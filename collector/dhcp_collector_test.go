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
)

type mockDHCPClient struct {
	responses []mockDHCPResponse
}

type mockDHCPResponse struct {
	ID          string
	DisplayName string
	Status      string
	Error       error
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

func buildDHCPResponse(id string, status string, err error) mockDHCPResponse {
	return mockDHCPResponse{
		ID:          fakeDHCPServerID + "-" + id,
		DisplayName: fakeDHCPServerDisplayName + "-" + id,
		Status:      status,
		Error:       err,
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
				buildDHCPResponse("01", "UP", nil),
				buildDHCPResponse("02", "DOWN", nil),
				buildDHCPResponse("03", "ERROR", nil),
				buildDHCPResponse("04", "NO_STANDBY", nil),
				buildDHCPResponse("05", "Up", nil),
				buildDHCPResponse("06", "dOwN", nil),
			},
			expectedMetrics: []dhcpStatusMetric{
				{
					ID:     "fake-dhcp-server-id-01",
					Name:   "fake-dhcp-server-name-01",
					Status: 1.0,
				},
				{
					ID:     "fake-dhcp-server-id-02",
					Name:   "fake-dhcp-server-name-02",
					Status: 0.0,
				},
				{
					ID:     "fake-dhcp-server-id-03",
					Name:   "fake-dhcp-server-name-03",
					Status: 0.0,
				},
				{
					ID:     "fake-dhcp-server-id-04",
					Name:   "fake-dhcp-server-name-04",
					Status: 0.0,
				},
				{
					ID:     "fake-dhcp-server-id-05",
					Name:   "fake-dhcp-server-name-05",
					Status: 1.0,
				},
				{
					ID:     "fake-dhcp-server-id-06",
					Name:   "fake-dhcp-server-name-06",
					Status: 0.0,
				},
			},
		},
		{
			description: "Should only return dhcp server with valid response",
			dhcpResponses: []mockDHCPResponse{
				buildDHCPResponse("01", "UP", nil),
				buildDHCPResponse("02", "UP", errors.New("error get dhcp")),
			},
			expectedMetrics: []dhcpStatusMetric{
				{
					ID:     "fake-dhcp-server-id-01",
					Name:   "fake-dhcp-server-name-01",
					Status: 1.0,
				},
			},
		},
		{
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
