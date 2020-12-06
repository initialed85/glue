package network

import (
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"testing"
	"time"
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
