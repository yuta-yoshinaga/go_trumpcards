//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupDurakForTransfer(enabled bool) *Durak {
	players := make([]*DurakPlayer, 4)
	players[0] = NewDurakPlayer(true)
	players[1] = NewDurakPlayer(true) // make player 1 human too
	players[2] = NewDurakPlayer(false)
	players[3] = NewDurakPlayer(false)

	d := NewDurak(NewTrumpCards(0), players)
	cfg := DefaultDurakConfig()
	cfg.TransferEnabled = enabled
	d.SetConfig(cfg)
	d.Reset()

	d.round.attackerIdx = 2
	d.round.defenderIdx = 0
	d.round.currentTurn = 0
	d.round.phase = DurakPhaseDefend

	attackCard := NewCard(CardDesignClover, 7, false)
	d.round.tablePairs = []*DurakTablePair{{Attack: attackCard}}

	d.players[0].Reset()
	d.players[0].AddCard(NewCard(CardDesignHeart, 7, false))
	d.players[0].AddCard(NewCard(CardDesignSpade, 10, false))

	d.players[1].Reset()
	d.players[1].AddCard(NewCard(CardDesignClover, 10, false)) // can beat club 7
	for i := 0; i < 5; i++ {
		d.players[1].AddCard(NewCard(CardDesignDiamond, i+1, false))
	}
	d.trumpSuit = int(CardDesignSpade)

	return d
}

func TestDurak_PlayerTransfer_Success(t *testing.T) {
	d := setupDurakForTransfer(true)
	err := d.PlayerTransfer(0)
	assert.NoError(t, err)

	pairs := d.GetTablePairs()
	assert.Equal(t, 2, len(pairs))
	assert.Equal(t, CardDesignHeart, pairs[1].Attack.GetDesign())

	assert.Equal(t, 0, d.GetAttackerIdx())
	assert.Equal(t, 1, d.GetDefenderIdx())
	assert.Equal(t, 1, d.GetCurrentTurn())
	assert.Equal(t, DurakPhaseDefend, d.GetPhase())

	err = d.PlayerDefend(0, 0)
	assert.NoError(t, err)
}

func TestDurak_PlayerTransfer_Disabled(t *testing.T) {
	d := setupDurakForTransfer(false)
	err := d.PlayerTransfer(0)
	assert.ErrorIs(t, err, ErrInvalidPlay)
}

func TestDurak_PlayerTransfer_AlreadyDefended(t *testing.T) {
	d := setupDurakForTransfer(true)
	d.round.tablePairs[0].Defense = NewCard(CardDesignSpade, 2, false)
	err := d.PlayerTransfer(0)
	assert.ErrorIs(t, err, ErrInvalidPlay)
}

func TestDurak_PlayerTransfer_RankMismatch(t *testing.T) {
	d := setupDurakForTransfer(true)
	err := d.PlayerTransfer(1) // spade 10
	assert.ErrorIs(t, err, ErrInvalidPlay)
}

func TestDurak_PlayerTransfer_NextDefenderCardsShort(t *testing.T) {
	d := setupDurakForTransfer(true)
	d.players[1].Reset()
	d.players[1].AddCard(NewCard(CardDesignClover, 10, false))
	// player 1 has only 1 card, but there are 2 attacks
	err := d.PlayerTransfer(0)
	assert.ErrorIs(t, err, ErrInvalidPlay)
}

func TestDurak_PlayerTransfer_NotDefenderTurn(t *testing.T) {
	d := setupDurakForTransfer(true)
	d.round.currentTurn = 2 // Not defender
	err := d.PlayerTransfer(0)
	assert.ErrorIs(t, err, ErrNotHumanTurn)
}

func TestDurak_PlayerTransfer_WrongPhase(t *testing.T) {
	d := setupDurakForTransfer(true)
	d.round.phase = DurakPhaseAttack
	err := d.PlayerTransfer(0)
	assert.ErrorIs(t, err, ErrWrongPhase)
}
