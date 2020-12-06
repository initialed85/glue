package network

import (
	"net"
	"sync"
)

type senderKey struct {
	rawDstAddr string
}

type receiverKey struct {
	rawDstAddr string
	intfcName  string
}

type Manager struct {
	mu                    sync.Mutex
	senderBySenderKey     map[senderKey]*Sender
	receiverByReceiverKey map[receiverKey]*Receiver
}

func NewManager() *Manager {
	return &Manager{
		senderBySenderKey:     make(map[senderKey]*Sender),
		receiverByReceiverKey: make(map[receiverKey]*Receiver),
	}
}

func (m *Manager) GetSender(
	rawDstAddr string,
) (*Sender, error) {
	senderKey := senderKey{
		rawDstAddr: rawDstAddr,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	var err error

	sender, ok := m.senderBySenderKey[senderKey]
	if !ok {
		sender = NewSender(rawDstAddr)

		err = sender.Open()
		if err != nil {
			sender = nil
		} else {
			m.senderBySenderKey[senderKey] = sender
		}
	}

	return sender, err
}

func (m *Manager) GetReceiver(
	rawDstAddr string,
	intfcName string,
) (*Receiver, error) {
	receiverKey := receiverKey{
		rawDstAddr: rawDstAddr,
		intfcName:  intfcName,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	var err error

	receiver, ok := m.receiverByReceiverKey[receiverKey]
	if !ok {
		receiver = NewReceiver(rawDstAddr, intfcName)

		err = receiver.Open()
		if err != nil {
			receiver = nil
		} else {
			m.receiverByReceiverKey[receiverKey] = receiver
		}

	}

	return receiver, err
}

func (m *Manager) GetRawSrcAddr(
	rawDstAddr string,
) (string, error) {
	sender, err := m.GetSender(rawDstAddr)
	if err != nil {
		return "", err
	}

	return sender.GetRawSrcAddr()
}

func (m *Manager) Send(
	rawDstAddr string,
	b []byte,
) error {
	sender, err := m.GetSender(rawDstAddr)
	if err != nil {
		return err
	}

	for i := 0; i < 2; i++ {
		err = sender.Send(b)
		if err != nil {
			sender.Close()
			err = sender.Open()
			if err != nil {
				return err
			}

			continue
		}

		break
	}

	return nil
}

func (m *Manager) RegisterCallback(
	rawDstAddr string,
	intfcName string,
	callback func(*net.UDPAddr, []byte),
) error {
	receiver, err := m.GetReceiver(rawDstAddr, intfcName)
	if err != nil {
		return err
	}

	return receiver.RegisterCallback(callback)
}

func (m *Manager) UnregisterCallback(
	rawDstAddr string,
	intfcName string,
	callback func(*net.UDPAddr, []byte),
) error {
	receiver, err := m.GetReceiver(rawDstAddr, intfcName)
	if err != nil {
		return err
	}

	return receiver.UnregisterCallback(callback)

}

func (m *Manager) Start() {
	// noop
}

func (m *Manager) Stop() {
	m.mu.Lock()
	m.mu.Unlock()

	for _, sender := range m.senderBySenderKey {
		sender.Close()
	}

	for _, receiver := range m.receiverByReceiverKey {
		receiver.Close()
	}
}
