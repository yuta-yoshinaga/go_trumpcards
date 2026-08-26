//go:build test && (!js || !wasm || extra)

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// specHand builds cards from (design, value) pairs.
func specHand(spec ...int) []*Card {
	cards := make([]*Card, 0, len(spec)/2)
	for i := 0; i < len(spec); i += 2 {
		cards = append(cards, NewCard(spec[i], spec[i+1], true))
	}
	return cards
}

// newSpecTable builds a table with known hands and a known trump, so nothing
// depends on the shuffle.
func newSpecTable(t *testing.T, trump int, hands ...[]*Card) *Speculation {
	t.Helper()
	cfg := NewDefaultSpeculationConfig()
	cfg.Players = len(hands)
	g := NewSpeculation(cfg)
	g.SetTrumpSuit(trump)
	for i, h := range hands {
		g.GetPlayers()[i].SetHidden(h)
		g.GetPlayers()[i].SetBest(nil)
	}
	g.SetBestSeat(-1)
	g.SetTurnSeat(0)
	g.SetPhase(SpeculationPhaseFlip)
	return g
}

// specFlip flips and answers any auction the flip opens by declining.
//
// **めくると競りが開くことがある。** 競り中は次をめくれないので、競りそのものを
// 試すテスト以外はここで閉じる。断るだけなので札もチップも動かず、検査したい
// 規則には影響しない。
func specFlip(t *testing.T, g *Speculation) {
	t.Helper()
	require.NoError(t, g.Flip())
	if g.GetPhase() == SpeculationPhaseAuction {
		require.NoError(t, g.Decline())
	}
}

func TestSpeculation_OnlyTrumpsCanTakeTheLead(t *testing.T) {
	// **切り札でなければ何を出しても関係ない。** スートが違う札はこのゲームでは
	// ただの紙で、競りの対象にすらならない。
	g := newSpecTable(t, 0,
		specHand(1, 13), // 座席0: ♥K — 切り札ではない
		specHand(0, 2),  // 座席1: ♠2 — 弱いが切り札
	)
	specFlip(t, g)
	assert.Equal(t, -1, g.GetBestSeat(), "切り札でない K は最高札にならない")

	specFlip(t, g)
	assert.Equal(t, 1, g.GetBestSeat(), "弱くても切り札なら先頭に立つ")
}

func TestSpeculation_AceIsTheHighestTrump(t *testing.T) {
	// A は 1 のままだと最弱になる。**14 として扱う。**
	g := newSpecTable(t, 0, specHand(0, 13), specHand(0, 1))
	specFlip(t, g) // ♠K
	require.Equal(t, 0, g.GetBestSeat())
	specFlip(t, g) // ♠A
	assert.Equal(t, 1, g.GetBestSeat(), "A は K より強い")
}

func TestSpeculation_ALowerTrumpDoesNotTakeTheLead(t *testing.T) {
	g := newSpecTable(t, 0, specHand(0, 12), specHand(0, 5))
	specFlip(t, g)
	require.Equal(t, 0, g.GetBestSeat())
	specFlip(t, g)
	assert.Equal(t, 0, g.GetBestSeat(), "5 は Q を上回らない")
	assert.Nil(t, g.GetPlayers()[1].GetBest())
}

func TestSpeculation_TakingTheLeadStripsThePreviousHolder(t *testing.T) {
	// **最高札は 1 枚しか存在しない。** 前の持ち主から取り上げないと、
	// 決着時に複数人が「最高札を持っている」ことになる。
	g := newSpecTable(t, 0, specHand(0, 5), specHand(0, 9))
	specFlip(t, g)
	require.NotNil(t, g.GetPlayers()[0].GetBest())
	specFlip(t, g)
	assert.Nil(t, g.GetPlayers()[0].GetBest(), "上回られた席は最高札を失う")
	assert.NotNil(t, g.GetPlayers()[1].GetBest())
}

