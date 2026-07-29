//go:build test

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupBraidCuiMockDefaults(g *interfaces.MockBraidGame) {
	g.On("GetPhase").Return(domain.BraidPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(7).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(71).Maybe()
	g.On("GetPassesUsed").Return(0).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()
	g.On("GetBraid").Return([]*domain.Card{domain.NewCard(domain.CardDesignClover, 3, true)}).Maybe()
	g.On("GetBaseRank").Return(5).Maybe()
	g.On("GetDirection").Return(domain.BraidDirectionAscending).Maybe()
	g.On("IsAwaitingDirection").Return(false).Maybe()

	var fields [domain.BraidFieldCnt]*domain.Card
	for i := range domain.BraidFieldCnt {
		fields[i] = domain.NewCard(domain.CardDesignSpade, i+2, true)
	}
	g.On("GetFields").Return(fields).Maybe()

	var helpers [domain.BraidHelperCnt]*domain.Card
	g.On("GetHelpers").Return(helpers).Maybe()

	var foundation [domain.BraidFoundationCnt][]*domain.Card
	foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)}
	g.On("GetFoundation").Return(foundation).Maybe()
}

func TestBraidCuiPresenter_Output(t *testing.T) {
	i18n.SetLang("ja")

	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		setupBraidCuiMockDefaults(g)

		out := new(BraidCuiPresenter).Output(g, nil)
		assert.Contains(t, out, i18n.T("braid.foundationHeader"))
		assert.Contains(t, out, "5")
		assert.Contains(t, out, i18n.T("braid.directionAscending"))
		assert.Contains(t, out, i18n.T("braid.fieldLabel"))
		assert.Contains(t, out, i18n.T("braid.helperLabel"))
		// 空のヘルパーは枠として残り、番号が保たれる。
		assert.Contains(t, out, i18n.T("cuiEmptyCol"))
	})

	t.Run("awaiting the direction leads the view", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		setupBraidCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsAwaitingDirection")
		g.On("IsAwaitingDirection").Return(true)

		out := new(BraidCuiPresenter).Output(g, nil)
		assert.Contains(t, out, "dir")
		// 盤面より先に出ること。向きを決めるまで基礎札に触れないので、これが最重要。
		assert.Less(t, strings.Index(out, "dir"), strings.Index(out, i18n.T("braid.foundationHeader")))
	})

	t.Run("descending direction", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		setupBraidCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetDirection")
		g.On("GetDirection").Return(domain.BraidDirectionDescending)

		assert.Contains(t, new(BraidCuiPresenter).Output(g, nil), i18n.T("braid.directionDescending"))
	})

	t.Run("empty braid and waste", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		setupBraidCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetBraid")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetWaste")
		g.On("GetBraid").Return([]*domain.Card(nil))
		g.On("GetWaste").Return([]*domain.Card(nil))

		out := new(BraidCuiPresenter).Output(g, nil)
		assert.Contains(t, out, i18n.T("braid.braidEmpty"))
		assert.Contains(t, out, i18n.T("braid.wasteEmpty"))
	})

	t.Run("stalemate shows the escape hint", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		setupBraidCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "UndoToEscape")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(3)

		out := new(BraidCuiPresenter).Output(g, nil)
		assert.Contains(t, out, i18n.T("cuiSolitaireStalemate"))
		assert.Contains(t, out, "3")
	})

	for _, tc := range []struct {
		name string
		val  domain.BraidPhase
		key  string
	}{
		{"game clear", domain.BraidPhaseGameClear, "cuiSolitaireGameClear"},
		{"game over", domain.BraidPhaseGameOver, "cuiSolitaireGameOver"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockBraidGame)
			setupBraidCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			assert.Contains(t, new(BraidCuiPresenter).Output(g, nil), i18n.T(tc.key))
		})
	}

	t.Run("error block", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		setupBraidCuiMockDefaults(g)

		assert.Contains(t, new(BraidCuiPresenter).Output(g, assertError{}), "boom")
	})
}

// assertError is a minimal error for the presenter's error block.
type assertError struct{}

func (assertError) Error() string { return "boom" }

func TestBraidCuiPresenter_HintOutput(t *testing.T) {
	i18n.SetLang("ja")

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		g.On("GetHint").Return((*domain.BraidHint)(nil))
		assert.Contains(t, new(BraidCuiPresenter).HintOutput(g), i18n.T("cuiHintNone"))
	})

	// 向きの選択は移動ではないので、from → to の形にせず専用の一文で出す。
	t.Run("choosing the direction gets its own sentence", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		g.On("GetHint").Return(&domain.BraidHint{FromZone: "direction", FromIdx: -1, ToZone: "foundation", ToIdx: -1})
		out := new(BraidCuiPresenter).HintOutput(g)
		assert.Equal(t, i18n.T("braid.hintChooseDirection")+"\n", out)
	})

	for _, tc := range []struct {
		name string
		hint *domain.BraidHint
		want []string
	}{
		{"braid to foundation", &domain.BraidHint{FromZone: "braid", FromIdx: -1, ToZone: "foundation", ToIdx: 2},
			[]string{i18n.T("braid.hintFromBraid"), "2"}},
		{"field to foundation", &domain.BraidHint{FromZone: "field", FromIdx: 1, ToZone: "foundation", ToIdx: 0},
			[]string{i18n.Tf("braid.hintFromField", "idx", "1")}},
		{"helper to foundation", &domain.BraidHint{FromZone: "helper", FromIdx: 4, ToZone: "foundation", ToIdx: 0},
			[]string{i18n.Tf("braid.hintFromHelper", "idx", "4")}},
		{"waste to helper", &domain.BraidHint{FromZone: "waste", FromIdx: -1, ToZone: "helper", ToIdx: 3},
			[]string{i18n.T("braid.hintFromWaste"), i18n.Tf("braid.hintToHelper", "idx", "3")}},
		{"stock to waste", &domain.BraidHint{FromZone: "stock", FromIdx: -1, ToZone: "waste", ToIdx: -1},
			[]string{i18n.T("braid.hintFromStock"), i18n.T("braid.hintToWaste")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockBraidGame)
			g.On("GetHint").Return(tc.hint)
			out := new(BraidCuiPresenter).HintOutput(g)
			for _, want := range tc.want {
				assert.Contains(t, out, want)
			}
		})
	}
}

func TestBraidCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase hides the log", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		g.On("GetPhase").Return(domain.BraidPhasePlaying)
		assert.NotContains(t, new(BraidCuiPresenter).ActionLogOutput(g), "move")
	})

	t.Run("game over shows the log", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		g.On("GetPhase").Return(domain.BraidPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test detail"},
		})
		assert.Contains(t, new(BraidCuiPresenter).ActionLogOutput(g), "test detail")
	})
}
