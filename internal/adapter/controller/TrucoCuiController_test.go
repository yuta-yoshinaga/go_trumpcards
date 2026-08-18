//go:build test

package controller_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestTrucoCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`
	newMock := func() *mockUsecases.MockTrucoInteractor {
		m := new(mockUsecases.MockTrucoInteractor)
		m.On("GetConfig").Return(domain.DefaultTrucoConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("Truco").Return(mockOutput)
		m.On("Respond", mock.Anything).Return(mockOutput)
		m.On("Next").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewTrucoCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewTrucoCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		got := controller.NewTrucoCuiController(m).Exec("r")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultTrucoConfig())
	})

	t.Run("play with index", func(t *testing.T) {
		m := newMock()
		got := controller.NewTrucoCuiController(m).Exec("p 1")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Play", 1)
	})

	t.Run("play missing index", func(t *testing.T) {
		got := controller.NewTrucoCuiController(newMock()).Exec("p")
		assert.Contains(t, got, msgCardIndexRequired())
	})

	t.Run("play invalid index", func(t *testing.T) {
		got := controller.NewTrucoCuiController(newMock()).Exec("p abc")
		assert.Contains(t, got, msgInvalidCardIndexPrefix())
	})

	t.Run("truco", func(t *testing.T) {
		m := newMock()
		got := controller.NewTrucoCuiController(m).Exec("t")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Truco")
	})

	t.Run("accept", func(t *testing.T) {
		m := newMock()
		got := controller.NewTrucoCuiController(m).Exec("a")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Respond", true)
	})

	t.Run("decline", func(t *testing.T) {
		m := newMock()
		got := controller.NewTrucoCuiController(m).Exec("d")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Respond", false)
	})

	t.Run("next", func(t *testing.T) {
		m := newMock()
		got := controller.NewTrucoCuiController(m).Exec("n")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Next")
	})

	t.Run("hint", func(t *testing.T) {
		m := newMock()
		got := controller.NewTrucoCuiController(m).Exec("h")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Hint")
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		got := controller.NewTrucoCuiController(m).Exec("l")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		got := controller.NewTrucoCuiController(newMock()).Exec("xyz")
		assert.NotEqual(t, "bye.", got)
		assert.NotEmpty(t, got)
	})
}

// #5618: Web は SettingsPanel で目標点 (9/12/15/18/24/30) を選んで reset に
// 渡せるのに、CUI には目標点を触るコマンドが 1 つも無く、既定の 15 でしか
// 遊べなかった。
func TestTrucoCuiControllerSetsTheMatchTarget(t *testing.T) {
	// **既定と違う設定を持たせる。**既定のままだと「今の設定を読んで書き換える」が
	// 「既定から作り直す」実装でも通ってしまい、何も確かめない。難易度は現状
	// Normal しか無いので、目標点そのものを既定から動かして区別する。
	current := domain.DefaultTrucoConfig()
	current.MatchTarget = 30
	newMock := func() *mockUsecases.MockTrucoInteractor {
		m := new(mockUsecases.MockTrucoInteractor)
		m.On("GetConfig").Return(current)
		m.On("ResetWithConfig", mock.Anything).Return(`{"phase":0}`)
		return m
	}

	for _, alias := range []string{"sm", "setmatchtarget"} {
		t.Run(alias, func(t *testing.T) {
			m := newMock()
			c := controller.NewTrucoCuiController(m)

			assert.Equal(t, `{"phase":0}`, c.Exec(alias+" 24"))
			// **設定を書き換えてリセットする。**目標点だけ変えて他の設定は保つ。
			m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.TrucoConfig) bool {
				return cfg.MatchTarget == 24
			}))
			// **今の設定を土台にする。**既定から作り直すと、他の設定 (将来増える
			// ぶんも含めて) が黙って戻る。
			m.AssertCalled(t, "GetConfig")
		})
	}

	t.Run("rejects a value outside the domain range", func(t *testing.T) {
		m := newMock()
		c := controller.NewTrucoCuiController(m)

		// 上限は domain.TrucoMaxMatchTarget。Web も同じ範囲でクランプしている。
		out := c.Exec("sm " + strconv.Itoa(domain.TrucoMaxMatchTarget+1))
		assert.Contains(t, out, strconv.Itoa(domain.TrucoMaxMatchTarget))
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})

	// **下限側も踏む。**上限だけだと、`v < min` を落としたミューテーションが
	// どのテストにも捕まらない (レビュー #5979)。
	t.Run("rejects a value below the domain range", func(t *testing.T) {
		m := newMock()
		c := controller.NewTrucoCuiController(m)

		out := c.Exec("sm " + strconv.Itoa(domain.TrucoMinMatchTarget-1))
		assert.Contains(t, out, strconv.Itoa(domain.TrucoMinMatchTarget))
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("asks for the value when it is missing", func(t *testing.T) {
		m := newMock()
		c := controller.NewTrucoCuiController(m)

		assert.NotEmpty(t, c.Exec("sm"))
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})
}
