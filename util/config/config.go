package config

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

func LoadWith(dst interface{}, path string, fn func(map[interface{}]interface{})) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	ty := map[interface{}]interface{}{}
	if err := yaml.Unmarshal(data, ty); err != nil {
		return err
	}

	fn(ty)

	data, err = yaml.Marshal(ty)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, dst)
}
