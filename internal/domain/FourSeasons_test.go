//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestFourSeasons() *FourSeasons {
	f := NewFourSeasons(NewTrumpCards(0))
	f.Reset()
	return f
}

// setFourSeasonsBoard installs an exact board. Reset shuffles, so nothing may be
// asserted on top of it.
func setFourSeasonsBoard(f *FourSeasons, base int, tableau [FourSeasonsTableauCnt][]*Card,
	foundation [FourSeasonsFoundationCnt][]*Card, waste, stock []*Card,
) {
	f.baseRank = base
	f.tableau = tableau
	f.foundation = foundation
	f.waste = waste
	f.stock = stock
	f.phase = FourSeasonsPhasePlaying
	f.moveCount = 0
	f.actionLog = nil
	f.history = nil
}

func TestFourSeasons_Reset_InitialState(t *testing.T) {
	f := newTestFourSeasons()

	assert.Equal(t, FourSeasonsPhasePlaying, f.GetPhase())
	assert.Equal(t, 0, f.GetMoveCount())
	assert.Empty(t, f.GetActionLog())
	assert.False(t, f.CanUndo())

	// The cross: five piles, one card each.
	for i, col := range f.GetTableau() {
		assert.Len(t, col, 1, "cross pile %d should open with one card", i)
	}
	// One foundation is opened by the base card; the other three start empty.
	opened := 0
	for _, pile := range f.GetFoundations() {
		if len(pile) > 0 {
			opened++
			assert.Len(t, pile, 1)
			assert.Equal(t, f.GetBaseRank(), pile[0].GetValue(), "the base card sets the base rank")
		}
	}
	assert.Equal(t, 1, opened, "exactly one foundation is seeded")
	// 52 - 5 cross - 1 base = 46 in the stock, nothing in the waste yet.
	assert.Equal(t, 46, f.GetStockCount())
	assert.Empty(t, f.GetWaste())
}

// The base rank is whatever the first card happened to be — not always an Ace.
// That is the rule this game shares with Canfield and that Klondike does not have.
func TestFourSeasons_Reset_BaseRankIsWhateverWasDealt(t *testing.T) {
	seen := map[int]bool{}
	for range 40 {
		f := newTestFourSeasons()
		require.GreaterOrEqual(t, f.GetBaseRank(), 1)
		require.LessOrEqual(t, f.GetBaseRank(), CardValueMax)
		seen[f.GetBaseRank()] = true
	}
	assert.Greater(t, len(seen), 1, "the base rank must vary between deals, not be pinned to Ace")
}

func TestFourSeasons_Draw(t *testing.T) {
	f := newTestFourSeasons()
	before := f.GetStockCount()

	require.NoError(t, f.Draw())

	assert.Equal(t, before-1, f.GetStockCount())
	assert.Len(t, f.GetWaste(), 1)
	assert.Equal(t, 1, f.GetMoveCount())
	assert.True(t, f.CanUndo())
}

func TestFourSeasons_Draw_EmptyStock(t *testing.T) {
	f := newTestFourSeasons()
	setFourSeasonsBoard(f, 5, [FourSeasonsTableauCnt][]*Card{}, [FourSeasonsFoundationCnt][]*Card{}, nil, nil)
	assert.Error(t, f.Draw())
}

func TestFourSeasons_Draw_RejectedWhenNotPlaying(t *testing.T) {
	f := newTestFourSeasons()
	f.GiveUp()
	assert.Error(t, f.Draw())
}

// Foundations build UP in suit and wrap: with a base of Q, the order is
// Q -> K -> A -> 2 ... A following K is the rule Klondike does not have.
func TestFourSeasons_Foundation_AscendsInSuitAndWraps(t *testing.T) {
	f := newTestFourSeasons()
	var fnd [FourSeasonsFoundationCnt][]*Card
	fnd[0] = []*Card{sp(12)} // base = Q of spades
	var tab [FourSeasonsTableauCnt][]*Card
	setFourSeasonsBoard(f, 12, tab, fnd, []*Card{sp(13)}, nil)

	require.NoError(t, f.MoveWasteToFoundation(0), "K follows Q")
	f.waste = []*Card{sp(1)}
	require.NoError(t, f.MoveWasteToFoundation(0), "A follows K — the wraparound")
	f.waste = []*Card{sp(2)}
	require.NoError(t, f.MoveWasteToFoundation(0), "2 follows A")

	assert.Len(t, f.GetFoundations()[0], 4)
}

