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

func newTestFourteenOut() *domain.FourteenOut {
	return domain.NewFourteenOut(domain.NewTrumpCards(0))
}

func playingFourteenOut(t *testing.T) *domain.FourteenOut {
	t.Helper()
	g := newTestFourteenOut()
	g.Reset()
	return g
}

func foCard(design, value int) *domain.Card { return domain.NewCard(design, value, true) }

// foBoard は列を並べた盤を作る。指定しなかった列は空になる。
func foBoard(cols ...[]*domain.Card) [][]*domain.Card {
	board := make([][]*domain.Card, domain.FourteenOutColumnCnt)
	for i := range board {
		if i < len(cols) {
			board[i] = cols[i]
			continue
		}
		board[i] = nil
	}
	return board
}

// --- the deal ---

// **「ほぼ均等」ではない。**52 = 12*4 + 4 なので、左から 4 列だけが 5 枚。
// クローン元の Monte Carlo は 5x5 のグリッドに 25 枚を配り、残り 27 枚を山札に
// 残す ── Fourteen Out は 52 枚すべてを配り切り、山札は存在しない。
func TestFourteenOut_DealsTwelveColumnsFirstFourLonger(t *testing.T) {
	g := playingFourteenOut(t)
	cols := g.GetColumns()

	require.Len(t, cols, domain.FourteenOutColumnCnt)
	total := 0
	for i, col := range cols {
		want := 4
		if i < domain.FourteenOutLongColumns {
			want = 5
		}
		assert.Len(t, col, want, "column %d", i)
		total += len(col)
	}
	assert.Equal(t, 52, total, "the whole deck is dealt; nothing is held back")
	assert.Equal(t, domain.FourteenOutPhasePlaying, g.GetPhase())
	assert.Equal(t, 0, g.GetRemovedCount())
}

func TestFourteenOut_ResetClearsHistoryAndLog(t *testing.T) {
	g := playingFourteenOut(t)
	g.SetColumns(foBoard(
		[]*domain.Card{foCard(domain.CardDesignSpade, 9)},
		[]*domain.Card{foCard(domain.CardDesignHeart, 5)},
	))
	require.NoError(t, g.Remove(0, 1))
	require.True(t, g.CanUndo())

	g.Reset()
	assert.False(t, g.CanUndo(), "history is cleared")
	assert.Empty(t, g.GetActionLog(), "log is cleared")
	assert.Equal(t, 0, g.GetRemovedCount())
}

// --- the sum rule ---

// **合計 14 だけが規則。**K と A も「特例」ではなく 13+1=14 の一例にすぎない。
// クローン元の Monte Carlo は同ランクどうしを組むので、その期待値のまま
// クローンすると通ってしまうのに間違っているテストになる。
func TestFourteenOut_RemovesEveryPairSummingToFourteen(t *testing.T) {
	pairs := []struct {
		name string
		a, b int
	}{
		{"K and A", domain.CardValueMax, 1},
		{"Q and 2", 12, 2},
		{"J and 3", 11, 3},
		{"10 and 4", 10, 4},
		{"9 and 5", 9, 5},
		{"8 and 6", 8, 6},
		{"7 and 7", 7, 7},
	}
	for _, p := range pairs {
		t.Run(p.name, func(t *testing.T) {
			g := playingFourteenOut(t)
			g.SetColumns(foBoard(
				[]*domain.Card{foCard(domain.CardDesignSpade, p.a)},
				[]*domain.Card{foCard(domain.CardDesignHeart, p.b)},
			))
			require.NoError(t, g.Remove(0, 1))
			assert.Empty(t, g.GetColumns()[0])
			assert.Empty(t, g.GetColumns()[1])
			assert.Equal(t, 2, g.GetRemovedCount())
		})
	}
}

