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

// setupCrazyEightsCuiMock creates a MockCrazyEightsGame with sensible defaults.
func setupCrazyEightsCuiMock() *interfaces.MockCrazyEightsGame {
	m := new(interfaces.MockCrazyEightsGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(30)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetChosenSuit").Return(-1)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.CrazyEightsPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// #4737: サーバー計算のヒントを返すようになった。既定は「ヒント無し」。
	m.On("GetHint").Return((*domain.CrazyEightsHint)(nil)).Maybe()

	return m
}

func makeCrazyEightsPlayers() []*domain.CrazyEightsPlayer {
	return []*domain.CrazyEightsPlayer{
		domain.NewCrazyEightsPlayer(true),
		domain.NewCrazyEightsPlayer(false),
		domain.NewCrazyEightsPlayer(false),
		domain.NewCrazyEightsPlayer(false),
	}
}

func setupCrazyEightsCuiMockWithPlayers() (*interfaces.MockCrazyEightsGame, []*domain.CrazyEightsPlayer) {
	m := setupCrazyEightsCuiMock()
	players := makeCrazyEightsPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestCrazyEightsCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.CrazyEightsCuiPresenter)

	t.Run("initial state with header and player info", func(t *testing.T) {
		m, players := setupCrazyEightsCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Crazy Eights (クレイジーエイト)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "山札: 30枚")
		assert.Contains(t, result, "あなた: 累積0点 ラウンド0点 2枚")
		assert.Contains(t, result, "[0]SPADE 1")
		assert.Contains(t, result, "[1]HEART 5")
		assert.Contains(t, result, "CPU 1: 累積0点 ラウンド0点 1枚")
		assert.Contains(t, result, "手番: あなた")
		assert.Contains(t, result, "play <idx>")
		assert.Contains(t, result, "draw")
	})

	t.Run("discard top shown", func(t *testing.T) {
		m, _ := setupCrazyEightsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		top := domain.NewCard(domain.CardDesignHeart, 7, false)
		m.On("GetDiscardTop").Return(top)

		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札: HEART 7")
	})

	t.Run("legal cards are starred on the human's play turn", func(t *testing.T) {
		m, players := setupCrazyEightsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))  // suit match -> legal
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))  // no match -> not legal
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 7, false)) // rank match -> legal
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))  // eight -> always legal

		result := p.Output(m, nil)
		assert.Contains(t, result, "[0]HEART 5*")
		assert.Contains(t, result, "[1]SPADE 3")
		assert.NotContains(t, result, "[1]SPADE 3*")
		assert.Contains(t, result, "[2]CLOVER 7*")
		assert.Contains(t, result, "[3]SPADE 8*")
	})

	t.Run("chosen suit governs legality after an eight is played", func(t *testing.T) {
		m, players := setupCrazyEightsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetChosenSuit")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignSpade, 8, false))
		m.On("GetChosenSuit").Return(domain.CardDesignHeart)
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))  // matches chosen suit -> legal
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))  // spade, not 8 -> not legal
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 8, false)) // eight -> always legal

		result := p.Output(m, nil)
		assert.Contains(t, result, "[0]HEART 2*")
		assert.Contains(t, result, "[1]SPADE 3")
		assert.NotContains(t, result, "[1]SPADE 3*")
		assert.Contains(t, result, "[2]CLOVER 8*")
	})

	t.Run("no legal markers when it is not the human's turn", func(t *testing.T) {
		m, players := setupCrazyEightsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
		m.On("GetCurrentPlayerIdx").Return(1)                                // a CPU is to act
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false)) // would be legal if marked

		result := p.Output(m, nil)
		assert.Contains(t, result, "[0]HEART 5")
		assert.NotContains(t, result, "[0]HEART 5*")
	})

	t.Run("discard top nil hides section", func(t *testing.T) {
		m, _ := setupCrazyEightsCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.NotContains(t, result, "捨て札:")
	})

	t.Run("chosen suit shown", func(t *testing.T) {
		m, _ := setupCrazyEightsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetChosenSuit")
		top := domain.NewCard(domain.CardDesignSpade, 8, false)
		m.On("GetDiscardTop").Return(top)
		m.On("GetChosenSuit").Return(domain.CardDesignHeart)

		result := p.Output(m, nil)
		assert.Contains(t, result, "(指定スート: ♥)")
	})

	t.Run("human with no cards does not print cards line", func(t *testing.T) {
		m, _ := setupCrazyEightsCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.Contains(t, result, "あなた: 累積0点 ラウンド0点 0枚")
		assert.NotContains(t, result, "[0]")
	})

	t.Run("player with scores", func(t *testing.T) {
		m, players := setupCrazyEightsCuiMockWithPlayers()
		players[1].SetCumulativeScore(150)
		players[1].SetRoundScore(30)

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1: 累積150点 ラウンド30点 0枚")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupCrazyEightsCuiMockWithPlayers()
		testErr := errors.New("invalid card index")

		result := p.Output(m, testErr)
		assert.Contains(t, result, "invalid card index")
	})

	t.Run("game ended shows winner human", func(t *testing.T) {
		m, _ := setupCrazyEightsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "あなたの勝利です！")
	})

	t.Run("game ended shows winner CPU", func(t *testing.T) {
		m, _ := setupCrazyEightsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(2)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "CPU 2の勝利です！")
	})

	t.Run("play phase shows current player CPU", func(t *testing.T) {
		m, _ := setupCrazyEightsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "手番: CPU 1")
	})

	t.Run("choose suit phase shows commands", func(t *testing.T) {
		m, _ := setupCrazyEightsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CrazyEightsPhaseChooseSuit)

		result := p.Output(m, nil)
		assert.Contains(t, result, "スート選択フェーズ")
		assert.Contains(t, result, "suit <1-4>")
	})

	t.Run("round end phase shows next command", func(t *testing.T) {
		m, _ := setupCrazyEightsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CrazyEightsPhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "nr / nextround・・・次のラウンドへ")
	})
}

