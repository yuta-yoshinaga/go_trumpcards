package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestFollowTheQueenCuiController_Quit(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestFollowTheQueenCuiController_Reset(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("Reset").Return("reset ok")
	assert.Equal(t, "reset ok", c.Exec("r"))
	assert.Equal(t, "reset ok", c.Exec("reset"))
}

func TestFollowTheQueenCuiController_Fold(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("Action", domain.FollowTheQueenActionFold, 0, 0).Return("fold ok")
	assert.Equal(t, "fold ok", c.Exec("f"))
	assert.Equal(t, "fold ok", c.Exec("fold"))
}

func TestFollowTheQueenCuiController_Check(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("Action", domain.FollowTheQueenActionCheck, 0, 0).Return("check ok")
	assert.Equal(t, "check ok", c.Exec("ck"))
	assert.Equal(t, "check ok", c.Exec("check"))
}

func TestFollowTheQueenCuiController_Call(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("Action", domain.FollowTheQueenActionCall, 0, 0).Return("call ok")
	assert.Equal(t, "call ok", c.Exec("c"))
	assert.Equal(t, "call ok", c.Exec("call"))
}

func TestFollowTheQueenCuiController_Bet(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("Action", domain.FollowTheQueenActionBet, 50, 0).Return("bet ok")
	assert.Equal(t, "bet ok", c.Exec("b 50"))
	assert.Equal(t, "bet ok", c.Exec("bet 50"))
}

func TestFollowTheQueenCuiController_Bet_NoAmount(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	result := c.Exec("b")
	assert.True(t, cuiutil.IsPromptRequest(result))
	_, tmpl := cuiutil.ParsePromptRequest(result)
	assert.Equal(t, "b {0}", tmpl)
}

func TestFollowTheQueenCuiController_Bet_InvalidAmount(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Contains(t, c.Exec("b abc"), "無効な金額です")
	assert.Contains(t, c.Exec("bet xyz"), "無効な金額です")
}

func TestFollowTheQueenCuiController_Raise(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("Action", domain.FollowTheQueenActionRaise, 30, 0).Return("raise ok")
	assert.Equal(t, "raise ok", c.Exec("ra 30"))
	assert.Equal(t, "raise ok", c.Exec("raise 30"))
}

func TestFollowTheQueenCuiController_Raise_NoAmount(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	result := c.Exec("ra")
	assert.True(t, cuiutil.IsPromptRequest(result))
	_, tmpl := cuiutil.ParsePromptRequest(result)
	assert.Equal(t, "ra {0}", tmpl)
}

func TestFollowTheQueenCuiController_Raise_InvalidAmount(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Contains(t, c.Exec("ra abc"), "無効な金額です")
	assert.Contains(t, c.Exec("raise xyz"), "無効な金額です")
}

func TestFollowTheQueenCuiController_Bet_NonPositiveAmount(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Contains(t, c.Exec("b -50"), "無効な金額です")
	assert.Contains(t, c.Exec("bet 0"), "無効な金額です")
}

func TestFollowTheQueenCuiController_Raise_NonPositiveAmount(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Contains(t, c.Exec("ra -30"), "無効な金額です")
	assert.Contains(t, c.Exec("raise 0"), "無効な金額です")
}

func TestFollowTheQueenCuiController_AllIn(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("Action", domain.FollowTheQueenActionAllIn, 0, 0).Return("allin ok")
	assert.Equal(t, "allin ok", c.Exec("a"))
	assert.Equal(t, "allin ok", c.Exec("allin"))
}

func TestFollowTheQueenCuiController_Empty(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

func TestFollowTheQueenCuiController_Unknown(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

// --- betting limit ---

func TestFollowTheQueenCuiController_BettingLimit_Valid(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFollowTheQueenConfig())
	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.BettingLimit = domain.BettingLimitPotLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bl 1"))
}

func TestFollowTheQueenCuiController_BettingLimit_LongCommand(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFollowTheQueenConfig())
	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.BettingLimit = domain.BettingLimitNoLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bettinglimit 2"))
}

func TestFollowTheQueenCuiController_BettingLimit_NoArgs(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Contains(t, c.Exec("bl"), "ベッティングリミット")
}

func TestFollowTheQueenCuiController_BettingLimit_InvalidValue(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Contains(t, c.Exec("bl abc"), "abc")
}

// --- tournament mode ---

func TestFollowTheQueenCuiController_TournamentMode_Valid(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFollowTheQueenConfig())
	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.TournamentMode = true
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("tm ok")
	assert.Equal(t, "tm ok", c.Exec("tm 1"))
}

