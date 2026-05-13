//go:build test

package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- Test helpers ---

func newTestMighty() *domain.Mighty {
	players := []*domain.MightyPlayer{
		domain.NewMightyPlayer(true),  // human
		domain.NewMightyPlayer(false), // CPU 1
		domain.NewMightyPlayer(false), // CPU 2
		domain.NewMightyPlayer(false), // CPU 3
		domain.NewMightyPlayer(false), // CPU 4
	}
	return domain.NewMighty(domain.NewTrumpCards(1), players, domain.DefaultMightyConfig())
}

// replaceHand resets a player's hand and adds the given cards in order.
func replaceHand(p *domain.MightyPlayer, cards ...*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// card is a brief constructor used to keep tables compact.
func mightyCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

// --- A. Construction and Reset ---

func TestNewDefaultMighty(t *testing.T) {
	m := domain.NewDefaultMighty()
	assert.NotNil(t, m)
	assert.Equal(t, domain.MightyPlayerCnt, m.GetPlayerCnt())
	assert.Equal(t, domain.MightyWinnerUndecided, m.GetWinnerTeam())
	assert.Equal(t, -1, m.GetDeclarerIdx())
	assert.Equal(t, -1, m.GetPartnerIdx())
	assert.Equal(t, -1, m.GetHighestBidder())
	assert.Equal(t, domain.MightyTrumpNone, m.GetTrumpSuit())

	humanCount := 0
	for _, p := range m.GetPlayers() {
		if p.GetIsHuman() {
			humanCount++
		}
	}
	assert.Equal(t, 1, humanCount)
}

func TestMighty_Reset_dealCards(t *testing.T) {
	m := newTestMighty()
	m.Reset()

	assert.Equal(t, domain.MightyPhaseBid, m.GetPhase())
	assert.Equal(t, 1, m.GetRoundNumber())
	assert.Len(t, m.GetKitty(), domain.MightyKittySize)

	seen := map[[2]int]bool{}
	total := 0
	for i := 0; i < m.GetPlayerCnt(); i++ {
		p := m.GetPlayer(i)
		assert.Equal(t, domain.MightyHandSize, p.GetCardsSize(), "player %d hand size", i)
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			key := [2]int{c.GetDesign(), c.GetValue()}
			assert.False(t, seen[key], "duplicate card %v in player %d", key, i)
			seen[key] = true
			total++
		}
	}
	for _, c := range m.GetKitty() {
		key := [2]int{c.GetDesign(), c.GetValue()}
		assert.False(t, seen[key], "kitty card duplicates a hand card")
		seen[key] = true
		total++
	}
	assert.Equal(t, domain.MightyTotalCards, total)
}

func TestMighty_GetPlayer_outOfRange(t *testing.T) {
	m := newTestMighty()
	m.Reset()
	assert.NotNil(t, m.GetPlayer(0))
	assert.NotNil(t, m.GetPlayer(domain.MightyPlayerCnt-1))
	assert.Nil(t, m.GetPlayer(-1))
	assert.Nil(t, m.GetPlayer(domain.MightyPlayerCnt))
}

func TestMighty_NextRound(t *testing.T) {
	m := newTestMighty()
	m.Reset()

	t.Run("no-op when wrong phase", func(t *testing.T) {
		m.Reset()
		startRound := m.GetRoundNumber()
		m.SetPhase(domain.MightyPhasePlay)
		m.NextRound()
		assert.Equal(t, startRound, m.GetRoundNumber())
	})

	t.Run("advances when in RoundEnd phase", func(t *testing.T) {
		m.Reset()
		// Mutate per-player state so we can confirm ResetRound was called.
		m.GetPlayer(0).SetIsDeclarer(true)
		m.GetPlayer(1).SetIsPartner(true)
		m.GetPlayer(2).SetPointCards(5)
		startRound := m.GetRoundNumber()
		m.SetPhase(domain.MightyPhaseRoundEnd)
		m.NextRound()

		assert.Equal(t, startRound+1, m.GetRoundNumber())
		assert.Equal(t, domain.MightyPhaseBid, m.GetPhase())
		assert.Equal(t, -1, m.GetDeclarerIdx())
		assert.Equal(t, -1, m.GetPartnerIdx())
		assert.Equal(t, -1, m.GetHighestBidder())
		assert.Equal(t, domain.MightyTrumpNone, m.GetTrumpSuit())
		for i := 0; i < m.GetPlayerCnt(); i++ {
			assert.False(t, m.GetPlayer(i).GetIsDeclarer(), "player %d declarer", i)
			assert.False(t, m.GetPlayer(i).GetIsPartner(), "player %d partner", i)
			assert.Equal(t, 0, m.GetPlayer(i).GetPointCards(), "player %d point cards", i)
			assert.Equal(t, domain.MightyHandSize, m.GetPlayer(i).GetCardsSize())
		}
		assert.Len(t, m.GetKitty(), domain.MightyKittySize)
	})
}

// --- B. Bid ---

