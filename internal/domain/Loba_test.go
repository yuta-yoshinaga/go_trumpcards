//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lbCard(design, value int) *Card { return NewCard(design, value, true) }
func lbJoker() *Card                 { return NewCard(CardDesignJoker, 0, true) }

func TestLoba_TheDeckIsTwoPacksPlusFourJokers(t *testing.T) {
	deck := newLobaDeck()
	assert.Len(t, deck, LobaDeckSize)
	assert.Equal(t, 108, LobaDeckSize)

	jokers, aceOfSpades := 0, 0
	for _, c := range deck {
		if c.GetDesign() == CardDesignJoker {
			jokers++
		}
		if c.GetDesign() == CardDesignSpade && c.GetValue() == 1 {
			aceOfSpades++
		}
	}
	assert.Equal(t, LobaJokerCnt, jokers)
	// **2 組ぶんなので同じ札が 2 枚ある。**単一デッキ前提で「その札は既に出た」と
	// 判断すると必ず食い違う。
	assert.Equal(t, 2, aceOfSpades, "every card appears twice")
}

func TestLoba_CardPointsCountTheJokerAsTen(t *testing.T) {
	assert.Equal(t, 10, LobaCardPoints(lbJoker()), "a joker is not free to hold")
	for _, v := range []int{1, 11, 12, 13} {
		assert.Equal(t, 10, LobaCardPoints(lbCard(CardDesignSpade, v)), "value %d", v)
	}
	assert.Equal(t, 7, LobaCardPoints(lbCard(CardDesignHeart, 7)))
	assert.Zero(t, LobaCardPoints(nil))
}

func TestLoba_APiernaNeedsThreeDifferentSuits(t *testing.T) {
	// #4414 が触れていない、Loba を Loba たらしめている規則。
	kind, err := LobaValidateMeld([]*Card{
		lbCard(CardDesignSpade, 7), lbCard(CardDesignHeart, 7), lbCard(CardDesignClover, 7),
	})
	require.NoError(t, err)
	assert.Equal(t, LobaMeldPierna, kind)

	// 2 組デッキなので同じスートが 2 枚揃うが、それではピエルナにならない。
	_, err = LobaValidateMeld([]*Card{
		lbCard(CardDesignSpade, 7), lbCard(CardDesignSpade, 7), lbCard(CardDesignHeart, 7),
	})
	assert.Error(t, err, "two of the same suit is only two suits")
}

func TestLoba_AJokerCannotJoinAPierna(t *testing.T) {
	// ジョーカーはワイルドではないので「同ランク」を満たしようがない。
	_, err := LobaValidateMeld([]*Card{
		lbCard(CardDesignSpade, 7), lbCard(CardDesignHeart, 7), lbJoker(),
	})
	assert.Error(t, err)
}

func TestLoba_AnEscaleraTakesAtMostOneJoker(t *testing.T) {
	kind, err := LobaValidateMeld([]*Card{
		lbCard(CardDesignSpade, 5), lbJoker(), lbCard(CardDesignSpade, 7),
	})
	require.NoError(t, err)
	assert.Equal(t, LobaMeldEscalera, kind)

	_, err = LobaValidateMeld([]*Card{
		lbCard(CardDesignSpade, 5), lbJoker(), lbJoker(), lbCard(CardDesignSpade, 8),
	})
	assert.Error(t, err, "two jokers in one escalera")
}

func TestLoba_AnEscaleraIsOneSuitInSequence(t *testing.T) {
	_, err := LobaValidateMeld([]*Card{
		lbCard(CardDesignSpade, 5), lbCard(CardDesignHeart, 6), lbCard(CardDesignSpade, 7),
	})
	assert.Error(t, err, "mixed suits")

	_, err = LobaValidateMeld([]*Card{
		lbCard(CardDesignSpade, 5), lbCard(CardDesignSpade, 7), lbCard(CardDesignSpade, 9),
	})
	assert.Error(t, err, "gaps with no joker to fill them")

	// 2 組デッキなので同じ札が 2 枚来ることがあるが、並びには使えない。
	_, err = LobaValidateMeld([]*Card{
		lbCard(CardDesignSpade, 5), lbCard(CardDesignSpade, 5), lbCard(CardDesignSpade, 6),
	})
	assert.Error(t, err, "a duplicate cannot fill its own place")
}

