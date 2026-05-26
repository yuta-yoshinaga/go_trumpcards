package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- helpers ---

func makeBigTwoPlayers() []*BigTwoPlayer {
	return []*BigTwoPlayer{
		NewBigTwoPlayer(true),
		NewBigTwoPlayer(false),
		NewBigTwoPlayer(false),
		NewBigTwoPlayer(false),
	}
}

func newTestBigTwo() *BigTwo {
	return NewBigTwo(NewTrumpCards(0), makeBigTwoPlayers(), DefaultBigTwoConfig())
}

func cardBT(value, design int) *Card {
	return NewCard(design, value, false)
}

// --- BigTwoCardStrength ---

func TestBigTwoCardStrength(t *testing.T) {
	// ♦3 is weakest, ♠2 is strongest
	d3 := cardBT(3, CardDesignDiamond)
	s2 := cardBT(2, CardDesignSpade)
	assert.Less(t, BigTwoCardStrength(d3), BigTwoCardStrength(s2))

	// Same value, suit ordering: ♦ < ♣ < ♥ < ♠
	d5 := cardBT(5, CardDesignDiamond)
	c5 := cardBT(5, CardDesignClover)
	h5 := cardBT(5, CardDesignHeart)
	s5 := cardBT(5, CardDesignSpade)
	assert.Less(t, BigTwoCardStrength(d5), BigTwoCardStrength(c5))
	assert.Less(t, BigTwoCardStrength(c5), BigTwoCardStrength(h5))
	assert.Less(t, BigTwoCardStrength(h5), BigTwoCardStrength(s5))

	// Value ordering: 3 < 4 < ... < K < A < 2
	assert.Less(t, BigTwoCardStrength(cardBT(3, CardDesignSpade)), BigTwoCardStrength(cardBT(4, CardDesignSpade)))
	assert.Less(t, BigTwoCardStrength(cardBT(13, CardDesignSpade)), BigTwoCardStrength(cardBT(1, CardDesignSpade)))
	assert.Less(t, BigTwoCardStrength(cardBT(1, CardDesignSpade)), BigTwoCardStrength(cardBT(2, CardDesignSpade)))
}

// --- bigTwoClassifyPlay ---

