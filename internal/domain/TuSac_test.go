//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTuSacForTest は配り終えた卓を返す。
func newTuSacForTest(t *testing.T) *TuSac {
	t.Helper()
	g := NewDefaultTuSac()
	g.Reset()
	return g
}

// tuSacPlayRound は人間の席を機械的に打たせてラウンドを閉じる。
//
// **人間の席は残したまま駆動する。** 全席を CPU にすると `HumanSeat()` が
// 席 0 に落ちて、その席を誰も動かさないまま盤面が止まる。
func tuSacPlayRound(t *testing.T, g *TuSac) {
	t.Helper()
	for steps := 0; g.GetPhase() != TuSacPhaseRoundEnd && !g.GetGameEndFlag(); steps++ {
		require.Less(t, steps, 600, "ラウンドが終わらない (フェーズ %d)", g.GetPhase())
		if !g.IsHumanTurn() {
			// CPU は自動で進む。ここに来るのは決着したときだけ。
			break
		}
		switch g.GetPhase() {
		case TuSacPhaseDraw:
			// **人間が引く場面になったなら、山は空であってはならない。**
			// 空のまま手番が回ると `Draw` が必ず失敗し、人間は引くことも
			// 進むこともできない ── ラウンドを閉じる判断が人間の手番より
			// 後ろにあると、この盤面が作れてしまう。
			require.Positive(t, g.GetStockCount(),
				"山が空なのに人間が引く場面になっている")
			require.NoError(t, g.Draw(false))
		case TuSacPhaseDiscard:
			// 出せる組み合わせがあれば出してから捨てる。
			for range TuSacHandSize {
				h := g.GetHint()
				if h == nil || h.Action != "meld" {
					break
				}
				require.NoError(t, g.Meld(h.Indexes))
			}
			// **出し切ったらそこで上がり。** 捨てる札が残っていない。
			if g.GetPhase() == TuSacPhaseRoundEnd || g.GetGameEndFlag() {
				return
			}
			require.NoError(t, g.Discard(0))
		}
	}
}

// --- 配札 ---

func TestTuSac_DealsTwentyToEverySeat(t *testing.T) {
	g := newTuSacForTest(t)
	for i, p := range g.GetPlayers() {
		// CPU の席は既に打っているので、人間の席だけ厳密に見る。
		if !p.GetIsHuman() {
			continue
		}
		assert.Len(t, p.GetCards(), TuSacHandSize, "席 %d の枚数", i)
	}
	// **山を配り切らない。** 引くところが無いと手番が「捨てるだけ」になる。
	assert.Positive(t, g.GetStockCount(), "山が空で始まっている")
	assert.Positive(t, g.GetDiscardCount(), "捨て札が空で始まっている")
}

// **札は増えも減りもしない。** 112 枚が山・捨て札・手札・場に散っているだけ。
func TestTuSac_CardsAreConserved(t *testing.T) {
	for round := range 20 {
		g := newTuSacForTest(t)
		for range 3 {
			total := g.GetStockCount() + g.GetDiscardCount()
			for _, p := range g.GetPlayers() {
				total += len(p.GetCards())
				for _, m := range p.GetMelds() {
					total += len(m.Cards)
				}
			}
			assert.Equal(t, TuSacDeckSize, total, "%d 局目で札の総数が %d に変わった", round, total)
			if g.GetPhase() == TuSacPhaseRoundEnd || g.GetGameEndFlag() {
				break
			}
			tuSacPlayRound(t, g)
		}
	}
}

// **ラウンドは必ず終わる。** 上がりは滅多に出ないので、終了性は山の枯渇が
// 担保する ── 引けない手番を回し続けると進まない。
func TestTuSac_RoundsAlwaysTerminate(t *testing.T) {
	for round := range 20 {
		g := newTuSacForTest(t)
		tuSacPlayRound(t, g)
		require.True(t, g.GetPhase() == TuSacPhaseRoundEnd || g.GetGameEndFlag(),
			"%d 局目が終わらなかった", round)
		// 山切れか上がりのどちらかで終わっている。
		assert.True(t, g.GetStockCount() == 0 || g.GetWentOutSeat() >= 0,
			"山も残っていて上がってもいないのに終わった")
	}
}