func TestFourSeasons_Foundation_Rejections(t *testing.T) {
	tests := []struct {
		name string
		card *Card
		fIdx int
	}{
		{"wrong suit", he(13), 0},
		{"wrong rank", sp(5), 0},
		{"foundation index below range", sp(13), -1},
		{"foundation index above range", sp(13), FourSeasonsFoundationCnt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTestFourSeasons()
			var fnd [FourSeasonsFoundationCnt][]*Card
			fnd[0] = []*Card{sp(12)}
			setFourSeasonsBoard(f, 12, [FourSeasonsTableauCnt][]*Card{}, fnd, []*Card{tt.card}, nil)
			assert.Error(t, f.MoveWasteToFoundation(tt.fIdx))
			assert.Equal(t, 0, f.GetMoveCount())
		})
	}
}

// An empty foundation only opens on the base rank — any suit, because which
// foundation a suit lands on is decided by whichever opens first.
func TestFourSeasons_Foundation_EmptyOpensOnBaseRankOnly(t *testing.T) {
	f := newTestFourSeasons()
	var fnd [FourSeasonsFoundationCnt][]*Card
	setFourSeasonsBoard(f, 7, [FourSeasonsTableauCnt][]*Card{}, fnd, []*Card{he(7)}, nil)
	require.NoError(t, f.MoveWasteToFoundation(1))
	assert.Len(t, f.GetFoundations()[1], 1)

	f2 := newTestFourSeasons()
	setFourSeasonsBoard(f2, 7, [FourSeasonsTableauCnt][]*Card{}, [FourSeasonsFoundationCnt][]*Card{}, []*Card{he(8)}, nil)
	assert.Error(t, f2.MoveWasteToFoundation(1), "an 8 cannot open a base-7 foundation")
}

// The cross builds DOWN, wraps, and IGNORES suit — all three differ from
// Canfield, which alternates colour and does not wrap on the tableau.
func TestFourSeasons_Tableau_DescendsIgnoringSuitAndWraps(t *testing.T) {
	f := newTestFourSeasons()
	var tab [FourSeasonsTableauCnt][]*Card
	tab[0] = []*Card{sp(2)}
	setFourSeasonsBoard(f, 5, tab, [FourSeasonsFoundationCnt][]*Card{}, []*Card{he(1)}, nil)

	require.NoError(t, f.MoveWasteToTableau(0), "A goes on 2, suit ignored")
	f.waste = []*Card{cl(13)}
	require.NoError(t, f.MoveWasteToTableau(0), "K goes on A — the wraparound")
	assert.Len(t, f.GetTableau()[0], 3)
}

func TestFourSeasons_Tableau_RejectsWrongRank(t *testing.T) {
	f := newTestFourSeasons()
	var tab [FourSeasonsTableauCnt][]*Card
	tab[0] = []*Card{sp(9)}
	setFourSeasonsBoard(f, 5, tab, [FourSeasonsFoundationCnt][]*Card{}, []*Card{he(7)}, nil)
	assert.Error(t, f.MoveWasteToTableau(0), "7 does not follow 9")
}

// An empty cross space takes ANY card — it is not reserved for a rank.
func TestFourSeasons_Tableau_EmptySpaceTakesAnyCard(t *testing.T) {
	f := newTestFourSeasons()
	var tab [FourSeasonsTableauCnt][]*Card
	setFourSeasonsBoard(f, 5, tab, [FourSeasonsFoundationCnt][]*Card{}, []*Card{he(7)}, nil)
	require.NoError(t, f.MoveWasteToTableau(3))
	assert.Len(t, f.GetTableau()[3], 1)
}

