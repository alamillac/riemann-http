package metric

import (
	"fmt"
	"log"
	"riemannhttp/domain/cerberus"
	"time"

	riemann "github.com/riemann/riemann-go-client"
)

type ASNService interface {
	GetASNForIP(string) (string, error)
}

type Service interface {
	Send(*MetricPayload) error
}

func NewService(rc *riemann.TCPClient, asnSvc ASNService, guardian *cerberus.Cerberus) Service {
	return &svc{
		client:   rc,
		asnSvc:   asnSvc,
		guardian: guardian,
	}
}

type svc struct {
	guardian *cerberus.Cerberus
	client   *riemann.TCPClient
	asnSvc   ASNService
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

func (s *svc) Analyze(m *MetricPayload) error {
	if m.Service != "core_api.response_time" {
		return nil
	}

	ip, hasIp := m.Attributes["ip"]
	if !hasIp {
		return fmt.Errorf("Ip not found in attributes")
	}

	asn, hasAsn := m.Attributes["asn"]
	if !hasAsn {
		return fmt.Errorf("ASN not found in attributes")
	}

	url, hasUrl := m.Attributes["url"]
	if !hasUrl {
		return fmt.Errorf("URL not found in attributes")
	}

	statusCode, hasStatusCode := m.Attributes["status_code"]
	if !hasStatusCode {
		return fmt.Errorf("Status Code not found in attributes")
	}

	isLogin := url == "/api/v2/access/login" || url == "/api/access/login"
	isUnauthorized := statusCode == "401" || statusCode == "403"
	s.guardian.Analyze(ip, asn, isLogin, isUnauthorized)
	return nil
}

func (s *svc) Send(m *MetricPayload) error {
	atts := s.GetAttributes(m)
	if err := s.Analyze(m); err != nil {
		log.Println(err)
	}

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
