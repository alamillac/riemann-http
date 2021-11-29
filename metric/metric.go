package metric

import (
	"time"

	riemann "github.com/riemann/riemann-go-client"
)

type MetricState string

const (
	MetricOK       = MetricState("ok")
	MetricWarning  = MetricState("warning")
	MetricError    = MetricState("error")
	MetricCritical = MetricState("critical")
)

var (
	connectTimeout = 10 * time.Second
	address        = "127.0.0.1:5555"
)

type Metric struct {
	Service     string            `json:"service" validate:"required"`
	Description string            `json:"description" validate:"required"`
	Metric      *int64            `json:"metric" validate:"required"`
	State       MetricState       `json:"state" validate:"required"`
	Host        string            `json:"host" validate:"required"`
	Tags        []string          `json:"tags,omitempty"`
	TTL         int64             `json:"ttl,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

func (m *Metric) Send() error {
	c, err := connect(address) //TODO move

	if err != nil {
		return err
	}

	e := &riemann.Event{
		Service:     m.Service,
		Description: m.Description,
		Metric:      *m.Metric,
		State:       string(m.State),
		Host:        m.Host,
		Tags:        m.Tags,
		TTL:         time.Duration(m.TTL) * time.Second,
		Attributes:  m.Attributes,
	}
	riemann.SendEvent(c, e)

	return err
}

func connect(address string) (*riemann.TCPClient, error) {
	c := riemann.NewTCPClient(address, connectTimeout)
	err := c.Connect()

	if err != nil {
		return nil, err
	}

	return c, nil
}
