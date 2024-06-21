package cerberus

import (
	"fmt"
	"log"
	"time"

	"github.com/patrickmn/go-cache"
)

type LoginRule struct {
	numMinRequests       uint16
	minRateLoginReq      float32
	minRateLoginErrorReq float32
	cache                *cache.Cache
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
	r.cache.Set(key, lastLoginError, cache.DefaultExpiration)
}

func NewLoginRule(numMinRequests uint16, minRateLoginReq float32, minRateLoginErrorReq float32) LoginRule {
	// Create a cache with a default expiration time of 10 minutes, and which
	// purges expired items every 10 minutes
	c := cache.New(10*time.Minute, 10*time.Minute)
	return LoginRule{
		cache:                c,
		numMinRequests:       numMinRequests,
		minRateLoginReq:      minRateLoginReq,
		minRateLoginErrorReq: minRateLoginErrorReq,
	}
}

type TotalRule struct {
	numMinRequests  uint16
	minRateErrorReq float32
	cache           *cache.Cache
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
	r.cache.Set(key, lastError, cache.DefaultExpiration)
}

func NewTotalRule(numMinRequests uint16, minRateErrorReq float32) TotalRule {
	c := cache.New(10*time.Minute, 10*time.Minute)
	return TotalRule{
		cache:           c,
		numMinRequests:  numMinRequests,
		minRateErrorReq: minRateErrorReq,
	}
}
