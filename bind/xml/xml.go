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

package xml

import (
	"encoding/xml"
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
	// ServerPath is the HTTP path of the v3 server resource.
	ServerPath = "/xml/v3/server"
	// StatusPath is the HTTP path of the v3 status resource.
	StatusPath = "/xml/v3/status"
	// TasksPath is the HTTP path of the v3 tasks resource.
	TasksPath = "/xml/v3/tasks"
	// TrafficPath is the HTTP path of the v3 traffic resource.
	TrafficPath = "/xml/v3/traffic"
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
	Server  Server           `xml:"server"`
	Taskmgr bind.TaskManager `xml:"taskmgr"`
	Views   []View           `xml:"views>view"`
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

type TrafficStatistics struct {
	UDPv4 []Counters `xml:"traffic>ipv4>udp>counters"`
	TCPv4 []Counters `xml:"traffic>ipv4>tcp>counters"`
	UDPv6 []Counters `xml:"traffic>ipv6>udp>counters"`
	TCPv6 []Counters `xml:"traffic>ipv6>tcp>counters"`
}

// Client implements bind.Client and can be used to query a BIND XML v3 API.
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
// v. The endpoint must return a valid XML representation which can be
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
		return fmt.Errorf("unexpected status %s", resp.Status)
	}

	if err := xml.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to unmarshal XML response: %s", err)
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

	var stats Statistics
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

	var zonestats ZoneStatistics
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

	if m[bind.TrafficStats] {
		var trafficStats TrafficStatistics
		if err := c.Get(TrafficPath, &trafficStats); err != nil {
			return s, err
		}

		var err error

		// Make IPv4 traffic histograms.
		for _, cGroup := range trafficStats.UDPv4 {
			switch cGroup.Type {
			case "request-size":
				if s.TrafficHistograms.ReceivedUDPv4, err = processTrafficCounters(cGroup.Counters, bind.TrafficInMaxSize); err != nil {
					return s, err
				}
			case "response-size":
				if s.TrafficHistograms.SentUDPv4, err = processTrafficCounters(cGroup.Counters, bind.TrafficOutMaxSize); err != nil {
					return s, err
				}
			}
		}
		for _, cGroup := range trafficStats.TCPv4 {
			switch cGroup.Type {
			case "request-size":
				if s.TrafficHistograms.ReceivedTCPv4, err = processTrafficCounters(cGroup.Counters, bind.TrafficInMaxSize); err != nil {
					return s, err
				}
			case "response-size":
				if s.TrafficHistograms.SentTCPv4, err = processTrafficCounters(cGroup.Counters, bind.TrafficOutMaxSize); err != nil {
					return s, err
				}
			}
		}

		// Make IPv6 traffic histograms.
		for _, cGroup := range trafficStats.UDPv6 {
			switch cGroup.Type {
			case "request-size":
				if s.TrafficHistograms.ReceivedUDPv6, err = processTrafficCounters(cGroup.Counters, bind.TrafficInMaxSize); err != nil {
					return s, err
				}
			case "response-size":
				if s.TrafficHistograms.SentUDPv6, err = processTrafficCounters(cGroup.Counters, bind.TrafficOutMaxSize); err != nil {
					return s, err
				}
			}
		}
		for _, cGroup := range trafficStats.TCPv6 {
			switch cGroup.Type {
			case "request-size":
				if s.TrafficHistograms.ReceivedTCPv6, err = processTrafficCounters(cGroup.Counters, bind.TrafficInMaxSize); err != nil {
					return s, err
				}
			case "response-size":
				if s.TrafficHistograms.SentTCPv6, err = processTrafficCounters(cGroup.Counters, bind.TrafficOutMaxSize); err != nil {
					return s, err
				}
			}
		}
	}

	return s, nil
}

func processTrafficCounters(traffic []bind.Counter, maxBucket uint) ([]uint64, error) {
	trafficHist := make([]uint64, maxBucket/bind.TrafficBucketSize)

	for _, c := range traffic {
		// Keys are in the format "lowerBound-upperBound". We are only interested in the upper
		// bound.
		parts := strings.Split(c.Name, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed traffic bucket range: %q", c.Name)
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
			trafficHist[idx] += c.Counter
		} else {
			// Final slice element aggregates packet sizes from maxBucket to +Inf.
			trafficHist[len(trafficHist)-1] += c.Counter
		}
	}

	return trafficHist, nil
}
