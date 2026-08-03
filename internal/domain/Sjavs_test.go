//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sjCard(design, value int) *Card { return NewCard(design, value, true) }

func TestSjavs_TheDeckIsThirtyTwoCardsWorthOneHundredAndTwenty(t *testing.T) {
	deck := newSjavsDeck()
	assert.Len(t, deck, SjavsDeckSize)

	total := 0
	for _, c := range deck {
		assert.NotContains(t, []int{2, 3, 4, 5, 6}, c.GetValue(), "2-6 are removed")
		total += SjavsCardPoints(c)
	}
	// A11 + 10*10 + K4 + Q3 + J2 = 30 per suit, 120 in all. 精算表の分母なので、
	// ここがずれると 61/90 の閾値の意味も変わる。
	assert.Equal(t, SjavsTotalPoints, total)
	assert.Equal(t, 120, total)
}

func TestSjavs_SixPermanentTrumpsInTheStatedOrder(t *testing.T) {
	// #4403 は常時切札を ♣Q ＞ ♠Q の 2 枚とするが、実際は 6 枚ある。
	want := []*Card{
		sjCard(CardDesignClover, 12), sjCard(CardDesignSpade, 12),
		sjCard(CardDesignClover, 11), sjCard(CardDesignSpade, 11),
		sjCard(CardDesignHeart, 11), sjCard(CardDesignDiamond, 11),
	}
	assert.Len(t, sjavsPermanentTrumps, 6)
	for i, c := range want {
		assert.Equal(t, i, sjavsPermanentRank(c), "rank of card %d", i)
	}
	// 常時切札でない札は -1。
	assert.Equal(t, -1, sjavsPermanentRank(sjCard(CardDesignHeart, 12)))
	assert.Equal(t, -1, sjavsPermanentRank(nil))

	// 強い順に並んでいること。
	for i := 1; i < len(want); i++ {
		assert.True(t, sjavsTrumpStrength(want[i-1]) > sjavsTrumpStrength(want[i]),
			"permanent trump %d must beat %d", i-1, i)
	}
}

func TestSjavs_TrumpCountIsThirteenForRedAndTwelveForBlack(t *testing.T) {
	// これが「常時切札は 6 枚」の検算。2 枚だと 9/8 にしかならず、原典の
	// 13/12 と合わない。
	assert.Equal(t, 13, SjavsTrumpCount(CardDesignHeart))
	assert.Equal(t, 13, SjavsTrumpCount(CardDesignDiamond))
	assert.Equal(t, 12, SjavsTrumpCount(CardDesignSpade), "the black suit's own Q and J are already permanent")
	assert.Equal(t, 12, SjavsTrumpCount(CardDesignClover))
}

func TestSjavs_PermanentTrumpsOutrankTheTrumpSuitsOwnCards(t *testing.T) {
	// ♥が切札でも ♦J (最弱の常時切札) が ♥A に勝つ。
	assert.True(t, sjavsBeats(sjCard(CardDesignDiamond, 11), sjCard(CardDesignHeart, 1), CardDesignHeart))
	// そして ♣Q が ♦J に勝つ。
	assert.True(t, sjavsBeats(sjCard(CardDesignClover, 12), sjCard(CardDesignDiamond, 11), CardDesignHeart))
	// 切札は平札に勝つ。
	assert.True(t, sjavsBeats(sjCard(CardDesignHeart, 7), sjCard(CardDesignSpade, 1), CardDesignHeart))
	// 追随していない平札は勝てない。
	assert.False(t, sjavsBeats(sjCard(CardDesignDiamond, 1), sjCard(CardDesignSpade, 7), CardDesignHeart))
}

