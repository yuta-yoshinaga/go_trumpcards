//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newKingoForTest は配り始めた卓を返す。
func newKingoForTest(t *testing.T) *Kingo {
	t.Helper()
	g := NewDefaultKingo()
	g.Reset()
	return g
}

// kingoTotalChips は卓のチップ総量を返す。
func kingoTotalChips(g *Kingo) int {
	total := 0
	for _, p := range g.GetPlayers() {
		total += p.GetChips()
	}
	return total
}

// kingoCards は数字の並びから手札を作る。
func kingoCards(values ...int) []*Card {
	out := make([]*Card, 0, len(values))
	for i, v := range values {
		out = append(out, NewCard(1+(i%4), v, false))
	}
	return out
}

// kingoPlayRound は人間の張りを 1 回入れてラウンドを閉じる。
func kingoPlayRound(t *testing.T, g *Kingo) {
	t.Helper()
	if g.GetGameEndFlag() {
		return
	}
	if g.IsHumanBanker() {
		// 親の回は配る合図を出す。
		require.NoError(t, g.Deal())
		return
	}
	require.NoError(t, g.PlaceBet(g.GetConfig().MinBet))
}

// --- 役 ---

// **競うのはそろい方であって、合計ではない。** おいちょかぶの下一桁を
// 持ち込むと別のゲームになる。
func TestKingoHandRank(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		cards  []*Card
		want   KingoRank
		wanted int
	}{
		{"嵐", kingoCards(7, 7, 7), KingoRankArashi, 7},
		{"2 枚そろい", kingoCards(3, 3, 9), KingoRankPair, 3},
		{"2 枚そろい (並び違い)", kingoCards(9, 3, 3), KingoRankPair, 3},
		{"役なし", kingoCards(1, 5, 9), KingoRankNone, 9},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, KingoHandRank(tc.cards))
			assert.Equal(t, tc.wanted, KingoMatchedValue(tc.cards))
		})
	}
}

// **合計が同じでも役が違えば勝敗が決まる。** 合計で競っていないことの証拠。
func TestKingo_TotalDoesNotDecideTheHand(t *testing.T) {
	t.Parallel()
	// どちらも合計 15。片方は 2 枚そろい、もう片方は役なし。
	pair := kingoCards(5, 5, 5)
	none := kingoCards(2, 4, 9)
	assert.Equal(t, 15, kingoSum(pair))
	assert.Equal(t, 15, kingoSum(none))
	assert.Positive(t, KingoCompare(pair, none), "そろえたほうが負けている")
}

func kingoSum(cards []*Card) int {
	n := 0
	for _, c := range cards {
		n += c.GetValue()
	}
	return n
}

// **同じ役どうしは、そろえた数字の大きいほうが勝つ。** ここが無いと、
// 同じ役の対戦がすべて引き分けになる。
func TestKingoCompare(t *testing.T) {
	t.Parallel()
	assert.Positive(t, KingoCompare(kingoCards(9, 9, 1), kingoCards(3, 3, 10)),
		"大きい数字でそろえたほうが負けている")
	assert.Negative(t, KingoCompare(kingoCards(3, 3, 10), kingoCards(9, 9, 1)))
	assert.Positive(t, KingoCompare(kingoCards(2, 2, 2), kingoCards(10, 10, 1)),
		"嵐が 2 枚そろいに負けている")

	// 役もそろえた数字も同じなら、残り札で決まる。
	assert.Positive(t, KingoCompare(kingoCards(4, 4, 10), kingoCards(4, 4, 1)))
	assert.Zero(t, KingoCompare(kingoCards(4, 4, 7), kingoCards(4, 4, 7)))
}

// **配当は役の出にくさに見合っている。** 同じ倍率だと「そろえに行く意味」が
// 消える。
func TestKingoPayout(t *testing.T) {
	t.Parallel()
	assert.Greater(t, KingoPayout(KingoRankArashi), KingoPayout(KingoRankPair),
		"嵐が 2 枚そろいと同じ配当になっている")
	assert.Positive(t, KingoPayout(KingoRankNone))
}

