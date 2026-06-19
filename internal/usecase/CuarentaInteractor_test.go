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

func newTestCuarenta() *domain.Cuarenta {
	return domain.NewDefaultCuarenta()
}

func TestNewCuarentaInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockCuarentaPresenter)
	t.Run("panics when cg is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "CuarentaInteractor: cg must not be nil", func() {
			usecase.NewCuarentaInteractor(nil, spMock)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "CuarentaInteractor: sp must not be nil", func() {
			usecase.NewCuarentaInteractor(newTestCuarenta(), nil)
		})
	})
}

func TestCuarentaInteractor_Methods(t *testing.T) {
	mockOutput := `{"players":[]}`
	spMock := new(presenter.MockCuarentaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	ci := usecase.NewCuarentaInteractor(newTestCuarenta(), spMock)

	t.Run("Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, ci.Reset())
	})
	t.Run("Play returns output", func(t *testing.T) {
		assert.Equal(t, mockOutput, ci.Play(0))
	})
	t.Run("ResetWithConfig valid", func(t *testing.T) {
		cfg := domain.DefaultCuarentaConfig()
		assert.Equal(t, mockOutput, ci.ResetWithConfig(cfg))
	})
	t.Run("GetConfig returns config", func(t *testing.T) {
		cfg := ci.GetConfig()
		assert.Equal(t, domain.CuarentaDefaultTargetScore, cfg.TargetScore)
	})
}

func TestCuarentaInteractor_ActionLog(t *testing.T) {
	spMock := new(presenter.MockCuarentaPresenter)
	gameMock := new(interfaces.MockCuarentaGame)
	spMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)
	ci := usecase.NewCuarentaInteractor(gameMock, spMock)
	assert.Equal(t, `{"entries":[]}`, ci.ActionLog())
	spMock.AssertExpectations(t)
}

func TestCuarentaInteractor_MockGame(t *testing.T) {
	mockOutput := `{"players":[]}`
	spMock := new(presenter.MockCuarentaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)

	gameMock := new(interfaces.MockCuarentaGame)
	gameMock.On("Reset").Return()
	gameMock.On("NextRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetPhase").Return(int(domain.CuarentaPhasePlay))
	gameMock.On("CpuPlay").Return()
	gameMock.On("PlayerPlay", mock.Anything).Return(nil)
	gameMock.On("SetConfig", mock.Anything).Return()
	gameMock.On("GetConfig").Return(domain.DefaultCuarentaConfig())

	ci := usecase.NewCuarentaInteractor(gameMock, spMock)

	t.Run("Reset delegates", func(t *testing.T) {
		assert.Equal(t, mockOutput, ci.Reset())
		gameMock.AssertCalled(t, "Reset")
	})
	t.Run("Play delegates", func(t *testing.T) {
		assert.Equal(t, mockOutput, ci.Play(2))
		gameMock.AssertCalled(t, "PlayerPlay", 2)
	})
	t.Run("NextRound delegates", func(t *testing.T) {
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertCalled(t, "NextRound")
	})
}

func TestCuarentaInteractor_NextRoundSkipsWhenGameEnded(t *testing.T) {
	spMock := new(presenter.MockCuarentaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("done")
	gameMock := new(interfaces.MockCuarentaGame)
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewCuarentaInteractor(gameMock, spMock)
	assert.Equal(t, "done", ci.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestCuarentaInteractor_PlayNotHumanTurn(t *testing.T) {
	spMock := new(presenter.MockCuarentaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("cpu")
	gameMock := new(interfaces.MockCuarentaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	ci := usecase.NewCuarentaInteractor(gameMock, spMock)
	assert.Equal(t, "cpu", ci.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestCuarentaInteractor_Snapshot(t *testing.T) {
	spMock := new(presenter.MockCuarentaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("x")
	ci := usecase.NewCuarentaInteractor(newTestCuarenta(), spMock)
	raw, err := ci.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, raw)
}

func TestRestoreCuarentaInteractor(t *testing.T) {
	spMock := new(presenter.MockCuarentaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("x")
	game := newTestCuarenta()
	game.Reset()
	ci := usecase.NewCuarentaInteractor(game, spMock)

	raw, err := ci.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreCuarentaInteractor(raw, spMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
	assert.Equal(t, ci.GetConfig(), restored.GetConfig())
}

func TestRestoreCuarentaInteractor_RejectsBadState(t *testing.T) {
	spMock := new(presenter.MockCuarentaPresenter)
	raw := []byte(`{"pl":[],"cf":{"ts":40,"di":1},"ph":0,"ct":0}`)
	restored, err := usecase.RestoreCuarentaInteractor(raw, spMock)
	assert.Error(t, err)
	assert.Nil(t, restored)
}