func TestBigTwoClassifyPlay(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		assert.Equal(t, BigTwoPlaySingle, bigTwoClassifyPlay([]*Card{cardBT(5, CardDesignSpade)}))
	})
	t.Run("pair", func(t *testing.T) {
		assert.Equal(t, BigTwoPlayPair, bigTwoClassifyPlay([]*Card{
			cardBT(5, CardDesignSpade), cardBT(5, CardDesignHeart),
		}))
	})
	t.Run("invalid pair", func(t *testing.T) {
		assert.Equal(t, BigTwoPlayInvalid, bigTwoClassifyPlay([]*Card{
			cardBT(5, CardDesignSpade), cardBT(6, CardDesignHeart),
		}))
	})
	t.Run("triple", func(t *testing.T) {
		assert.Equal(t, BigTwoPlayTriple, bigTwoClassifyPlay([]*Card{
			cardBT(7, CardDesignSpade), cardBT(7, CardDesignHeart), cardBT(7, CardDesignClover),
		}))
	})
	t.Run("invalid triple", func(t *testing.T) {
		assert.Equal(t, BigTwoPlayInvalid, bigTwoClassifyPlay([]*Card{
			cardBT(7, CardDesignSpade), cardBT(7, CardDesignHeart), cardBT(8, CardDesignClover),
		}))
	})
	t.Run("straight", func(t *testing.T) {
		assert.Equal(t, BigTwoPlayStraight, bigTwoClassifyPlay([]*Card{
			cardBT(3, CardDesignSpade), cardBT(4, CardDesignHeart), cardBT(5, CardDesignClover),
			cardBT(6, CardDesignDiamond), cardBT(7, CardDesignSpade),
		}))
	})
	t.Run("straight 10-J-Q-K-A", func(t *testing.T) {
		assert.Equal(t, BigTwoPlayStraight, bigTwoClassifyPlay([]*Card{
			cardBT(10, CardDesignSpade), cardBT(11, CardDesignHeart), cardBT(12, CardDesignClover),
			cardBT(13, CardDesignDiamond), cardBT(1, CardDesignSpade),
		}))
	})
	t.Run("no straight with 2", func(t *testing.T) {
		assert.Equal(t, BigTwoPlayInvalid, bigTwoClassifyPlay([]*Card{
			cardBT(10, CardDesignSpade), cardBT(11, CardDesignHeart), cardBT(12, CardDesignClover),
			cardBT(13, CardDesignDiamond), cardBT(2, CardDesignSpade),
		}))
	})
	t.Run("flush", func(t *testing.T) {
		assert.Equal(t, BigTwoPlayFlush, bigTwoClassifyPlay([]*Card{
			cardBT(3, CardDesignHeart), cardBT(5, CardDesignHeart), cardBT(7, CardDesignHeart),
			cardBT(9, CardDesignHeart), cardBT(11, CardDesignHeart),
		}))
	})
	t.Run("full house", func(t *testing.T) {
		assert.Equal(t, BigTwoPlayFullHouse, bigTwoClassifyPlay([]*Card{
			cardBT(5, CardDesignSpade), cardBT(5, CardDesignHeart), cardBT(5, CardDesignClover),
			cardBT(8, CardDesignDiamond), cardBT(8, CardDesignSpade),
		}))
	})
	t.Run("four of a kind", func(t *testing.T) {
		assert.Equal(t, BigTwoPlayFourOfAKind, bigTwoClassifyPlay([]*Card{
			cardBT(9, CardDesignSpade), cardBT(9, CardDesignHeart), cardBT(9, CardDesignClover),
			cardBT(9, CardDesignDiamond), cardBT(3, CardDesignSpade),
		}))
	})
	t.Run("straight flush", func(t *testing.T) {
		assert.Equal(t, BigTwoPlayStraightFlush, bigTwoClassifyPlay([]*Card{
			cardBT(3, CardDesignSpade), cardBT(4, CardDesignSpade), cardBT(5, CardDesignSpade),
			cardBT(6, CardDesignSpade), cardBT(7, CardDesignSpade),
		}))
	})
	t.Run("invalid 4 cards", func(t *testing.T) {
		assert.Equal(t, BigTwoPlayInvalid, bigTwoClassifyPlay([]*Card{
			cardBT(3, CardDesignSpade), cardBT(4, CardDesignHeart),
			cardBT(5, CardDesignClover), cardBT(6, CardDesignDiamond),
		}))
	})
	t.Run("invalid 5 cards - not a combo", func(t *testing.T) {
		assert.Equal(t, BigTwoPlayInvalid, bigTwoClassifyPlay([]*Card{
			cardBT(3, CardDesignSpade), cardBT(3, CardDesignHeart), cardBT(5, CardDesignClover),
			cardBT(7, CardDesignDiamond), cardBT(9, CardDesignSpade),
		}))
	})
}

// --- bigTwoIsPlayable ---

