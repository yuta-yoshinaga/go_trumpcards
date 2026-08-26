//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeCardStal(design, value int) *Card {
	return NewCard(design, value, false)
}

func newTestStalactites() *Stalactites {
	return NewStalactites(NewTrumpCards(0))
}

// setupPlayingStalactites deals and then PINS the base rank to Ace.
//
// The base rank is normally taken from the first stalactite, so it varies per
// deal -- fine for the game, useless for a test about move mechanics, which
// would otherwise pass or fail depending on the shuffle. Pinning it to Ace also
// keeps the FreeCell-derived tests below meaningful: with base = Ace the
// foundation sequence is A,2,3,... exactly as FreeCell's were written.
// TestStalactitesReset covers the real, deal-derived base rank.
func setupPlayingStalactites() *Stalactites {
	f := newTestStalactites()
	f.Reset()
	f.baseRank = 1
	return f
}

// clearCellsStal empties the four stalactite cells.
//
// Stalactites deals INTO the cells, so unlike FreeCell there is no free cell at
// the start. Any test about the cell-move mechanic has to make room first; the
// "cells begin full" property itself is asserted in TestStalactitesReset.
func clearCellsStal(f *Stalactites) {
	for i := 0; i < StalactitesCellCnt; i++ {
		f.cells[i] = nil
	}
}

func clearTableauFCStal(f *Stalactites) {
	for i := 0; i < StalactitesTableauCnt; i++ {
		f.tableau[i] = nil
	}
}

// --- Reset tests ---

func TestStalactitesReset(t *testing.T) {
	f := newTestStalactites()
	f.Reset()

	assert.Equal(t, StalactitesPhasePlaying, f.GetPhase())
	assert.Equal(t, 0, f.GetMoveCount())

	// **4 stalactites + 8 columns of 6 = 52.** FreeCell, which this domain was
	// cloned from, starts with EMPTY cells and deals 7/7/7/7/6/6/6/6. The issue
	// text ("4 revealed + the remaining 52 dealt") does not add up against a
	// 52-card deck; see the issue comment for how that was resolved.
	cells := f.GetCells()
	filled := 0
	for i := 0; i < StalactitesCellCnt; i++ {
		if cells[i] != nil {
			filled++
		}
	}
	assert.Equal(t, StalactitesCellCnt, filled, "every cell starts occupied")

	tableau := f.GetTableau()
	total := filled
	for i := 0; i < StalactitesTableauCnt; i++ {
		assert.Equal(t, 6, len(tableau[i]), "column %d holds 6 cards", i)
		total += len(tableau[i])
	}
	assert.Equal(t, 52, total, "4 stalactites + 48 tableau cards is the whole deck")

	// Foundations start empty and every one of them begins at the base rank,
	// which is taken from the first stalactite rather than fixed at Ace.
	foundation := f.GetFoundation()
	for i := 0; i < StalactitesFoundationCnt; i++ {
		assert.Empty(t, foundation[i])
	}
	assert.Equal(t, cells[0].GetValue(), f.GetBaseRank(),
		"base rank is the rank of the first stalactite")
}

// TestStalactitesSolverUsesBaseRank pins the stalemate solver to the SAME rules
// the game plays by.
//
// The solver kept FreeCell's model after the clone -- pile chosen by suit, empty
// pile takes only an Ace -- and nothing noticed, because every solver test
// seeded foundation state directly and so agreed with the solver's own wrong
// model. checkStalemate() runs after almost every move and feeds IsStalemate(),
// the stalemate message and the undo-to-escape count, so with a base rank other
// than Ace (12 deals in 13) the player was told about a different game.
//
// The board below is one move from won under Stalactites' rules: each pile holds
// 12 cards, and with base rank 7 the 13th card each wants is a 6. Under
// FreeCell's model a 12-card pile wants a King, so the same board reads as
// unwinnable -- which is exactly the wrong answer this test exists to catch.
func TestStalactitesSolverUsesBaseRank(t *testing.T) {
	f := newTestStalactites()
	f.Reset()
	clearTableauFCStal(f)
	clearCellsStal(f)
	f.baseRank = 7

	var foundation [StalactitesFoundationCnt][]*Card
	for i := range StalactitesFoundationCnt {
		foundation[i] = make([]*Card, 12)
		for j := range foundation[i] {
			foundation[i][j] = makeCardStal(CardDesignSpade, stalactitesRankAtOffset(7, j))
		}
	}
	f.SetFoundation(foundation)

	// The four remaining cards are the 6s each pile now wants.
	var cells [StalactitesCellCnt]*Card
	for i, d := range []int{CardDesignSpade, CardDesignHeart, CardDesignClover, CardDesignDiamond} {
		cells[i] = makeCardStal(d, 6)
	}
	f.SetCells(cells)

	f.checkStalemate()
	assert.False(t, f.IsStalemate(),
		"every remaining card is playable at base rank 7; only FreeCell's model calls this dead")
}

// --- foundation: suit-agnostic, ascending, wrapping from the base rank ---

// setBaseAndClear puts the deck into a known state: a chosen base rank, empty
// foundations, empty cells and an empty tableau, so a single placement exercises
// exactly canPlaceOnFoundation.
func setBaseAndClear(f *Stalactites, base int) {
	f.Reset()
	clearTableauFCStal(f)
	clearCellsStal(f)
	for i := 0; i < StalactitesCellCnt; i++ {
		f.cells[i] = nil
	}
	for i := 0; i < StalactitesFoundationCnt; i++ {
		f.foundation[i] = nil
	}
	f.baseRank = base
}

func TestStalactitesFoundationTakesOnlyTheBaseRankWhenEmpty(t *testing.T) {
	f := newTestStalactites()
	setBaseAndClear(f, 7)

	assert.True(t, f.canPlaceOnFoundation(makeCardStal(CardDesignSpade, 7), 0),
		"an empty pile takes the base rank")
	// FreeCell would take only an Ace here; Stalactites must not.
	assert.False(t, f.canPlaceOnFoundation(makeCardStal(CardDesignSpade, 1), 0),
		"an Ace is not special once the base rank is 7")
	assert.False(t, f.canPlaceOnFoundation(makeCardStal(CardDesignSpade, 8), 0),
		"the rank above the base is not a valid start")
}

