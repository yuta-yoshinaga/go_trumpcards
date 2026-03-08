package usecase_test

import (
	"errors"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewBlackJackInteractor_NilGuards(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	t.Run("panics when bj is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BlackJackInteractor: bj must not be nil", func() {
			usecase.NewBlackJackInteractor(nil, bjpMock)
		})
	})
	t.Run("panics when bjp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BlackJackInteractor: bjp must not be nil", func() {
			usecase.NewBlackJackInteractor(domain.NewDefaultBlackJack(), nil)
		})
	})
}

func TestBlackJackInteractor_Method(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n")
	tbj := usecase.NewBlackJackInteractor(domain.NewDefaultBlackJack(), bjpMock)
	t.Run("success Reset", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.Reset())
	})
	t.Run("success Hit", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.Hit())
	})
	t.Run("success Stand", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.Stand())
	})
	t.Run("success Bet", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.Bet(100, 0, 0, 0))
	})
	t.Run("success DoubleDown", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.DoubleDown())
	})
	t.Run("success Split", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.Split())
	})
	t.Run("success Insurance", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.Insurance())
	})
	t.Run("success DeclineInsurance", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.DeclineInsurance())
	})
}

func TestBlackJackInteractor_NewMethods(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("ok")
	tbj := usecase.NewBlackJackInteractor(domain.NewDefaultBlackJack(), bjpMock)
	t.Run("Surrender delegates to presenter", func(t *testing.T) {
		assert.Equal(t, "ok", tbj.Surrender())
	})
	t.Run("SetDeckCount delegates to presenter", func(t *testing.T) {
		assert.Equal(t, "ok", tbj.SetDeckCount(2))
	})
	t.Run("ToggleHint delegates to presenter", func(t *testing.T) {
		assert.Equal(t, "ok", tbj.ToggleHint())
	})
}

func TestToggleSoft17(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("toggled")

	bjMock := new(interfaces.MockBlackJackGame)
	// GetConfig returns DealerHitsSoft17=false, so toggle sets it to true
	bjMock.On("GetConfig").Return(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false})
	bjMock.On("SetConfig", domain.BlackJackConfig{DealerHitsSoft17: true, CpuPlayerCount: 0, CountingEnabled: false}).Return(nil)

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.ToggleSoft17()
	assert.Equal(t, "toggled", result)
	bjMock.AssertCalled(t, "SetConfig", domain.BlackJackConfig{DealerHitsSoft17: true, CpuPlayerCount: 0, CountingEnabled: false})
}

