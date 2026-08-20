//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMaoCuiMock() *mockUsecases.MockMaoInteractor {
	m := new(mockUsecases.MockMaoInteractor)
	m.On("GetConfig").Return(domain.DefaultMaoConfig())
	m.On("ResetWithConfig", mock.Anything).Return(`{"phase":0}`)
	m.On("Play", mock.Anything).Return(`{"phase":0}`)
	m.On("ChooseSuit", mock.Anything).Return(`{"phase":0}`)
	m.On("Draw").Return(`{"phase":0}`)
	m.On("Declare").Return(`{"phase":0}`)
	m.On("SkipDeclare").Return(`{"phase":0}`)
	m.On("DeclareWord", mock.Anything).Return(`{"phase":0}`)
	m.On("NextRound").Return(`{"phase":0}`)
	m.On("ActionLog").Return(`{"phase":0}`)
	m.On("IsHumanChooseSuitTurn").Return(false)
	m.On("IsHumanDeclareTurn").Return(false)
	m.On("IsHumanAwaitingWord").Return(false)
	return m
}

func TestMaoCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewMaoCuiController(newMaoCuiMock()).Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMaoCuiMock()
		c := controller.NewMaoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultMaoConfig())
	})

	t.Run("play with index", func(t *testing.T) {
		m := newMaoCuiMock()
		assert.Equal(t, mockOutput, controller.NewMaoCuiController(m).Exec("p 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		assert.Contains(t, controller.NewMaoCuiController(newMaoCuiMock()).Exec("p"), msgCardIndexRequired())
	})

	t.Run("draw", func(t *testing.T) {
		m := newMaoCuiMock()
		assert.Equal(t, mockOutput, controller.NewMaoCuiController(m).Exec("d"))
		m.AssertCalled(t, "Draw")
	})

	t.Run("suit", func(t *testing.T) {
		m := newMaoCuiMock()
		assert.Equal(t, mockOutput, controller.NewMaoCuiController(m).Exec("s 3"))
		m.AssertCalled(t, "ChooseSuit", 3)
	})

	t.Run("declare", func(t *testing.T) {
		m := newMaoCuiMock()
		assert.Equal(t, mockOutput, controller.NewMaoCuiController(m).Exec("dc"))
		m.AssertCalled(t, "Declare")
	})

	t.Run("skipdeclare", func(t *testing.T) {
		m := newMaoCuiMock()
		assert.Equal(t, mockOutput, controller.NewMaoCuiController(m).Exec("sk"))
		m.AssertCalled(t, "SkipDeclare")
	})

	t.Run("declareword", func(t *testing.T) {
		m := newMaoCuiMock()
		assert.Equal(t, mockOutput, controller.NewMaoCuiController(m).Exec("dw spade"))
		m.AssertCalled(t, "DeclareWord", "spade")
	})

	t.Run("declareword no arg", func(t *testing.T) {
		assert.Contains(t, controller.NewMaoCuiController(newMaoCuiMock()).Exec("dw"), msgStem("wordRequired"))
	})

	t.Run("nextround", func(t *testing.T) {
		m := newMaoCuiMock()
		assert.Equal(t, mockOutput, controller.NewMaoCuiController(m).Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMaoCuiMock()
		assert.Equal(t, mockOutput, controller.NewMaoCuiController(m).Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("setlimit", func(t *testing.T) {
		m := newMaoCuiMock()
		assert.Equal(t, mockOutput, controller.NewMaoCuiController(m).Exec("sl 100"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
}

func TestMaoCuiController_PlayWithPrompts(t *testing.T) {
	t.Run("awaiting word emits word prompt", func(t *testing.T) {
		m := newMaoCuiMock()
		m2 := new(mockUsecases.MockMaoInteractor)
		m2.On("Play", mock.Anything).Return(`{"phase":0}`)
		m2.On("IsHumanAwaitingWord").Return(true)
		m2.On("IsHumanChooseSuitTurn").Return(false)
		m2.On("IsHumanDeclareTurn").Return(false)
		out := controller.NewMaoCuiController(m2).Exec("p 0")
		assert.True(t, cuiutil.IsPromptRequest(out))
		_ = m
	})

	t.Run("choose-suit emits suit prompt", func(t *testing.T) {
		m := new(mockUsecases.MockMaoInteractor)
		m.On("Play", mock.Anything).Return(`{"phase":0}`)
		m.On("IsHumanAwaitingWord").Return(false)
		m.On("IsHumanChooseSuitTurn").Return(true)
		m.On("IsHumanDeclareTurn").Return(false)
		out := controller.NewMaoCuiController(m).Exec("p 0")
		assert.True(t, cuiutil.IsPromptRequest(out))
	})

	t.Run("declare emits declare prompt", func(t *testing.T) {
		m := new(mockUsecases.MockMaoInteractor)
		m.On("Play", mock.Anything).Return(`{"phase":0}`)
		m.On("IsHumanAwaitingWord").Return(false)
		m.On("IsHumanChooseSuitTurn").Return(false)
		m.On("IsHumanDeclareTurn").Return(true)
		out := controller.NewMaoCuiController(m).Exec("p 0")
		assert.True(t, cuiutil.IsPromptRequest(out))
	})
}
