//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// cqCard はテスト用のカード生成ヘルパー (design, value は plain int)。
func cqCard(d, v int) *domain.Card {
	return domain.NewCard(d, v, false)
}

// cqSetHand はプレイヤーの手札を明示的に設定する (Reset 後に AddCard)。
func cqSetHand(p *domain.ConquianPlayer, cards ...*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// newTestConquian は人間2人構成の Conquian を返す (CPU 自動進行を避けて決定的にテストするため)。
func newTestConquian() *domain.Conquian {
	players := []*domain.ConquianPlayer{
		domain.NewConquianPlayer(true),
		domain.NewConquianPlayer(true),
	}
	return domain.NewConquian(players, domain.DefaultConquianConfig())
}

func TestConquianConfig_Validate(t *testing.T) {
	t.Run("default is valid", func(t *testing.T) {
		assert.NoError(t, domain.DefaultConquianConfig().Validate())
	})
	t.Run("difficulty out of range", func(t *testing.T) {
		c := domain.DefaultConquianConfig()
		c.CpuDifficulty = domain.ConquianCpuDifficulty(99)
		assert.Error(t, c.Validate())
	})
	t.Run("target wins below 1", func(t *testing.T) {
		c := domain.DefaultConquianConfig()
		c.TargetWins = 0
		assert.Error(t, c.Validate())
	})
	t.Run("target wins above 100", func(t *testing.T) {
		c := domain.DefaultConquianConfig()
		c.TargetWins = 101
		assert.Error(t, c.Validate())
	})
}

func TestConquian_DeckIs40(t *testing.T) {
	g := domain.NewDefaultConquian()
	g.Reset()
	// 10 + 10 = 20 dealt, stock holds the rest. Total must be 40.
	total := g.GetPlayer(0).GetCardsSize() + g.GetPlayer(1).GetCardsSize() + g.GetDrawPileCount()
	assert.Equal(t, 40, total)
	assert.Equal(t, 10, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 10, g.GetPlayer(1).GetCardsSize())
	assert.Equal(t, 20, g.GetDrawPileCount())
	// No 8/9/10 anywhere in stock or hands.
	check := func(c *domain.Card) {
		v := c.GetValue()
		assert.NotContains(t, []int{8, 9, 10}, v)
	}
	for i := 0; i < 2; i++ {
		p := g.GetPlayer(i)
		for j := 0; j < p.GetCardsSize(); j++ {
			check(p.GetCard(j))
		}
	}
	for _, c := range g.GetDiscardPile() {
		check(c)
	}
}

func TestConquian_ResetInitialState(t *testing.T) {
	g := domain.NewDefaultConquian()
	g.Reset()
	assert.Equal(t, domain.ConquianPhaseDraw, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Nil(t, g.GetDiscardTop())
	assert.True(t, g.IsHumanTurn())
}

func TestConquian_DrawFromStock(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	before := g.GetPlayer(0).GetCardsSize()
	stockBefore := g.GetDrawPileCount()
	require.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, before+1, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, stockBefore-1, g.GetDrawPileCount())
	assert.Equal(t, domain.ConquianPhaseMeld, g.GetPhase())
}

func TestConquian_DrawFromStock_WrongPhase(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	g.SetPhase(domain.ConquianPhaseMeld)
	assert.ErrorIs(t, g.PlayerDrawFromStock(), domain.ErrWrongPhase)
}

func TestConquian_DrawFromStock_StockEmptyEndsRound(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	g.SetStock(nil)
	require.NoError(t, g.PlayerDrawFromStock())
	// stock empty + cannot proceed → game ends (draw).
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.ConquianPhaseGameEnd, g.GetPhase())
}