// **ゲームは設定ラウンド数で終わる。**
func TestTuSac_GameEndsAfterTheConfiguredRounds(t *testing.T) {
	g := newTuSacForTest(t)
	for steps := 0; !g.GetGameEndFlag(); steps++ {
		require.Less(t, steps, 100, "ゲームが終わらない")
		tuSacPlayRound(t, g)
		if g.GetGameEndFlag() {
			break
		}
		require.NoError(t, g.NextRound())
	}
	assert.Equal(t, g.GetConfig().Rounds, g.GetRoundNumber())
	assert.ErrorIs(t, g.NextRound(), errTuSacFinished)
	assert.ErrorIs(t, g.Draw(false), errTuSacFinished)
	assert.ErrorIs(t, g.Discard(0), errTuSacFinished)
	assert.ErrorIs(t, g.Meld([]int{0, 1, 2}), errTuSacFinished)
	assert.Nil(t, g.GetHint(), "終局後に助言が出ている")
}

// --- 手番 ---

// **引く前に捨てられない。** 順序が崩れると手札が減り続ける。
func TestTuSac_MustDrawBeforeDiscarding(t *testing.T) {
	g := newTuSacForTest(t)
	require.True(t, g.IsHumanTurn())
	require.Equal(t, TuSacPhaseDraw, g.GetPhase())

	assert.ErrorIs(t, g.Discard(0), errTuSacWrongPhase)
	assert.ErrorIs(t, g.Meld([]int{0, 1, 2}), errTuSacWrongPhase)

	require.NoError(t, g.Draw(false))
	assert.Equal(t, TuSacPhaseDiscard, g.GetPhase())
	// 引いたあとに引き直せない。
	assert.ErrorIs(t, g.Draw(false), errTuSacWrongPhase)
}

// **引くと手札が 1 枚増え、捨てると 1 枚減る。**
func TestTuSac_DrawAndDiscardMoveOneCard(t *testing.T) {
	g := newTuSacForTest(t)
	require.True(t, g.IsHumanTurn())
	p := g.GetPlayers()[g.HumanSeat()]

	before := len(p.GetCards())
	stockBefore := g.GetStockCount()
	require.NoError(t, g.Draw(false))
	assert.Len(t, p.GetCards(), before+1)
	assert.Equal(t, stockBefore-1, g.GetStockCount())

	// **捨てると CPU の手番もその場で走る。** 捨て札は自分の 1 枚だけでなく
	// CPU のぶんも増えるので、ちょうど +1 と書くと落ちる (実測でそうなった)。
	// 自分の手札は CPU の手番に影響されないので、そちらを厳密に見る。
	discardBefore := g.GetDiscardCount()
	discarded := p.GetCards()[0]
	require.NoError(t, g.Discard(0))
	assert.Len(t, p.GetCards(), before, "自分の手札が 1 枚減っていない")
	assert.Greater(t, g.GetDiscardCount(), discardBefore, "捨て札が増えていない")
	// **同じ色・同じ駒が 4 枚あるので、値では札を特定できない。** 値で
	// 比べると「別の同じ札」に当たって落ちる (実測でそうなった) ので、
	// ポインタの同一性で見る。
	for _, c := range p.GetCards() {
		assert.NotSame(t, discarded, c, "捨てたその札が手札に残っている")
	}
}

// **捨て札から引くと、その札が手に入る。**
func TestTuSac_DrawFromDiscardTakesTheTopCard(t *testing.T) {
	g := newTuSacForTest(t)
	require.True(t, g.IsHumanTurn())
	top := g.GetDiscardTop()
	require.NotNil(t, top)
	countBefore := g.GetDiscardCount()

	require.NoError(t, g.Draw(true))
	assert.Equal(t, countBefore-1, g.GetDiscardCount())
	found := false
	for _, c := range g.GetPlayers()[g.HumanSeat()].GetCards() {
		if c == top {
			found = true
		}
	}
	assert.True(t, found, "拾った札が手札に入っていない")
}

