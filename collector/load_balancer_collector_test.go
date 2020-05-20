package collector

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/go-vmware-nsxt/loadbalancer"
)

const (
	fakeLoadBalancerID             = "fake-load-balancer-id"
	fakeLoadBalancerName           = "fake-load-balancer-name"
	fakeLoadBalancerPoolID         = "fake-load-balancer-pool-id"
	fakeLoadbalancerPoolMemberIP   = "127.0.0.1"
	fakeLoadbalancerPoolMemberPort = "9732"
)

type mockLoadBalancerClient struct {
	responses []mockLoadBalancerResponse
}

type mockLoadBalancerResponse struct {
	ID               string
	Name             string
	Status           string
	PoolID           string
	PoolStatus       string
	PoolMemberStatus string
	Error            error
}

func (c *mockLoadBalancerClient) ListAllLoadBalancers() ([]loadbalancer.LbService, error) {
	panic("unused function. Only used to satisfy LoadBalancerClient interface")
}

func (c *mockLoadBalancerClient) GetLoadBalancerStatus(loadBalancerID string) (loadbalancer.LbServiceStatus, error) {
	for _, res := range c.responses {
		if res.ID == loadBalancerID {
			return loadbalancer.LbServiceStatus{
				ServiceId:     res.ID,
				ServiceStatus: res.Status,
				Pools: []loadbalancer.LbPoolStatus{
					{
						PoolId: res.PoolID,
						Status: res.PoolStatus,
						Members: []loadbalancer.LbPoolMemberStatus{
							{
								IPAddress: fakeLoadbalancerPoolMemberIP,
								Port:      fakeLoadbalancerPoolMemberPort,
								Status:    res.PoolMemberStatus,
							},
						},
					},
				},
			}, res.Error
		}
	}
	return loadbalancer.LbServiceStatus{}, errors.New("load balancer not found")
}

func buildLoadBalancerStatusResponse(id string, status string, poolStatus string, poolMemberStatus string, err error) mockLoadBalancerResponse {
	return mockLoadBalancerResponse{
		ID:               fmt.Sprintf("%s-%s", fakeLoadBalancerID, id),
		Name:             fmt.Sprintf("%s-%s", fakeLoadBalancerName, id),
		Status:           status,
		PoolID:           fmt.Sprintf("%s-%s", fakeLoadBalancerPoolID, id),
		PoolStatus:       poolStatus,
		PoolMemberStatus: poolMemberStatus,
		Error:            err,
	}
}

func buildLoadBalancers(loadBalancerResponses []mockLoadBalancerResponse) []loadbalancer.LbService {
	var loadBalancers []loadbalancer.LbService
	for _, res := range loadBalancerResponses {
		loadBalancer := loadbalancer.LbService{
			Id:          res.ID,
			DisplayName: res.Name,
		}
		loadBalancers = append(loadBalancers, loadBalancer)
	}
	return loadBalancers
}

