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
	ApplianceStatus string
	Error           error
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

func (c *mockSystemClient) ReadApplianceManagementServiceStatus() (administration.NodeServiceStatusProperties, error) {
	if c.clusterServiceStatusResponse.Error != nil {
		return administration.NodeServiceStatusProperties{}, c.clusterServiceStatusResponse.Error
	}
	return administration.NodeServiceStatusProperties{
		RuntimeState: c.clusterServiceStatusResponse.ApplianceStatus,
	}, nil
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

func TestSystemCollector_CollectClusterServiceMetrics(t *testing.T) {
	testcases := []struct {
		description     string
		response        mockClusterServiceStatusResponse
		expectedMetrics []systemStatusMetric
	}{
		{
			description: "Should return up value when appliance service is running",
			response: mockClusterServiceStatusResponse{
				ApplianceStatus: "RUNNING",
				Error:           nil,
			},
			expectedMetrics: []systemStatusMetric{
				{
					Status: 1.0,
				},
			},
		},
		{
			description: "Should return up value when appliance service is running with mixed cases",
			response: mockClusterServiceStatusResponse{
				ApplianceStatus: "Running",
				Error:           nil,
			},
			expectedMetrics: []systemStatusMetric{
				{
					Status: 1.0,
				},
			},
		},
		{
			description: "Should return down value when appliance service is not running",
			response: mockClusterServiceStatusResponse{
				ApplianceStatus: "STOPPED",
				Error:           nil,
			},
			expectedMetrics: []systemStatusMetric{
				{
					Status: 0.0,
				},
			},
		},
	}
	for _, tc := range testcases {
		mockSystemClient := &mockSystemClient{
			clusterServiceStatusResponse: tc.response,
		}
		logger := log.NewNopLogger()
		systemCollector := newSystemCollector(mockSystemClient, logger)
		serviceMetrics := systemCollector.collectClusterServiceMetrics()
		assert.ElementsMatch(t, tc.expectedMetrics, serviceMetrics, tc.description)
	}
}
