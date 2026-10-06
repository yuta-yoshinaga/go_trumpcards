//go:build test

package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoubtCui_HelpLinesDoNotAdvertiseWindowSetting(t *testing.T) {
	for _, line := range NewDoubtCui().HelpLines() {
		assert.NotContains(t, strings.ToLower(line), "window")
	}
}
