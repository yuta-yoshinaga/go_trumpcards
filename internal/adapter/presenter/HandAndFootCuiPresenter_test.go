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

func makeHandAndFootPlayers() []*domain.HandAndFootPlayer {
	return []*domain.HandAndFootPlayer{
		domain.NewHandAndFootPlayer(true),
		domain.NewHandAndFootPlayer(false),
		domain.NewHandAndFootPlayer(false),
		domain.NewHandAndFootPlayer(false),
	}
}

func setupHandAndFootCuiMock() *interfaces.MockHandAndFootGame {
	m := new(interfaces.MockHandAndFootGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(75)
	m.On("GetDiscardPileCount").Return(0)
	m.On("GetIsFrozen").Return(false)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.HandAndFootPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetTeamMelds", 0).Return(([]*domain.CanastaMeld)(nil))
	m.On("GetTeamMelds", 1).Return(([]*domain.CanastaMeld)(nil))
	m.On("GetTeamRed3s", 0).Return(([]*domain.Card)(nil))
	m.On("GetTeamRed3s", 1).Return(([]*domain.Card)(nil))
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupHandAndFootCuiMockWithPlayers() (*interfaces.MockHandAndFootGame, []*domain.HandAndFootPlayer) {
	m := setupHandAndFootCuiMock()
	players := makeHandAndFootPlayers()
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestHandAndFootCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.HandAndFootCuiPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupHandAndFootCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Hand and Foot")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "山札: 75枚")
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "[0]SPADE 5")
		assert.Contains(t, result, "ds")
		assert.Contains(t, result, "dd")
	})

	t.Run("frozen pile shown", func(t *testing.T) {
		m, _ := setupHandAndFootCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetIsFrozen")
		m.On("GetIsFrozen").Return(true)
		assert.Contains(t, p.Output(m, nil), "[フリーズ]")
	})

	t.Run("discard top shown", func(t *testing.T) {
		m, _ := setupHandAndFootCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
		assert.Contains(t, p.Output(m, nil), "捨て札: HEART 7")
	})

	t.Run("team meld shown", func(t *testing.T) {
		m, _ := setupHandAndFootCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamMelds")
		meld := &domain.CanastaMeld{
			Cards: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignClover, 7, false),
			},
			IsNatural: true,
		}
		m.On("GetTeamMelds", 0).Return([]*domain.CanastaMeld{meld})
		m.On("GetTeamMelds", 1).Return(([]*domain.CanastaMeld)(nil))
		result := p.Output(m, nil)
		assert.Contains(t, result, "チーム0")
		assert.Contains(t, result, "クリーン")
		assert.Contains(t, result, "SPADE 7")
	})

	t.Run("in foot tag shown", func(t *testing.T) {
		m, players := setupHandAndFootCuiMockWithPlayers()
		players[0].SetInFoot(true)
		assert.Contains(t, p.Output(m, nil), "★フット中")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupHandAndFootCuiMockWithPlayers()
		assert.Contains(t, p.Output(m, errors.New("invalid card index")), "invalid card index")
	})

	t.Run("game ended shows winning team", func(t *testing.T) {
		m, _ := setupHandAndFootCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(1)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "チーム1の勝利です！")
	})

	t.Run("meld phase commands", func(t *testing.T) {
		m, _ := setupHandAndFootCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.HandAndFootPhaseMeld)
		result := p.Output(m, nil)
		assert.Contains(t, result, "メルドフェーズ")
		assert.Contains(t, result, "sm")
	})

	t.Run("discard phase commands", func(t *testing.T) {
		m, _ := setupHandAndFootCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.HandAndFootPhaseDiscard)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ディスカードフェーズ")
		assert.Contains(t, result, "go")
	})

	t.Run("round end shows next command", func(t *testing.T) {
		m, _ := setupHandAndFootCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.HandAndFootPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "nr / nextround")
	})
}

func TestHandAndFootCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.HandAndFootCuiPresenter)

	t.Run("lists meld candidates", func(t *testing.T) {
		m := new(interfaces.MockHandAndFootGame)
		m.On("IsHumanTurn").Return(true)
		m.On("GetCurrentPlayerIdx").Return(0)
		m.On("SuggestMelds", 0).Return([][]*domain.Card{
			{
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignClover, 7, false),
			},
		})
		result := p.HintOutput(m)
		assert.Contains(t, result, "メルド候補")
		assert.Contains(t, result, "7")
	})

	t.Run("no meld available", func(t *testing.T) {
		m := new(interfaces.MockHandAndFootGame)
		m.On("IsHumanTurn").Return(true)
		m.On("GetCurrentPlayerIdx").Return(0)
		m.On("SuggestMelds", 0).Return(([][]*domain.Card)(nil))
		assert.Contains(t, p.HintOutput(m), "出せるメルドはありません")
	})

	t.Run("not the human's turn", func(t *testing.T) {
		m := new(interfaces.MockHandAndFootGame)
		m.On("IsHumanTurn").Return(false)
		assert.Contains(t, p.HintOutput(m), "あなたの番ではありません")
	})
}

func TestHandAndFootCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.HandAndFootCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockHandAndFootGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "draw_stock", Detail: "drew from stock"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "draw_stock")
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockHandAndFootGame)
		m.On("GetGameEndFlag").Return(false)
		result := p.ActionLogOutput(m)
		assert.NotEmpty(t, result)
		m.AssertExpectations(t)
	})
}