func TestBigTwoIsPlayable(t *testing.T) {
	t.Run("anything playable on empty table", func(t *testing.T) {
		cards := []*Card{cardBT(3, CardDesignDiamond)}
		assert.True(t, bigTwoIsPlayable(cards, nil, BigTwoPlayInvalid))
	})
	t.Run("stronger single beats weaker", func(t *testing.T) {
		table := []*Card{cardBT(5, CardDesignSpade)}
		play := []*Card{cardBT(6, CardDesignDiamond)}
		assert.True(t, bigTwoIsPlayable(play, table, BigTwoPlaySingle))
	})
	t.Run("weaker single cannot beat stronger", func(t *testing.T) {
		table := []*Card{cardBT(6, CardDesignSpade)}
		play := []*Card{cardBT(5, CardDesignSpade)}
		assert.False(t, bigTwoIsPlayable(play, table, BigTwoPlaySingle))
	})
	t.Run("same value higher suit wins", func(t *testing.T) {
		table := []*Card{cardBT(5, CardDesignDiamond)}
		play := []*Card{cardBT(5, CardDesignSpade)}
		assert.True(t, bigTwoIsPlayable(play, table, BigTwoPlaySingle))
	})
	t.Run("pair must beat pair", func(t *testing.T) {
		table := []*Card{cardBT(5, CardDesignSpade), cardBT(5, CardDesignHeart)}
		play := []*Card{cardBT(6, CardDesignDiamond), cardBT(6, CardDesignClover)}
		assert.True(t, bigTwoIsPlayable(play, table, BigTwoPlayPair))
	})
	t.Run("single cannot beat pair", func(t *testing.T) {
		table := []*Card{cardBT(5, CardDesignSpade), cardBT(5, CardDesignHeart)}
		play := []*Card{cardBT(13, CardDesignSpade)}
		assert.False(t, bigTwoIsPlayable(play, table, BigTwoPlayPair))
	})
	t.Run("flush beats straight in 5-card hierarchy", func(t *testing.T) {
		table := []*Card{
			cardBT(3, CardDesignSpade), cardBT(4, CardDesignHeart), cardBT(5, CardDesignClover),
			cardBT(6, CardDesignDiamond), cardBT(7, CardDesignSpade),
		}
		play := []*Card{
			cardBT(3, CardDesignHeart), cardBT(5, CardDesignHeart), cardBT(7, CardDesignHeart),
			cardBT(9, CardDesignHeart), cardBT(11, CardDesignHeart),
		}
		assert.True(t, bigTwoIsPlayable(play, table, BigTwoPlayStraight))
	})
	t.Run("straight cannot beat flush", func(t *testing.T) {
		table := []*Card{
			cardBT(3, CardDesignHeart), cardBT(5, CardDesignHeart), cardBT(7, CardDesignHeart),
			cardBT(9, CardDesignHeart), cardBT(11, CardDesignHeart),
		}
		play := []*Card{
			cardBT(8, CardDesignSpade), cardBT(9, CardDesignHeart), cardBT(10, CardDesignClover),
			cardBT(11, CardDesignDiamond), cardBT(12, CardDesignSpade),
		}
		assert.False(t, bigTwoIsPlayable(play, table, BigTwoPlayFlush))
	})
	t.Run("invalid play returns false", func(t *testing.T) {
		assert.False(t, bigTwoIsPlayable([]*Card{}, nil, BigTwoPlayInvalid))
	})
}

// --- BigTwo game flow ---

func TestBigTwo_NewDefaultBigTwo(t *testing.T) {
	bt := NewDefaultBigTwo()
	assert.Equal(t, BigTwoPlayerCnt, bt.GetPlayerCnt())
	assert.False(t, bt.GetGameEndFlag())
}

func TestBigTwo_Reset(t *testing.T) {
	bt := NewDefaultBigTwo()
	bt.Reset()

	totalCards := 0
	for i := 0; i < bt.GetPlayerCnt(); i++ {
		totalCards += bt.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 52, totalCards)
	assert.False(t, bt.GetGameEndFlag())
	assert.Nil(t, bt.GetTableCards())
}

func TestBigTwo_ResetFindsDiamond3(t *testing.T) {
	bt := NewDefaultBigTwo()
	bt.Reset()

	currentPlayer := bt.GetPlayer(bt.GetCurrentTurn())
	found := false
	for j := 0; j < currentPlayer.GetCardsSize(); j++ {
		c := currentPlayer.GetCard(j)
		if c.GetValue() == 3 && c.GetDesign() == CardDesignDiamond {
			found = true
			break
		}
	}
	assert.True(t, found, "current player should hold ♦3")
}

func TestBigTwo_PlayerPlay_Pass(t *testing.T) {
	bt := newTestBigTwo()
	players := bt.players
	players[0].AddCard(cardBT(3, CardDesignDiamond))
	players[0].AddCard(cardBT(5, CardDesignSpade))
	players[1].AddCard(cardBT(4, CardDesignSpade))
	players[2].AddCard(cardBT(6, CardDesignSpade))
	players[3].AddCard(cardBT(7, CardDesignSpade))
	bt.round.currentTurn = 0

	// Can't pass on empty table
	err := bt.PlayerPlay([]int{})
	assert.Error(t, err)

	// Play a card first
	err = bt.PlayerPlay([]int{0})
	assert.NoError(t, err)
}

