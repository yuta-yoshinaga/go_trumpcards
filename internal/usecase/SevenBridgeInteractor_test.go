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

const sbMockOutput = `{"phase":0}`

// setupSevenBridgeMocks returns a common mock pair that stubs runCpuTurns
// to break at the first iteration on a human turn.
func setupSevenBridgeMocks() (*presenter.MockSevenBridgePresenter, *interfaces.MockSevenBridgeGame) {
	pMock := new(presenter.MockSevenBridgePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(sbMockOutput)
	gameMock := new(interfaces.MockSevenBridgeGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SevenBridgePhaseDraw)
	gameMock.On("IsHumanTurn").Return(true)
	return pMock, gameMock
}

func TestNewSevenBridgeInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockSevenBridgePresenter)
	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "SevenBridgeInteractor: g must not be nil", func() {
			usecase.NewSevenBridgeInteractor(nil, pMock)
		})
	})
	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockSevenBridgeGame)
		assert.PanicsWithValue(t, "SevenBridgeInteractor: gp must not be nil", func() {
			usecase.NewSevenBridgeInteractor(gameMock, nil)
		})
	})
}

func TestSevenBridgeInteractor_Reset(t *testing.T) {
	pMock, gameMock := setupSevenBridgeMocks()
	gameMock.On("Reset").Return()

	ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
	assert.Equal(t, sbMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestSevenBridgeInteractor_Hint(t *testing.T) {
	pMock, gameMock := setupSevenBridgeMocks()
	pMock.On("HintOutput", gameMock).Return("hint_output")

	ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
	assert.Equal(t, "hint_output", ci.Hint())
	pMock.AssertCalled(t, "HintOutput", gameMock)
}

func TestSevenBridgeInteractor_ResetWithConfig_Valid(t *testing.T) {
	pMock, gameMock := setupSevenBridgeMocks()
	cfg := domain.SevenBridgeConfig{CpuDifficulty: domain.SevenBridgeCpuDifficultyHard, PointLimit: 200}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()

	ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
	assert.Equal(t, sbMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
	gameMock.AssertCalled(t, "Reset")
}

func TestSevenBridgeInteractor_ResetWithConfig_Invalid(t *testing.T) {
	pMock := new(presenter.MockSevenBridgePresenter)
	gameMock := new(interfaces.MockSevenBridgeGame)
	pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).
		Return("validation error")

	ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
	bad := domain.SevenBridgeConfig{CpuDifficulty: domain.SevenBridgeCpuDifficulty(-1), PointLimit: 100}
	assert.Equal(t, "validation error", ci.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	gameMock.AssertNotCalled(t, "Reset")
}

func TestSevenBridgeInteractor_DrawFromStock(t *testing.T) {
	t.Run("game ended → no-op", func(t *testing.T) {
		pMock := new(presenter.MockSevenBridgePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(sbMockOutput)
		gameMock := new(interfaces.MockSevenBridgeGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.DrawFromStock())
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})

	t.Run("not human turn → no-op", func(t *testing.T) {
		pMock := new(presenter.MockSevenBridgePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(sbMockOutput)
		gameMock := new(interfaces.MockSevenBridgeGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.DrawFromStock())
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})

	t.Run("error propagated", func(t *testing.T) {
		drawErr := errors.New("draw error")
		pMock := new(presenter.MockSevenBridgePresenter)
		pMock.On("Output", mock.Anything, drawErr).Return(sbMockOutput)
		gameMock := new(interfaces.MockSevenBridgeGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromStock").Return(drawErr)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.DrawFromStock())
	})

	t.Run("success runs CPU turns", func(t *testing.T) {
		pMock, gameMock := setupSevenBridgeMocks()
		gameMock.On("PlayerDrawFromStock").Return(nil)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.DrawFromStock())
		gameMock.AssertCalled(t, "PlayerDrawFromStock")
	})
}

func TestSevenBridgeInteractor_ClaimPon(t *testing.T) {
	t.Run("error propagated", func(t *testing.T) {
		e := errors.New("pon error")
		pMock := new(presenter.MockSevenBridgePresenter)
		pMock.On("Output", mock.Anything, e).Return(sbMockOutput)
		gameMock := new(interfaces.MockSevenBridgeGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerClaimPon", []int{0, 1}).Return(e)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.ClaimPon([]int{0, 1}))
	})

	t.Run("success runs CPU turns", func(t *testing.T) {
		pMock, gameMock := setupSevenBridgeMocks()
		gameMock.On("PlayerClaimPon", []int{0, 1}).Return(nil)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.ClaimPon([]int{0, 1}))
	})

	t.Run("game ended → no-op", func(t *testing.T) {
		pMock := new(presenter.MockSevenBridgePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(sbMockOutput)
		gameMock := new(interfaces.MockSevenBridgeGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.ClaimPon([]int{0, 1}))
		gameMock.AssertNotCalled(t, "PlayerClaimPon")
	})
}

func TestSevenBridgeInteractor_ClaimChi(t *testing.T) {
	t.Run("error propagated", func(t *testing.T) {
		e := errors.New("chi error")
		pMock := new(presenter.MockSevenBridgePresenter)
		pMock.On("Output", mock.Anything, e).Return(sbMockOutput)
		gameMock := new(interfaces.MockSevenBridgeGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerClaimChi", []int{0, 1}).Return(e)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.ClaimChi([]int{0, 1}))
	})

	t.Run("success runs CPU turns", func(t *testing.T) {
		pMock, gameMock := setupSevenBridgeMocks()
		gameMock.On("PlayerClaimChi", []int{0, 1}).Return(nil)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.ClaimChi([]int{0, 1}))
	})

	t.Run("not human turn → no-op", func(t *testing.T) {
		pMock := new(presenter.MockSevenBridgePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(sbMockOutput)
		gameMock := new(interfaces.MockSevenBridgeGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.ClaimChi([]int{0, 1}))
		gameMock.AssertNotCalled(t, "PlayerClaimChi")
	})
}

func TestSevenBridgeInteractor_Meld(t *testing.T) {
	t.Run("error propagated", func(t *testing.T) {
		e := errors.New("meld error")
		pMock := new(presenter.MockSevenBridgePresenter)
		pMock.On("Output", mock.Anything, e).Return(sbMockOutput)
		gameMock := new(interfaces.MockSevenBridgeGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerMeld", []int{0, 1, 2}).Return(e)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.Meld([]int{0, 1, 2}))
	})

	t.Run("success runs CPU turns", func(t *testing.T) {
		pMock, gameMock := setupSevenBridgeMocks()
		gameMock.On("PlayerMeld", []int{0, 1, 2}).Return(nil)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.Meld([]int{0, 1, 2}))
	})
}

func TestSevenBridgeInteractor_Layoff(t *testing.T) {
	t.Run("error propagated", func(t *testing.T) {
		e := errors.New("layoff error")
		pMock := new(presenter.MockSevenBridgePresenter)
		pMock.On("Output", mock.Anything, e).Return(sbMockOutput)
		gameMock := new(interfaces.MockSevenBridgeGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerLayoff", 1, 0, 0).Return(e)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.Layoff(1, 0, 0))
	})

	t.Run("success runs CPU turns", func(t *testing.T) {
		pMock, gameMock := setupSevenBridgeMocks()
		gameMock.On("PlayerLayoff", 1, 0, 0).Return(nil)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.Layoff(1, 0, 0))
	})
}

func TestSevenBridgeInteractor_Discard(t *testing.T) {
	t.Run("game ended → no-op", func(t *testing.T) {
		pMock := new(presenter.MockSevenBridgePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(sbMockOutput)
		gameMock := new(interfaces.MockSevenBridgeGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.Discard(0))
		gameMock.AssertNotCalled(t, "PlayerDiscard")
	})

	t.Run("error propagated", func(t *testing.T) {
		e := errors.New("discard error")
		pMock := new(presenter.MockSevenBridgePresenter)
		pMock.On("Output", mock.Anything, e).Return(sbMockOutput)
		gameMock := new(interfaces.MockSevenBridgeGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDiscard", 3).Return(e)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.Discard(3))
	})

	t.Run("success runs CPU turns", func(t *testing.T) {
		pMock, gameMock := setupSevenBridgeMocks()
		gameMock.On("PlayerDiscard", 3).Return(nil)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.Discard(3))
	})
}

func TestSevenBridgeInteractor_NextRound(t *testing.T) {
	t.Run("game ended → no-op", func(t *testing.T) {
		pMock := new(presenter.MockSevenBridgePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(sbMockOutput)
		gameMock := new(interfaces.MockSevenBridgeGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})

	t.Run("valid next round", func(t *testing.T) {
		pMock, gameMock := setupSevenBridgeMocks()
		gameMock.On("NextRound").Return()

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		assert.Equal(t, sbMockOutput, ci.NextRound())
		gameMock.AssertCalled(t, "NextRound")
	})
}

func TestSevenBridgeInteractor_GetConfig(t *testing.T) {
	pMock := new(presenter.MockSevenBridgePresenter)
	gameMock := new(interfaces.MockSevenBridgeGame)
	expected := domain.SevenBridgeConfig{CpuDifficulty: domain.SevenBridgeCpuDifficultyHard, PointLimit: 200}
	gameMock.On("GetConfig").Return(expected)

	ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
	assert.Equal(t, expected, ci.GetConfig())
}

func TestSevenBridgeInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockSevenBridgePresenter)
	gameMock := new(interfaces.MockSevenBridgeGame)
	pMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)

	ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
	assert.Equal(t, `{"entries":[]}`, ci.ActionLog())
	pMock.AssertExpectations(t)
}

func TestSevenBridgeInteractor_RunCpuTurns(t *testing.T) {
	t.Run("stops when game ended", func(t *testing.T) {
		pMock := new(presenter.MockSevenBridgePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(sbMockOutput)
		gameMock := new(interfaces.MockSevenBridgeGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops on RoundEnd phase", func(t *testing.T) {
		pMock := new(presenter.MockSevenBridgePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(sbMockOutput)
		gameMock := new(interfaces.MockSevenBridgeGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.SevenBridgePhaseRoundEnd)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops on GameEnd phase", func(t *testing.T) {
		pMock := new(presenter.MockSevenBridgePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(sbMockOutput)
		gameMock := new(interfaces.MockSevenBridgeGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.SevenBridgePhaseGameEnd)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("CPU plays then stops at human turn", func(t *testing.T) {
		pMock := new(presenter.MockSevenBridgePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(sbMockOutput)
		gameMock := new(interfaces.MockSevenBridgeGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.SevenBridgePhaseDraw)
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return()
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewSevenBridgeInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertCalled(t, "CpuPlay")
	})
}

func TestRestoreSevenBridgeInteractor(t *testing.T) {
	pMock := new(presenter.MockSevenBridgePresenter)
	// Minimal valid JSON for a SevenBridge (empty defaults).
	blob := []byte(`{}`)
	ci, err := usecase.RestoreSevenBridgeInteractor(blob, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}

func TestRestoreSevenBridgeInteractor_Invalid(t *testing.T) {
	pMock := new(presenter.MockSevenBridgePresenter)
	_, err := usecase.RestoreSevenBridgeInteractor([]byte("{"), pMock)
	assert.Error(t, err)
}
