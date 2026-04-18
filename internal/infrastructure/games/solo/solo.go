//go:build js && wasm

// Package solo binds the Cloudflare Worker KV-backed handlers for the
// 16 solitaire and rummy variants. A worker main must blank-import this
// package so the init below runs before games.RegisterCategory is called.
package solo

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/worker"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func init() {
	games.BindWorker("memory", games.CategorySolo, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/memory/exec", "memory:",
			func() usecase.MemoryInteractorIF {
				return usecase.NewMemoryInteractor(domain.NewDefaultMemory(), new(presenter.MemoryWebPresenter))
			},
			func(data []byte) (usecase.MemoryInteractorIF, error) {
				return usecase.RestoreMemoryInteractor(data, new(presenter.MemoryWebPresenter))
			},
			controller.NewMemoryWebControllerWithProvider,
		)
	})
	games.BindWorker("klondike", games.CategorySolo, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/klondike/exec", "klondike:",
			func() usecase.KlondikeInteractorIF {
				return usecase.NewKlondikeInteractor(domain.NewDefaultKlondike(), new(presenter.KlondikeWebPresenter))
			},
			func(data []byte) (usecase.KlondikeInteractorIF, error) {
				return usecase.RestoreKlondikeInteractor(data, new(presenter.KlondikeWebPresenter))
			},
			controller.NewKlondikeWebControllerWithProvider,
		)
	})
	games.BindWorker("freecell", games.CategorySolo, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/freecell/exec", "freecell:",
			func() usecase.FreeCellInteractorIF {
				return usecase.NewFreeCellInteractor(domain.NewDefaultFreeCell(), new(presenter.FreeCellWebPresenter))
			},
			func(data []byte) (usecase.FreeCellInteractorIF, error) {
				return usecase.RestoreFreeCellInteractor(data, new(presenter.FreeCellWebPresenter))
			},
			controller.NewFreeCellWebControllerWithProvider,
		)
	})
	games.BindWorker("ginrummy", games.CategorySolo, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/ginrummy/exec", "ginrummy:",
			func() usecase.GinRummyInteractorIF {
				return usecase.NewGinRummyInteractor(domain.NewDefaultGinRummy(), new(presenter.GinRummyWebPresenter))
			},
			func(data []byte) (usecase.GinRummyInteractorIF, error) {
				return usecase.RestoreGinRummyInteractor(data, new(presenter.GinRummyWebPresenter))
			},
			controller.NewGinRummyWebControllerWithProvider,
		)
	})
	games.BindWorker("canasta", games.CategorySolo, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/canasta/exec", "canasta:",
			func() usecase.CanastaInteractorIF {
				return usecase.NewCanastaInteractor(domain.NewDefaultCanasta(), new(presenter.CanastaWebPresenter))
			},
			func(data []byte) (usecase.CanastaInteractorIF, error) {
				return usecase.RestoreCanastaInteractor(data, new(presenter.CanastaWebPresenter))
			},
			controller.NewCanastaWebControllerWithProvider,
		)
	})
	games.BindWorker("spider", games.CategorySolo, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/spider/exec", "spider:",
			func() usecase.SpiderInteractorIF {
				return usecase.NewSpiderInteractor(domain.NewDefaultSpider(), new(presenter.SpiderWebPresenter))
			},
			func(data []byte) (usecase.SpiderInteractorIF, error) {
				return usecase.RestoreSpiderInteractor(data, new(presenter.SpiderWebPresenter))
			},
			controller.NewSpiderWebControllerWithProvider,
		)
	})
	games.BindWorker("pyramid", games.CategorySolo, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/pyramid/exec", "pyramid:",
			func() usecase.PyramidInteractorIF {
				return usecase.NewPyramidInteractor(domain.NewDefaultPyramid(), new(presenter.PyramidWebPresenter))
			},
			func(data []byte) (usecase.PyramidInteractorIF, error) {
				return usecase.RestorePyramidInteractor(data, new(presenter.PyramidWebPresenter))
			},
			controller.NewPyramidWebControllerWithProvider,
		)
	})
	games.BindWorker("tripeaks", games.CategorySolo, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/tripeaks/exec", "tripeaks:",
			func() usecase.TriPeaksInteractorIF {
				return usecase.NewTriPeaksInteractor(domain.NewDefaultTriPeaks(), new(presenter.TriPeaksWebPresenter))
			},
			func(data []byte) (usecase.TriPeaksInteractorIF, error) {
				return usecase.RestoreTriPeaksInteractor(data, new(presenter.TriPeaksWebPresenter))
			},
			controller.NewTriPeaksWebControllerWithProvider,
		)
	})
	games.BindWorker("cribbage", games.CategorySolo, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/cribbage/exec", "cribbage:",
			func() usecase.CribbageInteractorIF {
				return usecase.NewCribbageInteractor(domain.NewDefaultCribbage(), new(presenter.CribbageWebPresenter))
			},
			func(data []byte) (usecase.CribbageInteractorIF, error) {
				return usecase.RestoreCribbageInteractor(data, new(presenter.CribbageWebPresenter))
			},
			controller.NewCribbageWebControllerWithProvider,
		)
	})
	games.BindWorker("golf", games.CategorySolo, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/golf/exec", "golf:",
			func() usecase.GolfInteractorIF {
				return usecase.NewGolfInteractor(domain.NewDefaultGolf(), new(presenter.GolfWebPresenter))
			},
			func(data []byte) (usecase.GolfInteractorIF, error) {
				return usecase.RestoreGolfInteractor(data, new(presenter.GolfWebPresenter))
			},
			controller.NewGolfWebControllerWithProvider,
		)
	})
	games.BindWorker("clocksolitaire", games.CategorySolo, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/clocksolitaire/exec", "clocksolitaire:",
			func() usecase.ClockSolitaireInteractorIF {
				return usecase.NewClockSolitaireInteractor(domain.NewDefaultClockSolitaire(), new(presenter.ClockSolitaireWebPresenter))
			},
			func(data []byte) (usecase.ClockSolitaireInteractorIF, error) {
				return usecase.RestoreClockSolitaireInteractor(data, new(presenter.ClockSolitaireWebPresenter))
			},
			controller.NewClockSolitaireWebControllerWithProvider,
		)
	})
	games.BindWorker("fortythieves", games.CategorySolo, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/fortythieves/exec", "fortythieves:",
			func() usecase.FortyThievesInteractorIF {
				return usecase.NewFortyThievesInteractor(domain.NewDefaultFortyThieves(), new(presenter.FortyThievesWebPresenter))
			},
			func(data []byte) (usecase.FortyThievesInteractorIF, error) {
				return usecase.RestoreFortyThievesInteractor(data, new(presenter.FortyThievesWebPresenter))
			},
			controller.NewFortyThievesWebControllerWithProvider,
		)
	})
	games.BindWorker("canfield", games.CategorySolo, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/canfield/exec", "canfield:",
			func() usecase.CanfieldInteractorIF {
				return usecase.NewCanfieldInteractor(domain.NewDefaultCanfield(), new(presenter.CanfieldWebPresenter))
			},
			func(data []byte) (usecase.CanfieldInteractorIF, error) {
				return usecase.RestoreCanfieldInteractor(data, new(presenter.CanfieldWebPresenter))
			},
			controller.NewCanfieldWebControllerWithProvider,
		)
	})
	games.BindWorker("yukon", games.CategorySolo, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/yukon/exec", "yukon:",
			func() usecase.YukonInteractorIF {
				return usecase.NewYukonInteractor(domain.NewDefaultYukon(), new(presenter.YukonWebPresenter))
			},
			func(data []byte) (usecase.YukonInteractorIF, error) {
				return usecase.RestoreYukonInteractor(data, new(presenter.YukonWebPresenter))
			},
			controller.NewYukonWebControllerWithProvider,
		)
	})
	games.BindWorker("pokersquares", games.CategorySolo, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/pokersquares/exec", "pokersquares:",
			func() usecase.PokerSquaresInteractorIF {
				return usecase.NewPokerSquaresInteractor(domain.NewDefaultPokerSquares(), new(presenter.PokerSquaresWebPresenter))
			},
			func(data []byte) (usecase.PokerSquaresInteractorIF, error) {
				return usecase.RestorePokerSquaresInteractor(data, new(presenter.PokerSquaresWebPresenter))
			},
			controller.NewPokerSquaresWebControllerWithProvider,
		)
	})
	games.BindWorker("scorpion", games.CategorySolo, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/scorpion/exec", "scorpion:",
			func() usecase.ScorpionInteractorIF {
				return usecase.NewScorpionInteractor(domain.NewDefaultScorpion(), new(presenter.ScorpionWebPresenter))
			},
			func(data []byte) (usecase.ScorpionInteractorIF, error) {
				return usecase.RestoreScorpionInteractor(data, new(presenter.ScorpionWebPresenter))
			},
			controller.NewScorpionWebControllerWithProvider,
		)
	})
}
