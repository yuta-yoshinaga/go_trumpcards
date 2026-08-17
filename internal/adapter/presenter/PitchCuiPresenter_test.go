package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupPitchCuiMock() *interfaces.MockPitchGame {
	m := new(interfaces.MockPitchGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(0)
	m.On("GetDealerIdx").Return(3)
	m.On("GetCurrentBid").Return(0)
	m.On("GetTrumpSuit").Return(0)
	m.On("GetBidWinnerIdx").Return(-1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.PitchPhaseBid)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(-1)
	m.On("GetRoundBreakdown").Return(domain.PitchRoundBreakdown{
		High: domain.PitchNoScorer, Low: domain.PitchNoScorer,
		Jack: domain.PitchNoScorer, Game: domain.PitchNoScorer,
	}).Maybe()
	m.On("GetConfig").Return(domain.DefaultPitchConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makePitchPlayers() []*domain.PitchPlayer {
	return []*domain.PitchPlayer{
		domain.NewPitchPlayer(true),
		domain.NewPitchPlayer(false),
		domain.NewPitchPlayer(false),
		domain.NewPitchPlayer(false),
	}
}

func setupPitchCuiMockWithPlayers() (*interfaces.MockPitchGame, []*domain.PitchPlayer) {
	m := setupPitchCuiMock()
	players := makePitchPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestPitchCuiPresenter_Output_Bid(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.PitchCuiPresenter)

	m, players := setupPitchCuiMockWithPlayers()
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))

	result := p.Output(m, nil)
	assert.Contains(t, result, "Pitch (ピッチ)")
	assert.Contains(t, result, "ラウンド: 1")
	assert.Contains(t, result, "親: CPU 3")
	assert.Contains(t, result, "ビッド: 0")
	assert.Contains(t, result, "ビッドフェーズ: あなたの番")
	assert.Contains(t, result, "あなた: ビッド=未ビッド")
}

func TestPitchCuiPresenter_Output_PassedBidShown(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.PitchCuiPresenter)

	m, players := setupPitchCuiMockWithPlayers()
	players[0].SetBid(0)
	players[1].SetBid(3)

	result := p.Output(m, nil)
	assert.Contains(t, result, "あなた: ビッド=pass")
	assert.Contains(t, result, "CPU 1: ビッド=3")
}

func TestPitchCuiPresenter_Output_GameEnd(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.PitchCuiPresenter)

	m, _ := setupPitchCuiMockWithPlayers()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerIdx").Return(0)

	result := p.Output(m, nil)
	assert.Contains(t, result, "ゲーム終了")
}

func TestPitchCuiPresenter_Output_Error(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.PitchCuiPresenter)
	m, _ := setupPitchCuiMockWithPlayers()
	result := p.Output(m, errors.New("boom"))
	assert.Contains(t, result, "boom")
}

func TestPitchCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.PitchCuiPresenter)

	t.Run("nil hint returns no hint message", func(t *testing.T) {
		m := new(interfaces.MockPitchGame)
		m.On("GetHint").Return((*domain.PitchHint)(nil))
		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})

	t.Run("bid hint", func(t *testing.T) {
		m := new(interfaces.MockPitchGame)
		bid := 3
		m.On("GetHint").Return(&domain.PitchHint{Bid: &bid, Reason: "bid_strong"})
		assert.Contains(t, p.HintOutput(m), "ビッド 3")
	})

	t.Run("card hint", func(t *testing.T) {
		m := new(interfaces.MockPitchGame)
		idx := 0
		m.On("GetHint").Return(&domain.PitchHint{CardIndex: &idx, Reason: "trump_cut"})
		players := makePitchPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		m.On("GetPlayerCnt").Return(4)
		m.On("GetPlayer", 0).Return(players[0])
		m.On("GetPlayer", 1).Return(players[1])
		m.On("GetPlayer", 2).Return(players[2])
		m.On("GetPlayer", 3).Return(players[3])
		out := p.HintOutput(m)
		assert.Contains(t, out, "[0]")
		assert.Contains(t, out, "トランプでカット")
	})
}

func TestPitchCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PitchCuiPresenter)
	m := new(interfaces.MockPitchGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bid", Detail: "You bid 3"},
	})
	out := p.ActionLogOutput(m)
	assert.Contains(t, out, "You bid 3")
}

// **入札前に手札の得点価値を暗算させていた (#4751)。**Web は入札中にゲーム
// 得点バッジと内訳ポップオーバーを出している。
func TestPitchCuiPresenter_HandPips(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.PitchCuiPresenter)

	withHand := func(values ...int) *interfaces.MockPitchGame {
		m, players := setupPitchCuiMockWithPlayers()
		players[0].Reset()
		for _, v := range values {
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		return m
	}

	t.Run("totals the game pips in the human hand", func(t *testing.T) {
		// ♠10(10) ♠A(4) ♠K(3) ♠7(0) = 17
		out := p.Output(withHand(10, 1, 13, 7), nil)
		assert.Contains(t, out, "手札のゲーム得点: 17")
	})

	// **0点の札も内訳に並べる。**並べないと「見落とし」なのか「0点」なのか
	// 区別が付かない。
	t.Run("lists worthless cards in the breakdown too", func(t *testing.T) {
		out := p.Output(withHand(10, 7), nil)
		assert.Contains(t, out, "=10")
		assert.Contains(t, out, "=0")
	})

	// **人間の手札だけを見る。**手札を持っている最初の席を拾う実装だと、
	// 人間が配り終える前に CPU の手札を人間の得点として出してしまう。
	t.Run("shows nothing when only a CPU holds cards", func(t *testing.T) {
		m, players := setupPitchCuiMockWithPlayers()
		players[0].Reset()
		players[1].Reset()
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		assert.NotContains(t, p.Output(m, nil), "手札のゲーム得点")
	})

	t.Run("shows nothing outside the bid phase", func(t *testing.T) {
		m := withHand(10, 1)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PitchPhasePlay)
		assert.NotContains(t, p.Output(m, nil), "手札のゲーム得点")
	})
}

// #5584: 4 種の得点 (High/Low/Jack/Game) はこのゲームの骨格なのに、CUI も
// 合計しか出していなかった。
func TestPitchCuiPresenter_ShowsTheRoundBreakdown(t *testing.T) {
	i18n.SetLang("ja")
	build := func(bd domain.PitchRoundBreakdown) string {
		m, _ := setupPitchCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundBreakdown")
		m.On("GetPhase").Return(domain.PitchPhaseRoundEnd)
		m.On("GetRoundBreakdown").Return(bd)
		return new(presenter.PitchCuiPresenter).Output(m, nil)
	}

	out := build(domain.PitchRoundBreakdown{High: 0, Low: 1, Jack: 0, Game: domain.PitchNoScorer})
	// 4 カテゴリすべてが並ぶこと。
	for _, label := range []string{"High", "Low", "Jack", "Game"} {
		assert.Contains(t, out, label)
	}
	// **誰も取っていないカテゴリは「なし」。**黙って省くと、争われなかったのか
	// 見落としたのか区別が付かない。
	assert.Contains(t, out, i18n.T("pitch.scoringNobody"))

	// 全部「なし」の局面でも 4 つとも出る。
	none := domain.PitchRoundBreakdown{
		High: domain.PitchNoScorer, Low: domain.PitchNoScorer,
		Jack: domain.PitchNoScorer, Game: domain.PitchNoScorer,
	}
	assert.Equal(t, 4, strings.Count(build(none), i18n.T("pitch.scoringNobody")))
}

// ラウンド途中では出さない。まだ確定していない。
func TestPitchCuiPresenter_HidesTheBreakdownMidRound(t *testing.T) {
	i18n.SetLang("ja")
	m, _ := setupPitchCuiMockWithPlayers()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
	m.On("GetPhase").Return(domain.PitchPhasePlay)

	out := new(presenter.PitchCuiPresenter).Output(m, nil)
	assert.NotContains(t, out, strings.SplitN(i18n.T("pitch.scoringLine"), "{{", 2)[0])
}
