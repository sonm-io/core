package sonm

import (
	"errors"
	"strings"
	"time"
)

const MinDealDuration = time.Minute * 10

func (m *MarketIdentityLevel) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	key := "MARKET_" + strings.ToUpper(v)
	level, ok := MarketIdentityLevel_value[key]
	if !ok {
		return errors.New("unknown identity level")
	}

	*m = MarketIdentityLevel(level)
	return nil
}
