package network

import (
	"fmt"
	"log"
	"net"
	"sync"
)

func GetSenderConn(addr *net.UDPAddr, srcAddr *net.UDPAddr) (conn *net.UDPConn, err error) {
	network := GetNetwork(addr.String())

	conn, err = net.DialUDP(network, srcAddr, addr)
	if err != nil {
		err = fmt.Errorf("failed to DialUDP because %v", err)
		return
	}

	return
}

type Sender struct {
	rawDstAddr string
	rawSrcAddr string
	dstAddr    *net.UDPAddr
	mu         sync.Mutex
	conn       *net.UDPConn
}

func NewSender(
	rawDstAddr string,
) *Sender {
	s := Sender{
		rawDstAddr: rawDstAddr,
	}

	return &s
}

func (s *Sender) Open() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		return fmt.Errorf("cannot open, already opened")
	}

	dstAddr, err := GetAddress(s.rawDstAddr)

	conn, err := GetSenderConn(dstAddr, nil)
	if err != nil {
		return err
	}

	s.dstAddr = dstAddr
	s.conn = conn

	log.Printf("sender opened: dst=%#+v", dstAddr.String())

	return nil
}

func (s *Sender) getConn() (*net.UDPConn, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn == nil {
		return nil, fmt.Errorf("cannot send; conn is nil (possibly not connected)")
	}

	return s.conn, nil
}

func (s *Sender) GetRawSrcAddr() (string, error) {
	conn, err := s.getConn()
	if err != nil {
		return "", err
	}

	return conn.LocalAddr().String(), nil
}

func (s *Sender) Send(b []byte) error {
	conn, err := s.getConn()
	if err != nil {
		return err
	}

	_, err = conn.Write(b)

	return err
}

func (s *Sender) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn == nil {
		return
	}

	_ = s.conn.Close()

	log.Printf("sender closed: dst=%#+v", s.dstAddr.String())

	s.dstAddr = nil
	s.conn = nil
}
