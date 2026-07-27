//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// cegoStrongCpuHand seats a strong hand (Sküs + honour trumps + kings) on idx.
func cegoStrongCpuHand(g *domain.Cego, idx int) {
	cegoSetHand(g, idx,
		cegoSkusCard(), cegoTrumpCard(21), cegoTrumpCard(20), cegoTrumpCard(19),
		cegoTrumpCard(18), cegoTrumpCard(17), cegoTrumpCard(16), cegoTrumpCard(15),
		cegoKingCard(domain.CardDesignSpade), cegoKingCard(domain.CardDesignHeart),
		cegoKingCard(domain.CardDesignDiamond))
}

// cegoWeakCpuHand seats a weak hand (low pips + one low trump) on idx.
func cegoWeakCpuHand(g *domain.Cego, idx int) {
	cegoSetHand(g, idx,
		cegoSuitCard(domain.CardDesignSpade, 1), cegoSuitCard(domain.CardDesignSpade, 2),
		cegoSuitCard(domain.CardDesignHeart, 1), cegoSuitCard(domain.CardDesignHeart, 2),
		cegoSuitCard(domain.CardDesignDiamond, 1), cegoSuitCard(domain.CardDesignDiamond, 2),
		cegoSuitCard(domain.CardDesignClover, 1), cegoSuitCard(domain.CardDesignClover, 2),
		cegoSuitCard(domain.CardDesignClover, 3), cegoSuitCard(domain.CardDesignClover, 4),
		cegoTrumpCard(3))
}

// --- CPU bidding difficulty branches ---

func TestCegoCpuBidStrongBids(t *testing.T) {
	for _, diff := range []domain.CegoCpuDifficulty{
		domain.CegoCpuDifficultyEasy,
		domain.CegoCpuDifficultyNormal,
		domain.CegoCpuDifficultyHard,
	} {
		g := domain.NewDefaultCego()
		g.Reset()
		g.SetConfig(domain.CegoConfig{CpuDifficulty: diff, TargetDeals: 5})
		g.SetBidPlayerIdx(1) // a CPU seat
		g.SetHighestBid(domain.CegoBidPass)
		cegoStrongCpuHand(g, 1)
		g.CpuBid()
		assert.Equal(t, domain.CegoBidPlay, g.GetHighestBid())
		assert.Equal(t, 1, g.GetHighestBidder())
	}
}

func TestCegoCpuBidWeakPasses(t *testing.T) {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetBidPlayerIdx(1)
	g.SetHighestBid(domain.CegoBidPass)
	g.GetPlayer(1).Reset() // empty hand -> evalHand 0 -> pass
	g.CpuBid()
	assert.Equal(t, -1, g.GetHighestBidder())
}

func TestCegoCpuBidAlreadyBidPasses(t *testing.T) {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetBidPlayerIdx(1)
	g.SetHighestBid(domain.CegoBidPlay) // already at the top -> CPU cannot outbid
	cegoStrongCpuHand(g, 1)
	g.CpuBid()
	// highestBidder must stay unset (the CPU passed).
	assert.Equal(t, -1, g.GetHighestBidder())
}

func TestCegoCpuBidWrongPhaseNoop(t *testing.T) {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetPhase(domain.CegoPhasePlay)
	g.SetBidPlayerIdx(1)
	g.CpuBid() // must be a no-op outside the bid phase
	assert.Equal(t, -1, g.GetHighestBidder())
}

// --- CPU contract difficulty branches ---

func TestCegoCpuContractEasyStrongHandspiel(t *testing.T) {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetConfig(domain.CegoConfig{CpuDifficulty: domain.CegoCpuDifficultyEasy, TargetDeals: 5})
	g.SetDeclarerIdx(1)
	g.SetContract(domain.CegoBidPlay)
	g.SetPhase(domain.CegoPhaseContract)
	cegoStrongCpuHand(g, 1)
	g.CpuChooseContract()
	assert.Equal(t, domain.CegoContractHandspiel, g.GetContractType())
}

