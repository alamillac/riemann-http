package cerberus

import "fmt"

type Rule struct {
	numMinRequests       uint16
	minRateLoginReq      float32
	minRateLoginErrorReq float32
}

func (r Rule) Handle(name string, ip string, total uint32, loginOk uint32, loginError uint32) {
	if total < uint32(r.numMinRequests) {
		return
	}
	if float32(loginOk)/float32(total) < r.minRateLoginReq {
		return
	}
	loginTotal := loginOk + loginError
	if float32(loginError)/float32(loginTotal) < r.minRateLoginErrorReq {
		return
	}
  fmt.Printf("Rule triggered for %s %s Req: %d Login Req: %d Login Errors: %d\n", name, ip, total, loginTotal, loginError)
}

func NewRule(numMinRequests uint16, minRateLoginReq float32, minRateLoginErrorReq float32) Rule {
	return Rule{
		numMinRequests:       numMinRequests,
		minRateLoginReq:      minRateLoginReq,
		minRateLoginErrorReq: minRateLoginErrorReq,
	}
}
