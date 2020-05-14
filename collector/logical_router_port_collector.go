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
	registerCollector("logical_router_port", newLogicalRouterPortCollector)
}

type logicalRouterPortCollector struct {
	logicalRouterPortClient client.LogicalRouterPortClient
	logger                  log.Logger

	rxTotalPacket   *prometheus.Desc
	rxDroppedPacket *prometheus.Desc
	rxTotalByte     *prometheus.Desc
	txTotalPacket   *prometheus.Desc
	txDroppedPacket *prometheus.Desc
	txTotalByte     *prometheus.Desc
}

type logicalRouterPortStatisticMetric struct {
	LogicalRouterPort manager.LogicalRouterPort
	Rx                *manager.LogicalRouterPortCounters
	Tx                *manager.LogicalRouterPortCounters
}

func newLogicalRouterPortCollector(apiClient *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	nsxtClient := client.NewNSXTClient(apiClient, logger)
	rxTotalPacket := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_router_port", "rx_total_packet"),
		"Total packets received (rx) of logical router port",
		[]string{"id", "name", "logical_router_id"},
		nil,
	)
	rxDroppedPacket := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_router_port", "rx_dropped_packet"),
		"Total receive (rx) packets dropped of logical router port",
		[]string{"id", "name", "logical_router_id"},
		nil,
	)
	rxTotalByte := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_router_port", "rx_total_byte"),
		"Total bytes received (rx)  of logical router port rx",
		[]string{"id", "name", "logical_router_id"},
		nil,
	)
	txTotalPacket := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_router_port", "tx_total_packet"),
		"Total packets transmitted (rx) of logical router port",
		[]string{"id", "name", "logical_router_id"},
		nil,
	)
	txDroppedPacket := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_router_port", "tx_dropped_packet"),
		"Total transmit (tx) packets dropped of logical router port tx",
		[]string{"id", "name", "logical_router_id"},
		nil,
	)
	txTotalByte := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_router_port", "tx_total_byte"),
		"Total bytes transmitted (tx) of logical router port",
		[]string{"id", "name", "logical_router_id"},
		nil,
	)
	return &logicalRouterPortCollector{
		logicalRouterPortClient: nsxtClient,
		logger:                  logger,
		rxTotalPacket:           rxTotalPacket,
		rxTotalByte:             rxTotalByte,
		rxDroppedPacket:         rxDroppedPacket,
		txTotalPacket:           txTotalPacket,
		txTotalByte:             txTotalByte,
		txDroppedPacket:         txDroppedPacket,
	}
}

// Describe implements the prometheus.Collector interface.
func (c *logicalRouterPortCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.rxTotalPacket
	ch <- c.rxDroppedPacket
	ch <- c.rxTotalByte
	ch <- c.txTotalPacket
	ch <- c.txDroppedPacket
	ch <- c.txTotalByte
}

// Collect implements the prometheus.Collector interface.
func (c *logicalRouterPortCollector) Collect(ch chan<- prometheus.Metric) {
	logicalRouterPortStatisticMetrics := c.generateLogicalRouterPortStatisticMetrics()
	for _, metric := range logicalRouterPortStatisticMetrics {
		ch <- c.buildLogicalRouterPortMetric(metric.LogicalRouterPort, c.rxTotalPacket, float64(metric.Rx.TotalPackets))
		ch <- c.buildLogicalRouterPortMetric(metric.LogicalRouterPort, c.rxDroppedPacket, float64(metric.Rx.DroppedPackets))
		ch <- c.buildLogicalRouterPortMetric(metric.LogicalRouterPort, c.rxTotalByte, float64(metric.Rx.TotalBytes))
		ch <- c.buildLogicalRouterPortMetric(metric.LogicalRouterPort, c.txTotalPacket, float64(metric.Tx.TotalPackets))
		ch <- c.buildLogicalRouterPortMetric(metric.LogicalRouterPort, c.txDroppedPacket, float64(metric.Tx.DroppedPackets))
		ch <- c.buildLogicalRouterPortMetric(metric.LogicalRouterPort, c.txTotalByte, float64(metric.Tx.TotalBytes))
	}
}

func (c *logicalRouterPortCollector) buildLogicalRouterPortMetric(logicalRouterPort manager.LogicalRouterPort, desc *prometheus.Desc, value float64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		value,
		logicalRouterPort.Id,
		logicalRouterPort.DisplayName,
		logicalRouterPort.LogicalRouterId,
	)
}

func (c *logicalRouterPortCollector) generateLogicalRouterPortStatisticMetrics() (logicalRouterPortStatisticMetrics []logicalRouterPortStatisticMetric) {
	logicalRouterPorts, err := c.logicalRouterPortClient.ListAllLogicalRouterPorts()
	if err != nil {
		level.Error(c.logger).Log("msg", "Unable to list logical ports", "err", err)
		return
	}

	for _, logicalRouterPort := range logicalRouterPorts {
		statistic, err := c.logicalRouterPortClient.GetLogicalRouterPortStatisticsSummary(logicalRouterPort.Id)
		if err != nil {
			level.Error(c.logger).Log("msg", "Unable to get logical router port statistics", "id", logicalRouterPort.Id, "err", err)
			continue
		}
		logicalRouterPortStatisticMetric := logicalRouterPortStatisticMetric{
			LogicalRouterPort: logicalRouterPort,
			Rx:                statistic.Rx,
			Tx:                statistic.Tx,
		}
		logicalRouterPortStatisticMetrics = append(logicalRouterPortStatisticMetrics, logicalRouterPortStatisticMetric)
	}
	return
}