func TestStalactitesFoundationIgnoresSuit(t *testing.T) {
	f := newTestStalactites()
	setBaseAndClear(f, 7)
	f.foundation[0] = []*Card{makeCardStal(CardDesignSpade, 7)}

	// **All four suits are accepted.** FreeCell requires the same suit, so this
	// is the divergence that a clone would silently get wrong.
	for _, d := range []int{CardDesignSpade, CardDesignHeart, CardDesignClover, CardDesignDiamond} {
		assert.True(t, f.canPlaceOnFoundation(makeCardStal(d, 8), 0),
			"design %d should be accepted; suit is ignored", d)
	}
	// Negative control: ignoring suit must not mean ignoring rank.
	assert.False(t, f.canPlaceOnFoundation(makeCardStal(CardDesignHeart, 9), 0))
	assert.False(t, f.canPlaceOnFoundation(makeCardStal(CardDesignHeart, 7), 0))
	assert.False(t, f.canPlaceOnFoundation(makeCardStal(CardDesignHeart, 6), 0))
}

func TestStalactitesFoundationWrapsPastKing(t *testing.T) {
	f := newTestStalactites()
	setBaseAndClear(f, 12)
	f.foundation[0] = []*Card{makeCardStal(CardDesignSpade, 13)}

	assert.True(t, f.canPlaceOnFoundation(makeCardStal(CardDesignHeart, 1), 0),
		"Ace follows King when the base rank is not Ace")
	assert.False(t, f.canPlaceOnFoundation(makeCardStal(CardDesignHeart, 2), 0))
}

func TestStalactitesResetClearsHistory(t *testing.T) {
	f := setupPlayingStalactites()
	// actionLogとhistoryが初期化されること
	assert.Nil(t, f.GetActionLog())
}

// --- MoveTableauToTableau tests ---

func TestStalactitesMoveTableauToTableau(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// 列0にK♠、列1にQ♥を配置 → Q♥をK♠の上に移動可能
	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 13)}
	f.tableau[1] = []*Card{makeCardStal(CardDesignHeart, 12)}

	err := f.MoveTableauToTableau(1, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(f.tableau[0]))
	assert.Equal(t, 0, len(f.tableau[1]))
	assert.Equal(t, 1, f.GetMoveCount())
}

func TestStalactitesMoveTableauToTableauSupermove(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// 列0にK♠、列1にQ♥→J♠のシーケンス → 2枚移動
	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 13)}
	f.tableau[1] = []*Card{
		makeCardStal(CardDesignHeart, 12),
		makeCardStal(CardDesignSpade, 11),
	}

	err := f.MoveTableauToTableau(1, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(f.tableau[0]))
}

func TestStalactitesMoveTableauToTableauTopCard(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// cardIndex=-1 は一番上のカードを移動
	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 13)}
	f.tableau[1] = []*Card{makeCardStal(CardDesignHeart, 12)}

	err := f.MoveTableauToTableau(1, -1, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(f.tableau[0]))
	assert.Equal(t, 0, len(f.tableau[1]))
}

func TestStalactitesMoveTableauToTableauTopCardEmptyColumn(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// cardIndex=-1 で空の列からは移動できない
	err := f.MoveTableauToTableau(0, -1, 1)
	assert.Error(t, err)
}

func TestStalactitesMoveTableauToTableauKingToEmpty(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 13)}

	err := f.MoveTableauToTableau(0, 0, 1)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(f.tableau[0]))
	assert.Equal(t, 1, len(f.tableau[1]))
}

func TestStalactitesMoveTableauToTableauErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		f := setupPlayingStalactites()
		f.SetPhase(StalactitesPhaseGameOver)
		err := f.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("invalid from col negative", func(t *testing.T) {
		f := setupPlayingStalactites()
		err := f.MoveTableauToTableau(-1, 0, 1)
		assert.Error(t, err)
	})

	t.Run("invalid from col too large", func(t *testing.T) {
		f := setupPlayingStalactites()
		err := f.MoveTableauToTableau(8, 0, 1)
		assert.Error(t, err)
	})

	t.Run("invalid to col negative", func(t *testing.T) {
		f := setupPlayingStalactites()
		err := f.MoveTableauToTableau(0, 0, -1)
		assert.Error(t, err)
	})

	t.Run("invalid to col too large", func(t *testing.T) {
		f := setupPlayingStalactites()
		err := f.MoveTableauToTableau(0, 0, 8)
		assert.Error(t, err)
	})

	t.Run("same column", func(t *testing.T) {
		f := setupPlayingStalactites()
		err := f.MoveTableauToTableau(0, 0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid card index negative", func(t *testing.T) {
		f := setupPlayingStalactites()
		// -1 is a valid shortcut for "last card"; use -2 for truly invalid negative index
		err := f.MoveTableauToTableau(0, -2, 1)
		assert.Error(t, err)
	})

	t.Run("invalid card index too large", func(t *testing.T) {
		f := setupPlayingStalactites()
		clearTableauFCStal(f)
		clearCellsStal(f)
		f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 1)}
		err := f.MoveTableauToTableau(0, 5, 1)
		assert.Error(t, err)
	})

	t.Run("invalid sequence", func(t *testing.T) {
		f := setupPlayingStalactites()
		clearTableauFCStal(f)
		clearCellsStal(f)
		// 同色の連続は不正
		f.tableau[0] = []*Card{
			makeCardStal(CardDesignSpade, 10),
			makeCardStal(CardDesignClover, 9),
		}
		f.tableau[1] = []*Card{makeCardStal(CardDesignHeart, 11)}
		err := f.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("too many cards", func(t *testing.T) {
		f := setupPlayingStalactites()
		clearTableauFCStal(f)
		clearCellsStal(f)
		// フリーセル全部埋める
		for i := 0; i < StalactitesCellCnt; i++ {
			f.cells[i] = makeCardStal(CardDesignSpade, 1)
		}
		// 有効な3枚シーケンス, でもmax=1 (0 empty cells, 0 empty tableau excluding target)
		// Actually: empty cells=0, all other cols are empty -> 2^7 = 128... need to fill some
		// Let me fill all other tableau cols too
		for i := 2; i < StalactitesTableauCnt; i++ {
			f.tableau[i] = []*Card{makeCardStal(CardDesignSpade, 1)}
		}
		f.tableau[0] = []*Card{
			makeCardStal(CardDesignHeart, 5),
			makeCardStal(CardDesignSpade, 4),
			makeCardStal(CardDesignHeart, 3),
		}
		f.tableau[1] = []*Card{makeCardStal(CardDesignClover, 6)}
		// max = (1+0) * 2^0 = 1, trying to move 3 cards
		err := f.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("cannot place on tableau", func(t *testing.T) {
		f := setupPlayingStalactites()
		clearTableauFCStal(f)
		clearCellsStal(f)
		// 赤→赤は置けない
		f.tableau[0] = []*Card{makeCardStal(CardDesignHeart, 5)}
		f.tableau[1] = []*Card{makeCardStal(CardDesignDiamond, 6)}
		err := f.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("non-king to empty succeeds", func(t *testing.T) {
		// フリーセルでは空列に任意のカードを置けるため、非Kingの移動も成立する
		f := setupPlayingStalactites()
		clearTableauFCStal(f)
		clearCellsStal(f)
		f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 5)}
		err := f.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(f.tableau[0]))
		assert.Equal(t, 1, len(f.tableau[1]))
	})
}

// --- MoveTableauToFoundation tests ---

func TestStalactitesMoveTableauToFoundation(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// A♠をファンデーションへ
	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 1)}
	err := f.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(f.tableau[0]))
	assert.Equal(t, 1, len(f.foundation[0])) // Spade = design 1, index 0
}