func TestConquian_DrawFromDiscard_ForcedUse_Illegal(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	// Set hand that cannot use the discard top in any meld.
	cqSetHand(g.GetPlayer(0),
		cqCard(domain.CardDesignSpade, 1),
		cqCard(domain.CardDesignHeart, 4),
		cqCard(domain.CardDesignDiamond, 7),
	)
	// Discard top: a King with no support in hand → cannot be melded.
	g.SetDiscardPile([]*domain.Card{cqCard(domain.CardDesignClover, 13)})
	g.SetPhase(domain.ConquianPhaseDraw)
	g.SetCurrentPlayerIdx(0)
	err := g.PlayerDrawFromDiscard()
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	// hand unchanged
	assert.Equal(t, 3, g.GetPlayer(0).GetCardsSize())
}

func TestConquian_DrawFromDiscard_ForcedUse_Legal(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	// Hand has two more spades to make a set with the discard.
	cqSetHand(g.GetPlayer(0),
		cqCard(domain.CardDesignHeart, 5),
		cqCard(domain.CardDesignDiamond, 5),
		cqCard(domain.CardDesignClover, 2),
	)
	g.SetDiscardPile([]*domain.Card{cqCard(domain.CardDesignSpade, 5)})
	g.SetPhase(domain.ConquianPhaseDraw)
	g.SetCurrentPlayerIdx(0)
	require.NoError(t, g.PlayerDrawFromDiscard())
	assert.Equal(t, 4, g.GetPlayer(0).GetCardsSize())
	assert.True(t, g.GetTookDiscard())
	assert.Equal(t, domain.ConquianPhaseMeld, g.GetPhase())
}

func TestConquian_DrawFromDiscard_Empty(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	g.SetDiscardPile(nil)
	assert.ErrorIs(t, g.PlayerDrawFromDiscard(), domain.ErrInvalidPlay)
}

func TestConquian_MeldSet(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	cqSetHand(g.GetPlayer(0),
		cqCard(domain.CardDesignSpade, 5),
		cqCard(domain.CardDesignHeart, 5),
		cqCard(domain.CardDesignDiamond, 5),
		cqCard(domain.CardDesignClover, 2),
	)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ConquianPhaseMeld)
	require.NoError(t, g.PlayerMeld([][]int{{0, 1, 2}}))
	assert.Len(t, g.GetPlayer(0).GetMelds(), 1)
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize()) // only the 2 of clover left
}

func TestConquian_MeldRun_7JAdjacency(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	// 6,7,J of spades: positions 6,7,8 → consecutive run (7 and J adjacent).
	cqSetHand(g.GetPlayer(0),
		cqCard(domain.CardDesignSpade, 6),
		cqCard(domain.CardDesignSpade, 7),
		cqCard(domain.CardDesignSpade, 11), // J
		cqCard(domain.CardDesignHeart, 2),
	)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ConquianPhaseMeld)
	require.NoError(t, g.PlayerMeld([][]int{{0, 1, 2}}))
	assert.Len(t, g.GetPlayer(0).GetMelds(), 1)
}

func TestConquian_MeldRun_JQK(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	cqSetHand(g.GetPlayer(0),
		cqCard(domain.CardDesignDiamond, 11), // J = pos 8
		cqCard(domain.CardDesignDiamond, 12), // Q = pos 9
		cqCard(domain.CardDesignDiamond, 13), // K = pos 10
		cqCard(domain.CardDesignHeart, 2),
	)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ConquianPhaseMeld)
	require.NoError(t, g.PlayerMeld([][]int{{0, 1, 2}}))
	assert.Len(t, g.GetPlayer(0).GetMelds(), 1)
}

func TestConquian_MeldRun_GapInvalid(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	// 6,7,Q spades: positions 6,7,9 → NOT consecutive (8 = J missing).
	cqSetHand(g.GetPlayer(0),
		cqCard(domain.CardDesignSpade, 6),
		cqCard(domain.CardDesignSpade, 7),
		cqCard(domain.CardDesignSpade, 12), // Q
	)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ConquianPhaseMeld)
	assert.ErrorIs(t, g.PlayerMeld([][]int{{0, 1, 2}}), domain.ErrInvalidPlay)
}