// --- 配札 ---

func TestKingo_DealsThreeCardsToEverySeat(t *testing.T) {
	g := newKingoForTest(t)
	kingoPlayRound(t, g)
	for i, p := range g.GetPlayers() {
		assert.Len(t, p.GetCards(), KingoHandSize, "席 %d の枚数", i)
	}
	// 株札 40 枚から 席数 × 3 枚を配っている。
	assert.Equal(t, KingoDeckSize-len(g.GetPlayers())*KingoHandSize, g.GetRemainingCards())
}

// **山は株札 40 枚。** 1〜10 が 4 枚ずつで、52 枚デッキではない。
func TestKingo_UsesTheKabufudaDeck(t *testing.T) {
	deck := buildOichoKabuDeck()
	require.Len(t, deck, KingoDeckSize)
	counts := map[int]int{}
	for _, c := range deck {
		counts[c.GetValue()]++
	}
	assert.Len(t, counts, KingoValueMax, "数字の種類が 10 でない")
	for v, n := range counts {
		assert.Equal(t, 4, n, "数字 %d の枚数", v)
	}
}

// --- 親と精算 ---

// **親は張らない。** 受ける側なので、張らせると二重に払うことになる。
func TestKingo_BankerDoesNotBet(t *testing.T) {
	g := newKingoForTest(t)
	for range 10 {
		if g.IsHumanBanker() {
			assert.ErrorIs(t, g.PlaceBet(g.GetConfig().MinBet), errKingoBankerBets)
			assert.Zero(t, g.GetPlayers()[g.HumanSeat()].GetBet(),
				"親なのに張りが立っている")
			// 親の回に進める手は Deal のほう。
			require.NoError(t, g.Deal())
			return
		}
		kingoPlayRound(t, g)
		if g.GetGameEndFlag() {
			break
		}
		require.NoError(t, g.NextRound())
	}
	t.Fatal("10 ラウンド回しても人間が親にならなかった")
}

// **親は席を順に回る。** 総取りの側が固定されない。
func TestKingo_BankerRotates(t *testing.T) {
	g := newKingoForTest(t)
	seen := map[int]bool{}
	for range g.GetConfig().Rounds {
		seen[g.GetBankerSeat()] = true
		kingoPlayRound(t, g)
		if g.GetGameEndFlag() {
			break
		}
		require.NoError(t, g.NextRound())
	}
	assert.Len(t, seen, len(g.GetPlayers()), "親を一度もやらない席がある")
}

// **ラウンド数は席数以上でなければならない。** 少ないと親を一度もやらない
// 席が出て、有利不利が席順で固定される。
func TestKingoConfig_RoundsMustCoverEverySeat(t *testing.T) {
	t.Parallel()
	assert.ErrorIs(t, KingoConfig{Seats: 5, InitialChips: 1000, MinBet: 10, Rounds: 4}.Validate(),
		errKingoRoundsRange)
	assert.NoError(t, KingoConfig{Seats: 5, InitialChips: 1000, MinBet: 10, Rounds: 5}.Validate())
	assert.NoError(t, DefaultKingoConfig().Validate())

	assert.ErrorIs(t, KingoConfig{Seats: 1, InitialChips: 1000, MinBet: 10, Rounds: 10}.Validate(),
		errKingoSeatsRange)
	assert.ErrorIs(t, KingoConfig{Seats: 4, InitialChips: 10, MinBet: 10, Rounds: 10}.Validate(),
		errKingoChipsRange)
	assert.ErrorIs(t, KingoConfig{Seats: 4, InitialChips: 1000, MinBet: 15, Rounds: 10}.Validate(),
		errKingoBetRange)
	// 席数 × 3 枚が山を超えないこと (5 席 × 3 = 15 <= 40)。
	assert.LessOrEqual(t, KingoMaxSeats*KingoHandSize, KingoDeckSize)
}

