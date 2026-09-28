//go:build js && wasm

// Package extra7 reserves the tenth Cloudflare Worker size bucket.
package extra7

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func init() {
	games.RegisterKVGame("biriba", games.CategoryExtra7,
		func() usecase.BiribaInteractorIF {
			return usecase.NewBiribaInteractor(domain.NewDefaultBiriba(), new(presenter.BiribaWebPresenter))
		},
		func(data []byte) (usecase.BiribaInteractorIF, error) {
			return usecase.RestoreBiribaInteractor(data, new(presenter.BiribaWebPresenter))
		},
		controller.NewBiribaWebControllerWithProvider)
	games.RegisterKVGame("burraco", games.CategoryExtra7,
		func() usecase.BurracoInteractorIF {
			return usecase.NewBurracoInteractor(domain.NewDefaultBurraco(), new(presenter.BurracoWebPresenter))
		},
		func(data []byte) (usecase.BurracoInteractorIF, error) {
			return usecase.RestoreBurracoInteractor(data, new(presenter.BurracoWebPresenter))
		},
		controller.NewBurracoWebControllerWithProvider)
	games.RegisterKVGame("canasta", games.CategoryExtra7,
		func() usecase.CanastaInteractorIF {
			return usecase.NewCanastaInteractor(domain.NewDefaultCanasta(), new(presenter.CanastaWebPresenter))
		},
		func(data []byte) (usecase.CanastaInteractorIF, error) {
			return usecase.RestoreCanastaInteractor(data, new(presenter.CanastaWebPresenter))
		},
		controller.NewCanastaWebControllerWithProvider)
	games.RegisterKVGame("handandfoot", games.CategoryExtra7,
		func() usecase.HandAndFootInteractorIF {
			return usecase.NewHandAndFootInteractor(domain.NewDefaultHandAndFoot(), new(presenter.HandAndFootWebPresenter))
		},
		func(data []byte) (usecase.HandAndFootInteractorIF, error) {
			return usecase.RestoreHandAndFootInteractor(data, new(presenter.HandAndFootWebPresenter))
		},
		controller.NewHandAndFootWebControllerWithProvider)
	games.RegisterKVGame("samba", games.CategoryExtra7,
		func() usecase.SambaInteractorIF {
			return usecase.NewSambaInteractor(domain.NewDefaultSamba(), new(presenter.SambaWebPresenter))
		},
		func(data []byte) (usecase.SambaInteractorIF, error) {
			return usecase.RestoreSambaInteractor(data, new(presenter.SambaWebPresenter))
		},
		controller.NewSambaWebControllerWithProvider)
	games.RegisterKVGame("chinchon", games.CategoryExtra7,
		func() usecase.ChinchonInteractorIF {
			return usecase.NewChinchonInteractor(domain.NewDefaultChinchon(), new(presenter.ChinchonWebPresenter))
		},
		func(data []byte) (usecase.ChinchonInteractorIF, error) {
			return usecase.RestoreChinchonInteractor(data, new(presenter.ChinchonWebPresenter))
		},
		controller.NewChinchonWebControllerWithProvider)
	games.RegisterKVGame("conquian", games.CategoryExtra7,
		func() usecase.ConquianInteractorIF {
			return usecase.NewConquianInteractor(domain.NewDefaultConquian(), new(presenter.ConquianWebPresenter))
		},
		func(data []byte) (usecase.ConquianInteractorIF, error) {
			return usecase.RestoreConquianInteractor(data, new(presenter.ConquianWebPresenter))
		},
		controller.NewConquianWebControllerWithProvider)
	games.RegisterKVGame("minchiate", games.CategoryExtra7,
		func() usecase.MinchiateInteractorIF {
			return usecase.NewMinchiateInteractor(domain.NewDefaultMinchiate(), new(presenter.MinchiateWebPresenter))
		},
		func(data []byte) (usecase.MinchiateInteractorIF, error) {
			return usecase.RestoreMinchiateInteractor(data, new(presenter.MinchiateWebPresenter))
		},
		controller.NewMinchiateWebControllerWithProvider)
	games.RegisterKVGame("tarocchini", games.CategoryExtra7,
		func() usecase.TarocchiniInteractorIF {
			return usecase.NewTarocchiniInteractor(domain.NewDefaultTarocchini(), new(presenter.TarocchiniWebPresenter))
		},
		func(data []byte) (usecase.TarocchiniInteractorIF, error) {
			return usecase.RestoreTarocchiniInteractor(data, new(presenter.TarocchiniWebPresenter))
		},
		controller.NewTarocchiniWebControllerWithProvider)
	games.RegisterKVGame("thirtyone", games.CategoryExtra7,
		func() usecase.ThirtyOneInteractorIF {
			return usecase.NewThirtyOneInteractor(domain.NewDefaultThirtyOne(), new(presenter.ThirtyOneWebPresenter))
		},
		func(data []byte) (usecase.ThirtyOneInteractorIF, error) {
			return usecase.RestoreThirtyOneInteractor(data, new(presenter.ThirtyOneWebPresenter))
		},
		controller.NewThirtyOneWebControllerWithProvider)
	games.RegisterKVGame("andarbahar", games.CategoryExtra7,
		func() usecase.AndarBaharInteractorIF {
			return usecase.NewAndarBaharInteractor(domain.NewDefaultAndarBahar(), new(presenter.AndarBaharWebPresenter))
		},
		func(data []byte) (usecase.AndarBaharInteractorIF, error) {
			return usecase.RestoreAndarBaharInteractor(data, new(presenter.AndarBaharWebPresenter))
		},
		controller.NewAndarBaharWebControllerWithProvider)
	games.RegisterKVGame("crazyfourpoker", games.CategoryExtra7,
		func() usecase.CrazyFourPokerInteractorIF {
			return usecase.NewCrazyFourPokerInteractor(domain.NewDefaultCrazyFourPoker(), new(presenter.CrazyFourPokerWebPresenter))
		},
		func(data []byte) (usecase.CrazyFourPokerInteractorIF, error) {
			return usecase.RestoreCrazyFourPokerInteractor(data, new(presenter.CrazyFourPokerWebPresenter))
		},
		controller.NewCrazyFourPokerWebControllerWithProvider)
	games.RegisterKVGame("doudizhu", games.CategoryExtra7,
		func() usecase.DoudizhuInteractorIF {
			return usecase.NewDoudizhuInteractor(domain.NewDefaultDoudizhu(), new(presenter.DoudizhuWebPresenter))
		},
		func(data []byte) (usecase.DoudizhuInteractorIF, error) {
			return usecase.RestoreDoudizhuInteractor(data, new(presenter.DoudizhuWebPresenter))
		},
		controller.NewDoudizhuWebControllerWithProvider)
	games.RegisterKVGame("pasur", games.CategoryExtra7,
		func() usecase.PasurInteractorIF {
			return usecase.NewPasurInteractor(domain.NewDefaultPasur(), new(presenter.PasurWebPresenter))
		},
		func(data []byte) (usecase.PasurInteractorIF, error) {
			return usecase.RestorePasurInteractor(data, new(presenter.PasurWebPresenter))
		},
		controller.NewPasurWebControllerWithProvider)
	games.RegisterKVGame("schafkopf", games.CategoryExtra7,
		func() usecase.SchafkopfInteractorIF {
			return usecase.NewSchafkopfInteractor(domain.NewDefaultSchafkopf(), new(presenter.SchafkopfWebPresenter))
		},
		func(data []byte) (usecase.SchafkopfInteractorIF, error) {
			return usecase.RestoreSchafkopfInteractor(data, new(presenter.SchafkopfWebPresenter))
		},
		controller.NewSchafkopfWebControllerWithProvider)
	games.RegisterKVGame("brusquembille", games.CategoryExtra7,
		func() usecase.BrusquembilleInteractorIF {
			return usecase.NewBrusquembilleInteractor(domain.NewDefaultBrusquembille(), new(presenter.BrusquembilleWebPresenter))
		},
		func(data []byte) (usecase.BrusquembilleInteractorIF, error) {
			return usecase.RestoreBrusquembilleInteractor(data, new(presenter.BrusquembilleWebPresenter))
		},
		controller.NewBrusquembilleWebControllerWithProvider)
	games.RegisterKVGame("germanwhist", games.CategoryExtra7,
		func() usecase.GermanWhistInteractorIF {
			return usecase.NewGermanWhistInteractor(domain.NewDefaultGermanWhist(), new(presenter.GermanWhistWebPresenter))
		},
		func(data []byte) (usecase.GermanWhistInteractorIF, error) {
			return usecase.RestoreGermanWhistInteractor(data, new(presenter.GermanWhistWebPresenter))
		},
		controller.NewGermanWhistWebControllerWithProvider)
	games.RegisterKVGame("hearts", games.CategoryExtra7,
		func() usecase.HeartsInteractorIF {
			return usecase.NewHeartsInteractor(domain.NewDefaultHearts(), new(presenter.HeartsWebPresenter))
		},
		func(data []byte) (usecase.HeartsInteractorIF, error) {
			return usecase.RestoreHeartsInteractor(data, new(presenter.HeartsWebPresenter))
		},
		controller.NewHeartsWebControllerWithProvider)
	games.RegisterKVGame("hokm", games.CategoryExtra7,
		func() usecase.HokmInteractorIF {
			return usecase.NewHokmInteractor(domain.NewDefaultHokm(), new(presenter.HokmWebPresenter))
		},
		func(data []byte) (usecase.HokmInteractorIF, error) {
			return usecase.RestoreHokmInteractor(data, new(presenter.HokmWebPresenter))
		},
		controller.NewHokmWebControllerWithProvider)
	games.RegisterKVGame("klaverjas", games.CategoryExtra7,
		func() usecase.KlaverjasInteractorIF {
			return usecase.NewKlaverjasInteractor(domain.NewDefaultKlaverjas(), new(presenter.KlaverjasWebPresenter))
		},
		func(data []byte) (usecase.KlaverjasInteractorIF, error) {
			return usecase.RestoreKlaverjasInteractor(data, new(presenter.KlaverjasWebPresenter))
		},
		controller.NewKlaverjasWebControllerWithProvider)
	games.RegisterKVGame("manille", games.CategoryExtra7,
		func() usecase.ManilleInteractorIF {
			return usecase.NewManilleInteractor(domain.NewDefaultManille(), new(presenter.ManilleWebPresenter))
		},
		func(data []byte) (usecase.ManilleInteractorIF, error) {
			return usecase.RestoreManilleInteractor(data, new(presenter.ManilleWebPresenter))
		},
		controller.NewManilleWebControllerWithProvider)
	games.RegisterKVGame("slobberhannes", games.CategoryExtra7,
		func() usecase.SlobberhannesInteractorIF {
			return usecase.NewSlobberhannesInteractor(domain.NewDefaultSlobberhannes(), new(presenter.SlobberhannesWebPresenter))
		},
		func(data []byte) (usecase.SlobberhannesInteractorIF, error) {
			return usecase.RestoreSlobberhannesInteractor(data, new(presenter.SlobberhannesWebPresenter))
		},
		controller.NewSlobberhannesWebControllerWithProvider)
	games.RegisterKVGame("escoba", games.CategoryExtra7,
		func() usecase.EscobaInteractorIF {
			return usecase.NewEscobaInteractor(domain.NewDefaultEscoba(), new(presenter.EscobaWebPresenter))
		},
		func(data []byte) (usecase.EscobaInteractorIF, error) {
			return usecase.RestoreEscobaInteractor(data, new(presenter.EscobaWebPresenter))
		},
		controller.NewEscobaWebControllerWithProvider)
	games.RegisterKVGame("scopa", games.CategoryExtra7,
		func() usecase.ScopaInteractorIF {
			return usecase.NewScopaInteractor(domain.NewDefaultScopa(), new(presenter.ScopaWebPresenter))
		},
		func(data []byte) (usecase.ScopaInteractorIF, error) {
			return usecase.RestoreScopaInteractor(data, new(presenter.ScopaWebPresenter))
		},
		controller.NewScopaWebControllerWithProvider)
	games.RegisterKVGame("scopone", games.CategoryExtra7,
		func() usecase.ScoponeInteractorIF {
			return usecase.NewScoponeInteractor(domain.NewDefaultScopone(), new(presenter.ScoponeWebPresenter))
		},
		func(data []byte) (usecase.ScoponeInteractorIF, error) {
			return usecase.RestoreScoponeInteractor(data, new(presenter.ScoponeWebPresenter))
		},
		controller.NewScoponeWebControllerWithProvider)
	games.RegisterKVGame("bolivia", games.CategoryExtra7,
		func() usecase.BoliviaInteractorIF {
			return usecase.NewBoliviaInteractor(domain.NewDefaultBolivia(), new(presenter.BoliviaWebPresenter))
		},
		func(data []byte) (usecase.BoliviaInteractorIF, error) {
			return usecase.RestoreBoliviaInteractor(data, new(presenter.BoliviaWebPresenter))
		},
		controller.NewBoliviaWebControllerWithProvider)
}
