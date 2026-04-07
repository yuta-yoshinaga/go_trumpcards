package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// buildDaifugoEndedGame creates a Daifugo game where human has 1 card and
// all CPUs are finished. After human plays the card, checkGameEnd sets gameEndFlag.
func buildDaifugoEndedGame() *domain.Daifugo {
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
	return dg
}

// buildDaifugoCpuTurnGame creates a Daifugo game where currentTurn is a CPU.
func buildDaifugoCpuTurnGame() *domain.Daifugo {
	config := domain.DefaultDaifugoConfig()
	players := []*domain.DaifugoPlayer{
		domain.NewDaifugoPlayer(false), // player 0: CPU (current turn)
		domain.NewDaifugoPlayer(true),  // player 1: human
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
	}
	dg := domain.NewDaifugo(domain.NewTrumpCards(0), players, config)
	for i := 0; i < dg.GetPlayerCnt(); i++ {
		dg.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignSpade, i+1, false))
	}
	return dg
}

func newInternalTestDaifugo() *domain.Daifugo {
	config := domain.DefaultDaifugoConfig()
	players := []*domain.DaifugoPlayer{
		domain.NewDaifugoPlayer(true),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
	}
	return domain.NewDaifugo(domain.NewTrumpCards(config.JokerCount), players, config)
}

func TestDaifugoInteractor_Play_GameEndFlag(t *testing.T) {
	mockOutput := "mock"
	dgpMock := new(presenter.MockDaifugoPresenter)
	dgpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	di := NewDaifugoInteractor(newInternalTestDaifugo(), dgpMock)
	di.Game = buildDaifugoEndedGame()

	// Play the card at index 0 → human finishes → checkGameEnd → gameEndFlag = true.
	di.Play([]int{0})
	assert.True(t, di.Game.GetGameEndFlag())

	// Now call Play again with gameEndFlag == true → early return.
	result := di.Play([]int{})
	assert.Equal(t, mockOutput, result)
}

func TestDaifugoInteractor_Play_NotHumanTurn(t *testing.T) {
	mockOutput := "mock"
	dgpMock := new(presenter.MockDaifugoPresenter)
	dgpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	di := NewDaifugoInteractor(newInternalTestDaifugo(), dgpMock)
	di.Game = buildDaifugoCpuTurnGame()

	// currentTurn = 0, player 0 is CPU → !IsHumanTurn() is true.
	assert.False(t, di.Game.IsHumanTurn())
	result := di.Play([]int{})
	assert.Equal(t, mockOutput, result)
}
