//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newBanLuckForTest(t *testing.T) *BanLuck {
	t.Helper()
	g := NewDefaultBanLuck()
	g.Reset()
	return g
}

// blStackNext は次に引かれる札を指定する。
//
// **引く札を配りに委ねると検査が配り依存になる。** 役の判定も親の義務ヒットも
// 引いた札で分岐するので、固定しないと同じ assert が配りで通ったり落ちたりする。
func blStackNext(g *BanLuck, cards ...*Card) {
	for i, c := range cards {
		g.deck.deck[g.deck.deckDrawCnt+i] = c
	}
}

// blDealPlain は特別役の出ない配りを積む (全席 9+8 = 17)。
//
// **配りに委ねると `PlaceBet` がラウンドの最後まで走ることがある。** 特別役の
// 席は手番を持たないので、全席が確定するとその場で精算し、**親まで次へ移る**。
// その後で手札を差し替えても、親も局面も前のラウンドのままになっていて、
// 「子は 15 未満でも止まれる」のような検査が親の義務で落ちる (実測)。
func blDealPlain(g *BanLuck) {
	cards := make([]*Card, 0, 2*len(g.players))
	for range g.players {
		cards = append(cards, blCard(CardDesignSpade, 9), blCard(CardDesignHeart, 8))
	}
	blStackNext(g, cards...)
}

// blTotalChips は卓のチップ総量を返す。
func blTotalChips(g *BanLuck) int {
	total := 0
	for _, p := range g.GetPlayers() {
		total += p.GetChips()
	}
	return total
}

// blStep は 1 手だけ進める。人間の席は「義務があれば引く、無ければ止まる」。
//
// **全席を CPU にしてはいけない。** `humanSeat()` は人間が居なければ席 0 に
// 落ちるので、`isHuman` を全部 false にすると席 0 が人間扱いのまま誰も動かさず、
// テストが無限に回る (実測: 10 分でタイムアウト)。人間の席は残したまま、
// その席を機械的に打たせるのが正しい回し方。
func blStep(t *testing.T, g *BanLuck) {
	t.Helper()
	switch g.GetPhase() {
	case BanLuckPhaseBet:
		bet := g.GetConfig().DefaultBet
		if chips := g.GetPlayers()[g.GetHumanSeat()].GetChips(); chips < bet {
			bet = BanLuckMinBet
		}
		require.NoError(t, g.PlaceBet(bet))
	case BanLuckPhasePlay:
		if !g.IsHumanTurn() {
			before := g.GetTurnSeat()
			g.CpuPlay()
			require.False(t, g.GetPhase() == BanLuckPhasePlay && !g.IsHumanTurn() && g.GetTurnSeat() == before,
				"CpuPlay が席 %d を進めていない", before)
			return
		}
		if g.MustHit() {
			require.NoError(t, g.Hit())
			return
		}
		require.NoError(t, g.Stand())
	case BanLuckPhaseRoundEnd:
		require.NoError(t, g.NextRound())
	default:
	}
}

// blPlayOut は終局まで回す。
func blPlayOut(t *testing.T, g *BanLuck, onRoundEnd func()) {
	t.Helper()
	for steps := 0; !g.GetGameEndFlag(); steps++ {
		require.Less(t, steps, 3000, "局が終わらない")
		if g.GetPhase() == BanLuckPhaseRoundEnd && onRoundEnd != nil {
			onRoundEnd()
		}
		blStep(t, g)
	}
}

// --- 停止性 ---

// **規則が止まることを最初に確かめる。** 引き取り系や義務ヒットのあるゲームは
// 素直に書くと終わらない経路が生まれる。CPU だけで何百局も回して数える。
func TestBanLuck_AlwaysTerminates(t *testing.T) {
	t.Parallel()
	rounds := 0
	for range 200 {
		g := newBanLuckForTest(t)
		blPlayOut(t, g, func() { rounds++ })
		// **終わったことだけでは足りない。** 1 ラウンドも回さずに終局しても
		// 「止まった」ようには見えるので、実際に配られたことまで数える。
		require.Positive(t, g.GetRoundNumber())
	}
	assert.GreaterOrEqual(t, rounds, 200*BanLuckDefaultRounds,
		"200 局で %d ラウンドしか回っていない — 局が早々に終わっている", rounds)
}

// --- チップの保存 ---

