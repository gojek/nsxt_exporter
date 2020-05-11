package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	nsxt "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/manager"
)

func init() {
	registerCollector("logical_router_port", newLogicalRouterPortCollector)
}

type logicalRouterPortCollector struct {
	client *nsxt.APIClient
	logger log.Logger

	rxTotalPacketDesc   *prometheus.Desc
	rxDroppedPacketDesc *prometheus.Desc
	rxTotalByteDesc     *prometheus.Desc
	txTotalPacketDesc   *prometheus.Desc
	txDroppedPacketDesc *prometheus.Desc
	txTotalByteDesc     *prometheus.Desc
}

func newLogicalRouterPortCollector(client *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	rxTotalPacketDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_router_port", "rx_total_packet"),
		"Total packets of logical router port rx",
		[]string{"id", "name", "logical_router_id"},
		nil,
	)
	rxDroppedPacketDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_router_port", "rx_dropped_packet"),
		"Total dropped packets of logical router port rx",
		[]string{"id", "name", "logical_router_id"},
		nil,
	)
	rxTotalByteDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_router_port", "rx_total_byte"),
		"Total bytes of logical router port rx",
		[]string{"id", "name", "logical_router_id"},
		nil,
	)
	txTotalPacketDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_router_port", "tx_total_packet"),
		"Total packets of logical router port tx",
		[]string{"id", "name", "logical_router_id"},
		nil,
	)
	txDroppedPacketDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_router_port", "tx_dropped_packet"),
		"Total dropped packets of logical router port tx",
		[]string{"id", "name", "logical_router_id"},
		nil,
	)
	txTotalByteDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_router_port", "tx_total_byte"),
		"Total bytes of logical router port tx",
		[]string{"id", "name", "logical_router_id"},
		nil,
	)
	return &logicalRouterPortCollector{
		client:              client,
		logger:              logger,
		rxTotalPacketDesc:   rxTotalPacketDesc,
		rxTotalByteDesc:     rxTotalByteDesc,
		rxDroppedPacketDesc: rxDroppedPacketDesc,
		txTotalPacketDesc:   txTotalPacketDesc,
		txTotalByteDesc:     txTotalByteDesc,
		txDroppedPacketDesc: txDroppedPacketDesc,
	}
}

// Describe implements the prometheus.Collector interface.
func (c *logicalRouterPortCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.rxTotalPacketDesc
	ch <- c.rxDroppedPacketDesc
	ch <- c.rxTotalByteDesc
	ch <- c.txTotalPacketDesc
	ch <- c.txDroppedPacketDesc
	ch <- c.txTotalByteDesc
}

// Collect implements the prometheus.Collector interface.
func (c *logicalRouterPortCollector) Collect(ch chan<- prometheus.Metric) {
	logicalRouterPortStatisticMetrics := c.generateLogicalRouterPortStatisticMetrics()
	for _, metric := range logicalRouterPortStatisticMetrics {
		ch <- metric
	}
}

func (c *logicalRouterPortCollector) generateLogicalRouterPortStatisticMetrics() (logicalRouterPortStatisticMetrics []prometheus.Metric) {
	var logicalRouterPorts []manager.LogicalRouterPort
	var cursor string
	for {
		localVarOptionals := make(map[string]interface{})
		localVarOptionals["cursor"] = cursor
		logicalRouterPortsResult, _, err := c.client.LogicalRoutingAndServicesApi.ListLogicalRouterPorts(c.client.Context, nil)
		if err != nil {
			level.Error(c.logger).Log("msg", "Unable to list logical ports", "err", err)
			return
		}
		logicalRouterPorts = append(logicalRouterPorts, logicalRouterPortsResult.Results...)
		cursor = logicalRouterPortsResult.Cursor
		if len(cursor) == 0 {
			break
		}
	}

	for _, logicalRouterPort := range logicalRouterPorts {
		statistic, _, err := c.client.LogicalRoutingAndServicesApi.GetLogicalRouterPortStatisticsSummary(c.client.Context, logicalRouterPort.Id, nil)
		if err != nil {
			level.Error(c.logger).Log("msg", "Unable to get logical router port statistics", "id", logicalRouterPort.Id, "err", err)
			continue
		}
		rxTotalPacketMetric := prometheus.MustNewConstMetric(
			c.rxTotalPacketDesc,
			prometheus.GaugeValue,
			float64(statistic.Rx.TotalPackets),
			logicalRouterPort.Id,
			logicalRouterPort.DisplayName,
			logicalRouterPort.LogicalRouterId,
		)
		rxDroppedPacketMetric := prometheus.MustNewConstMetric(
			c.rxDroppedPacketDesc,
			prometheus.GaugeValue,
			float64(statistic.Rx.DroppedPackets),
			logicalRouterPort.Id,
			logicalRouterPort.DisplayName,
			logicalRouterPort.LogicalRouterId,
		)
		rxTotalByteMetric := prometheus.MustNewConstMetric(
			c.rxTotalByteDesc,
			prometheus.GaugeValue,
			float64(statistic.Rx.TotalBytes),
			logicalRouterPort.Id,
			logicalRouterPort.DisplayName,
			logicalRouterPort.LogicalRouterId,
		)
		txTotalPacketMetric := prometheus.MustNewConstMetric(
			c.txTotalPacketDesc,
			prometheus.GaugeValue,
			float64(statistic.Tx.TotalPackets),
			logicalRouterPort.Id,
			logicalRouterPort.DisplayName,
			logicalRouterPort.LogicalRouterId,
		)
		txDroppedPacketMetric := prometheus.MustNewConstMetric(
			c.txDroppedPacketDesc,
			prometheus.GaugeValue,
			float64(statistic.Tx.DroppedPackets),
			logicalRouterPort.Id,
			logicalRouterPort.DisplayName,
			logicalRouterPort.LogicalRouterId,
		)
		txTotalByteMetric := prometheus.MustNewConstMetric(
			c.txTotalByteDesc,
			prometheus.GaugeValue,
			float64(statistic.Tx.TotalBytes),
			logicalRouterPort.Id,
			logicalRouterPort.DisplayName,
			logicalRouterPort.LogicalRouterId,
		)
		logicalRouterPortStatisticMetrics = append(logicalRouterPortStatisticMetrics, rxTotalPacketMetric)
		logicalRouterPortStatisticMetrics = append(logicalRouterPortStatisticMetrics, rxDroppedPacketMetric)
		logicalRouterPortStatisticMetrics = append(logicalRouterPortStatisticMetrics, rxTotalByteMetric)
		logicalRouterPortStatisticMetrics = append(logicalRouterPortStatisticMetrics, txTotalPacketMetric)
		logicalRouterPortStatisticMetrics = append(logicalRouterPortStatisticMetrics, txDroppedPacketMetric)
		logicalRouterPortStatisticMetrics = append(logicalRouterPortStatisticMetrics, txTotalByteMetric)
	}
	return
}
