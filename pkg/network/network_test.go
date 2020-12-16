package network

import (
	"log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIntegration_Manager(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	var err error

	lastData := make([]byte, 0)
	callback := func(addr *net.UDPAddr, data []byte) {
		log.Printf("addr=%#+v, data=%#+v", addr.String(), string(data))
		lastData = data
	}

	m1 := NewManager()
	m2 := NewManager()

	m1.Start()
	m2.Start()

	defer func() {
		m1.Stop()
		m2.Stop()
	}()

	//
	// multicast test
	//

	err = m1.RegisterCallback("239.255.192.137:27320", "en0", callback)
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 1)

	err = m2.Send("239.255.192.137:27320", []byte("Hello, world!"))
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 1)
	assert.Equal(t, []byte("Hello, world!"), lastData)

	err = m1.UnregisterCallback("239.255.192.137:27320", "en0", callback)
	if err != nil {
		log.Fatal(err)
	}

	//
	// unicast test
	//

	err = m1.RegisterCallback("0.0.0.0:27321", "lo0", callback)
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 1)

	err = m2.Send("127.0.0.1:27321", []byte("Hello, world!"))
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 1)
	assert.Equal(t, []byte("Hello, world!"), lastData)

	err = m1.UnregisterCallback("0.0.0.0:27321", "lo0", callback)
	if err != nil {
		log.Fatal(err)
	}

	_ = lastData
}

func TestGetFreePort(t *testing.T) {
	for i := 0; i < 8; i++ {
		port, err := GetFreePort()
		if err != nil {
			log.Fatal(err)
		}

		assert.Greater(t, port, 1024)
	}
}

func TestGetDefaultInterface(t *testing.T) {
	addr, err := GetDefaultInterfaceName()
	if err != nil {
		log.Fatal(err)
	}

	assert.NotEmpty(t, addr)
}