func TestCrazyEightsCuiPresenter_SuitDisplayName(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.CrazyEightsCuiPresenter)

	tests := []struct {
		suit     int
		expected string
	}{
		{domain.CardDesignSpade, "♠"},
		{domain.CardDesignClover, "♣"},
		{domain.CardDesignHeart, "♥"},
		{domain.CardDesignDiamond, "♦"},
		{99, "?"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			m, _ := setupCrazyEightsCuiMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetChosenSuit")
			top := domain.NewCard(domain.CardDesignSpade, 8, false)
			m.On("GetDiscardTop").Return(top)
			m.On("GetChosenSuit").Return(tt.suit)

			result := p.Output(m, nil)
			assert.Contains(t, result, "(指定スート: "+tt.expected+")")
		})
	}
}

func TestCrazyEightsCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.CrazyEightsCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockCrazyEightsGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays SPADE 5"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "play")
		assert.Contains(t, result, "You plays SPADE 5")
		m.AssertExpectations(t)
	})

	t.Run("nil entries", func(t *testing.T) {
		m := new(interfaces.MockCrazyEightsGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockCrazyEightsGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})
}

// **Hearts / Spades はサーバー計算の理由付きヒントを返すのに、CrazyEights には
// HintOutput が無く、全ゲーム共通の簡易ヒューリスティックしか支援が無かった (#4737)。**
func TestCrazyEightsCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CrazyEightsCuiPresenter)

	t.Run("card hint names the card and the reason", func(t *testing.T) {
		m, players := setupCrazyEightsCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		idx := 0
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.CrazyEightsHint{CardIndex: &idx, Reason: "match_suit"})

		out := p.HintOutput(m)
		assert.Contains(t, out, "SPADE 5")
		assert.Contains(t, out, "スートが合う")
	})

	t.Run("suit hint names the suit", func(t *testing.T) {
		m, _ := setupCrazyEightsCuiMockWithPlayers()
		suit := domain.CardDesignHeart
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.CrazyEightsHint{Suit: &suit, Reason: "choose_longest_suit"})

		out := p.HintOutput(m)
		assert.Contains(t, out, "HEART")
		assert.Contains(t, out, "手札に一番多いスート")
	})

	// CardIndex も Suit も無いヒント (ありえないが防御的な分岐) でも、
	// 黙らずに「ヒントなし」と言うこと。
	t.Run("a hint with neither a card nor a suit falls back to none", func(t *testing.T) {
		m, _ := setupCrazyEightsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.CrazyEightsHint{Reason: "play_valid"})

		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})

	// **ヒントが無い側も踏む。**nil を「空文字」で返すと、CUI に何も出ず
	// プレイヤーはコマンドが効いたのか分からない。
	t.Run("no hint says so explicitly", func(t *testing.T) {
		m, _ := setupCrazyEightsCuiMockWithPlayers()
		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})
}
