//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func daCard(design, value int) *Card { return NewCard(design, value, false) }

func newDoubleAttackForTest(t *testing.T) *DoubleAttackBlackjack {
	t.Helper()
	g := NewDefaultDoubleAttackBlackjack()
	g.Reset()
	return g
}

// --- デッキ ---

// **10 が抜けている 48 枚 × 8。**
func TestDoubleAttack_UsesTheSpanishShoe(t *testing.T) {
	g := newDoubleAttackForTest(t)
	assert.Equal(t, DoubleAttackDeckSize*DoubleAttackDeckCount, g.GetRemainingCards())

	// シューを全部引いて 10 が 1 枚も出ないことを見る。
	tens := 0
	for range DoubleAttackDeckSize * DoubleAttackDeckCount {
		c := g.shoe.DrawCard()
		if c != nil && c.GetValue() == 10 {
			tens++
		}
	}
	assert.Zero(t, tens, "スパニッシュデッキに 10 が混ざっている")
}

// --- Bust It の配当表 ---

// **実測した分布から決めた表を守る。**
//
// 200 万回のシミュレーションで得た頻度。配当表を変えればハウスエッジが動くので、
// この期待値そのものを錨にする。
func TestDoubleAttack_BustItHouseEdge(t *testing.T) {
	t.Parallel()

	prob := map[int]float64{
		3: 0.1530965, 4: 0.0940075, 5: 0.0245730,
		6: 0.0035905, 7: 0.0003280, 8: 0.0000240,
	}
	win := 0.0
	ret := 0.0
	for n, p := range prob {
		win += p
		ret += p * float64(DoubleAttackBustItPayout(n))
	}
	ev := ret - (1 - win)

	assert.InDelta(t, 0.275620, win, 0.000001, "バスト率が実測から動いた")
	assert.InDelta(t, -0.05212, ev, 0.0005, "Bust It の期待値が動いた (配当表を触った?)")
	assert.Negative(t, ev, "プレイヤー有利な配当表になっている")
}

// **issue の配当表を採ると成立しない。** 置くだけで勝てる賭けになる。
func TestDoubleAttack_TheIssuePaytableWouldBeBroken(t *testing.T) {
	t.Parallel()

	prob := map[int]float64{
		3: 0.1530965, 4: 0.0940075, 5: 0.0245730,
		6: 0.0035905, 7: 0.0003280, 8: 0.0000240,
	}
	issue := map[int]int{3: 15, 4: 20, 5: 40, 6: 80, 7: 120, 8: 200}

	win, ret := 0.0, 0.0
	for n, p := range prob {
		win += p
		ret += p * float64(issue[n])
	}
	assert.Positive(t, ret-(1-win), "前提が崩れている: issue の表はプレイヤー有利のはず")

	// 3 枚バストは 15% 起きるので、そこに 15:1 は払えない。
	assert.Less(t, DoubleAttackBustItPayout(3), issue[3],
		"3 枚バストの配当が issue のまま残っている")
}

// 8 枚以上はまとめて同じ配当。
func TestDoubleAttack_BustItPayoutSaturates(t *testing.T) {
	t.Parallel()

	assert.Equal(t, DoubleAttackBustItPayout(8), DoubleAttackBustItPayout(9))
	assert.Equal(t, DoubleAttackBustItPayout(8), DoubleAttackBustItPayout(20))
	assert.Greater(t, DoubleAttackBustItPayout(7), DoubleAttackBustItPayout(6),
		"枚数が増えるほど配当が上がる")
}

// --- 情報の順序 ---

