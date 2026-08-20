//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupScorpionCuiMockDefaults(sg *interfaces.MockScorpionGame) {
	sg.On("GetPhase").Return(domain.ScorpionPhasePlaying).Maybe()
	sg.On("GetMoveCount").Return(0).Maybe()
	sg.On("IsStalemate").Return(false).Maybe()
	sg.On("UndoToEscape").Return(0).Maybe()
	sg.On("GetCompletedSuits").Return(0).Maybe()
	sg.On("GetStockCount").Return(3).Maybe()

	var tableau [domain.ScorpionTableauCnt][]*domain.KlondikeTableauCard
	for i := range domain.ScorpionTableauCnt {
		tableau[i] = []*domain.KlondikeTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, i+1, false), FaceUp: true},
		}
	}
	sg.On("GetTableau").Return(tableau).Maybe()
}

func TestScorpionCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		setupScorpionCuiMockDefaults(sg)
		p := new(ScorpionCuiPresenter)

		result := p.Output(sg, nil)
		assert.Contains(t, result, "Scorpion")
		assert.Contains(t, result, "完成スーツ")
		assert.Contains(t, result, "列0:")
	})

	t.Run("with error", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		setupScorpionCuiMockDefaults(sg)
		p := new(ScorpionCuiPresenter)

		result := p.Output(sg, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("stalemate", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhasePlaying).Maybe()
		sg.On("GetMoveCount").Return(5).Maybe()
		sg.On("IsStalemate").Return(true).Maybe()
		sg.On("UndoToEscape").Return(0).Maybe()
		sg.On("GetCompletedSuits").Return(0).Maybe()
		sg.On("GetStockCount").Return(0).Maybe()
		var tableau [domain.ScorpionTableauCnt][]*domain.KlondikeTableauCard
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(ScorpionCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "手詰まり")
	})

	t.Run("game clear", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhaseGameClear).Maybe()
		sg.On("GetMoveCount").Return(42).Maybe()
		sg.On("IsStalemate").Return(false).Maybe()
		sg.On("GetCompletedSuits").Return(4).Maybe()
		sg.On("GetStockCount").Return(0).Maybe()
		var tableau [domain.ScorpionTableauCnt][]*domain.KlondikeTableauCard
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(ScorpionCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhaseGameOver).Maybe()
		sg.On("GetMoveCount").Return(10).Maybe()
		sg.On("IsStalemate").Return(false).Maybe()
		sg.On("GetCompletedSuits").Return(0).Maybe()
		sg.On("GetStockCount").Return(0).Maybe()
		var tableau [domain.ScorpionTableauCnt][]*domain.KlondikeTableauCard
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(ScorpionCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})
}

func TestScorpionCuiPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetHint").Return(&domain.ScorpionHint{FromCol: 0, CardIndex: 1, ToCol: 3})

		p := new(ScorpionCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "タブロー列3")
	})

	// #5544: 裏カードを開ける手を優先しているのに、その理由を言っていなかった。
	t.Run("says when the move uncovers a face-down card", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetHint").Return(&domain.ScorpionHint{FromCol: 0, CardIndex: 1, ToCol: 3, ExposesFaceDown: true})

		result := new(ScorpionCuiPresenter).HintOutput(sg)
		assert.Contains(t, result, i18n.T("scorpion.hintExposes"))
	})

	// **開かない手では言わない。**常に出ると理由として機能しない。
	t.Run("stays silent when the move uncovers nothing", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetHint").Return(&domain.ScorpionHint{FromCol: 0, CardIndex: 1, ToCol: 3})

		result := new(ScorpionCuiPresenter).HintOutput(sg)
		assert.NotContains(t, result, i18n.T("scorpion.hintExposes"))
	})

	t.Run("deal hint", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetHint").Return(&domain.ScorpionHint{
			FromCol:   domain.ScorpionHintDeal,
			CardIndex: domain.ScorpionHintDeal,
			ToCol:     domain.ScorpionHintDeal,
		})

		p := new(ScorpionCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "ストック")
	})

	t.Run("no hint", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetHint").Return((*domain.ScorpionHint)(nil))

		p := new(ScorpionCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestScorpionCuiPresenter_LegalMovesOutput(t *testing.T) {
	// scorpionTableauWith builds a full-length tableau, overriding specific columns.
	scorpionTableauWith := func(overrides map[int][]*domain.KlondikeTableauCard) [domain.ScorpionTableauCnt][]*domain.KlondikeTableauCard {
		var tableau [domain.ScorpionTableauCnt][]*domain.KlondikeTableauCard
		for i := range domain.ScorpionTableauCnt {
			if cards, ok := overrides[i]; ok {
				tableau[i] = cards
			}
		}
		return tableau
	}

	t.Run("lists suit-matching column, empty not flagged for non-King", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhasePlaying).Maybe()
		// Column 0 top card: Spade 5 (movable, non-King). Column 1 top: Spade 6 (legal target).
		// Column 2 top: Heart 6 (wrong suit). Column 3 empty (NOT legal for a non-King in Scorpion).
		tableau := scorpionTableauWith(map[int][]*domain.KlondikeTableauCard{
			0: {{Card: domain.NewCard(domain.CardDesignSpade, 5, false), FaceUp: true}},
			1: {{Card: domain.NewCard(domain.CardDesignSpade, 6, false), FaceUp: true}},
			2: {{Card: domain.NewCard(domain.CardDesignHeart, 6, false), FaceUp: true}},
		})
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(ScorpionCuiPresenter)
		result := p.LegalMovesOutput(sg, 0)
		assert.Contains(t, result, "列0")
		assert.Contains(t, result, "列1")    // suit-matching target
		assert.NotContains(t, result, "空列") // empty columns NOT accepted for non-King
	})

	t.Run("flags empty columns for a King", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhasePlaying).Maybe()
		// Column 0 top card: Spade King (movable). Column 3 empty (legal for a King).
		tableau := scorpionTableauWith(map[int][]*domain.KlondikeTableauCard{
			0: {{Card: domain.NewCard(domain.CardDesignSpade, domain.CardValueMax, false), FaceUp: true}},
		})
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(ScorpionCuiPresenter)
		result := p.LegalMovesOutput(sg, 0)
		assert.Contains(t, result, "空列")
	})

	t.Run("no movable card in empty column", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhasePlaying).Maybe()
		sg.On("GetTableau").Return(scorpionTableauWith(nil)).Maybe()

		p := new(ScorpionCuiPresenter)
		result := p.LegalMovesOutput(sg, 0)
		assert.Contains(t, result, "移動可能")
	})

	t.Run("face-down top card", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhasePlaying).Maybe()
		tableau := scorpionTableauWith(map[int][]*domain.KlondikeTableauCard{
			0: {{Card: domain.NewCard(domain.CardDesignSpade, 5, false), FaceUp: false}},
		})
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(ScorpionCuiPresenter)
		result := p.LegalMovesOutput(sg, 0)
		assert.Contains(t, result, "移動可能")
	})

	t.Run("no legal destinations", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhasePlaying).Maybe()
		// Every other column occupied with a non-matching card, none empty; moving card non-King.
		overrides := map[int][]*domain.KlondikeTableauCard{}
		for i := range domain.ScorpionTableauCnt {
			overrides[i] = []*domain.KlondikeTableauCard{
				{Card: domain.NewCard(domain.CardDesignHeart, 2, false), FaceUp: true},
			}
		}
		overrides[0] = []*domain.KlondikeTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, 5, false), FaceUp: true},
		}
		sg.On("GetTableau").Return(scorpionTableauWith(overrides)).Maybe()

		p := new(ScorpionCuiPresenter)
		result := p.LegalMovesOutput(sg, 0)
		assert.Contains(t, result, "なし")
	})

	t.Run("invalid column", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhasePlaying).Maybe()

		p := new(ScorpionCuiPresenter)
		assert.NotEmpty(t, p.LegalMovesOutput(sg, -1))
		assert.NotEmpty(t, p.LegalMovesOutput(sg, domain.ScorpionTableauCnt))
	})

	t.Run("not in playing phase", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhaseGameOver).Maybe()

		p := new(ScorpionCuiPresenter)
		result := p.LegalMovesOutput(sg, 0)
		assert.NotEmpty(t, result)
	})
}

func TestScorpionCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhasePlaying)

		p := new(ScorpionCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.NotEmpty(t, result)
	})

	t.Run("game over", func(t *testing.T) {
		sg := new(interfaces.MockScorpionGame)
		sg.On("GetPhase").Return(domain.ScorpionPhaseGameOver)
		sg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(ScorpionCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.NotEmpty(t, result)
	})
}
