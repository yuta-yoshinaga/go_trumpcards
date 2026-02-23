package usecase

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDaifugoInteractor_Play_GameEndFlag(t *testing.T) {
	mockOutput := "mock"
	dgpMock := new(presenter.MockDaifugoPresenter)
	dgpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	di := NewDaifugoInteractor(dgpMock)

	// Human has 1 card and all CPUs are finished. When human plays the last card,
	// checkGameEnd sets gameEndFlag = true.
	config := domain.DefaultDaifugoConfig()
	players := []*domain.DaifugoPlayer{
		domain.NewDaifugoPlayer(true),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
	}
	dg := domain.NewDaifugo(domain.NewTrumpCards(0), players, config)
	for i := 1; i < dg.GetPlayerCnt(); i++ {
		dg.GetPlayer(i).SetIsFinished(true)
		dg.GetPlayer(i).SetRank(i)
	}
	dg.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	di.dg = dg

	// Play the card at index 0 → human finishes → checkGameEnd → gameEndFlag = true.
	di.Play([]int{0})
	assert.True(t, di.dg.GetGameEndFlag())

	// Now call Play again with gameEndFlag == true → early return.
	result := di.Play([]int{})
	assert.Equal(t, mockOutput, result)
}

func TestDaifugoInteractor_Play_NotHumanTurn(t *testing.T) {
	mockOutput := "mock"
	dgpMock := new(presenter.MockDaifugoPresenter)
	dgpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	di := NewDaifugoInteractor(dgpMock)

	config := domain.DefaultDaifugoConfig()
	players := []*domain.DaifugoPlayer{
		domain.NewDaifugoPlayer(false), // player 0: CPU (current turn)
		domain.NewDaifugoPlayer(true),  // player 1: human
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
	}
	dg := domain.NewDaifugo(domain.NewTrumpCards(0), players, config)
	// Give all players 1 card so game doesn't end.
	for i := 0; i < dg.GetPlayerCnt(); i++ {
		dg.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignSpade, i+1, false))
	}
	di.dg = dg

	// currentTurn = 0, player 0 is CPU → !IsHumanTurn() is true.
	assert.False(t, di.dg.IsHumanTurn())
	result := di.Play([]int{})
	assert.Equal(t, mockOutput, result)
}
