package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestFiveCardStudCuiController_Quit(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestFiveCardStudCuiController_Reset(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("Reset").Return("reset ok")
	assert.Equal(t, "reset ok", c.Exec("r"))
	assert.Equal(t, "reset ok", c.Exec("reset"))
}

func TestFiveCardStudCuiController_Fold(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("Action", domain.FiveCardStudActionFold, 0, 0).Return("fold ok")
	assert.Equal(t, "fold ok", c.Exec("f"))
	assert.Equal(t, "fold ok", c.Exec("fold"))
}

func TestFiveCardStudCuiController_Check(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("Action", domain.FiveCardStudActionCheck, 0, 0).Return("check ok")
	assert.Equal(t, "check ok", c.Exec("ck"))
	assert.Equal(t, "check ok", c.Exec("check"))
}

func TestFiveCardStudCuiController_Call(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("Action", domain.FiveCardStudActionCall, 0, 0).Return("call ok")
	assert.Equal(t, "call ok", c.Exec("c"))
	assert.Equal(t, "call ok", c.Exec("call"))
}

func TestFiveCardStudCuiController_Bet(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("Action", domain.FiveCardStudActionBet, 50, 0).Return("bet ok")
	assert.Equal(t, "bet ok", c.Exec("b 50"))
	assert.Equal(t, "bet ok", c.Exec("bet 50"))
}

func TestFiveCardStudCuiController_Bet_NoAmount(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	result := c.Exec("b")
	assert.True(t, cuiutil.IsPromptRequest(result))
	_, tmpl := cuiutil.ParsePromptRequest(result)
	assert.Equal(t, "b {0}", tmpl)
}

func TestFiveCardStudCuiController_Bet_InvalidAmount(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("b abc"), "無効な金額です")
	assert.Contains(t, c.Exec("bet xyz"), "無効な金額です")
}

func TestFiveCardStudCuiController_Raise(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("Action", domain.FiveCardStudActionRaise, 30, 0).Return("raise ok")
	assert.Equal(t, "raise ok", c.Exec("ra 30"))
	assert.Equal(t, "raise ok", c.Exec("raise 30"))
}

func TestFiveCardStudCuiController_Raise_NoAmount(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	result := c.Exec("ra")
	assert.True(t, cuiutil.IsPromptRequest(result))
	_, tmpl := cuiutil.ParsePromptRequest(result)
	assert.Equal(t, "ra {0}", tmpl)
}

func TestFiveCardStudCuiController_Raise_InvalidAmount(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("ra abc"), "無効な金額です")
	assert.Contains(t, c.Exec("raise xyz"), "無効な金額です")
}

func TestFiveCardStudCuiController_Bet_NonPositiveAmount(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("b -50"), "無効な金額です")
	assert.Contains(t, c.Exec("bet 0"), "無効な金額です")
}

func TestFiveCardStudCuiController_Raise_NonPositiveAmount(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("ra -30"), "無効な金額です")
	assert.Contains(t, c.Exec("raise 0"), "無効な金額です")
}

func TestFiveCardStudCuiController_AllIn(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("Action", domain.FiveCardStudActionAllIn, 0, 0).Return("allin ok")
	assert.Equal(t, "allin ok", c.Exec("a"))
	assert.Equal(t, "allin ok", c.Exec("allin"))
}

func TestFiveCardStudCuiController_Empty(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

func TestFiveCardStudCuiController_Unknown(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

// --- betting limit ---

func TestFiveCardStudCuiController_BettingLimit_Valid(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFiveCardStudConfig())
	cfg := domain.DefaultFiveCardStudConfig()
	cfg.BettingLimit = domain.BettingLimitPotLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bl 1"))
}

func TestFiveCardStudCuiController_BettingLimit_LongCommand(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFiveCardStudConfig())
	cfg := domain.DefaultFiveCardStudConfig()
	cfg.BettingLimit = domain.BettingLimitNoLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bettinglimit 2"))
}

func TestFiveCardStudCuiController_BettingLimit_NoArgs(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("bl"), "ベッティングリミット")
}

func TestFiveCardStudCuiController_BettingLimit_InvalidValue(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("bl abc"), "abc")
}

// --- tournament mode ---

func TestFiveCardStudCuiController_TournamentMode_Valid(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFiveCardStudConfig())
	cfg := domain.DefaultFiveCardStudConfig()
	cfg.TournamentMode = true
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("tm ok")
	assert.Equal(t, "tm ok", c.Exec("tm 1"))
}

