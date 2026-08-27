package controller_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOldMaidCuiController_Method(t *testing.T) {
	mockOutput := "==========\nOld Maid (ババ抜き)\n==========\n[You]: 0枚\n\nCPU 1: 0枚\nCPU 2: 0枚\nCPU 3: 0枚\n----------\n手番: あなた\n==========\n"
	omiMock := new(usecase.MockOldMaidInteractor)
	omiMock.On("GetConfig").Return(domain.DefaultOldMaidConfig())
	omiMock.On("Reset", mock.Anything, mock.Anything).Return(mockOutput)
	omiMock.On("Draw", -1).Return(mockOutput)
	omiMock.On("Draw", 0).Return(mockOutput)
	omiMock.On("Draw", 2).Return(mockOutput)
	omiMock.On("Shuffle").Return(mockOutput)

	tomc := controller.NewOldMaidCuiController(omiMock)

	t.Run("success Exec q", func(t *testing.T) {
		assert.Equal(t, "bye.", tomc.Exec("q"))
	})
	t.Run("success Exec quit", func(t *testing.T) {
		assert.Equal(t, "bye.", tomc.Exec("quit"))
	})
	t.Run("success Exec r preserves config", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("r"))
		omiMock.AssertCalled(t, "GetConfig")
	})
	t.Run("success Exec reset preserves config", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("reset"))
		omiMock.AssertCalled(t, "GetConfig")
	})
	t.Run("success Exec d", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("d"))
	})
	t.Run("success Exec draw", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("draw"))
	})
	t.Run("success Exec d 0", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("d 0"))
	})
	t.Run("success Exec draw 2", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("draw 2"))
	})
	// **打ち間違いを -1 (ランダム) として実行しない。** 30 配り中 17 配りで、
	// プレイヤーが選んでいない札を引いていた (issue #5390)。
	t.Run("draw refuses an unparseable index", func(t *testing.T) {
		refuseMock := new(usecase.MockOldMaidInteractor)
		refuseMock.On("Draw", mock.Anything).Return(mockOutput)
		c := controller.NewOldMaidCuiController(refuseMock)

		assert.Contains(t, c.Exec("d zz"), msgInvalidCardIndexPrefix())

		refuseMock.AssertNotCalled(t, "Draw", mock.Anything)
	})
	t.Run("success Exec s (shuffle)", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("s"))
		omiMock.AssertCalled(t, "Shuffle")
	})
	t.Run("success Exec shuffle", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("shuffle"))
		omiMock.AssertCalled(t, "Shuffle")
	})
	t.Run("success Exec other", func(t *testing.T) {
		assert.Equal(t, i18n.ErrorPrefix+"コマンドが不明です: other", tomc.Exec("other"))
	})
	t.Run("success Exec empty", func(t *testing.T) {
		assert.Equal(t, "'help' でコマンド一覧を表示します。", tomc.Exec(""))
	})
}

// --- reorder ---

func TestOldMaidCuiController_Reorder_Valid(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	mi.On("Reorder", []int{2, 0, 1}).Return("reorder ok")
	assert.Equal(t, "reorder ok", c.Exec("ro 2 0 1"))
}

func TestOldMaidCuiController_Reorder_LongCommand(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	mi.On("Reorder", []int{1, 0}).Return("reorder ok")
	assert.Equal(t, "reorder ok", c.Exec("reorder 1 0"))
}

func TestOldMaidCuiController_Reorder_NoArgs(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	mi.On("Reorder", []int{}).Return("reorder empty")
	assert.Equal(t, "reorder empty", c.Exec("ro"))
}

func TestOldMaidCuiController_Reorder_InvalidIndex_NonNumeric(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	mi.On("Reorder", []int{2}).Return("reorder ok")
	result := c.Exec("ro abc 2")
	assert.Contains(t, result, "abc")
	mi.AssertNotCalled(t, "Reorder", mock.Anything)
}

// --- set mode ---

