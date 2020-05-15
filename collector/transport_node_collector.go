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
	registerCollector("transport_node", createTransportNodeCollectorFactory)
}

type edgeClusterMembership struct {
	transportNodeID string
	edgeMemberIndex string
	edgeClusterID   string
}

type transportNodeCollector struct {
	transportNodeClient client.TransportNodeClient
	logger              log.Logger

	edgeClusterMembership *prometheus.Desc
}

type transportNodeMetric struct {
	ID               string
	Name             string
	Status           float64
	Type             string
	TransportZoneIDs []string
}

func createTransportNodeCollectorFactory(apiClient *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	nsxtClient := client.NewNSXTClient(apiClient, logger)
	return newTransportNodeCollector(nsxtClient, logger)
}

func newTransportNodeCollector(transportNodeClient client.TransportNodeClient, logger log.Logger) *transportNodeCollector {
	edgeClusterMembership := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transport_node", "edge_cluster_membership"),
		"Membership info of Transport Node in an Edge Cluster",
		[]string{"id", "edge_cluster_id", "edge_member_index"},
		nil,
	)
	return &transportNodeCollector{
		transportNodeClient:   transportNodeClient,
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
	transportNodes, err := c.transportNodeClient.ListAllTransportNodes()
	if err != nil {
		level.Error(c.logger).Log("msg", "Unable to list transport nodes", "err", err)
		return
	}
	edgeClusterMemberships, err := c.generateEdgeClusterMemberships()
	if err != nil {
		edgeClusterMemberships = nil
		level.Error(c.logger).Log("msg", "Unable to generate edge cluster membership", "err", err)
	}
	for _, membership := range edgeClusterMemberships {
		ch <- c.buildEdgeClusterMembershipMetrics(membership)
	}
	transportNodeMetrics := c.generateTransportNodeMetrics(transportNodes, edgeClusterMemberships)
	for _, metric := range transportNodeMetrics {
		transportZoneLabels := c.buildTransportZoneLabels(metric.TransportZoneIDs)
		desc := c.buildStatusDesc(transportZoneLabels)
		ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, metric.Status, metric.ID, metric.Name, metric.Type)
	}
}

func (c *transportNodeCollector) buildEdgeClusterMembershipMetrics(membership edgeClusterMembership) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		c.edgeClusterMembership,
		prometheus.GaugeValue,
		1.0,
		membership.transportNodeID,
		membership.edgeClusterID,
		membership.edgeMemberIndex,
	)
}

func (c *transportNodeCollector) generateTransportNodeMetrics(transportNodes []manager.TransportNode, edgeClusterMemberships []edgeClusterMembership) (transportNodeMetrics []transportNodeMetric) {
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
		if edgeClusterMemberships != nil {
			transportNodeType = "host"
		}
		for _, membership := range edgeClusterMemberships {
			if membership.transportNodeID == transportNode.Id {
				transportNodeType = "edge"
				break
			}
		}
		var transportZoneIDs []string
		for _, endpoint := range transportNode.TransportZoneEndpoints {
			transportZoneIDs = append(transportZoneIDs, endpoint.TransportZoneId)
		}
		transportNodeMetric := transportNodeMetric{
			ID:               transportNode.Id,
			Name:             transportNode.DisplayName,
			Status:           status,
			Type:             transportNodeType,
			TransportZoneIDs: transportZoneIDs,
		}
		transportNodeMetrics = append(transportNodeMetrics, transportNodeMetric)
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

func (c *transportNodeCollector) buildTransportZoneLabels(transportZoneIDs []string) prometheus.Labels {
	labels := make(prometheus.Labels)
	for _, transportZoneID := range transportZoneIDs {
		sanitizedID := regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(transportZoneID, "_")
		key := "transport_zone_" + sanitizedID
		value := transportZoneID
		labels[key] = value
	}
	return labels
}

func (c *transportNodeCollector) generateEdgeClusterMemberships() ([]edgeClusterMembership, error) {
	var edgeClusterMemberships []edgeClusterMembership
	edgeClusters, err := c.transportNodeClient.ListAllEdgeClusters()
	if err != nil {
		return nil, err
	}

	for _, ec := range edgeClusters {
		for _, member := range ec.Members {
			edgeClusterMembership := edgeClusterMembership{
				transportNodeID: member.TransportNodeId,
				edgeMemberIndex: strconv.Itoa(int(member.MemberIndex)),
				edgeClusterID:   ec.Id,
			}
			edgeClusterMemberships = append(edgeClusterMemberships, edgeClusterMembership)
		}
	}
	return edgeClusterMemberships, nil
}
