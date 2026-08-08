package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// makeDoubtGame ダウトゲームを生成するヘルパー
func makeDoubtGame() (*domain.Doubt, []*domain.DoubtPlayer) {
	players := []*domain.DoubtPlayer{
		domain.NewDoubtPlayer(true),  // 0: human
		domain.NewDoubtPlayer(false), // 1: CPU
		domain.NewDoubtPlayer(false), // 2: CPU
		domain.NewDoubtPlayer(false), // 3: CPU
	}
	tc := domain.NewTrumpCards(0)
	game := domain.NewDoubt(tc, players)
	return game, players
}

func TestNewDoubt(t *testing.T) {
	game, _ := makeDoubtGame()
	assert.False(t, game.GetGameEndFlag())
	assert.Equal(t, domain.DoubtPhasePlay, game.GetPhase())
	assert.Equal(t, 0, game.GetCurrentTurn())
	assert.Equal(t, 4, game.GetPlayerCnt())
	assert.Equal(t, -1, game.GetWinnerIdx())
	assert.Equal(t, 0, game.GetTableCardCount())
	assert.Nil(t, game.GetLastAction())
	assert.Nil(t, game.GetCpuDoubters())
	assert.Nil(t, game.GetCpuActions())
	assert.Nil(t, game.GetHumanAction())
	assert.Nil(t, game.GetLastDoubtResult())
}

func TestDoubt_Reset(t *testing.T) {
	game, _ := makeDoubtGame()
	game.Reset()

	assert.False(t, game.GetGameEndFlag())
	assert.Equal(t, domain.DoubtPhasePlay, game.GetPhase())
	assert.Equal(t, 0, game.GetCurrentTurn())
	assert.Equal(t, -1, game.GetWinnerIdx())
	assert.Equal(t, 0, game.GetTableCardCount())
	assert.Nil(t, game.GetLastAction())
	assert.Nil(t, game.GetCpuDoubters())
	assert.Nil(t, game.GetCpuActions())
	assert.Nil(t, game.GetHumanAction())
	assert.Nil(t, game.GetLastDoubtResult())

	// 52枚が4人に均等に配られているか
	total := 0
	for i := 0; i < game.GetPlayerCnt(); i++ {
		total += game.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 52, total)
}

func TestDoubt_GetPlayer(t *testing.T) {
	game, _ := makeDoubtGame()

	t.Run("valid index", func(t *testing.T) {
		p := game.GetPlayer(0)
		assert.NotNil(t, p)
		assert.True(t, p.GetIsHuman())
	})

	t.Run("negative index returns nil", func(t *testing.T) {
		assert.Nil(t, game.GetPlayer(-1))
	})

	t.Run("out of range index returns nil", func(t *testing.T) {
		assert.Nil(t, game.GetPlayer(99))
	})
}

func TestDoubt_IsHumanTurn(t *testing.T) {
	game, _ := makeDoubtGame()
	// currentTurn=0 is human
	assert.True(t, game.IsHumanTurn())
}

func TestDoubt_PlayerPlay_Errors(t *testing.T) {
	t.Run("game ended", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		game.SetPhase(domain.DoubtPhasePlay)
		// End the game manually
		err := game.PlayerPlay([]int{0}, 1, 0)
		assert.NoError(t, err)
		assert.True(t, game.GetGameEndFlag())
		// Try again
		err = game.PlayerPlay([]int{0}, 1, 0)
		assert.ErrorIs(t, err, domain.ErrGameEnded)
	})

	t.Run("wrong phase", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		game.SetPhase(domain.DoubtPhaseDoubt)
		err := game.PlayerPlay([]int{0}, 1, 0)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("not human turn", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		// Advance turn manually by setting phase and calling SkipDoubt
		// Use a fresh game where currentTurn=1 (CPU)
		players2 := []*domain.DoubtPlayer{
			domain.NewDoubtPlayer(false), // 0: CPU
			domain.NewDoubtPlayer(true),  // 1: human
		}
		tc := domain.NewTrumpCards(0)
		// Create a 2-player game is not possible since DoubtPlayerCnt=4
		// Instead set up so currentTurn points to CPU
		game2, players2b := makeDoubtGame()
		players2b[1].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		_ = players2
		_ = game
		// Set lastAction and skip to advance to CPU turn
		game2.SetLastAction(&domain.DoubtAction{PlayerIdx: 3, ClaimedValue: 1, CardCount: 1, PlayedCards: []*domain.Card{}})
		game2.SetPhase(domain.DoubtPhaseDoubt)
		game2.SkipDoubt() // advances currentTurn to (3+1)%4 = 0, but 0 is human...
		// Let's directly test: set phase=Play but currentTurn=1 (CPU)
		// We need a way to set currentTurn. Let's add SetLastAction to manipulate.
		// Use a different approach: call PlayerPlay on a game where human isn't at turn 0
		_ = tc
		// Just test via a direct manipulation using the exported test setter
		// SetPhase makes it PhasePlay; we need currentTurn to be CPU
		// Re-read: PlayerPlay returns ErrNotHumanTurn if !players[currentTurn].GetIsHuman()
		// If currentTurn=1 (CPU), it returns ErrNotHumanTurn
		// We can advance the turn by: setLastAction(playerIdx=0), setPhase=Doubt, SkipDoubt -> currentTurn=1
		game3, players3 := makeDoubtGame()
		players3[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players3[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		game3.SetLastAction(&domain.DoubtAction{
			PlayerIdx: 0, ClaimedValue: 1, CardCount: 1,
			PlayedCards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)},
		})
		game3.SetPhase(domain.DoubtPhaseDoubt)
		game3.SkipDoubt() // currentTurn = (0+1)%4 = 1 (CPU)
		// Now currentTurn=1 (CPU), phase=Play
		players3[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		err := game3.PlayerPlay([]int{0}, 1, 0)
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})

	t.Run("invalid claimed value too low", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		err := game.PlayerPlay([]int{0}, 0, 0)
		assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	})

	t.Run("invalid claimed value too high", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		err := game.PlayerPlay([]int{0}, 14, 0)
		assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	})

	t.Run("empty card indices", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		err := game.PlayerPlay([]int{}, 1, 0)
		assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	})

	t.Run("duplicate card index", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		err := game.PlayerPlay([]int{0, 0}, 1, 0)
		assert.ErrorIs(t, err, domain.ErrInvalidCard)
	})

	t.Run("out of range card index", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		err := game.PlayerPlay([]int{99}, 1, 0)
		assert.ErrorIs(t, err, domain.ErrInvalidCard)
	})
}

