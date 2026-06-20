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

func newRbWebController() *controller.RussianBankWebController {
	factory := func() uc.RussianBankInteractorIF {
		g := domain.NewDefaultRussianBank()
		g.Reset()
		return uc.NewRussianBankInteractor(g, new(presenter.RussianBankWebPresenter))
	}
	return controller.NewRussianBankWebController(factory)
}

func TestRussianBankWebController_Exec(t *testing.T) {
	ctrl := newRbWebController()
	defer ctrl.Stop()

	run := func(t *testing.T, body string, code int) *recorded {
		t.Helper()
		rec := execRequest(t, ctrl.Exec, bytes.NewReader([]byte(body)))
		rec.CodeIs(code)
		rec.ContentTypeIsJson()
		return rec
	}

	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1","config":{"cpuDifficulty":2}}`, http.StatusOK)
	})
	t.Run("foundation move", func(t *testing.T) {
		run(t, `{"command":"pf","zone":0,"fromOpp":false,"col":0,"sessionId":"s1"}`, http.StatusOK)
	})
	t.Run("tableau move", func(t *testing.T) {
		run(t, `{"command":"mt","zone":1,"fromOpp":false,"col":0,"toCol":2,"sessionId":"s1"}`, http.StatusOK)
	})
	t.Run("discard / stop / undo / hint / log", func(t *testing.T) {
		for _, cmd := range []string{"d", "s", "u", "hint", "log"} {
			run(t, `{"command":"`+cmd+`","sessionId":"s1"}`, http.StatusOK)
		}
	})
	t.Run("quit", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, http.StatusOK)
	})
	t.Run("pf without zone is a bad request", func(t *testing.T) {
		run(t, `{"command":"pf","sessionId":"s1"}`, http.StatusBadRequest)
	})
	t.Run("mt without toCol is a bad request", func(t *testing.T) {
		run(t, `{"command":"mt","zone":0,"sessionId":"s1"}`, http.StatusBadRequest)
	})
}