func TestFollowTheQueenCuiController_TournamentMode_LongCommand(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFollowTheQueenConfig())
	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.TournamentMode = false
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("tm ok")
	assert.Equal(t, "tm ok", c.Exec("tournament 0"))
}

func TestFollowTheQueenCuiController_TournamentMode_NoArgs(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Contains(t, c.Exec("tm"), "トーナメントモード")
}

func TestFollowTheQueenCuiController_TournamentMode_InvalidValue(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Contains(t, c.Exec("tm 5"), "5")
	assert.Contains(t, c.Exec("tm abc"), "abc")
	assert.Contains(t, c.Exec("tm -1"), "-1")
}

// --- ante ---

func TestFollowTheQueenCuiController_Ante_Valid(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFollowTheQueenConfig())
	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.Ante = 3
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("ante ok")
	assert.Equal(t, "ante ok", c.Exec("ante 3"))
}

func TestFollowTheQueenCuiController_Ante_NoArgs(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	result := c.Exec("ante")
	assert.Contains(t, result, "アンティ")
}

func TestFollowTheQueenCuiController_Ante_InvalidValue(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Contains(t, c.Exec("ante abc"), "アンティ")
}

// --- bring-in ---

func TestFollowTheQueenCuiController_BringIn_Valid(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFollowTheQueenConfig())
	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.BringIn = 5
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bi ok")
	assert.Equal(t, "bi ok", c.Exec("bi 5"))
}

func TestFollowTheQueenCuiController_BringIn_LongCommand(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFollowTheQueenConfig())
	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.BringIn = 5
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bi ok")
	assert.Equal(t, "bi ok", c.Exec("bringin 5"))
}

func TestFollowTheQueenCuiController_BringIn_NoArgs(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	result := c.Exec("bi")
	assert.Contains(t, result, "ブリングイン")
}

func TestFollowTheQueenCuiController_BringIn_InvalidValue(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Contains(t, c.Exec("bi abc"), "ブリングイン")
}

// --- small bet ---

func TestFollowTheQueenCuiController_SmallBet_Valid(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFollowTheQueenConfig())
	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.SmallBet = 10
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("sb ok")
	assert.Equal(t, "sb ok", c.Exec("sb 10"))
}

func TestFollowTheQueenCuiController_SmallBet_LongCommand(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFollowTheQueenConfig())
	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.SmallBet = 10
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("sb ok")
	assert.Equal(t, "sb ok", c.Exec("smallbet 10"))
}

func TestFollowTheQueenCuiController_SmallBet_NoArgs(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	result := c.Exec("sb")
	assert.Contains(t, result, "スモールベット")
}

func TestFollowTheQueenCuiController_SmallBet_InvalidValue(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Contains(t, c.Exec("sb abc"), "スモールベット")
}

// --- big bet ---

func TestFollowTheQueenCuiController_BigBet_Valid(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFollowTheQueenConfig())
	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.BigBet = 20
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bb ok")
	assert.Equal(t, "bb ok", c.Exec("bb 20"))
}

func TestFollowTheQueenCuiController_BigBet_LongCommand(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFollowTheQueenConfig())
	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.BigBet = 20
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bb ok")
	assert.Equal(t, "bb ok", c.Exec("bigbet 20"))
}

func TestFollowTheQueenCuiController_BigBet_NoArgs(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	result := c.Exec("bb")
	assert.Contains(t, result, "ビッグベット")
}

func TestFollowTheQueenCuiController_BigBet_InvalidValue(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Contains(t, c.Exec("bb abc"), "ビッグベット")
}

// --- level hand ---

func TestFollowTheQueenCuiController_LevelHand_Valid(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFollowTheQueenConfig())
	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.AnteLevelHands = 5
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("lh ok")
	assert.Equal(t, "lh ok", c.Exec("lh 5"))
}

func TestFollowTheQueenCuiController_LevelHand_LongCommand(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFollowTheQueenConfig())
	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.AnteLevelHands = 5
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("lh ok")
	assert.Equal(t, "lh ok", c.Exec("levelhand 5"))
}

func TestFollowTheQueenCuiController_LevelHand_NoArgs(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Contains(t, c.Exec("lh"), "ハンド数が必要")
}

