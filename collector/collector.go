package collector

import (
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	nsxt "github.com/vmware/go-vmware-nsxt"
)

const (
	namespace = "nsxt"
)

var (
	factories = make(map[string]func(client *nsxt.APIClient, logger log.Logger) prometheus.Collector)
)

func registerCollector(collector string, factory func(client *nsxt.APIClient, logger log.Logger) prometheus.Collector) {
	factories[collector] = factory
}

// nsxtCollector collects NSX-T stats from the given api server and exports them using
// the prometheus metrics package.
type nsxtCollector struct {
	collectors []prometheus.Collector
	client     *nsxt.APIClient
	logger     log.Logger
}

// NewNSXTCollector creates a new NSXTCollector.
func NewNSXTCollector(client *nsxt.APIClient, logger log.Logger) prometheus.Collector {
	var collectors []prometheus.Collector
	for key, factory := range factories {
		collector := factory(client, log.With(logger, "collector", key))
		collectors = append(collectors, collector)
	}
	return &nsxtCollector{
		collectors: collectors,
		client:     client,
		logger:     logger,
	}
}

// Describe implements the prometheus.Collector interface.
func (n *nsxtCollector) Describe(ch chan<- *prometheus.Desc) {
	wg := sync.WaitGroup{}
	wg.Add(len(n.collectors))
	for _, c := range n.collectors {
		go func(c prometheus.Collector) {
			c.Describe(ch)
			wg.Done()
		}(c)
	}
	wg.Wait()
}

// Collect implements the prometheus.Collector interface.
func (n *nsxtCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(n.collectors))
	for _, c := range n.collectors {
		go func(c prometheus.Collector) {
			c.Collect(ch)
			wg.Done()
		}(c)
	}
	wg.Wait()
}
