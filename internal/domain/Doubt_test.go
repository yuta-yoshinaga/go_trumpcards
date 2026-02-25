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
		err := game.PlayerPlay([]int{0}, 1)
		assert.NoError(t, err)
		assert.True(t, game.GetGameEndFlag())
		// Try again
		err = game.PlayerPlay([]int{0}, 1)
		assert.ErrorIs(t, err, domain.ErrGameEnded)
	})

	t.Run("wrong phase", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		game.SetPhase(domain.DoubtPhaseDoubt)
		err := game.PlayerPlay([]int{0}, 1)
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
		err := game3.PlayerPlay([]int{0}, 1)
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})

	t.Run("invalid claimed value too low", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		err := game.PlayerPlay([]int{0}, 0)
		assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	})

	t.Run("invalid claimed value too high", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		err := game.PlayerPlay([]int{0}, 14)
		assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	})

	t.Run("empty card indices", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		err := game.PlayerPlay([]int{}, 1)
		assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	})

	t.Run("out of range card index", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		err := game.PlayerPlay([]int{99}, 1)
		assert.ErrorIs(t, err, domain.ErrInvalidCard)
	})
}

func TestDoubt_PlayerPlay_Success(t *testing.T) {
	t.Run("normal play - phase changes to Doubt", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))

		err := game.PlayerPlay([]int{0}, 5)
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

		err := game.PlayerPlay([]int{0, 1}, 3)
		assert.NoError(t, err)
		assert.Equal(t, 2, game.GetTableCardCount())
		assert.Equal(t, 2, game.GetLastAction().CardCount)
		assert.Equal(t, 1, game.GetPlayer(0).GetCardsSize())
	})

	t.Run("game ends when last card played", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))

		err := game.PlayerPlay([]int{0}, 1)
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

func TestDoubt_CpuPlay(t *testing.T) {
	t.Run("no-op when game ended", func(t *testing.T) {
		_, p2 := makeDoubtGame()
		g2 := domain.NewDoubt(domain.NewTrumpCards(0), p2)
		p2[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		_ = g2.PlayerPlay([]int{0}, 1) // ends game
		before := g2.GetTableCardCount()
		g2.CpuPlay() // no-op
		assert.Equal(t, before, g2.GetTableCardCount())
	})

	t.Run("no-op when human turn", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		// currentTurn=0 (human)
		before := players[0].GetCardsSize()
		game.CpuPlay()
		assert.Equal(t, before, players[0].GetCardsSize())
	})

	t.Run("no-op when not in play phase", func(t *testing.T) {
		game, players := makeDoubtGame()
		advanceToCpuTurn(game)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		game.SetPhase(domain.DoubtPhaseDoubt) // not PhasePlay
		before := players[1].GetCardsSize()
		game.CpuPlay()
		assert.Equal(t, before, players[1].GetCardsSize())
	})

	t.Run("CPU plays and game ends when last card played", func(t *testing.T) {
		game, players := makeDoubtGame()
		advanceToCpuTurn(game)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		game.CpuPlay()

		assert.True(t, game.GetGameEndFlag())
		assert.Equal(t, 1, game.GetWinnerIdx())
		assert.True(t, players[1].GetIsFinished())
	})

	// CPU plays without ending game (deterministic: give many cards so hand can't be emptied in 1 play
	// with numCards = rand.Intn(N)+1 where N=13; probability of emptying is 1/13).
	// Loop guarantees both the "game doesn't end" branch AND the human-player-skip in decideCpuDoubters.
	t.Run("CPU plays without ending game - game continues to DoubtPhase", func(t *testing.T) {
		for attempt := 0; attempt < 1000; attempt++ {
			game, players := makeDoubtGame()
			advanceToCpuTurn(game)
			// Give CPU 1 exactly 13 cards; only 1/13 chance of emptying hand
			for i := 1; i <= 13; i++ {
				players[1].AddCard(domain.NewCard(domain.CardDesignSpade, i, false))
			}

			beforeTable := game.GetTableCardCount()
			game.CpuPlay()

			if !game.GetGameEndFlag() {
				// "Game doesn't end" branch hit:
				// - phase changed to DoubtPhaseDoubt
				// - decideCpuDoubters ran (checking human player at i=0)
				assert.Equal(t, domain.DoubtPhaseDoubt, game.GetPhase())
				assert.Greater(t, game.GetTableCardCount(), beforeTable)
				assert.NotNil(t, game.GetLastAction())
				assert.Equal(t, 1, game.GetLastAction().PlayerIdx)
				assert.NotEmpty(t, game.GetCpuActions())
				return // success
			}
		}
		t.Fatal("CPU never played without ending game after 1000 attempts")
	})

	// Cover the intentBluff=false (honest play) branch of CpuPlay.
	// IsBluff is now based on actual card matching: false when claimed value equals all played cards.
	// This happens when intentBluff=false AND numCards=1 (single card always matches its own value).
	t.Run("CPU plays honestly at least once", func(t *testing.T) {
		for attempt := 0; attempt < 1000; attempt++ {
			game, players := makeDoubtGame()
			advanceToCpuTurn(game)
			for i := 1; i <= 13; i++ {
				players[1].AddCard(domain.NewCard(domain.CardDesignSpade, i, false))
			}

			game.CpuPlay()

			cpuActions := game.GetCpuActions()
			if len(cpuActions) > 0 && !cpuActions[0].IsBluff {
				// actualIsBluff=false branch was hit (all played cards match claimed value)
				return // success
			}
		}
		t.Fatal("CPU never played honestly after 1000 attempts")
	})
}

func TestDoubt_DecideCpuDoubters(t *testing.T) {
	t.Run("after human plays, cpuDoubters only contains valid CPU indices", func(t *testing.T) {
		_, players := makeDoubtGame()
		g := domain.NewDoubt(domain.NewTrumpCards(0), players)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		_ = g.PlayerPlay([]int{0}, 1)
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
