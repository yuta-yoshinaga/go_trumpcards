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

func newDkWebController() *controller.DoubleKlondikeWebController {
	factory := func() uc.DoubleKlondikeInteractorIF {
		g := domain.NewDefaultDoubleKlondike()
		g.Reset()
		return uc.NewDoubleKlondikeInteractor(g, new(presenter.DoubleKlondikeWebPresenter))
	}
	return controller.NewDoubleKlondikeWebController(factory)
}

func TestDoubleKlondikeWebController_Exec(t *testing.T) {
	ctrl := newDkWebController()
	defer ctrl.Stop()

	run := func(t *testing.T, body string, code int) {
		t.Helper()
		rec := execRequest(t, ctrl.Exec, bytes.NewReader([]byte(body)))
		rec.CodeIs(code)
		rec.ContentTypeIsJson()
	}

	t.Run("reset", func(t *testing.T) { run(t, `{"command":"reset","sessionId":"s1"}`, http.StatusOK) })
	t.Run("draw", func(t *testing.T) { run(t, `{"command":"d","sessionId":"s1"}`, http.StatusOK) })
	t.Run("moves", func(t *testing.T) {
		run(t, `{"command":"mwt","col":0,"sessionId":"s1"}`, http.StatusOK)
		run(t, `{"command":"mwf","sessionId":"s1"}`, http.StatusOK)
		run(t, `{"command":"mtf","col":0,"sessionId":"s1"}`, http.StatusOK)
		run(t, `{"command":"mtt","fromCol":0,"cardIndex":0,"toCol":1,"sessionId":"s1"}`, http.StatusOK)
	})
	t.Run("simple commands", func(t *testing.T) {
		for _, cmd := range []string{"g", "ac", "u", "hint", "log"} {
			run(t, `{"command":"`+cmd+`","sessionId":"s1"}`, http.StatusOK)
		}
	})
	t.Run("undo_n", func(t *testing.T) { run(t, `{"command":"undo_n","n":1,"sessionId":"s1"}`, http.StatusOK) })
	t.Run("missing params are bad requests", func(t *testing.T) {
		run(t, `{"command":"mwt","sessionId":"s1"}`, http.StatusBadRequest)
		run(t, `{"command":"mtf","sessionId":"s1"}`, http.StatusBadRequest)
		run(t, `{"command":"mtt","fromCol":0,"sessionId":"s1"}`, http.StatusBadRequest)
	})
}
