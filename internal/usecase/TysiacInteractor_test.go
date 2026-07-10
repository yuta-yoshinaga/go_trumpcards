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

const tysiacMockOutput = `{"phase":0}`

func TestNewTysiacInteractor_NilGuards(t *testing.T) {
	tpMock := new(presenter.MockTysiacPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "TysiacInteractor: g must not be nil", func() {
			usecase.NewTysiacInteractor(nil, tpMock)
		})
	})
	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockTysiacGame)
		assert.PanicsWithValue(t, "TysiacInteractor: tp must not be nil", func() {
			usecase.NewTysiacInteractor(gameMock, nil)
		})
	})
}

func TestTysiacInteractor_Reset(t *testing.T) {
	tpMock := new(presenter.MockTysiacPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tysiacMockOutput)
	gameMock := new(interfaces.MockTysiacGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TysiacPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTysiacInteractor(gameMock, tpMock)
	assert.Equal(t, tysiacMockOutput, ti.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestTysiacInteractor_ResetWithConfig(t *testing.T) {
	tpMock := new(presenter.MockTysiacPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tysiacMockOutput)
	gameMock := new(interfaces.MockTysiacGame)
	cfg := domain.TysiacConfig{
		CpuDifficulty: domain.TysiacCpuDifficultyHard,
		TargetPoints:  500,
	}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TysiacPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTysiacInteractor(gameMock, tpMock)
	assert.Equal(t, tysiacMockOutput, ti.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestTysiacInteractor_ResetWithConfigInvalid(t *testing.T) {
	tpMock := new(presenter.MockTysiacPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tysiacMockOutput)
	gameMock := new(interfaces.MockTysiacGame)

	ti := usecase.NewTysiacInteractor(gameMock, tpMock)
	bad := domain.TysiacConfig{
		CpuDifficulty: domain.TysiacCpuDifficultyNormal,
		TargetPoints:  0,
	}
	assert.Equal(t, tysiacMockOutput, ti.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestTysiacInteractor_Bid(t *testing.T) {
	tpMock := new(presenter.MockTysiacPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tysiacMockOutput)
	gameMock := new(interfaces.MockTysiacGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TysiacPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerBid", true).Return(nil)

	ti := usecase.NewTysiacInteractor(gameMock, tpMock)
	assert.Equal(t, tysiacMockOutput, ti.Bid(true))
	gameMock.AssertCalled(t, "PlayerBid", true)
}

func TestTysiacInteractor_BidError(t *testing.T) {
	tpMock := new(presenter.MockTysiacPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tysiacMockOutput)
	gameMock := new(interfaces.MockTysiacGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", false).Return(errors.New("cannot bid"))

	ti := usecase.NewTysiacInteractor(gameMock, tpMock)
	assert.Equal(t, tysiacMockOutput, ti.Bid(false))
	gameMock.AssertCalled(t, "PlayerBid", false)
}

func TestTysiacInteractor_Discard(t *testing.T) {
	tpMock := new(presenter.MockTysiacPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tysiacMockOutput)
	gameMock := new(interfaces.MockTysiacGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TysiacPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerDiscard", 2).Return(nil)

	ti := usecase.NewTysiacInteractor(gameMock, tpMock)
	assert.Equal(t, tysiacMockOutput, ti.Discard(2))
	gameMock.AssertCalled(t, "PlayerDiscard", 2)
}

func TestTysiacInteractor_DiscardError(t *testing.T) {
	tpMock := new(presenter.MockTysiacPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tysiacMockOutput)
	gameMock := new(interfaces.MockTysiacGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerDiscard", 9).Return(errors.New("bad discard"))

	ti := usecase.NewTysiacInteractor(gameMock, tpMock)
	assert.Equal(t, tysiacMockOutput, ti.Discard(9))
	gameMock.AssertCalled(t, "PlayerDiscard", 9)
}

func TestTysiacInteractor_PlayResolvesTrick(t *testing.T) {
	tpMock := new(presenter.MockTysiacPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tysiacMockOutput)
	gameMock := new(interfaces.MockTysiacGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TysiacPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.TysiacPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	ti := usecase.NewTysiacInteractor(gameMock, tpMock)
	assert.Equal(t, tysiacMockOutput, ti.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestTysiacInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	tpMock := new(presenter.MockTysiacPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tysiacMockOutput)
	gameMock := new(interfaces.MockTysiacGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TysiacPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	ti := usecase.NewTysiacInteractor(gameMock, tpMock)
	assert.Equal(t, tysiacMockOutput, ti.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestTysiacInteractor_PlayError(t *testing.T) {
	tpMock := new(presenter.MockTysiacPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tysiacMockOutput)
	gameMock := new(interfaces.MockTysiacGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TysiacPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	ti := usecase.NewTysiacInteractor(gameMock, tpMock)
	assert.Equal(t, tysiacMockOutput, ti.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestTysiacInteractor_PlayNotHumanTurn(t *testing.T) {
	tpMock := new(presenter.MockTysiacPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tysiacMockOutput)
	gameMock := new(interfaces.MockTysiacGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TysiacPhasePlay)
	gameMock.On("IsHumanTurn").Return(false)

	ti := usecase.NewTysiacInteractor(gameMock, tpMock)
	assert.Equal(t, tysiacMockOutput, ti.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestTysiacInteractor_NextTrick(t *testing.T) {
	tpMock := new(presenter.MockTysiacPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tysiacMockOutput)
	gameMock := new(interfaces.MockTysiacGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TysiacPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTysiacInteractor(gameMock, tpMock)
	assert.Equal(t, tysiacMockOutput, ti.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestTysiacInteractor_NextRound(t *testing.T) {
	tpMock := new(presenter.MockTysiacPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tysiacMockOutput)
	gameMock := new(interfaces.MockTysiacGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.TysiacPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTysiacInteractor(gameMock, tpMock)
	assert.Equal(t, tysiacMockOutput, ti.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestTysiacInteractor_NextRoundGameEnded(t *testing.T) {
	tpMock := new(presenter.MockTysiacPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tysiacMockOutput)
	gameMock := new(interfaces.MockTysiacGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ti := usecase.NewTysiacInteractor(gameMock, tpMock)
	assert.Equal(t, tysiacMockOutput, ti.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestTysiacInteractor_GetConfigHintActionLog(t *testing.T) {
	tpMock := new(presenter.MockTysiacPresenter)
	tpMock.On("HintOutput", mock.Anything).Return("hint")
	tpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockTysiacGame)
	cfg := domain.DefaultTysiacConfig()
	gameMock.On("GetConfig").Return(cfg)

	ti := usecase.NewTysiacInteractor(gameMock, tpMock)
	assert.Equal(t, cfg, ti.GetConfig())
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestRestoreTysiacInteractor(t *testing.T) {
	tpMock := new(presenter.MockTysiacPresenter)
	src := domain.NewDefaultTysiac()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ti, err := usecase.RestoreTysiacInteractor(data, tpMock)
	assert.NoError(t, err)
	assert.NotNil(t, ti)

	_, err = usecase.RestoreTysiacInteractor([]byte(`{`), tpMock)
	assert.Error(t, err)
}