func TestBigTwo_PlayerPlay_FirstPlayMustIncludeDiamond3(t *testing.T) {
	bt := newTestBigTwo()
	players := bt.players
	players[0].AddCard(cardBT(3, CardDesignDiamond))
	players[0].AddCard(cardBT(5, CardDesignSpade))
	players[1].AddCard(cardBT(4, CardDesignSpade))
	players[2].AddCard(cardBT(6, CardDesignSpade))
	players[3].AddCard(cardBT(7, CardDesignSpade))
	bt.round.currentTurn = 0
	bt.round.lastPlayPlayerIdx = -1

	// Playing card without ♦3 on first turn should fail
	err := bt.PlayerPlay([]int{1})
	assert.Error(t, err)

	// Playing ♦3 should succeed
	err = bt.PlayerPlay([]int{0})
	assert.NoError(t, err)
}

func TestBigTwo_PlayerPlay_InvalidCard(t *testing.T) {
	bt := newTestBigTwo()
	players := bt.players
	players[0].AddCard(cardBT(3, CardDesignDiamond))
	bt.round.currentTurn = 0

	err := bt.PlayerPlay([]int{99})
	assert.Error(t, err)
}

func TestBigTwo_PlayerPlay_InvalidPlay(t *testing.T) {
	bt := newTestBigTwo()
	players := bt.players
	players[0].AddCard(cardBT(3, CardDesignDiamond))
	players[0].AddCard(cardBT(5, CardDesignSpade))
	players[1].AddCard(cardBT(10, CardDesignSpade))
	players[2].AddCard(cardBT(11, CardDesignSpade))
	players[3].AddCard(cardBT(12, CardDesignSpade))
	bt.round.currentTurn = 0
	bt.round.lastPlayPlayerIdx = -1

	// Two different values = invalid pair
	err := bt.PlayerPlay([]int{0, 1})
	assert.Error(t, err)
}

func TestBigTwo_PlayerPlay_GameEnded(t *testing.T) {
	bt := newTestBigTwo()
	bt.round.gameEndFlag = true
	err := bt.PlayerPlay([]int{0})
	assert.ErrorIs(t, err, ErrGameEnded)
}

func TestBigTwo_PlayerPlay_NotHumanTurn(t *testing.T) {
	bt := newTestBigTwo()
	bt.round.currentTurn = 1 // CPU player
	err := bt.PlayerPlay([]int{0})
	assert.ErrorIs(t, err, ErrNotHumanTurn)
}

func TestBigTwo_PlayerPlay_DuplicateIndices(t *testing.T) {
	bt := newTestBigTwo()
	players := bt.players
	players[0].AddCard(cardBT(3, CardDesignDiamond))
	players[0].AddCard(cardBT(5, CardDesignSpade))
	players[1].AddCard(cardBT(10, CardDesignSpade))
	players[2].AddCard(cardBT(11, CardDesignSpade))
	players[3].AddCard(cardBT(12, CardDesignSpade))
	bt.round.currentTurn = 0
	bt.round.lastPlayPlayerIdx = -1

	// Duplicate indices should be deduplicated → single card play
	err := bt.PlayerPlay([]int{0, 0, 0})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(bt.GetTableCards()))
}

func TestBigTwo_CpuPlay(t *testing.T) {
	bt := newTestBigTwo()
	players := bt.players
	players[0].AddCard(cardBT(3, CardDesignDiamond))
	players[1].AddCard(cardBT(4, CardDesignSpade))
	players[1].AddCard(cardBT(5, CardDesignSpade))
	players[2].AddCard(cardBT(6, CardDesignSpade))
	players[3].AddCard(cardBT(7, CardDesignSpade))

	// Set up: table has ♦3, turn is CPU 1
	bt.round.tableCards = []*Card{cardBT(3, CardDesignDiamond)}
	bt.round.tablePlayType = BigTwoPlaySingle
	bt.round.lastPlayPlayerIdx = 0
	bt.round.currentTurn = 1

	bt.CpuPlay()
	assert.NotNil(t, bt.GetCpuActions())
}

