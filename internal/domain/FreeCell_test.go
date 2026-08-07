//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeCard(design, value int) *Card {
	return NewCard(design, value, false)
}

func newTestFreeCell() *FreeCell {
	return NewFreeCell(NewTrumpCards(0))
}

func setupPlayingFreeCell() *FreeCell {
	f := newTestFreeCell()
	f.Reset()
	return f
}

func clearTableauFC(f *FreeCell) {
	for i := 0; i < FreeCellTableauCnt; i++ {
		f.tableau[i] = nil
	}
}

// --- Reset tests ---

func TestFreeCellReset(t *testing.T) {
	f := newTestFreeCell()
	f.Reset()

	assert.Equal(t, FreeCellPhasePlaying, f.GetPhase())
	assert.Equal(t, 0, f.GetMoveCount())

	// 最初の4列は7枚
	tableau := f.GetTableau()
	for i := 0; i < 4; i++ {
		assert.Equal(t, 7, len(tableau[i]))
	}
	// 残り4列は6枚
	for i := 4; i < 8; i++ {
		assert.Equal(t, 6, len(tableau[i]))
	}

	// フリーセルは全て空
	cells := f.GetFreeCells()
	for i := 0; i < FreeCellCellCnt; i++ {
		assert.Nil(t, cells[i])
	}

	// ファンデーションは全て空
	foundation := f.GetFoundation()
	for i := 0; i < FreeCellFoundationCnt; i++ {
		assert.Empty(t, foundation[i])
	}

	// 合計52枚
	totalCards := 0
	for i := 0; i < FreeCellTableauCnt; i++ {
		totalCards += len(tableau[i])
	}
	assert.Equal(t, 52, totalCards)
}

func TestFreeCellResetClearsHistory(t *testing.T) {
	f := setupPlayingFreeCell()
	// actionLogとhistoryが初期化されること
	assert.Nil(t, f.GetActionLog())
}

// --- MoveTableauToTableau tests ---

func TestFreeCellMoveTableauToTableau(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// 列0にK♠、列1にQ♥を配置 → Q♥をK♠の上に移動可能
	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}
	f.tableau[1] = []*Card{makeCard(CardDesignHeart, 12)}

	err := f.MoveTableauToTableau(1, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(f.tableau[0]))
	assert.Equal(t, 0, len(f.tableau[1]))
	assert.Equal(t, 1, f.GetMoveCount())
}

func TestFreeCellMoveTableauToTableauSupermove(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// 列0にK♠、列1にQ♥→J♠のシーケンス → 2枚移動
	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}
	f.tableau[1] = []*Card{
		makeCard(CardDesignHeart, 12),
		makeCard(CardDesignSpade, 11),
	}

	err := f.MoveTableauToTableau(1, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(f.tableau[0]))
}

func TestFreeCellMoveTableauToTableauTopCard(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// cardIndex=-1 は一番上のカードを移動
	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}
	f.tableau[1] = []*Card{makeCard(CardDesignHeart, 12)}

	err := f.MoveTableauToTableau(1, -1, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(f.tableau[0]))
	assert.Equal(t, 0, len(f.tableau[1]))
}

func TestFreeCellMoveTableauToTableauTopCardEmptyColumn(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// cardIndex=-1 で空の列からは移動できない
	err := f.MoveTableauToTableau(0, -1, 1)
	assert.Error(t, err)
}

func TestFreeCellMoveTableauToTableauKingToEmpty(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}

	err := f.MoveTableauToTableau(0, 0, 1)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(f.tableau[0]))
	assert.Equal(t, 1, len(f.tableau[1]))
}

func TestFreeCellMoveTableauToTableauErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		f := setupPlayingFreeCell()
		f.SetPhase(FreeCellPhaseGameOver)
		err := f.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("invalid from col negative", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveTableauToTableau(-1, 0, 1)
		assert.Error(t, err)
	})

	t.Run("invalid from col too large", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveTableauToTableau(8, 0, 1)
		assert.Error(t, err)
	})

	t.Run("invalid to col negative", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveTableauToTableau(0, 0, -1)
		assert.Error(t, err)
	})

	t.Run("invalid to col too large", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveTableauToTableau(0, 0, 8)
		assert.Error(t, err)
	})

	t.Run("same column", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveTableauToTableau(0, 0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid card index negative", func(t *testing.T) {
		f := setupPlayingFreeCell()
		// -1 is a valid shortcut for "last card"; use -2 for truly invalid negative index
		err := f.MoveTableauToTableau(0, -2, 1)
		assert.Error(t, err)
	})

	t.Run("invalid card index too large", func(t *testing.T) {
		f := setupPlayingFreeCell()
		clearTableauFC(f)
		f.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
		err := f.MoveTableauToTableau(0, 5, 1)
		assert.Error(t, err)
	})

	t.Run("invalid sequence", func(t *testing.T) {
		f := setupPlayingFreeCell()
		clearTableauFC(f)
		// 同色の連続は不正
		f.tableau[0] = []*Card{
			makeCard(CardDesignSpade, 10),
			makeCard(CardDesignClover, 9),
		}
		f.tableau[1] = []*Card{makeCard(CardDesignHeart, 11)}
		err := f.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("too many cards", func(t *testing.T) {
		f := setupPlayingFreeCell()
		clearTableauFC(f)
		// フリーセル全部埋める
		for i := 0; i < FreeCellCellCnt; i++ {
			f.freeCells[i] = makeCard(CardDesignSpade, 1)
		}
		// 有効な3枚シーケンス, でもmax=1 (0 empty cells, 0 empty tableau excluding target)
		// Actually: empty cells=0, all other cols are empty -> 2^7 = 128... need to fill some
		// Let me fill all other tableau cols too
		for i := 2; i < FreeCellTableauCnt; i++ {
			f.tableau[i] = []*Card{makeCard(CardDesignSpade, 1)}
		}
		f.tableau[0] = []*Card{
			makeCard(CardDesignHeart, 5),
			makeCard(CardDesignSpade, 4),
			makeCard(CardDesignHeart, 3),
		}
		f.tableau[1] = []*Card{makeCard(CardDesignClover, 6)}
		// max = (1+0) * 2^0 = 1, trying to move 3 cards
		err := f.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("cannot place on tableau", func(t *testing.T) {
		f := setupPlayingFreeCell()
		clearTableauFC(f)
		// 赤→赤は置けない
		f.tableau[0] = []*Card{makeCard(CardDesignHeart, 5)}
		f.tableau[1] = []*Card{makeCard(CardDesignDiamond, 6)}
		err := f.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("non-king to empty succeeds", func(t *testing.T) {
		// フリーセルでは空列に任意のカードを置けるため、非Kingの移動も成立する
		f := setupPlayingFreeCell()
		clearTableauFC(f)
		f.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
		err := f.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(f.tableau[0]))
		assert.Equal(t, 1, len(f.tableau[1]))
	})
}

// --- MoveTableauToFoundation tests ---

func TestFreeCellMoveTableauToFoundation(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// A♠をファンデーションへ
	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	err := f.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(f.tableau[0]))
	assert.Equal(t, 1, len(f.foundation[0])) // Spade = design 1, index 0
}

func TestFreeCellMoveTableauToFoundationSequence(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	f.foundation[0] = []*Card{makeCard(CardDesignSpade, 1)}
	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 2)}
	err := f.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(f.foundation[0]))
}

func TestFreeCellMoveTableauToFoundationErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		f := setupPlayingFreeCell()
		f.SetPhase(FreeCellPhaseGameOver)
		err := f.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("invalid column negative", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveTableauToFoundation(-1)
		assert.Error(t, err)
	})

	t.Run("invalid column too large", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveTableauToFoundation(8)
		assert.Error(t, err)
	})

	t.Run("empty column", func(t *testing.T) {
		f := setupPlayingFreeCell()
		clearTableauFC(f)
		err := f.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("invalid card for foundation (joker)", func(t *testing.T) {
		f := setupPlayingFreeCell()
		clearTableauFC(f)
		f.tableau[0] = []*Card{makeCard(CardDesignJoker, 0)}
		err := f.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("cannot place on foundation", func(t *testing.T) {
		f := setupPlayingFreeCell()
		clearTableauFC(f)
		f.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
		err := f.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})
}

// --- MoveTableauToFreeCell tests ---

func TestFreeCellMoveTableauToFreeCell(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
	err := f.MoveTableauToFreeCell(0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(f.tableau[0]))
	assert.NotNil(t, f.freeCells[0])
	assert.Equal(t, 5, f.freeCells[0].GetValue())
}

func TestFreeCellMoveTableauToFreeCellErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		f := setupPlayingFreeCell()
		f.SetPhase(FreeCellPhaseGameOver)
		err := f.MoveTableauToFreeCell(0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid column negative", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveTableauToFreeCell(-1, 0)
		assert.Error(t, err)
	})

	t.Run("invalid column too large", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveTableauToFreeCell(8, 0)
		assert.Error(t, err)
	})

	t.Run("invalid cell negative", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveTableauToFreeCell(0, -1)
		assert.Error(t, err)
	})

	t.Run("invalid cell too large", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveTableauToFreeCell(0, 4)
		assert.Error(t, err)
	})

	t.Run("empty column", func(t *testing.T) {
		f := setupPlayingFreeCell()
		clearTableauFC(f)
		err := f.MoveTableauToFreeCell(0, 0)
		assert.Error(t, err)
	})

	t.Run("cell occupied", func(t *testing.T) {
		f := setupPlayingFreeCell()
		f.freeCells[0] = makeCard(CardDesignSpade, 1)
		err := f.MoveTableauToFreeCell(0, 0)
		assert.Error(t, err)
	})
}

