//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAuldLangSyne() *AuldLangSyne {
	a := NewAuldLangSyne(NewTrumpCards(0))
	a.Reset()
	return a
}

// acesOnFoundations returns the opening foundations: one Ace on each of the four.
func acesOnFoundations() [AuldLangSyneFoundationCnt][]*Card {
	var fs [AuldLangSyneFoundationCnt][]*Card
	for i := range fs {
		fs[i] = []*Card{NewCard(i, 1, true)}
	}
	return fs
}

// setAuldLangSyneBoard installs an exact board so a test can drive the game
// deterministically. Reset shuffles, so nothing may be asserted on top of it.
func setAuldLangSyneBoard(a *AuldLangSyne, foundations, wastes [AuldLangSyneFoundationCnt][]*Card, stock []*Card) {
	a.foundations = foundations
	a.wastes = wastes
	a.stock = stock
	a.phase = AuldLangSynePhasePlaying
	a.isStalemate = false
	a.history = nil
	a.moveCount = 0
	a.actionLog = nil
}

func TestAuldLangSyne_Reset_InitialState(t *testing.T) {
	a := newTestAuldLangSyne()

	assert.Equal(t, AuldLangSynePhasePlaying, a.GetPhase())
	assert.Equal(t, 0, a.GetMoveCount())
	assert.Empty(t, a.GetActionLog())
	assert.False(t, a.IsStalemate())
	assert.False(t, a.GetGameEndFlag())
	assert.False(t, a.CanUndo())

	// The four Aces are pre-placed -- that is the difference from Sir Tommy,
	// where a foundation is opened by an Ace drawn out of the stock.
	designs := map[int]bool{}
	for i, f := range a.GetFoundations() {
		require.Len(t, f, 1, "foundation %d should open with one Ace", i)
		assert.Equal(t, 1, f[0].GetValue(), "foundation %d should open with an Ace", i)
		designs[f[0].GetDesign()] = true
	}
	assert.Len(t, designs, AuldLangSyneFoundationCnt, "all four Aces, one per suit")

	// The opening deal puts one card on each waste, so 48 - 4 = 44 remain.
	for i, w := range a.GetWastes() {
		assert.Len(t, w, 1, "waste %d should open with one card", i)
	}
	assert.Equal(t, 44, a.GetStockCount())
	assert.False(t, a.AllFaceUp())
}

// The Aces are removed from the deck, not merely dealt early: none may reappear.
func TestAuldLangSyne_Reset_NoAceOutsideFoundations(t *testing.T) {
	for range 20 {
		a := newTestAuldLangSyne()
		for i, w := range a.GetWastes() {
			for _, c := range w {
				assert.NotEqual(t, 1, c.GetValue(), "waste %d holds an Ace", i)
			}
		}
		for _, c := range a.stock {
			assert.NotEqual(t, 1, c.GetValue(), "stock holds an Ace")
		}
		assert.Equal(t, 48, a.GetStockCount()+4, "48 non-Ace cards are in play")
	}
}

func TestAuldLangSyne_Deal_AddsOneCardToEveryWaste(t *testing.T) {
	a := newTestAuldLangSyne()
	before := a.GetStockCount()

	require.NoError(t, a.Deal())

	assert.Equal(t, before-AuldLangSyneWasteCnt, a.GetStockCount())
	for i, w := range a.GetWastes() {
		assert.Len(t, w, 2, "waste %d should have received a second card", i)
	}
	assert.Equal(t, 1, a.GetMoveCount())
	assert.Len(t, a.GetActionLog(), 1)
	assert.True(t, a.CanUndo(), "a deal must be undoable")
}

// Dealing is forced onto all four wastes: unlike Sir Tommy the player never
// chooses where a card lands, which is why there is no PlayStockToWaste here.
func TestAuldLangSyne_Deal_ExhaustsStockInTwelveRows(t *testing.T) {
	a := newTestAuldLangSyne()
	rows := 1 // Reset already dealt the first row
	for a.GetStockCount() > 0 {
		require.NoError(t, a.Deal())
		rows++
	}
	assert.Equal(t, 12, rows, "48 non-Ace cards deal out as 12 rows of 4")
	assert.True(t, a.AllFaceUp())
	assert.Error(t, a.Deal(), "no deal once the stock is empty")
}