func TestFiveCardStudCuiController_TournamentMode_LongCommand(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFiveCardStudConfig())
	cfg := domain.DefaultFiveCardStudConfig()
	cfg.TournamentMode = false
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("tm ok")
	assert.Equal(t, "tm ok", c.Exec("tournament 0"))
}

func TestFiveCardStudCuiController_TournamentMode_NoArgs(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("tm"), "トーナメントモード")
}

func TestFiveCardStudCuiController_TournamentMode_InvalidValue(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("tm 5"), "5")
	assert.Contains(t, c.Exec("tm abc"), "abc")
	assert.Contains(t, c.Exec("tm -1"), "-1")
}

// --- ante ---

func TestFiveCardStudCuiController_Ante_Valid(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFiveCardStudConfig())
	cfg := domain.DefaultFiveCardStudConfig()
	cfg.Ante = 3
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("ante ok")
	assert.Equal(t, "ante ok", c.Exec("ante 3"))
}

func TestFiveCardStudCuiController_Ante_NoArgs(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("ante"), "アンティ")
}

func TestFiveCardStudCuiController_Ante_InvalidValue(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("ante abc"), "アンティ")
}

// --- bring-in ---

func TestFiveCardStudCuiController_BringIn_Valid(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFiveCardStudConfig())
	cfg := domain.DefaultFiveCardStudConfig()
	cfg.BringIn = 5
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bi ok")
	assert.Equal(t, "bi ok", c.Exec("bi 5"))
}

func TestFiveCardStudCuiController_BringIn_LongCommand(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFiveCardStudConfig())
	cfg := domain.DefaultFiveCardStudConfig()
	cfg.BringIn = 5
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bi ok")
	assert.Equal(t, "bi ok", c.Exec("bringin 5"))
}

func TestFiveCardStudCuiController_BringIn_NoArgs(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("bi"), "ブリングイン")
}

func TestFiveCardStudCuiController_BringIn_InvalidValue(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("bi abc"), "ブリングイン")
}

// --- small bet ---

func TestFiveCardStudCuiController_SmallBet_Valid(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFiveCardStudConfig())
	cfg := domain.DefaultFiveCardStudConfig()
	cfg.SmallBet = 10
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("sb ok")
	assert.Equal(t, "sb ok", c.Exec("sb 10"))
}

func TestFiveCardStudCuiController_SmallBet_LongCommand(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFiveCardStudConfig())
	cfg := domain.DefaultFiveCardStudConfig()
	cfg.SmallBet = 10
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("sb ok")
	assert.Equal(t, "sb ok", c.Exec("smallbet 10"))
}

func TestFiveCardStudCuiController_SmallBet_NoArgs(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("sb"), "スモールベット")
}

func TestFiveCardStudCuiController_SmallBet_InvalidValue(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("sb abc"), "スモールベット")
}

// --- big bet ---

func TestFiveCardStudCuiController_BigBet_Valid(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFiveCardStudConfig())
	cfg := domain.DefaultFiveCardStudConfig()
	cfg.BigBet = 20
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bb ok")
	assert.Equal(t, "bb ok", c.Exec("bb 20"))
}

func TestFiveCardStudCuiController_BigBet_LongCommand(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFiveCardStudConfig())
	cfg := domain.DefaultFiveCardStudConfig()
	cfg.BigBet = 20
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bb ok")
	assert.Equal(t, "bb ok", c.Exec("bigbet 20"))
}

func TestFiveCardStudCuiController_BigBet_NoArgs(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("bb"), "ビッグベット")
}

func TestFiveCardStudCuiController_BigBet_InvalidValue(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("bb abc"), "ビッグベット")
}

// --- level hand ---

func TestFiveCardStudCuiController_LevelHand_Valid(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFiveCardStudConfig())
	cfg := domain.DefaultFiveCardStudConfig()
	cfg.AnteLevelHands = 5
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("lh ok")
	assert.Equal(t, "lh ok", c.Exec("lh 5"))
}

func TestFiveCardStudCuiController_LevelHand_LongCommand(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFiveCardStudConfig())
	cfg := domain.DefaultFiveCardStudConfig()
	cfg.AnteLevelHands = 5
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("lh ok")
	assert.Equal(t, "lh ok", c.Exec("levelhand 5"))
}

