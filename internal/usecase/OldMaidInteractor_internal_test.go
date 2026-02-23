package usecase

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOldMaidInteractor_Draw_GameEndFlag(t *testing.T) {
	mockOutput := "mock"
	ompMock := new(presenter.MockOldMaidPresenter)
	ompMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	oi := NewOldMaidInteractor(ompMock)

	// Build a game where human has 1 card (the joker), one CPU has 1 card,
	// and the other CPUs are finished. Human draws from CPU → pair → both finish → game end.
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(true),  // player 0: human (current turn)
		domain.NewOldMaidPlayer(false), // player 1: CPU (has 1 card)
		domain.NewOldMaidPlayer(false), // player 2: finished
		domain.NewOldMaidPlayer(false), // player 3: finished
	}
	om := domain.NewOldMaid(domain.NewTrumpCards(1), players)

	// Mark players 2 and 3 as finished.
	players[2].SetIsFinished(true)
	players[3].SetIsFinished(true)

	// Give human exactly 1 card (spade A).
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	// Give CPU1 exactly 1 card (heart A) → will pair with human's spade A.
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

	oi.om = om

	// Human draws from CPU1 (card index 0 since CPU1 has 1 card).
	// → pairs are discarded → both finish → gameEndFlag = true.
	oi.Draw(0)
	assert.True(t, oi.om.GetGameEndFlag())

	// Now call Draw with gameEndFlag == true → early return.
	result := oi.Draw(0)
	assert.Equal(t, mockOutput, result)
}

func TestOldMaidInteractor_Draw_NotHumanTurn(t *testing.T) {
	mockOutput := "mock"
	ompMock := new(presenter.MockOldMaidPresenter)
	ompMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	oi := NewOldMaidInteractor(ompMock)

	// Build a game where currentTurn is a CPU player.
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(false), // player 0: CPU (current turn)
		domain.NewOldMaidPlayer(true),  // player 1: human
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
	}
	om := domain.NewOldMaid(domain.NewTrumpCards(1), players)

	// Give all players cards so game doesn't end.
	for i := 0; i < om.GetPlayerCnt(); i++ {
		// Give 2 different cards (no pairs) to each player.
		om.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignSpade, i*2+1, false))
		om.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignHeart, i*2+2, false))
	}
	oi.om = om

	// currentTurn = 0, player 0 is CPU → !IsHumanTurn() is true.
	assert.False(t, oi.om.IsHumanTurn())
	result := oi.Draw(0)
	assert.Equal(t, mockOutput, result)
}
