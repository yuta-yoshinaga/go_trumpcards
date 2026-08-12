//go:build test

package domain

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newColourWhistAllCpu は全席 CPU の卓を返す。
//
// **全席 CPU だと Reset() の中でラウンドが終わり切ります**（人間の入力待ちが
// 無いため）。途中を観測したいテストは newColourWhistWithHuman を使ってください。
func newColourWhistAllCpu(t *testing.T, seed int64) *ColourWhist {
	t.Helper()
	seats := make([]*ColourWhistPlayer, ColourWhistPlayerCnt)
	for i := range seats {
		seats[i] = NewColourWhistPlayer(false)
	}
	g := NewColourWhist(NewTrumpCards(0), seats, DefaultColourWhistConfig())
	g.SetRand(rand.New(rand.NewSource(seed)))
	return g
}

// newColourWhistWithHuman は席 0 が人間の卓を返す。
func newColourWhistWithHuman(t *testing.T, seed int64) *ColourWhist {
	t.Helper()
	g := NewDefaultColourWhist()
	g.SetRand(rand.New(rand.NewSource(seed)))
	g.Reset()
	return g
}

// **52 枚を 13 枚ずつ配る。**
//
// **Troel が出た配りでは人間の入力待ちが来ません**——競りを飛ばして CPU の
// 契約者がそのまま切り札を決め、ラウンドが終わり切ってしまうためです。
// 配り直後を見たいので、Troel の出なかった配りで検査します。
func TestColourWhistDealsThirteenEach(t *testing.T) {
	t.Parallel()

	var g *ColourWhist
	for seed := range 50 {
		c := newColourWhistWithHuman(t, int64(seed)+1)
		if c.GetPhase() == ColourWhistPhaseBid {
			g = c
			break
		}
	}
	require.NotNil(t, g, "Troel の出ない配りを引けなかった")

	for i := range ColourWhistPlayerCnt {
		assert.Equal(t, ColourWhistHandSize, g.GetPlayer(i).GetCardsSize(), "席 %d", i)
	}
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.False(t, g.IsTroelForced())
}

// **Troel が出た配りは競りを飛ばす。** そのときフェーズは Bid になりません。
func TestColourWhistTroelSkipsTheAuctionOnDeal(t *testing.T) {
	t.Parallel()

	seen := 0
	for seed := range 120 {
		g := newColourWhistWithHuman(t, int64(seed)+1)
		if !g.IsTroelForced() {
			continue
		}
		seen++
		assert.Equal(t, ColourWhistContractTroel, g.GetContract(), "seed %d", seed)
		assert.NotEqual(t, ColourWhistPhaseBid, g.GetPhase(), "seed %d: 競りに入っている", seed)
		for i := range ColourWhistPlayerCnt {
			assert.False(t, g.HasPassed(i), "seed %d 席 %d: 降りた記録が残っている", seed, i)
		}
	}
	assert.Positive(t, seen, "Troel の出る配りを一度も引けなかった")
}

// **エース 3 枚で Troel が強制成立する。** 競りをしません。
func TestColourWhistTroelIsForcedByThreeAces(t *testing.T) {
	t.Parallel()

	g := newColourWhistAllCpu(t, 1)
	// 手で「席 1 がエース 3 枚、席 2 が 4 枚目」の配りを作る。
	for i := range ColourWhistPlayerCnt {
		g.players[i].ResetRound()
	}
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart} {
		g.players[1].AddCard(NewCard(suit, 1, false))
	}
	g.players[2].AddCard(NewCard(CardDesignDiamond, 1, false))
	g.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	g.players[3].AddCard(NewCard(CardDesignHeart, 7, false))
	g.contract = ColourWhistContractNone
	g.declarerIdx = -1
	g.partnerIdx = -1
	g.passed = make([]bool, ColourWhistPlayerCnt)

	require.True(t, g.detectTroel(), "エース 3 枚なら成立する")
	assert.Equal(t, ColourWhistContractTroel, g.GetContract())
	assert.Equal(t, 1, g.GetDeclarerIdx(), "3 枚持っている席が契約者")
	assert.Equal(t, 2, g.GetPartnerIdx(), "4 枚目の持ち主が相方")
	assert.True(t, g.IsTroelForced())
	// **競りは飛ばされるので、降りた記録は残りません。**
	for i := range ColourWhistPlayerCnt {
		assert.False(t, g.HasPassed(i), "席 %d", i)
	}
}

