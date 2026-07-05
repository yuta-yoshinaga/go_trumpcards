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

// anacondaCard は指定デザイン・値のカードを生成するテストヘルパー。
func anacondaCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// anacondaSetKept は player の手札を任意枚数に差し替える。
func anacondaSetKept(p *domain.AnacondaPlayer, cs ...*domain.Card) {
	p.ClearHand()
	for _, c := range cs {
		p.AddCard(c)
	}
}

// anacondaWeakHand は「チェックしか誘発しない」弱いハイカードの 5 枚を返す。
// 固定 4 枚 {2,3,4,6} + base で、base ∈ {8,9,10,11} ならペア・ストレート・フラッシュを作らず
// ハイカードが Q 未満 (CPU がレイズせず、必要なら降りる) になる。
func anacondaWeakHand(base int) []*domain.Card {
	return []*domain.Card{
		anacondaCard(domain.CardDesignSpade, 2),
		anacondaCard(domain.CardDesignHeart, 3),
		anacondaCard(domain.CardDesignClover, 4),
		anacondaCard(domain.CardDesignDiamond, 6),
		anacondaCard(domain.CardDesignSpade, base),
	}
}

// --- 手役評価 ---

func TestAnacondaEval_Categories(t *testing.T) {
	sp, he, cl, di := domain.CardDesignSpade, domain.CardDesignHeart, domain.CardDesignClover, domain.CardDesignDiamond

	cases := []struct {
		name string
		hand []*domain.Card
		cat  int
		tb   []int
	}{
		{"straight flush", []*domain.Card{anacondaCard(sp, 5), anacondaCard(sp, 6), anacondaCard(sp, 7), anacondaCard(sp, 8), anacondaCard(sp, 9)}, domain.AnacondaStraightFlush, []int{9}},
		{"wheel straight flush", []*domain.Card{anacondaCard(sp, 1), anacondaCard(sp, 2), anacondaCard(sp, 3), anacondaCard(sp, 4), anacondaCard(sp, 5)}, domain.AnacondaStraightFlush, []int{5}},
		{"four of a kind", []*domain.Card{anacondaCard(sp, 8), anacondaCard(he, 8), anacondaCard(cl, 8), anacondaCard(di, 8), anacondaCard(sp, 13)}, domain.AnacondaFourKind, []int{8, 13}},
		{"full house", []*domain.Card{anacondaCard(sp, 8), anacondaCard(he, 8), anacondaCard(cl, 8), anacondaCard(di, 13), anacondaCard(sp, 13)}, domain.AnacondaFullHouse, []int{8, 13}},
		{"flush", []*domain.Card{anacondaCard(sp, 2), anacondaCard(sp, 5), anacondaCard(sp, 9), anacondaCard(sp, 11), anacondaCard(sp, 13)}, domain.AnacondaFlush, []int{13, 11, 9, 5, 2}},
		{"straight", []*domain.Card{anacondaCard(sp, 5), anacondaCard(he, 6), anacondaCard(sp, 7), anacondaCard(he, 8), anacondaCard(sp, 9)}, domain.AnacondaStraight, []int{9}},
		{"wheel straight", []*domain.Card{anacondaCard(sp, 1), anacondaCard(he, 2), anacondaCard(sp, 3), anacondaCard(he, 4), anacondaCard(sp, 5)}, domain.AnacondaStraight, []int{5}},
		{"broadway straight", []*domain.Card{anacondaCard(sp, 10), anacondaCard(he, 11), anacondaCard(sp, 12), anacondaCard(he, 13), anacondaCard(sp, 1)}, domain.AnacondaStraight, []int{14}},
		{"three of a kind", []*domain.Card{anacondaCard(sp, 8), anacondaCard(he, 8), anacondaCard(cl, 8), anacondaCard(di, 13), anacondaCard(sp, 2)}, domain.AnacondaThreeKind, []int{8, 13, 2}},
		{"two pair", []*domain.Card{anacondaCard(sp, 8), anacondaCard(he, 8), anacondaCard(cl, 13), anacondaCard(di, 13), anacondaCard(sp, 2)}, domain.AnacondaTwoPair, []int{13, 8, 2}},
		{"one pair", []*domain.Card{anacondaCard(sp, 8), anacondaCard(he, 8), anacondaCard(cl, 13), anacondaCard(di, 5), anacondaCard(sp, 2)}, domain.AnacondaOnePair, []int{8, 13, 5, 2}},
		{"high card", []*domain.Card{anacondaCard(sp, 13), anacondaCard(he, 11), anacondaCard(cl, 9), anacondaCard(di, 5), anacondaCard(sp, 2)}, domain.AnacondaHighCard, []int{13, 11, 9, 5, 2}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cat, tb := domain.AnacondaEval(tc.hand)
			assert.Equal(t, tc.cat, cat)
			assert.Equal(t, tc.tb, tb)
		})
	}
}