// Only the top card of a cross pile moves — the pile is not a sequence you can
// shift as a unit.
func TestFourSeasons_TableauToTableau_MovesOnlyTheTopCard(t *testing.T) {
	f := newTestFourSeasons()
	var tab [FourSeasonsTableauCnt][]*Card
	tab[0] = []*Card{sp(9), he(8)} // top is the 8
	tab[1] = []*Card{cl(9)}
	setFourSeasonsBoard(f, 5, tab, [FourSeasonsFoundationCnt][]*Card{}, nil, nil)

	require.NoError(t, f.MoveTableauToTableau(0, 1))
	assert.Len(t, f.GetTableau()[0], 1, "the buried 9 stays put")
	assert.Equal(t, 9, f.GetTableau()[0][0].GetValue())
	assert.Len(t, f.GetTableau()[1], 2)
}

func TestFourSeasons_TableauToTableau_Rejections(t *testing.T) {
	f := newTestFourSeasons()
	var tab [FourSeasonsTableauCnt][]*Card
	tab[0] = []*Card{sp(9)}
	setFourSeasonsBoard(f, 5, tab, [FourSeasonsFoundationCnt][]*Card{}, nil, nil)

	assert.Error(t, f.MoveTableauToTableau(0, 0), "a pile cannot move onto itself")
	assert.Error(t, f.MoveTableauToTableau(-1, 1))
	assert.Error(t, f.MoveTableauToTableau(0, FourSeasonsTableauCnt))
	assert.Error(t, f.MoveTableauToTableau(2, 1), "an empty source has nothing to move")
}

func TestFourSeasons_TableauToFoundation(t *testing.T) {
	f := newTestFourSeasons()
	var tab [FourSeasonsTableauCnt][]*Card
	tab[2] = []*Card{sp(13)}
	var fnd [FourSeasonsFoundationCnt][]*Card
	fnd[0] = []*Card{sp(12)}
	setFourSeasonsBoard(f, 12, tab, fnd, nil, nil)

	require.NoError(t, f.MoveTableauToFoundation(2, 0))
	assert.Empty(t, f.GetTableau()[2])
	assert.Len(t, f.GetFoundations()[0], 2)
}

func TestFourSeasons_GameClear(t *testing.T) {
	f := newTestFourSeasons()
	var fnd [FourSeasonsFoundationCnt][]*Card
	for i := range fnd {
		fnd[i] = make([]*Card, 0, CardValueMax)
		for v := 1; v < CardValueMax; v++ {
			fnd[i] = append(fnd[i], NewCard(i+1, v, true))
		}
	}
	var tab [FourSeasonsTableauCnt][]*Card
	tab[0] = []*Card{NewCard(1, CardValueMax, true)}
	setFourSeasonsBoard(f, 1, tab, fnd, nil, nil)
	// Three foundations are already 12 long; complete the fourth's last card.
	for i := 1; i < FourSeasonsFoundationCnt; i++ {
		f.foundation[i] = append(f.foundation[i], NewCard(i+1, CardValueMax, true))
	}

	require.NoError(t, f.MoveTableauToFoundation(0, 0))
	assert.Equal(t, FourSeasonsPhaseGameClear, f.GetPhase())
	assert.True(t, f.GetGameEndFlag())
}

func TestFourSeasons_GiveUp(t *testing.T) {
	f := newTestFourSeasons()
	f.GiveUp()
	assert.Equal(t, FourSeasonsPhaseGameOver, f.GetPhase())
	assert.True(t, f.GetGameEndFlag())
	assert.Len(t, f.GetActionLog(), 1)
	f.GiveUp()
	assert.Len(t, f.GetActionLog(), 1, "giving up twice logs once")
}