func TestSpeculation_TheHighestTrumpTakesThePot(t *testing.T) {
	g := newSpecTable(t, 0, specHand(0, 5), specHand(0, 9))
	g.SetPot(100)
	before := g.GetPlayers()[1].GetChips()

	specFlip(t, g)
	specFlip(t, g)

	assert.Equal(t, 1, g.GetWinnerSeat())
	assert.Equal(t, before+100, g.GetPlayers()[1].GetChips())
	assert.Equal(t, 0, g.GetPot(), "ポットは空になる")
}

func TestSpeculation_ARoundWithNoTrumpReturnsTheStakes(t *testing.T) {
	// **切り札が 1 枚も出ないことがある。** 繰り越すと、降りた席の負担だけが
	// 積み上がる。参加料は返す。
	g := newSpecTable(t, 0, specHand(1, 5), specHand(2, 9))
	g.SetPot(100)
	c0 := g.GetPlayers()[0].GetChips()
	c1 := g.GetPlayers()[1].GetChips()

	specFlip(t, g)
	specFlip(t, g)

	assert.Equal(t, -1, g.GetWinnerSeat(), "勝者なし")
	assert.Equal(t, c0+50, g.GetPlayers()[0].GetChips())
	assert.Equal(t, c1+50, g.GetPlayers()[1].GetChips())
	assert.Equal(t, 0, g.GetPot())
}

func TestSpeculation_AcceptingAnOfferMovesBothTheCardAndTheChips(t *testing.T) {
	g := newSpecTable(t, 0, specHand(0, 13), specHand(0, 2))
	require.NoError(t, g.Flip())
	if g.GetPhase() == SpeculationPhaseAuction {
		require.NoError(t, g.Decline())
	} // 座席0 が ♠K で先頭に立つ
	card := g.GetPlayers()[0].GetBest()
	require.NotNil(t, card)

	g.SetPhase(SpeculationPhaseAuction)
	g.SetOffer(1, 0, 30) // 座席1 が座席0 の札に 30 を提示
	g.GetPlayers()[1].SetChips(100)
	g.GetPlayers()[0].SetChips(100)

	require.NoError(t, g.Accept())

	assert.Equal(t, 130, g.GetPlayers()[0].GetChips(), "売り手はチップを受け取る")
	assert.Equal(t, 70, g.GetPlayers()[1].GetChips(), "買い手は払う")
	assert.Nil(t, g.GetPlayers()[0].GetBest(), "売った札は手元を離れる")
	assert.Equal(t, card, g.GetPlayers()[1].GetBest(), "買い手が同じ札を持つ")
	assert.Equal(t, 1, g.GetBestSeat())
}

func TestSpeculation_DecliningMovesNothing(t *testing.T) {
	g := newSpecTable(t, 0, specHand(0, 13), specHand(0, 2))
	require.NoError(t, g.Flip())
	if g.GetPhase() == SpeculationPhaseAuction {
		require.NoError(t, g.Decline())
	}
	card := g.GetPlayers()[0].GetBest()

	g.SetPhase(SpeculationPhaseAuction)
	g.SetOffer(1, 0, 30)
	g.GetPlayers()[0].SetChips(100)
	g.GetPlayers()[1].SetChips(100)

	require.NoError(t, g.Decline())

	assert.Equal(t, 100, g.GetPlayers()[0].GetChips())
	assert.Equal(t, 100, g.GetPlayers()[1].GetChips())
	assert.Equal(t, card, g.GetPlayers()[0].GetBest(), "断れば札は残る")
	assert.Equal(t, 0, g.GetBestSeat())
}