func TestCegoCpuContractHardStrongHandspiel(t *testing.T) {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetConfig(domain.CegoConfig{CpuDifficulty: domain.CegoCpuDifficultyHard, TargetDeals: 5})
	g.SetDeclarerIdx(1)
	g.SetContract(domain.CegoBidPlay)
	g.SetPhase(domain.CegoPhaseContract)
	cegoStrongCpuHand(g, 1)
	g.CpuChooseContract()
	assert.Equal(t, domain.CegoContractHandspiel, g.GetContractType())
}

func TestCegoCpuContractHardWeakCego(t *testing.T) {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetConfig(domain.CegoConfig{CpuDifficulty: domain.CegoCpuDifficultyHard, TargetDeals: 5})
	g.SetDeclarerIdx(1)
	g.SetContract(domain.CegoBidPlay)
	g.SetPhase(domain.CegoPhaseContract)
	cegoWeakCpuHand(g, 1)
	g.CpuChooseContract()
	assert.Equal(t, domain.CegoContractCego, g.GetContractType())
}

func TestCegoCpuChooseContractHumanNoop(t *testing.T) {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetDeclarerIdx(0) // human declarer
	g.SetContract(domain.CegoBidPlay)
	g.SetPhase(domain.CegoPhaseContract)
	g.CpuChooseContract() // must be a no-op for a human declarer
	assert.Equal(t, domain.CegoContractNone, g.GetContractType())
}

// --- CPU exchange (Cego keep) ---

func TestCegoCpuDiscardKeepsBest(t *testing.T) {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetDeclarerIdx(1) // CPU declarer
	g.SetContract(domain.CegoBidPlay)
	g.SetContractType(domain.CegoContractCego)
	g.SetPhase(domain.CegoPhaseExchange)
	// Mixed hand exercising every cegoKeepValue arm (trull / king / trump / plain).
	cegoSetHand(g, 1,
		cegoTrumpCard(1), // Pagat (trull) -> highest keep value
		cegoKingCard(domain.CardDesignSpade),
		cegoTrumpCard(9),
		cegoSuitCard(domain.CardDesignHeart, 2),
		cegoSuitCard(domain.CardDesignDiamond, 3))
	g.CpuDiscard()
	assert.Equal(t, domain.CegoPhasePlay, g.GetPhase())
	assert.Equal(t, 0, g.GetStashOwner())
}

// --- getValidPlayIndices: suit-follow branches ---

func cegoPlayable(t *testing.T, hand []*domain.Card, trick []*domain.TrickCard) []int {
	t.Helper()
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.CegoBidPlay)
	g.SetPhase(domain.CegoPhasePlay)
	cegoSetHand(g, 0, hand...)
	g.SetCurrentTrick(trick)
	return g.GetPlayableIndices(0)
}

func TestCegoValidPlaySuitFollowHasLed(t *testing.T) {
	got := cegoPlayable(t,
		[]*domain.Card{cegoSuitCard(domain.CardDesignSpade, 7), cegoSuitCard(domain.CardDesignHeart, 4), cegoTrumpCard(2)},
		[]*domain.TrickCard{{PlayerIdx: 1, Card: cegoSuitCard(domain.CardDesignSpade, 3)}})
	assert.Equal(t, []int{0}, got) // must follow the led spade
}

func TestCegoValidPlaySuitVoidMustOvertrump(t *testing.T) {
	got := cegoPlayable(t,
		[]*domain.Card{cegoTrumpCard(10), cegoTrumpCard(2), cegoSuitCard(domain.CardDesignHeart, 4)},
		[]*domain.TrickCard{
			{PlayerIdx: 1, Card: cegoSuitCard(domain.CardDesignSpade, 3)},
			{PlayerIdx: 2, Card: cegoTrumpCard(5)},
		})
	assert.Equal(t, []int{0}, got) // void in spades, a trump was played -> must beat T5
}

func TestCegoValidPlaySuitVoidTrumpNoHigher(t *testing.T) {
	got := cegoPlayable(t,
		[]*domain.Card{cegoTrumpCard(2), cegoTrumpCard(5), cegoSuitCard(domain.CardDesignHeart, 4)},
		[]*domain.TrickCard{
			{PlayerIdx: 1, Card: cegoSuitCard(domain.CardDesignSpade, 3)},
			{PlayerIdx: 2, Card: cegoTrumpCard(20)},
		})
	assert.Equal(t, []int{0, 1}, got) // cannot beat T20 -> any trump is legal
}

