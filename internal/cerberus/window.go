package cerberus

import (
	"log"
	"sync"
	"time"
)

type IpMap struct {
	LoginOk          uint32
	LoginError       uint32
	TotalOk          uint32
	TotalError       uint32
	LoginOkWindow    []uint16
	LoginErrorWindow []uint16
	TotalOkWindow    []uint16
	TotalErrorWindow []uint16
	LastError        int64
	LastLoginError   int64
}

type Window struct {
	name                 string
	ipMap                map[string]*IpMap
	index                uint16
	size                 uint16
	tick                 uint16
	mu                   sync.Mutex
	channelIncOk         chan string
	channelIncError      chan string
	channelIncLoginOk    chan string
	channelIncLoginError chan string
	consumer             Trigger
}

func (w *Window) Display() {
	log.Printf("%s: %d\n", w.name, len(w.ipMap))
}

func (w *Window) Inc(ip string, isLogin bool, isUnauthorized bool) {
	if isLogin {
		if isUnauthorized {
			w.channelIncLoginError <- ip
		} else {
			w.channelIncLoginOk <- ip
		}
	} else {
		if isUnauthorized {
			w.channelIncError <- ip
		} else {
			w.channelIncOk <- ip
		}
	}
}

func (w *Window) read() {
	for {
		select {
		case ip := <-w.channelIncOk:
			w.incOk(ip)
		case ip := <-w.channelIncError:
			w.incError(ip)
		case ip := <-w.channelIncLoginOk:
			w.incLoginOk(ip)
		case ip := <-w.channelIncLoginError:
			w.incLoginError(ip)
		}
	}
}

func (w *Window) incOk(ip string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if ipMap, ok := w.ipMap[ip]; ok {
		ipMap.TotalOk += 1
		ipMap.TotalOkWindow[w.index] += 1
	} else {
		ipMap := IpMap{
			LoginOk:          0,
			LoginError:       0,
			TotalOk:          1,
			TotalError:       0,
			LoginOkWindow:    make([]uint16, w.size),
			LoginErrorWindow: make([]uint16, w.size),
			TotalOkWindow:    make([]uint16, w.size),
			TotalErrorWindow: make([]uint16, w.size),
		}
		ipMap.TotalOkWindow[w.index] += 1
		w.ipMap[ip] = &ipMap
	}
}

func (w *Window) incError(ip string) {
	now := time.Now().Unix()
	w.mu.Lock()
	defer w.mu.Unlock()
	if ipMap, ok := w.ipMap[ip]; ok {
		ipMap.TotalError += 1
		ipMap.TotalErrorWindow[w.index] += 1
		ipMap.LastError = now
	} else {
		ipMap := IpMap{
			LoginOk:          0,
			LoginError:       0,
			TotalOk:          0,
			TotalError:       1,
			LastError:        now,
			LoginOkWindow:    make([]uint16, w.size),
			LoginErrorWindow: make([]uint16, w.size),
			TotalOkWindow:    make([]uint16, w.size),
			TotalErrorWindow: make([]uint16, w.size),
		}
		ipMap.TotalErrorWindow[w.index] += 1
		w.ipMap[ip] = &ipMap
	}
}

func (w *Window) incLoginOk(ip string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if ipMap, ok := w.ipMap[ip]; ok {
		ipMap.TotalOk += 1
		ipMap.LoginOk += 1
		ipMap.TotalOkWindow[w.index] += 1
		ipMap.LoginOkWindow[w.index] += 1
	} else {
		ipMap := IpMap{
			LoginOk:          1,
			LoginError:       0,
			TotalOk:          1,
			TotalError:       0,
			LoginOkWindow:    make([]uint16, w.size),
			LoginErrorWindow: make([]uint16, w.size),
			TotalOkWindow:    make([]uint16, w.size),
			TotalErrorWindow: make([]uint16, w.size),
		}
		ipMap.TotalOkWindow[w.index] += 1
		ipMap.LoginOkWindow[w.index] += 1
		w.ipMap[ip] = &ipMap
	}
}

func (w *Window) incLoginError(ip string) {
	now := time.Now().Unix()
	w.mu.Lock()
	defer w.mu.Unlock()
	if ipMap, ok := w.ipMap[ip]; ok {
		ipMap.TotalError += 1
		ipMap.LoginError += 1
		ipMap.TotalErrorWindow[w.index] += 1
		ipMap.LoginErrorWindow[w.index] += 1
		ipMap.LastError = now
		ipMap.LastLoginError = now
	} else {
		ipMap := IpMap{
			LoginOk:          0,
			LoginError:       1,
			TotalOk:          0,
			TotalError:       1,
			LastError:        now,
			LastLoginError:   now,
			LoginOkWindow:    make([]uint16, w.size),
			LoginErrorWindow: make([]uint16, w.size),
			TotalOkWindow:    make([]uint16, w.size),
			TotalErrorWindow: make([]uint16, w.size),
		}
		ipMap.TotalErrorWindow[w.index] += 1
		ipMap.LoginErrorWindow[w.index] += 1
		w.ipMap[ip] = &ipMap
	}
}

func (w *Window) step() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.index = (w.index + 1) % w.size
	for ip, ipMap := range w.ipMap {
		// Call consumer
		go w.consumer.Handle(w.name, ip, ipMap.TotalOk, ipMap.TotalError, ipMap.LastError, ipMap.LoginOk, ipMap.LoginError, ipMap.LastLoginError)

		// Remove oldest values
		ipMap.TotalOk -= uint32(ipMap.TotalOkWindow[w.index])       // remove the oldest value
		ipMap.TotalError -= uint32(ipMap.TotalErrorWindow[w.index]) // remove the oldest value
		total := ipMap.TotalOk + ipMap.TotalError
		if total == 0 {
			delete(w.ipMap, ip)
			// TODO: delete ipMap from window?
			continue
		}

		ipMap.LoginOk -= uint32(ipMap.LoginOkWindow[w.index])       // remove the oldest value
		ipMap.LoginError -= uint32(ipMap.LoginErrorWindow[w.index]) // remove the oldest value

		ipMap.TotalOkWindow[w.index] = 0
		ipMap.TotalErrorWindow[w.index] = 0
		ipMap.LoginOkWindow[w.index] = 0
		ipMap.LoginErrorWindow[w.index] = 0
	}
}

func (w *Window) startTick() {
	ticker := time.NewTicker(time.Duration(w.tick) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.step()
		}
	}
}

func (w *Window) startDisplay() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.Display()
		}
	}
}

func (w *Window) Start() {
	go w.read()
	go w.startTick()
	go w.startDisplay()
}

func NewWindow(name string, tickSecs uint16, windowSecs uint16, consumer Trigger) *Window {
	size := windowSecs / tickSecs
	return &Window{
		name:                 name,
		ipMap:                make(map[string]*IpMap),
		index:                0,
		size:                 size,
		tick:                 tickSecs,
		mu:                   sync.Mutex{},
		channelIncOk:         make(chan string, 1000),
		channelIncError:      make(chan string, 1000),
		channelIncLoginOk:    make(chan string, 1000),
		channelIncLoginError: make(chan string, 1000),
		consumer:             consumer,
	}
}
