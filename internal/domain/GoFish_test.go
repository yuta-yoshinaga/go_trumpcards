//go:build test

package domain

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestGoFish creates a GoFish game with empty hands for deterministic testing.
func newTestGoFish() *GoFish {
	players := []*GoFishPlayer{
		NewGoFishPlayer(true),  // P0: human
		NewGoFishPlayer(false), // P1: CPU
		NewGoFishPlayer(false), // P2: CPU
		NewGoFishPlayer(false), // P3: CPU
	}
	tc := NewTrumpCards(0)
	// Draw all cards so deck is empty; we'll set up manually
	for tc.DrawCard() != nil {
	}
	g := NewGoFish(tc, players)
	g.phase = GoFishPhasePlay
	g.currentTurn = 0
	g.turnNumber = 1
	g.config = DefaultGoFishConfig()
	return g
}

func TestGoFish_Reset(t *testing.T) {
	players := []*GoFishPlayer{
		NewGoFishPlayer(true),
		NewGoFishPlayer(false),
		NewGoFishPlayer(false),
		NewGoFishPlayer(false),
	}
	g := NewGoFish(NewTrumpCards(0), players)
	g.Reset()

	// 52 cards: 5 * 4 = 20 dealt, 32 remaining
	totalCards := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		totalCards += g.GetPlayer(i).GetCardsSize()
		// Account for any initial books formed (rare but possible)
		totalCards += g.GetPlayer(i).GetBookCount() * GoFishBookSize
	}
	totalCards += g.GetDeckRemaining()
	assert.Equal(t, 52, totalCards)

	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, GoFishPhasePlay, g.GetPhase())
	assert.Equal(t, 0, g.GetCurrentTurn())
	assert.Equal(t, 1, g.GetTurnNumber())
	assert.True(t, g.IsHumanTurn())
}

func TestGoFish_PlayerAsk_Success(t *testing.T) {
	g := newTestGoFish()
	// P0 (human) has: spade 3
	// P1 (cpu) has: heart 3, diamond 3
	g.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
	g.players[1].AddCard(NewCard(CardDesignHeart, 3, false))
	g.players[1].AddCard(NewCard(CardDesignDiamond, 3, false))
	// Give other players some cards
	g.players[2].AddCard(NewCard(CardDesignSpade, 5, false))
	g.players[3].AddCard(NewCard(CardDesignSpade, 7, false))

	err := g.PlayerAsk(1, 3)
	require.NoError(t, err)

	assert.True(t, g.GetLastAskSuccess())
	assert.Equal(t, 2, len(g.GetLastCardsReceived()))
	assert.Nil(t, g.GetLastDrawnCard())
	// P0 should now have 3 cards of rank 3
	assert.Equal(t, 3, g.players[0].GetCardsSize())
	// P1 should have 0 cards
	assert.Equal(t, 0, g.players[1].GetCardsSize())
	// Success → human gets another turn
	assert.Equal(t, 0, g.GetCurrentTurn())
	assert.True(t, g.IsHumanTurn())
}

func TestGoFish_PlayerAsk_GoFish(t *testing.T) {
	// Create a fresh game to have a deck
	players := []*GoFishPlayer{
		NewGoFishPlayer(true),
		NewGoFishPlayer(false),
		NewGoFishPlayer(false),
		NewGoFishPlayer(false),
	}
	g := NewGoFish(NewTrumpCards(0), players)
	g.Reset()

	// Clear hands and set up deterministic state
	for _, p := range g.players {
		p.Reset()
		p.ResetBooks()
	}
	g.currentTurn = 0
	g.turnNumber = 1

	g.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
	g.players[1].AddCard(NewCard(CardDesignHeart, 7, false))
	g.players[2].AddCard(NewCard(CardDesignSpade, 5, false))
	g.players[3].AddCard(NewCard(CardDesignSpade, 9, false))

	err := g.PlayerAsk(1, 3)
	require.NoError(t, err)

	assert.False(t, g.GetLastAskSuccess())
	assert.Nil(t, g.GetLastCardsReceived())
	// Drew a card from deck (not nil since deck has cards)
	assert.NotNil(t, g.GetLastDrawnCard())
}

