//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fbCard(design, value int) *Card { return NewCard(design, value, false) }

func newFreeBetForTest(t *testing.T) *FreeBetBlackjack {
	t.Helper()
	g := NewDefaultFreeBetBlackjack()
	g.Reset()
	return g
}

// fbStackNext は次に引かれる札を指定する。
//
// **引く札を配りに委ねると、検査そのものが配り依存になる。** FreeDouble は 1 枚
// 引いてその場で精算まで走るので、札を決めないと同じ assert が勝ち負けで通ったり
// 落ちたりする (実測 900 期待に対し 1200)。
func fbStackNext(g *FreeBetBlackjack, cards ...*Card) {
	for i, c := range cards {
		g.shoe.deck[g.shoe.deckDrawCnt+i] = c
	}
}

// fbStaged は指定の手を積んでプレイ待ちにした卓を返す。
func fbStaged(t *testing.T, ante int, player, dealer []*Card) *FreeBetBlackjack {
	t.Helper()
	g := newFreeBetForTest(t)
	// **最初の配りもこちらで決める。** ナチュラルが出ると PlaceBet がその場で
	// 精算まで走ってチップを動かし、この後で手札を差し替えても**チップだけが
	// 捨てたクープの結果を持ったまま**になる。9-8 なら双方 17 で決着しない。
	fbStackNext(g, fbCard(CardDesignSpade, 9), fbCard(CardDesignClover, 8),
		fbCard(CardDesignHeart, 9), fbCard(CardDesignDiamond, 8))
	require.NoError(t, g.PlaceBet(ante))
	require.Equal(t, FreeBetPhasePlay, g.phase, "積んだ配りで決着してしまった")

	h := NewBlackJackHand()
	for _, c := range player {
		h.AddCard(c)
	}
	h.SetBet(ante)
	g.hands = []*BlackJackHand{h}
	g.freeBets = []int{0}
	g.results = []FreeBetResult{FreeBetResultNone}
	g.activeHand = 0

	d := NewBlackJackHand()
	for _, c := range dealer {
		d.AddCard(c)
	}
	g.dealerHand = d
	g.phase = FreeBetPhasePlay
	return g
}

// --- 無料ダブルの条件 ---

// **ハードの 9・10・11 だけ。** ソフトも 3 枚目以降も対象外。
func TestFreeBet_CanFreeDouble(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name  string
		score int
		soft  bool
		cards int
		want  bool
	}{
		{"ハード 9 は可", 9, false, 2, true},
		{"ハード 10 は可", 10, false, 2, true},
		{"ハード 11 は可", 11, false, 2, true},
		{"ハード 8 は不可", 8, false, 2, false},
		{"ハード 12 は不可", 12, false, 2, false},
		{"ソフト 11 は不可", 11, true, 2, false},
		{"ソフト 9 は不可", 9, true, 2, false},
		{"3 枚目以降は不可", 10, false, 3, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FreeBetCanFreeDouble(tt.score, tt.soft, tt.cards))
		})
	}
}

// **10 点札のペアは割れない。**
func TestFreeBet_CanFreeSplit(t *testing.T) {
	t.Parallel()

	assert.True(t, FreeBetCanFreeSplit(8, 8))
	assert.True(t, FreeBetCanFreeSplit(1, 1), "エースのペアは割れる")
	assert.False(t, FreeBetCanFreeSplit(8, 9), "ペアでない")
	assert.False(t, FreeBetCanFreeSplit(10, 10), "10 のペアは割れない")
	assert.False(t, FreeBetCanFreeSplit(13, 13), "K のペアも 10 点なので割れない")
	assert.False(t, FreeBetCanFreeSplit(12, 12), "Q のペアも割れない")
}

// --- 無料ダブル ---

