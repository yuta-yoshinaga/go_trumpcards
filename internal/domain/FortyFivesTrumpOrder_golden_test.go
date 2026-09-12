//go:build test

package domain_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

type fortyFivesTrumpGoldenCard struct {
	Trump  bool   `json:"trump"`
	Design string `json:"design"`
	Value  int    `json:"value"`
}

type fortyFivesCardSpec struct {
	design int
	value  int
}

func TestFortyFivesTrumpOrder_GoldenValues(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(
		"..", "..", "frontend", "src", "constants", "fortyFivesTrumpOrder.json"))
	require.NoError(t, err)

	var golden []fortyFivesTrumpGoldenCard
	require.NoError(t, json.Unmarshal(raw, &golden))

	designs := map[string]int{
		"HEART":   domain.CardDesignHeart,
		"SPADE":   domain.CardDesignSpade,
		"CLOVER":  domain.CardDesignClover,
		"DIAMOND": domain.CardDesignDiamond,
	}
	for _, trumpSuit := range []int{domain.CardDesignHeart, domain.CardDesignSpade} {
		t.Run(map[int]string{
			domain.CardDesignHeart: "heart",
			domain.CardDesignSpade: "spade",
		}[trumpSuit], func(t *testing.T) {
			want := make([]fortyFivesCardSpec, 0, len(golden))
			for _, card := range golden {
				design := trumpSuit
				if !card.Trump {
					var ok bool
					design, ok = designs[card.Design]
					require.True(t, ok, "unknown design %q", card.Design)
				}
				if design == trumpSuit && card.Value == 1 && trumpSuit == domain.CardDesignHeart && card.Trump {
					continue
				}
				want = append(want, fortyFivesCardSpec{design: design, value: card.Value})
			}

			got := make([]fortyFivesCardSpec, 0, len(domain.FortyFivesTrumpOrder(trumpSuit)))
			for _, card := range domain.FortyFivesTrumpOrder(trumpSuit) {
				got = append(got, fortyFivesCardSpec{design: card.GetDesign(), value: card.GetValue()})
			}
			assert.Equal(t, want, got)
		})
	}
}
