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

const calabresellaMockOutput = `{"phase":0}`

func TestNewCalabresellaInteractor_NilGuards(t *testing.T) {
	tpMock := new(presenter.MockCalabresellaPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "CalabresellaInteractor: g must not be nil", func() {
			usecase.NewCalabresellaInteractor(nil, tpMock)
		})
	})
	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockCalabresellaGame)
		assert.PanicsWithValue(t, "CalabresellaInteractor: tp must not be nil", func() {
			usecase.NewCalabresellaInteractor(gameMock, nil)
		})
	})
}

func TestCalabresellaInteractor_Reset(t *testing.T) {
	tpMock := new(presenter.MockCalabresellaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(calabresellaMockOutput)
	gameMock := new(interfaces.MockCalabresellaGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	// advance: bid loop then play loop. Phase Play + human turn stops both loops immediately.
	gameMock.On("GetPhase").Return(domain.CalabresellaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewCalabresellaInteractor(gameMock, tpMock)
	assert.Equal(t, calabresellaMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestCalabresellaInteractor_ResetWithConfig(t *testing.T) {
	tpMock := new(presenter.MockCalabresellaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(calabresellaMockOutput)
	gameMock := new(interfaces.MockCalabresellaGame)
	cfg := domain.CalabresellaConfig{
		CpuDifficulty: domain.CalabresellaCpuDifficultyHard,
		TargetPoints:  21,
	}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.CalabresellaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewCalabresellaInteractor(gameMock, tpMock)
	assert.Equal(t, calabresellaMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestCalabresellaInteractor_ResetWithConfigInvalid(t *testing.T) {
	tpMock := new(presenter.MockCalabresellaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(calabresellaMockOutput)
	gameMock := new(interfaces.MockCalabresellaGame)

	ci := usecase.NewCalabresellaInteractor(gameMock, tpMock)
	bad := domain.CalabresellaConfig{
		CpuDifficulty: domain.CalabresellaCpuDifficultyNormal,
		TargetPoints:  0,
	}
	assert.Equal(t, calabresellaMockOutput, ci.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestCalabresellaInteractor_Bid(t *testing.T) {
	tpMock := new(presenter.MockCalabresellaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(calabresellaMockOutput)
	gameMock := new(interfaces.MockCalabresellaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.CalabresellaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerBid", domain.CalabresellaBidChiamo).Return(nil)

	ci := usecase.NewCalabresellaInteractor(gameMock, tpMock)
	assert.Equal(t, calabresellaMockOutput, ci.Bid(domain.CalabresellaBidChiamo))
	gameMock.AssertCalled(t, "PlayerBid", domain.CalabresellaBidChiamo)
}

func TestCalabresellaInteractor_BidError(t *testing.T) {
	tpMock := new(presenter.MockCalabresellaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(calabresellaMockOutput)
	gameMock := new(interfaces.MockCalabresellaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.CalabresellaBidSolo).Return(errors.New("cannot bid"))

	ci := usecase.NewCalabresellaInteractor(gameMock, tpMock)
	assert.Equal(t, calabresellaMockOutput, ci.Bid(domain.CalabresellaBidSolo))
	gameMock.AssertCalled(t, "PlayerBid", domain.CalabresellaBidSolo)
}

func TestCalabresellaInteractor_Discard(t *testing.T) {
	tpMock := new(presenter.MockCalabresellaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(calabresellaMockOutput)
	gameMock := new(interfaces.MockCalabresellaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.CalabresellaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerDiscard", 2).Return(nil)

	ci := usecase.NewCalabresellaInteractor(gameMock, tpMock)
	assert.Equal(t, calabresellaMockOutput, ci.Discard(2))
	gameMock.AssertCalled(t, "PlayerDiscard", 2)
}

func TestCalabresellaInteractor_DiscardError(t *testing.T) {
	tpMock := new(presenter.MockCalabresellaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(calabresellaMockOutput)
	gameMock := new(interfaces.MockCalabresellaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerDiscard", 9).Return(errors.New("bad discard"))

	ci := usecase.NewCalabresellaInteractor(gameMock, tpMock)
	assert.Equal(t, calabresellaMockOutput, ci.Discard(9))
	gameMock.AssertCalled(t, "PlayerDiscard", 9)
}

func TestCalabresellaInteractor_PlayResolvesTrick(t *testing.T) {
	tpMock := new(presenter.MockCalabresellaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(calabresellaMockOutput)
	gameMock := new(interfaces.MockCalabresellaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	// First GetPhase (in Play) sees TrickEnd so ResolveTrick fires; then RoundEnd stops advance.
	gameMock.On("GetPhase").Return(domain.CalabresellaPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.CalabresellaPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	ci := usecase.NewCalabresellaInteractor(gameMock, tpMock)
	assert.Equal(t, calabresellaMockOutput, ci.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestCalabresellaInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	tpMock := new(presenter.MockCalabresellaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(calabresellaMockOutput)
	gameMock := new(interfaces.MockCalabresellaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.CalabresellaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	ci := usecase.NewCalabresellaInteractor(gameMock, tpMock)
	assert.Equal(t, calabresellaMockOutput, ci.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestCalabresellaInteractor_PlayError(t *testing.T) {
	tpMock := new(presenter.MockCalabresellaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(calabresellaMockOutput)
	gameMock := new(interfaces.MockCalabresellaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.CalabresellaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	ci := usecase.NewCalabresellaInteractor(gameMock, tpMock)
	assert.Equal(t, calabresellaMockOutput, ci.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestCalabresellaInteractor_PlayNotHumanTurn(t *testing.T) {
	tpMock := new(presenter.MockCalabresellaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(calabresellaMockOutput)
	gameMock := new(interfaces.MockCalabresellaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.CalabresellaPhasePlay)
	gameMock.On("IsHumanTurn").Return(false)

	ci := usecase.NewCalabresellaInteractor(gameMock, tpMock)
	assert.Equal(t, calabresellaMockOutput, ci.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestCalabresellaInteractor_NextTrick(t *testing.T) {
	tpMock := new(presenter.MockCalabresellaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(calabresellaMockOutput)
	gameMock := new(interfaces.MockCalabresellaGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.CalabresellaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewCalabresellaInteractor(gameMock, tpMock)
	assert.Equal(t, calabresellaMockOutput, ci.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestCalabresellaInteractor_NextRound(t *testing.T) {
	tpMock := new(presenter.MockCalabresellaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(calabresellaMockOutput)
	gameMock := new(interfaces.MockCalabresellaGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.CalabresellaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewCalabresellaInteractor(gameMock, tpMock)
	assert.Equal(t, calabresellaMockOutput, ci.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestCalabresellaInteractor_NextRoundGameEnded(t *testing.T) {
	tpMock := new(presenter.MockCalabresellaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(calabresellaMockOutput)
	gameMock := new(interfaces.MockCalabresellaGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewCalabresellaInteractor(gameMock, tpMock)
	assert.Equal(t, calabresellaMockOutput, ci.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestCalabresellaInteractor_GetConfigHintActionLog(t *testing.T) {
	tpMock := new(presenter.MockCalabresellaPresenter)
	tpMock.On("HintOutput", mock.Anything).Return("hint")
	tpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockCalabresellaGame)
	cfg := domain.DefaultCalabresellaConfig()
	gameMock.On("GetConfig").Return(cfg)

	ci := usecase.NewCalabresellaInteractor(gameMock, tpMock)
	assert.Equal(t, cfg, ci.GetConfig())
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

func TestRestoreCalabresellaInteractor(t *testing.T) {
	tpMock := new(presenter.MockCalabresellaPresenter)
	src := domain.NewDefaultCalabresella()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ci, err := usecase.RestoreCalabresellaInteractor(data, tpMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)

	_, err = usecase.RestoreCalabresellaInteractor([]byte(`{`), tpMock)
	assert.Error(t, err)
}
