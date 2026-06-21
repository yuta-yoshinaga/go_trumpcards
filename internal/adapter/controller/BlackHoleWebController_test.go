//go:build test

package controller_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func newBhWebController() *controller.BlackHoleWebController {
	factory := func() uc.BlackHoleInteractorIF {
		g := domain.NewDefaultBlackHole()
		g.Reset()
		return uc.NewBlackHoleInteractor(g, new(presenter.BlackHoleWebPresenter))
	}
	return controller.NewBlackHoleWebController(factory)
}

func TestBlackHoleWebController_Exec(t *testing.T) {
	ctrl := newBhWebController()
	defer ctrl.Stop()

	run := func(t *testing.T, body string, code int) {
		t.Helper()
		rec := execRequest(t, ctrl.Exec, bytes.NewReader([]byte(body)))
		rec.CodeIs(code)
		rec.ContentTypeIsJson()
	}

	t.Run("reset", func(t *testing.T) { run(t, `{"command":"reset","sessionId":"s1"}`, http.StatusOK) })
	t.Run("move", func(t *testing.T) { run(t, `{"command":"mb","fan":0,"sessionId":"s1"}`, http.StatusOK) })
	t.Run("simple commands", func(t *testing.T) {
		for _, cmd := range []string{"g", "u", "hint", "log"} {
			run(t, `{"command":"`+cmd+`","sessionId":"s1"}`, http.StatusOK)
		}
	})
	t.Run("undo_n", func(t *testing.T) { run(t, `{"command":"undo_n","n":1,"sessionId":"s1"}`, http.StatusOK) })
	t.Run("missing params are bad requests", func(t *testing.T) {
		run(t, `{"command":"mb","sessionId":"s1"}`, http.StatusBadRequest)
		run(t, `{"command":"undo_n","sessionId":"s1"}`, http.StatusBadRequest)
	})
}