func TestSpeculation_AnUnaffordableOfferIsNotSettled(t *testing.T) {
	// 払えない申し出で札だけ動くと、チップが湧いたのと同じことになる。
	g := newSpecTable(t, 0, specHand(0, 13), specHand(0, 2))
	require.NoError(t, g.Flip())
	if g.GetPhase() == SpeculationPhaseAuction {
		require.NoError(t, g.Decline())
	}
	card := g.GetPlayers()[0].GetBest()

	g.SetPhase(SpeculationPhaseAuction)
	g.SetOffer(1, 0, 500)
	g.GetPlayers()[0].SetChips(100)
	g.GetPlayers()[1].SetChips(10) // 500 は払えない

	// **黙って閉じない。** 以前は競りを閉じて手番を進めていたので、
	// プレイヤーには理由も出ないまま申し出だけが消えていた。
	require.Error(t, g.Accept(), "払えない申し出は成立しない")

	assert.Equal(t, 100, g.GetPlayers()[0].GetChips())
	assert.Equal(t, 10, g.GetPlayers()[1].GetChips())
	assert.Equal(t, card, g.GetPlayers()[0].GetBest(), "札は動かない")
	assert.Equal(t, SpeculationPhaseAuction, g.GetPhase(),
		"競りは開いたまま —— 断る道が残る")
	require.NoError(t, g.Decline(), "断って先へ進める")
}

func TestSpeculation_BidOnlyGoesUp(t *testing.T) {
	// 提示額より低い値を「申し出」と呼ぶと、断るのと区別が付かない。
	g := newSpecTable(t, 0, specHand(0, 2), specHand(0, 13))
	g.SetPhase(SpeculationPhaseAuction)
	g.SetOffer(0, 1, 30) // 人間 (座席0) が買い手
	g.GetPlayers()[0].SetChips(100)
	g.GetPlayers()[1].SetBest(specHand(0, 13)[0])
	g.SetBestSeat(1)

	assert.Error(t, g.Bid(30), "同額は上乗せではない")
	assert.Error(t, g.Bid(10), "下回る額は上乗せではない")
	assert.Error(t, g.Bid(500), "所持チップを超える額は置けない")

	require.NoError(t, g.Bid(40))
	assert.Equal(t, 60, g.GetPlayers()[0].GetChips(), "上乗せした額で成立する")
	assert.Equal(t, 0, g.GetBestSeat())
}

func TestSpeculation_ActionsAreRejectedOutsideTheirPhase(t *testing.T) {
	g := newSpecTable(t, 0, specHand(0, 5), specHand(0, 9))
	assert.Error(t, g.Accept(), "競り中でなければ受けられない")
	assert.Error(t, g.Decline(), "競り中でなければ断れない")
	assert.Error(t, g.Bid(10), "競り中でなければ値を付けられない")
	assert.Error(t, g.NextRound(), "決着前に次のラウンドへは進めない")

	g.SetPhase(SpeculationPhaseAuction)
	assert.Error(t, g.Flip(), "競り中はめくれない")
}

func TestSpeculation_AnAuctionWithNoOfferCannotBeAnswered(t *testing.T) {
	g := newSpecTable(t, 0, specHand(0, 5), specHand(0, 9))
	g.SetPhase(SpeculationPhaseAuction)
	g.SetOffer(-1, -1, 0)
	assert.Error(t, g.Accept())
	assert.Error(t, g.Decline())
}

func TestSpeculation_FlippingAnEmptySeatSkipsIt(t *testing.T) {
	// 席ごとに残り枚数が違う。空の席で止まると進行が詰まる。
	g := newSpecTable(t, 0, nil, specHand(0, 9))
	specFlip(t, g) // 座席0 は空 → 手番だけ進む
	assert.Equal(t, 1, g.GetTurnSeat())
}

func TestSpeculation_RoundsRunOutAndEndTheGame(t *testing.T) {
	cfg := NewDefaultSpeculationConfig()
	cfg.Players, cfg.Rounds = 2, 1
	g := NewSpeculation(cfg)
	g.SetTrumpSuit(0)
	g.GetPlayers()[0].SetHidden(specHand(0, 5))
	g.GetPlayers()[1].SetHidden(specHand(0, 9))
	g.SetBestSeat(-1)
	g.SetTurnSeat(0)
	g.SetPhase(SpeculationPhaseFlip)

	specFlip(t, g)
	specFlip(t, g)

	assert.Equal(t, SpeculationPhaseGameEnd, g.GetPhase())
	assert.True(t, g.GetGameEndFlag())
	assert.Error(t, g.NextRound(), "終了後に次のラウンドは無い")
}

