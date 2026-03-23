//go:build test

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockConfig struct {
	value string
}

type mockWebConfig struct {
	value string
}

func (c *mockWebConfig) ToConfig() mockConfig {
	return mockConfig{value: c.value}
}

func TestConfigOrDefault_NilConfig(t *testing.T) {
	var cfg *mockWebConfig
	result := configOrDefault(cfg, (*mockWebConfig).ToConfig, mockConfig{value: "default"})
	assert.Equal(t, "default", result.value)
}

func TestConfigOrDefault_NonNilConfig(t *testing.T) {
	cfg := &mockWebConfig{value: "custom"}
	result := configOrDefault(cfg, (*mockWebConfig).ToConfig, mockConfig{value: "default"})
	assert.Equal(t, "custom", result.value)
}