func TestLoba_TheAceIsHighOrLowButNeverBoth(t *testing.T) {
	_, err := LobaValidateMeld([]*Card{
		lbCard(CardDesignSpade, 1), lbCard(CardDesignSpade, 2), lbCard(CardDesignSpade, 3),
	})
	assert.NoError(t, err, "ace low")

	_, err = LobaValidateMeld([]*Card{
		lbCard(CardDesignSpade, 12), lbCard(CardDesignSpade, 13), lbCard(CardDesignSpade, 1),
	})
	assert.NoError(t, err, "ace high")

	// K-A-2 は跨ぎなので通らない。
	_, err = LobaValidateMeld([]*Card{
		lbCard(CardDesignSpade, 13), lbCard(CardDesignSpade, 1), lbCard(CardDesignSpade, 2),
	})
	assert.Error(t, err, "the ace cannot be both ends at once")
}

func TestLoba_AMeldNeedsThreeCards(t *testing.T) {
	_, err := LobaValidateMeld([]*Card{lbCard(CardDesignSpade, 7), lbCard(CardDesignHeart, 7)})
	assert.ErrorContains(t, err, "at least 3")
	_, err = LobaValidateMeld([]*Card{lbCard(CardDesignSpade, 7), lbCard(CardDesignHeart, 7), nil})
	assert.Error(t, err)
}

// lbFirstDiscardable returns a hand index that may legally be discarded.
func lbFirstDiscardable(t *testing.T, l *Loba, seat int) int {
	t.Helper()
	p := l.GetPlayer(seat)
	for i := range p.GetCardsSize() {
		if !lobaIsJoker(p.GetCard(i)) {
			return i
		}
	}
	t.Fatalf("seat %d holds only jokers", seat)
	return -1
}

// lbReady puts a game into the act step for the given seat.
func lbReady(t *testing.T, seat int) *Loba {
	t.Helper()
	l := NewDefaultLoba()
	l.Reset()
	l.SetCurrentPlayerForTest(seat)
	l.SetPhaseForTest(LobaPhaseAct)
	return l
}

func TestLoba_LayingOffToAPiernaStaysInTheOriginalThreeSuits(t *testing.T) {
	// 4 つ目のスートは「異なる 3 スート」の枠を壊す。2 組デッキなので、元の
	// 3 スートの 2 枚目なら付けられる。
	l := lbReady(t, 0)
	p := l.GetPlayer(0)
	p.Reset()
	p.AddCard(lbCard(CardDesignSpade, 7))
	p.AddCard(lbCard(CardDesignHeart, 7))
	p.AddCard(lbCard(CardDesignClover, 7))
	p.AddCard(lbCard(CardDesignDiamond, 7)) // 4 つ目のスート
	p.AddCard(lbCard(CardDesignSpade, 7))   // 元のスートの 2 枚目
	p.AddCard(lbCard(CardDesignHeart, 2))

	require.NoError(t, l.Meld(0, []int{0, 1, 2}))
	require.Len(t, l.GetMelds(), 1)

	// 手札は [♦7, ♠7, ♥2] になっている。
	assert.ErrorContains(t, l.LayOff(0, 0, 0), "does not fit", "the fourth suit is refused")
	require.NoError(t, l.LayOff(0, 1, 0), "a second card of an original suit fits")
}

