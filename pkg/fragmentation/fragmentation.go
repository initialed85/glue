package fragmentation

import (
	"fmt"
)

func Fragment(payload []byte, fragmentSize int) ([][]byte, error) {
	if fragmentSize < 0 {
		return [][]byte{}, fmt.Errorf("fragmentSize must be 0 or above")
	}

	if fragmentSize == 0 {
		return [][]byte{payload}, nil
	}

	fragments := make([][]byte, 0)

	fragment := make([]byte, 0)

	for _, c := range payload {
		if len(fragment) == fragmentSize {
			fragments = append(fragments, fragment)
			fragment = make([]byte, 0)
		}

		fragment = append(fragment, c)
	}

	if len(fragment) > 0 {
		fragments = append(fragments, fragment)
	}

	return fragments, nil
}

func Defragment(fragments [][]byte) ([]byte, error) {
	payload := make([]byte, 0)

	for _, fragment := range fragments {
		for _, c := range fragment {
			payload = append(payload, c)
		}
	}

	return payload, nil
}
