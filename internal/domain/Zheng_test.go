package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- helpers (zheng-prefixed: domain tests share one package) ---

func makeZhengTestPlayers() []*ZhengPlayer {
	return []*ZhengPlayer{
		NewZhengPlayer(true),
		NewZhengPlayer(false),
		NewZhengPlayer(false),
		NewZhengPlayer(false),
	}
}

func newZhengTestGame() *Zheng {
	return NewZheng(NewTrumpCards(ZhengJokerCount), makeZhengTestPlayers(), DefaultZhengConfig())
}

func zhengCard(value, design int) *Card {
	return NewCard(design, value, false)
}

func zhengSmallJoker() *Card { return NewCard(CardDesignJoker, 1, false) }

func zhengBigJoker() *Card { return NewCard(CardDesignJoker, 2, false) }

// --- zhengRankStrength ---

func TestZhengRankStrength(t *testing.T) {
	// 3 weakest ... A, 2, small joker, big joker strongest
	assert.Equal(t, 0, zhengRankStrength(zhengCard(3, CardDesignSpade)))
	assert.Equal(t, 7, zhengRankStrength(zhengCard(10, CardDesignHeart)))
	assert.Equal(t, 8, zhengRankStrength(zhengCard(11, CardDesignClover)))
	assert.Equal(t, 9, zhengRankStrength(zhengCard(12, CardDesignDiamond)))
	assert.Equal(t, 10, zhengRankStrength(zhengCard(13, CardDesignSpade)))
	assert.Equal(t, 11, zhengRankStrength(zhengCard(1, CardDesignSpade)))
	assert.Equal(t, 12, zhengRankStrength(zhengCard(2, CardDesignSpade)))
	assert.Equal(t, 13, zhengRankStrength(zhengSmallJoker()))
	assert.Equal(t, 14, zhengRankStrength(zhengBigJoker()))

	// Suits are irrelevant: same value, different suit => identical strength
	assert.Equal(t,
		zhengRankStrength(zhengCard(5, CardDesignSpade)),
		zhengRankStrength(zhengCard(5, CardDesignHeart)))
	// Joker design is checked BEFORE value (value 1 is also Ace)
	assert.NotEqual(t,
		zhengRankStrength(zhengCard(1, CardDesignSpade)),
		zhengRankStrength(zhengSmallJoker()))
}

// --- zhengClassifyPlay ---

func TestZhengClassifyPlay(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		assert.Equal(t, ZhengPlayInvalid, zhengClassifyPlay(nil))
	})
	t.Run("single", func(t *testing.T) {
		assert.Equal(t, ZhengPlaySingle, zhengClassifyPlay([]*Card{zhengCard(5, CardDesignSpade)}))
	})
	t.Run("single joker", func(t *testing.T) {
		assert.Equal(t, ZhengPlaySingle, zhengClassifyPlay([]*Card{zhengBigJoker()}))
	})
	t.Run("pair", func(t *testing.T) {
		assert.Equal(t, ZhengPlayPair, zhengClassifyPlay([]*Card{
			zhengCard(5, CardDesignSpade), zhengCard(5, CardDesignHeart),
		}))
	})
	t.Run("two jokers are a joker bomb, never a pair", func(t *testing.T) {
		assert.Equal(t, ZhengPlayJokerBomb, zhengClassifyPlay([]*Card{zhengSmallJoker(), zhengBigJoker()}))
		assert.Equal(t, ZhengPlayJokerBomb, zhengClassifyPlay([]*Card{zhengBigJoker(), zhengSmallJoker()}))
	})
	t.Run("joker with ace is invalid", func(t *testing.T) {
		assert.Equal(t, ZhengPlayInvalid, zhengClassifyPlay([]*Card{
			zhengSmallJoker(), zhengCard(1, CardDesignSpade),
		}))
	})
	t.Run("mismatched pair invalid", func(t *testing.T) {
		assert.Equal(t, ZhengPlayInvalid, zhengClassifyPlay([]*Card{
			zhengCard(5, CardDesignSpade), zhengCard(6, CardDesignHeart),
		}))
	})
	t.Run("triple", func(t *testing.T) {
		assert.Equal(t, ZhengPlayTriple, zhengClassifyPlay([]*Card{
			zhengCard(7, CardDesignSpade), zhengCard(7, CardDesignHeart), zhengCard(7, CardDesignClover),
		}))
	})
	t.Run("3-card straight", func(t *testing.T) {
		assert.Equal(t, ZhengPlayStraight, zhengClassifyPlay([]*Card{
			zhengCard(3, CardDesignSpade), zhengCard(4, CardDesignHeart), zhengCard(5, CardDesignClover),
		}))
	})
	t.Run("invalid 3 cards", func(t *testing.T) {
		assert.Equal(t, ZhengPlayInvalid, zhengClassifyPlay([]*Card{
			zhengCard(3, CardDesignSpade), zhengCard(4, CardDesignHeart), zhengCard(9, CardDesignClover),
		}))
	})
	t.Run("four of a kind is a bomb", func(t *testing.T) {
		assert.Equal(t, ZhengPlayBomb, zhengClassifyPlay([]*Card{
			zhengCard(9, CardDesignSpade), zhengCard(9, CardDesignHeart),
			zhengCard(9, CardDesignClover), zhengCard(9, CardDesignDiamond),
		}))
	})
	t.Run("4-card straight", func(t *testing.T) {
		assert.Equal(t, ZhengPlayStraight, zhengClassifyPlay([]*Card{
			zhengCard(3, CardDesignSpade), zhengCard(4, CardDesignHeart),
			zhengCard(5, CardDesignClover), zhengCard(6, CardDesignDiamond),
		}))
	})
	t.Run("invalid 4 cards", func(t *testing.T) {
		assert.Equal(t, ZhengPlayInvalid, zhengClassifyPlay([]*Card{
			zhengCard(3, CardDesignSpade), zhengCard(4, CardDesignHeart),
			zhengCard(5, CardDesignClover), zhengCard(9, CardDesignDiamond),
		}))
	})
	t.Run("5-card straight", func(t *testing.T) {
		assert.Equal(t, ZhengPlayStraight, zhengClassifyPlay([]*Card{
			zhengCard(3, CardDesignSpade), zhengCard(4, CardDesignHeart), zhengCard(5, CardDesignClover),
			zhengCard(6, CardDesignDiamond), zhengCard(7, CardDesignSpade),
		}))
	})
	t.Run("pair run of 3", func(t *testing.T) {
		assert.Equal(t, ZhengPlayPairRun, zhengClassifyPlay([]*Card{
			zhengCard(4, CardDesignSpade), zhengCard(4, CardDesignHeart),
			zhengCard(5, CardDesignClover), zhengCard(5, CardDesignDiamond),
			zhengCard(6, CardDesignSpade), zhengCard(6, CardDesignHeart),
		}))
	})
	t.Run("pair run of 4", func(t *testing.T) {
		assert.Equal(t, ZhengPlayPairRun, zhengClassifyPlay([]*Card{
			zhengCard(4, CardDesignSpade), zhengCard(4, CardDesignHeart),
			zhengCard(5, CardDesignClover), zhengCard(5, CardDesignDiamond),
			zhengCard(6, CardDesignSpade), zhengCard(6, CardDesignHeart),
			zhengCard(7, CardDesignClover), zhengCard(7, CardDesignDiamond),
		}))
	})
	t.Run("6-card straight", func(t *testing.T) {
		assert.Equal(t, ZhengPlayStraight, zhengClassifyPlay([]*Card{
			zhengCard(3, CardDesignSpade), zhengCard(4, CardDesignHeart), zhengCard(5, CardDesignClover),
			zhengCard(6, CardDesignDiamond), zhengCard(7, CardDesignSpade), zhengCard(8, CardDesignHeart),
		}))
	})
	t.Run("invalid 6 cards", func(t *testing.T) {
		assert.Equal(t, ZhengPlayInvalid, zhengClassifyPlay([]*Card{
			zhengCard(4, CardDesignSpade), zhengCard(4, CardDesignHeart),
			zhengCard(5, CardDesignClover), zhengCard(5, CardDesignDiamond),
			zhengCard(9, CardDesignSpade), zhengCard(9, CardDesignHeart),
		}))
	})
	t.Run("invalid 7 cards", func(t *testing.T) {
		assert.Equal(t, ZhengPlayInvalid, zhengClassifyPlay([]*Card{
			zhengCard(3, CardDesignSpade), zhengCard(4, CardDesignHeart), zhengCard(5, CardDesignClover),
			zhengCard(6, CardDesignDiamond), zhengCard(7, CardDesignSpade), zhengCard(8, CardDesignHeart),
			zhengCard(10, CardDesignSpade),
		}))
	})
}

