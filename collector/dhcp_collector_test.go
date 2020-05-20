package collector

import (
	"errors"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/go-vmware-nsxt/manager"
)

const (
	fakeDHCPServerID          = "fake-dhcp-server-id"
	fakeDHCPServerDisplayName = "fake-dhcp-server-name"
	fakeDHCPPoolID            = "fake-dhcp-pool-id"
	fakeDHCPStatisticValue    = 785
)

type mockDHCPClient struct {
	responses []mockDHCPResponse
}

type mockDHCPResponse struct {
	ID          string
	DisplayName string
	Status      string
	Error       error
	Statistics  manager.DhcpStatistics
}

func (c *mockDHCPClient) ListAllDHCPServers() ([]manager.LogicalDhcpServer, error) {
	panic("unused function. Only used to satisfy DHCPClient interface")
}

func (c *mockDHCPClient) GetDhcpStatus(dhcpID string, localVarOptionals map[string]interface{}) (manager.DhcpServerStatus, error) {
	for _, res := range c.responses {
		if res.ID == dhcpID {
			return manager.DhcpServerStatus{
				ServiceStatus: res.Status,
			}, res.Error
		}
	}
	return manager.DhcpServerStatus{}, errors.New("error")
}

func (c *mockDHCPClient) GetDHCPStatistic(dhcpID string) (manager.DhcpStatistics, error) {
	for _, res := range c.responses {
		if res.ID == dhcpID {
			return res.Statistics, res.Error
		}
	}
	return manager.DhcpStatistics{}, errors.New("dhcp not found")
}

func buildDHCPStatusResponse(id string, status string, err error) mockDHCPResponse {
	return mockDHCPResponse{
		ID:          fakeDHCPServerID + "-" + id,
		DisplayName: fakeDHCPServerDisplayName + "-" + id,
		Status:      status,
		Error:       err,
	}
}

func buildDHCPStatisticResponse(id string, err error) mockDHCPResponse {
	return mockDHCPResponse{
		ID:          fakeDHCPServerID + "-" + id,
		DisplayName: fakeDHCPServerDisplayName + "-" + id,
		Error:       err,
		Statistics: manager.DhcpStatistics{
			Acks:      fakeDHCPStatisticValue,
			Declines:  fakeDHCPStatisticValue,
			Discovers: fakeDHCPStatisticValue,
			Errors:    fakeDHCPStatisticValue,
			Informs:   fakeDHCPStatisticValue,
			IpPoolStats: []manager.DhcpIpPoolUsage{
				{
					AllocatedNumber:     fakeDHCPStatisticValue,
					AllocatedPercentage: fakeDHCPStatisticValue,
					DhcpIpPoolId:        fakeDHCPPoolID + "-" + id,
					PoolSize:            fakeDHCPStatisticValue,
				},
			},
			Nacks:    fakeDHCPStatisticValue,
			Offers:   fakeDHCPStatisticValue,
			Releases: fakeDHCPStatisticValue,
			Requests: fakeDHCPStatisticValue,
		},
	}
}

func TestDHCPCollector_GenerateDHCPStatisticMetrics(t *testing.T) {
	testcases := []struct {
		description     string
		dhcpResponses   []mockDHCPResponse
		expectedMetrics []dhcpStatisticMetric
	}{
		{
			description: "Should return correct status value depending on dhcp server state",
			dhcpResponses: []mockDHCPResponse{
				buildDHCPStatisticResponse("01", nil),
				buildDHCPStatisticResponse("02", nil),
			},
			expectedMetrics: []dhcpStatisticMetric{
				{
					ID:   "fake-dhcp-server-id-01",
					Name: "fake-dhcp-server-name-01",
					Statistic: manager.DhcpStatistics{
						Acks:      fakeDHCPStatisticValue,
						Declines:  fakeDHCPStatisticValue,
						Discovers: fakeDHCPStatisticValue,
						Errors:    fakeDHCPStatisticValue,
						Informs:   fakeDHCPStatisticValue,
						IpPoolStats: []manager.DhcpIpPoolUsage{
							{
								AllocatedNumber:     fakeDHCPStatisticValue,
								AllocatedPercentage: fakeDHCPStatisticValue,
								DhcpIpPoolId:        "fake-dhcp-pool-id-01",
								PoolSize:            fakeDHCPStatisticValue,
							},
						},
						Nacks:    fakeDHCPStatisticValue,
						Offers:   fakeDHCPStatisticValue,
						Releases: fakeDHCPStatisticValue,
						Requests: fakeDHCPStatisticValue,
					},
				}, {
					ID:   "fake-dhcp-server-id-02",
					Name: "fake-dhcp-server-name-02",
					Statistic: manager.DhcpStatistics{
						Acks:      fakeDHCPStatisticValue,
						Declines:  fakeDHCPStatisticValue,
						Discovers: fakeDHCPStatisticValue,
						Errors:    fakeDHCPStatisticValue,
						Informs:   fakeDHCPStatisticValue,
						IpPoolStats: []manager.DhcpIpPoolUsage{
							{
								AllocatedNumber:     fakeDHCPStatisticValue,
								AllocatedPercentage: fakeDHCPStatisticValue,
								DhcpIpPoolId:        "fake-dhcp-pool-id-02",
								PoolSize:            fakeDHCPStatisticValue,
							},
						},
						Nacks:    fakeDHCPStatisticValue,
						Offers:   fakeDHCPStatisticValue,
						Releases: fakeDHCPStatisticValue,
						Requests: fakeDHCPStatisticValue,
					},
				},
			},
		}, {
			description: "Should only correct status value with valid response",
			dhcpResponses: []mockDHCPResponse{
				buildDHCPStatisticResponse("01", nil),
				buildDHCPStatisticResponse("02", errors.New("unable to get dhcp statistics")),
			},
			expectedMetrics: []dhcpStatisticMetric{
				{
					ID:   "fake-dhcp-server-id-01",
					Name: "fake-dhcp-server-name-01",
					Statistic: manager.DhcpStatistics{
						Acks:      fakeDHCPStatisticValue,
						Declines:  fakeDHCPStatisticValue,
						Discovers: fakeDHCPStatisticValue,
						Errors:    fakeDHCPStatisticValue,
						Informs:   fakeDHCPStatisticValue,
						IpPoolStats: []manager.DhcpIpPoolUsage{
							{
								AllocatedNumber:     fakeDHCPStatisticValue,
								AllocatedPercentage: fakeDHCPStatisticValue,
								DhcpIpPoolId:        "fake-dhcp-pool-id-01",
								PoolSize:            fakeDHCPStatisticValue,
							},
						},
						Nacks:    fakeDHCPStatisticValue,
						Offers:   fakeDHCPStatisticValue,
						Releases: fakeDHCPStatisticValue,
						Requests: fakeDHCPStatisticValue,
					},
				},
			},
		}, {
			description:     "Should return empty metrics when given empty dhcp server",
			dhcpResponses:   []mockDHCPResponse{},
			expectedMetrics: []dhcpStatisticMetric{},
		},
	}

	for _, tc := range testcases {
		mockDHCPClient := &mockDHCPClient{
			responses: tc.dhcpResponses,
		}
		var dhcpServers []manager.LogicalDhcpServer
		for _, response := range tc.dhcpResponses {
			dhcpServer := manager.LogicalDhcpServer{
				Id:          response.ID,
				DisplayName: response.DisplayName,
			}
			dhcpServers = append(dhcpServers, dhcpServer)
		}
		logger := log.NewNopLogger()
		dhcpCollector := newDHCPCollector(mockDHCPClient, logger)
		dhcpMetrics := dhcpCollector.generateDHCPStatisticMetrics(dhcpServers)
		assert.ElementsMatch(t, tc.expectedMetrics, dhcpMetrics, tc.description)
	}
}

