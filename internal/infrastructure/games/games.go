package games

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// registry is the single source of truth for all games. The order mirrors
// internal/infrastructure/ui/GameManager.go so that CLI, HTTP, and Worker
// surfaces share one canonical ordering.
//
// Adding a game requires:
//  1. Appending a Game entry here with its Web factory and Category.
//  2. Adding a matching RegisterWorker closure in games_wasm.go (or the
//     worker registration will be nil and TestWorkersPopulated will fail).
var registry = []*Game{
	{
		Name:     "blackjack",
		Category: CategoryCasino,
		NewWebController: func() WebController {
			return controller.NewBlackJackWebController(func() usecase.BlackJackInteractorIF {
				return usecase.NewBlackJackInteractor(domain.NewDefaultBlackJack(), new(presenter.BlackJackWebPresenter))
			})
		},
	},
	{
		Name:     "poker",
		Category: CategoryCasino,
		NewWebController: func() WebController {
			return controller.NewPokerWebController(func() usecase.PokerInteractorIF {
				return usecase.NewPokerInteractor(domain.NewDefaultPoker(), new(presenter.PokerWebPresenter))
			})
		},
	},
	{
		Name:     "oldmaid",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewOldMaidWebController(func() usecase.OldMaidInteractorIF {
				return usecase.NewOldMaidInteractor(domain.NewDefaultOldMaid(), new(presenter.OldMaidWebPresenter))
			})
		},
	},
	{
		Name:     "daifugo",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewDaifugoWebController(func() usecase.DaifugoInteractorIF {
				return usecase.NewDaifugoInteractor(domain.NewDefaultDaifugo(), new(presenter.DaifugoWebPresenter))
			})
		},
	},
	{
		Name:     "sevens",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewSevensWebController(func() usecase.SevensInteractorIF {
				return usecase.NewSevensInteractor(domain.NewDefaultSevens(), new(presenter.SevensWebPresenter))
			})
		},
	},
	{
		Name:     "doubt",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewDoubtWebController(func() usecase.DoubtInteractorIF {
				return usecase.NewDoubtInteractor(domain.NewDefaultDoubt(), new(presenter.DoubtWebPresenter))
			})
		},
	},
	{
		Name:     "holdem",
		Category: CategoryCasino,
		NewWebController: func() WebController {
			return controller.NewHoldemWebController(func() usecase.HoldemInteractorIF {
				return usecase.NewHoldemInteractor(domain.NewDefaultHoldem(), new(presenter.HoldemWebPresenter))
			})
		},
	},
	{
		Name:     "omaha",
		Category: CategoryCasino,
		NewWebController: func() WebController {
			return controller.NewOmahaWebController(func() usecase.OmahaInteractorIF {
				return usecase.NewOmahaInteractor(domain.NewDefaultOmaha(), new(presenter.OmahaWebPresenter))
			})
		},
	},
	{
		Name:     "shortdeck",
		Category: CategoryCasino,
		NewWebController: func() WebController {
			return controller.NewShortDeckWebController(func() usecase.ShortDeckInteractorIF {
				return usecase.NewShortDeckInteractor(domain.NewDefaultShortDeck(), new(presenter.ShortDeckWebPresenter))
			})
		},
	},
	{
		Name:     "pineapple",
		Category: CategoryCasino,
		NewWebController: func() WebController {
			return controller.NewPineappleWebController(func() usecase.PineappleInteractorIF {
				return usecase.NewPineappleInteractor(domain.NewDefaultPineapple(), new(presenter.PineappleWebPresenter))
			})
		},
	},
	{
		Name:     "hearts",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewHeartsWebController(func() usecase.HeartsInteractorIF {
				return usecase.NewHeartsInteractor(domain.NewDefaultHearts(), new(presenter.HeartsWebPresenter))
			})
		},
	},
	{
		Name:     "memory",
		Category: CategorySolo,
		NewWebController: func() WebController {
			return controller.NewMemoryWebController(func() usecase.MemoryInteractorIF {
				return usecase.NewMemoryInteractor(domain.NewDefaultMemory(), new(presenter.MemoryWebPresenter))
			})
		},
	},
	{
		Name:     "klondike",
		Category: CategorySolo,
		NewWebController: func() WebController {
			return controller.NewKlondikeWebController(func() usecase.KlondikeInteractorIF {
				return usecase.NewKlondikeInteractor(domain.NewDefaultKlondike(), new(presenter.KlondikeWebPresenter))
			})
		},
	},
	{
		Name:     "freecell",
		Category: CategorySolo,
		NewWebController: func() WebController {
			return controller.NewFreeCellWebController(func() usecase.FreeCellInteractorIF {
				return usecase.NewFreeCellInteractor(domain.NewDefaultFreeCell(), new(presenter.FreeCellWebPresenter))
			})
		},
	},
	{
		Name:     "baccarat",
		Category: CategoryCasino,
		NewWebController: func() WebController {
			return controller.NewBaccaratWebController(func() usecase.BaccaratInteractorIF {
				return usecase.NewBaccaratInteractor(domain.NewDefaultBaccarat(), new(presenter.BaccaratWebPresenter))
			})
		},
	},
	{
		Name:     "spades",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewSpadesWebController(func() usecase.SpadesInteractorIF {
				return usecase.NewSpadesInteractor(domain.NewDefaultSpades(), new(presenter.SpadesWebPresenter))
			})
		},
	},
	{
		Name:     "crazyeights",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewCrazyEightsWebController(func() usecase.CrazyEightsInteractorIF {
				return usecase.NewCrazyEightsInteractor(domain.NewDefaultCrazyEights(), new(presenter.CrazyEightsWebPresenter))
			})
		},
	},
	{
		Name:     "ginrummy",
		Category: CategorySolo,
		NewWebController: func() WebController {
			return controller.NewGinRummyWebController(func() usecase.GinRummyInteractorIF {
				return usecase.NewGinRummyInteractor(domain.NewDefaultGinRummy(), new(presenter.GinRummyWebPresenter))
			})
		},
	},
	{
		Name:     "canasta",
		Category: CategorySolo,
		NewWebController: func() WebController {
			return controller.NewCanastaWebController(func() usecase.CanastaInteractorIF {
				return usecase.NewCanastaInteractor(domain.NewDefaultCanasta(), new(presenter.CanastaWebPresenter))
			})
		},
	},
	{
		Name:     "spider",
		Category: CategorySolo,
		NewWebController: func() WebController {
			return controller.NewSpiderWebController(func() usecase.SpiderInteractorIF {
				return usecase.NewSpiderInteractor(domain.NewDefaultSpider(), new(presenter.SpiderWebPresenter))
			})
		},
	},
	{
		Name:     "napoleon",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewNapoleonWebController(func() usecase.NapoleonInteractorIF {
				return usecase.NewNapoleonInteractor(domain.NewDefaultNapoleon(), new(presenter.NapoleonWebPresenter))
			})
		},
	},
	{
		Name:     "indianpoker",
		Category: CategoryCasino,
		NewWebController: func() WebController {
			return controller.NewIndianPokerWebController(func() usecase.IndianPokerInteractorIF {
				return usecase.NewIndianPokerInteractor(domain.NewDefaultIndianPoker(), new(presenter.IndianPokerWebPresenter))
			})
		},
	},
	{
		Name:     "videopoker",
		Category: CategoryCasino,
		NewWebController: func() WebController {
			return controller.NewVideoPokerWebController(func() usecase.VideoPokerInteractorIF {
				return usecase.NewVideoPokerInteractor(domain.NewDefaultVideoPoker(), new(presenter.VideoPokerWebPresenter))
			})
		},
	},
	{
		Name:     "deuceswild",
		Category: CategoryCasino,
		NewWebController: func() WebController {
			return controller.NewVideoPokerWebController(func() usecase.VideoPokerInteractorIF {
				return usecase.NewVideoPokerInteractor(domain.NewDeucesWildVideoPoker(), new(presenter.VideoPokerWebPresenter))
			})
		},
	},
	{
		Name:     "jokerpoker",
		Category: CategoryCasino,
		NewWebController: func() WebController {
			return controller.NewVideoPokerWebController(func() usecase.VideoPokerInteractorIF {
				return usecase.NewVideoPokerInteractor(domain.NewJokerPokerVideoPoker(), new(presenter.VideoPokerWebPresenter))
			})
		},
	},
	{
		Name:     "euchre",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewEuchreWebController(func() usecase.EuchreInteractorIF {
				return usecase.NewEuchreInteractor(domain.NewDefaultEuchre(), new(presenter.EuchreWebPresenter))
			})
		},
	},
	{
		Name:     "pyramid",
		Category: CategorySolo,
		NewWebController: func() WebController {
			return controller.NewPyramidWebController(func() usecase.PyramidInteractorIF {
				return usecase.NewPyramidInteractor(domain.NewDefaultPyramid(), new(presenter.PyramidWebPresenter))
			})
		},
	},
	{
		Name:     "tripeaks",
		Category: CategorySolo,
		NewWebController: func() WebController {
			return controller.NewTriPeaksWebController(func() usecase.TriPeaksInteractorIF {
				return usecase.NewTriPeaksInteractor(domain.NewDefaultTriPeaks(), new(presenter.TriPeaksWebPresenter))
			})
		},
	},
	{
		Name:     "cribbage",
		Category: CategorySolo,
		NewWebController: func() WebController {
			return controller.NewCribbageWebController(func() usecase.CribbageInteractorIF {
				return usecase.NewCribbageInteractor(domain.NewDefaultCribbage(), new(presenter.CribbageWebPresenter))
			})
		},
	},
	{
		Name:     "threecard",
		Category: CategoryCasino,
		NewWebController: func() WebController {
			return controller.NewThreeCardWebController(func() usecase.ThreeCardInteractorIF {
				return usecase.NewThreeCardInteractor(domain.NewDefaultThreeCard(), new(presenter.ThreeCardWebPresenter))
			})
		},
	},
	{
		Name:     "ohhell",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewOhHellWebController(func() usecase.OhHellInteractorIF {
				return usecase.NewOhHellInteractor(domain.NewDefaultOhHell(), new(presenter.OhHellWebPresenter))
			})
		},
	},
	{
		Name:     "bridge",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewBridgeWebController(func() usecase.BridgeInteractorIF {
				return usecase.NewBridgeInteractor(domain.NewDefaultBridge(), new(presenter.BridgeWebPresenter))
			})
		},
	},
	{
		Name:     "speed",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewSpeedWebController(func() usecase.SpeedInteractorIF {
				return usecase.NewSpeedInteractor(domain.NewDefaultSpeed(), new(presenter.SpeedWebPresenter))
			})
		},
	},
	{
		Name:     "gofish",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewGoFishWebController(func() usecase.GoFishInteractorIF {
				return usecase.NewGoFishInteractor(domain.NewDefaultGoFish(), new(presenter.GoFishWebPresenter))
			})
		},
	},
	{
		Name:     "pinochle",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewPinochleWebController(func() usecase.PinochleInteractorIF {
				return usecase.NewPinochleInteractor(domain.NewDefaultPinochle(), new(presenter.PinochleWebPresenter))
			})
		},
	},
	{
		Name:     "golf",
		Category: CategorySolo,
		NewWebController: func() WebController {
			return controller.NewGolfWebController(func() usecase.GolfInteractorIF {
				return usecase.NewGolfInteractor(domain.NewDefaultGolf(), new(presenter.GolfWebPresenter))
			})
		},
	},
	{
		Name:     "pigtail",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewPigsTailWebController(func() usecase.PigsTailInteractorIF {
				return usecase.NewPigsTailInteractor(domain.NewDefaultPigsTail(), new(presenter.PigsTailWebPresenter))
			})
		},
	},
	{
		Name:     "sevencardstud",
		Category: CategoryCasino,
		NewWebController: func() WebController {
			return controller.NewSevenCardStudWebController(func() usecase.SevenCardStudInteractorIF {
				return usecase.NewSevenCardStudInteractor(domain.NewDefaultSevenCardStud(), new(presenter.SevenCardStudWebPresenter))
			})
		},
	},
	{
		Name:     "clocksolitaire",
		Category: CategorySolo,
		NewWebController: func() WebController {
			return controller.NewClockSolitaireWebController(func() usecase.ClockSolitaireInteractorIF {
				return usecase.NewClockSolitaireInteractor(domain.NewDefaultClockSolitaire(), new(presenter.ClockSolitaireWebPresenter))
			})
		},
	},
	{
		Name:     "durak",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewDurakWebController(func() usecase.DurakInteractorIF {
				return usecase.NewDurakInteractor(domain.NewDefaultDurak(), new(presenter.DurakWebPresenter))
			})
		},
	},
	{
		Name:     "fortythieves",
		Category: CategorySolo,
		NewWebController: func() WebController {
			return controller.NewFortyThievesWebController(func() usecase.FortyThievesInteractorIF {
				return usecase.NewFortyThievesInteractor(domain.NewDefaultFortyThieves(), new(presenter.FortyThievesWebPresenter))
			})
		},
	},
	{
		Name:     "paigow",
		Category: CategoryCasino,
		NewWebController: func() WebController {
			return controller.NewPaiGowWebController(func() usecase.PaiGowInteractorIF {
				return usecase.NewPaiGowInteractor(domain.NewDefaultPaiGow(), new(presenter.PaiGowWebPresenter))
			})
		},
	},
	{
		Name:     "twotenjack",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewTwoTenJackWebController(func() usecase.TwoTenJackInteractorIF {
				return usecase.NewTwoTenJackInteractor(domain.NewDefaultTwoTenJack(), new(presenter.TwoTenJackWebPresenter))
			})
		},
	},
	{
		Name:     "caribbeanstud",
		Category: CategoryCasino,
		NewWebController: func() WebController {
			return controller.NewCaribbeanStudWebController(func() usecase.CaribbeanStudInteractorIF {
				return usecase.NewCaribbeanStudInteractor(domain.NewDefaultCaribbeanStud(), new(presenter.CaribbeanStudWebPresenter))
			})
		},
	},
	{
		Name:     "war",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewWarWebController(func() usecase.WarInteractorIF {
				return usecase.NewWarInteractor(domain.NewDefaultWar(), new(presenter.WarWebPresenter))
			})
		},
	},
	{
		Name:     "canfield",
		Category: CategorySolo,
		NewWebController: func() WebController {
			return controller.NewCanfieldWebController(func() usecase.CanfieldInteractorIF {
				return usecase.NewCanfieldInteractor(domain.NewDefaultCanfield(), new(presenter.CanfieldWebPresenter))
			})
		},
	},
	{
		Name:     "fiftyone",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewFiftyOneWebController(func() usecase.FiftyOneInteractorIF {
				return usecase.NewFiftyOneInteractor(domain.NewDefaultFiftyOne(), new(presenter.FiftyOneWebPresenter))
			})
		},
	},
	{
		Name:     "yukon",
		Category: CategorySolo,
		NewWebController: func() WebController {
			return controller.NewYukonWebController(func() usecase.YukonInteractorIF {
				return usecase.NewYukonInteractor(domain.NewDefaultYukon(), new(presenter.YukonWebPresenter))
			})
		},
	},
	{
		Name:     "whist",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewWhistWebController(func() usecase.WhistInteractorIF {
				return usecase.NewWhistInteractor(domain.NewDefaultWhist(), new(presenter.WhistWebPresenter))
			})
		},
	},
	{
		Name:     "letitride",
		Category: CategoryCasino,
		NewWebController: func() WebController {
			return controller.NewLetItRideWebController(func() usecase.LetItRideInteractorIF {
				return usecase.NewLetItRideInteractor(domain.NewDefaultLetItRide(), new(presenter.LetItRideWebPresenter))
			})
		},
	},
	{
		Name:     "pokersquares",
		Category: CategorySolo,
		NewWebController: func() WebController {
			return controller.NewPokerSquaresWebController(func() usecase.PokerSquaresInteractorIF {
				return usecase.NewPokerSquaresInteractor(domain.NewDefaultPokerSquares(), new(presenter.PokerSquaresWebPresenter))
			})
		},
	},
	{
		Name:     "pageone",
		Category: CategoryClassic,
		NewWebController: func() WebController {
			return controller.NewPageOneWebController(func() usecase.PageOneInteractorIF {
				return usecase.NewPageOneInteractor(domain.NewDefaultPageOne(), new(presenter.PageOneWebPresenter))
			})
		},
	},
	{
		Name:     "reddog",
		Category: CategoryCasino,
		NewWebController: func() WebController {
			return controller.NewRedDogWebController(func() usecase.RedDogInteractorIF {
				return usecase.NewRedDogInteractor(domain.NewDefaultRedDog(), new(presenter.RedDogWebPresenter))
			})
		},
	},
	{
		Name:     "razz",
		Category: CategoryCasino,
		NewWebController: func() WebController {
			return controller.NewSevenCardStudWebController(func() usecase.SevenCardStudInteractorIF {
				return usecase.NewSevenCardStudInteractor(domain.NewDefaultRazz(), new(presenter.SevenCardStudWebPresenter))
			})
		},
	},
	{
		Name:     "scorpion",
		Category: CategorySolo,
		NewWebController: func() WebController {
			return controller.NewScorpionWebController(func() usecase.ScorpionInteractorIF {
				return usecase.NewScorpionInteractor(domain.NewDefaultScorpion(), new(presenter.ScorpionWebPresenter))
			})
		},
	},
}