// --- zhengCheckStraight / zhengCheckPairRun edge cases ---

func TestZhengCheckStraight(t *testing.T) {
	t.Run("too short", func(t *testing.T) {
		assert.False(t, zhengCheckStraight([]*Card{zhengCard(3, CardDesignSpade), zhengCard(4, CardDesignSpade)}))
	})
	t.Run("contains 2 is invalid", func(t *testing.T) {
		assert.False(t, zhengCheckStraight([]*Card{
			zhengCard(13, CardDesignSpade), zhengCard(1, CardDesignSpade), zhengCard(2, CardDesignSpade),
		}))
	})
	t.Run("contains joker is invalid", func(t *testing.T) {
		assert.False(t, zhengCheckStraight([]*Card{
			zhengCard(13, CardDesignSpade), zhengCard(1, CardDesignSpade), zhengSmallJoker(),
		}))
	})
	t.Run("duplicate ranks are invalid", func(t *testing.T) {
		assert.False(t, zhengCheckStraight([]*Card{
			zhengCard(4, CardDesignSpade), zhengCard(4, CardDesignHeart), zhengCard(5, CardDesignSpade),
		}))
	})
	t.Run("Q-K-A high straight", func(t *testing.T) {
		assert.True(t, zhengCheckStraight([]*Card{
			zhengCard(12, CardDesignSpade), zhengCard(13, CardDesignHeart), zhengCard(1, CardDesignClover),
		}))
	})
}

func TestZhengCheckPairRun(t *testing.T) {
	t.Run("too short", func(t *testing.T) {
		assert.False(t, zhengCheckPairRun([]*Card{
			zhengCard(4, CardDesignSpade), zhengCard(4, CardDesignHeart),
			zhengCard(5, CardDesignClover), zhengCard(5, CardDesignDiamond),
		}))
	})
	t.Run("odd size", func(t *testing.T) {
		assert.False(t, zhengCheckPairRun([]*Card{
			zhengCard(4, CardDesignSpade), zhengCard(4, CardDesignHeart),
			zhengCard(5, CardDesignClover), zhengCard(5, CardDesignDiamond),
			zhengCard(6, CardDesignSpade), zhengCard(6, CardDesignHeart),
			zhengCard(7, CardDesignClover),
		}))
	})
	t.Run("contains 2", func(t *testing.T) {
		assert.False(t, zhengCheckPairRun([]*Card{
			zhengCard(13, CardDesignSpade), zhengCard(13, CardDesignHeart),
			zhengCard(1, CardDesignSpade), zhengCard(1, CardDesignHeart),
			zhengCard(2, CardDesignSpade), zhengCard(2, CardDesignHeart),
		}))
	})
	t.Run("contains joker", func(t *testing.T) {
		assert.False(t, zhengCheckPairRun([]*Card{
			zhengCard(4, CardDesignSpade), zhengCard(4, CardDesignHeart),
			zhengCard(5, CardDesignClover), zhengCard(5, CardDesignDiamond),
			zhengSmallJoker(), zhengBigJoker(),
		}))
	})
	t.Run("not all pairs", func(t *testing.T) {
		assert.False(t, zhengCheckPairRun([]*Card{
			zhengCard(4, CardDesignSpade), zhengCard(4, CardDesignHeart), zhengCard(4, CardDesignClover),
			zhengCard(5, CardDesignDiamond), zhengCard(6, CardDesignSpade), zhengCard(6, CardDesignHeart),
		}))
	})
	t.Run("not consecutive", func(t *testing.T) {
		assert.False(t, zhengCheckPairRun([]*Card{
			zhengCard(4, CardDesignSpade), zhengCard(4, CardDesignHeart),
			zhengCard(5, CardDesignClover), zhengCard(5, CardDesignDiamond),
			zhengCard(8, CardDesignSpade), zhengCard(8, CardDesignHeart),
		}))
	})
}

// --- zhengIsPlayable ---

