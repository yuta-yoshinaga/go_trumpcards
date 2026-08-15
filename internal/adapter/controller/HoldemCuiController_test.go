package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestHoldemCuiController_Quit(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestHoldemCuiController_Reset(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Reset").Return("reset ok")
	assert.Equal(t, "reset ok", c.Exec("r"))
	assert.Equal(t, "reset ok", c.Exec("reset"))
}

func TestHoldemCuiController_Fold(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Action", domain.HoldemActionFold, 0, 0).Return("fold ok")
	assert.Equal(t, "fold ok", c.Exec("f"))
	assert.Equal(t, "fold ok", c.Exec("fold"))
}

func TestHoldemCuiController_Check(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Action", domain.HoldemActionCheck, 0, 0).Return("check ok")
	assert.Equal(t, "check ok", c.Exec("ck"))
	assert.Equal(t, "check ok", c.Exec("check"))
}

func TestHoldemCuiController_Call(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Action", domain.HoldemActionCall, 0, 0).Return("call ok")
	assert.Equal(t, "call ok", c.Exec("c"))
	assert.Equal(t, "call ok", c.Exec("call"))
}

func TestHoldemCuiController_Bet(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Action", domain.HoldemActionBet, 50, 0).Return("bet ok")
	assert.Equal(t, "bet ok", c.Exec("b 50"))
	assert.Equal(t, "bet ok", c.Exec("bet 50"))
}

func TestHoldemCuiController_Bet_NoAmount(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	result := c.Exec("b")
	assert.True(t, cuiutil.IsPromptRequest(result))
	_, tmpl := cuiutil.ParsePromptRequest(result)
	assert.Equal(t, "b {0}", tmpl)
}

func TestHoldemCuiController_Bet_InvalidAmount(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("b abc"), "無効な金額です")
	assert.Contains(t, c.Exec("bet xyz"), "無効な金額です")
}

func TestHoldemCuiController_Raise(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Action", domain.HoldemActionRaise, 30, 0).Return("raise ok")
	assert.Equal(t, "raise ok", c.Exec("ra 30"))
	assert.Equal(t, "raise ok", c.Exec("raise 30"))
}

func TestHoldemCuiController_Raise_NoAmount(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	result := c.Exec("ra")
	assert.True(t, cuiutil.IsPromptRequest(result))
	_, tmpl := cuiutil.ParsePromptRequest(result)
	assert.Equal(t, "ra {0}", tmpl)
}

func TestHoldemCuiController_Raise_InvalidAmount(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("ra abc"), "無効な金額です")
	assert.Contains(t, c.Exec("raise xyz"), "無効な金額です")
}

func TestHoldemCuiController_Bet_NonPositiveAmount(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("b -50"), "無効な金額です")
	assert.Contains(t, c.Exec("bet 0"), "無効な金額です")
}

func TestHoldemCuiController_Raise_NonPositiveAmount(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("ra -30"), "無効な金額です")
	assert.Contains(t, c.Exec("raise 0"), "無効な金額です")
}

func TestHoldemCuiController_AllIn(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Action", domain.HoldemActionAllIn, 0, 0).Return("allin ok")
	assert.Equal(t, "allin ok", c.Exec("a"))
	assert.Equal(t, "allin ok", c.Exec("allin"))
}

func TestHoldemCuiController_Empty(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

func TestHoldemCuiController_Unknown(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

// --- betting limit ---

func TestHoldemCuiController_BettingLimit_Valid(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.BettingLimit = domain.BettingLimitPotLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bl 1"))
}

func TestHoldemCuiController_BettingLimit_LongCommand(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.BettingLimit = domain.BettingLimitNoLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bettinglimit 2"))
}

func TestHoldemCuiController_BettingLimit_NoArgs(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("bl"), "ベッティングリミット")
}

func TestHoldemCuiController_BettingLimit_InvalidValue(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	// non-numeric: controller handles directly
	assert.Contains(t, c.Exec("bl abc"), "abc")
	// numeric out-of-range: delegated to interactor
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg5 := domain.DefaultHoldemConfig()
	cfg5.BettingLimit = domain.BettingLimitType(5)
	mi.On("ResetWithConfig", cfg5, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("bl 5"))
	cfgNeg := domain.DefaultHoldemConfig()
	cfgNeg.BettingLimit = domain.BettingLimitType(-1)
	mi.On("ResetWithConfig", cfgNeg, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("bl -1"))
}

// --- tournament mode ---

func TestHoldemCuiController_TournamentMode_Valid(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.TournamentMode = true
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("tm ok")
	assert.Equal(t, "tm ok", c.Exec("tm 1"))
}

func TestHoldemCuiController_TournamentMode_LongCommand(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.TournamentMode = false
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("tm ok")
	assert.Equal(t, "tm ok", c.Exec("tournament 0"))
}

func TestHoldemCuiController_TournamentMode_NoArgs(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("tm"), "トーナメントモード")
}

