package cerberus

import (
	"log"
	"sync"
	"time"
)

type Consumer interface {
	Handle(name string, ip string, reqOk uint32, reqError uint32, lastError int64, loginOk uint32, loginError uint32, lastLoginError int64)
}

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

type Memory struct {
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
	consumer             Consumer
}

func (m *Memory) Display() {
	log.Printf("%s: %d\n", m.name, len(m.ipMap))
}

func (m *Memory) Inc(ip string, isLogin bool, isUnauthorized bool) {
	if isLogin {
		if isUnauthorized {
			m.channelIncLoginError <- ip
		} else {
			m.channelIncLoginOk <- ip
		}
	} else {
		if isUnauthorized {
			m.channelIncError <- ip
		} else {
			m.channelIncOk <- ip
		}
	}
}

func (m *Memory) read() {
	for {
		select {
		case ip := <-m.channelIncOk:
			m.incOk(ip)
		case ip := <-m.channelIncError:
			m.incError(ip)
		case ip := <-m.channelIncLoginOk:
			m.incLoginOk(ip)
		case ip := <-m.channelIncLoginError:
			m.incLoginError(ip)
		}
	}
}

func (m *Memory) incOk(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ipMap, ok := m.ipMap[ip]; ok {
		ipMap.TotalOk += 1
		ipMap.TotalOkWindow[m.index] += 1
	} else {
		ipMap := IpMap{
			LoginOk:          0,
			LoginError:       0,
			TotalOk:          1,
			TotalError:       0,
			LoginOkWindow:    make([]uint16, m.size),
			LoginErrorWindow: make([]uint16, m.size),
			TotalOkWindow:    make([]uint16, m.size),
			TotalErrorWindow: make([]uint16, m.size),
		}
		ipMap.TotalOkWindow[m.index] += 1
		m.ipMap[ip] = &ipMap
	}
}

func (m *Memory) incError(ip string) {
	now := time.Now().Unix()
	m.mu.Lock()
	defer m.mu.Unlock()
	if ipMap, ok := m.ipMap[ip]; ok {
		ipMap.TotalError += 1
		ipMap.TotalErrorWindow[m.index] += 1
		ipMap.LastError = now
	} else {
		ipMap := IpMap{
			LoginOk:          0,
			LoginError:       0,
			TotalOk:          0,
			TotalError:       1,
			LastError:        now,
			LoginOkWindow:    make([]uint16, m.size),
			LoginErrorWindow: make([]uint16, m.size),
			TotalOkWindow:    make([]uint16, m.size),
			TotalErrorWindow: make([]uint16, m.size),
		}
		ipMap.TotalErrorWindow[m.index] += 1
		m.ipMap[ip] = &ipMap
	}
}

func (m *Memory) incLoginOk(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ipMap, ok := m.ipMap[ip]; ok {
		ipMap.TotalOk += 1
		ipMap.LoginOk += 1
		ipMap.TotalOkWindow[m.index] += 1
		ipMap.LoginOkWindow[m.index] += 1
	} else {
		ipMap := IpMap{
			LoginOk:          1,
			LoginError:       0,
			TotalOk:          1,
			TotalError:       0,
			LoginOkWindow:    make([]uint16, m.size),
			LoginErrorWindow: make([]uint16, m.size),
			TotalOkWindow:    make([]uint16, m.size),
			TotalErrorWindow: make([]uint16, m.size),
		}
		ipMap.TotalOkWindow[m.index] += 1
		ipMap.LoginOkWindow[m.index] += 1
		m.ipMap[ip] = &ipMap
	}
}

func (m *Memory) incLoginError(ip string) {
	now := time.Now().Unix()
	m.mu.Lock()
	defer m.mu.Unlock()
	if ipMap, ok := m.ipMap[ip]; ok {
		ipMap.TotalError += 1
		ipMap.LoginError += 1
		ipMap.TotalErrorWindow[m.index] += 1
		ipMap.LoginErrorWindow[m.index] += 1
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
			LoginOkWindow:    make([]uint16, m.size),
			LoginErrorWindow: make([]uint16, m.size),
			TotalOkWindow:    make([]uint16, m.size),
			TotalErrorWindow: make([]uint16, m.size),
		}
		ipMap.TotalErrorWindow[m.index] += 1
		ipMap.LoginErrorWindow[m.index] += 1
		m.ipMap[ip] = &ipMap
	}
}

func (m *Memory) step() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.index = (m.index + 1) % m.size
	for ip, ipMap := range m.ipMap {
		// Call consumer
		m.consumer.Handle(m.name, ip, ipMap.TotalOk, ipMap.TotalError, ipMap.LastError, ipMap.LoginOk, ipMap.LoginError, ipMap.LastLoginError)

		// Remove oldest values
		ipMap.TotalOk -= uint32(ipMap.TotalOkWindow[m.index])       // remove the oldest value
		ipMap.TotalError -= uint32(ipMap.TotalErrorWindow[m.index]) // remove the oldest value
		total := ipMap.TotalOk + ipMap.TotalError
		if total == 0 {
			delete(m.ipMap, ip)
			// TODO: delete ipMap from memory?
			continue
		}

		ipMap.LoginOk -= uint32(ipMap.LoginOkWindow[m.index])       // remove the oldest value
		ipMap.LoginError -= uint32(ipMap.LoginErrorWindow[m.index]) // remove the oldest value

		ipMap.TotalOkWindow[m.index] = 0
		ipMap.TotalErrorWindow[m.index] = 0
		ipMap.LoginOkWindow[m.index] = 0
		ipMap.LoginErrorWindow[m.index] = 0
	}
}

func (m *Memory) startTick() {
	ticker := time.NewTicker(time.Duration(m.tick) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.step()
		}
	}
}

func (m *Memory) startDisplay() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.Display()
		}
	}
}

func (m *Memory) Start() {
	go m.read()
	go m.startTick()
	go m.startDisplay()
}

func NewMemory(name string, tickSecs uint16, windowSecs uint16, consumer Consumer) *Memory {
	size := windowSecs / tickSecs
	return &Memory{
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
