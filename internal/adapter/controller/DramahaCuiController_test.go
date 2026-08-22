package controller

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestDramahaCuiController_Quit(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestDramahaCuiController_Reset(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("Reset").Return("reset ok")
	assert.Equal(t, "reset ok", c.Exec("r"))
	assert.Equal(t, "reset ok", c.Exec("reset"))
}

func TestDramahaCuiController_Fold(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("Action", domain.DramahaActionFold, 0, 0).Return("fold ok")
	assert.Equal(t, "fold ok", c.Exec("f"))
	assert.Equal(t, "fold ok", c.Exec("fold"))
}

func TestDramahaCuiController_Check(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("Action", domain.DramahaActionCheck, 0, 0).Return("check ok")
	assert.Equal(t, "check ok", c.Exec("ck"))
	assert.Equal(t, "check ok", c.Exec("check"))
}

func TestDramahaCuiController_Call(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("Action", domain.DramahaActionCall, 0, 0).Return("call ok")
	assert.Equal(t, "call ok", c.Exec("c"))
	assert.Equal(t, "call ok", c.Exec("call"))
}

func TestDramahaCuiController_Bet(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("Action", domain.DramahaActionBet, 50, 0).Return("bet ok")
	assert.Equal(t, "bet ok", c.Exec("b 50"))
	assert.Equal(t, "bet ok", c.Exec("bet 50"))
}

func TestDramahaCuiController_Bet_NoAmount(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	assert.True(t, cuiutil.IsPromptRequest(c.Exec("b")))
}

func TestDramahaCuiController_Bet_InvalidAmount(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	assert.Contains(t, c.Exec("b abc"), "無効な金額です")
	assert.Contains(t, c.Exec("bet xyz"), "無効な金額です")
}

func TestDramahaCuiController_Raise(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("Action", domain.DramahaActionRaise, 30, 0).Return("raise ok")
	assert.Equal(t, "raise ok", c.Exec("ra 30"))
	assert.Equal(t, "raise ok", c.Exec("raise 30"))
}

func TestDramahaCuiController_Raise_NoAmount(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	assert.True(t, cuiutil.IsPromptRequest(c.Exec("ra")))
}

func TestDramahaCuiController_Raise_InvalidAmount(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	assert.Contains(t, c.Exec("ra abc"), "無効な金額です")
	assert.Contains(t, c.Exec("raise xyz"), "無効な金額です")
}

func TestDramahaCuiController_Bet_NonPositiveAmount(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	assert.Contains(t, c.Exec("b -50"), "無効な金額です")
	assert.Contains(t, c.Exec("bet 0"), "無効な金額です")
}

func TestDramahaCuiController_Raise_NonPositiveAmount(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	assert.Contains(t, c.Exec("ra -30"), "無効な金額です")
	assert.Contains(t, c.Exec("raise 0"), "無効な金額です")
}

func TestDramahaCuiController_AllIn(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("Action", domain.DramahaActionAllIn, 0, 0).Return("allin ok")
	assert.Equal(t, "allin ok", c.Exec("a"))
	assert.Equal(t, "allin ok", c.Exec("allin"))
}

func TestDramahaCuiController_Empty(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

func TestDramahaCuiController_Unknown(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

// --- betting limit ---

func TestDramahaCuiController_BettingLimit_Valid(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDramahaConfig())
	cfg := domain.DefaultDramahaConfig()
	cfg.BettingLimit = domain.BettingLimitPotLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bl 1"))
}

func TestDramahaCuiController_BettingLimit_LongCommand(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDramahaConfig())
	cfg := domain.DefaultDramahaConfig()
	cfg.BettingLimit = domain.BettingLimitNoLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bettinglimit 2"))
}

func TestDramahaCuiController_BettingLimit_NoArgs(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	assert.Contains(t, c.Exec("bl"), "ベッティングリミット")
}

func TestDramahaCuiController_BettingLimit_InvalidValue(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	// non-numeric: controller handles directly
	assert.Contains(t, c.Exec("bl abc"), "abc")
	// numeric out-of-range: delegated to interactor
	mi.On("GetConfig").Return(domain.DefaultDramahaConfig())
	cfg5 := domain.DefaultDramahaConfig()
	cfg5.BettingLimit = domain.BettingLimitType(5)
	mi.On("ResetWithConfig", cfg5, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("bl 5"))
	cfgNeg := domain.DefaultDramahaConfig()
	cfgNeg.BettingLimit = domain.BettingLimitType(-1)
	mi.On("ResetWithConfig", cfgNeg, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("bl -1"))
}

// --- tournament mode ---

func TestDramahaCuiController_TournamentMode_Valid(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDramahaConfig())
	cfg := domain.DefaultDramahaConfig()
	cfg.TournamentMode = true
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("tm ok")
	assert.Equal(t, "tm ok", c.Exec("tm 1"))
}

