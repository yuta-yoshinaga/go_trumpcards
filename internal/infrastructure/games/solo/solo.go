//go:build js && wasm

// Package solo binds the Cloudflare Worker KV-backed handlers for the
// 17 solitaire and rummy variants. A worker main must blank-import this
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
	games.RegisterKVGame("ginrummy", games.CategorySolo,
		func() usecase.GinRummyInteractorIF {
			return usecase.NewGinRummyInteractor(domain.NewDefaultGinRummy(), new(presenter.GinRummyWebPresenter))
		},
		func(data []byte) (usecase.GinRummyInteractorIF, error) {
			return usecase.RestoreGinRummyInteractor(data, new(presenter.GinRummyWebPresenter))
		},
		controller.NewGinRummyWebControllerWithProvider)
	games.RegisterKVGame("canasta", games.CategorySolo,
		func() usecase.CanastaInteractorIF {
			return usecase.NewCanastaInteractor(domain.NewDefaultCanasta(), new(presenter.CanastaWebPresenter))
		},
		func(data []byte) (usecase.CanastaInteractorIF, error) {
			return usecase.RestoreCanastaInteractor(data, new(presenter.CanastaWebPresenter))
		},
		controller.NewCanastaWebControllerWithProvider)
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
	games.RegisterKVGame("cribbage", games.CategorySolo,
		func() usecase.CribbageInteractorIF {
			return usecase.NewCribbageInteractor(domain.NewDefaultCribbage(), new(presenter.CribbageWebPresenter))
		},
		func(data []byte) (usecase.CribbageInteractorIF, error) {
			return usecase.RestoreCribbageInteractor(data, new(presenter.CribbageWebPresenter))
		},
		controller.NewCribbageWebControllerWithProvider)
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
	games.RegisterKVGame("accordion", games.CategorySolo,
		func() usecase.AccordionInteractorIF {
			return usecase.NewAccordionInteractor(domain.NewDefaultAccordion(), new(presenter.AccordionWebPresenter))
		},
		func(data []byte) (usecase.AccordionInteractorIF, error) {
			return usecase.RestoreAccordionInteractor(data, new(presenter.AccordionWebPresenter))
		},
		controller.NewAccordionWebControllerWithProvider)
	games.RegisterKVGame("sevenbridge", games.CategorySolo,
		func() usecase.SevenBridgeInteractorIF {
			return usecase.NewSevenBridgeInteractor(domain.NewDefaultSevenBridge(), new(presenter.SevenBridgeWebPresenter))
		},
		func(data []byte) (usecase.SevenBridgeInteractorIF, error) {
			return usecase.RestoreSevenBridgeInteractor(data, new(presenter.SevenBridgeWebPresenter))
		},
		controller.NewSevenBridgeWebControllerWithProvider)
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
	games.RegisterKVGame("contractrummy", games.CategorySolo,
		func() usecase.ContractRummyInteractorIF {
			return usecase.NewContractRummyInteractor(domain.NewDefaultContractRummy(), new(presenter.ContractRummyWebPresenter))
		},
		func(data []byte) (usecase.ContractRummyInteractorIF, error) {
			return usecase.RestoreContractRummyInteractor(data, new(presenter.ContractRummyWebPresenter))
		},
		controller.NewContractRummyWebControllerWithProvider)
	games.RegisterKVGame("crescent", games.CategorySolo,
		func() usecase.CrescentInteractorIF {
			return usecase.NewCrescentInteractor(domain.NewDefaultCrescent(), new(presenter.CrescentWebPresenter))
		},
		func(data []byte) (usecase.CrescentInteractorIF, error) {
			return usecase.RestoreCrescentInteractor(data, new(presenter.CrescentWebPresenter))
		},
		controller.NewCrescentWebControllerWithProvider)
}