func TestZhengIsPlayable(t *testing.T) {
	bomb9 := []*Card{
		zhengCard(9, CardDesignSpade), zhengCard(9, CardDesignHeart),
		zhengCard(9, CardDesignClover), zhengCard(9, CardDesignDiamond),
	}
	bomb7 := []*Card{
		zhengCard(7, CardDesignSpade), zhengCard(7, CardDesignHeart),
		zhengCard(7, CardDesignClover), zhengCard(7, CardDesignDiamond),
	}
	jokerBomb := []*Card{zhengSmallJoker(), zhengBigJoker()}

	t.Run("anything playable on empty table", func(t *testing.T) {
		assert.True(t, zhengIsPlayable([]*Card{zhengCard(3, CardDesignSpade)}, nil, ZhengPlayInvalid))
		assert.True(t, zhengIsPlayable(bomb9, nil, ZhengPlayInvalid))
		assert.True(t, zhengIsPlayable(jokerBomb, nil, ZhengPlayInvalid))
	})
	t.Run("invalid play returns false", func(t *testing.T) {
		assert.False(t, zhengIsPlayable([]*Card{}, nil, ZhengPlayInvalid))
	})
	t.Run("stronger single beats weaker", func(t *testing.T) {
		table := []*Card{zhengCard(5, CardDesignHeart)}
		assert.True(t, zhengIsPlayable([]*Card{zhengCard(6, CardDesignSpade)}, table, ZhengPlaySingle))
	})
	t.Run("weaker single cannot beat stronger", func(t *testing.T) {
		table := []*Card{zhengCard(6, CardDesignSpade)}
		assert.False(t, zhengIsPlayable([]*Card{zhengCard(5, CardDesignHeart)}, table, ZhengPlaySingle))
	})
	t.Run("same rank never beats regardless of suit", func(t *testing.T) {
		table := []*Card{zhengCard(5, CardDesignSpade)}
		assert.False(t, zhengIsPlayable([]*Card{zhengCard(5, CardDesignHeart)}, table, ZhengPlaySingle))
	})
	t.Run("2 beats ace, small joker beats 2, big joker beats small joker", func(t *testing.T) {
		assert.True(t, zhengIsPlayable([]*Card{zhengCard(2, CardDesignSpade)},
			[]*Card{zhengCard(1, CardDesignHeart)}, ZhengPlaySingle))
		assert.True(t, zhengIsPlayable([]*Card{zhengSmallJoker()},
			[]*Card{zhengCard(2, CardDesignHeart)}, ZhengPlaySingle))
		assert.True(t, zhengIsPlayable([]*Card{zhengBigJoker()},
			[]*Card{zhengSmallJoker()}, ZhengPlaySingle))
	})
	t.Run("pair must beat pair", func(t *testing.T) {
		table := []*Card{zhengCard(5, CardDesignSpade), zhengCard(5, CardDesignHeart)}
		assert.True(t, zhengIsPlayable([]*Card{zhengCard(6, CardDesignSpade), zhengCard(6, CardDesignClover)}, table, ZhengPlayPair))
	})
	t.Run("single cannot beat pair", func(t *testing.T) {
		table := []*Card{zhengCard(5, CardDesignSpade), zhengCard(5, CardDesignHeart)}
		assert.False(t, zhengIsPlayable([]*Card{zhengCard(2, CardDesignHeart)}, table, ZhengPlayPair))
	})
	t.Run("longer straight cannot beat shorter", func(t *testing.T) {
		table := []*Card{zhengCard(5, CardDesignSpade), zhengCard(6, CardDesignSpade), zhengCard(7, CardDesignSpade)}
		play := []*Card{
			zhengCard(8, CardDesignSpade), zhengCard(9, CardDesignSpade),
			zhengCard(10, CardDesignSpade), zhengCard(11, CardDesignSpade),
		}
		assert.False(t, zhengIsPlayable(play, table, ZhengPlayStraight))
	})
	t.Run("stronger straight beats weaker by top rank", func(t *testing.T) {
		table := []*Card{zhengCard(5, CardDesignSpade), zhengCard(6, CardDesignSpade), zhengCard(7, CardDesignSpade)}
		play := []*Card{zhengCard(6, CardDesignHeart), zhengCard(7, CardDesignHeart), zhengCard(8, CardDesignHeart)}
		assert.True(t, zhengIsPlayable(play, table, ZhengPlayStraight))
	})
	t.Run("stronger pair run beats weaker", func(t *testing.T) {
		table := []*Card{
			zhengCard(4, CardDesignSpade), zhengCard(4, CardDesignHeart),
			zhengCard(5, CardDesignClover), zhengCard(5, CardDesignDiamond),
			zhengCard(6, CardDesignSpade), zhengCard(6, CardDesignHeart),
		}
		play := []*Card{
			zhengCard(5, CardDesignSpade), zhengCard(5, CardDesignHeart),
			zhengCard(6, CardDesignClover), zhengCard(6, CardDesignDiamond),
			zhengCard(7, CardDesignSpade), zhengCard(7, CardDesignHeart),
		}
		assert.True(t, zhengIsPlayable(play, table, ZhengPlayPairRun))
		assert.False(t, zhengIsPlayable(table, play, ZhengPlayPairRun))
	})

	// --- bombs ---
	t.Run("bomb beats any non-bomb play regardless of type and count", func(t *testing.T) {
		assert.True(t, zhengIsPlayable(bomb9, []*Card{zhengBigJoker()}, ZhengPlaySingle))
		assert.True(t, zhengIsPlayable(bomb9,
			[]*Card{zhengCard(2, CardDesignSpade), zhengCard(2, CardDesignHeart)}, ZhengPlayPair))
		assert.True(t, zhengIsPlayable(bomb9,
			[]*Card{zhengCard(12, CardDesignSpade), zhengCard(13, CardDesignHeart), zhengCard(1, CardDesignClover)},
			ZhengPlayStraight))
		assert.True(t, zhengIsPlayable(bomb9, []*Card{
			zhengCard(4, CardDesignSpade), zhengCard(4, CardDesignHeart),
			zhengCard(5, CardDesignClover), zhengCard(5, CardDesignDiamond),
			zhengCard(6, CardDesignSpade), zhengCard(6, CardDesignHeart),
		}, ZhengPlayPairRun))
	})
	t.Run("higher bomb beats lower bomb", func(t *testing.T) {
		assert.True(t, zhengIsPlayable(bomb9, bomb7, ZhengPlayBomb))
		assert.False(t, zhengIsPlayable(bomb7, bomb9, ZhengPlayBomb))
	})
	t.Run("non-bomb cannot beat a bomb", func(t *testing.T) {
		assert.False(t, zhengIsPlayable([]*Card{zhengBigJoker()}, bomb7, ZhengPlayBomb))
	})
	t.Run("joker bomb beats everything including bombs", func(t *testing.T) {
		assert.True(t, zhengIsPlayable(jokerBomb, []*Card{zhengCard(2, CardDesignHeart)}, ZhengPlaySingle))
		assert.True(t, zhengIsPlayable(jokerBomb, bomb9, ZhengPlayBomb))
	})
	t.Run("nothing beats the joker bomb", func(t *testing.T) {
		assert.False(t, zhengIsPlayable(bomb9, jokerBomb, ZhengPlayJokerBomb))
		assert.False(t, zhengIsPlayable([]*Card{zhengCard(2, CardDesignHeart)}, jokerBomb, ZhengPlayJokerBomb))
	})
}

func TestZhengPlayStrength_Invalid(t *testing.T) {
	assert.Equal(t, -1, zhengPlayStrength(nil, ZhengPlayInvalid))
}

