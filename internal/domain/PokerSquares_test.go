package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newPokerSquaresForTest() *PokerSquares {
	return NewPokerSquares(NewTrumpCards(0))
}

// buildBoardFromCards builds a 5x5 board from a row-major slice of 25 cards.
func buildBoardFromCards(cards []*Card) [PokerSquaresGridSize][PokerSquaresGridSize]*Card {
	var b [PokerSquaresGridSize][PokerSquaresGridSize]*Card
	for i, c := range cards {
		if i >= PokerSquaresTotalCells {
			break
		}
		b[i/PokerSquaresGridSize][i%PokerSquaresGridSize] = c
	}
	return b
}

func TestPokerSquares_Reset(t *testing.T) {
	g := newPokerSquaresForTest()
	g.Reset()
	assert.Equal(t, PokerSquaresPhasePlaying, g.GetPhase())
	assert.Equal(t, 0, g.GetPlacedCount())
	assert.NotNil(t, g.GetCurrentCard())
	assert.False(t, g.IsComplete())
	// After reset, 1 card drawn → 51 remain.
	// We can't easily assert remaining count directly, so trust DrawCard.
}

func TestPokerSquares_PlaceSuccess(t *testing.T) {
	g := newPokerSquaresForTest()
	g.Reset()
	first := g.GetCurrentCard()
	err := g.Place(0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, g.GetPlacedCount())
	assert.Equal(t, first, g.GetBoard()[0][0])
	assert.NotNil(t, g.GetCurrentCard())
	assert.NotEqual(t, first, g.GetCurrentCard())
}

func TestPokerSquares_PlaceErrors(t *testing.T) {
	g := newPokerSquaresForTest()
	g.Reset()

	// Out of bounds
	assert.Error(t, g.Place(-1, 0))
	assert.Error(t, g.Place(0, -1))
	assert.Error(t, g.Place(5, 0))
	assert.Error(t, g.Place(0, 5))

	// Occupied cell
	assert.NoError(t, g.Place(0, 0))
	assert.Error(t, g.Place(0, 0))

	// Wrong phase
	g.SetPhase(PokerSquaresPhaseComplete)
	assert.Error(t, g.Place(1, 1))

	// No current card
	g2 := newPokerSquaresForTest()
	g2.Reset()
	g2.SetCurrentCard(nil)
	assert.Error(t, g2.Place(0, 0))
}

func TestPokerSquares_GetHint_NotPlaying(t *testing.T) {
	g := newPokerSquaresForTest()
	g.SetCurrentCard(NewCard(CardDesignSpade, 5, false))
	g.SetPhase(PokerSquaresPhaseComplete)
	assert.Nil(t, g.GetHint())
}

func TestPokerSquares_GetHint_NoCurrentCard(t *testing.T) {
	g := newPokerSquaresForTest()
	g.SetCurrentCard(nil)
	assert.Nil(t, g.GetHint())
}

func TestPokerSquares_GetHint_EmptyBoard(t *testing.T) {
	g := newPokerSquaresForTest()
	g.SetCurrentCard(NewCard(CardDesignSpade, 5, false))
	hint := g.GetHint()
	assert.NotNil(t, hint)
	// No existing cards → no synergy; the first empty cell (0,0) is suggested.
	assert.Equal(t, 0, hint.Row)
	assert.Equal(t, 0, hint.Col)
	assert.Equal(t, 0, hint.Score)
	assert.False(t, hint.Synergy)
}

func TestPokerSquares_GetHint_ValuePairSynergy(t *testing.T) {
	g := newPokerSquaresForTest()
	g.SetCurrentCard(NewCard(CardDesignSpade, 5, false))
	var b [PokerSquaresGridSize][PokerSquaresGridSize]*Card
	// A same-value card in row 0 gives every empty row-0 cell a pair bonus (+4).
	b[0][1] = NewCard(CardDesignHeart, 5, false)
	g.SetBoard(b)

	hint := g.GetHint()
	assert.NotNil(t, hint)
	assert.True(t, hint.Synergy)
	assert.Equal(t, 4, hint.Score)
	assert.Equal(t, 0, hint.Row)
	assert.Equal(t, 0, hint.Col)
}

