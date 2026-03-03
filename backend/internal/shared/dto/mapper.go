package dto

func MapSlice[S any, T any](src []S, mapper func(S) T) []T {
	res := make([]T, len(src))
	for i, v := range src {
		res[i] = mapper(v)
	}
	return res
}
