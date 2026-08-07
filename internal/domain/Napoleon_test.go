//go:build test

package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestNapoleon() *domain.Napoleon {
	players := []*domain.NapoleonPlayer{
		domain.NewNapoleonPlayer(true),  // human
		domain.NewNapoleonPlayer(false), // CPU 1
		domain.NewNapoleonPlayer(false), // CPU 2
		domain.NewNapoleonPlayer(false), // CPU 3
	}
	cfg := domain.DefaultNapoleonConfig()
	tc := domain.NewTrumpCards(1)
	return domain.NewNapoleon(tc, players, cfg)
}

func TestNapoleon_NewNapoleon(t *testing.T) {
	n := newTestNapoleon()
	assert.Equal(t, domain.NapoleonWinnerUndecided, n.GetWinnerTeam())
	assert.Equal(t, -1, n.GetNapoleonIdx())
	assert.Equal(t, -1, n.GetAdjutantIdx())
	assert.Equal(t, -1, n.GetHighestBidder())
	assert.Equal(t, 0, n.GetRoundNumber())
	assert.False(t, n.GetGameEndFlag())
}

func TestNapoleon_Reset(t *testing.T) {
	n := newTestNapoleon()
	n.Reset()

	assert.Equal(t, domain.NapoleonPhaseBid, n.GetPhase())
	assert.Equal(t, 1, n.GetRoundNumber())
	assert.Equal(t, 0, n.GetTrickNumber())
	assert.Equal(t, 0, n.GetBidPlayerIdx())
	assert.False(t, n.GetGameEndFlag())
	assert.Equal(t, domain.NapoleonWinnerUndecided, n.GetWinnerTeam())
	assert.Equal(t, -1, n.GetNapoleonIdx())
	assert.False(t, n.GetAdjutantRevealed())
	assert.Nil(t, n.GetAdjutantCard())
	assert.NotNil(t, n.GetKitty())
	assert.Len(t, n.GetKitty(), 1)

	// 各プレイヤーに13枚配られている
	for i := 0; i < n.GetPlayerCnt(); i++ {
		assert.Equal(t, domain.NapoleonHandSize, n.GetPlayer(i).GetCardsSize())
	}
}

func TestNapoleon_GetPlayer(t *testing.T) {
	n := newTestNapoleon()
	n.Reset()
	assert.NotNil(t, n.GetPlayer(0))
	assert.NotNil(t, n.GetPlayer(3))
	assert.Nil(t, n.GetPlayer(-1))
	assert.Nil(t, n.GetPlayer(4))
}

