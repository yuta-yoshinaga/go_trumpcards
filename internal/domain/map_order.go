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

// namedInt は検証エラーに載せる名前と値の組。map リテラルと違い宣言順に反復できる (#8069)。
type namedInt struct {
	name  string
	value int
}