func TestSjavs_TrumpsAreTheirOwnSuitWhenFollowing(t *testing.T) {
	// ♥が切札のとき ♣J は「♣」ではなく「切札」。♣がリードされても追随しない。
	s := NewDefaultSjavs()
	s.Reset()
	s.SetTrumpSuitForTest(CardDesignHeart)
	s.SetPhaseForTest(SjavsPhasePlay)

	p := s.GetPlayer(1)
	p.Reset()
	p.AddCard(sjCard(CardDesignClover, 11)) // 切札 (常時)
	p.AddCard(sjCard(CardDesignClover, 9))  // ただの♣
	p.AddCard(sjCard(CardDesignSpade, 8))

	s.trick = []SjavsTrickCard{{PlayerIdx: 0, Card: sjCard(CardDesignClover, 13)}}
	s.SetCurrentPlayerForTest(1)

	// ♣がリードされたら、追随できるのは「切札でない♣」だけ。
	assert.Equal(t, []int{1}, s.GetValidPlayIndices(1))

	// 逆に切札がリードされたら、常時切札も含めて切札で追随する。
	s.trick = []SjavsTrickCard{{PlayerIdx: 0, Card: sjCard(CardDesignHeart, 13)}}
	assert.Equal(t, []int{0}, s.GetValidPlayIndices(1), "the club jack IS a trump here")
}

func TestSjavs_ABidCountsEveryTrumpIncludingThePermanentOnes(t *testing.T) {
	// 素のスート枚数を数えると、実際に持っている切札より少なく申告してしまう。
	s := NewDefaultSjavs()
	s.Reset()
	p := s.GetPlayer(0)
	p.Reset()
	// ♥は 3 枚だが、常時切札 3 枚を足すと♥切札は 6 枚になる。
	p.AddCard(sjCard(CardDesignHeart, 1))
	p.AddCard(sjCard(CardDesignHeart, 13))
	p.AddCard(sjCard(CardDesignHeart, 10))
	p.AddCard(sjCard(CardDesignClover, 12))
	p.AddCard(sjCard(CardDesignSpade, 12))
	p.AddCard(sjCard(CardDesignDiamond, 11))
	p.AddCard(sjCard(CardDesignSpade, 7))
	p.AddCard(sjCard(CardDesignSpade, 8))

	assert.Equal(t, 6, s.LongestTrumpLength(0), "three hearts plus three permanent trumps")
}

func TestSjavs_BiddingRejectsWhatYouDoNotHold(t *testing.T) {
	s := NewDefaultSjavs()
	s.Reset()
	cur := s.GetCurrentPlayerIdx()

	assert.ErrorContains(t, s.Bid(cur, 4), "at least 5")
	assert.ErrorContains(t, s.Bid(cur, 99), "you only hold")
	assert.Error(t, s.Bid((cur+1)%SjavsPlayerCnt, 0), "not that player's turn")
}

func TestSjavs_LongerBeatsShorterAndClubsBeatsAnEqualLength(t *testing.T) {
	s := NewDefaultSjavs()
	s.Reset()

	// 立っているビッドが無ければ何でも通る。
	assert.True(t, s.beatsStandingBid(5, false))

	s.bidderIdx, s.bidLength, s.bidIsClubs = 0, 6, false
	assert.True(t, s.beatsStandingBid(7, false), "longer wins")
	assert.False(t, s.beatsStandingBid(5, false), "shorter loses")
	assert.False(t, s.beatsStandingBid(6, false), "equal length without clubs does not overcall")
	assert.True(t, s.beatsStandingBid(6, true), "clubs beats an equal length")

	s.bidIsClubs = true
	assert.False(t, s.beatsStandingBid(6, true), "but clubs does not beat clubs")
}

func TestSjavs_ClubsIsForcedWhenItTiesForTheLongest(t *testing.T) {
	// 同長では♣が勝つので、♣が最長候補にあるなら♣を宣言しなければならない。
	s := NewDefaultSjavs()
	s.Reset()
	p := s.GetPlayer(0)
	p.Reset()
	// ♠が 5、♣が 5 (常時切札込み) になる手を作る。
	for _, c := range []*Card{
		sjCard(CardDesignSpade, 1), sjCard(CardDesignSpade, 13), sjCard(CardDesignSpade, 10),
		sjCard(CardDesignClover, 1), sjCard(CardDesignClover, 13), sjCard(CardDesignClover, 10),
		sjCard(CardDesignSpade, 12), sjCard(CardDesignClover, 12),
	} {
		p.AddCard(c)
	}
	suits := s.SjavsLongestSuits(0)
	require.Contains(t, suits, CardDesignClover)

	s.bidderIdx = 0
	s.finishBidding()
	assert.Equal(t, CardDesignClover, s.GetTrumpSuit(), "clubs must be chosen when it ties")
}