func TestStalactitesMoveTableauToFoundationSequence(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	f.foundation[0] = []*Card{makeCardStal(CardDesignSpade, 1)}
	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 2)}
	err := f.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(f.foundation[0]))
}

func TestStalactitesMoveTableauToFoundationErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		f := setupPlayingStalactites()
		f.SetPhase(StalactitesPhaseGameOver)
		err := f.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("invalid column negative", func(t *testing.T) {
		f := setupPlayingStalactites()
		err := f.MoveTableauToFoundation(-1)
		assert.Error(t, err)
	})

	t.Run("invalid column too large", func(t *testing.T) {
		f := setupPlayingStalactites()
		err := f.MoveTableauToFoundation(8)
		assert.Error(t, err)
	})

	t.Run("empty column", func(t *testing.T) {
		f := setupPlayingStalactites()
		clearTableauFCStal(f)
		clearCellsStal(f)
		err := f.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("invalid card for foundation (joker)", func(t *testing.T) {
		f := setupPlayingStalactites()
		clearTableauFCStal(f)
		clearCellsStal(f)
		f.tableau[0] = []*Card{makeCardStal(CardDesignJoker, 0)}
		err := f.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("cannot place on foundation", func(t *testing.T) {
		f := setupPlayingStalactites()
		clearTableauFCStal(f)
		clearCellsStal(f)
		f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 5)}
		err := f.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})
}

// --- MoveTableauToStalactites tests ---

func TestStalactitesMoveTableauToStalactites(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 5)}
	err := f.MoveTableauToStalactites(0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(f.tableau[0]))
	assert.NotNil(t, f.cells[0])
	assert.Equal(t, 5, f.cells[0].GetValue())
}

func TestStalactitesMoveTableauToStalactitesErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		f := setupPlayingStalactites()
		f.SetPhase(StalactitesPhaseGameOver)
		err := f.MoveTableauToStalactites(0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid column negative", func(t *testing.T) {
		f := setupPlayingStalactites()
		err := f.MoveTableauToStalactites(-1, 0)
		assert.Error(t, err)
	})

	t.Run("invalid column too large", func(t *testing.T) {
		f := setupPlayingStalactites()
		err := f.MoveTableauToStalactites(8, 0)
		assert.Error(t, err)
	})

	t.Run("invalid cell negative", func(t *testing.T) {
		f := setupPlayingStalactites()
		err := f.MoveTableauToStalactites(0, -1)
		assert.Error(t, err)
	})

	t.Run("invalid cell too large", func(t *testing.T) {
		f := setupPlayingStalactites()
		err := f.MoveTableauToStalactites(0, 4)
		assert.Error(t, err)
	})

	t.Run("empty column", func(t *testing.T) {
		f := setupPlayingStalactites()
		clearTableauFCStal(f)
		clearCellsStal(f)
		err := f.MoveTableauToStalactites(0, 0)
		assert.Error(t, err)
	})

	t.Run("cell occupied", func(t *testing.T) {
		f := setupPlayingStalactites()
		f.cells[0] = makeCardStal(CardDesignSpade, 1)
		err := f.MoveTableauToStalactites(0, 0)
		assert.Error(t, err)
	})
}

// --- MoveStalactitesToTableau tests ---

func TestStalactitesMoveStalactitesToTableau(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	f.cells[0] = makeCardStal(CardDesignSpade, 13)
	err := f.MoveStalactitesToTableau(0, 0)
	assert.NoError(t, err)
	assert.Nil(t, f.cells[0])
	assert.Equal(t, 1, len(f.tableau[0]))
}

func TestStalactitesMoveStalactitesToTableauOnCard(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 6)}
	f.cells[0] = makeCardStal(CardDesignHeart, 5)
	err := f.MoveStalactitesToTableau(0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(f.tableau[0]))
}

func TestStalactitesMoveStalactitesToTableauErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		f := setupPlayingStalactites()
		f.SetPhase(StalactitesPhaseGameOver)
		err := f.MoveStalactitesToTableau(0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid cell negative", func(t *testing.T) {
		f := setupPlayingStalactites()
		err := f.MoveStalactitesToTableau(-1, 0)
		assert.Error(t, err)
	})

	t.Run("invalid cell too large", func(t *testing.T) {
		f := setupPlayingStalactites()
		err := f.MoveStalactitesToTableau(4, 0)
		assert.Error(t, err)
	})

	t.Run("invalid col negative", func(t *testing.T) {
		f := setupPlayingStalactites()
		err := f.MoveStalactitesToTableau(0, -1)
		assert.Error(t, err)
	})

	t.Run("invalid col too large", func(t *testing.T) {
		f := setupPlayingStalactites()
		err := f.MoveStalactitesToTableau(0, 8)
		assert.Error(t, err)
	})

	t.Run("empty cell", func(t *testing.T) {
		// **セルを実際に空にすること。** Stalactites は**セルに配る**ので、
		// 配った直後のセルは埋まっている。空にせずに呼ぶと、通っていたのは
		// 「セルが空」ではなく「その札はその列に置けない」からで、置ける配りが
		// 出た瞬間に落ちる（パッケージ全体を走らせたときだけ再現した）。
		f := setupPlayingStalactites()
		clearCellsStal(f)
		require.Nil(t, f.cells[0], "セルが空になっていない")

		err := f.MoveStalactitesToTableau(0, 0)
		assert.Error(t, err)

		// **負のコントロール**: セルに札を戻し、置ける列を作れば通ること。
		// これが無いと「常にエラー」でも上の assert は通る。
		clearTableauFCStal(f)
		f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 6)}
		f.cells[0] = makeCardStal(CardDesignHeart, 5) // 異色の 1 つ下
		assert.NoError(t, f.MoveStalactitesToTableau(0, 0))
	})

	t.Run("cannot place on tableau", func(t *testing.T) {
		f := setupPlayingStalactites()
		clearTableauFCStal(f)
		clearCellsStal(f)
		f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 6)}
		f.cells[0] = makeCardStal(CardDesignClover, 5) // same color
		err := f.MoveStalactitesToTableau(0, 0)
		assert.Error(t, err)
	})
}

// --- MoveStalactitesToFoundation tests ---

func TestStalactitesMoveStalactitesToFoundation(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	f.cells[0] = makeCardStal(CardDesignSpade, 1)
	err := f.MoveStalactitesToFoundation(0)
	assert.NoError(t, err)
	assert.Nil(t, f.cells[0])
	assert.Equal(t, 1, len(f.foundation[0]))
}

func TestStalactitesMoveStalactitesToFoundationErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		f := setupPlayingStalactites()
		f.SetPhase(StalactitesPhaseGameOver)
		err := f.MoveStalactitesToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("invalid cell negative", func(t *testing.T) {
		f := setupPlayingStalactites()
		err := f.MoveStalactitesToFoundation(-1)
		assert.Error(t, err)
	})

	t.Run("invalid cell too large", func(t *testing.T) {
		f := setupPlayingStalactites()
		err := f.MoveStalactitesToFoundation(4)
		assert.Error(t, err)
	})

	t.Run("empty cell", func(t *testing.T) {
		f := setupPlayingStalactites()
		// Stalactites DEALS INTO the cells, so Reset leaves cell 0 occupied --
		// this only errored when the dealt card happened to be unplayable, i.e.
		// it failed roughly one run in thirteen. Empty it to test what the name
		// says.
		f.cells[0] = nil
		err := f.MoveStalactitesToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("invalid card for foundation (joker)", func(t *testing.T) {
		f := setupPlayingStalactites()
		f.cells[0] = makeCardStal(CardDesignJoker, 0)
		err := f.MoveStalactitesToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("cannot place on foundation", func(t *testing.T) {
		f := setupPlayingStalactites()
		f.cells[0] = makeCardStal(CardDesignSpade, 5)
		err := f.MoveStalactitesToFoundation(0)
		assert.Error(t, err)
	})
}

// --- GiveUp tests ---

func TestStalactitesGiveUp(t *testing.T) {
	f := setupPlayingStalactites()
	f.GiveUp()
	assert.Equal(t, StalactitesPhaseGameOver, f.GetPhase())
	assert.Equal(t, 1, len(f.GetActionLog()))
}

func TestStalactitesGiveUpNotPlaying(t *testing.T) {
	f := setupPlayingStalactites()
	f.SetPhase(StalactitesPhaseGameClear)
	f.GiveUp()
	assert.Equal(t, StalactitesPhaseGameClear, f.GetPhase())
}

// --- GetHint tests ---

func TestStalactitesGetHintNotPlaying(t *testing.T) {
	f := setupPlayingStalactites()
	f.SetPhase(StalactitesPhaseGameOver)
	assert.Nil(t, f.GetHint())
}

func TestStalactitesGetHintTableauToFoundation(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	f.tableau[2] = []*Card{makeCardStal(CardDesignHeart, 1)}
	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, 2, hint.FromCol)
	assert.Equal(t, "foundation", hint.ToZone)
}

func TestStalactitesGetHintStalactitesToFoundation(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	f.cells[1] = makeCardStal(CardDesignClover, 1)
	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "stalactites", hint.FromZone)
	assert.Equal(t, 1, hint.FromCol)
	assert.Equal(t, "foundation", hint.ToZone)
}

func TestStalactitesGetHintTableauToTableau(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 6)}
	f.tableau[1] = []*Card{makeCardStal(CardDesignHeart, 5)}
	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, "tableau", hint.ToZone)
}

func TestStalactitesGetHintStalactitesToTableau(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// フリーセルからタブローへ（空列へのKingではない配置）
	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 6)}
	f.cells[0] = makeCardStal(CardDesignHeart, 5)
	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "stalactites", hint.FromZone)
	assert.Equal(t, "tableau", hint.ToZone)
}

func TestStalactitesGetHintTableauToStalactites(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)
	// フリーセル3つ埋めて1つ空けておく（ファンデーションに置けない値にする）
	f.cells[0] = makeCardStal(CardDesignSpade, 5)
	f.cells[1] = makeCardStal(CardDesignClover, 6)
	f.cells[2] = makeCardStal(CardDesignHeart, 7)
	// タブロー: ファンデーションに置けず、タブロー間移動もできない
	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 9)}
	// 他のタブロー列を埋めて空列→Kingヒントを抑制
	for i := 1; i < StalactitesTableauCnt; i++ {
		f.tableau[i] = []*Card{makeCardStal(CardDesignSpade, 2)}
	}
	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, "stalactites", hint.ToZone)
	assert.Equal(t, 3, hint.ToCol) // cell 3が空
}