func TestPokerSquares_GetHint_SuitAndStraightSynergy(t *testing.T) {
	g := newPokerSquaresForTest()
	g.SetCurrentCard(NewCard(CardDesignSpade, 5, false))
	var b [PokerSquaresGridSize][PokerSquaresGridSize]*Card
	// Same suit (+2) and adjacent value (+1) in row 0 → +3 for empty row-0 cells.
	b[0][1] = NewCard(CardDesignSpade, 6, false)
	g.SetBoard(b)

	hint := g.GetHint()
	assert.NotNil(t, hint)
	assert.True(t, hint.Synergy)
	assert.Equal(t, 3, hint.Score)
}

func TestPokerSquares_IsCompleteAndPhase(t *testing.T) {
	g := newPokerSquaresForTest()
	g.Reset()
	// Fill 25 cells.
	for i := range PokerSquaresTotalCells {
		err := g.Place(i/PokerSquaresGridSize, i%PokerSquaresGridSize)
		assert.NoError(t, err)
	}
	assert.True(t, g.IsComplete())
	assert.Equal(t, PokerSquaresPhaseComplete, g.GetPhase())
	assert.Nil(t, g.GetCurrentCard())
}

func TestPokerSquares_UndoRestoresState(t *testing.T) {
	g := newPokerSquaresForTest()
	g.Reset()
	first := g.GetCurrentCard()
	err := g.Place(0, 0)
	assert.NoError(t, err)
	assert.True(t, g.CanUndo())

	err = g.Undo()
	assert.NoError(t, err)
	assert.Equal(t, 0, g.GetPlacedCount())
	assert.Nil(t, g.GetBoard()[0][0])
	assert.Equal(t, first, g.GetCurrentCard())
}

func TestPokerSquares_UndoRestoresDeckAndActionLog(t *testing.T) {
	g := newPokerSquaresForTest()
	g.Reset()

	// Place all 25 cards with an undo+replace partway through.
	// Without the deckDrawCnt fix, the extra undo would eventually exhaust the deck.
	for i := 0; i < PokerSquaresTotalCells; i++ {
		r, c := i/PokerSquaresGridSize, i%PokerSquaresGridSize
		if i == 10 {
			assert.NoError(t, g.Undo())
			assert.Equal(t, 9, len(g.GetActionLog()), "actionLog must be trimmed on undo")
			// Re-place the cell that was undone.
			prevR, prevC := 9/PokerSquaresGridSize, 9%PokerSquaresGridSize
			assert.NoError(t, g.Place(prevR, prevC))
		}
		assert.NoError(t, g.Place(r, c), "place %d at (%d,%d) must not exhaust deck", i, r, c)
	}
	assert.True(t, g.IsComplete())
}

func TestPokerSquares_UndoNoHistory(t *testing.T) {
	g := newPokerSquaresForTest()
	g.Reset()
	assert.False(t, g.CanUndo())
	assert.Error(t, g.Undo())
}

func TestPokerSquares_GiveUp(t *testing.T) {
	g := newPokerSquaresForTest()
	g.Reset()
	g.GiveUp()
	assert.Equal(t, PokerSquaresPhaseComplete, g.GetPhase())
	assert.Nil(t, g.GetCurrentCard())

	// Second giveup is a no-op.
	g.GiveUp()
	assert.Equal(t, PokerSquaresPhaseComplete, g.GetPhase())
}

func TestPokerSquares_EvaluateRow_IncompleteReturnsMinus1(t *testing.T) {
	g := newPokerSquaresForTest()
	g.Reset()
	assert.Equal(t, -1, g.EvaluateRow(0))
	assert.Equal(t, -1, g.EvaluateRow(-1))
	assert.Equal(t, -1, g.EvaluateRow(5))
	assert.Equal(t, 0, g.RowScore(0))
}

func TestPokerSquares_EvaluateCol_IncompleteReturnsMinus1(t *testing.T) {
	g := newPokerSquaresForTest()
	g.Reset()
	assert.Equal(t, -1, g.EvaluateCol(0))
	assert.Equal(t, -1, g.EvaluateCol(-1))
	assert.Equal(t, -1, g.EvaluateCol(5))
	assert.Equal(t, 0, g.ColScore(0))
}

// Helper rows used throughout evaluation tests.
func rowHighCard() []*Card {
	return []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 11, false),
	}
}

func rowOnePair() []*Card {
	return []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 11, false),
	}
}

func rowTwoPair() []*Card {
	return []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignSpade, 11, false),
	}
}

