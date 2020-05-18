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

	serviceMetrics := sc.collectServiceStatusMetrics()
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

func (sc *systemCollector) collectServiceStatusMetrics() (systemStatusMetrics []systemStatusMetric) {
	var collectors []func() (systemStatusMetric, error)
	collectors = append(collectors, sc.collectApplianceServiceMetric)
	collectors = append(collectors, sc.collectMessageBusServiceMetric)
	collectors = append(collectors, sc.collectNTPServiceMetric)
	collectors = append(collectors, sc.collectUpgradeAgentServiceMetric)
	collectors = append(collectors, sc.collectProtonServiceMetric)
	collectors = append(collectors, sc.collectProxyServiceMetric)
	collectors = append(collectors, sc.collectRabbitMQServiceMetric)
	collectors = append(collectors, sc.collectSNMPServiceMetric)
	collectors = append(collectors, sc.collectSSHServiceMetric)
	collectors = append(collectors, sc.collectSearchServiceMetric)
	collectors = append(collectors, sc.collectSyslogServiceMetric)

	for _, collectServiceStatusMetric := range collectors {
		m, err := collectServiceStatusMetric()
		if err != nil {
			level.Error(sc.logger).Log("msg", "Unable to collect system service status", "name", m.Name, "error", err.Error())
			continue
		}
		systemStatusMetrics = append(systemStatusMetrics, m)
	}
	return
}

func (sc *systemCollector) collectServiceStatusMetric(name string, collectSystemService func() (administration.NodeServiceStatusProperties, error)) (systemStatusMetric, error) {
	status, err := collectSystemService()
	if err != nil {
		return systemStatusMetric{}, err
	}
	statusMetric := systemStatusMetric{
		Name:   name,
		Status: 0.0,
	}
	if strings.ToUpper(status.RuntimeState) == "RUNNING" {
		statusMetric.Status = 1.0
	}
	return statusMetric, nil
}

func (sc *systemCollector) collectApplianceServiceMetric() (systemStatusMetric, error) {
	return sc.collectServiceStatusMetric("appliance", sc.systemClient.ReadApplianceManagementServiceStatus)
}

func (sc *systemCollector) collectMessageBusServiceMetric() (systemStatusMetric, error) {
	return sc.collectServiceStatusMetric("message_bus", sc.systemClient.ReadNSXMessageBusServiceStatus)
}

func (sc *systemCollector) collectNTPServiceMetric() (systemStatusMetric, error) {
	return sc.collectServiceStatusMetric("ntp", sc.systemClient.ReadNTPServiceStatus)
}

func (sc *systemCollector) collectUpgradeAgentServiceMetric() (systemStatusMetric, error) {
	return sc.collectServiceStatusMetric("upgrade_agent", sc.systemClient.ReadNsxUpgradeAgentServiceStatus)
}

func (sc *systemCollector) collectProtonServiceMetric() (systemStatusMetric, error) {
	return sc.collectServiceStatusMetric("proton", sc.systemClient.ReadProtonServiceStatus)
}

func (sc *systemCollector) collectProxyServiceMetric() (systemStatusMetric, error) {
	return sc.collectServiceStatusMetric("proxy", sc.systemClient.ReadProxyServiceStatus)
}

func (sc *systemCollector) collectRabbitMQServiceMetric() (systemStatusMetric, error) {
	return sc.collectServiceStatusMetric("rabbitmq", sc.systemClient.ReadRabbitMQServiceStatus)
}

func (sc *systemCollector) collectRepositoryServiceMetric() (systemStatusMetric, error) {
	return sc.collectServiceStatusMetric("repository", sc.systemClient.ReadRepositoryServiceStatus)
}

func (sc *systemCollector) collectSNMPServiceMetric() (systemStatusMetric, error) {
	return sc.collectServiceStatusMetric("snmp", sc.systemClient.ReadSNMPServiceStatus)
}

func (sc *systemCollector) collectSSHServiceMetric() (systemStatusMetric, error) {
	return sc.collectServiceStatusMetric("ssh", sc.systemClient.ReadSSHServiceStatus)
}

func (sc *systemCollector) collectSearchServiceMetric() (systemStatusMetric, error) {
	return sc.collectServiceStatusMetric("search", sc.systemClient.ReadSearchServiceStatus)
}

func (sc *systemCollector) collectSyslogServiceMetric() (systemStatusMetric, error) {
	return sc.collectServiceStatusMetric("syslog", sc.systemClient.ReadSyslogServiceStatus)
}
