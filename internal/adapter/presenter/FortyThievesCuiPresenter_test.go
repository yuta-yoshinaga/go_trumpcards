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
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupFortyThievesCuiMockDefaults(fg *interfaces.MockFortyThievesGame) {
	fg.On("GetPhase").Return(domain.FortyThievesPhasePlaying).Maybe()
	fg.On("GetMoveCount").Return(0).Maybe()
	fg.On("GetStockCount").Return(64).Maybe()
	fg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	fg.On("IsStalemate").Return(false).Maybe()
	fg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
	for i := range domain.FortyThievesTableauCnt {
		tableau[i] = make([]*domain.FortyThievesTableauCard, 4)
		for j := range 4 {
			tableau[i][j] = &domain.FortyThievesTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+1, false),
				FaceUp: true,
			}
		}
	}
	fg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.FortyThievesFoundationCnt][]*domain.Card
	fg.On("GetFoundation").Return(foundation).Maybe()
}

func TestFortyThievesCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesCuiMockDefaults(fg)
		p := new(FortyThievesCuiPresenter)

		result := p.Output(fg, nil)
		assert.Contains(t, result, "Forty Thieves")
		assert.Contains(t, result, "Foundation")
		assert.Contains(t, result, "Stock: 64枚")
		assert.Contains(t, result, "Waste: [空]")
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "手数: 0")
		assert.Contains(t, result, "操作: m で移動")

		foundationLine := fortyThievesFoundationLine(result)
		assert.Contains(t, foundationLine, "[空: どのAでも可]")
	})

	t.Run("with waste card", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetWaste")
		fg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})

		p := new(FortyThievesCuiPresenter)
		result := p.Output(fg, nil)
		assert.Contains(t, result, "Waste:")
		assert.NotContains(t, result, "Waste: [空]")
	})

	t.Run("with error", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesCuiMockDefaults(fg)
		p := new(FortyThievesCuiPresenter)

		result := p.Output(fg, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game clear", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetPhase")
		fg.On("GetPhase").Return(domain.FortyThievesPhaseGameClear)

		p := new(FortyThievesCuiPresenter)
		result := p.Output(fg, nil)
		assert.Contains(t, result, "ゲームクリア")
		assert.NotContains(t, result, "操作: m で移動")
	})

	t.Run("game over", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetPhase")
		fg.On("GetPhase").Return(domain.FortyThievesPhaseGameOver)

		p := new(FortyThievesCuiPresenter)
		result := p.Output(fg, nil)
		assert.Contains(t, result, "ゲームオーバー")
		assert.NotContains(t, result, "操作: m で移動")
	})

	t.Run("stalemate", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "IsStalemate")
		fg.On("IsStalemate").Return(true)
		fg.On("UndoToEscape").Return(0).Maybe()

		p := new(FortyThievesCuiPresenter)
		result := p.Output(fg, nil)
		assert.Contains(t, result, "手詰まり")
	})

	t.Run("empty column", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetTableau")
		var emptyTableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
		fg.On("GetTableau").Return(emptyTableau)

		p := new(FortyThievesCuiPresenter)
		result := p.Output(fg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("foundation with cards", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		setupFortyThievesCuiMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetFoundation")
		var foundation [domain.FortyThievesFoundationCnt][]*domain.Card
		for i := range domain.FortyThievesFoundationCnt {
			foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		}
		fg.On("GetFoundation").Return(foundation)

		p := new(FortyThievesCuiPresenter)
		result := p.Output(fg, nil)
		foundationLine := fortyThievesFoundationLine(result)
		assert.Contains(t, foundationLine, "SPADE 1")
		assert.NotContains(t, foundationLine, "[空: どのAでも可]")
	})
}

func TestFortyThievesCuiPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		fg.On("GetHint").Return(&domain.FortyThievesHint{
			FromZone:  "tableau",
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(FortyThievesCuiPresenter)
		result := p.HintOutput(fg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "タブロー列0")
		assert.Contains(t, result, "ファンデーション")
	})

	// #5525: ストックだけ残っている局面は行き詰まりではないので、
	// 「ヒントはありません」ではなく「引け」と言う。
	t.Run("stock hint tells the player to draw", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		fg.On("GetHint").Return(&domain.FortyThievesHint{
			FromZone:  "stock",
			FromCol:   -1,
			CardIndex: -1,
			ToZone:    "waste",
			ToCol:     -1,
		})

		result := new(FortyThievesCuiPresenter).HintOutput(fg)
		assert.Contains(t, result, i18n.T("fortythieves.hintDraw"))
		assert.NotContains(t, result, i18n.T("cuiHintNone"))
		// 移動ヒントの体裁 (「A → B」) には落とさない。列 -1 が漏れる。
		assert.NotContains(t, result, "-1")
	})

	t.Run("waste hint", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		fg.On("GetHint").Return(&domain.FortyThievesHint{
			FromZone:  "waste",
			FromCol:   -1,
			CardIndex: -1,
			ToZone:    "tableau",
			ToCol:     3,
		})

		p := new(FortyThievesCuiPresenter)
		result := p.HintOutput(fg)
		assert.Contains(t, result, "ウェイスト")
		assert.Contains(t, result, "タブロー列3")
	})

	t.Run("no hint", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		fg.On("GetHint").Return((*domain.FortyThievesHint)(nil))

		p := new(FortyThievesCuiPresenter)
		result := p.HintOutput(fg)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestFortyThievesCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		fg.On("GetPhase").Return(domain.FortyThievesPhasePlaying)

		p := new(FortyThievesCuiPresenter)
		result := p.ActionLogOutput(fg)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		fg := new(interfaces.MockFortyThievesGame)
		fg.On("GetPhase").Return(domain.FortyThievesPhaseGameOver)
		fg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "draw", Detail: "test"},
		})

		p := new(FortyThievesCuiPresenter)
		result := p.ActionLogOutput(fg)
		assert.Contains(t, result, "draw")
	})
}

// fortyThievesFoundationLine returns the Foundation row of the CUI output.
// The row has to be isolated before asserting on it: "[空]" also appears on the
// waste and tableau rows, so a whole-output Contains would pass no matter what
// the foundation actually rendered.
func fortyThievesFoundationLine(out string) string {
	for _, l := range strings.Split(out, "\n") {
		if strings.HasPrefix(l, i18n.T("fortythieves.foundationHeader")) {
			return l
		}
	}
	return ""
}
