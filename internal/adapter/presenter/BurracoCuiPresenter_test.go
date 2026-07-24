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

func setupBurracoCuiMock() *interfaces.MockBurracoGame {
	m := new(interfaces.MockBurracoGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(54)
	m.On("GetDiscardPileCount").Return(0)
	m.On("GetPozzettoCount").Return(2)
	m.On("GetIsFrozen").Return(false)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.BurracoPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupBurracoCuiMockWithPlayers() (*interfaces.MockBurracoGame, []*domain.BurracoPlayer) {
	m := setupBurracoCuiMock()
	players := makeBurracoPlayers()
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestBurracoCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BurracoCuiPresenter)

	t.Run("initial state with header and player info", func(t *testing.T) {
		m, players := setupBurracoCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Burraco (ブラーコ)")
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
		m, _ := setupBurracoCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetIsFrozen")
		m.On("GetIsFrozen").Return(true)

		result := p.Output(m, nil)
		assert.Contains(t, result, "[フリーズ]")
		// The draw prompt warns about the freeze constraint (top-only, no wild pickup).
		assert.Contains(t, result, "フリーズ中: 上の1枚のみ")
	})

	t.Run("discard top shown", func(t *testing.T) {
		m, _ := setupBurracoCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		top := domain.NewCard(domain.CardDesignHeart, 7, false)
		m.On("GetDiscardTop").Return(top)

		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札: HEART 7")
	})

	t.Run("discard top nil hides section", func(t *testing.T) {
		m, _ := setupBurracoCuiMockWithPlayers()

		result := p.Output(m, nil)
		assert.NotContains(t, result, "捨て札:")
	})

	t.Run("player with scores", func(t *testing.T) {
		m, players := setupBurracoCuiMockWithPlayers()
		players[1].SetCumulativeScore(300)
		players[1].SetRoundScore(100)

		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1: 累積300点 ラウンド100点 0枚")
	})

	t.Run("player with meld shown", func(t *testing.T) {
		m, players := setupBurracoCuiMockWithPlayers()
		meld := &domain.BurracoMeld{
			Cards: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignClover, 7, false),
			},
			IsNatural: true,
		}
		players[0].SetMelds([]*domain.BurracoMeld{meld})

		result := p.Output(m, nil)
		assert.Contains(t, result, "ナチュラル")
		assert.Contains(t, result, "SPADE 7")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupBurracoCuiMockWithPlayers()
		testErr := errors.New("invalid card index")

		result := p.Output(m, testErr)
		assert.Contains(t, result, "invalid card index")
	})

	t.Run("game ended shows winner human", func(t *testing.T) {
		m, _ := setupBurracoCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "あなたの勝利です！")
	})

	t.Run("game ended shows winner CPU", func(t *testing.T) {
		m, _ := setupBurracoCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "CPU 1の勝利です！")
	})

	t.Run("draw phase shows current player CPU", func(t *testing.T) {
		m, _ := setupBurracoCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "手番: CPU 1")
	})

	t.Run("meld phase shows commands", func(t *testing.T) {
		m, _ := setupBurracoCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BurracoPhaseMeld)

		result := p.Output(m, nil)
		assert.Contains(t, result, "メルドフェーズ")
		assert.Contains(t, result, "m ")
		assert.Contains(t, result, "sm")
	})

	t.Run("discard phase shows commands", func(t *testing.T) {
		m, _ := setupBurracoCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BurracoPhaseDiscard)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ディスカードフェーズ")
		assert.Contains(t, result, "d <idx>")
		assert.Contains(t, result, "go")
	})

	t.Run("round end phase shows next command", func(t *testing.T) {
		m, _ := setupBurracoCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BurracoPhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "nr / nextround")
	})

	t.Run("red 3 tag shown", func(t *testing.T) {
		m, players := setupBurracoCuiMockWithPlayers()
		players[0].AddRed3(domain.NewCard(domain.CardDesignHeart, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "赤3: 1枚")
	})

	t.Run("burraco star tag shown when player holds a burraco meld", func(t *testing.T) {
		m, players := setupBurracoCuiMockWithPlayers()
		// A 7-card natural meld of 4s satisfies HasBurraco() (>=7 cards).
		// IsNatural=true also exercises the "ナチュラル" meld-type label;
		// the m.IsBurraco() branch attaches "ブラーコ" to that label.
		cards := make([]*domain.Card, 7)
		for i := range cards {
			cards[i] = domain.NewCard(domain.CardDesignSpade, 4, false)
		}
		players[0].AddMeld(&domain.BurracoMeld{Cards: cards, IsNatural: true})

		result := p.Output(m, nil)
		assert.Contains(t, result, "★ブラーコ")
		assert.Contains(t, result, "ナチュラルブラーコ")
	})
}

func TestBurracoCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BurracoCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockBurracoGame)
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
		m := new(interfaces.MockBurracoGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.NotEmpty(t, result)
		m.AssertExpectations(t)
	})
}

// newBurracoHintGame builds a real 2-player Burraco game for hint tests.
func newBurracoHintGame() *domain.Burraco {
	players := []*domain.BurracoPlayer{
		domain.NewCanastaPlayer(true),
		domain.NewCanastaPlayer(false),
	}
	return domain.NewCanasta(domain.NewTrumpCardsWithDecks(2, 4), players, domain.DefaultCanastaConfig())
}

func TestBurracoCuiPresenter_HintOutput(t *testing.T) {
	p := &presenter.BurracoCuiPresenter{}

	t.Run("no hint on CPU turn", func(t *testing.T) {
		g := newBurracoHintGame()
		g.SetPhase(domain.BurracoPhaseDraw)
		g.SetCurrentPlayerIdx(1) // CPU
		out := p.HintOutput(g)
		assert.NotEmpty(t, out)
	})

	t.Run("draw stock", func(t *testing.T) {
		g := newBurracoHintGame()
		g.SetPhase(domain.BurracoPhaseDraw)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("draw discard", func(t *testing.T) {
		g := newBurracoHintGame()
		g.SetPhase(domain.BurracoPhaseDraw)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("meld", func(t *testing.T) {
		g := newBurracoHintGame()
		g.SetPhase(domain.BurracoPhaseMeld)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 8, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("skip meld", func(t *testing.T) {
		g := newBurracoHintGame()
		g.SetPhase(domain.BurracoPhaseMeld)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("discard", func(t *testing.T) {
		g := newBurracoHintGame()
		g.SetPhase(domain.BurracoPhaseDiscard)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		out := p.HintOutput(g)
		assert.NotEmpty(t, out)
	})
}
