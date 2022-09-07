// Copyright 2020 The Prometheus Authors
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

package v3

import (
	"net/http"
	"time"

	"github.com/prometheus-community/bind_exporter/bind"
)

const (
	// ServerPath is the HTTP path of the v3 server resource.
	ServerPath = "/xml/v3/server"
	// StatusPath is the HTTP path of the v3 status resource.
	StatusPath = "/xml/v3/status"
	// TasksPath is the HTTP path of the v3 tasks resource.
	TasksPath = "/xml/v3/tasks"
	// ZonesPath is the HTTP path of the v3 zones resource.
	ZonesPath = "/xml/v3/zones"

	nsstat   = "nsstat"
	opcode   = "opcode"
	qtype    = "qtype"
	resqtype = "resqtype"
	resstats = "resstats"
	zonestat = "zonestat"
	rcode    = "rcode"
)

type Statistics struct {
	Memory    struct{}         `xml:"memory"`
	Server    Server           `xml:"server"`
	Socketmgr struct{}         `xml:"socketmgr"`
	Taskmgr   bind.TaskManager `xml:"taskmgr"`
	Views     []View           `xml:"views>view"`
}

type ZoneStatistics struct {
	ZoneViews []ZoneView `xml:"views>view"`
}

type Server struct {
	BootTime   time.Time  `xml:"boot-time"`
	ConfigTime time.Time  `xml:"config-time"`
	Counters   []Counters `xml:"counters"`
}

type View struct {
	Name     string       `xml:"name,attr"`
	Cache    []bind.Gauge `xml:"cache>rrset"`
	Counters []Counters   `xml:"counters"`
}

type ZoneView struct {
	Name  string        `xml:"name,attr"`
	Zones []ZoneCounter `xml:"zones>zone"`
}

type Counters struct {
	Type     string         `xml:"type,attr"`
	Counters []bind.Counter `xml:"counter"`
}

type Counter struct {
	Name    string `xml:"name"`
	Counter uint64 `xml:"counter"`
}

type ZoneCounter struct {
	Name       string `xml:"name,attr"`
	Rdataclass string `xml:"rdataclass,attr"`
	Serial     string `xml:"serial"`
}

// Client implements bind.Client and can be used to query a BIND v3 API.
type Client struct {
	*bind.XMLClient
}

// NewClient returns an initialized Client.
func NewClient(url string, c *http.Client) *Client {
	return &Client{XMLClient: bind.NewXMLClient(url, c)}
}

// Stats implements bind.Stats.
func (c *Client) Stats(groups ...bind.StatisticGroup) (bind.Statistics, error) {
	s := bind.Statistics{}
	m := map[bind.StatisticGroup]bool{}
	for _, g := range groups {
		m[g] = true
	}

	var stats Statistics
	var zonestats ZoneStatistics
	if m[bind.ServerStats] || m[bind.ViewStats] {
		if err := c.Get(ServerPath, &stats); err != nil {
			return s, err
		}

		s.Server.BootTime = stats.Server.BootTime
		s.Server.ConfigTime = stats.Server.ConfigTime
		for _, c := range stats.Server.Counters {
			switch c.Type {
			case opcode:
				s.Server.IncomingRequests = c.Counters
			case qtype:
				s.Server.IncomingQueries = c.Counters
			case nsstat:
				s.Server.NameServerStats = c.Counters
			case zonestat:
				s.Server.ZoneStatistics = c.Counters
			case rcode:
				s.Server.ServerRcodes = c.Counters
			}
		}

		for _, view := range stats.Views {
			v := bind.View{
				Name:  view.Name,
				Cache: view.Cache,
			}
			for _, c := range view.Counters {
				switch c.Type {
				case resqtype:
					v.ResolverQueries = c.Counters
				case resstats:
					v.ResolverStats = c.Counters
				}
			}
			s.Views = append(s.Views, v)
		}
	}

	if err := c.Get(ZonesPath, &zonestats); err != nil {
		return s, err
	}

	for _, view := range zonestats.ZoneViews {
		v := bind.ZoneView{
			Name: view.Name,
		}
		for _, zone := range view.Zones {
			if zone.Rdataclass != "IN" {
				continue
			}
			z := bind.ZoneCounter{
				Name:   zone.Name,
				Serial: zone.Serial,
			}
			v.ZoneData = append(v.ZoneData, z)
		}
		s.ZoneViews = append(s.ZoneViews, v)
	}

	if m[bind.TaskStats] {
		if err := c.Get(TasksPath, &stats); err != nil {
			return s, err
		}
		s.TaskManager = stats.Taskmgr
	}

	return s, nil
}
