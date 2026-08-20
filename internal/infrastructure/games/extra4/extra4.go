//go:build js && wasm

// Package extra4 binds the Cloudflare Worker KV-backed handlers for the games
// assigned to the seventh size bucket. A worker main blank-imports this package
// for its registration side effects, so that whatever it registers is in place
// before games.RegisterCategory is called.
//
// Like casino/classic/solo/extra/extra2/extra3 this is purely a binary-size
// bucket, not a user-facing taxonomy (ADR-0037). The colourless name is
// deliberate: it holds whatever had to move to keep every TinyGo WASM binary
// under the Cloudflare Workers free-tier 1 MB gzipped limit, and nothing about
// a game's genre says it belongs here.
//
// Unlike extra2/extra3, this bucket was not introduced empty. ADR-0037 adds it
// and moves games in within one change, because extra3 had 188 bytes of gzip
// headroom left and an empty Phase 1 would have left that exposed for the whole
// gap between the two phases.
package extra4

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func init() {
	games.RegisterKVGame("anaconda", games.CategoryExtra4,
		func() usecase.AnacondaInteractorIF {
			return usecase.NewAnacondaInteractor(domain.NewDefaultAnaconda(), new(presenter.AnacondaWebPresenter))
		},
		func(data []byte) (usecase.AnacondaInteractorIF, error) {
			return usecase.RestoreAnacondaInteractor(data, new(presenter.AnacondaWebPresenter))
		},
		controller.NewAnacondaWebControllerWithProvider)
	games.RegisterKVGame("andarbahar", games.CategoryExtra4,
		func() usecase.AndarBaharInteractorIF {
			return usecase.NewAndarBaharInteractor(domain.NewDefaultAndarBahar(), new(presenter.AndarBaharWebPresenter))
		},
		func(data []byte) (usecase.AndarBaharInteractorIF, error) {
			return usecase.RestoreAndarBaharInteractor(data, new(presenter.AndarBaharWebPresenter))
		},
		controller.NewAndarBaharWebControllerWithProvider)
	games.RegisterKVGame("barbu", games.CategoryExtra4,
		func() usecase.BarbuInteractorIF {
			return usecase.NewBarbuInteractor(domain.NewDefaultBarbu(), new(presenter.BarbuWebPresenter))
		},
		func(data []byte) (usecase.BarbuInteractorIF, error) {
			return usecase.RestoreBarbuInteractor(data, new(presenter.BarbuWebPresenter))
		},
		controller.NewBarbuWebControllerWithProvider)
	games.RegisterKVGame("bezique", games.CategoryExtra4,
		func() usecase.BeziqueInteractorIF {
			return usecase.NewBeziqueInteractor(domain.NewDefaultBezique(), new(presenter.BeziqueWebPresenter))
		},
		func(data []byte) (usecase.BeziqueInteractorIF, error) {
			return usecase.RestoreBeziqueInteractor(data, new(presenter.BeziqueWebPresenter))
		},
		controller.NewBeziqueWebControllerWithProvider)
	games.RegisterKVGame("casinowar", games.CategoryExtra4,
		func() usecase.CasinoWarInteractorIF {
			return usecase.NewCasinoWarInteractor(domain.NewDefaultCasinoWar(), new(presenter.CasinoWarWebPresenter))
		},
		func(data []byte) (usecase.CasinoWarInteractorIF, error) {
			return usecase.RestoreCasinoWarInteractor(data, new(presenter.CasinoWarWebPresenter))
		},
		controller.NewCasinoWarWebControllerWithProvider)
	games.RegisterKVGame("cego", games.CategoryExtra4,
		func() usecase.CegoInteractorIF {
			return usecase.NewCegoInteractor(domain.NewDefaultCego(), new(presenter.CegoWebPresenter))
		},
		func(data []byte) (usecase.CegoInteractorIF, error) {
			return usecase.RestoreCegoInteractor(data, new(presenter.CegoWebPresenter))
		},
		controller.NewCegoWebControllerWithProvider)
	games.RegisterKVGame("chemindefer", games.CategoryExtra4,
		func() usecase.ChemindeFerInteractorIF {
			return usecase.NewChemindeFerInteractor(domain.NewDefaultChemindeFer(), new(presenter.ChemindeFerWebPresenter))
		},
		func(data []byte) (usecase.ChemindeFerInteractorIF, error) {
			return usecase.RestoreChemindeFerInteractor(data, new(presenter.ChemindeFerWebPresenter))
		},
		controller.NewChemindeFerWebControllerWithProvider)
	games.RegisterKVGame("crazyfourpoker", games.CategoryExtra4,
		func() usecase.CrazyFourPokerInteractorIF {
			return usecase.NewCrazyFourPokerInteractor(domain.NewDefaultCrazyFourPoker(), new(presenter.CrazyFourPokerWebPresenter))
		},
		func(data []byte) (usecase.CrazyFourPokerInteractorIF, error) {
			return usecase.RestoreCrazyFourPokerInteractor(data, new(presenter.CrazyFourPokerWebPresenter))
		},
		controller.NewCrazyFourPokerWebControllerWithProvider)
	games.RegisterKVGame("doudizhu", games.CategoryExtra4,
		func() usecase.DoudizhuInteractorIF {
			return usecase.NewDoudizhuInteractor(domain.NewDefaultDoudizhu(), new(presenter.DoudizhuWebPresenter))
		},
		func(data []byte) (usecase.DoudizhuInteractorIF, error) {
			return usecase.RestoreDoudizhuInteractor(data, new(presenter.DoudizhuWebPresenter))
		},
		controller.NewDoudizhuWebControllerWithProvider)
	games.RegisterKVGame("dragontiger", games.CategoryExtra4,
		func() usecase.DragonTigerInteractorIF {
			return usecase.NewDragonTigerInteractor(domain.NewDefaultDragonTiger(), new(presenter.DragonTigerWebPresenter))
		},
		func(data []byte) (usecase.DragonTigerInteractorIF, error) {
			return usecase.RestoreDragonTigerInteractor(data, new(presenter.DragonTigerWebPresenter))
		},
		controller.NewDragonTigerWebControllerWithProvider)
	games.RegisterKVGame("flowergarden", games.CategoryExtra4,
		func() usecase.FlowerGardenInteractorIF {
			return usecase.NewFlowerGardenInteractor(domain.NewDefaultFlowerGarden(), new(presenter.FlowerGardenWebPresenter))
		},
		func(data []byte) (usecase.FlowerGardenInteractorIF, error) {
			return usecase.RestoreFlowerGardenInteractor(data, new(presenter.FlowerGardenWebPresenter))
		},
		controller.NewFlowerGardenWebControllerWithProvider)
	games.RegisterKVGame("guts", games.CategoryExtra4,
		func() usecase.GutsInteractorIF {
			return usecase.NewGutsInteractor(domain.NewDefaultGuts(), new(presenter.GutsWebPresenter))
		},
		func(data []byte) (usecase.GutsInteractorIF, error) {
			return usecase.RestoreGutsInteractor(data, new(presenter.GutsWebPresenter))
		},
		controller.NewGutsWebControllerWithProvider)
	games.RegisterKVGame("highcardflush", games.CategoryExtra4,
		func() usecase.HighCardFlushInteractorIF {
			return usecase.NewHighCardFlushInteractor(domain.NewDefaultHighCardFlush(), new(presenter.HighCardFlushWebPresenter))
		},
		func(data []byte) (usecase.HighCardFlushInteractorIF, error) {
			return usecase.RestoreHighCardFlushInteractor(data, new(presenter.HighCardFlushWebPresenter))
		},
		controller.NewHighCardFlushWebControllerWithProvider)
	games.RegisterKVGame("honeymoonbridge", games.CategoryExtra4,
		func() usecase.HoneymoonBridgeInteractorIF {
			return usecase.NewHoneymoonBridgeInteractor(domain.NewDefaultHoneymoonBridge(), new(presenter.HoneymoonBridgeWebPresenter))
		},
		func(data []byte) (usecase.HoneymoonBridgeInteractorIF, error) {
			return usecase.RestoreHoneymoonBridgeInteractor(data, new(presenter.HoneymoonBridgeWebPresenter))
		},
		controller.NewHoneymoonBridgeWebControllerWithProvider)
	games.RegisterKVGame("israeliwhist", games.CategoryExtra4,
		func() usecase.IsraeliWhistInteractorIF {
			return usecase.NewIsraeliWhistInteractor(domain.NewDefaultIsraeliWhist(), new(presenter.IsraeliWhistWebPresenter))
		},
		func(data []byte) (usecase.IsraeliWhistInteractorIF, error) {
			return usecase.RestoreIsraeliWhistInteractor(data, new(presenter.IsraeliWhistWebPresenter))
		},
		controller.NewIsraeliWhistWebControllerWithProvider)
	games.RegisterKVGame("letitride", games.CategoryExtra4,
		func() usecase.LetItRideInteractorIF {
			return usecase.NewLetItRideInteractor(domain.NewDefaultLetItRide(), new(presenter.LetItRideWebPresenter))
		},
		func(data []byte) (usecase.LetItRideInteractorIF, error) {
			return usecase.RestoreLetItRideInteractor(data, new(presenter.LetItRideWebPresenter))
		},
		controller.NewLetItRideWebControllerWithProvider)
	games.RegisterKVGame("lingerlonger", games.CategoryExtra4,
		func() usecase.LingerLongerInteractorIF {
			return usecase.NewLingerLongerInteractor(domain.NewDefaultLingerLonger(), new(presenter.LingerLongerWebPresenter))
		},
		func(data []byte) (usecase.LingerLongerInteractorIF, error) {
			return usecase.RestoreLingerLongerInteractor(data, new(presenter.LingerLongerWebPresenter))
		},
		controller.NewLingerLongerWebControllerWithProvider)
	games.RegisterKVGame("literature", games.CategoryExtra4,
		func() usecase.LiteratureInteractorIF {
			return usecase.NewLiteratureInteractor(domain.NewDefaultLiterature(), new(presenter.LiteratureWebPresenter))
		},
		func(data []byte) (usecase.LiteratureInteractorIF, error) {
			return usecase.RestoreLiteratureInteractor(data, new(presenter.LiteratureWebPresenter))
		},
		controller.NewLiteratureWebControllerWithProvider)
	games.RegisterKVGame("mendikot", games.CategoryExtra4,
		func() usecase.MendikotInteractorIF {
			return usecase.NewMendikotInteractor(domain.NewDefaultMendikot(), new(presenter.MendikotWebPresenter))
		},
		func(data []byte) (usecase.MendikotInteractorIF, error) {
			return usecase.RestoreMendikotInteractor(data, new(presenter.MendikotWebPresenter))
		},
		controller.NewMendikotWebControllerWithProvider)
	games.RegisterKVGame("mighty", games.CategoryExtra4,
		func() usecase.MightyInteractorIF {
			return usecase.NewMightyInteractor(domain.NewDefaultMighty(), new(presenter.MightyWebPresenter))
		},
		func(data []byte) (usecase.MightyInteractorIF, error) {
			return usecase.RestoreMightyInteractor(data, new(presenter.MightyWebPresenter))
		},
		controller.NewMightyWebControllerWithProvider)
	games.RegisterKVGame("napoleon", games.CategoryExtra4,
		func() usecase.NapoleonInteractorIF {
			return usecase.NewNapoleonInteractor(domain.NewDefaultNapoleon(), new(presenter.NapoleonWebPresenter))
		},
		func(data []byte) (usecase.NapoleonInteractorIF, error) {
			return usecase.RestoreNapoleonInteractor(data, new(presenter.NapoleonWebPresenter))
		},
		controller.NewNapoleonWebControllerWithProvider)
	games.RegisterKVGame("ombre", games.CategoryExtra4,
		func() usecase.OmbreInteractorIF {
			return usecase.NewOmbreInteractor(domain.NewDefaultOmbre(), new(presenter.OmbreWebPresenter))
		},
		func(data []byte) (usecase.OmbreInteractorIF, error) {
			return usecase.RestoreOmbreInteractor(data, new(presenter.OmbreWebPresenter))
		},
		controller.NewOmbreWebControllerWithProvider)
	games.RegisterKVGame("piquet", games.CategoryExtra4,
		func() usecase.PiquetInteractorIF {
			return usecase.NewPiquetInteractor(domain.NewDefaultPiquet(), new(presenter.PiquetWebPresenter))
		},
		func(data []byte) (usecase.PiquetInteractorIF, error) {
			return usecase.RestorePiquetInteractor(data, new(presenter.PiquetWebPresenter))
		},
		controller.NewPiquetWebControllerWithProvider)
	games.RegisterKVGame("reddog", games.CategoryExtra4,
		func() usecase.RedDogInteractorIF {
			return usecase.NewRedDogInteractor(domain.NewDefaultRedDog(), new(presenter.RedDogWebPresenter))
		},
		func(data []byte) (usecase.RedDogInteractorIF, error) {
			return usecase.RestoreRedDogInteractor(data, new(presenter.RedDogWebPresenter))
		},
		controller.NewRedDogWebControllerWithProvider)
	games.RegisterKVGame("russianbank", games.CategoryExtra4,
		func() usecase.RussianBankInteractorIF {
			return usecase.NewRussianBankInteractor(domain.NewDefaultRussianBank(), new(presenter.RussianBankWebPresenter))
		},
		func(data []byte) (usecase.RussianBankInteractorIF, error) {
			return usecase.RestoreRussianBankInteractor(data, new(presenter.RussianBankWebPresenter))
		},
		controller.NewRussianBankWebControllerWithProvider)
	games.RegisterKVGame("scarto", games.CategoryExtra4,
		func() usecase.ScartoInteractorIF {
			return usecase.NewScartoInteractor(domain.NewDefaultScarto(), new(presenter.ScartoWebPresenter))
		},
		func(data []byte) (usecase.ScartoInteractorIF, error) {
			return usecase.RestoreScartoInteractor(data, new(presenter.ScartoWebPresenter))
		},
		controller.NewScartoWebControllerWithProvider)
	games.RegisterKVGame("sergeantmajor", games.CategoryExtra4,
		func() usecase.SergeantMajorInteractorIF {
			return usecase.NewSergeantMajorInteractor(domain.NewDefaultSergeantMajor(), new(presenter.SergeantMajorWebPresenter))
		},
		func(data []byte) (usecase.SergeantMajorInteractorIF, error) {
			return usecase.RestoreSergeantMajorInteractor(data, new(presenter.SergeantMajorWebPresenter))
		},
		controller.NewSergeantMajorWebControllerWithProvider)
	games.RegisterKVGame("sheepshead", games.CategoryExtra4,
		func() usecase.SheepsheadInteractorIF {
			return usecase.NewSheepsheadInteractor(domain.NewDefaultSheepshead(), new(presenter.SheepsheadWebPresenter))
		},
		func(data []byte) (usecase.SheepsheadInteractorIF, error) {
			return usecase.RestoreSheepsheadInteractor(data, new(presenter.SheepsheadWebPresenter))
		},
		controller.NewSheepsheadWebControllerWithProvider)
	games.RegisterKVGame("shengji", games.CategoryExtra4,
		func() usecase.ShengJiInteractorIF {
			return usecase.NewShengJiInteractor(domain.NewDefaultShengJi(), new(presenter.ShengJiWebPresenter))
		},
		func(data []byte) (usecase.ShengJiInteractorIF, error) {
			return usecase.RestoreShengJiInteractor(data, new(presenter.ShengJiWebPresenter))
		},
		controller.NewShengJiWebControllerWithProvider)
	games.RegisterKVGame("sixbidsolo", games.CategoryExtra4,
		func() usecase.SixBidSoloInteractorIF {
			return usecase.NewSixBidSoloInteractor(domain.NewDefaultSixBidSolo(), new(presenter.SixBidSoloWebPresenter))
		},
		func(data []byte) (usecase.SixBidSoloInteractorIF, error) {
			return usecase.RestoreSixBidSoloInteractor(data, new(presenter.SixBidSoloWebPresenter))
		},
		controller.NewSixBidSoloWebControllerWithProvider)
	games.RegisterKVGame("trenteetquarante", games.CategoryExtra4,
		func() usecase.TrenteEtQuaranteInteractorIF {
			return usecase.NewTrenteEtQuaranteInteractor(domain.NewDefaultTrenteEtQuarante(), new(presenter.TrenteEtQuaranteWebPresenter))
		},
		func(data []byte) (usecase.TrenteEtQuaranteInteractorIF, error) {
			return usecase.RestoreTrenteEtQuaranteInteractor(data, new(presenter.TrenteEtQuaranteWebPresenter))
		},
		controller.NewTrenteEtQuaranteWebControllerWithProvider)
	games.RegisterKVGame("ulti", games.CategoryExtra4,
		func() usecase.UltiInteractorIF {
			return usecase.NewUltiInteractor(domain.NewDefaultUlti(), new(presenter.UltiWebPresenter))
		},
		func(data []byte) (usecase.UltiInteractorIF, error) {
			return usecase.RestoreUltiInteractor(data, new(presenter.UltiWebPresenter))
		},
		controller.NewUltiWebControllerWithProvider)
	games.RegisterKVGame("colourwhist", games.CategoryExtra4,
		func() usecase.ColourWhistInteractorIF {
			return usecase.NewColourWhistInteractor(domain.NewDefaultColourWhist(), new(presenter.ColourWhistWebPresenter))
		},
		func(data []byte) (usecase.ColourWhistInteractorIF, error) {
			return usecase.RestoreColourWhistInteractor(data, new(presenter.ColourWhistWebPresenter))
		},
		controller.NewColourWhistWebControllerWithProvider)
	games.RegisterKVGame("estimation", games.CategoryExtra4,
		func() usecase.EstimationInteractorIF {
			return usecase.NewEstimationInteractor(domain.NewDefaultEstimation(), new(presenter.EstimationWebPresenter))
		},
		func(data []byte) (usecase.EstimationInteractorIF, error) {
			return usecase.RestoreEstimationInteractor(data, new(presenter.EstimationWebPresenter))
		},
		controller.NewEstimationWebControllerWithProvider)
	games.RegisterKVGame("preference", games.CategoryExtra4,
		func() usecase.PreferenceInteractorIF {
			return usecase.NewPreferenceInteractor(domain.NewDefaultPreference(), new(presenter.PreferenceWebPresenter))
		},
		func(data []byte) (usecase.PreferenceInteractorIF, error) {
			return usecase.RestorePreferenceInteractor(data, new(presenter.PreferenceWebPresenter))
		},
		controller.NewPreferenceWebControllerWithProvider)
	games.RegisterKVGame("bhabhi", games.CategoryExtra4,
		func() usecase.BhabhiInteractorIF {
			return usecase.NewBhabhiInteractor(domain.NewDefaultBhabhi(), new(presenter.BhabhiWebPresenter))
		},
		func(data []byte) (usecase.BhabhiInteractorIF, error) {
			return usecase.RestoreBhabhiInteractor(data, new(presenter.BhabhiWebPresenter))
		},
		controller.NewBhabhiWebControllerWithProvider)
	games.RegisterKVGame("pasur", games.CategoryExtra4,
		func() usecase.PasurInteractorIF {
			return usecase.NewPasurInteractor(domain.NewDefaultPasur(), new(presenter.PasurWebPresenter))
		},
		func(data []byte) (usecase.PasurInteractorIF, error) {
			return usecase.RestorePasurInteractor(data, new(presenter.PasurWebPresenter))
		},
		controller.NewPasurWebControllerWithProvider)
	games.RegisterKVGame("guandan", games.CategoryExtra4,
		func() usecase.GuandanInteractorIF {
			return usecase.NewGuandanInteractor(domain.NewDefaultGuandan(), new(presenter.GuandanWebPresenter))
		},
		func(data []byte) (usecase.GuandanInteractorIF, error) {
			return usecase.RestoreGuandanInteractor(data, new(presenter.GuandanWebPresenter))
		},
		controller.NewGuandanWebControllerWithProvider)
	games.RegisterKVGame("michigan", games.CategoryExtra4,
		func() usecase.MichiganInteractorIF {
			return usecase.NewMichiganInteractor(domain.NewDefaultMichigan(), new(presenter.MichiganWebPresenter))
		},
		func(data []byte) (usecase.MichiganInteractorIF, error) {
			return usecase.RestoreMichiganInteractor(data, new(presenter.MichiganWebPresenter))
		},
		controller.NewMichiganWebControllerWithProvider)
}
