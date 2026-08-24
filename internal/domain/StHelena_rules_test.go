//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- St. Helena's four divergences from Crescent ---
//
// Crescent is the clone source: two decks, eight foundations seeded with the
// four aces and four kings, a tableau of face-up piles, and redeals. Everything
// below is a rule Crescent does NOT have, so each test carries a negative
// control that the Crescent predicate would fail.

// **タブローはスートを見ない。**クレセントは同スートの ±1 なので、その述語を
// 残すと異なるスートの手が全部消える。
func TestStHelena_TableauIgnoresSuit(t *testing.T) {
	cr := setupPlayingStHelena()
	clearStHelenaTableau(cr)
	var tb [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
	tb[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignHeart, 7)}
	tb[1] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 8)}
	cr.SetTableau(tb)

	// 負のコントロール: ♥7 の上に ♠8 はクレセントならスート違いで拒む。
	require.NoError(t, cr.MoveTableauToTableau(1, 0))
	assert.Len(t, cr.GetTableau()[0], 2)
	assert.Len(t, cr.GetTableau()[1], 0)
}

// **K と A は繋がらない。**クレセントは A↔K の折り返しを許すので、その分岐を
// 残すと置けないはずの手が置ける。
func TestStHelena_KingAndAceDoNotWrap(t *testing.T) {
	cr := setupPlayingStHelena()
	clearStHelenaTableau(cr)
	var tb [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
	tb[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignHeart, domain.CardValueMax)}
	tb[1] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 1)}
	tb[2] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignClover, domain.CardValueMax-1)}
	cr.SetTableau(tb)

	assert.Error(t, cr.MoveTableauToTableau(1, 0), "A は K の上に置けない")
	assert.Error(t, cr.MoveTableauToTableau(0, 1), "K は A の上に置けない")
	// 負のコントロール: 折り返しを消したせいで ±1 まで壊していないこと。
	require.NoError(t, cr.MoveTableauToTableau(2, 0), "Q は K の上に置ける")
}

// **初回の配りでは、どの列がどの組札へ送れるかが決まっている。**上 4 列は K 段
// (降順) のみ、下 4 列は A 段 (昇順) のみ、左右 4 列はどちらでも。クレセントには
// この制限が無い。
func TestStHelena_FirstDealRestrictsWhichFoundationsAColumnCanReach(t *testing.T) {
	// 盤を毎回組み直す。1 つの盤で続けて送ると、先の手が組札を進めてしまい、
	// 次の手が「制限」ではなく「ランク違い」で落ちる。
	setup := func(col, design, value int) *domain.StHelena {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		clearStHelenaFoundation(cr)

		var fnd [domain.StHelenaFoundationCnt][]*domain.Card
		suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
		for i, s := range suits {
			fnd[i] = []*domain.Card{makeStHelenaCard(s, 1)}
			fnd[i+domain.StHelenaAscendingFoundationCnt] = []*domain.Card{makeStHelenaCard(s, domain.CardValueMax)}
		}
		cr.SetFoundation(fnd)

		var tb [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		tb[col] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(design, value)}
		cr.SetTableau(tb)
		return cr
	}

	const aceRow, kingRow = 0, domain.StHelenaAscendingFoundationCnt
	spade := domain.CardDesignSpade
	two, queen := 2, domain.CardValueMax-1

	t.Run("a top column reaches only the king row", func(t *testing.T) {
		assert.Error(t, setup(0, spade, two).MoveTableauToFoundation(0, aceRow))
		assert.NoError(t, setup(0, spade, queen).MoveTableauToFoundation(0, kingRow))
	})

	t.Run("a bottom column reaches only the ace row", func(t *testing.T) {
		assert.NoError(t, setup(6, spade, two).MoveTableauToFoundation(6, aceRow))
		assert.Error(t, setup(6, spade, queen).MoveTableauToFoundation(6, kingRow))
	})

	t.Run("a side column reaches both rows", func(t *testing.T) {
		for _, col := range domain.StHelenaSideColumns {
			assert.NoError(t, setup(col, spade, two).MoveTableauToFoundation(col, aceRow), "col %d", col)
			assert.NoError(t, setup(col, spade, queen).MoveTableauToFoundation(col, kingRow), "col %d", col)
		}
	})

	// **負のコントロール: 制限は「送り先」だけを絞る。**置けない札まで通すように
	// してしまうと、上の NoError は制限を測らずランク判定の穴を測ることになる。
	t.Run("reachability does not override the rank rule", func(t *testing.T) {
		assert.Error(t, setup(5, spade, two).MoveTableauToFoundation(5, kingRow),
			"横の列でも 2 は K 段には積めない")
	})
}

// **再配りで制限が解ける。**規則の眼目で、これを実装しないと後半に手が無くなる。
func TestStHelena_RedealLiftsTheRestriction(t *testing.T) {
	cr := setupPlayingStHelena()
	require.True(t, cr.RestrictionsActive(), "配り直後は制限あり")

	require.NoError(t, cr.Redeal())
	assert.False(t, cr.RestrictionsActive(), "1 回目の再配りで解ける")
}

