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
	registerCollector("transport_zone", newTransportZoneCollector)
}

type transportZoneCollector struct {
	transportZoneClient client.TransportZoneClient
	logger              log.Logger

	transportZoneLogicalPort   *prometheus.Desc
	transportZoneLogicalSwitch *prometheus.Desc
	transportZoneTransportNode *prometheus.Desc
}

func newTransportZoneCollector(apiClient *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	nsxtClient := client.NewNSXTClient(apiClient, logger)
	transportZoneLogicalPort := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transport_zone", "total_logical_port"),
		"Total number of logical port in transport zone",
		[]string{"id", "name"},
		nil,
	)
	transportZoneLogicalSwitch := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transport_zone", "total_logical_switch"),
		"Total number of logical switch in transport zone",
		[]string{"id", "name"},
		nil,
	)
	transportZoneTransportNode := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transport_zone", "total_transport_node"),
		"Total number of transport node in transport zone",
		[]string{"id", "name"},
		nil,
	)
	return &transportZoneCollector{
		transportZoneClient:        nsxtClient,
		logger:                     logger,
		transportZoneLogicalPort:   transportZoneLogicalPort,
		transportZoneLogicalSwitch: transportZoneLogicalSwitch,
		transportZoneTransportNode: transportZoneTransportNode,
	}
}

// Describe implements the prometheus.Collector interface.
func (c *transportZoneCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.transportZoneLogicalPort
	ch <- c.transportZoneLogicalSwitch
	ch <- c.transportZoneTransportNode
}

// Collect implements the prometheus.Collector interface.
func (c *transportZoneCollector) Collect(ch chan<- prometheus.Metric) {
	transportZones, err := c.transportZoneClient.ListAllTransportZones()
	if err != nil {
		level.Error(c.logger).Log("msg", "Unable to list transport zones", "err", err)
		return
	}
	c.collectTransportZonesStatus(transportZones, ch)
}

func (c *transportZoneCollector) collectTransportZonesStatus(transportZones []manager.TransportZone, ch chan<- prometheus.Metric) {
	for _, transportZone := range transportZones {
		transportZoneStatus, err := c.transportZoneClient.GetTransportZoneStatus(transportZone.Id)
		if err != nil {
			level.Error(c.logger).Log("msg", "Unable to get transport zone status", "id", transportZone.Id, "err", err)
			continue
		}
		transportZoneLabels := []string{transportZone.Id, transportZone.DisplayName}
		ch <- prometheus.MustNewConstMetric(c.transportZoneLogicalPort, prometheus.GaugeValue, float64(transportZoneStatus.NumLogicalPorts), transportZoneLabels...)
		ch <- prometheus.MustNewConstMetric(c.transportZoneLogicalSwitch, prometheus.GaugeValue, float64(transportZoneStatus.NumLogicalSwitches), transportZoneLabels...)
		ch <- prometheus.MustNewConstMetric(c.transportZoneTransportNode, prometheus.GaugeValue, float64(transportZoneStatus.NumTransportNodes), transportZoneLabels...)
	}
}