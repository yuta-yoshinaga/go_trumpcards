//go:build test

package controller_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestMinchiateCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockMinchiateInteractor {
		m := new(mockUsecases.MockMinchiateInteractor)
		m.On("GetConfig").Return(domain.DefaultMinchiateConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewMinchiateCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewMinchiateCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewMinchiateCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultMinchiateConfig())
	})

	// **枚数は定数から組み立てる。**Tarocchini から写すと 2 枚のままになるが、
	// Minchiate の余剰は 13 枚。テストに数字を直書きすると誤った枚数が仕様として
	// 読まれる。
	surplusArgs := func() (string, []int) {
		var sb strings.Builder
		idx := make([]int, 0, domain.MinchiateSurplus)
		for i := 0; i < domain.MinchiateSurplus; i++ {
			fmt.Fprintf(&sb, " %d", i)
			idx = append(idx, i)
		}
		return sb.String(), idx
	}

	t.Run("scarto and its discard alias", func(t *testing.T) {
		args, idx := surplusArgs()
		for _, cmd := range []string{"scarto" + args, "discard" + args} {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewMinchiateCuiController(m).Exec(cmd))
			m.AssertCalled(t, "Discard", idx)
		}
	})

	// 案内文の枚数は MinchiateSurplus から出す。数字を直接書くと余剰枚数が
	// 変わったときに案内だけ古くなる。
	t.Run("scarto with too few indices names the required count", func(t *testing.T) {
		out := controller.NewMinchiateCuiController(newMock()).Exec("scarto 0")
		assert.Contains(t, out, msgKey("cardIndicesRequiredScartoN", "n", fmt.Sprint(domain.MinchiateSurplus)))
	})

	// **枚数は足りているが中身が数字でない場合。**引数を 2 個だけ渡すと枚数検査で
	// 先に弾かれ、数値検査を一度も踏まない。
	t.Run("scarto with a non-numeric index", func(t *testing.T) {
		args, _ := surplusArgs()
		out := controller.NewMinchiateCuiController(newMock()).Exec("scarto" + args + " x")
		assert.Contains(t, out, msgInvalidCardIndexPrefix())
	})

	t.Run("play card", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewMinchiateCuiController(m).Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		assert.Contains(t, controller.NewMinchiateCuiController(newMock()).Exec("play"), msgCardIndexRequired())
	})

	// このゲームに入札は無い。bid/pass が黙って別の動作に落ちてはならない。
	t.Run("bidding commands are not accepted", func(t *testing.T) {
		for _, cmd := range []string{"bid 1", "pass"} {
			assert.Contains(t, controller.NewMinchiateCuiController(newMock()).Exec(cmd), "コマンドが不明です")
		}
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewMinchiateCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewMinchiateCuiController(m).Exec("sd 2"))
		expected := domain.DefaultMinchiateConfig()
		expected.CpuDifficulty = domain.MinchiateCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		assert.Contains(t, controller.NewMinchiateCuiController(newMock()).Exec("sd 9"), msgInvalidCpuDifficultyPrefix())
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewMinchiateCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		assert.Contains(t, controller.NewMinchiateCuiController(newMock()).Exec("zzz"), "コマンドが不明です")
	})
}
