//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newBeggarMyNeighbourForTest() *BeggarMyNeighbour {
	players := []*BeggarMyNeighbourPlayer{
		NewBeggarMyNeighbourPlayer(true),
		NewBeggarMyNeighbourPlayer(false),
	}
	return NewBeggarMyNeighbour(NewTrumpCards(0), players, DefaultBeggarMyNeighbourConfig())
}

// setupBeggarMyNeighbourWithPiles テスト用に山札を直接設定する
func setupBeggarMyNeighbourWithPiles(t *testing.T, playerCards, cpuCards []*Card) *BeggarMyNeighbour {
	t.Helper()
	g := newBeggarMyNeighbourForTest()
	g.Reset()
	g.players[0].ResetPiles()
	g.players[1].ResetPiles()
	g.players[0].AddToDrawPile(playerCards...)
	g.players[1].AddToDrawPile(cpuCards...)
	g.centralPile = nil
	g.lastCardPlayed = nil
	g.phase = BeggarMyNeighbourPhasePlay
	g.currentPlayerIdx = 0
	g.penaltyOwnerIdx = -1
	g.penaltyRemaining = 0
	g.roundsPlayed = 0
	g.gameEndFlag = false
	g.winnerIdx = -1
	return g
}

func bmnCard(value int) *Card {
	return NewCard(CardDesignSpade, value, false)
}

func TestBeggarMyNeighbour_Reset_Deals26Each(t *testing.T) {
	g := newBeggarMyNeighbourForTest()
	g.Reset()
	assert.Equal(t, BeggarMyNeighbourPhasePlay, g.GetPhase())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, 26, g.GetPlayer(0).GetDrawPileSize())
	assert.Equal(t, 26, g.GetPlayer(1).GetDrawPileSize())
	assert.Equal(t, 0, g.GetCentralPileSize())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.Equal(t, -1, g.GetPenaltyOwnerIdx())
	assert.Equal(t, -1, g.GetWinnerIdx())
}

func TestBeggarMyNeighbour_PenaltyValue(t *testing.T) {
	assert.Equal(t, 1, beggarMyNeighbourPenaltyValue(bmnCard(11))) // J
	assert.Equal(t, 2, beggarMyNeighbourPenaltyValue(bmnCard(12))) // Q
	assert.Equal(t, 3, beggarMyNeighbourPenaltyValue(bmnCard(13))) // K
	assert.Equal(t, 4, beggarMyNeighbourPenaltyValue(bmnCard(1)))  // A
	assert.Equal(t, 0, beggarMyNeighbourPenaltyValue(bmnCard(5)))  // number
	assert.Equal(t, 0, beggarMyNeighbourPenaltyValue(nil))
}

func TestBeggarMyNeighbour_Play_NormalCard_AdvancesTurn(t *testing.T) {
	// Player has 2, CPU has 3 - no penalty cards
	g := setupBeggarMyNeighbourWithPiles(t,
		[]*Card{bmnCard(2), bmnCard(4)},
		[]*Card{bmnCard(3), bmnCard(5)},
	)
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.NoError(t, g.Step())
	// After player plays non-penalty, turn switches to CPU
	assert.Equal(t, BeggarMyNeighbourPhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	assert.Equal(t, 1, g.GetCentralPileSize())
	assert.NotNil(t, g.GetLastCardPlayed())
}

func TestBeggarMyNeighbour_Play_PenaltyCard_TriggersPayPenalty(t *testing.T) {
	// Player plays J (penalty=1): CPU must pay 1
	g := setupBeggarMyNeighbourWithPiles(t,
		[]*Card{bmnCard(11)}, // J
		[]*Card{bmnCard(5)},
	)
	assert.NoError(t, g.Step())
	assert.Equal(t, BeggarMyNeighbourPhasePayPenalty, g.GetPhase())
	assert.Equal(t, 0, g.GetPenaltyOwnerIdx()) // player 0 played the penalty
	assert.Equal(t, 1, g.GetPenaltyRemaining())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx()) // CPU must pay
}

func TestBeggarMyNeighbour_PayPenalty_AllPaid_GoesToCollect(t *testing.T) {
	// Player plays J (1 penalty), CPU pays 1 non-penalty card → collect phase
	g := setupBeggarMyNeighbourWithPiles(t,
		[]*Card{bmnCard(11), bmnCard(4)}, // J then 4
		[]*Card{bmnCard(5)},              // CPU pays 5 (non-penalty)
	)
	// Step 1: player plays J
	assert.NoError(t, g.Step())
	assert.Equal(t, BeggarMyNeighbourPhasePayPenalty, g.GetPhase())
	assert.Equal(t, 1, g.GetPenaltyRemaining())

	// Step 2: CPU pays 5 (non-penalty), penaltyRemaining becomes 0
	assert.NoError(t, g.Step())
	assert.Equal(t, BeggarMyNeighbourPhaseCollect, g.GetPhase())
}

