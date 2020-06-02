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
	fakeLoadBalancerID                                        = "fake-load-balancer-id"
	fakeLoadBalancerName                                      = "fake-load-balancer-name"
	fakeLoadBalancerPoolID                                    = "fake-load-balancer-pool-id"
	fakeLoadBalancerVirtualServerID                           = "fake-load-balancer-virtual-server"
	fakeLoadbalancerPoolMemberIP                              = "127.0.0.1"
	fakeLoadbalancerPoolMemberPort                            = "9744"
	fakeLoadBalancerL4CurrentSessions                         = iota
	fakeLoadBalancerL4MaxSessions                             = iota
	fakeLoadBalancerL4TotalSessions                           = iota
	fakeLoadBalancerL7CurrentSessions                         = iota
	fakeLoadBalancerL7MaxSessions                             = iota
	fakeLoadBalancerL7TotalSessions                           = iota
	fakeLoadBalancerPoolBytesIn                               = iota
	fakeLoadBalancerPoolBytesOut                              = iota
	fakeLoadBalancerPoolCurrentSessions                       = iota
	fakeLoadBalancerPoolHttpRequests                          = iota
	fakeLoadBalancerPoolMaxSessions                           = iota
	fakeLoadBalancerPoolPacketsIn                             = iota
	fakeLoadBalancerPoolPacketsOut                            = iota
	fakeLoadBalancerPoolSourceIPPersistenceEntrySize          = iota
	fakeLoadBalancerPoolTotalSessions                         = iota
	fakeLoadBalancerPoolMemberBytesIn                         = iota
	fakeLoadBalancerPoolMemberBytesOut                        = iota
	fakeLoadBalancerPoolMemberCurrentSessions                 = iota
	fakeLoadBalancerPoolMemberHttpRequests                    = iota
	fakeLoadBalancerPoolMemberMaxSessions                     = iota
	fakeLoadBalancerPoolMemberPacketsIn                       = iota
	fakeLoadBalancerPoolMemberPacketsOut                      = iota
	fakeLoadBalancerPoolMemberSourceIPPersistenceEntrySize    = iota
	fakeLoadBalancerPoolMemberTotalSessions                   = iota
	fakeLoadBalancerVirtualServerBytesIn                      = iota
	fakeLoadBalancerVirtualServerBytesOut                     = iota
	fakeLoadBalancerVirtualServerCurrentSessions              = iota
	fakeLoadBalancerVirtualServerHttpRequests                 = iota
	fakeLoadBalancerVirtualServerMaxSessions                  = iota
	fakeLoadBalancerVirtualServerPacketsIn                    = iota
	fakeLoadBalancerVirtualServerPacketsOut                   = iota
	fakeLoadBalancerVirtualServerSourceIPPersistenceEntrySize = iota
	fakeLoadBalancerVirtualServerTotalSessions                = iota
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
	VirtualServerID  string
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

