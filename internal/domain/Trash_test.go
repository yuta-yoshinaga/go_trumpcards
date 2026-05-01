//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// faceDownSlots returns 10 face-down placeholders backed by the supplied cards.
// Pass nil for a card to leave a slot card-less (useful for shape-only tests).
func faceDownSlots(cards [TrashSlotCnt]*Card) [TrashSlotCnt]TrashSlot {
	var out [TrashSlotCnt]TrashSlot
	for i := range out {
		out[i] = TrashSlot{Card: cards[i], FaceUp: false}
	}
	return out
}

// allDifferent returns 10 face-down placeholder cards (Ace..10 of spades) so
// each swap produces a distinct rank (helps assert chain placements).
func aceThroughTen() [TrashSlotCnt]*Card {
	var out [TrashSlotCnt]*Card
	for i := range out {
		out[i] = NewCard(CardDesignSpade, i+1, false)
	}
	return out
}

func TestNewDefaultTrashConstruction(t *testing.T) {
	tr := NewDefaultTrash()
	assert.NotNil(t, tr)
	assert.Equal(t, -1, tr.GetWinner())
	assert.Equal(t, TrashPhasePlayerTurn, tr.GetPhase())
}

func TestTrashResetLayout(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()

	assert.Equal(t, TrashPhasePlayerTurn, tr.GetPhase())
	assert.Equal(t, TrashHumanIdx, tr.GetCurrent())
	assert.Equal(t, -1, tr.GetWinner())
	assert.Nil(t, tr.GetPending())
	assert.Equal(t, 0, tr.GetMoveCount())
	assert.Equal(t, 0, tr.GetDiscardSize())

	// 54 cards total - 20 dealt = 34 in stock
	assert.Equal(t, 34, tr.GetStockSize())

	for i := range TrashPlayerCnt {
		slots := tr.GetPlayerSlots(i)
		require.Len(t, slots, TrashSlotCnt)
		for _, s := range slots {
			assert.NotNil(t, s.Card)
			assert.False(t, s.FaceUp)
		}
	}
	assert.False(t, tr.IsCpuPlayer(TrashHumanIdx))
	assert.True(t, tr.IsCpuPlayer(TrashCpuIdx))
}

func TestTrashDrawNumericPlace(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()
	// Force a deterministic stock + slot layout
	tr.SetPlayerSlots(TrashHumanIdx, faceDownSlots(aceThroughTen()))
	// Draw an Ace → must place at position 1, then chain into the displaced 1 card (Ace of Spades).
	// Set stock to: Ace of Hearts. Slot 0 holds Ace of Spades (face-down).
	tr.SetStock([]*Card{NewCard(CardDesignHeart, 1, false)})

	err := tr.Draw()
	require.NoError(t, err)

	slots := tr.GetPlayerSlots(TrashHumanIdx)
	// Ace of Hearts placed at slot 0, then displaced Ace of Spades chained into slot 0 (already face-up, ends turn).
	assert.True(t, slots[0].FaceUp)
	assert.Equal(t, CardDesignHeart, slots[0].Card.GetDesign())
	// Turn ends with displaced card discarded
	assert.Equal(t, TrashPhasePlayerTurn, tr.GetPhase())
	assert.Equal(t, TrashCpuIdx, tr.GetCurrent())
	assert.Equal(t, 1, tr.GetDiscardSize())
}

func TestTrashDrawWildAwaitsPlacement(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()
	tr.SetPlayerSlots(TrashHumanIdx, faceDownSlots(aceThroughTen()))
	tr.SetStock([]*Card{NewCard(CardDesignSpade, 13, false)}) // King = wild

	err := tr.Draw()
	require.NoError(t, err)

	assert.Equal(t, TrashPhaseAwaitWild, tr.GetPhase())
	assert.NotNil(t, tr.GetPending())
	assert.Equal(t, TrashHumanIdx, tr.GetCurrent())
}

