package main

import (
	"encoding/xml"
	"flag"
	"io/ioutil"
	"net"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const (
	namespace = "bind"
)

type VecInfo struct {
	help   string
	labels []string
}

var (
	gaugeMetrics      = map[string]string{}
	counterMetrics    = map[string]string{}
	counterVecMetrics = map[string]*VecInfo{
		"incoming_requests": {
			help:   "number of inbound requests made",
			labels: []string{"name"},
		},
		"incoming_queries": {
			help:   "number of inbound queries made",
			labels: []string{"name"},
		},
	}

	gaugeVecMetrics = map[string]*VecInfo{}
)

// Exporter collects Binds stats from the given server and exports
// them using the prometheus metrics package.
type Exporter struct {
	URI   string
	mutex sync.RWMutex

	up prometheus.Gauge

	gauges      map[string]*prometheus.GaugeVec
	gaugeVecs   map[string]*prometheus.GaugeVec
	counters    map[string]*prometheus.CounterVec
	counterVecs map[string]*prometheus.CounterVec

	client *http.Client
}

// NewExporter returns an initialized Exporter.
func NewExporter(uri string, timeout time.Duration) *Exporter {
	counters := make(map[string]*prometheus.CounterVec, len(counterMetrics))
	counterVecs := make(map[string]*prometheus.CounterVec, len(counterVecMetrics))
	gauges := make(map[string]*prometheus.GaugeVec, len(gaugeMetrics))
	gaugeVecs := make(map[string]*prometheus.GaugeVec, len(gaugeVecMetrics))

	for name, info := range counterVecMetrics {
		counterVecs[name] = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      name,
			Help:      info.help,
		}, info.labels)
	}

	for name, info := range gaugeVecMetrics {
		gaugeVecs[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      name,
			Help:      info.help,
		}, info.labels)
	}

	for name, help := range counterMetrics {
		counters[name] = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      name,
			Help:      help,
		}, []string{})
	}

	for name, help := range gaugeMetrics {
		gauges[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      name,
			Help:      help,
		}, []string{})
	}

	// Init our exporter.
	return &Exporter{
		URI: uri,

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Was the Bind instance query successful?",
		}),

		counters:    counters,
		counterVecs: counterVecs,
		gauges:      gauges,
		gaugeVecs:   gaugeVecs,

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
	ch <- e.up.Desc()

	for _, vec := range e.counters {
		vec.Describe(ch)
	}

	for _, vec := range e.gauges {
		vec.Describe(ch)
	}
}

// Collect fetches the stats from configured bind location and
// delivers them as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	// Reset metrics.
	for _, vec := range e.gauges {
		vec.Reset()
	}

	for _, vec := range e.counters {
		vec.Reset()
	}

	for _, vec := range e.gaugeVecs {
		vec.Reset()
	}

	for _, vec := range e.counterVecs {
		vec.Reset()
	}

	resp, err := e.client.Get(e.URI)
	if err != nil {
		e.up.Set(0)
		log.Error("Error while querying Bind:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read Bind xml response body:", err)
		e.up.Set(0)
		return
	}

	root := Isc{}

	err = xml.Unmarshal([]byte(body), &root)
	if err != nil {
		log.Error(err)
		return
	}

	e.up.Set(1)

	serverNode := root.Bind.Statistics.Server

	// Incoming Queries
	for _, stat := range serverNode.Requests.Opcode {
		c := e.counterVecs["incoming_requests_total"]
		if c != nil {
			c.WithLabelValues(stat.Name).Set(float64(stat.Counter))
		}
	}

	// Incoming requests
	for _, stat := range serverNode.QueriesIn.Rdtype {
		c := e.counterVecs["incoming_queries_total"]
		if c != nil {
			c.WithLabelValues(stat.Name).Set(float64(stat.Counter))
		}
	}

	// Report metrics.
	ch <- e.up

	for _, vec := range e.counterVecs {
		vec.Collect(ch)
	}

	for _, vec := range e.gaugeVecs {
		vec.Collect(ch)
	}

	for _, vec := range e.counters {
		vec.Collect(ch)
	}

	for _, vec := range e.gauges {
		vec.Collect(ch)
	}

}

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9109", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
		bindURI       = flag.String("bind.statsuri", "http://localhost:8053/", "HTTP XML API address of an Bind server.")
		bindTimeout   = flag.Duration("bind.timeout", 10*time.Second, "Timeout for trying to get stats from Bind.")
	)
	flag.Parse()

	exporter := NewExporter(*bindURI, *bindTimeout)
	prometheus.MustRegister(exporter)

	log.Info("Starting Server:", *listenAddress)
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
