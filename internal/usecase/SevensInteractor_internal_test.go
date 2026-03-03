package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// buildSevensEndedGame creates a Sevens game where gameEndFlag is true.
// Human player has 1 playable card. All other players are finished.
// After human plays the card, checkGameEnd sets gameEndFlag.
func buildSevensEndedGame() *domain.Sevens {
	config := domain.SevensConfig{
		TunnelEnabled: false,
		JokerCount:    0,
		CpuStrategy:   false,
	}
	players := []*domain.SevensPlayer{
		domain.NewSevensPlayer(true),  // player 0: human
		domain.NewSevensPlayer(false), // player 1: CPU finished
		domain.NewSevensPlayer(false), // player 2: CPU finished
		domain.NewSevensPlayer(false), // player 3: CPU finished
	}
	s := domain.NewSevens(domain.NewTrumpCards(0), players, config)

	// Mark CPU players as finished.
	for i := 1; i < s.GetPlayerCnt(); i++ {
		s.GetPlayer(i).SetIsFinished(true)
		s.GetPlayer(i).SetRank(i)
	}

	// Give human a single playable card (spade 6, adjacent to 7 on board).
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

	return s
}

// buildSevensCpuTurnGame creates a Sevens game where currentTurn is a CPU.
func buildSevensCpuTurnGame() *domain.Sevens {
	config := domain.SevensConfig{
		TunnelEnabled: false,
		JokerCount:    0,
		CpuStrategy:   false,
	}
	players := []*domain.SevensPlayer{
		domain.NewSevensPlayer(false), // player 0: CPU (current turn)
		domain.NewSevensPlayer(true),  // player 1: human
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
	}
	s := domain.NewSevens(domain.NewTrumpCards(0), players, config)

	// Give all players cards so game doesn't end.
	for i := 0; i < s.GetPlayerCnt(); i++ {
		s.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	}

	return s
}

func newInternalTestSevens() *domain.Sevens {
	config := domain.DefaultSevensConfig()
	players := []*domain.SevensPlayer{
		domain.NewSevensPlayer(true),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
	}
	return domain.NewSevens(domain.NewTrumpCards(config.JokerCount), players, config)
}

func TestSevensInteractor_Play_GameEndFlag(t *testing.T) {
	mockOutput := "mock"
	spMock := new(presenter.MockSevensPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	si := NewSevensInteractor(newInternalTestSevens(), spMock)

	s := buildSevensEndedGame()
	si.s = s

	// Human plays card at index 0 → finishes → checkGameEnd → gameEndFlag = true.
	si.Play(0)
	assert.True(t, si.s.GetGameEndFlag())

	// Now call Play with gameEndFlag == true → early return.
	result := si.Play(0)
	assert.Equal(t, mockOutput, result)
}

func TestSevensInteractor_Play_NotHumanTurn(t *testing.T) {
	mockOutput := "mock"
	spMock := new(presenter.MockSevensPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	si := NewSevensInteractor(newInternalTestSevens(), spMock)

	si.s = buildSevensCpuTurnGame()

	assert.False(t, si.s.IsHumanTurn())
	result := si.Play(0)
	assert.Equal(t, mockOutput, result)
}

func TestSevensInteractor_PlayJoker_GameEndFlag(t *testing.T) {
	mockOutput := "mock"
	spMock := new(presenter.MockSevensPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	si := NewSevensInteractor(newInternalTestSevens(), spMock)

	// Build a game with joker. Human has 1 joker, CPUs finished.
	config := domain.SevensConfig{
		TunnelEnabled: false,
		JokerCount:    1,
		CpuStrategy:   false,
	}
	players := []*domain.SevensPlayer{
		domain.NewSevensPlayer(true),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
	}
	s := domain.NewSevens(domain.NewTrumpCards(1), players, config)
	for i := 1; i < s.GetPlayerCnt(); i++ {
		s.GetPlayer(i).SetIsFinished(true)
		s.GetPlayer(i).SetRank(i)
	}
	// Give human a joker.
	players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
	si.s = s

	// Play joker at a valid position (spade 6, adjacent to 7).
	si.PlayJoker(0, domain.CardDesignSpade, 6)
	assert.True(t, si.s.GetGameEndFlag())

	// Now call PlayJoker with gameEndFlag == true → early return.
	result := si.PlayJoker(0, domain.CardDesignSpade, 5)
	assert.Equal(t, mockOutput, result)
}

func TestSevensInteractor_PlayJoker_SuccessRunCpuTurns(t *testing.T) {
	mockOutput := "mock"
	spMock := new(presenter.MockSevensPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	si := NewSevensInteractor(newInternalTestSevens(), spMock)

	// Build a game where human plays joker successfully and game does NOT end.
	// This covers the runCpuTurns() call inside PlayJoker.
	config := domain.SevensConfig{
		TunnelEnabled: false,
		JokerCount:    1,
		CpuStrategy:   false,
	}
	players := []*domain.SevensPlayer{
		domain.NewSevensPlayer(true),  // player 0: human
		domain.NewSevensPlayer(false), // player 1: CPU
		domain.NewSevensPlayer(false), // player 2: CPU
		domain.NewSevensPlayer(false), // player 3: CPU
	}
	s := domain.NewSevens(domain.NewTrumpCards(1), players, config)

	// Give human a joker and another card so human doesn't finish.
	players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))

	// Give CPU players playable cards.
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	players[2].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))

	si.s = s

	// Human plays joker at spade 6 (adjacent to 7) → success → runCpuTurns.
	result := si.PlayJoker(0, domain.CardDesignSpade, 6)
	assert.Equal(t, mockOutput, result)
	assert.False(t, si.s.GetGameEndFlag())
}

