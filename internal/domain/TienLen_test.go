package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- helpers ---

func makeTienLenPlayers() []*TienLenPlayer {
	return []*TienLenPlayer{
		NewTienLenPlayer(true),
		NewTienLenPlayer(false),
		NewTienLenPlayer(false),
		NewTienLenPlayer(false),
	}
}

func newTestTienLen() *TienLen {
	return NewTienLen(NewTrumpCards(0), makeTienLenPlayers(), DefaultTienLenConfig())
}

func cardTL(value, design int) *Card {
	return NewCard(design, value, false)
}

// --- TienLenCardStrength ---

func TestTienLenCardStrength(t *testing.T) {
	// ♠3 is weakest, ♥2 is strongest
	s3 := cardTL(3, CardDesignSpade)
	h2 := cardTL(2, CardDesignHeart)
	assert.Less(t, TienLenCardStrength(s3), TienLenCardStrength(h2))

	// Same value, suit ordering: ♠ < ♣ < ♦ < ♥
	s5 := cardTL(5, CardDesignSpade)
	c5 := cardTL(5, CardDesignClover)
	d5 := cardTL(5, CardDesignDiamond)
	h5 := cardTL(5, CardDesignHeart)
	assert.Less(t, TienLenCardStrength(s5), TienLenCardStrength(c5))
	assert.Less(t, TienLenCardStrength(c5), TienLenCardStrength(d5))
	assert.Less(t, TienLenCardStrength(d5), TienLenCardStrength(h5))

	// Value ordering: 3 < 4 < ... < K < A < 2
	assert.Less(t, TienLenCardStrength(cardTL(3, CardDesignSpade)), TienLenCardStrength(cardTL(4, CardDesignSpade)))
	assert.Less(t, TienLenCardStrength(cardTL(13, CardDesignSpade)), TienLenCardStrength(cardTL(1, CardDesignSpade)))
	assert.Less(t, TienLenCardStrength(cardTL(1, CardDesignSpade)), TienLenCardStrength(cardTL(2, CardDesignSpade)))

	// Unknown design falls back to spade strength
	assert.Equal(t, 0, tienLenSuitStrength(999))
}

// --- tienLenClassifyPlay ---

func TestTienLenClassifyPlay(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		assert.Equal(t, TienLenPlayInvalid, tienLenClassifyPlay(nil))
	})
	t.Run("single", func(t *testing.T) {
		assert.Equal(t, TienLenPlaySingle, tienLenClassifyPlay([]*Card{cardTL(5, CardDesignSpade)}))
	})
	t.Run("pair", func(t *testing.T) {
		assert.Equal(t, TienLenPlayPair, tienLenClassifyPlay([]*Card{
			cardTL(5, CardDesignSpade), cardTL(5, CardDesignHeart),
		}))
	})
	t.Run("invalid pair", func(t *testing.T) {
		assert.Equal(t, TienLenPlayInvalid, tienLenClassifyPlay([]*Card{
			cardTL(5, CardDesignSpade), cardTL(6, CardDesignHeart),
		}))
	})
	t.Run("triple", func(t *testing.T) {
		assert.Equal(t, TienLenPlayTriple, tienLenClassifyPlay([]*Card{
			cardTL(7, CardDesignSpade), cardTL(7, CardDesignHeart), cardTL(7, CardDesignClover),
		}))
	})
	t.Run("3-card straight", func(t *testing.T) {
		assert.Equal(t, TienLenPlayStraight, tienLenClassifyPlay([]*Card{
			cardTL(3, CardDesignSpade), cardTL(4, CardDesignHeart), cardTL(5, CardDesignClover),
		}))
	})
	t.Run("invalid 3 cards", func(t *testing.T) {
		assert.Equal(t, TienLenPlayInvalid, tienLenClassifyPlay([]*Card{
			cardTL(3, CardDesignSpade), cardTL(4, CardDesignHeart), cardTL(9, CardDesignClover),
		}))
	})
	t.Run("4-card straight", func(t *testing.T) {
		assert.Equal(t, TienLenPlayStraight, tienLenClassifyPlay([]*Card{
			cardTL(3, CardDesignSpade), cardTL(4, CardDesignHeart), cardTL(5, CardDesignClover),
			cardTL(6, CardDesignDiamond),
		}))
	})
	t.Run("four of a kind", func(t *testing.T) {
		assert.Equal(t, TienLenPlayFourOfAKind, tienLenClassifyPlay([]*Card{
			cardTL(9, CardDesignSpade), cardTL(9, CardDesignHeart), cardTL(9, CardDesignClover),
			cardTL(9, CardDesignDiamond),
		}))
	})
	t.Run("invalid 4 cards", func(t *testing.T) {
		assert.Equal(t, TienLenPlayInvalid, tienLenClassifyPlay([]*Card{
			cardTL(3, CardDesignSpade), cardTL(4, CardDesignHeart),
			cardTL(5, CardDesignClover), cardTL(9, CardDesignDiamond),
		}))
	})
	t.Run("three pair run", func(t *testing.T) {
		assert.Equal(t, TienLenPlayThreePairRun, tienLenClassifyPlay([]*Card{
			cardTL(4, CardDesignSpade), cardTL(4, CardDesignHeart),
			cardTL(5, CardDesignClover), cardTL(5, CardDesignDiamond),
			cardTL(6, CardDesignSpade), cardTL(6, CardDesignHeart),
		}))
	})
	t.Run("6-card straight", func(t *testing.T) {
		assert.Equal(t, TienLenPlayStraight, tienLenClassifyPlay([]*Card{
			cardTL(3, CardDesignSpade), cardTL(4, CardDesignHeart), cardTL(5, CardDesignClover),
			cardTL(6, CardDesignDiamond), cardTL(7, CardDesignSpade), cardTL(8, CardDesignHeart),
		}))
	})
	t.Run("invalid 6 cards", func(t *testing.T) {
		assert.Equal(t, TienLenPlayInvalid, tienLenClassifyPlay([]*Card{
			cardTL(4, CardDesignSpade), cardTL(4, CardDesignHeart),
			cardTL(5, CardDesignClover), cardTL(5, CardDesignDiamond),
			cardTL(9, CardDesignSpade), cardTL(9, CardDesignHeart),
		}))
	})
	t.Run("invalid 7 cards", func(t *testing.T) {
		assert.Equal(t, TienLenPlayInvalid, tienLenClassifyPlay([]*Card{
			cardTL(3, CardDesignSpade), cardTL(4, CardDesignHeart), cardTL(5, CardDesignClover),
			cardTL(6, CardDesignDiamond), cardTL(7, CardDesignSpade), cardTL(8, CardDesignHeart),
			cardTL(10, CardDesignSpade),
		}))
	})
}

