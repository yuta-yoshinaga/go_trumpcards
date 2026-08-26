//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- helpers ---

func newTestRankAndFile() *domain.RankAndFile {
	tc := domain.NewTrumpCardsWithDecks(2, 0)
	ft := domain.NewRankAndFile(tc)
	return ft
}

func setupPlayingRankAndFile() *domain.RankAndFile {
	ft := newTestRankAndFile()
	ft.Reset()
	return ft
}

func makeRFCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func makeRFTableauCard(design, value int) *domain.RankAndFileTableauCard {
	return &domain.RankAndFileTableauCard{Card: makeRFCard(design, value), FaceUp: true}
}

func clearRFTableau(ft *domain.RankAndFile) {
	var empty [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
	ft.SetTableau(empty)
}

func TestNewRankAndFile(t *testing.T) {
	ft := newTestRankAndFile()
	assert.NotNil(t, ft)
	assert.Equal(t, domain.RankAndFilePhase(0), ft.GetPhase())
}

func TestRankAndFile_Reset(t *testing.T) {
	ft := setupPlayingRankAndFile()

	assert.Equal(t, domain.RankAndFilePhasePlaying, ft.GetPhase())
	assert.Equal(t, 0, ft.GetMoveCount())

	// **各列4枚のうち、下3枚は伏せて最上段だけ表向き。**クローン元の
	// Forty Thieves は40枚すべて表向きに配る。
	tableau := ft.GetTableau()
	totalTableauCards := 0
	for i := 0; i < domain.RankAndFileTableauCnt; i++ {
		assert.Len(t, tableau[i], domain.RankAndFileColSize, "column %d", i)
		for j, tc := range tableau[i] {
			want := j == domain.RankAndFileColSize-1
			assert.Equal(t, want, tc.FaceUp, "column %d card %d face-up state", i, j)
		}
		totalTableauCards += len(tableau[i])
	}
	assert.Equal(t, 40, totalTableauCards)
	assert.False(t, ft.AllFaceUp(), "three cards per column start hidden")

	// Stock: 64 cards (104 - 40)
	assert.Equal(t, 64, ft.GetStockCount())

	// Waste: empty
	assert.Nil(t, ft.GetWaste())

	// Foundation: empty
	foundation := ft.GetFoundation()
	for i := 0; i < domain.RankAndFileFoundationCnt; i++ {
		assert.Nil(t, foundation[i])
	}
}

func TestRankAndFile_Draw(t *testing.T) {
	t.Run("draw from stock", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		initialStock := ft.GetStockCount()
		err := ft.Draw()
		assert.NoError(t, err)
		assert.Equal(t, initialStock-1, ft.GetStockCount())
		assert.Equal(t, 1, len(ft.GetWaste()))
		assert.Equal(t, 1, ft.GetMoveCount())
	})

	t.Run("draw when stock is empty (no recycle)", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		ft.SetStock(nil)
		err := ft.Draw()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no cards in stock")
	})

	t.Run("draw when game is not playing", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.SetPhase(domain.RankAndFilePhaseGameOver)
		err := ft.Draw()
		assert.Error(t, err)
	})
}

