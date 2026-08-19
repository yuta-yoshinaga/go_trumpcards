//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockLetItRideInteractor() *usecase.MockLetItRideInteractor {
	m := new(usecase.MockLetItRideInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100).Return("bet result")
	m.On("Pull").Return("pull result")
	m.On("PullConfirm").Return("pull confirm result")
	m.On("LetItRide").Return("letitride result")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestLetItRideCuiController_Quit(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestLetItRideCuiController_Reset(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestLetItRideCuiController_Bet(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	t.Run("short", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("b 100"))
	})

	t.Run("long", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("bet 100"))
	})
}

func TestLetItRideCuiController_Bet_Errors(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	t.Run("missing args", func(t *testing.T) {
		result := c.Exec("b")
		assert.Contains(t, result, msgBetAmountRequired())
	})

	t.Run("invalid amount", func(t *testing.T) {
		result := c.Exec("b abc")
		assert.Contains(t, result, msgInvalidBetAmountPrefix())
	})

	t.Run("zero amount", func(t *testing.T) {
		result := c.Exec("b 0")
		assert.Contains(t, result, msgInvalidBetAmountPrefix())
	})

	t.Run("negative amount", func(t *testing.T) {
		result := c.Exec("b -10")
		assert.Contains(t, result, msgInvalidBetAmountPrefix())
	})
}

// **Pull は取り消せない。**Web は専用ダイアログでリスクの前後を見せてから
// 実行するのに、CUI は "p" 一発で確定していた (#4699)。
func TestLetItRideCuiController_Pull_RequiresConfirmation(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	assert.Equal(t, "pull confirm result", c.Exec("p"))
	m.AssertNotCalled(t, "Pull")

	assert.Equal(t, "pull result", c.Exec("y"))
	m.AssertCalled(t, "Pull")
}

func TestLetItRideCuiController_Pull_ConfirmAliases(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	assert.Equal(t, "pull confirm result", c.Exec("pull"))
	assert.Equal(t, "pull result", c.Exec("yes"))
}

// **確認待ちを跨がせない。**別のコマンドを打った後の "y" が、忘れたころに
// Pull を確定させてしまう事故を防ぐ。
func TestLetItRideCuiController_Pull_AnyOtherCommandCancels(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	c.Exec("p")
	assert.Equal(t, "letitride result", c.Exec("l"))
	assert.Contains(t, c.Exec("y"), msgStem("letitride.nothingToConfirm"))
	m.AssertNotCalled(t, "Pull")
}

// **リセットも確認待ちを消す** (#6076)。r / reset は execCuiCommand が
// gameHandler より先に拾うので、gameHandler 側のクリアだけでは待ちが残り、
// 配り直した卓に Pull が走っていた。
func TestLetItRideCuiController_Pull_ResetCancels(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	c.Exec("p")
	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Contains(t, c.Exec("y"), msgStem("letitride.nothingToConfirm"))
	m.AssertNotCalled(t, "Pull")
}

func TestLetItRideCuiController_Confirm_WithoutPendingPull(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	assert.Contains(t, c.Exec("y"), msgStem("letitride.nothingToConfirm"))
	m.AssertNotCalled(t, "Pull")
}

// **既存コマンドは何も変わらない。**
func TestLetItRideCuiController_OtherCommandsUnaffected(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	assert.Equal(t, "bet result", c.Exec("b 100"))
	assert.Equal(t, "letitride result", c.Exec("l"))
	assert.Equal(t, "action log result", c.Exec("log"))
	m.AssertNotCalled(t, "Pull")
	m.AssertNotCalled(t, "PullConfirm")
}

func TestLetItRideCuiController_LetItRide(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	assert.Equal(t, "letitride result", c.Exec("l"))
	assert.Equal(t, "letitride result", c.Exec("letitride"))
}

func TestLetItRideCuiController_ActionLog(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	assert.Equal(t, "action log result", c.Exec("log"))
}

func TestLetItRideCuiController_Unknown(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	result := c.Exec("xyz")
	assert.Contains(t, result, "コマンドが不明です")
}

func TestLetItRideCuiController_Empty(t *testing.T) {
	m := newMockLetItRideInteractor()
	c := controller.NewLetItRideCuiController(m)

	result := c.Exec("")
	assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
}
