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

func init() {
	registerCollector("dhcp", createDHCPCollectorFactory)
}

type dhcpCollector struct {
	dhcpClient client.DHCPClient
	logger     log.Logger

	dhcpStatus          *prometheus.Desc
	dhcpAckPacket       *prometheus.Desc
	dhcpDeclinePacket   *prometheus.Desc
	dhcpDiscoverPacket  *prometheus.Desc
	dhcpErrorTotal      *prometheus.Desc
	dhcpInformPacket    *prometheus.Desc
	dhcpNackPacket      *prometheus.Desc
	dhcpOfferPacket     *prometheus.Desc
	dhcpReleasePacket   *prometheus.Desc
	dhcpRequestPacket   *prometheus.Desc
	dhcpIPPoolSize      *prometheus.Desc
	dhcpIPPoolAllocated *prometheus.Desc
}

type dhcpStatusMetric struct {
	ID     string
	Name   string
	Status float64
}

type dhcpStatisticMetric struct {
	ID        string
	Name      string
	Statistic manager.DhcpStatistics
}

func createDHCPCollectorFactory(apiClient *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	nsxtClient := client.NewNSXTClient(apiClient, logger)
	return newDHCPCollector(nsxtClient, logger)
}

func newDHCPCollector(dhcpClient client.DHCPClient, logger log.Logger) *dhcpCollector {
	dhcpStatus := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dhcp", "status"),
		"Status of DCHP UP/DOWN",
		[]string{"id", "name"},
		nil,
	)
	dhcpAckPacket := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dhcp", "ack_packet"),
		"Number of DHCP ACK packets",
		[]string{"id", "name"},
		nil,
	)
	dhcpDeclinePacket := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dhcp", "decline_packet"),
		"Number of DHCP DECLINE packets",
		[]string{"id", "name"},
		nil,
	)
	dhcpDiscoverPacket := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dhcp", "discover_packet"),
		"Number of DHCP DISCOVER packets",
		[]string{"id", "name"},
		nil,
	)
	dhcpErrorTotal := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dhcp", "error_total"),
		"Number of DHCP errors",
		[]string{"id", "name"},
		nil,
	)
	dhcpInformPacket := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dhcp", "inform_packet"),
		"Number of DHCP INFORM packets",
		[]string{"id", "name"},
		nil,
	)
	dhcpNackPacket := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dhcp", "nack_packet"),
		"Number of DHCP NACK packets",
		[]string{"id", "name"},
		nil,
	)
	dhcpOfferPacket := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dhcp", "offer_packet"),
		"Number of DHCP OFFER packets",
		[]string{"id", "name"},
		nil,
	)
	dhcpReleasePacket := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dhcp", "release_packet"),
		"Number of DHCP RELEASE packets",
		[]string{"id", "name"},
		nil,
	)
	dhcpRequestPacket := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dhcp", "request_packet"),
		"Number of DHCP REQUEST packets",
		[]string{"id", "name"},
		nil,
	)
	dhcpIPPoolSize := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dhcp", "ip_pool_size_total"),
		"Total size of dhcp ip pool",
		[]string{"id", "dhcp_id"},
		nil,
	)
	dhcpIPPoolAllocated := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "dhcp", "ip_pool_allocated_total"),
		"Total number of allocated ip of dhcp ip pool",
		[]string{"id", "dhcp_id"},
		nil,
	)
	return &dhcpCollector{
		dhcpClient:          dhcpClient,
		logger:              logger,
		dhcpStatus:          dhcpStatus,
		dhcpAckPacket:       dhcpAckPacket,
		dhcpDeclinePacket:   dhcpDeclinePacket,
		dhcpDiscoverPacket:  dhcpDiscoverPacket,
		dhcpErrorTotal:      dhcpErrorTotal,
		dhcpInformPacket:    dhcpInformPacket,
		dhcpNackPacket:      dhcpNackPacket,
		dhcpOfferPacket:     dhcpOfferPacket,
		dhcpReleasePacket:   dhcpReleasePacket,
		dhcpRequestPacket:   dhcpRequestPacket,
		dhcpIPPoolSize:      dhcpIPPoolSize,
		dhcpIPPoolAllocated: dhcpIPPoolAllocated,
	}
}

// Describe implements the prometheus.Collector interface.
func (dc *dhcpCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- dc.dhcpStatus
	ch <- dc.dhcpAckPacket
	ch <- dc.dhcpDeclinePacket
	ch <- dc.dhcpDiscoverPacket
	ch <- dc.dhcpErrorTotal
	ch <- dc.dhcpInformPacket
	ch <- dc.dhcpNackPacket
	ch <- dc.dhcpOfferPacket
	ch <- dc.dhcpReleasePacket
	ch <- dc.dhcpRequestPacket
	ch <- dc.dhcpIPPoolSize
	ch <- dc.dhcpIPPoolAllocated
}

