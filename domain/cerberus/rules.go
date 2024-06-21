package cerberus

import (
	"fmt"
	"log"
	"time"

	"github.com/patrickmn/go-cache"
	riemann "github.com/riemann/riemann-go-client"
)

func sendAlert(rc *riemann.TCPClient, ip, name string) error {
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

type LoginRule struct {
	numMinRequests       uint16
	minRateLoginReq      float32
	minRateLoginErrorReq float32
	cache                *cache.Cache
	client               *riemann.TCPClient
}

func (r LoginRule) Handle(name string, ip string, reqOk uint32, reqError uint32, lastError int64, loginOk uint32, loginError uint32, lastLoginError int64) {
	total := reqOk + reqError
	if total == 0 || loginError == 0 {
		return
	}
	if total < uint32(r.numMinRequests) {
		return
	}
	loginTotal := loginOk + loginError
	if float32(loginTotal)/float32(total) < r.minRateLoginReq {
		return
	}
	if float32(loginError)/float32(loginTotal) < r.minRateLoginErrorReq {
		return
	}

	// Check if the rule was already triggered for this ip
	key := fmt.Sprintf("%s%s", name, ip)
	savedLastLoginError, foundInCache := r.cache.Get(key)
	wasTriggered := foundInCache && savedLastLoginError == lastLoginError
	if wasTriggered {
		return
	}

	log.Printf("Rule triggered for %s %s Req: %d Login Req: %d Login Errors: %d\n", name, ip, total, loginTotal, loginError)
	if err := sendAlert(r.client, ip, name); err == nil {
		r.cache.Set(key, lastLoginError, cache.DefaultExpiration)
	}
}

func NewLoginRule(numMinRequests uint16, minRateLoginReq float32, minRateLoginErrorReq float32, rc *riemann.TCPClient) LoginRule {
	// Create a cache with a default expiration time of 10 minutes, and which
	// purges expired items every 10 minutes
	c := cache.New(10*time.Minute, 10*time.Minute)
	return LoginRule{
		cache:                c,
		numMinRequests:       numMinRequests,
		minRateLoginReq:      minRateLoginReq,
		minRateLoginErrorReq: minRateLoginErrorReq,
		client:               rc,
	}
}

type TotalRule struct {
	numMinRequests  uint16
	minRateErrorReq float32
	cache           *cache.Cache
	client          *riemann.TCPClient
}

func (r TotalRule) Handle(name string, ip string, reqOk uint32, reqError uint32, lastError int64, loginOk uint32, loginError uint32, lastLoginError int64) {
	total := reqOk + reqError
	if total == 0 || reqError == 0 {
		return
	}
	if total < uint32(r.numMinRequests) {
		return
	}
	if float32(reqError)/float32(total) < r.minRateErrorReq {
		return
	}

	// Check if the rule was already triggered for this ip
	key := fmt.Sprintf("%s%s", name, ip)
	savedLastError, foundInCache := r.cache.Get(key)
	wasTriggered := foundInCache && savedLastError == lastError
	if wasTriggered {
		return
	}

	log.Printf("Rule triggered for %s %s Req: %d Errors: %d\n", name, ip, total, reqError)
	if err := sendAlert(r.client, ip, name); err == nil {
		r.cache.Set(key, lastError, cache.DefaultExpiration)
	}
}

func NewTotalRule(numMinRequests uint16, minRateErrorReq float32, rc *riemann.TCPClient) TotalRule {
	c := cache.New(10*time.Minute, 10*time.Minute)
	return TotalRule{
		cache:           c,
		numMinRequests:  numMinRequests,
		minRateErrorReq: minRateErrorReq,
		client:          rc,
	}
}
