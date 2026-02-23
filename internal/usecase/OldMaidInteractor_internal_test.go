package usecase

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// buildOldMaidEndedGame creates an OldMaid game where human has 1 card (spade A),
// one CPU has 1 card (heart A), and all other CPUs are finished.
// After human draws from CPU, the pair is discarded and the game ends.
func buildOldMaidEndedGame() *domain.OldMaid {
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(true),  // player 0: human (current turn)
		domain.NewOldMaidPlayer(false), // player 1: CPU (has 1 card)
		domain.NewOldMaidPlayer(false), // player 2: finished
		domain.NewOldMaidPlayer(false), // player 3: finished
	}
	om := domain.NewOldMaid(domain.NewTrumpCards(1), players)
	players[2].SetIsFinished(true)
	players[3].SetIsFinished(true)
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
	return om
}

// buildOldMaidCpuTurnGame creates an OldMaid game where currentTurn is a CPU.
func buildOldMaidCpuTurnGame() *domain.OldMaid {
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(false), // player 0: CPU (current turn)
		domain.NewOldMaidPlayer(true),  // player 1: human
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
	}
	om := domain.NewOldMaid(domain.NewTrumpCards(1), players)
	for i := 0; i < om.GetPlayerCnt(); i++ {
		om.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignSpade, i*2+1, false))
		om.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignHeart, i*2+2, false))
	}
	return om
}

func TestOldMaidInteractor_Draw_GameEndFlag(t *testing.T) {
	mockOutput := "mock"
	ompMock := new(presenter.MockOldMaidPresenter)
	ompMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	oi := NewOldMaidInteractor(ompMock)
	oi.om = buildOldMaidEndedGame()

	// Human draws from CPU1 (card index 0) → pair discarded → game ends.
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
	oi.om = buildOldMaidCpuTurnGame()

	// currentTurn = 0, player 0 is CPU → !IsHumanTurn() is true.
	assert.False(t, oi.om.IsHumanTurn())
	result := oi.Draw(0)
	assert.Equal(t, mockOutput, result)
}