// --- MoveFreeCellToTableau tests ---

func TestFreeCellMoveFreeCellToTableau(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	f.freeCells[0] = makeCard(CardDesignSpade, 13)
	err := f.MoveFreeCellToTableau(0, 0)
	assert.NoError(t, err)
	assert.Nil(t, f.freeCells[0])
	assert.Equal(t, 1, len(f.tableau[0]))
}

func TestFreeCellMoveFreeCellToTableauOnCard(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 6)}
	f.freeCells[0] = makeCard(CardDesignHeart, 5)
	err := f.MoveFreeCellToTableau(0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(f.tableau[0]))
}

func TestFreeCellMoveFreeCellToTableauErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		f := setupPlayingFreeCell()
		f.SetPhase(FreeCellPhaseGameOver)
		err := f.MoveFreeCellToTableau(0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid cell negative", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveFreeCellToTableau(-1, 0)
		assert.Error(t, err)
	})

	t.Run("invalid cell too large", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveFreeCellToTableau(4, 0)
		assert.Error(t, err)
	})

	t.Run("invalid col negative", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveFreeCellToTableau(0, -1)
		assert.Error(t, err)
	})

	t.Run("invalid col too large", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveFreeCellToTableau(0, 8)
		assert.Error(t, err)
	})

	t.Run("empty cell", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveFreeCellToTableau(0, 0)
		assert.Error(t, err)
	})

	t.Run("cannot place on tableau", func(t *testing.T) {
		f := setupPlayingFreeCell()
		clearTableauFC(f)
		f.tableau[0] = []*Card{makeCard(CardDesignSpade, 6)}
		f.freeCells[0] = makeCard(CardDesignClover, 5) // same color
		err := f.MoveFreeCellToTableau(0, 0)
		assert.Error(t, err)
	})
}

// --- MoveFreeCellToFoundation tests ---

func TestFreeCellMoveFreeCellToFoundation(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	f.freeCells[0] = makeCard(CardDesignSpade, 1)
	err := f.MoveFreeCellToFoundation(0)
	assert.NoError(t, err)
	assert.Nil(t, f.freeCells[0])
	assert.Equal(t, 1, len(f.foundation[0]))
}

func TestFreeCellMoveFreeCellToFoundationErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		f := setupPlayingFreeCell()
		f.SetPhase(FreeCellPhaseGameOver)
		err := f.MoveFreeCellToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("invalid cell negative", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveFreeCellToFoundation(-1)
		assert.Error(t, err)
	})

	t.Run("invalid cell too large", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveFreeCellToFoundation(4)
		assert.Error(t, err)
	})

	t.Run("empty cell", func(t *testing.T) {
		f := setupPlayingFreeCell()
		err := f.MoveFreeCellToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("invalid card for foundation (joker)", func(t *testing.T) {
		f := setupPlayingFreeCell()
		f.freeCells[0] = makeCard(CardDesignJoker, 0)
		err := f.MoveFreeCellToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("cannot place on foundation", func(t *testing.T) {
		f := setupPlayingFreeCell()
		f.freeCells[0] = makeCard(CardDesignSpade, 5)
		err := f.MoveFreeCellToFoundation(0)
		assert.Error(t, err)
	})
}

// --- GiveUp tests ---

func TestFreeCellGiveUp(t *testing.T) {
	f := setupPlayingFreeCell()
	f.GiveUp()
	assert.Equal(t, FreeCellPhaseGameOver, f.GetPhase())
	assert.Equal(t, 1, len(f.GetActionLog()))
}

func TestFreeCellGiveUpNotPlaying(t *testing.T) {
	f := setupPlayingFreeCell()
	f.SetPhase(FreeCellPhaseGameClear)
	f.GiveUp()
	assert.Equal(t, FreeCellPhaseGameClear, f.GetPhase())
}

// --- GetHint tests ---

func TestFreeCellGetHintNotPlaying(t *testing.T) {
	f := setupPlayingFreeCell()
	f.SetPhase(FreeCellPhaseGameOver)
	assert.Nil(t, f.GetHint())
}

func TestFreeCellGetHintTableauToFoundation(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	f.tableau[2] = []*Card{makeCard(CardDesignHeart, 1)}
	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, 2, hint.FromCol)
	assert.Equal(t, "foundation", hint.ToZone)
}

func TestFreeCellGetHintFreeCellToFoundation(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	f.freeCells[1] = makeCard(CardDesignClover, 1)
	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "freecell", hint.FromZone)
	assert.Equal(t, 1, hint.FromCol)
	assert.Equal(t, "foundation", hint.ToZone)
}

func TestFreeCellGetHintTableauToTableau(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 6)}
	f.tableau[1] = []*Card{makeCard(CardDesignHeart, 5)}
	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, "tableau", hint.ToZone)
}

