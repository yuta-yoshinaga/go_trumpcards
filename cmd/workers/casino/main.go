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
	if err := worker.RegisterKV(mux, "/blackjack/exec", "blackjack:",
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
	); err != nil {
		log.Fatal(err)
	}

	// Baccarat
	if err := worker.RegisterKV(mux, "/baccarat/exec", "baccarat:",
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
	); err != nil {
		log.Fatal(err)
	}

	// Poker
	if err := worker.RegisterKV(mux, "/poker/exec", "poker:",
		func() usecase.PokerInteractorIF {
			config := domain.DefaultPokerConfig()
			players := []*domain.PokerPlayer{
				domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
				domain.NewPokerPlayer(false, domain.PokerStyleConservative),
				domain.NewPokerPlayer(false, domain.PokerStyleAggressive),
				domain.NewPokerPlayer(false, domain.PokerStyleBluffer),
			}
			poker := domain.NewPoker(domain.NewTrumpCards(config.JokerCount), players, config)
			return usecase.NewPokerInteractor(poker, new(presenter.PokerWebPresenter))
		},
		func(data []byte) (usecase.PokerInteractorIF, error) {
			return usecase.RestorePokerInteractor(data, new(presenter.PokerWebPresenter))
		},
		controller.NewPokerWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Texas Hold'em
	if err := worker.RegisterKV(mux, "/holdem/exec", "holdem:",
		func() usecase.HoldemInteractorIF {
			cfg := domain.DefaultHoldemConfig()
			holdem := domain.NewHoldem(domain.NewTrumpCards(0), domain.NewPlayersForTable(cfg.TableSize), cfg)
			return usecase.NewHoldemInteractor(holdem, new(presenter.HoldemWebPresenter))
		},
		func(data []byte) (usecase.HoldemInteractorIF, error) {
			return usecase.RestoreHoldemInteractor(data, new(presenter.HoldemWebPresenter))
		},
		controller.NewHoldemWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Omaha
	if err := worker.RegisterKV(mux, "/omaha/exec", "omaha:",
		func() usecase.OmahaInteractorIF {
			cfg := domain.DefaultOmahaConfig()
			omaha := domain.NewOmaha(domain.NewTrumpCards(0), domain.NewOmahaPlayersForTable(cfg.TableSize), cfg)
			return usecase.NewOmahaInteractor(omaha, new(presenter.OmahaWebPresenter))
		},
		func(data []byte) (usecase.OmahaInteractorIF, error) {
			return usecase.RestoreOmahaInteractor(data, new(presenter.OmahaWebPresenter))
		},
		controller.NewOmahaWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Short Deck
	if err := worker.RegisterKV(mux, "/shortdeck/exec", "shortdeck:",
		func() usecase.ShortDeckInteractorIF {
			cfg := domain.DefaultShortDeckConfig()
			sd := domain.NewShortDeck(domain.NewTrumpCardsShortDeck(), domain.NewShortDeckPlayersForTable(cfg.TableSize), cfg)
			return usecase.NewShortDeckInteractor(sd, new(presenter.ShortDeckWebPresenter))
		},
		func(data []byte) (usecase.ShortDeckInteractorIF, error) {
			return usecase.RestoreShortDeckInteractor(data, new(presenter.ShortDeckWebPresenter))
		},
		controller.NewShortDeckWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Indian Poker
	if err := worker.RegisterKV(mux, "/indianpoker/exec", "indianpoker:",
		func() usecase.IndianPokerInteractorIF {
			cfg := domain.DefaultIndianPokerConfig()
			ip := domain.NewIndianPoker(domain.NewTrumpCards(0), domain.NewIndianPokerPlayers(), cfg)
			return usecase.NewIndianPokerInteractor(ip, new(presenter.IndianPokerWebPresenter))
		},
		func(data []byte) (usecase.IndianPokerInteractorIF, error) {
			return usecase.RestoreIndianPokerInteractor(data, new(presenter.IndianPokerWebPresenter))
		},
		controller.NewIndianPokerWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Video Poker
	if err := worker.RegisterKV(mux, "/videopoker/exec", "videopoker:",
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
	); err != nil {
		log.Fatal(err)
	}

	// Deuces Wild
	if err := worker.RegisterKV(mux, "/deuceswild/exec", "deuceswild:",
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
	); err != nil {
		log.Fatal(err)
	}

	// Joker Poker
	if err := worker.RegisterKV(mux, "/jokerpoker/exec", "jokerpoker:",
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
	); err != nil {
		log.Fatal(err)
	}

	// Three Card Poker
	if err := worker.RegisterKV(mux, "/threecard/exec", "threecard:",
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
	); err != nil {
		log.Fatal(err)
	}

	// Pineapple Poker
	if err := worker.RegisterKV(mux, "/pineapple/exec", "pineapple:",
		func() usecase.PineappleInteractorIF {
			cfg := domain.DefaultPineappleConfig()
			pineapple := domain.NewPineapple(domain.NewTrumpCards(0), domain.NewPineapplePlayersForTable(cfg.TableSize), cfg)
			return usecase.NewPineappleInteractor(pineapple, new(presenter.PineappleWebPresenter))
		},
		func(data []byte) (usecase.PineappleInteractorIF, error) {
			return usecase.RestorePineappleInteractor(data, new(presenter.PineappleWebPresenter))
		},
		controller.NewPineappleWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Seven Card Stud
	if err := worker.RegisterKV(mux, "/sevencardstud/exec", "sevencardstud:",
		func() usecase.SevenCardStudInteractorIF {
			cfg := domain.DefaultSevenCardStudConfig()
			scs := domain.NewSevenCardStud(domain.NewTrumpCards(0), domain.NewSevenCardStudPlayersForTable(cfg.TableSize), cfg)
			return usecase.NewSevenCardStudInteractor(scs, new(presenter.SevenCardStudWebPresenter))
		},
		func(data []byte) (usecase.SevenCardStudInteractorIF, error) {
			return usecase.RestoreSevenCardStudInteractor(data, new(presenter.SevenCardStudWebPresenter))
		},
		controller.NewSevenCardStudWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Pai Gow Poker
	if err := worker.RegisterKV(mux, "/paigow/exec", "paigow:",
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
	); err != nil {
		log.Fatal(err)
	}

	// Caribbean Stud Poker
	if err := worker.RegisterKV(mux, "/caribbeanstud/exec", "caribbeanstud:",
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
	); err != nil {
		log.Fatal(err)
	}

	// Let It Ride
	if err := worker.RegisterKV(mux, "/letitride/exec", "letitride:",
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
	); err != nil {
		log.Fatal(err)
	}

	var handler http.Handler = mux
	if origins := corsmw.ParseOrigins(cloudflare.Getenv("CORS_ALLOWED_ORIGINS")); origins != nil {
		handler = corsmw.Middleware(origins, mux)
	}
	workers.Serve(handler)
}
