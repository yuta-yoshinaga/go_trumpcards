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

func TestNewPitchInteractor_NilGuards(t *testing.T) {
	ppMock := new(presenter.MockPitchPresenter)

	t.Run("panics when p is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "PitchInteractor: p must not be nil", func() {
			usecase.NewPitchInteractor(nil, ppMock)
		})
	})
	t.Run("panics when pp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockPitchGame)
		assert.PanicsWithValue(t, "PitchInteractor: pp must not be nil", func() {
			usecase.NewPitchInteractor(gameMock, nil)
		})
	})
}

func TestPitchInteractor_Reset_StaysInBidPhase(t *testing.T) {
	mockOutput := `{"phase":0}`
	ppMock := new(presenter.MockPitchPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockPitchGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.PitchPhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	pi := usecase.NewPitchInteractor(gameMock, ppMock)
	result := pi.Reset()
	assert.Equal(t, mockOutput, result)
	gameMock.AssertCalled(t, "Reset")
	gameMock.AssertNotCalled(t, "CpuPlay")
}

func TestPitchInteractor_Reset_RunsCpuTurnsAfterPlay(t *testing.T) {
	mockOutput := `{"phase":1}`
	ppMock := new(presenter.MockPitchPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockPitchGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.PitchPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	pi := usecase.NewPitchInteractor(gameMock, ppMock)
	assert.Equal(t, mockOutput, pi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestPitchInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		mockOutput := `{"phase":0}`
		ppMock := new(presenter.MockPitchPresenter)
		ppMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockPitchGame)
		cfg := domain.PitchConfig{CpuDifficulty: domain.PitchCpuDifficultyHard, PointLimit: 11}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.PitchPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		pi := usecase.NewPitchInteractor(gameMock, ppMock)
		assert.Equal(t, mockOutput, pi.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config short-circuits", func(t *testing.T) {
		ppMock := new(presenter.MockPitchPresenter)
		gameMock := new(interfaces.MockPitchGame)
		ppMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

		pi := usecase.NewPitchInteractor(gameMock, ppMock)
		bad := domain.PitchConfig{CpuDifficulty: domain.PitchCpuDifficulty(-1), PointLimit: 7}
		assert.Equal(t, "validation error", pi.ResetWithConfig(bad))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestPitchInteractor_Bid(t *testing.T) {
	t.Run("game ended short-circuits", func(t *testing.T) {
		mockOutput := `{"phase":4}`
		ppMock := new(presenter.MockPitchPresenter)
		ppMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockPitchGame)
		gameMock.On("GetGameEndFlag").Return(true)

		pi := usecase.NewPitchInteractor(gameMock, ppMock)
		assert.Equal(t, mockOutput, pi.Bid(3))
		gameMock.AssertNotCalled(t, "PlayerBid")
	})

	t.Run("invalid bid returns error output", func(t *testing.T) {
		ppMock := new(presenter.MockPitchPresenter)
		gameMock := new(interfaces.MockPitchGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", 1).Return(errors.New("invalid"))
		ppMock.On("Output", gameMock, mock.Anything).Return("err")

		pi := usecase.NewPitchInteractor(gameMock, ppMock)
		assert.Equal(t, "err", pi.Bid(1))
	})

	t.Run("valid bid runs CPU bids", func(t *testing.T) {
		ppMock := new(presenter.MockPitchPresenter)
		gameMock := new(interfaces.MockPitchGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", 3).Return(nil)
		gameMock.On("GetPhase").Return(domain.PitchPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true) // CPU loop ends immediately
		ppMock.On("Output", gameMock, mock.Anything).Return("ok")

		pi := usecase.NewPitchInteractor(gameMock, ppMock)
		assert.Equal(t, "ok", pi.Bid(3))
		gameMock.AssertCalled(t, "PlayerBid", 3)
	})
}

func TestPitchInteractor_Play(t *testing.T) {
	t.Run("not playable short-circuits", func(t *testing.T) {
		ppMock := new(presenter.MockPitchPresenter)
		gameMock := new(interfaces.MockPitchGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ppMock.On("Output", mock.Anything, mock.Anything).Return("ended")

		pi := usecase.NewPitchInteractor(gameMock, ppMock)
		assert.Equal(t, "ended", pi.Play(0))
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("invalid play returns error", func(t *testing.T) {
		ppMock := new(presenter.MockPitchPresenter)
		gameMock := new(interfaces.MockPitchGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 99).Return(errors.New("oob"))
		ppMock.On("Output", gameMock, mock.Anything).Return("err")

		pi := usecase.NewPitchInteractor(gameMock, ppMock)
		assert.Equal(t, "err", pi.Play(99))
	})

	t.Run("valid play runs CPU turns", func(t *testing.T) {
		ppMock := new(presenter.MockPitchPresenter)
		gameMock := new(interfaces.MockPitchGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("GetPhase").Return(domain.PitchPhasePlay)
		gameMock.On("PlayerPlay", 0).Return(nil)
		ppMock.On("Output", gameMock, mock.Anything).Return("ok")

		pi := usecase.NewPitchInteractor(gameMock, ppMock)
		assert.Equal(t, "ok", pi.Play(0))
	})
}

func TestPitchInteractor_NextTrick(t *testing.T) {
	ppMock := new(presenter.MockPitchPresenter)
	gameMock := new(interfaces.MockPitchGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.PitchPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	ppMock.On("Output", gameMock, mock.Anything).Return("ok")

	pi := usecase.NewPitchInteractor(gameMock, ppMock)
	assert.Equal(t, "ok", pi.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestPitchInteractor_NextRound(t *testing.T) {
	t.Run("scoring then next round", func(t *testing.T) {
		ppMock := new(presenter.MockPitchPresenter)
		gameMock := new(interfaces.MockPitchGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.PitchPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)
		ppMock.On("Output", gameMock, mock.Anything).Return("ok")

		pi := usecase.NewPitchInteractor(gameMock, ppMock)
		assert.Equal(t, "ok", pi.NextRound())
	})

	t.Run("game ended after scoring", func(t *testing.T) {
		ppMock := new(presenter.MockPitchPresenter)
		gameMock := new(interfaces.MockPitchGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(true)
		ppMock.On("Output", gameMock, mock.Anything).Return("end")

		pi := usecase.NewPitchInteractor(gameMock, ppMock)
		assert.Equal(t, "end", pi.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})
}

func TestPitchInteractor_GetConfig(t *testing.T) {
	ppMock := new(presenter.MockPitchPresenter)
	gameMock := new(interfaces.MockPitchGame)
	cfg := domain.DefaultPitchConfig()
	gameMock.On("GetConfig").Return(cfg)

	pi := usecase.NewPitchInteractor(gameMock, ppMock)
	assert.Equal(t, cfg, pi.GetConfig())
}

func TestPitchInteractor_HintAndActionLog(t *testing.T) {
	ppMock := new(presenter.MockPitchPresenter)
	gameMock := new(interfaces.MockPitchGame)
	ppMock.On("HintOutput", gameMock).Return("hint")
	ppMock.On("ActionLogOutput", gameMock).Return("log")

	pi := usecase.NewPitchInteractor(gameMock, ppMock)
	assert.Equal(t, "hint", pi.Hint())
	assert.Equal(t, "log", pi.ActionLog())
}

func TestPitchInteractor_Snapshot(t *testing.T) {
	g := domain.NewDefaultPitch()
	pi := usecase.NewPitchInteractor(g, new(presenter.MockPitchPresenter))
	data, err := pi.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestorePitchInteractor(t *testing.T) {
	g := domain.NewDefaultPitch()
	g.Reset()
	pi := usecase.NewPitchInteractor(g, new(presenter.MockPitchPresenter))
	data, err := pi.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestorePitchInteractor(data, new(presenter.MockPitchPresenter))
	assert.NoError(t, err)
	assert.NotNil(t, restored)
	assert.Equal(t, g.GetConfig(), restored.GetConfig())
}

func TestRestorePitchInteractor_InvalidData(t *testing.T) {
	_, err := usecase.RestorePitchInteractor([]byte("not json"), new(presenter.MockPitchPresenter))
	assert.Error(t, err)
}