// **再配りは列を裏返すのではなく、最後の列から集めて配り直す。**クレセントは
// 各列をその場で逆順にするだけなので、列をまたいだ並べ替えが起きない。
func TestStHelena_RedealGathersFromTheLastPileAndDealsAgain(t *testing.T) {
	cr := setupPlayingStHelena()
	before := cr.GetTableau()
	// 配り直後は 12 列 × 8 枚。
	for i, col := range before {
		require.Len(t, col, domain.StHelenaTableauInitialSize, "col %d", i)
	}

	require.NoError(t, cr.Redeal())
	after := cr.GetTableau()

	// 枚数の形は変わらない (何も失われない)。
	total := 0
	for _, col := range after {
		total += len(col)
	}
	assert.Equal(t, domain.StHelenaTableauCnt*domain.StHelenaTableauInitialSize, total)

	// **列をまたいで並べ替わっている。**その場で逆順にするだけなら、列0 の
	// 中身は列0 のままになる。
	//
	// **見るのは列 0 全体で、底の 1 枚ではない。** 104 枚のダブルデッキには
	// 同じ札が 2 枚あるので、集め直した結果たまたま同じ絵柄・同じ数字が底に
	// 戻ることがある ── 底 1 枚だけを見ていた頃は、`internal/domain` の一括
	// 実行で 1% 前後の割合で落ちていた (実測)。8 枚すべてが一致する確率は
	// 無視できる。
	moved := false
	for i := range before[0] {
		b, a := before[0][i].Card, after[0][i].Card
		if b.GetDesign() != a.GetDesign() || b.GetValue() != a.GetValue() {
			moved = true
			break
		}
	}
	assert.True(t, moved, "集め直していれば列0の中身は入れ替わる")
}

// **12 列 × 8 枚で 96 枚。**種札 8 枚と合わせて 104 枚。クレセントの 16 × 6 の
// ままだと勘定が合わない。
func TestStHelena_DealIsTwelveColumnsOfEight(t *testing.T) {
	cr := setupPlayingStHelena()

	assert.Equal(t, 12, domain.StHelenaTableauCnt)
	assert.Equal(t, 8, domain.StHelenaTableauInitialSize)
	assert.Equal(t, 2, domain.StHelenaMaxRedeals)

	total := 0
	for _, col := range cr.GetTableau() {
		total += len(col)
	}
	for _, pile := range cr.GetFoundation() {
		total += len(pile)
	}
	assert.Equal(t, domain.CardCnt*2, total, "104 枚すべてが盤上にある")
}

// **制限は「送れる手」の定義そのもの。**`MoveTableauToFoundation` だけに書くと、
// 同じ盤を読む他の経路がそれを知らないまま先へ進む
// ([[feedback_new_variant_needs_every_consumer]] の形)。
func TestStHelena_TheRestrictionBindsEveryPathToAFoundation(t *testing.T) {
	// 列 0 (上) の上札が A 段にちょうど乗る盤。手で送るのは拒まれるので、
	// ヒントもオートコンプリートも同じ手を選んではいけない。
	setup := func() *domain.StHelena {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		clearStHelenaFoundation(cr)
		var fnd [domain.StHelenaFoundationCnt][]*domain.Card
		fnd[0] = []*domain.Card{makeStHelenaCard(domain.CardDesignSpade, 1)}
		cr.SetFoundation(fnd)
		var tb [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		tb[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 2)}
		cr.SetTableau(tb)
		return cr
	}

	// 前提: 手で送るのは拒まれる。ここが通ってしまうと以下は何も測らない。
	require.Error(t, setup().MoveTableauToFoundation(0, 0))

	t.Run("GetHint does not point at it", func(t *testing.T) {
		h := setup().GetHint()
		if h != nil && h.ToZone == "foundation" {
			assert.Fail(t, "拒まれる手をヒントが指している", "col %d -> foundation %d", h.FromCol, h.ToCol)
		}
	})

	t.Run("AutoComplete does not play it", func(t *testing.T) {
		cr := setup()
		_ = cr.AutoComplete()
		assert.Len(t, cr.GetFoundation()[0], 1, "組札は種札のまま")
		assert.Len(t, cr.GetTableau()[0], 1, "札は列に残る")
	})

	// 負のコントロール: 制限が解ければ、どちらの経路でも同じ手を選ぶ。
	// 「常に送らない」実装でも上の 2 つは通ってしまう。
	t.Run("both play it once the restriction is lifted", func(t *testing.T) {
		cr := setup()
		cr.SetRestrictionsActive(false)
		h := cr.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "foundation", h.ToZone)

		cr2 := setup()
		cr2.SetRestrictionsActive(false)
		require.NoError(t, cr2.AutoComplete())
		assert.Len(t, cr2.GetFoundation()[0], 2)
	})
}
