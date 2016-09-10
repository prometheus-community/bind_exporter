package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/digitalocean/bind_exporter/bind"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

var (
	serverStats = []string{
		`bind_incoming_queries_total{type="A"} 128417`,
		`bind_incoming_requests_total{opcode="QUERY"} 37634`,
		`bind_responses_total{result="Success"} 29313`,
		`bind_query_errors_total{error="Dropped"} 237`,
		`bind_query_errors_total{error="Duplicate"} 216`,
		`bind_query_errors_total{error="Failure"} 2950`,
	}
	viewStats = []string{
		`bind_resolver_cache_rrsets{type="A",view="_default"} 34324`,
		`bind_resolver_queries_total{type="CNAME",view="_default"} 28`,
		`bind_resolver_response_errors_total{error="FORMERR",view="_bind"} 0`,
		`bind_resolver_response_errors_total{error="FORMERR",view="_default"} 42906`,
		`bind_resolver_response_errors_total{error="NXDOMAIN",view="_bind"} 0`,
		`bind_resolver_response_errors_total{error="NXDOMAIN",view="_default"} 16707`,
		`bind_resolver_response_errors_total{error="OtherError",view="_bind"} 0`,
		`bind_resolver_response_errors_total{error="OtherError",view="_default"} 20660`,
		`bind_resolver_response_errors_total{error="SERVFAIL",view="_bind"} 0`,
		`bind_resolver_response_errors_total{error="SERVFAIL",view="_default"} 7596`,
		`bind_resolver_response_lame_total{view="_default"} 9108`,
		`bind_resolver_query_duration_seconds_bucket{view="_default",le="0.01"} 38334`,
		`bind_resolver_query_duration_seconds_bucket{view="_default",le="0.1"} 113122`,
		`bind_resolver_query_duration_seconds_bucket{view="_default",le="0.5"} 182658`,
		`bind_resolver_query_duration_seconds_bucket{view="_default",le="0.8"} 187375`,
		`bind_resolver_query_duration_seconds_bucket{view="_default",le="1.6"} 188409`,
		`bind_resolver_query_duration_seconds_bucket{view="_default",le="+Inf"} 227755`,
	}
	taskStats = []string{
		`bind_tasks_running 8`,
		`bind_worker_threads 16`,
	}
)

func TestBindExporterV2(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "fixtures/v2.xml")
	}))
	defer ts.Close()

	groups := []bind.StatisticGroup{bind.ServerStats, bind.ViewStats, bind.TaskStats}
	o, err := collect(NewExporter("xml.v2", ts.URL, 100*time.Millisecond, groups))
	if err != nil {
		t.Fatal(err)
	}

	shouldInclude(t, o, []string{`bind_up 1`}, serverStats, viewStats, taskStats)
}

func TestBindExporterV3(t *testing.T) {
	ts := newV3Server()
	defer ts.Close()

	groups := []bind.StatisticGroup{bind.ServerStats, bind.ViewStats, bind.TaskStats}
	o, err := collect(NewExporter("xml.v3", ts.URL, 100*time.Millisecond, groups))
	if err != nil {
		t.Fatal(err)
	}

	shouldInclude(t, o, []string{`bind_up 1`}, serverStats, viewStats, taskStats)
}

func TestBindExporterStatisticGroups(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "fixtures/v2.xml")
	}))
	defer ts.Close()

	groups := []bind.StatisticGroup{bind.ServerStats}
	o, err := collect(NewExporter("xml.v2", ts.URL, 100*time.Millisecond, groups))
	if err != nil {
		t.Fatal(err)
	}

	shouldInclude(t, o, []string{`bind_up 1`}, serverStats)
	shouldExclude(t, o, viewStats, taskStats, []string{`bind_tasks_running 0`, `bind_worker_threads 0`})
}

func shouldInclude(t *testing.T, text []byte, metrics ...[]string) {
	for _, m := range combine(metrics...) {
		if !bytes.Contains(text, []byte(m)) {
			t.Errorf("expected to find metric %q in output", m)
		}
	}
}

func shouldExclude(t *testing.T, text []byte, metrics ...[]string) {
	for _, m := range combine(metrics...) {
		if bytes.Contains(text, []byte(m)) {
			t.Errorf("expected to not find metric %q in output", m)
		}
	}
}

func combine(s ...[]string) []string {
	r := []string{}
	for _, i := range s {
		r = append(r, i...)
	}
	return r
}

func collect(c prometheus.Collector) ([]byte, error) {
	r := prometheus.NewRegistry()
	if err := r.Register(c); err != nil {
		return nil, err
	}
	m, err := r.Gather()
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	enc := expfmt.NewEncoder(&b, expfmt.FmtText)
	for _, f := range m {
		if err := enc.Encode(f); err != nil {
			return nil, err
		}
	}
	return b.Bytes(), nil
}

func newV3Server() *httptest.Server {
	m := map[string]string{
		"/xml/v3/server": "fixtures/v3/server",
		"/xml/v3/tasks":  "fixtures/v3/tasks",
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f, ok := m[r.RequestURI]; ok {
			http.ServeFile(w, r, f)
		} else {
			http.NotFound(w, r)
		}
	}))
}
