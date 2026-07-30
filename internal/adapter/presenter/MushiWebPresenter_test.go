//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func mushiTestGame(t *testing.T) *domain.Mushi {
	t.Helper()
	m := domain.NewDefaultMushi()
	m.Reset()
	return m
}

func mushiDecode(t *testing.T, raw string) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return out
}

func TestMushiWebPresenter_HidesTheCpuHandButNeverItsCaptures(t *testing.T) {
	// Captured cards are PUBLIC in hanafuda -- masking them would remove the
	// read the game is built on. Only the hand is withheld.
	m := mushiTestGame(t)
	out := mushiDecode(t, new(MushiWebPresenter).Output(m, nil))

	players := out["players"].([]any)
	require.Len(t, players, 2)

	human := players[0].(map[string]any)
	assert.Equal(t, false, human["hidden"])
	assert.Len(t, human["cards"], domain.MushiHandSize)

	cpu := players[1].(map[string]any)
	assert.Equal(t, true, cpu["hidden"])
	assert.Empty(t, cpu["cards"], "the CPU's hand must not reach the client")
	assert.Equal(t, float64(domain.MushiHandSize), cpu["cardCount"])
	assert.NotNil(t, cpu["captured"], "captures are public and must always be sent")
}

func TestMushiWebPresenter_CardsCarryTheirIdentityAndTheHanafudaDeckFlag(t *testing.T) {
	// design serialises to a SUIT NAME, so the month has to travel separately
	// and the deck flag has to switch the client to the procedural path
	// (ADR-0033). Without either, the card renders as a French-suit PNG.
	m := mushiTestGame(t)
	out := mushiDecode(t, new(MushiWebPresenter).Output(m, nil))

	field := out["field"].([]any)
	require.NotEmpty(t, field)
	card := field[0].(map[string]any)

	assert.Equal(t, "hanafuda", card["deck"])
	assert.NotEmpty(t, card["glyph"])
	month := card["month"].(float64)
	assert.GreaterOrEqual(t, month, float64(1))
	assert.LessOrEqual(t, month, float64(12))
	assert.NotEqual(t, float64(6), month, "June is not in the Mushi pack")
	assert.NotEqual(t, float64(7), month, "July is not in the Mushi pack")
	assert.Contains(t, card, "points")
	assert.Contains(t, card, "isWild")
}

func TestMushiWebPresenter_ShipsSelectableIndicesSoTheClientNeedNotDeriveThem(t *testing.T) {
	// Drive until a choice is open; the wild's "not another willow" rule lives
	// here and must reach the client as a list, not as a rule to re-implement.
	for range 300 {
		m := mushiTestGame(t)
		for range 40 {
			phase := m.GetPhase()
			if phase == domain.MushiPhaseSelect || phase == domain.MushiPhaseWildSelect {
				out := mushiDecode(t, new(MushiWebPresenter).Output(m, nil))
				assert.NotEmpty(t, out["selectableIndices"])
				assert.NotNil(t, out["pendingCard"])
				if phase == domain.MushiPhaseWildSelect {
					for _, idx := range out["selectableIndices"].([]any) {
						c := m.GetField()[int(idx.(float64))]
						assert.NotEqual(t, domain.MushiWildMonth, c.GetDesign(),
							"the wild may never take a willow card")
					}
				}
				return
			}
			if phase != domain.MushiPhasePlay {
				break
			}
			idx := m.GetCurrentPlayerIdx()
			if m.GetPlayer(idx).GetCardsSize() == 0 || m.PlayCard(idx, 0) != nil {
				break
			}
		}
	}
	t.Skip("no selection phase occurred")
}

