package ui

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GameRegistryEntry holds a game's name and CUI constructor.
// Description lives on games.Game (issue #1459 SSoT); use Description() to
// look it up for a given entry.
type GameRegistryEntry struct {
	Name   string
	NewCui func() cuiGame
}

// Description returns the display description for this entry, sourced from
// the games package SSoT. Uses the single-entry lookup so calling this in a
// loop does not repeatedly touch the descriptions map.
func (e GameRegistryEntry) Description() string {
	return games.Description(e.Name)
}

// cuiEntry builds a generic CUI game from a controller and a help spec.
// Used by gameRegistry entries that follow the standard help template.
func cuiEntry(ctrl CuiExecer, spec CuiHelpSpec) cuiGame {
	return newCuiGame(ctrl, BuildCuiHelp(spec))
}

// Shared poker-family help keys live here so a label change (e.g. renaming
// "small blind" wording) updates every entry that uses them. See issue #1511.
var (
	// tournamentRebuyAddOnKeys covers the rebuy / add-on row used by every
	// tournament-capable poker game.
	tournamentRebuyAddOnKeys = []string{
		"tournament.helpRebuy",
		"tournament.helpSkipRebuy",
		"tournament.helpAddOn",
		"tournament.helpSkipAddOn",
	}
	// pineappleRebuyAddOnKeys is the same row prefixed with the discard line
	// only Pineapple variants surface. Spelled out as a flat literal (rather
	// than `append([]string{...}, tournamentRebuyAddOnKeys...)`) so the keys
	// don't depend on package-level init order and can't accidentally alias
	// the shared slice.
	pineappleRebuyAddOnKeys = []string{
		"tournament.helpDiscard",
		"tournament.helpRebuy",
		"tournament.helpSkipRebuy",
		"tournament.helpAddOn",
		"tournament.helpSkipAddOn",
	}
	// holdemBlindKeys are the blind / level-up / table-size settings shared
	// by holdem, omaha, shortdeck, pineapple, crazypineapple.
	holdemBlindKeys = []string{
		"tournament.helpSmallBlind",
		"tournament.helpBigBlind",
		"tournament.helpLevelUpHands",
		"tournament.helpTableSize",
	}
	// studAnteKeys are the ante / bring-in settings shared by sevencardstud
	// and razz (both render through the SevenCardStud controller).
	studAnteKeys = []string{
		"stud.helpAnte",
		"stud.helpBringIn",
		"stud.helpSmallBet",
		"stud.helpBigBet",
		"stud.helpLevelUpHands",
		"stud.helpTableSize",
	}
)

