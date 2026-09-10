//go:build test
// +build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
)

func TestBassetWebController_Constructs(t *testing.T) {
	assert.NotNil(t, controller.NewBassetWebController)
}
