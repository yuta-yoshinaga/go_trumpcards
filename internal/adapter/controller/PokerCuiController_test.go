package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockUsecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// --- quit ---

func TestPokerCuiController_Quit(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

// --- reset ---

func TestPokerCuiController_Reset(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Reset").Return("reset ok")
	assert.Equal(t, "reset ok", c.Exec("r"))
	assert.Equal(t, "reset ok", c.Exec("reset"))
}

// --- exchange ---

func TestPokerCuiController_Exchange_WithValidIndices(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Exchange", []int{0, 2, 4}).Return("exchange ok")
	assert.Equal(t, "exchange ok", c.Exec("e 0 2 4"))
}

func TestPokerCuiController_Exchange_LongCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Exchange", []int{1, 3}).Return("exchange ok")
	assert.Equal(t, "exchange ok", c.Exec("exchange 1 3"))
}

func TestPokerCuiController_Exchange_NoIndices(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Exchange", []int{}).Return("exchange empty")
	assert.Equal(t, "exchange empty", c.Exec("e"))
}

func TestPokerCuiController_Exchange_InvalidIndex_NonNumeric(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// "abc" fails strconv.Atoi => err != nil => skipped; "2" is valid
	mi.On("Exchange", []int{2}).Return("exchange ok")
	result := c.Exec("e abc 2")
	assert.Contains(t, result, "'abc'")
	assert.Contains(t, result, "exchange ok")
}

func TestPokerCuiController_Exchange_InvalidIndex_OutOfRange_Negative(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// -1 fails 0 <= idx check => skipped; "3" is valid
	mi.On("Exchange", []int{3}).Return("exchange ok")
	result := c.Exec("e -1 3")
	assert.Contains(t, result, "'-1'")
	assert.Contains(t, result, "exchange ok")
}

func TestPokerCuiController_Exchange_InvalidIndex_OutOfRange_Above(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// 5 fails idx <= 4 check => skipped; "0" is valid
	mi.On("Exchange", []int{0}).Return("exchange ok")
	result := c.Exec("e 5 0")
	assert.Contains(t, result, "'5'")
	assert.Contains(t, result, "exchange ok")
}

func TestPokerCuiController_Exchange_AllInvalid(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// All indices invalid => empty slice
	mi.On("Exchange", []int{}).Return("exchange empty")
	result := c.Exec("e abc -1 5 99")
	assert.Contains(t, result, "'abc'")
	assert.Contains(t, result, "exchange empty")
}

// --- stand ---

func TestPokerCuiController_Stand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Stand").Return("stand ok")
	assert.Equal(t, "stand ok", c.Exec("s"))
	assert.Equal(t, "stand ok", c.Exec("stand"))
}

// --- bet ---

func TestPokerCuiController_Bet_WithValidAmount(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionBet, 50, 0).Return("bet ok")
	assert.Equal(t, "bet ok", c.Exec("b 50"))
	assert.Equal(t, "bet ok", c.Exec("bet 50"))
}

func TestPokerCuiController_Bet_NoAmount(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// No amount => amount stays 0
	mi.On("Action", domain.PokerActionBet, 0, 0).Return("bet zero")
	assert.Equal(t, "bet zero", c.Exec("b"))
}

func TestPokerCuiController_Bet_InvalidAmount_NonNumeric(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// "abc" fails strconv.Atoi => amount stays 0
	mi.On("Action", domain.PokerActionBet, 0, 0).Return("bet zero")
	assert.Equal(t, msgInvalidBetAmount("abc"), c.Exec("b abc"))
}

func TestPokerCuiController_Bet_InvalidAmount_Negative(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// -20 succeeds Atoi but fails a > 0 check => amount stays 0
	mi.On("Action", domain.PokerActionBet, 0, 0).Return("bet zero")
	assert.Equal(t, msgInvalidBetAmount("-20"), c.Exec("b -20"))
}

func TestPokerCuiController_Bet_InvalidAmount_Zero(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// 0 succeeds Atoi but fails a > 0 check => amount stays 0
	mi.On("Action", domain.PokerActionBet, 0, 0).Return("bet zero")
	assert.Equal(t, msgInvalidBetAmount("0"), c.Exec("b 0"))
}

// --- call ---

func TestPokerCuiController_Call(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionCall, 0, 0).Return("call ok")
	assert.Equal(t, "call ok", c.Exec("c"))
	assert.Equal(t, "call ok", c.Exec("call"))
}

// --- raise ---

func TestPokerCuiController_Raise_WithValidAmount(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionRaise, 30, 0).Return("raise ok")
	assert.Equal(t, "raise ok", c.Exec("ra 30"))
	assert.Equal(t, "raise ok", c.Exec("raise 30"))
}