func TestZhengCandidateStrength(t *testing.T) {
	bomb := []*Card{
		zhengCard(3, CardDesignSpade), zhengCard(3, CardDesignHeart),
		zhengCard(3, CardDesignClover), zhengCard(3, CardDesignDiamond),
	}
	single := []*Card{zhengBigJoker()}
	jokerBomb := []*Card{zhengSmallJoker(), zhengBigJoker()}
	assert.Greater(t, zhengCandidateStrength(bomb), zhengCandidateStrength(single))
	assert.Greater(t, zhengCandidateStrength(jokerBomb), zhengCandidateStrength(bomb))
}

// --- Zheng game flow ---

func TestZheng_NewDefaultZheng(t *testing.T) {
	z := NewDefaultZheng()
	assert.Equal(t, ZhengPlayerCnt, z.GetPlayerCnt())
	assert.False(t, z.GetGameEndFlag())
}

func TestZheng_Reset_UnevenDeal(t *testing.T) {
	z := NewDefaultZheng()
	z.Reset()

	// 54 cards round-robin: seats 0,1 get 14; seats 2,3 get 13.
	assert.Equal(t, 14, z.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 14, z.GetPlayer(1).GetCardsSize())
	assert.Equal(t, 13, z.GetPlayer(2).GetCardsSize())
	assert.Equal(t, 13, z.GetPlayer(3).GetCardsSize())

	totalCards := 0
	for i := 0; i < z.GetPlayerCnt(); i++ {
		totalCards += z.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 54, totalCards)
	assert.False(t, z.GetGameEndFlag())
	assert.Nil(t, z.GetTableCards())
}

func TestZheng_ResetFindsSpade3(t *testing.T) {
	z := NewDefaultZheng()
	z.Reset()

	currentPlayer := z.GetPlayer(z.GetCurrentTurn())
	found := false
	for j := 0; j < currentPlayer.GetCardsSize(); j++ {
		c := currentPlayer.GetCard(j)
		if c.GetValue() == 3 && c.GetDesign() == CardDesignSpade {
			found = true
			break
		}
	}
	assert.True(t, found, "current player should hold the 3 of spades")
}

func TestZheng_FindSpade3Holder_Fallback(t *testing.T) {
	z := newZhengTestGame()
	// no ♠3 dealt -> defaults to 0
	assert.Equal(t, 0, z.findSpade3Holder())
}

func TestZheng_PlayerPlay_LeaderMustPlay(t *testing.T) {
	z := newZhengTestGame()
	players := z.players
	players[0].AddCard(zhengCard(3, CardDesignSpade))
	players[0].AddCard(zhengCard(5, CardDesignSpade))
	players[1].AddCard(zhengCard(4, CardDesignSpade))
	players[2].AddCard(zhengCard(6, CardDesignSpade))
	players[3].AddCard(zhengCard(7, CardDesignSpade))
	z.round.currentTurn = 0

	// Cannot pass on an empty table
	assert.Error(t, z.PlayerPlay([]int{}))

	// Any card may lead (no forced ♠3 combo)
	assert.NoError(t, z.PlayerPlay([]int{1}))
}

func TestZheng_PlayerPlay_PassOnTable(t *testing.T) {
	z := newZhengTestGame()
	players := z.players
	players[0].AddCard(zhengCard(3, CardDesignSpade))
	players[1].AddCard(zhengCard(4, CardDesignSpade))
	players[2].AddCard(zhengCard(6, CardDesignSpade))
	players[3].AddCard(zhengCard(7, CardDesignSpade))
	z.round.currentTurn = 0
	z.round.tableCards = []*Card{zhengCard(10, CardDesignSpade)}
	z.round.tablePlayType = ZhengPlaySingle
	z.round.lastPlayPlayerIdx = 3

	assert.NoError(t, z.PlayerPlay([]int{}))
	require.NotNil(t, z.GetHumanAction())
	assert.Nil(t, z.GetHumanAction().PlayedCards)
	assert.Equal(t, 1, z.GetPassCount())
}

func TestZheng_PlayerPlay_InvalidCard(t *testing.T) {
	z := newZhengTestGame()
	z.players[0].AddCard(zhengCard(3, CardDesignSpade))
	z.round.currentTurn = 0
	assert.Error(t, z.PlayerPlay([]int{99}))
}

func TestZheng_PlayerPlay_InvalidPlay(t *testing.T) {
	z := newZhengTestGame()
	players := z.players
	players[0].AddCard(zhengCard(3, CardDesignSpade))
	players[0].AddCard(zhengCard(5, CardDesignSpade))
	players[1].AddCard(zhengCard(10, CardDesignSpade))
	players[2].AddCard(zhengCard(11, CardDesignSpade))
	players[3].AddCard(zhengCard(12, CardDesignSpade))
	z.round.currentTurn = 0

	// Two different values = invalid pair
	assert.Error(t, z.PlayerPlay([]int{0, 1}))
}

func TestZheng_PlayerPlay_GameEnded(t *testing.T) {
	z := newZhengTestGame()
	z.round.gameEndFlag = true
	assert.ErrorIs(t, z.PlayerPlay([]int{0}), ErrGameEnded)
}

func TestZheng_PlayerPlay_NotHumanTurn(t *testing.T) {
	z := newZhengTestGame()
	z.round.currentTurn = 1 // CPU player
	assert.ErrorIs(t, z.PlayerPlay([]int{0}), ErrNotHumanTurn)
}

func TestZheng_PlayerPlay_DuplicateIndices(t *testing.T) {
	z := newZhengTestGame()
	players := z.players
	players[0].AddCard(zhengCard(3, CardDesignSpade))
	players[0].AddCard(zhengCard(5, CardDesignSpade))
	players[1].AddCard(zhengCard(10, CardDesignSpade))
	players[2].AddCard(zhengCard(11, CardDesignSpade))
	players[3].AddCard(zhengCard(12, CardDesignSpade))
	z.round.currentTurn = 0

	assert.NoError(t, z.PlayerPlay([]int{0, 0, 0}))
	assert.Equal(t, 1, len(z.GetTableCards()))
}

func TestZheng_CpuPlay(t *testing.T) {
	z := newZhengTestGame()
	players := z.players
	players[0].AddCard(zhengCard(3, CardDesignSpade))
	players[1].AddCard(zhengCard(4, CardDesignSpade))
	players[1].AddCard(zhengCard(5, CardDesignSpade))
	players[2].AddCard(zhengCard(6, CardDesignSpade))
	players[3].AddCard(zhengCard(7, CardDesignSpade))

	z.round.tableCards = []*Card{zhengCard(3, CardDesignHeart)}
	z.round.tablePlayType = ZhengPlaySingle
	z.round.lastPlayPlayerIdx = 0
	z.round.currentTurn = 1

	z.CpuPlay()
	assert.NotNil(t, z.GetCpuActions())
}