func TestLoba_LayingOffNeedsAMeldOfYourOwnFirst(t *testing.T) {
	l := lbReady(t, 0)
	p := l.GetPlayer(0)
	p.Reset()
	p.AddCard(lbCard(CardDesignSpade, 7))
	p.AddCard(lbCard(CardDesignHeart, 7))
	p.AddCard(lbCard(CardDesignClover, 7))
	require.NoError(t, l.Meld(0, []int{0, 1, 2}))

	// 席 1 はまだ何も出していないので付けられない。
	l.SetCurrentPlayerForTest(1)
	l.SetPhaseForTest(LobaPhaseAct)
	q := l.GetPlayer(1)
	q.Reset()
	q.AddCard(lbCard(CardDesignDiamond, 7))
	assert.ErrorContains(t, l.LayOff(1, 0, 0), "must put down a meld of your own first")

	// 出していれば、**他家のメルドにも**付けられる (de Menos の規則)。
	l.SetHasMeldedForTest(1, true)
	q.Reset()
	q.AddCard(lbCard(CardDesignSpade, 7))
	require.NoError(t, l.LayOff(1, 0, 0))
	assert.Equal(t, 0, l.GetMelds()[0].Owner, "the meld still belongs to seat 0")
}

func TestLoba_AJokerCannotBeDiscardedUnlessItIsTheLastCard(t *testing.T) {
	l := lbReady(t, 0)
	p := l.GetPlayer(0)
	p.Reset()
	p.AddCard(lbJoker())
	p.AddCard(lbCard(CardDesignSpade, 3))
	assert.ErrorContains(t, l.Discard(0, 0), "cannot be discarded")
	require.NoError(t, l.Discard(0, 1), "an ordinary card is fine")

	// 手札がジョーカー 1 枚だけなら捨てられる。
	l2 := lbReady(t, 0)
	q := l2.GetPlayer(0)
	q.Reset()
	q.AddCard(lbJoker())
	require.NoError(t, l2.Discard(0, 0))
}

func TestLoba_GoingOutInOneGoTakesTenOff(t *testing.T) {
	l := lbReady(t, 0)
	for i := range l.GetPlayers() {
		l.GetPlayer(i).Reset()
	}
	p := l.GetPlayer(0)
	p.AddCard(lbCard(CardDesignSpade, 7))
	p.AddCard(lbCard(CardDesignHeart, 7))
	p.AddCard(lbCard(CardDesignClover, 7))
	// 相手に 1 枚残す。
	l.GetPlayer(1).AddCard(lbCard(CardDesignSpade, 13))

	require.NoError(t, l.Meld(0, []int{0, 1, 2}))

	assert.True(t, l.IsRoundClean())
	assert.Equal(t, -LobaGoOutCleanBonus, l.GetScore(0), "going out in one go is worth -10")
	assert.Equal(t, 10, l.GetScore(1), "a king left in hand is ten")
	assert.Equal(t, 0, l.GetRoundWinner())
	assert.Equal(t, LobaPhaseRoundEnd, l.GetPhase())
}

// TestLoba_ThreeMeldsInOneTurnIsStillClean は、9 枚を 3+3+3 で一気に出す
// **最も普通の一気上がり**が -10 を取れることを確かめる。メルドの数で判定して
// いると、この形が丸ごと弾かれる。
func TestLoba_ThreeMeldsInOneTurnIsStillClean(t *testing.T) {
	l := lbReady(t, 0)
	for i := range l.GetPlayers() {
		l.GetPlayer(i).Reset()
	}
	p := l.GetPlayer(0)
	for _, c := range []*Card{
		lbCard(CardDesignSpade, 7), lbCard(CardDesignHeart, 7), lbCard(CardDesignClover, 7),
		lbCard(CardDesignDiamond, 4), lbCard(CardDesignDiamond, 5), lbCard(CardDesignDiamond, 6),
		lbCard(CardDesignClover, 9), lbCard(CardDesignClover, 10), lbCard(CardDesignClover, 11),
	} {
		p.AddCard(c)
	}
	l.GetPlayer(1).AddCard(lbCard(CardDesignSpade, 13))

	require.NoError(t, l.Meld(0, []int{0, 1, 2}))
	require.NoError(t, l.Meld(0, []int{0, 1, 2}))
	require.NoError(t, l.Meld(0, []int{0, 1, 2}))

	assert.True(t, l.IsRoundClean(), "3+3+3 laid down in one turn is a clean go-out")
	assert.Equal(t, -LobaGoOutCleanBonus, l.GetScore(0))
}

