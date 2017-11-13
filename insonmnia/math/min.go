package math

// NOTE: What the hell with this language?
func Min(x int, other ...int) int {
	min := x

	for _, y := range other {
		if y < min {
			min = y
		}
	}
	return min
}
