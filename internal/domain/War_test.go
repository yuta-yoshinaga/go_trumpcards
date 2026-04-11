//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newWarForTest() *War {
	players := []*WarPlayer{NewWarPlayer(true), NewWarPlayer(false)}
	return NewWar(NewTrumpCards(0), players, DefaultWarConfig())
}

// setupWarWithPiles テスト用に山札を直接設定する
func setupWarWithPiles(t *testing.T, playerCards, cpuCards []*Card) *War {
	t.Helper()
	w := newWarForTest()
	w.Reset()
	// Reset で配られたカードを破棄して、明示的な山札に差し替える
	w.players[0].ResetPiles()
	w.players[1].ResetPiles()
	w.players[0].AddToDrawPile(playerCards...)
	w.players[1].AddToDrawPile(cpuCards...)
	w.warPot = nil
	w.playerRevealed = nil
	w.cpuRevealed = nil
	w.phase = WarPhaseReveal
	w.roundsPlayed = 0
	w.gameEndFlag = false
	w.winnerIdx = -1
	w.lastWinnerIdx = -1
	return w
}

func card(value int) *Card {
	return NewCard(CardDesignSpade, value, false)
}

func TestWar_Reset_Deals26Each(t *testing.T) {
	w := newWarForTest()
	w.Reset()
	assert.Equal(t, WarPhaseReveal, w.GetPhase())
	assert.False(t, w.GetGameEndFlag())
	assert.Equal(t, 26, w.GetPlayer(0).GetDrawPileSize())
	assert.Equal(t, 26, w.GetPlayer(1).GetDrawPileSize())
	assert.Equal(t, 0, w.GetWarPotSize())
}

func TestWar_WarRank_AceHigh(t *testing.T) {
	assert.Equal(t, 14, warRank(card(1)))
	assert.Equal(t, 13, warRank(card(13)))
	assert.Equal(t, 2, warRank(card(2)))
	assert.Equal(t, 0, warRank(nil))
}

func TestWar_Reveal_PlayerWins(t *testing.T) {
	w := setupWarWithPiles(t,
		[]*Card{card(10)}, // Player: 10
		[]*Card{card(5)},  // CPU: 5
	)
	assert.NoError(t, w.Step())
	assert.Equal(t, WarPhaseResolved, w.GetPhase())
	assert.Equal(t, 0, w.GetLastWinnerIdx())
	assert.Equal(t, 2, w.GetWarPotSize())
	assert.NotNil(t, w.GetPlayerRevealed())
	assert.NotNil(t, w.GetCpuRevealed())

	// Resolve: pot goes to player 0
	assert.NoError(t, w.Step())
	assert.Equal(t, WarPhaseGameEnd, w.GetPhase())
	assert.Equal(t, 0, w.GetWinnerIdx())
	assert.Equal(t, 2, w.GetPlayer(0).GetDiscardPileSize())
}

func TestWar_Reveal_CpuWins(t *testing.T) {
	w := setupWarWithPiles(t,
		[]*Card{card(3), card(2)},
		[]*Card{card(11), card(2)},
	)
	assert.NoError(t, w.Step())
	assert.Equal(t, 1, w.GetLastWinnerIdx())
	assert.NoError(t, w.Step())
	// after resolve: both have 1 card left, phase back to Reveal
	assert.Equal(t, WarPhaseReveal, w.GetPhase())
	assert.Equal(t, 2, w.GetPlayer(1).GetDiscardPileSize())
}

func TestWar_AceBeatsKing(t *testing.T) {
	w := setupWarWithPiles(t,
		[]*Card{card(1)}, // Ace
		[]*Card{card(13)},
	)
	assert.NoError(t, w.Step())
	assert.Equal(t, 0, w.GetLastWinnerIdx(), "Ace (high) beats King")
}