// 負のコントロール: 13 でも 15 でも組めない。境界の両側を見る。
func TestFourteenOut_RefusesPairsThatDoNotSumToFourteen(t *testing.T) {
	for _, tc := range []struct {
		name string
		a, b int
	}{
		{"sums to 13", 9, 4},
		{"sums to 15", 9, 6},
		{"two kings sum to 26", domain.CardValueMax, domain.CardValueMax},
		{"king with a two sums to 15", domain.CardValueMax, 2},
		{"same rank that is not 7-7", 5, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := playingFourteenOut(t)
			g.SetColumns(foBoard(
				[]*domain.Card{foCard(domain.CardDesignSpade, tc.a)},
				[]*domain.Card{foCard(domain.CardDesignHeart, tc.b)},
			))
			err := g.Remove(0, 1)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "14")
			assert.Len(t, g.GetColumns()[0], 1, "nothing is removed on a refusal")
			assert.Len(t, g.GetColumns()[1], 1)
		})
	}
}

// **スートは一切見ない。**同じ組み合わせがスート違いでも通ること。
func TestFourteenOut_IgnoresSuit(t *testing.T) {
	for _, d := range []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond} {
		g := playingFourteenOut(t)
		g.SetColumns(foBoard(
			[]*domain.Card{foCard(domain.CardDesignSpade, 9)},
			[]*domain.Card{foCard(d, 5)},
		))
		assert.NoError(t, g.Remove(0, 1), "design %d", d)
	}
}

// **隣接は関係ない。**離れた列どうしでも、露出していれば組める。
// クローン元の Monte Carlo は隣接セルしか組めないので、ここが分岐点。
func TestFourteenOut_PairsDistantColumns(t *testing.T) {
	g := playingFourteenOut(t)
	cols := foBoard()
	cols[0] = []*domain.Card{foCard(domain.CardDesignSpade, 9)}
	cols[11] = []*domain.Card{foCard(domain.CardDesignHeart, 5)}
	g.SetColumns(cols)

	assert.NoError(t, g.Remove(0, 11), "column 0 and column 11 are not adjacent")
}

// **末尾しか露出していない。**埋もれた札は合計が合っても組めない。
func TestFourteenOut_OnlyTheTailIsAvailable(t *testing.T) {
	g := playingFourteenOut(t)
	g.SetColumns(foBoard(
		// 列0 の末尾は ♠2。その下の ♠9 は 5 と組めるが、埋もれているので使えない。
		[]*domain.Card{foCard(domain.CardDesignSpade, 9), foCard(domain.CardDesignSpade, 2)},
		[]*domain.Card{foCard(domain.CardDesignHeart, 5)},
	))

	err := g.Remove(0, 1)
	require.Error(t, err, "2 + 5 = 7, and the buried 9 is not in play")

	// 末尾の ♠2 を Q と組んで取り除くと、その下の ♠9 が露出して 5 と組める。
	g.SetColumns(foBoard(
		[]*domain.Card{foCard(domain.CardDesignSpade, 9), foCard(domain.CardDesignSpade, 2)},
		[]*domain.Card{foCard(domain.CardDesignHeart, 5)},
		[]*domain.Card{foCard(domain.CardDesignClover, 12)},
	))
	require.NoError(t, g.Remove(0, 2), "♠2 pairs with ♣Q")
	require.NoError(t, g.Remove(0, 1), "the 9 underneath is exposed now")
}

func TestFourteenOut_RemoveErrors(t *testing.T) {
	t.Run("same column twice", func(t *testing.T) {
		g := playingFourteenOut(t)
		assert.Error(t, g.Remove(0, 0))
	})

	t.Run("out of range", func(t *testing.T) {
		g := playingFourteenOut(t)
		assert.Error(t, g.Remove(-1, 0))
		assert.Error(t, g.Remove(0, domain.FourteenOutColumnCnt))
	})

	t.Run("empty column", func(t *testing.T) {
		g := playingFourteenOut(t)
		g.SetColumns(foBoard([]*domain.Card{foCard(domain.CardDesignSpade, 9)}))
		assert.Error(t, g.Remove(0, 1), "column 1 is empty")
	})

	t.Run("not playing", func(t *testing.T) {
		g := playingFourteenOut(t)
		g.GiveUp()
		assert.Error(t, g.Remove(0, 1))
	})
}