func TestZheng_CpuPlay_GameEnded(t *testing.T) {
	z := newZhengTestGame()
	z.round.gameEndFlag = true
	z.round.currentTurn = 1
	z.CpuPlay()
	assert.Nil(t, z.GetCpuActions())
}

func TestZheng_CpuPlay_HumanTurn(t *testing.T) {
	z := newZhengTestGame()
	z.round.currentTurn = 0
	z.CpuPlay()
	assert.Nil(t, z.GetCpuActions())
}

func TestZheng_CpuPlay_PassWhenNoValidPlay(t *testing.T) {
	z := newZhengTestGame()
	players := z.players
	players[0].AddCard(zhengCard(3, CardDesignSpade))
	players[1].AddCard(zhengCard(4, CardDesignSpade)) // weaker than table
	players[2].AddCard(zhengCard(6, CardDesignSpade))
	players[3].AddCard(zhengCard(7, CardDesignSpade))

	z.round.tableCards = []*Card{zhengBigJoker()} // strongest single
	z.round.tablePlayType = ZhengPlaySingle
	z.round.lastPlayPlayerIdx = 0
	z.round.currentTurn = 1

	z.CpuPlay()
	require.Len(t, z.GetCpuActions(), 1)
	assert.Nil(t, z.GetCpuActions()[0].PlayedCards) // pass, never a phantom play
	assert.Equal(t, 1, players[1].GetCardsSize())
}

func TestZheng_FinishPlayerRankOrder(t *testing.T) {
	z := newZhengTestGame()
	players := z.players
	players[0].AddCard(zhengCard(3, CardDesignSpade))
	players[1].AddCard(zhengCard(4, CardDesignSpade))
	players[1].AddCard(zhengCard(5, CardDesignSpade))
	players[2].AddCard(zhengCard(6, CardDesignSpade))
	players[2].AddCard(zhengCard(7, CardDesignSpade))
	players[3].AddCard(zhengCard(8, CardDesignSpade))
	players[3].AddCard(zhengCard(9, CardDesignSpade))
	z.round.currentTurn = 0

	assert.NoError(t, z.PlayerPlay([]int{0}))
	assert.True(t, players[0].GetIsFinished())
	assert.Equal(t, 1, players[0].GetRank())
}

func TestZheng_PassClearAfterFinisher(t *testing.T) {
	z := newZhengTestGame()
	players := z.players
	players[0].AddCard(zhengBigJoker()) // strongest single, player 0's last card
	players[1].AddCard(zhengCard(4, CardDesignSpade))
	players[2].AddCard(zhengCard(5, CardDesignSpade))
	players[3].AddCard(zhengCard(6, CardDesignSpade))
	z.round.currentTurn = 0

	// Player 0 plays their last card and goes out — the table must NOT clear
	// immediately, so the remaining players still get a chance to respond.
	require.NoError(t, z.PlayerPlay([]int{0}))
	assert.True(t, players[0].GetIsFinished())
	assert.NotNil(t, z.GetTableCards())
	assert.Equal(t, 0, z.GetLastPlayPlayerIdx())

	// None of the active players can beat the big joker, so all pass and the
	// table clears; the lead falls to the next active player after the finisher.
	for i := 0; i < 3; i++ {
		z.CpuPlay()
	}
	assert.Nil(t, z.GetTableCards())
	assert.Equal(t, 1, z.GetCurrentTurn())
}

func TestZheng_GameEnd(t *testing.T) {
	z := newZhengTestGame()
	players := z.players
	players[1].SetIsFinished(true)
	players[1].SetRank(1)
	players[2].SetIsFinished(true)
	players[2].SetRank(2)
	players[3].SetIsFinished(true)
	players[3].SetRank(3)
	players[0].AddCard(zhengCard(3, CardDesignSpade))
	z.round.currentTurn = 0

	assert.NoError(t, z.PlayerPlay([]int{0}))
	assert.True(t, z.GetGameEndFlag())
	assert.Equal(t, 4, players[0].GetRank())
}

// --- Config ---

func TestZhengConfig_Validate(t *testing.T) {
	t.Run("valid default", func(t *testing.T) {
		assert.NoError(t, DefaultZhengConfig().Validate())
	})
	t.Run("invalid difficulty", func(t *testing.T) {
		cfg := DefaultZhengConfig()
		cfg.CpuDifficulty = 99
		assert.Error(t, cfg.Validate())
	})
}

func TestZhengConfig_JSON(t *testing.T) {
	cfg := DefaultZhengConfig()
	cfg.CpuDifficulty = ZhengDifficultyHard
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	var restored ZhengConfig
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, cfg, restored)
}

// --- Player ---

func TestZhengPlayer_SortCardsByZhengStrength(t *testing.T) {
	p := NewZhengPlayer(true)
	p.AddCard(zhengBigJoker())
	p.AddCard(zhengCard(2, CardDesignHeart))
	p.AddCard(zhengCard(3, CardDesignSpade))
	p.AddCard(zhengSmallJoker())
	p.AddCard(zhengCard(1, CardDesignHeart))
	p.SortCardsByZhengStrength()

	assert.Equal(t, 3, p.GetCard(0).GetValue())                // 3 weakest
	assert.Equal(t, 1, p.GetCard(1).GetValue())                // A
	assert.Equal(t, 2, p.GetCard(2).GetValue())                // 2
	assert.Equal(t, CardDesignJoker, p.GetCard(3).GetDesign()) // small joker
	assert.Equal(t, 1, p.GetCard(3).GetValue())
	assert.Equal(t, CardDesignJoker, p.GetCard(4).GetDesign()) // big joker strongest
	assert.Equal(t, 2, p.GetCard(4).GetValue())
}

func TestZhengPlayer_JSON(t *testing.T) {
	p := NewZhengPlayer(true)
	p.AddCard(zhengCard(5, CardDesignSpade))
	p.SetRank(1)
	data, err := json.Marshal(p)
	require.NoError(t, err)
	var restored ZhengPlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.True(t, restored.GetIsHuman())
	assert.Equal(t, 1, restored.GetRank())
	assert.Equal(t, 1, restored.GetCardsSize())
}

func TestZhengPlayer_JSON_NilRankedPlayer(t *testing.T) {
	data := []byte(`{"rp":null}`)
	var restored ZhengPlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.False(t, restored.GetIsHuman())
}

// --- Zheng JSON ---

func TestZheng_JSON(t *testing.T) {
	z := NewDefaultZheng()
	z.Reset()

	data, err := json.Marshal(z)
	require.NoError(t, err)

	var restored Zheng
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, z.GetCurrentTurn(), restored.GetCurrentTurn())
	assert.Equal(t, z.GetGameEndFlag(), restored.GetGameEndFlag())
	assert.Equal(t, z.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, 14, restored.GetPlayer(0).GetCardsSize())
}