// **チップは増えも減りもしない。** 親と子の間で動くだけ。
func TestKingo_ChipsAreConserved(t *testing.T) {
	for round := range 30 {
		g := newKingoForTest(t)
		want := kingoTotalChips(g)
		for range g.GetConfig().Rounds {
			kingoPlayRound(t, g)
			assert.Equal(t, want, kingoTotalChips(g),
				"%d 局目でチップ総量が動いた", round)
			if g.GetGameEndFlag() {
				break
			}
			require.NoError(t, g.NextRound())
		}
	}
}

// **親の残高が負にならない。** 3 人が同時に勝ったとき、集める前に払うと
// 手持ちを超えて払い出す。
func TestKingo_BankerNeverGoesNegative(t *testing.T) {
	for range 40 {
		g := newKingoForTest(t)
		for range g.GetConfig().Rounds {
			kingoPlayRound(t, g)
			for i, p := range g.GetPlayers() {
				require.GreaterOrEqual(t, p.GetChips(), 0,
					"席 %d のチップが負になった", i)
			}
			if g.GetGameEndFlag() {
				break
			}
			require.NoError(t, g.NextRound())
		}
	}
}

// **ゲームは必ず終わる。** ラウンド数で必ず打ち切られる。
func TestKingo_AlwaysTerminates(t *testing.T) {
	for range 30 {
		g := newKingoForTest(t)
		for steps := 0; !g.GetGameEndFlag(); steps++ {
			require.Less(t, steps, 200, "ゲームが終わらない")
			kingoPlayRound(t, g)
			if g.GetGameEndFlag() {
				break
			}
			require.NoError(t, g.NextRound())
		}
		assert.LessOrEqual(t, g.GetRoundNumber(), g.GetConfig().Rounds,
			"設定より多く回っている")
	}
}

// **人間が張れなくなったら終わる。**
//
// 精算のときだけ見ていると、**張れないのに張りを待つ盤面**で止まる ──
// 人間は張ることも進むこともできず、ゲームが動かなくなる。子の立場で
// 破産させて、次のラウンドの入口で止まることを見る。
func TestKingo_EndsWhenTheHumanCannotBet(t *testing.T) {
	g := newKingoForTest(t)
	// 親の回は張らないので、子になるラウンドまで進める。
	kingoPlayRound(t, g)
	require.False(t, g.GetGameEndFlag())

	g.GetPlayers()[g.HumanSeat()].SetChips(0)
	require.NoError(t, g.NextRound())
	require.False(t, g.IsHumanBanker(), "人間がまだ親のまま")

	assert.True(t, g.GetGameEndFlag(), "張れないのに続いている")
	assert.ErrorIs(t, g.NextRound(), errKingoFinished)
	assert.ErrorIs(t, g.PlaceBet(10), errKingoFinished)
	assert.ErrorIs(t, g.Deal(), errKingoFinished)
	assert.Nil(t, g.GetHint(), "終局後に助言が出ている")
}

// --- 張りの検証 ---

func TestKingo_BetValidation(t *testing.T) {
	g := newKingoForTest(t)
	for g.IsHumanBanker() && !g.GetGameEndFlag() {
		kingoPlayRound(t, g)
		require.NoError(t, g.NextRound())
	}
	require.False(t, g.IsHumanBanker(), "人間の張り待ちにならない")
	require.Equal(t, KingoPhaseBet, g.GetPhase())

	cfg := g.GetConfig()
	assert.ErrorIs(t, g.PlaceBet(cfg.MinBet-1), errKingoBetAmount)
	assert.ErrorIs(t, g.PlaceBet(KingoMaxBet+10), errKingoBetAmount)
	assert.ErrorIs(t, g.PlaceBet(cfg.MinBet+1), errKingoBetAmount, "刻みを外れた額が通った")
	assert.ErrorIs(t, g.PlaceBet(1000000), errKingoBetAmount, "手持ちを超える額が通った")

	require.NoError(t, g.PlaceBet(cfg.MinBet))
	assert.Equal(t, KingoPhaseResult, g.GetPhase())
	// 張った直後に同じラウンドで張り直せない。
	assert.ErrorIs(t, g.PlaceBet(cfg.MinBet), errKingoWrongPhase)
}