func TestFreeCellGetHintFreeCellToTableau(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// フリーセルからタブローへ（空列へのKingではない配置）
	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 6)}
	f.freeCells[0] = makeCard(CardDesignHeart, 5)
	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "freecell", hint.FromZone)
	assert.Equal(t, "tableau", hint.ToZone)
}

func TestFreeCellGetHintTableauToFreeCell(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)
	// フリーセル3つ埋めて1つ空けておく（ファンデーションに置けない値にする）
	f.freeCells[0] = makeCard(CardDesignSpade, 5)
	f.freeCells[1] = makeCard(CardDesignClover, 6)
	f.freeCells[2] = makeCard(CardDesignHeart, 7)
	// タブロー: ファンデーションに置けず、タブロー間移動もできない
	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 9)}
	// 他のタブロー列を埋めて空列→Kingヒントを抑制
	for i := 1; i < FreeCellTableauCnt; i++ {
		f.tableau[i] = []*Card{makeCard(CardDesignSpade, 2)}
	}
	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, "freecell", hint.ToZone)
	assert.Equal(t, 3, hint.ToCol) // cell 3が空
}

func TestFreeCellGetHintFreeCellKingToEmptyTableau(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// King in freecell should be hinted to move to empty tableau column
	f.freeCells[0] = makeCard(CardDesignSpade, 13)
	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "freecell", hint.FromZone)
	assert.Equal(t, 0, hint.FromCol)
	assert.Equal(t, "tableau", hint.ToZone)
}

func TestFreeCellGetHintNoHint(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)
	// 全フリーセル埋める
	for i := 0; i < FreeCellCellCnt; i++ {
		f.freeCells[i] = makeCard(CardDesignSpade, 7+i)
	}
	// タブロー: ファンデーションに置けない、タブロー間移動もできない
	// 全列に同色のカードで移動不能にする
	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
	f.tableau[1] = []*Card{makeCard(CardDesignSpade, 4)}
	f.tableau[2] = []*Card{makeCard(CardDesignSpade, 3)}
	f.tableau[3] = []*Card{makeCard(CardDesignSpade, 2)}
	f.tableau[4] = []*Card{makeCard(CardDesignClover, 5)}
	f.tableau[5] = []*Card{makeCard(CardDesignClover, 4)}
	f.tableau[6] = []*Card{makeCard(CardDesignClover, 3)}
	f.tableau[7] = []*Card{makeCard(CardDesignClover, 2)}
	// フリーセルからタブローへも不能（同色）
	hint := f.GetHint()
	assert.Nil(t, hint)
}

func TestFreeCellGetHintTableauToTableauKingToEmpty(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// King can move to empty column
	f.tableau[0] = []*Card{
		makeCard(CardDesignHeart, 5),
		makeCard(CardDesignSpade, 13),
	}
	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, "tableau", hint.ToZone)
	assert.Equal(t, 1, hint.CardIndex) // King at index 1
}

func TestFreeCellGetHintTableauToTableauSequenceMaxCards(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// Fill all free cells with cards that cannot be placed on any tableau
	f.freeCells[0] = makeCard(CardDesignSpade, 10)
	f.freeCells[1] = makeCard(CardDesignSpade, 11)
	f.freeCells[2] = makeCard(CardDesignClover, 10)
	f.freeCells[3] = makeCard(CardDesignClover, 11)
	// Fill most tableau columns
	for i := 2; i < FreeCellTableauCnt; i++ {
		f.tableau[i] = []*Card{makeCard(CardDesignSpade, 2)}
	}
	// Valid 2-card sequence, but max = 1 -> cannot hint
	f.tableau[0] = []*Card{
		makeCard(CardDesignHeart, 6),
		makeCard(CardDesignSpade, 5),
	}
	f.tableau[1] = []*Card{makeCard(CardDesignDiamond, 7)}
	// max for col 1 = (1+0)*2^0 = 1, but we need to move 2 cards
	// freecell→tableau: 5♠ on 2♠? No, same color. Can't place.
	// tableau→freecell: all occupied
	hint := f.GetHint()
	assert.Nil(t, hint)
}

func TestFreeCellGetHintEmptyTableauColumns(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// Empty columns exist, non-King card on one column -> hint should skip empty columns for non-King
	f.tableau[0] = []*Card{makeCard(CardDesignHeart, 5)}
	f.tableau[1] = []*Card{makeCard(CardDesignSpade, 4)}
	// col 2-7 are empty. hint should move 4♠ to 5♥ (not to empty column)
	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.ToZone)
	assert.Equal(t, 0, hint.ToCol) // move to col 0 (non-empty)
}

