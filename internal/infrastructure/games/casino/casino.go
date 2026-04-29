//go:build js && wasm

// Package casino binds the Cloudflare Worker KV-backed handlers for the
// 18 table and poker games. A worker main must blank-import this package
// so that the init below runs before games.RegisterCategory is called.
package casino

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func init() {
	games.RegisterKVGame("blackjack", games.CategoryCasino,
		func() usecase.BlackJackInteractorIF {
			return usecase.NewBlackJackInteractor(domain.NewDefaultBlackJack(), new(presenter.BlackJackWebPresenter))
		},
		func(data []byte) (usecase.BlackJackInteractorIF, error) {
			return usecase.RestoreBlackJackInteractor(data, new(presenter.BlackJackWebPresenter))
		},
		controller.NewBlackJackWebControllerWithProvider)
	games.RegisterKVGame("poker", games.CategoryCasino,
		func() usecase.PokerInteractorIF {
			return usecase.NewPokerInteractor(domain.NewDefaultPoker(), new(presenter.PokerWebPresenter))
		},
		func(data []byte) (usecase.PokerInteractorIF, error) {
			return usecase.RestorePokerInteractor(data, new(presenter.PokerWebPresenter))
		},
		controller.NewPokerWebControllerWithProvider)
	games.RegisterKVGame("holdem", games.CategoryCasino,
		func() usecase.HoldemInteractorIF {
			return usecase.NewHoldemInteractor(domain.NewDefaultHoldem(), new(presenter.HoldemWebPresenter))
		},
		func(data []byte) (usecase.HoldemInteractorIF, error) {
			return usecase.RestoreHoldemInteractor(data, new(presenter.HoldemWebPresenter))
		},
		controller.NewHoldemWebControllerWithProvider)
	games.RegisterKVGame("omaha", games.CategoryCasino,
		func() usecase.OmahaInteractorIF {
			return usecase.NewOmahaInteractor(domain.NewDefaultOmaha(), new(presenter.OmahaWebPresenter))
		},
		func(data []byte) (usecase.OmahaInteractorIF, error) {
			return usecase.RestoreOmahaInteractor(data, new(presenter.OmahaWebPresenter))
		},
		controller.NewOmahaWebControllerWithProvider)
	games.RegisterKVGame("shortdeck", games.CategoryCasino,
		func() usecase.ShortDeckInteractorIF {
			return usecase.NewShortDeckInteractor(domain.NewDefaultShortDeck(), new(presenter.ShortDeckWebPresenter))
		},
		func(data []byte) (usecase.ShortDeckInteractorIF, error) {
			return usecase.RestoreShortDeckInteractor(data, new(presenter.ShortDeckWebPresenter))
		},
		controller.NewShortDeckWebControllerWithProvider)
	games.RegisterKVGame("pineapple", games.CategoryCasino,
		func() usecase.PineappleInteractorIF {
			return usecase.NewPineappleInteractor(domain.NewDefaultPineapple(), new(presenter.PineappleWebPresenter))
		},
		func(data []byte) (usecase.PineappleInteractorIF, error) {
			return usecase.RestorePineappleInteractor(data, new(presenter.PineappleWebPresenter))
		},
		controller.NewPineappleWebControllerWithProvider)
	games.RegisterKVGame("crazypineapple", games.CategoryCasino,
		func() usecase.PineappleInteractorIF {
			return usecase.NewPineappleInteractor(domain.NewDefaultCrazyPineapple(), new(presenter.PineappleWebPresenter))
		},
		func(data []byte) (usecase.PineappleInteractorIF, error) {
			return usecase.RestorePineappleInteractor(data, new(presenter.PineappleWebPresenter))
		},
		controller.NewPineappleWebControllerWithProvider)
	games.RegisterKVGame("baccarat", games.CategoryCasino,
		func() usecase.BaccaratInteractorIF {
			return usecase.NewBaccaratInteractor(domain.NewDefaultBaccarat(), new(presenter.BaccaratWebPresenter))
		},
		func(data []byte) (usecase.BaccaratInteractorIF, error) {
			return usecase.RestoreBaccaratInteractor(data, new(presenter.BaccaratWebPresenter))
		},
		controller.NewBaccaratWebControllerWithProvider)
	games.RegisterKVGame("indianpoker", games.CategoryCasino,
		func() usecase.IndianPokerInteractorIF {
			return usecase.NewIndianPokerInteractor(domain.NewDefaultIndianPoker(), new(presenter.IndianPokerWebPresenter))
		},
		func(data []byte) (usecase.IndianPokerInteractorIF, error) {
			return usecase.RestoreIndianPokerInteractor(data, new(presenter.IndianPokerWebPresenter))
		},
		controller.NewIndianPokerWebControllerWithProvider)
	games.RegisterKVGame("videopoker", games.CategoryCasino,
		func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(domain.NewDefaultVideoPoker(), new(presenter.VideoPokerWebPresenter))
		},
		func(data []byte) (usecase.VideoPokerInteractorIF, error) {
			return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
		},
		controller.NewVideoPokerWebControllerWithProvider)
	games.RegisterKVGame("deuceswild", games.CategoryCasino,
		func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(domain.NewDeucesWildVideoPoker(), new(presenter.VideoPokerWebPresenter))
		},
		func(data []byte) (usecase.VideoPokerInteractorIF, error) {
			return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
		},
		controller.NewVideoPokerWebControllerWithProvider)
	games.RegisterKVGame("jokerpoker", games.CategoryCasino,
		func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(domain.NewJokerPokerVideoPoker(), new(presenter.VideoPokerWebPresenter))
		},
		func(data []byte) (usecase.VideoPokerInteractorIF, error) {
			return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
		},
		controller.NewVideoPokerWebControllerWithProvider)
	games.RegisterKVGame("threecard", games.CategoryCasino,
		func() usecase.ThreeCardInteractorIF {
			return usecase.NewThreeCardInteractor(domain.NewDefaultThreeCard(), new(presenter.ThreeCardWebPresenter))
		},
		func(data []byte) (usecase.ThreeCardInteractorIF, error) {
			return usecase.RestoreThreeCardInteractor(data, new(presenter.ThreeCardWebPresenter))
		},
		controller.NewThreeCardWebControllerWithProvider)
	games.RegisterKVGame("sevencardstud", games.CategoryCasino,
		func() usecase.SevenCardStudInteractorIF {
			return usecase.NewSevenCardStudInteractor(domain.NewDefaultSevenCardStud(), new(presenter.SevenCardStudWebPresenter))
		},
		func(data []byte) (usecase.SevenCardStudInteractorIF, error) {
			return usecase.RestoreSevenCardStudInteractor(data, new(presenter.SevenCardStudWebPresenter))
		},
		controller.NewSevenCardStudWebControllerWithProvider)
	games.RegisterKVGame("paigow", games.CategoryCasino,
		func() usecase.PaiGowInteractorIF {
			return usecase.NewPaiGowInteractor(domain.NewDefaultPaiGow(), new(presenter.PaiGowWebPresenter))
		},
		func(data []byte) (usecase.PaiGowInteractorIF, error) {
			return usecase.RestorePaiGowInteractor(data, new(presenter.PaiGowWebPresenter))
		},
		controller.NewPaiGowWebControllerWithProvider)
	games.RegisterKVGame("caribbeanstud", games.CategoryCasino,
		func() usecase.CaribbeanStudInteractorIF {
			return usecase.NewCaribbeanStudInteractor(domain.NewDefaultCaribbeanStud(), new(presenter.CaribbeanStudWebPresenter))
		},
		func(data []byte) (usecase.CaribbeanStudInteractorIF, error) {
			return usecase.RestoreCaribbeanStudInteractor(data, new(presenter.CaribbeanStudWebPresenter))
		},
		controller.NewCaribbeanStudWebControllerWithProvider)
	games.RegisterKVGame("texasholdembonus", games.CategoryCasino,
		func() usecase.TexasHoldemBonusInteractorIF {
			return usecase.NewTexasHoldemBonusInteractor(domain.NewDefaultTexasHoldemBonus(), new(presenter.TexasHoldemBonusWebPresenter))
		},
		func(data []byte) (usecase.TexasHoldemBonusInteractorIF, error) {
			return usecase.RestoreTexasHoldemBonusInteractor(data, new(presenter.TexasHoldemBonusWebPresenter))
		},
		controller.NewTexasHoldemBonusWebControllerWithProvider)
	games.RegisterKVGame("letitride", games.CategoryCasino,
		func() usecase.LetItRideInteractorIF {
			return usecase.NewLetItRideInteractor(domain.NewDefaultLetItRide(), new(presenter.LetItRideWebPresenter))
		},
		func(data []byte) (usecase.LetItRideInteractorIF, error) {
			return usecase.RestoreLetItRideInteractor(data, new(presenter.LetItRideWebPresenter))
		},
		controller.NewLetItRideWebControllerWithProvider)
	games.RegisterKVGame("reddog", games.CategoryCasino,
		func() usecase.RedDogInteractorIF {
			return usecase.NewRedDogInteractor(domain.NewDefaultRedDog(), new(presenter.RedDogWebPresenter))
		},
		func(data []byte) (usecase.RedDogInteractorIF, error) {
			return usecase.RestoreRedDogInteractor(data, new(presenter.RedDogWebPresenter))
		},
		controller.NewRedDogWebControllerWithProvider)
	games.RegisterKVGame("razz", games.CategoryCasino,
		func() usecase.SevenCardStudInteractorIF {
			return usecase.NewSevenCardStudInteractor(domain.NewDefaultRazz(), new(presenter.SevenCardStudWebPresenter))
		},
		func(data []byte) (usecase.SevenCardStudInteractorIF, error) {
			return usecase.RestoreSevenCardStudInteractor(data, new(presenter.SevenCardStudWebPresenter))
		},
		controller.NewSevenCardStudWebControllerWithProvider)
	games.RegisterKVGame("badugi", games.CategoryCasino,
		func() usecase.BadugiInteractorIF {
			return usecase.NewBadugiInteractor(domain.NewDefaultBadugi(), new(presenter.BadugiWebPresenter))
		},
		func(data []byte) (usecase.BadugiInteractorIF, error) {
			return usecase.RestoreBadugiInteractor(data, new(presenter.BadugiWebPresenter))
		},
		controller.NewBadugiWebControllerWithProvider)
	games.RegisterKVGame("spanish21", games.CategoryCasino,
		func() usecase.BlackJackInteractorIF {
			return usecase.NewBlackJackInteractor(domain.NewSpanish21BlackJack(), new(presenter.BlackJackWebPresenter))
		},
		func(data []byte) (usecase.BlackJackInteractorIF, error) {
			return usecase.RestoreBlackJackInteractor(data, new(presenter.BlackJackWebPresenter))
		},
		controller.NewBlackJackWebControllerWithProvider)
}