func TestOldMaidCuiController_SetMode_Valid(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultOldMaidConfig())
	cfg := domain.DefaultOldMaidConfig()
	cfg.Mode = domain.OldMaidModeJijiNuki
	mi.On("Reset", cfg, mock.Anything).Return("sm ok")
	assert.Equal(t, "sm ok", c.Exec("sm 1"))
}

func TestOldMaidCuiController_SetMode_LongCommand(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultOldMaidConfig())
	cfg := domain.DefaultOldMaidConfig()
	cfg.Mode = domain.OldMaidModeNormal
	mi.On("Reset", cfg, mock.Anything).Return("sm ok")
	assert.Equal(t, "sm ok", c.Exec("setmode 0"))
}

func TestOldMaidCuiController_SetMode_NoArgs(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	assert.Contains(t, c.Exec("sm"), msgStem("gameModeRequired0Normal1Jijinuki"))
}

func TestOldMaidCuiController_SetMode_InvalidValue(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	// non-numeric: controller catches
	assert.Contains(t, c.Exec("sm abc"), msgKey("invalidGameMode01", "val", "abc"))
	// numeric out-of-range: controller catches via ParseIntArg bounds
	assert.Equal(t, msgKey("invalidGameMode01", "val", "2"), c.Exec("sm 2"))
	assert.Equal(t, msgKey("invalidGameMode01", "val", "-1"), c.Exec("sm -1"))
}

// --- set placement strategy ---

func TestOldMaidCuiController_SetPlacementStrategy_Valid(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultOldMaidConfig())
	cfg := domain.DefaultOldMaidConfig()
	cfg.CpuPlacementStrategy = true
	mi.On("Reset", cfg, mock.Anything).Return("sps ok")
	assert.Equal(t, "sps ok", c.Exec("sps 1"))
}

func TestOldMaidCuiController_SetPlacementStrategy_LongCommand(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultOldMaidConfig())
	cfg := domain.DefaultOldMaidConfig()
	cfg.CpuPlacementStrategy = false
	mi.On("Reset", cfg, mock.Anything).Return("sps ok")
	assert.Equal(t, "sps ok", c.Exec("setplacementstrategy 0"))
}

func TestOldMaidCuiController_SetPlacementStrategy_NoArgs(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	assert.Contains(t, c.Exec("sps"), msgStem("cpuPlacementStrategyFlagRequired0Off1On"))
}

func TestOldMaidCuiController_SetPlacementStrategy_InvalidValue(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	assert.Contains(t, c.Exec("sps 2"), msgKey("invalidCpuPlacementStrategyFlag01", "val", "2"))
	assert.Contains(t, c.Exec("sps abc"), msgKey("invalidCpuPlacementStrategyFlag01", "val", "abc"))
	assert.Contains(t, c.Exec("sps -1"), msgKey("invalidCpuPlacementStrategyFlag01", "val", "-1"))
}

// --- set memory AI ---

func TestOldMaidCuiController_SetMemoryAI_Valid(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultOldMaidConfig())
	cfg := domain.DefaultOldMaidConfig()
	cfg.CpuMemoryAI = true
	mi.On("Reset", cfg, mock.Anything).Return("sma ok")
	assert.Equal(t, "sma ok", c.Exec("sma 1"))
}

func TestOldMaidCuiController_SetMemoryAI_LongCommand(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultOldMaidConfig())
	cfg := domain.DefaultOldMaidConfig()
	cfg.CpuMemoryAI = false
	mi.On("Reset", cfg, mock.Anything).Return("sma ok")
	assert.Equal(t, "sma ok", c.Exec("setmemoryai 0"))
}

func TestOldMaidCuiController_SetMemoryAI_NoArgs(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	assert.Contains(t, c.Exec("sma"), msgStem("cpuMemoryAiFlagRequired0Off1On"))
}

func TestOldMaidCuiController_SetMemoryAI_InvalidValue(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	assert.Contains(t, c.Exec("sma 2"), msgKey("invalidCpuMemoryAiFlag01", "val", "2"))
	assert.Contains(t, c.Exec("sma abc"), msgKey("invalidCpuMemoryAiFlag01", "val", "abc"))
	assert.Contains(t, c.Exec("sma -1"), msgKey("invalidCpuMemoryAiFlag01", "val", "-1"))
}

