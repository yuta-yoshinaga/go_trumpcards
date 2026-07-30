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

func toepenTestGame(t *testing.T) *domain.Toepen {
	t.Helper()
	tp := domain.NewDefaultToepen()
	tp.Reset()
	return tp
}

func toepenDecode(t *testing.T, raw string) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return out
}

// toepenStub wires a MockToepenGame with every accessor the presenters touch,
// so a test can pin an exact phase/winner rather than shuffling until one
// appears.
func toepenStub(phase domain.ToepenPhase, winner int, gameEnd bool) *interfaces.MockToepenGame {
	g := new(interfaces.MockToepenGame)
	g.On("GetPhase").Return(phase)
	g.On("GetWinnerIdx").Return(winner)
	g.On("GetGameEndFlag").Return(gameEnd)
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("GetLeadPlayerIdx").Return(0)
	g.On("GetDealerIdx").Return(0)
	g.On("GetCurrentTrick").Return([]*domain.TrickCard{})
	g.On("GetLeadSuit").Return(-1)
	g.On("GetTrickNumber").Return(0)
	g.On("GetStake").Return(1)
	g.On("GetKnockerIdx").Return(-1)
	g.On("GetPendingRespondent").Return(-1)
	g.On("GetLastTrickWinner").Return(-1)
	g.On("GetHandNumber").Return(1)
	g.On("GetConfig").Return(domain.DefaultToepenConfig())
	g.On("GetPlayers").Return([]*domain.ToepenPlayer{
		domain.NewToepenPlayer(true), domain.NewToepenPlayer(false),
	})
	g.On("GetPlayer", mock.Anything).Return(domain.NewToepenPlayer(false))
	g.On("GetLives", mock.Anything).Return(0)
	g.On("IsFolded", mock.Anything).Return(false)
	g.On("IsEliminated", mock.Anything).Return(false)
	g.On("GetValidPlayIndices", mock.Anything).Return([]int{})
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	g.On("ToepenCpuDecide", mock.Anything).Return(domain.ToepenCpuAction{HandIdx: 0})
	return g
}

func TestToepenWebPresenter_HidesTheCpuHandsButNotTheirLives(t *testing.T) {
	// Lives, folded and eliminated are public -- they are what the table reads
	// to decide whether a toep will stick. Only the CARDS are withheld.
	out := toepenDecode(t, new(ToepenWebPresenter).Output(toepenTestGame(t), nil))
	players := out["players"].([]any)
	require.Len(t, players, domain.ToepenPlayerCnt)

	human := players[0].(map[string]any)
	assert.Equal(t, false, human["hidden"])
	assert.Len(t, human["cards"], domain.ToepenHandSize)

	for i := 1; i < domain.ToepenPlayerCnt; i++ {
		cpu := players[i].(map[string]any)
		assert.Equal(t, true, cpu["hidden"], "seat %d", i)
		assert.Empty(t, cpu["cards"], "seat %d's cards must not reach the client", i)
		assert.Equal(t, float64(domain.ToepenHandSize), cpu["cardCount"])
		assert.Contains(t, cpu, "lives")
		assert.Contains(t, cpu, "folded")
	}
}

func TestToepenWebPresenter_ShipsTheFollowSuitDecision(t *testing.T) {
	tp := toepenTestGame(t)
	out := toepenDecode(t, new(ToepenWebPresenter).Output(tp, nil))
	// validPlayIndices is the human's legality list, computed once here.
	assert.Len(t, out["validPlayIndices"], len(tp.GetValidPlayIndices(0)))
	assert.Equal(t, float64(domain.ToepenMaxLives), out["maxLives"])
	assert.Equal(t, float64(1), out["stake"])
}

func TestToepenWebPresenter_ReportsTheOutcome(t *testing.T) {
	for _, tc := range []struct {
		name   string
		winner int
		want   string
	}{
		{"human wins", 0, "toepen.win"},
		{"human loses", 2, "toepen.lose"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := toepenDecode(t, new(ToepenWebPresenter).Output(toepenStub(domain.ToepenPhaseGameEnd, tc.winner, true), nil))
			assert.Equal(t, tc.want, out["messageCode"])
		})
	}
}

