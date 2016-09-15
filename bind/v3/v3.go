package v3

import (
	"net/http"
	"time"

	"github.com/digitalocean/bind_exporter/bind"
)

const (
	// ServerPath is the HTTP path of the v3 server resource.
	ServerPath = "/xml/v3/server"
	// StatusPath is the HTTP path of the v3 status resource.
	StatusPath = "/xml/v3/status"
	// TasksPath is the HTTP path of the v3 tasks resource.
	TasksPath = "/xml/v3/tasks"

	nsstat   = "nsstat"
	opcode   = "opcode"
	qtype    = "qtype"
	resqtype = "resqtype"
	resstats = "resstats"
)

type Statistics struct {
	Memory    struct{}         `xml:"memory"`
	Server    Server           `xml:"server"`
	Socketmgr struct{}         `xml:"socketmgr"`
	Taskmgr   bind.TaskManager `xml:"taskmgr"`
	Views     []View           `xml:"views>view"`
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
	Zones    struct{}     `xml:"zones>zone"`
}

type Counters struct {
	Type     string         `xml:"type,attr"`
	Counters []bind.Counter `xml:"counter"`
}

type Counter struct {
	Name    string `xml:"name"`
	Counter int    `xml:"counter"`
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
	if m[bind.TaskStats] {
		if err := c.Get(TasksPath, &stats); err != nil {
			return s, err
		}
		s.TaskManager = stats.Taskmgr
	}

	return s, nil
}