func TestBigTwo_CpuPlay_GameEnded(t *testing.T) {
	bt := newTestBigTwo()
	bt.round.gameEndFlag = true
	bt.round.currentTurn = 1
	bt.CpuPlay()
	// Should be a no-op
	assert.Nil(t, bt.GetCpuActions())
}

func TestBigTwo_CpuPlay_HumanTurn(t *testing.T) {
	bt := newTestBigTwo()
	bt.round.currentTurn = 0
	bt.CpuPlay()
	assert.Nil(t, bt.GetCpuActions())
}

func TestBigTwo_CpuPlay_Pass(t *testing.T) {
	bt := newTestBigTwo()
	players := bt.players
	players[0].AddCard(cardBT(3, CardDesignDiamond))
	players[1].AddCard(cardBT(4, CardDesignDiamond)) // weaker than table
	players[2].AddCard(cardBT(6, CardDesignSpade))
	players[3].AddCard(cardBT(7, CardDesignSpade))

	bt.round.tableCards = []*Card{cardBT(1, CardDesignSpade)} // Ace of spades (very strong)
	bt.round.tablePlayType = BigTwoPlaySingle
	bt.round.lastPlayPlayerIdx = 0
	bt.round.currentTurn = 1

	bt.CpuPlay()
	require.Len(t, bt.GetCpuActions(), 1)
	assert.Nil(t, bt.GetCpuActions()[0].PlayedCards) // pass
}

func TestBigTwo_FinishPlayer(t *testing.T) {
	bt := newTestBigTwo()
	players := bt.players
	// Player 0 has one card, playing it should finish them
	players[0].AddCard(cardBT(3, CardDesignDiamond))
	players[1].AddCard(cardBT(4, CardDesignSpade))
	players[1].AddCard(cardBT(5, CardDesignSpade))
	players[2].AddCard(cardBT(6, CardDesignSpade))
	players[2].AddCard(cardBT(7, CardDesignSpade))
	players[3].AddCard(cardBT(8, CardDesignSpade))
	players[3].AddCard(cardBT(9, CardDesignSpade))
	bt.round.currentTurn = 0
	bt.round.lastPlayPlayerIdx = -1

	err := bt.PlayerPlay([]int{0})
	assert.NoError(t, err)
	assert.True(t, players[0].GetIsFinished())
	assert.Equal(t, 1, players[0].GetRank())
}

func TestBigTwo_GameEnd(t *testing.T) {
	bt := newTestBigTwo()
	players := bt.players
	// Set up: 3 players already finished
	players[1].SetIsFinished(true)
	players[1].SetRank(1)
	players[2].SetIsFinished(true)
	players[2].SetRank(2)
	players[3].SetIsFinished(true)
	players[3].SetRank(3)
	players[0].AddCard(cardBT(3, CardDesignDiamond))
	bt.round.currentTurn = 0
	bt.round.lastPlayPlayerIdx = -1

	err := bt.PlayerPlay([]int{0})
	assert.NoError(t, err)
	assert.True(t, bt.GetGameEndFlag())
}

func TestBigTwo_CheckPassClear(t *testing.T) {
	bt := newTestBigTwo()
	players := bt.players
	players[0].AddCard(cardBT(3, CardDesignDiamond))
	players[0].AddCard(cardBT(1, CardDesignSpade))
	players[1].AddCard(cardBT(4, CardDesignDiamond))
	players[1].AddCard(cardBT(5, CardDesignDiamond))
	players[2].AddCard(cardBT(6, CardDesignDiamond))
	players[2].AddCard(cardBT(7, CardDesignDiamond))
	players[3].AddCard(cardBT(8, CardDesignDiamond))
	players[3].AddCard(cardBT(9, CardDesignDiamond))

	bt.round.currentTurn = 0
	bt.round.lastPlayPlayerIdx = -1

	// Player 0 plays ♦3
	err := bt.PlayerPlay([]int{0})
	assert.NoError(t, err)
	assert.NotNil(t, bt.GetTableCards())
}

// --- Config ---

func TestBigTwoConfig_Validate(t *testing.T) {
	t.Run("valid default", func(t *testing.T) {
		assert.NoError(t, DefaultBigTwoConfig().Validate())
	})
	t.Run("invalid difficulty", func(t *testing.T) {
		cfg := DefaultBigTwoConfig()
		cfg.CpuDifficulty = 99
		assert.Error(t, cfg.Validate())
	})
}

