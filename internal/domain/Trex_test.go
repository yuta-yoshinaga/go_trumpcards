//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func txCard(design, value int) *Card { return NewCard(design, value, true) }

func TestTrex_ThePenaltyTableBalancesAgainstTheDominoBonuses(t *testing.T) {
	// これが表の裏取り。1 王国 (5 ディール) はちょうどゼロサムになる。
	// どれか 1 つでも値が違えばこの一致は起きない。
	penalties := -TrexKingOfHeartsPenalty + // 75  … ♥K 1 枚
		-TrexDiamondPenalty*13 + // 130 … ♦ 13 枚
		-TrexQueenPenalty*4 + // 100 … Q 4 枚
		-TrexTrickPenalty*TrexHandSize // 195 … 13 トリック
	assert.Equal(t, 500, penalties)

	bonuses := 0
	for _, b := range TrexTrixBonuses {
		bonuses += b
	}
	assert.Equal(t, 500, bonuses)
	assert.Equal(t, penalties, bonuses, "one kingdom is exactly zero-sum")
	assert.Equal(t, []int{200, 150, 100, 50}, TrexTrixBonuses)
}

func TestTrex_TwentyDealsInFourKingdoms(t *testing.T) {
	assert.Equal(t, 20, TrexTotalDeals)
	assert.Equal(t, TrexPlayerCnt*TrexContractsPerKingdom, TrexTotalDeals)
}

func TestTrex_TheFiveContractsAreTheSourcesNotTheIssues(t *testing.T) {
	// #4410 は「ノーハート」「ノーキングス」とするが、実際はダイヤと ♥K 1 枚。
	tr := NewDefaultTrex()
	tr.Reset()

	tr.SetContractForTest(TrexContractDiamonds)
	assert.Equal(t, TrexDiamondPenalty, tr.cardPenalty(txCard(CardDesignDiamond, 3)))
	assert.Zero(t, tr.cardPenalty(txCard(CardDesignHeart, 3)), "hearts are not the diamond contract")

	tr.SetContractForTest(TrexContractKingOfHearts)
	assert.Equal(t, TrexKingOfHeartsPenalty, tr.cardPenalty(txCard(CardDesignHeart, 13)))
	// 「キング」複数ではなく ♥K の 1 枚だけ。
	assert.Zero(t, tr.cardPenalty(txCard(CardDesignSpade, 13)), "only the HEART king costs anything")
	assert.Zero(t, tr.cardPenalty(txCard(CardDesignClover, 13)))

	tr.SetContractForTest(TrexContractQueens)
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		assert.Equal(t, TrexQueenPenalty, tr.cardPenalty(txCard(suit, 12)), "suit %d", suit)
	}

	// トリック契約は札ではなくトリックに付く。
	tr.SetContractForTest(TrexContractTricks)
	assert.Zero(t, tr.cardPenalty(txCard(CardDesignHeart, 13)))
}

func TestTrex_TheDominoesStartFromTheJackNotTheSeven(t *testing.T) {
	// #4410 は「7 を起点に数字順へ連結」とするが、起点は J。7 起点にすると
	// BarbuDominoes と同じ 7 並べになり、別のゲームになる。
	tr := NewDefaultTrex()
	tr.Reset()
	tr.SetContractForTest(TrexContractTrix)
	tr.SetPhaseForTest(TrexPhasePlay)

	assert.True(t, tr.trixPlayable(txCard(CardDesignSpade, 11)), "a jack opens its suit")
	assert.False(t, tr.trixPlayable(txCard(CardDesignSpade, 7)), "a seven does not")
	assert.False(t, tr.trixPlayable(txCard(CardDesignSpade, 12)), "nor does a queen before the jack")

	// J を置くと、その両隣が開く。
	tr.runs[CardDesignSpade] = trexSuitRun{Started: true, Low: 11, High: 11}
	assert.True(t, tr.trixPlayable(txCard(CardDesignSpade, 10)))
	assert.True(t, tr.trixPlayable(txCard(CardDesignSpade, 12)))
	assert.False(t, tr.trixPlayable(txCard(CardDesignSpade, 9)), "not two away")
	// 別スートは別の列。♠の J は♥を開かない。
	assert.False(t, tr.trixPlayable(txCard(CardDesignHeart, 12)))
	assert.True(t, tr.trixPlayable(txCard(CardDesignHeart, 11)))
}

