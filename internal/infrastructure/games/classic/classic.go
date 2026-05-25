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
	games.RegisterKVGame("bigtwo", games.CategoryClassic,
		func() usecase.BigTwoInteractorIF {
			return usecase.NewBigTwoInteractor(domain.NewDefaultBigTwo(), new(presenter.BigTwoWebPresenter))
		},
		func(data []byte) (usecase.BigTwoInteractorIF, error) {
			return usecase.RestoreBigTwoInteractor(data, new(presenter.BigTwoWebPresenter))
		},
		controller.NewBigTwoWebControllerWithProvider)
	games.RegisterKVGame("sevens", games.CategoryClassic,
		func() usecase.SevensInteractorIF {
			return usecase.NewSevensInteractor(domain.NewDefaultSevens(), new(presenter.SevensWebPresenter))
		},
		func(data []byte) (usecase.SevensInteractorIF, error) {
			return usecase.RestoreSevensInteractor(data, new(presenter.SevensWebPresenter))
		},
		controller.NewSevensWebControllerWithProvider)
	games.RegisterKVGame("doubt", games.CategoryClassic,
		func() usecase.DoubtInteractorIF {
			return usecase.NewDoubtInteractor(domain.NewDefaultDoubt(), new(presenter.DoubtWebPresenter))
		},
		func(data []byte) (usecase.DoubtInteractorIF, error) {
			return usecase.RestoreDoubtInteractor(data, new(presenter.DoubtWebPresenter))
		},
		controller.NewDoubtWebControllerWithProvider)
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
	games.RegisterKVGame("napoleon", games.CategoryClassic,
		func() usecase.NapoleonInteractorIF {
			return usecase.NewNapoleonInteractor(domain.NewDefaultNapoleon(), new(presenter.NapoleonWebPresenter))
		},
		func(data []byte) (usecase.NapoleonInteractorIF, error) {
			return usecase.RestoreNapoleonInteractor(data, new(presenter.NapoleonWebPresenter))
		},
		controller.NewNapoleonWebControllerWithProvider)
	games.RegisterKVGame("euchre", games.CategoryClassic,
		func() usecase.EuchreInteractorIF {
			return usecase.NewEuchreInteractor(domain.NewDefaultEuchre(), new(presenter.EuchreWebPresenter))
		},
		func(data []byte) (usecase.EuchreInteractorIF, error) {
			return usecase.RestoreEuchreInteractor(data, new(presenter.EuchreWebPresenter))
		},
		controller.NewEuchreWebControllerWithProvider)
	games.RegisterKVGame("ohhell", games.CategoryClassic,
		func() usecase.OhHellInteractorIF {
			return usecase.NewOhHellInteractor(domain.NewDefaultOhHell(), new(presenter.OhHellWebPresenter))
		},
		func(data []byte) (usecase.OhHellInteractorIF, error) {
			return usecase.RestoreOhHellInteractor(data, new(presenter.OhHellWebPresenter))
		},
		controller.NewOhHellWebControllerWithProvider)
	games.RegisterKVGame("bridge", games.CategoryClassic,
		func() usecase.BridgeInteractorIF {
			return usecase.NewBridgeInteractor(domain.NewDefaultBridge(), new(presenter.BridgeWebPresenter))
		},
		func(data []byte) (usecase.BridgeInteractorIF, error) {
			return usecase.RestoreBridgeInteractor(data, new(presenter.BridgeWebPresenter))
		},
		controller.NewBridgeWebControllerWithProvider)
	games.RegisterKVGame("speed", games.CategoryClassic,
		func() usecase.SpeedInteractorIF {
			return usecase.NewSpeedInteractor(domain.NewDefaultSpeed(), new(presenter.SpeedWebPresenter))
		},
		func(data []byte) (usecase.SpeedInteractorIF, error) {
			return usecase.RestoreSpeedInteractor(data, new(presenter.SpeedWebPresenter))
		},
		controller.NewSpeedWebControllerWithProvider)
	games.RegisterKVGame("gofish", games.CategoryClassic,
		func() usecase.GoFishInteractorIF {
			return usecase.NewGoFishInteractor(domain.NewDefaultGoFish(), new(presenter.GoFishWebPresenter))
		},
		func(data []byte) (usecase.GoFishInteractorIF, error) {
			return usecase.RestoreGoFishInteractor(data, new(presenter.GoFishWebPresenter))
		},
		controller.NewGoFishWebControllerWithProvider)
	games.RegisterKVGame("pinochle", games.CategoryClassic,
		func() usecase.PinochleInteractorIF {
			return usecase.NewPinochleInteractor(domain.NewDefaultPinochle(), new(presenter.PinochleWebPresenter))
		},
		func(data []byte) (usecase.PinochleInteractorIF, error) {
			return usecase.RestorePinochleInteractor(data, new(presenter.PinochleWebPresenter))
		},
		controller.NewPinochleWebControllerWithProvider)
	games.RegisterKVGame("pigtail", games.CategoryClassic,
		func() usecase.PigsTailInteractorIF {
			return usecase.NewPigsTailInteractor(domain.NewDefaultPigsTail(), new(presenter.PigsTailWebPresenter))
		},
		func(data []byte) (usecase.PigsTailInteractorIF, error) {
			return usecase.RestorePigsTailInteractor(data, new(presenter.PigsTailWebPresenter))
		},
		controller.NewPigsTailWebControllerWithProvider)
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
	games.RegisterKVGame("war", games.CategoryClassic,
		func() usecase.WarInteractorIF {
			return usecase.NewWarInteractor(domain.NewDefaultWar(), new(presenter.WarWebPresenter))
		},
		func(data []byte) (usecase.WarInteractorIF, error) {
			return usecase.RestoreWarInteractor(data, new(presenter.WarWebPresenter))
		},
		controller.NewWarWebControllerWithProvider)
	games.RegisterKVGame("fiftyone", games.CategoryClassic,
		func() usecase.FiftyOneInteractorIF {
			return usecase.NewFiftyOneInteractor(domain.NewDefaultFiftyOne(), new(presenter.FiftyOneWebPresenter))
		},
		func(data []byte) (usecase.FiftyOneInteractorIF, error) {
			return usecase.RestoreFiftyOneInteractor(data, new(presenter.FiftyOneWebPresenter))
		},
		controller.NewFiftyOneWebControllerWithProvider)
	games.RegisterKVGame("whist", games.CategoryClassic,
		func() usecase.WhistInteractorIF {
			return usecase.NewWhistInteractor(domain.NewDefaultWhist(), new(presenter.WhistWebPresenter))
		},
		func(data []byte) (usecase.WhistInteractorIF, error) {
			return usecase.RestoreWhistInteractor(data, new(presenter.WhistWebPresenter))
		},
		controller.NewWhistWebControllerWithProvider)
	games.RegisterKVGame("pageone", games.CategoryClassic,
		func() usecase.PageOneInteractorIF {
			return usecase.NewPageOneInteractor(domain.NewDefaultPageOne(), new(presenter.PageOneWebPresenter))
		},
		func(data []byte) (usecase.PageOneInteractorIF, error) {
			return usecase.RestorePageOneInteractor(data, new(presenter.PageOneWebPresenter))
		},
		controller.NewPageOneWebControllerWithProvider)
	games.RegisterKVGame("trash", games.CategoryClassic,
		func() usecase.TrashInteractorIF {
			return usecase.NewTrashInteractor(domain.NewDefaultTrash(), new(presenter.TrashWebPresenter))
		},
		func(data []byte) (usecase.TrashInteractorIF, error) {
			return usecase.RestoreTrashInteractor(data, new(presenter.TrashWebPresenter))
		},
		controller.NewTrashWebControllerWithProvider)
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
	games.RegisterKVGame("spiteandmalice", games.CategoryClassic,
		func() usecase.SpiteAndMaliceInteractorIF {
			return usecase.NewSpiteAndMaliceInteractor(domain.NewDefaultSpiteAndMalice(), new(presenter.SpiteAndMaliceWebPresenter))
		},
		func(data []byte) (usecase.SpiteAndMaliceInteractorIF, error) {
			return usecase.RestoreSpiteAndMaliceInteractor(data, new(presenter.SpiteAndMaliceWebPresenter))
		},
		controller.NewSpiteAndMaliceWebControllerWithProvider)
	games.RegisterKVGame("skat", games.CategoryClassic,
		func() usecase.SkatInteractorIF {
			return usecase.NewSkatInteractor(domain.NewDefaultSkat(), new(presenter.SkatWebPresenter))
		},
		func(data []byte) (usecase.SkatInteractorIF, error) {
			return usecase.RestoreSkatInteractor(data, new(presenter.SkatWebPresenter))
		},
		controller.NewSkatWebControllerWithProvider)
	games.RegisterKVGame("shithead", games.CategoryClassic,
		func() usecase.ShitheadInteractorIF {
			return usecase.NewShitheadInteractor(domain.NewDefaultShithead(), new(presenter.ShitheadWebPresenter))
		},
		func(data []byte) (usecase.ShitheadInteractorIF, error) {
			return usecase.RestoreShitheadInteractor(data, new(presenter.ShitheadWebPresenter))
		},
		controller.NewShitheadWebControllerWithProvider)
	games.RegisterKVGame("nertz", games.CategoryClassic,
		func() usecase.NertzInteractorIF {
			return usecase.NewNertzInteractor(domain.NewDefaultNertz(), new(presenter.NertzWebPresenter))
		},
		func(data []byte) (usecase.NertzInteractorIF, error) {
			return usecase.RestoreNertzInteractor(data, new(presenter.NertzWebPresenter))
		},
		controller.NewNertzWebControllerWithProvider)
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
	games.RegisterKVGame("belote", games.CategoryClassic,
		func() usecase.BeloteInteractorIF {
			return usecase.NewBeloteInteractor(domain.NewDefaultBelote(), new(presenter.BeloteWebPresenter))
		},
		func(data []byte) (usecase.BeloteInteractorIF, error) {
			return usecase.RestoreBeloteInteractor(data, new(presenter.BeloteWebPresenter))
		},
		controller.NewBeloteWebControllerWithProvider)
	games.RegisterKVGame("mighty", games.CategoryClassic,
		func() usecase.MightyInteractorIF {
			return usecase.NewMightyInteractor(domain.NewDefaultMighty(), new(presenter.MightyWebPresenter))
		},
		func(data []byte) (usecase.MightyInteractorIF, error) {
			return usecase.RestoreMightyInteractor(data, new(presenter.MightyWebPresenter))
		},
		controller.NewMightyWebControllerWithProvider)
	games.RegisterKVGame("piquet", games.CategoryClassic,
		func() usecase.PiquetInteractorIF {
			return usecase.NewPiquetInteractor(domain.NewDefaultPiquet(), new(presenter.PiquetWebPresenter))
		},
		func(data []byte) (usecase.PiquetInteractorIF, error) {
			return usecase.RestorePiquetInteractor(data, new(presenter.PiquetWebPresenter))
		},
		controller.NewPiquetWebControllerWithProvider)
	games.RegisterKVGame("callbreak", games.CategoryClassic,
		func() usecase.CallBreakInteractorIF {
			return usecase.NewCallBreakInteractor(domain.NewDefaultCallBreak(), new(presenter.CallBreakWebPresenter))
		},
		func(data []byte) (usecase.CallBreakInteractorIF, error) {
			return usecase.RestoreCallBreakInteractor(data, new(presenter.CallBreakWebPresenter))
		},
		controller.NewCallBreakWebControllerWithProvider)
	games.RegisterKVGame("tarneeb", games.CategoryClassic,
		func() usecase.TarneebInteractorIF {
			return usecase.NewTarneebInteractor(domain.NewDefaultTarneeb(), new(presenter.TarneebWebPresenter))
		},
		func(data []byte) (usecase.TarneebInteractorIF, error) {
			return usecase.RestoreTarneebInteractor(data, new(presenter.TarneebWebPresenter))
		},
		controller.NewTarneebWebControllerWithProvider)
	games.RegisterKVGame("briscola", games.CategoryClassic,
		func() usecase.BriscolaInteractorIF {
			return usecase.NewBriscolaInteractor(domain.NewDefaultBriscola(), new(presenter.BriscolaWebPresenter))
		},
		func(data []byte) (usecase.BriscolaInteractorIF, error) {
			return usecase.RestoreBriscolaInteractor(data, new(presenter.BriscolaWebPresenter))
		},
		controller.NewBriscolaWebControllerWithProvider)
}
