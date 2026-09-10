//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMatrimony() *Matrimony {
	c := NewDefaultMatrimony()
	c.Reset()
	c.stock = nil
	c.waste = nil
	c.history = nil
	c.moveCount = 0
	c.phase = MatrimonyPhasePlaying
	for i := range MatrimonyTableauCnt {
		c.tableau[i] = nil
	}
	for i := range MatrimonyFoundationCnt {
		c.foundation[i] = nil
	}
	return c
}

func TestMatrimonyResetUsesTwoDecksAndSixteenTableauSlots(t *testing.T) {
	c := NewDefaultMatrimony()
	c.Reset()
	assert.Len(t, c.GetTableau(), MatrimonyTableauCnt)
	assert.Equal(t, 88, c.GetStockCount())
	assert.Equal(t, MatrimonyTotalCards, c.GetStockCount()+MatrimonyTableauCnt)
}

func TestMatrimonyFoundationsStartCorrectly(t *testing.T) {
	c := newTestMatrimony()
	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, matrimonyQueen, true), 0))
	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignDiamond, matrimonyJack, true), 2))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignHeart, matrimonyQueen, true), 0))
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, matrimonyJack, true), 2))
}

func TestMatrimonyFoundationWrapsBothDirections(t *testing.T) {
	c := newTestMatrimony()
	for value := matrimonyQueen; value >= matrimonyAce; value-- {
		c.foundation[0] = append(c.foundation[0], NewCard(CardDesignSpade, value, true))
	}
	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, matrimonyKing, true), 0))
	for value := matrimonyJack; value <= matrimonyKing; value++ {
		c.foundation[2] = append(c.foundation[2], NewCard(CardDesignDiamond, value, true))
	}
	assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignDiamond, matrimonyAce, true), 2))
}

func TestMatrimonyMoveAndRedeal(t *testing.T) {
	c := newTestMatrimony()
	c.tableau[0] = NewCard(CardDesignSpade, matrimonyQueen, true)
	require.NoError(t, c.MoveTableauToFoundation(0))
	assert.Len(t, c.GetFoundation()[0], 1)
	c.waste = []*Card{NewCard(CardDesignHeart, 5, true)}
	for range MatrimonyMaxRedeals {
		c.stock = []*Card{NewCard(CardDesignClover, 4, true)}
		require.NoError(t, c.Draw())
		c.stock = nil
		require.NoError(t, c.Draw())
		assert.Equal(t, 1, c.GetStockCount())
	}
	c.stock = nil
	c.waste = []*Card{NewCard(CardDesignHeart, 5, true)}
	assert.Error(t, c.Draw())
}

func TestMatrimonyClearWhenAllCardsAreOnFoundations(t *testing.T) {
	c := newTestMatrimony()
	for i := range MatrimonyFoundationCnt {
		start := matrimonyFoundationStart[i]
		for n := 0; n < MatrimonyFoundationTarget; n++ {
			value := start.value
			for j := 0; j < n; j++ {
				if start.desc {
					value = matrimonyPreviousRank(value)
				} else {
					value = matrimonyNextRank(value)
				}
			}
			c.foundation[i] = append(c.foundation[i], NewCard(start.design, value, true))
		}
	}
	c.checkGameClear()
	assert.Equal(t, MatrimonyPhaseGameClear, c.GetPhase())
}

func TestMatrimonyMovesCardsFromAllPlayableZones(t *testing.T) {
	c := newTestMatrimony()
	c.tableau[0] = NewCard(CardDesignSpade, matrimonyQueen, true)
	require.NoError(t, c.MoveTableauToFoundation(0))
	assert.Nil(t, c.GetTableau()[0])

	c.waste = []*Card{NewCard(CardDesignDiamond, matrimonyJack, true)}
	require.NoError(t, c.MoveWasteToFoundation())
	assert.Empty(t, c.GetWaste())

	c.stock = []*Card{NewCard(CardDesignHeart, 7, true)}
	require.NoError(t, c.MoveStockToTableau(1))
	assert.Equal(t, 7, c.GetTableau()[1].GetValue())

	c.waste = []*Card{NewCard(CardDesignClover, 9, true)}
	require.NoError(t, c.MoveWasteToTableau(0))
	assert.Equal(t, 9, c.GetTableau()[0].GetValue())
}