func TestDramahaCuiController_TournamentMode_LongCommand(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDramahaConfig())
	cfg := domain.DefaultDramahaConfig()
	cfg.TournamentMode = false
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("tm ok")
	assert.Equal(t, "tm ok", c.Exec("tournament 0"))
}

func TestDramahaCuiController_TournamentMode_NoArgs(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	assert.Contains(t, c.Exec("tm"), "トーナメントモード")
}

func TestDramahaCuiController_TournamentMode_InvalidValue(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	assert.Contains(t, c.Exec("tm 5"), "5")
	assert.Contains(t, c.Exec("tm abc"), "abc")
	assert.Contains(t, c.Exec("tm -1"), "-1")
}

// --- small blind ---

func TestDramahaCuiController_SmallBlind_Valid(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDramahaConfig())
	cfg := domain.DefaultDramahaConfig()
	cfg.SmallBlind = 3
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("sb ok")
	assert.Equal(t, "sb ok", c.Exec("sb 3"))
}

func TestDramahaCuiController_SmallBlind_LongCommand(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDramahaConfig())
	cfg := domain.DefaultDramahaConfig()
	cfg.SmallBlind = 3
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("sb ok")
	assert.Equal(t, "sb ok", c.Exec("smallblind 3"))
}

func TestDramahaCuiController_SmallBlind_NoArgs(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	assert.Contains(t, c.Exec("sb"), "スモールブラインド")
}

func TestDramahaCuiController_SmallBlind_InvalidValue(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	// non-numeric: controller handles directly
	assert.Contains(t, c.Exec("sb abc"), "abc")
	// numeric out-of-range: delegated to interactor
	mi.On("GetConfig").Return(domain.DefaultDramahaConfig())
	cfg0 := domain.DefaultDramahaConfig()
	cfg0.SmallBlind = 0
	mi.On("ResetWithConfig", cfg0, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("sb 0"))
	cfgNeg := domain.DefaultDramahaConfig()
	cfgNeg.SmallBlind = -1
	mi.On("ResetWithConfig", cfgNeg, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("sb -1"))
}

func TestDramahaCuiController_SmallBlind_NotLessThanBigBlind(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	baseCfg := domain.DramahaConfig{SmallBlind: 5, BigBlind: 10, InitChips: 1000, BlindLevelHands: 10, BlindMultiplier: 200}
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

func TestDramahaCuiController_BigBlind_Valid(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDramahaConfig())
	cfg := domain.DefaultDramahaConfig()
	cfg.BigBlind = 20
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bb ok")
	assert.Equal(t, "bb ok", c.Exec("bb 20"))
}

func TestDramahaCuiController_BigBlind_LongCommand(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDramahaConfig())
	cfg := domain.DefaultDramahaConfig()
	cfg.BigBlind = 20
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("bb ok")
	assert.Equal(t, "bb ok", c.Exec("bigblind 20"))
}

func TestDramahaCuiController_BigBlind_NoArgs(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	assert.Contains(t, c.Exec("bb"), "ビッグブラインド")
}

