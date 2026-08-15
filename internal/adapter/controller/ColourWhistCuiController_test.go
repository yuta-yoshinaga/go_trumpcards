//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockColourWhistInteractor() *usecase.MockColourWhistInteractor {
	m := new(usecase.MockColourWhistInteractor)
	m.On("Reset").Return("reset result")
	m.On("PlayCard", 2).Return("played 2")
	m.On("Bid", domain.ColourWhistContractNone).Return("passed")
	m.On("Bid", domain.ColourWhistContractSamen).Return("bid samen")
	m.On("Bid", domain.ColourWhistContractAlleen).Return("bid alleen")
	m.On("Bid", domain.ColourWhistContractMiserie).Return("bid miserie")
	m.On("Call", domain.CardDesignHeart).Return("called heart")
	m.On("NextRound").Return("next round")
	m.On("GiveUp").Return("gave up")
	m.On("Hint").Return("hint result")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestColourWhistCuiController_QuitAndReset(t *testing.T) {
	c := controller.NewColourWhistCuiController(newMockColourWhistInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "reset result", c.Exec("r"))
}

func TestColourWhistCuiController_Play(t *testing.T) {
	c := controller.NewColourWhistCuiController(newMockColourWhistInteractor())
	assert.Equal(t, "played 2", c.Exec("p 2"))
	assert.Contains(t, c.Exec("p"), msgCardIndexRequired())
	assert.Contains(t, c.Exec("p abc"), msgInvalidCardIndexPrefix())
}

// **競れる契約は3つだけ。** troel は配りでしか成立しないので語彙に入れません。
func TestColourWhistCuiController_Bid(t *testing.T) {
	c := controller.NewColourWhistCuiController(newMockColourWhistInteractor())
	assert.Equal(t, "bid samen", c.Exec("bid samen"))
	assert.Equal(t, "bid alleen", c.Exec("bid alleen"))
	assert.Equal(t, "bid miserie", c.Exec("bid miserie"))
	assert.Equal(t, "bid miserie", c.Exec("bid mis"))
	assert.Equal(t, "passed", c.Exec("bid pass"))
	assert.Equal(t, "passed", c.Exec("pass"))

	// **troel は競れない。** 語彙に無いのでパーサが弾きます。
	out := c.Exec("bid troel")
	assert.Contains(t, out, "Invalid contract")
	assert.Contains(t, out, "troel is dealt, not bid")

	assert.Contains(t, c.Exec("bid"), "Invalid contract")
	assert.Contains(t, c.Exec("bid nonsense"), "Invalid contract")
}

func TestColourWhistCuiController_CallAndOthers(t *testing.T) {
	c := controller.NewColourWhistCuiController(newMockColourWhistInteractor())
	assert.Equal(t, "called heart", c.Exec("call h"))
	assert.Equal(t, "called heart", c.Exec("call heart"))
	assert.Contains(t, c.Exec("call"), "Invalid trump")
	assert.Contains(t, c.Exec("call x"), "Invalid trump")

	assert.Equal(t, "next round", c.Exec("next"))
	assert.Equal(t, "gave up", c.Exec("giveup"))
	assert.Equal(t, "hint result", c.Exec("hint"))
	assert.Equal(t, "action log result", c.Exec("log"))
}

func TestColourWhistCuiController_UnknownCommand(t *testing.T) {
	c := controller.NewColourWhistCuiController(newMockColourWhistInteractor())
	out := c.Exec("nonsense")
	assert.NotEmpty(t, out)
	assert.NotEqual(t, "reset result", out, "未知のコマンドでリセットしない")
}