// **プレイヤーはチップを出さない。** 勝てば倍額、負けても元の賭け金だけ。
func TestFreeBet_FreeDoubleCostsNothing(t *testing.T) {
	g := fbStaged(t, 100,
		[]*Card{fbCard(CardDesignSpade, 6), fbCard(CardDesignHeart, 5)}, // ハード 11
		[]*Card{fbCard(CardDesignClover, 13), fbCard(CardDesignDiamond, 7)})
	before := g.GetChips()
	require.True(t, g.CanFreeDouble())
	// **負ける札を積む。** FreeDouble は 1 枚引いて打ち止めにするので、ここで
	// 勝ってしまうと配当がチップを押し上げ、「ハウスの出資をプレイヤーから
	// 引いていないか」という肝心の問いに答えられなくなる。
	fbStackNext(g, fbCard(CardDesignClover, 2)) // 11+2 = 13 でディーラー 17 に負け

	require.NoError(t, g.FreeDouble())
	assert.Equal(t, 100, g.GetFreeBet(0), "ハウスの出資が乗っていない")
	assert.Equal(t, 100, g.GetHands()[0].GetBet(), "プレイヤーの賭け金が変わっている")
	assert.Equal(t, FreeBetResultLose, g.GetResults()[0], "積んだ札で負けていない")
	assert.Equal(t, before, g.GetChips(), "無料ダブルでチップが減っている")
}

func TestFreeBet_FreeDoubleRejectedOutsideTheRange(t *testing.T) {
	g := fbStaged(t, 100,
		[]*Card{fbCard(CardDesignSpade, 10), fbCard(CardDesignHeart, 2)}, // ハード 12
		[]*Card{fbCard(CardDesignClover, 13), fbCard(CardDesignDiamond, 7)})
	assert.False(t, g.CanFreeDouble())
	assert.ErrorIs(t, g.FreeDouble(), errFreeBetCannotDouble)
}

// **負けても元の賭け金しか失わない。**
func TestFreeBet_FreeDoubleLosesOnlyTheOriginalBet(t *testing.T) {
	g := fbStaged(t, 100,
		[]*Card{fbCard(CardDesignSpade, 6), fbCard(CardDesignHeart, 5)},
		[]*Card{fbCard(CardDesignClover, 13), fbCard(CardDesignDiamond, 9)}) // 19
	g.freeBets[0] = 100
	g.hands[0].AddCard(fbCard(CardDesignSpade, 2)) // 13 で負け
	g.hands[0].SetStood(true)

	r, ret := g.settleHand(g.hands[0], 100, 19, false)
	assert.Equal(t, FreeBetResultLose, r)
	assert.Zero(t, ret, "負けたのに払い戻しがある")
}

// **勝てばハウスのぶんも配当として付く。**
func TestFreeBet_FreeDoubleWinPaysTheHouseShare(t *testing.T) {
	g := fbStaged(t, 100,
		[]*Card{fbCard(CardDesignSpade, 6), fbCard(CardDesignHeart, 5)},
		[]*Card{fbCard(CardDesignClover, 13), fbCard(CardDesignDiamond, 7)}) // 17
	g.hands[0].AddCard(fbCard(CardDesignSpade, 9)) // 20 で勝ち
	g.hands[0].SetStood(true)

	r, ret := g.settleHand(g.hands[0], 100, 17, false)
	assert.Equal(t, FreeBetResultWin, r)
	// 賭け 100 の返却 + 100 の配当 + ハウス 100 の配当 = 300
	assert.Equal(t, 300, ret)
}

// --- 無料スプリット ---

// **2 つ目の手札はまるごとハウスの金。**
func TestFreeBet_FreeSplitCostsNothing(t *testing.T) {
	g := fbStaged(t, 100,
		[]*Card{fbCard(CardDesignSpade, 8), fbCard(CardDesignHeart, 8)},
		[]*Card{fbCard(CardDesignClover, 13), fbCard(CardDesignDiamond, 7)})
	before := g.GetChips()
	require.True(t, g.CanFreeSplit())
	// 分けた 2 手に配る札も決めておく (8+10 = 18 が 2 つ)。
	fbStackNext(g, fbCard(CardDesignClover, 10), fbCard(CardDesignDiamond, 10))

	require.NoError(t, g.FreeSplit())
	assert.Equal(t, before, g.GetChips(), "無料スプリットでチップが減っている")
	assert.Equal(t, 2, g.GetHandCount())
	assert.Equal(t, 100, g.GetHands()[0].GetBet(), "1 つ目はプレイヤーの金のまま")
	assert.Zero(t, g.GetHands()[1].GetBet(), "2 つ目にプレイヤーの金が乗っている")
	assert.Equal(t, 100, g.GetFreeBet(1), "2 つ目がハウス持ちになっていない")
}

