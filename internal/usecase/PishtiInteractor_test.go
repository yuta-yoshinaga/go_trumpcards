package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newTestPishti() *domain.Pishti {
	return domain.NewDefaultPishti()
}

func TestNewPishtiInteractor_NilGuards(t *testing.T) {
	ppMock := new(presenter.MockPishtiPresenter)
	t.Run("panics when pg is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "PishtiInteractor: pg must not be nil", func() {
			usecase.NewPishtiInteractor(nil, ppMock)
		})
	})
	t.Run("panics when pp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "PishtiInteractor: pp must not be nil", func() {
			usecase.NewPishtiInteractor(newTestPishti(), nil)
		})
	})
}

func TestPishtiInteractor_Methods(t *testing.T) {
	mockOutput := `{"players":[]}`
	ppMock := new(presenter.MockPishtiPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	pi := usecase.NewPishtiInteractor(newTestPishti(), ppMock)

	t.Run("Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, pi.Reset())
	})

	t.Run("Play returns output", func(t *testing.T) {
		assert.Equal(t, mockOutput, pi.Play(0))
	})

	t.Run("NextRound returns output", func(t *testing.T) {
		assert.Equal(t, mockOutput, pi.NextRound())
	})

	t.Run("ResetWithConfig valid", func(t *testing.T) {
		cfg := domain.DefaultPishtiConfig()
		cfg.PlayerCnt = 3
		assert.Equal(t, mockOutput, pi.ResetWithConfig(cfg))
		assert.Equal(t, 3, pi.GetConfig().PlayerCnt)
	})

	t.Run("ResetWithConfig invalid returns output", func(t *testing.T) {
		cfg := domain.PishtiConfig{PlayerCnt: 1}
		assert.Equal(t, mockOutput, pi.ResetWithConfig(cfg))
	})
}

func TestPishtiInteractor_ActionLog(t *testing.T) {
	ppMock := new(presenter.MockPishtiPresenter)
	gameMock := new(interfaces.MockPishtiGame)
	ppMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)
	pi := usecase.NewPishtiInteractor(gameMock, ppMock)
	assert.Equal(t, `{"entries":[]}`, pi.ActionLog())
	ppMock.AssertExpectations(t)
}

func TestPishtiInteractor_MockGame(t *testing.T) {
	mockOutput := `{"players":[]}`
	ppMock := new(presenter.MockPishtiPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)

	gameMock := new(interfaces.MockPishtiGame)
	gameMock.On("Reset").Return()
	gameMock.On("NextRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()
	gameMock.On("PlayerPlay", mock.Anything).Return(nil)
	gameMock.On("SetConfig", mock.Anything).Return()
	gameMock.On("GetConfig").Return(domain.DefaultPishtiConfig())

	pi := usecase.NewPishtiInteractor(gameMock, ppMock)

	t.Run("Reset delegates", func(t *testing.T) {
		assert.Equal(t, mockOutput, pi.Reset())
		gameMock.AssertCalled(t, "Reset")
	})

	t.Run("Play delegates", func(t *testing.T) {
		assert.Equal(t, mockOutput, pi.Play(2))
		gameMock.AssertCalled(t, "PlayerPlay", 2)
	})

	t.Run("NextRound delegates", func(t *testing.T) {
		assert.Equal(t, mockOutput, pi.NextRound())
		gameMock.AssertCalled(t, "NextRound")
	})
}

func TestPishtiInteractor_RunsCpuTurns(t *testing.T) {
	mockOutput := `{"players":[]}`
	ppMock := new(presenter.MockPishtiPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)

	gameMock := new(interfaces.MockPishtiGame)
	gameMock.On("Reset").Return()
	// 1 回目: CPU 手番、2 回目以降: 人間手番でループ終了。
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()

	pi := usecase.NewPishtiInteractor(gameMock, ppMock)
	pi.Reset()
	gameMock.AssertCalled(t, "CpuPlay")
}

func TestPishtiInteractor_Snapshot(t *testing.T) {
	ppMock := new(presenter.MockPishtiPresenter)
	pi := usecase.NewPishtiInteractor(newTestPishti(), ppMock)
	data, err := pi.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestorePishtiInteractor(data, ppMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}
