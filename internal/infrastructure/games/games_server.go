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
// Web-server-only controllers/dispatchers do not drag in 219 games of code.
func init() {
	BindWebControllerFor("blackjack",
		func() usecase.BlackJackInteractorIF {
			return usecase.NewBlackJackInteractor(domain.NewDefaultBlackJack(), new(presenter.BlackJackWebPresenter))
		},
		controller.NewBlackJackWebController)
	BindWebControllerFor("poker",
		func() usecase.PokerInteractorIF {
			return usecase.NewPokerInteractor(domain.NewDefaultPoker(), new(presenter.PokerWebPresenter))
		},
		controller.NewPokerWebController)
	BindWebControllerFor("oldmaid",
		func() usecase.OldMaidInteractorIF {
			return usecase.NewOldMaidInteractor(domain.NewDefaultOldMaid(), new(presenter.OldMaidWebPresenter))
		},
		controller.NewOldMaidWebController)
	BindWebControllerFor("daifugo",
		func() usecase.DaifugoInteractorIF {
			return usecase.NewDaifugoInteractor(domain.NewDefaultDaifugo(), new(presenter.DaifugoWebPresenter))
		},
		controller.NewDaifugoWebController)
	BindWebControllerFor("bigtwo",
		func() usecase.BigTwoInteractorIF {
			return usecase.NewBigTwoInteractor(domain.NewDefaultBigTwo(), new(presenter.BigTwoWebPresenter))
		},
		controller.NewBigTwoWebController)
	BindWebControllerFor("sevens",
		func() usecase.SevensInteractorIF {
			return usecase.NewSevensInteractor(domain.NewDefaultSevens(), new(presenter.SevensWebPresenter))
		},
		controller.NewSevensWebController)
	BindWebControllerFor("doubt",
		func() usecase.DoubtInteractorIF {
			return usecase.NewDoubtInteractor(domain.NewDefaultDoubt(), new(presenter.DoubtWebPresenter))
		},
		controller.NewDoubtWebController)
	BindWebControllerFor("holdem",
		func() usecase.HoldemInteractorIF {
			return usecase.NewHoldemInteractor(domain.NewDefaultHoldem(), new(presenter.HoldemWebPresenter))
		},
		controller.NewHoldemWebController)
	BindWebControllerFor("omaha",
		func() usecase.OmahaInteractorIF {
			return usecase.NewOmahaInteractor(domain.NewDefaultOmaha(), new(presenter.OmahaWebPresenter))
		},
		controller.NewOmahaWebController)
	BindWebControllerFor("omahahilo",
		func() usecase.OmahaInteractorIF {
			return usecase.NewOmahaInteractor(domain.NewDefaultOmahaHiLo(), new(presenter.OmahaWebPresenter))
		},
		controller.NewOmahaWebController)
	BindWebControllerFor("bigo",
		func() usecase.OmahaInteractorIF {
			return usecase.NewOmahaInteractor(domain.NewDefaultBigO(), new(presenter.OmahaWebPresenter))
		},
		controller.NewOmahaWebController)
	BindWebControllerFor("courchevel",
		func() usecase.OmahaInteractorIF {
			return usecase.NewOmahaInteractor(domain.NewDefaultCourchevel(), new(presenter.OmahaWebPresenter))
		},
		controller.NewOmahaWebController)
	BindWebControllerFor("bigohilo",
		func() usecase.OmahaInteractorIF {
			return usecase.NewOmahaInteractor(domain.NewDefaultBigOHiLo(), new(presenter.OmahaWebPresenter))
		},
		controller.NewOmahaWebController)
	BindWebControllerFor("shortdeck",
		func() usecase.ShortDeckInteractorIF {
			return usecase.NewShortDeckInteractor(domain.NewDefaultShortDeck(), new(presenter.ShortDeckWebPresenter))
		},
		controller.NewShortDeckWebController)
	BindWebControllerFor("pineapple",
		func() usecase.PineappleInteractorIF {
			return usecase.NewPineappleInteractor(domain.NewDefaultPineapple(), new(presenter.PineappleWebPresenter))
		},
		controller.NewPineappleWebController)
	BindWebControllerFor("crazypineapple",
		func() usecase.PineappleInteractorIF {
			return usecase.NewPineappleInteractor(domain.NewDefaultCrazyPineapple(), new(presenter.PineappleWebPresenter))
		},
		controller.NewPineappleWebController)
	BindWebControllerFor("irishpoker",
		func() usecase.PineappleInteractorIF {
			return usecase.NewPineappleInteractor(domain.NewDefaultIrishPoker(), new(presenter.PineappleWebPresenter))
		},
		controller.NewPineappleWebController)
	BindWebControllerFor("hearts",
		func() usecase.HeartsInteractorIF {
			return usecase.NewHeartsInteractor(domain.NewDefaultHearts(), new(presenter.HeartsWebPresenter))
		},
		controller.NewHeartsWebController)
	BindWebControllerFor("memory",
		func() usecase.MemoryInteractorIF {
			return usecase.NewMemoryInteractor(domain.NewDefaultMemory(), new(presenter.MemoryWebPresenter))
		},
		controller.NewMemoryWebController)
	BindWebControllerFor("whitehead",
		func() usecase.WhiteheadInteractorIF {
			return usecase.NewWhiteheadInteractor(domain.NewDefaultWhitehead(), new(presenter.WhiteheadWebPresenter))
		},
		controller.NewWhiteheadWebController)
	BindWebControllerFor("klondike",
		func() usecase.KlondikeInteractorIF {
			return usecase.NewKlondikeInteractor(domain.NewDefaultKlondike(), new(presenter.KlondikeWebPresenter))
		},
		controller.NewKlondikeWebController)
	BindWebControllerFor("freecell",
		func() usecase.FreeCellInteractorIF {
			return usecase.NewFreeCellInteractor(domain.NewDefaultFreeCell(), new(presenter.FreeCellWebPresenter))
		},
		controller.NewFreeCellWebController)
	// Baker's Game reuses the FreeCell interactor/controller; the same-suit rule
	// lives in the domain (NewDefaultBakersGame) and the i18n namespace in the presenter.
	BindWebControllerFor("bakersgame",
		func() usecase.FreeCellInteractorIF {
			return usecase.NewFreeCellInteractor(domain.NewDefaultBakersGame(), new(presenter.BakersGameWebPresenter))
		},
		controller.NewFreeCellWebController)
	BindWebControllerFor("seahaventowers",
		func() usecase.SeahavenTowersInteractorIF {
			return usecase.NewSeahavenTowersInteractor(domain.NewDefaultSeahavenTowers(), new(presenter.SeahavenTowersWebPresenter))
		},
		controller.NewSeahavenTowersWebController)
	BindWebControllerFor("cruel",
		func() usecase.CruelInteractorIF {
			return usecase.NewCruelInteractor(domain.NewDefaultCruel(), new(presenter.CruelWebPresenter))
		},
		controller.NewCruelWebController)
	BindWebControllerFor("baccarat",
		func() usecase.BaccaratInteractorIF {
			return usecase.NewBaccaratInteractor(domain.NewDefaultBaccarat(), new(presenter.BaccaratWebPresenter))
		},
		controller.NewBaccaratWebController)
	BindWebControllerFor("spades",
		func() usecase.SpadesInteractorIF {
			return usecase.NewSpadesInteractor(domain.NewDefaultSpades(), new(presenter.SpadesWebPresenter))
		},
		controller.NewSpadesWebController)
	BindWebControllerFor("crazyeights",
		func() usecase.CrazyEightsInteractorIF {
			return usecase.NewCrazyEightsInteractor(domain.NewDefaultCrazyEights(), new(presenter.CrazyEightsWebPresenter))
		},
		controller.NewCrazyEightsWebController)
	BindWebControllerFor("ginrummy",
		func() usecase.GinRummyInteractorIF {
			return usecase.NewGinRummyInteractor(domain.NewDefaultGinRummy(), new(presenter.GinRummyWebPresenter))
		},
		controller.NewGinRummyWebController)
	BindWebControllerFor("indianrummy",
		func() usecase.IndianRummyInteractorIF {
			return usecase.NewIndianRummyInteractor(domain.NewDefaultIndianRummy(), new(presenter.IndianRummyWebPresenter))
		},
		controller.NewIndianRummyWebController)
	BindWebControllerFor("canasta",
		func() usecase.CanastaInteractorIF {
			return usecase.NewCanastaInteractor(domain.NewDefaultCanasta(), new(presenter.CanastaWebPresenter))
		},
		controller.NewCanastaWebController)
	BindWebControllerFor("spider",
		func() usecase.SpiderInteractorIF {
			return usecase.NewSpiderInteractor(domain.NewDefaultSpider(), new(presenter.SpiderWebPresenter))
		},
		controller.NewSpiderWebController)
	BindWebControllerFor("napoleon",
		func() usecase.NapoleonInteractorIF {
			return usecase.NewNapoleonInteractor(domain.NewDefaultNapoleon(), new(presenter.NapoleonWebPresenter))
		},
		controller.NewNapoleonWebController)
	BindWebControllerFor("indianpoker",
		func() usecase.IndianPokerInteractorIF {
			return usecase.NewIndianPokerInteractor(domain.NewDefaultIndianPoker(), new(presenter.IndianPokerWebPresenter))
		},
		controller.NewIndianPokerWebController)
	BindWebControllerFor("videopoker",
		func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(domain.NewDefaultVideoPoker(), new(presenter.VideoPokerWebPresenter))
		},
		controller.NewVideoPokerWebController)
	BindWebControllerFor("deuceswild",
		func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(domain.NewDeucesWildVideoPoker(), new(presenter.VideoPokerWebPresenter))
		},
		controller.NewVideoPokerWebController)
	BindWebControllerFor("jokerpoker",
		func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(domain.NewJokerPokerVideoPoker(), new(presenter.VideoPokerWebPresenter))
		},
		controller.NewVideoPokerWebController)
	BindWebControllerFor("euchre",
		func() usecase.EuchreInteractorIF {
			return usecase.NewEuchreInteractor(domain.NewDefaultEuchre(), new(presenter.EuchreWebPresenter))
		},
		controller.NewEuchreWebController)
	BindWebControllerFor("pyramid",
		func() usecase.PyramidInteractorIF {
			return usecase.NewPyramidInteractor(domain.NewDefaultPyramid(), new(presenter.PyramidWebPresenter))
		},
		controller.NewPyramidWebController)
	BindWebControllerFor("tripeaks",
		func() usecase.TriPeaksInteractorIF {
			return usecase.NewTriPeaksInteractor(domain.NewDefaultTriPeaks(), new(presenter.TriPeaksWebPresenter))
		},
		controller.NewTriPeaksWebController)
	BindWebControllerFor("cribbage",
		func() usecase.CribbageInteractorIF {
			return usecase.NewCribbageInteractor(domain.NewDefaultCribbage(), new(presenter.CribbageWebPresenter))
		},
		controller.NewCribbageWebController)
	BindWebControllerFor("threecard",
		func() usecase.ThreeCardInteractorIF {
			return usecase.NewThreeCardInteractor(domain.NewDefaultThreeCard(), new(presenter.ThreeCardWebPresenter))
		},
		controller.NewThreeCardWebController)
	BindWebControllerFor("ohhell",
		func() usecase.OhHellInteractorIF {
			return usecase.NewOhHellInteractor(domain.NewDefaultOhHell(), new(presenter.OhHellWebPresenter))
		},
		controller.NewOhHellWebController)
	BindWebControllerFor("ninetynine",
		func() usecase.NinetyNineInteractorIF {
			return usecase.NewNinetyNineInteractor(domain.NewDefaultNinetyNine(), new(presenter.NinetyNineWebPresenter))
		},
		controller.NewNinetyNineWebController)
	BindWebControllerFor("bridge",
		func() usecase.BridgeInteractorIF {
			return usecase.NewBridgeInteractor(domain.NewDefaultBridge(), new(presenter.BridgeWebPresenter))
		},
		controller.NewBridgeWebController)
	BindWebControllerFor("speed",
		func() usecase.SpeedInteractorIF {
			return usecase.NewSpeedInteractor(domain.NewDefaultSpeed(), new(presenter.SpeedWebPresenter))
		},
		controller.NewSpeedWebController)
	BindWebControllerFor("gofish",
		func() usecase.GoFishInteractorIF {
			return usecase.NewGoFishInteractor(domain.NewDefaultGoFish(), new(presenter.GoFishWebPresenter))
		},
		controller.NewGoFishWebController)
	BindWebControllerFor("pinochle",
		func() usecase.PinochleInteractorIF {
			return usecase.NewPinochleInteractor(domain.NewDefaultPinochle(), new(presenter.PinochleWebPresenter))
		},
		controller.NewPinochleWebController)
	BindWebControllerFor("golf",
		func() usecase.GolfInteractorIF {
			return usecase.NewGolfInteractor(domain.NewDefaultGolf(), new(presenter.GolfWebPresenter))
		},
		controller.NewGolfWebController)
	BindWebControllerFor("pigtail",
		func() usecase.PigsTailInteractorIF {
			return usecase.NewPigsTailInteractor(domain.NewDefaultPigsTail(), new(presenter.PigsTailWebPresenter))
		},
		controller.NewPigsTailWebController)
	BindWebControllerFor("sevencardstud",
		func() usecase.SevenCardStudInteractorIF {
			return usecase.NewSevenCardStudInteractor(domain.NewDefaultSevenCardStud(), new(presenter.SevenCardStudWebPresenter))
		},
		controller.NewSevenCardStudWebController)
	BindWebControllerFor("followthequeen",
		func() usecase.FollowTheQueenInteractorIF {
			return usecase.NewFollowTheQueenInteractor(domain.NewDefaultFollowTheQueen(), new(presenter.FollowTheQueenWebPresenter))
		},
		controller.NewFollowTheQueenWebController)
	BindWebControllerFor("clocksolitaire",
		func() usecase.ClockSolitaireInteractorIF {
			return usecase.NewClockSolitaireInteractor(domain.NewDefaultClockSolitaire(), new(presenter.ClockSolitaireWebPresenter))
		},
		controller.NewClockSolitaireWebController)
	BindWebControllerFor("durak",
		func() usecase.DurakInteractorIF {
			return usecase.NewDurakInteractor(domain.NewDefaultDurak(), new(presenter.DurakWebPresenter))
		},
		controller.NewDurakWebController)
	BindWebControllerFor("fortythieves",
		func() usecase.FortyThievesInteractorIF {
			return usecase.NewFortyThievesInteractor(domain.NewDefaultFortyThieves(), new(presenter.FortyThievesWebPresenter))
		},
		controller.NewFortyThievesWebController)
	BindWebControllerFor("paigow",
		func() usecase.PaiGowInteractorIF {
			return usecase.NewPaiGowInteractor(domain.NewDefaultPaiGow(), new(presenter.PaiGowWebPresenter))
		},
		controller.NewPaiGowWebController)
	BindWebControllerFor("twotenjack",
		func() usecase.TwoTenJackInteractorIF {
			return usecase.NewTwoTenJackInteractor(domain.NewDefaultTwoTenJack(), new(presenter.TwoTenJackWebPresenter))
		},
		controller.NewTwoTenJackWebController)
	BindWebControllerFor("caribbeanstud",
		func() usecase.CaribbeanStudInteractorIF {
			return usecase.NewCaribbeanStudInteractor(domain.NewDefaultCaribbeanStud(), new(presenter.CaribbeanStudWebPresenter))
		},
		controller.NewCaribbeanStudWebController)
	BindWebControllerFor("texasholdembonus",
		func() usecase.TexasHoldemBonusInteractorIF {
			return usecase.NewTexasHoldemBonusInteractor(domain.NewDefaultTexasHoldemBonus(), new(presenter.TexasHoldemBonusWebPresenter))
		},
		controller.NewTexasHoldemBonusWebController)
	BindWebControllerFor("war",
		func() usecase.WarInteractorIF {
			return usecase.NewWarInteractor(domain.NewDefaultWar(), new(presenter.WarWebPresenter))
		},
		controller.NewWarWebController)
	BindWebControllerFor("canfield",
		func() usecase.CanfieldInteractorIF {
			return usecase.NewCanfieldInteractor(domain.NewDefaultCanfield(), new(presenter.CanfieldWebPresenter))
		},
		controller.NewCanfieldWebController)
	BindWebControllerFor("fiftyone",
		func() usecase.FiftyOneInteractorIF {
			return usecase.NewFiftyOneInteractor(domain.NewDefaultFiftyOne(), new(presenter.FiftyOneWebPresenter))
		},
		controller.NewFiftyOneWebController)
	BindWebControllerFor("yukon",
		func() usecase.YukonInteractorIF {
			return usecase.NewYukonInteractor(domain.NewDefaultYukon(), new(presenter.YukonWebPresenter))
		},
		controller.NewYukonWebController)
	BindWebControllerFor("alaska",
		func() usecase.AlaskaInteractorIF {
			return usecase.NewAlaskaInteractor(domain.NewDefaultAlaska(), new(presenter.AlaskaWebPresenter))
		},
		controller.NewAlaskaWebController)
	BindWebControllerFor("russiansolitaire",
		func() usecase.RussianSolitaireInteractorIF {
			return usecase.NewRussianSolitaireInteractor(domain.NewDefaultRussianSolitaire(), new(presenter.RussianSolitaireWebPresenter))
		},
		controller.NewRussianSolitaireWebController)
	BindWebControllerFor("whist",
		func() usecase.WhistInteractorIF {
			return usecase.NewWhistInteractor(domain.NewDefaultWhist(), new(presenter.WhistWebPresenter))
		},
		controller.NewWhistWebController)
	BindWebControllerFor("catchten",
		func() usecase.CatchTenInteractorIF {
			return usecase.NewCatchTenInteractor(domain.NewDefaultCatchTen(), new(presenter.CatchTenWebPresenter))
		},
		controller.NewCatchTenWebController)
	BindWebControllerFor("letitride",
		func() usecase.LetItRideInteractorIF {
			return usecase.NewLetItRideInteractor(domain.NewDefaultLetItRide(), new(presenter.LetItRideWebPresenter))
		},
		controller.NewLetItRideWebController)
	BindWebControllerFor("pokersquares",
		func() usecase.PokerSquaresInteractorIF {
			return usecase.NewPokerSquaresInteractor(domain.NewDefaultPokerSquares(), new(presenter.PokerSquaresWebPresenter))
		},
		controller.NewPokerSquaresWebController)
	BindWebControllerFor("pageone",
		func() usecase.PageOneInteractorIF {
			return usecase.NewPageOneInteractor(domain.NewDefaultPageOne(), new(presenter.PageOneWebPresenter))
		},
		controller.NewPageOneWebController)
	BindWebControllerFor("reddog",
		func() usecase.RedDogInteractorIF {
			return usecase.NewRedDogInteractor(domain.NewDefaultRedDog(), new(presenter.RedDogWebPresenter))
		},
		controller.NewRedDogWebController)
	BindWebControllerFor("razz",
		func() usecase.SevenCardStudInteractorIF {
			return usecase.NewSevenCardStudInteractor(domain.NewDefaultRazz(), new(presenter.SevenCardStudWebPresenter))
		},
		controller.NewSevenCardStudWebController)
	BindWebControllerFor("sevencardstudhilo",
		func() usecase.SevenCardStudInteractorIF {
			return usecase.NewSevenCardStudInteractor(domain.NewDefaultSevenCardStudHiLo(), new(presenter.SevenCardStudWebPresenter))
		},
		controller.NewSevenCardStudWebController)
	BindWebControllerFor("badugi",
		func() usecase.BadugiInteractorIF {
			return usecase.NewBadugiInteractor(domain.NewDefaultBadugi(), new(presenter.BadugiWebPresenter))
		},
		controller.NewBadugiWebController)
	BindWebControllerFor("deucetoseven",
		func() usecase.DeuceToSevenInteractorIF {
			return usecase.NewDeuceToSevenInteractor(domain.NewDefaultDeuceToSeven(), new(presenter.DeuceToSevenWebPresenter))
		},
		controller.NewDeuceToSevenWebController)
	BindWebControllerFor("scorpion",
		func() usecase.ScorpionInteractorIF {
			return usecase.NewScorpionInteractor(domain.NewDefaultScorpion(), new(presenter.ScorpionWebPresenter))
		},
		controller.NewScorpionWebController)
	BindWebControllerFor("wasp",
		func() usecase.WaspInteractorIF {
			return usecase.NewWaspInteractor(domain.NewDefaultWasp(), new(presenter.WaspWebPresenter))
		},
		controller.NewWaspWebController)
	BindWebControllerFor("accordion",
		func() usecase.AccordionInteractorIF {
			return usecase.NewAccordionInteractor(domain.NewDefaultAccordion(), new(presenter.AccordionWebPresenter))
		},
		controller.NewAccordionWebController)
	BindWebControllerFor("trash",
		func() usecase.TrashInteractorIF {
			return usecase.NewTrashInteractor(domain.NewDefaultTrash(), new(presenter.TrashWebPresenter))
		},
		controller.NewTrashWebController)
	BindWebControllerFor("sevenbridge",
		func() usecase.SevenBridgeInteractorIF {
			return usecase.NewSevenBridgeInteractor(domain.NewDefaultSevenBridge(), new(presenter.SevenBridgeWebPresenter))
		},
		controller.NewSevenBridgeWebController)
	BindWebControllerFor("president",
		func() usecase.PresidentInteractorIF {
			return usecase.NewPresidentInteractor(domain.NewDefaultPresident(), new(presenter.PresidentWebPresenter))
		},
		controller.NewPresidentWebController)
	BindWebControllerFor("cassino",
		func() usecase.CassinoInteractorIF {
			return usecase.NewCassinoInteractor(domain.NewDefaultCassino(), new(presenter.CassinoWebPresenter))
		},
		controller.NewCassinoWebController)
	BindWebControllerFor("spanish21",
		func() usecase.BlackJackInteractorIF {
			return usecase.NewBlackJackInteractor(domain.NewSpanish21BlackJack(), new(presenter.BlackJackWebPresenter))
		},
		controller.NewBlackJackWebController)
	BindWebControllerFor("calculation",
		func() usecase.CalculationInteractorIF {
			return usecase.NewCalculationInteractor(domain.NewDefaultCalculation(), new(presenter.CalculationWebPresenter))
		},
		controller.NewCalculationWebController)
	BindWebControllerFor("sirtommy",
		func() usecase.SirTommyInteractorIF {
			return usecase.NewSirTommyInteractor(domain.NewDefaultSirTommy(), new(presenter.SirTommyWebPresenter))
		},
		controller.NewSirTommyWebController)
	BindWebControllerFor("fourseasons",
		func() usecase.FourSeasonsInteractorIF {
			return usecase.NewFourSeasonsInteractor(domain.NewDefaultFourSeasons(), new(presenter.FourSeasonsWebPresenter))
		},
		controller.NewFourSeasonsWebController)
	BindWebControllerFor("colorado",
		func() usecase.ColoradoInteractorIF {
			return usecase.NewColoradoInteractor(domain.NewDefaultColorado(), new(presenter.ColoradoWebPresenter))
		},
		controller.NewColoradoWebController)
	BindWebControllerFor("slyfox",
		func() usecase.SlyFoxInteractorIF {
			return usecase.NewSlyFoxInteractor(domain.NewDefaultSlyFox(), new(presenter.SlyFoxWebPresenter))
		},
		controller.NewSlyFoxWebController)
	BindWebControllerFor("cribbagesquares",
		func() usecase.CribbageSquaresInteractorIF {
			return usecase.NewCribbageSquaresInteractor(domain.NewDefaultCribbageSquares(), new(presenter.CribbageSquaresWebPresenter))
		},
		controller.NewCribbageSquaresWebController)
	BindWebControllerFor("diplomat",
		func() usecase.DiplomatInteractorIF {
			return usecase.NewDiplomatInteractor(domain.NewDefaultDiplomat(), new(presenter.DiplomatWebPresenter))
		},
		controller.NewDiplomatWebController)
	BindWebControllerFor("royalcotillion",
		func() usecase.RoyalCotillionInteractorIF {
			return usecase.NewRoyalCotillionInteractor(domain.NewDefaultRoyalCotillion(), new(presenter.RoyalCotillionWebPresenter))
		},
		controller.NewRoyalCotillionWebController)
	BindWebControllerFor("tarabish",
		func() usecase.TarabishInteractorIF {
			return usecase.NewTarabishInteractor(domain.NewDefaultTarabish(), new(presenter.TarabishWebPresenter))
		},
		controller.NewTarabishWebController)
	BindWebControllerFor("baloot",
		func() usecase.BalootInteractorIF {
			return usecase.NewBalootInteractor(domain.NewDefaultBaloot(), new(presenter.BalootWebPresenter))
		},
		controller.NewBalootWebController)
	BindWebControllerFor("estimation",
		func() usecase.EstimationInteractorIF {
			return usecase.NewEstimationInteractor(domain.NewDefaultEstimation(), new(presenter.EstimationWebPresenter))
		},
		controller.NewEstimationWebController)
	BindWebControllerFor("israeliwhist",
		func() usecase.IsraeliWhistInteractorIF {
			return usecase.NewIsraeliWhistInteractor(domain.NewDefaultIsraeliWhist(), new(presenter.IsraeliWhistWebPresenter))
		},
		controller.NewIsraeliWhistWebController)
	BindWebControllerFor("hokm",
		func() usecase.HokmInteractorIF {
			return usecase.NewHokmInteractor(domain.NewDefaultHokm(), new(presenter.HokmWebPresenter))
		},
		controller.NewHokmWebController)
	BindWebControllerFor("shelem",
		func() usecase.ShelemInteractorIF {
			return usecase.NewShelemInteractor(domain.NewDefaultShelem(), new(presenter.ShelemWebPresenter))
		},
		controller.NewShelemWebController)
	BindWebControllerFor("mendikot",
		func() usecase.MendikotInteractorIF {
			return usecase.NewMendikotInteractor(domain.NewDefaultMendikot(), new(presenter.MendikotWebPresenter))
		},
		controller.NewMendikotWebController)
	BindWebControllerFor("bhabhi",
		func() usecase.BhabhiInteractorIF {
			return usecase.NewBhabhiInteractor(domain.NewDefaultBhabhi(), new(presenter.BhabhiWebPresenter))
		},
		controller.NewBhabhiWebController)
	BindWebControllerFor("teendopaanch",
		func() usecase.TeenDoPaanchInteractorIF {
			return usecase.NewTeenDoPaanchInteractor(domain.NewDefaultTeenDoPaanch(), new(presenter.TeenDoPaanchWebPresenter))
		},
		controller.NewTeenDoPaanchWebController)
	BindWebControllerFor("hasenpfeffer",
		func() usecase.HasenpfefferInteractorIF {
			return usecase.NewHasenpfefferInteractor(domain.NewDefaultHasenpfeffer(), new(presenter.HasenpfefferWebPresenter))
		},
		controller.NewHasenpfefferWebController)
	BindWebControllerFor("sergeantmajor",
		func() usecase.SergeantMajorInteractorIF {
			return usecase.NewSergeantMajorInteractor(domain.NewDefaultSergeantMajor(), new(presenter.SergeantMajorWebPresenter))
		},
		controller.NewSergeantMajorWebController)
	BindWebControllerFor("honeymoonbridge",
		func() usecase.HoneymoonBridgeInteractorIF {
			return usecase.NewHoneymoonBridgeInteractor(domain.NewDefaultHoneymoonBridge(), new(presenter.HoneymoonBridgeWebPresenter))
		},
		controller.NewHoneymoonBridgeWebController)
	BindWebControllerFor("minibridge",
		func() usecase.MinibridgeInteractorIF {
			return usecase.NewMinibridgeInteractor(domain.NewDefaultMinibridge(), new(presenter.MinibridgeWebPresenter))
		},
		controller.NewMinibridgeWebController)
	BindWebControllerFor("pasur",
		func() usecase.PasurInteractorIF {
			return usecase.NewPasurInteractor(domain.NewDefaultPasur(), new(presenter.PasurWebPresenter))
		},
		controller.NewPasurWebController)
	BindWebControllerFor("snap",
		func() usecase.SnapInteractorIF {
			return usecase.NewSnapInteractor(domain.NewDefaultSnap(), new(presenter.SnapWebPresenter))
		},
		controller.NewSnapWebController)
	BindWebControllerFor("rollingstone",
		func() usecase.RollingStoneInteractorIF {
			return usecase.NewRollingStoneInteractor(domain.NewDefaultRollingStone(), new(presenter.RollingStoneWebPresenter))
		},
		controller.NewRollingStoneWebController)
	BindWebControllerFor("lingerlonger",
		func() usecase.LingerLongerInteractorIF {
			return usecase.NewLingerLongerInteractor(domain.NewDefaultLingerLonger(), new(presenter.LingerLongerWebPresenter))
		},
		controller.NewLingerLongerWebController)
	BindWebControllerFor("pig",
		func() usecase.PigInteractorIF {
			return usecase.NewPigInteractor(domain.NewDefaultPig(), new(presenter.PigWebPresenter))
		},
		controller.NewPigWebController)
	BindWebControllerFor("stealingbundles",
		func() usecase.StealingBundlesInteractorIF {
			return usecase.NewStealingBundlesInteractor(domain.NewDefaultStealingBundles(), new(presenter.StealingBundlesWebPresenter))
		},
		controller.NewStealingBundlesWebController)
	BindWebControllerFor("cucumber",
		func() usecase.CucumberInteractorIF {
			return usecase.NewCucumberInteractor(domain.NewDefaultCucumber(), new(presenter.CucumberWebPresenter))
		},
		controller.NewCucumberWebController)
	BindWebControllerFor("goofspiel",
		func() usecase.GoofspielInteractorIF {
			return usecase.NewGoofspielInteractor(domain.NewDefaultGoofspiel(), new(presenter.GoofspielWebPresenter))
		},
		controller.NewGoofspielWebController)
	BindWebControllerFor("rams",
		func() usecase.RamsInteractorIF {
			return usecase.NewRamsInteractor(domain.NewDefaultRams(), new(presenter.RamsWebPresenter))
		},
		controller.NewRamsWebController)
	BindWebControllerFor("reversis",
		func() usecase.ReversisInteractorIF {
			return usecase.NewReversisInteractor(domain.NewDefaultReversis(), new(presenter.ReversisWebPresenter))
		},
		controller.NewReversisWebController)
	BindWebControllerFor("polignac",
		func() usecase.PolignacInteractorIF {
			return usecase.NewPolignacInteractor(domain.NewDefaultPolignac(), new(presenter.PolignacWebPresenter))
		},
		controller.NewPolignacWebController)
	BindWebControllerFor("slobberhannes",
		func() usecase.SlobberhannesInteractorIF {
			return usecase.NewSlobberhannesInteractor(domain.NewDefaultSlobberhannes(), new(presenter.SlobberhannesWebPresenter))
		},
		controller.NewSlobberhannesWebController)
	BindWebControllerFor("germanwhist",
		func() usecase.GermanWhistInteractorIF {
			return usecase.NewGermanWhistInteractor(domain.NewDefaultGermanWhist(), new(presenter.GermanWhistWebPresenter))
		},
		controller.NewGermanWhistWebController)
	BindWebControllerFor("crazyquilt",
		func() usecase.CrazyQuiltInteractorIF {
			return usecase.NewCrazyQuiltInteractor(domain.NewDefaultCrazyQuilt(), new(presenter.CrazyQuiltWebPresenter))
		},
		controller.NewCrazyQuiltWebController)
	BindWebControllerFor("soko",
		func() usecase.FiveCardStudInteractorIF {
			return usecase.NewFiveCardStudInteractor(domain.NewDefaultSoko(), new(presenter.FiveCardStudWebPresenter))
		},
		controller.NewFiveCardStudWebController)
	BindWebControllerFor("auldlangsyne",
		func() usecase.AuldLangSyneInteractorIF {
			return usecase.NewAuldLangSyneInteractor(domain.NewDefaultAuldLangSyne(), new(presenter.AuldLangSyneWebPresenter))
		},
		controller.NewAuldLangSyneWebController)
	BindWebControllerFor("bisley",
		func() usecase.BisleyInteractorIF {
			return usecase.NewBisleyInteractor(domain.NewDefaultBisley(), new(presenter.BisleyWebPresenter))
		},
		controller.NewBisleyWebController)
	BindWebControllerFor("napoleonssquare",
		func() usecase.NapoleonsSquareInteractorIF {
			return usecase.NewNapoleonsSquareInteractor(domain.NewDefaultNapoleonsSquare(), new(presenter.NapoleonsSquareWebPresenter))
		},
		controller.NewNapoleonsSquareWebController)
	BindWebControllerFor("grandfathersclock",
		func() usecase.GrandfathersClockInteractorIF {
			return usecase.NewGrandfathersClockInteractor(domain.NewDefaultGrandfathersClock(), new(presenter.GrandfathersClockWebPresenter))
		},
		controller.NewGrandfathersClockWebController)
	BindWebControllerFor("bigben",
		func() usecase.BigBenInteractorIF {
			return usecase.NewBigBenInteractor(domain.NewDefaultBigBen(), new(presenter.BigBenWebPresenter))
		},
		controller.NewBigBenWebController)
	BindWebControllerFor("missmilligan",
		func() usecase.MissMilliganInteractorIF {
			return usecase.NewMissMilliganInteractor(domain.NewDefaultMissMilligan(), new(presenter.MissMilliganWebPresenter))
		},
		controller.NewMissMilliganWebController)
	BindWebControllerFor("duchess",
		func() usecase.DuchessInteractorIF {
			return usecase.NewDuchessInteractor(domain.NewDefaultDuchess(), new(presenter.DuchessWebPresenter))
		},
		controller.NewDuchessWebController)
	BindWebControllerFor("windmill",
		func() usecase.WindmillInteractorIF {
			return usecase.NewWindmillInteractor(domain.NewDefaultWindmill(), new(presenter.WindmillWebPresenter))
		},
		controller.NewWindmillWebController)
	BindWebControllerFor("americantoad",
		func() usecase.AmericanToadInteractorIF {
			return usecase.NewAmericanToadInteractor(domain.NewDefaultAmericanToad(), new(presenter.AmericanToadWebPresenter))
		},
		controller.NewAmericanToadWebController)
	BindWebControllerFor("spiteandmalice",
		func() usecase.SpiteAndMaliceInteractorIF {
			return usecase.NewSpiteAndMaliceInteractor(domain.NewDefaultSpiteAndMalice(), new(presenter.SpiteAndMaliceWebPresenter))
		},
		controller.NewSpiteAndMaliceWebController)
	BindWebControllerFor("skat",
		func() usecase.SkatInteractorIF {
			return usecase.NewSkatInteractor(domain.NewDefaultSkat(), new(presenter.SkatWebPresenter))
		},
		controller.NewSkatWebController)
	BindWebControllerFor("congress",
		func() usecase.CongressInteractorIF {
			return usecase.NewCongressInteractor(domain.NewDefaultCongress(), new(presenter.CongressWebPresenter))
		},
		controller.NewCongressWebController)
	BindWebControllerFor("saliclaw",
		func() usecase.SalicLawInteractorIF {
			return usecase.NewSalicLawInteractor(domain.NewDefaultSalicLaw(), new(presenter.SalicLawWebPresenter))
		},
		controller.NewSalicLawWebController)
	BindWebControllerFor("terrace",
		func() usecase.TerraceInteractorIF {
			return usecase.NewTerraceInteractor(domain.NewDefaultTerrace(), new(presenter.TerraceWebPresenter))
		},
		controller.NewTerraceWebController)
	BindWebControllerFor("braid",
		func() usecase.BraidInteractorIF {
			return usecase.NewBraidInteractor(domain.NewDefaultBraid(), new(presenter.BraidWebPresenter))
		},
		controller.NewBraidWebController)
	BindWebControllerFor("pontoon",
		func() usecase.PontoonInteractorIF {
			return usecase.NewPontoonInteractor(domain.NewDefaultPontoon(), new(presenter.PontoonWebPresenter))
		},
		controller.NewPontoonWebController)
	BindWebControllerFor("settemezzo",
		func() usecase.SetteEMezzoInteractorIF {
			return usecase.NewSetteEMezzoInteractor(domain.NewDefaultSetteEMezzo(), new(presenter.SetteEMezzoWebPresenter))
		},
		controller.NewSetteEMezzoWebController)
	BindWebControllerFor("niuniu",
		func() usecase.NiuNiuInteractorIF {
			return usecase.NewNiuNiuInteractor(domain.NewDefaultNiuNiu(), new(presenter.NiuNiuWebPresenter))
		},
		controller.NewNiuNiuWebController)
	BindWebControllerFor("bura",
		func() usecase.BuraInteractorIF {
			return usecase.NewBuraInteractor(domain.NewDefaultBura(), new(presenter.BuraWebPresenter))
		},
		controller.NewBuraWebController)
	BindWebControllerFor("mushi",
		func() usecase.MushiInteractorIF {
			return usecase.NewMushiInteractor(domain.NewDefaultMushi(), new(presenter.MushiWebPresenter))
		},
		controller.NewMushiWebController)
	BindWebControllerFor("toepen",
		func() usecase.ToepenInteractorIF {
			return usecase.NewToepenInteractor(domain.NewDefaultToepen(), new(presenter.ToepenWebPresenter))
		},
		controller.NewToepenWebController)
	BindWebControllerFor("chineseten",
		func() usecase.ChineseTenInteractorIF {
			return usecase.NewChineseTenInteractor(domain.NewDefaultChineseTen(), new(presenter.ChineseTenWebPresenter))
		},
		controller.NewChineseTenWebController)
	BindWebControllerFor("trex",
		func() usecase.TrexInteractorIF {
			return usecase.NewTrexInteractor(domain.NewDefaultTrex(), new(presenter.TrexWebPresenter))
		},
		controller.NewTrexWebController)
	BindWebControllerFor("skitgubbe",
		func() usecase.SkitgubbeInteractorIF {
			return usecase.NewSkitgubbeInteractor(domain.NewDefaultSkitgubbe(), new(presenter.SkitgubbeWebPresenter))
		},
		controller.NewSkitgubbeWebController)
	BindWebControllerFor("loba",
		func() usecase.LobaInteractorIF {
			return usecase.NewLobaInteractor(domain.NewDefaultLoba(), new(presenter.LobaWebPresenter))
		},
		controller.NewLobaWebController)
	BindWebControllerFor("sjavs",
		func() usecase.SjavsInteractorIF {
			return usecase.NewSjavsInteractor(domain.NewDefaultSjavs(), new(presenter.SjavsWebPresenter))
		},
		controller.NewSjavsWebController)
	BindWebControllerFor("laughandliedown",
		func() usecase.LaughAndLieDownInteractorIF {
			return usecase.NewLaughAndLieDownInteractor(domain.NewDefaultLaughAndLieDown(), new(presenter.LaughAndLieDownWebPresenter))
		},
		controller.NewLaughAndLieDownWebController)
	BindWebControllerFor("shithead",
		func() usecase.ShitheadInteractorIF {
			return usecase.NewShitheadInteractor(domain.NewDefaultShithead(), new(presenter.ShitheadWebPresenter))
		},
		controller.NewShitheadWebController)
	BindWebControllerFor("nertz",
		func() usecase.NertzInteractorIF {
			return usecase.NewNertzInteractor(domain.NewDefaultNertz(), new(presenter.NertzWebPresenter))
		},
		controller.NewNertzWebController)
	BindWebControllerFor("slapjack",
		func() usecase.SlapjackInteractorIF {
			return usecase.NewSlapjackInteractor(domain.NewDefaultSlapjack(), new(presenter.SlapjackWebPresenter))
		},
		controller.NewSlapjackWebController)
	BindWebControllerFor("egyptianratscrew",
		func() usecase.EgyptianRatscrewInteractorIF {
			return usecase.NewEgyptianRatscrewInteractor(domain.NewDefaultEgyptianRatscrew(), new(presenter.EgyptianRatscrewWebPresenter))
		},
		controller.NewEgyptianRatscrewWebController)
	BindWebControllerFor("bakersdozen",
		func() usecase.BakersDozenInteractorIF {
			return usecase.NewBakersDozenInteractor(domain.NewDefaultBakersDozen(), new(presenter.BakersDozenWebPresenter))
		},
		controller.NewBakersDozenWebController)
	BindWebControllerFor("perseverance",
		func() usecase.PerseveranceInteractorIF {
			return usecase.NewPerseveranceInteractor(domain.NewDefaultPerseverance(), new(presenter.PerseveranceWebPresenter))
		},
		controller.NewPerseveranceWebController)
	BindWebControllerFor("fourteenout",
		func() usecase.FourteenOutInteractorIF {
			return usecase.NewFourteenOutInteractor(domain.NewDefaultFourteenOut(), new(presenter.FourteenOutWebPresenter))
		},
		controller.NewFourteenOutWebController)
	BindWebControllerFor("narcotic",
		func() usecase.NarcoticInteractorIF {
			return usecase.NewNarcoticInteractor(domain.NewDefaultNarcotic(), new(presenter.NarcoticWebPresenter))
		},
		controller.NewNarcoticWebController)
	BindWebControllerFor("mrsmop",
		func() usecase.MrsMopInteractorIF {
			return usecase.NewMrsMopInteractor(domain.NewDefaultMrsMop(), new(presenter.MrsMopWebPresenter))
		},
		controller.NewMrsMopWebController)
	BindWebControllerFor("rankandfile",
		func() usecase.RankAndFileInteractorIF {
			return usecase.NewRankAndFileInteractor(domain.NewDefaultRankAndFile(), new(presenter.RankAndFileWebPresenter))
		},
		controller.NewRankAndFileWebController)
	BindWebControllerFor("tonk",
		func() usecase.TonkInteractorIF {
			return usecase.NewTonkInteractor(domain.NewDefaultTonk(), new(presenter.TonkWebPresenter))
		},
		controller.NewTonkWebController)
	BindWebControllerFor("thirtyone",
		func() usecase.ThirtyOneInteractorIF {
			return usecase.NewThirtyOneInteractor(domain.NewDefaultThirtyOne(), new(presenter.ThirtyOneWebPresenter))
		},
		controller.NewThirtyOneWebController)
	BindWebControllerFor("tienlen",
		func() usecase.TienLenInteractorIF {
			return usecase.NewTienLenInteractor(domain.NewDefaultTienLen(), new(presenter.TienLenWebPresenter))
		},
		controller.NewTienLenWebController)
	BindWebControllerFor("osmosis",
		func() usecase.OsmosisInteractorIF {
			return usecase.NewOsmosisInteractor(domain.NewDefaultOsmosis(), new(presenter.OsmosisWebPresenter))
		},
		controller.NewOsmosisWebController)
	BindWebControllerFor("fivehundred",
		func() usecase.FiveHundredInteractorIF {
			return usecase.NewFiveHundredInteractor(domain.NewDefaultFiveHundred(), new(presenter.FiveHundredWebPresenter))
		},
		controller.NewFiveHundredWebController)
	BindWebControllerFor("schnapsen",
		func() usecase.SchnapsenInteractorIF {
			return usecase.NewSchnapsenInteractor(domain.NewDefaultSchnapsen(), new(presenter.SchnapsenWebPresenter))
		},
		controller.NewSchnapsenWebController)
	BindWebControllerFor("burraco",
		func() usecase.BurracoInteractorIF {
			return usecase.NewBurracoInteractor(domain.NewDefaultBurraco(), new(presenter.BurracoWebPresenter))
		},
		controller.NewBurracoWebController)
	BindWebControllerFor("yaniv",
		func() usecase.YanivInteractorIF {
			return usecase.NewYanivInteractor(domain.NewDefaultYaniv(), new(presenter.YanivWebPresenter))
		},
		controller.NewYanivWebController)
	BindWebControllerFor("casinowar",
		func() usecase.CasinoWarInteractorIF {
			return usecase.NewCasinoWarInteractor(domain.NewDefaultCasinoWar(), new(presenter.CasinoWarWebPresenter))
		},
		controller.NewCasinoWarWebController)
	BindWebControllerFor("pitch",
		func() usecase.PitchInteractorIF {
			return usecase.NewPitchInteractor(domain.NewDefaultPitch(), new(presenter.PitchWebPresenter))
		},
		controller.NewPitchWebController)
	BindWebControllerFor("dragontiger",
		func() usecase.DragonTigerInteractorIF {
			return usecase.NewDragonTigerInteractor(domain.NewDefaultDragonTiger(), new(presenter.DragonTigerWebPresenter))
		},
		controller.NewDragonTigerWebController)
	BindWebControllerFor("andarbahar",
		func() usecase.AndarBaharInteractorIF {
			return usecase.NewAndarBaharInteractor(domain.NewDefaultAndarBahar(), new(presenter.AndarBaharWebPresenter))
		},
		controller.NewAndarBaharWebController)
	BindWebControllerFor("botifarra",
		func() usecase.BotifarraInteractorIF {
			return usecase.NewBotifarraInteractor(domain.NewDefaultBotifarra(), new(presenter.BotifarraWebPresenter))
		},
		controller.NewBotifarraWebController)
	BindWebControllerFor("rikken",
		func() usecase.RikkenInteractorIF {
			return usecase.NewRikkenInteractor(domain.NewDefaultRikken(), new(presenter.RikkenWebPresenter))
		},
		controller.NewRikkenWebController)
	BindWebControllerFor("doubleattack",
		func() usecase.DoubleAttackBlackjackInteractorIF {
			return usecase.NewDoubleAttackBlackjackInteractor(domain.NewDefaultDoubleAttackBlackjack(), new(presenter.DoubleAttackBlackjackWebPresenter))
		},
		controller.NewDoubleAttackBlackjackWebController)
	BindWebControllerFor("crazyfourpoker",
		func() usecase.CrazyFourPokerInteractorIF {
			return usecase.NewCrazyFourPokerInteractor(domain.NewDefaultCrazyFourPoker(), new(presenter.CrazyFourPokerWebPresenter))
		},
		controller.NewCrazyFourPokerWebController)
	BindWebControllerFor("chemindefer",
		func() usecase.ChemindeFerInteractorIF {
			return usecase.NewChemindeFerInteractor(domain.NewDefaultChemindeFer(), new(presenter.ChemindeFerWebPresenter))
		},
		controller.NewChemindeFerWebController)
	BindWebControllerFor("colourwhist",
		func() usecase.ColourWhistInteractorIF {
			return usecase.NewColourWhistInteractor(domain.NewDefaultColourWhist(), new(presenter.ColourWhistWebPresenter))
		},
		controller.NewColourWhistWebController)
	BindWebControllerFor("blackjackswitch",
		func() usecase.BlackJackSwitchInteractorIF {
			return usecase.NewBlackJackSwitchInteractor(domain.NewDefaultBlackJackSwitch(), new(presenter.BlackJackSwitchWebPresenter))
		},
		controller.NewBlackJackSwitchWebController)
	BindWebControllerFor("montecarlo",
		func() usecase.MonteCarloInteractorIF {
			return usecase.NewMonteCarloInteractor(domain.NewDefaultMonteCarlo(), new(presenter.MonteCarloWebPresenter))
		},
		controller.NewMonteCarloWebController)
	BindWebControllerFor("contractrummy",
		func() usecase.ContractRummyInteractorIF {
			return usecase.NewContractRummyInteractor(domain.NewDefaultContractRummy(), new(presenter.ContractRummyWebPresenter))
		},
		controller.NewContractRummyWebController)
	BindWebControllerFor("ultimatetexasholdem",
		func() usecase.UltimateTexasHoldemInteractorIF {
			return usecase.NewUltimateTexasHoldemInteractor(domain.NewDefaultUltimateTexasHoldem(), new(presenter.UltimateTexasHoldemWebPresenter))
		},
		controller.NewUltimateTexasHoldemWebController)
	BindWebControllerFor("crescent",
		func() usecase.CrescentInteractorIF {
			return usecase.NewCrescentInteractor(domain.NewDefaultCrescent(), new(presenter.CrescentWebPresenter))
		},
		controller.NewCrescentWebController)
	BindWebControllerFor("sthelena",
		func() usecase.StHelenaInteractorIF {
			return usecase.NewStHelenaInteractor(domain.NewDefaultStHelena(), new(presenter.StHelenaWebPresenter))
		},
		controller.NewStHelenaWebController)
	BindWebControllerFor("mississippistud",
		func() usecase.MississippiStudInteractorIF {
			return usecase.NewMississippiStudInteractor(domain.NewDefaultMississippiStud(), new(presenter.MississippiStudWebPresenter))
		},
		controller.NewMississippiStudWebController)
	BindWebControllerFor("belote",
		func() usecase.BeloteInteractorIF {
			return usecase.NewBeloteInteractor(domain.NewDefaultBelote(), new(presenter.BeloteWebPresenter))
		},
		controller.NewBeloteWebController)
	BindWebControllerFor("spiderette",
		func() usecase.SpideretteInteractorIF {
			return usecase.NewSpideretteInteractor(domain.NewDefaultSpiderette(), new(presenter.SpideretteWebPresenter))
		},
		controller.NewSpideretteWebController)
	BindWebControllerFor("mighty",
		func() usecase.MightyInteractorIF {
			return usecase.NewMightyInteractor(domain.NewDefaultMighty(), new(presenter.MightyWebPresenter))
		},
		controller.NewMightyWebController)
	BindWebControllerFor("oasispoker",
		func() usecase.OasisPokerInteractorIF {
			return usecase.NewOasisPokerInteractor(domain.NewDefaultOasisPoker(), new(presenter.OasisPokerWebPresenter))
		},
		controller.NewOasisPokerWebController)
	BindWebControllerFor("stalactites",
		func() usecase.StalactitesInteractorIF {
			return usecase.NewStalactitesInteractor(domain.NewDefaultStalactites(), new(presenter.StalactitesWebPresenter))
		},
		controller.NewStalactitesWebController)
	BindWebControllerFor("somerset",
		func() usecase.SomersetInteractorIF {
			return usecase.NewSomersetInteractor(domain.NewDefaultSomerset(), new(presenter.SomersetWebPresenter))
		},
		controller.NewSomersetWebController)
	BindWebControllerFor("fortress",
		func() usecase.FortressInteractorIF {
			return usecase.NewFortressInteractor(domain.NewDefaultFortress(), new(presenter.FortressWebPresenter))
		},
		controller.NewFortressWebController)
	BindWebControllerFor("beleagueredcastle",
		func() usecase.BeleagueredCastleInteractorIF {
			return usecase.NewBeleagueredCastleInteractor(domain.NewDefaultBeleagueredCastle(), new(presenter.BeleagueredCastleWebPresenter))
		},
		controller.NewBeleagueredCastleWebController)
	BindWebControllerFor("streetsandalleys",
		func() usecase.StreetsAndAlleysInteractorIF {
			return usecase.NewStreetsAndAlleysInteractor(domain.NewDefaultStreetsAndAlleys(), new(presenter.StreetsAndAlleysWebPresenter))
		},
		controller.NewStreetsAndAlleysWebController)
	BindWebControllerFor("kingalbert",
		func() usecase.KingAlbertInteractorIF {
			return usecase.NewKingAlbertInteractor(domain.NewDefaultKingAlbert(), new(presenter.KingAlbertWebPresenter))
		},
		controller.NewKingAlbertWebController)
	BindWebControllerFor("flowergarden",
		func() usecase.FlowerGardenInteractorIF {
			return usecase.NewFlowerGardenInteractor(domain.NewDefaultFlowerGarden(), new(presenter.FlowerGardenWebPresenter))
		},
		controller.NewFlowerGardenWebController)
	BindWebControllerFor("fortyandeight",
		func() usecase.FortyAndEightInteractorIF {
			return usecase.NewFortyAndEightInteractor(domain.NewDefaultFortyAndEight(), new(presenter.FortyAndEightWebPresenter))
		},
		controller.NewFortyAndEightWebController)
	BindWebControllerFor("agnes",
		func() usecase.AgnesInteractorIF {
			return usecase.NewAgnesInteractor(domain.NewDefaultAgnes(), new(presenter.AgnesWebPresenter))
		},
		controller.NewAgnesWebController)
	BindWebControllerFor("sultan",
		func() usecase.SultanInteractorIF {
			return usecase.NewSultanInteractor(domain.NewDefaultSultan(), new(presenter.SultanWebPresenter))
		},
		controller.NewSultanWebController)
	BindWebControllerFor("piquet",
		func() usecase.PiquetInteractorIF {
			return usecase.NewPiquetInteractor(domain.NewDefaultPiquet(), new(presenter.PiquetWebPresenter))
		},
		controller.NewPiquetWebController)
	BindWebControllerFor("casinoholdem",
		func() usecase.CasinoHoldemInteractorIF {
			return usecase.NewCasinoHoldemInteractor(domain.NewDefaultCasinoHoldem(), new(presenter.CasinoHoldemWebPresenter))
		},
		controller.NewCasinoHoldemWebController)
	BindWebControllerFor("callbreak",
		func() usecase.CallBreakInteractorIF {
			return usecase.NewCallBreakInteractor(domain.NewDefaultCallBreak(), new(presenter.CallBreakWebPresenter))
		},
		controller.NewCallBreakWebController)
	BindWebControllerFor("tarneeb",
		func() usecase.TarneebInteractorIF {
			return usecase.NewTarneebInteractor(domain.NewDefaultTarneeb(), new(presenter.TarneebWebPresenter))
		},
		controller.NewTarneebWebController)
	BindWebControllerFor("highcardflush",
		func() usecase.HighCardFlushInteractorIF {
			return usecase.NewHighCardFlushInteractor(domain.NewDefaultHighCardFlush(), new(presenter.HighCardFlushWebPresenter))
		},
		controller.NewHighCardFlushWebController)
	BindWebControllerFor("briscola",
		func() usecase.BriscolaInteractorIF {
			return usecase.NewBriscolaInteractor(domain.NewDefaultBriscola(), new(presenter.BriscolaWebPresenter))
		},
		controller.NewBriscolaWebController)
	BindWebControllerFor("gaps",
		func() usecase.GapsInteractorIF {
			return usecase.NewGapsInteractor(domain.NewDefaultGaps(), new(presenter.GapsWebPresenter))
		},
		controller.NewGapsWebController)
	BindWebControllerFor("fourcardpoker",
		func() usecase.FourCardPokerInteractorIF {
			return usecase.NewFourCardPokerInteractor(domain.NewDefaultFourCardPoker(), new(presenter.FourCardPokerWebPresenter))
		},
		controller.NewFourCardPokerWebController)
	BindWebControllerFor("rummy500",
		func() usecase.Rummy500InteractorIF {
			return usecase.NewRummy500Interactor(domain.NewDefaultRummy500(), new(presenter.Rummy500WebPresenter))
		},
		controller.NewRummy500WebController)
	BindWebControllerFor("eightoff",
		func() usecase.EightOffInteractorIF {
			return usecase.NewEightOffInteractor(domain.NewDefaultEightOff(), new(presenter.EightOffWebPresenter))
		},
		controller.NewEightOffWebController)
	BindWebControllerFor("russianpoker",
		func() usecase.RussianPokerInteractorIF {
			return usecase.NewRussianPokerInteractor(domain.NewDefaultRussianPoker(), new(presenter.RussianPokerWebPresenter))
		},
		controller.NewRussianPokerWebController)
	BindWebControllerFor("penguin",
		func() usecase.PenguinInteractorIF {
			return usecase.NewPenguinInteractor(domain.NewDefaultPenguin(), new(presenter.PenguinWebPresenter))
		},
		controller.NewPenguinWebController)
	BindWebControllerFor("chinesepoker",
		func() usecase.ChinesePokerInteractorIF {
			return usecase.NewChinesePokerInteractor(domain.NewDefaultChinesePoker(), new(presenter.ChinesePokerWebPresenter))
		},
		controller.NewChinesePokerWebController)
	BindWebControllerFor("sixcardgolf",
		func() usecase.SixCardGolfInteractorIF {
			return usecase.NewSixCardGolfInteractor(domain.NewDefaultSixCardGolf(), new(presenter.SixCardGolfWebPresenter))
		},
		controller.NewSixCardGolfWebController)
	BindWebControllerFor("doudizhu",
		func() usecase.DoudizhuInteractorIF {
			return usecase.NewDoudizhuInteractor(domain.NewDefaultDoudizhu(), new(presenter.DoudizhuWebPresenter))
		},
		controller.NewDoudizhuWebController)
	BindWebControllerFor("truco",
		func() usecase.TrucoInteractorIF {
			return usecase.NewTrucoInteractor(domain.NewDefaultTruco(), new(presenter.TrucoWebPresenter))
		},
		controller.NewTrucoWebController)
	BindWebControllerFor("scopa",
		func() usecase.ScopaInteractorIF {
			return usecase.NewScopaInteractor(domain.NewDefaultScopa(), new(presenter.ScopaWebPresenter))
		},
		controller.NewScopaWebController)
	BindWebControllerFor("acesup",
		func() usecase.AcesUpInteractorIF {
			return usecase.NewAcesUpInteractor(domain.NewDefaultAcesUp(), new(presenter.AcesUpWebPresenter))
		},
		controller.NewAcesUpWebController)
	BindWebControllerFor("barbu",
		func() usecase.BarbuInteractorIF {
			return usecase.NewBarbuInteractor(domain.NewDefaultBarbu(), new(presenter.BarbuWebPresenter))
		},
		controller.NewBarbuWebController)
	BindWebControllerFor("macau",
		func() usecase.MacauInteractorIF {
			return usecase.NewMacauInteractor(domain.NewDefaultMacau(), new(presenter.MacauWebPresenter))
		},
		controller.NewMacauWebController)
	BindWebControllerFor("russianbank",
		func() usecase.RussianBankInteractorIF {
			return usecase.NewRussianBankInteractor(domain.NewDefaultRussianBank(), new(presenter.RussianBankWebPresenter))
		},
		controller.NewRussianBankWebController)
	BindWebControllerFor("shamrocks",
		func() usecase.ShamrocksInteractorIF {
			return usecase.NewShamrocksInteractor(domain.NewDefaultShamrocks(), new(presenter.ShamrocksWebPresenter))
		},
		controller.NewShamrocksWebController)
	BindWebControllerFor("labellelucie",
		func() usecase.LaBelleLucieInteractorIF {
			return usecase.NewLaBelleLucieInteractor(domain.NewDefaultLaBelleLucie(), new(presenter.LaBelleLucieWebPresenter))
		},
		controller.NewLaBelleLucieWebController)
	BindWebControllerFor("curdsandwhey",
		func() usecase.CurdsAndWheyInteractorIF {
			return usecase.NewCurdsAndWheyInteractor(domain.NewDefaultCurdsAndWhey(), new(presenter.CurdsAndWheyWebPresenter))
		},
		controller.NewCurdsAndWheyWebController)
	BindWebControllerFor("simplesimon",
		func() usecase.SimpleSimonInteractorIF {
			return usecase.NewSimpleSimonInteractor(domain.NewDefaultSimpleSimon(), new(presenter.SimpleSimonWebPresenter))
		},
		controller.NewSimpleSimonWebController)
	BindWebControllerFor("doubleklondike",
		func() usecase.DoubleKlondikeInteractorIF {
			return usecase.NewDoubleKlondikeInteractor(domain.NewDefaultDoubleKlondike(), new(presenter.DoubleKlondikeWebPresenter))
		},
		controller.NewDoubleKlondikeWebController)
	BindWebControllerFor("blackhole",
		func() usecase.BlackHoleInteractorIF {
			return usecase.NewBlackHoleInteractor(domain.NewDefaultBlackHole(), new(presenter.BlackHoleWebPresenter))
		},
		controller.NewBlackHoleWebController)
	BindWebControllerFor("gongzhu",
		func() usecase.GongZhuInteractorIF {
			return usecase.NewGongZhuInteractor(domain.NewDefaultGongZhu(), new(presenter.GongZhuWebPresenter))
		},
		controller.NewGongZhuWebController)
	BindWebControllerFor("bristol",
		func() usecase.BristolInteractorIF {
			return usecase.NewBristolInteractor(domain.NewDefaultBristol(), new(presenter.BristolWebPresenter))
		},
		controller.NewBristolWebController)
	BindWebControllerFor("bidwhist",
		func() usecase.BidWhistInteractorIF {
			return usecase.NewBidWhistInteractor(domain.NewDefaultBidWhist(), new(presenter.BidWhistWebPresenter))
		},
		controller.NewBidWhistWebController)
	BindWebControllerFor("tressette",
		func() usecase.TressetteInteractorIF {
			return usecase.NewTressetteInteractor(domain.NewDefaultTressette(), new(presenter.TressetteWebPresenter))
		},
		controller.NewTressetteWebController)
	BindWebControllerFor("sheepshead",
		func() usecase.SheepsheadInteractorIF {
			return usecase.NewSheepsheadInteractor(domain.NewDefaultSheepshead(), new(presenter.SheepsheadWebPresenter))
		},
		controller.NewSheepsheadWebController)
	BindWebControllerFor("doppelkopf",
		func() usecase.DoppelkopfInteractorIF {
			return usecase.NewDoppelkopfInteractor(domain.NewDefaultDoppelkopf(), new(presenter.DoppelkopfWebPresenter))
		},
		controller.NewDoppelkopfWebController)
	BindWebControllerFor("mus",
		func() usecase.MusInteractorIF {
			return usecase.NewMusInteractor(domain.NewDefaultMus(), new(presenter.MusWebPresenter))
		},
		controller.NewMusWebController)
	BindWebControllerFor("tute",
		func() usecase.TuteInteractorIF {
			return usecase.NewTuteInteractor(domain.NewDefaultTute(), new(presenter.TuteWebPresenter))
		},
		controller.NewTuteWebController)
	BindWebControllerFor("sueca",
		func() usecase.SuecaInteractorIF {
			return usecase.NewSuecaInteractor(domain.NewDefaultSueca(), new(presenter.SuecaWebPresenter))
		},
		controller.NewSuecaWebController)
	BindWebControllerFor("fortyfives",
		func() usecase.FortyFivesInteractorIF {
			return usecase.NewFortyFivesInteractor(domain.NewDefaultFortyFives(), new(presenter.FortyFivesWebPresenter))
		},
		controller.NewFortyFivesWebController)
	BindWebControllerFor("twentynine",
		func() usecase.TwentyNineInteractorIF {
			return usecase.NewTwentyNineInteractor(domain.NewDefaultTwentyNine(), new(presenter.TwentyNineWebPresenter))
		},
		controller.NewTwentyNineWebController)
	BindWebControllerFor("klaverjas",
		func() usecase.KlaverjasInteractorIF {
			return usecase.NewKlaverjasInteractor(domain.NewDefaultKlaverjas(), new(presenter.KlaverjasWebPresenter))
		},
		controller.NewKlaverjasWebController)
	BindWebControllerFor("manille",
		func() usecase.ManilleInteractorIF {
			return usecase.NewManilleInteractor(domain.NewDefaultManille(), new(presenter.ManilleWebPresenter))
		},
		controller.NewManilleWebController)
	BindWebControllerFor("marias",
		func() usecase.MariasInteractorIF {
			return usecase.NewMariasInteractor(domain.NewDefaultMarias(), new(presenter.MariasWebPresenter))
		},
		controller.NewMariasWebController)
	BindWebControllerFor("sedma",
		func() usecase.SedmaInteractorIF {
			return usecase.NewSedmaInteractor(domain.NewDefaultSedma(), new(presenter.SedmaWebPresenter))
		},
		controller.NewSedmaWebController)
	BindWebControllerFor("solowhist",
		func() usecase.SoloWhistInteractorIF {
			return usecase.NewSoloWhistInteractor(domain.NewDefaultSoloWhist(), new(presenter.SoloWhistWebPresenter))
		},
		controller.NewSoloWhistWebController)
	BindWebControllerFor("knockoutwhist",
		func() usecase.KnockoutWhistInteractorIF {
			return usecase.NewKnockoutWhistInteractor(domain.NewDefaultKnockoutWhist(), new(presenter.KnockoutWhistWebPresenter))
		},
		controller.NewKnockoutWhistWebController)
	BindWebControllerFor("nap",
		func() usecase.NapInteractorIF {
			return usecase.NewNapInteractor(domain.NewDefaultNap(), new(presenter.NapWebPresenter))
		},
		controller.NewNapWebController)
	BindWebControllerFor("preference",
		func() usecase.PreferenceInteractorIF {
			return usecase.NewPreferenceInteractor(domain.NewDefaultPreference(), new(presenter.PreferenceWebPresenter))
		},
		controller.NewPreferenceWebController)
	BindWebControllerFor("ganjifa",
		func() usecase.GanjifaInteractorIF {
			return usecase.NewGanjifaInteractor(domain.NewDefaultGanjifa(), new(presenter.GanjifaWebPresenter))
		},
		controller.NewGanjifaWebController)
	BindWebControllerFor("vira",
		func() usecase.ViraInteractorIF {
			return usecase.NewViraInteractor(domain.NewDefaultVira(), new(presenter.ViraWebPresenter))
		},
		controller.NewViraWebController)
	BindWebControllerFor("spoilfive",
		func() usecase.SpoilFiveInteractorIF {
			return usecase.NewSpoilFiveInteractor(domain.NewDefaultSpoilFive(), new(presenter.SpoilFiveWebPresenter))
		},
		controller.NewSpoilFiveWebController)
	BindWebControllerFor("easthaven",
		func() usecase.EasthavenInteractorIF {
			return usecase.NewEasthavenInteractor(domain.NewDefaultEasthaven(), new(presenter.EasthavenWebPresenter))
		},
		controller.NewEasthavenWebController)
	BindWebControllerFor("tichu",
		func() usecase.TichuInteractorIF {
			return usecase.NewTichuInteractor(domain.NewDefaultTichu(), new(presenter.TichuWebPresenter))
		},
		controller.NewTichuWebController)
	BindWebControllerFor("bourre",
		func() usecase.BourreInteractorIF {
			return usecase.NewBourreInteractor(domain.NewDefaultBourre(), new(presenter.BourreWebPresenter))
		},
		controller.NewBourreWebController)
	BindWebControllerFor("courtpiece",
		func() usecase.CourtPieceInteractorIF {
			return usecase.NewCourtPieceInteractor(domain.NewDefaultCourtPiece(), new(presenter.CourtPieceWebPresenter))
		},
		controller.NewCourtPieceWebController)
	BindWebControllerFor("bezique",
		func() usecase.BeziqueInteractorIF {
			return usecase.NewBeziqueInteractor(domain.NewDefaultBezique(), new(presenter.BeziqueWebPresenter))
		},
		controller.NewBeziqueWebController)
	BindWebControllerFor("ecarte",
		func() usecase.EcarteInteractorIF {
			return usecase.NewEcarteInteractor(domain.NewDefaultEcarte(), new(presenter.EcarteWebPresenter))
		},
		controller.NewEcarteWebController)
	BindWebControllerFor("threecardbrag",
		func() usecase.ThreeCardBragInteractorIF {
			return usecase.NewThreeCardBragInteractor(domain.NewDefaultThreeCardBrag(), new(presenter.ThreeCardBragWebPresenter))
		},
		controller.NewThreeCardBragWebController)
	BindWebControllerFor("teenpatti",
		func() usecase.TeenPattiInteractorIF {
			return usecase.NewTeenPattiInteractor(domain.NewDefaultTeenPatti(), new(presenter.TeenPattiWebPresenter))
		},
		controller.NewTeenPattiWebController)
	BindWebControllerFor("scopone",
		func() usecase.ScoponeInteractorIF {
			return usecase.NewScoponeInteractor(domain.NewDefaultScopone(), new(presenter.ScoponeWebPresenter))
		},
		controller.NewScoponeWebController)
	BindWebControllerFor("escoba",
		func() usecase.EscobaInteractorIF {
			return usecase.NewEscobaInteractor(domain.NewDefaultEscoba(), new(presenter.EscobaWebPresenter))
		},
		controller.NewEscobaWebController)
	BindWebControllerFor("handandfoot",
		func() usecase.HandAndFootInteractorIF {
			return usecase.NewHandAndFootInteractor(domain.NewDefaultHandAndFoot(), new(presenter.HandAndFootWebPresenter))
		},
		controller.NewHandAndFootWebController)
	BindWebControllerFor("conquian",
		func() usecase.ConquianInteractorIF {
			return usecase.NewConquianInteractor(domain.NewDefaultConquian(), new(presenter.ConquianWebPresenter))
		},
		controller.NewConquianWebController)
	BindWebControllerFor("chinchon",
		func() usecase.ChinchonInteractorIF {
			return usecase.NewChinchonInteractor(domain.NewDefaultChinchon(), new(presenter.ChinchonWebPresenter))
		},
		controller.NewChinchonWebController)
	BindWebControllerFor("kalooki",
		func() usecase.KalookiInteractorIF {
			return usecase.NewKalookiInteractor(domain.NewDefaultKalooki(), new(presenter.KalookiWebPresenter))
		},
		controller.NewKalookiWebController)
	BindWebControllerFor("threethirteen",
		func() usecase.ThreeThirteenInteractorIF {
			return usecase.NewThreeThirteenInteractor(domain.NewDefaultThreeThirteen(), new(presenter.ThreeThirteenWebPresenter))
		},
		controller.NewThreeThirteenWebController)
	BindWebControllerFor("mao",
		func() usecase.MaoInteractorIF {
			return usecase.NewMaoInteractor(domain.NewDefaultMao(), new(presenter.MaoWebPresenter))
		},
		controller.NewMaoWebController)
	BindWebControllerFor("spoons",
		func() usecase.SpoonsInteractorIF {
			return usecase.NewSpoonsInteractor(domain.NewDefaultSpoons(), new(presenter.SpoonsWebPresenter))
		},
		controller.NewSpoonsWebController)
	BindWebControllerFor("kemps",
		func() usecase.KempsInteractorIF {
			return usecase.NewKempsInteractor(domain.NewDefaultKemps(), new(presenter.KempsWebPresenter))
		},
		controller.NewKempsWebController)
	BindWebControllerFor("cuckoo",
		func() usecase.CuckooInteractorIF {
			return usecase.NewCuckooInteractor(domain.NewDefaultCuckoo(), new(presenter.CuckooWebPresenter))
		},
		controller.NewCuckooWebController)
	BindWebControllerFor("pishti",
		func() usecase.PishtiInteractorIF {
			return usecase.NewPishtiInteractor(domain.NewDefaultPishti(), new(presenter.PishtiWebPresenter))
		},
		controller.NewPishtiWebController)
	BindWebControllerFor("cuarenta",
		func() usecase.CuarentaInteractorIF {
			return usecase.NewCuarentaInteractor(domain.NewDefaultCuarenta(), new(presenter.CuarentaWebPresenter))
		},
		controller.NewCuarentaWebController)
	BindWebControllerFor("fivecardstud",
		func() usecase.FiveCardStudInteractorIF {
			return usecase.NewFiveCardStudInteractor(domain.NewDefaultFiveCardStud(), new(presenter.FiveCardStudWebPresenter))
		},
		controller.NewFiveCardStudWebController)
	BindWebControllerFor("faro",
		func() usecase.FaroInteractorIF {
			return usecase.NewFaroInteractor(domain.NewDefaultFaro(), new(presenter.FaroWebPresenter))
		},
		controller.NewFaroWebController)
	BindWebControllerFor("openfacechinese",
		func() usecase.OpenFaceChineseInteractorIF {
			return usecase.NewOpenFaceChineseInteractor(domain.NewDefaultOpenFaceChinese(), new(presenter.OpenFaceChineseWebPresenter))
		},
		controller.NewOpenFaceChineseWebController)
	BindWebControllerFor("beggarmyneighbour",
		func() usecase.BeggarMyNeighbourInteractorIF {
			return usecase.NewBeggarMyNeighbourInteractor(domain.NewDefaultBeggarMyNeighbour(), new(presenter.BeggarMyNeighbourWebPresenter))
		},
		controller.NewBeggarMyNeighbourWebController)
	BindWebControllerFor("allfours",
		func() usecase.AllFoursInteractorIF {
			return usecase.NewAllFoursInteractor(domain.NewDefaultAllFours(), new(presenter.AllFoursWebPresenter))
		},
		controller.NewAllFoursWebController)
	BindWebControllerFor("prsi",
		func() usecase.PrsiInteractorIF {
			return usecase.NewPrsiInteractor(domain.NewDefaultPrsi(), new(presenter.PrsiWebPresenter))
		},
		controller.NewPrsiWebController)
	BindWebControllerFor("jass",
		func() usecase.JassInteractorIF {
			return usecase.NewJassInteractor(domain.NewDefaultJass(), new(presenter.JassWebPresenter))
		},
		controller.NewJassWebController)
	BindWebControllerFor("gaigel",
		func() usecase.GaigelInteractorIF {
			return usecase.NewGaigelInteractor(domain.NewDefaultGaigel(), new(presenter.GaigelWebPresenter))
		},
		controller.NewGaigelWebController)
	BindWebControllerFor("tysiac",
		func() usecase.TysiacInteractorIF {
			return usecase.NewTysiacInteractor(domain.NewDefaultTysiac(), new(presenter.TysiacWebPresenter))
		},
		controller.NewTysiacWebController)
	BindWebControllerFor("calabresella",
		func() usecase.CalabresellaInteractorIF {
			return usecase.NewCalabresellaInteractor(domain.NewDefaultCalabresella(), new(presenter.CalabresellaWebPresenter))
		},
		controller.NewCalabresellaWebController)
	BindWebControllerFor("ombre",
		func() usecase.OmbreInteractorIF {
			return usecase.NewOmbreInteractor(domain.NewDefaultOmbre(), new(presenter.OmbreWebPresenter))
		},
		controller.NewOmbreWebController)
	BindWebControllerFor("ulti",
		func() usecase.UltiInteractorIF {
			return usecase.NewUltiInteractor(domain.NewDefaultUlti(), new(presenter.UltiWebPresenter))
		},
		controller.NewUltiWebController)
	BindWebControllerFor("king",
		func() usecase.KingInteractorIF {
			return usecase.NewKingInteractor(domain.NewDefaultKing(), new(presenter.KingWebPresenter))
		},
		controller.NewKingWebController)
	BindWebControllerFor("cinch",
		func() usecase.CinchInteractorIF {
			return usecase.NewCinchInteractor(domain.NewDefaultCinch(), new(presenter.CinchWebPresenter))
		},
		controller.NewCinchWebController)
	BindWebControllerFor("loo",
		func() usecase.LooInteractorIF {
			return usecase.NewLooInteractor(domain.NewDefaultLoo(), new(presenter.LooWebPresenter))
		},
		controller.NewLooWebController)
	BindWebControllerFor("basra",
		func() usecase.BasraInteractorIF {
			return usecase.NewBasraInteractor(domain.NewDefaultBasra(), new(presenter.BasraWebPresenter))
		},
		controller.NewBasraWebController)
	BindWebControllerFor("tablanet",
		func() usecase.TablanetInteractorIF {
			return usecase.NewTablanetInteractor(domain.NewDefaultTablanet(), new(presenter.TablanetWebPresenter))
		},
		controller.NewTablanetWebController)
	BindWebControllerFor("trenteetquarante",
		func() usecase.TrenteEtQuaranteInteractorIF {
			return usecase.NewTrenteEtQuaranteInteractor(domain.NewDefaultTrenteEtQuarante(), new(presenter.TrenteEtQuaranteWebPresenter))
		},
		controller.NewTrenteEtQuaranteWebController)
	BindWebControllerFor("guts",
		func() usecase.GutsInteractorIF {
			return usecase.NewGutsInteractor(domain.NewDefaultGuts(), new(presenter.GutsWebPresenter))
		},
		controller.NewGutsWebController)
	BindWebControllerFor("bouillotte",
		func() usecase.BouillotteInteractorIF {
			return usecase.NewBouillotteInteractor(domain.NewDefaultBouillotte(), new(presenter.BouillotteWebPresenter))
		},
		controller.NewBouillotteWebController)
	BindWebControllerFor("primero",
		func() usecase.PrimeroInteractorIF {
			return usecase.NewPrimeroInteractor(domain.NewDefaultPrimero(), new(presenter.PrimeroWebPresenter))
		},
		controller.NewPrimeroWebController)
	BindWebControllerFor("michigan",
		func() usecase.MichiganInteractorIF {
			return usecase.NewMichiganInteractor(domain.NewDefaultMichigan(), new(presenter.MichiganWebPresenter))
		},
		controller.NewMichiganWebController)
	BindWebControllerFor("watten",
		func() usecase.WattenInteractorIF {
			return usecase.NewWattenInteractor(domain.NewDefaultWatten(), new(presenter.WattenWebPresenter))
		},
		controller.NewWattenWebController)
	BindWebControllerFor("carioca",
		func() usecase.CariocaInteractorIF {
			return usecase.NewCariocaInteractor(domain.NewDefaultCarioca(), new(presenter.CariocaWebPresenter))
		},
		controller.NewCariocaWebController)
	BindWebControllerFor("samba",
		func() usecase.SambaInteractorIF {
			return usecase.NewSambaInteractor(domain.NewDefaultSamba(), new(presenter.SambaWebPresenter))
		},
		controller.NewSambaWebController)
	BindWebControllerFor("anaconda",
		func() usecase.AnacondaInteractorIF {
			return usecase.NewAnacondaInteractor(domain.NewDefaultAnaconda(), new(presenter.AnacondaWebPresenter))
		},
		controller.NewAnacondaWebController)
	BindWebControllerFor("machiavelli",
		func() usecase.MachiavelliInteractorIF {
			return usecase.NewMachiavelliInteractor(domain.NewDefaultMachiavelli(), new(presenter.MachiavelliWebPresenter))
		},
		controller.NewMachiavelliWebController)
	BindWebControllerFor("pan",
		func() usecase.PanInteractorIF {
			return usecase.NewPanInteractor(domain.NewDefaultPan(), new(presenter.PanWebPresenter))
		},
		controller.NewPanWebController)
	BindWebControllerFor("wizard",
		func() usecase.WizardInteractorIF {
			return usecase.NewWizardInteractor(domain.NewDefaultWizard(), new(presenter.WizardWebPresenter))
		},
		controller.NewWizardWebController)
	BindWebControllerFor("oichokabu",
		func() usecase.OichoKabuInteractorIF {
			return usecase.NewOichoKabuInteractor(domain.NewDefaultOichoKabu(), new(presenter.OichoKabuWebPresenter))
		},
		controller.NewOichoKabuWebController)
	BindWebControllerFor("kingo",
		func() usecase.KingoInteractorIF {
			return usecase.NewKingoInteractor(domain.NewDefaultKingo(), new(presenter.KingoWebPresenter))
		},
		controller.NewKingoWebController)
	BindWebControllerFor("tusac",
		func() usecase.TuSacInteractorIF {
			return usecase.NewTuSacInteractor(domain.NewDefaultTuSac(), new(presenter.TuSacWebPresenter))
		},
		controller.NewTuSacWebController)
	BindWebControllerFor("rook",
		func() usecase.RookInteractorIF {
			return usecase.NewRookInteractor(domain.NewDefaultRook(), new(presenter.RookWebPresenter))
		},
		controller.NewRookWebController)
	BindWebControllerFor("koikoi",
		func() usecase.KoiKoiInteractorIF {
			return usecase.NewKoiKoiInteractor(domain.NewDefaultKoiKoi(), new(presenter.KoiKoiWebPresenter))
		},
		controller.NewKoiKoiWebController)
	BindWebControllerFor("gostop",
		func() usecase.GoStopInteractorIF {
			return usecase.NewGoStopInteractor(domain.NewDefaultGoStop(), new(presenter.GoStopWebPresenter))
		},
		controller.NewGoStopWebController)
	BindWebControllerFor("hachihachi",
		func() usecase.HachiHachiInteractorIF {
			return usecase.NewHachiHachiInteractor(domain.NewDefaultHachiHachi(), new(presenter.HachiHachiWebPresenter))
		},
		controller.NewHachiHachiWebController)
	BindWebControllerFor("frenchtarot",
		func() usecase.FrenchTarotInteractorIF {
			return usecase.NewFrenchTarotInteractor(domain.NewDefaultFrenchTarot(), new(presenter.FrenchTarotWebPresenter))
		},
		controller.NewFrenchTarotWebController)
	BindWebControllerFor("koenigrufen",
		func() usecase.KoenigrufenInteractorIF {
			return usecase.NewKoenigrufenInteractor(domain.NewDefaultKoenigrufen(), new(presenter.KoenigrufenWebPresenter))
		},
		controller.NewKoenigrufenWebController)
	BindWebControllerFor("aluette",
		func() usecase.AluetteInteractorIF {
			return usecase.NewAluetteInteractor(domain.NewDefaultAluette(), new(presenter.AluetteWebPresenter))
		},
		controller.NewAluetteWebController)
	BindWebControllerFor("minchiate",
		func() usecase.MinchiateInteractorIF {
			return usecase.NewMinchiateInteractor(domain.NewDefaultMinchiate(), new(presenter.MinchiateWebPresenter))
		},
		controller.NewMinchiateWebController)
	BindWebControllerFor("tarocchini",
		func() usecase.TarocchiniInteractorIF {
			return usecase.NewTarocchiniInteractor(domain.NewDefaultTarocchini(), new(presenter.TarocchiniWebPresenter))
		},
		controller.NewTarocchiniWebController)
	BindWebControllerFor("scarto",
		func() usecase.ScartoInteractorIF {
			return usecase.NewScartoInteractor(domain.NewDefaultScarto(), new(presenter.ScartoWebPresenter))
		},
		controller.NewScartoWebController)
	BindWebControllerFor("cego",
		func() usecase.CegoInteractorIF {
			return usecase.NewCegoInteractor(domain.NewDefaultCego(), new(presenter.CegoWebPresenter))
		},
		controller.NewCegoWebController)
	BindWebControllerFor("zheng",
		func() usecase.ZhengInteractorIF {
			return usecase.NewZhengInteractor(domain.NewDefaultZheng(), new(presenter.ZhengWebPresenter))
		},
		controller.NewZhengWebController)
	BindWebControllerFor("desmoche",
		func() usecase.DesmocheInteractorIF {
			return usecase.NewDesmocheInteractor(domain.NewDefaultDesmoche(), new(presenter.DesmocheWebPresenter))
		},
		controller.NewDesmocheWebController)
	BindWebControllerFor("zwicker",
		func() usecase.ZwickerInteractorIF {
			return usecase.NewZwickerInteractor(domain.NewDefaultZwicker(), new(presenter.ZwickerWebPresenter))
		},
		controller.NewZwickerWebController)
	BindWebControllerFor("poch",
		func() usecase.PochInteractorIF {
			return usecase.NewPochInteractor(domain.NewDefaultPoch(), new(presenter.PochWebPresenter))
		},
		controller.NewPochWebController)
	BindWebControllerFor("popejoan",
		func() usecase.PopeJoanInteractorIF {
			return usecase.NewPopeJoanInteractor(domain.NewDefaultPopeJoan(), new(presenter.PopeJoanWebPresenter))
		},
		controller.NewPopeJoanWebController)
	BindWebControllerFor("nainjaune",
		func() usecase.NainJauneInteractorIF {
			return usecase.NewNainJauneInteractor(domain.NewDefaultNainJaune(), new(presenter.NainJauneWebPresenter))
		},
		controller.NewNainJauneWebController)
	BindWebControllerFor("kille",
		func() usecase.KilleInteractorIF {
			return usecase.NewKilleInteractor(domain.NewDefaultKille(), new(presenter.KilleWebPresenter))
		},
		controller.NewKilleWebController)
	BindWebControllerFor("klaberjass",
		func() usecase.KlaberjassInteractorIF {
			return usecase.NewKlaberjassInteractor(domain.NewDefaultKlaberjass(), new(presenter.KlaberjassWebPresenter))
		},
		controller.NewKlaberjassWebController)
	BindWebControllerFor("kaiser",
		func() usecase.KaiserInteractorIF {
			return usecase.NewKaiserInteractor(domain.NewDefaultKaiser(), new(presenter.KaiserWebPresenter))
		},
		controller.NewKaiserWebController)
	BindWebControllerFor("boston",
		func() usecase.BostonInteractorIF {
			return usecase.NewBostonInteractor(domain.NewDefaultBoston(), new(presenter.BostonWebPresenter))
		},
		controller.NewBostonWebController)
	BindWebControllerFor("vint",
		func() usecase.VintInteractorIF {
			return usecase.NewVintInteractor(domain.NewDefaultVint(), new(presenter.VintWebPresenter))
		},
		controller.NewVintWebController)
	BindWebControllerFor("bideuchre",
		func() usecase.BidEuchreInteractorIF {
			return usecase.NewBidEuchreInteractor(domain.NewDefaultBidEuchre(), new(presenter.BidEuchreWebPresenter))
		},
		controller.NewBidEuchreWebController)
	BindWebControllerFor("sixbidsolo",
		func() usecase.SixBidSoloInteractorIF {
			return usecase.NewSixBidSoloInteractor(domain.NewDefaultSixBidSolo(), new(presenter.SixBidSoloWebPresenter))
		},
		controller.NewSixBidSoloWebController)
	BindWebControllerFor("karnoffel",
		func() usecase.KarnoffelInteractorIF {
			return usecase.NewKarnoffelInteractor(domain.NewDefaultKarnoffel(), new(presenter.KarnoffelWebPresenter))
		},
		controller.NewKarnoffelWebController)
	BindWebControllerFor("literature",
		func() usecase.LiteratureInteractorIF {
			return usecase.NewLiteratureInteractor(domain.NewDefaultLiterature(), new(presenter.LiteratureWebPresenter))
		},
		controller.NewLiteratureWebController)
	BindWebControllerFor("guandan",
		func() usecase.GuandanInteractorIF {
			return usecase.NewGuandanInteractor(domain.NewDefaultGuandan(), new(presenter.GuandanWebPresenter))
		},
		controller.NewGuandanWebController)
	BindWebControllerFor("freebet",
		func() usecase.FreeBetBlackjackInteractorIF {
			return usecase.NewFreeBetBlackjackInteractor(domain.NewDefaultFreeBetBlackjack(), new(presenter.FreeBetBlackjackWebPresenter))
		},
		controller.NewFreeBetBlackjackWebController)
	BindWebControllerFor("banluck",
		func() usecase.BanLuckInteractorIF {
			return usecase.NewBanLuckInteractor(domain.NewDefaultBanLuck(), new(presenter.BanLuckWebPresenter))
		},
		controller.NewBanLuckWebController)
	BindWebControllerFor("montebank",
		func() usecase.MonteBankInteractorIF {
			return usecase.NewMonteBankInteractor(domain.NewDefaultMonteBank(), new(presenter.MonteBankWebPresenter))
		},
		controller.NewMonteBankWebController)
	BindWebControllerFor("cincinnati",
		func() usecase.CincinnatiInteractorIF {
			return usecase.NewCincinnatiInteractor(domain.NewDefaultCincinnati(), new(presenter.CincinnatiWebPresenter))
		},
		controller.NewCincinnatiWebController)
	BindWebControllerFor("ironcross",
		func() usecase.IronCrossInteractorIF {
			return usecase.NewIronCrossInteractor(domain.NewDefaultIronCross(), new(presenter.IronCrossWebPresenter))
		},
		controller.NewIronCrossWebController)
	BindWebControllerFor("baseballpoker",
		func() usecase.BaseballPokerInteractorIF {
			return usecase.NewBaseballPokerInteractor(domain.NewDefaultBaseballPoker(), new(presenter.BaseballPokerWebPresenter))
		},
		controller.NewBaseballPokerWebController)
	BindWebControllerFor("shengji",
		func() usecase.ShengJiInteractorIF {
			return usecase.NewShengJiInteractor(domain.NewDefaultShengJi(), new(presenter.ShengJiWebPresenter))
		},
		controller.NewShengJiWebController)
	BindWebControllerFor("sakura",
		func() usecase.SakuraInteractorIF {
			return usecase.NewSakuraInteractor(domain.NewDefaultSakura(), new(presenter.SakuraWebPresenter))
		},
		controller.NewSakuraWebController)
	BindWebControllerFor("zwanzigerrufen",
		func() usecase.ZwanzigerrufenInteractorIF {
			return usecase.NewZwanzigerrufenInteractor(domain.NewDefaultZwanzigerrufen(), new(presenter.ZwanzigerrufenWebPresenter))
		},
		controller.NewZwanzigerrufenWebController)
	BindWebControllerFor("troggu",
		func() usecase.TrogguInteractorIF {
			return usecase.NewTrogguInteractor(domain.NewDefaultTroggu(), new(presenter.TrogguWebPresenter))
		},
		controller.NewTrogguWebController)
	BindWebControllerFor("horse",
		func() usecase.HorseInteractorIF {
			return usecase.NewHorseInteractor(domain.NewDefaultHorse(), new(presenter.HorseWebPresenter))
		},
		controller.NewHorseWebController)
	BindWebControllerFor("ramsch",
		func() usecase.RamschInteractorIF {
			return usecase.NewRamschInteractor(domain.NewDefaultRamsch(), new(presenter.RamschWebPresenter))
		},
		controller.NewRamschWebController)
	BindWebControllerFor("seventwentyseven",
		func() usecase.SevenTwentySevenInteractorIF {
			return usecase.NewSevenTwentySevenInteractor(domain.NewDefaultSevenTwentySeven(), new(presenter.SevenTwentySevenWebPresenter))
		},
		controller.NewSevenTwentySevenWebController)
	BindWebControllerFor("threecardrummy",
		func() usecase.ThreeCardRummyInteractorIF {
			return usecase.NewThreeCardRummyInteractor(domain.NewDefaultThreeCardRummy(), new(presenter.ThreeCardRummyWebPresenter))
		},
		controller.NewThreeCardRummyWebController)
	BindWebControllerFor("caribbeandraw",
		func() usecase.CaribbeanDrawInteractorIF {
			return usecase.NewCaribbeanDrawInteractor(domain.NewDefaultCaribbeanDraw(), new(presenter.CaribbeanDrawWebPresenter))
		},
		controller.NewCaribbeanDrawWebController)
}
