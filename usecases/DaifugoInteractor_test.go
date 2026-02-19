package usecases_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/usecases"
	"github.com/yuta-yoshinaga/go_trumpcards/usecases/presenters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDaifugoInteractor_Method(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableCards":[],"lastPlayPlayerIdx":-1,"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
	dgpMock := new(presenters.MockDaifugoPresenter)
	dgpMock.On("Output", mock.AnythingOfType("string")).Return(mockOutput)
	tdi := usecases.NewDaifugoInteractor(dgpMock)

	t.Run("success Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, tdi.Reset())
	})

	t.Run("success Play with pass (empty indices)", func(t *testing.T) {
		assert.Equal(t, mockOutput, tdi.Play([]int{}))
	})

	t.Run("success Play with indices", func(t *testing.T) {
		assert.Equal(t, mockOutput, tdi.Play([]int{0}))
	})
}
