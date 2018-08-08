package task_config

import (
	"bytes"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func LoadFromFile(path string, into interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	decoder := yaml.NewDecoder(bytes.NewBuffer(data))
	decoder.SetStrict(true)

	if err := decoder.Decode(into); err != nil {
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