func TestKingo_NextRoundNeedsAFinishedRound(t *testing.T) {
	g := newKingoForTest(t)
	if g.GetPhase() == KingoPhaseBet {
		assert.ErrorIs(t, g.NextRound(), errKingoWrongPhase)
	}
	kingoPlayRound(t, g)
	if !g.GetGameEndFlag() {
		before := g.GetBankerSeat()
		require.NoError(t, g.NextRound())
		assert.Equal(t, (before+1)%len(g.GetPlayers()), g.GetBankerSeat(),
			"親が回っていない")
		assert.Equal(t, 2, g.GetRoundNumber())
	}
}

// --- 助言 ---

// **配る前に手札は見えない。** 助言できるのは張りの重さだけ。
func TestKingo_HintDoesNotPretendToSeeTheHand(t *testing.T) {
	g := newKingoForTest(t)
	require.True(t, g.IsHumanTurn())
	h := g.GetHint()
	require.NotNil(t, h)
	if g.IsHumanBanker() {
		// 親の回は「配る」しか言えない ── 張りは子のもの。
		assert.Equal(t, "deal", h.Action)
	} else {
		assert.Equal(t, "bet", h.Action)
		assert.Equal(t, g.GetConfig().MinBet, h.Amount,
			"見えていない手札を根拠に額を変えている")
	}
	assert.NotEmpty(t, h.Reason)
	kingoPlayRound(t, g)
	if !g.GetGameEndFlag() {
		h := g.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "next", h.Action)
	}
}

func TestKingo_Accessors(t *testing.T) {
	g := newKingoForTest(t)
	assert.Len(t, g.GetPlayers(), KingoDefaultSeats)
	assert.Equal(t, 0, g.HumanSeat())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, DefaultKingoConfig(), g.GetConfig())
	assert.GreaterOrEqual(t, g.WinnerSeat(), 0)
	assert.NotEmpty(t, g.GetActionLog())
	assert.NotNil(t, g.GetResults())

	cfg := KingoConfig{Seats: 3, InitialChips: 500, MinBet: 20, Rounds: 5}
	g.SetConfig(cfg)
	assert.Equal(t, cfg, g.GetConfig())
	g.Reset()
	assert.Len(t, g.GetPlayers(), 3)
}

// **親が払いきれないときも、払うのは持っている分まで。**
//
// 3 人が同時に勝つと、親の手持ちを超える請求が立つ。集める前に払うと親の
// 残高が負になるが、**既定のチップ (1000) と最小の張り (10) では絶対に
// 起きない** ── 実際に払いきれない卓を組まないと、この分岐は一度も踏まれない。
func TestKingo_BankerPaysOnlyWhatItHas(t *testing.T) {
	g := NewDefaultKingo()
	g.Reset()
	players := g.GetPlayers()
	require.GreaterOrEqual(t, len(players), 4)

	banker := g.GetBankerSeat()
	// 親は役なしで手薄。子は全員 2 枚そろいで勝つ。
	players[banker].SetChips(30)
	players[banker].cards = kingoCards(1, 5, 9)
	for i, p := range players {
		if i == banker {
			continue
		}
		p.SetChips(500)
		p.SetBet(20)
		p.cards = kingoCards(10, 10, 2)
	}

	before := kingoTotalChips(g)
	g.resolve()

	for i, p := range players {
		require.GreaterOrEqual(t, p.GetChips(), 0, "席 %d のチップが負になった", i)
	}
	assert.Equal(t, before, kingoTotalChips(g), "チップ総量が動いた")
	// 親が持っていた 30 がちょうど出ていく。
	assert.Zero(t, players[banker].GetChips(), "親に払い残しがある")

	paid := 0
	for i, r := range g.GetResults() {
		if i == banker {
			continue
		}
		assert.GreaterOrEqual(t, r.WonAmount, 0, "勝った席が払っている")
		paid += r.WonAmount
	}
	assert.Equal(t, 30, paid, "親の手持ちを超えて配られている")
}
