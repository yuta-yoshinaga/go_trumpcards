//go:build js && wasm

// Package casino binds the Cloudflare Worker KV-backed handlers for the
// table and poker games. A worker main must blank-import this package
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
	games.RegisterKVGame("omahahilo", games.CategoryCasino,
		func() usecase.OmahaInteractorIF {
			return usecase.NewOmahaInteractor(domain.NewDefaultOmahaHiLo(), new(presenter.OmahaWebPresenter))
		},
		func(data []byte) (usecase.OmahaInteractorIF, error) {
			return usecase.RestoreOmahaInteractor(data, new(presenter.OmahaWebPresenter))
		},
		controller.NewOmahaWebControllerWithProvider)
	games.RegisterKVGame("bigo", games.CategoryCasino,
		func() usecase.OmahaInteractorIF {
			return usecase.NewOmahaInteractor(domain.NewDefaultBigO(), new(presenter.OmahaWebPresenter))
		},
		func(data []byte) (usecase.OmahaInteractorIF, error) {
			return usecase.RestoreOmahaInteractor(data, new(presenter.OmahaWebPresenter))
		},
		controller.NewOmahaWebControllerWithProvider)
	games.RegisterKVGame("bigohilo", games.CategoryCasino,
		func() usecase.OmahaInteractorIF {
			return usecase.NewOmahaInteractor(domain.NewDefaultBigOHiLo(), new(presenter.OmahaWebPresenter))
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
	games.RegisterKVGame("irishpoker", games.CategoryCasino,
		func() usecase.PineappleInteractorIF {
			return usecase.NewPineappleInteractor(domain.NewDefaultIrishPoker(), new(presenter.PineappleWebPresenter))
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
	games.RegisterKVGame("deucetoseven", games.CategoryCasino,
		func() usecase.DeuceToSevenInteractorIF {
			return usecase.NewDeuceToSevenInteractor(domain.NewDefaultDeuceToSeven(), new(presenter.DeuceToSevenWebPresenter))
		},
		func(data []byte) (usecase.DeuceToSevenInteractorIF, error) {
			return usecase.RestoreDeuceToSevenInteractor(data, new(presenter.DeuceToSevenWebPresenter))
		},
		controller.NewDeuceToSevenWebControllerWithProvider)
	games.RegisterKVGame("spanish21", games.CategoryCasino,
		func() usecase.BlackJackInteractorIF {
			return usecase.NewBlackJackInteractor(domain.NewSpanish21BlackJack(), new(presenter.BlackJackWebPresenter))
		},
		func(data []byte) (usecase.BlackJackInteractorIF, error) {
			return usecase.RestoreBlackJackInteractor(data, new(presenter.BlackJackWebPresenter))
		},
		controller.NewBlackJackWebControllerWithProvider)
	games.RegisterKVGame("casinowar", games.CategoryCasino,
		func() usecase.CasinoWarInteractorIF {
			return usecase.NewCasinoWarInteractor(domain.NewDefaultCasinoWar(), new(presenter.CasinoWarWebPresenter))
		},
		func(data []byte) (usecase.CasinoWarInteractorIF, error) {
			return usecase.RestoreCasinoWarInteractor(data, new(presenter.CasinoWarWebPresenter))
		},
		controller.NewCasinoWarWebControllerWithProvider)
	games.RegisterKVGame("dragontiger", games.CategoryCasino,
		func() usecase.DragonTigerInteractorIF {
			return usecase.NewDragonTigerInteractor(domain.NewDefaultDragonTiger(), new(presenter.DragonTigerWebPresenter))
		},
		func(data []byte) (usecase.DragonTigerInteractorIF, error) {
			return usecase.RestoreDragonTigerInteractor(data, new(presenter.DragonTigerWebPresenter))
		},
		controller.NewDragonTigerWebControllerWithProvider)
	games.RegisterKVGame("blackjackswitch", games.CategoryCasino,
		func() usecase.BlackJackSwitchInteractorIF {
			return usecase.NewBlackJackSwitchInteractor(domain.NewDefaultBlackJackSwitch(), new(presenter.BlackJackSwitchWebPresenter))
		},
		func(data []byte) (usecase.BlackJackSwitchInteractorIF, error) {
			return usecase.RestoreBlackJackSwitchInteractor(data, new(presenter.BlackJackSwitchWebPresenter))
		},
		controller.NewBlackJackSwitchWebControllerWithProvider)
	games.RegisterKVGame("ultimatetexasholdem", games.CategoryCasino,
		func() usecase.UltimateTexasHoldemInteractorIF {
			return usecase.NewUltimateTexasHoldemInteractor(domain.NewDefaultUltimateTexasHoldem(), new(presenter.UltimateTexasHoldemWebPresenter))
		},
		func(data []byte) (usecase.UltimateTexasHoldemInteractorIF, error) {
			return usecase.RestoreUltimateTexasHoldemInteractor(data, new(presenter.UltimateTexasHoldemWebPresenter))
		},
		controller.NewUltimateTexasHoldemWebControllerWithProvider)
	games.RegisterKVGame("mississippistud", games.CategoryCasino,
		func() usecase.MississippiStudInteractorIF {
			return usecase.NewMississippiStudInteractor(domain.NewDefaultMississippiStud(), new(presenter.MississippiStudWebPresenter))
		},
		func(data []byte) (usecase.MississippiStudInteractorIF, error) {
			return usecase.RestoreMississippiStudInteractor(data, new(presenter.MississippiStudWebPresenter))
		},
		controller.NewMississippiStudWebControllerWithProvider)
	games.RegisterKVGame("oasispoker", games.CategoryCasino,
		func() usecase.OasisPokerInteractorIF {
			return usecase.NewOasisPokerInteractor(domain.NewDefaultOasisPoker(), new(presenter.OasisPokerWebPresenter))
		},
		func(data []byte) (usecase.OasisPokerInteractorIF, error) {
			return usecase.RestoreOasisPokerInteractor(data, new(presenter.OasisPokerWebPresenter))
		},
		controller.NewOasisPokerWebControllerWithProvider)
	games.RegisterKVGame("casinoholdem", games.CategoryCasino,
		func() usecase.CasinoHoldemInteractorIF {
			return usecase.NewCasinoHoldemInteractor(domain.NewDefaultCasinoHoldem(), new(presenter.CasinoHoldemWebPresenter))
		},
		func(data []byte) (usecase.CasinoHoldemInteractorIF, error) {
			return usecase.RestoreCasinoHoldemInteractor(data, new(presenter.CasinoHoldemWebPresenter))
		},
		controller.NewCasinoHoldemWebControllerWithProvider)
	games.RegisterKVGame("highcardflush", games.CategoryCasino,
		func() usecase.HighCardFlushInteractorIF {
			return usecase.NewHighCardFlushInteractor(domain.NewDefaultHighCardFlush(), new(presenter.HighCardFlushWebPresenter))
		},
		func(data []byte) (usecase.HighCardFlushInteractorIF, error) {
			return usecase.RestoreHighCardFlushInteractor(data, new(presenter.HighCardFlushWebPresenter))
		},
		controller.NewHighCardFlushWebControllerWithProvider)
	games.RegisterKVGame("fourcardpoker", games.CategoryCasino,
		func() usecase.FourCardPokerInteractorIF {
			return usecase.NewFourCardPokerInteractor(domain.NewDefaultFourCardPoker(), new(presenter.FourCardPokerWebPresenter))
		},
		func(data []byte) (usecase.FourCardPokerInteractorIF, error) {
			return usecase.RestoreFourCardPokerInteractor(data, new(presenter.FourCardPokerWebPresenter))
		},
		controller.NewFourCardPokerWebControllerWithProvider)
	games.RegisterKVGame("russianpoker", games.CategoryCasino,
		func() usecase.RussianPokerInteractorIF {
			return usecase.NewRussianPokerInteractor(domain.NewDefaultRussianPoker(), new(presenter.RussianPokerWebPresenter))
		},
		func(data []byte) (usecase.RussianPokerInteractorIF, error) {
			return usecase.RestoreRussianPokerInteractor(data, new(presenter.RussianPokerWebPresenter))
		},
		controller.NewRussianPokerWebControllerWithProvider)
	games.RegisterKVGame("chinesepoker", games.CategoryCasino,
		func() usecase.ChinesePokerInteractorIF {
			return usecase.NewChinesePokerInteractor(domain.NewDefaultChinesePoker(), new(presenter.ChinesePokerWebPresenter))
		},
		func(data []byte) (usecase.ChinesePokerInteractorIF, error) {
			return usecase.RestoreChinesePokerInteractor(data, new(presenter.ChinesePokerWebPresenter))
		},
		controller.NewChinesePokerWebControllerWithProvider)
	// Scopa is a classic fishing game, but it is bucketed into the casino
	// worker because the classic worker is at the 1 MB gzip free-tier limit.
	// Workers are pure binary-size partitions with no user-facing meaning.

	games.RegisterKVGame("bridge", games.CategoryCasino,
		func() usecase.BridgeInteractorIF {
			return usecase.NewBridgeInteractor(domain.NewDefaultBridge(), new(presenter.BridgeWebPresenter))
		},
		func(data []byte) (usecase.BridgeInteractorIF, error) {
			return usecase.RestoreBridgeInteractor(data, new(presenter.BridgeWebPresenter))
		},
		controller.NewBridgeWebControllerWithProvider)

	games.RegisterKVGame("napoleon", games.CategoryCasino,
		func() usecase.NapoleonInteractorIF {
			return usecase.NewNapoleonInteractor(domain.NewDefaultNapoleon(), new(presenter.NapoleonWebPresenter))
		},
		func(data []byte) (usecase.NapoleonInteractorIF, error) {
			return usecase.RestoreNapoleonInteractor(data, new(presenter.NapoleonWebPresenter))
		},
		controller.NewNapoleonWebControllerWithProvider)

	games.RegisterKVGame("skat", games.CategoryCasino,
		func() usecase.SkatInteractorIF {
			return usecase.NewSkatInteractor(domain.NewDefaultSkat(), new(presenter.SkatWebPresenter))
		},
		func(data []byte) (usecase.SkatInteractorIF, error) {
			return usecase.RestoreSkatInteractor(data, new(presenter.SkatWebPresenter))
		},
		controller.NewSkatWebControllerWithProvider)

	games.RegisterKVGame("tarneeb", games.CategoryCasino,
		func() usecase.TarneebInteractorIF {
			return usecase.NewTarneebInteractor(domain.NewDefaultTarneeb(), new(presenter.TarneebWebPresenter))
		},
		func(data []byte) (usecase.TarneebInteractorIF, error) {
			return usecase.RestoreTarneebInteractor(data, new(presenter.TarneebWebPresenter))
		},
		controller.NewTarneebWebControllerWithProvider)

	games.RegisterKVGame("belote", games.CategoryCasino,
		func() usecase.BeloteInteractorIF {
			return usecase.NewBeloteInteractor(domain.NewDefaultBelote(), new(presenter.BeloteWebPresenter))
		},
		func(data []byte) (usecase.BeloteInteractorIF, error) {
			return usecase.RestoreBeloteInteractor(data, new(presenter.BeloteWebPresenter))
		},
		controller.NewBeloteWebControllerWithProvider)
	games.RegisterKVGame("tressette", games.CategoryCasino,
		func() usecase.TressetteInteractorIF {
			return usecase.NewTressetteInteractor(domain.NewDefaultTressette(), new(presenter.TressetteWebPresenter))
		},
		func(data []byte) (usecase.TressetteInteractorIF, error) {
			return usecase.RestoreTressetteInteractor(data, new(presenter.TressetteWebPresenter))
		},
		controller.NewTressetteWebControllerWithProvider)
	games.RegisterKVGame("sheepshead", games.CategoryCasino,
		func() usecase.SheepsheadInteractorIF {
			return usecase.NewSheepsheadInteractor(domain.NewDefaultSheepshead(), new(presenter.SheepsheadWebPresenter))
		},
		func(data []byte) (usecase.SheepsheadInteractorIF, error) {
			return usecase.RestoreSheepsheadInteractor(data, new(presenter.SheepsheadWebPresenter))
		},
		controller.NewSheepsheadWebControllerWithProvider)
	games.RegisterKVGame("doppelkopf", games.CategoryCasino,
		func() usecase.DoppelkopfInteractorIF {
			return usecase.NewDoppelkopfInteractor(domain.NewDefaultDoppelkopf(), new(presenter.DoppelkopfWebPresenter))
		},
		func(data []byte) (usecase.DoppelkopfInteractorIF, error) {
			return usecase.RestoreDoppelkopfInteractor(data, new(presenter.DoppelkopfWebPresenter))
		},
		controller.NewDoppelkopfWebControllerWithProvider)
	games.RegisterKVGame("mus", games.CategoryCasino,
		func() usecase.MusInteractorIF {
			return usecase.NewMusInteractor(domain.NewDefaultMus(), new(presenter.MusWebPresenter))
		},
		func(data []byte) (usecase.MusInteractorIF, error) {
			return usecase.RestoreMusInteractor(data, new(presenter.MusWebPresenter))
		},
		controller.NewMusWebControllerWithProvider)
	games.RegisterKVGame("tute", games.CategoryCasino,
		func() usecase.TuteInteractorIF {
			return usecase.NewTuteInteractor(domain.NewDefaultTute(), new(presenter.TuteWebPresenter))
		},
		func(data []byte) (usecase.TuteInteractorIF, error) {
			return usecase.RestoreTuteInteractor(data, new(presenter.TuteWebPresenter))
		},
		controller.NewTuteWebControllerWithProvider)
	games.RegisterKVGame("sueca", games.CategoryCasino,
		func() usecase.SuecaInteractorIF {
			return usecase.NewSuecaInteractor(domain.NewDefaultSueca(), new(presenter.SuecaWebPresenter))
		},
		func(data []byte) (usecase.SuecaInteractorIF, error) {
			return usecase.RestoreSuecaInteractor(data, new(presenter.SuecaWebPresenter))
		},
		controller.NewSuecaWebControllerWithProvider)
	games.RegisterKVGame("fortyfives", games.CategoryCasino,
		func() usecase.FortyFivesInteractorIF {
			return usecase.NewFortyFivesInteractor(domain.NewDefaultFortyFives(), new(presenter.FortyFivesWebPresenter))
		},
		func(data []byte) (usecase.FortyFivesInteractorIF, error) {
			return usecase.RestoreFortyFivesInteractor(data, new(presenter.FortyFivesWebPresenter))
		},
		controller.NewFortyFivesWebControllerWithProvider)
	games.RegisterKVGame("twentynine", games.CategoryCasino,
		func() usecase.TwentyNineInteractorIF {
			return usecase.NewTwentyNineInteractor(domain.NewDefaultTwentyNine(), new(presenter.TwentyNineWebPresenter))
		},
		func(data []byte) (usecase.TwentyNineInteractorIF, error) {
			return usecase.RestoreTwentyNineInteractor(data, new(presenter.TwentyNineWebPresenter))
		},
		controller.NewTwentyNineWebControllerWithProvider)
	games.RegisterKVGame("bourre", games.CategoryCasino,
		func() usecase.BourreInteractorIF {
			return usecase.NewBourreInteractor(domain.NewDefaultBourre(), new(presenter.BourreWebPresenter))
		},
		func(data []byte) (usecase.BourreInteractorIF, error) {
			return usecase.RestoreBourreInteractor(data, new(presenter.BourreWebPresenter))
		},
		controller.NewBourreWebControllerWithProvider)
	games.RegisterKVGame("courtpiece", games.CategoryCasino,
		func() usecase.CourtPieceInteractorIF {
			return usecase.NewCourtPieceInteractor(domain.NewDefaultCourtPiece(), new(presenter.CourtPieceWebPresenter))
		},
		func(data []byte) (usecase.CourtPieceInteractorIF, error) {
			return usecase.RestoreCourtPieceInteractor(data, new(presenter.CourtPieceWebPresenter))
		},
		controller.NewCourtPieceWebControllerWithProvider)
	games.RegisterKVGame("ecarte", games.CategoryCasino,
		func() usecase.EcarteInteractorIF {
			return usecase.NewEcarteInteractor(domain.NewDefaultEcarte(), new(presenter.EcarteWebPresenter))
		},
		func(data []byte) (usecase.EcarteInteractorIF, error) {
			return usecase.RestoreEcarteInteractor(data, new(presenter.EcarteWebPresenter))
		},
		controller.NewEcarteWebControllerWithProvider)
	games.RegisterKVGame("threecardbrag", games.CategoryCasino,
		func() usecase.ThreeCardBragInteractorIF {
			return usecase.NewThreeCardBragInteractor(domain.NewDefaultThreeCardBrag(), new(presenter.ThreeCardBragWebPresenter))
		},
		func(data []byte) (usecase.ThreeCardBragInteractorIF, error) {
			return usecase.RestoreThreeCardBragInteractor(data, new(presenter.ThreeCardBragWebPresenter))
		},
		controller.NewThreeCardBragWebControllerWithProvider)
	games.RegisterKVGame("teenpatti", games.CategoryCasino,
		func() usecase.TeenPattiInteractorIF {
			return usecase.NewTeenPattiInteractor(domain.NewDefaultTeenPatti(), new(presenter.TeenPattiWebPresenter))
		},
		func(data []byte) (usecase.TeenPattiInteractorIF, error) {
			return usecase.RestoreTeenPattiInteractor(data, new(presenter.TeenPattiWebPresenter))
		},
		controller.NewTeenPattiWebControllerWithProvider)

	games.RegisterKVGame("fivecardstud", games.CategoryCasino,
		func() usecase.FiveCardStudInteractorIF {
			return usecase.NewFiveCardStudInteractor(domain.NewDefaultFiveCardStud(), new(presenter.FiveCardStudWebPresenter))
		},
		func(data []byte) (usecase.FiveCardStudInteractorIF, error) {
			return usecase.RestoreFiveCardStudInteractor(data, new(presenter.FiveCardStudWebPresenter))
		},
		controller.NewFiveCardStudWebControllerWithProvider)

	games.RegisterKVGame("openfacechinese", games.CategoryCasino,
		func() usecase.OpenFaceChineseInteractorIF {
			return usecase.NewOpenFaceChineseInteractor(domain.NewDefaultOpenFaceChinese(), new(presenter.OpenFaceChineseWebPresenter))
		},
		func(data []byte) (usecase.OpenFaceChineseInteractorIF, error) {
			return usecase.RestoreOpenFaceChineseInteractor(data, new(presenter.OpenFaceChineseWebPresenter))
		},
		controller.NewOpenFaceChineseWebControllerWithProvider)

}
