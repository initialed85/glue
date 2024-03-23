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
	srcAddr *net.UDPAddr
	dstAddr *net.UDPAddr
	mu      sync.Mutex
	conn    *net.UDPConn
	opened  bool
}

func NewSender(
	dstAddr *net.UDPAddr,
) *Sender {
	s := Sender{
		dstAddr: dstAddr,
	}

	return &s
}

func (s *Sender) open() error {
	conn, err := GetSenderConn(s.dstAddr, s.srcAddr)
	if err != nil {
		return err
	}

	s.conn = conn
	s.srcAddr = conn.LocalAddr().(*net.UDPAddr)

	log.Printf("sender opened: dst=%#+v", s.dstAddr.String())

	return nil
}

func (s *Sender) Open() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.opened {
		return fmt.Errorf("cannot open, already opened")
	}

	s.opened = true

	return nil
}

func (s *Sender) getConn() (*net.UDPConn, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn == nil {
		s.close()

		err := s.open()
		if err != nil {
			log.Printf("warning: sender had error trying to open: %v", err)

			s.close()

			return nil, err
		}
	}

	return s.conn, nil
}

func (s *Sender) GetRawSrcAddr() (*net.UDPAddr, error) {
	conn, err := s.getConn()
	if err != nil {
		return nil, err
	}

	return conn.LocalAddr().(*net.UDPAddr), nil
}

func (s *Sender) Send(b []byte) error {
	conn, err := s.getConn()
	if err != nil {
		return err
	}

	_, err = conn.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func (s *Sender) close() {
	if s.conn != nil {
		_ = s.conn.Close()
		log.Printf("sender closed: dst=%#+v", s.dstAddr.String())
	}

	s.srcAddr = nil
	s.conn = nil
}

func (s *Sender) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.opened {
		return
	}

	s.close()
	s.opened = false
}
