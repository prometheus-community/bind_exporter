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

package auto

import (
	"net/http"

	"github.com/prometheus-community/bind_exporter/bind"
	"github.com/prometheus-community/bind_exporter/bind/v2"
	"github.com/prometheus-community/bind_exporter/bind/v3"
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
