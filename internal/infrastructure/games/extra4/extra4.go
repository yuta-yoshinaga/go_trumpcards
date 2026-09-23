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
	games.RegisterKVGame("chemindefer", games.CategoryExtra4,
		func() usecase.ChemindeFerInteractorIF {
			return usecase.NewChemindeFerInteractor(domain.NewDefaultChemindeFer(), new(presenter.ChemindeFerWebPresenter))
		},
		func(data []byte) (usecase.ChemindeFerInteractorIF, error) {
			return usecase.RestoreChemindeFerInteractor(data, new(presenter.ChemindeFerWebPresenter))
		},
		controller.NewChemindeFerWebControllerWithProvider)
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
	games.RegisterKVGame("russianbank", games.CategoryExtra4,
		func() usecase.RussianBankInteractorIF {
			return usecase.NewRussianBankInteractor(domain.NewDefaultRussianBank(), new(presenter.RussianBankWebPresenter))
		},
		func(data []byte) (usecase.RussianBankInteractorIF, error) {
			return usecase.RestoreRussianBankInteractor(data, new(presenter.RussianBankWebPresenter))
		},
		controller.NewRussianBankWebControllerWithProvider)
	games.RegisterKVGame("piedmontesetarot", games.CategoryExtra4,
		func() usecase.PiedmonteseTarotInteractorIF {
			return usecase.NewPiedmonteseTarotInteractor(domain.NewDefaultPiedmonteseTarot(), new(presenter.PiedmonteseTarotWebPresenter))
		},
		func(data []byte) (usecase.PiedmonteseTarotInteractorIF, error) {
			return usecase.RestorePiedmonteseTarotInteractor(data, new(presenter.PiedmonteseTarotWebPresenter))
		},
		controller.NewPiedmonteseTarotWebControllerWithProvider)
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
	games.RegisterKVGame("michigan", games.CategoryExtra4,
		func() usecase.MichiganInteractorIF {
			return usecase.NewMichiganInteractor(domain.NewDefaultMichigan(), new(presenter.MichiganWebPresenter))
		},
		func(data []byte) (usecase.MichiganInteractorIF, error) {
			return usecase.RestoreMichiganInteractor(data, new(presenter.MichiganWebPresenter))
		},
		controller.NewMichiganWebControllerWithProvider)
	games.RegisterKVGame("alaska", games.CategoryExtra4,
		func() usecase.AlaskaInteractorIF {
			return usecase.NewAlaskaInteractor(domain.NewDefaultAlaska(), new(presenter.AlaskaWebPresenter))
		},
		func(data []byte) (usecase.AlaskaInteractorIF, error) {
			return usecase.RestoreAlaskaInteractor(data, new(presenter.AlaskaWebPresenter))
		},
		controller.NewAlaskaWebControllerWithProvider)
	games.RegisterKVGame("perseverance", games.CategoryExtra4,
		func() usecase.PerseveranceInteractorIF {
			return usecase.NewPerseveranceInteractor(domain.NewDefaultPerseverance(), new(presenter.PerseveranceWebPresenter))
		},
		func(data []byte) (usecase.PerseveranceInteractorIF, error) {
			return usecase.RestorePerseveranceInteractor(data, new(presenter.PerseveranceWebPresenter))
		},
		controller.NewPerseveranceWebControllerWithProvider)
	games.RegisterKVGame("fourteenout", games.CategoryExtra4,
		func() usecase.FourteenOutInteractorIF {
			return usecase.NewFourteenOutInteractor(domain.NewDefaultFourteenOut(), new(presenter.FourteenOutWebPresenter))
		},
		func(data []byte) (usecase.FourteenOutInteractorIF, error) {
			return usecase.RestoreFourteenOutInteractor(data, new(presenter.FourteenOutWebPresenter))
		},
		controller.NewFourteenOutWebControllerWithProvider)
	games.RegisterKVGame("narcotic", games.CategoryExtra4,
		func() usecase.NarcoticInteractorIF {
			return usecase.NewNarcoticInteractor(domain.NewDefaultNarcotic(), new(presenter.NarcoticWebPresenter))
		},
		func(data []byte) (usecase.NarcoticInteractorIF, error) {
			return usecase.RestoreNarcoticInteractor(data, new(presenter.NarcoticWebPresenter))
		},
		controller.NewNarcoticWebControllerWithProvider)
	games.RegisterKVGame("mrsmop", games.CategoryExtra4,
		func() usecase.MrsMopInteractorIF {
			return usecase.NewMrsMopInteractor(domain.NewDefaultMrsMop(), new(presenter.MrsMopWebPresenter))
		},
		func(data []byte) (usecase.MrsMopInteractorIF, error) {
			return usecase.RestoreMrsMopInteractor(data, new(presenter.MrsMopWebPresenter))
		},
		controller.NewMrsMopWebControllerWithProvider)
	games.RegisterKVGame("rankandfile", games.CategoryExtra4,
		func() usecase.RankAndFileInteractorIF {
			return usecase.NewRankAndFileInteractor(domain.NewDefaultRankAndFile(), new(presenter.RankAndFileWebPresenter))
		},
		func(data []byte) (usecase.RankAndFileInteractorIF, error) {
			return usecase.RestoreRankAndFileInteractor(data, new(presenter.RankAndFileWebPresenter))
		},
		controller.NewRankAndFileWebControllerWithProvider)
	games.RegisterKVGame("put", games.CategoryExtra4,
		func() usecase.PutInteractorIF {
			return usecase.NewPutInteractor(domain.NewDefaultPut(), new(presenter.PutWebPresenter))
		},
		func(data []byte) (usecase.PutInteractorIF, error) {
			return usecase.RestorePutInteractor(data, new(presenter.PutWebPresenter))
		},
		controller.NewPutWebControllerWithProvider)
	games.RegisterKVGame("basset", games.CategoryExtra4,
		func() usecase.BassetInteractorIF {
			return usecase.NewBassetInteractor(domain.NewDefaultBasset(), new(presenter.BassetWebPresenter))
		},
		func(data []byte) (usecase.BassetInteractorIF, error) {
			return usecase.RestoreBassetInteractor(data, new(presenter.BassetWebPresenter))
		},
		controller.NewBassetWebControllerWithProvider)
	games.RegisterKVGame("gongzhu", games.CategoryExtra4,
		func() usecase.GongZhuInteractorIF {
			return usecase.NewGongZhuInteractor(domain.NewDefaultGongZhu(), new(presenter.GongZhuWebPresenter))
		},
		func(data []byte) (usecase.GongZhuInteractorIF, error) {
			return usecase.RestoreGongZhuInteractor(data, new(presenter.GongZhuWebPresenter))
		},
		controller.NewGongZhuWebControllerWithProvider)
}
