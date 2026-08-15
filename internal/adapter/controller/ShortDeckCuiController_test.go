package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestShortDeckCuiController_Quit(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestShortDeckCuiController_Reset(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("Reset").Return("reset ok")
	assert.Equal(t, "reset ok", c.Exec("r"))
	assert.Equal(t, "reset ok", c.Exec("reset"))
}

func TestShortDeckCuiController_Fold(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("Action", domain.ShortDeckActionFold, 0, 0).Return("fold ok")
	assert.Equal(t, "fold ok", c.Exec("f"))
	assert.Equal(t, "fold ok", c.Exec("fold"))
}

func TestShortDeckCuiController_Check(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("Action", domain.ShortDeckActionCheck, 0, 0).Return("check ok")
	assert.Equal(t, "check ok", c.Exec("ck"))
	assert.Equal(t, "check ok", c.Exec("check"))
}

func TestShortDeckCuiController_Call(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("Action", domain.ShortDeckActionCall, 0, 0).Return("call ok")
	assert.Equal(t, "call ok", c.Exec("c"))
	assert.Equal(t, "call ok", c.Exec("call"))
}

func TestShortDeckCuiController_Bet(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("Action", domain.ShortDeckActionBet, 50, 0).Return("bet ok")
	assert.Equal(t, "bet ok", c.Exec("b 50"))
	assert.Equal(t, "bet ok", c.Exec("bet 50"))
}

func TestShortDeckCuiController_Bet_NoAmount(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	assert.True(t, cuiutil.IsPromptRequest(c.Exec("b")))
}

func TestShortDeckCuiController_Bet_InvalidAmount(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	assert.Contains(t, c.Exec("b abc"), "無効な金額です")
	assert.Contains(t, c.Exec("bet xyz"), "無効な金額です")
}

func TestShortDeckCuiController_Raise(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("Action", domain.ShortDeckActionRaise, 30, 0).Return("raise ok")
	assert.Equal(t, "raise ok", c.Exec("ra 30"))
	assert.Equal(t, "raise ok", c.Exec("raise 30"))
}

func TestShortDeckCuiController_Raise_NoAmount(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	assert.True(t, cuiutil.IsPromptRequest(c.Exec("ra")))
}

func TestShortDeckCuiController_Raise_InvalidAmount(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	assert.Contains(t, c.Exec("ra abc"), "無効な金額です")
	assert.Contains(t, c.Exec("raise xyz"), "無効な金額です")
}

func TestShortDeckCuiController_Bet_NonPositiveAmount(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	assert.Contains(t, c.Exec("b -50"), "無効な金額です")
	assert.Contains(t, c.Exec("bet 0"), "無効な金額です")
}

func TestShortDeckCuiController_Raise_NonPositiveAmount(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	assert.Contains(t, c.Exec("ra -30"), "無効な金額です")
	assert.Contains(t, c.Exec("raise 0"), "無効な金額です")
}

func TestShortDeckCuiController_AllIn(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("Action", domain.ShortDeckActionAllIn, 0, 0).Return("allin ok")
	assert.Equal(t, "allin ok", c.Exec("a"))
	assert.Equal(t, "allin ok", c.Exec("allin"))
}

func TestShortDeckCuiController_Empty(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

func TestShortDeckCuiController_Unknown(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

// --- betting limit ---

func TestShortDeckCuiController_BettingLimit_Valid(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultShortDeckConfig())
	cfg := domain.DefaultShortDeckConfig()
	cfg.BettingLimit = domain.BettingLimitPotLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bl 1"))
}

func TestShortDeckCuiController_BettingLimit_LongCommand(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultShortDeckConfig())
	cfg := domain.DefaultShortDeckConfig()
	cfg.BettingLimit = domain.BettingLimitNoLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bettinglimit 2"))
}

func TestShortDeckCuiController_BettingLimit_NoArgs(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	assert.Contains(t, c.Exec("bl"), "ベッティングリミット")
}

func TestShortDeckCuiController_BettingLimit_InvalidValue(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	// non-numeric: controller handles directly
	assert.Contains(t, c.Exec("bl abc"), "abc")
	// numeric out-of-range: delegated to interactor
	mi.On("GetConfig").Return(domain.DefaultShortDeckConfig())
	cfg5 := domain.DefaultShortDeckConfig()
	cfg5.BettingLimit = domain.BettingLimitType(5)
	mi.On("ResetWithConfig", cfg5, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("bl 5"))
	cfgNeg := domain.DefaultShortDeckConfig()
	cfgNeg.BettingLimit = domain.BettingLimitType(-1)
	mi.On("ResetWithConfig", cfgNeg, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("bl -1"))
}

// --- tournament mode ---

func TestShortDeckCuiController_TournamentMode_Valid(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultShortDeckConfig())
	cfg := domain.DefaultShortDeckConfig()
	cfg.TournamentMode = true
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("tm ok")
	assert.Equal(t, "tm ok", c.Exec("tm 1"))
}

