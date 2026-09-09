//go:build test
// +build test

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

func setupTongitsCuiMock() *interfaces.MockTongitsGame {
	m := new(interfaces.MockTongitsGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(41)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TongitsPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetIsTongits").Return(false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

	return m
}

func makeTongitsPlayers() []*domain.TongitsPlayer {
	return []*domain.TongitsPlayer{
		domain.NewTongitsPlayer(true),
		domain.NewTongitsPlayer(false),
		domain.NewTongitsPlayer(false),
	}
}

func setupTongitsCuiMockWithPlayers() (*interfaces.MockTongitsGame, []*domain.TongitsPlayer) {
	m := setupTongitsCuiMock()
	players := makeTongitsPlayers()
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestTongitsCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.TongitsCuiPresenter)

	t.Run("initial state with header and player info", func(t *testing.T) {
		m, players := setupTongitsCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Tongits (トンギッツ)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "山札: 41枚")
		assert.Contains(t, result, "あなた: 累積0点 ラウンド0点 2枚")
		assert.Contains(t, result, "[0]SPADE 1")
		assert.Contains(t, result, "[1]HEART 5")
		assert.Contains(t, result, "CPU 1: 累積0点 ラウンド0点 1枚")
		assert.Contains(t, result, "手番: あなた")
		assert.Contains(t, result, "ds")
		assert.Contains(t, result, "dd")
	})

	t.Run("discard top shown", func(t *testing.T) {
		m, _ := setupTongitsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		top := domain.NewCard(domain.CardDesignHeart, 7, false)
		m.On("GetDiscardTop").Return(top)

		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札: HEART 7")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupTongitsCuiMockWithPlayers()
		testErr := errors.New("invalid card index")
		result := p.Output(m, testErr)
		assert.Contains(t, result, "invalid card index")
	})

	t.Run("game ended human winner", func(t *testing.T) {
		m, _ := setupTongitsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "あなたの勝利です！")
	})

	t.Run("game ended CPU winner", func(t *testing.T) {
		m, _ := setupTongitsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1の勝利です！")
	})

	t.Run("draw phase shows CPU current", func(t *testing.T) {
		m, _ := setupTongitsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "手番: CPU 1")
	})

	t.Run("discard phase commands", func(t *testing.T) {
		m, _ := setupTongitsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TongitsPhaseDiscard)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ディスカードフェーズ")
		assert.Contains(t, result, "d <idx>")
		assert.Contains(t, result, "m <札番号...>")
		assert.Contains(t, result, "sp <相手> <メルド番号> <札番号>")
		assert.Contains(t, result, "c             : ドローを宣言")
		assert.Contains(t, result, "現在の残り点: 0")
	})

	t.Run("round end shows next command", func(t *testing.T) {
		m, _ := setupTongitsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TongitsPhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "nr / nextround")
	})

	t.Run("round end with tongits on deal flag", func(t *testing.T) {
		m, _ := setupTongitsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetIsTongits")
		m.On("GetPhase").Return(domain.TongitsPhaseRoundEnd)
		m.On("GetIsTongits").Return(true)

		result := p.Output(m, nil)
		assert.Contains(t, result, "配牌Tongits成立")
	})

	t.Run("round end reveals all melds and CPU hands", func(t *testing.T) {
		m, players := setupTongitsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TongitsPhaseRoundEnd)
		players[0].AppendMeld([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false), domain.NewCard(domain.CardDesignHeart, 7, false), domain.NewCard(domain.CardDesignClover, 7, false)})
		players[0].AppendMeld([]*domain.Card{domain.NewCard(domain.CardDesignClover, 4, false), domain.NewCard(domain.CardDesignClover, 5, false), domain.NewCard(domain.CardDesignClover, 6, false)})
		// The CPU's remaining hand is revealed at round end.
		players[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 12, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "あなた の公開メルド")
		assert.Contains(t, result, "1. [セット]")
		assert.Contains(t, result, "2. [ラン]")
		assert.Contains(t, result, "CPU 1の手札: DIAMOND 12")
	})

	t.Run("CPU hands are not revealed during play", func(t *testing.T) {
		m, players := setupTongitsCuiMockWithPlayers()
		players[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 12, false))

		result := p.Output(m, nil) // default phase is Draw
		assert.NotContains(t, result, "CPU 1の手札:")
	})
}

func TestTongitsCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.TongitsCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockTongitsGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "meld", Detail: "Player 0 melds"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		m.On("GetPlayer", mock.Anything).Return(domain.NewTongitsPlayer(true)).Maybe()

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "meld")
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockTongitsGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
	})
}
