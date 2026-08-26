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

func TestSchafkopfCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockSchafkopfInteractor {
		m := new(mockUsecases.MockSchafkopfInteractor)
		m.On("GetConfig").Return(domain.DefaultSchafkopfConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Declare", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Call", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewSchafkopfCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewSchafkopfCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewSchafkopfCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultSchafkopfConfig())
	})

	// **契約ごとに別のコマンド。**どれか 1 つでも Rufspiel に落ちると、
	// 宣言した契約と盤面の切り札が食い違う。
	for _, tc := range []struct {
		name     string
		input    string
		contract domain.SchafkopfContract
		soloSuit int
	}{
		{"pick declares rufspiel", "pick", domain.SchafkopfContractRufspiel, 0},
		{"pick shorthand p", "p", domain.SchafkopfContractRufspiel, 0},
		{"wenz", "wenz", domain.SchafkopfContractWenz, 0},
		{"wenz shorthand w", "w", domain.SchafkopfContractWenz, 0},
		{"solo carries its suit", "solo 3", domain.SchafkopfContractSolo, domain.CardDesignHeart},
		{"solo shorthand so", "so 4", domain.SchafkopfContractSolo, domain.CardDesignDiamond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newMock()
			c := controller.NewSchafkopfCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(tc.input))
			m.AssertCalled(t, "Declare", true, tc.contract, tc.soloSuit)
		})
	}

	t.Run("pass", func(t *testing.T) {
		m := newMock()
		c := controller.NewSchafkopfCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("pass"))
		m.AssertCalled(t, "Declare", false, domain.SchafkopfContractRufspiel, 0)
	})

	// Solo は切り札スートを要る。既定値に落ちると、指定を忘れた宣言が
	// 黙って別のスートの Solo になる。
	for _, input := range []string{"solo", "so x", "solo 0", "solo 5"} {
		t.Run("solo rejects "+input, func(t *testing.T) {
			m := newMock()
			c := controller.NewSchafkopfCuiController(m)
			c.Exec(input)
			m.AssertNotCalled(t, "Declare", mock.Anything, mock.Anything, mock.Anything)
		})
	}

	t.Run("call suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewSchafkopfCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("c 2"))
		m.AssertCalled(t, "Call", 2)
	})

	t.Run("call alias", func(t *testing.T) {
		m := newMock()
		c := controller.NewSchafkopfCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("call 1"))
		m.AssertCalled(t, "Call", 1)
	})

	t.Run("call no args", func(t *testing.T) {
		result := controller.NewSchafkopfCuiController(newMock()).Exec("c")
		assert.Contains(t, result, msgStem("suitRequiredThree"))
	})

	t.Run("call invalid suit", func(t *testing.T) {
		result := controller.NewSchafkopfCuiController(newMock()).Exec("c 9")
		assert.Contains(t, result, msgStem("invalidSuitThree"))
	})

	t.Run("play card", func(t *testing.T) {
		m := newMock()
		c := controller.NewSchafkopfCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		result := controller.NewSchafkopfCuiController(newMock()).Exec("play")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewSchafkopfCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewSchafkopfCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultSchafkopfConfig()
		expected.CpuDifficulty = domain.SchafkopfCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		result := controller.NewSchafkopfCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setchips", func(t *testing.T) {
		m := newMock()
		c := controller.NewSchafkopfCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sb 5"))
		expected := domain.DefaultSchafkopfConfig()
		expected.BaseChips = 5
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setchips invalid", func(t *testing.T) {
		result := controller.NewSchafkopfCuiController(newMock()).Exec("sb 0")
		assert.Contains(t, result, msgStem("invalidBaseChips1OrMore"))
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewSchafkopfCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		result := controller.NewSchafkopfCuiController(newMock()).Exec("zzz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
