package collector

import (
	"nsxt_exporter/client"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	nsxt "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/manager"
)

func init() {
	registerCollector("logical_router", createLogicalRouterCollectorFactory)
}

type logicalRouterCollector struct {
	logicalRouterClient client.LogicalRouterClient
	logger              log.Logger

	natTotalPackets *prometheus.Desc
	natTotalBytes   *prometheus.Desc
}

type logicalRouterNatStatisticMetric struct {
	LogicalRouterID string
	NatTotalPackets float64
	NatTotalBytes   float64
}

func createLogicalRouterCollectorFactory(apiClient *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	nsxtClient := client.NewNSXTClient(apiClient, logger)
	return newLogicalRouterCollector(nsxtClient, logger)
}

func newLogicalRouterCollector(logicalRouterClient client.LogicalRouterClient, logger log.Logger) *logicalRouterCollector {
	natTotalPackets := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_router", "nat_total_packets"),
		"Total packets processed by the NAT rule associated with logical router",
		[]string{"logical_router_id"},
		nil,
	)
	natTotalBytes := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_router", "nat_total_bytes"),
		"Total bytes processed by the NAT rule associated with logical router",
		[]string{"logical_router_id"},
		nil,
	)
	return &logicalRouterCollector{
		logicalRouterClient: logicalRouterClient,
		logger:              logger,
		natTotalPackets:     natTotalPackets,
		natTotalBytes:       natTotalBytes,
	}
}

func (c *logicalRouterCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.natTotalPackets
	ch <- c.natTotalBytes
}

func (c *logicalRouterCollector) Collect(ch chan<- prometheus.Metric) {
	logicalRouters, err := c.logicalRouterClient.ListAllLogicalRouters()
	if err != nil {
		level.Error(c.logger).Log("msg", "Unable to list logical routers", "err", err)
		return
	}
	logicalRouterNatStatisticMetrics := c.generateLogicalRouterNatStatisticMetrics(logicalRouters)
	for _, metric := range logicalRouterNatStatisticMetrics {
		ch <- prometheus.MustNewConstMetric(c.natTotalPackets, prometheus.GaugeValue, metric.NatTotalPackets, metric.LogicalRouterID)
		ch <- prometheus.MustNewConstMetric(c.natTotalBytes, prometheus.GaugeValue, metric.NatTotalBytes, metric.LogicalRouterID)
	}
}

func (c *logicalRouterCollector) generateLogicalRouterNatStatisticMetrics(logicalRouters []manager.LogicalRouter) (logicalRouterNatStatisticMetrics []logicalRouterNatStatisticMetric) {
	for _, logicalRouter := range logicalRouters {
		statistic, err := c.logicalRouterClient.GetNatStatisticsPerLogicalRouter(logicalRouter.Id)
		if err != nil {
			level.Error(c.logger).Log("msg", "Unable to get logical router statistics", "id", logicalRouter.Id, "err", err)
			continue
		}
		logicalRouterNatStatisticMetric := logicalRouterNatStatisticMetric{
			LogicalRouterID: logicalRouter.Id,
			NatTotalPackets: float64(statistic.StatisticsAcrossAllNodes.TotalPackets),
			NatTotalBytes:   float64(statistic.StatisticsAcrossAllNodes.TotalBytes),
		}
		logicalRouterNatStatisticMetrics = append(logicalRouterNatStatisticMetrics, logicalRouterNatStatisticMetric)
	}
	return
}
