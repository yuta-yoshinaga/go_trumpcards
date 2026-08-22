//go:build test && (!js || !wasm || extra)

package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// newSpeculationInteractorFixture wires a mock game and mock presenter into a
// real interactor. `ended` drives GetGameEndFlag, which is the only state
// runGuarded reads before it lets an action through.
func newSpeculationInteractorFixture(ended bool) (*interfaces.MockSpeculationGame, *presenter.MockSpeculationPresenter, *SpeculationInteractor) {
	sg := new(interfaces.MockSpeculationGame)
	sp := new(presenter.MockSpeculationPresenter)
	sg.On("GetGameEndFlag").Return(ended).Maybe()
	return sg, sp, NewSpeculationInteractor(sg, sp)
}

func TestNewSpeculationInteractor(t *testing.T) {
	_, _, si := newSpeculationInteractorFixture(false)
	assert.NotNil(t, si)
}

func TestNewSpeculationInteractorPanicsOnNil(t *testing.T) {
	sp := new(presenter.MockSpeculationPresenter)
	assert.Panics(t, func() { NewSpeculationInteractor(nil, sp) })
	sg := new(interfaces.MockSpeculationGame)
	assert.Panics(t, func() { NewSpeculationInteractor(sg, nil) })
}

func TestSpeculationInteractorReset(t *testing.T) {
	sg, sp, si := newSpeculationInteractorFixture(false)

	sg.On("Reset").Return()
	sp.On("Output", sg, nil).Return("reset_output")

	assert.Equal(t, "reset_output", si.Reset())
	sg.AssertCalled(t, "Reset")
}

// TestSpeculationInteractorResetIsNotGuarded pins that a finished game can
// still be dealt again -- Reset is how the player starts over, so guarding it
// would leave the table permanently dead.
func TestSpeculationInteractorResetIsNotGuarded(t *testing.T) {
	sg, sp, si := newSpeculationInteractorFixture(true)

	sg.On("Reset").Return()
	sp.On("Output", sg, nil).Return("reset_output")

	assert.Equal(t, "reset_output", si.Reset())
	sg.AssertCalled(t, "Reset")
}

func TestSpeculationInteractorFlip(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg, sp, si := newSpeculationInteractorFixture(false)

		sg.On("Flip").Return(nil)
		sp.On("Output", sg, nil).Return("flip_output")

		assert.Equal(t, "flip_output", si.Flip())
		sg.AssertCalled(t, "Flip")
	})

	t.Run("error surfaces", func(t *testing.T) {
		sg, sp, si := newSpeculationInteractorFixture(false)

		err := errors.New("not allowed in this phase")
		sg.On("Flip").Return(err)
		sp.On("Output", sg, err).Return("flip_error_output")

		assert.Equal(t, "flip_error_output", si.Flip())
		// The presenter must receive the very error the domain returned;
		// swallowing it would leave the player with no idea why nothing moved.
		sp.AssertCalled(t, "Output", sg, err)
	})
}

func TestSpeculationInteractorAccept(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg, sp, si := newSpeculationInteractorFixture(false)

		sg.On("Accept").Return(nil)
		sp.On("Output", sg, nil).Return("accept_output")

		assert.Equal(t, "accept_output", si.Accept())
		sg.AssertCalled(t, "Accept")
	})

	t.Run("error surfaces", func(t *testing.T) {
		sg, sp, si := newSpeculationInteractorFixture(false)

		err := errors.New("there is no offer to answer")
		sg.On("Accept").Return(err)
		sp.On("Output", sg, err).Return("accept_error_output")

		assert.Equal(t, "accept_error_output", si.Accept())
		sp.AssertCalled(t, "Output", sg, err)
	})
}

func TestSpeculationInteractorDecline(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg, sp, si := newSpeculationInteractorFixture(false)

		sg.On("Decline").Return(nil)
		sp.On("Output", sg, nil).Return("decline_output")

		assert.Equal(t, "decline_output", si.Decline())
		sg.AssertCalled(t, "Decline")
	})

	t.Run("error surfaces", func(t *testing.T) {
		sg, sp, si := newSpeculationInteractorFixture(false)

		err := errors.New("there is no offer to answer")
		sg.On("Decline").Return(err)
		sp.On("Output", sg, err).Return("decline_error_output")

		assert.Equal(t, "decline_error_output", si.Decline())
		sp.AssertCalled(t, "Output", sg, err)
	})
}

