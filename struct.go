package main

import "time"

const qryRTT = "QryRTT"

type Zone struct {
	Name       string   `xml:"name,attr"`
	Rdataclass string   `xml:"rdataclass,attr"`
	Serial     string   `xml:"serial"`
	Counters   Counters `xml:"counters"`
}

type Stat struct {
	Counter int    `xml:"counter"`
	Name    string `xml:"name,attr"`
}

type Counters struct {
	Type    string    `xml:"type,attr"`
	Counter []Counter `xml:"counter"`
}

type Counter struct {
	Counter int    `xml:",chardata"`
	Name    string `xml:"name,attr"`
}

type View struct {
	Name    string `xml:"name,attr"`
	Zones   []Zone `xml:"zones>zone"`
	Cache   []Stat `xml:"cache>rrset"`
	Rdtype  []Stat `xml:"rdtype"`
	Resstat []Stat `xml:"resstat"`
}

//TODO expand
type Socket struct {
	ID           string `xml:"id"`
	Name         string `xml:"name"`
	LocalAddress string `xml:"local-address"`
	References   uint   `xml:"references"`
}

type Socketmgr struct {
	References uint     `xml:"references"`
	Sockets    []Socket `xml:"sockets>socket"`
}

type Task struct {
	ID         string `xml:"id"`
	Name       string `xml:"name"`
	Quantum    uint   `xml:"quantum"`
	References uint   `xml:"references"`
	State      string `xml:"state"`
}

type ThreadModel struct {
	Type           string `xml:"type"`
	WorkerThreads  uint   `xml:"worker-threads"`
	DefaultQuantum uint   `xml:"default-quantum"`
	TasksRunning   uint   `xml:"tasks-running"`
}

type Taskmgr struct {
	Tasks       []Task      `xml:"tasks>task"`
	ThreadModel ThreadModel `xml:"thread-model"`
}

type QueriesIn struct {
	Rdtype []Stat `xml:"rdtype"`
}

type Requests struct {
	Opcode []Stat `xml:"opcode"`
}

type Server struct {
	BootTime   time.Time  `xml:"boot-time"`
	ConfigTime time.Time  `xml:"config-time"`
	Counters   []Counters `xml:"counters"`
}

type Memory struct {
	//TODO
}

type Statistics struct {
	Server Server `xml:"server"`

	Views []View `xml:"views>view"`

	Socketmgr Socketmgr `xml:"socketmgr"`
	Taskmgr   Taskmgr   `xml:"taskmgr"`
	Memory    Memory    `xml:"memory"`
}
