package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestSevenCardStudCuiController_Quit(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestSevenCardStudCuiController_Reset(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("Reset").Return("reset ok")
	assert.Equal(t, "reset ok", c.Exec("r"))
	assert.Equal(t, "reset ok", c.Exec("reset"))
}

func TestSevenCardStudCuiController_Fold(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("Action", domain.SevenCardStudActionFold, 0, 0).Return("fold ok")
	assert.Equal(t, "fold ok", c.Exec("f"))
	assert.Equal(t, "fold ok", c.Exec("fold"))
}

func TestSevenCardStudCuiController_Check(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("Action", domain.SevenCardStudActionCheck, 0, 0).Return("check ok")
	assert.Equal(t, "check ok", c.Exec("ck"))
	assert.Equal(t, "check ok", c.Exec("check"))
}

func TestSevenCardStudCuiController_Call(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("Action", domain.SevenCardStudActionCall, 0, 0).Return("call ok")
	assert.Equal(t, "call ok", c.Exec("c"))
	assert.Equal(t, "call ok", c.Exec("call"))
}

func TestSevenCardStudCuiController_Bet(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("Action", domain.SevenCardStudActionBet, 50, 0).Return("bet ok")
	assert.Equal(t, "bet ok", c.Exec("b 50"))
	assert.Equal(t, "bet ok", c.Exec("bet 50"))
}

func TestSevenCardStudCuiController_Bet_NoAmount(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	result := c.Exec("b")
	assert.True(t, cuiutil.IsPromptRequest(result))
	_, tmpl := cuiutil.ParsePromptRequest(result)
	assert.Equal(t, "b {0}", tmpl)
}

func TestSevenCardStudCuiController_Bet_InvalidAmount(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Contains(t, c.Exec("b abc"), "無効な金額です")
	assert.Contains(t, c.Exec("bet xyz"), "無効な金額です")
}

func TestSevenCardStudCuiController_Raise(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("Action", domain.SevenCardStudActionRaise, 30, 0).Return("raise ok")
	assert.Equal(t, "raise ok", c.Exec("ra 30"))
	assert.Equal(t, "raise ok", c.Exec("raise 30"))
}

func TestSevenCardStudCuiController_Raise_NoAmount(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	result := c.Exec("ra")
	assert.True(t, cuiutil.IsPromptRequest(result))
	_, tmpl := cuiutil.ParsePromptRequest(result)
	assert.Equal(t, "ra {0}", tmpl)
}

func TestSevenCardStudCuiController_Raise_InvalidAmount(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Contains(t, c.Exec("ra abc"), "無効な金額です")
	assert.Contains(t, c.Exec("raise xyz"), "無効な金額です")
}

func TestSevenCardStudCuiController_Bet_NonPositiveAmount(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Contains(t, c.Exec("b -50"), "無効な金額です")
	assert.Contains(t, c.Exec("bet 0"), "無効な金額です")
}

func TestSevenCardStudCuiController_Raise_NonPositiveAmount(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Contains(t, c.Exec("ra -30"), "無効な金額です")
	assert.Contains(t, c.Exec("raise 0"), "無効な金額です")
}

func TestSevenCardStudCuiController_AllIn(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("Action", domain.SevenCardStudActionAllIn, 0, 0).Return("allin ok")
	assert.Equal(t, "allin ok", c.Exec("a"))
	assert.Equal(t, "allin ok", c.Exec("allin"))
}

func TestSevenCardStudCuiController_Empty(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

func TestSevenCardStudCuiController_Unknown(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

// --- betting limit ---

func TestSevenCardStudCuiController_BettingLimit_Valid(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultSevenCardStudConfig())
	cfg := domain.DefaultSevenCardStudConfig()
	cfg.BettingLimit = domain.BettingLimitPotLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bl 1"))
}

func TestSevenCardStudCuiController_BettingLimit_LongCommand(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultSevenCardStudConfig())
	cfg := domain.DefaultSevenCardStudConfig()
	cfg.BettingLimit = domain.BettingLimitNoLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bettinglimit 2"))
}

func TestSevenCardStudCuiController_BettingLimit_NoArgs(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Contains(t, c.Exec("bl"), "ベッティングリミット")
}

func TestSevenCardStudCuiController_BettingLimit_InvalidValue(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Contains(t, c.Exec("bl abc"), "abc")
}

// --- tournament mode ---

func TestSevenCardStudCuiController_TournamentMode_Valid(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultSevenCardStudConfig())
	cfg := domain.DefaultSevenCardStudConfig()
	cfg.TournamentMode = true
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("tm ok")
	assert.Equal(t, "tm ok", c.Exec("tm 1"))
}

