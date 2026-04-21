//go:build js && wasm

package games

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/worker"
)

// RegisterKVGame binds a KV-backed handler for name under the given category.
// URL path and KV prefix are derived from name by the project convention
// (/<name>/exec and "<name>:"), so each call site only has to supply the
// ingredients that actually vary per game.
//
// This helper is the DRY/SSoT answer to issue #1458 for Cloudflare Workers —
// Go infers the I/P/O type parameters from newCtrl, so call sites stay
// untyped and a game registration stays to the minimum information (name,
// category, factory, restore, constructor).
func RegisterKVGame[I worker.SnapshotIF, P controller.WebInput, O any](
	name string,
	category Category,
	factory func() I,
	restore func([]byte) (I, error),
	newCtrl func(controller.SessionProvider[I], func() I) *controller.GameWebController[I, P, O],
) {
	BindWorker(name, category, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/"+name+"/exec", name+":", factory, restore, newCtrl)
	})
}