func TestPokerCuiController_Raise_NoAmount(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionRaise, 0, 0).Return("raise zero")
	assert.Equal(t, "raise zero", c.Exec("ra"))
}

func TestPokerCuiController_Raise_InvalidAmount_NonNumeric(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionRaise, 0, 0).Return("raise zero")
	assert.Equal(t, msgInvalidBetAmount("abc"), c.Exec("ra abc"))
}

func TestPokerCuiController_Raise_InvalidAmount_Negative(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionRaise, 0, 0).Return("raise zero")
	assert.Equal(t, msgInvalidBetAmount("-30"), c.Exec("ra -30"))
}

func TestPokerCuiController_Raise_InvalidAmount_Zero(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionRaise, 0, 0).Return("raise zero")
	assert.Equal(t, msgInvalidBetAmount("0"), c.Exec("ra 0"))
}

// --- fold ---

func TestPokerCuiController_Fold(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionFold, 0, 0).Return("fold ok")
	assert.Equal(t, "fold ok", c.Exec("f"))
	assert.Equal(t, "fold ok", c.Exec("fold"))
}

// --- check ---

func TestPokerCuiController_Check(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionCheck, 0, 0).Return("check ok")
	assert.Equal(t, "check ok", c.Exec("ck"))
	assert.Equal(t, "check ok", c.Exec("check"))
}

// --- allin ---

func TestPokerCuiController_AllIn(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionAllIn, 0, 0).Return("allin ok")
	assert.Equal(t, "allin ok", c.Exec("a"))
	assert.Equal(t, "allin ok", c.Exec("allin"))
}

// --- empty / unknown ---

func TestPokerCuiController_Empty(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	assert.Equal(t, "'help' でコマンド一覧を表示します。", c.Exec(""))
}

func TestPokerCuiController_Unknown(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	assert.Equal(t, i18n.ErrorPrefix+"コマンドが不明です: xyz", c.Exec("xyz"))
}

// --- betting limit ---

func TestPokerCuiController_BettingLimit_Valid(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultPokerConfig())
	cfg := domain.DefaultPokerConfig()
	cfg.BettingLimit = domain.BettingLimitPotLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bl 1"))
}

func TestPokerCuiController_BettingLimit_LongCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultPokerConfig())
	cfg := domain.DefaultPokerConfig()
	cfg.BettingLimit = domain.BettingLimitNoLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bettinglimit 2"))
}

func TestPokerCuiController_BettingLimit_NoArgs(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	assert.Contains(t, c.Exec("bl"), "Betting limit type is required")
}

func TestPokerCuiController_BettingLimit_InvalidValue(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// non-numeric: controller catches
	assert.Contains(t, c.Exec("bl abc"), "Invalid betting limit: abc")
	// numeric out-of-range: controller catches via ParseIntArg bounds
	assert.Equal(t, "Invalid betting limit: 5. Please enter 0-2.", c.Exec("bl 5"))
	assert.Equal(t, "Invalid betting limit: -1. Please enter 0-2.", c.Exec("bl -1"))
}

// --- lowball ---

func TestPokerCuiController_Lowball(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	cfg := domain.DefaultPokerConfig()
	mi.On("GetConfig").Return(cfg)
	expectedCfg := domain.DefaultPokerConfig()
	expectedCfg.IsLowball = true
	mi.On("ResetWithConfig", expectedCfg, mock.Anything).Return("lw ok")
	assert.Equal(t, "lw ok", c.Exec("lw"))
	mi.AssertCalled(t, "GetConfig")
	mi.AssertCalled(t, "ResetWithConfig", expectedCfg, mock.Anything)
}

// --- set cpu count ---

func TestPokerCuiController_SetCpuCount_Valid(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultPokerConfig())
	cfg := domain.DefaultPokerConfig()
	cfg.CpuCount = 2
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("scc ok")
	assert.Equal(t, "scc ok", c.Exec("scc 2"))
}

func TestPokerCuiController_SetCpuCount_LongCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultPokerConfig())
	cfg := domain.DefaultPokerConfig()
	cfg.CpuCount = 1
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("scc ok")
	assert.Equal(t, "scc ok", c.Exec("setcpucount 1"))
}

func TestPokerCuiController_SetCpuCount_MaxValue(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultPokerConfig())
	cfg := domain.DefaultPokerConfig()
	cfg.CpuCount = 3
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("scc ok")
	assert.Equal(t, "scc ok", c.Exec("scc 3"))
}

func TestPokerCuiController_SetCpuCount_NoArgs(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	assert.Contains(t, c.Exec("scc"), "CPU player count is required")
}