func TestBeggarMyNeighbour_Collect_PlayerGetsCards(t *testing.T) {
	// Player plays J, CPU pays one non-penalty card, then player collects.
	// Give CPU 2 cards so it still has cards after paying 1, preventing immediate game-end.
	g := setupBeggarMyNeighbourWithPiles(t,
		[]*Card{bmnCard(11), bmnCard(4)}, // J then 4
		[]*Card{bmnCard(5), bmnCard(6)},  // CPU pays 5, keeps 6
	)
	assert.NoError(t, g.Step()) // player plays J
	assert.NoError(t, g.Step()) // CPU pays 5 → collect phase
	assert.Equal(t, BeggarMyNeighbourPhaseCollect, g.GetPhase())

	// Step 3: collect - player 0 (penaltyOwnerIdx) gets the pile
	assert.NoError(t, g.Step())
	assert.Equal(t, BeggarMyNeighbourPhasePlay, g.GetPhase())
	// Player 0 should have collected 2 cards (J + 5) into discardPile
	assert.Equal(t, 2, g.GetPlayer(0).GetDiscardPileSize())
	// currentPlayerIdx should be the collector (0 = Player) who leads the next round
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.Equal(t, 1, g.GetRoundsPlayed())
}

func TestBeggarMyNeighbour_PayPenalty_NewPenaltyFlipsObligation(t *testing.T) {
	// Player plays Q (2 penalty), CPU pays J (1 penalty) → obligation flips back to player
	g := setupBeggarMyNeighbourWithPiles(t,
		[]*Card{bmnCard(12), bmnCard(5)}, // Q, then 5
		[]*Card{bmnCard(11), bmnCard(7)}, // J, then 7
	)
	// Step 1: player plays Q → CPU must pay 2
	assert.NoError(t, g.Step())
	assert.Equal(t, BeggarMyNeighbourPhasePayPenalty, g.GetPhase())
	assert.Equal(t, 0, g.GetPenaltyOwnerIdx())
	assert.Equal(t, 2, g.GetPenaltyRemaining())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())

	// Step 2: CPU pays J (penalty=1) → obligation flips to CPU (currentPlayerIdx=1 becomes penaltyOwner)
	assert.NoError(t, g.Step())
	assert.Equal(t, BeggarMyNeighbourPhasePayPenalty, g.GetPhase())
	assert.Equal(t, 1, g.GetPenaltyOwnerIdx())
	assert.Equal(t, 1, g.GetPenaltyRemaining())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx()) // player must now pay
}

func TestBeggarMyNeighbour_AcePenalty_FourCards(t *testing.T) {
	// Player plays A (4 penalty), CPU pays 4 non-penalty cards
	g := setupBeggarMyNeighbourWithPiles(t,
		[]*Card{bmnCard(1), bmnCard(2)},                         // A
		[]*Card{bmnCard(5), bmnCard(6), bmnCard(7), bmnCard(8)}, // 4 non-penalty
	)
	assert.NoError(t, g.Step()) // player plays A
	assert.Equal(t, BeggarMyNeighbourPhasePayPenalty, g.GetPhase())
	assert.Equal(t, 4, g.GetPenaltyRemaining())

	for range 4 {
		assert.NoError(t, g.Step())
	}
	assert.Equal(t, BeggarMyNeighbourPhaseCollect, g.GetPhase())
}

func TestBeggarMyNeighbour_GameEnd_PlayerWins(t *testing.T) {
	// Player plays J, CPU has no cards to pay → game ends
	g := setupBeggarMyNeighbourWithPiles(t,
		[]*Card{bmnCard(11)},
		[]*Card{},
	)
	// CPU has no cards, so playing J leads to a finishByTotal check eventually
	// Let's start with player having all cards
	// Actually the game would end on the next step when CPU tries to draw
	assert.NoError(t, g.Step()) // player plays J → PayPenalty, CPU must pay
	// CPU has no cards → finishByTotal
	assert.NoError(t, g.Step())
	assert.True(t, g.GetGameEndFlag())
	// Player had 1 card (J on pile), CPU had 0, so player wins
	assert.Equal(t, 0, g.GetWinnerIdx())
}

func TestBeggarMyNeighbour_GameEnd_Draw(t *testing.T) {
	// Player plays J, CPU pays one non-penalty card, player collects (2 cards),
	// leaving both players with 2 cards. With MaxRounds=1 the round cap fires on
	// collect and finishByTotal sees an equal split → genuine draw (winnerIdx -1).
	g := setupBeggarMyNeighbourWithPiles(t,
		[]*Card{bmnCard(11)},                        // J
		[]*Card{bmnCard(5), bmnCard(6), bmnCard(7)}, // pays 5, keeps 6,7
	)
	g.SetConfig(BeggarMyNeighbourConfig{MaxRounds: 1})

	assert.NoError(t, g.Step()) // player plays J → PayPenalty
	assert.NoError(t, g.Step()) // CPU pays 5 → Collect
	assert.NoError(t, g.Step()) // player collects → MaxRounds cap fires
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 2, g.GetPlayer(0).TotalCards())
	assert.Equal(t, 2, g.GetPlayer(1).TotalCards())
	assert.Equal(t, -1, g.GetWinnerIdx()) // draw
	assert.False(t, g.GetPlayer(0).GetIsFinished())
	assert.False(t, g.GetPlayer(1).GetIsFinished())
}