func TestWar_TieTriggersWar(t *testing.T) {
	// Both reveal 7 -> war. Each has 3 buried + face-up:
	// Player: [7, b1, b2, b3, 10]  (draw pile bottom to top as indexed)
	// CPU   : [7, b1, b2, b3, 5]
	w := setupWarWithPiles(t,
		[]*Card{card(7), card(2), card(2), card(2), card(10)},
		[]*Card{card(7), card(3), card(3), card(3), card(5)},
	)
	// Step 1: reveal -> tie -> WarBury
	assert.NoError(t, w.Step())
	assert.Equal(t, WarPhaseWarBury, w.GetPhase())
	assert.Equal(t, 2, w.GetWarPotSize())

	// Step 2: bury 3 each + reveal new faces (10 vs 5 -> player wins)
	assert.NoError(t, w.Step())
	assert.Equal(t, WarPhaseResolved, w.GetPhase())
	assert.Equal(t, 0, w.GetLastWinnerIdx())
	assert.Equal(t, 3, w.GetLastBurialCount())
	assert.Equal(t, 10, w.GetWarPotSize()) // 2 + 6 buried + 2 new faces

	// Step 3: award pot to player
	assert.NoError(t, w.Step())
	assert.Equal(t, 10, w.GetPlayer(0).GetDiscardPileSize())
	assert.Equal(t, 0, w.GetPlayer(1).TotalCards())
	assert.True(t, w.GetGameEndFlag())
	assert.Equal(t, 0, w.GetWinnerIdx())
}

func TestWar_NestedWar(t *testing.T) {
	// First reveal ties at 5, war burial yields another tie at 6, third reveal resolves
	w := setupWarWithPiles(t,
		[]*Card{card(5), card(2), card(2), card(2), card(6), card(2), card(2), card(2), card(12)},
		[]*Card{card(5), card(3), card(3), card(3), card(6), card(3), card(3), card(3), card(4)},
	)
	assert.NoError(t, w.Step()) // reveal 5=5 -> WarBury
	assert.Equal(t, WarPhaseWarBury, w.GetPhase())
	assert.NoError(t, w.Step()) // bury3+reveal 6=6 -> WarBury again
	assert.Equal(t, WarPhaseWarBury, w.GetPhase())
	assert.NoError(t, w.Step()) // bury3+reveal 12 vs 4 -> player
	assert.Equal(t, WarPhaseResolved, w.GetPhase())
	assert.Equal(t, 0, w.GetLastWinnerIdx())
	assert.Equal(t, 18, w.GetWarPotSize())
}

func TestWar_ShortWar_FewerThanFourCards(t *testing.T) {
	// Tie on the opener, but CPU only has 2 cards remaining after reveal -> bury 1 + reveal 1
	w := setupWarWithPiles(t,
		[]*Card{card(8), card(2), card(2), card(2), card(9)},
		[]*Card{card(8), card(3), card(2)}, // 3 cards; after reveal 2 remain -> bury 1, reveal 1
	)
	assert.NoError(t, w.Step()) // reveal 8=8
	assert.Equal(t, WarPhaseWarBury, w.GetPhase())
	assert.NoError(t, w.Step()) // player buries 3 + reveal 9, cpu buries 1 + reveal 2
	assert.Equal(t, WarPhaseResolved, w.GetPhase())
	assert.Equal(t, 0, w.GetLastWinnerIdx())
	assert.NoError(t, w.Step()) // game ends (cpu empty)
	assert.True(t, w.GetGameEndFlag())
	assert.Equal(t, 0, w.GetWinnerIdx())
}

func TestWar_RunOutDuringReveal(t *testing.T) {
	w := setupWarWithPiles(t,
		[]*Card{card(10), card(11)},
		[]*Card{card(5)},
	)
	assert.NoError(t, w.Step()) // reveal round 1
	assert.NoError(t, w.Step()) // resolve -> cpu empty -> finishByTotal
	assert.True(t, w.GetGameEndFlag())
	assert.Equal(t, 0, w.GetWinnerIdx())
}

func TestWar_DiscardRefill(t *testing.T) {
	// Player has 1 card in draw pile + 3 in discard; DrawOne should auto-refill from discard.
	w := newWarForTest()
	w.Reset()
	w.players[0].ResetPiles()
	w.players[1].ResetPiles()
	w.players[0].AddToDrawPile(card(10))
	w.players[0].AddToDiscardPile(card(4), card(6), card(8))
	w.players[1].AddToDrawPile(card(5), card(5), card(5), card(5))
	w.phase = WarPhaseReveal
	w.roundsPlayed = 0

	// Round 1: 10 vs 5 -> player wins
	assert.NoError(t, w.Step())
	assert.Equal(t, 0, w.GetLastWinnerIdx())
	assert.NoError(t, w.Step())

	// Round 2: player draws -> triggers refill from discard
	assert.NoError(t, w.Step())
	// Player revealed card must come from the refilled pile
	assert.NotNil(t, w.GetPlayerRevealed())
}