func TestAuldLangSyne_Deal_RejectedWhenNotPlaying(t *testing.T) {
	a := newTestAuldLangSyne()
	a.GiveUp()
	assert.Error(t, a.Deal())
}

func TestAuldLangSyne_PlayWasteToFoundation_Success(t *testing.T) {
	a := newTestAuldLangSyne()
	var wastes [AuldLangSyneWasteCnt][]*Card
	wastes[1] = []*Card{NewCard(2, 7, true), NewCard(0, 2, true)}
	setAuldLangSyneBoard(a, acesOnFoundations(), wastes, nil)

	require.NoError(t, a.PlayWasteToFoundation(1, 0))

	assert.Len(t, a.GetFoundations()[0], 2)
	assert.Equal(t, 2, a.GetFoundations()[0][1].GetValue())
	assert.Len(t, a.GetWastes()[1], 1, "only the top card moves")
	assert.Equal(t, 7, a.GetWastes()[1][0].GetValue(), "the buried card stays buried")
	assert.Equal(t, 1, a.GetMoveCount())
	assert.Len(t, a.GetActionLog(), 1)
}

func TestAuldLangSyne_PlayWasteToFoundation_Rejections(t *testing.T) {
	tests := []struct {
		name     string
		wasteIdx int
		fIdx     int
	}{
		{"waste index below range", -1, 0},
		{"waste index above range", AuldLangSyneWasteCnt, 0},
		{"foundation index below range", 1, -1},
		{"foundation index above range", 1, AuldLangSyneFoundationCnt},
		{"empty waste", 2, 0},
		{"rank does not follow", 3, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestAuldLangSyne()
			var wastes [AuldLangSyneWasteCnt][]*Card
			wastes[1] = []*Card{NewCard(0, 2, true)} // playable on an Ace
			wastes[3] = []*Card{NewCard(0, 9, true)} // not playable on an Ace
			setAuldLangSyneBoard(a, acesOnFoundations(), wastes, nil)

			assert.Error(t, a.PlayWasteToFoundation(tt.wasteIdx, tt.fIdx))
			assert.Equal(t, 0, a.GetMoveCount(), "a rejected move must not count")
			assert.False(t, a.CanUndo(), "a rejected move must not snapshot")
		})
	}
}

func TestAuldLangSyne_PlayWasteToFoundation_RejectedWhenNotPlaying(t *testing.T) {
	a := newTestAuldLangSyne()
	var wastes [AuldLangSyneWasteCnt][]*Card
	wastes[0] = []*Card{NewCard(0, 2, true)}
	setAuldLangSyneBoard(a, acesOnFoundations(), wastes, nil)
	a.GiveUp()

	assert.Error(t, a.PlayWasteToFoundation(0, 0))
}

// A foundation that is already complete takes nothing more. The guard is on the
// pile length, so it also holds for a pile a corrupt KV snapshot made too long.
func TestAuldLangSyne_PlayWasteToFoundation_RejectedOnCompleteFoundation(t *testing.T) {
	a := newTestAuldLangSyne()
	fs := acesOnFoundations()
	fs[0] = make([]*Card, CardValueMax)
	for i := range fs[0] {
		fs[0][i] = NewCard(0, 5, true) // top value 5, so a 6 would otherwise follow
	}
	var wastes [AuldLangSyneWasteCnt][]*Card
	wastes[0] = []*Card{NewCard(1, 6, true)}
	setAuldLangSyneBoard(a, fs, wastes, nil)

	assert.Error(t, a.PlayWasteToFoundation(0, 0))
}

func TestAuldLangSyne_GameClear(t *testing.T) {
	a := newTestAuldLangSyne()
	var fs [AuldLangSyneFoundationCnt][]*Card
	var wastes [AuldLangSyneWasteCnt][]*Card
	for i := range fs {
		fs[i] = make([]*Card, 0, CardValueMax)
		for v := 1; v < CardValueMax; v++ {
			fs[i] = append(fs[i], NewCard(i, v, true))
		}
		wastes[i] = []*Card{NewCard(i, CardValueMax, true)}
	}
	setAuldLangSyneBoard(a, fs, wastes, nil)

	for i := range AuldLangSyneWasteCnt {
		require.NoError(t, a.PlayWasteToFoundation(i, i))
	}

	assert.Equal(t, AuldLangSynePhaseGameClear, a.GetPhase())
	assert.True(t, a.GetGameEndFlag())
	assert.False(t, a.IsStalemate(), "a win is not a stalemate")
}