func TestDramahaCuiController_BigBlind_InvalidValue(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	// non-numeric: controller handles directly
	assert.Contains(t, c.Exec("bb abc"), "abc")
	// numeric out-of-range: delegated to interactor
	mi.On("GetConfig").Return(domain.DefaultDramahaConfig())
	cfg1 := domain.DefaultDramahaConfig()
	cfg1.BigBlind = 1
	mi.On("ResetWithConfig", cfg1, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("bb 1"))
	cfgNeg := domain.DefaultDramahaConfig()
	cfgNeg.BigBlind = -1
	mi.On("ResetWithConfig", cfgNeg, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("bb -1"))
}

func TestDramahaCuiController_BigBlind_NotGreaterThanSmallBlind(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	baseCfg := domain.DramahaConfig{SmallBlind: 5, BigBlind: 10, InitChips: 1000, BlindLevelHands: 10, BlindMultiplier: 200}
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

func TestDramahaCuiController_LevelHand_Valid(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDramahaConfig())
	cfg := domain.DefaultDramahaConfig()
	cfg.BlindLevelHands = 5
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("lh ok")
	assert.Equal(t, "lh ok", c.Exec("lh 5"))
}

func TestDramahaCuiController_LevelHand_LongCommand(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDramahaConfig())
	cfg := domain.DefaultDramahaConfig()
	cfg.BlindLevelHands = 5
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("lh ok")
	assert.Equal(t, "lh ok", c.Exec("levelhand 5"))
}

func TestDramahaCuiController_LevelHand_NoArgs(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	assert.Contains(t, c.Exec("lh"), "ハンド数が必要")
}

func TestDramahaCuiController_LevelHand_InvalidValue(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	// non-numeric: controller handles directly
	assert.Contains(t, c.Exec("lh abc"), "abc")
	// numeric out-of-range: delegated to interactor
	mi.On("GetConfig").Return(domain.DefaultDramahaConfig())
	cfg0 := domain.DefaultDramahaConfig()
	cfg0.BlindLevelHands = 0
	mi.On("ResetWithConfig", cfg0, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("lh 0"))
	cfgNeg := domain.DefaultDramahaConfig()
	cfgNeg.BlindLevelHands = -1
	mi.On("ResetWithConfig", cfgNeg, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("lh -1"))
}

// --- table size ---

func TestDramahaCuiController_TableSize_Valid(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDramahaConfig())
	cfg := domain.DefaultDramahaConfig()
	cfg.TableSize = domain.HoldemTableSize4
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("ts ok")
	// 山に収まる唯一の席数。6/9 は TableSizeRejectsWhatTheDeckCannotDeal を参照。
	assert.Equal(t, "ts ok", c.Exec("ts 4"))
}

func TestDramahaCuiController_TableSize_LongCommand(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultDramahaConfig())
	cfg := domain.DefaultDramahaConfig()
	cfg.TableSize = domain.HoldemTableSize4
	mi.On("ResetWithConfig", cfg, mock.Anything).Return("ts ok")
	assert.Equal(t, "ts ok", c.Exec("tablesize 4"))
}

func TestDramahaCuiController_TableSize_NoArgs(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	assert.Contains(t, c.Exec("ts"), "テーブルサイズ")
}

func TestDramahaCuiController_TableSize_InvalidValue(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	// non-numeric: controller handles directly
	assert.Contains(t, c.Exec("ts abc"), "abc")
	// numeric out-of-range: delegated to interactor
	mi.On("GetConfig").Return(domain.DefaultDramahaConfig())
	cfg5 := domain.DefaultDramahaConfig()
	cfg5.TableSize = 5
	mi.On("ResetWithConfig", cfg5, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("ts 5"))
	cfgNeg := domain.DefaultDramahaConfig()
	cfgNeg.TableSize = -1
	mi.On("ResetWithConfig", cfgNeg, mock.Anything).Return("error from domain")
	assert.Equal(t, "error from domain", c.Exec("ts -1"))
}

