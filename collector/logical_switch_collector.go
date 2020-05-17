package collector

import (
	"strings"

	"nsxt_exporter/client"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	nsxt "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/manager"
)

func init() {
	registerCollector("logical_switch", createLogicalSwitchFactory)
}

type logicalSwitchCollector struct {
	logicalSwitchClient client.LogicalSwitchClient
	logger              log.Logger

	logicalSwitchStatus *prometheus.Desc
	rxByteTotal         *prometheus.Desc
	rxByteDropped       *prometheus.Desc
	rxPacketTotal       *prometheus.Desc
	rxPacketDropped     *prometheus.Desc
	txByteTotal         *prometheus.Desc
	txByteDropped       *prometheus.Desc
	txPacketTotal       *prometheus.Desc
	txPacketDropped     *prometheus.Desc
}

type logicalSwitchStatusMetric struct {
	ID              string
	Name            string
	TransportZoneID string
	Status          float64
}

type logicalSwitchStatisticMetric struct {
	ID              string
	Name            string
	TransportZoneID string
	RxByteTotal     float64
	RxByteDropped   float64
	RxPacketTotal   float64
	RxPacketDropped float64
	TxByteTotal     float64
	TxByteDropped   float64
	TxPacketTotal   float64
	TxPacketDropped float64
}

func createLogicalSwitchFactory(apiClient *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	nsxtClient := client.NewNSXTClient(apiClient, logger)
	return newLogicalSwitchCollector(nsxtClient, logger)
}

func newLogicalSwitchCollector(lswitchClient client.LogicalSwitchClient, logger log.Logger) *logicalSwitchCollector {
	logicalSwitchStatus := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_switch", "status"),
		"Status of logical switch success/other",
		[]string{"id", "name", "transport_zone_id"},
		nil,
	)
	rxByteTotal := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_switch", "rx_byte"),
		"Total bytes received (rx) on logical switch",
		[]string{"id", "name", "transport_zone_id"},
		nil,
	)
	rxByteDropped := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_switch", "rx_dropped_byte"),
		"Total receive (rx) bytes dropped on logical switch",
		[]string{"id", "name", "transport_zone_id"},
		nil,
	)
	rxPacketTotal := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_switch", "rx_packet"),
		"Total packets received (rx) on logical switch",
		[]string{"id", "name", "transport_zone_id"},
		nil,
	)
	rxPacketDropped := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_switch", "rx_dropped_packet"),
		"Total receive (rx) packets dropped on logical switch",
		[]string{"id", "name", "transport_zone_id"},
		nil,
	)
	txByteTotal := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_switch", "tx_byte"),
		"Total bytes transmitted (tx) on logical switch",
		[]string{"id", "name", "transport_zone_id"},
		nil,
	)
	txByteDropped := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_switch", "tx_dropped_byte"),
		"Total transmit (tx) bytes dropped on logical switch",
		[]string{"id", "name", "transport_zone_id"},
		nil,
	)
	txPacketTotal := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_switch", "tx_packet"),
		"Total packets transmitted (tx) on logical switch",
		[]string{"id", "name", "transport_zone_id"},
		nil,
	)
	txPacketDropped := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_switch", "tx_dropped_packet"),
		"Total transmit (tx) packets dropped on logical switch",
		[]string{"id", "name", "transport_zone_id"},
		nil,
	)
	return &logicalSwitchCollector{
		logicalSwitchClient: lswitchClient,
		logger:              logger,
		logicalSwitchStatus: logicalSwitchStatus,
		rxPacketTotal:       rxPacketTotal,
		rxPacketDropped:     rxPacketDropped,
		rxByteTotal:         rxByteTotal,
		rxByteDropped:       rxByteDropped,
		txPacketTotal:       txPacketTotal,
		txPacketDropped:     txPacketDropped,
		txByteTotal:         txByteTotal,
		txByteDropped:       txByteDropped,
	}
}

// Describe implements the prometheus.Collector interface.
func (c *logicalSwitchCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.logicalSwitchStatus
	ch <- c.rxByteTotal
	ch <- c.rxByteDropped
	ch <- c.rxPacketTotal
	ch <- c.rxPacketDropped
	ch <- c.txByteTotal
	ch <- c.txByteDropped
	ch <- c.txPacketTotal
	ch <- c.txPacketDropped
}

