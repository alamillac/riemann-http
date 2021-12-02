package metric

import (
	"time"

	riemann "github.com/riemann/riemann-go-client"
)

type Service interface {
	Send(*MetricPayload) error
}

func NewService(rc *riemann.TCPClient) Service {
	return &svc{
		client: rc,
	}
}

type svc struct {
	client *riemann.TCPClient
}

func (s *svc) Send(m *MetricPayload) error {
	e := &riemann.Event{
		Service:     m.Service,
		Description: m.Description,
		Metric:      *m.Metric.Metric,
		State:       string(m.State),
		Host:        m.Host,
		Tags:        m.Tags,
		TTL:         time.Duration(m.TTL) * time.Second,
		Attributes:  m.Attributes,
	}
	if _, err := riemann.SendEvent(s.client, e); err == nil {
		return nil
	}

	// If fail to send event retry the connection
	if err := s.client.Connect(); err != nil {
		return err
	}

	_, err := riemann.SendEvent(s.client, e)
	return err
}
