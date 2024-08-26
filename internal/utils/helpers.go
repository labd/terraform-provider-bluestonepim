package utils

func Ref[T any](s T) *T {
	return &s
}
