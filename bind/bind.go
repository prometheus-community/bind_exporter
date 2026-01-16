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

package bind

import (
	"time"
)

// Client queries the BIND API, parses the response and returns stats in a
// generic format.
type Client interface {
	Stats(...StatisticGroup) (Statistics, error)
}

const (
	// QryRTT is the common prefix of query round-trip histogram counters.
	QryRTT = "QryRTT"
)

// StatisticGroup describes a sub-group of BIND statistics.
type StatisticGroup string

// Available statistic groups.
const (
	ServerStats StatisticGroup = "server"
	ViewStats   StatisticGroup = "view"
	TaskStats   StatisticGroup = "tasks"
)

// Statistics is a generic representation of BIND statistics.
type Statistics struct {
	Server      Server
	Views       []View
	ZoneViews   []ZoneView
	TaskManager TaskManager
}

// Server represents BIND server statistics.
type Server struct {
	BootTime         time.Time
	ConfigTime       time.Time
	IncomingQueries  []Counter
	IncomingRequests []Counter
	NameServerStats  []Counter
	ZoneStatistics   []Counter
	ServerRcodes     []Counter
}

// View represents statistics for a single BIND view.
type View struct {
	Name            string
	Cache           []Gauge
	ResolverStats   []Counter
	ResolverQueries []Counter
	CacheStats      []Counter
}

// View represents statistics for a single BIND zone view.
type ZoneView struct {
	Name     string
	ZoneData []ZoneCounter
}

// TaskManager contains information about all running tasks.
type TaskManager struct {
	Tasks       []Task      `xml:"tasks>task"`
	ThreadModel ThreadModel `xml:"thread-model"`
}

// Counter represents a single counter value.
type Counter struct {
	Name    string `xml:"name,attr"`
	Counter uint64 `xml:",chardata"`
}

// Counter represents a single zone counter value.
type ZoneCounter struct {
	Name   string
	Serial string
}

// Gauge represents a single gauge value.
type Gauge struct {
	Name  string `xml:"name"`
	Gauge uint64 `xml:"counter"`
}

// Task represents a single running task.
type Task struct {
	ID         string `xml:"id"`
	Name       string `xml:"name"`
	Quantum    int64  `xml:"quantum"`
	References uint64 `xml:"references"`
	State      string `xml:"state"`
}

// ThreadModel contains task and worker information.
type ThreadModel struct {
	Type           string `xml:"type"`
	WorkerThreads  uint64 `xml:"worker-threads"`
	DefaultQuantum uint64 `xml:"default-quantum"`
	TasksRunning   uint64 `xml:"tasks-running"`
}
