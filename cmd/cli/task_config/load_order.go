package task_config

import (
	"github.com/jinzhu/configor"
)

func LoadFromFile(path string, into interface{}) error {
	if err := configor.Load(into, path); err != nil {
		return err
	}

	if v, ok := into.(interface {
		Validate() error
	}); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	return nil
}
