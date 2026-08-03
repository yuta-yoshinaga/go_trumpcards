//go:build test

package presenter

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func ctTestGame(t *testing.T) *domain.ChineseTen {
	t.Helper()
	c := domain.NewDefaultChineseTen()
	c.Reset()
	return c
}

func ctDecode(t *testing.T, raw string) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return out
}

// ctStub wires a MockChineseTenGame with every accessor the presenters touch,
// so a test can pin an exact phase/winner rather than shuffling until one
// appears.
func ctStub(phase domain.ChineseTenPhase, winner int, gameEnd bool) *interfaces.MockChineseTenGame {
	g := new(interfaces.MockChineseTenGame)
	g.On("GetPhase").Return(phase)
	g.On("GetWinnerIdx").Return(winner)
	g.On("GetGameEndFlag").Return(gameEnd)
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("GetStockCount").Return(0)
	g.On("GetLayout").Return([]*domain.Card{})
	g.On("GetPendingCard").Return((*domain.Card)(nil))
	g.On("GetSelectableIndices").Return([]int{})
	g.On("GetConfig").Return(domain.DefaultChineseTenConfig())
	g.On("GetPlayers").Return([]*domain.ChineseTenPlayer{
		domain.NewChineseTenPlayer(true), domain.NewChineseTenPlayer(false),
	})
	g.On("GetPlayer", mock.Anything).Return(domain.NewChineseTenPlayer(false))
	g.On("GetCaptured", mock.Anything).Return([]*domain.Card{})
	g.On("GetScore", mock.Anything).Return(0)
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	g.On("ChineseTenCpuDecide", mock.Anything).Return(domain.ChineseTenCpuAction{HandIdx: 0, LayoutIdx: 0})
	return g
}

func TestChineseTenWebPresenter_HidesTheCpuHandButNeverItsCaptures(t *testing.T) {
	// Which cards have already gone is public -- it is how the remaining ones
	// are worked out. Only the hand is withheld.
	out := ctDecode(t, new(ChineseTenWebPresenter).Output(ctTestGame(t), nil))
	players := out["players"].([]any)
	require.Len(t, players, domain.ChineseTenPlayerCnt)

	human := players[0].(map[string]any)
	assert.Equal(t, false, human["hidden"])
	assert.Len(t, human["cards"], domain.ChineseTenHandSize)

	cpu := players[1].(map[string]any)
	assert.Equal(t, true, cpu["hidden"])
	assert.Empty(t, cpu["cards"], "the CPU's hand must not reach the client")
	assert.Equal(t, float64(domain.ChineseTenHandSize), cpu["cardCount"])
	assert.NotNil(t, cpu["captured"], "captures are public and must always be sent")
}

func TestChineseTenWebPresenter_CardsCarryTheirValue(t *testing.T) {
	// The 52-card scoring table lives on the server; sending points/isRed means
	// the client never keeps a second copy.
	out := ctDecode(t, new(ChineseTenWebPresenter).Output(ctTestGame(t), nil))
	layout := out["layout"].([]any)
	require.NotEmpty(t, layout)
	for _, raw := range layout {
		card := raw.(map[string]any)
		assert.Contains(t, card, "points")
		assert.Contains(t, card, "isRed")
		if card["isRed"] == false {
			assert.Equal(t, float64(0), card["points"], "black cards score nothing")
		}
	}
	assert.Equal(t, float64(domain.ChineseTenTieScore), out["tieScore"])
}

func TestChineseTenWebPresenter_ShipsSelectableIndices(t *testing.T) {
	// Both capture rules live here; the list is what stops the client from
	// re-implementing a pair of rules that do not overlap.
	c := ctTestGame(t)
	c.SetLayoutForTest([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, true),
		domain.NewCard(domain.CardDesignDiamond, 5, true),
		domain.NewCard(domain.CardDesignSpade, 13, true),
	})
	c.SetStockForTest([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, true)})
	p := c.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignClover, 5, true))
	c.SetCurrentPlayerForTest(0)
	require.NoError(t, c.PlayCard(0, 0))

	out := ctDecode(t, new(ChineseTenWebPresenter).Output(c, nil))
	assert.Len(t, out["selectableIndices"], 2, "both fives, and not the king")
	assert.NotNil(t, out["pendingCard"])
}

func TestChineseTenWebPresenter_ReportsEveryOutcome(t *testing.T) {
	for _, tc := range []struct {
		name   string
		winner int
		want   string
	}{
		{"human wins", 0, "chineseten.win"},
		{"draw", -1, "chineseten.draw"},
		{"cpu wins", 1, "chineseten.lose"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := ctDecode(t, new(ChineseTenWebPresenter).Output(ctStub(domain.ChineseTenPhaseGameEnd, tc.winner, true), nil))
			assert.Equal(t, tc.want, out["messageCode"])
		})
	}
}

