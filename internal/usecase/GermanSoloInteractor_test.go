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

const germanSoloMockOutput = `{"phase":0}`

func TestNewGermanSoloInteractor_NilGuards(t *testing.T) {
	tpMock := new(presenter.MockGermanSoloPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "GermanSoloInteractor: g must not be nil", func() {
			usecase.NewGermanSoloInteractor(nil, tpMock)
		})
	})
	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockGermanSoloGame)
		assert.PanicsWithValue(t, "GermanSoloInteractor: tp must not be nil", func() {
			usecase.NewGermanSoloInteractor(gameMock, nil)
		})
	})
}

func TestGermanSoloInteractor_Reset(t *testing.T) {
	tpMock := new(presenter.MockGermanSoloPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(germanSoloMockOutput)
	gameMock := new(interfaces.MockGermanSoloGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.GermanSoloPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewGermanSoloInteractor(gameMock, tpMock)
	assert.Equal(t, germanSoloMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestGermanSoloInteractor_ResetWithConfig(t *testing.T) {
	tpMock := new(presenter.MockGermanSoloPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(germanSoloMockOutput)
	gameMock := new(interfaces.MockGermanSoloGame)
	cfg := domain.GermanSoloConfig{CpuDifficulty: domain.GermanSoloCpuDifficultyHard, TargetRounds: 5}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.GermanSoloPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewGermanSoloInteractor(gameMock, tpMock)
	assert.Equal(t, germanSoloMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestGermanSoloInteractor_ResetWithConfigInvalid(t *testing.T) {
	tpMock := new(presenter.MockGermanSoloPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(germanSoloMockOutput)
	gameMock := new(interfaces.MockGermanSoloGame)

	ci := usecase.NewGermanSoloInteractor(gameMock, tpMock)
	bad := domain.GermanSoloConfig{CpuDifficulty: domain.GermanSoloCpuDifficultyNormal, TargetRounds: 0}
	assert.Equal(t, germanSoloMockOutput, ci.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestGermanSoloInteractor_Bid(t *testing.T) {
	tpMock := new(presenter.MockGermanSoloPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(germanSoloMockOutput)
	gameMock := new(interfaces.MockGermanSoloGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.GermanSoloPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerBid", domain.GermanSoloBidFrage, domain.CardDesignHeart).Return(nil)

	ci := usecase.NewGermanSoloInteractor(gameMock, tpMock)
	assert.Equal(t, germanSoloMockOutput, ci.Bid(domain.GermanSoloBidFrage, domain.CardDesignHeart))
	gameMock.AssertCalled(t, "PlayerBid", domain.GermanSoloBidFrage, domain.CardDesignHeart)
}

func TestGermanSoloInteractor_BidError(t *testing.T) {
	tpMock := new(presenter.MockGermanSoloPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(germanSoloMockOutput)
	gameMock := new(interfaces.MockGermanSoloGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.GermanSoloBidSolo, -1).Return(errors.New("cannot bid"))

	ci := usecase.NewGermanSoloInteractor(gameMock, tpMock)
	assert.Equal(t, germanSoloMockOutput, ci.Bid(domain.GermanSoloBidSolo, -1))
	gameMock.AssertCalled(t, "PlayerBid", domain.GermanSoloBidSolo, -1)
}

func TestGermanSoloInteractor_BidGameEnded(t *testing.T) {
	tpMock := new(presenter.MockGermanSoloPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(germanSoloMockOutput)
	gameMock := new(interfaces.MockGermanSoloGame)
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewGermanSoloInteractor(gameMock, tpMock)
	assert.Equal(t, germanSoloMockOutput, ci.Bid(domain.GermanSoloBidFrage, domain.CardDesignHeart))
	gameMock.AssertNotCalled(t, "PlayerBid", mock.Anything, mock.Anything)
}

func TestGermanSoloInteractor_PlayResolvesTrick(t *testing.T) {
	tpMock := new(presenter.MockGermanSoloPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(germanSoloMockOutput)
	gameMock := new(interfaces.MockGermanSoloGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.GermanSoloPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.GermanSoloPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	ci := usecase.NewGermanSoloInteractor(gameMock, tpMock)
	assert.Equal(t, germanSoloMockOutput, ci.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestGermanSoloInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	tpMock := new(presenter.MockGermanSoloPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(germanSoloMockOutput)
	gameMock := new(interfaces.MockGermanSoloGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.GermanSoloPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	ci := usecase.NewGermanSoloInteractor(gameMock, tpMock)
	assert.Equal(t, germanSoloMockOutput, ci.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestGermanSoloInteractor_PlayError(t *testing.T) {
	tpMock := new(presenter.MockGermanSoloPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(germanSoloMockOutput)
	gameMock := new(interfaces.MockGermanSoloGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.GermanSoloPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	ci := usecase.NewGermanSoloInteractor(gameMock, tpMock)
	assert.Equal(t, germanSoloMockOutput, ci.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestGermanSoloInteractor_PlayNotHumanTurn(t *testing.T) {
	tpMock := new(presenter.MockGermanSoloPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(germanSoloMockOutput)
	gameMock := new(interfaces.MockGermanSoloGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.GermanSoloPhasePlay)
	gameMock.On("IsHumanTurn").Return(false)

	ci := usecase.NewGermanSoloInteractor(gameMock, tpMock)
	assert.Equal(t, germanSoloMockOutput, ci.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestGermanSoloInteractor_NextTrick(t *testing.T) {
	tpMock := new(presenter.MockGermanSoloPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(germanSoloMockOutput)
	gameMock := new(interfaces.MockGermanSoloGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.GermanSoloPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewGermanSoloInteractor(gameMock, tpMock)
	assert.Equal(t, germanSoloMockOutput, ci.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestGermanSoloInteractor_NextRound(t *testing.T) {
	tpMock := new(presenter.MockGermanSoloPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(germanSoloMockOutput)
	gameMock := new(interfaces.MockGermanSoloGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.GermanSoloPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewGermanSoloInteractor(gameMock, tpMock)
	assert.Equal(t, germanSoloMockOutput, ci.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestGermanSoloInteractor_NextRoundGameEnded(t *testing.T) {
	tpMock := new(presenter.MockGermanSoloPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(germanSoloMockOutput)
	gameMock := new(interfaces.MockGermanSoloGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewGermanSoloInteractor(gameMock, tpMock)
	assert.Equal(t, germanSoloMockOutput, ci.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestGermanSoloInteractor_GetConfigHintActionLog(t *testing.T) {
	tpMock := new(presenter.MockGermanSoloPresenter)
	tpMock.On("HintOutput", mock.Anything).Return("hint")
	tpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockGermanSoloGame)
	cfg := domain.DefaultGermanSoloConfig()
	gameMock.On("GetConfig").Return(cfg)

	ci := usecase.NewGermanSoloInteractor(gameMock, tpMock)
	assert.Equal(t, cfg, ci.GetConfig())
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

func TestRestoreGermanSoloInteractor(t *testing.T) {
	tpMock := new(presenter.MockGermanSoloPresenter)
	src := domain.NewDefaultGermanSolo()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ci, err := usecase.RestoreGermanSoloInteractor(data, tpMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)

	_, err = usecase.RestoreGermanSoloInteractor([]byte(`{`), tpMock)
	assert.Error(t, err)
}