func TestTrex_TheAceIsTheTopOfTheDominoRun(t *testing.T) {
	// J,Q,K,A と伸びる。A を 1 として扱うと K の次が繋がらない。
	assert.Equal(t, 14, TrexRank(txCard(CardDesignSpade, 1)))
	assert.Equal(t, 13, TrexRank(txCard(CardDesignSpade, 13)))
	assert.Equal(t, 2, TrexRank(txCard(CardDesignSpade, 2)))
	assert.Zero(t, TrexRank(nil))

	tr := NewDefaultTrex()
	tr.Reset()
	tr.SetContractForTest(TrexContractTrix)
	tr.SetPhaseForTest(TrexPhasePlay)
	tr.runs[CardDesignSpade] = trexSuitRun{Started: true, Low: 11, High: 13}
	assert.True(t, tr.trixPlayable(txCard(CardDesignSpade, 1)), "the ace follows the king")
}

func TestTrex_TheSevenOfHeartsPicksTheFirstKing(t *testing.T) {
	// 任意の席から始めると、王国が移る順序の起点が決まらない。
	for range 50 {
		tr := NewDefaultTrex()
		tr.Reset()
		assert.Equal(t, tr.holderOf(CardDesignHeart, 7), tr.GetKingIdx())
	}
}

func TestTrex_OnlyTheKingChoosesAndNeverTheSameContractTwice(t *testing.T) {
	tr := NewDefaultTrex()
	tr.Reset()
	king := tr.GetKingIdx()

	assert.Error(t, tr.ChooseContract((king+1)%TrexPlayerCnt, TrexContractQueens), "only the king chooses")
	assert.Error(t, tr.ChooseContract(king, TrexContractNone), "None is not selectable")
	assert.Error(t, tr.ChooseContract(king, TrexContract(99)))

	require.Len(t, tr.AvailableContracts(), TrexContractsPerKingdom)
	require.NoError(t, tr.ChooseContract(king, TrexContractQueens))
	assert.True(t, tr.IsContractUsed(king, TrexContractQueens))
	assert.Len(t, tr.AvailableContracts(), TrexContractsPerKingdom-1)

	// 同じ王が同じ契約を二度は選べない。
	tr.SetPhaseForTest(TrexPhaseChoose)
	assert.ErrorContains(t, tr.ChooseContract(king, TrexContractQueens), "already been played")
}

func TestTrex_TrickPenaltiesLandOnTheWinner(t *testing.T) {
	tr := NewDefaultTrex()
	tr.Reset()
	tr.SetKingForTest(0)
	tr.SetPhaseForTest(TrexPhasePlay)
	tr.SetCurrentPlayerForTest(0)

	// クイーン契約: Q を 2 枚含むトリックは -50。
	tr.SetContractForTest(TrexContractQueens)
	tr.trick = []TrexTrickCard{
		{PlayerIdx: 0, Card: txCard(CardDesignSpade, 1)},
		{PlayerIdx: 1, Card: txCard(CardDesignSpade, 12)},
		{PlayerIdx: 2, Card: txCard(CardDesignHeart, 12)},
		{PlayerIdx: 3, Card: txCard(CardDesignSpade, 2)},
	}
	tr.resolveTrick()
	assert.Equal(t, 2*TrexQueenPenalty, tr.GetDealScore(0), "the ace took both queens")
	assert.Zero(t, tr.GetDealScore(1))
}