func TestChineseTenWebPresenter_SurfacesAnError(t *testing.T) {
	out := ctDecode(t, new(ChineseTenWebPresenter).Output(ctTestGame(t), assert.AnError))
	assert.Equal(t, assert.AnError.Error(), out["message"])
	assert.Empty(t, out["messageCode"])
}

func TestChineseTenWebPresenter_ShipsTheHintOnAnOrdinaryResponse(t *testing.T) {
	out := ctDecode(t, new(ChineseTenWebPresenter).Output(ctTestGame(t), nil))
	hint, ok := out["hint"].(map[string]any)
	require.True(t, ok, "the hint must ride along with ordinary state")
	assert.Equal(t, "chineseten.hint.play", hint["reason"])
}

func TestChineseTenWebPresenter_HintCoversEveryBranch(t *testing.T) {
	p := new(ChineseTenWebPresenter)

	t.Run("a finished game", func(t *testing.T) {
		out := ctDecode(t, p.HintOutput(ctStub(domain.ChineseTenPhaseGameEnd, 0, true)))
		assert.Equal(t, "chineseten.hint.game_end", out["hint"].(map[string]any)["reason"])
	})

	t.Run("someone else's turn", func(t *testing.T) {
		g := ctStub(domain.ChineseTenPhasePlay, -1, false)
		g.ExpectedCalls = nil
		g.On("GetPhase").Return(domain.ChineseTenPhasePlay)
		g.On("GetGameEndFlag").Return(false)
		g.On("GetCurrentPlayerIdx").Return(1)
		g.On("GetWinnerIdx").Return(-1)
		g.On("GetStockCount").Return(0)
		g.On("GetLayout").Return([]*domain.Card{})
		g.On("GetPendingCard").Return((*domain.Card)(nil))
		g.On("GetSelectableIndices").Return([]int{})
		g.On("GetConfig").Return(domain.DefaultChineseTenConfig())
		g.On("GetPlayers").Return([]*domain.ChineseTenPlayer{domain.NewChineseTenPlayer(true)})
		g.On("GetCaptured", mock.Anything).Return([]*domain.Card{})
		g.On("GetScore", mock.Anything).Return(0)

		out := ctDecode(t, p.HintOutput(g))
		assert.Equal(t, "chineseten.hint.not_your_turn", out["hint"].(map[string]any)["reason"])
	})

	t.Run("a pending selection", func(t *testing.T) {
		g := ctStub(domain.ChineseTenPhaseSelect, -1, false)
		g.ExpectedCalls = nil
		g.On("GetPhase").Return(domain.ChineseTenPhaseSelect)
		g.On("GetGameEndFlag").Return(false)
		g.On("GetCurrentPlayerIdx").Return(0)
		g.On("GetWinnerIdx").Return(-1)
		g.On("GetStockCount").Return(0)
		g.On("GetLayout").Return([]*domain.Card{})
		g.On("GetPendingCard").Return((*domain.Card)(nil))
		g.On("GetSelectableIndices").Return([]int{1})
		g.On("GetConfig").Return(domain.DefaultChineseTenConfig())
		g.On("GetPlayers").Return([]*domain.ChineseTenPlayer{domain.NewChineseTenPlayer(true)})
		g.On("GetCaptured", mock.Anything).Return([]*domain.Card{})
		g.On("GetScore", mock.Anything).Return(0)
		g.On("ChineseTenCpuDecide", 0).Return(domain.ChineseTenCpuAction{HandIdx: -1, LayoutIdx: 1})

		out := ctDecode(t, p.HintOutput(g))
		hint := out["hint"].(map[string]any)
		assert.Equal(t, "chineseten.hint.select", hint["reason"])
		assert.Equal(t, float64(1), hint["layoutIndex"])
	})

	t.Run("nothing to suggest", func(t *testing.T) {
		for _, tc := range []struct {
			name   string
			phase  domain.ChineseTenPhase
			action domain.ChineseTenCpuAction
		}{
			{"no card", domain.ChineseTenPhasePlay, domain.ChineseTenCpuAction{HandIdx: -1, LayoutIdx: -1}},
			{"no layout card", domain.ChineseTenPhaseSelect, domain.ChineseTenCpuAction{HandIdx: -1, LayoutIdx: -1}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				g := ctStub(tc.phase, -1, false)
				g.ExpectedCalls = nil
				g.On("GetPhase").Return(tc.phase)
				g.On("GetGameEndFlag").Return(false)
				g.On("GetCurrentPlayerIdx").Return(0)
				g.On("GetWinnerIdx").Return(-1)
				g.On("GetStockCount").Return(0)
				g.On("GetLayout").Return([]*domain.Card{})
				g.On("GetPendingCard").Return((*domain.Card)(nil))
				g.On("GetSelectableIndices").Return([]int{})
				g.On("GetConfig").Return(domain.DefaultChineseTenConfig())
				g.On("GetPlayers").Return([]*domain.ChineseTenPlayer{domain.NewChineseTenPlayer(true)})
				g.On("GetCaptured", mock.Anything).Return([]*domain.Card{})
				g.On("GetScore", mock.Anything).Return(0)
				// A -1 index shipped as-is would have the page highlight card "-1".
				g.On("ChineseTenCpuDecide", 0).Return(tc.action)

				out := ctDecode(t, p.HintOutput(g))
				hint := out["hint"].(map[string]any)
				assert.Equal(t, "chineseten.hint.none", hint["reason"])
				assert.NotContains(t, hint, "cardIndex")
				assert.NotContains(t, hint, "layoutIndex")
			})
		}
	})
}