func rowThreeOfAKind() []*Card {
	return []*Card{
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 11, false),
	}
}

func rowStraight() []*Card {
	return []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignClover, 6, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 8, false),
		NewCard(CardDesignSpade, 9, false),
	}
}

func rowFlush() []*Card {
	return []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignSpade, 11, false),
	}
}

func rowFullHouse() []*Card {
	return []*Card{
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 9, false),
	}
}

func rowFourOfAKind() []*Card {
	return []*Card{
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignSpade, 11, false),
	}
}

func rowStraightFlush() []*Card {
	return []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 6, false),
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignSpade, 9, false),
	}
}

func rowRoyalFlush() []*Card {
	return []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 13, false),
	}
}

func TestPokerSquares_EvaluateRow_AllHandTypes(t *testing.T) {
	tests := []struct {
		name      string
		row       []*Card
		wantRank  int
		wantScore int
	}{
		{"high card", rowHighCard(), PokerHandHighCard, 0},
		{"one pair", rowOnePair(), PokerHandOnePair, 2},
		{"two pair", rowTwoPair(), PokerHandTwoPair, 5},
		{"three of a kind", rowThreeOfAKind(), PokerHandThreeOfAKind, 10},
		{"straight", rowStraight(), PokerHandStraight, 15},
		{"flush", rowFlush(), PokerHandFlush, 20},
		{"full house", rowFullHouse(), PokerHandFullHouse, 25},
		{"four of a kind", rowFourOfAKind(), PokerHandFourOfAKind, 50},
		{"straight flush", rowStraightFlush(), PokerHandStraightFlush, 75},
		{"royal flush", rowRoyalFlush(), PokerHandRoyalFlush, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newPokerSquaresForTest()
			var b [PokerSquaresGridSize][PokerSquaresGridSize]*Card
			copy(b[0][:], tt.row)
			g.SetBoard(b)
			assert.Equal(t, tt.wantRank, g.EvaluateRow(0))
			assert.Equal(t, tt.wantScore, g.RowScore(0))
		})
	}
}

func TestPokerSquares_EvaluateCol_Basic(t *testing.T) {
	g := newPokerSquaresForTest()
	var b [PokerSquaresGridSize][PokerSquaresGridSize]*Card
	for r, card := range rowFlush() {
		b[r][0] = card
	}
	g.SetBoard(b)
	assert.Equal(t, PokerHandFlush, g.EvaluateCol(0))
	assert.Equal(t, 20, g.ColScore(0))
}

func TestPokerSquares_TotalScore(t *testing.T) {
	// Build a board where row 0 = royal flush, all others = high card, columns = mixed.
	// Simpler: a full board of all-high-card rows/cols yields total 0.
	g := newPokerSquaresForTest()
	// Use unique cards: 25 distinct (design, value) combinations.
	cards := make([]*Card, 0, 25)
	designs := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	val := 2
	for i := range 25 {
		d := designs[i%4]
		cards = append(cards, NewCard(d, val, false))
		val++
		if val > 13 {
			val = 2
		}
	}
	g.SetBoard(buildBoardFromCards(cards))
	// Just verify TotalScore is sum of row + col scores.
	expected := 0
	for i := range PokerSquaresGridSize {
		expected += g.RowScore(i) + g.ColScore(i)
	}
	assert.Equal(t, expected, g.TotalScore())
}

func TestPokerSquares_TotalScore_WithRoyalFlushRow(t *testing.T) {
	g := newPokerSquaresForTest()
	var b [PokerSquaresGridSize][PokerSquaresGridSize]*Card
	for c, card := range rowRoyalFlush() {
		b[0][c] = card
	}
	// Fill remaining rows with cards so columns are not evaluable (nil cells).
	// Actually we need all rows+cols complete for scoring - but missing rows return 0.
	g.SetBoard(b)
	// Row 0 complete (royal flush = 100). Other rows incomplete (0 each). Columns incomplete (0 each).
	assert.Equal(t, 100, g.TotalScore())
}

func TestPokerSquares_JSONRoundTrip(t *testing.T) {
	g := newPokerSquaresForTest()
	g.Reset()
	_ = g.Place(0, 0)
	_ = g.Place(0, 1)

	data, err := json.Marshal(g)
	assert.NoError(t, err)

	var g2 PokerSquares
	err = json.Unmarshal(data, &g2)
	assert.NoError(t, err)
	assert.Equal(t, g.GetPlacedCount(), g2.GetPlacedCount())
	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetBoard()[0][0].GetValue(), g2.GetBoard()[0][0].GetValue())
	assert.Equal(t, g.GetBoard()[0][0].GetDesign(), g2.GetBoard()[0][0].GetDesign())
}

