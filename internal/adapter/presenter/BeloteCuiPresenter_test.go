package presenter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupBeloteCuiMock() *interfaces.MockBeloteGame {
	m := new(interfaces.MockBeloteGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.BelotePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetFaceUpCard").Return((*domain.Card)(nil))
	m.On("GetMakerTeam").Return(0)
	m.On("GetMakerPlayerIdx").Return(0)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetRoundPoints", 0).Return(0)
	m.On("GetRoundPoints", 1).Return(0)
	m.On("GetRoundBeloteBonus", 0).Return(0)
	m.On("GetRoundBeloteBonus", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultBeloteConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeBelotePlayers() []*domain.BelotePlayer {
	return []*domain.BelotePlayer{
		domain.NewBelotePlayer(true, 0),
		domain.NewBelotePlayer(false, 1),
		domain.NewBelotePlayer(false, 0),
		domain.NewBelotePlayer(false, 1),
	}
}

func setupBeloteCuiMockWithPlayers() (*interfaces.MockBeloteGame, []*domain.BelotePlayer) {
	m := setupBeloteCuiMock()
	players := makeBelotePlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestBeloteCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BeloteCuiPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupBeloteCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Belote (ベロート)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "トリック: 1")
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "切り札: SPADE (メイカー: チーム0)")
		assert.Contains(t, result, "[0]SPADE 1")
	})

	t.Run("trump undecided", func(t *testing.T) {
		m, _ := setupBeloteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: 未決定")
	})

	t.Run("face-up card", func(t *testing.T) {
		m, _ := setupBeloteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetFaceUpCard")
		m.On("GetFaceUpCard").Return(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "表向きカード: HEART 11")
	})

	t.Run("phase: bid pickup", func(t *testing.T) {
		m, _ := setupBeloteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BelotePhaseBidPickUp)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ピックアップフェーズ")
	})

	t.Run("phase: call trump", func(t *testing.T) {
		m, _ := setupBeloteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BelotePhaseBidCallTrump)

		result := p.Output(m, nil)
		assert.Contains(t, result, "コールトランプフェーズ")
	})

	t.Run("phase: trick end", func(t *testing.T) {
		m, _ := setupBeloteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BelotePhaseTrickEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
	})

	t.Run("phase: round end", func(t *testing.T) {
		m, _ := setupBeloteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BelotePhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupBeloteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
	})
}

func TestBeloteCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BeloteCuiPresenter)

	t.Run("nil hint", func(t *testing.T) {
		m := setupBeloteCuiMock()
		m.On("GetHint").Return((*domain.BeloteHint)(nil))

		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("order up hint", func(t *testing.T) {
		m := setupBeloteCuiMock()
		ok := true
		m.On("GetHint").Return(&domain.BeloteHint{OrderUp: &ok, Reason: "strategic_pickup"})

		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("pass hint", func(t *testing.T) {
		m := setupBeloteCuiMock()
		ok := false
		m.On("GetHint").Return(&domain.BeloteHint{OrderUp: &ok, Reason: "pass_recommended"})

		result := p.HintOutput(m)
		assert.Contains(t, result, "パス")
	})

	t.Run("call suit hint", func(t *testing.T) {
		m := setupBeloteCuiMock()
		suit := 2
		m.On("GetHint").Return(&domain.BeloteHint{Suit: &suit, Reason: "strategic_call"})

		result := p.HintOutput(m)
		assert.Contains(t, result, "コール")
	})

	t.Run("card hint", func(t *testing.T) {
		m, players := setupBeloteCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		idx := 0
		m.On("GetHint").Return(&domain.BeloteHint{CardIndex: &idx, Reason: "trump_cut"})

		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestBeloteCuiPresenter_ActionLogOutput(t *testing.T) {
	m := setupBeloteCuiMock()
	p := new(presenter.BeloteCuiPresenter)
	result := p.ActionLogOutput(m)
	// even empty log should not panic; result should be a string
	assert.NotNil(t, result)
}