func TestFreeBet_FreeSplitRejectedOnTens(t *testing.T) {
	g := fbStaged(t, 100,
		[]*Card{fbCard(CardDesignSpade, 13), fbCard(CardDesignHeart, 12)},
		[]*Card{fbCard(CardDesignClover, 13), fbCard(CardDesignDiamond, 7)})
	assert.False(t, g.CanFreeSplit(), "10 点札のペアが割れてしまう")
	assert.ErrorIs(t, g.FreeSplit(), errFreeBetCannotSplit)
}

func TestFreeBet_SplitAcesGetOneCardEach(t *testing.T) {
	g := fbStaged(t, 100,
		[]*Card{fbCard(CardDesignSpade, 1), fbCard(CardDesignHeart, 1)},
		[]*Card{fbCard(CardDesignClover, 13), fbCard(CardDesignDiamond, 7)})
	fbStackNext(g, fbCard(CardDesignClover, 9), fbCard(CardDesignDiamond, 9))
	require.NoError(t, g.FreeSplit())

	for i, h := range g.GetHands() {
		assert.True(t, h.IsStood(), "手札 %d が打ち止めになっていない", i)
		assert.Equal(t, 2, h.GetCardsSize())
	}
}

// **ハウス持ちの手札で負けても、プレイヤーは何も失わない。**
func TestFreeBet_FreeSplitHandLosesNothing(t *testing.T) {
	g := fbStaged(t, 100,
		[]*Card{fbCard(CardDesignSpade, 8), fbCard(CardDesignHeart, 8)},
		[]*Card{fbCard(CardDesignClover, 13), fbCard(CardDesignDiamond, 9)})
	// **引く札を決める。** 素の配りだと 8 にエースが付いて 19 になり、
	// ディーラー 19 と同点で勝敗が引き分けに化ける (実測)。
	fbStackNext(g, fbCard(CardDesignClover, 10), fbCard(CardDesignDiamond, 10))
	require.NoError(t, g.FreeSplit())

	// 2 つ目 (bet=0, free=100) が負けても払い戻しは 0 = 失うものが無い。
	r, ret := g.settleHand(g.hands[1], g.GetFreeBet(1), 19, false)
	assert.Equal(t, FreeBetResultLose, r)
	assert.Zero(t, ret)

	// 勝てばハウスのぶんが配当として付く。18 対 17 なので必ずこちらを通る。
	g.hands[1].SetStood(true)
	r2, ret2 := g.settleHand(g.hands[1], g.GetFreeBet(1), 17, false)
	require.Equal(t, FreeBetResultWin, r2)
	assert.Equal(t, 100, ret2, "ハウス持ちの手札の配当が違う")
}

// --- ディーラー 22 プッシュ ---

// **これが無料ダブル / 無料スプリットの対価。**
//
// ここを外すとハウスエッジが消える。22 だけが特別で、23 以上は普通のバスト。
func TestFreeBet_Dealer22Pushes(t *testing.T) {
	tests := []struct {
		name       string
		dealer     []*Card
		wantResult FreeBetResult
		wantPayout int
		wantPushed bool
	}{
		{
			name: "22 でバストは引き分け",
			dealer: []*Card{fbCard(CardDesignClover, 13), fbCard(CardDesignDiamond, 5),
				fbCard(CardDesignSpade, 7)}, // 10+5+7 = 22
			wantResult: FreeBetResultDealer22Push, wantPayout: 100, wantPushed: true,
		},
		{
			name: "23 でバストは普通に勝ち",
			dealer: []*Card{fbCard(CardDesignClover, 13), fbCard(CardDesignDiamond, 6),
				fbCard(CardDesignSpade, 7)}, // 10+6+7 = 23
			wantResult: FreeBetResultWin, wantPayout: 200, wantPushed: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := fbStaged(t, 100,
				[]*Card{fbCard(CardDesignSpade, 10), fbCard(CardDesignHeart, 9)}, // 19
				tt.dealer)
			g.hands[0].SetStood(true)
			before := g.GetChips()

			g.settle()

			assert.Equal(t, tt.wantPushed, g.IsDealerPushed22())
			assert.Equal(t, tt.wantResult, g.GetResults()[0])
			assert.Equal(t, before+tt.wantPayout, g.GetChips())
		})
	}
}

