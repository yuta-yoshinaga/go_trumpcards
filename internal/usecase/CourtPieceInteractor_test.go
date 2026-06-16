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

func TestNewCourtPieceInteractor_NilGuards(t *testing.T) {
	tp := new(presenter.MockCourtPiecePresenter)

	t.Run("panics when game is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "CourtPieceInteractor: t must not be nil", func() {
			usecase.NewCourtPieceInteractor(nil, tp)
		})
	})

	t.Run("panics when presenter is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockCourtPieceGame)
		assert.PanicsWithValue(t, "CourtPieceInteractor: tp must not be nil", func() {
			usecase.NewCourtPieceInteractor(gameMock, nil)
		})
	})
}

func TestCourtPieceInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("stays in trump declaration waiting for human caller", func(t *testing.T) {
		tp := new(presenter.MockCourtPiecePresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockCourtPieceGame)
		game.On("Reset").Return()
		game.On("GetPhase").Return(domain.CourtPiecePhaseTrumpDeclaration)
		game.On("IsHumanTrumpTurn").Return(true)

		ti := usecase.NewCourtPieceInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.Reset())
		game.AssertCalled(t, "Reset")
		game.AssertNotCalled(t, "CpuDeclareTrump")
	})

	t.Run("CPU caller declares then transitions to play", func(t *testing.T) {
		tp := new(presenter.MockCourtPiecePresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockCourtPieceGame)
		game.On("Reset").Return()
		game.On("GetPhase").Return(domain.CourtPiecePhaseTrumpDeclaration).Once()
		game.On("IsHumanTrumpTurn").Return(false)
		game.On("CpuDeclareTrump").Return()
		game.On("GetPhase").Return(domain.CourtPiecePhasePlay)
		game.On("GetGameEndFlag").Return(false)
		game.On("IsHumanTurn").Return(true)

		ti := usecase.NewCourtPieceInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.Reset())
		game.AssertCalled(t, "CpuDeclareTrump")
	})

	t.Run("transitions to play and runs CPU turns", func(t *testing.T) {
		tp := new(presenter.MockCourtPiecePresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockCourtPieceGame)
		game.On("Reset").Return()
		game.On("GetPhase").Return(domain.CourtPiecePhasePlay)
		game.On("GetGameEndFlag").Return(false)
		game.On("IsHumanTurn").Return(true)

		ti := usecase.NewCourtPieceInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.Reset())
	})
}

func TestCourtPieceInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("sets config then resets", func(t *testing.T) {
		tp := new(presenter.MockCourtPiecePresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockCourtPieceGame)
		cfg := domain.CourtPieceConfig{CpuDifficulty: domain.CourtPieceCpuDifficultyHard, PointLimit: 9}
		game.On("SetConfig", cfg).Return()
		game.On("Reset").Return()
		game.On("GetPhase").Return(domain.CourtPiecePhaseTrumpDeclaration)
		game.On("IsHumanTrumpTurn").Return(true)

		ti := usecase.NewCourtPieceInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.ResetWithConfig(cfg))
		game.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config short-circuits", func(t *testing.T) {
		tp := new(presenter.MockCourtPiecePresenter)
		game := new(interfaces.MockCourtPieceGame)
		tp.On("Output", game, mock.MatchedBy(func(err error) bool { return err != nil })).Return("err")

		ti := usecase.NewCourtPieceInteractor(game, tp)
		cfg := domain.CourtPieceConfig{CpuDifficulty: domain.CourtPieceCpuDifficulty(-1), PointLimit: 7}
		assert.Equal(t, "err", ti.ResetWithConfig(cfg))
		game.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestCourtPieceInteractor_DeclareTrump(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended short-circuits", func(t *testing.T) {
		tp := new(presenter.MockCourtPiecePresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockCourtPieceGame)
		game.On("GetGameEndFlag").Return(true)
		ti := usecase.NewCourtPieceInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.DeclareTrump(domain.CardDesignSpade))
		game.AssertNotCalled(t, "PlayerDeclareTrump")
	})

	t.Run("error surfaced", func(t *testing.T) {
		errBad := errors.New("bad trump")
		tp := new(presenter.MockCourtPiecePresenter)
		tp.On("Output", mock.Anything, errBad).Return(mockOutput)
		game := new(interfaces.MockCourtPieceGame)
		game.On("GetGameEndFlag").Return(false)
		game.On("PlayerDeclareTrump", 99).Return(errBad)
		ti := usecase.NewCourtPieceInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.DeclareTrump(99))
	})

	t.Run("valid declaration transitions to play (human turn)", func(t *testing.T) {
		tp := new(presenter.MockCourtPiecePresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockCourtPieceGame)
		game.On("GetGameEndFlag").Return(false)
		game.On("PlayerDeclareTrump", domain.CardDesignSpade).Return(nil)
		game.On("GetPhase").Return(domain.CourtPiecePhasePlay)
		game.On("IsHumanTurn").Return(true)
		ti := usecase.NewCourtPieceInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.DeclareTrump(domain.CardDesignSpade))
	})
}

