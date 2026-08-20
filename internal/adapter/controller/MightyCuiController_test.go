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

// newMightyCuiMock returns a MockMightyInteractor that accepts every
// interactor call and returns mockOutput so subtests can focus on
// dispatch / validation only.
func newMightyCuiMock(mockOutput string) *mockUsecases.MockMightyInteractor {
	m := new(mockUsecases.MockMightyInteractor)
	m.On("GetConfig").Return(domain.DefaultMightyConfig())
	m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	m.On("Bid", mock.Anything, mock.Anything).Return(mockOutput)
	m.On("DeclareTrumpAndFriend", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
	m.On("ExchangeKitty", mock.Anything).Return(mockOutput)
	m.On("Play", mock.Anything).Return(mockOutput)
	m.On("PlayJokerLead", mock.Anything, mock.Anything).Return(mockOutput)
	m.On("NextTrick").Return(mockOutput)
	m.On("NextRound").Return(mockOutput)
	m.On("Hint").Return(mockOutput)
	m.On("ActionLog").Return(mockOutput)
	return m
}

func TestMightyCuiController_Exec(t *testing.T) {
	const mockOutput = `{"phase":0}`

	t.Run("quit q", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit quit", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultMightyConfig())
	})

	// bid: covers no-trump axis (issue #1677 extension)
	t.Run("bid plain", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("b 14"))
		m.AssertCalled(t, "Bid", 14, false)
	})

	t.Run("bid pass", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("bid 0"))
		m.AssertCalled(t, "Bid", 0, false)
	})

	t.Run("bid nt trailing flag", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("b 15 nt"))
		m.AssertCalled(t, "Bid", 15, true)
	})

	t.Run("bid notrump alias", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("b 16 notrump"))
		m.AssertCalled(t, "Bid", 16, true)
	})

	t.Run("bid no args errors", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("b"), msgStem("bidValueRequiredPass1320"))
	})

	t.Run("bid invalid", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("b abc"), msgStem("invalidBidValue"))
	})

	t.Run("bid over max", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("b 21"), msgKey("invalidBidValue", "val", "21"))
	})

	// trump+friend
	t.Run("trump with suits and partner value", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("t 1 2 13"))
		m.AssertCalled(t, "DeclareTrumpAndFriend", 1, 2, 13)
	})

	t.Run("trump no-trump suit=-1", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("trump -1 1 1"))
		m.AssertCalled(t, "DeclareTrumpAndFriend", -1, 1, 1)
	})

	t.Run("trump partner suit 0 (joker)", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("t 3 0 1"))
		m.AssertCalled(t, "DeclareTrumpAndFriend", 3, 0, 1)
	})

	t.Run("trump no args", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("t"), msgUsage("usageTrumpSuitPartnersuitPartnervalSuit1NoTrump1Spad"))
	})

	t.Run("trump partial args", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("t 1 2"), msgUsage("usageTrumpSuitPartnersuitPartnervalSuit1NoTrump1Spad"))
	})

	t.Run("trump invalid suit value", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("t 5 2 13"), msgKey("invalidSuit", "val", "5"))
	})

	t.Run("trump invalid partner suit", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("t 1 5 13"), msgKey("invalidPartnerSuit", "val", "5"))
	})

	t.Run("trump invalid partner value", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("t 1 2 0"), msgKey("invalidPartnerValue", "val", "0"))
	})

	// exchange: three indices
	t.Run("exchange three indices", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("e 0 1 2"))
		m.AssertCalled(t, "ExchangeKitty", []int{0, 1, 2})
	})

	t.Run("exchange long form", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("exchange 3 4 5"))
		m.AssertCalled(t, "ExchangeKitty", []int{3, 4, 5})
	})

	t.Run("exchange no args", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("e"), msgUsage("usageExchangeIJKThreeCardIndicesToDiscardFromKittyPi"))
	})

	t.Run("exchange only two args (insufficient)", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("e 1 2"), msgUsage("usageExchangeIJKThreeCardIndicesToDiscardFromKittyPi"))
	})

	t.Run("exchange invalid index", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("e abc 1 2"), msgInvalidCardIndexPrefix())
	})

	// play
	t.Run("play p", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("p"), msgCardIndexRequired())
	})

	t.Run("play invalid", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("play foo"), msgInvalidCardIndexPrefix())
	})

	// jokerlead (Mighty-specific)
	t.Run("jokerlead jl with args", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("jl 4 2"))
		m.AssertCalled(t, "PlayJokerLead", 4, 2)
	})

	t.Run("jokerlead long form", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("jokerlead 0 1"))
		m.AssertCalled(t, "PlayJokerLead", 0, 1)
	})

	t.Run("jokerlead no args", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("jl"), msgUsage("usageJokerleadCardindexDemandsuitDemandsuit1Spade2Cl"))
	})

	t.Run("jokerlead missing demand suit", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("jl 1"), msgUsage("usageJokerleadCardindexDemandsuitDemandsuit1Spade2Cl"))
	})

	t.Run("jokerlead invalid demand suit", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("jl 1 5"), msgKey("invalidDemandSuit", "val", "5"))
	})

	// next / nextround
	t.Run("next n", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextTrick")
	})

	t.Run("nextround nr", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	// config setters: setdifficulty / setlimit / setminbid / setnotrumpextra
	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultMightyConfig()
		expected.CpuDifficulty = domain.MightyCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty over range", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Equal(t, msgInvalidCpuDifficulty("3"), c.Exec("sd 3"))
	})

	t.Run("setdifficulty negative", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Equal(t, msgInvalidCpuDifficulty("-1"), c.Exec("sd -1"))
	})

	t.Run("setlimit valid", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sl 200"))
		expected := domain.DefaultMightyConfig()
		expected.PointLimit = 200
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit zero", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("sl 0"), msgInvalidPointLimitPrefix())
	})

	t.Run("setminbid valid", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sm 15"))
		expected := domain.DefaultMightyConfig()
		expected.MinBid = 15
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setminbid over max", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("sm 21"), msgKey("invalidMinBid", "val", "21"))
	})

	t.Run("setnotrumpextra valid", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sn 3"))
		expected := domain.DefaultMightyConfig()
		expected.NoTrumpExtra = 3
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setnotrumpextra no args", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("sn"), msgStem("noTrumpExtraRequired"))
	})

	t.Run("setnotrumpextra over max", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("sn 20"), msgKey("invalidNoTrumpExtra", "val", "20"))
	})

	// log / hint
	t.Run("log", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("hint h", func(t *testing.T) {
		m := newMightyCuiMock(mockOutput)
		c := controller.NewMightyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		m.AssertCalled(t, "Hint")
	})

	// unknown / empty
	t.Run("unknown", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})

	t.Run("empty", func(t *testing.T) {
		c := controller.NewMightyCuiController(newMightyCuiMock(mockOutput))
		assert.Contains(t, c.Exec(""), "'help'")
	})
}
