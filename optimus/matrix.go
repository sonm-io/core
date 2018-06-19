package optimus

type row = []float64
type matrix = []row

func transpose(m matrix) matrix {
	if len(m) == 0 {
		return m
	}
	r := make(matrix, len(m[0]))
	for x := range r {
		r[x] = make(row, len(m))
	}
	for y, s := range m {
		for x, e := range s {
			r[x][y] = e
		}
	}
	return r
}
