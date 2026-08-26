package controller

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newMockSlyFoxInteractor() *mockusecase.MockSlyFoxInteractor {
	return new(mockusecase.MockSlyFoxInteractor)
}

func TestSlyFoxCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"Reset", []string{"r", "reset"}},
		{"GiveUp", []string{"g", "giveup"}},
		{"Hint", []string{"h", "hint"}},
		{"AutoComplete", []string{"ac", "autocomplete"}},
		{"ActionLog", []string{"log", "l"}},
		{"Undo", []string{"u", "undo"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			ci := newMockSlyFoxInteractor()
			c := NewSlyFoxCuiController(ci)
			ci.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestSlyFoxCuiControllerQuit(t *testing.T) {
	c := NewSlyFoxCuiController(newMockSlyFoxInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestSlyFoxCuiControllerMoves(t *testing.T) {
	t.Run("tableau to a foundation", func(t *testing.T) {
		ci := newMockSlyFoxInteractor()
		c := NewSlyFoxCuiController(ci)
		ci.On("MoveTableauToFoundation", 2).Return("tf")
		assert.Equal(t, "tf", c.Exec("m t 2 f"))
	})

	// A tableau card has one legal destination, so the destination zone is
	// optional -- and anything other than `f` in that slot is a mistake, not a
	// silently-ignored extra word.
	t.Run("tableau without a destination zone", func(t *testing.T) {
		ci := newMockSlyFoxInteractor()
		c := NewSlyFoxCuiController(ci)
		ci.On("MoveTableauToFoundation", 2).Return("tf")
		assert.Equal(t, "tf", c.Exec("m t 2"))
	})

	t.Run("tableau to a tableau is rejected", func(t *testing.T) {
		ci := newMockSlyFoxInteractor()
		c := NewSlyFoxCuiController(ci)
		// 1 文字の Contains は本文がほぼ何であっても通ってしまうので、
		// 実際に出る文言そのものを見る。
		assert.Equal(t, invalidArg("slyfox.invalidToZone", "val", "t"), c.Exec("m t 0 t 5"))
		ci.AssertNotCalled(t, "MoveTableauToFoundation", 0)
	})

	t.Run("deal onto a slot", func(t *testing.T) {
		ci := newMockSlyFoxInteractor()
		c := NewSlyFoxCuiController(ci)
		ci.On("DealToPile", 4).Return("dp")
		assert.Equal(t, "dp", c.Exec("d 4"))
	})

	t.Run("deal straight to a foundation", func(t *testing.T) {
		ci := newMockSlyFoxInteractor()
		c := NewSlyFoxCuiController(ci)
		ci.On("DealToFoundation", 1).Return("df")
		assert.Equal(t, "df", c.Exec("d f 1"))
	})

	// **捨て札も山札も移動元ではない。**クローン元のコロラドから引き継いだ
	// `m w ...` / `m s ...` をリネームだけして残すと、ヘルプに無い構文が
	// インタラクターまで届いてしまう。
	for _, cmd := range []string{"m w f", "m w t 2", "m s t 3"} {
		t.Run("rejects the removed syntax "+cmd, func(t *testing.T) {
			ci := newMockSlyFoxInteractor()
			c := NewSlyFoxCuiController(ci)
			out := c.Exec(cmd)
			assert.False(t, cuiutil.IsPromptRequest(out), "%q は入力待ちにならない", cmd)
			assert.Contains(t, out, i18n.Tf("slyfox.invalidFromZone", "val", strings.Fields(cmd)[1]))
			ci.AssertNotCalled(t, "MoveTableauToFoundation")
			ci.AssertNotCalled(t, "DealToPile")
		})
	}

}

func TestSlyFoxCuiControllerPrompts(t *testing.T) {
	// 配りは行き先を必ず訊く。捨て札が無いので、決めずには配れない。
	for _, cmd := range []string{"m", "m t", "d", "d f"} {
		t.Run(cmd, func(t *testing.T) {
			c := NewSlyFoxCuiController(newMockSlyFoxInteractor())
			assert.True(t, cuiutil.IsPromptRequest(c.Exec(cmd)), "%q should prompt for more input", cmd)
		})
	}
}

func TestSlyFoxCuiControllerErrors(t *testing.T) {
	// 期待値は完全一致で持つ。部分一致だと "t" のような 1 文字が
	// 「たまたま含まれている」だけで通り、壊れても気付けない。
	for _, tc := range []struct{ cmd, want string }{
		{"m x f", invalidArg("slyfox.invalidFromZone", "val", "x")},
		{"m t abc f", invalidArg("slyfox.invalidPile", "val", "abc")},
		{"m t 0 z", invalidArg("slyfox.invalidToZone", "val", "z")},
		{"m t 0 t", invalidArg("slyfox.invalidToZone", "val", "t")},
		{"d abc", invalidArg("slyfox.invalidPile", "val", "abc")},
		{"d f abc", invalidArg("slyfox.invalidFoundation", "val", "abc")},
	} {
		t.Run(tc.cmd, func(t *testing.T) {
			ci := newMockSlyFoxInteractor()
			c := NewSlyFoxCuiController(ci)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			ci.AssertExpectations(t)
		})
	}
}
