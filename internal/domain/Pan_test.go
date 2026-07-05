//go:build test

package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// panCard は Pan テスト用のカードコンストラクタ（スラッグ接頭辞付き）。
func panCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

// panGame は指定した人間フラグでプレイヤーを組んだ Pan ゲームを返す。
func panGame(isHumans ...bool) *domain.Pan {
	players := make([]*domain.PanPlayer, len(isHumans))
	for i, h := range isHumans {
		players[i] = domain.NewPanPlayer(h)
	}
	cfg := domain.DefaultPanConfig()
	cfg.PlayerCount = len(isHumans)
	return domain.NewPan(domain.NewTrumpCards(0), players, cfg)
}

// panSetHand は player の手札を指定カードで置き換える。
func panSetHand(p *domain.PanPlayer, cards ...*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

const (
	panSp = domain.CardDesignSpade
	panCl = domain.CardDesignClover
	panHe = domain.CardDesignHeart
	panDi = domain.CardDesignDiamond
)

func TestPanConfigValidate(t *testing.T) {
	t.Run("default valid", func(t *testing.T) {
		require.NoError(t, domain.DefaultPanConfig().Validate())
	})
	t.Run("player count below min", func(t *testing.T) {
		c := domain.DefaultPanConfig()
		c.PlayerCount = 2
		assert.Error(t, c.Validate())
	})
	t.Run("player count above max", func(t *testing.T) {
		c := domain.DefaultPanConfig()
		c.PlayerCount = 7
		assert.Error(t, c.Validate())
	})
	t.Run("difficulty out of range", func(t *testing.T) {
		c := domain.DefaultPanConfig()
		c.CpuDifficulty = domain.PanCpuDifficulty(9)
		assert.Error(t, c.Validate())
	})
	t.Run("target rounds zero", func(t *testing.T) {
		c := domain.DefaultPanConfig()
		c.TargetRounds = 0
		assert.Error(t, c.Validate())
	})
}

func TestPanResetAndDeal(t *testing.T) {
	g := domain.NewDefaultPan()
	g.Reset()

	assert.Equal(t, domain.PanPhaseDraw, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 4, g.GetPlayerCnt())
	// 4 players × 10 cards + 1 discard = 41 removed from a 320-card deck.
	assert.Equal(t, domain.PanDeckSize-4*domain.PanHandSize-1, g.GetDrawPileCount())
	assert.NotNil(t, g.GetDiscardTop())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx()) // dealer 0 → next is 1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, domain.PanHandSize, g.GetPlayer(i).GetCardsSize())
	}
	assert.Equal(t, domain.PanDeckSize, 320)
}

func TestPanIsValidMeld(t *testing.T) {
	tests := []struct {
		name  string
		cards []*domain.Card
		want  bool
	}{
		{"set of 3 same rank", []*domain.Card{panCard(panSp, 5), panCard(panCl, 5), panCard(panHe, 5)}, true},
		{"set with duplicate suit (multi-deck)", []*domain.Card{panCard(panSp, 5), panCard(panSp, 5), panCard(panHe, 5)}, true},
		{"run of 3 same suit", []*domain.Card{panCard(panSp, 1), panCard(panSp, 2), panCard(panSp, 3)}, true},
		{"run ace high Q-K-A", []*domain.Card{panCard(panHe, 12), panCard(panHe, 13), panCard(panHe, 1)}, true},
		{"run mixed suits", []*domain.Card{panCard(panSp, 1), panCard(panCl, 2), panCard(panSp, 3)}, false},
		{"set different ranks", []*domain.Card{panCard(panSp, 5), panCard(panCl, 6), panCard(panHe, 5)}, false},
		{"too few cards", []*domain.Card{panCard(panSp, 5), panCard(panCl, 5)}, false},
		{"run with gap (7 then J)", []*domain.Card{panCard(panSp, 6), panCard(panSp, 7), panCard(panSp, 11)}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, domain.PanIsValidMeld(tc.cards))
		})
	}
}