func TestMushiWebPresenter_ReportsTheOutcome(t *testing.T) {
	m := domain.NewDefaultMushi()
	cfg := m.GetConfig()
	cfg.TargetRounds = 1
	m.SetConfig(cfg)
	m.Reset()
	for range 400 {
		if m.GetGameEndFlag() {
			break
		}
		idx := m.GetCurrentPlayerIdx()
		action := m.MushiCpuDecide(idx)
		phase := m.GetPhase()
		if phase == domain.MushiPhaseSelect || phase == domain.MushiPhaseWildSelect {
			require.NoError(t, m.SelectCapture(idx, action.FieldIdx))
			continue
		}
		if phase != domain.MushiPhasePlay {
			break
		}
		require.NoError(t, m.PlayCard(idx, action.HandIdx))
	}
	require.True(t, m.GetGameEndFlag())

	out := mushiDecode(t, new(MushiWebPresenter).Output(m, nil))
	assert.Contains(t, []any{"mushi.win", "mushi.lose", "mushi.draw"}, out["messageCode"])
	// Everything is revealed once the game is over.
	cpu := out["players"].([]any)[1].(map[string]any)
	assert.Equal(t, false, cpu["hidden"])
}

func TestMushiWebPresenter_ReportsARoundEnd(t *testing.T) {
	m := mushiTestGame(t)
	for range 400 {
		if m.GetPhase() == domain.MushiPhaseRoundEnd {
			break
		}
		idx := m.GetCurrentPlayerIdx()
		action := m.MushiCpuDecide(idx)
		phase := m.GetPhase()
		if phase == domain.MushiPhaseSelect || phase == domain.MushiPhaseWildSelect {
			require.NoError(t, m.SelectCapture(idx, action.FieldIdx))
			continue
		}
		require.NoError(t, m.PlayCard(idx, action.HandIdx))
	}
	require.Equal(t, domain.MushiPhaseRoundEnd, m.GetPhase())

	out := mushiDecode(t, new(MushiWebPresenter).Output(m, nil))
	assert.Equal(t, "mushi.round_end", out["messageCode"])
}

func TestMushiWebPresenter_SurfacesAnError(t *testing.T) {
	out := mushiDecode(t, new(MushiWebPresenter).Output(mushiTestGame(t), assert.AnError))
	assert.Equal(t, assert.AnError.Error(), out["message"])
	assert.Empty(t, out["messageCode"])
}

func TestMushiWebPresenter_ShipsTheHintOnAnOrdinaryResponse(t *testing.T) {
	// Every other game sets Hint only in HintOutput, which no page calls.
	out := mushiDecode(t, new(MushiWebPresenter).Output(mushiTestGame(t), nil))
	hint, ok := out["hint"].(map[string]any)
	require.True(t, ok, "the hint must ride along with ordinary state")
	assert.Equal(t, "mushi.hint.play", hint["reason"])
	assert.Contains(t, hint, "cardIndex")
}

func TestMushiWebPresenter_HintDeclinesWhenItIsNotTheHumansTurn(t *testing.T) {
	m := mushiTestGame(t)
	// Play once so the turn passes to the CPU.
	require.NoError(t, m.PlayCard(m.GetCurrentPlayerIdx(), 0))
	if m.GetCurrentPlayerIdx() == 0 {
		t.Skip("the human retained the turn on this deal")
	}
	out := mushiDecode(t, new(MushiWebPresenter).HintOutput(m))
	assert.Equal(t, "mushi.hint.not_your_turn", out["hint"].(map[string]any)["reason"])
	// ...and the ordinary response carries no hint at all in that state.
	assert.Nil(t, mushiDecode(t, new(MushiWebPresenter).Output(m, nil))["hint"])
}

func TestMushiWebPresenter_ActionLogRenders(t *testing.T) {
	assert.NotEmpty(t, new(MushiWebPresenter).ActionLogOutput(mushiTestGame(t)))
}

