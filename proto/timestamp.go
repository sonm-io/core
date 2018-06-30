package sonm

import "time"

// Unix returns the local time.Time corresponding to the given Unix time
// since January 1, 1970 UTC.
func (m *Timestamp) Unix() time.Time {
	if m == nil {
		return time.Unix(0, 0).In(time.UTC)
	}
	return time.Unix(m.Seconds, int64(m.Nanos)).In(time.UTC)
}

func CurrentTimestamp() *Timestamp {
	now := time.Now().UnixNano()
	return &Timestamp{
		Seconds: now / 1e9,
		Nanos:   int32(now % 1e9),
	}
}

func (m Timestamp) MarshalText() (text []byte, err error) {
	return m.Unix().MarshalText()
}

func (m *Timestamp) UnmarshalText(text []byte) error {
	t := &time.Time{}
	if err := t.UnmarshalText(text); err != nil {
		return err
	}
	m.Seconds = t.Unix()
	m.Nanos = int32(t.UnixNano() % 1e9)
	return nil
}
