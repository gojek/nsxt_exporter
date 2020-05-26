package collector

import (
	"strconv"
	"strings"

	"nsxt_exporter/client"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	nsxt "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/manager"
)

var transportNodePossibleStatus = [...]string{"UP", "DOWN", "DEGRADED", "UNKNOWN"}

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

	transportNodeStatus   *prometheus.Desc
	edgeClusterMembership *prometheus.Desc
}

type transportNodeMetric struct {
	ID               string
	Name             string
	StatusDetail     map[string]float64
	Type             string
	TransportZoneIDs []string
}

func createTransportNodeCollectorFactory(apiClient *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	nsxtClient := client.NewNSXTClient(apiClient, logger)
	return newTransportNodeCollector(nsxtClient, logger)
}

func newTransportNodeCollector(transportNodeClient client.TransportNodeClient, logger log.Logger) *transportNodeCollector {
	transportNodeStatus := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transport_node", "status"),
		"Status of Transport Node",
		[]string{"id", "name", "type", "transport_zone_id", "status"},
		nil,
	)
	edgeClusterMembership := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transport_node", "edge_cluster_membership"),
		"Membership info of Transport Node in an Edge Cluster",
		[]string{"id", "edge_cluster_id", "edge_member_index"},
		nil,
	)
	return &transportNodeCollector{
		transportNodeClient:   transportNodeClient,
		logger:                logger,
		transportNodeStatus:   transportNodeStatus,
		edgeClusterMembership: edgeClusterMembership,
	}
}

// Describe implements the prometheus.Collector interface.
func (c *transportNodeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.transportNodeStatus
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
	for _, tnMetric := range transportNodeMetrics {
		for _, tzID := range tnMetric.TransportZoneIDs {
			for status, value := range tnMetric.StatusDetail {
				ch <- prometheus.MustNewConstMetric(c.transportNodeStatus, prometheus.GaugeValue, value, tnMetric.ID, tnMetric.Name, tnMetric.Type, tzID, status)
			}
		}
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
		statusDetail := map[string]float64{}
		for _, status := range transportNodePossibleStatus {
			statusValue := 0.0
			if status == strings.ToUpper(transportNodeStatus.Status) {
				statusValue = 1.0
			}
			statusDetail[status] = statusValue
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
			Type:             transportNodeType,
			TransportZoneIDs: transportZoneIDs,
			StatusDetail:     statusDetail,
		}
		transportNodeMetrics = append(transportNodeMetrics, transportNodeMetric)
	}
	return
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
