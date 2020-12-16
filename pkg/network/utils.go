package network

import (
	"fmt"
	"net"
	"strings"
)

const UDPv4 = "udp4"
const UDPv6 = "udp6"
const MaxDatagramSize = 65507

func GetNetwork(rawAddr string) string {
	network := UDPv4
	if strings.Count(rawAddr, ":") > 1 {
		network = UDPv6
	}

	return network
}

func GetAddress(rawAddr string) (addr *net.UDPAddr, err error) {
	addr, err = net.ResolveUDPAddr(GetNetwork(rawAddr), rawAddr)
	if err != nil {
		err = fmt.Errorf("failed to get addr because %v", err)
		return
	}

	return addr, nil
}

func GetAddressesAndInterfaces(rawIntfc, rawAddr string) (addr *net.UDPAddr, intfc *net.Interface, srcAddr *net.UDPAddr, err error) {
	intfc, err = net.InterfaceByName(rawIntfc)
	if err != nil {
		err = fmt.Errorf("failed to get interface because %v", err)

		return
	}

	intfcAddrs, err := intfc.Addrs()
	if err != nil {
		err = fmt.Errorf("failed to get intfcAddrs because %v", err)

		return
	}

	network := GetNetwork(rawAddr)

	addr, err = GetAddress(rawAddr)
	if err != nil {
		err = fmt.Errorf("failed to get addr because %v", err)

		return
	}

	srcAddr = &net.UDPAddr{}
	for _, v := range intfcAddrs {
		ipNet := v.(*net.IPNet)

		if network == "udp4" && strings.Count(ipNet.IP.String(), ":") > 0 {
			continue
		} else if network == "udp6" && strings.Count(ipNet.IP.String(), ":") == 0 {
			continue
		} else if ipNet.IP.IsLinkLocalUnicast() || ipNet.IP.IsInterfaceLocalMulticast() {
			continue
		}

		srcAddr.IP = ipNet.IP
		srcAddr.Zone = intfc.Name

		break
	}

	return
}

// TODO: DRY this up w/ the above
func GetFreePort() (int, error) {
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenUDP("udp", addr)
	if err != nil {
		return 0, err
	}

	defer func() {
		_ = l.Close()
	}()

	return l.LocalAddr().(*net.UDPAddr).Port, nil
}

// TODO: DRY this up w/ the above
func GetDefaultInterfaceName() (string, error) {
	addr, err := GetAddress("1.1.1.1:23720")
	if err != nil {
		return "", err
	}

	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return "", err
	}

	defaultAddr := conn.LocalAddr().(*net.UDPAddr).IP

	intfcs, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	interfaceName := ""
	for _, intfc := range intfcs {
		addrs, err := intfc.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if addr.(*net.IPNet).IP.Equal(defaultAddr) {
				interfaceName = intfc.Name
			}
		}
	}

	return interfaceName, nil
}