func TestStalactitesGetHintStalactitesKingToEmptyTableau(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// King in stalactites should be hinted to move to empty tableau column
	f.cells[0] = makeCardStal(CardDesignSpade, 13)
	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "stalactites", hint.FromZone)
	assert.Equal(t, 0, hint.FromCol)
	assert.Equal(t, "tableau", hint.ToZone)
}

func TestStalactitesGetHintNoHint(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)
	// 全フリーセル埋める
	for i := 0; i < StalactitesCellCnt; i++ {
		f.cells[i] = makeCardStal(CardDesignSpade, 7+i)
	}
	// タブロー: ファンデーションに置けない、タブロー間移動もできない
	// 全列に同色のカードで移動不能にする
	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 5)}
	f.tableau[1] = []*Card{makeCardStal(CardDesignSpade, 4)}
	f.tableau[2] = []*Card{makeCardStal(CardDesignSpade, 3)}
	f.tableau[3] = []*Card{makeCardStal(CardDesignSpade, 2)}
	f.tableau[4] = []*Card{makeCardStal(CardDesignClover, 5)}
	f.tableau[5] = []*Card{makeCardStal(CardDesignClover, 4)}
	f.tableau[6] = []*Card{makeCardStal(CardDesignClover, 3)}
	f.tableau[7] = []*Card{makeCardStal(CardDesignClover, 2)}
	// フリーセルからタブローへも不能（同色）
	hint := f.GetHint()
	assert.Nil(t, hint)
}

func TestStalactitesGetHintTableauToTableauKingToEmpty(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// King can move to empty column
	f.tableau[0] = []*Card{
		makeCardStal(CardDesignHeart, 5),
		makeCardStal(CardDesignSpade, 13),
	}
	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, "tableau", hint.ToZone)
	assert.Equal(t, 1, hint.CardIndex) // King at index 1
}

func TestStalactitesGetHintTableauToTableauSequenceMaxCards(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// Fill all free cells with cards that cannot be placed on any tableau
	f.cells[0] = makeCardStal(CardDesignSpade, 10)
	f.cells[1] = makeCardStal(CardDesignSpade, 11)
	f.cells[2] = makeCardStal(CardDesignClover, 10)
	f.cells[3] = makeCardStal(CardDesignClover, 11)
	// Fill most tableau columns
	for i := 2; i < StalactitesTableauCnt; i++ {
		f.tableau[i] = []*Card{makeCardStal(CardDesignSpade, 2)}
	}
	// Valid 2-card sequence, but max = 1 -> cannot hint
	f.tableau[0] = []*Card{
		makeCardStal(CardDesignHeart, 6),
		makeCardStal(CardDesignSpade, 5),
	}
	f.tableau[1] = []*Card{makeCardStal(CardDesignDiamond, 7)}
	// max for col 1 = (1+0)*2^0 = 1, but we need to move 2 cards
	// stalactites→tableau: 5♠ on 2♠? No, same color. Can't place.
	// tableau→stalactites: all occupied
	hint := f.GetHint()
	assert.Nil(t, hint)
}

func TestStalactitesGetHintEmptyTableauColumns(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// Empty columns exist, non-King card on one column -> hint should skip empty columns for non-King
	f.tableau[0] = []*Card{makeCardStal(CardDesignHeart, 5)}
	f.tableau[1] = []*Card{makeCardStal(CardDesignSpade, 4)}
	// col 2-7 are empty. hint should move 4♠ to 5♥ (not to empty column)
	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.ToZone)
	assert.Equal(t, 0, hint.ToCol) // move to col 0 (non-empty)
}

func TestStalactitesGetHintStalactitesNonKingToEmptyColumn(t *testing.T) {
	// フリーセルにしか置き場所がなく、かつ非Kingカードであっても空列へのヒントを返す
	// （Issue #1283: Stalactitesでは空列に任意のカードを置ける）。
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// フリーセルの5♥はタブロー上の6♥（同色）には置けない。他の列は空。
	f.tableau[0] = []*Card{makeCardStal(CardDesignHeart, 6)}
	f.cells[0] = makeCardStal(CardDesignHeart, 5)

	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "stalactites", hint.FromZone)
	assert.Equal(t, 0, hint.FromCol)
	assert.Equal(t, "tableau", hint.ToZone)
	// col 0以外の最初の空列（col 1）
	assert.Equal(t, 1, hint.ToCol)
}

func TestStalactitesGetHintTableauNonKingSequenceToEmptyColumn(t *testing.T) {
	// タブロー奥のカードを露出させる目的で、非Kingのシーケンスを空列に移動する
	// ヒントを返す（Issue #1283）。
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// col 0: 隠したい9♥の上に非Kingの有効シーケンス（4♠, 3♥）が乗っている
	f.tableau[0] = []*Card{
		makeCardStal(CardDesignHeart, 9),
		makeCardStal(CardDesignSpade, 4),
		makeCardStal(CardDesignHeart, 3),
	}
	// col 1: 他のタブロー間移動を成立させないための同色カード
	f.tableau[1] = []*Card{makeCardStal(CardDesignSpade, 10)}
	// col 2-7は空。フリーセルは空のまま。

	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, 0, hint.FromCol)
	assert.Equal(t, 1, hint.CardIndex) // 4♠の位置
	assert.Equal(t, "tableau", hint.ToZone)
	// col 2が最初の空列
	assert.Equal(t, 2, hint.ToCol)
}

func TestStalactitesGetHintSkipsWholeColumnMoveToEmpty(t *testing.T) {
	// 列全体を空列に移動するのは無意味な空列交換なので、空列への
	// フォールバックヒントには含めない。その場合はタブロー→フリーセルに
	// フォールバックする。
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// col 0には単独カード（これを空列に動かしても空列交換にしかならない）
	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 5)}
	// col 1-7は空、フリーセルは空

	hint := f.GetHint()
	assert.NotNil(t, hint)
	// 空列交換を避けて、タブロー→フリーセルが選ばれる
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, 0, hint.FromCol)
	assert.Equal(t, "stalactites", hint.ToZone)
}