func TestZheng_JSON_NilOptionalFields(t *testing.T) {
	z := newZhengTestGame()
	data, err := json.Marshal(z)
	require.NoError(t, err)

	var restored Zheng
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.NotNil(t, restored.trumpCards)
	assert.NotNil(t, restored.round.actionLog)
	assert.Nil(t, restored.GetTableCards())
	assert.Nil(t, restored.GetHumanAction())
}

func TestZheng_JSON_NilTrumpCards(t *testing.T) {
	z := newZhengTestGame()
	z.trumpCards = nil
	data, err := json.Marshal(z)
	require.NoError(t, err)

	var restored Zheng
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.NotNil(t, restored.trumpCards)
}

// 復元時に "tb" が非nilの空スライス ([]) でも場クリア (リード) として扱う。
func TestZheng_UnmarshalJSON_EmptyTableCardsIsLead(t *testing.T) {
	z := newZhengTestGame()
	for i := 0; i < ZhengPlayerCnt; i++ {
		z.players[i].AddCard(zhengCard(5+i, CardDesignSpade))
	}
	data, err := json.Marshal(z)
	require.NoError(t, err)

	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))
	raw["tb"] = json.RawMessage(`[]`)
	data, err = json.Marshal(raw)
	require.NoError(t, err)

	var restored Zheng
	require.NoError(t, json.Unmarshal(data, &restored))

	// リード扱いなのでパスは拒否され、任意の有効役を出せる
	require.Error(t, restored.PlayerPlay(nil))
	require.NoError(t, restored.PlayerPlay([]int{0}))
}

func TestZheng_UnmarshalJSON_Invalid(t *testing.T) {
	marshal := func(t *testing.T, z *Zheng) []byte {
		t.Helper()
		data, err := json.Marshal(z)
		require.NoError(t, err)
		return data
	}

	cases := []struct {
		name  string
		build func(t *testing.T) []byte
	}{
		{"not json", func(t *testing.T) []byte { return []byte("not json") }},
		{"oversized players", func(t *testing.T) []byte {
			huge := make([]*ZhengPlayer, zhengMaxSliceLen+1)
			for i := range huge {
				huge[i] = NewZhengPlayer(false)
			}
			return marshal(t, &Zheng{trumpCards: NewTrumpCards(ZhengJokerCount), players: huge})
		}},
		{"wrong player count", func(t *testing.T) []byte {
			z := newZhengTestGame()
			z.players = z.players[:3]
			return marshal(t, z)
		}},
		{"nil player", func(t *testing.T) []byte {
			z := newZhengTestGame()
			z.players[1] = nil
			return marshal(t, z)
		}},
		{"rank out of range", func(t *testing.T) []byte {
			z := newZhengTestGame()
			z.players[0].SetRank(9)
			return marshal(t, z)
		}},
		{"invalid hand card design", func(t *testing.T) []byte {
			z := newZhengTestGame()
			z.players[0].AddCard(NewCard(9, 5, false))
			return marshal(t, z)
		}},
		{"invalid hand card value", func(t *testing.T) []byte {
			z := newZhengTestGame()
			z.players[0].AddCard(NewCard(CardDesignSpade, 14, false))
			return marshal(t, z)
		}},
		{"invalid joker value", func(t *testing.T) []byte {
			z := newZhengTestGame()
			z.players[0].AddCard(NewCard(CardDesignJoker, 3, false))
			return marshal(t, z)
		}},
		{"currentTurn out of range", func(t *testing.T) []byte {
			z := newZhengTestGame()
			z.round.currentTurn = 7
			return marshal(t, z)
		}},
		{"lastPlayPlayerIdx out of range", func(t *testing.T) []byte {
			z := newZhengTestGame()
			z.round.lastPlayPlayerIdx = 4
			return marshal(t, z)
		}},
		{"invalid table play type", func(t *testing.T) []byte {
			z := newZhengTestGame()
			z.round.tablePlayType = 99
			return marshal(t, z)
		}},
		{"invalid table card", func(t *testing.T) []byte {
			z := newZhengTestGame()
			z.round.tableCards = []*Card{NewCard(CardDesignSpade, 0, false)}
			z.round.tablePlayType = ZhengPlaySingle
			return marshal(t, z)
		}},
		{"nil cpu action", func(t *testing.T) []byte {
			z := newZhengTestGame()
			z.round.cpuActions = []*ZhengAction{nil}
			return marshal(t, z)
		}},
		{"cpu action player out of range", func(t *testing.T) []byte {
			z := newZhengTestGame()
			z.round.cpuActions = []*ZhengAction{{PlayerIdx: 9}}
			return marshal(t, z)
		}},
		{"cpu action invalid card", func(t *testing.T) []byte {
			z := newZhengTestGame()
			z.round.cpuActions = []*ZhengAction{{PlayerIdx: 1, PlayedCards: []*Card{NewCard(7, 7, false)}}}
			return marshal(t, z)
		}},
		{"human action player out of range", func(t *testing.T) []byte {
			z := newZhengTestGame()
			z.round.humanAction = &ZhengAction{PlayerIdx: -5}
			return marshal(t, z)
		}},
		{"invalid config", func(t *testing.T) []byte {
			z := newZhengTestGame()
			z.config.CpuDifficulty = 99
			return marshal(t, z)
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var restored Zheng
			assert.Error(t, json.Unmarshal(tc.build(t), &restored))
		})
	}
}

func TestZhengAction_JSON(t *testing.T) {
	a := &ZhengAction{PlayerIdx: 2, PlayedCards: []*Card{zhengCard(5, CardDesignSpade)}}
	data, err := json.Marshal(a)
	require.NoError(t, err)
	var restored ZhengAction
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, 2, restored.PlayerIdx)
	assert.Len(t, restored.PlayedCards, 1)
}

// --- Getters ---

func TestZheng_Getters(t *testing.T) {
	z := newZhengTestGame()
	z.players[0].AddCard(zhengCard(3, CardDesignSpade))
	z.players[1].AddCard(zhengCard(4, CardDesignSpade))
	z.players[2].AddCard(zhengCard(5, CardDesignSpade))
	z.players[3].AddCard(zhengCard(6, CardDesignSpade))

	assert.Equal(t, ZhengPlayerCnt, z.GetPlayerCnt())
	assert.False(t, z.GetGameEndFlag())
	assert.Nil(t, z.GetTableCards())
	assert.Equal(t, ZhengPlayInvalid, z.GetTablePlayType())
	assert.Equal(t, -1, z.GetLastPlayPlayerIdx())
	assert.Equal(t, 0, z.GetPassCount())
	assert.Nil(t, z.GetCpuActions())
	assert.Nil(t, z.GetHumanAction())
	assert.False(t, z.HasPendingAction())
	assert.True(t, z.IsHumanTurn())
	assert.NotNil(t, z.GetConfig())
	assert.NotNil(t, z.GetActionLog())
	assert.Nil(t, z.GetPlayer(-1))
	assert.Nil(t, z.GetPlayer(99))
	assert.Equal(t, 0, z.GetCurrentTurn())
}

