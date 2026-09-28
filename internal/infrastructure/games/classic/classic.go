//go:build js && wasm

// Package classic binds the Cloudflare Worker KV-backed handlers for the
// 21 trick-taking, matching, and family card games. A worker main must
// blank-import this package so the init below runs before
// games.RegisterCategory is called.
package classic

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func init() {
	games.RegisterKVGame("oldmaid", games.CategoryClassic,
		func() usecase.OldMaidInteractorIF {
			return usecase.NewOldMaidInteractor(domain.NewDefaultOldMaid(), new(presenter.OldMaidWebPresenter))
		},
		func(data []byte) (usecase.OldMaidInteractorIF, error) {
			return usecase.RestoreOldMaidInteractor(data, new(presenter.OldMaidWebPresenter))
		},
		controller.NewOldMaidWebControllerWithProvider)
	games.RegisterKVGame("sevens", games.CategoryClassic,
		func() usecase.SevensInteractorIF {
			return usecase.NewSevensInteractor(domain.NewDefaultSevens(), new(presenter.SevensWebPresenter))
		},
		func(data []byte) (usecase.SevensInteractorIF, error) {
			return usecase.RestoreSevensInteractor(data, new(presenter.SevensWebPresenter))
		},
		controller.NewSevensWebControllerWithProvider)
	games.RegisterKVGame("spades", games.CategoryClassic,
		func() usecase.SpadesInteractorIF {
			return usecase.NewSpadesInteractor(domain.NewDefaultSpades(), new(presenter.SpadesWebPresenter))
		},
		func(data []byte) (usecase.SpadesInteractorIF, error) {
			return usecase.RestoreSpadesInteractor(data, new(presenter.SpadesWebPresenter))
		},
		controller.NewSpadesWebControllerWithProvider)
	games.RegisterKVGame("crazyeights", games.CategoryClassic,
		func() usecase.CrazyEightsInteractorIF {
			return usecase.NewCrazyEightsInteractor(domain.NewDefaultCrazyEights(), new(presenter.CrazyEightsWebPresenter))
		},
		func(data []byte) (usecase.CrazyEightsInteractorIF, error) {
			return usecase.RestoreCrazyEightsInteractor(data, new(presenter.CrazyEightsWebPresenter))
		},
		controller.NewCrazyEightsWebControllerWithProvider)
	games.RegisterKVGame("ohhell", games.CategoryClassic,
		func() usecase.OhHellInteractorIF {
			return usecase.NewOhHellInteractor(domain.NewDefaultOhHell(), new(presenter.OhHellWebPresenter))
		},
		func(data []byte) (usecase.OhHellInteractorIF, error) {
			return usecase.RestoreOhHellInteractor(data, new(presenter.OhHellWebPresenter))
		},
		controller.NewOhHellWebControllerWithProvider)
	games.RegisterKVGame("ninetynine", games.CategoryClassic,
		func() usecase.NinetyNineInteractorIF {
			return usecase.NewNinetyNineInteractor(domain.NewDefaultNinetyNine(), new(presenter.NinetyNineWebPresenter))
		},
		func(data []byte) (usecase.NinetyNineInteractorIF, error) {
			return usecase.RestoreNinetyNineInteractor(data, new(presenter.NinetyNineWebPresenter))
		},
		controller.NewNinetyNineWebControllerWithProvider)
	games.RegisterKVGame("durak", games.CategoryClassic,
		func() usecase.DurakInteractorIF {
			return usecase.NewDurakInteractor(domain.NewDefaultDurak(), new(presenter.DurakWebPresenter))
		},
		func(data []byte) (usecase.DurakInteractorIF, error) {
			return usecase.RestoreDurakInteractor(data, new(presenter.DurakWebPresenter))
		},
		controller.NewDurakWebControllerWithProvider)
	games.RegisterKVGame("twotenjack", games.CategoryClassic,
		func() usecase.TwoTenJackInteractorIF {
			return usecase.NewTwoTenJackInteractor(domain.NewDefaultTwoTenJack(), new(presenter.TwoTenJackWebPresenter))
		},
		func(data []byte) (usecase.TwoTenJackInteractorIF, error) {
			return usecase.RestoreTwoTenJackInteractor(data, new(presenter.TwoTenJackWebPresenter))
		},
		controller.NewTwoTenJackWebControllerWithProvider)
	games.RegisterKVGame("whist", games.CategoryClassic,
		func() usecase.WhistInteractorIF {
			return usecase.NewWhistInteractor(domain.NewDefaultWhist(), new(presenter.WhistWebPresenter))
		},
		func(data []byte) (usecase.WhistInteractorIF, error) {
			return usecase.RestoreWhistInteractor(data, new(presenter.WhistWebPresenter))
		},
		controller.NewWhistWebControllerWithProvider)
	games.RegisterKVGame("catchten", games.CategoryClassic,
		func() usecase.CatchTenInteractorIF {
			return usecase.NewCatchTenInteractor(domain.NewDefaultCatchTen(), new(presenter.CatchTenWebPresenter))
		},
		func(data []byte) (usecase.CatchTenInteractorIF, error) {
			return usecase.RestoreCatchTenInteractor(data, new(presenter.CatchTenWebPresenter))
		},
		controller.NewCatchTenWebControllerWithProvider)
	games.RegisterKVGame("pageone", games.CategoryClassic,
		func() usecase.PageOneInteractorIF {
			return usecase.NewPageOneInteractor(domain.NewDefaultPageOne(), new(presenter.PageOneWebPresenter))
		},
		func(data []byte) (usecase.PageOneInteractorIF, error) {
			return usecase.RestorePageOneInteractor(data, new(presenter.PageOneWebPresenter))
		},
		controller.NewPageOneWebControllerWithProvider)
	games.RegisterKVGame("president", games.CategoryClassic,
		func() usecase.PresidentInteractorIF {
			return usecase.NewPresidentInteractor(domain.NewDefaultPresident(), new(presenter.PresidentWebPresenter))
		},
		func(data []byte) (usecase.PresidentInteractorIF, error) {
			return usecase.RestorePresidentInteractor(data, new(presenter.PresidentWebPresenter))
		},
		controller.NewPresidentWebControllerWithProvider)
	games.RegisterKVGame("shithead", games.CategoryClassic,
		func() usecase.ShitheadInteractorIF {
			return usecase.NewShitheadInteractor(domain.NewDefaultShithead(), new(presenter.ShitheadWebPresenter))
		},
		func(data []byte) (usecase.ShitheadInteractorIF, error) {
			return usecase.RestoreShitheadInteractor(data, new(presenter.ShitheadWebPresenter))
		},
		controller.NewShitheadWebControllerWithProvider)
	games.RegisterKVGame("slapjack", games.CategoryClassic,
		func() usecase.SlapjackInteractorIF {
			return usecase.NewSlapjackInteractor(domain.NewDefaultSlapjack(), new(presenter.SlapjackWebPresenter))
		},
		func(data []byte) (usecase.SlapjackInteractorIF, error) {
			return usecase.RestoreSlapjackInteractor(data, new(presenter.SlapjackWebPresenter))
		},
		controller.NewSlapjackWebControllerWithProvider)
	games.RegisterKVGame("egyptianratscrew", games.CategoryClassic,
		func() usecase.EgyptianRatscrewInteractorIF {
			return usecase.NewEgyptianRatscrewInteractor(domain.NewDefaultEgyptianRatscrew(), new(presenter.EgyptianRatscrewWebPresenter))
		},
		func(data []byte) (usecase.EgyptianRatscrewInteractorIF, error) {
			return usecase.RestoreEgyptianRatscrewInteractor(data, new(presenter.EgyptianRatscrewWebPresenter))
		},
		controller.NewEgyptianRatscrewWebControllerWithProvider)
	games.RegisterKVGame("tonk", games.CategoryClassic,
		func() usecase.TonkInteractorIF {
			return usecase.NewTonkInteractor(domain.NewDefaultTonk(), new(presenter.TonkWebPresenter))
		},
		func(data []byte) (usecase.TonkInteractorIF, error) {
			return usecase.RestoreTonkInteractor(data, new(presenter.TonkWebPresenter))
		},
		controller.NewTonkWebControllerWithProvider)
	games.RegisterKVGame("callbreak", games.CategoryClassic,
		func() usecase.CallBreakInteractorIF {
			return usecase.NewCallBreakInteractor(domain.NewDefaultCallBreak(), new(presenter.CallBreakWebPresenter))
		},
		func(data []byte) (usecase.CallBreakInteractorIF, error) {
			return usecase.RestoreCallBreakInteractor(data, new(presenter.CallBreakWebPresenter))
		},
		controller.NewCallBreakWebControllerWithProvider)
	games.RegisterKVGame("briscola", games.CategoryClassic,
		func() usecase.BriscolaInteractorIF {
			return usecase.NewBriscolaInteractor(domain.NewDefaultBriscola(), new(presenter.BriscolaWebPresenter))
		},
		func(data []byte) (usecase.BriscolaInteractorIF, error) {
			return usecase.RestoreBriscolaInteractor(data, new(presenter.BriscolaWebPresenter))
		},
		controller.NewBriscolaWebControllerWithProvider)
	games.RegisterKVGame("truco", games.CategoryClassic,
		func() usecase.TrucoInteractorIF {
			return usecase.NewTrucoInteractor(domain.NewDefaultTruco(), new(presenter.TrucoWebPresenter))
		},
		func(data []byte) (usecase.TrucoInteractorIF, error) {
			return usecase.RestoreTrucoInteractor(data, new(presenter.TrucoWebPresenter))
		},
		controller.NewTrucoWebControllerWithProvider)
	games.RegisterKVGame("marias", games.CategoryClassic,
		func() usecase.MariasInteractorIF {
			return usecase.NewMariasInteractor(domain.NewDefaultMarias(), new(presenter.MariasWebPresenter))
		},
		func(data []byte) (usecase.MariasInteractorIF, error) {
			return usecase.RestoreMariasInteractor(data, new(presenter.MariasWebPresenter))
		},
		controller.NewMariasWebControllerWithProvider)
	games.RegisterKVGame("sedma", games.CategoryClassic,
		func() usecase.SedmaInteractorIF {
			return usecase.NewSedmaInteractor(domain.NewDefaultSedma(), new(presenter.SedmaWebPresenter))
		},
		func(data []byte) (usecase.SedmaInteractorIF, error) {
			return usecase.RestoreSedmaInteractor(data, new(presenter.SedmaWebPresenter))
		},
		controller.NewSedmaWebControllerWithProvider)

	games.RegisterKVGame("shamrocks", games.CategoryClassic,
		func() usecase.ShamrocksInteractorIF {
			return usecase.NewShamrocksInteractor(domain.NewDefaultShamrocks(), new(presenter.ShamrocksWebPresenter))
		},
		func(data []byte) (usecase.ShamrocksInteractorIF, error) {
			return usecase.RestoreShamrocksInteractor(data, new(presenter.ShamrocksWebPresenter))
		},
		controller.NewShamrocksWebControllerWithProvider)



	games.RegisterKVGame("prsi", games.CategoryClassic,
		func() usecase.PrsiInteractorIF {
			return usecase.NewPrsiInteractor(domain.NewDefaultPrsi(), new(presenter.PrsiWebPresenter))
		},
		func(data []byte) (usecase.PrsiInteractorIF, error) {
			return usecase.RestorePrsiInteractor(data, new(presenter.PrsiWebPresenter))
		},
		controller.NewPrsiWebControllerWithProvider)

	games.RegisterKVGame("unsunkaruta", games.CategoryClassic,
		func() usecase.UnsunKarutaInteractorIF {
			return usecase.NewUnsunKarutaInteractor(domain.NewDefaultUnsunKaruta(), new(presenter.UnsunKarutaWebPresenter))
		},
		func(data []byte) (usecase.UnsunKarutaInteractorIF, error) {
			return usecase.RestoreUnsunKarutaInteractor(data, new(presenter.UnsunKarutaWebPresenter))
		},
		controller.NewUnsunKarutaWebControllerWithProvider)
	games.RegisterKVGame("karnoffel", games.CategoryClassic,
		func() usecase.KarnoffelInteractorIF {
			return usecase.NewKarnoffelInteractor(domain.NewDefaultKarnoffel(), new(presenter.KarnoffelWebPresenter))
		},
		func(data []byte) (usecase.KarnoffelInteractorIF, error) {
			return usecase.RestoreKarnoffelInteractor(data, new(presenter.KarnoffelWebPresenter))
		},
		controller.NewKarnoffelWebControllerWithProvider)
	games.RegisterKVGame("colorado", games.CategoryClassic,
		func() usecase.ColoradoInteractorIF {
			return usecase.NewColoradoInteractor(domain.NewDefaultColorado(), new(presenter.ColoradoWebPresenter))
		},
		func(data []byte) (usecase.ColoradoInteractorIF, error) {
			return usecase.RestoreColoradoInteractor(data, new(presenter.ColoradoWebPresenter))
		},
		controller.NewColoradoWebControllerWithProvider)
	games.RegisterKVGame("royalcotillion", games.CategoryClassic,
		func() usecase.RoyalCotillionInteractorIF {
			return usecase.NewRoyalCotillionInteractor(domain.NewDefaultRoyalCotillion(), new(presenter.RoyalCotillionWebPresenter))
		},
		func(data []byte) (usecase.RoyalCotillionInteractorIF, error) {
			return usecase.RestoreRoyalCotillionInteractor(data, new(presenter.RoyalCotillionWebPresenter))
		},
		controller.NewRoyalCotillionWebControllerWithProvider)
	games.RegisterKVGame("cucumber", games.CategoryClassic,
		func() usecase.CucumberInteractorIF {
			return usecase.NewCucumberInteractor(domain.NewDefaultCucumber(), new(presenter.CucumberWebPresenter))
		},
		func(data []byte) (usecase.CucumberInteractorIF, error) {
			return usecase.RestoreCucumberInteractor(data, new(presenter.CucumberWebPresenter))
		},
		controller.NewCucumberWebControllerWithProvider)
	games.RegisterKVGame("ginrummy", games.CategoryClassic,
		func() usecase.GinRummyInteractorIF {
			return usecase.NewGinRummyInteractor(domain.NewDefaultGinRummy(), new(presenter.GinRummyWebPresenter))
		},
		func(data []byte) (usecase.GinRummyInteractorIF, error) {
			return usecase.RestoreGinRummyInteractor(data, new(presenter.GinRummyWebPresenter))
		},
		controller.NewGinRummyWebControllerWithProvider)
}
