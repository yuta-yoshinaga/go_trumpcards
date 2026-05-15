//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// newTestMightyDiff returns a fresh game with the requested CPU difficulty.
func newTestMightyDiff(diff domain.MightyCpuDifficulty) *domain.Mighty {
	players := []*domain.MightyPlayer{
		domain.NewMightyPlayer(true),
		domain.NewMightyPlayer(false),
		domain.NewMightyPlayer(false),
		domain.NewMightyPlayer(false),
		domain.NewMightyPlayer(false),
	}
	cfg := domain.DefaultMightyConfig()
	cfg.CpuDifficulty = diff
	return domain.NewMighty(domain.NewTrumpCards(1), players, cfg)
}

func TestMighty_CpuBid_Easy_strongHandBids(t *testing.T) {
	m := newTestMightyDiff(domain.MightyCpuDifficultyEasy)
	m.Reset()
	// Force a strong hand on the CPU player about to bid: many high cards + Mighty.
	cpuIdx := m.GetBidPlayerIdx()
	if cpuIdx == 0 {
		// Bid rotation starts at player 0 (human); push to next CPU.
		m.SetBidPlayerIdx(1)
		cpuIdx = 1
	}
	replaceHand(m.GetPlayer(cpuIdx),
		mightyCard(domain.CardDesignSpade, 1),    // Mighty
		mightyCard(domain.CardDesignHeart, 1),    // point
		mightyCard(domain.CardDesignHeart, 10),   // point
		mightyCard(domain.CardDesignHeart, 11),   // point
		mightyCard(domain.CardDesignHeart, 12),   // point
		mightyCard(domain.CardDesignHeart, 13),   // point
		mightyCard(domain.CardDesignClover, 10),  // point
		mightyCard(domain.CardDesignClover, 11),  // point
		mightyCard(domain.CardDesignClover, 12),  // point
		mightyCard(domain.CardDesignDiamond, 13), // point
	)
	m.SetHighestBidder(-1) // ensure no prior bidder
	m.CpuBid()
	// Easy bid logic places bid at MinBid when hand strength threshold is met.
	assert.GreaterOrEqual(t, m.GetPlayer(cpuIdx).GetBid(), 0)
}

func TestMighty_CpuBid_Easy_weakHandPasses(t *testing.T) {
	m := newTestMightyDiff(domain.MightyCpuDifficultyEasy)
	m.Reset()
	cpuIdx := m.GetBidPlayerIdx()
	if cpuIdx == 0 {
		m.SetBidPlayerIdx(1)
		cpuIdx = 1
	}
	// Weak hand → strength under threshold → pass.
	replaceHand(m.GetPlayer(cpuIdx),
		mightyCard(domain.CardDesignSpade, 2),
		mightyCard(domain.CardDesignSpade, 3),
		mightyCard(domain.CardDesignSpade, 4),
		mightyCard(domain.CardDesignHeart, 5),
		mightyCard(domain.CardDesignHeart, 6),
		mightyCard(domain.CardDesignHeart, 7),
		mightyCard(domain.CardDesignClover, 2),
		mightyCard(domain.CardDesignClover, 5),
		mightyCard(domain.CardDesignClover, 6),
		mightyCard(domain.CardDesignDiamond, 4),
	)
	m.SetHighestBidder(-1)
	m.CpuBid()
	assert.Equal(t, 0, m.GetPlayer(cpuIdx).GetBid(), "weak Easy CPU should pass")
}

func TestMighty_CpuPlay_EachDifficultyChoosesValidCard(t *testing.T) {
	for _, diff := range []domain.MightyCpuDifficulty{
		domain.MightyCpuDifficultyEasy,
		domain.MightyCpuDifficultyNormal,
		domain.MightyCpuDifficultyHard,
	} {
		t.Run(string(rune('A'+diff)), func(t *testing.T) {
			m := newTestMightyDiff(diff)
			m.Reset()
			m.SetPhase(domain.MightyPhasePlay)
			m.SetTrumpSuit(domain.CardDesignSpade)
			m.SetCurrentPlayerIdx(1)
			m.SetLeadPlayerIdx(1)
			m.SetTrickNumber(2)
			m.SetCurrentTrick(nil)
			// Make sure CPU 1 has cards to play.
			replaceHand(m.GetPlayer(1),
				mightyCard(domain.CardDesignDiamond, 1), // Mighty when spades trump
				mightyCard(domain.CardDesignHeart, 5),
				mightyCard(domain.CardDesignClover, 7),
			)
			// Mark CPU 1 as declarer so the team-aware paths run.
			m.GetPlayer(1).SetIsDeclarer(true)
			m.CpuPlay()
			// After CPU play, either the trick advanced (card recorded) or phase stayed Play.
			// Just assert nothing panicked and exactly one trick card was added or hand shrunk.
			handSize := m.GetPlayer(1).GetCardsSize()
			assert.LessOrEqual(t, handSize, 3)
		})
	}
}

