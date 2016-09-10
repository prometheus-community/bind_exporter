package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

func TestBindExporter(t *testing.T) {
	tests := []string{
		`bind_up 1`,
		`bind_incoming_queries_total{type="A"} 128417`,
		`bind_incoming_requests_total{opcode="QUERY"} 37634`,
		`bind_responses_total{result="Success"} 29313`,
		`bind_query_errors_total{error="Dropped"} 237`,
		`bind_query_errors_total{error="Duplicate"} 216`,
		`bind_query_errors_total{error="Failure"} 2950`,
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
		`bind_tasks_running 8`,
		`bind_worker_threads 16`,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "fixtures/v2.xml")
	}))
	defer ts.Close()

	o, err := collect(NewExporter(ts.URL, 100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		if !bytes.Contains(o, []byte(test)) {
			t.Errorf("expected to find %q in output:\n%s", string(test), string(o))
		}
	}
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
