package domain

import "sort"

// sortedIntKeys は map のキーを昇順で返し、反復順に結果が依存するのを防ぐ (#8065)。
func sortedIntKeys[V any](m map[int]V) []int {
	keys := make([]int, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	return keys
}