// --- end states ---

func TestFourteenOut_GameClearWhenEveryCardIsGone(t *testing.T) {
	g := playingFourteenOut(t)
	g.SetColumns(foBoard(
		[]*domain.Card{foCard(domain.CardDesignSpade, 9)},
		[]*domain.Card{foCard(domain.CardDesignHeart, 5)},
	))
	g.SetRemovedCount(50) // 残り 2 枚という状況を作る

	require.NoError(t, g.Remove(0, 1))
	assert.Equal(t, domain.FourteenOutPhaseGameClear, g.GetPhase())
	assert.True(t, g.IsComplete())
}

// **山札が無いので、組が尽きた時点で敗北。**クローン元は補充で救われるので、
// あちらの「山札があれば手詰まりでない」分岐をそのまま残すと永久に詰まない。
func TestFourteenOut_StalemateWhenNoPairRemains(t *testing.T) {
	g := playingFourteenOut(t)
	g.SetColumns(foBoard(
		[]*domain.Card{foCard(domain.CardDesignSpade, 2)},
		[]*domain.Card{foCard(domain.CardDesignHeart, 3)},
		[]*domain.Card{foCard(domain.CardDesignClover, 4)},
	))
	g.CheckFourteenOutStalemate()

	assert.True(t, g.IsStalemate(), "2+3=5, 2+4=6, 3+4=7 — nothing reaches 14")
	assert.Nil(t, g.Hint())
	assert.Equal(t, 0, g.CountRemovablePairs())
}

func TestFourteenOut_NotStalemateWhileAPairRemains(t *testing.T) {
	g := playingFourteenOut(t)
	g.SetColumns(foBoard(
		[]*domain.Card{foCard(domain.CardDesignSpade, 9)},
		[]*domain.Card{foCard(domain.CardDesignHeart, 5)},
		[]*domain.Card{foCard(domain.CardDesignClover, 4)},
	))
	g.CheckFourteenOutStalemate()

	assert.False(t, g.IsStalemate())
	assert.Equal(t, 1, g.CountRemovablePairs(), "only 9+5")
}

func TestFourteenOut_GiveUp(t *testing.T) {
	g := playingFourteenOut(t)
	g.GiveUp()
	assert.Equal(t, domain.FourteenOutPhaseGameOver, g.GetPhase())
	assert.True(t, g.GetGameEndFlag())

	// 終局後の GiveUp は何もしない。
	g.GiveUp()
	assert.Equal(t, domain.FourteenOutPhaseGameOver, g.GetPhase())
}

// --- hint / counting ---

func TestFourteenOut_HintNamesARemovablePair(t *testing.T) {
	g := playingFourteenOut(t)
	g.SetColumns(foBoard(
		[]*domain.Card{foCard(domain.CardDesignSpade, 2)},
		[]*domain.Card{foCard(domain.CardDesignHeart, 9)},
		[]*domain.Card{foCard(domain.CardDesignClover, 5)},
	))

	hint := g.Hint()
	require.NotNil(t, hint)
	assert.Equal(t, domain.FourteenOutHintActionRemove, hint.Action)
	assert.Equal(t, 1, hint.FromCol)
	assert.Equal(t, 2, hint.ToCol)
	// ヒントが名指しした手は本当に指せる。
	assert.NoError(t, g.Remove(hint.FromCol, hint.ToCol))
}

func TestFourteenOut_HintNilWhenGameEnded(t *testing.T) {
	g := playingFourteenOut(t)
	g.GiveUp()
	assert.Nil(t, g.Hint())
}

