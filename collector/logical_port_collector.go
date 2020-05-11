package collector

import (
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
	client *nsxt.APIClient
	logger log.Logger

	logicalPortTotal  *prometheus.Desc
	logicalPortUp     *prometheus.Desc
	logicalPortStatus *prometheus.Desc
}

func newLogicalPortCollector(client *nsxt.APIClient, logger log.Logger) prometheus.Collector {
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
		client:            client,
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
	lportStatus, _, err := lpc.client.LogicalSwitchingApi.GetLogicalPortStatusSummary(lpc.client.Context, nil)
	if err != nil {
		level.Error(lpc.logger).Log("msg", "Unable to collect logical port status summary", "err", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(lpc.logicalPortTotal, prometheus.GaugeValue, float64(lportStatus.TotalPorts))
	ch <- prometheus.MustNewConstMetric(lpc.logicalPortUp, prometheus.GaugeValue, float64(lportStatus.UpPorts))

	lportMetrics := lpc.generateLogicalPortMetrics()
	for _, lportMetric := range lportMetrics {
		ch <- prometheus.MustNewConstMetric(lpc.logicalPortStatus, prometheus.GaugeValue, lportMetric.Status, lportMetric.ID, lportMetric.Name)
	}
}

type LportMetric struct {
	Name   string
	ID     string
	Status float64
}

func (lpc *logicalPortCollector) generateLogicalPortMetrics() (lportMetrics []LportMetric) {
	var lports []manager.LogicalPort
	var cursor string
	for {
		localVarOptionals := make(map[string]interface{})
		localVarOptionals["cursor"] = cursor
		lportsResult, _, err := lpc.client.LogicalSwitchingApi.ListLogicalPorts(lpc.client.Context, localVarOptionals)
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
		lportStatus, _, err := lpc.client.LogicalSwitchingApi.GetLogicalPortOperationalStatus(lpc.client.Context, lport.Id, nil)
		if err != nil {
			level.Error(lpc.logger).Log("msg", "Unable to get logical port status", "id", lport.Id, "err", err)
			continue
		}
		status := 0
		if lportStatus.Status == "UP" {
			status = 1
		}
		lportMetric := LportMetric{
			Name:   lport.DisplayName,
			ID:     lport.Id,
			Status: float64(status),
		}
		lportMetrics = append(lportMetrics, lportMetric)
	}
	return
}
