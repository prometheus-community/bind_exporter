// Copyright 2023 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package json

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus-community/bind_exporter/bind"
)

const (
	// ServerPath is the HTTP path of the JSON v1 server resource.
	ServerPath = "/json/v1/server"
	// TasksPath is the HTTP path of the JSON v1 tasks resource.
	TasksPath = "/json/v1/tasks"
	// TrafficPath is the HTTP path of the JSON v1 traffic resource.
	TrafficPath = "/json/v1/traffic"
	// ZonesPath is the HTTP path of the JSON v1 zones resource.
	ZonesPath = "/json/v1/zones"
)

type Gauges map[string]uint64
type Counters map[string]uint64

type Statistics struct {
	BootTime   time.Time `json:"boot-time"`
	ConfigTime time.Time `json:"config-time"`
	Opcodes    Counters  `json:"opcodes"`
	QTypes     Counters  `json:"qtypes"`
	NSStats    Counters  `json:"nsstats"`
	Rcodes     Counters  `json:"rcodes"`
	ZoneStats  Counters  `json:"zonestats"`
	Views      map[string]struct {
		Resolver struct {
			Cache  Gauges   `json:"cache"`
			Qtypes Counters `json:"qtypes"`
			Stats  Counters `json:"stats"`
		} `json:"resolver"`
	} `json:"views"`
}

type ZoneStatistics struct {
	Views map[string]struct {
		Zones []struct {
			Name   string `json:"name"`
			Class  string `json:"class"`
			Serial uint32 `json:"serial"` // RFC 1035 specifies SOA serial number as uint32
		} `json:"zones"`
	} `json:"views"`
}

type TaskStatistics struct {
	TaskMgr struct {
		TasksRunning  uint64 `json:"tasks-running"`
		WorkerThreads uint64 `json:"worker-threads"`
	} `json:"taskmgr"`
}

type TrafficStatistics struct {
	Traffic struct {
		ReceivedUDPv4 map[string]uint64 `json:"dns-udp-requests-sizes-received-ipv4"`
		SentUDPv4     map[string]uint64 `json:"dns-udp-responses-sizes-sent-ipv4"`
		ReceivedTCPv4 map[string]uint64 `json:"dns-tcp-requests-sizes-sent-ipv4"`
		SentTCPv4     map[string]uint64 `json:"dns-tcp-responses-sizes-sent-ipv4"`
		ReceivedUDPv6 map[string]uint64 `json:"dns-udp-requests-sizes-received-ipv6"`
		SentUDPv6     map[string]uint64 `json:"dns-udp-responses-sizes-sent-ipv6"`
		ReceivedTCPv6 map[string]uint64 `json:"dns-tcp-requests-sizes-sent-ipv6"`
		SentTCPv6     map[string]uint64 `json:"dns-tcp-responses-sizes-sent-ipv6"`
	} `json:"traffic"`
}

// Client implements bind.Client and can be used to query a BIND JSON v1 API.
type Client struct {
	url  string
	http *http.Client
}

// NewClient returns an initialized Client.
func NewClient(url string, c *http.Client) *Client {
	return &Client{
		url:  url,
		http: c,
	}
}

// Get queries the given path and stores the result in the value pointed to by
// v. The endpoint must return a valid JSON representation which can be
// unmarshaled into the provided value.
func (c *Client) Get(p string, v interface{}) error {
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %s", c.url, err)
	}
	u.Path = path.Join(u.Path, p)

	resp, err := c.http.Get(u.String())
	if err != nil {
		return fmt.Errorf("error querying stats: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status for %q: %s", u, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON response: %s", err)
	}

	return nil
}

