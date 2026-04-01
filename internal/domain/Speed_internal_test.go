//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newInternalTestSpeed() *Speed {
	tc := NewTrumpCards(0)
	players := []*SpeedPlayer{
		NewSpeedPlayer(true),
		NewSpeedPlayer(false),
	}
	return NewSpeed(tc, players, SpeedConfig{CpuDifficulty: SpeedCpuDifficultyHard})
}

// setupInternalSpeed creates a Speed game with manual state for internal testing.
func setupInternalSpeed(
	humanHand []*Card,
	cpuHand []*Card,
	cp0 *Card,
	cp1 *Card,
) *Speed {
	s := newInternalTestSpeed()
	s.phase = SpeedPhasePlay
	s.gameEndFlag = false
	s.winnerIdx = -1

	s.players[0].Reset()
	s.players[0].ResetDrawPile()
	s.players[1].Reset()
	s.players[1].ResetDrawPile()

	for _, c := range humanHand {
		s.players[0].AddCard(c)
	}
	for _, c := range cpuHand {
		s.players[1].AddCard(c)
	}
	s.centerPiles[0] = cp0
	s.centerPiles[1] = cp1
	return s
}

// --- countAdjacentCards ---

func TestSpeed_countAdjacentCards(t *testing.T) {
	tests := []struct {
		name      string
		hand      []*Card
		value     int
		playerIdx int
		want      int
	}{
		{
			"basic adjacency",
			[]*Card{
				NewCard(CardDesignSpade, 3, false),
				NewCard(CardDesignHeart, 5, false),
				NewCard(CardDesignClover, 7, false),
				NewCard(CardDesignDiamond, 9, false),
			},
			4, 0, 2, // 3 and 5 are adjacent to 4
		},
		{
			"K-A wrap",
			[]*Card{
				NewCard(CardDesignSpade, 13, false),
				NewCard(CardDesignHeart, 2, false),
			},
			1, 0, 2, // K(13) and 2 are both adjacent to A(1) -- wait, 2 is adjacent to 1 (diff=1), K is adjacent to 1 (diff=12=CardValueMax-1)
		},
		{
			"no adjacency",
			[]*Card{
				NewCard(CardDesignSpade, 10, false),
				NewCard(CardDesignHeart, 8, false),
			},
			5, 0, 0,
		},
		{
			"all adjacent",
			[]*Card{
				NewCard(CardDesignSpade, 4, false),
				NewCard(CardDesignHeart, 6, false),
			},
			5, 0, 2,
		},
		{
			"empty hand",
			[]*Card{},
			5, 0, 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var humanHand, cpuHand []*Card
			if tt.playerIdx == 0 {
				humanHand = tt.hand
				cpuHand = []*Card{NewCard(CardDesignClover, 1, false)}
			} else {
				humanHand = []*Card{NewCard(CardDesignClover, 1, false)}
				cpuHand = tt.hand
			}
			s := setupInternalSpeed(
				humanHand, cpuHand,
				NewCard(CardDesignDiamond, 5, false),
				NewCard(CardDesignSpade, 9, false),
			)
			got := s.countAdjacentCards(tt.playerIdx, tt.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- scoreHardPlay ---

func TestSpeed_scoreHardPlay_BlocksOpponent(t *testing.T) {
	// CPU has 5 and 8. Human has 7.
	// Pile 0 = 6, Pile 1 = 10.
	// Playing 5 on pile 0: newValue=5, human's 7 NOT adjacent to 5 -> opponentPlays=0
	// Playing 8 on pile 1 (not adjacent to 10, diff=2) -> not valid
	// But let's make pile 1 = 9 so 8 is adjacent.
	// Playing 8 on pile 1: newValue=8, human's 7 IS adjacent to 8 -> opponentPlays=1
	// So playing 5 on pile 0 should score higher (blocks better).
	s := setupInternalSpeed(
		[]*Card{NewCard(CardDesignHeart, 7, false)},
		[]*Card{
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignDiamond, 8, false),
		},
		NewCard(CardDesignSpade, 6, false), // pile 0
		NewCard(CardDesignHeart, 9, false), // pile 1
	)

	score5on0 := s.scoreHardPlay(s.players[1].GetCard(0)) // play 5 (6->5, human 7 not adj to 5)
	score8on1 := s.scoreHardPlay(s.players[1].GetCard(1)) // play 8 (9->8, human 7 adj to 8)
	assert.Greater(t, score5on0, score8on1, "should prefer blocking play")
}

func TestSpeed_scoreHardPlay_PrefersCombo(t *testing.T) {
	// CPU has 4, 3. Human has 10 (no adjacency concern for either).
	// Pile 0 = 5, Pile 1 = 9.
	// Playing 4 on pile 0: newValue=4, CPU's remaining 3 IS adjacent to 4 -> ownFuture=1
	// Playing 4 on pile 1: not valid (diff=5)
	// But we need two valid plays to compare. Let's adjust:
	// CPU has 6, 4. Pile 0 = 5, Pile 1 = 7.
	// Playing 6 on pile 0: newValue=6, CPU's remaining 4 NOT adjacent to 6 -> ownFuture=0
	// Playing 6 on pile 1: newValue=6, CPU's remaining 4 NOT adjacent to 6 -> ownFuture=0
	// Both same. Let me think differently.
	// CPU has 4 and 3. Pile 0 = 5, Pile 1 = 5.
	// Playing 4 on pile 0: newValue=4, remaining 3 adjacent to 4 -> ownFuture=1
	// Playing 3 on pile 0: newValue=3, remaining 4 adjacent to 3 -> ownFuture=1
	// Still equal. Let's use:
	// CPU has 4, 3, 2. Pile 0 = 5, Pile 1 = 9.
	// Playing 4 on pile 0: newValue=4, remaining [3,2]: 3 adj to 4 -> ownFuture=1
	// Playing 3 on pile 0: newValue=3, remaining [4,2]: both adj to 3 -> ownFuture=2
	// Hmm, 3 scores higher. Actually that's fine. We just need to verify combo matters.
	// Let's compare plays where combo differs.
	// CPU has 6, 4. Pile 0 = 5, Pile 1 = 5. Human has 10.
	// Playing 6 on pile 0: newValue=6, remaining 4 NOT adj to 6 (diff=2) -> ownFuture=0
	// Playing 4 on pile 0: newValue=4, remaining 6 NOT adj to 4 (diff=2) -> ownFuture=0
	// Both 0. Need a better setup.
	// CPU has 4, 3. Pile 0 = 5, Pile 1 = 12. Human has 10.
	// Playing 4 on pile 0: newValue=4, remaining 3 adj to 4 -> ownFuture=1
	// Playing 3 on pile 0: not valid (diff from 5 = 2)
	// Only one valid play. Let's make pile 1 = 3 so playing 4 on pile 1 is also valid.
	// Pile 0 = 5, Pile 1 = 3. CPU has 4, 3.
	// Playing 4 on pile 0: newValue=4, remaining 3 adj to 4 -> ownFuture=1 (and 3 adj to pile1=3? no, same value)
	// Playing 4 on pile 1: newValue=4, remaining 3 adj to 4 -> ownFuture=1
	// Still equal. OK, simplest approach:
	// CPU has 6 and 4. Pile 0 = 5. Human has 10.
	// card 0 (6): newValue=6, remaining card 1 (4) not adj to 6 -> ownFuture=0
	// card 1 (4): newValue=4, remaining card 0 (6) not adj to 4 -> ownFuture=0
	// No combo diff. Let me use 3 cards.
	// CPU has 6, 4, 3. Pile 0 = 5. Human has 10.
	// card 0 (idx=0, val=6) on pile 0: newValue=6, remaining [4,3]. 4 not adj 6, 3 not adj 6 -> own=0
	// card 1 (idx=1, val=4) on pile 0: newValue=4, remaining [6,3]. 6 not adj 4, 3 IS adj 4 -> own=1
	// So card 1 (val=4) should score higher due to combo.
	s := setupInternalSpeed(
		[]*Card{NewCard(CardDesignHeart, 10, false)}, // human: no adjacency to anything relevant
		[]*Card{
			NewCard(CardDesignClover, 6, false),
			NewCard(CardDesignDiamond, 4, false),
			NewCard(CardDesignSpade, 3, false),
		},
		NewCard(CardDesignHeart, 5, false),  // pile 0
		NewCard(CardDesignSpade, 12, false), // pile 1 (far away, irrelevant)
	)

	score6on0 := s.scoreHardPlay(s.players[1].GetCard(0)) // play 6 -> no combo
	score4on0 := s.scoreHardPlay(s.players[1].GetCard(1)) // play 4 -> enables 3 combo
	assert.Greater(t, score4on0, score6on0, "should prefer play that enables combo")
}