func TestRankAndFile_MoveWasteToTableau(t *testing.T) {
	// **異色降順。**クローン元の Forty Thieves は同スート降順なので、
	// その期待値のままクローンすると「通ってしまうのに間違っている」テストになる。
	t.Run("valid move alternating colour descending", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		// ♠5 の上に ♥4 (黒→赤)。
		var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		tableau[0] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 5)}
		ft.SetTableau(tableau)
		ft.SetWaste([]*domain.Card{makeRFCard(domain.CardDesignHeart, 4)})
		ft.SetStock(nil)

		err := ft.MoveWasteToTableau(0)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(ft.GetTableau()[0]))
	})

	// **負のコントロール: クローン元なら通る手。**♠5 の上の ♠4 は同スート降順で、
	// Forty Thieves では合法。Rank and File では同色なので拒む。
	t.Run("reject the same colour", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		tableau[0] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 5)}
		ft.SetTableau(tableau)
		ft.SetWaste([]*domain.Card{makeRFCard(domain.CardDesignSpade, 4)})
		ft.SetStock(nil)

		err := ft.MoveWasteToTableau(0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot place card on tableau")
	})

	t.Run("valid move to empty column", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		ft.SetWaste([]*domain.Card{makeRFCard(domain.CardDesignSpade, 7)})
		ft.SetStock(nil)

		err := ft.MoveWasteToTableau(0)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(ft.GetTableau()[0]))
	})

	t.Run("waste empty", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		err := ft.MoveWasteToTableau(0)
		assert.Error(t, err)
	})

	t.Run("invalid column", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		ft.SetWaste([]*domain.Card{makeRFCard(domain.CardDesignSpade, 1)})
		err := ft.MoveWasteToTableau(-1)
		assert.Error(t, err)
		err = ft.MoveWasteToTableau(10)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.SetPhase(domain.RankAndFilePhaseGameOver)
		err := ft.MoveWasteToTableau(0)
		assert.Error(t, err)
	})
}

func TestRankAndFile_MoveWasteToFoundation(t *testing.T) {
	t.Run("place ace on empty foundation", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		ft.SetWaste([]*domain.Card{makeRFCard(domain.CardDesignSpade, 1)})
		ft.SetStock(nil)

		err := ft.MoveWasteToFoundation()
		assert.NoError(t, err)
		// Check that one foundation has 1 card
		found := false
		for i := 0; i < domain.RankAndFileFoundationCnt; i++ {
			if len(ft.GetFoundation()[i]) == 1 {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("place card on matching foundation", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		var foundation [domain.RankAndFileFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{makeRFCard(domain.CardDesignSpade, 1)}
		ft.SetFoundation(foundation)
		ft.SetWaste([]*domain.Card{makeRFCard(domain.CardDesignSpade, 2)})
		ft.SetStock(nil)

		err := ft.MoveWasteToFoundation()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(ft.GetFoundation()[0]))
	})

	t.Run("cannot place non-ace on empty foundation", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		ft.SetWaste([]*domain.Card{makeRFCard(domain.CardDesignSpade, 5)})
		ft.SetStock(nil)

		err := ft.MoveWasteToFoundation()
		assert.Error(t, err)
	})

	t.Run("waste empty", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		err := ft.MoveWasteToFoundation()
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.SetPhase(domain.RankAndFilePhaseGameClear)
		err := ft.MoveWasteToFoundation()
		assert.Error(t, err)
	})
}

