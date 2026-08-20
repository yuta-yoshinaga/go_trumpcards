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
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func tcbMakePlayers() []*domain.ThreeCardBragPlayer {
	return []*domain.ThreeCardBragPlayer{
		domain.NewThreeCardBragPlayer(true, 30),
		domain.NewThreeCardBragPlayer(false, 30),
		domain.NewThreeCardBragPlayer(false, 30),
		domain.NewThreeCardBragPlayer(false, 30),
	}
}

func tcbSetupBaseMock() *interfaces.MockThreeCardBragGame {
	m := new(interfaces.MockThreeCardBragGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetPot").Return(4)
	m.On("GetStake").Return(1)
	m.On("GetPhase").Return(domain.ThreeCardBragPhaseBetting)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetRoundWinnerIdx").Return(-1)
	m.On("GetMatchWinnerIdx").Return(-1)
	m.On("IsShowdown").Return(false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func tcbSetupMockWithPlayers() (*interfaces.MockThreeCardBragGame, []*domain.ThreeCardBragPlayer) {
	m := tcbSetupBaseMock()
	players := tcbMakePlayers()
	m.On("GetPlayerCnt").Return(domain.ThreeCardBragPlayerCnt)
	for i, p := range players {
		m.On("GetPlayer", i).Return(p)
	}
	return m, players
}

func TestThreeCardBragCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ThreeCardBragCuiPresenter)

	t.Run("betting phase shows prompt and human cards", func(t *testing.T) {
		m, players := tcbSetupMockWithPlayers()
		players[0].SetSeen(true)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "SPADE")
		// Seen player with 30 chips at stake 1: raise range max = 30/2 = 15.
		assert.Contains(t, result, "15")
	})

	t.Run("betting phase shows raise unavailable when chips too low", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		m := tcbSetupBaseMock()
		players := []*domain.ThreeCardBragPlayer{
			domain.NewThreeCardBragPlayer(true, 1), // blind, 1 chip, stake 1 -> min 2 > max 1
			domain.NewThreeCardBragPlayer(false, 30),
			domain.NewThreeCardBragPlayer(false, 30),
			domain.NewThreeCardBragPlayer(false, 30),
		}
		m.On("GetPlayerCnt").Return(domain.ThreeCardBragPlayerCnt)
		for i, pl := range players {
			m.On("GetPlayer", i).Return(pl)
		}
		result := p.Output(m, nil)
		assert.Contains(t, result, "Not enough chips to raise.")
	})

	t.Run("betting phase omits raise range on a CPU turn", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		m, _ := tcbSetupMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(1) // CPU at turn
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		assert.NotContains(t, result, "Raise range")
	})

	t.Run("showdown reveals all non-folded hands", func(t *testing.T) {
		m, players := tcbSetupMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ThreeCardBragPhaseShowdown)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsShowdown")
		m.On("IsShowdown").Return(true)
		for i := range players {
			players[i].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
			players[i].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
		}
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end shows winner prompt", func(t *testing.T) {
		m, _ := tcbSetupMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ThreeCardBragPhaseRoundEnd)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundWinnerIdx")
		m.On("GetRoundWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := tcbSetupMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMatchWinnerIdx")
		m.On("GetMatchWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("folded and out statuses render", func(t *testing.T) {
		m, players := tcbSetupMockWithPlayers()
		players[1].SetFolded(true)
		players[2].SetOut(true)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := tcbSetupMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestThreeCardBragCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ThreeCardBragCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := tcbSetupMockWithPlayers()
		m.On("GetHint").Return((*domain.ThreeCardBragHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("see hint", func(t *testing.T) {
		m, _ := tcbSetupMockWithPlayers()
		m.On("GetHint").Return(&domain.ThreeCardBragHint{Action: "see", Reason: "see_first"})
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "see")
	})

	t.Run("raise hint", func(t *testing.T) {
		m, _ := tcbSetupMockWithPlayers()
		m.On("GetHint").Return(&domain.ThreeCardBragHint{Action: "raise", Reason: "strong_hand"})
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "raise")
	})
}

func TestThreeCardBragCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ThreeCardBragCuiPresenter)
	m := new(interfaces.MockThreeCardBragGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bet", Detail: "You bets 1"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "bet")
}

// #5659: 通常のプレイヤー表示は cuiPlayerName で「あなた」「CPU 1」と出している
// のに、**勝者を出す2箇所だけ生の席番号**を埋めていた。他のゲームは全部名前を
// 出しているので、ここだけ「Player 0 の勝ち」に見える。
func TestThreeCardBragCuiPresenter_NamesTheWinner(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ThreeCardBragCuiPresenter)

	t.Run("round end names the winner", func(t *testing.T) {
		m, _ := tcbSetupMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundWinnerIdx")
		m.On("GetPhase").Return(domain.ThreeCardBragPhaseRoundEnd)
		m.On("GetRoundWinnerIdx").Return(1)

		out := p.Output(m, nil)

		assert.Contains(t, out, i18n.Tf("threecardbrag.promptRoundEnd",
			"player", i18n.Tf("cuiPlayerCpu", "idx", "1")))
		assert.NotContains(t, out, i18n.Tf("threecardbrag.promptRoundEnd", "player", "1"))
	})

	// **人間が勝ったら「あなた」。**席番号のままだと自分が勝ったのかも読み取れない。
	t.Run("the human wins as あなた, not as seat 0", func(t *testing.T) {
		m, _ := tcbSetupMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMatchWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetMatchWinnerIdx").Return(0)

		out := p.Output(m, nil)

		assert.Contains(t, out, i18n.Tf("threecardbrag.gameEnd", "player", i18n.T("cuiPlayerYou")))
		assert.NotContains(t, out, i18n.Tf("threecardbrag.gameEnd", "player", "0"))
	})
}
