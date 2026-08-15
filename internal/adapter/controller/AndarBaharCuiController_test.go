//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockAndarBaharInteractor() *usecase.MockAndarBaharInteractor {
	m := new(usecase.MockAndarBaharInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100, domain.AndarBaharBetAndar, 0, domain.AndarBaharSideNone).Return("bet andar")
	m.On("Bet", 100, domain.AndarBaharBetBahar, 0, domain.AndarBaharSideNone).Return("bet bahar")
	m.On("Bet", 100, domain.AndarBaharBetAndar, 50, domain.AndarBaharSide6To10).Return("bet andar with side")
	m.On("ClearHistory").Return("history cleared")
	m.On("Hint").Return("hint result")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestAndarBaharCuiController_Quit(t *testing.T) {
	c := controller.NewAndarBaharCuiController(newMockAndarBaharInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestAndarBaharCuiController_Reset(t *testing.T) {
	c := controller.NewAndarBaharCuiController(newMockAndarBaharInteractor())
	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestAndarBaharCuiController_Bet(t *testing.T) {
	c := controller.NewAndarBaharCuiController(newMockAndarBaharInteractor())
	assert.Equal(t, "bet andar", c.Exec("b 100 a"))
	assert.Equal(t, "bet andar", c.Exec("bet 100 andar"))
	assert.Equal(t, "bet andar", c.Exec("b 100 0"))
	assert.Equal(t, "bet bahar", c.Exec("b 100 b"))
	assert.Equal(t, "bet bahar", c.Exec("bet 100 bahar"))
	assert.Equal(t, "bet bahar", c.Exec("b 100 1"))
}

// **サイドベットは 3・4 番目の引数。** 省略すれば賭けません。
func TestAndarBaharCuiController_BetWithSideBet(t *testing.T) {
	c := controller.NewAndarBaharCuiController(newMockAndarBaharInteractor())
	assert.Equal(t, "bet andar with side", c.Exec("b 100 a 50 2"))
	// 3 つ目だけ与えても帯が無いのでサイドベットは付かない。
	assert.Equal(t, "bet andar", c.Exec("b 100 a 50"))
}

func TestAndarBaharCuiController_Bet_InvalidInput(t *testing.T) {
	c := controller.NewAndarBaharCuiController(newMockAndarBaharInteractor())

	assert.Contains(t, c.Exec("b"), "required")
	assert.Contains(t, c.Exec("b 100"), "required")
	assert.Contains(t, c.Exec("b abc a"), msgInvalidBetAmountPrefix())
	assert.Contains(t, c.Exec("b 100 x"), "Invalid bet target")
	assert.Contains(t, c.Exec("b 100 a abc 2"), msgStem("invalidSideBetAmountANumber"))
	assert.Contains(t, c.Exec("b 100 a 50 abc"), msgStem("invalidSideBetBandANumber"))
	// 帯は 0..6 の範囲外を弾く。
	assert.Contains(t, c.Exec("b 100 a 50 9"), msgStem("invalidSideBetBandANumber"))
}

func TestAndarBaharCuiController_ClearHintAndLog(t *testing.T) {
	c := controller.NewAndarBaharCuiController(newMockAndarBaharInteractor())
	assert.Equal(t, "history cleared", c.Exec("clear"))
	assert.Equal(t, "hint result", c.Exec("hint"))
	assert.Equal(t, "action log result", c.Exec("log"))
}

func TestAndarBaharCuiController_UnknownCommand(t *testing.T) {
	c := controller.NewAndarBaharCuiController(newMockAndarBaharInteractor())
	out := c.Exec("nonsense")
	assert.NotEmpty(t, out)
	assert.NotEqual(t, "reset result", out, "未知のコマンドでリセットしない")
}
