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

func TestNewNinetyNineInteractor_NilGuards(t *testing.T) {
	opMock := new(presenter.MockNinetyNinePresenter)

	t.Run("panics when o is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "NinetyNineInteractor: o must not be nil", func() {
			usecase.NewNinetyNineInteractor(nil, opMock)
		})
	})

	t.Run("panics when op is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockNinetyNineGame)
		assert.PanicsWithValue(t, "NinetyNineInteractor: op must not be nil", func() {
			usecase.NewNinetyNineInteractor(gameMock, nil)
		})
	})
}

func TestNinetyNineInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("reset stays in bid phase", func(t *testing.T) {
		opMock := new(presenter.MockNinetyNinePresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNinetyNineGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.NinetyNinePhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		oi := usecase.NewNinetyNineInteractor(gameMock, opMock)
		assert.Equal(t, mockOutput, oi.Reset())
		gameMock.AssertCalled(t, "Reset")
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("reset transitions to play phase", func(t *testing.T) {
		opMock := new(presenter.MockNinetyNinePresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNinetyNineGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.NinetyNinePhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		oi := usecase.NewNinetyNineInteractor(gameMock, opMock)
		assert.Equal(t, mockOutput, oi.Reset())
		gameMock.AssertCalled(t, "Reset")
	})
}

func TestNinetyNineInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("sets config then resets", func(t *testing.T) {
		opMock := new(presenter.MockNinetyNinePresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNinetyNineGame)
		cfg := domain.NinetyNineConfig{CpuDifficulty: domain.NinetyNineCpuDifficultyHard, TargetScore: 100}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.NinetyNinePhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		oi := usecase.NewNinetyNineInteractor(gameMock, opMock)
		assert.Equal(t, mockOutput, oi.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("validation error skips config", func(t *testing.T) {
		opMock := new(presenter.MockNinetyNinePresenter)
		gameMock := new(interfaces.MockNinetyNineGame)
		opMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

		oi := usecase.NewNinetyNineInteractor(gameMock, opMock)
		cfg := domain.NinetyNineConfig{CpuDifficulty: domain.NinetyNineCpuDifficulty(-1), TargetScore: 100}
		assert.Equal(t, "validation error", oi.ResetWithConfig(cfg))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestNinetyNineInteractor_Bid(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended returns output without bidding", func(t *testing.T) {
		opMock := new(presenter.MockNinetyNinePresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNinetyNineGame)
		gameMock.On("GetGameEndFlag").Return(true)

		oi := usecase.NewNinetyNineInteractor(gameMock, opMock)
		assert.Equal(t, mockOutput, oi.Bid([]int{0, 1, 2}))
		gameMock.AssertNotCalled(t, "PlayerBid")
	})

	t.Run("bid error returns error output", func(t *testing.T) {
		opMock := new(presenter.MockNinetyNinePresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return("bid error")
		gameMock := new(interfaces.MockNinetyNineGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", []int{0, 0, 0}).Return(errors.New("invalid bid"))

		oi := usecase.NewNinetyNineInteractor(gameMock, opMock)
		assert.Equal(t, "bid error", oi.Bid([]int{0, 0, 0}))
	})

	t.Run("successful bid runs CPU bids", func(t *testing.T) {
		opMock := new(presenter.MockNinetyNinePresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNinetyNineGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", []int{0, 1, 2}).Return(nil)
		gameMock.On("GetPhase").Return(domain.NinetyNinePhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		oi := usecase.NewNinetyNineInteractor(gameMock, opMock)
		assert.Equal(t, mockOutput, oi.Bid([]int{0, 1, 2}))
		gameMock.AssertCalled(t, "PlayerBid", []int{0, 1, 2})
	})
}

func TestNinetyNineInteractor_Play(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended returns output without playing", func(t *testing.T) {
		opMock := new(presenter.MockNinetyNinePresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNinetyNineGame)
		gameMock.On("GetGameEndFlag").Return(true)

		oi := usecase.NewNinetyNineInteractor(gameMock, opMock)
		assert.Equal(t, mockOutput, oi.Play(0))
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("not human turn returns output", func(t *testing.T) {
		opMock := new(presenter.MockNinetyNinePresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNinetyNineGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		oi := usecase.NewNinetyNineInteractor(gameMock, opMock)
		assert.Equal(t, mockOutput, oi.Play(0))
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("play error returns error output", func(t *testing.T) {
		opMock := new(presenter.MockNinetyNinePresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return("play error")
		gameMock := new(interfaces.MockNinetyNineGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 99).Return(errors.New("invalid"))

		oi := usecase.NewNinetyNineInteractor(gameMock, opMock)
		assert.Equal(t, "play error", oi.Play(99))
	})

	t.Run("successful play runs CPU turns", func(t *testing.T) {
		opMock := new(presenter.MockNinetyNinePresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNinetyNineGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerPlay", 0).Return(nil)
		gameMock.On("GetPhase").Return(domain.NinetyNinePhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		oi := usecase.NewNinetyNineInteractor(gameMock, opMock)
		assert.Equal(t, mockOutput, oi.Play(0))
		gameMock.AssertCalled(t, "PlayerPlay", 0)
		gameMock.AssertNotCalled(t, "ResolveTrick")
	})

	t.Run("human completes trick calls ResolveTrick", func(t *testing.T) {
		opMock := new(presenter.MockNinetyNinePresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNinetyNineGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerPlay", 0).Return(nil)
		gameMock.On("GetPhase").Return(domain.NinetyNinePhaseTrickEnd)
		gameMock.On("ResolveTrick").Return()

		oi := usecase.NewNinetyNineInteractor(gameMock, opMock)
		assert.Equal(t, mockOutput, oi.Play(0))
		gameMock.AssertCalled(t, "ResolveTrick")
	})
}

func TestNinetyNineInteractor_NextTrick(t *testing.T) {
	mockOutput := `{"phase":1}`
	opMock := new(presenter.MockNinetyNinePresenter)
	opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockNinetyNineGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.NinetyNinePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	oi := usecase.NewNinetyNineInteractor(gameMock, opMock)
	assert.Equal(t, mockOutput, oi.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestNinetyNineInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("scores round then starts next", func(t *testing.T) {
		opMock := new(presenter.MockNinetyNinePresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNinetyNineGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.NinetyNinePhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		oi := usecase.NewNinetyNineInteractor(gameMock, opMock)
		assert.Equal(t, mockOutput, oi.NextRound())
		gameMock.AssertCalled(t, "ScoreRound")
		gameMock.AssertCalled(t, "NextRound")
	})

	t.Run("game ends after score", func(t *testing.T) {
		opMock := new(presenter.MockNinetyNinePresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEndFlag":true}`)
		gameMock := new(interfaces.MockNinetyNineGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		oi := usecase.NewNinetyNineInteractor(gameMock, opMock)
		assert.Equal(t, `{"gameEndFlag":true}`, oi.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})
}

func TestNinetyNineInteractor_GetConfig(t *testing.T) {
	opMock := new(presenter.MockNinetyNinePresenter)
	gameMock := new(interfaces.MockNinetyNineGame)
	cfg := domain.DefaultNinetyNineConfig()
	gameMock.On("GetConfig").Return(cfg)

	oi := usecase.NewNinetyNineInteractor(gameMock, opMock)
	assert.Equal(t, cfg, oi.GetConfig())
}

func TestNinetyNineInteractor_Hint(t *testing.T) {
	opMock := new(presenter.MockNinetyNinePresenter)
	gameMock := new(interfaces.MockNinetyNineGame)
	opMock.On("HintOutput", gameMock).Return("hint result")

	oi := usecase.NewNinetyNineInteractor(gameMock, opMock)
	assert.Equal(t, "hint result", oi.Hint())
}

func TestNinetyNineInteractor_ActionLog(t *testing.T) {
	opMock := new(presenter.MockNinetyNinePresenter)
	gameMock := new(interfaces.MockNinetyNineGame)
	opMock.On("ActionLogOutput", gameMock).Return("log result")

	oi := usecase.NewNinetyNineInteractor(gameMock, opMock)
	assert.Equal(t, "log result", oi.ActionLog())
}

func TestRestoreNinetyNineInteractor(t *testing.T) {
	opMock := new(presenter.MockNinetyNinePresenter)
	o := domain.NewDefaultNinetyNine()
	o.Reset()
	data, err := o.MarshalJSON()
	assert.NoError(t, err)

	oi, err := usecase.RestoreNinetyNineInteractor(data, opMock)
	assert.NoError(t, err)
	assert.NotNil(t, oi)

	_, err = usecase.RestoreNinetyNineInteractor([]byte("not json"), opMock)
	assert.Error(t, err)
}