func TestConquian_Meld_IllegalSetDuplicateSuit(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	// Three 5s but two are spades → invalid set.
	cqSetHand(g.GetPlayer(0),
		cqCard(domain.CardDesignSpade, 5),
		cqCard(domain.CardDesignSpade, 5),
		cqCard(domain.CardDesignHeart, 5),
	)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ConquianPhaseMeld)
	assert.ErrorIs(t, g.PlayerMeld([][]int{{0, 1, 2}}), domain.ErrInvalidPlay)
}

func TestConquian_Meld_IndexOutOfRange(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	cqSetHand(g.GetPlayer(0), cqCard(domain.CardDesignSpade, 5))
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ConquianPhaseMeld)
	assert.ErrorIs(t, g.PlayerMeld([][]int{{5}}), domain.ErrInvalidCard)
}

func TestConquian_Meld_DuplicateIndex(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	cqSetHand(g.GetPlayer(0),
		cqCard(domain.CardDesignSpade, 5),
		cqCard(domain.CardDesignHeart, 5),
		cqCard(domain.CardDesignDiamond, 5),
	)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ConquianPhaseMeld)
	assert.ErrorIs(t, g.PlayerMeld([][]int{{0, 0, 1}}), domain.ErrInvalidCard)
}

func TestConquian_Meld_ForcedUse_DiscardCardNotInMeld(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	// Draw a discard that forms a set; then try to meld a DIFFERENT group → illegal.
	cqSetHand(g.GetPlayer(0),
		cqCard(domain.CardDesignHeart, 5),
		cqCard(domain.CardDesignDiamond, 5),
		// unrelated set already possible:
		cqCard(domain.CardDesignSpade, 3),
		cqCard(domain.CardDesignHeart, 3),
		cqCard(domain.CardDesignDiamond, 3),
	)
	g.SetDiscardPile([]*domain.Card{cqCard(domain.CardDesignSpade, 5)})
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ConquianPhaseDraw)
	require.NoError(t, g.PlayerDrawFromDiscard()) // now hand has the spade 5; tookDiscard=true
	// Meld the 3-set (indices for the three 3s), NOT using the spade-5.
	// After sortHand the indices change, so locate the three 3s.
	p := g.GetPlayer(0)
	var threes []int
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetValue() == 3 {
			threes = append(threes, i)
		}
	}
	require.Len(t, threes, 3)
	err := g.PlayerMeld([][]int{threes})
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
}

func TestConquian_Discard_BlockedAfterTakingDiscard(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	cqSetHand(g.GetPlayer(0),
		cqCard(domain.CardDesignHeart, 5),
		cqCard(domain.CardDesignDiamond, 5),
		cqCard(domain.CardDesignClover, 2),
	)
	g.SetDiscardPile([]*domain.Card{cqCard(domain.CardDesignSpade, 5)})
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ConquianPhaseDraw)
	require.NoError(t, g.PlayerDrawFromDiscard())
	// Cannot discard before using the taken card.
	assert.ErrorIs(t, g.PlayerDiscard(0), domain.ErrInvalidPlay)
}

func TestConquian_Discard_NormalFlow(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	cqSetHand(g.GetPlayer(0),
		cqCard(domain.CardDesignHeart, 5),
		cqCard(domain.CardDesignClover, 2),
	)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ConquianPhaseMeld)
	require.NoError(t, g.PlayerDiscard(0))
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx()) // turn advanced
	assert.Equal(t, domain.ConquianPhaseDraw, g.GetPhase())
	assert.NotNil(t, g.GetDiscardTop())
}

func TestConquian_Discard_IndexOutOfRange(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	cqSetHand(g.GetPlayer(0), cqCard(domain.CardDesignHeart, 5))
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ConquianPhaseMeld)
	assert.ErrorIs(t, g.PlayerDiscard(9), domain.ErrInvalidCard)
}