func TestDoubt_PlayerPlay_Success(t *testing.T) {
	t.Run("normal play - phase changes to Doubt", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))

		err := game.PlayerPlay([]int{0}, 5, 0)
		assert.NoError(t, err)
		assert.Equal(t, domain.DoubtPhaseDoubt, game.GetPhase())
		assert.Equal(t, 1, game.GetTableCardCount())
		assert.NotNil(t, game.GetLastAction())
		assert.Equal(t, 5, game.GetLastAction().ClaimedValue)
		assert.Equal(t, 1, game.GetLastAction().CardCount)
		assert.Equal(t, 0, game.GetLastAction().PlayerIdx)
		assert.NotNil(t, game.GetHumanAction())
		assert.Equal(t, 5, game.GetHumanAction().ClaimedValue)
		assert.Nil(t, game.GetCpuActions())
	})

	t.Run("play multiple cards", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))

		err := game.PlayerPlay([]int{0, 1}, 3, 0)
		assert.NoError(t, err)
		assert.Equal(t, 2, game.GetTableCardCount())
		assert.Equal(t, 2, game.GetLastAction().CardCount)
		assert.Equal(t, 1, game.GetPlayer(0).GetCardsSize())
	})

	t.Run("game ends when last card played", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))

		err := game.PlayerPlay([]int{0}, 1, 0)
		assert.NoError(t, err)
		assert.True(t, game.GetGameEndFlag())
		assert.Equal(t, 0, game.GetWinnerIdx())
		assert.True(t, game.GetPlayer(0).GetIsFinished())
		// Phase should NOT be DoubtPhaseDoubt when game ends
		assert.NotEqual(t, domain.DoubtPhaseDoubt, game.GetPhase())
	})
}

// advanceToCpuTurn ヘルパー: player[0] が出した後 SkipDoubt で CPU 1 の手番にする
func advanceToCpuTurn(game *domain.Doubt) {
	game.SetLastAction(&domain.DoubtAction{
		PlayerIdx:    0,
		ClaimedValue: 1,
		CardCount:    1,
		PlayedCards:  []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)},
	})
	game.SetPhase(domain.DoubtPhaseDoubt)
	game.SkipDoubt() // currentTurn = 1 (CPU)
}

func TestDoubt_DecideCpuDoubters(t *testing.T) {
	t.Run("after human plays, cpuDoubters only contains valid CPU indices", func(t *testing.T) {
		_, players := makeDoubtGame()
		g := domain.NewDoubt(domain.NewTrumpCards(0), players)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		_ = g.PlayerPlay([]int{0}, 1, 0)
		// cpuDoubters should only contain CPU players (1, 2, 3), not card player (0)
		for _, idx := range g.GetCpuDoubters() {
			assert.NotEqual(t, 0, idx)
			assert.GreaterOrEqual(t, idx, 1)
			assert.LessOrEqual(t, idx, 3)
		}
	})
}

