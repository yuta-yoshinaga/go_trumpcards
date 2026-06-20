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

func newLlWebController() *controller.LaBelleLucieWebController {
	factory := func() uc.LaBelleLucieInteractorIF {
		g := domain.NewDefaultLaBelleLucie()
		g.Reset()
		return uc.NewLaBelleLucieInteractor(g, new(presenter.LaBelleLucieWebPresenter))
	}
	return controller.NewLaBelleLucieWebController(factory)
}

func TestLaBelleLucieWebController_Exec(t *testing.T) {
	ctrl := newLlWebController()
	defer ctrl.Stop()

	run := func(t *testing.T, body string, code int) {
		t.Helper()
		rec := execRequest(t, ctrl.Exec, bytes.NewReader([]byte(body)))
		rec.CodeIs(code)
		rec.ContentTypeIsJson()
	}

	t.Run("reset", func(t *testing.T) { run(t, `{"command":"reset","sessionId":"s1"}`, http.StatusOK) })
	t.Run("fan to fan", func(t *testing.T) { run(t, `{"command":"mf","from":0,"to":1,"sessionId":"s1"}`, http.StatusOK) })
	t.Run("fan to foundation", func(t *testing.T) { run(t, `{"command":"ff","from":0,"sessionId":"s1"}`, http.StatusOK) })
	t.Run("simple commands", func(t *testing.T) {
		for _, cmd := range []string{"rd", "g", "ac", "u", "hint", "log"} {
			run(t, `{"command":"`+cmd+`","sessionId":"s1"}`, http.StatusOK)
		}
	})
	t.Run("undo_n", func(t *testing.T) { run(t, `{"command":"undo_n","n":2,"sessionId":"s1"}`, http.StatusOK) })
	t.Run("mf without to is a bad request", func(t *testing.T) { run(t, `{"command":"mf","from":0,"sessionId":"s1"}`, http.StatusBadRequest) })
	t.Run("ff without from is a bad request", func(t *testing.T) { run(t, `{"command":"ff","sessionId":"s1"}`, http.StatusBadRequest) })
}