// **これがこのゲームの本体。** アップカードだけを見せ、2 枚目は追加ベットの後。
func TestDoubleAttack_DealerHoleCardWaitsForTheAttack(t *testing.T) {
	g := newDoubleAttackForTest(t)
	require.NoError(t, g.PlaceBet(50, 0))

	assert.Equal(t, DoubleAttackPhaseAttack, g.GetPhase())
	assert.Len(t, g.GetDealerCards(), 1, "追加ベットの前にディーラーの 2 枚目が配られている")
	assert.False(t, g.IsDealerHoleDealt())

	require.NoError(t, g.Attack(0))
	assert.True(t, g.IsDealerHoleDealt())
	assert.GreaterOrEqual(t, len(g.GetDealerCards()), 2)
}

// **追加ベットはアンティまで。**
func TestDoubleAttack_AttackIsCappedAtTheAnte(t *testing.T) {
	g := newDoubleAttackForTest(t)
	require.NoError(t, g.PlaceBet(50, 0))

	assert.Equal(t, 50, g.MaxAttackBet())
	assert.ErrorIs(t, g.Attack(51), errDoubleAttackAttackRange)
	assert.ErrorIs(t, g.Attack(-1), errDoubleAttackAttackRange)

	before := g.GetChips()
	require.NoError(t, g.Attack(50))
	assert.Equal(t, 50, g.GetAttackBet())
	assert.Equal(t, before-50, g.GetChips())
}

// 追加ベットを見送っても進む。
func TestDoubleAttack_AttackMayBeDeclined(t *testing.T) {
	g := newDoubleAttackForTest(t)
	require.NoError(t, g.PlaceBet(50, 0))
	before := g.GetChips()

	require.NoError(t, g.Attack(0))
	assert.Zero(t, g.GetAttackBet())
	assert.Equal(t, before, g.GetChips(), "見送ったのにチップが減っている")
}

// --- 配当 ---

// daStaged は指定の手を積んで決着直前まで進めた卓を返す。
func daStaged(t *testing.T, ante, bustIt int, player, dealer []*Card) *DoubleAttackBlackjack {
	t.Helper()
	g := newDoubleAttackForTest(t)
	require.NoError(t, g.PlaceBet(ante, bustIt))

	h := NewBlackJackHand()
	for _, c := range player {
		h.AddCard(c)
	}
	h.SetBet(ante)
	g.hands = []*BlackJackHand{h}
	g.results = []DoubleAttackResult{DoubleAttackResultNone}

	d := NewBlackJackHand()
	for _, c := range dealer {
		d.AddCard(c)
	}
	g.dealerHand = d
	g.dealerHoleDealt = true
	g.phase = DoubleAttackPhasePlay
	return g
}

// **プレイヤーのブラックジャックは 1:1。** 3:2 ではない。
//
// ここを 3:2 に戻すと、アップカードを見てから賭け増しできる有利さに対価が無くなり
// 控除率が消える。
func TestDoubleAttack_BlackjackPaysEven(t *testing.T) {
	bj := []*Card{daCard(CardDesignSpade, 1), daCard(CardDesignHeart, 13)}
	dealer := []*Card{daCard(CardDesignClover, 9), daCard(CardDesignDiamond, 8)}
	g := daStaged(t, 100, 0, bj, dealer)
	before := g.GetChips()

	g.dealerPlay()

	require.Equal(t, DoubleAttackResultBlackjack, g.GetResults()[0])
	// 賭け 100 に対し 100 の返却 + 100 の配当 = 200。3:2 なら 250。
	assert.Equal(t, before+200, g.GetChips(), "ブラックジャックが 1:1 で払われていない")
}

