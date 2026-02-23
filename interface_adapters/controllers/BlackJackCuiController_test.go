package controllers_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers/usecases"

	"github.com/stretchr/testify/assert"
)

func TestBlackJackCuiController_Method(t *testing.T) {
	mockOutput := "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n"
	bjiMock := new(usecases.MockBlackJackInteractor)
	bjiMock.On("Reset").Return(mockOutput)
	bjiMock.On("Hit").Return(mockOutput)
	bjiMock.On("Stand").Return(mockOutput)
	bjiMock.On("Bet", 100).Return(mockOutput)
	bjiMock.On("DoubleDown").Return(mockOutput)
	bjiMock.On("Split").Return(mockOutput)
	bjiMock.On("Insurance").Return(mockOutput)
	bjiMock.On("DeclineInsurance").Return(mockOutput)
	tbc := controllers.NewBlackJackCuiController(bjiMock)
	t.Run("success Exec q", func(t *testing.T) {
		assert.Equal(t, "bye.", tbc.Exec("q"))
	})
	t.Run("success Exec quit", func(t *testing.T) {
		assert.Equal(t, "bye.", tbc.Exec("quit"))
	})
	t.Run("success Exec r", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("r"))
	})
	t.Run("success Exec reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("reset"))
	})
	t.Run("success Exec h", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("h"))
	})
	t.Run("success Exec hit", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("hit"))
	})
	t.Run("success Exec s", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("s"))
	})
	t.Run("success Exec stand", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("stand"))
	})
	t.Run("success Exec b with amount", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("b 100"))
	})
	t.Run("success Exec bet with amount", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("bet 100"))
	})
	t.Run("error Exec b without amount", func(t *testing.T) {
		assert.Equal(t, "Bet amount is required.", tbc.Exec("b"))
	})
	t.Run("error Exec b with invalid amount", func(t *testing.T) {
		assert.Equal(t, "Invalid bet amount. Please enter a number.", tbc.Exec("b foo"))
	})
	t.Run("success Exec d", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("d"))
	})
	t.Run("success Exec doubledown", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("doubledown"))
	})
	t.Run("success Exec sp", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("sp"))
	})
	t.Run("success Exec split", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("split"))
	})
	t.Run("success Exec i", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("i"))
	})
	t.Run("success Exec insurance", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("insurance"))
	})
	t.Run("success Exec di", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("di"))
	})
	t.Run("success Exec declineinsurance", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("declineinsurance"))
	})
	t.Run("success Exec other", func(t *testing.T) {
		assert.Equal(t, "Unsupported command.", tbc.Exec("other"))
	})
	t.Run("success Exec empty", func(t *testing.T) {
		assert.Equal(t, "Unsupported command.", tbc.Exec(""))
	})
}
