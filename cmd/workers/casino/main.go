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

	// BlackJack (KV-backed session)
	blackjackFactory := func() usecase.BlackJackInteractorIF {
		return usecase.NewBlackJackInteractor(
			domain.NewDefaultBlackJack(),
			new(presenter.BlackJackWebPresenter),
		)
	}
	blackjackKV, err := controller.NewKVSessionProvider[usecase.BlackJackInteractorIF](
		"GAME_SESSIONS", "blackjack:",
		func(bi usecase.BlackJackInteractorIF) ([]byte, error) {
			snap, ok := bi.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.BlackJackInteractorIF, error) {
			return usecase.RestoreBlackJackInteractor(data, new(presenter.BlackJackWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	bjc := controller.NewBlackJackWebControllerWithProvider(blackjackKV, blackjackFactory)
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

	// Poker (KV-backed session)
	pokerFactory := func() usecase.PokerInteractorIF {
		config := domain.DefaultPokerConfig()
		players := []*domain.PokerPlayer{
			domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
			domain.NewPokerPlayer(false, domain.PokerStyleConservative),
			domain.NewPokerPlayer(false, domain.PokerStyleAggressive),
			domain.NewPokerPlayer(false, domain.PokerStyleBluffer),
		}
		poker := domain.NewPoker(domain.NewTrumpCards(config.JokerCount), players, config)
		return usecase.NewPokerInteractor(poker, new(presenter.PokerWebPresenter))
	}
	pokerKV, err := controller.NewKVSessionProvider[usecase.PokerInteractorIF](
		"GAME_SESSIONS", "poker:",
		func(pi usecase.PokerInteractorIF) ([]byte, error) {
			snap, ok := pi.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.PokerInteractorIF, error) {
			return usecase.RestorePokerInteractor(data, new(presenter.PokerWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	pkc := controller.NewPokerWebControllerWithProvider(pokerKV, pokerFactory)
	mux.HandleFunc("/poker/exec", pkc.Exec)

	// Texas Hold'em (KV-backed session)
	holdemFactory := func() usecase.HoldemInteractorIF {
		cfg := domain.DefaultHoldemConfig()
		holdem := domain.NewHoldem(domain.NewTrumpCards(0), domain.NewPlayersForTable(cfg.TableSize), cfg)
		return usecase.NewHoldemInteractor(holdem, new(presenter.HoldemWebPresenter))
	}
	holdemKV, err := controller.NewKVSessionProvider[usecase.HoldemInteractorIF](
		"GAME_SESSIONS", "holdem:",
		func(hi usecase.HoldemInteractorIF) ([]byte, error) {
			snap, ok := hi.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.HoldemInteractorIF, error) {
			return usecase.RestoreHoldemInteractor(data, new(presenter.HoldemWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	hmc := controller.NewHoldemWebControllerWithProvider(holdemKV, holdemFactory)
	mux.HandleFunc("/holdem/exec", hmc.Exec)

	// Omaha (KV-backed session)
	omahaFactory := func() usecase.OmahaInteractorIF {
		cfg := domain.DefaultOmahaConfig()
		omaha := domain.NewOmaha(domain.NewTrumpCards(0), domain.NewOmahaPlayersForTable(cfg.TableSize), cfg)
		return usecase.NewOmahaInteractor(omaha, new(presenter.OmahaWebPresenter))
	}
	omahaKV, err := controller.NewKVSessionProvider[usecase.OmahaInteractorIF](
		"GAME_SESSIONS", "omaha:",
		func(oi usecase.OmahaInteractorIF) ([]byte, error) {
			snap, ok := oi.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.OmahaInteractorIF, error) {
			return usecase.RestoreOmahaInteractor(data, new(presenter.OmahaWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	ohc := controller.NewOmahaWebControllerWithProvider(omahaKV, omahaFactory)
	mux.HandleFunc("/omaha/exec", ohc.Exec)

	// Short Deck (KV-backed session)
	shortdeckFactory := func() usecase.ShortDeckInteractorIF {
		cfg := domain.DefaultShortDeckConfig()
		sd := domain.NewShortDeck(domain.NewTrumpCardsShortDeck(), domain.NewShortDeckPlayersForTable(cfg.TableSize), cfg)
		return usecase.NewShortDeckInteractor(sd, new(presenter.ShortDeckWebPresenter))
	}
	shortdeckKV, err := controller.NewKVSessionProvider[usecase.ShortDeckInteractorIF](
		"GAME_SESSIONS", "shortdeck:",
		func(si usecase.ShortDeckInteractorIF) ([]byte, error) {
			snap, ok := si.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.ShortDeckInteractorIF, error) {
			return usecase.RestoreShortDeckInteractor(data, new(presenter.ShortDeckWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	skc := controller.NewShortDeckWebControllerWithProvider(shortdeckKV, shortdeckFactory)
	mux.HandleFunc("/shortdeck/exec", skc.Exec)

	// Indian Poker (KV-backed session)
	indianpokerFactory := func() usecase.IndianPokerInteractorIF {
		cfg := domain.DefaultIndianPokerConfig()
		ip := domain.NewIndianPoker(domain.NewTrumpCards(0), domain.NewIndianPokerPlayers(), cfg)
		return usecase.NewIndianPokerInteractor(ip, new(presenter.IndianPokerWebPresenter))
	}
	indianpokerKV, err := controller.NewKVSessionProvider[usecase.IndianPokerInteractorIF](
		"GAME_SESSIONS", "indianpoker:",
		func(ipi usecase.IndianPokerInteractorIF) ([]byte, error) {
			snap, ok := ipi.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.IndianPokerInteractorIF, error) {
			return usecase.RestoreIndianPokerInteractor(data, new(presenter.IndianPokerWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	ipc := controller.NewIndianPokerWebControllerWithProvider(indianpokerKV, indianpokerFactory)
	mux.HandleFunc("/indianpoker/exec", ipc.Exec)

	// Video Poker (KV-backed session)
	videopokerFactory := func() usecase.VideoPokerInteractorIF {
		return usecase.NewVideoPokerInteractor(
			domain.NewDefaultVideoPoker(),
			new(presenter.VideoPokerWebPresenter),
		)
	}
	videopokerKV, err := controller.NewKVSessionProvider[usecase.VideoPokerInteractorIF](
		"GAME_SESSIONS", "videopoker:",
		func(vi usecase.VideoPokerInteractorIF) ([]byte, error) {
			snap, ok := vi.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.VideoPokerInteractorIF, error) {
			return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	vpc := controller.NewVideoPokerWebControllerWithProvider(videopokerKV, videopokerFactory)
	mux.HandleFunc("/videopoker/exec", vpc.Exec)

	// Deuces Wild (KV-backed session)
	deuceswildFactory := func() usecase.VideoPokerInteractorIF {
		return usecase.NewVideoPokerInteractor(
			domain.NewDeucesWildVideoPoker(),
			new(presenter.VideoPokerWebPresenter),
		)
	}
	deuceswildKV, err := controller.NewKVSessionProvider[usecase.VideoPokerInteractorIF](
		"GAME_SESSIONS", "deuceswild:",
		func(vi usecase.VideoPokerInteractorIF) ([]byte, error) {
			snap, ok := vi.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.VideoPokerInteractorIF, error) {
			return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	dwc := controller.NewVideoPokerWebControllerWithProvider(deuceswildKV, deuceswildFactory)
	mux.HandleFunc("/deuceswild/exec", dwc.Exec)

	// Joker Poker (KV-backed session)
	jokerpokerFactory := func() usecase.VideoPokerInteractorIF {
		return usecase.NewVideoPokerInteractor(
			domain.NewJokerPokerVideoPoker(),
			new(presenter.VideoPokerWebPresenter),
		)
	}
	jokerpokerKV, err := controller.NewKVSessionProvider[usecase.VideoPokerInteractorIF](
		"GAME_SESSIONS", "jokerpoker:",
		func(vi usecase.VideoPokerInteractorIF) ([]byte, error) {
			snap, ok := vi.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.VideoPokerInteractorIF, error) {
			return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	jpc := controller.NewVideoPokerWebControllerWithProvider(jokerpokerKV, jokerpokerFactory)
	mux.HandleFunc("/jokerpoker/exec", jpc.Exec)

	var handler http.Handler = mux
	if origins := corsmw.ParseOrigins(cloudflare.Getenv("CORS_ALLOWED_ORIGINS")); origins != nil {
		handler = corsmw.Middleware(origins, mux)
	}
	workers.Serve(handler)
}