func TestFollowTheQueenCuiController_LevelHand_InvalidValue(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Contains(t, c.Exec("lh abc"), "abc")
}

// --- table size ---

func TestFollowTheQueenCuiController_TableSize_Valid(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFollowTheQueenConfig())
	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.TableSize = 4
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("ts ok")
	assert.Equal(t, "ts ok", c.Exec("ts 4"))
}

func TestFollowTheQueenCuiController_TableSize_LongCommand(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultFollowTheQueenConfig())
	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.TableSize = 6
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("ts ok")
	assert.Equal(t, "ts ok", c.Exec("tablesize 6"))
}

func TestFollowTheQueenCuiController_TableSize_NoArgs(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Contains(t, c.Exec("ts"), "テーブルサイズ")
}

func TestFollowTheQueenCuiController_TableSize_InvalidValue(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Contains(t, c.Exec("ts abc"), "abc")
}

// --- rebuy ---

func TestFollowTheQueenCuiController_Rebuy(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("Rebuy").Return("rebuy ok")
	assert.Equal(t, "rebuy ok", c.Exec("rb"))
	assert.Equal(t, "rebuy ok", c.Exec("rebuy"))
}

func TestFollowTheQueenCuiController_SkipRebuy(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("SkipRebuy").Return("skiprebuy ok")
	assert.Equal(t, "skiprebuy ok", c.Exec("sr"))
	assert.Equal(t, "skiprebuy ok", c.Exec("skiprebuy"))
}

// --- addon ---

func TestFollowTheQueenCuiController_Addon(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("Addon").Return("addon ok")
	assert.Equal(t, "addon ok", c.Exec("ad"))
	assert.Equal(t, "addon ok", c.Exec("addon"))
}

func TestFollowTheQueenCuiController_SkipAddon(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("SkipAddon").Return("skipaddon ok")
	assert.Equal(t, "skipaddon ok", c.Exec("sa"))
	assert.Equal(t, "skipaddon ok", c.Exec("skipaddon"))
}

// --- muck / show ---

func TestFollowTheQueenCuiController_Muck(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("Muck").Return("muck ok")
	assert.Equal(t, "muck ok", c.Exec("m"))
	assert.Equal(t, "muck ok", c.Exec("muck"))
}

func TestFollowTheQueenCuiController_ShowHand(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("ShowHand").Return("show ok")
	assert.Equal(t, "show ok", c.Exec("sh"))
	assert.Equal(t, "show ok", c.Exec("show"))
}

// --- metaai ---

func TestFollowTheQueenCuiController_MetaAI_On(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	cfg := domain.DefaultFollowTheQueenConfig()
	mi.On("GetConfig").Return(cfg)
	expected := domain.DefaultFollowTheQueenConfig()
	expected.CpuMetaAI = true
	mi.On("ResetWithConfig", expected, mock.Anything).Return("mai ok")
	assert.Equal(t, "mai ok", c.Exec("mai 1"))
}

func TestFollowTheQueenCuiController_MetaAI_Off(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	cfg := domain.DefaultFollowTheQueenConfig()
	mi.On("GetConfig").Return(cfg)
	expected := domain.DefaultFollowTheQueenConfig()
	expected.CpuMetaAI = false
	mi.On("ResetWithConfig", expected, mock.Anything).Return("mai off")
	assert.Equal(t, "mai off", c.Exec("metaai 0"))
}

func TestFollowTheQueenCuiController_MetaAI_MissingArg(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Equal(t, "メタAI設定が必要です (0=オフ, 1=オン)。", c.Exec("mai"))
}

func TestFollowTheQueenCuiController_MetaAI_InvalidArg(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	assert.Equal(t, invalidArg("invalidMetaAI", "val", "abc"), c.Exec("mai abc"))
}

// --- log ---

func TestFollowTheQueenCuiController_Log(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("ActionLog").Return("log output")
	assert.Equal(t, "log output", c.Exec("log"))
	assert.Equal(t, "log output", c.Exec("l"))
}

func TestFollowTheQueenCuiController_Hint(t *testing.T) {
	mi := new(usecase.MockFollowTheQueenInteractor)
	c := NewFollowTheQueenCuiController(mi)
	mi.On("Hint").Return("hint output")
	assert.Equal(t, "hint output", c.Exec("h"))
	assert.Equal(t, "hint output", c.Exec("hint"))
}