// **4 枚とも持っていれば相方は付かない。** 単独で 8 トリックです。
func TestColourWhistTroelWithAllFourAcesHasNoPartner(t *testing.T) {
	t.Parallel()

	g := newColourWhistAllCpu(t, 2)
	for i := range ColourWhistPlayerCnt {
		g.players[i].ResetRound()
	}
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		g.players[0].AddCard(NewCard(suit, 1, false))
	}
	g.players[1].AddCard(NewCard(CardDesignSpade, 5, false))
	g.contract = ColourWhistContractNone
	g.declarerIdx = -1
	g.partnerIdx = -1

	require.True(t, g.detectTroel())
	assert.Equal(t, 0, g.GetDeclarerIdx())
	assert.Equal(t, -1, g.GetPartnerIdx(), "相方は居ない")
	assert.False(t, g.IsDeclarerSide(1))
	assert.False(t, g.IsDeclarerSide(2))
}

// **エースが 2 枚以下なら Troel にならない。** 負のコントロールです。
func TestColourWhistNoTroelWithTwoAces(t *testing.T) {
	t.Parallel()

	g := newColourWhistAllCpu(t, 3)
	for i := range ColourWhistPlayerCnt {
		g.players[i].ResetRound()
	}
	g.players[0].AddCard(NewCard(CardDesignSpade, 1, false))
	g.players[0].AddCard(NewCard(CardDesignClover, 1, false))
	g.players[1].AddCard(NewCard(CardDesignHeart, 1, false))
	g.players[2].AddCard(NewCard(CardDesignDiamond, 1, false))
	g.contract = ColourWhistContractNone
	g.declarerIdx = -1

	assert.False(t, g.detectTroel(), "2 枚では成立しない")
	assert.Equal(t, ColourWhistContractNone, g.GetContract())
	assert.False(t, g.IsTroelForced())
}

// **Troel は競りで宣言できない。** 配りでしか成立しません。
func TestColourWhistTroelCannotBeBid(t *testing.T) {
	t.Parallel()

	g := newColourWhistWithHuman(t, 5)
	if g.GetPhase() != ColourWhistPhaseBid || !g.IsHumanTurn() {
		t.Skip("この配りでは人間の競り番が来ない")
	}
	assert.Error(t, g.Bid(ColourWhistContractTroel), "競りでは選べない")
	assert.Error(t, g.Bid(99))
	assert.Error(t, g.Bid(-1))
}

// **契約の梯子は上へしか積めない。**
func TestColourWhistBidsMustRise(t *testing.T) {
	t.Parallel()

	g := newColourWhistWithHuman(t, 7)
	if g.GetPhase() != ColourWhistPhaseBid || !g.IsHumanTurn() {
		t.Skip("この配りでは人間の競り番が来ない")
	}
	below := g.GetContract()
	if below >= ColourWhistBidMax {
		t.Skip("すでに最強まで競り上がっている")
	}
	require.NoError(t, g.Bid(below+1))
	assert.Equal(t, below+1, g.GetContract())
	if g.IsHumanTurn() && g.GetPhase() == ColourWhistPhaseBid {
		assert.Error(t, g.Bid(g.GetContract()), "同じ契約は積めない")
	}
}

// **得点はゼロサム。** 卓の合計は常に 0 です。
func TestColourWhistScoresAreZeroSum(t *testing.T) {
	for seed := range 40 {
		g := newColourWhistAllCpu(t, int64(seed)+1)
		g.Reset()

		for round := 0; round < 30; round++ {
			total := 0
			for i := range ColourWhistPlayerCnt {
				total += g.GetPlayer(i).GetScore()
			}
			require.Zero(t, total, "seed %d ラウンド %d: 合計が %d", seed, round, total)
			if g.GetGameEndFlag() {
				break
			}
			require.Equal(t, ColourWhistPhaseRoundEnd, g.GetPhase(), "seed %d", seed)
			require.NoError(t, g.NextRound())
		}
		require.True(t, g.GetGameEndFlag(), "seed %d: 規定ラウンドで終わらなかった", seed)
	}
}