func TestAnacondaEval_Invalid(t *testing.T) {
	cat, tb := domain.AnacondaEval([]*domain.Card{anacondaCard(domain.CardDesignSpade, 5)})
	assert.Equal(t, -1, cat)
	assert.Nil(t, tb)

	cat2, tb2 := domain.AnacondaEval([]*domain.Card{
		anacondaCard(domain.CardDesignSpade, 5), anacondaCard(domain.CardDesignSpade, 6),
		anacondaCard(domain.CardDesignSpade, 7), anacondaCard(domain.CardDesignSpade, 8), nil,
	})
	assert.Equal(t, -1, cat2)
	assert.Nil(t, tb2)
}

func TestAnacondaCompare(t *testing.T) {
	sp := domain.CardDesignSpade
	sf, sfTb := domain.AnacondaEval([]*domain.Card{anacondaCard(sp, 5), anacondaCard(sp, 6), anacondaCard(sp, 7), anacondaCard(sp, 8), anacondaCard(sp, 9)})
	quad, quadTb := domain.AnacondaEval([]*domain.Card{anacondaCard(sp, 8), anacondaCard(domain.CardDesignHeart, 8), anacondaCard(domain.CardDesignClover, 8), anacondaCard(domain.CardDesignDiamond, 8), anacondaCard(sp, 13)})

	assert.Equal(t, 1, domain.AnacondaCompare(sf, sfTb, quad, quadTb))
	assert.Equal(t, -1, domain.AnacondaCompare(quad, quadTb, sf, sfTb))
	assert.Equal(t, 0, domain.AnacondaCompare(sf, sfTb, sf, sfTb))

	// Same category, tiebreak decides (A-high pair beats K-high pair).
	pA, tbA := domain.AnacondaEval([]*domain.Card{anacondaCard(sp, 1), anacondaCard(domain.CardDesignHeart, 1), anacondaCard(sp, 5), anacondaCard(sp, 7), anacondaCard(sp, 9)})
	pK, tbK := domain.AnacondaEval([]*domain.Card{anacondaCard(sp, 13), anacondaCard(domain.CardDesignHeart, 13), anacondaCard(sp, 5), anacondaCard(sp, 7), anacondaCard(sp, 9)})
	assert.Equal(t, 1, domain.AnacondaCompare(pA, tbA, pK, tbK))
}

// --- 設定 ---

func TestAnacondaConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultAnacondaConfig().Validate())
	assert.Error(t, domain.AnacondaConfig{PlayerCount: 2, Ante: 10, StartingChips: 200, TargetRounds: 10}.Validate())
	assert.Error(t, domain.AnacondaConfig{PlayerCount: 8, Ante: 10, StartingChips: 200, TargetRounds: 10}.Validate())
	assert.Error(t, domain.AnacondaConfig{PlayerCount: 4, Ante: 0, StartingChips: 200, TargetRounds: 10}.Validate())
	assert.Error(t, domain.AnacondaConfig{PlayerCount: 4, Ante: 10, StartingChips: 1, TargetRounds: 10}.Validate())
	assert.Error(t, domain.AnacondaConfig{PlayerCount: 4, Ante: 50, StartingChips: 10, TargetRounds: 10}.Validate())
	assert.Error(t, domain.AnacondaConfig{PlayerCount: 4, Ante: 10, StartingChips: 200, TargetRounds: 0}.Validate())
}