func TestTrex_TheTricksContractChargesPerTrickNotPerCard(t *testing.T) {
	tr := NewDefaultTrex()
	tr.Reset()
	tr.SetPhaseForTest(TrexPhasePlay)
	tr.SetContractForTest(TrexContractTricks)
	tr.trick = []TrexTrickCard{
		{PlayerIdx: 0, Card: txCard(CardDesignSpade, 1)},
		{PlayerIdx: 1, Card: txCard(CardDesignSpade, 12)},
		{PlayerIdx: 2, Card: txCard(CardDesignSpade, 13)},
		{PlayerIdx: 3, Card: txCard(CardDesignSpade, 2)},
	}
	tr.resolveTrick()
	assert.Equal(t, TrexTrickPenalty, tr.GetDealScore(0), "one trick, one charge")
}

func TestTrex_TrickWinnerIgnoresOffSuitCards(t *testing.T) {
	// 切札は無い。リードのスートに追随していない札は、いくら強くても勝てない。
	trick := []TrexTrickCard{
		{PlayerIdx: 0, Card: txCard(CardDesignSpade, 5)},
		{PlayerIdx: 1, Card: txCard(CardDesignHeart, 1)},
		{PlayerIdx: 2, Card: txCard(CardDesignSpade, 7)},
		{PlayerIdx: 3, Card: txCard(CardDesignDiamond, 13)},
	}
	assert.Equal(t, 2, TrexTrickWinner(trick))
	assert.Equal(t, -1, TrexTrickWinner(nil))
}

func TestTrex_FollowingSuitIsForcedWhenPossible(t *testing.T) {
	tr := NewDefaultTrex()
	tr.Reset()
	tr.SetPhaseForTest(TrexPhasePlay)
	tr.SetContractForTest(TrexContractQueens)
	p := tr.GetPlayer(1)
	p.Reset()
	p.AddCard(txCard(CardDesignSpade, 3))
	p.AddCard(txCard(CardDesignHeart, 9))
	tr.trick = []TrexTrickCard{{PlayerIdx: 0, Card: txCard(CardDesignSpade, 10)}}
	tr.SetCurrentPlayerForTest(1)

	assert.Equal(t, []int{0}, tr.GetValidPlayIndices(1))
	assert.ErrorContains(t, tr.PlayCard(1, 1), "not a legal play")
}

func TestTrex_TheKingOfHeartsDealEndsAsSoonAsItIsTaken(t *testing.T) {
	// 残り 12 トリックを消化しても点は動かない。付き合わせる理由がない。
	tr := NewDefaultTrex()
	tr.Reset()
	tr.SetPhaseForTest(TrexPhasePlay)
	tr.SetContractForTest(TrexContractKingOfHearts)
	for i := range tr.GetPlayers() {
		tr.GetPlayer(i).Reset()
	}
	tr.trick = []TrexTrickCard{
		{PlayerIdx: 0, Card: txCard(CardDesignHeart, 1)},
		{PlayerIdx: 1, Card: txCard(CardDesignHeart, 13)},
		{PlayerIdx: 2, Card: txCard(CardDesignHeart, 2)},
		{PlayerIdx: 3, Card: txCard(CardDesignHeart, 3)},
	}
	tr.resolveTrick()

	assert.Equal(t, TrexKingOfHeartsPenalty, tr.GetScore(0))
	assert.Equal(t, TrexPhaseDealEnd, tr.GetPhase(), "no reason to play the remaining tricks")
}

func TestTrex_TheDominoesPayOutTwoHundredDownToFifty(t *testing.T) {
	tr := NewDefaultTrex()
	tr.Reset()
	tr.SetContractForTest(TrexContractTrix)
	tr.SetPhaseForTest(TrexPhasePlay)
	tr.dealScores = make([]int, TrexPlayerCnt)
	tr.finishOrder = []int{2, 0, 3}
	for i := range tr.GetPlayers() {
		tr.GetPlayer(i).Reset()
	}
	tr.GetPlayer(1).AddCard(txCard(CardDesignSpade, 2)) // 最後の 1 人
	tr.finishTrix()

	assert.Equal(t, 200, tr.GetScore(2))
	assert.Equal(t, 150, tr.GetScore(0))
	assert.Equal(t, 100, tr.GetScore(3))
	assert.Equal(t, 50, tr.GetScore(1), "the last player still scores")
	total := 0
	for i := range tr.GetPlayers() {
		total += tr.GetScore(i)
	}
	assert.Equal(t, 500, total)
}