func TestRankAndFile_MoveTableauToTableau(t *testing.T) {
	// **異色降順。**♠6 の上に ♥5。
	t.Run("valid single card move alternating colour descending", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		tableau[0] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignHeart, 5)}
		tableau[1] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 6)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(ft.GetTableau()[0]))
		assert.Equal(t, 2, len(ft.GetTableau()[1]))
	})

	// **クローン元は先頭以外を拒む。**Rank and File は並びの一部を一括で動かせる
	// ので、元のケースは違法から合法に変わる (Deauville は上札のみ、Emperor は
	// 列まるごと、Rank and File だけが任意の並び)。
	t.Run("accepts a partial alternating-colour sequence", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		// ♠K を挟んで ♠6 ♥5 の並び。index 1 から2枚が動く。
		tableau[0] = []*domain.RankAndFileTableauCard{
			makeRFTableauCard(domain.CardDesignSpade, domain.CardValueMax),
			makeRFTableauCard(domain.CardDesignSpade, 6),
			makeRFTableauCard(domain.CardDesignHeart, 5),
		}
		tableau[1] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignHeart, 7)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		require.NoError(t, ft.MoveTableauToTableau(0, 1, 1))
		assert.Len(t, ft.GetTableau()[0], 1, "only the ♠K stays behind")
		assert.Len(t, ft.GetTableau()[1], 3, "♥7 plus the two that travelled with it")
	})

	// **負のコントロール: 並びが崩れていれば1枚も動かない。**
	t.Run("rejects a group that is not an alternating-colour sequence", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		// ♠6 ♠5: 降順だが同色なので並びではない。
		tableau[0] = []*domain.RankAndFileTableauCard{
			makeRFTableauCard(domain.CardDesignSpade, 6),
			makeRFTableauCard(domain.CardDesignSpade, 5),
		}
		tableau[1] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignHeart, 7)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToTableau(0, 0, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sequence")
		assert.Len(t, ft.GetTableau()[0], 2, "nothing moves when the sequence is broken")
	})

	// **負のコントロール: クローン元なら通る手。**♠6 の上の ♠5 は同スート降順で
	// Forty Thieves では合法だが、同色なのでここでは拒む。
	t.Run("reject the same colour", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		tableau[0] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 5)}
		tableau[1] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 6)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("move to empty column", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		tableau[0] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 5)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(ft.GetTableau()[0]))
		assert.Equal(t, 1, len(ft.GetTableau()[1]))
	})

	t.Run("same column", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		err := ft.MoveTableauToTableau(0, 0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid columns", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		err := ft.MoveTableauToTableau(-1, 0, 1)
		assert.Error(t, err)
		err = ft.MoveTableauToTableau(0, 0, 10)
		assert.Error(t, err)
	})

	t.Run("invalid card index", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		// **-1 を渡してはいけない。** -1 は「その列の上札」を意味する CUI 短縮形
		// (`m <from> <to>`) で、合法な着手になりうる ── 配り次第でエラーにならず、
		// このテストはパッケージ全体を走らせたときだけ落ちていた。
		// 範囲外を見たいなら -2 と 100。
		err := ft.MoveTableauToTableau(0, -2, 1)
		assert.Error(t, err)
		err = ft.MoveTableauToTableau(0, 100, 1)
		assert.Error(t, err)
	})

	t.Run("card index -1 means the column's top card", func(t *testing.T) {
		// 短縮形が生きていることを固定する。上のテストが -1 をエラー扱いに
		// 戻されたら、ここが落ちて気付ける。
		//
		// **盤面は組み立てる。** 以前は setupPlayingRankAndFile() を 2 回呼んで
		// 「-1 を渡した盤」と「明示添字を渡した*別の*盤」の成否を比べていた。
		// Reset() はシャッフルするので二つは別の配りで、合法性が食い違って
		// 当然だった —— clean な develop でも落ちる (30 回中数回)。
		build := func() *domain.RankAndFile {
			ft := newTestRankAndFile()
			ft.Reset()
			clearRFTableau(ft)
			var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
			// 列 0 の一番上が ♥4、列 1 の一番上が ♠5。黒→赤の降順なので合法。
			tableau[0] = []*domain.RankAndFileTableauCard{
				makeRFTableauCard(domain.CardDesignHeart, 9),
				makeRFTableauCard(domain.CardDesignHeart, 4),
			}
			tableau[1] = []*domain.RankAndFileTableauCard{
				makeRFTableauCard(domain.CardDesignSpade, 5),
			}
			ft.SetTableau(tableau)
			return ft
		}

		shorthand := build()
		explicit := build()
		top := shorthand.GetTableau()[0]
		require.Len(t, top, 2)

		errShorthand := shorthand.MoveTableauToTableau(0, -1, 1)
		errExplicit := explicit.MoveTableauToTableau(0, len(top)-1, 1)

		require.NoError(t, errExplicit, "明示添字での移動は合法な盤面であること")
		assert.NoError(t, errShorthand, "-1 が「上札」として扱われていない")
		assert.Equal(t, explicit.GetTableau(), shorthand.GetTableau(),
			"-1 と最後の添字で結果の盤面まで一致すること")
	})

	t.Run("not playing phase", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.SetPhase(domain.RankAndFilePhaseGameOver)
		err := ft.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
}