func TestDoubt_ResolveDoubt(t *testing.T) {
	t.Run("no-op when wrong phase", func(t *testing.T) {
		game, _ := makeDoubtGame()
		// phase is PhasePlay, not PhaseDoubt
		game.ResolveDoubt([]int{1})
		assert.Nil(t, game.GetLastDoubtResult())
	})

	t.Run("no-op when no last action", func(t *testing.T) {
		game, _ := makeDoubtGame()
		game.SetPhase(domain.DoubtPhaseDoubt)
		game.ResolveDoubt([]int{1})
		assert.Nil(t, game.GetLastDoubtResult())
	})

	t.Run("falls through to SkipDoubt when no valid doubter", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		// Set up: player 0 played, phase=Doubt
		game.SetLastAction(&domain.DoubtAction{PlayerIdx: 0, ClaimedValue: 1, CardCount: 1,
			PlayedCards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)}})
		game.SetPhase(domain.DoubtPhaseDoubt)
		tableCard := domain.NewCard(domain.CardDesignHeart, 1, false)
		game.SetTableCards([]*domain.Card{tableCard})

		// ResolveDoubt with empty list → SkipDoubt
		game.ResolveDoubt([]int{})
		// After SkipDoubt: phase=PhasePlay, currentTurn advanced
		assert.Equal(t, domain.DoubtPhasePlay, game.GetPhase())
		assert.Nil(t, game.GetLastDoubtResult())
		// Table cards remain (SkipDoubt doesn't clear them)
		assert.Equal(t, 1, game.GetTableCardCount())
	})

	t.Run("card player was lying - doubter wins, card player takes cards", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // value 5
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		// Player 0 played a SPADE 5 but claimed value 3 (lie)
		playedCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		tableCards := []*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 1, false),
			domain.NewCard(domain.CardDesignHeart, 2, false),
			playedCard,
		}
		game.SetLastAction(&domain.DoubtAction{
			PlayerIdx:    0,
			ClaimedValue: 3, // but card value is 5 → lie
			CardCount:    1,
			PlayedCards:  []*domain.Card{playedCard},
		})
		game.SetPhase(domain.DoubtPhaseDoubt)
		game.SetTableCards(tableCards)
		originalP0Cards := players[0].GetCardsSize()

		game.ResolveDoubt([]int{1})

		result := game.GetLastDoubtResult()
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.DoubterIdx)
		assert.Equal(t, 0, result.CardPlayerIdx)
		assert.True(t, result.WasLying)
		assert.Equal(t, 0, result.LoserIdx)
		// Player 0 should have gained the 3 table cards
		assert.Equal(t, originalP0Cards+3, players[0].GetCardsSize())
		// Table cleared
		assert.Equal(t, 0, game.GetTableCardCount())
		// Phase reset to Play
		assert.Equal(t, domain.DoubtPhasePlay, game.GetPhase())
		// Turn advanced to (0+1)%4 = 1
		assert.Equal(t, 1, game.GetCurrentTurn())
	})

	t.Run("card player was honest - doubter loses, doubter takes cards", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		// Player 0 played a card with correct value
		playedCard := domain.NewCard(domain.CardDesignSpade, 3, false)
		tableCards := []*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 1, false),
			playedCard,
		}
		game.SetLastAction(&domain.DoubtAction{
			PlayerIdx:    0,
			ClaimedValue: 3, // matches played card value → honest
			CardCount:    1,
			PlayedCards:  []*domain.Card{playedCard},
		})
		game.SetPhase(domain.DoubtPhaseDoubt)
		game.SetTableCards(tableCards)
		originalP1Cards := players[1].GetCardsSize()

		game.ResolveDoubt([]int{1})

		result := game.GetLastDoubtResult()
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.DoubterIdx)
		assert.Equal(t, 0, result.CardPlayerIdx)
		assert.False(t, result.WasLying)
		assert.Equal(t, 1, result.LoserIdx)
		// Player 1 should have gained the 2 table cards
		assert.Equal(t, originalP1Cards+2, players[1].GetCardsSize())
		assert.Equal(t, 0, game.GetTableCardCount())
	})

	t.Run("priority resolution - next player after card player wins", func(t *testing.T) {
		game, _ := makeDoubtGame()
		// Card player is 0, doubters are 2 and 1
		// Priority: 1 (next after 0), then 2
		playedCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		game.SetLastAction(&domain.DoubtAction{
			PlayerIdx:    0,
			ClaimedValue: 3, // lie
			CardCount:    1,
			PlayedCards:  []*domain.Card{playedCard},
		})
		game.SetPhase(domain.DoubtPhaseDoubt)
		game.SetTableCards([]*domain.Card{playedCard})

		game.ResolveDoubt([]int{2, 1}) // 1 has priority over 2

		result := game.GetLastDoubtResult()
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.DoubterIdx) // 1 was chosen
	})

	t.Run("no valid doubter in list - falls through to SkipDoubt", func(t *testing.T) {
		// doubterIndices non-empty but contains only invalid player indices (99)
		// This hits the return -1 at end of findHighestPriorityDoubter (not the early return)
		game, _ := makeDoubtGame()
		playedCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		game.SetLastAction(&domain.DoubtAction{
			PlayerIdx:    0,
			ClaimedValue: 3,
			CardCount:    1,
			PlayedCards:  []*domain.Card{playedCard},
		})
		game.SetPhase(domain.DoubtPhaseDoubt)
		game.SetTableCards([]*domain.Card{playedCard})

		// Index 99 doesn't correspond to any player → no match → return -1 at end → SkipDoubt
		game.ResolveDoubt([]int{99})

		assert.Equal(t, domain.DoubtPhasePlay, game.GetPhase())
		assert.Nil(t, game.GetLastDoubtResult())
		// Table cards remain (SkipDoubt doesn't clear)
		assert.Equal(t, 1, game.GetTableCardCount())
	})
}

func TestDoubt_SkipDoubt(t *testing.T) {
	t.Run("no-op when wrong phase", func(t *testing.T) {
		game, _ := makeDoubtGame()
		// phase is PhasePlay
		game.SkipDoubt()
		assert.Equal(t, domain.DoubtPhasePlay, game.GetPhase())
		assert.Equal(t, 0, game.GetCurrentTurn())
	})

	t.Run("no-op when no last action", func(t *testing.T) {
		game, _ := makeDoubtGame()
		game.SetPhase(domain.DoubtPhaseDoubt)
		game.SkipDoubt()
		// phase remains Doubt since lastAction=nil
		assert.Equal(t, domain.DoubtPhaseDoubt, game.GetPhase())
	})

	t.Run("success - advances turn and clears lastDoubtResult", func(t *testing.T) {
		game, _ := makeDoubtGame()
		playedCard := domain.NewCard(domain.CardDesignSpade, 1, false)
		tableCard := domain.NewCard(domain.CardDesignHeart, 5, false)
		game.SetLastAction(&domain.DoubtAction{PlayerIdx: 2, ClaimedValue: 1, CardCount: 1,
			PlayedCards: []*domain.Card{playedCard}})
		game.SetPhase(domain.DoubtPhaseDoubt)
		game.SetTableCards([]*domain.Card{tableCard})
		game.SetLastDoubtResult(&domain.DoubtDoubtResult{DoubterIdx: 1})

		game.SkipDoubt()

		assert.Equal(t, domain.DoubtPhasePlay, game.GetPhase())
		assert.Equal(t, 3, game.GetCurrentTurn()) // (2+1)%4
		assert.Nil(t, game.GetLastDoubtResult())
		// Table cards remain
		assert.Equal(t, 1, game.GetTableCardCount())
	})
}

func TestDoubt_GetTableCards(t *testing.T) {
	game, _ := makeDoubtGame()
	assert.Nil(t, game.GetTableCards())

	cards := []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	game.SetTableCards(cards)
	assert.Equal(t, cards, game.GetTableCards())
}

func TestDoubt_GetSetConfig(t *testing.T) {
	game, _ := makeDoubtGame()

	// Default config
	cfg := game.GetConfig()
	assert.Equal(t, 10, cfg.DoubtWindowSec)
	assert.Equal(t, domain.DoubtMemoryLevelNormal, cfg.CpuMemoryLevel)
	assert.Equal(t, 0, cfg.PenaltyDrawLimit) // default = unlimited

	// Custom config
	custom := domain.DoubtConfig{DoubtWindowSec: 3, CpuMemoryLevel: domain.DoubtMemoryLevelHard, PenaltyDrawLimit: 5}
	game.SetConfig(custom)
	got := game.GetConfig()
	assert.Equal(t, 3, got.DoubtWindowSec)
	assert.Equal(t, domain.DoubtMemoryLevelHard, got.CpuMemoryLevel)
	assert.Equal(t, 5, got.PenaltyDrawLimit)
}

