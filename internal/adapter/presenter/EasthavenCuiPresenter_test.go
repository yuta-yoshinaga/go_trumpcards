//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupEasthavenCuiMockDefaults(eg *interfaces.MockEasthavenGame) {
	eg.On("GetPhase").Return(domain.EasthavenPhasePlaying).Maybe()
	eg.On("GetMoveCount").Return(0).Maybe()
	eg.On("GetStockCount").Return(31).Maybe()
	eg.On("IsStalemate").Return(false).Maybe()
	eg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
	for i := range domain.EasthavenTableauCnt {
		tableau[i] = []*domain.KlondikeTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, i+1, false), FaceUp: true},
		}
	}
	eg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.EasthavenFoundationCnt][]*domain.Card
	eg.On("GetFoundation").Return(foundation).Maybe()
}

func TestEasthavenCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		setupEasthavenCuiMockDefaults(eg)
		p := new(EasthavenCuiPresenter)

		result := p.Output(eg, nil)
		assert.Contains(t, result, "Easthaven")
		assert.Contains(t, result, "Foundation")
		assert.Contains(t, result, "列0:")
	})

	t.Run("with error", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		setupEasthavenCuiMockDefaults(eg)
		p := new(EasthavenCuiPresenter)
		assert.Contains(t, p.Output(eg, assert.AnError), assert.AnError.Error())
	})

	t.Run("stalemate", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetPhase").Return(domain.EasthavenPhasePlaying).Maybe()
		eg.On("GetMoveCount").Return(5).Maybe()
		eg.On("GetStockCount").Return(0).Maybe()
		eg.On("IsStalemate").Return(true).Maybe()
		eg.On("UndoToEscape").Return(0).Maybe()
		var tableau [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		eg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.EasthavenFoundationCnt][]*domain.Card
		eg.On("GetFoundation").Return(foundation).Maybe()

		p := new(EasthavenCuiPresenter)
		assert.Contains(t, p.Output(eg, nil), "手詰まり")
	})

	t.Run("game clear", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetPhase").Return(domain.EasthavenPhaseGameClear).Maybe()
		eg.On("GetMoveCount").Return(42).Maybe()
		eg.On("GetStockCount").Return(0).Maybe()
		eg.On("IsStalemate").Return(false).Maybe()
		var tableau [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		eg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.EasthavenFoundationCnt][]*domain.Card
		eg.On("GetFoundation").Return(foundation).Maybe()

		p := new(EasthavenCuiPresenter)
		assert.Contains(t, p.Output(eg, nil), "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetPhase").Return(domain.EasthavenPhaseGameOver).Maybe()
		eg.On("GetMoveCount").Return(10).Maybe()
		eg.On("GetStockCount").Return(0).Maybe()
		eg.On("IsStalemate").Return(false).Maybe()
		var tableau [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		eg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.EasthavenFoundationCnt][]*domain.Card
		eg.On("GetFoundation").Return(foundation).Maybe()

		p := new(EasthavenCuiPresenter)
		assert.Contains(t, p.Output(eg, nil), "ゲームオーバー")
	})

	t.Run("foundation with cards", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetPhase").Return(domain.EasthavenPhasePlaying).Maybe()
		eg.On("GetMoveCount").Return(0).Maybe()
		eg.On("GetStockCount").Return(0).Maybe()
		eg.On("IsStalemate").Return(false).Maybe()
		var tableau [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		eg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.EasthavenFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		eg.On("GetFoundation").Return(foundation).Maybe()

		p := new(EasthavenCuiPresenter)
		assert.Contains(t, p.Output(eg, nil), "SPADE 1")
	})
}

func TestEasthavenCuiPresenter_HintOutput(t *testing.T) {
	t.Run("to foundation", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetHint").Return(&domain.EasthavenHint{FromCol: 0, CardIndex: 1, ToZone: "foundation", ToCol: 0})
		p := new(EasthavenCuiPresenter)
		result := p.HintOutput(eg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "ファンデーション")
	})

	t.Run("to tableau", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetHint").Return(&domain.EasthavenHint{FromCol: 0, CardIndex: 1, ToZone: "tableau", ToCol: 3})
		p := new(EasthavenCuiPresenter)
		assert.Contains(t, p.HintOutput(eg), "タブロー列3")
	})

	t.Run("no hint", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetHint").Return((*domain.EasthavenHint)(nil))
		p := new(EasthavenCuiPresenter)
		assert.Contains(t, p.HintOutput(eg), "ヒントはありません")
	})
}

func TestEasthavenCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetPhase").Return(domain.EasthavenPhasePlaying)
		p := new(EasthavenCuiPresenter)
		assert.NotEmpty(t, p.ActionLogOutput(eg))
	})

	t.Run("game over", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetPhase").Return(domain.EasthavenPhaseGameOver)
		eg.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "move", Detail: "test"}})
		p := new(EasthavenCuiPresenter)
		assert.NotEmpty(t, p.ActionLogOutput(eg))
	})
}
