//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupWaspCuiMockDefaults(sg *interfaces.MockWaspGame) {
	sg.On("GetPhase").Return(domain.WaspPhasePlaying).Maybe()
	sg.On("GetMoveCount").Return(0).Maybe()
	sg.On("IsStalemate").Return(false).Maybe()
	sg.On("UndoToEscape").Return(0).Maybe()
	sg.On("GetCompletedSuits").Return(0).Maybe()
	sg.On("GetStockCount").Return(3).Maybe()

	var tableau [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
	for i := range domain.WaspTableauCnt {
		tableau[i] = []*domain.KlondikeTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, i+1, false), FaceUp: true},
		}
	}
	sg.On("GetTableau").Return(tableau).Maybe()
}

func TestWaspCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		setupWaspCuiMockDefaults(sg)
		p := new(WaspCuiPresenter)

		result := p.Output(sg, nil)
		assert.Contains(t, result, "Wasp")
		assert.Contains(t, result, "完成スーツ")
		assert.Contains(t, result, "列0:")
	})

	t.Run("with error", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		setupWaspCuiMockDefaults(sg)
		p := new(WaspCuiPresenter)

		result := p.Output(sg, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("stalemate", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetPhase").Return(domain.WaspPhasePlaying).Maybe()
		sg.On("GetMoveCount").Return(5).Maybe()
		sg.On("IsStalemate").Return(true).Maybe()
		sg.On("UndoToEscape").Return(0).Maybe()
		sg.On("GetCompletedSuits").Return(0).Maybe()
		sg.On("GetStockCount").Return(0).Maybe()
		var tableau [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(WaspCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "手詰まり")
	})

	t.Run("game clear", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetPhase").Return(domain.WaspPhaseGameClear).Maybe()
		sg.On("GetMoveCount").Return(42).Maybe()
		sg.On("IsStalemate").Return(false).Maybe()
		sg.On("GetCompletedSuits").Return(4).Maybe()
		sg.On("GetStockCount").Return(0).Maybe()
		var tableau [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(WaspCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetPhase").Return(domain.WaspPhaseGameOver).Maybe()
		sg.On("GetMoveCount").Return(10).Maybe()
		sg.On("IsStalemate").Return(false).Maybe()
		sg.On("GetCompletedSuits").Return(0).Maybe()
		sg.On("GetStockCount").Return(0).Maybe()
		var tableau [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(WaspCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})
}

func TestWaspCuiPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetHint").Return(&domain.WaspHint{FromCol: 0, CardIndex: 1, ToCol: 3})

		p := new(WaspCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "タブロー列3")
	})

	// #6340: 裏カードを開ける手を優先しているのに、その理由を言っていなかった。
	t.Run("says when the move uncovers a face-down card", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetHint").Return(&domain.WaspHint{FromCol: 0, CardIndex: 1, ToCol: 3, ExposesFaceDown: true})

		result := new(WaspCuiPresenter).HintOutput(sg)
		assert.Contains(t, result, i18n.T("wasp.hintExposes"))
	})

	// **開かない手では言わない。**常に出ると理由として機能しない。
	t.Run("stays silent when the move uncovers nothing", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetHint").Return(&domain.WaspHint{FromCol: 0, CardIndex: 1, ToCol: 3})

		result := new(WaspCuiPresenter).HintOutput(sg)
		assert.NotContains(t, result, i18n.T("wasp.hintExposes"))
	})

	t.Run("deal hint", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetHint").Return(&domain.WaspHint{
			FromCol:   domain.WaspHintDeal,
			CardIndex: domain.WaspHintDeal,
			ToCol:     domain.WaspHintDeal,
		})

		p := new(WaspCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "ストック")
	})

	t.Run("no hint", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetHint").Return((*domain.WaspHint)(nil))

		p := new(WaspCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestWaspCuiPresenter_LegalMovesOutput(t *testing.T) {
	// waspTableauWith builds a full-length tableau, overriding specific columns.
	waspTableauWith := func(overrides map[int][]*domain.KlondikeTableauCard) [domain.WaspTableauCnt][]*domain.KlondikeTableauCard {
		var tableau [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		for i := range domain.WaspTableauCnt {
			if cards, ok := overrides[i]; ok {
				tableau[i] = cards
				continue
			}
			tableau[i] = nil
		}
		return tableau
	}

	t.Run("lists matching column and flags empty columns", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetPhase").Return(domain.WaspPhasePlaying).Maybe()
		// Column 0 top card: Spade 5 (movable). Column 1 top card: Spade 6 (legal target).
		// Column 2 top card: Heart 6 (wrong suit). Column 3 empty (always legal).
		tableau := waspTableauWith(map[int][]*domain.KlondikeTableauCard{
			0: {{Card: domain.NewCard(domain.CardDesignSpade, 5, false), FaceUp: true}},
			1: {{Card: domain.NewCard(domain.CardDesignSpade, 6, false), FaceUp: true}},
			2: {{Card: domain.NewCard(domain.CardDesignHeart, 6, false), FaceUp: true}},
		})
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(WaspCuiPresenter)
		result := p.LegalMovesOutput(sg, 0)
		assert.Contains(t, result, "列0")
		assert.Contains(t, result, "列1")      // suit-matching target
		assert.Contains(t, result, "空列")      // empty columns flagged
		assert.NotContains(t, result, "列2 (") // wrong-suit column not listed as a target
	})

	t.Run("no movable card in empty column", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetPhase").Return(domain.WaspPhasePlaying).Maybe()
		sg.On("GetTableau").Return(waspTableauWith(nil)).Maybe()

		p := new(WaspCuiPresenter)
		result := p.LegalMovesOutput(sg, 0)
		assert.Contains(t, result, "移動可能")
	})

	t.Run("face-down top card", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetPhase").Return(domain.WaspPhasePlaying).Maybe()
		tableau := waspTableauWith(map[int][]*domain.KlondikeTableauCard{
			0: {{Card: domain.NewCard(domain.CardDesignSpade, 5, false), FaceUp: false}},
		})
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(WaspCuiPresenter)
		result := p.LegalMovesOutput(sg, 0)
		assert.Contains(t, result, "移動可能")
	})

	t.Run("no legal destinations", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetPhase").Return(domain.WaspPhasePlaying).Maybe()
		// Every other column occupied with a non-matching card, none empty.
		overrides := map[int][]*domain.KlondikeTableauCard{}
		for i := range domain.WaspTableauCnt {
			overrides[i] = []*domain.KlondikeTableauCard{
				{Card: domain.NewCard(domain.CardDesignHeart, 2, false), FaceUp: true},
			}
		}
		sg.On("GetTableau").Return(waspTableauWith(overrides)).Maybe()

		p := new(WaspCuiPresenter)
		result := p.LegalMovesOutput(sg, 0)
		assert.Contains(t, result, "なし")
	})

	t.Run("invalid column", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetPhase").Return(domain.WaspPhasePlaying).Maybe()

		p := new(WaspCuiPresenter)
		assert.NotEmpty(t, p.LegalMovesOutput(sg, -1))
		assert.NotEmpty(t, p.LegalMovesOutput(sg, domain.WaspTableauCnt))
	})

	t.Run("not in playing phase", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetPhase").Return(domain.WaspPhaseGameOver).Maybe()

		p := new(WaspCuiPresenter)
		result := p.LegalMovesOutput(sg, 0)
		assert.NotEmpty(t, result)
	})
}

func TestWaspCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetPhase").Return(domain.WaspPhasePlaying)

		p := new(WaspCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.NotEmpty(t, result)
	})

	t.Run("game over", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetPhase").Return(domain.WaspPhaseGameOver)
		sg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(WaspCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.NotEmpty(t, result)
	})
}