func TestCourtPieceInteractor_Play(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended short-circuits", func(t *testing.T) {
		tp := new(presenter.MockCourtPiecePresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockCourtPieceGame)
		game.On("GetGameEndFlag").Return(true)
		ti := usecase.NewCourtPieceInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.Play(0))
		game.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("not human turn short-circuits", func(t *testing.T) {
		tp := new(presenter.MockCourtPiecePresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockCourtPieceGame)
		game.On("GetGameEndFlag").Return(false)
		game.On("IsHumanTurn").Return(false)
		ti := usecase.NewCourtPieceInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.Play(0))
		game.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("play error surfaced", func(t *testing.T) {
		errBad := errors.New("bad play")
		tp := new(presenter.MockCourtPiecePresenter)
		tp.On("Output", mock.Anything, errBad).Return(mockOutput)
		game := new(interfaces.MockCourtPieceGame)
		game.On("GetGameEndFlag").Return(false)
		game.On("IsHumanTurn").Return(true)
		game.On("PlayerPlay", 0).Return(errBad)
		ti := usecase.NewCourtPieceInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.Play(0))
	})

	t.Run("valid play loops CPU turns", func(t *testing.T) {
		tp := new(presenter.MockCourtPiecePresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockCourtPieceGame)
		game.On("GetGameEndFlag").Return(false)
		game.On("IsHumanTurn").Return(true).Once()
		game.On("PlayerPlay", 2).Return(nil)
		game.On("GetPhase").Return(domain.CourtPiecePhasePlay)
		game.On("IsHumanTurn").Return(true)
		ti := usecase.NewCourtPieceInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.Play(2))
		game.AssertCalled(t, "PlayerPlay", 2)
	})

	t.Run("human completes trick calls ResolveTrick", func(t *testing.T) {
		// When the human plays the 4th card of a trick, PlayerPlay flips the
		// phase to TrickEnd. The interactor must call ResolveTrick before
		// looping to CPU turns; otherwise leadPlayerIdx and trick counts drift.
		tp := new(presenter.MockCourtPiecePresenter)
		tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := new(interfaces.MockCourtPieceGame)
		game.On("GetGameEndFlag").Return(false)
		game.On("IsHumanTurn").Return(true).Once()
		game.On("PlayerPlay", 0).Return(nil)
		game.On("GetPhase").Return(domain.CourtPiecePhaseTrickEnd).Once()
		game.On("ResolveTrick").Return()
		game.On("GetPhase").Return(domain.CourtPiecePhaseTrickEnd)
		ti := usecase.NewCourtPieceInteractor(game, tp)
		assert.Equal(t, mockOutput, ti.Play(0))
		game.AssertCalled(t, "ResolveTrick")
	})
}

func TestCourtPieceInteractor_NextTrick(t *testing.T) {
	mockOutput := `{"phase":2}`
	tp := new(presenter.MockCourtPiecePresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	game := new(interfaces.MockCourtPieceGame)
	game.On("GetGameEndFlag").Return(false)
	game.On("NextTrick").Return()
	game.On("GetPhase").Return(domain.CourtPiecePhasePlay)
	game.On("IsHumanTurn").Return(true)

	ti := usecase.NewCourtPieceInteractor(game, tp)
	assert.Equal(t, mockOutput, ti.NextTrick())
	game.AssertCalled(t, "NextTrick")
}

func TestCourtPieceInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":3}`
	tp := new(presenter.MockCourtPiecePresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	game := new(interfaces.MockCourtPieceGame)
	game.On("ScoreRound").Return()
	game.On("GetGameEndFlag").Return(false)
	game.On("NextRound").Return()
	game.On("GetPhase").Return(domain.CourtPiecePhaseTrumpDeclaration)
	game.On("IsHumanTrumpTurn").Return(true)

	ti := usecase.NewCourtPieceInteractor(game, tp)
	assert.Equal(t, mockOutput, ti.NextRound())
	game.AssertCalled(t, "ScoreRound")
	game.AssertCalled(t, "NextRound")
}

func TestCourtPieceInteractor_GetConfig(t *testing.T) {
	game := new(interfaces.MockCourtPieceGame)
	cfg := domain.DefaultCourtPieceConfig()
	game.On("GetConfig").Return(cfg)
	tp := new(presenter.MockCourtPiecePresenter)
	ti := usecase.NewCourtPieceInteractor(game, tp)
	assert.Equal(t, cfg, ti.GetConfig())
}

func TestCourtPieceInteractor_Hint(t *testing.T) {
	game := new(interfaces.MockCourtPieceGame)
	tp := new(presenter.MockCourtPiecePresenter)
	tp.On("HintOutput", game).Return("hint")
	ti := usecase.NewCourtPieceInteractor(game, tp)
	assert.Equal(t, "hint", ti.Hint())
}

func TestCourtPieceInteractor_ActionLog(t *testing.T) {
	game := new(interfaces.MockCourtPieceGame)
	tp := new(presenter.MockCourtPiecePresenter)
	tp.On("ActionLogOutput", game).Return("log")
	ti := usecase.NewCourtPieceInteractor(game, tp)
	assert.Equal(t, "log", ti.ActionLog())
}

func TestRestoreCourtPieceInteractor(t *testing.T) {
	cp := domain.NewDefaultCourtPiece()
	cp.Reset()
	data, err := cp.MarshalJSON()
	assert.NoError(t, err)
	tp := new(presenter.MockCourtPiecePresenter)
	restored, err := usecase.RestoreCourtPieceInteractor(data, tp)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}