// TestLoba_MeldingOnAnEarlierTurnIsNotClean は逆方向 — 前の手番で出しておいて
// 残りをレイオフで捌いた人に -10 を与えないことを確かめる。所有メルドは 1 つ
// のままなので、数で判定していると通ってしまう。
func TestLoba_MeldingOnAnEarlierTurnIsNotClean(t *testing.T) {
	l := lbReady(t, 0)
	for i := range l.GetPlayers() {
		l.GetPlayer(i).Reset()
	}
	p := l.GetPlayer(0)
	for _, c := range []*Card{
		lbCard(CardDesignSpade, 7), lbCard(CardDesignHeart, 7), lbCard(CardDesignClover, 7),
		lbCard(CardDesignSpade, 7), // 2 組デッキなので同じ札が 2 枚ある
	} {
		p.AddCard(c)
	}
	l.GetPlayer(1).AddCard(lbCard(CardDesignSpade, 13))
	require.NoError(t, l.Meld(0, []int{0, 1, 2}), "this is an earlier turn")
	require.False(t, l.GetGameEndFlag())

	// 次の手番。引いた時点で「もう出している」が控えられる。
	l.SetPhaseForTest(LobaPhaseDraw)
	l.SetStockForTest([]*Card{lbCard(CardDesignHeart, 7)})
	require.NoError(t, l.DrawFromStock(0))
	require.NoError(t, l.LayOff(0, 0, 0))
	require.NoError(t, l.LayOff(0, 0, 0))

	assert.Equal(t, 0, l.GetRoundWinner())
	assert.False(t, l.IsRoundClean(), "the melding happened on an earlier turn")
	assert.Equal(t, 0, l.GetScore(0), "no bonus")
}

func TestLoba_ReachingAHundredAndOneEliminates(t *testing.T) {
	l := lbReady(t, 0)
	for i := range l.GetPlayers() {
		l.GetPlayer(i).Reset()
	}
	p := l.GetPlayer(0)
	p.AddCard(lbCard(CardDesignSpade, 7))
	p.AddCard(lbCard(CardDesignHeart, 7))
	p.AddCard(lbCard(CardDesignClover, 7))
	// 席 1 はあと 10 点で 101 に届く。
	l.SetScoreForTest(1, LobaKnockOut-10)
	l.GetPlayer(1).AddCard(lbCard(CardDesignSpade, 13))

	require.NoError(t, l.Meld(0, []int{0, 1, 2}))

	assert.True(t, l.IsEliminated(1))
	assert.False(t, l.IsEliminated(0))
	assert.False(t, l.GetGameEndFlag(), "three players are still in")
}

func TestLoba_TheLastPlayerStandingWins(t *testing.T) {
	l := lbReady(t, 0)
	for i := range l.GetPlayers() {
		l.GetPlayer(i).Reset()
	}
	p := l.GetPlayer(0)
	p.AddCard(lbCard(CardDesignSpade, 7))
	p.AddCard(lbCard(CardDesignHeart, 7))
	p.AddCard(lbCard(CardDesignClover, 7))
	for _, seat := range []int{1, 2, 3} {
		l.SetScoreForTest(seat, LobaKnockOut-10)
		l.GetPlayer(seat).AddCard(lbCard(CardDesignSpade, 13))
	}

	require.NoError(t, l.Meld(0, []int{0, 1, 2}))

	assert.True(t, l.GetGameEndFlag())
	assert.Equal(t, 0, l.GetWinnerIdx(), "the survivor wins, not the lowest score")
	assert.Equal(t, LobaPhaseGameEnd, l.GetPhase())
}

