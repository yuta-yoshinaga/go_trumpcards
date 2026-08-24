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

func TestGleekCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockGleekInteractor {
		m := new(mockUsecases.MockGleekInteractor)
		m.On("GetConfig").Return(domain.DefaultGleekConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewGleekCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewGleekCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewGleekCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultGleekConfig())
	})

	t.Run("bid raises to the given amount", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewGleekCuiController(m).Exec("bid 14"))
		m.AssertCalled(t, "Bid", 14)
	})

	t.Run("bid pass drops out", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewGleekCuiController(m).Exec("bid pass"))
		m.AssertCalled(t, "Bid", 0)
	})

	t.Run("bid rejects a missing or non-numeric amount", func(t *testing.T) {
		assert.Contains(t, controller.NewGleekCuiController(newMock()).Exec("bid"), msgStem("bidAmountRequired"))
		assert.Contains(t, controller.NewGleekCuiController(newMock()).Exec("bid zzz"), msgStem("invalidBidAmount"))
	})

	// **捨て札フェーズを抜ける唯一の入力。** これを落とすと落札の直後で盤面が
	// 固まり、play は「フェーズが違う」で弾かれ続ける。
	t.Run("discard passes every index through", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewGleekCuiController(m).Exec("discard 0 1 2 3 4 5 6"))
		m.AssertCalled(t, "Discard", []int{0, 1, 2, 3, 4, 5, 6})
	})

	t.Run("discard accepts commas and the short form", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewGleekCuiController(m).Exec("d 0,1,2"))
		m.AssertCalled(t, "Discard", []int{0, 1, 2})
	})

	t.Run("discard rejects a missing or non-numeric index", func(t *testing.T) {
		missing := controller.NewGleekCuiController(newMock()).Exec("discard")
		assert.Contains(t, missing, msgStem("discardIndicesRequired"))
		// **キーが訳に解決していることまで見る。** i18n.T は未翻訳のキーを
		// そのまま返すので、Contains(out, T(key)) は鍵が無くても通る。
		assert.NotContains(t, missing, "discardIndicesRequired", "生のキーが出てはいけない")
		assert.Contains(t, controller.NewGleekCuiController(newMock()).Exec("d 0 x"), msgStem("invalidCardIndex"))
	})

	t.Run("play card", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewGleekCuiController(m).Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		assert.Contains(t, controller.NewGleekCuiController(newMock()).Exec("play"), msgCardIndexRequired())
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewGleekCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewGleekCuiController(m).Exec("sd 2"))
		expected := domain.DefaultGleekConfig()
		expected.CpuDifficulty = domain.GleekCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		assert.Contains(t, controller.NewGleekCuiController(newMock()).Exec("sd 9"), msgInvalidCpuDifficultyPrefix())
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewGleekCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		assert.Contains(t, controller.NewGleekCuiController(newMock()).Exec("zzz"), "コマンドが不明です")
	})
}