func TestCegoValidPlaySuitVoidNoTrumpInTrick(t *testing.T) {
	got := cegoPlayable(t,
		[]*domain.Card{cegoTrumpCard(2), cegoTrumpCard(5)},
		[]*domain.TrickCard{
			{PlayerIdx: 1, Card: cegoSuitCard(domain.CardDesignSpade, 3)},
			{PlayerIdx: 2, Card: cegoSuitCard(domain.CardDesignHeart, 4)},
		})
	assert.Equal(t, []int{0, 1}, got) // void, no trump led yet -> may play any trump
}

func TestCegoValidPlaySuitVoidNoTrump(t *testing.T) {
	got := cegoPlayable(t,
		[]*domain.Card{cegoSuitCard(domain.CardDesignHeart, 4), cegoSuitCard(domain.CardDesignDiamond, 5)},
		[]*domain.TrickCard{{PlayerIdx: 1, Card: cegoSuitCard(domain.CardDesignSpade, 3)}})
	assert.Equal(t, []int{0, 1}, got) // void in led suit and no trumps -> anything goes
}

// --- getValidPlayIndices: trump-led branches ---

func TestCegoValidPlayTrumpLedMustOvertrump(t *testing.T) {
	got := cegoPlayable(t,
		[]*domain.Card{cegoTrumpCard(10), cegoTrumpCard(2), cegoSuitCard(domain.CardDesignSpade, 3)},
		[]*domain.TrickCard{{PlayerIdx: 1, Card: cegoTrumpCard(5)}})
	assert.Equal(t, []int{0}, got) // trump led -> must play a higher trump (T10)
}

func TestCegoValidPlayTrumpLedNoHigher(t *testing.T) {
	got := cegoPlayable(t,
		[]*domain.Card{cegoTrumpCard(2), cegoTrumpCard(5), cegoSuitCard(domain.CardDesignSpade, 3)},
		[]*domain.TrickCard{{PlayerIdx: 1, Card: cegoTrumpCard(20)}})
	assert.Equal(t, []int{0, 1}, got) // cannot beat T20 -> any trump is legal
}

func TestCegoValidPlayTrumpLedVoid(t *testing.T) {
	got := cegoPlayable(t,
		[]*domain.Card{cegoSuitCard(domain.CardDesignSpade, 3), cegoSuitCard(domain.CardDesignHeart, 4)},
		[]*domain.TrickCard{{PlayerIdx: 1, Card: cegoTrumpCard(5)}})
	assert.Equal(t, []int{0, 1}, got) // no trumps -> discard anything
}

// --- CPU play strategy branches ---

// cegoCpuPlaySetup builds a Play-phase game with a fixed declarer, difficulty,
// current trick and the acting CPU's hand, then returns the game.
func cegoCpuPlaySetup(declarer, actor int, diff domain.CegoCpuDifficulty, hand []*domain.Card, trick []*domain.TrickCard) *domain.Cego {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetConfig(domain.CegoConfig{CpuDifficulty: diff, TargetDeals: 5})
	g.SetDeclarerIdx(declarer)
	g.SetContract(domain.CegoBidPlay)
	g.SetPhase(domain.CegoPhasePlay)
	g.SetCurrentPlayerIdx(actor)
	cegoSetHand(g, actor, hand...)
	g.SetCurrentTrick(trick)
	return g
}

func TestCegoCpuPlayLeadDeclarerSide(t *testing.T) {
	g := cegoCpuPlaySetup(1, 1, domain.CegoCpuDifficultyNormal,
		[]*domain.Card{cegoSuitCard(domain.CardDesignSpade, 3), cegoTrumpCard(10)}, nil)
	before := g.GetPlayer(1).GetCardsSize()
	g.CpuPlay()
	assert.Equal(t, before-1, g.GetPlayer(1).GetCardsSize())
}

func TestCegoCpuPlayLeadOpponent(t *testing.T) {
	g := cegoCpuPlaySetup(0, 1, domain.CegoCpuDifficultyNormal,
		[]*domain.Card{cegoSuitCard(domain.CardDesignSpade, 3), cegoKingCard(domain.CardDesignSpade)}, nil)
	before := g.GetPlayer(1).GetCardsSize()
	g.CpuPlay()
	assert.Equal(t, before-1, g.GetPlayer(1).GetCardsSize())
}

