package cerberus

import (
	riemann "github.com/riemann/riemann-go-client"
	"log"
	"time"
)

func sendMetric(rc *riemann.TCPClient, name, ip string) error {
	atts := make(map[string]string)
	atts["ip-asn"] = ip
	atts["name"] = name
	e := &riemann.Event{
		Service:     "cerberus.alert",
		Description: "",
		Metric:      1,
		State:       "error",
		Host:        "cerberus",
		TTL:         time.Duration(1) * time.Minute,
		Attributes:  atts,
	}
	if _, err := riemann.SendEvent(rc, e); err == nil {
		return nil
	}

	// If fail to send event retry the connection
	if err := rc.Connect(); err != nil {
		return err
	}

	if _, err := riemann.SendEvent(rc, e); err != nil {
		return err
	}
	return nil
}

type BlockIp struct {
	Client  *riemann.TCPClient
	Jenkins *Jenkins
}

func (b *BlockIp) Send(name, ip string) error {
	if err := sendMetric(b.Client, name, ip); err != nil {
		log.Printf("Error sending metric: %s\n", err)
	}
	return b.Jenkins.BlockIp(ip)
}

type BlockAsn struct {
	Client  *riemann.TCPClient
	Jenkins *Jenkins
}

func (b *BlockAsn) Send(name, asn string) error {
	if err := sendMetric(b.Client, name, asn); err != nil {
		log.Printf("Error sending metric: %s\n", err)
	}
	return b.Jenkins.BlockAsn(asn)
}
