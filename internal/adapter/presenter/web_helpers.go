package presenter

import (
	"fmt"
	"strings"
)

// encodeIndexedScores joins "seat:score" pairs with "," for seats 0..n-1,
// skipping every seat for which score reports ok == false.
func encodeIndexedScores(n int, score func(i int) (int, bool)) string {
	parts := make([]string, 0, n)
	for i := 0; i < n; i++ {
		value, ok := score(i)
		if !ok {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d:%d", i, value))
	}
	return strings.Join(parts, ",")
}