func (c *mockLoadBalancerClient) GetLoadBalancerStatistic(loadBalancerID string) (loadbalancer.LbServiceStatistics, error) {
	for _, res := range c.responses {
		if res.ID == loadBalancerID {
			return loadbalancer.LbServiceStatistics{
				Statistics: loadbalancer.LbServiceStatisticsCounter{
					L4CurrentSessions: fakeLoadBalancerL4CurrentSessions,
					L4MaxSessions:     fakeLoadBalancerL4MaxSessions,
					L4TotalSessions:   fakeLoadBalancerL4TotalSessions,
					L7CurrentSessions: fakeLoadBalancerL7CurrentSessions,
					L7MaxSessions:     fakeLoadBalancerL7MaxSessions,
					L7TotalSessions:   fakeLoadBalancerL7TotalSessions,
				},
				Pools: []loadbalancer.LbPoolStatistics{
					{
						PoolId: res.PoolID,
						Statistics: loadbalancer.LbStatisticsCounter{
							BytesIn:                      fakeLoadBalancerPoolBytesIn,
							BytesOut:                     fakeLoadBalancerPoolBytesOut,
							CurrentSessions:              fakeLoadBalancerPoolCurrentSessions,
							HttpRequests:                 fakeLoadBalancerPoolHttpRequests,
							MaxSessions:                  fakeLoadBalancerPoolMaxSessions,
							PacketsIn:                    fakeLoadBalancerPoolPacketsIn,
							PacketsOut:                   fakeLoadBalancerPoolPacketsOut,
							SourceIPPersistenceEntrySize: fakeLoadBalancerPoolSourceIPPersistenceEntrySize,
							TotalSessions:                fakeLoadBalancerPoolTotalSessions,
						},
						Members: []loadbalancer.LbPoolMemberStatistics{
							{
								IPAddress: fakeLoadbalancerPoolMemberIP,
								Port:      fakeLoadbalancerPoolMemberPort,
								Statistics: loadbalancer.LbStatisticsCounter{
									BytesIn:                      fakeLoadBalancerPoolMemberBytesIn,
									BytesOut:                     fakeLoadBalancerPoolMemberBytesOut,
									CurrentSessions:              fakeLoadBalancerPoolMemberCurrentSessions,
									HttpRequests:                 fakeLoadBalancerPoolMemberHttpRequests,
									MaxSessions:                  fakeLoadBalancerPoolMemberMaxSessions,
									PacketsIn:                    fakeLoadBalancerPoolMemberPacketsIn,
									PacketsOut:                   fakeLoadBalancerPoolMemberPacketsOut,
									SourceIPPersistenceEntrySize: fakeLoadBalancerPoolMemberSourceIPPersistenceEntrySize,
									TotalSessions:                fakeLoadBalancerPoolMemberTotalSessions,
								},
							},
						},
					},
				},
				VirtualServes: []loadbalancer.LbVirtualServerStatistics{
					{
						VirtualServerId: res.VirtualServerID,
						Statistics: loadbalancer.LbStatisticsCounter{
							BytesIn:                      fakeLoadBalancerVirtualServerBytesIn,
							BytesOut:                     fakeLoadBalancerVirtualServerBytesOut,
							CurrentSessions:              fakeLoadBalancerVirtualServerCurrentSessions,
							HttpRequests:                 fakeLoadBalancerVirtualServerHttpRequests,
							MaxSessions:                  fakeLoadBalancerVirtualServerMaxSessions,
							PacketsIn:                    fakeLoadBalancerVirtualServerPacketsIn,
							PacketsOut:                   fakeLoadBalancerVirtualServerPacketsOut,
							SourceIPPersistenceEntrySize: fakeLoadBalancerVirtualServerSourceIPPersistenceEntrySize,
							TotalSessions:                fakeLoadBalancerVirtualServerTotalSessions,
						},
					},
				},
			}, res.Error
		}
	}
	return loadbalancer.LbServiceStatistics{}, errors.New("load balancer not found")
}

func buildLoadBalancerStatusResponse(id string, status string, poolStatus string, poolMemberStatus string, err error) mockLoadBalancerResponse {
	return mockLoadBalancerResponse{
		ID:               fmt.Sprintf("%s-%s", fakeLoadBalancerID, id),
		Name:             fmt.Sprintf("%s-%s", fakeLoadBalancerName, id),
		Status:           status,
		PoolID:           fmt.Sprintf("%s-%s", fakeLoadBalancerPoolID, id),
		PoolStatus:       poolStatus,
		PoolMemberStatus: poolMemberStatus,
		VirtualServerID:  fmt.Sprintf("%s-%s", fakeLoadBalancerVirtualServerID, id),
		Error:            err,
	}
}

