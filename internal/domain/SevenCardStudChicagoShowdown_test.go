//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chiTable builds a 4-seat Chicago game with an even pot and no side pots.
func chiTable(t *testing.T, pot int) *SevenCardStud {
	t.Helper()
	s := NewDefaultSevenCardStudChicago()
	require.True(t, s.GetIsChicago())
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

func chiWon(s *SevenCardStud, idx int) int {
	for _, r := range s.GetRoundResults() {
		if r.PlayerIdx == idx {
			return r.WonAmount
		}
	}
	return 0
}

func chiResult(t *testing.T, s *SevenCardStud, idx int) SevenCardStudResult {
	t.Helper()
	for _, r := range s.GetRoundResults() {
		if r.PlayerIdx == idx {
			return r
		}
	}
	t.Fatalf("no result for seat %d", idx)
	return SevenCardStudResult{}
}

// **半分は伏せ札の最高スペードへ。** 役の勝者とスペードの持ち主が別なら、
// ポットは割れる。
func TestSevenCardStudChicago_TheHighestSpadeInTheHoleTakesHalf(t *testing.T) {
	s := chiTable(t, 400)
	// 席 0: エースのペアでハイ勝ち。伏せ札にスペードは無い。
	hlDeal(s, 0, hlCard(CardDesignHeart, 1), hlCard(CardDesignClover, 1), hlCard(CardDesignDiamond, 13),
		hlCard(CardDesignDiamond, 12), hlCard(CardDesignHeart, 11), hlCard(CardDesignClover, 10), hlCard(CardDesignDiamond, 9))
	// 席 1: 役は弱いが伏せ札に A♠。
	hlDeal(s, 1, hlCard(CardDesignSpade, 1), hlCard(CardDesignHeart, 3), hlCard(CardDesignClover, 4),
		hlCard(CardDesignDiamond, 5), hlCard(CardDesignHeart, 7), hlCard(CardDesignClover, 8), hlCard(CardDesignDiamond, 10))
	for i := 2; i < len(s.players); i++ {
		s.players[i].SetFolded(true)
	}

	s.resolveShowdown()

	assert.Equal(t, 400, chiWon(s, 0)+chiWon(s, 1), "the pot must balance")
	assert.Equal(t, 200, chiWon(s, 0), "high half")
	assert.Equal(t, 200, chiWon(s, 1), "spade half")

	spadeSide := chiResult(t, s, 1)
	require.NotNil(t, spadeSide.SpadeCard)
	assert.Equal(t, 1, spadeSide.SpadeCard.GetValue(), "A♠")
	assert.Equal(t, 200, spadeSide.WonSpade)
}

// **エースは 1 でもスペードでは最高。** 額面で比べると K♠ に負ける。
func TestSevenCardStudChicago_AceOfSpadesBeatsTheKingOfSpades(t *testing.T) {
	s := chiTable(t, 400)
	hlDeal(s, 0, hlCard(CardDesignSpade, 13), hlCard(CardDesignHeart, 1), hlCard(CardDesignClover, 1),
		hlCard(CardDesignDiamond, 12), hlCard(CardDesignHeart, 11), hlCard(CardDesignClover, 10), hlCard(CardDesignDiamond, 9))
	hlDeal(s, 1, hlCard(CardDesignSpade, 1), hlCard(CardDesignHeart, 3), hlCard(CardDesignClover, 4),
		hlCard(CardDesignDiamond, 5), hlCard(CardDesignHeart, 7), hlCard(CardDesignClover, 8), hlCard(CardDesignDiamond, 6))
	for i := 2; i < len(s.players); i++ {
		s.players[i].SetFolded(true)
	}

	s.resolveShowdown()

	assert.Equal(t, 200, chiResult(t, s, 1).WonSpade, "A♠ takes the spade half")
	assert.Zero(t, chiResult(t, s, 0).WonSpade, "K♠ loses it")
}

// **1 人も伏せ札にスペードを持っていなければハイが総取り。**
func TestSevenCardStudChicago_WithNoSpadeInAnyHoleTheHighTakesEverything(t *testing.T) {
	s := chiTable(t, 400)
	hlDeal(s, 0, hlCard(CardDesignHeart, 1), hlCard(CardDesignClover, 1), hlCard(CardDesignDiamond, 13),
		hlCard(CardDesignSpade, 12), hlCard(CardDesignSpade, 11), hlCard(CardDesignSpade, 10), hlCard(CardDesignSpade, 9))
	hlDeal(s, 1, hlCard(CardDesignHeart, 3), hlCard(CardDesignClover, 4), hlCard(CardDesignDiamond, 5),
		hlCard(CardDesignSpade, 7), hlCard(CardDesignSpade, 8), hlCard(CardDesignHeart, 10), hlCard(CardDesignClover, 6))
	for i := 2; i < len(s.players); i++ {
		s.players[i].SetFolded(true)
	}

	s.resolveShowdown()

	// 表札にはスペードが並んでいるが、伏せ札には 1 枚も無い。
	assert.Equal(t, 400, chiWon(s, 0), "the high side scoops")
	assert.Zero(t, chiWon(s, 1))
	assert.Nil(t, chiResult(t, s, 0).SpadeCard)
}

// **同じ人が両方取ることもある。**
func TestSevenCardStudChicago_OnePlayerCanScoopBothHalves(t *testing.T) {
	s := chiTable(t, 400)
	hlDeal(s, 0, hlCard(CardDesignSpade, 1), hlCard(CardDesignHeart, 1), hlCard(CardDesignClover, 1),
		hlCard(CardDesignDiamond, 12), hlCard(CardDesignHeart, 11), hlCard(CardDesignClover, 10), hlCard(CardDesignDiamond, 9))
	// **連番を渡さない。** 3-4-5-6-7 を持たせると**ストレートになってハイを
	// 取ってしまい**、スクープの検証にならない (Hi-Lo 側の同じ罠と同型)。
	hlDeal(s, 1, hlCard(CardDesignHeart, 3), hlCard(CardDesignClover, 4), hlCard(CardDesignDiamond, 5),
		hlCard(CardDesignHeart, 9), hlCard(CardDesignClover, 11), hlCard(CardDesignDiamond, 12), hlCard(CardDesignHeart, 2))
	for i := 2; i < len(s.players); i++ {
		s.players[i].SetFolded(true)
	}

	s.resolveShowdown()

	assert.Equal(t, 400, chiWon(s, 0), "a scoop takes both halves")
	assert.Zero(t, chiWon(s, 1))
}

// **奇数チップはハイ側に寄せる** (ポーカー慣例、Hi-Lo と同じ)。
func TestSevenCardStudChicago_TheOddChipGoesToTheHighSide(t *testing.T) {
	s := chiTable(t, 401)
	hlDeal(s, 0, hlCard(CardDesignHeart, 1), hlCard(CardDesignClover, 1), hlCard(CardDesignDiamond, 13),
		hlCard(CardDesignDiamond, 12), hlCard(CardDesignHeart, 11), hlCard(CardDesignClover, 10), hlCard(CardDesignDiamond, 9))
	hlDeal(s, 1, hlCard(CardDesignSpade, 1), hlCard(CardDesignHeart, 3), hlCard(CardDesignClover, 4),
		hlCard(CardDesignDiamond, 5), hlCard(CardDesignHeart, 7), hlCard(CardDesignClover, 8), hlCard(CardDesignDiamond, 10))
	for i := 2; i < len(s.players); i++ {
		s.players[i].SetFolded(true)
	}

	s.resolveShowdown()

	assert.Equal(t, 401, chiWon(s, 0)+chiWon(s, 1))
	assert.Equal(t, 201, chiWon(s, 0), "the odd chip goes high")
	assert.Equal(t, 200, chiWon(s, 1))
}

// **降りた席のスペードは数えない。** 降りた人が卓に残っている扱いになると、
// ポットの半分が場から消える。
func TestSevenCardStudChicago_AFoldedSeatCannotWinTheSpadeHalf(t *testing.T) {
	s := chiTable(t, 400)
	hlDeal(s, 0, hlCard(CardDesignHeart, 1), hlCard(CardDesignClover, 1), hlCard(CardDesignDiamond, 13),
		hlCard(CardDesignDiamond, 12), hlCard(CardDesignHeart, 11), hlCard(CardDesignClover, 10), hlCard(CardDesignDiamond, 9))
	// 席 1 は A♠ を伏せているが降りている。
	hlDeal(s, 1, hlCard(CardDesignSpade, 1), hlCard(CardDesignHeart, 3), hlCard(CardDesignClover, 4),
		hlCard(CardDesignDiamond, 5), hlCard(CardDesignHeart, 7), hlCard(CardDesignClover, 8), hlCard(CardDesignDiamond, 10))
	s.players[1].SetFolded(true)
	// 席 2 は 9♠ を伏せている。
	hlDeal(s, 2, hlCard(CardDesignSpade, 9), hlCard(CardDesignHeart, 4), hlCard(CardDesignClover, 5),
		hlCard(CardDesignDiamond, 6), hlCard(CardDesignHeart, 8), hlCard(CardDesignClover, 2), hlCard(CardDesignDiamond, 3))
	s.players[3].SetFolded(true)

	s.resolveShowdown()

	assert.Equal(t, 200, chiWon(s, 0), "high half")
	assert.Equal(t, 200, chiWon(s, 2), "the 9 of spades takes it once the ace has folded")
	assert.Zero(t, chiWon(s, 1))
}

// **共有カードは伏せ札ではない。** デッキが尽きたときの 1 枚は全員に見えている
// ので、それでスペードの半分を主張できてはいけない。
func TestSevenCardStudChicago_TheCommunityCardIsNotAHoleCard(t *testing.T) {
	s := chiTable(t, 400)
	// どの席も伏せ札にスペードを持たない。
	hlDeal(s, 0, hlCard(CardDesignHeart, 1), hlCard(CardDesignClover, 1), hlCard(CardDesignDiamond, 13),
		hlCard(CardDesignDiamond, 12), hlCard(CardDesignHeart, 11), hlCard(CardDesignClover, 10), hlCard(CardDesignDiamond, 9))
	hlDeal(s, 1, hlCard(CardDesignHeart, 3), hlCard(CardDesignClover, 4), hlCard(CardDesignDiamond, 5),
		hlCard(CardDesignHeart, 7), hlCard(CardDesignClover, 8), hlCard(CardDesignDiamond, 10), hlCard(CardDesignHeart, 6))
	for i := 2; i < len(s.players); i++ {
		s.players[i].SetFolded(true)
	}
	// 共有カードが A♠。全員の手に一時的に加わる。
	s.communityCard = hlCard(CardDesignSpade, 1)

	s.resolveShowdown()

	assert.Equal(t, 400, chiWon(s, 0), "the shared ace of spades gives nobody the half")
	assert.Nil(t, chiResult(t, s, 0).SpadeCard)
	assert.Nil(t, chiResult(t, s, 1).SpadeCard)
}
