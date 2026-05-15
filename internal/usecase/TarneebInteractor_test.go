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

func TestNewTarneebInteractor_NilGuards(t *testing.T) {
	tp := new(presenter.MockTarneebPresenter)

	t.Run("panics when game is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "TarneebInteractor: t must not be nil", func() {
			usecase.NewTarneebInteractor(nil, tp)
		})
	})

	t.Run("panics when presenter is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockTarneebGame)
		assert.PanicsWithValue(t, "TarneebInteractor: tp must not be nil", func() {
			usecase.NewTarneebInteractor(gameMock, nil)
		})
	})
}

func TestTarneebInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("stays in bid phase waiting for human", func(t *testing.T) {
		tp := new(presenter.MockTarneebPresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockTarneebGame)
		game.On("Reset").Return()
		game.On("GetGameEndFlag").Return(false)
		game.On("GetPhase").Return(domain.TarneebPhaseBid)
		game.On("IsHumanBidTurn").Return(true)

		ti := usecase.NewTarneebInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.Reset())
		game.AssertCalled(t, "Reset")
		game.AssertNotCalled(t, "CpuBid")
	})

	t.Run("transitions to play and runs CPU turns", func(t *testing.T) {
		tp := new(presenter.MockTarneebPresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockTarneebGame)
		game.On("Reset").Return()
		game.On("GetGameEndFlag").Return(false)
		game.On("GetPhase").Return(domain.TarneebPhasePlay)
		game.On("IsHumanTurn").Return(true)

		ti := usecase.NewTarneebInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.Reset())
	})
}

func TestTarneebInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("sets config then resets", func(t *testing.T) {
		tp := new(presenter.MockTarneebPresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockTarneebGame)
		cfg := domain.TarneebConfig{CpuDifficulty: domain.TarneebCpuDifficultyHard, PointLimit: 41, MinBid: 7}
		game.On("SetConfig", cfg).Return()
		game.On("Reset").Return()
		game.On("GetGameEndFlag").Return(false)
		game.On("GetPhase").Return(domain.TarneebPhaseBid)
		game.On("IsHumanBidTurn").Return(true)

		ti := usecase.NewTarneebInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.ResetWithConfig(cfg))
		game.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config short-circuits", func(t *testing.T) {
		tp := new(presenter.MockTarneebPresenter)
		game := new(interfaces.MockTarneebGame)
		tp.On("Output", game, mock.MatchedBy(func(err error) bool { return err != nil })).Return("err")

		ti := usecase.NewTarneebInteractor(game, tp)
		cfg := domain.TarneebConfig{CpuDifficulty: domain.TarneebCpuDifficulty(-1), PointLimit: 31, MinBid: 7}
		assert.Equal(t, "err", ti.ResetWithConfig(cfg))
		game.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestTarneebInteractor_Bid(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended short-circuits", func(t *testing.T) {
		tp := new(presenter.MockTarneebPresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockTarneebGame)
		game.On("GetGameEndFlag").Return(true)
		ti := usecase.NewTarneebInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.Bid(7))
		game.AssertNotCalled(t, "PlayerBid")
	})

	t.Run("bid error surfaced", func(t *testing.T) {
		bidErr := errors.New("bad bid")
		tp := new(presenter.MockTarneebPresenter)
		tp.On("Output", mock.Anything, bidErr).Return(mockOutput)
		game := new(interfaces.MockTarneebGame)
		game.On("GetGameEndFlag").Return(false)
		game.On("PlayerBid", 5).Return(bidErr)
		ti := usecase.NewTarneebInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.Bid(5))
	})

	t.Run("valid bid transitions to trump declaration with CPU", func(t *testing.T) {
		tp := new(presenter.MockTarneebPresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockTarneebGame)
		game.On("GetGameEndFlag").Return(false)
		game.On("PlayerBid", 8).Return(nil)
		// Bid phase ends → trump declaration with CPU as bid winner → CpuDeclareTrump
		// returnSeq is intentionally ordered. First call: after PlayerBid we still inspect Bid phase loop.
		game.On("IsHumanBidTurn").Return(true)
		game.On("GetPhase").Return(domain.TarneebPhaseTrumpDeclaration)
		game.On("IsHumanTrumpTurn").Return(false)
		game.On("CpuDeclareTrump").Return()
		// After CPU declared, GetPhase is now mock'd as Play in the same `On` chain isn't ergonomic;
		// the easiest path: stub it to remain TrumpDeclaration so the play branch is skipped.
		ti := usecase.NewTarneebInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.Bid(8))
		game.AssertCalled(t, "CpuDeclareTrump")
	})
}