// --- tienLenCheckStraight / tienLenCheckThreePairRun edge cases ---

func TestTienLenCheckStraight(t *testing.T) {
	t.Run("too short", func(t *testing.T) {
		assert.False(t, tienLenCheckStraight([]*Card{cardTL(3, CardDesignSpade), cardTL(4, CardDesignSpade)}))
	})
	t.Run("contains 2 is invalid", func(t *testing.T) {
		assert.False(t, tienLenCheckStraight([]*Card{
			cardTL(13, CardDesignSpade), cardTL(1, CardDesignSpade), cardTL(2, CardDesignSpade),
		}))
	})
	t.Run("J-Q-K-A high straight", func(t *testing.T) {
		assert.True(t, tienLenCheckStraight([]*Card{
			cardTL(11, CardDesignSpade), cardTL(12, CardDesignSpade),
			cardTL(13, CardDesignSpade), cardTL(1, CardDesignSpade),
		}))
	})
}

func TestTienLenCheckThreePairRun(t *testing.T) {
	t.Run("wrong size", func(t *testing.T) {
		assert.False(t, tienLenCheckThreePairRun([]*Card{cardTL(4, CardDesignSpade), cardTL(4, CardDesignHeart)}))
	})
	t.Run("contains 2", func(t *testing.T) {
		assert.False(t, tienLenCheckThreePairRun([]*Card{
			cardTL(13, CardDesignSpade), cardTL(13, CardDesignHeart),
			cardTL(1, CardDesignSpade), cardTL(1, CardDesignHeart),
			cardTL(2, CardDesignSpade), cardTL(2, CardDesignHeart),
		}))
	})
	t.Run("not all pairs", func(t *testing.T) {
		assert.False(t, tienLenCheckThreePairRun([]*Card{
			cardTL(4, CardDesignSpade), cardTL(4, CardDesignHeart), cardTL(4, CardDesignClover),
			cardTL(5, CardDesignDiamond), cardTL(6, CardDesignSpade), cardTL(6, CardDesignHeart),
		}))
	})
	t.Run("not consecutive", func(t *testing.T) {
		assert.False(t, tienLenCheckThreePairRun([]*Card{
			cardTL(4, CardDesignSpade), cardTL(4, CardDesignHeart),
			cardTL(5, CardDesignClover), cardTL(5, CardDesignDiamond),
			cardTL(8, CardDesignSpade), cardTL(8, CardDesignHeart),
		}))
	})
}

// --- tienLenIsPlayable ---