func TestSevenCardStudCuiController_TournamentMode_LongCommand(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultSevenCardStudConfig())
	cfg := domain.DefaultSevenCardStudConfig()
	cfg.TournamentMode = false
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("tm ok")
	assert.Equal(t, "tm ok", c.Exec("tournament 0"))
}

func TestSevenCardStudCuiController_TournamentMode_NoArgs(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Contains(t, c.Exec("tm"), "トーナメントモード")
}

func TestSevenCardStudCuiController_TournamentMode_InvalidValue(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Contains(t, c.Exec("tm 5"), "5")
	assert.Contains(t, c.Exec("tm abc"), "abc")
	assert.Contains(t, c.Exec("tm -1"), "-1")
}

// --- ante ---

func TestSevenCardStudCuiController_Ante_Valid(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultSevenCardStudConfig())
	cfg := domain.DefaultSevenCardStudConfig()
	cfg.Ante = 3
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("ante ok")
	assert.Equal(t, "ante ok", c.Exec("ante 3"))
}

func TestSevenCardStudCuiController_Ante_NoArgs(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	result := c.Exec("ante")
	assert.Contains(t, result, "アンティ")
}

func TestSevenCardStudCuiController_Ante_InvalidValue(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Contains(t, c.Exec("ante abc"), "アンティ")
}

// --- bring-in ---

func TestSevenCardStudCuiController_BringIn_Valid(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultSevenCardStudConfig())
	cfg := domain.DefaultSevenCardStudConfig()
	cfg.BringIn = 5
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bi ok")
	assert.Equal(t, "bi ok", c.Exec("bi 5"))
}

func TestSevenCardStudCuiController_BringIn_LongCommand(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultSevenCardStudConfig())
	cfg := domain.DefaultSevenCardStudConfig()
	cfg.BringIn = 5
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bi ok")
	assert.Equal(t, "bi ok", c.Exec("bringin 5"))
}

func TestSevenCardStudCuiController_BringIn_NoArgs(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	result := c.Exec("bi")
	assert.Contains(t, result, "ブリングイン")
}

func TestSevenCardStudCuiController_BringIn_InvalidValue(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Contains(t, c.Exec("bi abc"), "ブリングイン")
}

// --- small bet ---

func TestSevenCardStudCuiController_SmallBet_Valid(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultSevenCardStudConfig())
	cfg := domain.DefaultSevenCardStudConfig()
	cfg.SmallBet = 10
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("sb ok")
	assert.Equal(t, "sb ok", c.Exec("sb 10"))
}

func TestSevenCardStudCuiController_SmallBet_LongCommand(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultSevenCardStudConfig())
	cfg := domain.DefaultSevenCardStudConfig()
	cfg.SmallBet = 10
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("sb ok")
	assert.Equal(t, "sb ok", c.Exec("smallbet 10"))
}

func TestSevenCardStudCuiController_SmallBet_NoArgs(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	result := c.Exec("sb")
	assert.Contains(t, result, "スモールベット")
}

func TestSevenCardStudCuiController_SmallBet_InvalidValue(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Contains(t, c.Exec("sb abc"), "スモールベット")
}

// --- big bet ---

func TestSevenCardStudCuiController_BigBet_Valid(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultSevenCardStudConfig())
	cfg := domain.DefaultSevenCardStudConfig()
	cfg.BigBet = 20
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bb ok")
	assert.Equal(t, "bb ok", c.Exec("bb 20"))
}

func TestSevenCardStudCuiController_BigBet_LongCommand(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultSevenCardStudConfig())
	cfg := domain.DefaultSevenCardStudConfig()
	cfg.BigBet = 20
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bb ok")
	assert.Equal(t, "bb ok", c.Exec("bigbet 20"))
}

func TestSevenCardStudCuiController_BigBet_NoArgs(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	result := c.Exec("bb")
	assert.Contains(t, result, "ビッグベット")
}

func TestSevenCardStudCuiController_BigBet_InvalidValue(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Contains(t, c.Exec("bb abc"), "ビッグベット")
}

// --- level hand ---

func TestSevenCardStudCuiController_LevelHand_Valid(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultSevenCardStudConfig())
	cfg := domain.DefaultSevenCardStudConfig()
	cfg.AnteLevelHands = 5
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("lh ok")
	assert.Equal(t, "lh ok", c.Exec("lh 5"))
}

func TestSevenCardStudCuiController_LevelHand_LongCommand(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultSevenCardStudConfig())
	cfg := domain.DefaultSevenCardStudConfig()
	cfg.AnteLevelHands = 5
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("lh ok")
	assert.Equal(t, "lh ok", c.Exec("levelhand 5"))
}

