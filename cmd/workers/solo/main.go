//go:build js && wasm

package main

import (
	"log"
	"net/http"

	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	corsmw "github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/cors"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// registerKV creates a KV-backed session provider for a game and registers the
// controller on the given mux. It eliminates the repeated boilerplate of
// creating a provider, building a controller, and wiring up the route.
func registerKV[I any](
	mux *http.ServeMux,
	path string,
	kvPrefix string,
	factory func() I,
	restore func([]byte) (I, error),
	newCtrl func(controller.SessionProvider[I], func() I) interface {
		Exec(http.ResponseWriter, *http.Request)
		Stop()
	},
) {
	kvProvider, err := controller.NewKVSessionProvider[I](
		"GAME_SESSIONS", kvPrefix,
		func(i I) ([]byte, error) {
			return any(i).(interface{ Snapshot() ([]byte, error) }).Snapshot()
		},
		restore,
	)
	if err != nil {
		log.Fatal(err)
	}
	ctrl := newCtrl(kvProvider, factory)
	mux.HandleFunc(path, ctrl.Exec)
}

func main() {
	mux := http.NewServeMux()

	// Klondike
	registerKV(mux, "/klondike/exec", "klondike:",
		func() usecase.KlondikeInteractorIF {
			klondike := domain.NewKlondike(domain.NewTrumpCards(0))
			return usecase.NewKlondikeInteractor(klondike, new(presenter.KlondikeWebPresenter))
		},
		func(data []byte) (usecase.KlondikeInteractorIF, error) {
			return usecase.RestoreKlondikeInteractor(data, new(presenter.KlondikeWebPresenter))
		},
		func(p controller.SessionProvider[usecase.KlondikeInteractorIF], f func() usecase.KlondikeInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewKlondikeWebControllerWithProvider(p, f)
		},
	)

	// FreeCell
	registerKV(mux, "/freecell/exec", "freecell:",
		func() usecase.FreeCellInteractorIF {
			freeCell := domain.NewFreeCell(domain.NewTrumpCards(0))
			return usecase.NewFreeCellInteractor(freeCell, new(presenter.FreeCellWebPresenter))
		},
		func(data []byte) (usecase.FreeCellInteractorIF, error) {
			return usecase.RestoreFreeCellInteractor(data, new(presenter.FreeCellWebPresenter))
		},
		func(p controller.SessionProvider[usecase.FreeCellInteractorIF], f func() usecase.FreeCellInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewFreeCellWebControllerWithProvider(p, f)
		},
	)

	// Spider
	registerKV(mux, "/spider/exec", "spider:",
		func() usecase.SpiderInteractorIF {
			spider := domain.NewSpider(domain.NewTrumpCardsWithSuits(domain.SpiderTotalCards, []int{domain.CardDesignSpade}))
			return usecase.NewSpiderInteractor(spider, new(presenter.SpiderWebPresenter))
		},
		func(data []byte) (usecase.SpiderInteractorIF, error) {
			return usecase.RestoreSpiderInteractor(data, new(presenter.SpiderWebPresenter))
		},
		func(p controller.SessionProvider[usecase.SpiderInteractorIF], f func() usecase.SpiderInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewSpiderWebControllerWithProvider(p, f)
		},
	)

	// Pyramid
	registerKV(mux, "/pyramid/exec", "pyramid:",
		func() usecase.PyramidInteractorIF {
			pyramid := domain.NewPyramid(domain.NewTrumpCards(0))
			return usecase.NewPyramidInteractor(pyramid, new(presenter.PyramidWebPresenter))
		},
		func(data []byte) (usecase.PyramidInteractorIF, error) {
			return usecase.RestorePyramidInteractor(data, new(presenter.PyramidWebPresenter))
		},
		func(p controller.SessionProvider[usecase.PyramidInteractorIF], f func() usecase.PyramidInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewPyramidWebControllerWithProvider(p, f)
		},
	)

	// Memory
	registerKV(mux, "/memory/exec", "memory:",
		func() usecase.MemoryInteractorIF {
			config := domain.DefaultMemoryConfig()
			players := []*domain.MemoryPlayer{
				domain.NewMemoryPlayer(true),
				domain.NewMemoryPlayer(false),
				domain.NewMemoryPlayer(false),
				domain.NewMemoryPlayer(false),
			}
			memory := domain.NewMemory(domain.NewTrumpCards(0), players, config)
			return usecase.NewMemoryInteractor(memory, new(presenter.MemoryWebPresenter))
		},
		func(data []byte) (usecase.MemoryInteractorIF, error) {
			return usecase.RestoreMemoryInteractor(data, new(presenter.MemoryWebPresenter))
		},
		func(p controller.SessionProvider[usecase.MemoryInteractorIF], f func() usecase.MemoryInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewMemoryWebControllerWithProvider(p, f)
		},
	)

	// Gin Rummy
	registerKV(mux, "/ginrummy/exec", "ginrummy:",
		func() usecase.GinRummyInteractorIF {
			config := domain.DefaultGinRummyConfig()
			players := []*domain.GinRummyPlayer{
				domain.NewGinRummyPlayer(true),
				domain.NewGinRummyPlayer(false),
			}
			gr := domain.NewGinRummy(domain.NewTrumpCards(0), players, config)
			return usecase.NewGinRummyInteractor(gr, new(presenter.GinRummyWebPresenter))
		},
		func(data []byte) (usecase.GinRummyInteractorIF, error) {
			return usecase.RestoreGinRummyInteractor(data, new(presenter.GinRummyWebPresenter))
		},
		func(p controller.SessionProvider[usecase.GinRummyInteractorIF], f func() usecase.GinRummyInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewGinRummyWebControllerWithProvider(p, f)
		},
	)

	// Cribbage
	registerKV(mux, "/cribbage/exec", "cribbage:",
		func() usecase.CribbageInteractorIF {
			config := domain.DefaultCribbageConfig()
			players := []*domain.CribbagePlayer{
				domain.NewCribbagePlayer(true),
				domain.NewCribbagePlayer(false),
			}
			cribbage := domain.NewCribbage(domain.NewTrumpCards(0), players, config)
			return usecase.NewCribbageInteractor(cribbage, new(presenter.CribbageWebPresenter))
		},
		func(data []byte) (usecase.CribbageInteractorIF, error) {
			return usecase.RestoreCribbageInteractor(data, new(presenter.CribbageWebPresenter))
		},
		func(p controller.SessionProvider[usecase.CribbageInteractorIF], f func() usecase.CribbageInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewCribbageWebControllerWithProvider(p, f)
		},
	)

	var handler http.Handler = mux
	if origins := corsmw.ParseOrigins(cloudflare.Getenv("CORS_ALLOWED_ORIGINS")); origins != nil {
		handler = corsmw.Middleware(origins, mux)
	}
	workers.Serve(handler)
}