func TestNapoleon_PlayerBid(t *testing.T) {
	n := newTestNapoleon()
	n.Reset()

	t.Run("valid bid", func(t *testing.T) {
		n.Reset()
		err := n.PlayerBid(12)
		assert.NoError(t, err)
		assert.Equal(t, 12, n.GetPlayer(0).GetBid())
		assert.Equal(t, 12, n.GetHighestBid())
	})

	t.Run("pass (bid 0)", func(t *testing.T) {
		n.Reset()
		err := n.PlayerBid(0)
		assert.NoError(t, err)
		assert.Equal(t, 0, n.GetPlayer(0).GetBid())
	})

	t.Run("bid below min", func(t *testing.T) {
		n.Reset()
		err := n.PlayerBid(11)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("bid above max", func(t *testing.T) {
		n.Reset()
		err := n.PlayerBid(domain.NapoleonMaxPictureCards + 1)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("bid not higher than current highest", func(t *testing.T) {
		n.Reset()
		n.SetHighestBid(13)
		err := n.PlayerBid(13)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("error when game ended", func(t *testing.T) {
		n.Reset()
		n.SetGameEndFlag(true)
		err := n.PlayerBid(12)
		assert.True(t, errors.Is(err, domain.ErrGameEnded))
	})

	t.Run("error when wrong phase", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhasePlay)
		err := n.PlayerBid(12)
		assert.True(t, errors.Is(err, domain.ErrWrongPhase))
	})

	t.Run("error when not human turn", func(t *testing.T) {
		n.Reset()
		n.SetBidPlayerIdx(1) // CPU turn
		err := n.PlayerBid(12)
		assert.True(t, errors.Is(err, domain.ErrNotHumanTurn))
	})
}

func TestNapoleon_CpuBid(t *testing.T) {
	n := newTestNapoleon()
	n.Reset()

	t.Run("CPU bids when it is their turn", func(t *testing.T) {
		n.Reset()
		n.SetBidPlayerIdx(1)
		n.CpuBid()
		// After CPU bids, bidPlayerIdx advances
		assert.True(t, n.GetBidPlayerIdx() >= 2)
	})

	t.Run("no-op when game ended", func(t *testing.T) {
		n.Reset()
		n.SetBidPlayerIdx(1)
		n.SetPhase(domain.NapoleonPhaseGameEnd)
		n.CpuBid()
		assert.Equal(t, 1, n.GetBidPlayerIdx())
	})

	t.Run("no-op when wrong phase", func(t *testing.T) {
		n.Reset()
		n.SetBidPlayerIdx(1)
		n.SetPhase(domain.NapoleonPhasePlay)
		n.CpuBid()
		assert.Equal(t, 1, n.GetBidPlayerIdx())
	})

	t.Run("no-op when bid already complete", func(t *testing.T) {
		n.Reset()
		n.SetBidPlayerIdx(domain.NapoleonPlayerCnt)
		n.CpuBid()
	})

	t.Run("no-op when human turn", func(t *testing.T) {
		n.Reset()
		n.SetBidPlayerIdx(0) // Human
		n.CpuBid()
		assert.Equal(t, 0, n.GetBidPlayerIdx())
	})
}

func TestNapoleon_AllPass_ForcedBid(t *testing.T) {
	n := newTestNapoleon()
	n.Reset()

	// Human passes
	err := n.PlayerBid(0)
	require.NoError(t, err)

	// All CPUs pass by setting up the state
	for n.GetPhase() == domain.NapoleonPhaseBid {
		n.CpuBid()
	}

	// If all passed, player 0 should be forced Napoleon
	// The phase should move to TrumpDeclaration
	assert.Equal(t, domain.NapoleonPhaseTrumpDeclaration, n.GetPhase())
	assert.True(t, n.GetPlayer(n.GetNapoleonIdx()).GetIsNapoleon())
}

func TestNapoleon_PlayerDeclareTrump(t *testing.T) {
	n := newTestNapoleon()

	setupNapoleonForDeclare := func() {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseTrumpDeclaration)
		n.SetNapoleonIdx(0)
		n.GetPlayer(0).SetIsNapoleon(true)
		n.SetKitty([]*domain.Card{domain.NewCard(domain.CardDesignClover, 5, false)})
	}

	t.Run("valid declaration", func(t *testing.T) {
		setupNapoleonForDeclare()
		err := n.PlayerDeclareTrump(domain.CardDesignSpade, domain.CardDesignHeart, 1)
		assert.NoError(t, err)
		assert.Equal(t, domain.CardDesignSpade, n.GetTrumpSuit())
		assert.Equal(t, domain.NapoleonPhaseKittyExchange, n.GetPhase())
	})

	t.Run("joker as adjutant", func(t *testing.T) {
		setupNapoleonForDeclare()
		err := n.PlayerDeclareTrump(domain.CardDesignHeart, domain.CardDesignJoker, 1)
		assert.NoError(t, err)
		adj := n.GetAdjutantCard()
		assert.Equal(t, domain.CardDesignJoker, adj.GetDesign())
	})

	t.Run("invalid joker value", func(t *testing.T) {
		setupNapoleonForDeclare()
		err := n.PlayerDeclareTrump(domain.CardDesignHeart, domain.CardDesignJoker, 2)
		assert.Error(t, err)
	})

	t.Run("invalid suit", func(t *testing.T) {
		setupNapoleonForDeclare()
		err := n.PlayerDeclareTrump(0, domain.CardDesignHeart, 1) // 0 = Joker suit, invalid as trump
		assert.Error(t, err)
	})

	t.Run("invalid adjutant suit", func(t *testing.T) {
		setupNapoleonForDeclare()
		err := n.PlayerDeclareTrump(domain.CardDesignSpade, 5, 1) // suit 5 invalid
		assert.Error(t, err)
	})

	t.Run("invalid adjutant value", func(t *testing.T) {
		setupNapoleonForDeclare()
		err := n.PlayerDeclareTrump(domain.CardDesignSpade, domain.CardDesignHeart, 0)
		assert.Error(t, err)
	})

	t.Run("adjutant value too high", func(t *testing.T) {
		setupNapoleonForDeclare()
		err := n.PlayerDeclareTrump(domain.CardDesignSpade, domain.CardDesignHeart, 14)
		assert.Error(t, err)
	})

	t.Run("error when game ended", func(t *testing.T) {
		setupNapoleonForDeclare()
		n.SetGameEndFlag(true)
		err := n.PlayerDeclareTrump(domain.CardDesignSpade, domain.CardDesignHeart, 1)
		assert.True(t, errors.Is(err, domain.ErrGameEnded))
	})

	t.Run("error when wrong phase", func(t *testing.T) {
		setupNapoleonForDeclare()
		n.SetPhase(domain.NapoleonPhasePlay)
		err := n.PlayerDeclareTrump(domain.CardDesignSpade, domain.CardDesignHeart, 1)
		assert.True(t, errors.Is(err, domain.ErrWrongPhase))
	})

	t.Run("error when CPU is napoleon", func(t *testing.T) {
		setupNapoleonForDeclare()
		n.SetNapoleonIdx(1)
		err := n.PlayerDeclareTrump(domain.CardDesignSpade, domain.CardDesignHeart, 1)
		assert.True(t, errors.Is(err, domain.ErrNotHumanTurn))
	})

	t.Run("self-named adjutant", func(t *testing.T) {
		setupNapoleonForDeclare()
		// Give human a specific card and name it
		n.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 1, false))
		err := n.PlayerDeclareTrump(domain.CardDesignSpade, domain.CardDesignDiamond, 1)
		assert.NoError(t, err)
		assert.True(t, n.GetAdjutantRevealed())
		assert.Equal(t, 0, n.GetAdjutantIdx())
	})
}

func TestNapoleon_CpuDeclareTrump(t *testing.T) {
	n := newTestNapoleon()
	n.Reset()

	t.Run("CPU declares when it is Napoleon", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseTrumpDeclaration)
		n.SetNapoleonIdx(1)
		n.GetPlayer(1).SetIsNapoleon(true)
		n.SetKitty([]*domain.Card{domain.NewCard(domain.CardDesignClover, 5, false)})
		n.CpuDeclareTrump()
		assert.Equal(t, domain.NapoleonPhaseKittyExchange, n.GetPhase())
	})

	t.Run("no-op when game ended", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseGameEnd)
		n.SetNapoleonIdx(1)
		n.CpuDeclareTrump()
	})

	t.Run("no-op when wrong phase", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhasePlay)
		n.SetNapoleonIdx(1)
		n.CpuDeclareTrump()
	})

	t.Run("no-op when human is napoleon", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseTrumpDeclaration)
		n.SetNapoleonIdx(0)
		n.CpuDeclareTrump()
	})
}