func TestFourSeasons_GetHint(t *testing.T) {
	f := newTestFourSeasons()
	var tab [FourSeasonsTableauCnt][]*Card
	tab[1] = []*Card{sp(13)}
	var fnd [FourSeasonsFoundationCnt][]*Card
	fnd[0] = []*Card{sp(12)}
	setFourSeasonsBoard(f, 12, tab, fnd, nil, nil)

	h := f.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "tableau", h.FromZone)
	assert.Equal(t, 1, h.FromIdx)
	assert.Equal(t, "foundation", h.ToZone)

	f2 := newTestFourSeasons()
	setFourSeasonsBoard(f2, 12, [FourSeasonsTableauCnt][]*Card{}, [FourSeasonsFoundationCnt][]*Card{}, nil, nil)
	assert.Nil(t, f2.GetHint(), "nothing to move")
}

func TestFourSeasons_GetHint_NilWhenNotPlaying(t *testing.T) {
	f := newTestFourSeasons()
	f.GiveUp()
	assert.Nil(t, f.GetHint())
}

func TestFourSeasons_AutoComplete(t *testing.T) {
	f := newTestFourSeasons()
	var tab [FourSeasonsTableauCnt][]*Card
	tab[0] = []*Card{sp(1), sp(13)} // K on top, A beneath — both go once the K lands
	var fnd [FourSeasonsFoundationCnt][]*Card
	fnd[0] = []*Card{sp(12)}
	setFourSeasonsBoard(f, 12, tab, fnd, nil, nil)

	require.NoError(t, f.AutoComplete())
	assert.Empty(t, f.GetTableau()[0])
	assert.Len(t, f.GetFoundations()[0], 3, "Q-K-A")
}

func TestFourSeasons_AutoComplete_NoMove(t *testing.T) {
	f := newTestFourSeasons()
	setFourSeasonsBoard(f, 12, [FourSeasonsTableauCnt][]*Card{}, [FourSeasonsFoundationCnt][]*Card{}, nil, nil)
	assert.Error(t, f.AutoComplete())
}

func TestFourSeasons_AutoComplete_RejectedWhenNotPlaying(t *testing.T) {
	f := newTestFourSeasons()
	f.GiveUp()
	assert.Error(t, f.AutoComplete())
}

func TestFourSeasons_Undo(t *testing.T) {
	f := newTestFourSeasons()
	assert.Error(t, f.Undo(), "nothing to undo yet")

	require.NoError(t, f.Draw())
	require.NoError(t, f.Undo())
	assert.Empty(t, f.GetWaste())
	assert.Equal(t, 0, f.GetMoveCount())
	assert.False(t, f.CanUndo())
}

func TestFourSeasons_UndoN(t *testing.T) {
	f := newTestFourSeasons()
	require.NoError(t, f.Draw())
	require.NoError(t, f.Draw())

	assert.Error(t, f.UndoN(0))
	assert.Error(t, f.UndoN(3))
	require.NoError(t, f.UndoN(2))
	assert.Empty(t, f.GetWaste())
}

func TestFourSeasons_JSONRoundTrip(t *testing.T) {
	f := newTestFourSeasons()
	require.NoError(t, f.Draw())

	data, err := json.Marshal(f)
	require.NoError(t, err)

	restored := NewFourSeasons(NewTrumpCards(0))
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, f.GetBaseRank(), restored.GetBaseRank(), "the base rank must survive — every placement rule depends on it")
	assert.Equal(t, f.GetTableau(), restored.GetTableau())
	assert.Equal(t, f.GetFoundations(), restored.GetFoundations())
	assert.Equal(t, f.GetStockCount(), restored.GetStockCount())
	assert.Equal(t, f.GetMoveCount(), restored.GetMoveCount())

	// The Worker rebuilds from KV every request, so an unpersisted undo stack
	// means Undo silently never works (#4478).
	require.True(t, restored.CanUndo())
	require.NoError(t, restored.Undo())
	assert.Empty(t, restored.GetWaste())
}

