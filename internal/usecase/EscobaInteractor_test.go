package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func escobaNewTestGame() *domain.Escoba {
	return domain.NewDefaultEscoba()
}

func TestNewEscobaInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockEscobaPresenter)
	t.Run("panics when eg is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "EscobaInteractor: eg must not be nil", func() {
			usecase.NewEscobaInteractor(nil, spMock)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "EscobaInteractor: sp must not be nil", func() {
			usecase.NewEscobaInteractor(escobaNewTestGame(), nil)
		})
	})
}

func TestEscobaInteractor_Methods(t *testing.T) {
	mockOutput := `{"players":[]}`
	spMock := new(presenter.MockEscobaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	ei := usecase.NewEscobaInteractor(escobaNewTestGame(), spMock)

	t.Run("Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, ei.Reset())
	})

	t.Run("Play returns output", func(t *testing.T) {
		assert.Equal(t, mockOutput, ei.Play(0, nil))
	})

	t.Run("ResetWithConfig valid", func(t *testing.T) {
		cfg := domain.DefaultEscobaConfig()
		assert.Equal(t, mockOutput, ei.ResetWithConfig(cfg))
	})

	t.Run("GetConfig returns config", func(t *testing.T) {
		cfg := ei.GetConfig()
		assert.Equal(t, domain.EscobaDefaultTargetScore, cfg.TargetScore)
	})
}

func TestEscobaInteractor_ActionLog(t *testing.T) {
	spMock := new(presenter.MockEscobaPresenter)
	gameMock := new(interfaces.MockEscobaGame)
	spMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)
	ei := usecase.NewEscobaInteractor(gameMock, spMock)
	assert.Equal(t, `{"entries":[]}`, ei.ActionLog())
	spMock.AssertExpectations(t)
}

func TestEscobaInteractor_MockGame(t *testing.T) {
	mockOutput := `{"players":[]}`
	spMock := new(presenter.MockEscobaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)

	gameMock := new(interfaces.MockEscobaGame)
	gameMock.On("Reset").Return()
	gameMock.On("NextRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetPhase").Return(domain.EscobaPhasePlayerTurn)
	gameMock.On("CpuPlay").Return()
	gameMock.On("PlayerPlay", mock.Anything, mock.Anything).Return(nil)
	gameMock.On("SetConfig", mock.Anything).Return()
	gameMock.On("GetConfig").Return(domain.DefaultEscobaConfig())

	ei := usecase.NewEscobaInteractor(gameMock, spMock)

	t.Run("Reset delegates", func(t *testing.T) {
		assert.Equal(t, mockOutput, ei.Reset())
		gameMock.AssertCalled(t, "Reset")
	})

	t.Run("Play delegates", func(t *testing.T) {
		assert.Equal(t, mockOutput, ei.Play(0, []int{1}))
		gameMock.AssertCalled(t, "PlayerPlay", 0, []int{1})
	})

	t.Run("NextRound delegates", func(t *testing.T) {
		assert.Equal(t, mockOutput, ei.NextRound())
		gameMock.AssertCalled(t, "NextRound")
	})
}

func TestEscobaInteractor_CpuTurnsStopOnRoundEnd(t *testing.T) {
	spMock := new(presenter.MockEscobaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("x")
	gameMock := new(interfaces.MockEscobaGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	// Phase is RoundEnd → loop must stop without calling CpuPlay or NextRound.
	gameMock.On("GetPhase").Return(domain.EscobaPhaseRoundEnd)

	ei := usecase.NewEscobaInteractor(gameMock, spMock)
	assert.Equal(t, "x", ei.Reset())
	gameMock.AssertNotCalled(t, "CpuPlay")
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestEscobaInteractor_NextRoundSkipsWhenGameEnded(t *testing.T) {
	spMock := new(presenter.MockEscobaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("done")
	gameMock := new(interfaces.MockEscobaGame)
	gameMock.On("GetGameEndFlag").Return(true)

	ei := usecase.NewEscobaInteractor(gameMock, spMock)
	assert.Equal(t, "done", ei.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestEscobaInteractor_PlayNotHumanTurn(t *testing.T) {
	spMock := new(presenter.MockEscobaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("cpu")
	gameMock := new(interfaces.MockEscobaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	ei := usecase.NewEscobaInteractor(gameMock, spMock)
	assert.Equal(t, "cpu", ei.Play(0, nil))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything, mock.Anything)
}

func TestEscobaInteractor_ResetWithConfigInvalid(t *testing.T) {
	spMock := new(presenter.MockEscobaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("invalid")
	ei := usecase.NewEscobaInteractor(escobaNewTestGame(), spMock)
	bad := domain.EscobaConfig{CpuDifficulty: 99, TargetScore: 10}
	assert.NotEmpty(t, ei.ResetWithConfig(bad))
}

func TestEscobaInteractor_Snapshot(t *testing.T) {
	spMock := new(presenter.MockEscobaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("x")
	ei := usecase.NewEscobaInteractor(escobaNewTestGame(), spMock)
	raw, err := ei.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, raw)
}

func TestRestoreEscobaInteractor(t *testing.T) {
	spMock := new(presenter.MockEscobaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("x")
	game := escobaNewTestGame()
	game.Reset()
	ei := usecase.NewEscobaInteractor(game, spMock)

	raw, err := ei.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreEscobaInteractor(raw, spMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
	assert.Equal(t, ei.GetConfig(), restored.GetConfig())
}

func TestRestoreEscobaInteractor_RejectsInvalid(t *testing.T) {
	spMock := new(presenter.MockEscobaPresenter)
	raw := []byte(`{"ps":[],"ph":"playerTurn"}`)
	restored, err := usecase.RestoreEscobaInteractor(raw, spMock)
	assert.Error(t, err)
	assert.Nil(t, restored)
}