func TestNapoleon_PlayerExchangeKitty(t *testing.T) {
	n := newTestNapoleon()

	setupForExchange := func() {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseKittyExchange)
		n.SetNapoleonIdx(0)
		n.GetPlayer(0).SetIsNapoleon(true)
		n.SetTrumpSuit(domain.CardDesignSpade)
	}

	t.Run("valid exchange", func(t *testing.T) {
		setupForExchange()
		cardsBefore := n.GetPlayer(0).GetCardsSize()
		err := n.PlayerExchangeKitty(0)
		assert.NoError(t, err)
		assert.Equal(t, cardsBefore-1, n.GetPlayer(0).GetCardsSize())
		assert.Equal(t, domain.NapoleonPhasePlay, n.GetPhase())
	})

	t.Run("invalid index negative", func(t *testing.T) {
		setupForExchange()
		err := n.PlayerExchangeKitty(-1)
		assert.Error(t, err)
	})

	t.Run("invalid index too high", func(t *testing.T) {
		setupForExchange()
		err := n.PlayerExchangeKitty(100)
		assert.Error(t, err)
	})

	t.Run("error when game ended", func(t *testing.T) {
		setupForExchange()
		n.SetGameEndFlag(true)
		err := n.PlayerExchangeKitty(0)
		assert.True(t, errors.Is(err, domain.ErrGameEnded))
	})

	t.Run("error when wrong phase", func(t *testing.T) {
		setupForExchange()
		n.SetPhase(domain.NapoleonPhasePlay)
		err := n.PlayerExchangeKitty(0)
		assert.True(t, errors.Is(err, domain.ErrWrongPhase))
	})

	t.Run("error when CPU is napoleon", func(t *testing.T) {
		setupForExchange()
		n.SetNapoleonIdx(1)
		err := n.PlayerExchangeKitty(0)
		assert.True(t, errors.Is(err, domain.ErrNotHumanTurn))
	})
}

func TestNapoleon_CpuExchangeKitty(t *testing.T) {
	n := newTestNapoleon()

	t.Run("CPU exchanges when it is Napoleon", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseKittyExchange)
		n.SetNapoleonIdx(1)
		n.GetPlayer(1).SetIsNapoleon(true)
		n.SetTrumpSuit(domain.CardDesignSpade)
		n.CpuExchangeKitty()
		assert.Equal(t, domain.NapoleonPhasePlay, n.GetPhase())
	})

	t.Run("no-op when game ended", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseGameEnd)
		n.SetNapoleonIdx(1)
		n.CpuExchangeKitty()
	})

	t.Run("no-op when wrong phase", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhasePlay)
		n.SetNapoleonIdx(1)
		n.CpuExchangeKitty()
	})

	t.Run("no-op when human is napoleon", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseKittyExchange)
		n.SetNapoleonIdx(0)
		n.CpuExchangeKitty()
	})
}

func TestNapoleon_PlayerPlay(t *testing.T) {
	n := newTestNapoleon()

	setupForPlay := func() {
		n.Reset()
		n.SetPhase(domain.NapoleonPhasePlay)
		n.SetCurrentPlayerIdx(0)
		n.SetTrickNumber(2) // Not first trick
		n.SetTrumpSuit(domain.CardDesignSpade)
		n.SetNapoleonIdx(0)
	}

	t.Run("valid play on lead", func(t *testing.T) {
		setupForPlay()
		err := n.PlayerPlay(0)
		assert.NoError(t, err)
	})

	t.Run("error when game ended", func(t *testing.T) {
		setupForPlay()
		n.SetGameEndFlag(true)
		err := n.PlayerPlay(0)
		assert.True(t, errors.Is(err, domain.ErrGameEnded))
	})

	t.Run("error when wrong phase", func(t *testing.T) {
		setupForPlay()
		n.SetPhase(domain.NapoleonPhaseBid)
		err := n.PlayerPlay(0)
		assert.True(t, errors.Is(err, domain.ErrWrongPhase))
	})

	t.Run("error when CPU turn", func(t *testing.T) {
		setupForPlay()
		n.SetCurrentPlayerIdx(1)
		err := n.PlayerPlay(0)
		assert.True(t, errors.Is(err, domain.ErrNotHumanTurn))
	})

	t.Run("error when card index out of range", func(t *testing.T) {
		setupForPlay()
		err := n.PlayerPlay(-1)
		assert.Error(t, err)
	})

	t.Run("error when card index too high", func(t *testing.T) {
		setupForPlay()
		err := n.PlayerPlay(100)
		assert.Error(t, err)
	})

	t.Run("follow suit violation", func(t *testing.T) {
		setupForPlay()
		// Set a lead card as heart
		n.SetCurrentTrick([]*domain.NapoleonTrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		})
		// Human has hearts but tries to play a non-heart
		n.GetPlayer(0).Reset()
		n.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
		n.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		err := n.PlayerPlay(0) // clover 3
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("can play off-suit when void", func(t *testing.T) {
		setupForPlay()
		n.SetCurrentTrick([]*domain.NapoleonTrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		})
		n.GetPlayer(0).Reset()
		n.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
		err := n.PlayerPlay(0)
		assert.NoError(t, err)
	})

	t.Run("joker can always be played", func(t *testing.T) {
		setupForPlay()
		n.SetCurrentTrick([]*domain.NapoleonTrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		})
		n.GetPlayer(0).Reset()
		n.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
		n.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		err := n.PlayerPlay(0) // joker
		assert.NoError(t, err)
	})

	t.Run("joker lead allows any follow", func(t *testing.T) {
		setupForPlay()
		n.SetCurrentTrick([]*domain.NapoleonTrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignJoker, 1, false)},
		})
		n.GetPlayer(0).Reset()
		n.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
		err := n.PlayerPlay(0)
		assert.NoError(t, err)
	})
}