func TestTrashDrawJokerAwaitsPlacement(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()
	tr.SetPlayerSlots(TrashHumanIdx, faceDownSlots(aceThroughTen()))
	tr.SetStock([]*Card{NewCard(CardDesignJoker, 1, false)})

	err := tr.Draw()
	require.NoError(t, err)
	assert.Equal(t, TrashPhaseAwaitWild, tr.GetPhase())
}

// TestTrashDrawWildAutoPlacesWithSingleEmptySlot pins issue #1565: when a
// wild card surfaces and the human has exactly one face-down slot, the
// game must auto-place the wild instead of asking the user to pick — there
// is no meaningful choice and the forced click breaks Trash's tempo.
func TestTrashDrawWildAutoPlacesWithSingleEmptySlot(t *testing.T) {
	t.Run("King wild auto-placed wins and discards the displaced card", func(t *testing.T) {
		// Filling the only face-down slot completes the board, so isWin
		// is always true on auto-placement (PR #1584 review). The
		// displaced face-down card cannot continue the chain — it is
		// pushed to the discard pile by the win path. We pin both
		// invariants here: AwaitWild is skipped, GameOver is reached,
		// and the displaced 5 lands on the discard pile.
		tr := NewDefaultTrash()
		tr.Reset()
		var slots [TrashSlotCnt]TrashSlot
		for i := range TrashSlotCnt - 1 {
			slots[i] = TrashSlot{Card: NewCard(CardDesignSpade, i+1, false), FaceUp: true}
		}
		slots[TrashSlotCnt-1] = TrashSlot{Card: NewCard(CardDesignHeart, 5, false), FaceUp: false}
		tr.SetPlayerSlots(TrashHumanIdx, slots)
		tr.SetStock([]*Card{NewCard(CardDesignDiamond, 13, false)})

		require.NoError(t, tr.Draw())

		out := tr.GetPlayerSlots(TrashHumanIdx)
		assert.True(t, out[TrashSlotCnt-1].FaceUp, "single-empty slot should be auto-flipped")
		assert.Equal(t, 13, out[TrashSlotCnt-1].Card.GetValue(), "wild should occupy the auto-placed slot")
		// Phase must NOT be AwaitWild — the forced-click was eliminated.
		assert.NotEqual(t, TrashPhaseAwaitWild, tr.GetPhase(),
			"auto-placement must skip the AwaitWild prompt")
		// The single-slot auto-fill always completes the board, so the
		// game ends and the displaced 5 is discarded.
		assert.Equal(t, TrashPhaseGameOver, tr.GetPhase())
		assert.Equal(t, TrashHumanIdx, tr.GetWinner())
		require.NotNil(t, tr.GetDiscardTop())
		assert.Equal(t, 5, tr.GetDiscardTop().GetValue(),
			"displaced face-down card should land on the discard pile")
	})

	t.Run("Joker wild auto-placed and triggers win", func(t *testing.T) {
		tr := NewDefaultTrash()
		tr.Reset()
		// Single face-down slot at position 10 holds a Q (end-turn card). The
		// wild lands there flipping it; the displaced Q would normally end the
		// turn, but a complete face-up board first triggers a win.
		var slots [TrashSlotCnt]TrashSlot
		for i := range TrashSlotCnt - 1 {
			slots[i] = TrashSlot{Card: NewCard(CardDesignSpade, i+1, false), FaceUp: true}
		}
		slots[TrashSlotCnt-1] = TrashSlot{Card: NewCard(CardDesignSpade, 12, false), FaceUp: false}
		tr.SetPlayerSlots(TrashHumanIdx, slots)
		tr.SetStock([]*Card{NewCard(CardDesignJoker, 1, false)})

		require.NoError(t, tr.Draw())

		assert.Equal(t, TrashPhaseGameOver, tr.GetPhase(), "completing the board must win")
		assert.Equal(t, TrashHumanIdx, tr.GetWinner())
	})

	t.Run("two empty slots still prompts (no auto)", func(t *testing.T) {
		tr := NewDefaultTrash()
		tr.Reset()
		var slots [TrashSlotCnt]TrashSlot
		for i := range TrashSlotCnt - 2 {
			slots[i] = TrashSlot{Card: NewCard(CardDesignSpade, i+1, false), FaceUp: true}
		}
		slots[8] = TrashSlot{Card: NewCard(CardDesignHeart, 9, false), FaceUp: false}
		slots[9] = TrashSlot{Card: NewCard(CardDesignHeart, 10, false), FaceUp: false}
		tr.SetPlayerSlots(TrashHumanIdx, slots)
		tr.SetStock([]*Card{NewCard(CardDesignDiamond, 13, false)})

		require.NoError(t, tr.Draw())

		assert.Equal(t, TrashPhaseAwaitWild, tr.GetPhase(),
			"two empty slots is a real choice; the human must pick")
	})
}

