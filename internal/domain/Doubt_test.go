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

	t.Run("duplicate card index", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		err := game.PlayerPlay([]int{0, 0}, 1)
		assert.ErrorIs(t, err, domain.ErrInvalidCard)
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

func TestDoubt_CpuPlay_MixedBluff(t *testing.T) {
	t.Run("mixed bluff branch hit", func(t *testing.T) {
		for attempt := 0; attempt < 1000; attempt++ {
			game, players := makeDoubtGame()
			advanceToCpuTurn(game)
			// Give CPU 1 five cards with different values to enable mixed bluff
			for i := 1; i <= 5; i++ {
				players[1].AddCard(domain.NewCard(domain.CardDesignSpade, i, false))
			}

			game.CpuPlay()

			cpuActions := game.GetCpuActions()
			lastAction := game.GetLastAction()
			if len(cpuActions) == 0 || lastAction == nil {
				continue
			}
			action := cpuActions[0]
			if !action.IsBluff || action.CardCount < 2 {
				continue
			}
			// Check if at least one played card matches claimed value AND
			// at least one doesn't (signature of mixed bluff)
			hasMatch := false
			hasNonMatch := false
			for _, card := range lastAction.PlayedCards {
				if card.GetValue() == action.ClaimedValue {
					hasMatch = true
				} else {
					hasNonMatch = true
				}
			}
			if hasMatch && hasNonMatch {
				return // mixed bluff detected
			}
		}
		t.Fatal("mixed bluff branch never hit after 1000 attempts")
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

func TestDoubt_GetSetConfig(t *testing.T) {
	game, _ := makeDoubtGame()

	// Default config
	cfg := game.GetConfig()
	assert.Equal(t, 10, cfg.DoubtWindowSec)
	assert.Equal(t, domain.DoubtMemoryLevelNormal, cfg.CpuMemoryLevel)

	// Custom config
	custom := domain.DoubtConfig{DoubtWindowSec: 3, CpuMemoryLevel: domain.DoubtMemoryLevelHard}
	game.SetConfig(custom)
	got := game.GetConfig()
	assert.Equal(t, 3, got.DoubtWindowSec)
	assert.Equal(t, domain.DoubtMemoryLevelHard, got.CpuMemoryLevel)
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
	err := game.PlayerPlay([]int{0}, 5)
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

	_ = game.PlayerPlay([]int{0}, 1)

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

func TestDoubt_MemoryRetentionChanceLevels(t *testing.T) {
	// Easy level: sometimes skips recording
	t.Run("easy level - partial retention", func(t *testing.T) {
		game, players := makeDoubtGame()
		game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: domain.DoubtMemoryLevelEasy})

		// Try many resolves to see if some are NOT recorded (retention = 0.3)
		skipped := false
		for attempt := 0; attempt < 1000; attempt++ {
			p := domain.NewDoubtPlayer(false)
			players[2] = p // replace CPU 2 with fresh player
			game2, players2 := makeDoubtGame()
			game2.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: domain.DoubtMemoryLevelEasy})
			playedCard := domain.NewCard(domain.CardDesignSpade, 11, false)
			game2.SetLastAction(&domain.DoubtAction{
				PlayerIdx: 0, ClaimedValue: 5, CardCount: 1,
				PlayedCards: []*domain.Card{playedCard},
			})
			game2.SetPhase(domain.DoubtPhaseDoubt)
			game2.SetTableCards([]*domain.Card{playedCard})
			game2.ResolveDoubt([]int{1})
			if players2[2].CountKnownCards(11) == 0 {
				skipped = true
				break
			}
		}
		assert.True(t, skipped, "easy level should sometimes skip recording")
	})

	// Normal level: sometimes skips recording
	t.Run("normal level - partial retention", func(t *testing.T) {
		skipped := false
		for attempt := 0; attempt < 1000; attempt++ {
			game2, players2 := makeDoubtGame()
			game2.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: domain.DoubtMemoryLevelNormal})
			playedCard := domain.NewCard(domain.CardDesignSpade, 12, false)
			game2.SetLastAction(&domain.DoubtAction{
				PlayerIdx: 0, ClaimedValue: 5, CardCount: 1,
				PlayedCards: []*domain.Card{playedCard},
			})
			game2.SetPhase(domain.DoubtPhaseDoubt)
			game2.SetTableCards([]*domain.Card{playedCard})
			game2.ResolveDoubt([]int{1})
			if players2[2].CountKnownCards(12) == 0 {
				skipped = true
				break
			}
		}
		assert.True(t, skipped, "normal level should sometimes skip recording")
	})

	// Hard level: always records
	t.Run("hard level - full retention", func(t *testing.T) {
		game2, players2 := makeDoubtGame()
		game2.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: domain.DoubtMemoryLevelHard})
		playedCard := domain.NewCard(domain.CardDesignSpade, 13, false)
		game2.SetLastAction(&domain.DoubtAction{
			PlayerIdx: 0, ClaimedValue: 5, CardCount: 1,
			PlayedCards: []*domain.Card{playedCard},
		})
		game2.SetPhase(domain.DoubtPhaseDoubt)
		game2.SetTableCards([]*domain.Card{playedCard})
		game2.ResolveDoubt([]int{1})
		// CPU 2 and CPU 3 should have recorded value 13
		assert.Equal(t, 1, players2[2].CountKnownCards(13))
		assert.Equal(t, 1, players2[3].CountKnownCards(13))
	})
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