func TestDoubt_Reset_ClearsMemory(t *testing.T) {
	game, players := makeDoubtGame()
	// Use value 99 (impossible in deck) so hand cards after reset won't interfere
	players[1].RecordRevealedCard(99, 1.0, 0)
	assert.Equal(t, 1, players[1].CountKnownCards(99))

	game.Reset()
	// After reset, memory is cleared and no dealt cards have value 99
	assert.Equal(t, 0, players[1].CountKnownCards(99))
}

func TestDoubt_DecideCpuDoubters_ImpossibleClaim(t *testing.T) {
	// CPU player 1 has 4 cards of value 5 in memory → impossible claim for claimedValue=5, count=1
	game, players := makeDoubtGame()
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))

	// Give CPU 1 knowledge of 4 cards of value 5 (max in deck)
	for i := 0; i < 4; i++ {
		players[1].RecordRevealedCard(5, 1.0, 0)
	}

	// Human plays value=5 (claimed), count=1 → known(5)+1 = 5 > 4 → impossible
	err := game.PlayerPlay([]int{0}, 5, 0)
	assert.NoError(t, err)

	// CPU 1 should ALWAYS doubt (impossible claim)
	found1 := false
	for _, idx := range game.GetCpuDoubters() {
		if idx == 1 {
			found1 = true
		}
	}
	assert.True(t, found1, "CPU 1 should always doubt an impossible claim")
}

func TestDoubt_DecideCpuDoubters_SkipsFinishedPlayers(t *testing.T) {
	game, players := makeDoubtGame()
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	players[1].SetIsFinished(true) // CPU 1 finished

	_ = game.PlayerPlay([]int{0}, 1, 0)

	// CPU 1 should not be in doubters (finished)
	for _, idx := range game.GetCpuDoubters() {
		assert.NotEqual(t, 1, idx, "finished CPU should not be in doubters")
	}
}

func TestDoubt_ResolveDoubt_UpdatesMemory(t *testing.T) {
	// Use Hard memory so retention is 100%
	game, players := makeDoubtGame()
	game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: domain.DoubtMemoryLevelHard})

	// Player 0 (human) played cards with value 7 (lie: claimed 3)
	playedCard := domain.NewCard(domain.CardDesignSpade, 7, false)
	tableCards := []*domain.Card{playedCard}
	game.SetLastAction(&domain.DoubtAction{
		PlayerIdx:    0,
		ClaimedValue: 3,
		CardCount:    1,
		PlayedCards:  []*domain.Card{playedCard},
	})
	game.SetPhase(domain.DoubtPhaseDoubt)
	game.SetTableCards(tableCards)

	// CPU 1 doubts → player 0 was lying → loser = 0
	game.ResolveDoubt([]int{1})

	// CPU 2 and CPU 3 (non-loser, non-human) should have recorded the revealed card (value=7)
	// CPU 1 (doubter, non-loser) should also have recorded
	result := game.GetLastDoubtResult()
	assert.NotNil(t, result)
	assert.True(t, result.WasLying)
	assert.Equal(t, 0, result.LoserIdx)

	// CPU 1 is doubter and non-loser → should record memory
	assert.Equal(t, 1, players[1].CountKnownCards(7))
	// CPU 2, CPU 3 also non-loser → should record
	assert.Equal(t, 1, players[2].CountKnownCards(7))
	assert.Equal(t, 1, players[3].CountKnownCards(7))
	// Human (player 0) is the loser, memory not updated (human doesn't track anyway)
}

func TestDoubt_ResolveDoubt_LoserDoesNotRecord(t *testing.T) {
	// CPU 1 doubts but player 0 was honest → loser = CPU 1
	// CPU 1 (loser) should NOT record memory
	game, players := makeDoubtGame()
	game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: domain.DoubtMemoryLevelHard})

	playedCard := domain.NewCard(domain.CardDesignHeart, 4, false)
	tableCards := []*domain.Card{playedCard}
	game.SetLastAction(&domain.DoubtAction{
		PlayerIdx:    0,
		ClaimedValue: 4, // honest
		CardCount:    1,
		PlayedCards:  []*domain.Card{playedCard},
	})
	game.SetPhase(domain.DoubtPhaseDoubt)
	game.SetTableCards(tableCards)

	// Give player 1 a card so it can receive table cards
	players[1].AddCard(domain.NewCard(domain.CardDesignClover, 2, false))

	game.ResolveDoubt([]int{1})

	result := game.GetLastDoubtResult()
	assert.NotNil(t, result)
	assert.False(t, result.WasLying)
	assert.Equal(t, 1, result.LoserIdx) // CPU 1 loses

	// CPU 1 (loser): received the table card (value 4) in hand via resolve,
	// memory NOT updated → CountKnownCards(4) = 0 (memory) + 1 (hand card) = 1
	assert.Equal(t, 1, players[1].CountKnownCards(4))
	// CPU 2, CPU 3 (non-loser) recorded value 4 in memory, none in hand
	assert.Equal(t, 1, players[2].CountKnownCards(4))
	assert.Equal(t, 1, players[3].CountKnownCards(4))
}

func TestDoubt_SettersForTest(t *testing.T) {
	game, _ := makeDoubtGame()

	cpuAction := &domain.DoubtCpuAction{PlayerIdx: 1, ClaimedValue: 3, CardCount: 1}
	game.SetCpuActions([]*domain.DoubtCpuAction{cpuAction})
	assert.Len(t, game.GetCpuActions(), 1)

	humanAction := &domain.DoubtCpuAction{PlayerIdx: 0, ClaimedValue: 5, CardCount: 2}
	game.SetHumanAction(humanAction)
	assert.Equal(t, humanAction, game.GetHumanAction())

	game.SetCpuDoubters([]int{1, 2})
	assert.Equal(t, []int{1, 2}, game.GetCpuDoubters())

	result := &domain.DoubtDoubtResult{DoubterIdx: 1, WasLying: true}
	game.SetLastDoubtResult(result)
	assert.Equal(t, result, game.GetLastDoubtResult())

	game.SetWinnerIdx(2)
	assert.Equal(t, 2, game.GetWinnerIdx())
}

