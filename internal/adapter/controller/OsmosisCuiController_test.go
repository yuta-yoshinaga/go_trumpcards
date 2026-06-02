package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockOsmosisInteractor() *mockusecase.MockOsmosisInteractor {
	return new(mockusecase.MockOsmosisInteractor)
}

func TestOsmosisCuiControllerQuit(t *testing.T) {
	c := NewOsmosisCuiController(newMockOsmosisInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestOsmosisCuiControllerReset(t *testing.T) {
	oi := newMockOsmosisInteractor()
	c := NewOsmosisCuiController(oi)
	oi.On("Reset").Return("reset")
	assert.Equal(t, "reset", c.Exec("r"))
	assert.Equal(t, "reset", c.Exec("reset"))
}

func TestOsmosisCuiControllerNoArgCommands(t *testing.T) {
	cases := []struct {
		method  string
		aliases []string
	}{
		{"Draw", []string{"d", "draw"}},
		{"GiveUp", []string{"g", "giveup"}},
		{"AutoComplete", []string{"ac", "autocomplete"}},
		{"Undo", []string{"u", "undo"}},
		{"Hint", []string{"h", "hint"}},
		{"ActionLog", []string{"log", "l"}},
	}
	for _, tc := range cases {
		for _, alias := range tc.aliases {
			oi := newMockOsmosisInteractor()
			c := NewOsmosisCuiController(oi)
			oi.On(tc.method).Return("out")
			assert.Equal(t, "out", c.Exec(alias), "method=%s alias=%s", tc.method, alias)
		}
	}
}

func TestOsmosisCuiControllerMoveWasteToFoundation(t *testing.T) {
	oi := newMockOsmosisInteractor()
	c := NewOsmosisCuiController(oi)
	oi.On("MoveWasteToFoundation", 2).Return("mwf")
	assert.Equal(t, "mwf", c.Exec("m w f 2"))
}

func TestOsmosisCuiControllerMoveReserveToFoundation(t *testing.T) {
	oi := newMockOsmosisInteractor()
	c := NewOsmosisCuiController(oi)
	oi.On("MoveReserveToFoundation", 1, 3).Return("mrf")
	assert.Equal(t, "mrf", c.Exec("m r 1 f 3"))
}

func TestOsmosisCuiControllerMoveErrors(t *testing.T) {
	t.Run("empty move prompts", func(t *testing.T) {
		c := NewOsmosisCuiController(newMockOsmosisInteractor())
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m")))
	})
	t.Run("invalid from", func(t *testing.T) {
		c := NewOsmosisCuiController(newMockOsmosisInteractor())
		assert.Contains(t, c.Exec("m x f 0"), "x")
	})
	t.Run("waste usage when not f", func(t *testing.T) {
		c := NewOsmosisCuiController(newMockOsmosisInteractor())
		out := c.Exec("m w x")
		assert.NotEmpty(t, out)
		assert.False(t, cuiutil.IsPromptRequest(out))
	})
	t.Run("waste missing target", func(t *testing.T) {
		c := NewOsmosisCuiController(newMockOsmosisInteractor())
		assert.NotEmpty(t, c.Exec("m w"))
	})
	t.Run("waste foundation prompt", func(t *testing.T) {
		c := NewOsmosisCuiController(newMockOsmosisInteractor())
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m w f")))
	})
	t.Run("waste invalid foundation", func(t *testing.T) {
		c := NewOsmosisCuiController(newMockOsmosisInteractor())
		assert.Contains(t, c.Exec("m w f abc"), "abc")
	})
	t.Run("reserve prompt for column", func(t *testing.T) {
		c := NewOsmosisCuiController(newMockOsmosisInteractor())
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m r")))
	})
	t.Run("reserve invalid column", func(t *testing.T) {
		c := NewOsmosisCuiController(newMockOsmosisInteractor())
		assert.Contains(t, c.Exec("m r abc"), "abc")
	})
	t.Run("reserve usage when missing f", func(t *testing.T) {
		c := NewOsmosisCuiController(newMockOsmosisInteractor())
		assert.NotEmpty(t, c.Exec("m r 0"))
	})
	t.Run("reserve usage when not f", func(t *testing.T) {
		c := NewOsmosisCuiController(newMockOsmosisInteractor())
		assert.NotEmpty(t, c.Exec("m r 0 x"))
	})
	t.Run("reserve foundation prompt", func(t *testing.T) {
		c := NewOsmosisCuiController(newMockOsmosisInteractor())
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m r 0 f")))
	})
	t.Run("reserve invalid foundation", func(t *testing.T) {
		c := NewOsmosisCuiController(newMockOsmosisInteractor())
		assert.Contains(t, c.Exec("m r 0 f abc"), "abc")
	})
}

func TestOsmosisCuiControllerUnknown(t *testing.T) {
	c := NewOsmosisCuiController(newMockOsmosisInteractor())
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

func TestOsmosisCuiControllerEmpty(t *testing.T) {
	c := NewOsmosisCuiController(newMockOsmosisInteractor())
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}