// --- 進行 ---

func TestAnaconda_ResetDealsRound(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	assert.Equal(t, domain.AnacondaPhasePass, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, domain.AnacondaPassStart, g.GetPassCount())
	cfg := g.GetConfig()
	assert.Equal(t, cfg.Ante*cfg.PlayerCount, g.GetPot())
	assert.Equal(t, cfg.StartingChips-cfg.Ante, g.GetChips())
	assert.Equal(t, domain.AnacondaDealSize, g.GetPlayer(0).GetCardsSize())
}

func TestAnaconda_PassPhase_ThreeTwoOne(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	// 3 → 2 → 1、その後セットフェーズ。各パス後も手札は 7 枚。
	require.NoError(t, g.Pass([]int{0, 1, 2}))
	assert.Equal(t, 2, g.GetPassCount())
	assert.Equal(t, domain.AnacondaDealSize, g.GetPlayer(0).GetCardsSize())

	require.NoError(t, g.Pass([]int{0, 1}))
	assert.Equal(t, 1, g.GetPassCount())

	require.NoError(t, g.Pass([]int{0}))
	assert.Equal(t, domain.AnacondaPhaseSet, g.GetPhase())
	assert.Equal(t, domain.AnacondaDealSize, g.GetPlayer(0).GetCardsSize())
}

func TestAnaconda_Pass_Errors(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	// wrong count
	assert.True(t, errors.Is(g.Pass([]int{0, 1}), domain.ErrInvalidPlay))
	// out of range
	assert.True(t, errors.Is(g.Pass([]int{0, 1, 99}), domain.ErrInvalidPlay))
	// duplicate
	assert.True(t, errors.Is(g.Pass([]int{0, 0, 1}), domain.ErrInvalidPlay))

	// wrong phase
	g.SetPhase(domain.AnacondaPhaseSet)
	assert.True(t, errors.Is(g.Pass([]int{0, 1, 2}), domain.ErrWrongPhase))
}

func TestAnaconda_SetPhase_KeepFive(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	require.NoError(t, g.Pass([]int{0, 1, 2}))
	require.NoError(t, g.Pass([]int{0, 1}))
	require.NoError(t, g.Pass([]int{0}))
	require.Equal(t, domain.AnacondaPhaseSet, g.GetPhase())

	require.NoError(t, g.Keep([]int{0, 1, 2, 3, 4}))
	assert.Equal(t, domain.AnacondaKeepSize, g.GetPlayer(0).GetCardsSize())
	// CPU も 5 枚に絞られている。
	for i := 1; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, domain.AnacondaKeepSize, g.GetPlayer(i).GetCardsSize())
	}
	// ロールフェーズへ (または 1 人残りで即 Result)。
	assert.Contains(t, []domain.AnacondaPhase{domain.AnacondaPhaseRoll, domain.AnacondaPhaseResult}, g.GetPhase())
}

func TestAnaconda_Keep_Errors(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	// wrong phase (still in pass)
	assert.True(t, errors.Is(g.Keep([]int{0, 1, 2, 3, 4}), domain.ErrWrongPhase))

	g.SetPhase(domain.AnacondaPhaseSet)
	assert.True(t, errors.Is(g.Keep([]int{0, 1, 2, 3}), domain.ErrInvalidPlay))     // too few
	assert.True(t, errors.Is(g.Keep([]int{0, 1, 2, 3, 99}), domain.ErrInvalidPlay)) // range
	assert.True(t, errors.Is(g.Keep([]int{0, 0, 1, 2, 3}), domain.ErrInvalidPlay))  // duplicate
}