func TestDoubt_TurnCounter(t *testing.T) {
	t.Run("initial value is 0", func(t *testing.T) {
		game, _ := makeDoubtGame()
		assert.Equal(t, 0, game.GetTurnCounter())
	})

	t.Run("SkipDoubt increments turnCounter", func(t *testing.T) {
		game, _ := makeDoubtGame()
		game.SetLastAction(&domain.DoubtAction{
			PlayerIdx: 0, ClaimedValue: 1, CardCount: 1,
			PlayedCards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)},
		})
		game.SetPhase(domain.DoubtPhaseDoubt)
		game.SkipDoubt()
		assert.Equal(t, 1, game.GetTurnCounter())
	})

	t.Run("ResolveDoubt increments turnCounter", func(t *testing.T) {
		game, players := makeDoubtGame()
		playedCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		game.SetLastAction(&domain.DoubtAction{
			PlayerIdx: 0, ClaimedValue: 3, CardCount: 1,
			PlayedCards: []*domain.Card{playedCard},
		})
		game.SetPhase(domain.DoubtPhaseDoubt)
		game.SetTableCards([]*domain.Card{playedCard})
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))

		game.ResolveDoubt([]int{1})
		assert.Equal(t, 1, game.GetTurnCounter())
	})

	t.Run("Reset clears turnCounter", func(t *testing.T) {
		game, _ := makeDoubtGame()
		game.SetTurnCounter(10)
		game.Reset()
		assert.Equal(t, 0, game.GetTurnCounter())
	})

	t.Run("SetTurnCounter sets value", func(t *testing.T) {
		game, _ := makeDoubtGame()
		game.SetTurnCounter(42)
		assert.Equal(t, 42, game.GetTurnCounter())
	})
}

func TestDoubt_MemoryDecay_HardLevel(t *testing.T) {
	// Hard level: memories never decay (decayRate=0)
	game, players := makeDoubtGame()
	game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: domain.DoubtMemoryLevelHard})

	// Record card for CPU 1 at turn 0
	players[1].RecordRevealedCard(5, 1.0, 0)

	// Advance turnCounter very far
	game.SetTurnCounter(100)

	// Trigger decideCpuDoubters which calls DecayMemories
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	_ = game.PlayerPlay([]int{0}, 1, 0)

	// Hard level: memory should never be lost
	assert.Equal(t, 1, players[1].CountKnownCards(5))
}

func TestDoubt_DynamicBluffChance(t *testing.T) {
	// When CPU has only 1 card, bluff rate should be lower (0.1 vs 0.4)
	// We test by running 1000 trials and checking bluff rate is significantly lower
	t.Run("last card reduces bluff rate", func(t *testing.T) {
		bluffCount := 0
		trials := 1000
		for i := 0; i < trials; i++ {
			game, players := makeDoubtGame()
			advanceToCpuTurn(game)
			// Give CPU 1 exactly 1 card (handSize=1 after play will be 0)
			players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

			game.CpuPlay()

			cpuActions := game.GetCpuActions()
			if len(cpuActions) > 0 && cpuActions[0].IsBluff {
				bluffCount++
			}
		}
		// With bluffChance=0.1, expected bluff rate ~10%
		// With old bluffChance=0.4, expected bluff rate ~40%
		// Use threshold of 25% to distinguish
		bluffRate := float64(bluffCount) / float64(trials)
		assert.Less(t, bluffRate, 0.25, "bluff rate with 1 card should be much lower than 40%%")
	})
}

func TestDoubt_ResolveDoubt_PenaltyDrawLimit(t *testing.T) {
	testCases := []struct {
		name              string
		limit             int
		tableCardCount    int
		expectedTakeCount int
		expectedDiscard   int
	}{
		{"limit=0 (unlimited), table=5", 0, 5, 5, 0},
		{"limit=3, table=5", 3, 5, 3, 2},
		{"limit=5, table=3", 5, 3, 3, 0},
		{"limit=5, table=5", 5, 5, 5, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			game, players := makeDoubtGame()
			game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: domain.DoubtMemoryLevelHard, PenaltyDrawLimit: tc.limit})

			tableCards := make([]*domain.Card, tc.tableCardCount)
			for i := 0; i < tc.tableCardCount; i++ {
				tableCards[i] = domain.NewCard(domain.CardDesignHeart, i+1, false)
			}
			playedCard := tableCards[len(tableCards)-1]

			game.SetLastAction(&domain.DoubtAction{
				PlayerIdx: 0, ClaimedValue: 13, CardCount: 1,
				PlayedCards: []*domain.Card{playedCard},
			})
			game.SetPhase(domain.DoubtPhaseDoubt)
			game.SetTableCards(tableCards)

			game.ResolveDoubt([]int{1})

			result := game.GetLastDoubtResult()
			assert.NotNil(t, result)
			assert.Equal(t, tc.expectedTakeCount, result.CardCount, "CardCount")
			assert.Equal(t, tc.expectedDiscard, result.DiscardedCount, "DiscardedCount")
			assert.Equal(t, tc.expectedTakeCount, players[0].GetCardsSize(), "Loser hand size")
		})
	}
}