func TestShortDeckCuiController_TournamentMode_LongCommand(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultShortDeckConfig())
	cfg := domain.DefaultShortDeckConfig()
	cfg.TournamentMode = false
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("tm ok")
	assert.Equal(t, "tm ok", c.Exec("tournament 0"))
}

func TestShortDeckCuiController_TournamentMode_NoArgs(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	assert.Contains(t, c.Exec("tm"), "トーナメントモード")
}

func TestShortDeckCuiController_TournamentMode_InvalidValue(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	assert.Contains(t, c.Exec("tm 5"), "5")
	assert.Contains(t, c.Exec("tm abc"), "abc")
	assert.Contains(t, c.Exec("tm -1"), "-1")
}

// --- small blind ---

func TestShortDeckCuiController_SmallBlind_Valid(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultShortDeckConfig())
	cfg := domain.DefaultShortDeckConfig()
	cfg.SmallBlind = 3
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("sb ok")
	assert.Equal(t, "sb ok", c.Exec("sb 3"))
}

func TestShortDeckCuiController_SmallBlind_LongCommand(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultShortDeckConfig())
	cfg := domain.DefaultShortDeckConfig()
	cfg.SmallBlind = 3
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("sb ok")
	assert.Equal(t, "sb ok", c.Exec("smallblind 3"))
}

func TestShortDeckCuiController_SmallBlind_NoArgs(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	assert.Contains(t, c.Exec("sb"), "スモールブラインド")
}

func TestShortDeckCuiController_SmallBlind_InvalidValue(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	// non-numeric: controller handles directly
	assert.Contains(t, c.Exec("sb abc"), "abc")
	// numeric out-of-range: delegated to interactor
	mi.On("GetConfig").Return(domain.DefaultShortDeckConfig())
	cfg0 := domain.DefaultShortDeckConfig()
	cfg0.SmallBlind = 0
	mi.On("ResetWithConfig", cfg0, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("sb 0"))
	cfgNeg := domain.DefaultShortDeckConfig()
	cfgNeg.SmallBlind = -1
	mi.On("ResetWithConfig", cfgNeg, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("sb -1"))
}

func TestShortDeckCuiController_SmallBlind_NotLessThanBigBlind(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	baseCfg := domain.ShortDeckConfig{SmallBlind: 5, BigBlind: 10, InitChips: 1000, BlindLevelHands: 10, BlindMultiplier: 200}
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

func TestShortDeckCuiController_BigBlind_Valid(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultShortDeckConfig())
	cfg := domain.DefaultShortDeckConfig()
	cfg.BigBlind = 20
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bb ok")
	assert.Equal(t, "bb ok", c.Exec("bb 20"))
}

func TestShortDeckCuiController_BigBlind_LongCommand(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultShortDeckConfig())
	cfg := domain.DefaultShortDeckConfig()
	cfg.BigBlind = 20
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bb ok")
	assert.Equal(t, "bb ok", c.Exec("bigblind 20"))
}

func TestShortDeckCuiController_BigBlind_NoArgs(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	assert.Contains(t, c.Exec("bb"), "ビッグブラインド")
}

func TestShortDeckCuiController_BigBlind_InvalidValue(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	// non-numeric: controller handles directly
	assert.Contains(t, c.Exec("bb abc"), "abc")
	// numeric out-of-range: delegated to interactor
	mi.On("GetConfig").Return(domain.DefaultShortDeckConfig())
	cfg1 := domain.DefaultShortDeckConfig()
	cfg1.BigBlind = 1
	mi.On("ResetWithConfig", cfg1, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("bb 1"))
	cfgNeg := domain.DefaultShortDeckConfig()
	cfgNeg.BigBlind = -1
	mi.On("ResetWithConfig", cfgNeg, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("bb -1"))
}

func TestShortDeckCuiController_BigBlind_NotGreaterThanSmallBlind(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	baseCfg := domain.ShortDeckConfig{SmallBlind: 5, BigBlind: 10, InitChips: 1000, BlindLevelHands: 10, BlindMultiplier: 200}
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

func TestShortDeckCuiController_LevelHand_Valid(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultShortDeckConfig())
	cfg := domain.DefaultShortDeckConfig()
	cfg.BlindLevelHands = 5
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("lh ok")
	assert.Equal(t, "lh ok", c.Exec("lh 5"))
}

func TestShortDeckCuiController_LevelHand_LongCommand(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultShortDeckConfig())
	cfg := domain.DefaultShortDeckConfig()
	cfg.BlindLevelHands = 5
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("lh ok")
	assert.Equal(t, "lh ok", c.Exec("levelhand 5"))
}

func TestShortDeckCuiController_LevelHand_NoArgs(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	assert.Contains(t, c.Exec("lh"), "ハンド数が必要")
}

