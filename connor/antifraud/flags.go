package antifraud

const (
	AllChecks = iota
	SkipBlacklisting
)

type flags int

func (f flags) SkipBlacklist() bool {
	return int(f)&SkipBlacklisting == 1
}
