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
	games.RegisterKVGame("sultan", games.CategoryExtra,
		func() usecase.SultanInteractorIF {
			return usecase.NewSultanInteractor(domain.NewDefaultSultan(), new(presenter.SultanWebPresenter))
		},
		func(data []byte) (usecase.SultanInteractorIF, error) {
			return usecase.RestoreSultanInteractor(data, new(presenter.SultanWebPresenter))
		},
		controller.NewSultanWebControllerWithProvider)
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
	games.RegisterKVGame("machiavelli", games.CategoryExtra,
		func() usecase.MachiavelliInteractorIF {
			return usecase.NewMachiavelliInteractor(domain.NewDefaultMachiavelli(), new(presenter.MachiavelliWebPresenter))
		},
		func(data []byte) (usecase.MachiavelliInteractorIF, error) {
			return usecase.RestoreMachiavelliInteractor(data, new(presenter.MachiavelliWebPresenter))
		},
		controller.NewMachiavelliWebControllerWithProvider)
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
	games.RegisterKVGame("diplomat", games.CategoryExtra,
		func() usecase.DiplomatInteractorIF {
			return usecase.NewDiplomatInteractor(domain.NewDefaultDiplomat(), new(presenter.DiplomatWebPresenter))
		},
		func(data []byte) (usecase.DiplomatInteractorIF, error) {
			return usecase.RestoreDiplomatInteractor(data, new(presenter.DiplomatWebPresenter))
		},
		controller.NewDiplomatWebControllerWithProvider)
	games.RegisterKVGame("zwanzigerrufen", games.CategoryExtra,
		func() usecase.ZwanzigerrufenInteractorIF {
			return usecase.NewZwanzigerrufenInteractor(domain.NewDefaultZwanzigerrufen(), new(presenter.ZwanzigerrufenWebPresenter))
		},
		func(data []byte) (usecase.ZwanzigerrufenInteractorIF, error) {
			return usecase.RestoreZwanzigerrufenInteractor(data, new(presenter.ZwanzigerrufenWebPresenter))
		},
		controller.NewZwanzigerrufenWebControllerWithProvider)
	games.RegisterKVGame("troggu", games.CategoryExtra,
		func() usecase.TrogguInteractorIF {
			return usecase.NewTrogguInteractor(domain.NewDefaultTroggu(), new(presenter.TrogguWebPresenter))
		},
		func(data []byte) (usecase.TrogguInteractorIF, error) {
			return usecase.RestoreTrogguInteractor(data, new(presenter.TrogguWebPresenter))
		},
		controller.NewTrogguWebControllerWithProvider)
	games.RegisterKVGame("sthelena", games.CategoryExtra,
		func() usecase.StHelenaInteractorIF {
			return usecase.NewStHelenaInteractor(domain.NewDefaultStHelena(), new(presenter.StHelenaWebPresenter))
		},
		func(data []byte) (usecase.StHelenaInteractorIF, error) {
			return usecase.RestoreStHelenaInteractor(data, new(presenter.StHelenaWebPresenter))
		},
		controller.NewStHelenaWebControllerWithProvider)
	games.RegisterKVGame("matrimony", games.CategoryExtra,
		func() usecase.MatrimonyInteractorIF {
			return usecase.NewMatrimonyInteractor(domain.NewDefaultMatrimony(), new(presenter.MatrimonyWebPresenter))
		},
		func(data []byte) (usecase.MatrimonyInteractorIF, error) {
			return usecase.RestoreMatrimonyInteractor(data, new(presenter.MatrimonyWebPresenter))
		},
		controller.NewMatrimonyWebControllerWithProvider)
	games.RegisterKVGame("tapptarock", games.CategoryExtra,
		func() usecase.TappTarockInteractorIF {
			return usecase.NewTappTarockInteractor(domain.NewDefaultTappTarock(), new(presenter.TappTarockWebPresenter))
		},
		func(data []byte) (usecase.TappTarockInteractorIF, error) {
			return usecase.RestoreTappTarockInteractor(data, new(presenter.TappTarockWebPresenter))
		},
		controller.NewTappTarockWebControllerWithProvider)
	games.RegisterKVGame("indianrummy", games.CategoryExtra,
		func() usecase.IndianRummyInteractorIF {
			return usecase.NewIndianRummyInteractor(domain.NewDefaultIndianRummy(), new(presenter.IndianRummyWebPresenter))
		},
		func(data []byte) (usecase.IndianRummyInteractorIF, error) {
			return usecase.RestoreIndianRummyInteractor(data, new(presenter.IndianRummyWebPresenter))
		},
		controller.NewIndianRummyWebControllerWithProvider)
	games.RegisterKVGame("pan", games.CategoryExtra,
		func() usecase.PanInteractorIF {
			return usecase.NewPanInteractor(domain.NewDefaultPan(), new(presenter.PanWebPresenter))
		},
		func(data []byte) (usecase.PanInteractorIF, error) {
			return usecase.RestorePanInteractor(data, new(presenter.PanWebPresenter))
		},
		controller.NewPanWebControllerWithProvider)
}