// The stock is the only escape hatch: while cards remain the player can always
// deal, so a blocked board is not yet a stalemate.
func TestAuldLangSyne_NotStalemateWhileStockRemains(t *testing.T) {
	a := newTestAuldLangSyne()
	var wastes [AuldLangSyneWasteCnt][]*Card
	for i := range wastes {
		wastes[i] = []*Card{NewCard(i, 9, true)}
	}
	// Five cards: a deal consumes four and leaves one, so the board is blocked
	// but the player still has a deal in hand.
	stock := make([]*Card, 0, 5)
	for range 5 {
		stock = append(stock, NewCard(0, 9, true))
	}
	setAuldLangSyneBoard(a, acesOnFoundations(), wastes, stock)

	require.NoError(t, a.Deal())

	require.Equal(t, 1, a.GetStockCount())
	assert.Nil(t, a.GetHint(), "the board really is blocked")
	assert.False(t, a.IsStalemate(), "but a deal remains, so it is not a stalemate")
}

// Dealing the last card into a blocked board is the moment it becomes one.
func TestAuldLangSyne_StalemateOnFinalDeal(t *testing.T) {
	a := newTestAuldLangSyne()
	var wastes [AuldLangSyneWasteCnt][]*Card
	for i := range wastes {
		wastes[i] = []*Card{NewCard(i, 9, true)}
	}
	setAuldLangSyneBoard(a, acesOnFoundations(), wastes, []*Card{NewCard(0, 5, true)})

	require.NoError(t, a.Deal())

	assert.True(t, a.AllFaceUp())
	assert.True(t, a.IsStalemate())
}

func TestAuldLangSyne_Stalemate(t *testing.T) {
	a := newTestAuldLangSyne()
	var wastes [AuldLangSyneWasteCnt][]*Card
	wastes[0] = []*Card{NewCard(0, 9, true)}
	wastes[1] = []*Card{NewCard(1, 2, true)} // the only playable card
	setAuldLangSyneBoard(a, acesOnFoundations(), wastes, nil)

	assert.Equal(t, 0, a.UndoToEscape(), "not stalemate yet")
	require.NoError(t, a.PlayWasteToFoundation(1, 0))

	assert.True(t, a.IsStalemate(), "nothing left to play and no stock to deal")
	assert.Equal(t, AuldLangSynePhasePlaying, a.GetPhase(), "stalemate keeps the game open for undo")
	assert.Equal(t, 1, a.UndoToEscape(), "one undo returns to a position with a move")
}

func TestAuldLangSyne_UndoToEscape_Unreachable(t *testing.T) {
	a := newTestAuldLangSyne()
	setAuldLangSyneBoard(a, acesOnFoundations(), [AuldLangSyneWasteCnt][]*Card{}, nil)
	a.isStalemate = true

	assert.Equal(t, -1, a.UndoToEscape(), "no history to rewind into")
}

func TestAuldLangSyne_GetHint(t *testing.T) {
	a := newTestAuldLangSyne()
	var wastes [AuldLangSyneWasteCnt][]*Card
	wastes[0] = []*Card{NewCard(0, 9, true)}
	wastes[2] = []*Card{NewCard(1, 2, true)}
	setAuldLangSyneBoard(a, acesOnFoundations(), wastes, nil)

	h := a.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, 2, h.WasteIdx)
	assert.GreaterOrEqual(t, h.FoundationIdx, 0)

	require.NoError(t, a.PlayWasteToFoundation(h.WasteIdx, h.FoundationIdx))
	assert.Nil(t, a.GetHint(), "nothing playable remains")
}

func TestAuldLangSyne_GetHint_NilWhenNotPlaying(t *testing.T) {
	a := newTestAuldLangSyne()
	a.GiveUp()
	assert.Nil(t, a.GetHint())
}

// AutoComplete finishes an endgame; it does not play the deal for you. The
// frontend button already requires an empty stock, so the API must agree.
func TestAuldLangSyne_AutoComplete_RequiresEmptyStock(t *testing.T) {
	a := newTestAuldLangSyne()
	assert.Error(t, a.AutoComplete())
}

