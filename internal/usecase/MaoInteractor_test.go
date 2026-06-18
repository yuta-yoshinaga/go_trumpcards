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

// newPlayableMaoMock returns a game mock whose runCpuTurns loop exits
// immediately (human turn, not awaiting word, not ended).
func newPlayableMaoMock() *interfaces.MockMaoGame {
	g := new(interfaces.MockMaoGame)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetAwaitingWord").Return(false)
	g.On("GetPhase").Return(domain.MaoPhasePlay)
	g.On("IsHumanTurn").Return(true)
	return g
}

func TestNewMaoInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockMaoPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "MaoInteractor: g must not be nil", func() {
			usecase.NewMaoInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockMaoGame)
		assert.PanicsWithValue(t, "MaoInteractor: gp must not be nil", func() {
			usecase.NewMaoInteractor(gameMock, nil)
		})
	})
}

func TestMaoInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`
	pMock := new(presenter.MockMaoPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := newPlayableMaoMock()
	gameMock.On("Reset").Return()

	ci := usecase.NewMaoInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestMaoInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("valid config sets then resets", func(t *testing.T) {
		pMock := new(presenter.MockMaoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := newPlayableMaoMock()
		cfg := domain.MaoConfig{CpuDifficulty: domain.MaoCpuDifficultyHard, PointLimit: 300}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		ci := usecase.NewMaoInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error without resetting", func(t *testing.T) {
		pMock := new(presenter.MockMaoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockMaoGame)
		ci := usecase.NewMaoInteractor(gameMock, pMock)
		bad := domain.MaoConfig{CpuDifficulty: 99, PointLimit: 200}
		assert.Equal(t, mockOutput, ci.ResetWithConfig(bad))
		gameMock.AssertNotCalled(t, "Reset")
	})
}

func TestMaoInteractor_Play(t *testing.T) {
	mockOutput := `{"ok":1}`

	t.Run("success", func(t *testing.T) {
		pMock := new(presenter.MockMaoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := newPlayableMaoMock()
		gameMock.On("PlayerPlay", 2).Return(nil)
		ci := usecase.NewMaoInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Play(2))
		gameMock.AssertCalled(t, "PlayerPlay", 2)
	})

	t.Run("error path returns output with error", func(t *testing.T) {
		pMock := new(presenter.MockMaoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := newPlayableMaoMock()
		gameMock.On("PlayerPlay", 0).Return(errors.New("bad"))
		ci := usecase.NewMaoInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Play(0))
	})
}

func TestMaoInteractor_OtherActions(t *testing.T) {
	mockOutput := `{"ok":1}`
	pMock := new(presenter.MockMaoPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	pMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := newPlayableMaoMock()
	gameMock.On("PlayerChooseSuit", 3).Return(nil)
	gameMock.On("PlayerDraw").Return(nil)
	gameMock.On("PlayerDeclare").Return(nil)
	gameMock.On("PlayerSkipDeclare").Return(nil)
	gameMock.On("PlayerDeclareWord", "spade").Return(nil)
	gameMock.On("ScoreRound").Return()
	gameMock.On("NextRound").Return()
	gameMock.On("GetConfig").Return(domain.DefaultMaoConfig())

	ci := usecase.NewMaoInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.ChooseSuit(3))
	assert.Equal(t, mockOutput, ci.Draw())
	assert.Equal(t, mockOutput, ci.Declare())
	assert.Equal(t, mockOutput, ci.SkipDeclare())
	assert.Equal(t, mockOutput, ci.DeclareWord("spade"))
	assert.Equal(t, mockOutput, ci.NextRound())
	assert.Equal(t, "log", ci.ActionLog())
	assert.Equal(t, domain.DefaultMaoConfig(), ci.GetConfig())
}

func TestMaoInteractor_ActionErrorPaths(t *testing.T) {
	mockOutput := `{"e":1}`
	pMock := new(presenter.MockMaoPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := newPlayableMaoMock()
	gameMock.On("PlayerChooseSuit", 9).Return(errors.New("x"))
	gameMock.On("PlayerDraw").Return(errors.New("x"))
	gameMock.On("PlayerDeclare").Return(errors.New("x"))
	gameMock.On("PlayerSkipDeclare").Return(errors.New("x"))
	gameMock.On("PlayerDeclareWord", "bad").Return(errors.New("x"))

	ci := usecase.NewMaoInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.ChooseSuit(9))
	assert.Equal(t, mockOutput, ci.Draw())
	assert.Equal(t, mockOutput, ci.Declare())
	assert.Equal(t, mockOutput, ci.SkipDeclare())
	assert.Equal(t, mockOutput, ci.DeclareWord("bad"))
}

func TestMaoInteractor_TurnHelpers(t *testing.T) {
	gameMock := new(interfaces.MockMaoGame)
	gameMock.On("GetPhase").Return(domain.MaoPhaseChooseSuit).Once()
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetAwaitingWord").Return(true)
	pMock := new(presenter.MockMaoPresenter)
	ci := usecase.NewMaoInteractor(gameMock, pMock)
	assert.True(t, ci.IsHumanChooseSuitTurn())
	assert.True(t, ci.IsHumanAwaitingWord())

	gameMock2 := new(interfaces.MockMaoGame)
	gameMock2.On("GetPhase").Return(domain.MaoPhaseMustDeclare)
	gameMock2.On("IsHumanTurn").Return(true)
	ci2 := usecase.NewMaoInteractor(gameMock2, pMock)
	assert.True(t, ci2.IsHumanDeclareTurn())
}

func TestMaoInteractor_RunCpuTurnsPausesOnAwaitingWord(t *testing.T) {
	mockOutput := `{"x":1}`
	pMock := new(presenter.MockMaoPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockMaoGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)     // passes guardNotPlayable
	gameMock.On("GetAwaitingWord").Return(true) // pause immediately in runCpuTurns
	gameMock.On("PlayerPlay", 0).Return(nil)

	ci := usecase.NewMaoInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.Play(0))
	// CpuPlay must NOT be invoked while awaiting word.
	gameMock.AssertNotCalled(t, "CpuPlay")
}

func TestMaoInteractor_RunCpuTurnsAdvancesCpu(t *testing.T) {
	mockOutput := `{"x":1}`
	pMock := new(presenter.MockMaoPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockMaoGame)
	gameMock.On("PlayerPlay", 0).Return(nil)
	gameMock.On("GetAwaitingWord").Return(false)
	gameMock.On("GetPhase").Return(domain.MaoPhasePlay)
	// guard: GameEndFlag=false, IsHumanTurn=true => proceed.
	gameMock.On("GetGameEndFlag").Return(false).Once()
	gameMock.On("IsHumanTurn").Return(true).Once()
	// runCpuTurns loop 1: GameEndFlag=false, IsHumanTurn=false => CpuPlay.
	gameMock.On("GetGameEndFlag").Return(false).Once()
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("CpuPlay").Return().Once()
	// loop 2: ended.
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewMaoInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.Play(0))
	gameMock.AssertCalled(t, "CpuPlay")
}