func TestNapoleon_CpuPlay(t *testing.T) {
	n := newTestNapoleon()

	t.Run("CPU plays when it is their turn", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhasePlay)
		n.SetCurrentPlayerIdx(1)
		n.SetTrickNumber(1)
		n.SetTrumpSuit(domain.CardDesignSpade)
		cardsBefore := n.GetPlayer(1).GetCardsSize()
		n.CpuPlay()
		assert.Equal(t, cardsBefore-1, n.GetPlayer(1).GetCardsSize())
	})

	t.Run("no-op when game ended", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseGameEnd)
		n.SetCurrentPlayerIdx(1)
		n.CpuPlay()
	})

	t.Run("no-op when wrong phase", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseBid)
		n.SetCurrentPlayerIdx(1)
		n.CpuPlay()
	})

	t.Run("no-op when human turn", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhasePlay)
		n.SetCurrentPlayerIdx(0)
		n.CpuPlay()
	})
}

func TestNapoleon_ResolveTrick(t *testing.T) {
	n := newTestNapoleon()

	setupTrick := func(cards []*domain.NapoleonTrickCard) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseTrickEnd)
		n.SetTrickNumber(1)
		n.SetTrumpSuit(domain.CardDesignSpade)
		n.SetCurrentTrick(cards)
	}

	t.Run("highest lead suit wins", func(t *testing.T) {
		setupTrick([]*domain.NapoleonTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 10, false)},
			{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 3, false)},
			{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 1, false)}, // A = 14 strongest
		})
		n.ResolveTrick()
		assert.Equal(t, 1, n.GetPlayer(3).GetTrickCount())
	})

	t.Run("trump beats lead suit", func(t *testing.T) {
		setupTrick([]*domain.NapoleonTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 2, false)}, // trump
			{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},
			{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignClover, 1, false)},
		})
		n.ResolveTrick()
		assert.Equal(t, 1, n.GetPlayer(1).GetTrickCount())
	})

	t.Run("joker wins", func(t *testing.T) {
		setupTrick([]*domain.NapoleonTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignJoker, 1, false)},
			{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 1, false)}, // trump A
			{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
		})
		n.ResolveTrick()
		assert.Equal(t, 1, n.GetPlayer(1).GetTrickCount())
	})

	t.Run("spade 3 kills joker", func(t *testing.T) {
		setupTrick([]*domain.NapoleonTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignJoker, 1, false)},
			{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 3, false)}, // joker killer
			{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
		})
		n.ResolveTrick()
		assert.Equal(t, 1, n.GetPlayer(2).GetTrickCount())
	})

	t.Run("picture cards counted", func(t *testing.T) {
		setupTrick([]*domain.NapoleonTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 11, false)}, // J
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 12, false)}, // Q
			{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 13, false)}, // K
			{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},  // A (trump, wins)
		})
		n.ResolveTrick()
		assert.Equal(t, 4, n.GetPlayer(3).GetPictureCards()) // J+Q+K+A = 4
	})

	t.Run("no-op when wrong phase", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhasePlay)
		n.ResolveTrick()
	})

	t.Run("no-op when trick incomplete", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseTrickEnd)
		n.SetCurrentTrick([]*domain.NapoleonTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		})
		n.ResolveTrick()
	})

	t.Run("round end after 13 tricks", func(t *testing.T) {
		setupTrick([]*domain.NapoleonTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 10, false)},
			{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 3, false)},
			{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},
		})
		n.SetTrickNumber(domain.NapoleonHandSize)
		n.ResolveTrick()
		assert.Equal(t, domain.NapoleonPhaseRoundEnd, n.GetPhase())
	})

	t.Run("trick end before 13 tricks", func(t *testing.T) {
		setupTrick([]*domain.NapoleonTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 10, false)},
			{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 3, false)},
			{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},
		})
		n.SetTrickNumber(5)
		n.ResolveTrick()
		assert.Equal(t, domain.NapoleonPhaseTrickEnd, n.GetPhase())
	})

	t.Run("trump same suit comparison", func(t *testing.T) {
		setupTrick([]*domain.NapoleonTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 2, false)},  // trump low
			{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 10, false)}, // trump high
			{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},
		})
		n.ResolveTrick()
		assert.Equal(t, 1, n.GetPlayer(2).GetTrickCount())
	})

	t.Run("off-suit non-trump does not win", func(t *testing.T) {
		setupTrick([]*domain.NapoleonTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 1, false)},  // off-suit A
			{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignDiamond, 1, false)}, // off-suit A
			{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 3, false)},
		})
		n.ResolveTrick()
		assert.Equal(t, 1, n.GetPlayer(0).GetTrickCount()) // lead heart 5 wins (only lead suit counted)
	})
}

