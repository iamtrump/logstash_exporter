package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/sequra/logstash_exporter/collector"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/prometheus/exporter-toolkit/web"
)

var (
	scrapeDurations = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: collector.Namespace,
			Subsystem: "exporter",
			Name:      "scrape_duration_seconds",
			Help:      "logstash_exporter: Duration of a scrape job.",
		},
		[]string{"collector", "result"},
	)
	logger              log.Logger
	logstashEndpoint    *string
	exporterBindAddress *string
	configFile          *string
)

// LogstashCollector collector type
type LogstashCollector struct {
	collectors map[string]collector.Collector
}

// NewLogstashCollector register a logstash collector
func NewLogstashCollector(logstashEndpoint string) (*LogstashCollector, error) {
	nodeStatsCollector, err := collector.NewNodeStatsCollector(logstashEndpoint, logger)
	if err != nil {
		level.Error(logger).Log("msg", "Cannot register a new collector", "err", err)
		os.Exit(1)
	}

	nodeInfoCollector, err := collector.NewNodeInfoCollector(logstashEndpoint, logger)
	if err != nil {
		level.Error(logger).Log("msg", "Cannot register a new collector", "err", err)
		os.Exit(1)
	}

	return &LogstashCollector{
		collectors: map[string]collector.Collector{
			"node": nodeStatsCollector,
			"info": nodeInfoCollector,
		},
	}, nil
}

func listen() {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/metrics", http.StatusMovedPermanently)
	})

	level.Info(logger).Log("msg", "Starting server", "bind_address", exporterBindAddress)
	server := &http.Server{Addr: *exporterBindAddress}
	if err := web.ListenAndServe(server, *configFile, logger); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
}

// Describe logstash metrics
func (coll LogstashCollector) Describe(ch chan<- *prometheus.Desc) {
	scrapeDurations.Describe(ch)
}

// Collect logstash metrics
func (coll LogstashCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(coll.collectors))
	for name, c := range coll.collectors {
		go func(name string, c collector.Collector) {
			execute(name, c, ch)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
	scrapeDurations.Collect(ch)
}

func execute(name string, c collector.Collector, ch chan<- prometheus.Metric) {
	begin := time.Now()
	err := c.Collect(ch)
	duration := time.Since(begin)
	var result string

	if err != nil {
		level.Debug(logger).Log("msg", "Collector failed", "name", name, "duration", duration.Seconds(), "err", err)
		result = "error"
	} else {
		level.Debug(logger).Log("msg", "Collector succeeded", "name", name, "duration", duration.Seconds())
		result = "success"
	}
	scrapeDurations.WithLabelValues(name, result).Observe(duration.Seconds())
}

func init() {
	prometheus.MustRegister(version.NewCollector("logstash_exporter"))

	logstashEndpoint = kingpin.Flag("logstash.endpoint", "The protocol, host and port on which logstash metrics API listens").Default("http://localhost:9600").String()
	exporterBindAddress = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9198").String()
	configFile = kingpin.Flag("web.config", "[EXPERIMENTAL] Path to config yaml file that can enable TLS or authentication.").Default("").String()

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("logstash_exporter"))
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger = promlog.New(promlogConfig)
}

func main() {
	logstashCollector, err := NewLogstashCollector(*logstashEndpoint)
	if err != nil {
		level.Error(logger).Log("msg", "Cannot register a new Logstash Collector", "err", err)
		os.Exit(1)
	}

	prometheus.MustRegister(logstashCollector)

	level.Info(logger).Log("msg", "Starting Logstash exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())
	listen()
}