func TestToepenWebPresenter_ReportsAHandEnd(t *testing.T) {
	out := toepenDecode(t, new(ToepenWebPresenter).Output(toepenStub(domain.ToepenPhaseHandEnd, -1, false), nil))
	assert.Equal(t, "toepen.hand_end", out["messageCode"])
}

func TestToepenWebPresenter_SurfacesAnError(t *testing.T) {
	out := toepenDecode(t, new(ToepenWebPresenter).Output(toepenTestGame(t), assert.AnError))
	assert.Equal(t, assert.AnError.Error(), out["message"])
	assert.Empty(t, out["messageCode"])
}

func TestToepenWebPresenter_ShipsTheHintOnAnOrdinaryResponse(t *testing.T) {
	out := toepenDecode(t, new(ToepenWebPresenter).Output(toepenTestGame(t), nil))
	hint, ok := out["hint"].(map[string]any)
	require.True(t, ok, "the hint must ride along with ordinary state")
	assert.NotEmpty(t, hint["reason"])
}

func TestToepenWebPresenter_HintCoversEveryBranch(t *testing.T) {
	p := new(ToepenWebPresenter)

	t.Run("a finished game", func(t *testing.T) {
		out := toepenDecode(t, p.HintOutput(toepenStub(domain.ToepenPhaseGameEnd, 0, true)))
		assert.Equal(t, "toepen.hint.game_end", out["hint"].(map[string]any)["reason"])
	})

	t.Run("a settled hand", func(t *testing.T) {
		out := toepenDecode(t, p.HintOutput(toepenStub(domain.ToepenPhaseHandEnd, -1, false)))
		assert.Equal(t, "toepen.hint.hand_end", out["hint"].(map[string]any)["reason"])
	})

	t.Run("someone else's turn", func(t *testing.T) {
		g := toepenStub(domain.ToepenPhasePlay, -1, false)
		g.ExpectedCalls = nil
		*g = *toepenStub(domain.ToepenPhasePlay, -1, false)
		g.ExpectedCalls = nil
		g2 := toepenStub(domain.ToepenPhasePlay, -1, false)
		g2.ExpectedCalls = nil
		gg := toepenStub(domain.ToepenPhasePlay, -1, false)
		gg.ExpectedCalls = nil
		gg.On("GetPhase").Return(domain.ToepenPhasePlay)
		gg.On("GetGameEndFlag").Return(false)
		gg.On("GetCurrentPlayerIdx").Return(2)
		gg.On("GetWinnerIdx").Return(-1)
		gg.On("GetLeadPlayerIdx").Return(0)
		gg.On("GetDealerIdx").Return(0)
		gg.On("GetCurrentTrick").Return([]*domain.TrickCard{})
		gg.On("GetLeadSuit").Return(-1)
		gg.On("GetTrickNumber").Return(0)
		gg.On("GetStake").Return(1)
		gg.On("GetKnockerIdx").Return(-1)
		gg.On("GetPendingRespondent").Return(-1)
		gg.On("GetLastTrickWinner").Return(-1)
		gg.On("GetHandNumber").Return(1)
		gg.On("GetConfig").Return(domain.DefaultToepenConfig())
		gg.On("GetPlayers").Return([]*domain.ToepenPlayer{domain.NewToepenPlayer(true), domain.NewToepenPlayer(false)})
		gg.On("GetLives", mock.Anything).Return(0)
		gg.On("IsFolded", mock.Anything).Return(false)
		gg.On("IsEliminated", mock.Anything).Return(false)
		gg.On("GetValidPlayIndices", mock.Anything).Return([]int{})

		out := toepenDecode(t, p.HintOutput(gg))
		assert.Equal(t, "toepen.hint.not_your_turn", out["hint"].(map[string]any)["reason"])
	})

	t.Run("no card to suggest", func(t *testing.T) {
		g := toepenStub(domain.ToepenPhasePlay, -1, false)
		g.ExpectedCalls = nil
		g.On("GetPhase").Return(domain.ToepenPhasePlay)
		g.On("GetGameEndFlag").Return(false)
		g.On("GetCurrentPlayerIdx").Return(0)
		g.On("GetWinnerIdx").Return(-1)
		g.On("GetLeadPlayerIdx").Return(0)
		g.On("GetDealerIdx").Return(0)
		g.On("GetCurrentTrick").Return([]*domain.TrickCard{})
		g.On("GetLeadSuit").Return(-1)
		g.On("GetTrickNumber").Return(0)
		g.On("GetStake").Return(1)
		g.On("GetKnockerIdx").Return(-1)
		g.On("GetPendingRespondent").Return(-1)
		g.On("GetLastTrickWinner").Return(-1)
		g.On("GetHandNumber").Return(1)
		g.On("GetConfig").Return(domain.DefaultToepenConfig())
		g.On("GetPlayers").Return([]*domain.ToepenPlayer{domain.NewToepenPlayer(true)})
		g.On("GetLives", mock.Anything).Return(0)
		g.On("IsFolded", mock.Anything).Return(false)
		g.On("IsEliminated", mock.Anything).Return(false)
		g.On("GetValidPlayIndices", mock.Anything).Return([]int{})
		// An empty hand makes the CPU return -1; shipping that index would have
		// the page highlight card "-1".
		g.On("ToepenCpuDecide", 0).Return(domain.ToepenCpuAction{HandIdx: -1})

		out := toepenDecode(t, p.HintOutput(g))
		hint := out["hint"].(map[string]any)
		assert.Equal(t, "toepen.hint.none", hint["reason"])
		assert.NotContains(t, hint, "cardIndex")
	})

	t.Run("an outstanding toep on the human", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			fold bool
			want string
		}{
			{"fold", true, "toepen.hint.fold"},
			{"stay", false, "toepen.hint.stay"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				g := toepenStub(domain.ToepenPhaseRespond, -1, false)
				g.ExpectedCalls = nil
				g.On("GetPhase").Return(domain.ToepenPhaseRespond)
				g.On("GetGameEndFlag").Return(false)
				g.On("GetPendingRespondent").Return(0)
				g.On("GetCurrentPlayerIdx").Return(0)
				g.On("GetWinnerIdx").Return(-1)
				g.On("GetLeadPlayerIdx").Return(0)
				g.On("GetDealerIdx").Return(0)
				g.On("GetCurrentTrick").Return([]*domain.TrickCard{})
				g.On("GetLeadSuit").Return(-1)
				g.On("GetTrickNumber").Return(0)
				g.On("GetStake").Return(2)
				g.On("GetKnockerIdx").Return(1)
				g.On("GetLastTrickWinner").Return(-1)
				g.On("GetHandNumber").Return(1)
				g.On("GetConfig").Return(domain.DefaultToepenConfig())
				g.On("GetPlayers").Return([]*domain.ToepenPlayer{domain.NewToepenPlayer(true)})
				g.On("GetLives", mock.Anything).Return(0)
				g.On("IsFolded", mock.Anything).Return(false)
				g.On("IsEliminated", mock.Anything).Return(false)
				g.On("GetValidPlayIndices", mock.Anything).Return([]int{})
				g.On("ToepenCpuDecide", 0).Return(domain.ToepenCpuAction{HandIdx: -1, Fold: tc.fold})

				out := toepenDecode(t, p.HintOutput(g))
				assert.Equal(t, tc.want, out["hint"].(map[string]any)["reason"])
			})
		}
	})
}

