package serialization

import (
	"fmt"
	"net"

	"github.com/initialed85/glue/pkg/network"
	"github.com/initialed85/glue/pkg/types"
	"github.com/vmihailenco/msgpack/v5"
)

func Serialize(base *types.Container) ([]byte, error) {
	if base.Announcement != nil && base.Frame != nil {
		return []byte{}, fmt.Errorf("base cannot be both announcement and frame")
	}

	return msgpack.Marshal(base)
}

func Deserialize(data []byte) (*types.Container, error) {
	base := &types.Container{}

	err := msgpack.Unmarshal(data, base)

	if base.Announcement != nil && base.Frame != nil {
		return base, fmt.Errorf("base cannot be both announcement and frame")
	}

	base.SentByAddr, _ = net.ResolveUDPAddr(
		network.GetNetwork(base.SentBy),
		base.SentBy,
	)

	base.SentToAddr, _ = net.ResolveUDPAddr(
		network.GetNetwork(base.SentTo),
		base.SentTo,
	)

	if base.Announcement != nil {
		base.Announcement.DiscoveryListenAddr, _ = net.ResolveUDPAddr(
			network.GetNetwork(base.Announcement.DiscoveryListenAddress),
			base.Announcement.DiscoveryListenAddress,
		)

		base.Announcement.DiscoveryTargetAddr, _ = net.ResolveUDPAddr(
			network.GetNetwork(base.Announcement.DiscoveryTargetAddress),
			base.Announcement.DiscoveryTargetAddress,
		)
	}

	return base, err
}
