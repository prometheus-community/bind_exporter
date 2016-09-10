package v2

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/digitalocean/bind_exporter/bind"
)

// Client implements bind.Client and can be used to query a BIND v2 API.
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

// Stats implements bind.Stats.
func (c *Client) Stats() (bind.Statistics, error) {
	s := bind.Statistics{}

	resp, err := c.http.Get(c.url)
	if err != nil {
		return s, fmt.Errorf("error querying stats: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return s, fmt.Errorf("failed to read response: %s", err)
	}

	root := Isc{}
	if err := xml.Unmarshal([]byte(body), &root); err != nil {
		return s, fmt.Errorf("Failed to unmarshal XML response: %s", err)
	}

	stats := root.Bind.Statistics
	for _, t := range stats.Server.QueriesIn.Rdtype {
		s.Server.IncomingQueries = append(s.Server.IncomingQueries, stat(t))
	}
	for _, t := range stats.Server.Requests.Opcode {
		s.Server.IncomingRequests = append(s.Server.IncomingRequests, stat(t))
	}
	for _, t := range stats.Server.NsStats {
		s.Server.NSStats = append(s.Server.NSStats, stat(t))
	}
	for _, view := range stats.Views {
		v := bind.View{Name: view.Name}
		for _, t := range view.Cache {
			v.Cache = append(v.Cache, stat(t))
		}
		for _, t := range view.Rdtype {
			v.ResolverQueries = append(v.ResolverQueries, stat(t))
		}
		for _, t := range view.Resstat {
			v.ResolverStats = append(v.ResolverStats, stat(t))
		}
		s.Views = append(s.Views, v)
	}
	s.TaskManager = stats.Taskmgr

	return s, nil
}

func stat(s Stat) bind.Stat {
	return bind.Stat{
		Name:    s.Name,
		Counter: s.Counter,
	}
}
