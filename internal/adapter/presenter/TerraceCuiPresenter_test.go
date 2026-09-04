//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupTerraceCuiMockDefaults(g *interfaces.MockTerraceGame) {
	g.On("GetPhase").Return(domain.TerracePhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("GetStockCount").Return(84).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()
	g.On("GetReserve").Return([]*domain.Card{domain.NewCard(domain.CardDesignClover, 3, true)}).Maybe()
	g.On("GetBaseRank").Return(5).Maybe()
	g.On("IsAwaitingBaseRank").Return(false).Maybe()

	var tableau [domain.TerraceTableauCnt][]*domain.Card
	for i := range domain.TerraceTableauCnt {
		tableau[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, i+2, true)}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.TerraceFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func TestTerraceCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceCuiMockDefaults(g)

		result := new(TerraceCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "Terrace")
		assert.Contains(t, result, i18n.T("terrace.foundationHeader"))
		assert.Contains(t, result, "山0:")
		assert.Contains(t, result, "山8:", "all nine piles are rendered")
		assert.Contains(t, result, "84")
		assert.Contains(t, result, i18n.Tf("terrace.baseRankLine", "rank", "5"))
		assert.Contains(t, result, "手数: 0")
	})

	// Until the base rank is fixed, that one decision dominates the board.
	t.Run("awaiting the base rank is called out", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsAwaitingBaseRank")
		g.On("IsAwaitingBaseRank").Return(true)

		result := new(TerraceCuiPresenter).Output(g, nil)
		assert.Contains(t, result, i18n.T("terrace.awaitingBase"))
		assert.NotContains(t, result, i18n.Tf("terrace.baseRankLine", "rank", "5"))
	})

	// The terrace never refills, so its depth is the number that matters.
	t.Run("the terrace shows its depth and its restriction", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceCuiMockDefaults(g)

		assert.Contains(t, new(TerraceCuiPresenter).Output(g, nil),
			i18n.Tf("terrace.reserveLine", "card", "CLOVER 3", "count", "1"))
	})

	t.Run("empty terrace, empty waste and empty piles", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetReserve")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetWaste")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetTableau")
		g.On("GetReserve").Return([]*domain.Card(nil))
		g.On("GetWaste").Return([]*domain.Card(nil))
		g.On("GetTableau").Return([domain.TerraceTableauCnt][]*domain.Card{})

		result := new(TerraceCuiPresenter).Output(g, nil)
		assert.Contains(t, result, i18n.T("terrace.reserveEmpty"))
		assert.Contains(t, result, i18n.T("terrace.wasteEmpty"))
		assert.Contains(t, result, "[空]")
	})

	t.Run("stalemate shows undo-to-escape guidance", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(3)

		assert.Contains(t, new(TerraceCuiPresenter).Output(g, nil),
			i18n.Tf("terrace.undoToEscape", "count", "3"))
	})

	t.Run("stalemate with no escape hides the guidance", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(0)

		result := new(TerraceCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "手詰まり")
		assert.NotContains(t, result, "脱出には")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceCuiMockDefaults(g)

		assert.Contains(t, new(TerraceCuiPresenter).Output(g, errors.New("test error")), "test error")
	})

	t.Run("domain error code prints in correct locale", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		setupTerraceCuiMockDefaults(g)

		err := domain.NewDomainErrorCode(domain.ErrInvalidPlay, "terrace.errStockEmpty", nil)

		i18n.SetLang("ja")
		jaOut := new(TerraceCuiPresenter).Output(g, err)
		assert.Contains(t, jaOut, "山札が空で、再配布もありません")
		assert.NotContains(t, jaOut, "The stock is empty")

		i18n.SetLang("en")
		enOut := new(TerraceCuiPresenter).Output(g, err)
		assert.Contains(t, enOut, "The stock is empty and there is no redeal.")
		assert.NotContains(t, enOut, "山札が空で、再配布もありません")

		i18n.SetLang("ja") // restore
	})

	for _, tc := range []struct {
		name string
		val  domain.TerracePhase
		want string
	}{
		{"game clear", domain.TerracePhaseGameClear, "ゲームクリア"},
		{"game over", domain.TerracePhaseGameOver, "ゲームオーバー"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockTerraceGame)
			setupTerraceCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			assert.Contains(t, new(TerraceCuiPresenter).Output(g, nil), tc.want)
		})
	}
}

func TestTerraceCuiPresenter_HintOutput(t *testing.T) {
	for _, tc := range []struct {
		name     string
		hint     *domain.TerraceHint
		contains []string
	}{
		{"terrace to a foundation",
			&domain.TerraceHint{FromZone: "reserve", FromIdx: -1, ToZone: "foundation", ToIdx: 1},
			[]string{"テラス", "基礎札1"}},
		{"waste to a foundation",
			&domain.TerraceHint{FromZone: "waste", FromIdx: -1, ToZone: "foundation", ToIdx: 2},
			[]string{"捨て札", "基礎札2"}},
		{"waste to a pile",
			&domain.TerraceHint{FromZone: "waste", FromIdx: -1, ToZone: "tableau", ToIdx: 3},
			[]string{"捨て札", "タブロー山3"}},
		{"between piles",
			&domain.TerraceHint{FromZone: "tableau", FromIdx: 1, ToZone: "tableau", ToIdx: 5},
			[]string{"タブロー山1", "タブロー山5"}},
		{"draw from the stock",
			&domain.TerraceHint{FromZone: "stock", FromIdx: -1, ToZone: "waste", ToIdx: -1},
			[]string{"山札", i18n.T("terrace.hintToWaste")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockTerraceGame)
			g.On("GetHint").Return(tc.hint)

			result := new(TerraceCuiPresenter).HintOutput(g)
			assert.Contains(t, result, "ヒント")
			for _, want := range tc.contains {
				assert.Contains(t, result, want)
			}
		})
	}

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		g.On("GetHint").Return((*domain.TerraceHint)(nil))

		assert.Contains(t, new(TerraceCuiPresenter).HintOutput(g), "ヒントはありません")
	})
}

func TestTerraceCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		g.On("GetPhase").Return(domain.TerracePhasePlaying)

		assert.Contains(t, new(TerraceCuiPresenter).ActionLogOutput(g), "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockTerraceGame)
		g.On("GetPhase").Return(domain.TerracePhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(TerraceCuiPresenter).ActionLogOutput(g), "move")
	})
}