// --- rebuy ---

func TestDramahaCuiController_Rebuy(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("Rebuy").Return("rebuy ok")
	assert.Equal(t, "rebuy ok", c.Exec("rb"))
	assert.Equal(t, "rebuy ok", c.Exec("rebuy"))
}

func TestDramahaCuiController_SkipRebuy(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("SkipRebuy").Return("skiprebuy ok")
	assert.Equal(t, "skiprebuy ok", c.Exec("sr"))
	assert.Equal(t, "skiprebuy ok", c.Exec("skiprebuy"))
}

// --- addon ---

func TestDramahaCuiController_Addon(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("Addon").Return("addon ok")
	assert.Equal(t, "addon ok", c.Exec("ad"))
	assert.Equal(t, "addon ok", c.Exec("addon"))
}

func TestDramahaCuiController_SkipAddon(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("SkipAddon").Return("skipaddon ok")
	assert.Equal(t, "skipaddon ok", c.Exec("sa"))
	assert.Equal(t, "skipaddon ok", c.Exec("skipaddon"))
}

// --- muck / show ---

func TestDramahaCuiController_Muck(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("Muck").Return("muck ok")
	assert.Equal(t, "muck ok", c.Exec("m"))
	assert.Equal(t, "muck ok", c.Exec("muck"))
}

func TestDramahaCuiController_ShowHand(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	mi.On("ShowHand").Return("show ok")
	assert.Equal(t, "show ok", c.Exec("sh"))
	assert.Equal(t, "show ok", c.Exec("show"))
}

// --- metaai ---

func TestDramahaCuiController_MetaAI_On(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	cfg := domain.DefaultDramahaConfig()
	mi.On("GetConfig").Return(cfg)
	expected := domain.DefaultDramahaConfig()
	expected.CpuMetaAI = true
	mi.On("ResetWithConfig", expected, mock.Anything).Return("mai ok")
	assert.Equal(t, "mai ok", c.Exec("mai 1"))
}

func TestDramahaCuiController_MetaAI_Off(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	cfg := domain.DefaultDramahaConfig()
	mi.On("GetConfig").Return(cfg)
	expected := domain.DefaultDramahaConfig()
	expected.CpuMetaAI = false
	mi.On("ResetWithConfig", expected, mock.Anything).Return("mai off")
	assert.Equal(t, "mai off", c.Exec("metaai 0"))
}

func TestDramahaCuiController_MetaAI_MissingArg(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	assert.Equal(t, "メタAI設定が必要です (0=オフ, 1=オン)。", c.Exec("mai"))
}

func TestDramahaCuiController_MetaAI_InvalidArg(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	c := NewDramahaCuiController(mi)
	assert.Equal(t, invalidArg("invalidMetaAI", "val", "abc"), c.Exec("mai abc"))
}

// --- draw ---

// TestDramahaCuiController_Draw_ConvertsOneBasedInputToZeroBased is the test
// for Dramaha's only off-by-one risk: the screen numbers the hole cards from 1
// and the domain indexes them from 0. `d 1 3` must discard the first and third
// cards, i.e. domain indices 0 and 2.
func TestDramahaCuiController_Draw_ConvertsOneBasedInputToZeroBased(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  []int
	}{
		{"d 1 3", []int{0, 2}},
		{"draw 1 3", []int{0, 2}},
		{"d 1", []int{0}},
		{"d 5", []int{4}},
		{"d 1 2 3 4 5", []int{0, 1, 2, 3, 4}},
		{"d 3 1", []int{2, 0}},
	} {
		t.Run(tc.input, func(t *testing.T) {
			mi := new(usecase.MockDramahaInteractor)
			c := NewDramahaCuiController(mi)
			mi.On("Draw", tc.want).Return("draw ok")

			assert.Equal(t, "draw ok", c.Exec(tc.input))
			mi.AssertCalled(t, "Draw", tc.want)
		})
	}
}

