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

func setupBakersDozenCuiMockDefaults(bg *interfaces.MockBakersDozenGame) {
	bg.On("GetPhase").Return(domain.BakersDozenPhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
	bg.On("IsStalemate").Return(false).Maybe()

	var tableau [domain.BakersDozenTableauCnt][]*domain.BakersDozenTableauCard
	for i := range domain.BakersDozenTableauCnt {
		tableau[i] = make([]*domain.BakersDozenTableauCard, 4)
		for j := range 4 {
			tableau[i][j] = &domain.BakersDozenTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+1, false),
				FaceUp: true,
			}
		}
	}
	bg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.BakersDozenFoundationCnt][]*domain.Card
	bg.On("GetFoundation").Return(foundation).Maybe()
}

func TestBakersDozenCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		setupBakersDozenCuiMockDefaults(bg)
		p := new(BakersDozenCuiPresenter)

		result := p.Output(bg, nil)
		assert.Contains(t, result, "Baker's Dozen")
		assert.Contains(t, result, "Foundation")
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "手数: 0")
		// Playing phase surfaces the empty-column caveat; no column is at 1 card.
		assert.Contains(t, result, "空列は再利用できません")
		assert.NotContains(t, result, "残り1枚")
	})

	t.Run("single-card column is flagged with a warning marker", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		setupBakersDozenCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetTableau")
		var tableau [domain.BakersDozenTableauCnt][]*domain.BakersDozenTableauCard
		// Column 0 has a single card (at-risk); the rest keep four.
		tableau[0] = []*domain.BakersDozenTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, 5, false), FaceUp: true},
		}
		for i := 1; i < domain.BakersDozenTableauCnt; i++ {
			tableau[i] = make([]*domain.BakersDozenTableauCard, 4)
			for j := range 4 {
				tableau[i][j] = &domain.BakersDozenTableauCard{
					Card:   domain.NewCard(domain.CardDesignHeart, j+1, false),
					FaceUp: true,
				}
			}
		}
		bg.On("GetTableau").Return(tableau)
		p := new(BakersDozenCuiPresenter)

		result := p.Output(bg, nil)
		assert.Contains(t, result, "残り1枚")
	})

	t.Run("with error", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		setupBakersDozenCuiMockDefaults(bg)
		p := new(BakersDozenCuiPresenter)

		result := p.Output(bg, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		setupBakersDozenCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.BakersDozenPhaseGameClear)

		p := new(BakersDozenCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		setupBakersDozenCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.BakersDozenPhaseGameOver)

		p := new(BakersDozenCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})

	t.Run("stalemate", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		setupBakersDozenCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)

		p := new(BakersDozenCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "手詰まり")
	})

	t.Run("empty column", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		setupBakersDozenCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetTableau")
		var emptyTableau [domain.BakersDozenTableauCnt][]*domain.BakersDozenTableauCard
		bg.On("GetTableau").Return(emptyTableau)

		p := new(BakersDozenCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("foundation with cards", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		setupBakersDozenCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetFoundation")
		var foundation [domain.BakersDozenFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		bg.On("GetFoundation").Return(foundation)

		p := new(BakersDozenCuiPresenter)
		result := p.Output(bg, nil)
		assert.NotEmpty(t, result)
	})
}

func TestBakersDozenCuiPresenter_HintOutput(t *testing.T) {
	t.Run("foundation hint", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		bg.On("GetHint").Return(&domain.BakersDozenHint{
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(BakersDozenCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "タブロー列0")
		assert.Contains(t, result, "ファンデーション")
	})

	t.Run("tableau hint", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		bg.On("GetHint").Return(&domain.BakersDozenHint{
			FromCol:   1,
			CardIndex: 0,
			ToZone:    "tableau",
			ToCol:     3,
		})

		p := new(BakersDozenCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "タブロー列1")
		assert.Contains(t, result, "タブロー列3")
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		bg.On("GetHint").Return((*domain.BakersDozenHint)(nil))

		p := new(BakersDozenCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestBakersDozenCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		bg.On("GetGameEndFlag").Return(false)

		p := new(BakersDozenCuiPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		bg := new(interfaces.MockBakersDozenGame)
		bg.On("GetGameEndFlag").Return(true)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(BakersDozenCuiPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "move")
	})
}
