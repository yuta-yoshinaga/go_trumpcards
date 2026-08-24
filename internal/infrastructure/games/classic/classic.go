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
	games.RegisterKVGame("daifugo", games.CategoryClassic,
		func() usecase.DaifugoInteractorIF {
			return usecase.NewDaifugoInteractor(domain.NewDefaultDaifugo(), new(presenter.DaifugoWebPresenter))
		},
		func(data []byte) (usecase.DaifugoInteractorIF, error) {
			return usecase.RestoreDaifugoInteractor(data, new(presenter.DaifugoWebPresenter))
		},
		controller.NewDaifugoWebControllerWithProvider)
	games.RegisterKVGame("sevens", games.CategoryClassic,
		func() usecase.SevensInteractorIF {
			return usecase.NewSevensInteractor(domain.NewDefaultSevens(), new(presenter.SevensWebPresenter))
		},
		func(data []byte) (usecase.SevensInteractorIF, error) {
			return usecase.RestoreSevensInteractor(data, new(presenter.SevensWebPresenter))
		},
		controller.NewSevensWebControllerWithProvider)
	games.RegisterKVGame("hearts", games.CategoryClassic,
		func() usecase.HeartsInteractorIF {
			return usecase.NewHeartsInteractor(domain.NewDefaultHearts(), new(presenter.HeartsWebPresenter))
		},
		func(data []byte) (usecase.HeartsInteractorIF, error) {
			return usecase.RestoreHeartsInteractor(data, new(presenter.HeartsWebPresenter))
		},
		controller.NewHeartsWebControllerWithProvider)
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
	games.RegisterKVGame("cassino", games.CategoryClassic,
		func() usecase.CassinoInteractorIF {
			return usecase.NewCassinoInteractor(domain.NewDefaultCassino(), new(presenter.CassinoWebPresenter))
		},
		func(data []byte) (usecase.CassinoInteractorIF, error) {
			return usecase.RestoreCassinoInteractor(data, new(presenter.CassinoWebPresenter))
		},
		controller.NewCassinoWebControllerWithProvider)
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
	games.RegisterKVGame("pitch", games.CategoryClassic,
		func() usecase.PitchInteractorIF {
			return usecase.NewPitchInteractor(domain.NewDefaultPitch(), new(presenter.PitchWebPresenter))
		},
		func(data []byte) (usecase.PitchInteractorIF, error) {
			return usecase.RestorePitchInteractor(data, new(presenter.PitchWebPresenter))
		},
		controller.NewPitchWebControllerWithProvider)
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
	games.RegisterKVGame("brusquembille", games.CategoryClassic,
		func() usecase.BrusquembilleInteractorIF {
			return usecase.NewBrusquembilleInteractor(domain.NewDefaultBrusquembille(), new(presenter.BrusquembilleWebPresenter))
		},
		func(data []byte) (usecase.BrusquembilleInteractorIF, error) {
			return usecase.RestoreBrusquembilleInteractor(data, new(presenter.BrusquembilleWebPresenter))
		},
		controller.NewBrusquembilleWebControllerWithProvider)
	games.RegisterKVGame("truco", games.CategoryClassic,
		func() usecase.TrucoInteractorIF {
			return usecase.NewTrucoInteractor(domain.NewDefaultTruco(), new(presenter.TrucoWebPresenter))
		},
		func(data []byte) (usecase.TrucoInteractorIF, error) {
			return usecase.RestoreTrucoInteractor(data, new(presenter.TrucoWebPresenter))
		},
		controller.NewTrucoWebControllerWithProvider)
	games.RegisterKVGame("klaverjas", games.CategoryClassic,
		func() usecase.KlaverjasInteractorIF {
			return usecase.NewKlaverjasInteractor(domain.NewDefaultKlaverjas(), new(presenter.KlaverjasWebPresenter))
		},
		func(data []byte) (usecase.KlaverjasInteractorIF, error) {
			return usecase.RestoreKlaverjasInteractor(data, new(presenter.KlaverjasWebPresenter))
		},
		controller.NewKlaverjasWebControllerWithProvider)
	games.RegisterKVGame("manille", games.CategoryClassic,
		func() usecase.ManilleInteractorIF {
			return usecase.NewManilleInteractor(domain.NewDefaultManille(), new(presenter.ManilleWebPresenter))
		},
		func(data []byte) (usecase.ManilleInteractorIF, error) {
			return usecase.RestoreManilleInteractor(data, new(presenter.ManilleWebPresenter))
		},
		controller.NewManilleWebControllerWithProvider)
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
	games.RegisterKVGame("solowhist", games.CategoryClassic,
		func() usecase.SoloWhistInteractorIF {
			return usecase.NewSoloWhistInteractor(domain.NewDefaultSoloWhist(), new(presenter.SoloWhistWebPresenter))
		},
		func(data []byte) (usecase.SoloWhistInteractorIF, error) {
			return usecase.RestoreSoloWhistInteractor(data, new(presenter.SoloWhistWebPresenter))
		},
		controller.NewSoloWhistWebControllerWithProvider)
	games.RegisterKVGame("knockoutwhist", games.CategoryClassic,
		func() usecase.KnockoutWhistInteractorIF {
			return usecase.NewKnockoutWhistInteractor(domain.NewDefaultKnockoutWhist(), new(presenter.KnockoutWhistWebPresenter))
		},
		func(data []byte) (usecase.KnockoutWhistInteractorIF, error) {
			return usecase.RestoreKnockoutWhistInteractor(data, new(presenter.KnockoutWhistWebPresenter))
		},
		controller.NewKnockoutWhistWebControllerWithProvider)
	games.RegisterKVGame("nap", games.CategoryClassic,
		func() usecase.NapInteractorIF {
			return usecase.NewNapInteractor(domain.NewDefaultNap(), new(presenter.NapWebPresenter))
		},
		func(data []byte) (usecase.NapInteractorIF, error) {
			return usecase.RestoreNapInteractor(data, new(presenter.NapWebPresenter))
		},
		controller.NewNapWebControllerWithProvider)
	games.RegisterKVGame("spoilfive", games.CategoryClassic,
		func() usecase.SpoilFiveInteractorIF {
			return usecase.NewSpoilFiveInteractor(domain.NewDefaultSpoilFive(), new(presenter.SpoilFiveWebPresenter))
		},
		func(data []byte) (usecase.SpoilFiveInteractorIF, error) {
			return usecase.RestoreSpoilFiveInteractor(data, new(presenter.SpoilFiveWebPresenter))
		},
		controller.NewSpoilFiveWebControllerWithProvider)
	games.RegisterKVGame("scopa", games.CategoryClassic,
		func() usecase.ScopaInteractorIF {
			return usecase.NewScopaInteractor(domain.NewDefaultScopa(), new(presenter.ScopaWebPresenter))
		},
		func(data []byte) (usecase.ScopaInteractorIF, error) {
			return usecase.RestoreScopaInteractor(data, new(presenter.ScopaWebPresenter))
		},
		controller.NewScopaWebControllerWithProvider)
	games.RegisterKVGame("scopone", games.CategoryClassic,
		func() usecase.ScoponeInteractorIF {
			return usecase.NewScoponeInteractor(domain.NewDefaultScopone(), new(presenter.ScoponeWebPresenter))
		},
		func(data []byte) (usecase.ScoponeInteractorIF, error) {
			return usecase.RestoreScoponeInteractor(data, new(presenter.ScoponeWebPresenter))
		},
		controller.NewScoponeWebControllerWithProvider)
	games.RegisterKVGame("escoba", games.CategoryClassic,
		func() usecase.EscobaInteractorIF {
			return usecase.NewEscobaInteractor(domain.NewDefaultEscoba(), new(presenter.EscobaWebPresenter))
		},
		func(data []byte) (usecase.EscobaInteractorIF, error) {
			return usecase.RestoreEscobaInteractor(data, new(presenter.EscobaWebPresenter))
		},
		controller.NewEscobaWebControllerWithProvider)

	games.RegisterKVGame("shamrocks", games.CategoryClassic,
		func() usecase.ShamrocksInteractorIF {
			return usecase.NewShamrocksInteractor(domain.NewDefaultShamrocks(), new(presenter.ShamrocksWebPresenter))
		},
		func(data []byte) (usecase.ShamrocksInteractorIF, error) {
			return usecase.RestoreShamrocksInteractor(data, new(presenter.ShamrocksWebPresenter))
		},
		controller.NewShamrocksWebControllerWithProvider)
	games.RegisterKVGame("labellelucie", games.CategoryClassic,
		func() usecase.LaBelleLucieInteractorIF {
			return usecase.NewLaBelleLucieInteractor(domain.NewDefaultLaBelleLucie(), new(presenter.LaBelleLucieWebPresenter))
		},
		func(data []byte) (usecase.LaBelleLucieInteractorIF, error) {
			return usecase.RestoreLaBelleLucieInteractor(data, new(presenter.LaBelleLucieWebPresenter))
		},
		controller.NewLaBelleLucieWebControllerWithProvider)

	games.RegisterKVGame("curdsandwhey", games.CategoryClassic,
		func() usecase.CurdsAndWheyInteractorIF {
			return usecase.NewCurdsAndWheyInteractor(domain.NewDefaultCurdsAndWhey(), new(presenter.CurdsAndWheyWebPresenter))
		},
		func(data []byte) (usecase.CurdsAndWheyInteractorIF, error) {
			return usecase.RestoreCurdsAndWheyInteractor(data, new(presenter.CurdsAndWheyWebPresenter))
		},
		controller.NewCurdsAndWheyWebControllerWithProvider)
	games.RegisterKVGame("simplesimon", games.CategoryClassic,
		func() usecase.SimpleSimonInteractorIF {
			return usecase.NewSimpleSimonInteractor(domain.NewDefaultSimpleSimon(), new(presenter.SimpleSimonWebPresenter))
		},
		func(data []byte) (usecase.SimpleSimonInteractorIF, error) {
			return usecase.RestoreSimpleSimonInteractor(data, new(presenter.SimpleSimonWebPresenter))
		},
		controller.NewSimpleSimonWebControllerWithProvider)

	games.RegisterKVGame("allfours", games.CategoryClassic,
		func() usecase.AllFoursInteractorIF {
			return usecase.NewAllFoursInteractor(domain.NewDefaultAllFours(), new(presenter.AllFoursWebPresenter))
		},
		func(data []byte) (usecase.AllFoursInteractorIF, error) {
			return usecase.RestoreAllFoursInteractor(data, new(presenter.AllFoursWebPresenter))
		},
		controller.NewAllFoursWebControllerWithProvider)

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
	games.RegisterKVGame("germanwhist", games.CategoryClassic,
		func() usecase.GermanWhistInteractorIF {
			return usecase.NewGermanWhistInteractor(domain.NewDefaultGermanWhist(), new(presenter.GermanWhistWebPresenter))
		},
		func(data []byte) (usecase.GermanWhistInteractorIF, error) {
			return usecase.RestoreGermanWhistInteractor(data, new(presenter.GermanWhistWebPresenter))
		},
		controller.NewGermanWhistWebControllerWithProvider)
	games.RegisterKVGame("slobberhannes", games.CategoryClassic,
		func() usecase.SlobberhannesInteractorIF {
			return usecase.NewSlobberhannesInteractor(domain.NewDefaultSlobberhannes(), new(presenter.SlobberhannesWebPresenter))
		},
		func(data []byte) (usecase.SlobberhannesInteractorIF, error) {
			return usecase.RestoreSlobberhannesInteractor(data, new(presenter.SlobberhannesWebPresenter))
		},
		controller.NewSlobberhannesWebControllerWithProvider)
	games.RegisterKVGame("reversis", games.CategoryClassic,
		func() usecase.ReversisInteractorIF {
			return usecase.NewReversisInteractor(domain.NewDefaultReversis(), new(presenter.ReversisWebPresenter))
		},
		func(data []byte) (usecase.ReversisInteractorIF, error) {
			return usecase.RestoreReversisInteractor(data, new(presenter.ReversisWebPresenter))
		},
		controller.NewReversisWebControllerWithProvider)
	games.RegisterKVGame("hokm", games.CategoryClassic,
		func() usecase.HokmInteractorIF {
			return usecase.NewHokmInteractor(domain.NewDefaultHokm(), new(presenter.HokmWebPresenter))
		},
		func(data []byte) (usecase.HokmInteractorIF, error) {
			return usecase.RestoreHokmInteractor(data, new(presenter.HokmWebPresenter))
		},
		controller.NewHokmWebControllerWithProvider)
	games.RegisterKVGame("cucumber", games.CategoryClassic,
		func() usecase.CucumberInteractorIF {
			return usecase.NewCucumberInteractor(domain.NewDefaultCucumber(), new(presenter.CucumberWebPresenter))
		},
		func(data []byte) (usecase.CucumberInteractorIF, error) {
			return usecase.RestoreCucumberInteractor(data, new(presenter.CucumberWebPresenter))
		},
		controller.NewCucumberWebControllerWithProvider)
	games.RegisterKVGame("botifarra", games.CategoryClassic,
		func() usecase.BotifarraInteractorIF {
			return usecase.NewBotifarraInteractor(domain.NewDefaultBotifarra(), new(presenter.BotifarraWebPresenter))
		},
		func(data []byte) (usecase.BotifarraInteractorIF, error) {
			return usecase.RestoreBotifarraInteractor(data, new(presenter.BotifarraWebPresenter))
		},
		controller.NewBotifarraWebControllerWithProvider)
	games.RegisterKVGame("germansolo", games.CategoryClassic,
		func() usecase.GermanSoloInteractorIF {
			return usecase.NewGermanSoloInteractor(domain.NewDefaultGermanSolo(), new(presenter.GermanSoloWebPresenter))
		},
		func(data []byte) (usecase.GermanSoloInteractorIF, error) {
			return usecase.RestoreGermanSoloInteractor(data, new(presenter.GermanSoloWebPresenter))
		},
		controller.NewGermanSoloWebControllerWithProvider)
}
