package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	nsxt "github.com/vmware/go-vmware-nsxt"
)

func init() {
	registerCollector("logical_port", newLogicalPortCollector)
}

type logicalPortCollector struct {
	client *nsxt.APIClient
	logger log.Logger

	logicalPortTotal *prometheus.Desc
	logicalPortUp    *prometheus.Desc
}

func newLogicalPortCollector(client *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	logicalPortTotal := prometheus.NewDesc(prometheus.BuildFQName(namespace, "logical_port", "total"), "Total number of logical ports.", nil, nil)
	logicalPortUp := prometheus.NewDesc(prometheus.BuildFQName(namespace, "logical_port", "up"), "Number of logical ports currently up.", nil, nil)
	return &logicalPortCollector{
		client:           client,
		logger:           logger,
		logicalPortTotal: logicalPortTotal,
		logicalPortUp:    logicalPortUp,
	}
}

// Describe implements the prometheus.Collector interface.
func (lpc *logicalPortCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- lpc.logicalPortUp
	ch <- lpc.logicalPortTotal
}

// Collect implements the prometheus.Collector interface.
func (lpc *logicalPortCollector) Collect(ch chan<- prometheus.Metric) {
	lportStatus, _, err := lpc.client.LogicalSwitchingApi.GetLogicalPortStatusSummary(lpc.client.Context, nil)
	if err != nil {
		level.Error(lpc.logger).Log("msg", "Unable to collect logical port status summary", "err", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(lpc.logicalPortTotal, prometheus.GaugeValue, float64(lportStatus.TotalPorts))
	ch <- prometheus.MustNewConstMetric(lpc.logicalPortUp, prometheus.GaugeValue, float64(lportStatus.UpPorts))
}