// **総量が変わらないだけでは足りない。** 親と子の間でしか動かないので、
// どこかに取り残されていないことも見る。
func TestBanLuck_ChipsAreConserved(t *testing.T) {
	t.Parallel()
	for range 100 {
		g := newBanLuckForTest(t)
		want := blTotalChips(g)
		blPlayOut(t, g, func() {
			assert.Equal(t, want, blTotalChips(g), "ラウンド終了時に総量が変わっている")
		})
		assert.Equal(t, want, blTotalChips(g), "終局時に総量が変わっている")
		for i, p := range g.GetPlayers() {
			assert.GreaterOrEqual(t, p.GetChips(), 0, "席 %d の残高が負", i)
		}
	}
}

// **親が持っていない額は払えない。** 集めてから払う順序にしていないと、
// 3 席が同時に勝ったときに親の残高が負になる。
func TestBanLuck_BankerCannotPayMoreThanItHolds(t *testing.T) {
	t.Parallel()
	g := newBanLuckForTest(t)
	g.players[g.banker].SetChips(30) // 子 3 席 × 賭け 50 より少ない
	blDealPlain(g)
	require.NoError(t, g.PlaceBet(50))
	// 親をバストさせ、子は全員生かす。
	for i := range g.players {
		h := NewBlackJackHand()
		if i == g.banker {
			h.AddCard(blCard(CardDesignSpade, 10))
			h.AddCard(blCard(CardDesignHeart, 9))
			h.AddCard(blCard(CardDesignClover, 8)) // 27
			h.SetBusted(true)
		} else {
			h.AddCard(blCard(CardDesignSpade, 10))
			h.AddCard(blCard(CardDesignHeart, 9)) // 19
			h.SetStood(true)
		}
		g.hands[i] = h
	}
	before := blTotalChips(g)
	g.settled = false
	g.settle()

	assert.GreaterOrEqual(t, g.players[g.banker].GetChips(), 0, "親の残高が負になっている")
	assert.Equal(t, before, blTotalChips(g), "払えない額を配って総量が増えている")
}

// --- 親の義務ヒット ---

// **人間が親でも止まれない。** CPU 側だけで守ると、人間が親のときだけ規則が消える。
func TestBanLuck_HumanBankerCannotStandBelowMinimum(t *testing.T) {
	t.Parallel()
	g := newBanLuckForTest(t)
	require.Equal(t, 0, g.GetBankerSeat(), "既定では席 0 が親")
	blDealPlain(g)
	require.NoError(t, g.PlaceBet(50))

	// 子を全員止め、親 (人間) に 14 を持たせて手番を回す。
	for i := range g.players {
		h := NewBlackJackHand()
		if i == g.banker {
			h.AddCard(blCard(CardDesignSpade, 10))
			h.AddCard(blCard(CardDesignHeart, 4)) // 14 = 15 未満
		} else {
			h.AddCard(blCard(CardDesignSpade, 10))
			h.AddCard(blCard(CardDesignHeart, 9))
			h.SetStood(true)
		}
		g.hands[i] = h
	}
	g.turn = g.banker
	g.phase = BanLuckPhasePlay

	assert.True(t, g.MustHit(), "義務が画面に伝わっていない")
	assert.ErrorIs(t, g.Stand(), errBanLuckMustHit)

	// 15 以上にすれば止まれる。
	g.hands[g.banker].AddCard(blCard(CardDesignClover, 3)) // 17
	assert.False(t, g.MustHit())
	assert.NoError(t, g.Stand())
}

// **子は 15 未満でも止まれる。** 義務は親だけのもの。
func TestBanLuck_SeatMayStandBelowMinimum(t *testing.T) {
	t.Parallel()
	g := newBanLuckForTest(t)
	g.banker = 1 // 人間 (席 0) を子にする
	blDealPlain(g)
	require.NoError(t, g.PlaceBet(50))

	g.hands[0] = blHand(blCard(CardDesignSpade, 5), blCard(CardDesignHeart, 6)) // 11
	g.turn = 0
	g.phase = BanLuckPhasePlay

	assert.False(t, g.MustHit(), "子に義務が出ている")
	assert.NoError(t, g.Stand())
}

// --- 特別役の扱い ---

// **配られた時点で確定した役に手番は要らない。** 引かせると役が消える。
func TestBanLuck_SpecialHandsGetNoTurn(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		cards []*Card
	}{
		{"Ban Ban", []*Card{blCard(CardDesignSpade, 1), blCard(CardDesignHeart, 1)}},
		{"Ban Luck", []*Card{blCard(CardDesignSpade, 1), blCard(CardDesignHeart, 13)}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := newBanLuckForTest(t)
			g.banker = 1
			blDealPlain(g)
			require.NoError(t, g.PlaceBet(50))
			g.hands[0] = blHand(tt.cards...)
			assert.True(t, g.seatDone(0), "特別役なのに手番が残っている")
		})
	}
}