// **プレイヤーがバストしていれば、ディーラーの 22 は関係ない。**
func TestFreeBet_PlayerBustLosesEvenOnDealer22(t *testing.T) {
	g := fbStaged(t, 100,
		[]*Card{fbCard(CardDesignSpade, 10), fbCard(CardDesignHeart, 9), fbCard(CardDesignClover, 5)},
		[]*Card{fbCard(CardDesignClover, 13), fbCard(CardDesignDiamond, 5), fbCard(CardDesignSpade, 7)})
	g.hands[0].SetBusted(true)
	before := g.GetChips()

	g.settle()
	assert.Equal(t, FreeBetResultLose, g.GetResults()[0])
	assert.Equal(t, before, g.GetChips(), "バストしたのに払い戻された")
}

// **22 プッシュではハウスのぶんは払われない。**
func TestFreeBet_Dealer22DoesNotPayTheFreeShare(t *testing.T) {
	g := fbStaged(t, 100,
		[]*Card{fbCard(CardDesignSpade, 6), fbCard(CardDesignHeart, 5)},
		[]*Card{fbCard(CardDesignClover, 13), fbCard(CardDesignDiamond, 5), fbCard(CardDesignSpade, 7)})
	g.freeBets[0] = 100
	g.hands[0].AddCard(fbCard(CardDesignSpade, 9)) // 20
	g.hands[0].SetStood(true)
	before := g.GetChips()

	g.settle()
	assert.Equal(t, FreeBetResultDealer22Push, g.GetResults()[0])
	// 賭け金 100 が返るだけ。ハウスの 100 は配当にならない。
	assert.Equal(t, before+100, g.GetChips())
}

// --- 通常の配当 ---

func TestFreeBet_BlackjackPaysThreeToTwo(t *testing.T) {
	g := fbStaged(t, 100,
		[]*Card{fbCard(CardDesignSpade, 1), fbCard(CardDesignHeart, 13)},
		[]*Card{fbCard(CardDesignClover, 10), fbCard(CardDesignDiamond, 9)})
	before := g.GetChips()

	g.settle()
	assert.Equal(t, FreeBetResultBlackjack, g.GetResults()[0])
	// 100 の返却 + 150 の配当 = 250
	assert.Equal(t, before+250, g.GetChips())
}

func TestFreeBet_WinLosePush(t *testing.T) {
	for _, tt := range []struct {
		name           string
		player, dealer []*Card
		want           FreeBetResult
		wantPayout     int
	}{
		{
			name:   "高いほうが勝つ",
			player: []*Card{fbCard(CardDesignSpade, 10), fbCard(CardDesignHeart, 9)},
			dealer: []*Card{fbCard(CardDesignClover, 10), fbCard(CardDesignDiamond, 8)},
			want:   FreeBetResultWin, wantPayout: 200,
		},
		{
			name:   "低いほうが負ける",
			player: []*Card{fbCard(CardDesignSpade, 10), fbCard(CardDesignHeart, 7)},
			dealer: []*Card{fbCard(CardDesignClover, 10), fbCard(CardDesignDiamond, 8)},
			want:   FreeBetResultLose, wantPayout: 0,
		},
		{
			name:   "同点はプッシュ",
			player: []*Card{fbCard(CardDesignSpade, 10), fbCard(CardDesignHeart, 8)},
			dealer: []*Card{fbCard(CardDesignClover, 10), fbCard(CardDesignDiamond, 8)},
			want:   FreeBetResultPush, wantPayout: 100,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := fbStaged(t, 100, tt.player, tt.dealer)
			g.hands[0].SetStood(true)
			before := g.GetChips()
			g.settle()
			assert.Equal(t, tt.want, g.GetResults()[0])
			assert.Equal(t, before+tt.wantPayout, g.GetChips())
		})
	}
}

// --- ディーラーの進行 ---

