//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// **半分だけ取った席で盤面が止まってはいけない (#5435 の実測)。**
//
// ショーダウンは「人間が勝てなかったとき」だけマックを訊くために留まる。
// ところが分割ポットでは、留まる条件が**ハイの取り分だけ**を見ているのに、
// マックを出す条件は**合計の取り分**を見ていた ── ローやスペードで半分を
// 取った席はその食い違いのちょうど間に落ちる:
//
//   - 留まる (humanLost = ハイが 0 なので true) → END へ進まない
//   - 訊かない (IsMuckAvailable = 合計が 0 でないので false)
//
// Web は `muckAvailable` でボタンを出すので、**押せる手が 1 つも無い画面**に
// なる。CUI は `show` を打てば抜けられるが、勧めるものが何も出ない。
func TestSevenCardStudHiLo_WinningOnlyTheLowDoesNotFreezeTheShowdown(t *testing.T) {
	s := hlTable(t, 400)

	// 人間 (席 0): 8 以下のローだけが強い。ハイは弱い。
	hlDeal(s, 0,
		hlCard(CardDesignSpade, 1), hlCard(CardDesignHeart, 2), hlCard(CardDesignClover, 3),
		hlCard(CardDesignDiamond, 4), hlCard(CardDesignSpade, 6),
		hlCard(CardDesignHeart, 9), hlCard(CardDesignClover, 10))
	// 席 1: フラッシュ。ハイは確実にこちら。ローは無い。
	hlDeal(s, 1,
		hlCard(CardDesignDiamond, 13), hlCard(CardDesignDiamond, 11), hlCard(CardDesignDiamond, 9),
		hlCard(CardDesignDiamond, 7), hlCard(CardDesignDiamond, 5),
		hlCard(CardDesignSpade, 12), hlCard(CardDesignHeart, 12))
	for _, idx := range []int{2, 3} {
		s.players[idx].SetFolded(true)
	}

	s.SetPhase(SevenCardStudPhaseShowdown)
	s.resolveShowdown()

	require.Positive(t, hlWon(s, 0), "人間はローで半分を取っている")
	assert.False(t, s.IsMuckAvailable(),
		"取り分がある席にマックを訊く理由は無い")
	assert.Equal(t, SevenCardStudPhaseEnd, s.GetPhase(),
		"訊かないなら留まってもいけない (押せる手が無い画面になる)")
}

// 何も取れなかった席には今までどおりマックを訊く。**負のコントロール** ──
// 上の修正で「常に END へ進む」にしてしまうと、マックの面ごと消える。
func TestSevenCardStudHiLo_WinningNothingStillOffersTheMuck(t *testing.T) {
	s := hlTable(t, 400)

	// 人間 (席 0): ハイもローも無い。
	hlDeal(s, 0,
		hlCard(CardDesignSpade, 9), hlCard(CardDesignHeart, 10), hlCard(CardDesignClover, 11),
		hlCard(CardDesignDiamond, 2), hlCard(CardDesignSpade, 13),
		hlCard(CardDesignHeart, 4), hlCard(CardDesignClover, 6))
	// 席 1: フラッシュ + ホイールのロー。スクープする。
	hlDeal(s, 1,
		hlCard(CardDesignDiamond, 1), hlCard(CardDesignDiamond, 2), hlCard(CardDesignDiamond, 3),
		hlCard(CardDesignDiamond, 4), hlCard(CardDesignDiamond, 5),
		hlCard(CardDesignSpade, 12), hlCard(CardDesignHeart, 12))
	for _, idx := range []int{2, 3} {
		s.players[idx].SetFolded(true)
	}

	s.SetPhase(SevenCardStudPhaseShowdown)
	s.resolveShowdown()

	require.Zero(t, hlWon(s, 0), "人間は 1 チップも取っていない")
	assert.True(t, s.IsMuckAvailable(), "負けた席にはマックを訊く")
	assert.Equal(t, SevenCardStudPhaseShowdown, s.GetPhase(), "決めるまで留まる")
}
