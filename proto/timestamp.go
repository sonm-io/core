package sonm

import "time"

// Unix returns the local time.Time corresponding to the given Unix time
// since January 1, 1970 UTC.
func (m Timestamp) Unix() time.Time {
	return time.Unix(m.Seconds, int64(m.Nanos))
}