func TestFreeCellGetHintFreeCellNonKingToEmptyColumn(t *testing.T) {
	// フリーセルにしか置き場所がなく、かつ非Kingカードであっても空列へのヒントを返す
	// （Issue #1283: FreeCellでは空列に任意のカードを置ける）。
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// フリーセルの5♥はタブロー上の6♥（同色）には置けない。他の列は空。
	f.tableau[0] = []*Card{makeCard(CardDesignHeart, 6)}
	f.freeCells[0] = makeCard(CardDesignHeart, 5)

	hint := f.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "freecell", hint.FromZone)
	assert.Equal(t, 0, hint.FromCol)
	assert.Equal(t, "tableau", hint.ToZone)
	// col 0以外の最初の空列（col 1）
	assert.Equal(t, 1, hint.ToCol)
}

func TestFreeCellGetHintTableauNonKingSequenceToEmptyColumn(t *testing.T) {
	// タブロー奥のカードを露出させる目的で、非Kingのシーケンスを空列に移動する
	// ヒントを返す（Issue #1283）。
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// col 0: 隠したい9♥の上に非Kingの有効シーケンス（4♠, 3♥）が乗っている
	f.tableau[0] = []*Card{
		makeCard(CardDesignHeart, 9),
		makeCard(CardDesignSpade, 4),
		makeCard(CardDesignHeart, 3),
	}
	// col 1: 他のタブロー間移動を成立させないための同色カード
	f.tableau[1] = []*Card{makeCard(CardDesignSpade, 10)}
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

func TestFreeCellGetHintSkipsWholeColumnMoveToEmpty(t *testing.T) {
	// 列全体を空列に移動するのは無意味な空列交換なので、空列への
	// フォールバックヒントには含めない。その場合はタブロー→フリーセルに
	// フォールバックする。
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// col 0には単独カード（これを空列に動かしても空列交換にしかならない）
	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
	// col 1-7は空、フリーセルは空

	hint := f.GetHint()
	assert.NotNil(t, hint)
	// 空列交換を避けて、タブロー→フリーセルが選ばれる
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, 0, hint.FromCol)
	assert.Equal(t, "freecell", hint.ToZone)
}

// --- AutoComplete tests ---

func TestFreeCellAutoComplete(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// 全カードをファンデーションに置ける状態を作る
	for suit := 1; suit <= 4; suit++ {
		for val := 1; val <= 13; val++ {
			f.foundation[suit-1] = append(f.foundation[suit-1], makeCard(suit, val))
		}
	}
	// ファンデーションが全部埋まっているが、removeして一部をタブロー/フリーセルに戻す
	// シンプルに: 各スートの最後の1枚をタブローに移す
	for suit := 0; suit < 4; suit++ {
		f.foundation[suit] = f.foundation[suit][:12]
	}
	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}
	f.tableau[1] = []*Card{makeCard(CardDesignClover, 13)}
	f.freeCells[0] = makeCard(CardDesignHeart, 13)
	f.freeCells[1] = makeCard(CardDesignDiamond, 13)

	err := f.AutoComplete()
	assert.NoError(t, err)
	assert.Equal(t, FreeCellPhaseGameClear, f.GetPhase())
}

func TestFreeCellAutoCompleteNotPlaying(t *testing.T) {
	f := setupPlayingFreeCell()
	f.SetPhase(FreeCellPhaseGameOver)
	err := f.AutoComplete()
	assert.Error(t, err)
}

func TestFreeCellAutoCompletePartial(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// A♠だけファンデーションに移せる
	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	f.tableau[1] = []*Card{makeCard(CardDesignHeart, 5)}

	err := f.AutoComplete()
	assert.NoError(t, err)
	assert.Equal(t, FreeCellPhasePlaying, f.GetPhase())
	assert.Equal(t, 1, len(f.foundation[0]))
}

func TestFreeCellAutoCompleteFreeCellInvalidFoundation(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// Joker in freecell - invalid foundation index
	f.freeCells[0] = makeCard(CardDesignJoker, 0)
	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}

	err := f.AutoComplete()
	assert.NoError(t, err)
	// Joker should remain in freecell, Ace should go to foundation
	assert.NotNil(t, f.freeCells[0])
	assert.Equal(t, 1, len(f.foundation[0]))
}

func TestFreeCellAutoCompleteTableauInvalidFoundation(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// Joker at top of tableau - invalid foundation index
	f.tableau[0] = []*Card{makeCard(CardDesignJoker, 0)}

	err := f.AutoComplete()
	assert.NoError(t, err)
	// Joker should remain
	assert.Equal(t, 1, len(f.tableau[0]))
}

// --- Undo tests ---

func TestFreeCellUndo(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	err := f.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(f.foundation[0]))

	err = f.Undo()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(f.tableau[0]))
	assert.Equal(t, 0, len(f.foundation[0]))
}

