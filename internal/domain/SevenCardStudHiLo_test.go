//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func hlCard(design, value int) *Card { return NewCard(design, value, true) }

// hlDeal gives seat idx exactly seven cards (3 down, 4 up), which is what a
// finished stud hand looks like.
func hlDeal(s *SevenCardStud, idx int, cards ...*Card) {
	p := s.players[idx]
	p.ClearCards()
	for i, c := range cards {
		if i < 3 {
			p.AddHoleCard(c)
			continue
		}
		p.AddDoorCard(c)
	}
}

// hlTable builds a 4-seat Hi-Lo game with an even pot and no side pots.
func hlTable(t *testing.T, pot int) *SevenCardStud {
	t.Helper()
	s := NewDefaultSevenCardStudHiLo()
	require.True(t, s.GetIsHiLo())
	for _, p := range s.players {
		p.SetChips(0)
		p.SetFolded(false)
		p.SetCurrentBet(pot / len(s.players))
	}
	s.SetPot(pot)
	starting := make([]int, len(s.players))
	for i := range starting {
		starting[i] = pot / len(s.players)
	}
	s.SetStartingChips(starting)
	return s
}

func hlWon(s *SevenCardStud, idx int) int {
	for _, r := range s.GetRoundResults() {
		if r.PlayerIdx == idx {
			return r.WonAmount
		}
	}
	return 0
}

func TestSevenCardStudHiLo_TheLowMustBeEightOrBetterWithNoPair(t *testing.T) {
	p := NewSevenCardStudPlayer(false, HoldemStyleTAG)

	// A-2-3-4-5 (the wheel) with two high cards alongside: the low is there.
	for _, c := range []*Card{
		hlCard(CardDesignSpade, 1), hlCard(CardDesignHeart, 2), hlCard(CardDesignClover, 3),
		hlCard(CardDesignDiamond, 4), hlCard(CardDesignSpade, 5),
		hlCard(CardDesignHeart, 13), hlCard(CardDesignClover, 12),
	} {
		p.AddHoleCard(c)
	}
	require.True(t, p.EvalBestLowHandEightOrBetter())
	assert.Len(t, p.GetLowBestHand(), 5)

	// Nothing under nine: no qualifying low at all.
	nine := NewSevenCardStudPlayer(false, HoldemStyleTAG)
	for _, c := range []*Card{
		hlCard(CardDesignSpade, 9), hlCard(CardDesignHeart, 10), hlCard(CardDesignClover, 11),
		hlCard(CardDesignDiamond, 12), hlCard(CardDesignSpade, 13),
		hlCard(CardDesignHeart, 9), hlCard(CardDesignClover, 10),
	} {
		nine.AddHoleCard(c)
	}
	assert.False(t, nine.EvalBestLowHandEightOrBetter(), "a nine is already too high")
	assert.Nil(t, nine.GetLowBestHand())

	// Low cards but not five DISTINCT ranks: a pair kills the low.
	paired := NewSevenCardStudPlayer(false, HoldemStyleTAG)
	for _, c := range []*Card{
		hlCard(CardDesignSpade, 2), hlCard(CardDesignHeart, 2), hlCard(CardDesignClover, 3),
		hlCard(CardDesignDiamond, 3), hlCard(CardDesignSpade, 4),
		hlCard(CardDesignHeart, 13), hlCard(CardDesignClover, 12),
	} {
		paired.AddHoleCard(c)
	}
	assert.False(t, paired.EvalBestLowHandEightOrBetter(), "only four distinct low ranks")
}

func TestSevenCardStudHiLo_TheLowIsPickedIndependentlyOfTheHigh(t *testing.T) {
	// 同じ 7 枚からハイとローを別々に選んでよい。ハイに使った札をローで使い回せる。
	p := NewSevenCardStudPlayer(false, HoldemStyleTAG)
	for _, c := range []*Card{
		hlCard(CardDesignSpade, 1), hlCard(CardDesignSpade, 2), hlCard(CardDesignSpade, 3),
		hlCard(CardDesignSpade, 4), hlCard(CardDesignSpade, 5),
		hlCard(CardDesignHeart, 13), hlCard(CardDesignClover, 13),
	} {
		p.AddHoleCard(c)
	}
	p.EvalBestHand()
	require.True(t, p.EvalBestLowHandEightOrBetter())

	// A-5 のスペードは同時にストレートフラッシュでもある。ローの判定は
	// ストレートもフラッシュも無視するので、両取りになる。
	assert.Equal(t, PokerHandStraightFlush, p.GetHandRank())
	assert.Len(t, p.GetLowBestHand(), 5)
}