func TestRankAndFile_MoveTableauToFoundation(t *testing.T) {
	t.Run("move ace to foundation", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		tableau[0] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 1)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToFoundation(0)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(ft.GetTableau()[0]))
	})

	t.Run("move card to existing foundation pile", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		var foundation [domain.RankAndFileFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{makeRFCard(domain.CardDesignSpade, 1)}
		ft.SetFoundation(foundation)
		var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		tableau[0] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 2)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToFoundation(0)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(ft.GetFoundation()[0]))
	})

	t.Run("empty column", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		ft.SetStock(nil)

		err := ft.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("cannot place on foundation", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		tableau[0] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 5)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("invalid column", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		err := ft.MoveTableauToFoundation(-1)
		assert.Error(t, err)
		err = ft.MoveTableauToFoundation(10)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.SetPhase(domain.RankAndFilePhaseGameOver)
		err := ft.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})
}

func TestRankAndFile_GiveUp(t *testing.T) {
	t.Run("give up during playing", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		ft.GiveUp()
		assert.Equal(t, domain.RankAndFilePhaseGameOver, ft.GetPhase())
	})

	t.Run("give up when not playing does nothing", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.SetPhase(domain.RankAndFilePhaseGameClear)
		ft.GiveUp()
		assert.Equal(t, domain.RankAndFilePhaseGameClear, ft.GetPhase())
	})
}

func TestRankAndFile_GetHint(t *testing.T) {
	t.Run("hint tableau to foundation", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		tableau[0] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 1)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		hint := ft.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.FromZone)
		assert.Equal(t, "foundation", hint.ToZone)
	})

	t.Run("hint waste to foundation", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		ft.SetWaste([]*domain.Card{makeRFCard(domain.CardDesignHeart, 1)})
		ft.SetStock(nil)

		hint := ft.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "waste", hint.FromZone)
		assert.Equal(t, "foundation", hint.ToZone)
	})

	t.Run("hint tableau to tableau", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		// **異色でなければ動かせない。**同スートのままだとヒントは出ない。
		tableau[0] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignHeart, 5)}
		tableau[1] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 6)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		hint := ft.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.FromZone)
		assert.Equal(t, "tableau", hint.ToZone)
	})

	t.Run("hint waste to tableau", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		tableau[0] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 6)}
		ft.SetTableau(tableau)
		ft.SetWaste([]*domain.Card{makeRFCard(domain.CardDesignSpade, 5)})
		ft.SetStock(nil)

		hint := ft.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "waste", hint.FromZone)
		assert.Equal(t, "tableau", hint.ToZone)
	})

	t.Run("no hint available", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		// Place cards that cannot move anywhere
		var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		tableau[0] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 13)}
		tableau[1] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignHeart, 13)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		hint := ft.GetHint()
		assert.Nil(t, hint)
	})

	// #5525: 他に手が無くストックだけ残っている局面で「ヒントはありません」を
	// 返していた。行き詰まりではないのに、プレイヤーには詰んだのか引けば良いのか
	// 区別が付かない。
	t.Run("suggests drawing when nothing else moves but stock remains", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		tableau[0] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 13)}
		tableau[1] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignHeart, 13)}
		ft.SetTableau(tableau)
		ft.SetStock([]*domain.Card{makeRFCard(domain.CardDesignClover, 7)})

		hint := ft.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "stock", hint.FromZone)
		assert.Equal(t, "waste", hint.ToZone)
		assert.Equal(t, -1, hint.FromCol)
		assert.Equal(t, -1, hint.ToCol)
		assert.Equal(t, -1, hint.CardIndex)
	})

	// **盤上に手があるならそちらが先。**引くのは最後の手段。
	t.Run("prefers a real move over drawing", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		tableau[0] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 1)}
		ft.SetTableau(tableau)
		ft.SetStock([]*domain.Card{makeRFCard(domain.CardDesignClover, 7)})

		hint := ft.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "foundation", hint.ToZone)
	})

	t.Run("no hint when not playing", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.SetPhase(domain.RankAndFilePhaseGameOver)
		hint := ft.GetHint()
		assert.Nil(t, hint)
	})
}

