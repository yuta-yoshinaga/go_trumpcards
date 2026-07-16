//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupFortyAndEightCuiMockDefaults(fg *interfaces.MockFortyAndEightGame) {
	fg.On("GetPhase").Return(domain.FortyAndEightPhasePlaying).Maybe()
	fg.On("GetMoveCount").Return(0).Maybe()
	fg.On("GetStockCount").Return(64).Maybe()
	fg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	fg.On("IsStalemate").Return(false).Maybe()
	fg.On("GetRedealUsed").Return(false).Maybe()

	var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
	for i := range domain.FortyAndEightTableauCnt {
		tableau[i] = make([]*domain.FortyAndEightTableauCard, 5)
		for j := range 5 {
			tableau[i][j] = &domain.FortyAndEightTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+1, false),
				FaceUp: true,
			}
		}
	}
	fg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.FortyAndEightFoundationCnt][]*domain.Card
	fg.On("GetFoundation").Return(foundation).Maybe()
}

func TestFortyAndEightCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightCuiMockDefaults(fg)
		p := new(FortyAndEightCuiPresenter)

		result := p.Output(fg, nil)
		assert.Contains(t, result, "Forty and Eight")
		assert.Contains(t, result, "Foundation")
		assert.Contains(t, result, "Stock: 64枚")
		assert.Contains(t, result, "Waste: [空]")
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "手数: 0")
		assert.Contains(t, result, "操作: m で移動")
		// Each foundation pile carries its move-command index label.
		assert.Contains(t, result, "[f0]")
		assert.Contains(t, result, "[f7]")
	})

	t.Run("with waste card", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetWaste")
		fg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})

		p := new(FortyAndEightCuiPresenter)
		result := p.Output(fg, nil)
		assert.Contains(t, result, "Waste:")
		assert.NotContains(t, result, "Waste: [空]")
	})

	t.Run("redeal used label", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetRedealUsed")
		fg.On("GetRedealUsed").Return(true)

		p := new(FortyAndEightCuiPresenter)
		result := p.Output(fg, nil)
		assert.Contains(t, result, "リディール")
	})

	t.Run("with error", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightCuiMockDefaults(fg)
		p := new(FortyAndEightCuiPresenter)

		result := p.Output(fg, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game clear", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetPhase")
		fg.On("GetPhase").Return(domain.FortyAndEightPhaseGameClear)

		p := new(FortyAndEightCuiPresenter)
		result := p.Output(fg, nil)
		assert.Contains(t, result, "ゲームクリア")
		assert.NotContains(t, result, "操作: m で移動")
	})

	t.Run("game over", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetPhase")
		fg.On("GetPhase").Return(domain.FortyAndEightPhaseGameOver)

		p := new(FortyAndEightCuiPresenter)
		result := p.Output(fg, nil)
		assert.Contains(t, result, "ゲームオーバー")
		assert.NotContains(t, result, "操作: m で移動")
	})

	t.Run("stalemate", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "IsStalemate")
		fg.On("IsStalemate").Return(true)

		p := new(FortyAndEightCuiPresenter)
		result := p.Output(fg, nil)
		assert.Contains(t, result, "手詰まり")
	})

	t.Run("empty column", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetTableau")
		var emptyTableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
		fg.On("GetTableau").Return(emptyTableau)

		p := new(FortyAndEightCuiPresenter)
		result := p.Output(fg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("foundation with cards", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetFoundation")
		var foundation [domain.FortyAndEightFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		fg.On("GetFoundation").Return(foundation)

		p := new(FortyAndEightCuiPresenter)
		result := p.Output(fg, nil)
		assert.Contains(t, result, "SPADE 1")
	})
}

func TestFortyAndEightCuiPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		fg.On("GetHint").Return(&domain.FortyAndEightHint{
			FromZone:  "tableau",
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(FortyAndEightCuiPresenter)
		result := p.HintOutput(fg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "タブロー列0")
		assert.Contains(t, result, "ファンデーション")
	})

	t.Run("waste hint", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		fg.On("GetHint").Return(&domain.FortyAndEightHint{
			FromZone:  "waste",
			FromCol:   -1,
			CardIndex: -1,
			ToZone:    "tableau",
			ToCol:     3,
		})

		p := new(FortyAndEightCuiPresenter)
		result := p.HintOutput(fg)
		assert.Contains(t, result, "ウェイスト")
		assert.Contains(t, result, "タブロー列3")
	})

	t.Run("no hint", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		fg.On("GetHint").Return((*domain.FortyAndEightHint)(nil))

		p := new(FortyAndEightCuiPresenter)
		result := p.HintOutput(fg)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestFortyAndEightCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		fg.On("GetPhase").Return(domain.FortyAndEightPhasePlaying)

		p := new(FortyAndEightCuiPresenter)
		result := p.ActionLogOutput(fg)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		fg.On("GetPhase").Return(domain.FortyAndEightPhaseGameOver)
		fg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "draw", Detail: "test"},
		})

		p := new(FortyAndEightCuiPresenter)
		result := p.ActionLogOutput(fg)
		assert.Contains(t, result, "draw")
	})
}
