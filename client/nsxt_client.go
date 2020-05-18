package client

import (
	"github.com/go-kit/kit/log"
	nsxt "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/administration"
	"github.com/vmware/go-vmware-nsxt/manager"
)

type nsxtClient struct {
	apiClient *nsxt.APIClient
	logger    log.Logger
}

func NewNSXTClient(apiClient *nsxt.APIClient, logger log.Logger) *nsxtClient {
	return &nsxtClient{
		apiClient: apiClient,
		logger:    logger,
	}
}

func (c *nsxtClient) ListLogicalPorts(localVarOptionals map[string]interface{}) (manager.LogicalPortListResult, error) {
	lportsResult, _, err := c.apiClient.LogicalSwitchingApi.ListLogicalPorts(c.apiClient.Context, localVarOptionals)
	return lportsResult, err
}

func (c *nsxtClient) GetLogicalPortOperationalStatus(lportId string, localVarOptionals map[string]interface{}) (manager.LogicalPortOperationalStatus, error) {
	lportStatus, _, err := c.apiClient.LogicalSwitchingApi.GetLogicalPortOperationalStatus(c.apiClient.Context, lportId, localVarOptionals)
	return lportStatus, err
}

func (c *nsxtClient) ListAllLogicalRouterPorts() ([]manager.LogicalRouterPort, error) {
	var logicalRouterPorts []manager.LogicalRouterPort
	var cursor string
	for {
		localVarOptionals := make(map[string]interface{})
		localVarOptionals["cursor"] = cursor
		logicalRouterPortsResult, _, err := c.apiClient.LogicalRoutingAndServicesApi.ListLogicalRouterPorts(c.apiClient.Context, localVarOptionals)
		if err != nil {
			return nil, err
		}
		logicalRouterPorts = append(logicalRouterPorts, logicalRouterPortsResult.Results...)
		cursor = logicalRouterPortsResult.Cursor
		if len(cursor) == 0 {
			break
		}
	}
	return logicalRouterPorts, nil
}

func (c *nsxtClient) GetLogicalRouterPortStatisticsSummary(lrportID string) (manager.LogicalRouterPortStatisticsSummary, error) {
	lrportsStatus, _, err := c.apiClient.LogicalRoutingAndServicesApi.GetLogicalRouterPortStatisticsSummary(c.apiClient.Context, lrportID, nil)
	return lrportsStatus, err
}

func (c *nsxtClient) ListAllDHCPServers() ([]manager.LogicalDhcpServer, error) {
	var dhcps []manager.LogicalDhcpServer
	var cursor string
	for {
		localVarOptionals := make(map[string]interface{})
		localVarOptionals["cursor"] = cursor
		dhcpListResponse, _, err := c.apiClient.ServicesApi.ListDhcpServers(c.apiClient.Context, localVarOptionals)
		if err != nil {
			return nil, err
		}
		dhcps = append(dhcps, dhcpListResponse.Results...)
		cursor = dhcpListResponse.Cursor
		if len(cursor) == 0 {
			break
		}
	}
	return dhcps, nil
}

func (c *nsxtClient) GetDhcpStatus(dhcpID string, localVarOptionals map[string]interface{}) (manager.DhcpServerStatus, error) {
	dhcpServerStatus, _, err := c.apiClient.ServicesApi.GetDhcpStatus(c.apiClient.Context, dhcpID)
	return dhcpServerStatus, err
}

func (c *nsxtClient) GetDHCPStatistic(dhcpID string) (manager.DhcpStatistics, error) {
	dhcpServerStatistic, _, err := c.apiClient.ServicesApi.GetDhcpStatistics(c.apiClient.Context, dhcpID)
	return dhcpServerStatistic, err
}

func (c *nsxtClient) ListAllTransportNodes() ([]manager.TransportNode, error) {
	var transportNodes []manager.TransportNode
	var cursor string
	for {
		localVarOptionals := make(map[string]interface{})
		localVarOptionals["cursor"] = cursor
		transportNodesResult, _, err := c.apiClient.NetworkTransportApi.ListTransportNodes(c.apiClient.Context, localVarOptionals)
		if err != nil {
			return nil, err
		}
		transportNodes = append(transportNodes, transportNodesResult.Results...)
		cursor = transportNodesResult.Cursor
		if len(cursor) == 0 {
			break
		}
	}
	return transportNodes, nil
}

func (c *nsxtClient) GetTransportNodeStatus(nodeID string) (manager.TransportNodeStatus, error) {
	transportNodeStatus, _, err := c.apiClient.TroubleshootingAndMonitoringApi.GetTransportNodeStatus(c.apiClient.Context, nodeID, nil)
	return transportNodeStatus, err
}

func (c *nsxtClient) ListAllEdgeClusters() ([]manager.EdgeCluster, error) {
	var edgeClusters []manager.EdgeCluster
	var cursor string
	for {
		localVarOptionals := make(map[string]interface{})
		localVarOptionals["cursor"] = cursor
		res, _, err := c.apiClient.NetworkTransportApi.ListEdgeClusters(c.apiClient.Context, localVarOptionals)
		if err != nil {
			return nil, err
		}
		edgeClusters = append(edgeClusters, res.Results...)
		cursor = res.Cursor
		if len(cursor) == 0 {
			break
		}
	}
	return edgeClusters, nil
}

func (c *nsxtClient) ReadClusterStatus() (administration.ClusterStatus, error) {
	clusterStatus, _, err := c.apiClient.NsxComponentAdministrationApi.ReadClusterStatus(c.apiClient.Context, nil)
	return clusterStatus, err
}

func (c *nsxtClient) ReadClusterNodesAggregateStatus() (administration.ClustersAggregateInfo, error) {
	clusterNodesStatus, _, err := c.apiClient.NsxComponentAdministrationApi.ReadClusterNodesAggregateStatus(c.apiClient.Context)
	return clusterNodesStatus, err
}

func (c *nsxtClient) ReadApplianceManagementServiceStatus() (administration.NodeServiceStatusProperties, error) {
	applianceServiceStatus, _, err := c.apiClient.NsxComponentAdministrationApi.ReadApplianceManagementServiceStatus(c.apiClient.Context)
	return applianceServiceStatus, err
}

func (c *nsxtClient) ListAllLogicalSwitches() ([]manager.LogicalSwitch, error) {
	var logicalSwitches []manager.LogicalSwitch
	var cursor string
	for {
		localVarOptionals := make(map[string]interface{})
		localVarOptionals["cursor"] = cursor
		logicalSwitchListResult, _, err := c.apiClient.LogicalSwitchingApi.ListLogicalSwitches(c.apiClient.Context, localVarOptionals)
		if err != nil {
			return nil, err
		}
		logicalSwitches = append(logicalSwitches, logicalSwitchListResult.Results...)
		cursor = logicalSwitchListResult.Cursor
		if len(cursor) == 0 {
			break
		}
	}
	return logicalSwitches, nil
}

func (c *nsxtClient) GetLogicalSwitchState(lswitchID string) (manager.LogicalSwitchState, error) {
	logicalSwitchesStatus, _, err := c.apiClient.LogicalSwitchingApi.GetLogicalSwitchState(c.apiClient.Context, lswitchID)
	return logicalSwitchesStatus, err
}

func (c *nsxtClient) GetLogicalSwitchStatistic(lswitchID string) (manager.LogicalSwitchStatistics, error) {
	logicalSwitchStatistic, _, err := c.apiClient.LogicalSwitchingApi.GetLogicalSwitchStatistics(c.apiClient.Context, lswitchID, nil)
	return logicalSwitchStatistic, err
}
