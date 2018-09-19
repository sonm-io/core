package xdocker

import (
	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

type Reference struct {
	reference.Reference
}

func NewReference(source string) (Reference, error) {
	ref := Reference{}
	if err := ref.Parse(source); err != nil {
		return Reference{}, err
	}
	return ref, nil
}

func (m *Reference) Parse(source string) error {
	ref, err := reference.ParseAnyReference(source)
	if err != nil {
		return err
	}
	m.Reference = ref
	return nil
}

func (m *Reference) UnmarshalText(source []byte) error {
	return m.Parse(string(source))
}

func (m Reference) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

func (m *Reference) Named() reference.Named {
	if ref, ok := m.Reference.(reference.Named); ok {
		return ref
	}
	return nil
}

func (m *Reference) WithTag(tag string) (Reference, error) {
	named := m.Named()
	if named == nil {
		return Reference{}, errors.New("reference is not named and thus can not be tagged")
	}
	named = reference.TrimNamed(named)
	ref, err := reference.WithTag(named, tag)
	if err != nil {
		return Reference{}, err
	}
	return Reference{ref}, nil
}

func (m *Reference) HasDigest() bool {
	digested, ok := m.Reference.(reference.Digested)
	return ok && len(digested.Digest()) > 0
}

func (m *Reference) Digest() digest.Digest {
	digested, ok := m.Reference.(reference.Digested)
	if !ok {
		return ""
	}
	return digested.Digest()
}

func (m *Reference) WithDigest(digest digest.Digest) (Reference, error) {
	named := m.Named()
	if named == nil {
		return Reference{}, errors.New("reference is not named and thus can not be digested")
	}
	named = reference.TrimNamed(named)
	ref, err := reference.WithDigest(named, digest)
	if err != nil {
		return Reference{}, err
	}
	return Reference{ref}, nil
}

func (m *Reference) HasName() bool {
	named, ok := m.Reference.(reference.Named)
	return ok && len(named.Name()) > 0
}

func (m *Reference) Name() string {
	named, ok := m.Reference.(reference.Named)
	if !ok {
		return ""
	}
	return named.Name()
}