func TestRankAndFile_AutoComplete(t *testing.T) {
	t.Run("auto complete when all face up", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		ft.SetStock(nil)

		// Set up tableau with cards ready for foundation (ordered aces and 2s)
		var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		tableau[0] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 1)}
		tableau[1] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignHeart, 1)}
		ft.SetTableau(tableau)

		err := ft.AutoComplete()
		assert.NoError(t, err)
	})

	t.Run("cannot auto complete with stock remaining", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		err := ft.AutoComplete()
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.SetPhase(domain.RankAndFilePhaseGameOver)
		err := ft.AutoComplete()
		assert.Error(t, err)
	})
}

func TestRankAndFile_AllFaceUp(t *testing.T) {
	// **クローン元 (Forty Thieves) は「ストックが空 = 全部見えている」。**
	// あちらは40枚すべて表向きに配るので正しいが、Rank and File は各列3枚を
	// 伏せるので、その判定のままだと伏せ札が残ったまま true を返す。
	t.Run("still false when the stock is empty but cards are hidden", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		ft.SetStock(nil)
		ft.SetWaste(nil)
		assert.False(t, ft.AllFaceUp())
	})

	t.Run("true only once nothing is hidden anywhere", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		ft.SetStock(nil)
		ft.SetWaste(nil)
		tableau := ft.GetTableau()
		for i := range domain.RankAndFileTableauCnt {
			for _, tc := range tableau[i] {
				tc.FaceUp = true
			}
		}
		ft.SetTableau(tableau)
		assert.True(t, ft.AllFaceUp())
	})

	t.Run("false while the stock has cards", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		assert.False(t, ft.AllFaceUp())
	})
}

func TestRankAndFile_Undo(t *testing.T) {
	t.Run("undo draw", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		initialStock := ft.GetStockCount()
		_ = ft.Draw()
		assert.Equal(t, initialStock-1, ft.GetStockCount())

		err := ft.Undo()
		assert.NoError(t, err)
		assert.Equal(t, initialStock, ft.GetStockCount())
		assert.Equal(t, 0, ft.GetMoveCount())
	})

	t.Run("cannot undo with no history", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		err := ft.Undo()
		assert.Error(t, err)
	})

	t.Run("cannot undo when not playing", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.SetPhase(domain.RankAndFilePhaseGameOver)
		err := ft.Undo()
		assert.Error(t, err)
	})
}

func TestRankAndFile_CanUndo(t *testing.T) {
	t.Run("true after action", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		_ = ft.Draw()
		assert.True(t, ft.CanUndo())
	})

	t.Run("false with no history", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		assert.False(t, ft.CanUndo())
	})
}

func TestRankAndFile_UndoN(t *testing.T) {
	ft := setupPlayingRankAndFile()
	_ = ft.Draw()
	_ = ft.Draw()
	_ = ft.Draw()
	assert.Equal(t, 3, ft.GetMoveCount())

	err := ft.UndoN(3)
	assert.NoError(t, err)
	assert.Equal(t, 0, ft.GetMoveCount())
}

func TestRankAndFile_UndoToEscape(t *testing.T) {
	t.Run("returns 0 when not stalemate", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		assert.Equal(t, 0, ft.UndoToEscape())
	})

	t.Run("returns undo count to escape", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		// Draw a few cards then force stalemate
		_ = ft.Draw()
		_ = ft.Draw()
		ft.SetIsStalemate(true)
		result := ft.UndoToEscape()
		assert.True(t, result > 0)
	})
}

