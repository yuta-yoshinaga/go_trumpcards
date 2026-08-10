//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCribbageSquares() *CribbageSquares {
	c := NewDefaultCribbageSquares()
	c.Reset()
	return c
}

// fillCribbageSquares places every card the deal offers, so the board reaches
// the state where the starter is turned. It never asserts on which cards land
// where -- the deal is shuffled.
func fillCribbageSquares(t *testing.T, c *CribbageSquares) {
	t.Helper()
	for i := range CribbageSquaresTotalCells {
		require.NoError(t, c.Place(i/CribbageSquaresGridSize, i%CribbageSquaresGridSize))
	}
}

// setCribbageBoard installs an exact board so a test can state the position it
// cares about. Never assert on a freshly Reset board -- the deal is shuffled.
func setCribbageBoard(c *CribbageSquares, rows [CribbageSquaresGridSize][CribbageSquaresGridSize]*Card, starter *Card) {
	c.board = rows
	c.starter = starter
	c.placedCount = CribbageSquaresTotalCells
	c.phase = CribbageSquaresPhaseComplete
	c.currentCard = nil
}

func TestNewCribbageSquares(t *testing.T) {
	assert.NotNil(t, NewCribbageSquares(NewTrumpCards(0)))
	assert.NotNil(t, NewDefaultCribbageSquares())
}

func TestCribbageSquares_Reset(t *testing.T) {
	c := newTestCribbageSquares()

	assert.Equal(t, CribbageSquaresPhasePlaying, c.GetPhase())
	assert.Equal(t, 0, c.GetPlacedCount())
	assert.NotNil(t, c.GetCurrentCard(), "a card is waiting to be placed")
	assert.Nil(t, c.GetStarter(), "the starter is not turned until the board is full")
	assert.False(t, c.GetGameEndFlag())
	assert.False(t, c.CanUndo())
	assert.Empty(t, c.GetActionLog())

	for r, row := range c.GetBoard() {
		for col, cell := range row {
			assert.Nil(t, cell, "cell (%d,%d)", r, col)
		}
	}
}

func TestCribbageSquares_GridIs4x4(t *testing.T) {
	assert.Equal(t, 4, CribbageSquaresGridSize)
	assert.Equal(t, 16, CribbageSquaresTotalCells)
	// Four rows plus four columns is the eight hands the game scores.
	assert.Equal(t, 8, CribbageSquaresLineCnt)
}

func TestCribbageSquares_ResetTwiceIsClean(t *testing.T) {
	c := newTestCribbageSquares()
	require.NoError(t, c.Place(0, 0))
	c.Reset()
	assert.Equal(t, 0, c.GetPlacedCount())
	assert.False(t, c.CanUndo())
	assert.Empty(t, c.GetActionLog())
	assert.Nil(t, c.GetStarter())
}

// --- Place ---

func TestCribbageSquares_Place(t *testing.T) {
	c := newTestCribbageSquares()
	card := c.GetCurrentCard()

	require.NoError(t, c.Place(1, 2))
	assert.Same(t, card, c.GetBoard()[1][2])
	assert.Equal(t, 1, c.GetPlacedCount())
	assert.NotNil(t, c.GetCurrentCard(), "the next card is drawn straight away")
	assert.True(t, c.CanUndo())
}