func TestTuSac_DiscardIndexIsChecked(t *testing.T) {
	g := newTuSacForTest(t)
	require.True(t, g.IsHumanTurn())
	require.NoError(t, g.Draw(false))
	assert.ErrorIs(t, g.Discard(-1), errTuSacBadCardIndex)
	assert.ErrorIs(t, g.Discard(999), errTuSacBadCardIndex)
}

// **組み合わせでない札は場に出せない。**
func TestTuSac_MeldRejectsANonCombination(t *testing.T) {
	g := newTuSacForTest(t)
	require.True(t, g.IsHumanTurn())
	require.NoError(t, g.Draw(false))

	p := g.GetPlayers()[g.HumanSeat()]
	// 3 枚の適当な組が必ずメルドとは限らないので、メルドでない組を探す。
	for a := range len(p.GetCards()) {
		for b := a + 1; b < len(p.GetCards()); b++ {
			for c := b + 1; c < len(p.GetCards()); c++ {
				if _, kind := TuSacFindMeld(p.GetCards(), []int{a, b, c}); kind == TuSacMeldNone {
					assert.ErrorIs(t, g.Meld([]int{a, b, c}), errTuSacNotAMeld)
					return
				}
			}
		}
	}
	t.Skip("手札 21 枚がすべてメルドになる配り (起こり得ないが安全側)")
}

// **出した組み合わせは手札から消える。**
func TestTuSac_MeldMovesCardsOutOfTheHand(t *testing.T) {
	g := newTuSacForTest(t)
	require.True(t, g.IsHumanTurn())
	require.NoError(t, g.Draw(false))

	p := g.GetPlayers()[g.HumanSeat()]
	// 出せる組み合わせが出るまで、この配りでは見つからないこともある。
	h := g.GetHint()
	if h == nil || h.Action != "meld" {
		t.Skip("この配りでは出せる組み合わせが無い")
	}
	before := len(p.GetCards())
	meldsBefore := len(p.GetMelds())

	require.NoError(t, g.Meld(h.Indexes))
	assert.Len(t, p.GetCards(), before-len(h.Indexes), "出した枚数ぶん減っていない")
	assert.Len(t, p.GetMelds(), meldsBefore+1)
	assert.Positive(t, p.MeldPoints())
}

// --- 得点 ---

// **得点はメルド - 手残り。** 手札を抱えたままだと減点になる。
func TestTuSac_ScoresMeldsMinusHeldCards(t *testing.T) {
	g := newTuSacForTest(t)
	tuSacPlayRound(t, g)
	require.Equal(t, TuSacPhaseRoundEnd, g.GetPhase())

	results := g.GetResults()
	require.Len(t, results, len(g.GetPlayers()))
	for i, r := range results {
		p := g.GetPlayers()[i]
		assert.Equal(t, p.MeldPoints(), r.MeldPoints, "席 %d のメルド点", i)
		assert.Equal(t, len(p.GetCards()), r.HandPenalty, "席 %d の手残り", i)
		assert.Equal(t, r.MeldPoints-r.HandPenalty, r.RoundScore, "席 %d の差し引き", i)
		assert.Equal(t, r.RoundScore, p.GetRoundScore())
	}
}

func TestTuSac_Accessors(t *testing.T) {
	g := newTuSacForTest(t)
	assert.Len(t, g.GetPlayers(), TuSacDefaultSeats)
	assert.Equal(t, 0, g.HumanSeat())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, DefaultTuSacConfig(), g.GetConfig())
	assert.GreaterOrEqual(t, g.WinnerSeat(), 0)
	assert.NotEmpty(t, g.GetActionLog())
	assert.Equal(t, -1, g.GetWentOutSeat(), "配った直後に上がりが立っている")

	cfg := TuSacConfig{Seats: 2, Rounds: 2}
	g.SetConfig(cfg)
	assert.Equal(t, cfg, g.GetConfig())
	g.Reset()
	assert.Len(t, g.GetPlayers(), 2)
}

