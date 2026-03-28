//go:build js && wasm

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	corsmw "github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/cors"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func main() {
	mux := http.NewServeMux()

	// Hearts (KV-backed session)
	heartsFactory := func() usecase.HeartsInteractorIF {
		config := domain.DefaultHeartsConfig()
		players := []*domain.HeartsPlayer{
			domain.NewHeartsPlayer(true),
			domain.NewHeartsPlayer(false),
			domain.NewHeartsPlayer(false),
			domain.NewHeartsPlayer(false),
		}
		hearts := domain.NewHearts(domain.NewTrumpCards(0), players, config)
		return usecase.NewHeartsInteractor(hearts, new(presenter.HeartsWebPresenter))
	}
	heartsKV, err := controller.NewKVSessionProvider[usecase.HeartsInteractorIF](
		"GAME_SESSIONS", "hearts:",
		func(hi usecase.HeartsInteractorIF) ([]byte, error) {
			snap, ok := hi.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.HeartsInteractorIF, error) {
			return usecase.RestoreHeartsInteractor(data, new(presenter.HeartsWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	htc := controller.NewHeartsWebControllerWithProvider(heartsKV, heartsFactory)
	mux.HandleFunc("/hearts/exec", htc.Exec)

	// Spades (KV-backed session)
	spadesFactory := func() usecase.SpadesInteractorIF {
		config := domain.DefaultSpadesConfig()
		players := []*domain.SpadesPlayer{
			domain.NewSpadesPlayer(true),
			domain.NewSpadesPlayer(false),
			domain.NewSpadesPlayer(false),
			domain.NewSpadesPlayer(false),
		}
		spades := domain.NewSpades(domain.NewTrumpCards(0), players, config)
		return usecase.NewSpadesInteractor(spades, new(presenter.SpadesWebPresenter))
	}
	spadesKV, err := controller.NewKVSessionProvider[usecase.SpadesInteractorIF](
		"GAME_SESSIONS", "spades:",
		func(si usecase.SpadesInteractorIF) ([]byte, error) {
			snap, ok := si.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.SpadesInteractorIF, error) {
			return usecase.RestoreSpadesInteractor(data, new(presenter.SpadesWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	spc := controller.NewSpadesWebControllerWithProvider(spadesKV, spadesFactory)
	mux.HandleFunc("/spades/exec", spc.Exec)

	// Euchre (KV-backed session)
	euchreFactory := func() usecase.EuchreInteractorIF {
		config := domain.DefaultEuchreConfig()
		players := []*domain.EuchrePlayer{
			domain.NewEuchrePlayer(true, 0),
			domain.NewEuchrePlayer(false, 1),
			domain.NewEuchrePlayer(false, 0),
			domain.NewEuchrePlayer(false, 1),
		}
		euchre := domain.NewEuchre(domain.NewTrumpCardsEuchre(), players, config)
		return usecase.NewEuchreInteractor(euchre, new(presenter.EuchreWebPresenter))
	}
	euchreKV, err := controller.NewKVSessionProvider[usecase.EuchreInteractorIF](
		"GAME_SESSIONS", "euchre:",
		func(ei usecase.EuchreInteractorIF) ([]byte, error) {
			snap, ok := ei.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.EuchreInteractorIF, error) {
			return usecase.RestoreEuchreInteractor(data, new(presenter.EuchreWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	euc := controller.NewEuchreWebControllerWithProvider(euchreKV, euchreFactory)
	mux.HandleFunc("/euchre/exec", euc.Exec)

	// Napoleon (KV-backed session)
	napoleonFactory := func() usecase.NapoleonInteractorIF {
		config := domain.DefaultNapoleonConfig()
		players := []*domain.NapoleonPlayer{
			domain.NewNapoleonPlayer(true),
			domain.NewNapoleonPlayer(false),
			domain.NewNapoleonPlayer(false),
			domain.NewNapoleonPlayer(false),
		}
		napoleon := domain.NewNapoleon(domain.NewTrumpCards(1), players, config)
		return usecase.NewNapoleonInteractor(napoleon, new(presenter.NapoleonWebPresenter))
	}
	napoleonKV, err := controller.NewKVSessionProvider[usecase.NapoleonInteractorIF](
		"GAME_SESSIONS", "napoleon:",
		func(ni usecase.NapoleonInteractorIF) ([]byte, error) {
			snap, ok := ni.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.NapoleonInteractorIF, error) {
			return usecase.RestoreNapoleonInteractor(data, new(presenter.NapoleonWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	npc := controller.NewNapoleonWebControllerWithProvider(napoleonKV, napoleonFactory)
	mux.HandleFunc("/napoleon/exec", npc.Exec)

	// Old Maid (KV-backed session)
	oldmaidFactory := func() usecase.OldMaidInteractorIF {
		players := []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(true),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}
		oldMaid := domain.NewOldMaid(domain.NewTrumpCards(1), players)
		return usecase.NewOldMaidInteractor(oldMaid, new(presenter.OldMaidWebPresenter))
	}
	oldmaidKV, err := controller.NewKVSessionProvider[usecase.OldMaidInteractorIF](
		"GAME_SESSIONS", "oldmaid:",
		func(oi usecase.OldMaidInteractorIF) ([]byte, error) {
			snap, ok := oi.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.OldMaidInteractorIF, error) {
			return usecase.RestoreOldMaidInteractor(data, new(presenter.OldMaidWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	omc := controller.NewOldMaidWebControllerWithProvider(oldmaidKV, oldmaidFactory)
	mux.HandleFunc("/oldmaid/exec", omc.Exec)

	// Doubt (KV-backed session)
	doubtFactory := func() usecase.DoubtInteractorIF {
		players := []*domain.DoubtPlayer{
			domain.NewDoubtPlayer(true),
			domain.NewDoubtPlayer(false),
			domain.NewDoubtPlayer(false),
			domain.NewDoubtPlayer(false),
		}
		doubt := domain.NewDoubt(domain.NewTrumpCards(0), players)
		return usecase.NewDoubtInteractor(doubt, new(presenter.DoubtWebPresenter))
	}
	doubtKV, err := controller.NewKVSessionProvider[usecase.DoubtInteractorIF](
		"GAME_SESSIONS", "doubt:",
		func(di usecase.DoubtInteractorIF) ([]byte, error) {
			snap, ok := di.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.DoubtInteractorIF, error) {
			return usecase.RestoreDoubtInteractor(data, new(presenter.DoubtWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	dwc := controller.NewDoubtWebControllerWithProvider(doubtKV, doubtFactory)
	mux.HandleFunc("/doubt/exec", dwc.Exec)

	// Daifugo (KV-backed session)
	daifugoFactory := func() usecase.DaifugoInteractorIF {
		config := domain.DefaultDaifugoConfig()
		players := []*domain.DaifugoPlayer{
			domain.NewDaifugoPlayer(true),
			domain.NewDaifugoPlayer(false),
			domain.NewDaifugoPlayer(false),
			domain.NewDaifugoPlayer(false),
		}
		daifugo := domain.NewDaifugo(domain.NewTrumpCards(config.JokerCount), players, config)
		return usecase.NewDaifugoInteractor(daifugo, new(presenter.DaifugoWebPresenter))
	}
	daifugoKV, err := controller.NewKVSessionProvider[usecase.DaifugoInteractorIF](
		"GAME_SESSIONS", "daifugo:",
		func(di usecase.DaifugoInteractorIF) ([]byte, error) {
			snap, ok := di.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.DaifugoInteractorIF, error) {
			return usecase.RestoreDaifugoInteractor(data, new(presenter.DaifugoWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	dgc := controller.NewDaifugoWebControllerWithProvider(daifugoKV, daifugoFactory)
	mux.HandleFunc("/daifugo/exec", dgc.Exec)

	// Sevens (KV-backed session)
	sevensFactory := func() usecase.SevensInteractorIF {
		config := domain.DefaultSevensConfig()
		players := []*domain.SevensPlayer{
			domain.NewSevensPlayer(true),
			domain.NewSevensPlayer(false),
			domain.NewSevensPlayer(false),
			domain.NewSevensPlayer(false),
		}
		sevens := domain.NewSevens(domain.NewTrumpCards(config.JokerCount), players, config)
		return usecase.NewSevensInteractor(sevens, new(presenter.SevensWebPresenter))
	}
	sevensKV, err := controller.NewKVSessionProvider[usecase.SevensInteractorIF](
		"GAME_SESSIONS", "sevens:",
		func(si usecase.SevensInteractorIF) ([]byte, error) {
			snap, ok := si.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.SevensInteractorIF, error) {
			return usecase.RestoreSevensInteractor(data, new(presenter.SevensWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	sgc := controller.NewSevensWebControllerWithProvider(sevensKV, sevensFactory)
	mux.HandleFunc("/sevens/exec", sgc.Exec)

	// Crazy Eights (KV-backed session)
	crazyeightsFactory := func() usecase.CrazyEightsInteractorIF {
		config := domain.DefaultCrazyEightsConfig()
		players := []*domain.CrazyEightsPlayer{
			domain.NewCrazyEightsPlayer(true),
			domain.NewCrazyEightsPlayer(false),
			domain.NewCrazyEightsPlayer(false),
			domain.NewCrazyEightsPlayer(false),
		}
		ce := domain.NewCrazyEights(domain.NewTrumpCards(0), players, config)
		return usecase.NewCrazyEightsInteractor(ce, new(presenter.CrazyEightsWebPresenter))
	}
	crazyeightsKV, err := controller.NewKVSessionProvider[usecase.CrazyEightsInteractorIF](
		"GAME_SESSIONS", "crazyeights:",
		func(ci usecase.CrazyEightsInteractorIF) ([]byte, error) {
			snap, ok := ci.(interface{ Snapshot() ([]byte, error) })
			if !ok {
				return nil, fmt.Errorf("interactor does not support snapshotting")
			}
			return snap.Snapshot()
		},
		func(data []byte) (usecase.CrazyEightsInteractorIF, error) {
			return usecase.RestoreCrazyEightsInteractor(data, new(presenter.CrazyEightsWebPresenter))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	cec := controller.NewCrazyEightsWebControllerWithProvider(crazyeightsKV, crazyeightsFactory)
	mux.HandleFunc("/crazyeights/exec", cec.Exec)

	var handler http.Handler = mux
	if origins := corsmw.ParseOrigins(cloudflare.Getenv("CORS_ALLOWED_ORIGINS")); origins != nil {
		handler = corsmw.Middleware(origins, mux)
	}
	workers.Serve(handler)
}
