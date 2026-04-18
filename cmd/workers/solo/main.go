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
	must(worker.RegisterKV(mux, "/klondike/exec", "klondike:",
		func() usecase.KlondikeInteractorIF {
			return usecase.NewKlondikeInteractor(domain.NewDefaultKlondike(), new(presenter.KlondikeWebPresenter))
		},
		func(data []byte) (usecase.KlondikeInteractorIF, error) {
			return usecase.RestoreKlondikeInteractor(data, new(presenter.KlondikeWebPresenter))
		},
		controller.NewKlondikeWebControllerWithProvider,
	))

	// FreeCell
	must(worker.RegisterKV(mux, "/freecell/exec", "freecell:",
		func() usecase.FreeCellInteractorIF {
			return usecase.NewFreeCellInteractor(domain.NewDefaultFreeCell(), new(presenter.FreeCellWebPresenter))
		},
		func(data []byte) (usecase.FreeCellInteractorIF, error) {
			return usecase.RestoreFreeCellInteractor(data, new(presenter.FreeCellWebPresenter))
		},
		controller.NewFreeCellWebControllerWithProvider,
	))

	// Spider
	must(worker.RegisterKV(mux, "/spider/exec", "spider:",
		func() usecase.SpiderInteractorIF {
			return usecase.NewSpiderInteractor(domain.NewDefaultSpider(), new(presenter.SpiderWebPresenter))
		},
		func(data []byte) (usecase.SpiderInteractorIF, error) {
			return usecase.RestoreSpiderInteractor(data, new(presenter.SpiderWebPresenter))
		},
		controller.NewSpiderWebControllerWithProvider,
	))

	// Pyramid
	must(worker.RegisterKV(mux, "/pyramid/exec", "pyramid:",
		func() usecase.PyramidInteractorIF {
			return usecase.NewPyramidInteractor(domain.NewDefaultPyramid(), new(presenter.PyramidWebPresenter))
		},
		func(data []byte) (usecase.PyramidInteractorIF, error) {
			return usecase.RestorePyramidInteractor(data, new(presenter.PyramidWebPresenter))
		},
		controller.NewPyramidWebControllerWithProvider,
	))

	// TriPeaks
	must(worker.RegisterKV(mux, "/tripeaks/exec", "tripeaks:",
		func() usecase.TriPeaksInteractorIF {
			return usecase.NewTriPeaksInteractor(domain.NewDefaultTriPeaks(), new(presenter.TriPeaksWebPresenter))
		},
		func(data []byte) (usecase.TriPeaksInteractorIF, error) {
			return usecase.RestoreTriPeaksInteractor(data, new(presenter.TriPeaksWebPresenter))
		},
		controller.NewTriPeaksWebControllerWithProvider,
	))

	// Memory
	must(worker.RegisterKV(mux, "/memory/exec", "memory:",
		func() usecase.MemoryInteractorIF {
			return usecase.NewMemoryInteractor(domain.NewDefaultMemory(), new(presenter.MemoryWebPresenter))
		},
		func(data []byte) (usecase.MemoryInteractorIF, error) {
			return usecase.RestoreMemoryInteractor(data, new(presenter.MemoryWebPresenter))
		},
		controller.NewMemoryWebControllerWithProvider,
	))

	// Gin Rummy
	must(worker.RegisterKV(mux, "/ginrummy/exec", "ginrummy:",
		func() usecase.GinRummyInteractorIF {
			return usecase.NewGinRummyInteractor(domain.NewDefaultGinRummy(), new(presenter.GinRummyWebPresenter))
		},
		func(data []byte) (usecase.GinRummyInteractorIF, error) {
			return usecase.RestoreGinRummyInteractor(data, new(presenter.GinRummyWebPresenter))
		},
		controller.NewGinRummyWebControllerWithProvider,
	))

	// Cribbage
	must(worker.RegisterKV(mux, "/cribbage/exec", "cribbage:",
		func() usecase.CribbageInteractorIF {
			return usecase.NewCribbageInteractor(domain.NewDefaultCribbage(), new(presenter.CribbageWebPresenter))
		},
		func(data []byte) (usecase.CribbageInteractorIF, error) {
			return usecase.RestoreCribbageInteractor(data, new(presenter.CribbageWebPresenter))
		},
		controller.NewCribbageWebControllerWithProvider,
	))

	// Canasta
	must(worker.RegisterKV(mux, "/canasta/exec", "canasta:",
		func() usecase.CanastaInteractorIF {
			return usecase.NewCanastaInteractor(domain.NewDefaultCanasta(), new(presenter.CanastaWebPresenter))
		},
		func(data []byte) (usecase.CanastaInteractorIF, error) {
			return usecase.RestoreCanastaInteractor(data, new(presenter.CanastaWebPresenter))
		},
		controller.NewCanastaWebControllerWithProvider,
	))

	// Golf
	must(worker.RegisterKV(mux, "/golf/exec", "golf:",
		func() usecase.GolfInteractorIF {
			return usecase.NewGolfInteractor(domain.NewDefaultGolf(), new(presenter.GolfWebPresenter))
		},
		func(data []byte) (usecase.GolfInteractorIF, error) {
			return usecase.RestoreGolfInteractor(data, new(presenter.GolfWebPresenter))
		},
		controller.NewGolfWebControllerWithProvider,
	))

	// Clock Solitaire
	must(worker.RegisterKV(mux, "/clocksolitaire/exec", "clocksolitaire:",
		func() usecase.ClockSolitaireInteractorIF {
			return usecase.NewClockSolitaireInteractor(domain.NewDefaultClockSolitaire(), new(presenter.ClockSolitaireWebPresenter))
		},
		func(data []byte) (usecase.ClockSolitaireInteractorIF, error) {
			return usecase.RestoreClockSolitaireInteractor(data, new(presenter.ClockSolitaireWebPresenter))
		},
		controller.NewClockSolitaireWebControllerWithProvider,
	))

	// Canfield
	must(worker.RegisterKV(mux, "/canfield/exec", "canfield:",
		func() usecase.CanfieldInteractorIF {
			return usecase.NewCanfieldInteractor(domain.NewDefaultCanfield(), new(presenter.CanfieldWebPresenter))
		},
		func(data []byte) (usecase.CanfieldInteractorIF, error) {
			return usecase.RestoreCanfieldInteractor(data, new(presenter.CanfieldWebPresenter))
		},
		controller.NewCanfieldWebControllerWithProvider,
	))

	// Forty Thieves
	must(worker.RegisterKV(mux, "/fortythieves/exec", "fortythieves:",
		func() usecase.FortyThievesInteractorIF {
			return usecase.NewFortyThievesInteractor(domain.NewDefaultFortyThieves(), new(presenter.FortyThievesWebPresenter))
		},
		func(data []byte) (usecase.FortyThievesInteractorIF, error) {
			return usecase.RestoreFortyThievesInteractor(data, new(presenter.FortyThievesWebPresenter))
		},
		controller.NewFortyThievesWebControllerWithProvider,
	))

	// Poker Squares
	must(worker.RegisterKV(mux, "/pokersquares/exec", "pokersquares:",
		func() usecase.PokerSquaresInteractorIF {
			return usecase.NewPokerSquaresInteractor(domain.NewDefaultPokerSquares(), new(presenter.PokerSquaresWebPresenter))
		},
		func(data []byte) (usecase.PokerSquaresInteractorIF, error) {
			return usecase.RestorePokerSquaresInteractor(data, new(presenter.PokerSquaresWebPresenter))
		},
		controller.NewPokerSquaresWebControllerWithProvider,
	))

	// Yukon
	must(worker.RegisterKV(mux, "/yukon/exec", "yukon:",
		func() usecase.YukonInteractorIF {
			return usecase.NewYukonInteractor(domain.NewDefaultYukon(), new(presenter.YukonWebPresenter))
		},
		func(data []byte) (usecase.YukonInteractorIF, error) {
			return usecase.RestoreYukonInteractor(data, new(presenter.YukonWebPresenter))
		},
		controller.NewYukonWebControllerWithProvider,
	))

	// Scorpion
	must(worker.RegisterKV(mux, "/scorpion/exec", "scorpion:",
		func() usecase.ScorpionInteractorIF {
			return usecase.NewScorpionInteractor(domain.NewDefaultScorpion(), new(presenter.ScorpionWebPresenter))
		},
		func(data []byte) (usecase.ScorpionInteractorIF, error) {
			return usecase.RestoreScorpionInteractor(data, new(presenter.ScorpionWebPresenter))
		},
		controller.NewScorpionWebControllerWithProvider,
	))

	var handler http.Handler = mux
	if origins := corsmw.ParseOrigins(cloudflare.Getenv("CORS_ALLOWED_ORIGINS")); origins != nil {
		handler = corsmw.Middleware(origins, mux)
	}
	workers.Serve(handler)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
