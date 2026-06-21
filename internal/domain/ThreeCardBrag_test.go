//go:build test

package domain_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestBrag(humanSeat0 bool) *domain.ThreeCardBrag {
	cfg := domain.DefaultThreeCardBragConfig()
	players := make([]*domain.ThreeCardBragPlayer, domain.ThreeCardBragPlayerCnt)
	for i := range players {
		players[i] = domain.NewThreeCardBragPlayer(i == 0 && humanSeat0, cfg.StartingChips)
	}
	return domain.NewThreeCardBrag(domain.NewTrumpCards(0), players, cfg)
}

func tcbCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

func tcbSetHand(p *domain.ThreeCardBragPlayer, cards ...*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestThreeCardBragConfig_Validate(t *testing.T) {
	cfg := domain.DefaultThreeCardBragConfig()
	assert.NoError(t, cfg.Validate())
	assert.Error(t, domain.ThreeCardBragConfig{CpuDifficulty: 99, Ante: 1, StartingChips: 30}.Validate())
	assert.Error(t, domain.ThreeCardBragConfig{CpuDifficulty: 0, Ante: 0, StartingChips: 30}.Validate())
	assert.Error(t, domain.ThreeCardBragConfig{CpuDifficulty: 0, Ante: 1, StartingChips: 1}.Validate())
}

func TestThreeCardBragEval_Categories(t *testing.T) {
	S, C, H, D := domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond
	cat := func(cards ...*domain.Card) int { c, _ := domain.ThreeCardBragEval(cards); return c }

	assert.Equal(t, domain.ThreeCardBragPrial, cat(tcbCard(S, 13), tcbCard(C, 13), tcbCard(H, 13)))     // KKK
	assert.Equal(t, domain.ThreeCardBragRunningFlush, cat(tcbCard(S, 4), tcbCard(S, 5), tcbCard(S, 6))) // 4-5-6♠
	assert.Equal(t, domain.ThreeCardBragRun, cat(tcbCard(S, 4), tcbCard(C, 5), tcbCard(H, 6)))          // 4-5-6 mixed
	assert.Equal(t, domain.ThreeCardBragRun, cat(tcbCard(S, 1), tcbCard(C, 2), tcbCard(H, 3)))          // A-2-3 run
	assert.Equal(t, domain.ThreeCardBragFlush, cat(tcbCard(D, 2), tcbCard(D, 7), tcbCard(D, 13)))       // flush, not run
	assert.Equal(t, domain.ThreeCardBragPair, cat(tcbCard(S, 13), tcbCard(C, 13), tcbCard(H, 5)))       // KK5
	assert.Equal(t, domain.ThreeCardBragHighCard, cat(tcbCard(S, 1), tcbCard(C, 13), tcbCard(H, 9)))    // A-K-9

	_, tb := domain.ThreeCardBragEval([]*domain.Card{tcbCard(S, 1)}) // wrong length
	assert.Nil(t, tb)
}

func TestThreeCardBragEval_Specials(t *testing.T) {
	S, C, H, D := domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond
	// 3-3-3 is the best prial (beats A-A-A).
	c333, tb333 := domain.ThreeCardBragEval([]*domain.Card{tcbCard(S, 3), tcbCard(C, 3), tcbCard(H, 3)})
	cAAA, tbAAA := domain.ThreeCardBragEval([]*domain.Card{tcbCard(S, 1), tcbCard(C, 1), tcbCard(H, 1)})
	assert.Equal(t, 1, domain.ThreeCardBragCompare(c333, tb333, cAAA, tbAAA))
	// A-2-3 is the best run (beats A-K-Q run).
	cA23, tbA23 := domain.ThreeCardBragEval([]*domain.Card{tcbCard(S, 1), tcbCard(C, 2), tcbCard(H, 3)})
	cAKQ, tbAKQ := domain.ThreeCardBragEval([]*domain.Card{tcbCard(S, 1), tcbCard(C, 13), tcbCard(D, 12)})
	assert.Equal(t, 1, domain.ThreeCardBragCompare(cA23, tbA23, cAKQ, tbAKQ))
	// Run beats flush.
	cRun, tbRun := domain.ThreeCardBragEval([]*domain.Card{tcbCard(S, 4), tcbCard(C, 5), tcbCard(H, 6)})
	cFl, tbFl := domain.ThreeCardBragEval([]*domain.Card{tcbCard(D, 2), tcbCard(D, 7), tcbCard(D, 13)})
	assert.Equal(t, 1, domain.ThreeCardBragCompare(cRun, tbRun, cFl, tbFl))
}

func TestNewDefaultThreeCardBrag(t *testing.T) {
	g := domain.NewDefaultThreeCardBrag()
	require.NotNil(t, g)
	assert.Equal(t, domain.ThreeCardBragPlayerCnt, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	assert.Equal(t, -1, g.GetMatchWinnerIdx())
	assert.Nil(t, g.GetPlayer(9))
}

func TestThreeCardBrag_ResetDealsAntesAndCards(t *testing.T) {
	g := newTestBrag(true)
	g.Reset()
	assert.Equal(t, domain.ThreeCardBragPhaseBetting, g.GetPhase())
	for i := 0; i < domain.ThreeCardBragPlayerCnt; i++ {
		assert.Equal(t, domain.ThreeCardBragHandSize, g.GetPlayer(i).GetCardsSize(), "player %d cards", i)
		assert.False(t, g.GetPlayer(i).GetSeen())
	}
	// pot = 4 antes.
	assert.Equal(t, domain.ThreeCardBragPlayerCnt*g.GetConfig().Ante, g.GetPot())
	assert.Equal(t, g.GetConfig().StartingChips-g.GetConfig().Ante, g.GetPlayer(0).GetChips())
}

func TestThreeCardBrag_SeeAndBet(t *testing.T) {
	g := newTestBrag(true)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	require.False(t, g.GetPlayer(0).GetSeen())
	require.NoError(t, g.PlayerSee())
	assert.True(t, g.GetPlayer(0).GetSeen())
	assert.Error(t, g.PlayerSee()) // already seen
	potBefore := g.GetPot()
	require.NoError(t, g.PlayerBet()) // seen -> pays 2*stake
	assert.Equal(t, potBefore+g.GetStake()*2, g.GetPot())
}

func TestThreeCardBrag_RaiseAndWrongTurn(t *testing.T) {
	g := newTestBrag(true)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	assert.Error(t, g.PlayerRaise(g.GetStake())) // not greater
	require.NoError(t, g.PlayerRaise(g.GetStake()+2))
	// Wrong phase guard.
	g.SetPhase(domain.ThreeCardBragPhaseRoundEnd)
	assert.ErrorIs(t, g.PlayerBet(), domain.ErrWrongPhase)
}

func TestThreeCardBrag_FoldToOneWinsPot(t *testing.T) {
	g := newTestBrag(true)
	g.Reset()
	// Fold seats 2 and 3 directly; active = {0,1}. Human seat 0 folds -> seat 1 wins.
	g.GetPlayer(2).SetFolded(true)
	g.GetPlayer(3).SetFolded(true)
	g.SetCurrentPlayerIdx(0)
	pot := g.GetPot()
	chips1 := g.GetPlayer(1).GetChips()
	require.NoError(t, g.PlayerFold())
	assert.Equal(t, 1, g.GetRoundWinnerIdx())
	assert.Equal(t, chips1+pot, g.GetPlayer(1).GetChips())
	assert.Equal(t, domain.ThreeCardBragPhaseRoundEnd, g.GetPhase())
}

func TestThreeCardBrag_ShowResolvesShowdown(t *testing.T) {
	g := newTestBrag(true)
	g.Reset()
	g.GetPlayer(2).SetFolded(true)
	g.GetPlayer(3).SetFolded(true)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(0).SetSeen(true)
	g.SetStake(1)
	// seat 0 has a Prial (KKK), seat 1 a high card -> seat 0 wins the show.
	tcbSetHand(g.GetPlayer(0), tcbCard(domain.CardDesignSpade, 13), tcbCard(domain.CardDesignClover, 13), tcbCard(domain.CardDesignHeart, 13))
	tcbSetHand(g.GetPlayer(1), tcbCard(domain.CardDesignSpade, 2), tcbCard(domain.CardDesignClover, 7), tcbCard(domain.CardDesignHeart, 9))
	require.True(t, g.CanShow())
	require.NoError(t, g.PlayerShow())
	assert.True(t, g.IsShowdown())
	assert.Equal(t, 0, g.GetRoundWinnerIdx())
}

func TestThreeCardBrag_Hint(t *testing.T) {
	g := newTestBrag(true)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "see", h.Action) // blind -> see first
}

func TestThreeCardBrag_FullCpuGame(t *testing.T) {
	g := newTestBrag(false) // all CPU
	g.Reset()
	guard := 0
	for !g.GetGameEndFlag() && guard < 500000 {
		guard++
		switch g.GetPhase() {
		case domain.ThreeCardBragPhaseBetting:
			g.CpuAct()
		case domain.ThreeCardBragPhaseRoundEnd:
			g.NextRound()
		}
	}
	assert.True(t, g.GetGameEndFlag(), "match must terminate")
	assert.GreaterOrEqual(t, g.GetMatchWinnerIdx(), 0)
	// The match winner holds all the chips.
	total := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		total += g.GetPlayer(i).GetChips()
	}
	assert.Equal(t, g.GetConfig().StartingChips*domain.ThreeCardBragPlayerCnt, total)
}