func TestSpeculation_NextRoundKeepsChipsAndRoundCount(t *testing.T) {
	cfg := NewDefaultSpeculationConfig()
	cfg.Players, cfg.Rounds = 2, 5
	g := NewSpeculation(cfg)
	g.SetTrumpSuit(0)
	g.GetPlayers()[0].SetHidden(specHand(0, 5))
	g.GetPlayers()[1].SetHidden(specHand(0, 9))
	g.SetBestSeat(-1)
	g.SetTurnSeat(0)
	g.SetPhase(SpeculationPhaseFlip)
	specFlip(t, g)
	specFlip(t, g)
	require.Equal(t, SpeculationPhaseResult, g.GetPhase())

	chipsBefore := g.GetPlayers()[1].GetChips()
	round := g.GetRoundNo()
	require.NoError(t, g.NextRound())

	assert.Equal(t, round, g.GetRoundNo(), "ラウンド数は巻き戻らない")
	assert.Equal(t, SpeculationPhaseFlip, g.GetPhase())
	assert.Less(t, g.GetPlayers()[1].GetChips(), chipsBefore, "次の参加料が引かれる")
	assert.Equal(t, SpeculationCardsPerPlayer, g.GetPlayers()[0].GetHiddenCount(), "配り直される")
}

func TestSpeculation_ResetCollectsTheStakeFromEverySeat(t *testing.T) {
	cfg := NewDefaultSpeculationConfig()
	cfg.Players, cfg.Stake = 3, 10
	g := NewSpeculation(cfg)
	assert.Equal(t, 30, g.GetPot(), "3 席 × 10")
	for i, p := range g.GetPlayers() {
		assert.Equal(t, cfg.InitialChips-10, p.GetChips(), "座席 %d", i)
	}
}

func TestSpeculation_ABrokeSeatPaysWhatItHasAndStays(t *testing.T) {
	// **席を弾かない。** 座席番号がずれるとラウンドを跨いだ集計が崩れる。
	cfg := NewDefaultSpeculationConfig()
	cfg.Players, cfg.Stake = 2, 10
	g := NewSpeculation(cfg)
	// 未精算ポットを先に空にする。**返却が絡むとこのテストの主題がぼやける**
	// —— 返ってきた分で参加料を払えてしまい、「払えない席」でなくなる。
	g.SetPot(0)
	g.GetPlayers()[1].SetChips(3)
	g.SetRoundNo(0)
	g.Reset()

	assert.Len(t, g.GetPlayers(), 2, "席は減らない")
	assert.Equal(t, 0, g.GetPlayers()[1].GetChips(), "出せるだけ出す")
	assert.Equal(t, 13, g.GetPot(), "10 + 3")
}

func TestSpeculation_ConfigNormalizesOutOfRangeValues(t *testing.T) {
	c := SpeculationConfig{Players: 99, InitialChips: -5, Stake: 0, Rounds: 0}
	c.Normalize()
	assert.Equal(t, SpeculationDefaultPlayers, c.Players)
	assert.Equal(t, SpeculationDefaultChips, c.InitialChips)
	assert.Equal(t, SpeculationDefaultStake, c.Stake)
	assert.Equal(t, SpeculationDefaultRounds, c.Rounds)

	ok := SpeculationConfig{Players: 6, InitialChips: 500, Stake: 25, Rounds: 10}
	ok.Normalize()
	assert.Equal(t, SpeculationConfig{Players: 6, InitialChips: 500, Stake: 25, Rounds: 10}, ok,
		"範囲内の設定は触らない")
}