func TestCegoCpuPlayAllyWinning(t *testing.T) {
	// Declarer=0 (human). Opponent 2 leads a strong trump (winning), opponent 3 follows.
	g := cegoCpuPlaySetup(0, 3, domain.CegoCpuDifficultyNormal,
		[]*domain.Card{cegoTrumpCard(2), cegoTrumpCard(5)},
		[]*domain.TrickCard{{PlayerIdx: 2, Card: cegoTrumpCard(20)}})
	before := g.GetPlayer(3).GetCardsSize()
	g.CpuPlay()
	assert.Equal(t, before-1, g.GetPlayer(3).GetCardsSize())
}

func TestCegoCpuPlayCanBeatDeclarer(t *testing.T) {
	// Declarer=0 leads a low spade; opponent 1 can overtake it.
	g := cegoCpuPlaySetup(0, 1, domain.CegoCpuDifficultyNormal,
		[]*domain.Card{cegoSuitCard(domain.CardDesignSpade, 7), cegoSuitCard(domain.CardDesignSpade, 1)},
		[]*domain.TrickCard{{PlayerIdx: 0, Card: cegoSuitCard(domain.CardDesignSpade, 3)}})
	before := g.GetPlayer(1).GetCardsSize()
	g.CpuPlay()
	assert.Equal(t, before-1, g.GetPlayer(1).GetCardsSize())
}

func TestCegoCpuPlayCannotBeatDucks(t *testing.T) {
	// Declarer=0 leads a high spade; opponent 1 cannot beat it -> ducks.
	g := cegoCpuPlaySetup(0, 1, domain.CegoCpuDifficultyNormal,
		[]*domain.Card{cegoSuitCard(domain.CardDesignSpade, 1), cegoSuitCard(domain.CardDesignSpade, 2)},
		[]*domain.TrickCard{{PlayerIdx: 0, Card: cegoSuitCard(domain.CardDesignSpade, 7)}})
	before := g.GetPlayer(1).GetCardsSize()
	g.CpuPlay()
	assert.Equal(t, before-1, g.GetPlayer(1).GetCardsSize())
}

func TestCegoCpuPlayEasyRandom(t *testing.T) {
	g := cegoCpuPlaySetup(0, 1, domain.CegoCpuDifficultyEasy,
		[]*domain.Card{cegoSuitCard(domain.CardDesignSpade, 3), cegoSuitCard(domain.CardDesignHeart, 4), cegoTrumpCard(2)}, nil)
	before := g.GetPlayer(1).GetCardsSize()
	g.CpuPlay()
	assert.Equal(t, before-1, g.GetPlayer(1).GetCardsSize())
}

func TestCegoCpuPlaySingleValid(t *testing.T) {
	g := cegoCpuPlaySetup(0, 1, domain.CegoCpuDifficultyNormal,
		[]*domain.Card{cegoSuitCard(domain.CardDesignSpade, 3)},
		[]*domain.TrickCard{{PlayerIdx: 0, Card: cegoSuitCard(domain.CardDesignSpade, 7)}})
	g.CpuPlay()
	assert.Equal(t, 0, g.GetPlayer(1).GetCardsSize())
}

func TestCegoCpuPlayWrongPhaseNoop(t *testing.T) {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetPhase(domain.CegoPhaseBid)
	g.SetCurrentPlayerIdx(1)
	before := g.GetPlayer(1).GetCardsSize()
	g.CpuPlay()
	assert.Equal(t, before, g.GetPlayer(1).GetCardsSize())
}

// --- checkGameEnd tie vs solo winner + humanResult ---

