//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestBisley() *Bisley {
	b := NewBisley(NewTrumpCards(0))
	b.Reset()
	return b
}

// setBoard installs an exact position. Tests must never lean on the shuffle:
// asserting what is or is not playable against a real deal is a flake waiting
// to happen (see #4467).
func setBoard(b *Bisley, aces, kings [BisleyFoundationCnt][]*Card, cols [][]*Card) {
	b.aceFoundations = aces
	b.kingFoundations = kings
	for i := range BisleyTableauCnt {
		b.tableau[i] = nil
	}
	for i, c := range cols {
		pile := make([]*BisleyTableauCard, 0, len(c))
		for _, card := range c {
			pile = append(pile, &BisleyTableauCard{Card: card, FaceUp: true})
		}
		b.tableau[i] = pile
	}
	b.phase = BisleyPhasePlaying
	b.isStalemate = false
}

func TestBisley_Reset_DealsAcesUpAndFortyEightToTableau(t *testing.T) {
	b := newTestBisley()

	assert.Equal(t, BisleyPhasePlaying, b.GetPhase())
	assert.Equal(t, 0, b.GetMoveCount())
	assert.False(t, b.GetGameEndFlag())

	// Only the four Aces are seeded. The King foundations open later, when a
	// King is moved up -- the issue's "place the Kings too" reading contradicts
	// its own "48 cards remain" (52-8 = 44).
	total := 0
	for i, f := range b.GetAceFoundations() {
		require.Len(t, f, 1, "ace foundation %d", i)
		assert.Equal(t, 1, f[0].GetValue())
		total += len(f)
	}
	for i, f := range b.GetKingFoundations() {
		assert.Empty(t, f, "king foundation %d starts empty", i)
	}
	for _, col := range b.GetTableau() {
		total += len(col)
	}
	assert.Equal(t, CardCnt, total, "every card is accounted for")

	// 13 columns: four of three cards, nine of four.
	threes, fours := 0, 0
	for _, col := range b.GetTableau() {
		switch len(col) {
		case 3:
			threes++
		case 4:
			fours++
		}
	}
	assert.Equal(t, 4, threes, "four short columns")
	assert.Equal(t, 9, fours, "nine full columns")
	assert.True(t, b.AllFaceUp(), "Bisley is an open game: nothing is hidden")
}

