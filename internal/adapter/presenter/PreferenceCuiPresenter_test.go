//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func makePreferencePlayers() []*domain.PreferencePlayer {
	return []*domain.PreferencePlayer{
		domain.NewPreferencePlayer(true),
		domain.NewPreferencePlayer(false),
		domain.NewPreferencePlayer(false),
	}
}

func setupPreferenceCuiMock() *interfaces.MockPreferenceGame {
	m := new(interfaces.MockPreferenceGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.PreferencePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetContract").Return(domain.PreferenceBidSix)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayerScores").Return([domain.PreferencePlayerCnt]int{0, 0, 0})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupPreferenceCuiMockWithPlayers() (*interfaces.MockPreferenceGame, []*domain.PreferencePlayer) {
	m := setupPreferenceCuiMock()
	players := makePreferencePlayers()
	m.On("GetPlayerCnt").Return(3)
	for i := 0; i < 3; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestPreferenceCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.PreferenceCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupPreferenceCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Préférence")
		assert.NotEmpty(t, result)
	})

	t.Run("bid phase prompt", func(t *testing.T) {
		m, _ := setupPreferenceCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDeclarerIdx")
		m.On("GetPhase").Return(domain.PreferencePhaseBid)
		m.On("GetDeclarerIdx").Return(-1)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("misere shows no trump", func(t *testing.T) {
		m, _ := setupPreferenceCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupPreferenceCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PreferencePhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupPreferenceCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PreferencePhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end shows the declarer's contract as achieved", func(t *testing.T) {
		m, players := setupPreferenceCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PreferencePhaseRoundEnd)
		// Declarer (seat 0) took 6 tricks on a Six contract → achieved.
		for range 6 {
			players[0].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
		}
		players[1].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, false)})
		players[2].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 7, false)})
		result := p.Output(m, nil)
		assert.Contains(t, result, "達成")
		assert.NotContains(t, result, "失敗")
		// Per-player trick tally is present.
		assert.Contains(t, result, "各プレイヤーのトリック数")
	})

	t.Run("round end shows the declarer's contract as failed", func(t *testing.T) {
		m, players := setupPreferenceCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PreferencePhaseRoundEnd)
		// Declarer took only 2 tricks on a Six contract → failed.
		for range 2 {
			players[0].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
		}
		result := p.Output(m, nil)
		assert.Contains(t, result, "失敗")
	})

	t.Run("round end with no declarer skips the result line", func(t *testing.T) {
		m, _ := setupPreferenceCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDeclarerIdx")
		m.On("GetPhase").Return(domain.PreferencePhaseRoundEnd)
		m.On("GetDeclarerIdx").Return(-1)
		result := p.Output(m, nil)
		assert.NotContains(t, result, "各プレイヤーのトリック数")
	})

	t.Run("round end with a nil declarer player skips the result", func(t *testing.T) {
		m, players := setupPreferenceCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayer")
		m.On("GetPhase").Return(domain.PreferencePhaseRoundEnd)
		m.On("GetPlayer", 0).Return((*domain.PreferencePlayer)(nil)) // declarer missing
		m.On("GetPlayer", 1).Return(players[1])
		m.On("GetPlayer", 2).Return(players[2])
		result := p.Output(m, nil)
		assert.NotContains(t, result, "各プレイヤーのトリック数")
	})

	t.Run("round end skips a nil player in the trick tally", func(t *testing.T) {
		m, players := setupPreferenceCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayer")
		m.On("GetPhase").Return(domain.PreferencePhaseRoundEnd)
		m.On("GetPlayer", 0).Return(players[0]) // declarer present
		m.On("GetPlayer", 1).Return((*domain.PreferencePlayer)(nil))
		m.On("GetPlayer", 2).Return(players[2])
		result := p.Output(m, nil)
		assert.Contains(t, result, "各プレイヤーのトリック数")
	})

	// Each contract has a distinct trick target; a Misère is achieved by taking zero.
	contractCases := []struct {
		name     string
		contract domain.PreferenceBid
		tricks   int
		achieved bool
	}{
		{"misere achieved with zero tricks", domain.PreferenceBidMisere, 0, true},
		{"misere failed with a trick", domain.PreferenceBidMisere, 1, false},
		{"seven achieved", domain.PreferenceBidSeven, 7, true},
		{"eight failed", domain.PreferenceBidEight, 7, false},
	}
	for _, tc := range contractCases {
		t.Run("round end "+tc.name, func(t *testing.T) {
			m, players := setupPreferenceCuiMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetContract")
			m.On("GetPhase").Return(domain.PreferencePhaseRoundEnd)
			m.On("GetContract").Return(tc.contract)
			for range tc.tricks {
				players[0].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
			}
			result := p.Output(m, nil)
			if tc.achieved {
				assert.Contains(t, result, "達成")
				assert.NotContains(t, result, "失敗")
			} else {
				assert.Contains(t, result, "失敗")
			}
		})
	}

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupPreferenceCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupPreferenceCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestPreferenceCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.PreferenceCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupPreferenceCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.PreferenceHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupPreferenceCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.PreferenceHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupPreferenceCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.PreferenceHint{CardIndices: nil, Reason: "follow_win"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestPreferenceCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PreferenceCuiPresenter)
	m := new(interfaces.MockPreferenceGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewPreferencePlayer(true)).Maybe()
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}