// **どの契約でもゼロサムが崩れず、端数も出ない。**
func TestColourWhistEveryContractIsZeroSum(t *testing.T) {
	t.Parallel()

	for _, contract := range []int{
		ColourWhistContractSamen, ColourWhistContractAlleen,
		ColourWhistContractMiserie, ColourWhistContractTroel,
	} {
		for _, made := range []bool{true, false} {
			g := newColourWhistAllCpu(t, 9)
			g.Reset()
			for i := range ColourWhistPlayerCnt {
				g.players[i].SetScore(0)
			}
			g.contract = contract
			g.declarerIdx = 1
			g.partnerIdx = -1
			if ColourWhistHasPartner(contract) {
				g.partnerIdx = 3
			}
			g.declarerTricks = 9

			declarers, defenders := g.sideSizes()
			require.Positive(t, declarers)
			assert.Zero(t, (g.perDefenderStake()*defenders)%declarers,
				"%s: 端数が出る", ColourWhistContractName(contract))

			total := 0
			for i := range ColourWhistPlayerCnt {
				total += g.roundScoreFor(i, made)
			}
			assert.Zero(t, total, "%s made=%v: 合計が %d",
				ColourWhistContractName(contract), made, total)
		}
	}
}

// **CPU 同士で規定ラウンドまで進み切る。**
func TestColourWhistAlwaysTerminates(t *testing.T) {
	const games = 40
	stalled := 0

	for seed := range games {
		g := newColourWhistAllCpu(t, int64(seed)+1)
		g.Reset()

		steps := 0
		for !g.GetGameEndFlag() {
			steps++
			if steps > 500 {
				stalled++
				break
			}
			if g.GetPhase() == ColourWhistPhaseRoundEnd {
				require.NoError(t, g.NextRound())
				continue
			}
			before := g.GetTrickCount() + g.GetRoundNumber()
			g.CpuPlay()
			if before == g.GetTrickCount()+g.GetRoundNumber() && g.GetPhase() == ColourWhistPhasePlay {
				stalled++
				break
			}
		}
	}
	assert.Zero(t, stalled, "%d 局中 %d 局が進まなくなった", games, stalled)
}

// **13 トリックちょうどで 1 ラウンドが終わる。**
func TestColourWhistPlaysThirteenTricks(t *testing.T) {
	for seed := range 20 {
		g := newColourWhistAllCpu(t, int64(seed)+1)
		g.Reset()
		assert.Equal(t, ColourWhistTrickCnt, g.GetTrickCount(), "seed %d", seed)
		for i := range ColourWhistPlayerCnt {
			assert.Zero(t, g.GetPlayer(i).GetCardsSize(), "seed %d 席 %d", seed, i)
		}
	}
}

// **Miserie だけが切り札なし。**
func TestColourWhistContractShapes(t *testing.T) {
	t.Parallel()

	assert.True(t, ColourWhistIsMiserie(ColourWhistContractMiserie))
	assert.False(t, ColourWhistNeedsTrump(ColourWhistContractMiserie))
	assert.False(t, ColourWhistHasPartner(ColourWhistContractMiserie))
	assert.Zero(t, ColourWhistContractTarget(ColourWhistContractMiserie))

	for _, c := range []int{ColourWhistContractSamen, ColourWhistContractAlleen, ColourWhistContractTroel} {
		assert.True(t, ColourWhistNeedsTrump(c), "%s", ColourWhistContractName(c))
		assert.Equal(t, ColourWhistTrickTarget, ColourWhistContractTarget(c))
	}
	// **2 対 2 は Samen と Troel だけ。**
	assert.True(t, ColourWhistHasPartner(ColourWhistContractSamen))
	assert.True(t, ColourWhistHasPartner(ColourWhistContractTroel))
	assert.False(t, ColourWhistHasPartner(ColourWhistContractAlleen))

	assert.Equal(t, "samen", ColourWhistContractName(ColourWhistContractSamen))
	assert.Equal(t, "alleen", ColourWhistContractName(ColourWhistContractAlleen))
	assert.Equal(t, "miserie", ColourWhistContractName(ColourWhistContractMiserie))
	assert.Equal(t, "troel", ColourWhistContractName(ColourWhistContractTroel))
	assert.Equal(t, "none", ColourWhistContractName(99))
}

