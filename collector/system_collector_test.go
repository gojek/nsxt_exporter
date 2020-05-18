package collector

import (
	"errors"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/go-vmware-nsxt/administration"
)

const (
	fakeClusterNodeIPAddress = "1.2.3.4"
)

type mockSystemClient struct {
	clusterStatusResponse        mockClusterStatusResponse
	clusterNodeStatusResponse    mockClusterNodeStatusResponse
	clusterServiceStatusResponse mockClusterServiceStatusResponse
}

type mockClusterStatusResponse struct {
	ControllerStatus string
	ManagementStatus string
	Error            error
}

type mockClusterNodeStatusResponse struct {
	ControlClusterStatus    []mockControlClusterStatus
	ManagementClusterStatus []string
	Error                   error
}

type mockClusterServiceStatusResponse struct {
	ServiceStatus string
	Error         error
}

type mockControlClusterStatus struct {
	Status                 string
	MgmtConnectivityStatus string
}

func (c *mockSystemClient) ReadClusterStatus() (administration.ClusterStatus, error) {
	if c.clusterStatusResponse.Error != nil {
		return administration.ClusterStatus{}, c.clusterStatusResponse.Error
	}
	return administration.ClusterStatus{
		ControlClusterStatus: &administration.ControllerClusterStatus{
			Status: c.clusterStatusResponse.ControllerStatus,
		},
		MgmtClusterStatus: &administration.ManagementClusterStatus{
			Status: c.clusterStatusResponse.ManagementStatus,
		},
	}, nil
}

func (c *mockSystemClient) ReadClusterNodesAggregateStatus() (administration.ClustersAggregateInfo, error) {
	if c.clusterNodeStatusResponse.Error != nil {
		return administration.ClustersAggregateInfo{}, c.clusterNodeStatusResponse.Error
	}
	var controllerStatus []administration.ControllerNodeAggregateInfo
	for _, cs := range c.clusterNodeStatusResponse.ControlClusterStatus {
		controller := administration.ControllerNodeAggregateInfo{
			RoleConfig: &administration.ControllerClusterRoleConfig{
				ControlPlaneListenAddr: &administration.ServiceEndpoint{
					IpAddress: fakeClusterNodeIPAddress,
				},
			},
			NodeStatus: &administration.ClusterNodeStatus{
				ControlClusterStatus: &administration.ControlClusterNodeStatus{
					ControlClusterStatus: cs.Status,
					MgmtConnectionStatus: &administration.MgmtConnStatus{
						ConnectivityStatus: cs.MgmtConnectivityStatus,
					},
				},
			},
		}
		controllerStatus = append(controllerStatus, controller)
	}

	var managementStatus []administration.ManagementNodeAggregateInfo
	for _, ms := range c.clusterNodeStatusResponse.ManagementClusterStatus {
		management := administration.ManagementNodeAggregateInfo{
			RoleConfig: &administration.ManagementClusterRoleConfig{
				MgmtPlaneListenAddr: &administration.ServiceEndpoint{
					IpAddress: fakeClusterNodeIPAddress,
				},
			},
			NodeStatus: &administration.ClusterNodeStatus{
				MgmtClusterStatus: &administration.ManagementClusterNodeStatus{
					MgmtClusterStatus: ms,
				},
			},
		}
		managementStatus = append(managementStatus, management)
	}

	return administration.ClustersAggregateInfo{
		ControllerCluster: controllerStatus,
		ManagementCluster: managementStatus,
	}, nil
}

func (c *mockSystemClient) buildServiceStatusResponse() (administration.NodeServiceStatusProperties, error) {
	if c.clusterServiceStatusResponse.Error != nil {
		return administration.NodeServiceStatusProperties{}, c.clusterServiceStatusResponse.Error
	}
	return administration.NodeServiceStatusProperties{
		RuntimeState: c.clusterServiceStatusResponse.ServiceStatus,
	}, nil
}

func (c *mockSystemClient) ReadApplianceManagementServiceStatus() (administration.NodeServiceStatusProperties, error) {
	return c.buildServiceStatusResponse()
}

