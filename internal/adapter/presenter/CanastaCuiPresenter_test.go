//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupCanastaCuiMock() *interfaces.MockCanastaGame {
	m := new(interfaces.MockCanastaGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(54)
	m.On("GetDiscardPileCount").Return(0)
	m.On("GetDiscardPile").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetIsFrozen").Return(false)
	m.On("GetDrawFromDiscardBlocker").Return("").Maybe()
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.CanastaPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupCanastaCuiMockWithPlayers() (*interfaces.MockCanastaGame, []*domain.CanastaPlayer) {
	m := setupCanastaCuiMock()
	players := makeCanastaPlayers()
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

// **山ごと取るゲームなので捨て札の中身は公開情報 (#5043)。**Burraco (#4833) と
// 同じ制限が残っていた。
func TestCanastaCuiPresenter_ListsTheDiscardPile(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CanastaCuiPresenter)

	withPile := func(n int) *interfaces.MockCanastaGame {
		m, _ := setupCanastaCuiMockWithPlayers()
		pile := make([]*domain.Card, 0, n)
		for i := 0; i < n; i++ {
			pile = append(pile, domain.NewCard(domain.CardDesignHeart, i%13+1, false))
		}
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardPile")
		if n > 0 {
			m.On("GetDiscardTop").Return(pile[len(pile)-1])
		} else {
			m.On("GetDiscardTop").Return((*domain.Card)(nil))
		}
		m.On("GetDiscardPile").Return(pile)
		return m
	}

	t.Run("lists every card with its index", func(t *testing.T) {
		out := p.Output(withPile(3), nil)
		assert.Contains(t, out, "山の中身:")
		assert.Contains(t, out, "[0]HEART 1")
		assert.Contains(t, out, "[2]HEART 3")
	})

	t.Run("wraps a long pile over several lines", func(t *testing.T) {
		out := p.Output(withPile(20), nil)
		assert.Equal(t, 3, strings.Count(out, "山の中身:"), "8 枚ごとに折り返す")
	})

	t.Run("says nothing when the pile is empty", func(t *testing.T) {
		assert.NotContains(t, p.Output(withPile(0), nil), "山の中身:")
	})
}

func TestCanastaCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.CanastaCuiPresenter)

	t.Run("initial state with header and player info", func(t *testing.T) {
		m, players := setupCanastaCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Canasta (カナスタ)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "山札: 54枚")
		assert.Contains(t, result, "あなた: 累積0点 ラウンド0点 1枚")
		assert.Contains(t, result, "[0]SPADE 5")
		assert.Contains(t, result, "CPU 1: 累積0点 ラウンド0点 1枚")
		assert.Contains(t, result, "手番: あなた")
		assert.Contains(t, result, "ds")
		assert.Contains(t, result, "dd")
	})

	t.Run("frozen pile shown", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetIsFrozen")
		m.On("GetIsFrozen").Return(true)

		result := p.Output(m, nil)
		assert.Contains(t, result, "[フリーズ]")
		// Draw phase with a frozen pile also shows the pickup-restriction note.
		assert.Contains(t, result, "捨て札は凍結中")
	})

	t.Run("meld phase shows the score-tiered initial-meld minimum", func(t *testing.T) {
		m, players := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CanastaPhaseMeld)
		// Human (player 0), no initial meld, cumulative 0 -> 50-point band.
		players[0].SetCumulativeScore(0)
		players[0].SetHasInitMeld(false)

		result := p.Output(m, nil)
		assert.Contains(t, result, "初回メルド最低点: 50点")
	})

	t.Run("meld phase minimum rises with the score band", func(t *testing.T) {
		m, players := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CanastaPhaseMeld)
		players[0].SetCumulativeScore(1500) // 1500-2999 band -> 90
		players[0].SetHasInitMeld(false)

		result := p.Output(m, nil)
		assert.Contains(t, result, "初回メルド最低点: 90点")
	})

	t.Run("meld phase minimum is 15 for a negative score", func(t *testing.T) {
		m, players := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CanastaPhaseMeld)
		players[0].SetCumulativeScore(-100) // negative band -> 15
		players[0].SetHasInitMeld(false)

		result := p.Output(m, nil)
		assert.Contains(t, result, "初回メルド最低点: 15点")
	})

	t.Run("meld phase minimum is 120 at 3000+", func(t *testing.T) {
		m, players := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CanastaPhaseMeld)
		players[0].SetCumulativeScore(3000) // 3000+ band -> 120
		players[0].SetHasInitMeld(false)

		result := p.Output(m, nil)
		assert.Contains(t, result, "初回メルド最低点: 120点")
	})

	t.Run("meld phase hides the minimum once the initial meld is done", func(t *testing.T) {
		m, players := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CanastaPhaseMeld)
		players[0].SetHasInitMeld(true)

		result := p.Output(m, nil)
		assert.NotContains(t, result, "初回メルド最低点")
	})

	t.Run("discard top shown", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		top := domain.NewCard(domain.CardDesignHeart, 7, false)
		m.On("GetDiscardTop").Return(top)

		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札: HEART 7")
	})

	t.Run("discard top nil hides section", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.NotContains(t, result, "捨て札:")
	})

	t.Run("player with scores", func(t *testing.T) {
		m, players := setupCanastaCuiMockWithPlayers()
		players[1].SetCumulativeScore(300)
		players[1].SetRoundScore(100)

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1: 累積300点 ラウンド100点 0枚")
	})

	t.Run("player with meld shown", func(t *testing.T) {
		m, players := setupCanastaCuiMockWithPlayers()
		meld := &domain.CanastaMeld{
			Cards: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignClover, 7, false),
			},
			IsNatural: true,
		}
		players[0].SetMelds([]*domain.CanastaMeld{meld})

		result := p.Output(m, nil)
		assert.Contains(t, result, "ナチュラル")
		assert.Contains(t, result, "SPADE 7")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()
		testErr := errors.New("invalid card index")

		result := p.Output(m, testErr)
		assert.Contains(t, result, "invalid card index")
	})

	t.Run("game ended shows winner human", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "あなたの勝利です！")
	})

	t.Run("game ended shows winner CPU", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "CPU 1の勝利です！")
	})

	t.Run("draw phase shows current player CPU", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "手番: CPU 1")
	})

	t.Run("meld phase shows commands", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CanastaPhaseMeld)

		result := p.Output(m, nil)
		assert.Contains(t, result, "メルドフェーズ")
		assert.Contains(t, result, "m ")
		assert.Contains(t, result, "sm")
	})

	t.Run("discard phase shows commands", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CanastaPhaseDiscard)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ディスカードフェーズ")
		assert.Contains(t, result, "d <idx>")
		assert.Contains(t, result, "go")
	})

	t.Run("round end phase shows next command", func(t *testing.T) {
		m, _ := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CanastaPhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "nr / nextround")
	})

	t.Run("red 3 tag shown", func(t *testing.T) {
		m, players := setupCanastaCuiMockWithPlayers()
		players[0].AddRed3(domain.NewCard(domain.CardDesignHeart, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "赤3: 1枚")
	})

	t.Run("canasta star tag shown when player holds a canasta meld", func(t *testing.T) {
		m, players := setupCanastaCuiMockWithPlayers()
		// A 7-card natural meld of 4s satisfies HasCanasta() (>=7 cards).
		// IsNatural=true also exercises the "ナチュラル" meld-type label;
		// the m.IsCanasta() branch attaches "カナスタ" to that label.
		cards := make([]*domain.Card, 7)
		for i := range cards {
			cards[i] = domain.NewCard(domain.CardDesignSpade, 4, false)
		}
		players[0].AddMeld(&domain.CanastaMeld{Cards: cards, IsNatural: true})

		result := p.Output(m, nil)
		assert.Contains(t, result, "★カナスタ")
		assert.Contains(t, result, "ナチュラルカナスタ")
	})
}

func TestCanastaCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.CanastaCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockCanastaGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "draw_stock", Detail: "drew from stock"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "draw_stock")
		m.AssertExpectations(t)
	})

	t.Run("game not ended returns empty", func(t *testing.T) {
		m := new(interfaces.MockCanastaGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.NotEmpty(t, result)
		m.AssertExpectations(t)
	})
}

// #5502: 取れない理由をインデックスを打つ前に出す。
func TestCanastaCuiPresenter_DrawBlockerLine(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CanastaCuiPresenter)

	withBlocker := func(code string) *interfaces.MockCanastaGame {
		m, _ := setupCanastaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDrawFromDiscardBlocker")
		m.On("GetDrawFromDiscardBlocker").Return(code)
		return m
	}

	t.Run("names the reason the pile cannot be taken", func(t *testing.T) {
		out := p.Output(withBlocker(domain.CanastaDrawBlockerWildTop), nil)
		assert.Contains(t, out, i18n.T("canasta.drawBlocker"+domain.CanastaDrawBlockerWildTop))
	})

	// **問題が無いときは行ごと出さない。** 個別のコードだけを NotContains で見ると、
	// 空のキー ("canasta.drawBlocker") が生で出ていても気づけない。接頭辞そのものが
	// 出力に現れないことを見る。
	t.Run("stays quiet when the pile can be taken", func(t *testing.T) {
		out := p.Output(withBlocker(""), nil)
		assert.NotContains(t, out, "canasta.drawBlocker")
		for _, code := range []string{
			domain.CanastaDrawBlockerPileEmpty, domain.CanastaDrawBlockerBlackThree,
			domain.CanastaDrawBlockerWildTop, domain.CanastaDrawBlockerNoPair,
		} {
			assert.NotContains(t, out, i18n.T("canasta.drawBlocker"+code))
		}
	})
}
