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

func newTestTeenPatti(humanSeat0 bool) *domain.TeenPatti {
	cfg := domain.DefaultTeenPattiConfig()
	players := make([]*domain.TeenPattiPlayer, domain.TeenPattiPlayerCnt)
	for i := range players {
		players[i] = domain.NewTeenPattiPlayer(i == 0 && humanSeat0, cfg.StartingChips)
	}
	return domain.NewTeenPatti(domain.NewTrumpCards(0), players, cfg)
}

func newAllHumanTeenPatti() *domain.TeenPatti {
	cfg := domain.DefaultTeenPattiConfig()
	players := make([]*domain.TeenPattiPlayer, domain.TeenPattiPlayerCnt)
	for i := range players {
		players[i] = domain.NewTeenPattiPlayer(true, cfg.StartingChips)
	}
	return domain.NewTeenPatti(domain.NewTrumpCards(0), players, cfg)
}

func tpCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

func tpSetHand(p *domain.TeenPattiPlayer, cards ...*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestTeenPattiConfig_Validate(t *testing.T) {
	cfg := domain.DefaultTeenPattiConfig()
	assert.NoError(t, cfg.Validate())
	assert.Error(t, domain.TeenPattiConfig{CpuDifficulty: 99, Ante: 1, StartingChips: 30}.Validate())
	assert.Error(t, domain.TeenPattiConfig{CpuDifficulty: 0, Ante: 0, StartingChips: 30}.Validate())
}

func TestNewDefaultTeenPatti(t *testing.T) {
	g := domain.NewDefaultTeenPatti()
	require.NotNil(t, g)
	assert.Equal(t, domain.TeenPattiPlayerCnt, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	assert.Equal(t, -1, g.GetMatchWinnerIdx())
	assert.Nil(t, g.GetPlayer(9))
}

func TestTeenPatti_ResetDealsBootAndCards(t *testing.T) {
	g := newTestTeenPatti(true)
	g.Reset()
	assert.Equal(t, domain.TeenPattiPhaseBetting, g.GetPhase())
	for i := 0; i < domain.TeenPattiPlayerCnt; i++ {
		assert.Equal(t, domain.TeenPattiHandSize, g.GetPlayer(i).GetCardsSize())
		assert.False(t, g.GetPlayer(i).GetSeen())
	}
	assert.Equal(t, domain.TeenPattiPlayerCnt*g.GetConfig().Ante, g.GetPot())
}

func TestTeenPatti_SeeBetRaiseFold(t *testing.T) {
	g := newTestTeenPatti(true)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	require.NoError(t, g.PlayerSee())
	assert.True(t, g.GetPlayer(0).GetSeen())
	potBefore := g.GetPot()
	require.NoError(t, g.PlayerBet()) // seen pays 2*stake
	assert.Equal(t, potBefore+g.GetStake()*2, g.GetPot())

	g.SetCurrentPlayerIdx(0)
	assert.Error(t, g.PlayerRaise(g.GetStake())) // not greater
	require.NoError(t, g.PlayerRaise(g.GetStake()+2))
}

func TestTeenPatti_FoldToOneWinsPot(t *testing.T) {
	g := newTestTeenPatti(true)
	g.Reset()
	g.GetPlayer(2).SetFolded(true)
	g.GetPlayer(3).SetFolded(true)
	g.SetCurrentPlayerIdx(0)
	pot := g.GetPot()
	chips1 := g.GetPlayer(1).GetChips()
	require.NoError(t, g.PlayerFold())
	assert.Equal(t, 1, g.GetRoundWinnerIdx())
	assert.Equal(t, chips1+pot, g.GetPlayer(1).GetChips())
}

func TestTeenPatti_ShowResolvesShowdown(t *testing.T) {
	g := newTestTeenPatti(true)
	g.Reset()
	g.GetPlayer(2).SetFolded(true)
	g.GetPlayer(3).SetFolded(true)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(0).SetSeen(true)
	g.SetStake(1)
	tpSetHand(g.GetPlayer(0), tpCard(domain.CardDesignSpade, 13), tpCard(domain.CardDesignClover, 13), tpCard(domain.CardDesignHeart, 13))
	tpSetHand(g.GetPlayer(1), tpCard(domain.CardDesignSpade, 2), tpCard(domain.CardDesignClover, 7), tpCard(domain.CardDesignHeart, 9))
	require.True(t, g.CanShow())
	require.NoError(t, g.PlayerShow())
	assert.Equal(t, 0, g.GetRoundWinnerIdx())
}

func TestTeenPatti_SideShowAcceptLoserFolds(t *testing.T) {
	g := newAllHumanTeenPatti()
	g.Reset()
	// All four seen & active; seat 0 requests a side show vs the previous seen (seat 3).
	for i := 0; i < domain.TeenPattiPlayerCnt; i++ {
		g.GetPlayer(i).SetSeen(true)
	}
	g.SetStake(1)
	g.SetCurrentPlayerIdx(0)
	// seat 0 strong (KKK), seat 3 weak -> seat 3 should lose and fold.
	tpSetHand(g.GetPlayer(0), tpCard(domain.CardDesignSpade, 13), tpCard(domain.CardDesignClover, 13), tpCard(domain.CardDesignHeart, 13))
	tpSetHand(g.GetPlayer(3), tpCard(domain.CardDesignSpade, 2), tpCard(domain.CardDesignClover, 7), tpCard(domain.CardDesignHeart, 9))
	require.True(t, g.CanRequestSideShow())
	require.NoError(t, g.PlayerRequestSideShow())
	assert.Equal(t, domain.TeenPattiPhaseSideShow, g.GetPhase())
	assert.Equal(t, 3, g.GetSideShowTarget())
	assert.Equal(t, 3, g.GetCurrentPlayerIdx())
	require.NoError(t, g.PlayerRespondSideShow(true))
	assert.True(t, g.GetPlayer(3).GetFolded()) // seat 3 lost and folded
	assert.Equal(t, domain.TeenPattiPhaseBetting, g.GetPhase())
	// The accepted side show result is retained for display.
	req, tgt, loser, ok := g.GetLastSideShow()
	require.True(t, ok)
	assert.Equal(t, 0, req)
	assert.Equal(t, 3, tgt)
	assert.Equal(t, 3, loser)
}

func TestTeenPatti_LastSideShowLifecycle(t *testing.T) {
	g := newAllHumanTeenPatti()
	g.Reset()
	// No side show has happened yet.
	_, _, _, ok := g.GetLastSideShow()
	assert.False(t, ok)

	for i := 0; i < domain.TeenPattiPlayerCnt; i++ {
		g.GetPlayer(i).SetSeen(true)
	}
	g.SetStake(1)
	g.SetCurrentPlayerIdx(0)

	// Declining does not record a comparison result.
	require.NoError(t, g.PlayerRequestSideShow())
	require.NoError(t, g.PlayerRespondSideShow(false))
	_, _, _, ok = g.GetLastSideShow()
	assert.False(t, ok)

	// Accepting records a result; a new deal clears it.
	g.SetCurrentPlayerIdx(0)
	require.NoError(t, g.PlayerRequestSideShow())
	require.NoError(t, g.PlayerRespondSideShow(true))
	_, _, _, ok = g.GetLastSideShow()
	require.True(t, ok)
	g.SetPhase(domain.TeenPattiPhaseRoundEnd)
	g.NextRound()
	_, _, _, ok = g.GetLastSideShow()
	assert.False(t, ok)
}

func TestTeenPatti_SideShowDeclineContinues(t *testing.T) {
	g := newAllHumanTeenPatti()
	g.Reset()
	for i := 0; i < domain.TeenPattiPlayerCnt; i++ {
		g.GetPlayer(i).SetSeen(true)
	}
	g.SetStake(1)
	g.SetCurrentPlayerIdx(0)
	require.NoError(t, g.PlayerRequestSideShow())
	require.NoError(t, g.PlayerRespondSideShow(false))
	assert.Equal(t, domain.TeenPattiPhaseBetting, g.GetPhase())
	// Nobody folded from a declined side show.
	folded := 0
	for i := 0; i < domain.TeenPattiPlayerCnt; i++ {
		if g.GetPlayer(i).GetFolded() {
			folded++
		}
	}
	assert.Equal(t, 0, folded)
}

func TestTeenPatti_CannotSideShowWhenBlindOrTwoLeft(t *testing.T) {
	g := newTestTeenPatti(true)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	// Blind -> cannot side show.
	assert.False(t, g.CanRequestSideShow())
	assert.Error(t, g.PlayerRequestSideShow())
	// Two players left -> cannot side show.
	g.GetPlayer(0).SetSeen(true)
	g.GetPlayer(1).SetSeen(true)
	g.GetPlayer(2).SetFolded(true)
	g.GetPlayer(3).SetFolded(true)
	assert.False(t, g.CanRequestSideShow())
}

func TestTeenPatti_CpuRespondsToSideShow(t *testing.T) {
	g := newTestTeenPatti(true) // seat 0 human, seats 1-3 CPU
	g.Reset()
	for i := 0; i < domain.TeenPattiPlayerCnt; i++ {
		g.GetPlayer(i).SetSeen(true)
	}
	g.SetStake(1)
	g.SetCurrentPlayerIdx(0)
	require.NoError(t, g.PlayerRequestSideShow())
	require.Equal(t, domain.TeenPattiPhaseSideShow, g.GetPhase())
	require.False(t, g.GetPlayer(g.GetCurrentPlayerIdx()).GetIsHuman()) // CPU target
	g.CpuAct()                                                          // CPU responds
	assert.Equal(t, domain.TeenPattiPhaseBetting, g.GetPhase())
}

func TestTeenPatti_Hint(t *testing.T) {
	g := newTestTeenPatti(true)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "see", h.Action)
}

func TestTeenPatti_FullCpuGame(t *testing.T) {
	g := newTestTeenPatti(false)
	g.Reset()
	guard := 0
	for !g.GetGameEndFlag() && guard < 500000 {
		guard++
		switch g.GetPhase() {
		case domain.TeenPattiPhaseBetting, domain.TeenPattiPhaseSideShow:
			g.CpuAct()
		case domain.TeenPattiPhaseRoundEnd:
			g.NextRound()
		}
	}
	assert.True(t, g.GetGameEndFlag(), "match must terminate")
	assert.GreaterOrEqual(t, g.GetMatchWinnerIdx(), 0)
	total := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		total += g.GetPlayer(i).GetChips()
	}
	assert.Equal(t, g.GetConfig().StartingChips*domain.TeenPattiPlayerCnt, total)
}

