package service

func unique[T comparable](values []T) []T {
	uniqueValues := make(map[T]struct{})
	for _, value := range values {
		uniqueValues[value] = struct{}{}
	}

	result := make([]T, 0, len(uniqueValues))
	for value := range uniqueValues {
		result = append(result, value)
	}

	return result
}
