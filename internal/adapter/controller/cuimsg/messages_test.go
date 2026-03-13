package cuimsg_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuimsg"
)

func TestRequired(t *testing.T) {
	assert.Equal(t, "Bet amount is required.", cuimsg.Required("Bet amount"))
	assert.Equal(t, "CPU player count is required.", cuimsg.Required("CPU player count"))
}

func TestInvalidNotANumber(t *testing.T) {
	assert.Equal(t, "Invalid bet amount. Please enter a number.", cuimsg.InvalidNotANumber("bet amount"))
	assert.Equal(t, "Invalid deck count. Please enter a number.", cuimsg.InvalidNotANumber("deck count"))
}

func TestRequiredWithHint(t *testing.T) {
	assert.Equal(t, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", cuimsg.RequiredWithHint("CPU difficulty", "(0=Easy, 1=Normal, 2=Hard)"))
	assert.Equal(t, "Meta-AI flag is required (0=OFF, 1=ON).", cuimsg.RequiredWithHint("Meta-AI flag", "(0=OFF, 1=ON)"))
}

func TestInvalidValue(t *testing.T) {
	assert.Equal(t, "Invalid card index: abc.", cuimsg.InvalidValue("card index", "abc"))
	assert.Equal(t, "Invalid position: 99.", cuimsg.InvalidValue("position", "99"))
}

func TestInvalidOutOfRange(t *testing.T) {
	assert.Equal(t, "Invalid betting limit: abc. Please enter 0-2.", cuimsg.InvalidOutOfRange("betting limit", "abc", "Please enter 0-2."))
	assert.Equal(t, "Invalid CPU difficulty: 5. Please enter 0-2.", cuimsg.InvalidOutOfRange("CPU difficulty", "5", "Please enter 0-2."))
}