func TestFourSeasons_UnmarshalJSON_Rejections(t *testing.T) {
	tests := []struct{ name, data string }{
		{"malformed", `{`},
		{"phase below range", `{"ps":-1}`},
		{"phase above range", `{"ps":99}`},
		{"negative move count", `{"ps":0,"mc":-1}`},
		{"base rank out of range", `{"ps":0,"mc":0,"br":14}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFourSeasons(NewTrumpCards(0))
			assert.Error(t, json.Unmarshal([]byte(tt.data), f))
		})
	}
}

func TestFourSeasons_NewDefault(t *testing.T) {
	f := NewDefaultFourSeasons()
	require.NotNil(t, f)
	f.Reset()
	assert.Equal(t, 46, f.GetStockCount())
}

// Every move must refuse once the game has ended. These are the branches that,
// if missed, let a finished board keep accepting cards.
func TestFourSeasons_AllMovesRejectedWhenNotPlaying(t *testing.T) {
	moves := map[string]func(*FourSeasons) error{
		"waste->foundation":   func(f *FourSeasons) error { return f.MoveWasteToFoundation(0) },
		"waste->tableau":      func(f *FourSeasons) error { return f.MoveWasteToTableau(0) },
		"tableau->foundation": func(f *FourSeasons) error { return f.MoveTableauToFoundation(0, 0) },
		"tableau->tableau":    func(f *FourSeasons) error { return f.MoveTableauToTableau(0, 1) },
		"draw":                func(f *FourSeasons) error { return f.Draw() },
	}
	for name, move := range moves {
		t.Run(name, func(t *testing.T) {
			f := newTestFourSeasons()
			f.GiveUp()
			assert.Error(t, move(f))
		})
	}
}

func TestFourSeasons_WasteMovesRejectedWhenWasteEmpty(t *testing.T) {
	f := newTestFourSeasons()
	var fnd [FourSeasonsFoundationCnt][]*Card
	fnd[0] = []*Card{sp(12)}
	setFourSeasonsBoard(f, 12, [FourSeasonsTableauCnt][]*Card{}, fnd, nil, nil)
	assert.Error(t, f.MoveWasteToFoundation(0))
	assert.Error(t, f.MoveWasteToTableau(0))
}

func TestFourSeasons_TableauMoveRejections(t *testing.T) {
	f := newTestFourSeasons()
	var tab [FourSeasonsTableauCnt][]*Card
	tab[0] = []*Card{sp(5)}
	var fnd [FourSeasonsFoundationCnt][]*Card
	fnd[0] = []*Card{sp(12)}
	setFourSeasonsBoard(f, 12, tab, fnd, nil, nil)

	assert.Error(t, f.MoveTableauToFoundation(1, 0), "empty source pile")
	assert.Error(t, f.MoveTableauToFoundation(-1, 0), "tableau index below range")
	assert.Error(t, f.MoveTableauToFoundation(FourSeasonsTableauCnt, 0), "tableau index above range")
	assert.Error(t, f.MoveTableauToFoundation(0, -1), "foundation index below range")
	assert.Error(t, f.MoveTableauToFoundation(0, 0), "a 5 does not follow a Q")
	assert.Error(t, f.MoveWasteToTableau(FourSeasonsTableauCnt), "tableau index above range")
}

// A complete foundation takes nothing more. The guard is on the pile length, so
// it also holds for a pile a corrupt KV snapshot made too long.
func TestFourSeasons_Foundation_RejectsWhenComplete(t *testing.T) {
	f := newTestFourSeasons()
	var fnd [FourSeasonsFoundationCnt][]*Card
	fnd[0] = make([]*Card, CardValueMax)
	for i := range fnd[0] {
		fnd[0][i] = sp(5) // top value 5, so a 6 would otherwise follow
	}
	setFourSeasonsBoard(f, 5, [FourSeasonsTableauCnt][]*Card{}, fnd, []*Card{sp(6)}, nil)
	assert.Error(t, f.MoveWasteToFoundation(0))
}

// Wraparound arithmetic, both directions, at the seam.
func TestFourSeasons_RankWraparound(t *testing.T) {
	assert.Equal(t, 1, fourSeasonsNextRank(CardValueMax), "K -> A")
	assert.Equal(t, 2, fourSeasonsNextRank(1), "A -> 2")
	assert.Equal(t, CardValueMax, FourSeasonsPrevRank(1), "A -> K going down")
	assert.Equal(t, 1, FourSeasonsPrevRank(2), "2 -> A going down")
}

// The hint prefers a tableau card over the waste, because a tableau top blocks
// the cards beneath it while the waste can be drawn again.
func TestFourSeasons_GetHint_PrefersTableauOverWaste(t *testing.T) {
	f := newTestFourSeasons()
	var tab [FourSeasonsTableauCnt][]*Card
	tab[2] = []*Card{sp(13)}
	var fnd [FourSeasonsFoundationCnt][]*Card
	fnd[0] = []*Card{sp(12)}
	setFourSeasonsBoard(f, 12, tab, fnd, []*Card{sp(13)}, nil)

	h := f.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "tableau", h.FromZone)
}

// ...but it still finds a waste move when no tableau card can go.
func TestFourSeasons_GetHint_FallsBackToWaste(t *testing.T) {
	f := newTestFourSeasons()
	var fnd [FourSeasonsFoundationCnt][]*Card
	fnd[0] = []*Card{sp(12)}
	setFourSeasonsBoard(f, 12, [FourSeasonsTableauCnt][]*Card{}, fnd, []*Card{sp(13)}, nil)

	h := f.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "waste", h.FromZone)
	assert.Equal(t, -1, h.FromIdx)
	require.NoError(t, f.AutoComplete(), "and auto-complete can drive that branch too")
}

func TestFourSeasons_Snapshot_UnmarshalJSON_Rejections(t *testing.T) {
	oversized := func(field string) string {
		out := `{"` + field + `":[`
		for i := range fourSeasonsMaxSliceLen + 1 {
			if i > 0 {
				out += ","
			}
			out += `{"design":1,"value":1}`
		}
		return out + `]}`
	}
	for _, tt := range []struct{ name, data string }{
		{"malformed", `{`},
		{"stock too long", oversized("st")},
		{"waste too long", oversized("wa")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var s fourSeasonsSnapshot
			assert.Error(t, json.Unmarshal([]byte(tt.data), &s))
		})
	}
}

func TestFourSeasons_UnmarshalJSON_RejectsOversizedHistory(t *testing.T) {
	data := `{"ps":0,"mc":0,"br":1,"hi":[`
	for i := range fourSeasonsMaxSliceLen + 1 {
		if i > 0 {
			data += ","
		}
		data += `{"ps":0,"mc":0}`
	}
	data += `]}`
	f := NewFourSeasons(NewTrumpCards(0))
	assert.Error(t, json.Unmarshal([]byte(data), f))
}

func TestFourSeasons_UnmarshalJSON_RejectsOverlongFoundation(t *testing.T) {
	pile := ""
	for i := range CardValueMax + 1 {
		if i > 0 {
			pile += ","
		}
		pile += `{"design":1,"value":1}`
	}
	data := `{"ps":0,"mc":0,"br":1,"fd":[[` + pile + `],[],[],[]]}`
	f := NewFourSeasons(NewTrumpCards(0))
	assert.Error(t, json.Unmarshal([]byte(data), f))
}

// The restore path rebuilds state from a KV blob on every Worker request, so
// every array has to be bounded — not just the ones an honest client sends.
func TestFourSeasons_UnmarshalJSON_RejectsOversizedLiveArrays(t *testing.T) {
	big := func() string {
		out := ""
		for i := range fourSeasonsMaxSliceLen + 1 {
			if i > 0 {
				out += ","
			}
			out += `{"design":1,"value":1}`
		}
		return out
	}()
	for _, tt := range []struct{ name, data string }{
		{"stock", `{"ps":0,"mc":0,"br":1,"st":[` + big + `]}`},
		{"waste", `{"ps":0,"mc":0,"br":1,"wa":[` + big + `]}`},
		{"tableau pile", `{"ps":0,"mc":0,"br":1,"tb":[[` + big + `],[],[],[],[]]}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFourSeasons(NewTrumpCards(0))
			assert.Error(t, json.Unmarshal([]byte(tt.data), f))
		})
	}
}