// cegoScoreTerminalDeal scores a single terminal deal where the human declarer
// (seat 0) captures exactly 53 points (a loss: declarer -30, each opponent +10),
// starting from the supplied cumulative scores. TargetDeals=1 so it ends the match.
func cegoScoreTerminalDeal(t *testing.T, pre [domain.CegoPlayerCnt]int) *domain.Cego {
	t.Helper()
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetConfig(domain.CegoConfig{CpuDifficulty: domain.CegoCpuDifficultyNormal, TargetDeals: 1})
	g.SetRoundNumber(1)
	g.SetDeclarerIdx(0)
	g.SetContract(domain.CegoBidPlay)
	g.SetContractType(domain.CegoContractHandspiel)
	// Stash owned by the declarer summing to exactly 53 points (10 kings + 1 Cavalier).
	stash := make([]*domain.Card, 0, 11)
	for i := 0; i < 10; i++ {
		stash = append(stash, cegoKingCard(domain.CardDesignSpade)) // 5 pts each -> 50
	}
	stash = append(stash, cegoSuitCard(domain.CardDesignSpade, 6)) // Cavalier -> 3 pts
	g.SetStash(stash)
	g.SetStashOwner(0)
	g.SetPlayerScores(pre)
	g.SetPhase(domain.CegoPhaseRoundEnd)
	g.ScoreRound()
	return g
}

func TestCegoGameEndTie(t *testing.T) {
	// pre[0]-30 == pre[i]+10 for all i  =>  all four end on 10.
	g := cegoScoreTerminalDeal(t, [domain.CegoPlayerCnt]int{40, 0, 0, 0})
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.CegoPhaseGameEnd, g.GetPhase())
	assert.Equal(t, -1, g.GetWinnerPlayer())
	assert.Equal(t, domain.CegoResultNone, g.GetResult())
}

func TestCegoGameEndHumanWins(t *testing.T) {
	g := cegoScoreTerminalDeal(t, [domain.CegoPlayerCnt]int{100, 0, 0, 0})
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerPlayer())
	assert.Equal(t, domain.CegoResultWin, g.GetResult())
}

func TestCegoGameEndHumanLoses(t *testing.T) {
	g := cegoScoreTerminalDeal(t, [domain.CegoPlayerCnt]int{0, 100, 0, 0})
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetWinnerPlayer())
	assert.Equal(t, domain.CegoResultLose, g.GetResult())
}

// --- Human action error branches ---

func TestCegoPlayerBidErrors(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		g := cegoNewReset()
		g.SetPhase(domain.CegoPhasePlay)
		assert.Error(t, g.PlayerBid(domain.CegoBidPlay))
	})
	t.Run("not human turn", func(t *testing.T) {
		g := cegoNewReset()
		g.SetBidPlayerIdx(1)
		assert.Error(t, g.PlayerBid(domain.CegoBidPlay))
	})
	t.Run("invalid bid", func(t *testing.T) {
		g := cegoNewReset()
		g.SetBidPlayerIdx(0)
		assert.Error(t, g.PlayerBid(domain.CegoBidPass))
	})
	t.Run("bid not above highest", func(t *testing.T) {
		g := cegoNewReset()
		g.SetBidPlayerIdx(0)
		g.SetHighestBid(domain.CegoBidPlay)
		assert.Error(t, g.PlayerBid(domain.CegoBidPlay))
	})
}

func TestCegoPlayerPassErrors(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		g := cegoNewReset()
		g.SetPhase(domain.CegoPhasePlay)
		assert.Error(t, g.PlayerPass())
	})
	t.Run("not human turn", func(t *testing.T) {
		g := cegoNewReset()
		g.SetBidPlayerIdx(1)
		assert.Error(t, g.PlayerPass())
	})
	t.Run("valid pass", func(t *testing.T) {
		g := cegoNewReset()
		g.SetBidPlayerIdx(0)
		assert.NoError(t, g.PlayerPass())
	})
}

func TestCegoPlayerChooseContractErrors(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		g := cegoNewReset()
		assert.Error(t, g.PlayerChooseContract(domain.CegoContractCego))
	})
	t.Run("not human declarer", func(t *testing.T) {
		g := cegoNewReset()
		g.SetDeclarerIdx(1)
		g.SetContract(domain.CegoBidPlay)
		g.SetPhase(domain.CegoPhaseContract)
		assert.Error(t, g.PlayerChooseContract(domain.CegoContractCego))
	})
	t.Run("invalid contract", func(t *testing.T) {
		g := cegoNewReset()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.CegoBidPlay)
		g.SetPhase(domain.CegoPhaseContract)
		assert.Error(t, g.PlayerChooseContract(domain.CegoContractNone))
	})
}