func TestRankAndFile_Stalemate(t *testing.T) {
	t.Run("stalemate when no moves and stock empty", func(t *testing.T) {
		ft := newTestRankAndFile()
		ft.Reset()
		clearRFTableau(ft)
		// Kings can't go to foundation and can't stack
		var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
		tableau[0] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 13)}
		tableau[1] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignHeart, 13)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)
		ft.SetWaste(nil)

		// Trigger stalemate check via a draw (need stock for that)
		// Instead, move waste to tableau to trigger check - but waste is nil
		// Just verify GetHint returns nil and IsStalemate works
		assert.Nil(t, ft.GetHint())
	})

	t.Run("not stalemate when stock has cards", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		// After reset, stock has cards, so not stalemate
		assert.False(t, ft.IsStalemate())
	})
}

func TestRankAndFile_GameClear(t *testing.T) {
	ft := newTestRankAndFile()
	ft.Reset()
	clearRFTableau(ft)
	ft.SetStock(nil)

	// Fill all 8 foundations with 13 cards each
	var foundation [domain.RankAndFileFoundationCnt][]*domain.Card
	designs := []int{domain.CardDesignSpade, domain.CardDesignHeart, domain.CardDesignDiamond, domain.CardDesignClover}
	for i := 0; i < domain.RankAndFileFoundationCnt; i++ {
		foundation[i] = make([]*domain.Card, 0, 13)
		design := designs[i%4]
		for v := 1; v < 13; v++ {
			foundation[i] = append(foundation[i], makeRFCard(design, v))
		}
	}
	ft.SetFoundation(foundation)

	// Place the last card (value 13) on tableau and move to foundation
	var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
	tableau[0] = []*domain.RankAndFileTableauCard{makeRFTableauCard(designs[0], 13)}
	ft.SetTableau(tableau)

	err := ft.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	// Still not clear since other foundations aren't full yet
	// Need all 8 foundations to be 13 cards each
}

func TestRankAndFile_DuplicateSuitFoundation(t *testing.T) {
	// Test that two aces of the same suit go to different foundations
	ft := newTestRankAndFile()
	ft.Reset()
	clearRFTableau(ft)
	ft.SetStock(nil)

	// Place two spade aces
	ft.SetWaste([]*domain.Card{makeRFCard(domain.CardDesignSpade, 1)})
	err := ft.MoveWasteToFoundation()
	assert.NoError(t, err)

	ft.SetWaste([]*domain.Card{makeRFCard(domain.CardDesignSpade, 1)})
	err = ft.MoveWasteToFoundation()
	assert.NoError(t, err)

	// Two different foundation piles should have spade aces
	count := 0
	for i := 0; i < domain.RankAndFileFoundationCnt; i++ {
		if len(ft.GetFoundation()[i]) > 0 {
			count++
		}
	}
	assert.Equal(t, 2, count)
}

func TestRankAndFile_JSON(t *testing.T) {
	t.Run("marshal and unmarshal", func(t *testing.T) {
		ft := setupPlayingRankAndFile()
		_ = ft.Draw()
		_ = ft.Draw()

		data, err := json.Marshal(ft)
		require.NoError(t, err)

		ft2 := &domain.RankAndFile{}
		err = json.Unmarshal(data, ft2)
		require.NoError(t, err)

		assert.Equal(t, ft.GetPhase(), ft2.GetPhase())
		assert.Equal(t, ft.GetMoveCount(), ft2.GetMoveCount())
		assert.Equal(t, ft.GetStockCount(), ft2.GetStockCount())
		assert.Equal(t, len(ft.GetWaste()), len(ft2.GetWaste()))
		assert.True(t, ft2.CanUndo()) // history is now serialized (#1654)
	})

	t.Run("unmarshal with nil trumpCards", func(t *testing.T) {
		// Marshal a fresh game to get valid JSON structure, then unmarshal
		ft1 := newTestRankAndFile()
		ft1.Reset()
		data, err := json.Marshal(ft1)
		require.NoError(t, err)
		ft2 := &domain.RankAndFile{}
		err = json.Unmarshal(data, ft2)
		assert.NoError(t, err)
		assert.Equal(t, ft1.GetPhase(), ft2.GetPhase())
	})

	t.Run("unmarshal rejects oversized arrays", func(t *testing.T) {
		// Create a JSON with >1000 stock cards
		bigSlice := make([]*domain.Card, 1001)
		data, _ := json.Marshal(map[string]interface{}{
			"tc": nil,
			"st": bigSlice,
			"wa": nil,
			"al": nil,
		})
		ft := &domain.RankAndFile{}
		err := json.Unmarshal(data, ft)
		assert.Error(t, err)
	})
}