func TestSpeculation_JSONRoundTripKeepsEveryField(t *testing.T) {
	// **非公開フィールドだけの構造体は marshaller が無いと `{}` になる。**
	g := newSpecTable(t, 2, specHand(2, 13, 2, 4), specHand(2, 9, 1, 3))
	g.SetPot(77)
	specFlip(t, g)
	g.SetPhase(SpeculationPhaseAuction)
	g.SetOffer(1, 0, 42)

	data, err := json.Marshal(g)
	require.NoError(t, err)
	assert.Greater(t, len(data), 100, "`{}` で出荷されていないこと")

	var got Speculation
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, g.GetPhase(), got.GetPhase())
	assert.Equal(t, g.GetTrumpSuit(), got.GetTrumpSuit())
	assert.Equal(t, g.GetPot(), got.GetPot())
	assert.Equal(t, g.GetTurnSeat(), got.GetTurnSeat())
	assert.Equal(t, g.GetOfferFrom(), got.GetOfferFrom())
	assert.Equal(t, g.GetOfferTo(), got.GetOfferTo())
	assert.Equal(t, g.GetOfferAmount(), got.GetOfferAmount())
	assert.Equal(t, g.GetBestSeat(), got.GetBestSeat())
	assert.Equal(t, g.GetRoundNo(), got.GetRoundNo())
	assert.Equal(t, g.GetConfig(), got.GetConfig())
	require.Len(t, got.GetPlayers(), len(g.GetPlayers()))
	for i, p := range g.GetPlayers() {
		assert.Equal(t, p.GetChips(), got.GetPlayers()[i].GetChips(), "座席 %d のチップ", i)
		assert.Equal(t, p.GetHiddenCount(), got.GetPlayers()[i].GetHiddenCount(), "座席 %d の伏せ札", i)
	}
	assert.NotNil(t, got.GetPlayers()[0].GetBest(), "最高札が往復で消えないこと")
	assert.Equal(t, g.GetActionLog(), got.GetActionLog())

	// 復元した卓でそのまま続けられること。
	require.NoError(t, got.Decline())
	assert.Equal(t, SpeculationPhaseFlip, got.GetPhase())
}

func TestSpeculation_UnmarshalClampsSeatsIntoRange(t *testing.T) {
	// **席番号が席数の外を指したまま戻ると、次のめくりで即座に落ちる。**
	raw := `{"pl":[{"n":"You","c":100},{"n":"CPU1","c":100}],"tn":99,"bs":42,"of":-7,"ph":0}`
	var g Speculation
	require.NoError(t, json.Unmarshal([]byte(raw), &g))
	assert.Equal(t, 0, g.GetTurnSeat(), "範囲外の手番は 0 に戻す")
	assert.Equal(t, -1, g.GetBestSeat(), "範囲外の最高札席は「無し」に戻す")
	assert.Equal(t, -1, g.GetOfferFrom())
	require.NotPanics(t, func() { _ = g.Flip() }, "復元直後にめくっても落ちない")
}

func TestSpeculation_UnmarshalRejectsOversizedArrays(t *testing.T) {
	huge := "["
	for i := range speculationMaxSliceLen + 1 {
		if i > 0 {
			huge += ","
		}
		huge += "null"
	}
	huge += "]"
	assert.Error(t, json.Unmarshal([]byte(`{"pl":`+huge+`}`), new(Speculation)))
	assert.Error(t, json.Unmarshal([]byte(`not json`), new(Speculation)))
}

func TestSpeculation_UnmarshalRebuildsASeatlessTable(t *testing.T) {
	// 席が 0 の卓は進行できない。既定人数で作り直す。
	var g Speculation
	require.NoError(t, json.Unmarshal([]byte(`{"pl":[]}`), &g))
	assert.Len(t, g.GetPlayers(), SpeculationDefaultPlayers)
	require.NotPanics(t, func() { _ = g.Flip() })
}

func TestSpeculation_ResettingTwiceDoesNotDestroyChips(t *testing.T) {
	// **Reset は何度でも呼ばれる。** コンストラクタが一度、ユースケースが
	// 起動時にもう一度。そのたびに参加料だけ取って古いポットを捨てると
	// チップが消える —— 実測で 200 → 190 → 180 と減り、ポットは 40 のまま
	// だった (4 席で 40 チップが蒸発)。
	cfg := NewDefaultSpeculationConfig()
	cfg.Players, cfg.Stake, cfg.InitialChips = 4, 10, 200
	g := NewSpeculation(cfg)

	total := func() int {
		n := g.GetPot()
		for _, p := range g.GetPlayers() {
			n += p.GetChips()
		}
		return n
	}
	want := 4 * 200
	require.Equal(t, want, total(), "卓上の総額は初期チップの合計")

	for range 5 {
		g.Reset()
		assert.Equal(t, want, total(), "Reset を繰り返しても総額は変わらない")
	}
	assert.Equal(t, 190, g.GetPlayers()[0].GetChips(), "参加料は 1 ラウンド分しか引かれない")
	assert.Equal(t, 40, g.GetPot())
}

