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

func TestKaiserCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockKaiserInteractor {
		m := new(mockUsecases.MockKaiserInteractor)
		m.On("GetConfig").Return(domain.DefaultKaiserConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("PassBid").Return(mockOutput)
		m.On("SetTrump", mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("PlayCard", mock.Anything).Return(mockOutput)
		m.On("NextHand").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		return m
	}

	// **CUI にもヒントが要る。**Web にはあるのに hint/h が未登録で、
	// 「そんなコマンドは無い」と弾かれていた (#4938)。
	t.Run("hint", func(t *testing.T) {
		m := newMock()
		c := controller.NewKaiserCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("hint"))
		m.AssertNumberOfCalls(t, "Hint", 2)
		// 既存コマンドは巻き添えを食わない。
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("quit and reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewKaiserCuiController(m)
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultKaiserConfig())
	})

	// **契約は省略できる。**省略時は切札あり。
	t.Run("bid with and without a contract", func(t *testing.T) {
		m := newMock()
		c := controller.NewKaiserCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("b 8"))
		m.AssertCalled(t, "Bid", 8, domain.KaiserContractTrump)
		assert.Equal(t, mockOutput, c.Exec("bid 9 2"))
		m.AssertCalled(t, "Bid", 9, domain.KaiserContractLowNoTrump)
		assert.Equal(t, mockOutput, c.Exec("ps"))
		m.AssertCalled(t, "PassBid")
	})

	// **最低は 7。**6 以下はドメインへ行く前に弾く。
	t.Run("bid rejects a value outside 7-12", func(t *testing.T) {
		c := controller.NewKaiserCuiController(newMock())
		assert.True(t, msgRejected(c.Exec("b")))
		assert.Contains(t, c.Exec("b abc"), "Invalid bid")
		assert.Contains(t, c.Exec("b 6"), "Invalid bid")
		assert.Contains(t, c.Exec("b 13"), "Invalid bid")
		assert.Contains(t, c.Exec("b 8 9"), msgStem("invalidContract02"))
	})

	t.Run("trump", func(t *testing.T) {
		m := newMock()
		c := controller.NewKaiserCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("t 3"))
		m.AssertCalled(t, "SetTrump", domain.CardDesignHeart)
		assert.Contains(t, c.Exec("t"), msgStem("suitRequiredLetters"))
		assert.Contains(t, c.Exec("t 5"), msgStem("invalidSuitRange"))
	})

	// **捨て札は必ず 2 枚。**キティと同数でなければ手札が合わなくなる。
	t.Run("discard takes exactly two indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewKaiserCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 0 3"))
		m.AssertCalled(t, "Discard", []int{0, 3})
		assert.Contains(t, c.Exec("d 0"), msgStem("twoIndicesRequired"))
		assert.Contains(t, c.Exec("d"), msgStem("twoIndicesRequired"))
		assert.Contains(t, c.Exec("d a b"), msgInvalidCardIndexPrefix())
	})

	t.Run("play and next", func(t *testing.T) {
		m := newMock()
		c := controller.NewKaiserCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 2"))
		m.AssertCalled(t, "PlayCard", 2)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextHand")
		assert.Contains(t, c.Exec("p abc"), msgInvalidCardIndexPrefix())
	})

	t.Run("log and unknown", func(t *testing.T) {
		m := newMock()
		c := controller.NewKaiserCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})
}
