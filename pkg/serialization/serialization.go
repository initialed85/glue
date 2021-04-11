package serialization

import (
	"encoding/json"
	"fmt"

	"github.com/initialed85/glue/pkg/types"
)

func Serialize(base types.Container) ([]byte, error) {
	if base.Announcement != nil && base.Frame != nil {
		return []byte{}, fmt.Errorf("base cannot be both announcement and frame")
	}

	return json.Marshal(base)
}

func Deserialize(data []byte) (types.Container, error) {
	base := types.Container{}

	err := json.Unmarshal(data, &base)

	if base.Announcement != nil && base.Frame != nil {
		return base, fmt.Errorf("base cannot be both announcement and frame")
	}

	return base, err
}