func TestDoubt_MetaAI_ProfileSurvivesReset(t *testing.T) {
	game, _ := makeDoubtGame()
	game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMetaAI: true})

	// First Reset creates a new profile
	game.Reset()
	profile := game.GetHumanProfile()
	assert.NotNil(t, profile, "profile should be created on first Reset with CpuMetaAI=true")
	assert.Equal(t, 0, profile.GamesPlayed, "GamesPlayed should be 0 on first Reset")

	// Record some data on the profile
	profile.RecordPlay(5, true)
	profile.RecordDoubt(true)
	assert.Equal(t, 1, profile.BluffsByBracket[1].Total)
	assert.Equal(t, 1, profile.DoubtTotal)

	// Second Reset should preserve profile data and increment GamesPlayed
	game.Reset()
	profile2 := game.GetHumanProfile()
	assert.NotNil(t, profile2, "profile should survive Reset")
	assert.Equal(t, 1, profile2.GamesPlayed, "GamesPlayed should be incremented on second Reset")
	assert.Equal(t, 1, profile2.BluffsByBracket[1].Total, "recorded data should survive Reset")
	assert.Equal(t, 1, profile2.DoubtTotal, "doubt data should survive Reset")
}

func TestDoubt_MetaAI_ProfileNotCreatedWhenDisabled(t *testing.T) {
	game, _ := makeDoubtGame()
	// CpuMetaAI defaults to false
	game.Reset()
	assert.Nil(t, game.GetHumanProfile(), "profile should be nil when CpuMetaAI is disabled")
}

func TestDoubt_MetaAI_ResetProfileClearsProfile(t *testing.T) {
	game, _ := makeDoubtGame()
	game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMetaAI: true})
	game.Reset()
	assert.NotNil(t, game.GetHumanProfile(), "profile should exist after Reset with CpuMetaAI=true")

	game.ResetProfile()
	assert.Nil(t, game.GetHumanProfile(), "profile should be nil after ResetProfile")
}

func TestDoubt_MetaAI_PlayerPlayRecordsBluff(t *testing.T) {
	t.Run("bluff is recorded when card value does not match claimed value", func(t *testing.T) {
		game, players := makeDoubtGame()
		game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMetaAI: true})
		game.SetHumanProfile(&domain.DoubtHumanProfile{})

		// Give human cards: spade 5 and spade 6
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		// Play card index 0 (value=5) but claim value=3 → bluff
		err := game.PlayerPlay([]int{0}, 3, 0)
		assert.NoError(t, err)

		profile := game.GetHumanProfile()
		assert.NotNil(t, profile)
		// Hand size after play is 1 → bracket 0 (small: 1-4)
		assert.Equal(t, 1, profile.BluffsByBracket[0].Bluffs, "bluff should be recorded")
		assert.Equal(t, 1, profile.BluffsByBracket[0].Total, "total should be recorded")
	})

	t.Run("honest play is recorded when card value matches claimed value", func(t *testing.T) {
		game, players := makeDoubtGame()
		game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMetaAI: true})
		game.SetHumanProfile(&domain.DoubtHumanProfile{})

		// Give human cards: spade 5 and spade 6
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		// Play card index 0 (value=5) and claim value=5 → honest
		err := game.PlayerPlay([]int{0}, 5, 0)
		assert.NoError(t, err)

		profile := game.GetHumanProfile()
		assert.NotNil(t, profile)
		assert.Equal(t, 0, profile.BluffsByBracket[0].Bluffs, "bluff count should be 0 for honest play")
		assert.Equal(t, 1, profile.BluffsByBracket[0].Total, "total should be recorded")
	})
}

func TestDoubt_MetaAI_ResolveDoubtRecordsHumanDoubtAccuracy(t *testing.T) {
	t.Run("human doubts a lying CPU - wasCorrect=true recorded", func(t *testing.T) {
		game, players := makeDoubtGame()
		game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMetaAI: true})
		game.SetHumanProfile(&domain.DoubtHumanProfile{})

		// CPU (player 1) played card value=5 but claimed value=3 → lying
		playedCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		game.SetLastAction(&domain.DoubtAction{
			PlayerIdx:    1,
			ClaimedValue: 3,
			CardCount:    1,
			PlayedCards:  []*domain.Card{playedCard},
		})
		game.SetPhase(domain.DoubtPhaseDoubt)
		game.SetTableCards([]*domain.Card{playedCard})
		// Give CPU 1 a card so it can receive table cards if it loses
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))

		// Human (player 0) doubts
		game.ResolveDoubt([]int{0})

		profile := game.GetHumanProfile()
		assert.NotNil(t, profile)
		assert.Equal(t, 1, profile.DoubtCorrect, "correct doubt should be recorded")
		assert.Equal(t, 1, profile.DoubtTotal, "total doubt should be recorded")
	})

	t.Run("human doubts an honest CPU - wasCorrect=false recorded", func(t *testing.T) {
		game, players := makeDoubtGame()
		game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMetaAI: true})
		game.SetHumanProfile(&domain.DoubtHumanProfile{})

		// CPU (player 1) played card value=3 and claimed value=3 → honest
		playedCard := domain.NewCard(domain.CardDesignSpade, 3, false)
		game.SetLastAction(&domain.DoubtAction{
			PlayerIdx:    1,
			ClaimedValue: 3,
			CardCount:    1,
			PlayedCards:  []*domain.Card{playedCard},
		})
		game.SetPhase(domain.DoubtPhaseDoubt)
		game.SetTableCards([]*domain.Card{playedCard})
		// Human (player 0) will be the loser, needs cards to receive table
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))

		// Human (player 0) doubts
		game.ResolveDoubt([]int{0})

		profile := game.GetHumanProfile()
		assert.NotNil(t, profile)
		assert.Equal(t, 0, profile.DoubtCorrect, "incorrect doubt should not increment DoubtCorrect")
		assert.Equal(t, 1, profile.DoubtTotal, "total doubt should be recorded")
	})
}