func TestDoubt_MemoryDecay_EasyLevel(t *testing.T) {
	// Easy level: memories should decay over turns
	// Record a card at turn 0, then advance to a high turn and check if decay happens
	forgotten := false
	for attempt := 0; attempt < 1000; attempt++ {
		game, players := makeDoubtGame()
		game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: domain.DoubtMemoryLevelEasy})

		// Record card for CPU 1 at turn 0
		players[1].RecordRevealedCard(5, 1.0, 0)
		assert.Equal(t, 1, players[1].CountKnownCards(5))

		// Advance turnCounter to simulate many turns passed
		game.SetTurnCounter(10)

		// Trigger decideCpuDoubters which calls DecayMemories
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		_ = game.PlayerPlay([]int{0}, 1)

		if players[1].CountKnownCards(5) == 0 {
			forgotten = true
			break
		}
	}
	assert.True(t, forgotten, "easy level should eventually forget old memories")
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
	_ = game.PlayerPlay([]int{0}, 1)

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

func TestDoubt_HasTell(t *testing.T) {
	t.Run("HasTell is set on bluff actions", func(t *testing.T) {
		tellSeen := false
		for attempt := 0; attempt < 1000; attempt++ {
			game, players := makeDoubtGame()
			game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: domain.DoubtMemoryLevelEasy})
			advanceToCpuTurn(game)
			players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
			players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))

			game.CpuPlay()

			cpuActions := game.GetCpuActions()
			if len(cpuActions) > 0 && cpuActions[0].IsBluff && cpuActions[0].HasTell {
				tellSeen = true
				break
			}
		}
		assert.True(t, tellSeen, "HasTell should be set on at least one bluff action after 1000 attempts (Easy=40%% tell chance)")
	})

	t.Run("HasTell is false on non-bluff actions", func(t *testing.T) {
		for attempt := 0; attempt < 1000; attempt++ {
			game, players := makeDoubtGame()
			advanceToCpuTurn(game)
			players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
			players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))

			game.CpuPlay()

			cpuActions := game.GetCpuActions()
			if len(cpuActions) > 0 && !cpuActions[0].IsBluff {
				assert.False(t, cpuActions[0].HasTell, "HasTell should be false when not bluffing")
				return
			}
		}
		t.Fatal("could not find a non-bluff action after 1000 attempts")
	})

	t.Run("HasTell not set on bluff with retry (false branch)", func(t *testing.T) {
		noTellSeen := false
		for attempt := 0; attempt < 1000; attempt++ {
			game, players := makeDoubtGame()
			game.SetConfig(domain.DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: domain.DoubtMemoryLevelEasy})
			advanceToCpuTurn(game)
			players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
			players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))

			game.CpuPlay()

			cpuActions := game.GetCpuActions()
			if len(cpuActions) > 0 && cpuActions[0].IsBluff && !cpuActions[0].HasTell {
				noTellSeen = true
				break
			}
		}
		assert.True(t, noTellSeen, "HasTell should be false on at least one bluff action after 1000 attempts")
	})

	t.Run("human action never has HasTell", func(t *testing.T) {
		game, players := makeDoubtGame()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		_ = game.PlayerPlay([]int{0}, 5)
		ha := game.GetHumanAction()
		assert.NotNil(t, ha)
		assert.False(t, ha.HasTell)
	})
}
