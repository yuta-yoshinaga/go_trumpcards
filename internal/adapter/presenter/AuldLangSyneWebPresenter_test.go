//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupAuldLangSyneWebMockDefaults(g *interfaces.MockAuldLangSyneGame) {
	g.On("GetPhase").Return(domain.AuldLangSynePhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(44).Maybe()
	g.On("GetFoundations").Return(acesOnMockFoundations()).Maybe()

	var wastes [domain.AuldLangSyneWasteCnt][]*domain.Card
	wastes[0] = []*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, false)}
	g.On("GetWastes").Return(wastes).Maybe()
}

func parseAuldLangSyneOutput(t *testing.T, jsonStr string) *controller.AuldLangSyneWebOutput {
	t.Helper()
	var out controller.AuldLangSyneWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupAuldLangSyneOutputMock is the Output() default. **Output() fills the
// passive hint too** (#4483), so GetHint must be callable. Putting this in the
// shared helper would let the first-registered expectation eat the HintOutput
// tests' "hint present" case.
func setupAuldLangSyneOutputMock(g *interfaces.MockAuldLangSyneGame) {
	setupAuldLangSyneWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestAuldLangSyneWebPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		setupAuldLangSyneOutputMock(g)
		p := new(AuldLangSyneWebPresenter)

		out := parseAuldLangSyneOutput(t, p.Output(g, nil))
		assert.Equal(t, "auldlangsyne.playing", out.MessageCode)
		assert.Len(t, out.Foundations, domain.AuldLangSyneFoundationCnt)
		assert.Len(t, out.Wastes, domain.AuldLangSyneWasteCnt)
		assert.Equal(t, 44, out.StockCount)
	})

	// The passive hint is what the page's highlight reads; without it every
	// hint-driven branch on the page is dead (#4483).
	t.Run("passive hint is filled while playing", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		g.On("GetHint").Return(&domain.AuldLangSyneHint{WasteIdx: 1, FoundationIdx: 3})
		setupAuldLangSyneWebMockDefaults(g)
		p := new(AuldLangSyneWebPresenter)

		out := parseAuldLangSyneOutput(t, p.Output(g, nil))
		assert.NotNil(t, out.Hint)
		assert.Equal(t, 1, out.Hint.WasteIdx)
		assert.Equal(t, 3, out.Hint.FoundationIdx)
	})

	t.Run("no passive hint once stalemate", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		g.On("IsStalemate").Return(true)
		setupAuldLangSyneOutputMock(g)
		p := new(AuldLangSyneWebPresenter)

		out := parseAuldLangSyneOutput(t, p.Output(g, nil))
		assert.Nil(t, out.Hint)
		assert.Equal(t, "auldlangsyne.stalemate", out.MessageCode)
	})

	t.Run("game clear", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		g.On("GetPhase").Return(domain.AuldLangSynePhaseGameClear)
		g.On("GetMoveCount").Return(37)
		setupAuldLangSyneOutputMock(g)
		p := new(AuldLangSyneWebPresenter)

		out := parseAuldLangSyneOutput(t, p.Output(g, nil))
		assert.Equal(t, "auldlangsyne.gameClear", out.MessageCode)
		assert.Equal(t, "37", out.MessageParams["moveCount"])
	})

	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		g.On("GetPhase").Return(domain.AuldLangSynePhaseGameOver)
		setupAuldLangSyneOutputMock(g)
		p := new(AuldLangSyneWebPresenter)

		out := parseAuldLangSyneOutput(t, p.Output(g, nil))
		assert.Equal(t, "auldlangsyne.gameOver", out.MessageCode)
	})

	t.Run("error takes precedence over the phase message", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		setupAuldLangSyneOutputMock(g)
		p := new(AuldLangSyneWebPresenter)

		out := parseAuldLangSyneOutput(t, p.Output(g, errors.New("stock is empty")))
		assert.Equal(t, "stock is empty", out.Message)
		assert.Empty(t, out.MessageCode)
	})
}

func TestAuldLangSyneWebPresenter_HintOutput(t *testing.T) {
	t.Run("hint available", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		g.On("GetHint").Return(&domain.AuldLangSyneHint{WasteIdx: 2, FoundationIdx: 0})
		setupAuldLangSyneWebMockDefaults(g)
		p := new(AuldLangSyneWebPresenter)

		out := parseAuldLangSyneOutput(t, p.HintOutput(g))
		assert.Equal(t, "auldlangsyne.hintAvailable", out.MessageCode)
		assert.Equal(t, 2, out.Hint.WasteIdx)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		g.On("GetHint").Return(nil)
		setupAuldLangSyneWebMockDefaults(g)
		p := new(AuldLangSyneWebPresenter)

		out := parseAuldLangSyneOutput(t, p.HintOutput(g))
		assert.Equal(t, "auldlangsyne.noHint", out.MessageCode)
		assert.Nil(t, out.Hint)
	})
}

func TestAuldLangSyneWebPresenter_ActionLogOutput(t *testing.T) {
	g := new(interfaces.MockAuldLangSyneGame)
	// actionLogOutputJSON gates on the end flag: the transcript is only emitted
	// once the game is over.
	g.On("GetGameEndFlag").Return(true)
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, ActionType: "deal", Detail: "4枚を配りました"},
	})
	p := new(AuldLangSyneWebPresenter)

	assert.Contains(t, p.ActionLogOutput(g), "deal")
}