func TestAuldLangSyne_AutoComplete_DrainsPlayableCards(t *testing.T) {
	a := newTestAuldLangSyne()
	var wastes [AuldLangSyneWasteCnt][]*Card
	// 4 buried under 3 buried under 2: only sequential play frees them, which is
	// why AutoComplete re-scans the whole board after every single card.
	wastes[0] = []*Card{NewCard(0, 4, true), NewCard(0, 3, true), NewCard(0, 2, true)}
	setAuldLangSyneBoard(a, acesOnFoundations(), wastes, nil)

	require.NoError(t, a.AutoComplete())

	assert.Empty(t, a.GetWastes()[0])
	assert.Len(t, a.GetFoundations()[0], 4, "A-2-3-4 all landed on one foundation")
}

func TestAuldLangSyne_AutoComplete_NoMoveAvailable(t *testing.T) {
	a := newTestAuldLangSyne()
	var wastes [AuldLangSyneWasteCnt][]*Card
	wastes[0] = []*Card{NewCard(0, 9, true)}
	setAuldLangSyneBoard(a, acesOnFoundations(), wastes, nil)

	assert.Error(t, a.AutoComplete())
}

func TestAuldLangSyne_AutoComplete_RejectedWhenNotPlaying(t *testing.T) {
	a := newTestAuldLangSyne()
	a.GiveUp()
	assert.Error(t, a.AutoComplete())
}

func TestAuldLangSyne_Undo(t *testing.T) {
	a := newTestAuldLangSyne()
	var wastes [AuldLangSyneWasteCnt][]*Card
	wastes[0] = []*Card{NewCard(0, 2, true)}
	setAuldLangSyneBoard(a, acesOnFoundations(), wastes, nil)

	assert.Error(t, a.Undo(), "nothing to undo yet")
	require.NoError(t, a.PlayWasteToFoundation(0, 0))
	require.NoError(t, a.Undo())

	assert.Len(t, a.GetFoundations()[0], 1)
	assert.Len(t, a.GetWastes()[0], 1)
	assert.Equal(t, 0, a.GetMoveCount())
	assert.False(t, a.CanUndo())
}

func TestAuldLangSyne_UndoN(t *testing.T) {
	a := newTestAuldLangSyne()
	var wastes [AuldLangSyneWasteCnt][]*Card
	wastes[0] = []*Card{NewCard(0, 3, true), NewCard(0, 2, true)}
	setAuldLangSyneBoard(a, acesOnFoundations(), wastes, nil)
	require.NoError(t, a.PlayWasteToFoundation(0, 0))
	require.NoError(t, a.PlayWasteToFoundation(0, 0))

	assert.Error(t, a.UndoN(0), "n must be positive")
	assert.Error(t, a.UndoN(3), "not enough history")
	require.NoError(t, a.UndoN(2))

	assert.Len(t, a.GetFoundations()[0], 1)
	assert.Len(t, a.GetWastes()[0], 2)
	assert.Equal(t, 0, a.GetMoveCount())
}

func TestAuldLangSyne_GiveUp(t *testing.T) {
	a := newTestAuldLangSyne()
	a.GiveUp()

	assert.Equal(t, AuldLangSynePhaseGameOver, a.GetPhase())
	assert.True(t, a.GetGameEndFlag())
	assert.Len(t, a.GetActionLog(), 1)

	a.GiveUp()
	assert.Len(t, a.GetActionLog(), 1, "giving up twice logs once")
}

