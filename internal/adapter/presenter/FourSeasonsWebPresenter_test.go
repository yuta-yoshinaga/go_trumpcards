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

func setupFourSeasonsWebMockDefaults(g *interfaces.MockFourSeasonsGame) {
	g.On("GetPhase").Return(domain.FourSeasonsPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("GetBaseRank").Return(7).Maybe()
	g.On("GetStockCount").Return(44).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 6, false)}).Maybe()

	var fnd [domain.FourSeasonsFoundationCnt][]*domain.Card
	fnd[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}
	g.On("GetFoundations").Return(fnd).Maybe()

	var tab [domain.FourSeasonsTableauCnt][]*domain.Card
	tab[0] = []*domain.Card{domain.NewCard(domain.CardDesignHeart, 12, false)}
	g.On("GetTableau").Return(tab).Maybe()
}

func parseFourSeasonsOutput(t *testing.T, jsonStr string) *controller.FourSeasonsWebOutput {
	t.Helper()
	var out controller.FourSeasonsWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

// Output() also fills the passive hint (#4483); registering GetHint here would
// eat the HintOutput tests' "hint present" case, so it stays local.
func setupFourSeasonsOutputMock(g *interfaces.MockFourSeasonsGame) {
	setupFourSeasonsWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestFourSeasonsWebPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockFourSeasonsGame)
		setupFourSeasonsOutputMock(g)
		out := parseFourSeasonsOutput(t, new(FourSeasonsWebPresenter).Output(g, nil))

		assert.Equal(t, "fourseasons.playing", out.MessageCode)
		assert.Len(t, out.Tableau, domain.FourSeasonsTableauCnt)
		assert.Len(t, out.Foundation, domain.FourSeasonsFoundationCnt)
		assert.Equal(t, 44, out.StockCount)
		// The base rank must reach the client: the page renders every "next
		// required rank" badge from it.
		assert.Equal(t, 7, out.BaseRank)
	})

	t.Run("passive hint is filled while playing", func(t *testing.T) {
		g := new(interfaces.MockFourSeasonsGame)
		g.On("GetHint").Return(&domain.FourSeasonsHint{FromZone: "tableau", FromIdx: 3, ToZone: "foundation", ToIdx: 2})
		setupFourSeasonsWebMockDefaults(g)
		out := parseFourSeasonsOutput(t, new(FourSeasonsWebPresenter).Output(g, nil))

		assert.NotNil(t, out.Hint)
		assert.Equal(t, "tableau", out.Hint.FromZone)
		assert.Equal(t, 3, out.Hint.FromIdx)
		assert.Equal(t, 2, out.Hint.ToIdx)
	})

	t.Run("game clear", func(t *testing.T) {
		g := new(interfaces.MockFourSeasonsGame)
		g.On("GetPhase").Return(domain.FourSeasonsPhaseGameClear)
		g.On("GetMoveCount").Return(42)
		setupFourSeasonsOutputMock(g)
		out := parseFourSeasonsOutput(t, new(FourSeasonsWebPresenter).Output(g, nil))

		assert.Equal(t, "fourseasons.gameClear", out.MessageCode)
		assert.Equal(t, "42", out.MessageParams["moveCount"])
	})

	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockFourSeasonsGame)
		g.On("GetPhase").Return(domain.FourSeasonsPhaseGameOver)
		setupFourSeasonsOutputMock(g)
		out := parseFourSeasonsOutput(t, new(FourSeasonsWebPresenter).Output(g, nil))
		assert.Equal(t, "fourseasons.gameOver", out.MessageCode)
	})

	t.Run("error takes precedence over the phase message", func(t *testing.T) {
		g := new(interfaces.MockFourSeasonsGame)
		setupFourSeasonsOutputMock(g)
		out := parseFourSeasonsOutput(t, new(FourSeasonsWebPresenter).Output(g, errors.New("stock is empty")))
		assert.Equal(t, "stock is empty", out.Message)
		assert.Empty(t, out.MessageCode)
	})
}

func TestFourSeasonsWebPresenter_HintOutput(t *testing.T) {
	t.Run("hint available", func(t *testing.T) {
		g := new(interfaces.MockFourSeasonsGame)
		g.On("GetHint").Return(&domain.FourSeasonsHint{FromZone: "waste", FromIdx: -1, ToZone: "foundation", ToIdx: 0})
		setupFourSeasonsWebMockDefaults(g)
		out := parseFourSeasonsOutput(t, new(FourSeasonsWebPresenter).HintOutput(g))

		assert.Equal(t, "fourseasons.hintAvailable", out.MessageCode)
		assert.Equal(t, "waste", out.Hint.FromZone)
		assert.Equal(t, -1, out.Hint.FromIdx)
	})
	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockFourSeasonsGame)
		g.On("GetHint").Return(nil)
		setupFourSeasonsWebMockDefaults(g)
		out := parseFourSeasonsOutput(t, new(FourSeasonsWebPresenter).HintOutput(g))

		assert.Equal(t, "fourseasons.noHint", out.MessageCode)
		assert.Nil(t, out.Hint)
	})
}

func TestFourSeasonsWebPresenter_ActionLogOutput(t *testing.T) {
	g := new(interfaces.MockFourSeasonsGame)
	// actionLogOutputJSON gates on the end flag.
	g.On("GetGameEndFlag").Return(true)
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "draw", Detail: "引きました"}})
	assert.Contains(t, new(FourSeasonsWebPresenter).ActionLogOutput(g), "draw")
}