func TestLoba_TheTurnIsDrawThenAct(t *testing.T) {
	l := NewDefaultLoba()
	l.Reset()
	cur := l.GetCurrentPlayerIdx()

	assert.Equal(t, LobaPhaseDraw, l.GetPhase())
	assert.ErrorContains(t, l.Discard(cur, 0), "must draw first")

	before := l.GetPlayer(cur).GetCardsSize()
	require.NoError(t, l.DrawFromStock(cur))
	assert.Equal(t, before+1, l.GetPlayer(cur).GetCardsSize())
	assert.Equal(t, LobaPhaseAct, l.GetPhase())
	assert.ErrorContains(t, l.DrawFromStock(cur), "not the draw step")

	// 添字 0 を決め打ちで捨てるとジョーカーを引いた回で落ちる。ジョーカーは
	// 捨てられないので、捨てられる札を選ぶ。
	require.NoError(t, l.Discard(cur, lbFirstDiscardable(t, l, cur)))
	assert.Equal(t, LobaPhaseDraw, l.GetPhase(), "the turn passes")
	assert.NotEqual(t, cur, l.GetCurrentPlayerIdx())
}

func TestLoba_TakingTheDiscardRemovesItFromThePile(t *testing.T) {
	l := NewDefaultLoba()
	l.Reset()
	cur := l.GetCurrentPlayerIdx()
	top := l.GetDiscardTop()
	require.NotNil(t, top)

	require.NoError(t, l.DrawFromDiscard(cur))
	assert.NotEqual(t, top, l.GetDiscardTop(), "the pile must shrink")

	// 空の捨て札からは取れない。
	l.SetPhaseForTest(LobaPhaseDraw)
	l.SetDiscardForTest(nil)
	assert.ErrorContains(t, l.DrawFromDiscard(cur), "discard pile is empty")
}

func TestLoba_AnEmptyStockIsRefilledFromTheDiscard(t *testing.T) {
	l := NewDefaultLoba()
	l.Reset()
	cur := l.GetCurrentPlayerIdx()
	l.SetStockForTest(nil)
	l.SetDiscardForTest([]*Card{
		lbCard(CardDesignSpade, 2), lbCard(CardDesignHeart, 3), lbCard(CardDesignClover, 4),
	})

	require.NoError(t, l.DrawFromStock(cur))
	assert.Positive(t, l.GetStockCount(), "the pile was turned back into the stock")
	assert.NotNil(t, l.GetDiscardTop(), "but the top card stays face up")
}

func TestLoba_RejectsIllegalRequests(t *testing.T) {
	l := NewDefaultLoba()
	l.Reset()
	cur := l.GetCurrentPlayerIdx()

	assert.Error(t, l.DrawFromStock((cur+1)%LobaPlayerCnt), "not that player's turn")
	assert.ErrorContains(t, l.NextRound(), "still in progress")

	require.NoError(t, l.DrawFromStock(cur))
	assert.Error(t, l.Discard(cur, -1))
	assert.Error(t, l.Discard(cur, 99))
	assert.Error(t, l.Meld(cur, []int{0, 0, 1}), "an index listed twice")
	assert.Error(t, l.Meld(cur, []int{0, 1, 99}))
	assert.Error(t, l.LayOff(cur, 0, 99), "no such meld")
}

// lbPlayRound drives one round with CPU decisions. Returns false if it stalls.
func lbPlayRound(t *testing.T, l *Loba) bool {
	t.Helper()
	for range 2000 {
		if l.GetPhase() == LobaPhaseRoundEnd || l.GetPhase() == LobaPhaseGameEnd {
			return true
		}
		idx := l.GetCurrentPlayerIdx()
		action := l.LobaCpuDecide(idx)
		switch l.GetPhase() {
		case LobaPhaseDraw:
			require.NoError(t, l.DrawFromStock(idx))
		case LobaPhaseAct:
			if action.MeldIdxs != nil {
				require.NoError(t, l.Meld(idx, action.MeldIdxs))
				continue
			}
			if action.DiscardIdx < 0 {
				return false
			}
			require.NoError(t, l.Discard(idx, action.DiscardIdx))
		}
	}
	return false
}