func TestPokerSquares_UnmarshalEmptyActionLog(t *testing.T) {
	g := newPokerSquaresForTest()
	data, err := json.Marshal(g)
	assert.NoError(t, err)
	var g2 PokerSquares
	assert.NoError(t, json.Unmarshal(data, &g2))
	assert.NotNil(t, g2.GetActionLog())
}

func TestPokerSquares_SettersAndGetters(t *testing.T) {
	g := newPokerSquaresForTest()
	g.SetPhase(PokerSquaresPhaseComplete)
	assert.Equal(t, PokerSquaresPhaseComplete, g.GetPhase())

	c := NewCard(CardDesignSpade, 7, false)
	g.SetCurrentCard(c)
	assert.Equal(t, c, g.GetCurrentCard())

	g.SetPlacedCount(5)
	assert.Equal(t, 5, g.GetPlacedCount())
}

func TestPokerSquaresRankToScore_Unknown(t *testing.T) {
	assert.Equal(t, 0, pokerSquaresRankToScore(-1))
	// FiveOfAKind (not in primary table) should map to four-of-a-kind score.
	assert.Equal(t, 50, pokerSquaresRankToScore(PokerHandFiveOfAKind))
}

func TestPokerSquares_PartialRank(t *testing.T) {
	tests := []struct {
		name     string
		cards    []*Card
		wantRank int
	}{
		{"0 cards", []*Card{}, -1},
		{"1 card", []*Card{NewCard(CardDesignSpade, 2, false)}, PokerHandHighCard},
		{"2 cards (pair)", []*Card{NewCard(CardDesignSpade, 2, false), NewCard(CardDesignClover, 2, false)}, PokerHandOnePair},
		{"2 cards (high)", []*Card{NewCard(CardDesignSpade, 2, false), NewCard(CardDesignClover, 3, false)}, PokerHandHighCard},
		{"3 cards (three of a kind)", []*Card{NewCard(CardDesignSpade, 3, false), NewCard(CardDesignClover, 3, false), NewCard(CardDesignHeart, 3, false)}, PokerHandThreeOfAKind},
		{"3 cards (pair)", []*Card{NewCard(CardDesignSpade, 3, false), NewCard(CardDesignClover, 3, false), NewCard(CardDesignHeart, 4, false)}, PokerHandOnePair},
		{"4 cards (four of a kind)", []*Card{NewCard(CardDesignSpade, 4, false), NewCard(CardDesignClover, 4, false), NewCard(CardDesignHeart, 4, false), NewCard(CardDesignDiamond, 4, false)}, PokerHandFourOfAKind},
		{"4 cards (two pair)", []*Card{NewCard(CardDesignSpade, 2, false), NewCard(CardDesignClover, 2, false), NewCard(CardDesignHeart, 3, false), NewCard(CardDesignDiamond, 3, false)}, PokerHandTwoPair},
		{"5 cards (complete)", rowStraight(), -1}, // negative control
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newPokerSquaresForTest()
			var b [PokerSquaresGridSize][PokerSquaresGridSize]*Card

			// Setup Row 0
			copy(b[0][:], tt.cards)
			g.SetBoard(b)
			assert.Equal(t, tt.wantRank, g.PartialRowRank(0), "Row partial rank mismatch")

			// Clear and setup Col 0
			var b2 [PokerSquaresGridSize][PokerSquaresGridSize]*Card
			for i, c := range tt.cards {
				b2[i][0] = c // Setup col 0
			}
			g.SetBoard(b2)
			assert.Equal(t, tt.wantRank, g.PartialColRank(0), "Col partial rank mismatch")
		})
	}
}

func TestPokerSquares_PartialRank_OutOfBounds(t *testing.T) {
	g := newPokerSquaresForTest()
	assert.Equal(t, -1, g.PartialRowRank(-1))
	assert.Equal(t, -1, g.PartialRowRank(5))
	assert.Equal(t, -1, g.PartialColRank(-1))
	assert.Equal(t, -1, g.PartialColRank(5))
}
