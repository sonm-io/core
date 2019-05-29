package plugin

import (
	"fmt"

	"github.com/sonm-io/core/insonmnia/worker/volume"
)

// Cleanup describes an interface for resource freeing.
//
// The behavior of Close after the first call is undefined.
type Cleanup interface {
	// Close performs cleanup handling for the resource acquired.
	Close() error
}

type nestedCleanup struct {
	children []Cleanup
}

func newNestedCleanup() *nestedCleanup {
	return &nestedCleanup{
		children: make([]Cleanup, 0),
	}
}

func (c *nestedCleanup) Add(v Cleanup) {
	c.children = append(c.children, v)
}

func (c *nestedCleanup) Close() error {
	errors := make([]error, 0)
	for _, v := range c.children {
		if err := v.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	// Avoiding double call.
	*c = *newNestedCleanup()

	if len(errors) == 0 {
		return nil
	} else {
		return fmt.Errorf("%+v", errors)
	}
}

type volumeCleanup struct {
	driver volume.VolumeDriver
	id     string
}

func (v volumeCleanup) Close() error {
	return v.driver.RemoveVolume(v.id)
}
