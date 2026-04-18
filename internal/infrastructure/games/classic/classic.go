//go:build js && wasm

// Package classic binds the Cloudflare Worker KV-backed handlers for the
// 21 trick-taking, matching, and family card games. A worker main must
// blank-import this package so the init below runs before
// games.RegisterCategory is called.
package classic

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/worker"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func init() {
	games.BindWorker("oldmaid", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/oldmaid/exec", "oldmaid:",
			func() usecase.OldMaidInteractorIF {
				return usecase.NewOldMaidInteractor(domain.NewDefaultOldMaid(), new(presenter.OldMaidWebPresenter))
			},
			func(data []byte) (usecase.OldMaidInteractorIF, error) {
				return usecase.RestoreOldMaidInteractor(data, new(presenter.OldMaidWebPresenter))
			},
			controller.NewOldMaidWebControllerWithProvider,
		)
	})
	games.BindWorker("daifugo", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/daifugo/exec", "daifugo:",
			func() usecase.DaifugoInteractorIF {
				return usecase.NewDaifugoInteractor(domain.NewDefaultDaifugo(), new(presenter.DaifugoWebPresenter))
			},
			func(data []byte) (usecase.DaifugoInteractorIF, error) {
				return usecase.RestoreDaifugoInteractor(data, new(presenter.DaifugoWebPresenter))
			},
			controller.NewDaifugoWebControllerWithProvider,
		)
	})
	games.BindWorker("sevens", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/sevens/exec", "sevens:",
			func() usecase.SevensInteractorIF {
				return usecase.NewSevensInteractor(domain.NewDefaultSevens(), new(presenter.SevensWebPresenter))
			},
			func(data []byte) (usecase.SevensInteractorIF, error) {
				return usecase.RestoreSevensInteractor(data, new(presenter.SevensWebPresenter))
			},
			controller.NewSevensWebControllerWithProvider,
		)
	})
	games.BindWorker("doubt", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/doubt/exec", "doubt:",
			func() usecase.DoubtInteractorIF {
				return usecase.NewDoubtInteractor(domain.NewDefaultDoubt(), new(presenter.DoubtWebPresenter))
			},
			func(data []byte) (usecase.DoubtInteractorIF, error) {
				return usecase.RestoreDoubtInteractor(data, new(presenter.DoubtWebPresenter))
			},
			controller.NewDoubtWebControllerWithProvider,
		)
	})
	games.BindWorker("hearts", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/hearts/exec", "hearts:",
			func() usecase.HeartsInteractorIF {
				return usecase.NewHeartsInteractor(domain.NewDefaultHearts(), new(presenter.HeartsWebPresenter))
			},
			func(data []byte) (usecase.HeartsInteractorIF, error) {
				return usecase.RestoreHeartsInteractor(data, new(presenter.HeartsWebPresenter))
			},
			controller.NewHeartsWebControllerWithProvider,
		)
	})
	games.BindWorker("spades", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/spades/exec", "spades:",
			func() usecase.SpadesInteractorIF {
				return usecase.NewSpadesInteractor(domain.NewDefaultSpades(), new(presenter.SpadesWebPresenter))
			},
			func(data []byte) (usecase.SpadesInteractorIF, error) {
				return usecase.RestoreSpadesInteractor(data, new(presenter.SpadesWebPresenter))
			},
			controller.NewSpadesWebControllerWithProvider,
		)
	})
	games.BindWorker("crazyeights", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/crazyeights/exec", "crazyeights:",
			func() usecase.CrazyEightsInteractorIF {
				return usecase.NewCrazyEightsInteractor(domain.NewDefaultCrazyEights(), new(presenter.CrazyEightsWebPresenter))
			},
			func(data []byte) (usecase.CrazyEightsInteractorIF, error) {
				return usecase.RestoreCrazyEightsInteractor(data, new(presenter.CrazyEightsWebPresenter))
			},
			controller.NewCrazyEightsWebControllerWithProvider,
		)
	})
	games.BindWorker("napoleon", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/napoleon/exec", "napoleon:",
			func() usecase.NapoleonInteractorIF {
				return usecase.NewNapoleonInteractor(domain.NewDefaultNapoleon(), new(presenter.NapoleonWebPresenter))
			},
			func(data []byte) (usecase.NapoleonInteractorIF, error) {
				return usecase.RestoreNapoleonInteractor(data, new(presenter.NapoleonWebPresenter))
			},
			controller.NewNapoleonWebControllerWithProvider,
		)
	})
	games.BindWorker("euchre", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/euchre/exec", "euchre:",
			func() usecase.EuchreInteractorIF {
				return usecase.NewEuchreInteractor(domain.NewDefaultEuchre(), new(presenter.EuchreWebPresenter))
			},
			func(data []byte) (usecase.EuchreInteractorIF, error) {
				return usecase.RestoreEuchreInteractor(data, new(presenter.EuchreWebPresenter))
			},
			controller.NewEuchreWebControllerWithProvider,
		)
	})
	games.BindWorker("ohhell", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/ohhell/exec", "ohhell:",
			func() usecase.OhHellInteractorIF {
				return usecase.NewOhHellInteractor(domain.NewDefaultOhHell(), new(presenter.OhHellWebPresenter))
			},
			func(data []byte) (usecase.OhHellInteractorIF, error) {
				return usecase.RestoreOhHellInteractor(data, new(presenter.OhHellWebPresenter))
			},
			controller.NewOhHellWebControllerWithProvider,
		)
	})
	games.BindWorker("bridge", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/bridge/exec", "bridge:",
			func() usecase.BridgeInteractorIF {
				return usecase.NewBridgeInteractor(domain.NewDefaultBridge(), new(presenter.BridgeWebPresenter))
			},
			func(data []byte) (usecase.BridgeInteractorIF, error) {
				return usecase.RestoreBridgeInteractor(data, new(presenter.BridgeWebPresenter))
			},
			controller.NewBridgeWebControllerWithProvider,
		)
	})
	games.BindWorker("speed", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/speed/exec", "speed:",
			func() usecase.SpeedInteractorIF {
				return usecase.NewSpeedInteractor(domain.NewDefaultSpeed(), new(presenter.SpeedWebPresenter))
			},
			func(data []byte) (usecase.SpeedInteractorIF, error) {
				return usecase.RestoreSpeedInteractor(data, new(presenter.SpeedWebPresenter))
			},
			controller.NewSpeedWebControllerWithProvider,
		)
	})
	games.BindWorker("gofish", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/gofish/exec", "gofish:",
			func() usecase.GoFishInteractorIF {
				return usecase.NewGoFishInteractor(domain.NewDefaultGoFish(), new(presenter.GoFishWebPresenter))
			},
			func(data []byte) (usecase.GoFishInteractorIF, error) {
				return usecase.RestoreGoFishInteractor(data, new(presenter.GoFishWebPresenter))
			},
			controller.NewGoFishWebControllerWithProvider,
		)
	})
	games.BindWorker("pinochle", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/pinochle/exec", "pinochle:",
			func() usecase.PinochleInteractorIF {
				return usecase.NewPinochleInteractor(domain.NewDefaultPinochle(), new(presenter.PinochleWebPresenter))
			},
			func(data []byte) (usecase.PinochleInteractorIF, error) {
				return usecase.RestorePinochleInteractor(data, new(presenter.PinochleWebPresenter))
			},
			controller.NewPinochleWebControllerWithProvider,
		)
	})
	games.BindWorker("pigtail", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/pigtail/exec", "pigtail:",
			func() usecase.PigsTailInteractorIF {
				return usecase.NewPigsTailInteractor(domain.NewDefaultPigsTail(), new(presenter.PigsTailWebPresenter))
			},
			func(data []byte) (usecase.PigsTailInteractorIF, error) {
				return usecase.RestorePigsTailInteractor(data, new(presenter.PigsTailWebPresenter))
			},
			controller.NewPigsTailWebControllerWithProvider,
		)
	})
	games.BindWorker("durak", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/durak/exec", "durak:",
			func() usecase.DurakInteractorIF {
				return usecase.NewDurakInteractor(domain.NewDefaultDurak(), new(presenter.DurakWebPresenter))
			},
			func(data []byte) (usecase.DurakInteractorIF, error) {
				return usecase.RestoreDurakInteractor(data, new(presenter.DurakWebPresenter))
			},
			controller.NewDurakWebControllerWithProvider,
		)
	})
	games.BindWorker("twotenjack", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/twotenjack/exec", "twotenjack:",
			func() usecase.TwoTenJackInteractorIF {
				return usecase.NewTwoTenJackInteractor(domain.NewDefaultTwoTenJack(), new(presenter.TwoTenJackWebPresenter))
			},
			func(data []byte) (usecase.TwoTenJackInteractorIF, error) {
				return usecase.RestoreTwoTenJackInteractor(data, new(presenter.TwoTenJackWebPresenter))
			},
			controller.NewTwoTenJackWebControllerWithProvider,
		)
	})
	games.BindWorker("war", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/war/exec", "war:",
			func() usecase.WarInteractorIF {
				return usecase.NewWarInteractor(domain.NewDefaultWar(), new(presenter.WarWebPresenter))
			},
			func(data []byte) (usecase.WarInteractorIF, error) {
				return usecase.RestoreWarInteractor(data, new(presenter.WarWebPresenter))
			},
			controller.NewWarWebControllerWithProvider,
		)
	})
	games.BindWorker("fiftyone", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/fiftyone/exec", "fiftyone:",
			func() usecase.FiftyOneInteractorIF {
				return usecase.NewFiftyOneInteractor(domain.NewDefaultFiftyOne(), new(presenter.FiftyOneWebPresenter))
			},
			func(data []byte) (usecase.FiftyOneInteractorIF, error) {
				return usecase.RestoreFiftyOneInteractor(data, new(presenter.FiftyOneWebPresenter))
			},
			controller.NewFiftyOneWebControllerWithProvider,
		)
	})
	games.BindWorker("whist", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/whist/exec", "whist:",
			func() usecase.WhistInteractorIF {
				return usecase.NewWhistInteractor(domain.NewDefaultWhist(), new(presenter.WhistWebPresenter))
			},
			func(data []byte) (usecase.WhistInteractorIF, error) {
				return usecase.RestoreWhistInteractor(data, new(presenter.WhistWebPresenter))
			},
			controller.NewWhistWebControllerWithProvider,
		)
	})
	games.BindWorker("pageone", games.CategoryClassic, func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/pageone/exec", "pageone:",
			func() usecase.PageOneInteractorIF {
				return usecase.NewPageOneInteractor(domain.NewDefaultPageOne(), new(presenter.PageOneWebPresenter))
			},
			func(data []byte) (usecase.PageOneInteractorIF, error) {
				return usecase.RestorePageOneInteractor(data, new(presenter.PageOneWebPresenter))
			},
			controller.NewPageOneWebControllerWithProvider,
		)
	})
}
