package usecase_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOldMaidInteractor_Method(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"nextDrawTargetIdx":1,"gameEndFlag":false,"loserIdx":-1,"lastDrawPlayerIdx":-1,"lastDrawFromIdx":-1,"lastDiscardedPairs":0,"hasDrawn":false,"message":""}`
	ompMock := new(presenter.MockOldMaidPresenter)
	ompMock.On("Output", mock.AnythingOfType("string")).Return(mockOutput)
	toi := usecase.NewOldMaidInteractor(ompMock)

	t.Run("success Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, toi.Reset())
	})
	t.Run("success Draw", func(t *testing.T) {
		assert.Equal(t, mockOutput, toi.Draw(-1))
	})
}
