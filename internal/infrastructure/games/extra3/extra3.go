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
	games.RegisterKVGame("cego", games.CategoryExtra3,
		func() usecase.CegoInteractorIF {
			return usecase.NewCegoInteractor(domain.NewDefaultCego(), new(presenter.CegoWebPresenter))
		},
		func(data []byte) (usecase.CegoInteractorIF, error) {
			return usecase.RestoreCegoInteractor(data, new(presenter.CegoWebPresenter))
		},
		controller.NewCegoWebControllerWithProvider)
	games.RegisterKVGame("toepen", games.CategoryExtra3,
		func() usecase.ToepenInteractorIF {
			return usecase.NewToepenInteractor(domain.NewDefaultToepen(), new(presenter.ToepenWebPresenter))
		},
		func(data []byte) (usecase.ToepenInteractorIF, error) {
			return usecase.RestoreToepenInteractor(data, new(presenter.ToepenWebPresenter))
		},
		controller.NewToepenWebControllerWithProvider)
	games.RegisterKVGame("ulti", games.CategoryExtra3,
		func() usecase.UltiInteractorIF {
			return usecase.NewUltiInteractor(domain.NewDefaultUlti(), new(presenter.UltiWebPresenter))
		},
		func(data []byte) (usecase.UltiInteractorIF, error) {
			return usecase.RestoreUltiInteractor(data, new(presenter.UltiWebPresenter))
		},
		controller.NewUltiWebControllerWithProvider)
	games.RegisterKVGame("scarto", games.CategoryExtra3,
		func() usecase.ScartoInteractorIF {
			return usecase.NewScartoInteractor(domain.NewDefaultScarto(), new(presenter.ScartoWebPresenter))
		},
		func(data []byte) (usecase.ScartoInteractorIF, error) {
			return usecase.RestoreScartoInteractor(data, new(presenter.ScartoWebPresenter))
		},
		controller.NewScartoWebControllerWithProvider)
	games.RegisterKVGame("piquet", games.CategoryExtra3,
		func() usecase.PiquetInteractorIF {
			return usecase.NewPiquetInteractor(domain.NewDefaultPiquet(), new(presenter.PiquetWebPresenter))
		},
		func(data []byte) (usecase.PiquetInteractorIF, error) {
			return usecase.RestorePiquetInteractor(data, new(presenter.PiquetWebPresenter))
		},
		controller.NewPiquetWebControllerWithProvider)
	games.RegisterKVGame("cribbage", games.CategoryExtra3,
		func() usecase.CribbageInteractorIF {
			return usecase.NewCribbageInteractor(domain.NewDefaultCribbage(), new(presenter.CribbageWebPresenter))
		},
		func(data []byte) (usecase.CribbageInteractorIF, error) {
			return usecase.RestoreCribbageInteractor(data, new(presenter.CribbageWebPresenter))
		},
		controller.NewCribbageWebControllerWithProvider)
	games.RegisterKVGame("mao", games.CategoryExtra3,
		func() usecase.MaoInteractorIF {
			return usecase.NewMaoInteractor(domain.NewDefaultMao(), new(presenter.MaoWebPresenter))
		},
		func(data []byte) (usecase.MaoInteractorIF, error) {
			return usecase.RestoreMaoInteractor(data, new(presenter.MaoWebPresenter))
		},
		controller.NewMaoWebControllerWithProvider)
	games.RegisterKVGame("sevenbridge", games.CategoryExtra3,
		func() usecase.SevenBridgeInteractorIF {
			return usecase.NewSevenBridgeInteractor(domain.NewDefaultSevenBridge(), new(presenter.SevenBridgeWebPresenter))
		},
		func(data []byte) (usecase.SevenBridgeInteractorIF, error) {
			return usecase.RestoreSevenBridgeInteractor(data, new(presenter.SevenBridgeWebPresenter))
		},
		controller.NewSevenBridgeWebControllerWithProvider)
	games.RegisterKVGame("ombre", games.CategoryExtra3,
		func() usecase.OmbreInteractorIF {
			return usecase.NewOmbreInteractor(domain.NewDefaultOmbre(), new(presenter.OmbreWebPresenter))
		},
		func(data []byte) (usecase.OmbreInteractorIF, error) {
			return usecase.RestoreOmbreInteractor(data, new(presenter.OmbreWebPresenter))
		},
		controller.NewOmbreWebControllerWithProvider)
	games.RegisterKVGame("koikoi", games.CategoryExtra3,
		func() usecase.KoiKoiInteractorIF {
			return usecase.NewKoiKoiInteractor(domain.NewDefaultKoiKoi(), new(presenter.KoiKoiWebPresenter))
		},
		func(data []byte) (usecase.KoiKoiInteractorIF, error) {
			return usecase.RestoreKoiKoiInteractor(data, new(presenter.KoiKoiWebPresenter))
		},
		controller.NewKoiKoiWebControllerWithProvider)
	games.RegisterKVGame("rook", games.CategoryExtra3,
		func() usecase.RookInteractorIF {
			return usecase.NewRookInteractor(domain.NewDefaultRook(), new(presenter.RookWebPresenter))
		},
		func(data []byte) (usecase.RookInteractorIF, error) {
			return usecase.RestoreRookInteractor(data, new(presenter.RookWebPresenter))
		},
		controller.NewRookWebControllerWithProvider)
	games.RegisterKVGame("jass", games.CategoryExtra3,
		func() usecase.JassInteractorIF {
			return usecase.NewJassInteractor(domain.NewDefaultJass(), new(presenter.JassWebPresenter))
		},
		func(data []byte) (usecase.JassInteractorIF, error) {
			return usecase.RestoreJassInteractor(data, new(presenter.JassWebPresenter))
		},
		controller.NewJassWebControllerWithProvider)
	games.RegisterKVGame("michigan", games.CategoryExtra3,
		func() usecase.MichiganInteractorIF {
			return usecase.NewMichiganInteractor(domain.NewDefaultMichigan(), new(presenter.MichiganWebPresenter))
		},
		func(data []byte) (usecase.MichiganInteractorIF, error) {
			return usecase.RestoreMichiganInteractor(data, new(presenter.MichiganWebPresenter))
		},
		controller.NewMichiganWebControllerWithProvider)
	games.RegisterKVGame("loo", games.CategoryExtra3,
		func() usecase.LooInteractorIF {
			return usecase.NewLooInteractor(domain.NewDefaultLoo(), new(presenter.LooWebPresenter))
		},
		func(data []byte) (usecase.LooInteractorIF, error) {
			return usecase.RestoreLooInteractor(data, new(presenter.LooWebPresenter))
		},
		controller.NewLooWebControllerWithProvider)
	games.RegisterKVGame("wizard", games.CategoryExtra3,
		func() usecase.WizardInteractorIF {
			return usecase.NewWizardInteractor(domain.NewDefaultWizard(), new(presenter.WizardWebPresenter))
		},
		func(data []byte) (usecase.WizardInteractorIF, error) {
			return usecase.RestoreWizardInteractor(data, new(presenter.WizardWebPresenter))
		},
		controller.NewWizardWebControllerWithProvider)
	games.RegisterKVGame("bouillotte", games.CategoryExtra3,
		func() usecase.BouillotteInteractorIF {
			return usecase.NewBouillotteInteractor(domain.NewDefaultBouillotte(), new(presenter.BouillotteWebPresenter))
		},
		func(data []byte) (usecase.BouillotteInteractorIF, error) {
			return usecase.RestoreBouillotteInteractor(data, new(presenter.BouillotteWebPresenter))
		},
		controller.NewBouillotteWebControllerWithProvider)
	games.RegisterKVGame("tablanet", games.CategoryExtra3,
		func() usecase.TablanetInteractorIF {
			return usecase.NewTablanetInteractor(domain.NewDefaultTablanet(), new(presenter.TablanetWebPresenter))
		},
		func(data []byte) (usecase.TablanetInteractorIF, error) {
			return usecase.RestoreTablanetInteractor(data, new(presenter.TablanetWebPresenter))
		},
		controller.NewTablanetWebControllerWithProvider)
	games.RegisterKVGame("primero", games.CategoryExtra3,
		func() usecase.PrimeroInteractorIF {
			return usecase.NewPrimeroInteractor(domain.NewDefaultPrimero(), new(presenter.PrimeroWebPresenter))
		},
		func(data []byte) (usecase.PrimeroInteractorIF, error) {
			return usecase.RestorePrimeroInteractor(data, new(presenter.PrimeroWebPresenter))
		},
		controller.NewPrimeroWebControllerWithProvider)
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
	games.RegisterKVGame("skat", games.CategoryExtra3,
		func() usecase.SkatInteractorIF {
			return usecase.NewSkatInteractor(domain.NewDefaultSkat(), new(presenter.SkatWebPresenter))
		},
		func(data []byte) (usecase.SkatInteractorIF, error) {
			return usecase.RestoreSkatInteractor(data, new(presenter.SkatWebPresenter))
		},
		controller.NewSkatWebControllerWithProvider)
	games.RegisterKVGame("congress", games.CategoryExtra3,
		func() usecase.CongressInteractorIF {
			return usecase.NewCongressInteractor(domain.NewDefaultCongress(), new(presenter.CongressWebPresenter))
		},
		func(data []byte) (usecase.CongressInteractorIF, error) {
			return usecase.RestoreCongressInteractor(data, new(presenter.CongressWebPresenter))
		},
		controller.NewCongressWebControllerWithProvider)
	games.RegisterKVGame("terrace", games.CategoryExtra3,
		func() usecase.TerraceInteractorIF {
			return usecase.NewTerraceInteractor(domain.NewDefaultTerrace(), new(presenter.TerraceWebPresenter))
		},
		func(data []byte) (usecase.TerraceInteractorIF, error) {
			return usecase.RestoreTerraceInteractor(data, new(presenter.TerraceWebPresenter))
		},
		controller.NewTerraceWebControllerWithProvider)
	games.RegisterKVGame("belote", games.CategoryExtra3,
		func() usecase.BeloteInteractorIF {
			return usecase.NewBeloteInteractor(domain.NewDefaultBelote(), new(presenter.BeloteWebPresenter))
		},
		func(data []byte) (usecase.BeloteInteractorIF, error) {
			return usecase.RestoreBeloteInteractor(data, new(presenter.BeloteWebPresenter))
		},
		controller.NewBeloteWebControllerWithProvider)
	games.RegisterKVGame("sheepshead", games.CategoryExtra3,
		func() usecase.SheepsheadInteractorIF {
			return usecase.NewSheepsheadInteractor(domain.NewDefaultSheepshead(), new(presenter.SheepsheadWebPresenter))
		},
		func(data []byte) (usecase.SheepsheadInteractorIF, error) {
			return usecase.RestoreSheepsheadInteractor(data, new(presenter.SheepsheadWebPresenter))
		},
		controller.NewSheepsheadWebControllerWithProvider)
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
	games.RegisterKVGame("poch", games.CategoryExtra3,
		func() usecase.PochInteractorIF {
			return usecase.NewPochInteractor(domain.NewDefaultPoch(), new(presenter.PochWebPresenter))
		},
		func(data []byte) (usecase.PochInteractorIF, error) {
			return usecase.RestorePochInteractor(data, new(presenter.PochWebPresenter))
		},
		controller.NewPochWebControllerWithProvider)
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
}
