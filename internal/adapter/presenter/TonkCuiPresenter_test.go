//go:build test

package presenter_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupTonkCuiMock() *interfaces.MockTonkGame {
	m := new(interfaces.MockTonkGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(41)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TonkPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetIsTonk").Return(false)
	m.On("GetKnockerMelds").Return(([][]*domain.Card)(nil)).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// #4750: ディスカード表示が最小デッドウッドを引くようになった。
	m.On("GetBestDeadwood", 0).Return(3, 0).Maybe()

	return m
}

func makeTonkPlayers() []*domain.TonkPlayer {
	return []*domain.TonkPlayer{
		domain.NewTonkPlayer(true),
		domain.NewTonkPlayer(false),
	}
}

func setupTonkCuiMockWithPlayers() (*interfaces.MockTonkGame, []*domain.TonkPlayer) {
	m := setupTonkCuiMock()
	players := makeTonkPlayers()
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestTonkCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.TonkCuiPresenter)

	t.Run("initial state with header and player info", func(t *testing.T) {
		m, players := setupTonkCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Tonk (トンク)")
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
		m, _ := setupTonkCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		top := domain.NewCard(domain.CardDesignHeart, 7, false)
		m.On("GetDiscardTop").Return(top)

		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札: HEART 7")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupTonkCuiMockWithPlayers()
		testErr := errors.New("invalid card index")
		result := p.Output(m, testErr)
		assert.Contains(t, result, "invalid card index")
	})

	t.Run("game ended human winner", func(t *testing.T) {
		m, _ := setupTonkCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "あなたの勝利です！")
	})

	t.Run("game ended CPU winner", func(t *testing.T) {
		m, _ := setupTonkCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1の勝利です！")
	})

	t.Run("draw phase shows CPU current", func(t *testing.T) {
		m, _ := setupTonkCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "手番: CPU 1")
	})

	t.Run("discard phase commands", func(t *testing.T) {
		m, _ := setupTonkCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TonkPhaseDiscard)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ディスカードフェーズ")
		assert.Contains(t, result, "d <idx>")
		assert.Contains(t, result, "k <idx>")
	})

	t.Run("discard phase shows knockable for a low hand", func(t *testing.T) {
		m, players := setupTonkCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TonkPhaseDiscard)
		// **計算はドメインの仕事になった (#4750)。**ここで確かめるのは
		// 「閾値と比べて正しい文言を選ぶこと」だけ。値そのものの正しさは
		// TestTonk_GetBestDeadwood が実際の手札で見ている。
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetBestDeadwood")
		m.On("GetBestDeadwood", 0).Return(domain.TonkKnockThreshold, 0)
		_ = players

		result := p.Output(m, nil)
		assert.Contains(t, result, "最小デッドウッド(1枚捨て後):")
		assert.Contains(t, result, "ノック可能")
	})

	t.Run("discard phase shows not-knockable for a high hand", func(t *testing.T) {
		m, players := setupTonkCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TonkPhaseDiscard)
		// 閾値のすぐ上。境界 (== 閾値) は上の subtest が押さえている。
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetBestDeadwood")
		m.On("GetBestDeadwood", 0).Return(domain.TonkKnockThreshold+1, 0)
		_ = players

		result := p.Output(m, nil)
		assert.Contains(t, result, "最小デッドウッド(1枚捨て後):")
		assert.Contains(t, result, "ノック不可")
		assert.NotContains(t, result, "ノック可能")
	})

	t.Run("round end shows next command", func(t *testing.T) {
		m, _ := setupTonkCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TonkPhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "nr / nextround")
	})

	t.Run("round end with tonk on deal flag", func(t *testing.T) {
		m, _ := setupTonkCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetIsTonk")
		m.On("GetPhase").Return(domain.TonkPhaseRoundEnd)
		m.On("GetIsTonk").Return(true)

		result := p.Output(m, nil)
		assert.Contains(t, result, "配牌Tonk成立")
	})

	t.Run("round end reveals knocker melds and CPU hands", func(t *testing.T) {
		m, players := setupTonkCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetKnockerMelds")
		m.On("GetPhase").Return(domain.TonkPhaseRoundEnd)
		m.On("GetKnockerMelds").Return([][]*domain.Card{
			{ // set of 7s
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignClover, 7, false),
			},
			{ // run of clubs 4-5-6
				domain.NewCard(domain.CardDesignClover, 4, false),
				domain.NewCard(domain.CardDesignClover, 5, false),
				domain.NewCard(domain.CardDesignClover, 6, false),
			},
		})
		// The CPU's remaining hand is revealed at round end.
		players[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 12, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "[ノッカーのメルド]")
		assert.Contains(t, result, "メルド1(セット):")
		assert.Contains(t, result, "メルド2(ラン):")
		assert.Contains(t, result, "CPU 1の手札: DIAMOND 12")
	})

	t.Run("CPU hands are not revealed during play", func(t *testing.T) {
		m, players := setupTonkCuiMockWithPlayers()
		players[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 12, false))

		result := p.Output(m, nil) // default phase is Draw
		assert.NotContains(t, result, "CPU 1の手札:")
	})
}

func TestTonkCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.TonkCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockTonkGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "knock", Detail: "Player 0 knocks"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "knock")
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockTonkGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
	})
}

// #5582: 相手の残りが少ないほどノックは裏目 (#1939)。Web はボタンに警告リングと
// ⚠️ を出しているのに、CUI は各行の枚数を見比べさせるだけだった。
func TestTonkCuiPresenter_WarnsAboutTheUndercutRisk(t *testing.T) {
	i18n.SetLang("ja")
	build := func(oppCards int) string {
		m, players := setupTonkCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TonkPhaseDiscard)
		for range oppCards {
			players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 3, true))
		}
		return new(presenter.TonkCuiPresenter).Output(m, nil)
	}

	// 境界値。閾値ちょうどでは出し、1 枚上では出さない (受け入れ条件3)。
	assert.Contains(t, build(domain.TonkUndercutRiskMax),
		i18n.Tf("tonk.knockUndercutWarning", "count", strconv.Itoa(domain.TonkUndercutRiskMax)))
	assert.NotContains(t, build(domain.TonkUndercutRiskMax+1),
		i18n.Tf("tonk.knockUndercutWarning", "count", strconv.Itoa(domain.TonkUndercutRiskMax+1)))

	// **1 枚でも出ること。**「ちょうど 2 枚」だけを見る実装では通らない。
	assert.Contains(t, build(1), i18n.Tf("tonk.knockUndercutWarning", "count", "1"))
}

// ドロー中は出さない。ノックできない局面で「ノックは危ない」と言っても仕方がない。
func TestTonkCuiPresenter_DoesNotWarnOutsideTheDiscardPhase(t *testing.T) {
	i18n.SetLang("ja")
	m, players := setupTonkCuiMockWithPlayers()
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 3, true))

	out := new(presenter.TonkCuiPresenter).Output(m, nil)
	assert.NotContains(t, out, i18n.Tf("tonk.knockUndercutWarning", "count", "1"))
}

// レビュー (#5941) の指摘: ノックを決めるのは人間なので、CPU の捨て札中に
// 「相手の手札が少ない」と警告しても行動できない。上のデッドウッド表示と同じ条件。
func TestTonkCuiPresenter_WarnsOnlyOnTheHumanTurn(t *testing.T) {
	i18n.SetLang("ja")
	m, players := setupTonkCuiMockWithPlayers()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
	m.On("GetPhase").Return(domain.TonkPhaseDiscard)
	m.On("GetCurrentPlayerIdx").Return(1) // CPU の捨て札中
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 3, true))

	out := new(presenter.TonkCuiPresenter).Output(m, nil)
	assert.NotContains(t, out, i18n.Tf("tonk.knockUndercutWarning", "count", "1"))
}
