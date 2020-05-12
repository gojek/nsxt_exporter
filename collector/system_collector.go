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
	registerCollector("system", newSystemCollector)
}

type systemCollector struct {
	systemClient client.SystemClient
	logger       log.Logger

	clusterStatus                    *prometheus.Desc
	clusterNodeStatus                *prometheus.Desc
	applianceManagementServiceStatus *prometheus.Desc
}

func newSystemCollector(apiClient *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	nsxtClient := client.NewNSXTClient(apiClient, logger)
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
		systemClient: nsxtClient,
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
	sc.collectClusterStatusMetric(ch)
	sc.collectClusterNodeMetric(ch)
	sc.collectClusterServiceMetric(ch)
}

func (sc *systemCollector) collectClusterStatusMetric(ch chan<- prometheus.Metric) {
	clusterStatus, err := sc.systemClient.ReadClusterStatus()
	if err != nil {
		level.Error(sc.logger).Log("msg", "Unable to collect cluster status")
		return
	}

	if strings.ToUpper(clusterStatus.ControlClusterStatus.Status) == "STABLE" &&
		strings.ToUpper(clusterStatus.MgmtClusterStatus.Status) == "STABLE" {
		ch <- prometheus.MustNewConstMetric(sc.clusterStatus, prometheus.GaugeValue, 1.0)
	} else {
		ch <- prometheus.MustNewConstMetric(sc.clusterStatus, prometheus.GaugeValue, 0.0)
	}
}

func (sc *systemCollector) collectClusterNodeMetric(ch chan<- prometheus.Metric) {
	clusterNodesStatus, err := sc.systemClient.ReadClusterNodesAggregateStatus()
	if err != nil {
		level.Error(sc.logger).Log("msg", "Unable to collect cluster nodes status")
		return
	}

	for _, c := range clusterNodesStatus.ControllerCluster {
		ipAddress := c.RoleConfig.ControlPlaneListenAddr.IpAddress
		if strings.ToUpper(c.NodeStatus.ControlClusterStatus.ControlClusterStatus) == "CONNECTED" &&
			strings.ToUpper(c.NodeStatus.ControlClusterStatus.MgmtConnectionStatus.ConnectivityStatus) == "CONNECTED" {
			ch <- prometheus.MustNewConstMetric(sc.clusterNodeStatus, prometheus.GaugeValue, 1.0, ipAddress, "controller")
		} else {
			ch <- prometheus.MustNewConstMetric(sc.clusterNodeStatus, prometheus.GaugeValue, 0.0, ipAddress, "controller")
		}
	}

	for _, m := range clusterNodesStatus.ManagementCluster {
		ipAddress := m.RoleConfig.MgmtPlaneListenAddr.IpAddress
		if strings.ToUpper(m.NodeStatus.MgmtClusterStatus.MgmtClusterStatus) == "CONNECTED" {
			ch <- prometheus.MustNewConstMetric(sc.clusterNodeStatus, prometheus.GaugeValue, 1.0, ipAddress, "management")
		} else {
			ch <- prometheus.MustNewConstMetric(sc.clusterNodeStatus, prometheus.GaugeValue, 0.0, ipAddress, "management")
		}
	}
}

func (sc *systemCollector) collectClusterServiceMetric(ch chan<- prometheus.Metric) {
	applianceStatus, err := sc.systemClient.ReadApplianceManagementServiceStatus()
	if err != nil {
		level.Error(sc.logger).Log("msg", "Unable to collect appliance management service status")
		return
	}

	if strings.ToUpper(applianceStatus.RuntimeState) == "RUNNING" {
		ch <- prometheus.MustNewConstMetric(sc.applianceManagementServiceStatus, prometheus.GaugeValue, 1.0)
	} else {
		ch <- prometheus.MustNewConstMetric(sc.applianceManagementServiceStatus, prometheus.GaugeValue, 0.0)
	}
}
