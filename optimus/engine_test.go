package optimus

import (
	"testing"

	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/require"
)

func genPrice(price int64) *sonm.Price {
	return &sonm.Price{
		PerSecond: sonm.NewBigIntFromInt(price),
	}
}

func genAsk(price int64) *sonm.AskPlan {
	return &sonm.AskPlan{
		Price: genPrice(price),
	}
}

func TestRemoveDuplicates(t *testing.T) {
	// simple case
	create := []*sonm.AskPlan{
		genAsk(1),
		genAsk(2),
		genAsk(3),
	}
	remove := []*sonm.AskPlan{
		genAsk(2),
	}

	create, remove = removeDuplicates(create, remove)
	require.Equal(t, 2, len(create))
	require.Equal(t, 0, len(remove))
	require.Equal(t, int64(1), create[0].Price.PerSecond.Unwrap().Int64())
	require.Equal(t, int64(3), create[1].Price.PerSecond.Unwrap().Int64())

	// 2
	create = []*sonm.AskPlan{
		genAsk(1),
		genAsk(2),
		genAsk(3),
	}
	remove = []*sonm.AskPlan{
		genAsk(3),
		genAsk(4),
	}

	create, remove = removeDuplicates(create, remove)
	require.Equal(t, 2, len(create))
	require.Equal(t, 1, len(remove))
	require.Equal(t, int64(1), create[0].Price.PerSecond.Unwrap().Int64())
	require.Equal(t, int64(2), create[1].Price.PerSecond.Unwrap().Int64())
	require.Equal(t, int64(4), remove[0].Price.PerSecond.Unwrap().Int64())

	// 3
	create = []*sonm.AskPlan{
		genAsk(1),
		genAsk(1),
		genAsk(2),
	}
	remove = []*sonm.AskPlan{
		genAsk(1),
		genAsk(4),
		genAsk(4),
	}

	create, remove = removeDuplicates(create, remove)
	require.Equal(t, 2, len(create))
	require.Equal(t, 2, len(remove))
	require.Equal(t, int64(1), create[0].Price.PerSecond.Unwrap().Int64())
	require.Equal(t, int64(2), create[1].Price.PerSecond.Unwrap().Int64())
	require.Equal(t, int64(4), remove[0].Price.PerSecond.Unwrap().Int64())
	require.Equal(t, int64(4), remove[1].Price.PerSecond.Unwrap().Int64())

	//4
	create = []*sonm.AskPlan{
		genAsk(1),
		genAsk(1),
		genAsk(1),
	}
	remove = []*sonm.AskPlan{
		genAsk(1),
		genAsk(1),
		genAsk(1),
	}

	create, remove = removeDuplicates(create, remove)
	require.Equal(t, 0, len(create))
	require.Equal(t, 0, len(remove))

	// 5

	create = []*sonm.AskPlan{
		genAsk(1),
		genAsk(1),
		genAsk(1),
	}
	remove = []*sonm.AskPlan{
		genAsk(4),
		genAsk(1),
		genAsk(1),
		genAsk(1),
	}

	create, remove = removeDuplicates(create, remove)
	require.Equal(t, 0, len(create))
	require.Equal(t, 1, len(remove))
	require.Equal(t, int64(4), remove[0].Price.PerSecond.Unwrap().Int64())

	//6

	create = []*sonm.AskPlan{
		genAsk(1),
		genAsk(1),
		genAsk(1),
	}
	remove = []*sonm.AskPlan{
		genAsk(1),
		genAsk(1),
	}

	create, remove = removeDuplicates(create, remove)
	require.Equal(t, 1, len(create))
	require.Equal(t, 0, len(remove))
	require.Equal(t, int64(1), create[0].Price.PerSecond.Unwrap().Int64())

	//7

	create = []*sonm.AskPlan{
		genAsk(1),
		genAsk(1),
		genAsk(1),
		genAsk(2),
		genAsk(2),
	}
	remove = []*sonm.AskPlan{
		genAsk(2),
		genAsk(1),
		genAsk(1),
	}

	create, remove = removeDuplicates(create, remove)
	require.Equal(t, 2, len(create))
	require.Equal(t, 0, len(remove))
	require.Equal(t, int64(1), create[0].Price.PerSecond.Unwrap().Int64())
	require.Equal(t, int64(2), create[1].Price.PerSecond.Unwrap().Int64())

	//8

	create = []*sonm.AskPlan{
		genAsk(2),
		genAsk(3),
		genAsk(5),
		genAsk(4),
		genAsk(1),
		genAsk(4),
	}
	remove = []*sonm.AskPlan{
		genAsk(4),
		genAsk(2),
		genAsk(2),
		genAsk(3),
		genAsk(3),
	}

	create, remove = removeDuplicates(create, remove)
	require.Equal(t, 3, len(create))
	require.Equal(t, 2, len(remove))
	require.Equal(t, int64(1), create[0].Price.PerSecond.Unwrap().Int64())
	require.Equal(t, int64(4), create[1].Price.PerSecond.Unwrap().Int64())
	require.Equal(t, int64(5), create[2].Price.PerSecond.Unwrap().Int64())
	require.Equal(t, int64(2), remove[0].Price.PerSecond.Unwrap().Int64())
	require.Equal(t, int64(3), remove[1].Price.PerSecond.Unwrap().Int64())
}

func genBench() ([]*sonm.AskPlan, []*sonm.AskPlan) {
	create := []*sonm.AskPlan{
		genAsk(2),
		genAsk(3),
		genAsk(5),
		genAsk(4),
		genAsk(1),
		genAsk(4),
		genAsk(2),
		genAsk(3),
		genAsk(5),
		genAsk(4),
		genAsk(1),
		genAsk(4),
		genAsk(2),
		genAsk(3),
		genAsk(5),
		genAsk(4),
		genAsk(1),
		genAsk(4),
		genAsk(2),
		genAsk(3),
		genAsk(5),
		genAsk(4),
		genAsk(1),
		genAsk(4),
	}
	remove := []*sonm.AskPlan{
		genAsk(4),
		genAsk(2),
		genAsk(2),
		genAsk(3),
		genAsk(3),
		genAsk(4),
		genAsk(2),
		genAsk(2),
		genAsk(3),
		genAsk(3),
		genAsk(4),
		genAsk(2),
		genAsk(2),
		genAsk(3),
		genAsk(3),
		genAsk(4),
		genAsk(2),
		genAsk(2),
		genAsk(3),
		genAsk(3),
	}
	return create, remove
}

func BenchmarkRemoveDuplicatesVeipera(b *testing.B) {
	create, remove := genBench()

	for i := 0; i < b.N; i++ {
		removeDuplicates(create, remove)
	}
}

func BenchmarkRemoveDuplicatesKurilshchika(b *testing.B) {
	create, remove := genBench()

	for i := 0; i < b.N; i++ {
		removeDuplicates2(create, remove)
	}
}
