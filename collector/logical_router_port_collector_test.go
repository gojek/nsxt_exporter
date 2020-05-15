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
	fakeLogicalRouterPortID          = "fake-logical-router-port-id"
	fakeLogicalRouterPortDisplayName = "fake-logical-router-port-name"
	fakeLogicalRouterID              = "fake-logical-router-id"
)

type mockLogicalRouterPortClient struct {
	responses                  []mockLogicalRouterPortResponse
	logicalRouterPortListError error
}

type mockLogicalRouterPortResponse struct {
	ID              string
	DisplayName     string
	LogicalRouterID string
	Error           error

	RxTotalBytes     int64
	RxTotalPackets   int64
	RxDroppedPackets int64
	TxTotalBytes     int64
	TxTotalPackets   int64
	TxDroppedPackets int64
}

func (c *mockLogicalRouterPortClient) ListAllLogicalRouterPorts() ([]manager.LogicalRouterPort, error) {
	if c.logicalRouterPortListError != nil {
		return nil, c.logicalRouterPortListError
	}
	var logicalRouterPorts []manager.LogicalRouterPort
	for _, response := range c.responses {
		logicalRouterPort := manager.LogicalRouterPort{
			Id:              response.ID,
			DisplayName:     response.DisplayName,
			LogicalRouterId: response.LogicalRouterID,
		}
		logicalRouterPorts = append(logicalRouterPorts, logicalRouterPort)
	}
	return logicalRouterPorts, nil
}

func (c *mockLogicalRouterPortClient) GetLogicalRouterPortStatisticsSummary(lrportID string) (manager.LogicalRouterPortStatisticsSummary, error) {
	for _, res := range c.responses {
		if res.ID == lrportID {
			return manager.LogicalRouterPortStatisticsSummary{
				Rx: &manager.LogicalRouterPortCounters{
					TotalBytes:     res.RxTotalBytes,
					TotalPackets:   res.RxTotalPackets,
					DroppedPackets: res.RxDroppedPackets,
				},
				Tx: &manager.LogicalRouterPortCounters{
					TotalBytes:     res.TxTotalBytes,
					TotalPackets:   res.TxTotalPackets,
					DroppedPackets: res.TxDroppedPackets,
				},
			}, res.Error
		}
	}
	return manager.LogicalRouterPortStatisticsSummary{}, errors.New("error")
}

func buildLogicalRouterPortResponse(id string, baseValue int64, err error) mockLogicalRouterPortResponse {
	return mockLogicalRouterPortResponse{
		ID:               fmt.Sprintf("%s-%s", fakeLogicalRouterPortID, id),
		DisplayName:      fmt.Sprintf("%s-%s", fakeLogicalRouterPortDisplayName, id),
		LogicalRouterID:  fmt.Sprintf("%s-%s", fakeLogicalRouterID, id),
		Error:            err,
		RxTotalBytes:     baseValue,
		RxTotalPackets:   baseValue,
		RxDroppedPackets: baseValue,
		TxTotalBytes:     baseValue,
		TxTotalPackets:   baseValue,
		TxDroppedPackets: baseValue,
	}
}

func TestLogicalRouterPortCollector_GenerateLogicalRouterPortStatisticMetrics(t *testing.T) {
	testcases := []struct {
		description                string
		logicalRouterPortListError error
		logicalRouterPortResponses []mockLogicalRouterPortResponse
		expectedMetrics            []logicalRouterPortStatisticMetric
	}{
		{
			description:                "Should return correct statistics metrics",
			logicalRouterPortListError: nil,
			logicalRouterPortResponses: []mockLogicalRouterPortResponse{
				buildLogicalRouterPortResponse("01", 2, nil),
				buildLogicalRouterPortResponse("02", 3, nil),
			},
			expectedMetrics: []logicalRouterPortStatisticMetric{
				{
					LogicalRouterPort: manager.LogicalRouterPort{
						Id:              "fake-logical-router-port-id-01",
						DisplayName:     "fake-logical-router-port-name-01",
						LogicalRouterId: "fake-logical-router-id-01",
					},
					Rx: &manager.LogicalRouterPortCounters{
						TotalBytes:     2,
						TotalPackets:   2,
						DroppedPackets: 2,
					},
					Tx: &manager.LogicalRouterPortCounters{
						TotalBytes:     2,
						TotalPackets:   2,
						DroppedPackets: 2,
					},
				}, {
					LogicalRouterPort: manager.LogicalRouterPort{
						Id:              "fake-logical-router-port-id-02",
						DisplayName:     "fake-logical-router-port-name-02",
						LogicalRouterId: "fake-logical-router-id-02",
					},
					Rx: &manager.LogicalRouterPortCounters{
						TotalBytes:     3,
						TotalPackets:   3,
						DroppedPackets: 3,
					},
					Tx: &manager.LogicalRouterPortCounters{
						TotalBytes:     3,
						TotalPackets:   3,
						DroppedPackets: 3,
					},
				},
			},
		}, {
			description:                "Should only return logical router port with valid response",
			logicalRouterPortListError: nil,
			logicalRouterPortResponses: []mockLogicalRouterPortResponse{
				buildLogicalRouterPortResponse("01", 2, nil),
				buildLogicalRouterPortResponse("02", 3, errors.New("error get statistic")),
			},
			expectedMetrics: []logicalRouterPortStatisticMetric{
				{
					LogicalRouterPort: manager.LogicalRouterPort{
						Id:              "fake-logical-router-port-id-01",
						DisplayName:     "fake-logical-router-port-name-01",
						LogicalRouterId: "fake-logical-router-id-01",
					},
					Rx: &manager.LogicalRouterPortCounters{
						TotalBytes:     2,
						TotalPackets:   2,
						DroppedPackets: 2,
					},
					Tx: &manager.LogicalRouterPortCounters{
						TotalBytes:     2,
						TotalPackets:   2,
						DroppedPackets: 2,
					},
				},
			},
		}, {
			description:                "Should return empty metrics when fail to list logical router port",
			logicalRouterPortListError: errors.New("failed to list logical router port"),
			logicalRouterPortResponses: []mockLogicalRouterPortResponse{
				buildLogicalRouterPortResponse("01", 2, nil),
				buildLogicalRouterPortResponse("02", 3, nil),
			},
			expectedMetrics: []logicalRouterPortStatisticMetric{},
		}, {
			description:                "Should return empty metrics when there's no logical router port",
			logicalRouterPortListError: nil,
			logicalRouterPortResponses: []mockLogicalRouterPortResponse{},
			expectedMetrics:            []logicalRouterPortStatisticMetric{},
		},
	}
	for _, tc := range testcases {
		mockLogicalRouterPortClient := &mockLogicalRouterPortClient{
			responses:                  tc.logicalRouterPortResponses,
			logicalRouterPortListError: tc.logicalRouterPortListError,
		}
		logger := log.NewNopLogger()
		logicalRouterPortCollector := newLogicalRouterPortCollector(mockLogicalRouterPortClient, logger)
		logicalRouterPortMetrics := logicalRouterPortCollector.generateLogicalRouterPortStatisticMetrics()
		assert.ElementsMatch(t, tc.expectedMetrics, logicalRouterPortMetrics, tc.description)
	}
}
