package collector

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/go-vmware-nsxt/manager"
)

const (
	fakeFirewallSectionID        = "fake-firewall-section-id"
	fakeFirewallRuleID           = "fake-firewall-rule-id"
	fakeFirewallRuleName         = "fake-firewall-rule-name"
	fakeFirewallRuleTotalPackets = 1
	fakeFirewallRuleTotalBytes   = 1
)

type mockFirewallClient struct {
	responses             []mockFirewallResponse
	firewallRuleListError error
}

type mockFirewallResponse struct {
	Section    manager.FirewallSection
	Rules      []manager.FirewallRule
	Statistics []manager.FirewallStats
	Error      error
}

func (c *mockFirewallClient) ListAllFirewallSections() ([]manager.FirewallSection, error) {
	panic("unused function. Only used to satisfy FirewallClient interface")
}

func (c *mockFirewallClient) GetAllFirewallRules(sectionID string) ([]manager.FirewallRule, error) {
	if c.firewallRuleListError != nil {
		return nil, c.firewallRuleListError
	}
	var firewallRules []manager.FirewallRule
	for _, res := range c.responses {
		if res.Section.Id != sectionID {
			continue
		}
		for _, rule := range res.Rules {
			firewallRule := manager.FirewallRule{
				Id:          rule.Id,
				DisplayName: rule.DisplayName,
			}
			firewallRules = append(firewallRules, firewallRule)
		}
	}
	return firewallRules, nil
}

func (c *mockFirewallClient) GetFirewallStats(sectionID string, ruleID string) (manager.FirewallStats, error) {
	for _, res := range c.responses {
		if res.Section.Id != sectionID {
			continue
		}
		for _, rule := range res.Rules {
			if rule.Id != ruleID {
				continue
			}
			return manager.FirewallStats{
				RuleId:      ruleID,
				PacketCount: fakeFirewallRuleTotalPackets,
				ByteCount:   fakeFirewallRuleTotalBytes,
			}, res.Error
		}
	}
	return manager.FirewallStats{}, errors.New("error firewall rule not found")
}

func buildFirewallResponse(sectionID string, ruleIDs []string, err error) mockFirewallResponse {
	var firewallRules []manager.FirewallRule
	var firewallStatistics []manager.FirewallStats
	for _, ruleID := range ruleIDs {
		firewallRules = append(firewallRules, manager.FirewallRule{
			Id:          fmt.Sprintf("%s-%s", fakeFirewallRuleID, ruleID),
			DisplayName: fmt.Sprintf("%s-%s", fakeFirewallRuleName, ruleID),
		})
		firewallStatistics = append(firewallStatistics, manager.FirewallStats{
			RuleId:      fmt.Sprintf("%s-%s", fakeFirewallRuleID, ruleID),
			PacketCount: fakeFirewallRuleTotalPackets,
			ByteCount:   fakeFirewallRuleTotalBytes,
		})
	}
	return mockFirewallResponse{
		Section: manager.FirewallSection{
			Id: fmt.Sprintf("%s-%s", fakeFirewallSectionID, sectionID),
		},
		Rules:      firewallRules,
		Statistics: firewallStatistics,
		Error:      err,
	}
}

func buildFirewallSections(firewallResponses []mockFirewallResponse) []manager.FirewallSection {
	var firewallSections []manager.FirewallSection
	for _, res := range firewallResponses {
		firewallSections = append(firewallSections, res.Section)
	}
	return firewallSections
}

func TestFirewallCollector_GenerateFirewallStatisticMetrics(t *testing.T) {
	testcases := []struct {
		description           string
		firewallRuleListError error
		firewallResponses     []mockFirewallResponse
		expectedMetrics       []firewallStatisticMetric
	}{
		{
			description: "Should return correct statistics metrics",
			firewallResponses: []mockFirewallResponse{
				buildFirewallResponse("1", []string{"1", "2"}, nil),
				buildFirewallResponse("2", []string{"3"}, nil),
			},
			expectedMetrics: []firewallStatisticMetric{
				{
					SectionID:    fmt.Sprintf("%s-1", fakeFirewallSectionID),
					RuleID:       fmt.Sprintf("%s-1", fakeFirewallRuleID),
					RuleName:     fmt.Sprintf("%s-1", fakeFirewallRuleName),
					TotalPackets: fakeFirewallRuleTotalPackets,
					TotalBytes:   fakeFirewallRuleTotalBytes,
				},
				{
					SectionID:    fmt.Sprintf("%s-1", fakeFirewallSectionID),
					RuleID:       fmt.Sprintf("%s-2", fakeFirewallRuleID),
					RuleName:     fmt.Sprintf("%s-2", fakeFirewallRuleName),
					TotalPackets: fakeFirewallRuleTotalPackets,
					TotalBytes:   fakeFirewallRuleTotalBytes,
				},
				{
					SectionID:    fmt.Sprintf("%s-2", fakeFirewallSectionID),
					RuleID:       fmt.Sprintf("%s-3", fakeFirewallRuleID),
					RuleName:     fmt.Sprintf("%s-3", fakeFirewallRuleName),
					TotalPackets: fakeFirewallRuleTotalPackets,
					TotalBytes:   fakeFirewallRuleTotalBytes,
				},
			},
		},
		{
			description:       "Should return empty metrics when given empty firewall sections",
			firewallResponses: []mockFirewallResponse{},
			expectedMetrics:   []firewallStatisticMetric{},
		},
		{
			description:           "Should return empty metrics when error listing nat rules",
			firewallRuleListError: errors.New("error list firewall rules"),
			firewallResponses:     []mockFirewallResponse{},
			expectedMetrics:       []firewallStatisticMetric{},
		},
	}
	for _, tc := range testcases {
		mockFirewallClient := &mockFirewallClient{
			responses: tc.firewallResponses,
		}
		logger := log.NewNopLogger()
		firewallCollector := newFirewallCollector(mockFirewallClient, logger)
		firewallSections := buildFirewallSections(tc.firewallResponses)
		metrics := firewallCollector.generateFirewallStatisticMetrics(firewallSections)
		assert.ElementsMatch(t, tc.expectedMetrics, metrics, tc.description)
	}
}
