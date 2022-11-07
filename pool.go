package stepper

import (
	"context"
)

func Pool[T any](ctx context.Context, count int, consumer func(*T)) chan *T {
	ch := make(chan *T)

	for i := 0; i < count; i++ {
		go func() {
			for item := range ch {
				consumer(item)
			}
		}()
	}

	return ch
}
