package collector

import (
	"strings"

	"nsxt_exporter/client"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	nsxt "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/loadbalancer"
)

var loadBalancerPossibleStatus = []string{"UP", "DOWN", "ERROR", "NO_STANDBY", "DETACHED", "DISABLED", "UNKNOWN"}
var loadBalancerPoolPossibleStatus = []string{"UP", "PARTIALLY_UP", "PRIMARY_DOWN", "DOWN", "DETACHED", "UNKNOWN"}
var loadBalancerPoolMemberPossibleStatus = []string{"UP", "DOWN", "DISABLED", "GRACEFUL_DISABLED", "UNUSED"}

func init() {
	registerCollector("load_balancer", createLoadBalancerCollectorFactory)
}

type loadBalancerCollector struct {
	client client.LoadBalancerClient
	logger log.Logger

	loadBalancerStatus            *prometheus.Desc
	loadBalancerPoolStatus        *prometheus.Desc
	loadBalancerPoolMemberStatus  *prometheus.Desc
	loadBalancerL4CurrentSessions *prometheus.Desc
	loadBalancerL4MaxSessions     *prometheus.Desc
	loadBalancerL4TotalSessions   *prometheus.Desc
	loadBalancerL7CurrentSessions *prometheus.Desc
	loadBalancerL7MaxSessions     *prometheus.Desc
	loadBalancerL7TotalSessions   *prometheus.Desc

	loadBalancerPoolBytesIn                      *prometheus.Desc
	loadBalancerPoolBytesOut                     *prometheus.Desc
	loadBalancerPoolCurrentSessions              *prometheus.Desc
	loadBalancerPoolHttpRequests                 *prometheus.Desc
	loadBalancerPoolMaxSessions                  *prometheus.Desc
	loadBalancerPoolPacketsIn                    *prometheus.Desc
	loadBalancerPoolPacketsOut                   *prometheus.Desc
	loadBalancerPoolSourceIPPersistenceEntrySize *prometheus.Desc
	loadBalancerPoolTotalSessions                *prometheus.Desc

	loadBalancerPoolMemberBytesIn                      *prometheus.Desc
	loadBalancerPoolMemberBytesOut                     *prometheus.Desc
	loadBalancerPoolMemberCurrentSessions              *prometheus.Desc
	loadBalancerPoolMemberHttpRequests                 *prometheus.Desc
	loadBalancerPoolMemberMaxSessions                  *prometheus.Desc
	loadBalancerPoolMemberPacketsIn                    *prometheus.Desc
	loadBalancerPoolMemberPacketsOut                   *prometheus.Desc
	loadBalancerPoolMemberSourceIPPersistenceEntrySize *prometheus.Desc
	loadBalancerPoolMemberTotalSessions                *prometheus.Desc

	loadBalancerVirtualServerBytesIn                      *prometheus.Desc
	loadBalancerVirtualServerBytesOut                     *prometheus.Desc
	loadBalancerVirtualServerCurrentSessions              *prometheus.Desc
	loadBalancerVirtualServerHttpRequests                 *prometheus.Desc
	loadBalancerVirtualServerMaxSessions                  *prometheus.Desc
	loadBalancerVirtualServerPacketsIn                    *prometheus.Desc
	loadBalancerVirtualServerPacketsOut                   *prometheus.Desc
	loadBalancerVirtualServerSourceIPPersistenceEntrySize *prometheus.Desc
	loadBalancerVirtualServerTotalSessions                *prometheus.Desc
}

type loadBalancerStatusMetric struct {
	ID           string
	Name         string
	StatusDetail map[string]float64
	PoolsStatus  []loadBalancerPoolStatusMetric
}

type loadBalancerPoolStatusMetric struct {
	ID            string
	StatusDetail  map[string]float64
	MembersStatus []loadBalancerPoolMemberStatusMetric
}

type loadBalancerPoolMemberStatusMetric struct {
	IPAddress    string
	Port         string
	StatusDetail map[string]float64
}

