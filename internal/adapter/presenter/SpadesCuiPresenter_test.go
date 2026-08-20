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

// setupSpadesCuiMock creates a MockSpadesGame with sensible defaults for CUI tests.
func setupSpadesCuiMock() *interfaces.MockSpadesGame {
	m := new(interfaces.MockSpadesGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetSpadesBroken").Return(false)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SpadesPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultSpadesConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeSpadesPlayers() []*domain.SpadesPlayer {
	return []*domain.SpadesPlayer{
		domain.NewSpadesPlayer(true),
		domain.NewSpadesPlayer(false),
		domain.NewSpadesPlayer(false),
		domain.NewSpadesPlayer(false),
	}
}

func setupSpadesCuiMockWithPlayers() (*interfaces.MockSpadesGame, []*domain.SpadesPlayer) {
	m := setupSpadesCuiMock()
	players := makeSpadesPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestSpadesCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.SpadesCuiPresenter)

	t.Run("initial state with header and player info", func(t *testing.T) {
		m, players := setupSpadesCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Spades (スペード)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "トリック: 1")
		assert.Contains(t, result, "スペードブレイク: なし")
		assert.Contains(t, result, "あなた: ビッド=未ビッド 獲得0トリック バッグ0 累積0点 ラウンド0点 2枚")
		assert.Contains(t, result, "[0]SPADE 1")
		assert.Contains(t, result, "[1]HEART 5")
		assert.Contains(t, result, "CPU 1: ビッド=未ビッド 獲得0トリック バッグ0 累積0点 ラウンド0点 1枚")
		assert.Contains(t, result, "手番: あなた")
		assert.Contains(t, result, "play <idx>")
		// Point-limit progress line (default limit 500, all scores at 0).
		assert.Contains(t, result, "上限: 500点")
		assert.Contains(t, result, "首位:")
	})

	t.Run("bag count is colored as it nears the penalty threshold", func(t *testing.T) {
		origNo := color.NoColor()
		color.SetNoColor(false)
		defer color.SetNoColor(origNo)
		m, players := setupSpadesCuiMockWithPlayers()
		players[0].SetBags(9) // threshold 10, remaining 1 -> red
		players[1].SetBags(8) // remaining 2 -> yellow
		players[2].SetBags(3) // remaining 7 -> plain

		result := p.Output(m, nil)
		assert.Contains(t, result, color.Red("9"))
		assert.Contains(t, result, color.Yellow("8"))
		assert.NotContains(t, result, color.Red("3"))
		assert.NotContains(t, result, color.Yellow("3"))
	})

	t.Run("spades broken shows あり", func(t *testing.T) {
		m, _ := setupSpadesCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetSpadesBroken")
		m.On("GetSpadesBroken").Return(true)

		result := p.Output(m, nil)
		assert.Contains(t, result, "スペードブレイク: あり")
	})

	t.Run("player with scores and tricks and bags", func(t *testing.T) {
		m, players := setupSpadesCuiMockWithPlayers()
		players[1].SetCumulativeScore(150)
		players[1].SetRoundScore(30)
		players[1].SetBags(3)
		players[1].SetBid(4)
		players[1].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, false)})

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1: ビッド=4 獲得1トリック バッグ3 累積150点 ラウンド30点 0枚")
	})

	t.Run("human with no cards does not print extra cards line", func(t *testing.T) {
		m, _ := setupSpadesCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.Contains(t, result, "あなた: ビッド=未ビッド 獲得0トリック バッグ0 累積0点 ラウンド0点 0枚")
	})

	t.Run("human cards with separator", func(t *testing.T) {
		m, players := setupSpadesCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "[0]SPADE 1  [1]DIAMOND 10")
	})

	t.Run("current trick shown", func(t *testing.T) {
		m, _ := setupSpadesCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
		}
		m.On("GetCurrentTrick").Return(trick)

		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック: あなた=CLOVER 3, CPU 1=CLOVER 7")
	})

	t.Run("no trick cards hides trick section", func(t *testing.T) {
		m, _ := setupSpadesCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.NotContains(t, result, "トリック: あなた")
		assert.NotContains(t, result, "トリック: CPU")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupSpadesCuiMockWithPlayers()
		testErr := errors.New("invalid card index")

		result := p.Output(m, testErr)
		assert.Contains(t, result, "invalid card index")
	})

	t.Run("game ended shows winner human", func(t *testing.T) {
		m, _ := setupSpadesCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "あなたの勝利です！")
	})

	t.Run("game ended shows winner CPU", func(t *testing.T) {
		m, _ := setupSpadesCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(2)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "CPU 2の勝利です！")
	})

	t.Run("bid phase shows bidder and command", func(t *testing.T) {
		m, _ := setupSpadesCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SpadesPhaseBid)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ビッドフェーズ: あなたの番")
		assert.Contains(t, result, "b <n>")
	})

	t.Run("bid phase CPU bidder", func(t *testing.T) {
		m, _ := setupSpadesCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetBidPlayerIdx")
		m.On("GetPhase").Return(domain.SpadesPhaseBid)
		m.On("GetBidPlayerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ビッドフェーズ: CPU 1の番")
	})

	t.Run("play phase shows current player CPU", func(t *testing.T) {
		m, _ := setupSpadesCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "手番: CPU 1")
		assert.Contains(t, result, "play <idx>")
	})

	t.Run("trick end phase shows next command", func(t *testing.T) {
		m, _ := setupSpadesCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SpadesPhaseTrickEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
		assert.Contains(t, result, "next・・・次のトリックへ")
	})

	t.Run("round end phase shows next command", func(t *testing.T) {
		m, _ := setupSpadesCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SpadesPhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "nr / nextround・・・次のラウンドへ")
	})

	t.Run("nil player at winnerIdx shows UNKNOWN", func(t *testing.T) {
		m := setupSpadesCuiMock()
		m.On("GetPlayerCnt").Return(1)
		players := makeSpadesPlayers()
		m.On("GetPlayer", 0).Return(players[0])
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(99)
		m.On("GetPlayer", 99).Return((*domain.SpadesPlayer)(nil))

		result := p.Output(m, nil)
		assert.Contains(t, result, "UNKNOWN")
	})
}

func TestSpadesCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.SpadesCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockSpadesGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played SPADE 5"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		m.On("GetPlayer", mock.Anything).Return(domain.NewSpadesPlayer(true)).Maybe()

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "play")
		assert.Contains(t, result, "あなた", "棋譜の座席名が他の行と揃っていない")
		assert.Contains(t, result, "played SPADE 5")
		m.AssertExpectations(t)
	})

	t.Run("nil entries", func(t *testing.T) {
		m := new(interfaces.MockSpadesGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockSpadesGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})
}

func TestSpadesCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("no hint", func(t *testing.T) {
		m := new(interfaces.MockSpadesGame)
		m.On("GetHint").Return((*domain.SpadesHint)(nil))

		p := new(presenter.SpadesCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("bid hint", func(t *testing.T) {
		bid := 3
		m := new(interfaces.MockSpadesGame)
		m.On("GetHint").Return(&domain.SpadesHint{
			Bid:    &bid,
			Reason: "strategic_bid",
		})

		p := new(presenter.SpadesCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
		assert.Contains(t, result, "ビッド 3")
		assert.Contains(t, result, "戦略的なビッド")
	})

	t.Run("hint with nil bid and nil card index", func(t *testing.T) {
		m := new(interfaces.MockSpadesGame)
		m.On("GetHint").Return(&domain.SpadesHint{
			Reason: "unknown",
		})

		p := new(presenter.SpadesCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("play hint", func(t *testing.T) {
		idx := 1
		m := new(interfaces.MockSpadesGame)
		m.On("GetHint").Return(&domain.SpadesHint{
			CardIndex: &idx,
			Reason:    "follow_suit",
		})
		player := domain.NewSpadesPlayer(true)
		player.Reset()
		player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
		m.On("GetPlayer", 0).Return(player)

		p := new(presenter.SpadesCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
		assert.Contains(t, result, "リードスートに追随")
	})

	t.Run("hint reason strings", func(t *testing.T) {
		reasons := map[string]string{
			"lead_strong":    "強いカードでリード",
			"lead_low":       "低いカードでリード",
			"trump_cut":      "スペードでカット",
			"discard_high":   "高いカードを捨てる",
			"unknown_reason": "unknown_reason",
		}
		for key, expected := range reasons {
			idx := 0
			m := new(interfaces.MockSpadesGame)
			m.On("GetHint").Return(&domain.SpadesHint{
				CardIndex: &idx,
				Reason:    key,
			})
			player := domain.NewSpadesPlayer(true)
			player.Reset()
			player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
			m.On("GetPlayer", 0).Return(player)

			p := new(presenter.SpadesCuiPresenter)
			result := p.HintOutput(m)
			assert.Contains(t, result, expected, "reason: "+key)
		}
	})
}