func TestTrashDrawJackEndsTurn(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()
	tr.SetPlayerSlots(TrashHumanIdx, faceDownSlots(aceThroughTen()))
	tr.SetStock([]*Card{NewCard(CardDesignSpade, 11, false)})

	err := tr.Draw()
	require.NoError(t, err)

	assert.Equal(t, TrashPhasePlayerTurn, tr.GetPhase())
	assert.Equal(t, TrashCpuIdx, tr.GetCurrent())
	assert.Equal(t, 1, tr.GetDiscardSize())
	assert.Equal(t, 11, tr.GetDiscardTop().GetValue())
}

func TestTrashDrawQueenEndsTurn(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()
	tr.SetPlayerSlots(TrashHumanIdx, faceDownSlots(aceThroughTen()))
	tr.SetStock([]*Card{NewCard(CardDesignSpade, 12, false)})

	err := tr.Draw()
	require.NoError(t, err)
	assert.Equal(t, TrashPhasePlayerTurn, tr.GetPhase())
	assert.Equal(t, TrashCpuIdx, tr.GetCurrent())
}

func TestTrashDrawErrors(t *testing.T) {
	t.Run("not in player turn", func(t *testing.T) {
		tr := NewDefaultTrash()
		tr.Reset()
		tr.SetPhase(TrashPhaseGameOver)
		assert.Error(t, tr.Draw())
	})
	t.Run("pending already set", func(t *testing.T) {
		tr := NewDefaultTrash()
		tr.Reset()
		tr.SetPending(NewCard(CardDesignSpade, 5, false))
		assert.Error(t, tr.Draw())
	})
	t.Run("empty stock and discard", func(t *testing.T) {
		tr := NewDefaultTrash()
		tr.Reset()
		tr.SetStock(nil)
		tr.SetDiscard(nil)
		assert.Error(t, tr.Draw())
	})
}

func TestTrashPlaceWildSwapsAndChains(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()
	// Slots: position 1 has a face-down 5; positions 2-10 have face-down Q (end-turn).
	var slots [TrashSlotCnt]TrashSlot
	slots[0] = TrashSlot{Card: NewCard(CardDesignHeart, 5, false), FaceUp: false}
	for i := 1; i < TrashSlotCnt; i++ {
		slots[i] = TrashSlot{Card: NewCard(CardDesignSpade, 12, false), FaceUp: false}
	}
	tr.SetPlayerSlots(TrashHumanIdx, slots)
	tr.SetStock([]*Card{NewCard(CardDesignDiamond, 13, false)}) // King wild

	require.NoError(t, tr.Draw())
	require.Equal(t, TrashPhaseAwaitWild, tr.GetPhase())

	// Place wild at position 1 → flips face-down 5; chain places 5 at position 5 → flips face-down Q → end turn.
	require.NoError(t, tr.PlaceWild(1))
	out := tr.GetPlayerSlots(TrashHumanIdx)
	assert.True(t, out[0].FaceUp)
	assert.Equal(t, 13, out[0].Card.GetValue())
	assert.True(t, out[4].FaceUp)
	assert.Equal(t, 5, out[4].Card.GetValue())
	assert.Equal(t, TrashCpuIdx, tr.GetCurrent())
	assert.Equal(t, TrashPhasePlayerTurn, tr.GetPhase())
}

