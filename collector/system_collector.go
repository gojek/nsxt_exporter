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

	clusterStatus            *prometheus.Desc
	clusterNodeStatus        *prometheus.Desc
	clusterNodeCPUCoresUse   *prometheus.Desc
	clusterNodeCPUCoresTotal *prometheus.Desc
	clusterNodeMemoryUse     *prometheus.Desc
	clusterNodeMemoryTotal   *prometheus.Desc
	clusterNodeMemoryCached  *prometheus.Desc
	clusterNodeSwapUse       *prometheus.Desc
	clusterNodeSwapTotal     *prometheus.Desc
	clusterNodeDiskUse       *prometheus.Desc
	clusterNodeDiskTotal     *prometheus.Desc
	systemServiceStatus      *prometheus.Desc
}

type systemStatusMetric struct {
	Name      string
	IPAddress string
	Type      string
	Status    float64

	CPUCores                  float64
	LoadAverageOneMinute      float64
	LoadAverageFiveMinutes    float64
	LoadAverageFifteenMinutes float64
	MemoryCached              float64
	MemoryUse                 float64
	MemoryTotal               float64
	SwapUse                   float64
	SwapTotal                 float64
	DiskUse                   map[string]float64
	DiskTotal                 map[string]float64
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
	clusterNodeCPUCoresUse := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_node", "cpu_use_cores"),
		"NSX-T system cluster node average load",
		[]string{"ip_address", "type", "minutes"},
		nil,
	)
	clusterNodeCPUCoresTotal := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_node", "cpu_total_cores"),
		"NSX-T system cluster nodes cpu cores total",
		[]string{"ip_address", "type"},
		nil,
	)
	clusterNodeMemoryUse := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_node", "memory_use_kilobytes"),
		"NSX-T system cluster node memory use",
		[]string{"ip_address", "type"},
		nil,
	)
	clusterNodeMemoryTotal := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_node", "memory_total_kilobytes"),
		"NSX-T system cluster node memory total",
		[]string{"ip_address", "type"},
		nil,
	)
	clusterNodeMemoryCached := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_node", "memory_cached_kilobytes"),
		"NSX-T system cluster node cached memory",
		[]string{"ip_address", "type"},
		nil,
	)
	clusterNodeSwapUse := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_node", "swap_use_kilobytes"),
		"NSX-T system cluster node swap use",
		[]string{"ip_address", "type"},
		nil,
	)
	clusterNodeSwapTotal := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_node", "swap_total_kilobytes"),
		"NSX-T system cluster node swap total",
		[]string{"ip_address", "type"},
		nil,
	)
	clusterNodeDiskUse := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_node", "disk_use_kilobytes"),
		"NSX-T system cluster node disk use",
		[]string{"ip_address", "type", "filesystem"},
		nil,
	)
	clusterNodeDiskTotal := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_node", "disk_total_kilobytes"),
		"NSX-T system cluster node disk total",
		[]string{"ip_address", "type", "filesystem"},
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

		clusterStatus:            clusterStatus,
		clusterNodeStatus:        clusterNodeStatus,
		clusterNodeCPUCoresUse:   clusterNodeCPUCoresUse,
		clusterNodeCPUCoresTotal: clusterNodeCPUCoresTotal,
		clusterNodeMemoryUse:     clusterNodeMemoryUse,
		clusterNodeMemoryTotal:   clusterNodeMemoryTotal,
		clusterNodeMemoryCached:  clusterNodeMemoryCached,
		clusterNodeSwapUse:       clusterNodeSwapUse,
		clusterNodeSwapTotal:     clusterNodeSwapTotal,
		clusterNodeDiskUse:       clusterNodeDiskUse,
		clusterNodeDiskTotal:     clusterNodeDiskTotal,
		systemServiceStatus:      systemServiceStatus,
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
		if nm.Type == "management" {
			ch <- prometheus.MustNewConstMetric(sc.clusterNodeCPUCoresUse, prometheus.GaugeValue, nm.LoadAverageOneMinute, nm.IPAddress, nm.Type, "1")
			ch <- prometheus.MustNewConstMetric(sc.clusterNodeCPUCoresUse, prometheus.GaugeValue, nm.LoadAverageFiveMinutes, nm.IPAddress, nm.Type, "5")
			ch <- prometheus.MustNewConstMetric(sc.clusterNodeCPUCoresUse, prometheus.GaugeValue, nm.LoadAverageFifteenMinutes, nm.IPAddress, nm.Type, "15")
			ch <- prometheus.MustNewConstMetric(sc.clusterNodeCPUCoresTotal, prometheus.GaugeValue, nm.CPUCores, nm.IPAddress, nm.Type)
			ch <- prometheus.MustNewConstMetric(sc.clusterNodeMemoryUse, prometheus.GaugeValue, nm.MemoryUse, nm.IPAddress, nm.Type)
			ch <- prometheus.MustNewConstMetric(sc.clusterNodeMemoryTotal, prometheus.GaugeValue, nm.MemoryTotal, nm.IPAddress, nm.Type)
			ch <- prometheus.MustNewConstMetric(sc.clusterNodeMemoryCached, prometheus.GaugeValue, nm.MemoryCached, nm.IPAddress, nm.Type)
			ch <- prometheus.MustNewConstMetric(sc.clusterNodeSwapUse, prometheus.GaugeValue, nm.SwapUse, nm.IPAddress, nm.Type)
			ch <- prometheus.MustNewConstMetric(sc.clusterNodeSwapTotal, prometheus.GaugeValue, nm.SwapTotal, nm.IPAddress, nm.Type)

			for filesystem, diskUse := range nm.DiskUse {
				ch <- prometheus.MustNewConstMetric(sc.clusterNodeDiskUse, prometheus.GaugeValue, diskUse, nm.IPAddress, nm.Type, filesystem)
			}
			for filesystem, diskTotal := range nm.DiskTotal {
				ch <- prometheus.MustNewConstMetric(sc.clusterNodeDiskTotal, prometheus.GaugeValue, diskTotal, nm.IPAddress, nm.Type, filesystem)
			}
		}
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
	clusterNodes, err := sc.systemClient.ReadClusterNodesAggregateStatus()
	if err != nil {
		level.Error(sc.logger).Log("msg", "Unable to collect cluster nodes status")
		return
	}

	controllerStatusMetrics := sc.extractControllerStatusMetrics(clusterNodes.ControllerCluster)
	systemStatusMetrics = append(systemStatusMetrics, controllerStatusMetrics...)

	managementNodeMetrics := sc.extractManagementNodeMetrics(clusterNodes.ManagementCluster)
	systemStatusMetrics = append(systemStatusMetrics, managementNodeMetrics...)

	return
}

