package cerberus

import "fmt"

type LoginRule struct {
	numMinRequests       uint16
	minRateLoginReq      float32
	minRateLoginErrorReq float32
}

func (r LoginRule) Handle(name string, ip string, reqOk uint32, reqError uint32, loginOk uint32, loginError uint32) {
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
	fmt.Printf("Rule triggered for %s %s Req: %d Login Req: %d Login Errors: %d\n", name, ip, total, loginTotal, loginError)
}

func NewLoginRule(numMinRequests uint16, minRateLoginReq float32, minRateLoginErrorReq float32) LoginRule {
	return LoginRule{
		numMinRequests:       numMinRequests,
		minRateLoginReq:      minRateLoginReq,
		minRateLoginErrorReq: minRateLoginErrorReq,
	}
}

type TotalRule struct {
	numMinRequests  uint16
	minRateErrorReq float32
}

func (r TotalRule) Handle(name string, ip string, reqOk uint32, reqError uint32, loginOk uint32, loginError uint32) {
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
	fmt.Printf("Rule triggered for %s %s Req: %d Errors: %d\n", name, ip, total, reqError)
}

func NewTotalRule(numMinRequests uint16, minRateErrorReq float32) TotalRule {
	return TotalRule{
		numMinRequests:  numMinRequests,
		minRateErrorReq: minRateErrorReq,
	}
}