func TestSevenCardStudHiLo_WithNoQualifyingLowTheHighTakesEverything(t *testing.T) {
	// 8 or Better の肝。ローを取りに行った人が空振りすると、ポットは丸ごと
	// ハイへ行く。半分が浮いたままになってはいけない。
	s := hlTable(t, 400)
	// 全員 9 以上しか持たない = qualifying なローが 1 人もいない。
	hlDeal(s, 0, hlCard(CardDesignSpade, 1), hlCard(CardDesignHeart, 1), hlCard(CardDesignClover, 13),
		hlCard(CardDesignDiamond, 12), hlCard(CardDesignSpade, 11), hlCard(CardDesignHeart, 10), hlCard(CardDesignClover, 9))
	for i := 1; i < len(s.players); i++ {
		hlDeal(s, i, hlCard(CardDesignSpade, 9), hlCard(CardDesignHeart, 10), hlCard(CardDesignClover, 11),
			hlCard(CardDesignDiamond, 12), hlCard(CardDesignSpade, 13), hlCard(CardDesignHeart, 9+i), hlCard(CardDesignClover, 10))
	}

	s.resolveShowdown()

	total := 0
	for i := range s.players {
		total += hlWon(s, i)
	}
	assert.Equal(t, 400, total, "the whole pot must be paid out")
	for _, r := range s.GetRoundResults() {
		assert.False(t, r.LowQualifies, "seat %d should have no low", r.PlayerIdx)
		assert.Zero(t, r.WonLow)
	}
}

func TestSevenCardStudHiLo_AQualifyingLowTakesHalfAndTheOddChipGoesHigh(t *testing.T) {
	s := hlTable(t, 401)
	// 席 0: エースのペア (ハイ勝ち)、ローは無い。
	hlDeal(s, 0, hlCard(CardDesignSpade, 1), hlCard(CardDesignHeart, 1), hlCard(CardDesignClover, 13),
		hlCard(CardDesignDiamond, 12), hlCard(CardDesignSpade, 11), hlCard(CardDesignHeart, 10), hlCard(CardDesignClover, 9))
	// 席 1: 2-3-4-5-7 の qualifying なロー。ハイは弱い。
	hlDeal(s, 1, hlCard(CardDesignSpade, 2), hlCard(CardDesignHeart, 3), hlCard(CardDesignClover, 4),
		hlCard(CardDesignDiamond, 5), hlCard(CardDesignSpade, 7), hlCard(CardDesignHeart, 12), hlCard(CardDesignClover, 13))
	// 残りは降りている。
	for i := 2; i < len(s.players); i++ {
		s.players[i].SetFolded(true)
	}

	s.resolveShowdown()

	hi, lo := hlWon(s, 0), hlWon(s, 1)
	assert.Equal(t, 401, hi+lo, "the pot must balance")
	assert.Equal(t, 201, hi, "the odd chip goes to the high side")
	assert.Equal(t, 200, lo)

	for _, r := range s.GetRoundResults() {
		if r.PlayerIdx == 1 {
			assert.True(t, r.LowQualifies)
			assert.Equal(t, 200, r.WonLow)
		}
	}
}