func TestDoubt_MetaAI_CpuUsesAdjustedDoubtChance(t *testing.T) {
	// When meta-AI is enabled with a high bluff rate profile, CPU should doubt more
	// than baseline. Compare doubt frequency with and without meta-AI.
	t.Run("CPU doubts more when human has high bluff rate", func(t *testing.T) {
		trials := 5000

		// Count doubts WITHOUT meta-AI
		baselineDoubts := 0
		for i := 0; i < trials; i++ {
			game, players := makeDoubtGame()
			game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMetaAI: false})
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
			_ = game.PlayerPlay([]int{0}, 1, 0)
			baselineDoubts += len(game.GetCpuDoubters())
		}

		// Count doubts WITH meta-AI and high bluff rate profile
		metaDoubts := 0
		for i := 0; i < trials; i++ {
			game, players := makeDoubtGame()
			game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMetaAI: true})
			// High bluff rate in all brackets: 90% bluff rate, max adapt
			profile := &domain.DoubtHumanProfile{GamesPlayed: 5}
			profile.BluffsByBracket[0] = struct{ Bluffs, Total int }{9, 10}
			profile.BluffsByBracket[1] = struct{ Bluffs, Total int }{9, 10}
			profile.BluffsByBracket[2] = struct{ Bluffs, Total int }{9, 10}
			game.SetHumanProfile(profile)
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
			_ = game.PlayerPlay([]int{0}, 1, 0)
			metaDoubts += len(game.GetCpuDoubters())
		}

		// Meta-AI with high bluff rate should produce more doubts than baseline
		assert.Greater(t, metaDoubts, baselineDoubts,
			"meta-AI with high bluff rate should cause CPU to doubt more (meta=%d, baseline=%d)", metaDoubts, baselineDoubts)
	})
}

// doubtUniformHandValue is the rank every card in the CPU's hand gets in the
// meta-AI bluff test. Any rank works; it only has to be a single one.
const doubtUniformHandValue = 7

// setUniformDoubtHand replaces a player's whole hand with n cards of one rank.
//
// This is what makes the meta-AI bluff test measure what it claims to. The meta
// profile scales `intentBluff` (Doubt.go: AdjustedBluffChance), but the recorded
// `IsBluff` is `isActuallyBluff` -- "did the declared value differ from the cards
// played" -- and those two come apart on the honest path: an honest CPU playing
// several cards declares only played[0]'s rank while dumping the rest of the
// front of its hand, so with a mixed hand it is recorded as bluffing even though
// it never intended to. With 26 cards of mixed ranks that path swallowed ~96% of
// plays, both arms measured ~94.9%, and the strict `assert.Less` came down to
// binomial noise -- it tied at 18980 vs 18980 in CI (issue #5177).
//
// A single-rank hand collapses the difference: every honest play (plain or the
// mixed-bluff branch, which falls back to the first n cards when it cannot find a
// non-matching one) declares that rank and plays only that rank, so it is not a
// bluff; an intentional bluff declares a random rank and mismatches 12/13 of the
// time. `IsBluff` then tracks intent, which is the thing under test.
func setUniformDoubtHand(p *domain.DoubtPlayer, value, n int) {
	if size := p.GetCardsSize(); size > 0 {
		indices := make([]int, size)
		for i := range indices {
			indices[i] = i
		}
		p.RemoveCards(indices)
	}
	for i := 0; i < n; i++ {
		p.AddCard(domain.NewCard(domain.CardDesignSpade, value, false))
	}
}

func TestDoubt_MetaAI_CpuUsesAdjustedBluffChance(t *testing.T) {
	// When human has high doubt accuracy, CPU should bluff less
	t.Run("CPU bluffs less when human has high doubt accuracy", func(t *testing.T) {
		trials := 20000
		// Large enough that most plays leave >1 card behind, so calcBluffChance
		// uses the normal base rather than its last-card special case.
		const handSize = 26

		// Count bluffs WITHOUT meta-AI
		baselineBluffs := 0
		for i := 0; i < trials; i++ {
			game, players := makeDoubtGame()
			game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMetaAI: false})
			advanceToCpuTurn(game)
			setUniformDoubtHand(players[1], doubtUniformHandValue, handSize)
			game.CpuPlay()
			cpuActions := game.GetCpuActions()
			if len(cpuActions) > 0 && cpuActions[0].IsBluff {
				baselineBluffs++
			}
		}

		// Count bluffs WITH meta-AI and high doubt accuracy profile
		metaBluffs := 0
		for i := 0; i < trials; i++ {
			game, players := makeDoubtGame()
			game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMetaAI: true})
			// High doubt accuracy: 95%, max adapt (AdaptStrength caps at 0.2, so
			// the adjustment is base x (1 - 0.95*0.2) = base x 0.81).
			profile := &domain.DoubtHumanProfile{GamesPlayed: 5, DoubtCorrect: 19, DoubtTotal: 20}
			game.SetHumanProfile(profile)
			advanceToCpuTurn(game)
			setUniformDoubtHand(players[1], doubtUniformHandValue, handSize)
			game.CpuPlay()
			cpuActions := game.GetCpuActions()
			if len(cpuActions) > 0 && cpuActions[0].IsBluff {
				metaBluffs++
			}
		}

		// Both arms must actually bluff sometimes, otherwise "meta < baseline"
		// could be satisfied by a code path that stopped bluffing entirely.
		assert.Positive(t, metaBluffs, "meta arm recorded no bluffs at all -- the measurement is broken, not the AI")
		assert.Positive(t, baselineBluffs, "baseline arm recorded no bluffs at all -- the measurement is broken, not the AI")

		// Meta-AI with high doubt accuracy should produce fewer bluffs than baseline.
		// The expected gap is ~19% of the baseline (AdjustedBluffChance x 0.81) against
		// a binomial sigma of ~90 at these counts, so this is a many-sigma margin
		// rather than the coin-flip it used to be.
		assert.Less(t, metaBluffs, baselineBluffs,
			"meta-AI with high doubt accuracy should cause CPU to bluff less (meta=%d, baseline=%d)", metaBluffs, baselineBluffs)
	})
}