func TestNapoleon_NextTrick(t *testing.T) {
	n := newTestNapoleon()

	t.Run("advances to play phase", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseTrickEnd)
		n.SetLeadPlayerIdx(2)
		n.SetTrickNumber(3)
		n.NextTrick()
		assert.Equal(t, domain.NapoleonPhasePlay, n.GetPhase())
		assert.Equal(t, 4, n.GetTrickNumber())
		assert.Equal(t, 2, n.GetCurrentPlayerIdx())
		assert.Nil(t, n.GetCurrentTrick())
	})

	t.Run("no-op when wrong phase", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhasePlay)
		n.NextTrick()
	})
}

func TestNapoleon_ScoreRound(t *testing.T) {
	n := newTestNapoleon()

	setupForScoring := func(napoleonPictures, alliedPictures int) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseRoundEnd)
		n.SetHighestBid(12)
		n.SetNapoleonIdx(0)
		n.GetPlayer(0).SetIsNapoleon(true)
		n.SetAdjutantIdx(1)
		n.GetPlayer(1).SetIsAdjutant(true)
		n.GetPlayer(0).SetPictureCards(napoleonPictures - 3) // Napoleon's share
		n.GetPlayer(1).SetPictureCards(3)                    // Adjutant's share
		n.GetPlayer(2).SetPictureCards(alliedPictures / 2)
		n.GetPlayer(3).SetPictureCards(alliedPictures - alliedPictures/2)
	}

	t.Run("napoleon wins", func(t *testing.T) {
		setupForScoring(12, 5) // Napoleon team: 9+3=12, Allied: 5
		n.ScoreRound()
		assert.Equal(t, domain.NapoleonWinnerNapoleon, n.GetWinnerTeam())
		assert.Equal(t, 12, n.GetPlayer(0).GetRoundScore())  // Napoleon +bid
		assert.Equal(t, 6, n.GetPlayer(1).GetRoundScore())   // Adjutant +bid/2
		assert.Equal(t, -12, n.GetPlayer(2).GetRoundScore()) // Allied -bid
	})

	t.Run("allied wins", func(t *testing.T) {
		setupForScoring(8, 9) // Napoleon team: 5+3=8, Allied: 9
		n.ScoreRound()
		assert.Equal(t, domain.NapoleonWinnerAllied, n.GetWinnerTeam())
		assert.Equal(t, -24, n.GetPlayer(0).GetRoundScore()) // Napoleon -bid*2
		assert.Equal(t, -12, n.GetPlayer(1).GetRoundScore()) // Adjutant -bid
		assert.Equal(t, 12, n.GetPlayer(2).GetRoundScore())  // Allied +bid
	})

	t.Run("no-op when wrong phase", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhasePlay)
		n.ScoreRound()
	})

	t.Run("game end when point limit reached", func(t *testing.T) {
		setupForScoring(12, 5)
		cfg := n.GetConfig()
		cfg.PointLimit = 10
		n.SetConfig(cfg)
		n.ScoreRound()
		assert.True(t, n.GetGameEndFlag())
		assert.Equal(t, domain.NapoleonPhaseGameEnd, n.GetPhase())
	})

	t.Run("game end when negative point limit reached", func(t *testing.T) {
		setupForScoring(8, 9) // Allied wins, Napoleon gets -24
		cfg := n.GetConfig()
		cfg.PointLimit = 20
		n.SetConfig(cfg)
		n.ScoreRound()
		assert.True(t, n.GetGameEndFlag())
	})
}

func TestNapoleon_NextRound(t *testing.T) {
	n := newTestNapoleon()

	t.Run("advances to next round", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseRoundEnd)
		r := n.GetRoundNumber()
		n.NextRound()
		assert.Equal(t, r+1, n.GetRoundNumber())
		assert.Equal(t, domain.NapoleonPhaseBid, n.GetPhase())
		assert.Equal(t, 0, n.GetBidPlayerIdx())
		assert.Equal(t, -1, n.GetNapoleonIdx())
	})

	t.Run("no-op when wrong phase", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhasePlay)
		n.NextRound()
	})
}

func TestNapoleon_IsHumanTurn(t *testing.T) {
	n := newTestNapoleon()
	n.Reset()

	n.SetCurrentPlayerIdx(0)
	assert.True(t, n.IsHumanTurn())

	n.SetCurrentPlayerIdx(1)
	assert.False(t, n.IsHumanTurn())

	n.SetCurrentPlayerIdx(-1)
	assert.False(t, n.IsHumanTurn())

	n.SetCurrentPlayerIdx(100)
	assert.False(t, n.IsHumanTurn())
}