type loadBalancerStatisticMetric struct {
	ID                                        string
	Name                                      string
	L4CurrentSessions                         float64
	L4MaxSessions                             float64
	L4TotalSessions                           float64
	L7CurrentSessions                         float64
	L7MaxSessions                             float64
	L7TotalSessions                           float64
	loadBalancerPoolStatisticMetrics          []loadBalancerPoolStatisticMetric
	loadBalancerVirtualServerStatisticMetrics []loadBalancerVirtualServerStatisticMetric
}

type loadBalancerVirtualServerStatisticMetric struct {
	ID                           string
	BytesIn                      float64
	BytesOut                     float64
	CurrentSessions              float64
	HttpRequests                 float64
	MaxSessions                  float64
	PacketsIn                    float64
	PacketsOut                   float64
	SourceIPPersistenceEntrySize float64
	TotalSessions                float64
}

type loadBalancerPoolStatisticMetric struct {
	ID                                     string
	BytesIn                                float64
	BytesOut                               float64
	CurrentSessions                        float64
	HttpRequests                           float64
	MaxSessions                            float64
	PacketsIn                              float64
	PacketsOut                             float64
	SourceIPPersistenceEntrySize           float64
	TotalSessions                          float64
	loadBalancerPoolMemberStatisticMetrics []loadBalancerPoolMemberStatisticMetric
}

type loadBalancerPoolMemberStatisticMetric struct {
	IPAddress                    string
	Port                         string
	BytesIn                      float64
	BytesOut                     float64
	CurrentSessions              float64
	HttpRequests                 float64
	MaxSessions                  float64
	PacketsIn                    float64
	PacketsOut                   float64
	SourceIPPersistenceEntrySize float64
	TotalSessions                float64
}

func createLoadBalancerCollectorFactory(apiClient *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	nsxtClient := client.NewNSXTClient(apiClient, logger)
	return newLoadBalancerCollector(nsxtClient, logger)
}

