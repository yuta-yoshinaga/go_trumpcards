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

func makeTTJPlayers() []*domain.TwoTenJackPlayer {
	return []*domain.TwoTenJackPlayer{
		domain.NewTwoTenJackPlayer(true),
		domain.NewTwoTenJackPlayer(false),
		domain.NewTwoTenJackPlayer(false),
		domain.NewTwoTenJackPlayer(false),
	}
}

func setupTTJCuiMock() (*interfaces.MockTwoTenJackGame, []*domain.TwoTenJackPlayer) {
	m := new(interfaces.MockTwoTenJackGame)
	players := makeTTJPlayers()
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TwoTenJackPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultTwoTenJackConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestTwoTenJackCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.TwoTenJackCuiPresenter)

	t.Run("basic render", func(t *testing.T) {
		m, players := setupTTJCuiMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Two Ten Jack")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "SPADE")
		assert.Contains(t, result, "手番:")
	})

	t.Run("declare phase", func(t *testing.T) {
		m, _ := setupTTJCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TwoTenJackPhaseDeclare)
		result := p.Output(m, nil)
		assert.Contains(t, result, "宣言フェーズ")
	})

	t.Run("trick end", func(t *testing.T) {
		m, _ := setupTTJCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TwoTenJackPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
	})

	t.Run("round end", func(t *testing.T) {
		m, _ := setupTTJCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TwoTenJackPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
	})

	t.Run("game end shows winner team", func(t *testing.T) {
		m, _ := setupTTJCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
		assert.Contains(t, result, "チーム0")
	})

	t.Run("error shown", func(t *testing.T) {
		m, _ := setupTTJCuiMock()
		result := p.Output(m, errors.New("bad"))
		assert.Contains(t, result, "bad")
	})

	t.Run("trump unknown", func(t *testing.T) {
		m, _ := setupTTJCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(-1)
		result := p.Output(m, nil)
		assert.Contains(t, result, "未宣言")
	})

	t.Run("trick shown", func(t *testing.T) {
		m, _ := setupTTJCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
		}
		m.On("GetCurrentTrick").Return(trick)
		result := p.Output(m, nil)
		assert.Contains(t, result, "CLOVER 3")
	})
}

func TestTwoTenJackCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.TwoTenJackCuiPresenter)

	t.Run("nil hint", func(t *testing.T) {
		m := new(interfaces.MockTwoTenJackGame)
		m.On("GetHint").Return((*domain.TwoTenJackHint)(nil))
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("trump hint", func(t *testing.T) {
		m := new(interfaces.MockTwoTenJackGame)
		suit := domain.CardDesignSpade
		m.On("GetHint").Return(&domain.TwoTenJackHint{TrumpSuit: &suit, Reason: "strategic_trump"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
		assert.Contains(t, result, "SPADE")
	})

	t.Run("play hint", func(t *testing.T) {
		m := new(interfaces.MockTwoTenJackGame)
		idx := 0
		m.On("GetHint").Return(&domain.TwoTenJackHint{CardIndex: &idx, Reason: "follow_suit"})
		player := domain.NewTwoTenJackPlayer(true)
		player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		m.On("GetPlayer", 0).Return(player)
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
		assert.Contains(t, result, "リードスートに追随")
	})

	t.Run("empty hint", func(t *testing.T) {
		m := new(interfaces.MockTwoTenJackGame)
		m.On("GetHint").Return(&domain.TwoTenJackHint{Reason: "none"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestTwoTenJackCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TwoTenJackCuiPresenter)
	m := new(interfaces.MockTwoTenJackGame)
	m.On("GetGameEndFlag").Return(false)
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "棋譜はありません")
}
