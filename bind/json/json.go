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
	"time"

	"github.com/prometheus-community/bind_exporter/bind"
)

const (
	// ServerPath is the HTTP path of the JSON v1 server resource.
	ServerPath = "/json/v1/server"
	// TasksPath is the HTTP path of the JSON v1 tasks resource.
	TasksPath = "/json/v1/tasks"
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
			Cache      Gauges   `json:"cache"`
			Qtypes     Counters `json:"qtypes"`
			Stats      Counters `json:"stats"`
			CacheStats Counters `json:"cachestats"`
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
			for k, val := range view.Resolver.CacheStats {
				v.CacheStats = append(v.CacheStats, bind.Counter{Name: k, Counter: val})
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

	return s, nil
}
