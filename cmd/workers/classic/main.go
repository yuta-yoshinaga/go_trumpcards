//go:build js && wasm

package main

import (
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

// registerKV creates a KV-backed session provider for a game and registers the
// controller on the given mux. It eliminates the repeated boilerplate of
// creating a provider, building a controller, and wiring up the route.
func registerKV[I any](
	mux *http.ServeMux,
	path string,
	kvPrefix string,
	factory func() I,
	restore func([]byte) (I, error),
	newCtrl func(controller.SessionProvider[I], func() I) interface {
		Exec(http.ResponseWriter, *http.Request)
		Stop()
	},
) {
	kvProvider, err := controller.NewKVSessionProvider[I](
		"GAME_SESSIONS", kvPrefix,
		func(i I) ([]byte, error) {
			return any(i).(interface{ Snapshot() ([]byte, error) }).Snapshot()
		},
		restore,
	)
	if err != nil {
		log.Fatal(err)
	}
	ctrl := newCtrl(kvProvider, factory)
	mux.HandleFunc(path, ctrl.Exec)
}

func main() {
	mux := http.NewServeMux()

	// Hearts
	registerKV(mux, "/hearts/exec", "hearts:",
		func() usecase.HeartsInteractorIF {
			config := domain.DefaultHeartsConfig()
			players := []*domain.HeartsPlayer{
				domain.NewHeartsPlayer(true),
				domain.NewHeartsPlayer(false),
				domain.NewHeartsPlayer(false),
				domain.NewHeartsPlayer(false),
			}
			hearts := domain.NewHearts(domain.NewTrumpCards(0), players, config)
			return usecase.NewHeartsInteractor(hearts, new(presenter.HeartsWebPresenter))
		},
		func(data []byte) (usecase.HeartsInteractorIF, error) {
			return usecase.RestoreHeartsInteractor(data, new(presenter.HeartsWebPresenter))
		},
		func(p controller.SessionProvider[usecase.HeartsInteractorIF], f func() usecase.HeartsInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewHeartsWebControllerWithProvider(p, f)
		},
	)

	// Spades
	registerKV(mux, "/spades/exec", "spades:",
		func() usecase.SpadesInteractorIF {
			config := domain.DefaultSpadesConfig()
			players := []*domain.SpadesPlayer{
				domain.NewSpadesPlayer(true),
				domain.NewSpadesPlayer(false),
				domain.NewSpadesPlayer(false),
				domain.NewSpadesPlayer(false),
			}
			spades := domain.NewSpades(domain.NewTrumpCards(0), players, config)
			return usecase.NewSpadesInteractor(spades, new(presenter.SpadesWebPresenter))
		},
		func(data []byte) (usecase.SpadesInteractorIF, error) {
			return usecase.RestoreSpadesInteractor(data, new(presenter.SpadesWebPresenter))
		},
		func(p controller.SessionProvider[usecase.SpadesInteractorIF], f func() usecase.SpadesInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewSpadesWebControllerWithProvider(p, f)
		},
	)

	// Euchre
	registerKV(mux, "/euchre/exec", "euchre:",
		func() usecase.EuchreInteractorIF {
			config := domain.DefaultEuchreConfig()
			players := []*domain.EuchrePlayer{
				domain.NewEuchrePlayer(true, 0),
				domain.NewEuchrePlayer(false, 1),
				domain.NewEuchrePlayer(false, 0),
				domain.NewEuchrePlayer(false, 1),
			}
			euchre := domain.NewEuchre(domain.NewTrumpCardsEuchre(), players, config)
			return usecase.NewEuchreInteractor(euchre, new(presenter.EuchreWebPresenter))
		},
		func(data []byte) (usecase.EuchreInteractorIF, error) {
			return usecase.RestoreEuchreInteractor(data, new(presenter.EuchreWebPresenter))
		},
		func(p controller.SessionProvider[usecase.EuchreInteractorIF], f func() usecase.EuchreInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewEuchreWebControllerWithProvider(p, f)
		},
	)

	// Napoleon
	registerKV(mux, "/napoleon/exec", "napoleon:",
		func() usecase.NapoleonInteractorIF {
			config := domain.DefaultNapoleonConfig()
			players := []*domain.NapoleonPlayer{
				domain.NewNapoleonPlayer(true),
				domain.NewNapoleonPlayer(false),
				domain.NewNapoleonPlayer(false),
				domain.NewNapoleonPlayer(false),
			}
			napoleon := domain.NewNapoleon(domain.NewTrumpCards(1), players, config)
			return usecase.NewNapoleonInteractor(napoleon, new(presenter.NapoleonWebPresenter))
		},
		func(data []byte) (usecase.NapoleonInteractorIF, error) {
			return usecase.RestoreNapoleonInteractor(data, new(presenter.NapoleonWebPresenter))
		},
		func(p controller.SessionProvider[usecase.NapoleonInteractorIF], f func() usecase.NapoleonInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewNapoleonWebControllerWithProvider(p, f)
		},
	)

	// Old Maid
	registerKV(mux, "/oldmaid/exec", "oldmaid:",
		func() usecase.OldMaidInteractorIF {
			players := []*domain.OldMaidPlayer{
				domain.NewOldMaidPlayer(true),
				domain.NewOldMaidPlayer(false),
				domain.NewOldMaidPlayer(false),
				domain.NewOldMaidPlayer(false),
			}
			oldMaid := domain.NewOldMaid(domain.NewTrumpCards(1), players)
			return usecase.NewOldMaidInteractor(oldMaid, new(presenter.OldMaidWebPresenter))
		},
		func(data []byte) (usecase.OldMaidInteractorIF, error) {
			return usecase.RestoreOldMaidInteractor(data, new(presenter.OldMaidWebPresenter))
		},
		func(p controller.SessionProvider[usecase.OldMaidInteractorIF], f func() usecase.OldMaidInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewOldMaidWebControllerWithProvider(p, f)
		},
	)

	// Doubt
	registerKV(mux, "/doubt/exec", "doubt:",
		func() usecase.DoubtInteractorIF {
			players := []*domain.DoubtPlayer{
				domain.NewDoubtPlayer(true),
				domain.NewDoubtPlayer(false),
				domain.NewDoubtPlayer(false),
				domain.NewDoubtPlayer(false),
			}
			doubt := domain.NewDoubt(domain.NewTrumpCards(0), players)
			return usecase.NewDoubtInteractor(doubt, new(presenter.DoubtWebPresenter))
		},
		func(data []byte) (usecase.DoubtInteractorIF, error) {
			return usecase.RestoreDoubtInteractor(data, new(presenter.DoubtWebPresenter))
		},
		func(p controller.SessionProvider[usecase.DoubtInteractorIF], f func() usecase.DoubtInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewDoubtWebControllerWithProvider(p, f)
		},
	)

	// Daifugo
	registerKV(mux, "/daifugo/exec", "daifugo:",
		func() usecase.DaifugoInteractorIF {
			config := domain.DefaultDaifugoConfig()
			players := []*domain.DaifugoPlayer{
				domain.NewDaifugoPlayer(true),
				domain.NewDaifugoPlayer(false),
				domain.NewDaifugoPlayer(false),
				domain.NewDaifugoPlayer(false),
			}
			daifugo := domain.NewDaifugo(domain.NewTrumpCards(config.JokerCount), players, config)
			return usecase.NewDaifugoInteractor(daifugo, new(presenter.DaifugoWebPresenter))
		},
		func(data []byte) (usecase.DaifugoInteractorIF, error) {
			return usecase.RestoreDaifugoInteractor(data, new(presenter.DaifugoWebPresenter))
		},
		func(p controller.SessionProvider[usecase.DaifugoInteractorIF], f func() usecase.DaifugoInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewDaifugoWebControllerWithProvider(p, f)
		},
	)

	// Sevens
	registerKV(mux, "/sevens/exec", "sevens:",
		func() usecase.SevensInteractorIF {
			config := domain.DefaultSevensConfig()
			players := []*domain.SevensPlayer{
				domain.NewSevensPlayer(true),
				domain.NewSevensPlayer(false),
				domain.NewSevensPlayer(false),
				domain.NewSevensPlayer(false),
			}
			sevens := domain.NewSevens(domain.NewTrumpCards(config.JokerCount), players, config)
			return usecase.NewSevensInteractor(sevens, new(presenter.SevensWebPresenter))
		},
		func(data []byte) (usecase.SevensInteractorIF, error) {
			return usecase.RestoreSevensInteractor(data, new(presenter.SevensWebPresenter))
		},
		func(p controller.SessionProvider[usecase.SevensInteractorIF], f func() usecase.SevensInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewSevensWebControllerWithProvider(p, f)
		},
	)

	// Crazy Eights
	registerKV(mux, "/crazyeights/exec", "crazyeights:",
		func() usecase.CrazyEightsInteractorIF {
			config := domain.DefaultCrazyEightsConfig()
			players := []*domain.CrazyEightsPlayer{
				domain.NewCrazyEightsPlayer(true),
				domain.NewCrazyEightsPlayer(false),
				domain.NewCrazyEightsPlayer(false),
				domain.NewCrazyEightsPlayer(false),
			}
			ce := domain.NewCrazyEights(domain.NewTrumpCards(0), players, config)
			return usecase.NewCrazyEightsInteractor(ce, new(presenter.CrazyEightsWebPresenter))
		},
		func(data []byte) (usecase.CrazyEightsInteractorIF, error) {
			return usecase.RestoreCrazyEightsInteractor(data, new(presenter.CrazyEightsWebPresenter))
		},
		func(p controller.SessionProvider[usecase.CrazyEightsInteractorIF], f func() usecase.CrazyEightsInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewCrazyEightsWebControllerWithProvider(p, f)
		},
	)

	// Oh Hell
	registerKV(mux, "/ohhell/exec", "ohhell:",
		func() usecase.OhHellInteractorIF {
			config := domain.DefaultOhHellConfig()
			players := []*domain.OhHellPlayer{
				domain.NewOhHellPlayer(true),
				domain.NewOhHellPlayer(false),
				domain.NewOhHellPlayer(false),
				domain.NewOhHellPlayer(false),
			}
			ohHell := domain.NewOhHell(domain.NewTrumpCards(0), players, config)
			return usecase.NewOhHellInteractor(ohHell, new(presenter.OhHellWebPresenter))
		},
		func(data []byte) (usecase.OhHellInteractorIF, error) {
			return usecase.RestoreOhHellInteractor(data, new(presenter.OhHellWebPresenter))
		},
		func(p controller.SessionProvider[usecase.OhHellInteractorIF], f func() usecase.OhHellInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewOhHellWebControllerWithProvider(p, f)
		},
	)

	// Contract Bridge
	registerKV(mux, "/bridge/exec", "bridge:",
		func() usecase.BridgeInteractorIF {
			config := domain.DefaultBridgeConfig()
			players := []*domain.BridgePlayer{
				domain.NewBridgePlayer(true, 0),
				domain.NewBridgePlayer(false, 1),
				domain.NewBridgePlayer(false, 0),
				domain.NewBridgePlayer(false, 1),
			}
			bridge := domain.NewBridge(domain.NewTrumpCards(0), players, config)
			return usecase.NewBridgeInteractor(bridge, new(presenter.BridgeWebPresenter))
		},
		func(data []byte) (usecase.BridgeInteractorIF, error) {
			return usecase.RestoreBridgeInteractor(data, new(presenter.BridgeWebPresenter))
		},
		func(p controller.SessionProvider[usecase.BridgeInteractorIF], f func() usecase.BridgeInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewBridgeWebControllerWithProvider(p, f)
		},
	)

	// Speed
	registerKV(mux, "/speed/exec", "speed:",
		func() usecase.SpeedInteractorIF {
			config := domain.DefaultSpeedConfig()
			players := []*domain.SpeedPlayer{
				domain.NewSpeedPlayer(true),
				domain.NewSpeedPlayer(false),
			}
			speed := domain.NewSpeed(domain.NewTrumpCards(0), players, config)
			return usecase.NewSpeedInteractor(speed, new(presenter.SpeedWebPresenter))
		},
		func(data []byte) (usecase.SpeedInteractorIF, error) {
			return usecase.RestoreSpeedInteractor(data, new(presenter.SpeedWebPresenter))
		},
		func(p controller.SessionProvider[usecase.SpeedInteractorIF], f func() usecase.SpeedInteractorIF) interface {
			Exec(http.ResponseWriter, *http.Request)
			Stop()
		} {
			return controller.NewSpeedWebControllerWithProvider(p, f)
		},
	)

	var handler http.Handler = mux
	if origins := corsmw.ParseOrigins(cloudflare.Getenv("CORS_ALLOWED_ORIGINS")); origins != nil {
		handler = corsmw.Middleware(origins, mux)
	}
	workers.Serve(handler)
}
