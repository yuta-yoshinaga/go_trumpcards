package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
	mi.On("Action", domain.HoldemActionFold, 0).Return("fold ok")
	assert.Equal(t, "fold ok", c.Exec("f"))
	assert.Equal(t, "fold ok", c.Exec("fold"))
}

func TestHoldemCuiController_Check(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Action", domain.HoldemActionCheck, 0).Return("check ok")
	assert.Equal(t, "check ok", c.Exec("ck"))
	assert.Equal(t, "check ok", c.Exec("check"))
}

func TestHoldemCuiController_Call(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Action", domain.HoldemActionCall, 0).Return("call ok")
	assert.Equal(t, "call ok", c.Exec("c"))
	assert.Equal(t, "call ok", c.Exec("call"))
}

func TestHoldemCuiController_Bet(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Action", domain.HoldemActionBet, 50).Return("bet ok")
	assert.Equal(t, "bet ok", c.Exec("b 50"))
	assert.Equal(t, "bet ok", c.Exec("bet 50"))
}

func TestHoldemCuiController_Bet_NoAmount(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("b"), "金額の指定が必要です")
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
	mi.On("Action", domain.HoldemActionRaise, 30).Return("raise ok")
	assert.Equal(t, "raise ok", c.Exec("ra 30"))
	assert.Equal(t, "raise ok", c.Exec("raise 30"))
}

func TestHoldemCuiController_Raise_NoAmount(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("ra"), "金額の指定が必要です")
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
	mi.On("Action", domain.HoldemActionAllIn, 0).Return("allin ok")
	assert.Equal(t, "allin ok", c.Exec("a"))
	assert.Equal(t, "allin ok", c.Exec("allin"))
}

func TestHoldemCuiController_Empty(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec(""), "コマンドが不明です")
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
	mi.On("ResetWithConfig", cfg).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bl 1"))
}

func TestHoldemCuiController_BettingLimit_LongCommand(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.BettingLimit = domain.BettingLimitNoLimit
	mi.On("ResetWithConfig", cfg).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bettinglimit 2"))
}

func TestHoldemCuiController_BettingLimit_NoArgs(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("bl"), "Betting limit type is required")
}

func TestHoldemCuiController_BettingLimit_InvalidValue(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("bl 5"), "Invalid betting limit: 5")
	assert.Contains(t, c.Exec("bl abc"), "Invalid betting limit: abc")
	assert.Contains(t, c.Exec("bl -1"), "Invalid betting limit: -1")
}

// --- tournament mode ---

func TestHoldemCuiController_TournamentMode_Valid(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.TournamentMode = true
	mi.On("ResetWithConfig", cfg).Return("tm ok")
	assert.Equal(t, "tm ok", c.Exec("tm 1"))
}

func TestHoldemCuiController_TournamentMode_LongCommand(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.TournamentMode = false
	mi.On("ResetWithConfig", cfg).Return("tm ok")
	assert.Equal(t, "tm ok", c.Exec("tournament 0"))
}

func TestHoldemCuiController_TournamentMode_NoArgs(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("tm"), "Tournament mode is required")
}

func TestHoldemCuiController_TournamentMode_InvalidValue(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("tm 5"), "Invalid tournament mode: 5")
	assert.Contains(t, c.Exec("tm abc"), "Invalid tournament mode: abc")
	assert.Contains(t, c.Exec("tm -1"), "Invalid tournament mode: -1")
}

// --- small blind ---

func TestHoldemCuiController_SmallBlind_Valid(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.SmallBlind = 3
	mi.On("ResetWithConfig", cfg).Return("sb ok")
	assert.Equal(t, "sb ok", c.Exec("sb 3"))
}

func TestHoldemCuiController_SmallBlind_LongCommand(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.SmallBlind = 3
	mi.On("ResetWithConfig", cfg).Return("sb ok")
	assert.Equal(t, "sb ok", c.Exec("smallblind 3"))
}

func TestHoldemCuiController_SmallBlind_NoArgs(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("sb"), "Small blind amount is required")
}

func TestHoldemCuiController_SmallBlind_InvalidValue(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("sb 0"), "Invalid small blind: 0")
	assert.Contains(t, c.Exec("sb abc"), "Invalid small blind: abc")
	assert.Contains(t, c.Exec("sb -1"), "Invalid small blind: -1")
}

func TestHoldemCuiController_SmallBlind_NotLessThanBigBlind(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.HoldemConfig{SmallBlind: 5, BigBlind: 10, InitChips: 1000, BlindLevelHands: 10, BlindMultiplier: 200})
	assert.Contains(t, c.Exec("sb 10"), "Small blind must be less than big blind")
	assert.Contains(t, c.Exec("sb 15"), "Small blind must be less than big blind")
}

// --- big blind ---

func TestHoldemCuiController_BigBlind_Valid(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.BigBlind = 20
	mi.On("ResetWithConfig", cfg).Return("bb ok")
	assert.Equal(t, "bb ok", c.Exec("bb 20"))
}

func TestHoldemCuiController_BigBlind_LongCommand(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.BigBlind = 20
	mi.On("ResetWithConfig", cfg).Return("bb ok")
	assert.Equal(t, "bb ok", c.Exec("bigblind 20"))
}

func TestHoldemCuiController_BigBlind_NoArgs(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("bb"), "Big blind amount is required")
}

func TestHoldemCuiController_BigBlind_InvalidValue(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("bb 1"), "Invalid big blind: 1")
	assert.Contains(t, c.Exec("bb abc"), "Invalid big blind: abc")
	assert.Contains(t, c.Exec("bb -1"), "Invalid big blind: -1")
}

func TestHoldemCuiController_BigBlind_NotGreaterThanSmallBlind(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.HoldemConfig{SmallBlind: 5, BigBlind: 10, InitChips: 1000, BlindLevelHands: 10, BlindMultiplier: 200})
	assert.Contains(t, c.Exec("bb 5"), "Big blind must be greater than small blind")
	assert.Contains(t, c.Exec("bb 3"), "Big blind must be greater than small blind")
}

// --- level-up hands ---

func TestHoldemCuiController_LevelHand_Valid(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.BlindLevelHands = 5
	mi.On("ResetWithConfig", cfg).Return("lh ok")
	assert.Equal(t, "lh ok", c.Exec("lh 5"))
}

func TestHoldemCuiController_LevelHand_LongCommand(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultHoldemConfig())
	cfg := domain.DefaultHoldemConfig()
	cfg.BlindLevelHands = 5
	mi.On("ResetWithConfig", cfg).Return("lh ok")
	assert.Equal(t, "lh ok", c.Exec("levelhand 5"))
}

func TestHoldemCuiController_LevelHand_NoArgs(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("lh"), "Level-up hands is required")
}

func TestHoldemCuiController_LevelHand_InvalidValue(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("lh 0"), "Invalid level-up hands: 0")
	assert.Contains(t, c.Exec("lh abc"), "Invalid level-up hands: abc")
	assert.Contains(t, c.Exec("lh -1"), "Invalid level-up hands: -1")
}
