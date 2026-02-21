package entities_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"

	"github.com/stretchr/testify/assert"
)

func makeDaifugoPlayers() []*entities.DaifugoPlayer {
	return []*entities.DaifugoPlayer{
		entities.NewDaifugoPlayer(true),
		entities.NewDaifugoPlayer(false),
		entities.NewDaifugoPlayer(false),
		entities.NewDaifugoPlayer(false),
	}
}

func TestDaifugo_Method(t *testing.T) {
	t.Run("success NewDaifugo", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		assert.NotNil(t, dg)
		assert.Equal(t, 4, dg.GetPlayerCnt())
		assert.False(t, dg.GetGameEndFlag())
		assert.Nil(t, dg.GetTableCards())
		assert.Equal(t, -1, dg.GetLastPlayPlayerIdx())
		assert.Equal(t, 0, dg.GetCurrentTurn())
	})

	t.Run("success Reset distributes 52 cards", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		dg.Reset()
		total := 0
		for i := 0; i < dg.GetPlayerCnt(); i++ {
			total += dg.GetPlayer(i).GetCardsSize()
		}
		assert.Equal(t, 52, total)
	})

	t.Run("success Reset clears state", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		dg.Reset()
		assert.False(t, dg.GetGameEndFlag())
		assert.Nil(t, dg.GetTableCards())
		assert.Equal(t, -1, dg.GetLastPlayPlayerIdx())
		assert.Equal(t, 0, dg.GetPassCount())
		assert.Nil(t, dg.GetHumanAction())
		assert.Nil(t, dg.GetCpuActions())
	})

	t.Run("success GetPlayer valid index", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		assert.NotNil(t, dg.GetPlayer(0))
		assert.True(t, dg.GetPlayer(0).GetIsHuman())
		assert.NotNil(t, dg.GetPlayer(1))
		assert.False(t, dg.GetPlayer(1).GetIsHuman())
	})

	t.Run("success GetPlayer invalid index returns nil", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		assert.Nil(t, dg.GetPlayer(-1))
		assert.Nil(t, dg.GetPlayer(10))
	})

	t.Run("success IsHumanTurn at start", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		assert.True(t, dg.IsHumanTurn())
	})

	t.Run("success PlayerPlay on clear table", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		// Human has 2 cards so they don't finish when playing one
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		ok := dg.PlayerPlay([]int{0}) // play 3
		assert.True(t, ok)
		assert.NotNil(t, dg.GetTableCards())
		assert.Equal(t, 3, dg.GetTableCards()[0].GetValue())
		assert.Equal(t, 0, dg.GetLastPlayPlayerIdx())
		assert.Equal(t, 1, players[0].GetCardsSize())
	})

	t.Run("success PlayerPlay fails with invalid index", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))

		ok := dg.PlayerPlay([]int{5}) // out of range
		assert.False(t, ok)
	})

	t.Run("success PlayerPlay fails with different values", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))

		ok := dg.PlayerPlay([]int{0, 1}) // different values → invalid
		assert.False(t, ok)
	})

	t.Run("success PlayerPlay table card stays after valid play", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		ok := dg.PlayerPlay([]int{0}) // play 7
		assert.True(t, ok)
		assert.Equal(t, 7, dg.GetTableCards()[0].GetValue())
	})

	t.Run("success PlayerPlay pass", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		ok := dg.PlayerPlay([]int{}) // pass
		assert.True(t, ok)
		assert.Equal(t, 1, dg.GetPassCount())
		assert.NotNil(t, dg.GetHumanAction())
		assert.Nil(t, dg.GetHumanAction().PlayedCards) // pass → nil
	})

	t.Run("success PlayerPlay does nothing when not human turn", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		dg.PlayerPlay([]int{0}) // advance to CPU turn
		if !dg.IsHumanTurn() && !dg.GetGameEndFlag() {
			ok := dg.PlayerPlay([]int{0})
			assert.False(t, ok)
		}
	})

	t.Run("success CpuPlay passes on table with unbeatable card", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		// Human has 2 cards: [2 (idx0), 3 (idx1)]  — play the 2 (strongest), keep 3
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 2, false)) // idx0 → play this
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false)) // idx1 → kept
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 3, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 4, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		dg.PlayerPlay([]int{0}) // play 2 → human keeps [3], not finished
		// CPUs all pass (can't beat 2) → table clears → back to human
		dg.CpuPlay() // CPU 1 passes
		dg.CpuPlay() // CPU 2 passes
		dg.CpuPlay() // CPU 3 passes → checkPassClear triggers, table clears
		assert.Nil(t, dg.GetTableCards())
		assert.True(t, dg.IsHumanTurn())
	})

	t.Run("success CpuPlay does nothing on human turn", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		dg.CpuPlay() // does nothing on human turn
		assert.Nil(t, dg.GetTableCards())
		assert.True(t, dg.IsHumanTurn())
	})

	t.Run("success game ends when only 1 player remains", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		// 3 CPUs already finished, human has 1 card left
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		dg.PlayerPlay([]int{0}) // human plays last card → finishes → game ends
		assert.True(t, dg.GetGameEndFlag())
		// countFinished was 3 before human finished → rank = 4
		assert.Equal(t, 4, players[0].GetRank())
	})

	t.Run("success GetHumanAction after play", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		dg.PlayerPlay([]int{0}) // play 5
		action := dg.GetHumanAction()
		assert.NotNil(t, action)
		assert.Equal(t, 0, action.PlayerIdx)
		assert.Len(t, action.PlayedCards, 1)
		assert.Equal(t, 5, action.PlayedCards[0].GetValue())
	})

	t.Run("success GetCpuActions is nil at start", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		assert.Nil(t, dg.GetCpuActions())
	})

	t.Run("success pair play on clear table keeps table alive", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		// Human has 3 cards: pair of 5s + extra 3 (human doesn't finish after playing pair)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))  // idx0
		players[0].AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))  // idx1
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))  // idx2 (kept)
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		ok := dg.PlayerPlay([]int{0, 1}) // play pair of 5s
		assert.True(t, ok)
		assert.Len(t, dg.GetTableCards(), 2)
		assert.Equal(t, 1, players[0].GetCardsSize()) // 1 card (3) remains
	})

	t.Run("success PlayerPlay deduplicates indices so only unique cards are played", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		// Human has 3 cards. [0,0] must be treated as [0] — only 1 card goes to the table.
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false)) // idx0
		players[0].AddCard(entities.NewCard(entities.CardDesignHeart, 5, false)) // idx1
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false)) // idx2
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		ok := dg.PlayerPlay([]int{0, 0}) // duplicate → deduped to [0]
		assert.True(t, ok)
		assert.Len(t, dg.GetTableCards(), 1)
		assert.Equal(t, 5, dg.GetTableCards()[0].GetValue())
		assert.Equal(t, 2, players[0].GetCardsSize()) // 2 cards remain
	})

	t.Run("success PlayerPlay rejects fake pair from duplicate indices when table has a pair", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		// Human plays pair of 3s; CPU 1 responds with stronger pair of 5s;
		// CPUs 2 & 3 have singles so they cannot beat a pair → they pass.
		// Turn returns to human with pair of 5s on the table.
		// CPU 1 gets a spare card (9) so it does not finish when it plays the pair of 5s
		// (finishing would clear the table immediately, breaking the test scenario).
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))   // idx0 – played first
		players[0].AddCard(entities.NewCard(entities.CardDesignHeart, 3, false))   // idx1 – played first
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))   // idx2 – remains
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))   // pair of 5s (beats 3s pair)
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))   // spare – keeps CPU 1 alive
		players[2].AddCard(entities.NewCard(entities.CardDesignClover, 6, false))  // single – cannot beat pair
		players[3].AddCard(entities.NewCard(entities.CardDesignDiamond, 6, false))

		dg.PlayerPlay([]int{0, 1}) // human plays pair of 3s → table=[3,3], turn→CPU 1
		dg.CpuPlay()               // CPU 1 plays pair of 5s → table=[5,5], turn→CPU 2
		dg.CpuPlay()               // CPU 2 passes (single cannot beat pair)
		dg.CpuPlay()               // CPU 3 passes; currentTurn→0, lastPlay=1 → no clear
		assert.True(t, dg.IsHumanTurn())
		assert.Len(t, dg.GetTableCards(), 2) // pair of 5s still on table

		// Human has [7♦] at idx0; tries [0,0] as a fake pair → deduped to [0] (1 card)
		// isPlayable([7]) fails: 1 card ≠ 2 needed
		ok := dg.PlayerPlay([]int{0, 0})
		assert.False(t, ok)                         // correctly rejected
		assert.Len(t, dg.GetTableCards(), 2)         // table unchanged
		assert.Equal(t, 1, players[0].GetCardsSize()) // hand unchanged
	})

	t.Run("success finishPlayer rank based on already-finished count", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		// 2 already finished → human finishes → gets rank 3
		dg.PlayerPlay([]int{0})
		assert.Equal(t, 3, players[0].GetRank())
	})

	// --- Revolution rule tests ---

	t.Run("success GetRevolutionActive is false initially", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		assert.False(t, dg.GetRevolutionActive())
	})

	t.Run("success playing 4 cards triggers revolution", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		// Human has 4 fives + extra card (does not finish), CPUs have unbeatable 2s
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignDiamond, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false)) // extra card keeps human alive
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		ok := dg.PlayerPlay([]int{0, 1, 2, 3}) // play four 5s
		assert.True(t, ok)
		assert.True(t, dg.GetRevolutionActive())
	})

	t.Run("success isPlayable respects revolution (3 beats 2 during revolution)", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		// Set up: table has a 2 (strongest normally), revolution is active
		// Human plays four 5s to trigger revolution, then on clear table tries to play 3 over 2
		// Instead: manually set up revolution by playing 4 cards first
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignDiamond, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false)) // will play this next
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 2, false)) // 2 also in hand
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false)) // CPU passes (can't beat 4 of kind)
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		// Play four 5s → revolution, table = four 5s, advance to CPU1
		dg.PlayerPlay([]int{0, 1, 2, 3})
		assert.True(t, dg.GetRevolutionActive())
		// CPUs have a 2 (revolution-weakest), but the table has 4 cards, CPUs only have 1 → they pass
		dg.CpuPlay() // CPU1 passes
		dg.CpuPlay() // CPU2 passes
		dg.CpuPlay() // CPU3 passes → table clears
		assert.Nil(t, dg.GetTableCards())
		// Now on clear table, human plays single 2. Then we verify 3 can beat 2 during revolution.
		// After table clear, human plays 2 (revolution-weakest single)
		// Human hand after revolution+re-sort should be sorted by revolution strength (weakest=2 first): [2, 3]
		assert.True(t, dg.IsHumanTurn())
		assert.Equal(t, 2, players[0].GetCard(0).GetValue()) // 2 is now at index 0 (weakest in revolution)
		// Play the 2 on clear table
		ok := dg.PlayerPlay([]int{0})
		assert.True(t, ok)
		assert.Equal(t, 2, dg.GetTableCards()[0].GetValue())
		// CPUs pass again since they only have singles and can't match (or table has 2 which is weakest)
		// Actually CPUs have a single 2 → revolution strength 3, table has 2 (rev strength 3) → can't beat → pass
		dg.CpuPlay()
		dg.CpuPlay()
		dg.CpuPlay()
		// Back to human, table has 2 on it. Human has 3 (rev-strongest, rev-strength=15 > 3)
		// Human should be able to play 3 over 2 during revolution
		assert.True(t, dg.IsHumanTurn())
		assert.Equal(t, 3, players[0].GetCard(0).GetValue()) // only 3 left
		ok2 := dg.PlayerPlay([]int{0})
		assert.True(t, ok2) // 3 beats 2 during revolution (verified by successful play)
		// player 0 emptied their hand → finishPlayer clears the table
		assert.Nil(t, dg.GetTableCards())
	})

	t.Run("success double revolution reverts to normal", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		// Human plays four 5s → revolution active
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignDiamond, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignClover, 7, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignDiamond, 7, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false)) // spare
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		dg.PlayerPlay([]int{0, 1, 2, 3}) // play four 5s → revolution active
		assert.True(t, dg.GetRevolutionActive())
		// CPUs pass (can't match 4 cards)
		dg.CpuPlay()
		dg.CpuPlay()
		dg.CpuPlay()
		assert.Nil(t, dg.GetTableCards())
		// Human plays four 7s on clear table → revolution cancelled
		// After first revolution: hand sorted by rev strength (weakest first = 2,A,K,...,3)
		// Human has [2(rev-weak), 3(rev-strong), 7,7,7,7] — wait: CPUs have 2s, human's hand had 5,5,5,5 (played), 7,7,7,7, 3
		// After revolution, human's remaining cards: [7,7,7,7,3] sorted by rev strength (weakest first)
		// Rev strengths: 7→11, 3→15. So sorted: [7,7,7,7,3] where 7 (rev-str=11) comes before 3 (rev-str=15)
		// Indices 0-3 are 7s
		assert.True(t, dg.IsHumanTurn())
		ok := dg.PlayerPlay([]int{0, 1, 2, 3}) // play four 7s → revolution cancelled
		assert.True(t, ok)
		assert.False(t, dg.GetRevolutionActive())
	})

	t.Run("success revolution resets on Reset", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		// Trigger revolution
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignDiamond, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		dg.PlayerPlay([]int{0, 1, 2, 3})
		assert.True(t, dg.GetRevolutionActive())
		// Reset clears revolution
		dg.Reset()
		assert.False(t, dg.GetRevolutionActive())
	})

	t.Run("success findBestPlay during revolution picks weakest by revolution strength", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		// Trigger revolution: human plays four 5s
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignDiamond, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false)) // spare
		// CPU1 has 2 and K: after revolution, 2 is weakest (rev-str=3), K is stronger (rev-str=5)
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 13, false)) // K
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		// Play four 5s → revolution, turn advances to CPU1
		dg.PlayerPlay([]int{0, 1, 2, 3})
		assert.True(t, dg.GetRevolutionActive())
		// CPUs pass (table has 4 cards, CPUs only have 1-2 cards → can't match)
		dg.CpuPlay() // CPU1: table=4 cards, CPU1 has 2 cards → can't match → pass
		dg.CpuPlay() // CPU2 passes
		dg.CpuPlay() // CPU3 passes → table clears
		assert.Nil(t, dg.GetTableCards())
		// Human's turn on clear table — human has [3] (spare), plays 3 (clear table = anything)
		// Actually after revolution, human has [3] which in rev order is strongest (rev-str=15)
		// but there's only one card. Let's pass human and let CPU1 play on clear table.
		dg.PlayerPlay([]int{}) // human passes
		// CPU1 has [2, K] sorted by revolution strength: 2(rev-str=3), K(rev-str=5)
		// CPU1 should play the weakest by rev strength = 2 (index 0)
		dg.CpuPlay() // CPU1 plays on clear table (plays weakest = 2 in revolution)
		// Table should have the 2 (CPU1's weakest in revolution)
		assert.NotNil(t, dg.GetTableCards())
		assert.Equal(t, 2, dg.GetTableCards()[0].GetValue())
	})

	t.Run("success CPU triggers revolution with 4-card play", func(t *testing.T) {
		tc := entities.NewTrumpCards(0)
		players := makeDaifugoPlayers()
		dg := entities.NewDaifugo(tc, players)
		// Human plays four 5s → revolution active (rev-strength(5)=13)
		// CPU1 has four 4s: rev-strength(4)=14 > rev-strength(5)=13 → can beat four 5s in revolution
		// CPU1 playing four 4s triggers a second revolution, reverting to normal order.
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignDiamond, 5, false))
		players[0].AddCard(entities.NewCard(entities.CardDesignSpade, 3, false)) // spare keeps human alive
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 4, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignClover, 4, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignDiamond, 4, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignSpade, 4, false))
		players[1].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false)) // spare keeps CPU1 alive
		players[2].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))
		players[3].AddCard(entities.NewCard(entities.CardDesignHeart, 2, false))

		dg.PlayerPlay([]int{0, 1, 2, 3}) // human plays four 5s → revolution active
		assert.True(t, dg.GetRevolutionActive())

		// CPU1's hand after revolution: sorted by rev-strength ascending → [2(rev=3), 4(rev=14), 4, 4, 4]
		// findBestPlay: table has 4 cards (needed=4), tableStrength=rev(5)=13
		// Group of 4s at indices 1-4, count=4 >= 4, rev(4)=14 > 13 → plays four 4s
		dg.CpuPlay() // CPU1 plays four 4s → double revolution (back to normal)
		assert.False(t, dg.GetRevolutionActive())
	})
}
