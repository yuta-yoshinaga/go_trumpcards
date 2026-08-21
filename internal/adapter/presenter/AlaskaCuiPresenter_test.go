//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupAlaskaCuiMockDefaults(rg *interfaces.MockAlaskaGame) {
	rg.On("GetPhase").Return(domain.AlaskaPhasePlaying).Maybe()
	rg.On("GetMoveCount").Return(0).Maybe()
	rg.On("IsStalemate").Return(false).Maybe()
	rg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
	for i := range domain.AlaskaTableauCnt {
		tableau[i] = []*domain.AlaskaTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, i+1, false), FaceUp: true},
		}
	}
	rg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.AlaskaFoundationCnt][]*domain.Card
	rg.On("GetFoundation").Return(foundation).Maybe()
}

func TestAlaskaCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		setupAlaskaCuiMockDefaults(rg)
		p := new(AlaskaCuiPresenter)

		result := p.Output(rg, nil)
		assert.Contains(t, result, "Alaska")
		assert.Contains(t, result, "Foundation")
		assert.Contains(t, result, "列0:")
	})

	t.Run("with error", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		setupAlaskaCuiMockDefaults(rg)
		p := new(AlaskaCuiPresenter)

		result := p.Output(rg, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("stalemate", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		rg.On("GetPhase").Return(domain.AlaskaPhasePlaying).Maybe()
		rg.On("GetMoveCount").Return(5).Maybe()
		rg.On("IsStalemate").Return(true).Maybe()
		rg.On("UndoToEscape").Return(0).Maybe()
		var tableau [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		rg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.AlaskaFoundationCnt][]*domain.Card
		rg.On("GetFoundation").Return(foundation).Maybe()

		p := new(AlaskaCuiPresenter)
		result := p.Output(rg, nil)
		assert.Contains(t, result, "手詰まり")
	})

	t.Run("game clear", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		rg.On("GetPhase").Return(domain.AlaskaPhaseGameClear).Maybe()
		rg.On("GetMoveCount").Return(42).Maybe()
		rg.On("IsStalemate").Return(false).Maybe()
		var tableau [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		rg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.AlaskaFoundationCnt][]*domain.Card
		rg.On("GetFoundation").Return(foundation).Maybe()

		p := new(AlaskaCuiPresenter)
		result := p.Output(rg, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		rg.On("GetPhase").Return(domain.AlaskaPhaseGameOver).Maybe()
		rg.On("GetMoveCount").Return(10).Maybe()
		rg.On("IsStalemate").Return(false).Maybe()
		var tableau [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		rg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.AlaskaFoundationCnt][]*domain.Card
		rg.On("GetFoundation").Return(foundation).Maybe()

		p := new(AlaskaCuiPresenter)
		result := p.Output(rg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})

	t.Run("foundation with cards", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		rg.On("GetPhase").Return(domain.AlaskaPhasePlaying).Maybe()
		rg.On("GetMoveCount").Return(0).Maybe()
		rg.On("IsStalemate").Return(false).Maybe()
		var tableau [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		rg.On("GetTableau").Return(tableau).Maybe()
		// Stack a single card on foundation[0] to exercise the non-empty branch.
		var foundation [domain.AlaskaFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		rg.On("GetFoundation").Return(foundation).Maybe()

		p := new(AlaskaCuiPresenter)
		result := p.Output(rg, nil)
		assert.Contains(t, result, "SPADE 1")
	})
}

func TestAlaskaCuiPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		rg.On("GetHint").Return(&domain.AlaskaHint{
			FromCol:   0,
			CardIndex: 1,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(AlaskaCuiPresenter)
		result := p.HintOutput(rg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "ファンデーション")
	})

	t.Run("hint to tableau", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		rg.On("GetHint").Return(&domain.AlaskaHint{
			FromCol:   0,
			CardIndex: 1,
			ToZone:    "tableau",
			ToCol:     3,
		})

		p := new(AlaskaCuiPresenter)
		result := p.HintOutput(rg)
		assert.Contains(t, result, "タブロー列3")
	})

	t.Run("no hint", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		rg.On("GetHint").Return((*domain.AlaskaHint)(nil))

		p := new(AlaskaCuiPresenter)
		result := p.HintOutput(rg)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestAlaskaCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		rg.On("GetPhase").Return(domain.AlaskaPhasePlaying)

		p := new(AlaskaCuiPresenter)
		result := p.ActionLogOutput(rg)
		assert.NotEmpty(t, result)
	})

	t.Run("game over", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		rg.On("GetPhase").Return(domain.AlaskaPhaseGameOver)
		rg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(AlaskaCuiPresenter)
		result := p.ActionLogOutput(rg)
		assert.NotEmpty(t, result)
	})
}