func TestMatrimonyMoveRejectionsAndGameEnd(t *testing.T) {
	c := newTestMatrimony()
	assert.Error(t, c.MoveTableauToFoundation(-1))
	assert.Error(t, c.MoveTableauToFoundation(MatrimonyTableauCnt))
	assert.Error(t, c.MoveTableauToFoundation(0))
	c.tableau[0] = NewCard(CardDesignHeart, 5, true)
	assert.Error(t, c.MoveTableauToFoundation(0))

	assert.Error(t, c.MoveWasteToFoundation())
	c.waste = []*Card{NewCard(CardDesignHeart, 5, true)}
	assert.Error(t, c.MoveWasteToFoundation())

	assert.Error(t, c.MoveWasteToTableau(-1))
	assert.Error(t, c.MoveWasteToTableau(MatrimonyTableauCnt))
	assert.Error(t, c.MoveWasteToTableau(0))
	c.tableau[0] = nil
	c.waste = nil
	assert.Error(t, c.MoveWasteToTableau(0))
	c.waste = []*Card{NewCard(CardDesignHeart, 5, true)}
	c.tableau[0] = NewCard(CardDesignSpade, 5, true)
	assert.Error(t, c.MoveWasteToTableau(0))

	assert.Error(t, c.MoveStockToTableau(-1))
	assert.Error(t, c.MoveStockToTableau(MatrimonyTableauCnt))
	c.tableau[0] = nil
	c.stock = nil
	assert.Error(t, c.MoveStockToTableau(0))
	c.stock = []*Card{NewCard(CardDesignHeart, 5, true)}
	c.tableau[0] = NewCard(CardDesignSpade, 5, true)
	assert.Error(t, c.MoveStockToTableau(0))

	c.GiveUp()
	assert.Equal(t, MatrimonyPhaseGameOver, c.GetPhase())
	assert.Error(t, c.Draw())
	assert.Error(t, c.MoveTableauToFoundation(1))
	assert.Error(t, c.MoveWasteToFoundation())
	assert.Error(t, c.MoveWasteToTableau(1))
	assert.Error(t, c.MoveStockToTableau(1))
	assert.Error(t, c.AutoComplete())
}

func TestMatrimonyFoundationMovesWrapAtBothEnds(t *testing.T) {
	c := newTestMatrimony()
	for value := matrimonyQueen; value >= matrimonyAce; value-- {
		c.foundation[0] = append(c.foundation[0], NewCard(CardDesignSpade, value, true))
	}
	c.tableau[0] = NewCard(CardDesignSpade, matrimonyKing, true)
	require.NoError(t, c.MoveTableauToFoundation(0))
	assert.Equal(t, matrimonyKing, c.GetFoundation()[0][len(c.foundation[0])-1].GetValue())

	for value := matrimonyJack; value <= matrimonyKing; value++ {
		c.foundation[2] = append(c.foundation[2], NewCard(CardDesignDiamond, value, true))
	}
	c.waste = []*Card{NewCard(CardDesignDiamond, matrimonyAce, true)}
	require.NoError(t, c.MoveWasteToFoundation())
	assert.Equal(t, matrimonyAce, c.GetFoundation()[2][len(c.foundation[2])-1].GetValue())

	c.tableau[1] = NewCard(CardDesignSpade, matrimonyAce, true)
	assert.Error(t, c.MoveTableauToFoundation(1))
}

