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

const threeCardBragMockOutput = `{"phase":0}`

func TestNewThreeCardBragInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockThreeCardBragPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "ThreeCardBragInteractor: g must not be nil", func() {
			usecase.NewThreeCardBragInteractor(nil, spMock)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockThreeCardBragGame)
		assert.PanicsWithValue(t, "ThreeCardBragInteractor: sp must not be nil", func() {
			usecase.NewThreeCardBragInteractor(gameMock, nil)
		})
	})
}

func TestThreeCardBragInteractor_Reset(t *testing.T) {
	spMock := new(presenter.MockThreeCardBragPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(threeCardBragMockOutput)
	gameMock := new(interfaces.MockThreeCardBragGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ThreeCardBragPhaseBetting)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewThreeCardBragInteractor(gameMock, spMock)
	assert.Equal(t, threeCardBragMockOutput, ti.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestThreeCardBragInteractor_ResetWithConfig(t *testing.T) {
	spMock := new(presenter.MockThreeCardBragPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(threeCardBragMockOutput)
	gameMock := new(interfaces.MockThreeCardBragGame)
	cfg := domain.ThreeCardBragConfig{
		CpuDifficulty: domain.ThreeCardBragCpuDifficultyHard,
		Ante:          2,
		StartingChips: 50,
	}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ThreeCardBragPhaseBetting)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewThreeCardBragInteractor(gameMock, spMock)
	assert.Equal(t, threeCardBragMockOutput, ti.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestThreeCardBragInteractor_ResetWithConfigInvalid(t *testing.T) {
	spMock := new(presenter.MockThreeCardBragPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(threeCardBragMockOutput)
	gameMock := new(interfaces.MockThreeCardBragGame)

	ti := usecase.NewThreeCardBragInteractor(gameMock, spMock)
	bad := domain.ThreeCardBragConfig{
		CpuDifficulty: domain.ThreeCardBragCpuDifficulty(99),
		Ante:          1,
		StartingChips: 30,
	}
	assert.Equal(t, threeCardBragMockOutput, ti.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestThreeCardBragInteractor_See(t *testing.T) {
	spMock := new(presenter.MockThreeCardBragPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(threeCardBragMockOutput)

	t.Run("see success advances CPU", func(t *testing.T) {
		gameMock := new(interfaces.MockThreeCardBragGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerSee").Return(nil)
		gameMock.On("GetPhase").Return(domain.ThreeCardBragPhaseBetting)
		gameMock.On("IsHumanTurn").Return(true)

		ti := usecase.NewThreeCardBragInteractor(gameMock, spMock)
		assert.Equal(t, threeCardBragMockOutput, ti.See())
		gameMock.AssertCalled(t, "PlayerSee")
	})

	t.Run("see error", func(t *testing.T) {
		gameMock := new(interfaces.MockThreeCardBragGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerSee").Return(errors.New("already seen"))

		ti := usecase.NewThreeCardBragInteractor(gameMock, spMock)
		assert.Equal(t, threeCardBragMockOutput, ti.See())
	})

	t.Run("game ended blocks see", func(t *testing.T) {
		gameMock := new(interfaces.MockThreeCardBragGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ti := usecase.NewThreeCardBragInteractor(gameMock, spMock)
		assert.Equal(t, threeCardBragMockOutput, ti.See())
		gameMock.AssertNotCalled(t, "PlayerSee")
	})
}

func TestThreeCardBragInteractor_Bet(t *testing.T) {
	spMock := new(presenter.MockThreeCardBragPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(threeCardBragMockOutput)

	t.Run("bet success", func(t *testing.T) {
		gameMock := new(interfaces.MockThreeCardBragGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBet").Return(nil)
		gameMock.On("GetPhase").Return(domain.ThreeCardBragPhaseRoundEnd)
		gameMock.On("IsHumanTurn").Return(true)

		ti := usecase.NewThreeCardBragInteractor(gameMock, spMock)
		assert.Equal(t, threeCardBragMockOutput, ti.Bet())
		gameMock.AssertCalled(t, "PlayerBet")
	})

	t.Run("bet error", func(t *testing.T) {
		gameMock := new(interfaces.MockThreeCardBragGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBet").Return(errors.New("not your turn"))

		ti := usecase.NewThreeCardBragInteractor(gameMock, spMock)
		assert.Equal(t, threeCardBragMockOutput, ti.Bet())
	})
}

func TestThreeCardBragInteractor_Raise(t *testing.T) {
	spMock := new(presenter.MockThreeCardBragPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(threeCardBragMockOutput)

	t.Run("raise success", func(t *testing.T) {
		gameMock := new(interfaces.MockThreeCardBragGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerRaise", 4).Return(nil)
		gameMock.On("GetPhase").Return(domain.ThreeCardBragPhaseBetting)
		gameMock.On("IsHumanTurn").Return(true)

		ti := usecase.NewThreeCardBragInteractor(gameMock, spMock)
		assert.Equal(t, threeCardBragMockOutput, ti.Raise(4))
		gameMock.AssertCalled(t, "PlayerRaise", 4)
	})

	t.Run("raise error", func(t *testing.T) {
		gameMock := new(interfaces.MockThreeCardBragGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerRaise", 1).Return(errors.New("raise too small"))

		ti := usecase.NewThreeCardBragInteractor(gameMock, spMock)
		assert.Equal(t, threeCardBragMockOutput, ti.Raise(1))
	})
}

func TestThreeCardBragInteractor_FoldShow(t *testing.T) {
	spMock := new(presenter.MockThreeCardBragPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(threeCardBragMockOutput)

	t.Run("fold success", func(t *testing.T) {
		gameMock := new(interfaces.MockThreeCardBragGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerFold").Return(nil)
		gameMock.On("GetPhase").Return(domain.ThreeCardBragPhaseRoundEnd)
		gameMock.On("IsHumanTurn").Return(true)

		ti := usecase.NewThreeCardBragInteractor(gameMock, spMock)
		assert.Equal(t, threeCardBragMockOutput, ti.Fold())
		gameMock.AssertCalled(t, "PlayerFold")
	})

	t.Run("show success", func(t *testing.T) {
		gameMock := new(interfaces.MockThreeCardBragGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerShow").Return(nil)
		gameMock.On("GetPhase").Return(domain.ThreeCardBragPhaseRoundEnd)
		gameMock.On("IsHumanTurn").Return(true)

		ti := usecase.NewThreeCardBragInteractor(gameMock, spMock)
		assert.Equal(t, threeCardBragMockOutput, ti.Show())
		gameMock.AssertCalled(t, "PlayerShow")
	})

	t.Run("show error", func(t *testing.T) {
		gameMock := new(interfaces.MockThreeCardBragGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerShow").Return(errors.New("cannot show"))

		ti := usecase.NewThreeCardBragInteractor(gameMock, spMock)
		assert.Equal(t, threeCardBragMockOutput, ti.Show())
	})
}

func TestThreeCardBragInteractor_NextRound(t *testing.T) {
	spMock := new(presenter.MockThreeCardBragPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(threeCardBragMockOutput)
	gameMock := new(interfaces.MockThreeCardBragGame)
	gameMock.On("NextRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ThreeCardBragPhaseBetting)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewThreeCardBragInteractor(gameMock, spMock)
	assert.Equal(t, threeCardBragMockOutput, ti.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestThreeCardBragInteractor_RunCpuTurns(t *testing.T) {
	spMock := new(presenter.MockThreeCardBragPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(threeCardBragMockOutput)
	gameMock := new(interfaces.MockThreeCardBragGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	// First a CPU turn (betting + not human), then it becomes the human's turn.
	gameMock.On("GetPhase").Return(domain.ThreeCardBragPhaseBetting)
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuAct").Return()

	ti := usecase.NewThreeCardBragInteractor(gameMock, spMock)
	assert.Equal(t, threeCardBragMockOutput, ti.Reset())
	gameMock.AssertCalled(t, "CpuAct")
}

func TestThreeCardBragInteractor_GetConfigHintActionLog(t *testing.T) {
	spMock := new(presenter.MockThreeCardBragPresenter)
	spMock.On("HintOutput", mock.Anything).Return("hint")
	spMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockThreeCardBragGame)
	cfg := domain.DefaultThreeCardBragConfig()
	gameMock.On("GetConfig").Return(cfg)

	ti := usecase.NewThreeCardBragInteractor(gameMock, spMock)
	assert.Equal(t, cfg, ti.GetConfig())
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestRestoreThreeCardBragInteractor(t *testing.T) {
	spMock := new(presenter.MockThreeCardBragPresenter)
	src := domain.NewDefaultThreeCardBrag()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ti, err := usecase.RestoreThreeCardBragInteractor(data, spMock)
	assert.NoError(t, err)
	assert.NotNil(t, ti)

	_, err = usecase.RestoreThreeCardBragInteractor([]byte(`{`), spMock)
	assert.Error(t, err)
}
