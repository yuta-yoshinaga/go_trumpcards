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
	games.RegisterKVGame("fortyandeight", games.CategoryExtra,
		func() usecase.FortyAndEightInteractorIF {
			return usecase.NewFortyAndEightInteractor(domain.NewDefaultFortyAndEight(), new(presenter.FortyAndEightWebPresenter))
		},
		func(data []byte) (usecase.FortyAndEightInteractorIF, error) {
			return usecase.RestoreFortyAndEightInteractor(data, new(presenter.FortyAndEightWebPresenter))
		},
		controller.NewFortyAndEightWebControllerWithProvider)
	games.RegisterKVGame("agnes", games.CategoryExtra,
		func() usecase.AgnesInteractorIF {
			return usecase.NewAgnesInteractor(domain.NewDefaultAgnes(), new(presenter.AgnesWebPresenter))
		},
		func(data []byte) (usecase.AgnesInteractorIF, error) {
			return usecase.RestoreAgnesInteractor(data, new(presenter.AgnesWebPresenter))
		},
		controller.NewAgnesWebControllerWithProvider)
}
