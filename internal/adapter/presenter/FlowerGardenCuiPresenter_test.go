//go:build test

package presenter

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupFlowerGardenCuiMockDefaults(bg *interfaces.MockFlowerGardenGame) {
	bg.On("GetPhase").Return(domain.FlowerGardenPhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
	bg.On("IsStalemate").Return(false).Maybe()
	bg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
	for i := range domain.FlowerGardenTableauCnt {
		tableau[i] = make([]*domain.FlowerGardenTableauCard, domain.FlowerGardenColumnLen)
		for j := range domain.FlowerGardenColumnLen {
			tableau[i][j] = &domain.FlowerGardenTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	bg.On("GetTableau").Return(tableau).Maybe()

	reserve := make([]*domain.Card, domain.FlowerGardenReserveCnt)
	for i := range reserve {
		reserve[i] = domain.NewCard(domain.CardDesignHeart, (i%13)+1, false)
	}
	bg.On("GetReserve").Return(reserve).Maybe()

	var foundation [domain.FlowerGardenFoundationCnt][]*domain.Card
	for i := range domain.FlowerGardenFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade+i, 1, false)}
	}
	bg.On("GetFoundation").Return(foundation).Maybe()
}

func TestFlowerGardenCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenCuiMockDefaults(bg)
		p := new(FlowerGardenCuiPresenter)

		result := p.Output(bg, nil)
		assert.Contains(t, result, "Flower Garden")
		assert.Contains(t, result, "Foundation")
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "手数: 0")
		// The 16 reserve cards wrap across rows: r0 and r8 land on different lines.
		assert.Contains(t, result, "[r0]")
		assert.Contains(t, result, "[r15]")
		var r0Line, r8Line string
		for _, l := range strings.Split(result, "\n") {
			if strings.Contains(l, "[r0]") {
				r0Line = l
			}
			if strings.Contains(l, "[r8]") {
				r8Line = l
			}
		}
		assert.NotEmpty(t, r0Line)
		assert.NotEmpty(t, r8Line)
		assert.NotContains(t, r0Line, "[r8]") // wrapped onto a separate row
	})

	t.Run("with error", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenCuiMockDefaults(bg)
		p := new(FlowerGardenCuiPresenter)

		result := p.Output(bg, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.FlowerGardenPhaseGameClear)

		p := new(FlowerGardenCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.FlowerGardenPhaseGameOver)

		p := new(FlowerGardenCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})

	t.Run("stalemate", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)
		bg.On("UndoToEscape").Return(0).Maybe()

		p := new(FlowerGardenCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "手詰まり")
	})

	t.Run("empty column and depleted reserve", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetTableau")
		var emptyTableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
		bg.On("GetTableau").Return(emptyTableau)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetReserve")
		bg.On("GetReserve").Return([]*domain.Card{nil, nil})

		p := new(FlowerGardenCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("empty foundation", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		setupFlowerGardenCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetFoundation")
		var emptyFoundation [domain.FlowerGardenFoundationCnt][]*domain.Card
		bg.On("GetFoundation").Return(emptyFoundation)

		p := new(FlowerGardenCuiPresenter)
		result := p.Output(bg, nil)
		assert.NotEmpty(t, result)
	})
}

func TestFlowerGardenCuiPresenter_HintOutput(t *testing.T) {
	t.Run("foundation hint from tableau", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		bg.On("GetHint").Return(&domain.FlowerGardenHint{
			FromZone:  "tableau",
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(FlowerGardenCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "フラワーベッド0")
		assert.Contains(t, result, "ファンデーション")
	})

	t.Run("tableau hint", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		bg.On("GetHint").Return(&domain.FlowerGardenHint{
			FromZone:  "tableau",
			FromCol:   1,
			CardIndex: 0,
			ToZone:    "tableau",
			ToCol:     3,
		})

		p := new(FlowerGardenCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "フラワーベッド1")
		assert.Contains(t, result, "フラワーベッド3")
	})

	t.Run("reserve hint", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		bg.On("GetHint").Return(&domain.FlowerGardenHint{
			FromZone:  "reserve",
			FromCol:   2,
			CardIndex: -1,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(FlowerGardenCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "ブーケ2")
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		bg.On("GetHint").Return((*domain.FlowerGardenHint)(nil))

		p := new(FlowerGardenCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestFlowerGardenCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		bg.On("GetPhase").Return(domain.FlowerGardenPhasePlaying)

		p := new(FlowerGardenCuiPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		bg := new(interfaces.MockFlowerGardenGame)
		bg.On("GetPhase").Return(domain.FlowerGardenPhaseGameOver)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(FlowerGardenCuiPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "move")
	})
}
