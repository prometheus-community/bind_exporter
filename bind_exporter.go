package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const (
	namespace = "bind"
	resolver  = "resolver"
)

var (
	up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Is the Bind instance up and running?",
		nil, nil,
	)
	// Total incoming queries, per type (opcode/rrtype) and code (opcode number, RR)
	incomingQueries = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "incoming_queries_total"),
		"Number of incoming DNS queries.",
		[]string{"type", "code"}, nil,
	)
	resolverCache = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, resolver, "cache_rrsets"),
		"Number of RRSets in Cache database.",
		[]string{"view", "type"}, nil,
	)
	resolverQueries = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, resolver, "queries_total"),
		"Number of outgoing DNS queries.",
		[]string{"view", "type"}, nil,
	)
	resolverQueryDuration = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, resolver, "query_duration_seconds"),
		"Resolver query round-trip time in seconds.",
		[]string{"view"}, nil,
	)
	resolverQueryErrors = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, resolver, "query_errors_total"),
		"Number of resolver queries failed.",
		[]string{"view", "error"}, nil,
	)
	resolverResponseErrors = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, resolver, "response_errors_total"),
		"Number of resolver reponse errors received.",
		[]string{"view", "error"}, nil,
	)
	resolverDNSSECSucess = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, resolver, "dnssec_validation_success_total"),
		"Number of DNSSEC validation attempts succeeded.",
		[]string{"view", "result"}, nil,
	)
	resolverMetricStats = map[string]*prometheus.Desc{
		"Lame": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resolver, "response_lame_total"),
			"Number of lame delegation responses received.",
			[]string{"view"}, nil,
		),
		"EDNS0Fail": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resolver, "query_edns0_errors_total"),
			"Number of EDNS(0) query errors.",
			[]string{"view"}, nil,
		),
		"Mismatch": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resolver, "response_mismatch_total"),
			"Number of mismatch responses received.",
			[]string{"view"}, nil,
		),
		"Retry": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resolver, "query_retries_total"),
			"Number of resolver query retries.",
			[]string{"view"}, nil,
		),
		"Truncated": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resolver, "response_truncated_total"),
			"Number of truncated responses received.",
			[]string{"view"}, nil,
		),
		"ValFail": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resolver, "dnssec_validation_errors_total"),
			"Number of DNSSEC validation attempt errors.",
			[]string{"view"}, nil,
		),
	}
	resolverLabelStats = map[string]*prometheus.Desc{
		"QueryAbort":    resolverQueryErrors,
		"QuerySockFail": resolverQueryErrors,
		"QueryTimeout":  resolverQueryErrors,
		"NXDOMAIN":      resolverResponseErrors,
		"SERVFAIL":      resolverResponseErrors,
		"FORMERR":       resolverResponseErrors,
		"OtherError":    resolverResponseErrors,
		"ValOk":         resolverDNSSECSucess,
		"ValNegOk":      resolverDNSSECSucess,
	}
	serverQueryErrors = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "query_errors_total"),
		"Number of query failures.",
		[]string{"error"}, nil,
	)
	serverReponses = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "responses_total"),
		"Number of responses sent.",
		[]string{"result"}, nil,
	)
	serverLabelStats = map[string]*prometheus.Desc{
		"QryDuplicate": serverQueryErrors,
		"QryDropped":   serverQueryErrors,
		"QryFailure":   serverQueryErrors,
		"QrySuccess":   serverReponses,
		"QryReferral":  serverReponses,
		"QryNxrrset":   serverReponses,
		"QrySERVFAIL":  serverReponses,
		"QryFORMERR":   serverReponses,
		"QryNXDOMAIN":  serverReponses,
	}
	tasksRunning = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "tasks_running"),
		"Number of running tasks.",
		nil, nil,
	)
	workerThreads = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "worker_threads"),
		"Total number of available worker threads.",
		nil, nil,
	)
)

// Exporter collects Binds stats from the given server and exports
// them using the prometheus metrics package.
type Exporter struct {
	URI    string
	client *http.Client
}

// NewExporter returns an initialized Exporter.
func NewExporter(uri string, timeout time.Duration) *Exporter {
	return &Exporter{
		URI: uri,
		client: &http.Client{
			Transport: &http.Transport{
				Dial: func(netw, addr string) (net.Conn, error) {
					c, err := net.DialTimeout(netw, addr, timeout)
					if err != nil {
						return nil, err
					}
					if err := c.SetDeadline(time.Now().Add(timeout)); err != nil {
						return nil, err
					}
					return c, nil
				},
			},
		},
	}
}

