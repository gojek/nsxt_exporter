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
	fakeLogicalRouterID              = "fake-logical-router-id"
	fakeLogicalRouterName            = "fake-logical-router-name"
	fakeLogicalRouterServiceRouterID = "fake-service-router-id"
	fakeLogicalRouterTransportNodeID = "fake-transport-node-id"
	fakeNatRuleID                    = "fake-nat-rule-id"
	fakeNatRuleName                  = "fake-nat-rule-name"
	fakeNatRuleType                  = "fake-nat-rule-type"
	fakeNatTotalPackets              = 1
	fakeNatTotalBytes                = 1
)

type mockLogicalRouterClient struct {
	responses        []mockLogicalRouterResponse
	natRuleListError error
}

type mockLogicalRouterResponse struct {
	LogicalRouter       manager.LogicalRouter
	LogicalRouterStatus []manager.LogicalRouterStatusPerNode
	NatRules            []manager.NatRule
	Error               error

	NatTotalPackets int64
	NatTotalBytes   int64
}

func (c *mockLogicalRouterClient) ListAllLogicalRouters() ([]manager.LogicalRouter, error) {
	panic("unused function. Only used to satisfy LogicalRouterClient interface")
}

func (c *mockLogicalRouterClient) GetLogicalRouterStatus(lrouterID string) (manager.LogicalRouterStatus, error) {
	for _, res := range c.responses {
		if res.LogicalRouter.Id == lrouterID {
			return manager.LogicalRouterStatus{
				LogicalRouterId: lrouterID,
				PerNodeStatus:   res.LogicalRouterStatus,
			}, res.Error
		}
	}
	return manager.LogicalRouterStatus{}, errors.New("error logical router not found")
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
				Action:          rule.Action,
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

func buildLogicalRouterResponseWithStatus(lrouterID string, highAvailabilityStatus []string, err error) mockLogicalRouterResponse {
	var lrouterStatus []manager.LogicalRouterStatusPerNode
	for _, status := range highAvailabilityStatus {
		lrouterStatus = append(lrouterStatus, manager.LogicalRouterStatusPerNode{
			HighAvailabilityStatus: status,
			ServiceRouterId:        fakeLogicalRouterServiceRouterID,
			TransportNodeId:        fakeLogicalRouterTransportNodeID,
		})
	}
	return mockLogicalRouterResponse{
		LogicalRouter: manager.LogicalRouter{
			Id:          fmt.Sprintf("%s-%s", fakeLogicalRouterID, lrouterID),
			DisplayName: fmt.Sprintf("%s-%s", fakeLogicalRouterName, lrouterID),
		},
		LogicalRouterStatus: lrouterStatus,
		Error:               err,
	}
}

func buildLogicalRouterResponseWithNatRules(lrouterID string, ruleIDs []string, err error) mockLogicalRouterResponse {
	var natRules []manager.NatRule
	for _, ruleID := range ruleIDs {
		natRules = append(natRules, manager.NatRule{
			Id:          fmt.Sprintf("%s-%s", fakeNatRuleID, ruleID),
			DisplayName: fmt.Sprintf("%s-%s", fakeNatRuleName, ruleID),
			Action:      fakeNatRuleType,
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

func buildExpectedHighAvailabilityStatusDetails(nonZeroStatus string) map[string]float64 {
	statusDetails := map[string]float64{
		"ACTIVE":  0.0,
		"STANDBY": 0.0,
	}
	statusDetails[nonZeroStatus] = 1.0
	return statusDetails
}

func TestLogicalRouterCollector_GenerateLogicalRouterStatusMetrics(t *testing.T) {
	testcases := []struct {
		description            string
		logicalRouterResponses []mockLogicalRouterResponse
		expectedMetrics        []logicalRouterStatusMetric
	}{
		{
			description: "Should return correct status metrics",
			logicalRouterResponses: []mockLogicalRouterResponse{
				buildLogicalRouterResponseWithStatus("1", []string{"ACTIVE", "STANDBY"}, nil),
				buildLogicalRouterResponseWithStatus("2", []string{"ACTIVE"}, nil),
			},
			expectedMetrics: []logicalRouterStatusMetric{
				{
					ID:                           fmt.Sprintf("%s-1", fakeLogicalRouterID),
					Name:                         fmt.Sprintf("%s-1", fakeLogicalRouterName),
					TransportNodeID:              fakeLogicalRouterTransportNodeID,
					ServiceRouterID:              fakeLogicalRouterServiceRouterID,
					HighAvailabilityStatusDetail: buildExpectedHighAvailabilityStatusDetails("ACTIVE"),
				},
				{
					ID:                           fmt.Sprintf("%s-1", fakeLogicalRouterID),
					Name:                         fmt.Sprintf("%s-1", fakeLogicalRouterName),
					TransportNodeID:              fakeLogicalRouterTransportNodeID,
					ServiceRouterID:              fakeLogicalRouterServiceRouterID,
					HighAvailabilityStatusDetail: buildExpectedHighAvailabilityStatusDetails("STANDBY"),
				},
				{
					ID:                           fmt.Sprintf("%s-2", fakeLogicalRouterID),
					Name:                         fmt.Sprintf("%s-2", fakeLogicalRouterName),
					TransportNodeID:              fakeLogicalRouterTransportNodeID,
					ServiceRouterID:              fakeLogicalRouterServiceRouterID,
					HighAvailabilityStatusDetail: buildExpectedHighAvailabilityStatusDetails("ACTIVE"),
				},
			},
		},
		{
			description: "Should only return status with valid response",
			logicalRouterResponses: []mockLogicalRouterResponse{
				buildLogicalRouterResponseWithStatus("1", []string{"ACTIVE", "STANDBY"}, errors.New("error get logical router status")),
				buildLogicalRouterResponseWithStatus("2", []string{"ACTIVE"}, nil),
			},
			expectedMetrics: []logicalRouterStatusMetric{
				{
					ID:                           fmt.Sprintf("%s-2", fakeLogicalRouterID),
					Name:                         fmt.Sprintf("%s-2", fakeLogicalRouterName),
					TransportNodeID:              fakeLogicalRouterTransportNodeID,
					ServiceRouterID:              fakeLogicalRouterServiceRouterID,
					HighAvailabilityStatusDetail: buildExpectedHighAvailabilityStatusDetails("ACTIVE"),
				},
			},
		},
		{
			description:            "Should return empty metrics when empty response",
			logicalRouterResponses: []mockLogicalRouterResponse{},
			expectedMetrics:        []logicalRouterStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockLogicalRouterClient := &mockLogicalRouterClient{
			responses: tc.logicalRouterResponses,
		}
		logger := log.NewNopLogger()
		lrouterCollector := newLogicalRouterCollector(mockLogicalRouterClient, logger)
		logicalRouters := buildLogicalRouters(tc.logicalRouterResponses)
		metrics := lrouterCollector.generateLogicalRouterStatusMetrics(logicalRouters)
		assert.ElementsMatch(t, tc.expectedMetrics, metrics, tc.description)
	}
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
				buildLogicalRouterResponseWithNatRules("1", []string{"1", "2"}, nil),
				buildLogicalRouterResponseWithNatRules("2", []string{"3"}, nil),
			},
			expectedMetrics: []natRuleStatisticMetric{
				{
					ID:              fmt.Sprintf("%s-1", fakeNatRuleID),
					Name:            fmt.Sprintf("%s-1", fakeNatRuleName),
					Type:            fakeNatRuleType,
					LogicalRouterID: fmt.Sprintf("%s-1", fakeLogicalRouterID),
					NatTotalPackets: fakeNatTotalPackets,
					NatTotalBytes:   fakeNatTotalBytes,
				},
				{
					ID:              fmt.Sprintf("%s-2", fakeNatRuleID),
					Name:            fmt.Sprintf("%s-2", fakeNatRuleName),
					Type:            fakeNatRuleType,
					LogicalRouterID: fmt.Sprintf("%s-1", fakeLogicalRouterID),
					NatTotalPackets: fakeNatTotalPackets,
					NatTotalBytes:   fakeNatTotalBytes,
				},
				{
					ID:              fmt.Sprintf("%s-3", fakeNatRuleID),
					Name:            fmt.Sprintf("%s-3", fakeNatRuleName),
					Type:            fakeNatRuleType,
					LogicalRouterID: fmt.Sprintf("%s-2", fakeLogicalRouterID),
					NatTotalPackets: fakeNatTotalPackets,
					NatTotalBytes:   fakeNatTotalBytes,
				},
			},
		},
		{
			description: "Should only return statistics with valid response",
			logicalRouterResponses: []mockLogicalRouterResponse{
				buildLogicalRouterResponseWithNatRules("1", []string{"1", "2"}, errors.New("error get nat rule statistic")),
				buildLogicalRouterResponseWithNatRules("2", []string{"3"}, nil),
			},
			expectedMetrics: []natRuleStatisticMetric{
				{
					ID:              fmt.Sprintf("%s-3", fakeNatRuleID),
					Name:            fmt.Sprintf("%s-3", fakeNatRuleName),
					Type:            fakeNatRuleType,
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
		mockLogicalRouterClient := &mockLogicalRouterClient{
			responses: tc.logicalRouterResponses,
		}
		logger := log.NewNopLogger()
		lrouterCollector := newLogicalRouterCollector(mockLogicalRouterClient, logger)
		logicalRouters := buildLogicalRouters(tc.logicalRouterResponses)
		metrics := lrouterCollector.generateNatRuleStatisticMetrics(logicalRouters)
		assert.ElementsMatch(t, tc.expectedMetrics, metrics, tc.description)
	}
}