func TestTienLenIsPlayable(t *testing.T) {
	t.Run("anything playable on empty table", func(t *testing.T) {
		cards := []*Card{cardTL(3, CardDesignSpade)}
		assert.True(t, tienLenIsPlayable(cards, nil, TienLenPlayInvalid))
	})
	t.Run("invalid play returns false", func(t *testing.T) {
		assert.False(t, tienLenIsPlayable([]*Card{}, nil, TienLenPlayInvalid))
	})
	t.Run("stronger single beats weaker", func(t *testing.T) {
		table := []*Card{cardTL(5, CardDesignSpade)}
		play := []*Card{cardTL(6, CardDesignSpade)}
		assert.True(t, tienLenIsPlayable(play, table, TienLenPlaySingle))
	})
	t.Run("weaker single cannot beat stronger", func(t *testing.T) {
		table := []*Card{cardTL(6, CardDesignSpade)}
		play := []*Card{cardTL(5, CardDesignSpade)}
		assert.False(t, tienLenIsPlayable(play, table, TienLenPlaySingle))
	})
	t.Run("same value higher suit wins", func(t *testing.T) {
		table := []*Card{cardTL(5, CardDesignSpade)}
		play := []*Card{cardTL(5, CardDesignHeart)}
		assert.True(t, tienLenIsPlayable(play, table, TienLenPlaySingle))
	})
	t.Run("pair must beat pair", func(t *testing.T) {
		table := []*Card{cardTL(5, CardDesignSpade), cardTL(5, CardDesignHeart)}
		play := []*Card{cardTL(6, CardDesignSpade), cardTL(6, CardDesignClover)}
		assert.True(t, tienLenIsPlayable(play, table, TienLenPlayPair))
	})
	t.Run("single cannot beat pair", func(t *testing.T) {
		table := []*Card{cardTL(5, CardDesignSpade), cardTL(5, CardDesignHeart)}
		play := []*Card{cardTL(2, CardDesignHeart)}
		assert.False(t, tienLenIsPlayable(play, table, TienLenPlayPair))
	})
	t.Run("longer straight cannot beat shorter", func(t *testing.T) {
		table := []*Card{cardTL(5, CardDesignSpade), cardTL(6, CardDesignSpade), cardTL(7, CardDesignSpade)}
		play := []*Card{
			cardTL(8, CardDesignSpade), cardTL(9, CardDesignSpade),
			cardTL(10, CardDesignSpade), cardTL(11, CardDesignSpade),
		}
		assert.False(t, tienLenIsPlayable(play, table, TienLenPlayStraight))
	})
	t.Run("stronger straight beats weaker", func(t *testing.T) {
		table := []*Card{cardTL(5, CardDesignSpade), cardTL(6, CardDesignSpade), cardTL(7, CardDesignSpade)}
		play := []*Card{cardTL(6, CardDesignSpade), cardTL(7, CardDesignHeart), cardTL(8, CardDesignSpade)}
		assert.True(t, tienLenIsPlayable(play, table, TienLenPlayStraight))
	})

	// --- bombs / chops ---
	singleTwo := []*Card{cardTL(2, CardDesignHeart)}
	threePairRun := []*Card{
		cardTL(4, CardDesignSpade), cardTL(4, CardDesignHeart),
		cardTL(5, CardDesignClover), cardTL(5, CardDesignDiamond),
		cardTL(6, CardDesignSpade), cardTL(6, CardDesignHeart),
	}
	strongerThreePairRun := []*Card{
		cardTL(5, CardDesignSpade), cardTL(5, CardDesignHeart),
		cardTL(6, CardDesignClover), cardTL(6, CardDesignDiamond),
		cardTL(7, CardDesignSpade), cardTL(7, CardDesignHeart),
	}
	fourKind := []*Card{
		cardTL(9, CardDesignSpade), cardTL(9, CardDesignHeart),
		cardTL(9, CardDesignClover), cardTL(9, CardDesignDiamond),
	}

	t.Run("three pair run cuts single 2", func(t *testing.T) {
		assert.True(t, tienLenIsPlayable(threePairRun, singleTwo, TienLenPlaySingle))
	})
	t.Run("four of a kind cuts single 2", func(t *testing.T) {
		assert.True(t, tienLenIsPlayable(fourKind, singleTwo, TienLenPlaySingle))
	})
	t.Run("stronger three pair run beats weaker", func(t *testing.T) {
		assert.True(t, tienLenIsPlayable(strongerThreePairRun, threePairRun, TienLenPlayThreePairRun))
	})
	t.Run("weaker three pair run cannot beat stronger", func(t *testing.T) {
		assert.False(t, tienLenIsPlayable(threePairRun, strongerThreePairRun, TienLenPlayThreePairRun))
	})
	t.Run("four of a kind beats three pair run", func(t *testing.T) {
		assert.True(t, tienLenIsPlayable(fourKind, threePairRun, TienLenPlayThreePairRun))
	})
	t.Run("three pair run cannot beat four of a kind", func(t *testing.T) {
		assert.False(t, tienLenIsPlayable(threePairRun, fourKind, TienLenPlayFourOfAKind))
	})
	t.Run("stronger four of a kind beats weaker", func(t *testing.T) {
		weakFour := []*Card{
			cardTL(7, CardDesignSpade), cardTL(7, CardDesignHeart),
			cardTL(7, CardDesignClover), cardTL(7, CardDesignDiamond),
		}
		assert.True(t, tienLenIsPlayable(fourKind, weakFour, TienLenPlayFourOfAKind))
	})
	t.Run("bomb cannot cut an unrelated single", func(t *testing.T) {
		ordinarySingle := []*Card{cardTL(13, CardDesignSpade)}
		assert.False(t, tienLenIsPlayable(fourKind, ordinarySingle, TienLenPlaySingle))
	})
	t.Run("bomb cannot cut a straight", func(t *testing.T) {
		straight := []*Card{cardTL(3, CardDesignSpade), cardTL(4, CardDesignSpade), cardTL(5, CardDesignSpade)}
		assert.False(t, tienLenIsPlayable(threePairRun, straight, TienLenPlayStraight))
	})
}

