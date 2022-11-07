package stepper

func Or[T comparable](first, second T) T {
	var zero T

	if first == zero {
		return second
	}

	return first
}

func Apply[T any](initial *T, callbacks []func(*T)) *T {
	for _, callback := range callbacks {
		callback(initial)
	}

	return initial
}