func (c *mockSystemClient) ReadNSXMessageBusServiceStatus() (administration.NodeServiceStatusProperties, error) {
	return c.buildServiceStatusResponse()
}

func (c *mockSystemClient) ReadNTPServiceStatus() (administration.NodeServiceStatusProperties, error) {
	return c.buildServiceStatusResponse()
}

func (c *mockSystemClient) ReadNsxUpgradeAgentServiceStatus() (administration.NodeServiceStatusProperties, error) {
	return c.buildServiceStatusResponse()
}

func (c *mockSystemClient) ReadProtonServiceStatus() (administration.NodeServiceStatusProperties, error) {
	return c.buildServiceStatusResponse()
}

func (c *mockSystemClient) ReadProxyServiceStatus() (administration.NodeServiceStatusProperties, error) {
	return c.buildServiceStatusResponse()
}

func (c *mockSystemClient) ReadRabbitMQServiceStatus() (administration.NodeServiceStatusProperties, error) {
	return c.buildServiceStatusResponse()
}

func (c *mockSystemClient) ReadRepositoryServiceStatus() (administration.NodeServiceStatusProperties, error) {
	return c.buildServiceStatusResponse()
}

func (c *mockSystemClient) ReadSNMPServiceStatus() (administration.NodeServiceStatusProperties, error) {
	return c.buildServiceStatusResponse()
}

func (c *mockSystemClient) ReadSSHServiceStatus() (administration.NodeServiceStatusProperties, error) {
	return c.buildServiceStatusResponse()
}

func (c *mockSystemClient) ReadSearchServiceStatus() (administration.NodeServiceStatusProperties, error) {
	return c.buildServiceStatusResponse()
}

func (c *mockSystemClient) ReadSyslogServiceStatus() (administration.NodeServiceStatusProperties, error) {
	return c.buildServiceStatusResponse()
}