func TestConquian_GoOutWins(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	// Hand of exactly one meld of 3 → melding empties the hand → win.
	cqSetHand(g.GetPlayer(0),
		cqCard(domain.CardDesignSpade, 5),
		cqCard(domain.CardDesignHeart, 5),
		cqCard(domain.CardDesignDiamond, 5),
	)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ConquianPhaseMeld)
	require.NoError(t, g.PlayerMeld([][]int{{0, 1, 2}}))
	assert.Equal(t, 0, g.GetPlayer(0).GetCardsSize())
	// TargetWins default = 1 → match ends.
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerIdx())
	assert.Equal(t, domain.ConquianPhaseGameEnd, g.GetPhase())
}

func TestConquian_Meld_ExtendExistingMeld(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	p := g.GetPlayer(0)
	// Pre-place a set of three 5s on the table.
	p.SetMelds([][]*domain.Card{{
		cqCard(domain.CardDesignSpade, 5),
		cqCard(domain.CardDesignHeart, 5),
		cqCard(domain.CardDesignDiamond, 5),
	}})
	// Hand has the 4th 5 (clover) plus a spare so we don't go out.
	cqSetHand(p,
		cqCard(domain.CardDesignClover, 5),
		cqCard(domain.CardDesignHeart, 2),
	)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ConquianPhaseMeld)
	// locate the clover 5
	idx := -1
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetValue() == 5 {
			idx = i
		}
	}
	require.GreaterOrEqual(t, idx, 0)
	require.NoError(t, g.PlayerMeld([][]int{{idx}}))
	assert.Len(t, p.GetMelds()[0], 4)
	assert.Equal(t, 1, p.GetCardsSize())
}

func TestConquian_NextRound(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	g.SetPhase(domain.ConquianPhaseRoundEnd)
	rn := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, rn+1, g.GetRoundNumber())
	assert.Equal(t, domain.ConquianPhaseDraw, g.GetPhase())
	assert.Equal(t, 10, g.GetPlayer(0).GetCardsSize())
}

func TestConquian_NextRound_WrongPhase(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	g.SetPhase(domain.ConquianPhaseDraw)
	rn := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, rn, g.GetRoundNumber()) // no-op
}

func TestConquian_FullCpuGameTerminates(t *testing.T) {
	players := []*domain.ConquianPlayer{
		domain.NewConquianPlayer(false),
		domain.NewConquianPlayer(false),
	}
	g := domain.NewConquian(players, domain.DefaultConquianConfig())
	g.Reset()
	// Drive CPU turns until termination; bounded to avoid infinite loop.
	for i := 0; i < 100000 && !g.GetGameEndFlag(); i++ {
		phase := g.GetPhase()
		if phase == domain.ConquianPhaseRoundEnd {
			g.NextRound()
			continue
		}
		g.CpuPlay()
	}
	assert.True(t, g.GetGameEndFlag(), "full CPU game must terminate")
	assert.Equal(t, domain.ConquianPhaseGameEnd, g.GetPhase())
}

func TestConquian_CpuPlay_NoOpForHumanTurn(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	g.SetCurrentPlayerIdx(0) // human
	before := g.GetPlayer(0).GetCardsSize()
	g.CpuPlay()
	assert.Equal(t, before, g.GetPlayer(0).GetCardsSize())
}

func TestConquian_JSONRoundTrip(t *testing.T) {
	g := domain.NewDefaultConquian()
	g.Reset()
	require.NoError(t, g.PlayerDrawFromStock())

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored domain.Conquian
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetCurrentPlayerIdx(), restored.GetCurrentPlayerIdx())
	assert.Equal(t, g.GetPlayer(0).GetCardsSize(), restored.GetPlayer(0).GetCardsSize())
	assert.Equal(t, g.GetDrawPileCount(), restored.GetDrawPileCount())
}

