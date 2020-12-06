package fragmentation

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestFragmentAndDefragment(t *testing.T) {
	expected := []byte("Some payload")

	fragments, err := Fragment(expected, 5)
	if err != nil {
		log.Fatal(err)
	}

	actual, err := Defragment(fragments)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, expected, actual)
}