// setupControlledRoll は 4 人・弱いハイカード手・ロール最終ストリートの決定的状態を作る。
func setupControlledRoll(t *testing.T) *domain.Anaconda {
	t.Helper()
	g := domain.NewDefaultAnaconda()
	if g.GetGameEndFlag() {
		t.Skip("unexpected game end")
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		anacondaSetKept(g.GetPlayer(i), anacondaWeakHand(8+i)...)
	}
	g.SetPhase(domain.AnacondaPhaseRoll)
	g.SetRollIndex(domain.AnacondaKeepSize)
	g.SetPot(40)
	g.SetCurrentBet(0)
	g.SetDealerIdx(g.GetPlayerCnt() - 1)
	g.SetCurrentPlayerIdx(0)
	return g
}

func TestAnaconda_Roll_CallShowdown(t *testing.T) {
	g := setupControlledRoll(t)
	require.True(t, g.IsHumanTurn())
	require.NoError(t, g.PlayerCall())
	// 全員チェック → ショーダウン。
	assert.Equal(t, domain.AnacondaPhaseResult, g.GetPhase())
	assert.GreaterOrEqual(t, g.GetWinnerIdx(), 0)
}

func TestAnaconda_Roll_RaiseFoldsOthers(t *testing.T) {
	g := setupControlledRoll(t)
	require.True(t, g.CanRaise())
	require.NoError(t, g.PlayerRaise())
	// 弱い CPU は降り、人間が総取り。
	assert.Equal(t, domain.AnacondaPhaseResult, g.GetPhase())
	assert.Equal(t, 0, g.GetWinnerIdx())
	assert.Equal(t, domain.AnacondaResultWin, g.GetResult())
}

func TestAnaconda_Roll_HumanFold(t *testing.T) {
	g := setupControlledRoll(t)
	require.NoError(t, g.PlayerFold())
	assert.True(t, g.GetPlayer(0).GetFolded())
	assert.Equal(t, domain.AnacondaResultNone, g.GetResult())
	assert.Equal(t, domain.AnacondaPhaseResult, g.GetPhase())
}

func TestAnaconda_Roll_ActionErrors(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	// wrong phase (still pass)
	assert.True(t, errors.Is(g.PlayerCall(), domain.ErrWrongPhase))

	g2 := setupControlledRoll(t)
	g2.SetCurrentPlayerIdx(1) // CPU's turn
	assert.True(t, errors.Is(g2.PlayerCall(), domain.ErrNotHumanTurn))
}

func TestAnaconda_Showdown_BestHandWins(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	sp, he, cl, di := domain.CardDesignSpade, domain.CardDesignHeart, domain.CardDesignClover, domain.CardDesignDiamond
	// seat0: four of a kind (winner), others high card.
	anacondaSetKept(g.GetPlayer(0), anacondaCard(sp, 8), anacondaCard(he, 8), anacondaCard(cl, 8), anacondaCard(di, 8), anacondaCard(sp, 2))
	anacondaSetKept(g.GetPlayer(1), anacondaWeakHand(9)...)
	anacondaSetKept(g.GetPlayer(2), anacondaWeakHand(8)...)
	anacondaSetKept(g.GetPlayer(3), anacondaWeakHand(7)...)
	g.SetPhase(domain.AnacondaPhaseRoll)
	g.SetPot(100)
	before := g.GetPlayer(0).GetChips()

	g.ResolveShowdownForTest()
	assert.Equal(t, 0, g.GetWinnerIdx())
	assert.Equal(t, domain.AnacondaResultWin, g.GetResult())
	assert.Equal(t, before+100, g.GetPlayer(0).GetChips())
	assert.Equal(t, domain.AnacondaPhaseResult, g.GetPhase())
}