func TestCegoPlayerDiscardPhaseErrors(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		g := cegoNewReset()
		g.SetPhase(domain.CegoPhasePlay)
		assert.Error(t, g.PlayerDiscard([]int{0}))
	})
	t.Run("not human declarer", func(t *testing.T) {
		g := cegoNewReset()
		g.SetDeclarerIdx(1)
		g.SetContract(domain.CegoBidPlay)
		g.SetContractType(domain.CegoContractCego)
		g.SetPhase(domain.CegoPhaseExchange)
		assert.Error(t, g.PlayerDiscard([]int{0}))
	})
}

func TestCegoPlayerPlayErrors(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		g := cegoNewReset()
		assert.Error(t, g.PlayerPlay(0))
	})
	t.Run("not human turn", func(t *testing.T) {
		g := cegoNewReset()
		g.SetPhase(domain.CegoPhasePlay)
		g.SetCurrentPlayerIdx(1)
		assert.Error(t, g.PlayerPlay(0))
	})
	t.Run("index out of range", func(t *testing.T) {
		g := cegoNewReset()
		g.SetPhase(domain.CegoPhasePlay)
		g.SetCurrentPlayerIdx(0)
		assert.Error(t, g.PlayerPlay(999))
	})
	t.Run("illegal play must follow", func(t *testing.T) {
		g := domain.NewDefaultCego()
		g.Reset()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.CegoBidPlay)
		g.SetPhase(domain.CegoPhasePlay)
		g.SetCurrentPlayerIdx(0)
		cegoSetHand(g, 0, cegoSuitCard(domain.CardDesignSpade, 7), cegoSuitCard(domain.CardDesignHeart, 4))
		g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: cegoSuitCard(domain.CardDesignSpade, 3)}})
		assert.Error(t, g.PlayerPlay(1)) // must follow spade, cannot dump the heart
	})
	t.Run("legal play", func(t *testing.T) {
		g := domain.NewDefaultCego()
		g.Reset()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.CegoBidPlay)
		g.SetPhase(domain.CegoPhasePlay)
		g.SetCurrentPlayerIdx(0)
		cegoSetHand(g, 0, cegoSuitCard(domain.CardDesignSpade, 7), cegoSuitCard(domain.CardDesignHeart, 4))
		g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: cegoSuitCard(domain.CardDesignSpade, 3)}})
		assert.NoError(t, g.PlayerPlay(0))
	})
}

// --- Play-phase hint reason branches ---

func cegoPlayHint(t *testing.T, declarer int, hand []*domain.Card, trick []*domain.TrickCard) *domain.CegoHint {
	t.Helper()
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetDeclarerIdx(declarer)
	g.SetContract(domain.CegoBidPlay)
	g.SetPhase(domain.CegoPhasePlay)
	g.SetCurrentPlayerIdx(0)
	cegoSetHand(g, 0, hand...)
	g.SetCurrentTrick(trick)
	h := g.GetHint()
	require.NotNil(t, h)
	return h
}

func TestCegoHintLeadHigh(t *testing.T) {
	h := cegoPlayHint(t, 0,
		[]*domain.Card{cegoSuitCard(domain.CardDesignSpade, 3), cegoTrumpCard(10)}, nil)
	assert.Equal(t, "lead_high", h.Reason)
}

func TestCegoHintLeadLow(t *testing.T) {
	h := cegoPlayHint(t, 1,
		[]*domain.Card{cegoSuitCard(domain.CardDesignSpade, 3), cegoTrumpCard(10)}, nil)
	assert.Equal(t, "lead_low", h.Reason)
}

func TestCegoHintFollowWin(t *testing.T) {
	h := cegoPlayHint(t, 1,
		[]*domain.Card{cegoSuitCard(domain.CardDesignSpade, 7), cegoSuitCard(domain.CardDesignSpade, 1)},
		[]*domain.TrickCard{{PlayerIdx: 1, Card: cegoSuitCard(domain.CardDesignSpade, 3)}})
	assert.Equal(t, "follow_win", h.Reason)
}