func TestSpeculationInteractorBid(t *testing.T) {
	t.Run("forwards the amount", func(t *testing.T) {
		sg, sp, si := newSpeculationInteractorFixture(false)

		// **The amount must reach the domain unchanged.** A raise that arrives
		// as some other number is a different bid than the player made.
		sg.On("Bid", 57).Return(nil)
		sp.On("Output", sg, nil).Return("bid_output")

		assert.Equal(t, "bid_output", si.Bid(57))
		sg.AssertCalled(t, "Bid", 57)
		sg.AssertNotCalled(t, "Bid", 0)
	})

	t.Run("a different amount is a different call", func(t *testing.T) {
		sg, sp, si := newSpeculationInteractorFixture(false)

		sg.On("Bid", 12).Return(nil)
		sp.On("Output", sg, nil).Return("bid_output")

		assert.Equal(t, "bid_output", si.Bid(12))
		sg.AssertCalled(t, "Bid", 12)
		sg.AssertNotCalled(t, "Bid", 57)
	})

	t.Run("error surfaces", func(t *testing.T) {
		sg, sp, si := newSpeculationInteractorFixture(false)

		err := errors.New("invalid amount")
		sg.On("Bid", 3).Return(err)
		sp.On("Output", sg, err).Return("bid_error_output")

		assert.Equal(t, "bid_error_output", si.Bid(3))
		sp.AssertCalled(t, "Output", sg, err)
	})
}

func TestSpeculationInteractorNextRound(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg, sp, si := newSpeculationInteractorFixture(false)

		sg.On("NextRound").Return(nil)
		sp.On("Output", sg, nil).Return("next_output")

		assert.Equal(t, "next_output", si.NextRound())
		sg.AssertCalled(t, "NextRound")
	})

	t.Run("error surfaces", func(t *testing.T) {
		sg, sp, si := newSpeculationInteractorFixture(false)

		err := errors.New("not allowed in this phase")
		sg.On("NextRound").Return(err)
		sp.On("Output", sg, err).Return("next_error_output")

		assert.Equal(t, "next_error_output", si.NextRound())
		sp.AssertCalled(t, "Output", sg, err)
	})
}

// TestSpeculationInteractorGuardsAfterGameEnd pins runGuarded: once the game is
// over, every play command presents the final board and **nothing reaches the
// domain**. Asserting only the returned string would pass even if the action
// still ran, so each case also asserts the domain method was never called.
func TestSpeculationInteractorGuardsAfterGameEnd(t *testing.T) {
	cases := []struct {
		name   string
		method string
		run    func(si *SpeculationInteractor) string
	}{
		{"flip", "Flip", func(si *SpeculationInteractor) string { return si.Flip() }},
		{"accept", "Accept", func(si *SpeculationInteractor) string { return si.Accept() }},
		{"decline", "Decline", func(si *SpeculationInteractor) string { return si.Decline() }},
		{"bid", "Bid", func(si *SpeculationInteractor) string { return si.Bid(40) }},
		{"nextRound", "NextRound", func(si *SpeculationInteractor) string { return si.NextRound() }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sg, sp, si := newSpeculationInteractorFixture(true)
			sp.On("Output", sg, nil).Return("game_over_output")

			assert.Equal(t, "game_over_output", tc.run(si))

			// Not "no call with these arguments" but "no call at all": the
			// only thing the interactor may touch is the end flag itself.
			for _, call := range sg.Calls {
				assert.Equal(t, "GetGameEndFlag", call.Method,
					"%s reached the domain after the game ended", tc.method)
			}
			sg.AssertCalled(t, "GetGameEndFlag")
		})
	}
}

func TestSpeculationInteractorHint(t *testing.T) {
	sg, sp, si := newSpeculationInteractorFixture(false)

	sp.On("HintOutput", mock.Anything).Return("hint_output")

	assert.Equal(t, "hint_output", si.Hint())
	sp.AssertCalled(t, "HintOutput", sg)
}

// TestSpeculationInteractorHintIsNotGuarded pins that advice is still available
// after the last round -- the hint reads the board, it does not change it.
func TestSpeculationInteractorHintIsNotGuarded(t *testing.T) {
	sg, sp, si := newSpeculationInteractorFixture(true)

	sp.On("HintOutput", sg).Return("hint_output")

	assert.Equal(t, "hint_output", si.Hint())
}

func TestSpeculationInteractorActionLog(t *testing.T) {
	sg, sp, si := newSpeculationInteractorFixture(false)

	sp.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "log_output", si.ActionLog())
	sp.AssertCalled(t, "ActionLogOutput", sg)
}
