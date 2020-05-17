package collector

import (
	"errors"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/go-vmware-nsxt/manager"
)

const (
	fakeLogicalSwitchID              = "fake-logical-switch-id"
	fakeLogicalSwitchDisplayName     = "fake-logical-switch-name"
	fakeLogicalSwitchTransportZoneID = "fake-transport-zone-id"
)

type mockLogicalSwitchClient struct {
	responses []mockLogicalSwitchResponse
}

type mockLogicalSwitchResponse struct {
	logicalSwitch  manager.LogicalSwitch
	Status         string
	StatisticValue int64
	Error          error
}

func (c *mockLogicalSwitchClient) ListAllLogicalSwitches() ([]manager.LogicalSwitch, error) {
	panic("unused function. Only used to satisfy LogicalSwitchClient interface")
}

func (c *mockLogicalSwitchClient) GetLogicalSwitchState(lswitchID string) (manager.LogicalSwitchState, error) {
	for _, res := range c.responses {
		if res.logicalSwitch.Id == lswitchID {
			return manager.LogicalSwitchState{
				State: res.Status,
			}, res.Error
		}
	}
	return manager.LogicalSwitchState{}, errors.New("error")
}

func (c *mockLogicalSwitchClient) GetLogicalSwitchStatistic(lswitchID string) (manager.LogicalSwitchStatistics, error) {
	for _, res := range c.responses {
		if res.logicalSwitch.Id == lswitchID {
			dataCounter := &manager.DataCounter{
				Total:   res.StatisticValue,
				Dropped: res.StatisticValue,
			}
			return manager.LogicalSwitchStatistics{
				RxPackets: dataCounter,
				RxBytes:   dataCounter,
				TxPackets: dataCounter,
				TxBytes:   dataCounter,
			}, res.Error
		}
	}
	panic("implement me")
}

func buildLogicalSwitchStatusResponse(id string, status string, err error) mockLogicalSwitchResponse {
	return mockLogicalSwitchResponse{
		logicalSwitch: manager.LogicalSwitch{
			Id:              fakeLogicalSwitchID + "-" + id,
			DisplayName:     fakeLogicalSwitchDisplayName + "-" + id,
			TransportZoneId: fakeLogicalSwitchTransportZoneID + "-" + id,
		},
		Status: status,
		Error:  err,
	}
}

func buildLogicalSwitchStatisticsResponse(id string, value int64, err error) mockLogicalSwitchResponse {
	return mockLogicalSwitchResponse{
		logicalSwitch: manager.LogicalSwitch{
			Id:              fakeLogicalSwitchID + "-" + id,
			DisplayName:     fakeLogicalSwitchDisplayName + "-" + id,
			TransportZoneId: fakeLogicalSwitchTransportZoneID + "-" + id,
		},
		StatisticValue: value,
		Error:          err,
	}
}

