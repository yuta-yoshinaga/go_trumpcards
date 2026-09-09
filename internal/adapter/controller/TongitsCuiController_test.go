//go:build test
// +build test

package controller_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestTongitsCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockTongitsInteractor {
		m := new(mockUsecases.MockTongitsInteractor)
		m.On("GetConfig").Return(domain.DefaultTongitsConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("DrawFromStock").Return(mockOutput)
		m.On("DrawFromDiscard").Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Meld", mock.Anything).Return(mockOutput)
		m.On("Sapaw", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Challenge", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		c := controller.NewTongitsCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewTongitsCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		assert.Equal(t, mockOutput, c.Exec("reset"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultTongitsConfig())
	})

	t.Run("drawstock", func(t *testing.T) {
		m := newMock()
		c := controller.NewTongitsCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("ds"))
		assert.Equal(t, mockOutput, c.Exec("drawstock"))
		m.AssertCalled(t, "DrawFromStock")
	})

	t.Run("drawdiscard", func(t *testing.T) {
		m := newMock()
		c := controller.NewTongitsCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("dd"))
		assert.Equal(t, mockOutput, c.Exec("drawdiscard"))
		m.AssertCalled(t, "DrawFromDiscard")
	})

	t.Run("discard with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewTongitsCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 3"))
		assert.Equal(t, mockOutput, c.Exec("discard 5"))
		m.AssertCalled(t, "Discard", 3)
		m.AssertCalled(t, "Discard", 5)
	})

	t.Run("discard no args", func(t *testing.T) {
		c := controller.NewTongitsCuiController(newMock())
		assert.Contains(t, c.Exec("d"), msgCardIndexRequired())
		assert.Contains(t, c.Exec("d abc"), msgInvalidCardIndexPrefix())
	})

	t.Run("meld with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewTongitsCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("m 2 3 4"))
		assert.Equal(t, mockOutput, c.Exec("meld 5 6 7"))
		m.AssertCalled(t, "Meld", []int{2, 3, 4})
		m.AssertCalled(t, "Meld", []int{5, 6, 7})
	})

	t.Run("sapaw and challenge", func(t *testing.T) {
		m := newMock()
		c := controller.NewTongitsCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sp 1 2 3"))
		assert.Equal(t, mockOutput, c.Exec("sapaw 2 0 4"))
		assert.Equal(t, mockOutput, c.Exec("c"))
		assert.Equal(t, mockOutput, c.Exec("challenge"))
		m.AssertCalled(t, "Sapaw", 1, 2, 3)
		m.AssertCalled(t, "Sapaw", 2, 0, 4)
		m.AssertCalled(t, "Challenge", []bool{true, true})
	})

	t.Run("meld and sapaw invalid args", func(t *testing.T) {
		c := controller.NewTongitsCuiController(newMock())
		// **キー名そのものを期待値にしない。** i18n.T は未知のキーをそのまま返すので、
		// "tongits.cardIndexRequired" と比べる試験は翻訳が無いときだけ通り、
		// 翻訳を足した瞬間に落ちる (実際そうなった)。解決後の文言を見て、
		// ついでに生のキーが漏れていないことも確かめる。
		noIdx := c.Exec("m")
		assert.Equal(t, i18n.T("tongits.cardIndexRequired"), noIdx)
		assert.NotContains(t, noIdx, "tongits.", "生の i18n キーが利用者に見えている")

		badIdx := c.Exec("m abc")
		assert.Equal(t, i18n.Tf("tongits.invalidCardIndex", "val", "abc"), badIdx)
		assert.Contains(t, badIdx, "abc", "どの引数が悪いのかが出ていない")

		shortSapaw := c.Exec("sp 1 2")
		assert.Equal(t, i18n.T("tongits.sapawArgsRequired"), shortSapaw)
		assert.NotContains(t, shortSapaw, "tongits.", "生の i18n キーが利用者に見えている")
	})

	t.Run("nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewTongitsCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		assert.Equal(t, mockOutput, c.Exec("nextround"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewTongitsCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultTongitsConfig()
		expected.CpuDifficulty = domain.TongitsCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty errors", func(t *testing.T) {
		c := controller.NewTongitsCuiController(newMock())
		assert.Contains(t, c.Exec("sd"), msgCpuDifficultyRequired())
		assert.Contains(t, c.Exec("sd abc"), msgInvalidCpuDifficultyPrefix())
		assert.Equal(t, msgInvalidCpuDifficulty("-1"), c.Exec("sd -1"))
		assert.Equal(t, msgInvalidCpuDifficulty("3"), c.Exec("sd 3"))
	})

	t.Run("setlimit valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewTongitsCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sl 100"))
		expected := domain.DefaultTongitsConfig()
		expected.PointLimit = 100
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit errors", func(t *testing.T) {
		c := controller.NewTongitsCuiController(newMock())
		assert.Contains(t, c.Exec("sl"), msgPointLimitRequired())
		assert.Contains(t, c.Exec("sl abc"), msgInvalidPointLimitPrefix())
		assert.Equal(t, msgInvalidPointLimit("0"), c.Exec("sl 0"))
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewTongitsCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		assert.Equal(t, mockOutput, c.Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewTongitsCuiController(newMock())
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewTongitsCuiController(newMock())
		assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
	})
}