func TestMushiFace_ColoursByCategoryAndFlagsTheWild(t *testing.T) {
	assert.Nil(t, mushiFace(nil))
	assert.Equal(t, "gold", mushiFace(domain.NewCard(1, 1, true)).Color, "bright")
	assert.Equal(t, "purple", mushiFace(domain.NewCard(2, 1, true)).Color, "animal")
	assert.Equal(t, "red", mushiFace(domain.NewCard(1, 2, true)).Color, "ribbon")
	assert.Equal(t, "black", mushiFace(domain.NewCard(1, 3, true)).Color, "chaff")
	// The lightning card is chaff but is picked out, because its role is not.
	assert.Equal(t, "gold", mushiFace(domain.NewCard(11, 4, true)).Color)
	assert.Nil(t, mushiCardOutput(nil))
}

func TestMushiWebPresenter_HintCoversEveryTerminalState(t *testing.T) {
	t.Run("a finished game", func(t *testing.T) {
		m := domain.NewDefaultMushi()
		cfg := m.GetConfig()
		cfg.TargetRounds = 1
		m.SetConfig(cfg)
		m.Reset()
		for range 400 {
			if m.GetGameEndFlag() {
				break
			}
			idx := m.GetCurrentPlayerIdx()
			action := m.MushiCpuDecide(idx)
			phase := m.GetPhase()
			if phase == domain.MushiPhaseSelect || phase == domain.MushiPhaseWildSelect {
				require.NoError(t, m.SelectCapture(idx, action.FieldIdx))
				continue
			}
			if phase != domain.MushiPhasePlay {
				break
			}
			require.NoError(t, m.PlayCard(idx, action.HandIdx))
		}
		out := mushiDecode(t, new(MushiWebPresenter).HintOutput(m))
		assert.Equal(t, "mushi.hint.game_end", out["hint"].(map[string]any)["reason"])
	})

	t.Run("a settled round", func(t *testing.T) {
		m := mushiTestGame(t)
		for range 400 {
			if m.GetPhase() == domain.MushiPhaseRoundEnd {
				break
			}
			idx := m.GetCurrentPlayerIdx()
			action := m.MushiCpuDecide(idx)
			phase := m.GetPhase()
			if phase == domain.MushiPhaseSelect || phase == domain.MushiPhaseWildSelect {
				require.NoError(t, m.SelectCapture(idx, action.FieldIdx))
				continue
			}
			require.NoError(t, m.PlayCard(idx, action.HandIdx))
		}
		out := mushiDecode(t, new(MushiWebPresenter).HintOutput(m))
		assert.Equal(t, "mushi.hint.round_end", out["hint"].(map[string]any)["reason"])
	})

	t.Run("a pending selection on the human's turn", func(t *testing.T) {
		for range 300 {
			m := mushiTestGame(t)
			for range 40 {
				phase := m.GetPhase()
				if (phase == domain.MushiPhaseSelect || phase == domain.MushiPhaseWildSelect) &&
					m.GetCurrentPlayerIdx() == 0 {
					out := mushiDecode(t, new(MushiWebPresenter).HintOutput(m))
					hint := out["hint"].(map[string]any)
					assert.Equal(t, "mushi.hint.select", hint["reason"])
					assert.Contains(t, hint, "fieldIndex")
					return
				}
				if phase != domain.MushiPhasePlay {
					break
				}
				idx := m.GetCurrentPlayerIdx()
				if m.GetPlayer(idx).GetCardsSize() == 0 || m.PlayCard(idx, 0) != nil {
					break
				}
			}
		}
		t.Skip("no human selection phase occurred")
	})
}

