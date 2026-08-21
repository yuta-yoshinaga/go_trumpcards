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

func newSsWebControllerCW() *controller.CurdsAndWheyWebController {
	factory := func() uc.CurdsAndWheyInteractorIF {
		g := domain.NewDefaultCurdsAndWhey()
		g.Reset()
		return uc.NewCurdsAndWheyInteractor(g, new(presenter.CurdsAndWheyWebPresenter))
	}
	return controller.NewCurdsAndWheyWebController(factory)
}

func TestCurdsAndWheyWebController_Exec(t *testing.T) {
	ctrl := newSsWebControllerCW()
	defer ctrl.Stop()

	run := func(t *testing.T, body string, code int) {
		t.Helper()
		rec := execRequest(t, ctrl.Exec, bytes.NewReader([]byte(body)))
		rec.CodeIs(code)
		rec.ContentTypeIsJson()
	}

	t.Run("reset", func(t *testing.T) { run(t, `{"command":"reset","sessionId":"s1"}`, http.StatusOK) })
	t.Run("move", func(t *testing.T) {
		run(t, `{"command":"m","fromCol":0,"cardIndex":0,"toCol":1,"sessionId":"s1"}`, http.StatusOK)
	})
	t.Run("simple commands", func(t *testing.T) {
		for _, cmd := range []string{"g", "u", "hint", "log"} {
			run(t, `{"command":"`+cmd+`","sessionId":"s1"}`, http.StatusOK)
		}
	})
	t.Run("undo_n", func(t *testing.T) { run(t, `{"command":"undo_n","n":1,"sessionId":"s1"}`, http.StatusOK) })
	t.Run("move missing params is bad request", func(t *testing.T) { run(t, `{"command":"m","fromCol":0,"sessionId":"s1"}`, http.StatusBadRequest) })
}