// --- AutoComplete tests ---

func TestStalactitesAutoComplete(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// 全カードをファンデーションに置ける状態を作る
	for suit := 1; suit <= 4; suit++ {
		for val := 1; val <= 13; val++ {
			f.foundation[suit-1] = append(f.foundation[suit-1], makeCardStal(suit, val))
		}
	}
	// ファンデーションが全部埋まっているが、removeして一部をタブロー/フリーセルに戻す
	// シンプルに: 各スートの最後の1枚をタブローに移す
	for suit := 0; suit < 4; suit++ {
		f.foundation[suit] = f.foundation[suit][:12]
	}
	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 13)}
	f.tableau[1] = []*Card{makeCardStal(CardDesignClover, 13)}
	f.cells[0] = makeCardStal(CardDesignHeart, 13)
	f.cells[1] = makeCardStal(CardDesignDiamond, 13)

	err := f.AutoComplete()
	assert.NoError(t, err)
	assert.Equal(t, StalactitesPhaseGameClear, f.GetPhase())
}

func TestStalactitesAutoCompleteNotPlaying(t *testing.T) {
	f := setupPlayingStalactites()
	f.SetPhase(StalactitesPhaseGameOver)
	err := f.AutoComplete()
	assert.Error(t, err)
}

func TestStalactitesAutoCompletePartial(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// A♠だけファンデーションに移せる
	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 1)}
	f.tableau[1] = []*Card{makeCardStal(CardDesignHeart, 5)}

	err := f.AutoComplete()
	assert.NoError(t, err)
	assert.Equal(t, StalactitesPhasePlaying, f.GetPhase())
	assert.Equal(t, 1, len(f.foundation[0]))
}

func TestStalactitesAutoCompleteStalactitesInvalidFoundation(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// Joker in stalactites - invalid foundation index
	f.cells[0] = makeCardStal(CardDesignJoker, 0)
	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 1)}

	err := f.AutoComplete()
	assert.NoError(t, err)
	// Joker should remain in stalactites, Ace should go to foundation
	assert.NotNil(t, f.cells[0])
	assert.Equal(t, 1, len(f.foundation[0]))
}

func TestStalactitesAutoCompleteTableauInvalidFoundation(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// Joker at top of tableau - invalid foundation index
	f.tableau[0] = []*Card{makeCardStal(CardDesignJoker, 0)}

	err := f.AutoComplete()
	assert.NoError(t, err)
	// Joker should remain
	assert.Equal(t, 1, len(f.tableau[0]))
}

// --- Undo tests ---

func TestStalactitesUndo(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 1)}
	err := f.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(f.foundation[0]))

	err = f.Undo()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(f.tableau[0]))
	assert.Equal(t, 0, len(f.foundation[0]))
}

func TestStalactitesUndoNotPlaying(t *testing.T) {
	f := setupPlayingStalactites()
	f.SetPhase(StalactitesPhaseGameOver)
	err := f.Undo()
	assert.Error(t, err)
}

func TestStalactitesUndoNoHistory(t *testing.T) {
	f := setupPlayingStalactites()
	err := f.Undo()
	assert.Error(t, err)
}

func TestStalactitesCanUndo(t *testing.T) {
	f := setupPlayingStalactites()
	assert.False(t, f.CanUndo())

	clearTableauFCStal(f)
	clearCellsStal(f)
	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 1)}
	_ = f.MoveTableauToFoundation(0)
	assert.True(t, f.CanUndo())
}

func TestStalactitesCanUndoNotPlaying(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)
	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 1)}
	_ = f.MoveTableauToFoundation(0)
	f.SetPhase(StalactitesPhaseGameOver)
	assert.False(t, f.CanUndo())
}

// --- GameClear test ---

func TestStalactitesGameClear(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// ファンデーションをほぼ完成状態にして最後の1枚を移動
	for suit := 0; suit < 4; suit++ {
		for val := 1; val <= 12; val++ {
			f.foundation[suit] = append(f.foundation[suit], makeCardStal(suit+1, val))
		}
	}
	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 13)}
	f.tableau[1] = []*Card{makeCardStal(CardDesignClover, 13)}
	f.tableau[2] = []*Card{makeCardStal(CardDesignHeart, 13)}
	f.tableau[3] = []*Card{makeCardStal(CardDesignDiamond, 13)}

	_ = f.MoveTableauToFoundation(0)
	_ = f.MoveTableauToFoundation(1)
	_ = f.MoveTableauToFoundation(2)
	_ = f.MoveTableauToFoundation(3)

	assert.Equal(t, StalactitesPhaseGameClear, f.GetPhase())
}

func TestStalactitesGameClearFromStalactites(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	for suit := 0; suit < 4; suit++ {
		for val := 1; val <= 12; val++ {
			f.foundation[suit] = append(f.foundation[suit], makeCardStal(suit+1, val))
		}
	}
	f.cells[0] = makeCardStal(CardDesignSpade, 13)
	f.cells[1] = makeCardStal(CardDesignClover, 13)
	f.cells[2] = makeCardStal(CardDesignHeart, 13)
	f.cells[3] = makeCardStal(CardDesignDiamond, 13)

	_ = f.MoveStalactitesToFoundation(0)
	_ = f.MoveStalactitesToFoundation(1)
	_ = f.MoveStalactitesToFoundation(2)
	_ = f.MoveStalactitesToFoundation(3)

	assert.Equal(t, StalactitesPhaseGameClear, f.GetPhase())
}

// --- Snapshot deep copy tests ---

func TestStalactitesSnapshotDeepCopy(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 1)}
	_ = f.MoveTableauToFoundation(0)

	// Modify current state
	f.foundation[0] = nil
	f.tableau[0] = []*Card{makeCardStal(CardDesignHeart, 5)}

	// Undo should restore
	_ = f.Undo()
	assert.Equal(t, 1, len(f.tableau[0]))
	assert.Equal(t, 1, f.tableau[0][0].GetValue())
	assert.Equal(t, 0, len(f.foundation[0]))
}

