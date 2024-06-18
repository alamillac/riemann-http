package cerberus

import (
	"fmt"
	"sync"
	"time"
)

type Consumer interface {
	Handle(name string, ip string, total uint32, loginOk uint32, loginError uint32)
}

type IpMap struct {
	LoginOk          uint32
	LoginError       uint32
	Total            uint32
	LoginOkWindow    []uint16
	LoginErrorWindow []uint16
	TotalWindow      []uint16
}

type Memory struct {
	name                 string
	ipMap                map[string]*IpMap
	index                uint16
	size                 uint16
	tick                 uint16
	mu                   sync.Mutex
	channelInc           chan string
	channelIncLoginOk    chan string
	channelIncLoginError chan string
	consumer             Consumer
}

func (m *Memory) Display() {
	fmt.Printf("%s: %d\n", m.name, len(m.ipMap))
}

func (m *Memory) Inc(ip string) {
	m.channelInc <- ip
}

func (m *Memory) IncLoginOk(ip string) {
	m.channelIncLoginOk <- ip
}

func (m *Memory) IncLoginError(ip string) {
	m.channelIncLoginError <- ip
}

func (m *Memory) read() {
	for {
		select {
		case ip := <-m.channelInc:
			m.inc(ip)
		case ip := <-m.channelIncLoginOk:
			m.incLoginOk(ip)
		case ip := <-m.channelIncLoginError:
			m.incLoginError(ip)
		}
	}
}

func (m *Memory) inc(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ipMap, ok := m.ipMap[ip]; ok {
		ipMap.Total += 1
		ipMap.TotalWindow[m.index] += 1
	} else {
		ipMap := IpMap{
			LoginOk:          0,
			LoginError:       0,
			Total:            1,
			LoginOkWindow:    make([]uint16, m.size),
			LoginErrorWindow: make([]uint16, m.size),
			TotalWindow:      make([]uint16, m.size),
		}
		ipMap.TotalWindow[m.index] += 1
		m.ipMap[ip] = &ipMap
	}
}

func (m *Memory) incLoginOk(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ipMap, ok := m.ipMap[ip]; ok {
		ipMap.Total += 1
		ipMap.LoginOk += 1
		ipMap.TotalWindow[m.index] += 1
		ipMap.LoginOkWindow[m.index] += 1
	} else {
		ipMap := IpMap{
			LoginOk:          1,
			LoginError:       0,
			Total:            1,
			LoginOkWindow:    make([]uint16, m.size),
			LoginErrorWindow: make([]uint16, m.size),
			TotalWindow:      make([]uint16, m.size),
		}
		ipMap.TotalWindow[m.index] += 1
		ipMap.LoginOkWindow[m.index] += 1
		m.ipMap[ip] = &ipMap
	}
}

func (m *Memory) incLoginError(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ipMap, ok := m.ipMap[ip]; ok {
		ipMap.Total += 1
		ipMap.LoginError += 1
		ipMap.TotalWindow[m.index] += 1
		ipMap.LoginErrorWindow[m.index] += 1
	} else {
		ipMap := IpMap{
			LoginOk:          0,
			LoginError:       1,
			Total:            1,
			LoginOkWindow:    make([]uint16, m.size),
			LoginErrorWindow: make([]uint16, m.size),
			TotalWindow:      make([]uint16, m.size),
		}
		ipMap.TotalWindow[m.index] += 1
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
		m.consumer.Handle(m.name, ip, ipMap.Total, ipMap.LoginOk, ipMap.LoginError)

		// Remove oldest values
		ipMap.Total -= uint32(ipMap.TotalWindow[m.index]) // remove the oldest value
		if ipMap.Total == 0 {
			delete(m.ipMap, ip)
			// TODO: delete ipMap from memory?
			continue
		}

		ipMap.LoginOk -= uint32(ipMap.LoginOkWindow[m.index])       // remove the oldest value
		ipMap.LoginError -= uint32(ipMap.LoginErrorWindow[m.index]) // remove the oldest value

		ipMap.TotalWindow[m.index] = 0
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
		channelInc:           make(chan string, 1000),
		channelIncLoginOk:    make(chan string, 1000),
		channelIncLoginError: make(chan string, 1000),
		consumer:             consumer,
	}
}
