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

func newTestSlapjack() *domain.Slapjack {
	return domain.NewDefaultSlapjack()
}

func TestNewSlapjackInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockSlapjackPresenter)

	t.Run("panics when s is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "SlapjackInteractor: s must not be nil", func() {
			usecase.NewSlapjackInteractor(nil, spMock)
		})
	})

	t.Run("panics when sp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "SlapjackInteractor: sp must not be nil", func() {
			usecase.NewSlapjackInteractor(newTestSlapjack(), nil)
		})
	})
}

func TestSlapjackInteractor_Reset(t *testing.T) {
	out := `{"phase":0}`
	spMock := new(presenter.MockSlapjackPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(out)
	gameMock := new(interfaces.MockSlapjackGame)
	gameMock.On("Reset").Return()

	si := usecase.NewSlapjackInteractor(gameMock, spMock)
	assert.Equal(t, out, si.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestSlapjackInteractor_ResetWithConfig(t *testing.T) {
	out := `{"phase":0}`

	t.Run("valid config", func(t *testing.T) {
		spMock := new(presenter.MockSlapjackPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(out)
		gameMock := new(interfaces.MockSlapjackGame)
		cfg := domain.SlapjackConfig{CpuDifficulty: domain.SlapjackCpuHard}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		si := usecase.NewSlapjackInteractor(gameMock, spMock)
		assert.Equal(t, out, si.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error", func(t *testing.T) {
		errOut := `{"error":"invalid"}`
		spMock := new(presenter.MockSlapjackPresenter)
		spMock.On("Output", mock.Anything, mock.MatchedBy(func(err error) bool { return err != nil })).Return(errOut)
		gameMock := new(interfaces.MockSlapjackGame)

		si := usecase.NewSlapjackInteractor(gameMock, spMock)
		assert.Equal(t, errOut, si.ResetWithConfig(domain.SlapjackConfig{CpuDifficulty: 99}))
	})
}

func TestSlapjackInteractor_Step(t *testing.T) {
	out := `{"phase":0}`

	t.Run("success", func(t *testing.T) {
		spMock := new(presenter.MockSlapjackPresenter)
		spMock.On("Output", mock.Anything, nil).Return(out)
		gameMock := new(interfaces.MockSlapjackGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("Step").Return(nil)

		si := usecase.NewSlapjackInteractor(gameMock, spMock)
		assert.Equal(t, out, si.Step())
	})

	t.Run("error from domain", func(t *testing.T) {
		errOut := `{"error":"invalid"}`
		spMock := new(presenter.MockSlapjackPresenter)
		spMock.On("Output", mock.Anything, mock.MatchedBy(func(err error) bool { return err != nil })).Return(errOut)
		gameMock := new(interfaces.MockSlapjackGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("Step").Return(domain.ErrInvalidPlay)

		si := usecase.NewSlapjackInteractor(gameMock, spMock)
		assert.Equal(t, errOut, si.Step())
	})

	t.Run("blocked when game ended", func(t *testing.T) {
		endOut := `{"gameEnd":true}`
		spMock := new(presenter.MockSlapjackPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(endOut)
		gameMock := new(interfaces.MockSlapjackGame)
		gameMock.On("GetGameEndFlag").Return(true)

		si := usecase.NewSlapjackInteractor(gameMock, spMock)
		assert.Equal(t, endOut, si.Step())
		gameMock.AssertNotCalled(t, "Step")
	})
}

func TestSlapjackInteractor_Slap(t *testing.T) {
	out := `{"phase":0}`

	t.Run("success", func(t *testing.T) {
		spMock := new(presenter.MockSlapjackPresenter)
		spMock.On("Output", mock.Anything, nil).Return(out)
		gameMock := new(interfaces.MockSlapjackGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("Slap", 0).Return(nil)

		si := usecase.NewSlapjackInteractor(gameMock, spMock)
		assert.Equal(t, out, si.Slap(0))
	})

	t.Run("error from domain", func(t *testing.T) {
		errOut := `{"error":"invalid"}`
		spMock := new(presenter.MockSlapjackPresenter)
		spMock.On("Output", mock.Anything, mock.MatchedBy(func(err error) bool { return err != nil })).Return(errOut)
		gameMock := new(interfaces.MockSlapjackGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("Slap", 0).Return(domain.ErrInvalidPlay)

		si := usecase.NewSlapjackInteractor(gameMock, spMock)
		assert.Equal(t, errOut, si.Slap(0))
	})

	t.Run("blocked when game ended", func(t *testing.T) {
		endOut := `{"gameEnd":true}`
		spMock := new(presenter.MockSlapjackPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(endOut)
		gameMock := new(interfaces.MockSlapjackGame)
		gameMock.On("GetGameEndFlag").Return(true)

		si := usecase.NewSlapjackInteractor(gameMock, spMock)
		assert.Equal(t, endOut, si.Slap(0))
		gameMock.AssertNotCalled(t, "Slap", mock.Anything)
	})
}

func TestSlapjackInteractor_Tick(t *testing.T) {
	out := `{"phase":0}`
	spMock := new(presenter.MockSlapjackPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(out)
	gameMock := new(interfaces.MockSlapjackGame)
	gameMock.On("Tick").Return(domain.SlapjackPendingNone)

	si := usecase.NewSlapjackInteractor(gameMock, spMock)
	assert.Equal(t, out, si.Tick())
	gameMock.AssertCalled(t, "Tick")
}

func TestSlapjackInteractor_GetConfig(t *testing.T) {
	cfg := domain.SlapjackConfig{CpuDifficulty: domain.SlapjackCpuHard}
	gameMock := new(interfaces.MockSlapjackGame)
	gameMock.On("GetConfig").Return(cfg)
	spMock := new(presenter.MockSlapjackPresenter)

	si := usecase.NewSlapjackInteractor(gameMock, spMock)
	assert.Equal(t, cfg, si.GetConfig())
}

func TestSlapjackInteractor_ActionLog(t *testing.T) {
	logOut := `{"log":[]}`
	spMock := new(presenter.MockSlapjackPresenter)
	spMock.On("ActionLogOutput", mock.Anything).Return(logOut)
	gameMock := new(interfaces.MockSlapjackGame)

	si := usecase.NewSlapjackInteractor(gameMock, spMock)
	assert.Equal(t, logOut, si.ActionLog())
}

func TestSlapjackInteractor_SnapshotAndRestore(t *testing.T) {
	game := newTestSlapjack()
	game.Reset()
	spMock := new(presenter.MockSlapjackPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("{}")

	si := usecase.NewSlapjackInteractor(game, spMock)
	data, err := si.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreSlapjackInteractor(data, spMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestoreSlapjackInteractor_InvalidJSON(t *testing.T) {
	spMock := new(presenter.MockSlapjackPresenter)
	_, err := usecase.RestoreSlapjackInteractor([]byte("not-json"), spMock)
	assert.Error(t, err)
}