func TestGoFish_PlayerAsk_BookFormed(t *testing.T) {
	g := newTestGoFish()
	// P0 has 3 cards of rank 5, P1 has the 4th
	g.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	g.players[0].AddCard(NewCard(CardDesignClover, 5, false))
	g.players[0].AddCard(NewCard(CardDesignHeart, 5, false))
	g.players[1].AddCard(NewCard(CardDesignDiamond, 5, false))
	g.players[2].AddCard(NewCard(CardDesignSpade, 7, false))
	g.players[3].AddCard(NewCard(CardDesignSpade, 9, false))

	err := g.PlayerAsk(1, 5)
	require.NoError(t, err)

	assert.True(t, g.GetLastAskSuccess())
	assert.True(t, g.GetLastBookFormed())
	assert.Equal(t, 5, g.GetLastBookRank())
	assert.Equal(t, 1, g.players[0].GetBookCount())
	assert.Equal(t, 0, g.players[0].GetCardsSize()) // All 4 went to book
}

func TestGoFish_PlayerAsk_Errors(t *testing.T) {
	g := newTestGoFish()
	g.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
	g.players[1].AddCard(NewCard(CardDesignHeart, 7, false))
	g.players[2].AddCard(NewCard(CardDesignSpade, 5, false))
	g.players[3].AddCard(NewCard(CardDesignSpade, 9, false))

	// Ask self
	err := g.PlayerAsk(0, 3)
	assert.Equal(t, ErrGoFishAskSelf, err)

	// Invalid target
	err = g.PlayerAsk(-1, 3)
	assert.Equal(t, ErrGoFishInvalidTarget, err)
	err = g.PlayerAsk(99, 3)
	assert.Equal(t, ErrGoFishInvalidTarget, err)

	// Rank not in hand
	err = g.PlayerAsk(1, 7)
	assert.Equal(t, ErrGoFishInvalidRank, err)

	// Game ended
	g.gameEndFlag = true
	err = g.PlayerAsk(1, 3)
	assert.Equal(t, ErrGoFishGameEnded, err)

	// Not your turn
	g.gameEndFlag = false
	g.currentTurn = 1
	err = g.PlayerAsk(0, 3)
	assert.Equal(t, ErrGoFishNotYourTurn, err)
}

func TestGoFish_PlayerAsk_TargetNoCards(t *testing.T) {
	g := newTestGoFish()
	g.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
	// P1 has no cards
	g.players[2].AddCard(NewCard(CardDesignSpade, 5, false))
	g.players[3].AddCard(NewCard(CardDesignSpade, 9, false))

	err := g.PlayerAsk(1, 3)
	assert.Equal(t, ErrGoFishInvalidTarget, err)
}

func TestGoFish_CpuAsk(t *testing.T) {
	g := newTestGoFish()
	g.config.CpuDifficulty = GoFishCpuDifficultyEasy
	g.currentTurn = 1 // CPU turn

	g.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
	g.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	g.players[2].AddCard(NewCard(CardDesignSpade, 7, false))
	g.players[3].AddCard(NewCard(CardDesignSpade, 9, false))

	err := g.CpuAsk()
	require.NoError(t, err)

	assert.Len(t, g.GetCpuActions(), 1)
	action := g.GetCpuActions()[0]
	assert.Equal(t, 1, action.AskPlayerIdx)
	assert.Equal(t, 5, action.AskRank) // Only has rank 5
}

func TestGoFish_CpuAsk_NotHumanTurn(t *testing.T) {
	g := newTestGoFish()
	g.currentTurn = 0 // Human turn

	g.players[0].AddCard(NewCard(CardDesignSpade, 3, false))

	err := g.CpuAsk()
	assert.Equal(t, ErrGoFishNotYourTurn, err)
}

func TestGoFish_CpuAsk_GameEnded(t *testing.T) {
	g := newTestGoFish()
	g.gameEndFlag = true
	g.currentTurn = 1

	err := g.CpuAsk()
	assert.Equal(t, ErrGoFishGameEnded, err)
}

