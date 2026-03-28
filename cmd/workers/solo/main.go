//go:build js && wasm

package main

import (
	"net/http"

	"github.com/syumai/workers/cloudflare"

	"github.com/syumai/workers"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	corsmw "github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/cors"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func main() {
	mux := http.NewServeMux()

	// Klondike
	klc := controller.NewKlondikeWebController(func() usecase.KlondikeInteractorIF {
		klondike := domain.NewKlondike(domain.NewTrumpCards(0))
		return usecase.NewKlondikeInteractor(klondike, new(presenter.KlondikeWebPresenter))
	})
	mux.HandleFunc("/klondike/exec", klc.Exec)

	// FreeCell
	fcc := controller.NewFreeCellWebController(func() usecase.FreeCellInteractorIF {
		freeCell := domain.NewFreeCell(domain.NewTrumpCards(0))
		return usecase.NewFreeCellInteractor(freeCell, new(presenter.FreeCellWebPresenter))
	})
	mux.HandleFunc("/freecell/exec", fcc.Exec)

	// Spider
	sdc := controller.NewSpiderWebController(func() usecase.SpiderInteractorIF {
		spider := domain.NewSpider(domain.NewTrumpCardsWithSuits(domain.SpiderTotalCards, []int{domain.CardDesignSpade}))
		return usecase.NewSpiderInteractor(spider, new(presenter.SpiderWebPresenter))
	})
	mux.HandleFunc("/spider/exec", sdc.Exec)

	// Pyramid
	pyc := controller.NewPyramidWebController(func() usecase.PyramidInteractorIF {
		pyramid := domain.NewPyramid(domain.NewTrumpCards(0))
		return usecase.NewPyramidInteractor(pyramid, new(presenter.PyramidWebPresenter))
	})
	mux.HandleFunc("/pyramid/exec", pyc.Exec)

	// Memory
	myc := controller.NewMemoryWebController(func() usecase.MemoryInteractorIF {
		config := domain.DefaultMemoryConfig()
		players := []*domain.MemoryPlayer{
			domain.NewMemoryPlayer(true),
			domain.NewMemoryPlayer(false),
			domain.NewMemoryPlayer(false),
			domain.NewMemoryPlayer(false),
		}
		memory := domain.NewMemory(domain.NewTrumpCards(0), players, config)
		return usecase.NewMemoryInteractor(memory, new(presenter.MemoryWebPresenter))
	})
	mux.HandleFunc("/memory/exec", myc.Exec)

	// Gin Rummy
	grc := controller.NewGinRummyWebController(func() usecase.GinRummyInteractorIF {
		config := domain.DefaultGinRummyConfig()
		players := []*domain.GinRummyPlayer{
			domain.NewGinRummyPlayer(true),
			domain.NewGinRummyPlayer(false),
		}
		gr := domain.NewGinRummy(domain.NewTrumpCards(0), players, config)
		return usecase.NewGinRummyInteractor(gr, new(presenter.GinRummyWebPresenter))
	})
	mux.HandleFunc("/ginrummy/exec", grc.Exec)

	// Cribbage
	cbc := controller.NewCribbageWebController(func() usecase.CribbageInteractorIF {
		config := domain.DefaultCribbageConfig()
		players := []*domain.CribbagePlayer{
			domain.NewCribbagePlayer(true),
			domain.NewCribbagePlayer(false),
		}
		cribbage := domain.NewCribbage(domain.NewTrumpCards(0), players, config)
		return usecase.NewCribbageInteractor(cribbage, new(presenter.CribbageWebPresenter))
	})
	mux.HandleFunc("/cribbage/exec", cbc.Exec)

	var handler http.Handler = mux
	if origins := corsmw.ParseOrigins(cloudflare.Getenv("CORS_ALLOWED_ORIGINS")); origins != nil {
		handler = corsmw.Middleware(origins, mux)
	}
	workers.Serve(handler)
}
