//go:build test

package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecAndPresent(t *testing.T) {
	t.Run("action returns nil error", func(t *testing.T) {
		g := &mockGameEndChecker{}
		p := &recordingPresenter[*mockGameEndChecker]{successOutput: "ok"}
		out := execAndPresent(g, p, func() error { return nil })
		assert.Equal(t, "ok", out)
		assert.NoError(t, p.lastErr)
	})

	t.Run("action returns error", func(t *testing.T) {
		g := &mockGameEndChecker{}
		called := false
		p := &recordingPresenter[*mockGameEndChecker]{}
		out := execAndPresent(g, p, func() error {
			called = true
			return errors.New("fail")
		})
		assert.True(t, called)
		assert.Equal(t, "fail", out)
		assert.Error(t, p.lastErr)
	})
}

func TestRunAndPresent(t *testing.T) {
	t.Run("action is called and result is presented with nil error", func(t *testing.T) {
		g := &mockGameEndChecker{}
		called := false
		p := &recordingPresenter[*mockGameEndChecker]{}
		out := runAndPresent(g, p, func() { called = true })
		assert.True(t, called)
		assert.Equal(t, "", out)
		assert.NoError(t, p.lastErr)
	})
}

func TestResetWithValidatedConfig(t *testing.T) {
	t.Run("validation error returns error output", func(t *testing.T) {
		g := &mockGameEndChecker{}
		p := &recordingPresenter[*mockGameEndChecker]{}
		cfg := &mockValidatableConfig{err: errors.New("bad config")}
		resetCalled := false
		out := resetWithValidatedConfig(g, p, cfg, func(_ *mockValidatableConfig) {}, func() string {
			resetCalled = true
			return "reset"
		})
		assert.False(t, resetCalled)
		assert.Equal(t, "bad config", out)
		assert.Error(t, p.lastErr)
	})

	t.Run("valid config calls setConfig and resetAndPresent", func(t *testing.T) {
		g := &mockGameEndChecker{}
		p := &recordingPresenter[*mockGameEndChecker]{successOutput: "ok"}
		cfg := &mockValidatableConfig{}
		configSet := false
		out := resetWithValidatedConfig(g, p, cfg, func(c *mockValidatableConfig) {
			configSet = true
		}, func() string {
			return "reset-output"
		})
		assert.True(t, configSet)
		assert.Equal(t, "reset-output", out)
	})
}

type mockValidatableConfig struct {
	err error
}

func (c *mockValidatableConfig) Validate() error { return c.err }

// recordingPresenter records the last error passed to Output.
type recordingPresenter[G any] struct {
	lastErr       error
	successOutput string
}

func (r *recordingPresenter[G]) Output(_ G, err error) string {
	r.lastErr = err
	if err != nil {
		return err.Error()
	}
	return r.successOutput
}