func TestBigTwoConfig_JSON(t *testing.T) {
	cfg := DefaultBigTwoConfig()
	cfg.CpuDifficulty = BigTwoDifficultyHard
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	var restored BigTwoConfig
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, cfg, restored)
}

// --- Player ---

func TestBigTwoPlayer_SortCardsByBigTwoStrength(t *testing.T) {
	p := NewBigTwoPlayer(true)
	p.AddCard(cardBT(2, CardDesignSpade))   // strongest
	p.AddCard(cardBT(3, CardDesignDiamond)) // weakest
	p.AddCard(cardBT(1, CardDesignHeart))   // Ace
	p.SortCardsByBigTwoStrength()

	assert.Equal(t, 3, p.GetCard(0).GetValue()) // ♦3 weakest
	assert.Equal(t, 1, p.GetCard(1).GetValue()) // A
	assert.Equal(t, 2, p.GetCard(2).GetValue()) // ♠2 strongest
}

func TestBigTwoPlayer_JSON(t *testing.T) {
	p := NewBigTwoPlayer(true)
	p.AddCard(cardBT(5, CardDesignSpade))
	p.SetRank(1)
	data, err := json.Marshal(p)
	require.NoError(t, err)
	var restored BigTwoPlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.True(t, restored.GetIsHuman())
	assert.Equal(t, 1, restored.GetRank())
	assert.Equal(t, 1, restored.GetCardsSize())
}

func TestBigTwoPlayer_JSON_NilRankedPlayer(t *testing.T) {
	data := []byte(`{"rp":null}`)
	var restored BigTwoPlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.False(t, restored.GetIsHuman())
}

// --- BigTwo JSON ---

func TestBigTwo_JSON(t *testing.T) {
	bt := NewDefaultBigTwo()
	bt.Reset()

	data, err := json.Marshal(bt)
	require.NoError(t, err)

	var restored BigTwo
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, bt.GetCurrentTurn(), restored.GetCurrentTurn())
	assert.Equal(t, bt.GetGameEndFlag(), restored.GetGameEndFlag())
	assert.Equal(t, bt.GetPlayerCnt(), restored.GetPlayerCnt())
}

func TestBigTwo_JSON_NilFields(t *testing.T) {
	data := []byte(`{"tc":null,"pl":null,"cf":{},"ct":0,"tb":null,"tt":0,"lp":-1,"ge":false,"pc":0,"ca":null,"ha":null,"al":null}`)
	var restored BigTwo
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.NotNil(t, restored.trumpCards)
	assert.NotNil(t, restored.players)
	assert.NotNil(t, restored.round.actionLog)
}

func TestBigTwo_JSON_MaxSliceLen(t *testing.T) {
	huge := make([]*BigTwoPlayer, bigTwoMaxSliceLen+1)
	for i := range huge {
		huge[i] = NewBigTwoPlayer(false)
	}
	bt := &BigTwo{
		trumpCards: NewTrumpCards(0),
		players:    huge,
	}
	data, err := json.Marshal(bt)
	require.NoError(t, err)

	var restored BigTwo
	err = json.Unmarshal(data, &restored)
	assert.Error(t, err)
}

// --- BigTwoAction JSON ---

func TestBigTwoAction_JSON(t *testing.T) {
	a := &BigTwoAction{
		PlayerIdx:   2,
		PlayedCards: []*Card{cardBT(5, CardDesignSpade)},
	}
	data, err := json.Marshal(a)
	require.NoError(t, err)
	var restored BigTwoAction
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, 2, restored.PlayerIdx)
	assert.Len(t, restored.PlayedCards, 1)
}

// --- Getters ---

