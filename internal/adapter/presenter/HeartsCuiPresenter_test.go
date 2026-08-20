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

// removeMockCall removes the first expected call matching the given method name.
func removeMockCall(calls []*mock.Call, method string) []*mock.Call {
	result := make([]*mock.Call, 0, len(calls))
	found := false
	for _, c := range calls {
		if !found && c.Method == method {
			found = true
			continue
		}
		result = append(result, c)
	}
	return result
}

// setupHeartsCuiMock creates a MockHeartsGame with sensible defaults for CUI tests.
func setupHeartsCuiMock() *interfaces.MockHeartsGame {
	m := new(interfaces.MockHeartsGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetHeartsBroken").Return(false)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.HeartsPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetPassDirection").Return(domain.HeartsPassLeft)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultHeartsConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeHeartsPlayers() []*domain.HeartsPlayer {
	return []*domain.HeartsPlayer{
		domain.NewHeartsPlayer(true),
		domain.NewHeartsPlayer(false),
		domain.NewHeartsPlayer(false),
		domain.NewHeartsPlayer(false),
	}
}

func setupHeartsCuiMockWithPlayers() (*interfaces.MockHeartsGame, []*domain.HeartsPlayer) {
	m := setupHeartsCuiMock()
	players := makeHeartsPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestHeartsCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.HeartsCuiPresenter)

	t.Run("initial state with header and player info", func(t *testing.T) {
		m, players := setupHeartsCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Hearts (ハーツ)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "トリック: 1")
		assert.Contains(t, result, "ハートブレイク: なし")
		assert.Contains(t, result, "あなた: 累積0点 ラウンド0点 2枚 0トリック")
		assert.Contains(t, result, "[0]SPADE 1")
		assert.Contains(t, result, "[1]HEART 5")
		assert.Contains(t, result, "CPU 1: 累積0点 ラウンド0点 1枚 0トリック")
		assert.Contains(t, result, "手番: あなた")
		assert.Contains(t, result, "play <idx>")
		// Point-limit progress line (default limit 100, all at 0 → leader score 0).
		assert.Contains(t, result, "上限: 100点")
		assert.Contains(t, result, "最多失点:")
	})

	t.Run("progress line names the highest-scoring player", func(t *testing.T) {
		m, players := setupHeartsCuiMockWithPlayers()
		players[2].SetCumulativeScore(42)
		result := p.Output(m, nil)
		assert.Contains(t, result, "最多失点: CPU 2 42点")
	})

	t.Run("cumulative score over 80% of the limit is highlighted", func(t *testing.T) {
		// Colors are enabled here so the ANSI yellow wrapper is observable.
		origNo := color.NoColor()
		color.SetNoColor(false)
		defer color.SetNoColor(origNo)
		m, players := setupHeartsCuiMockWithPlayers()
		players[1].SetCumulativeScore(85) // >= 80% of 100
		players[0].SetCumulativeScore(10) // below threshold
		result := p.Output(m, nil)
		yellow := color.Yellow("85")
		assert.Contains(t, result, yellow)
		assert.NotContains(t, result, color.Yellow("10"))
	})

	t.Run("hearts broken shows あり", func(t *testing.T) {
		m, _ := setupHeartsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHeartsBroken")
		m.On("GetHeartsBroken").Return(true)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ハートブレイク: あり")
	})

	t.Run("player with scores and tricks", func(t *testing.T) {
		m, players := setupHeartsCuiMockWithPlayers()
		players[1].SetCumulativeScore(15)
		players[1].SetRoundScore(7)
		players[1].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, false)})

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1: 累積15点 ラウンド7点 0枚 1トリック")
	})

	t.Run("human with no cards does not print extra newline", func(t *testing.T) {
		m, _ := setupHeartsCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.Contains(t, result, "あなた: 累積0点 ラウンド0点 0枚 0トリック")
	})

	t.Run("human cards with separator", func(t *testing.T) {
		m, players := setupHeartsCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "[0]SPADE 1  [1]DIAMOND 10")
	})

	t.Run("current trick shown", func(t *testing.T) {
		m, _ := setupHeartsCuiMockWithPlayers()
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
		m, _ := setupHeartsCuiMockWithPlayers()

		result := p.Output(m, nil)
		// "トリック: あなた=" would only appear when trick cards are shown
		assert.NotContains(t, result, "トリック: あなた")
		assert.NotContains(t, result, "トリック: CPU")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupHeartsCuiMockWithPlayers()
		testErr := errors.New("invalid card index")

		result := p.Output(m, testErr)
		assert.Contains(t, result, "invalid card index")
	})

	t.Run("game ended shows winner human", func(t *testing.T) {
		m, _ := setupHeartsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！ あなたの勝利です！")
	})

	t.Run("game ended shows winner CPU", func(t *testing.T) {
		m, _ := setupHeartsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(2)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！ CPU 2の勝利です！")
	})

	t.Run("pass phase shows direction and command", func(t *testing.T) {
		m, _ := setupHeartsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.HeartsPhasePass)

		result := p.Output(m, nil)
		assert.Contains(t, result, "パスフェーズ: 左へ渡す")
		assert.Contains(t, result, "pass <idx> <idx> <idx>")
	})

	t.Run("pass phase direction right", func(t *testing.T) {
		m, _ := setupHeartsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPassDirection")
		m.On("GetPhase").Return(domain.HeartsPhasePass)
		m.On("GetPassDirection").Return(domain.HeartsPassRight)

		result := p.Output(m, nil)
		assert.Contains(t, result, "パスフェーズ: 右へ渡す")
	})

	t.Run("pass phase direction across", func(t *testing.T) {
		m, _ := setupHeartsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPassDirection")
		m.On("GetPhase").Return(domain.HeartsPhasePass)
		m.On("GetPassDirection").Return(domain.HeartsPassAcross)

		result := p.Output(m, nil)
		assert.Contains(t, result, "パスフェーズ: 向かいへ渡す")
	})

	t.Run("pass phase direction none", func(t *testing.T) {
		m, _ := setupHeartsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPassDirection")
		m.On("GetPhase").Return(domain.HeartsPhasePass)
		m.On("GetPassDirection").Return(domain.HeartsPassNone)

		result := p.Output(m, nil)
		assert.Contains(t, result, "パスフェーズ: 交換なし")
	})

	t.Run("pass phase direction unknown", func(t *testing.T) {
		m, _ := setupHeartsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPassDirection")
		m.On("GetPhase").Return(domain.HeartsPhasePass)
		m.On("GetPassDirection").Return(domain.HeartsPassDirection(99))

		result := p.Output(m, nil)
		assert.Contains(t, result, "パスフェーズ: 不明")
	})

	t.Run("play phase shows current player CPU", func(t *testing.T) {
		m, _ := setupHeartsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "手番: CPU 1")
		assert.Contains(t, result, "play <idx>")
	})

	t.Run("trick end phase shows next command", func(t *testing.T) {
		m, _ := setupHeartsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.HeartsPhaseTrickEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
		assert.Contains(t, result, "next・・・次のトリックへ")
	})

	t.Run("round end phase shows next command", func(t *testing.T) {
		m, _ := setupHeartsCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.HeartsPhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "nr / nextround・・・次のラウンドへ")
	})

	t.Run("nil player at winnerIdx shows UNKNOWN", func(t *testing.T) {
		m := setupHeartsCuiMock()
		m.On("GetPlayerCnt").Return(1)
		players := makeHeartsPlayers()
		m.On("GetPlayer", 0).Return(players[0])
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(99)
		m.On("GetPlayer", 99).Return((*domain.HeartsPlayer)(nil))

		result := p.Output(m, nil)
		assert.Contains(t, result, "UNKNOWN")
	})
}

func TestHeartsCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.HeartsCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockHeartsGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played SPADE 5"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		m.On("GetPlayer", mock.Anything).Return(domain.NewHeartsPlayer(true)).Maybe()

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "play")
		assert.Contains(t, result, "played SPADE 5")
		m.AssertExpectations(t)
	})

	t.Run("nil entries", func(t *testing.T) {
		m := new(interfaces.MockHeartsGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		m.On("GetPlayer", mock.Anything).Return(domain.NewHeartsPlayer(true)).Maybe()

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockHeartsGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})
}

func TestHeartsCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("no hint", func(t *testing.T) {
		m := new(interfaces.MockHeartsGame)
		m.On("GetHint").Return((*domain.HeartsHint)(nil))

		p := new(presenter.HeartsCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("play hint", func(t *testing.T) {
		m := new(interfaces.MockHeartsGame)
		m.On("GetHint").Return(&domain.HeartsHint{
			CardIndices: []int{2},
			Reason:      "follow_suit",
		})
		player := domain.NewHeartsPlayer(true)
		player.Reset()
		player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
		player.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
		m.On("GetPlayer", 0).Return(player)

		p := new(presenter.HeartsCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
		assert.Contains(t, result, "リードスートに追随")
	})

	t.Run("pass hint", func(t *testing.T) {
		m := new(interfaces.MockHeartsGame)
		m.On("GetHint").Return(&domain.HeartsHint{
			CardIndices: []int{0, 1, 2},
			Reason:      "pass_high_risk_cards",
		})
		player := domain.NewHeartsPlayer(true)
		player.Reset()
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		player.AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))
		player.AddCard(domain.NewCard(domain.CardDesignClover, 11, false))
		m.On("GetPlayer", 0).Return(player)

		p := new(presenter.HeartsCuiPresenter)
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
		assert.Contains(t, result, "リスクの高いカードを渡す")
	})

	t.Run("hint reason strings", func(t *testing.T) {
		reasons := map[string]string{
			"lead_low":             "低いカードでリード",
			"discard_queen_spades": "Q♠を捨てるチャンス",
			"discard_hearts":       "ハートを捨てる",
			"discard_high":         "高いカードを捨てる",
			"unknown_reason":       "unknown_reason",
		}
		for key, expected := range reasons {
			m := new(interfaces.MockHeartsGame)
			m.On("GetHint").Return(&domain.HeartsHint{
				CardIndices: []int{0},
				Reason:      key,
			})
			player := domain.NewHeartsPlayer(true)
			player.Reset()
			player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
			m.On("GetPlayer", 0).Return(player)

			p := new(presenter.HeartsCuiPresenter)
			result := p.HintOutput(m)
			assert.Contains(t, result, expected, "reason: "+key)
		}
	})
}
