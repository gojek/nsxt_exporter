package generator_test

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"nsxt_exporter/generator"
	"testing"

	"nsxt_exporter/client"

	"github.com/go-kit/kit/log"
	"github.com/vmware/go-vmware-nsxt/manager"
)

type logicalPortMetricGeneratorTestContext struct {
	nsxtClientMock       *client.NSXTClientMock
	lportMetricGenerator generator.LogicalPortMetricGenerator
}

func (ctx *logicalPortMetricGeneratorTestContext) setUp() {
	ctx.nsxtClientMock = &client.NSXTClientMock{}
	ctx.lportMetricGenerator = generator.NewLogicalPortMetricGenerator(ctx.nsxtClientMock, log.NewNopLogger())
}

func TestLogicalPortMetricGenerator_GenerateLogicalPortStatusSummary(t *testing.T) {
	tests := []struct {
		mockLogicalStatusSummary manager.LogicalPortStatusSummary
		mockError                error
		expectedResult           manager.LogicalPortStatusSummary
		expectedSuccess          bool
	}{
		{
			mockLogicalStatusSummary: manager.LogicalPortStatusSummary{TotalPorts: 10, UpPorts: 7},
			mockError:                nil,
			expectedResult:           manager.LogicalPortStatusSummary{TotalPorts: 10, UpPorts: 7},
			expectedSuccess:          true,
		}, {
			mockLogicalStatusSummary: manager.LogicalPortStatusSummary{TotalPorts: 12, UpPorts: 12},
			mockError:                nil,
			expectedResult:           manager.LogicalPortStatusSummary{TotalPorts: 12, UpPorts: 12},
			expectedSuccess:          true,
		}, {
			mockLogicalStatusSummary: manager.LogicalPortStatusSummary{},
			mockError:                errors.New("generated error"),
			expectedResult:           manager.LogicalPortStatusSummary{},
			expectedSuccess:          false,
		},
	}
	for _, test := range tests {
		ctx := logicalPortMetricGeneratorTestContext{}
		ctx.setUp()
		ctx.nsxtClientMock.
			On("GetLogicalPortStatusSummary", mock.Anything).
			Return(test.mockLogicalStatusSummary, test.mockError).
			Once()
		result, ok := ctx.lportMetricGenerator.GenerateLogicalPortStatusSummary()
		assert.Equal(t, test.expectedResult, result)
		assert.Equal(t, test.expectedSuccess, ok)
		ctx.nsxtClientMock.AssertExpectations(t)
	}
}

type listLogicalPortsMockParam struct {
	options map[string]interface{}
	result  manager.LogicalPortListResult
	error   error
}

type getLogicalPortOperationalStatusParam struct {
	id     string
	result manager.LogicalPortOperationalStatus
	error  error
}

func TestLogicalPortMetricGenerator_GenerateLogicalPortStatusMetrics(t *testing.T) {
	tests := map[string]struct {
		listLogicalPortsMockParams                []listLogicalPortsMockParam
		getLogicalPortOperationalStatusMockParams []getLogicalPortOperationalStatusParam
		expectedResult                            []generator.LogicalPortStatus
		expectedStatus                            bool
	}{
		"single-logical-port": {
			listLogicalPortsMockParams: []listLogicalPortsMockParam{{
				options: map[string]interface{}{"cursor": ""},
				result: manager.LogicalPortListResult{
					Results: []manager.LogicalPort{{
						Id:          "id01",
						DisplayName: "name01",
					}},
					Cursor: "",
				},
				error: nil,
			}},
			getLogicalPortOperationalStatusMockParams: []getLogicalPortOperationalStatusParam{{
				id: "id01",
				result: manager.LogicalPortOperationalStatus{
					Status: "UP",
				},
				error: nil,
			}},
			expectedResult: []generator.LogicalPortStatus{{
				Status: 1,
				ID:     "id01",
				Name:   "name01",
			}},
			expectedStatus: true,
		},
		"list-logical-ports-failed": {
			listLogicalPortsMockParams: []listLogicalPortsMockParam{{
				options: map[string]interface{}{"cursor": ""},
				result:  manager.LogicalPortListResult{},
				error:   errors.New("generated error"),
			}},
			getLogicalPortOperationalStatusMockParams: []getLogicalPortOperationalStatusParam{},
			expectedResult: nil,
			expectedStatus: false,
		},
		"multiple-logical-ports-with-one-failed": {
			listLogicalPortsMockParams: []listLogicalPortsMockParam{{
				options: map[string]interface{}{"cursor": ""},
				result: manager.LogicalPortListResult{
					Results: []manager.LogicalPort{{
						Id:          "id01",
						DisplayName: "name01",
					}, {
						Id:          "id02",
						DisplayName: "name02",
					}},
					Cursor: "",
				},
				error: nil,
			}},
			getLogicalPortOperationalStatusMockParams: []getLogicalPortOperationalStatusParam{{
				id:     "id01",
				result: manager.LogicalPortOperationalStatus{},
				error:  errors.New("generated error"),
			}, {
				id: "id02",
				result: manager.LogicalPortOperationalStatus{
					Status: "DOWN",
				},
				error: nil,
			}},
			expectedResult: []generator.LogicalPortStatus{{
				Status: 0,
				ID:     "id02",
				Name:   "name02",
			}},
			expectedStatus: false,
		},
		"multiple-logical-ports-with-multiple-pages": {
			listLogicalPortsMockParams: []listLogicalPortsMockParam{{
				options: map[string]interface{}{"cursor": ""},
				result: manager.LogicalPortListResult{
					Results: []manager.LogicalPort{{
						Id:          "id01",
						DisplayName: "name01",
					}},
					Cursor: "next-page",
				},
				error: nil,
			}, {
				options: map[string]interface{}{"cursor": "next-page"},
				result: manager.LogicalPortListResult{
					Results: []manager.LogicalPort{{
						Id:          "id02",
						DisplayName: "name02",
					}},
					Cursor: "",
				},
				error: nil,
			}},
			getLogicalPortOperationalStatusMockParams: []getLogicalPortOperationalStatusParam{{
				id: "id01",
				result: manager.LogicalPortOperationalStatus{
					Status: "UP",
				},
				error: nil,
			}, {
				id: "id02",
				result: manager.LogicalPortOperationalStatus{
					Status: "DOWN",
				},
				error: nil,
			}},
			expectedResult: []generator.LogicalPortStatus{{
				Status: 1,
				ID:     "id01",
				Name:   "name01",
			}, {
				Status: 0,
				ID:     "id02",
				Name:   "name02",
			}},
			expectedStatus: true,
		},
	}

	for _, test := range tests {
		ctx := logicalPortMetricGeneratorTestContext{}
		ctx.setUp()
		for _, param := range test.listLogicalPortsMockParams {
			ctx.nsxtClientMock.
				On("ListLogicalPorts", param.options).
				Return(param.result, param.error).
				Once()
		}
		for _, param := range test.getLogicalPortOperationalStatusMockParams {
			ctx.nsxtClientMock.
				On("GetLogicalPortOperationalStatus", param.id, mock.Anything).
				Return(param.result, param.error).
				Once()
		}
		metrics, ok := ctx.lportMetricGenerator.GenerateLogicalPortStatusMetrics()
		assert.Equal(t, test.expectedResult, metrics)
		assert.Equal(t, test.expectedStatus, ok)
		ctx.nsxtClientMock.AssertExpectations(t)
	}
}