// --- ActionLog tests ---

func TestStalactitesActionLog(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 1)}
	_ = f.MoveTableauToFoundation(0)

	log := f.GetActionLog()
	assert.Equal(t, 1, len(log))
	assert.Equal(t, "move", log[0].ActionType)
}

// --- maxMovableCards tests ---

func TestStalactitesMaxMovableCards(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// All cells empty, all tableau empty
	// Target col 0: empty cells=4, empty tableau cols (excluding col 0) = 7
	max := f.maxMovableCards(0)
	assert.Equal(t, (1+4)*(1<<7), max) // 5 * 128 = 640

	// One stalactites occupied, one tableau col occupied
	f.cells[0] = makeCardStal(CardDesignSpade, 1)
	f.tableau[1] = []*Card{makeCardStal(CardDesignSpade, 2)}
	max = f.maxMovableCards(0)
	// empty cells=3, empty tableau (excl 0) = 6
	assert.Equal(t, (1+3)*(1<<6), max) // 4 * 64 = 256

	// Target is empty: should not count target as empty
	max = f.maxMovableCards(2) // col 2 is empty
	// empty cells=3, empty tableau (excl 2): cols 0,3,4,5,6,7 are empty (6), col 1 occupied → 6 empty
	assert.Equal(t, (1+3)*(1<<6), max) // 4 * 64 = 256
}

// --- isValidTableauSequence tests ---

func TestStalactitesIsValidTableauSequence(t *testing.T) {
	f := newTestStalactites()

	t.Run("single card", func(t *testing.T) {
		assert.True(t, f.isValidTableauSequence([]*Card{makeCardStal(CardDesignSpade, 5)}))
	})

	t.Run("valid sequence", func(t *testing.T) {
		cards := []*Card{
			makeCardStal(CardDesignSpade, 5),
			makeCardStal(CardDesignHeart, 4),
			makeCardStal(CardDesignClover, 3),
		}
		assert.True(t, f.isValidTableauSequence(cards))
	})

	t.Run("invalid same color", func(t *testing.T) {
		cards := []*Card{
			makeCardStal(CardDesignSpade, 5),
			makeCardStal(CardDesignClover, 4),
		}
		assert.False(t, f.isValidTableauSequence(cards))
	})

	t.Run("invalid not descending", func(t *testing.T) {
		cards := []*Card{
			makeCardStal(CardDesignSpade, 5),
			makeCardStal(CardDesignHeart, 5),
		}
		assert.False(t, f.isValidTableauSequence(cards))
	})
}

// --- isAlternateColor / isBlack tests ---

func TestStalactitesIsAlternateColor(t *testing.T) {
	f := newTestStalactites()

	assert.True(t, f.isAlternateColor(makeCardStal(CardDesignSpade, 1), makeCardStal(CardDesignHeart, 1)))
	assert.True(t, f.isAlternateColor(makeCardStal(CardDesignClover, 1), makeCardStal(CardDesignDiamond, 1)))
	assert.False(t, f.isAlternateColor(makeCardStal(CardDesignSpade, 1), makeCardStal(CardDesignClover, 1)))
	assert.False(t, f.isAlternateColor(makeCardStal(CardDesignHeart, 1), makeCardStal(CardDesignDiamond, 1)))
}

func TestStalactitesIsBlack(t *testing.T) {
	f := newTestStalactites()

	assert.True(t, f.isBlack(makeCardStal(CardDesignSpade, 1)))
	assert.True(t, f.isBlack(makeCardStal(CardDesignClover, 1)))
	assert.False(t, f.isBlack(makeCardStal(CardDesignHeart, 1)))
	assert.False(t, f.isBlack(makeCardStal(CardDesignDiamond, 1)))
}

// --- canPlaceOnTableau tests ---

func TestStalactitesCanPlaceOnTableau(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	t.Run("king on empty", func(t *testing.T) {
		assert.True(t, f.canPlaceOnTableau(makeCardStal(CardDesignSpade, 13), 0))
	})

	t.Run("non-king on empty", func(t *testing.T) {
		// フリーセルでは空列に任意のカードを置ける（クロンダイクと異なる）
		assert.True(t, f.canPlaceOnTableau(makeCardStal(CardDesignSpade, 5), 0))
	})

	t.Run("alternate color descending", func(t *testing.T) {
		f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 6)}
		assert.True(t, f.canPlaceOnTableau(makeCardStal(CardDesignHeart, 5), 0))
	})

	t.Run("same color", func(t *testing.T) {
		f.tableau[1] = []*Card{makeCardStal(CardDesignSpade, 6)}
		assert.False(t, f.canPlaceOnTableau(makeCardStal(CardDesignClover, 5), 1))
	})

	t.Run("not descending", func(t *testing.T) {
		f.tableau[2] = []*Card{makeCardStal(CardDesignSpade, 6)}
		assert.False(t, f.canPlaceOnTableau(makeCardStal(CardDesignHeart, 6), 2))
	})
}

// --- canPlaceOnFoundation tests ---

func TestStalactitesCanPlaceOnFoundation(t *testing.T) {
	f := setupPlayingStalactites()

	t.Run("ace on empty", func(t *testing.T) {
		assert.True(t, f.canPlaceOnFoundation(makeCardStal(CardDesignSpade, 1), 0))
	})

	t.Run("non-ace on empty", func(t *testing.T) {
		assert.False(t, f.canPlaceOnFoundation(makeCardStal(CardDesignSpade, 5), 0))
	})

	t.Run("same suit ascending", func(t *testing.T) {
		f.foundation[0] = []*Card{makeCardStal(CardDesignSpade, 1)}
		assert.True(t, f.canPlaceOnFoundation(makeCardStal(CardDesignSpade, 2), 0))
	})

	// **Stalactites ignores suit.** FreeCell, which this was cloned from,
	// rejected a different suit here; that assertion was inverted rather than
	// deleted so the divergence stays visible.
	t.Run("different suit is ACCEPTED", func(t *testing.T) {
		f.foundation[1] = []*Card{makeCardStal(CardDesignClover, 1)}
		assert.True(t, f.canPlaceOnFoundation(makeCardStal(CardDesignSpade, 2), 1))
	})

	t.Run("not ascending", func(t *testing.T) {
		f.foundation[2] = []*Card{makeCardStal(CardDesignHeart, 1)}
		assert.False(t, f.canPlaceOnFoundation(makeCardStal(CardDesignHeart, 3), 2))
	})
}