func buildLoadBalancerStatisticResponse(id string, err error) mockLoadBalancerResponse {
	return mockLoadBalancerResponse{
		ID:              fmt.Sprintf("%s-%s", fakeLoadBalancerID, id),
		Name:            fmt.Sprintf("%s-%s", fakeLoadBalancerName, id),
		PoolID:          fmt.Sprintf("%s-%s", fakeLoadBalancerPoolID, id),
		VirtualServerID: fmt.Sprintf("%s-%s", fakeLoadBalancerVirtualServerID, id),
		Error:           err,
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

func buildExpectedLoadBalancerStatusDetails(nonZeroStatus string) map[string]float64 {
	statusDetails := map[string]float64{
		"UP":         0.0,
		"DOWN":       0.0,
		"ERROR":      0.0,
		"NO_STANDBY": 0.0,
		"DETACHED":   0.0,
		"DISABLED":   0.0,
		"UNKNOWN":    0.0,
	}
	statusDetails[nonZeroStatus] = 1.0
	return statusDetails
}

func buildExpectedLoadBalancerPoolStatusDetails(nonZeroStatus string) map[string]float64 {
	statusDetails := map[string]float64{
		"UP":           0.0,
		"PARTIALLY_UP": 0.0,
		"PRIMARY_DOWN": 0.0,
		"DOWN":         0.0,
		"DETACHED":     0.0,
		"UNKNOWN":      0.0,
	}
	statusDetails[nonZeroStatus] = 1.0
	return statusDetails
}

func buildExpectedLoadBalancerPoolMemberStatusDetails(nonZeroStatus string) map[string]float64 {
	statusDetails := map[string]float64{
		"UP":                0.0,
		"DOWN":              0.0,
		"DISABLED":          0.0,
		"GRACEFUL_DISABLED": 0.0,
		"UNUSED":            0.0,
	}
	statusDetails[nonZeroStatus] = 1.0
	return statusDetails
}

func buildExpectedLoadBalancerStatistic(id string) loadBalancerStatisticMetric {
	return loadBalancerStatisticMetric{
		ID:                fmt.Sprintf("%s-%s", fakeLoadBalancerID, id),
		Name:              fmt.Sprintf("%s-%s", fakeLoadBalancerName, id),
		L4CurrentSessions: float64(fakeLoadBalancerL4CurrentSessions),
		L4MaxSessions:     float64(fakeLoadBalancerL4MaxSessions),
		L4TotalSessions:   float64(fakeLoadBalancerL4TotalSessions),
		L7CurrentSessions: float64(fakeLoadBalancerL7CurrentSessions),
		L7MaxSessions:     float64(fakeLoadBalancerL7MaxSessions),
		L7TotalSessions:   float64(fakeLoadBalancerL7TotalSessions),
		loadBalancerPoolStatisticMetrics: []loadBalancerPoolStatisticMetric{
			{
				ID:                           fmt.Sprintf("%s-%s", fakeLoadBalancerPoolID, id),
				BytesIn:                      float64(fakeLoadBalancerPoolBytesIn),
				BytesOut:                     float64(fakeLoadBalancerPoolBytesOut),
				CurrentSessions:              float64(fakeLoadBalancerPoolCurrentSessions),
				HttpRequests:                 float64(fakeLoadBalancerPoolHttpRequests),
				MaxSessions:                  float64(fakeLoadBalancerPoolMaxSessions),
				PacketsIn:                    float64(fakeLoadBalancerPoolPacketsIn),
				PacketsOut:                   float64(fakeLoadBalancerPoolPacketsOut),
				SourceIPPersistenceEntrySize: float64(fakeLoadBalancerPoolSourceIPPersistenceEntrySize),
				TotalSessions:                float64(fakeLoadBalancerPoolTotalSessions),
				loadBalancerPoolMemberStatisticMetrics: []loadBalancerPoolMemberStatisticMetric{
					{
						IPAddress:                    fakeLoadbalancerPoolMemberIP,
						Port:                         fakeLoadbalancerPoolMemberPort,
						BytesIn:                      float64(fakeLoadBalancerPoolMemberBytesIn),
						BytesOut:                     float64(fakeLoadBalancerPoolMemberBytesOut),
						CurrentSessions:              float64(fakeLoadBalancerPoolMemberCurrentSessions),
						HttpRequests:                 float64(fakeLoadBalancerPoolMemberHttpRequests),
						MaxSessions:                  float64(fakeLoadBalancerPoolMemberMaxSessions),
						PacketsIn:                    float64(fakeLoadBalancerPoolMemberPacketsIn),
						PacketsOut:                   float64(fakeLoadBalancerPoolMemberPacketsOut),
						SourceIPPersistenceEntrySize: float64(fakeLoadBalancerPoolMemberSourceIPPersistenceEntrySize),
						TotalSessions:                float64(fakeLoadBalancerPoolMemberTotalSessions),
					},
				},
			},
		},
		loadBalancerVirtualServerStatisticMetrics: []loadBalancerVirtualServerStatisticMetric{
			{
				ID:                           fmt.Sprintf("%s-%s", fakeLoadBalancerVirtualServerID, id),
				BytesIn:                      float64(fakeLoadBalancerVirtualServerBytesIn),
				BytesOut:                     float64(fakeLoadBalancerVirtualServerBytesOut),
				CurrentSessions:              float64(fakeLoadBalancerVirtualServerCurrentSessions),
				HttpRequests:                 float64(fakeLoadBalancerVirtualServerHttpRequests),
				MaxSessions:                  float64(fakeLoadBalancerVirtualServerMaxSessions),
				PacketsIn:                    float64(fakeLoadBalancerVirtualServerPacketsIn),
				PacketsOut:                   float64(fakeLoadBalancerVirtualServerPacketsOut),
				SourceIPPersistenceEntrySize: float64(fakeLoadBalancerVirtualServerSourceIPPersistenceEntrySize),
				TotalSessions:                float64(fakeLoadBalancerVirtualServerTotalSessions),
			},
		},
	}
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
					ID:           "fake-load-balancer-id-01",
					Name:         "fake-load-balancer-name-01",
					StatusDetail: buildExpectedLoadBalancerStatusDetails("UP"),
					PoolsStatus: []loadBalancerPoolStatusMetric{
						{
							ID:           "fake-load-balancer-pool-id-01",
							StatusDetail: buildExpectedLoadBalancerPoolStatusDetails("UP"),
							MembersStatus: []loadBalancerPoolMemberStatusMetric{
								{
									IPAddress:    fakeLoadbalancerPoolMemberIP,
									Port:         fakeLoadbalancerPoolMemberPort,
									StatusDetail: buildExpectedLoadBalancerPoolMemberStatusDetails("UP"),
								},
							},
						},
					},
				}, {
					ID:           "fake-load-balancer-id-02",
					Name:         "fake-load-balancer-name-02",
					StatusDetail: buildExpectedLoadBalancerStatusDetails("DOWN"),
					PoolsStatus: []loadBalancerPoolStatusMetric{
						{
							ID:           "fake-load-balancer-pool-id-02",
							StatusDetail: buildExpectedLoadBalancerPoolStatusDetails("PARTIALLY_UP"),
							MembersStatus: []loadBalancerPoolMemberStatusMetric{
								{
									IPAddress:    fakeLoadbalancerPoolMemberIP,
									Port:         fakeLoadbalancerPoolMemberPort,
									StatusDetail: buildExpectedLoadBalancerPoolMemberStatusDetails("DOWN"),
								},
							},
						},
					},
				}, {
					ID:           "fake-load-balancer-id-03",
					Name:         "fake-load-balancer-name-03",
					StatusDetail: buildExpectedLoadBalancerStatusDetails("ERROR"),
					PoolsStatus: []loadBalancerPoolStatusMetric{
						{
							ID:           "fake-load-balancer-pool-id-03",
							StatusDetail: buildExpectedLoadBalancerPoolStatusDetails("PRIMARY_DOWN"),
							MembersStatus: []loadBalancerPoolMemberStatusMetric{
								{
									IPAddress:    fakeLoadbalancerPoolMemberIP,
									Port:         fakeLoadbalancerPoolMemberPort,
									StatusDetail: buildExpectedLoadBalancerPoolMemberStatusDetails("DISABLED"),
								},
							},
						},
					},
				}, {
					ID:           "fake-load-balancer-id-04",
					Name:         "fake-load-balancer-name-04",
					StatusDetail: buildExpectedLoadBalancerStatusDetails("NO_STANDBY"),
					PoolsStatus: []loadBalancerPoolStatusMetric{
						{
							ID:           "fake-load-balancer-pool-id-04",
							StatusDetail: buildExpectedLoadBalancerPoolStatusDetails("DOWN"),
							MembersStatus: []loadBalancerPoolMemberStatusMetric{
								{
									IPAddress:    fakeLoadbalancerPoolMemberIP,
									Port:         fakeLoadbalancerPoolMemberPort,
									StatusDetail: buildExpectedLoadBalancerPoolMemberStatusDetails("GRACEFUL_DISABLED"),
								},
							},
						},
					},
				}, {
					ID:           "fake-load-balancer-id-05",
					Name:         "fake-load-balancer-name-05",
					StatusDetail: buildExpectedLoadBalancerStatusDetails("DETACHED"),
					PoolsStatus: []loadBalancerPoolStatusMetric{
						{
							ID:           "fake-load-balancer-pool-id-05",
							StatusDetail: buildExpectedLoadBalancerPoolStatusDetails("DETACHED"),
							MembersStatus: []loadBalancerPoolMemberStatusMetric{
								{
									IPAddress:    fakeLoadbalancerPoolMemberIP,
									Port:         fakeLoadbalancerPoolMemberPort,
									StatusDetail: buildExpectedLoadBalancerPoolMemberStatusDetails("UNUSED"),
								},
							},
						},
					},
				}, {
					ID:           "fake-load-balancer-id-06",
					Name:         "fake-load-balancer-name-06",
					StatusDetail: buildExpectedLoadBalancerStatusDetails("DISABLED"),
					PoolsStatus: []loadBalancerPoolStatusMetric{
						{
							ID:           "fake-load-balancer-pool-id-06",
							StatusDetail: buildExpectedLoadBalancerPoolStatusDetails("UNKNOWN"),
							MembersStatus: []loadBalancerPoolMemberStatusMetric{
								{
									IPAddress:    fakeLoadbalancerPoolMemberIP,
									Port:         fakeLoadbalancerPoolMemberPort,
									StatusDetail: buildExpectedLoadBalancerPoolMemberStatusDetails("UP"),
								},
							},
						},
					},
				}, {
					ID:           "fake-load-balancer-id-07",
					Name:         "fake-load-balancer-name-07",
					StatusDetail: buildExpectedLoadBalancerStatusDetails("UNKNOWN"),
					PoolsStatus: []loadBalancerPoolStatusMetric{
						{
							ID:           "fake-load-balancer-pool-id-07",
							StatusDetail: buildExpectedLoadBalancerPoolStatusDetails("UP"),
							MembersStatus: []loadBalancerPoolMemberStatusMetric{
								{
									IPAddress:    fakeLoadbalancerPoolMemberIP,
									Port:         fakeLoadbalancerPoolMemberPort,
									StatusDetail: buildExpectedLoadBalancerPoolMemberStatusDetails("DOWN"),
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
					ID:           "fake-load-balancer-id-01",
					Name:         "fake-load-balancer-name-01",
					StatusDetail: buildExpectedLoadBalancerStatusDetails("UP"),
					PoolsStatus: []loadBalancerPoolStatusMetric{
						{
							ID:           "fake-load-balancer-pool-id-01",
							StatusDetail: buildExpectedLoadBalancerPoolStatusDetails("UP"),
							MembersStatus: []loadBalancerPoolMemberStatusMetric{
								{
									IPAddress:    fakeLoadbalancerPoolMemberIP,
									Port:         fakeLoadbalancerPoolMemberPort,
									StatusDetail: buildExpectedLoadBalancerPoolMemberStatusDetails("UP"),
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

func TestLoadBalancerCollector_GenerateLoadBalancerStatisticMetrics(t *testing.T) {
	testcases := []struct {
		description           string
		loadBalancerResponses []mockLoadBalancerResponse
		expectedMetrics       []loadBalancerStatisticMetric
	}{
		{
			description: "Should return correct statistic value depending on load balancer statistic",
			loadBalancerResponses: []mockLoadBalancerResponse{
				buildLoadBalancerStatisticResponse("01", nil),
			},
			expectedMetrics: []loadBalancerStatisticMetric{
				buildExpectedLoadBalancerStatistic("01"),
			},
		}, {
			description: "Should only return metric with valid response",
			loadBalancerResponses: []mockLoadBalancerResponse{
				buildLoadBalancerStatisticResponse("01", errors.New("unable to fetch load balancer statistic")),
				buildLoadBalancerStatisticResponse("02", nil),
			},
			expectedMetrics: []loadBalancerStatisticMetric{
				buildExpectedLoadBalancerStatistic("02"),
			},
		}, {
			description:           "Should return empty metrics when given empty load balancer",
			loadBalancerResponses: []mockLoadBalancerResponse{},
			expectedMetrics:       []loadBalancerStatisticMetric{},
		},
	}
	for _, tc := range testcases {
		client := &mockLoadBalancerClient{
			responses: tc.loadBalancerResponses,
		}
		loadBalancers := buildLoadBalancers(tc.loadBalancerResponses)
		logger := log.NewNopLogger()
		loadBalancerCollector := newLoadBalancerCollector(client, logger)
		loadBalancerStatisticMetrics := loadBalancerCollector.generateLoadBalancerStatisticMetrics(loadBalancers)
		assert.ElementsMatch(t, tc.expectedMetrics, loadBalancerStatisticMetrics, tc.description)
	}
}