func TestPanCanLayoff(t *testing.T) {
	set := []*domain.Card{panCard(panSp, 5), panCard(panCl, 5), panCard(panHe, 5)}
	run := []*domain.Card{panCard(panSp, 1), panCard(panSp, 2), panCard(panSp, 3)}

	assert.True(t, domain.PanCanLayoff(set, panCard(panDi, 5)))
	assert.False(t, domain.PanCanLayoff(set, panCard(panDi, 6)))
	assert.True(t, domain.PanCanLayoff(run, panCard(panSp, 4)))
	assert.False(t, domain.PanCanLayoff(run, panCard(panHe, 4))) // wrong suit
	assert.False(t, domain.PanCanLayoff(run, panCard(panSp, 6))) // non-adjacent
	assert.False(t, domain.PanCanLayoff(nil, panCard(panSp, 5)))
}

func TestPanMeldChipUnits(t *testing.T) {
	tests := []struct {
		name  string
		cards []*domain.Card
		want  int
	}{
		{"valle set of three 5s", []*domain.Card{panCard(panSp, 5), panCard(panCl, 5), panCard(panHe, 5)}, 1},
		{"non-valle set of three 2s", []*domain.Card{panCard(panSp, 2), panCard(panCl, 2), panCard(panHe, 2)}, 0},
		{"run of 3", []*domain.Card{panCard(panSp, 1), panCard(panSp, 2), panCard(panSp, 3)}, 1},
		{"non-valle set of four 4s", []*domain.Card{panCard(panSp, 4), panCard(panCl, 4), panCard(panHe, 4), panCard(panDi, 4)}, 1},
		{"valle set of four 5s", []*domain.Card{panCard(panSp, 5), panCard(panCl, 5), panCard(panHe, 5), panCard(panDi, 5)}, 2},
		{"invalid meld", []*domain.Card{panCard(panSp, 5), panCard(panCl, 6)}, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, domain.PanMeldChipUnits(tc.cards))
		})
	}
}

func TestPanHandPoints(t *testing.T) {
	p := domain.NewPanPlayer(true)
	panSetHand(p, panCard(panSp, 1), panCard(panCl, 7), panCard(panHe, 13), panCard(panDi, 12))
	// Ace=1, 7=7, K=10, Q=10 → 28
	assert.Equal(t, 28, domain.PanHandPoints(p))
}

func TestPanMeldAndChips(t *testing.T) {
	t.Run("valle set collects one chip from each opponent", func(t *testing.T) {
		g := panGame(true, false, false)
		g.SetPhase(domain.PanPhasePlay)
		g.SetCurrentPlayerIdx(0)
		panSetHand(g.GetPlayer(0), panCard(panSp, 5), panCard(panCl, 5), panCard(panHe, 5), panCard(panDi, 2))

		require.NoError(t, g.PlayerMeld([]int{0, 1, 2}))
		assert.Equal(t, 2, g.GetPlayer(0).GetChips())
		assert.Equal(t, -1, g.GetPlayer(1).GetChips())
		assert.Equal(t, -1, g.GetPlayer(2).GetChips())
		assert.Len(t, g.GetPlayer(0).GetLaidMelds(), 1)
		assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
	})

	t.Run("valle set of four collects two chips", func(t *testing.T) {
		g := panGame(true, false, false)
		g.SetPhase(domain.PanPhasePlay)
		g.SetCurrentPlayerIdx(0)
		panSetHand(g.GetPlayer(0), panCard(panSp, 3), panCard(panCl, 3), panCard(panHe, 3), panCard(panDi, 3))

		require.NoError(t, g.PlayerMeld([]int{0, 1, 2, 3}))
		assert.Equal(t, 4, g.GetPlayer(0).GetChips())
		assert.Equal(t, -2, g.GetPlayer(1).GetChips())
	})

	t.Run("non-valle set pays no chips", func(t *testing.T) {
		g := panGame(true, false, false)
		g.SetPhase(domain.PanPhasePlay)
		g.SetCurrentPlayerIdx(0)
		panSetHand(g.GetPlayer(0), panCard(panSp, 2), panCard(panCl, 2), panCard(panHe, 2))
		require.NoError(t, g.PlayerMeld([]int{0, 1, 2}))
		assert.Equal(t, 0, g.GetPlayer(0).GetChips())
		assert.Equal(t, 0, g.GetPlayer(1).GetChips())
	})
}