func TestLoba_ARoundPlaysOutToSomebodyGoingOut(t *testing.T) {
	l := NewDefaultLoba()
	l.Reset()
	require.True(t, lbPlayRound(t, l))
	assert.GreaterOrEqual(t, l.GetRoundWinner(), 0)
	assert.Zero(t, l.GetPlayer(l.GetRoundWinner()).GetCardsSize())
	assert.Equal(t, 1, l.GetRoundNumber())
}

func TestLoba_TheGameRunsUntilOnePlayerIsLeft(t *testing.T) {
	l := NewDefaultLoba()
	l.Reset()
	for range 200 {
		if l.GetGameEndFlag() {
			break
		}
		require.True(t, lbPlayRound(t, l))
		if l.GetGameEndFlag() {
			break
		}
		require.NoError(t, l.NextRound())
	}
	require.True(t, l.GetGameEndFlag())

	alive := 0
	for i := range l.GetPlayers() {
		if !l.IsEliminated(i) {
			alive++
		}
	}
	assert.LessOrEqual(t, alive, 1, "everyone else reached 101")
	assert.GreaterOrEqual(t, l.GetWinnerIdx(), 0)
}

func TestLoba_SurvivesAKVRoundTrip(t *testing.T) {
	l := NewDefaultLoba()
	l.Reset()
	require.True(t, lbPlayRound(t, l))

	data, err := json.Marshal(l)
	require.NoError(t, err)

	restored := NewDefaultLoba()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, l.GetPhase(), restored.GetPhase())
	assert.Equal(t, l.GetStockCount(), restored.GetStockCount())
	assert.Equal(t, l.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, len(l.GetMelds()), len(restored.GetMelds()))
	for i := range l.GetPlayers() {
		assert.Equal(t, l.GetScore(i), restored.GetScore(i), "score %d", i)
		// これが落ちると脱落者が復活する。
		assert.Equal(t, l.IsEliminated(i), restored.IsEliminated(i), "eliminated %d", i)
		assert.Equal(t, l.HasMelded(i), restored.HasMelded(i), "hasMelded %d", i)
	}
}

func TestLoba_UnmarshalRejectsAndClampsHostileSnapshots(t *testing.T) {
	for name, payload := range map[string]string{
		"not json":      "{",
		"seat count":    `{"pl":[],"cfg":{"cd":0},"ph":0}`,
		"bad config":    `{"pl":[{},{},{},{}],"cfg":{"cd":99},"ph":0}`,
		"unknown phase": `{"pl":[{},{},{},{}],"cfg":{"cd":0},"ph":9}`,
	} {
		t.Run(name, func(t *testing.T) {
			assert.Error(t, json.Unmarshal([]byte(payload), NewDefaultLoba()))
		})
	}

	short := `{"pl":[{},{},{},{}],"cfg":{"cd":0},"ph":0,"sc":[7],"el":[true],"cur":99,"rw":42,"wi":42,
	"me":[{"Owner":99,"Kind":0,"Cards":[{},{},{}]},{"Owner":0,"Kind":0,"Cards":[{}]}]}`
	l := NewDefaultLoba()
	require.NoError(t, json.Unmarshal([]byte(short), l))
	assert.Equal(t, 0, l.GetCurrentPlayerIdx(), "an out-of-range seat is clamped")
	assert.Equal(t, -1, l.GetRoundWinner())
	assert.Equal(t, -1, l.GetWinnerIdx())
	assert.Equal(t, 7, l.GetScore(0))
	assert.Zero(t, l.GetScore(3), "the padded tail must not read past the supplied slice")
	assert.True(t, l.IsEliminated(0))
	assert.False(t, l.IsEliminated(3))
	// 席番号が範囲外のメルドも、3 枚に満たないメルドも捨てる。
	assert.Empty(t, l.GetMelds())
}
