//go:build js && wasm

package main

import (
	"fmt"
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

func main() {
	mux := http.NewServeMux()

	// Klondike (KV-backed session)
	klondikeFactory := func() usecase.KlondikeInteractorIF {
		klondike := domain.NewKlondike(domain.NewTrumpCards(0))
		return usecase.NewKlondikeInteractor(klondike, new(presenter.KlondikeWebPresenter))
	}
	klondikeKV, err := controller.NewKVSessionProvider[usecase.KlondikeInteractorIF](
		"GAME_SESSIONS", "klondike:",
		func(ki usecase.KlondikeInteractorIF) ([]byte, error) {
			snap, ok := ki.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.KlondikeInteractorIF, error) {
			return usecase.RestoreKlondikeInteractor(data, new(presenter.KlondikeWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	klc := controller.NewKlondikeWebControllerWithProvider(klondikeKV, klondikeFactory)
	mux.HandleFunc("/klondike/exec", klc.Exec)

	// FreeCell (KV-backed session)
	freecellFactory := func() usecase.FreeCellInteractorIF {
		freeCell := domain.NewFreeCell(domain.NewTrumpCards(0))
		return usecase.NewFreeCellInteractor(freeCell, new(presenter.FreeCellWebPresenter))
	}
	freecellKV, err := controller.NewKVSessionProvider[usecase.FreeCellInteractorIF](
		"GAME_SESSIONS", "freecell:",
		func(fi usecase.FreeCellInteractorIF) ([]byte, error) {
			snap, ok := fi.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.FreeCellInteractorIF, error) {
			return usecase.RestoreFreeCellInteractor(data, new(presenter.FreeCellWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fcc := controller.NewFreeCellWebControllerWithProvider(freecellKV, freecellFactory)
	mux.HandleFunc("/freecell/exec", fcc.Exec)

	// Spider (KV-backed session)
	spiderFactory := func() usecase.SpiderInteractorIF {
		spider := domain.NewSpider(domain.NewTrumpCardsWithSuits(domain.SpiderTotalCards, []int{domain.CardDesignSpade}))
		return usecase.NewSpiderInteractor(spider, new(presenter.SpiderWebPresenter))
	}
	spiderKV, err := controller.NewKVSessionProvider[usecase.SpiderInteractorIF](
		"GAME_SESSIONS", "spider:",
		func(si usecase.SpiderInteractorIF) ([]byte, error) {
			snap, ok := si.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.SpiderInteractorIF, error) {
			return usecase.RestoreSpiderInteractor(data, new(presenter.SpiderWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	sdc := controller.NewSpiderWebControllerWithProvider(spiderKV, spiderFactory)
	mux.HandleFunc("/spider/exec", sdc.Exec)

	// Pyramid (KV-backed session)
	pyramidFactory := func() usecase.PyramidInteractorIF {
		pyramid := domain.NewPyramid(domain.NewTrumpCards(0))
		return usecase.NewPyramidInteractor(pyramid, new(presenter.PyramidWebPresenter))
	}
	pyramidKV, err := controller.NewKVSessionProvider[usecase.PyramidInteractorIF](
		"GAME_SESSIONS", "pyramid:",
		func(pi usecase.PyramidInteractorIF) ([]byte, error) {
			snap, ok := pi.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.PyramidInteractorIF, error) {
			return usecase.RestorePyramidInteractor(data, new(presenter.PyramidWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	pyc := controller.NewPyramidWebControllerWithProvider(pyramidKV, pyramidFactory)
	mux.HandleFunc("/pyramid/exec", pyc.Exec)

	// Memory (KV-backed session)
	memoryFactory := func() usecase.MemoryInteractorIF {
		config := domain.DefaultMemoryConfig()
		players := []*domain.MemoryPlayer{
			domain.NewMemoryPlayer(true),
			domain.NewMemoryPlayer(false),
			domain.NewMemoryPlayer(false),
			domain.NewMemoryPlayer(false),
		}
		memory := domain.NewMemory(domain.NewTrumpCards(0), players, config)
		return usecase.NewMemoryInteractor(memory, new(presenter.MemoryWebPresenter))
	}
	memoryKV, err := controller.NewKVSessionProvider[usecase.MemoryInteractorIF](
		"GAME_SESSIONS", "memory:",
		func(mi usecase.MemoryInteractorIF) ([]byte, error) {
			snap, ok := mi.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.MemoryInteractorIF, error) {
			return usecase.RestoreMemoryInteractor(data, new(presenter.MemoryWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	myc := controller.NewMemoryWebControllerWithProvider(memoryKV, memoryFactory)
	mux.HandleFunc("/memory/exec", myc.Exec)

	// Gin Rummy (KV-backed session)
	ginrummyFactory := func() usecase.GinRummyInteractorIF {
		config := domain.DefaultGinRummyConfig()
		players := []*domain.GinRummyPlayer{
			domain.NewGinRummyPlayer(true),
			domain.NewGinRummyPlayer(false),
		}
		gr := domain.NewGinRummy(domain.NewTrumpCards(0), players, config)
		return usecase.NewGinRummyInteractor(gr, new(presenter.GinRummyWebPresenter))
	}
	ginrummyKV, err := controller.NewKVSessionProvider[usecase.GinRummyInteractorIF](
		"GAME_SESSIONS", "ginrummy:",
		func(gi usecase.GinRummyInteractorIF) ([]byte, error) {
			snap, ok := gi.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.GinRummyInteractorIF, error) {
			return usecase.RestoreGinRummyInteractor(data, new(presenter.GinRummyWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	grc := controller.NewGinRummyWebControllerWithProvider(ginrummyKV, ginrummyFactory)
	mux.HandleFunc("/ginrummy/exec", grc.Exec)

	// Cribbage (KV-backed session)
	cribbageFactory := func() usecase.CribbageInteractorIF {
		config := domain.DefaultCribbageConfig()
		players := []*domain.CribbagePlayer{
			domain.NewCribbagePlayer(true),
			domain.NewCribbagePlayer(false),
		}
		cribbage := domain.NewCribbage(domain.NewTrumpCards(0), players, config)
		return usecase.NewCribbageInteractor(cribbage, new(presenter.CribbageWebPresenter))
	}
	cribbageKV, err := controller.NewKVSessionProvider[usecase.CribbageInteractorIF](
		"GAME_SESSIONS", "cribbage:",
		func(ci usecase.CribbageInteractorIF) ([]byte, error) {
			snap, ok := ci.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.CribbageInteractorIF, error) {
			return usecase.RestoreCribbageInteractor(data, new(presenter.CribbageWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	cbc := controller.NewCribbageWebControllerWithProvider(cribbageKV, cribbageFactory)
	mux.HandleFunc("/cribbage/exec", cbc.Exec)

	var handler http.Handler = mux
	if origins := corsmw.ParseOrigins(cloudflare.Getenv("CORS_ALLOWED_ORIGINS")); origins != nil {
		handler = corsmw.Middleware(origins, mux)
	}
	workers.Serve(handler)
}