func TestSpeculation_ChipsAreConservedAcrossARound(t *testing.T) {
	// 競りも決着も、卓上の総額を変えない。**チップが湧く/消える形のバグは
	// 個別の額を見ていると通り抜ける。**
	g := newSpecTable(t, 0, specHand(0, 5, 0, 9), specHand(0, 13, 1, 3))
	total := func() int {
		n := g.GetPot()
		for _, p := range g.GetPlayers() {
			n += p.GetChips()
		}
		return n
	}
	want := total()

	for range 4 {
		if g.GetPhase() == SpeculationPhaseAuction {
			require.NoError(t, g.Accept())
		} else if g.GetPhase() == SpeculationPhaseFlip {
			require.NoError(t, g.Flip())
		}
		assert.Equal(t, want, total(), "1 手ごとに総額が保存されること")
	}
}

func TestSpeculation_ResetStartsANewGameFromRoundOne(t *testing.T) {
	// **Reset は「新しいゲーム」、NextRound は「次のラウンド」。** ラウンド数を
	// 持ち越すと、終局画面からリセットしたとき roundNo が上限のままで、最初の
	// 決着で即座にまた終局する。
	cfg := NewDefaultSpeculationConfig()
	cfg.Players, cfg.Rounds = 2, 3
	g := NewSpeculation(cfg)
	g.SetRoundNo(3) // 終局した状態
	g.SetPhase(SpeculationPhaseGameEnd)

	g.Reset()
	assert.Equal(t, 0, g.GetRoundNo(), "リセットは 1 ラウンド目から")
	assert.Equal(t, SpeculationPhaseFlip, g.GetPhase())
	assert.False(t, g.GetGameEndFlag())

	// 実際に 1 ラウンド回しても終局しないこと (規定は 3 ラウンド)。
	g.SetTrumpSuit(0)
	g.GetPlayers()[0].SetHidden(specHand(0, 5))
	g.GetPlayers()[1].SetHidden(specHand(0, 9))
	g.SetBestSeat(-1)
	g.SetTurnSeat(0)
	specFlip(t, g)
	specFlip(t, g)
	assert.Equal(t, SpeculationPhaseResult, g.GetPhase(), "1 ラウンド目で終局しない")
	assert.False(t, g.GetGameEndFlag())
}

func TestSpeculation_ACPUNeverAsksMoreThanYouCanPay(t *testing.T) {
	// 払えない額を提示されると、受けても黙って競りが閉じるだけで理由が出ない。
	// **提示の時点で手の届く額に丸める。**
	cfg := NewDefaultSpeculationConfig()
	cfg.Players = 3
	g := NewSpeculation(cfg)
	g.SetTrumpSuit(0)
	g.SetPot(10000) // CPU の評価額を吊り上げる
	g.GetPlayers()[0].SetChips(5)
	g.GetPlayers()[1].SetHidden(specHand(0, 1))
	g.GetPlayers()[2].SetHidden(nil)
	g.GetPlayers()[0].SetHidden(nil)
	g.SetBestSeat(-1)
	g.SetTurnSeat(1)
	g.SetPhase(SpeculationPhaseFlip)

	require.NoError(t, g.Flip())
	if g.GetPhase() == SpeculationPhaseAuction {
		assert.LessOrEqual(t, g.GetOfferAmount(), 5, "所持チップを超える額は提示しない")
		assert.NoError(t, g.Accept(), "提示された額なら必ず払える")
	}
}