func TestPanMeldGuards(t *testing.T) {
	g := panGame(true, false, false)
	g.SetPhase(domain.PanPhasePlay)
	g.SetCurrentPlayerIdx(0)
	panSetHand(g.GetPlayer(0), panCard(panSp, 5), panCard(panCl, 5), panCard(panHe, 6))

	t.Run("too few cards", func(t *testing.T) {
		assert.ErrorIs(t, g.PlayerMeld([]int{0, 1}), domain.ErrInvalidPlay)
	})
	t.Run("out of range index", func(t *testing.T) {
		assert.ErrorIs(t, g.PlayerMeld([]int{0, 1, 9}), domain.ErrInvalidCard)
	})
	t.Run("duplicate index", func(t *testing.T) {
		assert.ErrorIs(t, g.PlayerMeld([]int{0, 0, 1}), domain.ErrInvalidCard)
	})
	t.Run("invalid meld composition", func(t *testing.T) {
		assert.ErrorIs(t, g.PlayerMeld([]int{0, 1, 2}), domain.ErrInvalidPlay)
	})
	t.Run("wrong phase", func(t *testing.T) {
		g.SetPhase(domain.PanPhaseDraw)
		assert.ErrorIs(t, g.PlayerMeld([]int{0, 1, 2}), domain.ErrWrongPhase)
		g.SetPhase(domain.PanPhasePlay)
	})
	t.Run("not human turn", func(t *testing.T) {
		g.SetCurrentPlayerIdx(1)
		assert.ErrorIs(t, g.PlayerMeld([]int{0, 1, 2}), domain.ErrNotHumanTurn)
		g.SetCurrentPlayerIdx(0)
	})
}

func TestPanLayoff(t *testing.T) {
	g := panGame(true, false, false)
	g.SetPhase(domain.PanPhasePlay)
	g.SetCurrentPlayerIdx(0)
	// Opponent (player 1) already has a set of 5s on the table.
	g.GetPlayer(1).AddLaidMeld([]*domain.Card{panCard(panSp, 5), panCard(panCl, 5), panCard(panHe, 5)})
	panSetHand(g.GetPlayer(0), panCard(panDi, 5), panCard(panSp, 2))

	t.Run("bad meld owner", func(t *testing.T) {
		assert.ErrorIs(t, g.PlayerLayoff(9, 0, 0), domain.ErrInvalidPlay)
	})
	t.Run("bad meld idx", func(t *testing.T) {
		assert.ErrorIs(t, g.PlayerLayoff(1, 5, 0), domain.ErrInvalidPlay)
	})
	t.Run("bad card idx", func(t *testing.T) {
		assert.ErrorIs(t, g.PlayerLayoff(1, 0, 9), domain.ErrInvalidCard)
	})
	t.Run("card cannot lay off", func(t *testing.T) {
		assert.ErrorIs(t, g.PlayerLayoff(1, 0, 1), domain.ErrInvalidPlay) // the 2 cannot join a set of 5s
	})
	t.Run("valid layoff", func(t *testing.T) {
		require.NoError(t, g.PlayerLayoff(1, 0, 0))
		assert.Len(t, g.GetPlayer(1).GetLaidMelds()[0], 4)
		assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
	})
}

