//go:build js && wasm

// Package casino binds the Cloudflare Worker KV-backed handlers for the
// 18 table and poker games. A worker main must blank-import this package
// so that the init below runs before games.RegisterCategory is called.
package casino

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
	games.BindWorker("blackjack", games.CategoryCasino, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/blackjack/exec", "blackjack:",
			func() usecase.BlackJackInteractorIF {
				return usecase.NewBlackJackInteractor(domain.NewDefaultBlackJack(), new(presenter.BlackJackWebPresenter))
			},
			func(data []byte) (usecase.BlackJackInteractorIF, error) {
				return usecase.RestoreBlackJackInteractor(data, new(presenter.BlackJackWebPresenter))
			},
			controller.NewBlackJackWebControllerWithProvider,
		)
	})
	games.BindWorker("poker", games.CategoryCasino, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/poker/exec", "poker:",
			func() usecase.PokerInteractorIF {
				return usecase.NewPokerInteractor(domain.NewDefaultPoker(), new(presenter.PokerWebPresenter))
			},
			func(data []byte) (usecase.PokerInteractorIF, error) {
				return usecase.RestorePokerInteractor(data, new(presenter.PokerWebPresenter))
			},
			controller.NewPokerWebControllerWithProvider,
		)
	})
	games.BindWorker("holdem", games.CategoryCasino, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/holdem/exec", "holdem:",
			func() usecase.HoldemInteractorIF {
				return usecase.NewHoldemInteractor(domain.NewDefaultHoldem(), new(presenter.HoldemWebPresenter))
			},
			func(data []byte) (usecase.HoldemInteractorIF, error) {
				return usecase.RestoreHoldemInteractor(data, new(presenter.HoldemWebPresenter))
			},
			controller.NewHoldemWebControllerWithProvider,
		)
	})
	games.BindWorker("omaha", games.CategoryCasino, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/omaha/exec", "omaha:",
			func() usecase.OmahaInteractorIF {
				return usecase.NewOmahaInteractor(domain.NewDefaultOmaha(), new(presenter.OmahaWebPresenter))
			},
			func(data []byte) (usecase.OmahaInteractorIF, error) {
				return usecase.RestoreOmahaInteractor(data, new(presenter.OmahaWebPresenter))
			},
			controller.NewOmahaWebControllerWithProvider,
		)
	})
	games.BindWorker("shortdeck", games.CategoryCasino, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/shortdeck/exec", "shortdeck:",
			func() usecase.ShortDeckInteractorIF {
				return usecase.NewShortDeckInteractor(domain.NewDefaultShortDeck(), new(presenter.ShortDeckWebPresenter))
			},
			func(data []byte) (usecase.ShortDeckInteractorIF, error) {
				return usecase.RestoreShortDeckInteractor(data, new(presenter.ShortDeckWebPresenter))
			},
			controller.NewShortDeckWebControllerWithProvider,
		)
	})
	games.BindWorker("pineapple", games.CategoryCasino, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/pineapple/exec", "pineapple:",
			func() usecase.PineappleInteractorIF {
				return usecase.NewPineappleInteractor(domain.NewDefaultPineapple(), new(presenter.PineappleWebPresenter))
			},
			func(data []byte) (usecase.PineappleInteractorIF, error) {
				return usecase.RestorePineappleInteractor(data, new(presenter.PineappleWebPresenter))
			},
			controller.NewPineappleWebControllerWithProvider,
		)
	})
	games.BindWorker("baccarat", games.CategoryCasino, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/baccarat/exec", "baccarat:",
			func() usecase.BaccaratInteractorIF {
				return usecase.NewBaccaratInteractor(domain.NewDefaultBaccarat(), new(presenter.BaccaratWebPresenter))
			},
			func(data []byte) (usecase.BaccaratInteractorIF, error) {
				return usecase.RestoreBaccaratInteractor(data, new(presenter.BaccaratWebPresenter))
			},
			controller.NewBaccaratWebControllerWithProvider,
		)
	})
	games.BindWorker("indianpoker", games.CategoryCasino, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/indianpoker/exec", "indianpoker:",
			func() usecase.IndianPokerInteractorIF {
				return usecase.NewIndianPokerInteractor(domain.NewDefaultIndianPoker(), new(presenter.IndianPokerWebPresenter))
			},
			func(data []byte) (usecase.IndianPokerInteractorIF, error) {
				return usecase.RestoreIndianPokerInteractor(data, new(presenter.IndianPokerWebPresenter))
			},
			controller.NewIndianPokerWebControllerWithProvider,
		)
	})
	games.BindWorker("videopoker", games.CategoryCasino, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/videopoker/exec", "videopoker:",
			func() usecase.VideoPokerInteractorIF {
				return usecase.NewVideoPokerInteractor(domain.NewDefaultVideoPoker(), new(presenter.VideoPokerWebPresenter))
			},
			func(data []byte) (usecase.VideoPokerInteractorIF, error) {
				return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
			},
			controller.NewVideoPokerWebControllerWithProvider,
		)
	})
	games.BindWorker("deuceswild", games.CategoryCasino, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/deuceswild/exec", "deuceswild:",
			func() usecase.VideoPokerInteractorIF {
				return usecase.NewVideoPokerInteractor(domain.NewDeucesWildVideoPoker(), new(presenter.VideoPokerWebPresenter))
			},
			func(data []byte) (usecase.VideoPokerInteractorIF, error) {
				return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
			},
			controller.NewVideoPokerWebControllerWithProvider,
		)
	})
	games.BindWorker("jokerpoker", games.CategoryCasino, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/jokerpoker/exec", "jokerpoker:",
			func() usecase.VideoPokerInteractorIF {
				return usecase.NewVideoPokerInteractor(domain.NewJokerPokerVideoPoker(), new(presenter.VideoPokerWebPresenter))
			},
			func(data []byte) (usecase.VideoPokerInteractorIF, error) {
				return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
			},
			controller.NewVideoPokerWebControllerWithProvider,
		)
	})
	games.BindWorker("threecard", games.CategoryCasino, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/threecard/exec", "threecard:",
			func() usecase.ThreeCardInteractorIF {
				return usecase.NewThreeCardInteractor(domain.NewDefaultThreeCard(), new(presenter.ThreeCardWebPresenter))
			},
			func(data []byte) (usecase.ThreeCardInteractorIF, error) {
				return usecase.RestoreThreeCardInteractor(data, new(presenter.ThreeCardWebPresenter))
			},
			controller.NewThreeCardWebControllerWithProvider,
		)
	})
	games.BindWorker("sevencardstud", games.CategoryCasino, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/sevencardstud/exec", "sevencardstud:",
			func() usecase.SevenCardStudInteractorIF {
				return usecase.NewSevenCardStudInteractor(domain.NewDefaultSevenCardStud(), new(presenter.SevenCardStudWebPresenter))
			},
			func(data []byte) (usecase.SevenCardStudInteractorIF, error) {
				return usecase.RestoreSevenCardStudInteractor(data, new(presenter.SevenCardStudWebPresenter))
			},
			controller.NewSevenCardStudWebControllerWithProvider,
		)
	})
	games.BindWorker("paigow", games.CategoryCasino, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/paigow/exec", "paigow:",
			func() usecase.PaiGowInteractorIF {
				return usecase.NewPaiGowInteractor(domain.NewDefaultPaiGow(), new(presenter.PaiGowWebPresenter))
			},
			func(data []byte) (usecase.PaiGowInteractorIF, error) {
				return usecase.RestorePaiGowInteractor(data, new(presenter.PaiGowWebPresenter))
			},
			controller.NewPaiGowWebControllerWithProvider,
		)
	})
	games.BindWorker("caribbeanstud", games.CategoryCasino, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/caribbeanstud/exec", "caribbeanstud:",
			func() usecase.CaribbeanStudInteractorIF {
				return usecase.NewCaribbeanStudInteractor(domain.NewDefaultCaribbeanStud(), new(presenter.CaribbeanStudWebPresenter))
			},
			func(data []byte) (usecase.CaribbeanStudInteractorIF, error) {
				return usecase.RestoreCaribbeanStudInteractor(data, new(presenter.CaribbeanStudWebPresenter))
			},
			controller.NewCaribbeanStudWebControllerWithProvider,
		)
	})
	games.BindWorker("letitride", games.CategoryCasino, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/letitride/exec", "letitride:",
			func() usecase.LetItRideInteractorIF {
				return usecase.NewLetItRideInteractor(domain.NewDefaultLetItRide(), new(presenter.LetItRideWebPresenter))
			},
			func(data []byte) (usecase.LetItRideInteractorIF, error) {
				return usecase.RestoreLetItRideInteractor(data, new(presenter.LetItRideWebPresenter))
			},
			controller.NewLetItRideWebControllerWithProvider,
		)
	})
	games.BindWorker("reddog", games.CategoryCasino, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/reddog/exec", "reddog:",
			func() usecase.RedDogInteractorIF {
				return usecase.NewRedDogInteractor(domain.NewDefaultRedDog(), new(presenter.RedDogWebPresenter))
			},
			func(data []byte) (usecase.RedDogInteractorIF, error) {
				return usecase.RestoreRedDogInteractor(data, new(presenter.RedDogWebPresenter))
			},
			controller.NewRedDogWebControllerWithProvider,
		)
	})
	games.BindWorker("razz", games.CategoryCasino, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/razz/exec", "razz:",
			func() usecase.SevenCardStudInteractorIF {
				return usecase.NewSevenCardStudInteractor(domain.NewDefaultRazz(), new(presenter.SevenCardStudWebPresenter))
			},
			func(data []byte) (usecase.SevenCardStudInteractorIF, error) {
				return usecase.RestoreSevenCardStudInteractor(data, new(presenter.SevenCardStudWebPresenter))
			},
			controller.NewSevenCardStudWebControllerWithProvider,
		)
	})
}