func TestWar_MaxRoundsTimeout(t *testing.T) {
	w := setupWarWithPiles(t,
		[]*Card{card(10), card(10)},
		[]*Card{card(5), card(5)},
	)
	w.config = WarConfig{MaxRounds: 1}
	assert.NoError(t, w.Step()) // reveal
	assert.NoError(t, w.Step()) // resolve -> roundsPlayed==1 -> timeout -> finishByTotal
	assert.True(t, w.GetGameEndFlag())
	assert.Equal(t, 0, w.GetWinnerIdx()) // player has more cards
}

func TestWar_StepAfterEnd(t *testing.T) {
	w := setupWarWithPiles(t, []*Card{card(10)}, []*Card{card(5)})
	assert.NoError(t, w.Step())
	assert.NoError(t, w.Step())
	assert.True(t, w.GetGameEndFlag())
	err := w.Step()
	assert.ErrorIs(t, err, ErrGameEnded)
}

func TestWar_JSON(t *testing.T) {
	w := setupWarWithPiles(t, []*Card{card(10)}, []*Card{card(5)})
	assert.NoError(t, w.Step())

	data, err := json.Marshal(w)
	assert.NoError(t, err)

	decoded := NewWar(NewTrumpCards(0), []*WarPlayer{NewWarPlayer(true), NewWarPlayer(false)}, DefaultWarConfig())
	assert.NoError(t, json.Unmarshal(data, decoded))
	assert.Equal(t, w.GetPhase(), decoded.GetPhase())
	assert.Equal(t, w.GetLastWinnerIdx(), decoded.GetLastWinnerIdx())
	assert.Equal(t, w.GetWarPotSize(), decoded.GetWarPotSize())
}

func TestWar_Config_GetSet(t *testing.T) {
	w := newWarForTest()
	w.Reset()
	cfg := WarConfig{MaxRounds: 123}
	w.SetConfig(cfg)
	assert.Equal(t, cfg, w.GetConfig())
}

func TestWar_IsHumanTurn(t *testing.T) {
	w := newWarForTest()
	w.Reset()
	assert.True(t, w.IsHumanTurn())
}

func TestWar_RevealRunOut_PlayerHasNone(t *testing.T) {
	// Player's pile empty; CPU retains cards after the drained reveal.
	w := setupWarWithPiles(t, []*Card{}, []*Card{card(5), card(6)})
	assert.NoError(t, w.Step())
	assert.True(t, w.GetGameEndFlag())
	assert.Equal(t, 1, w.GetWinnerIdx())
}

func TestWar_RevealRunOut_CpuHasNone(t *testing.T) {
	w := setupWarWithPiles(t, []*Card{card(5), card(6)}, []*Card{})
	assert.NoError(t, w.Step())
	assert.True(t, w.GetGameEndFlag())
	assert.Equal(t, 0, w.GetWinnerIdx())
}

func TestWar_WarBury_PlayerRunsOut(t *testing.T) {
	// Tie on opener; player has no card to reveal face-up after burial.
	// CPU has extra cards so it retains some after burial + face-up draw.
	w := setupWarWithPiles(t,
		[]*Card{card(8)},
		[]*Card{card(8), card(2), card(2), card(2), card(9), card(9), card(9)},
	)
	assert.NoError(t, w.Step()) // reveal tie -> WarBury
	assert.Equal(t, WarPhaseWarBury, w.GetPhase())
	assert.NoError(t, w.Step()) // player can't produce face-up -> finishByTotal
	assert.True(t, w.GetGameEndFlag())
	assert.Equal(t, 1, w.GetWinnerIdx())
}

func TestWar_WarBury_CpuRunsOut(t *testing.T) {
	w := setupWarWithPiles(t,
		[]*Card{card(8), card(2), card(2), card(2), card(9), card(9), card(9)},
		[]*Card{card(8)},
	)
	assert.NoError(t, w.Step())
	assert.Equal(t, WarPhaseWarBury, w.GetPhase())
	assert.NoError(t, w.Step())
	assert.True(t, w.GetGameEndFlag())
	assert.Equal(t, 0, w.GetWinnerIdx())
}

func TestWar_GetPlayerCnt(t *testing.T) {
	w := newWarForTest()
	assert.Equal(t, WarPlayerCnt, w.GetPlayerCnt())
	assert.Nil(t, w.GetPlayer(-1))
	assert.Nil(t, w.GetPlayer(99))
}

func TestWar_ActionLog(t *testing.T) {
	w := setupWarWithPiles(t, []*Card{card(10)}, []*Card{card(5)})
	assert.NoError(t, w.Step())
	logs := w.GetActionLog()
	assert.NotEmpty(t, logs)
	assert.Equal(t, "reveal", logs[0].ActionType)
}