func TestToepenWebPresenter_SkipsANilSeatAndRendersTheLog(t *testing.T) {
	g := toepenStub(domain.ToepenPhasePlay, -1, false)
	g.ExpectedCalls = nil
	g.On("GetPhase").Return(domain.ToepenPhasePlay)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetCurrentPlayerIdx").Return(-1)
	g.On("GetWinnerIdx").Return(-1)
	g.On("GetLeadPlayerIdx").Return(0)
	g.On("GetDealerIdx").Return(0)
	g.On("GetCurrentTrick").Return([]*domain.TrickCard{})
	g.On("GetLeadSuit").Return(-1)
	g.On("GetTrickNumber").Return(0)
	g.On("GetStake").Return(1)
	g.On("GetKnockerIdx").Return(-1)
	g.On("GetPendingRespondent").Return(-1)
	g.On("GetLastTrickWinner").Return(-1)
	g.On("GetHandNumber").Return(1)
	g.On("GetConfig").Return(domain.DefaultToepenConfig())
	g.On("GetPlayers").Return([]*domain.ToepenPlayer{domain.NewToepenPlayer(true), nil})
	g.On("GetLives", mock.Anything).Return(0)
	g.On("IsFolded", mock.Anything).Return(false)
	g.On("IsEliminated", mock.Anything).Return(false)
	g.On("GetValidPlayIndices", mock.Anything).Return([]int{})
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	g.On("ToepenCpuDecide", mock.Anything).Return(domain.ToepenCpuAction{HandIdx: -1})

	out := toepenDecode(t, new(ToepenWebPresenter).Output(g, nil))
	assert.Len(t, out["players"], 1, "the nil seat is dropped, not rendered")
	assert.NotEmpty(t, new(ToepenWebPresenter).ActionLogOutput(g))
}