func TestFreeBet_DealerHitsSoft17(t *testing.T) {
	g := fbStaged(t, 100,
		[]*Card{fbCard(CardDesignSpade, 10), fbCard(CardDesignHeart, 9)},
		[]*Card{fbCard(CardDesignClover, 1), fbCard(CardDesignDiamond, 6)}) // ソフト 17
	g.hands[0].SetStood(true)
	g.dealerPlay()
	assert.Greater(t, len(g.GetDealerCards()), 2, "ソフト 17 で止まっている")
}

func TestFreeBet_DealerStandsOnHard17(t *testing.T) {
	g := fbStaged(t, 100,
		[]*Card{fbCard(CardDesignSpade, 10), fbCard(CardDesignHeart, 9)},
		[]*Card{fbCard(CardDesignClover, 13), fbCard(CardDesignDiamond, 7)})
	g.hands[0].SetStood(true)
	g.dealerPlay()
	assert.Len(t, g.GetDealerCards(), 2, "ハード 17 で引いている")
}

// --- 進行 ---

func TestFreeBet_PlaceBetValidation(t *testing.T) {
	for _, tt := range []struct {
		name    string
		ante    int
		wantErr error
	}{
		{"最低額は通る", FreeBetAnteMin, nil},
		{"上限額は通る", FreeBetAnteMax, nil},
		{"低すぎる", FreeBetAnteMin - 1, errFreeBetAnteRange},
		{"高すぎる", FreeBetAnteMax + 1, errFreeBetAnteRange},
		{"刻みに合わない", 15, errFreeBetAnteUnit},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := newFreeBetForTest(t)
			g.SetChips(FreeBetChipsMax)
			err := g.PlaceBet(tt.ante)
			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestFreeBet_PhaseGuards(t *testing.T) {
	g := newFreeBetForTest(t)
	assert.ErrorIs(t, g.Hit(), errFreeBetWrongPhase)
	assert.ErrorIs(t, g.Stand(), errFreeBetWrongPhase)
	assert.ErrorIs(t, g.FreeDouble(), errFreeBetWrongPhase)
	assert.ErrorIs(t, g.FreeSplit(), errFreeBetWrongPhase)
	assert.ErrorIs(t, g.NextRound(), errFreeBetWrongPhase)
}

func TestFreeBet_EndsWhenOutOfChips(t *testing.T) {
	g := newFreeBetForTest(t)
	require.NoError(t, g.PlaceBet(FreeBetAnteMin))
	for g.GetPhase() == FreeBetPhasePlay {
		require.NoError(t, g.Stand())
	}
	g.SetChips(0)
	require.NoError(t, g.NextRound())
	assert.True(t, g.GetGameEndFlag())
	assert.ErrorIs(t, g.PlaceBet(FreeBetAnteMin), errFreeBetFinished)
}

// **通しで回してもチップが負にならない。**
func TestFreeBet_FullSessionsStayConsistent(t *testing.T) {
	for seed := range 30 {
		g := newFreeBetForTest(t)
		g.SetChips(5000)
		for step := 0; step < 300 && !g.GetGameEndFlag(); step++ {
			switch g.GetPhase() {
			case FreeBetPhaseBet:
				require.NoError(t, g.PlaceBet(FreeBetAnteMin))
			case FreeBetPhasePlay:
				switch {
				case g.CanFreeSplit():
					require.NoError(t, g.FreeSplit())
				case g.CanFreeDouble():
					require.NoError(t, g.FreeDouble())
				default:
					require.NoError(t, g.Stand())
				}
			default:
				require.NoError(t, g.NextRound())
			}
			require.GreaterOrEqual(t, g.GetChips(), 0, "seed %d: チップが負になった", seed)
			require.Equal(t, len(g.GetHands()), len(g.GetFreeBets()), "seed %d: 手札とハウス出資の数が合わない", seed)
			require.Equal(t, len(g.GetHands()), len(g.GetResults()), "seed %d: 手札と決着の数が合わない", seed)
		}
	}
}

// --- ヒント ---

func TestFreeBet_GetHint(t *testing.T) {
	t.Run("賭ける前は助言しない", func(t *testing.T) {
		assert.Nil(t, newFreeBetForTest(t).GetHint())
	})

	// **タダなら使う。** 負けても元の賭け金しか失わない。
	t.Run("無料スプリットできるなら薦める", func(t *testing.T) {
		g := fbStaged(t, 100,
			[]*Card{fbCard(CardDesignSpade, 8), fbCard(CardDesignHeart, 8)},
			[]*Card{fbCard(CardDesignClover, 13), fbCard(CardDesignDiamond, 7)})
		h := g.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "freeSplit", h.Action)
		assert.Equal(t, "freeIsFree", h.Reason)
	})

	t.Run("無料ダブルできるなら薦める", func(t *testing.T) {
		g := fbStaged(t, 100,
			[]*Card{fbCard(CardDesignSpade, 6), fbCard(CardDesignHeart, 5)},
			[]*Card{fbCard(CardDesignClover, 13), fbCard(CardDesignDiamond, 7)})
		h := g.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "freeDouble", h.Action)
	})

	t.Run("17 以上は立つ", func(t *testing.T) {
		g := fbStaged(t, 100,
			[]*Card{fbCard(CardDesignSpade, 10), fbCard(CardDesignHeart, 8)},
			[]*Card{fbCard(CardDesignClover, 13), fbCard(CardDesignDiamond, 7)})
		h := g.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "stand", h.Action)
	})
}