// Stats implements bind.Stats.
func (c *Client) Stats(groups ...bind.StatisticGroup) (bind.Statistics, error) {
	s := bind.Statistics{}
	m := map[bind.StatisticGroup]bool{}
	for _, g := range groups {
		m[g] = true
	}

	if m[bind.ServerStats] || m[bind.ViewStats] {
		var stats Statistics
		if err := c.Get(ServerPath, &stats); err != nil {
			return s, err
		}

		s.Server.BootTime = stats.BootTime
		s.Server.ConfigTime = stats.ConfigTime

		for k, val := range stats.Opcodes {
			s.Server.IncomingRequests = append(s.Server.IncomingRequests, bind.Counter{Name: k, Counter: val})
		}
		for k, val := range stats.QTypes {
			s.Server.IncomingQueries = append(s.Server.IncomingQueries, bind.Counter{Name: k, Counter: val})
		}
		for k, val := range stats.NSStats {
			s.Server.NameServerStats = append(s.Server.NameServerStats, bind.Counter{Name: k, Counter: val})
		}
		for k, val := range stats.Rcodes {
			s.Server.ServerRcodes = append(s.Server.ServerRcodes, bind.Counter{Name: k, Counter: val})
		}
		for k, val := range stats.ZoneStats {
			s.Server.ZoneStatistics = append(s.Server.ZoneStatistics, bind.Counter{Name: k, Counter: val})
		}

		for name, view := range stats.Views {
			v := bind.View{Name: name}
			for k, val := range view.Resolver.Cache {
				v.Cache = append(v.Cache, bind.Gauge{Name: k, Gauge: val})
			}
			for k, val := range view.Resolver.Qtypes {
				v.ResolverQueries = append(v.ResolverQueries, bind.Counter{Name: k, Counter: val})
			}
			for k, val := range view.Resolver.Stats {
				v.ResolverStats = append(v.ResolverStats, bind.Counter{Name: k, Counter: val})
			}
			s.Views = append(s.Views, v)
		}
	}

	var zonestats ZoneStatistics
	if err := c.Get(ZonesPath, &zonestats); err != nil {
		return s, err
	}

	for name, view := range zonestats.Views {
		v := bind.ZoneView{
			Name: name,
		}
		for _, zone := range view.Zones {
			if zone.Class != "IN" {
				continue
			}
			z := bind.ZoneCounter{
				Name:   zone.Name,
				Serial: strconv.FormatUint(uint64(zone.Serial), 10),
			}
			v.ZoneData = append(v.ZoneData, z)
		}
		s.ZoneViews = append(s.ZoneViews, v)
	}

	if m[bind.TaskStats] {
		var taskstats TaskStatistics
		if err := c.Get(TasksPath, &taskstats); err != nil {
			return s, err
		}
		s.TaskManager.ThreadModel.TasksRunning = taskstats.TaskMgr.TasksRunning
		s.TaskManager.ThreadModel.WorkerThreads = taskstats.TaskMgr.WorkerThreads
	}

	if m[bind.TrafficStats] {
		var trafficStats TrafficStatistics
		if err := c.Get(TrafficPath, &trafficStats); err != nil {
			return s, err
		}

		var err error

		// Make IPv4 traffic histograms.
		if s.TrafficHistograms.ReceivedUDPv4, err = parseTrafficHist(trafficStats.Traffic.ReceivedUDPv4, bind.TrafficInMaxSize); err != nil {
			return s, err
		}
		if s.TrafficHistograms.SentUDPv4, err = parseTrafficHist(trafficStats.Traffic.SentUDPv4, bind.TrafficOutMaxSize); err != nil {
			return s, err
		}
		if s.TrafficHistograms.ReceivedTCPv4, err = parseTrafficHist(trafficStats.Traffic.ReceivedTCPv4, bind.TrafficInMaxSize); err != nil {
			return s, err
		}
		if s.TrafficHistograms.SentTCPv4, err = parseTrafficHist(trafficStats.Traffic.SentTCPv4, bind.TrafficOutMaxSize); err != nil {
			return s, err
		}

		// Make IPv6 traffic histograms.
		if s.TrafficHistograms.ReceivedUDPv6, err = parseTrafficHist(trafficStats.Traffic.ReceivedUDPv6, bind.TrafficInMaxSize); err != nil {
			return s, err
		}
		if s.TrafficHistograms.SentUDPv6, err = parseTrafficHist(trafficStats.Traffic.SentUDPv6, bind.TrafficOutMaxSize); err != nil {
			return s, err
		}
		if s.TrafficHistograms.ReceivedTCPv6, err = parseTrafficHist(trafficStats.Traffic.ReceivedTCPv6, bind.TrafficInMaxSize); err != nil {
			return s, err
		}
		if s.TrafficHistograms.SentTCPv6, err = parseTrafficHist(trafficStats.Traffic.SentTCPv6, bind.TrafficOutMaxSize); err != nil {
			return s, err
		}
	}

	return s, nil
}

func parseTrafficHist(traffic map[string]uint64, maxBucket uint) ([]uint64, error) {
	trafficHist := make([]uint64, maxBucket/bind.TrafficBucketSize)

	for k, v := range traffic {
		// Keys are in the format "lowerBound-upperBound". We are only interested in the upper
		// bound.
		parts := strings.Split(k, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed traffic bucket range: %q", k)
		}

		upperBound, err := strconv.ParseUint(parts[1], 10, 16)
		if err != nil {
			return nil, fmt.Errorf("cannot convert bucket upper bound to uint: %w", err)
		}

		if (upperBound+1)%bind.TrafficBucketSize != 0 {
			return nil, fmt.Errorf("upper bucket bound is not a multiple of %d minus one: %d",
				bind.TrafficBucketSize, upperBound)
		}

		if upperBound < uint64(maxBucket) {
			// idx is offset, since there is no 0-16 bucket reported by BIND.
			idx := (upperBound+1)/bind.TrafficBucketSize - 2
			trafficHist[idx] += v
		} else {
			// Final slice element aggregates packet sizes from maxBucket to +Inf.
			trafficHist[len(trafficHist)-1] += v
		}
	}

	return trafficHist, nil
}
