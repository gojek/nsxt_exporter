package client

import (
	"github.com/go-kit/kit/log"
	nsxt "github.com/vmware/go-vmware-nsxt"
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

func (c *nsxtClient) GetTransportNodeStatus(nodeId string, localVarOptionals map[string]interface{}) (manager.TransportNodeStatus, error) {
	transportNodeStatus, _, err := c.apiClient.TroubleshootingAndMonitoringApi.GetTransportNodeStatus(c.apiClient.Context, nodeId, localVarOptionals)
	return transportNodeStatus, err
}