func TestSevenCardStudCuiController_LevelHand_NoArgs(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Contains(t, c.Exec("lh"), "ハンド数が必要")
}

func TestSevenCardStudCuiController_LevelHand_InvalidValue(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Contains(t, c.Exec("lh abc"), "abc")
}

// --- table size ---

func TestSevenCardStudCuiController_TableSize_Valid(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultSevenCardStudConfig())
	cfg := domain.DefaultSevenCardStudConfig()
	cfg.TableSize = 4
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("ts ok")
	assert.Equal(t, "ts ok", c.Exec("ts 4"))
}

func TestSevenCardStudCuiController_TableSize_LongCommand(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultSevenCardStudConfig())
	cfg := domain.DefaultSevenCardStudConfig()
	cfg.TableSize = 6
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("ts ok")
	assert.Equal(t, "ts ok", c.Exec("tablesize 6"))
}

func TestSevenCardStudCuiController_TableSize_NoArgs(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Contains(t, c.Exec("ts"), "テーブルサイズ")
}

func TestSevenCardStudCuiController_TableSize_InvalidValue(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Contains(t, c.Exec("ts abc"), "abc")
}

// --- rebuy ---

func TestSevenCardStudCuiController_Rebuy(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("Rebuy").Return("rebuy ok")
	assert.Equal(t, "rebuy ok", c.Exec("rb"))
	assert.Equal(t, "rebuy ok", c.Exec("rebuy"))
}

func TestSevenCardStudCuiController_SkipRebuy(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("SkipRebuy").Return("skiprebuy ok")
	assert.Equal(t, "skiprebuy ok", c.Exec("sr"))
	assert.Equal(t, "skiprebuy ok", c.Exec("skiprebuy"))
}

// --- addon ---

func TestSevenCardStudCuiController_Addon(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("Addon").Return("addon ok")
	assert.Equal(t, "addon ok", c.Exec("ad"))
	assert.Equal(t, "addon ok", c.Exec("addon"))
}

func TestSevenCardStudCuiController_SkipAddon(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("SkipAddon").Return("skipaddon ok")
	assert.Equal(t, "skipaddon ok", c.Exec("sa"))
	assert.Equal(t, "skipaddon ok", c.Exec("skipaddon"))
}

// --- muck / show ---

func TestSevenCardStudCuiController_Muck(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("Muck").Return("muck ok")
	assert.Equal(t, "muck ok", c.Exec("m"))
	assert.Equal(t, "muck ok", c.Exec("muck"))
}

func TestSevenCardStudCuiController_ShowHand(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("ShowHand").Return("show ok")
	assert.Equal(t, "show ok", c.Exec("sh"))
	assert.Equal(t, "show ok", c.Exec("show"))
}

// --- metaai ---

func TestSevenCardStudCuiController_MetaAI_On(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	cfg := domain.DefaultSevenCardStudConfig()
	mi.On("GetConfig").Return(cfg)
	expected := domain.DefaultSevenCardStudConfig()
	expected.CpuMetaAI = true
	mi.On("ResetWithConfig", expected, mock.Anything).Return("mai ok")
	assert.Equal(t, "mai ok", c.Exec("mai 1"))
}

func TestSevenCardStudCuiController_MetaAI_Off(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	cfg := domain.DefaultSevenCardStudConfig()
	mi.On("GetConfig").Return(cfg)
	expected := domain.DefaultSevenCardStudConfig()
	expected.CpuMetaAI = false
	mi.On("ResetWithConfig", expected, mock.Anything).Return("mai off")
	assert.Equal(t, "mai off", c.Exec("metaai 0"))
}

func TestSevenCardStudCuiController_MetaAI_MissingArg(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Equal(t, "メタAI設定が必要です (0=オフ, 1=オン)。", c.Exec("mai"))
}

func TestSevenCardStudCuiController_MetaAI_InvalidArg(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	assert.Equal(t, "無効な値です: abc。0 または 1 を入力してください。", c.Exec("mai abc"))
}

// --- log ---

func TestSevenCardStudCuiController_Log(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("ActionLog").Return("log output")
	assert.Equal(t, "log output", c.Exec("log"))
	assert.Equal(t, "log output", c.Exec("l"))
}

func TestSevenCardStudCuiController_Hint(t *testing.T) {
	mi := new(usecase.MockSevenCardStudInteractor)
	c := NewSevenCardStudCuiController(mi)
	mi.On("Hint").Return("hint output")
	assert.Equal(t, "hint output", c.Exec("h"))
	assert.Equal(t, "hint output", c.Exec("hint"))
}
