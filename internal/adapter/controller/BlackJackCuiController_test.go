package controller_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"

	"github.com/stretchr/testify/assert"
)

func TestBlackJackCuiController_Method(t *testing.T) {
	mockOutput := "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n"
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("Reset").Return(mockOutput)
	bjiMock.On("Hit").Return(mockOutput)
	bjiMock.On("Stand").Return(mockOutput)
	bjiMock.On("Bet", 100, 0, 0, 0).Return(mockOutput)
	bjiMock.On("DoubleDown").Return(mockOutput)
	bjiMock.On("Split").Return(mockOutput)
	bjiMock.On("Insurance").Return(mockOutput)
	bjiMock.On("DeclineInsurance").Return(mockOutput)
	tbc := controller.NewBlackJackCuiController(bjiMock)
	t.Run("success Exec q", func(t *testing.T) {
		assert.Equal(t, "bye.", tbc.Exec("q"))
	})
	t.Run("success Exec quit", func(t *testing.T) {
		assert.Equal(t, "bye.", tbc.Exec("quit"))
	})
	t.Run("success Exec r", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("r"))
	})
	t.Run("success Exec reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("reset"))
	})
	t.Run("success Exec h", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("h"))
	})
	t.Run("success Exec hit", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("hit"))
	})
	t.Run("success Exec s", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("s"))
	})
	t.Run("success Exec stand", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("stand"))
	})
	t.Run("success Exec b with amount", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("b 100"))
	})
	t.Run("success Exec bet with amount", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("bet 100"))
	})
	t.Run("error Exec b without amount", func(t *testing.T) {
		assert.Equal(t, msgBetAmountRequired(), tbc.Exec("b"))
	})
	t.Run("error Exec b with invalid amount", func(t *testing.T) {
		assert.Equal(t, msgInvalidBetAmount(""), tbc.Exec("b foo"))
	})
	t.Run("error Exec b with negative amount", func(t *testing.T) {
		assert.Equal(t, msgInvalidBetAmount(""), tbc.Exec("b -100"))
	})
	t.Run("error Exec b with zero amount", func(t *testing.T) {
		assert.Equal(t, msgInvalidBetAmount(""), tbc.Exec("b 0"))
	})
	t.Run("success Exec d", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("d"))
	})
	t.Run("success Exec doubledown", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("doubledown"))
	})
	t.Run("success Exec sp", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("sp"))
	})
	t.Run("success Exec split", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("split"))
	})
	t.Run("success Exec i", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("i"))
	})
	t.Run("success Exec insurance", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("insurance"))
	})
	t.Run("success Exec di", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("di"))
	})
	t.Run("success Exec declineinsurance", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("declineinsurance"))
	})
	t.Run("success Exec other", func(t *testing.T) {
		assert.Equal(t, i18n.ErrorPrefix+"コマンドが不明です: other", tbc.Exec("other"))
	})
	t.Run("success Exec empty", func(t *testing.T) {
		assert.Equal(t, "'help' でコマンド一覧を表示します。", tbc.Exec(""))
	})
}

func TestBlackJackCuiController_NewCommands(t *testing.T) {
	mockOutput := "----------\n"
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("Surrender").Return(mockOutput)
	bjiMock.On("ToggleHint").Return(mockOutput)
	bjiMock.On("SetDeckCount", 6).Return(mockOutput)
	bjiMock.On("SetDeckCount", -1).Return("error from domain")
	bjiMock.On("SetDeckCount", 0).Return("error from domain")
	tbc := controller.NewBlackJackCuiController(bjiMock)

	t.Run("sur", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("sur"))
	})
	t.Run("surrender", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("surrender"))
	})
	t.Run("hint", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("hint"))
	})
	t.Run("togglehint", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("togglehint"))
	})
	t.Run("sd with valid count", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("sd 6"))
	})
	t.Run("setdeckcount with valid count", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("setdeckcount 6"))
	})
	t.Run("sd without count", func(t *testing.T) {
		assert.Equal(t, "Deck count is required.", tbc.Exec("sd"))
	})
	t.Run("sd with invalid count", func(t *testing.T) {
		assert.Equal(t, "Invalid deck count. Please enter a number.", tbc.Exec("sd abc"))
	})
	t.Run("sd with negative count", func(t *testing.T) {
		assert.Equal(t, "Invalid deck count. Please enter a number.", tbc.Exec("sd -1"))
	})
	t.Run("sd with zero count", func(t *testing.T) {
		assert.Equal(t, "Invalid deck count. Please enter a number.", tbc.Exec("sd 0"))
	})
}

