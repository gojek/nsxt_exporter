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

	natRuleTotalPackets *prometheus.Desc
	natRuleTotalBytes   *prometheus.Desc
}

type natRuleStatisticMetric struct {
	ID              string
	Name            string
	LogicalRouterID string
	NatTotalPackets float64
	NatTotalBytes   float64
}

func createLogicalRouterCollectorFactory(apiClient *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	nsxtClient := client.NewNSXTClient(apiClient, logger)
	return newLogicalRouterCollector(nsxtClient, logger)
}

func newLogicalRouterCollector(logicalRouterClient client.LogicalRouterClient, logger log.Logger) *logicalRouterCollector {
	natRuleTotalPackets := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "nat_rule", "total_packets"),
		"Total packets processed by the NAT rule associated with logical router",
		[]string{"id", "name", "logical_router_id"},
		nil,
	)
	natRuleTotalBytes := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "nat_rule", "total_bytes"),
		"Total bytes processed by the NAT rule associated with logical router",
		[]string{"id", "name", "logical_router_id"},
		nil,
	)
	return &logicalRouterCollector{
		logicalRouterClient: logicalRouterClient,
		logger:              logger,
		natRuleTotalPackets: natRuleTotalPackets,
		natRuleTotalBytes:   natRuleTotalBytes,
	}
}

func (c *logicalRouterCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.natRuleTotalPackets
	ch <- c.natRuleTotalBytes
}

func (c *logicalRouterCollector) Collect(ch chan<- prometheus.Metric) {
	logicalRouters, err := c.logicalRouterClient.ListAllLogicalRouters()
	if err != nil {
		level.Error(c.logger).Log("msg", "Unable to list logical routers", "err", err)
		return
	}
	natRuleStatisticMetrics := c.generateNatRuleStatisticMetrics(logicalRouters)
	for _, metric := range natRuleStatisticMetrics {
		labels := []string{metric.ID, metric.Name, metric.LogicalRouterID}
		ch <- prometheus.MustNewConstMetric(c.natRuleTotalPackets, prometheus.GaugeValue, metric.NatTotalPackets, labels...)
		ch <- prometheus.MustNewConstMetric(c.natRuleTotalBytes, prometheus.GaugeValue, metric.NatTotalBytes, labels...)
	}
}

func (c *logicalRouterCollector) generateNatRuleStatisticMetrics(logicalRouters []manager.LogicalRouter) (natRuleStatisticMetrics []natRuleStatisticMetric) {
	for _, logicalRouter := range logicalRouters {
		natRules, err := c.logicalRouterClient.ListAllNatRules(logicalRouter.Id)
		if err != nil {
			level.Error(c.logger).Log("msg", "Unable to get nat rules from logical router", "id", logicalRouter.Id, "err", err)
			continue
		}
		for _, rule := range natRules {
			statistic, err := c.logicalRouterClient.GetNatStatisticsPerRule(logicalRouter.Id, rule.Id)
			if err != nil {
				level.Error(c.logger).Log("msg", "Unable to get nat rule statistics", "id", rule.Id, "logicalRouterID", logicalRouter.Id, "err", err)
				continue
			}
			natRuleStatisticMetric := natRuleStatisticMetric{
				ID:              rule.Id,
				Name:            rule.DisplayName,
				LogicalRouterID: logicalRouter.Id,
				NatTotalPackets: float64(statistic.TotalPackets),
				NatTotalBytes:   float64(statistic.TotalBytes),
			}
			natRuleStatisticMetrics = append(natRuleStatisticMetrics, natRuleStatisticMetric)
		}
	}
	return
}
