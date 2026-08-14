//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func blHand(cards ...*Card) *BlackJackHand {
	h := NewBlackJackHand()
	for _, c := range cards {
		h.AddCard(c)
	}
	return h
}

func blCard(design, value int) *Card { return NewCard(design, value, false) }

// --- 役の判定 ---

func TestBanLuck_EvalHand(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		cards []*Card
		want  BanLuckRank
	}{
		{
			// **A+A の合計は 21 ではない (12)。** 合計値だけ見ていると取りこぼす。
			name:  "A+A は Ban Ban",
			cards: []*Card{blCard(CardDesignSpade, 1), blCard(CardDesignHeart, 1)},
			want:  BanLuckRankBanBan,
		},
		{
			name:  "2 枚 21 は Ban Luck",
			cards: []*Card{blCard(CardDesignSpade, 1), blCard(CardDesignHeart, 13)},
			want:  BanLuckRankBanLuck,
		},
		{
			// **引いて到達した 21 は Ban Luck ではない。**
			name: "3 枚 21 は普通の手",
			cards: []*Card{blCard(CardDesignSpade, 7), blCard(CardDesignHeart, 7),
				blCard(CardDesignClover, 7)},
			want: BanLuckRankPoint,
		},
		{
			// **合計が低くても Five Dragon。** 16 の 5 枚が 20 の 2 枚に勝つ。
			name: "5 枚 21 以下は Five Dragon",
			cards: []*Card{blCard(CardDesignSpade, 2), blCard(CardDesignHeart, 3),
				blCard(CardDesignClover, 4), blCard(CardDesignDiamond, 3),
				blCard(CardDesignSpade, 4)}, // 16
			want: BanLuckRankFiveDragon,
		},
		{
			name: "5 枚で 21 ちょうども Five Dragon",
			cards: []*Card{blCard(CardDesignSpade, 2), blCard(CardDesignHeart, 3),
				blCard(CardDesignClover, 4), blCard(CardDesignDiamond, 5),
				blCard(CardDesignSpade, 7)}, // 21
			want: BanLuckRankFiveDragon,
		},
		{
			// **枚数が揃っていても 21 を超えたらバスト。** Five Dragon にならない。
			name: "5 枚で 21 超えはバスト",
			cards: []*Card{blCard(CardDesignSpade, 5), blCard(CardDesignHeart, 5),
				blCard(CardDesignClover, 5), blCard(CardDesignDiamond, 5),
				blCard(CardDesignSpade, 5)}, // 25
			want: BanLuckRankBust,
		},
		{
			name:  "普通の手",
			cards: []*Card{blCard(CardDesignSpade, 10), blCard(CardDesignHeart, 8)},
			want:  BanLuckRankPoint,
		},
		{
			name:  "21 超えはバスト",
			cards: []*Card{blCard(CardDesignSpade, 10), blCard(CardDesignHeart, 8), blCard(CardDesignClover, 9)},
			want:  BanLuckRankBust,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, EvalBanLuckHand(blHand(tt.cards...)))
		})
	}

	t.Run("nil はバスト扱い", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, BanLuckRankBust, EvalBanLuckHand(nil))
	})
}

// **役の順序そのものが規則。** 値の並びを入れ替えると比較が静かに壊れる。
func TestBanLuck_RankOrder(t *testing.T) {
	t.Parallel()
	assert.Less(t, BanLuckRankBust, BanLuckRankPoint)
	assert.Less(t, BanLuckRankPoint, BanLuckRankFiveDragon)
	assert.Less(t, BanLuckRankFiveDragon, BanLuckRankBanLuck)
	assert.Less(t, BanLuckRankBanLuck, BanLuckRankBanBan)
	assert.Equal(t, BanLuckRankBanBan, BanLuckRankMax)
}

func TestBanLuck_RankNames(t *testing.T) {
	t.Parallel()
	for r, want := range map[BanLuckRank]string{
		BanLuckRankBust:       "bust",
		BanLuckRankPoint:      "point",
		BanLuckRankFiveDragon: "fiveDragon",
		BanLuckRankBanLuck:    "banLuck",
		BanLuckRankBanBan:     "banBan",
		BanLuckRank(99):       "bust",
	} {
		assert.Equal(t, want, BanLuckRankName(r))
	}
}

func TestBanLuck_PayoutFor(t *testing.T) {
	t.Parallel()
	assert.Equal(t, BanLuckPayoutBanBan, BanLuckPayoutFor(BanLuckRankBanBan))
	assert.Equal(t, BanLuckPayoutBanLuck, BanLuckPayoutFor(BanLuckRankBanLuck))
	assert.Equal(t, BanLuckPayoutFiveDragon, BanLuckPayoutFor(BanLuckRankFiveDragon))
	assert.Equal(t, BanLuckPayoutNormal, BanLuckPayoutFor(BanLuckRankPoint))
	assert.Equal(t, BanLuckPayoutNormal, BanLuckPayoutFor(BanLuckRankBust))
}

// --- 子と親の比較 ---