func TestHoldemCuiController_TournamentMode_InvalidValue(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("tm 5"), "5")
	assert.Contains(t, c.Exec("tm abc"), "abc")
	assert.Contains(t, c.Exec("tm -1"), "-1")
}

// --- small blind ---

func TestHoldemCuiController_SmallBlind_Valid(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.SmallBlind = 3
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("sb ok")
	assert.Equal(t, "sb ok", c.Exec("sb 3"))
}

func TestHoldemCuiController_SmallBlind_LongCommand(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.SmallBlind = 3
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("sb ok")
	assert.Equal(t, "sb ok", c.Exec("smallblind 3"))
}

func TestHoldemCuiController_SmallBlind_NoArgs(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("sb"), "スモールブラインド")
}

func TestHoldemCuiController_SmallBlind_InvalidValue(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	// non-numeric: controller handles directly
	assert.Contains(t, c.Exec("sb abc"), "abc")
	// numeric out-of-range: delegated to interactor
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg0 := domain.DefaultHoldemConfig()
	cfg0.SmallBlind = 0
	mi.On("ResetWithConfig", cfg0, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("sb 0"))
	cfgNeg := domain.DefaultHoldemConfig()
	cfgNeg.SmallBlind = -1
	mi.On("ResetWithConfig", cfgNeg, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("sb -1"))
}

func TestHoldemCuiController_SmallBlind_NotLessThanBigBlind(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	baseCfg := domain.HoldemConfig{SmallBlind: 5, BigBlind: 10, InitChips: 1000, BlindLevelHands: 10, BlindMultiplier: 200}
	mi.On("GetConfig").Return(baseCfg)
	cfg10 := baseCfg
	cfg10.SmallBlind = 10
	mi.On("ResetWithConfig", cfg10, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("sb 10"))
	cfg15 := baseCfg
	cfg15.SmallBlind = 15
	mi.On("ResetWithConfig", cfg15, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("sb 15"))
}

// --- big blind ---

func TestHoldemCuiController_BigBlind_Valid(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.BigBlind = 20
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bb ok")
	assert.Equal(t, "bb ok", c.Exec("bb 20"))
}

func TestHoldemCuiController_BigBlind_LongCommand(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.BigBlind = 20
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bb ok")
	assert.Equal(t, "bb ok", c.Exec("bigblind 20"))
}

func TestHoldemCuiController_BigBlind_NoArgs(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("bb"), "ビッグブラインド")
}

func TestHoldemCuiController_BigBlind_InvalidValue(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	// non-numeric: controller handles directly
	assert.Contains(t, c.Exec("bb abc"), "abc")
	// numeric out-of-range: delegated to interactor
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg1 := domain.DefaultHoldemConfig()
	cfg1.BigBlind = 1
	mi.On("ResetWithConfig", cfg1, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("bb 1"))
	cfgNeg := domain.DefaultHoldemConfig()
	cfgNeg.BigBlind = -1
	mi.On("ResetWithConfig", cfgNeg, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("bb -1"))
}

func TestHoldemCuiController_BigBlind_NotGreaterThanSmallBlind(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	baseCfg := domain.HoldemConfig{SmallBlind: 5, BigBlind: 10, InitChips: 1000, BlindLevelHands: 10, BlindMultiplier: 200}
	mi.On("GetConfig").Return(baseCfg)
	cfg5 := baseCfg
	cfg5.BigBlind = 5
	mi.On("ResetWithConfig", cfg5, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("bb 5"))
	cfg3 := baseCfg
	cfg3.BigBlind = 3
	mi.On("ResetWithConfig", cfg3, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("bb 3"))
}

// --- level-up hands ---

func TestHoldemCuiController_LevelHand_Valid(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.BlindLevelHands = 5
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("lh ok")
	assert.Equal(t, "lh ok", c.Exec("lh 5"))
}

func TestHoldemCuiController_LevelHand_LongCommand(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.BlindLevelHands = 5
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("lh ok")
	assert.Equal(t, "lh ok", c.Exec("levelhand 5"))
}

