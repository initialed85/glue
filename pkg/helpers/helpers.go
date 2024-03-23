package helpers

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/initialed85/glue/pkg/network"
	"github.com/segmentio/ksuid"
)

func WaitForCtrlC() {
	var wg sync.WaitGroup

	wg.Add(1)

	sig := make(chan os.Signal, 1)

	signal.Notify(sig, os.Interrupt)

	go func() {
		<-sig
		wg.Done()
	}()

	wg.Wait()
}

func getKSUIDFromEnv(key string) (ksuid.KSUID, error) {
	rawValue := strings.TrimSpace(os.Getenv(key))

	value, err := ksuid.Parse(rawValue)
	if err != nil {
		return ksuid.Nil, fmt.Errorf("failed to parse %v=%#+v as KSUID: %v", key, rawValue, err)
	}

	return value, nil
}

func getStringFromEnv(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("empty or unset %v", key)
	}

	return value, nil
}

func getIntFromEnv(key string) (int, error) {
	rawValue := os.Getenv(key)

	value, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %v=%#+v as int: %v", key, rawValue, err)
	}

	return int(value), nil
}

func getInt64FromEnv(key string) (int64, error) {
	value, err := getIntFromEnv(key)
	if err != nil {
		return 0, err
	}

	return int64(value), err
}

func getAddrFromEnv(key string) (*net.UDPAddr, error) {
	rawValue := strings.TrimSpace(os.Getenv(key))
	if rawValue == "" {
		return nil, fmt.Errorf("empty or unset %v", key)
	}

	if rawValue == "0" {
		return nil, nil
	}

	addr, err := net.ResolveUDPAddr(network.GetNetwork(rawValue), rawValue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %v=%#+v as UDP address: %v", key, rawValue, err)
	}

	return addr, nil
}

func getDurationFromEnv(key string) (time.Duration, error) {
	rawValue := os.Getenv(key)

	value, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %v=%#+v as int: %v", key, rawValue, err)
	}

	return time.Millisecond * time.Duration(value), nil
}

func getFloat64FromEnv(key string) (float64, error) {
	rawValue := os.Getenv(key)

	value, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %v=%#+v as float: %v", key, rawValue, err)
	}

	return value, nil
}

func GetNetworkIDFromEnv() (int64, error) {
	return getInt64FromEnv("GLUE_NETWORK_ID")
}

func GetEndpointIDFromEnv() (ksuid.KSUID, error) {
	return getKSUIDFromEnv("GLUE_ENDPOINT_ID")
}

func GetEndpointNameFromEnv() (string, error) {
	return getStringFromEnv("GLUE_ENDPOINT_NAME")
}

func GetListenAddressFromEnv() (*net.UDPAddr, error) {
	return getAddrFromEnv("GLUE_LISTEN_ADDRESS")
}

func GetListenInterfaceFromEnv() (string, error) {
	return getStringFromEnv("GLUE_LISTEN_INTERFACE")
}

func GetDiscoveryTargetAddressFromEnv() (*net.UDPAddr, error) {
	return getAddrFromEnv("GLUE_DISCOVERY_TARGET_ADDRESS")
}

func GetDiscoveryListenAddressFromEnv() (*net.UDPAddr, error) {
	return getAddrFromEnv("GLUE_DISCOVERY_LISTEN_ADDRESS")
}

func GetDiscoveryRateFromEnv() (time.Duration, error) {
	return getDurationFromEnv("GLUE_DISCOVERY_RATE_MILLISECONDS")
}

func GetDiscoveryRateTimeoutMultiplierFromEnv() (float64, error) {
	return getFloat64FromEnv("GLUE_DISCOVERY_RATE_TIMEOUT_MULTIPLIER")
}
