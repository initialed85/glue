package worker

import (
	"github.com/stretchr/testify/assert"
	"glue/internal/test_utils"
	"testing"
	"time"
)

type MockThing struct {
	mock test_utils.Mock
}

func NewMockThing() MockThing {
	return MockThing{
		mock: test_utils.NewMock(),
	}
}

func (m *MockThing) onStart() {
	m.mock.Call("onStart")
}

func (m *MockThing) onWork() {
	m.mock.PauseCall("onWork")
}

func (m *MockThing) onStop() {
	m.mock.Call("onStop")
}

func TestNewBlockedWorker(t *testing.T) {
	m := NewMockThing()

	w := NewBlockedWorker(
		m.onStart,
		m.onWork,
		m.onStop,
	)

	m.mock.Pause()
	assert.False(t, m.mock.HasCall("onStart"))
	w.Start()
	assert.True(t, m.mock.WaitForCall("onStart", time.Second))

	assert.True(t, m.mock.WaitForEnterCall("onWork", time.Second))
	assert.False(t, m.mock.HasExitCall("onWork"))

	m.mock.Unpause()
	assert.True(t, m.mock.WaitForExitCall("onWork", time.Second))

	assert.False(t, m.mock.HasCall("onStop"))
	w.Stop()
	assert.True(t, m.mock.WaitForCall("onStop", time.Second))
}

func TestNewScheduledWorker(t *testing.T) {
	m := NewMockThing()

	w := NewScheduledWorker(
		m.onStart,
		m.onWork,
		m.onStop,
		time.Millisecond*100,
	)

	m.mock.Pause()
	assert.False(t, m.mock.HasCall("onStart"))
	w.Start()
	assert.True(t, m.mock.WaitForCall("onStart", time.Second))

	m.mock.Unpause()

	before := m.mock.EnterCallCount("onWork")
	time.Sleep(time.Second)
	after := m.mock.EnterCallCount("onWork")
	diff := after - before
	assert.GreaterOrEqual(t, 11, diff)
	assert.LessOrEqual(t, 9, diff)

	assert.False(t, m.mock.HasCall("onStop"))
	w.Stop()
	assert.True(t, m.mock.WaitForCall("onStop", time.Second))
}