// sjSettle pins a settlement scenario without playing the hand out.
func sjSettle(t *testing.T, trump, declPts, declTricks int) *SjavsHandResult {
	t.Helper()
	s := NewDefaultSjavs()
	s.Reset()
	s.SetTrumpSuitForTest(trump)
	s.SetBidderForTest(0) // team 0
	s.SetTeamPointsForTest(declPts, SjavsTotalPoints-declPts)
	won := make([]int, SjavsPlayerCnt)
	won[0] = declTricks
	won[1] = SjavsHandSize - declTricks
	s.SetTricksWonForTest(won)
	s.SettleHandForTest()
	return s.GetHandResult()
}

func TestSjavs_SettlementFollowsTheSourceTable(t *testing.T) {
	for name, tc := range map[string]struct {
		trump, pts, tricks int
		wantTeam, wantAmt  int
		wantVol            bool
	}{
		"vol":            {CardDesignHeart, 120, 8, 0, 12, true},
		"vol in clubs":   {CardDesignClover, 120, 8, 0, 16, true},
		"90 or more":     {CardDesignHeart, 95, 6, 0, 4, false},
		"90 in clubs":    {CardDesignClover, 95, 6, 0, 8, false},
		"61 to 89":       {CardDesignHeart, 70, 5, 0, 2, false},
		"61 in clubs":    {CardDesignClover, 70, 5, 0, 4, false},
		"31 to 59":       {CardDesignHeart, 45, 3, 1, 4, false},
		"31 to 59 clubs": {CardDesignClover, 45, 3, 1, 8, false},
		"under 31":       {CardDesignHeart, 20, 1, 1, 8, false},
		"under 31 clubs": {CardDesignClover, 20, 1, 1, 16, false},
	} {
		t.Run(name, func(t *testing.T) {
			res := sjSettle(t, tc.trump, tc.pts, tc.tricks)
			require.NotNil(t, res)
			assert.Equal(t, tc.wantTeam, res.ScoringTeam)
			assert.Equal(t, tc.wantAmt, res.Amount)
			assert.Equal(t, tc.wantVol, res.Vol)
		})
	}
}

func TestSjavs_NoTricksIsSixteenWhateverTheTrumpSuit(t *testing.T) {
	// ここだけ♣の倍額規則が効かない。倍にすると 32 になってしまう。
	for _, trump := range []int{CardDesignHeart, CardDesignClover} {
		res := sjSettle(t, trump, 0, 0)
		require.NotNil(t, res)
		assert.Equal(t, 1, res.ScoringTeam)
		assert.Equal(t, 16, res.Amount, "trump %d", trump)
	}
}

func TestSjavs_SixtySixtyScoresNothingButRaisesTheNextGame(t *testing.T) {
	s := NewDefaultSjavs()
	s.Reset()
	s.SetTrumpSuitForTest(CardDesignHeart)
	s.SetBidderForTest(0)
	s.SetTeamPointsForTest(60, 60)
	s.SetTricksWonForTest([]int{4, 4, 0, 0})
	s.SettleHandForTest()

	require.NotNil(t, s.GetHandResult())
	assert.Equal(t, -1, s.GetHandResult().ScoringTeam, "nobody scores")
	assert.Equal(t, SjavsRubber, s.GetRemaining(0), "and nothing is deducted")
	assert.Equal(t, SjavsRubber, s.GetRemaining(1))
	assert.Equal(t, 2, s.GetCarryOver(), "the next game is worth two more")

	// 次のハンドでその 2 が上乗せされる。
	require.NoError(t, s.NextHand())
	s.SetTrumpSuitForTest(CardDesignHeart)
	s.SetBidderForTest(0)
	s.SetTeamPointsForTest(70, 50)
	s.SetTricksWonForTest([]int{4, 4, 0, 0})
	s.SettleHandForTest()

	assert.Equal(t, 4, s.GetHandResult().Amount, "2 for the hand plus the 2 carried over")
	assert.Zero(t, s.GetCarryOver(), "and the carry-over is spent")
}

