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

// The whole game state lives in unexported fields, so a session that round-trips
// through KV on the Worker only survives if the custom codec carries it. Plain
// json.Marshal on the struct would emit "{}" and silently reset every hand.
func TestGanjifa_JSONRoundTripPreservesHands(t *testing.T) {
	src := NewDefaultGanjifa()
	src.SetRand(rand.New(rand.NewSource(7)))
	src.Reset()
	src.SetPhase(GanjifaPhaseTrickEnd)

	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	var got Ganjifa
	assert.NoError(t, got.UnmarshalJSON(data))
	assert.Equal(t, GanjifaPhaseTrickEnd, got.GetPhase())
	assert.Equal(t, src.GetTrumpSuit(), got.GetTrumpSuit())
	assert.Equal(t, src.GetDealerIdx(), got.GetDealerIdx())
	assert.Equal(t, GanjifaPlayerCnt, got.GetPlayerCnt())
	for i := 0; i < GanjifaPlayerCnt; i++ {
		assert.Equal(t, GanjifaHandSize, got.GetPlayer(i).GetCardsSize())
		assert.Equal(t, src.GetPlayer(i).GetCard(0).GetDesign(), got.GetPlayer(i).GetCard(0).GetDesign())
		assert.Equal(t, src.GetPlayer(i).GetCard(0).GetValue(), got.GetPlayer(i).GetCard(0).GetValue())
	}
}

// A weak-group trump is a legal state (designs 5-8), so borrowing the standard
// 52-card bound of 1-4 here would make half of all sessions unrestorable.
func TestGanjifa_UnmarshalAcceptsWeakGroupTrump(t *testing.T) {
	src := NewDefaultGanjifa()
	src.Reset()
	for suit := 1; suit <= GanjifaSuitCnt; suit++ {
		src.trumpSuit = suit
		data, err := src.MarshalJSON()
		assert.NoError(t, err)
		var got Ganjifa
		assert.NoError(t, got.UnmarshalJSON(data), "trump suit %d must round-trip", suit)
		assert.Equal(t, suit, got.GetTrumpSuit())
	}
}

func TestGanjifa_UnmarshalRejectsBadState(t *testing.T) {
	cases := map[string]string{
		"not json":   `{`,
		"no players": `{"pl":[],"rn":1,"tn":1}`,
		"trump out of range": `{"pl":[{},{},{}],"rn":1,"tn":1,"ts":9,` +
			`"cfg":{"cd":1,"tr":3}}`,
		"trick number too high": `{"pl":[{},{},{}],"rn":1,"tn":99,"cfg":{"cd":1,"tr":3}}`,
		"nil trick card":        `{"pl":[{},{},{}],"rn":1,"tn":1,"ct":[null],"cfg":{"cd":1,"tr":3}}`,
		"invalid config":        `{"pl":[{},{},{}],"rn":1,"tn":1,"cfg":{"cd":1,"tr":0}}`,
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			var got Ganjifa
			assert.Error(t, got.UnmarshalJSON([]byte(payload)))
		})
	}
}

func TestGanjifa_SuitNamesAndGlyphs(t *testing.T) {
	seenNames := map[string]bool{}
	seenGlyphs := map[string]bool{}
	for suit := 1; suit <= GanjifaSuitCnt; suit++ {
		name, glyph := GanjifaSuitName(suit), GanjifaSuitGlyph(suit)
		assert.NotEmpty(t, name)
		assert.NotEmpty(t, glyph)
		assert.False(t, seenNames[name], "suit name %q is reused", name)
		assert.False(t, seenGlyphs[glyph], "suit glyph %q is reused", glyph)
		seenNames[name], seenGlyphs[glyph] = true, true
	}
	for _, bad := range []int{0, -1, GanjifaSuitCnt + 1} {
		assert.Empty(t, GanjifaSuitName(bad))
		assert.Empty(t, GanjifaSuitGlyph(bad))
	}
}
