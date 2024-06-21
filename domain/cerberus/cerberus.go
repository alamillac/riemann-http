package cerberus

import (
	riemann "github.com/riemann/riemann-go-client"
)

type Cerberus struct {
	ipMetrics      *Memory
	asnMetricsLow  *Memory
	asnMetricsHigh *Memory
}

// Analyze is the function that will analyze the metrics and apply the rules
func (c Cerberus) Analyze(ip string, asn string, isLogin bool, isUnauthorized bool) {
	// Increment IP metrics
	c.ipMetrics.Inc(ip, isLogin, isUnauthorized)

	// Increment ASN metrics
	if asn == "27725" {
		// Ignore Cuba ASN
		return
	}

	c.asnMetricsLow.Inc(asn, isLogin, isUnauthorized)
	c.asnMetricsHigh.Inc(asn, isLogin, isUnauthorized)
}

func (c Cerberus) Start() {
	c.ipMetrics.Start()
	c.asnMetricsLow.Start()
	c.asnMetricsHigh.Start()
}

func NewCerberus(rc *riemann.TCPClient) *Cerberus {
	ipRule := NewLoginRule(5, 0.9, 0.9, rc)                     // Min 5 requests, 90% of login requests, 90% of login errors
	ipMetrics := NewMemory("ip", 5, 600, ipRule)                // 10 minutes window (600 seconds) with tick every 5 seconds
	asnRule := NewTotalRule(12, 0.8, rc)                        // Min 12 requests, 80% of errors
	asnMetricsHighFreq := NewMemory("asn-high", 1, 30, asnRule) // 30 seconds window with tick every 1 second
	asnMetricsLowFreq := NewMemory("asn-low", 5, 300, asnRule)  // 5 minutes window (300 seconds) with tick every 5 seconds

	return &Cerberus{
		ipMetrics:      ipMetrics,
		asnMetricsLow:  asnMetricsLowFreq,
		asnMetricsHigh: asnMetricsHighFreq,
	}
}
