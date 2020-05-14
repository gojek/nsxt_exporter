package collector

import (
	"nsxt_exporter/client"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	nsxt "github.com/vmware/go-vmware-nsxt"
)

func init() {
	registerCollector("system", createSystemCollectorFactory)
}

type systemCollector struct {
	systemClient client.SystemClient
	logger       log.Logger

	clusterStatus                    *prometheus.Desc
	clusterNodeStatus                *prometheus.Desc
	applianceManagementServiceStatus *prometheus.Desc
}

type systemStatusMetric struct {
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
	applianceManagementServiceStatus := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "appliance_service", "status"),
		"Status of NSX-T appliance management service UP/DOWN.",
		nil,
		nil,
	)
	return &systemCollector{
		systemClient: systemClient,
		logger:       logger,

		clusterStatus:                    clusterStatus,
		clusterNodeStatus:                clusterNodeStatus,
		applianceManagementServiceStatus: applianceManagementServiceStatus,
	}
}

// Describe implements the prometheus.Collector interface.
func (sc *systemCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- sc.clusterStatus
	ch <- sc.clusterNodeStatus
	ch <- sc.applianceManagementServiceStatus
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

	serviceMetrics := sc.collectClusterServiceMetrics()
	for _, svm := range serviceMetrics {
		ch <- prometheus.MustNewConstMetric(sc.applianceManagementServiceStatus, prometheus.GaugeValue, svm.Status)
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

func (sc *systemCollector) collectClusterServiceMetrics() (systemStatusMetrics []systemStatusMetric) {
	applianceStatus, err := sc.systemClient.ReadApplianceManagementServiceStatus()
	if err != nil {
		level.Error(sc.logger).Log("msg", "Unable to collect appliance management service status")
		return
	}

	applianceStatusMetric := systemStatusMetric{
		Status: 0.0,
	}
	if strings.ToUpper(applianceStatus.RuntimeState) == "RUNNING" {
		applianceStatusMetric.Status = 1.0
	}
	systemStatusMetrics = append(systemStatusMetrics, applianceStatusMetric)
	return
}