func TestDoubt_MetaAI_PlayerPlayRecordsHesitation(t *testing.T) {
	t.Run("hesitation is recorded when humanPlayMs > 0", func(t *testing.T) {
		game, players := makeDoubtGame()
		game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMetaAI: true})
		game.SetHumanProfile(&domain.DoubtHumanProfile{})

		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		err := game.PlayerPlay([]int{0}, 5, 1500)
		assert.NoError(t, err)

		profile := game.GetHumanProfile()
		assert.Equal(t, 1, profile.HesitationCount)
		assert.InDelta(t, 1500.0, profile.HesitationMean, 0.001)
	})

	t.Run("hesitation is not recorded when humanPlayMs is 0", func(t *testing.T) {
		game, players := makeDoubtGame()
		game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMetaAI: true})
		game.SetHumanProfile(&domain.DoubtHumanProfile{})

		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		err := game.PlayerPlay([]int{0}, 5, 0)
		assert.NoError(t, err)

		profile := game.GetHumanProfile()
		assert.Equal(t, 0, profile.HesitationCount)
	})

	t.Run("hesitation not recorded when metaAI is disabled", func(t *testing.T) {
		game, players := makeDoubtGame()
		game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMetaAI: false})

		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		_ = game.PlayerPlay([]int{0}, 5, 2000)
		assert.Nil(t, game.GetHumanProfile())
	})
}

// ---------------------------------------------------------------------------
// ActionLog tests
// ---------------------------------------------------------------------------

func TestDoubt_ActionLog_PlayerPlay(t *testing.T) {
	game, players := makeDoubtGame()
	// Give human 2 cards so game doesn't end
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	game.SetPhase(domain.DoubtPhasePlay)

	err := game.PlayerPlay([]int{0}, 1, 0)
	assert.NoError(t, err)

	log := game.GetActionLog()
	assert.GreaterOrEqual(t, len(log), 1)
	entry := log[0]
	assert.Equal(t, 0, entry.PlayerIdx)
	assert.Equal(t, "play", entry.ActionType)
	assert.Contains(t, entry.Detail, "declared 1")
	assert.Contains(t, entry.Detail, "1 card(s)")
	assert.Len(t, entry.Cards, 1)
}

func TestDoubt_ActionLog_CpuPlay(t *testing.T) {
	game, players := makeDoubtGame()
	advanceToCpuTurn(game)
	// Give CPU 1 enough cards
	for v := 1; v <= 13; v++ {
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
	}
	game.CpuPlay()

	log := game.GetActionLog()
	// Should contain play entries from CPU actions (excluding the nodoubt from advanceToCpuTurn)
	found := false
	for _, e := range log {
		if e.ActionType == "play" && e.PlayerIdx != 0 {
			found = true
			assert.NotEmpty(t, e.Detail)
			assert.NotEmpty(t, e.Cards)
			break
		}
	}
	assert.True(t, found, "expected CPU play action log entry")
}

func TestDoubt_ActionLog_ResolveDoubt(t *testing.T) {
	game, players := makeDoubtGame()
	// Human plays a card that is a bluff (claims 5 but card is 1)
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	game.SetPhase(domain.DoubtPhasePlay)
	err := game.PlayerPlay([]int{0}, 5, 0) // bluff: card is 1, claims 5
	assert.NoError(t, err)

	// Set phase to doubt and resolve
	game.SetPhase(domain.DoubtPhaseDoubt)
	game.ResolveDoubt([]int{1})

	log := game.GetActionLog()
	var doubtEntry, penaltyEntry *domain.ActionLogEntry
	for _, e := range log {
		if e.ActionType == "doubt" {
			doubtEntry = e
		}
		if e.ActionType == "penalty" {
			penaltyEntry = e
		}
	}
	assert.NotNil(t, doubtEntry, "expected doubt action log entry")
	assert.Equal(t, 1, doubtEntry.PlayerIdx)
	assert.Contains(t, doubtEntry.Detail, "lying")
	assert.NotNil(t, penaltyEntry, "expected penalty action log entry")
	assert.Contains(t, penaltyEntry.Detail, "card(s)")
}

func TestDoubt_ActionLog_SkipDoubt(t *testing.T) {
	game, players := makeDoubtGame()
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	game.SetPhase(domain.DoubtPhasePlay)
	err := game.PlayerPlay([]int{0}, 1, 0)
	assert.NoError(t, err)

	game.SetPhase(domain.DoubtPhaseDoubt)
	game.SkipDoubt()

	log := game.GetActionLog()
	found := false
	for _, e := range log {
		if e.ActionType == "nodoubt" {
			found = true
			assert.Equal(t, -1, e.PlayerIdx)
			assert.Equal(t, "no one doubted", e.Detail)
			break
		}
	}
	assert.True(t, found, "expected nodoubt action log entry")
}

func TestDoubt_ActionLog_Finish(t *testing.T) {
	game, players := makeDoubtGame()
	// Give human exactly 1 card so game ends on play
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	game.SetPhase(domain.DoubtPhasePlay)
	err := game.PlayerPlay([]int{0}, 1, 0)
	assert.NoError(t, err)
	assert.True(t, game.GetGameEndFlag())

	log := game.GetActionLog()
	found := false
	for _, e := range log {
		if e.ActionType == "finish" {
			found = true
			assert.Equal(t, -1, e.PlayerIdx)
			assert.Contains(t, e.Detail, "wins")
			break
		}
	}
	assert.True(t, found, "expected finish action log entry")
}

func TestDoubt_ActionLog_Reset(t *testing.T) {
	game, players := makeDoubtGame()
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	game.SetPhase(domain.DoubtPhasePlay)
	_ = game.PlayerPlay([]int{0}, 1, 0)
	assert.NotEmpty(t, game.GetActionLog())

	game.Reset()
	assert.Nil(t, game.GetActionLog())
}

// **Web の honestValue と同じ式であること。**両方が (値 % 13) + 1、開始直後は A。
func TestDoubtHonestClaimValue(t *testing.T) {
	assert.Equal(t, 1, domain.DoubtHonestClaimValue(nil))
	assert.Equal(t, 2, domain.DoubtHonestClaimValue(&domain.DoubtAction{ClaimedValue: 1}))
	assert.Equal(t, 8, domain.DoubtHonestClaimValue(&domain.DoubtAction{ClaimedValue: 7}))
	assert.Equal(t, 1, domain.DoubtHonestClaimValue(&domain.DoubtAction{ClaimedValue: 13}))
}