func TestLoadBalancerCollector_GenerateLoadBalancerStatusMetrics(t *testing.T) {
	testcases := []struct {
		description           string
		loadBalancerResponses []mockLoadBalancerResponse
		expectedMetrics       []loadBalancerStatusMetric
	}{
		{
			description: "Should return correct status value depending on load balancer state",
			loadBalancerResponses: []mockLoadBalancerResponse{
				buildLoadBalancerStatusResponse("01", "UP", "UP", "UP", nil),
				buildLoadBalancerStatusResponse("02", "DOWN", "PARTIALLY_UP", "DOWN", nil),
				buildLoadBalancerStatusResponse("03", "ERROR", "PRIMARY_DOWN", "DISABLED", nil),
				buildLoadBalancerStatusResponse("04", "NO_STANDBY", "DOWN", "GRACEFUL_DISABLED", nil),
				buildLoadBalancerStatusResponse("05", "DETACHED", "DETACHED", "UNUSED", nil),
				buildLoadBalancerStatusResponse("06", "DISABLED", "UNKNOWN", "up", nil),
				buildLoadBalancerStatusResponse("07", "UNKNOWN", "up", "down", nil),
			},
			expectedMetrics: []loadBalancerStatusMetric{
				{
					ID:   "fake-load-balancer-id-01",
					Name: "fake-load-balancer-name-01",
					StatusDetail: map[string]float64{
						"UP":         1.0,
						"DOWN":       0.0,
						"ERROR":      0.0,
						"NO_STANDBY": 0.0,
						"DETACHED":   0.0,
						"DISABLED":   0.0,
						"UNKNOWN":    0.0,
					},
					PoolsStatus: []loadBalancerPoolStatusMetric{
						{
							ID: "fake-load-balancer-pool-id-01",
							StatusDetail: map[string]float64{
								"UP":           1.0,
								"PARTIALLY_UP": 0.0,
								"PRIMARY_DOWN": 0.0,
								"DOWN":         0.0,
								"DETACHED":     0.0,
								"UNKNOWN":      0.0,
							},
							MembersStatus: []loadBalancerPoolMemberStatusMetric{
								{
									IPAddress: fakeLoadbalancerPoolMemberIP,
									Port:      fakeLoadbalancerPoolMemberPort,
									Status:    1.0,
								},
							},
						},
					},
				}, {
					ID:   "fake-load-balancer-id-02",
					Name: "fake-load-balancer-name-02",
					StatusDetail: map[string]float64{
						"UP":         0.0,
						"DOWN":       1.0,
						"ERROR":      0.0,
						"NO_STANDBY": 0.0,
						"DETACHED":   0.0,
						"DISABLED":   0.0,
						"UNKNOWN":    0.0,
					},
					PoolsStatus: []loadBalancerPoolStatusMetric{
						{
							ID: "fake-load-balancer-pool-id-02",
							StatusDetail: map[string]float64{
								"UP":           0.0,
								"PARTIALLY_UP": 1.0,
								"PRIMARY_DOWN": 0.0,
								"DOWN":         0.0,
								"DETACHED":     0.0,
								"UNKNOWN":      0.0,
							},
							MembersStatus: []loadBalancerPoolMemberStatusMetric{
								{
									IPAddress: fakeLoadbalancerPoolMemberIP,
									Port:      fakeLoadbalancerPoolMemberPort,
									Status:    0.0,
								},
							},
						},
					},
				}, {
					ID:   "fake-load-balancer-id-03",
					Name: "fake-load-balancer-name-03",
					StatusDetail: map[string]float64{
						"UP":         0.0,
						"DOWN":       0.0,
						"ERROR":      1.0,
						"NO_STANDBY": 0.0,
						"DETACHED":   0.0,
						"DISABLED":   0.0,
						"UNKNOWN":    0.0,
					},
					PoolsStatus: []loadBalancerPoolStatusMetric{
						{
							ID: "fake-load-balancer-pool-id-03",
							StatusDetail: map[string]float64{
								"UP":           0.0,
								"PARTIALLY_UP": 0.0,
								"PRIMARY_DOWN": 1.0,
								"DOWN":         0.0,
								"DETACHED":     0.0,
								"UNKNOWN":      0.0,
							},
							MembersStatus: []loadBalancerPoolMemberStatusMetric{
								{
									IPAddress: fakeLoadbalancerPoolMemberIP,
									Port:      fakeLoadbalancerPoolMemberPort,
									Status:    0.0,
								},
							},
						},
					},
				}, {
					ID:   "fake-load-balancer-id-04",
					Name: "fake-load-balancer-name-04",
					StatusDetail: map[string]float64{
						"UP":         0.0,
						"DOWN":       0.0,
						"ERROR":      0.0,
						"NO_STANDBY": 1.0,
						"DETACHED":   0.0,
						"DISABLED":   0.0,
						"UNKNOWN":    0.0,
					},
					PoolsStatus: []loadBalancerPoolStatusMetric{
						{
							ID: "fake-load-balancer-pool-id-04",
							StatusDetail: map[string]float64{
								"UP":           0.0,
								"PARTIALLY_UP": 0.0,
								"PRIMARY_DOWN": 0.0,
								"DOWN":         1.0,
								"DETACHED":     0.0,
								"UNKNOWN":      0.0,
							},
							MembersStatus: []loadBalancerPoolMemberStatusMetric{
								{
									IPAddress: fakeLoadbalancerPoolMemberIP,
									Port:      fakeLoadbalancerPoolMemberPort,
									Status:    0.0,
								},
							},
						},
					},
				}, {
					ID:   "fake-load-balancer-id-05",
					Name: "fake-load-balancer-name-05",
					StatusDetail: map[string]float64{
						"UP":         0.0,
						"DOWN":       0.0,
						"ERROR":      0.0,
						"NO_STANDBY": 0.0,
						"DETACHED":   1.0,
						"DISABLED":   0.0,
						"UNKNOWN":    0.0,
					},
					PoolsStatus: []loadBalancerPoolStatusMetric{
						{
							ID: "fake-load-balancer-pool-id-05",
							StatusDetail: map[string]float64{
								"UP":           0.0,
								"PARTIALLY_UP": 0.0,
								"PRIMARY_DOWN": 0.0,
								"DOWN":         0.0,
								"DETACHED":     1.0,
								"UNKNOWN":      0.0,
							},
							MembersStatus: []loadBalancerPoolMemberStatusMetric{
								{
									IPAddress: fakeLoadbalancerPoolMemberIP,
									Port:      fakeLoadbalancerPoolMemberPort,
									Status:    0.0,
								},
							},
						},
					},
				}, {
					ID:   "fake-load-balancer-id-06",
					Name: "fake-load-balancer-name-06",
					StatusDetail: map[string]float64{
						"UP":         0.0,
						"DOWN":       0.0,
						"ERROR":      0.0,
						"NO_STANDBY": 0.0,
						"DETACHED":   0.0,
						"DISABLED":   1.0,
						"UNKNOWN":    0.0,
					},
					PoolsStatus: []loadBalancerPoolStatusMetric{
						{
							ID: "fake-load-balancer-pool-id-06",
							StatusDetail: map[string]float64{
								"UP":           0.0,
								"PARTIALLY_UP": 0.0,
								"PRIMARY_DOWN": 0.0,
								"DOWN":         0.0,
								"DETACHED":     0.0,
								"UNKNOWN":      1.0,
							},
							MembersStatus: []loadBalancerPoolMemberStatusMetric{
								{
									IPAddress: fakeLoadbalancerPoolMemberIP,
									Port:      fakeLoadbalancerPoolMemberPort,
									Status:    1.0,
								},
							},
						},
					},
				}, {
					ID:   "fake-load-balancer-id-07",
					Name: "fake-load-balancer-name-07",
					StatusDetail: map[string]float64{
						"UP":         0.0,
						"DOWN":       0.0,
						"ERROR":      0.0,
						"NO_STANDBY": 0.0,
						"DETACHED":   0.0,
						"DISABLED":   0.0,
						"UNKNOWN":    1.0,
					},
					PoolsStatus: []loadBalancerPoolStatusMetric{
						{
							ID: "fake-load-balancer-pool-id-07",
							StatusDetail: map[string]float64{
								"UP":           1.0,
								"PARTIALLY_UP": 0.0,
								"PRIMARY_DOWN": 0.0,
								"DOWN":         0.0,
								"DETACHED":     0.0,
								"UNKNOWN":      0.0,
							},
							MembersStatus: []loadBalancerPoolMemberStatusMetric{
								{
									IPAddress: fakeLoadbalancerPoolMemberIP,
									Port:      fakeLoadbalancerPoolMemberPort,
									Status:    0.0,
								},
							},
						},
					},
				},
			},
		}, {
			description: "Should only return metric with valid response",
			loadBalancerResponses: []mockLoadBalancerResponse{
				buildLoadBalancerStatusResponse("01", "UP", "UP", "UP", nil),
				buildLoadBalancerStatusResponse("02", "UP", "UP", "UP", errors.New("unable to get load balancer status")),
			},
			expectedMetrics: []loadBalancerStatusMetric{
				{
					ID:   "fake-load-balancer-id-01",
					Name: "fake-load-balancer-name-01",
					StatusDetail: map[string]float64{
						"UP":         1.0,
						"DOWN":       0.0,
						"ERROR":      0.0,
						"NO_STANDBY": 0.0,
						"DETACHED":   0.0,
						"DISABLED":   0.0,
						"UNKNOWN":    0.0,
					},
					PoolsStatus: []loadBalancerPoolStatusMetric{
						{
							ID: "fake-load-balancer-pool-id-01",
							StatusDetail: map[string]float64{
								"UP":           1.0,
								"PARTIALLY_UP": 0.0,
								"PRIMARY_DOWN": 0.0,
								"DOWN":         0.0,
								"DETACHED":     0.0,
								"UNKNOWN":      0.0,
							},
							MembersStatus: []loadBalancerPoolMemberStatusMetric{
								{
									IPAddress: fakeLoadbalancerPoolMemberIP,
									Port:      fakeLoadbalancerPoolMemberPort,
									Status:    1.0,
								},
							},
						},
					},
				},
			},
		}, {
			description:           "Should return empty metrics when given empty load balancer",
			loadBalancerResponses: []mockLoadBalancerResponse{},
			expectedMetrics:       []loadBalancerStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockLoadBalancerClient := &mockLoadBalancerClient{
			responses: tc.loadBalancerResponses,
		}
		loadBalancers := buildLoadBalancers(tc.loadBalancerResponses)
		logger := log.NewNopLogger()
		loadBalancerCollector := newLoadBalancerCollector(mockLoadBalancerClient, logger)
		loadBalancerStatusMetrics := loadBalancerCollector.generateLoadBalancerStatusMetrics(loadBalancers)
		assert.ElementsMatch(t, tc.expectedMetrics, loadBalancerStatusMetrics, tc.description)
	}
}
