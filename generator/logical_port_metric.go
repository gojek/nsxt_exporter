package generator

import (
	"strings"

	"nsxt_exporter/client"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/vmware/go-vmware-nsxt/manager"
)

func NewLogicalPortMetricGenerator(client client.NSXTClient, logger log.Logger) *logicalPortMetricGenerator {
	return &logicalPortMetricGenerator{
		nsxtClient: client,
		logger:     logger,
	}
}

type LogicalPortMetricGenerator interface {
	GenerateLogicalPortStatusSummary() (manager.LogicalPortStatusSummary, bool)
	GenerateLogicalPortStatusMetrics() ([]LogicalPortStatus, bool)
}

type logicalPortMetricGenerator struct {
	nsxtClient client.NSXTClient
	logger     log.Logger
}

type LogicalPortStatus struct {
	Status int64
	Name   string
	ID     string
}

func (g *logicalPortMetricGenerator) GenerateLogicalPortStatusSummary() (manager.LogicalPortStatusSummary, bool) {
	lportStatus, err := g.nsxtClient.GetLogicalPortStatusSummary(nil)
	if err != nil {
		level.Error(g.logger).Log("msg", "Unable to collect logical port status summary", "err", err)
		return manager.LogicalPortStatusSummary{}, false
	}
	return lportStatus, true
}

func (g *logicalPortMetricGenerator) GenerateLogicalPortStatusMetrics() ([]LogicalPortStatus, bool) {
	var lportsStatus []LogicalPortStatus
	var lports []manager.LogicalPort
	var cursor string
	for {
		localVarOptionals := make(map[string]interface{})
		localVarOptionals["cursor"] = cursor
		lportsResult, err := g.nsxtClient.ListLogicalPorts(localVarOptionals)
		if err != nil {
			level.Error(g.logger).Log("msg", "Unable to list logical ports", "err", err)
			return lportsStatus, false
		}
		lports = append(lports, lportsResult.Results...)
		cursor = lportsResult.Cursor
		if len(cursor) == 0 {
			break
		}
	}
	ok := true
	for _, lport := range lports {
		lportStatus, err := g.nsxtClient.GetLogicalPortOperationalStatus(lport.Id, nil)
		if err != nil {
			level.Error(g.logger).Log("msg", "Unable to get logical port status", "id", lport.Id, "err", err)
			ok = false
			continue
		}
		var status int64
		if strings.ToUpper(lportStatus.Status) == "UP" {
			status = 1
		} else {
			status = 0
		}
		lportsStatus = append(lportsStatus, LogicalPortStatus{
			ID:     lport.Id,
			Name:   lport.DisplayName,
			Status: status,
		})
	}
	return lportsStatus, ok
}