func TestTienLenPlayStrength_Invalid(t *testing.T) {
	assert.Equal(t, -1, tienLenPlayStrength(nil, TienLenPlayInvalid))
}

func TestTienLenStraightStrength(t *testing.T) {
	s1 := []*Card{cardTL(3, CardDesignSpade), cardTL(4, CardDesignHeart), cardTL(5, CardDesignClover)}
	s2 := []*Card{cardTL(4, CardDesignSpade), cardTL(5, CardDesignHeart), cardTL(6, CardDesignClover)}
	assert.Less(t,
		tienLenPlayStrength(s1, TienLenPlayStraight),
		tienLenPlayStrength(s2, TienLenPlayStraight))
}

// --- TienLen game flow ---

func TestTienLen_NewDefaultTienLen(t *testing.T) {
	tl := NewDefaultTienLen()
	assert.Equal(t, TienLenPlayerCnt, tl.GetPlayerCnt())
	assert.False(t, tl.GetGameEndFlag())
}

func TestTienLen_Reset(t *testing.T) {
	tl := NewDefaultTienLen()
	tl.Reset()

	totalCards := 0
	for i := 0; i < tl.GetPlayerCnt(); i++ {
		totalCards += tl.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 52, totalCards)
	assert.False(t, tl.GetGameEndFlag())
	assert.Nil(t, tl.GetTableCards())
}

func TestTienLen_ResetFindsSpade3(t *testing.T) {
	tl := NewDefaultTienLen()
	tl.Reset()

	currentPlayer := tl.GetPlayer(tl.GetCurrentTurn())
	found := false
	for j := 0; j < currentPlayer.GetCardsSize(); j++ {
		c := currentPlayer.GetCard(j)
		if c.GetValue() == 3 && c.GetDesign() == CardDesignSpade {
			found = true
			break
		}
	}
	assert.True(t, found, "current player should hold ♠3")
}

func TestTienLen_FindSpade3Holder_Fallback(t *testing.T) {
	tl := newTestTienLen()
	// no ♠3 dealt -> defaults to 0
	assert.Equal(t, 0, tl.findSpade3Holder())
}

func TestTienLen_PlayerPlay_Pass(t *testing.T) {
	tl := newTestTienLen()
	players := tl.players
	players[0].AddCard(cardTL(3, CardDesignSpade))
	players[0].AddCard(cardTL(5, CardDesignSpade))
	players[1].AddCard(cardTL(4, CardDesignSpade))
	players[2].AddCard(cardTL(6, CardDesignSpade))
	players[3].AddCard(cardTL(7, CardDesignSpade))
	tl.round.currentTurn = 0

	// Can't pass on empty table
	assert.Error(t, tl.PlayerPlay([]int{}))

	// Play a card first (must include ♠3 on first play)
	assert.NoError(t, tl.PlayerPlay([]int{0}))
}

func TestTienLen_PlayerPlay_FirstPlayMustIncludeSpade3(t *testing.T) {
	tl := newTestTienLen()
	players := tl.players
	players[0].AddCard(cardTL(3, CardDesignSpade))
	players[0].AddCard(cardTL(5, CardDesignSpade))
	players[1].AddCard(cardTL(4, CardDesignSpade))
	players[2].AddCard(cardTL(6, CardDesignSpade))
	players[3].AddCard(cardTL(7, CardDesignSpade))
	tl.round.currentTurn = 0
	tl.round.lastPlayPlayerIdx = -1

	// Playing card without ♠3 on first turn should fail
	assert.Error(t, tl.PlayerPlay([]int{1}))
	// Playing ♠3 should succeed
	assert.NoError(t, tl.PlayerPlay([]int{0}))
}

func TestTienLen_PlayerPlay_InvalidCard(t *testing.T) {
	tl := newTestTienLen()
	tl.players[0].AddCard(cardTL(3, CardDesignSpade))
	tl.round.currentTurn = 0
	assert.Error(t, tl.PlayerPlay([]int{99}))
}

func TestTienLen_PlayerPlay_InvalidPlay(t *testing.T) {
	tl := newTestTienLen()
	players := tl.players
	players[0].AddCard(cardTL(3, CardDesignSpade))
	players[0].AddCard(cardTL(5, CardDesignSpade))
	players[1].AddCard(cardTL(10, CardDesignSpade))
	players[2].AddCard(cardTL(11, CardDesignSpade))
	players[3].AddCard(cardTL(12, CardDesignSpade))
	tl.round.currentTurn = 0
	tl.round.lastPlayPlayerIdx = -1

	// Two different values = invalid pair
	assert.Error(t, tl.PlayerPlay([]int{0, 1}))
}

func TestTienLen_PlayerPlay_GameEnded(t *testing.T) {
	tl := newTestTienLen()
	tl.round.gameEndFlag = true
	assert.ErrorIs(t, tl.PlayerPlay([]int{0}), ErrGameEnded)
}

func TestTienLen_PlayerPlay_NotHumanTurn(t *testing.T) {
	tl := newTestTienLen()
	tl.round.currentTurn = 1 // CPU player
	assert.ErrorIs(t, tl.PlayerPlay([]int{0}), ErrNotHumanTurn)
}