func TestBlackJackCuiController_SetCountingSystem(t *testing.T) {
	mockOutput := "----------\n"
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("SetCountingSystem", 2).Return(mockOutput)
	tbc := controller.NewBlackJackCuiController(bjiMock)

	t.Run("scs with valid system", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("scs 2"))
	})
	t.Run("setcountingsystem with valid system", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("setcountingsystem 2"))
	})
	t.Run("scs without arg", func(t *testing.T) {
		assert.Equal(t, "Counting system is required.", tbc.Exec("scs"))
	})
	t.Run("scs with invalid arg", func(t *testing.T) {
		assert.Equal(t, "Invalid counting system: abc. Please enter a number (0-3).", tbc.Exec("scs abc"))
	})
	t.Run("scs with negative arg", func(t *testing.T) {
		assert.Equal(t, "Invalid counting system: -1. Please enter a number (0-3).", tbc.Exec("scs -1"))
	})
	t.Run("scs with out-of-range arg", func(t *testing.T) {
		assert.Equal(t, "Invalid counting system: 4. Please enter a number (0-3).", tbc.Exec("scs 4"))
	})
}

func TestBlackJackCuiController_BetWithSideBets(t *testing.T) {
	mockOutput := "----------\n"
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("Bet", 100, 10, 20, 0).Return(mockOutput)
	bjiMock.On("Bet", 100, 10, 0, 0).Return(mockOutput)
	bjiMock.On("Bet", 100, 0, 0, 0).Return(mockOutput)
	tbc := controller.NewBlackJackCuiController(bjiMock)

	t.Run("bet with PP and T3", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("b 100 10 20"))
	})
	t.Run("bet with PP only", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("b 100 10"))
	})
	t.Run("bet with invalid PP (ignored)", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("b 100 abc"))
	})
	t.Run("bet with PP and invalid T3 (ignored)", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("b 100 10 xyz"))
	})
}

func TestBlackJackCuiController_BetWithHandCount(t *testing.T) {
	mockOutput := "----------\n"
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("Bet", 100, 0, 0, 2).Return(mockOutput)
	bjiMock.On("Bet", 100, 10, 20, 3).Return(mockOutput)
	bjiMock.On("Bet", 100, 0, 0, 0).Return(mockOutput)
	tbc := controller.NewBlackJackCuiController(bjiMock)

	t.Run("bet with handCount", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("b 100 0 0 2"))
	})
	t.Run("bet with side bets and handCount", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("b 100 10 20 3"))
	})
	t.Run("bet with invalid handCount (ignored)", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("b 100 0 0 abc"))
	})
}

func TestBlackJackCuiController_Soft17AndCountingCommands(t *testing.T) {
	mockOutput := "----------\n"
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("ToggleSoft17").Return(mockOutput)
	bjiMock.On("ToggleCounting").Return(mockOutput)
	bjiMock.On("ToggleDAS").Return(mockOutput)
	tbc := controller.NewBlackJackCuiController(bjiMock)

	t.Run("soft17", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("soft17"))
	})
	t.Run("togglesoft17", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("togglesoft17"))
	})
	t.Run("counting", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("counting"))
	})
	t.Run("togglecounting", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("togglecounting"))
	})
	t.Run("das", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("das"))
	})
	t.Run("toggledas", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("toggledas"))
	})
}

func TestBlackJackCuiController_SetCpuPlayerCount(t *testing.T) {
	mockOutput := "----------\n"
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("SetCpuPlayerCount", 2).Return(mockOutput)
	tbc := controller.NewBlackJackCuiController(bjiMock)

	t.Run("scc with valid count", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("scc 2"))
	})
	t.Run("setcpucount with valid count", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("setcpucount 2"))
	})
	t.Run("scc without arg", func(t *testing.T) {
		assert.Equal(t, "CPU player count is required.", tbc.Exec("scc"))
	})
	t.Run("scc with invalid arg", func(t *testing.T) {
		assert.Equal(t, "Invalid CPU player count: abc. Please enter a number (0-3).", tbc.Exec("scc abc"))
	})
	t.Run("scc with negative arg", func(t *testing.T) {
		assert.Equal(t, "Invalid CPU player count: -1. Please enter a number (0-3).", tbc.Exec("scc -1"))
	})
	t.Run("scc with out-of-range arg", func(t *testing.T) {
		assert.Equal(t, "Invalid CPU player count: 4. Please enter a number (0-3).", tbc.Exec("scc 4"))
	})
	t.Run("scc with zero (valid)", func(t *testing.T) {
		bjiMock.On("SetCpuPlayerCount", 0).Return(mockOutput)
		assert.Equal(t, mockOutput, tbc.Exec("scc 0"))
	})
}

