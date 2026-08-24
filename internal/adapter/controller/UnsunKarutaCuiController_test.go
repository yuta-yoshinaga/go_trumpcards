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

func TestUnsunKarutaCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockUnsunKarutaInteractor {
		m := new(mockUsecases.MockUnsunKarutaInteractor)
		m.On("GetConfig").Return(domain.DefaultUnsunKarutaConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewUnsunKarutaCuiController(newMock()).Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewUnsunKarutaCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultUnsunKarutaConfig())
	})

	// **宣言は札と一緒に送る。** 別命令にすると「宣言したが札を出していない」
	// 盤面が生まれる。
	// 短縮形も同じ意味 — Web の CLI が `p` / `m` を受けるので、CUI で弾かれると
	// 同じ手が端末によって通ったり通らなかったりする。
	for _, cmd := range []string{"play", "p"} {
		t.Run(cmd+" does not declare", func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewUnsunKarutaCuiController(m).Exec(cmd+" 3"))
			m.AssertCalled(t, "Play", 3, false)
		})
	}

	for _, cmd := range []string{"meri", "monchi", "m"} {
		t.Run(cmd+" declares", func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewUnsunKarutaCuiController(m).Exec(cmd+" 2"))
			m.AssertCalled(t, "Play", 2, true)
		})
	}

	t.Run("play needs an index", func(t *testing.T) {
		m := newMock()
		out := controller.NewUnsunKarutaCuiController(m).Exec("play")
		assert.Contains(t, out, msgStem("cardIndexRequired"))
		m.AssertNotCalled(t, "Play", mock.Anything, mock.Anything)
	})

	t.Run("play rejects a non-numeric index", func(t *testing.T) {
		m := newMock()
		out := controller.NewUnsunKarutaCuiController(m).Exec("play x")
		assert.Contains(t, out, msgStem("invalidCardIndex"))
		m.AssertNotCalled(t, "Play", mock.Anything, mock.Anything)
	})

	t.Run("next and nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewUnsunKarutaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("set difficulty", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewUnsunKarutaCuiController(m).Exec("sd 2"))
		cfg := domain.DefaultUnsunKarutaConfig()
		cfg.CpuDifficulty = domain.UnsunKarutaCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})

	// **ディール数は 1〜8。** 8 人が 1 回ずつ親を務めて 1 巡なので、それ以上は
	// 同じ席が二度親になる。
	t.Run("set deals", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewUnsunKarutaCuiController(m).Exec("st 8"))
		cfg := domain.DefaultUnsunKarutaConfig()
		cfg.TargetDeals = 8
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})

	for _, deals := range []string{"0", "9", "99"} {
		t.Run("set deals rejects "+deals, func(t *testing.T) {
			m := newMock()
			out := controller.NewUnsunKarutaCuiController(m).Exec("st " + deals)
			assert.Contains(t, out, msgStem("invalidNumberOfDeals18"))
			m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
		})
	}

	t.Run("hint and log", func(t *testing.T) {
		m := newMock()
		c := controller.NewUnsunKarutaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("l"))
	})

	t.Run("unknown command suggests a close one", func(t *testing.T) {
		out := controller.NewUnsunKarutaCuiController(newMock()).Exec("playy")
		assert.NotEqual(t, mockOutput, out)
		assert.Contains(t, out, "play")
	})
}
