//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	ucmock "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newBuraCui() (*controller.BuraCuiController, *ucmock.MockBuraInteractor) {
	bi := new(ucmock.MockBuraInteractor)
	return controller.NewBuraCuiController(bi), bi
}

func TestBuraCuiController_PlayParsesOneToThreeIndices(t *testing.T) {
	c, bi := newBuraCui()
	bi.On("Play", []int{0, 2}).Return("played")

	assert.Equal(t, "played", c.Exec("p 0 2"))
	bi.AssertCalled(t, "Play", []int{0, 2})
}

func TestBuraCuiController_PlayRejectsBadInput(t *testing.T) {
	cases := map[string]string{
		"p":     msgCardIndexRequired(),
		"p x":   msgInvalidCardIndex("x"),
		"p 1 1": "Duplicate card index: 1.",
	}
	for input, want := range cases {
		t.Run(input, func(t *testing.T) {
			c, bi := newBuraCui()
			assert.Contains(t, c.Exec(input), want)
			// A rejected parse must not reach the interactor -- a duplicate
			// index would otherwise read as a longer play than the hand holds.
			bi.AssertNotCalled(t, "Play", mock.Anything)
		})
	}
}

func TestBuraCuiController_ClaimAndDeclare(t *testing.T) {
	c, bi := newBuraCui()
	bi.On("Claim").Return("claimed")
	bi.On("Declare").Return("declared")

	assert.Equal(t, "claimed", c.Exec("c"))
	assert.Equal(t, "claimed", c.Exec("claim"))
	assert.Equal(t, "declared", c.Exec("d"))
	assert.Equal(t, "declared", c.Exec("declare"))
}

func TestBuraCuiController_ResetKeepsTheCurrentConfig(t *testing.T) {
	c, bi := newBuraCui()
	cfg := domain.DefaultBuraConfig()
	bi.On("GetConfig").Return(cfg)
	bi.On("ResetWithConfig", cfg).Return("reset")

	assert.Equal(t, "reset", c.Exec("r"))
	bi.AssertCalled(t, "ResetWithConfig", cfg)
}

func TestBuraCuiController_HintAndLog(t *testing.T) {
	c, bi := newBuraCui()
	bi.On("Hint").Return("hint")
	bi.On("ActionLog").Return("log")

	assert.Equal(t, "hint", c.Exec("h"))
	assert.Equal(t, "log", c.Exec("log"))
	assert.Equal(t, "log", c.Exec("l"))
}

func TestBuraCuiController_QuitAndUnknownCommand(t *testing.T) {
	c, _ := newBuraCui()
	assert.Contains(t, c.Exec("q"), "bye")
	// A near miss should suggest rather than just reject.
	assert.NotEmpty(t, c.Exec("clam"))
	assert.NotEmpty(t, c.Exec("zzzz"))
}
