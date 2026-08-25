//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dc は 1 枚作る短縮。
func dc(design, value int) *Card { return NewCard(design, value, true) }

// dcards はテーブルを組む短縮。
func dcards(pairs ...[2]int) []*Card {
	out := make([]*Card, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, dc(p[0], p[1]))
	}
	return out
}

func TestDilotiCardValue(t *testing.T) {
	// **A は 1、絵札は値を持たない。** 合計に絵札を混ぜると、J を 11 と数えて
	// 「J+A で 12」のような、この系統には存在しない捕獲が生まれる。
	assert.Equal(t, 1, DilotiCardValue(dc(CardDesignSpade, 1)))
	assert.Equal(t, 10, DilotiCardValue(dc(CardDesignHeart, 10)))
	for _, v := range []int{11, 12, 13} {
		assert.Equal(t, 0, DilotiCardValue(dc(CardDesignClover, v)),
			"絵札 %d が合計値を持ってしまっている", v)
	}
}

// **#5458 の「合計 10 固定」は誤り。** 捕獲の基準は出した札のランクであって
// 10 ではない。10 固定にすると、5 で 2+3 を取る当たり前の手が打てなくなる。
func TestDiloti_CaptureTargetIsThePlayedRankNotTen(t *testing.T) {
	table := dcards([2]int{CardDesignSpade, 2}, [2]int{CardDesignHeart, 3},
		[2]int{CardDesignClover, 7})
	opts := EnumerateDilotiTakes(dc(CardDesignDiamond, 5), table, nil)
	require.NotEmpty(t, opts, "5 で 2+3 が取れない — 基準が 10 に固定されている")

	found := false
	for _, o := range opts {
		if len(o.TableIdxs) == 2 && sumOfIdxs(table, o.TableIdxs) == 5 {
			found = true
		}
	}
	assert.True(t, found, "合計 5 の組が候補に無い")

	// 逆に、合計 10 になる 3+7 は 5 では取れない。
	for _, o := range opts {
		assert.NotEqual(t, 10, sumOfIdxs(table, o.TableIdxs),
			"5 で合計 10 の組が取れてしまっている")
	}
}

// **同ランクは合計 1 枚の特別扱いではない。** 数札では「合計がランクに等しい
// 部分集合」がそのまま同ランク捕獲を含む。
func TestDiloti_NumeralTakesEveryDisjointGroup(t *testing.T) {
	// 10 は 3+5、2、10 のすべてを同時に取れる。
	table := dcards([2]int{CardDesignSpade, 3}, [2]int{CardDesignHeart, 5},
		[2]int{CardDesignClover, 2}, [2]int{CardDesignDiamond, 10})
	opts := EnumerateDilotiTakes(dc(CardDesignSpade, 10), table, nil)

	best := 0
	for _, o := range opts {
		if len(o.TableIdxs) > best {
			best = len(o.TableIdxs)
		}
	}
	assert.Equal(t, 4, best, "3+5・2・10 をまとめて取る手が挙がっていない")
}

// **絵札はちょうど 1 枚。** 場に J が 2 枚あっても、J で取れるのは片方だけ。
func TestDiloti_FaceCardTakesExactlyOne(t *testing.T) {
	table := dcards([2]int{CardDesignSpade, 11}, [2]int{CardDesignHeart, 11},
		[2]int{CardDesignClover, 12})
	opts := EnumerateDilotiTakes(dc(CardDesignDiamond, 11), table, nil)
	require.Len(t, opts, 2, "J で取れる相手は 2 通りのはず")
	for _, o := range opts {
		assert.Len(t, o.TableIdxs, 1, "絵札が 2 枚まとめて取れてしまっている")
		assert.Empty(t, o.DeclIdxs)
	}
}

// **絵札は合計に混ざらない。** 場の K を 3+10 のように使えてはいけない。
func TestDiloti_FaceCardsNeverJoinASum(t *testing.T) {
	table := dcards([2]int{CardDesignSpade, 13}, [2]int{CardDesignHeart, 3})
	opts := EnumerateDilotiTakes(dc(CardDesignClover, 3), table, nil)
	for _, o := range opts {
		for _, i := range o.TableIdxs {
			assert.NotEqual(t, 13, table[i].GetValue(), "K が合計に使われている")
		}
	}
}

