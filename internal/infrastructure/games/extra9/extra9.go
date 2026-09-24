//go:build js && wasm

// Package extra9 reserves the twelfth Cloudflare Worker size bucket.
package extra9

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func init() {
	games.RegisterKVGame("blackhole", games.CategoryExtra9,
		func() usecase.BlackHoleInteractorIF {
			return usecase.NewBlackHoleInteractor(domain.NewDefaultBlackHole(), new(presenter.BlackHoleWebPresenter))
		},
		func(data []byte) (usecase.BlackHoleInteractorIF, error) {
			return usecase.RestoreBlackHoleInteractor(data, new(presenter.BlackHoleWebPresenter))
		},
		controller.NewBlackHoleWebControllerWithProvider)
	games.RegisterKVGame("clocksolitaire", games.CategoryExtra9,
		func() usecase.ClockSolitaireInteractorIF {
			return usecase.NewClockSolitaireInteractor(domain.NewDefaultClockSolitaire(), new(presenter.ClockSolitaireWebPresenter))
		},
		func(data []byte) (usecase.ClockSolitaireInteractorIF, error) {
			return usecase.RestoreClockSolitaireInteractor(data, new(presenter.ClockSolitaireWebPresenter))
		},
		controller.NewClockSolitaireWebControllerWithProvider)
	games.RegisterKVGame("memory", games.CategoryExtra9,
		func() usecase.MemoryInteractorIF {
			return usecase.NewMemoryInteractor(domain.NewDefaultMemory(), new(presenter.MemoryWebPresenter))
		},
		func(data []byte) (usecase.MemoryInteractorIF, error) {
			return usecase.RestoreMemoryInteractor(data, new(presenter.MemoryWebPresenter))
		},
		controller.NewMemoryWebControllerWithProvider)
	games.RegisterKVGame("montecarlo", games.CategoryExtra9,
		func() usecase.MonteCarloInteractorIF {
			return usecase.NewMonteCarloInteractor(domain.NewDefaultMonteCarlo(), new(presenter.MonteCarloWebPresenter))
		},
		func(data []byte) (usecase.MonteCarloInteractorIF, error) {
			return usecase.RestoreMonteCarloInteractor(data, new(presenter.MonteCarloWebPresenter))
		},
		controller.NewMonteCarloWebControllerWithProvider)
	games.RegisterKVGame("penguin", games.CategoryExtra9,
		func() usecase.PenguinInteractorIF {
			return usecase.NewPenguinInteractor(domain.NewDefaultPenguin(), new(presenter.PenguinWebPresenter))
		},
		func(data []byte) (usecase.PenguinInteractorIF, error) {
			return usecase.RestorePenguinInteractor(data, new(presenter.PenguinWebPresenter))
		},
		controller.NewPenguinWebControllerWithProvider)
	games.RegisterKVGame("pokersquares", games.CategoryExtra9,
		func() usecase.PokerSquaresInteractorIF {
			return usecase.NewPokerSquaresInteractor(domain.NewDefaultPokerSquares(), new(presenter.PokerSquaresWebPresenter))
		},
		func(data []byte) (usecase.PokerSquaresInteractorIF, error) {
			return usecase.RestorePokerSquaresInteractor(data, new(presenter.PokerSquaresWebPresenter))
		},
		controller.NewPokerSquaresWebControllerWithProvider)
	games.RegisterKVGame("schnapsen", games.CategoryExtra9,
		func() usecase.SchnapsenInteractorIF {
			return usecase.NewSchnapsenInteractor(domain.NewDefaultSchnapsen(), new(presenter.SchnapsenWebPresenter))
		},
		func(data []byte) (usecase.SchnapsenInteractorIF, error) {
			return usecase.RestoreSchnapsenInteractor(data, new(presenter.SchnapsenWebPresenter))
		},
		controller.NewSchnapsenWebControllerWithProvider)
	games.RegisterKVGame("seahaventowers", games.CategoryExtra9,
		func() usecase.SeahavenTowersInteractorIF {
			return usecase.NewSeahavenTowersInteractor(domain.NewDefaultSeahavenTowers(), new(presenter.SeahavenTowersWebPresenter))
		},
		func(data []byte) (usecase.SeahavenTowersInteractorIF, error) {
			return usecase.RestoreSeahavenTowersInteractor(data, new(presenter.SeahavenTowersWebPresenter))
		},
		controller.NewSeahavenTowersWebControllerWithProvider)
	games.RegisterKVGame("snap", games.CategoryExtra9,
		func() usecase.SnapInteractorIF {
			return usecase.NewSnapInteractor(domain.NewDefaultSnap(), new(presenter.SnapWebPresenter))
		},
		func(data []byte) (usecase.SnapInteractorIF, error) {
			return usecase.RestoreSnapInteractor(data, new(presenter.SnapWebPresenter))
		},
		controller.NewSnapWebControllerWithProvider)
	games.RegisterKVGame("stalactites", games.CategoryExtra9,
		func() usecase.StalactitesInteractorIF {
			return usecase.NewStalactitesInteractor(domain.NewDefaultStalactites(), new(presenter.StalactitesWebPresenter))
		},
		func(data []byte) (usecase.StalactitesInteractorIF, error) {
			return usecase.RestoreStalactitesInteractor(data, new(presenter.StalactitesWebPresenter))
		},
		controller.NewStalactitesWebControllerWithProvider)
	games.RegisterKVGame("tienlen", games.CategoryExtra9,
		func() usecase.TienLenInteractorIF {
			return usecase.NewTienLenInteractor(domain.NewDefaultTienLen(), new(presenter.TienLenWebPresenter))
		},
		func(data []byte) (usecase.TienLenInteractorIF, error) {
			return usecase.RestoreTienLenInteractor(data, new(presenter.TienLenWebPresenter))
		},
		controller.NewTienLenWebControllerWithProvider)
	games.RegisterKVGame("yaniv", games.CategoryExtra9,
		func() usecase.YanivInteractorIF {
			return usecase.NewYanivInteractor(domain.NewDefaultYaniv(), new(presenter.YanivWebPresenter))
		},
		func(data []byte) (usecase.YanivInteractorIF, error) {
			return usecase.RestoreYanivInteractor(data, new(presenter.YanivWebPresenter))
		},
		controller.NewYanivWebControllerWithProvider)
	games.RegisterKVGame("zheng", games.CategoryExtra9,
		func() usecase.ZhengInteractorIF {
			return usecase.NewZhengInteractor(domain.NewDefaultZheng(), new(presenter.ZhengWebPresenter))
		},
		func(data []byte) (usecase.ZhengInteractorIF, error) {
			return usecase.RestoreZhengInteractor(data, new(presenter.ZhengWebPresenter))
		},
		controller.NewZhengWebControllerWithProvider)
	games.RegisterKVGame("bauernschnapsen", games.CategoryExtra9,
		func() usecase.BauernschnapsenInteractorIF {
			return usecase.NewBauernschnapsenInteractor(domain.NewDefaultBauernschnapsen(), new(presenter.BauernschnapsenWebPresenter))
		},
		func(data []byte) (usecase.BauernschnapsenInteractorIF, error) {
			return usecase.RestoreBauernschnapsenInteractor(data, new(presenter.BauernschnapsenWebPresenter))
		},
		controller.NewBauernschnapsenWebControllerWithProvider)
	games.RegisterKVGame("calabresella", games.CategoryExtra9,
		func() usecase.CalabresellaInteractorIF {
			return usecase.NewCalabresellaInteractor(domain.NewDefaultCalabresella(), new(presenter.CalabresellaWebPresenter))
		},
		func(data []byte) (usecase.CalabresellaInteractorIF, error) {
			return usecase.RestoreCalabresellaInteractor(data, new(presenter.CalabresellaWebPresenter))
		},
		controller.NewCalabresellaWebControllerWithProvider)
	games.RegisterKVGame("gaigel", games.CategoryExtra9,
		func() usecase.GaigelInteractorIF {
			return usecase.NewGaigelInteractor(domain.NewDefaultGaigel(), new(presenter.GaigelWebPresenter))
		},
		func(data []byte) (usecase.GaigelInteractorIF, error) {
			return usecase.RestoreGaigelInteractor(data, new(presenter.GaigelWebPresenter))
		},
		controller.NewGaigelWebControllerWithProvider)
	games.RegisterKVGame("pontoon", games.CategoryExtra9,
		func() usecase.PontoonInteractorIF {
			return usecase.NewPontoonInteractor(domain.NewDefaultPontoon(), new(presenter.PontoonWebPresenter))
		},
		func(data []byte) (usecase.PontoonInteractorIF, error) {
			return usecase.RestorePontoonInteractor(data, new(presenter.PontoonWebPresenter))
		},
		controller.NewPontoonWebControllerWithProvider)
	games.RegisterKVGame("quinze", games.CategoryExtra9,
		func() usecase.QuinzeInteractorIF {
			return usecase.NewQuinzeInteractor(domain.NewDefaultQuinze(), new(presenter.QuinzeWebPresenter))
		},
		func(data []byte) (usecase.QuinzeInteractorIF, error) {
			return usecase.RestoreQuinzeInteractor(data, new(presenter.QuinzeWebPresenter))
		},
		controller.NewQuinzeWebControllerWithProvider)
	games.RegisterKVGame("settemezzo", games.CategoryExtra9,
		func() usecase.SetteEMezzoInteractorIF {
			return usecase.NewSetteEMezzoInteractor(domain.NewDefaultSetteEMezzo(), new(presenter.SetteEMezzoWebPresenter))
		},
		func(data []byte) (usecase.SetteEMezzoInteractorIF, error) {
			return usecase.RestoreSetteEMezzoInteractor(data, new(presenter.SetteEMezzoWebPresenter))
		},
		controller.NewSetteEMezzoWebControllerWithProvider)
	games.RegisterKVGame("tysiac", games.CategoryExtra9,
		func() usecase.TysiacInteractorIF {
			return usecase.NewTysiacInteractor(domain.NewDefaultTysiac(), new(presenter.TysiacWebPresenter))
		},
		func(data []byte) (usecase.TysiacInteractorIF, error) {
			return usecase.RestoreTysiacInteractor(data, new(presenter.TysiacWebPresenter))
		},
		controller.NewTysiacWebControllerWithProvider)
	games.RegisterKVGame("vira", games.CategoryExtra9,
		func() usecase.ViraInteractorIF {
			return usecase.NewViraInteractor(domain.NewDefaultVira(), new(presenter.ViraWebPresenter))
		},
		func(data []byte) (usecase.ViraInteractorIF, error) {
			return usecase.RestoreViraInteractor(data, new(presenter.ViraWebPresenter))
		},
		controller.NewViraWebControllerWithProvider)
	games.RegisterKVGame("aluette", games.CategoryExtra9,
		func() usecase.AluetteInteractorIF {
			return usecase.NewAluetteInteractor(domain.NewDefaultAluette(), new(presenter.AluetteWebPresenter))
		},
		func(data []byte) (usecase.AluetteInteractorIF, error) {
			return usecase.RestoreAluetteInteractor(data, new(presenter.AluetteWebPresenter))
		},
		controller.NewAluetteWebControllerWithProvider)
	games.RegisterKVGame("baloot", games.CategoryExtra9,
		func() usecase.BalootInteractorIF {
			return usecase.NewBalootInteractor(domain.NewDefaultBaloot(), new(presenter.BalootWebPresenter))
		},
		func(data []byte) (usecase.BalootInteractorIF, error) {
			return usecase.RestoreBalootInteractor(data, new(presenter.BalootWebPresenter))
		},
		controller.NewBalootWebControllerWithProvider)
	games.RegisterKVGame("bideuchre", games.CategoryExtra9,
		func() usecase.BidEuchreInteractorIF {
			return usecase.NewBidEuchreInteractor(domain.NewDefaultBidEuchre(), new(presenter.BidEuchreWebPresenter))
		},
		func(data []byte) (usecase.BidEuchreInteractorIF, error) {
			return usecase.RestoreBidEuchreInteractor(data, new(presenter.BidEuchreWebPresenter))
		},
		controller.NewBidEuchreWebControllerWithProvider)
	games.RegisterKVGame("julepe", games.CategoryExtra9,
		func() usecase.JulepeInteractorIF {
			return usecase.NewJulepeInteractor(domain.NewDefaultJulepe(), new(presenter.JulepeWebPresenter))
		},
		func(data []byte) (usecase.JulepeInteractorIF, error) {
			return usecase.RestoreJulepeInteractor(data, new(presenter.JulepeWebPresenter))
		},
		controller.NewJulepeWebControllerWithProvider)
	games.RegisterKVGame("rikken", games.CategoryExtra9,
		func() usecase.RikkenInteractorIF {
			return usecase.NewRikkenInteractor(domain.NewDefaultRikken(), new(presenter.RikkenWebPresenter))
		},
		func(data []byte) (usecase.RikkenInteractorIF, error) {
			return usecase.RestoreRikkenInteractor(data, new(presenter.RikkenWebPresenter))
		},
		controller.NewRikkenWebControllerWithProvider)
	games.RegisterKVGame("sjavs", games.CategoryExtra9,
		func() usecase.SjavsInteractorIF {
			return usecase.NewSjavsInteractor(domain.NewDefaultSjavs(), new(presenter.SjavsWebPresenter))
		},
		func(data []byte) (usecase.SjavsInteractorIF, error) {
			return usecase.RestoreSjavsInteractor(data, new(presenter.SjavsWebPresenter))
		},
		controller.NewSjavsWebControllerWithProvider)
	games.RegisterKVGame("spiteandmalice", games.CategoryExtra9,
		func() usecase.SpiteAndMaliceInteractorIF {
			return usecase.NewSpiteAndMaliceInteractor(domain.NewDefaultSpiteAndMalice(), new(presenter.SpiteAndMaliceWebPresenter))
		},
		func(data []byte) (usecase.SpiteAndMaliceInteractorIF, error) {
			return usecase.RestoreSpiteAndMaliceInteractor(data, new(presenter.SpiteAndMaliceWebPresenter))
		},
		controller.NewSpiteAndMaliceWebControllerWithProvider)
	games.RegisterKVGame("zwicker", games.CategoryExtra9,
		func() usecase.ZwickerInteractorIF {
			return usecase.NewZwickerInteractor(domain.NewDefaultZwicker(), new(presenter.ZwickerWebPresenter))
		},
		func(data []byte) (usecase.ZwickerInteractorIF, error) {
			return usecase.RestoreZwickerInteractor(data, new(presenter.ZwickerWebPresenter))
		},
		controller.NewZwickerWebControllerWithProvider)
}
