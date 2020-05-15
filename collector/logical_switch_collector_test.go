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
	responses        []mockLogicalSwitchResponse
	lswitchListError error
}

type mockLogicalSwitchResponse struct {
	logicalSwitch manager.LogicalSwitch
	Status        string
	Error         error
}

func (c *mockLogicalSwitchClient) ListAllLogicalSwitches() ([]manager.LogicalSwitch, error) {
	if c.lswitchListError != nil {
		return nil, c.lswitchListError
	}
	var lswitches []manager.LogicalSwitch
	for _, res := range c.responses {
		lswitches = append(lswitches, res.logicalSwitch)
	}
	return lswitches, nil
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

func buildLogicalSwitchResponse(id string, status string, err error) mockLogicalSwitchResponse {
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

func TestLogicalSwitchCollector_GenerateLogicalSwitchStatusMetrics(t *testing.T) {
	testcases := []struct {
		description      string
		lswitchListError error
		lswitchResponses []mockLogicalSwitchResponse
		expectedMetrics  []logicalSwitchStatusMetric
	}{
		{
			description:      "Should return correct status value depending on logical switch state",
			lswitchListError: nil,
			lswitchResponses: []mockLogicalSwitchResponse{
				buildLogicalSwitchResponse("01", "SUCCESS", nil),
				buildLogicalSwitchResponse("02", "PARTIAL_SUCCESS", nil),
				buildLogicalSwitchResponse("03", "IN_PROGRESS", nil),
				buildLogicalSwitchResponse("04", "PENDING", nil),
				buildLogicalSwitchResponse("05", "FAILED", nil),
				buildLogicalSwitchResponse("06", "ORPHANED", nil),
				buildLogicalSwitchResponse("07", "Success", nil),
				buildLogicalSwitchResponse("08", "fAiLeD", nil),
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
			description:      "Should only return logical switch with valid response",
			lswitchListError: nil,
			lswitchResponses: []mockLogicalSwitchResponse{
				buildLogicalSwitchResponse("01", "SUCCESS", nil),
				buildLogicalSwitchResponse("02", "SUCCESS", errors.New("error get logical switch")),
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
			description:      "Should return empty metrics when fail listing logical switch",
			lswitchListError: errors.New("error list logical switch"),
			lswitchResponses: []mockLogicalSwitchResponse{
				buildLogicalSwitchResponse("01", "SUCCESS", nil),
				buildLogicalSwitchResponse("02", "SUCCESS", nil),
			},
			expectedMetrics: []logicalSwitchStatusMetric{},
		},
		{
			description:      "Should return empty metrics when empty response",
			lswitchListError: nil,
			lswitchResponses: []mockLogicalSwitchResponse{},
			expectedMetrics:  []logicalSwitchStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockLogicalSwitchClient := &mockLogicalSwitchClient{
			lswitchListError: tc.lswitchListError,
			responses:        tc.lswitchResponses,
		}
		logger := log.NewNopLogger()
		lswitchCollector := newLogicalSwitchCollector(mockLogicalSwitchClient, logger)
		lswitchMetrics := lswitchCollector.generateLogicalSwitchStatusMetrics()
		assert.ElementsMatch(t, tc.expectedMetrics, lswitchMetrics, tc.description)
	}
}
