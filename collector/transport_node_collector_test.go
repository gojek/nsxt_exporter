package collector

import (
	"errors"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/go-vmware-nsxt/manager"
)

type transportNodeClientMock struct {
	edgeClustersResponse []manager.EdgeCluster
	edgeClustersError    error
}

func (c *transportNodeClientMock) ListAllTransportNodes() ([]manager.TransportNode, error) {
	panic("implement me")
}

func (c *transportNodeClientMock) GetTransportNodeStatus(nodeID string) (manager.TransportNodeStatus, error) {
	panic("implement me")
}

func (c *transportNodeClientMock) ListAllEdgeClusters() ([]manager.EdgeCluster, error) {
	return c.edgeClustersResponse, c.edgeClustersError
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
					Id: "fake-edge-cluster-id-01",
					Members: []manager.EdgeClusterMember{
						{
							TransportNodeId: "fake-transport-node-id-01",
							MemberIndex:     0,
						}, {
							TransportNodeId: "fake-transport-node-id-02",
							MemberIndex:     1,
						},
					},
				}, {
					Id: "fake-edge-cluster-id-02",
					Members: []manager.EdgeClusterMember{
						{
							TransportNodeId: "fake-transport-node-id-01",
							MemberIndex:     0,
						},
					},
				}, {
					Id:      "fake-edge-cluster-id-03",
					Members: []manager.EdgeClusterMember{},
				},
			},
			edgeClustersError: nil,
			expectedMembership: []edgeClusterMembership{
				{
					edgeClusterID:   "fake-edge-cluster-id-01",
					transportNodeID: "fake-transport-node-id-01",
					edgeMemberIndex: "0",
				}, {
					edgeClusterID:   "fake-edge-cluster-id-01",
					transportNodeID: "fake-transport-node-id-02",
					edgeMemberIndex: "1",
				}, {
					edgeClusterID:   "fake-edge-cluster-id-02",
					transportNodeID: "fake-transport-node-id-01",
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