func TestFreeCellUndoNotPlaying(t *testing.T) {
	f := setupPlayingFreeCell()
	f.SetPhase(FreeCellPhaseGameOver)
	err := f.Undo()
	assert.Error(t, err)
}

func TestFreeCellUndoNoHistory(t *testing.T) {
	f := setupPlayingFreeCell()
	err := f.Undo()
	assert.Error(t, err)
}

func TestFreeCellCanUndo(t *testing.T) {
	f := setupPlayingFreeCell()
	assert.False(t, f.CanUndo())

	clearTableauFC(f)
	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	_ = f.MoveTableauToFoundation(0)
	assert.True(t, f.CanUndo())
}

func TestFreeCellCanUndoNotPlaying(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)
	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	_ = f.MoveTableauToFoundation(0)
	f.SetPhase(FreeCellPhaseGameOver)
	assert.False(t, f.CanUndo())
}

// --- GameClear test ---

func TestFreeCellGameClear(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// ファンデーションをほぼ完成状態にして最後の1枚を移動
	for suit := 0; suit < 4; suit++ {
		for val := 1; val <= 12; val++ {
			f.foundation[suit] = append(f.foundation[suit], makeCard(suit+1, val))
		}
	}
	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}
	f.tableau[1] = []*Card{makeCard(CardDesignClover, 13)}
	f.tableau[2] = []*Card{makeCard(CardDesignHeart, 13)}
	f.tableau[3] = []*Card{makeCard(CardDesignDiamond, 13)}

	_ = f.MoveTableauToFoundation(0)
	_ = f.MoveTableauToFoundation(1)
	_ = f.MoveTableauToFoundation(2)
	_ = f.MoveTableauToFoundation(3)

	assert.Equal(t, FreeCellPhaseGameClear, f.GetPhase())
}

func TestFreeCellGameClearFromFreeCell(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	for suit := 0; suit < 4; suit++ {
		for val := 1; val <= 12; val++ {
			f.foundation[suit] = append(f.foundation[suit], makeCard(suit+1, val))
		}
	}
	f.freeCells[0] = makeCard(CardDesignSpade, 13)
	f.freeCells[1] = makeCard(CardDesignClover, 13)
	f.freeCells[2] = makeCard(CardDesignHeart, 13)
	f.freeCells[3] = makeCard(CardDesignDiamond, 13)

	_ = f.MoveFreeCellToFoundation(0)
	_ = f.MoveFreeCellToFoundation(1)
	_ = f.MoveFreeCellToFoundation(2)
	_ = f.MoveFreeCellToFoundation(3)

	assert.Equal(t, FreeCellPhaseGameClear, f.GetPhase())
}

// --- Snapshot deep copy tests ---

func TestFreeCellSnapshotDeepCopy(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	_ = f.MoveTableauToFoundation(0)

	// Modify current state
	f.foundation[0] = nil
	f.tableau[0] = []*Card{makeCard(CardDesignHeart, 5)}

	// Undo should restore
	_ = f.Undo()
	assert.Equal(t, 1, len(f.tableau[0]))
	assert.Equal(t, 1, f.tableau[0][0].GetValue())
	assert.Equal(t, 0, len(f.foundation[0]))
}

// --- ActionLog tests ---

func TestFreeCellActionLog(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	_ = f.MoveTableauToFoundation(0)

	log := f.GetActionLog()
	assert.Equal(t, 1, len(log))
	assert.Equal(t, "move", log[0].ActionType)
}

// --- maxMovableCards tests ---

func TestFreeCellMaxMovableCards(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// All freecells empty, all tableau empty
	// Target col 0: empty freecells=4, empty tableau cols (excluding col 0) = 7
	max := f.maxMovableCards(0)
	assert.Equal(t, (1+4)*(1<<7), max) // 5 * 128 = 640

	// One freecell occupied, one tableau col occupied
	f.freeCells[0] = makeCard(CardDesignSpade, 1)
	f.tableau[1] = []*Card{makeCard(CardDesignSpade, 2)}
	max = f.maxMovableCards(0)
	// empty freecells=3, empty tableau (excl 0) = 6
	assert.Equal(t, (1+3)*(1<<6), max) // 4 * 64 = 256

	// Target is empty: should not count target as empty
	max = f.maxMovableCards(2) // col 2 is empty
	// empty freecells=3, empty tableau (excl 2): cols 0,3,4,5,6,7 are empty (6), col 1 occupied → 6 empty
	assert.Equal(t, (1+3)*(1<<6), max) // 4 * 64 = 256
}

// --- isValidTableauSequence tests ---