// Collect implements the prometheus.Collector interface.
func (c *logicalSwitchCollector) Collect(ch chan<- prometheus.Metric) {
	logicalSwitches, err := c.logicalSwitchClient.ListAllLogicalSwitches()
	if err != nil {
		level.Error(c.logger).Log("msg", "Unable to list logical switches", "err", err)
		return
	}
	lswitchStatusMetrics := c.generateLogicalSwitchStatusMetrics(logicalSwitches)
	for _, m := range lswitchStatusMetrics {
		labels := []string{m.ID, m.Name, m.TransportZoneID}
		ch <- prometheus.MustNewConstMetric(c.logicalSwitchStatus, prometheus.GaugeValue, m.Status, labels...)
	}
	lswitchStatisticMetrics := c.generateLogicalSwitchStatisticMetrics(logicalSwitches)
	for _, metric := range lswitchStatisticMetrics {
		labels := []string{metric.ID, metric.Name, metric.TransportZoneID}
		ch <- prometheus.MustNewConstMetric(c.rxByteTotal, prometheus.GaugeValue, metric.RxByteTotal, labels...)
		ch <- prometheus.MustNewConstMetric(c.rxByteDropped, prometheus.GaugeValue, metric.RxByteDropped, labels...)
		ch <- prometheus.MustNewConstMetric(c.rxPacketTotal, prometheus.GaugeValue, metric.RxPacketTotal, labels...)
		ch <- prometheus.MustNewConstMetric(c.rxPacketDropped, prometheus.GaugeValue, metric.RxPacketDropped, labels...)
		ch <- prometheus.MustNewConstMetric(c.txByteTotal, prometheus.GaugeValue, metric.TxByteTotal, labels...)
		ch <- prometheus.MustNewConstMetric(c.txByteDropped, prometheus.GaugeValue, metric.TxByteDropped, labels...)
		ch <- prometheus.MustNewConstMetric(c.txPacketTotal, prometheus.GaugeValue, metric.TxPacketTotal, labels...)
		ch <- prometheus.MustNewConstMetric(c.txPacketDropped, prometheus.GaugeValue, metric.TxPacketDropped, labels...)
	}
}

func (c *logicalSwitchCollector) generateLogicalSwitchStatusMetrics(logicalSwitches []manager.LogicalSwitch) (logicalSwitchStatusMetrics []logicalSwitchStatusMetric) {
	for _, logicalSwitch := range logicalSwitches {
		logicalSwitchStatus, err := c.logicalSwitchClient.GetLogicalSwitchState(logicalSwitch.Id)
		if err != nil {
			level.Error(c.logger).Log("msg", "Unable to get logical switch status", "id", logicalSwitch.Id, "err", err)
			continue
		}
		var status float64
		if strings.ToUpper(logicalSwitchStatus.State) == "SUCCESS" {
			status = 1
		} else {
			status = 0
		}
		logicalSwitchStatusMetric := logicalSwitchStatusMetric{
			ID:              logicalSwitch.Id,
			Name:            logicalSwitch.DisplayName,
			TransportZoneID: logicalSwitch.TransportZoneId,
			Status:          status,
		}
		logicalSwitchStatusMetrics = append(logicalSwitchStatusMetrics, logicalSwitchStatusMetric)
	}
	return
}

func (c *logicalSwitchCollector) generateLogicalSwitchStatisticMetrics(logicalSwitches []manager.LogicalSwitch) (logicalSwitchStatisticMetrics []logicalSwitchStatisticMetric) {
	for _, logicalSwitch := range logicalSwitches {
		logicalSwitchStatistic, err := c.logicalSwitchClient.GetLogicalSwitchStatistic(logicalSwitch.Id)
		if err != nil {
			level.Error(c.logger).Log("msg", "Unable to get logical switch statistic", "id", logicalSwitch.Id, "err", err)
			continue
		}
		logicalSwitchStatisticMetric := logicalSwitchStatisticMetric{
			ID:              logicalSwitch.Id,
			Name:            logicalSwitch.DisplayName,
			TransportZoneID: logicalSwitch.TransportZoneId,
			RxByteTotal:     float64(logicalSwitchStatistic.RxBytes.Total),
			RxByteDropped:   float64(logicalSwitchStatistic.RxBytes.Dropped),
			RxPacketTotal:   float64(logicalSwitchStatistic.RxPackets.Total),
			RxPacketDropped: float64(logicalSwitchStatistic.RxPackets.Dropped),
			TxByteTotal:     float64(logicalSwitchStatistic.TxBytes.Total),
			TxByteDropped:   float64(logicalSwitchStatistic.TxBytes.Dropped),
			TxPacketTotal:   float64(logicalSwitchStatistic.TxPackets.Total),
			TxPacketDropped: float64(logicalSwitchStatistic.TxPackets.Dropped),
		}
		logicalSwitchStatisticMetrics = append(logicalSwitchStatisticMetrics, logicalSwitchStatisticMetric)
	}
	return
}
