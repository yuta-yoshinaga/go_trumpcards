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

const ultiMockOutput = `{"phase":0}`

func newUltiPlayMock() *interfaces.MockUltiGame {
	m := new(interfaces.MockUltiGame)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.UltiPhasePlay)
	m.On("IsHumanTurn").Return(true)
	m.On("IsHumanBidTurn").Return(false)
	return m
}

func TestNewUltiInteractor_NilGuards(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "UltiInteractor: g must not be nil", func() {
			usecase.NewUltiInteractor(nil, tpMock)
		})
	})
	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockUltiGame)
		assert.PanicsWithValue(t, "UltiInteractor: tp must not be nil", func() {
			usecase.NewUltiInteractor(gameMock, nil)
		})
	})
}

func TestUltiInteractor_Reset(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ultiMockOutput)
	gameMock := newUltiPlayMock()
	gameMock.On("Reset").Return()

	ci := usecase.NewUltiInteractor(gameMock, tpMock)
	assert.Equal(t, ultiMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestUltiInteractor_ResetWithConfig(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ultiMockOutput)
	gameMock := newUltiPlayMock()
	cfg := domain.UltiConfig{CpuDifficulty: domain.UltiCpuDifficultyHard, TargetRounds: 5}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()

	ci := usecase.NewUltiInteractor(gameMock, tpMock)
	assert.Equal(t, ultiMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestUltiInteractor_ResetWithConfigInvalid(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ultiMockOutput)
	gameMock := new(interfaces.MockUltiGame)

	ci := usecase.NewUltiInteractor(gameMock, tpMock)
	bad := domain.UltiConfig{CpuDifficulty: domain.UltiCpuDifficultyNormal, TargetRounds: 0}
	assert.Equal(t, ultiMockOutput, ci.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestUltiInteractor_Bid(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ultiMockOutput)
	gameMock := newUltiPlayMock()
	gameMock.On("PlayerBid", domain.UltiContractParty, domain.CardDesignHeart).Return(nil)

	ci := usecase.NewUltiInteractor(gameMock, tpMock)
	assert.Equal(t, ultiMockOutput, ci.Bid(domain.UltiContractParty, domain.CardDesignHeart))
	gameMock.AssertCalled(t, "PlayerBid", domain.UltiContractParty, domain.CardDesignHeart)
}

func TestUltiInteractor_BidError(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ultiMockOutput)
	gameMock := new(interfaces.MockUltiGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.UltiContractBetli, -1).Return(errors.New("cannot bid"))

	ci := usecase.NewUltiInteractor(gameMock, tpMock)
	assert.Equal(t, ultiMockOutput, ci.Bid(domain.UltiContractBetli, -1))
	gameMock.AssertCalled(t, "PlayerBid", domain.UltiContractBetli, -1)
}

func TestUltiInteractor_BidGameEnded(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ultiMockOutput)
	gameMock := new(interfaces.MockUltiGame)
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewUltiInteractor(gameMock, tpMock)
	assert.Equal(t, ultiMockOutput, ci.Bid(domain.UltiContractParty, domain.CardDesignHeart))
	gameMock.AssertNotCalled(t, "PlayerBid", mock.Anything, mock.Anything)
}

func TestUltiInteractor_Discard(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ultiMockOutput)
	gameMock := newUltiPlayMock()
	gameMock.On("PlayerDiscard", []int{0, 1}).Return(nil)

	ci := usecase.NewUltiInteractor(gameMock, tpMock)
	assert.Equal(t, ultiMockOutput, ci.Discard([]int{0, 1}))
	gameMock.AssertCalled(t, "PlayerDiscard", []int{0, 1})
}

func TestUltiInteractor_DiscardError(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ultiMockOutput)
	gameMock := new(interfaces.MockUltiGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerDiscard", []int{0}).Return(errors.New("bad discard"))

	ci := usecase.NewUltiInteractor(gameMock, tpMock)
	assert.Equal(t, ultiMockOutput, ci.Discard([]int{0}))
	gameMock.AssertCalled(t, "PlayerDiscard", []int{0})
}

func TestUltiInteractor_DiscardGameEnded(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ultiMockOutput)
	gameMock := new(interfaces.MockUltiGame)
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewUltiInteractor(gameMock, tpMock)
	assert.Equal(t, ultiMockOutput, ci.Discard([]int{0, 1}))
	gameMock.AssertNotCalled(t, "PlayerDiscard", mock.Anything)
}

func TestUltiInteractor_PlayResolvesTrick(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ultiMockOutput)
	gameMock := new(interfaces.MockUltiGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.UltiPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.UltiPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	ci := usecase.NewUltiInteractor(gameMock, tpMock)
	assert.Equal(t, ultiMockOutput, ci.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestUltiInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ultiMockOutput)
	gameMock := newUltiPlayMock()
	gameMock.On("PlayerPlay", 1).Return(nil)

	ci := usecase.NewUltiInteractor(gameMock, tpMock)
	assert.Equal(t, ultiMockOutput, ci.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestUltiInteractor_PlayError(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ultiMockOutput)
	gameMock := new(interfaces.MockUltiGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.UltiPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	ci := usecase.NewUltiInteractor(gameMock, tpMock)
	assert.Equal(t, ultiMockOutput, ci.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestUltiInteractor_PlayNotHumanTurn(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ultiMockOutput)
	gameMock := new(interfaces.MockUltiGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.UltiPhasePlay)
	gameMock.On("IsHumanTurn").Return(false)

	ci := usecase.NewUltiInteractor(gameMock, tpMock)
	assert.Equal(t, ultiMockOutput, ci.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestUltiInteractor_NextTrick(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ultiMockOutput)
	gameMock := newUltiPlayMock()
	gameMock.On("NextTrick").Return()

	ci := usecase.NewUltiInteractor(gameMock, tpMock)
	assert.Equal(t, ultiMockOutput, ci.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestUltiInteractor_NextRound(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ultiMockOutput)
	gameMock := newUltiPlayMock()
	gameMock.On("ScoreRound").Return()
	gameMock.On("NextRound").Return()

	ci := usecase.NewUltiInteractor(gameMock, tpMock)
	assert.Equal(t, ultiMockOutput, ci.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestUltiInteractor_NextRoundGameEnded(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(ultiMockOutput)
	gameMock := new(interfaces.MockUltiGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewUltiInteractor(gameMock, tpMock)
	assert.Equal(t, ultiMockOutput, ci.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestUltiInteractor_GetConfigHintActionLog(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)
	tpMock.On("HintOutput", mock.Anything).Return("hint")
	tpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockUltiGame)
	cfg := domain.DefaultUltiConfig()
	gameMock.On("GetConfig").Return(cfg)

	ci := usecase.NewUltiInteractor(gameMock, tpMock)
	assert.Equal(t, cfg, ci.GetConfig())
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

func TestRestoreUltiInteractor(t *testing.T) {
	tpMock := new(presenter.MockUltiPresenter)
	src := domain.NewDefaultUlti()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ci, err := usecase.RestoreUltiInteractor(data, tpMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)

	_, err = usecase.RestoreUltiInteractor([]byte(`{`), tpMock)
	assert.Error(t, err)
}
