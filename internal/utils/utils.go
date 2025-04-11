package utils

func Ptr[T any](v T) *T {
	return &v
}

func Where[C ~[]T, T any](collection C, predicate func(T) bool) (out C) {
	for _, v := range collection {
		if predicate(v) {
			out = append(out, v)
		}
	}
	return
}

func First[C ~[]T, T any](collection C, predicate func(T) bool) (out T, ok bool) {
	for _, v := range collection {
		if predicate(v) {
			return v, true
		}
	}
	return out, false
}

func FirstPtr[T any](collection []T, predicate func(T) bool) (*T, bool) {
	for i := range collection {
		if predicate(collection[i]) {
			return &collection[i], true
		}
	}
	return nil, false
}

func AppendAndGetPtr[T any](slice *[]T, value T) *T {
	*slice = append(*slice, value)
	return &(*slice)[len(*slice)-1]
}

func PtrsEqual[T comparable](a, b *T) bool {
	return (a == nil && b == nil) || (a != nil && b != nil && *a == *b)
}