func TestZheng_SetConfig(t *testing.T) {
	z := NewDefaultZheng()
	cfg := ZhengConfig{CpuDifficulty: ZhengDifficultyHard}
	z.SetConfig(cfg)
	assert.Equal(t, ZhengDifficultyHard, z.GetConfig().CpuDifficulty)
}

// --- CPU difficulty variants ---

func TestZheng_CpuDifficulty(t *testing.T) {
	for _, diff := range []ZhengCpuDifficulty{ZhengDifficultyEasy, ZhengDifficultyNormal, ZhengDifficultyHard} {
		t.Run("difficulty_"+string(rune('0'+diff)), func(t *testing.T) {
			z := newZhengTestGame()
			z.config.CpuDifficulty = diff
			players := z.players
			players[0].AddCard(zhengCard(3, CardDesignSpade))
			players[1].AddCard(zhengCard(4, CardDesignSpade))
			players[1].AddCard(zhengCard(5, CardDesignSpade))
			players[2].AddCard(zhengCard(6, CardDesignSpade))
			players[3].AddCard(zhengCard(7, CardDesignSpade))

			z.round.tableCards = []*Card{zhengCard(3, CardDesignHeart)}
			z.round.tablePlayType = ZhengPlaySingle
			z.round.lastPlayPlayerIdx = 0
			z.round.currentTurn = 1

			z.CpuPlay()
			assert.NotEmpty(t, z.GetCpuActions())
		})
	}
}

func TestZheng_CpuEasyAndHard_PassWhenNoValidPlay(t *testing.T) {
	for _, diff := range []ZhengCpuDifficulty{ZhengDifficultyEasy, ZhengDifficultyHard} {
		z := newZhengTestGame()
		z.config.CpuDifficulty = diff
		z.players[0].AddCard(zhengCard(3, CardDesignSpade))
		z.players[1].AddCard(zhengCard(4, CardDesignSpade))
		z.players[2].AddCard(zhengCard(6, CardDesignSpade))
		z.players[3].AddCard(zhengCard(7, CardDesignSpade))

		z.round.tableCards = []*Card{zhengBigJoker()}
		z.round.tablePlayType = ZhengPlaySingle
		z.round.lastPlayPlayerIdx = 0
		z.round.currentTurn = 1

		z.CpuPlay()
		require.Len(t, z.GetCpuActions(), 1)
		assert.Nil(t, z.GetCpuActions()[0].PlayedCards)
	}
}

func TestZheng_CpuNormal_PlaysWeakest(t *testing.T) {
	z := newZhengTestGame()
	cpu := z.players[1]
	cpu.AddCard(zhengCard(4, CardDesignSpade))
	cpu.AddCard(zhengCard(9, CardDesignSpade))
	z.players[0].AddCard(zhengCard(5, CardDesignSpade))
	z.players[2].AddCard(zhengCard(7, CardDesignSpade))
	z.players[3].AddCard(zhengCard(8, CardDesignSpade))

	z.round.tableCards = []*Card{zhengCard(3, CardDesignHeart)}
	z.round.tablePlayType = ZhengPlaySingle
	z.round.lastPlayPlayerIdx = 0
	z.round.currentTurn = 1
	z.CpuPlay()
	require.Len(t, z.GetCpuActions(), 1)
	require.NotNil(t, z.GetCpuActions()[0].PlayedCards)
	assert.Equal(t, 4, z.GetCpuActions()[0].PlayedCards[0].GetValue())
}

func TestZheng_CpuHard_LeadsWeakest(t *testing.T) {
	z := newZhengTestGame()
	z.config.CpuDifficulty = ZhengDifficultyHard
	p := z.players[1]
	p.AddCard(zhengBigJoker())
	p.AddCard(zhengCard(4, CardDesignSpade))
	z.players[0].AddCard(zhengCard(5, CardDesignSpade))
	z.players[0].AddCard(zhengCard(6, CardDesignSpade))
	z.players[2].AddCard(zhengCard(7, CardDesignSpade))
	z.players[3].AddCard(zhengCard(8, CardDesignSpade))
	z.round.currentTurn = 1
	z.CpuPlay()
	require.Len(t, z.GetCpuActions(), 1)
	require.NotNil(t, z.GetCpuActions()[0].PlayedCards)
	assert.Equal(t, 4, z.GetCpuActions()[0].PlayedCards[0].GetValue()) // weakest single
}

func TestZheng_CpuHard_OpponentLowCardsPlaysStrong(t *testing.T) {
	z := newZhengTestGame()
	z.config.CpuDifficulty = ZhengDifficultyHard
	cpu := z.players[1]
	cpu.AddCard(zhengCard(5, CardDesignSpade))
	cpu.AddCard(zhengCard(9, CardDesignSpade))
	// opponent with only 1 card
	z.players[2].AddCard(zhengCard(7, CardDesignSpade))
	z.players[0].AddCard(zhengCard(3, CardDesignSpade))
	z.players[3].AddCard(zhengCard(8, CardDesignSpade))

	z.round.tableCards = []*Card{zhengCard(4, CardDesignHeart)}
	z.round.tablePlayType = ZhengPlaySingle
	z.round.lastPlayPlayerIdx = 0
	z.round.currentTurn = 1
	z.CpuPlay()
	require.Len(t, z.GetCpuActions(), 1)
	require.NotNil(t, z.GetCpuActions()[0].PlayedCards)
	assert.Equal(t, 9, z.GetCpuActions()[0].PlayedCards[0].GetValue()) // strongest single
}

func TestZheng_CpuHard_SavesBombWhenNotNeeded(t *testing.T) {
	z := newZhengTestGame()
	z.config.CpuDifficulty = ZhengDifficultyHard
	cpu := z.players[1]
	cpu.AddCard(zhengCard(9, CardDesignSpade))
	cpu.AddCard(zhengCard(9, CardDesignHeart))
	cpu.AddCard(zhengCard(9, CardDesignClover))
	cpu.AddCard(zhengCard(9, CardDesignDiamond))
	cpu.AddCard(zhengCard(5, CardDesignSpade))
	// every opponent still holds plenty of cards
	for _, idx := range []int{0, 2, 3} {
		z.players[idx].AddCard(zhengCard(6, CardDesignSpade))
		z.players[idx].AddCard(zhengCard(7, CardDesignSpade))
		z.players[idx].AddCard(zhengCard(8, CardDesignSpade))
	}

	z.round.tableCards = []*Card{zhengCard(4, CardDesignHeart)}
	z.round.tablePlayType = ZhengPlaySingle
	z.round.lastPlayPlayerIdx = 0
	z.round.currentTurn = 1
	z.CpuPlay()
	require.Len(t, z.GetCpuActions(), 1)
	played := z.GetCpuActions()[0].PlayedCards
	require.NotNil(t, played)
	require.Len(t, played, 1) // single, NOT the bomb
	assert.Equal(t, 5, played[0].GetValue())
}

