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
	Counter int    `xml:"counter"`
}

type View struct {
	Name    string `xml:"name"`
	Zones   []Zone `xml:"zones>zone"`
	Resstat []Stat `xml:"resstat"`
	Rdtype  []Stat `xml:"rdtype"`
}

//TODO expand
type Socket struct {
	Id           string `xml:"id"`
	Name         string `xml:"name"`
	LocalAddress string `xml:"local-address"`
}

type Socketmgr struct {
	References int      `xml:"references"`
	Sockets    []Socket `xml:"sockets>socket"`
}

type Taskmgr struct {
	//TODO
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