func TestGoFish_GameEnd_AllBooksFormed(t *testing.T) {
	g := newTestGoFish()
	// Give P0 all 13 books directly
	for rank := 1; rank <= 13; rank++ {
		g.players[0].AddBook([]*Card{
			NewCard(CardDesignSpade, rank, false),
			NewCard(CardDesignClover, rank, false),
			NewCard(CardDesignHeart, rank, false),
			NewCard(CardDesignDiamond, rank, false),
		})
	}
	g.checkGameEnd()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, GoFishPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetWinnerIdx())
}

func TestGoFish_GameEnd_WinnerMostBooks(t *testing.T) {
	g := newTestGoFish()
	// P0: 3 books, P1: 5 books, P2: 3 books, P3: 2 books = 13 total
	for i := range 3 {
		g.players[0].AddBook([]*Card{NewCard(1, i+1, false), NewCard(2, i+1, false), NewCard(3, i+1, false), NewCard(4, i+1, false)})
	}
	for i := range 5 {
		g.players[1].AddBook([]*Card{NewCard(1, i+4, false), NewCard(2, i+4, false), NewCard(3, i+4, false), NewCard(4, i+4, false)})
	}
	for i := range 3 {
		g.players[2].AddBook([]*Card{NewCard(1, i+9, false), NewCard(2, i+9, false), NewCard(3, i+9, false), NewCard(4, i+9, false)})
	}
	for i := range 2 {
		g.players[3].AddBook([]*Card{NewCard(1, i+12, false), NewCard(2, i+12, false), NewCard(3, i+12, false), NewCard(4, i+12, false)})
	}
	g.checkGameEnd()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetWinnerIdx()) // P1 has most books
}

func TestGoFish_GetPlayer(t *testing.T) {
	g := newTestGoFish()
	assert.NotNil(t, g.GetPlayer(0))
	assert.NotNil(t, g.GetPlayer(3))
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(4))
}

func TestGoFish_JSON_RoundTrip(t *testing.T) {
	players := []*GoFishPlayer{
		NewGoFishPlayer(true),
		NewGoFishPlayer(false),
		NewGoFishPlayer(false),
		NewGoFishPlayer(false),
	}
	g := NewGoFish(NewTrumpCards(0), players)
	g.Reset()

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored GoFish
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, g.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, g.GetCurrentTurn(), restored.GetCurrentTurn())
	assert.Equal(t, g.GetGameEndFlag(), restored.GetGameEndFlag())
	assert.Equal(t, g.GetTurnNumber(), restored.GetTurnNumber())
	assert.Equal(t, g.GetDeckRemaining(), restored.GetDeckRemaining())
	assert.Equal(t, g.GetConfig(), restored.GetConfig())
}

// TestGoFish_UnmarshalJSON_RejectsOversizedCpuMemories pins ADR-0028's policy
// that every deserialised slice field is bounded — the GoFish unmarshaller
// previously had no guard at all, so a hostile session blob could allocate
// arbitrary heap before any check ran.
func TestGoFish_UnmarshalJSON_RejectsOversizedCpuMemories(t *testing.T) {
	var sb strings.Builder
	sb.WriteString(`{"cm":[`)
	for i := 0; i < goFishMaxSliceLen+1; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(`{"a":0,"g":1,"r":2,"h":false,"t":0}`)
	}
	sb.WriteString(`]}`)

	var g GoFish
	err := json.Unmarshal([]byte(sb.String()), &g)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

// TestGoFish_JSON_RoundTrip_PreservesCpuMemories pins #1655: goFishMemoryEntry
// fields are unexported, so the previous wire format silently flattened every
// entry to its zero value on restore. The round-trip must keep ask history.
func TestGoFish_JSON_RoundTrip_PreservesCpuMemories(t *testing.T) {
	g := newTestGoFish()
	g.cpuMemories = []goFishMemoryEntry{
		{askerIdx: 0, targetIdx: 2, rank: 5, hadCards: false, turnSeen: 1},
		{askerIdx: 1, targetIdx: 3, rank: 11, hadCards: true, turnSeen: 4},
	}

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored GoFish
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, g.cpuMemories, restored.cpuMemories)
}