// Describe describes all the metrics ever exported by the bind
// exporter. It implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- incomingQueries
	ch <- resolverDNSSECSucess
	ch <- resolverQueries
	ch <- resolverQueryDuration
	ch <- resolverQueryErrors
	ch <- resolverResponseErrors
	for _, desc := range resolverMetricStats {
		ch <- desc
	}
	ch <- serverReponses
	ch <- tasksRunning
	ch <- workerThreads
}

// Collect fetches the stats from configured bind location and
// delivers them as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	var status float64
	defer func() {
		ch <- prometheus.MustNewConstMetric(up, prometheus.GaugeValue, status)
	}()

	resp, err := e.client.Get(e.URI)
	if err != nil {
		log.Error("Error while querying Bind: ", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read XML response body: ", err)
		return
	}

	status = 1

	stats, err := unmarshal(body)
	if err != nil {
		log.Error("Failed to unmarshal XML response: ", err)
		return
	}

	for _, s := range stats.Server.Counters {
		if s.Type != "opcode" && s.Type != "qtype" {
			continue
		}
		for _, c := range s.Counter {
			ch <- prometheus.MustNewConstMetric(
				incomingQueries, prometheus.CounterValue, float64(c.Counter), s.Type, c.Name,
			)
		}
	}

	// stats.Views.Zones has zone stats

	for _, v := range stats.Views {
		for _, s := range v.Cache {
			ch <- prometheus.MustNewConstMetric(
				resolverCache, prometheus.GaugeValue, float64(s.Counter), v.Name, s.Name,
			)
		}

		for _, s := range v.Rdtype {
			ch <- prometheus.MustNewConstMetric(
				resolverQueries, prometheus.CounterValue, float64(s.Counter), v.Name, s.Name,
			)
		}

		for _, s := range v.Resstat {
			if desc, ok := resolverMetricStats[s.Name]; ok {
				ch <- prometheus.MustNewConstMetric(
					desc, prometheus.CounterValue, float64(s.Counter), v.Name,
				)
			}
			if desc, ok := resolverLabelStats[s.Name]; ok {
				ch <- prometheus.MustNewConstMetric(
					desc, prometheus.CounterValue, float64(s.Counter), v.Name, s.Name,
				)
			}
		}

		if buckets, count, err := histogram(v.Resstat); err == nil {
			ch <- prometheus.MustNewConstHistogram(
				resolverQueryDuration, count, math.NaN(), buckets, v.Name,
			)
		} else {
			log.Warn("Error parsing RTT:", err)
		}
	}
	threadModel := stats.Taskmgr.ThreadModel
	ch <- prometheus.MustNewConstMetric(
		tasksRunning, prometheus.GaugeValue, float64(threadModel.TasksRunning),
	)
	ch <- prometheus.MustNewConstMetric(
		workerThreads, prometheus.GaugeValue, float64(threadModel.WorkerThreads),
	)

}

func histogram(stats []Stat) (map[float64]uint64, uint64, error) {
	buckets := map[float64]uint64{}
	var count uint64

	for _, s := range stats {
		if strings.HasPrefix(s.Name, qryRTT) {
			b := math.Inf(0)
			if !strings.HasSuffix(s.Name, "+") {
				var err error
				rrt := strings.TrimPrefix(s.Name, qryRTT)
				b, err = strconv.ParseFloat(rrt, 32)
				if err != nil {
					return buckets, 0, fmt.Errorf("could not parse RTT: %s", rrt)
				}
			}

			buckets[b/1000] = count + uint64(s.Counter)
			count += uint64(s.Counter)
		}
	}
	return buckets, count, nil
}

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9119", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
		bindURI       = flag.String("bind.statsuri", "http://localhost:8053/", "HTTP XML API address of an Bind server.")
		bindTimeout   = flag.Duration("bind.timeout", 10*time.Second, "Timeout for trying to get stats from Bind.")
		bindPidFile   = flag.String("bind.pid-file", "", "Path to Bind's pid file to export process information.")
	)
	flag.Parse()

	prometheus.MustRegister(NewExporter(*bindURI, *bindTimeout))
	if *bindPidFile != "" {
		procExporter := prometheus.NewProcessCollectorPIDFn(
			func() (int, error) {
				content, err := ioutil.ReadFile(*bindPidFile)
				if err != nil {
					return 0, fmt.Errorf("Can't read pid file: %s", err)
				}
				value, err := strconv.Atoi(strings.TrimSpace(string(content)))
				if err != nil {
					return 0, fmt.Errorf("Can't parse pid file: %s", err)
				}
				return value, nil
			}, namespace)
		prometheus.MustRegister(procExporter)
	}

	log.Info("Starting Server: ", *listenAddress)
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Bind Exporter</title></head>
             <body>
             <h1>Bind Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func unmarshal(body []byte) (Statistics, error) {
	root := Statistics{}
	if err := xml.Unmarshal([]byte(body), &root); err != nil {
		return root, err
	}
	return root, nil
}
