package cerberus

type Cerberus struct {
	ipMetrics      *Memory
	asnMetricsLow  *Memory
	asnMetricsHigh *Memory
}

// Analyze is the function that will analyze the metrics and apply the rules
func (c Cerberus) Analyze(ip string, asn string, isLogin bool, isUnauthorized bool) {
	// Increment IP metrics
	if isLogin {
		if isUnauthorized {
			c.ipMetrics.IncLoginError(ip)
		} else {
			c.ipMetrics.IncLoginOk(ip)
		}
	} else {
		c.ipMetrics.Inc(ip)
	}

	// Increment ASN metrics
	if asn == "27725" {
		// Ignore Cuba ASN
		return
	}

	if isLogin {
		if isUnauthorized {
			c.asnMetricsLow.IncLoginError(asn)
			c.asnMetricsHigh.IncLoginError(asn)
		} else {
			c.asnMetricsLow.IncLoginOk(asn)
			c.asnMetricsHigh.IncLoginOk(asn)
		}
	} else {
		c.asnMetricsLow.Inc(asn)
		c.asnMetricsHigh.IncLoginOk(asn)
	}
}

func (c Cerberus) Start() {
	c.ipMetrics.Start()
	c.asnMetricsLow.Start()
	c.asnMetricsHigh.Start()
}

func NewCerberus() *Cerberus {
	ipRule := NewRule(5, 0.9, 0.9)                              // Min 5 requests, 90% of login requests, 90% of login errors
	ipMetrics := NewMemory("ip", 5, 600, ipRule)                // 10 minutes window (600 seconds) with tick every 5 seconds
	asnRule := NewRule(12, 0, 0.8)                              // Min 12 requests, 80% of login errors
	asnMetricsHighFreq := NewMemory("asn high", 1, 30, asnRule) // 30 seconds window with tick every 1 second
	asnMetricsLowFreq := NewMemory("asn low", 5, 300, asnRule)  // 5 minutes window (300 seconds) with tick every 5 seconds

	return &Cerberus{
		ipMetrics:      ipMetrics,
		asnMetricsLow:  asnMetricsLowFreq,
		asnMetricsHigh: asnMetricsHighFreq,
	}
}
