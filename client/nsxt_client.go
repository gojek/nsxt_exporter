package client

import (
	"github.com/go-kit/kit/log"
	nsxt "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/administration"
	"github.com/vmware/go-vmware-nsxt/manager"
	"github.com/vmware/go-vmware-nsxt/monitoring"
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

func (c *nsxtClient) GetLogicalPortStatusSummary(localVarOptionals map[string]interface{}) (manager.LogicalPortStatusSummary, error) {
	lportStatus, _, err := c.apiClient.LogicalSwitchingApi.GetLogicalPortStatusSummary(c.apiClient.Context, localVarOptionals)
	return lportStatus, err
}

func (c *nsxtClient) ListLogicalPorts(localVarOptionals map[string]interface{}) (manager.LogicalPortListResult, error) {
	lportsResult, _, err := c.apiClient.LogicalSwitchingApi.ListLogicalPorts(c.apiClient.Context, localVarOptionals)
	return lportsResult, err
}

func (c *nsxtClient) GetLogicalPortOperationalStatus(lportId string, localVarOptionals map[string]interface{}) (manager.LogicalPortOperationalStatus, error) {
	lportStatus, _, err := c.apiClient.LogicalSwitchingApi.GetLogicalPortOperationalStatus(c.apiClient.Context, lportId, localVarOptionals)
	return lportStatus, err
}

func (c *nsxtClient) ListLogicalRouterPorts(localVarOptionals map[string]interface{}) (manager.LogicalRouterPortListResult, error) {
	lroutersResult, _, err := c.apiClient.LogicalRoutingAndServicesApi.ListLogicalRouterPorts(c.apiClient.Context, localVarOptionals)
	return lroutersResult, err
}

func (c *nsxtClient) GetLogicalRouterPortStatisticsSummary(lrportID string) (manager.LogicalRouterPortStatisticsSummary, error) {
	lrportsStatus, _, err := c.apiClient.LogicalRoutingAndServicesApi.GetLogicalRouterPortStatisticsSummary(c.apiClient.Context, lrportID, nil)
	return lrportsStatus, err
}

func (c *nsxtClient) ListDhcpServers(localVarOptionals map[string]interface{}) (manager.LogicalDhcpServerListResult, error) {
	dhcpServersResult, _, err := c.apiClient.ServicesApi.ListDhcpServers(c.apiClient.Context, localVarOptionals)
	return dhcpServersResult, err
}

func (c *nsxtClient) GetDhcpStatus(dhcpID string, localVarOptionals map[string]interface{}) (manager.DhcpServerStatus, error) {
	dhcpServerStatus, _, err := c.apiClient.ServicesApi.GetDhcpStatus(c.apiClient.Context, dhcpID)
	return dhcpServerStatus, err
}

func (c *nsxtClient) ListTransportNodes(localVarOptionals map[string]interface{}) (manager.TransportNodeListResult, error) {
	transportNodeStatus, _, err := c.apiClient.NetworkTransportApi.ListTransportNodes(c.apiClient.Context, localVarOptionals)
	return transportNodeStatus, err
}

func (c *nsxtClient) GetTransportNodeStatus(nodeID string) (manager.TransportNodeStatus, error) {
	transportNodeStatus, _, err := c.apiClient.TroubleshootingAndMonitoringApi.GetTransportNodeStatus(c.apiClient.Context, nodeID, nil)
	return transportNodeStatus, err
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

func (c *nsxtClient) ListLogicalSwitches(localVarOptionals map[string]interface{}) (manager.LogicalSwitchListResult, error) {
	logicalSwitchesResult, _, err := c.apiClient.LogicalSwitchingApi.ListLogicalSwitches(c.apiClient.Context, localVarOptionals)
	return logicalSwitchesResult, err
}

func (c *nsxtClient) GetLogicalSwitchState(lswitchID string) (manager.LogicalSwitchState, error) {
	logicalSwitchesStatus, _, err := c.apiClient.LogicalSwitchingApi.GetLogicalSwitchState(c.apiClient.Context, lswitchID)
	return logicalSwitchesStatus, err
}

func (c *nsxtClient) ListAllTransportZones() ([]manager.TransportZone, error) {
	var transportZones []manager.TransportZone
	var cursor string
	for {
		localVarOptionals := make(map[string]interface{})
		localVarOptionals["cursor"] = cursor
		transportZoneListResult, _, err := c.apiClient.NetworkTransportApi.ListTransportZones(c.apiClient.Context, localVarOptionals)
		if err != nil {
			return transportZones, err
		}
		transportZones = append(transportZones, transportZoneListResult.Results...)
		cursor = transportZoneListResult.Cursor
		if len(cursor) == 0 {
			break
		}
	}
	return transportZones, nil
}

func (c *nsxtClient) GetHeatmapTransportZoneStatus(zoneID string) (monitoring.HeatMapTransportZoneStatus, error) {
	heatmapTransportZoneStatus, _, err := c.apiClient.TroubleshootingAndMonitoringApi.GetHeatmapTransportZoneStatus(c.apiClient.Context, zoneID, nil)
	return heatmapTransportZoneStatus, err
}
