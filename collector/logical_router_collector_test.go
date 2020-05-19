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
	fakeNatTotalPackets = 1
	fakeNatTotalBytes   = 1
)

type mockLogicalRouterClient struct {
	responses []mockLogicalRouterResponse
}

type mockLogicalRouterResponse struct {
	LogicalRouter manager.LogicalRouter
	Error         error

	NatTotalPackets int64
	NatTotalBytes   int64
}

func (c *mockLogicalRouterClient) ListAllLogicalRouters() ([]manager.LogicalRouter, error) {
	panic("unused function. Only used to satisfy LogicalRouterClient interface")
}

func (c *mockLogicalRouterClient) GetNatStatisticsPerLogicalRouter(lrouterID string) (manager.NatStatisticsPerLogicalRouter, error) {
	for _, res := range c.responses {
		if res.LogicalRouter.Id == lrouterID {
			return manager.NatStatisticsPerLogicalRouter{
				StatisticsAcrossAllNodes: &manager.NatCounters{
					TotalPackets: res.NatTotalPackets,
					TotalBytes:   res.NatTotalBytes,
				},
			}, res.Error
		}
	}
	return manager.NatStatisticsPerLogicalRouter{}, errors.New("error logical router not found")
}

func buildLogicalRouterResponse(id string, err error) mockLogicalRouterResponse {
	return mockLogicalRouterResponse{
		LogicalRouter: manager.LogicalRouter{
			Id: fmt.Sprintf("%s-%s", fakeLogicalRouterID, id),
		},
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
		logicalRouterResponses []mockLogicalRouterResponse
		expectedMetrics        []logicalRouterNatStatisticMetric
	}{
		{
			description: "Should return correct statistics metrics",
			logicalRouterResponses: []mockLogicalRouterResponse{
				buildLogicalRouterResponse("01", nil),
				buildLogicalRouterResponse("02", nil),
			},
			expectedMetrics: []logicalRouterNatStatisticMetric{
				{
					LogicalRouterID: fmt.Sprintf("%s-01", fakeLogicalRouterID),
					NatTotalPackets: fakeNatTotalPackets,
					NatTotalBytes:   fakeNatTotalBytes,
				},
				{
					LogicalRouterID: fmt.Sprintf("%s-02", fakeLogicalRouterID),
					NatTotalPackets: fakeNatTotalPackets,
					NatTotalBytes:   fakeNatTotalBytes,
				},
			},
		},
		{
			description: "Should only return statistics with valid response",
			logicalRouterResponses: []mockLogicalRouterResponse{
				buildLogicalRouterResponse("01", nil),
				buildLogicalRouterResponse("02", errors.New("error get logical router nat statistic")),
			},
			expectedMetrics: []logicalRouterNatStatisticMetric{
				{
					LogicalRouterID: fmt.Sprintf("%s-01", fakeLogicalRouterID),
					NatTotalPackets: fakeNatTotalPackets,
					NatTotalBytes:   fakeNatTotalBytes,
				},
			},
		},
		{
			description:            "Should return empty metrics when given empty logical router",
			logicalRouterResponses: []mockLogicalRouterResponse{},
			expectedMetrics:        []logicalRouterNatStatisticMetric{},
		},
	}
	for _, tc := range testcases {
		mockLogicalRouterCLient := &mockLogicalRouterClient{
			responses: tc.logicalRouterResponses,
		}
		logger := log.NewNopLogger()
		lrouterCollector := newLogicalRouterCollector(mockLogicalRouterCLient, logger)
		logicalRouters := buildLogicalRouters(tc.logicalRouterResponses)
		metrics := lrouterCollector.generateLogicalRouterNatStatisticMetrics(logicalRouters)
		assert.ElementsMatch(t, tc.expectedMetrics, metrics, tc.description)
	}
}