func TestCegoHintFollowDuck(t *testing.T) {
	h := cegoPlayHint(t, 1,
		[]*domain.Card{cegoSuitCard(domain.CardDesignSpade, 1), cegoSuitCard(domain.CardDesignSpade, 2)},
		[]*domain.TrickCard{{PlayerIdx: 1, Card: cegoSuitCard(domain.CardDesignSpade, 7)}})
	assert.Equal(t, "follow_duck", h.Reason)
}

func TestCegoHintBidTakeAndExchange(t *testing.T) {
	// Bid-take: strong human hand in the bid phase.
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetBidPlayerIdx(0)
	cegoStrongCpuHand(g, 0)
	hb := g.GetHint()
	require.NotNil(t, hb)
	assert.Equal(t, "bid_take", hb.Reason)
	require.NotNil(t, hb.Bid)

	// Keep-best: human declarer in the exchange phase.
	g2 := domain.NewDefaultCego()
	g2.Reset()
	g2.SetDeclarerIdx(0)
	g2.SetContract(domain.CegoBidPlay)
	g2.SetContractType(domain.CegoContractCego)
	g2.SetPhase(domain.CegoPhaseExchange)
	he := g2.GetHint()
	require.NotNil(t, he)
	assert.Equal(t, "keep_best", he.Reason)
	assert.NotEmpty(t, he.CardIndices)
}

// --- Additional UnmarshalJSON validation branches ---

func TestCegoUnmarshalMoreErrors(t *testing.T) {
	base := func() map[string]any {
		g := cegoNewReset()
		data, _ := json.Marshal(g)
		var m map[string]any
		_ = json.Unmarshal(data, &m)
		return m
	}
	bad := func(t *testing.T, mutate func(map[string]any)) {
		t.Helper()
		m := base()
		mutate(m)
		data, _ := json.Marshal(m)
		var g domain.Cego
		assert.Error(t, json.Unmarshal(data, &g))
	}

	t.Run("oversized deck", func(t *testing.T) {
		bad(t, func(m map[string]any) { m["dk"] = make([]any, 5001) })
	})
	t.Run("invalid blind card", func(t *testing.T) {
		bad(t, func(m map[string]any) { m["bl"] = []any{nil} })
	})
	t.Run("invalid stash card", func(t *testing.T) {
		bad(t, func(m map[string]any) { m["st"] = []any{nil} })
	})
	t.Run("invalid last-trick card", func(t *testing.T) {
		bad(t, func(m map[string]any) { m["lc"] = []any{nil} })
	})
	t.Run("nil trick card", func(t *testing.T) {
		bad(t, func(m map[string]any) { m["ct"] = []any{nil} })
	})
	t.Run("currentPlayerIdx out of range", func(t *testing.T) {
		bad(t, func(m map[string]any) { m["ci"] = 5 })
	})
	t.Run("dealerIdx out of range", func(t *testing.T) {
		bad(t, func(m map[string]any) { m["di"] = 5 })
	})
	t.Run("bidPlayerIdx out of range", func(t *testing.T) {
		bad(t, func(m map[string]any) { m["bi"] = 9 })
	})
	t.Run("leadPlayerIdx out of range", func(t *testing.T) {
		bad(t, func(m map[string]any) { m["li"] = 9 })
	})
	t.Run("winnerPlayer out of range", func(t *testing.T) {
		bad(t, func(m map[string]any) { m["wp"] = 9 })
	})
	t.Run("outcome out of range", func(t *testing.T) {
		bad(t, func(m map[string]any) { m["oc"] = 9 })
	})
	t.Run("result out of range", func(t *testing.T) {
		bad(t, func(m map[string]any) { m["rs"] = 9 })
	})
	t.Run("contract bid out of range", func(t *testing.T) {
		bad(t, func(m map[string]any) { m["co"] = 9 })
	})
	t.Run("play phase without declarer", func(t *testing.T) {
		bad(t, func(m map[string]any) { m["ph"] = int(domain.CegoPhasePlay) })
	})
	t.Run("invalid config", func(t *testing.T) {
		bad(t, func(m map[string]any) { m["cf"] = map[string]any{"cd": 9, "td": 5} })
	})
}
