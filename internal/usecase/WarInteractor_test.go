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

func newTestWar() *domain.War {
	players := []*domain.WarPlayer{
		domain.NewWarPlayer(true),
		domain.NewWarPlayer(false),
	}
	return domain.NewWar(domain.NewTrumpCards(0), players, domain.DefaultWarConfig())
}

func TestNewWarInteractor_NilGuards(t *testing.T) {
	wpMock := new(presenter.MockWarPresenter)

	t.Run("panics when w is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "WarInteractor: w must not be nil", func() {
			usecase.NewWarInteractor(nil, wpMock)
		})
	})

	t.Run("panics when wp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "WarInteractor: wp must not be nil", func() {
			usecase.NewWarInteractor(newTestWar(), nil)
		})
	})
}

func TestWarInteractor_Reset(t *testing.T) {
	out := `{"phase":0}`
	wpMock := new(presenter.MockWarPresenter)
	wpMock.On("Output", mock.Anything, mock.Anything).Return(out)
	gameMock := new(interfaces.MockWarGame)
	gameMock.On("Reset").Return()

	wi := usecase.NewWarInteractor(gameMock, wpMock)
	assert.Equal(t, out, wi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestWarInteractor_ResetWithConfig(t *testing.T) {
	out := `{"phase":0}`

	t.Run("valid config", func(t *testing.T) {
		wpMock := new(presenter.MockWarPresenter)
		wpMock.On("Output", mock.Anything, mock.Anything).Return(out)
		gameMock := new(interfaces.MockWarGame)
		cfg := domain.WarConfig{MaxRounds: 200}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		wi := usecase.NewWarInteractor(gameMock, wpMock)
		assert.Equal(t, out, wi.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error", func(t *testing.T) {
		errOut := `{"error":"invalid"}`
		wpMock := new(presenter.MockWarPresenter)
		wpMock.On("Output", mock.Anything, mock.MatchedBy(func(err error) bool { return err != nil })).Return(errOut)
		gameMock := new(interfaces.MockWarGame)

		wi := usecase.NewWarInteractor(gameMock, wpMock)
		assert.Equal(t, errOut, wi.ResetWithConfig(domain.WarConfig{MaxRounds: 0}))
	})
}

func TestWarInteractor_Step(t *testing.T) {
	out := `{"phase":1}`

	t.Run("success", func(t *testing.T) {
		wpMock := new(presenter.MockWarPresenter)
		wpMock.On("Output", mock.Anything, nil).Return(out)
		gameMock := new(interfaces.MockWarGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("Step").Return(nil)

		wi := usecase.NewWarInteractor(gameMock, wpMock)
		assert.Equal(t, out, wi.Step())
		gameMock.AssertCalled(t, "Step")
	})

	t.Run("step error", func(t *testing.T) {
		errOut := `{"error":"wrong phase"}`
		wpMock := new(presenter.MockWarPresenter)
		wpMock.On("Output", mock.Anything, mock.MatchedBy(func(err error) bool { return err != nil })).Return(errOut)
		gameMock := new(interfaces.MockWarGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("Step").Return(domain.ErrWrongPhase)

		wi := usecase.NewWarInteractor(gameMock, wpMock)
		assert.Equal(t, errOut, wi.Step())
	})

	t.Run("game ended blocks step", func(t *testing.T) {
		endOut := `{"gameEnd":true}`
		wpMock := new(presenter.MockWarPresenter)
		wpMock.On("Output", mock.Anything, mock.Anything).Return(endOut)
		gameMock := new(interfaces.MockWarGame)
		gameMock.On("GetGameEndFlag").Return(true)

		wi := usecase.NewWarInteractor(gameMock, wpMock)
		assert.Equal(t, endOut, wi.Step())
		gameMock.AssertNotCalled(t, "Step")
	})
}

func TestWarInteractor_AutoPlay(t *testing.T) {
	out := `{"phase":3,"gameEndFlag":true}`

	t.Run("success", func(t *testing.T) {
		wpMock := new(presenter.MockWarPresenter)
		wpMock.On("Output", mock.Anything, nil).Return(out)
		gameMock := new(interfaces.MockWarGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("AutoPlay").Return(nil)

		wi := usecase.NewWarInteractor(gameMock, wpMock)
		assert.Equal(t, out, wi.AutoPlay())
		gameMock.AssertCalled(t, "AutoPlay")
	})

	t.Run("autoplay error", func(t *testing.T) {
		errOut := `{"error":"wrong phase"}`
		wpMock := new(presenter.MockWarPresenter)
		wpMock.On("Output", mock.Anything, mock.MatchedBy(func(err error) bool { return err != nil })).Return(errOut)
		gameMock := new(interfaces.MockWarGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("AutoPlay").Return(domain.ErrWrongPhase)

		wi := usecase.NewWarInteractor(gameMock, wpMock)
		assert.Equal(t, errOut, wi.AutoPlay())
	})

	t.Run("game ended blocks autoplay", func(t *testing.T) {
		endOut := `{"gameEnd":true}`
		wpMock := new(presenter.MockWarPresenter)
		wpMock.On("Output", mock.Anything, mock.Anything).Return(endOut)
		gameMock := new(interfaces.MockWarGame)
		gameMock.On("GetGameEndFlag").Return(true)

		wi := usecase.NewWarInteractor(gameMock, wpMock)
		assert.Equal(t, endOut, wi.AutoPlay())
		gameMock.AssertNotCalled(t, "AutoPlay")
	})
}

func TestWarInteractor_GetConfig(t *testing.T) {
	cfg := domain.WarConfig{MaxRounds: 300}
	gameMock := new(interfaces.MockWarGame)
	gameMock.On("GetConfig").Return(cfg)
	wpMock := new(presenter.MockWarPresenter)

	wi := usecase.NewWarInteractor(gameMock, wpMock)
	assert.Equal(t, cfg, wi.GetConfig())
}

func TestWarInteractor_ActionLog(t *testing.T) {
	logOut := `{"log":[]}`
	wpMock := new(presenter.MockWarPresenter)
	wpMock.On("ActionLogOutput", mock.Anything).Return(logOut)
	gameMock := new(interfaces.MockWarGame)

	wi := usecase.NewWarInteractor(gameMock, wpMock)
	assert.Equal(t, logOut, wi.ActionLog())
}

func TestWarInteractor_SnapshotAndRestore(t *testing.T) {
	game := newTestWar()
	game.Reset()
	wpMock := new(presenter.MockWarPresenter)
	wpMock.On("Output", mock.Anything, mock.Anything).Return("{}")

	wi := usecase.NewWarInteractor(game, wpMock)
	data, err := wi.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreWarInteractor(data, wpMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestoreWarInteractor_InvalidJSON(t *testing.T) {
	wpMock := new(presenter.MockWarPresenter)
	_, err := usecase.RestoreWarInteractor([]byte("not-json"), wpMock)
	assert.Error(t, err)
}