func TestAnaconda_Showdown_FoldToWin(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		anacondaSetKept(g.GetPlayer(i), anacondaWeakHand(9)...)
	}
	// Only seat 2 remains active.
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetFolded(i != 2)
	}
	g.SetPhase(domain.AnacondaPhaseRoll)
	g.SetPot(60)
	before := g.GetPlayer(2).GetChips()

	g.ResolveShowdownForTest()
	assert.Equal(t, 2, g.GetWinnerIdx())
	assert.Equal(t, before+60, g.GetPlayer(2).GetChips())
}

func TestAnaconda_NextRound(t *testing.T) {
	g := setupControlledRoll(t)
	require.NoError(t, g.PlayerCall())
	require.Equal(t, domain.AnacondaPhaseResult, g.GetPhase())
	if g.GetGameEndFlag() {
		t.Skip("game ended")
	}
	g.NextRound()
	assert.Equal(t, domain.AnacondaPhasePass, g.GetPhase())
	assert.Equal(t, 2, g.GetRoundNumber())
	// NextRound while in pass phase is a no-op.
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
}

func TestAnaconda_GameEndByRounds(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	// Drive one full round via controlled roll resolution.
	for i := 0; i < g.GetPlayerCnt(); i++ {
		anacondaSetKept(g.GetPlayer(i), anacondaWeakHand(9)...)
	}
	g.SetPhase(domain.AnacondaPhaseRoll)
	g.SetPot(40)
	g.ResolveShowdownForTest()
	assert.True(t, g.GetGameEndFlag())
	assert.GreaterOrEqual(t, g.GetMatchWinnerIdx(), 0)
	// NextRound after end is a no-op.
	g.NextRound()
	assert.True(t, g.GetGameEndFlag())
}

func TestAnaconda_GameEnd_ActionsBlocked(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		anacondaSetKept(g.GetPlayer(i), anacondaWeakHand(9)...)
	}
	g.SetPhase(domain.AnacondaPhaseRoll)
	g.ResolveShowdownForTest()
	require.True(t, g.GetGameEndFlag())
	assert.True(t, errors.Is(g.Pass([]int{0, 1, 2}), domain.ErrGameEnded))
	assert.True(t, errors.Is(g.Keep([]int{0, 1, 2, 3, 4}), domain.ErrGameEnded))
	assert.True(t, errors.Is(g.PlayerCall(), domain.ErrGameEnded))
}

func TestAnaconda_Hint(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	// Pass phase hint.
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "pass", h.Action)
	assert.Len(t, h.CardIndices, domain.AnacondaPassStart)

	// Set phase hint.
	g.SetPhase(domain.AnacondaPhaseSet)
	h = g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "keep", h.Action)
	assert.Len(t, h.CardIndices, domain.AnacondaKeepSize)

	// Roll phase hint (strong hand → raise).
	sp, he, cl, di := domain.CardDesignSpade, domain.CardDesignHeart, domain.CardDesignClover, domain.CardDesignDiamond
	anacondaSetKept(g.GetPlayer(0), anacondaCard(sp, 8), anacondaCard(he, 8), anacondaCard(cl, 8), anacondaCard(di, 13), anacondaCard(sp, 2))
	g.SetPhase(domain.AnacondaPhaseRoll)
	g.SetCurrentPlayerIdx(0)
	h = g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "raise", h.Action)

	// Result phase → no hint.
	g.SetPhase(domain.AnacondaPhaseResult)
	assert.Nil(t, g.GetHint())
}

func TestAnaconda_RevealedCards(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		anacondaSetKept(g.GetPlayer(i), anacondaWeakHand(9)...)
	}
	// Human always sees own full hand.
	g.SetPhase(domain.AnacondaPhaseRoll)
	g.SetRollIndex(2)
	assert.Len(t, g.GetRevealedCards(0), domain.AnacondaKeepSize)
	// CPU shows only revealed prefix during roll.
	assert.Len(t, g.GetRevealedCards(1), 2)
	// At result an active CPU is fully revealed.
	g.SetPhase(domain.AnacondaPhaseResult)
	assert.True(t, g.IsHandFullyRevealed(1))
	// A folded CPU hides at result.
	g.GetPlayer(2).SetFolded(true)
	assert.Nil(t, g.GetRevealedCards(2))
	// Out of range.
	assert.Nil(t, g.GetRevealedCards(99))
}

