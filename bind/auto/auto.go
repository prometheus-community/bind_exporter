package auto

import (
	"net/http"

	"github.com/digitalocean/bind_exporter/bind"
	"github.com/digitalocean/bind_exporter/bind/v2"
	"github.com/digitalocean/bind_exporter/bind/v3"
)

// Client is a client which automatically detects the statistics version of the
// BIND server and delegates the request to the correct versioned client.
type Client struct {
	*bind.XMLClient
}

// NewClient returns an initialized Client.
func NewClient(url string, c *http.Client) *Client {
	return &Client{XMLClient: bind.NewXMLClient(url, c)}
}

// Stats implements bind.Stats.
func (c *Client) Stats(g ...bind.StatisticGroup) (bind.Statistics, error) {
	if err := c.Get(v3.StatusPath, &bind.Statistics{}); err == nil {
		return (&v3.Client{XMLClient: c.XMLClient}).Stats(g...)
	}
	return (&v2.Client{XMLClient: c.XMLClient}).Stats(g...)
}
