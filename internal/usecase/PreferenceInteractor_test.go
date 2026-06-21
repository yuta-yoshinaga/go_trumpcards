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

const preferenceMockOutput = `{"phase":0}`

func newPreferenceMocks() (*presenter.MockPreferencePresenter, *interfaces.MockPreferenceGame) {
	sp := new(presenter.MockPreferencePresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(preferenceMockOutput)
	sp.On("HintOutput", mock.Anything).Return("hint")
	sp.On("ActionLogOutput", mock.Anything).Return("log")
	return sp, new(interfaces.MockPreferenceGame)
}

func TestNewPreferenceInteractor_NilGuards(t *testing.T) {
	sp := new(presenter.MockPreferencePresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "PreferenceInteractor: g must not be nil", func() {
			usecase.NewPreferenceInteractor(nil, sp)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockPreferenceGame)
		assert.PanicsWithValue(t, "PreferenceInteractor: sp must not be nil", func() {
			usecase.NewPreferenceInteractor(gameMock, nil)
		})
	})
}

func TestPreferenceInteractor_Reset(t *testing.T) {
	sp, gameMock := newPreferenceMocks()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.PreferencePhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	pi := usecase.NewPreferenceInteractor(gameMock, sp)
	assert.Equal(t, preferenceMockOutput, pi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestPreferenceInteractor_ResetWithConfig(t *testing.T) {
	sp, gameMock := newPreferenceMocks()
	cfg := domain.PreferenceConfig{CpuDifficulty: domain.PreferenceCpuDifficultyHard, TargetPoints: 15}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.PreferencePhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	pi := usecase.NewPreferenceInteractor(gameMock, sp)
	assert.Equal(t, preferenceMockOutput, pi.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestPreferenceInteractor_ResetWithConfigInvalid(t *testing.T) {
	sp, gameMock := newPreferenceMocks()

	pi := usecase.NewPreferenceInteractor(gameMock, sp)
	bad := domain.PreferenceConfig{CpuDifficulty: domain.PreferenceCpuDifficultyNormal, TargetPoints: 0}
	assert.Equal(t, preferenceMockOutput, pi.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestPreferenceInteractor_Bid(t *testing.T) {
	sp, gameMock := newPreferenceMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.PreferenceBidSix).Return(nil)
	gameMock.On("GetPhase").Return(domain.PreferencePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	pi := usecase.NewPreferenceInteractor(gameMock, sp)
	assert.Equal(t, preferenceMockOutput, pi.Bid(int(domain.PreferenceBidSix)))
	gameMock.AssertCalled(t, "PlayerBid", domain.PreferenceBidSix)
}

func TestPreferenceInteractor_BidError(t *testing.T) {
	sp, gameMock := newPreferenceMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.PreferenceBidSix).Return(errors.New("must beat current bid"))

	pi := usecase.NewPreferenceInteractor(gameMock, sp)
	assert.Equal(t, preferenceMockOutput, pi.Bid(int(domain.PreferenceBidSix)))
	gameMock.AssertNotCalled(t, "CpuBid")
}

func TestPreferenceInteractor_BidGameEnded(t *testing.T) {
	sp, gameMock := newPreferenceMocks()
	gameMock.On("GetGameEndFlag").Return(true)

	pi := usecase.NewPreferenceInteractor(gameMock, sp)
	assert.Equal(t, preferenceMockOutput, pi.Bid(0))
	gameMock.AssertNotCalled(t, "PlayerBid", mock.Anything)
}

func TestPreferenceInteractor_PlayResolvesTrick(t *testing.T) {
	sp, gameMock := newPreferenceMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.PreferencePhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.PreferencePhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	pi := usecase.NewPreferenceInteractor(gameMock, sp)
	assert.Equal(t, preferenceMockOutput, pi.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestPreferenceInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	sp, gameMock := newPreferenceMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.PreferencePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	pi := usecase.NewPreferenceInteractor(gameMock, sp)
	assert.Equal(t, preferenceMockOutput, pi.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestPreferenceInteractor_PlayError(t *testing.T) {
	sp, gameMock := newPreferenceMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.PreferencePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	pi := usecase.NewPreferenceInteractor(gameMock, sp)
	assert.Equal(t, preferenceMockOutput, pi.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestPreferenceInteractor_PlayNotHumanTurn(t *testing.T) {
	sp, gameMock := newPreferenceMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	pi := usecase.NewPreferenceInteractor(gameMock, sp)
	assert.Equal(t, preferenceMockOutput, pi.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestPreferenceInteractor_NextTrick(t *testing.T) {
	sp, gameMock := newPreferenceMocks()
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.PreferencePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	pi := usecase.NewPreferenceInteractor(gameMock, sp)
	assert.Equal(t, preferenceMockOutput, pi.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestPreferenceInteractor_NextRound(t *testing.T) {
	sp, gameMock := newPreferenceMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.PreferencePhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	pi := usecase.NewPreferenceInteractor(gameMock, sp)
	assert.Equal(t, preferenceMockOutput, pi.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestPreferenceInteractor_NextRoundGameEnded(t *testing.T) {
	sp, gameMock := newPreferenceMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	pi := usecase.NewPreferenceInteractor(gameMock, sp)
	assert.Equal(t, preferenceMockOutput, pi.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestPreferenceInteractor_GetConfigHintActionLog(t *testing.T) {
	sp, gameMock := newPreferenceMocks()
	cfg := domain.DefaultPreferenceConfig()
	gameMock.On("GetConfig").Return(cfg)

	pi := usecase.NewPreferenceInteractor(gameMock, sp)
	assert.Equal(t, cfg, pi.GetConfig())
	assert.Equal(t, "hint", pi.Hint())
	assert.Equal(t, "log", pi.ActionLog())
}

func TestPreferenceInteractor_AdvanceCpuBids(t *testing.T) {
	sp, gameMock := newPreferenceMocks()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	// Two CPU bids, then the human's bid turn.
	gameMock.On("GetPhase").Return(domain.PreferencePhaseBid)
	gameMock.On("IsHumanBidTurn").Return(false).Twice()
	gameMock.On("IsHumanBidTurn").Return(true)
	gameMock.On("CpuBid").Return()

	pi := usecase.NewPreferenceInteractor(gameMock, sp)
	assert.Equal(t, preferenceMockOutput, pi.Reset())
	gameMock.AssertNumberOfCalls(t, "CpuBid", 2)
}

func TestRestorePreferenceInteractor(t *testing.T) {
	sp := new(presenter.MockPreferencePresenter)
	src := domain.NewDefaultPreference()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	pi, err := usecase.RestorePreferenceInteractor(data, sp)
	assert.NoError(t, err)
	assert.NotNil(t, pi)

	_, err = usecase.RestorePreferenceInteractor([]byte(`{`), sp)
	assert.Error(t, err)
}
