package metric

import (
  "time"
  "log"

  riemann "github.com/riemann/riemann-go-client"
)

type ASNService interface {
  GetASNForIP(string) (string, error)
}

type Service interface {
  Send(*MetricPayload) error
}

func NewService(rc *riemann.TCPClient, asnSvc ASNService) Service {
  return &svc{
    client: rc,
    asnSvc: asnSvc,
  }
}

type svc struct {
  client *riemann.TCPClient
  asnSvc ASNService
}

func (s *svc) GetAttributes(m *MetricPayload) map[string]string {
  if m.Service != "core_api.response_time" {
    return m.Attributes
  }

  ip, hasIp := m.Attributes["ip"]
  if !hasIp {
    log.Printf("Ip not found in attributes")
    return m.Attributes
  }

  log.Printf("Getting asn for metric")
  asn, err := s.asnSvc.GetASNForIP(ip)
  if err != nil {
    log.Printf("Error getting asn %s", err)
    return m.Attributes
  }

  m.Attributes["asn"] = asn
  return m.Attributes
}

func (s *svc) Send(m *MetricPayload) error {
  atts := s.GetAttributes(m)

  e := &riemann.Event{
    Service:     m.Service,
    Description: m.Description,
    Metric:      *m.Metric.Metric,
    State:       string(m.State),
    Host:        m.Host,
    Tags:        m.Tags,
    TTL:         time.Duration(m.TTL) * time.Second,
    Attributes:  atts,
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