func TestToepenCuiPresenter_PrintsTheRankingAndHidesCpuHands(t *testing.T) {
	tp := toepenTestGame(t)
	out := new(ToepenCuiPresenter).Output(tp, nil)

	// The ranking is on every screen: it is inverted from every other game here.
	assert.Contains(t, out, "10 > 9 > 8 > 7 > A > K > Q > J")
	// Exactly one indexed hand is printed -- the human's.
	assert.Equal(t, 1, strings.Count(out, "[0]"), "only the human hand may be printed")
}

func TestToepenCuiPresenter_RendersEachPhase(t *testing.T) {
	p := new(ToepenCuiPresenter)
	assert.Contains(t, p.Output(toepenTestGame(t), assert.AnError), assert.AnError.Error())

	tp := toepenTestGame(t)
	require.NoError(t, tp.Toep(0))
	assert.NotEmpty(t, p.Output(tp, nil), "a response phase renders")

	assert.NotEmpty(t, p.Output(toepenStub(domain.ToepenPhaseHandEnd, -1, false), nil))
	assert.NotEmpty(t, p.Output(toepenStub(domain.ToepenPhaseGameEnd, 0, true), nil))
	assert.NotEmpty(t, p.Output(toepenStub(domain.ToepenPhaseGameEnd, 1, true), nil))
	assert.NotEmpty(t, p.ActionLogOutput(toepenTestGame(t)))
}

func TestToepenCuiPresenter_HintResolvesItsReasonKey(t *testing.T) {
	for range 50 {
		out := new(ToepenCuiPresenter).HintOutput(toepenTestGame(t))
		assert.NotContains(t, out, "toepen.hint.", "the reason must be translated, not printed raw")
		assert.NotEmpty(t, strings.TrimSpace(out))
	}
}

func TestToepenCuiPresenter_EveryReasonTheHintCanReturnIsMapped(t *testing.T) {
	// The fallback for an unmapped reason is unreachable TODAY, and this is
	// what makes that true. Adding a reason on the Web side without an entry
	// turns this red rather than printing `toepen.hint.x` at a player.
	for _, reason := range []string{
		"toepen.hint.game_end", "toepen.hint.hand_end", "toepen.hint.not_your_turn",
		"toepen.hint.stay", "toepen.hint.fold", "toepen.hint.play", "toepen.hint.none",
	} {
		assert.NotEmpty(t, toepenHintReasonKeys[reason], "reason %q has no i18n key", reason)
	}
	assert.Len(t, toepenHintReasonKeys, 7, "a new reason needs an entry here too")
}

func TestToepenCuiPresenter_HintPrintsAnIndexOnlyWhenThereIsOne(t *testing.T) {
	// The CUI hint has two shapes -- a hand index, and a bare reason -- and
	// each formats differently. A missing branch prints the wrong line.
	p := new(ToepenCuiPresenter)

	// Bare reason: the game is over, so there is nothing to point at.
	bare := p.HintOutput(toepenStub(domain.ToepenPhaseGameEnd, 0, true))
	assert.NotContains(t, bare, "手札[")
	assert.NotContains(t, bare, "toepen.hint.")

	// With an index: a live hand on the HUMAN's turn. A fresh deal starts on
	// the seat after the dealer, so play round to seat 0 first rather than
	// assuming the human leads.
	tp := toepenTestGame(t)
	for range 10 {
		if tp.GetCurrentPlayerIdx() == 0 {
			break
		}
		idx := tp.GetCurrentPlayerIdx()
		require.NoError(t, tp.PlayCard(idx, tp.GetValidPlayIndices(idx)[0]))
	}
	require.Equal(t, 0, tp.GetCurrentPlayerIdx())
	assert.Contains(t, p.HintOutput(tp), "手札[")
}
