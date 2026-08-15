package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDoubtCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockDoubtInteractor {
		m := new(mockUsecases.MockDoubtInteractor)
		m.On("GetConfig").Return(domain.DefaultDoubtConfig())
		m.On("ResetWithConfig", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
		m.On("ResolveDoubt", mock.Anything).Return(mockOutput)
		m.On("SkipDoubt").Return(mockOutput)
		return m
	}

	t.Run("quit command q", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit command quit", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset command r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("r")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultDoubtConfig(), mock.Anything)
	})

	t.Run("reset command reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("reset")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultDoubtConfig(), mock.Anything)
	})

	t.Run("play command p with value and indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("p 5 0 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", []int{0, 2}, 5, 0)
	})

	t.Run("play command play with value and indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("play 3 0 1 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", []int{0, 1, 2}, 3, 0)
	})

	t.Run("play command p with no args uses 0 value and empty indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("p")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", []int{}, 0, 0)
	})

	t.Run("play command p with only value and no card indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("p 7")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", []int{}, 7, 0)
	})

	t.Run("doubt command d with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("d 1 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResolveDoubt", []int{1, 2})
	})

	t.Run("doubt command doubt with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("doubt 0")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResolveDoubt", []int{0})
	})

	t.Run("doubt command d with no indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("d")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResolveDoubt", []int{})
	})

	t.Run("skip command s", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("s")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "SkipDoubt")
	})

	t.Run("skip command skip", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("skip")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "SkipDoubt")
	})

	// setwindow tests
	t.Run("setwindow sw valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("sw 30")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultDoubtConfig()
		expected.DoubtWindowSec = 30
		m.AssertCalled(t, "ResetWithConfig", expected, mock.Anything)
	})

	t.Run("setwindow long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("setwindow 5")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultDoubtConfig()
		expected.DoubtWindowSec = 5
		m.AssertCalled(t, "ResetWithConfig", expected, mock.Anything)
	})

	t.Run("setwindow no args", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		result := c.Exec("sw")
		assert.True(t, msgRejected(result))
	})

	t.Run("setwindow invalid value", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		result := c.Exec("sw abc")
		assert.Contains(t, result, msgStem("invalidDoubtWindow160"))
	})

	t.Run("setwindow zero", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		result := c.Exec("sw 0")
		assert.Contains(t, result, msgStem("invalidDoubtWindow160"))
	})

	t.Run("setwindow over 60", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		result := c.Exec("sw 61")
		assert.Contains(t, result, msgStem("invalidDoubtWindow160"))
	})

	// setmemory tests
	t.Run("setmemory sm valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("sm 2")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultDoubtConfig()
		expected.CpuMemoryLevel = domain.DoubtMemoryLevelHard
		m.AssertCalled(t, "ResetWithConfig", expected, mock.Anything)
	})

	t.Run("setmemory long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("setmemory 0")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultDoubtConfig()
		expected.CpuMemoryLevel = domain.DoubtMemoryLevelEasy
		m.AssertCalled(t, "ResetWithConfig", expected, mock.Anything)
	})

	t.Run("setmemory no args", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		result := c.Exec("sm")
		assert.True(t, msgRejected(result))
	})

	t.Run("setmemory invalid value", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		result := c.Exec("sm abc")
		assert.Contains(t, result, msgStem("invalidCpuMemoryLevel02"))
	})

	t.Run("setmemory negative", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		result := c.Exec("sm -1")
		assert.Contains(t, result, msgStem("invalidCpuMemoryLevel02"))
	})

	t.Run("setmemory over 2", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		result := c.Exec("sm 3")
		assert.Contains(t, result, msgStem("invalidCpuMemoryLevel02"))
	})

	// setpenalty tests
	t.Run("setpenalty sp valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("sp 5")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultDoubtConfig()
		expected.PenaltyDrawLimit = 5
		m.AssertCalled(t, "ResetWithConfig", expected, mock.Anything)
	})

	t.Run("setpenalty long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("setpenalty 0")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultDoubtConfig()
		expected.PenaltyDrawLimit = 0
		m.AssertCalled(t, "ResetWithConfig", expected, mock.Anything)
	})

	t.Run("setpenalty no args", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		result := c.Exec("sp")
		assert.True(t, msgRejected(result))
	})

	t.Run("setpenalty invalid value", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		result := c.Exec("sp abc")
		assert.Contains(t, result, msgStem("invalidPenaltyDrawLimit0OrMore"))
	})

	t.Run("setpenalty negative", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		result := c.Exec("sp -1")
		assert.Contains(t, result, msgStem("invalidPenaltyDrawLimit0OrMore"))
	})

	// smetaai / smai tests
	t.Run("smetaai valid ON", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("smetaai 1")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultDoubtConfig()
		expected.CpuMetaAI = true
		m.AssertCalled(t, "ResetWithConfig", expected, mock.Anything)
	})

	t.Run("smai valid OFF", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("smai 0")
		assert.Equal(t, mockOutput, result)
		expected := domain.DefaultDoubtConfig()
		expected.CpuMetaAI = false
		m.AssertCalled(t, "ResetWithConfig", expected, mock.Anything)
	})

	t.Run("smetaai no args", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		result := c.Exec("smetaai")
		assert.Contains(t, result, msgStem("metaAiFlagRequired0Off1On"))
	})

	t.Run("smetaai invalid value", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		assert.Contains(t, c.Exec("smetaai 2"), msgStem("invalidMetaAiFlag01"))
		assert.Contains(t, c.Exec("smai abc"), msgStem("invalidMetaAiFlag01"))
		assert.Contains(t, c.Exec("smai -1"), msgStem("invalidMetaAiFlag01"))
	})

	// rp / resetprofile tests
	t.Run("rp command", func(t *testing.T) {
		m := new(mockUsecases.MockDoubtInteractor)
		m.On("ResetProfile").Return(mockOutput)
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("rp")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResetProfile")
	})

	t.Run("resetprofile command", func(t *testing.T) {
		m := new(mockUsecases.MockDoubtInteractor)
		m.On("ResetProfile").Return(mockOutput)
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("resetprofile")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResetProfile")
	})

	t.Run("play command refuses invalid card args", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("p 5 abc 2")
		assert.Contains(t, result, "abc")
		m.AssertNotCalled(t, "Play", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("doubt command refuses invalid args", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("d xyz 1")
		assert.Contains(t, result, "xyz")
		m.AssertNotCalled(t, "ResolveDoubt", mock.Anything)
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		result := c.Exec("unknown")
		assert.Contains(t, result, "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		result := c.Exec("")
		assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
	})
}

// #5390: `p abc` は宣言する値を 0 に落として通っていた。
func TestDoubtCuiController_PlayRefusesMistypedValue(t *testing.T) {
	mockOutput := `{"players":[]}`
	m := new(mockUsecases.MockDoubtInteractor)
	m.On("Reset").Return(mockOutput)
	m.On("Play", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)

	out := controller.NewDoubtCuiController(m).Exec("p abc")
	assert.Equal(t, msgKey("invalidClaimedValue", "val", "abc"), out)
	m.AssertNotCalled(t, "Play", mock.Anything, mock.Anything, mock.Anything)
}