// **フォローできるならフォローする。**
func TestColourWhistMustFollowSuit(t *testing.T) {
	checked := 0
	for seed := range 10 {
		g := newColourWhistWithHuman(t, int64(seed)+1)
		for step := 0; step < 400 && !g.GetGameEndFlag(); step++ {
			switch {
			case g.GetPhase() == ColourWhistPhaseRoundEnd:
				require.NoError(t, g.NextRound())
			case !g.IsHumanTurn():
				g.CpuPlay()
			case g.GetPhase() == ColourWhistPhaseBid:
				require.NoError(t, g.Bid(ColourWhistContractNone))
			case g.GetPhase() == ColourWhistPhaseCall:
				require.NoError(t, g.Call(CardDesignSpade))
			case g.GetPhase() == ColourWhistPhasePlay:
				if len(g.GetTrick()) > 0 {
					p := g.GetPlayer(0)
					lead := g.GetTrick()[0].Card.GetDesign()
					has := false
					for k := range p.GetCardsSize() {
						if p.GetCard(k).GetDesign() == lead {
							has = true
							break
						}
					}
					if has {
						for _, v := range g.GetValidPlayIndices(0) {
							assert.Equal(t, lead, p.GetCard(v).GetDesign(), "フォローを外せてしまう")
						}
						checked++
					}
				}
				valid := g.GetValidPlayIndices(0)
				require.NotEmpty(t, valid)
				require.NoError(t, g.PlayCard(valid[0]))
			default:
				step = 400
			}
		}
	}
	assert.Positive(t, checked, "人間の手番を一度も踏めなかった")
}

func TestColourWhistConfigValidate(t *testing.T) {
	t.Parallel()

	assert.NoError(t, DefaultColourWhistConfig().Validate())
	assert.Error(t, ColourWhistConfig{Rounds: ColourWhistRoundsMin - 1}.Validate())
	assert.Error(t, ColourWhistConfig{Rounds: ColourWhistRoundsMax + 1}.Validate())

	g := NewColourWhist(nil, nil, ColourWhistConfig{Rounds: 9999})
	assert.Equal(t, DefaultColourWhistConfig(), g.GetConfig())
	assert.True(t, g.GetPlayer(0).GetIsHuman())

	g.SetConfig(ColourWhistConfig{Rounds: 12})
	assert.Equal(t, 12, g.GetConfig().Rounds)
	g.SetConfig(ColourWhistConfig{Rounds: 1})
	assert.Equal(t, 12, g.GetConfig().Rounds, "範囲外は無視する")
}

func TestColourWhistGiveUpAndAccessors(t *testing.T) {
	g := newColourWhistWithHuman(t, 11)
	g.GiveUp()
	assert.True(t, g.GetGameEndFlag())
	assert.NotEqual(t, 0, g.GetWinnerIdx(), "投了した人は勝たない")
	assert.False(t, g.IsHumanTurn())
	assert.Nil(t, g.GetHint())
	g.GiveUp()
	assert.True(t, g.GetGameEndFlag())

	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
	assert.Nil(t, g.GetValidPlayIndices(-1))
	assert.False(t, g.HasPassed(-1))
	assert.False(t, g.HasPassed(99))
	assert.Zero(t, colourWhistRank(nil))
	assert.Equal(t, 14, colourWhistRank(NewCard(CardDesignSpade, 1, false)), "エースが最強")
	assert.Equal(t, "spade", colourWhistSuitName(CardDesignSpade))
	assert.Equal(t, "notrump", colourWhistSuitName(ColourWhistNoTrump))
}

func TestColourWhistCallRejectsBadSuit(t *testing.T) {
	t.Parallel()

	g := newColourWhistWithHuman(t, 13)
	g.phase = ColourWhistPhaseCall
	g.contract = ColourWhistContractSamen
	g.declarerIdx = 0
	g.currentTurn = 0
	assert.Error(t, g.call(0, 99))
	assert.Error(t, g.call(0, 0), "スートは 1..4")
}

func TestColourWhistHint(t *testing.T) {
	g := newColourWhistWithHuman(t, 4)
	if !g.IsHumanTurn() || g.GetPhase() != ColourWhistPhaseBid {
		t.Skip("この配りでは人間の競り番が来ない")
	}
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "colourWhistBidStrength", h.Reason)
	require.NotNil(t, h.Contract)
	assert.Nil(t, h.CardIndex)
}
