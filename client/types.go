package client

import (
	"github.com/vmware/go-vmware-nsxt/administration"
	"github.com/vmware/go-vmware-nsxt/manager"
)

// LogicalPortClient represents API group logical port for NSX-T client.
type LogicalPortClient interface {
	ListLogicalPorts(localVarOptionals map[string]interface{}) (manager.LogicalPortListResult, error)
	GetLogicalPortOperationalStatus(lportID string, localVarOptionals map[string]interface{}) (manager.LogicalPortOperationalStatus, error)
}

// LogicalRouterPortClient represents API group logical router port for NSX-T client.
type LogicalRouterPortClient interface {
	ListAllLogicalRouterPorts() ([]manager.LogicalRouterPort, error)
	GetLogicalRouterPortStatisticsSummary(lrportID string) (manager.LogicalRouterPortStatisticsSummary, error)
}

// DHCPClient represents API group DHCP for NSX-T client.
type DHCPClient interface {
	ListAllDHCPServers() ([]manager.LogicalDhcpServer, error)
	GetDhcpStatus(dhcpID string, localVarOptionals map[string]interface{}) (manager.DhcpServerStatus, error)
	GetDHCPStatistic(dhcpID string) (manager.DhcpStatistics, error)
}

// TransportNodeClient represents API group Transport Node for NSX-T client.
type TransportNodeClient interface {
	ListAllTransportNodes() ([]manager.TransportNode, error)
	GetTransportNodeStatus(nodeID string) (manager.TransportNodeStatus, error)
	ListAllEdgeClusters() ([]manager.EdgeCluster, error)
}

// SystemClient represents API group system for NSX-t client.
type SystemClient interface {
	ReadClusterStatus() (administration.ClusterStatus, error)
	ReadClusterNodesAggregateStatus() (administration.ClustersAggregateInfo, error)
	ReadApplianceManagementServiceStatus() (administration.NodeServiceStatusProperties, error)
	ReadNSXMessageBusServiceStatus() (administration.NodeServiceStatusProperties, error)
	ReadNTPServiceStatus() (administration.NodeServiceStatusProperties, error)
	ReadNsxUpgradeAgentServiceStatus() (administration.NodeServiceStatusProperties, error)
	ReadProtonServiceStatus() (administration.NodeServiceStatusProperties, error)
	ReadProxyServiceStatus() (administration.NodeServiceStatusProperties, error)
	ReadRabbitMQServiceStatus() (administration.NodeServiceStatusProperties, error)
	ReadRepositoryServiceStatus() (administration.NodeServiceStatusProperties, error)
	ReadSNMPServiceStatus() (administration.NodeServiceStatusProperties, error)
	ReadSSHServiceStatus() (administration.NodeServiceStatusProperties, error)
	ReadSearchServiceStatus() (administration.NodeServiceStatusProperties, error)
	ReadSyslogServiceStatus() (administration.NodeServiceStatusProperties, error)
}

// LogicalSwitchClient represents API group Logical Switch for NSX-T client.
type LogicalSwitchClient interface {
	ListAllLogicalSwitches() ([]manager.LogicalSwitch, error)
	GetLogicalSwitchState(lswitchID string) (manager.LogicalSwitchState, error)
	GetLogicalSwitchStatistic(lswitchID string) (manager.LogicalSwitchStatistics, error)
}