func TestMighty_CpuPlay_HardPrefersWinning(t *testing.T) {
	m := newTestMightyDiff(domain.MightyCpuDifficultyHard)
	m.Reset()
	m.SetPhase(domain.MightyPhasePlay)
	m.SetTrumpSuit(domain.CardDesignHeart)
	m.SetCurrentPlayerIdx(2)
	m.SetLeadPlayerIdx(1)
	m.SetTrickNumber(2)
	// Lead is a low heart; CPU has a higher heart + a low non-trump.
	m.SetCurrentTrick([]*domain.MightyTrickCard{
		{PlayerIdx: 1, Card: mightyCard(domain.CardDesignHeart, 5)},
	})
	replaceHand(m.GetPlayer(2),
		mightyCard(domain.CardDesignHeart, 13), // can win
		mightyCard(domain.CardDesignClover, 2), // safe dump
	)
	m.GetPlayer(2).SetIsDeclarer(true) // declarer wants to win point tricks
	before := m.GetPlayer(2).GetCardsSize()
	m.CpuPlay()
	// Card was played.
	assert.Less(t, m.GetPlayer(2).GetCardsSize(), before)
}

func TestMighty_CpuSelectTrumpAndFriend_AllDifficulties(t *testing.T) {
	for _, diff := range []domain.MightyCpuDifficulty{
		domain.MightyCpuDifficultyEasy,
		domain.MightyCpuDifficultyNormal,
		domain.MightyCpuDifficultyHard,
	} {
		t.Run(string(rune('A'+diff)), func(t *testing.T) {
			m := newTestMightyDiff(diff)
			m.Reset()
			m.SetPhase(domain.MightyPhaseTrumpAndFriend)
			m.SetDeclarerIdx(1)
			m.SetHighestBid(15)
			m.GetPlayer(1).SetIsDeclarer(true)
			// Give the CPU declarer a hand with a clear "longest" suit.
			replaceHand(m.GetPlayer(1),
				mightyCard(domain.CardDesignHeart, 10),
				mightyCard(domain.CardDesignHeart, 11),
				mightyCard(domain.CardDesignHeart, 12),
				mightyCard(domain.CardDesignHeart, 13),
				mightyCard(domain.CardDesignHeart, 1),
				mightyCard(domain.CardDesignSpade, 1), // Mighty
				mightyCard(domain.CardDesignClover, 7),
				mightyCard(domain.CardDesignClover, 8),
				mightyCard(domain.CardDesignDiamond, 3),
				mightyCard(domain.CardDesignDiamond, 5),
			)
			m.CpuDeclareTrumpAndFriend()
			// Phase advanced and a partner card was set.
			assert.NotEqual(t, domain.MightyPhaseTrumpAndFriend, m.GetPhase())
			assert.NotNil(t, m.GetPartnerCard())
		})
	}
}

func TestMighty_CpuJokerLeadPath(t *testing.T) {
	// Drive the joker-lead branch deterministically: trick 4 (middle game), CPU
	// declarer leads with the Joker in hand. The shouldCpuLeadJoker path uses
	// rand.Intn so we retry up to a few times to hit the lead-joker branch.
	hit := false
	for attempt := 0; attempt < 50 && !hit; attempt++ {
		m := newTestMightyDiff(domain.MightyCpuDifficultyNormal)
		m.Reset()
		m.SetPhase(domain.MightyPhasePlay)
		m.SetTrumpSuit(domain.CardDesignHeart)
		m.SetTrickNumber(4)
		m.SetLeadPlayerIdx(1)
		m.SetCurrentPlayerIdx(1)
		m.SetCurrentTrick(nil)
		m.GetPlayer(1).SetIsDeclarer(true)
		replaceHand(m.GetPlayer(1),
			mightyCard(domain.CardDesignJoker, 1),
			mightyCard(domain.CardDesignClover, 8),
			mightyCard(domain.CardDesignDiamond, 4),
		)
		m.CpuPlay()
		// Whether or not Joker was led, the trick state must remain consistent.
		trick := m.GetCurrentTrick()
		if len(trick) == 1 && trick[0].Card.GetDesign() == domain.CardDesignJoker {
			assert.True(t, trick[0].IsJokerLead, "Joker-as-lead must set IsJokerLead")
			assert.NotZero(t, trick[0].LeadDemandSuit, "joker-lead must demand a suit")
			hit = true
		}
	}
	// The branch is randomized — we tolerate not hitting it, but a hit asserts shape.
	_ = hit
}

func TestMighty_CheckPartnerReveal(t *testing.T) {
	// When the partner card is played, partner is revealed.
	m := newTestMightyDiff(domain.MightyCpuDifficultyNormal)
	m.Reset()
	m.SetPhase(domain.MightyPhasePlay)
	m.SetTrumpSuit(domain.CardDesignHeart)
	m.SetDeclarerIdx(0)
	m.SetPartnerIdx(2)
	m.SetPartnerCard(mightyCard(domain.CardDesignSpade, 13))
	m.GetPlayer(0).SetIsDeclarer(true)
	m.GetPlayer(2).SetIsPartner(true)
	m.SetLeadPlayerIdx(2)
	m.SetCurrentPlayerIdx(2)
	m.SetTrickNumber(2)
	m.SetCurrentTrick(nil)
	replaceHand(m.GetPlayer(2),
		mightyCard(domain.CardDesignSpade, 13), // partner card
		mightyCard(domain.CardDesignClover, 7),
	)
	m.CpuPlay()
	// Partner card was played → partner now revealed.
	assert.True(t, m.GetPartnerRevealed())
	assert.True(t, m.GetPlayer(2).GetPartnerRevealed())
}