func newLoadBalancerCollector(client client.LoadBalancerClient, logger log.Logger) *loadBalancerCollector {
	loadBalancerStatus := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer", "status"),
		"Status of Load Balancer",
		[]string{"id", "name", "status"},
		nil,
	)
	loadBalancerPoolStatus := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer", "pool_status"),
		"Status of Load Balancer pool",
		[]string{"id", "load_balancer_id", "status"},
		nil,
	)
	loadBalancerPoolMemberStatus := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer", "pool_member_status"),
		"Status of Load Balancer pool member",
		[]string{"ip_address", "port", "load_balancer_pool_id", "load_balancer_id", "status"},
		nil,
	)
	loadBalancerL4CurrentSessions := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer", "l4_current_sessions"),
		"Number of Load Balancer L4 current sessions",
		[]string{"id", "name"},
		nil,
	)
	loadBalancerL4MaxSessions := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer", "l4_max_sessions"),
		"Number of Load Balancer L4 max sessions",
		[]string{"id", "name"},
		nil,
	)
	loadBalancerL4TotalSessions := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer", "l4_total_sessions"),
		"Number of Load Balancer L4 total sessions",
		[]string{"id", "name"},
		nil,
	)
	loadBalancerL7CurrentSessions := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer", "l7_current_sessions"),
		"Number of Load Balancer L7 current sessions",
		[]string{"id", "name"},
		nil,
	)
	loadBalancerL7MaxSessions := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer", "l7_max_sessions"),
		"Number of Load Balancer L7 max sessions",
		[]string{"id", "name"},
		nil,
	)
	loadBalancerL7TotalSessions := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer", "l7_total_sessions"),
		"Number of Load Balancer L7 total sessions",
		[]string{"id", "name"},
		nil,
	)
	loadBalancerPoolBytesIn := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_pool", "incoming_bytes"),
		"Number of incoming bytes to Load Balancer Pool",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	loadBalancerPoolBytesOut := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_pool", "outgoing_bytes"),
		"Number of outgoing bytes to Load Balancer Pool",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	loadBalancerPoolCurrentSessions := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_pool", "current_sessions"),
		"Number of current sessions in Load Balancer Pool",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	loadBalancerPoolHttpRequests := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_pool", "http_request"),
		"Number of http request in Load Balancer Pool",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	loadBalancerPoolMaxSessions := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_pool", "max_sessions"),
		"Number of max sessions in Load Balancer Pool",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	loadBalancerPoolPacketsIn := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_pool", "incoming_packets"),
		"Number of incoming packets to Load Balancer Pool",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	loadBalancerPoolPacketsOut := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_pool", "outgoing_packets"),
		"Number of outgoing packets to Load Balancer Pool",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	loadBalancerPoolSourceIPPersistenceEntrySize := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_pool", "source_ip_persistence_entry"),
		"Number of source IP persistence entries to Load Balancer Pool",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	loadBalancerPoolTotalSessions := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_pool", "total_sessions"),
		"Number of total sessions in Load Balancer Pool",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	loadBalancerPoolMemberBytesIn := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_pool_member", "incoming_bytes"),
		"Number of incoming bytes to Load Balancer Pool Member",
		[]string{"ip", "port", "pool_id", "load_balancer_id"},
		nil,
	)
	loadBalancerPoolMemberBytesOut := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_pool_member", "outgoing_bytes"),
		"Number of outgoing bytes to Load Balancer Pool Member",
		[]string{"ip", "port", "pool_id", "load_balancer_id"},
		nil,
	)
	loadBalancerPoolMemberCurrentSessions := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_pool_member", "current_sessions"),
		"Number of current sessions in Load Balancer Pool Member",
		[]string{"ip", "port", "pool_id", "load_balancer_id"},
		nil,
	)
	loadBalancerPoolMemberHttpRequests := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_pool_member", "http_request"),
		"Number of http request in Load Balancer Pool Member",
		[]string{"ip", "port", "pool_id", "load_balancer_id"},
		nil,
	)
	loadBalancerPoolMemberMaxSessions := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_pool_member", "max_sessions"),
		"Number of max sessions in Load Balancer Pool Member",
		[]string{"ip", "port", "pool_id", "load_balancer_id"},
		nil,
	)
	loadBalancerPoolMemberPacketsIn := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_pool_member", "incoming_packets"),
		"Number of incoming packets to Load Balancer Pool Member",
		[]string{"ip", "port", "pool_id", "load_balancer_id"},
		nil,
	)
	loadBalancerPoolMemberPacketsOut := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_pool_member", "outgoing_packets"),
		"Number of outgoing packets to Load Balancer Pool Member",
		[]string{"ip", "port", "pool_id", "load_balancer_id"},
		nil,
	)
	loadBalancerPoolMemberSourceIPPersistenceEntrySize := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_pool_member", "source_ip_persistence_entry"),
		"Number of source IP persistence entries to Load Balancer Pool Member",
		[]string{"ip", "port", "pool_id", "load_balancer_id"},
		nil,
	)
	loadBalancerPoolMemberTotalSessions := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_pool_member", "total_sessions"),
		"Number of total sessions in Load Balancer Pool Member",
		[]string{"ip", "port", "pool_id", "load_balancer_id"},
		nil,
	)
	loadBalancerVirtualServerBytesIn := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_virtual_server", "incoming_bytes"),
		"Number of incoming bytes to Load Balancer Virtual Server",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	loadBalancerVirtualServerBytesOut := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_virtual_server", "outgoing_bytes"),
		"Number of outgoing bytes to Load Balancer Virtual Server",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	loadBalancerVirtualServerCurrentSessions := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_virtual_server", "current_sessions"),
		"Number of current sessions in Load Balancer Virtual Server",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	loadBalancerVirtualServerHttpRequests := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_virtual_server", "http_request"),
		"Number of http request in Load Balancer Virtual Server",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	loadBalancerVirtualServerMaxSessions := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_virtual_server", "max_sessions"),
		"Number of max sessions in Load Balancer Virtual Server",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	loadBalancerVirtualServerPacketsIn := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_virtual_server", "incoming_packets"),
		"Number of incoming packets to Load Balancer Virtual Server",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	loadBalancerVirtualServerPacketsOut := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_virtual_server", "outgoing_packets"),
		"Number of outgoing packets to Load Balancer Virtual Server",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	loadBalancerVirtualServerSourceIPPersistenceEntrySize := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_virtual_server", "source_ip_persistence_entry"),
		"Number of source IP persistence entries to Load Balancer Virtual Server",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	loadBalancerVirtualServerTotalSessions := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer_virtual_server", "total_sessions"),
		"Number of total sessions in Load Balancer Virtual Server",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	return &loadBalancerCollector{
		client: client,
		logger: logger,

		loadBalancerStatus:            loadBalancerStatus,
		loadBalancerPoolStatus:        loadBalancerPoolStatus,
		loadBalancerPoolMemberStatus:  loadBalancerPoolMemberStatus,
		loadBalancerL4CurrentSessions: loadBalancerL4CurrentSessions,
		loadBalancerL4MaxSessions:     loadBalancerL4MaxSessions,
		loadBalancerL4TotalSessions:   loadBalancerL4TotalSessions,
		loadBalancerL7CurrentSessions: loadBalancerL7CurrentSessions,
		loadBalancerL7MaxSessions:     loadBalancerL7MaxSessions,
		loadBalancerL7TotalSessions:   loadBalancerL7TotalSessions,

		loadBalancerPoolBytesIn:                      loadBalancerPoolBytesIn,
		loadBalancerPoolBytesOut:                     loadBalancerPoolBytesOut,
		loadBalancerPoolCurrentSessions:              loadBalancerPoolCurrentSessions,
		loadBalancerPoolHttpRequests:                 loadBalancerPoolHttpRequests,
		loadBalancerPoolMaxSessions:                  loadBalancerPoolMaxSessions,
		loadBalancerPoolPacketsIn:                    loadBalancerPoolPacketsIn,
		loadBalancerPoolPacketsOut:                   loadBalancerPoolPacketsOut,
		loadBalancerPoolSourceIPPersistenceEntrySize: loadBalancerPoolSourceIPPersistenceEntrySize,
		loadBalancerPoolTotalSessions:                loadBalancerPoolTotalSessions,

		loadBalancerPoolMemberBytesIn:                      loadBalancerPoolMemberBytesIn,
		loadBalancerPoolMemberBytesOut:                     loadBalancerPoolMemberBytesOut,
		loadBalancerPoolMemberCurrentSessions:              loadBalancerPoolMemberCurrentSessions,
		loadBalancerPoolMemberHttpRequests:                 loadBalancerPoolMemberHttpRequests,
		loadBalancerPoolMemberMaxSessions:                  loadBalancerPoolMemberMaxSessions,
		loadBalancerPoolMemberPacketsIn:                    loadBalancerPoolMemberPacketsIn,
		loadBalancerPoolMemberPacketsOut:                   loadBalancerPoolMemberPacketsOut,
		loadBalancerPoolMemberSourceIPPersistenceEntrySize: loadBalancerPoolMemberSourceIPPersistenceEntrySize,
		loadBalancerPoolMemberTotalSessions:                loadBalancerPoolMemberTotalSessions,

		loadBalancerVirtualServerBytesIn:                      loadBalancerVirtualServerBytesIn,
		loadBalancerVirtualServerBytesOut:                     loadBalancerVirtualServerBytesOut,
		loadBalancerVirtualServerCurrentSessions:              loadBalancerVirtualServerCurrentSessions,
		loadBalancerVirtualServerHttpRequests:                 loadBalancerVirtualServerHttpRequests,
		loadBalancerVirtualServerMaxSessions:                  loadBalancerVirtualServerMaxSessions,
		loadBalancerVirtualServerPacketsIn:                    loadBalancerVirtualServerPacketsIn,
		loadBalancerVirtualServerPacketsOut:                   loadBalancerVirtualServerPacketsOut,
		loadBalancerVirtualServerSourceIPPersistenceEntrySize: loadBalancerVirtualServerSourceIPPersistenceEntrySize,
		loadBalancerVirtualServerTotalSessions:                loadBalancerVirtualServerTotalSessions,
	}
}