func TestBlackJackCuiController_SetPenetration_Valid(t *testing.T) {
	mockOutput := "----------\n"
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("SetDeckPenetration", 50).Return(mockOutput)
	tbc := controller.NewBlackJackCuiController(bjiMock)
	assert.Equal(t, mockOutput, tbc.Exec("pen 50"))
}

func TestBlackJackCuiController_SetPenetration_75(t *testing.T) {
	mockOutput := "----------\n"
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("SetDeckPenetration", 75).Return(mockOutput)
	tbc := controller.NewBlackJackCuiController(bjiMock)
	assert.Equal(t, mockOutput, tbc.Exec("pen 75"))
}

func TestBlackJackCuiController_SetPenetration_MissingArg(t *testing.T) {
	bjiMock := new(usecase.MockBlackJackInteractor)
	tbc := controller.NewBlackJackCuiController(bjiMock)
	assert.Equal(t, "Penetration rate is required.", tbc.Exec("pen"))
}

func TestBlackJackCuiController_SetPenetration_NonNumeric(t *testing.T) {
	bjiMock := new(usecase.MockBlackJackInteractor)
	tbc := controller.NewBlackJackCuiController(bjiMock)
	assert.Equal(t, "Invalid penetration rate. Please enter a number.", tbc.Exec("pen abc"))
}

func TestBlackJackCuiController_SetPenetration_InvalidValue(t *testing.T) {
	mockOutput := "error from domain"
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("SetDeckPenetration", 60).Return(mockOutput)
	tbc := controller.NewBlackJackCuiController(bjiMock)
	assert.Equal(t, mockOutput, tbc.Exec("pen 60"))
}

func TestBlackJackCuiController_SetPenetration_LongForm(t *testing.T) {
	mockOutput := "----------\n"
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("SetDeckPenetration", 50).Return(mockOutput)
	tbc := controller.NewBlackJackCuiController(bjiMock)
	assert.Equal(t, mockOutput, tbc.Exec("setpenetration 50"))
}

func TestBlackJackCuiController_EarlySurrender(t *testing.T) {
	mockOutput := "----------\n"
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("EarlySurrender").Return(mockOutput)
	tbc := controller.NewBlackJackCuiController(bjiMock)

	t.Run("es", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("es"))
	})
	t.Run("earlysurrender", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("earlysurrender"))
	})
}

func TestBlackJackCuiController_DeclineEarlySurrender(t *testing.T) {
	mockOutput := "----------\n"
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("DeclineEarlySurrender").Return(mockOutput)
	tbc := controller.NewBlackJackCuiController(bjiMock)

	t.Run("des", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("des"))
	})
	t.Run("declineearlysurrender", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("declineearlysurrender"))
	})
}

func TestBlackJackCuiController_SetSurrenderRule(t *testing.T) {
	mockOutput := "----------\n"
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("SetSurrenderRule", 1).Return(mockOutput)
	tbc := controller.NewBlackJackCuiController(bjiMock)

	t.Run("ssr with valid rule", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("ssr 1"))
	})
	t.Run("setsurrenderrule with valid rule", func(t *testing.T) {
		assert.Equal(t, mockOutput, tbc.Exec("setsurrenderrule 1"))
	})
	t.Run("ssr without arg", func(t *testing.T) {
		assert.Equal(t, "Surrender rule is required.", tbc.Exec("ssr"))
	})
	t.Run("ssr with invalid arg", func(t *testing.T) {
		assert.Equal(t, "Invalid surrender rule: abc. Please enter a number (0-2).", tbc.Exec("ssr abc"))
	})
	t.Run("ssr with negative arg", func(t *testing.T) {
		assert.Equal(t, "Invalid surrender rule: -1. Please enter a number (0-2).", tbc.Exec("ssr -1"))
	})
	t.Run("ssr with out-of-range arg", func(t *testing.T) {
		assert.Equal(t, "Invalid surrender rule: 3. Please enter a number (0-2).", tbc.Exec("ssr 3"))
	})
}