func TestDoubleAttack_WinLosePush(t *testing.T) {
	tests := []struct {
		name           string
		player, dealer []*Card
		want           DoubleAttackResult
		wantReturn     int
	}{
		{
			name:   "高いほうが勝つ",
			player: []*Card{daCard(CardDesignSpade, 13), daCard(CardDesignHeart, 9)},
			dealer: []*Card{daCard(CardDesignClover, 13), daCard(CardDesignDiamond, 8)},
			want:   DoubleAttackResultWin, wantReturn: 200,
		},
		{
			name:   "低いほうが負ける",
			player: []*Card{daCard(CardDesignSpade, 13), daCard(CardDesignHeart, 8)},
			dealer: []*Card{daCard(CardDesignClover, 13), daCard(CardDesignDiamond, 9)},
			want:   DoubleAttackResultLose, wantReturn: 0,
		},
		{
			name:   "同点はプッシュ",
			player: []*Card{daCard(CardDesignSpade, 13), daCard(CardDesignHeart, 9)},
			dealer: []*Card{daCard(CardDesignClover, 12), daCard(CardDesignDiamond, 9)},
			want:   DoubleAttackResultPush, wantReturn: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := daStaged(t, 100, 0, tt.player, tt.dealer)
			before := g.GetChips()
			g.dealerPlay()
			assert.Equal(t, tt.want, g.GetResults()[0])
			assert.Equal(t, before+tt.wantReturn, g.GetChips())
		})
	}
}

// **プレイヤーがバストしたら、ディーラーがバストしても負け。**
func TestDoubleAttack_PlayerBustLosesEvenIfDealerBusts(t *testing.T) {
	g := daStaged(t, 100, 0,
		[]*Card{daCard(CardDesignSpade, 13), daCard(CardDesignHeart, 12), daCard(CardDesignClover, 5)},
		[]*Card{daCard(CardDesignDiamond, 13), daCard(CardDesignSpade, 12), daCard(CardDesignHeart, 5)})
	g.hands[0].SetBusted(true)
	before := g.GetChips()

	g.dealerPlay()
	assert.Equal(t, DoubleAttackResultLose, g.GetResults()[0])
	assert.Equal(t, before, g.GetChips(), "バストしたのに払い戻された")
}

// **全部バストしてもディーラーは引き切る。** Bust It が生きているため。
func TestDoubleAttack_DealerPlaysOutForTheBustItBet(t *testing.T) {
	g := daStaged(t, 100, 50,
		[]*Card{daCard(CardDesignSpade, 13), daCard(CardDesignHeart, 12), daCard(CardDesignClover, 5)},
		[]*Card{daCard(CardDesignDiamond, 6), daCard(CardDesignSpade, 6)})
	g.hands[0].SetBusted(true)
	before := g.GetChips()

	g.dealerPlay()

	assert.Greater(t, len(g.GetDealerCards()), 2, "ディーラーが引いていない")
	if g.GetDealerScore() > 21 {
		assert.Positive(t, g.GetBustItPayout(), "ディーラーがバストしたのに Bust It が払われていない")
		assert.Greater(t, g.GetChips(), before)
	} else {
		assert.Zero(t, g.GetBustItPayout())
	}
}

// Bust It はディーラーがバストしなければ没収。
func TestDoubleAttack_BustItLosesWhenDealerStands(t *testing.T) {
	g := daStaged(t, 100, 50,
		[]*Card{daCard(CardDesignSpade, 13), daCard(CardDesignHeart, 9)},
		[]*Card{daCard(CardDesignClover, 13), daCard(CardDesignDiamond, 9)})
	g.dealerPlay()
	require.LessOrEqual(t, g.GetDealerScore(), 21)
	assert.Zero(t, g.GetBustItPayout())
}

// **ディーラーはソフト 17 でヒットする。**
func TestDoubleAttack_DealerHitsSoft17(t *testing.T) {
	g := daStaged(t, 100, 0,
		[]*Card{daCard(CardDesignSpade, 13), daCard(CardDesignHeart, 9)},
		[]*Card{daCard(CardDesignClover, 1), daCard(CardDesignDiamond, 6)}) // A+6 = soft 17
	g.dealerPlay()
	assert.Greater(t, len(g.GetDealerCards()), 2, "ソフト 17 で止まっている")
}

