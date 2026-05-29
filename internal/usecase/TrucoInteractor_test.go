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

func TestNewTrucoInteractor_NilGuards(t *testing.T) {
	tpMock := new(presenter.MockTrucoPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "TrucoInteractor: g must not be nil", func() {
			usecase.NewTrucoInteractor(nil, tpMock)
		})
	})

	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockTrucoGame)
		assert.PanicsWithValue(t, "TrucoInteractor: tp must not be nil", func() {
			usecase.NewTrucoInteractor(gameMock, nil)
		})
	})
}

func TestTrucoInteractor_Reset(t *testing.T) {
	out := `{"phase":0}`
	tpMock := new(presenter.MockTrucoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(out)
	gameMock := new(interfaces.MockTrucoGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TrucoPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTrucoInteractor(gameMock, tpMock)
	assert.Equal(t, out, ti.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestTrucoInteractor_ResetWithConfig(t *testing.T) {
	out := `{"phase":0}`
	tpMock := new(presenter.MockTrucoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(out)
	gameMock := new(interfaces.MockTrucoGame)
	cfg := domain.DefaultTrucoConfig()
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TrucoPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTrucoInteractor(gameMock, tpMock)
	assert.Equal(t, out, ti.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestTrucoInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	tpMock := new(presenter.MockTrucoPresenter)
	gameMock := new(interfaces.MockTrucoGame)
	tpMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

	ti := usecase.NewTrucoInteractor(gameMock, tpMock)
	got := ti.ResetWithConfig(domain.TrucoConfig{MatchTarget: 0})
	assert.Equal(t, "validation error", got)
	gameMock.AssertNotCalled(t, "Reset")
}

func TestTrucoInteractor_Play_Valid(t *testing.T) {
	out := `{"phase":0}`
	tpMock := new(presenter.MockTrucoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(out)
	gameMock := new(interfaces.MockTrucoGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)
	gameMock.On("GetPhase").Return(domain.TrucoPhasePlay)

	ti := usecase.NewTrucoInteractor(gameMock, tpMock)
	assert.Equal(t, out, ti.Play(1))
	gameMock.AssertCalled(t, "PlayerPlay", 1)
}

func TestTrucoInteractor_Play_Error(t *testing.T) {
	wantErr := errors.New("boom")
	tpMock := new(presenter.MockTrucoPresenter)
	tpMock.On("Output", mock.Anything, wantErr).Return("error output")
	gameMock := new(interfaces.MockTrucoGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 99).Return(wantErr)

	ti := usecase.NewTrucoInteractor(gameMock, tpMock)
	assert.Equal(t, "error output", ti.Play(99))
}

func TestTrucoInteractor_Play_GuardBlocksWhenGameEnded(t *testing.T) {
	tpMock := new(presenter.MockTrucoPresenter)
	gameMock := new(interfaces.MockTrucoGame)
	gameMock.On("GetGameEndFlag").Return(true)
	tpMock.On("Output", gameMock, nil).Return("game ended")

	ti := usecase.NewTrucoInteractor(gameMock, tpMock)
	assert.Equal(t, "game ended", ti.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay")
}

func TestTrucoInteractor_Truco_Valid(t *testing.T) {
	out := `{"phase":1}`
	tpMock := new(presenter.MockTrucoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(out)
	gameMock := new(interfaces.MockTrucoGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("DeclareTruco").Return(nil)
	gameMock.On("GetPhase").Return(domain.TrucoPhaseRespond)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTrucoInteractor(gameMock, tpMock)
	assert.Equal(t, out, ti.Truco())
	gameMock.AssertCalled(t, "DeclareTruco")
}

func TestTrucoInteractor_Truco_Error(t *testing.T) {
	wantErr := errors.New("cannot call")
	tpMock := new(presenter.MockTrucoPresenter)
	tpMock.On("Output", mock.Anything, wantErr).Return("err")
	gameMock := new(interfaces.MockTrucoGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("DeclareTruco").Return(wantErr)

	ti := usecase.NewTrucoInteractor(gameMock, tpMock)
	assert.Equal(t, "err", ti.Truco())
}

func TestTrucoInteractor_Truco_BlocksWhenEnded(t *testing.T) {
	tpMock := new(presenter.MockTrucoPresenter)
	gameMock := new(interfaces.MockTrucoGame)
	gameMock.On("GetGameEndFlag").Return(true)
	tpMock.On("Output", gameMock, nil).Return("ended")

	ti := usecase.NewTrucoInteractor(gameMock, tpMock)
	assert.Equal(t, "ended", ti.Truco())
	gameMock.AssertNotCalled(t, "DeclareTruco")
}

func TestTrucoInteractor_Respond_Valid(t *testing.T) {
	out := `{"phase":0}`
	tpMock := new(presenter.MockTrucoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(out)
	gameMock := new(interfaces.MockTrucoGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("RespondTruco", true).Return(nil)
	gameMock.On("GetPhase").Return(domain.TrucoPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTrucoInteractor(gameMock, tpMock)
	assert.Equal(t, out, ti.Respond(true))
	gameMock.AssertCalled(t, "RespondTruco", true)
}

func TestTrucoInteractor_Respond_Error(t *testing.T) {
	wantErr := errors.New("bad respond")
	tpMock := new(presenter.MockTrucoPresenter)
	tpMock.On("Output", mock.Anything, wantErr).Return("err")
	gameMock := new(interfaces.MockTrucoGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("RespondTruco", false).Return(wantErr)

	ti := usecase.NewTrucoInteractor(gameMock, tpMock)
	assert.Equal(t, "err", ti.Respond(false))
}

func TestTrucoInteractor_Next(t *testing.T) {
	out := `{"phase":0}`
	tpMock := new(presenter.MockTrucoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(out)
	gameMock := new(interfaces.MockTrucoGame)
	gameMock.On("Next").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TrucoPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTrucoInteractor(gameMock, tpMock)
	assert.Equal(t, out, ti.Next())
	gameMock.AssertCalled(t, "Next")
}

func TestTrucoInteractor_Next_BlocksOnGameEnd(t *testing.T) {
	tpMock := new(presenter.MockTrucoPresenter)
	gameMock := new(interfaces.MockTrucoGame)
	gameMock.On("Next").Return()
	gameMock.On("GetGameEndFlag").Return(true)
	tpMock.On("Output", gameMock, nil).Return("ended")

	ti := usecase.NewTrucoInteractor(gameMock, tpMock)
	assert.Equal(t, "ended", ti.Next())
}

func TestTrucoInteractor_GetConfig(t *testing.T) {
	tpMock := new(presenter.MockTrucoPresenter)
	gameMock := new(interfaces.MockTrucoGame)
	cfg := domain.DefaultTrucoConfig()
	gameMock.On("GetConfig").Return(cfg)

	ti := usecase.NewTrucoInteractor(gameMock, tpMock)
	assert.Equal(t, cfg, ti.GetConfig())
}

func TestTrucoInteractor_HintAndActionLog(t *testing.T) {
	tpMock := new(presenter.MockTrucoPresenter)
	gameMock := new(interfaces.MockTrucoGame)
	tpMock.On("HintOutput", gameMock).Return("hint")
	tpMock.On("ActionLogOutput", gameMock).Return("log")

	ti := usecase.NewTrucoInteractor(gameMock, tpMock)
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestTrucoInteractor_Snapshot_RoundtripsViaRealGame(t *testing.T) {
	tpMock := new(presenter.MockTrucoPresenter)
	game := domain.NewDefaultTruco()
	game.Reset()

	ti := usecase.NewTrucoInteractor(game, tpMock)
	data, err := ti.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	ti2, err := usecase.RestoreTrucoInteractor(data, tpMock)
	assert.NoError(t, err)
	assert.Equal(t, game.GetPhase(), ti2.Game.GetPhase())
}
