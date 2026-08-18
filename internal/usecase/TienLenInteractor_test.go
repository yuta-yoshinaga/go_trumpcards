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

func TestNewTienLenInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockTienLenPresenter)

	t.Run("panics when tg is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "TienLenInteractor: tg must not be nil", func() {
			usecase.NewTienLenInteractor(nil, pMock)
		})
	})

	t.Run("panics when tlp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockTienLenGame)
		assert.PanicsWithValue(t, "TienLenInteractor: tlp must not be nil", func() {
			usecase.NewTienLenInteractor(gameMock, nil)
		})
	})
}

func TestTienLenInteractor_Reset(t *testing.T) {
	mockOutput := `{"players":[]}`
	pMock := new(presenter.MockTienLenPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTienLenGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTienLenInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ti.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestTienLenInteractor_Reset_RunsCpuTurns(t *testing.T) {
	mockOutput := `{}`
	pMock := new(presenter.MockTienLenPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTienLenGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	// CPU turn first, then the human's turn ends the loop.
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()

	ti := usecase.NewTienLenInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ti.Reset())
	gameMock.AssertCalled(t, "CpuPlay")
}

func TestTienLenInteractor_Play(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid play runs cpu turns", func(t *testing.T) {
		pMock := new(presenter.MockTienLenPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTienLenGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", []int{0}).Return(nil)
		gameMock.On("HasPendingAction").Return(false)

		ti := usecase.NewTienLenInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ti.Play([]int{0}))
		gameMock.AssertCalled(t, "PlayerPlay", []int{0})
	})

	t.Run("play error surfaces", func(t *testing.T) {
		playErr := errors.New("invalid play")
		pMock := new(presenter.MockTienLenPresenter)
		pMock.On("Output", mock.Anything, playErr).Return(mockOutput)
		gameMock := new(interfaces.MockTienLenGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", []int{1}).Return(playErr)

		ti := usecase.NewTienLenInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ti.Play([]int{1}))
	})

	t.Run("game ended blocks play", func(t *testing.T) {
		pMock := new(presenter.MockTienLenPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTienLenGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ti := usecase.NewTienLenInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ti.Play([]int{0}))
		gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
	})
}

func TestTienLenInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockTienLenPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockTienLenGame)
		cfg := domain.TienLenConfig{CpuDifficulty: domain.TienLenDifficultyHard}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)

		ti := usecase.NewTienLenInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ti.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error without resetting", func(t *testing.T) {
		pMock := new(presenter.MockTienLenPresenter)
		gameMock := new(interfaces.MockTienLenGame)
		pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

		ti := usecase.NewTienLenInteractor(gameMock, pMock)
		cfg := domain.TienLenConfig{CpuDifficulty: domain.TienLenCpuDifficulty(-1)}
		assert.Equal(t, "validation error", ti.ResetWithConfig(cfg))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestTienLenInteractor_GetConfigAndLog(t *testing.T) {
	pMock := new(presenter.MockTienLenPresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockTienLenGame)
	cfg := domain.TienLenConfig{CpuDifficulty: domain.TienLenDifficultyHard}
	gameMock.On("GetConfig").Return(cfg)

	ti := usecase.NewTienLenInteractor(gameMock, pMock)
	assert.Equal(t, cfg, ti.GetConfig())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestRestoreTienLenInteractor(t *testing.T) {
	pMock := new(presenter.MockTienLenPresenter)
	g := domain.NewDefaultTienLen()
	data, err := g.MarshalJSON()
	assert.NoError(t, err)

	ti, err := usecase.RestoreTienLenInteractor(data, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, ti)
}

func TestRestoreTienLenInteractor_InvalidJSON(t *testing.T) {
	pMock := new(presenter.MockTienLenPresenter)
	_, err := usecase.RestoreTienLenInteractor([]byte("not json"), pMock)
	assert.Error(t, err)
}

// #5624: Hint はプレゼンターへ素通しするだけだが、その 1 本が繋がっていないと
// CUI の `h` が何も返さない。
func TestTienLenInteractor_Hint(t *testing.T) {
	gMock := new(interfaces.MockTienLenGame)
	pMock := new(presenter.MockTienLenPresenter)
	pMock.On("HintOutput", gMock).Return("hint-output")

	ti := usecase.NewTienLenInteractor(gMock, pMock)
	assert.Equal(t, "hint-output", ti.Hint())
	pMock.AssertCalled(t, "HintOutput", gMock)
}
