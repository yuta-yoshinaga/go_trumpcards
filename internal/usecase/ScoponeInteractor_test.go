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

func newTestScopone() *domain.Scopone {
	return domain.NewDefaultScopone()
}

func TestNewScoponeInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockScoponePresenter)
	t.Run("panics when sg is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "ScoponeInteractor: sg must not be nil", func() {
			usecase.NewScoponeInteractor(nil, spMock)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "ScoponeInteractor: sp must not be nil", func() {
			usecase.NewScoponeInteractor(newTestScopone(), nil)
		})
	})
}

func TestScoponeInteractor_Methods(t *testing.T) {
	mockOutput := `{"players":[]}`
	spMock := new(presenter.MockScoponePresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	si := usecase.NewScoponeInteractor(newTestScopone(), spMock)

	t.Run("Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, si.Reset())
	})

	t.Run("Play returns output", func(t *testing.T) {
		assert.Equal(t, mockOutput, si.Play(0, nil))
	})

	t.Run("ResetWithConfig valid", func(t *testing.T) {
		cfg := domain.DefaultScoponeConfig()
		assert.Equal(t, mockOutput, si.ResetWithConfig(cfg))
	})

	t.Run("GetConfig returns config", func(t *testing.T) {
		cfg := si.GetConfig()
		assert.Equal(t, domain.ScoponeDefaultTargetScore, cfg.TargetScore)
	})
}

func TestScoponeInteractor_ActionLog(t *testing.T) {
	spMock := new(presenter.MockScoponePresenter)
	gameMock := new(interfaces.MockScoponeGame)
	spMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)
	si := usecase.NewScoponeInteractor(gameMock, spMock)
	assert.Equal(t, `{"entries":[]}`, si.ActionLog())
	spMock.AssertExpectations(t)
}

func TestScoponeInteractor_MockGame(t *testing.T) {
	mockOutput := `{"players":[]}`
	spMock := new(presenter.MockScoponePresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)

	gameMock := new(interfaces.MockScoponeGame)
	gameMock.On("Reset").Return()
	gameMock.On("NextRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetPhase").Return(domain.ScoponePhasePlayerTurn)
	gameMock.On("CpuPlay").Return()
	gameMock.On("PlayerPlay", mock.Anything, mock.Anything).Return(nil)
	gameMock.On("SetConfig", mock.Anything).Return()
	gameMock.On("GetConfig").Return(domain.DefaultScoponeConfig())

	si := usecase.NewScoponeInteractor(gameMock, spMock)

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

// TestScoponeInteractor_CpuTurnsAdvance verifies runCpuTurns spins CPU turns
// until the human's turn and does NOT auto-advance a RoundEnd phase.
func TestScoponeInteractor_CpuTurnsStopOnRoundEnd(t *testing.T) {
	spMock := new(presenter.MockScoponePresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("x")
	gameMock := new(interfaces.MockScoponeGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	// Phase is RoundEnd → loop must stop without calling CpuPlay or NextRound.
	gameMock.On("GetPhase").Return(domain.ScoponePhaseRoundEnd)

	si := usecase.NewScoponeInteractor(gameMock, spMock)
	assert.Equal(t, "x", si.Reset())
	gameMock.AssertNotCalled(t, "CpuPlay")
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestScoponeInteractor_NextRoundSkipsWhenGameEnded(t *testing.T) {
	spMock := new(presenter.MockScoponePresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("done")
	gameMock := new(interfaces.MockScoponeGame)
	gameMock.On("GetGameEndFlag").Return(true)

	si := usecase.NewScoponeInteractor(gameMock, spMock)
	assert.Equal(t, "done", si.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestScoponeInteractor_PlayNotHumanTurn(t *testing.T) {
	spMock := new(presenter.MockScoponePresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("cpu")
	gameMock := new(interfaces.MockScoponeGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	si := usecase.NewScoponeInteractor(gameMock, spMock)
	assert.Equal(t, "cpu", si.Play(0, nil))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything, mock.Anything)
}

func TestScoponeInteractor_ResetWithConfigInvalid(t *testing.T) {
	spMock := new(presenter.MockScoponePresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("invalid")
	si := usecase.NewScoponeInteractor(newTestScopone(), spMock)
	bad := domain.ScoponeConfig{CpuDifficulty: 99, TargetScore: 11}
	assert.NotEmpty(t, si.ResetWithConfig(bad))
}

func TestScoponeInteractor_Snapshot(t *testing.T) {
	spMock := new(presenter.MockScoponePresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("x")
	si := usecase.NewScoponeInteractor(newTestScopone(), spMock)
	raw, err := si.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, raw)
}

func TestRestoreScoponeInteractor(t *testing.T) {
	spMock := new(presenter.MockScoponePresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("x")
	game := newTestScopone()
	game.Reset()
	si := usecase.NewScoponeInteractor(game, spMock)

	raw, err := si.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreScoponeInteractor(raw, spMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
	assert.Equal(t, si.GetConfig(), restored.GetConfig())
}

func TestRestoreScoponeInteractor_RejectsInvalid(t *testing.T) {
	spMock := new(presenter.MockScoponePresenter)
	raw := []byte(`{"ps":[],"ph":"playerTurn"}`)
	restored, err := usecase.RestoreScoponeInteractor(raw, spMock)
	assert.Error(t, err)
	assert.Nil(t, restored)
}
