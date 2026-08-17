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

func setupSevenBridgeCuiMock() *interfaces.MockSevenBridgeGame {
	m := new(interfaces.MockSevenBridgeGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(37)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SevenBridgePhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetClaimedThisTurn").Return(false).Maybe()
	return m
}

func makeSevenBridgePlayers() []*domain.SevenBridgePlayer {
	return []*domain.SevenBridgePlayer{
		domain.NewSevenBridgePlayer(true),
		domain.NewSevenBridgePlayer(false),
	}
}

func setupSevenBridgeCuiMockWithPlayers() (*interfaces.MockSevenBridgeGame, []*domain.SevenBridgePlayer) {
	m := setupSevenBridgeCuiMock()
	players := makeSevenBridgePlayers()
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestSevenBridgeCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SevenBridgeCuiPresenter)

	t.Run("initial header/state", func(t *testing.T) {
		m, players := setupSevenBridgeCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Seven Bridge (セブンブリッジ)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "山札: 37枚")
		assert.Contains(t, result, "あなた: 累積0点 ラウンド0点 2枚")
		assert.Contains(t, result, "CPU 1: 累積0点 ラウンド0点 1枚")
		assert.Contains(t, result, "手番: あなた")
		assert.Contains(t, result, "ds")
		assert.Contains(t, result, "pon")
		assert.Contains(t, result, "chi")
	})

	t.Run("discard top shown", func(t *testing.T) {
		m, _ := setupSevenBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札: HEART 7")
	})

	t.Run("meld rendered for player", func(t *testing.T) {
		m, players := setupSevenBridgeCuiMockWithPlayers()
		players[0].AppendMeld([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignClover, 3, false),
			domain.NewCard(domain.CardDesignHeart, 3, false),
		})

		result := p.Output(m, nil)
		assert.Contains(t, result, "場: ")
		assert.Contains(t, result, "SPADE 3")
	})

	t.Run("error shown", func(t *testing.T) {
		m, _ := setupSevenBridgeCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})

	t.Run("game ended → winner human", func(t *testing.T) {
		m, _ := setupSevenBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "あなたの勝利です！")
	})

	t.Run("game ended → winner CPU", func(t *testing.T) {
		m, _ := setupSevenBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1の勝利です！")
	})

	t.Run("play phase commands", func(t *testing.T) {
		m, _ := setupSevenBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SevenBridgePhasePlay)

		result := p.Output(m, nil)
		assert.Contains(t, result, "プレイフェーズ")
		assert.Contains(t, result, "m <idx")
		assert.Contains(t, result, "lo <pIdx>")
		assert.Contains(t, result, "d <idx>")
	})

	t.Run("round end phase", func(t *testing.T) {
		m, _ := setupSevenBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SevenBridgePhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "nr / nextround")
	})

	t.Run("draw phase CPU turn", func(t *testing.T) {
		m, _ := setupSevenBridgeCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "手番: CPU 1")
	})
}

func TestSevenBridgeCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.SevenBridgeCuiPresenter)

	newHintMock := func() *interfaces.MockSevenBridgeGame {
		m := new(interfaces.MockSevenBridgeGame)
		human := domain.NewSevenBridgePlayer(true)
		for _, v := range []int{3, 3, 3, 8} {
			human.AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		m.On("GetPhase").Return(domain.SevenBridgePhasePlay)
		m.On("IsHumanTurn").Return(true)
		m.On("GetCurrentPlayerIdx").Return(0)
		m.On("GetPlayer", 0).Return(human)
		return m
	}

	t.Run("recommends a meld when one is available", func(t *testing.T) {
		m := newHintMock()
		m.On("SuggestMeld", 0).Return([]int{0, 1, 2})
		assert.Contains(t, p.HintOutput(m), "メルド")
	})

	t.Run("recommends a discard when no meld is available", func(t *testing.T) {
		m := newHintMock()
		m.On("SuggestMeld", 0).Return(([]int)(nil))
		m.On("SuggestDiscard", 0).Return(3)
		assert.Contains(t, p.HintOutput(m), "捨てる")
	})

	t.Run("declines outside the human's play phase", func(t *testing.T) {
		m := new(interfaces.MockSevenBridgeGame)
		m.On("GetPhase").Return(domain.SevenBridgePhaseDraw)
		// **ドローフェーズでも人間の手番でなければ断る。**手番のときは
		// ポン・チーを案内するようになった (#4904)。
		m.On("IsHumanTurn").Return(false)
		assert.Contains(t, p.HintOutput(m), "プレイフェーズではありません")
	})

	// **ドローフェーズこそ一番迷う。**ポン・チーの判断を無支援にしていた (#4904)。
	t.Run("advises pon, chi or drawing during the draw phase", func(t *testing.T) {
		build := func(pon, chi []int) *interfaces.MockSevenBridgeGame {
			m := new(interfaces.MockSevenBridgeGame)
			m.On("GetPhase").Return(domain.SevenBridgePhaseDraw)
			m.On("IsHumanTurn").Return(true)
			m.On("GetCurrentPlayerIdx").Return(0)
			human := domain.NewSevenBridgePlayer(true)
			human.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
			human.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
			m.On("GetPlayer", 0).Return(human)
			m.On("SuggestPon", 0).Return(pon)
			m.On("SuggestChi", 0).Return(chi)
			return m
		}

		// **ポンが先。**同ランク 3 枚は連番より確実に面子になる。
		both := p.HintOutput(build([]int{0, 1}, []int{0, 1}))
		assert.Contains(t, both, "ポンできます")
		assert.NotContains(t, both, "チーできます")

		assert.Contains(t, p.HintOutput(build(nil, []int{0, 1})), "チーできます")

		// どちらも無ければ山札を勧める。黙らない。
		none := p.HintOutput(build(nil, nil))
		assert.Contains(t, none, "山札から引きましょう")
		assert.NotContains(t, none, "プレイフェーズではありません")
	})
}

func TestSevenBridgeCuiPresenter_ActionLogOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SevenBridgeCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockSevenBridgeGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "draw_stock", Detail: "You draws from stock"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "draw_stock")
		m.AssertExpectations(t)
	})

	t.Run("no entries", func(t *testing.T) {
		m := new(interfaces.MockSevenBridgeGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockSevenBridgeGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
		m.AssertExpectations(t)
	})
}

// #5547: ポン/チーで捨て札を割り込んで取ったのか、素直に山から引いたのかが
// 画面から判別できなかった。`GetClaimedThisTurn()` は保存までされているのに
// どちらの UI も一度も読んでいない。
func TestSevenBridgeCuiPresenter_Output_ClaimedThisTurn(t *testing.T) {
	p := new(presenter.SevenBridgeCuiPresenter)

	outputWith := func(claimed bool) string {
		m := new(interfaces.MockSevenBridgeGame)
		m.On("GetClaimedThisTurn").Return(claimed)
		m.On("GetRoundNumber").Return(1)
		m.On("GetDrawPileCount").Return(37)
		m.On("GetDiscardTop").Return((*domain.Card)(nil))
		m.On("GetGameEndFlag").Return(false)
		m.On("GetPhase").Return(domain.SevenBridgePhasePlay)
		m.On("GetCurrentPlayerIdx").Return(0)
		m.On("GetWinnerIdx").Return(-1)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		players := makeSevenBridgePlayers()
		m.On("GetPlayerCnt").Return(2)
		m.On("GetPlayer", 0).Return(players[0])
		m.On("GetPlayer", 1).Return(players[1])
		return p.Output(m, nil)
	}

	assert.Contains(t, outputWith(true), i18n.T("sevenbridge.claimedThisTurn"))
	// **山から引いたターンでは出さない。**毎ターン出ると区別にならない。
	assert.NotContains(t, outputWith(false), i18n.T("sevenbridge.claimedThisTurn"))
}