func TestBeggarMyNeighbour_AutoPlay_RunsToEnd(t *testing.T) {
	g := newBeggarMyNeighbourForTest()
	g.Reset()
	assert.NoError(t, g.AutoPlay())
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, BeggarMyNeighbourPhaseGameEnd, g.GetPhase())
}

func TestBeggarMyNeighbour_AutoPlay_AfterEndReturnsErr(t *testing.T) {
	g := setupBeggarMyNeighbourWithPiles(t,
		[]*Card{bmnCard(11)},
		[]*Card{},
	)
	assert.NoError(t, g.AutoPlay())
	assert.True(t, g.GetGameEndFlag())
	err := g.AutoPlay()
	assert.ErrorIs(t, err, ErrGameEnded)
}

func TestBeggarMyNeighbour_AutoPlay_HitsCapWithoutFinishing(t *testing.T) {
	orig := beggarMyNeighbourAutoPlayMaxSteps
	beggarMyNeighbourAutoPlayMaxSteps = 1
	t.Cleanup(func() { beggarMyNeighbourAutoPlayMaxSteps = orig })

	g := setupBeggarMyNeighbourWithPiles(t,
		[]*Card{bmnCard(2), bmnCard(3), bmnCard(4)},
		[]*Card{bmnCard(5), bmnCard(6), bmnCard(7)},
	)
	err := g.AutoPlay()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auto-play reached maximum steps")
}

func TestBeggarMyNeighbour_StepAfterEnd(t *testing.T) {
	g := setupBeggarMyNeighbourWithPiles(t,
		[]*Card{bmnCard(11)},
		[]*Card{},
	)
	assert.NoError(t, g.AutoPlay())
	assert.True(t, g.GetGameEndFlag())
	err := g.Step()
	assert.ErrorIs(t, err, ErrGameEnded)
}

func TestBeggarMyNeighbour_JSON(t *testing.T) {
	g := setupBeggarMyNeighbourWithPiles(t,
		[]*Card{bmnCard(11), bmnCard(4)},
		[]*Card{bmnCard(5)},
	)
	assert.NoError(t, g.Step()) // play J → PayPenalty

	data, err := json.Marshal(g)
	assert.NoError(t, err)

	decoded := NewBeggarMyNeighbour(
		NewTrumpCards(0),
		[]*BeggarMyNeighbourPlayer{NewBeggarMyNeighbourPlayer(true), NewBeggarMyNeighbourPlayer(false)},
		DefaultBeggarMyNeighbourConfig(),
	)
	assert.NoError(t, json.Unmarshal(data, decoded))
	assert.Equal(t, g.GetPhase(), decoded.GetPhase())
	assert.Equal(t, g.GetPenaltyOwnerIdx(), decoded.GetPenaltyOwnerIdx())
	assert.Equal(t, g.GetPenaltyRemaining(), decoded.GetPenaltyRemaining())
	assert.Equal(t, g.GetCentralPileSize(), decoded.GetCentralPileSize())
	assert.Equal(t, g.GetCurrentPlayerIdx(), decoded.GetCurrentPlayerIdx())
}

func TestBeggarMyNeighbour_Config_GetSet(t *testing.T) {
	g := newBeggarMyNeighbourForTest()
	g.Reset()
	cfg := BeggarMyNeighbourConfig{MaxRounds: 500}
	g.SetConfig(cfg)
	assert.Equal(t, cfg, g.GetConfig())
}

func TestBeggarMyNeighbour_IsHumanTurn(t *testing.T) {
	g := newBeggarMyNeighbourForTest()
	g.Reset()
	assert.True(t, g.IsHumanTurn())
}

func TestBeggarMyNeighbour_GetPlayerCnt(t *testing.T) {
	g := newBeggarMyNeighbourForTest()
	assert.Equal(t, BeggarMyNeighbourPlayerCnt, g.GetPlayerCnt())
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
}

func TestBeggarMyNeighbour_ActionLog(t *testing.T) {
	g := setupBeggarMyNeighbourWithPiles(t,
		[]*Card{bmnCard(5), bmnCard(6)},
		[]*Card{bmnCard(7), bmnCard(8)},
	)
	assert.NoError(t, g.Step())
	logs := g.GetActionLog()
	assert.NotEmpty(t, logs)
	assert.Equal(t, "play", logs[0].ActionType)
}

func TestBeggarMyNeighbour_MaxRoundsTimeout(t *testing.T) {
	g := setupBeggarMyNeighbourWithPiles(t,
		[]*Card{bmnCard(11), bmnCard(4), bmnCard(5)},
		[]*Card{bmnCard(12), bmnCard(6), bmnCard(7)},
	)
	g.config = BeggarMyNeighbourConfig{MaxRounds: 1}
	// Run enough steps to complete a collect cycle
	for i := range 20 {
		if g.GetGameEndFlag() {
			break
		}
		_ = g.Step()
		_ = i
	}
	assert.True(t, g.GetGameEndFlag())
}
