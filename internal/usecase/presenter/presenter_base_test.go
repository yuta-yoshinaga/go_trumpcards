package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// compile-time interface satisfaction checks
var (
	_ presenter.BlackJackPresenter = (*presenter.MockBlackJackPresenter)(nil)
	_ presenter.OldMaidPresenter   = (*presenter.MockOldMaidPresenter)(nil)
	_ presenter.DaifugoPresenter   = (*presenter.MockDaifugoPresenter)(nil)
	_ presenter.SevensPresenter    = (*presenter.MockSevensPresenter)(nil)
	_ presenter.DoubtPresenter     = (*presenter.MockDoubtPresenter)(nil)
	_ presenter.HoldemPresenter    = (*presenter.MockHoldemPresenter)(nil)
	_ presenter.HeartsPresenter    = (*presenter.MockHeartsPresenter)(nil)
	_ presenter.MemoryPresenter    = (*presenter.MockMemoryPresenter)(nil)
	_ presenter.BaccaratPresenter  = (*presenter.MockBaccaratPresenter)(nil)
	_ presenter.PokerPresenter     = (*presenter.MockPokerPresenter)(nil)
	_ presenter.KlondikePresenter  = (*presenter.MockKlondikePresenter)(nil)
)

func TestMockGamePresenter_Output(t *testing.T) {
	m := new(presenter.MockGamePresenter[interfaces.BlackJackGame])
	mg := new(interfaces.MockBlackJackGame)
	err := errors.New("test error")
	m.On("Output", mg, err).Return("output text")

	result := m.Output(mg, err)

	assert.Equal(t, "output text", result)
	m.AssertExpectations(t)
}

func TestMockGamePresenter_Output_NilError(t *testing.T) {
	m := new(presenter.MockGamePresenter[interfaces.BlackJackGame])
	mg := new(interfaces.MockBlackJackGame)
	m.On("Output", mg, mock.Anything).Return("no error output")

	result := m.Output(mg, nil)

	assert.Equal(t, "no error output", result)
	m.AssertExpectations(t)
}

func TestMockGamePresenter_ActionLogOutput(t *testing.T) {
	m := new(presenter.MockGamePresenter[interfaces.BlackJackGame])
	mg := new(interfaces.MockBlackJackGame)
	m.On("ActionLogOutput", mg).Return("action log")

	result := m.ActionLogOutput(mg)

	assert.Equal(t, "action log", result)
	m.AssertExpectations(t)
}

func TestMockGamePresenter_SatisfiesInterface(t *testing.T) {
	tests := []struct {
		name string
		fn   func()
	}{
		{"BlackJack", func() {
			var _ presenter.GamePresenter[interfaces.BlackJackGame] = new(presenter.MockGamePresenter[interfaces.BlackJackGame])
		}},
		{"OldMaid", func() {
			var _ presenter.GamePresenter[interfaces.OldMaidGame] = new(presenter.MockGamePresenter[interfaces.OldMaidGame])
		}},
		{"Daifugo", func() {
			var _ presenter.GamePresenter[interfaces.DaifugoGame] = new(presenter.MockGamePresenter[interfaces.DaifugoGame])
		}},
		{"Sevens", func() {
			var _ presenter.GamePresenter[interfaces.SevensGame] = new(presenter.MockGamePresenter[interfaces.SevensGame])
		}},
		{"Doubt", func() {
			var _ presenter.GamePresenter[interfaces.DoubtGame] = new(presenter.MockGamePresenter[interfaces.DoubtGame])
		}},
		{"Holdem", func() {
			var _ presenter.GamePresenter[interfaces.HoldemGame] = new(presenter.MockGamePresenter[interfaces.HoldemGame])
		}},
		{"Hearts", func() {
			var _ presenter.GamePresenter[interfaces.HeartsGame] = new(presenter.MockGamePresenter[interfaces.HeartsGame])
		}},
		{"Memory", func() {
			var _ presenter.GamePresenter[interfaces.MemoryGame] = new(presenter.MockGamePresenter[interfaces.MemoryGame])
		}},
		{"Baccarat", func() {
			var _ presenter.GamePresenter[interfaces.BaccaratGame] = new(presenter.MockGamePresenter[interfaces.BaccaratGame])
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, tt.fn)
		})
	}
}
