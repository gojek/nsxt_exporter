package collector

import (
	"regexp"
	"strings"

	"nsxt_exporter/client"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	nsxt "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/manager"
)

func init() {
	registerCollector("transport_node", newTransportNodeCollector)
}

type transportNodeCollector struct {
	transportNodeClient client.TransportNodeClient
	logger              log.Logger

	transportNodeStatus *prometheus.Desc
}

func newTransportNodeCollector(apiClient *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	nsxtClient := client.NewNSXTClient(apiClient, logger)
	transportNodeStatus := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transport_node", "status"),
		"Status of Transport Node UP/DOWN",
		[]string{"id", "name"},
		nil,
	)
	return &transportNodeCollector{
		transportNodeClient: nsxtClient,
		logger:              logger,
		transportNodeStatus: transportNodeStatus,
	}
}

// Describe implements the prometheus.Collector interface.
func (c *transportNodeCollector) Describe(ch chan<- *prometheus.Desc) {
}

// Collect implements the prometheus.Collector interface.
func (c *transportNodeCollector) Collect(ch chan<- prometheus.Metric) {
	metrics := c.generateTransportNodeStatusMetrics()
	for _, metric := range metrics {
		ch <- metric
	}
}

func (c *transportNodeCollector) generateTransportNodeStatusMetrics() (transportNodeMetrics []prometheus.Metric) {
	var transportNodes []manager.TransportNode
	var cursor string
	for {
		localVarOptionals := make(map[string]interface{})
		localVarOptionals["cursor"] = cursor
		transportNodesResult, err := c.transportNodeClient.ListTransportNodes(localVarOptionals)
		if err != nil {
			level.Error(c.logger).Log("msg", "Unable to list transport nodes", "err", err)
			return
		}
		transportNodes = append(transportNodes, transportNodesResult.Results...)
		cursor = transportNodesResult.Cursor
		if len(cursor) == 0 {
			break
		}
	}

	for _, transportNode := range transportNodes {
		transportNodeStatus, err := c.transportNodeClient.GetTransportNodeStatus(transportNode.Id)
		if err != nil {
			level.Error(c.logger).Log("msg", "Unable to get transport node status", "id", transportNode.Id, "err", err)
			continue
		}
		var status float64
		if strings.ToUpper(transportNodeStatus.Status) == "UP" {
			status = 1
		} else {
			status = 0
		}
		transportZoneLabels := c.buildTransportZoneEndpointLabels(transportNode.TransportZoneEndpoints)
		desc := c.buildStatusDesc(transportZoneLabels)
		transportNodeStatusMetric := prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, status, transportNode.Id, transportNode.DisplayName)
		transportNodeMetrics = append(transportNodeMetrics, transportNodeStatusMetric)
	}
	return
}

func (c *transportNodeCollector) buildStatusDesc(extraLabels prometheus.Labels) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transport_node", "status"),
		"Status of Transport Node UP/DOWN",
		[]string{"id", "name"},
		extraLabels,
	)
}

func (c *transportNodeCollector) buildTransportZoneEndpointLabels(transportZoneEndpoints []manager.TransportZoneEndPoint) prometheus.Labels {
	labels := make(prometheus.Labels)
	for _, endpoint := range transportZoneEndpoints {
		sanitizedID := regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(endpoint.TransportZoneId, "_")
		key := "transport_zone_" + sanitizedID
		value := endpoint.TransportZoneId
		labels[key] = value
	}
	return labels
}
