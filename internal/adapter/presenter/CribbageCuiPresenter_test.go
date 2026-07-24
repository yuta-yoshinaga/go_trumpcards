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

func setupCribbageCuiMock() *interfaces.MockCribbageGame {
	m := new(interfaces.MockCribbageGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDealerIdx").Return(0)
	m.On("GetStarter").Return((*domain.Card)(nil))
	m.On("GetPegCount").Return(0)
	m.On("GetPegPlayedCards").Return(([]*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.CribbagePhaseDiscard)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetHandScoreDetails").Return([3]*domain.CribbageScoreDetail{})
	return m
}

func makeCribbagePlayers() []*domain.CribbagePlayer {
	return []*domain.CribbagePlayer{
		domain.NewCribbagePlayer(true),
		domain.NewCribbagePlayer(false),
	}
}

func setupCribbageCuiMockWithPlayers() (*interfaces.MockCribbageGame, []*domain.CribbagePlayer) {
	m := setupCribbageCuiMock()
	players := makeCribbagePlayers()
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestCribbageCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.CribbageCuiPresenter)

	t.Run("initial state with header and player info", func(t *testing.T) {
		m, players := setupCribbageCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Cribbage (クリベッジ)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "[0]SPADE 1")
		assert.Contains(t, result, "[1]HEART 5")
		assert.Contains(t, result, "CPU 1")
		assert.Contains(t, result, "ディスカードフェーズ")
		assert.Contains(t, result, "d <idx,idx>")
	})

	t.Run("starter card shown", func(t *testing.T) {
		m, _ := setupCribbageCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetStarter")
		starter := domain.NewCard(domain.CardDesignHeart, 7, false)
		m.On("GetStarter").Return(starter)

		result := p.Output(m, nil)
		assert.Contains(t, result, "スターター: HEART 7")
	})

	t.Run("starter nil hides section", func(t *testing.T) {
		m, _ := setupCribbageCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.NotContains(t, result, "スターター:")
	})

	t.Run("human with no cards does not print cards line", func(t *testing.T) {
		m, _ := setupCribbageCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "0枚")
		assert.NotContains(t, result, "[0]")
	})

	t.Run("player with scores", func(t *testing.T) {
		m, players := setupCribbageCuiMockWithPlayers()
		players[1].SetCumulativeScore(80)
		players[1].SetRoundScore(15)

		result := p.Output(m, nil)
		assert.Contains(t, result, "累積80点")
		assert.Contains(t, result, "ラウンド15点")
	})

	t.Run("dealer marker shown", func(t *testing.T) {
		m, _ := setupCribbageCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.Contains(t, result, "[D]")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupCribbageCuiMockWithPlayers()
		testErr := errors.New("invalid card index")

		result := p.Output(m, testErr)
		assert.Contains(t, result, "invalid card index")
	})

	t.Run("pegging legend lists indices playable within 31 (boundary)", func(t *testing.T) {
		m, players := setupCribbageCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPegCount")
		m.On("GetPhase").Return(domain.CribbagePhasePegging)
		m.On("GetPegCount").Return(25)
		// [0]=6 → 25+6=31 (exactly the limit, legal); [1]=7 → 32 (illegal).
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "出せる手札: [0]")
		assert.NotContains(t, result, "出せる手札: [0] [1]")
	})

	t.Run("pegging warns to declare go when nothing fits", func(t *testing.T) {
		m, players := setupCribbageCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPegCount")
		m.On("GetPhase").Return(domain.CribbagePhasePegging)
		m.On("GetPegCount").Return(30)
		// Only a 2 in hand → 30+2=32 > 31, so nothing is playable.
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "go を宣言してください")
		assert.NotContains(t, result, "出せる手札:")
	})

	t.Run("game ended shows winner human", func(t *testing.T) {
		m, _ := setupCribbageCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "あなたの勝利です！")
	})

	t.Run("game ended shows winner CPU", func(t *testing.T) {
		m, _ := setupCribbageCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "CPU 1の勝利です！")
	})

	t.Run("pegging phase shows commands", func(t *testing.T) {
		m, _ := setupCribbageCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CribbagePhasePegging)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ペギングフェーズ")
		assert.Contains(t, result, "p <idx>")
		assert.Contains(t, result, "go")
	})

	t.Run("pegging phase shows peg count and played cards", func(t *testing.T) {
		m, _ := setupCribbageCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPegCount")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPegPlayedCards")
		m.On("GetPhase").Return(domain.CribbagePhasePegging)
		m.On("GetPegCount").Return(15)
		m.On("GetPegPlayedCards").Return([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 10, false),
		})

		result := p.Output(m, nil)
		assert.Contains(t, result, "ペギング合計: 15/31")
		assert.Contains(t, result, "出されたカード:")
		assert.Contains(t, result, "SPADE 5")
		assert.Contains(t, result, "HEART 10")
	})

	t.Run("show phase shows commands and score details", func(t *testing.T) {
		m, _ := setupCribbageCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHandScoreDetails")
		m.On("GetPhase").Return(domain.CribbagePhaseShow)
		m.On("GetHandScoreDetails").Return([3]*domain.CribbageScoreDetail{
			{Total: 8, Fifteens: 4, Pairs: 2, Runs: 0, Flush: 0, Nobs: 2},
			nil,
			nil,
		})

		result := p.Output(m, nil)
		assert.Contains(t, result, "ショーフェーズ")
		assert.Contains(t, result, "sn / shownext")
		assert.Contains(t, result, "非ディーラー手札")
		assert.Contains(t, result, "8点")
	})

	t.Run("round end phase shows next command", func(t *testing.T) {
		m, _ := setupCribbageCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHandScoreDetails")
		m.On("GetPhase").Return(domain.CribbagePhaseRoundEnd)
		m.On("GetHandScoreDetails").Return([3]*domain.CribbageScoreDetail{})

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "nr / nextround・・・次のラウンドへ")
	})

	t.Run("discard phase shows current player CPU", func(t *testing.T) {
		m, _ := setupCribbageCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "手番: CPU 1")
	})
}

func TestCribbageCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.CribbageCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockCribbageGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "discard", Detail: "Player discards 2 cards"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "discard")
		assert.Contains(t, result, "Player discards 2 cards")
		m.AssertExpectations(t)
	})

	t.Run("nil entries", func(t *testing.T) {
		m := new(interfaces.MockCribbageGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockCribbageGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})
}

func TestCribbageCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.CribbageCuiPresenter)

	t.Run("discard hint names both crib indices", func(t *testing.T) {
		m := new(interfaces.MockCribbageGame)
		m.On("GetHint").Return(&domain.CribbageHint{Type: "discard", Indices: []int{0, 1}})
		out := p.HintOutput(m)
		assert.Contains(t, out, "0")
		assert.Contains(t, out, "1")
		assert.Contains(t, out, "クリブ")
	})

	t.Run("play hint names the card index", func(t *testing.T) {
		m := new(interfaces.MockCribbageGame)
		m.On("GetHint").Return(&domain.CribbageHint{Type: "play", Indices: []int{2}})
		out := p.HintOutput(m)
		assert.Contains(t, out, "2")
		assert.Contains(t, out, "出す")
	})

	t.Run("nil hint falls back to no-hint message", func(t *testing.T) {
		m := new(interfaces.MockCribbageGame)
		m.On("GetHint").Return((*domain.CribbageHint)(nil))
		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})

	t.Run("unknown hint type falls back to no-hint message", func(t *testing.T) {
		m := new(interfaces.MockCribbageGame)
		m.On("GetHint").Return(&domain.CribbageHint{Type: "??"})
		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})

	t.Run("discard hint with short indices falls back to no-hint message", func(t *testing.T) {
		m := new(interfaces.MockCribbageGame)
		m.On("GetHint").Return(&domain.CribbageHint{Type: "discard", Indices: []int{0}})
		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})

	t.Run("play hint with no indices falls back to no-hint message", func(t *testing.T) {
		m := new(interfaces.MockCribbageGame)
		m.On("GetHint").Return(&domain.CribbageHint{Type: "play"})
		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})
}
