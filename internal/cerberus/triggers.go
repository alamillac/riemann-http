package cerberus

import (
	"fmt"
	"log"
	"time"

	"github.com/patrickmn/go-cache"
)

type Trigger interface {
	Handle(name string, ip string, reqOk uint32, reqError uint32, lastError int64, loginOk uint32, loginError uint32, lastLoginError int64)
}

type Action interface {
	Send(name string, ip string) error
}

type LoginTrigger struct {
	numMinRequests       uint16
	minRateLoginReq      float32
	minRateLoginErrorReq float32
	cache                *cache.Cache
	action               Action
}

func (r *LoginTrigger) Handle(name string, ip string, reqOk uint32, reqError uint32, lastError int64, loginOk uint32, loginError uint32, lastLoginError int64) {
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
	if err := r.action.Send(name, ip); err == nil {
		r.cache.Set(key, lastLoginError, cache.DefaultExpiration)
	}
}

type LoginTriggerOpts struct {
	MinRequests       uint16
	MinRateLogin      float32
	MinRateLoginError float32
	Action            Action
}

func (o LoginTriggerOpts) NewTrigger() Trigger {
	// Create a cache with a default expiration time of 10 minutes, and which
	// purges expired items every 10 minutes
	c := cache.New(10*time.Minute, 10*time.Minute)
	return &LoginTrigger{
		cache:                c,
		numMinRequests:       o.MinRequests,
		minRateLoginReq:      o.MinRateLoginError,
		minRateLoginErrorReq: o.MinRateLoginError,
		action:               o.Action,
	}
}

type RateTrigger struct {
	numMinRequests  uint16
	minRateErrorReq float32
	cache           *cache.Cache
	action          Action
}

func (r *RateTrigger) Handle(name string, ip string, reqOk uint32, reqError uint32, lastError int64, loginOk uint32, loginError uint32, lastLoginError int64) {
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
	if err := r.action.Send(name, ip); err == nil {
		r.cache.Set(key, lastError, cache.DefaultExpiration)
	}
}

type RateTriggerOpts struct {
	MinRequests  uint16
	MinRateError float32
	Action       Action
}

func (o RateTriggerOpts) NewTrigger() Trigger {
	// Create a cache with a default expiration time of 10 minutes, and which
	// purges expired items every 10 minutes
	c := cache.New(10*time.Minute, 10*time.Minute)
	return &RateTrigger{
		cache:           c,
		numMinRequests:  o.MinRequests,
		minRateErrorReq: o.MinRateError,
		action:          o.Action,
	}
}
