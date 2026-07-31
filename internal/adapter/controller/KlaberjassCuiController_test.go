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

func TestKlaberjassCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockKlaberjassInteractor {
		m := new(mockUsecases.MockKlaberjassInteractor)
		m.On("GetConfig").Return(domain.DefaultKlaberjassConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("AcceptTrump").Return(mockOutput)
		m.On("CallTrump", mock.Anything).Return(mockOutput)
		m.On("Pass").Return(mockOutput)
		m.On("Schmeiss").Return(mockOutput)
		m.On("AnswerSchmeiss", mock.Anything).Return(mockOutput)
		m.On("PlayCard", mock.Anything).Return(mockOutput)
		m.On("NextDeal").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit and reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewKlaberjassCuiController(m)
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultKlaberjassConfig())
	})

	t.Run("bidding", func(t *testing.T) {
		m := newMock()
		c := controller.NewKlaberjassCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("a"))
		assert.Equal(t, mockOutput, c.Exec("accept"))
		assert.Equal(t, mockOutput, c.Exec("ps"))
		assert.Equal(t, mockOutput, c.Exec("pass"))
		assert.Equal(t, mockOutput, c.Exec("c 3"))
		m.AssertCalled(t, "AcceptTrump")
		m.AssertCalled(t, "Pass")
		m.AssertCalled(t, "CallTrump", domain.CardDesignHeart)
	})

	// **スートは 1〜4。**0 や 5 を通すとドメインで弾かれるだけの往復になる。
	t.Run("call rejects a bad suit", func(t *testing.T) {
		c := controller.NewKlaberjassCuiController(newMock())
		assert.Contains(t, c.Exec("c"), "required")
		assert.Contains(t, c.Exec("c abc"), "Invalid suit")
		assert.Contains(t, c.Exec("c 0"), "Invalid suit")
		assert.Contains(t, c.Exec("c 5"), "Invalid suit")
	})

	t.Run("schmeiss and its answer", func(t *testing.T) {
		m := newMock()
		c := controller.NewKlaberjassCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sm"))
		assert.Equal(t, mockOutput, c.Exec("y"))
		assert.Equal(t, mockOutput, c.Exec("no"))
		m.AssertCalled(t, "Schmeiss")
		m.AssertCalled(t, "AnswerSchmeiss", true)
		m.AssertCalled(t, "AnswerSchmeiss", false)
	})

	t.Run("play and next", func(t *testing.T) {
		m := newMock()
		c := controller.NewKlaberjassCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 2"))
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("next"))
		m.AssertCalled(t, "PlayCard", 2)
		m.AssertCalled(t, "NextDeal")
	})

	t.Run("play rejects a bad index", func(t *testing.T) {
		c := controller.NewKlaberjassCuiController(newMock())
		assert.Contains(t, c.Exec("p"), "required")
		assert.Contains(t, c.Exec("p abc"), "Invalid card index")
		assert.Contains(t, c.Exec("p 9"), "Invalid card index")
	})

	t.Run("settarget", func(t *testing.T) {
		m := newMock()
		c := controller.NewKlaberjassCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("st 300"))
		expected := domain.DefaultKlaberjassConfig()
		expected.TargetScore = 300
		m.AssertCalled(t, "ResetWithConfig", expected)
		assert.Contains(t, c.Exec("st 5"), "Invalid target score")
	})

	t.Run("log and unknown", func(t *testing.T) {
		m := newMock()
		c := controller.NewKlaberjassCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})
}