// Describe implements the prometheus.Collector interface.
func (c *loadBalancerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.loadBalancerStatus
	ch <- c.loadBalancerPoolStatus
	ch <- c.loadBalancerPoolMemberStatus
	ch <- c.loadBalancerL4CurrentSessions
	ch <- c.loadBalancerL4MaxSessions
	ch <- c.loadBalancerL4TotalSessions
	ch <- c.loadBalancerL7CurrentSessions
	ch <- c.loadBalancerL7MaxSessions
	ch <- c.loadBalancerL7TotalSessions

	ch <- c.loadBalancerPoolBytesIn
	ch <- c.loadBalancerPoolBytesOut
	ch <- c.loadBalancerPoolCurrentSessions
	ch <- c.loadBalancerPoolHttpRequests
	ch <- c.loadBalancerPoolMaxSessions
	ch <- c.loadBalancerPoolPacketsIn
	ch <- c.loadBalancerPoolPacketsOut
	ch <- c.loadBalancerPoolSourceIPPersistenceEntrySize
	ch <- c.loadBalancerPoolTotalSessions

	ch <- c.loadBalancerPoolMemberBytesIn
	ch <- c.loadBalancerPoolMemberBytesOut
	ch <- c.loadBalancerPoolMemberCurrentSessions
	ch <- c.loadBalancerPoolMemberHttpRequests
	ch <- c.loadBalancerPoolMemberMaxSessions
	ch <- c.loadBalancerPoolMemberPacketsIn
	ch <- c.loadBalancerPoolMemberPacketsOut
	ch <- c.loadBalancerPoolMemberSourceIPPersistenceEntrySize
	ch <- c.loadBalancerPoolMemberTotalSessions

	ch <- c.loadBalancerVirtualServerBytesIn
	ch <- c.loadBalancerVirtualServerBytesOut
	ch <- c.loadBalancerVirtualServerCurrentSessions
	ch <- c.loadBalancerVirtualServerHttpRequests
	ch <- c.loadBalancerVirtualServerMaxSessions
	ch <- c.loadBalancerVirtualServerPacketsIn
	ch <- c.loadBalancerVirtualServerPacketsOut
	ch <- c.loadBalancerVirtualServerSourceIPPersistenceEntrySize
	ch <- c.loadBalancerVirtualServerTotalSessions
	return
}