func TestTeenPatti_StartDealEndsWhenTooFewCanAnte(t *testing.T) {
	g := newTestTeenPatti(false)
	g.Reset()
	// Only seat 0 can afford the next ante -> the new deal must end the match,
	// not start a one-player betting loop.
	g.SetPhase(domain.TeenPattiPhaseRoundEnd)
	g.GetPlayer(1).SetChips(0)
	g.GetPlayer(2).SetChips(0)
	g.GetPlayer(3).SetChips(0)
	g.NextRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.TeenPattiPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetMatchWinnerIdx())
}

func TestTeenPatti_JSONRoundTrip(t *testing.T) {
	g := newTestTeenPatti(true)
	g.Reset()
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var g2 domain.TeenPatti
	require.NoError(t, json.Unmarshal(data, &g2))
	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetPot(), g2.GetPot())
}

func TestTeenPatti_UnmarshalRejectsInvalid(t *testing.T) {
	g := newTestTeenPatti(true)
	g.Reset()
	data, err := json.Marshal(g)
	require.NoError(t, err)
	tampered := strings.Replace(string(data), `"di":0`, `"di":9`, 1)
	require.NotEqual(t, string(data), tampered)
	var bad domain.TeenPatti
	assert.Error(t, bad.UnmarshalJSON([]byte(tampered)))
	var bad2 domain.TeenPatti
	assert.Error(t, bad2.UnmarshalJSON([]byte(`{"ps":[null]}`)))
	var bad3 domain.TeenPatti
	assert.Error(t, bad3.UnmarshalJSON([]byte(`not json`)))
}

