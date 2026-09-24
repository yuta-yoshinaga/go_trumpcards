//go:build js && wasm

// Package extra10 reserves the thirteenth Cloudflare Worker size bucket.
package extra10

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func init() {
	games.RegisterKVGame("bouillotte", games.CategoryExtra10,
		func() usecase.BouillotteInteractorIF {
			return usecase.NewBouillotteInteractor(domain.NewDefaultBouillotte(), new(presenter.BouillotteWebPresenter))
		},
		func(data []byte) (usecase.BouillotteInteractorIF, error) {
			return usecase.RestoreBouillotteInteractor(data, new(presenter.BouillotteWebPresenter))
		},
		controller.NewBouillotteWebControllerWithProvider)
	games.RegisterKVGame("cirulla", games.CategoryExtra10,
		func() usecase.CirullaInteractorIF {
			return usecase.NewCirullaInteractor(domain.NewDefaultCirulla(), new(presenter.CirullaWebPresenter))
		},
		func(data []byte) (usecase.CirullaInteractorIF, error) {
			return usecase.RestoreCirullaInteractor(data, new(presenter.CirullaWebPresenter))
		},
		controller.NewCirullaWebControllerWithProvider)
	games.RegisterKVGame("kaiser", games.CategoryExtra10,
		func() usecase.KaiserInteractorIF {
			return usecase.NewKaiserInteractor(domain.NewDefaultKaiser(), new(presenter.KaiserWebPresenter))
		},
		func(data []byte) (usecase.KaiserInteractorIF, error) {
			return usecase.RestoreKaiserInteractor(data, new(presenter.KaiserWebPresenter))
		},
		controller.NewKaiserWebControllerWithProvider)
	games.RegisterKVGame("kille", games.CategoryExtra10,
		func() usecase.KilleInteractorIF {
			return usecase.NewKilleInteractor(domain.NewDefaultKille(), new(presenter.KilleWebPresenter))
		},
		func(data []byte) (usecase.KilleInteractorIF, error) {
			return usecase.RestoreKilleInteractor(data, new(presenter.KilleWebPresenter))
		},
		controller.NewKilleWebControllerWithProvider)
	games.RegisterKVGame("mao", games.CategoryExtra10,
		func() usecase.MaoInteractorIF {
			return usecase.NewMaoInteractor(domain.NewDefaultMao(), new(presenter.MaoWebPresenter))
		},
		func(data []byte) (usecase.MaoInteractorIF, error) {
			return usecase.RestoreMaoInteractor(data, new(presenter.MaoWebPresenter))
		},
		controller.NewMaoWebControllerWithProvider)
	games.RegisterKVGame("poch", games.CategoryExtra10,
		func() usecase.PochInteractorIF {
			return usecase.NewPochInteractor(domain.NewDefaultPoch(), new(presenter.PochWebPresenter))
		},
		func(data []byte) (usecase.PochInteractorIF, error) {
			return usecase.RestorePochInteractor(data, new(presenter.PochWebPresenter))
		},
		controller.NewPochWebControllerWithProvider)
	games.RegisterKVGame("primero", games.CategoryExtra10,
		func() usecase.PrimeroInteractorIF {
			return usecase.NewPrimeroInteractor(domain.NewDefaultPrimero(), new(presenter.PrimeroWebPresenter))
		},
		func(data []byte) (usecase.PrimeroInteractorIF, error) {
			return usecase.RestorePrimeroInteractor(data, new(presenter.PrimeroWebPresenter))
		},
		controller.NewPrimeroWebControllerWithProvider)
	games.RegisterKVGame("rook", games.CategoryExtra10,
		func() usecase.RookInteractorIF {
			return usecase.NewRookInteractor(domain.NewDefaultRook(), new(presenter.RookWebPresenter))
		},
		func(data []byte) (usecase.RookInteractorIF, error) {
			return usecase.RestoreRookInteractor(data, new(presenter.RookWebPresenter))
		},
		controller.NewRookWebControllerWithProvider)
	games.RegisterKVGame("cego", games.CategoryExtra10,
		func() usecase.CegoInteractorIF {
			return usecase.NewCegoInteractor(domain.NewDefaultCego(), new(presenter.CegoWebPresenter))
		},
		func(data []byte) (usecase.CegoInteractorIF, error) {
			return usecase.RestoreCegoInteractor(data, new(presenter.CegoWebPresenter))
		},
		controller.NewCegoWebControllerWithProvider)
	games.RegisterKVGame("kingo", games.CategoryExtra10,
		func() usecase.KingoInteractorIF {
			return usecase.NewKingoInteractor(domain.NewDefaultKingo(), new(presenter.KingoWebPresenter))
		},
		func(data []byte) (usecase.KingoInteractorIF, error) {
			return usecase.RestoreKingoInteractor(data, new(presenter.KingoWebPresenter))
		},
		controller.NewKingoWebControllerWithProvider)
	games.RegisterKVGame("napoleon", games.CategoryExtra10,
		func() usecase.NapoleonInteractorIF {
			return usecase.NewNapoleonInteractor(domain.NewDefaultNapoleon(), new(presenter.NapoleonWebPresenter))
		},
		func(data []byte) (usecase.NapoleonInteractorIF, error) {
			return usecase.RestoreNapoleonInteractor(data, new(presenter.NapoleonWebPresenter))
		},
		controller.NewNapoleonWebControllerWithProvider)
	games.RegisterKVGame("oichokabu", games.CategoryExtra10,
		func() usecase.OichoKabuInteractorIF {
			return usecase.NewOichoKabuInteractor(domain.NewDefaultOichoKabu(), new(presenter.OichoKabuWebPresenter))
		},
		func(data []byte) (usecase.OichoKabuInteractorIF, error) {
			return usecase.RestoreOichoKabuInteractor(data, new(presenter.OichoKabuWebPresenter))
		},
		controller.NewOichoKabuWebControllerWithProvider)
	games.RegisterKVGame("quadrille", games.CategoryExtra10,
		func() usecase.QuadrilleInteractorIF {
			return usecase.NewQuadrilleInteractor(domain.NewDefaultQuadrille(), new(presenter.QuadrilleWebPresenter))
		},
		func(data []byte) (usecase.QuadrilleInteractorIF, error) {
			return usecase.RestoreQuadrilleInteractor(data, new(presenter.QuadrilleWebPresenter))
		},
		controller.NewQuadrilleWebControllerWithProvider)
	games.RegisterKVGame("quodlibet", games.CategoryExtra10,
		func() usecase.QuodlibetInteractorIF {
			return usecase.NewQuodlibetInteractor(domain.NewDefaultQuodlibet(), new(presenter.QuodlibetWebPresenter))
		},
		func(data []byte) (usecase.QuodlibetInteractorIF, error) {
			return usecase.RestoreQuodlibetInteractor(data, new(presenter.QuodlibetWebPresenter))
		},
		controller.NewQuodlibetWebControllerWithProvider)
	games.RegisterKVGame("tusac", games.CategoryExtra10,
		func() usecase.TuSacInteractorIF {
			return usecase.NewTuSacInteractor(domain.NewDefaultTuSac(), new(presenter.TuSacWebPresenter))
		},
		func(data []byte) (usecase.TuSacInteractorIF, error) {
			return usecase.RestoreTuSacInteractor(data, new(presenter.TuSacWebPresenter))
		},
		controller.NewTuSacWebControllerWithProvider)
	games.RegisterKVGame("gostop", games.CategoryExtra10,
		func() usecase.GoStopInteractorIF {
			return usecase.NewGoStopInteractor(domain.NewDefaultGoStop(), new(presenter.GoStopWebPresenter))
		},
		func(data []byte) (usecase.GoStopInteractorIF, error) {
			return usecase.RestoreGoStopInteractor(data, new(presenter.GoStopWebPresenter))
		},
		controller.NewGoStopWebControllerWithProvider)
	games.RegisterKVGame("kingalbert", games.CategoryExtra10,
		func() usecase.KingAlbertInteractorIF {
			return usecase.NewKingAlbertInteractor(domain.NewDefaultKingAlbert(), new(presenter.KingAlbertWebPresenter))
		},
		func(data []byte) (usecase.KingAlbertInteractorIF, error) {
			return usecase.RestoreKingAlbertInteractor(data, new(presenter.KingAlbertWebPresenter))
		},
		controller.NewKingAlbertWebControllerWithProvider)
	games.RegisterKVGame("baccarat", games.CategoryExtra10,
		func() usecase.BaccaratInteractorIF {
			return usecase.NewBaccaratInteractor(domain.NewDefaultBaccarat(), new(presenter.BaccaratWebPresenter))
		},
		func(data []byte) (usecase.BaccaratInteractorIF, error) {
			return usecase.RestoreBaccaratInteractor(data, new(presenter.BaccaratWebPresenter))
		},
		controller.NewBaccaratWebControllerWithProvider)
	games.RegisterKVGame("caribbeanstud", games.CategoryExtra10,
		func() usecase.CaribbeanStudInteractorIF {
			return usecase.NewCaribbeanStudInteractor(domain.NewDefaultCaribbeanStud(), new(presenter.CaribbeanStudWebPresenter))
		},
		func(data []byte) (usecase.CaribbeanStudInteractorIF, error) {
			return usecase.RestoreCaribbeanStudInteractor(data, new(presenter.CaribbeanStudWebPresenter))
		},
		controller.NewCaribbeanStudWebControllerWithProvider)
}
