package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestIndianPokerCuiController_Quit(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestIndianPokerCuiController_Reset(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	mi.On("Reset").Return("reset ok")
	assert.Equal(t, "reset ok", c.Exec("r"))
	assert.Equal(t, "reset ok", c.Exec("reset"))
}

func TestIndianPokerCuiController_Fold(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	mi.On("Action", domain.IndianPokerActionFold, 0, 0).Return("fold ok")
	assert.Equal(t, "fold ok", c.Exec("f"))
	assert.Equal(t, "fold ok", c.Exec("fold"))
}

func TestIndianPokerCuiController_Check(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	mi.On("Action", domain.IndianPokerActionCheck, 0, 0).Return("check ok")
	assert.Equal(t, "check ok", c.Exec("ck"))
	assert.Equal(t, "check ok", c.Exec("check"))
}

func TestIndianPokerCuiController_Call(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	mi.On("Action", domain.IndianPokerActionCall, 0, 0).Return("call ok")
	assert.Equal(t, "call ok", c.Exec("c"))
	assert.Equal(t, "call ok", c.Exec("call"))
}

func TestIndianPokerCuiController_Bet(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	mi.On("Action", domain.IndianPokerActionBet, 100, 0).Return("bet ok")
	assert.Equal(t, "bet ok", c.Exec("b 100"))
	assert.Equal(t, "bet ok", c.Exec("bet 100"))
}

func TestIndianPokerCuiController_Bet_NoAmount(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	result := c.Exec("b")
	assert.Equal(t, i18n.T("indianpoker.amountRequired"), result)
}

func TestIndianPokerCuiController_Bet_InvalidAmount(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	assert.Equal(t, invalidArg("indianpoker.invalidAmount", "val", "abc"), c.Exec("b abc"))
	assert.Equal(t, invalidArg("indianpoker.invalidAmount", "val", "xyz"), c.Exec("bet xyz"))
}

func TestIndianPokerCuiController_Bet_NonPositiveAmount(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	assert.Equal(t, invalidArg("indianpoker.invalidAmount", "val", "-50"), c.Exec("b -50"))
	assert.Equal(t, invalidArg("indianpoker.invalidAmount", "val", "0"), c.Exec("bet 0"))
}

func TestIndianPokerCuiController_Raise(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	mi.On("Action", domain.IndianPokerActionRaise, 200, 0).Return("raise ok")
	assert.Equal(t, "raise ok", c.Exec("ra 200"))
	assert.Equal(t, "raise ok", c.Exec("raise 200"))
}

func TestIndianPokerCuiController_Raise_NoAmount(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	result := c.Exec("ra")
	assert.Equal(t, i18n.T("indianpoker.amountRequired"), result)
}

func TestIndianPokerCuiController_Raise_InvalidAmount(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	assert.Equal(t, invalidArg("indianpoker.invalidAmount", "val", "abc"), c.Exec("ra abc"))
	assert.Equal(t, invalidArg("indianpoker.invalidAmount", "val", "xyz"), c.Exec("raise xyz"))
}

func TestIndianPokerCuiController_Raise_NonPositiveAmount(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	assert.Equal(t, invalidArg("indianpoker.invalidAmount", "val", "-30"), c.Exec("ra -30"))
	assert.Equal(t, invalidArg("indianpoker.invalidAmount", "val", "0"), c.Exec("raise 0"))
}

func TestIndianPokerCuiController_AllIn(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	mi.On("Action", domain.IndianPokerActionAllIn, 0, 0).Return("allin ok")
	assert.Equal(t, "allin ok", c.Exec("a"))
	assert.Equal(t, "allin ok", c.Exec("allin"))
}

func TestIndianPokerCuiController_BettingLimit_Valid(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultIndianPokerConfig())
	cfg := domain.DefaultIndianPokerConfig()
	cfg.BettingLimit = domain.BettingLimitNoLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bl 2"))
}

func TestIndianPokerCuiController_BettingLimit_LongCommand(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultIndianPokerConfig())
	cfg := domain.DefaultIndianPokerConfig()
	cfg.BettingLimit = domain.BettingLimitPotLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bettinglimit 1"))
}

func TestIndianPokerCuiController_BettingLimit_NoArgs(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	result := c.Exec("bl")
	assert.Equal(t, i18n.T("indianpoker.bettingLimitRequired"), result)
}

func TestIndianPokerCuiController_BettingLimit_InvalidValue(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	assert.Equal(t, invalidArg("indianpoker.invalidBettingLimit", "val", "abc"), c.Exec("bl abc"))
}

func TestIndianPokerCuiController_MetaAI_On(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	cfg := domain.DefaultIndianPokerConfig()
	mi.On("GetConfig").Return(cfg)
	expected := domain.DefaultIndianPokerConfig()
	expected.CpuMetaAI = true
	mi.On("ResetWithConfig", expected, mock.Anything).Return("mai ok")
	assert.Equal(t, "mai ok", c.Exec("mai 1"))
}

func TestIndianPokerCuiController_MetaAI_Off(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	cfg := domain.DefaultIndianPokerConfig()
	mi.On("GetConfig").Return(cfg)
	expected := domain.DefaultIndianPokerConfig()
	expected.CpuMetaAI = false
	mi.On("ResetWithConfig", expected, mock.Anything).Return("mai off")
	assert.Equal(t, "mai off", c.Exec("metaai 0"))
}

func TestIndianPokerCuiController_MetaAI_MissingArg(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	assert.Equal(t, i18n.T("metaAIRequired"), c.Exec("mai"))
}

func TestIndianPokerCuiController_MetaAI_InvalidArg(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	assert.Equal(t, invalidArg("invalidMetaAI", "val", "abc"), c.Exec("mai abc"))
}

func TestIndianPokerCuiController_MetaAI_OutOfRange(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	assert.Equal(t, invalidArg("invalidMetaAI", "val", "5"), c.Exec("mai 5"))
	assert.Equal(t, invalidArg("invalidMetaAI", "val", "-1"), c.Exec("mai -1"))
}

func TestIndianPokerCuiController_Ante_Valid(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultIndianPokerConfig())
	cfg := domain.DefaultIndianPokerConfig()
	cfg.Ante = 50
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("ante ok")
	assert.Equal(t, "ante ok", c.Exec("an 50"))
	assert.Equal(t, "ante ok", c.Exec("ante 50"))
}

func TestIndianPokerCuiController_Ante_NoArgs(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	result := c.Exec("an")
	assert.Equal(t, i18n.T("indianpoker.anteRequired"), result)
}

func TestIndianPokerCuiController_Ante_InvalidValue(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	assert.Equal(t, invalidArg("indianpoker.invalidAnte", "val", "abc"), c.Exec("an abc"))
}

func TestIndianPokerCuiController_Ante_NonPositive(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	assert.Equal(t, invalidArg("indianpoker.invalidAnte", "val", "0"), c.Exec("an 0"))
	assert.Equal(t, invalidArg("indianpoker.invalidAnte", "val", "-1"), c.Exec("an -1"))
}

func TestIndianPokerCuiController_ActionLog(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	mi.On("ActionLog").Return("log output")
	assert.Equal(t, "log output", c.Exec("log"))
}

func TestIndianPokerCuiController_Empty(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

func TestIndianPokerCuiController_Unknown(t *testing.T) {
	mi := new(usecase.MockIndianPokerInteractor)
	c := NewIndianPokerCuiController(mi)
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}
