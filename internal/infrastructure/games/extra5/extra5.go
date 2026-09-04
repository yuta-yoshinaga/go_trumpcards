//go:build js && wasm

// Package extra5 binds the Cloudflare Worker KV-backed handlers for the games
// assigned to the eighth size bucket. A worker main blank-imports this package
// for its registration side effects, so that whatever it registers is in place
// before games.RegisterCategory is called.
//
// Like casino/classic/solo/extra/extra2/extra3/extra4 this is purely a binary-size
// bucket, not a user-facing taxonomy (ADR-0038). The colourless name is
// deliberate: it holds whatever had to move to keep every TinyGo WASM binary
// under the Cloudflare Workers free-tier 1 MB gzipped limit, and nothing about
// a game's genre says it belongs here.
//
package extra5

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func init() {
	games.RegisterKVGame("daifugo", games.CategoryExtra5,
		func() usecase.DaifugoInteractorIF {
			return usecase.NewDaifugoInteractor(domain.NewDefaultDaifugo(), new(presenter.DaifugoWebPresenter))
		},
		func(data []byte) (usecase.DaifugoInteractorIF, error) {
			return usecase.RestoreDaifugoInteractor(data, new(presenter.DaifugoWebPresenter))
		},
		controller.NewDaifugoWebControllerWithProvider)
	games.RegisterKVGame("diloti", games.CategoryExtra5,
		func() usecase.DilotiInteractorIF {
			return usecase.NewDilotiInteractor(domain.NewDefaultDiloti(), new(presenter.DilotiWebPresenter))
		},
		func(data []byte) (usecase.DilotiInteractorIF, error) {
			return usecase.RestoreDilotiInteractor(data, new(presenter.DilotiWebPresenter))
		},
		controller.NewDilotiWebControllerWithProvider)
	games.RegisterKVGame("pitch", games.CategoryExtra5,
		func() usecase.PitchInteractorIF {
			return usecase.NewPitchInteractor(domain.NewDefaultPitch(), new(presenter.PitchWebPresenter))
		},
		func(data []byte) (usecase.PitchInteractorIF, error) {
			return usecase.RestorePitchInteractor(data, new(presenter.PitchWebPresenter))
		},
		controller.NewPitchWebControllerWithProvider)
	games.RegisterKVGame("quodlibet", games.CategoryExtra5,
		func() usecase.QuodlibetInteractorIF {
			return usecase.NewQuodlibetInteractor(domain.NewDefaultQuodlibet(), new(presenter.QuodlibetWebPresenter))
		},
		func(data []byte) (usecase.QuodlibetInteractorIF, error) {
			return usecase.RestoreQuodlibetInteractor(data, new(presenter.QuodlibetWebPresenter))
		},
		controller.NewQuodlibetWebControllerWithProvider)
	games.RegisterKVGame("teendopaanch", games.CategoryExtra5,
		func() usecase.TeenDoPaanchInteractorIF {
			return usecase.NewTeenDoPaanchInteractor(domain.NewDefaultTeenDoPaanch(), new(presenter.TeenDoPaanchWebPresenter))
		},
		func(data []byte) (usecase.TeenDoPaanchInteractorIF, error) {
			return usecase.RestoreTeenDoPaanchInteractor(data, new(presenter.TeenDoPaanchWebPresenter))
		},
		controller.NewTeenDoPaanchWebControllerWithProvider)
	games.RegisterKVGame("tusac", games.CategoryExtra5,
		func() usecase.TuSacInteractorIF {
			return usecase.NewTuSacInteractor(domain.NewDefaultTuSac(), new(presenter.TuSacWebPresenter))
		},
		func(data []byte) (usecase.TuSacInteractorIF, error) {
			return usecase.RestoreTuSacInteractor(data, new(presenter.TuSacWebPresenter))
		},
		controller.NewTuSacWebControllerWithProvider)
	games.RegisterKVGame("macau", games.CategoryExtra5,
		func() usecase.MacauInteractorIF {
			return usecase.NewMacauInteractor(domain.NewDefaultMacau(), new(presenter.MacauWebPresenter))
		},
		func(data []byte) (usecase.MacauInteractorIF, error) {
			return usecase.RestoreMacauInteractor(data, new(presenter.MacauWebPresenter))
		},
		controller.NewMacauWebControllerWithProvider)
	games.RegisterKVGame("gleek", games.CategoryExtra5,
		func() usecase.GleekInteractorIF {
			return usecase.NewGleekInteractor(domain.NewDefaultGleek(), new(presenter.GleekWebPresenter))
		},
		func(data []byte) (usecase.GleekInteractorIF, error) {
			return usecase.RestoreGleekInteractor(data, new(presenter.GleekWebPresenter))
		},
		controller.NewGleekWebControllerWithProvider)
	games.RegisterKVGame("costlycolours", games.CategoryExtra5,
		func() usecase.CostlyColoursInteractorIF {
			return usecase.NewCostlyColoursInteractor(domain.NewDefaultCostlyColours(), new(presenter.CostlyColoursWebPresenter))
		},
		func(data []byte) (usecase.CostlyColoursInteractorIF, error) {
			return usecase.RestoreCostlyColoursInteractor(data, new(presenter.CostlyColoursWebPresenter))
		},
		controller.NewCostlyColoursWebControllerWithProvider)
	games.RegisterKVGame("dehlapakad", games.CategoryExtra5,
		func() usecase.DehlaPakadInteractorIF {
			return usecase.NewDehlaPakadInteractor(domain.NewDefaultDehlaPakad(), new(presenter.DehlaPakadWebPresenter))
		},
		func(data []byte) (usecase.DehlaPakadInteractorIF, error) {
			return usecase.RestoreDehlaPakadInteractor(data, new(presenter.DehlaPakadWebPresenter))
		},
		controller.NewDehlaPakadWebControllerWithProvider)
	games.RegisterKVGame("goofspiel", games.CategoryExtra5,
		func() usecase.GoofspielInteractorIF {
			return usecase.NewGoofspielInteractor(domain.NewDefaultGoofspiel(), new(presenter.GoofspielWebPresenter))
		},
		func(data []byte) (usecase.GoofspielInteractorIF, error) {
			return usecase.RestoreGoofspielInteractor(data, new(presenter.GoofspielWebPresenter))
		},
		controller.NewGoofspielWebControllerWithProvider)
	games.RegisterKVGame("agnes", games.CategoryExtra5,
		func() usecase.AgnesInteractorIF {
			return usecase.NewAgnesInteractor(domain.NewDefaultAgnes(), new(presenter.AgnesWebPresenter))
		},
		func(data []byte) (usecase.AgnesInteractorIF, error) {
			return usecase.RestoreAgnesInteractor(data, new(presenter.AgnesWebPresenter))
		},
		controller.NewAgnesWebControllerWithProvider)
	games.RegisterKVGame("nertz", games.CategoryExtra5,
		func() usecase.NertzInteractorIF {
			return usecase.NewNertzInteractor(domain.NewDefaultNertz(), new(presenter.NertzWebPresenter))
		},
		func(data []byte) (usecase.NertzInteractorIF, error) {
			return usecase.RestoreNertzInteractor(data, new(presenter.NertzWebPresenter))
		},
		controller.NewNertzWebControllerWithProvider)
	games.RegisterKVGame("pinochle", games.CategoryExtra5,
		func() usecase.PinochleInteractorIF {
			return usecase.NewPinochleInteractor(domain.NewDefaultPinochle(), new(presenter.PinochleWebPresenter))
		},
		func(data []byte) (usecase.PinochleInteractorIF, error) {
			return usecase.RestorePinochleInteractor(data, new(presenter.PinochleWebPresenter))
		},
		controller.NewPinochleWebControllerWithProvider)
	games.RegisterKVGame("shelem", games.CategoryExtra5,
		func() usecase.ShelemInteractorIF {
			return usecase.NewShelemInteractor(domain.NewDefaultShelem(), new(presenter.ShelemWebPresenter))
		},
		func(data []byte) (usecase.ShelemInteractorIF, error) {
			return usecase.RestoreShelemInteractor(data, new(presenter.ShelemWebPresenter))
		},
		controller.NewShelemWebControllerWithProvider)
	games.RegisterKVGame("doubt", games.CategoryExtra5,
		func() usecase.DoubtInteractorIF {
			return usecase.NewDoubtInteractor(domain.NewDefaultDoubt(), new(presenter.DoubtWebPresenter))
		},
		func(data []byte) (usecase.DoubtInteractorIF, error) {
			return usecase.RestoreDoubtInteractor(data, new(presenter.DoubtWebPresenter))
		},
		controller.NewDoubtWebControllerWithProvider)
	games.RegisterKVGame("cribbage", games.CategoryExtra5,
		func() usecase.CribbageInteractorIF {
			return usecase.NewCribbageInteractor(domain.NewDefaultCribbage(), new(presenter.CribbageWebPresenter))
		},
		func(data []byte) (usecase.CribbageInteractorIF, error) {
			return usecase.RestoreCribbageInteractor(data, new(presenter.CribbageWebPresenter))
		},
		controller.NewCribbageWebControllerWithProvider)
	games.RegisterKVGame("hasenpfeffer", games.CategoryExtra5,
		func() usecase.HasenpfefferInteractorIF {
			return usecase.NewHasenpfefferInteractor(domain.NewDefaultHasenpfeffer(), new(presenter.HasenpfefferWebPresenter))
		},
		func(data []byte) (usecase.HasenpfefferInteractorIF, error) {
			return usecase.RestoreHasenpfefferInteractor(data, new(presenter.HasenpfefferWebPresenter))
		},
		controller.NewHasenpfefferWebControllerWithProvider)
	games.RegisterKVGame("minibridge", games.CategoryExtra5,
		func() usecase.MinibridgeInteractorIF {
			return usecase.NewMinibridgeInteractor(domain.NewDefaultMinibridge(), new(presenter.MinibridgeWebPresenter))
		},
		func(data []byte) (usecase.MinibridgeInteractorIF, error) {
			return usecase.RestoreMinibridgeInteractor(data, new(presenter.MinibridgeWebPresenter))
		},
		controller.NewMinibridgeWebControllerWithProvider)
	games.RegisterKVGame("loo", games.CategoryExtra5,
		func() usecase.LooInteractorIF {
			return usecase.NewLooInteractor(domain.NewDefaultLoo(), new(presenter.LooWebPresenter))
		},
		func(data []byte) (usecase.LooInteractorIF, error) {
			return usecase.RestoreLooInteractor(data, new(presenter.LooWebPresenter))
		},
		controller.NewLooWebControllerWithProvider)
	games.RegisterKVGame("mighty", games.CategoryExtra5,
		func() usecase.MightyInteractorIF {
			return usecase.NewMightyInteractor(domain.NewDefaultMighty(), new(presenter.MightyWebPresenter))
		},
		func(data []byte) (usecase.MightyInteractorIF, error) {
			return usecase.RestoreMightyInteractor(data, new(presenter.MightyWebPresenter))
		},
		controller.NewMightyWebControllerWithProvider)
	games.RegisterKVGame("quadrille", games.CategoryExtra5,
		func() usecase.QuadrilleInteractorIF {
			return usecase.NewQuadrilleInteractor(domain.NewDefaultQuadrille(), new(presenter.QuadrilleWebPresenter))
		},
		func(data []byte) (usecase.QuadrilleInteractorIF, error) {
			return usecase.RestoreQuadrilleInteractor(data, new(presenter.QuadrilleWebPresenter))
		},
		controller.NewQuadrilleWebControllerWithProvider)
	games.RegisterKVGame("cego", games.CategoryExtra5,
		func() usecase.CegoInteractorIF {
			return usecase.NewCegoInteractor(domain.NewDefaultCego(), new(presenter.CegoWebPresenter))
		},
		func(data []byte) (usecase.CegoInteractorIF, error) {
			return usecase.RestoreCegoInteractor(data, new(presenter.CegoWebPresenter))
		},
		controller.NewCegoWebControllerWithProvider)
	games.RegisterKVGame("comet", games.CategoryExtra5,
		func() usecase.CometInteractorIF {
			return usecase.NewCometInteractor(domain.NewDefaultComet(), new(presenter.CometWebPresenter))
		},
		func(data []byte) (usecase.CometInteractorIF, error) {
			return usecase.RestoreCometInteractor(data, new(presenter.CometWebPresenter))
		},
		controller.NewCometWebControllerWithProvider)
	games.RegisterKVGame("kingo", games.CategoryExtra5,
		func() usecase.KingoInteractorIF {
			return usecase.NewKingoInteractor(domain.NewDefaultKingo(), new(presenter.KingoWebPresenter))
		},
		func(data []byte) (usecase.KingoInteractorIF, error) {
			return usecase.RestoreKingoInteractor(data, new(presenter.KingoWebPresenter))
		},
		controller.NewKingoWebControllerWithProvider)
	games.RegisterKVGame("oichokabu", games.CategoryExtra5,
		func() usecase.OichoKabuInteractorIF {
			return usecase.NewOichoKabuInteractor(domain.NewDefaultOichoKabu(), new(presenter.OichoKabuWebPresenter))
		},
		func(data []byte) (usecase.OichoKabuInteractorIF, error) {
			return usecase.RestoreOichoKabuInteractor(data, new(presenter.OichoKabuWebPresenter))
		},
		controller.NewOichoKabuWebControllerWithProvider)
	games.RegisterKVGame("speculation", games.CategoryExtra5,
		func() usecase.SpeculationInteractorIF {
			return usecase.NewSpeculationInteractor(domain.NewDefaultSpeculation(), new(presenter.SpeculationWebPresenter))
		},
		func(data []byte) (usecase.SpeculationInteractorIF, error) {
			return usecase.RestoreSpeculationInteractor(data, new(presenter.SpeculationWebPresenter))
		},
		controller.NewSpeculationWebControllerWithProvider)
	games.RegisterKVGame("continentalrummy", games.CategoryExtra5,
		func() usecase.ContinentalRummyInteractorIF {
			return usecase.NewContinentalRummyInteractor(domain.NewDefaultContinentalRummy(), new(presenter.ContinentalRummyWebPresenter))
		},
		func(data []byte) (usecase.ContinentalRummyInteractorIF, error) {
			return usecase.RestoreContinentalRummyInteractor(data, new(presenter.ContinentalRummyWebPresenter))
		},
		controller.NewContinentalRummyWebControllerWithProvider)
	games.RegisterKVGame("wizard", games.CategoryExtra5,
		func() usecase.WizardInteractorIF {
			return usecase.NewWizardInteractor(domain.NewDefaultWizard(), new(presenter.WizardWebPresenter))
		},
		func(data []byte) (usecase.WizardInteractorIF, error) {
			return usecase.RestoreWizardInteractor(data, new(presenter.WizardWebPresenter))
		},
		controller.NewWizardWebControllerWithProvider)
	games.RegisterKVGame("napoleon", games.CategoryExtra5,
		func() usecase.NapoleonInteractorIF {
			return usecase.NewNapoleonInteractor(domain.NewDefaultNapoleon(), new(presenter.NapoleonWebPresenter))
		},
		func(data []byte) (usecase.NapoleonInteractorIF, error) {
			return usecase.RestoreNapoleonInteractor(data, new(presenter.NapoleonWebPresenter))
		},
		controller.NewNapoleonWebControllerWithProvider)
	games.RegisterKVGame("ulti", games.CategoryExtra5,
		func() usecase.UltiInteractorIF {
			return usecase.NewUltiInteractor(domain.NewDefaultUlti(), new(presenter.UltiWebPresenter))
		},
		func(data []byte) (usecase.UltiInteractorIF, error) {
			return usecase.RestoreUltiInteractor(data, new(presenter.UltiWebPresenter))
		},
		controller.NewUltiWebControllerWithProvider)
}
