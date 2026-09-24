//go:build js && wasm

// Package extra11 reserves the fourteenth Cloudflare Worker size bucket.
package extra11

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func init() {
	games.RegisterKVGame("chineseten", games.CategoryExtra11,
		func() usecase.ChineseTenInteractorIF {
			return usecase.NewChineseTenInteractor(domain.NewDefaultChineseTen(), new(presenter.ChineseTenWebPresenter))
		},
		func(data []byte) (usecase.ChineseTenInteractorIF, error) {
			return usecase.RestoreChineseTenInteractor(data, new(presenter.ChineseTenWebPresenter))
		},
		controller.NewChineseTenWebControllerWithProvider)
	games.RegisterKVGame("cuckoo", games.CategoryExtra11,
		func() usecase.CuckooInteractorIF {
			return usecase.NewCuckooInteractor(domain.NewDefaultCuckoo(), new(presenter.CuckooWebPresenter))
		},
		func(data []byte) (usecase.CuckooInteractorIF, error) {
			return usecase.RestoreCuckooInteractor(data, new(presenter.CuckooWebPresenter))
		},
		controller.NewCuckooWebControllerWithProvider)
	games.RegisterKVGame("daifugo", games.CategoryExtra11,
		func() usecase.DaifugoInteractorIF {
			return usecase.NewDaifugoInteractor(domain.NewDefaultDaifugo(), new(presenter.DaifugoWebPresenter))
		},
		func(data []byte) (usecase.DaifugoInteractorIF, error) {
			return usecase.RestoreDaifugoInteractor(data, new(presenter.DaifugoWebPresenter))
		},
		controller.NewDaifugoWebControllerWithProvider)
	games.RegisterKVGame("fortyfives", games.CategoryExtra11,
		func() usecase.FortyFivesInteractorIF {
			return usecase.NewFortyFivesInteractor(domain.NewDefaultFortyFives(), new(presenter.FortyFivesWebPresenter))
		},
		func(data []byte) (usecase.FortyFivesInteractorIF, error) {
			return usecase.RestoreFortyFivesInteractor(data, new(presenter.FortyFivesWebPresenter))
		},
		controller.NewFortyFivesWebControllerWithProvider)
	games.RegisterKVGame("marjapussi", games.CategoryExtra11,
		func() usecase.MarjapussiInteractorIF {
			return usecase.NewMarjapussiInteractor(domain.NewDefaultMarjapussi(), new(presenter.MarjapussiWebPresenter))
		},
		func(data []byte) (usecase.MarjapussiInteractorIF, error) {
			return usecase.RestoreMarjapussiInteractor(data, new(presenter.MarjapussiWebPresenter))
		},
		controller.NewMarjapussiWebControllerWithProvider)
	games.RegisterKVGame("marriage", games.CategoryExtra11,
		func() usecase.MarriageInteractorIF {
			return usecase.NewMarriageInteractor(domain.NewDefaultMarriage(), new(presenter.MarriageWebPresenter))
		},
		func(data []byte) (usecase.MarriageInteractorIF, error) {
			return usecase.RestoreMarriageInteractor(data, new(presenter.MarriageWebPresenter))
		},
		controller.NewMarriageWebControllerWithProvider)
	games.RegisterKVGame("montebank", games.CategoryExtra11,
		func() usecase.MonteBankInteractorIF {
			return usecase.NewMonteBankInteractor(domain.NewDefaultMonteBank(), new(presenter.MonteBankWebPresenter))
		},
		func(data []byte) (usecase.MonteBankInteractorIF, error) {
			return usecase.RestoreMonteBankInteractor(data, new(presenter.MonteBankWebPresenter))
		},
		controller.NewMonteBankWebControllerWithProvider)
	games.RegisterKVGame("mushi", games.CategoryExtra11,
		func() usecase.MushiInteractorIF {
			return usecase.NewMushiInteractor(domain.NewDefaultMushi(), new(presenter.MushiWebPresenter))
		},
		func(data []byte) (usecase.MushiInteractorIF, error) {
			return usecase.RestoreMushiInteractor(data, new(presenter.MushiWebPresenter))
		},
		controller.NewMushiWebControllerWithProvider)
	games.RegisterKVGame("tute", games.CategoryExtra11,
		func() usecase.TuteInteractorIF {
			return usecase.NewTuteInteractor(domain.NewDefaultTute(), new(presenter.TuteWebPresenter))
		},
		func(data []byte) (usecase.TuteInteractorIF, error) {
			return usecase.RestoreTuteInteractor(data, new(presenter.TuteWebPresenter))
		},
		controller.NewTuteWebControllerWithProvider)
	games.RegisterKVGame("twentynine", games.CategoryExtra11,
		func() usecase.TwentyNineInteractorIF {
			return usecase.NewTwentyNineInteractor(domain.NewDefaultTwentyNine(), new(presenter.TwentyNineWebPresenter))
		},
		func(data []byte) (usecase.TwentyNineInteractorIF, error) {
			return usecase.RestoreTwentyNineInteractor(data, new(presenter.TwentyNineWebPresenter))
		},
		controller.NewTwentyNineWebControllerWithProvider)
	games.RegisterKVGame("bidwhist", games.CategoryExtra11,
		func() usecase.BidWhistInteractorIF {
			return usecase.NewBidWhistInteractor(domain.NewDefaultBidWhist(), new(presenter.BidWhistWebPresenter))
		},
		func(data []byte) (usecase.BidWhistInteractorIF, error) {
			return usecase.RestoreBidWhistInteractor(data, new(presenter.BidWhistWebPresenter))
		},
		controller.NewBidWhistWebControllerWithProvider)
	games.RegisterKVGame("casinowar", games.CategoryExtra11,
		func() usecase.CasinoWarInteractorIF {
			return usecase.NewCasinoWarInteractor(domain.NewDefaultCasinoWar(), new(presenter.CasinoWarWebPresenter))
		},
		func(data []byte) (usecase.CasinoWarInteractorIF, error) {
			return usecase.RestoreCasinoWarInteractor(data, new(presenter.CasinoWarWebPresenter))
		},
		controller.NewCasinoWarWebControllerWithProvider)
	games.RegisterKVGame("germansolo", games.CategoryExtra11,
		func() usecase.GermanSoloInteractorIF {
			return usecase.NewGermanSoloInteractor(domain.NewDefaultGermanSolo(), new(presenter.GermanSoloWebPresenter))
		},
		func(data []byte) (usecase.GermanSoloInteractorIF, error) {
			return usecase.RestoreGermanSoloInteractor(data, new(presenter.GermanSoloWebPresenter))
		},
		controller.NewGermanSoloWebControllerWithProvider)
	games.RegisterKVGame("guandan", games.CategoryExtra11,
		func() usecase.GuandanInteractorIF {
			return usecase.NewGuandanInteractor(domain.NewDefaultGuandan(), new(presenter.GuandanWebPresenter))
		},
		func(data []byte) (usecase.GuandanInteractorIF, error) {
			return usecase.RestoreGuandanInteractor(data, new(presenter.GuandanWebPresenter))
		},
		controller.NewGuandanWebControllerWithProvider)
	games.RegisterKVGame("guts", games.CategoryExtra11,
		func() usecase.GutsInteractorIF {
			return usecase.NewGutsInteractor(domain.NewDefaultGuts(), new(presenter.GutsWebPresenter))
		},
		func(data []byte) (usecase.GutsInteractorIF, error) {
			return usecase.RestoreGutsInteractor(data, new(presenter.GutsWebPresenter))
		},
		controller.NewGutsWebControllerWithProvider)
	games.RegisterKVGame("reddog", games.CategoryExtra11,
		func() usecase.RedDogInteractorIF {
			return usecase.NewRedDogInteractor(domain.NewDefaultRedDog(), new(presenter.RedDogWebPresenter))
		},
		func(data []byte) (usecase.RedDogInteractorIF, error) {
			return usecase.RestoreRedDogInteractor(data, new(presenter.RedDogWebPresenter))
		},
		controller.NewRedDogWebControllerWithProvider)
	games.RegisterKVGame("seventwentyseven", games.CategoryExtra11,
		func() usecase.SevenTwentySevenInteractorIF {
			return usecase.NewSevenTwentySevenInteractor(domain.NewDefaultSevenTwentySeven(), new(presenter.SevenTwentySevenWebPresenter))
		},
		func(data []byte) (usecase.SevenTwentySevenInteractorIF, error) {
			return usecase.RestoreSevenTwentySevenInteractor(data, new(presenter.SevenTwentySevenWebPresenter))
		},
		controller.NewSevenTwentySevenWebControllerWithProvider)
	games.RegisterKVGame("sheepshead", games.CategoryExtra11,
		func() usecase.SheepsheadInteractorIF {
			return usecase.NewSheepsheadInteractor(domain.NewDefaultSheepshead(), new(presenter.SheepsheadWebPresenter))
		},
		func(data []byte) (usecase.SheepsheadInteractorIF, error) {
			return usecase.RestoreSheepsheadInteractor(data, new(presenter.SheepsheadWebPresenter))
		},
		controller.NewSheepsheadWebControllerWithProvider)
}
