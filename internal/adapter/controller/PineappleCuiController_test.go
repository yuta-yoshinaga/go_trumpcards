//go:build test

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func TestPineappleCuiController_Bet_NoAmount_Prompt(t *testing.T) {
	mi := new(usecase.MockPineappleInteractor)
	c := NewPineappleCuiController(mi)
	result := c.Exec("b")
	assert.True(t, cuiutil.IsPromptRequest(result))
	_, tmpl := cuiutil.ParsePromptRequest(result)
	assert.Equal(t, "b {0}", tmpl)
}

func TestPineappleCuiController_Raise_NoAmount_Prompt(t *testing.T) {
	mi := new(usecase.MockPineappleInteractor)
	c := NewPineappleCuiController(mi)
	result := c.Exec("ra")
	assert.True(t, cuiutil.IsPromptRequest(result))
	_, tmpl := cuiutil.ParsePromptRequest(result)
	assert.Equal(t, "ra {0}", tmpl)
}