func TestSjavs_TheRubberCountsDownFromTwentyFour(t *testing.T) {
	// #4403 は「クロスを1つ消す/全部消したら勝ち」とするが、実際は 24 からの
	// 引き算で、クロスはラバー勝利数の記録。
	s := NewDefaultSjavs()
	s.Reset()
	assert.Equal(t, SjavsRubber, s.GetRemaining(0))
	assert.Equal(t, SjavsRubber, s.GetRemaining(1))
	assert.Zero(t, s.GetCrosses(0))

	s.SetTrumpSuitForTest(CardDesignHeart)
	s.SetBidderForTest(0)
	s.SetRemainingForTest(3, 24)
	s.SetTeamPointsForTest(95, 25)
	s.SetTricksWonForTest([]int{5, 3, 0, 0})
	s.SettleHandForTest()

	assert.Equal(t, -1, s.GetRemaining(0), "passing zero is enough")
	assert.True(t, s.GetGameEndFlag())
	assert.Equal(t, 0, s.GetWinnerTeam())
	assert.Equal(t, 1, s.GetCrosses(0), "a cross records the rubber, not the hand")
	assert.True(t, s.IsDoubleVictory(), "the losers never left 24")
}

func TestSjavs_ADoubleVictoryNeedsTheLosersStillOnTwentyFour(t *testing.T) {
	s := NewDefaultSjavs()
	s.Reset()
	s.SetTrumpSuitForTest(CardDesignHeart)
	s.SetBidderForTest(0)
	s.SetRemainingForTest(2, 20)
	s.SetTeamPointsForTest(95, 25)
	s.SetTricksWonForTest([]int{5, 3, 0, 0})
	s.SettleHandForTest()

	require.True(t, s.GetGameEndFlag())
	assert.False(t, s.IsDoubleVictory(), "the losers had already scored")
}

func TestSjavs_TeamsSitAcross(t *testing.T) {
	assert.Equal(t, SjavsTeamOf(0), SjavsTeamOf(2))
	assert.Equal(t, SjavsTeamOf(1), SjavsTeamOf(3))
	assert.NotEqual(t, SjavsTeamOf(0), SjavsTeamOf(1))
}

func TestSjavs_EveryDealCanBeBid(t *testing.T) {
	// 誰も 5 枚に届かなければ配り直す。届かない配りのまま進むと、
	// ビッドフェーズから抜けられない。
	for range 200 {
		s := NewDefaultSjavs()
		s.Reset()
		best := 0
		for i := range s.GetPlayers() {
			if n := s.LongestTrumpLength(i); n > best {
				best = n
			}
		}
		require.GreaterOrEqual(t, best, SjavsMinBid)
	}
}

// sjPlayHand drives one hand with CPU decisions. Returns false if it stalls.
func sjPlayHand(t *testing.T, s *Sjavs) bool {
	t.Helper()
	for range 200 {
		if s.GetPhase() == SjavsPhaseHandEnd || s.GetPhase() == SjavsPhaseGameEnd {
			return true
		}
		idx := s.GetCurrentPlayerIdx()
		action := s.SjavsCpuDecide(idx)
		if s.GetPhase() == SjavsPhaseBid {
			require.NoError(t, s.Bid(idx, action.BidLength))
			continue
		}
		if action.HandIdx < 0 {
			return false
		}
		require.NoError(t, s.PlayCard(idx, action.HandIdx))
	}
	return false
}