func TestChineseTenWebPresenter_SkipsANilSeatAndRendersTheLog(t *testing.T) {
	g := ctStub(domain.ChineseTenPhasePlay, -1, false)
	g.ExpectedCalls = nil
	g.On("GetPhase").Return(domain.ChineseTenPhasePlay)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetCurrentPlayerIdx").Return(-1)
	g.On("GetWinnerIdx").Return(-1)
	g.On("GetStockCount").Return(0)
	g.On("GetLayout").Return([]*domain.Card{})
	g.On("GetPendingCard").Return((*domain.Card)(nil))
	g.On("GetSelectableIndices").Return([]int{})
	g.On("GetConfig").Return(domain.DefaultChineseTenConfig())
	g.On("GetPlayers").Return([]*domain.ChineseTenPlayer{domain.NewChineseTenPlayer(true), nil})
	g.On("GetCaptured", mock.Anything).Return([]*domain.Card{})
	g.On("GetScore", mock.Anything).Return(0)
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{})

	out := ctDecode(t, new(ChineseTenWebPresenter).Output(g, nil))
	assert.Len(t, out["players"], 1, "the nil seat is dropped, not rendered")
	assert.NotEmpty(t, new(ChineseTenWebPresenter).ActionLogOutput(g))
	assert.Nil(t, chineseTenCardOutput(nil))
}

func TestChineseTenCuiPresenter_PrintsTheRulesAndHidesTheCpuHand(t *testing.T) {
	c := ctTestGame(t)
	out := new(ChineseTenCuiPresenter).Output(c, nil)

	// Both rules on every screen: they do not overlap and that is the trap.
	assert.Contains(t, out, "A〜9は合計10")
	// Exactly one indexed hand is printed -- the human's. Counting "[0]"
	// rather than "[": ANSI colour escapes contain one.
	assert.Equal(t, 2, strings.Count(out, "[0]"), "the layout and the human hand, and nothing else")
}

func TestChineseTenCuiPresenter_RendersEachPhaseAndAnnotatesRedCards(t *testing.T) {
	p := new(ChineseTenCuiPresenter)
	assert.Contains(t, p.Output(ctTestGame(t), assert.AnError), assert.AnError.Error())

	// A red card carries its value; a black one does not.
	assert.Contains(t, chineseTenCardStr(domain.NewCard(domain.CardDesignHeart, 5, true)), "(5)")
	assert.NotContains(t, chineseTenCardStr(domain.NewCard(domain.CardDesignSpade, 5, true)), "(")
	assert.Equal(t, "--", chineseTenCardStr(nil))
	assert.Equal(t, "-", chineseTenCardListStr(nil, false))

	assert.NotEmpty(t, p.Output(ctStub(domain.ChineseTenPhaseSelect, -1, false), nil))
	assert.NotEmpty(t, p.Output(ctStub(domain.ChineseTenPhaseGameEnd, 0, true), nil))
	assert.NotEmpty(t, p.Output(ctStub(domain.ChineseTenPhaseGameEnd, -1, true), nil))
	assert.NotEmpty(t, p.Output(ctStub(domain.ChineseTenPhaseGameEnd, 1, true), nil))
	assert.NotEmpty(t, p.ActionLogOutput(ctTestGame(t)))
}

func TestChineseTenCuiPresenter_HintResolvesItsReasonKey(t *testing.T) {
	for range 50 {
		out := new(ChineseTenCuiPresenter).HintOutput(ctTestGame(t))
		assert.NotContains(t, out, "chineseten.hint.", "the reason must be translated, not printed raw")
		assert.NotEmpty(t, strings.TrimSpace(out))
	}
	// The bare-reason shape, with nothing to point at.
	bare := new(ChineseTenCuiPresenter).HintOutput(ctStub(domain.ChineseTenPhaseGameEnd, 0, true))
	assert.NotContains(t, bare, "手札[")
	assert.NotContains(t, bare, "場札[")
}

func TestChineseTenCuiPresenter_EveryReasonTheHintCanReturnIsMapped(t *testing.T) {
	// The fallback for an unmapped reason is unreachable TODAY, and this is
	// what makes that true.
	for _, reason := range []string{
		"chineseten.hint.game_end", "chineseten.hint.not_your_turn",
		"chineseten.hint.select", "chineseten.hint.play", "chineseten.hint.none",
	} {
		assert.NotEmpty(t, chineseTenHintReasonKeys[reason], "reason %q has no i18n key", reason)
	}
	assert.Len(t, chineseTenHintReasonKeys, 5, "a new reason needs an entry here too")
}
