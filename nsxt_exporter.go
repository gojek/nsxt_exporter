package main

import (
	"net/http"
	"os"

	"nsxt_exporter/client"
	"nsxt_exporter/collector"

	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9732").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		opts          = client.NSXTOpts{}
	)
	kingpin.Flag("nsxt.host", "URI of NSX-T manager.").Default("localhost").StringVar(&opts.Host)
	kingpin.Flag("nsxt.username", "The username to connect to the NSX-T manager as.").StringVar(&opts.Username)
	kingpin.Flag("nsxt.password", "The password for the NSX-T manager user.").StringVar(&opts.Password)
	kingpin.Flag("nsxt.insecure", "Disable TLS host verification.").Default("true").BoolVar(&opts.Insecure)

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting nsxt_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "context", version.BuildContext())

	nsxtClient, err := client.NewNSXTClient(opts)
	if err != nil {
		level.Error(logger).Log("msg", "Error creating nsx-t client", "err", err)
		os.Exit(1)
	}

	collector := collector.NewNSXTCollector(nsxtClient, logger)
	prometheus.MustRegister(collector)
	prometheus.MustRegister(version.NewCollector("nsxt_exporter"))

	level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>NSX-T Exporter</title></head>
		<body>
		<h1>NSX-T Exporter</h1>
		<p><a href="` + *metricsPath + `"></p>
		</body>
		</html>`))
	})
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
