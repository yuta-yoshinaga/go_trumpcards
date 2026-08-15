//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockBotifarraInteractor() *usecase.MockBotifarraInteractor {
	m := new(usecase.MockBotifarraInteractor)
	m.On("Reset").Return("reset result")
	m.On("PlayCard", 3).Return("played 3")
	m.On("Declare", domain.CardDesignSpade).Return("declared spade")
	m.On("Declare", domain.CardDesignClover).Return("declared clover")
	m.On("Declare", domain.CardDesignHeart).Return("declared heart")
	m.On("Declare", domain.CardDesignDiamond).Return("declared diamond")
	m.On("Declare", domain.BotifarraNoTrump).Return("declared no trump")
	m.On("Delegate").Return("delegated")
	m.On("Double").Return("doubled")
	m.On("PassDouble").Return("passed")
	m.On("NextRound").Return("next round")
	m.On("GiveUp").Return("gave up")
	m.On("Hint").Return("hint result")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestBotifarraCuiController_QuitAndReset(t *testing.T) {
	c := controller.NewBotifarraCuiController(newMockBotifarraInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestBotifarraCuiController_Play(t *testing.T) {
	c := controller.NewBotifarraCuiController(newMockBotifarraInteractor())
	assert.Equal(t, "played 3", c.Exec("p 3"))
	assert.Equal(t, "played 3", c.Exec("play 3"))
	assert.Contains(t, c.Exec("p"), msgCardIndexRequired())
	assert.Contains(t, c.Exec("p abc"), msgInvalidCardIndexPrefix())
}

// **切り札なし (n) も有効な宣言。** 「指定が無い」とは区別します。
func TestBotifarraCuiController_Declare(t *testing.T) {
	c := controller.NewBotifarraCuiController(newMockBotifarraInteractor())
	assert.Equal(t, "declared spade", c.Exec("declare s"))
	assert.Equal(t, "declared spade", c.Exec("declare spade"))
	assert.Equal(t, "declared clover", c.Exec("declare c"))
	assert.Equal(t, "declared clover", c.Exec("declare club"))
	assert.Equal(t, "declared heart", c.Exec("declare h"))
	assert.Equal(t, "declared diamond", c.Exec("declare d"))
	assert.Equal(t, "declared no trump", c.Exec("declare n"))
	assert.Equal(t, "declared no trump", c.Exec("declare none"))

	assert.Contains(t, c.Exec("declare"), msgStem("invalidTrumpSCHDN"))
	assert.Contains(t, c.Exec("declare x"), msgStem("invalidTrumpSCHDN"))
}

func TestBotifarraCuiController_OtherCommands(t *testing.T) {
	c := controller.NewBotifarraCuiController(newMockBotifarraInteractor())
	assert.Equal(t, "delegated", c.Exec("delegate"))
	assert.Equal(t, "doubled", c.Exec("double"))
	assert.Equal(t, "passed", c.Exec("pass"))
	assert.Equal(t, "next round", c.Exec("next"))
	assert.Equal(t, "gave up", c.Exec("giveup"))
	assert.Equal(t, "hint result", c.Exec("hint"))
	assert.Equal(t, "action log result", c.Exec("log"))
}

func TestBotifarraCuiController_UnknownCommand(t *testing.T) {
	c := controller.NewBotifarraCuiController(newMockBotifarraInteractor())
	out := c.Exec("nonsense")
	assert.NotEmpty(t, out)
	assert.NotEqual(t, "reset result", out, "未知のコマンドでリセットしない")
}