func TestNapoleon_IsHumanBidTurn(t *testing.T) {
	n := newTestNapoleon()
	n.Reset()

	n.SetBidPlayerIdx(0)
	assert.True(t, n.IsHumanBidTurn())

	n.SetBidPlayerIdx(1)
	assert.False(t, n.IsHumanBidTurn())

	n.SetBidPlayerIdx(-1)
	assert.False(t, n.IsHumanBidTurn())

	n.SetBidPlayerIdx(100)
	assert.False(t, n.IsHumanBidTurn())
}

func TestNapoleon_IsHumanDeclareTurn(t *testing.T) {
	n := newTestNapoleon()
	n.Reset()

	n.SetNapoleonIdx(0)
	assert.True(t, n.IsHumanDeclareTurn())

	n.SetNapoleonIdx(1)
	assert.False(t, n.IsHumanDeclareTurn())

	n.SetNapoleonIdx(-1)
	assert.False(t, n.IsHumanDeclareTurn())

	n.SetNapoleonIdx(100)
	assert.False(t, n.IsHumanDeclareTurn())
}

func TestNapoleon_IsHumanExchangeTurn(t *testing.T) {
	n := newTestNapoleon()
	n.Reset()

	n.SetNapoleonIdx(0)
	assert.True(t, n.IsHumanExchangeTurn())

	n.SetNapoleonIdx(1)
	assert.False(t, n.IsHumanExchangeTurn())
}

func TestNapoleon_GetSetConfig(t *testing.T) {
	n := newTestNapoleon()
	cfg := domain.NapoleonConfig{
		CpuDifficulty: domain.NapoleonCpuDifficultyHard,
		MinBid:        13,
		PointLimit:    200,
	}
	n.SetConfig(cfg)
	assert.Equal(t, cfg, n.GetConfig())
}

func TestNapoleon_GetActionLog(t *testing.T) {
	n := newTestNapoleon()
	n.Reset()
	assert.Nil(t, n.GetActionLog())

	// Trigger some actions to generate logs
	_ = n.PlayerBid(12)
	assert.True(t, len(n.GetActionLog()) > 0)
}

func TestNapoleon_GetHint(t *testing.T) {
	n := newTestNapoleon()

	t.Run("bid phase hint", func(t *testing.T) {
		n.Reset()
		n.SetBidPlayerIdx(0)
		hint := n.GetHint()
		assert.NotNil(t, hint)
		assert.NotNil(t, hint.Bid)
		assert.Equal(t, "strategic_bid", hint.Reason)
	})

	t.Run("play phase hint", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhasePlay)
		n.SetCurrentPlayerIdx(0)
		n.SetTrickNumber(2)
		n.SetTrumpSuit(domain.CardDesignSpade)
		hint := n.GetHint()
		assert.NotNil(t, hint)
		assert.NotNil(t, hint.CardIndex)
	})

	t.Run("declare phase hint", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseTrumpDeclaration)
		n.SetNapoleonIdx(0)
		hint := n.GetHint()
		assert.NotNil(t, hint)
		assert.NotNil(t, hint.TrumpSuit)
		assert.NotNil(t, hint.AdjutantSuit)
		assert.NotNil(t, hint.AdjutantValue)
		assert.Equal(t, "strategic_declare", hint.Reason)
	})

	t.Run("exchange phase hint", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseKittyExchange)
		n.SetNapoleonIdx(0)
		n.SetTrumpSuit(domain.CardDesignSpade)
		hint := n.GetHint()
		assert.NotNil(t, hint)
		assert.NotNil(t, hint.DiscardIndex)
		assert.Equal(t, "strategic_discard", hint.Reason)
	})

	t.Run("no hint for non-human phases", func(t *testing.T) {
		n.Reset()
		n.SetBidPlayerIdx(1) // CPU turn
		hint := n.GetHint()
		assert.Nil(t, hint)
	})

	t.Run("no hint for game end", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseGameEnd)
		hint := n.GetHint()
		assert.Nil(t, hint)
	})

	t.Run("play phase hint not human turn", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhasePlay)
		n.SetCurrentPlayerIdx(1)
		hint := n.GetHint()
		assert.Nil(t, hint)
	})

	t.Run("declare phase hint not human napoleon", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseTrumpDeclaration)
		n.SetNapoleonIdx(1)
		hint := n.GetHint()
		assert.Nil(t, hint)
	})

	t.Run("exchange phase hint not human napoleon", func(t *testing.T) {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseKittyExchange)
		n.SetNapoleonIdx(1)
		hint := n.GetHint()
		assert.Nil(t, hint)
	})
}

func TestNapoleon_GetValidPlayIndices(t *testing.T) {
	n := newTestNapoleon()
	n.Reset()
	n.SetPhase(domain.NapoleonPhasePlay)
	n.SetCurrentPlayerIdx(0)
	n.SetTrickNumber(2)
	n.SetTrumpSuit(domain.CardDesignSpade)

	indices := n.GetValidPlayIndices(0)
	// On lead, all cards are valid
	assert.Equal(t, n.GetPlayer(0).GetCardsSize(), len(indices))
}