// --- 親の交代 ---

// **特別役で親を破った席が次の親。** 普通の勝ちでは奪えない。
func TestBanLuck_BankerRotation(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name       string
		seat2Cards []*Card
		wantBanker int
	}{
		{
			name:       "Ban Luck で破れば奪う",
			seat2Cards: []*Card{blCard(CardDesignSpade, 1), blCard(CardDesignHeart, 13)},
			wantBanker: 2,
		},
		{
			// **普通の勝ちでは奪えない。** 既定の「次の席」へ回る。
			name:       "普通の勝ちでは奪えない",
			seat2Cards: []*Card{blCard(CardDesignSpade, 10), blCard(CardDesignHeart, 9)},
			wantBanker: 1,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := newBanLuckForTest(t)
			require.Equal(t, 0, g.GetBankerSeat())
			blDealPlain(g)
			require.NoError(t, g.PlaceBet(50))

			for i := range g.players {
				var h *BlackJackHand
				switch i {
				case g.banker:
					h = blHand(blCard(CardDesignSpade, 10), blCard(CardDesignClover, 6)) // 16
				case 2:
					h = blHand(tt.seat2Cards...)
				default:
					h = blHand(blCard(CardDesignSpade, 10), blCard(CardDesignDiamond, 2)) // 12 で負け
				}
				h.SetStood(true)
				g.hands[i] = h
			}
			g.settled = false
			g.settle()
			assert.Equal(t, tt.wantBanker, g.GetBankerSeat())
		})
	}
}

// **親は必ず誰かになる。** 「誰も親になれない」経路を作っていないこと。
func TestBanLuck_BankerAlwaysExists(t *testing.T) {
	t.Parallel()
	g := newBanLuckForTest(t)
	blPlayOut(t, g, func() {
		assert.GreaterOrEqual(t, g.GetBankerSeat(), 0)
		assert.Less(t, g.GetBankerSeat(), len(g.GetPlayers()))
	})
}

// --- 賭け金の検証 ---

func TestBanLuck_PlaceBetValidation(t *testing.T) {
	t.Parallel()
	g := newBanLuckForTest(t)
	g.banker = 1 // 人間を子にする

	assert.ErrorIs(t, g.PlaceBet(BanLuckMinBet-BanLuckBetUnit), errBanLuckBetRangeIn)
	assert.ErrorIs(t, g.PlaceBet(BanLuckMaxBet+BanLuckBetUnit), errBanLuckBetRangeIn)
	assert.ErrorIs(t, g.PlaceBet(55), errBanLuckBetUnitIn)

	g.players[0].SetChips(20)
	assert.ErrorIs(t, g.PlaceBet(50), errBanLuckNotEnough)

	g.players[0].SetChips(1000)
	blDealPlain(g)
	assert.NoError(t, g.PlaceBet(50))
	assert.Equal(t, BanLuckPhasePlay, g.GetPhase())
}

// **人間が親のラウンドは額を取らない。** 親は賭ける側ではない。
func TestBanLuck_BankerPlacesNoBet(t *testing.T) {
	t.Parallel()
	g := newBanLuckForTest(t)
	require.Equal(t, g.GetHumanSeat(), g.GetBankerSeat())
	before := g.players[0].GetChips()
	// **配りを固定する。** 特別役が出るとその場で精算まで走り、親のチップが
	// 配当で動くので「賭け金を取っていない」が読めなくなる。全席 17。
	cards := make([]*Card, 0, 2*len(g.players))
	for range g.players {
		cards = append(cards, blCard(CardDesignSpade, 9), blCard(CardDesignHeart, 8))
	}
	blStackNext(g, cards...)

	// 範囲外の額でも親なら通る (無視されるため)。
	require.NoError(t, g.PlaceBet(0))
	require.Equal(t, BanLuckPhasePlay, g.GetPhase(), "積んだ配りで決着してしまった")
	assert.Equal(t, before, g.players[0].GetChips(), "親から賭け金を取っている")
	assert.Zero(t, g.players[0].GetBet())
}

func TestBanLuck_PhaseGuards(t *testing.T) {
	t.Parallel()
	g := newBanLuckForTest(t)
	assert.ErrorIs(t, g.Hit(), errBanLuckWrongPhase)
	assert.ErrorIs(t, g.Stand(), errBanLuckWrongPhase)
	assert.ErrorIs(t, g.NextRound(), errBanLuckWrongPhase)

	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlaceBet(50), errBanLuckFinished)
	assert.ErrorIs(t, g.Hit(), errBanLuckFinished)
	assert.ErrorIs(t, g.NextRound(), errBanLuckFinished)
}

