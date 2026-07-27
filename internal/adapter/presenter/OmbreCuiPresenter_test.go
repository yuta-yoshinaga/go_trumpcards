//go:build test

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

func makeOmbrePlayers() []*domain.OmbrePlayer {
	return []*domain.OmbrePlayer{
		domain.NewOmbrePlayer(true),
		domain.NewOmbrePlayer(false),
		domain.NewOmbrePlayer(false),
	}
}

func setupOmbreCuiMock() *interfaces.MockOmbreGame {
	m := new(interfaces.MockOmbreGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetWinningBid").Return(domain.OmbreBidEntrar)
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.OmbrePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetCurrentBidderIdx").Return(1)
	m.On("GetOmbreIdx").Return(0)
	m.On("GetOutcome").Return(domain.OmbreOutcomeSacar)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayerScores").Return([domain.OmbrePlayerCnt]int{0, 0, 0})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupOmbreCuiMockWithPlayers() (*interfaces.MockOmbreGame, []*domain.OmbrePlayer) {
	m := setupOmbreCuiMock()
	players := makeOmbrePlayers()
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestOmbreCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.OmbreCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupOmbreCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "オンブル")    // translated helpTitle / role
		assert.Contains(t, result, "マストフォロー") // play-phase help mentions must-follow
		assert.NotEmpty(t, result)
	})

	t.Run("bid phase prompt", func(t *testing.T) {
		m, _ := setupOmbreCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OmbrePhaseBid)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "ビッド") // translated bid prompt/help
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupOmbreCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OmbrePhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupOmbreCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OmbrePhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "サカール") // outcome label
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupOmbreCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupOmbreCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestOmbreCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.OmbreCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupOmbreCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.OmbreHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupOmbreCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.OmbreHint{CardIndices: []int{0}, Reason: "lead_high"})
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("bid decision hint shows the action, not an empty card list", func(t *testing.T) {
		m, _ := setupOmbreCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.OmbreHint{CardIndices: nil, Reason: "bid_solo"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ソロ")  // recommended action name
		assert.Contains(t, result, "を推奨") // hintDecision format
		assert.NotContains(t, result, "HINT: -")
	})

	t.Run("non-bid empty-card hint falls back to the card line", func(t *testing.T) {
		m, _ := setupOmbreCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.OmbreHint{CardIndices: nil, Reason: "discard_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "-") // hintCard fallback keeps the placeholder
		assert.NotContains(t, result, "を推奨")
	})
}

func TestOmbreCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.OmbreCuiPresenter)
	m := new(interfaces.MockOmbreGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}