func TestRankAndFile_ActionLog(t *testing.T) {
	ft := setupPlayingRankAndFile()
	assert.Nil(t, ft.GetActionLog())

	_ = ft.Draw()
	log := ft.GetActionLog()
	assert.Equal(t, 1, len(log))
	assert.Equal(t, "draw", log[0].ActionType)
}

// --- Rank and File's three divergences from Forty Thieves ---
//
// A clone's tests pass by default: they assert the SOURCE game's rules. Each
// test below carries a negative control that Forty Thieves would fail.
//
// Sources: solitairecentral, esolutions.se, solavant. Deauville / Emperor /
// Rank and File are near-identical and differ only in how much of a sequence
// may move — Deauville the top card, Emperor a whole pile, Rank and File any
// part of one.

// **札が減ったら伏せ札をめくる。**Forty Thieves は40枚すべて表向きに配るので
// この処理が存在しない。めくり忘れると、下に札があるのに何も置けない列ができて
// 盤が静かに詰む。
func TestRankAndFile_AutoFlipsAfterTheTopCardLeaves(t *testing.T) {
	ft := setupPlayingRankAndFile()
	clearRFTableau(ft)

	var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
	// 列0: 伏せた ♦9 の上に表の ♥5。列1 は ♠6。
	hidden := makeRFTableauCard(domain.CardDesignDiamond, 9)
	hidden.FaceUp = false
	tableau[0] = []*domain.RankAndFileTableauCard{hidden, makeRFTableauCard(domain.CardDesignHeart, 5)}
	tableau[1] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 6)}
	ft.SetTableau(tableau)
	ft.SetStock(nil)

	require.NoError(t, ft.MoveTableauToTableau(0, 1, 1))
	require.Len(t, ft.GetTableau()[0], 1)
	assert.True(t, ft.GetTableau()[0][0].FaceUp, "the card underneath is turned up")
}

// 伏せ札は積み先にならないし、組札にも送れない。
func TestRankAndFile_FaceDownCardsAreNotInPlay(t *testing.T) {
	ft := setupPlayingRankAndFile()
	clearRFTableau(ft)

	var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
	hidden := makeRFTableauCard(domain.CardDesignSpade, 6)
	hidden.FaceUp = false
	tableau[0] = []*domain.RankAndFileTableauCard{hidden}
	tableau[1] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignHeart, 5)}
	ft.SetTableau(tableau)
	ft.SetStock(nil)

	// ♥5 は ♠6 の上に置けるはずだが、その ♠6 は伏せられている。
	assert.Error(t, ft.MoveTableauToTableau(1, 0, 0), "cannot stack onto a face-down card")

	// 伏せた A も組札へ送れない。
	ace := makeRFTableauCard(domain.CardDesignSpade, 1)
	ace.FaceUp = false
	var t2 [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
	t2[0] = []*domain.RankAndFileTableauCard{ace}
	ft.SetTableau(t2)
	assert.Error(t, ft.MoveTableauToFoundation(0), "a face-down ace is not visible")
}

