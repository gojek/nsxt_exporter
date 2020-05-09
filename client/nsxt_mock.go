package client

import (
	"github.com/stretchr/testify/mock"
	"github.com/vmware/go-vmware-nsxt/manager"
)

type NSXTClientMock struct {
	mock.Mock
}

func (m *NSXTClientMock) GetLogicalPortStatusSummary(localVarOptionals map[string]interface{}) (manager.LogicalPortStatusSummary, error) {
	arguments := m.Called(localVarOptionals)
	return arguments.Get(0).(manager.LogicalPortStatusSummary), arguments.Error(1)
}

func (m *NSXTClientMock) ListLogicalPorts(localVarOptionals map[string]interface{}) (manager.LogicalPortListResult, error) {
	arguments := m.Called(localVarOptionals)
	return arguments.Get(0).(manager.LogicalPortListResult), arguments.Error(1)
}

func (m *NSXTClientMock) GetLogicalPortOperationalStatus(lportId string, localVarOptionals map[string]interface{}) (manager.LogicalPortOperationalStatus, error) {
	arguments := m.Called(lportId, localVarOptionals)
	return arguments.Get(0).(manager.LogicalPortOperationalStatus), arguments.Error(1)
}