func TestTienLen_PlayerPlay_DuplicateIndices(t *testing.T) {
	tl := newTestTienLen()
	players := tl.players
	players[0].AddCard(cardTL(3, CardDesignSpade))
	players[0].AddCard(cardTL(5, CardDesignSpade))
	players[1].AddCard(cardTL(10, CardDesignSpade))
	players[2].AddCard(cardTL(11, CardDesignSpade))
	players[3].AddCard(cardTL(12, CardDesignSpade))
	tl.round.currentTurn = 0
	tl.round.lastPlayPlayerIdx = -1

	assert.NoError(t, tl.PlayerPlay([]int{0, 0, 0}))
	assert.Equal(t, 1, len(tl.GetTableCards()))
}

func TestTienLen_CpuPlay(t *testing.T) {
	tl := newTestTienLen()
	players := tl.players
	players[0].AddCard(cardTL(3, CardDesignSpade))
	players[1].AddCard(cardTL(4, CardDesignSpade))
	players[1].AddCard(cardTL(5, CardDesignSpade))
	players[2].AddCard(cardTL(6, CardDesignSpade))
	players[3].AddCard(cardTL(7, CardDesignSpade))

	tl.round.tableCards = []*Card{cardTL(3, CardDesignSpade)}
	tl.round.tablePlayType = TienLenPlaySingle
	tl.round.lastPlayPlayerIdx = 0
	tl.round.currentTurn = 1

	tl.CpuPlay()
	assert.NotNil(t, tl.GetCpuActions())
}

func TestTienLen_CpuPlay_GameEnded(t *testing.T) {
	tl := newTestTienLen()
	tl.round.gameEndFlag = true
	tl.round.currentTurn = 1
	tl.CpuPlay()
	assert.Nil(t, tl.GetCpuActions())
}

func TestTienLen_CpuPlay_HumanTurn(t *testing.T) {
	tl := newTestTienLen()
	tl.round.currentTurn = 0
	tl.CpuPlay()
	assert.Nil(t, tl.GetCpuActions())
}

func TestTienLen_CpuPlay_Pass(t *testing.T) {
	tl := newTestTienLen()
	players := tl.players
	players[0].AddCard(cardTL(3, CardDesignSpade))
	players[1].AddCard(cardTL(4, CardDesignSpade)) // weaker than table
	players[2].AddCard(cardTL(6, CardDesignSpade))
	players[3].AddCard(cardTL(7, CardDesignSpade))

	tl.round.tableCards = []*Card{cardTL(2, CardDesignHeart)} // ♥2 strongest single
	tl.round.tablePlayType = TienLenPlaySingle
	tl.round.lastPlayPlayerIdx = 0
	tl.round.currentTurn = 1

	tl.CpuPlay()
	require.Len(t, tl.GetCpuActions(), 1)
	assert.Nil(t, tl.GetCpuActions()[0].PlayedCards) // pass
}

func TestTienLen_FinishPlayer(t *testing.T) {
	tl := newTestTienLen()
	players := tl.players
	players[0].AddCard(cardTL(3, CardDesignSpade))
	players[1].AddCard(cardTL(4, CardDesignSpade))
	players[1].AddCard(cardTL(5, CardDesignSpade))
	players[2].AddCard(cardTL(6, CardDesignSpade))
	players[2].AddCard(cardTL(7, CardDesignSpade))
	players[3].AddCard(cardTL(8, CardDesignSpade))
	players[3].AddCard(cardTL(9, CardDesignSpade))
	tl.round.currentTurn = 0
	tl.round.lastPlayPlayerIdx = -1

	assert.NoError(t, tl.PlayerPlay([]int{0}))
	assert.True(t, players[0].GetIsFinished())
	assert.Equal(t, 1, players[0].GetRank())
}

func TestTienLen_PassClearAfterFinisher(t *testing.T) {
	tl := newTestTienLen()
	players := tl.players
	players[0].AddCard(cardTL(2, CardDesignHeart)) // strongest single, player 0's last card
	players[1].AddCard(cardTL(4, CardDesignSpade))
	players[2].AddCard(cardTL(5, CardDesignSpade))
	players[3].AddCard(cardTL(6, CardDesignSpade))
	tl.round.currentTurn = 0
	tl.round.firstPlayDone = true
	tl.round.lastPlayPlayerIdx = -1

	// Player 0 plays their last card and goes out — the table must NOT clear
	// immediately, so the remaining players still get a chance to respond.
	require.NoError(t, tl.PlayerPlay([]int{0}))
	assert.True(t, players[0].GetIsFinished())
	assert.NotNil(t, tl.GetTableCards())
	assert.Equal(t, 0, tl.GetLastPlayPlayerIdx())

	// None of the active players can beat the ♥2, so all pass and the table clears.
	for i := 0; i < 3; i++ {
		tl.CpuPlay()
	}
	assert.Nil(t, tl.GetTableCards())
}