// --- set meta-AI ---

func TestOldMaidCuiController_SetMetaAI_Valid(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultOldMaidConfig())
	cfg := domain.DefaultOldMaidConfig()
	cfg.CpuMetaAI = true
	mi.On("Reset", cfg, mock.Anything).Return("smai ok")
	assert.Equal(t, "smai ok", c.Exec("smai 1"))
}

func TestOldMaidCuiController_SetMetaAI_LongCommand(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	mi.On("GetConfig").Return(domain.DefaultOldMaidConfig())
	cfg := domain.DefaultOldMaidConfig()
	cfg.CpuMetaAI = false
	mi.On("Reset", cfg, mock.Anything).Return("smai ok")
	assert.Equal(t, "smai ok", c.Exec("smetaai 0"))
}

func TestOldMaidCuiController_SetMetaAI_NoArgs(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	assert.Contains(t, c.Exec("smai"), msgStem("metaAiFlagRequired0Off1On"))
}

func TestOldMaidCuiController_SetMetaAI_InvalidValue(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	assert.Contains(t, c.Exec("smai 2"), msgKey("invalidMetaAiFlag01", "val", "2"))
	assert.Contains(t, c.Exec("smetaai abc"), msgKey("invalidMetaAiFlag01", "val", "abc"))
	assert.Contains(t, c.Exec("smai -1"), msgKey("invalidMetaAiFlag01", "val", "-1"))
}

// --- reset profile ---

func TestOldMaidCuiController_ResetProfile(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	mi.On("ResetProfile").Return("profile reset ok")
	assert.Equal(t, "profile reset ok", c.Exec("rp"))
}

func TestOldMaidCuiController_ResetProfile_LongCommand(t *testing.T) {
	mi := new(usecase.MockOldMaidInteractor)
	c := controller.NewOldMaidCuiController(mi)
	mi.On("ResetProfile").Return("profile reset ok")
	assert.Equal(t, "profile reset ok", c.Exec("resetprofile"))
}

// CpuHesitationEnabled was settable from the web dialog but from nowhere in the
// CUI, so a CUI session was pinned to the default forever.
func TestOldMaidCuiController_SetHesitation(t *testing.T) {
	t.Run("turns the delay on", func(t *testing.T) {
		mi := new(usecase.MockOldMaidInteractor)
		c := controller.NewOldMaidCuiController(mi)
		mi.On("GetConfig").Return(domain.DefaultOldMaidConfig())
		cfg := domain.DefaultOldMaidConfig()
		cfg.CpuHesitationEnabled = true
		mi.On("Reset", cfg, mock.Anything).Return("sh on")
		assert.Equal(t, "sh on", c.Exec("sh 1"))
	})

	t.Run("turns the delay off through the long form", func(t *testing.T) {
		mi := new(usecase.MockOldMaidInteractor)
		c := controller.NewOldMaidCuiController(mi)
		mi.On("GetConfig").Return(domain.DefaultOldMaidConfig())
		cfg := domain.DefaultOldMaidConfig()
		cfg.CpuHesitationEnabled = false
		mi.On("Reset", cfg, mock.Anything).Return("sh off")
		assert.Equal(t, "sh off", c.Exec("sethesitation 0"))
	})

	t.Run("refuses a missing or out-of-range flag", func(t *testing.T) {
		mi := new(usecase.MockOldMaidInteractor)
		c := controller.NewOldMaidCuiController(mi)
		assert.Contains(t, c.Exec("sh"), msgStem("hesitationFlagRequired0Off1On"))
		assert.Contains(t, c.Exec("sh 2"), msgKey("invalidHesitationFlag01", "val", "2"))
		assert.Contains(t, c.Exec("sh abc"), msgKey("invalidHesitationFlag01", "val", "abc"))
	})
}