func TestToggleSoft17_Error(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("error output")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("GetConfig").Return(domain.BlackJackConfig{DealerHitsSoft17: true, CpuPlayerCount: 0, CountingEnabled: false})
	bjMock.On("SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false}).Return(errors.New("config error"))

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.ToggleSoft17()
	assert.Equal(t, "error output", result)
}

func TestToggleCounting(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("counting toggled")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("GetConfig").Return(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false})
	bjMock.On("SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: true}).Return(nil)

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.ToggleCounting()
	assert.Equal(t, "counting toggled", result)
	bjMock.AssertCalled(t, "SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: true})
}

func TestToggleCounting_Error(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("error output")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("GetConfig").Return(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: true})
	bjMock.On("SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false}).Return(errors.New("config error"))

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.ToggleCounting()
	assert.Equal(t, "error output", result)
}

func TestResetWithConfig(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("reset done")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("SetConfig", domain.BlackJackConfig{DealerHitsSoft17: true, CpuPlayerCount: 2, CountingEnabled: true, DoubleAfterSplit: true, CountingSystem: domain.BJCountingKO}).Return(nil)
	bjMock.On("Reset").Return()

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.ResetWithConfig(true, 2, true, true, domain.BJCountingKO, 0, 0)
	assert.Equal(t, "reset done", result)
	bjMock.AssertCalled(t, "SetConfig", domain.BlackJackConfig{DealerHitsSoft17: true, CpuPlayerCount: 2, CountingEnabled: true, DoubleAfterSplit: true, CountingSystem: domain.BJCountingKO})
	bjMock.AssertNumberOfCalls(t, "Reset", 2)
}

func TestResetWithConfig_Error(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("config error output")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("SetConfig", domain.BlackJackConfig{DealerHitsSoft17: true, CpuPlayerCount: 5, CountingEnabled: false, DoubleAfterSplit: false}).Return(errors.New("invalid cpu count"))
	bjMock.On("Reset").Return()

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.ResetWithConfig(true, 5, false, false, 0, 0, 0)
	assert.Equal(t, "config error output", result)
	bjMock.AssertNumberOfCalls(t, "Reset", 1)
}

func TestSetCountingSystem(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("system changed")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("GetConfig").Return(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: true, DoubleAfterSplit: true})
	bjMock.On("SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: true, DoubleAfterSplit: true, CountingSystem: domain.BJCountingKO}).Return(nil)

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.SetCountingSystem(domain.BJCountingKO)
	assert.Equal(t, "system changed", result)
	bjMock.AssertCalled(t, "SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: true, DoubleAfterSplit: true, CountingSystem: domain.BJCountingKO})
}

func TestSetCountingSystem_Error(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("error output")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("GetConfig").Return(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: true, DoubleAfterSplit: true})
	bjMock.On("SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: true, DoubleAfterSplit: true, CountingSystem: 99}).Return(errors.New("invalid system"))

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.SetCountingSystem(99)
	assert.Equal(t, "error output", result)
}

func TestToggleDAS(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("das toggled")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("GetConfig").Return(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false, DoubleAfterSplit: true})
	bjMock.On("SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false, DoubleAfterSplit: false}).Return(nil)

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.ToggleDAS()
	assert.Equal(t, "das toggled", result)
	bjMock.AssertCalled(t, "SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false, DoubleAfterSplit: false})
}

func TestToggleDAS_Error(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("error output")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("GetConfig").Return(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false, DoubleAfterSplit: false})
	bjMock.On("SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false, DoubleAfterSplit: true}).Return(errors.New("config error"))

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.ToggleDAS()
	assert.Equal(t, "error output", result)
}

func TestSetDeckPenetration(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("penetration changed")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("GetConfig").Return(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false, DoubleAfterSplit: true})
	bjMock.On("SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false, DoubleAfterSplit: true, DeckPenetration: 50}).Return(nil)

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.SetDeckPenetration(50)
	assert.Equal(t, "penetration changed", result)
	bjMock.AssertCalled(t, "SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false, DoubleAfterSplit: true, DeckPenetration: 50})
}

func TestSetDeckPenetration_Error(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("error output")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("GetConfig").Return(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false, DoubleAfterSplit: true})
	bjMock.On("SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false, DoubleAfterSplit: true, DeckPenetration: 60}).Return(errors.New("invalid penetration"))

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.SetDeckPenetration(60)
	assert.Equal(t, "error output", result)
}

func TestSetCpuPlayerCount(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("cpu count changed")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("GetConfig").Return(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false, DoubleAfterSplit: true})
	bjMock.On("SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 2, CountingEnabled: false, DoubleAfterSplit: true}).Return(nil)

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.SetCpuPlayerCount(2)
	assert.Equal(t, "cpu count changed", result)
	bjMock.AssertCalled(t, "SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 2, CountingEnabled: false, DoubleAfterSplit: true})
}

func TestSetCpuPlayerCount_Error(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("error output")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("GetConfig").Return(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false, DoubleAfterSplit: true})
	bjMock.On("SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 5, CountingEnabled: false, DoubleAfterSplit: true}).Return(errors.New("invalid cpu count"))

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.SetCpuPlayerCount(5)
	assert.Equal(t, "error output", result)
}

func TestResetWithConfig_WithPenetration(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("reset done")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("SetConfig", domain.BlackJackConfig{DealerHitsSoft17: true, CpuPlayerCount: 1, CountingEnabled: true, DoubleAfterSplit: true, CountingSystem: domain.BJCountingHiLo, DeckPenetration: 50}).Return(nil)
	bjMock.On("Reset").Return()

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.ResetWithConfig(true, 1, true, true, domain.BJCountingHiLo, 50, 0)
	assert.Equal(t, "reset done", result)
	bjMock.AssertCalled(t, "SetConfig", domain.BlackJackConfig{DealerHitsSoft17: true, CpuPlayerCount: 1, CountingEnabled: true, DoubleAfterSplit: true, CountingSystem: domain.BJCountingHiLo, DeckPenetration: 50})
	bjMock.AssertNumberOfCalls(t, "Reset", 2)
}

func TestEarlySurrender(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("early surrender")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("PlayerEarlySurrender").Return(nil)

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.EarlySurrender()
	assert.Equal(t, "early surrender", result)
	bjMock.AssertCalled(t, "PlayerEarlySurrender")
}

func TestDeclineEarlySurrender(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("decline early surrender")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("PlayerDeclineEarlySurrender").Return(nil)

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.DeclineEarlySurrender()
	assert.Equal(t, "decline early surrender", result)
	bjMock.AssertCalled(t, "PlayerDeclineEarlySurrender")
}

func TestSetSurrenderRule(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("surrender rule changed")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("GetConfig").Return(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false, DoubleAfterSplit: true})
	bjMock.On("SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false, DoubleAfterSplit: true, SurrenderRule: domain.BJSurrenderEarly}).Return(nil)

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.SetSurrenderRule(domain.BJSurrenderEarly)
	assert.Equal(t, "surrender rule changed", result)
	bjMock.AssertCalled(t, "SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false, DoubleAfterSplit: true, SurrenderRule: domain.BJSurrenderEarly})
}

func TestSetSurrenderRule_Error(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("error output")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("GetConfig").Return(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false, DoubleAfterSplit: true})
	bjMock.On("SetConfig", domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: false, DoubleAfterSplit: true, SurrenderRule: 99}).Return(errors.New("invalid surrender rule"))

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.SetSurrenderRule(99)
	assert.Equal(t, "error output", result)
}

func TestResetWithConfig_WithSurrenderRule(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("reset done")

	bjMock := new(interfaces.MockBlackJackGame)
	bjMock.On("SetConfig", domain.BlackJackConfig{DealerHitsSoft17: true, CpuPlayerCount: 1, CountingEnabled: true, DoubleAfterSplit: true, CountingSystem: domain.BJCountingHiLo, DeckPenetration: 50, SurrenderRule: domain.BJSurrenderEarly}).Return(nil)
	bjMock.On("Reset").Return()

	tbj := usecase.NewBlackJackInteractor(bjMock, bjpMock)
	result := tbj.ResetWithConfig(true, 1, true, true, domain.BJCountingHiLo, 50, domain.BJSurrenderEarly)
	assert.Equal(t, "reset done", result)
	bjMock.AssertCalled(t, "SetConfig", domain.BlackJackConfig{DealerHitsSoft17: true, CpuPlayerCount: 1, CountingEnabled: true, DoubleAfterSplit: true, CountingSystem: domain.BJCountingHiLo, DeckPenetration: 50, SurrenderRule: domain.BJSurrenderEarly})
	bjMock.AssertNumberOfCalls(t, "Reset", 2)
}
