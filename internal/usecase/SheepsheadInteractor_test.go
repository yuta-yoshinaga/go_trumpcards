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

const shMockOutput = `{"phase":0}`

func TestNewSheepsheadInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockSheepsheadPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "SheepsheadInteractor: g must not be nil", func() {
			usecase.NewSheepsheadInteractor(nil, spMock)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockSheepsheadGame)
		assert.PanicsWithValue(t, "SheepsheadInteractor: sp must not be nil", func() {
			usecase.NewSheepsheadInteractor(gameMock, nil)
		})
	})
}

func TestSheepsheadInteractor_Reset(t *testing.T) {
	spMock := new(presenter.MockSheepsheadPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(shMockOutput)
	gameMock := new(interfaces.MockSheepsheadGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SheepsheadPhasePick)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSheepsheadInteractor(gameMock, spMock)
	assert.Equal(t, shMockOutput, si.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestSheepsheadInteractor_ResetWithConfig(t *testing.T) {
	spMock := new(presenter.MockSheepsheadPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(shMockOutput)
	gameMock := new(interfaces.MockSheepsheadGame)
	cfg := domain.SheepsheadConfig{
		CpuDifficulty: domain.SheepsheadCpuDifficultyHard,
		BaseChips:     2,
		StartChips:    20,
		TargetChips:   60,
	}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SheepsheadPhasePick)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSheepsheadInteractor(gameMock, spMock)
	assert.Equal(t, shMockOutput, si.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestSheepsheadInteractor_ResetWithConfigInvalid(t *testing.T) {
	spMock := new(presenter.MockSheepsheadPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(shMockOutput)
	gameMock := new(interfaces.MockSheepsheadGame)

	si := usecase.NewSheepsheadInteractor(gameMock, spMock)
	// TargetChips must be > StartChips; StartChips=10, TargetChips=5 is invalid
	bad := domain.SheepsheadConfig{
		CpuDifficulty: domain.SheepsheadCpuDifficultyNormal,
		BaseChips:     2,
		StartChips:    10,
		TargetChips:   5,
	}
	assert.Equal(t, shMockOutput, si.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestSheepsheadInteractor_Pick(t *testing.T) {
	spMock := new(presenter.MockSheepsheadPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(shMockOutput)

	t.Run("pick true advances CPU", func(t *testing.T) {
		gameMock := new(interfaces.MockSheepsheadGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerPick", true).Return(nil)
		gameMock.On("GetPhase").Return(domain.SheepsheadPhaseBury)
		gameMock.On("IsHumanTurn").Return(true)

		si := usecase.NewSheepsheadInteractor(gameMock, spMock)
		assert.Equal(t, shMockOutput, si.Pick(true))
		gameMock.AssertCalled(t, "PlayerPick", true)
	})

	t.Run("pick returns error", func(t *testing.T) {
		gameMock := new(interfaces.MockSheepsheadGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerPick", false).Return(errors.New("cannot pass"))

		si := usecase.NewSheepsheadInteractor(gameMock, spMock)
		assert.Equal(t, shMockOutput, si.Pick(false))
	})

	t.Run("game ended blocks pick", func(t *testing.T) {
		gameMock := new(interfaces.MockSheepsheadGame)
		gameMock.On("GetGameEndFlag").Return(true)

		si := usecase.NewSheepsheadInteractor(gameMock, spMock)
		assert.Equal(t, shMockOutput, si.Pick(true))
		gameMock.AssertNotCalled(t, "PlayerPick", mock.Anything)
	})
}

func TestSheepsheadInteractor_Bury(t *testing.T) {
	spMock := new(presenter.MockSheepsheadPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(shMockOutput)
	indices := []int{0, 1}

	t.Run("bury success", func(t *testing.T) {
		gameMock := new(interfaces.MockSheepsheadGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBury", indices).Return(nil)
		gameMock.On("GetPhase").Return(domain.SheepsheadPhaseCall)
		gameMock.On("IsHumanTurn").Return(true)

		si := usecase.NewSheepsheadInteractor(gameMock, spMock)
		assert.Equal(t, shMockOutput, si.Bury(indices))
		gameMock.AssertCalled(t, "PlayerBury", indices)
	})

	t.Run("bury error", func(t *testing.T) {
		gameMock := new(interfaces.MockSheepsheadGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBury", indices).Return(errors.New("invalid bury"))

		si := usecase.NewSheepsheadInteractor(gameMock, spMock)
		assert.Equal(t, shMockOutput, si.Bury(indices))
	})

	t.Run("game ended blocks bury", func(t *testing.T) {
		gameMock := new(interfaces.MockSheepsheadGame)
		gameMock.On("GetGameEndFlag").Return(true)

		si := usecase.NewSheepsheadInteractor(gameMock, spMock)
		assert.Equal(t, shMockOutput, si.Bury(indices))
		gameMock.AssertNotCalled(t, "PlayerBury", mock.Anything)
	})
}

func TestSheepsheadInteractor_Call(t *testing.T) {
	spMock := new(presenter.MockSheepsheadPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(shMockOutput)

	t.Run("call success", func(t *testing.T) {
		gameMock := new(interfaces.MockSheepsheadGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerCall", domain.CardDesignClover).Return(nil)
		gameMock.On("GetPhase").Return(domain.SheepsheadPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		si := usecase.NewSheepsheadInteractor(gameMock, spMock)
		assert.Equal(t, shMockOutput, si.Call(domain.CardDesignClover))
		gameMock.AssertCalled(t, "PlayerCall", domain.CardDesignClover)
	})

	t.Run("call error", func(t *testing.T) {
		gameMock := new(interfaces.MockSheepsheadGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerCall", 0).Return(errors.New("invalid suit"))

		si := usecase.NewSheepsheadInteractor(gameMock, spMock)
		assert.Equal(t, shMockOutput, si.Call(0))
	})

	t.Run("game ended blocks call", func(t *testing.T) {
		gameMock := new(interfaces.MockSheepsheadGame)
		gameMock.On("GetGameEndFlag").Return(true)

		si := usecase.NewSheepsheadInteractor(gameMock, spMock)
		assert.Equal(t, shMockOutput, si.Call(domain.CardDesignSpade))
		gameMock.AssertNotCalled(t, "PlayerCall", mock.Anything)
	})
}

func TestSheepsheadInteractor_PlayResolvesTrick(t *testing.T) {
	spMock := new(presenter.MockSheepsheadPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(shMockOutput)
	gameMock := new(interfaces.MockSheepsheadGame)
	gameMock.On("GetGameEndFlag").Return(false)
	// Phase sequence:
	//   1st call: PhasePlay  (guard check in Play())
	//   2nd call: TrickEnd   (post-PlayerPlay check → triggers ResolveTrick + NextTrick)
	//   3rd call: TrickEnd   (post-ResolveTrick check inside Play(), resolves to RoundEnd path... use RoundEnd to exit)
	//   Remaining: RoundEnd  (runCpuTurns exits on RoundEnd)
	gameMock.On("GetPhase").Return(domain.SheepsheadPhasePlay).Once()
	gameMock.On("GetPhase").Return(domain.SheepsheadPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.SheepsheadPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	si := usecase.NewSheepsheadInteractor(gameMock, spMock)
	assert.Equal(t, shMockOutput, si.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestSheepsheadInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	spMock := new(presenter.MockSheepsheadPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(shMockOutput)
	gameMock := new(interfaces.MockSheepsheadGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SheepsheadPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	si := usecase.NewSheepsheadInteractor(gameMock, spMock)
	assert.Equal(t, shMockOutput, si.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestSheepsheadInteractor_PlayError(t *testing.T) {
	spMock := new(presenter.MockSheepsheadPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(shMockOutput)
	gameMock := new(interfaces.MockSheepsheadGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SheepsheadPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("boom"))

	si := usecase.NewSheepsheadInteractor(gameMock, spMock)
	assert.Equal(t, shMockOutput, si.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestSheepsheadInteractor_PlayNotHumanTurn(t *testing.T) {
	spMock := new(presenter.MockSheepsheadPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(shMockOutput)
	gameMock := new(interfaces.MockSheepsheadGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SheepsheadPhasePlay)
	gameMock.On("IsHumanTurn").Return(false)

	si := usecase.NewSheepsheadInteractor(gameMock, spMock)
	assert.Equal(t, shMockOutput, si.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestSheepsheadInteractor_NextTrick(t *testing.T) {
	spMock := new(presenter.MockSheepsheadPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(shMockOutput)
	gameMock := new(interfaces.MockSheepsheadGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SheepsheadPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSheepsheadInteractor(gameMock, spMock)
	assert.Equal(t, shMockOutput, si.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestSheepsheadInteractor_NextRound(t *testing.T) {
	spMock := new(presenter.MockSheepsheadPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(shMockOutput)
	gameMock := new(interfaces.MockSheepsheadGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.SheepsheadPhasePick)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSheepsheadInteractor(gameMock, spMock)
	assert.Equal(t, shMockOutput, si.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestSheepsheadInteractor_NextRoundGameEnded(t *testing.T) {
	spMock := new(presenter.MockSheepsheadPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(shMockOutput)
	gameMock := new(interfaces.MockSheepsheadGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	si := usecase.NewSheepsheadInteractor(gameMock, spMock)
	assert.Equal(t, shMockOutput, si.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestSheepsheadInteractor_GetConfigHintActionLog(t *testing.T) {
	spMock := new(presenter.MockSheepsheadPresenter)
	spMock.On("HintOutput", mock.Anything).Return("hint")
	spMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockSheepsheadGame)
	cfg := domain.DefaultSheepsheadConfig()
	gameMock.On("GetConfig").Return(cfg)

	si := usecase.NewSheepsheadInteractor(gameMock, spMock)
	assert.Equal(t, cfg, si.GetConfig())
	assert.Equal(t, "hint", si.Hint())
	assert.Equal(t, "log", si.ActionLog())
}

func TestRestoreSheepsheadInteractor(t *testing.T) {
	spMock := new(presenter.MockSheepsheadPresenter)
	src := domain.NewDefaultSheepshead()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	si, err := usecase.RestoreSheepsheadInteractor(data, spMock)
	assert.NoError(t, err)
	assert.NotNil(t, si)

	_, err = usecase.RestoreSheepsheadInteractor([]byte(`{`), spMock)
	assert.Error(t, err)
}
