package controller

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newMockSalicLawInteractor() *mockusecase.MockSalicLawInteractor {
	return new(mockusecase.MockSalicLawInteractor)
}

func TestSalicLawCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"Reset", []string{"r", "reset"}},
		{"Draw", []string{"d", "draw"}},
		{"GiveUp", []string{"g", "giveup"}},
		{"Hint", []string{"h", "hint"}},
		{"AutoComplete", []string{"ac", "autocomplete"}},
		{"ActionLog", []string{"log", "l"}},
		{"Undo", []string{"u", "undo"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			ci := newMockSalicLawInteractor()
			c := NewSalicLawCuiController(ci)
			ci.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestSalicLawCuiControllerQuit(t *testing.T) {
	c := NewSalicLawCuiController(newMockSalicLawInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestSalicLawCuiControllerMoves(t *testing.T) {
	t.Run("tableau to a foundation", func(t *testing.T) {
		ci := newMockSalicLawInteractor()
		c := NewSalicLawCuiController(ci)
		ci.On("MoveTableauToFoundation", 2).Return("tf")
		assert.Equal(t, "tf", c.Exec("m t 2 f"))
	})

	t.Run("between tableau piles", func(t *testing.T) {
		ci := newMockSalicLawInteractor()
		c := NewSalicLawCuiController(ci)
		ci.On("MoveTableauToTableau", 0, 5).Return("tt")
		assert.Equal(t, "tt", c.Exec("m t 0 t 5"))
	})

	// **捨て札も山札からの直接置きも無い。**Congress から引き継いだ `m w ...` /
	// `m s ...` は、リネームだけして残すとヘルプに無い構文が半分動く状態になる。
	// コントローラごと消したので、移動元として拒まれること自体を見る。
	for _, cmd := range []string{"m w f", "m w t 2", "m s t 3", "m s f"} {
		t.Run("rejects the removed syntax "+cmd, func(t *testing.T) {
			ci := newMockSalicLawInteractor()
			c := NewSalicLawCuiController(ci)
			out := c.Exec(cmd)
			assert.False(t, cuiutil.IsPromptRequest(out), "%q は入力待ちにならない", cmd)
			assert.Contains(t, out, i18n.Tf("saliclaw.invalidFromZone", "val", strings.Fields(cmd)[1]))
			// インタラクターまで届いていないこと。届いていれば呼び出しが記録される。
			ci.AssertNotCalled(t, "MoveTableauToTableau")
			ci.AssertNotCalled(t, "MoveTableauToFoundation")
		})
	}
}

func TestSalicLawCuiControllerPrompts(t *testing.T) {
	for _, cmd := range []string{"m", "m t", "m t 0", "m t 0 t"} {
		t.Run(cmd, func(t *testing.T) {
			c := NewSalicLawCuiController(newMockSalicLawInteractor())
			assert.True(t, cuiutil.IsPromptRequest(c.Exec(cmd)), "%q should prompt for more input", cmd)
		})
	}
}

func TestSalicLawCuiControllerErrors(t *testing.T) {
	for _, tc := range []struct{ cmd, contains string }{
		{"m x f", "x"},
		{"m t abc f", "abc"},
		{"m t 0 z", "z"},
		{"m t 0 t abc", "abc"},
	} {
		t.Run(tc.cmd, func(t *testing.T) {
			c := NewSalicLawCuiController(newMockSalicLawInteractor())
			assert.Contains(t, c.Exec(tc.cmd), tc.contains)
		})
	}
}