// TestDramahaCuiController_Draw_RejectsZeroAndNegative guards the lower edge of
// the 1-based numbering: `d 0` is not "the first card", it is a typo, and
// letting it through would silently discard index -1.
func TestDramahaCuiController_Draw_RejectsOutOfRangeInput(t *testing.T) {
	for _, input := range []string{"d 0", "d -1", "d abc", "d 1 0"} {
		t.Run(input, func(t *testing.T) {
			mi := new(usecase.MockDramahaInteractor)
			c := NewDramahaCuiController(mi)

			out := c.Exec(input)

			assert.Contains(t, out, "無効なカードインデックスです",
				"the player has to be told which value was rejected")
			mi.AssertNotCalled(t, "Draw", mock.Anything)
		})
	}
}

// TestDramahaCuiController_Draw_BareCommandStandsPat: no arguments means "keep
// all five", which is a legal move, not an error.
func TestDramahaCuiController_Draw_BareCommandStandsPat(t *testing.T) {
	for _, input := range []string{"d", "draw"} {
		mi := new(usecase.MockDramahaInteractor)
		c := NewDramahaCuiController(mi)
		mi.On("Draw", []int{}).Return("stood pat")

		assert.Equal(t, "stood pat", c.Exec(input), input)
		mi.AssertCalled(t, "Draw", []int{})
	}
}

// TestDramahaCuiController_DrawIsAdvertised keeps the help wiring honest: a
// command the parser accepts has to be listed, or players never find it.
func TestDramahaCuiController_DrawIsAdvertised(t *testing.T) {
	assert.Contains(t, dramahaArgfulCommands, "d")
	assert.Contains(t, dramahaArgfulCommands, "draw")
}

// TestDramahaCuiController_TableSizeRejectsWhatTheDeckCannotDeal は、
// 収まらない席数を頼まれたときに黙って丸めないことを見る。
//
// **丸めるだけだと、効いたのか無視されたのか画面から分からない。** しかも
// `ts 6` は「6 人卓になった」と読めてしまう。実際にはドラマハは 1 席が最悪
// 10 枚 (ホール 5 + 交換 5) 使い、ボードに 5 枚要るので 10N+5 ——
// 6-max は 65 枚で、52 枚の山に収まらない。
func TestDramahaCuiController_TableSizeRejectsWhatTheDeckCannotDeal(t *testing.T) {
	for _, lang := range []string{"ja", "en"} {
		t.Run(lang, func(t *testing.T) {
			defer i18n.SetLang(i18n.Lang())
			i18n.SetLang(lang)

			for _, size := range []int{6, 9} {
				mi := new(usecase.MockDramahaInteractor)
				c := NewDramahaCuiController(mi)

				out := c.Exec("ts " + strconv.Itoa(size))

				assert.Contains(t, out, strconv.Itoa(size),
					"断るなら、頼まれた数を見せる")
				assert.NotContains(t, out, "{{", "プレースホルダが素通りしている")
				mi.AssertNotCalled(t, "ResetWithConfig", mock.Anything, mock.Anything)
			}
		})
	}
}

// TestDramahaCuiController_TableSizeAcceptsFourMax は上の負のコントロール。
// **4 まで弾いていたら、上のテストは「常に断る」実装でも通る。**
func TestDramahaCuiController_TableSizeAcceptsFourMax(t *testing.T) {
	mi := new(usecase.MockDramahaInteractor)
	mi.On("GetConfig").Return(domain.DefaultDramahaConfig())
	mi.On("ResetWithConfig", mock.Anything, mock.Anything).Return("ok")
	c := NewDramahaCuiController(mi)

	out := c.Exec("ts 4")

	assert.Equal(t, "ok", out)
	mi.AssertCalled(t, "ResetWithConfig", mock.Anything, mock.Anything)
}