func TestAuldLangSyne_JSONRoundTrip(t *testing.T) {
	a := newTestAuldLangSyne()
	require.NoError(t, a.Deal())
	var wastes [AuldLangSyneWasteCnt][]*Card
	wastes[0] = []*Card{NewCard(0, 2, true)}
	setAuldLangSyneBoard(a, acesOnFoundations(), wastes, []*Card{NewCard(1, 8, true)})
	require.NoError(t, a.PlayWasteToFoundation(0, 0))

	data, err := json.Marshal(a)
	require.NoError(t, err)

	restored := NewAuldLangSyne(NewTrumpCards(0))
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, a.GetFoundations(), restored.GetFoundations())
	assert.Equal(t, a.GetWastes(), restored.GetWastes())
	assert.Equal(t, a.GetStockCount(), restored.GetStockCount())
	assert.Equal(t, a.GetPhase(), restored.GetPhase())
	assert.Equal(t, a.GetMoveCount(), restored.GetMoveCount())
	assert.Len(t, restored.GetActionLog(), len(a.GetActionLog()))

	// The undo stack must survive: the Worker rebuilds the game from KV on every
	// request, so an unpersisted history means Undo silently never works (#4478).
	require.True(t, restored.CanUndo())
	require.NoError(t, restored.Undo())
	assert.Len(t, restored.GetWastes()[0], 1, "undo rewinds the board, not just the depth")
	assert.Len(t, restored.GetFoundations()[0], 1)
}

func TestAuldLangSyne_UnmarshalJSON_Rejections(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"malformed", `{`},
		{"phase below range", `{"ps":-1}`},
		{"phase above range", `{"ps":99}`},
		{"negative move count", `{"ps":0,"mc":-1}`},
		{"foundation longer than a suit", `{"ps":0,"mc":0,"fd":[` +
			`[{"design":0,"value":1},{"design":0,"value":1},{"design":0,"value":1},{"design":0,"value":1},` +
			`{"design":0,"value":1},{"design":0,"value":1},{"design":0,"value":1},{"design":0,"value":1},` +
			`{"design":0,"value":1},{"design":0,"value":1},{"design":0,"value":1},{"design":0,"value":1},` +
			`{"design":0,"value":1},{"design":0,"value":1}],[],[],[]]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAuldLangSyne(NewTrumpCards(0))
			assert.Error(t, json.Unmarshal([]byte(tt.data), a))
		})
	}
}

func TestAuldLangSyne_NewDefault(t *testing.T) {
	a := NewDefaultAuldLangSyne()
	require.NotNil(t, a)
	a.Reset()
	assert.Equal(t, 44, a.GetStockCount())
}

// The empty-foundation branch only happens after restoring a corrupt snapshot --
// Reset seeds all four -- but it must still accept nothing but an Ace.
func TestAuldLangSyne_EmptyFoundationTakesOnlyAnAce(t *testing.T) {
	a := newTestAuldLangSyne()
	var fs [AuldLangSyneFoundationCnt][]*Card
	var wastes [AuldLangSyneWasteCnt][]*Card
	wastes[0] = []*Card{NewCard(0, 5, true)}
	wastes[1] = []*Card{NewCard(1, 1, true)}
	setAuldLangSyneBoard(a, fs, wastes, nil)

	assert.Error(t, a.PlayWasteToFoundation(0, 0), "a 5 cannot open a foundation")
	require.NoError(t, a.PlayWasteToFoundation(1, 0), "an Ace can")
}

func TestAuldLangSyne_Snapshot_UnmarshalJSON_Rejections(t *testing.T) {
	oversized := func(field string) string {
		out := `{"` + field + `":[`
		for i := range auldLangSyneMaxSliceLen + 1 {
			if i > 0 {
				out += ","
			}
			out += `{"design":0,"value":1}`
		}
		return out + `]}`
	}
	nestedOversized := func(field string) string {
		out := `{"` + field + `":[[`
		for i := range auldLangSyneMaxSliceLen + 1 {
			if i > 0 {
				out += ","
			}
			out += `{"design":0,"value":1}`
		}
		return out + `],[],[],[]]}`
	}
	tests := []struct {
		name string
		data string
	}{
		{"malformed", `{`},
		{"stock too long", oversized("st")},
		{"foundation pile too long", nestedOversized("fs")},
		{"waste pile too long", nestedOversized("ws")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s auldLangSyneSnapshot
			assert.Error(t, json.Unmarshal([]byte(tt.data), &s))
		})
	}
}

func TestAuldLangSyne_UnmarshalJSON_RejectsOversizedHistory(t *testing.T) {
	data := `{"ps":0,"mc":0,"hi":[`
	for i := range auldLangSyneMaxSliceLen + 1 {
		if i > 0 {
			data += ","
		}
		data += `{"ps":0,"mc":0}`
	}
	data += `]}`

	a := NewAuldLangSyne(NewTrumpCards(0))
	assert.Error(t, json.Unmarshal([]byte(data), a))
}