func TestMatrimonyAutoCompleteMovesTableauAndWasteUntilBlocked(t *testing.T) {
	c := newTestMatrimony()
	c.tableau[0] = NewCard(CardDesignSpade, matrimonyQueen, true)
	c.tableau[1] = NewCard(CardDesignSpade, matrimonyJack, true)
	c.waste = []*Card{NewCard(CardDesignSpade, matrimonyQueen, true)}
	require.NoError(t, c.AutoComplete())
	assert.Len(t, c.GetFoundation()[0], 2)
	assert.Len(t, c.GetFoundation()[1], 1)
	assert.Empty(t, c.GetWaste())

	assert.Error(t, c.AutoComplete())
}

func TestMatrimonyUndoRestoresMovesAndRejectsEmptyHistory(t *testing.T) {
	c := newTestMatrimony()
	assert.Error(t, c.Undo())
	c.stock = []*Card{NewCard(CardDesignHeart, 6, true)}
	require.NoError(t, c.MoveStockToTableau(0))
	assert.True(t, c.CanUndo())
	require.NoError(t, c.Undo())
	assert.Nil(t, c.GetTableau()[0])
	assert.Equal(t, 1, c.GetStockCount())
	assert.False(t, c.CanUndo())

	c.stock = []*Card{NewCard(CardDesignHeart, 6, true)}
	require.NoError(t, c.Draw())
	require.NoError(t, c.Draw())
	assert.Error(t, c.UndoN(4))
	require.NoError(t, c.UndoN(1))
	assert.Empty(t, c.GetWaste())
	assert.Equal(t, 1, c.GetStockCount())
	assert.Error(t, c.UndoN(0))
}

func TestMatrimonyJSONRoundTripRestoresPlayableBoardAndHistory(t *testing.T) {
	c := newTestMatrimony()
	c.tableau[0] = NewCard(CardDesignSpade, matrimonyQueen, true)
	c.stock = []*Card{NewCard(CardDesignHeart, 6, true)}
	require.NoError(t, c.Draw())
	data, err := json.Marshal(c)
	require.NoError(t, err)

	restored := NewDefaultMatrimony()
	require.NoError(t, json.Unmarshal(data, restored))
	assert.Equal(t, c.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, c.GetStockCount(), restored.GetStockCount())
	assert.Equal(t, c.GetWaste()[0].GetValue(), restored.GetWaste()[0].GetValue())
	assert.True(t, restored.CanUndo())
	require.NoError(t, restored.Undo())
	assert.Empty(t, restored.GetWaste())
	require.NoError(t, restored.MoveTableauToFoundation(0))
	assert.Len(t, restored.GetFoundation()[0], 1)
}

func TestMatrimonyJSONMissingFieldsUseZeroValues(t *testing.T) {
	restored := NewDefaultMatrimony()
	require.NoError(t, json.Unmarshal([]byte(`{"ps":0,"st":[{"d":3,"v":6,"w":true}],"tb":[null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null]}`), restored))
	require.NoError(t, restored.MoveStockToTableau(0))
	assert.Equal(t, 6, restored.GetTableau()[0].GetValue())

	var snap matrimonySnapshot
	require.NoError(t, json.Unmarshal([]byte(`{"tb":[null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null]}`), &snap))
	assert.Nil(t, snap.tableau[0])
}

func TestMatrimonyJSONRejectsInvalidAndOversizedInput(t *testing.T) {
	for _, data := range []string{`{`, `{"ps":-1}`, `{"ps":99}`, `{"mc":-1}`} {
		assert.Error(t, json.Unmarshal([]byte(data), NewDefaultMatrimony()))
	}
	big := make([]*Card, MatrimonyTotalCards+1)
	for i := range big {
		big[i] = NewCard(CardDesignSpade, matrimonyQueen, true)
	}
	for _, field := range []string{"Stock", "Waste"} {
		j := &matrimonyJSON{}
		if field == "Stock" {
			j.Stock = big
		} else {
			j.Waste = big
		}
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultMatrimony()))
	}
	var snap matrimonySnapshot
	data, err := json.Marshal(matrimonySnapshotJSON{Stock: make([]*Card, matrimonyMaxSliceLen+1)})
	require.NoError(t, err)
	assert.Error(t, json.Unmarshal(data, &snap))
}
