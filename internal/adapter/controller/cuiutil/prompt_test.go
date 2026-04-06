//go:build test

package cuiutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
)

func TestPromptRequest_Format(t *testing.T) {
	got := cuiutil.PromptRequest("Enter bet amount:", "b {0}")
	assert.Equal(t, "PROMPT:Enter bet amount:\tb {0}", got)
}

func TestIsPromptRequest(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"valid prompt", "PROMPT:msg\ttmpl", true},
		{"prefix only", "PROMPT:", true},
		{"empty", "", false},
		{"no prefix", "Enter bet amount:", false},
		{"partial prefix", "PROMPT", false},
		{"lowercase", "prompt:msg\ttmpl", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, cuiutil.IsPromptRequest(tt.s))
		})
	}
}

func TestParsePromptRequest(t *testing.T) {
	tests := []struct {
		name         string
		s            string
		wantPrompt   string
		wantTemplate string
	}{
		{
			name:         "valid",
			s:            "PROMPT:Enter bet amount:\tb {0}",
			wantPrompt:   "Enter bet amount:",
			wantTemplate: "b {0}",
		},
		{
			name:         "no tab (malformed)",
			s:            "PROMPT:Enter bet amount:",
			wantPrompt:   "Enter bet amount:",
			wantTemplate: "",
		},
		{
			name:         "not a prompt request",
			s:            "some error message",
			wantPrompt:   "",
			wantTemplate: "",
		},
		{
			name:         "empty after prefix",
			s:            "PROMPT:",
			wantPrompt:   "",
			wantTemplate: "",
		},
		{
			name:         "multiple tabs",
			s:            "PROMPT:msg\ttmpl\textra",
			wantPrompt:   "msg",
			wantTemplate: "tmpl\textra",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, tmpl := cuiutil.ParsePromptRequest(tt.s)
			assert.Equal(t, tt.wantPrompt, p)
			assert.Equal(t, tt.wantTemplate, tmpl)
		})
	}
}

func TestFillTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		value    string
		want     string
	}{
		{"simple", "b {0}", "100", "b 100"},
		{"no placeholder", "fold", "100", "fold"},
		{"multiple placeholders", "m t {0} {0}", "3", "m t 3 {0}"},
		{"empty value", "b {0}", "", "b "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, cuiutil.FillTemplate(tt.template, tt.value))
		})
	}
}

func TestPromptRequest_Roundtrip(t *testing.T) {
	msg := "Enter source zone (t/c):"
	tmpl := "m {0}"
	encoded := cuiutil.PromptRequest(msg, tmpl)
	assert.True(t, cuiutil.IsPromptRequest(encoded))
	gotMsg, gotTmpl := cuiutil.ParsePromptRequest(encoded)
	assert.Equal(t, msg, gotMsg)
	assert.Equal(t, tmpl, gotTmpl)
	filled := cuiutil.FillTemplate(gotTmpl, "t")
	assert.Equal(t, "m t", filled)
}
