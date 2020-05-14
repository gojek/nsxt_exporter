package collector

import (
	"errors"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/go-vmware-nsxt/manager"
	"testing"
)

const (
	fakeLogicalPortID          = "fake-logical-port-id"
	fakeLogicalPortDisplayName = "fake-logical-port-name"
	faceLogicalSwitchID        = "fake-logical-switch-id"
)

type mockLogicalPortClient struct {
	responses            []mockLogicalPortResponse
	logicalPortListError error
}

type mockLogicalPortResponse struct {
	ID              string
	DisplayName     string
	Status          string
	LogicalSwitchID string
	Error           error
}

func (c *mockLogicalPortClient) ListLogicalPorts(localVarOptionals map[string]interface{}) (manager.LogicalPortListResult, error) {
	if c.logicalPortListError != nil {
		return manager.LogicalPortListResult{}, c.logicalPortListError
	}
	var logicalPorts []manager.LogicalPort
	for _, response := range c.responses {
		logicalPort := manager.LogicalPort{
			Id:              response.ID,
			DisplayName:     response.DisplayName,
			LogicalSwitchId: response.LogicalSwitchID,
		}
		logicalPorts = append(logicalPorts, logicalPort)
	}
	return manager.LogicalPortListResult{
		Results: logicalPorts,
	}, nil
}

func (c *mockLogicalPortClient) GetLogicalPortOperationalStatus(lportID string, localVarOptionals map[string]interface{}) (manager.LogicalPortOperationalStatus, error) {
	for _, res := range c.responses {
		if res.ID == lportID {
			return manager.LogicalPortOperationalStatus{
				Status: res.Status,
			}, res.Error
		}
	}
	return manager.LogicalPortOperationalStatus{}, errors.New("error")
}

func buildLogicalPortResponse(id string, status string, err error) mockLogicalPortResponse {
	return mockLogicalPortResponse{
		ID:              fmt.Sprintf("%s-%s", fakeLogicalPortID, id),
		DisplayName:     fmt.Sprintf("%s-%s", fakeLogicalPortDisplayName, id),
		LogicalSwitchID: fmt.Sprintf("%s-%s", faceLogicalSwitchID, id),
		Status:          status,
		Error:           err,
	}
}

func TestLogicalPortCollector_GenerateLogicalPortStatusMetrics(t *testing.T) {
	testcases := []struct {
		logicalPortListError error
		logicalPortResponses []mockLogicalPortResponse
		expectedMetrics      []logicalPortStatusMetric
	}{
		{
			logicalPortListError: nil,
			logicalPortResponses: []mockLogicalPortResponse{
				buildLogicalPortResponse("01", "UP", nil),
				buildLogicalPortResponse("02", "DOWN", nil),
				buildLogicalPortResponse("03", "UNKNOWN", nil),
			},
			expectedMetrics: []logicalPortStatusMetric{
				{
					ID:              "fake-logical-port-id-01",
					Name:            "fake-logical-port-name-01",
					LogicalSwitchID: "fake-logical-switch-id-01",
					Status:          1.0,
				}, {
					ID:              "fake-logical-port-id-02",
					Name:            "fake-logical-port-name-02",
					LogicalSwitchID: "fake-logical-switch-id-02",
					Status:          0.0,
				}, {
					ID:              "fake-logical-port-id-03",
					Name:            "fake-logical-port-name-03",
					LogicalSwitchID: "fake-logical-switch-id-03",
					Status:          0.0,
				},
			},
		}, {
			logicalPortListError: nil,
			logicalPortResponses: []mockLogicalPortResponse{
				buildLogicalPortResponse("01", "UP", nil),
				buildLogicalPortResponse("02", "UP", errors.New("error get logical port status")),
			},
			expectedMetrics: []logicalPortStatusMetric{
				{
					ID:              "fake-logical-port-id-01",
					Name:            "fake-logical-port-name-01",
					LogicalSwitchID: "fake-logical-switch-id-01",
					Status:          1.0,
				},
			},
		}, {
			logicalPortListError: errors.New("error list logical ports"),
			logicalPortResponses: []mockLogicalPortResponse{
				buildLogicalPortResponse("01", "UP", nil),
				buildLogicalPortResponse("02", "UP", nil),
			},
			expectedMetrics: []logicalPortStatusMetric{},
		}, {
			logicalPortListError: nil,
			logicalPortResponses: []mockLogicalPortResponse{},
			expectedMetrics:      []logicalPortStatusMetric{},
		},
	}
	for _, testcase := range testcases {
		mockLogicalPortClient := &mockLogicalPortClient{
			responses:            testcase.logicalPortResponses,
			logicalPortListError: testcase.logicalPortListError,
		}
		logger := log.NewNopLogger()
		logicalPortCollector := newLogicalPortCollector(mockLogicalPortClient, logger)
		logicalPortMetrics := logicalPortCollector.generateLogicalPortStatusMetrics()
		assert.ElementsMatch(t, testcase.expectedMetrics, logicalPortMetrics)
	}
}