func TestTarneebInteractor_DeclareTrump(t *testing.T) {
	mockOutput := `{"phase":2}`

	t.Run("game ended short-circuits", func(t *testing.T) {
		tp := new(presenter.MockTarneebPresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockTarneebGame)
		game.On("GetGameEndFlag").Return(true)
		ti := usecase.NewTarneebInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.DeclareTrump(domain.CardDesignSpade))
		game.AssertNotCalled(t, "PlayerDeclareTrump")
	})

	t.Run("error surfaced", func(t *testing.T) {
		errBad := errors.New("bad trump")
		tp := new(presenter.MockTarneebPresenter)
		tp.On("Output", mock.Anything, errBad).Return(mockOutput)
		game := new(interfaces.MockTarneebGame)
		game.On("GetGameEndFlag").Return(false)
		game.On("PlayerDeclareTrump", 99).Return(errBad)
		ti := usecase.NewTarneebInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.DeclareTrump(99))
	})

	t.Run("valid declaration transitions to play (human turn)", func(t *testing.T) {
		tp := new(presenter.MockTarneebPresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockTarneebGame)
		game.On("GetGameEndFlag").Return(false)
		game.On("PlayerDeclareTrump", domain.CardDesignSpade).Return(nil)
		game.On("GetPhase").Return(domain.TarneebPhasePlay)
		game.On("IsHumanTurn").Return(true)
		ti := usecase.NewTarneebInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.DeclareTrump(domain.CardDesignSpade))
	})
}

func TestTarneebInteractor_Play(t *testing.T) {
	mockOutput := `{"phase":2}`

	t.Run("game ended short-circuits", func(t *testing.T) {
		tp := new(presenter.MockTarneebPresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockTarneebGame)
		game.On("GetGameEndFlag").Return(true)
		ti := usecase.NewTarneebInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.Play(0))
		game.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("not human turn short-circuits", func(t *testing.T) {
		tp := new(presenter.MockTarneebPresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockTarneebGame)
		game.On("GetGameEndFlag").Return(false)
		game.On("IsHumanTurn").Return(false)
		ti := usecase.NewTarneebInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.Play(0))
		game.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("play error surfaced", func(t *testing.T) {
		errBad := errors.New("bad play")
		tp := new(presenter.MockTarneebPresenter)
		tp.On("Output", mock.Anything, errBad).Return(mockOutput)
		game := new(interfaces.MockTarneebGame)
		game.On("GetGameEndFlag").Return(false)
		game.On("IsHumanTurn").Return(true)
		game.On("PlayerPlay", 0).Return(errBad)
		ti := usecase.NewTarneebInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.Play(0))
	})

	t.Run("valid play loops CPU turns", func(t *testing.T) {
		tp := new(presenter.MockTarneebPresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockTarneebGame)
		game.On("GetGameEndFlag").Return(false)
		game.On("IsHumanTurn").Return(true).Once()
		game.On("PlayerPlay", 2).Return(nil)
		game.On("GetPhase").Return(domain.TarneebPhasePlay)
		game.On("IsHumanTurn").Return(true)
		ti := usecase.NewTarneebInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.Play(2))
		game.AssertCalled(t, "PlayerPlay", 2)
	})
}

func TestTarneebInteractor_NextTrick(t *testing.T) {
	mockOutput := `{"phase":3}`
	tp := new(presenter.MockTarneebPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	game := new(interfaces.MockTarneebGame)
	game.On("GetGameEndFlag").Return(false)
	game.On("NextTrick").Return()
	game.On("GetPhase").Return(domain.TarneebPhasePlay)
	game.On("IsHumanTurn").Return(true)

	ti := usecase.NewTarneebInteractor(game, tp)
	assert.Equal(t, mockOutput, ti.NextTrick())
	game.AssertCalled(t, "NextTrick")
}

func TestTarneebInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":4}`
	tp := new(presenter.MockTarneebPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	game := new(interfaces.MockTarneebGame)
	game.On("ScoreRound").Return()
	game.On("GetGameEndFlag").Return(false)
	game.On("NextRound").Return()
	game.On("GetPhase").Return(domain.TarneebPhaseBid)
	game.On("IsHumanBidTurn").Return(true)

	ti := usecase.NewTarneebInteractor(game, tp)
	assert.Equal(t, mockOutput, ti.NextRound())
	game.AssertCalled(t, "ScoreRound")
	game.AssertCalled(t, "NextRound")
}

func TestTarneebInteractor_GetConfig(t *testing.T) {
	game := new(interfaces.MockTarneebGame)
	cfg := domain.DefaultTarneebConfig()
	game.On("GetConfig").Return(cfg)
	tp := new(presenter.MockTarneebPresenter)
	ti := usecase.NewTarneebInteractor(game, tp)
	assert.Equal(t, cfg, ti.GetConfig())
}

func TestTarneebInteractor_Hint(t *testing.T) {
	game := new(interfaces.MockTarneebGame)
	tp := new(presenter.MockTarneebPresenter)
	tp.On("HintOutput", game).Return("hint")
	ti := usecase.NewTarneebInteractor(game, tp)
	assert.Equal(t, "hint", ti.Hint())
}

func TestTarneebInteractor_ActionLog(t *testing.T) {
	game := new(interfaces.MockTarneebGame)
	tp := new(presenter.MockTarneebPresenter)
	tp.On("ActionLogOutput", game).Return("log")
	ti := usecase.NewTarneebInteractor(game, tp)
	assert.Equal(t, "log", ti.ActionLog())
}

func TestRestoreTarneebInteractor(t *testing.T) {
	tn := domain.NewDefaultTarneeb()
	tn.Reset()
	data, err := tn.MarshalJSON()
	assert.NoError(t, err)
	tp := new(presenter.MockTarneebPresenter)
	restored, err := usecase.RestoreTarneebInteractor(data, tp)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}
