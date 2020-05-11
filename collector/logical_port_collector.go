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
	registerCollector("logical_port", newLogicalPortCollector)
}

type logicalPortCollector struct {
	logicalPortClient client.LogicalPortClient
	logger            log.Logger

	logicalPortTotal  *prometheus.Desc
	logicalPortUp     *prometheus.Desc
	logicalPortStatus *prometheus.Desc
}

func newLogicalPortCollector(apiClient *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	nsxtClient := client.NewNSXTClient(apiClient, logger)
	logicalPortTotal := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_port", "total"),
		"Total number of logical ports.",
		nil,
		nil,
	)
	logicalPortUp := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_port", "up"),
		"Number of logical ports currently up.",
		nil,
		nil,
	)
	logicalPortStatus := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_port", "status"),
		"Status of logical port UP/DOWN",
		[]string{"id", "name"},
		nil,
	)
	return &logicalPortCollector{
		logicalPortClient: nsxtClient,
		logger:            logger,

		logicalPortTotal:  logicalPortTotal,
		logicalPortUp:     logicalPortUp,
		logicalPortStatus: logicalPortStatus,
	}
}

// Describe implements the prometheus.Collector interface.
func (lpc *logicalPortCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- lpc.logicalPortUp
	ch <- lpc.logicalPortTotal
	ch <- lpc.logicalPortStatus
}

// Collect implements the prometheus.Collector interface.
func (lpc *logicalPortCollector) Collect(ch chan<- prometheus.Metric) {
	lportStatus, err := lpc.logicalPortClient.GetLogicalPortStatusSummary(nil)
	if err != nil {
		level.Error(lpc.logger).Log("msg", "Unable to collect logical port status summary", "err", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(lpc.logicalPortTotal, prometheus.GaugeValue, float64(lportStatus.TotalPorts))
	ch <- prometheus.MustNewConstMetric(lpc.logicalPortUp, prometheus.GaugeValue, float64(lportStatus.UpPorts))

	lportStatusMetrics := lpc.generateLogicalPortStatusMetrics()
	for _, lportStatusMetric := range lportStatusMetrics {
		ch <- lportStatusMetric
	}
}

func (lpc *logicalPortCollector) generateLogicalPortStatusMetrics() (lportStatusMetrics []prometheus.Metric) {
	var lports []manager.LogicalPort
	var cursor string
	for {
		localVarOptionals := make(map[string]interface{})
		localVarOptionals["cursor"] = cursor
		lportsResult, err := lpc.logicalPortClient.ListLogicalPorts(localVarOptionals)
		if err != nil {
			level.Error(lpc.logger).Log("msg", "Unable to list logical ports", "err", err)
			return
		}
		lports = append(lports, lportsResult.Results...)
		cursor = lportsResult.Cursor
		if len(cursor) == 0 {
			break
		}
	}
	for _, lport := range lports {
		lportStatus, err := lpc.logicalPortClient.GetLogicalPortOperationalStatus(lport.Id, nil)
		if err != nil {
			level.Error(lpc.logger).Log("msg", "Unable to get logical port status", "id", lport.Id, "err", err)
			continue
		}
		var status float64
		if strings.ToUpper(lportStatus.Status) == "UP" {
			status = 1
		} else {
			status = 0
		}
		lportStatusMetric := prometheus.MustNewConstMetric(lpc.logicalPortStatus, prometheus.GaugeValue, status, lport.Id, lport.DisplayName)
		lportStatusMetrics = append(lportStatusMetrics, lportStatusMetric)
	}
	return
}