func TestFreeCellIsValidTableauSequence(t *testing.T) {
	f := newTestFreeCell()

	t.Run("single card", func(t *testing.T) {
		assert.True(t, f.isValidTableauSequence([]*Card{makeCard(CardDesignSpade, 5)}))
	})

	t.Run("valid sequence", func(t *testing.T) {
		cards := []*Card{
			makeCard(CardDesignSpade, 5),
			makeCard(CardDesignHeart, 4),
			makeCard(CardDesignClover, 3),
		}
		assert.True(t, f.isValidTableauSequence(cards))
	})

	t.Run("invalid same color", func(t *testing.T) {
		cards := []*Card{
			makeCard(CardDesignSpade, 5),
			makeCard(CardDesignClover, 4),
		}
		assert.False(t, f.isValidTableauSequence(cards))
	})

	t.Run("invalid not descending", func(t *testing.T) {
		cards := []*Card{
			makeCard(CardDesignSpade, 5),
			makeCard(CardDesignHeart, 5),
		}
		assert.False(t, f.isValidTableauSequence(cards))
	})
}

// --- isAlternateColor / isBlack tests ---

func TestFreeCellIsAlternateColor(t *testing.T) {
	f := newTestFreeCell()

	assert.True(t, f.isAlternateColor(makeCard(CardDesignSpade, 1), makeCard(CardDesignHeart, 1)))
	assert.True(t, f.isAlternateColor(makeCard(CardDesignClover, 1), makeCard(CardDesignDiamond, 1)))
	assert.False(t, f.isAlternateColor(makeCard(CardDesignSpade, 1), makeCard(CardDesignClover, 1)))
	assert.False(t, f.isAlternateColor(makeCard(CardDesignHeart, 1), makeCard(CardDesignDiamond, 1)))
}

func TestFreeCellIsBlack(t *testing.T) {
	f := newTestFreeCell()

	assert.True(t, f.isBlack(makeCard(CardDesignSpade, 1)))
	assert.True(t, f.isBlack(makeCard(CardDesignClover, 1)))
	assert.False(t, f.isBlack(makeCard(CardDesignHeart, 1)))
	assert.False(t, f.isBlack(makeCard(CardDesignDiamond, 1)))
}

// --- canPlaceOnTableau tests ---

func TestFreeCellCanPlaceOnTableau(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	t.Run("king on empty", func(t *testing.T) {
		assert.True(t, f.canPlaceOnTableau(makeCard(CardDesignSpade, 13), 0))
	})

	t.Run("non-king on empty", func(t *testing.T) {
		// フリーセルでは空列に任意のカードを置ける（クロンダイクと異なる）
		assert.True(t, f.canPlaceOnTableau(makeCard(CardDesignSpade, 5), 0))
	})

	t.Run("alternate color descending", func(t *testing.T) {
		f.tableau[0] = []*Card{makeCard(CardDesignSpade, 6)}
		assert.True(t, f.canPlaceOnTableau(makeCard(CardDesignHeart, 5), 0))
	})

	t.Run("same color", func(t *testing.T) {
		f.tableau[1] = []*Card{makeCard(CardDesignSpade, 6)}
		assert.False(t, f.canPlaceOnTableau(makeCard(CardDesignClover, 5), 1))
	})

	t.Run("not descending", func(t *testing.T) {
		f.tableau[2] = []*Card{makeCard(CardDesignSpade, 6)}
		assert.False(t, f.canPlaceOnTableau(makeCard(CardDesignHeart, 6), 2))
	})
}

// --- canPlaceOnFoundation tests ---

func TestFreeCellCanPlaceOnFoundation(t *testing.T) {
	f := setupPlayingFreeCell()

	t.Run("ace on empty", func(t *testing.T) {
		assert.True(t, f.canPlaceOnFoundation(makeCard(CardDesignSpade, 1), 0))
	})

	t.Run("non-ace on empty", func(t *testing.T) {
		assert.False(t, f.canPlaceOnFoundation(makeCard(CardDesignSpade, 5), 0))
	})

	t.Run("same suit ascending", func(t *testing.T) {
		f.foundation[0] = []*Card{makeCard(CardDesignSpade, 1)}
		assert.True(t, f.canPlaceOnFoundation(makeCard(CardDesignSpade, 2), 0))
	})

	t.Run("different suit", func(t *testing.T) {
		f.foundation[1] = []*Card{makeCard(CardDesignClover, 1)}
		assert.False(t, f.canPlaceOnFoundation(makeCard(CardDesignSpade, 2), 1))
	})

	t.Run("not ascending", func(t *testing.T) {
		f.foundation[2] = []*Card{makeCard(CardDesignHeart, 1)}
		assert.False(t, f.canPlaceOnFoundation(makeCard(CardDesignHeart, 3), 2))
	})
}

// --- UndoToEscape / UndoN tests ---

func TestFreeCellUndoToEscape_NotInStalemate(t *testing.T) {
	f := setupPlayingFreeCell()
	assert.Equal(t, 0, f.UndoToEscape())
}

