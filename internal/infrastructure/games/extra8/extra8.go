//go:build js && wasm

// Package extra8 reserves the eleventh Cloudflare Worker size bucket.
package extra8

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func init() {
	games.RegisterKVGame("banluck", games.CategoryExtra8,
		func() usecase.BanLuckInteractorIF {
			return usecase.NewBanLuckInteractor(domain.NewDefaultBanLuck(), new(presenter.BanLuckWebPresenter))
		},
		func(data []byte) (usecase.BanLuckInteractorIF, error) {
			return usecase.RestoreBanLuckInteractor(data, new(presenter.BanLuckWebPresenter))
		},
		controller.NewBanLuckWebControllerWithProvider)
	games.RegisterKVGame("deuceswild", games.CategoryExtra8,
		func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(domain.NewDeucesWildVideoPoker(), new(presenter.VideoPokerWebPresenter))
		},
		func(data []byte) (usecase.VideoPokerInteractorIF, error) {
			return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
		},
		controller.NewVideoPokerWebControllerWithProvider)
	games.RegisterKVGame("doubleattack", games.CategoryExtra8,
		func() usecase.DoubleAttackBlackjackInteractorIF {
			return usecase.NewDoubleAttackBlackjackInteractor(domain.NewDefaultDoubleAttackBlackjack(), new(presenter.DoubleAttackBlackjackWebPresenter))
		},
		func(data []byte) (usecase.DoubleAttackBlackjackInteractorIF, error) {
			return usecase.RestoreDoubleAttackBlackjackInteractor(data, new(presenter.DoubleAttackBlackjackWebPresenter))
		},
		controller.NewDoubleAttackBlackjackWebControllerWithProvider)
	games.RegisterKVGame("fourcardpoker", games.CategoryExtra8,
		func() usecase.FourCardPokerInteractorIF {
			return usecase.NewFourCardPokerInteractor(domain.NewDefaultFourCardPoker(), new(presenter.FourCardPokerWebPresenter))
		},
		func(data []byte) (usecase.FourCardPokerInteractorIF, error) {
			return usecase.RestoreFourCardPokerInteractor(data, new(presenter.FourCardPokerWebPresenter))
		},
		controller.NewFourCardPokerWebControllerWithProvider)
	games.RegisterKVGame("freebet", games.CategoryExtra8,
		func() usecase.FreeBetBlackjackInteractorIF {
			return usecase.NewFreeBetBlackjackInteractor(domain.NewDefaultFreeBetBlackjack(), new(presenter.FreeBetBlackjackWebPresenter))
		},
		func(data []byte) (usecase.FreeBetBlackjackInteractorIF, error) {
			return usecase.RestoreFreeBetBlackjackInteractor(data, new(presenter.FreeBetBlackjackWebPresenter))
		},
		controller.NewFreeBetBlackjackWebControllerWithProvider)
	games.RegisterKVGame("jokerpoker", games.CategoryExtra8,
		func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(domain.NewJokerPokerVideoPoker(), new(presenter.VideoPokerWebPresenter))
		},
		func(data []byte) (usecase.VideoPokerInteractorIF, error) {
			return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
		},
		controller.NewVideoPokerWebControllerWithProvider)
	games.RegisterKVGame("russianpoker", games.CategoryExtra8,
		func() usecase.RussianPokerInteractorIF {
			return usecase.NewRussianPokerInteractor(domain.NewDefaultRussianPoker(), new(presenter.RussianPokerWebPresenter))
		},
		func(data []byte) (usecase.RussianPokerInteractorIF, error) {
			return usecase.RestoreRussianPokerInteractor(data, new(presenter.RussianPokerWebPresenter))
		},
		controller.NewRussianPokerWebControllerWithProvider)
	games.RegisterKVGame("threecardrummy", games.CategoryExtra8,
		func() usecase.ThreeCardRummyInteractorIF {
			return usecase.NewThreeCardRummyInteractor(domain.NewDefaultThreeCardRummy(), new(presenter.ThreeCardRummyWebPresenter))
		},
		func(data []byte) (usecase.ThreeCardRummyInteractorIF, error) {
			return usecase.RestoreThreeCardRummyInteractor(data, new(presenter.ThreeCardRummyWebPresenter))
		},
		controller.NewThreeCardRummyWebControllerWithProvider)
	games.RegisterKVGame("videopoker", games.CategoryExtra8,
		func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(domain.NewDefaultVideoPoker(), new(presenter.VideoPokerWebPresenter))
		},
		func(data []byte) (usecase.VideoPokerInteractorIF, error) {
			return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
		},
		controller.NewVideoPokerWebControllerWithProvider)
	games.RegisterKVGame("blackjack", games.CategoryExtra8,
		func() usecase.BlackJackInteractorIF {
			return usecase.NewBlackJackInteractor(domain.NewDefaultBlackJack(), new(presenter.BlackJackWebPresenter))
		},
		func(data []byte) (usecase.BlackJackInteractorIF, error) {
			return usecase.RestoreBlackJackInteractor(data, new(presenter.BlackJackWebPresenter))
		},
		controller.NewBlackJackWebControllerWithProvider)
	games.RegisterKVGame("doubleexposure", games.CategoryExtra8,
		func() usecase.BlackJackInteractorIF {
			return usecase.NewBlackJackInteractor(domain.NewDoubleExposureBlackJack(), new(presenter.BlackJackWebPresenter))
		},
		func(data []byte) (usecase.BlackJackInteractorIF, error) {
			return usecase.RestoreBlackJackInteractor(data, new(presenter.BlackJackWebPresenter))
		},
		controller.NewBlackJackWebControllerWithProvider)
	games.RegisterKVGame("spanish21", games.CategoryExtra8,
		func() usecase.BlackJackInteractorIF {
			return usecase.NewBlackJackInteractor(domain.NewSpanish21BlackJack(), new(presenter.BlackJackWebPresenter))
		},
		func(data []byte) (usecase.BlackJackInteractorIF, error) {
			return usecase.RestoreBlackJackInteractor(data, new(presenter.BlackJackWebPresenter))
		},
		controller.NewBlackJackWebControllerWithProvider)
	games.RegisterKVGame("blackjackswitch", games.CategoryExtra8,
		func() usecase.BlackJackSwitchInteractorIF {
			return usecase.NewBlackJackSwitchInteractor(domain.NewDefaultBlackJackSwitch(), new(presenter.BlackJackSwitchWebPresenter))
		},
		func(data []byte) (usecase.BlackJackSwitchInteractorIF, error) {
			return usecase.RestoreBlackJackSwitchInteractor(data, new(presenter.BlackJackSwitchWebPresenter))
		},
		controller.NewBlackJackSwitchWebControllerWithProvider)
	games.RegisterKVGame("allfours", games.CategoryExtra8,
		func() usecase.AllFoursInteractorIF {
			return usecase.NewAllFoursInteractor(domain.NewDefaultAllFours(), new(presenter.AllFoursWebPresenter))
		},
		func(data []byte) (usecase.AllFoursInteractorIF, error) {
			return usecase.RestoreAllFoursInteractor(data, new(presenter.AllFoursWebPresenter))
		},
		controller.NewAllFoursWebControllerWithProvider)
	games.RegisterKVGame("botifarra", games.CategoryExtra8,
		func() usecase.BotifarraInteractorIF {
			return usecase.NewBotifarraInteractor(domain.NewDefaultBotifarra(), new(presenter.BotifarraWebPresenter))
		},
		func(data []byte) (usecase.BotifarraInteractorIF, error) {
			return usecase.RestoreBotifarraInteractor(data, new(presenter.BotifarraWebPresenter))
		},
		controller.NewBotifarraWebControllerWithProvider)
	games.RegisterKVGame("cassino", games.CategoryExtra8,
		func() usecase.CassinoInteractorIF {
			return usecase.NewCassinoInteractor(domain.NewDefaultCassino(), new(presenter.CassinoWebPresenter))
		},
		func(data []byte) (usecase.CassinoInteractorIF, error) {
			return usecase.RestoreCassinoInteractor(data, new(presenter.CassinoWebPresenter))
		},
		controller.NewCassinoWebControllerWithProvider)
	games.RegisterKVGame("curdsandwhey", games.CategoryExtra8,
		func() usecase.CurdsAndWheyInteractorIF {
			return usecase.NewCurdsAndWheyInteractor(domain.NewDefaultCurdsAndWhey(), new(presenter.CurdsAndWheyWebPresenter))
		},
		func(data []byte) (usecase.CurdsAndWheyInteractorIF, error) {
			return usecase.RestoreCurdsAndWheyInteractor(data, new(presenter.CurdsAndWheyWebPresenter))
		},
		controller.NewCurdsAndWheyWebControllerWithProvider)
	games.RegisterKVGame("knockoutwhist", games.CategoryExtra8,
		func() usecase.KnockoutWhistInteractorIF {
			return usecase.NewKnockoutWhistInteractor(domain.NewDefaultKnockoutWhist(), new(presenter.KnockoutWhistWebPresenter))
		},
		func(data []byte) (usecase.KnockoutWhistInteractorIF, error) {
			return usecase.RestoreKnockoutWhistInteractor(data, new(presenter.KnockoutWhistWebPresenter))
		},
		controller.NewKnockoutWhistWebControllerWithProvider)
	games.RegisterKVGame("labellelucie", games.CategoryExtra8,
		func() usecase.LaBelleLucieInteractorIF {
			return usecase.NewLaBelleLucieInteractor(domain.NewDefaultLaBelleLucie(), new(presenter.LaBelleLucieWebPresenter))
		},
		func(data []byte) (usecase.LaBelleLucieInteractorIF, error) {
			return usecase.RestoreLaBelleLucieInteractor(data, new(presenter.LaBelleLucieWebPresenter))
		},
		controller.NewLaBelleLucieWebControllerWithProvider)
	games.RegisterKVGame("nap", games.CategoryExtra8,
		func() usecase.NapInteractorIF {
			return usecase.NewNapInteractor(domain.NewDefaultNap(), new(presenter.NapWebPresenter))
		},
		func(data []byte) (usecase.NapInteractorIF, error) {
			return usecase.RestoreNapInteractor(data, new(presenter.NapWebPresenter))
		},
		controller.NewNapWebControllerWithProvider)
	games.RegisterKVGame("reversis", games.CategoryExtra8,
		func() usecase.ReversisInteractorIF {
			return usecase.NewReversisInteractor(domain.NewDefaultReversis(), new(presenter.ReversisWebPresenter))
		},
		func(data []byte) (usecase.ReversisInteractorIF, error) {
			return usecase.RestoreReversisInteractor(data, new(presenter.ReversisWebPresenter))
		},
		controller.NewReversisWebControllerWithProvider)
	games.RegisterKVGame("simplesimon", games.CategoryExtra8,
		func() usecase.SimpleSimonInteractorIF {
			return usecase.NewSimpleSimonInteractor(domain.NewDefaultSimpleSimon(), new(presenter.SimpleSimonWebPresenter))
		},
		func(data []byte) (usecase.SimpleSimonInteractorIF, error) {
			return usecase.RestoreSimpleSimonInteractor(data, new(presenter.SimpleSimonWebPresenter))
		},
		controller.NewSimpleSimonWebControllerWithProvider)
	games.RegisterKVGame("solowhist", games.CategoryExtra8,
		func() usecase.SoloWhistInteractorIF {
			return usecase.NewSoloWhistInteractor(domain.NewDefaultSoloWhist(), new(presenter.SoloWhistWebPresenter))
		},
		func(data []byte) (usecase.SoloWhistInteractorIF, error) {
			return usecase.RestoreSoloWhistInteractor(data, new(presenter.SoloWhistWebPresenter))
		},
		controller.NewSoloWhistWebControllerWithProvider)
	games.RegisterKVGame("spoilfive", games.CategoryExtra8,
		func() usecase.SpoilFiveInteractorIF {
			return usecase.NewSpoilFiveInteractor(domain.NewDefaultSpoilFive(), new(presenter.SpoilFiveWebPresenter))
		},
		func(data []byte) (usecase.SpoilFiveInteractorIF, error) {
			return usecase.RestoreSpoilFiveInteractor(data, new(presenter.SpoilFiveWebPresenter))
		},
		controller.NewSpoilFiveWebControllerWithProvider)
	games.RegisterKVGame("anaconda", games.CategoryExtra8,
		func() usecase.AnacondaInteractorIF {
			return usecase.NewAnacondaInteractor(domain.NewDefaultAnaconda(), new(presenter.AnacondaWebPresenter))
		},
		func(data []byte) (usecase.AnacondaInteractorIF, error) {
			return usecase.RestoreAnacondaInteractor(data, new(presenter.AnacondaWebPresenter))
		},
		controller.NewAnacondaWebControllerWithProvider)
	games.RegisterKVGame("barbu", games.CategoryExtra8,
		func() usecase.BarbuInteractorIF {
			return usecase.NewBarbuInteractor(domain.NewDefaultBarbu(), new(presenter.BarbuWebPresenter))
		},
		func(data []byte) (usecase.BarbuInteractorIF, error) {
			return usecase.RestoreBarbuInteractor(data, new(presenter.BarbuWebPresenter))
		},
		controller.NewBarbuWebControllerWithProvider)
	games.RegisterKVGame("chemindefer", games.CategoryExtra8,
		func() usecase.ChemindeFerInteractorIF {
			return usecase.NewChemindeFerInteractor(domain.NewDefaultChemindeFer(), new(presenter.ChemindeFerWebPresenter))
		},
		func(data []byte) (usecase.ChemindeFerInteractorIF, error) {
			return usecase.RestoreChemindeFerInteractor(data, new(presenter.ChemindeFerWebPresenter))
		},
		controller.NewChemindeFerWebControllerWithProvider)
	games.RegisterKVGame("ombre", games.CategoryExtra8,
		func() usecase.OmbreInteractorIF {
			return usecase.NewOmbreInteractor(domain.NewDefaultOmbre(), new(presenter.OmbreWebPresenter))
		},
		func(data []byte) (usecase.OmbreInteractorIF, error) {
			return usecase.RestoreOmbreInteractor(data, new(presenter.OmbreWebPresenter))
		},
		controller.NewOmbreWebControllerWithProvider)
	games.RegisterKVGame("piquet", games.CategoryExtra8,
		func() usecase.PiquetInteractorIF {
			return usecase.NewPiquetInteractor(domain.NewDefaultPiquet(), new(presenter.PiquetWebPresenter))
		},
		func(data []byte) (usecase.PiquetInteractorIF, error) {
			return usecase.RestorePiquetInteractor(data, new(presenter.PiquetWebPresenter))
		},
		controller.NewPiquetWebControllerWithProvider)
	games.RegisterKVGame("shengji", games.CategoryExtra8,
		func() usecase.ShengJiInteractorIF {
			return usecase.NewShengJiInteractor(domain.NewDefaultShengJi(), new(presenter.ShengJiWebPresenter))
		},
		func(data []byte) (usecase.ShengJiInteractorIF, error) {
			return usecase.RestoreShengJiInteractor(data, new(presenter.ShengJiWebPresenter))
		},
		controller.NewShengJiWebControllerWithProvider)
	games.RegisterKVGame("sixbidsolo", games.CategoryExtra8,
		func() usecase.SixBidSoloInteractorIF {
			return usecase.NewSixBidSoloInteractor(domain.NewDefaultSixBidSolo(), new(presenter.SixBidSoloWebPresenter))
		},
		func(data []byte) (usecase.SixBidSoloInteractorIF, error) {
			return usecase.RestoreSixBidSoloInteractor(data, new(presenter.SixBidSoloWebPresenter))
		},
		controller.NewSixBidSoloWebControllerWithProvider)
}
