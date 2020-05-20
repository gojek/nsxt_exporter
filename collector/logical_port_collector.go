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

var logicalPortPossibleStatus = [...]string{"UP", "DOWN", "UNKNOWN"}

func init() {
	registerCollector("logical_port", createLogicalPortCollectorFactory)
}

type logicalPortCollector struct {
	logicalPortClient client.LogicalPortClient
	logger            log.Logger

	logicalPortStatus *prometheus.Desc
}

type logicalPortStatusMetric struct {
	ID              string
	Name            string
	StatusDetail    map[string]float64
	LogicalSwitchID string
}

func createLogicalPortCollectorFactory(apiClient *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	nsxtClient := client.NewNSXTClient(apiClient, logger)
	return newLogicalPortCollector(nsxtClient, logger)
}

func newLogicalPortCollector(logicalPortClient client.LogicalPortClient, logger log.Logger) *logicalPortCollector {
	logicalPortStatus := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_port", "status"),
		"Status of logical port",
		[]string{"id", "name", "logical_switch_id", "status"},
		nil,
	)
	return &logicalPortCollector{
		logicalPortClient: logicalPortClient,
		logger:            logger,

		logicalPortStatus: logicalPortStatus,
	}
}

// Describe implements the prometheus.Collector interface.
func (lpc *logicalPortCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- lpc.logicalPortStatus
}

// Collect implements the prometheus.Collector interface.
func (lpc *logicalPortCollector) Collect(ch chan<- prometheus.Metric) {
	lportStatusMetrics := lpc.generateLogicalPortStatusMetrics()
	for _, lportStatusMetric := range lportStatusMetrics {
		for status, value := range lportStatusMetric.StatusDetail {
			ch <- prometheus.MustNewConstMetric(
				lpc.logicalPortStatus,
				prometheus.GaugeValue,
				value,
				lportStatusMetric.ID,
				lportStatusMetric.Name,
				lportStatusMetric.LogicalSwitchID,
				status,
			)
		}
	}
}

func (lpc *logicalPortCollector) generateLogicalPortStatusMetrics() (lportStatusMetrics []logicalPortStatusMetric) {
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
		lportStatusMetric := logicalPortStatusMetric{
			ID:              lport.Id,
			Name:            lport.DisplayName,
			StatusDetail:    map[string]float64{},
			LogicalSwitchID: lport.LogicalSwitchId,
		}
		for _, possibleStatus := range logicalPortPossibleStatus {
			statusValue := 0.0
			if possibleStatus == strings.ToUpper(lportStatus.Status) {
				statusValue = 1.0
			}
			lportStatusMetric.StatusDetail[possibleStatus] = statusValue
		}
		lportStatusMetrics = append(lportStatusMetrics, lportStatusMetric)
	}
	return
}