// ハード 17 では止まる。
func TestDoubleAttack_DealerStandsOnHard17(t *testing.T) {
	g := daStaged(t, 100, 0,
		[]*Card{daCard(CardDesignSpade, 13), daCard(CardDesignHeart, 9)},
		[]*Card{daCard(CardDesignClover, 13), daCard(CardDesignDiamond, 7)}) // 10+7 = hard 17
	g.dealerPlay()
	assert.Len(t, g.GetDealerCards(), 2, "ハード 17 で引いている")
}

// --- スプリットとダブル ---

func TestDoubleAttack_SplitCreatesTwoHands(t *testing.T) {
	g := daStaged(t, 100, 0,
		[]*Card{daCard(CardDesignSpade, 8), daCard(CardDesignHeart, 8)},
		[]*Card{daCard(CardDesignClover, 13), daCard(CardDesignDiamond, 6)})
	require.True(t, g.CanSplit())
	before := g.GetChips()

	require.NoError(t, g.Split())
	assert.Equal(t, 2, g.GetHandCount())
	assert.Equal(t, before-100, g.GetChips(), "2 つ目の手札に賭け金が出ていない")
	for _, h := range g.GetHands() {
		assert.Equal(t, 2, h.GetCardsSize(), "分けた手札に 1 枚ずつ配られていない")
	}
}

// **エースを割ったら 1 枚ずつで打ち止め。**
func TestDoubleAttack_SplitAcesGetOneCardEach(t *testing.T) {
	g := daStaged(t, 100, 0,
		[]*Card{daCard(CardDesignSpade, 1), daCard(CardDesignHeart, 1)},
		[]*Card{daCard(CardDesignClover, 13), daCard(CardDesignDiamond, 6)})
	require.NoError(t, g.Split())

	for i, h := range g.GetHands() {
		assert.True(t, h.IsStood(), "手札 %d が打ち止めになっていない", i)
		assert.Equal(t, 2, h.GetCardsSize())
	}
	assert.NotEqual(t, DoubleAttackPhasePlay, g.GetPhase(), "エース分割後も操作を求めている")
}

func TestDoubleAttack_SplitRejectedWhenNotAPair(t *testing.T) {
	g := daStaged(t, 100, 0,
		[]*Card{daCard(CardDesignSpade, 8), daCard(CardDesignHeart, 9)},
		[]*Card{daCard(CardDesignClover, 13), daCard(CardDesignDiamond, 6)})
	assert.False(t, g.CanSplit())
	assert.ErrorIs(t, g.Split(), errDoubleAttackCannotSplit)
}

func TestDoubleAttack_DoubleTakesOneCardAndEndsTheHand(t *testing.T) {
	g := daStaged(t, 100, 0,
		[]*Card{daCard(CardDesignSpade, 6), daCard(CardDesignHeart, 5)},
		[]*Card{daCard(CardDesignClover, 13), daCard(CardDesignDiamond, 6)})
	require.True(t, g.CanDouble())

	require.NoError(t, g.Double())
	// **チップは見ない。** ダブルは手札を終わらせるので、その場でディーラーが引き、
	// 精算まで走る。ここで残高を比べると「倍にした」ではなく「勝ったか」を測ってしまう。
	assert.Equal(t, 200, g.GetHands()[0].GetBet(), "賭け金が倍になっていない")
	assert.Equal(t, 3, g.GetHands()[0].GetCardsSize(), "ダブルで 1 枚だけ引いていない")
	assert.NotEqual(t, DoubleAttackPhasePlay, g.GetPhase(), "ダブルの後も操作を求めている")
}

func TestDoubleAttack_DoubleRejectedAfterHitting(t *testing.T) {
	g := daStaged(t, 100, 0,
		[]*Card{daCard(CardDesignSpade, 4), daCard(CardDesignHeart, 3)},
		[]*Card{daCard(CardDesignClover, 13), daCard(CardDesignDiamond, 6)})
	require.NoError(t, g.Hit())
	if g.GetPhase() == DoubleAttackPhasePlay {
		assert.False(t, g.CanDouble(), "3 枚目を引いた後にダブルできてしまう")
		assert.ErrorIs(t, g.Double(), errDoubleAttackCannotDouble)
	}
}