func TestTrex_PassingIsOnlyForTheDominoesAndOnlyWhenStuck(t *testing.T) {
	tr := NewDefaultTrex()
	tr.Reset()
	tr.SetPhaseForTest(TrexPhasePlay)
	tr.SetContractForTest(TrexContractQueens)
	tr.SetCurrentPlayerForTest(0)
	assert.ErrorContains(t, tr.Pass(0), "only possible in the dominoes")

	tr.SetContractForTest(TrexContractTrix)
	p := tr.GetPlayer(0)
	p.Reset()
	p.AddCard(txCard(CardDesignSpade, 11)) // 出せる
	assert.ErrorContains(t, tr.Pass(0), "you have a legal play")

	p.Reset()
	p.AddCard(txCard(CardDesignSpade, 5)) // 出せない
	require.NoError(t, tr.Pass(0))
}

func TestTrex_TheKingdomMovesAfterFiveContracts(t *testing.T) {
	tr := NewDefaultTrex()
	tr.Reset()
	first := tr.GetKingIdx()

	// 4 ディール消化しただけでは王国は動かない。
	tr.SetDealNumberForTest(4)
	tr.SetPhaseForTest(TrexPhaseDealEnd)
	require.NoError(t, tr.NextDeal())
	assert.Equal(t, first, tr.GetKingIdx())

	tr.SetDealNumberForTest(5)
	tr.SetPhaseForTest(TrexPhaseDealEnd)
	require.NoError(t, tr.NextDeal())
	assert.Equal(t, (first+1)%TrexPlayerCnt, tr.GetKingIdx(), "five contracts hand the kingdom on")
}

// txPlayDeal drives one deal with CPU decisions. Returns false if it stalls.
func txPlayDeal(t *testing.T, tr *Trex) bool {
	t.Helper()
	for range 400 {
		switch tr.GetPhase() {
		case TrexPhaseDealEnd, TrexPhaseGameEnd:
			return true
		case TrexPhaseChoose:
			a := tr.TrexCpuDecide(tr.GetKingIdx())
			require.NoError(t, tr.ChooseContract(tr.GetKingIdx(), a.Contract))
		case TrexPhasePlay:
			idx := tr.GetCurrentPlayerIdx()
			a := tr.TrexCpuDecide(idx)
			if a.Pass {
				require.NoError(t, tr.Pass(idx))
				continue
			}
			if a.HandIdx < 0 {
				return false
			}
			require.NoError(t, tr.PlayCard(idx, a.HandIdx))
		}
	}
	return false
}

func TestTrex_EveryContractPlaysOut(t *testing.T) {
	tr := NewDefaultTrex()
	tr.Reset()
	for deal := range TrexTotalDeals {
		require.True(t, txPlayDeal(t, tr), "deal %d stalled", deal)
		if tr.GetGameEndFlag() {
			break
		}
		require.NoError(t, tr.NextDeal())
	}
	require.True(t, tr.GetGameEndFlag())
	assert.Equal(t, TrexTotalDeals, tr.GetDealNumber())
	assert.GreaterOrEqual(t, tr.GetWinnerIdx(), 0)
}

func TestTrex_TheWholeGameIsZeroSum(t *testing.T) {
	// 罰点 500 と加点 500 が釣り合うので、20 ディール後の全員の合計は 0 に
	// なる。どこかで点が湧いたり消えたりしていれば必ずここで出る。
	tr := NewDefaultTrex()
	tr.Reset()
	for range TrexTotalDeals {
		require.True(t, txPlayDeal(t, tr))
		if tr.GetGameEndFlag() {
			break
		}
		require.NoError(t, tr.NextDeal())
	}
	require.True(t, tr.GetGameEndFlag())

	total := 0
	for i := range tr.GetPlayers() {
		total += tr.GetScore(i)
	}
	assert.Zero(t, total, "penalties and bonuses must cancel over the whole game")
}