func TestFreeCellUndoToEscape_StalemateNoHistory(t *testing.T) {
	f := setupPlayingFreeCell()
	f.SetIsStalemate(true)
	assert.Equal(t, -1, f.UndoToEscape())
}

func TestFreeCellUndoToEscape_StalemateWithEscape(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// Make a move to create a non-stalemate history entry
	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	err := f.MoveTableauToFoundation(0)
	assert.NoError(t, err)

	// Now set stalemate
	f.SetIsStalemate(true)
	n := f.UndoToEscape()
	assert.Equal(t, 1, n)
}

func TestFreeCellUndoToEscape_AllHistoryStalemate(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// Make a move, then set stalemate on the game, make another move
	f.SetIsStalemate(true)
	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	err := f.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	// History snapshot was taken while isStalemate was true

	f.SetIsStalemate(true)
	assert.Equal(t, -1, f.UndoToEscape())
}

func TestFreeCellUndoN_Zero(t *testing.T) {
	f := setupPlayingFreeCell()
	err := f.UndoN(0)
	assert.NoError(t, err)
}

func TestFreeCellUndoN_Valid(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	// Make two moves
	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	_ = f.MoveTableauToFoundation(0)
	f.tableau[1] = []*Card{makeCard(CardDesignHeart, 1)}
	_ = f.MoveTableauToFoundation(1)

	err := f.UndoN(2)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(f.foundation[0]))
	assert.Equal(t, 0, len(f.foundation[1]))
}

func TestFreeCellUndoN_Excessive(t *testing.T) {
	f := setupPlayingFreeCell()
	clearTableauFC(f)

	f.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	_ = f.MoveTableauToFoundation(0)

	err := f.UndoN(5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undo step")
}

// **何枚まとめて動かせるかは CUI に出ていなかった (#4777)。**Web は
// fc-supermove-limit で常時出している。
func TestFreeCell_GetMaxMovableCards(t *testing.T) {
	card := func(v int) *Card { return NewCard(CardDesignSpade, v, false) }
	board := func(filledCells int, filledCols int) *FreeCell {
		f := NewFreeCell(NewTrumpCards(0))
		f.Reset()
		var cells [FreeCellCellCnt]*Card
		for i := 0; i < filledCells && i < FreeCellCellCnt; i++ {
			cells[i] = card(i + 2)
		}
		f.SetFreeCells(cells)
		var tableau [FreeCellTableauCnt][]*Card
		for i := 0; i < FreeCellTableauCnt; i++ {
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
		assert.Equal(t, 5, board(0, FreeCellTableauCnt).GetMaxMovableCards())
		// セル4つ空き・列1つ空き → (1+4) << 1 = 10。
		assert.Equal(t, 10, board(0, FreeCellTableauCnt-1).GetMaxMovableCards())
		// セル4つ空き・列2つ空き → (1+4) << 2 = 20。
		assert.Equal(t, 20, board(0, FreeCellTableauCnt-2).GetMaxMovableCards())
	})

	// **一般の上限はどの列も除外しない。**特定の列を除外した値を返すと、
	// 空き列が1つ少ない前提の小さすぎる数を出すことになる。
	t.Run("the general limit excludes no column, not even the first", func(t *testing.T) {
		f := NewFreeCell(NewTrumpCards(0))
		f.Reset()
		f.SetFreeCells([FreeCellCellCnt]*Card{})
		var tableau [FreeCellTableauCnt][]*Card
		// 0 列目だけを空にし、残りを埋める。
		for i := 1; i < FreeCellTableauCnt; i++ {
			tableau[i] = []*Card{card(5)}
		}
		f.SetTableau(tableau)
		assert.Equal(t, 10, f.GetMaxMovableCards(), "(1+4) << 1")
	})

	t.Run("a filled free cell lowers the limit", func(t *testing.T) {
		assert.Equal(t, 4, board(1, FreeCellTableauCnt).GetMaxMovableCards())
	})

	t.Run("with everything full only a single card moves", func(t *testing.T) {
		assert.Equal(t, 1, board(FreeCellCellCnt, FreeCellTableauCnt).GetMaxMovableCards())
	})

	// **空き列を移動先にすると上限は下がる。**その列自身を経由地に使えない。
	t.Run("moving onto an empty column halves the limit", func(t *testing.T) {
		f := board(0, FreeCellTableauCnt-1)
		assert.Equal(t, 10, f.GetMaxMovableCards())
		assert.Equal(t, 5, f.GetMaxMovableCardsToEmptyColumn())
	})

	t.Run("reports zero when there is no empty column to move onto", func(t *testing.T) {
		assert.Equal(t, 0, board(0, FreeCellTableauCnt).GetMaxMovableCardsToEmptyColumn())
	})
}