func TestMighty_PlayerBid_validations(t *testing.T) {
	m := newTestMighty()

	t.Run("error when game ended", func(t *testing.T) {
		m.Reset()
		m.SetGameEndFlag(true)
		err := m.PlayerBid(13, false)
		assert.True(t, errors.Is(err, domain.ErrGameEnded))
	})

	t.Run("error when wrong phase", func(t *testing.T) {
		m.Reset()
		m.SetPhase(domain.MightyPhasePlay)
		err := m.PlayerBid(13, false)
		assert.True(t, errors.Is(err, domain.ErrWrongPhase))
	})

	t.Run("error when not human turn", func(t *testing.T) {
		m.Reset()
		m.SetBidPlayerIdx(1) // CPU turn
		err := m.PlayerBid(13, false)
		assert.True(t, errors.Is(err, domain.ErrNotHumanTurn))
	})

	t.Run("bid below MinBid", func(t *testing.T) {
		m.Reset()
		err := m.PlayerBid(m.GetConfig().MinBid-1, false)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("bid above MightyMaxPoints", func(t *testing.T) {
		m.Reset()
		err := m.PlayerBid(domain.MightyMaxPoints+1, false)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("bid not higher than highestBid", func(t *testing.T) {
		m.Reset()
		m.SetHighestBid(15)
		err := m.PlayerBid(15, false)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("no-trump bid below MinBid + NoTrumpExtra", func(t *testing.T) {
		m.Reset()
		cfg := m.GetConfig()
		err := m.PlayerBid(cfg.MinBid+cfg.NoTrumpExtra-1, true)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("valid no-trump bid at the minimum threshold", func(t *testing.T) {
		m.Reset()
		cfg := m.GetConfig()
		err := m.PlayerBid(cfg.MinBid+cfg.NoTrumpExtra, true)
		assert.NoError(t, err)
		assert.True(t, m.GetWinningBidNoTrump())
	})
}

func TestMighty_PlayerBid_pass(t *testing.T) {
	m := newTestMighty()
	m.Reset()
	prev := m.GetPassCount()
	err := m.PlayerBid(0, true) // noTrump flag ignored on pass
	require.NoError(t, err)
	assert.Equal(t, 0, m.GetPlayer(0).GetBid())
	assert.False(t, m.GetPlayer(0).GetBidNoTrump())
	assert.Equal(t, prev+1, m.GetPassCount())
}

func TestMighty_PlayerBid_raise(t *testing.T) {
	m := newTestMighty()
	m.Reset()
	err := m.PlayerBid(15, false)
	require.NoError(t, err)
	assert.Equal(t, 15, m.GetHighestBid())
	assert.Equal(t, 0, m.GetHighestBidder())
	assert.False(t, m.GetWinningBidNoTrump())
}

func TestMighty_Bid_allPass_forcedDeclarer(t *testing.T) {
	m := newTestMighty()
	m.Reset()
	// Empty every CPU hand so hand strength is 0 → they always pass.
	for i := 1; i < m.GetPlayerCnt(); i++ {
		m.GetPlayer(i).Reset()
	}
	require.NoError(t, m.PlayerBid(0, false))
	for m.GetPhase() == domain.MightyPhaseBid {
		m.CpuBid()
	}
	// All pass → player 0 forced declarer at MinBid.
	assert.Equal(t, domain.MightyPhaseTrumpAndFriend, m.GetPhase())
	assert.Equal(t, 0, m.GetDeclarerIdx())
	assert.Equal(t, m.GetConfig().MinBid, m.GetHighestBid())
	assert.True(t, m.GetPlayer(0).GetIsDeclarer())
}

func TestMighty_Bid_phaseAdvances_afterAllBidsTaken(t *testing.T) {
	m := newTestMighty()
	m.Reset()
	// Bid the maximum possible value so no CPU can outbid us.
	require.NoError(t, m.PlayerBid(domain.MightyMaxPoints, false))
	for m.GetPhase() == domain.MightyPhaseBid {
		m.CpuBid()
	}
	assert.Equal(t, domain.MightyPhaseTrumpAndFriend, m.GetPhase())
	assert.Equal(t, 0, m.GetHighestBidder())
	assert.True(t, m.GetPlayer(0).GetIsDeclarer())
}

func TestMighty_CpuBid_guards(t *testing.T) {
	m := newTestMighty()

	t.Run("no-op when game ended", func(t *testing.T) {
		m.Reset()
		m.SetBidPlayerIdx(1)
		m.SetGameEndFlag(true)
		m.CpuBid()
		assert.Equal(t, 1, m.GetBidPlayerIdx())
	})

	t.Run("no-op when wrong phase", func(t *testing.T) {
		m.Reset()
		m.SetBidPlayerIdx(1)
		m.SetPhase(domain.MightyPhasePlay)
		m.CpuBid()
		assert.Equal(t, 1, m.GetBidPlayerIdx())
	})

	t.Run("no-op when bid index out of range", func(t *testing.T) {
		m.Reset()
		m.SetBidPlayerIdx(domain.MightyPlayerCnt)
		m.CpuBid()
	})

	t.Run("no-op when human is next bidder", func(t *testing.T) {
		m.Reset()
		m.SetBidPlayerIdx(0)
		m.CpuBid()
		assert.Equal(t, 0, m.GetBidPlayerIdx())
	})
}

// --- C. Declare trump and friend ---

func setupForDeclare(t *testing.T) *domain.Mighty {
	t.Helper()
	m := newTestMighty()
	m.Reset()
	m.SetPhase(domain.MightyPhaseTrumpAndFriend)
	m.SetDeclarerIdx(0)
	m.SetHighestBid(15)
	m.GetPlayer(0).SetIsDeclarer(true)
	return m
}

func TestMighty_DeclareTrumpAndFriend_validations(t *testing.T) {
	t.Run("error when game ended", func(t *testing.T) {
		m := setupForDeclare(t)
		m.SetGameEndFlag(true)
		err := m.PlayerDeclareTrumpAndFriend(domain.CardDesignSpade, domain.CardDesignHeart, 1)
		assert.True(t, errors.Is(err, domain.ErrGameEnded))
	})

	t.Run("error when wrong phase", func(t *testing.T) {
		m := setupForDeclare(t)
		m.SetPhase(domain.MightyPhasePlay)
		err := m.PlayerDeclareTrumpAndFriend(domain.CardDesignSpade, domain.CardDesignHeart, 1)
		assert.True(t, errors.Is(err, domain.ErrWrongPhase))
	})

	t.Run("error when declarer is CPU", func(t *testing.T) {
		m := setupForDeclare(t)
		m.SetDeclarerIdx(1)
		err := m.PlayerDeclareTrumpAndFriend(domain.CardDesignSpade, domain.CardDesignHeart, 1)
		assert.True(t, errors.Is(err, domain.ErrNotHumanTurn))
	})

	t.Run("invalid trump suit (Joker)", func(t *testing.T) {
		m := setupForDeclare(t)
		err := m.PlayerDeclareTrumpAndFriend(domain.CardDesignJoker, domain.CardDesignHeart, 1)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("invalid partner suit", func(t *testing.T) {
		m := setupForDeclare(t)
		err := m.PlayerDeclareTrumpAndFriend(domain.CardDesignSpade, domain.CardDesignJoker+99, 1)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("invalid partner value (0)", func(t *testing.T) {
		m := setupForDeclare(t)
		err := m.PlayerDeclareTrumpAndFriend(domain.CardDesignSpade, domain.CardDesignHeart, 0)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("invalid partner value (>13)", func(t *testing.T) {
		m := setupForDeclare(t)
		err := m.PlayerDeclareTrumpAndFriend(domain.CardDesignSpade, domain.CardDesignHeart, domain.CardValueMax+1)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("joker as partner requires value 1", func(t *testing.T) {
		m := setupForDeclare(t)
		err := m.PlayerDeclareTrumpAndFriend(domain.CardDesignSpade, domain.CardDesignJoker, 2)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("joker as partner value 1 is valid", func(t *testing.T) {
		m := setupForDeclare(t)
		err := m.PlayerDeclareTrumpAndFriend(domain.CardDesignSpade, domain.CardDesignJoker, 1)
		assert.NoError(t, err)
	})
}

func TestMighty_DeclareTrumpAndFriend_noTrumpRequiresNoneSentinel(t *testing.T) {
	m := setupForDeclare(t)
	m.SetWinningBidNoTrump(true)

	t.Run("rejects a normal suit when no-trump", func(t *testing.T) {
		err := m.PlayerDeclareTrumpAndFriend(domain.CardDesignSpade, domain.CardDesignHeart, 1)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("accepts MightyTrumpNone when no-trump", func(t *testing.T) {
		m2 := setupForDeclare(t)
		m2.SetWinningBidNoTrump(true)
		err := m2.PlayerDeclareTrumpAndFriend(domain.MightyTrumpNone, domain.CardDesignHeart, 1)
		assert.NoError(t, err)
		assert.Equal(t, domain.MightyTrumpNone, m2.GetTrumpSuit())
	})
}

func TestMighty_DeclareTrumpAndFriend_selfFriend(t *testing.T) {
	m := setupForDeclare(t)
	// Put ♥A into declarer's hand so they themselves are partner.
	replaceHand(m.GetPlayer(0), mightyCard(domain.CardDesignHeart, 1))
	m.SetKitty([]*domain.Card{
		mightyCard(domain.CardDesignClover, 2),
		mightyCard(domain.CardDesignClover, 3),
		mightyCard(domain.CardDesignClover, 4),
	})
	err := m.PlayerDeclareTrumpAndFriend(domain.CardDesignSpade, domain.CardDesignHeart, 1)
	require.NoError(t, err)
	assert.Equal(t, 0, m.GetPartnerIdx())
	assert.True(t, m.GetPartnerRevealed())
	assert.True(t, m.GetPlayer(0).GetIsPartner())
}

func TestMighty_DeclareTrumpAndFriend_partnerHidden(t *testing.T) {
	m := setupForDeclare(t)
	// Strip every player's hand so the partner card lives in player 2 only.
	for i := 0; i < m.GetPlayerCnt(); i++ {
		m.GetPlayer(i).Reset()
	}
	m.GetPlayer(0).AddCard(mightyCard(domain.CardDesignSpade, 5))
	m.GetPlayer(2).AddCard(mightyCard(domain.CardDesignHeart, 1))
	m.SetKitty([]*domain.Card{
		mightyCard(domain.CardDesignClover, 2),
		mightyCard(domain.CardDesignClover, 3),
		mightyCard(domain.CardDesignClover, 4),
	})
	err := m.PlayerDeclareTrumpAndFriend(domain.CardDesignSpade, domain.CardDesignHeart, 1)
	require.NoError(t, err)
	assert.Equal(t, 2, m.GetPartnerIdx())
	assert.False(t, m.GetPartnerRevealed())
	assert.True(t, m.GetPlayer(2).GetIsPartner())
}

func TestMighty_DeclareTrumpAndFriend_advancesToKittyExchange(t *testing.T) {
	m := setupForDeclare(t)
	err := m.PlayerDeclareTrumpAndFriend(domain.CardDesignSpade, domain.CardDesignHeart, 13)
	require.NoError(t, err)
	assert.Equal(t, domain.MightyPhaseKittyExchange, m.GetPhase())
}

func TestMighty_CpuDeclareTrumpAndFriend_guards(t *testing.T) {
	t.Run("CPU declares when CPU is declarer", func(t *testing.T) {
		m := newTestMighty()
		m.Reset()
		m.SetPhase(domain.MightyPhaseTrumpAndFriend)
		m.SetDeclarerIdx(1)
		m.SetHighestBid(15)
		m.GetPlayer(1).SetIsDeclarer(true)
		m.CpuDeclareTrumpAndFriend()
		assert.Equal(t, domain.MightyPhaseKittyExchange, m.GetPhase())
	})

	t.Run("no-op when game ended", func(t *testing.T) {
		m := newTestMighty()
		m.Reset()
		m.SetPhase(domain.MightyPhaseTrumpAndFriend)
		m.SetDeclarerIdx(1)
		m.SetGameEndFlag(true)
		m.CpuDeclareTrumpAndFriend()
		assert.Equal(t, domain.MightyPhaseTrumpAndFriend, m.GetPhase())
	})

	t.Run("no-op when human is declarer", func(t *testing.T) {
		m := newTestMighty()
		m.Reset()
		m.SetPhase(domain.MightyPhaseTrumpAndFriend)
		m.SetDeclarerIdx(0)
		m.CpuDeclareTrumpAndFriend()
		assert.Equal(t, domain.MightyPhaseTrumpAndFriend, m.GetPhase())
	})

	t.Run("no-op when wrong phase", func(t *testing.T) {
		m := newTestMighty()
		m.Reset()
		m.SetDeclarerIdx(1)
		m.SetPhase(domain.MightyPhasePlay)
		m.CpuDeclareTrumpAndFriend()
		assert.Equal(t, domain.MightyPhasePlay, m.GetPhase())
	})
}

// --- D. Kitty exchange ---

// setupForKittyExchange puts m into KittyExchange with the declarer holding
// 13 (=10+3) deterministic cards so discards are predictable.
func setupForKittyExchange(t *testing.T, declarerIdx int) *domain.Mighty {
	t.Helper()
	m := newTestMighty()
	m.Reset()
	m.SetPhase(domain.MightyPhaseKittyExchange)
	m.SetDeclarerIdx(declarerIdx)
	m.SetHighestBid(15)
	m.GetPlayer(declarerIdx).SetIsDeclarer(true)
	m.SetTrumpSuit(domain.CardDesignSpade)
	m.SetPartnerCard(mightyCard(domain.CardDesignHeart, 1))
	// 13 cards in declarer hand. Mix of high & low so picks are deterministic.
	replaceHand(m.GetPlayer(declarerIdx),
		mightyCard(domain.CardDesignHeart, 2),
		mightyCard(domain.CardDesignHeart, 3),
		mightyCard(domain.CardDesignHeart, 4),
		mightyCard(domain.CardDesignHeart, 5),
		mightyCard(domain.CardDesignHeart, 6),
		mightyCard(domain.CardDesignHeart, 7),
		mightyCard(domain.CardDesignHeart, 8),
		mightyCard(domain.CardDesignHeart, 9),
		mightyCard(domain.CardDesignHeart, 10), // point card
		mightyCard(domain.CardDesignDiamond, 11),
		mightyCard(domain.CardDesignDiamond, 12),
		mightyCard(domain.CardDesignDiamond, 13),
		mightyCard(domain.CardDesignDiamond, 1),
	)
	return m
}

func TestMighty_ExchangeKitty_validations(t *testing.T) {
	t.Run("error when game ended", func(t *testing.T) {
		m := setupForKittyExchange(t, 0)
		m.SetGameEndFlag(true)
		err := m.PlayerExchangeKitty([]int{0, 1, 2})
		assert.True(t, errors.Is(err, domain.ErrGameEnded))
	})

	t.Run("error when wrong phase", func(t *testing.T) {
		m := setupForKittyExchange(t, 0)
		m.SetPhase(domain.MightyPhasePlay)
		err := m.PlayerExchangeKitty([]int{0, 1, 2})
		assert.True(t, errors.Is(err, domain.ErrWrongPhase))
	})

	t.Run("error when CPU is declarer", func(t *testing.T) {
		m := setupForKittyExchange(t, 1)
		err := m.PlayerExchangeKitty([]int{0, 1, 2})
		assert.True(t, errors.Is(err, domain.ErrNotHumanTurn))
	})

	t.Run("wrong number of indices", func(t *testing.T) {
		m := setupForKittyExchange(t, 0)
		err := m.PlayerExchangeKitty([]int{0, 1})
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("index out of range", func(t *testing.T) {
		m := setupForKittyExchange(t, 0)
		err := m.PlayerExchangeKitty([]int{0, 1, 100})
		assert.True(t, errors.Is(err, domain.ErrInvalidCard))
	})

	t.Run("duplicate indices", func(t *testing.T) {
		m := setupForKittyExchange(t, 0)
		err := m.PlayerExchangeKitty([]int{0, 0, 1})
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})
}

func TestMighty_ExchangeKitty_keepsHandSizeTen(t *testing.T) {
	m := setupForKittyExchange(t, 0)
	assert.Equal(t, 13, m.GetPlayer(0).GetCardsSize(), "precondition: 10 hand + 3 kitty")
	err := m.PlayerExchangeKitty([]int{0, 1, 2})
	require.NoError(t, err)
	assert.Equal(t, domain.MightyHandSize, m.GetPlayer(0).GetCardsSize())
	assert.Equal(t, domain.MightyPhasePlay, m.GetPhase())
}

func TestMighty_ExchangeKitty_pointCardsCountedForDeclarer(t *testing.T) {
	m := setupForKittyExchange(t, 0)
	// Hand has ♥10 (point card) at index 8; discard it + two non-point lows.
	require.NoError(t, m.PlayerExchangeKitty([]int{0, 1, 8}))
	assert.Equal(t, 1, m.GetPlayer(0).GetPointCards(), "♥10 discarded should add 1 point card to declarer")
}

func TestMighty_ExchangeKitty_phaseAdvancesAndStartsPlay(t *testing.T) {
	m := setupForKittyExchange(t, 0)
	require.NoError(t, m.PlayerExchangeKitty([]int{0, 1, 2}))
	assert.Equal(t, domain.MightyPhasePlay, m.GetPhase())
	assert.Equal(t, 0, m.GetCurrentPlayerIdx())
	assert.Equal(t, 0, m.GetLeadPlayerIdx())
	assert.Equal(t, 1, m.GetTrickNumber())
}

func TestMighty_CpuExchangeKitty_picksLowestThree(t *testing.T) {
	m := setupForKittyExchange(t, 1) // CPU declarer
	// Seed CPU declarer's hand: 3 obvious "junk" lows (♥2, ♥3, ♥4) plus 10
	// cards that are clearly more valuable (Mighty/Joker/trump/points).
	replaceHand(m.GetPlayer(1),
		mightyCard(domain.CardDesignHeart, 2),    // low junk
		mightyCard(domain.CardDesignHeart, 3),    // low junk
		mightyCard(domain.CardDesignHeart, 4),    // low junk
		mightyCard(domain.CardDesignSpade, 1),    // Mighty
		mightyCard(domain.CardDesignSpade, 13),   // trump K
		mightyCard(domain.CardDesignSpade, 12),   // trump Q
		mightyCard(domain.CardDesignSpade, 11),   // trump J
		mightyCard(domain.CardDesignSpade, 10),   // trump point
		mightyCard(domain.CardDesignDiamond, 1),  // ace
		mightyCard(domain.CardDesignDiamond, 13), // K
		mightyCard(domain.CardDesignDiamond, 12), // Q
		mightyCard(domain.CardDesignDiamond, 11), // J
		mightyCard(domain.CardDesignJoker, 1),    // Joker
	)
	m.CpuExchangeKitty()

	kitty := m.GetKitty()
	require.Len(t, kitty, domain.MightyKittySize)
	// CPU should have discarded ♥2, ♥3, ♥4.
	gotValues := map[int]bool{}
	for _, c := range kitty {
		assert.Equal(t, domain.CardDesignHeart, c.GetDesign())
		gotValues[c.GetValue()] = true
	}
	assert.True(t, gotValues[2])
	assert.True(t, gotValues[3])
	assert.True(t, gotValues[4])
}

func TestMighty_CpuExchangeKitty_guards(t *testing.T) {
	t.Run("no-op when game ended", func(t *testing.T) {
		m := setupForKittyExchange(t, 1)
		m.SetGameEndFlag(true)
		m.CpuExchangeKitty()
		assert.Equal(t, domain.MightyPhaseKittyExchange, m.GetPhase())
	})

	t.Run("no-op when wrong phase", func(t *testing.T) {
		m := setupForKittyExchange(t, 1)
		m.SetPhase(domain.MightyPhasePlay)
		m.CpuExchangeKitty()
		assert.Equal(t, domain.MightyPhasePlay, m.GetPhase())
	})

	t.Run("no-op when human is declarer", func(t *testing.T) {
		m := setupForKittyExchange(t, 0)
		m.CpuExchangeKitty()
		assert.Equal(t, domain.MightyPhaseKittyExchange, m.GetPhase())
	})
}

// --- E. Trick play — Mighty card ---

// setupForPlay returns a Mighty fixture in Play phase, trick 2 (so Joker Call
// can activate when desired), current player = 0 (human), with every player's
// hand cleared.
func setupForPlay(t *testing.T, trump int) *domain.Mighty {
	t.Helper()
	m := newTestMighty()
	m.Reset()
	m.SetPhase(domain.MightyPhasePlay)
	m.SetTrumpSuit(trump)
	m.SetTrickNumber(2)
	m.SetCurrentPlayerIdx(0)
	m.SetLeadPlayerIdx(0)
	m.SetDeclarerIdx(0)
	m.SetCurrentTrick(nil)
	for i := 0; i < m.GetPlayerCnt(); i++ {
		m.GetPlayer(i).Reset()
	}
	return m
}

// runFullTrick lets every player play their single seeded card, drives
// CpuPlay where necessary, then resolves and returns the trick winner index.
func runFullTrick(t *testing.T, m *domain.Mighty) int {
	t.Helper()
	for i := 0; i < domain.MightyPlayerCnt; i++ {
		require.Equal(t, domain.MightyPhasePlay, m.GetPhase(), "trick should still be in Play before step %d", i)
		cur := m.GetCurrentPlayerIdx()
		if m.IsHumanTurn() {
			require.NoError(t, m.PlayerPlay(0))
		} else {
			m.CpuPlay()
		}
		_ = cur
	}
	require.Equal(t, domain.MightyPhaseTrickEnd, m.GetPhase())
	m.ResolveTrick()
	for i := 0; i < domain.MightyPlayerCnt; i++ {
		if m.GetPlayer(i).GetTrickCount() > 0 {
			return i
		}
	}
	t.Fatalf("no trick winner detected")
	return -1
}

func TestMighty_Trick_MightyBeatsEverything_nonSpadeTrump(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignHeart)
	// Each player has exactly one card; play one card each.
	replaceHand(m.GetPlayer(0), mightyCard(domain.CardDesignSpade, 1))   // Mighty (♠A)
	replaceHand(m.GetPlayer(1), mightyCard(domain.CardDesignHeart, 1))   // trump A
	replaceHand(m.GetPlayer(2), mightyCard(domain.CardDesignJoker, 1))   // Joker
	replaceHand(m.GetPlayer(3), mightyCard(domain.CardDesignHeart, 13))  // trump K
	replaceHand(m.GetPlayer(4), mightyCard(domain.CardDesignDiamond, 1)) // off-suit A
	winner := runFullTrick(t, m)
	assert.Equal(t, 0, winner, "♠A should win when trump is hearts")
}

func TestMighty_Trick_MightyBeatsEverything_spadeTrump(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignSpade)
	// Trump = ♠. Mighty becomes ♦A. ♠A is now just a regular trump ace.
	replaceHand(m.GetPlayer(0), mightyCard(domain.CardDesignDiamond, 1)) // Mighty under spade trump
	replaceHand(m.GetPlayer(1), mightyCard(domain.CardDesignSpade, 1))   // trump A (no longer Mighty)
	replaceHand(m.GetPlayer(2), mightyCard(domain.CardDesignJoker, 1))
	replaceHand(m.GetPlayer(3), mightyCard(domain.CardDesignSpade, 13))
	replaceHand(m.GetPlayer(4), mightyCard(domain.CardDesignHeart, 1))
	winner := runFullTrick(t, m)
	assert.Equal(t, 0, winner, "♦A should win when trump is spades")
}

func TestMighty_Trick_MightyIsAlwaysLegal(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignHeart)
	// Lead is hearts. Player has ♥3 (lead suit) AND ♠A (Mighty).
	// Playing Mighty must be allowed even though the player holds lead suit.
	m.SetCurrentTrick([]*domain.MightyTrickCard{
		{PlayerIdx: 1, Card: mightyCard(domain.CardDesignHeart, 5)},
	})
	m.SetCurrentPlayerIdx(0)
	replaceHand(m.GetPlayer(0),
		mightyCard(domain.CardDesignHeart, 3), // lead suit
		mightyCard(domain.CardDesignSpade, 1), // Mighty
	)
	valid := m.GetValidPlayIndices(0)
	assert.Contains(t, valid, 1, "Mighty must be in valid plays")
	// Actually play Mighty (index 1) and confirm no error.
	require.NoError(t, m.PlayerPlay(1))
}

// --- F. Trick play — Joker ---

func TestMighty_Trick_JokerSecondHighest_whenNotLed(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignHeart)
	// Player 0 leads hearts; Joker (played by player 2) wins over hearts but
	// not over ♠A (Mighty); we omit Mighty here so Joker wins.
	replaceHand(m.GetPlayer(0), mightyCard(domain.CardDesignHeart, 5))
	replaceHand(m.GetPlayer(1), mightyCard(domain.CardDesignHeart, 13))
	replaceHand(m.GetPlayer(2), mightyCard(domain.CardDesignJoker, 1))
	replaceHand(m.GetPlayer(3), mightyCard(domain.CardDesignHeart, 1)) // trump A
	replaceHand(m.GetPlayer(4), mightyCard(domain.CardDesignDiamond, 9))
	winner := runFullTrick(t, m)
	assert.Equal(t, 2, winner, "Joker (non-lead) should beat all non-Mighty cards")
}

func TestMighty_Trick_JokerLegalEvenWithLeadSuit(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignHeart)
	m.SetCurrentTrick([]*domain.MightyTrickCard{
		{PlayerIdx: 1, Card: mightyCard(domain.CardDesignClover, 5)},
	})
	m.SetCurrentPlayerIdx(0)
	replaceHand(m.GetPlayer(0),
		mightyCard(domain.CardDesignClover, 7), // lead suit
		mightyCard(domain.CardDesignJoker, 1),  // Joker
	)
	valid := m.GetValidPlayIndices(0)
	assert.Contains(t, valid, 1, "Joker should be a valid follow even when player has lead suit")
}

// --- G. Trick play — Joker Call ---

func TestMighty_Trick_JokerCallForcesJoker_afterFirstTrick(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignHeart) // clubs (♣) not trump
	m.SetTrickNumber(2)
	m.SetJokerPlayed(false)
	m.SetCurrentTrick([]*domain.MightyTrickCard{
		{PlayerIdx: 1, Card: mightyCard(domain.CardDesignClover, 3)}, // Joker Call card
	})
	m.SetCurrentPlayerIdx(0)
	replaceHand(m.GetPlayer(0),
		mightyCard(domain.CardDesignClover, 9),
		mightyCard(domain.CardDesignJoker, 1),
	)
	valid := m.GetValidPlayIndices(0)
	// Joker (index 1) must be the only legal play.
	assert.Equal(t, []int{1}, valid)
	err := m.PlayerPlay(0) // try to play ♣9 instead of Joker
	assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
}

func TestMighty_Trick_JokerCallSuppressed_onFirstTrick(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignHeart)
	m.SetTrickNumber(1)
	m.SetCurrentTrick([]*domain.MightyTrickCard{
		{PlayerIdx: 1, Card: mightyCard(domain.CardDesignClover, 3)}, // Joker Call card on trick 1
	})
	m.SetCurrentPlayerIdx(0)
	replaceHand(m.GetPlayer(0),
		mightyCard(domain.CardDesignClover, 9),
		mightyCard(domain.CardDesignJoker, 1),
	)
	valid := m.GetValidPlayIndices(0)
	// Both should be valid because Joker Call is suppressed on trick 1.
	assert.Contains(t, valid, 0)
	assert.Contains(t, valid, 1)
}

func TestMighty_Trick_JokerCallSuppressed_afterJokerPlayed(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignHeart)
	m.SetTrickNumber(3)
	m.SetJokerPlayed(true)
	m.SetCurrentTrick([]*domain.MightyTrickCard{
		{PlayerIdx: 1, Card: mightyCard(domain.CardDesignClover, 3)},
	})
	m.SetCurrentPlayerIdx(0)
	replaceHand(m.GetPlayer(0),
		mightyCard(domain.CardDesignClover, 9),
		mightyCard(domain.CardDesignJoker, 1),
	)
	valid := m.GetValidPlayIndices(0)
	assert.Contains(t, valid, 0, "♣9 should be legal after Joker already played")
	assert.Contains(t, valid, 1)
}

func TestMighty_Trick_JokerCallShifts_whenClubsTrump(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignClover) // clubs ARE trump
	m.SetTrickNumber(2)
	m.SetJokerPlayed(false)
	// ♠3 is now the Joker Call card.
	m.SetCurrentTrick([]*domain.MightyTrickCard{
		{PlayerIdx: 1, Card: mightyCard(domain.CardDesignSpade, 3)},
	})
	m.SetCurrentPlayerIdx(0)
	replaceHand(m.GetPlayer(0),
		mightyCard(domain.CardDesignSpade, 9),
		mightyCard(domain.CardDesignJoker, 1),
	)
	valid := m.GetValidPlayIndices(0)
	assert.Equal(t, []int{1}, valid, "Only Joker is legal when ♠3 leads with clubs trump")

	// ♣3 should NOT force Joker when clubs is trump.
	m2 := setupForPlay(t, domain.CardDesignClover)
	m2.SetTrickNumber(2)
	m2.SetJokerPlayed(false)
	m2.SetCurrentTrick([]*domain.MightyTrickCard{
		{PlayerIdx: 1, Card: mightyCard(domain.CardDesignClover, 3)},
	})
	m2.SetCurrentPlayerIdx(0)
	replaceHand(m2.GetPlayer(0),
		mightyCard(domain.CardDesignClover, 9), // can follow clubs
		mightyCard(domain.CardDesignJoker, 1),
	)
	valid2 := m2.GetValidPlayIndices(0)
	assert.Contains(t, valid2, 0, "♣9 should be legal: ♣3 is not the joker-call when clubs is trump")
}

// --- H. Joker lead ---

func TestMighty_PlayerPlay_RejectsJokerLead(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignHeart)
	m.SetCurrentTrick(nil) // empty trick → leading
	replaceHand(m.GetPlayer(0), mightyCard(domain.CardDesignJoker, 1))
	err := m.PlayerPlay(0)
	assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
}

func TestMighty_PlayerPlayJokerLead_validations(t *testing.T) {
	t.Run("error when wrong phase", func(t *testing.T) {
		m := setupForPlay(t, domain.CardDesignHeart)
		m.SetPhase(domain.MightyPhaseBid)
		replaceHand(m.GetPlayer(0), mightyCard(domain.CardDesignJoker, 1))
		err := m.PlayerPlayJokerLead(0, domain.CardDesignHeart)
		assert.True(t, errors.Is(err, domain.ErrWrongPhase))
	})

	t.Run("error when not human turn", func(t *testing.T) {
		m := setupForPlay(t, domain.CardDesignHeart)
		m.SetCurrentPlayerIdx(1)
		replaceHand(m.GetPlayer(1), mightyCard(domain.CardDesignJoker, 1))
		err := m.PlayerPlayJokerLead(0, domain.CardDesignHeart)
		assert.True(t, errors.Is(err, domain.ErrNotHumanTurn))
	})

	t.Run("error when not first card in trick", func(t *testing.T) {
		m := setupForPlay(t, domain.CardDesignHeart)
		m.SetCurrentTrick([]*domain.MightyTrickCard{
			{PlayerIdx: 1, Card: mightyCard(domain.CardDesignClover, 5)},
		})
		replaceHand(m.GetPlayer(0), mightyCard(domain.CardDesignJoker, 1))
		err := m.PlayerPlayJokerLead(0, domain.CardDesignHeart)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("error when card is not a joker", func(t *testing.T) {
		m := setupForPlay(t, domain.CardDesignHeart)
		replaceHand(m.GetPlayer(0), mightyCard(domain.CardDesignSpade, 9))
		err := m.PlayerPlayJokerLead(0, domain.CardDesignHeart)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("error when demand suit invalid", func(t *testing.T) {
		m := setupForPlay(t, domain.CardDesignHeart)
		replaceHand(m.GetPlayer(0), mightyCard(domain.CardDesignJoker, 1))
		err := m.PlayerPlayJokerLead(0, domain.CardDesignJoker)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("error when card index out of range", func(t *testing.T) {
		m := setupForPlay(t, domain.CardDesignHeart)
		replaceHand(m.GetPlayer(0), mightyCard(domain.CardDesignJoker, 1))
		err := m.PlayerPlayJokerLead(99, domain.CardDesignHeart)
		assert.True(t, errors.Is(err, domain.ErrInvalidCard))
	})

	t.Run("error when game ended", func(t *testing.T) {
		m := setupForPlay(t, domain.CardDesignHeart)
		m.SetGameEndFlag(true)
		replaceHand(m.GetPlayer(0), mightyCard(domain.CardDesignJoker, 1))
		err := m.PlayerPlayJokerLead(0, domain.CardDesignHeart)
		assert.True(t, errors.Is(err, domain.ErrGameEnded))
	})
}

func TestMighty_PlayerPlayJokerLead_demandsSuitAndLoses(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignDiamond)
	// Player 0 leads Joker, demands hearts. Others holding hearts must play hearts.
	replaceHand(m.GetPlayer(0), mightyCard(domain.CardDesignJoker, 1))
	replaceHand(m.GetPlayer(1), mightyCard(domain.CardDesignHeart, 5), mightyCard(domain.CardDesignClover, 9))
	replaceHand(m.GetPlayer(2), mightyCard(domain.CardDesignHeart, 13))
	replaceHand(m.GetPlayer(3), mightyCard(domain.CardDesignClover, 7)) // void in hearts
	replaceHand(m.GetPlayer(4), mightyCard(domain.CardDesignHeart, 8))

	require.NoError(t, m.PlayerPlayJokerLead(0, domain.CardDesignHeart))

	// Player 1 holds hearts → must play ♥5, not ♣9.
	valid := m.GetValidPlayIndices(1)
	// Hand order may be sorted after seed, so confirm by card identity.
	for _, idx := range valid {
		c := m.GetPlayer(1).GetCard(idx)
		assert.Equal(t, domain.CardDesignHeart, c.GetDesign(), "player 1 must follow hearts when holding them")
	}

	m.CpuPlay() // player 1
	m.CpuPlay() // player 2
	m.CpuPlay() // player 3 (void → any card)
	m.CpuPlay() // player 4
	require.Equal(t, domain.MightyPhaseTrickEnd, m.GetPhase())
	m.ResolveTrick()
	// Player 0's Joker should lose (treated as weakest demand-suit card).
	assert.Equal(t, 0, m.GetPlayer(0).GetTrickCount())
}

func TestMighty_PlayerPlayJokerLead_MightyStillBeatsJokerLead(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignHeart) // Mighty = ♠A
	replaceHand(m.GetPlayer(0), mightyCard(domain.CardDesignJoker, 1))
	replaceHand(m.GetPlayer(1), mightyCard(domain.CardDesignSpade, 1)) // Mighty
	replaceHand(m.GetPlayer(2), mightyCard(domain.CardDesignHeart, 5))
	replaceHand(m.GetPlayer(3), mightyCard(domain.CardDesignHeart, 7))
	replaceHand(m.GetPlayer(4), mightyCard(domain.CardDesignHeart, 9))

	require.NoError(t, m.PlayerPlayJokerLead(0, domain.CardDesignHeart))
	m.CpuPlay()
	m.CpuPlay()
	m.CpuPlay()
	m.CpuPlay()
	require.Equal(t, domain.MightyPhaseTrickEnd, m.GetPhase())
	m.ResolveTrick()
	assert.Equal(t, 1, m.GetPlayer(1).GetTrickCount(), "Mighty still beats a Joker lead")
}

// --- I. Follow-suit ---

func TestMighty_FollowSuit_required(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignHeart)
	m.SetCurrentTrick([]*domain.MightyTrickCard{
		{PlayerIdx: 1, Card: mightyCard(domain.CardDesignClover, 5)},
	})
	m.SetCurrentPlayerIdx(0)
	replaceHand(m.GetPlayer(0),
		mightyCard(domain.CardDesignClover, 9), // lead suit
		mightyCard(domain.CardDesignDiamond, 2),
	)
	valid := m.GetValidPlayIndices(0)
	for _, idx := range valid {
		c := m.GetPlayer(0).GetCard(idx)
		assert.Equal(t, domain.CardDesignClover, c.GetDesign(), "must follow clubs when holding one")
	}
}

func TestMighty_FollowSuit_voidAllowsAnything(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignHeart)
	m.SetCurrentTrick([]*domain.MightyTrickCard{
		{PlayerIdx: 1, Card: mightyCard(domain.CardDesignClover, 5)},
	})
	m.SetCurrentPlayerIdx(0)
	replaceHand(m.GetPlayer(0),
		mightyCard(domain.CardDesignDiamond, 2),
		mightyCard(domain.CardDesignDiamond, 9),
	)
	valid := m.GetValidPlayIndices(0)
	assert.Len(t, valid, 2, "void in lead suit → any card legal")
}

// --- J. Round scoring ---

func setupForScoring(t *testing.T, bid, declarerPts int, opts func(*domain.Mighty)) *domain.Mighty {
	t.Helper()
	m := newTestMighty()
	m.Reset()
	m.SetPhase(domain.MightyPhaseRoundEnd)
	m.SetHighestBid(bid)
	m.SetHighestBidder(0)
	m.SetDeclarerIdx(0)
	m.SetPartnerIdx(1)
	m.GetPlayer(0).SetIsDeclarer(true)
	m.GetPlayer(1).SetIsPartner(true)
	// Split declarer-side point cards between declarer & partner.
	half := declarerPts / 2
	m.GetPlayer(0).SetPointCards(declarerPts - half)
	m.GetPlayer(1).SetPointCards(half)
	if opts != nil {
		opts(m)
	}
	return m
}

func TestMighty_ScoreRound_declarerWins(t *testing.T) {
	m := setupForScoring(t, 13, 14, nil)
	m.ScoreRound()
	assert.Equal(t, domain.MightyWinnerDeclarer, m.GetWinnerTeam())
	// gain = (14-13+1)*1 = 2. Declarer & partner: +2 each. Opposition: -2 each.
	assert.Equal(t, 2, m.GetPlayer(0).GetRoundScore(), "declarer")
	assert.Equal(t, 2, m.GetPlayer(1).GetRoundScore(), "partner")
	for i := 2; i < domain.MightyPlayerCnt; i++ {
		assert.Equal(t, -2, m.GetPlayer(i).GetRoundScore(), "opposition %d", i)
	}
}

func TestMighty_ScoreRound_declarerLoses(t *testing.T) {
	m := setupForScoring(t, 14, 12, nil)
	m.ScoreRound()
	assert.Equal(t, domain.MightyWinnerOpposition, m.GetWinnerTeam())
	// loss = (14-12+1)*1 = 3. Declarer & partner: -3 each. Opposition: +3 each.
	assert.Equal(t, -3, m.GetPlayer(0).GetRoundScore(), "declarer")
	assert.Equal(t, -3, m.GetPlayer(1).GetRoundScore(), "partner")
	for i := 2; i < domain.MightyPlayerCnt; i++ {
		assert.Equal(t, 3, m.GetPlayer(i).GetRoundScore(), "opposition %d", i)
	}
}

func TestMighty_ScoreRound_noTrumpMultiplier(t *testing.T) {
	m := setupForScoring(t, 13, 14, func(m *domain.Mighty) {
		m.SetWinningBidNoTrump(true)
	})
	m.ScoreRound()
	// multiplier = 2 → gain = (14-13+1)*2 = 4.
	assert.Equal(t, 4, m.GetPlayer(0).GetRoundScore())
	assert.Equal(t, 4, m.GetPlayer(1).GetRoundScore())
	for i := 2; i < domain.MightyPlayerCnt; i++ {
		assert.Equal(t, -4, m.GetPlayer(i).GetRoundScore())
	}
}

func TestMighty_ScoreRound_soloDeclarer(t *testing.T) {
	m := newTestMighty()
	m.Reset()
	m.SetPhase(domain.MightyPhaseRoundEnd)
	m.SetHighestBid(13)
	m.SetHighestBidder(0)
	m.SetDeclarerIdx(0)
	m.SetPartnerIdx(0) // solo: partner == declarer
	m.GetPlayer(0).SetIsDeclarer(true)
	m.GetPlayer(0).SetIsPartner(true)
	m.GetPlayer(0).SetPointCards(14)

	m.ScoreRound()
	// gain = (14-13+1)*1 = 2. Solo declarer gets gain*2 = 4.
	assert.Equal(t, 4, m.GetPlayer(0).GetRoundScore(), "solo declarer doubles")
	for i := 1; i < domain.MightyPlayerCnt; i++ {
		assert.Equal(t, -2, m.GetPlayer(i).GetRoundScore(), "opposition %d", i)
	}
}

func TestMighty_ScoreRound_soloDeclarer_loses(t *testing.T) {
	m := newTestMighty()
	m.Reset()
	m.SetPhase(domain.MightyPhaseRoundEnd)
	m.SetHighestBid(14)
	m.SetHighestBidder(0)
	m.SetDeclarerIdx(0)
	m.SetPartnerIdx(0)
	m.GetPlayer(0).SetIsDeclarer(true)
	m.GetPlayer(0).SetIsPartner(true)
	m.GetPlayer(0).SetPointCards(12)
	m.ScoreRound()
	// loss = (14-12+1)*1 = 3, solo doubles → declarer -6.
	assert.Equal(t, -6, m.GetPlayer(0).GetRoundScore())
	for i := 1; i < domain.MightyPlayerCnt; i++ {
		assert.Equal(t, 3, m.GetPlayer(i).GetRoundScore())
	}
}

func TestMighty_ScoreRound_noOpWhenWrongPhase(t *testing.T) {
	m := newTestMighty()
	m.Reset()
	m.SetPhase(domain.MightyPhasePlay)
	m.ScoreRound()
	assert.Equal(t, domain.MightyPhasePlay, m.GetPhase())
}

// --- K. Game end ---

func TestMighty_GameEnd_atPointLimit(t *testing.T) {
	m := newTestMighty()
	m.Reset()
	m.SetPhase(domain.MightyPhaseRoundEnd)
	m.SetHighestBid(13)
	m.SetHighestBidder(0)
	m.SetDeclarerIdx(0)
	m.SetPartnerIdx(0) // solo declarer doubles the gain
	m.GetPlayer(0).SetIsDeclarer(true)
	m.GetPlayer(0).SetIsPartner(true)
	cfg := m.GetConfig()
	cfg.PointLimit = 5
	m.SetConfig(cfg)
	// gain=(20-13+1)*1 = 8; solo declarer gets 16 → over PointLimit=5.
	m.GetPlayer(0).SetPointCards(20)

	m.ScoreRound()
	assert.True(t, m.GetGameEndFlag())
	assert.Equal(t, domain.MightyPhaseGameEnd, m.GetPhase())
}

func TestMighty_GameEnd_atNegativePointLimit(t *testing.T) {
	m := newTestMighty()
	m.Reset()
	m.SetPhase(domain.MightyPhaseRoundEnd)
	m.SetHighestBid(20) // declarer needs 20 to win; will lose with 0.
	m.SetHighestBidder(0)
	m.SetDeclarerIdx(0)
	m.SetPartnerIdx(0) // solo declarer → -loss*2
	m.GetPlayer(0).SetIsDeclarer(true)
	m.GetPlayer(0).SetIsPartner(true)
	cfg := m.GetConfig()
	cfg.PointLimit = 10
	m.SetConfig(cfg)
	m.GetPlayer(0).SetPointCards(0) // loss=(20-0+1)*1 = 21, solo doubles → -42

	m.ScoreRound()
	assert.True(t, m.GetGameEndFlag())
}

// --- L. JSON snapshot round-trip ---

func TestMighty_JSONRoundTrip(t *testing.T) {
	m := newTestMighty()
	m.Reset()
	m.SetPhase(domain.MightyPhasePlay)
	m.SetRoundNumber(3)
	m.SetTrickNumber(4)
	m.SetTrumpSuit(domain.CardDesignHeart)
	m.SetPartnerCard(mightyCard(domain.CardDesignDiamond, 1))
	m.SetDeclarerIdx(2)
	m.SetPartnerIdx(4)
	m.SetPartnerRevealed(true)
	m.SetLeadPlayerIdx(2)
	m.SetCurrentPlayerIdx(3)
	m.SetBidPlayerIdx(domain.MightyPlayerCnt)
	m.SetKitty([]*domain.Card{
		mightyCard(domain.CardDesignSpade, 7),
		mightyCard(domain.CardDesignClover, 8),
		mightyCard(domain.CardDesignDiamond, 9),
	})
	m.SetHighestBid(15)
	m.SetHighestBidder(2)
	m.SetWinningBidNoTrump(false)
	m.SetPassCount(2)
	m.SetJokerPlayed(true)
	m.SetCurrentTrick([]*domain.MightyTrickCard{
		{PlayerIdx: 2, Card: mightyCard(domain.CardDesignHeart, 10), IsJokerLead: false, LeadDemandSuit: 0},
		{PlayerIdx: 3, Card: mightyCard(domain.CardDesignJoker, 1), IsJokerLead: false, LeadDemandSuit: 0},
	})

	data, err := json.Marshal(m)
	require.NoError(t, err)

	m2 := domain.NewDefaultMighty()
	require.NoError(t, json.Unmarshal(data, m2))

	assert.Equal(t, m.GetPhase(), m2.GetPhase())
	assert.Equal(t, m.GetRoundNumber(), m2.GetRoundNumber())
	assert.Equal(t, m.GetTrickNumber(), m2.GetTrickNumber())
	assert.Equal(t, m.GetTrumpSuit(), m2.GetTrumpSuit())
	assert.Equal(t, m.GetDeclarerIdx(), m2.GetDeclarerIdx())
	assert.Equal(t, m.GetPartnerIdx(), m2.GetPartnerIdx())
	assert.Equal(t, m.GetPartnerRevealed(), m2.GetPartnerRevealed())
	assert.Equal(t, m.GetLeadPlayerIdx(), m2.GetLeadPlayerIdx())
	assert.Equal(t, m.GetCurrentPlayerIdx(), m2.GetCurrentPlayerIdx())
	assert.Equal(t, m.GetBidPlayerIdx(), m2.GetBidPlayerIdx())
	assert.Equal(t, m.GetHighestBid(), m2.GetHighestBid())
	assert.Equal(t, m.GetHighestBidder(), m2.GetHighestBidder())
	assert.Equal(t, m.GetWinningBidNoTrump(), m2.GetWinningBidNoTrump())
	assert.Equal(t, m.GetPassCount(), m2.GetPassCount())
	assert.Equal(t, m.GetJokerPlayed(), m2.GetJokerPlayed())

	require.NotNil(t, m2.GetPartnerCard())
	assert.Equal(t, m.GetPartnerCard().GetDesign(), m2.GetPartnerCard().GetDesign())
	assert.Equal(t, m.GetPartnerCard().GetValue(), m2.GetPartnerCard().GetValue())

	assert.Equal(t, len(m.GetKitty()), len(m2.GetKitty()))
	for i := range m.GetKitty() {
		assert.Equal(t, m.GetKitty()[i].GetDesign(), m2.GetKitty()[i].GetDesign())
		assert.Equal(t, m.GetKitty()[i].GetValue(), m2.GetKitty()[i].GetValue())
	}

	assert.Equal(t, len(m.GetCurrentTrick()), len(m2.GetCurrentTrick()))
	for i := range m.GetCurrentTrick() {
		assert.Equal(t, m.GetCurrentTrick()[i].PlayerIdx, m2.GetCurrentTrick()[i].PlayerIdx)
		assert.Equal(t, m.GetCurrentTrick()[i].IsJokerLead, m2.GetCurrentTrick()[i].IsJokerLead)
	}

	assert.Equal(t, m.GetPlayerCnt(), m2.GetPlayerCnt())
	for i := 0; i < m.GetPlayerCnt(); i++ {
		assert.Equal(t, m.GetPlayer(i).GetCardsSize(), m2.GetPlayer(i).GetCardsSize())
	}
}

func TestMighty_UnmarshalJSON_arraySizeLimit(t *testing.T) {
	bigJSON := `{"al":[`
	for i := 0; i < 1001; i++ {
		if i > 0 {
			bigJSON += ","
		}
		bigJSON += `{"t":0,"p":0,"a":"x","d":"y","c":[]}`
	}
	bigJSON += `]}`
	m := domain.NewDefaultMighty()
	err := json.Unmarshal([]byte(bigJSON), m)
	assert.Error(t, err)
}

// --- M. Hint ---

func TestMighty_Hint_Bid(t *testing.T) {
	m := newTestMighty()
	m.Reset()
	// Seed human with a clearly strong hand so Hard CPU bids.
	replaceHand(m.GetPlayer(0),
		mightyCard(domain.CardDesignSpade, 1),    // Mighty
		mightyCard(domain.CardDesignJoker, 1),    // Joker
		mightyCard(domain.CardDesignHeart, 1),    // A
		mightyCard(domain.CardDesignHeart, 13),   // K
		mightyCard(domain.CardDesignHeart, 12),   // Q
		mightyCard(domain.CardDesignHeart, 11),   // J
		mightyCard(domain.CardDesignHeart, 10),   // point
		mightyCard(domain.CardDesignDiamond, 1),  // A
		mightyCard(domain.CardDesignDiamond, 13), // K
		mightyCard(domain.CardDesignDiamond, 10), // point
	)
	hint := m.GetHint()
	require.NotNil(t, hint)
	assert.NotNil(t, hint.Bid)
}

func TestMighty_Hint_Play(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignHeart)
	// Human leads with a varied hand; just need a non-trivial valid index.
	replaceHand(m.GetPlayer(0),
		mightyCard(domain.CardDesignHeart, 5),
		mightyCard(domain.CardDesignDiamond, 9),
		mightyCard(domain.CardDesignSpade, 7),
	)
	hint := m.GetHint()
	require.NotNil(t, hint)
	require.NotNil(t, hint.CardIndex)
	assert.GreaterOrEqual(t, *hint.CardIndex, 0)
	assert.Less(t, *hint.CardIndex, m.GetPlayer(0).GetCardsSize())
}

func TestMighty_Hint_NoHumanInGame(t *testing.T) {
	// All CPU game → no hint.
	players := []*domain.MightyPlayer{
		domain.NewMightyPlayer(false),
		domain.NewMightyPlayer(false),
		domain.NewMightyPlayer(false),
		domain.NewMightyPlayer(false),
		domain.NewMightyPlayer(false),
	}
	m := domain.NewMighty(domain.NewTrumpCards(1), players, domain.DefaultMightyConfig())
	m.Reset()
	assert.Nil(t, m.GetHint())
}

func TestMighty_Hint_WrongPhase(t *testing.T) {
	m := newTestMighty()
	m.Reset()
	m.SetPhase(domain.MightyPhaseGameEnd)
	assert.Nil(t, m.GetHint())
}

func TestMighty_Hint_KittyExchange(t *testing.T) {
	m := setupForKittyExchange(t, 0)
	hint := m.GetHint()
	require.NotNil(t, hint)
	assert.Len(t, hint.DiscardIndices, domain.MightyKittySize)
}

func TestMighty_Hint_TrumpAndFriend(t *testing.T) {
	m := setupForDeclare(t)
	hint := m.GetHint()
	require.NotNil(t, hint)
	require.NotNil(t, hint.TrumpSuit)
	require.NotNil(t, hint.PartnerSuit)
	require.NotNil(t, hint.PartnerValue)
}

// --- Convenience predicate setters/getters ---

func TestMighty_IsHumanTurn(t *testing.T) {
	m := newTestMighty()
	m.Reset()
	m.SetCurrentPlayerIdx(0)
	assert.True(t, m.IsHumanTurn())
	m.SetCurrentPlayerIdx(1)
	assert.False(t, m.IsHumanTurn())
	m.SetCurrentPlayerIdx(-1)
	assert.False(t, m.IsHumanTurn())
	m.SetCurrentPlayerIdx(domain.MightyPlayerCnt)
	assert.False(t, m.IsHumanTurn())
}

func TestMighty_IsHuman_TurnHelpers(t *testing.T) {
	m := newTestMighty()
	m.Reset()
	cases := []struct {
		name    string
		set     func(int)
		check   func() bool
		humanIs int
	}{
		{"bid", m.SetBidPlayerIdx, m.IsHumanBidTurn, 0},
		{"declare", m.SetDeclarerIdx, m.IsHumanDeclareTurn, 0},
		{"exchange", m.SetDeclarerIdx, m.IsHumanExchangeTurn, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name+"/human", func(t *testing.T) {
			tc.set(tc.humanIs)
			assert.True(t, tc.check())
		})
		t.Run(tc.name+"/cpu", func(t *testing.T) {
			tc.set(1)
			assert.False(t, tc.check())
		})
		t.Run(tc.name+"/oob_low", func(t *testing.T) {
			tc.set(-1)
			assert.False(t, tc.check())
		})
		t.Run(tc.name+"/oob_high", func(t *testing.T) {
			tc.set(domain.MightyPlayerCnt)
			assert.False(t, tc.check())
		})
	}
}

func TestMighty_GetTrumpCards(t *testing.T) {
	m := newTestMighty()
	assert.NotNil(t, m.GetTrumpCards())
}

func TestMighty_GetActionLog(t *testing.T) {
	m := newTestMighty()
	m.Reset()
	require.NoError(t, m.PlayerBid(15, false))
	assert.NotEmpty(t, m.GetActionLog())
}

// --- N. Play-time misc ---

func TestMighty_PlayerPlay_validations(t *testing.T) {
	t.Run("error when game ended", func(t *testing.T) {
		m := setupForPlay(t, domain.CardDesignHeart)
		m.SetGameEndFlag(true)
		replaceHand(m.GetPlayer(0), mightyCard(domain.CardDesignSpade, 5))
		err := m.PlayerPlay(0)
		assert.True(t, errors.Is(err, domain.ErrGameEnded))
	})

	t.Run("error when wrong phase", func(t *testing.T) {
		m := setupForPlay(t, domain.CardDesignHeart)
		m.SetPhase(domain.MightyPhaseBid)
		err := m.PlayerPlay(0)
		assert.True(t, errors.Is(err, domain.ErrWrongPhase))
	})

	t.Run("error when CPU turn", func(t *testing.T) {
		m := setupForPlay(t, domain.CardDesignHeart)
		m.SetCurrentPlayerIdx(1)
		err := m.PlayerPlay(0)
		assert.True(t, errors.Is(err, domain.ErrNotHumanTurn))
	})

	t.Run("error when card index out of range", func(t *testing.T) {
		m := setupForPlay(t, domain.CardDesignHeart)
		replaceHand(m.GetPlayer(0), mightyCard(domain.CardDesignSpade, 5))
		err := m.PlayerPlay(99)
		assert.True(t, errors.Is(err, domain.ErrInvalidCard))
	})

	t.Run("follow-suit violation", func(t *testing.T) {
		m := setupForPlay(t, domain.CardDesignHeart)
		m.SetCurrentTrick([]*domain.MightyTrickCard{
			{PlayerIdx: 1, Card: mightyCard(domain.CardDesignClover, 5)},
		})
		m.SetCurrentPlayerIdx(0)
		replaceHand(m.GetPlayer(0),
			mightyCard(domain.CardDesignClover, 9), // lead suit
			mightyCard(domain.CardDesignDiamond, 2),
		)
		err := m.PlayerPlay(1) // try the off-suit
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})
}

func TestMighty_CpuPlay_guards(t *testing.T) {
	t.Run("no-op when game ended", func(t *testing.T) {
		m := setupForPlay(t, domain.CardDesignHeart)
		m.SetGameEndFlag(true)
		m.SetCurrentPlayerIdx(1)
		replaceHand(m.GetPlayer(1), mightyCard(domain.CardDesignSpade, 5))
		m.CpuPlay()
		assert.Equal(t, 1, m.GetPlayer(1).GetCardsSize(), "no card should be removed")
	})

	t.Run("no-op when wrong phase", func(t *testing.T) {
		m := setupForPlay(t, domain.CardDesignHeart)
		m.SetPhase(domain.MightyPhaseBid)
		m.SetCurrentPlayerIdx(1)
		replaceHand(m.GetPlayer(1), mightyCard(domain.CardDesignSpade, 5))
		m.CpuPlay()
		assert.Equal(t, 1, m.GetPlayer(1).GetCardsSize())
	})

	t.Run("no-op when human turn", func(t *testing.T) {
		m := setupForPlay(t, domain.CardDesignHeart)
		m.SetCurrentPlayerIdx(0)
		replaceHand(m.GetPlayer(0), mightyCard(domain.CardDesignSpade, 5))
		m.CpuPlay()
		assert.Equal(t, 1, m.GetPlayer(0).GetCardsSize())
	})
}

func TestMighty_NextTrick(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignHeart)
	m.SetPhase(domain.MightyPhaseTrickEnd)
	m.SetLeadPlayerIdx(3)
	m.SetTrickNumber(4)
	m.NextTrick()
	assert.Equal(t, domain.MightyPhasePlay, m.GetPhase())
	assert.Equal(t, 3, m.GetCurrentPlayerIdx())
	assert.Equal(t, 5, m.GetTrickNumber())
	assert.Nil(t, m.GetCurrentTrick())
}

func TestMighty_NextTrick_noOp_wrongPhase(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignHeart)
	m.SetPhase(domain.MightyPhasePlay)
	m.SetTrickNumber(4)
	m.NextTrick()
	assert.Equal(t, 4, m.GetTrickNumber(), "should not advance trick when not in TrickEnd")
}

func TestMighty_ResolveTrick_advancesPhase_atFinalTrick(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignHeart)
	m.SetPhase(domain.MightyPhaseTrickEnd)
	m.SetTrickNumber(domain.MightyTricksPerRound)
	m.SetCurrentTrick([]*domain.MightyTrickCard{
		{PlayerIdx: 0, Card: mightyCard(domain.CardDesignHeart, 1)},
		{PlayerIdx: 1, Card: mightyCard(domain.CardDesignHeart, 2)},
		{PlayerIdx: 2, Card: mightyCard(domain.CardDesignHeart, 3)},
		{PlayerIdx: 3, Card: mightyCard(domain.CardDesignHeart, 4)},
		{PlayerIdx: 4, Card: mightyCard(domain.CardDesignHeart, 5)},
	})
	m.ResolveTrick()
	assert.Equal(t, domain.MightyPhaseRoundEnd, m.GetPhase())
}

func TestMighty_ResolveTrick_noOp_wrongPhase(t *testing.T) {
	m := setupForPlay(t, domain.CardDesignHeart)
	m.SetPhase(domain.MightyPhasePlay) // not TrickEnd
	m.SetCurrentTrick([]*domain.MightyTrickCard{
		{PlayerIdx: 0, Card: mightyCard(domain.CardDesignHeart, 1)},
	})
	m.ResolveTrick()
	assert.Equal(t, domain.MightyPhasePlay, m.GetPhase())
}

func TestMighty_TrickCard_JSONRoundTrip(t *testing.T) {
	tc := &domain.MightyTrickCard{
		PlayerIdx:      2,
		Card:           mightyCard(domain.CardDesignJoker, 1),
		IsJokerLead:    true,
		LeadDemandSuit: domain.CardDesignHeart,
	}
	data, err := json.Marshal(tc)
	require.NoError(t, err)
	var out domain.MightyTrickCard
	require.NoError(t, json.Unmarshal(data, &out))
	assert.Equal(t, tc.PlayerIdx, out.PlayerIdx)
	assert.Equal(t, tc.IsJokerLead, out.IsJokerLead)
	assert.Equal(t, tc.LeadDemandSuit, out.LeadDemandSuit)
	assert.Equal(t, tc.Card.GetDesign(), out.Card.GetDesign())
	assert.Equal(t, tc.Card.GetValue(), out.Card.GetValue())
}
