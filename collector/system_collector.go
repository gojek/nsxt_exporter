package collector

import (
	"nsxt_exporter/client"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	nsxt "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/administration"
)

func init() {
	registerCollector("system", createSystemCollectorFactory)
}

type systemCollector struct {
	systemClient client.SystemClient
	logger       log.Logger

	clusterStatus       *prometheus.Desc
	clusterNodeStatus   *prometheus.Desc
	systemServiceStatus *prometheus.Desc
}

type systemStatusMetric struct {
	Name      string
	IPAddress string
	Type      string
	Status    float64
}

func createSystemCollectorFactory(apiClient *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	nsxtClient := client.NewNSXTClient(apiClient, logger)
	return newSystemCollector(nsxtClient, logger)
}

func newSystemCollector(systemClient client.SystemClient, logger log.Logger) *systemCollector {
	clusterStatus := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster", "status"),
		"Status of NSX-T system controller and management cluster.",
		nil,
		nil,
	)
	clusterNodeStatus := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_node", "status"),
		"Status of NSX-T system cluster nodes UP/DOWN.",
		[]string{"ip_address", "type"},
		nil,
	)
	systemServiceStatus := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "system_service", "status"),
		"Status of NSX-T system service UP/DOWN.",
		[]string{"name"},
		nil,
	)
	return &systemCollector{
		systemClient: systemClient,
		logger:       logger,

		clusterStatus:       clusterStatus,
		clusterNodeStatus:   clusterNodeStatus,
		systemServiceStatus: systemServiceStatus,
	}
}

// Describe implements the prometheus.Collector interface.
func (sc *systemCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- sc.clusterStatus
	ch <- sc.clusterNodeStatus
	ch <- sc.systemServiceStatus
}

// Collect implements the prometheus.Collector interface.
func (sc *systemCollector) Collect(ch chan<- prometheus.Metric) {
	statusMetrics := sc.collectClusterStatusMetrics()
	for _, sm := range statusMetrics {
		ch <- prometheus.MustNewConstMetric(sc.clusterStatus, prometheus.GaugeValue, sm.Status)
	}

	nodeMetrics := sc.collectClusterNodeMetrics()
	for _, nm := range nodeMetrics {
		ch <- prometheus.MustNewConstMetric(sc.clusterNodeStatus, prometheus.GaugeValue, nm.Status, nm.IPAddress, nm.Type)
	}

	var serviceMetrics []systemStatusMetric
	serviceMetrics = append(serviceMetrics, sc.collectApplianceServiceMetric())
	serviceMetrics = append(serviceMetrics, sc.collectMessageBusServiceMetric())
	serviceMetrics = append(serviceMetrics, sc.collectNTPServiceMetric())
	serviceMetrics = append(serviceMetrics, sc.collectUpgradeAgentServiceMetric())
	serviceMetrics = append(serviceMetrics, sc.collectProtonServiceMetric())
	serviceMetrics = append(serviceMetrics, sc.collectProxyServiceMetric())
	serviceMetrics = append(serviceMetrics, sc.collectRabbitMQServiceMetric())
	serviceMetrics = append(serviceMetrics, sc.collectSNMPServiceMetric())
	serviceMetrics = append(serviceMetrics, sc.collectSSHServiceMetric())
	serviceMetrics = append(serviceMetrics, sc.collectSearchServiceMetric())
	serviceMetrics = append(serviceMetrics, sc.collectSyslogServiceMetric())
	for _, svm := range serviceMetrics {
		ch <- prometheus.MustNewConstMetric(sc.systemServiceStatus, prometheus.GaugeValue, svm.Status, svm.Name)
	}
}

func (sc *systemCollector) collectClusterStatusMetrics() (systemStatusMetrics []systemStatusMetric) {
	clusterStatus, err := sc.systemClient.ReadClusterStatus()
	if err != nil {
		level.Error(sc.logger).Log("msg", "Unable to collect cluster status")
		return
	}
	systemStatusMetric := systemStatusMetric{
		Status: 0.0,
	}
	if strings.ToUpper(clusterStatus.ControlClusterStatus.Status) == "STABLE" &&
		strings.ToUpper(clusterStatus.MgmtClusterStatus.Status) == "STABLE" {
		systemStatusMetric.Status = 1.0
	}
	systemStatusMetrics = append(systemStatusMetrics, systemStatusMetric)
	return
}

