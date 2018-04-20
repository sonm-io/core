package sonm

import (
	"errors"
	"strings"
	"time"
)

const MinDealDuration = time.Minute * 10

func (m *IdentityLevel) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	level, ok := IdentityLevel_value[strings.ToUpper(v)]
	if !ok {
		return errors.New("unknown identity level")
	}

	*m = IdentityLevel(level)
	return nil
}
