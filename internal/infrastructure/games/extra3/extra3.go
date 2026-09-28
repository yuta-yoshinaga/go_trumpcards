//go:build js && wasm

// Package extra3 binds the Cloudflare Worker KV-backed handlers for the games
// assigned to the sixth size bucket. A worker main blank-imports this package
// for its registration side effects, so that whatever it registers is in place
// before games.RegisterCategory is called.
//
// Like casino/classic/solo/extra this is purely a binary-size bucket, not a
// user-facing taxonomy (ADR-0036). The colourless name is deliberate: it holds
// whatever had to move to keep every TinyGo WASM binary under the Cloudflare
// Workers free-tier 1 MB gzipped limit, and nothing about a game's genre says
// it belongs here.
package extra3

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func init() {
	games.RegisterKVGame("madrasso", games.CategoryExtra3,
		func() usecase.MadrassoInteractorIF {
			return usecase.NewMadrassoInteractor(domain.NewDefaultMadrasso(), new(presenter.MadrassoWebPresenter))
		},
		func(data []byte) (usecase.MadrassoInteractorIF, error) {
			return usecase.RestoreMadrassoInteractor(data, new(presenter.MadrassoWebPresenter))
		},
		controller.NewMadrassoWebControllerWithProvider)
	games.RegisterKVGame("toepen", games.CategoryExtra3,
		func() usecase.ToepenInteractorIF {
			return usecase.NewToepenInteractor(domain.NewDefaultToepen(), new(presenter.ToepenWebPresenter))
		},
		func(data []byte) (usecase.ToepenInteractorIF, error) {
			return usecase.RestoreToepenInteractor(data, new(presenter.ToepenWebPresenter))
		},
		controller.NewToepenWebControllerWithProvider)
	games.RegisterKVGame("sevenbridge", games.CategoryExtra3,
		func() usecase.SevenBridgeInteractorIF {
			return usecase.NewSevenBridgeInteractor(domain.NewDefaultSevenBridge(), new(presenter.SevenBridgeWebPresenter))
		},
		func(data []byte) (usecase.SevenBridgeInteractorIF, error) {
			return usecase.RestoreSevenBridgeInteractor(data, new(presenter.SevenBridgeWebPresenter))
		},
		controller.NewSevenBridgeWebControllerWithProvider)
	games.RegisterKVGame("koikoi", games.CategoryExtra3,
		func() usecase.KoiKoiInteractorIF {
			return usecase.NewKoiKoiInteractor(domain.NewDefaultKoiKoi(), new(presenter.KoiKoiWebPresenter))
		},
		func(data []byte) (usecase.KoiKoiInteractorIF, error) {
			return usecase.RestoreKoiKoiInteractor(data, new(presenter.KoiKoiWebPresenter))
		},
		controller.NewKoiKoiWebControllerWithProvider)
	games.RegisterKVGame("tablanet", games.CategoryExtra3,
		func() usecase.TablanetInteractorIF {
			return usecase.NewTablanetInteractor(domain.NewDefaultTablanet(), new(presenter.TablanetWebPresenter))
		},
		func(data []byte) (usecase.TablanetInteractorIF, error) {
			return usecase.RestoreTablanetInteractor(data, new(presenter.TablanetWebPresenter))
		},
		controller.NewTablanetWebControllerWithProvider)
	games.RegisterKVGame("basra", games.CategoryExtra3,
		func() usecase.BasraInteractorIF {
			return usecase.NewBasraInteractor(domain.NewDefaultBasra(), new(presenter.BasraWebPresenter))
		},
		func(data []byte) (usecase.BasraInteractorIF, error) {
			return usecase.RestoreBasraInteractor(data, new(presenter.BasraWebPresenter))
		},
		controller.NewBasraWebControllerWithProvider)
	games.RegisterKVGame("fortyandeight", games.CategoryExtra3,
		func() usecase.FortyAndEightInteractorIF {
			return usecase.NewFortyAndEightInteractor(domain.NewDefaultFortyAndEight(), new(presenter.FortyAndEightWebPresenter))
		},
		func(data []byte) (usecase.FortyAndEightInteractorIF, error) {
			return usecase.RestoreFortyAndEightInteractor(data, new(presenter.FortyAndEightWebPresenter))
		},
		controller.NewFortyAndEightWebControllerWithProvider)
	games.RegisterKVGame("bridge", games.CategoryExtra3,
		func() usecase.BridgeInteractorIF {
			return usecase.NewBridgeInteractor(domain.NewDefaultBridge(), new(presenter.BridgeWebPresenter))
		},
		func(data []byte) (usecase.BridgeInteractorIF, error) {
			return usecase.RestoreBridgeInteractor(data, new(presenter.BridgeWebPresenter))
		},
		controller.NewBridgeWebControllerWithProvider)
	games.RegisterKVGame("congress", games.CategoryExtra3,
		func() usecase.CongressInteractorIF {
			return usecase.NewCongressInteractor(domain.NewDefaultCongress(), new(presenter.CongressWebPresenter))
		},
		func(data []byte) (usecase.CongressInteractorIF, error) {
			return usecase.RestoreCongressInteractor(data, new(presenter.CongressWebPresenter))
		},
		controller.NewCongressWebControllerWithProvider)
	games.RegisterKVGame("saliclaw", games.CategoryExtra3,
		func() usecase.SalicLawInteractorIF {
			return usecase.NewSalicLawInteractor(domain.NewDefaultSalicLaw(), new(presenter.SalicLawWebPresenter))
		},
		func(data []byte) (usecase.SalicLawInteractorIF, error) {
			return usecase.RestoreSalicLawInteractor(data, new(presenter.SalicLawWebPresenter))
		},
		controller.NewSalicLawWebControllerWithProvider)
	games.RegisterKVGame("terrace", games.CategoryExtra3,
		func() usecase.TerraceInteractorIF {
			return usecase.NewTerraceInteractor(domain.NewDefaultTerrace(), new(presenter.TerraceWebPresenter))
		},
		func(data []byte) (usecase.TerraceInteractorIF, error) {
			return usecase.RestoreTerraceInteractor(data, new(presenter.TerraceWebPresenter))
		},
		controller.NewTerraceWebControllerWithProvider)
	games.RegisterKVGame("niuniu", games.CategoryExtra3,
		func() usecase.NiuNiuInteractorIF {
			return usecase.NewNiuNiuInteractor(domain.NewDefaultNiuNiu(), new(presenter.NiuNiuWebPresenter))
		},
		func(data []byte) (usecase.NiuNiuInteractorIF, error) {
			return usecase.RestoreNiuNiuInteractor(data, new(presenter.NiuNiuWebPresenter))
		},
		controller.NewNiuNiuWebControllerWithProvider)
	games.RegisterKVGame("bura", games.CategoryExtra3,
		func() usecase.BuraInteractorIF {
			return usecase.NewBuraInteractor(domain.NewDefaultBura(), new(presenter.BuraWebPresenter))
		},
		func(data []byte) (usecase.BuraInteractorIF, error) {
			return usecase.RestoreBuraInteractor(data, new(presenter.BuraWebPresenter))
		},
		controller.NewBuraWebControllerWithProvider)
	games.RegisterKVGame("trex", games.CategoryExtra3,
		func() usecase.TrexInteractorIF {
			return usecase.NewTrexInteractor(domain.NewDefaultTrex(), new(presenter.TrexWebPresenter))
		},
		func(data []byte) (usecase.TrexInteractorIF, error) {
			return usecase.RestoreTrexInteractor(data, new(presenter.TrexWebPresenter))
		},
		controller.NewTrexWebControllerWithProvider)
	games.RegisterKVGame("skitgubbe", games.CategoryExtra3,
		func() usecase.SkitgubbeInteractorIF {
			return usecase.NewSkitgubbeInteractor(domain.NewDefaultSkitgubbe(), new(presenter.SkitgubbeWebPresenter))
		},
		func(data []byte) (usecase.SkitgubbeInteractorIF, error) {
			return usecase.RestoreSkitgubbeInteractor(data, new(presenter.SkitgubbeWebPresenter))
		},
		controller.NewSkitgubbeWebControllerWithProvider)
	games.RegisterKVGame("desmoche", games.CategoryExtra3,
		func() usecase.DesmocheInteractorIF {
			return usecase.NewDesmocheInteractor(domain.NewDefaultDesmoche(), new(presenter.DesmocheWebPresenter))
		},
		func(data []byte) (usecase.DesmocheInteractorIF, error) {
			return usecase.RestoreDesmocheInteractor(data, new(presenter.DesmocheWebPresenter))
		},
		controller.NewDesmocheWebControllerWithProvider)
	games.RegisterKVGame("popejoan", games.CategoryExtra3,
		func() usecase.PopeJoanInteractorIF {
			return usecase.NewPopeJoanInteractor(domain.NewDefaultPopeJoan(), new(presenter.PopeJoanWebPresenter))
		},
		func(data []byte) (usecase.PopeJoanInteractorIF, error) {
			return usecase.RestorePopeJoanInteractor(data, new(presenter.PopeJoanWebPresenter))
		},
		controller.NewPopeJoanWebControllerWithProvider)
	games.RegisterKVGame("nainjaune", games.CategoryExtra3,
		func() usecase.NainJauneInteractorIF {
			return usecase.NewNainJauneInteractor(domain.NewDefaultNainJaune(), new(presenter.NainJauneWebPresenter))
		},
		func(data []byte) (usecase.NainJauneInteractorIF, error) {
			return usecase.RestoreNainJauneInteractor(data, new(presenter.NainJauneWebPresenter))
		},
		controller.NewNainJauneWebControllerWithProvider)
	games.RegisterKVGame("boston", games.CategoryExtra3,
		func() usecase.BostonInteractorIF {
			return usecase.NewBostonInteractor(domain.NewDefaultBoston(), new(presenter.BostonWebPresenter))
		},
		func(data []byte) (usecase.BostonInteractorIF, error) {
			return usecase.RestoreBostonInteractor(data, new(presenter.BostonWebPresenter))
		},
		controller.NewBostonWebControllerWithProvider)
	games.RegisterKVGame("vint", games.CategoryExtra3,
		func() usecase.VintInteractorIF {
			return usecase.NewVintInteractor(domain.NewDefaultVint(), new(presenter.VintWebPresenter))
		},
		func(data []byte) (usecase.VintInteractorIF, error) {
			return usecase.RestoreVintInteractor(data, new(presenter.VintWebPresenter))
		},
		controller.NewVintWebControllerWithProvider)
	games.RegisterKVGame("rollingstone", games.CategoryExtra3,
		func() usecase.RollingStoneInteractorIF {
			return usecase.NewRollingStoneInteractor(domain.NewDefaultRollingStone(), new(presenter.RollingStoneWebPresenter))
		},
		func(data []byte) (usecase.RollingStoneInteractorIF, error) {
			return usecase.RestoreRollingStoneInteractor(data, new(presenter.RollingStoneWebPresenter))
		},
		controller.NewRollingStoneWebControllerWithProvider)
	games.RegisterKVGame("stealingbundles", games.CategoryExtra3,
		func() usecase.StealingBundlesInteractorIF {
			return usecase.NewStealingBundlesInteractor(domain.NewDefaultStealingBundles(), new(presenter.StealingBundlesWebPresenter))
		},
		func(data []byte) (usecase.StealingBundlesInteractorIF, error) {
			return usecase.RestoreStealingBundlesInteractor(data, new(presenter.StealingBundlesWebPresenter))
		},
		controller.NewStealingBundlesWebControllerWithProvider)
	games.RegisterKVGame("sakura", games.CategoryExtra3,
		func() usecase.SakuraInteractorIF {
			return usecase.NewSakuraInteractor(domain.NewDefaultSakura(), new(presenter.SakuraWebPresenter))
		},
		func(data []byte) (usecase.SakuraInteractorIF, error) {
			return usecase.RestoreSakuraInteractor(data, new(presenter.SakuraWebPresenter))
		},
		controller.NewSakuraWebControllerWithProvider)
	games.RegisterKVGame("bigben", games.CategoryExtra3,
		func() usecase.BigBenInteractorIF {
			return usecase.NewBigBenInteractor(domain.NewDefaultBigBen(), new(presenter.BigBenWebPresenter))
		},
		func(data []byte) (usecase.BigBenInteractorIF, error) {
			return usecase.RestoreBigBenInteractor(data, new(presenter.BigBenWebPresenter))
		},
		controller.NewBigBenWebControllerWithProvider)
}