func TestTrashPlaceWildErrors(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()
	t.Run("wrong phase", func(t *testing.T) {
		assert.Error(t, tr.PlaceWild(1))
	})
	t.Run("invalid position", func(t *testing.T) {
		tr.SetPhase(TrashPhaseAwaitWild)
		tr.SetPending(NewCard(CardDesignSpade, 13, false))
		assert.Error(t, tr.PlaceWild(0))
		assert.Error(t, tr.PlaceWild(11))
	})
	t.Run("slot already filled", func(t *testing.T) {
		tr.SetPhase(TrashPhaseAwaitWild)
		tr.SetPending(NewCard(CardDesignSpade, 13, false))
		var slots [TrashSlotCnt]TrashSlot
		for i := range slots {
			slots[i] = TrashSlot{Card: NewCard(CardDesignSpade, i+1, false), FaceUp: i == 0}
		}
		tr.SetPlayerSlots(TrashHumanIdx, slots)
		assert.Error(t, tr.PlaceWild(1))
	})
}

func TestTrashWinDetection(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()
	// Set human to be one slot away from winning (position 7 is the only face-down).
	var slots [TrashSlotCnt]TrashSlot
	for i := range slots {
		slots[i] = TrashSlot{Card: NewCard(CardDesignSpade, i+1, false), FaceUp: i != 6}
	}
	slots[6] = TrashSlot{Card: NewCard(CardDesignHeart, 12, false), FaceUp: false}
	tr.SetPlayerSlots(TrashHumanIdx, slots)
	tr.SetStock([]*Card{NewCard(CardDesignDiamond, 7, false)})

	require.NoError(t, tr.Draw())
	assert.Equal(t, TrashPhaseGameOver, tr.GetPhase())
	assert.Equal(t, TrashHumanIdx, tr.GetWinner())
}

func TestTrashStockRefillFromDiscard(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()
	tr.SetPlayerSlots(TrashHumanIdx, faceDownSlots(aceThroughTen()))
	tr.SetStock(nil)
	// All three discard cards are J/Q (end-turn) so the chain can never place anything
	// from stock — the only post-refill card to land in discard is whatever Draw consumed.
	preservedTop := NewCard(CardDesignDiamond, 12, false) // Q of diamonds, top of discard
	tr.SetDiscard([]*Card{
		NewCard(CardDesignSpade, 11, false), // J of spades — refilled into stock
		NewCard(CardDesignSpade, 12, false), // Q of spades — refilled into stock
		preservedTop,
	})

	require.Equal(t, 0, tr.GetStockSize())
	require.NoError(t, tr.Draw())

	// Refill behaviour: the drawn card was a J/Q from the shuffled rest, ending the turn
	// and being appended to the discard. So the discard now has the preserved top (still
	// present from before refill) plus the just-drawn card on top — total size 2.
	assert.Equal(t, 2, tr.GetDiscardSize())
	// The preserved top must still be present somewhere in the discard.
	found := false
	for i := range tr.GetDiscardSize() {
		c := tr.discard[i]
		if c.GetDesign() == preservedTop.GetDesign() && c.GetValue() == preservedTop.GetValue() {
			found = true
			break
		}
	}
	assert.True(t, found, "preserved discard-top card must remain after refill")
	// Refill moved 2 cards into stock; Draw consumed 1, so 1 remains.
	assert.Equal(t, 1, tr.GetStockSize())
	assert.Equal(t, TrashCpuIdx, tr.GetCurrent())
}

func TestTrashEndTurnAlternatesPlayers(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()
	// Force human's draw to be a Q (end turn immediately).
	tr.SetPlayerSlots(TrashHumanIdx, faceDownSlots(aceThroughTen()))
	tr.SetPlayerSlots(TrashCpuIdx, faceDownSlots(aceThroughTen()))
	tr.SetStock([]*Card{
		NewCard(CardDesignSpade, 11, false), // human draws → end turn
		NewCard(CardDesignSpade, 11, false), // cpu draws → end turn
	})

	require.NoError(t, tr.Draw())
	assert.Equal(t, TrashCpuIdx, tr.GetCurrent())
	require.NoError(t, tr.CpuStep())
	assert.Equal(t, TrashHumanIdx, tr.GetCurrent())
}

