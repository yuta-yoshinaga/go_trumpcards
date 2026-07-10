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

const ombreMockOutput = `{"phase":0}`

func TestNewOmbreInteractor_NilGuards(t *testing.T) {
	tpMock := new(presenter.MockOmbrePresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "OmbreInteractor: g must not be nil", func() {
			usecase.NewOmbreInteractor(nil, tpMock)
		})
	})
	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockOmbreGame)
		assert.PanicsWithValue(t, "OmbreInteractor: tp must not be nil", func() {
			usecase.NewOmbreInteractor(gameMock, nil)
		})
	})
}

func TestOmbreInteractor_Reset(t *testing.T) {
	tpMock := new(presenter.MockOmbrePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ombreMockOutput)
	gameMock := new(interfaces.MockOmbreGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.OmbrePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewOmbreInteractor(gameMock, tpMock)
	assert.Equal(t, ombreMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestOmbreInteractor_ResetWithConfig(t *testing.T) {
	tpMock := new(presenter.MockOmbrePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ombreMockOutput)
	gameMock := new(interfaces.MockOmbreGame)
	cfg := domain.OmbreConfig{CpuDifficulty: domain.OmbreCpuDifficultyHard, TargetRounds: 5}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.OmbrePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewOmbreInteractor(gameMock, tpMock)
	assert.Equal(t, ombreMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestOmbreInteractor_ResetWithConfigInvalid(t *testing.T) {
	tpMock := new(presenter.MockOmbrePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ombreMockOutput)
	gameMock := new(interfaces.MockOmbreGame)

	ci := usecase.NewOmbreInteractor(gameMock, tpMock)
	bad := domain.OmbreConfig{CpuDifficulty: domain.OmbreCpuDifficultyNormal, TargetRounds: 0}
	assert.Equal(t, ombreMockOutput, ci.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestOmbreInteractor_Bid(t *testing.T) {
	tpMock := new(presenter.MockOmbrePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ombreMockOutput)
	gameMock := new(interfaces.MockOmbreGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.OmbrePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerBid", domain.OmbreBidEntrar, domain.CardDesignHeart).Return(nil)

	ci := usecase.NewOmbreInteractor(gameMock, tpMock)
	assert.Equal(t, ombreMockOutput, ci.Bid(domain.OmbreBidEntrar, domain.CardDesignHeart))
	gameMock.AssertCalled(t, "PlayerBid", domain.OmbreBidEntrar, domain.CardDesignHeart)
}

func TestOmbreInteractor_BidError(t *testing.T) {
	tpMock := new(presenter.MockOmbrePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ombreMockOutput)
	gameMock := new(interfaces.MockOmbreGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.OmbreBidSolo, -1).Return(errors.New("cannot bid"))

	ci := usecase.NewOmbreInteractor(gameMock, tpMock)
	assert.Equal(t, ombreMockOutput, ci.Bid(domain.OmbreBidSolo, -1))
	gameMock.AssertCalled(t, "PlayerBid", domain.OmbreBidSolo, -1)
}

func TestOmbreInteractor_BidGameEnded(t *testing.T) {
	tpMock := new(presenter.MockOmbrePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ombreMockOutput)
	gameMock := new(interfaces.MockOmbreGame)
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewOmbreInteractor(gameMock, tpMock)
	assert.Equal(t, ombreMockOutput, ci.Bid(domain.OmbreBidEntrar, domain.CardDesignHeart))
	gameMock.AssertNotCalled(t, "PlayerBid", mock.Anything, mock.Anything)
}

func TestOmbreInteractor_PlayResolvesTrick(t *testing.T) {
	tpMock := new(presenter.MockOmbrePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ombreMockOutput)
	gameMock := new(interfaces.MockOmbreGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.OmbrePhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.OmbrePhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	ci := usecase.NewOmbreInteractor(gameMock, tpMock)
	assert.Equal(t, ombreMockOutput, ci.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestOmbreInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	tpMock := new(presenter.MockOmbrePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ombreMockOutput)
	gameMock := new(interfaces.MockOmbreGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.OmbrePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	ci := usecase.NewOmbreInteractor(gameMock, tpMock)
	assert.Equal(t, ombreMockOutput, ci.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestOmbreInteractor_PlayError(t *testing.T) {
	tpMock := new(presenter.MockOmbrePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ombreMockOutput)
	gameMock := new(interfaces.MockOmbreGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.OmbrePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	ci := usecase.NewOmbreInteractor(gameMock, tpMock)
	assert.Equal(t, ombreMockOutput, ci.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestOmbreInteractor_PlayNotHumanTurn(t *testing.T) {
	tpMock := new(presenter.MockOmbrePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ombreMockOutput)
	gameMock := new(interfaces.MockOmbreGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.OmbrePhasePlay)
	gameMock.On("IsHumanTurn").Return(false)

	ci := usecase.NewOmbreInteractor(gameMock, tpMock)
	assert.Equal(t, ombreMockOutput, ci.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestOmbreInteractor_NextTrick(t *testing.T) {
	tpMock := new(presenter.MockOmbrePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ombreMockOutput)
	gameMock := new(interfaces.MockOmbreGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.OmbrePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewOmbreInteractor(gameMock, tpMock)
	assert.Equal(t, ombreMockOutput, ci.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestOmbreInteractor_NextRound(t *testing.T) {
	tpMock := new(presenter.MockOmbrePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ombreMockOutput)
	gameMock := new(interfaces.MockOmbreGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.OmbrePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewOmbreInteractor(gameMock, tpMock)
	assert.Equal(t, ombreMockOutput, ci.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestOmbreInteractor_NextRoundGameEnded(t *testing.T) {
	tpMock := new(presenter.MockOmbrePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ombreMockOutput)
	gameMock := new(interfaces.MockOmbreGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewOmbreInteractor(gameMock, tpMock)
	assert.Equal(t, ombreMockOutput, ci.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestOmbreInteractor_GetConfigHintActionLog(t *testing.T) {
	tpMock := new(presenter.MockOmbrePresenter)
	tpMock.On("HintOutput", mock.Anything).Return("hint")
	tpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockOmbreGame)
	cfg := domain.DefaultOmbreConfig()
	gameMock.On("GetConfig").Return(cfg)

	ci := usecase.NewOmbreInteractor(gameMock, tpMock)
	assert.Equal(t, cfg, ci.GetConfig())
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

func TestRestoreOmbreInteractor(t *testing.T) {
	tpMock := new(presenter.MockOmbrePresenter)
	src := domain.NewDefaultOmbre()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ci, err := usecase.RestoreOmbreInteractor(data, tpMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)

	_, err = usecase.RestoreOmbreInteractor([]byte(`{`), tpMock)
	assert.Error(t, err)
}
