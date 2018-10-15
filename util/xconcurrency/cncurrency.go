package xconcurrency

import (
	"reflect"
	"sync"
)

func Run(concurrency int, t interface{}, cb func(elem interface{})) {
	iterable := reflect.ValueOf(t)

	ln := iterable.Len()
	if ln < concurrency {
		concurrency = ln
	}

	ch := make(chan interface{})
	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for item := range ch {
				cb(item)
			}
		}()
	}

	switch iterable.Type().Kind() {
	case reflect.Slice:
		for i := 0; i < ln; i++ {
			ch <- iterable.Index(i).Interface()
		}

	case reflect.Map:
		for _, key := range iterable.MapKeys() {
			ch <- iterable.MapIndex(key).Interface()
		}
	default:
		panic("not a slice")
	}

	close(ch)
	wg.Wait()
}