func TestBisley_AceFoundationBuildsUpInSuit(t *testing.T) {
	b := newTestBisley()
	var aces, kings [BisleyFoundationCnt][]*Card
	aces[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	setBoard(b, aces, kings, [][]*Card{{NewCard(CardDesignSpade, 2, true)}})

	require.NoError(t, b.MoveTableauToAceFoundation(0))
	assert.Len(t, b.GetAceFoundations()[0], 2)
	assert.Empty(t, b.GetTableau()[0])
}

func TestBisley_AceFoundationRejectsOffSuitAndWrongRank(t *testing.T) {
	b := newTestBisley()
	var aces, kings [BisleyFoundationCnt][]*Card
	aces[0] = []*Card{NewCard(CardDesignSpade, 1, true)}

	setBoard(b, aces, kings, [][]*Card{{NewCard(CardDesignHeart, 2, true)}})
	require.Error(t, b.MoveTableauToAceFoundation(0), "suit must match")

	setBoard(b, aces, kings, [][]*Card{{NewCard(CardDesignSpade, 3, true)}})
	require.Error(t, b.MoveTableauToAceFoundation(0), "must be exactly one higher")
}

func TestBisley_KingFoundationOpensOnKingAndBuildsDown(t *testing.T) {
	b := newTestBisley()
	var aces, kings [BisleyFoundationCnt][]*Card
	setBoard(b, aces, kings, [][]*Card{{NewCard(CardDesignHeart, CardValueMax, true)}})

	require.NoError(t, b.MoveTableauToKingFoundation(0), "a King opens its own pile")
	require.Len(t, b.GetKingFoundations()[BisleySuitIndex(CardDesignHeart)], 1)

	setBoard(b, b.GetAceFoundations(), b.GetKingFoundations(), [][]*Card{{NewCard(CardDesignHeart, 12, true)}})
	require.NoError(t, b.MoveTableauToKingFoundation(0), "builds down by one, same suit")
	assert.Len(t, b.GetKingFoundations()[BisleySuitIndex(CardDesignHeart)], 2)
}

func TestBisley_KingFoundationRejectsNonKingOnEmptyPile(t *testing.T) {
	b := newTestBisley()
	var aces, kings [BisleyFoundationCnt][]*Card
	setBoard(b, aces, kings, [][]*Card{{NewCard(CardDesignHeart, 12, true)}})

	require.Error(t, b.MoveTableauToKingFoundation(0), "only a King may open a descending pile")
}

func TestBisley_TableauMovesSameSuitEitherDirection(t *testing.T) {
	b := newTestBisley()
	var aces, kings [BisleyFoundationCnt][]*Card
	// The real rule, and the reason the game is not trivial: a tableau top may
	// only sit on a same-suit neighbour one rank away, up or down.
	setBoard(b, aces, kings, [][]*Card{
		{NewCard(CardDesignSpade, 7, true)},
		{NewCard(CardDesignSpade, 8, true)},
		{NewCard(CardDesignHeart, 8, true)},
		{NewCard(CardDesignSpade, 10, true)},
	})

	require.NoError(t, b.MoveTableauToTableau(0, 1), "7 onto 8 of the same suit")
	assert.Len(t, b.GetTableau()[1], 2)

	require.Error(t, b.MoveTableauToTableau(1, 2), "different suit")
	require.Error(t, b.MoveTableauToTableau(1, 3), "two ranks apart")
}

func TestBisley_TableauMoveToEmptyColumnIsRejected(t *testing.T) {
	b := newTestBisley()
	var aces, kings [BisleyFoundationCnt][]*Card
	setBoard(b, aces, kings, [][]*Card{{NewCard(CardDesignSpade, 7, true)}, {}})

	require.Error(t, b.MoveTableauToTableau(0, 1),
		"an empty column is not a free parking space in Bisley")
}

func TestBisley_InvalidIndexesAndEmptyPiles(t *testing.T) {
	b := newTestBisley()
	var aces, kings [BisleyFoundationCnt][]*Card
	setBoard(b, aces, kings, [][]*Card{{}})

	assert.Error(t, b.MoveTableauToTableau(-1, 0))
	assert.Error(t, b.MoveTableauToTableau(0, BisleyTableauCnt))
	assert.Error(t, b.MoveTableauToTableau(0, 0), "a column cannot move onto itself")
	assert.Error(t, b.MoveTableauToAceFoundation(-1))
	assert.Error(t, b.MoveTableauToAceFoundation(0), "empty column")
	assert.Error(t, b.MoveTableauToKingFoundation(BisleyTableauCnt))
}

func TestBisley_GameClearWhenBothFoundationsMeet(t *testing.T) {
	b := newTestBisley()
	var aces, kings [BisleyFoundationCnt][]*Card
	// Each suit split between its ascending and descending pile: A-6 up, K-7 down.
	for s, design := range bisleySuitOrder {
		up := make([]*Card, 0, 6)
		for v := 1; v <= 6; v++ {
			up = append(up, NewCard(design, v, true))
		}
		down := make([]*Card, 0, 7)
		for v := CardValueMax; v >= 7; v-- {
			down = append(down, NewCard(design, v, true))
		}
		aces[s], kings[s] = up, down
	}
	setBoard(b, aces, kings, nil)
	b.checkGameClear()

	assert.Equal(t, BisleyPhaseGameClear, b.GetPhase())
	assert.True(t, b.GetGameEndFlag())
}

func TestBisley_Hint(t *testing.T) {
	b := newTestBisley()
	var aces, kings [BisleyFoundationCnt][]*Card
	setBoard(b, aces, kings, [][]*Card{{NewCard(CardDesignSpade, 9, true)}})
	assert.Nil(t, b.GetHint(), "nothing playable")

	aces[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	setBoard(b, aces, kings, [][]*Card{{NewCard(CardDesignSpade, 2, true)}})
	h := b.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, 0, h.FromCol)
	assert.Equal(t, "ace", h.ToZone)

	b.GiveUp()
	assert.Nil(t, b.GetHint(), "no hint once the game is over")
}

func TestBisley_GiveUpAndUndo(t *testing.T) {
	b := newTestBisley()
	var aces, kings [BisleyFoundationCnt][]*Card
	aces[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	setBoard(b, aces, kings, [][]*Card{{NewCard(CardDesignSpade, 2, true)}})

	assert.False(t, b.CanUndo())
	require.NoError(t, b.MoveTableauToAceFoundation(0))
	assert.True(t, b.CanUndo())
	require.NoError(t, b.Undo())
	assert.Len(t, b.GetTableau()[0], 1, "the card is back on the tableau")
	assert.Equal(t, 0, b.GetMoveCount())

	b.GiveUp()
	assert.Equal(t, BisleyPhaseGameOver, b.GetPhase())
	before := len(b.GetActionLog())
	b.GiveUp()
	assert.Len(t, b.GetActionLog(), before, "giving up twice logs once")
}

func TestBisley_AutoCompletePlaysEverythingReachable(t *testing.T) {
	b := newTestBisley()
	var aces, kings [BisleyFoundationCnt][]*Card
	for s, design := range bisleySuitOrder {
		aces[s] = []*Card{NewCard(design, 1, true)}
	}
	cols := make([][]*Card, 0, BisleyTableauCnt)
	for _, design := range bisleySuitOrder {
		for v := CardValueMax; v >= 2; v-- {
			cols = append(cols, []*Card{NewCard(design, v, true)})
		}
	}
	setBoard(b, aces, kings, cols[:BisleyTableauCnt])

	require.NoError(t, b.AutoComplete())
	assert.Positive(t, b.GetMoveCount())
}

func TestBisley_JSONRoundTrip(t *testing.T) {
	b := newTestBisley()
	data, err := json.Marshal(b)
	require.NoError(t, err)

	restored := NewBisley(NewTrumpCards(0))
	require.NoError(t, json.Unmarshal(data, restored))
	assert.Equal(t, b.GetPhase(), restored.GetPhase())
	assert.Equal(t, b.GetMoveCount(), restored.GetMoveCount())
	assert.Len(t, restored.GetAceFoundations()[0], 1)
}

func TestBisley_UnmarshalRejectsOutOfRangeState(t *testing.T) {
	b := NewBisley(NewTrumpCards(0))
	assert.Error(t, json.Unmarshal([]byte(`{"ps":99}`), b))
	assert.Error(t, json.Unmarshal([]byte(`{"mc":-1}`), b))
	assert.Error(t, json.Unmarshal([]byte(`nope`), b))
}

func TestBisley_NewDefaultUsesAStandardDeck(t *testing.T) {
	b := NewDefaultBisley()
	b.Reset()
	total := 0
	for _, f := range b.GetAceFoundations() {
		total += len(f)
	}
	for _, col := range b.GetTableau() {
		total += len(col)
	}
	assert.Equal(t, CardCnt, total)
}

func TestBisley_UndoNAndStalemateEscape(t *testing.T) {
	b := newTestBisley()
	var aces, kings [BisleyFoundationCnt][]*Card
	aces[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	setBoard(b, aces, kings, [][]*Card{
		{NewCard(CardDesignSpade, 3, true), NewCard(CardDesignSpade, 2, true)},
	})

	assert.Error(t, b.UndoN(0), "n must be positive")
	assert.Error(t, b.UndoN(1), "no history yet")
	assert.Equal(t, 0, b.UndoToEscape(), "not stalemated")

	require.NoError(t, b.MoveTableauToAceFoundation(0))
	require.NoError(t, b.MoveTableauToAceFoundation(0))
	assert.Error(t, b.UndoN(3), "more than the history holds")
	require.NoError(t, b.UndoN(2))
	assert.Equal(t, 0, b.GetMoveCount())
	assert.Len(t, b.GetTableau()[0], 2)
}

func TestBisley_StalemateWhenNothingCanMove(t *testing.T) {
	b := newTestBisley()
	var aces, kings [BisleyFoundationCnt][]*Card
	// Two tops that neither reach a foundation nor pair with each other.
	setBoard(b, aces, kings, [][]*Card{
		{NewCard(CardDesignSpade, 5, true)},
		{NewCard(CardDesignHeart, 9, true)},
	})
	b.checkStalemate()
	assert.True(t, b.IsStalemate())
	assert.Equal(t, -1, b.UndoToEscape(), "no history to rewind into")

	// A same-suit neighbour is a legal move, so this is not a stalemate.
	setBoard(b, aces, kings, [][]*Card{
		{NewCard(CardDesignSpade, 5, true)},
		{NewCard(CardDesignSpade, 6, true)},
	})
	b.checkStalemate()
	assert.False(t, b.IsStalemate())
}

// The action-log detail is rendered verbatim in the UI, so its column numbers
// have to match what the board, the hint and the CLI show. Nothing asserted on
// the string before, which is how a 1-indexed log shipped next to a 0-indexed
// board (review on #4468).
func TestBisley_ActionLogDetailUsesZeroBasedColumns(t *testing.T) {
	b := newTestBisley()
	var aces, kings [BisleyFoundationCnt][]*Card
	aces[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	setBoard(b, aces, kings, [][]*Card{
		{},
		{NewCard(CardDesignSpade, 2, true)},
		{NewCard(CardDesignHeart, 13, true)},
		{NewCard(CardDesignHeart, 7, true)},
		{NewCard(CardDesignHeart, 6, true)},
	})

	require.NoError(t, b.MoveTableauToAceFoundation(1))
	require.NoError(t, b.MoveTableauToKingFoundation(2))
	require.NoError(t, b.MoveTableauToTableau(3, 4))

	log := b.GetActionLog()
	require.Len(t, log, 3)
	// Spade is foundation 0, heart is foundation 2 — both raw indices.
	assert.Equal(t, "タブロー列1→昇順基礎0", log[0].Detail)
	assert.Equal(t, "タブロー列2→降順基礎2", log[1].Detail)
	assert.Equal(t, "タブロー列3→タブロー列4", log[2].Detail)
}

// GetHint used to only look at the foundations, so a board with legal tableau
// moves and no foundation move reported "no hint" while not being stalemated —
// the two disagreed. It also left the "tableau" branch of every consumer dead.
func TestBisley_GetHintFallsBackToATableauMove(t *testing.T) {
	b := newTestBisley()
	var aces, kings [BisleyFoundationCnt][]*Card
	// No ace pile is open and no King is exposed, so nothing reaches a
	// foundation; but ♠5 and ♠6 are a legal tableau pair.
	setBoard(b, aces, kings, [][]*Card{
		{NewCard(CardDesignSpade, 5, true)},
		{NewCard(CardDesignSpade, 6, true)},
	})

	h := b.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "tableau", h.ToZone)
	assert.Equal(t, 0, h.FromCol)
	assert.Equal(t, 1, h.ToIdx)
	b.checkStalemate()
	assert.False(t, b.IsStalemate(), "a hint exists, so this is not a dead end")

	// The hint the player is shown must be a move the domain actually accepts.
	// (This consumes the only move, so it comes after the stalemate check.)
	require.NoError(t, b.MoveTableauToTableau(h.FromCol, h.ToIdx))
}

// A foundation move always outranks a tableau shuffle.
func TestBisley_GetHintPrefersTheFoundation(t *testing.T) {
	b := newTestBisley()
	var aces, kings [BisleyFoundationCnt][]*Card
	aces[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	setBoard(b, aces, kings, [][]*Card{
		{NewCard(CardDesignHeart, 5, true)},
		{NewCard(CardDesignHeart, 6, true)},
		{NewCard(CardDesignSpade, 2, true)},
	})

	h := b.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "ace", h.ToZone, "♠2 goes up even though ♥5/♥6 pair")
	assert.Equal(t, 2, h.FromCol)
}

// AutoComplete must never make a tableau move, or it would shuffle the board
// instead of clearing it — it drives off foundationHint, not GetHint.
func TestBisley_AutoCompleteIgnoresTableauMoves(t *testing.T) {
	b := newTestBisley()
	var aces, kings [BisleyFoundationCnt][]*Card
	setBoard(b, aces, kings, [][]*Card{
		{NewCard(CardDesignSpade, 5, true)},
		{NewCard(CardDesignSpade, 6, true)},
	})

	require.NotNil(t, b.GetHint(), "a tableau move exists")
	assert.Error(t, b.AutoComplete(), "but nothing can go to a foundation")
	assert.Equal(t, 0, b.GetMoveCount())
	assert.Len(t, b.GetTableau()[0], 1, "the board is untouched")
	assert.Len(t, b.GetTableau()[1], 1)
}

// #5553: 空列への移動は `errors.New("cannot move onto an empty column")` を返し、
// cuiErrorBlock がそれをそのまま赤字で出していた。日本語でプレイしていても
// 英語の一文が画面に混ざる。
func TestBisley_ErrorsNameI18nKeys(t *testing.T) {
	b := newTestBisley()
	b.Reset()

	// 空列への移動。片方の列を空にしてから試す。
	b.tableau[0] = nil
	err := b.MoveTableauToTableau(1, 0)
	require.Error(t, err)
	code, _ := ErrorMessageCode(err)
	assert.Equal(t, "bisley.errEmptyColumn", code)
	// **生の英文が残っていないこと。**
	assert.NotContains(t, err.Error(), "empty column")

	// 同じ列への移動。
	code, _ = ErrorMessageCode(b.MoveTableauToTableau(1, 1))
	assert.Equal(t, "bisley.errSameColumn", code)

	// 範囲外の列。
	code, _ = ErrorMessageCode(b.MoveTableauToTableau(1, BisleyTableauCnt))
	assert.Equal(t, "bisley.errBadToColumn", code)

	// 取り消せる手が無いとき。
	fresh := newTestBisley()
	fresh.Reset()
	code, _ = ErrorMessageCode(fresh.Undo())
	assert.Equal(t, "bisley.errNothingToUndo", code)

	// **プレイ中でないときも。**
	fresh.GiveUp()
	code, _ = ErrorMessageCode(fresh.MoveTableauToAceFoundation(0))
	assert.Equal(t, "bisley.errNotPlaying", code)
}