// Collect implements the prometheus.Collector interface.
func (dc *dhcpCollector) Collect(ch chan<- prometheus.Metric) {
	dhcpServers, err := dc.dhcpClient.ListAllDHCPServers()
	if err != nil {
		level.Error(dc.logger).Log("msg", "Unable to list dhcp servers", "err", err)
		return
	}
	dhcpStatusMetrics := dc.generateDHCPStatusMetrics(dhcpServers)
	for _, m := range dhcpStatusMetrics {
		ch <- prometheus.MustNewConstMetric(dc.dhcpStatus, prometheus.GaugeValue, m.Status, m.ID, m.Name)
	}
	dhcpStatisticMetrics := dc.generateDHCPStatisticMetrics(dhcpServers)
	for _, m := range dhcpStatisticMetrics {
		dhcpLabels := []string{m.ID, m.Name}
		ch <- prometheus.MustNewConstMetric(dc.dhcpAckPacket, prometheus.GaugeValue, float64(m.Statistic.Acks), dhcpLabels...)
		ch <- prometheus.MustNewConstMetric(dc.dhcpDeclinePacket, prometheus.GaugeValue, float64(m.Statistic.Declines), dhcpLabels...)
		ch <- prometheus.MustNewConstMetric(dc.dhcpDiscoverPacket, prometheus.GaugeValue, float64(m.Statistic.Discovers), dhcpLabels...)
		ch <- prometheus.MustNewConstMetric(dc.dhcpErrorTotal, prometheus.GaugeValue, float64(m.Statistic.Errors), dhcpLabels...)
		ch <- prometheus.MustNewConstMetric(dc.dhcpInformPacket, prometheus.GaugeValue, float64(m.Statistic.Informs), dhcpLabels...)
		ch <- prometheus.MustNewConstMetric(dc.dhcpNackPacket, prometheus.GaugeValue, float64(m.Statistic.Nacks), dhcpLabels...)
		ch <- prometheus.MustNewConstMetric(dc.dhcpOfferPacket, prometheus.GaugeValue, float64(m.Statistic.Offers), dhcpLabels...)
		ch <- prometheus.MustNewConstMetric(dc.dhcpReleasePacket, prometheus.GaugeValue, float64(m.Statistic.Releases), dhcpLabels...)
		ch <- prometheus.MustNewConstMetric(dc.dhcpRequestPacket, prometheus.GaugeValue, float64(m.Statistic.Requests), dhcpLabels...)
		for _, ipPoolStat := range m.Statistic.IpPoolStats {
			ipPoolLabels := []string{ipPoolStat.DhcpIpPoolId, m.ID}
			ch <- prometheus.MustNewConstMetric(dc.dhcpIPPoolSize, prometheus.GaugeValue, float64(ipPoolStat.PoolSize), ipPoolLabels...)
			ch <- prometheus.MustNewConstMetric(dc.dhcpIPPoolAllocated, prometheus.GaugeValue, float64(ipPoolStat.AllocatedNumber), ipPoolLabels...)
		}
	}
}

func (dc *dhcpCollector) generateDHCPStatusMetrics(dhcpServers []manager.LogicalDhcpServer) (dhcpStatusMetrics []dhcpStatusMetric) {
	for _, dhcp := range dhcpServers {
		dhcpStatus, err := dc.dhcpClient.GetDhcpStatus(dhcp.Id, nil)
		if err != nil {
			level.Error(dc.logger).Log("msg", "Unable to get dhcp status", "id", dhcp.Id, "err", err)
			continue
		}
		var status float64
		if strings.ToUpper(dhcpStatus.ServiceStatus) == "UP" {
			status = 1.0
		} else {
			status = 0.0
		}
		dhcpStatusMetric := dhcpStatusMetric{
			Name:   dhcp.DisplayName,
			ID:     dhcp.Id,
			Status: status,
		}
		dhcpStatusMetrics = append(dhcpStatusMetrics, dhcpStatusMetric)
	}
	return
}

func (dc *dhcpCollector) generateDHCPStatisticMetrics(dhcpServers []manager.LogicalDhcpServer) (dhcpStatisticMetrics []dhcpStatisticMetric) {
	for _, dhcp := range dhcpServers {
		dhcpStatistic, err := dc.dhcpClient.GetDHCPStatistic(dhcp.Id)
		if err != nil {
			level.Error(dc.logger).Log("msg", "Unable to get dhcp statistic", "id", dhcp.Id, "err", err)
			continue
		}
		dhcpStatisticMetric := dhcpStatisticMetric{
			ID:        dhcp.Id,
			Name:      dhcp.DisplayName,
			Statistic: dhcpStatistic,
		}
		dhcpStatisticMetrics = append(dhcpStatisticMetrics, dhcpStatisticMetric)
	}
	return
}
