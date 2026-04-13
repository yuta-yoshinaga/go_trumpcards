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
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/worker"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func main() {
	mux := http.NewServeMux()

	// Klondike
	if err := worker.RegisterKV(mux, "/klondike/exec", "klondike:",
		func() usecase.KlondikeInteractorIF {
			klondike := domain.NewKlondike(domain.NewTrumpCards(0))
			return usecase.NewKlondikeInteractor(klondike, new(presenter.KlondikeWebPresenter))
		},
		func(data []byte) (usecase.KlondikeInteractorIF, error) {
			return usecase.RestoreKlondikeInteractor(data, new(presenter.KlondikeWebPresenter))
		},
		controller.NewKlondikeWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// FreeCell
	if err := worker.RegisterKV(mux, "/freecell/exec", "freecell:",
		func() usecase.FreeCellInteractorIF {
			freeCell := domain.NewFreeCell(domain.NewTrumpCards(0))
			return usecase.NewFreeCellInteractor(freeCell, new(presenter.FreeCellWebPresenter))
		},
		func(data []byte) (usecase.FreeCellInteractorIF, error) {
			return usecase.RestoreFreeCellInteractor(data, new(presenter.FreeCellWebPresenter))
		},
		controller.NewFreeCellWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Spider
	if err := worker.RegisterKV(mux, "/spider/exec", "spider:",
		func() usecase.SpiderInteractorIF {
			spider := domain.NewSpider(domain.NewTrumpCardsWithSuits(domain.SpiderTotalCards, []int{domain.CardDesignSpade}))
			return usecase.NewSpiderInteractor(spider, new(presenter.SpiderWebPresenter))
		},
		func(data []byte) (usecase.SpiderInteractorIF, error) {
			return usecase.RestoreSpiderInteractor(data, new(presenter.SpiderWebPresenter))
		},
		controller.NewSpiderWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Pyramid
	if err := worker.RegisterKV(mux, "/pyramid/exec", "pyramid:",
		func() usecase.PyramidInteractorIF {
			pyramid := domain.NewPyramid(domain.NewTrumpCards(0))
			return usecase.NewPyramidInteractor(pyramid, new(presenter.PyramidWebPresenter))
		},
		func(data []byte) (usecase.PyramidInteractorIF, error) {
			return usecase.RestorePyramidInteractor(data, new(presenter.PyramidWebPresenter))
		},
		controller.NewPyramidWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// TriPeaks
	if err := worker.RegisterKV(mux, "/tripeaks/exec", "tripeaks:",
		func() usecase.TriPeaksInteractorIF {
			triPeaks := domain.NewTriPeaks(domain.NewTrumpCards(0))
			return usecase.NewTriPeaksInteractor(triPeaks, new(presenter.TriPeaksWebPresenter))
		},
		func(data []byte) (usecase.TriPeaksInteractorIF, error) {
			return usecase.RestoreTriPeaksInteractor(data, new(presenter.TriPeaksWebPresenter))
		},
		controller.NewTriPeaksWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Memory
	if err := worker.RegisterKV(mux, "/memory/exec", "memory:",
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
		controller.NewMemoryWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Gin Rummy
	if err := worker.RegisterKV(mux, "/ginrummy/exec", "ginrummy:",
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
		controller.NewGinRummyWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Cribbage
	if err := worker.RegisterKV(mux, "/cribbage/exec", "cribbage:",
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
		controller.NewCribbageWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Canasta
	if err := worker.RegisterKV(mux, "/canasta/exec", "canasta:",
		func() usecase.CanastaInteractorIF {
			config := domain.DefaultCanastaConfig()
			players := []*domain.CanastaPlayer{
				domain.NewCanastaPlayer(true),
				domain.NewCanastaPlayer(false),
			}
			canasta := domain.NewCanasta(domain.NewTrumpCardsWithDecks(2, 4), players, config)
			return usecase.NewCanastaInteractor(canasta, new(presenter.CanastaWebPresenter))
		},
		func(data []byte) (usecase.CanastaInteractorIF, error) {
			return usecase.RestoreCanastaInteractor(data, new(presenter.CanastaWebPresenter))
		},
		controller.NewCanastaWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Golf
	if err := worker.RegisterKV(mux, "/golf/exec", "golf:",
		func() usecase.GolfInteractorIF {
			golf := domain.NewGolf(domain.NewTrumpCards(0))
			return usecase.NewGolfInteractor(golf, new(presenter.GolfWebPresenter))
		},
		func(data []byte) (usecase.GolfInteractorIF, error) {
			return usecase.RestoreGolfInteractor(data, new(presenter.GolfWebPresenter))
		},
		controller.NewGolfWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Clock Solitaire
	if err := worker.RegisterKV(mux, "/clocksolitaire/exec", "clocksolitaire:",
		func() usecase.ClockSolitaireInteractorIF {
			cs := domain.NewClockSolitaire(domain.NewTrumpCards(0))
			return usecase.NewClockSolitaireInteractor(cs, new(presenter.ClockSolitaireWebPresenter))
		},
		func(data []byte) (usecase.ClockSolitaireInteractorIF, error) {
			return usecase.RestoreClockSolitaireInteractor(data, new(presenter.ClockSolitaireWebPresenter))
		},
		controller.NewClockSolitaireWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Canfield
	if err := worker.RegisterKV(mux, "/canfield/exec", "canfield:",
		func() usecase.CanfieldInteractorIF {
			canfield := domain.NewCanfield(domain.NewTrumpCards(0))
			return usecase.NewCanfieldInteractor(canfield, new(presenter.CanfieldWebPresenter))
		},
		func(data []byte) (usecase.CanfieldInteractorIF, error) {
			return usecase.RestoreCanfieldInteractor(data, new(presenter.CanfieldWebPresenter))
		},
		controller.NewCanfieldWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Forty Thieves
	if err := worker.RegisterKV(mux, "/fortythieves/exec", "fortythieves:",
		func() usecase.FortyThievesInteractorIF {
			ft := domain.NewFortyThieves(domain.NewTrumpCardsWithDecks(2, 0))
			return usecase.NewFortyThievesInteractor(ft, new(presenter.FortyThievesWebPresenter))
		},
		func(data []byte) (usecase.FortyThievesInteractorIF, error) {
			return usecase.RestoreFortyThievesInteractor(data, new(presenter.FortyThievesWebPresenter))
		},
		controller.NewFortyThievesWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Yukon
	if err := worker.RegisterKV(mux, "/yukon/exec", "yukon:",
		func() usecase.YukonInteractorIF {
			yukon := domain.NewYukon(domain.NewTrumpCards(0))
			return usecase.NewYukonInteractor(yukon, new(presenter.YukonWebPresenter))
		},
		func(data []byte) (usecase.YukonInteractorIF, error) {
			return usecase.RestoreYukonInteractor(data, new(presenter.YukonWebPresenter))
		},
		controller.NewYukonWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	var handler http.Handler = mux
	if origins := corsmw.ParseOrigins(cloudflare.Getenv("CORS_ALLOWED_ORIGINS")); origins != nil {
		handler = corsmw.Middleware(origins, mux)
	}
	workers.Serve(handler)
}
