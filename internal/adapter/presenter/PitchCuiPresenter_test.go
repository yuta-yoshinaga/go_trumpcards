package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
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
