package util

type Optional[T any] struct {
	value   T
	present bool
}

func OptionalSome[T any](value T) Optional[T] {
	return Optional[T]{value: value, present: true}
}

func OptionalNone[T any]() Optional[T] {
	return Optional[T]{present: false}
}

func (o *Optional[T]) IsPresent() bool {
	return o.present
}

func (o *Optional[T]) Unwrap() T {
	if !o.present {
		panic("optional is none")
	}
	return o.value
}
