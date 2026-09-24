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
	games.RegisterKVGame("courchevel", games.CategoryCasino,
		func() usecase.OmahaInteractorIF {
			return usecase.NewOmahaInteractor(domain.NewDefaultCourchevel(), new(presenter.OmahaWebPresenter))
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
	games.RegisterKVGame("courchevelhilo", games.CategoryCasino,
		func() usecase.OmahaInteractorIF {
			return usecase.NewOmahaInteractor(domain.NewDefaultCourchevelHiLo(), new(presenter.OmahaWebPresenter))
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
	games.RegisterKVGame("caribbeanstud", games.CategoryCasino,
		func() usecase.CaribbeanStudInteractorIF {
			return usecase.NewCaribbeanStudInteractor(domain.NewDefaultCaribbeanStud(), new(presenter.CaribbeanStudWebPresenter))
		},
		func(data []byte) (usecase.CaribbeanStudInteractorIF, error) {
			return usecase.RestoreCaribbeanStudInteractor(data, new(presenter.CaribbeanStudWebPresenter))
		},
		controller.NewCaribbeanStudWebControllerWithProvider)
	games.RegisterKVGame("razz", games.CategoryCasino,
		func() usecase.SevenCardStudInteractorIF {
			return usecase.NewSevenCardStudInteractor(domain.NewDefaultRazz(), new(presenter.SevenCardStudWebPresenter))
		},
		func(data []byte) (usecase.SevenCardStudInteractorIF, error) {
			return usecase.RestoreSevenCardStudInteractor(data, new(presenter.SevenCardStudWebPresenter))
		},
		controller.NewSevenCardStudWebControllerWithProvider)
	games.RegisterKVGame("chicago", games.CategoryCasino,
		func() usecase.SevenCardStudInteractorIF {
			return usecase.NewSevenCardStudInteractor(domain.NewDefaultSevenCardStudChicago(), new(presenter.SevenCardStudWebPresenter))
		},
		func(data []byte) (usecase.SevenCardStudInteractorIF, error) {
			return usecase.RestoreSevenCardStudInteractor(data, new(presenter.SevenCardStudWebPresenter))
		},
		controller.NewSevenCardStudWebControllerWithProvider)
	games.RegisterKVGame("sevencardstudhilo", games.CategoryCasino,
		func() usecase.SevenCardStudInteractorIF {
			return usecase.NewSevenCardStudInteractor(domain.NewDefaultSevenCardStudHiLo(), new(presenter.SevenCardStudWebPresenter))
		},
		func(data []byte) (usecase.SevenCardStudInteractorIF, error) {
			return usecase.RestoreSevenCardStudInteractor(data, new(presenter.SevenCardStudWebPresenter))
		},
		controller.NewSevenCardStudWebControllerWithProvider)
	games.RegisterKVGame("soko", games.CategoryCasino,
		func() usecase.FiveCardStudInteractorIF {
			return usecase.NewFiveCardStudInteractor(domain.NewDefaultSoko(), new(presenter.FiveCardStudWebPresenter))
		},
		func(data []byte) (usecase.FiveCardStudInteractorIF, error) {
			return usecase.RestoreFiveCardStudInteractor(data, new(presenter.FiveCardStudWebPresenter))
		},
		controller.NewFiveCardStudWebControllerWithProvider)
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

	games.RegisterKVGame("cincinnati", games.CategoryCasino,
		func() usecase.CincinnatiInteractorIF {
			return usecase.NewCincinnatiInteractor(domain.NewDefaultCincinnati(), new(presenter.CincinnatiWebPresenter))
		},
		func(data []byte) (usecase.CincinnatiInteractorIF, error) {
			return usecase.RestoreCincinnatiInteractor(data, new(presenter.CincinnatiWebPresenter))
		},
		controller.NewCincinnatiWebControllerWithProvider)
	games.RegisterKVGame("ironcross", games.CategoryCasino,
		func() usecase.IronCrossInteractorIF {
			return usecase.NewIronCrossInteractor(domain.NewDefaultIronCross(), new(presenter.IronCrossWebPresenter))
		},
		func(data []byte) (usecase.IronCrossInteractorIF, error) {
			return usecase.RestoreIronCrossInteractor(data, new(presenter.IronCrossWebPresenter))
		},
		controller.NewIronCrossWebControllerWithProvider)
	games.RegisterKVGame("baseballpoker", games.CategoryCasino,
		func() usecase.BaseballPokerInteractorIF {
			return usecase.NewBaseballPokerInteractor(domain.NewDefaultBaseballPoker(), new(presenter.BaseballPokerWebPresenter))
		},
		func(data []byte) (usecase.BaseballPokerInteractorIF, error) {
			return usecase.RestoreBaseballPokerInteractor(data, new(presenter.BaseballPokerWebPresenter))
		},
		controller.NewBaseballPokerWebControllerWithProvider)
	games.RegisterKVGame("openfacechinese", games.CategoryCasino,
		func() usecase.OpenFaceChineseInteractorIF {
			return usecase.NewOpenFaceChineseInteractor(domain.NewDefaultOpenFaceChinese(), new(presenter.OpenFaceChineseWebPresenter))
		},
		func(data []byte) (usecase.OpenFaceChineseInteractorIF, error) {
			return usecase.RestoreOpenFaceChineseInteractor(data, new(presenter.OpenFaceChineseWebPresenter))
		},
		controller.NewOpenFaceChineseWebControllerWithProvider)
	games.RegisterKVGame("eightgame", games.CategoryCasino,
		func() usecase.HorseInteractorIF {
			return usecase.NewHorseInteractor(domain.NewDefaultEightGame(), new(presenter.HorseWebPresenter))
		},
		func(data []byte) (usecase.HorseInteractorIF, error) {
			return usecase.RestoreHorseInteractor(data, new(presenter.HorseWebPresenter))
		},
		controller.NewHorseWebControllerWithProvider)

	games.RegisterKVGame("horse", games.CategoryCasino,
		func() usecase.HorseInteractorIF {
			return usecase.NewHorseInteractor(domain.NewDefaultHorse(), new(presenter.HorseWebPresenter))
		},
		func(data []byte) (usecase.HorseInteractorIF, error) {
			return usecase.RestoreHorseInteractor(data, new(presenter.HorseWebPresenter))
		},
		controller.NewHorseWebControllerWithProvider)

	games.RegisterKVGame("followthequeen", games.CategoryCasino,
		func() usecase.FollowTheQueenInteractorIF {
			return usecase.NewFollowTheQueenInteractor(domain.NewDefaultFollowTheQueen(), new(presenter.FollowTheQueenWebPresenter))
		},
		func(data []byte) (usecase.FollowTheQueenInteractorIF, error) {
			return usecase.RestoreFollowTheQueenInteractor(data, new(presenter.FollowTheQueenWebPresenter))
		},
		controller.NewFollowTheQueenWebControllerWithProvider)
	games.RegisterKVGame("dramaha", games.CategoryCasino,
		func() usecase.DramahaInteractorIF {
			return usecase.NewDramahaInteractor(domain.NewDefaultDramaha(), new(presenter.DramahaWebPresenter))
		},
		func(data []byte) (usecase.DramahaInteractorIF, error) {
			return usecase.RestoreDramahaInteractor(data, new(presenter.DramahaWebPresenter))
		},
		controller.NewDramahaWebControllerWithProvider)
	games.RegisterKVGame("casinoholdem", games.CategoryCasino,
		func() usecase.CasinoHoldemInteractorIF {
			return usecase.NewCasinoHoldemInteractor(domain.NewDefaultCasinoHoldem(), new(presenter.CasinoHoldemWebPresenter))
		},
		func(data []byte) (usecase.CasinoHoldemInteractorIF, error) {
			return usecase.RestoreCasinoHoldemInteractor(data, new(presenter.CasinoHoldemWebPresenter))
		},
		controller.NewCasinoHoldemWebControllerWithProvider)
	games.RegisterKVGame("ultimatetexasholdem", games.CategoryCasino,
		func() usecase.UltimateTexasHoldemInteractorIF {
			return usecase.NewUltimateTexasHoldemInteractor(domain.NewDefaultUltimateTexasHoldem(), new(presenter.UltimateTexasHoldemWebPresenter))
		},
		func(data []byte) (usecase.UltimateTexasHoldemInteractorIF, error) {
			return usecase.RestoreUltimateTexasHoldemInteractor(data, new(presenter.UltimateTexasHoldemWebPresenter))
		},
		controller.NewUltimateTexasHoldemWebControllerWithProvider)
	games.RegisterKVGame("texasholdembonus", games.CategoryCasino,
		func() usecase.TexasHoldemBonusInteractorIF {
			return usecase.NewTexasHoldemBonusInteractor(domain.NewDefaultTexasHoldemBonus(), new(presenter.TexasHoldemBonusWebPresenter))
		},
		func(data []byte) (usecase.TexasHoldemBonusInteractorIF, error) {
			return usecase.RestoreTexasHoldemBonusInteractor(data, new(presenter.TexasHoldemBonusWebPresenter))
		},
		controller.NewTexasHoldemBonusWebControllerWithProvider)
}