// --- 進行 ---

func TestDoubleAttack_PlaceBetValidation(t *testing.T) {
	for _, tt := range []struct {
		name    string
		ante    int
		bustIt  int
		wantErr error
	}{
		{"最低額は通る", DoubleAttackAnteMin, 0, nil},
		{"上限額は通る", DoubleAttackAnteMax, 0, nil},
		{"低すぎる", DoubleAttackAnteMin - 1, 0, errDoubleAttackAnteRange},
		{"高すぎる", DoubleAttackAnteMax + 1, 0, errDoubleAttackAnteRange},
		{"負のサイドベット", 50, -1, errDoubleAttackSideRange},
		{"サイドベットが上限超え", 50, DoubleAttackBustItMax + 1, errDoubleAttackSideRange},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := newDoubleAttackForTest(t)
			g.SetChips(DoubleAttackChipsMax)
			err := g.PlaceBet(tt.ante, tt.bustIt)
			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestDoubleAttack_PhaseGuards(t *testing.T) {
	g := newDoubleAttackForTest(t)
	assert.ErrorIs(t, g.Attack(10), errDoubleAttackWrongPhase)
	assert.ErrorIs(t, g.Hit(), errDoubleAttackWrongPhase)
	assert.ErrorIs(t, g.Stand(), errDoubleAttackWrongPhase)
	assert.ErrorIs(t, g.NextRound(), errDoubleAttackWrongPhase)

	require.NoError(t, g.PlaceBet(50, 0))
	assert.ErrorIs(t, g.PlaceBet(50, 0), errDoubleAttackWrongPhase)
	assert.ErrorIs(t, g.Hit(), errDoubleAttackWrongPhase, "追加ベットの前に引けてしまう")
}

// **資金が尽きたら終わる。**
func TestDoubleAttack_EndsWhenOutOfChips(t *testing.T) {
	g := newDoubleAttackForTest(t)
	require.NoError(t, g.PlaceBet(DoubleAttackAnteMin, 0))
	require.NoError(t, g.Attack(0))
	for g.GetPhase() == DoubleAttackPhasePlay {
		require.NoError(t, g.Stand())
	}
	g.SetChips(0)
	require.NoError(t, g.NextRound())
	assert.True(t, g.GetGameEndFlag())
	assert.ErrorIs(t, g.PlaceBet(DoubleAttackAnteMin, 0), errDoubleAttackFinished)
}

// **CPU 相手ではなく自分だけの卓なので、通しで回して破綻しないことを見る。**
func TestDoubleAttack_FullSessionsStayConsistent(t *testing.T) {
	for seed := range 30 {
		g := newDoubleAttackForTest(t)
		g.SetChips(5000)
		for step := 0; step < 300 && !g.GetGameEndFlag(); step++ {
			switch g.GetPhase() {
			case DoubleAttackPhaseBet:
				require.NoError(t, g.PlaceBet(DoubleAttackAnteMin, 10))
			case DoubleAttackPhaseAttack:
				require.NoError(t, g.Attack(0))
			case DoubleAttackPhasePlay:
				require.NoError(t, g.Stand())
			default:
				require.NoError(t, g.NextRound())
			}
			require.GreaterOrEqual(t, g.GetChips(), 0, "seed %d: チップが負になった", seed)
		}
	}
}

// --- ヒント ---

func TestDoubleAttack_GetHint(t *testing.T) {
	t.Run("賭ける前は助言しない", func(t *testing.T) {
		assert.Nil(t, newDoubleAttackForTest(t).GetHint())
	})

	t.Run("弱いアップカードなら乗せる", func(t *testing.T) {
		g := newDoubleAttackForTest(t)
		require.NoError(t, g.PlaceBet(50, 0))
		g.dealerHand = NewBlackJackHand()
		g.dealerHand.AddCard(daCard(CardDesignSpade, 5))
		h := g.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "attack", h.Action)
		assert.Equal(t, "weakUpCard", h.Reason)
	})

	t.Run("強いアップカードなら乗せない", func(t *testing.T) {
		g := newDoubleAttackForTest(t)
		require.NoError(t, g.PlaceBet(50, 0))
		g.dealerHand = NewBlackJackHand()
		g.dealerHand.AddCard(daCard(CardDesignSpade, 1))
		h := g.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "stand", h.Action)
		assert.Equal(t, "strongUpCard", h.Reason)
	})

	t.Run("11 以下は引く", func(t *testing.T) {
		g := daStaged(t, 100, 0,
			[]*Card{daCard(CardDesignSpade, 4), daCard(CardDesignHeart, 5)},
			[]*Card{daCard(CardDesignClover, 13), daCard(CardDesignDiamond, 6)})
		h := g.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "hit", h.Action)
	})
}

