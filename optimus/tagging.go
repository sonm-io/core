package optimus

import (
	"fmt"

	"github.com/sonm-io/core/proto"
)

type Tagger struct {
	value []byte
}

func newTagger(version string) *Tagger {
	return &Tagger{
		value: makeTag(version),
	}
}

func (m *Tagger) Tag() []byte {
	return m.value
}

func makeTag(version string) []byte {
	value := fmt.Sprintf("optimus/%s", version)

	if len(value) < sonm.MaxTagLength {
		return []byte(value)
	}

	return []byte(value[:sonm.MaxTagLength])
}
