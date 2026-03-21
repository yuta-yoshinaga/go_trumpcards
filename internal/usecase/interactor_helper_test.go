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
		p := &mockPresenter[*mockGameEndChecker]{output: "ok"}
		out := execAndPresent(g, p, func() error { return nil })
		assert.Equal(t, "ok", out)
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

// recordingPresenter records the last error passed to Output.
type recordingPresenter[G any] struct {
	lastErr error
}

func (r *recordingPresenter[G]) Output(_ G, err error) string {
	r.lastErr = err
	if err != nil {
		return err.Error()
	}
	return ""
}