// --- UndoToEscape / UndoN tests ---

func TestStalactitesUndoToEscape_NotInStalemate(t *testing.T) {
	f := setupPlayingStalactites()
	assert.Equal(t, 0, f.UndoToEscape())
}

func TestStalactitesUndoToEscape_StalemateNoHistory(t *testing.T) {
	f := setupPlayingStalactites()
	f.SetIsStalemate(true)
	assert.Equal(t, -1, f.UndoToEscape())
}

func TestStalactitesUndoToEscape_StalemateWithEscape(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// Make a move to create a non-stalemate history entry
	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 1)}
	err := f.MoveTableauToFoundation(0)
	assert.NoError(t, err)

	// Now set stalemate
	f.SetIsStalemate(true)
	n := f.UndoToEscape()
	assert.Equal(t, 1, n)
}

func TestStalactitesUndoToEscape_AllHistoryStalemate(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// Make a move, then set stalemate on the game, make another move
	f.SetIsStalemate(true)
	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 1)}
	err := f.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	// History snapshot was taken while isStalemate was true

	f.SetIsStalemate(true)
	assert.Equal(t, -1, f.UndoToEscape())
}

func TestStalactitesUndoN_Zero(t *testing.T) {
	f := setupPlayingStalactites()
	err := f.UndoN(0)
	assert.NoError(t, err)
}

func TestStalactitesUndoN_Valid(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	// Make two moves
	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 1)}
	_ = f.MoveTableauToFoundation(0)
	f.tableau[1] = []*Card{makeCardStal(CardDesignHeart, 1)}
	_ = f.MoveTableauToFoundation(1)

	err := f.UndoN(2)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(f.foundation[0]))
	assert.Equal(t, 0, len(f.foundation[1]))
}

func TestStalactitesUndoN_Excessive(t *testing.T) {
	f := setupPlayingStalactites()
	clearTableauFCStal(f)
	clearCellsStal(f)

	f.tableau[0] = []*Card{makeCardStal(CardDesignSpade, 1)}
	_ = f.MoveTableauToFoundation(0)

	err := f.UndoN(5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undo step")
}

// **何枚まとめて動かせるかは CUI に出ていなかった (#4777)。**Web は
// fc-supermove-limit で常時出している。
func TestStalactites_GetMaxMovableCards(t *testing.T) {
	card := func(v int) *Card { return NewCard(CardDesignSpade, v, false) }
	board := func(filledCells int, filledCols int) *Stalactites {
		f := NewStalactites(NewTrumpCards(0))
		f.Reset()
		var cells [StalactitesCellCnt]*Card
		for i := 0; i < filledCells && i < StalactitesCellCnt; i++ {
			cells[i] = card(i + 2)
		}
		f.SetCells(cells)
		var tableau [StalactitesTableauCnt][]*Card
		for i := 0; i < StalactitesTableauCnt; i++ {
			if i < filledCols {
				tableau[i] = []*Card{card(5)}
			}
		}
		f.SetTableau(tableau)
		return f
	}

	// **(1 + 空きセル) × 2^(空き列)。**セルは足し算、列は掛け算。
	t.Run("counts free cells additively and empty columns as doubling", func(t *testing.T) {
		// セル4つ空き・列は全部埋まり → (1+4) << 0 = 5。
		assert.Equal(t, 5, board(0, StalactitesTableauCnt).GetMaxMovableCards())
		// セル4つ空き・列1つ空き → (1+4) << 1 = 10。
		assert.Equal(t, 10, board(0, StalactitesTableauCnt-1).GetMaxMovableCards())
		// セル4つ空き・列2つ空き → (1+4) << 2 = 20。
		assert.Equal(t, 20, board(0, StalactitesTableauCnt-2).GetMaxMovableCards())
	})

	// **一般の上限はどの列も除外しない。**特定の列を除外した値を返すと、
	// 空き列が1つ少ない前提の小さすぎる数を出すことになる。
	t.Run("the general limit excludes no column, not even the first", func(t *testing.T) {
		f := NewStalactites(NewTrumpCards(0))
		f.Reset()
		f.SetCells([StalactitesCellCnt]*Card{})
		var tableau [StalactitesTableauCnt][]*Card
		// 0 列目だけを空にし、残りを埋める。
		for i := 1; i < StalactitesTableauCnt; i++ {
			tableau[i] = []*Card{card(5)}
		}
		f.SetTableau(tableau)
		assert.Equal(t, 10, f.GetMaxMovableCards(), "(1+4) << 1")
	})

	t.Run("a filled free cell lowers the limit", func(t *testing.T) {
		assert.Equal(t, 4, board(1, StalactitesTableauCnt).GetMaxMovableCards())
	})

	t.Run("with everything full only a single card moves", func(t *testing.T) {
		assert.Equal(t, 1, board(StalactitesCellCnt, StalactitesTableauCnt).GetMaxMovableCards())
	})

	// **空き列を移動先にすると上限は下がる。**その列自身を経由地に使えない。
	t.Run("moving onto an empty column halves the limit", func(t *testing.T) {
		f := board(0, StalactitesTableauCnt-1)
		assert.Equal(t, 10, f.GetMaxMovableCards())
		assert.Equal(t, 5, f.GetMaxMovableCardsToEmptyColumn())
	})

	t.Run("reports zero when there is no empty column to move onto", func(t *testing.T) {
		assert.Equal(t, 0, board(0, StalactitesTableauCnt).GetMaxMovableCardsToEmptyColumn())
	})
}