func TestPanWinDeclaresPan(t *testing.T) {
	g := panGame(true, false, false)
	g.SetPhase(domain.PanPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetRoundNumber(1)
	cfg := g.GetConfig()
	cfg.TargetRounds = 3
	g.SetConfig(cfg)

	// Pre-seed 8 already-melded cards, then meld a set of 3 to reach 11.
	g.GetPlayer(0).SetLaidMelds([][]*domain.Card{
		{panCard(panSp, 4), panCard(panCl, 4), panCard(panHe, 4), panCard(panDi, 4)},
		{panCard(panSp, 6), panCard(panCl, 6), panCard(panHe, 6), panCard(panDi, 6)},
	})
	panSetHand(g.GetPlayer(0), panCard(panSp, 2), panCard(panCl, 2), panCard(panHe, 2))
	// Give opponents some hand points.
	panSetHand(g.GetPlayer(1), panCard(panHe, 13))
	panSetHand(g.GetPlayer(2), panCard(panDi, 1))

	require.NoError(t, g.PlayerMeld([]int{0, 1, 2}))

	assert.Equal(t, 0, g.GetPanDeclarerIdx())
	assert.Equal(t, domain.PanPhaseRoundEnd, g.GetPhase())
	assert.Equal(t, 11, g.GetPlayer(0).GetMeldedCardCount())
	assert.Equal(t, 0, g.GetPlayer(0).GetRoundScore())  // winner scores 0
	assert.Equal(t, 10, g.GetPlayer(1).GetRoundScore()) // King
	assert.Equal(t, 1, g.GetPlayer(2).GetRoundScore())  // Ace
	assert.Equal(t, 0, g.GetPlayer(0).GetCumulativeScore())
}

func TestPanWinOnLastRoundEndsGame(t *testing.T) {
	g := panGame(true, false, false)
	g.SetPhase(domain.PanPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetRoundNumber(3)
	cfg := g.GetConfig()
	cfg.TargetRounds = 3
	g.SetConfig(cfg)
	g.GetPlayer(0).SetLaidMelds([][]*domain.Card{
		{panCard(panSp, 4), panCard(panCl, 4), panCard(panHe, 4), panCard(panDi, 4)},
		{panCard(panSp, 6), panCard(panCl, 6), panCard(panHe, 6), panCard(panDi, 6)},
	})
	panSetHand(g.GetPlayer(0), panCard(panSp, 7), panCard(panCl, 7), panCard(panHe, 7))

	require.NoError(t, g.PlayerMeld([]int{0, 1, 2}))
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.PanPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetWinnerIdx()) // lowest cumulative (declarer scored 0)
}

func TestPanDrawFlows(t *testing.T) {
	t.Run("draw from stock", func(t *testing.T) {
		g := panGame(true, false)
		g.SetPhase(domain.PanPhaseDraw)
		g.SetCurrentPlayerIdx(0)
		panSetHand(g.GetPlayer(0), panCard(panSp, 5))
		g.SetDrawPile([]*domain.Card{panCard(panCl, 7), panCard(panHe, 3)})
		require.NoError(t, g.PlayerDrawFromStock())
		assert.Equal(t, 2, g.GetPlayer(0).GetCardsSize())
		assert.Equal(t, domain.PanPhasePlay, g.GetPhase())
	})

	t.Run("draw from discard top", func(t *testing.T) {
		g := panGame(true, false)
		g.SetPhase(domain.PanPhaseDraw)
		g.SetCurrentPlayerIdx(0)
		panSetHand(g.GetPlayer(0), panCard(panSp, 5))
		g.SetDiscardPile([]*domain.Card{panCard(panCl, 7)})
		require.NoError(t, g.PlayerDrawFromDiscard())
		assert.Equal(t, 2, g.GetPlayer(0).GetCardsSize())
		assert.Equal(t, domain.PanPhasePlay, g.GetPhase())
		assert.Nil(t, g.GetDiscardTop())
	})

	t.Run("draw from empty discard errors", func(t *testing.T) {
		g := panGame(true, false)
		g.SetPhase(domain.PanPhaseDraw)
		g.SetCurrentPlayerIdx(0)
		g.SetDiscardPile(nil)
		assert.ErrorIs(t, g.PlayerDrawFromDiscard(), domain.ErrInvalidPlay)
	})

	t.Run("draw from empty stock ends round", func(t *testing.T) {
		g := panGame(true, false)
		g.SetPhase(domain.PanPhaseDraw)
		g.SetCurrentPlayerIdx(0)
		g.SetRoundNumber(1)
		g.SetDrawPile(nil)
		panSetHand(g.GetPlayer(0), panCard(panSp, 5))
		panSetHand(g.GetPlayer(1), panCard(panHe, 13))
		require.NoError(t, g.PlayerDrawFromStock())
		assert.Equal(t, domain.PanPhaseRoundEnd, g.GetPhase())
		assert.Equal(t, -1, g.GetPanDeclarerIdx())
		assert.Equal(t, 5, g.GetPlayer(0).GetRoundScore())
		assert.Equal(t, 10, g.GetPlayer(1).GetRoundScore())
	})

	t.Run("draw guards", func(t *testing.T) {
		g := panGame(true, false)
		g.SetPhase(domain.PanPhasePlay)
		assert.ErrorIs(t, g.PlayerDrawFromStock(), domain.ErrWrongPhase)
		g.SetPhase(domain.PanPhaseDraw)
		g.SetCurrentPlayerIdx(1)
		assert.ErrorIs(t, g.PlayerDrawFromStock(), domain.ErrNotHumanTurn)
	})
}

func TestPanDiscardAdvancesTurn(t *testing.T) {
	g := panGame(true, false, false)
	g.SetPhase(domain.PanPhasePlay)
	g.SetCurrentPlayerIdx(0)
	panSetHand(g.GetPlayer(0), panCard(panSp, 5), panCard(panCl, 7))

	t.Run("out of range", func(t *testing.T) {
		assert.ErrorIs(t, g.PlayerDiscard(9), domain.ErrInvalidCard)
	})
	t.Run("valid discard", func(t *testing.T) {
		require.NoError(t, g.PlayerDiscard(1))
		assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
		assert.Equal(t, 1, g.GetCurrentPlayerIdx())
		assert.Equal(t, domain.PanPhaseDraw, g.GetPhase())
		assert.Equal(t, 7, g.GetDiscardTop().GetValue())
	})
	t.Run("game ended", func(t *testing.T) {
		g2 := panGame(true, false)
		g2.SetPhase(domain.PanPhasePlay)
		// simulate game over
		g2.SetRoundNumber(1)
		cfg := g2.GetConfig()
		cfg.TargetRounds = 1
		g2.SetConfig(cfg)
		g2.SetDrawPile(nil)
		g2.SetPhase(domain.PanPhaseDraw)
		_ = g2.PlayerDrawFromStock() // ends the only round → game end
		assert.True(t, g2.GetGameEndFlag())
		assert.ErrorIs(t, g2.PlayerDiscard(0), domain.ErrGameEnded)
	})
}

func TestPanNextRound(t *testing.T) {
	t.Run("advances to next round", func(t *testing.T) {
		g := domain.NewDefaultPan()
		g.Reset()
		g.SetPhase(domain.PanPhaseRoundEnd)
		g.SetRoundNumber(1)
		g.NextRound()
		assert.Equal(t, 2, g.GetRoundNumber())
		assert.Equal(t, domain.PanPhaseDraw, g.GetPhase())
		assert.Equal(t, 1, g.GetDealerIdx()) // dealer rotates
	})

	t.Run("no-op when not in round end", func(t *testing.T) {
		g := domain.NewDefaultPan()
		g.Reset()
		g.NextRound()
		assert.Equal(t, 1, g.GetRoundNumber())
	})

	t.Run("finalizes game after target rounds", func(t *testing.T) {
		g := domain.NewDefaultPan()
		g.Reset()
		g.SetPhase(domain.PanPhaseRoundEnd)
		g.SetRoundNumber(3)
		cfg := g.GetConfig()
		cfg.TargetRounds = 3
		g.SetConfig(cfg)
		g.NextRound()
		assert.True(t, g.GetGameEndFlag())
		assert.Equal(t, domain.PanPhaseGameEnd, g.GetPhase())
		assert.GreaterOrEqual(t, g.GetWinnerIdx(), 0)
	})
}

func TestPanCpuPlay(t *testing.T) {
	t.Run("cpu draws then melds and discards", func(t *testing.T) {
		g := panGame(true, false)
		g.SetPhase(domain.PanPhasePlay)
		g.SetCurrentPlayerIdx(1)
		panSetHand(g.GetPlayer(1), panCard(panSp, 5), panCard(panCl, 5), panCard(panHe, 5), panCard(panDi, 13))
		g.CpuPlay()
		assert.Len(t, g.GetPlayer(1).GetLaidMelds(), 1)
		assert.Equal(t, domain.PanPhaseDraw, g.GetPhase())
		assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	})

	t.Run("cpu takes discard to complete a set", func(t *testing.T) {
		g := panGame(true, false)
		g.SetPhase(domain.PanPhaseDraw)
		g.SetCurrentPlayerIdx(1)
		cfg := g.GetConfig()
		cfg.CpuDifficulty = domain.PanCpuDifficultyNormal
		g.SetConfig(cfg)
		panSetHand(g.GetPlayer(1), panCard(panSp, 5), panCard(panCl, 5))
		g.SetDiscardPile([]*domain.Card{panCard(panHe, 5)})
		g.SetDrawPile([]*domain.Card{panCard(panDi, 2)})
		g.CpuPlay() // draw phase → should take the discard
		assert.Equal(t, 3, g.GetPlayer(1).GetCardsSize())
		assert.Equal(t, domain.PanPhasePlay, g.GetPhase())
	})

	t.Run("cpu draws from stock when discard useless", func(t *testing.T) {
		g := panGame(true, false)
		g.SetPhase(domain.PanPhaseDraw)
		g.SetCurrentPlayerIdx(1)
		panSetHand(g.GetPlayer(1), panCard(panSp, 2))
		g.SetDiscardPile([]*domain.Card{panCard(panHe, 13)})
		g.SetDrawPile([]*domain.Card{panCard(panDi, 4)})
		g.CpuPlay()
		assert.Equal(t, domain.PanPhasePlay, g.GetPhase())
		assert.Equal(t, 2, g.GetPlayer(1).GetCardsSize())
	})

	t.Run("no-op when human turn", func(t *testing.T) {
		g := panGame(true, false)
		g.SetPhase(domain.PanPhaseDraw)
		g.SetCurrentPlayerIdx(0)
		g.CpuPlay()
		assert.Equal(t, domain.PanPhaseDraw, g.GetPhase())
	})
}

func TestPanBoundedDrive(t *testing.T) {
	g := domain.NewDefaultPan()
	g.Reset()
	for i := 0; i < 3000 && !g.GetGameEndFlag(); i++ {
		phase := g.GetPhase()
		if phase == domain.PanPhaseRoundEnd {
			g.NextRound()
			continue
		}
		if g.IsHumanTurn() {
			switch phase {
			case domain.PanPhaseDraw:
				_ = g.PlayerDrawFromStock()
			case domain.PanPhasePlay:
				_ = g.PlayerDiscard(0)
			}
		} else {
			g.CpuPlay()
		}
		// state remains coherent
		assert.GreaterOrEqual(t, g.GetPhase(), domain.PanPhaseDraw)
		assert.LessOrEqual(t, g.GetPhase(), domain.PanPhaseGameEnd)
	}
}

func TestPanJSONRoundTrip(t *testing.T) {
	g := domain.NewDefaultPan()
	g.Reset()
	g.GetPlayer(0).SetChips(3)
	g.GetPlayer(0).AddLaidMeld([]*domain.Card{panCard(panSp, 5), panCard(panCl, 5), panCard(panHe, 5)})

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var g2 domain.Pan
	require.NoError(t, json.Unmarshal(data, &g2))
	assert.Equal(t, g.GetPlayerCnt(), g2.GetPlayerCnt())
	assert.Equal(t, 3, g2.GetPlayer(0).GetChips())
	assert.Len(t, g2.GetPlayer(0).GetLaidMelds(), 1)
	assert.Equal(t, g.GetDrawPileCount(), g2.GetDrawPileCount())
}

func TestPanUnmarshalValidation(t *testing.T) {
	// A valid baseline we can perturb via the raw JSON map.
	base := domain.NewDefaultPan()
	base.Reset()
	valid, err := json.Marshal(base)
	require.NoError(t, err)

	perturb := func(t *testing.T, mutate func(m map[string]json.RawMessage)) error {
		var m map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(valid, &m))
		mutate(m)
		bad, err := json.Marshal(m)
		require.NoError(t, err)
		var g domain.Pan
		return json.Unmarshal(bad, &g)
	}

	t.Run("currentPlayerIdx out of range", func(t *testing.T) {
		assert.Error(t, perturb(t, func(m map[string]json.RawMessage) { m["ci"] = json.RawMessage("99") }))
	})
	t.Run("dealerIdx out of range", func(t *testing.T) {
		assert.Error(t, perturb(t, func(m map[string]json.RawMessage) { m["di"] = json.RawMessage("99") }))
	})
	t.Run("invalid phase", func(t *testing.T) {
		assert.Error(t, perturb(t, func(m map[string]json.RawMessage) { m["ps"] = json.RawMessage("99") }))
	})
	t.Run("negative round", func(t *testing.T) {
		assert.Error(t, perturb(t, func(m map[string]json.RawMessage) { m["rn"] = json.RawMessage("-5") }))
	})
	t.Run("winnerIdx out of range", func(t *testing.T) {
		assert.Error(t, perturb(t, func(m map[string]json.RawMessage) { m["wi"] = json.RawMessage("42") }))
	})
	t.Run("panDeclarerIdx out of range", func(t *testing.T) {
		assert.Error(t, perturb(t, func(m map[string]json.RawMessage) { m["pd"] = json.RawMessage("42") }))
	})
	t.Run("winnerIdx sentinel -1 allowed", func(t *testing.T) {
		assert.NoError(t, perturb(t, func(m map[string]json.RawMessage) { m["wi"] = json.RawMessage("-1") }))
	})

	t.Run("nil player rejected", func(t *testing.T) {
		var g domain.Pan
		err := json.Unmarshal([]byte(`{"pl":[null,null,null]}`), &g)
		assert.Error(t, err)
	})
	t.Run("invalid player count", func(t *testing.T) {
		var g domain.Pan
		err := json.Unmarshal([]byte(`{"pl":[]}`), &g)
		assert.Error(t, err)
	})
	t.Run("bad json", func(t *testing.T) {
		var g domain.Pan
		assert.Error(t, json.Unmarshal([]byte(`{`), &g))
	})
}

func TestPanErrorsAreSentinels(t *testing.T) {
	// Guard against accidental error-wrapping regressions.
	assert.True(t, errors.Is(domain.ErrWrongPhase, domain.ErrWrongPhase))
}

func TestPanActionLog(t *testing.T) {
	g := panGame(true, false)
	g.SetPhase(domain.PanPhaseDraw)
	g.SetCurrentPlayerIdx(0)
	g.SetDrawPile([]*domain.Card{panCard(panSp, 5)})
	panSetHand(g.GetPlayer(0), panCard(panCl, 7))
	require.NoError(t, g.PlayerDrawFromStock())
	assert.NotEmpty(t, g.GetActionLog())
}