// Collect implements the prometheus.Collector interface.
func (c *loadBalancerCollector) Collect(ch chan<- prometheus.Metric) {
	loadBalancers, err := c.client.ListAllLoadBalancers()
	if err != nil {
		level.Error(c.logger).Log("msg", "Unable to list load balancers", "err", err)
		return
	}
	statusMetrics := c.generateLoadBalancerStatusMetrics(loadBalancers)
	for _, metric := range statusMetrics {
		for status, value := range metric.StatusDetail {
			ch <- prometheus.MustNewConstMetric(c.loadBalancerStatus, prometheus.GaugeValue, value, metric.ID, metric.Name, status)
		}
		for _, poolStatus := range metric.PoolsStatus {
			for status, value := range poolStatus.StatusDetail {
				ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolStatus, prometheus.GaugeValue, value, poolStatus.ID, metric.ID, status)
			}
			for _, memberStatus := range poolStatus.MembersStatus {
				for status, value := range memberStatus.StatusDetail {
					ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolMemberStatus, prometheus.GaugeValue, value, memberStatus.IPAddress, memberStatus.Port, poolStatus.ID, metric.ID, status)
				}
			}
		}
	}
	statisticMetrics := c.generateLoadBalancerStatisticMetrics(loadBalancers)
	for _, metric := range statisticMetrics {
		ch <- prometheus.MustNewConstMetric(c.loadBalancerL4CurrentSessions, prometheus.GaugeValue, metric.L4CurrentSessions, metric.ID, metric.Name)
		ch <- prometheus.MustNewConstMetric(c.loadBalancerL4MaxSessions, prometheus.GaugeValue, metric.L4MaxSessions, metric.ID, metric.Name)
		ch <- prometheus.MustNewConstMetric(c.loadBalancerL4TotalSessions, prometheus.GaugeValue, metric.L4TotalSessions, metric.ID, metric.Name)
		ch <- prometheus.MustNewConstMetric(c.loadBalancerL7CurrentSessions, prometheus.GaugeValue, metric.L7CurrentSessions, metric.ID, metric.Name)
		ch <- prometheus.MustNewConstMetric(c.loadBalancerL7MaxSessions, prometheus.GaugeValue, metric.L7MaxSessions, metric.ID, metric.Name)
		ch <- prometheus.MustNewConstMetric(c.loadBalancerL7TotalSessions, prometheus.GaugeValue, metric.L7TotalSessions, metric.ID, metric.Name)
		for _, poolStatistic := range metric.loadBalancerPoolStatisticMetrics {
			poolLabels := []string{poolStatistic.ID, metric.ID}
			ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolBytesIn, prometheus.GaugeValue, poolStatistic.BytesIn, poolLabels...)
			ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolBytesOut, prometheus.GaugeValue, poolStatistic.BytesOut, poolLabels...)
			ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolCurrentSessions, prometheus.GaugeValue, poolStatistic.CurrentSessions, poolLabels...)
			ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolHttpRequests, prometheus.GaugeValue, poolStatistic.HttpRequests, poolLabels...)
			ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolMaxSessions, prometheus.GaugeValue, poolStatistic.MaxSessions, poolLabels...)
			ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolPacketsIn, prometheus.GaugeValue, poolStatistic.PacketsIn, poolLabels...)
			ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolPacketsOut, prometheus.GaugeValue, poolStatistic.PacketsOut, poolLabels...)
			ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolSourceIPPersistenceEntrySize, prometheus.GaugeValue, poolStatistic.SourceIPPersistenceEntrySize, poolLabels...)
			ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolTotalSessions, prometheus.GaugeValue, poolStatistic.TotalSessions, poolLabels...)
			for _, memberStatistic := range poolStatistic.loadBalancerPoolMemberStatisticMetrics {
				memberLabels := []string{memberStatistic.IPAddress, memberStatistic.Port, poolStatistic.ID, metric.ID}
				ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolMemberBytesIn, prometheus.GaugeValue, memberStatistic.BytesIn, memberLabels...)
				ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolMemberBytesOut, prometheus.GaugeValue, memberStatistic.BytesOut, memberLabels...)
				ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolMemberCurrentSessions, prometheus.GaugeValue, memberStatistic.CurrentSessions, memberLabels...)
				ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolMemberHttpRequests, prometheus.GaugeValue, memberStatistic.HttpRequests, memberLabels...)
				ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolMemberMaxSessions, prometheus.GaugeValue, memberStatistic.MaxSessions, memberLabels...)
				ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolMemberPacketsIn, prometheus.GaugeValue, memberStatistic.PacketsIn, memberLabels...)
				ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolMemberPacketsOut, prometheus.GaugeValue, memberStatistic.PacketsOut, memberLabels...)
				ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolMemberSourceIPPersistenceEntrySize, prometheus.GaugeValue, memberStatistic.SourceIPPersistenceEntrySize, memberLabels...)
				ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolMemberTotalSessions, prometheus.GaugeValue, memberStatistic.TotalSessions, memberLabels...)
			}
		}
		for _, virtualServerStatistic := range metric.loadBalancerVirtualServerStatisticMetrics {
			virtualServerLabels := []string{virtualServerStatistic.ID, metric.ID}
			ch <- prometheus.MustNewConstMetric(c.loadBalancerVirtualServerBytesIn, prometheus.GaugeValue, virtualServerStatistic.BytesIn, virtualServerLabels...)
			ch <- prometheus.MustNewConstMetric(c.loadBalancerVirtualServerBytesOut, prometheus.GaugeValue, virtualServerStatistic.BytesOut, virtualServerLabels...)
			ch <- prometheus.MustNewConstMetric(c.loadBalancerVirtualServerCurrentSessions, prometheus.GaugeValue, virtualServerStatistic.CurrentSessions, virtualServerLabels...)
			ch <- prometheus.MustNewConstMetric(c.loadBalancerVirtualServerHttpRequests, prometheus.GaugeValue, virtualServerStatistic.HttpRequests, virtualServerLabels...)
			ch <- prometheus.MustNewConstMetric(c.loadBalancerVirtualServerMaxSessions, prometheus.GaugeValue, virtualServerStatistic.MaxSessions, virtualServerLabels...)
			ch <- prometheus.MustNewConstMetric(c.loadBalancerVirtualServerPacketsIn, prometheus.GaugeValue, virtualServerStatistic.PacketsIn, virtualServerLabels...)
			ch <- prometheus.MustNewConstMetric(c.loadBalancerVirtualServerPacketsOut, prometheus.GaugeValue, virtualServerStatistic.PacketsOut, virtualServerLabels...)
			ch <- prometheus.MustNewConstMetric(c.loadBalancerVirtualServerSourceIPPersistenceEntrySize, prometheus.GaugeValue, virtualServerStatistic.SourceIPPersistenceEntrySize, virtualServerLabels...)
			ch <- prometheus.MustNewConstMetric(c.loadBalancerVirtualServerTotalSessions, prometheus.GaugeValue, virtualServerStatistic.TotalSessions, virtualServerLabels...)
		}
	}
	return
}

