//go:build test

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func TestChinesePokerCuiController_Exec_Reset(t *testing.T) {
	mock := new(mockusecase.MockChinesePokerInteractor)
	mock.On("Reset").Return("reset ok")
	cc := NewChinesePokerCuiController(mock)
	result := cc.Exec("r")
	assert.Equal(t, "reset ok", result)
}

func TestChinesePokerCuiController_Exec_Bet(t *testing.T) {
	mock := new(mockusecase.MockChinesePokerInteractor)
	mock.On("Bet", 100).Return("bet ok")
	cc := NewChinesePokerCuiController(mock)
	result := cc.Exec("b 100")
	assert.Equal(t, "bet ok", result)
}

func TestChinesePokerCuiController_Exec_BetMissingAmount(t *testing.T) {
	mock := new(mockusecase.MockChinesePokerInteractor)
	cc := NewChinesePokerCuiController(mock)
	result := cc.Exec("b")
	assert.Contains(t, result, msgBetAmountRequired())
}

func TestChinesePokerCuiController_Exec_Set(t *testing.T) {
	mock := new(mockusecase.MockChinesePokerInteractor)
	mock.On("SetHands", []int{0, 1, 2}, []int{3, 4, 5, 6, 7}).Return("set ok")
	cc := NewChinesePokerCuiController(mock)
	result := cc.Exec("s 0 1 2 3 4 5 6 7")
	assert.Equal(t, "set ok", result)
}

func TestChinesePokerCuiController_Exec_SetMissingArgs(t *testing.T) {
	mock := new(mockusecase.MockChinesePokerInteractor)
	cc := NewChinesePokerCuiController(mock)
	result := cc.Exec("s 0 1 2")
	assert.Contains(t, result, "8 card indices")
}

func TestChinesePokerCuiController_Exec_SetInvalidFrontIndex(t *testing.T) {
	mock := new(mockusecase.MockChinesePokerInteractor)
	cc := NewChinesePokerCuiController(mock)
	result := cc.Exec("s a 1 2 3 4 5 6 7")
	assert.Contains(t, result, "Invalid front index")
}

func TestChinesePokerCuiController_Exec_SetInvalidMiddleIndex(t *testing.T) {
	mock := new(mockusecase.MockChinesePokerInteractor)
	cc := NewChinesePokerCuiController(mock)
	result := cc.Exec("s 0 1 2 x 4 5 6 7")
	assert.Contains(t, result, "Invalid middle index")
}

func TestChinesePokerCuiController_Exec_Log(t *testing.T) {
	mock := new(mockusecase.MockChinesePokerInteractor)
	mock.On("ActionLog").Return("[]")
	cc := NewChinesePokerCuiController(mock)
	result := cc.Exec("l")
	assert.Equal(t, "[]", result)
}

func TestChinesePokerCuiController_Exec_Quit(t *testing.T) {
	mock := new(mockusecase.MockChinesePokerInteractor)
	cc := NewChinesePokerCuiController(mock)
	result := cc.Exec("q")
	assert.NotEmpty(t, result)
}

// **CUI は13枚を自力で 3/5/5 に分けるしかなく、ファウルしても無警告だった
// (#4717)。**
func TestChinesePokerCuiController_Exec_Hint(t *testing.T) {
	mock := new(mockusecase.MockChinesePokerInteractor)
	mock.On("Hint").Return("hint ok")
	cc := NewChinesePokerCuiController(mock)

	assert.Equal(t, "hint ok", cc.Exec("h"))
	assert.Equal(t, "hint ok", cc.Exec("hint"))
}

// **既存コマンドは何も変わらない。**
func TestChinesePokerCuiController_Exec_HintDoesNotShadowSet(t *testing.T) {
	mock := new(mockusecase.MockChinesePokerInteractor)
	mock.On("Bet", 100).Return("bet ok")
	cc := NewChinesePokerCuiController(mock)

	assert.Equal(t, "bet ok", cc.Exec("b 100"))
	mock.AssertNotCalled(t, "Hint")
}