func TestCribbageSquares_Place_Rejections(t *testing.T) {
	c := newTestCribbageSquares()

	for _, tc := range []struct {
		name     string
		row, col int
	}{
		{"row below zero", -1, 0},
		{"row past the grid", CribbageSquaresGridSize, 0},
		{"col below zero", 0, -1},
		{"col past the grid", 0, CribbageSquaresGridSize},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := c.Place(tc.row, tc.col)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid cell position")
		})
	}

	require.NoError(t, c.Place(0, 0))
	err := c.Place(0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already occupied")

	c.GiveUp()
	err = c.Place(2, 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not in playing phase")
}

func TestCribbageSquares_Place_RejectsWithoutACurrentCard(t *testing.T) {
	c := newTestCribbageSquares()
	c.currentCard = nil
	err := c.Place(0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no current card")
}

// The starter is the whole reason this is a cribbage game rather than a
// 16-card arrangement puzzle: #5272 leaves it out entirely.
func TestCribbageSquares_StarterIsTurnedOnlyWhenTheBoardIsFull(t *testing.T) {
	c := newTestCribbageSquares()

	for i := range CribbageSquaresTotalCells - 1 {
		require.NoError(t, c.Place(i/CribbageSquaresGridSize, i%CribbageSquaresGridSize))
		assert.Nil(t, c.GetStarter(), "still hidden after %d placements", i+1)
		assert.Equal(t, CribbageSquaresPhasePlaying, c.GetPhase())
	}

	require.NoError(t, c.Place(CribbageSquaresGridSize-1, CribbageSquaresGridSize-1))
	assert.NotNil(t, c.GetStarter(), "the 17th card is turned once the 16th is placed")
	assert.Equal(t, CribbageSquaresPhaseComplete, c.GetPhase())
	assert.Nil(t, c.GetCurrentCard())
	assert.True(t, c.IsComplete())
	assert.True(t, c.GetGameEndFlag())
}

func TestCribbageSquares_StarterIsNotAlreadyOnTheBoard(t *testing.T) {
	// The 17th card comes off the same deck, so it can never duplicate a placed
	// card. Run several deals -- one deal proves nothing about a shuffle.
	for range 20 {
		c := newTestCribbageSquares()
		fillCribbageSquares(t, c)
		starter := c.GetStarter()
		require.NotNil(t, starter)
		for _, row := range c.GetBoard() {
			for _, cell := range row {
				assert.False(t, cell.GetDesign() == starter.GetDesign() && cell.GetValue() == starter.GetValue(),
					"starter duplicates a board card")
			}
		}
	}
}

func TestCribbageSquares_PlaceLogsTheStarter(t *testing.T) {
	c := newTestCribbageSquares()
	fillCribbageSquares(t, c)
	log := c.GetActionLog()
	require.NotEmpty(t, log)
	assert.Equal(t, "starter", log[len(log)-1].ActionType)
	assert.Len(t, log[len(log)-1].Cards, 1)
}

// --- Scoring ---

// A row of 5-5-5-J with a 5 starter is cribbage's canonical 29-hand shape;
// here the four 5s and the Jack make the maximum a single line can hold.
func TestCribbageSquares_RowScore_TheTwentyNineHand(t *testing.T) {
	c := newTestCribbageSquares()
	var board [CribbageSquaresGridSize][CribbageSquaresGridSize]*Card
	board[0] = [CribbageSquaresGridSize]*Card{
		NewCard(CardDesignSpade, 5, true),
		NewCard(CardDesignClover, 5, true),
		NewCard(CardDesignDiamond, 5, true),
		NewCard(CardDesignHeart, CribbageJackValue, true),
	}
	setCribbageBoard(c, board, NewCard(CardDesignHeart, 5, true))

	detail := c.RowDetail(0)
	assert.Equal(t, 29, detail.Total, "eight fifteens, four of a kind, and nobs")
	assert.Equal(t, 16, detail.Fifteens)
	assert.Equal(t, 12, detail.Pairs)
	assert.Equal(t, 1, detail.Nobs, "the Jack matches the starter's suit")
	assert.Equal(t, 29, c.RowScore(0))
}

func TestCribbageSquares_ColScore(t *testing.T) {
	c := newTestCribbageSquares()
	var board [CribbageSquaresGridSize][CribbageSquaresGridSize]*Card
	// Column 0 is a 4-card run in one suit: 4-5-6-7 of spades.
	for i, v := range []int{4, 5, 6, 7} {
		board[i][0] = NewCard(CardDesignSpade, v, true)
	}
	setCribbageBoard(c, board, NewCard(CardDesignSpade, 8, true))

	detail := c.ColDetail(0)
	// 4-5-6-7-8 is a five-card run (5) plus 7+8 and 4+5+6 make fifteen (4),
	// plus a five-card flush (5).
	assert.Equal(t, 5, detail.Runs)
	assert.Equal(t, 4, detail.Fifteens)
	assert.Equal(t, 5, detail.Flush, "four in the line plus a matching starter")
	assert.Equal(t, 14, c.ColScore(0))
}

// A four-card flush scores 4 on its own, and 5 when the starter matches.
func TestCribbageSquares_Flush(t *testing.T) {
	c := newTestCribbageSquares()
	var board [CribbageSquaresGridSize][CribbageSquaresGridSize]*Card
	for i, v := range []int{2, 4, 9, CribbageJackValue} {
		board[0][i] = NewCard(CardDesignClover, v, true)
	}

	setCribbageBoard(c, board, NewCard(CardDesignSpade, 7, true))
	assert.Equal(t, 4, c.RowDetail(0).Flush, "the line alone is a flush")

	setCribbageBoard(c, board, NewCard(CardDesignClover, 7, true))
	assert.Equal(t, 5, c.RowDetail(0).Flush, "the starter matches too")
}

// An unfilled line scores nothing, and so does a full board with no starter --
// scoring four cards would report a number the game never awards.
func TestCribbageSquares_IncompleteLinesScoreZero(t *testing.T) {
	c := newTestCribbageSquares()
	assert.Equal(t, 0, c.RowScore(0))
	assert.Equal(t, 0, c.ColScore(0))
	assert.Equal(t, CribbageScoreDetail{}, c.RowDetail(0))
	assert.Equal(t, 0, c.TotalScore())

	var board [CribbageSquaresGridSize][CribbageSquaresGridSize]*Card
	for i, v := range []int{5, 5, 5, CribbageJackValue} {
		board[0][i] = NewCard(CardDesignSpade, v, true)
	}
	setCribbageBoard(c, board, nil)
	assert.Equal(t, 0, c.RowScore(0), "no starter means no score yet")
}

func TestCribbageSquares_RowColCards_OutOfRange(t *testing.T) {
	c := newTestCribbageSquares()
	assert.Nil(t, c.RowCards(-1))
	assert.Nil(t, c.RowCards(CribbageSquaresGridSize))
	assert.Nil(t, c.ColCards(-1))
	assert.Nil(t, c.ColCards(CribbageSquaresGridSize))
	assert.Equal(t, 0, c.RowScore(-1))
	assert.Equal(t, 0, c.ColScore(CribbageSquaresGridSize))
}

func TestCribbageSquares_TotalScoreSumsEightLines(t *testing.T) {
	c := newTestCribbageSquares()
	var board [CribbageSquaresGridSize][CribbageSquaresGridSize]*Card
	// Every card a 5 except one Jack, so both rows and columns score heavily.
	for r := range CribbageSquaresGridSize {
		for col := range CribbageSquaresGridSize {
			board[r][col] = NewCard(r, 5, true)
		}
	}
	setCribbageBoard(c, board, NewCard(CardDesignHeart, 5, true))

	want := 0
	for i := range CribbageSquaresGridSize {
		want += c.RowScore(i) + c.ColScore(i)
	}
	assert.Equal(t, want, c.TotalScore())
	assert.Positive(t, c.TotalScore())
}

func TestCribbageSquares_IsWin(t *testing.T) {
	c := newTestCribbageSquares()
	var board [CribbageSquaresGridSize][CribbageSquaresGridSize]*Card
	for r := range CribbageSquaresGridSize {
		for col := range CribbageSquaresGridSize {
			board[r][col] = NewCard(r, 5, true)
		}
	}
	setCribbageBoard(c, board, NewCard(CardDesignHeart, 5, true))
	assert.GreaterOrEqual(t, c.TotalScore(), CribbageSquaresWinScore)
	assert.True(t, c.IsWin())

	// Negative control: an empty board is nowhere near the target.
	empty := newTestCribbageSquares()
	assert.False(t, empty.IsWin())
	assert.Equal(t, 61, CribbageSquaresWinScore)
}

// --- Hint ---

func TestCribbageSquares_GetHint_PrefersTheCellThatScores(t *testing.T) {
	c := newTestCribbageSquares()
	// Row 0 already holds a 5; the current card is a 10, so pairing them into
	// row 0 makes fifteen. Every other cell is empty and gains nothing.
	c.board[0][0] = NewCard(CardDesignSpade, 5, true)
	c.placedCount = 1
	c.currentCard = NewCard(CardDesignHeart, 10, true)

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, 0, h.Row)
	assert.Positive(t, h.Score)
	assert.True(t, h.Synergy)
	assert.NotEqual(t, 0, h.Col, "cell (0,0) is taken")
}

func TestCribbageSquares_GetHint_NoSynergyOnAnEmptyBoard(t *testing.T) {
	c := newTestCribbageSquares()
	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, 0, h.Score)
	assert.False(t, h.Synergy, "nothing to combine with yet")
}