func TestPokerCuiController_SetCpuCount_InvalidValue(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// non-numeric: controller catches
	assert.Contains(t, c.Exec("scc abc"), "Invalid CPU player count: abc")
	// numeric out-of-range: controller catches via ParseIntArg bounds
	assert.Equal(t, "Invalid CPU player count: 0. Please enter 1-3.", c.Exec("scc 0"))
	assert.Equal(t, "Invalid CPU player count: 4. Please enter 1-3.", c.Exec("scc 4"))
	assert.Equal(t, "Invalid CPU player count: -1. Please enter 1-3.", c.Exec("scc -1"))
}

// --- set joker count ---

func TestPokerCuiController_SetJokerCount_Valid(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultPokerConfig())
	cfg := domain.DefaultPokerConfig()
	cfg.JokerCount = 1
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("sjc ok")
	assert.Equal(t, "sjc ok", c.Exec("sjc 1"))
}

func TestPokerCuiController_SetJokerCount_LongCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultPokerConfig())
	cfg := domain.DefaultPokerConfig()
	cfg.JokerCount = 2
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("sjc ok")
	assert.Equal(t, "sjc ok", c.Exec("setjokercount 2"))
}

func TestPokerCuiController_SetJokerCount_MinValue(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultPokerConfig())
	cfg := domain.DefaultPokerConfig()
	cfg.JokerCount = 0
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("sjc ok")
	assert.Equal(t, "sjc ok", c.Exec("sjc 0"))
}

func TestPokerCuiController_SetJokerCount_NoArgs(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	assert.Contains(t, c.Exec("sjc"), "Joker count is required")
}

func TestPokerCuiController_SetJokerCount_InvalidValue(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// non-numeric: controller catches
	assert.Contains(t, c.Exec("sjc abc"), "Invalid joker count: abc")
	// numeric out-of-range: controller catches via ParseIntArg bounds
	assert.Equal(t, "Invalid joker count: -1. Please enter 0-2.", c.Exec("sjc -1"))
	assert.Equal(t, "Invalid joker count: 3. Please enter 0-2.", c.Exec("sjc 3"))
}

// --- odds ---

func TestPokerCuiController_Odds_WithValidIndices(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Odds", []int{0, 2, 4}).Return("odds ok")
	assert.Equal(t, "odds ok", c.Exec("o 0 2 4"))
}

func TestPokerCuiController_Odds_LongCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Odds", []int{1, 3}).Return("odds ok")
	assert.Equal(t, "odds ok", c.Exec("odds 1 3"))
}

func TestPokerCuiController_Odds_NoIndices(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Odds", []int{}).Return("odds empty")
	assert.Equal(t, "odds empty", c.Exec("o"))
}

func TestPokerCuiController_Odds_InvalidIndex_NonNumeric(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Odds", []int{2}).Return("odds ok")
	result := c.Exec("o abc 2")
	assert.Contains(t, result, "'abc'")
	assert.Contains(t, result, "odds ok")
}

func TestPokerCuiController_Odds_InvalidIndex_OutOfRange(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Odds", []int{0}).Return("odds ok")
	result := c.Exec("o -1 0 5")
	assert.Contains(t, result, "'-1'")
	assert.Contains(t, result, "odds ok")
}

func TestPokerCuiController_Odds_AllInvalid(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Odds", []int{}).Return("odds empty")
	result := c.Exec("o abc -1 5 99")
	assert.Contains(t, result, "'abc'")
	assert.Contains(t, result, "odds empty")
}

// --- metaai ---

func TestPokerCuiController_MetaAI_On(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	cfg := domain.DefaultPokerConfig()
	mi.On("GetConfig").Return(cfg)
	expected := domain.DefaultPokerConfig()
	expected.CpuMetaAI = true
	mi.On("ResetWithConfig", expected, mock.Anything).Return("mai ok")
	assert.Equal(t, "mai ok", c.Exec("mai 1"))
}

func TestPokerCuiController_MetaAI_Off(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	cfg := domain.DefaultPokerConfig()
	mi.On("GetConfig").Return(cfg)
	expected := domain.DefaultPokerConfig()
	expected.CpuMetaAI = false
	mi.On("ResetWithConfig", expected, mock.Anything).Return("mai off")
	assert.Equal(t, "mai off", c.Exec("metaai 0"))
}

func TestPokerCuiController_MetaAI_MissingArg(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	assert.Equal(t, "メタAI設定が必要です (0=オフ, 1=オン)。", c.Exec("mai"))
}

func TestPokerCuiController_MetaAI_InvalidArg(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	assert.Equal(t, invalidArg("invalidMetaAI", "val", "abc"), c.Exec("mai abc"))
}
