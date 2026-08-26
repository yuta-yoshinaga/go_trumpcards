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

const quadrilleMockOutput = `{"phase":0}`

func TestNewQuadrilleInteractor_NilGuards(t *testing.T) {
	tpMock := new(presenter.MockQuadrillePresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "QuadrilleInteractor: g must not be nil", func() {
			usecase.NewQuadrilleInteractor(nil, tpMock)
		})
	})
	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockQuadrilleGame)
		assert.PanicsWithValue(t, "QuadrilleInteractor: tp must not be nil", func() {
			usecase.NewQuadrilleInteractor(gameMock, nil)
		})
	})
}

func TestQuadrilleInteractor_Reset(t *testing.T) {
	tpMock := new(presenter.MockQuadrillePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(quadrilleMockOutput)
	gameMock := new(interfaces.MockQuadrilleGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.QuadrillePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewQuadrilleInteractor(gameMock, tpMock)
	assert.Equal(t, quadrilleMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestQuadrilleInteractor_ResetWithConfig(t *testing.T) {
	tpMock := new(presenter.MockQuadrillePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(quadrilleMockOutput)
	gameMock := new(interfaces.MockQuadrilleGame)
	cfg := domain.QuadrilleConfig{CpuDifficulty: domain.QuadrilleCpuDifficultyHard, TargetRounds: 5}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.QuadrillePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewQuadrilleInteractor(gameMock, tpMock)
	assert.Equal(t, quadrilleMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestQuadrilleInteractor_ResetWithConfigInvalid(t *testing.T) {
	tpMock := new(presenter.MockQuadrillePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(quadrilleMockOutput)
	gameMock := new(interfaces.MockQuadrilleGame)

	ci := usecase.NewQuadrilleInteractor(gameMock, tpMock)
	bad := domain.QuadrilleConfig{CpuDifficulty: domain.QuadrilleCpuDifficultyNormal, TargetRounds: 0}
	assert.Equal(t, quadrilleMockOutput, ci.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestQuadrilleInteractor_Bid(t *testing.T) {
	tpMock := new(presenter.MockQuadrillePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(quadrilleMockOutput)
	gameMock := new(interfaces.MockQuadrilleGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.QuadrillePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerBid", domain.QuadrilleBidEntrar, domain.CardDesignHeart).Return(nil)

	ci := usecase.NewQuadrilleInteractor(gameMock, tpMock)
	assert.Equal(t, quadrilleMockOutput, ci.Bid(domain.QuadrilleBidEntrar, domain.CardDesignHeart))
	gameMock.AssertCalled(t, "PlayerBid", domain.QuadrilleBidEntrar, domain.CardDesignHeart)
}

func TestQuadrilleInteractor_BidError(t *testing.T) {
	tpMock := new(presenter.MockQuadrillePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(quadrilleMockOutput)
	gameMock := new(interfaces.MockQuadrilleGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.QuadrilleBidSolo, -1).Return(errors.New("cannot bid"))

	ci := usecase.NewQuadrilleInteractor(gameMock, tpMock)
	assert.Equal(t, quadrilleMockOutput, ci.Bid(domain.QuadrilleBidSolo, -1))
	gameMock.AssertCalled(t, "PlayerBid", domain.QuadrilleBidSolo, -1)
}

func TestQuadrilleInteractor_BidGameEnded(t *testing.T) {
	tpMock := new(presenter.MockQuadrillePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(quadrilleMockOutput)
	gameMock := new(interfaces.MockQuadrilleGame)
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewQuadrilleInteractor(gameMock, tpMock)
	assert.Equal(t, quadrilleMockOutput, ci.Bid(domain.QuadrilleBidEntrar, domain.CardDesignHeart))
	gameMock.AssertNotCalled(t, "PlayerBid", mock.Anything, mock.Anything)
}

func TestQuadrilleInteractor_PlayResolvesTrick(t *testing.T) {
	tpMock := new(presenter.MockQuadrillePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(quadrilleMockOutput)
	gameMock := new(interfaces.MockQuadrilleGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.QuadrillePhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.QuadrillePhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	ci := usecase.NewQuadrilleInteractor(gameMock, tpMock)
	assert.Equal(t, quadrilleMockOutput, ci.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestQuadrilleInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	tpMock := new(presenter.MockQuadrillePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(quadrilleMockOutput)
	gameMock := new(interfaces.MockQuadrilleGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.QuadrillePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	ci := usecase.NewQuadrilleInteractor(gameMock, tpMock)
	assert.Equal(t, quadrilleMockOutput, ci.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestQuadrilleInteractor_PlayError(t *testing.T) {
	tpMock := new(presenter.MockQuadrillePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(quadrilleMockOutput)
	gameMock := new(interfaces.MockQuadrilleGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.QuadrillePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	ci := usecase.NewQuadrilleInteractor(gameMock, tpMock)
	assert.Equal(t, quadrilleMockOutput, ci.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestQuadrilleInteractor_PlayNotHumanTurn(t *testing.T) {
	tpMock := new(presenter.MockQuadrillePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(quadrilleMockOutput)
	gameMock := new(interfaces.MockQuadrilleGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.QuadrillePhasePlay)
	gameMock.On("IsHumanTurn").Return(false)

	ci := usecase.NewQuadrilleInteractor(gameMock, tpMock)
	assert.Equal(t, quadrilleMockOutput, ci.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestQuadrilleInteractor_NextTrick(t *testing.T) {
	tpMock := new(presenter.MockQuadrillePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(quadrilleMockOutput)
	gameMock := new(interfaces.MockQuadrilleGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.QuadrillePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewQuadrilleInteractor(gameMock, tpMock)
	assert.Equal(t, quadrilleMockOutput, ci.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestQuadrilleInteractor_NextRound(t *testing.T) {
	tpMock := new(presenter.MockQuadrillePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(quadrilleMockOutput)
	gameMock := new(interfaces.MockQuadrilleGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.QuadrillePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewQuadrilleInteractor(gameMock, tpMock)
	assert.Equal(t, quadrilleMockOutput, ci.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestQuadrilleInteractor_NextRoundGameEnded(t *testing.T) {
	tpMock := new(presenter.MockQuadrillePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(quadrilleMockOutput)
	gameMock := new(interfaces.MockQuadrilleGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewQuadrilleInteractor(gameMock, tpMock)
	assert.Equal(t, quadrilleMockOutput, ci.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestQuadrilleInteractor_GetConfigHintActionLog(t *testing.T) {
	tpMock := new(presenter.MockQuadrillePresenter)
	tpMock.On("HintOutput", mock.Anything).Return("hint")
	tpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockQuadrilleGame)
	cfg := domain.DefaultQuadrilleConfig()
	gameMock.On("GetConfig").Return(cfg)

	ci := usecase.NewQuadrilleInteractor(gameMock, tpMock)
	assert.Equal(t, cfg, ci.GetConfig())
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

func TestRestoreQuadrilleInteractor(t *testing.T) {
	tpMock := new(presenter.MockQuadrillePresenter)
	src := domain.NewDefaultQuadrille()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ci, err := usecase.RestoreQuadrilleInteractor(data, tpMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)

	_, err = usecase.RestoreQuadrilleInteractor([]byte(`{`), tpMock)
	assert.Error(t, err)
}
