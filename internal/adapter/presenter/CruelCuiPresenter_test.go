//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupCruelCuiMockDefaults(cg *interfaces.MockCruelGame) {
	cg.On("GetPhase").Return(domain.CruelPhasePlaying).Maybe()
	cg.On("GetMoveCount").Return(0).Maybe()
	cg.On("IsStalemate").Return(false).Maybe()

	var tableau [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	for i := range domain.CruelTableauCnt {
		tableau[i] = []*domain.KlondikeTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, (i%13)+2, false), FaceUp: true},
		}
	}
	cg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.CruelFoundationCnt][]*domain.Card
	cg.On("GetFoundation").Return(foundation).Maybe()
}

func TestCruelCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		setupCruelCuiMockDefaults(cg)
		p := new(CruelCuiPresenter)

		result := p.Output(cg, nil)
		assert.Contains(t, result, "Cruel")
		assert.Contains(t, result, "Foundation")
		assert.Contains(t, result, "列0:")
		// Empty foundation -> 0/52 progress; the shift/move help is always shown.
		assert.Contains(t, result, "進捗: 0/52")
		assert.Contains(t, result, "操作:")
	})

	t.Run("with error", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		setupCruelCuiMockDefaults(cg)
		p := new(CruelCuiPresenter)

		result := p.Output(cg, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("stalemate", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		cg.On("GetPhase").Return(domain.CruelPhasePlaying).Maybe()
		cg.On("GetMoveCount").Return(5).Maybe()
		cg.On("IsStalemate").Return(true).Maybe()
		var tableau [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
		cg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.CruelFoundationCnt][]*domain.Card
		cg.On("GetFoundation").Return(foundation).Maybe()

		p := new(CruelCuiPresenter)
		result := p.Output(cg, nil)
		assert.Contains(t, result, "手詰まり")
		assert.Contains(t, result, "シフト")
		// Stalemate adds the give-up guidance beyond the shift hint.
		assert.Contains(t, result, "ギブアップ")
	})

	t.Run("game clear", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		cg.On("GetPhase").Return(domain.CruelPhaseGameClear).Maybe()
		cg.On("GetMoveCount").Return(42).Maybe()
		cg.On("IsStalemate").Return(false).Maybe()
		var tableau [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
		cg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.CruelFoundationCnt][]*domain.Card
		cg.On("GetFoundation").Return(foundation).Maybe()

		p := new(CruelCuiPresenter)
		result := p.Output(cg, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		cg.On("GetPhase").Return(domain.CruelPhaseGameOver).Maybe()
		cg.On("GetMoveCount").Return(10).Maybe()
		cg.On("IsStalemate").Return(false).Maybe()
		var tableau [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
		cg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.CruelFoundationCnt][]*domain.Card
		cg.On("GetFoundation").Return(foundation).Maybe()

		p := new(CruelCuiPresenter)
		result := p.Output(cg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})

	t.Run("foundation with cards", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		cg.On("GetPhase").Return(domain.CruelPhasePlaying).Maybe()
		cg.On("GetMoveCount").Return(0).Maybe()
		cg.On("IsStalemate").Return(false).Maybe()
		var tableau [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
		cg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.CruelFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		cg.On("GetFoundation").Return(foundation).Maybe()

		p := new(CruelCuiPresenter)
		result := p.Output(cg, nil)
		assert.Contains(t, result, "SPADE 1")
	})
}

func TestCruelCuiPresenter_HintOutput(t *testing.T) {
	t.Run("hint to foundation", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		cg.On("GetHint").Return(&domain.CruelHint{
			FromCol:   0,
			CardIndex: 1,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(CruelCuiPresenter)
		result := p.HintOutput(cg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "ファウンデーション")
	})

	t.Run("hint to tableau", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		cg.On("GetHint").Return(&domain.CruelHint{
			FromCol:   0,
			CardIndex: 1,
			ToZone:    "tableau",
			ToCol:     3,
		})

		p := new(CruelCuiPresenter)
		result := p.HintOutput(cg)
		assert.Contains(t, result, "タブロー列3")
	})

	t.Run("no hint", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		cg.On("GetHint").Return((*domain.CruelHint)(nil))

		p := new(CruelCuiPresenter)
		result := p.HintOutput(cg)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestCruelCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		cg.On("GetPhase").Return(domain.CruelPhasePlaying)

		p := new(CruelCuiPresenter)
		result := p.ActionLogOutput(cg)
		assert.NotEmpty(t, result)
	})

	t.Run("game over", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		cg.On("GetPhase").Return(domain.CruelPhaseGameOver)
		cg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(CruelCuiPresenter)
		result := p.ActionLogOutput(cg)
		assert.NotEmpty(t, result)
	})
}
