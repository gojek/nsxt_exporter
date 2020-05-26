package collector

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/go-vmware-nsxt/manager"
)

func fakeTransportNodeID(id string) string {
	return fmt.Sprintf("fake-transport-node-id-%s", id)
}

func fakeTransportNodeName(id string) string {
	return fmt.Sprintf("fake-transport-node-name-%s", id)
}

func fakeTransportZoneID(id string) string {
	return fmt.Sprintf("fake-transport-zone-id-%s", id)
}

func fakeEdgeClusterID(id string) string {
	return fmt.Sprintf("fake-edge-cluster-id-%s", id)
}

type transportNodeStatusResponse struct {
	ID     string
	Status string
	Error  error
}

type transportNodeClientMock struct {
	edgeClustersResponse         []manager.EdgeCluster
	edgeClustersError            error
	transportNodeStatusResponses []transportNodeStatusResponse
}

func (c *transportNodeClientMock) ListAllTransportNodes() ([]manager.TransportNode, error) {
	panic("implement me")
}

func (c *transportNodeClientMock) GetTransportNodeStatus(nodeID string) (manager.TransportNodeStatus, error) {
	for _, response := range c.transportNodeStatusResponses {
		if response.ID == nodeID {
			return manager.TransportNodeStatus{
				Status: response.Status,
			}, response.Error
		}
	}
	return manager.TransportNodeStatus{}, errors.New("transport node status not foud")
}

func (c *transportNodeClientMock) ListAllEdgeClusters() ([]manager.EdgeCluster, error) {
	return c.edgeClustersResponse, c.edgeClustersError
}

func buildExpectedTransportNodeStatusDetails(nonZeroStatus string) map[string]float64 {
	statusDetails := map[string]float64{
		"UP":       0.0,
		"DOWN":     0.0,
		"DEGRADED": 0.0,
		"UNKNOWN":  0.0,
	}
	statusDetails[nonZeroStatus] = 1.0
	return statusDetails
}

func TestTransportNodeCollector_GenerateEdgeClusterMemberships(t *testing.T) {
	testcases := []struct {
		description          string
		edgeClustersResponse []manager.EdgeCluster
		edgeClustersError    error
		expectedMembership   []edgeClusterMembership
		expectingError       bool
	}{
		{
			description: "Should return correct edge cluster memberships",
			edgeClustersResponse: []manager.EdgeCluster{
				{
					Id: fakeEdgeClusterID("01"),
					Members: []manager.EdgeClusterMember{
						{
							TransportNodeId: fakeTransportNodeID("01"),
							MemberIndex:     0,
						}, {
							TransportNodeId: fakeTransportNodeID("02"),
							MemberIndex:     1,
						},
					},
				}, {
					Id: fakeEdgeClusterID("02"),
					Members: []manager.EdgeClusterMember{
						{
							TransportNodeId: fakeTransportNodeID("01"),
							MemberIndex:     0,
						},
					},
				}, {
					Id:      fakeEdgeClusterID("03"),
					Members: []manager.EdgeClusterMember{},
				},
			},
			edgeClustersError: nil,
			expectedMembership: []edgeClusterMembership{
				{
					edgeClusterID:   fakeEdgeClusterID("01"),
					transportNodeID: fakeTransportNodeID("01"),
					edgeMemberIndex: "0",
				}, {
					edgeClusterID:   fakeEdgeClusterID("01"),
					transportNodeID: fakeTransportNodeID("02"),
					edgeMemberIndex: "1",
				}, {
					edgeClusterID:   fakeEdgeClusterID("02"),
					transportNodeID: fakeTransportNodeID("01"),
					edgeMemberIndex: "0",
				},
			},
			expectingError: false,
		}, {
			description:          "Should return empty memberships when got no edge cluster",
			edgeClustersResponse: []manager.EdgeCluster{},
			edgeClustersError:    nil,
			expectedMembership:   []edgeClusterMembership{},
			expectingError:       false,
		}, {
			description:          "Should error when failed to list edge cluster",
			edgeClustersResponse: []manager.EdgeCluster{},
			edgeClustersError:    errors.New("failed to list edge cluster"),
			expectedMembership:   nil,
			expectingError:       true,
		},
	}
	for _, tc := range testcases {
		client := &transportNodeClientMock{
			edgeClustersResponse: tc.edgeClustersResponse,
			edgeClustersError:    tc.edgeClustersError,
		}
		logger := log.NewNopLogger()
		collector := newTransportNodeCollector(client, logger)
		memberships, err := collector.generateEdgeClusterMemberships()
		if tc.expectingError {
			assert.Error(t, err, tc.description)
		}
		assert.ElementsMatch(t, memberships, tc.expectedMembership, tc.description)
	}
}

