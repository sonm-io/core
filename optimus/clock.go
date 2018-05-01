package optimus

import "time"

// Clock describes whatever that can return current time.
// Used primarily while testing.
type Clock func() time.Time