func TestZheng_CpuBombsSingleWhenOpponentNearOut(t *testing.T) {
	z := newZhengTestGame()
	z.config.CpuDifficulty = ZhengDifficultyHard
	cpu := z.players[1]
	cpu.AddCard(zhengCard(9, CardDesignSpade))
	cpu.AddCard(zhengCard(9, CardDesignHeart))
	cpu.AddCard(zhengCard(9, CardDesignClover))
	cpu.AddCard(zhengCard(9, CardDesignDiamond))
	// opponent about to go out
	z.players[0].AddCard(zhengCard(3, CardDesignSpade))
	z.players[2].AddCard(zhengCard(6, CardDesignSpade))
	z.players[3].AddCard(zhengCard(7, CardDesignSpade))

	z.round.tableCards = []*Card{zhengBigJoker()}
	z.round.tablePlayType = ZhengPlaySingle
	z.round.lastPlayPlayerIdx = 0
	z.round.currentTurn = 1
	z.CpuPlay()
	require.Len(t, z.GetCpuActions(), 1)
	played := z.GetCpuActions()[0].PlayedCards
	require.NotNil(t, played)
	assert.Len(t, played, 4) // bomb cuts the single
}

func TestZheng_CpuPlaysJokerBombOverBomb(t *testing.T) {
	z := newZhengTestGame()
	cpu := z.players[1]
	cpu.AddCard(zhengSmallJoker())
	cpu.AddCard(zhengBigJoker())
	z.players[0].AddCard(zhengCard(3, CardDesignSpade))
	z.players[2].AddCard(zhengCard(6, CardDesignSpade))
	z.players[3].AddCard(zhengCard(7, CardDesignSpade))

	z.round.tableCards = []*Card{
		zhengCard(9, CardDesignSpade), zhengCard(9, CardDesignHeart),
		zhengCard(9, CardDesignClover), zhengCard(9, CardDesignDiamond),
	}
	z.round.tablePlayType = ZhengPlayBomb
	z.round.lastPlayPlayerIdx = 0
	z.round.currentTurn = 1
	z.CpuPlay()
	require.Len(t, z.GetCpuActions(), 1)
	played := z.GetCpuActions()[0].PlayedCards
	require.NotNil(t, played)
	assert.Len(t, played, 2)
	assert.Equal(t, CardDesignJoker, played[0].GetDesign())
}

// --- CPU set finders ---

func TestZheng_FindAllPlayableSets_EmptyHand(t *testing.T) {
	z := newZhengTestGame()
	assert.Empty(t, z.findAllPlayableSets(z.players[0]))
}

func TestZheng_FindPairRuns(t *testing.T) {
	z := newZhengTestGame()
	p := z.players[0]
	p.AddCard(zhengCard(4, CardDesignSpade))
	p.AddCard(zhengCard(4, CardDesignHeart))
	p.AddCard(zhengCard(5, CardDesignSpade))
	p.AddCard(zhengCard(5, CardDesignHeart))
	p.AddCard(zhengCard(6, CardDesignSpade))
	p.AddCard(zhengCard(6, CardDesignHeart))

	t.Run("all lengths", func(t *testing.T) {
		runs := z.findPairRuns(p, 0)
		assert.Len(t, runs, 1)
		assert.Len(t, runs[0], 6)
	})
	t.Run("exact length", func(t *testing.T) {
		assert.Len(t, z.findPairRuns(p, 6), 1)
	})
	t.Run("exact odd or too-short lengths yield nothing", func(t *testing.T) {
		assert.Nil(t, z.findPairRuns(p, 5))
		assert.Nil(t, z.findPairRuns(p, 4))
	})
}

func TestZheng_FindBombsAndJokerBomb(t *testing.T) {
	z := newZhengTestGame()
	p := z.players[0]
	p.AddCard(zhengCard(9, CardDesignSpade))
	p.AddCard(zhengCard(9, CardDesignHeart))
	p.AddCard(zhengCard(9, CardDesignClover))
	p.AddCard(zhengCard(9, CardDesignDiamond))
	p.AddCard(zhengSmallJoker())

	assert.Len(t, z.findBombs(p), 1)
	assert.Nil(t, z.findJokerBomb(p)) // only one joker

	p.AddCard(zhengBigJoker())
	assert.Len(t, z.findJokerBomb(p), 1)
}

func TestZheng_FindStraights_ExactLen(t *testing.T) {
	z := newZhengTestGame()
	p := z.players[0]
	p.AddCard(zhengCard(3, CardDesignSpade))
	p.AddCard(zhengCard(4, CardDesignSpade))
	p.AddCard(zhengCard(5, CardDesignSpade))
	p.AddCard(zhengCard(6, CardDesignSpade))
	p.AddCard(zhengCard(2, CardDesignSpade)) // never part of a straight
	got := z.findStraights(p, 3)
	assert.NotEmpty(t, got)
	for _, c := range got {
		assert.Len(t, c, 3)
		for _, idx := range c {
			assert.NotEqual(t, 2, p.GetCard(idx).GetValue())
		}
	}
}

// --- Full game simulation ---

func TestZheng_FullGame(t *testing.T) {
	z := NewDefaultZheng()
	z.Reset()

	for i := 0; i < 3000 && !z.GetGameEndFlag(); i++ {
		if z.IsHumanTurn() {
			player := z.GetPlayer(z.GetCurrentTurn())
			played := false
			for j := 0; j < player.GetCardsSize(); j++ {
				if z.PlayerPlay([]int{j}) == nil {
					played = true
					break
				}
			}
			if !played {
				if z.GetTableCards() == nil {
					require.Fail(t, "human must be able to play on empty table")
				} else {
					require.NoError(t, z.PlayerPlay([]int{}))
				}
			}
		} else {
			z.CpuPlay()
		}
	}

	assert.True(t, z.GetGameEndFlag(), "game should end within 3000 iterations")

	// Every player must end with a distinct rank 1..4.
	seen := map[int]bool{}
	for i := 0; i < z.GetPlayerCnt(); i++ {
		r := z.GetPlayer(i).GetRank()
		assert.GreaterOrEqual(t, r, 1)
		assert.LessOrEqual(t, r, 4)
		assert.False(t, seen[r], "duplicate rank")
		seen[r] = true
	}
}
