//go:build test

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

func newTestEgyptianRatscrew() *domain.EgyptianRatscrew {
	return domain.NewDefaultEgyptianRatscrew()
}

func TestNewEgyptianRatscrewInteractor_NilGuards(t *testing.T) {
	epMock := new(presenter.MockEgyptianRatscrewPresenter)

	t.Run("panics when e is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "EgyptianRatscrewInteractor: e must not be nil", func() {
			usecase.NewEgyptianRatscrewInteractor(nil, epMock)
		})
	})

	t.Run("panics when ep is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "EgyptianRatscrewInteractor: ep must not be nil", func() {
			usecase.NewEgyptianRatscrewInteractor(newTestEgyptianRatscrew(), nil)
		})
	})
}

func TestEgyptianRatscrewInteractor_Reset(t *testing.T) {
	out := `{"phase":0}`
	epMock := new(presenter.MockEgyptianRatscrewPresenter)
	epMock.On("Output", mock.Anything, mock.Anything).Return(out)
	gameMock := new(interfaces.MockEgyptianRatscrewGame)
	gameMock.On("Reset").Return()

	ei := usecase.NewEgyptianRatscrewInteractor(gameMock, epMock)
	assert.Equal(t, out, ei.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestEgyptianRatscrewInteractor_ResetWithConfig(t *testing.T) {
	out := `{"phase":0}`

	t.Run("valid config", func(t *testing.T) {
		epMock := new(presenter.MockEgyptianRatscrewPresenter)
		epMock.On("Output", mock.Anything, mock.Anything).Return(out)
		gameMock := new(interfaces.MockEgyptianRatscrewGame)
		cfg := domain.EgyptianRatscrewConfig{CpuDifficulty: domain.EgyptianRatscrewCpuHard}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		ei := usecase.NewEgyptianRatscrewInteractor(gameMock, epMock)
		assert.Equal(t, out, ei.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error", func(t *testing.T) {
		errOut := `{"error":"invalid"}`
		epMock := new(presenter.MockEgyptianRatscrewPresenter)
		epMock.On("Output", mock.Anything, mock.MatchedBy(func(err error) bool { return err != nil })).Return(errOut)
		gameMock := new(interfaces.MockEgyptianRatscrewGame)

		ei := usecase.NewEgyptianRatscrewInteractor(gameMock, epMock)
		assert.Equal(t, errOut, ei.ResetWithConfig(domain.EgyptianRatscrewConfig{CpuDifficulty: 99}))
	})
}

func TestEgyptianRatscrewInteractor_Step(t *testing.T) {
	out := `{"phase":0}`

	t.Run("success", func(t *testing.T) {
		epMock := new(presenter.MockEgyptianRatscrewPresenter)
		epMock.On("Output", mock.Anything, nil).Return(out)
		gameMock := new(interfaces.MockEgyptianRatscrewGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("Step").Return(nil)

		ei := usecase.NewEgyptianRatscrewInteractor(gameMock, epMock)
		assert.Equal(t, out, ei.Step())
	})

	t.Run("error from domain", func(t *testing.T) {
		errOut := `{"error":"invalid"}`
		epMock := new(presenter.MockEgyptianRatscrewPresenter)
		epMock.On("Output", mock.Anything, mock.MatchedBy(func(err error) bool { return err != nil })).Return(errOut)
		gameMock := new(interfaces.MockEgyptianRatscrewGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("Step").Return(domain.ErrInvalidPlay)

		ei := usecase.NewEgyptianRatscrewInteractor(gameMock, epMock)
		assert.Equal(t, errOut, ei.Step())
	})

	t.Run("blocked when game ended", func(t *testing.T) {
		endOut := `{"gameEnd":true}`
		epMock := new(presenter.MockEgyptianRatscrewPresenter)
		epMock.On("Output", mock.Anything, mock.Anything).Return(endOut)
		gameMock := new(interfaces.MockEgyptianRatscrewGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ei := usecase.NewEgyptianRatscrewInteractor(gameMock, epMock)
		assert.Equal(t, endOut, ei.Step())
		gameMock.AssertNotCalled(t, "Step")
	})
}

func TestEgyptianRatscrewInteractor_Slap(t *testing.T) {
	out := `{"phase":0}`

	t.Run("success", func(t *testing.T) {
		epMock := new(presenter.MockEgyptianRatscrewPresenter)
		epMock.On("Output", mock.Anything, nil).Return(out)
		gameMock := new(interfaces.MockEgyptianRatscrewGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("Slap", 0).Return(nil)

		ei := usecase.NewEgyptianRatscrewInteractor(gameMock, epMock)
		assert.Equal(t, out, ei.Slap(0))
	})

	t.Run("error from domain", func(t *testing.T) {
		errOut := `{"error":"invalid"}`
		epMock := new(presenter.MockEgyptianRatscrewPresenter)
		epMock.On("Output", mock.Anything, mock.MatchedBy(func(err error) bool { return err != nil })).Return(errOut)
		gameMock := new(interfaces.MockEgyptianRatscrewGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("Slap", 0).Return(domain.ErrInvalidPlay)

		ei := usecase.NewEgyptianRatscrewInteractor(gameMock, epMock)
		assert.Equal(t, errOut, ei.Slap(0))
	})

	t.Run("blocked when game ended", func(t *testing.T) {
		endOut := `{"gameEnd":true}`
		epMock := new(presenter.MockEgyptianRatscrewPresenter)
		epMock.On("Output", mock.Anything, mock.Anything).Return(endOut)
		gameMock := new(interfaces.MockEgyptianRatscrewGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ei := usecase.NewEgyptianRatscrewInteractor(gameMock, epMock)
		assert.Equal(t, endOut, ei.Slap(0))
		gameMock.AssertNotCalled(t, "Slap", mock.Anything)
	})
}

func TestEgyptianRatscrewInteractor_Tick(t *testing.T) {
	out := `{"phase":0}`
	epMock := new(presenter.MockEgyptianRatscrewPresenter)
	epMock.On("Output", mock.Anything, mock.Anything).Return(out)
	gameMock := new(interfaces.MockEgyptianRatscrewGame)
	gameMock.On("Tick").Return(domain.EgyptianRatscrewPendingNone)

	ei := usecase.NewEgyptianRatscrewInteractor(gameMock, epMock)
	assert.Equal(t, out, ei.Tick())
	gameMock.AssertCalled(t, "Tick")
}

func TestEgyptianRatscrewInteractor_GetConfig(t *testing.T) {
	cfg := domain.EgyptianRatscrewConfig{CpuDifficulty: domain.EgyptianRatscrewCpuHard}
	gameMock := new(interfaces.MockEgyptianRatscrewGame)
	gameMock.On("GetConfig").Return(cfg)
	epMock := new(presenter.MockEgyptianRatscrewPresenter)

	ei := usecase.NewEgyptianRatscrewInteractor(gameMock, epMock)
	assert.Equal(t, cfg, ei.GetConfig())
}

func TestEgyptianRatscrewInteractor_ActionLog(t *testing.T) {
	logOut := `{"log":[]}`
	epMock := new(presenter.MockEgyptianRatscrewPresenter)
	epMock.On("ActionLogOutput", mock.Anything).Return(logOut)
	gameMock := new(interfaces.MockEgyptianRatscrewGame)

	ei := usecase.NewEgyptianRatscrewInteractor(gameMock, epMock)
	assert.Equal(t, logOut, ei.ActionLog())
}

func TestEgyptianRatscrewInteractor_SnapshotAndRestore(t *testing.T) {
	game := newTestEgyptianRatscrew()
	game.Reset()
	epMock := new(presenter.MockEgyptianRatscrewPresenter)
	epMock.On("Output", mock.Anything, mock.Anything).Return("{}")

	ei := usecase.NewEgyptianRatscrewInteractor(game, epMock)
	data, err := ei.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreEgyptianRatscrewInteractor(data, epMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestoreEgyptianRatscrewInteractor_InvalidJSON(t *testing.T) {
	epMock := new(presenter.MockEgyptianRatscrewPresenter)
	_, err := usecase.RestoreEgyptianRatscrewInteractor([]byte("not-json"), epMock)
	assert.Error(t, err)
}