// **同ランクの絵札が場にあるなら、絵札は置けない。** 置けてしまうと、取れる札を
// 見送って場に絵札を積むだけの手が成立してしまう。
func TestDiloti_CannotTrailAFaceCardOntoItsMatch(t *testing.T) {
	table := dcards([2]int{CardDesignSpade, 12}, [2]int{CardDesignHeart, 4})
	assert.False(t, CanTrailDiloti(dc(CardDesignClover, 12), table),
		"同ランクの Q が場にあるのに Q を置けてしまう")
	assert.True(t, CanTrailDiloti(dc(CardDesignClover, 11), table))
	assert.True(t, CanTrailDiloti(dc(CardDesignClover, 4), table),
		"数札は取れても置ける (捕獲は強制ではない)")
}

// **宣言は値ちょうどの 1 枚で取る。**
func TestDiloti_DeclarationIsTakenByItsValue(t *testing.T) {
	decl := NewDilotiDeclaration(1, 8, dcards([2]int{CardDesignSpade, 3}, [2]int{CardDesignHeart, 5}))
	opts := EnumerateDilotiTakes(dc(CardDesignClover, 8), nil, []*DilotiDeclaration{decl})
	require.NotEmpty(t, opts)
	assert.Equal(t, []int{0}, opts[0].DeclIdxs)

	assert.Empty(t, EnumerateDilotiTakes(dc(CardDesignClover, 7), nil, []*DilotiDeclaration{decl}),
		"宣言値と違う札で宣言が取れてしまっている")
	// **絵札では宣言を取れない。** 宣言に絵札は入らず、値も 10 以下。
	assert.Empty(t, EnumerateDilotiTakes(dc(CardDesignClover, 11), nil, []*DilotiDeclaration{decl}))
}

// **グループ宣言は部分では取れない。** まとめて 1 枚で取るか、取らないか。
func TestDiloti_GroupDeclarationIsAllOrNothing(t *testing.T) {
	decl := NewDilotiDeclaration(1, 6, dcards([2]int{CardDesignSpade, 6}))
	decl.AddGroup(dcards([2]int{CardDesignHeart, 2}, [2]int{CardDesignClover, 4}))
	require.True(t, decl.IsGroup)
	assert.Len(t, decl.AllCards(), 3)

	opts := EnumerateDilotiTakes(dc(CardDesignDiamond, 6), nil, []*DilotiDeclaration{decl})
	require.Len(t, opts, 1)
	assert.Equal(t, []int{0}, opts[0].DeclIdxs, "グループ宣言が丸ごと取れない")
}

func TestIsValidDilotiCapture(t *testing.T) {
	table := dcards([2]int{CardDesignSpade, 2}, [2]int{CardDesignHeart, 3},
		[2]int{CardDesignClover, 5}, [2]int{CardDesignDiamond, 11})
	decl := NewDilotiDeclaration(1, 5, dcards([2]int{CardDesignSpade, 5}))
	decls := []*DilotiDeclaration{decl}

	// 5 で 2+3 と 5 と 宣言 5 を同時に。
	assert.True(t, IsValidDilotiCapture(dc(CardDesignHeart, 5), table, decls, []int{0, 1, 2}, []int{0}))
	// 合計が割り切れない選択は不可。
	assert.False(t, IsValidDilotiCapture(dc(CardDesignHeart, 5), table, decls, []int{0}, nil),
		"合計 2 を 5 で取れてしまっている")
	// **束に割れない選択は不可。** 合計 10 でも 5+5 に割れなければ駄目。
	table2 := dcards([2]int{CardDesignSpade, 1}, [2]int{CardDesignHeart, 9})
	assert.False(t, IsValidDilotiCapture(dc(CardDesignClover, 5), table2, nil, []int{0, 1}, nil),
		"1+9 が 5 の捕獲として通っている")
	// 絵札を巻き込む選択は不可。
	assert.False(t, IsValidDilotiCapture(dc(CardDesignHeart, 5), table, decls, []int{0, 1, 3}, nil))
	// 絵札は 1 枚だけ、同ランクだけ。
	assert.True(t, IsValidDilotiCapture(dc(CardDesignSpade, 11), table, nil, []int{3}, nil))
	assert.False(t, IsValidDilotiCapture(dc(CardDesignSpade, 12), table, nil, []int{3}, nil))
	// **同ランクの絵札が 2 枚あってもまとめては取れない。**
	twoJacks := dcards([2]int{CardDesignSpade, 11}, [2]int{CardDesignHeart, 11})
	assert.True(t, IsValidDilotiCapture(dc(CardDesignClover, 11), twoJacks, nil, []int{0}, nil))
	assert.False(t, IsValidDilotiCapture(dc(CardDesignClover, 11), twoJacks, nil, []int{0, 1}, nil),
		"J が 2 枚まとめて取れてしまっている")
	assert.False(t, IsValidDilotiCapture(dc(CardDesignSpade, 11), table, decls, []int{3}, []int{0}),
		"絵札が宣言を巻き込んでいる")
	// 範囲外は弾く。
	assert.False(t, IsValidDilotiCapture(dc(CardDesignHeart, 5), table, decls, []int{9}, nil))
	assert.False(t, IsValidDilotiCapture(dc(CardDesignHeart, 5), table, decls, nil, []int{9}))
	// 何も選ばないのは捕獲ではない。
	assert.False(t, IsValidDilotiCapture(dc(CardDesignHeart, 5), table, decls, nil, nil))
}