// mushiStubGame wires a MockMushiGame with every accessor the presenters touch,
// so a test can pin an exact phase/winner instead of shuffling until one turns
// up. The outcome branches are reachable in play but not reachable ON DEMAND
// from a real game, which is how they stayed uncovered.
func mushiStubGame(phase domain.MushiPhase, winner int, gameEnd bool) *interfaces.MockMushiGame {
	g := new(interfaces.MockMushiGame)
	g.On("GetPhase").Return(phase)
	g.On("GetWinnerIdx").Return(winner)
	g.On("GetGameEndFlag").Return(gameEnd)
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("GetDealerIdx").Return(0)
	g.On("GetRoundNumber").Return(1)
	g.On("GetStockCount").Return(0)
	g.On("GetField").Return([]*domain.Card{})
	g.On("GetPendingCard").Return((*domain.Card)(nil))
	g.On("GetSelectableIndices").Return([]int{})
	g.On("GetConfig").Return(domain.DefaultMushiConfig())
	g.On("GetPlayers").Return([]*domain.MushiPlayer{domain.NewMushiPlayer(true), domain.NewMushiPlayer(false)})
	g.On("GetCaptured", mock.Anything).Return([]*domain.Card{})
	g.On("GetScore", mock.Anything).Return(0)
	g.On("GetRoundResult", mock.Anything).Return(0)
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	g.On("MushiCpuDecide", mock.Anything).Return(domain.MushiCpuAction{HandIdx: 0, FieldIdx: 0})
	return g
}

func TestMushiWebPresenter_EveryOutcomeHasItsOwnCode(t *testing.T) {
	for _, tc := range []struct {
		name   string
		winner int
		want   string
	}{
		{"human wins", 0, "mushi.win"},
		{"draw", -1, "mushi.draw"},
		{"cpu wins", 1, "mushi.lose"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := mushiDecode(t, new(MushiWebPresenter).Output(mushiStubGame(domain.MushiPhaseGameEnd, tc.winner, true), nil))
			assert.Equal(t, tc.want, out["messageCode"])
		})
	}
}

func TestMushiWebPresenter_HintDeclinesOnTheOpponentsSelectionTurn(t *testing.T) {
	// A selection phase belonging to the CPU: the human has nothing to choose,
	// and suggesting a field card would invite a click that cannot land.
	g := new(interfaces.MockMushiGame)
	g.On("GetPhase").Return(domain.MushiPhaseSelect)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("GetWinnerIdx").Return(-1)
	g.On("GetDealerIdx").Return(0)
	g.On("GetRoundNumber").Return(1)
	g.On("GetStockCount").Return(0)
	g.On("GetField").Return([]*domain.Card{})
	g.On("GetPendingCard").Return((*domain.Card)(nil))
	g.On("GetSelectableIndices").Return([]int{})
	g.On("GetConfig").Return(domain.DefaultMushiConfig())
	g.On("GetPlayers").Return([]*domain.MushiPlayer{domain.NewMushiPlayer(true), domain.NewMushiPlayer(false)})
	g.On("GetCaptured", mock.Anything).Return([]*domain.Card{})
	g.On("GetScore", mock.Anything).Return(0)
	g.On("GetRoundResult", mock.Anything).Return(0)

	out := mushiDecode(t, new(MushiWebPresenter).HintOutput(g))
	assert.Equal(t, "mushi.hint.not_your_turn", out["hint"].(map[string]any)["reason"])
}

func TestMushiWebPresenter_HintReportsNoneWhenTheCpuHasNothingToOffer(t *testing.T) {
	// An empty hand makes MushiCpuDecide return -1; the hint must say so rather
	// than ship an index of -1 the page would happily highlight.
	for _, tc := range []struct {
		name   string
		phase  domain.MushiPhase
		action domain.MushiCpuAction
	}{
		{"nothing to play", domain.MushiPhasePlay, domain.MushiCpuAction{HandIdx: -1, FieldIdx: -1}},
		{"nothing to select", domain.MushiPhaseSelect, domain.MushiCpuAction{HandIdx: -1, FieldIdx: -1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := mushiStubGame(tc.phase, -1, false)
			g.ExpectedCalls = nil
			g.On("GetPhase").Return(tc.phase)
			g.On("GetGameEndFlag").Return(false)
			g.On("GetCurrentPlayerIdx").Return(0)
			g.On("GetWinnerIdx").Return(-1)
			g.On("GetDealerIdx").Return(0)
			g.On("GetRoundNumber").Return(1)
			g.On("GetStockCount").Return(0)
			g.On("GetField").Return([]*domain.Card{})
			g.On("GetPendingCard").Return((*domain.Card)(nil))
			g.On("GetSelectableIndices").Return([]int{})
			g.On("GetConfig").Return(domain.DefaultMushiConfig())
			g.On("GetPlayers").Return([]*domain.MushiPlayer{domain.NewMushiPlayer(true), domain.NewMushiPlayer(false)})
			g.On("GetCaptured", mock.Anything).Return([]*domain.Card{})
			g.On("GetScore", mock.Anything).Return(0)
			g.On("GetRoundResult", mock.Anything).Return(0)
			g.On("MushiCpuDecide", 0).Return(tc.action)

			out := mushiDecode(t, new(MushiWebPresenter).HintOutput(g))
			hint := out["hint"].(map[string]any)
			assert.Equal(t, "mushi.hint.none", hint["reason"])
			assert.NotContains(t, hint, "cardIndex")
			assert.NotContains(t, hint, "fieldIndex")
		})
	}
}

