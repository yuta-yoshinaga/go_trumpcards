package main

import (
	"net/http"

	"github.com/syumai/workers"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func main() {
	mux := http.NewServeMux()

	// BlackJack
	bjc := controller.NewBlackJackWebController(func() usecase.BlackJackInteractorIF {
		return usecase.NewBlackJackInteractor(
			domain.NewDefaultBlackJack(),
			new(presenter.BlackJackWebPresenter),
		)
	})
	mux.HandleFunc("POST /blackjack/exec", bjc.Exec)

	// Baccarat
	bcc := controller.NewBaccaratWebController(func() usecase.BaccaratInteractorIF {
		baccarat := domain.NewDefaultBaccarat()
		return usecase.NewBaccaratInteractor(baccarat, new(presenter.BaccaratWebPresenter))
	})
	mux.HandleFunc("POST /baccarat/exec", bcc.Exec)

	workers.Serve(mux)
}