func TestCribbageSquares_GetHint_NilOutsidePlay(t *testing.T) {
	c := newTestCribbageSquares()
	c.GiveUp()
	assert.Nil(t, c.GetHint())

	c2 := newTestCribbageSquares()
	c2.currentCard = nil
	assert.Nil(t, c2.GetHint())
}

func TestCribbageSquares_LineGain(t *testing.T) {
	five := NewCard(CardDesignSpade, 5, true)
	ten := NewCard(CardDesignHeart, 10, true)

	// A ten onto a five makes fifteen: 2 points.
	assert.Equal(t, 2, cribbageSquaresLineGain([]*Card{five}, ten))
	// A second five is a pair: 2 points.
	assert.Equal(t, 2, cribbageSquaresLineGain([]*Card{five}, NewCard(CardDesignHeart, 5, true)))
	// Nothing in common.
	assert.Equal(t, 0, cribbageSquaresLineGain([]*Card{five}, NewCard(CardDesignHeart, 8, true)))
	assert.Equal(t, 0, cribbageSquaresLineGain(nil, ten))
}

// The heuristic must not count flush or nobs: both need a full line and the
// starter, so crediting them mid-board would point at cells that pay nothing.
func TestCribbageSquares_PartialScoreExcludesFlushAndNobs(t *testing.T) {
	sameSuit := []*Card{
		NewCard(CardDesignSpade, 2, true),
		NewCard(CardDesignSpade, 4, true),
		NewCard(CardDesignSpade, 9, true),
		NewCard(CardDesignSpade, CribbageJackValue, true),
	}
	// 2+4+9 = 15 is the only combination here; a flush would add 4 more.
	assert.Equal(t, 2, cribbageSquaresPartialScore(sameSuit))
	assert.Equal(t, 0, cribbageSquaresPartialScore(nil))
}

