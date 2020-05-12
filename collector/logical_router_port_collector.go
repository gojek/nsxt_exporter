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
		ch <- metric
	}
}

func (c *logicalRouterPortCollector) generateLogicalRouterPortStatisticMetrics() (logicalRouterPortStatisticMetrics []prometheus.Metric) {
	var logicalRouterPorts []manager.LogicalRouterPort
	var cursor string
	for {
		localVarOptionals := make(map[string]interface{})
		localVarOptionals["cursor"] = cursor
		logicalRouterPortsResult, err := c.logicalRouterPortClient.ListLogicalRouterPorts(localVarOptionals)
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
		statistic, err := c.logicalRouterPortClient.GetLogicalRouterPortStatisticsSummary(logicalRouterPort.Id)
		if err != nil {
			level.Error(c.logger).Log("msg", "Unable to get logical router port statistics", "id", logicalRouterPort.Id, "err", err)
			continue
		}
		rxTotalPacketMetric := prometheus.MustNewConstMetric(
			c.rxTotalPacket,
			prometheus.GaugeValue,
			float64(statistic.Rx.TotalPackets),
			logicalRouterPort.Id,
			logicalRouterPort.DisplayName,
			logicalRouterPort.LogicalRouterId,
		)
		rxDroppedPacketMetric := prometheus.MustNewConstMetric(
			c.rxDroppedPacket,
			prometheus.GaugeValue,
			float64(statistic.Rx.DroppedPackets),
			logicalRouterPort.Id,
			logicalRouterPort.DisplayName,
			logicalRouterPort.LogicalRouterId,
		)
		rxTotalByteMetric := prometheus.MustNewConstMetric(
			c.rxTotalByte,
			prometheus.GaugeValue,
			float64(statistic.Rx.TotalBytes),
			logicalRouterPort.Id,
			logicalRouterPort.DisplayName,
			logicalRouterPort.LogicalRouterId,
		)
		txTotalPacketMetric := prometheus.MustNewConstMetric(
			c.txTotalPacket,
			prometheus.GaugeValue,
			float64(statistic.Tx.TotalPackets),
			logicalRouterPort.Id,
			logicalRouterPort.DisplayName,
			logicalRouterPort.LogicalRouterId,
		)
		txDroppedPacketMetric := prometheus.MustNewConstMetric(
			c.txDroppedPacket,
			prometheus.GaugeValue,
			float64(statistic.Tx.DroppedPackets),
			logicalRouterPort.Id,
			logicalRouterPort.DisplayName,
			logicalRouterPort.LogicalRouterId,
		)
		txTotalByteMetric := prometheus.MustNewConstMetric(
			c.txTotalByte,
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
