//go:build !js || !wasm

package games

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// init attaches the HTTP-server-side factory for every game. The build tag
// excludes this file from Cloudflare Worker (TinyGo) binaries so that their
// Web-server-only controllers/dispatchers do not drag in 55 games of code.
func init() {
	BindWebController("blackjack", func() WebController {
		return controller.NewBlackJackWebController(func() usecase.BlackJackInteractorIF {
			return usecase.NewBlackJackInteractor(domain.NewDefaultBlackJack(), new(presenter.BlackJackWebPresenter))
		})
	})
	BindWebController("poker", func() WebController {
		return controller.NewPokerWebController(func() usecase.PokerInteractorIF {
			return usecase.NewPokerInteractor(domain.NewDefaultPoker(), new(presenter.PokerWebPresenter))
		})
	})
	BindWebController("oldmaid", func() WebController {
		return controller.NewOldMaidWebController(func() usecase.OldMaidInteractorIF {
			return usecase.NewOldMaidInteractor(domain.NewDefaultOldMaid(), new(presenter.OldMaidWebPresenter))
		})
	})
	BindWebController("daifugo", func() WebController {
		return controller.NewDaifugoWebController(func() usecase.DaifugoInteractorIF {
			return usecase.NewDaifugoInteractor(domain.NewDefaultDaifugo(), new(presenter.DaifugoWebPresenter))
		})
	})
	BindWebController("sevens", func() WebController {
		return controller.NewSevensWebController(func() usecase.SevensInteractorIF {
			return usecase.NewSevensInteractor(domain.NewDefaultSevens(), new(presenter.SevensWebPresenter))
		})
	})
	BindWebController("doubt", func() WebController {
		return controller.NewDoubtWebController(func() usecase.DoubtInteractorIF {
			return usecase.NewDoubtInteractor(domain.NewDefaultDoubt(), new(presenter.DoubtWebPresenter))
		})
	})
	BindWebController("holdem", func() WebController {
		return controller.NewHoldemWebController(func() usecase.HoldemInteractorIF {
			return usecase.NewHoldemInteractor(domain.NewDefaultHoldem(), new(presenter.HoldemWebPresenter))
		})
	})
	BindWebController("omaha", func() WebController {
		return controller.NewOmahaWebController(func() usecase.OmahaInteractorIF {
			return usecase.NewOmahaInteractor(domain.NewDefaultOmaha(), new(presenter.OmahaWebPresenter))
		})
	})
	BindWebController("shortdeck", func() WebController {
		return controller.NewShortDeckWebController(func() usecase.ShortDeckInteractorIF {
			return usecase.NewShortDeckInteractor(domain.NewDefaultShortDeck(), new(presenter.ShortDeckWebPresenter))
		})
	})
	BindWebController("pineapple", func() WebController {
		return controller.NewPineappleWebController(func() usecase.PineappleInteractorIF {
			return usecase.NewPineappleInteractor(domain.NewDefaultPineapple(), new(presenter.PineappleWebPresenter))
		})
	})
	BindWebController("hearts", func() WebController {
		return controller.NewHeartsWebController(func() usecase.HeartsInteractorIF {
			return usecase.NewHeartsInteractor(domain.NewDefaultHearts(), new(presenter.HeartsWebPresenter))
		})
	})
	BindWebController("memory", func() WebController {
		return controller.NewMemoryWebController(func() usecase.MemoryInteractorIF {
			return usecase.NewMemoryInteractor(domain.NewDefaultMemory(), new(presenter.MemoryWebPresenter))
		})
	})
	BindWebController("klondike", func() WebController {
		return controller.NewKlondikeWebController(func() usecase.KlondikeInteractorIF {
			return usecase.NewKlondikeInteractor(domain.NewDefaultKlondike(), new(presenter.KlondikeWebPresenter))
		})
	})
	BindWebController("freecell", func() WebController {
		return controller.NewFreeCellWebController(func() usecase.FreeCellInteractorIF {
			return usecase.NewFreeCellInteractor(domain.NewDefaultFreeCell(), new(presenter.FreeCellWebPresenter))
		})
	})
	BindWebController("baccarat", func() WebController {
		return controller.NewBaccaratWebController(func() usecase.BaccaratInteractorIF {
			return usecase.NewBaccaratInteractor(domain.NewDefaultBaccarat(), new(presenter.BaccaratWebPresenter))
		})
	})
	BindWebController("spades", func() WebController {
		return controller.NewSpadesWebController(func() usecase.SpadesInteractorIF {
			return usecase.NewSpadesInteractor(domain.NewDefaultSpades(), new(presenter.SpadesWebPresenter))
		})
	})
	BindWebController("crazyeights", func() WebController {
		return controller.NewCrazyEightsWebController(func() usecase.CrazyEightsInteractorIF {
			return usecase.NewCrazyEightsInteractor(domain.NewDefaultCrazyEights(), new(presenter.CrazyEightsWebPresenter))
		})
	})
	BindWebController("ginrummy", func() WebController {
		return controller.NewGinRummyWebController(func() usecase.GinRummyInteractorIF {
			return usecase.NewGinRummyInteractor(domain.NewDefaultGinRummy(), new(presenter.GinRummyWebPresenter))
		})
	})
	BindWebController("canasta", func() WebController {
		return controller.NewCanastaWebController(func() usecase.CanastaInteractorIF {
			return usecase.NewCanastaInteractor(domain.NewDefaultCanasta(), new(presenter.CanastaWebPresenter))
		})
	})
	BindWebController("spider", func() WebController {
		return controller.NewSpiderWebController(func() usecase.SpiderInteractorIF {
			return usecase.NewSpiderInteractor(domain.NewDefaultSpider(), new(presenter.SpiderWebPresenter))
		})
	})
	BindWebController("napoleon", func() WebController {
		return controller.NewNapoleonWebController(func() usecase.NapoleonInteractorIF {
			return usecase.NewNapoleonInteractor(domain.NewDefaultNapoleon(), new(presenter.NapoleonWebPresenter))
		})
	})
	BindWebController("indianpoker", func() WebController {
		return controller.NewIndianPokerWebController(func() usecase.IndianPokerInteractorIF {
			return usecase.NewIndianPokerInteractor(domain.NewDefaultIndianPoker(), new(presenter.IndianPokerWebPresenter))
		})
	})
	BindWebController("videopoker", func() WebController {
		return controller.NewVideoPokerWebController(func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(domain.NewDefaultVideoPoker(), new(presenter.VideoPokerWebPresenter))
		})
	})
	BindWebController("deuceswild", func() WebController {
		return controller.NewVideoPokerWebController(func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(domain.NewDeucesWildVideoPoker(), new(presenter.VideoPokerWebPresenter))
		})
	})
	BindWebController("jokerpoker", func() WebController {
		return controller.NewVideoPokerWebController(func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(domain.NewJokerPokerVideoPoker(), new(presenter.VideoPokerWebPresenter))
		})
	})
	BindWebController("euchre", func() WebController {
		return controller.NewEuchreWebController(func() usecase.EuchreInteractorIF {
			return usecase.NewEuchreInteractor(domain.NewDefaultEuchre(), new(presenter.EuchreWebPresenter))
		})
	})
	BindWebController("pyramid", func() WebController {
		return controller.NewPyramidWebController(func() usecase.PyramidInteractorIF {
			return usecase.NewPyramidInteractor(domain.NewDefaultPyramid(), new(presenter.PyramidWebPresenter))
		})
	})
	BindWebController("tripeaks", func() WebController {
		return controller.NewTriPeaksWebController(func() usecase.TriPeaksInteractorIF {
			return usecase.NewTriPeaksInteractor(domain.NewDefaultTriPeaks(), new(presenter.TriPeaksWebPresenter))
		})
	})
	BindWebController("cribbage", func() WebController {
		return controller.NewCribbageWebController(func() usecase.CribbageInteractorIF {
			return usecase.NewCribbageInteractor(domain.NewDefaultCribbage(), new(presenter.CribbageWebPresenter))
		})
	})
	BindWebController("threecard", func() WebController {
		return controller.NewThreeCardWebController(func() usecase.ThreeCardInteractorIF {
			return usecase.NewThreeCardInteractor(domain.NewDefaultThreeCard(), new(presenter.ThreeCardWebPresenter))
		})
	})
	BindWebController("ohhell", func() WebController {
		return controller.NewOhHellWebController(func() usecase.OhHellInteractorIF {
			return usecase.NewOhHellInteractor(domain.NewDefaultOhHell(), new(presenter.OhHellWebPresenter))
		})
	})
	BindWebController("bridge", func() WebController {
		return controller.NewBridgeWebController(func() usecase.BridgeInteractorIF {
			return usecase.NewBridgeInteractor(domain.NewDefaultBridge(), new(presenter.BridgeWebPresenter))
		})
	})
	BindWebController("speed", func() WebController {
		return controller.NewSpeedWebController(func() usecase.SpeedInteractorIF {
			return usecase.NewSpeedInteractor(domain.NewDefaultSpeed(), new(presenter.SpeedWebPresenter))
		})
	})
	BindWebController("gofish", func() WebController {
		return controller.NewGoFishWebController(func() usecase.GoFishInteractorIF {
			return usecase.NewGoFishInteractor(domain.NewDefaultGoFish(), new(presenter.GoFishWebPresenter))
		})
	})
	BindWebController("pinochle", func() WebController {
		return controller.NewPinochleWebController(func() usecase.PinochleInteractorIF {
			return usecase.NewPinochleInteractor(domain.NewDefaultPinochle(), new(presenter.PinochleWebPresenter))
		})
	})
	BindWebController("golf", func() WebController {
		return controller.NewGolfWebController(func() usecase.GolfInteractorIF {
			return usecase.NewGolfInteractor(domain.NewDefaultGolf(), new(presenter.GolfWebPresenter))
		})
	})
	BindWebController("pigtail", func() WebController {
		return controller.NewPigsTailWebController(func() usecase.PigsTailInteractorIF {
			return usecase.NewPigsTailInteractor(domain.NewDefaultPigsTail(), new(presenter.PigsTailWebPresenter))
		})
	})
	BindWebController("sevencardstud", func() WebController {
		return controller.NewSevenCardStudWebController(func() usecase.SevenCardStudInteractorIF {
			return usecase.NewSevenCardStudInteractor(domain.NewDefaultSevenCardStud(), new(presenter.SevenCardStudWebPresenter))
		})
	})
	BindWebController("clocksolitaire", func() WebController {
		return controller.NewClockSolitaireWebController(func() usecase.ClockSolitaireInteractorIF {
			return usecase.NewClockSolitaireInteractor(domain.NewDefaultClockSolitaire(), new(presenter.ClockSolitaireWebPresenter))
		})
	})
	BindWebController("durak", func() WebController {
		return controller.NewDurakWebController(func() usecase.DurakInteractorIF {
			return usecase.NewDurakInteractor(domain.NewDefaultDurak(), new(presenter.DurakWebPresenter))
		})
	})
	BindWebController("fortythieves", func() WebController {
		return controller.NewFortyThievesWebController(func() usecase.FortyThievesInteractorIF {
			return usecase.NewFortyThievesInteractor(domain.NewDefaultFortyThieves(), new(presenter.FortyThievesWebPresenter))
		})
	})
	BindWebController("paigow", func() WebController {
		return controller.NewPaiGowWebController(func() usecase.PaiGowInteractorIF {
			return usecase.NewPaiGowInteractor(domain.NewDefaultPaiGow(), new(presenter.PaiGowWebPresenter))
		})
	})
	BindWebController("twotenjack", func() WebController {
		return controller.NewTwoTenJackWebController(func() usecase.TwoTenJackInteractorIF {
			return usecase.NewTwoTenJackInteractor(domain.NewDefaultTwoTenJack(), new(presenter.TwoTenJackWebPresenter))
		})
	})
	BindWebController("caribbeanstud", func() WebController {
		return controller.NewCaribbeanStudWebController(func() usecase.CaribbeanStudInteractorIF {
			return usecase.NewCaribbeanStudInteractor(domain.NewDefaultCaribbeanStud(), new(presenter.CaribbeanStudWebPresenter))
		})
	})
	BindWebController("war", func() WebController {
		return controller.NewWarWebController(func() usecase.WarInteractorIF {
			return usecase.NewWarInteractor(domain.NewDefaultWar(), new(presenter.WarWebPresenter))
		})
	})
	BindWebController("canfield", func() WebController {
		return controller.NewCanfieldWebController(func() usecase.CanfieldInteractorIF {
			return usecase.NewCanfieldInteractor(domain.NewDefaultCanfield(), new(presenter.CanfieldWebPresenter))
		})
	})
	BindWebController("fiftyone", func() WebController {
		return controller.NewFiftyOneWebController(func() usecase.FiftyOneInteractorIF {
			return usecase.NewFiftyOneInteractor(domain.NewDefaultFiftyOne(), new(presenter.FiftyOneWebPresenter))
		})
	})
	BindWebController("yukon", func() WebController {
		return controller.NewYukonWebController(func() usecase.YukonInteractorIF {
			return usecase.NewYukonInteractor(domain.NewDefaultYukon(), new(presenter.YukonWebPresenter))
		})
	})
	BindWebController("whist", func() WebController {
		return controller.NewWhistWebController(func() usecase.WhistInteractorIF {
			return usecase.NewWhistInteractor(domain.NewDefaultWhist(), new(presenter.WhistWebPresenter))
		})
	})
	BindWebController("letitride", func() WebController {
		return controller.NewLetItRideWebController(func() usecase.LetItRideInteractorIF {
			return usecase.NewLetItRideInteractor(domain.NewDefaultLetItRide(), new(presenter.LetItRideWebPresenter))
		})
	})
	BindWebController("pokersquares", func() WebController {
		return controller.NewPokerSquaresWebController(func() usecase.PokerSquaresInteractorIF {
			return usecase.NewPokerSquaresInteractor(domain.NewDefaultPokerSquares(), new(presenter.PokerSquaresWebPresenter))
		})
	})
	BindWebController("pageone", func() WebController {
		return controller.NewPageOneWebController(func() usecase.PageOneInteractorIF {
			return usecase.NewPageOneInteractor(domain.NewDefaultPageOne(), new(presenter.PageOneWebPresenter))
		})
	})
	BindWebController("reddog", func() WebController {
		return controller.NewRedDogWebController(func() usecase.RedDogInteractorIF {
			return usecase.NewRedDogInteractor(domain.NewDefaultRedDog(), new(presenter.RedDogWebPresenter))
		})
	})
	BindWebController("razz", func() WebController {
		return controller.NewSevenCardStudWebController(func() usecase.SevenCardStudInteractorIF {
			return usecase.NewSevenCardStudInteractor(domain.NewDefaultRazz(), new(presenter.SevenCardStudWebPresenter))
		})
	})
	BindWebController("scorpion", func() WebController {
		return controller.NewScorpionWebController(func() usecase.ScorpionInteractorIF {
			return usecase.NewScorpionInteractor(domain.NewDefaultScorpion(), new(presenter.ScorpionWebPresenter))
		})
	})
}
