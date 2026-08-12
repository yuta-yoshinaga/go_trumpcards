//go:build js && wasm

// Package extra binds the Cloudflare Worker KV-backed handlers for the games
// assigned to the fourth ("extra") size bucket. A worker main blank-imports
// this package so the init below runs before games.RegisterCategory is called.
//
// Like casino/classic/solo this is purely a binary-size bucket, not a
// user-facing taxonomy: it holds an overflow mix of games moved off the other
// three workers to keep every TinyGo WASM binary under the Cloudflare Workers
// free-tier 1 MB gzipped limit.
package extra

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func init() {
	games.RegisterKVGame("canasta", games.CategoryExtra,
		func() usecase.CanastaInteractorIF {
			return usecase.NewCanastaInteractor(domain.NewDefaultCanasta(), new(presenter.CanastaWebPresenter))
		},
		func(data []byte) (usecase.CanastaInteractorIF, error) {
			return usecase.RestoreCanastaInteractor(data, new(presenter.CanastaWebPresenter))
		},
		controller.NewCanastaWebControllerWithProvider)
	games.RegisterKVGame("ginrummy", games.CategoryExtra,
		func() usecase.GinRummyInteractorIF {
			return usecase.NewGinRummyInteractor(domain.NewDefaultGinRummy(), new(presenter.GinRummyWebPresenter))
		},
		func(data []byte) (usecase.GinRummyInteractorIF, error) {
			return usecase.RestoreGinRummyInteractor(data, new(presenter.GinRummyWebPresenter))
		},
		controller.NewGinRummyWebControllerWithProvider)
	games.RegisterKVGame("indianrummy", games.CategoryExtra,
		func() usecase.IndianRummyInteractorIF {
			return usecase.NewIndianRummyInteractor(domain.NewDefaultIndianRummy(), new(presenter.IndianRummyWebPresenter))
		},
		func(data []byte) (usecase.IndianRummyInteractorIF, error) {
			return usecase.RestoreIndianRummyInteractor(data, new(presenter.IndianRummyWebPresenter))
		},
		controller.NewIndianRummyWebControllerWithProvider)
	games.RegisterKVGame("contractrummy", games.CategoryExtra,
		func() usecase.ContractRummyInteractorIF {
			return usecase.NewContractRummyInteractor(domain.NewDefaultContractRummy(), new(presenter.ContractRummyWebPresenter))
		},
		func(data []byte) (usecase.ContractRummyInteractorIF, error) {
			return usecase.RestoreContractRummyInteractor(data, new(presenter.ContractRummyWebPresenter))
		},
		controller.NewContractRummyWebControllerWithProvider)
	games.RegisterKVGame("kalooki", games.CategoryExtra,
		func() usecase.KalookiInteractorIF {
			return usecase.NewKalookiInteractor(domain.NewDefaultKalooki(), new(presenter.KalookiWebPresenter))
		},
		func(data []byte) (usecase.KalookiInteractorIF, error) {
			return usecase.RestoreKalookiInteractor(data, new(presenter.KalookiWebPresenter))
		},
		controller.NewKalookiWebControllerWithProvider)
	games.RegisterKVGame("burraco", games.CategoryExtra,
		func() usecase.BurracoInteractorIF {
			return usecase.NewBurracoInteractor(domain.NewDefaultBurraco(), new(presenter.BurracoWebPresenter))
		},
		func(data []byte) (usecase.BurracoInteractorIF, error) {
			return usecase.RestoreBurracoInteractor(data, new(presenter.BurracoWebPresenter))
		},
		controller.NewBurracoWebControllerWithProvider)
	games.RegisterKVGame("handandfoot", games.CategoryExtra,
		func() usecase.HandAndFootInteractorIF {
			return usecase.NewHandAndFootInteractor(domain.NewDefaultHandAndFoot(), new(presenter.HandAndFootWebPresenter))
		},
		func(data []byte) (usecase.HandAndFootInteractorIF, error) {
			return usecase.RestoreHandAndFootInteractor(data, new(presenter.HandAndFootWebPresenter))
		},
		controller.NewHandAndFootWebControllerWithProvider)
	games.RegisterKVGame("conquian", games.CategoryExtra,
		func() usecase.ConquianInteractorIF {
			return usecase.NewConquianInteractor(domain.NewDefaultConquian(), new(presenter.ConquianWebPresenter))
		},
		func(data []byte) (usecase.ConquianInteractorIF, error) {
			return usecase.RestoreConquianInteractor(data, new(presenter.ConquianWebPresenter))
		},
		controller.NewConquianWebControllerWithProvider)
	games.RegisterKVGame("chinchon", games.CategoryExtra,
		func() usecase.ChinchonInteractorIF {
			return usecase.NewChinchonInteractor(domain.NewDefaultChinchon(), new(presenter.ChinchonWebPresenter))
		},
		func(data []byte) (usecase.ChinchonInteractorIF, error) {
			return usecase.RestoreChinchonInteractor(data, new(presenter.ChinchonWebPresenter))
		},
		controller.NewChinchonWebControllerWithProvider)
	games.RegisterKVGame("threethirteen", games.CategoryExtra,
		func() usecase.ThreeThirteenInteractorIF {
			return usecase.NewThreeThirteenInteractor(domain.NewDefaultThreeThirteen(), new(presenter.ThreeThirteenWebPresenter))
		},
		func(data []byte) (usecase.ThreeThirteenInteractorIF, error) {
			return usecase.RestoreThreeThirteenInteractor(data, new(presenter.ThreeThirteenWebPresenter))
		},
		controller.NewThreeThirteenWebControllerWithProvider)
	games.RegisterKVGame("rummy500", games.CategoryExtra,
		func() usecase.Rummy500InteractorIF {
			return usecase.NewRummy500Interactor(domain.NewDefaultRummy500(), new(presenter.Rummy500WebPresenter))
		},
		func(data []byte) (usecase.Rummy500InteractorIF, error) {
			return usecase.RestoreRummy500Interactor(data, new(presenter.Rummy500WebPresenter))
		},
		controller.NewRummy500WebControllerWithProvider)
	games.RegisterKVGame("streetsandalleys", games.CategoryExtra,
		func() usecase.StreetsAndAlleysInteractorIF {
			return usecase.NewStreetsAndAlleysInteractor(domain.NewDefaultStreetsAndAlleys(), new(presenter.StreetsAndAlleysWebPresenter))
		},
		func(data []byte) (usecase.StreetsAndAlleysInteractorIF, error) {
			return usecase.RestoreStreetsAndAlleysInteractor(data, new(presenter.StreetsAndAlleysWebPresenter))
		},
		controller.NewStreetsAndAlleysWebControllerWithProvider)
	games.RegisterKVGame("kingalbert", games.CategoryExtra,
		func() usecase.KingAlbertInteractorIF {
			return usecase.NewKingAlbertInteractor(domain.NewDefaultKingAlbert(), new(presenter.KingAlbertWebPresenter))
		},
		func(data []byte) (usecase.KingAlbertInteractorIF, error) {
			return usecase.RestoreKingAlbertInteractor(data, new(presenter.KingAlbertWebPresenter))
		},
		controller.NewKingAlbertWebControllerWithProvider)
	games.RegisterKVGame("flowergarden", games.CategoryExtra,
		func() usecase.FlowerGardenInteractorIF {
			return usecase.NewFlowerGardenInteractor(domain.NewDefaultFlowerGarden(), new(presenter.FlowerGardenWebPresenter))
		},
		func(data []byte) (usecase.FlowerGardenInteractorIF, error) {
			return usecase.RestoreFlowerGardenInteractor(data, new(presenter.FlowerGardenWebPresenter))
		},
		controller.NewFlowerGardenWebControllerWithProvider)
	games.RegisterKVGame("agnes", games.CategoryExtra,
		func() usecase.AgnesInteractorIF {
			return usecase.NewAgnesInteractor(domain.NewDefaultAgnes(), new(presenter.AgnesWebPresenter))
		},
		func(data []byte) (usecase.AgnesInteractorIF, error) {
			return usecase.RestoreAgnesInteractor(data, new(presenter.AgnesWebPresenter))
		},
		controller.NewAgnesWebControllerWithProvider)
	games.RegisterKVGame("sultan", games.CategoryExtra,
		func() usecase.SultanInteractorIF {
			return usecase.NewSultanInteractor(domain.NewDefaultSultan(), new(presenter.SultanWebPresenter))
		},
		func(data []byte) (usecase.SultanInteractorIF, error) {
			return usecase.RestoreSultanInteractor(data, new(presenter.SultanWebPresenter))
		},
		controller.NewSultanWebControllerWithProvider)
	games.RegisterKVGame("gaigel", games.CategoryExtra,
		func() usecase.GaigelInteractorIF {
			return usecase.NewGaigelInteractor(domain.NewDefaultGaigel(), new(presenter.GaigelWebPresenter))
		},
		func(data []byte) (usecase.GaigelInteractorIF, error) {
			return usecase.RestoreGaigelInteractor(data, new(presenter.GaigelWebPresenter))
		},
		controller.NewGaigelWebControllerWithProvider)
	games.RegisterKVGame("tysiac", games.CategoryExtra,
		func() usecase.TysiacInteractorIF {
			return usecase.NewTysiacInteractor(domain.NewDefaultTysiac(), new(presenter.TysiacWebPresenter))
		},
		func(data []byte) (usecase.TysiacInteractorIF, error) {
			return usecase.RestoreTysiacInteractor(data, new(presenter.TysiacWebPresenter))
		},
		controller.NewTysiacWebControllerWithProvider)
	games.RegisterKVGame("calabresella", games.CategoryExtra,
		func() usecase.CalabresellaInteractorIF {
			return usecase.NewCalabresellaInteractor(domain.NewDefaultCalabresella(), new(presenter.CalabresellaWebPresenter))
		},
		func(data []byte) (usecase.CalabresellaInteractorIF, error) {
			return usecase.RestoreCalabresellaInteractor(data, new(presenter.CalabresellaWebPresenter))
		},
		controller.NewCalabresellaWebControllerWithProvider)
	games.RegisterKVGame("king", games.CategoryExtra,
		func() usecase.KingInteractorIF {
			return usecase.NewKingInteractor(domain.NewDefaultKing(), new(presenter.KingWebPresenter))
		},
		func(data []byte) (usecase.KingInteractorIF, error) {
			return usecase.RestoreKingInteractor(data, new(presenter.KingWebPresenter))
		},
		controller.NewKingWebControllerWithProvider)
	games.RegisterKVGame("cinch", games.CategoryExtra,
		func() usecase.CinchInteractorIF {
			return usecase.NewCinchInteractor(domain.NewDefaultCinch(), new(presenter.CinchWebPresenter))
		},
		func(data []byte) (usecase.CinchInteractorIF, error) {
			return usecase.RestoreCinchInteractor(data, new(presenter.CinchWebPresenter))
		},
		controller.NewCinchWebControllerWithProvider)
	games.RegisterKVGame("trenteetquarante", games.CategoryExtra,
		func() usecase.TrenteEtQuaranteInteractorIF {
			return usecase.NewTrenteEtQuaranteInteractor(domain.NewDefaultTrenteEtQuarante(), new(presenter.TrenteEtQuaranteWebPresenter))
		},
		func(data []byte) (usecase.TrenteEtQuaranteInteractorIF, error) {
			return usecase.RestoreTrenteEtQuaranteInteractor(data, new(presenter.TrenteEtQuaranteWebPresenter))
		},
		controller.NewTrenteEtQuaranteWebControllerWithProvider)
	games.RegisterKVGame("guts", games.CategoryExtra,
		func() usecase.GutsInteractorIF {
			return usecase.NewGutsInteractor(domain.NewDefaultGuts(), new(presenter.GutsWebPresenter))
		},
		func(data []byte) (usecase.GutsInteractorIF, error) {
			return usecase.RestoreGutsInteractor(data, new(presenter.GutsWebPresenter))
		},
		controller.NewGutsWebControllerWithProvider)
	games.RegisterKVGame("watten", games.CategoryExtra,
		func() usecase.WattenInteractorIF {
			return usecase.NewWattenInteractor(domain.NewDefaultWatten(), new(presenter.WattenWebPresenter))
		},
		func(data []byte) (usecase.WattenInteractorIF, error) {
			return usecase.RestoreWattenInteractor(data, new(presenter.WattenWebPresenter))
		},
		controller.NewWattenWebControllerWithProvider)
	games.RegisterKVGame("carioca", games.CategoryExtra,
		func() usecase.CariocaInteractorIF {
			return usecase.NewCariocaInteractor(domain.NewDefaultCarioca(), new(presenter.CariocaWebPresenter))
		},
		func(data []byte) (usecase.CariocaInteractorIF, error) {
			return usecase.RestoreCariocaInteractor(data, new(presenter.CariocaWebPresenter))
		},
		controller.NewCariocaWebControllerWithProvider)
	games.RegisterKVGame("samba", games.CategoryExtra,
		func() usecase.SambaInteractorIF {
			return usecase.NewSambaInteractor(domain.NewDefaultSamba(), new(presenter.SambaWebPresenter))
		},
		func(data []byte) (usecase.SambaInteractorIF, error) {
			return usecase.RestoreSambaInteractor(data, new(presenter.SambaWebPresenter))
		},
		controller.NewSambaWebControllerWithProvider)
	games.RegisterKVGame("anaconda", games.CategoryExtra,
		func() usecase.AnacondaInteractorIF {
			return usecase.NewAnacondaInteractor(domain.NewDefaultAnaconda(), new(presenter.AnacondaWebPresenter))
		},
		func(data []byte) (usecase.AnacondaInteractorIF, error) {
			return usecase.RestoreAnacondaInteractor(data, new(presenter.AnacondaWebPresenter))
		},
		controller.NewAnacondaWebControllerWithProvider)
	games.RegisterKVGame("machiavelli", games.CategoryExtra,
		func() usecase.MachiavelliInteractorIF {
			return usecase.NewMachiavelliInteractor(domain.NewDefaultMachiavelli(), new(presenter.MachiavelliWebPresenter))
		},
		func(data []byte) (usecase.MachiavelliInteractorIF, error) {
			return usecase.RestoreMachiavelliInteractor(data, new(presenter.MachiavelliWebPresenter))
		},
		controller.NewMachiavelliWebControllerWithProvider)
	games.RegisterKVGame("pan", games.CategoryExtra,
		func() usecase.PanInteractorIF {
			return usecase.NewPanInteractor(domain.NewDefaultPan(), new(presenter.PanWebPresenter))
		},
		func(data []byte) (usecase.PanInteractorIF, error) {
			return usecase.RestorePanInteractor(data, new(presenter.PanWebPresenter))
		},
		controller.NewPanWebControllerWithProvider)
	games.RegisterKVGame("oichokabu", games.CategoryExtra,
		func() usecase.OichoKabuInteractorIF {
			return usecase.NewOichoKabuInteractor(domain.NewDefaultOichoKabu(), new(presenter.OichoKabuWebPresenter))
		},
		func(data []byte) (usecase.OichoKabuInteractorIF, error) {
			return usecase.RestoreOichoKabuInteractor(data, new(presenter.OichoKabuWebPresenter))
		},
		controller.NewOichoKabuWebControllerWithProvider)
	games.RegisterKVGame("gostop", games.CategoryExtra,
		func() usecase.GoStopInteractorIF {
			return usecase.NewGoStopInteractor(domain.NewDefaultGoStop(), new(presenter.GoStopWebPresenter))
		},
		func(data []byte) (usecase.GoStopInteractorIF, error) {
			return usecase.RestoreGoStopInteractor(data, new(presenter.GoStopWebPresenter))
		},
		controller.NewGoStopWebControllerWithProvider)
	games.RegisterKVGame("hachihachi", games.CategoryExtra,
		func() usecase.HachiHachiInteractorIF {
			return usecase.NewHachiHachiInteractor(domain.NewDefaultHachiHachi(), new(presenter.HachiHachiWebPresenter))
		},
		func(data []byte) (usecase.HachiHachiInteractorIF, error) {
			return usecase.RestoreHachiHachiInteractor(data, new(presenter.HachiHachiWebPresenter))
		},
		controller.NewHachiHachiWebControllerWithProvider)
	games.RegisterKVGame("frenchtarot", games.CategoryExtra,
		func() usecase.FrenchTarotInteractorIF {
			return usecase.NewFrenchTarotInteractor(domain.NewDefaultFrenchTarot(), new(presenter.FrenchTarotWebPresenter))
		},
		func(data []byte) (usecase.FrenchTarotInteractorIF, error) {
			return usecase.RestoreFrenchTarotInteractor(data, new(presenter.FrenchTarotWebPresenter))
		},
		controller.NewFrenchTarotWebControllerWithProvider)
	games.RegisterKVGame("koenigrufen", games.CategoryExtra,
		func() usecase.KoenigrufenInteractorIF {
			return usecase.NewKoenigrufenInteractor(domain.NewDefaultKoenigrufen(), new(presenter.KoenigrufenWebPresenter))
		},
		func(data []byte) (usecase.KoenigrufenInteractorIF, error) {
			return usecase.RestoreKoenigrufenInteractor(data, new(presenter.KoenigrufenWebPresenter))
		},
		controller.NewKoenigrufenWebControllerWithProvider)
	games.RegisterKVGame("ganjifa", games.CategoryExtra,
		func() usecase.GanjifaInteractorIF {
			return usecase.NewGanjifaInteractor(domain.NewDefaultGanjifa(), new(presenter.GanjifaWebPresenter))
		},
		func(data []byte) (usecase.GanjifaInteractorIF, error) {
			return usecase.RestoreGanjifaInteractor(data, new(presenter.GanjifaWebPresenter))
		},
		controller.NewGanjifaWebControllerWithProvider)
	games.RegisterKVGame("vira", games.CategoryExtra,
		func() usecase.ViraInteractorIF {
			return usecase.NewViraInteractor(domain.NewDefaultVira(), new(presenter.ViraWebPresenter))
		},
		func(data []byte) (usecase.ViraInteractorIF, error) {
			return usecase.RestoreViraInteractor(data, new(presenter.ViraWebPresenter))
		},
		controller.NewViraWebControllerWithProvider)
	games.RegisterKVGame("diplomat", games.CategoryExtra,
		func() usecase.DiplomatInteractorIF {
			return usecase.NewDiplomatInteractor(domain.NewDefaultDiplomat(), new(presenter.DiplomatWebPresenter))
		},
		func(data []byte) (usecase.DiplomatInteractorIF, error) {
			return usecase.RestoreDiplomatInteractor(data, new(presenter.DiplomatWebPresenter))
		},
		controller.NewDiplomatWebControllerWithProvider)
	games.RegisterKVGame("mendikot", games.CategoryExtra,
		func() usecase.MendikotInteractorIF {
			return usecase.NewMendikotInteractor(domain.NewDefaultMendikot(), new(presenter.MendikotWebPresenter))
		},
		func(data []byte) (usecase.MendikotInteractorIF, error) {
			return usecase.RestoreMendikotInteractor(data, new(presenter.MendikotWebPresenter))
		},
		controller.NewMendikotWebControllerWithProvider)
	games.RegisterKVGame("bhabhi", games.CategoryExtra,
		func() usecase.BhabhiInteractorIF {
			return usecase.NewBhabhiInteractor(domain.NewDefaultBhabhi(), new(presenter.BhabhiWebPresenter))
		},
		func(data []byte) (usecase.BhabhiInteractorIF, error) {
			return usecase.RestoreBhabhiInteractor(data, new(presenter.BhabhiWebPresenter))
		},
		controller.NewBhabhiWebControllerWithProvider)
	games.RegisterKVGame("sergeantmajor", games.CategoryExtra,
		func() usecase.SergeantMajorInteractorIF {
			return usecase.NewSergeantMajorInteractor(domain.NewDefaultSergeantMajor(), new(presenter.SergeantMajorWebPresenter))
		},
		func(data []byte) (usecase.SergeantMajorInteractorIF, error) {
			return usecase.RestoreSergeantMajorInteractor(data, new(presenter.SergeantMajorWebPresenter))
		},
		controller.NewSergeantMajorWebControllerWithProvider)
	games.RegisterKVGame("pasur", games.CategoryExtra,
		func() usecase.PasurInteractorIF {
			return usecase.NewPasurInteractor(domain.NewDefaultPasur(), new(presenter.PasurWebPresenter))
		},
		func(data []byte) (usecase.PasurInteractorIF, error) {
			return usecase.RestorePasurInteractor(data, new(presenter.PasurWebPresenter))
		},
		controller.NewPasurWebControllerWithProvider)
	games.RegisterKVGame("lingerlonger", games.CategoryExtra,
		func() usecase.LingerLongerInteractorIF {
			return usecase.NewLingerLongerInteractor(domain.NewDefaultLingerLonger(), new(presenter.LingerLongerWebPresenter))
		},
		func(data []byte) (usecase.LingerLongerInteractorIF, error) {
			return usecase.RestoreLingerLongerInteractor(data, new(presenter.LingerLongerWebPresenter))
		},
		controller.NewLingerLongerWebControllerWithProvider)
}
