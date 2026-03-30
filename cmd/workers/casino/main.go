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

	// BlackJack
	registerKV(mux, "/blackjack/exec", "blackjack:",
		func() usecase.BlackJackInteractorIF {
			return usecase.NewBlackJackInteractor(
				domain.NewDefaultBlackJack(),
				new(presenter.BlackJackWebPresenter),
			)
		},
		func(data []byte) (usecase.BlackJackInteractorIF, error) {
			return usecase.RestoreBlackJackInteractor(data, new(presenter.BlackJackWebPresenter))
		},
		func(p controller.SessionProvider[usecase.BlackJackInteractorIF], f func() usecase.BlackJackInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewBlackJackWebControllerWithProvider(p, f)
		},
	)

	// Baccarat
	registerKV(mux, "/baccarat/exec", "baccarat:",
		func() usecase.BaccaratInteractorIF {
			return usecase.NewBaccaratInteractor(
				domain.NewDefaultBaccarat(),
				new(presenter.BaccaratWebPresenter),
			)
		},
		func(data []byte) (usecase.BaccaratInteractorIF, error) {
			return usecase.RestoreBaccaratInteractor(data, new(presenter.BaccaratWebPresenter))
		},
		func(p controller.SessionProvider[usecase.BaccaratInteractorIF], f func() usecase.BaccaratInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewBaccaratWebControllerWithProvider(p, f)
		},
	)

	// Poker
	registerKV(mux, "/poker/exec", "poker:",
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
		func(p controller.SessionProvider[usecase.PokerInteractorIF], f func() usecase.PokerInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewPokerWebControllerWithProvider(p, f)
		},
	)

	// Texas Hold'em
	registerKV(mux, "/holdem/exec", "holdem:",
		func() usecase.HoldemInteractorIF {
			cfg := domain.DefaultHoldemConfig()
			holdem := domain.NewHoldem(domain.NewTrumpCards(0), domain.NewPlayersForTable(cfg.TableSize), cfg)
			return usecase.NewHoldemInteractor(holdem, new(presenter.HoldemWebPresenter))
		},
		func(data []byte) (usecase.HoldemInteractorIF, error) {
			return usecase.RestoreHoldemInteractor(data, new(presenter.HoldemWebPresenter))
		},
		func(p controller.SessionProvider[usecase.HoldemInteractorIF], f func() usecase.HoldemInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewHoldemWebControllerWithProvider(p, f)
		},
	)

	// Omaha
	registerKV(mux, "/omaha/exec", "omaha:",
		func() usecase.OmahaInteractorIF {
			cfg := domain.DefaultOmahaConfig()
			omaha := domain.NewOmaha(domain.NewTrumpCards(0), domain.NewOmahaPlayersForTable(cfg.TableSize), cfg)
			return usecase.NewOmahaInteractor(omaha, new(presenter.OmahaWebPresenter))
		},
		func(data []byte) (usecase.OmahaInteractorIF, error) {
			return usecase.RestoreOmahaInteractor(data, new(presenter.OmahaWebPresenter))
		},
		func(p controller.SessionProvider[usecase.OmahaInteractorIF], f func() usecase.OmahaInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewOmahaWebControllerWithProvider(p, f)
		},
	)

	// Short Deck
	registerKV(mux, "/shortdeck/exec", "shortdeck:",
		func() usecase.ShortDeckInteractorIF {
			cfg := domain.DefaultShortDeckConfig()
			sd := domain.NewShortDeck(domain.NewTrumpCardsShortDeck(), domain.NewShortDeckPlayersForTable(cfg.TableSize), cfg)
			return usecase.NewShortDeckInteractor(sd, new(presenter.ShortDeckWebPresenter))
		},
		func(data []byte) (usecase.ShortDeckInteractorIF, error) {
			return usecase.RestoreShortDeckInteractor(data, new(presenter.ShortDeckWebPresenter))
		},
		func(p controller.SessionProvider[usecase.ShortDeckInteractorIF], f func() usecase.ShortDeckInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewShortDeckWebControllerWithProvider(p, f)
		},
	)

	// Indian Poker
	registerKV(mux, "/indianpoker/exec", "indianpoker:",
		func() usecase.IndianPokerInteractorIF {
			cfg := domain.DefaultIndianPokerConfig()
			ip := domain.NewIndianPoker(domain.NewTrumpCards(0), domain.NewIndianPokerPlayers(), cfg)
			return usecase.NewIndianPokerInteractor(ip, new(presenter.IndianPokerWebPresenter))
		},
		func(data []byte) (usecase.IndianPokerInteractorIF, error) {
			return usecase.RestoreIndianPokerInteractor(data, new(presenter.IndianPokerWebPresenter))
		},
		func(p controller.SessionProvider[usecase.IndianPokerInteractorIF], f func() usecase.IndianPokerInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewIndianPokerWebControllerWithProvider(p, f)
		},
	)

	// Video Poker
	registerKV(mux, "/videopoker/exec", "videopoker:",
		func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(
				domain.NewDefaultVideoPoker(),
				new(presenter.VideoPokerWebPresenter),
			)
		},
		func(data []byte) (usecase.VideoPokerInteractorIF, error) {
			return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
		},
		func(p controller.SessionProvider[usecase.VideoPokerInteractorIF], f func() usecase.VideoPokerInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewVideoPokerWebControllerWithProvider(p, f)
		},
	)

	// Deuces Wild
	registerKV(mux, "/deuceswild/exec", "deuceswild:",
		func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(
				domain.NewDeucesWildVideoPoker(),
				new(presenter.VideoPokerWebPresenter),
			)
		},
		func(data []byte) (usecase.VideoPokerInteractorIF, error) {
			return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
		},
		func(p controller.SessionProvider[usecase.VideoPokerInteractorIF], f func() usecase.VideoPokerInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewVideoPokerWebControllerWithProvider(p, f)
		},
	)

	// Joker Poker
	registerKV(mux, "/jokerpoker/exec", "jokerpoker:",
		func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(
				domain.NewJokerPokerVideoPoker(),
				new(presenter.VideoPokerWebPresenter),
			)
		},
		func(data []byte) (usecase.VideoPokerInteractorIF, error) {
			return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
		},
		func(p controller.SessionProvider[usecase.VideoPokerInteractorIF], f func() usecase.VideoPokerInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewVideoPokerWebControllerWithProvider(p, f)
		},
	)

	// Three Card Poker
	registerKV(mux, "/threecard/exec", "threecard:",
		func() usecase.ThreeCardInteractorIF {
			return usecase.NewThreeCardInteractor(
				domain.NewDefaultThreeCard(),
				new(presenter.ThreeCardWebPresenter),
			)
		},
		func(data []byte) (usecase.ThreeCardInteractorIF, error) {
			return usecase.RestoreThreeCardInteractor(data, new(presenter.ThreeCardWebPresenter))
		},
		func(p controller.SessionProvider[usecase.ThreeCardInteractorIF], f func() usecase.ThreeCardInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewThreeCardWebControllerWithProvider(p, f)
		},
	)

	// Pineapple Poker
	registerKV(mux, "/pineapple/exec", "pineapple:",
		func() usecase.PineappleInteractorIF {
			cfg := domain.DefaultPineappleConfig()
			pineapple := domain.NewPineapple(domain.NewTrumpCards(0), domain.NewPineapplePlayersForTable(cfg.TableSize), cfg)
			return usecase.NewPineappleInteractor(pineapple, new(presenter.PineappleWebPresenter))
		},
		func(data []byte) (usecase.PineappleInteractorIF, error) {
			return usecase.RestorePineappleInteractor(data, new(presenter.PineappleWebPresenter))
		},
		func(p controller.SessionProvider[usecase.PineappleInteractorIF], f func() usecase.PineappleInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewPineappleWebControllerWithProvider(p, f)
		},
	)

	var handler http.Handler = mux
	if origins := corsmw.ParseOrigins(cloudflare.Getenv("CORS_ALLOWED_ORIGINS")); origins != nil {
		handler = corsmw.Middleware(origins, mux)
	}
	workers.Serve(handler)
}