func TestTrex_RejectsIllegalRequests(t *testing.T) {
	tr := NewDefaultTrex()
	tr.Reset()

	// 選択フェーズではプレイできない。
	assert.ErrorContains(t, tr.PlayCard(tr.GetKingIdx(), 0), "not the play phase")
	// ディール進行中に次へは進めない。
	assert.ErrorContains(t, tr.NextDeal(), "still in progress")

	require.NoError(t, tr.ChooseContract(tr.GetKingIdx(), TrexContractQueens))
	cur := tr.GetCurrentPlayerIdx()
	assert.Error(t, tr.PlayCard(cur, -1))
	assert.Error(t, tr.PlayCard(cur, 99))
	assert.Error(t, tr.PlayCard((cur+1)%TrexPlayerCnt, 0), "not that player's turn")
	assert.Error(t, tr.ChooseContract(tr.GetKingIdx(), TrexContractTricks), "choosing is over")
}

func TestTrex_SurvivesAKVRoundTrip(t *testing.T) {
	tr := NewDefaultTrex()
	tr.Reset()
	require.True(t, txPlayDeal(t, tr))

	data, err := json.Marshal(tr)
	require.NoError(t, err)

	restored := NewDefaultTrex()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, tr.GetPhase(), restored.GetPhase())
	assert.Equal(t, tr.GetContract(), restored.GetContract())
	assert.Equal(t, tr.GetKingIdx(), restored.GetKingIdx())
	assert.Equal(t, tr.GetDealNumber(), restored.GetDealNumber())
	for i := range tr.GetPlayers() {
		assert.Equal(t, tr.GetScore(i), restored.GetScore(i), "score %d", i)
		// これが落ちると同じ契約を二度選べてしまう。
		for c := TrexContractKingOfHearts; c <= TrexContractTrix; c++ {
			assert.Equal(t, tr.IsContractUsed(i, c), restored.IsContractUsed(i, c), "used %d/%d", i, c)
		}
	}
}

func TestTrex_UnmarshalRejectsAndClampsHostileSnapshots(t *testing.T) {
	for name, payload := range map[string]string{
		"not json":         "{",
		"seat count":       `{"pl":[],"cfg":{"cd":0},"ph":0,"ct":5}`,
		"bad config":       `{"pl":[{},{},{},{}],"cfg":{"cd":99},"ph":0,"ct":5}`,
		"unknown phase":    `{"pl":[{},{},{},{}],"cfg":{"cd":0},"ph":9,"ct":5}`,
		"unknown contract": `{"pl":[{},{},{},{}],"cfg":{"cd":0},"ph":0,"ct":42}`,
	} {
		t.Run(name, func(t *testing.T) {
			assert.Error(t, json.Unmarshal([]byte(payload), NewDefaultTrex()))
		})
	}

	short := `{"pl":[{},{},{},{}],"cfg":{"cd":0},"ph":0,"ct":5,"uc":[[true]],"sc":[7],"ki":99,"cur":-5,"fo":[9,1]}`
	tr := NewDefaultTrex()
	require.NoError(t, json.Unmarshal([]byte(short), tr))
	assert.Equal(t, 0, tr.GetKingIdx(), "an out-of-range seat is clamped")
	assert.Equal(t, 0, tr.GetCurrentPlayerIdx())
	assert.Equal(t, 7, tr.GetScore(0))
	assert.Zero(t, tr.GetScore(3), "the padded tail must not read past the supplied slice")
	assert.True(t, tr.IsContractUsed(0, TrexContractKingOfHearts))
	assert.False(t, tr.IsContractUsed(3, TrexContractKingOfHearts))
	assert.Equal(t, []int{1}, tr.GetFinishOrder(), "an out-of-range seat is dropped, not kept")
}
