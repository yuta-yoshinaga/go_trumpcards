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
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/worker"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func main() {
	mux := http.NewServeMux()

	// Hearts
	if err := worker.RegisterKV(mux, "/hearts/exec", "hearts:",
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
		controller.NewHeartsWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Spades
	if err := worker.RegisterKV(mux, "/spades/exec", "spades:",
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
		controller.NewSpadesWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Euchre
	if err := worker.RegisterKV(mux, "/euchre/exec", "euchre:",
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
		controller.NewEuchreWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Napoleon
	if err := worker.RegisterKV(mux, "/napoleon/exec", "napoleon:",
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
		controller.NewNapoleonWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Old Maid
	if err := worker.RegisterKV(mux, "/oldmaid/exec", "oldmaid:",
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
		controller.NewOldMaidWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Doubt
	if err := worker.RegisterKV(mux, "/doubt/exec", "doubt:",
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
		controller.NewDoubtWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Daifugo
	if err := worker.RegisterKV(mux, "/daifugo/exec", "daifugo:",
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
		controller.NewDaifugoWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Sevens
	if err := worker.RegisterKV(mux, "/sevens/exec", "sevens:",
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
		controller.NewSevensWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Crazy Eights
	if err := worker.RegisterKV(mux, "/crazyeights/exec", "crazyeights:",
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
		controller.NewCrazyEightsWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Oh Hell
	if err := worker.RegisterKV(mux, "/ohhell/exec", "ohhell:",
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
		controller.NewOhHellWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Contract Bridge
	if err := worker.RegisterKV(mux, "/bridge/exec", "bridge:",
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
		controller.NewBridgeWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Speed
	if err := worker.RegisterKV(mux, "/speed/exec", "speed:",
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
		controller.NewSpeedWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Go Fish
	if err := worker.RegisterKV(mux, "/gofish/exec", "gofish:",
		func() usecase.GoFishInteractorIF {
			players := []*domain.GoFishPlayer{
				domain.NewGoFishPlayer(true),
				domain.NewGoFishPlayer(false),
				domain.NewGoFishPlayer(false),
				domain.NewGoFishPlayer(false),
			}
			goFish := domain.NewGoFish(domain.NewTrumpCards(0), players)
			return usecase.NewGoFishInteractor(goFish, new(presenter.GoFishWebPresenter))
		},
		func(data []byte) (usecase.GoFishInteractorIF, error) {
			return usecase.RestoreGoFishInteractor(data, new(presenter.GoFishWebPresenter))
		},
		controller.NewGoFishWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Pinochle
	if err := worker.RegisterKV(mux, "/pinochle/exec", "pinochle:",
		func() usecase.PinochleInteractorIF {
			config := domain.DefaultPinochleConfig()
			players := []*domain.PinochlePlayer{
				domain.NewPinochlePlayer(true, 0),
				domain.NewPinochlePlayer(false, 1),
				domain.NewPinochlePlayer(false, 0),
				domain.NewPinochlePlayer(false, 1),
			}
			pinochle := domain.NewPinochle(domain.NewTrumpCardsPinochle(), players, config)
			return usecase.NewPinochleInteractor(pinochle, new(presenter.PinochleWebPresenter))
		},
		func(data []byte) (usecase.PinochleInteractorIF, error) {
			return usecase.RestorePinochleInteractor(data, new(presenter.PinochleWebPresenter))
		},
		controller.NewPinochleWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Pig's Tail
	if err := worker.RegisterKV(mux, "/pigtail/exec", "pigtail:",
		func() usecase.PigsTailInteractorIF {
			players := []*domain.PigsTailPlayer{
				domain.NewPigsTailPlayer(true),
				domain.NewPigsTailPlayer(false),
				domain.NewPigsTailPlayer(false),
				domain.NewPigsTailPlayer(false),
			}
			pigsTail := domain.NewPigsTail(domain.NewTrumpCards(0), players)
			return usecase.NewPigsTailInteractor(pigsTail, new(presenter.PigsTailWebPresenter))
		},
		func(data []byte) (usecase.PigsTailInteractorIF, error) {
			return usecase.RestorePigsTailInteractor(data, new(presenter.PigsTailWebPresenter))
		},
		controller.NewPigsTailWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Two Ten Jack
	if err := worker.RegisterKV(mux, "/twotenjack/exec", "twotenjack:",
		func() usecase.TwoTenJackInteractorIF {
			config := domain.DefaultTwoTenJackConfig()
			players := []*domain.TwoTenJackPlayer{
				domain.NewTwoTenJackPlayer(true),
				domain.NewTwoTenJackPlayer(false),
				domain.NewTwoTenJackPlayer(false),
				domain.NewTwoTenJackPlayer(false),
			}
			ttj := domain.NewTwoTenJack(domain.NewTrumpCards(0), players, config)
			return usecase.NewTwoTenJackInteractor(ttj, new(presenter.TwoTenJackWebPresenter))
		},
		func(data []byte) (usecase.TwoTenJackInteractorIF, error) {
			return usecase.RestoreTwoTenJackInteractor(data, new(presenter.TwoTenJackWebPresenter))
		},
		controller.NewTwoTenJackWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// War
	if err := worker.RegisterKV(mux, "/war/exec", "war:",
		func() usecase.WarInteractorIF {
			config := domain.DefaultWarConfig()
			players := []*domain.WarPlayer{
				domain.NewWarPlayer(true),
				domain.NewWarPlayer(false),
			}
			war := domain.NewWar(domain.NewTrumpCards(0), players, config)
			return usecase.NewWarInteractor(war, new(presenter.WarWebPresenter))
		},
		func(data []byte) (usecase.WarInteractorIF, error) {
			return usecase.RestoreWarInteractor(data, new(presenter.WarWebPresenter))
		},
		controller.NewWarWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Durak
	if err := worker.RegisterKV(mux, "/durak/exec", "durak:",
		func() usecase.DurakInteractorIF {
			players := []*domain.DurakPlayer{
				domain.NewDurakPlayer(true),
				domain.NewDurakPlayer(false),
				domain.NewDurakPlayer(false),
				domain.NewDurakPlayer(false),
			}
			d := domain.NewDurak(domain.NewTrumpCardsShortDeck(), players)
			return usecase.NewDurakInteractor(d, new(presenter.DurakWebPresenter))
		},
		func(data []byte) (usecase.DurakInteractorIF, error) {
			return usecase.RestoreDurakInteractor(data, new(presenter.DurakWebPresenter))
		},
		controller.NewDurakWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Fifty-one
	if err := worker.RegisterKV(mux, "/fiftyone/exec", "fiftyone:",
		func() usecase.FiftyOneInteractorIF {
			players := []*domain.FiftyOnePlayer{
				domain.NewFiftyOnePlayer(true),
				domain.NewFiftyOnePlayer(false),
				domain.NewFiftyOnePlayer(false),
				domain.NewFiftyOnePlayer(false),
			}
			fo := domain.NewFiftyOne(domain.NewTrumpCards(0), players)
			return usecase.NewFiftyOneInteractor(fo, new(presenter.FiftyOneWebPresenter))
		},
		func(data []byte) (usecase.FiftyOneInteractorIF, error) {
			return usecase.RestoreFiftyOneInteractor(data, new(presenter.FiftyOneWebPresenter))
		},
		controller.NewFiftyOneWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	// Whist
	if err := worker.RegisterKV(mux, "/whist/exec", "whist:",
		func() usecase.WhistInteractorIF {
			config := domain.DefaultWhistConfig()
			players := []*domain.WhistPlayer{
				domain.NewWhistPlayer(true, 0),
				domain.NewWhistPlayer(false, 1),
				domain.NewWhistPlayer(false, 0),
				domain.NewWhistPlayer(false, 1),
			}
			whist := domain.NewWhist(domain.NewTrumpCards(0), players, config)
			return usecase.NewWhistInteractor(whist, new(presenter.WhistWebPresenter))
		},
		func(data []byte) (usecase.WhistInteractorIF, error) {
			return usecase.RestoreWhistInteractor(data, new(presenter.WhistWebPresenter))
		},
		controller.NewWhistWebControllerWithProvider,
	); err != nil {
		log.Fatal(err)
	}

	var handler http.Handler = mux
	if origins := corsmw.ParseOrigins(cloudflare.Getenv("CORS_ALLOWED_ORIGINS")); origins != nil {
		handler = corsmw.Middleware(origins, mux)
	}
	workers.Serve(handler)
}