func TestNapoleon_CpuDifficulty_Bid(t *testing.T) {
	for _, diff := range []domain.NapoleonCpuDifficulty{
		domain.NapoleonCpuDifficultyEasy,
		domain.NapoleonCpuDifficultyNormal,
		domain.NapoleonCpuDifficultyHard,
	} {
		t.Run("difficulty_"+string(rune('0'+int(diff))), func(t *testing.T) {
			n := newTestNapoleon()
			cfg := domain.DefaultNapoleonConfig()
			cfg.CpuDifficulty = diff
			n.SetConfig(cfg)
			n.Reset()
			n.SetBidPlayerIdx(1)
			n.CpuBid()
			// Just verifying no panic
		})
	}
}

func TestNapoleon_CpuDifficulty_Play(t *testing.T) {
	for _, diff := range []domain.NapoleonCpuDifficulty{
		domain.NapoleonCpuDifficultyEasy,
		domain.NapoleonCpuDifficultyNormal,
		domain.NapoleonCpuDifficultyHard,
	} {
		t.Run("difficulty_"+string(rune('0'+int(diff))), func(t *testing.T) {
			n := newTestNapoleon()
			cfg := domain.DefaultNapoleonConfig()
			cfg.CpuDifficulty = diff
			n.SetConfig(cfg)
			n.Reset()
			n.SetPhase(domain.NapoleonPhasePlay)
			n.SetCurrentPlayerIdx(1)
			n.SetTrickNumber(1)
			n.SetTrumpSuit(domain.CardDesignSpade)
			n.CpuPlay()
			// No panic
		})
	}
}

func TestNapoleon_CpuDifficulty_Declare(t *testing.T) {
	for _, diff := range []domain.NapoleonCpuDifficulty{
		domain.NapoleonCpuDifficultyEasy,
		domain.NapoleonCpuDifficultyNormal,
		domain.NapoleonCpuDifficultyHard,
	} {
		t.Run("difficulty_"+string(rune('0'+int(diff))), func(t *testing.T) {
			n := newTestNapoleon()
			cfg := domain.DefaultNapoleonConfig()
			cfg.CpuDifficulty = diff
			n.SetConfig(cfg)
			n.Reset()
			n.SetPhase(domain.NapoleonPhaseTrumpDeclaration)
			n.SetNapoleonIdx(1)
			n.GetPlayer(1).SetIsNapoleon(true)
			n.SetKitty([]*domain.Card{domain.NewCard(domain.CardDesignClover, 5, false)})
			n.CpuDeclareTrump()
			assert.Equal(t, domain.NapoleonPhaseKittyExchange, n.GetPhase())
		})
	}
}

func TestNapoleon_CpuDifficulty_Exchange(t *testing.T) {
	for _, diff := range []domain.NapoleonCpuDifficulty{
		domain.NapoleonCpuDifficultyEasy,
		domain.NapoleonCpuDifficultyNormal,
		domain.NapoleonCpuDifficultyHard,
	} {
		t.Run("difficulty_"+string(rune('0'+int(diff))), func(t *testing.T) {
			n := newTestNapoleon()
			cfg := domain.DefaultNapoleonConfig()
			cfg.CpuDifficulty = diff
			n.SetConfig(cfg)
			n.Reset()
			n.SetPhase(domain.NapoleonPhaseKittyExchange)
			n.SetNapoleonIdx(1)
			n.GetPlayer(1).SetIsNapoleon(true)
			n.SetTrumpSuit(domain.CardDesignSpade)
			n.CpuExchangeKitty()
			assert.Equal(t, domain.NapoleonPhasePlay, n.GetPhase())
		})
	}
}

func TestNapoleon_FullGameFlow(t *testing.T) {
	// Test a complete game from Reset through bidding, declaration, exchange, play, and scoring
	n := newTestNapoleon()
	cfg := domain.DefaultNapoleonConfig()
	cfg.CpuDifficulty = domain.NapoleonCpuDifficultyEasy
	cfg.PointLimit = 1 // End game after 1 round
	n.SetConfig(cfg)
	n.Reset()

	// Bid phase
	assert.Equal(t, domain.NapoleonPhaseBid, n.GetPhase())
	err := n.PlayerBid(12)
	require.NoError(t, err)

	// CPU bids
	for n.GetPhase() == domain.NapoleonPhaseBid {
		n.CpuBid()
	}

	// Should be either TrumpDeclaration or have a napoleon
	if n.IsHumanDeclareTurn() {
		// Human is napoleon
		err = n.PlayerDeclareTrump(domain.CardDesignSpade, domain.CardDesignHeart, 1)
		require.NoError(t, err)
		assert.Equal(t, domain.NapoleonPhaseKittyExchange, n.GetPhase())

		err = n.PlayerExchangeKitty(0)
		require.NoError(t, err)
	} else {
		// CPU is napoleon
		n.CpuDeclareTrump()
		assert.Equal(t, domain.NapoleonPhaseKittyExchange, n.GetPhase())
		n.CpuExchangeKitty()
	}

	assert.Equal(t, domain.NapoleonPhasePlay, n.GetPhase())

	// Play all 13 tricks
	for range domain.NapoleonHandSize {
		for n.GetPhase() == domain.NapoleonPhasePlay {
			if n.IsHumanTurn() {
				valid := n.GetValidPlayIndices(0)
				require.True(t, len(valid) > 0, "human has no valid plays")
				err = n.PlayerPlay(valid[0])
				require.NoError(t, err)
			} else {
				n.CpuPlay()
			}
		}

		assert.True(t, n.GetPhase() == domain.NapoleonPhaseTrickEnd || n.GetPhase() == domain.NapoleonPhaseRoundEnd)
		n.ResolveTrick()

		if n.GetPhase() == domain.NapoleonPhaseRoundEnd {
			break
		}
		n.NextTrick()
	}

	// Score round
	assert.Equal(t, domain.NapoleonPhaseRoundEnd, n.GetPhase())
	n.ScoreRound()

	// Game should end (PointLimit=1)
	assert.True(t, n.GetGameEndFlag())
	assert.Equal(t, domain.NapoleonPhaseGameEnd, n.GetPhase())
}

