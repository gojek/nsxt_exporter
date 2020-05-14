package collector

import (
	"regexp"
	"strconv"
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

type edgeClusterMembership struct {
	edgeMemberIndex string
	edgeClusterID   string
}

type transportNodeCollector struct {
	transportNodeClient client.TransportNodeClient
	logger              log.Logger

	edgeClusterMembership *prometheus.Desc
}

func newTransportNodeCollector(apiClient *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	nsxtClient := client.NewNSXTClient(apiClient, logger)
	edgeClusterMembership := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transport_node", "edge_cluster_membership"),
		"Membership info of Transport Node in an Edge Cluster",
		[]string{"id", "edge_cluster_id", "edge_member_index"},
		nil,
	)
	return &transportNodeCollector{
		transportNodeClient:   nsxtClient,
		logger:                logger,
		edgeClusterMembership: edgeClusterMembership,
	}
}

// Describe implements the prometheus.Collector interface.
func (c *transportNodeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.edgeClusterMembership
}

// Collect implements the prometheus.Collector interface.
func (c *transportNodeCollector) Collect(ch chan<- prometheus.Metric) {
	transportNodeMetrics := c.generateTransportNodeMetrics()
	for _, metric := range transportNodeMetrics {
		ch <- metric
	}
}

func (c *transportNodeCollector) generateTransportNodeMetrics() (transportNodeMetrics []prometheus.Metric) {
	transportNodes, err := c.transportNodeClient.ListAllTransportNodes()
	if err != nil {
		level.Error(c.logger).Log("msg", "Unable to list transport nodes", "err", err)
		return
	}
	edgeClusterMap, err := c.initEdgeClusterMap()
	if err != nil {
		level.Error(c.logger).Log("msg", "Unable to initalize edge cluster map", "err", err)
		return
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

		var transportNodeType string
		if e, ok := edgeClusterMap[transportNode.Id]; ok {
			edgeClusterMembershipMetric := prometheus.MustNewConstMetric(c.edgeClusterMembership, prometheus.GaugeValue, 1.0, transportNode.Id, e.edgeClusterID, e.edgeMemberIndex)
			transportNodeMetrics = append(transportNodeMetrics, edgeClusterMembershipMetric)
			transportNodeType = "edge"
		} else {
			transportNodeType = "host"
		}

		transportZoneLabels := c.buildTransportZoneEndpointLabels(transportNode.TransportZoneEndpoints)
		desc := c.buildStatusDesc(transportZoneLabels)
		transportNodeStatusMetric := prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, status, transportNode.Id, transportNode.DisplayName, transportNodeType)
		transportNodeMetrics = append(transportNodeMetrics, transportNodeStatusMetric)
	}
	return
}

func (c *transportNodeCollector) buildStatusDesc(extraLabels prometheus.Labels) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transport_node", "status"),
		"Status of Transport Node UP/DOWN",
		[]string{"id", "name", "type"},
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

func (c *transportNodeCollector) initEdgeClusterMap() (map[string]edgeClusterMembership, error) {
	edgeClusterMap := make(map[string]edgeClusterMembership)
	edgeClusters, err := c.transportNodeClient.ListAllEdgeClusters()
	if err != nil {
		return nil, err
	}

	for _, ec := range edgeClusters {
		for _, member := range ec.Members {
			edgeClusterMembership := edgeClusterMembership{
				edgeMemberIndex: strconv.Itoa(int(member.MemberIndex)),
				edgeClusterID:   ec.Id,
			}
			edgeClusterMap[member.TransportNodeId] = edgeClusterMembership
		}
	}
	return edgeClusterMap, nil
}