func TestShortDeckCuiController_LevelHand_InvalidValue(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	// non-numeric: controller handles directly
	assert.Contains(t, c.Exec("lh abc"), "abc")
	// numeric out-of-range: delegated to interactor
	mi.On("GetConfig").Return(domain.DefaultShortDeckConfig())
	cfg0 := domain.DefaultShortDeckConfig()
	cfg0.BlindLevelHands = 0
	mi.On("ResetWithConfig", cfg0, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("lh 0"))
	cfgNeg := domain.DefaultShortDeckConfig()
	cfgNeg.BlindLevelHands = -1
	mi.On("ResetWithConfig", cfgNeg, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("lh -1"))
}

// --- table size ---

func TestShortDeckCuiController_TableSize_Valid(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultShortDeckConfig())
	cfg := domain.DefaultShortDeckConfig()
	cfg.TableSize = domain.HoldemTableSize6
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("ts ok")
	assert.Equal(t, "ts ok", c.Exec("ts 6"))
}

func TestShortDeckCuiController_TableSize_LongCommand(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultShortDeckConfig())
	cfg := domain.DefaultShortDeckConfig()
	cfg.TableSize = domain.HoldemTableSize9
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("ts ok")
	assert.Equal(t, "ts ok", c.Exec("tablesize 9"))
}

func TestShortDeckCuiController_TableSize_NoArgs(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	assert.Contains(t, c.Exec("ts"), "テーブルサイズ")
}

func TestShortDeckCuiController_TableSize_InvalidValue(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	// non-numeric: controller handles directly
	assert.Contains(t, c.Exec("ts abc"), "abc")
	// numeric out-of-range: delegated to interactor
	mi.On("GetConfig").Return(domain.DefaultShortDeckConfig())
	cfg5 := domain.DefaultShortDeckConfig()
	cfg5.TableSize = 5
	mi.On("ResetWithConfig", cfg5, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("ts 5"))
	cfgNeg := domain.DefaultShortDeckConfig()
	cfgNeg.TableSize = -1
	mi.On("ResetWithConfig", cfgNeg, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("ts -1"))
}

// --- rebuy ---

func TestShortDeckCuiController_Rebuy(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("Rebuy").Return("rebuy ok")
	assert.Equal(t, "rebuy ok", c.Exec("rb"))
	assert.Equal(t, "rebuy ok", c.Exec("rebuy"))
}

func TestShortDeckCuiController_SkipRebuy(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("SkipRebuy").Return("skiprebuy ok")
	assert.Equal(t, "skiprebuy ok", c.Exec("sr"))
	assert.Equal(t, "skiprebuy ok", c.Exec("skiprebuy"))
}

// --- addon ---

func TestShortDeckCuiController_Addon(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("Addon").Return("addon ok")
	assert.Equal(t, "addon ok", c.Exec("ad"))
	assert.Equal(t, "addon ok", c.Exec("addon"))
}

func TestShortDeckCuiController_SkipAddon(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("SkipAddon").Return("skipaddon ok")
	assert.Equal(t, "skipaddon ok", c.Exec("sa"))
	assert.Equal(t, "skipaddon ok", c.Exec("skipaddon"))
}

// --- muck / show ---

func TestShortDeckCuiController_Muck(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("Muck").Return("muck ok")
	assert.Equal(t, "muck ok", c.Exec("m"))
	assert.Equal(t, "muck ok", c.Exec("muck"))
}

func TestShortDeckCuiController_ShowHand(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	mi.On("ShowHand").Return("show ok")
	assert.Equal(t, "show ok", c.Exec("sh"))
	assert.Equal(t, "show ok", c.Exec("show"))
}

// --- metaai ---

func TestShortDeckCuiController_MetaAI_On(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	cfg := domain.DefaultShortDeckConfig()
	mi.On("GetConfig").Return(cfg)
	expected := domain.DefaultShortDeckConfig()
	expected.CpuMetaAI = true
	mi.On("ResetWithConfig", expected, mock.Anything).Return("mai ok")
	assert.Equal(t, "mai ok", c.Exec("mai 1"))
}

func TestShortDeckCuiController_MetaAI_Off(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	cfg := domain.DefaultShortDeckConfig()
	mi.On("GetConfig").Return(cfg)
	expected := domain.DefaultShortDeckConfig()
	expected.CpuMetaAI = false
	mi.On("ResetWithConfig", expected, mock.Anything).Return("mai off")
	assert.Equal(t, "mai off", c.Exec("metaai 0"))
}

func TestShortDeckCuiController_MetaAI_MissingArg(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	assert.Equal(t, "メタAI設定が必要です (0=オフ, 1=オン)。", c.Exec("mai"))
}

func TestShortDeckCuiController_MetaAI_InvalidArg(t *testing.T) {
	mi := new(usecase.MockShortDeckInteractor)
	c := NewShortDeckCuiController(mi)
	assert.Equal(t, invalidArg("invalidMetaAI", "val", "abc"), c.Exec("mai abc"))
}