func (sc *systemCollector) extractControllerStatusMetrics(controllerNodes []administration.ControllerNodeAggregateInfo) (systemStatusMetrics []systemStatusMetric) {
	for _, c := range controllerNodes {
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
	return
}

func (sc *systemCollector) extractManagementNodeMetrics(managementNodes []administration.ManagementNodeAggregateInfo) (systemStatusMetrics []systemStatusMetric) {
	for _, m := range managementNodes {
		managementNodeMetric := systemStatusMetric{
			IPAddress: m.RoleConfig.MgmtPlaneListenAddr.IpAddress,
			Type:      "management",
			Status:    0.0,
		}
		if strings.ToUpper(m.NodeStatus.MgmtClusterStatus.MgmtClusterStatus) == "CONNECTED" {
			managementNodeMetric.Status = 1.0
		}
		if len(m.NodeStatusProperties) > 0 {
			const latestDataIndex = 0
			prop := m.NodeStatusProperties[latestDataIndex]

			managementNodeMetric.CPUCores = float64(prop.CpuCores)

			managementNodeMetric.LoadAverageOneMinute = float64(prop.LoadAverage[0])
			managementNodeMetric.LoadAverageFiveMinutes = float64(prop.LoadAverage[1])
			managementNodeMetric.LoadAverageFifteenMinutes = float64(prop.LoadAverage[2])

			managementNodeMetric.MemoryUse = float64(prop.MemUsed)
			managementNodeMetric.MemoryTotal = float64(prop.MemTotal)
			managementNodeMetric.MemoryCached = float64(prop.MemCache)

			managementNodeMetric.SwapUse = float64(prop.SwapUsed)
			managementNodeMetric.SwapTotal = float64(prop.SwapTotal)

			managementNodeMetric.DiskUse = make(map[string]float64)
			managementNodeMetric.DiskTotal = make(map[string]float64)
			for _, disk := range prop.FileSystems {
				managementNodeMetric.DiskUse[disk.Mount] = float64(disk.Used)
				managementNodeMetric.DiskTotal[disk.Mount] = float64(disk.Total)
			}
		}
		systemStatusMetrics = append(systemStatusMetrics, managementNodeMetric)
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
