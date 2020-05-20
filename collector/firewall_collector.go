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
	registerCollector("firewall", createFirewallCollectorFactory)
}

type firewallCollector struct {
	firewallClient client.FirewallClient
	logger         log.Logger

	totalPackets *prometheus.Desc
	totalBytes   *prometheus.Desc
}

type firewallStatisticMetric struct {
	SectionID    string
	RuleID       string
	RuleName     string
	TotalPackets float64
	TotalBytes   float64
}

func createFirewallCollectorFactory(apiClient *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	nsxtClient := client.NewNSXTClient(apiClient, logger)
	return newFirewallCollector(nsxtClient, logger)
}

func newFirewallCollector(firewallClient client.FirewallClient, logger log.Logger) *firewallCollector {
	totalPackets := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "firewall", "total_packets"),
		"Total packets processed by the firewall rule",
		[]string{"id", "name", "section_id"},
		nil,
	)
	totalBytes := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "firewall", "total_bytes"),
		"Total bytes processed by the firewall rule",
		[]string{"id", "name", "section_id"},
		nil,
	)
	return &firewallCollector{
		firewallClient: firewallClient,
		logger:         logger,
		totalPackets:   totalPackets,
		totalBytes:     totalBytes,
	}
}

// Describe implements the prometheus.Collector interface.
func (c *firewallCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.totalPackets
	ch <- c.totalBytes
}

// Collect implements the prometheus.Collector interface.
func (c *firewallCollector) Collect(ch chan<- prometheus.Metric) {
	firewallSections, err := c.firewallClient.ListAllSections()
	if err != nil {
		level.Error(c.logger).Log("msg", "Unable to list firewall sections", "err", err)
	}
	firewallStatisticMetrics := c.generateFirewallStatisticMetrics(firewallSections)
	for _, m := range firewallStatisticMetrics {
		labels := []string{m.RuleID, m.RuleName, m.SectionID}
		ch <- prometheus.MustNewConstMetric(c.totalPackets, prometheus.GaugeValue, m.TotalPackets, labels...)
		ch <- prometheus.MustNewConstMetric(c.totalBytes, prometheus.GaugeValue, m.TotalBytes, labels...)
	}
}

func (c *firewallCollector) generateFirewallStatisticMetrics(firewallSections []manager.FirewallSection) (firewallStatisticMetrics []firewallStatisticMetric) {
	for _, sec := range firewallSections {
		rules, err := c.firewallClient.GetAllRules(sec.Id)
		if err != nil {
			level.Error(c.logger).Log("msg", "Unable to get firewall rules", "section", sec.Id, "err", err)
			continue
		}
		for _, rule := range rules {
			stats, err := c.firewallClient.GetFirewallStats(sec.Id, rule.Id)
			if err != nil {
				level.Error(c.logger).Log("msg", "Unable to get firewall statistic", "section", sec.Id, "rule", rule.Id, "err", err)
				continue
			}
			firewallStatisticMetric := firewallStatisticMetric{
				SectionID:    sec.Id,
				RuleID:       rule.Id,
				RuleName:     rule.DisplayName,
				TotalPackets: float64(stats.PacketCount),
				TotalBytes:   float64(stats.ByteCount),
			}
			firewallStatisticMetrics = append(firewallStatisticMetrics, firewallStatisticMetric)
		}
	}
	return
}