func TestThreeCardBrag_StartDealEndsWhenTooFewCanAnte(t *testing.T) {
	g := newTestBrag(false)
	g.Reset()
	g.SetPhase(domain.ThreeCardBragPhaseRoundEnd)
	g.GetPlayer(1).SetChips(0)
	g.GetPlayer(2).SetChips(0)
	g.GetPlayer(3).SetChips(0)
	g.NextRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.ThreeCardBragPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetMatchWinnerIdx())
}

func TestThreeCardBrag_JSONRoundTrip(t *testing.T) {
	g := newTestBrag(true)
	g.Reset()
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var g2 domain.ThreeCardBrag
	require.NoError(t, json.Unmarshal(data, &g2))
	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetPot(), g2.GetPot())
	assert.Equal(t, g.GetPlayerCnt(), g2.GetPlayerCnt())
}

func TestThreeCardBrag_UnmarshalRejectsInvalid(t *testing.T) {
	g := newTestBrag(true)
	g.Reset()
	data, err := json.Marshal(g)
	require.NoError(t, err)
	tampered := strings.Replace(string(data), `"di":0`, `"di":9`, 1)
	require.NotEqual(t, string(data), tampered)
	var bad domain.ThreeCardBrag
	assert.Error(t, bad.UnmarshalJSON([]byte(tampered)))
	var bad2 domain.ThreeCardBrag
	assert.Error(t, bad2.UnmarshalJSON([]byte(`{"ps":[null]}`)))
	var bad3 domain.ThreeCardBrag
	assert.Error(t, bad3.UnmarshalJSON([]byte(`not json`)))
}
