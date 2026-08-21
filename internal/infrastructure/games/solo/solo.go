//go:build js && wasm

// Package solo binds the Cloudflare Worker KV-backed handlers for the
// 27 solitaire and rummy variants. A worker main must blank-import this
// package so the init below runs before games.RegisterCategory is called.
package solo

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func init() {
	games.RegisterKVGame("minchiate", games.CategorySolo,
		func() usecase.MinchiateInteractorIF {
			return usecase.NewMinchiateInteractor(domain.NewDefaultMinchiate(), new(presenter.MinchiateWebPresenter))
		},
		func(data []byte) (usecase.MinchiateInteractorIF, error) {
			return usecase.RestoreMinchiateInteractor(data, new(presenter.MinchiateWebPresenter))
		},
		controller.NewMinchiateWebControllerWithProvider)
	games.RegisterKVGame("tarocchini", games.CategorySolo,
		func() usecase.TarocchiniInteractorIF {
			return usecase.NewTarocchiniInteractor(domain.NewDefaultTarocchini(), new(presenter.TarocchiniWebPresenter))
		},
		func(data []byte) (usecase.TarocchiniInteractorIF, error) {
			return usecase.RestoreTarocchiniInteractor(data, new(presenter.TarocchiniWebPresenter))
		},
		controller.NewTarocchiniWebControllerWithProvider)
	games.RegisterKVGame("memory", games.CategorySolo,
		func() usecase.MemoryInteractorIF {
			return usecase.NewMemoryInteractor(domain.NewDefaultMemory(), new(presenter.MemoryWebPresenter))
		},
		func(data []byte) (usecase.MemoryInteractorIF, error) {
			return usecase.RestoreMemoryInteractor(data, new(presenter.MemoryWebPresenter))
		},
		controller.NewMemoryWebControllerWithProvider)
	games.RegisterKVGame("klondike", games.CategorySolo,
		func() usecase.KlondikeInteractorIF {
			return usecase.NewKlondikeInteractor(domain.NewDefaultKlondike(), new(presenter.KlondikeWebPresenter))
		},
		func(data []byte) (usecase.KlondikeInteractorIF, error) {
			return usecase.RestoreKlondikeInteractor(data, new(presenter.KlondikeWebPresenter))
		},
		controller.NewKlondikeWebControllerWithProvider)
	games.RegisterKVGame("freecell", games.CategorySolo,
		func() usecase.FreeCellInteractorIF {
			return usecase.NewFreeCellInteractor(domain.NewDefaultFreeCell(), new(presenter.FreeCellWebPresenter))
		},
		func(data []byte) (usecase.FreeCellInteractorIF, error) {
			return usecase.RestoreFreeCellInteractor(data, new(presenter.FreeCellWebPresenter))
		},
		controller.NewFreeCellWebControllerWithProvider)
	// Baker's Game reuses the FreeCell engine (same-suit stacking variant).
	// The sameSuit flag is serialised in the snapshot, so the restore path
	// rebuilds the correct variant from KV automatically.
	games.RegisterKVGame("bakersgame", games.CategorySolo,
		func() usecase.FreeCellInteractorIF {
			return usecase.NewFreeCellInteractor(domain.NewDefaultBakersGame(), new(presenter.BakersGameWebPresenter))
		},
		func(data []byte) (usecase.FreeCellInteractorIF, error) {
			return usecase.RestoreFreeCellInteractor(data, new(presenter.BakersGameWebPresenter))
		},
		controller.NewFreeCellWebControllerWithProvider)
	games.RegisterKVGame("seahaventowers", games.CategorySolo,
		func() usecase.SeahavenTowersInteractorIF {
			return usecase.NewSeahavenTowersInteractor(domain.NewDefaultSeahavenTowers(), new(presenter.SeahavenTowersWebPresenter))
		},
		func(data []byte) (usecase.SeahavenTowersInteractorIF, error) {
			return usecase.RestoreSeahavenTowersInteractor(data, new(presenter.SeahavenTowersWebPresenter))
		},
		controller.NewSeahavenTowersWebControllerWithProvider)
	games.RegisterKVGame("cruel", games.CategorySolo,
		func() usecase.CruelInteractorIF {
			return usecase.NewCruelInteractor(domain.NewDefaultCruel(), new(presenter.CruelWebPresenter))
		},
		func(data []byte) (usecase.CruelInteractorIF, error) {
			return usecase.RestoreCruelInteractor(data, new(presenter.CruelWebPresenter))
		},
		controller.NewCruelWebControllerWithProvider)
	games.RegisterKVGame("spider", games.CategorySolo,
		func() usecase.SpiderInteractorIF {
			return usecase.NewSpiderInteractor(domain.NewDefaultSpider(), new(presenter.SpiderWebPresenter))
		},
		func(data []byte) (usecase.SpiderInteractorIF, error) {
			return usecase.RestoreSpiderInteractor(data, new(presenter.SpiderWebPresenter))
		},
		controller.NewSpiderWebControllerWithProvider)
	games.RegisterKVGame("pyramid", games.CategorySolo,
		func() usecase.PyramidInteractorIF {
			return usecase.NewPyramidInteractor(domain.NewDefaultPyramid(), new(presenter.PyramidWebPresenter))
		},
		func(data []byte) (usecase.PyramidInteractorIF, error) {
			return usecase.RestorePyramidInteractor(data, new(presenter.PyramidWebPresenter))
		},
		controller.NewPyramidWebControllerWithProvider)
	games.RegisterKVGame("tripeaks", games.CategorySolo,
		func() usecase.TriPeaksInteractorIF {
			return usecase.NewTriPeaksInteractor(domain.NewDefaultTriPeaks(), new(presenter.TriPeaksWebPresenter))
		},
		func(data []byte) (usecase.TriPeaksInteractorIF, error) {
			return usecase.RestoreTriPeaksInteractor(data, new(presenter.TriPeaksWebPresenter))
		},
		controller.NewTriPeaksWebControllerWithProvider)
	games.RegisterKVGame("golf", games.CategorySolo,
		func() usecase.GolfInteractorIF {
			return usecase.NewGolfInteractor(domain.NewDefaultGolf(), new(presenter.GolfWebPresenter))
		},
		func(data []byte) (usecase.GolfInteractorIF, error) {
			return usecase.RestoreGolfInteractor(data, new(presenter.GolfWebPresenter))
		},
		controller.NewGolfWebControllerWithProvider)
	games.RegisterKVGame("clocksolitaire", games.CategorySolo,
		func() usecase.ClockSolitaireInteractorIF {
			return usecase.NewClockSolitaireInteractor(domain.NewDefaultClockSolitaire(), new(presenter.ClockSolitaireWebPresenter))
		},
		func(data []byte) (usecase.ClockSolitaireInteractorIF, error) {
			return usecase.RestoreClockSolitaireInteractor(data, new(presenter.ClockSolitaireWebPresenter))
		},
		controller.NewClockSolitaireWebControllerWithProvider)
	games.RegisterKVGame("fortythieves", games.CategorySolo,
		func() usecase.FortyThievesInteractorIF {
			return usecase.NewFortyThievesInteractor(domain.NewDefaultFortyThieves(), new(presenter.FortyThievesWebPresenter))
		},
		func(data []byte) (usecase.FortyThievesInteractorIF, error) {
			return usecase.RestoreFortyThievesInteractor(data, new(presenter.FortyThievesWebPresenter))
		},
		controller.NewFortyThievesWebControllerWithProvider)
	games.RegisterKVGame("fourseasons", games.CategorySolo,
		func() usecase.FourSeasonsInteractorIF {
			return usecase.NewFourSeasonsInteractor(domain.NewDefaultFourSeasons(), new(presenter.FourSeasonsWebPresenter))
		},
		func(data []byte) (usecase.FourSeasonsInteractorIF, error) {
			return usecase.RestoreFourSeasonsInteractor(data, new(presenter.FourSeasonsWebPresenter))
		},
		controller.NewFourSeasonsWebControllerWithProvider)
	games.RegisterKVGame("canfield", games.CategorySolo,
		func() usecase.CanfieldInteractorIF {
			return usecase.NewCanfieldInteractor(domain.NewDefaultCanfield(), new(presenter.CanfieldWebPresenter))
		},
		func(data []byte) (usecase.CanfieldInteractorIF, error) {
			return usecase.RestoreCanfieldInteractor(data, new(presenter.CanfieldWebPresenter))
		},
		controller.NewCanfieldWebControllerWithProvider)
	games.RegisterKVGame("yukon", games.CategorySolo,
		func() usecase.YukonInteractorIF {
			return usecase.NewYukonInteractor(domain.NewDefaultYukon(), new(presenter.YukonWebPresenter))
		},
		func(data []byte) (usecase.YukonInteractorIF, error) {
			return usecase.RestoreYukonInteractor(data, new(presenter.YukonWebPresenter))
		},
		controller.NewYukonWebControllerWithProvider)
	games.RegisterKVGame("russiansolitaire", games.CategorySolo,
		func() usecase.RussianSolitaireInteractorIF {
			return usecase.NewRussianSolitaireInteractor(domain.NewDefaultRussianSolitaire(), new(presenter.RussianSolitaireWebPresenter))
		},
		func(data []byte) (usecase.RussianSolitaireInteractorIF, error) {
			return usecase.RestoreRussianSolitaireInteractor(data, new(presenter.RussianSolitaireWebPresenter))
		},
		controller.NewRussianSolitaireWebControllerWithProvider)
	games.RegisterKVGame("pokersquares", games.CategorySolo,
		func() usecase.PokerSquaresInteractorIF {
			return usecase.NewPokerSquaresInteractor(domain.NewDefaultPokerSquares(), new(presenter.PokerSquaresWebPresenter))
		},
		func(data []byte) (usecase.PokerSquaresInteractorIF, error) {
			return usecase.RestorePokerSquaresInteractor(data, new(presenter.PokerSquaresWebPresenter))
		},
		controller.NewPokerSquaresWebControllerWithProvider)
	games.RegisterKVGame("scorpion", games.CategorySolo,
		func() usecase.ScorpionInteractorIF {
			return usecase.NewScorpionInteractor(domain.NewDefaultScorpion(), new(presenter.ScorpionWebPresenter))
		},
		func(data []byte) (usecase.ScorpionInteractorIF, error) {
			return usecase.RestoreScorpionInteractor(data, new(presenter.ScorpionWebPresenter))
		},
		controller.NewScorpionWebControllerWithProvider)
	games.RegisterKVGame("wasp", games.CategorySolo,
		func() usecase.WaspInteractorIF {
			return usecase.NewWaspInteractor(domain.NewDefaultWasp(), new(presenter.WaspWebPresenter))
		},
		func(data []byte) (usecase.WaspInteractorIF, error) {
			return usecase.RestoreWaspInteractor(data, new(presenter.WaspWebPresenter))
		},
		controller.NewWaspWebControllerWithProvider)
	games.RegisterKVGame("accordion", games.CategorySolo,
		func() usecase.AccordionInteractorIF {
			return usecase.NewAccordionInteractor(domain.NewDefaultAccordion(), new(presenter.AccordionWebPresenter))
		},
		func(data []byte) (usecase.AccordionInteractorIF, error) {
			return usecase.RestoreAccordionInteractor(data, new(presenter.AccordionWebPresenter))
		},
		controller.NewAccordionWebControllerWithProvider)
	games.RegisterKVGame("calculation", games.CategorySolo,
		func() usecase.CalculationInteractorIF {
			return usecase.NewCalculationInteractor(domain.NewDefaultCalculation(), new(presenter.CalculationWebPresenter))
		},
		func(data []byte) (usecase.CalculationInteractorIF, error) {
			return usecase.RestoreCalculationInteractor(data, new(presenter.CalculationWebPresenter))
		},
		controller.NewCalculationWebControllerWithProvider)
	games.RegisterKVGame("bakersdozen", games.CategorySolo,
		func() usecase.BakersDozenInteractorIF {
			return usecase.NewBakersDozenInteractor(domain.NewDefaultBakersDozen(), new(presenter.BakersDozenWebPresenter))
		},
		func(data []byte) (usecase.BakersDozenInteractorIF, error) {
			return usecase.RestoreBakersDozenInteractor(data, new(presenter.BakersDozenWebPresenter))
		},
		controller.NewBakersDozenWebControllerWithProvider)
	games.RegisterKVGame("montecarlo", games.CategorySolo,
		func() usecase.MonteCarloInteractorIF {
			return usecase.NewMonteCarloInteractor(domain.NewDefaultMonteCarlo(), new(presenter.MonteCarloWebPresenter))
		},
		func(data []byte) (usecase.MonteCarloInteractorIF, error) {
			return usecase.RestoreMonteCarloInteractor(data, new(presenter.MonteCarloWebPresenter))
		},
		controller.NewMonteCarloWebControllerWithProvider)
	games.RegisterKVGame("crescent", games.CategorySolo,
		func() usecase.CrescentInteractorIF {
			return usecase.NewCrescentInteractor(domain.NewDefaultCrescent(), new(presenter.CrescentWebPresenter))
		},
		func(data []byte) (usecase.CrescentInteractorIF, error) {
			return usecase.RestoreCrescentInteractor(data, new(presenter.CrescentWebPresenter))
		},
		controller.NewCrescentWebControllerWithProvider)
	games.RegisterKVGame("spiderette", games.CategorySolo,
		func() usecase.SpideretteInteractorIF {
			return usecase.NewSpideretteInteractor(domain.NewDefaultSpiderette(), new(presenter.SpideretteWebPresenter))
		},
		func(data []byte) (usecase.SpideretteInteractorIF, error) {
			return usecase.RestoreSpideretteInteractor(data, new(presenter.SpideretteWebPresenter))
		},
		controller.NewSpideretteWebControllerWithProvider)
	games.RegisterKVGame("fortress", games.CategorySolo,
		func() usecase.FortressInteractorIF {
			return usecase.NewFortressInteractor(domain.NewDefaultFortress(), new(presenter.FortressWebPresenter))
		},
		func(data []byte) (usecase.FortressInteractorIF, error) {
			return usecase.RestoreFortressInteractor(data, new(presenter.FortressWebPresenter))
		},
		controller.NewFortressWebControllerWithProvider)
	games.RegisterKVGame("beleagueredcastle", games.CategorySolo,
		func() usecase.BeleagueredCastleInteractorIF {
			return usecase.NewBeleagueredCastleInteractor(domain.NewDefaultBeleagueredCastle(), new(presenter.BeleagueredCastleWebPresenter))
		},
		func(data []byte) (usecase.BeleagueredCastleInteractorIF, error) {
			return usecase.RestoreBeleagueredCastleInteractor(data, new(presenter.BeleagueredCastleWebPresenter))
		},
		controller.NewBeleagueredCastleWebControllerWithProvider)
	games.RegisterKVGame("gaps", games.CategorySolo,
		func() usecase.GapsInteractorIF {
			return usecase.NewGapsInteractor(domain.NewDefaultGaps(), new(presenter.GapsWebPresenter))
		},
		func(data []byte) (usecase.GapsInteractorIF, error) {
			return usecase.RestoreGapsInteractor(data, new(presenter.GapsWebPresenter))
		},
		controller.NewGapsWebControllerWithProvider)
	games.RegisterKVGame("eightoff", games.CategorySolo,
		func() usecase.EightOffInteractorIF {
			return usecase.NewEightOffInteractor(domain.NewDefaultEightOff(), new(presenter.EightOffWebPresenter))
		},
		func(data []byte) (usecase.EightOffInteractorIF, error) {
			return usecase.RestoreEightOffInteractor(data, new(presenter.EightOffWebPresenter))
		},
		controller.NewEightOffWebControllerWithProvider)
	games.RegisterKVGame("penguin", games.CategorySolo,
		func() usecase.PenguinInteractorIF {
			return usecase.NewPenguinInteractor(domain.NewDefaultPenguin(), new(presenter.PenguinWebPresenter))
		},
		func(data []byte) (usecase.PenguinInteractorIF, error) {
			return usecase.RestorePenguinInteractor(data, new(presenter.PenguinWebPresenter))
		},
		controller.NewPenguinWebControllerWithProvider)
	games.RegisterKVGame("acesup", games.CategorySolo,
		func() usecase.AcesUpInteractorIF {
			return usecase.NewAcesUpInteractor(domain.NewDefaultAcesUp(), new(presenter.AcesUpWebPresenter))
		},
		func(data []byte) (usecase.AcesUpInteractorIF, error) {
			return usecase.RestoreAcesUpInteractor(data, new(presenter.AcesUpWebPresenter))
		},
		controller.NewAcesUpWebControllerWithProvider)
	// Barbu is bucketed here (solo worker) for binary-size reasons; the classic
	// worker is at the 1 MB gzip free-tier limit. See registry.go.
	// Macau is a Crazy Eights variant bucketed here (solo worker) for binary-size
	// reasons; the classic worker is at the 1 MB gzip free-tier limit. See registry.go.
	games.RegisterKVGame("macau", games.CategorySolo,
		func() usecase.MacauInteractorIF {
			return usecase.NewMacauInteractor(domain.NewDefaultMacau(), new(presenter.MacauWebPresenter))
		},
		func(data []byte) (usecase.MacauInteractorIF, error) {
			return usecase.RestoreMacauInteractor(data, new(presenter.MacauWebPresenter))
		},
		controller.NewMacauWebControllerWithProvider)
	games.RegisterKVGame("thirtyone", games.CategorySolo,
		func() usecase.ThirtyOneInteractorIF {
			return usecase.NewThirtyOneInteractor(domain.NewDefaultThirtyOne(), new(presenter.ThirtyOneWebPresenter))
		},
		func(data []byte) (usecase.ThirtyOneInteractorIF, error) {
			return usecase.RestoreThirtyOneInteractor(data, new(presenter.ThirtyOneWebPresenter))
		},
		controller.NewThirtyOneWebControllerWithProvider)
	games.RegisterKVGame("tienlen", games.CategorySolo,
		func() usecase.TienLenInteractorIF {
			return usecase.NewTienLenInteractor(domain.NewDefaultTienLen(), new(presenter.TienLenWebPresenter))
		},
		func(data []byte) (usecase.TienLenInteractorIF, error) {
			return usecase.RestoreTienLenInteractor(data, new(presenter.TienLenWebPresenter))
		},
		controller.NewTienLenWebControllerWithProvider)
	games.RegisterKVGame("osmosis", games.CategorySolo,
		func() usecase.OsmosisInteractorIF {
			return usecase.NewOsmosisInteractor(domain.NewDefaultOsmosis(), new(presenter.OsmosisWebPresenter))
		},
		func(data []byte) (usecase.OsmosisInteractorIF, error) {
			return usecase.RestoreOsmosisInteractor(data, new(presenter.OsmosisWebPresenter))
		},
		controller.NewOsmosisWebControllerWithProvider)
	games.RegisterKVGame("fivehundred", games.CategorySolo,
		func() usecase.FiveHundredInteractorIF {
			return usecase.NewFiveHundredInteractor(domain.NewDefaultFiveHundred(), new(presenter.FiveHundredWebPresenter))
		},
		func(data []byte) (usecase.FiveHundredInteractorIF, error) {
			return usecase.RestoreFiveHundredInteractor(data, new(presenter.FiveHundredWebPresenter))
		},
		controller.NewFiveHundredWebControllerWithProvider)
	// Schnapsen / Sixty-Six is a 2-player trick-taking game bucketed here (solo
	// worker) for binary-size reasons; the classic worker is at the 1 MB gzip
	// free-tier limit. See registry.go.
	games.RegisterKVGame("schnapsen", games.CategorySolo,
		func() usecase.SchnapsenInteractorIF {
			return usecase.NewSchnapsenInteractor(domain.NewDefaultSchnapsen(), new(presenter.SchnapsenWebPresenter))
		},
		func(data []byte) (usecase.SchnapsenInteractorIF, error) {
			return usecase.RestoreSchnapsenInteractor(data, new(presenter.SchnapsenWebPresenter))
		},
		controller.NewSchnapsenWebControllerWithProvider)
	games.RegisterKVGame("euchre", games.CategorySolo,
		func() usecase.EuchreInteractorIF {
			return usecase.NewEuchreInteractor(domain.NewDefaultEuchre(), new(presenter.EuchreWebPresenter))
		},
		func(data []byte) (usecase.EuchreInteractorIF, error) {
			return usecase.RestoreEuchreInteractor(data, new(presenter.EuchreWebPresenter))
		},
		controller.NewEuchreWebControllerWithProvider)

	games.RegisterKVGame("gongzhu", games.CategorySolo,
		func() usecase.GongZhuInteractorIF {
			return usecase.NewGongZhuInteractor(domain.NewDefaultGongZhu(), new(presenter.GongZhuWebPresenter))
		},
		func(data []byte) (usecase.GongZhuInteractorIF, error) {
			return usecase.RestoreGongZhuInteractor(data, new(presenter.GongZhuWebPresenter))
		},
		controller.NewGongZhuWebControllerWithProvider)

	games.RegisterKVGame("bristol", games.CategorySolo,
		func() usecase.BristolInteractorIF {
			return usecase.NewBristolInteractor(domain.NewDefaultBristol(), new(presenter.BristolWebPresenter))
		},
		func(data []byte) (usecase.BristolInteractorIF, error) {
			return usecase.RestoreBristolInteractor(data, new(presenter.BristolWebPresenter))
		},
		controller.NewBristolWebControllerWithProvider)

	games.RegisterKVGame("bidwhist", games.CategorySolo,
		func() usecase.BidWhistInteractorIF {
			return usecase.NewBidWhistInteractor(domain.NewDefaultBidWhist(), new(presenter.BidWhistWebPresenter))
		},
		func(data []byte) (usecase.BidWhistInteractorIF, error) {
			return usecase.RestoreBidWhistInteractor(data, new(presenter.BidWhistWebPresenter))
		},
		controller.NewBidWhistWebControllerWithProvider)

	games.RegisterKVGame("easthaven", games.CategorySolo,
		func() usecase.EasthavenInteractorIF {
			return usecase.NewEasthavenInteractor(domain.NewDefaultEasthaven(), new(presenter.EasthavenWebPresenter))
		},
		func(data []byte) (usecase.EasthavenInteractorIF, error) {
			return usecase.RestoreEasthavenInteractor(data, new(presenter.EasthavenWebPresenter))
		},
		controller.NewEasthavenWebControllerWithProvider)

	games.RegisterKVGame("blackhole", games.CategorySolo,
		func() usecase.BlackHoleInteractorIF {
			return usecase.NewBlackHoleInteractor(domain.NewDefaultBlackHole(), new(presenter.BlackHoleWebPresenter))
		},
		func(data []byte) (usecase.BlackHoleInteractorIF, error) {
			return usecase.RestoreBlackHoleInteractor(data, new(presenter.BlackHoleWebPresenter))
		},
		controller.NewBlackHoleWebControllerWithProvider)
	// Scarto (78-card Italian tarot trick-taker) is bucketed here for binary-size
	// headroom; the extra worker hit the 1 MB gzip free-tier limit. Category is a
	// size bucket, not a user-facing taxonomy.
	// Cego (54-card Baden tarock, Cego-blind swap) — bucketed in solo (extra full).
	// Zheng Shangyou (54-card Chinese climbing game, suit-blind ranks).
	games.RegisterKVGame("zheng", games.CategorySolo,
		func() usecase.ZhengInteractorIF {
			return usecase.NewZhengInteractor(domain.NewDefaultZheng(), new(presenter.ZhengWebPresenter))
		},
		func(data []byte) (usecase.ZhengInteractorIF, error) {
			return usecase.RestoreZhengInteractor(data, new(presenter.ZhengWebPresenter))
		},
		controller.NewZhengWebControllerWithProvider)
	games.RegisterKVGame("yaniv", games.CategorySolo,
		func() usecase.YanivInteractorIF {
			return usecase.NewYanivInteractor(domain.NewDefaultYaniv(), new(presenter.YanivWebPresenter))
		},
		func(data []byte) (usecase.YanivInteractorIF, error) {
			return usecase.RestoreYanivInteractor(data, new(presenter.YanivWebPresenter))
		},
		controller.NewYanivWebControllerWithProvider)
	games.RegisterKVGame("crazyquilt", games.CategorySolo,
		func() usecase.CrazyQuiltInteractorIF {
			return usecase.NewCrazyQuiltInteractor(domain.NewDefaultCrazyQuilt(), new(presenter.CrazyQuiltWebPresenter))
		},
		func(data []byte) (usecase.CrazyQuiltInteractorIF, error) {
			return usecase.RestoreCrazyQuiltInteractor(data, new(presenter.CrazyQuiltWebPresenter))
		},
		controller.NewCrazyQuiltWebControllerWithProvider)
	games.RegisterKVGame("teendopaanch", games.CategorySolo,
		func() usecase.TeenDoPaanchInteractorIF {
			return usecase.NewTeenDoPaanchInteractor(domain.NewDefaultTeenDoPaanch(), new(presenter.TeenDoPaanchWebPresenter))
		},
		func(data []byte) (usecase.TeenDoPaanchInteractorIF, error) {
			return usecase.RestoreTeenDoPaanchInteractor(data, new(presenter.TeenDoPaanchWebPresenter))
		},
		controller.NewTeenDoPaanchWebControllerWithProvider)
	games.RegisterKVGame("snap", games.CategorySolo,
		func() usecase.SnapInteractorIF {
			return usecase.NewSnapInteractor(domain.NewDefaultSnap(), new(presenter.SnapWebPresenter))
		},
		func(data []byte) (usecase.SnapInteractorIF, error) {
			return usecase.RestoreSnapInteractor(data, new(presenter.SnapWebPresenter))
		},
		controller.NewSnapWebControllerWithProvider)
	games.RegisterKVGame("tusac", games.CategorySolo,
		func() usecase.TuSacInteractorIF {
			return usecase.NewTuSacInteractor(domain.NewDefaultTuSac(), new(presenter.TuSacWebPresenter))
		},
		func(data []byte) (usecase.TuSacInteractorIF, error) {
			return usecase.RestoreTuSacInteractor(data, new(presenter.TuSacWebPresenter))
		},
		controller.NewTuSacWebControllerWithProvider)
}