func TestDHCPCollector_GenerateDHCPStatusMetrics(t *testing.T) {
	testcases := []struct {
		description     string
		dhcpResponses   []mockDHCPResponse
		expectedMetrics []dhcpStatusMetric
	}{
		{
			description: "Should return correct status value depending on dhcp server state",
			dhcpResponses: []mockDHCPResponse{
				buildDHCPStatusResponse("01", "UP", nil),
				buildDHCPStatusResponse("02", "DOWN", nil),
				buildDHCPStatusResponse("03", "ERROR", nil),
				buildDHCPStatusResponse("04", "NO_STANDBY", nil),
				buildDHCPStatusResponse("05", "Up", nil),
				buildDHCPStatusResponse("06", "dOwN", nil),
			},
			expectedMetrics: []dhcpStatusMetric{
				{
					ID:         "fake-dhcp-server-id-01",
					Name:       "fake-dhcp-server-name-01",
					Status:     1.0,
					StatusEnum: "UP",
				}, {
					ID:         "fake-dhcp-server-id-02",
					Name:       "fake-dhcp-server-name-02",
					Status:     0.0,
					StatusEnum: "DOWN",
				}, {
					ID:         "fake-dhcp-server-id-03",
					Name:       "fake-dhcp-server-name-03",
					Status:     0.0,
					StatusEnum: "ERROR",
				}, {
					ID:         "fake-dhcp-server-id-04",
					Name:       "fake-dhcp-server-name-04",
					Status:     0.0,
					StatusEnum: "NO_STANDBY",
				}, {
					ID:         "fake-dhcp-server-id-05",
					Name:       "fake-dhcp-server-name-05",
					Status:     1.0,
					StatusEnum: "UP",
				}, {
					ID:         "fake-dhcp-server-id-06",
					Name:       "fake-dhcp-server-name-06",
					Status:     0.0,
					StatusEnum: "DOWN",
				},
			},
		}, {
			description: "Should only return dhcp server with valid response",
			dhcpResponses: []mockDHCPResponse{
				buildDHCPStatusResponse("01", "UP", nil),
				buildDHCPStatusResponse("02", "UP", errors.New("error get dhcp")),
			},
			expectedMetrics: []dhcpStatusMetric{
				{
					ID:         "fake-dhcp-server-id-01",
					Name:       "fake-dhcp-server-name-01",
					Status:     1.0,
					StatusEnum: "UP",
				},
			},
		}, {
			dhcpResponses:   []mockDHCPResponse{},
			expectedMetrics: []dhcpStatusMetric{},
		},
	}
	for _, tc := range testcases {
		mockDHCPClient := &mockDHCPClient{
			responses: tc.dhcpResponses,
		}
		var dhcpServers []manager.LogicalDhcpServer
		for _, response := range tc.dhcpResponses {
			dhcpServer := manager.LogicalDhcpServer{
				Id:          response.ID,
				DisplayName: response.DisplayName,
			}
			dhcpServers = append(dhcpServers, dhcpServer)
		}
		logger := log.NewNopLogger()
		dhcpCollector := newDHCPCollector(mockDHCPClient, logger)
		dhcpMetrics := dhcpCollector.generateDHCPStatusMetrics(dhcpServers)
		assert.ElementsMatch(t, tc.expectedMetrics, dhcpMetrics, tc.description)
	}
}