func TestSystemCollector_CollectClusterStatusMetrics(t *testing.T) {
	testcases := []struct {
		description     string
		response        mockClusterStatusResponse
		expectedMetrics []systemStatusMetric
	}{
		{
			description: "Should return up value when both controller and management stable",
			response: mockClusterStatusResponse{
				ControllerStatus: "STABLE",
				ManagementStatus: "STABLE",
				Error:            nil,
			},
			expectedMetrics: []systemStatusMetric{
				{
					Status: 1.0,
				},
			},
		},
		{
			description: "Should return up value when stable with mixed cases",
			response: mockClusterStatusResponse{
				ControllerStatus: "Stable",
				ManagementStatus: "sTaBLe",
				Error:            nil,
			},
			expectedMetrics: []systemStatusMetric{
				{
					Status: 1.0,
				},
			},
		},
		{
			description: "Should return down value when controller is unstable",
			response: mockClusterStatusResponse{
				ControllerStatus: "UNSTABLE",
				ManagementStatus: "STABLE",
				Error:            nil,
			},
			expectedMetrics: []systemStatusMetric{
				{
					Status: 0.0,
				},
			},
		},
		{
			description: "Should return down value when management is unstable",
			response: mockClusterStatusResponse{
				ControllerStatus: "STABLE",
				ManagementStatus: "UNSTABLE",
				Error:            nil,
			},
			expectedMetrics: []systemStatusMetric{
				{
					Status: 0.0,
				},
			},
		},
		{
			description: "Should return down value when both are unstable",
			response: mockClusterStatusResponse{
				ControllerStatus: "UNSTABLE",
				ManagementStatus: "UNSTABLE",
				Error:            nil,
			},
			expectedMetrics: []systemStatusMetric{
				{
					Status: 0.0,
				},
			},
		},
		{
			description: "Should return down value when both are degraded",
			response: mockClusterStatusResponse{
				ControllerStatus: "DEGRADED",
				ManagementStatus: "DEGRADED",
				Error:            nil,
			},
			expectedMetrics: []systemStatusMetric{
				{
					Status: 0.0,
				},
			},
		},
		{
			description: "Should return down value when both are unknown",
			response: mockClusterStatusResponse{
				ControllerStatus: "UNKNOWN",
				ManagementStatus: "UNKNOWN",
				Error:            nil,
			},
			expectedMetrics: []systemStatusMetric{
				{
					Status: 0.0,
				},
			},
		},
		{
			description: "Should return down value when both are no controllers",
			response: mockClusterStatusResponse{
				ControllerStatus: "NO_CONTROLLERS",
				ManagementStatus: "NO_CONTROLLERS",
				Error:            nil,
			},
			expectedMetrics: []systemStatusMetric{
				{
					Status: 0.0,
				},
			},
		},
		{
			description: "Should return empty metrics value when error retrieving response",
			response: mockClusterStatusResponse{
				ControllerStatus: "STABLE",
				ManagementStatus: "STABLE",
				Error:            errors.New("error read cluster status"),
			},
			expectedMetrics: []systemStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockSystemClient := &mockSystemClient{
			clusterStatusResponse: tc.response,
		}
		logger := log.NewNopLogger()
		systemCollector := newSystemCollector(mockSystemClient, logger)
		clusterMetrics := systemCollector.collectClusterStatusMetrics()
		assert.ElementsMatch(t, tc.expectedMetrics, clusterMetrics, tc.description)
	}
}

func TestSystemCollector_CollectClusterNodeMetrics(t *testing.T) {
	testcases := []struct {
		description     string
		response        mockClusterNodeStatusResponse
		expectedMetrics []systemStatusMetric
	}{
		{
			description: "Should return up value for connected nodes",
			response: mockClusterNodeStatusResponse{
				ControlClusterStatus: []mockControlClusterStatus{
					{
						Status:                 "CONNECTED",
						MgmtConnectivityStatus: "CONNECTED",
					},
					{
						Status:                 "Connected",
						MgmtConnectivityStatus: "connEcTed",
					},
				},
				ManagementClusterStatus: []string{"CONNECTED", "ConNected"},
				Error:                   nil,
			},
			expectedMetrics: []systemStatusMetric{
				{
					IPAddress: fakeClusterNodeIPAddress,
					Type:      "controller",
					Status:    1.0,
				},
				{
					IPAddress: fakeClusterNodeIPAddress,
					Type:      "controller",
					Status:    1.0,
				},
				{
					IPAddress: fakeClusterNodeIPAddress,
					Type:      "management",
					Status:    1.0,
				},
				{
					IPAddress: fakeClusterNodeIPAddress,
					Type:      "management",
					Status:    1.0,
				},
			},
		},
		{
			description: "Should return down value for disconnected nodes",
			response: mockClusterNodeStatusResponse{
				ControlClusterStatus: []mockControlClusterStatus{
					{
						Status:                 "DISCONNECTED",
						MgmtConnectivityStatus: "DISCONNECTED",
					},
					{
						Status:                 "UNKNOWN",
						MgmtConnectivityStatus: "UNKNOWN",
					},
					{
						Status:                 "CONNECTED",
						MgmtConnectivityStatus: "DISCONNECTED",
					},
					{
						Status:                 "DISCONNECTED",
						MgmtConnectivityStatus: "CONNECTED",
					},
				},
				ManagementClusterStatus: []string{"DISCONNECTED", "UNKNOWN"},
				Error:                   nil,
			},
			expectedMetrics: []systemStatusMetric{
				{
					IPAddress: fakeClusterNodeIPAddress,
					Type:      "controller",
					Status:    0.0,
				},
				{
					IPAddress: fakeClusterNodeIPAddress,
					Type:      "controller",
					Status:    0.0,
				},
				{
					IPAddress: fakeClusterNodeIPAddress,
					Type:      "controller",
					Status:    0.0,
				},
				{
					IPAddress: fakeClusterNodeIPAddress,
					Type:      "controller",
					Status:    0.0,
				},
				{
					IPAddress: fakeClusterNodeIPAddress,
					Type:      "management",
					Status:    0.0,
				},
				{
					IPAddress: fakeClusterNodeIPAddress,
					Type:      "management",
					Status:    0.0,
				},
			},
		},
		{
			description: "Should return empty metrics value when error retrieving response",
			response: mockClusterNodeStatusResponse{
				ControlClusterStatus: []mockControlClusterStatus{
					{
						Status:                 "CONNECTED",
						MgmtConnectivityStatus: "CONNECTED",
					},
				},
				ManagementClusterStatus: []string{"CONNECTED"},
				Error:                   errors.New("error read cluster node status"),
			},
			expectedMetrics: []systemStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockSystemClient := &mockSystemClient{
			clusterNodeStatusResponse: tc.response,
		}
		logger := log.NewNopLogger()
		systemCollector := newSystemCollector(mockSystemClient, logger)
		nodeMetrics := systemCollector.collectClusterNodeMetrics()
		assert.ElementsMatch(t, tc.expectedMetrics, nodeMetrics, tc.description)
	}
}

func TestSystemCollector_CollectApplianceServiceMetric(t *testing.T) {
	testcases := []struct {
		description    string
		response       mockClusterServiceStatusResponse
		expectedMetric systemStatusMetric
	}{
		{
			description:    "Should return up value when appliance service is running",
			response:       mockClusterServiceStatusResponse{"RUNNING", nil},
			expectedMetric: systemStatusMetric{Name: "appliance", Status: 1.0},
		},
		{
			description:    "Should return up value when appliance service is running with mixed cases",
			response:       mockClusterServiceStatusResponse{"Running", nil},
			expectedMetric: systemStatusMetric{Name: "appliance", Status: 1.0},
		},
		{
			description:    "Should return down value when appliance service is not running",
			response:       mockClusterServiceStatusResponse{"STOPPED", nil},
			expectedMetric: systemStatusMetric{Name: "appliance", Status: 0.0},
		},
		{
			description:    "Should return empty when failed reading appliance service state",
			response:       mockClusterServiceStatusResponse{"RUNNING", errors.New("error read state")},
			expectedMetric: systemStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockSystemClient := &mockSystemClient{
			clusterServiceStatusResponse: tc.response,
		}
		logger := log.NewNopLogger()
		systemCollector := newSystemCollector(mockSystemClient, logger)
		serviceMetric, err := systemCollector.collectApplianceServiceMetric()
		assert.Equal(t, tc.expectedMetric, serviceMetric, tc.description)
		if tc.expectedMetric == (systemStatusMetric{}) {
			assert.Error(t, err)
		}
	}
}

func TestSystemCollector_CollectMessageBusServiceMetric(t *testing.T) {
	testcases := []struct {
		description    string
		response       mockClusterServiceStatusResponse
		expectedMetric systemStatusMetric
	}{
		{
			description:    "Should return up value when message bus service is running",
			response:       mockClusterServiceStatusResponse{"RUNNING", nil},
			expectedMetric: systemStatusMetric{Name: "message_bus", Status: 1.0},
		},
		{
			description:    "Should return up value when message bus service is running with mixed cases",
			response:       mockClusterServiceStatusResponse{"Running", nil},
			expectedMetric: systemStatusMetric{Name: "message_bus", Status: 1.0},
		},
		{
			description:    "Should return down value when message bus service is not running",
			response:       mockClusterServiceStatusResponse{"STOPPED", nil},
			expectedMetric: systemStatusMetric{Name: "message_bus", Status: 0.0},
		},
		{
			description:    "Should return empty when failed reading message bus service state",
			response:       mockClusterServiceStatusResponse{"RUNNING", errors.New("error read state")},
			expectedMetric: systemStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockSystemClient := &mockSystemClient{
			clusterServiceStatusResponse: tc.response,
		}
		logger := log.NewNopLogger()
		systemCollector := newSystemCollector(mockSystemClient, logger)
		serviceMetric, err := systemCollector.collectMessageBusServiceMetric()
		assert.Equal(t, tc.expectedMetric, serviceMetric, tc.description)
		if tc.expectedMetric == (systemStatusMetric{}) {
			assert.Error(t, err)
		}
	}
}

func TestSystemCollector_CollectNTPServiceMetric(t *testing.T) {
	testcases := []struct {
		description    string
		response       mockClusterServiceStatusResponse
		expectedMetric systemStatusMetric
	}{
		{
			description:    "Should return up value when ntp service is running",
			response:       mockClusterServiceStatusResponse{"RUNNING", nil},
			expectedMetric: systemStatusMetric{Name: "ntp", Status: 1.0},
		},
		{
			description:    "Should return up value when ntp service is running with mixed cases",
			response:       mockClusterServiceStatusResponse{"Running", nil},
			expectedMetric: systemStatusMetric{Name: "ntp", Status: 1.0},
		},
		{
			description:    "Should return down value when ntp service is not running",
			response:       mockClusterServiceStatusResponse{"STOPPED", nil},
			expectedMetric: systemStatusMetric{Name: "ntp", Status: 0.0},
		},
		{
			description:    "Should return empty when failed reading ntp service state",
			response:       mockClusterServiceStatusResponse{"RUNNING", errors.New("error read state")},
			expectedMetric: systemStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockSystemClient := &mockSystemClient{
			clusterServiceStatusResponse: tc.response,
		}
		logger := log.NewNopLogger()
		systemCollector := newSystemCollector(mockSystemClient, logger)
		serviceMetric, err := systemCollector.collectNTPServiceMetric()
		assert.Equal(t, tc.expectedMetric, serviceMetric, tc.description)
		if tc.expectedMetric == (systemStatusMetric{}) {
			assert.Error(t, err)
		}
	}
}

func TestSystemCollector_CollectUpgradeServiceMetric(t *testing.T) {
	testcases := []struct {
		description    string
		response       mockClusterServiceStatusResponse
		expectedMetric systemStatusMetric
	}{
		{
			description:    "Should return up value when upgrade agent service is running",
			response:       mockClusterServiceStatusResponse{"RUNNING", nil},
			expectedMetric: systemStatusMetric{Name: "upgrade_agent", Status: 1.0},
		},
		{
			description:    "Should return up value when upgrade agent service is running with mixed cases",
			response:       mockClusterServiceStatusResponse{"Running", nil},
			expectedMetric: systemStatusMetric{Name: "upgrade_agent", Status: 1.0},
		},
		{
			description:    "Should return down value when upgrade agent service is not running",
			response:       mockClusterServiceStatusResponse{"STOPPED", nil},
			expectedMetric: systemStatusMetric{Name: "upgrade_agent", Status: 0.0},
		},
		{
			description:    "Should return empty when failed reading upgrade agent service state",
			response:       mockClusterServiceStatusResponse{"RUNNING", errors.New("error read state")},
			expectedMetric: systemStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockSystemClient := &mockSystemClient{
			clusterServiceStatusResponse: tc.response,
		}
		logger := log.NewNopLogger()
		systemCollector := newSystemCollector(mockSystemClient, logger)
		serviceMetric, err := systemCollector.collectUpgradeAgentServiceMetric()
		assert.Equal(t, tc.expectedMetric, serviceMetric, tc.description)
		if tc.expectedMetric == (systemStatusMetric{}) {
			assert.Error(t, err)
		}
	}
}

func TestSystemCollector_CollectProtonServiceMetric(t *testing.T) {
	testcases := []struct {
		description    string
		response       mockClusterServiceStatusResponse
		expectedMetric systemStatusMetric
	}{
		{
			description:    "Should return up value when proton service is running",
			response:       mockClusterServiceStatusResponse{"RUNNING", nil},
			expectedMetric: systemStatusMetric{Name: "proton", Status: 1.0},
		},
		{
			description:    "Should return up value when proton service is running with mixed cases",
			response:       mockClusterServiceStatusResponse{"Running", nil},
			expectedMetric: systemStatusMetric{Name: "proton", Status: 1.0},
		},
		{
			description:    "Should return down value when proton service is not running",
			response:       mockClusterServiceStatusResponse{"STOPPED", nil},
			expectedMetric: systemStatusMetric{Name: "proton", Status: 0.0},
		},
		{
			description:    "Should return empty when failed reading proton service state",
			response:       mockClusterServiceStatusResponse{"RUNNING", errors.New("error read state")},
			expectedMetric: systemStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockSystemClient := &mockSystemClient{
			clusterServiceStatusResponse: tc.response,
		}
		logger := log.NewNopLogger()
		systemCollector := newSystemCollector(mockSystemClient, logger)
		serviceMetric, err := systemCollector.collectProtonServiceMetric()
		assert.Equal(t, tc.expectedMetric, serviceMetric, tc.description)
		if tc.expectedMetric == (systemStatusMetric{}) {
			assert.Error(t, err)
		}
	}
}

func TestSystemCollector_CollectProxyServiceMetric(t *testing.T) {
	testcases := []struct {
		description    string
		response       mockClusterServiceStatusResponse
		expectedMetric systemStatusMetric
	}{
		{
			description:    "Should return up value when proxy service is running",
			response:       mockClusterServiceStatusResponse{"RUNNING", nil},
			expectedMetric: systemStatusMetric{Name: "proxy", Status: 1.0},
		},
		{
			description:    "Should return up value when proxy service is running with mixed cases",
			response:       mockClusterServiceStatusResponse{"Running", nil},
			expectedMetric: systemStatusMetric{Name: "proxy", Status: 1.0},
		},
		{
			description:    "Should return down value when proxy service is not running",
			response:       mockClusterServiceStatusResponse{"STOPPED", nil},
			expectedMetric: systemStatusMetric{Name: "proxy", Status: 0.0},
		},
		{
			description:    "Should return empty when failed reading proxy service state",
			response:       mockClusterServiceStatusResponse{"RUNNING", errors.New("error read state")},
			expectedMetric: systemStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockSystemClient := &mockSystemClient{
			clusterServiceStatusResponse: tc.response,
		}
		logger := log.NewNopLogger()
		systemCollector := newSystemCollector(mockSystemClient, logger)
		serviceMetric, err := systemCollector.collectProxyServiceMetric()
		assert.Equal(t, tc.expectedMetric, serviceMetric, tc.description)
		if tc.expectedMetric == (systemStatusMetric{}) {
			assert.Error(t, err)
		}
	}
}

func TestSystemCollector_CollectRabbitMQServiceMetric(t *testing.T) {
	testcases := []struct {
		description    string
		response       mockClusterServiceStatusResponse
		expectedMetric systemStatusMetric
	}{
		{
			description:    "Should return up value when rabbitmq service is running",
			response:       mockClusterServiceStatusResponse{"RUNNING", nil},
			expectedMetric: systemStatusMetric{Name: "rabbitmq", Status: 1.0},
		},
		{
			description:    "Should return up value when rabbitmq service is running with mixed cases",
			response:       mockClusterServiceStatusResponse{"Running", nil},
			expectedMetric: systemStatusMetric{Name: "rabbitmq", Status: 1.0},
		},
		{
			description:    "Should return down value when rabbitmq service is not running",
			response:       mockClusterServiceStatusResponse{"STOPPED", nil},
			expectedMetric: systemStatusMetric{Name: "rabbitmq", Status: 0.0},
		},
		{
			description:    "Should return empty when failed reading rabbitmq service state",
			response:       mockClusterServiceStatusResponse{"RUNNING", errors.New("error read state")},
			expectedMetric: systemStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockSystemClient := &mockSystemClient{
			clusterServiceStatusResponse: tc.response,
		}
		logger := log.NewNopLogger()
		systemCollector := newSystemCollector(mockSystemClient, logger)
		serviceMetric, err := systemCollector.collectRabbitMQServiceMetric()
		assert.Equal(t, tc.expectedMetric, serviceMetric, tc.description)
		if tc.expectedMetric == (systemStatusMetric{}) {
			assert.Error(t, err)
		}
	}
}

func TestSystemCollector_CollectRepositoryServiceMetric(t *testing.T) {
	testcases := []struct {
		description    string
		response       mockClusterServiceStatusResponse
		expectedMetric systemStatusMetric
	}{
		{
			description:    "Should return up value when repository service is running",
			response:       mockClusterServiceStatusResponse{"RUNNING", nil},
			expectedMetric: systemStatusMetric{Name: "repository", Status: 1.0},
		},
		{
			description:    "Should return up value when repository service is running with mixed cases",
			response:       mockClusterServiceStatusResponse{"Running", nil},
			expectedMetric: systemStatusMetric{Name: "repository", Status: 1.0},
		},
		{
			description:    "Should return down value when repository service is not running",
			response:       mockClusterServiceStatusResponse{"STOPPED", nil},
			expectedMetric: systemStatusMetric{Name: "repository", Status: 0.0},
		},
		{
			description:    "Should return empty when failed reading repository service state",
			response:       mockClusterServiceStatusResponse{"RUNNING", errors.New("error read state")},
			expectedMetric: systemStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockSystemClient := &mockSystemClient{
			clusterServiceStatusResponse: tc.response,
		}
		logger := log.NewNopLogger()
		systemCollector := newSystemCollector(mockSystemClient, logger)
		serviceMetric, err := systemCollector.collectRepositoryServiceMetric()
		assert.Equal(t, tc.expectedMetric, serviceMetric, tc.description)
		if tc.expectedMetric == (systemStatusMetric{}) {
			assert.Error(t, err)
		}
	}
}

func TestSystemCollector_CollectSNMPServiceMetric(t *testing.T) {
	testcases := []struct {
		description    string
		response       mockClusterServiceStatusResponse
		expectedMetric systemStatusMetric
	}{
		{
			description:    "Should return up value when snmp service is running",
			response:       mockClusterServiceStatusResponse{"RUNNING", nil},
			expectedMetric: systemStatusMetric{Name: "snmp", Status: 1.0},
		},
		{
			description:    "Should return up value when snmp service is running with mixed cases",
			response:       mockClusterServiceStatusResponse{"Running", nil},
			expectedMetric: systemStatusMetric{Name: "snmp", Status: 1.0},
		},
		{
			description:    "Should return down value when snmp service is not running",
			response:       mockClusterServiceStatusResponse{"STOPPED", nil},
			expectedMetric: systemStatusMetric{Name: "snmp", Status: 0.0},
		},
		{
			description:    "Should return empty when failed reading snmp service state",
			response:       mockClusterServiceStatusResponse{"RUNNING", errors.New("error read state")},
			expectedMetric: systemStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockSystemClient := &mockSystemClient{
			clusterServiceStatusResponse: tc.response,
		}
		logger := log.NewNopLogger()
		systemCollector := newSystemCollector(mockSystemClient, logger)
		serviceMetric, err := systemCollector.collectSNMPServiceMetric()
		assert.Equal(t, tc.expectedMetric, serviceMetric, tc.description)
		if tc.expectedMetric == (systemStatusMetric{}) {
			assert.Error(t, err)
		}
	}
}

func TestSystemCollector_CollectSSHServiceMetric(t *testing.T) {
	testcases := []struct {
		description    string
		response       mockClusterServiceStatusResponse
		expectedMetric systemStatusMetric
	}{
		{
			description:    "Should return up value when ssh service is running",
			response:       mockClusterServiceStatusResponse{"RUNNING", nil},
			expectedMetric: systemStatusMetric{Name: "ssh", Status: 1.0},
		},
		{
			description:    "Should return up value when ssh service is running with mixed cases",
			response:       mockClusterServiceStatusResponse{"Running", nil},
			expectedMetric: systemStatusMetric{Name: "ssh", Status: 1.0},
		},
		{
			description:    "Should return down value when ssh service is not running",
			response:       mockClusterServiceStatusResponse{"STOPPED", nil},
			expectedMetric: systemStatusMetric{Name: "ssh", Status: 0.0},
		},
		{
			description:    "Should return empty when failed reading ssh service state",
			response:       mockClusterServiceStatusResponse{"RUNNING", errors.New("error read state")},
			expectedMetric: systemStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockSystemClient := &mockSystemClient{
			clusterServiceStatusResponse: tc.response,
		}
		logger := log.NewNopLogger()
		systemCollector := newSystemCollector(mockSystemClient, logger)
		serviceMetric, err := systemCollector.collectSSHServiceMetric()
		assert.Equal(t, tc.expectedMetric, serviceMetric, tc.description)
		if tc.expectedMetric == (systemStatusMetric{}) {
			assert.Error(t, err)
		}
	}
}

func TestSystemCollector_CollectSearchServiceMetric(t *testing.T) {
	testcases := []struct {
		description    string
		response       mockClusterServiceStatusResponse
		expectedMetric systemStatusMetric
	}{
		{
			description:    "Should return up value when search service is running",
			response:       mockClusterServiceStatusResponse{"RUNNING", nil},
			expectedMetric: systemStatusMetric{Name: "search", Status: 1.0},
		},
		{
			description:    "Should return up value when search service is running with mixed cases",
			response:       mockClusterServiceStatusResponse{"Running", nil},
			expectedMetric: systemStatusMetric{Name: "search", Status: 1.0},
		},
		{
			description:    "Should return down value when search service is not running",
			response:       mockClusterServiceStatusResponse{"STOPPED", nil},
			expectedMetric: systemStatusMetric{Name: "search", Status: 0.0},
		},
		{
			description:    "Should return empty when failed reading search service state",
			response:       mockClusterServiceStatusResponse{"RUNNING", errors.New("error read state")},
			expectedMetric: systemStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockSystemClient := &mockSystemClient{
			clusterServiceStatusResponse: tc.response,
		}
		logger := log.NewNopLogger()
		systemCollector := newSystemCollector(mockSystemClient, logger)
		serviceMetric, err := systemCollector.collectSearchServiceMetric()
		assert.Equal(t, tc.expectedMetric, serviceMetric, tc.description)
		if tc.expectedMetric == (systemStatusMetric{}) {
			assert.Error(t, err)
		}
	}
}

func TestSystemCollector_CollectSyslogServiceMetric(t *testing.T) {
	testcases := []struct {
		description    string
		response       mockClusterServiceStatusResponse
		expectedMetric systemStatusMetric
	}{
		{
			description:    "Should return up value when syslog service is running",
			response:       mockClusterServiceStatusResponse{"RUNNING", nil},
			expectedMetric: systemStatusMetric{Name: "syslog", Status: 1.0},
		},
		{
			description:    "Should return up value when syslog service is running with mixed cases",
			response:       mockClusterServiceStatusResponse{"Running", nil},
			expectedMetric: systemStatusMetric{Name: "syslog", Status: 1.0},
		},
		{
			description:    "Should return down value when syslog service is not running",
			response:       mockClusterServiceStatusResponse{"STOPPED", nil},
			expectedMetric: systemStatusMetric{Name: "syslog", Status: 0.0},
		},
		{
			description:    "Should return empty when failed reading syslog service state",
			response:       mockClusterServiceStatusResponse{"RUNNING", errors.New("error read state")},
			expectedMetric: systemStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockSystemClient := &mockSystemClient{
			clusterServiceStatusResponse: tc.response,
		}
		logger := log.NewNopLogger()
		systemCollector := newSystemCollector(mockSystemClient, logger)
		serviceMetric, err := systemCollector.collectSyslogServiceMetric()
		assert.Equal(t, tc.expectedMetric, serviceMetric, tc.description)
		if tc.expectedMetric == (systemStatusMetric{}) {
			assert.Error(t, err)
		}
	}
}
