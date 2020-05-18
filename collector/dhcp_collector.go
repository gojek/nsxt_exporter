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

	dhcpStatus *prometheus.Desc
}

type dhcpStatusMetric struct {
	ID     string
	Name   string
	Status float64
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
	return &dhcpCollector{
		dhcpClient: dhcpClient,
		logger:     logger,
		dhcpStatus: dhcpStatus,
	}
}

// Describe implements the prometheus.Collector interface.
func (dc *dhcpCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- dc.dhcpStatus
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
