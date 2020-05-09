package client

import (
	nsxt "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/manager"
)

type NSXTClient interface {
	GetLogicalPortStatusSummary(localVarOptionals map[string]interface{}) (manager.LogicalPortStatusSummary, error)
	ListLogicalPorts(localVarOptionals map[string]interface{}) (manager.LogicalPortListResult, error)
	GetLogicalPortOperationalStatus(lportId string, localVarOptionals map[string]interface{}) (manager.LogicalPortOperationalStatus, error)
}

type NSXTOpts struct {
	Host     string
	Username string
	Password string
	Insecure bool
}

type nsxtClient struct {
	apiClient *nsxt.APIClient
}

func NewNSXTClient(opts NSXTOpts) (*nsxtClient, error) {
	cfg := nsxt.Configuration{
		BasePath:           "/api/v1",
		Host:               opts.Host,
		Scheme:             "https",
		UserAgent:          "nsxt_exporter/1.0",
		ClientAuthCertFile: "",
		RemoteAuth:         false,
		UserName:           opts.Username,
		Password:           opts.Password,
		Insecure:           opts.Insecure,
	}
	apiClient, err := nsxt.NewAPIClient(&cfg)
	if err != nil {
		return nil, err
	}
	return &nsxtClient{
		apiClient: apiClient,
	}, nil
}

func (c *nsxtClient) GetLogicalPortStatusSummary(localVarOptionals map[string]interface{}) (manager.LogicalPortStatusSummary, error) {
	lportStatus, _, err := c.apiClient.LogicalSwitchingApi.GetLogicalPortStatusSummary(c.apiClient.Context, localVarOptionals)
	return lportStatus, err
}

func (c *nsxtClient) ListLogicalPorts(localVarOptionals map[string]interface{}) (manager.LogicalPortListResult, error) {
	lportsResult, _, err := c.apiClient.LogicalSwitchingApi.ListLogicalPorts(c.apiClient.Context, localVarOptionals)
	return lportsResult, err
}

func (c *nsxtClient) GetLogicalPortOperationalStatus(lportId string, localVarOptionals map[string]interface{}) (manager.LogicalPortOperationalStatus, error) {
	lportStatus, _, err := c.apiClient.LogicalSwitchingApi.GetLogicalPortOperationalStatus(c.apiClient.Context, lportId, localVarOptionals)
	return lportStatus, err
}