// --- 助言 ---

func TestTuSac_HintFollowsThePhase(t *testing.T) {
	g := newTuSacForTest(t)
	require.True(t, g.IsHumanTurn())

	h := g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "draw", h.Action, "引く場面で引く以外を薦めている")

	require.NoError(t, g.Draw(false))
	h = g.GetHint()
	require.NotNil(t, h)
	assert.Contains(t, []string{"meld", "discard"}, h.Action)
	assert.NotEmpty(t, h.Indexes, "薦める札が空")

	tuSacPlayRound(t, g)
	if !g.GetGameEndFlag() {
		h = g.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "next", h.Action)
	}
}

// **席数を変えても、人間が空の山に当たらない。**
//
// 既定の 4 席では山 31 枚がちょうど CPU の手番で尽きるので、ラウンドを閉じる
// 判断がどこにあっても人間は困らない ── **席数が変わると尽きる席も変わる**。
// 3 席なら山 51 枚が 3 番目の席で尽き、その次は人間なので、閉じる判断が人間の
// 手番の検査より後ろにあると「引けない Draw を待つ盤面」が作れてしまう。
func TestTuSac_HumanNeverFacesAnEmptyStock(t *testing.T) {
	for _, seats := range []int{2, 3, 4} {
		t.Run("seats="+string(rune('0'+seats)), func(t *testing.T) {
			for range 6 {
				g := NewTuSac(NewTuSacPlayersForTable(seats), TuSacConfig{Seats: seats, Rounds: 2})
				g.Reset()
				tuSacPlayRound(t, g)
				require.True(t, g.GetPhase() == TuSacPhaseRoundEnd || g.GetGameEndFlag(),
					"%d 席でラウンドが終わらなかった", seats)
			}
		})
	}
}

// **手札を出し切ったらそこで上がり。** 捨てる札が残っていないので、
// 「捨てるまで手番が終わらない」規則のままだと、上がった席が捨てられない
// まま進めない盤面で固まる ── 21 枚が 3a + 5b でちょうど割り切れると
// 実際に起きる (2 席の卓で実測して見つかった)。
func TestTuSac_MeldingOutEndsTheRound(t *testing.T) {
	g := NewTuSac(NewTuSacPlayersForTable(2), TuSacConfig{Seats: 2, Rounds: 2})
	g.Reset()
	require.True(t, g.IsHumanTurn())
	require.NoError(t, g.Draw(false))

	// 手札を丸ごと差し替えて、卒 5 枚 × 3 + 同色同種 3 枚 × 2 = 21 枚にする。
	p := g.GetPlayers()[g.HumanSeat()]
	p.cards = p.cards[:0]
	for range 15 {
		p.AddCard(tsCard(TuSacColorRed, TuSacPieceSoldier))
	}
	for range 3 {
		p.AddCard(tsCard(TuSacColorGreen, TuSacPieceElephant))
	}
	for range 3 {
		p.AddCard(tsCard(TuSacColorWhite, TuSacPieceHorse))
	}
	require.Len(t, p.GetCards(), TuSacHandSize+1)

	// 出せるだけ出す。最後の 1 組で手札が空になる。
	for range TuSacHandSize {
		h := g.GetHint()
		if h == nil || h.Action != "meld" {
			break
		}
		require.NoError(t, g.Meld(h.Indexes))
		if g.GetPhase() == TuSacPhaseRoundEnd {
			break
		}
	}

	assert.Empty(t, p.GetCards(), "出し切れていない")
	assert.Equal(t, TuSacPhaseRoundEnd, g.GetPhase(), "出し切ってもラウンドが終わらない")
	assert.Equal(t, g.HumanSeat(), g.GetWentOutSeat(), "上がった席が記録されていない")
	// 上がった席は手残りが 0 なので減点が無い。
	assert.Zero(t, g.GetResults()[g.HumanSeat()].HandPenalty)
	assert.Positive(t, g.GetResults()[g.HumanSeat()].RoundScore, "出し切ったのに得点が無い")
}
