package test_utils

import (
	"sync"
	"time"
)

type Mock struct {
	mu    sync.Mutex
	calls []string
	pause bool
}

func NewMock() Mock {
	return Mock{
		pause: true,
		calls: make([]string, 0),
	}
}

func (m *Mock) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = nil
	m.calls = make([]string, 0)
	m.pause = false
}

func (m *Mock) Pause() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pause = true
}

func (m *Mock) Unpause() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pause = false
}

func (m *Mock) Call(value string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = append(m.calls, value)
}

func (m *Mock) PauseCall(value string) {
	m.Call(value + "__enter")
	for m.pause == true {
		time.Sleep(time.Millisecond)
	}
	m.Call(value + "__exit")
}

func (m *Mock) CallCount(value string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for _, call := range m.calls {
		if call == value {
			count++
		}
	}

	return count
}

func (m *Mock) EnterCallCount(value string) int {
	return m.CallCount(value + "__enter")
}

func (m *Mock) ExitCallCount(value string) int {
	return m.CallCount(value + "__exit")
}

func (m *Mock) HasCall(value string) bool {
	return m.CallCount(value) > 0
}

func (m *Mock) HasEnterCall(value string) bool {
	return m.CallCount(value+"__enter") > 0
}

func (m *Mock) HasExitCall(value string) bool {
	return m.CallCount(value+"__exit") > 0
}

func (m *Mock) WaitForCall(value string, duration time.Duration) bool {
	timeout := time.Now().Add(duration)

	for time.Now().Before(timeout) {
		if m.CallCount(value) > 0 {
			return true
		}

		time.Sleep(time.Millisecond * 1)
	}

	return false
}

func (m *Mock) WaitForEnterCall(value string, duration time.Duration) bool {
	return m.WaitForCall(value+"__enter", duration)
}

func (m *Mock) WaitForExitCall(value string, duration time.Duration) bool {
	return m.WaitForCall(value+"__exit", duration)
}