func TestTrashCpuStepHeuristicAwaitWild(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()
	tr.SetCurrent(TrashCpuIdx)
	tr.SetPhase(TrashPhaseAwaitWild)
	tr.SetPending(NewCard(CardDesignDiamond, 13, false)) // King wild

	// CPU layout: position 10 face-down (target), positions 1-9 already face-up (with placeholder cards).
	var slots [TrashSlotCnt]TrashSlot
	for i := range slots {
		slots[i] = TrashSlot{Card: NewCard(CardDesignSpade, i+1, false), FaceUp: i != 9}
	}
	tr.SetPlayerSlots(TrashCpuIdx, slots)

	require.NoError(t, tr.CpuStep())
	out := tr.GetPlayerSlots(TrashCpuIdx)
	assert.True(t, out[9].FaceUp, "CPU should have placed wild at highest face-down index")
	assert.Equal(t, 13, out[9].Card.GetValue())
	// All 10 face-up → win
	assert.Equal(t, TrashPhaseGameOver, tr.GetPhase())
	assert.Equal(t, TrashCpuIdx, tr.GetWinner())
}

func TestTrashCpuStepDraw(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()
	tr.SetCurrent(TrashCpuIdx)
	tr.SetPlayerSlots(TrashCpuIdx, faceDownSlots(aceThroughTen()))
	tr.SetStock([]*Card{NewCard(CardDesignSpade, 11, false)}) // J → end turn

	require.NoError(t, tr.CpuStep())
	assert.Equal(t, TrashHumanIdx, tr.GetCurrent())
}

func TestTrashCpuStepErrors(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()
	t.Run("not cpu turn", func(t *testing.T) {
		assert.Error(t, tr.CpuStep())
	})
	t.Run("game over", func(t *testing.T) {
		tr.SetCurrent(TrashCpuIdx)
		tr.SetPhase(TrashPhaseGameOver)
		assert.Error(t, tr.CpuStep())
	})
	t.Run("await wild but no face-down slot", func(t *testing.T) {
		tr2 := NewDefaultTrash()
		tr2.Reset()
		tr2.SetCurrent(TrashCpuIdx)
		tr2.SetPhase(TrashPhaseAwaitWild)
		tr2.SetPending(NewCard(CardDesignSpade, 13, false))
		var slots [TrashSlotCnt]TrashSlot
		for i := range slots {
			slots[i] = TrashSlot{Card: NewCard(CardDesignSpade, i+1, false), FaceUp: true}
		}
		tr2.SetPlayerSlots(TrashCpuIdx, slots)
		assert.Error(t, tr2.CpuStep())
	})
}

func TestTrashIsCpuTurn(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()
	assert.False(t, tr.IsCpuTurn())
	tr.SetCurrent(TrashCpuIdx)
	assert.True(t, tr.IsCpuTurn())
	tr.SetPhase(TrashPhaseGameOver)
	assert.False(t, tr.IsCpuTurn())
}

func TestTrashGetPlayerSlotsBounds(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()
	assert.Nil(t, tr.GetPlayerSlots(-1))
	assert.Nil(t, tr.GetPlayerSlots(TrashPlayerCnt))
}

func TestTrashSetPlayerSlotsBounds(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()
	// Out-of-bounds writes are no-ops.
	tr.SetPlayerSlots(-1, faceDownSlots(aceThroughTen()))
	tr.SetPlayerSlots(TrashPlayerCnt, faceDownSlots(aceThroughTen()))
	// Slots remain populated from Reset.
	assert.Len(t, tr.GetPlayerSlots(TrashHumanIdx), TrashSlotCnt)
}

func TestTrashIsCpuPlayerBounds(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()
	assert.False(t, tr.IsCpuPlayer(-1))
	assert.False(t, tr.IsCpuPlayer(TrashPlayerCnt))
}

func TestTrashActionLogPopulated(t *testing.T) {
	tr := NewDefaultTrash()
	tr.Reset()
	tr.SetPlayerSlots(TrashHumanIdx, faceDownSlots(aceThroughTen()))
	tr.SetStock([]*Card{NewCard(CardDesignSpade, 11, false)})
	require.NoError(t, tr.Draw())
	log := tr.GetActionLog()
	assert.NotEmpty(t, log)
	// last entry should be "end" since J ends the turn
	assert.Equal(t, "end", log[len(log)-1].ActionType)
}