func TestTeenPatti_JSONRoundTripSideShow(t *testing.T) {
	g := newAllHumanTeenPatti()
	g.Reset()
	for i := 0; i < domain.TeenPattiPlayerCnt; i++ {
		g.GetPlayer(i).SetSeen(true)
	}
	g.SetStake(1)
	g.SetCurrentPlayerIdx(0)
	tpSetHand(g.GetPlayer(0), tpCard(domain.CardDesignSpade, 13), tpCard(domain.CardDesignClover, 13), tpCard(domain.CardDesignHeart, 13))
	tpSetHand(g.GetPlayer(3), tpCard(domain.CardDesignSpade, 2), tpCard(domain.CardDesignClover, 7), tpCard(domain.CardDesignHeart, 9))
	require.NoError(t, g.PlayerRequestSideShow())
	require.NoError(t, g.PlayerRespondSideShow(true))

	data, err := json.Marshal(g)
	require.NoError(t, err)
	var g2 domain.TeenPatti
	require.NoError(t, json.Unmarshal(data, &g2))
	req, tgt, loser, ok := g2.GetLastSideShow()
	require.True(t, ok)
	assert.Equal(t, 0, req)
	assert.Equal(t, 3, tgt)
	assert.Equal(t, 3, loser)

	// A side show whose loser is neither participant is rejected.
	tampered := strings.Replace(string(data), `"ls":{"rq":0,"tg":3,"ls":3}`, `"ls":{"rq":0,"tg":3,"ls":1}`, 1)
	require.NotEqual(t, string(data), tampered)
	var bad domain.TeenPatti
	assert.Error(t, bad.UnmarshalJSON([]byte(tampered)))
}