func TestTienLen_GameEnd(t *testing.T) {
	tl := newTestTienLen()
	players := tl.players
	players[1].SetIsFinished(true)
	players[1].SetRank(1)
	players[2].SetIsFinished(true)
	players[2].SetRank(2)
	players[3].SetIsFinished(true)
	players[3].SetRank(3)
	players[0].AddCard(cardTL(3, CardDesignSpade))
	tl.round.currentTurn = 0
	tl.round.lastPlayPlayerIdx = -1

	assert.NoError(t, tl.PlayerPlay([]int{0}))
	assert.True(t, tl.GetGameEndFlag())
}

// --- Config ---

func TestTienLenConfig_Validate(t *testing.T) {
	t.Run("valid default", func(t *testing.T) {
		assert.NoError(t, DefaultTienLenConfig().Validate())
	})
	t.Run("invalid difficulty", func(t *testing.T) {
		cfg := DefaultTienLenConfig()
		cfg.CpuDifficulty = 99
		assert.Error(t, cfg.Validate())
	})
}

func TestTienLenConfig_JSON(t *testing.T) {
	cfg := DefaultTienLenConfig()
	cfg.CpuDifficulty = TienLenDifficultyHard
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	var restored TienLenConfig
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, cfg, restored)
}

// --- Player ---

func TestTienLenPlayer_SortCardsByTienLenStrength(t *testing.T) {
	p := NewTienLenPlayer(true)
	p.AddCard(cardTL(2, CardDesignHeart)) // strongest
	p.AddCard(cardTL(3, CardDesignSpade)) // weakest
	p.AddCard(cardTL(1, CardDesignHeart)) // Ace
	p.SortCardsByTienLenStrength()

	assert.Equal(t, 3, p.GetCard(0).GetValue()) // ♠3 weakest
	assert.Equal(t, 1, p.GetCard(1).GetValue()) // A
	assert.Equal(t, 2, p.GetCard(2).GetValue()) // ♥2 strongest
}

func TestTienLenPlayer_JSON(t *testing.T) {
	p := NewTienLenPlayer(true)
	p.AddCard(cardTL(5, CardDesignSpade))
	p.SetRank(1)
	data, err := json.Marshal(p)
	require.NoError(t, err)
	var restored TienLenPlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.True(t, restored.GetIsHuman())
	assert.Equal(t, 1, restored.GetRank())
	assert.Equal(t, 1, restored.GetCardsSize())
}

func TestTienLenPlayer_JSON_NilRankedPlayer(t *testing.T) {
	data := []byte(`{"rp":null}`)
	var restored TienLenPlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.False(t, restored.GetIsHuman())
}

// --- TienLen JSON ---

func TestTienLen_JSON(t *testing.T) {
	tl := NewDefaultTienLen()
	tl.Reset()

	data, err := json.Marshal(tl)
	require.NoError(t, err)

	var restored TienLen
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, tl.GetCurrentTurn(), restored.GetCurrentTurn())
	assert.Equal(t, tl.GetGameEndFlag(), restored.GetGameEndFlag())
	assert.Equal(t, tl.GetPlayerCnt(), restored.GetPlayerCnt())
}

func TestTienLen_JSON_NilFields(t *testing.T) {
	data := []byte(`{"tc":null,"pl":null,"cf":{},"ct":0,"tb":null,"tt":0,"lp":-1,"ge":false,"pc":0,"ca":null,"ha":null,"al":null}`)
	var restored TienLen
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.NotNil(t, restored.trumpCards)
	assert.NotNil(t, restored.players)
	assert.NotNil(t, restored.round.actionLog)
}

func TestTienLen_JSON_MaxSliceLen(t *testing.T) {
	huge := make([]*TienLenPlayer, tienLenMaxSliceLen+1)
	for i := range huge {
		huge[i] = NewTienLenPlayer(false)
	}
	tl := &TienLen{trumpCards: NewTrumpCards(0), players: huge}
	data, err := json.Marshal(tl)
	require.NoError(t, err)

	var restored TienLen
	assert.Error(t, json.Unmarshal(data, &restored))
}

func TestTienLenAction_JSON(t *testing.T) {
	a := &TienLenAction{PlayerIdx: 2, PlayedCards: []*Card{cardTL(5, CardDesignSpade)}}
	data, err := json.Marshal(a)
	require.NoError(t, err)
	var restored TienLenAction
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, 2, restored.PlayerIdx)
	assert.Len(t, restored.PlayedCards, 1)
}

// --- Getters ---

func TestTienLen_Getters(t *testing.T) {
	tl := newTestTienLen()
	tl.players[0].AddCard(cardTL(3, CardDesignSpade))
	tl.players[1].AddCard(cardTL(4, CardDesignSpade))
	tl.players[2].AddCard(cardTL(5, CardDesignSpade))
	tl.players[3].AddCard(cardTL(6, CardDesignSpade))

	assert.Equal(t, TienLenPlayerCnt, tl.GetPlayerCnt())
	assert.False(t, tl.GetGameEndFlag())
	assert.Nil(t, tl.GetTableCards())
	assert.Equal(t, TienLenPlayInvalid, tl.GetTablePlayType())
	assert.Equal(t, -1, tl.GetLastPlayPlayerIdx())
	assert.Equal(t, 0, tl.GetPassCount())
	assert.Nil(t, tl.GetCpuActions())
	assert.Nil(t, tl.GetHumanAction())
	assert.False(t, tl.HasPendingAction())
	assert.NotNil(t, tl.GetConfig())
	assert.NotNil(t, tl.GetActionLog())
	assert.Nil(t, tl.GetPlayer(-1))
	assert.Nil(t, tl.GetPlayer(99))
}