// --- 名前と設定 ---

func TestDoubleAttackNames(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "bet", DoubleAttackPhaseName(DoubleAttackPhaseBet))
	assert.Equal(t, "attack", DoubleAttackPhaseName(DoubleAttackPhaseAttack))
	assert.Equal(t, "play", DoubleAttackPhaseName(DoubleAttackPhasePlay))
	assert.Equal(t, "result", DoubleAttackPhaseName(DoubleAttackPhaseResult))

	assert.Equal(t, "win", DoubleAttackResultName(DoubleAttackResultWin))
	assert.Equal(t, "lose", DoubleAttackResultName(DoubleAttackResultLose))
	assert.Equal(t, "push", DoubleAttackResultName(DoubleAttackResultPush))
	assert.Equal(t, "blackjack", DoubleAttackResultName(DoubleAttackResultBlackjack))
	assert.Equal(t, "none", DoubleAttackResultName(DoubleAttackResultNone))
}

func TestDoubleAttackConfig_Validate(t *testing.T) {
	t.Parallel()

	assert.NoError(t, DefaultDoubleAttackBlackjackConfig().Validate())
	for _, cfg := range []DoubleAttackBlackjackConfig{
		{InitialChips: DoubleAttackChipsMin - 1, DefaultAnte: 50},
		{InitialChips: DoubleAttackChipsMax + 1, DefaultAnte: 50},
		{InitialChips: 1000, DefaultAnte: DoubleAttackAnteMin - 1},
		{InitialChips: 1000, DefaultAnte: DoubleAttackAnteMax + 1},
	} {
		assert.Error(t, cfg.Validate())
	}
}

func TestDoubleAttack_Accessors(t *testing.T) {
	g := newDoubleAttackForTest(t)
	assert.NotNil(t, g.GetPlayer())
	assert.Equal(t, DoubleAttackDefaultChips, g.GetConfig().InitialChips)
	assert.NotEmpty(t, g.GetActionLog())
	assert.Zero(t, g.GetHandCount())
	assert.Nil(t, g.GetDealerCards())
	assert.Zero(t, g.GetDealerScore())

	require.NoError(t, g.PlaceBet(50, 20))
	assert.Equal(t, 50, g.GetAnteBet())
	assert.Equal(t, 20, g.GetBustItBet())
	assert.Equal(t, 1, g.GetHandCount())
	assert.Zero(t, g.GetActiveHandIdx())

	g.SetConfig(DoubleAttackBlackjackConfig{InitialChips: 500, DefaultAnte: 20})
	assert.Equal(t, 500, g.GetConfig().InitialChips)
}

func TestDoubleAttack_ActionLogIsBounded(t *testing.T) {
	g := newDoubleAttackForTest(t)
	for range doubleAttackMaxSliceLen + 50 {
		g.appendLog("noise", "x", nil)
	}
	assert.Len(t, g.GetActionLog(), doubleAttackMaxSliceLen)
}
