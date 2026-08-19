//go:build test && (!js || !wasm || extra)

package presenter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// anacondaGolden mirrors frontend/src/utils/__fixtures__/anacondaPassRecipient.golden.json.
type anacondaGolden struct {
	Cases []struct {
		Name  string `json:"name"`
		Seats []struct {
			Human bool `json:"human"`
			Out   bool `json:"out"`
		} `json:"seats"`
		Recipient int `json:"recipient"`
	} `json:"cases"`
}

// #5703: 「左隣」は脱落者を飛ばすので、席番号 +1 ではない。Web は受取人を
// 名指しできるよう同じ判定を TypeScript に持つことになったので、両者を同じ
// golden vector で縛る。
func TestAnacondaPassRecipient_GoldenVectors(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(
		"..", "..", "..", "frontend", "src", "utils", "__fixtures__", "anacondaPassRecipient.golden.json"))
	require.NoError(t, err)
	var golden anacondaGolden
	require.NoError(t, json.Unmarshal(raw, &golden))
	require.NotEmpty(t, golden.Cases)

	sawSkip, sawNone := false, false
	for _, c := range golden.Cases {
		t.Run(c.Name, func(t *testing.T) {
			players := make([]*domain.AnacondaPlayer, 0, len(c.Seats))
			for _, s := range c.Seats {
				pl := domain.NewAnacondaPlayer(s.Human, 100)
				pl.SetOut(s.Out)
				players = append(players, pl)
			}
			g := domain.NewAnaconda(domain.NewTrumpCards(0), players, domain.DefaultAnacondaConfig())

			assert.Equal(t, c.Recipient, anacondaPassRecipient(g))
		})
		if c.Recipient < 0 {
			sawNone = true
		}
		for i, s := range c.Seats {
			if s.Out && c.Recipient >= 0 && i < c.Recipient {
				sawSkip = true
			}
		}
	}
	// 負のコントロール: 脱落者を飛ばす case と受取人不在の case が要る。
	assert.True(t, sawSkip, "the golden vectors must include a skipped seat")
	assert.True(t, sawNone, "the golden vectors must include a table with no recipient")
}
