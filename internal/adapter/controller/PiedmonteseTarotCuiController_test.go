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

func TestPiedmonteseTarotCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":1}`

	newMock := func(cfg domain.PiedmonteseTarotConfig) *mockUsecases.MockPiedmonteseTarotInteractor {
		m := new(mockUsecases.MockPiedmonteseTarotInteractor)
		m.On("GetConfig").Return(cfg)
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}
	fourHanded := domain.DefaultPiedmonteseTarotConfig()
	threeHanded := domain.DefaultPiedmonteseTarotConfig()
	threeHanded.Seats = 3

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewPiedmonteseTarotCuiController(newMock(fourHanded)).Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock(fourHanded)
		assert.Equal(t, mockOutput, controller.NewPiedmonteseTarotCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", fourHanded)
	})

	// **枚数は卓が決める。** 4 人卓は 2 枚、3 人卓は 3 枚 ── 固定で検査すると
	// 片方の卓では必ず弾かれる。
	t.Run("a four-handed table buries two", func(t *testing.T) {
		m := newMock(fourHanded)
		c := controller.NewPiedmonteseTarotCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("scarto 0 1"))
		m.AssertCalled(t, "Discard", []int{0, 1})
	})

	t.Run("a three-handed table refuses two and takes three", func(t *testing.T) {
		m := newMock(threeHanded)
		c := controller.NewPiedmonteseTarotCuiController(m)
		out := c.Exec("scarto 0 1")
		assert.NotEqual(t, mockOutput, out, "2 枚で通ってしまう")
		m.AssertNotCalled(t, "Discard", mock.Anything)

		assert.Equal(t, mockOutput, c.Exec("discard 0 1 2"))
		m.AssertCalled(t, "Discard", []int{0, 1, 2})
	})

	t.Run("scarto rejects a non-numeric index", func(t *testing.T) {
		m := newMock(fourHanded)
		out := controller.NewPiedmonteseTarotCuiController(m).Exec("scarto 0 x")
		assert.Contains(t, out, msgStem("invalidCardIndex"))
		m.AssertNotCalled(t, "Discard", mock.Anything)
	})

	t.Run("play", func(t *testing.T) {
		m := newMock(fourHanded)
		assert.Equal(t, mockOutput, controller.NewPiedmonteseTarotCuiController(m).Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play needs an index", func(t *testing.T) {
		m := newMock(fourHanded)
		out := controller.NewPiedmonteseTarotCuiController(m).Exec("play")
		assert.Contains(t, out, msgStem("cardIndexRequired"))
		m.AssertNotCalled(t, "Play", mock.Anything)
	})

	t.Run("next and nextround", func(t *testing.T) {
		m := newMock(fourHanded)
		c := controller.NewPiedmonteseTarotCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	// **席数は 3 か 4 だけ。** 配り方が決まっているのはその 2 つで、他の数には
	// 配りが存在しない。
	for _, seats := range []int{3, 4} {
		t.Run("set seats accepts a table the rules deal for", func(t *testing.T) {
			m := newMock(fourHanded)
			assert.Equal(t, mockOutput,
				controller.NewPiedmonteseTarotCuiController(m).Exec("ss "+string(rune('0'+seats))))
			cfg := fourHanded
			cfg.Seats = seats
			m.AssertCalled(t, "ResetWithConfig", cfg)
		})
	}
	for _, seats := range []string{"2", "5", "9"} {
		t.Run("set seats rejects "+seats, func(t *testing.T) {
			m := newMock(fourHanded)
			out := controller.NewPiedmonteseTarotCuiController(m).Exec("ss " + seats)
			assert.Contains(t, out, msgKey("invalidNumberOfSeats34"))
			m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
		})
	}
	t.Run("set seats with no argument asks for one", func(t *testing.T) {
		m := newMock(fourHanded)
		out := controller.NewPiedmonteseTarotCuiController(m).Exec("ss")
		assert.Equal(t, msgKey("numberOfSeatsRequired34"), out)
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("set difficulty", func(t *testing.T) {
		m := newMock(fourHanded)
		assert.Equal(t, mockOutput, controller.NewPiedmonteseTarotCuiController(m).Exec("sd 2"))
		cfg := fourHanded
		cfg.CpuDifficulty = domain.PiedmonteseTarotCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})

	t.Run("hint and log", func(t *testing.T) {
		m := newMock(fourHanded)
		c := controller.NewPiedmonteseTarotCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("l"))
	})

	t.Run("unknown command suggests a close one", func(t *testing.T) {
		out := controller.NewPiedmonteseTarotCuiController(newMock(fourHanded)).Exec("playy")
		assert.NotEqual(t, mockOutput, out)
		assert.Contains(t, out, "play")
	})
}