// gameRegistry wires each game's CUI constructor. Name and Description are
// carried by the games package (see internal/infrastructure/games/registry.go)
// — this slice only holds the CLI-specific NewCui factory. Order mirrors the
// games registry; drift is enforced by TestRegistryMatchesCLI.
var gameRegistry = []GameRegistryEntry{
	{Name: "blackjack", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBlackJackCuiController(usecase.NewBlackJackInteractor(
				domain.NewDefaultBlackJack(), new(presenter.BlackJackCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "blackjack.helpTitle",
				CommandKeys: []string{
					"blackjack.helpBet",
					"blackjack.helpHit",
					"blackjack.helpStand",
					"blackjack.helpDouble",
					"blackjack.helpSplit",
					"blackjack.helpInsurance",
					"blackjack.helpDeclineInsurance",
				},
				SettingKeys: []string{"blackjack.helpSetCpuCount"},
			})
	}},
	{Name: "poker", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPokerCuiController(usecase.NewPokerInteractor(
				domain.NewDefaultPoker(), new(presenter.PokerCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "poker.helpTitle",
				CommandKeys: []string{
					"poker.helpBet",
					"poker.helpCall",
					"poker.helpRaise",
					"poker.helpCheck",
					"poker.helpFold",
					"poker.helpAllIn",
					"poker.helpExchange",
					"poker.helpStand",
				},
				SettingKeys: []string{"poker.helpBettingLimit", "poker.helpLowball"},
			})
	}},
	{Name: "oldmaid", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewOldMaidCuiController(usecase.NewOldMaidInteractor(
				domain.NewDefaultOldMaid(), new(presenter.OldMaidCuiPresenter))),
			CuiHelpSpec{
				TitleKey:    "oldmaid.helpTitle",
				CommandKeys: []string{"oldmaid.helpDraw", "oldmaid.helpShuffle", "oldmaid.helpReorder"},
				SettingKeys: []string{"oldmaid.helpSetMode", "oldmaid.helpSetPlacement", "oldmaid.helpSetMemoryAI"},
			})
	}},
	{Name: "daifugo", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewDaifugoCuiController(usecase.NewDaifugoInteractor(
				domain.NewDefaultDaifugo(), new(presenter.DaifugoCuiPresenter))),
			CuiHelpSpec{
				TitleKey:    "daifugo.helpTitle",
				CommandKeys: []string{"daifugo.helpPlay", "daifugo.helpSort"},
				SettingKeys: []string{"daifugo.helpSetDifficulty", "daifugo.helpSetJoker", "daifugo.helpSetRule"},
			})
	}},
	{Name: "sevens", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSevensCuiController(usecase.NewSevensInteractor(
				domain.NewDefaultSevens(), new(presenter.SevensCuiPresenter))),
			CuiHelpSpec{
				TitleKey:      "sevens.helpTitle",
				CommandKeys:   []string{"sevens.helpPlay"},
				ResetOverride: "  r [tunnel] [joker=N] [strategy] [passes=N]  reset with options",
			})
	}},
	{Name: "doubt", NewCui: func() cuiGame { return NewDoubtCui() }},
	{Name: "holdem", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewHoldemCuiController(usecase.NewHoldemInteractor(
				domain.NewDefaultHoldem(), new(presenter.HoldemCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "holdem.helpTitle",
				CommandKeys: append([]string{
					"holdem.helpFold",
					"holdem.helpCheck",
					"holdem.helpCall",
					"holdem.helpBet",
					"holdem.helpRaise",
					"holdem.helpAllIn",
				}, tournamentRebuyAddOnKeys...),
				SettingKeys: append([]string{
					"holdem.helpBettingLimit", "holdem.helpTournament",
				}, holdemBlindKeys...),
			})
	}},
	{Name: "omaha", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewOmahaCuiController(usecase.NewOmahaInteractor(
				domain.NewDefaultOmaha(), new(presenter.OmahaCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "omaha.helpTitle",
				CommandKeys: append([]string{
					"omaha.helpFold",
					"omaha.helpCheck",
					"omaha.helpCall",
					"omaha.helpBet",
					"omaha.helpRaise",
					"omaha.helpAllIn",
				}, tournamentRebuyAddOnKeys...),
				SettingKeys: append([]string{
					"omaha.helpBettingLimit", "omaha.helpTournament",
				}, holdemBlindKeys...),
			})
	}},
	{Name: "omahahilo", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewOmahaCuiController(usecase.NewOmahaInteractor(
				domain.NewDefaultOmahaHiLo(), new(presenter.OmahaCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "omahahilo.helpTitle",
				CommandKeys: append([]string{
					"omaha.helpFold",
					"omaha.helpCheck",
					"omaha.helpCall",
					"omaha.helpBet",
					"omaha.helpRaise",
					"omaha.helpAllIn",
				}, tournamentRebuyAddOnKeys...),
				SettingKeys: append([]string{
					"omaha.helpBettingLimit", "omaha.helpTournament",
				}, holdemBlindKeys...),
			})
	}},
	{Name: "shortdeck", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewShortDeckCuiController(usecase.NewShortDeckInteractor(
				domain.NewDefaultShortDeck(), new(presenter.ShortDeckCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "shortdeck.helpTitle",
				CommandKeys: append([]string{
					"shortdeck.helpFold",
					"shortdeck.helpCheck",
					"shortdeck.helpCall",
					"shortdeck.helpBet",
					"shortdeck.helpRaise",
					"shortdeck.helpAllIn",
				}, tournamentRebuyAddOnKeys...),
				SettingKeys: append([]string{
					"shortdeck.helpBettingLimit", "shortdeck.helpTournament",
				}, holdemBlindKeys...),
			})
	}},
	{Name: "pineapple", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPineappleCuiController(usecase.NewPineappleInteractor(
				domain.NewDefaultPineapple(), new(presenter.PineappleCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "pineapple.helpTitle",
				CommandKeys: append([]string{
					"pineapple.helpFold",
					"pineapple.helpCheck",
					"pineapple.helpCall",
					"pineapple.helpBet",
					"pineapple.helpRaise",
					"pineapple.helpAllIn",
				}, pineappleRebuyAddOnKeys...),
				SettingKeys: append([]string{
					"pineapple.helpBettingLimit", "pineapple.helpTournament",
				}, holdemBlindKeys...),
			})
	}},
	{Name: "crazypineapple", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPineappleCuiController(usecase.NewPineappleInteractor(
				domain.NewDefaultCrazyPineapple(), new(presenter.PineappleCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "crazypineapple.helpTitle",
				CommandKeys: append([]string{
					"crazypineapple.helpFold",
					"crazypineapple.helpCheck",
					"crazypineapple.helpCall",
					"crazypineapple.helpBet",
					"crazypineapple.helpRaise",
					"crazypineapple.helpAllIn",
				}, pineappleRebuyAddOnKeys...),
				SettingKeys: append([]string{
					"crazypineapple.helpBettingLimit", "crazypineapple.helpTournament",
				}, holdemBlindKeys...),
			})
	}},
	{Name: "hearts", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewHeartsCuiController(usecase.NewHeartsInteractor(
				domain.NewDefaultHearts(), new(presenter.HeartsCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "hearts.helpTitle",
				CommandKeys:       []string{"hearts.helpPass", "hearts.helpPlay", "hearts.helpNext", "hearts.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"hearts.helpSetDifficulty", "hearts.helpSetLimit"},
			})
	}},
	{Name: "memory", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewMemoryCuiController(usecase.NewMemoryInteractor(
				domain.NewDefaultMemory(), new(presenter.MemoryCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "memory.helpTitle",
				CommandKeys:       []string{"memory.helpFlip", "memory.helpNext"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"memory.helpSetDifficulty"},
			})
	}},
	{Name: "klondike", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewKlondikeCuiController(usecase.NewKlondikeInteractor(
				domain.NewDefaultKlondike(), new(presenter.KlondikeCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "klondike.helpTitle",
				CommandKeys: []string{
					"klondike.helpDraw",
					"klondike.helpMove",
					"klondike.helpMoveWF",
					"klondike.helpMoveTF",
					"klondike.helpMoveTT",
					"klondike.helpGiveUp",
					"klondike.helpHint",
					"klondike.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "freecell", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewFreeCellCuiController(usecase.NewFreeCellInteractor(
				domain.NewDefaultFreeCell(), new(presenter.FreeCellCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "freecell.helpTitle",
				CommandKeys: []string{
					"freecell.helpMove",
					"freecell.helpMoveTF",
					"freecell.helpMoveTT",
					"freecell.helpMoveTC",
					"freecell.helpMoveCT",
					"freecell.helpMoveCF",
					"freecell.helpGiveUp",
					"freecell.helpHint",
					"freecell.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "baccarat", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBaccaratCuiController(usecase.NewBaccaratInteractor(
				domain.NewDefaultBaccarat(), new(presenter.BaccaratCuiPresenter))),
			CuiHelpSpec{
				TitleKey:    "baccarat.helpTitle",
				CommandKeys: []string{"baccarat.helpBet"},
				ExtraCommandLines: []string{
					"  log                  action log",
					"  ch                   clear history",
				},
			})
	}},
	{Name: "spades", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSpadesCuiController(usecase.NewSpadesInteractor(
				domain.NewDefaultSpades(), new(presenter.SpadesCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "spades.helpTitle",
				CommandKeys:       []string{"spades.helpBid", "spades.helpPlay", "spades.helpNext", "spades.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"spades.helpSetDifficulty", "spades.helpSetLimit"},
			})
	}},
	{Name: "crazyeights", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCrazyEightsCuiController(usecase.NewCrazyEightsInteractor(
				domain.NewDefaultCrazyEights(), new(presenter.CrazyEightsCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "crazyeights.helpTitle",
				CommandKeys:       []string{"crazyeights.helpPlay", "crazyeights.helpDraw", "crazyeights.helpSuit", "crazyeights.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"crazyeights.helpSetDifficulty", "crazyeights.helpSetLimit"},
			})
	}},
	{Name: "ginrummy", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewGinRummyCuiController(usecase.NewGinRummyInteractor(
				domain.NewDefaultGinRummy(), new(presenter.GinRummyCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "ginrummy.helpTitle",
				CommandKeys: []string{
					"ginrummy.helpDrawStock",
					"ginrummy.helpDrawDiscard",
					"ginrummy.helpDiscard",
					"ginrummy.helpKnock",
					"ginrummy.helpLayoff",
					"ginrummy.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"ginrummy.helpSetDifficulty", "ginrummy.helpSetLimit"},
			})
	}},
	{Name: "canasta", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCanastaCuiController(usecase.NewCanastaInteractor(
				domain.NewDefaultCanasta(), new(presenter.CanastaCuiPresenter))),
			CuiHelpSpec{Body: []string{
				"Canasta (カナスタ) Help",
				"",
				"Game Commands:",
				"  ds                   draw from stock",
				"  dd <idx,idx>         pick up discard pile (natural pair indices)",
				"  m <idx,idx;idx,idx>  meld (semicolon-separated groups)",
				"  sm                   skip meld phase",
				"  d <idx>              discard a card",
				"  go                   go out (requires canasta)",
				"  nr                   next round",
				"  l                    action log",
				"",
				"Settings:",
				"  sd <0-2>             set CPU difficulty (0=Easy, 1=Normal, 2=Hard)",
				"  sl <n>               set point limit",
				"",
				"Session:",
				"  r / reset            reset game",
				"  q / quit             quit",
				"  ? / help             show help",
			}})
	}},
	{Name: "spider", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSpiderCuiController(usecase.NewSpiderInteractor(
				domain.NewDefaultSpider(), new(presenter.SpiderCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "spider.helpTitle",
				CommandKeys: []string{
					"spider.helpDeal",
					"spider.helpMove",
					"spider.helpGiveUp",
					"spider.helpHint",
					"spider.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "napoleon", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewNapoleonCuiController(usecase.NewNapoleonInteractor(
				domain.NewDefaultNapoleon(), new(presenter.NapoleonCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "napoleon.helpTitle",
				CommandKeys: []string{
					"napoleon.helpBid",
					"napoleon.helpTrump",
					"napoleon.helpExchange",
					"napoleon.helpPlay",
					"napoleon.helpNext",
					"napoleon.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"napoleon.helpSetDifficulty", "napoleon.helpSetLimit", "napoleon.helpSetMinBid"},
			})
	}},
	{Name: "indianpoker", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewIndianPokerCuiController(usecase.NewIndianPokerInteractor(
				domain.NewDefaultIndianPoker(), new(presenter.IndianPokerCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "indianpoker.helpTitle",
				CommandKeys: []string{
					"indianpoker.helpFold",
					"indianpoker.helpCheck",
					"indianpoker.helpCall",
					"indianpoker.helpBet",
					"indianpoker.helpRaise",
					"indianpoker.helpAllIn",
				},
				SettingKeys: []string{"indianpoker.helpAnte", "indianpoker.helpBettingLimit", "indianpoker.helpMetaAI"},
			})
	}},
	{Name: "videopoker", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewVideoPokerCuiController(usecase.NewVideoPokerInteractor(
				domain.NewDefaultVideoPoker(), new(presenter.VideoPokerCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "videopoker.helpTitle",
				CommandKeys:       []string{"videopoker.helpBet", "videopoker.helpHold"},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "deuceswild", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewVideoPokerCuiController(usecase.NewVideoPokerInteractor(
				domain.NewDeucesWildVideoPoker(), new(presenter.VideoPokerCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "deuceswild.helpTitle",
				CommandKeys:       []string{"videopoker.helpBet", "videopoker.helpHold"},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "jokerpoker", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewVideoPokerCuiController(usecase.NewVideoPokerInteractor(
				domain.NewJokerPokerVideoPoker(), new(presenter.VideoPokerCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "jokerpoker.helpTitle",
				CommandKeys:       []string{"videopoker.helpBet", "videopoker.helpHold"},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "euchre", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewEuchreCuiController(usecase.NewEuchreInteractor(
				domain.NewDefaultEuchre(), new(presenter.EuchreCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "euchre.helpTitle",
				CommandKeys: []string{
					"euchre.helpOrderUp",
					"euchre.helpOrderUpAlone",
					"euchre.helpPass",
					"euchre.helpCall",
					"euchre.helpCallAlone",
					"euchre.helpDiscard",
					"euchre.helpPlay",
					"euchre.helpNext",
					"euchre.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"euchre.helpSetDifficulty", "euchre.helpSetLimit"},
			})
	}},
	{Name: "pyramid", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPyramidCuiController(usecase.NewPyramidInteractor(
				domain.NewDefaultPyramid(), new(presenter.PyramidCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "pyramid.helpTitle",
				CommandKeys: []string{
					"pyramid.helpDraw",
					"pyramid.helpRemoveKing",
					"pyramid.helpRemovePair",
					"pyramid.helpRemoveWaste",
					"pyramid.helpRemoveWasteKing",
					"pyramid.helpGiveUp",
					"pyramid.helpHint",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "tripeaks", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewTriPeaksCuiController(usecase.NewTriPeaksInteractor(
				domain.NewDefaultTriPeaks(), new(presenter.TriPeaksCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "tripeaks.helpTitle",
				CommandKeys: []string{
					"tripeaks.helpDraw",
					"tripeaks.helpRemove",
					"tripeaks.helpGiveUp",
					"tripeaks.helpHint",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "cribbage", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCribbageCuiController(usecase.NewCribbageInteractor(
				domain.NewDefaultCribbage(), new(presenter.CribbageCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "cribbage.helpTitle",
				CommandKeys:       []string{"cribbage.helpDiscard", "cribbage.helpPeg", "cribbage.helpGo", "cribbage.helpShowNext", "cribbage.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"cribbage.helpSetDifficulty", "cribbage.helpSetLimit"},
			})
	}},
	{Name: "threecard", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewThreeCardCuiController(usecase.NewThreeCardInteractor(
				domain.NewDefaultThreeCard(), new(presenter.ThreeCardCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "threecard.helpTitle",
				CommandKeys:       []string{"threecard.helpBet", "threecard.helpPlay", "threecard.helpFold"},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "ohhell", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewOhHellCuiController(usecase.NewOhHellInteractor(
				domain.NewDefaultOhHell(), new(presenter.OhHellCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "ohhell.helpTitle",
				CommandKeys:       []string{"ohhell.helpBid", "ohhell.helpPlay", "ohhell.helpNext", "ohhell.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"ohhell.helpSetDifficulty", "ohhell.helpSetMaxHand"},
			})
	}},
	{Name: "bridge", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBridgeCuiController(usecase.NewBridgeInteractor(
				domain.NewDefaultBridge(), new(presenter.BridgeCuiPresenter))),
			CuiHelpSpec{Body: []string{
				"=== Contract Bridge ===",
				"",
				"Game Commands:",
				"  b <type> <level> <suit>  bid (type: 0=pass,1=bid,2=dbl,3=rdbl; level: 1-7; suit: 1-5)",
				"  p <index>                play a card",
				"  n                        next trick",
				"  nr                       next round (score & proceed)",
				"  h                        hint",
				"  l                        action log",
				"",
				"Settings:",
				"  sd <0-2>                 set CPU difficulty (0=Easy,1=Normal,2=Hard)",
				"",
				"Session:",
				"  r                        reset game",
				"  q                        quit",
				"  help                     show this help",
			}})
	}},
	{Name: "speed", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSpeedCuiController(usecase.NewSpeedInteractor(
				domain.NewDefaultSpeed(), new(presenter.SpeedCuiPresenter))),
			CuiHelpSpec{
				TitleKey:    "speed.helpTitle",
				CommandKeys: []string{"speed.helpPlay", "speed.helpFlip", "speed.helpHint"},
				SettingKeys: []string{"speed.helpSetDifficulty"},
			})
	}},
	{Name: "gofish", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewGoFishCuiController(usecase.NewGoFishInteractor(
				domain.NewDefaultGoFish(), new(presenter.GoFishCuiPresenter))),
			CuiHelpSpec{
				TitleKey:    "gofish.helpTitle",
				CommandKeys: []string{"gofish.helpAsk"},
				SettingKeys: []string{"gofish.helpSetDifficulty"},
			})
	}},
	{Name: "pinochle", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPinochleCuiController(usecase.NewPinochleInteractor(
				domain.NewDefaultPinochle(), new(presenter.PinochleCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "pinochle.helpTitle",
				CommandKeys: []string{
					"pinochle.helpBid",
					"pinochle.helpPass",
					"pinochle.helpTrump",
					"pinochle.helpMeld",
					"pinochle.helpPlay",
					"pinochle.helpNext",
					"pinochle.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"pinochle.helpSetDifficulty", "pinochle.helpSetLimit"},
			})
	}},
	{Name: "golf", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewGolfCuiController(usecase.NewGolfInteractor(
				domain.NewDefaultGolf(), new(presenter.GolfCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "golf.helpTitle",
				CommandKeys:       []string{"golf.helpDraw", "golf.helpRemove", "golf.helpGiveUp", "golf.helpHint"},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "pigtail", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPigsTailCuiController(usecase.NewPigsTailInteractor(
				domain.NewDefaultPigsTail(), new(presenter.PigsTailCuiPresenter))),
			CuiHelpSpec{
				TitleKey:    "pigtail.helpTitle",
				CommandKeys: []string{"pigtail.helpAction"},
			})
	}},
	{Name: "sevencardstud", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSevenCardStudCuiController(usecase.NewSevenCardStudInteractor(
				domain.NewDefaultSevenCardStud(), new(presenter.SevenCardStudCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "sevencardstud.helpTitle",
				CommandKeys: append([]string{
					"sevencardstud.helpFold",
					"sevencardstud.helpCheck",
					"sevencardstud.helpCall",
					"sevencardstud.helpBet",
					"sevencardstud.helpRaise",
					"sevencardstud.helpAllIn",
				}, tournamentRebuyAddOnKeys...),
				SettingKeys: append([]string{
					"sevencardstud.helpBettingLimit", "sevencardstud.helpTournament",
				}, studAnteKeys...),
			})
	}},
	{Name: "clocksolitaire", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewClockSolitaireCuiController(usecase.NewClockSolitaireInteractor(
				domain.NewDefaultClockSolitaire(), new(presenter.ClockSolitaireCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "clocksolitaire.helpTitle",
				CommandKeys:       []string{"clocksolitaire.helpStep", "clocksolitaire.helpAutoPlay"},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "durak", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewDurakCuiController(usecase.NewDurakInteractor(
				domain.NewDefaultDurak(), new(presenter.DurakCuiPresenter))),
			CuiHelpSpec{Body: []string{
				i18n.T("durak.helpTitle"),
				"",
				i18n.T("gameCommands"),
				"  a <idx>                  attack with card",
				"  d <atkIdx> <handIdx>     defend attack card",
				"  p                        pass (stop attacking)",
				"  t                        take cards (give up defense)",
				"  sort <0|1>               sort hand (0=suit, 1=value)",
				"  sd <0-2>                 set CPU difficulty",
				"  l                        action log",
				"",
				i18n.T("commonCommands"),
			}})
	}},
	{Name: "fortythieves", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewFortyThievesCuiController(usecase.NewFortyThievesInteractor(
				domain.NewDefaultFortyThieves(), new(presenter.FortyThievesCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "fortythieves.helpTitle",
				CommandKeys: []string{
					"fortythieves.helpDraw",
					"fortythieves.helpMove",
					"fortythieves.helpMoveWF",
					"fortythieves.helpMoveTF",
					"fortythieves.helpMoveTT",
					"fortythieves.helpGiveUp",
					"fortythieves.helpHint",
					"fortythieves.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "paigow", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPaiGowCuiController(usecase.NewPaiGowInteractor(
				domain.NewDefaultPaiGow(), new(presenter.PaiGowCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "paigow.helpTitle",
				CommandKeys:       []string{"paigow.helpBet", "paigow.helpSet"},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "twotenjack", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewTwoTenJackCuiController(usecase.NewTwoTenJackInteractor(
				domain.NewDefaultTwoTenJack(), new(presenter.TwoTenJackCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "twotenjack.helpTitle",
				CommandKeys: []string{
					"twotenjack.helpDeclare",
					"twotenjack.helpPlay",
					"twotenjack.helpNext",
					"twotenjack.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"twotenjack.helpSetDifficulty", "twotenjack.helpSetLimit"},
			})
	}},
	{Name: "caribbeanstud", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCaribbeanStudCuiController(usecase.NewCaribbeanStudInteractor(
				domain.NewDefaultCaribbeanStud(), new(presenter.CaribbeanStudCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "caribbeanstud.helpTitle",
				CommandKeys:       []string{"caribbeanstud.helpBet", "caribbeanstud.helpPlay", "caribbeanstud.helpFold"},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "texasholdembonus", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewTexasHoldemBonusCuiController(usecase.NewTexasHoldemBonusInteractor(
				domain.NewDefaultTexasHoldemBonus(), new(presenter.TexasHoldemBonusCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "texasholdembonus.helpTitle",
				CommandKeys: []string{
					"texasholdembonus.helpBet",
					"texasholdembonus.helpPlay",
					"texasholdembonus.helpFold",
					"texasholdembonus.helpCheck",
					"texasholdembonus.helpRaise",
				},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "war", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewWarCuiController(usecase.NewWarInteractor(
				domain.NewDefaultWar(), new(presenter.WarCuiPresenter))),
			CuiHelpSpec{
				TitleKey:    "war.helpTitle",
				CommandKeys: []string{"war.helpStep", "war.helpAutoPlay"},
				SettingKeys: []string{"war.helpSetMax"},
			})
	}},
	{Name: "canfield", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCanfieldCuiController(usecase.NewCanfieldInteractor(
				domain.NewDefaultCanfield(), new(presenter.CanfieldCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "canfield.helpTitle",
				CommandKeys: []string{
					"canfield.helpDraw",
					"canfield.helpMove",
					"canfield.helpMoveWF",
					"canfield.helpMoveRT",
					"canfield.helpMoveRF",
					"canfield.helpMoveTF",
					"canfield.helpMoveTT",
					"canfield.helpGiveUp",
					"canfield.helpHint",
					"canfield.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "fiftyone", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewFiftyOneCuiController(usecase.NewFiftyOneInteractor(
				domain.NewDefaultFiftyOne(), new(presenter.FiftyOneCuiPresenter))),
			CuiHelpSpec{
				TitleKey:    "fiftyone.helpTitle",
				CommandKeys: []string{"fiftyone.helpPlay", "fiftyone.helpAll", "fiftyone.helpStop"},
				SettingKeys: []string{"fiftyone.helpSetDifficulty"},
			})
	}},
	{Name: "yukon", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewYukonCuiController(usecase.NewYukonInteractor(
				domain.NewDefaultYukon(), new(presenter.YukonCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "yukon.helpTitle",
				CommandKeys: []string{
					"yukon.helpMove",
					"yukon.helpMoveTF",
					"yukon.helpMoveTT",
					"yukon.helpGiveUp",
					"yukon.helpHint",
					"yukon.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "russiansolitaire", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewRussianSolitaireCuiController(usecase.NewRussianSolitaireInteractor(
				domain.NewDefaultRussianSolitaire(), new(presenter.RussianSolitaireCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "russiansolitaire.helpTitle",
				CommandKeys: []string{
					"russiansolitaire.helpMove",
					"russiansolitaire.helpMoveTF",
					"russiansolitaire.helpMoveTT",
					"russiansolitaire.helpGiveUp",
					"russiansolitaire.helpHint",
					"russiansolitaire.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "whist", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewWhistCuiController(usecase.NewWhistInteractor(
				domain.NewDefaultWhist(), new(presenter.WhistCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "whist.helpTitle",
				CommandKeys:       []string{"whist.helpPlay", "whist.helpNext", "whist.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"whist.helpSetDifficulty", "whist.helpSetLimit"},
			})
	}},
	{Name: "letitride", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewLetItRideCuiController(usecase.NewLetItRideInteractor(
				domain.NewDefaultLetItRide(), new(presenter.LetItRideCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "letitride.helpTitle",
				CommandKeys:       []string{"letitride.helpBet", "letitride.helpPull", "letitride.helpLetItRide"},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "pokersquares", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPokerSquaresCuiController(usecase.NewPokerSquaresInteractor(
				domain.NewDefaultPokerSquares(), new(presenter.PokerSquaresCuiPresenter))),
			CuiHelpSpec{Body: []string{
				"Poker Squares (ポーカー・スクエアズ)",
				"",
				i18n.T("gameCommands"),
				"  p <row> <col>            カードを配置 (0-4)",
				"  u                        アンドゥ",
				"  g                        ギブアップ",
				"  l                        action log",
				"",
				i18n.T("session"),
				i18n.T("resetEntry"),
				i18n.T("quitEntry"),
				i18n.T("helpEntry"),
			}})
	}},
	{Name: "pageone", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPageOneCuiController(usecase.NewPageOneInteractor(
				domain.NewDefaultPageOne(), new(presenter.PageOneCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "pageone.helpTitle",
				CommandKeys: []string{
					"pageone.helpPlay",
					"pageone.helpDraw",
					"pageone.helpDeclare",
					"pageone.helpSkip",
					"pageone.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"pageone.helpSetDifficulty", "pageone.helpSetLimit"},
			})
	}},
	{Name: "reddog", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewRedDogCuiController(usecase.NewRedDogInteractor(
				domain.NewDefaultRedDog(), new(presenter.RedDogCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "reddog.helpTitle",
				CommandKeys:       []string{"reddog.helpBet", "reddog.helpRaise", "reddog.helpStay"},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "badugi", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBadugiCuiController(usecase.NewBadugiInteractor(
				domain.NewDefaultBadugi(), new(presenter.BadugiCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "badugi.helpTitle",
				CommandKeys: []string{
					"badugi.helpBet",
					"badugi.helpCall",
					"badugi.helpRaise",
					"badugi.helpCheck",
					"badugi.helpFold",
					"badugi.helpAllIn",
					"badugi.helpExchange",
					"badugi.helpStand",
				},
				SettingKeys: []string{"badugi.helpBettingLimit", "badugi.helpCpuCount"},
			})
	}},
	{Name: "razz", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSevenCardStudCuiController(usecase.NewSevenCardStudInteractor(
				domain.NewDefaultRazz(), new(presenter.SevenCardStudCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "razz.helpTitle",
				CommandKeys: append([]string{
					"sevencardstud.helpFold",
					"sevencardstud.helpCheck",
					"sevencardstud.helpCall",
					"sevencardstud.helpBet",
					"sevencardstud.helpRaise",
					"sevencardstud.helpAllIn",
				}, tournamentRebuyAddOnKeys...),
				SettingKeys: append([]string{
					"sevencardstud.helpBettingLimit", "sevencardstud.helpTournament",
				}, studAnteKeys...),
			})
	}},
	{Name: "scorpion", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewScorpionCuiController(usecase.NewScorpionInteractor(
				domain.NewDefaultScorpion(), new(presenter.ScorpionCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "scorpion.helpTitle",
				CommandKeys: []string{
					"scorpion.helpMove",
					"scorpion.helpMoveTT",
					"scorpion.helpDeal",
					"scorpion.helpGiveUp",
					"scorpion.helpHint",
					"scorpion.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "accordion", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewAccordionCuiController(usecase.NewAccordionInteractor(
				domain.NewDefaultAccordion(), new(presenter.AccordionCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "accordion.helpTitle",
				CommandKeys: []string{
					"accordion.helpMove",
					"accordion.helpGiveup",
					"accordion.helpHint",
					"accordion.helpLog",
					"accordion.helpUndo",
				},
			})
	}},
	{Name: "trash", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewTrashCuiController(usecase.NewTrashInteractor(
				domain.NewDefaultTrash(), new(presenter.TrashCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "trash.helpTitle",
				CommandKeys: []string{
					"trash.helpDraw",
					"trash.helpPlace",
					"trash.helpCpu",
					"trash.helpLog",
				},
			})
	}},
	{Name: "sevenbridge", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSevenBridgeCuiController(usecase.NewSevenBridgeInteractor(
				domain.NewDefaultSevenBridge(), new(presenter.SevenBridgeCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "sevenbridge.helpTitle",
				CommandKeys: []string{
					"sevenbridge.helpDrawStock",
					"sevenbridge.helpPon",
					"sevenbridge.helpChi",
					"sevenbridge.helpMeld",
					"sevenbridge.helpLayoff",
					"sevenbridge.helpDiscard",
					"sevenbridge.helpNextRound",
				},
				SettingKeys: []string{"sevenbridge.helpSetDifficulty", "sevenbridge.helpSetLimit"},
			})
	}},
	{Name: "president", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPresidentCuiController(usecase.NewPresidentInteractor(
				domain.NewDefaultPresident(), new(presenter.PresidentCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "president.helpTitle",
				CommandKeys: []string{
					"president.helpPlay",
					"president.helpPass",
					"president.helpLog",
				},
				SettingKeys: []string{
					"president.helpSetDifficulty",
					"president.helpSetRule",
				},
			})
	}},
	{Name: "cassino", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCassinoCuiController(usecase.NewCassinoInteractor(
				domain.NewDefaultCassino(), new(presenter.CassinoCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "cassino.helpTitle",
				CommandKeys: []string{
					"cassino.helpTake",
					"cassino.helpBuild",
					"cassino.helpTrail",
					"cassino.helpNext",
					"cassino.helpLog",
				},
				SettingKeys: []string{
					"cassino.helpSetDifficulty",
					"cassino.helpSetRule",
				},
			})
	}},
	{Name: "spanish21", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBlackJackCuiController(usecase.NewBlackJackInteractor(
				domain.NewSpanish21BlackJack(), new(presenter.BlackJackCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "spanish21.helpTitle",
				CommandKeys: []string{
					"blackjack.helpBet",
					"blackjack.helpHit",
					"blackjack.helpStand",
					"blackjack.helpDouble",
					"blackjack.helpSplit",
					"blackjack.helpInsurance",
					"blackjack.helpDeclineInsurance",
				},
				SettingKeys: []string{"blackjack.helpSetCpuCount"},
			})
	}},
	{Name: "calculation", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCalculationCuiController(usecase.NewCalculationInteractor(
				domain.NewDefaultCalculation(), new(presenter.CalculationCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "calculation.helpTitle",
				CommandKeys: []string{
					"calculation.helpStockToFoundation",
					"calculation.helpStockToWaste",
					"calculation.helpWasteToFoundation",
					"calculation.helpGiveUp",
					"calculation.helpHint",
					"calculation.helpAutoComplete",
					"calculation.helpUndo",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "spiteandmalice", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSpiteAndMaliceCuiController(usecase.NewSpiteAndMaliceInteractor(
				domain.NewDefaultSpiteAndMalice(), new(presenter.SpiteAndMaliceCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "spiteandmalice.helpTitle",
				CommandKeys: []string{
					"spiteandmalice.helpPlayHand",
					"spiteandmalice.helpPlayGoal",
					"spiteandmalice.helpPlaySide",
					"spiteandmalice.helpDiscard",
					"spiteandmalice.helpCpu",
					"spiteandmalice.helpHint",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "skat", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSkatCuiController(usecase.NewSkatInteractor(
				domain.NewDefaultSkat(), new(presenter.SkatCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "skat.helpTitle",
				CommandKeys: []string{
					"skat.helpBid",
					"skat.helpPickSkat",
					"skat.helpDiscard",
					"skat.helpGame",
					"skat.helpPlay",
					"skat.helpNext",
					"skat.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"skat.helpSetDifficulty", "skat.helpSetTarget"},
			})
	}},
	{Name: "shithead", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewShitheadCuiController(usecase.NewShitheadInteractor(
				domain.NewDefaultShithead(), new(presenter.ShitheadCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "shithead.helpTitle",
				CommandKeys: []string{
					"shithead.helpPlay",
					"shithead.helpPickup",
					"shithead.helpLog",
				},
				SettingKeys: []string{
					"shithead.helpSetDifficulty",
					"shithead.helpSetRule",
				},
			})
	}},
	{Name: "nertz", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewNertzCuiController(usecase.NewNertzInteractor(
				domain.NewDefaultNertz(), new(presenter.NertzCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "nertz.helpTitle",
				CommandKeys: []string{
					"nertz.helpDraw",
					"nertz.helpMoveNF",
					"nertz.helpMoveNT",
					"nertz.helpMoveWF",
					"nertz.helpMoveWT",
					"nertz.helpMoveTF",
					"nertz.helpMoveTT",
					"nertz.helpTick",
					"nertz.helpUndo",
					"nertz.helpNextRound",
					"nertz.helpHint",
				},
			})
	}},
	{Name: "slapjack", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSlapjackCuiController(usecase.NewSlapjackInteractor(
				domain.NewDefaultSlapjack(), new(presenter.SlapjackCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "slapjack.helpTitle",
				CommandKeys: []string{
					"slapjack.helpStep",
					"slapjack.helpSlap",
					"slapjack.helpTick",
				},
				SettingKeys: []string{"slapjack.helpSetDifficulty"},
			})
	}},
	{Name: "egyptianratscrew", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewEgyptianRatscrewCuiController(usecase.NewEgyptianRatscrewInteractor(
				domain.NewDefaultEgyptianRatscrew(), new(presenter.EgyptianRatscrewCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "egyptianratscrew.helpTitle",
				CommandKeys: []string{
					"egyptianratscrew.helpStep",
					"egyptianratscrew.helpSlap",
					"egyptianratscrew.helpTick",
				},
				SettingKeys: []string{"egyptianratscrew.helpSetDifficulty"},
			})
	}},
	{Name: "bakersdozen", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBakersDozenCuiController(usecase.NewBakersDozenInteractor(
				domain.NewDefaultBakersDozen(), new(presenter.BakersDozenCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "bakersdozen.helpTitle",
				CommandKeys: []string{
					"bakersdozen.helpMoveTT",
					"bakersdozen.helpMoveTF",
					"bakersdozen.helpGiveUp",
					"bakersdozen.helpHint",
					"bakersdozen.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "tonk", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewTonkCuiController(usecase.NewTonkInteractor(
				domain.NewDefaultTonk(), new(presenter.TonkCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "tonk.helpTitle",
				CommandKeys: []string{
					"tonk.helpDrawStock",
					"tonk.helpDrawDiscard",
					"tonk.helpDiscard",
					"tonk.helpKnock",
					"tonk.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"tonk.helpSetDifficulty", "tonk.helpSetLimit"},
			})
	}},
	{Name: "casinowar", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCasinoWarCuiController(usecase.NewCasinoWarInteractor(
				domain.NewDefaultCasinoWar(), new(presenter.CasinoWarCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "casinowar.helpTitle",
				CommandKeys:       []string{"casinowar.helpBet", "casinowar.helpSurrender", "casinowar.helpWar"},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "pitch", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPitchCuiController(usecase.NewPitchInteractor(
				domain.NewDefaultPitch(), new(presenter.PitchCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "pitch.helpTitle",
				CommandKeys:       []string{"pitch.helpBid", "pitch.helpPlay", "pitch.helpNext", "pitch.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"pitch.helpSetDifficulty", "pitch.helpSetLimit"},
			})
	}},
	{Name: "dragontiger", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewDragonTigerCuiController(usecase.NewDragonTigerInteractor(
				domain.NewDefaultDragonTiger(), new(presenter.DragonTigerCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "dragontiger.helpTitle",
				CommandKeys:       []string{"dragontiger.helpBet", "dragontiger.helpClear"},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "blackjackswitch", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBlackJackSwitchCuiController(usecase.NewBlackJackSwitchInteractor(
				domain.NewDefaultBlackJackSwitch(), new(presenter.BlackJackSwitchCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "blackjackswitch.helpTitle",
				CommandKeys: []string{
					"blackjackswitch.helpBet",
					"blackjackswitch.helpSwitch",
					"blackjackswitch.helpKeep",
					"blackjackswitch.helpHit",
					"blackjackswitch.helpStand",
					"blackjackswitch.helpDoubleDown",
				},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "montecarlo", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewMonteCarloCuiController(usecase.NewMonteCarloInteractor(
				domain.NewDefaultMonteCarlo(), new(presenter.MonteCarloCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "montecarlo.helpTitle",
				CommandKeys: []string{
					"montecarlo.helpRemove",
					"montecarlo.helpDeal",
					"montecarlo.helpUndo",
					"montecarlo.helpHint",
					"montecarlo.helpGiveup",
					"montecarlo.helpLog",
				},
			})
	}},
	{Name: "contractrummy", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewContractRummyCuiController(usecase.NewContractRummyInteractor(
				domain.NewDefaultContractRummy(), new(presenter.ContractRummyCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "contractrummy.helpTitle",
				CommandKeys: []string{
					"contractrummy.helpDrawStock",
					"contractrummy.helpDrawDiscard",
					"contractrummy.helpMeldContract",
					"contractrummy.helpMeldExtra",
					"contractrummy.helpLayoff",
					"contractrummy.helpDiscard",
					"contractrummy.helpNextRound",
				},
				SettingKeys: []string{"contractrummy.helpSetDifficulty", "contractrummy.helpSetPenalty"},
			})
	}},
	{Name: "ultimatetexasholdem", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewUltimateTexasHoldemCuiController(usecase.NewUltimateTexasHoldemInteractor(
				domain.NewDefaultUltimateTexasHoldem(), new(presenter.UltimateTexasHoldemCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "ultimatetexasholdem.helpTitle",
				CommandKeys: []string{
					"ultimatetexasholdem.helpBet",
					"ultimatetexasholdem.helpPlay",
					"ultimatetexasholdem.helpCheck",
					"ultimatetexasholdem.helpFold",
				},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "crescent", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCrescentCuiController(usecase.NewCrescentInteractor(
				domain.NewDefaultCrescent(), new(presenter.CrescentCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "crescent.helpTitle",
				CommandKeys: []string{
					"crescent.helpMoveTT",
					"crescent.helpMoveTF",
					"crescent.helpRedeal",
					"crescent.helpGiveUp",
					"crescent.helpHint",
					"crescent.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "mississippistud", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewMississippiStudCuiController(usecase.NewMississippiStudInteractor(
				domain.NewDefaultMississippiStud(), new(presenter.MississippiStudCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "mississippistud.helpTitle",
				CommandKeys: []string{
					"mississippistud.helpBet",
					"mississippistud.helpPlay",
					"mississippistud.helpFold",
				},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "belote", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBeloteCuiController(usecase.NewBeloteInteractor(
				domain.NewDefaultBelote(), new(presenter.BeloteCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "belote.helpTitle",
				CommandKeys: []string{
					"belote.helpOrderUp",
					"belote.helpPass",
					"belote.helpCall",
					"belote.helpPlay",
					"belote.helpNext",
					"belote.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"belote.helpSetDifficulty", "belote.helpSetTarget"},
			})
	}},
	{Name: "spiderette", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSpideretteCuiController(usecase.NewSpideretteInteractor(
				domain.NewDefaultSpiderette(), new(presenter.SpideretteCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "spiderette.helpTitle",
				CommandKeys: []string{
					"spiderette.helpDeal",
					"spiderette.helpMove",
					"spiderette.helpGiveUp",
					"spiderette.helpHint",
					"spiderette.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "mighty", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewMightyCuiController(usecase.NewMightyInteractor(
				domain.NewDefaultMighty(), new(presenter.MightyCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "mighty.helpTitle",
				CommandKeys: []string{
					"mighty.helpBid",
					"mighty.helpTrump",
					"mighty.helpExchange",
					"mighty.helpPlay",
					"mighty.helpJokerLead",
					"mighty.helpNext",
					"mighty.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"mighty.helpSetDifficulty", "mighty.helpSetLimit", "mighty.helpSetMinBid", "mighty.helpSetNoTrumpExtra"},
			})
	}},
	{Name: "oasispoker", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewOasisPokerCuiController(usecase.NewOasisPokerInteractor(
				domain.NewDefaultOasisPoker(), new(presenter.OasisPokerCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "oasispoker.helpTitle",
				CommandKeys: []string{
					"oasispoker.helpBet",
					"oasispoker.helpExchange",
					"oasispoker.helpStand",
					"oasispoker.helpPlay",
					"oasispoker.helpFold",
				},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "beleagueredcastle", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBeleagueredCastleCuiController(usecase.NewBeleagueredCastleInteractor(
				domain.NewDefaultBeleagueredCastle(), new(presenter.BeleagueredCastleCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "beleagueredcastle.helpTitle",
				CommandKeys: []string{
					"beleagueredcastle.helpMoveTT",
					"beleagueredcastle.helpMoveTF",
					"beleagueredcastle.helpGiveUp",
					"beleagueredcastle.helpHint",
					"beleagueredcastle.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "piquet", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPiquetCuiController(usecase.NewPiquetInteractor(
				domain.NewDefaultPiquet(), new(presenter.PiquetCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "piquet.helpTitle",
				CommandKeys: []string{
					"piquet.helpExchange",
					"piquet.helpExchangeY",
					"piquet.helpDeclare",
					"piquet.helpPlay",
					"piquet.helpNextDeal",
					"piquet.helpHint",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
}

// GameRegistry returns a copy of the game registry for external use.
func GameRegistry() []GameRegistryEntry {
	cp := make([]GameRegistryEntry, len(gameRegistry))
	copy(cp, gameRegistry)
	return cp
}

// GameNames returns the canonical ordered list of available game names,
// derived from gameRegistry. Returns a fresh copy on each call.
func GameNames() []string {
	names := make([]string, len(gameRegistry))
	for i, e := range gameRegistry {
		names[i] = e.Name
	}
	return names
}

// GameDescriptions returns a map of game names to their display descriptions.
// Descriptions are sourced from the games package (issue #1459 SSoT); this
// wrapper exists so CLI call sites don't have to import the games package
// directly.
func GameDescriptions() map[string]string {
	return games.Descriptions()
}

// GameAliases maps short alias names to their canonical game names.
// Aliases are not shown in help or game lists.
var GameAliases = map[string]string{
	"7stud":  "sevencardstud",
	"7cs":    "sevencardstud",
	"clock":  "clocksolitaire",
	"crazy8": "crazyeights",
	"indian": "indianpoker",
	"video":  "videopoker",
	"deuces": "deuceswild",
	"joker":  "jokerpoker",
	"short":  "shortdeck",
	"6plus":  "shortdeck",
	"gin":    "ginrummy",
	"3card":  "threecard",
	"csp":    "caribbeanstud",
	"stud":   "caribbeanstud",
	"oasis":  "oasispoker",
	"oasp":   "oasispoker",
	"thb":    "texasholdembonus",
	"thbp":   "texasholdembonus",
	"uth":    "ultimatetexasholdem",
	"uthe":   "ultimatetexasholdem",
	"40t":    "fortythieves",
	"pgp":    "paigow",
	"lir":    "letitride",
	"ride":   "letitride",
	"ms":     "mississippistud",
	"mstud":  "mississippistud",
	"sp21":   "spanish21",
	"s21":    "spanish21",
}

// cuiGame is implemented by each *Cui struct to expose its controller and help lines.
type cuiGame interface {
	Controller() CuiExecer
	HelpLines() []string
}

// GameManager manages multiple game CUI controllers and enables dynamic switching.
type GameManager struct {
	games       map[string]CuiExecer
	helpLines   map[string][]string
	initialized map[string]bool
	currentGame string
	gameOrder   []string
}

// NewGameManager creates a GameManager starting with startGame (must be a valid game name).
// i18n.SetLang must be called before NewGameManager: each game's HelpLines() is evaluated
// once at construction time, so changing the language afterwards will not update cached help text.
func NewGameManager(startGame string) *GameManager {
	controllers, helpLines := buildGameEntries()
	if _, ok := controllers[startGame]; !ok {
		panic(fmt.Sprintf("NewGameManager: unknown start game %q", startGame))
	}
	return &GameManager{
		games:       controllers,
		helpLines:   helpLines,
		initialized: make(map[string]bool),
		currentGame: startGame,
		gameOrder:   GameNames(),
	}
}

// Exec processes a command. "switch <game>" and "games" are handled by the manager;
// all other commands are delegated to the current game's controller.
func (m *GameManager) Exec(cmd string) string {
	fields := strings.Fields(cmd)
	if len(fields) > 0 {
		switch fields[0] {
		case "switch":
			if len(fields) < 2 {
				return i18n.T("switchUsage")
			}
			return m.switchGame(fields[1])
		case "games":
			return m.listGames()
		case "help", "?":
			return strings.Join(m.HelpLines(), "\n")
		}
	}
	return m.games[m.currentGame].Exec(cmd)
}

// HelpLines returns the current game's help lines plus interactive-mode commands.
func (m *GameManager) HelpLines() []string {
	base := m.helpLines[m.currentGame]
	extra := []string{
		"",
		i18n.Tf("interactiveMode", "name", m.currentGame),
		i18n.T("switchCmd"),
		i18n.T("gamesCmd"),
	}
	lines := make([]string, len(base)+len(extra))
	copy(lines, base)
	copy(lines[len(base):], extra)
	return lines
}

// CurrentGame returns the name of the currently active game.
func (m *GameManager) CurrentGame() string {
	return m.currentGame
}

// InitCurrentGame initializes (resets) the current game if not yet done and returns the reset output.
// This should be called once at startup before entering the game loop.
func (m *GameManager) InitCurrentGame() string {
	return m.initGame(m.currentGame)
}

func (m *GameManager) initGame(name string) string {
	if !m.initialized[name] {
		m.initialized[name] = true
		return m.games[name].Exec("r")
	}
	return ""
}

func (m *GameManager) switchGame(name string) string {
	name = strings.ToLower(name)
	if canonical, ok := GameAliases[name]; ok {
		name = canonical
	}
	if _, ok := m.games[name]; !ok {
		msg := i18n.Tf("unknownGame", "name", name)
		if suggestion := cuiutil.SuggestCommand(name, m.suggestionCandidates(), 2); suggestion != "" {
			msg += "\n  " + i18n.Tf("didYouMean", "name", suggestion)
		}
		return i18n.MarkError(msg)
	}
	if name == m.currentGame {
		return i18n.Tf("alreadyPlaying", "name", name)
	}
	m.currentGame = name
	initMsg := m.initGame(name)
	msg := i18n.Tf("switchedTo", "name", name)
	if initMsg != "" {
		return msg + "\n" + initMsg
	}
	return msg
}

// CompletionCandidates returns the manager-level commands valid as a first
// token: `switch` and `games`. Bare game names are intentionally NOT
// returned here because they aren't standalone commands — the manager's Exec
// would forward `blackjack` to the active controller, which would reject it.
// Game names are reachable via tab-completion as the second token of
// `switch` (see ArgumentCandidates). Issue #1608.
func (m *GameManager) CompletionCandidates() []string {
	return []string{"switch", "games"}
}

// ArgumentCandidates returns valid completions for the token after cmd. For
// `switch`, that's the canonical game name + alias set, so `switch bla<Tab>`
// expands to `blackjack`. Returns nil for any other command since the
// manager has no other argful commands at this layer (`help`, `?`, `q`,
// `r`, `games` are all nullary). Issue #1608.
func (m *GameManager) ArgumentCandidates(cmd string) []string {
	if cmd == "switch" {
		return m.suggestionCandidates()
	}
	return nil
}

// suggestionCandidates returns the deduplicated set of canonical game names
// and aliases for "did you mean" suggestions on `switch <typo>`. Mirrors the
// helper in cmd/trumpcards/main.go added in #1555 so a typo of an alias (e.g.
// "gni" for "gin") recovers the alias in interactive mode the same way it does
// at the top-level CLI. The local `add` closure matches the style used by
// helpSuggestionCandidates / suggestionCandidates(commands) in main.go so
// future readers see one dedup pattern instead of two. See issues #1602, #1625.
func (m *GameManager) suggestionCandidates() []string {
	capacity := len(m.gameOrder) + len(GameAliases)
	seen := make(map[string]struct{}, capacity)
	out := make([]string, 0, capacity)
	add := func(name string) {
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	for _, n := range m.gameOrder {
		add(n)
	}
	for alias := range GameAliases {
		add(alias)
	}
	return out
}

func (m *GameManager) listGames() string {
	var sb strings.Builder
	sb.WriteString(i18n.T("availableGames") + "\n")
	for _, name := range m.gameOrder {
		if name == m.currentGame {
			fmt.Fprintf(&sb, "  * %s %s\n", name, i18n.T("currentGame"))
		} else {
			fmt.Fprintf(&sb, "    %s\n", name)
		}
	}
	sb.WriteString(i18n.T("useSwitchCmd"))
	return sb.String()
}

func buildGameEntries() (map[string]CuiExecer, map[string][]string) {
	controllers := make(map[string]CuiExecer, len(gameRegistry))
	helpLines := make(map[string][]string, len(gameRegistry))
	for _, entry := range gameRegistry {
		g := entry.NewCui()
		controllers[entry.Name] = g.Controller()
		helpLines[entry.Name] = g.HelpLines()
	}
	return controllers, helpLines
}