// **宣言値は 10 を超えられない。**
func TestEnumerateDilotiDeclarations(t *testing.T) {
	table := dcards([2]int{CardDesignSpade, 4}, [2]int{CardDesignHeart, 9})
	hand := dcards([2]int{CardDesignClover, 3}, [2]int{CardDesignDiamond, 7},
		[2]int{CardDesignSpade, 13})

	// 3 を出して場の 4 と組めば 7 の宣言。手札に 7 があるので合法。
	cands := EnumerateDilotiDeclarations(hand[0], 0, hand, table)
	require.NotEmpty(t, cands)
	for _, c := range cands {
		// **定数と定数を比べても実験にならない。** 上限が動けば期待値も動いて
		// しまうので、リテラルの 10 で見る。
		assert.LessOrEqual(t, c.Value, 10, "宣言値が 10 を超えている")
		assert.Greater(t, c.Value, DilotiCardValue(hand[0]), "出した札より小さい宣言")
	}
	assert.Equal(t, 10, DilotiMaxDeclaration, "宣言値の上限が 10 から動いている")
	assert.True(t, containsDeclValue(cands, 7), "3+4=7 の宣言が挙がっていない")

	// **10 を超える宣言は挙がらない。** 手札に 12 の裏付け札は存在し得ないので、
	// 上限が効いていることは「12 が出ない」では確かめられない ── 場に 9 を置き、
	// 3+9=12 の組が候補から落ちることを、裏付けの有無と切り離して見る。
	assert.False(t, containsDeclValue(cands, 12))
	bigTable := dcards([2]int{CardDesignSpade, 8})
	bigHand := dcards([2]int{CardDesignClover, 3}, [2]int{CardDesignDiamond, 10},
		[2]int{CardDesignHeart, 11})
	// 3+8=11 は上限超え。手札に 11 (J) はあるが絵札なので裏付けにもならない。
	assert.Empty(t, EnumerateDilotiDeclarations(bigHand[0], 0, bigHand, bigTable),
		"3+8=11 の宣言が通っている")

	// **裏付けの札を持たない宣言はできない。** 7 を捨てると 7 の宣言は消える。
	handNo7 := dcards([2]int{CardDesignClover, 3}, [2]int{CardDesignDiamond, 2})
	assert.False(t, containsDeclValue(
		EnumerateDilotiDeclarations(handNo7[0], 0, handNo7, table), 7),
		"手札に 7 が無いのに 7 を宣言できてしまう")

	// **絵札は宣言に使えない。**
	assert.Empty(t, EnumerateDilotiDeclarations(hand[2], 2, hand, table))
}

func containsDeclValue(cands []DilotiDeclCandidate, v int) bool {
	for _, c := range cands {
		if c.Value == v {
			return true
		}
	}
	return false
}

func sumOfIdxs(cards []*Card, idxs []int) int {
	n := 0
	for _, i := range idxs {
		n += DilotiCardValue(cards[i])
	}
	return n
}
