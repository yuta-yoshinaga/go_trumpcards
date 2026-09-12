//go:build test

package domain_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNiuNiuMultipliers_GoldenVectors(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(
		"..", "..", "frontend", "src", "utils", "__fixtures__", "niuniuMultipliers.golden.json"))
	require.NoError(t, err)

	var golden map[string]int
	require.NoError(t, json.Unmarshal(raw, &golden))
	require.NotEmpty(t, golden)

	n := domain.NewDefaultNiuNiu()

	for key, expectedMultiplier := range golden {
		t.Run(key, func(t *testing.T) {
			var rank domain.NiuNiuRank
			switch {
			case key == "none":
				rank = domain.NiuNiuRankNone
			case key == "niuniu":
				rank = domain.NiuNiuRankNiuNiu
			case strings.HasPrefix(key, "n"):
				num, err := strconv.Atoi(strings.TrimPrefix(key, "n"))
				require.NoError(t, err)
				rank = domain.NiuNiuRank(num)
			default:
				t.Fatalf("unknown key: %s", key)
			}

			assert.Equal(t, expectedMultiplier, n.GetMultiplier(rank))
		})
	}
}
