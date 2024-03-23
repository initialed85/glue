package network

import (
	"fmt"
	"net"
	"sync"
)

type senderKey struct {
	rawDstAddr string
}

type receiverKey struct {
	rawDstAddr    string
	interfaceName string
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
	dstAddr *net.UDPAddr,
) (*Sender, error) {
	senderKey := senderKey{
		rawDstAddr: dstAddr.String(),
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	var err error

	sender, ok := m.senderBySenderKey[senderKey]
	if !ok || sender == nil {
		sender = NewSender(dstAddr)

		err = sender.Open()
		if err != nil {
			return nil, err
		}

		m.senderBySenderKey[senderKey] = sender
	}

	return sender, nil
}

func (m *Manager) GetReceiver(
	dstAddr *net.UDPAddr,
	interfaceName string,
) (*Receiver, error) {
	receiverKey := receiverKey{
		rawDstAddr:    dstAddr.String(),
		interfaceName: interfaceName,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	var err error

	receiver, ok := m.receiverByReceiverKey[receiverKey]
	if !ok || receiver == nil {
		receiver = NewReceiver(dstAddr, interfaceName)

		err = receiver.Open()
		if err != nil {
			return nil, err
		}

		m.receiverByReceiverKey[receiverKey] = receiver
	}

	return receiver, err
}

func (m *Manager) GetRawSrcAddr(
	dstAddr *net.UDPAddr,
) (*net.UDPAddr, error) {
	sender, err := m.GetSender(dstAddr)
	if err != nil {
		return nil, fmt.Errorf("network manager failed to get sender while getting src addr for %v: %v", dstAddr.String(), err)
	}

	srcAddr, err := sender.GetRawSrcAddr()
	if err != nil {
		return nil, fmt.Errorf("network manager failed to get src addr from sender for %v: %v", dstAddr.String(), err)
	}

	return srcAddr, nil
}

func (m *Manager) Send(
	dstAddr *net.UDPAddr,
	b []byte,
) error {
	sender, err := m.GetSender(dstAddr)
	if err != nil {
		return err
	}

	err = sender.Send(b)
	if err != nil {
		sender.Close()
	}

	return err
}

func (m *Manager) RegisterCallback(
	dstAddr *net.UDPAddr,
	interfaceName string,
	callback func(*net.UDPAddr, *net.UDPAddr, []byte),
) error {
	receiver, err := m.GetReceiver(dstAddr, interfaceName)
	if err != nil {
		return err
	}

	return receiver.RegisterCallback(callback)
}

func (m *Manager) UnregisterCallback(
	dstAddr *net.UDPAddr,
	interfaceName string,
	callback func(*net.UDPAddr, *net.UDPAddr, []byte),
) error {
	receiver, err := m.GetReceiver(dstAddr, interfaceName)
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
	defer m.mu.Unlock()

	for _, sender := range m.senderBySenderKey {
		sender.Close()
	}

	for _, receiver := range m.receiverByReceiverKey {
		receiver.Close()
	}
}
