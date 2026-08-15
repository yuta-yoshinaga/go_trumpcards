//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockRikkenInteractor() *usecase.MockRikkenInteractor {
	m := new(usecase.MockRikkenInteractor)
	m.On("Reset").Return("reset result")
	m.On("PlayCard", 3).Return("played 3")
	m.On("Bid", domain.RikkenContractNone).Return("passed")
	m.On("Bid", domain.RikkenContractRik).Return("bid rik")
	m.On("Bid", domain.RikkenContractMisere).Return("bid misere")
	m.On("Bid", domain.RikkenContractSolo).Return("bid solo")
	m.On("Bid", domain.RikkenContractOpenMisere).Return("bid open")
	m.On("Call", domain.CardDesignSpade).Return("called spade")
	m.On("Call", domain.CardDesignHeart).Return("called heart")
	m.On("NextRound").Return("next round")
	m.On("GiveUp").Return("gave up")
	m.On("Hint").Return("hint result")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestRikkenCuiController_QuitAndReset(t *testing.T) {
	c := controller.NewRikkenCuiController(newMockRikkenInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestRikkenCuiController_Play(t *testing.T) {
	c := controller.NewRikkenCuiController(newMockRikkenInteractor())
	assert.Equal(t, "played 3", c.Exec("p 3"))
	assert.Equal(t, "played 3", c.Exec("play 3"))
	assert.Contains(t, c.Exec("p"), msgCardIndexRequired())
	assert.Contains(t, c.Exec("p abc"), msgInvalidCardIndexPrefix())
}

// **4 種類の契約とパスが全部通る。** パスは契約 0 で、別経路にはしていません。
func TestRikkenCuiController_Bid(t *testing.T) {
	c := controller.NewRikkenCuiController(newMockRikkenInteractor())
	assert.Equal(t, "bid rik", c.Exec("bid rik"))
	assert.Equal(t, "bid misere", c.Exec("bid misere"))
	assert.Equal(t, "bid misere", c.Exec("bid mis"))
	assert.Equal(t, "bid solo", c.Exec("bid solo"))
	assert.Equal(t, "bid open", c.Exec("bid open"))
	assert.Equal(t, "bid open", c.Exec("bid openmisere"))
	assert.Equal(t, "passed", c.Exec("bid pass"))
	assert.Equal(t, "passed", c.Exec("pass"))

	assert.Contains(t, c.Exec("bid"), "Invalid contract")
	assert.Contains(t, c.Exec("bid nonsense"), "Invalid contract")
}

func TestRikkenCuiController_Call(t *testing.T) {
	c := controller.NewRikkenCuiController(newMockRikkenInteractor())
	assert.Equal(t, "called spade", c.Exec("call s"))
	assert.Equal(t, "called spade", c.Exec("call spade"))
	assert.Equal(t, "called heart", c.Exec("call h"))
	assert.Contains(t, c.Exec("call"), "Invalid trump")
	assert.Contains(t, c.Exec("call x"), "Invalid trump")
}

func TestRikkenCuiController_OtherCommands(t *testing.T) {
	c := controller.NewRikkenCuiController(newMockRikkenInteractor())
	assert.Equal(t, "next round", c.Exec("next"))
	assert.Equal(t, "gave up", c.Exec("giveup"))
	assert.Equal(t, "hint result", c.Exec("hint"))
	assert.Equal(t, "action log result", c.Exec("log"))
}

func TestRikkenCuiController_UnknownCommand(t *testing.T) {
	c := controller.NewRikkenCuiController(newMockRikkenInteractor())
	out := c.Exec("nonsense")
	assert.NotEmpty(t, out)
	assert.NotEqual(t, "reset result", out, "未知のコマンドでリセットしない")
}
