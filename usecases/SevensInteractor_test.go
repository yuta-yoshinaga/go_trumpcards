package usecases_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/usecases"
	"github.com/yuta-yoshinaga/go_trumpcards/usecases/presenters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSevensInteractor_Method(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
	spMock := new(presenters.MockSevensPresenter)
	spMock.On("Output", mock.AnythingOfType("string")).Return(mockOutput)
	tsi := usecases.NewSevensInteractor(spMock)

	t.Run("success Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, tsi.Reset())
	})

	t.Run("success ResetWithConfig all enabled", func(t *testing.T) {
		result := tsi.ResetWithConfig(true, 2, true)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("success ResetWithConfig default values", func(t *testing.T) {
		result := tsi.ResetWithConfig(false, 0, false)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("success ResetWithConfig jokerCount clamped to 0", func(t *testing.T) {
		result := tsi.ResetWithConfig(false, -5, false)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("success ResetWithConfig jokerCount clamped to 2", func(t *testing.T) {
		result := tsi.ResetWithConfig(false, 10, false)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("success Play with pass (idx -1)", func(t *testing.T) {
		assert.Equal(t, mockOutput, tsi.Play(-1))
	})

	t.Run("success Play with index", func(t *testing.T) {
		assert.Equal(t, mockOutput, tsi.Play(0))
	})

	t.Run("success PlayJoker", func(t *testing.T) {
		assert.Equal(t, mockOutput, tsi.PlayJoker(0, 1, 6))
	})
}
