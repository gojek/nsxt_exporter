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

var possibleSystemServiceStatus = [...]string{"RUNNING", "STOPPED"}
var possibleNodeStatus = [...]string{"CONNECTED", "DISCONNECTED", "UNKNOWN"}

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

type clusterStatusMetric struct {
	Status float64
}

type managementNodeMetric struct {
	Name         string
	IPAddress    string
	Type         string
	StatusDetail map[string]float64

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

type controllerNodeStatusMetric struct {
	IPAddress    string
	Status       float64
	StatusDetail map[string]float64
}

type serviceStatusMetric struct {
	Name         string
	StatusDetail map[string]float64
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
		"Status of NSX-T system cluster nodes",
		[]string{"ip_address", "type", "status"},
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
		"Status of NSX-T system service",
		[]string{"name", "status"},
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
	clusterStatusMetrics := sc.collectClusterStatusMetrics()
	for _, sm := range clusterStatusMetrics {
		ch <- prometheus.MustNewConstMetric(sc.clusterStatus, prometheus.GaugeValue, sm.Status)
	}

	controllerNodeStatusMetrics, nodeMetrics := sc.collectClusterNodeMetrics()
	for _, nm := range nodeMetrics {
		nodeType := "management"
		for status, value := range nm.StatusDetail {
			ch <- prometheus.MustNewConstMetric(sc.clusterNodeStatus, prometheus.GaugeValue, value, nm.IPAddress, nodeType, status)
		}
		ch <- prometheus.MustNewConstMetric(sc.clusterNodeCPUCoresUse, prometheus.GaugeValue, nm.LoadAverageOneMinute, nm.IPAddress, nodeType, "1")
		ch <- prometheus.MustNewConstMetric(sc.clusterNodeCPUCoresUse, prometheus.GaugeValue, nm.LoadAverageFiveMinutes, nm.IPAddress, nodeType, "5")
		ch <- prometheus.MustNewConstMetric(sc.clusterNodeCPUCoresUse, prometheus.GaugeValue, nm.LoadAverageFifteenMinutes, nm.IPAddress, nodeType, "15")
		ch <- prometheus.MustNewConstMetric(sc.clusterNodeCPUCoresTotal, prometheus.GaugeValue, nm.CPUCores, nm.IPAddress, nodeType)
		ch <- prometheus.MustNewConstMetric(sc.clusterNodeMemoryUse, prometheus.GaugeValue, nm.MemoryUse, nm.IPAddress, nodeType)
		ch <- prometheus.MustNewConstMetric(sc.clusterNodeMemoryTotal, prometheus.GaugeValue, nm.MemoryTotal, nm.IPAddress, nodeType)
		ch <- prometheus.MustNewConstMetric(sc.clusterNodeMemoryCached, prometheus.GaugeValue, nm.MemoryCached, nm.IPAddress, nodeType)
		ch <- prometheus.MustNewConstMetric(sc.clusterNodeSwapUse, prometheus.GaugeValue, nm.SwapUse, nm.IPAddress, nodeType)
		ch <- prometheus.MustNewConstMetric(sc.clusterNodeSwapTotal, prometheus.GaugeValue, nm.SwapTotal, nm.IPAddress, nodeType)

		for filesystem, diskUse := range nm.DiskUse {
			ch <- prometheus.MustNewConstMetric(sc.clusterNodeDiskUse, prometheus.GaugeValue, diskUse, nm.IPAddress, nodeType, filesystem)
		}
		for filesystem, diskTotal := range nm.DiskTotal {
			ch <- prometheus.MustNewConstMetric(sc.clusterNodeDiskTotal, prometheus.GaugeValue, diskTotal, nm.IPAddress, nodeType, filesystem)
		}
	}
	for _, nm := range controllerNodeStatusMetrics {
		nodeType := "controller"
		for status, value := range nm.StatusDetail {
			ch <- prometheus.MustNewConstMetric(sc.clusterNodeStatus, prometheus.GaugeValue, value, nm.IPAddress, nodeType, status)
		}
	}

	serviceMetrics := sc.collectServiceStatusMetrics()
	for _, svm := range serviceMetrics {
		for status, value := range svm.StatusDetail {
			ch <- prometheus.MustNewConstMetric(sc.systemServiceStatus, prometheus.GaugeValue, value, svm.Name, status)
		}
	}
}

func (sc *systemCollector) collectClusterStatusMetrics() (clusterStatusMetrics []clusterStatusMetric) {
	clusterStatus, err := sc.systemClient.ReadClusterStatus()
	if err != nil {
		level.Error(sc.logger).Log("msg", "Unable to collect cluster status")
		return
	}
	clusterStatusMetric := clusterStatusMetric{
		Status: 0.0,
	}
	if strings.ToUpper(clusterStatus.ControlClusterStatus.Status) == "STABLE" &&
		strings.ToUpper(clusterStatus.MgmtClusterStatus.Status) == "STABLE" {
		clusterStatusMetric.Status = 1.0
	}
	clusterStatusMetrics = append(clusterStatusMetrics, clusterStatusMetric)
	return
}

func (sc *systemCollector) collectClusterNodeMetrics() (controllerNodeStatusMetrics []controllerNodeStatusMetric, managementNodeMetrics []managementNodeMetric) {
	clusterNodes, err := sc.systemClient.ReadClusterNodesAggregateStatus()
	if err != nil {
		level.Error(sc.logger).Log("msg", "Unable to collect cluster nodes status")
		return
	}

	controllerNodeStatusMetrics = sc.extractControllerStatusMetrics(clusterNodes.ControllerCluster)
	managementNodeMetrics = sc.extractManagementNodeMetrics(clusterNodes.ManagementCluster)

	return
}

func (sc *systemCollector) extractControllerStatusMetrics(controllerNodes []administration.ControllerNodeAggregateInfo) (controllerNodeStatusMetrics []controllerNodeStatusMetric) {
	for _, c := range controllerNodes {
		controllerStatusMetric := controllerNodeStatusMetric{
			IPAddress:    c.RoleConfig.ControlPlaneListenAddr.IpAddress,
			StatusDetail: map[string]float64{},
		}
		for _, status := range possibleNodeStatus {
			controllerStatusMetric.StatusDetail[status] = 0.0
		}
		if strings.ToUpper(c.NodeStatus.ControlClusterStatus.ControlClusterStatus) == "CONNECTED" &&
			strings.ToUpper(c.NodeStatus.ControlClusterStatus.MgmtConnectionStatus.ConnectivityStatus) == "CONNECTED" {
			controllerStatusMetric.StatusDetail["CONNECTED"] = 1.0
		} else {
			controllerStatusMetric.StatusDetail["DISCONNECTED"] = 1.0
		}
		controllerNodeStatusMetrics = append(controllerNodeStatusMetrics, controllerStatusMetric)
	}
	return
}

func (sc *systemCollector) extractManagementNodeMetrics(managementNodes []administration.ManagementNodeAggregateInfo) (managementNodeMetrics []managementNodeMetric) {
	for _, m := range managementNodes {
		managementNodeMetric := managementNodeMetric{
			IPAddress:    m.RoleConfig.MgmtPlaneListenAddr.IpAddress,
			Type:         "management",
			StatusDetail: map[string]float64{},
		}
		for _, possibleStatus := range possibleNodeStatus {
			managementNodeMetric.StatusDetail[possibleStatus] = 0.0
			if strings.ToUpper(m.NodeStatus.MgmtClusterStatus.MgmtClusterStatus) == possibleStatus {
				managementNodeMetric.StatusDetail[possibleStatus] = 1.0
			}
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
		managementNodeMetrics = append(managementNodeMetrics, managementNodeMetric)
	}
	return
}

func (sc *systemCollector) collectServiceStatusMetrics() (serviceStatusMetrics []serviceStatusMetric) {
	var collectors []func() (serviceStatusMetric, error)
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
		serviceStatusMetrics = append(serviceStatusMetrics, m)
	}
	return
}

func (sc *systemCollector) collectServiceStatusMetric(name string, collectSystemService func() (administration.NodeServiceStatusProperties, error)) (serviceStatusMetric, error) {
	status, err := collectSystemService()
	if err != nil {
		return serviceStatusMetric{}, err
	}
	statusMetric := serviceStatusMetric{
		Name: name,
	}
	statusDetail := map[string]float64{}
	for _, possibleStatus := range possibleSystemServiceStatus {
		statusDetail[possibleStatus] = 0.0
		if strings.ToUpper(status.RuntimeState) == possibleStatus {
			statusDetail[possibleStatus] = 1.0
		}
	}
	statusMetric.StatusDetail = statusDetail
	return statusMetric, nil
}

func (sc *systemCollector) collectApplianceServiceMetric() (serviceStatusMetric, error) {
	return sc.collectServiceStatusMetric("appliance", sc.systemClient.ReadApplianceManagementServiceStatus)
}

func (sc *systemCollector) collectMessageBusServiceMetric() (serviceStatusMetric, error) {
	return sc.collectServiceStatusMetric("message_bus", sc.systemClient.ReadNSXMessageBusServiceStatus)
}

func (sc *systemCollector) collectNTPServiceMetric() (serviceStatusMetric, error) {
	return sc.collectServiceStatusMetric("ntp", sc.systemClient.ReadNTPServiceStatus)
}

func (sc *systemCollector) collectUpgradeAgentServiceMetric() (serviceStatusMetric, error) {
	return sc.collectServiceStatusMetric("upgrade_agent", sc.systemClient.ReadNsxUpgradeAgentServiceStatus)
}

func (sc *systemCollector) collectProtonServiceMetric() (serviceStatusMetric, error) {
	return sc.collectServiceStatusMetric("proton", sc.systemClient.ReadProtonServiceStatus)
}

func (sc *systemCollector) collectProxyServiceMetric() (serviceStatusMetric, error) {
	return sc.collectServiceStatusMetric("proxy", sc.systemClient.ReadProxyServiceStatus)
}

func (sc *systemCollector) collectRabbitMQServiceMetric() (serviceStatusMetric, error) {
	return sc.collectServiceStatusMetric("rabbitmq", sc.systemClient.ReadRabbitMQServiceStatus)
}

func (sc *systemCollector) collectRepositoryServiceMetric() (serviceStatusMetric, error) {
	return sc.collectServiceStatusMetric("repository", sc.systemClient.ReadRepositoryServiceStatus)
}

func (sc *systemCollector) collectSNMPServiceMetric() (serviceStatusMetric, error) {
	return sc.collectServiceStatusMetric("snmp", sc.systemClient.ReadSNMPServiceStatus)
}

func (sc *systemCollector) collectSSHServiceMetric() (serviceStatusMetric, error) {
	return sc.collectServiceStatusMetric("ssh", sc.systemClient.ReadSSHServiceStatus)
}

func (sc *systemCollector) collectSearchServiceMetric() (serviceStatusMetric, error) {
	return sc.collectServiceStatusMetric("search", sc.systemClient.ReadSearchServiceStatus)
}

func (sc *systemCollector) collectSyslogServiceMetric() (serviceStatusMetric, error) {
	return sc.collectServiceStatusMetric("syslog", sc.systemClient.ReadSyslogServiceStatus)
}