func TestBigTwo_Getters(t *testing.T) {
	bt := newTestBigTwo()
	bt.players[0].AddCard(cardBT(3, CardDesignDiamond))
	bt.players[1].AddCard(cardBT(4, CardDesignSpade))
	bt.players[2].AddCard(cardBT(5, CardDesignSpade))
	bt.players[3].AddCard(cardBT(6, CardDesignSpade))

	assert.Equal(t, BigTwoPlayerCnt, bt.GetPlayerCnt())
	assert.False(t, bt.GetGameEndFlag())
	assert.Nil(t, bt.GetTableCards())
	assert.Equal(t, BigTwoPlayInvalid, bt.GetTablePlayType())
	assert.Equal(t, -1, bt.GetLastPlayPlayerIdx())
	assert.Equal(t, 0, bt.GetPassCount())
	assert.Nil(t, bt.GetCpuActions())
	assert.Nil(t, bt.GetHumanAction())
	assert.False(t, bt.HasPendingAction())
	assert.NotNil(t, bt.GetConfig())
	assert.NotNil(t, bt.GetActionLog())
	assert.Nil(t, bt.GetPlayer(-1))
	assert.Nil(t, bt.GetPlayer(99))
}

// --- SetConfig ---

func TestBigTwo_SetConfig(t *testing.T) {
	bt := NewDefaultBigTwo()
	cfg := BigTwoConfig{CpuDifficulty: BigTwoDifficultyHard}
	bt.SetConfig(cfg)
	assert.Equal(t, BigTwoDifficultyHard, bt.GetConfig().CpuDifficulty)
}

// --- CPU difficulty variants ---

func TestBigTwo_CpuDifficulty(t *testing.T) {
	for _, diff := range []BigTwoCpuDifficulty{BigTwoDifficultyEasy, BigTwoDifficultyNormal, BigTwoDifficultyHard} {
		t.Run("difficulty_"+string(rune('0'+diff)), func(t *testing.T) {
			bt := newTestBigTwo()
			bt.config.CpuDifficulty = diff
			players := bt.players
			players[0].AddCard(cardBT(3, CardDesignDiamond))
			players[1].AddCard(cardBT(4, CardDesignSpade))
			players[1].AddCard(cardBT(5, CardDesignSpade))
			players[2].AddCard(cardBT(6, CardDesignSpade))
			players[3].AddCard(cardBT(7, CardDesignSpade))

			bt.round.tableCards = []*Card{cardBT(3, CardDesignDiamond)}
			bt.round.tablePlayType = BigTwoPlaySingle
			bt.round.lastPlayPlayerIdx = 0
			bt.round.currentTurn = 1

			bt.CpuPlay()
			assert.NotEmpty(t, bt.GetCpuActions())
		})
	}
}

// --- 5-card strength comparison ---

func TestBigTwoFlushStrength(t *testing.T) {
	spadeFlush := []*Card{
		cardBT(3, CardDesignSpade), cardBT(5, CardDesignSpade), cardBT(7, CardDesignSpade),
		cardBT(9, CardDesignSpade), cardBT(11, CardDesignSpade),
	}
	heartFlush := []*Card{
		cardBT(3, CardDesignHeart), cardBT(5, CardDesignHeart), cardBT(7, CardDesignHeart),
		cardBT(9, CardDesignHeart), cardBT(11, CardDesignHeart),
	}
	// Spade flush > Heart flush (suit ordering)
	assert.Greater(t, bigTwoFlushStrength(spadeFlush), bigTwoFlushStrength(heartFlush))
}

func TestBigTwoFullHouseStrength(t *testing.T) {
	fh1 := []*Card{
		cardBT(5, CardDesignSpade), cardBT(5, CardDesignHeart), cardBT(5, CardDesignClover),
		cardBT(3, CardDesignDiamond), cardBT(3, CardDesignSpade),
	}
	fh2 := []*Card{
		cardBT(6, CardDesignSpade), cardBT(6, CardDesignHeart), cardBT(6, CardDesignClover),
		cardBT(3, CardDesignDiamond), cardBT(3, CardDesignHeart),
	}
	assert.Less(t, bigTwoFullHouseStrength(fh1), bigTwoFullHouseStrength(fh2))
}

