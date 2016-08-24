package main

import "time"

type ZoneV3 struct {
	Name       string     `xml:"name,attr"`
	Rdataclass string     `xml:"rdataclass,attr"`
	Serial     string     `xml:"serial"`
	Counters   CountersV3 `xml:"counters"`
}

type StatV3 struct {
	Counter int    `xml:"counter"`
	Name    string `xml:"name,attr"`
}

func dedupStatV3(stats []StatV3) []StatV3 {
	m := make(map[string]StatV3)
	for _, s := range stats {
		if s.Name == "" {
			continue
		}
		if _, ok := m[s.Name]; !ok || m[s.Name].Counter == 0 {
			m[s.Name] = s
		}
	}
	result := make([]StatV3, len(m))
	i := 0
	for _, s := range m {
		result[i] = s
		i++
	}
	return result
}

type RRSetV3 struct {
        Counter int    `xml:"counter"`
        Name    string `xml:"name"`
}

func dedupRRSetV3(stats []RRSetV3) []RRSetV3 {
        m := make(map[string]RRSetV3)
        for _, s := range stats {
                if _, ok := m[s.Name]; !ok || m[s.Name].Counter == 0 {
                        m[s.Name] = s
                }
        }
        result := make([]RRSetV3, len(m))
        i := 0
        for _, s := range m {
                result[i] = s
                i++
        }
        return result
}

type CountersV3 struct {
	Type    string      `xml:"type,attr"`
	Counter []CounterV3 `xml:"counter"`
}

func (c *CountersV3) dedup() {
	c.Counter = dedupCounterV3(c.Counter)
}

func (c *CountersV3) merge(d *CountersV3) {
	c.Counter = append(c.Counter, d.Counter...)
}

func dedupCounterV3(stats []CounterV3) []CounterV3 {
	m := make(map[string]CounterV3)
	for _, s := range stats {
		if _, ok := m[s.Name]; !ok || m[s.Name].Counter == 0 {
			m[s.Name] = s
		}
	}
	result := make([]CounterV3, len(m))
	i := 0
	for _, s := range m {
		result[i] = s
		i++
	}
	return result
}

func dedupCountersV3(counters []CountersV3) []CountersV3 {
	cnts := make(map[string]CountersV3)
	for _, cs := range counters {
		if old, ok := cnts[cs.Type]; ok {
			old.merge(&cs)
		} else {
			cnts[cs.Type] = cs
		}
	}
	result := make([]CountersV3, len(cnts))
	i := 0
	for _, cs := range cnts {
		cs.dedup()
		result[i] = cs
		i++
	}
	return result
}

type CounterV3 struct {
	Counter int    `xml:",chardata"`
	Name    string `xml:"name,attr"`
}

type ViewV3 struct {
	Name     string       `xml:"name,attr"`
	Counters []CountersV3 `xml:"counters"`
	Zones    []ZoneV3     `xml:"zones>zone"`
	Cache    []RRSetV3    `xml:"cache>rrset"`
}

func (v *ViewV3) merge(w ViewV3) {
	v.Zones = append(v.Zones, w.Zones...)
	v.Cache = append(v.Cache, w.Cache...)
	v.Counters = append(v.Counters, w.Counters...)
}

func (v *ViewV3) dedup() {
	v.Cache = dedupRRSetV3(v.Cache)
	v.Counters = dedupCountersV3(v.Counters)
}

//TODO expand
type SocketV3 struct {
	ID           string `xml:"id"`
	Name         string `xml:"name"`
	LocalAddress string `xml:"local-address"`
	References   uint   `xml:"references"`
}

type SocketmgrV3 struct {
	References uint       `xml:"references"`
	Sockets    []SocketV3 `xml:"sockets>socket"`
}

type TaskV3 struct {
	ID         string `xml:"id"`
	Name       string `xml:"name"`
	Quantum    uint   `xml:"quantum"`
	References uint   `xml:"references"`
	State      string `xml:"state"`
}

type ThreadModelV3 struct {
	Type           string `xml:"type"`
	WorkerThreads  uint   `xml:"worker-threads"`
	DefaultQuantum uint   `xml:"default-quantum"`
	TasksRunning   uint   `xml:"tasks-running"`
}

type TaskmgrV3 struct {
	Tasks       []TaskV3      `xml:"tasks>task"`
	ThreadModel ThreadModelV3 `xml:"thread-model"`
}

type ServerV3 struct {
	BootTime   time.Time    `xml:"boot-time"`
	ConfigTime time.Time    `xml:"config-time"`
	Counters   []CountersV3 `xml:"counters"`
}

func (s *ServerV3) dedup() {
	s.Counters = dedupCountersV3(s.Counters)
}

type MemorySummaryV3 struct {
	Total       int `xml:"TotalUse"`
	Used        int `xml:"InUse"`
	BlockSize   int `xml:"BlockSize"`
	ContextSize int `xml:"ContextSize"`
	Lost        int `xml:"Lost"`
}

type MemoryV3 struct {
	Summary MemorySummaryV3 `xml:"summary"`
}

type BindRootV3 struct {
	Server ServerV3 `xml:"server"`

	Views []ViewV3 `xml:"views>view"`

	Socketmgr SocketmgrV3 `xml:"socketmgr"`
	Taskmgr   TaskmgrV3   `xml:"taskmgr"`
	Memory    MemoryV3    `xml:"memory"`
}

func (r *BindRootV3) dedup() {
	views := make(map[string]ViewV3)
	for _, v := range r.Views {
		if old, ok := views[v.Name]; ok {
			old.merge(v)
		} else {
			views[v.Name] = v
		}
	}
	r.Views = make([]ViewV3, len(views))
	i := 0
	for _, v := range views {
		r.Views[i] = v
		r.Views[i].dedup()
		i++
	}

	r.Server.dedup()
}