func TestSjavs_AHandPlaysOutAndTheHundredAndTwentyPointsAreAllAccountedFor(t *testing.T) {
	s := NewDefaultSjavs()
	s.Reset()
	require.True(t, sjPlayHand(t, s))

	assert.Equal(t, SjavsHandSize, s.GetTrickNumber())
	assert.Equal(t, SjavsTotalPoints, s.GetTeamPoints(0)+s.GetTeamPoints(1),
		"every point in the deck must land with one team or the other")
	for i := range s.GetPlayers() {
		assert.Zero(t, s.GetPlayer(i).GetCardsSize(), "seat %d still holds cards", i)
	}
}

func TestSjavs_RejectsIllegalRequests(t *testing.T) {
	s := NewDefaultSjavs()
	s.Reset()

	// ビッド中はプレイできない。
	assert.ErrorContains(t, s.PlayCard(s.GetCurrentPlayerIdx(), 0), "not the play phase")

	s.SetPhaseForTest(SjavsPhasePlay)
	s.SetTrumpSuitForTest(CardDesignHeart)
	cur := s.GetCurrentPlayerIdx()
	assert.Error(t, s.PlayCard(cur, -1))
	assert.Error(t, s.PlayCard(cur, 99))
	assert.Error(t, s.PlayCard((cur+1)%SjavsPlayerCnt, 0), "not that player's turn")

	// ハンド進行中に次のハンドへは進めない。
	assert.ErrorContains(t, s.NextHand(), "still in progress")
}

func TestSjavs_SurvivesAKVRoundTrip(t *testing.T) {
	s := NewDefaultSjavs()
	s.Reset()
	require.True(t, sjPlayHand(t, s))

	data, err := json.Marshal(s)
	require.NoError(t, err)

	restored := NewDefaultSjavs()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, s.GetPhase(), restored.GetPhase())
	assert.Equal(t, s.GetTrumpSuit(), restored.GetTrumpSuit())
	assert.Equal(t, s.GetBidderIdx(), restored.GetBidderIdx())
	assert.Equal(t, s.GetDealerIdx(), restored.GetDealerIdx())
	// ラバーの進行が巻き戻ると、決着済みの点数が消える。
	for team := range 2 {
		assert.Equal(t, s.GetRemaining(team), restored.GetRemaining(team), "remaining %d", team)
		assert.Equal(t, s.GetCrosses(team), restored.GetCrosses(team), "crosses %d", team)
		assert.Equal(t, s.GetTeamPoints(team), restored.GetTeamPoints(team), "points %d", team)
	}
	assert.Equal(t, s.GetCarryOver(), restored.GetCarryOver())
}

func TestSjavs_UnmarshalRejectsAndClampsHostileSnapshots(t *testing.T) {
	for name, payload := range map[string]string{
		"not json":      "{",
		"seat count":    `{"pl":[],"cfg":{"cd":0},"ph":0}`,
		"bad config":    `{"pl":[{},{},{},{}],"cfg":{"cd":99},"ph":0}`,
		"unknown phase": `{"pl":[{},{},{},{}],"cfg":{"cd":0},"ph":9}`,
	} {
		t.Run(name, func(t *testing.T) {
			assert.Error(t, json.Unmarshal([]byte(payload), NewDefaultSjavs()))
		})
	}

	// 欠けた remaining は 0 ではなく 24 で埋める。0 だと復元しただけで
	// ラバー勝ちが成立してしまう。
	short := `{"pl":[{},{},{},{}],"cfg":{"cd":0},"ph":0,"rm":[5],"cr":[],"bi":42,"ts":99,"cur":99,"wt":7}`
	s := NewDefaultSjavs()
	require.NoError(t, json.Unmarshal([]byte(short), s))
	assert.Equal(t, 5, s.GetRemaining(0))
	assert.Equal(t, SjavsRubber, s.GetRemaining(1), "the missing side starts full, not at zero")
	assert.Equal(t, -1, s.GetBidderIdx(), "an out-of-range seat is not trusted")
	assert.Equal(t, -1, s.GetTrumpSuit())
	assert.Equal(t, 0, s.GetCurrentPlayerIdx())
	assert.Equal(t, -1, s.GetWinnerTeam())
}