// **上札だけを見た走査は、まだ手のある盤を見落とす。**Forty Thieves は1枚ずつ
// しか動かせないので上札の走査で足りるが、Rank and File は並びを一括で動かせる。
func TestRankAndFile_HintSeesASequenceBelowTheTop(t *testing.T) {
	ft := setupPlayingRankAndFile()
	clearRFTableau(ft)

	var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
	// 列0 の上札 ♥5 の行き先 (黒6) は列1 にあるが、♠6 ♥5 の並びごと ♥7 へ動かせる。
	tableau[0] = []*domain.RankAndFileTableauCard{
		makeRFTableauCard(domain.CardDesignDiamond, domain.CardValueMax),
		makeRFTableauCard(domain.CardDesignSpade, 6),
		makeRFTableauCard(domain.CardDesignHeart, 5),
	}
	tableau[1] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignHeart, 7)}
	ft.SetTableau(tableau)
	ft.SetStock(nil)
	ft.SetWaste(nil)

	hint := ft.GetHint()
	require.NotNil(t, hint, "the sequence move exists")
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, 0, hint.FromCol)
	assert.Equal(t, 1, hint.CardIndex, "the hint names the sequence start, not the top card")
	// ヒントが名指しした手は本当に指せる。
	assert.NoError(t, ft.MoveTableauToTableau(hint.FromCol, hint.CardIndex, hint.ToCol))
}

func TestRankAndFile_SequenceStartsStopsAtTheBreakAndAtHiddenCards(t *testing.T) {
	ft := setupPlayingRankAndFile()
	clearRFTableau(ft)

	var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
	hidden := makeRFTableauCard(domain.CardDesignClover, 2)
	hidden.FaceUp = false
	// 伏せ ♣2 / ♦K / ♠6 / ♥5: 並びは上2枚 (index 2,3)。♦K で色も階級も切れる。
	tableau[0] = []*domain.RankAndFileTableauCard{
		hidden,
		makeRFTableauCard(domain.CardDesignDiamond, domain.CardValueMax),
		makeRFTableauCard(domain.CardDesignSpade, 6),
		makeRFTableauCard(domain.CardDesignHeart, 5),
	}
	ft.SetTableau(tableau)

	assert.Equal(t, []int{3, 2}, ft.SequenceStarts(0), "top card first, stopping where the sequence breaks")
	assert.Empty(t, ft.SequenceStarts(1), "an empty column has no sequence start")
	assert.Nil(t, ft.SequenceStarts(domain.RankAndFileTableauCnt), "out of range is nil, not a panic")
}

// **CUI の短縮形が届くこと。**`m <from> <to>` は cardIndex に -1 を渡す
// (コントローラのコメントどおり「上札を動かす」の意)。範囲チェックが素直に
// -1 を弾くと、ヘルプに載っている短縮形が必ず "invalid card index" で失敗する。
func TestRankAndFile_ShorthandIndexMovesTheTopCard(t *testing.T) {
	ft := newTestRankAndFile()
	ft.Reset()
	clearRFTableau(ft)
	var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
	// 列0は ♠J の上に ♥8。並びではないので、-1 が「上札」でなく「列の先頭」や
	// 「全部」と解釈されたらここで落ちる。
	tableau[0] = []*domain.RankAndFileTableauCard{
		makeRFTableauCard(domain.CardDesignSpade, 11),
		makeRFTableauCard(domain.CardDesignHeart, 8),
	}
	tableau[1] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 9)}
	ft.SetTableau(tableau)
	ft.SetStock(nil)

	require.NoError(t, ft.MoveTableauToTableau(0, -1, 1))

	assert.Len(t, ft.GetTableau()[0], 1, "上札だけが抜ける")
	require.Len(t, ft.GetTableau()[1], 2)
	assert.Equal(t, 8, ft.GetTableau()[1][1].Card.GetValue())
}

// **負のコントロール。**-1 を通すために範囲チェックごと外していないことを見る。
func TestRankAndFile_OtherNegativeIndicesAreStillRejected(t *testing.T) {
	ft := newTestRankAndFile()
	ft.Reset()
	clearRFTableau(ft)
	var tableau [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
	tableau[0] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignHeart, 8)}
	tableau[1] = []*domain.RankAndFileTableauCard{makeRFTableauCard(domain.CardDesignSpade, 9)}
	ft.SetTableau(tableau)
	ft.SetStock(nil)

	assert.Error(t, ft.MoveTableauToTableau(0, -2, 1))
}
