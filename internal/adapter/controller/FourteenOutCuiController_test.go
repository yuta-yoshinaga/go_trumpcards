package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockFourteenOutInteractor() *mockusecase.MockFourteenOutInteractor {
	return new(mockusecase.MockFourteenOutInteractor)
}

func TestFourteenOutCuiController_Quit(t *testing.T) {
	mi := newMockFourteenOutInteractor()
	c := NewFourteenOutCuiController(mi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestFourteenOutCuiController_Reset(t *testing.T) {
	mi := newMockFourteenOutInteractor()
	c := NewFourteenOutCuiController(mi)
	mi.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestFourteenOutCuiController_Remove(t *testing.T) {
	mi := newMockFourteenOutInteractor()
	c := NewFourteenOutCuiController(mi)
	// **列番号 2 つで足りる。**クローン元の Monte Carlo は (行,列) を 4 つ取るが、
	// Fourteen Out で動かせるのは各列の末尾だけなので一意に決まる。
	mi.On("Remove", 0, 3).Return("remove_output")
	assert.Equal(t, "remove_output", c.Exec("m 0 3"))
	assert.Equal(t, "remove_output", c.Exec("move 0 3"))
	assert.Equal(t, "remove_output", c.Exec("remove 0 3"))
}

func TestFourteenOutCuiController_RemoveInvalid(t *testing.T) {
	mi := newMockFourteenOutInteractor()
	c := NewFourteenOutCuiController(mi)

	assert.Contains(t, c.Exec("m"), msgUsage("fourteenout.usageMC1C2"))
	assert.Contains(t, c.Exec("m 0"), msgUsage("fourteenout.usageMC1C2"))
	// **4 引数はクローン元の文法。**受け付けてしまうと、余った引数が黙って
	// 捨てられる。
	assert.Contains(t, c.Exec("m 0 1 1 2"), msgUsage("fourteenout.usageMC1C2"))
	assert.True(t, msgRejected(c.Exec("m abc 3")))
	assert.True(t, msgRejected(c.Exec("m 0 zzz")))
}

// **補充コマンドは存在しない。**山札が無いので、受け付けると盤が変わらない
// 無言の no-op になる。
func TestFourteenOutCuiController_DealIsNotACommand(t *testing.T) {
	mi := newMockFourteenOutInteractor()
	c := NewFourteenOutCuiController(mi)
	assert.True(t, msgRejected(c.Exec("d")))
	assert.True(t, msgRejected(c.Exec("deal")))
}

func TestFourteenOutCuiController_Undo(t *testing.T) {
	mi := newMockFourteenOutInteractor()
	c := NewFourteenOutCuiController(mi)
	mi.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestFourteenOutCuiController_GiveUp(t *testing.T) {
	mi := newMockFourteenOutInteractor()
	c := NewFourteenOutCuiController(mi)
	mi.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestFourteenOutCuiController_Hint(t *testing.T) {
	mi := newMockFourteenOutInteractor()
	c := NewFourteenOutCuiController(mi)
	mi.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestFourteenOutCuiController_ActionLog(t *testing.T) {
	mi := newMockFourteenOutInteractor()
	c := NewFourteenOutCuiController(mi)
	mi.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("l"))
	assert.Equal(t, "log_output", c.Exec("log"))
}
