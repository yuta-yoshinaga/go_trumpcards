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

	// BlackJack
	must(worker.RegisterKV(mux, "/blackjack/exec", "blackjack:",
		func() usecase.BlackJackInteractorIF {
			return usecase.NewBlackJackInteractor(
				domain.NewDefaultBlackJack(),
				new(presenter.BlackJackWebPresenter),
			)
		},
		func(data []byte) (usecase.BlackJackInteractorIF, error) {
			return usecase.RestoreBlackJackInteractor(data, new(presenter.BlackJackWebPresenter))
		},
		controller.NewBlackJackWebControllerWithProvider,
	))

	// Baccarat
	must(worker.RegisterKV(mux, "/baccarat/exec", "baccarat:",
		func() usecase.BaccaratInteractorIF {
			return usecase.NewBaccaratInteractor(
				domain.NewDefaultBaccarat(),
				new(presenter.BaccaratWebPresenter),
			)
		},
		func(data []byte) (usecase.BaccaratInteractorIF, error) {
			return usecase.RestoreBaccaratInteractor(data, new(presenter.BaccaratWebPresenter))
		},
		controller.NewBaccaratWebControllerWithProvider,
	))

	// Poker
	must(worker.RegisterKV(mux, "/poker/exec", "poker:",
		func() usecase.PokerInteractorIF {
			return usecase.NewPokerInteractor(domain.NewDefaultPoker(), new(presenter.PokerWebPresenter))
		},
		func(data []byte) (usecase.PokerInteractorIF, error) {
			return usecase.RestorePokerInteractor(data, new(presenter.PokerWebPresenter))
		},
		controller.NewPokerWebControllerWithProvider,
	))

	// Texas Hold'em
	must(worker.RegisterKV(mux, "/holdem/exec", "holdem:",
		func() usecase.HoldemInteractorIF {
			return usecase.NewHoldemInteractor(domain.NewDefaultHoldem(), new(presenter.HoldemWebPresenter))
		},
		func(data []byte) (usecase.HoldemInteractorIF, error) {
			return usecase.RestoreHoldemInteractor(data, new(presenter.HoldemWebPresenter))
		},
		controller.NewHoldemWebControllerWithProvider,
	))

	// Omaha
	must(worker.RegisterKV(mux, "/omaha/exec", "omaha:",
		func() usecase.OmahaInteractorIF {
			return usecase.NewOmahaInteractor(domain.NewDefaultOmaha(), new(presenter.OmahaWebPresenter))
		},
		func(data []byte) (usecase.OmahaInteractorIF, error) {
			return usecase.RestoreOmahaInteractor(data, new(presenter.OmahaWebPresenter))
		},
		controller.NewOmahaWebControllerWithProvider,
	))

	// Short Deck
	must(worker.RegisterKV(mux, "/shortdeck/exec", "shortdeck:",
		func() usecase.ShortDeckInteractorIF {
			return usecase.NewShortDeckInteractor(domain.NewDefaultShortDeck(), new(presenter.ShortDeckWebPresenter))
		},
		func(data []byte) (usecase.ShortDeckInteractorIF, error) {
			return usecase.RestoreShortDeckInteractor(data, new(presenter.ShortDeckWebPresenter))
		},
		controller.NewShortDeckWebControllerWithProvider,
	))

	// Indian Poker
	must(worker.RegisterKV(mux, "/indianpoker/exec", "indianpoker:",
		func() usecase.IndianPokerInteractorIF {
			return usecase.NewIndianPokerInteractor(domain.NewDefaultIndianPoker(), new(presenter.IndianPokerWebPresenter))
		},
		func(data []byte) (usecase.IndianPokerInteractorIF, error) {
			return usecase.RestoreIndianPokerInteractor(data, new(presenter.IndianPokerWebPresenter))
		},
		controller.NewIndianPokerWebControllerWithProvider,
	))

	// Video Poker
	must(worker.RegisterKV(mux, "/videopoker/exec", "videopoker:",
		func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(
				domain.NewDefaultVideoPoker(),
				new(presenter.VideoPokerWebPresenter),
			)
		},
		func(data []byte) (usecase.VideoPokerInteractorIF, error) {
			return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
		},
		controller.NewVideoPokerWebControllerWithProvider,
	))

	// Deuces Wild
	must(worker.RegisterKV(mux, "/deuceswild/exec", "deuceswild:",
		func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(
				domain.NewDeucesWildVideoPoker(),
				new(presenter.VideoPokerWebPresenter),
			)
		},
		func(data []byte) (usecase.VideoPokerInteractorIF, error) {
			return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
		},
		controller.NewVideoPokerWebControllerWithProvider,
	))

	// Joker Poker
	must(worker.RegisterKV(mux, "/jokerpoker/exec", "jokerpoker:",
		func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(
				domain.NewJokerPokerVideoPoker(),
				new(presenter.VideoPokerWebPresenter),
			)
		},
		func(data []byte) (usecase.VideoPokerInteractorIF, error) {
			return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
		},
		controller.NewVideoPokerWebControllerWithProvider,
	))

	// Three Card Poker
	must(worker.RegisterKV(mux, "/threecard/exec", "threecard:",
		func() usecase.ThreeCardInteractorIF {
			return usecase.NewThreeCardInteractor(
				domain.NewDefaultThreeCard(),
				new(presenter.ThreeCardWebPresenter),
			)
		},
		func(data []byte) (usecase.ThreeCardInteractorIF, error) {
			return usecase.RestoreThreeCardInteractor(data, new(presenter.ThreeCardWebPresenter))
		},
		controller.NewThreeCardWebControllerWithProvider,
	))

	// Pineapple Poker
	must(worker.RegisterKV(mux, "/pineapple/exec", "pineapple:",
		func() usecase.PineappleInteractorIF {
			return usecase.NewPineappleInteractor(domain.NewDefaultPineapple(), new(presenter.PineappleWebPresenter))
		},
		func(data []byte) (usecase.PineappleInteractorIF, error) {
			return usecase.RestorePineappleInteractor(data, new(presenter.PineappleWebPresenter))
		},
		controller.NewPineappleWebControllerWithProvider,
	))

	// Seven Card Stud
	must(worker.RegisterKV(mux, "/sevencardstud/exec", "sevencardstud:",
		func() usecase.SevenCardStudInteractorIF {
			return usecase.NewSevenCardStudInteractor(domain.NewDefaultSevenCardStud(), new(presenter.SevenCardStudWebPresenter))
		},
		func(data []byte) (usecase.SevenCardStudInteractorIF, error) {
			return usecase.RestoreSevenCardStudInteractor(data, new(presenter.SevenCardStudWebPresenter))
		},
		controller.NewSevenCardStudWebControllerWithProvider,
	))

	// Razz
	must(worker.RegisterKV(mux, "/razz/exec", "razz:",
		func() usecase.SevenCardStudInteractorIF {
			return usecase.NewSevenCardStudInteractor(domain.NewDefaultRazz(), new(presenter.SevenCardStudWebPresenter))
		},
		func(data []byte) (usecase.SevenCardStudInteractorIF, error) {
			return usecase.RestoreSevenCardStudInteractor(data, new(presenter.SevenCardStudWebPresenter))
		},
		controller.NewSevenCardStudWebControllerWithProvider,
	))

	// Pai Gow Poker
	must(worker.RegisterKV(mux, "/paigow/exec", "paigow:",
		func() usecase.PaiGowInteractorIF {
			return usecase.NewPaiGowInteractor(
				domain.NewDefaultPaiGow(),
				new(presenter.PaiGowWebPresenter),
			)
		},
		func(data []byte) (usecase.PaiGowInteractorIF, error) {
			return usecase.RestorePaiGowInteractor(data, new(presenter.PaiGowWebPresenter))
		},
		controller.NewPaiGowWebControllerWithProvider,
	))

	// Caribbean Stud Poker
	must(worker.RegisterKV(mux, "/caribbeanstud/exec", "caribbeanstud:",
		func() usecase.CaribbeanStudInteractorIF {
			return usecase.NewCaribbeanStudInteractor(
				domain.NewDefaultCaribbeanStud(),
				new(presenter.CaribbeanStudWebPresenter),
			)
		},
		func(data []byte) (usecase.CaribbeanStudInteractorIF, error) {
			return usecase.RestoreCaribbeanStudInteractor(data, new(presenter.CaribbeanStudWebPresenter))
		},
		controller.NewCaribbeanStudWebControllerWithProvider,
	))

	// Let It Ride
	must(worker.RegisterKV(mux, "/letitride/exec", "letitride:",
		func() usecase.LetItRideInteractorIF {
			return usecase.NewLetItRideInteractor(
				domain.NewDefaultLetItRide(),
				new(presenter.LetItRideWebPresenter),
			)
		},
		func(data []byte) (usecase.LetItRideInteractorIF, error) {
			return usecase.RestoreLetItRideInteractor(data, new(presenter.LetItRideWebPresenter))
		},
		controller.NewLetItRideWebControllerWithProvider,
	))

	// Red Dog
	must(worker.RegisterKV(mux, "/reddog/exec", "reddog:",
		func() usecase.RedDogInteractorIF {
			return usecase.NewRedDogInteractor(
				domain.NewDefaultRedDog(),
				new(presenter.RedDogWebPresenter),
			)
		},
		func(data []byte) (usecase.RedDogInteractorIF, error) {
			return usecase.RestoreRedDogInteractor(data, new(presenter.RedDogWebPresenter))
		},
		controller.NewRedDogWebControllerWithProvider,
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