func TestGoFish_ActionLog(t *testing.T) {
	g := newTestGoFish()
	g.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
	g.players[1].AddCard(NewCard(CardDesignHeart, 3, false))
	g.players[2].AddCard(NewCard(CardDesignSpade, 5, false))
	g.players[3].AddCard(NewCard(CardDesignSpade, 9, false))

	_ = g.PlayerAsk(1, 3)
	log := g.GetActionLog()
	assert.NotEmpty(t, log)
	assert.Equal(t, "ask_hit", log[0].ActionType)
}

func TestGoFish_CpuAskNormal(t *testing.T) {
	g := newTestGoFish()
	g.config.CpuDifficulty = GoFishCpuDifficultyNormal
	g.currentTurn = 1

	g.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
	g.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	g.players[2].AddCard(NewCard(CardDesignSpade, 7, false))
	g.players[3].AddCard(NewCard(CardDesignSpade, 9, false))

	// Add memory that P0 asked for rank 5
	g.cpuMemories = append(g.cpuMemories, goFishMemoryEntry{
		askerIdx: 0, targetIdx: 2, rank: 5, hadCards: false, turnSeen: 1,
	})

	err := g.CpuAsk()
	require.NoError(t, err)
	// CPU1 has rank 5, and memory says P0 asked for 5 → CPU might target P0
	assert.Len(t, g.GetCpuActions(), 1)
}

func TestGoFish_CpuAskHard(t *testing.T) {
	g := newTestGoFish()
	g.config.CpuDifficulty = GoFishCpuDifficultyHard
	g.currentTurn = 1

	g.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
	g.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	g.players[2].AddCard(NewCard(CardDesignSpade, 7, false))
	g.players[3].AddCard(NewCard(CardDesignSpade, 9, false))

	err := g.CpuAsk()
	require.NoError(t, err)
	assert.Len(t, g.GetCpuActions(), 1)
}

func TestGoFish_AdvanceTurn_SkipsEmptyPlayers(t *testing.T) {
	g := newTestGoFish()
	g.currentTurn = 0
	// P0 has cards, P1 empty, P2 has cards, P3 empty
	g.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
	g.players[2].AddCard(NewCard(CardDesignSpade, 5, false))

	g.advanceTurn()
	// Should skip P1 (empty, no deck) and go to P2
	assert.Equal(t, 2, g.GetCurrentTurn())
}

func TestGoFish_CpuAsk_EmptyHand_DrawFromDeck(t *testing.T) {
	players := []*GoFishPlayer{
		NewGoFishPlayer(true),
		NewGoFishPlayer(false),
		NewGoFishPlayer(false),
		NewGoFishPlayer(false),
	}
	g := NewGoFish(NewTrumpCards(0), players)
	g.Reset()

	// Clear CPU hand to test draw from deck
	g.players[1].Reset()
	g.currentTurn = 1
	g.config.CpuDifficulty = GoFishCpuDifficultyEasy

	err := g.CpuAsk()
	require.NoError(t, err)
	// CPU should have drawn a card from deck and then made an ask
}

func TestGoFish_HumanAction(t *testing.T) {
	g := newTestGoFish()
	g.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
	g.players[1].AddCard(NewCard(CardDesignHeart, 3, false))
	g.players[2].AddCard(NewCard(CardDesignSpade, 5, false))
	g.players[3].AddCard(NewCard(CardDesignSpade, 9, false))

	_ = g.PlayerAsk(1, 3)
	ha := g.GetHumanAction()
	require.NotNil(t, ha)
	assert.Equal(t, 0, ha.AskPlayerIdx)
	assert.Equal(t, 1, ha.AskTargetIdx)
	assert.Equal(t, 3, ha.AskRank)
	assert.True(t, ha.Success)
}

func TestGoFishMemoryEntry_GetTurnSeen(t *testing.T) {
	entry := goFishMemoryEntry{turnSeen: 5}
	assert.Equal(t, 5, entry.GetTurnSeen())
}
