//go:build test

package presenter_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupGanjifaCuiMock() *interfaces.MockGanjifaGame {
	m := new(interfaces.MockGanjifaGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.GanjifaPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayerScores").Return([domain.GanjifaPlayerCnt]int{0, 0, 0})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetHint").Return(nil).Maybe()
	return m
}

func setupGanjifaCuiMockWithPlayers() (*interfaces.MockGanjifaGame, []*domain.GanjifaPlayer) {
	m := setupGanjifaCuiMock()
	players := makeGanjifaPlayers()
	m.On("GetPlayerCnt").Return(3)
	m.On("GetLeadPlayerIdx").Return(-1).Maybe()
	for i := 0; i < 3; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestGanjifaCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.GanjifaCuiPresenter)

	t.Run("play phase shows the hand and the current player", func(t *testing.T) {
		m, players := setupGanjifaCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(1, 12, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "ガンジファ")
		assert.Contains(t, result, "[0]")
	})

	// The whole point of the game is that ranks read in opposite directions, so
	// the frame has to say which direction applies right now.
	t.Run("strong trump states that higher numbers win", func(t *testing.T) {
		m, _ := setupGanjifaCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "強い群")
		assert.NotContains(t, result, "数字が小さいほど強い")
	})

	t.Run("weak trump states that lower numbers win", func(t *testing.T) {
		m, _ := setupGanjifaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(6)
		result := p.Output(m, nil)
		assert.Contains(t, result, "数字が小さいほど強い")
	})

	// cuiCardStr renders designs 5-8 as "UNKNOWN", which would collapse the 48
	// weak-group cards into one indistinguishable string in the hand listing.
	t.Run("weak-group cards are named, not UNKNOWN", func(t *testing.T) {
		m, players := setupGanjifaCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(7, 4, false))
		result := p.Output(m, nil)
		assert.NotContains(t, result, "UNKNOWN")
		assert.Contains(t, result, domain.GanjifaSuitName(7))
		assert.Contains(t, result, domain.GanjifaSuitGlyph(7))
	})

	t.Run("every suit renders with a distinct label", func(t *testing.T) {
		seen := map[string]bool{}
		for suit := 1; suit <= domain.GanjifaSuitCnt; suit++ {
			m, players := setupGanjifaCuiMockWithPlayers()
			players[0].AddCard(domain.NewCard(suit, 5, false))
			result := p.Output(m, nil)
			label := domain.GanjifaSuitGlyph(suit) + " " + domain.GanjifaSuitName(suit) + " 5"
			assert.Contains(t, result, label)
			assert.False(t, seen[label], "label %q is reused", label)
			seen[label] = true
		}
	})

	t.Run("trick cards are rendered", func(t *testing.T) {
		m, _ := setupGanjifaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(3, 9, false)},
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, domain.GanjifaSuitName(3))
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupGanjifaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GanjifaPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
	})

	t.Run("round end lists every player's trick count", func(t *testing.T) {
		m, players := setupGanjifaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GanjifaPhaseRoundEnd)
		players[0].AddTrick([]*domain.Card{domain.NewCard(1, 7, false)})
		result := p.Output(m, nil)
		assert.Contains(t, result, "各プレイヤーのトリック数")
	})

	t.Run("error block is rendered", func(t *testing.T) {
		m, _ := setupGanjifaCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})

	t.Run("game end names the winner", func(t *testing.T) {
		m, _ := setupGanjifaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(1)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	// A tie leaves winnerPlayer at -1; the banner must not index a player with it.
	t.Run("game end with no winner does not crash", func(t *testing.T) {
		m, _ := setupGanjifaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(-1)
		assert.NotPanics(t, func() { p.Output(m, nil) })
	})

	t.Run("nil player is skipped", func(t *testing.T) {
		m := setupGanjifaCuiMock()
		m.On("GetPlayerCnt").Return(1)
		m.On("GetPlayer", 0).Return((*domain.GanjifaPlayer)(nil))
		assert.NotPanics(t, func() { p.Output(m, nil) })
	})
}

func TestGanjifaCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.GanjifaCuiPresenter)

	t.Run("hint names the card and the reason", func(t *testing.T) {
		m, players := setupGanjifaCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(5, 1, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.GanjifaHint{CardIndices: []int{0}, Reason: "lead_high"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "[0]")
		assert.Contains(t, result, domain.GanjifaSuitName(5))
	})

	t.Run("hint with no card indices still renders the reason", func(t *testing.T) {
		m, _ := setupGanjifaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.GanjifaHint{Reason: "follow_duck"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "-")
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupGanjifaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.GanjifaHint)(nil))
		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})
}

func TestGanjifaCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.GanjifaCuiPresenter)
	m := new(interfaces.MockGanjifaGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays Taj 12"},
	})
	assert.Contains(t, p.ActionLogOutput(m), "play")
}

// #5653: Web は TrickDisplay に winnerIdx を渡してトリック勝者にバッジを出して
// いるのに、CUI は「次のトリックへ」としか言わず誰が取ったのか示していなかった。
// 同じトリックテイキングの Sedma は sedma.trickWinner で既に出している。
func TestGanjifaCuiPresenter_TrickEndNamesTheWinner(t *testing.T) {
	p := new(presenter.GanjifaCuiPresenter)

	trickEndMock := func(winner int) *interfaces.MockGanjifaGame {
		m, _ := setupGanjifaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLeadPlayerIdx")
		m.On("GetPhase").Return(domain.GanjifaPhaseTrickEnd)
		m.On("GetLeadPlayerIdx").Return(winner)
		return m
	}

	// **3 人全員分を確かめる。**席0 だけ見ていると、常に自分を勝者と呼ぶ実装が通る。
	names := []string{
		color.Bold(i18n.T("cuiPlayerYou")),
		color.Bold(i18n.Tf("cuiPlayerCpu", "idx", "1")),
		color.Bold(i18n.Tf("cuiPlayerCpu", "idx", "2")),
	}
	for idx, name := range names {
		t.Run("names seat "+strconv.Itoa(idx), func(t *testing.T) {
			out := p.Output(trickEndMock(idx), nil)

			assert.Contains(t, out, i18n.Tf("ganjifa.trickWinner", "name", name))
			// 既存の案内文は残す。
			assert.Contains(t, out, i18n.T("ganjifa.promptTrickEndHelp"))
		})
	}

	t.Run("says nothing before anyone has led", func(t *testing.T) {
		out := p.Output(trickEndMock(-1), nil)

		// **この文言は {{name}} で始まる**ので前置きが空になる。空文字で
		// NotContains を掛けても常に失敗する (= 何も確かめられない)。名前の
		// 後ろ側「がトリックを獲得」で見る。
		_, suffix, ok := strings.Cut(i18n.Tf("ganjifa.trickWinner", "name", "\x00"), "\x00")
		require.True(t, ok)
		require.NotEmpty(t, strings.TrimSpace(suffix))
		assert.NotContains(t, out, suffix)
	})
}
