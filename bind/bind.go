package bind

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"
)

// Client queries the BIND API, parses the response and returns stats in a
// generic format.
type Client interface {
	Stats(...StatisticGroup) (Statistics, error)
}

// XMLClient is a generic BIND API client to retrieve statistics in XML format.
type XMLClient struct {
	url  string
	http *http.Client
}

// NewXMLClient returns an initialized XMLClient.
func NewXMLClient(url string, c *http.Client) *XMLClient {
	return &XMLClient{
		url:  url,
		http: c,
	}
}

// Get queries the given path and stores the result in the value pointed to by
// v. The endpoint must return a valid XML representation which can be
// unmarshaled into the provided value.
func (c *XMLClient) Get(p string, v interface{}) error {
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %s", c.url, err)
	}
	u.Path = path.Join(u.Path, p)

	resp, err := c.http.Get(u.String())
	if err != nil {
		return fmt.Errorf("error querying stats: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %s", err)
	}
	if err := xml.Unmarshal([]byte(body), v); err != nil {
		return fmt.Errorf("failed to unmarshal XML response: %s", err)
	}

	return nil
}

const (
	// QryRTT is the common prefix of query round-trip histogram counters.
	QryRTT = "QryRTT"
)

// StatisticGroup describes a sub-group of BIND statistics.
type StatisticGroup string

// Available statistic groups.
const (
	ServerStats StatisticGroup = "server"
	ViewStats   StatisticGroup = "view"
	TaskStats   StatisticGroup = "tasks"
)

// Statistics is a generic representation of BIND statistics.
type Statistics struct {
	Server      Server
	Views       []View
	TaskManager TaskManager
}

// Server represents BIND server statistics.
type Server struct {
	BootTime         time.Time
	ConfigTime       time.Time
	IncomingQueries  []Counter
	IncomingRequests []Counter
	NameServerStats  []Counter
}

// View represents statistics for a single BIND view.
type View struct {
	Name            string
	Cache           []Gauge
	ResolverStats   []Counter
	ResolverQueries []Counter
}

// TaskManager contains information about all running tasks.
type TaskManager struct {
	Tasks       []Task      `xml:"tasks>task"`
	ThreadModel ThreadModel `xml:"thread-model"`
}

// Counter represents a single counter value.
type Counter struct {
	Name    string `xml:"name,attr"`
	Counter uint   `xml:",chardata"`
}

// Gauge represents a single gauge value.
type Gauge struct {
	Name  string `xml:"name"`
	Gauge int    `xml:"counter"`
}

// Task represents a single running task.
type Task struct {
	ID         string `xml:"id"`
	Name       string `xml:"name"`
	Quantum    uint   `xml:"quantum"`
	References uint   `xml:"references"`
	State      string `xml:"state"`
}

// ThreadModel contains task and worker information.
type ThreadModel struct {
	Type           string `xml:"type"`
	WorkerThreads  uint   `xml:"worker-threads"`
	DefaultQuantum uint   `xml:"default-quantum"`
	TasksRunning   uint   `xml:"tasks-running"`
}