func TestConquian_UnmarshalJSON_RejectsInvalid(t *testing.T) {
	cases := map[string]string{
		"wrong player count":     `{"pl":[{}],"cf":{"cd":1,"tw":1},"ps":0,"ci":0}`,
		"nil player element":     `{"pl":[null,null],"cf":{"cd":1,"tw":1},"ps":0,"ci":0}`,
		"phase out of range":     `{"pl":[{},{}],"cf":{"cd":1,"tw":1},"ps":99,"ci":0}`,
		"current idx out of rng": `{"pl":[{},{}],"cf":{"cd":1,"tw":1},"ps":0,"ci":5}`,
		"invalid config":         `{"pl":[{},{}],"cf":{"cd":99,"tw":1},"ps":0,"ci":0}`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			var g domain.Conquian
			assert.Error(t, json.Unmarshal([]byte(raw), &g))
		})
	}
}

func TestConquian_UnmarshalJSON_Malformed(t *testing.T) {
	var g domain.Conquian
	assert.Error(t, json.Unmarshal([]byte(`{not json`), &g))
}

func TestConquianPlayer_JSONRoundTrip(t *testing.T) {
	p := domain.NewConquianPlayer(true)
	p.AddCard(cqCard(domain.CardDesignSpade, 5))
	p.AddMeld([]*domain.Card{
		cqCard(domain.CardDesignSpade, 3),
		cqCard(domain.CardDesignHeart, 3),
		cqCard(domain.CardDesignDiamond, 3),
	})
	p.SetWins(2)

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored domain.ConquianPlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, 1, restored.GetCardsSize())
	assert.Len(t, restored.GetMelds(), 1)
	assert.Equal(t, 2, restored.GetWins())
	assert.True(t, restored.GetIsHuman())
}

// **延長先が一意とは限らない。**♠5 は「5 のセット」も「♠4-6-7 のラン」も延長できる。
// 先頭一致で決め打つと意図した側を選べない (#4837)。
func TestConquian_PlayerMeldWithTargets(t *testing.T) {
	setup := func() *domain.Conquian {
		g := newTestConquian()
		g.Reset()
		p := g.GetPlayer(0)
		p.SetMelds([][]*domain.Card{
			{
				cqCard(domain.CardDesignHeart, 5),
				cqCard(domain.CardDesignClover, 5),
				cqCard(domain.CardDesignDiamond, 5),
			},
			{
				cqCard(domain.CardDesignSpade, 4),
				cqCard(domain.CardDesignSpade, 6),
				cqCard(domain.CardDesignSpade, 7),
			},
		})
		cqSetHand(p, cqCard(domain.CardDesignSpade, 5), cqCard(domain.CardDesignHeart, 2))
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.ConquianPhaseMeld)
		return g
	}

	g := setup()
	// ♠5 はセット (4 枚目の 5) にもラン (♠4-5-6-7) にも足せる。
	assert.Equal(t, []int{0, 1},
		g.GetExtendableMeldIndices(0, cqCard(domain.CardDesignSpade, 5)))

	// 指定なしなら従来どおり先頭 (セット)。
	require.NoError(t, g.PlayerMeld([][]int{{0}}))
	assert.Len(t, g.GetPlayer(0).GetMelds()[0], 4)
	assert.Len(t, g.GetPlayer(0).GetMelds()[1], 3)

	// 指定すればラン側に付く。
	g2 := setup()
	require.NoError(t, g2.PlayerMeldWithTargets([][]int{{0}}, []int{1}))
	assert.Len(t, g2.GetPlayer(0).GetMelds()[1], 4)
	assert.Len(t, g2.GetPlayer(0).GetMelds()[0], 3, "セットは増えていない")

	// 足せないメルドを指定したら従来どおり最初に足せるメルドへ。
	g3 := setup()
	require.NoError(t, g3.PlayerMeldWithTargets([][]int{{0}}, []int{99}))
	assert.Len(t, g3.GetPlayer(0).GetMelds()[0], 4)

	// 範囲外・nil のカードは候補なし。
	assert.Nil(t, g.GetExtendableMeldIndices(99, cqCard(domain.CardDesignSpade, 5)))
	assert.Nil(t, g.GetExtendableMeldIndices(0, nil))
}