func TestFiveCardStudCuiController_LevelHand_NoArgs(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("lh"), "ハンド数が必要")
}

func TestFiveCardStudCuiController_LevelHand_InvalidValue(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("lh abc"), "abc")
}

// --- table size ---

func TestFiveCardStudCuiController_TableSize_Valid(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFiveCardStudConfig())
	cfg := domain.DefaultFiveCardStudConfig()
	cfg.TableSize = 4
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("ts ok")
	assert.Equal(t, "ts ok", c.Exec("ts 4"))
}

func TestFiveCardStudCuiController_TableSize_LongCommand(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFiveCardStudConfig())
	cfg := domain.DefaultFiveCardStudConfig()
	cfg.TableSize = 6
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("ts ok")
	assert.Equal(t, "ts ok", c.Exec("tablesize 6"))
}

func TestFiveCardStudCuiController_TableSize_NoArgs(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("ts"), "テーブルサイズ")
}

func TestFiveCardStudCuiController_TableSize_InvalidValue(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Contains(t, c.Exec("ts abc"), "abc")
}

// --- rebuy ---

func TestFiveCardStudCuiController_Rebuy(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("Rebuy").Return("rebuy ok")
	assert.Equal(t, "rebuy ok", c.Exec("rb"))
	assert.Equal(t, "rebuy ok", c.Exec("rebuy"))
}

func TestFiveCardStudCuiController_SkipRebuy(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("SkipRebuy").Return("skiprebuy ok")
	assert.Equal(t, "skiprebuy ok", c.Exec("sr"))
	assert.Equal(t, "skiprebuy ok", c.Exec("skiprebuy"))
}

// --- addon ---

func TestFiveCardStudCuiController_Addon(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("Addon").Return("addon ok")
	assert.Equal(t, "addon ok", c.Exec("ad"))
	assert.Equal(t, "addon ok", c.Exec("addon"))
}

func TestFiveCardStudCuiController_SkipAddon(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("SkipAddon").Return("skipaddon ok")
	assert.Equal(t, "skipaddon ok", c.Exec("sa"))
	assert.Equal(t, "skipaddon ok", c.Exec("skipaddon"))
}

// --- muck / show ---

func TestFiveCardStudCuiController_Muck(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("Muck").Return("muck ok")
	assert.Equal(t, "muck ok", c.Exec("m"))
	assert.Equal(t, "muck ok", c.Exec("muck"))
}

func TestFiveCardStudCuiController_ShowHand(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("ShowHand").Return("show ok")
	assert.Equal(t, "show ok", c.Exec("sh"))
	assert.Equal(t, "show ok", c.Exec("show"))
}

// --- metaai ---

func TestFiveCardStudCuiController_MetaAI_On(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	cfg := domain.DefaultFiveCardStudConfig()
	mi.On("GetConfig").Return(cfg)
	expected := domain.DefaultFiveCardStudConfig()
	expected.CpuMetaAI = true
	mi.On("ResetWithConfig", expected, mock.Anything).Return("mai ok")
	assert.Equal(t, "mai ok", c.Exec("mai 1"))
}

func TestFiveCardStudCuiController_MetaAI_Off(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	cfg := domain.DefaultFiveCardStudConfig()
	mi.On("GetConfig").Return(cfg)
	expected := domain.DefaultFiveCardStudConfig()
	expected.CpuMetaAI = false
	mi.On("ResetWithConfig", expected, mock.Anything).Return("mai off")
	assert.Equal(t, "mai off", c.Exec("metaai 0"))
}

func TestFiveCardStudCuiController_MetaAI_MissingArg(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Equal(t, "メタAI設定が必要です (0=オフ, 1=オン)。", c.Exec("mai"))
}

func TestFiveCardStudCuiController_MetaAI_InvalidArg(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	assert.Equal(t, invalidArg("invalidMetaAI", "val", "abc"), c.Exec("mai abc"))
}

// --- log ---

func TestFiveCardStudCuiController_Log(t *testing.T) {
	mi := new(usecase.MockFiveCardStudInteractor)
	c := NewFiveCardStudCuiController(mi)
	mi.On("ActionLog").Return("log output")
	assert.Equal(t, "log output", c.Exec("log"))
	assert.Equal(t, "log output", c.Exec("l"))
}
