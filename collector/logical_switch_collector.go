package collector

import (
	"strings"

	"nsxt_exporter/client"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	nsxt "github.com/vmware/go-vmware-nsxt"
)

func init() {
	registerCollector("logical_switch", createLogicalSwitchFactory)
}

type logicalSwitchCollector struct {
	logicalSwitchClient client.LogicalSwitchClient
	logger              log.Logger

	logicalSwitchStatus *prometheus.Desc
}

type logicalSwitchStatusMetric struct {
	ID              string
	Name            string
	TransportZoneID string
	Status          float64
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
	return &logicalSwitchCollector{
		logicalSwitchClient: lswitchClient,
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
	lswitchStatusMetrics := c.generateLogicalSwitchStatusMetrics()
	for _, m := range lswitchStatusMetrics {
		labels := []string{m.ID, m.Name, m.TransportZoneID}
		ch <- prometheus.MustNewConstMetric(c.logicalSwitchStatus, prometheus.GaugeValue, m.Status, labels...)
	}
}

func (c *logicalSwitchCollector) generateLogicalSwitchStatusMetrics() (logicalSwitchStatusMetrics []logicalSwitchStatusMetric) {
	logicalSwitches, err := c.logicalSwitchClient.ListAllLogicalSwitches()
	if err != nil {
		level.Error(c.logger).Log("msg", "Unable to list logical switches", "err", err)
		return
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