func TestBanLuck_Compare(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name                     string
		player, banker           BanLuckRank
		playerScore, bankerScore int
		wantOutcome              BanLuckOutcome
		wantMult                 int
	}{
		{
			// **子のバストが最優先。** 親が後で飛んでも救われない。
			name: "両方バストなら子の負け", player: BanLuckRankBust, banker: BanLuckRankBust,
			playerScore: 25, bankerScore: 24,
			wantOutcome: BanLuckOutcomeLose, wantMult: BanLuckPayoutNormal,
		},
		{
			name: "子だけバストなら負け", player: BanLuckRankBust, banker: BanLuckRankPoint,
			playerScore: 25, bankerScore: 17,
			wantOutcome: BanLuckOutcomeLose, wantMult: BanLuckPayoutNormal,
		},
		{
			name: "親だけバストなら勝ち", player: BanLuckRankPoint, banker: BanLuckRankBust,
			playerScore: 15, bankerScore: 25,
			wantOutcome: BanLuckOutcomeWin, wantMult: BanLuckPayoutNormal,
		},
		{
			// **倍率は勝った側の役で決まる。**
			name: "子の Ban Ban は 3 倍", player: BanLuckRankBanBan, banker: BanLuckRankPoint,
			playerScore: 12, bankerScore: 20,
			wantOutcome: BanLuckOutcomeWin, wantMult: BanLuckPayoutBanBan,
		},
		{
			name: "親の Ban Ban も 3 倍取る", player: BanLuckRankPoint, banker: BanLuckRankBanBan,
			playerScore: 20, bankerScore: 12,
			wantOutcome: BanLuckOutcomeLose, wantMult: BanLuckPayoutBanBan,
		},
		{
			name: "Ban Luck は Five Dragon に勝つ", player: BanLuckRankBanLuck, banker: BanLuckRankFiveDragon,
			playerScore: 21, bankerScore: 21,
			wantOutcome: BanLuckOutcomeWin, wantMult: BanLuckPayoutBanLuck,
		},
		{
			// **合計が低くても Five Dragon が勝つ。** ここがこのゲームの見どころ。
			name: "Five Dragon 16 は普通の 20 に勝つ", player: BanLuckRankFiveDragon, banker: BanLuckRankPoint,
			playerScore: 16, bankerScore: 20,
			wantOutcome: BanLuckOutcomeWin, wantMult: BanLuckPayoutFiveDragon,
		},
		{
			// **特別役同士は合計を比べない。**
			name: "Five Dragon 同士は合計が違っても引き分け", player: BanLuckRankFiveDragon, banker: BanLuckRankFiveDragon,
			playerScore: 16, bankerScore: 21,
			wantOutcome: BanLuckOutcomePush, wantMult: BanLuckPayoutFiveDragon,
		},
		{
			name: "Ban Luck 同士は引き分け", player: BanLuckRankBanLuck, banker: BanLuckRankBanLuck,
			playerScore: 21, bankerScore: 21,
			wantOutcome: BanLuckOutcomePush, wantMult: BanLuckPayoutBanLuck,
		},
		{
			name: "Ban Ban 同士は引き分け", player: BanLuckRankBanBan, banker: BanLuckRankBanBan,
			playerScore: 12, bankerScore: 12,
			wantOutcome: BanLuckOutcomePush, wantMult: BanLuckPayoutBanBan,
		},
		{
			name: "普通の手は合計で決まる (勝ち)", player: BanLuckRankPoint, banker: BanLuckRankPoint,
			playerScore: 20, bankerScore: 17,
			wantOutcome: BanLuckOutcomeWin, wantMult: BanLuckPayoutNormal,
		},
		{
			name: "普通の手は合計で決まる (負け)", player: BanLuckRankPoint, banker: BanLuckRankPoint,
			playerScore: 16, bankerScore: 17,
			wantOutcome: BanLuckOutcomeLose, wantMult: BanLuckPayoutNormal,
		},
		{
			// **同点は引き分け。** 親の総取りにはしない。
			name: "同点は引き分け", player: BanLuckRankPoint, banker: BanLuckRankPoint,
			playerScore: 18, bankerScore: 18,
			wantOutcome: BanLuckOutcomePush, wantMult: BanLuckPayoutNormal,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, mult := CompareBanLuck(tt.player, tt.banker, tt.playerScore, tt.bankerScore)
			assert.Equal(t, tt.wantOutcome, got)
			assert.Equal(t, tt.wantMult, mult)
		})
	}
}

// **比較は必ず対称。** 子と親を入れ替えたら勝敗も裏返り、倍率は変わらない。
// 片側だけを場合分けで書くと、ここが崩れる。
func TestBanLuck_CompareIsSymmetric(t *testing.T) {
	t.Parallel()
	ranks := []BanLuckRank{
		BanLuckRankBust, BanLuckRankPoint, BanLuckRankFiveDragon,
		BanLuckRankBanLuck, BanLuckRankBanBan,
	}
	scores := []int{14, 18, 21}
	for _, a := range ranks {
		for _, b := range ranks {
			for _, sa := range scores {
				for _, sb := range scores {
					if a == BanLuckRankBust && b == BanLuckRankBust {
						// 両方バストだけは「子が先に飛んでいる」ので非対称。
						continue
					}
					out, mult := CompareBanLuck(a, b, sa, sb)
					rev, revMult := CompareBanLuck(b, a, sb, sa)
					assert.Equal(t, mult, revMult, "%v vs %v の倍率が入れ替えで変わった", a, b)
					switch out {
					case BanLuckOutcomeWin:
						assert.Equal(t, BanLuckOutcomeLose, rev, "%v vs %v", a, b)
					case BanLuckOutcomeLose:
						assert.Equal(t, BanLuckOutcomeWin, rev, "%v vs %v", a, b)
					default:
						assert.Equal(t, BanLuckOutcomePush, rev, "%v vs %v", a, b)
					}
				}
			}
		}
	}
}