func TestBigTwoFourOfAKindStrength(t *testing.T) {
	foak1 := []*Card{
		cardBT(5, CardDesignSpade), cardBT(5, CardDesignHeart), cardBT(5, CardDesignClover),
		cardBT(5, CardDesignDiamond), cardBT(3, CardDesignSpade),
	}
	foak2 := []*Card{
		cardBT(6, CardDesignSpade), cardBT(6, CardDesignHeart), cardBT(6, CardDesignClover),
		cardBT(6, CardDesignDiamond), cardBT(3, CardDesignSpade),
	}
	assert.Less(t, bigTwoFourOfAKindStrength(foak1), bigTwoFourOfAKindStrength(foak2))
}

func TestBigTwoStraightStrength(t *testing.T) {
	s1 := []*Card{
		cardBT(3, CardDesignSpade), cardBT(4, CardDesignHeart), cardBT(5, CardDesignClover),
		cardBT(6, CardDesignDiamond), cardBT(7, CardDesignSpade),
	}
	s2 := []*Card{
		cardBT(4, CardDesignSpade), cardBT(5, CardDesignHeart), cardBT(6, CardDesignClover),
		cardBT(7, CardDesignDiamond), cardBT(8, CardDesignSpade),
	}
	assert.Less(t, bigTwoStraightStrength(s1), bigTwoStraightStrength(s2))
}

func TestBigTwoPlayStrength_Invalid(t *testing.T) {
	assert.Equal(t, -1, bigTwoPlayStrength(nil, BigTwoPlayInvalid))
}

// --- bigTwoCheckStraight edge cases ---

func TestBigTwoCheckStraight_AceLow(t *testing.T) {
	// A-2-3-4-5 is NOT a valid straight in Big Two (2 cannot be in straights)
	assert.False(t, bigTwoCheckStraight([]int{1, 2, 3, 4, 5}))
}

func TestBigTwoCheckStraight_AceHigh(t *testing.T) {
	// 10-J-Q-K-A is valid
	assert.True(t, bigTwoCheckStraight([]int{1, 10, 11, 12, 13}))
}

func TestBigTwoCheckStraight_AceMidInvalid(t *testing.T) {
	// A in middle position is invalid
	assert.False(t, bigTwoCheckStraight([]int{1, 3, 4, 5, 6}))
}

// --- CPU findAllPlayableSets edge cases ---

func TestBigTwo_FindAllPlayableSets_EmptyHand(t *testing.T) {
	bt := newTestBigTwo()
	// No cards
	results := bt.findAllPlayableSets(bt.players[0])
	assert.Empty(t, results)
}

func TestBigTwo_FindFiveCardCombos_NotEnoughCards(t *testing.T) {
	bt := newTestBigTwo()
	bt.players[0].AddCard(cardBT(3, CardDesignDiamond))
	bt.players[0].AddCard(cardBT(4, CardDesignSpade))
	results := bt.findFiveCardCombos(bt.players[0], false)
	assert.Empty(t, results)
}

// --- Full game simulation ---

func TestBigTwo_FullGame(t *testing.T) {
	bt := NewDefaultBigTwo()
	bt.Reset()

	for i := 0; i < 1000 && !bt.GetGameEndFlag(); i++ {
		if bt.IsHumanTurn() {
			player := bt.GetPlayer(bt.GetCurrentTurn())
			if bt.GetTableCards() == nil {
				// Must play; try each card
				played := false
				for j := 0; j < player.GetCardsSize(); j++ {
					err := bt.PlayerPlay([]int{j})
					if err == nil {
						played = true
						break
					}
				}
				if !played {
					// Try pairs
					for j := 0; j < player.GetCardsSize() && !played; j++ {
						for k := j + 1; k < player.GetCardsSize() && !played; k++ {
							err := bt.PlayerPlay([]int{j, k})
							if err == nil {
								played = true
							}
						}
					}
				}
				require.True(t, played, "human player must be able to play on empty table")
			} else {
				// Try to play, or pass
				played := false
				for j := 0; j < player.GetCardsSize(); j++ {
					err := bt.PlayerPlay([]int{j})
					if err == nil {
						played = true
						break
					}
				}
				if !played {
					err := bt.PlayerPlay([]int{})
					require.NoError(t, err)
				}
			}
		} else {
			bt.CpuPlay()
		}
	}

	assert.True(t, bt.GetGameEndFlag(), "game should end within 1000 iterations")
}