func TestMushiWebPresenter_SkipsANilSeat(t *testing.T) {
	// A corrupted snapshot could leave a nil player; the presenter must drop it
	// rather than dereference it on the way to the browser.
	g := mushiStubGame(domain.MushiPhasePlay, -1, false)
	g.ExpectedCalls = nil
	g.On("GetPhase").Return(domain.MushiPhasePlay)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetCurrentPlayerIdx").Return(-1)
	g.On("GetWinnerIdx").Return(-1)
	g.On("GetDealerIdx").Return(0)
	g.On("GetRoundNumber").Return(1)
	g.On("GetStockCount").Return(0)
	g.On("GetField").Return([]*domain.Card{})
	g.On("GetPendingCard").Return((*domain.Card)(nil))
	g.On("GetSelectableIndices").Return([]int{})
	g.On("GetConfig").Return(domain.DefaultMushiConfig())
	g.On("GetPlayers").Return([]*domain.MushiPlayer{domain.NewMushiPlayer(true), nil})
	g.On("GetCaptured", mock.Anything).Return([]*domain.Card{})
	g.On("GetScore", mock.Anything).Return(0)
	g.On("GetRoundResult", mock.Anything).Return(0)

	out := mushiDecode(t, new(MushiWebPresenter).Output(g, nil))
	assert.Len(t, out["players"], 1, "the nil seat is dropped, not rendered")
}

func TestMushiWebPresenter_HintNamesTheFieldCardOnTheHumansSelectionTurn(t *testing.T) {
	g := mushiStubGame(domain.MushiPhaseSelect, -1, false)
	g.ExpectedCalls = nil
	g.On("GetPhase").Return(domain.MushiPhaseSelect)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("GetWinnerIdx").Return(-1)
	g.On("GetDealerIdx").Return(0)
	g.On("GetRoundNumber").Return(1)
	g.On("GetStockCount").Return(0)
	g.On("GetField").Return([]*domain.Card{})
	g.On("GetPendingCard").Return((*domain.Card)(nil))
	g.On("GetSelectableIndices").Return([]int{2})
	g.On("GetConfig").Return(domain.DefaultMushiConfig())
	g.On("GetPlayers").Return([]*domain.MushiPlayer{domain.NewMushiPlayer(true), domain.NewMushiPlayer(false)})
	g.On("GetCaptured", mock.Anything).Return([]*domain.Card{})
	g.On("GetScore", mock.Anything).Return(0)
	g.On("GetRoundResult", mock.Anything).Return(0)
	g.On("MushiCpuDecide", 0).Return(domain.MushiCpuAction{HandIdx: -1, FieldIdx: 2})

	out := mushiDecode(t, new(MushiWebPresenter).HintOutput(g))
	hint := out["hint"].(map[string]any)
	assert.Equal(t, "mushi.hint.select", hint["reason"])
	assert.Equal(t, float64(2), hint["fieldIndex"])
}