func TestLogicalSwitchCollector_GenerateLogicalSwitchStatusMetrics(t *testing.T) {
	testcases := []struct {
		description      string
		lswitchResponses []mockLogicalSwitchResponse
		expectedMetrics  []logicalSwitchStatusMetric
	}{
		{
			description: "Should return correct status value depending on logical switch state",
			lswitchResponses: []mockLogicalSwitchResponse{
				buildLogicalSwitchStatusResponse("01", "SUCCESS", nil),
				buildLogicalSwitchStatusResponse("02", "PARTIAL_SUCCESS", nil),
				buildLogicalSwitchStatusResponse("03", "IN_PROGRESS", nil),
				buildLogicalSwitchStatusResponse("04", "PENDING", nil),
				buildLogicalSwitchStatusResponse("05", "FAILED", nil),
				buildLogicalSwitchStatusResponse("06", "ORPHANED", nil),
				buildLogicalSwitchStatusResponse("07", "Success", nil),
				buildLogicalSwitchStatusResponse("08", "fAiLeD", nil),
			},
			expectedMetrics: []logicalSwitchStatusMetric{
				{
					ID:              "fake-logical-switch-id-01",
					Name:            "fake-logical-switch-name-01",
					TransportZoneID: "fake-transport-zone-id-01",
					Status:          1.0,
				},
				{
					ID:              "fake-logical-switch-id-02",
					Name:            "fake-logical-switch-name-02",
					TransportZoneID: "fake-transport-zone-id-02",
					Status:          0.0,
				},
				{
					ID:              "fake-logical-switch-id-03",
					Name:            "fake-logical-switch-name-03",
					TransportZoneID: "fake-transport-zone-id-03",
					Status:          0.0,
				},
				{
					ID:              "fake-logical-switch-id-04",
					Name:            "fake-logical-switch-name-04",
					TransportZoneID: "fake-transport-zone-id-04",
					Status:          0.0,
				},
				{
					ID:              "fake-logical-switch-id-05",
					Name:            "fake-logical-switch-name-05",
					TransportZoneID: "fake-transport-zone-id-05",
					Status:          0.0,
				},
				{
					ID:              "fake-logical-switch-id-06",
					Name:            "fake-logical-switch-name-06",
					TransportZoneID: "fake-transport-zone-id-06",
					Status:          0.0,
				},
				{
					ID:              "fake-logical-switch-id-07",
					Name:            "fake-logical-switch-name-07",
					TransportZoneID: "fake-transport-zone-id-07",
					Status:          1.0,
				},
				{
					ID:              "fake-logical-switch-id-08",
					Name:            "fake-logical-switch-name-08",
					TransportZoneID: "fake-transport-zone-id-08",
					Status:          0.0,
				},
			},
		},
		{
			description: "Should only return logical switch with valid response",
			lswitchResponses: []mockLogicalSwitchResponse{
				buildLogicalSwitchStatusResponse("01", "SUCCESS", nil),
				buildLogicalSwitchStatusResponse("02", "SUCCESS", errors.New("error get logical switch")),
			},
			expectedMetrics: []logicalSwitchStatusMetric{
				{
					ID:              "fake-logical-switch-id-01",
					Name:            "fake-logical-switch-name-01",
					TransportZoneID: "fake-transport-zone-id-01",
					Status:          1.0,
				},
			},
		},
		{
			description:      "Should return empty metrics when empty response",
			lswitchResponses: []mockLogicalSwitchResponse{},
			expectedMetrics:  []logicalSwitchStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockLogicalSwitchClient := &mockLogicalSwitchClient{
			responses: tc.lswitchResponses,
		}
		logger := log.NewNopLogger()
		lswitchCollector := newLogicalSwitchCollector(mockLogicalSwitchClient, logger)
		var logicalSwitches []manager.LogicalSwitch
		for _, res := range tc.lswitchResponses {
			logicalSwitches = append(logicalSwitches, res.logicalSwitch)
		}
		lswitchMetrics := lswitchCollector.generateLogicalSwitchStatusMetrics(logicalSwitches)
		assert.ElementsMatch(t, tc.expectedMetrics, lswitchMetrics, tc.description)
	}
}

func TestLogicalSwitchCollector_GenerateLogicalSwitchStatisticsMetrics(t *testing.T) {
	testcases := []struct {
		description      string
		lswitchResponses []mockLogicalSwitchResponse
		expectedMetrics  []logicalSwitchStatisticMetric
	}{
		{
			description: "Should return correct logical switch statistics value",
			lswitchResponses: []mockLogicalSwitchResponse{
				buildLogicalSwitchStatisticsResponse("01", 2, nil),
				buildLogicalSwitchStatisticsResponse("02", 3, nil),
			},
			expectedMetrics: []logicalSwitchStatisticMetric{
				{
					ID:              "fake-logical-switch-id-01",
					Name:            "fake-logical-switch-name-01",
					TransportZoneID: "fake-transport-zone-id-01",
					RxByteTotal:     2,
					RxByteDropped:   2,
					RxPacketTotal:   2,
					RxPacketDropped: 2,
					TxByteTotal:     2,
					TxByteDropped:   2,
					TxPacketTotal:   2,
					TxPacketDropped: 2,
				}, {
					ID:              "fake-logical-switch-id-02",
					Name:            "fake-logical-switch-name-02",
					TransportZoneID: "fake-transport-zone-id-02",
					RxByteTotal:     3,
					RxByteDropped:   3,
					RxPacketTotal:   3,
					RxPacketDropped: 3,
					TxByteTotal:     3,
					TxByteDropped:   3,
					TxPacketTotal:   3,
					TxPacketDropped: 3,
				},
			},
		}, {
			description: "Should only return logical switch with valid response",
			lswitchResponses: []mockLogicalSwitchResponse{
				buildLogicalSwitchStatisticsResponse("01", 2, nil),
				buildLogicalSwitchStatisticsResponse("02", 3, errors.New("error get logical switch statistic")),
			},
			expectedMetrics: []logicalSwitchStatisticMetric{
				{
					ID:              "fake-logical-switch-id-01",
					Name:            "fake-logical-switch-name-01",
					TransportZoneID: "fake-transport-zone-id-01",
					RxByteTotal:     2,
					RxByteDropped:   2,
					RxPacketTotal:   2,
					RxPacketDropped: 2,
					TxByteTotal:     2,
					TxByteDropped:   2,
					TxPacketTotal:   2,
					TxPacketDropped: 2,
				},
			},
		}, {
			description:      "Should return empty metrics when given empty logical switch",
			lswitchResponses: []mockLogicalSwitchResponse{},
			expectedMetrics:  []logicalSwitchStatisticMetric{},
		},
	}
	for _, tc := range testcases {
		mockLogicalSwitchClient := &mockLogicalSwitchClient{
			responses: tc.lswitchResponses,
		}
		logger := log.NewNopLogger()
		lswitchCollector := newLogicalSwitchCollector(mockLogicalSwitchClient, logger)
		var logicalSwitches []manager.LogicalSwitch
		for _, res := range tc.lswitchResponses {
			logicalSwitches = append(logicalSwitches, res.logicalSwitch)
		}
		metrics := lswitchCollector.generateLogicalSwitchStatisticMetrics(logicalSwitches)
		assert.ElementsMatch(t, tc.expectedMetrics, metrics, tc.description)
	}
}