func TestSevenCardStudHiLo_OnePlayerCanScoopBothHalves(t *testing.T) {
	s := hlTable(t, 400)
	// 席 0: A-2-3-4-5 のホイール = ストレートでもあり、ローも成立する。
	hlDeal(s, 0, hlCard(CardDesignSpade, 1), hlCard(CardDesignHeart, 2), hlCard(CardDesignClover, 3),
		hlCard(CardDesignDiamond, 4), hlCard(CardDesignSpade, 5), hlCard(CardDesignHeart, 13), hlCard(CardDesignClover, 12))
	// 席 1: K と 9 のツーペア。ストレートには負け、8 以下は 2 の 1 枚だけなので
	// ローも成立しない。9-10-J-Q-K を持たせると**ストレートになってハイを取って
	// しまう**ので、そこは避ける。
	hlDeal(s, 1, hlCard(CardDesignSpade, 9), hlCard(CardDesignHeart, 10), hlCard(CardDesignClover, 11),
		hlCard(CardDesignDiamond, 13), hlCard(CardDesignSpade, 13), hlCard(CardDesignHeart, 9), hlCard(CardDesignClover, 2))
	for i := 2; i < len(s.players); i++ {
		s.players[i].SetFolded(true)
	}

	s.resolveShowdown()

	assert.Equal(t, 400, hlWon(s, 0), "a scoop takes both halves")
	assert.Zero(t, hlWon(s, 1))
}

func TestSevenCardStudHiLo_TheLowerLowWinsTheLowHalf(t *testing.T) {
	s := hlTable(t, 400)
	// 席 0 はハイ担当 (ローなし)。
	hlDeal(s, 0, hlCard(CardDesignSpade, 1), hlCard(CardDesignHeart, 1), hlCard(CardDesignClover, 1),
		hlCard(CardDesignDiamond, 12), hlCard(CardDesignSpade, 11), hlCard(CardDesignHeart, 10), hlCard(CardDesignClover, 9))
	// 席 1: 2-3-4-5-8。
	hlDeal(s, 1, hlCard(CardDesignSpade, 2), hlCard(CardDesignHeart, 3), hlCard(CardDesignClover, 4),
		hlCard(CardDesignDiamond, 5), hlCard(CardDesignSpade, 8), hlCard(CardDesignHeart, 13), hlCard(CardDesignClover, 12))
	// 席 2: A-2-3-4-6 -- こちらが低い。
	hlDeal(s, 2, hlCard(CardDesignSpade, 1), hlCard(CardDesignHeart, 2), hlCard(CardDesignClover, 3),
		hlCard(CardDesignDiamond, 4), hlCard(CardDesignSpade, 6), hlCard(CardDesignHeart, 13), hlCard(CardDesignClover, 11))
	s.players[3].SetFolded(true)

	s.resolveShowdown()

	assert.Zero(t, hlWon(s, 1), "the higher low wins nothing")
	assert.Equal(t, 200, hlWon(s, 2), "the lower low takes the low half")
}

func TestSevenCardStudHiLo_PlainStudAndRazzAreUnaffected(t *testing.T) {
	// Hi-Lo は既存の 2 モードに触らない。フラグを足したことで通常の
	// スタッドがスプリットになっては困る。
	assert.False(t, NewDefaultSevenCardStud().GetIsHiLo())
	assert.False(t, NewDefaultRazz().GetIsHiLo())
	assert.True(t, NewDefaultRazz().GetIsLowball())
	assert.False(t, NewDefaultSevenCardStudHiLo().GetIsLowball(), "Hi-Lo is not lowball")
}

func TestSevenCardStudHiLo_SurvivesAKVRoundTrip(t *testing.T) {
	s := NewDefaultSevenCardStudHiLo()
	require.NoError(t, s.Reset())
	hlDeal(s, 0, hlCard(CardDesignSpade, 1), hlCard(CardDesignHeart, 2), hlCard(CardDesignClover, 3),
		hlCard(CardDesignDiamond, 4), hlCard(CardDesignSpade, 5), hlCard(CardDesignHeart, 13), hlCard(CardDesignClover, 12))
	require.True(t, s.players[0].EvalBestLowHandEightOrBetter())

	data, err := json.Marshal(s)
	require.NoError(t, err)

	restored := NewDefaultSevenCardStud()
	require.NoError(t, json.Unmarshal(data, restored))

	// フラグが落ちると Worker 側でハイの総取りになり、静かに払い出しが変わる。
	assert.True(t, restored.GetIsHiLo(), "the hi-lo flag must survive")
	assert.True(t, restored.GetPlayers()[0].GetLowQualifies(), "and so must the low itself")
	assert.Len(t, restored.GetPlayers()[0].GetLowBestHand(), 5)
}