// **数え方と探し方が食い違うと「組があるのに手詰まり」になる (#5587)。**
// 同じ走査を通していることを、件数とヒントの両方から確かめる。
func TestFourteenOut_CountAgreesWithTheHint(t *testing.T) {
	g := playingFourteenOut(t)
	g.SetColumns(foBoard(
		[]*domain.Card{foCard(domain.CardDesignSpade, 7)},
		[]*domain.Card{foCard(domain.CardDesignHeart, 7)},
		[]*domain.Card{foCard(domain.CardDesignClover, 7)},
	))
	// 7 が 3 枚 → 組は (0,1) (0,2) (1,2) の 3 通り。同じ組を 2 度数えない。
	assert.Equal(t, 3, g.CountRemovablePairs())
	assert.NotNil(t, g.Hint())

	g.SetColumns(foBoard([]*domain.Card{foCard(domain.CardDesignSpade, 7)}))
	assert.Equal(t, 0, g.CountRemovablePairs(), "a lone 7 has no partner")
	assert.Nil(t, g.Hint())
}

// --- undo ---

func TestFourteenOut_UndoRestoresBothCards(t *testing.T) {
	g := playingFourteenOut(t)
	g.SetColumns(foBoard(
		[]*domain.Card{foCard(domain.CardDesignSpade, 3), foCard(domain.CardDesignSpade, 9)},
		[]*domain.Card{foCard(domain.CardDesignHeart, 5)},
	))
	require.NoError(t, g.Remove(0, 1))
	require.Equal(t, 2, g.GetRemovedCount())

	require.NoError(t, g.Undo())
	assert.Equal(t, 0, g.GetRemovedCount())
	require.Len(t, g.GetColumns()[0], 2, "the 9 comes back on top of the 3")
	assert.Equal(t, 9, g.GetColumns()[0][1].GetValue())
	require.Len(t, g.GetColumns()[1], 1)
	assert.Equal(t, 5, g.GetColumns()[1][0].GetValue())
}

func TestFourteenOut_UndoFailsWithNoHistory(t *testing.T) {
	g := playingFourteenOut(t)
	assert.Error(t, g.Undo())
	assert.False(t, g.CanUndo())
}

func TestFourteenOut_CanUndoFalseAfterGameEnd(t *testing.T) {
	g := playingFourteenOut(t)
	g.SetColumns(foBoard(
		[]*domain.Card{foCard(domain.CardDesignSpade, 9)},
		[]*domain.Card{foCard(domain.CardDesignHeart, 5)},
	))
	require.NoError(t, g.Remove(0, 1))
	require.True(t, g.CanUndo())
	g.GiveUp()
	assert.False(t, g.CanUndo())
}

// --- persistence ---

func TestFourteenOut_JSONRoundTrip(t *testing.T) {
	g := playingFourteenOut(t)
	g.SetColumns(foBoard(
		[]*domain.Card{foCard(domain.CardDesignSpade, 3), foCard(domain.CardDesignSpade, 9)},
		[]*domain.Card{foCard(domain.CardDesignHeart, 5)},
	))
	require.NoError(t, g.Remove(0, 1))

	data, err := json.Marshal(g)
	require.NoError(t, err)

	restored := domain.NewDefaultFourteenOut()
	require.NoError(t, json.Unmarshal(data, restored))
	assert.Equal(t, g.GetRemovedCount(), restored.GetRemovedCount())
	assert.Equal(t, len(g.GetColumns()), len(restored.GetColumns()))

	// **アンドゥ履歴も往復しないと、Worker では Undo が黙って効かない (#4478)。**
	require.True(t, restored.CanUndo(), "history survived the round trip")
	require.NoError(t, restored.Undo())
	assert.Equal(t, 0, restored.GetRemovedCount())
}

func TestFourteenOut_JSONRejectsOutOfRangeInput(t *testing.T) {
	// 列数は配り切った 12 を超えられない。
	cols := make([]string, 0, 20)
	for range 20 {
		cols = append(cols, "[]")
	}
	huge := `{"cl":[` + join(cols, ",") + `]}`
	assert.Error(t, json.Unmarshal([]byte(huge), domain.NewDefaultFourteenOut()))
}

func join(parts []string, sep string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += sep
		}
		out += p
	}
	return out
}