func (sc *systemCollector) collectClusterNodeMetrics() (systemStatusMetrics []systemStatusMetric) {
	clusterNodesStatus, err := sc.systemClient.ReadClusterNodesAggregateStatus()
	if err != nil {
		level.Error(sc.logger).Log("msg", "Unable to collect cluster nodes status")
		return
	}

	for _, c := range clusterNodesStatus.ControllerCluster {
		controllerStatusMetric := systemStatusMetric{
			IPAddress: c.RoleConfig.ControlPlaneListenAddr.IpAddress,
			Type:      "controller",
			Status:    0.0,
		}
		if strings.ToUpper(c.NodeStatus.ControlClusterStatus.ControlClusterStatus) == "CONNECTED" &&
			strings.ToUpper(c.NodeStatus.ControlClusterStatus.MgmtConnectionStatus.ConnectivityStatus) == "CONNECTED" {
			controllerStatusMetric.Status = 1.0
		}
		systemStatusMetrics = append(systemStatusMetrics, controllerStatusMetric)
	}

	for _, m := range clusterNodesStatus.ManagementCluster {
		managementStatusMetric := systemStatusMetric{
			IPAddress: m.RoleConfig.MgmtPlaneListenAddr.IpAddress,
			Type:      "management",
			Status:    0.0,
		}
		if strings.ToUpper(m.NodeStatus.MgmtClusterStatus.MgmtClusterStatus) == "CONNECTED" {
			managementStatusMetric.Status = 1.0
		}
		systemStatusMetrics = append(systemStatusMetrics, managementStatusMetric)
	}
	return
}

func (sc *systemCollector) collectServiceMetric(name string, collectSystemService func() (administration.NodeServiceStatusProperties, error)) systemStatusMetric {
	status, err := collectSystemService()
	if err != nil {
		level.Error(sc.logger).Log("msg", "Unable to collect system service status", "name", name)
		return systemStatusMetric{}
	}
	statusMetric := systemStatusMetric{
		Name:   name,
		Status: 0.0,
	}
	if strings.ToUpper(status.RuntimeState) == "RUNNING" {
		statusMetric.Status = 1.0
	}
	return statusMetric
}

func (sc *systemCollector) collectApplianceServiceMetric() systemStatusMetric {
	return sc.collectServiceMetric("appliance", sc.systemClient.ReadApplianceManagementServiceStatus)
}

func (sc *systemCollector) collectMessageBusServiceMetric() systemStatusMetric {
	return sc.collectServiceMetric("message_bus", sc.systemClient.ReadNSXMessageBusServiceStatus)
}

func (sc *systemCollector) collectNTPServiceMetric() systemStatusMetric {
	return sc.collectServiceMetric("ntp", sc.systemClient.ReadNTPServiceStatus)
}

func (sc *systemCollector) collectUpgradeAgentServiceMetric() systemStatusMetric {
	return sc.collectServiceMetric("upgrade_agent", sc.systemClient.ReadNsxUpgradeAgentServiceStatus)
}

func (sc *systemCollector) collectProtonServiceMetric() systemStatusMetric {
	return sc.collectServiceMetric("proton", sc.systemClient.ReadProtonServiceStatus)
}

func (sc *systemCollector) collectProxyServiceMetric() systemStatusMetric {
	return sc.collectServiceMetric("proxy", sc.systemClient.ReadProxyServiceStatus)
}

func (sc *systemCollector) collectRabbitMQServiceMetric() systemStatusMetric {
	return sc.collectServiceMetric("rabbitmq", sc.systemClient.ReadRabbitMQServiceStatus)
}

func (sc *systemCollector) collectRepositoryServiceMetric() systemStatusMetric {
	return sc.collectServiceMetric("repository", sc.systemClient.ReadRepositoryServiceStatus)
}

func (sc *systemCollector) collectSNMPServiceMetric() systemStatusMetric {
	return sc.collectServiceMetric("snmp", sc.systemClient.ReadSNMPServiceStatus)
}

func (sc *systemCollector) collectSSHServiceMetric() systemStatusMetric {
	return sc.collectServiceMetric("ssh", sc.systemClient.ReadSSHServiceStatus)
}

func (sc *systemCollector) collectSearchServiceMetric() systemStatusMetric {
	return sc.collectServiceMetric("search", sc.systemClient.ReadSearchServiceStatus)
}

func (sc *systemCollector) collectSyslogServiceMetric() (syslogStatusMetric systemStatusMetric) {
	return sc.collectServiceMetric("syslog", sc.systemClient.ReadSyslogServiceStatus)
}
