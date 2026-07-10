//go:build test

package usecase_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// newSambaGuardMocks builds a presenter/game pair whose Output always returns
// the sentinel string, so each command's guard/error branch can be asserted.
func newSambaGuardMocks() (*presenter.MockSambaPresenter, *interfaces.MockSambaGame, *usecase.SambaInteractor) {
	pMock := new(presenter.MockSambaPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return("out")
	gameMock := new(interfaces.MockSambaGame)
	ci := usecase.NewSambaInteractor(gameMock, pMock)
	return pMock, gameMock, ci
}

func TestSambaInteractor_GameEndedGuards(t *testing.T) {
	t.Run("DrawFromDiscard blocked when game ended", func(t *testing.T) {
		_, gameMock, ci := newSambaGuardMocks()
		gameMock.On("GetGameEndFlag").Return(true)
		assert.Equal(t, "out", ci.DrawFromDiscard([]int{0, 1}))
		gameMock.AssertNotCalled(t, "PlayerDrawFromDiscard", mock.Anything)
	})

	t.Run("DrawFromDiscard blocked when not human", func(t *testing.T) {
		_, gameMock, ci := newSambaGuardMocks()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)
		assert.Equal(t, "out", ci.DrawFromDiscard([]int{0, 1}))
		gameMock.AssertNotCalled(t, "PlayerDrawFromDiscard", mock.Anything)
	})

	t.Run("Meld blocked when not human", func(t *testing.T) {
		_, gameMock, ci := newSambaGuardMocks()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)
		assert.Equal(t, "out", ci.Meld([][]int{{0, 1, 2}}))
		gameMock.AssertNotCalled(t, "PlayerMeld", mock.Anything)
	})

	t.Run("SkipMeld blocked when game ended", func(t *testing.T) {
		_, gameMock, ci := newSambaGuardMocks()
		gameMock.On("GetGameEndFlag").Return(true)
		assert.Equal(t, "out", ci.SkipMeld())
		gameMock.AssertNotCalled(t, "PlayerSkipMeld")
	})

	t.Run("Discard blocked when game ended", func(t *testing.T) {
		_, gameMock, ci := newSambaGuardMocks()
		gameMock.On("GetGameEndFlag").Return(true)
		assert.Equal(t, "out", ci.Discard(0))
		gameMock.AssertNotCalled(t, "PlayerDiscard", mock.Anything)
	})

	t.Run("GoOut blocked when game ended", func(t *testing.T) {
		_, gameMock, ci := newSambaGuardMocks()
		gameMock.On("GetGameEndFlag").Return(true)
		assert.Equal(t, "out", ci.GoOut())
		gameMock.AssertNotCalled(t, "PlayerGoOut")
	})
}

func TestSambaInteractor_ErrorBranches(t *testing.T) {
	someErr := errors.New("boom")

	t.Run("SkipMeld error is returned", func(t *testing.T) {
		pMock, gameMock, ci := newSambaGuardMocks()
		_ = pMock
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerSkipMeld").Return(someErr)
		assert.Equal(t, "out", ci.SkipMeld())
	})

	t.Run("Discard error is returned", func(t *testing.T) {
		_, gameMock, ci := newSambaGuardMocks()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDiscard", 2).Return(someErr)
		assert.Equal(t, "out", ci.Discard(2))
	})

	t.Run("GoOut error is returned", func(t *testing.T) {
		_, gameMock, ci := newSambaGuardMocks()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerGoOut").Return(someErr)
		assert.Equal(t, "out", ci.GoOut())
	})
}

func TestSambaInteractor_RunCpuTurns_GameEndBreak(t *testing.T) {
	pMock := new(presenter.MockSambaPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return("out")
	gameMock := new(interfaces.MockSambaGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(true) // loop exits immediately

	ci := usecase.NewSambaInteractor(gameMock, pMock)
	assert.Equal(t, "out", ci.Reset())
	gameMock.AssertNotCalled(t, "CpuPlay")
}

func TestSambaInteractor_RunCpuTurns_GameEndPhaseBreak(t *testing.T) {
	pMock := new(presenter.MockSambaPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return("out")
	gameMock := new(interfaces.MockSambaGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SambaPhaseGameEnd)

	ci := usecase.NewSambaInteractor(gameMock, pMock)
	assert.Equal(t, "out", ci.Reset())
	gameMock.AssertNotCalled(t, "CpuPlay")
}
