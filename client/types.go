package client

import (
	"github.com/vmware/go-vmware-nsxt/administration"
	"github.com/vmware/go-vmware-nsxt/manager"
)

// LogicalPortClient represents API group logical port for NSX-T client.
type LogicalPortClient interface {
	ListLogicalPorts(localVarOptionals map[string]interface{}) (manager.LogicalPortListResult, error)
	GetLogicalPortStatusSummary(localVarOptionals map[string]interface{}) (manager.LogicalPortStatusSummary, error)
	GetLogicalPortOperationalStatus(lportID string, localVarOptionals map[string]interface{}) (manager.LogicalPortOperationalStatus, error)
}

// LogicalRouterPortClient represents API group logical router port for NSX-T client.
type LogicalRouterPortClient interface {
	ListLogicalRouterPorts(localVarOptionals map[string]interface{}) (manager.LogicalRouterPortListResult, error)
	GetLogicalRouterPortStatisticsSummary(lrportID string) (manager.LogicalRouterPortStatisticsSummary, error)
}

// DHCPClient represents API group DHCP for NSX-T client.
type DHCPClient interface {
	ListDhcpServers(localVarOptionals map[string]interface{}) (manager.LogicalDhcpServerListResult, error)
	GetDhcpStatus(dhcpID string, localVarOptionals map[string]interface{}) (manager.DhcpServerStatus, error)
}

// TransportNodeClient represents API group Transport Node for NSX-T client.
type TransportNodeClient interface {
	ListTransportNodes(localVarOptionals map[string]interface{}) (manager.TransportNodeListResult, error)
	GetTransportNodeStatus(nodeID string) (manager.TransportNodeStatus, error)
	ListEdgeClusters() (manager.EdgeClusterListResult, error)
}

// SystemClient represents API group system for NSX-t client.
type SystemClient interface {
	ReadClusterStatus() (administration.ClusterStatus, error)
	ReadClusterNodesAggregateStatus() (administration.ClustersAggregateInfo, error)
	ReadApplianceManagementServiceStatus() (administration.NodeServiceStatusProperties, error)
}

// LogicalSwitchClient represents API group Logical Switch for NSX-T client.
type LogicalSwitchClient interface {
	ListLogicalSwitches(localVarOptionals map[string]interface{}) (manager.LogicalSwitchListResult, error)
	GetLogicalSwitchState(lswitchID string) (manager.LogicalSwitchState, error)
}
