//go:build test && (!js || !wasm || extra4)

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

func TestNewPutInteractor_NilGuards(t *testing.T) {
	tpMock := new(presenter.MockPutPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "PutInteractor: g must not be nil", func() {
			usecase.NewPutInteractor(nil, tpMock)
		})
	})

	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockPutGame)
		assert.PanicsWithValue(t, "PutInteractor: tp must not be nil", func() {
			usecase.NewPutInteractor(gameMock, nil)
		})
	})
}

func TestPutInteractor_Reset(t *testing.T) {
	out := `{"phase":0}`
	tpMock := new(presenter.MockPutPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(out)
	gameMock := new(interfaces.MockPutGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.PutPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewPutInteractor(gameMock, tpMock)
	assert.Equal(t, out, ti.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestPutInteractor_ResetWithConfig(t *testing.T) {
	out := `{"phase":0}`
	tpMock := new(presenter.MockPutPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(out)
	gameMock := new(interfaces.MockPutGame)
	cfg := domain.DefaultPutConfig()
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.PutPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewPutInteractor(gameMock, tpMock)
	assert.Equal(t, out, ti.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestPutInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	tpMock := new(presenter.MockPutPresenter)
	gameMock := new(interfaces.MockPutGame)
	tpMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

	ti := usecase.NewPutInteractor(gameMock, tpMock)
	got := ti.ResetWithConfig(domain.PutConfig{MatchTarget: 0})
	assert.Equal(t, "validation error", got)
	gameMock.AssertNotCalled(t, "Reset")
}

func TestPutInteractor_Play_Valid(t *testing.T) {
	out := `{"phase":0}`
	tpMock := new(presenter.MockPutPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(out)
	gameMock := new(interfaces.MockPutGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)
	gameMock.On("GetPhase").Return(domain.PutPhasePlay)

	ti := usecase.NewPutInteractor(gameMock, tpMock)
	assert.Equal(t, out, ti.Play(1))
	gameMock.AssertCalled(t, "PlayerPlay", 1)
}

func TestPutInteractor_Play_Error(t *testing.T) {
	wantErr := errors.New("boom")
	tpMock := new(presenter.MockPutPresenter)
	tpMock.On("Output", mock.Anything, wantErr).Return("error output")
	gameMock := new(interfaces.MockPutGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 99).Return(wantErr)

	ti := usecase.NewPutInteractor(gameMock, tpMock)
	assert.Equal(t, "error output", ti.Play(99))
}

func TestPutInteractor_Play_GuardBlocksWhenGameEnded(t *testing.T) {
	tpMock := new(presenter.MockPutPresenter)
	gameMock := new(interfaces.MockPutGame)
	gameMock.On("GetGameEndFlag").Return(true)
	tpMock.On("Output", gameMock, nil).Return("game ended")

	ti := usecase.NewPutInteractor(gameMock, tpMock)
	assert.Equal(t, "game ended", ti.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay")
}

func TestPutInteractor_Put_Valid(t *testing.T) {
	out := `{"phase":1}`
	tpMock := new(presenter.MockPutPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(out)
	gameMock := new(interfaces.MockPutGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("DeclarePut").Return(nil)
	gameMock.On("GetPhase").Return(domain.PutPhaseRespond)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewPutInteractor(gameMock, tpMock)
	assert.Equal(t, out, ti.Put())
	gameMock.AssertCalled(t, "DeclarePut")
}

func TestPutInteractor_Put_Error(t *testing.T) {
	wantErr := errors.New("cannot call")
	tpMock := new(presenter.MockPutPresenter)
	tpMock.On("Output", mock.Anything, wantErr).Return("err")
	gameMock := new(interfaces.MockPutGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("DeclarePut").Return(wantErr)

	ti := usecase.NewPutInteractor(gameMock, tpMock)
	assert.Equal(t, "err", ti.Put())
}

func TestPutInteractor_Put_BlocksWhenEnded(t *testing.T) {
	tpMock := new(presenter.MockPutPresenter)
	gameMock := new(interfaces.MockPutGame)
	gameMock.On("GetGameEndFlag").Return(true)
	tpMock.On("Output", gameMock, nil).Return("ended")

	ti := usecase.NewPutInteractor(gameMock, tpMock)
	assert.Equal(t, "ended", ti.Put())
	gameMock.AssertNotCalled(t, "DeclarePut")
}

func TestPutInteractor_Respond_Valid(t *testing.T) {
	out := `{"phase":0}`
	tpMock := new(presenter.MockPutPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(out)
	gameMock := new(interfaces.MockPutGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("RespondPut", true).Return(nil)
	gameMock.On("GetPhase").Return(domain.PutPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewPutInteractor(gameMock, tpMock)
	assert.Equal(t, out, ti.Respond(true))
	gameMock.AssertCalled(t, "RespondPut", true)
}

func TestPutInteractor_Respond_Error(t *testing.T) {
	wantErr := errors.New("bad respond")
	tpMock := new(presenter.MockPutPresenter)
	tpMock.On("Output", mock.Anything, wantErr).Return("err")
	gameMock := new(interfaces.MockPutGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("RespondPut", false).Return(wantErr)

	ti := usecase.NewPutInteractor(gameMock, tpMock)
	assert.Equal(t, "err", ti.Respond(false))
}

func TestPutInteractor_Next(t *testing.T) {
	out := `{"phase":0}`
	tpMock := new(presenter.MockPutPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(out)
	gameMock := new(interfaces.MockPutGame)
	gameMock.On("Next").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.PutPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewPutInteractor(gameMock, tpMock)
	assert.Equal(t, out, ti.Next())
	gameMock.AssertCalled(t, "Next")
}

func TestPutInteractor_Next_BlocksOnGameEnd(t *testing.T) {
	tpMock := new(presenter.MockPutPresenter)
	gameMock := new(interfaces.MockPutGame)
	gameMock.On("Next").Return()
	gameMock.On("GetGameEndFlag").Return(true)
	tpMock.On("Output", gameMock, nil).Return("ended")

	ti := usecase.NewPutInteractor(gameMock, tpMock)
	assert.Equal(t, "ended", ti.Next())
}

func TestPutInteractor_GetConfig(t *testing.T) {
	tpMock := new(presenter.MockPutPresenter)
	gameMock := new(interfaces.MockPutGame)
	cfg := domain.DefaultPutConfig()
	gameMock.On("GetConfig").Return(cfg)

	ti := usecase.NewPutInteractor(gameMock, tpMock)
	assert.Equal(t, cfg, ti.GetConfig())
}

func TestPutInteractor_HintAndActionLog(t *testing.T) {
	tpMock := new(presenter.MockPutPresenter)
	gameMock := new(interfaces.MockPutGame)
	tpMock.On("HintOutput", gameMock).Return("hint")
	tpMock.On("ActionLogOutput", gameMock).Return("log")

	ti := usecase.NewPutInteractor(gameMock, tpMock)
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestPutInteractor_Snapshot_RoundtripsViaRealGame(t *testing.T) {
	tpMock := new(presenter.MockPutPresenter)
	game := domain.NewDefaultPut()
	game.Reset()

	ti := usecase.NewPutInteractor(game, tpMock)
	data, err := ti.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	ti2, err := usecase.RestorePutInteractor(data, tpMock)
	assert.NoError(t, err)
	assert.Equal(t, game.GetPhase(), ti2.Game.GetPhase())
}
