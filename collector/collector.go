package collector

import (
	"sync"

	"nsxt_exporter/client"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "nsxt"
)

var (
	factories = make(map[string]func(client client.NSXTClient, logger log.Logger) prometheus.Collector)
)

func registerCollector(collector string, factory func(client client.NSXTClient, logger log.Logger) prometheus.Collector) {
	factories[collector] = factory
}

// nsxtCollector collects NSX-T stats from the given api server and exports them using
// the prometheus metrics package.
type nsxtCollector struct {
	collectors []prometheus.Collector
}

// NewNSXTCollector creates a new NSXTCollector.
func NewNSXTCollector(client client.NSXTClient, logger log.Logger) prometheus.Collector {
	var collectors []prometheus.Collector
	for key, factory := range factories {
		collector := factory(client, log.With(logger, "collector", key))
		collectors = append(collectors, collector)
	}
	return &nsxtCollector{
		collectors: collectors,
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
