//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestViraCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockViraInteractor {
		m := new(mockUsecases.MockViraInteractor)
		m.On("GetConfig").Return(domain.DefaultViraConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewViraCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewViraCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewViraCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultViraConfig())
	})

	t.Run("every rung of the bid ladder is accepted", func(t *testing.T) {
		for _, bid := range []domain.ViraBid{
			domain.ViraBidPass, domain.ViraBidGask, domain.ViraBidSolo,
			domain.ViraBidMisere, domain.ViraBidVira,
		} {
			m := newMock()
			c := controller.NewViraCuiController(m)
			assert.Equal(t, mockOutput, c.Exec("bid "+string(rune('0'+int(bid)))))
			m.AssertCalled(t, "Bid", int(bid))
		}
	})

	t.Run("bid no args names the Vira ladder, not another game's", func(t *testing.T) {
		result := controller.NewViraCuiController(newMock()).Exec("bid")
		assert.Contains(t, result, "Bid is required")
		for _, want := range []string{"Gask", "Solo", "Vira"} {
			assert.Contains(t, result, want)
		}
		// Préférence's rungs — the prompt was copied from it and named these.
		for _, wrong := range []string{"Six", "Seven", "Eight"} {
			assert.NotContains(t, result, wrong, "%q belongs to Préférence, not Vira", wrong)
		}
	})

	t.Run("bid out of range", func(t *testing.T) {
		result := controller.NewViraCuiController(newMock()).Exec("bid 9")
		assert.Contains(t, result, "Invalid bid")
	})

	t.Run("pass maps to bid 0", func(t *testing.T) {
		m := newMock()
		c := controller.NewViraCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("pass"))
		m.AssertCalled(t, "Bid", int(domain.ViraBidPass))
	})

	t.Run("play card", func(t *testing.T) {
		m := newMock()
		c := controller.NewViraCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		result := controller.NewViraCuiController(newMock()).Exec("play")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewViraCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewViraCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultViraConfig()
		expected.CpuDifficulty = domain.ViraCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		result := controller.NewViraCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, "Invalid CPU difficulty")
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewViraCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		result := controller.NewViraCuiController(newMock()).Exec("zzz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