func TestTransportNodeCollector_GenerateTransportNodeMetrics(t *testing.T) {
	testcases := []struct {
		description                 string
		transportNodes              []manager.TransportNode
		edgeClusterMemberships      []edgeClusterMembership
		transportNodeStatusResponse []transportNodeStatusResponse
		expectedMetrics             []transportNodeMetric
	}{
		{
			description: "Should return correct transport node metrics",
			transportNodes: []manager.TransportNode{
				{
					Id:          fakeTransportNodeID("01"),
					DisplayName: fakeTransportNodeName("01"),
					TransportZoneEndpoints: []manager.TransportZoneEndPoint{
						{
							TransportZoneId: fakeTransportZoneID("01"),
						},
					},
				}, {
					Id:          fakeTransportNodeID("02"),
					DisplayName: fakeTransportNodeName("02"),
					TransportZoneEndpoints: []manager.TransportZoneEndPoint{
						{
							TransportZoneId: fakeTransportZoneID("02"),
						}, {
							TransportZoneId: fakeTransportZoneID("03"),
						},
					},
				}, {
					Id:          fakeTransportNodeID("03"),
					DisplayName: fakeTransportNodeName("03"),
					TransportZoneEndpoints: []manager.TransportZoneEndPoint{
						{
							TransportZoneId: fakeTransportZoneID("03"),
						}, {
							TransportZoneId: fakeTransportZoneID("04"),
						},
					},
				}, {
					Id:          fakeTransportNodeID("04"),
					DisplayName: fakeTransportNodeName("04"),
					TransportZoneEndpoints: []manager.TransportZoneEndPoint{
						{
							TransportZoneId: fakeTransportZoneID("04"),
						}, {
							TransportZoneId: fakeTransportZoneID("05"),
						},
					},
				},
			},
			edgeClusterMemberships: []edgeClusterMembership{
				{
					transportNodeID: fakeTransportNodeID("01"),
					edgeMemberIndex: "0",
					edgeClusterID:   fakeEdgeClusterID("01"),
				},
			},
			transportNodeStatusResponse: []transportNodeStatusResponse{
				{
					ID:     fakeTransportNodeID("01"),
					Status: "UP",
				}, {
					ID:     fakeTransportNodeID("02"),
					Status: "DOWN",
				}, {
					ID:     fakeTransportNodeID("03"),
					Status: "DEGRADED",
				}, {
					ID:     fakeTransportNodeID("04"),
					Status: "UNKNOWN",
				},
			},
			expectedMetrics: []transportNodeMetric{
				{
					ID:               fakeTransportNodeID("01"),
					Name:             fakeTransportNodeName("01"),
					StatusDetail:     buildExpectedTransportNodeStatusDetails("UP"),
					Type:             "edge",
					TransportZoneIDs: []string{fakeTransportZoneID("01")},
				}, {
					ID:               fakeTransportNodeID("02"),
					Name:             fakeTransportNodeName("02"),
					StatusDetail:     buildExpectedTransportNodeStatusDetails("DOWN"),
					Type:             "host",
					TransportZoneIDs: []string{fakeTransportZoneID("02"), fakeTransportZoneID("03")},
				}, {
					ID:               fakeTransportNodeID("03"),
					Name:             fakeTransportNodeName("03"),
					StatusDetail:     buildExpectedTransportNodeStatusDetails("DEGRADED"),
					Type:             "host",
					TransportZoneIDs: []string{fakeTransportZoneID("03"), fakeTransportZoneID("04")},
				}, {
					ID:               fakeTransportNodeID("04"),
					Name:             fakeTransportNodeName("04"),
					StatusDetail:     buildExpectedTransportNodeStatusDetails("UNKNOWN"),
					Type:             "host",
					TransportZoneIDs: []string{fakeTransportZoneID("04"), fakeTransportZoneID("05")},
				},
			},
		}, {
			description: "Should only return transport node metrics with valid status response",
			transportNodes: []manager.TransportNode{
				{
					Id:          fakeTransportNodeID("01"),
					DisplayName: fakeTransportNodeName("01"),
					TransportZoneEndpoints: []manager.TransportZoneEndPoint{
						{
							TransportZoneId: fakeTransportZoneID("01"),
						},
					},
				}, {
					Id:          fakeTransportNodeID("02"),
					DisplayName: fakeTransportNodeName("02"),
					TransportZoneEndpoints: []manager.TransportZoneEndPoint{
						{
							TransportZoneId: fakeTransportZoneID("02"),
						}, {
							TransportZoneId: fakeTransportZoneID("03"),
						},
					},
				},
			},
			edgeClusterMemberships: []edgeClusterMembership{
				{
					transportNodeID: fakeTransportNodeID("01"),
					edgeMemberIndex: "0",
					edgeClusterID:   fakeEdgeClusterID("01"),
				},
			},
			transportNodeStatusResponse: []transportNodeStatusResponse{
				{
					ID:     fakeTransportNodeID("01"),
					Status: "UP",
				}, {
					ID:     fakeTransportNodeID("02"),
					Status: "UP",
					Error:  errors.New("error getting transport node status"),
				},
			},
			expectedMetrics: []transportNodeMetric{
				{
					ID:               fakeTransportNodeID("01"),
					Name:             fakeTransportNodeName("01"),
					StatusDetail:     buildExpectedTransportNodeStatusDetails("UP"),
					Type:             "edge",
					TransportZoneIDs: []string{fakeTransportZoneID("01")},
				},
			},
		}, {
			description: "Should return transport node with empty type when given nil cluster memberships",
			transportNodes: []manager.TransportNode{
				{
					Id:          fakeTransportNodeID("01"),
					DisplayName: fakeTransportNodeName("01"),
					TransportZoneEndpoints: []manager.TransportZoneEndPoint{
						{
							TransportZoneId: fakeTransportZoneID("01"),
						},
					},
				}, {
					Id:          fakeTransportNodeID("02"),
					DisplayName: fakeTransportNodeName("02"),
					TransportZoneEndpoints: []manager.TransportZoneEndPoint{
						{
							TransportZoneId: fakeTransportZoneID("02"),
						}, {
							TransportZoneId: fakeTransportZoneID("03"),
						},
					},
				},
			},
			edgeClusterMemberships: nil,
			transportNodeStatusResponse: []transportNodeStatusResponse{
				{
					ID:     fakeTransportNodeID("01"),
					Status: "UP",
				}, {
					ID:     fakeTransportNodeID("02"),
					Status: "DOWN",
				},
			},
			expectedMetrics: []transportNodeMetric{
				{
					ID:               fakeTransportNodeID("01"),
					Name:             fakeTransportNodeName("01"),
					StatusDetail:     buildExpectedTransportNodeStatusDetails("UP"),
					Type:             "",
					TransportZoneIDs: []string{fakeTransportZoneID("01")},
				}, {
					ID:               fakeTransportNodeID("02"),
					Name:             fakeTransportNodeName("02"),
					StatusDetail:     buildExpectedTransportNodeStatusDetails("DOWN"),
					Type:             "",
					TransportZoneIDs: []string{fakeTransportZoneID("02"), fakeTransportZoneID("03")},
				},
			},
		}, {
			description:                 "Should return empty metrics when given empty transport node",
			transportNodes:              []manager.TransportNode{},
			edgeClusterMemberships:      []edgeClusterMembership{},
			transportNodeStatusResponse: []transportNodeStatusResponse{},
			expectedMetrics:             []transportNodeMetric{},
		},
	}

	for _, tc := range testcases {
		client := &transportNodeClientMock{
			transportNodeStatusResponses: tc.transportNodeStatusResponse,
		}
		logger := log.NewNopLogger()
		collector := newTransportNodeCollector(client, logger)
		metrics := collector.generateTransportNodeMetrics(tc.transportNodes, tc.edgeClusterMemberships)
		assert.ElementsMatch(t, tc.expectedMetrics, metrics, tc.description)
	}
}
