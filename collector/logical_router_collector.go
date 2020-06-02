package collector

import (
	"nsxt_exporter/client"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	nsxt "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/manager"
)

var logicalRouterPossibleHAStatus = [...]string{"ACTIVE", "STANDBY"}

func init() {
	registerCollector("logical_router", createLogicalRouterCollectorFactory)
}

type logicalRouterCollector struct {
	logicalRouterClient client.LogicalRouterClient
	logger              log.Logger

	logicalRouterStatus *prometheus.Desc
	natRuleTotalPackets *prometheus.Desc
	natRuleTotalBytes   *prometheus.Desc
}

type logicalRouterStatusMetric struct {
	ID                           string
	Name                         string
	TransportNodeID              string
	ServiceRouterID              string
	HighAvailabilityStatusDetail map[string]float64
}

type natRuleStatisticMetric struct {
	ID              string
	Name            string
	Type            string
	LogicalRouterID string
	NatTotalPackets float64
	NatTotalBytes   float64
}

func createLogicalRouterCollectorFactory(apiClient *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	nsxtClient := client.NewNSXTClient(apiClient, logger)
	return newLogicalRouterCollector(nsxtClient, logger)
}

func newLogicalRouterCollector(logicalRouterClient client.LogicalRouterClient, logger log.Logger) *logicalRouterCollector {
	logicalRouterStatus := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_router", "status"),
		"Status of logical router which includes high availability status associated with transport node",
		[]string{"id", "name", "transport_node_id", "service_router_id", "high_availability_status"},
		nil,
	)
	natRuleTotalPackets := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "nat_rule", "total_packets"),
		"Total packets processed by the NAT rule associated with logical router",
		[]string{"id", "name", "type", "logical_router_id"},
		nil,
	)
	natRuleTotalBytes := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "nat_rule", "total_bytes"),
		"Total bytes processed by the NAT rule associated with logical router",
		[]string{"id", "name", "type", "logical_router_id"},
		nil,
	)
	return &logicalRouterCollector{
		logicalRouterClient: logicalRouterClient,
		logger:              logger,
		logicalRouterStatus: logicalRouterStatus,
		natRuleTotalPackets: natRuleTotalPackets,
		natRuleTotalBytes:   natRuleTotalBytes,
	}
}

func (c *logicalRouterCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.logicalRouterStatus
	ch <- c.natRuleTotalPackets
	ch <- c.natRuleTotalBytes
}

func (c *logicalRouterCollector) Collect(ch chan<- prometheus.Metric) {
	logicalRouters, err := c.logicalRouterClient.ListAllLogicalRouters()
	if err != nil {
		level.Error(c.logger).Log("msg", "Unable to list logical routers", "err", err)
		return
	}
	logicalRouterStatusMetrics := c.generateLogicalRouterStatusMetrics(logicalRouters)
	for _, lrouterMetric := range logicalRouterStatusMetrics {
		for haStatus, value := range lrouterMetric.HighAvailabilityStatusDetail {
			labels := []string{lrouterMetric.ID, lrouterMetric.Name, lrouterMetric.TransportNodeID, lrouterMetric.ServiceRouterID, haStatus}
			ch <- prometheus.MustNewConstMetric(c.logicalRouterStatus, prometheus.GaugeValue, value, labels...)
		}
	}
	natRuleStatisticMetrics := c.generateNatRuleStatisticMetrics(logicalRouters)
	for _, natMetric := range natRuleStatisticMetrics {
		labels := []string{natMetric.ID, natMetric.Name, natMetric.Type, natMetric.LogicalRouterID}
		ch <- prometheus.MustNewConstMetric(c.natRuleTotalPackets, prometheus.GaugeValue, natMetric.NatTotalPackets, labels...)
		ch <- prometheus.MustNewConstMetric(c.natRuleTotalBytes, prometheus.GaugeValue, natMetric.NatTotalBytes, labels...)
	}
}

func (c *logicalRouterCollector) generateLogicalRouterStatusMetrics(logicalRouters []manager.LogicalRouter) (logicalRouterStatusMetrics []logicalRouterStatusMetric) {
	for _, logicalRouter := range logicalRouters {
		lrouterStatus, err := c.logicalRouterClient.GetLogicalRouterStatus(logicalRouter.Id)
		if err != nil {
			level.Error(c.logger).Log("msg", "Unable to get logical router status", "id", logicalRouter.Id, "err", err)
			continue
		}
		for _, status := range lrouterStatus.PerNodeStatus {
			logicalRouterStatusMetric := logicalRouterStatusMetric{
				ID:              logicalRouter.Id,
				Name:            logicalRouter.DisplayName,
				TransportNodeID: status.TransportNodeId,
				ServiceRouterID: status.ServiceRouterId,
			}
			logicalRouterStatusMetric.HighAvailabilityStatusDetail = make(map[string]float64)
			for _, haStatus := range logicalRouterPossibleHAStatus {
				statusValue := 0.0
				if haStatus == strings.ToUpper(status.HighAvailabilityStatus) {
					statusValue = 1.0
				}
				logicalRouterStatusMetric.HighAvailabilityStatusDetail[haStatus] = statusValue
			}
			logicalRouterStatusMetrics = append(logicalRouterStatusMetrics, logicalRouterStatusMetric)
		}
	}
	return
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
				Type:            rule.Action,
				LogicalRouterID: logicalRouter.Id,
				NatTotalPackets: float64(statistic.TotalPackets),
				NatTotalBytes:   float64(statistic.TotalBytes),
			}
			natRuleStatisticMetrics = append(natRuleStatisticMetrics, natRuleStatisticMetric)
		}
	}
	return
}