func TestSevensInteractor_PlayJoker_NotHumanTurn(t *testing.T) {
	mockOutput := "mock"
	spMock := new(presenter.MockSevensPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	si := NewSevensInteractor(newInternalTestSevens(), spMock)

	si.s = buildSevensCpuTurnGame()

	assert.False(t, si.s.IsHumanTurn())
	result := si.PlayJoker(0, domain.CardDesignSpade, 6)
	assert.Equal(t, mockOutput, result)
}

func TestSevensInteractor_runCpuTurns_HumanAutoHandleNoOption(t *testing.T) {
	mockOutput := "mock"
	spMock := new(presenter.MockSevensPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	si := NewSevensInteractor(newInternalTestSevens(), spMock)

	// Build a game where the human player has no playable cards and no passes left.
	// This triggers the AutoHandleNoOption branch in runCpuTurns.
	config := domain.SevensConfig{
		TunnelEnabled: false,
		JokerCount:    0,
		CpuStrategy:   false,
	}
	players := []*domain.SevensPlayer{
		domain.NewSevensPlayer(true),  // player 0: human
		domain.NewSevensPlayer(false), // player 1: CPU
		domain.NewSevensPlayer(false), // player 2: CPU
		domain.NewSevensPlayer(false), // player 3: CPU
	}
	s := domain.NewSevens(domain.NewTrumpCards(0), players, config)

	// Exhaust all human passes so CanPass() returns false.
	for players[0].CanPass() {
		players[0].IncrPassesUsed()
	}

	// Give human an unplayable card (e.g. spade 1 when only 7 is on board
	// and 6 is not placed, so 1 is not adjacent to any placed card).
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))

	// Give CPU players playable cards so the game continues past CPU turns.
	// CPU1 gets spade 6 (playable, adjacent to 7).
	players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	// CPU2 gets spade 8 (playable, adjacent to 7).
	players[2].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	// CPU3 gets heart 6 (playable, adjacent to 7).
	players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))

	si.s = s

	// runCpuTurns is called: human is turn 0, has no option → AutoHandleNoOption
	// → human is eliminated → turn advances to CPU → CPUs play → eventually
	// game may end or loop back.
	si.runCpuTurns()

	// Verify human was eliminated (AutoHandleNoOption path was taken).
	assert.True(t, si.s.GetPlayer(0).GetIsEliminated())
	assert.True(t, si.s.GetPlayer(0).GetIsFinished())

	// Verify subsequent CPU turns ran and the game ended.
	// The last remaining player is auto-finished by checkGameEnd (may still hold cards).
	assert.True(t, si.s.GetGameEndFlag())
	for i := 1; i < si.s.GetPlayerCnt(); i++ {
		assert.True(t, si.s.GetPlayer(i).GetIsFinished(), "player %d should be finished", i)
	}
}