// **拒否されたレイズが賭け単位を壊していた (#4729 の調査中に発見)。**
// applyRaise は g.stake を更新してからチップ不足を判定しており、弾いた後も
// 上がったままだった。以降の全員がその額でコールを迫られる。
func TestTeenPatti_RejectedRaiseLeavesTheStakeUntouched(t *testing.T) {
	g := domain.NewDefaultTeenPatti()
	g.Reset()
	g.SetCurrentPlayerIdx(0)

	before := g.GetStake()
	chips := g.GetPlayer(0).GetChips()

	err := g.PlayerRaise(chips * 10) // 到底払えない額
	if err == nil {
		t.Fatal("チップ不足のレイズは弾かれるはず")
	}
	if got := g.GetStake(); got != before {
		t.Errorf("弾かれたのに stake が %d -> %d に変わった", before, got)
	}
	// チップも減っていないこと。
	if got := g.GetPlayer(0).GetChips(); got != chips {
		t.Errorf("弾かれたのにチップが %d -> %d に減った", chips, got)
	}
}

// **レイズ可能域は1箇所から出す (#4729)。**CUI の表示と Web の入力上限が
// 別式だと「入力できたのに弾かれる」ずれになる。
func TestTeenPatti_GetRaiseRange(t *testing.T) {
	newGame := func(t *testing.T, chips int, seen bool) *domain.TeenPatti {
		t.Helper()
		g := domain.NewDefaultTeenPatti()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetStake(10)
		p := g.GetPlayer(0)
		p.SetChips(chips)
		p.SetSeen(seen)
		return g
	}

	t.Run("blind can raise up to its chips", func(t *testing.T) {
		g := newGame(t, 100, false)
		lo, hi, ok := g.GetRaiseRange(0)
		if !ok || lo != 11 || hi != 100 {
			t.Errorf("GetRaiseRange = (%d, %d, %v), want (11, 100, true)", lo, hi, ok)
		}
	})

	// Seen は倍払うので上限が半分になる。ここが CUI と食い違いやすい。
	t.Run("seen can only raise up to half its chips", func(t *testing.T) {
		g := newGame(t, 100, true)
		_, hi, ok := g.GetRaiseRange(0)
		if !ok || hi != 50 {
			t.Errorf("seen max = %d (ok=%v), want 50", hi, ok)
		}
	})

	t.Run("not enough chips reports ok=false", func(t *testing.T) {
		g := newGame(t, 5, false) // stake 10 なので最低 11 が必要
		if _, _, ok := g.GetRaiseRange(0); ok {
			t.Error("チップ不足なのに ok=true")
		}
	})

	// **上限ちょうどは実際に通ること。**式が1つずれていると、表示上は出せる額が
	// サーバーに弾かれる。範囲の正しさを実際の PlayerRaise で裏取りする。
	t.Run("the reported maximum is actually accepted", func(t *testing.T) {
		for _, seen := range []bool{false, true} {
			g := newGame(t, 100, seen)
			_, hi, ok := g.GetRaiseRange(0)
			if !ok {
				t.Fatalf("seen=%v: レイズ可能なはず", seen)
			}
			if err := g.PlayerRaise(hi); err != nil {
				t.Errorf("seen=%v: 上限 %d が弾かれた: %v", seen, hi, err)
			}
		}
	})

	t.Run("one over the reported maximum is rejected", func(t *testing.T) {
		g := newGame(t, 100, true)
		_, hi, _ := g.GetRaiseRange(0)
		if err := g.PlayerRaise(hi + 1); err == nil {
			t.Errorf("上限+1 (%d) は弾かれるはず", hi+1)
		}
	})
}
