//go:build js && wasm

// Package extra2 binds the Cloudflare Worker KV-backed handlers for the games
// assigned to the fifth size bucket. A worker main blank-imports this package
// for its registration side effects, so that whatever it registers is in place
// before games.RegisterCategory is called.
//
// Like casino/classic/solo/extra this is purely a binary-size bucket, not a
// user-facing taxonomy (ADR-0036). The colourless name is deliberate: it holds
// whatever had to move to keep every TinyGo WASM binary under the Cloudflare
// Workers free-tier 1 MB gzipped limit, and nothing about a game's genre says
// it belongs here.
package extra2

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func init() {
	games.RegisterKVGame("mighty", games.CategoryExtra2,
		func() usecase.MightyInteractorIF {
			return usecase.NewMightyInteractor(domain.NewDefaultMighty(), new(presenter.MightyWebPresenter))
		},
		func(data []byte) (usecase.MightyInteractorIF, error) {
			return usecase.RestoreMightyInteractor(data, new(presenter.MightyWebPresenter))
		},
		controller.NewMightyWebControllerWithProvider)
	games.RegisterKVGame("kemps", games.CategoryExtra2,
		func() usecase.KempsInteractorIF {
			return usecase.NewKempsInteractor(domain.NewDefaultKemps(), new(presenter.KempsWebPresenter))
		},
		func(data []byte) (usecase.KempsInteractorIF, error) {
			return usecase.RestoreKempsInteractor(data, new(presenter.KempsWebPresenter))
		},
		controller.NewKempsWebControllerWithProvider)
	games.RegisterKVGame("cuarenta", games.CategoryExtra2,
		func() usecase.CuarentaInteractorIF {
			return usecase.NewCuarentaInteractor(domain.NewDefaultCuarenta(), new(presenter.CuarentaWebPresenter))
		},
		func(data []byte) (usecase.CuarentaInteractorIF, error) {
			return usecase.RestoreCuarentaInteractor(data, new(presenter.CuarentaWebPresenter))
		},
		controller.NewCuarentaWebControllerWithProvider)
	games.RegisterKVGame("pishti", games.CategoryExtra2,
		func() usecase.PishtiInteractorIF {
			return usecase.NewPishtiInteractor(domain.NewDefaultPishti(), new(presenter.PishtiWebPresenter))
		},
		func(data []byte) (usecase.PishtiInteractorIF, error) {
			return usecase.RestorePishtiInteractor(data, new(presenter.PishtiWebPresenter))
		},
		controller.NewPishtiWebControllerWithProvider)
	games.RegisterKVGame("faro", games.CategoryExtra2,
		func() usecase.FaroInteractorIF {
			return usecase.NewFaroInteractor(domain.NewDefaultFaro(), new(presenter.FaroWebPresenter))
		},
		func(data []byte) (usecase.FaroInteractorIF, error) {
			return usecase.RestoreFaroInteractor(data, new(presenter.FaroWebPresenter))
		},
		controller.NewFaroWebControllerWithProvider)
	games.RegisterKVGame("beggarmyneighbour", games.CategoryExtra2,
		func() usecase.BeggarMyNeighbourInteractorIF {
			return usecase.NewBeggarMyNeighbourInteractor(domain.NewDefaultBeggarMyNeighbour(), new(presenter.BeggarMyNeighbourWebPresenter))
		},
		func(data []byte) (usecase.BeggarMyNeighbourInteractorIF, error) {
			return usecase.RestoreBeggarMyNeighbourInteractor(data, new(presenter.BeggarMyNeighbourWebPresenter))
		},
		controller.NewBeggarMyNeighbourWebControllerWithProvider)
	games.RegisterKVGame("nertz", games.CategoryExtra2,
		func() usecase.NertzInteractorIF {
			return usecase.NewNertzInteractor(domain.NewDefaultNertz(), new(presenter.NertzWebPresenter))
		},
		func(data []byte) (usecase.NertzInteractorIF, error) {
			return usecase.RestoreNertzInteractor(data, new(presenter.NertzWebPresenter))
		},
		controller.NewNertzWebControllerWithProvider)
	games.RegisterKVGame("pinochle", games.CategoryExtra2,
		func() usecase.PinochleInteractorIF {
			return usecase.NewPinochleInteractor(domain.NewDefaultPinochle(), new(presenter.PinochleWebPresenter))
		},
		func(data []byte) (usecase.PinochleInteractorIF, error) {
			return usecase.RestorePinochleInteractor(data, new(presenter.PinochleWebPresenter))
		},
		controller.NewPinochleWebControllerWithProvider)
	games.RegisterKVGame("doubt", games.CategoryExtra2,
		func() usecase.DoubtInteractorIF {
			return usecase.NewDoubtInteractor(domain.NewDefaultDoubt(), new(presenter.DoubtWebPresenter))
		},
		func(data []byte) (usecase.DoubtInteractorIF, error) {
			return usecase.RestoreDoubtInteractor(data, new(presenter.DoubtWebPresenter))
		},
		controller.NewDoubtWebControllerWithProvider)
	games.RegisterKVGame("spiteandmalice", games.CategoryExtra2,
		func() usecase.SpiteAndMaliceInteractorIF {
			return usecase.NewSpiteAndMaliceInteractor(domain.NewDefaultSpiteAndMalice(), new(presenter.SpiteAndMaliceWebPresenter))
		},
		func(data []byte) (usecase.SpiteAndMaliceInteractorIF, error) {
			return usecase.RestoreSpiteAndMaliceInteractor(data, new(presenter.SpiteAndMaliceWebPresenter))
		},
		controller.NewSpiteAndMaliceWebControllerWithProvider)
	games.RegisterKVGame("tichu", games.CategoryExtra2,
		func() usecase.TichuInteractorIF {
			return usecase.NewTichuInteractor(domain.NewDefaultTichu(), new(presenter.TichuWebPresenter))
		},
		func(data []byte) (usecase.TichuInteractorIF, error) {
			return usecase.RestoreTichuInteractor(data, new(presenter.TichuWebPresenter))
		},
		controller.NewTichuWebControllerWithProvider)
	games.RegisterKVGame("sixcardgolf", games.CategoryExtra2,
		func() usecase.SixCardGolfInteractorIF {
			return usecase.NewSixCardGolfInteractor(domain.NewDefaultSixCardGolf(), new(presenter.SixCardGolfWebPresenter))
		},
		func(data []byte) (usecase.SixCardGolfInteractorIF, error) {
			return usecase.RestoreSixCardGolfInteractor(data, new(presenter.SixCardGolfWebPresenter))
		},
		controller.NewSixCardGolfWebControllerWithProvider)
	games.RegisterKVGame("gofish", games.CategoryExtra2,
		func() usecase.GoFishInteractorIF {
			return usecase.NewGoFishInteractor(domain.NewDefaultGoFish(), new(presenter.GoFishWebPresenter))
		},
		func(data []byte) (usecase.GoFishInteractorIF, error) {
			return usecase.RestoreGoFishInteractor(data, new(presenter.GoFishWebPresenter))
		},
		controller.NewGoFishWebControllerWithProvider)
	games.RegisterKVGame("cuckoo", games.CategoryExtra2,
		func() usecase.CuckooInteractorIF {
			return usecase.NewCuckooInteractor(domain.NewDefaultCuckoo(), new(presenter.CuckooWebPresenter))
		},
		func(data []byte) (usecase.CuckooInteractorIF, error) {
			return usecase.RestoreCuckooInteractor(data, new(presenter.CuckooWebPresenter))
		},
		controller.NewCuckooWebControllerWithProvider)
	games.RegisterKVGame("bigtwo", games.CategoryExtra2,
		func() usecase.BigTwoInteractorIF {
			return usecase.NewBigTwoInteractor(domain.NewDefaultBigTwo(), new(presenter.BigTwoWebPresenter))
		},
		func(data []byte) (usecase.BigTwoInteractorIF, error) {
			return usecase.RestoreBigTwoInteractor(data, new(presenter.BigTwoWebPresenter))
		},
		controller.NewBigTwoWebControllerWithProvider)
	games.RegisterKVGame("spoons", games.CategoryExtra2,
		func() usecase.SpoonsInteractorIF {
			return usecase.NewSpoonsInteractor(domain.NewDefaultSpoons(), new(presenter.SpoonsWebPresenter))
		},
		func(data []byte) (usecase.SpoonsInteractorIF, error) {
			return usecase.RestoreSpoonsInteractor(data, new(presenter.SpoonsWebPresenter))
		},
		controller.NewSpoonsWebControllerWithProvider)
	games.RegisterKVGame("fiftyone", games.CategoryExtra2,
		func() usecase.FiftyOneInteractorIF {
			return usecase.NewFiftyOneInteractor(domain.NewDefaultFiftyOne(), new(presenter.FiftyOneWebPresenter))
		},
		func(data []byte) (usecase.FiftyOneInteractorIF, error) {
			return usecase.RestoreFiftyOneInteractor(data, new(presenter.FiftyOneWebPresenter))
		},
		controller.NewFiftyOneWebControllerWithProvider)
	games.RegisterKVGame("doubleklondike", games.CategoryExtra2,
		func() usecase.DoubleKlondikeInteractorIF {
			return usecase.NewDoubleKlondikeInteractor(domain.NewDefaultDoubleKlondike(), new(presenter.DoubleKlondikeWebPresenter))
		},
		func(data []byte) (usecase.DoubleKlondikeInteractorIF, error) {
			return usecase.RestoreDoubleKlondikeInteractor(data, new(presenter.DoubleKlondikeWebPresenter))
		},
		controller.NewDoubleKlondikeWebControllerWithProvider)
	games.RegisterKVGame("speed", games.CategoryExtra2,
		func() usecase.SpeedInteractorIF {
			return usecase.NewSpeedInteractor(domain.NewDefaultSpeed(), new(presenter.SpeedWebPresenter))
		},
		func(data []byte) (usecase.SpeedInteractorIF, error) {
			return usecase.RestoreSpeedInteractor(data, new(presenter.SpeedWebPresenter))
		},
		controller.NewSpeedWebControllerWithProvider)
	games.RegisterKVGame("pigtail", games.CategoryExtra2,
		func() usecase.PigsTailInteractorIF {
			return usecase.NewPigsTailInteractor(domain.NewDefaultPigsTail(), new(presenter.PigsTailWebPresenter))
		},
		func(data []byte) (usecase.PigsTailInteractorIF, error) {
			return usecase.RestorePigsTailInteractor(data, new(presenter.PigsTailWebPresenter))
		},
		controller.NewPigsTailWebControllerWithProvider)
	games.RegisterKVGame("trash", games.CategoryExtra2,
		func() usecase.TrashInteractorIF {
			return usecase.NewTrashInteractor(domain.NewDefaultTrash(), new(presenter.TrashWebPresenter))
		},
		func(data []byte) (usecase.TrashInteractorIF, error) {
			return usecase.RestoreTrashInteractor(data, new(presenter.TrashWebPresenter))
		},
		controller.NewTrashWebControllerWithProvider)
	games.RegisterKVGame("war", games.CategoryExtra2,
		func() usecase.WarInteractorIF {
			return usecase.NewWarInteractor(domain.NewDefaultWar(), new(presenter.WarWebPresenter))
		},
		func(data []byte) (usecase.WarInteractorIF, error) {
			return usecase.RestoreWarInteractor(data, new(presenter.WarWebPresenter))
		},
		controller.NewWarWebControllerWithProvider)
	games.RegisterKVGame("sirtommy", games.CategoryExtra2,
		func() usecase.SirTommyInteractorIF {
			return usecase.NewSirTommyInteractor(domain.NewDefaultSirTommy(), new(presenter.SirTommyWebPresenter))
		},
		func(data []byte) (usecase.SirTommyInteractorIF, error) {
			return usecase.RestoreSirTommyInteractor(data, new(presenter.SirTommyWebPresenter))
		},
		controller.NewSirTommyWebControllerWithProvider)
	games.RegisterKVGame("bisley", games.CategoryExtra2,
		func() usecase.BisleyInteractorIF {
			return usecase.NewBisleyInteractor(domain.NewDefaultBisley(), new(presenter.BisleyWebPresenter))
		},
		func(data []byte) (usecase.BisleyInteractorIF, error) {
			return usecase.RestoreBisleyInteractor(data, new(presenter.BisleyWebPresenter))
		},
		controller.NewBisleyWebControllerWithProvider)
	games.RegisterKVGame("napoleonssquare", games.CategoryExtra2,
		func() usecase.NapoleonsSquareInteractorIF {
			return usecase.NewNapoleonsSquareInteractor(domain.NewDefaultNapoleonsSquare(), new(presenter.NapoleonsSquareWebPresenter))
		},
		func(data []byte) (usecase.NapoleonsSquareInteractorIF, error) {
			return usecase.RestoreNapoleonsSquareInteractor(data, new(presenter.NapoleonsSquareWebPresenter))
		},
		controller.NewNapoleonsSquareWebControllerWithProvider)
	games.RegisterKVGame("grandfathersclock", games.CategoryExtra2,
		func() usecase.GrandfathersClockInteractorIF {
			return usecase.NewGrandfathersClockInteractor(domain.NewDefaultGrandfathersClock(), new(presenter.GrandfathersClockWebPresenter))
		},
		func(data []byte) (usecase.GrandfathersClockInteractorIF, error) {
			return usecase.RestoreGrandfathersClockInteractor(data, new(presenter.GrandfathersClockWebPresenter))
		},
		controller.NewGrandfathersClockWebControllerWithProvider)
	games.RegisterKVGame("missmilligan", games.CategoryExtra2,
		func() usecase.MissMilliganInteractorIF {
			return usecase.NewMissMilliganInteractor(domain.NewDefaultMissMilligan(), new(presenter.MissMilliganWebPresenter))
		},
		func(data []byte) (usecase.MissMilliganInteractorIF, error) {
			return usecase.RestoreMissMilliganInteractor(data, new(presenter.MissMilliganWebPresenter))
		},
		controller.NewMissMilliganWebControllerWithProvider)
	games.RegisterKVGame("duchess", games.CategoryExtra2,
		func() usecase.DuchessInteractorIF {
			return usecase.NewDuchessInteractor(domain.NewDefaultDuchess(), new(presenter.DuchessWebPresenter))
		},
		func(data []byte) (usecase.DuchessInteractorIF, error) {
			return usecase.RestoreDuchessInteractor(data, new(presenter.DuchessWebPresenter))
		},
		controller.NewDuchessWebControllerWithProvider)
	games.RegisterKVGame("windmill", games.CategoryExtra2,
		func() usecase.WindmillInteractorIF {
			return usecase.NewWindmillInteractor(domain.NewDefaultWindmill(), new(presenter.WindmillWebPresenter))
		},
		func(data []byte) (usecase.WindmillInteractorIF, error) {
			return usecase.RestoreWindmillInteractor(data, new(presenter.WindmillWebPresenter))
		},
		controller.NewWindmillWebControllerWithProvider)
}
