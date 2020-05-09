package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	nsxt "github.com/vmware/go-vmware-nsxt"

	"nsxt_exporter/generator"
)

func init() {
	registerCollector("logical_port", newLogicalPortCollector)
}

type logicalPortCollector struct {
	metricGenerator generator.LogicalPortMetricGenerator

	logicalPortTotal            *prometheus.Desc
	logicalPortUp               *prometheus.Desc
	logicalPortStatus           *prometheus.Desc
	logicalPortCollectorSuccess *prometheus.Desc
}

func newLogicalPortCollector(client *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	metricGenerator := generator.NewLogicalPortMetricGenerator(client, logger)
	logicalPortTotal := prometheus.NewDesc(prometheus.BuildFQName(namespace, "logical_port", "total"), "Total number of logical ports.", nil, nil)
	logicalPortUp := prometheus.NewDesc(prometheus.BuildFQName(namespace, "logical_port", "up"), "Number of logical ports currently up.", nil, nil)
	logicalPortStatus := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_port", "status"),
		"Status of logical port UP/DOWN",
		[]string{"id", "name"},
		nil,
	)
	logicalPortCollectorSuccess := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "collector", "success"),
		"Whether a collector success",
		[]string{"collector"},
		nil,
	)
	return &logicalPortCollector{
		logicalPortTotal:            logicalPortTotal,
		logicalPortUp:               logicalPortUp,
		logicalPortStatus:           logicalPortStatus,
		metricGenerator:             metricGenerator,
		logicalPortCollectorSuccess: logicalPortCollectorSuccess,
	}
}

// Describe implements the prometheus.Collector interface.
func (lpc *logicalPortCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- lpc.logicalPortUp
	ch <- lpc.logicalPortTotal
}

// Collect implements the prometheus.Collector interface.
func (lpc *logicalPortCollector) Collect(ch chan<- prometheus.Metric) {
	lportStatus, ok := lpc.metricGenerator.GenerateLogicalPortStatusSummary()
	if !ok {
		ch <- prometheus.MustNewConstMetric(lpc.logicalPortCollectorSuccess, prometheus.GaugeValue, 0, "logical_port")
		return
	}
	ch <- prometheus.MustNewConstMetric(lpc.logicalPortTotal, prometheus.GaugeValue, float64(lportStatus.TotalPorts))
	ch <- prometheus.MustNewConstMetric(lpc.logicalPortUp, prometheus.GaugeValue, float64(lportStatus.UpPorts))

	lportsStatus, ok := lpc.metricGenerator.GenerateLogicalPortStatusMetrics()
	for _, lportStatus := range lportsStatus {
		ch <- prometheus.MustNewConstMetric(
			lpc.logicalPortStatus,
			prometheus.GaugeValue,
			float64(lportStatus.Status),
			lportStatus.ID,
			lportStatus.Name,
		)
	}

	if ok {
		ch <- prometheus.MustNewConstMetric(lpc.logicalPortCollectorSuccess, prometheus.GaugeValue, 1, "logical_port")
	} else {
		ch <- prometheus.MustNewConstMetric(lpc.logicalPortCollectorSuccess, prometheus.GaugeValue, 0, "logical_port")
	}
}
