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

func newTestBeggarMyNeighbour() *domain.BeggarMyNeighbour {
	players := []*domain.BeggarMyNeighbourPlayer{
		domain.NewBeggarMyNeighbourPlayer(true),
		domain.NewBeggarMyNeighbourPlayer(false),
	}
	return domain.NewBeggarMyNeighbour(domain.NewTrumpCards(0), players, domain.DefaultBeggarMyNeighbourConfig())
}

func TestNewBeggarMyNeighbourInteractor_NilGuards(t *testing.T) {
	wpMock := new(presenter.MockBeggarMyNeighbourPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BeggarMyNeighbourInteractor: g must not be nil", func() {
			usecase.NewBeggarMyNeighbourInteractor(nil, wpMock)
		})
	})

	t.Run("panics when wp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BeggarMyNeighbourInteractor: wp must not be nil", func() {
			usecase.NewBeggarMyNeighbourInteractor(newTestBeggarMyNeighbour(), nil)
		})
	})
}

func TestBeggarMyNeighbourInteractor_Reset(t *testing.T) {
	out := `{"phase":0}`
	wpMock := new(presenter.MockBeggarMyNeighbourPresenter)
	wpMock.On("Output", mock.Anything, mock.Anything).Return(out)
	gameMock := new(interfaces.MockBeggarMyNeighbourGame)
	gameMock.On("Reset").Return()

	bi := usecase.NewBeggarMyNeighbourInteractor(gameMock, wpMock)
	assert.Equal(t, out, bi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestBeggarMyNeighbourInteractor_ResetWithConfig(t *testing.T) {
	out := `{"phase":0}`

	t.Run("valid config", func(t *testing.T) {
		wpMock := new(presenter.MockBeggarMyNeighbourPresenter)
		wpMock.On("Output", mock.Anything, mock.Anything).Return(out)
		gameMock := new(interfaces.MockBeggarMyNeighbourGame)
		cfg := domain.BeggarMyNeighbourConfig{MaxRounds: 500}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		bi := usecase.NewBeggarMyNeighbourInteractor(gameMock, wpMock)
		assert.Equal(t, out, bi.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error", func(t *testing.T) {
		errOut := `{"error":"invalid"}`
		wpMock := new(presenter.MockBeggarMyNeighbourPresenter)
		wpMock.On("Output", mock.Anything, mock.MatchedBy(func(err error) bool { return err != nil })).Return(errOut)
		gameMock := new(interfaces.MockBeggarMyNeighbourGame)

		bi := usecase.NewBeggarMyNeighbourInteractor(gameMock, wpMock)
		assert.Equal(t, errOut, bi.ResetWithConfig(domain.BeggarMyNeighbourConfig{MaxRounds: 0}))
	})
}

func TestBeggarMyNeighbourInteractor_Step(t *testing.T) {
	out := `{"phase":1}`

	t.Run("success", func(t *testing.T) {
		wpMock := new(presenter.MockBeggarMyNeighbourPresenter)
		wpMock.On("Output", mock.Anything, nil).Return(out)
		gameMock := new(interfaces.MockBeggarMyNeighbourGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("Step").Return(nil)

		bi := usecase.NewBeggarMyNeighbourInteractor(gameMock, wpMock)
		assert.Equal(t, out, bi.Step())
		gameMock.AssertCalled(t, "Step")
	})

	t.Run("step error", func(t *testing.T) {
		errOut := `{"error":"wrong phase"}`
		wpMock := new(presenter.MockBeggarMyNeighbourPresenter)
		wpMock.On("Output", mock.Anything, mock.MatchedBy(func(err error) bool { return err != nil })).Return(errOut)
		gameMock := new(interfaces.MockBeggarMyNeighbourGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("Step").Return(domain.ErrWrongPhase)

		bi := usecase.NewBeggarMyNeighbourInteractor(gameMock, wpMock)
		assert.Equal(t, errOut, bi.Step())
	})

	t.Run("game ended blocks step", func(t *testing.T) {
		endOut := `{"gameEnd":true}`
		wpMock := new(presenter.MockBeggarMyNeighbourPresenter)
		wpMock.On("Output", mock.Anything, mock.Anything).Return(endOut)
		gameMock := new(interfaces.MockBeggarMyNeighbourGame)
		gameMock.On("GetGameEndFlag").Return(true)

		bi := usecase.NewBeggarMyNeighbourInteractor(gameMock, wpMock)
		assert.Equal(t, endOut, bi.Step())
		gameMock.AssertNotCalled(t, "Step")
	})
}

func TestBeggarMyNeighbourInteractor_AutoPlay(t *testing.T) {
	out := `{"phase":3,"gameEndFlag":true}`

	t.Run("success", func(t *testing.T) {
		wpMock := new(presenter.MockBeggarMyNeighbourPresenter)
		wpMock.On("Output", mock.Anything, nil).Return(out)
		gameMock := new(interfaces.MockBeggarMyNeighbourGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("AutoPlay").Return(nil)

		bi := usecase.NewBeggarMyNeighbourInteractor(gameMock, wpMock)
		assert.Equal(t, out, bi.AutoPlay())
		gameMock.AssertCalled(t, "AutoPlay")
	})

	t.Run("autoplay error", func(t *testing.T) {
		errOut := `{"error":"wrong phase"}`
		wpMock := new(presenter.MockBeggarMyNeighbourPresenter)
		wpMock.On("Output", mock.Anything, mock.MatchedBy(func(err error) bool { return err != nil })).Return(errOut)
		gameMock := new(interfaces.MockBeggarMyNeighbourGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("AutoPlay").Return(domain.ErrWrongPhase)

		bi := usecase.NewBeggarMyNeighbourInteractor(gameMock, wpMock)
		assert.Equal(t, errOut, bi.AutoPlay())
	})

	t.Run("game ended blocks autoplay", func(t *testing.T) {
		endOut := `{"gameEnd":true}`
		wpMock := new(presenter.MockBeggarMyNeighbourPresenter)
		wpMock.On("Output", mock.Anything, mock.Anything).Return(endOut)
		gameMock := new(interfaces.MockBeggarMyNeighbourGame)
		gameMock.On("GetGameEndFlag").Return(true)

		bi := usecase.NewBeggarMyNeighbourInteractor(gameMock, wpMock)
		assert.Equal(t, endOut, bi.AutoPlay())
		gameMock.AssertNotCalled(t, "AutoPlay")
	})
}

func TestBeggarMyNeighbourInteractor_GetConfig(t *testing.T) {
	cfg := domain.BeggarMyNeighbourConfig{MaxRounds: 1000}
	gameMock := new(interfaces.MockBeggarMyNeighbourGame)
	gameMock.On("GetConfig").Return(cfg)
	wpMock := new(presenter.MockBeggarMyNeighbourPresenter)

	bi := usecase.NewBeggarMyNeighbourInteractor(gameMock, wpMock)
	assert.Equal(t, cfg, bi.GetConfig())
}

func TestBeggarMyNeighbourInteractor_ActionLog(t *testing.T) {
	logOut := `{"log":[]}`
	wpMock := new(presenter.MockBeggarMyNeighbourPresenter)
	wpMock.On("ActionLogOutput", mock.Anything).Return(logOut)
	gameMock := new(interfaces.MockBeggarMyNeighbourGame)

	bi := usecase.NewBeggarMyNeighbourInteractor(gameMock, wpMock)
	assert.Equal(t, logOut, bi.ActionLog())
}

func TestBeggarMyNeighbourInteractor_SnapshotAndRestore(t *testing.T) {
	game := newTestBeggarMyNeighbour()
	game.Reset()
	wpMock := new(presenter.MockBeggarMyNeighbourPresenter)
	wpMock.On("Output", mock.Anything, mock.Anything).Return("{}")

	bi := usecase.NewBeggarMyNeighbourInteractor(game, wpMock)
	data, err := bi.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreBeggarMyNeighbourInteractor(data, wpMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestoreBeggarMyNeighbourInteractor_InvalidJSON(t *testing.T) {
	wpMock := new(presenter.MockBeggarMyNeighbourPresenter)
	_, err := usecase.RestoreBeggarMyNeighbourInteractor([]byte("not-json"), wpMock)
	assert.Error(t, err)
}
