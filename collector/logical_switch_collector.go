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
	registerCollector("logical_switch", newLogicalSwitchCollector)
}

type logicalSwitchCollector struct {
	logicalSwitchClient client.LogicalSwitchClient
	logger              log.Logger

	logicalSwitchStatus *prometheus.Desc
}

func newLogicalSwitchCollector(apiClient *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	nsxtClient := client.NewNSXTClient(apiClient, logger)
	logicalSwitchStatus := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "logical_switch", "status"),
		"Status of logical switch success/other",
		[]string{"id", "name", "transport_zone_id"},
		nil,
	)
	return &logicalSwitchCollector{
		logicalSwitchClient: nsxtClient,
		logger:              logger,
		logicalSwitchStatus: logicalSwitchStatus,
	}
}

// Describe implements the prometheus.Collector interface.
func (c *logicalSwitchCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.logicalSwitchStatus
}

// Collect implements the prometheus.Collector interface.
func (c *logicalSwitchCollector) Collect(ch chan<- prometheus.Metric) {
	metrics := c.generateLogicalSwitchStatusMetrics()
	for _, metric := range metrics {
		ch <- metric
	}
}

func (c *logicalSwitchCollector) generateLogicalSwitchStatusMetrics() (logicalSwitchMetrics []prometheus.Metric) {
	var logicalSwitches []manager.LogicalSwitch
	var cursor string
	for {
		localVarOptionals := make(map[string]interface{})
		localVarOptionals["cursor"] = cursor
		logicalSwitchsResult, err := c.logicalSwitchClient.ListLogicalSwitches(localVarOptionals)
		if err != nil {
			level.Error(c.logger).Log("msg", "Unable to list logical switches", "err", err)
			return
		}
		logicalSwitches = append(logicalSwitches, logicalSwitchsResult.Results...)
		cursor = logicalSwitchsResult.Cursor
		if len(cursor) == 0 {
			break
		}
	}
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
		labels := []string{logicalSwitch.Id, logicalSwitch.DisplayName, logicalSwitch.TransportZoneId}
		logicalSwitchStatusMetric := prometheus.MustNewConstMetric(c.logicalSwitchStatus, prometheus.GaugeValue, status, labels...)
		logicalSwitchMetrics = append(logicalSwitchMetrics, logicalSwitchStatusMetric)
	}
	return
}
