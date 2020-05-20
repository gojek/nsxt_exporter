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

var loadBalancerPossibleStatus = [...]string{"UP", "DOWN", "ERROR", "NO_STANDBY", "DETACHED", "DISABLED", "UNKNOWN"}

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
	Status        float64
	MembersStatus []loadBalancerPoolMemberStatusMetric
}

type loadBalancerPoolMemberStatusMetric struct {
	IPAddress string
	Port      string
	Status    float64
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
		"Status of Load Balancer pool UP/DOWN",
		[]string{"id", "load_balancer_id"},
		nil,
	)
	loadBalancerPoolMemberStatus := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "load_balancer", "pool_member_status"),
		"Status of Load Balancer pool member UP/DOWN",
		[]string{"ip_address", "port", "load_balancer_pool_id", "load_balancer_id"},
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
			ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolStatus, prometheus.GaugeValue, poolStatus.Status, poolStatus.ID, metric.ID)
			for _, memberStatus := range poolStatus.MembersStatus {
				ch <- prometheus.MustNewConstMetric(c.loadBalancerPoolMemberStatus, prometheus.GaugeValue, memberStatus.Status, memberStatus.IPAddress, memberStatus.Port, poolStatus.ID, metric.ID)
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
			StatusDetail: map[string]float64{},
		}
		for _, status := range loadBalancerPossibleStatus {
			statusValue := 0.0
			if status == strings.ToUpper(lbStatus.ServiceStatus) {
				statusValue = 1.0
			}
			loadBalancerStatusMetric.StatusDetail[status] = statusValue
		}
		for _, poolStatus := range lbStatus.Pools {
			poolStatusValue := 0.0
			if strings.ToUpper(poolStatus.Status) == "UP" {
				poolStatusValue = 1.0
			}
			poolStatusMetric := loadBalancerPoolStatusMetric{
				ID:     poolStatus.PoolId,
				Status: poolStatusValue,
			}
			for _, memberStatus := range poolStatus.Members {
				memberStatusValue := 0.0
				if strings.ToUpper(memberStatus.Status) == "UP" {
					memberStatusValue = 1.0
				}
				memberStatusMetric := loadBalancerPoolMemberStatusMetric{
					IPAddress: memberStatus.IPAddress,
					Port:      memberStatus.Port,
					Status:    memberStatusValue,
				}
				poolStatusMetric.MembersStatus = append(poolStatusMetric.MembersStatus, memberStatusMetric)
			}
			loadBalancerStatusMetric.PoolsStatus = append(loadBalancerStatusMetric.PoolsStatus, poolStatusMetric)
		}
		loadBalancerStatusMetrics = append(loadBalancerStatusMetrics, loadBalancerStatusMetric)
	}
	return
}
