package network

import (
	"fmt"
	"github.com/segmentio/ksuid"
	"glue/pkg/worker"
	"log"
	"net"
	"reflect"
	"runtime"
	"sync"
	"time"
)

const Timeout = time.Second * 1

func GetReceiverConn(addr *net.UDPAddr, intfc *net.Interface) (conn *net.UDPConn, err error) {
	network := GetNetwork(addr.String())

	if addr.IP.IsMulticast() {
		conn, err = net.ListenMulticastUDP(network, intfc, addr)
		if err != nil {
			err = fmt.Errorf("failed to ListenMulticastUDP because %v", err)
			return
		}
	} else {
		conn, err = net.ListenUDP(network, addr)
		if err != nil {
			err = fmt.Errorf("failed to ListenUDP because %v", err)
			return
		}
	}

	err = conn.SetReadBuffer(MaxDatagramSize)
	if err != nil {
		err = fmt.Errorf("failed to SetReadBuffer to %v because %v", MaxDatagramSize, err)
		return
	}

	// TODO: seems to cause SetDeadline to be ignored by ReadFromUDP (blocks forever)
	// /*
	// 	file, err := conn.File()
	// 	if err != nil {
	// 		err = fmt.Errorf("failed to File for the socket because %v", err)
	// 		return
	// 	}
	//
	// 	fd := int(file.Fd())
	//
	// 	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	// 	if err != nil {
	// 		err = fmt.Errorf("failed to set SO_REUSADDR on the socket's file descriptor because %v", err)
	// 		return
	// 	}
	// */

	return
}

type Receiver struct {
	rawDstAddr string
	intfcName  string
	dstAddr    *net.UDPAddr
	srcAddr    *net.UDPAddr
	conn       *net.UDPConn
	mu         sync.Mutex
	worker     *worker.BlockedWorker
	callbacks  map[ksuid.KSUID]func(*net.UDPAddr, []byte)
}

func NewReceiver(
	rawDstAddr string,
	intfcName string,
) *Receiver {
	r := Receiver{
		rawDstAddr: rawDstAddr,
		intfcName:  intfcName,
		callbacks:  make(map[ksuid.KSUID]func(*net.UDPAddr, []byte)),
	}

	r.worker = worker.NewBlockedWorker(
		func() {},
		r.work,
		func() {},
	)

	return &r
}

func (r *Receiver) work() {
	r.mu.Lock()
	conn := r.conn
	r.mu.Unlock()

	if conn == nil {
		return
	}

	err := conn.SetDeadline(time.Now().Add(Timeout))
	if err != nil {
		panic(fmt.Errorf("caught %#+v during SetDeadline; cannot continue", err))
		return
	}

	b := make([]byte, MaxDatagramSize)

	n, srcAddr, err := conn.ReadFromUDP(b)
	if err != nil {
		// timeout- expected
		return
	}

	data := b[:n]

	r.mu.Lock()
	for _, callback := range r.callbacks {
		// TODO: fix unbounded goroutine use
		go callback(srcAddr, data)
	}
	r.mu.Unlock()
}

func (r *Receiver) RegisterCallback(
	callback func(*net.UDPAddr, []byte),
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, knownCallback := range r.callbacks {
		if reflect.ValueOf(knownCallback).Pointer() == reflect.ValueOf(callback).Pointer() {
			return fmt.Errorf(
				"cannot register callback %#+v; already registered",
				runtime.FuncForPC(reflect.ValueOf(callback).Pointer()).Name(),
			)
		}
	}

	identifier := ksuid.New()

	r.callbacks[identifier] = callback

	log.Printf("callback registered: %#+v", r.dstAddr.String())

	return nil
}

func (r *Receiver) UnregisterCallback(
	callback func(*net.UDPAddr, []byte),
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	found := false
	identifier := ksuid.KSUID{}

	for knownIdentifier, knownCallback := range r.callbacks {
		if reflect.ValueOf(knownCallback).Pointer() == reflect.ValueOf(callback).Pointer() {
			found = true
			identifier = knownIdentifier
			break
		}
	}

	if !found {
		return fmt.Errorf(
			"cannot unregister callback %#+v; not registered",
			runtime.FuncForPC(reflect.ValueOf(callback).Pointer()).Name(),
		)
	}

	delete(r.callbacks, identifier)

	log.Printf("callback unregistered: %#+v", r.dstAddr.String())

	return nil
}

func (r *Receiver) Open() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.conn != nil {
		return fmt.Errorf("cannot open, already opened")
	}

	dstAddr, intfc, srcAddr, err := GetAddressesAndInterfaces(r.intfcName, r.rawDstAddr)
	if err != nil {
		return err
	}

	conn, err := GetReceiverConn(dstAddr, intfc)
	if err != nil {
		return err
	}

	r.dstAddr = dstAddr
	r.srcAddr = srcAddr
	r.conn = conn

	r.worker.Start()

	log.Printf("receiver opened: dst=%+#v, src=%#+v", r.dstAddr.String(), r.srcAddr.String())

	return nil
}

func (r *Receiver) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.conn == nil {
		return
	}

	r.worker.Stop()

	_ = r.conn.Close()

	log.Printf("receiver closed: dst=%+#v, src=%#+v", r.dstAddr.String(), r.srcAddr.String())

	r.dstAddr = nil
	r.srcAddr = nil
	r.conn = nil
}
