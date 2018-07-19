package optimus

import (
	"fmt"

	"github.com/sonm-io/core/proto"
)

type Tagger struct {
	version string
}

func newTagger(version string) *Tagger {
	return &Tagger{
		version: version,
	}
}

func (m *Tagger) Tag() []byte {
	value := fmt.Sprintf("optimus/%s", m.version)

	if len(value) < sonm.MaxTagLength {
		return []byte(value)
	}

	return []byte(value[:sonm.MaxTagLength])
}
