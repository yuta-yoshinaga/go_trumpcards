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

	// BlackJack
	bjc := controller.NewBlackJackWebController(func() usecase.BlackJackInteractorIF {
		return usecase.NewBlackJackInteractor(
			domain.NewDefaultBlackJack(),
			new(presenter.BlackJackWebPresenter),
		)
	})
	mux.HandleFunc("/blackjack/exec", bjc.Exec)

	// Baccarat (KV-backed session)
	baccaratFactory := func() usecase.BaccaratInteractorIF {
		return usecase.NewBaccaratInteractor(
			domain.NewDefaultBaccarat(),
			new(presenter.BaccaratWebPresenter),
		)
	}
	baccaratKV, err := controller.NewKVSessionProvider[usecase.BaccaratInteractorIF](
		"GAME_SESSIONS", "baccarat:",
		func(bi usecase.BaccaratInteractorIF) ([]byte, error) {
			snap, ok := bi.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.BaccaratInteractorIF, error) {
			return usecase.RestoreBaccaratInteractor(data, new(presenter.BaccaratWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	bcc := controller.NewBaccaratWebControllerWithProvider(baccaratKV, baccaratFactory)
	mux.HandleFunc("/baccarat/exec", bcc.Exec)

	// Poker
	pkc := controller.NewPokerWebController(func() usecase.PokerInteractorIF {
		config := domain.DefaultPokerConfig()
		players := []*domain.PokerPlayer{
			domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
			domain.NewPokerPlayer(false, domain.PokerStyleConservative),
			domain.NewPokerPlayer(false, domain.PokerStyleAggressive),
			domain.NewPokerPlayer(false, domain.PokerStyleBluffer),
		}
		poker := domain.NewPoker(domain.NewTrumpCards(config.JokerCount), players, config)
		return usecase.NewPokerInteractor(poker, new(presenter.PokerWebPresenter))
	})
	mux.HandleFunc("/poker/exec", pkc.Exec)

	// Texas Hold'em
	hmc := controller.NewHoldemWebController(func() usecase.HoldemInteractorIF {
		cfg := domain.DefaultHoldemConfig()
		holdem := domain.NewHoldem(domain.NewTrumpCards(0), domain.NewPlayersForTable(cfg.TableSize), cfg)
		return usecase.NewHoldemInteractor(holdem, new(presenter.HoldemWebPresenter))
	})
	mux.HandleFunc("/holdem/exec", hmc.Exec)

	// Omaha
	ohc := controller.NewOmahaWebController(func() usecase.OmahaInteractorIF {
		cfg := domain.DefaultOmahaConfig()
		omaha := domain.NewOmaha(domain.NewTrumpCards(0), domain.NewOmahaPlayersForTable(cfg.TableSize), cfg)
		return usecase.NewOmahaInteractor(omaha, new(presenter.OmahaWebPresenter))
	})
	mux.HandleFunc("/omaha/exec", ohc.Exec)

	// Short Deck
	skc := controller.NewShortDeckWebController(func() usecase.ShortDeckInteractorIF {
		cfg := domain.DefaultShortDeckConfig()
		sd := domain.NewShortDeck(domain.NewTrumpCardsShortDeck(), domain.NewShortDeckPlayersForTable(cfg.TableSize), cfg)
		return usecase.NewShortDeckInteractor(sd, new(presenter.ShortDeckWebPresenter))
	})
	mux.HandleFunc("/shortdeck/exec", skc.Exec)

	// Indian Poker
	ipc := controller.NewIndianPokerWebController(func() usecase.IndianPokerInteractorIF {
		cfg := domain.DefaultIndianPokerConfig()
		ip := domain.NewIndianPoker(domain.NewTrumpCards(0), domain.NewIndianPokerPlayers(), cfg)
		return usecase.NewIndianPokerInteractor(ip, new(presenter.IndianPokerWebPresenter))
	})
	mux.HandleFunc("/indianpoker/exec", ipc.Exec)

	// Video Poker
	vpc := controller.NewVideoPokerWebController(func() usecase.VideoPokerInteractorIF {
		return usecase.NewVideoPokerInteractor(
			domain.NewDefaultVideoPoker(),
			new(presenter.VideoPokerWebPresenter),
		)
	})
	mux.HandleFunc("/videopoker/exec", vpc.Exec)

	// Deuces Wild
	dwc := controller.NewVideoPokerWebController(func() usecase.VideoPokerInteractorIF {
		return usecase.NewVideoPokerInteractor(
			domain.NewDeucesWildVideoPoker(),
			new(presenter.VideoPokerWebPresenter),
		)
	})
	mux.HandleFunc("/deuceswild/exec", dwc.Exec)

	// Joker Poker
	jpc := controller.NewVideoPokerWebController(func() usecase.VideoPokerInteractorIF {
		return usecase.NewVideoPokerInteractor(
			domain.NewJokerPokerVideoPoker(),
			new(presenter.VideoPokerWebPresenter),
		)
	})
	mux.HandleFunc("/jokerpoker/exec", jpc.Exec)

	var handler http.Handler = mux
	if origins := corsmw.ParseOrigins(cloudflare.Getenv("CORS_ALLOWED_ORIGINS")); origins != nil {
		handler = corsmw.Middleware(origins, mux)
	}
	workers.Serve(handler)
}
