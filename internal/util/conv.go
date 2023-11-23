package util

func StringToAnySlice(strings []string) []any {
	result := make([]any, len(strings))
	for i, str := range strings {
		result[i] = str
	}
	return result
}
