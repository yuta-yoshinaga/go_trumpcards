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

const teenPattiMockOutput = `{"phase":0}`

func TestNewTeenPattiInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockTeenPattiPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "TeenPattiInteractor: g must not be nil", func() {
			usecase.NewTeenPattiInteractor(nil, spMock)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockTeenPattiGame)
		assert.PanicsWithValue(t, "TeenPattiInteractor: sp must not be nil", func() {
			usecase.NewTeenPattiInteractor(gameMock, nil)
		})
	})
}

func TestTeenPattiInteractor_Reset(t *testing.T) {
	spMock := new(presenter.MockTeenPattiPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(teenPattiMockOutput)
	gameMock := new(interfaces.MockTeenPattiGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TeenPattiPhaseBetting)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
	assert.Equal(t, teenPattiMockOutput, ti.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestTeenPattiInteractor_ResetWithConfig(t *testing.T) {
	spMock := new(presenter.MockTeenPattiPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(teenPattiMockOutput)
	gameMock := new(interfaces.MockTeenPattiGame)
	cfg := domain.TeenPattiConfig{
		CpuDifficulty: domain.TeenPattiCpuDifficultyHard,
		Ante:          2,
		StartingChips: 50,
	}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TeenPattiPhaseBetting)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
	assert.Equal(t, teenPattiMockOutput, ti.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestTeenPattiInteractor_ResetWithConfigInvalid(t *testing.T) {
	spMock := new(presenter.MockTeenPattiPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(teenPattiMockOutput)
	gameMock := new(interfaces.MockTeenPattiGame)

	ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
	bad := domain.TeenPattiConfig{
		CpuDifficulty: domain.TeenPattiCpuDifficulty(99),
		Ante:          1,
		StartingChips: 30,
	}
	assert.Equal(t, teenPattiMockOutput, ti.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestTeenPattiInteractor_See(t *testing.T) {
	spMock := new(presenter.MockTeenPattiPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(teenPattiMockOutput)

	t.Run("see success advances CPU", func(t *testing.T) {
		gameMock := new(interfaces.MockTeenPattiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerSee").Return(nil)
		gameMock.On("GetPhase").Return(domain.TeenPattiPhaseBetting)
		gameMock.On("IsHumanTurn").Return(true)

		ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
		assert.Equal(t, teenPattiMockOutput, ti.See())
		gameMock.AssertCalled(t, "PlayerSee")
	})

	t.Run("see error", func(t *testing.T) {
		gameMock := new(interfaces.MockTeenPattiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerSee").Return(errors.New("already seen"))

		ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
		assert.Equal(t, teenPattiMockOutput, ti.See())
	})

	t.Run("game ended blocks see", func(t *testing.T) {
		gameMock := new(interfaces.MockTeenPattiGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
		assert.Equal(t, teenPattiMockOutput, ti.See())
		gameMock.AssertNotCalled(t, "PlayerSee")
	})
}

func TestTeenPattiInteractor_Bet(t *testing.T) {
	spMock := new(presenter.MockTeenPattiPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(teenPattiMockOutput)

	t.Run("bet success", func(t *testing.T) {
		gameMock := new(interfaces.MockTeenPattiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBet").Return(nil)
		gameMock.On("GetPhase").Return(domain.TeenPattiPhaseRoundEnd)
		gameMock.On("IsHumanTurn").Return(true)

		ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
		assert.Equal(t, teenPattiMockOutput, ti.Bet())
		gameMock.AssertCalled(t, "PlayerBet")
	})

	t.Run("bet error", func(t *testing.T) {
		gameMock := new(interfaces.MockTeenPattiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBet").Return(errors.New("not your turn"))

		ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
		assert.Equal(t, teenPattiMockOutput, ti.Bet())
	})
}

func TestTeenPattiInteractor_Raise(t *testing.T) {
	spMock := new(presenter.MockTeenPattiPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(teenPattiMockOutput)

	t.Run("raise success", func(t *testing.T) {
		gameMock := new(interfaces.MockTeenPattiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerRaise", 4).Return(nil)
		gameMock.On("GetPhase").Return(domain.TeenPattiPhaseBetting)
		gameMock.On("IsHumanTurn").Return(true)

		ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
		assert.Equal(t, teenPattiMockOutput, ti.Raise(4))
		gameMock.AssertCalled(t, "PlayerRaise", 4)
	})

	t.Run("raise error", func(t *testing.T) {
		gameMock := new(interfaces.MockTeenPattiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerRaise", 1).Return(errors.New("raise too small"))

		ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
		assert.Equal(t, teenPattiMockOutput, ti.Raise(1))
	})
}

func TestTeenPattiInteractor_FoldShow(t *testing.T) {
	spMock := new(presenter.MockTeenPattiPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(teenPattiMockOutput)

	t.Run("fold success", func(t *testing.T) {
		gameMock := new(interfaces.MockTeenPattiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerFold").Return(nil)
		gameMock.On("GetPhase").Return(domain.TeenPattiPhaseRoundEnd)
		gameMock.On("IsHumanTurn").Return(true)

		ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
		assert.Equal(t, teenPattiMockOutput, ti.Fold())
		gameMock.AssertCalled(t, "PlayerFold")
	})

	t.Run("show success", func(t *testing.T) {
		gameMock := new(interfaces.MockTeenPattiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerShow").Return(nil)
		gameMock.On("GetPhase").Return(domain.TeenPattiPhaseRoundEnd)
		gameMock.On("IsHumanTurn").Return(true)

		ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
		assert.Equal(t, teenPattiMockOutput, ti.Show())
		gameMock.AssertCalled(t, "PlayerShow")
	})

	t.Run("show error", func(t *testing.T) {
		gameMock := new(interfaces.MockTeenPattiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerShow").Return(errors.New("cannot show"))

		ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
		assert.Equal(t, teenPattiMockOutput, ti.Show())
	})
}

func TestTeenPattiInteractor_SideShow(t *testing.T) {
	spMock := new(presenter.MockTeenPattiPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(teenPattiMockOutput)

	t.Run("request side show success", func(t *testing.T) {
		gameMock := new(interfaces.MockTeenPattiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerRequestSideShow").Return(nil)
		gameMock.On("GetPhase").Return(domain.TeenPattiPhaseSideShow)
		gameMock.On("IsHumanTurn").Return(true)

		ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
		assert.Equal(t, teenPattiMockOutput, ti.RequestSideShow())
		gameMock.AssertCalled(t, "PlayerRequestSideShow")
	})

	t.Run("request side show error", func(t *testing.T) {
		gameMock := new(interfaces.MockTeenPattiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerRequestSideShow").Return(errors.New("cannot side show"))

		ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
		assert.Equal(t, teenPattiMockOutput, ti.RequestSideShow())
	})

	t.Run("respond accept advances CPU in side show", func(t *testing.T) {
		gameMock := new(interfaces.MockTeenPattiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerRespondSideShow", true).Return(nil)
		gameMock.On("GetPhase").Return(domain.TeenPattiPhaseBetting)
		gameMock.On("IsHumanTurn").Return(true)

		ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
		assert.Equal(t, teenPattiMockOutput, ti.RespondSideShow(true))
		gameMock.AssertCalled(t, "PlayerRespondSideShow", true)
	})

	t.Run("respond decline", func(t *testing.T) {
		gameMock := new(interfaces.MockTeenPattiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerRespondSideShow", false).Return(nil)
		gameMock.On("GetPhase").Return(domain.TeenPattiPhaseBetting)
		gameMock.On("IsHumanTurn").Return(true)

		ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
		assert.Equal(t, teenPattiMockOutput, ti.RespondSideShow(false))
		gameMock.AssertCalled(t, "PlayerRespondSideShow", false)
	})
}

func TestTeenPattiInteractor_NextRound(t *testing.T) {
	spMock := new(presenter.MockTeenPattiPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(teenPattiMockOutput)
	gameMock := new(interfaces.MockTeenPattiGame)
	gameMock.On("NextRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TeenPattiPhaseBetting)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
	assert.Equal(t, teenPattiMockOutput, ti.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestTeenPattiInteractor_RunCpuTurns(t *testing.T) {
	spMock := new(presenter.MockTeenPattiPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(teenPattiMockOutput)
	gameMock := new(interfaces.MockTeenPattiGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	// First a CPU turn in the side-show phase, then it becomes the human's turn.
	gameMock.On("GetPhase").Return(domain.TeenPattiPhaseSideShow)
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuAct").Return()

	ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
	assert.Equal(t, teenPattiMockOutput, ti.Reset())
	gameMock.AssertCalled(t, "CpuAct")
}

func TestTeenPattiInteractor_GetConfigHintActionLog(t *testing.T) {
	spMock := new(presenter.MockTeenPattiPresenter)
	spMock.On("HintOutput", mock.Anything).Return("hint")
	spMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockTeenPattiGame)
	cfg := domain.DefaultTeenPattiConfig()
	gameMock.On("GetConfig").Return(cfg)

	ti := usecase.NewTeenPattiInteractor(gameMock, spMock)
	assert.Equal(t, cfg, ti.GetConfig())
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestRestoreTeenPattiInteractor(t *testing.T) {
	spMock := new(presenter.MockTeenPattiPresenter)
	src := domain.NewDefaultTeenPatti()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ti, err := usecase.RestoreTeenPattiInteractor(data, spMock)
	assert.NoError(t, err)
	assert.NotNil(t, ti)

	_, err = usecase.RestoreTeenPattiInteractor([]byte(`{`), spMock)
	assert.Error(t, err)
}
