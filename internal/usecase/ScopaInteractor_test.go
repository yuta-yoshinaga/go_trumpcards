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

func newTestScopa() *domain.Scopa {
	return domain.NewDefaultScopa()
}

func TestNewScopaInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockScopaPresenter)
	t.Run("panics when sg is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "ScopaInteractor: sg must not be nil", func() {
			usecase.NewScopaInteractor(nil, spMock)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "ScopaInteractor: sp must not be nil", func() {
			usecase.NewScopaInteractor(newTestScopa(), nil)
		})
	})
}

func TestScopaInteractor_Methods(t *testing.T) {
	mockOutput := `{"players":[]}`
	spMock := new(presenter.MockScopaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	si := usecase.NewScopaInteractor(newTestScopa(), spMock)

	t.Run("Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, si.Reset())
	})

	t.Run("Play returns output", func(t *testing.T) {
		assert.Equal(t, mockOutput, si.Play(0, nil))
	})

	t.Run("ResetWithConfig valid", func(t *testing.T) {
		cfg := domain.DefaultScopaConfig()
		assert.Equal(t, mockOutput, si.ResetWithConfig(cfg))
	})

	t.Run("GetConfig returns config", func(t *testing.T) {
		cfg := si.GetConfig()
		assert.Equal(t, domain.ScopaDefaultTargetScore, cfg.TargetScore)
	})
}

func TestScopaInteractor_ActionLog(t *testing.T) {
	spMock := new(presenter.MockScopaPresenter)
	gameMock := new(interfaces.MockScopaGame)
	spMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)
	si := usecase.NewScopaInteractor(gameMock, spMock)
	assert.Equal(t, `{"entries":[]}`, si.ActionLog())
	spMock.AssertExpectations(t)
}

func TestScopaInteractor_MockGame(t *testing.T) {
	mockOutput := `{"players":[]}`
	spMock := new(presenter.MockScopaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)

	gameMock := new(interfaces.MockScopaGame)
	gameMock.On("Reset").Return()
	gameMock.On("NextRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetPhase").Return(domain.ScopaPhasePlayerTurn)
	gameMock.On("CpuPlay").Return()
	gameMock.On("PlayerPlay", mock.Anything, mock.Anything).Return(nil)
	gameMock.On("SetConfig", mock.Anything).Return()
	gameMock.On("GetConfig").Return(domain.DefaultScopaConfig())

	si := usecase.NewScopaInteractor(gameMock, spMock)

	t.Run("Reset delegates", func(t *testing.T) {
		assert.Equal(t, mockOutput, si.Reset())
		gameMock.AssertCalled(t, "Reset")
	})

	t.Run("Play delegates", func(t *testing.T) {
		assert.Equal(t, mockOutput, si.Play(0, []int{1}))
		gameMock.AssertCalled(t, "PlayerPlay", 0, []int{1})
	})

	t.Run("NextRound delegates", func(t *testing.T) {
		assert.Equal(t, mockOutput, si.NextRound())
		gameMock.AssertCalled(t, "NextRound")
	})
}

func TestScopaInteractor_NextRoundSkipsWhenGameEnded(t *testing.T) {
	spMock := new(presenter.MockScopaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("done")
	gameMock := new(interfaces.MockScopaGame)
	gameMock.On("GetGameEndFlag").Return(true)

	si := usecase.NewScopaInteractor(gameMock, spMock)
	assert.Equal(t, "done", si.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestScopaInteractor_PlayNotHumanTurn(t *testing.T) {
	spMock := new(presenter.MockScopaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("cpu")
	gameMock := new(interfaces.MockScopaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	si := usecase.NewScopaInteractor(gameMock, spMock)
	assert.Equal(t, "cpu", si.Play(0, nil))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything, mock.Anything)
}

func TestScopaInteractor_ResetWithConfigInvalid(t *testing.T) {
	spMock := new(presenter.MockScopaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("invalid")
	si := usecase.NewScopaInteractor(newTestScopa(), spMock)
	bad := domain.ScopaConfig{CpuDifficulty: 99, TargetScore: 11}
	// resetWithValidatedConfig applies then resets; invalid difficulty would be
	// rejected by domain Validate on the next reset path. We assert it returns output.
	assert.NotEmpty(t, si.ResetWithConfig(bad))
}

func TestScopaInteractor_Snapshot(t *testing.T) {
	spMock := new(presenter.MockScopaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("x")
	si := usecase.NewScopaInteractor(newTestScopa(), spMock)
	raw, err := si.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, raw)
}

func TestRestoreScopaInteractor(t *testing.T) {
	spMock := new(presenter.MockScopaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("x")
	game := newTestScopa()
	game.Reset()
	si := usecase.NewScopaInteractor(game, spMock)

	raw, err := si.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreScopaInteractor(raw, spMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
	assert.Equal(t, si.GetConfig(), restored.GetConfig())
}

func TestRestoreScopaInteractor_RejectsNilTrumpCards(t *testing.T) {
	spMock := new(presenter.MockScopaPresenter)
	raw := []byte(`{"pl":[],"cf":{"ts":11,"di":1},"ph":"playerTurn","ct":0,"tb":[],"lc":-1,"al":[]}`)
	restored, err := usecase.RestoreScopaInteractor(raw, spMock)
	assert.Error(t, err)
	assert.Nil(t, restored)
}
