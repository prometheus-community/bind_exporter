package v2

import (
	"encoding/xml"

	"github.com/digitalocean/bind_exporter/bind"
)

type Zone struct {
	Name       string `xml:"name"`
	Rdataclass string `xml:"rdataclass"`
	Serial     string `xml:"serial"`
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
	References uint     `xml:"references"`
	Sockets    []Socket `xml:"sockets>socket"`
}

type QueriesIn struct {
	Rdtype []Stat `xml:"rdtype"`
}

type Requests struct {
	Opcode []Stat `xml:"opcode"`
}

type Server struct {
	Requests  Requests  `xml:"requests"`
	QueriesIn QueriesIn `xml:"queries-in"`

	NsStats     []Stat `xml:"nsstat"`
	SocketStats []Stat `xml:"socketstat"`
	ZoneStats   []Stat `xml:"zonestats"`
}

type Statistics struct {
	Views     []View           `xml:"views>view"`
	Socketmgr Socketmgr        `xml:"socketmgr"`
	Taskmgr   bind.TaskManager `xml:"taskmgr"`
	Server    Server           `xml:"server"`
	Memory    struct{}         `xml:"memory"`
}
type Bind struct {
	Statistics Statistics `xml:"statistics"`
}
type Isc struct {
	XMLName xml.Name `xml:"isc"`
	Bind    Bind     `xml:"bind"`
}