func (c *loadBalancerCollector) generateLoadBalancerStatusMetrics(loadBalancers []loadbalancer.LbService) (loadBalancerStatusMetrics []loadBalancerStatusMetric) {
	for _, lb := range loadBalancers {
		lbStatus, err := c.client.GetLoadBalancerStatus(lb.Id)
		if err != nil {
			level.Error(c.logger).Log("msg", "Unable to get load balancer status", "id", lb.Id, "err", err)
			continue
		}
		loadBalancerStatusMetric := loadBalancerStatusMetric{
			ID:           lbStatus.ServiceId,
			Name:         lb.DisplayName,
			StatusDetail: c.constructStatusDetail(loadBalancerPossibleStatus, lbStatus.ServiceStatus),
		}
		for _, poolStatus := range lbStatus.Pools {
			poolStatusMetric := loadBalancerPoolStatusMetric{
				ID:           poolStatus.PoolId,
				StatusDetail: c.constructStatusDetail(loadBalancerPoolPossibleStatus, poolStatus.Status),
			}
			for _, memberStatus := range poolStatus.Members {
				memberStatusMetric := loadBalancerPoolMemberStatusMetric{
					IPAddress:    memberStatus.IPAddress,
					Port:         memberStatus.Port,
					StatusDetail: c.constructStatusDetail(loadBalancerPoolMemberPossibleStatus, memberStatus.Status),
				}
				poolStatusMetric.MembersStatus = append(poolStatusMetric.MembersStatus, memberStatusMetric)
			}
			loadBalancerStatusMetric.PoolsStatus = append(loadBalancerStatusMetric.PoolsStatus, poolStatusMetric)
		}
		loadBalancerStatusMetrics = append(loadBalancerStatusMetrics, loadBalancerStatusMetric)
	}
	return
}