// --- 名前と設定 ---

func TestFreeBetNames(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "bet", FreeBetPhaseName(FreeBetPhaseBet))
	assert.Equal(t, "play", FreeBetPhaseName(FreeBetPhasePlay))
	assert.Equal(t, "result", FreeBetPhaseName(FreeBetPhaseResult))

	assert.Equal(t, "win", FreeBetResultName(FreeBetResultWin))
	assert.Equal(t, "lose", FreeBetResultName(FreeBetResultLose))
	assert.Equal(t, "push", FreeBetResultName(FreeBetResultPush))
	assert.Equal(t, "blackjack", FreeBetResultName(FreeBetResultBlackjack))
	assert.Equal(t, "dealer22Push", FreeBetResultName(FreeBetResultDealer22Push))
	assert.Equal(t, "none", FreeBetResultName(FreeBetResultNone))
}

func TestFreeBetConfig_Validate(t *testing.T) {
	t.Parallel()

	assert.NoError(t, DefaultFreeBetBlackjackConfig().Validate())
	for _, cfg := range []FreeBetBlackjackConfig{
		{InitialChips: FreeBetChipsMin - 1, DefaultAnte: 50},
		{InitialChips: FreeBetChipsMax + 1, DefaultAnte: 50},
		{InitialChips: 1000, DefaultAnte: FreeBetAnteMin - 1},
		{InitialChips: 1000, DefaultAnte: FreeBetAnteMax + 1},
	} {
		assert.Error(t, cfg.Validate())
	}
}

func TestFreeBet_Accessors(t *testing.T) {
	g := newFreeBetForTest(t)
	assert.Equal(t, 52*FreeBetDeckCount, g.GetRemainingCards())
	assert.NotNil(t, g.GetPlayer())
	assert.NotEmpty(t, g.GetActionLog())
	assert.Zero(t, g.GetHandCount())
	assert.Nil(t, g.GetDealerCards())
	assert.Zero(t, g.GetDealerScore())
	assert.Zero(t, g.GetFreeBet(0), "範囲外は 0")
	assert.Zero(t, g.GetFreeBet(-1))

	require.NoError(t, g.PlaceBet(50))
	assert.Equal(t, 50, g.GetAnteBet())
	assert.Equal(t, 1, g.GetHandCount())
	assert.Len(t, g.GetDealerCards(), 2)

	g.SetConfig(FreeBetBlackjackConfig{InitialChips: 500, DefaultAnte: 20})
	assert.Equal(t, 500, g.GetConfig().InitialChips)
}

func TestFreeBet_ActionLogIsBounded(t *testing.T) {
	g := newFreeBetForTest(t)
	for range freeBetMaxSliceLen + 50 {
		g.appendLog("noise", "x", nil)
	}
	assert.Len(t, g.GetActionLog(), freeBetMaxSliceLen)
}