// --- Undo ---

func TestCribbageSquares_Undo(t *testing.T) {
	c := newTestCribbageSquares()
	first := c.GetCurrentCard()
	require.NoError(t, c.Place(2, 3))
	require.NoError(t, c.Undo())

	assert.Nil(t, c.GetBoard()[2][3])
	assert.Equal(t, 0, c.GetPlacedCount())
	assert.Same(t, first, c.GetCurrentCard(), "the same card comes back, not a fresh draw")
	assert.False(t, c.CanUndo())
	assert.Empty(t, c.GetActionLog())
}

func TestCribbageSquares_Undo_Rejections(t *testing.T) {
	c := newTestCribbageSquares()
	err := c.Undo()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no history")
}

// Undoing the final placement has to hide the starter again, or the player
// keeps the knowledge they were not meant to have while replacing the card.
func TestCribbageSquares_Undo_HidesTheStarterAgain(t *testing.T) {
	c := newTestCribbageSquares()
	fillCribbageSquares(t, c)
	require.NotNil(t, c.GetStarter())

	require.NoError(t, c.Undo())
	assert.Nil(t, c.GetStarter(), "the starter goes back face-down")
	assert.Equal(t, CribbageSquaresPhasePlaying, c.GetPhase())
	assert.Equal(t, CribbageSquaresTotalCells-1, c.GetPlacedCount())
	assert.NotNil(t, c.GetCurrentCard(), "the 16th card is back in hand")
	assert.False(t, c.IsComplete())
}

func TestCribbageSquares_Undo_ReturnsCardsToTheDeck(t *testing.T) {
	c := newTestCribbageSquares()
	require.NoError(t, c.Place(0, 0))
	require.NoError(t, c.Place(0, 1))
	drawnBefore := c.trumpCards.deckDrawCnt

	require.NoError(t, c.Undo())
	assert.Less(t, c.trumpCards.deckDrawCnt, drawnBefore, "the drawn card is put back")
}

// --- GiveUp ---

func TestCribbageSquares_GiveUp(t *testing.T) {
	c := newTestCribbageSquares()
	c.GiveUp()
	assert.Equal(t, CribbageSquaresPhaseComplete, c.GetPhase())
	assert.Nil(t, c.GetCurrentCard())
	assert.True(t, c.GetGameEndFlag())
	assert.Nil(t, c.GetStarter(), "an abandoned board never turns the starter")
	require.NotEmpty(t, c.GetActionLog())
	assert.Equal(t, "giveup", c.GetActionLog()[len(c.GetActionLog())-1].ActionType)

	before := len(c.GetActionLog())
	c.GiveUp()
	assert.Len(t, c.GetActionLog(), before, "a second give-up adds nothing")
}

// --- JSON round-trip ---

func TestCribbageSquares_JSONRoundTrip(t *testing.T) {
	c := newTestCribbageSquares()
	require.NoError(t, c.Place(0, 0))
	require.NoError(t, c.Place(1, 1))

	data, err := json.Marshal(c)
	require.NoError(t, err)

	restored := NewDefaultCribbageSquares()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, c.GetPlacedCount(), restored.GetPlacedCount())
	assert.Equal(t, c.GetPhase(), restored.GetPhase())
	assert.Equal(t, c.GetBoard()[0][0].GetValue(), restored.GetBoard()[0][0].GetValue())
	assert.Equal(t, c.GetCurrentCard().GetValue(), restored.GetCurrentCard().GetValue())
}