func TestTienLen_SetConfig(t *testing.T) {
	tl := NewDefaultTienLen()
	cfg := TienLenConfig{CpuDifficulty: TienLenDifficultyHard}
	tl.SetConfig(cfg)
	assert.Equal(t, TienLenDifficultyHard, tl.GetConfig().CpuDifficulty)
}

// --- CPU difficulty variants ---

func TestTienLen_CpuDifficulty(t *testing.T) {
	for _, diff := range []TienLenCpuDifficulty{TienLenDifficultyEasy, TienLenDifficultyNormal, TienLenDifficultyHard} {
		t.Run("difficulty_"+string(rune('0'+diff)), func(t *testing.T) {
			tl := newTestTienLen()
			tl.config.CpuDifficulty = diff
			players := tl.players
			players[0].AddCard(cardTL(3, CardDesignSpade))
			players[1].AddCard(cardTL(4, CardDesignSpade))
			players[1].AddCard(cardTL(5, CardDesignSpade))
			players[2].AddCard(cardTL(6, CardDesignSpade))
			players[3].AddCard(cardTL(7, CardDesignSpade))

			tl.round.tableCards = []*Card{cardTL(3, CardDesignSpade)}
			tl.round.tablePlayType = TienLenPlaySingle
			tl.round.lastPlayPlayerIdx = 0
			tl.round.currentTurn = 1

			tl.CpuPlay()
			assert.NotEmpty(t, tl.GetCpuActions())
		})
	}
}

func TestTienLen_CpuHard_LeadsWeakest(t *testing.T) {
	tl := newTestTienLen()
	tl.config.CpuDifficulty = TienLenDifficultyHard
	p := tl.players[1]
	p.AddCard(cardTL(2, CardDesignHeart))
	p.AddCard(cardTL(4, CardDesignSpade))
	tl.players[0].AddCard(cardTL(5, CardDesignSpade))
	tl.players[0].AddCard(cardTL(6, CardDesignSpade))
	tl.players[2].AddCard(cardTL(7, CardDesignSpade))
	tl.players[3].AddCard(cardTL(8, CardDesignSpade))
	tl.round.firstPlayDone = true
	tl.round.currentTurn = 1
	tl.CpuPlay()
	require.Len(t, tl.GetCpuActions(), 1)
	require.NotNil(t, tl.GetCpuActions()[0].PlayedCards)
	assert.Equal(t, 4, tl.GetCpuActions()[0].PlayedCards[0].GetValue()) // weakest single
}

func TestTienLen_CpuHard_OpponentLowCardsPlaysStrong(t *testing.T) {
	tl := newTestTienLen()
	tl.config.CpuDifficulty = TienLenDifficultyHard
	cpu := tl.players[1]
	cpu.AddCard(cardTL(5, CardDesignSpade))
	cpu.AddCard(cardTL(9, CardDesignSpade))
	// opponent with only 1 card
	tl.players[2].AddCard(cardTL(7, CardDesignSpade))
	tl.players[0].AddCard(cardTL(3, CardDesignSpade))
	tl.players[3].AddCard(cardTL(8, CardDesignSpade))

	tl.round.tableCards = []*Card{cardTL(4, CardDesignSpade)}
	tl.round.tablePlayType = TienLenPlaySingle
	tl.round.lastPlayPlayerIdx = 0
	tl.round.firstPlayDone = true
	tl.round.currentTurn = 1
	tl.CpuPlay()
	require.Len(t, tl.GetCpuActions(), 1)
	require.NotNil(t, tl.GetCpuActions()[0].PlayedCards)
	assert.Equal(t, 9, tl.GetCpuActions()[0].PlayedCards[0].GetValue()) // strongest single
}

// --- CPU set finders ---

func TestTienLen_FindAllPlayableSets_EmptyHand(t *testing.T) {
	tl := newTestTienLen()
	assert.Empty(t, tl.findAllPlayableSets(tl.players[0]))
}

func TestTienLen_FindThreePairRuns(t *testing.T) {
	tl := newTestTienLen()
	p := tl.players[0]
	p.AddCard(cardTL(4, CardDesignSpade))
	p.AddCard(cardTL(4, CardDesignHeart))
	p.AddCard(cardTL(5, CardDesignSpade))
	p.AddCard(cardTL(5, CardDesignHeart))
	p.AddCard(cardTL(6, CardDesignSpade))
	p.AddCard(cardTL(6, CardDesignHeart))
	runs := tl.findThreePairRuns(p)
	assert.Len(t, runs, 1)
}

func TestTienLen_FindFourOfAKinds(t *testing.T) {
	tl := newTestTienLen()
	p := tl.players[0]
	p.AddCard(cardTL(9, CardDesignSpade))
	p.AddCard(cardTL(9, CardDesignHeart))
	p.AddCard(cardTL(9, CardDesignClover))
	p.AddCard(cardTL(9, CardDesignDiamond))
	assert.Len(t, tl.findFourOfAKinds(p), 1)
}

