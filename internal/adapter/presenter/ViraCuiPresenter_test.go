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

func setupViraCuiMock() *interfaces.MockViraGame {
	m := new(interfaces.MockViraGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.ViraPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetContract").Return(domain.ViraBidGask)
	m.On("GetPot").Return(30)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayerScores").Return([domain.ViraPlayerCnt]int{0, 0, 0})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetHint").Return(nil).Maybe()
	return m
}

func setupViraCuiMockWithPlayers() (*interfaces.MockViraGame, []*domain.ViraPlayer) {
	m := setupViraCuiMock()
	players := makeViraPlayers()
	m.On("GetPlayerCnt").Return(3)
	for i := 0; i < 3; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestViraCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ViraCuiPresenter)

	t.Run("play phase shows the hand and the declarer", func(t *testing.T) {
		m, players := setupViraCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "ヴィーラ")
		assert.Contains(t, result, "[0]")
		assert.Contains(t, result, "宣言者")
	})

	// The pot carries forward between rounds and every settlement is paid out of
	// it, so a frame that omits it hides the one number the player is tracking --
	// and the Web view shows it, so dropping it here desyncs the two screens.
	t.Run("the pot is on screen", func(t *testing.T) {
		m, _ := setupViraCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPot")
		m.On("GetPot").Return(180)
		assert.Contains(t, p.Output(m, nil), "180")
	})

	t.Run("bid phase prompt names the Vira ladder, not another game's", func(t *testing.T) {
		m, _ := setupViraCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDeclarerIdx")
		m.On("GetPhase").Return(domain.ViraPhaseBid)
		m.On("GetDeclarerIdx").Return(-1)
		m.On("IsHumanBidTurn").Return(true)
		m.On("GetBids").Return([domain.ViraPlayerCnt]domain.ViraBid{})
		result := p.Output(m, nil)
		for _, want := range []string{"ガスク", "ソロ", "ミゼール", "ヴィーラ"} {
			assert.Contains(t, result, want)
		}
		// Préférence's ladder — the help line was copied from it and said these.
		for _, wrong := range []string{"シックス", "セブン", "エイト"} {
			assert.NotContains(t, result, wrong, "%q belongs to Préférence, not Vira", wrong)
		}
	})

	t.Run("no trump shown for Misere", func(t *testing.T) {
		m, _ := setupViraCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(0)
		assert.Contains(t, p.Output(m, nil), "なし")
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupViraCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ViraPhaseTrickEnd)
		assert.Contains(t, p.Output(m, nil), "トリック終了")
	})

	t.Run("round end shows the contract as achieved", func(t *testing.T) {
		m, players := setupViraCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ViraPhaseRoundEnd)
		// Declarer (seat 0) took 7 tricks on a Gask contract → achieved.
		for range 7 {
			players[0].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
		}
		result := p.Output(m, nil)
		assert.Contains(t, result, "達成")
		assert.NotContains(t, result, "失敗")
		assert.Contains(t, result, "各プレイヤーのトリック数")
	})

	t.Run("round end shows the contract as failed", func(t *testing.T) {
		m, players := setupViraCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ViraPhaseRoundEnd)
		for range 2 {
			players[0].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
		}
		assert.Contains(t, p.Output(m, nil), "失敗")
	})

	// Misère inverts the test: any trick at all fails it.
	t.Run("misere fails the instant the declarer takes a trick", func(t *testing.T) {
		m, players := setupViraCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetContract")
		m.On("GetPhase").Return(domain.ViraPhaseRoundEnd)
		m.On("GetContract").Return(domain.ViraBidMisere)
		players[0].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
		assert.Contains(t, p.Output(m, nil), "失敗")
	})

	t.Run("misere is made on zero tricks", func(t *testing.T) {
		m, _ := setupViraCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetContract")
		m.On("GetPhase").Return(domain.ViraPhaseRoundEnd)
		m.On("GetContract").Return(domain.ViraBidMisere)
		assert.Contains(t, p.Output(m, nil), "達成")
	})

	t.Run("round end with no declarer skips the result line", func(t *testing.T) {
		m, _ := setupViraCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDeclarerIdx")
		m.On("GetPhase").Return(domain.ViraPhaseRoundEnd)
		m.On("GetDeclarerIdx").Return(-1)
		assert.NotContains(t, p.Output(m, nil), "各プレイヤーのトリック数")
	})

	t.Run("trick cards are rendered", func(t *testing.T) {
		m, _ := setupViraCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 9, false)},
		})
		assert.Contains(t, p.Output(m, nil), "HEART")
	})

	t.Run("error block is rendered", func(t *testing.T) {
		m, _ := setupViraCuiMockWithPlayers()
		assert.Contains(t, p.Output(m, errors.New("boom")), "boom")
	})

	t.Run("game end names the winner", func(t *testing.T) {
		m, _ := setupViraCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(1)
		assert.Contains(t, p.Output(m, nil), "ゲーム終了")
	})

	// A tie leaves winnerPlayer at -1; the banner must not index a player with it.
	t.Run("game end with no winner does not crash", func(t *testing.T) {
		m, _ := setupViraCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(-1)
		assert.NotPanics(t, func() { p.Output(m, nil) })
	})

	t.Run("nil player is skipped", func(t *testing.T) {
		m := setupViraCuiMock()
		m.On("GetPlayerCnt").Return(1)
		m.On("GetPlayer", 0).Return((*domain.ViraPlayer)(nil))
		assert.NotPanics(t, func() { p.Output(m, nil) })
	})
}

func TestViraCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ViraCuiPresenter)

	// Every reason the domain can emit must resolve to a translation; three of
	// these had none and would have printed the raw key to the player.
	t.Run("every emitted reason resolves to prose", func(t *testing.T) {
		for _, reason := range []string{
			"lead_high", "lead_low", "follow_win", "follow_block", "misere_duck", "misere_force",
		} {
			m, players := setupViraCuiMockWithPlayers()
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
			m.On("GetHint").Return(&domain.ViraHint{CardIndices: []int{0}, Reason: reason})
			out := p.HintOutput(m)
			assert.NotContains(t, out, reason, "reason %q printed as a raw key", reason)
			assert.Contains(t, out, "[0]")
		}
	})

	t.Run("hint with no card indices still renders the reason", func(t *testing.T) {
		m, _ := setupViraCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.ViraHint{Reason: "lead_low"})
		assert.Contains(t, p.HintOutput(m), "-")
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupViraCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.ViraHint)(nil))
		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})
}

func TestViraCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ViraCuiPresenter)
	m := new(interfaces.MockViraGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewViraPlayer(true)).Maybe()
	assert.Contains(t, p.ActionLogOutput(m), "play")
}