func TestHoldemCuiController_LevelHand_NoArgs(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("lh"), "ハンド数が必要")
}

func TestHoldemCuiController_LevelHand_InvalidValue(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	// non-numeric: controller handles directly
	assert.Contains(t, c.Exec("lh abc"), "abc")
	// numeric out-of-range: delegated to interactor
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg0 := domain.DefaultHoldemConfig()
	cfg0.BlindLevelHands = 0
	mi.On("ResetWithConfig", cfg0, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("lh 0"))
	cfgNeg := domain.DefaultHoldemConfig()
	cfgNeg.BlindLevelHands = -1
	mi.On("ResetWithConfig", cfgNeg, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("lh -1"))
}

// --- table size ---

func TestHoldemCuiController_TableSize_Valid(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.TableSize = domain.HoldemTableSize6
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("ts ok")
	assert.Equal(t, "ts ok", c.Exec("ts 6"))
}

func TestHoldemCuiController_TableSize_LongCommand(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.TableSize = domain.HoldemTableSize9
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("ts ok")
	assert.Equal(t, "ts ok", c.Exec("tablesize 9"))
}

func TestHoldemCuiController_TableSize_NoArgs(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("ts"), "テーブルサイズ")
}

func TestHoldemCuiController_TableSize_InvalidValue(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	// non-numeric: controller handles directly
	assert.Contains(t, c.Exec("ts abc"), "abc")
	// numeric out-of-range: delegated to interactor
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg5 := domain.DefaultHoldemConfig()
	cfg5.TableSize = 5
	mi.On("ResetWithConfig", cfg5, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("ts 5"))
	cfgNeg := domain.DefaultHoldemConfig()
	cfgNeg.TableSize = -1
	mi.On("ResetWithConfig", cfgNeg, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("ts -1"))
}

// --- rebuy ---

func TestHoldemCuiController_Rebuy(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Rebuy").Return("rebuy ok")
	assert.Equal(t, "rebuy ok", c.Exec("rb"))
	assert.Equal(t, "rebuy ok", c.Exec("rebuy"))
}

func TestHoldemCuiController_SkipRebuy(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("SkipRebuy").Return("skiprebuy ok")
	assert.Equal(t, "skiprebuy ok", c.Exec("sr"))
	assert.Equal(t, "skiprebuy ok", c.Exec("skiprebuy"))
}

// --- addon ---

func TestHoldemCuiController_Addon(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Addon").Return("addon ok")
	assert.Equal(t, "addon ok", c.Exec("ad"))
	assert.Equal(t, "addon ok", c.Exec("addon"))
}

func TestHoldemCuiController_SkipAddon(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("SkipAddon").Return("skipaddon ok")
	assert.Equal(t, "skipaddon ok", c.Exec("sa"))
	assert.Equal(t, "skipaddon ok", c.Exec("skipaddon"))
}

// --- muck / show ---

func TestHoldemCuiController_Muck(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Muck").Return("muck ok")
	assert.Equal(t, "muck ok", c.Exec("m"))
	assert.Equal(t, "muck ok", c.Exec("muck"))
}

func TestHoldemCuiController_ShowHand(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("ShowHand").Return("show ok")
	assert.Equal(t, "show ok", c.Exec("sh"))
	assert.Equal(t, "show ok", c.Exec("show"))
}

// --- metaai ---

func TestHoldemCuiController_MetaAI_On(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	cfg := domain.DefaultHoldemConfig()
	mi.On("GetConfig").Return(cfg)
	expected := domain.DefaultHoldemConfig()
	expected.CpuMetaAI = true
	mi.On("ResetWithConfig", expected, mock.Anything).Return("mai ok")
	assert.Equal(t, "mai ok", c.Exec("mai 1"))
}

func TestHoldemCuiController_MetaAI_Off(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	cfg := domain.DefaultHoldemConfig()
	mi.On("GetConfig").Return(cfg)
	expected := domain.DefaultHoldemConfig()
	expected.CpuMetaAI = false
	mi.On("ResetWithConfig", expected, mock.Anything).Return("mai off")
	assert.Equal(t, "mai off", c.Exec("metaai 0"))
}

func TestHoldemCuiController_MetaAI_MissingArg(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Equal(t, "メタAI設定が必要です (0=オフ, 1=オン)。", c.Exec("mai"))
}

func TestHoldemCuiController_MetaAI_InvalidArg(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Equal(t, invalidArg("invalidMetaAI", "val", "abc"), c.Exec("mai abc"))
}
