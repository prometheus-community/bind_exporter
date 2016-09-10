package bind

// Client queries the BIND API, parses the response and returns stats in a
// generic format.
type Client interface {
	Stats() (Statistics, error)
}

// Common bind field types.
const (
	QryRTT = "QryRTT"
)

// Statistics is a generic representation of BIND statistics.
type Statistics struct {
	Server struct {
		IncomingQueries  []Stat
		IncomingRequests []Stat
		NSStats          []Stat
	}
	Views       []View
	TaskManager TaskManager
}

// View represents a statistics for a single BIND view.
type View struct {
	Name            string
	Cache           []Stat
	ResolverStats   []Stat
	ResolverQueries []Stat
}

// Stat represents a single counter value.
type Stat struct {
	Name    string
	Counter uint
}

// Task represents a single running task.
type Task struct {
	ID         string `xml:"id"`
	Name       string `xml:"name"`
	Quantum    uint   `xml:"quantum"`
	References uint   `xml:"references"`
	State      string `xml:"state"`
}

// TaskManager contains information about all running tasks.
type TaskManager struct {
	Tasks       []Task      `xml:"tasks>task"`
	ThreadModel ThreadModel `xml:"thread-model"`
}

// ThreadModel contains task and worker information.
type ThreadModel struct {
	Type           string `xml:"type"`
	WorkerThreads  uint   `xml:"worker-threads"`
	DefaultQuantum uint   `xml:"default-quantum"`
	TasksRunning   uint   `xml:"tasks-running"`
}
