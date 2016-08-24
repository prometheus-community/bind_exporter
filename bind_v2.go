package main

import (
	"encoding/xml"
)

type ZoneV2 struct {
	Name       string `xml:"name"`
	Rdataclass string `xml:"rdataclass"`
	Serial     string `xml:"serial"`
	//TODO a zone can also have a huge number of counters
	//              <counters>
}

type StatV2 struct {
	Name    string `xml:"name"`
	Counter uint   `xml:"counter"`
}

type ViewV2 struct {
	Name    string   `xml:"name"`
	Cache   []StatV2 `xml:"cache>rrset"`
	Rdtype  []StatV2 `xml:"rdtype"`
	Resstat []StatV2 `xml:"resstat"`
	Zones   []ZoneV2 `xml:"zones>zone"`
}

//TODO expand
type SocketV2 struct {
	ID           string `xml:"id"`
	Name         string `xml:"name"`
	LocalAddress string `xml:"local-address"`
	References   uint   `xml:"references"`
}

type SocketmgrV2 struct {
	References uint       `xml:"references"`
	Sockets    []SocketV2 `xml:"sockets>socket"`
}

type TaskV2 struct {
	ID         string `xml:"id"`
	Name       string `xml:"name"`
	Quantum    uint   `xml:"quantum"`
	References uint   `xml:"references"`
	State      string `xml:"state"`
}

type ThreadModelV2 struct {
	Type           string `xml:"type"`
	WorkerThreads  uint   `xml:"worker-threads"`
	DefaultQuantum uint   `xml:"default-quantum"`
	TasksRunning   uint   `xml:"tasks-running"`
}

type TaskmgrV2 struct {
	Tasks       []TaskV2      `xml:"tasks>task"`
	ThreadModel ThreadModelV2 `xml:"thread-model"`
}

type QueriesInV2 struct {
	Rdtype []StatV2 `xml:"rdtype"`
}

type RequestsV2 struct {
	Opcode []StatV2 `xml:"opcode"`
}

type ServerV2 struct {
	Requests  RequestsV2  `xml:"requests"`   //Most important stats
	QueriesIn QueriesInV2 `xml:"queries-in"` //Most important stats

	NsStats     []StatV2 `xml:"nsstat"`
	SocketStats []StatV2 `xml:"socketstat"`
	ZoneStats   []StatV2 `xml:"zonestats"`
}

type MemoryV2 struct {
	//TODO
}

type StatisticsV2 struct {
	Views     []ViewV2    `xml:"views>view"`
	Socketmgr SocketmgrV2 `xml:"socketmgr"`
	Taskmgr   TaskmgrV2   `xml:"taskmgr"`
	Server    ServerV2    `xml:"server"`
	Memory    MemoryV2    `xml:"memory"`
}
type BindV2 struct {
	Statistics StatisticsV2 `xml:"statistics"`
}
type BindRootV2 struct {
	XMLName xml.Name `xml:"isc"`
	Bind    BindV2   `xml:"bind"`
}
