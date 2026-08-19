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

func TestLiteratureCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockLiteratureInteractor {
		m := new(mockUsecases.MockLiteratureInteractor)
		m.On("GetConfig").Return(domain.DefaultLiteratureConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Ask", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Claim", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit and reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewLiteratureCuiController(m)
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultLiteratureConfig())
	})

	// **要求は 相手・スート・ランク の 3 つ。**
	t.Run("asks for a card", func(t *testing.T) {
		m := newMock()
		c := controller.NewLiteratureCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("a 1 1 2"))
		m.AssertCalled(t, "Ask", 1, domain.CardDesignSpade, 2)
		assert.Equal(t, mockOutput, c.Exec("ask 5 4 13"))
		m.AssertCalled(t, "Ask", 5, domain.CardDesignDiamond, 13)
	})

	t.Run("ask rejects bad input", func(t *testing.T) {
		c := controller.NewLiteratureCuiController(newMock())
		assert.Contains(t, c.Exec("a"), msgStem("askNeedsSeatSuitRank"))
		assert.Contains(t, c.Exec("a 1 1"), msgStem("askNeedsSeatSuitRank"))
		assert.Contains(t, c.Exec("a 9 1 2"), msgStem("invalidSeat0Max"))
		assert.Contains(t, c.Exec("a abc 1 2"), msgStem("invalidSeat0Max"))
		assert.Contains(t, c.Exec("a 1 9 2"), msgStem("invalidSuit14Letters"))
		assert.Contains(t, c.Exec("a 1 abc 2"), msgStem("invalidSuit14Letters"))
		assert.Contains(t, c.Exec("a 1 1 14"), msgStem("invalidRank113"))
		assert.Contains(t, c.Exec("a 1 1 abc"), msgStem("invalidRank113"))
	})

	// **宣言は組と 6 席。**所在を 6 枚ぶん申告する。
	// **1 打で確定しない** (#5733)。1 度目は読み上げるだけで、同じ内容の
	// 打ち直しか y で初めて Claim が呼ばれる。
	t.Run("claims a half-suit after a confirmation", func(t *testing.T) {
		m := newMock()
		c := controller.NewLiteratureCuiController(m)

		preview := c.Exec("c 0 0 0 2 2 4 4")
		m.AssertNotCalled(t, "Claim", 0, []int{0, 0, 2, 2, 4, 4})
		// 読み上げに札と席の組が出る (組 0 は ♠ の 2-7)。
		assert.Contains(t, preview, "組0")
		assert.Contains(t, preview, "2→席0")
		assert.Contains(t, preview, "7→席4")

		assert.Equal(t, mockOutput, c.Exec("y"))
		m.AssertCalled(t, "Claim", 0, []int{0, 0, 2, 2, 4, 4})

		// 同じ内容をもう一度打つのも確認になる。
		assert.NotEqual(t, mockOutput, c.Exec("claim 7 1 1 1 3 3 5"))
		assert.Equal(t, mockOutput, c.Exec("claim 7 1 1 1 3 3 5"))
		m.AssertCalled(t, "Claim", 7, []int{1, 1, 1, 3, 3, 5})
		claims := 0
		for _, call := range m.Calls {
			if call.Method == "Claim" {
				claims++
			}
		}
		assert.Equal(t, 2, claims, "確定したのは 2 回だけ。読み上げの 2 回は数に入らない")
	})

	// **確認待ちは他のコマンドで消える** (#5733)。残したままだと、あとで打った
	// y が意図しない宣言を確定させる。
	t.Run("a claim waiting for confirmation is dropped by any other command", func(t *testing.T) {
		m := newMock()
		c := controller.NewLiteratureCuiController(m)

		c.Exec("c 0 0 0 2 2 4 4")
		assert.Equal(t, mockOutput, c.Exec("a 1 1 2"))
		assert.Contains(t, c.Exec("y"), msgStem("literature.nothingToConfirm"))
		m.AssertNotCalled(t, "Claim", 0, []int{0, 0, 2, 2, 4, 4})

		// 中身の違う宣言は上書きするだけで、古いほうは確定しない。
		c.Exec("c 0 0 0 2 2 4 4")
		c.Exec("c 1 5 5 5 5 5 5")
		assert.Equal(t, mockOutput, c.Exec("y"))
		m.AssertCalled(t, "Claim", 1, []int{5, 5, 5, 5, 5, 5})
		m.AssertNotCalled(t, "Claim", 0, []int{0, 0, 2, 2, 4, 4})
	})

	// **リセットも待ちを消す** (レビュー指摘 #6070)。r / reset は
	// execCuiCommand が gameHandler より先に拾うので、gameHandler 側の
	// クリアだけでは配り直した卓に古い宣言が確定してしまう。
	t.Run("reset drops a claim waiting for confirmation", func(t *testing.T) {
		m := newMock()
		c := controller.NewLiteratureCuiController(m)

		c.Exec("c 0 0 0 2 2 4 4")
		assert.Equal(t, mockOutput, c.Exec("r"))
		assert.Contains(t, c.Exec("y"), msgStem("literature.nothingToConfirm"))
		m.AssertNotCalled(t, "Claim", 0, []int{0, 0, 2, 2, 4, 4})
	})

	// 打ち間違いは待ちを引き継がない。
	t.Run("an invalid claim clears the pending one", func(t *testing.T) {
		m := newMock()
		c := controller.NewLiteratureCuiController(m)
		c.Exec("c 0 0 0 2 2 4 4")
		assert.Contains(t, c.Exec("c 0 0 0 2 2 4"), msgStem("claimNeedsHalfSuitAndHolders"))
		assert.Contains(t, c.Exec("y"), msgStem("literature.nothingToConfirm"))
		m.AssertNotCalled(t, "Claim", 0, []int{0, 0, 2, 2, 4, 4})
	})

	// 確認待ちが無いのに y を打っても何も起きない。
	t.Run("y without a pending claim confirms nothing", func(t *testing.T) {
		m := newMock()
		c := controller.NewLiteratureCuiController(m)
		assert.Contains(t, c.Exec("y"), msgStem("literature.nothingToConfirm"))
		assert.Contains(t, c.Exec("yes"), msgStem("literature.nothingToConfirm"))
		m.AssertNotCalled(t, "Claim", mock.Anything, mock.Anything)
	})

	t.Run("claim rejects bad input", func(t *testing.T) {
		c := controller.NewLiteratureCuiController(newMock())
		assert.Contains(t, c.Exec("c"), msgStem("claimNeedsHalfSuitAndHolders"))
		assert.Contains(t, c.Exec("c 0 0 0 2"), msgStem("claimNeedsHalfSuitAndHolders"))
		assert.Contains(t, c.Exec("c 9 0 0 2 2 4 4"), msgStem("invalidHalfSuit0Max"))
		assert.Contains(t, c.Exec("c abc 0 0 2 2 4 4"), msgStem("invalidHalfSuit0Max"))
		assert.Contains(t, c.Exec("c 0 0 0 2 2 4 9"), msgStem("invalidSeat0Max"))
		assert.Contains(t, c.Exec("c 0 0 0 2 2 4 abc"), msgStem("invalidSeat0Max"))
	})

	t.Run("log and unknown", func(t *testing.T) {
		m := newMock()
		c := controller.NewLiteratureCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})
}