func (c *loadBalancerCollector) constructStatusDetail(possibleStatus []string, currentStatus string) map[string]float64 {
	statusDetail := map[string]float64{}
	for _, status := range possibleStatus {
		statusValue := 0.0
		if status == strings.ToUpper(currentStatus) {
			statusValue = 1.0
		}
		statusDetail[status] = statusValue
	}
	return statusDetail
}

func (c *loadBalancerCollector) generateLoadBalancerStatisticMetrics(loadBalancers []loadbalancer.LbService) (loadBalancerStatisticMetrics []loadBalancerStatisticMetric) {
	for _, lb := range loadBalancers {
		lbStatistic, err := c.client.GetLoadBalancerStatistic(lb.Id)
		if err != nil {
			level.Error(c.logger).Log("msg", "Unable to get load balancer statistic", "id", lb.Id, "err", err)
			continue
		}
		loadBalancerStatisticMetric := loadBalancerStatisticMetric{
			ID:                lb.Id,
			Name:              lb.DisplayName,
			L4CurrentSessions: float64(lbStatistic.Statistics.L4CurrentSessions),
			L4MaxSessions:     float64(lbStatistic.Statistics.L4MaxSessions),
			L4TotalSessions:   float64(lbStatistic.Statistics.L4TotalSessions),
			L7CurrentSessions: float64(lbStatistic.Statistics.L7CurrentSessions),
			L7MaxSessions:     float64(lbStatistic.Statistics.L7MaxSessions),
			L7TotalSessions:   float64(lbStatistic.Statistics.L7TotalSessions),
		}
		var loadBalancerPoolStatisticMetrics []loadBalancerPoolStatisticMetric
		for _, poolStatistic := range lbStatistic.Pools {
			loadBalancerPoolStatisticMetric := loadBalancerPoolStatisticMetric{
				ID:                           poolStatistic.PoolId,
				BytesIn:                      float64(poolStatistic.Statistics.BytesIn),
				BytesOut:                     float64(poolStatistic.Statistics.BytesOut),
				CurrentSessions:              float64(poolStatistic.Statistics.CurrentSessions),
				HttpRequests:                 float64(poolStatistic.Statistics.HttpRequests),
				MaxSessions:                  float64(poolStatistic.Statistics.MaxSessions),
				PacketsIn:                    float64(poolStatistic.Statistics.PacketsIn),
				PacketsOut:                   float64(poolStatistic.Statistics.PacketsOut),
				SourceIPPersistenceEntrySize: float64(poolStatistic.Statistics.SourceIPPersistenceEntrySize),
				TotalSessions:                float64(poolStatistic.Statistics.TotalSessions),
			}
			var loadBalancerPoolMemberStatisticMetrics []loadBalancerPoolMemberStatisticMetric
			for _, memberStatistic := range poolStatistic.Members {
				loadBalancerPoolMemberStatisticMetric := loadBalancerPoolMemberStatisticMetric{
					IPAddress:                    memberStatistic.IPAddress,
					Port:                         memberStatistic.Port,
					BytesIn:                      float64(memberStatistic.Statistics.BytesIn),
					BytesOut:                     float64(memberStatistic.Statistics.BytesOut),
					CurrentSessions:              float64(memberStatistic.Statistics.CurrentSessions),
					HttpRequests:                 float64(memberStatistic.Statistics.HttpRequests),
					MaxSessions:                  float64(memberStatistic.Statistics.MaxSessions),
					PacketsIn:                    float64(memberStatistic.Statistics.PacketsIn),
					PacketsOut:                   float64(memberStatistic.Statistics.PacketsOut),
					SourceIPPersistenceEntrySize: float64(memberStatistic.Statistics.SourceIPPersistenceEntrySize),
					TotalSessions:                float64(memberStatistic.Statistics.TotalSessions),
				}
				loadBalancerPoolMemberStatisticMetrics = append(loadBalancerPoolMemberStatisticMetrics, loadBalancerPoolMemberStatisticMetric)
			}
			loadBalancerPoolStatisticMetric.loadBalancerPoolMemberStatisticMetrics = loadBalancerPoolMemberStatisticMetrics
			loadBalancerPoolStatisticMetrics = append(loadBalancerPoolStatisticMetrics, loadBalancerPoolStatisticMetric)
		}
		var loadBalancerVirtualServerStatisticMetrics []loadBalancerVirtualServerStatisticMetric
		for _, virtualServerStatistics := range lbStatistic.VirtualServes {
			loadBalancerVirtualServerStatisticMetric := loadBalancerVirtualServerStatisticMetric{
				ID:                           virtualServerStatistics.VirtualServerId,
				BytesIn:                      float64(virtualServerStatistics.Statistics.BytesIn),
				BytesOut:                     float64(virtualServerStatistics.Statistics.BytesOut),
				CurrentSessions:              float64(virtualServerStatistics.Statistics.CurrentSessions),
				HttpRequests:                 float64(virtualServerStatistics.Statistics.HttpRequests),
				MaxSessions:                  float64(virtualServerStatistics.Statistics.MaxSessions),
				PacketsIn:                    float64(virtualServerStatistics.Statistics.PacketsIn),
				PacketsOut:                   float64(virtualServerStatistics.Statistics.PacketsOut),
				SourceIPPersistenceEntrySize: float64(virtualServerStatistics.Statistics.SourceIPPersistenceEntrySize),
				TotalSessions:                float64(virtualServerStatistics.Statistics.TotalSessions),
			}
			loadBalancerVirtualServerStatisticMetrics = append(loadBalancerVirtualServerStatisticMetrics, loadBalancerVirtualServerStatisticMetric)
		}
		loadBalancerStatisticMetric.loadBalancerPoolStatisticMetrics = loadBalancerPoolStatisticMetrics
		loadBalancerStatisticMetric.loadBalancerVirtualServerStatisticMetrics = loadBalancerVirtualServerStatisticMetrics
		loadBalancerStatisticMetrics = append(loadBalancerStatisticMetrics, loadBalancerStatisticMetric)
	}
	return
}