func TestTienLen_FindStraights_ExactLen(t *testing.T) {
	tl := newTestTienLen()
	p := tl.players[0]
	p.AddCard(cardTL(3, CardDesignSpade))
	p.AddCard(cardTL(4, CardDesignSpade))
	p.AddCard(cardTL(5, CardDesignSpade))
	p.AddCard(cardTL(6, CardDesignSpade))
	got := tl.findStraights(p, 3)
	assert.NotEmpty(t, got)
	for _, c := range got {
		assert.Len(t, c, 3)
	}
}

func TestTienLen_CpuPlaysBombOnSingleTwo(t *testing.T) {
	tl := newTestTienLen()
	cpu := tl.players[1]
	cpu.AddCard(cardTL(9, CardDesignSpade))
	cpu.AddCard(cardTL(9, CardDesignHeart))
	cpu.AddCard(cardTL(9, CardDesignClover))
	cpu.AddCard(cardTL(9, CardDesignDiamond))
	tl.players[0].AddCard(cardTL(3, CardDesignSpade))
	tl.players[2].AddCard(cardTL(5, CardDesignSpade))
	tl.players[3].AddCard(cardTL(6, CardDesignSpade))

	tl.round.tableCards = []*Card{cardTL(2, CardDesignHeart)}
	tl.round.tablePlayType = TienLenPlaySingle
	tl.round.lastPlayPlayerIdx = 0
	tl.round.firstPlayDone = true
	tl.round.currentTurn = 1
	tl.CpuPlay()
	require.Len(t, tl.GetCpuActions(), 1)
	require.NotNil(t, tl.GetCpuActions()[0].PlayedCards)
	assert.Len(t, tl.GetCpuActions()[0].PlayedCards, 4) // four of a kind cut
}

// --- Full game simulation ---

func TestTienLen_FullGame(t *testing.T) {
	tl := NewDefaultTienLen()
	tl.Reset()

	for i := 0; i < 2000 && !tl.GetGameEndFlag(); i++ {
		if tl.IsHumanTurn() {
			player := tl.GetPlayer(tl.GetCurrentTurn())
			played := false
			for j := 0; j < player.GetCardsSize(); j++ {
				if tl.PlayerPlay([]int{j}) == nil {
					played = true
					break
				}
			}
			if !played {
				if tl.GetTableCards() == nil {
					require.Fail(t, "human must be able to play on empty table")
				} else {
					require.NoError(t, tl.PlayerPlay([]int{}))
				}
			}
		} else {
			tl.CpuPlay()
		}
	}

	assert.True(t, tl.GetGameEndFlag(), "game should end within 2000 iterations")
}

// #5624: CPU の着手選択はドメインにあるのに、それを人間向けに取り出す経路が
// 無く、CUI には hint が存在しなかった (Web はフロント独自のヒューリスティック)。
func TestTienLenGetHintRecommendsAPlayableSet(t *testing.T) {
	tl := newTestTienLen()
	players := tl.players
	players[0].AddCard(cardTL(3, CardDesignSpade))
	players[0].AddCard(cardTL(5, CardDesignSpade))
	players[1].AddCard(cardTL(4, CardDesignSpade))
	players[2].AddCard(cardTL(6, CardDesignSpade))
	players[3].AddCard(cardTL(7, CardDesignSpade))
	tl.round.currentTurn = 0

	hint := tl.GetHint()
	require.NotNil(t, hint)
	assert.False(t, hint.Pass, "出せる手があるならパスを勧めない")
	require.NotEmpty(t, hint.Indices)
	// **勧める手は実際に通ること。**ここがずれると、ヒント通りに打ってエラーになる。
	assert.NoError(t, tl.PlayerPlay(hint.Indices))
}

// 出せる手が無ければパスを勧める。
func TestTienLenGetHintSuggestsPassWhenNothingBeats(t *testing.T) {
	tl := newTestTienLen()
	players := tl.players
	players[0].AddCard(cardTL(3, CardDesignSpade))
	players[1].AddCard(cardTL(4, CardDesignSpade))
	players[2].AddCard(cardTL(6, CardDesignSpade))
	players[3].AddCard(cardTL(7, CardDesignSpade))
	// 場に ♥2 (最強) が出ている状態。♠3 では返せない。
	tl.round.tableCards = []*Card{cardTL(2, CardDesignHeart)}
	tl.round.tablePlayType = TienLenPlaySingle
	tl.round.lastPlayPlayerIdx = 1
	tl.round.currentTurn = 0

	hint := tl.GetHint()
	require.NotNil(t, hint)
	assert.True(t, hint.Pass)
	assert.Empty(t, hint.Indices)
}

// 人間の手番でなければヒントを出さない。相手の手札を推測させる材料になる。
func TestTienLenGetHintOnlyOnTheHumansTurn(t *testing.T) {
	tl := newTestTienLen()
	tl.players[0].AddCard(cardTL(3, CardDesignSpade))
	tl.round.currentTurn = 1

	assert.Nil(t, tl.GetHint())
}