func TestCribbageSquares_JSONRoundTripKeepsTheStarter(t *testing.T) {
	c := newTestCribbageSquares()
	fillCribbageSquares(t, c)
	want := c.GetStarter()

	data, err := json.Marshal(c)
	require.NoError(t, err)
	restored := NewDefaultCribbageSquares()
	require.NoError(t, json.Unmarshal(data, restored))

	require.NotNil(t, restored.GetStarter())
	assert.Equal(t, want.GetValue(), restored.GetStarter().GetValue())
	assert.Equal(t, want.GetDesign(), restored.GetStarter().GetDesign())
	assert.Equal(t, c.TotalScore(), restored.TotalScore(), "the score survives the trip")
}

// The undo stack has to survive the trip: the Worker rebuilds the game from KV
// on every request, so an unpersisted history means Undo silently never works.
func TestCribbageSquares_JSONRoundTripKeepsUndoHistory(t *testing.T) {
	c := newTestCribbageSquares()
	require.NoError(t, c.Place(0, 0))
	require.NoError(t, c.Place(0, 1))

	data, err := json.Marshal(c)
	require.NoError(t, err)
	restored := NewDefaultCribbageSquares()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.True(t, restored.CanUndo())
	require.NoError(t, restored.Undo())
	assert.Nil(t, restored.GetBoard()[0][1], "the snapshot carried the board, not a blank")
	assert.Equal(t, 1, restored.GetPlacedCount())
}

func TestCribbageSquares_UnmarshalJSON_Rejections(t *testing.T) {
	for _, tc := range []struct{ name, data string }{
		{"broken json", `{`},
		{"phase too low", `{"ps":-1}`},
		{"phase too high", `{"ps":99}`},
		{"negative placed count", `{"pc":-1}`},
		{"placed count past the grid", `{"pc":17}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Error(t, json.Unmarshal([]byte(tc.data), NewDefaultCribbageSquares()))
		})
	}
}

func TestCribbageSquares_UnmarshalJSON_RejectsOversizedArrays(t *testing.T) {
	t.Run("action log", func(t *testing.T) {
		j := cribbageSquaresJSON{ActionLog: make([]*ActionLogEntry, cribbageSquaresMaxSliceLen+1)}
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultCribbageSquares()))
	})
	t.Run("history", func(t *testing.T) {
		j := cribbageSquaresJSON{History: make([]*cribbageSquaresSnapshot, cribbageSquaresMaxSliceLen+1)}
		for i := range j.History {
			j.History[i] = &cribbageSquaresSnapshot{}
		}
		data, err := json.Marshal(j)
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultCribbageSquares()))
	})
}

// Undo() slices the deck and the action log with these two numbers, so an
// out-of-range value from a crafted KV payload would panic the worker.
func TestCribbageSquares_SnapshotUnmarshalJSON_Rejections(t *testing.T) {
	for _, tc := range []struct {
		name string
		j    cribbageSquaresSnapshotJSON
	}{
		{"deck draw count below zero", cribbageSquaresSnapshotJSON{DeckDrawCnt: -1}},
		{"deck draw count past the deck", cribbageSquaresSnapshotJSON{DeckDrawCnt: CardCnt + 1}},
		{"action log length below zero", cribbageSquaresSnapshotJSON{ActionLogLn: -1}},
		{"action log length past the cap", cribbageSquaresSnapshotJSON{ActionLogLn: cribbageSquaresMaxSliceLen + 1}},
		{"placed count past the grid", cribbageSquaresSnapshotJSON{PlacedCount: CribbageSquaresTotalCells + 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.j)
			require.NoError(t, err)
			var s cribbageSquaresSnapshot
			assert.Error(t, s.UnmarshalJSON(data))
		})
	}

	t.Run("broken json", func(t *testing.T) {
		var s cribbageSquaresSnapshot
		assert.Error(t, s.UnmarshalJSON([]byte(`{`)))
	})

	// Negative control: a snapshot inside the bounds is accepted.
	t.Run("valid snapshot", func(t *testing.T) {
		data, err := json.Marshal(cribbageSquaresSnapshotJSON{DeckDrawCnt: 3, ActionLogLn: 2, PlacedCount: 2})
		require.NoError(t, err)
		var s cribbageSquaresSnapshot
		require.NoError(t, s.UnmarshalJSON(data))
		assert.Equal(t, 3, s.deckDrawCnt)
		assert.Equal(t, 2, s.placedCount)
	})
}

func TestCribbageSquares_UnmarshalJSON_FillsNilCollections(t *testing.T) {
	restored := NewDefaultCribbageSquares()
	require.NoError(t, json.Unmarshal([]byte(`{"ps":0,"pc":0}`), restored))
	assert.NotNil(t, restored.GetActionLog())
	assert.NotNil(t, restored.history)
	assert.NotNil(t, restored.trumpCards, "a missing deck is replaced rather than left nil")
}