// **他人の手番では動かせない。**
func TestBanLuck_NotYourTurn(t *testing.T) {
	t.Parallel()
	g := newBanLuckForTest(t)
	g.banker = 1
	blDealPlain(g)
	require.NoError(t, g.PlaceBet(50))
	g.phase = BanLuckPhasePlay
	g.turn = 2 // CPU の席
	assert.ErrorIs(t, g.Hit(), errBanLuckNotYourRun)
	assert.ErrorIs(t, g.Stand(), errBanLuckNotYourRun)
}

// **5 枚を超えて引けない。** Five Dragon が上限。
func TestBanLuck_HandStopsAtFiveCards(t *testing.T) {
	t.Parallel()
	g := newBanLuckForTest(t)
	g.banker = 1
	blDealPlain(g)
	require.NoError(t, g.PlaceBet(50))
	g.hands[0] = blHand(blCard(CardDesignSpade, 2), blCard(CardDesignHeart, 2),
		blCard(CardDesignClover, 2), blCard(CardDesignDiamond, 2), blCard(CardDesignSpade, 3))
	g.turn = 0
	g.phase = BanLuckPhasePlay
	assert.ErrorIs(t, g.hitSeat(0), errBanLuckHandFull)
	assert.True(t, g.seatDone(0))
}

func TestBanLuck_Accessors(t *testing.T) {
	t.Parallel()
	g := newBanLuckForTest(t)
	assert.Equal(t, BanLuckPhaseBet, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Positive(t, g.GetRemainingCards())
	assert.NotEmpty(t, g.GetActionLog())
	assert.Len(t, g.GetPlayers(), BanLuckDefaultSeats)
	assert.Zero(t, g.GetHumanSeat())
	assert.Zero(t, g.GetTurnSeat())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, BanLuckDefaultChips, g.GetConfig().InitialChips)

	g.SetConfig(BanLuckConfig{Seats: 3, InitialChips: 500, Rounds: 5, DefaultBet: 20})
	assert.Equal(t, 500, g.GetConfig().InitialChips)
}

func TestBanLuck_WinnerSeat(t *testing.T) {
	t.Parallel()
	g := newBanLuckForTest(t)
	g.players[2].SetChips(9999)
	assert.Equal(t, 2, g.WinnerSeat())

	// 同点なら若い席。
	for _, p := range g.players {
		p.SetChips(100)
	}
	assert.Zero(t, g.WinnerSeat())
}

// **チップが尽きた席が 2 未満になったら終わる。** 勝負が成立しないため。
func TestBanLuck_EndsWhenTooFewSeatsHaveChips(t *testing.T) {
	t.Parallel()
	g := newBanLuckForTest(t)
	blDealPlain(g)
	require.NoError(t, g.PlaceBet(0))
	for steps := 0; g.GetPhase() == BanLuckPhasePlay; steps++ {
		require.Less(t, steps, 200, "ラウンドが終わらない")
		blStep(t, g)
	}
	require.Equal(t, BanLuckPhaseRoundEnd, g.GetPhase())

	for i, p := range g.players {
		if i > 0 {
			p.SetChips(0)
		}
	}
	require.NoError(t, g.NextRound())
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, BanLuckPhaseGameEnd, g.GetPhase())
}

func TestBanLuck_ConfigValidate(t *testing.T) {
	t.Parallel()
	assert.NoError(t, DefaultBanLuckConfig().Validate())
	for _, tt := range []struct {
		name string
		cfg  BanLuckConfig
		want error
	}{
		{"席が少ない", BanLuckConfig{Seats: 1, InitialChips: 1000, Rounds: 10, DefaultBet: 50}, errBanLuckSeatsRange},
		{"席が多い", BanLuckConfig{Seats: 7, InitialChips: 1000, Rounds: 10, DefaultBet: 50}, errBanLuckSeatsRange},
		{"チップ範囲外", BanLuckConfig{Seats: 4, InitialChips: 1, Rounds: 10, DefaultBet: 50}, errBanLuckChipsRange},
		{"ラウンド範囲外", BanLuckConfig{Seats: 4, InitialChips: 1000, Rounds: 1, DefaultBet: 50}, errBanLuckRoundsRange},
		{"賭け金範囲外", BanLuckConfig{Seats: 4, InitialChips: 1000, Rounds: 10, DefaultBet: 5}, errBanLuckBetRange},
		{"賭け金の刻み外れ", BanLuckConfig{Seats: 4, InitialChips: 1000, Rounds: 10, DefaultBet: 55}, errBanLuckBetUnit},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.ErrorIs(t, tt.cfg.Validate(), tt.want)
		})
	}
}