func TestNapoleon_AdjutantRevealDuringPlay(t *testing.T) {
	n := newTestNapoleon()
	n.Reset()
	n.SetPhase(domain.NapoleonPhasePlay)
	n.SetCurrentPlayerIdx(1)
	n.SetTrickNumber(2)
	n.SetTrumpSuit(domain.CardDesignSpade)
	n.SetNapoleonIdx(0)
	n.GetPlayer(0).SetIsNapoleon(true)

	// Set adjutant card as Heart A
	adjCard := domain.NewCard(domain.CardDesignHeart, 1, false)
	n.SetAdjutantCard(adjCard)
	// Player 2 holds the adjutant card
	n.SetAdjutantIdx(2)
	n.GetPlayer(2).SetIsAdjutant(true)

	// Give player 1 the heart A to play (simulating adjutant reassignment after kitty exchange)
	n.GetPlayer(1).Reset()
	n.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

	n.CpuPlay()

	// Now adjutant should be revealed
	assert.True(t, n.GetAdjutantRevealed())
	assert.True(t, n.GetPlayer(1).GetAdjutantRevealed())
}

func TestNapoleon_PlayCard_AdvancesPlayer(t *testing.T) {
	n := newTestNapoleon()
	n.Reset()
	n.SetPhase(domain.NapoleonPhasePlay)
	n.SetCurrentPlayerIdx(0)
	n.SetTrickNumber(2)
	n.SetTrumpSuit(domain.CardDesignSpade)

	err := n.PlayerPlay(0)
	require.NoError(t, err)
	assert.Equal(t, 1, n.GetCurrentPlayerIdx())             // Advanced to next player
	assert.Equal(t, domain.NapoleonPhasePlay, n.GetPhase()) // Still play phase
}

func TestNapoleon_PlayCard_TrickEnd(t *testing.T) {
	n := newTestNapoleon()
	n.Reset()
	n.SetPhase(domain.NapoleonPhasePlay)
	n.SetCurrentPlayerIdx(0)
	n.SetTrickNumber(2)
	n.SetTrumpSuit(domain.CardDesignSpade)

	// Set 3 cards already in trick (heart lead)
	n.SetCurrentTrick([]*domain.NapoleonTrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 6, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 7, false)},
	})

	// Give human a heart to satisfy follow suit
	n.GetPlayer(0).Reset()
	n.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	err := n.PlayerPlay(0)
	require.NoError(t, err)
	assert.Equal(t, domain.NapoleonPhaseTrickEnd, n.GetPhase())
}

func TestNapoleon_TrickWinner_EmptyTrick(t *testing.T) {
	n := newTestNapoleon()
	n.Reset()
	n.SetPhase(domain.NapoleonPhaseTrickEnd)
	n.SetTrickNumber(1)
	n.SetTrumpSuit(domain.CardDesignSpade)
	n.SetCurrentTrick(nil) // empty
	n.ResolveTrick()       // no-op because trick incomplete
}

// **人間の席はコンストラクタ次第 (#4689)。**CUI の表示が GetPlayer(0) を
// 決め打ちしていた件で公開したアクセサ。並びを変えても正しく解決すること。
func TestNapoleon_GetHumanIdx(t *testing.T) {
	t.Run("default layout puts the human first", func(t *testing.T) {
		n := domain.NewDefaultNapoleon()
		n.Reset()
		if got := n.GetHumanIdx(); got != 0 {
			t.Errorf("GetHumanIdx() = %d, want 0", got)
		}
	})

	// **席順を変えて踏む。**既定配置だけ試しても「0 を返すだけ」の実装と
	// 区別が付かない。
	t.Run("finds the human at a later seat", func(t *testing.T) {
		players := []*domain.NapoleonPlayer{
			domain.NewNapoleonPlayer(false),
			domain.NewNapoleonPlayer(false),
			domain.NewNapoleonPlayer(true),
			domain.NewNapoleonPlayer(false),
		}
		n := domain.NewNapoleon(domain.NewTrumpCards(1), players, domain.DefaultNapoleonConfig())
		if got := n.GetHumanIdx(); got != 2 {
			t.Errorf("GetHumanIdx() = %d, want 2", got)
		}
	})

	t.Run("reports -1 when no player is human", func(t *testing.T) {
		players := []*domain.NapoleonPlayer{
			domain.NewNapoleonPlayer(false),
			domain.NewNapoleonPlayer(false),
			domain.NewNapoleonPlayer(false),
			domain.NewNapoleonPlayer(false),
		}
		n := domain.NewNapoleon(domain.NewTrumpCards(1), players, domain.DefaultNapoleonConfig())
		if got := n.GetHumanIdx(); got != -1 {
			t.Errorf("GetHumanIdx() = %d, want -1", got)
		}
	})
}
