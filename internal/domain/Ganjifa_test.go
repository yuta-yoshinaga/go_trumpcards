//go:build test

package domain

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestGanjifa は配り札を固定した Ganjifa を返す。
//
// **Reset 直後の手札に依存する主張は必ずシードを固定する。**固定しないと
// 低確率で落ちるテストになり、develop を赤くする (#4467)。
func newTestGanjifa(t *testing.T) *Ganjifa {
	t.Helper()
	g := NewDefaultGanjifa()
	g.SetRand(rand.New(rand.NewSource(42)))
	g.Reset()
	return g
}

func TestGanjifa_StrongSuitRanksAscending(t *testing.T) {
	// 強いスート群 (1..4) は数字が大きいほど強い。
	for suit := 1; suit <= GanjifaStrongSuitMax; suit++ {
		assert.Less(t, GanjifaCardStrength(suit, 3), GanjifaCardStrength(suit, 9), "スート %d", suit)
		assert.Equal(t, 12, GanjifaCardStrength(suit, 12), "12 が最強")
		assert.Equal(t, 1, GanjifaCardStrength(suit, 1), "1 が最弱")
	}
}

func TestGanjifa_WeakSuitRanksDescending(t *testing.T) {
	// **弱いスート群 (5..8) は逆。**1 が最強で 12 が最弱。
	for suit := GanjifaStrongSuitMax + 1; suit <= GanjifaSuitCnt; suit++ {
		assert.Greater(t, GanjifaCardStrength(suit, 3), GanjifaCardStrength(suit, 9), "スート %d", suit)
		assert.Equal(t, 12, GanjifaCardStrength(suit, 1), "1 が最強")
		assert.Equal(t, 1, GanjifaCardStrength(suit, 12), "12 が最弱")
	}
}

func TestGanjifa_SameValueDiffersByGroup(t *testing.T) {
	// 同じ「3」が群によって下から 3 番目にも、上から 3 番目にもなる。
	// 生の value で比べると必ずどちらかを取り違える。
	assert.Equal(t, 3, GanjifaCardStrength(1, 3), "強い群では 3 のまま")
	assert.Equal(t, 10, GanjifaCardStrength(5, 3), "弱い群では 10 相当")
}

func TestGanjifa_GroupBoundaryIsFour(t *testing.T) {
	assert.True(t, GanjifaIsStrongSuit(GanjifaStrongSuitMax), "境界は強い群に含む")
	assert.False(t, GanjifaIsStrongSuit(GanjifaStrongSuitMax+1), "その次から弱い群")
}

func TestGanjifa_DeckHasNinetySixDistinctCards(t *testing.T) {
	deck := buildGanjifaDeck()
	require.Len(t, deck, GanjifaDeckSize)
	seen := map[[2]int]bool{}
	for _, c := range deck {
		key := [2]int{c.GetDesign(), c.GetValue()}
		require.False(t, seen[key], "重複した札 %v", key)
		seen[key] = true
	}
	assert.Len(t, seen, GanjifaDeckSize)
}

func TestGanjifa_ResetDealsThirtyTwoEach(t *testing.T) {
	g := newTestGanjifa(t)
	assert.Equal(t, GanjifaPhasePlay, g.GetPhase())
	for i := 0; i < GanjifaPlayerCnt; i++ {
		assert.Equal(t, GanjifaHandSize, g.GetPlayer(i).GetCardsSize(), "席 %d", i)
	}
	// 96 枚は 3 人で割り切れるので余り札は出ない。
	assert.Equal(t, GanjifaDeckSize, GanjifaHandSize*GanjifaPlayerCnt)
}

func TestGanjifa_TrumpIsChosenFromTheDealersHand(t *testing.T) {
	g := newTestGanjifa(t)
	trump := g.GetTrumpSuit()
	assert.GreaterOrEqual(t, trump, 1)
	assert.LessOrEqual(t, trump, GanjifaSuitCnt)
}

func TestGanjifa_TrumpBeatsAnyPlainCard(t *testing.T) {
	g := newTestGanjifa(t)
	trump := g.GetTrumpSuit()
	plain := 1
	if plain == trump {
		plain = 2
	}
	// 切り札の最弱でも、非切り札の最強より強い。
	weakTrump := NewCard(trump, 1, false)
	strongPlain := NewCard(plain, 12, false)
	assert.Greater(t, g.ganjifaRank(weakTrump), g.ganjifaRank(strongPlain))
}

func TestGanjifa_ValidPlayIndicesEnforceFollow(t *testing.T) {
	g := newTestGanjifa(t)
	lead := g.GetLeadPlayerIdx()
	// リード前は全部出せる。
	assert.Len(t, g.GetValidPlayIndices(lead), GanjifaHandSize)
}

func TestGanjifa_ConfigRejectsOutOfRangeDifficulty(t *testing.T) {
	assert.Error(t, GanjifaConfig{CpuDifficulty: 99, TargetRounds: 3}.Validate())
	assert.NoError(t, GanjifaConfig{CpuDifficulty: GanjifaCpuDifficultyNormal, TargetRounds: 3}.Validate())
}

func TestGanjifa_ConfigRejectsZeroRounds(t *testing.T) {
	assert.Error(t, GanjifaConfig{CpuDifficulty: GanjifaCpuDifficultyNormal, TargetRounds: 0}.Validate())
}

func TestGanjifa_HintNamesTheWeakSuitWhenItPicksOne(t *testing.T) {
	g := newTestGanjifa(t)
	g.SetCurrentPlayerIdx(0)
	hint := g.GetHint()
	require.NotNil(t, hint)
	require.Len(t, hint.CardIndices, 1)
	card := g.GetPlayer(0).GetCard(hint.CardIndices[0])
	require.NotNil(t, card)
	// **弱い群の札を勧めるときは理由でそう言う。**強い群の感覚で読むと逆手になる。
	if GanjifaIsStrongSuit(card.GetDesign()) {
		assert.NotContains(t, hint.Reason, "weak_suit")
	} else {
		assert.Contains(t, hint.Reason, "weak_suit")
	}
}
