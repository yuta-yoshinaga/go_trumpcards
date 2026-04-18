//go:build js && wasm

package games

import (
	"fmt"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/worker"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RegisterCategory registers every game in cat onto mux using the KV-backed
// session provider. Returns the first error encountered; callers should
// treat a non-nil error as fatal (typically log.Fatal in a worker's main).
func RegisterCategory(mux *http.ServeMux, cat Category) error {
	for _, g := range ByCategory(cat) {
		if g.RegisterWorker == nil {
			return fmt.Errorf("games: %q has no RegisterWorker (missing from games_wasm.go)", g.Name)
		}
		if err := g.RegisterWorker(mux); err != nil {
			return fmt.Errorf("games: register %q: %w", g.Name, err)
		}
	}
	return nil
}

// find locates a game by name; returns nil if not found.
func find(name string) *Game {
	for _, g := range registry {
		if g.Name == name {
			return g
		}
	}
	return nil
}

// bind attaches a RegisterWorker closure to the named game. Panics at
// package-init time if the name is unknown, which surfaces typos early.
func bind(name string, r func(mux *http.ServeMux) error) {
	g := find(name)
	if g == nil {
		panic(fmt.Sprintf("games: bind worker: unknown game %q", name))
	}
	g.RegisterWorker = r
}

// init populates the RegisterWorker field for every entry in the registry.
// Each closure captures the per-game types needed by worker.RegisterKV
// (interactor interface, restore func, provider-aware controller ctor).
func init() {
	bind("blackjack", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/blackjack/exec", "blackjack:",
			func() usecase.BlackJackInteractorIF {
				return usecase.NewBlackJackInteractor(domain.NewDefaultBlackJack(), new(presenter.BlackJackWebPresenter))
			},
			func(data []byte) (usecase.BlackJackInteractorIF, error) {
				return usecase.RestoreBlackJackInteractor(data, new(presenter.BlackJackWebPresenter))
			},
			controller.NewBlackJackWebControllerWithProvider,
		)
	})
	bind("poker", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/poker/exec", "poker:",
			func() usecase.PokerInteractorIF {
				return usecase.NewPokerInteractor(domain.NewDefaultPoker(), new(presenter.PokerWebPresenter))
			},
			func(data []byte) (usecase.PokerInteractorIF, error) {
				return usecase.RestorePokerInteractor(data, new(presenter.PokerWebPresenter))
			},
			controller.NewPokerWebControllerWithProvider,
		)
	})
	bind("oldmaid", func(mux *http.ServeMux) error {
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
	bind("daifugo", func(mux *http.ServeMux) error {
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
	bind("sevens", func(mux *http.ServeMux) error {
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
	bind("doubt", func(mux *http.ServeMux) error {
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
	bind("holdem", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/holdem/exec", "holdem:",
			func() usecase.HoldemInteractorIF {
				return usecase.NewHoldemInteractor(domain.NewDefaultHoldem(), new(presenter.HoldemWebPresenter))
			},
			func(data []byte) (usecase.HoldemInteractorIF, error) {
				return usecase.RestoreHoldemInteractor(data, new(presenter.HoldemWebPresenter))
			},
			controller.NewHoldemWebControllerWithProvider,
		)
	})
	bind("omaha", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/omaha/exec", "omaha:",
			func() usecase.OmahaInteractorIF {
				return usecase.NewOmahaInteractor(domain.NewDefaultOmaha(), new(presenter.OmahaWebPresenter))
			},
			func(data []byte) (usecase.OmahaInteractorIF, error) {
				return usecase.RestoreOmahaInteractor(data, new(presenter.OmahaWebPresenter))
			},
			controller.NewOmahaWebControllerWithProvider,
		)
	})
	bind("shortdeck", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/shortdeck/exec", "shortdeck:",
			func() usecase.ShortDeckInteractorIF {
				return usecase.NewShortDeckInteractor(domain.NewDefaultShortDeck(), new(presenter.ShortDeckWebPresenter))
			},
			func(data []byte) (usecase.ShortDeckInteractorIF, error) {
				return usecase.RestoreShortDeckInteractor(data, new(presenter.ShortDeckWebPresenter))
			},
			controller.NewShortDeckWebControllerWithProvider,
		)
	})
	bind("pineapple", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/pineapple/exec", "pineapple:",
			func() usecase.PineappleInteractorIF {
				return usecase.NewPineappleInteractor(domain.NewDefaultPineapple(), new(presenter.PineappleWebPresenter))
			},
			func(data []byte) (usecase.PineappleInteractorIF, error) {
				return usecase.RestorePineappleInteractor(data, new(presenter.PineappleWebPresenter))
			},
			controller.NewPineappleWebControllerWithProvider,
		)
	})
	bind("hearts", func(mux *http.ServeMux) error {
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
	bind("memory", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/memory/exec", "memory:",
			func() usecase.MemoryInteractorIF {
				return usecase.NewMemoryInteractor(domain.NewDefaultMemory(), new(presenter.MemoryWebPresenter))
			},
			func(data []byte) (usecase.MemoryInteractorIF, error) {
				return usecase.RestoreMemoryInteractor(data, new(presenter.MemoryWebPresenter))
			},
			controller.NewMemoryWebControllerWithProvider,
		)
	})
	bind("klondike", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/klondike/exec", "klondike:",
			func() usecase.KlondikeInteractorIF {
				return usecase.NewKlondikeInteractor(domain.NewDefaultKlondike(), new(presenter.KlondikeWebPresenter))
			},
			func(data []byte) (usecase.KlondikeInteractorIF, error) {
				return usecase.RestoreKlondikeInteractor(data, new(presenter.KlondikeWebPresenter))
			},
			controller.NewKlondikeWebControllerWithProvider,
		)
	})
	bind("freecell", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/freecell/exec", "freecell:",
			func() usecase.FreeCellInteractorIF {
				return usecase.NewFreeCellInteractor(domain.NewDefaultFreeCell(), new(presenter.FreeCellWebPresenter))
			},
			func(data []byte) (usecase.FreeCellInteractorIF, error) {
				return usecase.RestoreFreeCellInteractor(data, new(presenter.FreeCellWebPresenter))
			},
			controller.NewFreeCellWebControllerWithProvider,
		)
	})
	bind("baccarat", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/baccarat/exec", "baccarat:",
			func() usecase.BaccaratInteractorIF {
				return usecase.NewBaccaratInteractor(domain.NewDefaultBaccarat(), new(presenter.BaccaratWebPresenter))
			},
			func(data []byte) (usecase.BaccaratInteractorIF, error) {
				return usecase.RestoreBaccaratInteractor(data, new(presenter.BaccaratWebPresenter))
			},
			controller.NewBaccaratWebControllerWithProvider,
		)
	})
	bind("spades", func(mux *http.ServeMux) error {
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
	bind("crazyeights", func(mux *http.ServeMux) error {
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
	bind("ginrummy", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/ginrummy/exec", "ginrummy:",
			func() usecase.GinRummyInteractorIF {
				return usecase.NewGinRummyInteractor(domain.NewDefaultGinRummy(), new(presenter.GinRummyWebPresenter))
			},
			func(data []byte) (usecase.GinRummyInteractorIF, error) {
				return usecase.RestoreGinRummyInteractor(data, new(presenter.GinRummyWebPresenter))
			},
			controller.NewGinRummyWebControllerWithProvider,
		)
	})
	bind("canasta", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/canasta/exec", "canasta:",
			func() usecase.CanastaInteractorIF {
				return usecase.NewCanastaInteractor(domain.NewDefaultCanasta(), new(presenter.CanastaWebPresenter))
			},
			func(data []byte) (usecase.CanastaInteractorIF, error) {
				return usecase.RestoreCanastaInteractor(data, new(presenter.CanastaWebPresenter))
			},
			controller.NewCanastaWebControllerWithProvider,
		)
	})
	bind("spider", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/spider/exec", "spider:",
			func() usecase.SpiderInteractorIF {
				return usecase.NewSpiderInteractor(domain.NewDefaultSpider(), new(presenter.SpiderWebPresenter))
			},
			func(data []byte) (usecase.SpiderInteractorIF, error) {
				return usecase.RestoreSpiderInteractor(data, new(presenter.SpiderWebPresenter))
			},
			controller.NewSpiderWebControllerWithProvider,
		)
	})
	bind("napoleon", func(mux *http.ServeMux) error {
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
	bind("indianpoker", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/indianpoker/exec", "indianpoker:",
			func() usecase.IndianPokerInteractorIF {
				return usecase.NewIndianPokerInteractor(domain.NewDefaultIndianPoker(), new(presenter.IndianPokerWebPresenter))
			},
			func(data []byte) (usecase.IndianPokerInteractorIF, error) {
				return usecase.RestoreIndianPokerInteractor(data, new(presenter.IndianPokerWebPresenter))
			},
			controller.NewIndianPokerWebControllerWithProvider,
		)
	})
	bind("videopoker", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/videopoker/exec", "videopoker:",
			func() usecase.VideoPokerInteractorIF {
				return usecase.NewVideoPokerInteractor(domain.NewDefaultVideoPoker(), new(presenter.VideoPokerWebPresenter))
			},
			func(data []byte) (usecase.VideoPokerInteractorIF, error) {
				return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
			},
			controller.NewVideoPokerWebControllerWithProvider,
		)
	})
	bind("deuceswild", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/deuceswild/exec", "deuceswild:",
			func() usecase.VideoPokerInteractorIF {
				return usecase.NewVideoPokerInteractor(domain.NewDeucesWildVideoPoker(), new(presenter.VideoPokerWebPresenter))
			},
			func(data []byte) (usecase.VideoPokerInteractorIF, error) {
				return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
			},
			controller.NewVideoPokerWebControllerWithProvider,
		)
	})
	bind("jokerpoker", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/jokerpoker/exec", "jokerpoker:",
			func() usecase.VideoPokerInteractorIF {
				return usecase.NewVideoPokerInteractor(domain.NewJokerPokerVideoPoker(), new(presenter.VideoPokerWebPresenter))
			},
			func(data []byte) (usecase.VideoPokerInteractorIF, error) {
				return usecase.RestoreVideoPokerInteractor(data, new(presenter.VideoPokerWebPresenter))
			},
			controller.NewVideoPokerWebControllerWithProvider,
		)
	})
	bind("euchre", func(mux *http.ServeMux) error {
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
	bind("pyramid", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/pyramid/exec", "pyramid:",
			func() usecase.PyramidInteractorIF {
				return usecase.NewPyramidInteractor(domain.NewDefaultPyramid(), new(presenter.PyramidWebPresenter))
			},
			func(data []byte) (usecase.PyramidInteractorIF, error) {
				return usecase.RestorePyramidInteractor(data, new(presenter.PyramidWebPresenter))
			},
			controller.NewPyramidWebControllerWithProvider,
		)
	})
	bind("tripeaks", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/tripeaks/exec", "tripeaks:",
			func() usecase.TriPeaksInteractorIF {
				return usecase.NewTriPeaksInteractor(domain.NewDefaultTriPeaks(), new(presenter.TriPeaksWebPresenter))
			},
			func(data []byte) (usecase.TriPeaksInteractorIF, error) {
				return usecase.RestoreTriPeaksInteractor(data, new(presenter.TriPeaksWebPresenter))
			},
			controller.NewTriPeaksWebControllerWithProvider,
		)
	})
	bind("cribbage", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/cribbage/exec", "cribbage:",
			func() usecase.CribbageInteractorIF {
				return usecase.NewCribbageInteractor(domain.NewDefaultCribbage(), new(presenter.CribbageWebPresenter))
			},
			func(data []byte) (usecase.CribbageInteractorIF, error) {
				return usecase.RestoreCribbageInteractor(data, new(presenter.CribbageWebPresenter))
			},
			controller.NewCribbageWebControllerWithProvider,
		)
	})
	bind("threecard", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/threecard/exec", "threecard:",
			func() usecase.ThreeCardInteractorIF {
				return usecase.NewThreeCardInteractor(domain.NewDefaultThreeCard(), new(presenter.ThreeCardWebPresenter))
			},
			func(data []byte) (usecase.ThreeCardInteractorIF, error) {
				return usecase.RestoreThreeCardInteractor(data, new(presenter.ThreeCardWebPresenter))
			},
			controller.NewThreeCardWebControllerWithProvider,
		)
	})
	bind("ohhell", func(mux *http.ServeMux) error {
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
	bind("bridge", func(mux *http.ServeMux) error {
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
	bind("speed", func(mux *http.ServeMux) error {
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
	bind("gofish", func(mux *http.ServeMux) error {
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
	bind("pinochle", func(mux *http.ServeMux) error {
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
	bind("golf", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/golf/exec", "golf:",
			func() usecase.GolfInteractorIF {
				return usecase.NewGolfInteractor(domain.NewDefaultGolf(), new(presenter.GolfWebPresenter))
			},
			func(data []byte) (usecase.GolfInteractorIF, error) {
				return usecase.RestoreGolfInteractor(data, new(presenter.GolfWebPresenter))
			},
			controller.NewGolfWebControllerWithProvider,
		)
	})
	bind("pigtail", func(mux *http.ServeMux) error {
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
	bind("sevencardstud", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/sevencardstud/exec", "sevencardstud:",
			func() usecase.SevenCardStudInteractorIF {
				return usecase.NewSevenCardStudInteractor(domain.NewDefaultSevenCardStud(), new(presenter.SevenCardStudWebPresenter))
			},
			func(data []byte) (usecase.SevenCardStudInteractorIF, error) {
				return usecase.RestoreSevenCardStudInteractor(data, new(presenter.SevenCardStudWebPresenter))
			},
			controller.NewSevenCardStudWebControllerWithProvider,
		)
	})
	bind("clocksolitaire", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/clocksolitaire/exec", "clocksolitaire:",
			func() usecase.ClockSolitaireInteractorIF {
				return usecase.NewClockSolitaireInteractor(domain.NewDefaultClockSolitaire(), new(presenter.ClockSolitaireWebPresenter))
			},
			func(data []byte) (usecase.ClockSolitaireInteractorIF, error) {
				return usecase.RestoreClockSolitaireInteractor(data, new(presenter.ClockSolitaireWebPresenter))
			},
			controller.NewClockSolitaireWebControllerWithProvider,
		)
	})
	bind("durak", func(mux *http.ServeMux) error {
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
	bind("fortythieves", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/fortythieves/exec", "fortythieves:",
			func() usecase.FortyThievesInteractorIF {
				return usecase.NewFortyThievesInteractor(domain.NewDefaultFortyThieves(), new(presenter.FortyThievesWebPresenter))
			},
			func(data []byte) (usecase.FortyThievesInteractorIF, error) {
				return usecase.RestoreFortyThievesInteractor(data, new(presenter.FortyThievesWebPresenter))
			},
			controller.NewFortyThievesWebControllerWithProvider,
		)
	})
	bind("paigow", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/paigow/exec", "paigow:",
			func() usecase.PaiGowInteractorIF {
				return usecase.NewPaiGowInteractor(domain.NewDefaultPaiGow(), new(presenter.PaiGowWebPresenter))
			},
			func(data []byte) (usecase.PaiGowInteractorIF, error) {
				return usecase.RestorePaiGowInteractor(data, new(presenter.PaiGowWebPresenter))
			},
			controller.NewPaiGowWebControllerWithProvider,
		)
	})
	bind("twotenjack", func(mux *http.ServeMux) error {
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
	bind("caribbeanstud", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/caribbeanstud/exec", "caribbeanstud:",
			func() usecase.CaribbeanStudInteractorIF {
				return usecase.NewCaribbeanStudInteractor(domain.NewDefaultCaribbeanStud(), new(presenter.CaribbeanStudWebPresenter))
			},
			func(data []byte) (usecase.CaribbeanStudInteractorIF, error) {
				return usecase.RestoreCaribbeanStudInteractor(data, new(presenter.CaribbeanStudWebPresenter))
			},
			controller.NewCaribbeanStudWebControllerWithProvider,
		)
	})
	bind("war", func(mux *http.ServeMux) error {
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
	bind("canfield", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/canfield/exec", "canfield:",
			func() usecase.CanfieldInteractorIF {
				return usecase.NewCanfieldInteractor(domain.NewDefaultCanfield(), new(presenter.CanfieldWebPresenter))
			},
			func(data []byte) (usecase.CanfieldInteractorIF, error) {
				return usecase.RestoreCanfieldInteractor(data, new(presenter.CanfieldWebPresenter))
			},
			controller.NewCanfieldWebControllerWithProvider,
		)
	})
	bind("fiftyone", func(mux *http.ServeMux) error {
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
	bind("yukon", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/yukon/exec", "yukon:",
			func() usecase.YukonInteractorIF {
				return usecase.NewYukonInteractor(domain.NewDefaultYukon(), new(presenter.YukonWebPresenter))
			},
			func(data []byte) (usecase.YukonInteractorIF, error) {
				return usecase.RestoreYukonInteractor(data, new(presenter.YukonWebPresenter))
			},
			controller.NewYukonWebControllerWithProvider,
		)
	})
	bind("whist", func(mux *http.ServeMux) error {
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
	bind("letitride", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/letitride/exec", "letitride:",
			func() usecase.LetItRideInteractorIF {
				return usecase.NewLetItRideInteractor(domain.NewDefaultLetItRide(), new(presenter.LetItRideWebPresenter))
			},
			func(data []byte) (usecase.LetItRideInteractorIF, error) {
				return usecase.RestoreLetItRideInteractor(data, new(presenter.LetItRideWebPresenter))
			},
			controller.NewLetItRideWebControllerWithProvider,
		)
	})
	bind("pokersquares", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/pokersquares/exec", "pokersquares:",
			func() usecase.PokerSquaresInteractorIF {
				return usecase.NewPokerSquaresInteractor(domain.NewDefaultPokerSquares(), new(presenter.PokerSquaresWebPresenter))
			},
			func(data []byte) (usecase.PokerSquaresInteractorIF, error) {
				return usecase.RestorePokerSquaresInteractor(data, new(presenter.PokerSquaresWebPresenter))
			},
			controller.NewPokerSquaresWebControllerWithProvider,
		)
	})
	bind("pageone", func(mux *http.ServeMux) error {
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
	bind("reddog", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/reddog/exec", "reddog:",
			func() usecase.RedDogInteractorIF {
				return usecase.NewRedDogInteractor(domain.NewDefaultRedDog(), new(presenter.RedDogWebPresenter))
			},
			func(data []byte) (usecase.RedDogInteractorIF, error) {
				return usecase.RestoreRedDogInteractor(data, new(presenter.RedDogWebPresenter))
			},
			controller.NewRedDogWebControllerWithProvider,
		)
	})
	bind("razz", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/razz/exec", "razz:",
			func() usecase.SevenCardStudInteractorIF {
				return usecase.NewSevenCardStudInteractor(domain.NewDefaultRazz(), new(presenter.SevenCardStudWebPresenter))
			},
			func(data []byte) (usecase.SevenCardStudInteractorIF, error) {
				return usecase.RestoreSevenCardStudInteractor(data, new(presenter.SevenCardStudWebPresenter))
			},
			controller.NewSevenCardStudWebControllerWithProvider,
		)
	})
	bind("scorpion", func(mux *http.ServeMux) error {
		return worker.RegisterKV(mux, "/scorpion/exec", "scorpion:",
			func() usecase.ScorpionInteractorIF {
				return usecase.NewScorpionInteractor(domain.NewDefaultScorpion(), new(presenter.ScorpionWebPresenter))
			},
			func(data []byte) (usecase.ScorpionInteractorIF, error) {
				return usecase.RestoreScorpionInteractor(data, new(presenter.ScorpionWebPresenter))
			},
			controller.NewScorpionWebControllerWithProvider,
		)
	})
}
