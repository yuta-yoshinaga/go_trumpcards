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

	// Drive game to ended state: mark all players as finished.
	config := domain.DefaultDaifugoConfig()
	players := []*domain.DaifugoPlayer{
		domain.NewDaifugoPlayer(true),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
	}
	dg := domain.NewDaifugo(domain.NewTrumpCards(config.JokerCount), players, config)

	// Set all players as finished to trigger checkGameEnd via public API.
	for i := 0; i < dg.GetPlayerCnt(); i++ {
		dg.GetPlayer(i).SetIsFinished(true)
		dg.GetPlayer(i).SetRank(i + 1)
	}
	// Replace the private dg field with our prepared game.
	di.dg = dg

	// Now call Reset to trigger checkGameEnd which sets gameEndFlag.
	// Actually, Reset() calls dg.Reset() which clears gameEndFlag.
	// Instead, we need to craft a state where gameEndFlag is already true.
	// We can do this by manually playing the game to completion.
	// Simpler: build a game, give human 1 card, all others 0 cards (finished),
	// then human plays that card → finishes → checkGameEnd sets flag.

	// Rebuild with a controlled setup.
	players2 := []*domain.DaifugoPlayer{
		domain.NewDaifugoPlayer(true),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
	}
	dg2 := domain.NewDaifugo(domain.NewTrumpCards(0), players2, config)
	// Mark CPU players as finished.
	for i := 1; i < dg2.GetPlayerCnt(); i++ {
		dg2.GetPlayer(i).SetIsFinished(true)
		dg2.GetPlayer(i).SetRank(i)
	}
	// Give human a single card so human can play it → finish.
	card := domain.NewCard(domain.CardDesignSpade, 5, false)
	dg2.GetPlayer(0).AddCard(card)
	di.dg = dg2

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
