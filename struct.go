package main

import (
	"encoding/xml"
)

const (
	qryRTT = "QryRTT"
)

type Zone struct {
	Name       string `xml:"name"`
	Rdataclass string `xml:"rdataclass"`
	Serial     string `xml:"serial"`
	//TODO a zone can also have a huge number of counters
	//              <counters>
}

type Stat struct {
	Name    string `xml:"name"`
	Counter uint   `xml:"counter"`
}

type View struct {
	Name    string `xml:"name"`
	Cache   []Stat `xml:"cache>rrset"`
	Rdtype  []Stat `xml:"rdtype"`
	Resstat []Stat `xml:"resstat"`
	Zones   []Zone `xml:"zones>zone"`
}

//TODO expand
type Socket struct {
	ID           string `xml:"id"`
	Name         string `xml:"name"`
	LocalAddress string `xml:"local-address"`
	References   uint   `xml:"references"`
}

type Socketmgr struct {
	Sockets []Socket `xml:"sockets>socket"`
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
	Requests  Requests  `xml:"requests"`   //Most important stats
	QueriesIn QueriesIn `xml:"queries-in"` //Most important stats

	NsStats     []Stat `xml:"nstat"`
	SocketStats []Stat `xml:"socketstat"`
	ZoneStats   []Stat `xml:"zonestats"`
}

type Memory struct {
	//TODO
}

type Statistics struct {
	Views     []View    `xml:"views>view"`
	Socketmgr Socketmgr `xml:"socketmgr"`
	Taskmgr   Taskmgr   `xml:"taskmgr"`
	Server    Server    `xml:"server"`
	Memory    Memory    `xml:"memory"`
}
type Bind struct {
	Statistics Statistics `xml:"statistics"`
}
type Isc struct {
	XMLName xml.Name `xml:"isc"`
	Bind    Bind     `xml:"bind"`
}
