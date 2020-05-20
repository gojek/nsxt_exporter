package collector

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/go-vmware-nsxt/manager"
)

const (
	fakeLogicalRouterID = "fake-logical-router-id"
	fakeNatRuleID       = "fake-nat-rule-id"
	fakeNatRuleName     = "fake-nat-rule-name"
	fakeNatTotalPackets = 1
	fakeNatTotalBytes   = 1
)

type mockLogicalRouterClient struct {
	responses        []mockLogicalRouterResponse
	natRuleListError error
}

type mockLogicalRouterResponse struct {
	LogicalRouter manager.LogicalRouter
	NatRules      []manager.NatRule
	Error         error

	NatTotalPackets int64
	NatTotalBytes   int64
}

func (c *mockLogicalRouterClient) ListAllLogicalRouters() ([]manager.LogicalRouter, error) {
	panic("unused function. Only used to satisfy LogicalRouterClient interface")
}

func (c *mockLogicalRouterClient) ListAllNatRules(lrouterID string) ([]manager.NatRule, error) {
	if c.natRuleListError != nil {
		return nil, c.natRuleListError
	}
	var natRules []manager.NatRule
	for _, response := range c.responses {
		if response.LogicalRouter.Id != lrouterID {
			continue
		}
		for _, rule := range response.NatRules {
			natRule := manager.NatRule{
				Id:              rule.Id,
				DisplayName:     rule.DisplayName,
				LogicalRouterId: lrouterID,
			}
			natRules = append(natRules, natRule)
		}
	}
	return natRules, nil
}

func (c *mockLogicalRouterClient) GetNatStatisticsPerRule(lrouterID, ruleID string) (manager.NatStatisticsPerRule, error) {
	for _, res := range c.responses {
		if res.LogicalRouter.Id != lrouterID {
			continue
		}
		for _, rule := range res.NatRules {
			if rule.Id != ruleID {
				continue
			}
			return manager.NatStatisticsPerRule{
				Id:              ruleID,
				LogicalRouterId: lrouterID,
				TotalPackets:    fakeNatTotalPackets,
				TotalBytes:      fakeNatTotalBytes,
			}, res.Error
		}
	}
	return manager.NatStatisticsPerRule{}, errors.New("error nat rule not found")
}

func buildLogicalRouterResponse(lrouterID string, ruleIDs []string, err error) mockLogicalRouterResponse {
	var natRules []manager.NatRule
	for _, ruleID := range ruleIDs {
		natRules = append(natRules, manager.NatRule{
			Id:          fmt.Sprintf("%s-%s", fakeNatRuleID, ruleID),
			DisplayName: fmt.Sprintf("%s-%s", fakeNatRuleName, ruleID),
		})
	}
	return mockLogicalRouterResponse{
		LogicalRouter: manager.LogicalRouter{
			Id: fmt.Sprintf("%s-%s", fakeLogicalRouterID, lrouterID),
		},
		NatRules:        natRules,
		Error:           err,
		NatTotalPackets: fakeNatTotalPackets,
		NatTotalBytes:   fakeNatTotalBytes,
	}
}

func buildLogicalRouters(logicalRouterResponses []mockLogicalRouterResponse) []manager.LogicalRouter {
	var logicalRouters []manager.LogicalRouter
	for _, res := range logicalRouterResponses {
		logicalRouters = append(logicalRouters, res.LogicalRouter)
	}
	return logicalRouters
}

func TestLogicalRouterCollector_GenerateLogicalRouterNatStatisticMetrics(t *testing.T) {
	testcases := []struct {
		description            string
		natRuleListError       error
		logicalRouterResponses []mockLogicalRouterResponse
		expectedMetrics        []natRuleStatisticMetric
	}{
		{
			description: "Should return correct statistics metrics",
			logicalRouterResponses: []mockLogicalRouterResponse{
				buildLogicalRouterResponse("1", []string{"1", "2"}, nil),
				buildLogicalRouterResponse("2", []string{"3"}, nil),
			},
			expectedMetrics: []natRuleStatisticMetric{
				{
					ID:              fmt.Sprintf("%s-1", fakeNatRuleID),
					Name:            fmt.Sprintf("%s-1", fakeNatRuleName),
					LogicalRouterID: fmt.Sprintf("%s-1", fakeLogicalRouterID),
					NatTotalPackets: fakeNatTotalPackets,
					NatTotalBytes:   fakeNatTotalBytes,
				},
				{
					ID:              fmt.Sprintf("%s-2", fakeNatRuleID),
					Name:            fmt.Sprintf("%s-2", fakeNatRuleName),
					LogicalRouterID: fmt.Sprintf("%s-1", fakeLogicalRouterID),
					NatTotalPackets: fakeNatTotalPackets,
					NatTotalBytes:   fakeNatTotalBytes,
				},
				{
					ID:              fmt.Sprintf("%s-3", fakeNatRuleID),
					Name:            fmt.Sprintf("%s-3", fakeNatRuleName),
					LogicalRouterID: fmt.Sprintf("%s-2", fakeLogicalRouterID),
					NatTotalPackets: fakeNatTotalPackets,
					NatTotalBytes:   fakeNatTotalBytes,
				},
			},
		},
		{
			description: "Should only return statistics with valid response",
			logicalRouterResponses: []mockLogicalRouterResponse{
				buildLogicalRouterResponse("1", []string{"1", "2"}, errors.New("error get nat rule statistic")),
				buildLogicalRouterResponse("2", []string{"3"}, nil),
			},
			expectedMetrics: []natRuleStatisticMetric{
				{
					ID:              fmt.Sprintf("%s-3", fakeNatRuleID),
					Name:            fmt.Sprintf("%s-3", fakeNatRuleName),
					LogicalRouterID: fmt.Sprintf("%s-2", fakeLogicalRouterID),
					NatTotalPackets: fakeNatTotalPackets,
					NatTotalBytes:   fakeNatTotalBytes,
				},
			},
		},
		{
			description:            "Should return empty metrics when given empty logical router",
			logicalRouterResponses: []mockLogicalRouterResponse{},
			expectedMetrics:        []natRuleStatisticMetric{},
		},
		{
			description:            "Should return empty metrics when error listing nat rules",
			natRuleListError:       errors.New("error list nat rules"),
			logicalRouterResponses: []mockLogicalRouterResponse{},
			expectedMetrics:        []natRuleStatisticMetric{},
		},
	}
	for _, tc := range testcases {
		mockLogicalRouterCLient := &mockLogicalRouterClient{
			responses: tc.logicalRouterResponses,
		}
		logger := log.NewNopLogger()
		lrouterCollector := newLogicalRouterCollector(mockLogicalRouterCLient, logger)
		logicalRouters := buildLogicalRouters(tc.logicalRouterResponses)
		metrics := lrouterCollector.generateNatRuleStatisticMetrics(logicalRouters)
		fmt.Println(metrics)
		assert.ElementsMatch(t, tc.expectedMetrics, metrics, tc.description)
	}
}