func TestAnaconda_ActionLog(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	require.NoError(t, g.Pass([]int{0, 1, 2}))
	assert.NotEmpty(t, g.GetActionLog())
}

// --- JSON ---

func TestAnacondaPlayer_JSON(t *testing.T) {
	p := domain.NewAnacondaPlayer(true, 300)
	p.SetFolded(true)
	p.AddRoundBet(20)
	p.AddStreetBet(5)
	data, err := json.Marshal(p)
	require.NoError(t, err)
	var got domain.AnacondaPlayer
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, 300, got.GetChips())
	assert.True(t, got.GetFolded())
	assert.Equal(t, 20, got.GetRoundBet())
	assert.Equal(t, 5, got.GetStreetBet())
	assert.True(t, got.GetIsHuman())
}

func TestAnacondaPlayer_UnmarshalErrors(t *testing.T) {
	var p domain.AnacondaPlayer
	assert.Error(t, json.Unmarshal([]byte(`{"ch":-5}`), &p))
	assert.Error(t, json.Unmarshal([]byte(`{"ch":10,"rb":-1}`), &p))
	assert.Error(t, json.Unmarshal([]byte(`{"ch":10,"sb":-1}`), &p))
	assert.Error(t, json.Unmarshal([]byte(`not json`), &p))
}

func TestAnaconda_JSONRoundTrip(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	require.NoError(t, g.Pass([]int{0, 1, 2}))

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var got domain.Anaconda
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, g.GetPhase(), got.GetPhase())
	assert.Equal(t, g.GetChips(), got.GetChips())
	assert.Equal(t, g.GetRoundNumber(), got.GetRoundNumber())
	assert.Equal(t, g.GetPassCount(), got.GetPassCount())
	assert.Equal(t, g.GetPlayerCnt(), got.GetPlayerCnt())
}

func TestAnaconda_UnmarshalValidation(t *testing.T) {
	base := `{"cf":{"pc":3,"an":10,"sc":200,"tr":10},"ps":[{"ch":200},{"ch":200},{"ch":200}],"ph":0,"rn":1`
	cases := map[string]string{
		"not json":        `not json`,
		"invalid config":  `{"cf":{"pc":9,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1},{"ch":1}]}`,
		"player mismatch": `{"cf":{"pc":4,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1},{"ch":1}],"ph":0,"rn":1}`,
		"too few players": `{"cf":{"pc":3,"an":10,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1}],"ph":0,"rn":1}`,
		"invalid phase":   base + `,"ph":9}`,
		"round zero":      base + `,"rn":0}`,
		"negative pot":    base + `,"pt":-1}`,
		"pass range":      base + `,"pn":9}`,
		"roll range":      base + `,"ri":9}`,
		"dealer range":    base + `,"di":9}`,
		"current range":   base + `,"ci":9}`,
		"winner range":    base + `,"wi":9}`,
		"result range":    base + `,"re":9}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			var g domain.Anaconda
			assert.Error(t, json.Unmarshal([]byte(body), &g))
		})
	}
}

func TestAnaconda_UnmarshalDefaults(t *testing.T) {
	var g domain.Anaconda
	require.NoError(t, json.Unmarshal([]byte(`{"cf":{"pc":3,"an":10,"sc":200,"tr":10},"ps":[{"ch":200},{"ch":200},{"ch":200}],"ph":0,"rn":1}`), &g))
	assert.Equal(t, 3, g.GetPlayerCnt())
	assert.NotNil(t, g.GetActionLog())
}
