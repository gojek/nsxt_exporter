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

	loadBalancerStatus           *prometheus.Desc
	loadBalancerPoolStatus       *prometheus.Desc
	loadBalancerPoolMemberStatus *prometheus.Desc
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
	return &loadBalancerCollector{
		client: client,
		logger: logger,

		loadBalancerStatus:           loadBalancerStatus,
		loadBalancerPoolStatus:       loadBalancerPoolStatus,
		loadBalancerPoolMemberStatus: loadBalancerPoolMemberStatus,
	}
}

// Describe implements the prometheus.Collector interface.
func (c *loadBalancerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.loadBalancerStatus
	ch <- c.loadBalancerPoolStatus
	ch <- c.loadBalancerPoolMemberStatus
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