func TestTrashJSONRoundtrip(t *testing.T) {
	original := NewDefaultTrash()
	original.Reset()
	original.SetPlayerSlots(TrashHumanIdx, faceDownSlots(aceThroughTen()))
	original.SetStock([]*Card{NewCard(CardDesignSpade, 5, false)})
	require.NoError(t, original.Draw())

	data, err := json.Marshal(original)
	require.NoError(t, err)

	restored := &Trash{}
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, original.GetPhase(), restored.GetPhase())
	assert.Equal(t, original.GetCurrent(), restored.GetCurrent())
	assert.Equal(t, original.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, original.GetWinner(), restored.GetWinner())
	assert.Equal(t, original.GetStockSize(), restored.GetStockSize())
	assert.Equal(t, original.GetDiscardSize(), restored.GetDiscardSize())
}

func TestTrashUnmarshalRejectsOversizedSlices(t *testing.T) {
	// Build a payload with stock of 201 entries.
	bigStock := make([]map[string]any, 201)
	for i := range bigStock {
		bigStock[i] = map[string]any{"d": 1, "v": 5, "w": false}
	}
	payload := map[string]any{"st": bigStock}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	tr := &Trash{}
	assert.Error(t, json.Unmarshal(data, tr))
}

func TestTrashUnmarshalRejectsWinnerWhileNotGameOver(t *testing.T) {
	cases := []struct {
		name    string
		payload string
	}{
		{"winner=1 with PlayerTurn", `{"wn":1,"ph":0}`},
		{"winner=0 with AwaitWild", `{"wn":0,"ph":1}`},
		{"winner=1 with AwaitWild", `{"wn":1,"ph":1}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tr := &Trash{}
			require.NoError(t, json.Unmarshal([]byte(c.payload), tr))
			assert.Equal(t, -1, tr.GetWinner(), "winner must reset to -1 until GameOver")
		})
	}

	t.Run("winner preserved on GameOver", func(t *testing.T) {
		tr := &Trash{}
		require.NoError(t, json.Unmarshal([]byte(`{"wn":1,"ph":2}`), tr))
		assert.Equal(t, 1, tr.GetWinner())
		assert.Equal(t, TrashPhaseGameOver, tr.GetPhase())
	})
}

func TestTrashUnmarshalDefaults(t *testing.T) {
	tr := &Trash{}
	require.NoError(t, json.Unmarshal([]byte(`{}`), tr))
	assert.NotNil(t, tr.GetPlayerSlots(TrashHumanIdx))
	assert.Equal(t, 0, tr.GetStockSize())
	assert.Equal(t, 0, tr.GetDiscardSize())
	assert.Equal(t, -1, tr.GetWinner())
}

func TestTrashHelperPredicates(t *testing.T) {
	assert.False(t, isTrashWild(nil))
	assert.True(t, isTrashWild(NewCard(CardDesignSpade, 13, false)))
	assert.True(t, isTrashWild(NewCard(CardDesignJoker, 1, false)))
	assert.False(t, isTrashWild(NewCard(CardDesignSpade, 5, false)))

	assert.False(t, isTrashEndTurn(nil))
	assert.True(t, isTrashEndTurn(NewCard(CardDesignSpade, 11, false)))
	assert.True(t, isTrashEndTurn(NewCard(CardDesignSpade, 12, false)))
	assert.False(t, isTrashEndTurn(NewCard(CardDesignJoker, 1, false)))

	assert.Equal(t, 0, trashCardPosition(nil))
	assert.Equal(t, 0, trashCardPosition(NewCard(CardDesignJoker, 1, false)))
	assert.Equal(t, 0, trashCardPosition(NewCard(CardDesignSpade, 11, false)))
	assert.Equal(t, 1, trashCardPosition(NewCard(CardDesignSpade, 1, false)))
	assert.Equal(t, 10, trashCardPosition(NewCard(CardDesignSpade, 10, false)))
}
