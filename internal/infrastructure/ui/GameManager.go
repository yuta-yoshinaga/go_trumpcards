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
	// by holdem, omaha, shortdeck, pineapple, crazypineapple, irishpoker.
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
	{Name: "bigtwo", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBigTwoCuiController(usecase.NewBigTwoInteractor(
				domain.NewDefaultBigTwo(), new(presenter.BigTwoCuiPresenter))),
			CuiHelpSpec{
				TitleKey:    "bigtwo.helpTitle",
				CommandKeys: []string{"bigtwo.helpPlay"},
				SettingKeys: []string{"bigtwo.helpSetDifficulty"},
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
	{Name: "bigo", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewOmahaCuiController(usecase.NewOmahaInteractor(
				domain.NewDefaultBigO(), new(presenter.OmahaCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "omaha.helpTitleBigO",
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
	{Name: "bigohilo", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewOmahaCuiController(usecase.NewOmahaInteractor(
				domain.NewDefaultBigOHiLo(), new(presenter.OmahaCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "omaha.helpTitleBigOHiLo",
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
	{Name: "irishpoker", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPineappleCuiController(usecase.NewPineappleInteractor(
				domain.NewDefaultIrishPoker(), new(presenter.PineappleCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "irishpoker.helpTitle",
				CommandKeys: append([]string{
					"irishpoker.helpFold",
					"irishpoker.helpCheck",
					"irishpoker.helpCall",
					"irishpoker.helpBet",
					"irishpoker.helpRaise",
					"irishpoker.helpAllIn",
				}, pineappleRebuyAddOnKeys...),
				SettingKeys: append([]string{
					"irishpoker.helpBettingLimit", "irishpoker.helpTournament",
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
	{Name: "seahaventowers", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSeahavenTowersCuiController(usecase.NewSeahavenTowersInteractor(
				domain.NewDefaultSeahavenTowers(), new(presenter.SeahavenTowersCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "seahaventowers.helpTitle",
				CommandKeys: []string{
					"seahaventowers.helpMove",
					"seahaventowers.helpMoveTF",
					"seahaventowers.helpMoveTT",
					"seahaventowers.helpMoveTC",
					"seahaventowers.helpMoveCT",
					"seahaventowers.helpMoveCF",
					"seahaventowers.helpGiveUp",
					"seahaventowers.helpHint",
					"seahaventowers.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "cruel", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCruelCuiController(usecase.NewCruelInteractor(
				domain.NewDefaultCruel(), new(presenter.CruelCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "cruel.helpTitle",
				CommandKeys: []string{
					"cruel.helpMove",
					"cruel.helpMoveTF",
					"cruel.helpShift",
					"cruel.helpGiveUp",
					"cruel.helpHint",
					"cruel.helpAutoComplete",
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
	{Name: "indianrummy", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewIndianRummyCuiController(usecase.NewIndianRummyInteractor(
				domain.NewDefaultIndianRummy(), new(presenter.IndianRummyCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "indianrummy.helpTitle",
				CommandKeys: []string{
					"indianrummy.helpDrawStock",
					"indianrummy.helpDrawDiscard",
					"indianrummy.helpDiscard",
					"indianrummy.helpDeclare",
					"indianrummy.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"indianrummy.helpSetPlayers", "indianrummy.helpSetDifficulty", "indianrummy.helpSetRounds"},
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
				CommandKeys:       []string{"videopoker.helpBet", "videopoker.helpHold", "videopoker.helpHint"},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "deuceswild", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewVideoPokerCuiController(usecase.NewVideoPokerInteractor(
				domain.NewDeucesWildVideoPoker(), new(presenter.VideoPokerCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "deuceswild.helpTitle",
				CommandKeys:       []string{"videopoker.helpBet", "videopoker.helpHold", "videopoker.helpHint"},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "jokerpoker", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewVideoPokerCuiController(usecase.NewVideoPokerInteractor(
				domain.NewJokerPokerVideoPoker(), new(presenter.VideoPokerCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "jokerpoker.helpTitle",
				CommandKeys:       []string{"videopoker.helpBet", "videopoker.helpHold", "videopoker.helpHint"},
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
				CommandKeys:       []string{"cribbage.helpDiscard", "cribbage.helpCut", "cribbage.helpPeg", "cribbage.helpGo", "cribbage.helpHint", "cribbage.helpShowNext", "cribbage.helpNextRound"},
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
	{Name: "ninetynine", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewNinetyNineCuiController(usecase.NewNinetyNineInteractor(
				domain.NewDefaultNinetyNine(), new(presenter.NinetyNineCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "ninetynine.helpTitle",
				CommandKeys:       []string{"ninetynine.helpBid", "ninetynine.helpPlay", "ninetynine.helpNext", "ninetynine.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"ninetynine.helpSetDifficulty", "ninetynine.helpSetTarget"},
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
				CommandKeys:       []string{"clocksolitaire.helpStep", "clocksolitaire.helpAutoPlay", "clocksolitaire.helpUndo"},
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
	{Name: "catchten", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCatchTenCuiController(usecase.NewCatchTenInteractor(
				domain.NewDefaultCatchTen(), new(presenter.CatchTenCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "catchten.helpTitle",
				CommandKeys:       []string{"catchten.helpPlay", "catchten.helpNext", "catchten.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"catchten.helpSetDifficulty", "catchten.helpSetLimit"},
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
				"  h                        ヒント (現在のカードの最善配置)",
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
				CommandKeys:       []string{"reddog.helpBet", "reddog.helpRaise", "reddog.helpStay", "reddog.helpHint"},
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
	{Name: "deucetoseven", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewDeuceToSevenCuiController(usecase.NewDeuceToSevenInteractor(
				domain.NewDefaultDeuceToSeven(), new(presenter.DeuceToSevenCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "deucetoseven.helpTitle",
				CommandKeys: []string{
					"deucetoseven.helpBet",
					"deucetoseven.helpCall",
					"deucetoseven.helpRaise",
					"deucetoseven.helpCheck",
					"deucetoseven.helpFold",
					"deucetoseven.helpAllIn",
					"deucetoseven.helpExchange",
					"deucetoseven.helpStand",
					"deucetoseven.helpHint",
				},
				SettingKeys: []string{"deucetoseven.helpBettingLimit", "deucetoseven.helpCpuCount"},
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
					"scorpion.helpLegal",
					"scorpion.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "wasp", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewWaspCuiController(usecase.NewWaspInteractor(
				domain.NewDefaultWasp(), new(presenter.WaspCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "wasp.helpTitle",
				CommandKeys: []string{
					"wasp.helpMove",
					"wasp.helpMoveTT",
					"wasp.helpDeal",
					"wasp.helpGiveUp",
					"wasp.helpHint",
					"wasp.helpLegal",
					"wasp.helpAutoComplete",
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
					"trash.helpHint",
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
					"sevenbridge.helpHint",
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
					"president.helpHint",
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
					"cassino.helpHint",
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
	{Name: "sirtommy", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSirTommyCuiController(usecase.NewSirTommyInteractor(
				domain.NewDefaultSirTommy(), new(presenter.SirTommyCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "sirtommy.helpTitle",
				CommandKeys: []string{
					"sirtommy.helpStockToFoundation",
					"sirtommy.helpStockToWaste",
					"sirtommy.helpWasteToFoundation",
					"sirtommy.helpGiveUp",
					"sirtommy.helpHint",
					"sirtommy.helpAutoComplete",
					"sirtommy.helpUndo",
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
	{Name: "streetsandalleys", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewStreetsAndAlleysCuiController(usecase.NewStreetsAndAlleysInteractor(
				domain.NewDefaultStreetsAndAlleys(), new(presenter.StreetsAndAlleysCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "streetsandalleys.helpTitle",
				CommandKeys: []string{
					"streetsandalleys.helpMoveTT",
					"streetsandalleys.helpMoveTF",
					"streetsandalleys.helpGiveUp",
					"streetsandalleys.helpHint",
					"streetsandalleys.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "kingalbert", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewKingAlbertCuiController(usecase.NewKingAlbertInteractor(
				domain.NewDefaultKingAlbert(), new(presenter.KingAlbertCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "kingalbert.helpTitle",
				CommandKeys: []string{
					"kingalbert.helpMoveTT",
					"kingalbert.helpMoveTF",
					"kingalbert.helpMoveRT",
					"kingalbert.helpMoveRF",
					"kingalbert.helpGiveUp",
					"kingalbert.helpHint",
					"kingalbert.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "flowergarden", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewFlowerGardenCuiController(usecase.NewFlowerGardenInteractor(
				domain.NewDefaultFlowerGarden(), new(presenter.FlowerGardenCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "flowergarden.helpTitle",
				CommandKeys: []string{
					"flowergarden.helpMoveTT",
					"flowergarden.helpMoveTF",
					"flowergarden.helpMoveRT",
					"flowergarden.helpMoveRF",
					"flowergarden.helpGiveUp",
					"flowergarden.helpHint",
					"flowergarden.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "fortyandeight", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewFortyAndEightCuiController(usecase.NewFortyAndEightInteractor(
				domain.NewDefaultFortyAndEight(), new(presenter.FortyAndEightCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "fortyandeight.helpTitle",
				CommandKeys: []string{
					"fortyandeight.helpDraw",
					"fortyandeight.helpRedeal",
					"fortyandeight.helpMove",
					"fortyandeight.helpMoveWF",
					"fortyandeight.helpMoveTF",
					"fortyandeight.helpMoveTT",
					"fortyandeight.helpGiveUp",
					"fortyandeight.helpHint",
					"fortyandeight.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "agnes", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewAgnesCuiController(usecase.NewAgnesInteractor(
				domain.NewDefaultAgnes(), new(presenter.AgnesCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "agnes.helpTitle",
				CommandKeys: []string{
					"agnes.helpDeal",
					"agnes.helpMoveTT",
					"agnes.helpMoveTF",
					"agnes.helpGiveUp",
					"agnes.helpHint",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "sultan", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSultanCuiController(usecase.NewSultanInteractor(
				domain.NewDefaultSultan(), new(presenter.SultanCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "sultan.helpTitle",
				CommandKeys: []string{
					"sultan.helpDraw",
					"sultan.helpRedeal",
					"sultan.helpMoveDF",
					"sultan.helpMoveWF",
					"sultan.helpGiveUp",
					"sultan.helpHint",
					"sultan.helpAutoComplete",
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
	{Name: "casinoholdem", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCasinoHoldemCuiController(usecase.NewCasinoHoldemInteractor(
				domain.NewDefaultCasinoHoldem(), new(presenter.CasinoHoldemCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "casinoholdem.helpTitle",
				CommandKeys: []string{
					"casinoholdem.helpBet",
					"casinoholdem.helpCall",
					"casinoholdem.helpFold",
				},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "callbreak", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCallBreakCuiController(usecase.NewCallBreakInteractor(
				domain.NewDefaultCallBreak(), new(presenter.CallBreakCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "callbreak.helpTitle",
				CommandKeys:       []string{"callbreak.helpBid", "callbreak.helpPlay", "callbreak.helpNext", "callbreak.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"callbreak.helpSetDifficulty", "callbreak.helpSetRounds"},
			})
	}},
	{Name: "tarneeb", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewTarneebCuiController(usecase.NewTarneebInteractor(
				domain.NewDefaultTarneeb(), new(presenter.TarneebCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "tarneeb.helpTitle",
				CommandKeys: []string{
					"tarneeb.helpBid",
					"tarneeb.helpTrump",
					"tarneeb.helpPlay",
					"tarneeb.helpNext",
					"tarneeb.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"tarneeb.helpSetDifficulty", "tarneeb.helpSetLimit", "tarneeb.helpSetMinBid"},
			})
	}},
	{Name: "highcardflush", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewHighCardFlushCuiController(usecase.NewHighCardFlushInteractor(
				domain.NewDefaultHighCardFlush(), new(presenter.HighCardFlushCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "highcardflush.helpTitle",
				CommandKeys:       []string{"highcardflush.helpBet", "highcardflush.helpRaise", "highcardflush.helpFold"},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "briscola", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBriscolaCuiController(usecase.NewBriscolaInteractor(
				domain.NewDefaultBriscola(), new(presenter.BriscolaCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "briscola.helpTitle",
				CommandKeys:       []string{"briscola.helpPlay", "briscola.helpNext"},
				ExtraCommandLines: []string{"  l                    action log"},
			})
	}},
	{Name: "gaps", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewGapsCuiController(usecase.NewGapsInteractor(
				domain.NewDefaultGaps(), new(presenter.GapsCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "gaps.helpTitle",
				CommandKeys: []string{
					"gaps.helpMove",
					"gaps.helpRedeal",
					"gaps.helpUndo",
					"gaps.helpGiveUp",
					"gaps.helpHint",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "fourcardpoker", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewFourCardPokerCuiController(usecase.NewFourCardPokerInteractor(
				domain.NewDefaultFourCardPoker(), new(presenter.FourCardPokerCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "fourcardpoker.helpTitle",
				CommandKeys:       []string{"fourcardpoker.helpBet", "fourcardpoker.helpPlay", "fourcardpoker.helpFold"},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "rummy500", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewRummy500CuiController(usecase.NewRummy500Interactor(
				domain.NewDefaultRummy500(), new(presenter.Rummy500CuiPresenter))),
			CuiHelpSpec{
				TitleKey: "rummy500.helpTitle",
				CommandKeys: []string{
					"rummy500.helpDrawStock",
					"rummy500.helpDrawDiscard",
					"rummy500.helpMeld",
					"rummy500.helpLayoff",
					"rummy500.helpDiscard",
					"rummy500.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"rummy500.helpSetDifficulty", "rummy500.helpSetLimit"},
			})
	}},
	{Name: "eightoff", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewEightOffCuiController(usecase.NewEightOffInteractor(
				domain.NewDefaultEightOff(), new(presenter.EightOffCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "eightoff.helpTitle",
				CommandKeys: []string{
					"eightoff.helpMove",
					"eightoff.helpMoveTF",
					"eightoff.helpMoveTT",
					"eightoff.helpMoveTC",
					"eightoff.helpMoveCT",
					"eightoff.helpMoveCF",
					"eightoff.helpGiveUp",
					"eightoff.helpHint",
					"eightoff.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "russianpoker", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewRussianPokerCuiController(usecase.NewRussianPokerInteractor(
				domain.NewDefaultRussianPoker(), new(presenter.RussianPokerCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "russianpoker.helpTitle",
				CommandKeys: []string{
					"russianpoker.helpBet",
					"russianpoker.helpExchange",
					"russianpoker.helpBuy6th",
					"russianpoker.helpSelect",
					"russianpoker.helpPlay",
					"russianpoker.helpFold",
					"russianpoker.helpForce",
					"russianpoker.helpDecline",
				},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "penguin", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPenguinCuiController(usecase.NewPenguinInteractor(
				domain.NewDefaultPenguin(), new(presenter.PenguinCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "penguin.helpTitle",
				CommandKeys: []string{
					"penguin.helpMove",
					"penguin.helpMoveTF",
					"penguin.helpMoveTT",
					"penguin.helpMoveTC",
					"penguin.helpMoveCT",
					"penguin.helpMoveCF",
					"penguin.helpGiveUp",
					"penguin.helpHint",
					"penguin.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "chinesepoker", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewChinesePokerCuiController(usecase.NewChinesePokerInteractor(
				domain.NewDefaultChinesePoker(), new(presenter.ChinesePokerCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "chinesepoker.helpTitle",
				CommandKeys:       []string{"chinesepoker.helpBet", "chinesepoker.helpSet"},
				ExtraCommandLines: []string{"  l                                            action log"},
			})
	}},
	{Name: "sixcardgolf", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSixCardGolfCuiController(usecase.NewSixCardGolfInteractor(
				domain.NewDefaultSixCardGolf(), new(presenter.SixCardGolfCuiPresenter))),
			CuiHelpSpec{
				TitleKey:    "sixcardgolf.helpTitle",
				CommandKeys: []string{"sixcardgolf.helpFlipInitial", "sixcardgolf.helpDrawStock", "sixcardgolf.helpDrawDiscard", "sixcardgolf.helpSwap", "sixcardgolf.helpDiscard", "sixcardgolf.helpFlip", "sixcardgolf.helpSkipFlip", "sixcardgolf.helpNextRound"},
				SettingKeys: []string{"sixcardgolf.helpSetDifficulty", "sixcardgolf.helpSetPlayers", "sixcardgolf.helpSetRounds"},
			})
	}},
	{Name: "doudizhu", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewDoudizhuCuiController(usecase.NewDoudizhuInteractor(
				domain.NewDefaultDoudizhu(), new(presenter.DoudizhuCuiPresenter))),
			CuiHelpSpec{
				TitleKey:    "doudizhu.helpTitle",
				CommandKeys: []string{"doudizhu.helpPlay", "doudizhu.helpBid"},
				SettingKeys: []string{"doudizhu.helpSetDifficulty"},
			})
	}},
	{Name: "truco", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewTrucoCuiController(usecase.NewTrucoInteractor(
				domain.NewDefaultTruco(), new(presenter.TrucoCuiPresenter))),
			CuiHelpSpec{
				TitleKey:    "truco.helpTitle",
				CommandKeys: []string{"truco.helpPlay", "truco.helpTruco", "truco.helpRespond", "truco.helpNext"},
			})
	}},
	{Name: "scopa", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewScopaCuiController(usecase.NewScopaInteractor(
				domain.NewDefaultScopa(), new(presenter.ScopaCuiPresenter))),
			CuiHelpSpec{
				TitleKey:    "scopa.helpTitle",
				CommandKeys: []string{"scopa.helpPlay", "scopa.helpNext"},
				SettingKeys: []string{"scopa.helpSetDifficulty"},
			})
	}},
	{Name: "acesup", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewAcesUpCuiController(usecase.NewAcesUpInteractor(
				domain.NewDefaultAcesUp(), new(presenter.AcesUpCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "acesup.helpTitle",
				CommandKeys:       []string{"acesup.helpDraw", "acesup.helpRemove", "acesup.helpMove", "acesup.helpGiveUp", "acesup.helpHint"},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "barbu", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBarbuCuiController(usecase.NewBarbuInteractor(
				domain.NewDefaultBarbu(), new(presenter.BarbuCuiPresenter))),
			CuiHelpSpec{
				TitleKey:    "barbu.helpTitle",
				CommandKeys: []string{"barbu.helpContract", "barbu.helpPlay", "barbu.helpNext"},
				SettingKeys: []string{"barbu.helpSetDifficulty"},
			})
	}},
	{Name: "macau", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewMacauCuiController(usecase.NewMacauInteractor(
				domain.NewDefaultMacau(), new(presenter.MacauCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "macau.helpTitle",
				CommandKeys:       []string{"macau.helpPlay", "macau.helpDraw", "macau.helpSuit", "macau.helpDeclare", "macau.helpSkipDeclare", "macau.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"macau.helpSetDifficulty", "macau.helpSetLimit"},
			})
	}},
	{Name: "thirtyone", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewThirtyOneCuiController(usecase.NewThirtyOneInteractor(
				domain.NewDefaultThirtyOne(), new(presenter.ThirtyOneCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "thirtyone.helpTitle",
				CommandKeys: []string{
					"thirtyone.helpDrawStock",
					"thirtyone.helpDrawDiscard",
					"thirtyone.helpDiscard",
					"thirtyone.helpKnock",
					"thirtyone.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"thirtyone.helpSetDifficulty", "thirtyone.helpSetLives"},
			})
	}},
	{Name: "tienlen", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewTienLenCuiController(usecase.NewTienLenInteractor(
				domain.NewDefaultTienLen(), new(presenter.TienLenCuiPresenter))),
			CuiHelpSpec{
				TitleKey:    "tienlen.helpTitle",
				CommandKeys: []string{"tienlen.helpPlay"},
				SettingKeys: []string{"tienlen.helpSetDifficulty"},
			})
	}},
	{Name: "osmosis", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewOsmosisCuiController(usecase.NewOsmosisInteractor(
				domain.NewDefaultOsmosis(), new(presenter.OsmosisCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "osmosis.helpTitle",
				CommandKeys: []string{
					"osmosis.helpDraw",
					"osmosis.helpMoveWF",
					"osmosis.helpMoveRF",
					"osmosis.helpGiveUp",
					"osmosis.helpHint",
					"osmosis.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "fivehundred", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewFiveHundredCuiController(usecase.NewFiveHundredInteractor(
				domain.NewDefaultFiveHundred(), new(presenter.FiveHundredCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "fivehundred.helpTitle",
				CommandKeys: []string{
					"fivehundred.helpBid",
					"fivehundred.helpBidNoTrump",
					"fivehundred.helpMisere",
					"fivehundred.helpOpenMisere",
					"fivehundred.helpPass",
					"fivehundred.helpExchange",
					"fivehundred.helpPlay",
					"fivehundred.helpNext",
					"fivehundred.helpNextRound",
				},
				SettingKeys: []string{
					"fivehundred.helpSetDifficulty",
					"fivehundred.helpSetTarget",
				},
			})
	}},
	{Name: "schnapsen", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSchnapsenCuiController(usecase.NewSchnapsenInteractor(
				domain.NewDefaultSchnapsen(), new(presenter.SchnapsenCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "schnapsen.helpTitle",
				CommandKeys: []string{
					"schnapsen.helpPlay",
					"schnapsen.helpMarriage",
					"schnapsen.helpNext",
				},
				ExtraCommandLines: []string{"  l                    action log"},
			})
	}},
	{Name: "burraco", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBurracoCuiController(usecase.NewBurracoInteractor(
				domain.NewDefaultBurraco(), new(presenter.BurracoCuiPresenter))),
			CuiHelpSpec{Body: []string{
				"Burraco (ブラーコ) Help",
				"",
				"Game Commands:",
				"  ds                   draw from stock",
				"  dd <idx,idx>         pick up discard pile (natural pair indices)",
				"  m <idx,idx;idx,idx>  meld (semicolon-separated groups)",
				"  sm                   skip meld phase",
				"  d <idx>              discard a card",
				"  go                   go out (requires the pozzetto + a burraco)",
				"  nr                   next round",
				"  h                    hint (recommended action)",
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
	{Name: "yaniv", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewYanivCuiController(usecase.NewYanivInteractor(
				domain.NewDefaultYaniv(), new(presenter.YanivCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "yaniv.helpTitle",
				CommandKeys: []string{
					"yaniv.helpDiscard",
					"yaniv.helpYaniv",
					"yaniv.helpDrawStock",
					"yaniv.helpDrawPickup",
					"yaniv.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"yaniv.helpSetDifficulty", "yaniv.helpSetLimit"},
			})
	}},
	{Name: "gongzhu", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewGongZhuCuiController(usecase.NewGongZhuInteractor(
				domain.NewDefaultGongZhu(), new(presenter.GongZhuCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "gongzhu.helpTitle",
				CommandKeys: []string{
					"gongzhu.helpExpose",
					"gongzhu.helpPlay",
					"gongzhu.helpNext",
					"gongzhu.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"gongzhu.helpSetDifficulty", "gongzhu.helpSetLimit"},
			})
	}},
	{Name: "bristol", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBristolCuiController(usecase.NewBristolInteractor(
				domain.NewDefaultBristol(), new(presenter.BristolCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "bristol.helpTitle",
				CommandKeys: []string{
					"bristol.helpDraw",
					"bristol.helpMoveTT",
					"bristol.helpMoveTF",
					"bristol.helpMoveNT",
					"bristol.helpMoveNF",
					"bristol.helpGiveUp",
					"bristol.helpHint",
					"bristol.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                    action log"},
			})
	}},
	{Name: "bidwhist", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBidWhistCuiController(usecase.NewBidWhistInteractor(
				domain.NewDefaultBidWhist(), new(presenter.BidWhistCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "bidwhist.helpTitle",
				CommandKeys: []string{
					"bidwhist.helpBid",
					"bidwhist.helpPass",
					"bidwhist.helpTrump",
					"bidwhist.helpExchange",
					"bidwhist.helpPlay",
					"bidwhist.helpNext",
					"bidwhist.helpNextRound",
				},
				SettingKeys: []string{
					"bidwhist.helpSetDifficulty",
					"bidwhist.helpSetTarget",
				},
			})
	}},
	{Name: "tressette", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewTressetteCuiController(usecase.NewTressetteInteractor(
				domain.NewDefaultTressette(), new(presenter.TressetteCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "tressette.helpTitle",
				CommandKeys: []string{
					"tressette.helpPlay",
					"tressette.helpNext",
					"tressette.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys: []string{
					"tressette.helpSetDifficulty",
					"tressette.helpSetTarget",
				},
			})
	}},
	{Name: "easthaven", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewEasthavenCuiController(usecase.NewEasthavenInteractor(
				domain.NewDefaultEasthaven(), new(presenter.EasthavenCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "easthaven.helpTitle",
				CommandKeys: []string{
					"easthaven.helpMove",
					"easthaven.helpMoveTF",
					"easthaven.helpMoveTT",
					"easthaven.helpDeal",
					"easthaven.helpGiveUp",
					"easthaven.helpHint",
					"easthaven.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "tichu", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewTichuCuiController(usecase.NewTichuInteractor(
				domain.NewDefaultTichu(), new(presenter.TichuCuiPresenter))),
			CuiHelpSpec{
				TitleKey:    "tichu.helpTitle",
				CommandKeys: []string{"tichu.helpPlay", "tichu.helpDeclare"},
				SettingKeys: []string{"tichu.helpSetDifficulty"},
			})
	}},
	{Name: "bakersgame", NewCui: func() cuiGame {
		// Baker's Game reuses the FreeCell interactor/controller; only the
		// domain (same-suit stacking) and presenter (i18n namespace) differ.
		return cuiEntry(
			controller.NewFreeCellCuiController(usecase.NewFreeCellInteractor(
				domain.NewDefaultBakersGame(), new(presenter.BakersGameCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "bakersgame.helpTitle",
				CommandKeys: []string{
					"bakersgame.helpMove",
					"bakersgame.helpMoveTF",
					"bakersgame.helpMoveTT",
					"bakersgame.helpMoveTC",
					"bakersgame.helpMoveCT",
					"bakersgame.helpMoveCF",
					"bakersgame.helpGiveUp",
					"bakersgame.helpHint",
					"bakersgame.helpAutoComplete",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "bourre", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBourreCuiController(usecase.NewBourreInteractor(
				domain.NewDefaultBourre(), new(presenter.BourreCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "bourre.helpTitle",
				CommandKeys: []string{
					"bourre.helpDecide",
					"bourre.helpDraw",
					"bourre.helpPlay",
					"bourre.helpNext",
					"bourre.helpSetDifficulty",
				},
				ExtraCommandLines: []string{"  l                        action log"},
			})
	}},
	{Name: "sheepshead", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSheepsheadCuiController(usecase.NewSheepsheadInteractor(
				domain.NewDefaultSheepshead(), new(presenter.SheepsheadCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "sheepshead.helpTitle",
				CommandKeys: []string{
					"sheepshead.helpPick",
					"sheepshead.helpPass",
					"sheepshead.helpBury",
					"sheepshead.helpCall",
					"sheepshead.helpPlay",
					"sheepshead.helpNext",
					"sheepshead.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys: []string{
					"sheepshead.helpSetDifficulty",
					"sheepshead.helpSetChips",
				},
			})
	}},
	{Name: "doppelkopf", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewDoppelkopfCuiController(usecase.NewDoppelkopfInteractor(
				domain.NewDefaultDoppelkopf(), new(presenter.DoppelkopfCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "doppelkopf.helpTitle",
				CommandKeys: []string{
					"doppelkopf.helpPlay",
					"doppelkopf.helpAnnounce",
					"doppelkopf.helpNext",
					"doppelkopf.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys: []string{
					"doppelkopf.helpSetDifficulty",
					"doppelkopf.helpSetChips",
				},
			})
	}},
	{Name: "mus", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewMusCuiController(usecase.NewMusInteractor(
				domain.NewDefaultMus(), new(presenter.MusCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "mus.helpTitle",
				CommandKeys: []string{
					"mus.helpMus",
					"mus.helpCut",
					"mus.helpDiscard",
					"mus.helpPaso",
					"mus.helpEnvido",
					"mus.helpOrdago",
					"mus.helpQuiero",
					"mus.helpNoQuiero",
					"mus.helpNext",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"mus.helpSetDifficulty"},
			})
	}},
	{Name: "tute", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewTuteCuiController(usecase.NewTuteInteractor(
				domain.NewDefaultTute(), new(presenter.TuteCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "tute.helpTitle",
				CommandKeys: []string{
					"tute.helpPlay",
					"tute.helpMarriage",
					"tute.helpTute",
					"tute.helpNext",
					"tute.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"tute.helpSetDifficulty"},
			})
	}},
	{Name: "sueca", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSuecaCuiController(usecase.NewSuecaInteractor(
				domain.NewDefaultSueca(), new(presenter.SuecaCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "sueca.helpTitle",
				CommandKeys: []string{
					"sueca.helpPlay",
					"sueca.helpNext",
					"sueca.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"sueca.helpSetDifficulty"},
			})
	}},
	{Name: "fortyfives", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewFortyFivesCuiController(usecase.NewFortyFivesInteractor(
				domain.NewDefaultFortyFives(), new(presenter.FortyFivesCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "fortyfives.helpTitle",
				CommandKeys: []string{
					"fortyfives.helpBid",
					"fortyfives.helpPass",
					"fortyfives.helpPlay",
					"fortyfives.helpNext",
					"fortyfives.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"fortyfives.helpSetDifficulty"},
			})
	}},
	{Name: "twentynine", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewTwentyNineCuiController(usecase.NewTwentyNineInteractor(
				domain.NewDefaultTwentyNine(), new(presenter.TwentyNineCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "twentynine.helpTitle",
				CommandKeys: []string{
					"twentynine.helpBid",
					"twentynine.helpPass",
					"twentynine.helpPlay",
					"twentynine.helpNext",
					"twentynine.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"twentynine.helpSetDifficulty"},
			})
	}},
	{Name: "klaverjas", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewKlaverjasCuiController(usecase.NewKlaverjasInteractor(
				domain.NewDefaultKlaverjas(), new(presenter.KlaverjasCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "klaverjas.helpTitle",
				CommandKeys: []string{
					"klaverjas.helpPlay",
					"klaverjas.helpNext",
					"klaverjas.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"klaverjas.helpSetDifficulty"},
			})
	}},
	{Name: "manille", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewManilleCuiController(usecase.NewManilleInteractor(
				domain.NewDefaultManille(), new(presenter.ManilleCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "manille.helpTitle",
				CommandKeys: []string{
					"manille.helpPlay",
					"manille.helpNext",
					"manille.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"manille.helpSetDifficulty"},
			})
	}},
	{Name: "marias", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewMariasCuiController(usecase.NewMariasInteractor(
				domain.NewDefaultMarias(), new(presenter.MariasCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "marias.helpTitle",
				CommandKeys: []string{
					"marias.helpPlay",
					"marias.helpNext",
					"marias.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"marias.helpSetDifficulty"},
			})
	}},
	{Name: "sedma", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSedmaCuiController(usecase.NewSedmaInteractor(
				domain.NewDefaultSedma(), new(presenter.SedmaCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "sedma.helpTitle",
				CommandKeys: []string{
					"sedma.helpPlay",
					"sedma.helpNext",
					"sedma.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"sedma.helpSetDifficulty"},
			})
	}},
	{Name: "solowhist", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSoloWhistCuiController(usecase.NewSoloWhistInteractor(
				domain.NewDefaultSoloWhist(), new(presenter.SoloWhistCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "solowhist.helpTitle",
				CommandKeys: []string{
					"solowhist.helpBid",
					"solowhist.helpPass",
					"solowhist.helpPlay",
					"solowhist.helpNext",
					"solowhist.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"solowhist.helpSetDifficulty"},
			})
	}},
	{Name: "knockoutwhist", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewKnockoutWhistCuiController(usecase.NewKnockoutWhistInteractor(
				domain.NewDefaultKnockoutWhist(), new(presenter.KnockoutWhistCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "knockoutwhist.helpTitle",
				CommandKeys: []string{
					"knockoutwhist.helpPlay",
					"knockoutwhist.helpNext",
					"knockoutwhist.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"knockoutwhist.helpSetDifficulty"},
			})
	}},
	{Name: "nap", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewNapCuiController(usecase.NewNapInteractor(
				domain.NewDefaultNap(), new(presenter.NapCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "nap.helpTitle",
				CommandKeys: []string{
					"nap.helpBid",
					"nap.helpPass",
					"nap.helpPlay",
					"nap.helpNext",
					"nap.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"nap.helpSetDifficulty"},
			})
	}},
	{Name: "preference", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPreferenceCuiController(usecase.NewPreferenceInteractor(
				domain.NewDefaultPreference(), new(presenter.PreferenceCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "preference.helpTitle",
				CommandKeys: []string{
					"preference.helpBid",
					"preference.helpPass",
					"preference.helpPlay",
					"preference.helpNext",
					"preference.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"preference.helpSetDifficulty"},
			})
	}},
	{Name: "spoilfive", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSpoilFiveCuiController(usecase.NewSpoilFiveInteractor(
				domain.NewDefaultSpoilFive(), new(presenter.SpoilFiveCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "spoilfive.helpTitle",
				CommandKeys: []string{
					"spoilfive.helpPlay",
					"spoilfive.helpNext",
					"spoilfive.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"spoilfive.helpSetDifficulty"},
			})
	}},
	{Name: "courtpiece", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCourtPieceCuiController(usecase.NewCourtPieceInteractor(
				domain.NewDefaultCourtPiece(), new(presenter.CourtPieceCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "courtpiece.helpTitle",
				CommandKeys: []string{
					"courtpiece.helpTrump",
					"courtpiece.helpPlay",
					"courtpiece.helpNext",
					"courtpiece.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"courtpiece.helpSetDifficulty"},
			})
	}},
	{Name: "bezique", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBeziqueCuiController(usecase.NewBeziqueInteractor(
				domain.NewDefaultBezique(), new(presenter.BeziqueCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "bezique.helpTitle",
				CommandKeys: []string{
					"bezique.helpPlay",
					"bezique.helpMeld",
					"bezique.helpSkip",
					"bezique.helpNext",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"bezique.helpSetDifficulty"},
			})
	}},
	{Name: "ecarte", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewEcarteCuiController(usecase.NewEcarteInteractor(
				domain.NewDefaultEcarte(), new(presenter.EcarteCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "ecarte.helpTitle",
				CommandKeys: []string{
					"ecarte.helpPropose",
					"ecarte.helpStand",
					"ecarte.helpDiscard",
					"ecarte.helpPlay",
					"ecarte.helpNext",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"ecarte.helpSetDifficulty"},
			})
	}},
	{Name: "threecardbrag", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewThreeCardBragCuiController(usecase.NewThreeCardBragInteractor(
				domain.NewDefaultThreeCardBrag(), new(presenter.ThreeCardBragCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "threecardbrag.helpTitle",
				CommandKeys: []string{
					"threecardbrag.helpSee",
					"threecardbrag.helpBet",
					"threecardbrag.helpRaise",
					"threecardbrag.helpFold",
					"threecardbrag.helpShow",
					"threecardbrag.helpNext",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"threecardbrag.helpSetDifficulty"},
			})
	}},
	{Name: "teenpatti", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewTeenPattiCuiController(usecase.NewTeenPattiInteractor(
				domain.NewDefaultTeenPatti(), new(presenter.TeenPattiCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "teenpatti.helpTitle",
				CommandKeys: []string{
					"teenpatti.helpSee",
					"teenpatti.helpBet",
					"teenpatti.helpRaise",
					"teenpatti.helpFold",
					"teenpatti.helpShow",
					"teenpatti.helpSideShow",
					"teenpatti.helpNext",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"teenpatti.helpSetDifficulty"},
			})
	}},
	{Name: "scopone", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewScoponeCuiController(usecase.NewScoponeInteractor(
				domain.NewDefaultScopone(), new(presenter.ScoponeCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "scopone.helpTitle",
				CommandKeys: []string{
					"scopone.helpPlay",
					"scopone.helpNext",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"scopone.helpSetDifficulty"},
			})
	}},
	{Name: "escoba", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewEscobaCuiController(usecase.NewEscobaInteractor(
				domain.NewDefaultEscoba(), new(presenter.EscobaCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "escoba.helpTitle",
				CommandKeys: []string{
					"escoba.helpPlay",
					"escoba.helpNext",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"escoba.helpSetDifficulty"},
			})
	}},
	{Name: "handandfoot", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewHandAndFootCuiController(usecase.NewHandAndFootInteractor(
				domain.NewDefaultHandAndFoot(), new(presenter.HandAndFootCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "handandfoot.helpTitle",
				CommandKeys: []string{
					"handandfoot.helpDrawStock",
					"handandfoot.helpDrawDiscard",
					"handandfoot.helpMeld",
					"handandfoot.helpSkipMeld",
					"handandfoot.helpDiscard",
					"handandfoot.helpGoOut",
					"handandfoot.helpNextRound",
					"handandfoot.helpHint",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys: []string{
					"handandfoot.helpSetDifficulty",
					"handandfoot.helpSetLimit",
				},
			})
	}},
	{Name: "conquian", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewConquianCuiController(usecase.NewConquianInteractor(
				domain.NewDefaultConquian(), new(presenter.ConquianCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "conquian.helpTitle",
				CommandKeys: []string{
					"conquian.helpDrawStock",
					"conquian.helpDrawDiscard",
					"conquian.helpMeld",
					"conquian.helpDiscard",
					"conquian.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys: []string{
					"conquian.helpSetDifficulty",
					"conquian.helpSetWins",
				},
			})
	}},
	{Name: "chinchon", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewChinchonCuiController(usecase.NewChinchonInteractor(
				domain.NewDefaultChinchon(), new(presenter.ChinchonCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "chinchon.helpTitle",
				CommandKeys: []string{
					"chinchon.helpDrawStock",
					"chinchon.helpDrawDiscard",
					"chinchon.helpDiscard",
					"chinchon.helpKnock",
					"chinchon.helpLayoff",
					"chinchon.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys: []string{
					"chinchon.helpSetDifficulty",
					"chinchon.helpSetPlayers",
				},
			})
	}},
	{Name: "kalooki", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewKalookiCuiController(usecase.NewKalookiInteractor(
				domain.NewDefaultKalooki(), new(presenter.KalookiCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "kalooki.helpTitle",
				CommandKeys: []string{
					"kalooki.helpDrawStock",
					"kalooki.helpDrawDiscard",
					"kalooki.helpMeld",
					"kalooki.helpLayoff",
					"kalooki.helpDiscard",
					"kalooki.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys: []string{
					"kalooki.helpSetDifficulty",
					"kalooki.helpSetPlayers",
					"kalooki.helpSetThreshold",
				},
			})
	}},
	{Name: "threethirteen", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewThreeThirteenCuiController(usecase.NewThreeThirteenInteractor(
				domain.NewDefaultThreeThirteen(), new(presenter.ThreeThirteenCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "threethirteen.helpTitle",
				CommandKeys: []string{
					"threethirteen.helpDrawStock",
					"threethirteen.helpDrawDiscard",
					"threethirteen.helpDiscard",
					"threethirteen.helpKnock",
					"threethirteen.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys: []string{
					"threethirteen.helpSetDifficulty",
					"threethirteen.helpSetPlayers",
				},
			})
	}},
	{Name: "mao", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewMaoCuiController(usecase.NewMaoInteractor(
				domain.NewDefaultMao(), new(presenter.MaoCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "mao.helpTitle",
				CommandKeys: []string{
					"mao.helpPlay",
					"mao.helpDraw",
					"mao.helpSuit",
					"mao.helpDeclare",
					"mao.helpDeclareWord",
					"mao.helpSkipDeclare",
					"mao.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys: []string{
					"mao.helpSetDifficulty",
					"mao.helpSetLimit",
				},
			})
	}},
	{Name: "spoons", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSpoonsCuiController(usecase.NewSpoonsInteractor(
				domain.NewDefaultSpoons(), new(presenter.SpoonsCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "spoons.helpTitle",
				CommandKeys: []string{
					"spoons.helpPass",
					"spoons.helpGrab",
					"spoons.helpNext",
				},
				ExtraCommandLines: []string{"  l                    action log"},
			})
	}},
	{Name: "kemps", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewKempsCuiController(usecase.NewKempsInteractor(
				domain.NewDefaultKemps(), new(presenter.KempsCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "kemps.helpTitle",
				CommandKeys: []string{
					"kemps.helpSwap",
					"kemps.helpPass",
					"kemps.helpSignal",
					"kemps.helpKemps",
					"kemps.helpCounter",
					"kemps.helpNext",
				},
				ExtraCommandLines: []string{"  l                    action log"},
			})
	}},
	{Name: "cuckoo", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCuckooCuiController(usecase.NewCuckooInteractor(
				domain.NewDefaultCuckoo(), new(presenter.CuckooCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "cuckoo.helpTitle",
				CommandKeys: []string{
					"cuckoo.helpKeep",
					"cuckoo.helpSwap",
					"cuckoo.helpRefuse",
					"cuckoo.helpAccept",
					"cuckoo.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys: []string{
					"cuckoo.helpSetDifficulty",
					"cuckoo.helpSetLives",
				},
			})
	}},
	{Name: "pishti", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPishtiCuiController(usecase.NewPishtiInteractor(
				domain.NewDefaultPishti(), new(presenter.PishtiCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "pishti.helpTitle",
				CommandKeys: []string{
					"pishti.helpPlay",
					"pishti.helpNext",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys: []string{
					"pishti.helpSetDifficulty",
					"pishti.helpSetPlayers",
				},
			})
	}},
	{Name: "cuarenta", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCuarentaCuiController(usecase.NewCuarentaInteractor(
				domain.NewDefaultCuarenta(), new(presenter.CuarentaCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "cuarenta.helpTitle",
				CommandKeys: []string{
					"cuarenta.helpPlay",
					"cuarenta.helpNext",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys: []string{
					"cuarenta.helpSetDifficulty",
				},
			})
	}},
	{Name: "fivecardstud", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewFiveCardStudCuiController(usecase.NewFiveCardStudInteractor(
				domain.NewDefaultFiveCardStud(), new(presenter.FiveCardStudCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "fivecardstud.helpTitle",
				CommandKeys: []string{
					"fivecardstud.helpFold",
					"fivecardstud.helpCheck",
					"fivecardstud.helpCall",
					"fivecardstud.helpBet",
					"fivecardstud.helpRaise",
					"fivecardstud.helpAllIn",
				},
				SettingKeys: []string{
					"fivecardstud.helpBettingLimit",
					"fivecardstud.helpTournament",
				},
			})
	}},
	{Name: "faro", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewFaroCuiController(usecase.NewFaroInteractor(
				domain.NewDefaultFaro(), new(presenter.FaroCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "faro.helpTitle",
				CommandKeys: []string{
					"faro.helpBet",
					"faro.helpClearBet",
					"faro.helpClearAll",
					"faro.helpDeal",
					"faro.helpCall",
					"faro.helpNext",
				},
				ExtraCommandLines: []string{"  l                    action log"},
			})
	}},
	{Name: "openfacechinese", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewOpenFaceChineseCuiController(usecase.NewOpenFaceChineseInteractor(
				domain.NewDefaultOpenFaceChinese(), new(presenter.OpenFaceChineseCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "openfacechinese.helpTitle",
				CommandKeys: []string{
					"openfacechinese.helpPlace",
					"openfacechinese.helpFront",
					"openfacechinese.helpMiddle",
					"openfacechinese.helpBack",
					"openfacechinese.helpNextRound",
				},
				SettingKeys: []string{
					"openfacechinese.helpSetDifficulty",
					"openfacechinese.helpSetPlayers",
				},
			})
	}},
	{Name: "russianbank", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewRussianBankCuiController(usecase.NewRussianBankInteractor(
				domain.NewDefaultRussianBank(), new(presenter.RussianBankCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "russianbank.helpTitle",
				CommandKeys: []string{
					"russianbank.helpFoundation",
					"russianbank.helpTableau",
					"russianbank.helpDiscard",
					"russianbank.helpStop",
					"russianbank.helpUndo",
				},
				SettingKeys: []string{
					"russianbank.helpSetDifficulty",
				},
			})
	}},
	{Name: "labellelucie", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewLaBelleLucieCuiController(usecase.NewLaBelleLucieInteractor(
				domain.NewDefaultLaBelleLucie(), new(presenter.LaBelleLucieCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "labellelucie.helpTitle",
				CommandKeys: []string{
					"labellelucie.helpMove",
					"labellelucie.helpRedeal",
					"labellelucie.helpAutoComplete",
					"labellelucie.helpUndo",
					"labellelucie.helpGiveUp",
				},
			})
	}},
	{Name: "simplesimon", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSimpleSimonCuiController(usecase.NewSimpleSimonInteractor(
				domain.NewDefaultSimpleSimon(), new(presenter.SimpleSimonCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "simplesimon.helpTitle",
				CommandKeys: []string{
					"simplesimon.helpMove",
					"simplesimon.helpUndo",
					"simplesimon.helpGiveUp",
				},
			})
	}},
	{Name: "doubleklondike", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewDoubleKlondikeCuiController(usecase.NewDoubleKlondikeInteractor(
				domain.NewDefaultDoubleKlondike(), new(presenter.DoubleKlondikeCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "doubleklondike.helpTitle",
				CommandKeys: []string{
					"doubleklondike.helpDraw",
					"doubleklondike.helpMoveWaste",
					"doubleklondike.helpMoveTableau",
					"doubleklondike.helpAuto",
					"doubleklondike.helpUndo",
					"doubleklondike.helpGiveUp",
				},
			})
	}},
	{Name: "blackhole", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBlackHoleCuiController(usecase.NewBlackHoleInteractor(
				domain.NewDefaultBlackHole(), new(presenter.BlackHoleCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "blackhole.helpTitle",
				CommandKeys: []string{
					"blackhole.helpMove",
					"blackhole.helpUndo",
					"blackhole.helpGiveUp",
				},
			})
	}},
	{Name: "beggarmyneighbour", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBeggarMyNeighbourCuiController(usecase.NewBeggarMyNeighbourInteractor(
				domain.NewDefaultBeggarMyNeighbour(), new(presenter.BeggarMyNeighbourCuiPresenter))),
			CuiHelpSpec{
				TitleKey:    "beggarmyneighbour.helpTitle",
				CommandKeys: []string{"beggarmyneighbour.helpStep", "beggarmyneighbour.helpAutoPlay"},
				SettingKeys: []string{"beggarmyneighbour.helpSetMax"},
			})
	}},
	{Name: "allfours", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewAllFoursCuiController(usecase.NewAllFoursInteractor(
				domain.NewDefaultAllFours(), new(presenter.AllFoursCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "allfours.helpTitle",
				CommandKeys: []string{
					"allfours.helpStand", "allfours.helpBeg", "allfours.helpGift",
					"allfours.helpRun", "allfours.helpPlay", "allfours.helpNext", "allfours.helpNextRound",
				},
				SettingKeys: []string{"allfours.helpSetDifficulty", "allfours.helpSetLimit"},
			})
	}},
	{Name: "prsi", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPrsiCuiController(usecase.NewPrsiInteractor(
				domain.NewDefaultPrsi(), new(presenter.PrsiCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "prsi.helpTitle",
				CommandKeys:       []string{"prsi.helpPlay", "prsi.helpDraw"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"prsi.helpSetDifficulty"},
			})
	}},
	{Name: "jass", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewJassCuiController(usecase.NewJassInteractor(
				domain.NewDefaultJass(), new(presenter.JassCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "jass.helpTitle",
				CommandKeys: []string{
					"jass.helpCall",
					"jass.helpSchieben",
					"jass.helpPlay",
					"jass.helpNext",
					"jass.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"jass.helpSetDifficulty", "jass.helpSetTarget"},
			})
	}},
	{Name: "gaigel", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewGaigelCuiController(usecase.NewGaigelInteractor(
				domain.NewDefaultGaigel(), new(presenter.GaigelCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "gaigel.helpTitle",
				CommandKeys: []string{
					"gaigel.helpPlay",
					"gaigel.helpMarriage",
					"gaigel.helpNext",
					"gaigel.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"gaigel.helpSetDifficulty", "gaigel.helpSetTarget"},
			})
	}},
	{Name: "tysiac", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewTysiacCuiController(usecase.NewTysiacInteractor(
				domain.NewDefaultTysiac(), new(presenter.TysiacCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "tysiac.helpTitle",
				CommandKeys: []string{
					"tysiac.helpBid",
					"tysiac.helpDiscard",
					"tysiac.helpPlay",
					"tysiac.helpNext",
					"tysiac.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"tysiac.helpSetDifficulty"},
			})
	}},
	{Name: "calabresella", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCalabresellaCuiController(usecase.NewCalabresellaInteractor(
				domain.NewDefaultCalabresella(), new(presenter.CalabresellaCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "calabresella.helpTitle",
				CommandKeys: []string{
					"calabresella.helpBid",
					"calabresella.helpDiscard",
					"calabresella.helpPlay",
					"calabresella.helpNext",
					"calabresella.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"calabresella.helpSetDifficulty"},
			})
	}},
	{Name: "ombre", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewOmbreCuiController(usecase.NewOmbreInteractor(
				domain.NewDefaultOmbre(), new(presenter.OmbreCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "ombre.helpTitle",
				CommandKeys: []string{
					"ombre.helpBid",
					"ombre.helpPlay",
					"ombre.helpNext",
					"ombre.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"ombre.helpSetDifficulty"},
			})
	}},
	{Name: "ulti", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewUltiCuiController(usecase.NewUltiInteractor(
				domain.NewDefaultUlti(), new(presenter.UltiCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "ulti.helpTitle",
				CommandKeys: []string{
					"ulti.helpBid",
					"ulti.helpDiscard",
					"ulti.helpPlay",
					"ulti.helpNext",
					"ulti.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"ulti.helpSetDifficulty"},
			})
	}},
	{Name: "king", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewKingCuiController(usecase.NewKingInteractor(
				domain.NewDefaultKing(), new(presenter.KingCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "king.helpTitle",
				CommandKeys:       []string{"king.helpContract", "king.helpPlay", "king.helpNext"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"king.helpSetDifficulty"},
			})
	}},
	{Name: "cinch", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCinchCuiController(usecase.NewCinchInteractor(
				domain.NewDefaultCinch(), new(presenter.CinchCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "cinch.helpTitle",
				CommandKeys:       []string{"cinch.helpBid", "cinch.helpTrump", "cinch.helpPlay", "cinch.helpNext"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"cinch.helpSetDifficulty"},
			})
	}},
	{Name: "loo", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewLooCuiController(usecase.NewLooInteractor(
				domain.NewDefaultLoo(), new(presenter.LooCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "loo.helpTitle",
				CommandKeys:       []string{"loo.helpDecide", "loo.helpPlay", "loo.helpNext"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"loo.helpSetDifficulty"},
			})
	}},
	{Name: "basra", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBasraCuiController(usecase.NewBasraInteractor(
				domain.NewDefaultBasra(), new(presenter.BasraCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "basra.helpTitle",
				CommandKeys:       []string{"basra.helpPlay", "basra.helpNext"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"basra.helpSetDifficulty"},
			})
	}},
	{Name: "tablanet", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewTablanetCuiController(usecase.NewTablanetInteractor(
				domain.NewDefaultTablanet(), new(presenter.TablanetCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "tablanet.helpTitle",
				CommandKeys:       []string{"tablanet.helpPlay", "tablanet.helpNext"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"tablanet.helpSetDifficulty"},
			})
	}},
	{Name: "trenteetquarante", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewTrenteEtQuaranteCuiController(usecase.NewTrenteEtQuaranteInteractor(
				domain.NewDefaultTrenteEtQuarante(), new(presenter.TrenteEtQuaranteCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "trenteetquarante.helpTitle",
				CommandKeys:       []string{"trenteetquarante.helpBet", "trenteetquarante.helpNext"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"trenteetquarante.helpSetDefaultBet"},
			})
	}},
	{Name: "guts", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewGutsCuiController(usecase.NewGutsInteractor(
				domain.NewDefaultGuts(), new(presenter.GutsCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "guts.helpTitle",
				CommandKeys:       []string{"guts.helpDeclare", "guts.helpNext"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"guts.helpSetPlayers", "guts.helpSetAnte", "guts.helpSetChips", "guts.helpSetRounds"},
			})
	}},
	{Name: "bouillotte", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewBouillotteCuiController(usecase.NewBouillotteInteractor(
				domain.NewDefaultBouillotte(), new(presenter.BouillotteCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "bouillotte.helpTitle",
				CommandKeys:       []string{"bouillotte.helpBet", "bouillotte.helpNext"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"bouillotte.helpSetPlayers", "bouillotte.helpSetAnte", "bouillotte.helpSetChips", "bouillotte.helpSetRounds"},
			})
	}},
	{Name: "primero", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPrimeroCuiController(usecase.NewPrimeroInteractor(
				domain.NewDefaultPrimero(), new(presenter.PrimeroCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "primero.helpTitle",
				CommandKeys:       []string{"primero.helpBet", "primero.helpNext"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"primero.helpSetPlayers", "primero.helpSetAnte", "primero.helpSetChips", "primero.helpSetRounds"},
			})
	}},
	{Name: "michigan", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewMichiganCuiController(usecase.NewMichiganInteractor(
				domain.NewDefaultMichigan(), new(presenter.MichiganCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "michigan.helpTitle",
				CommandKeys:       []string{"michigan.helpBet", "michigan.helpPlay", "michigan.helpNext"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"michigan.helpSetPlayers", "michigan.helpSetAnte", "michigan.helpSetChips", "michigan.helpSetRounds"},
			})
	}},
	{Name: "watten", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewWattenCuiController(usecase.NewWattenInteractor(
				domain.NewDefaultWatten(), new(presenter.WattenCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "watten.helpTitle",
				CommandKeys: []string{
					"watten.helpDeclare",
					"watten.helpPlay",
					"watten.helpRaise",
					"watten.helpRespond",
					"watten.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"watten.helpSetDifficulty", "watten.helpSetTarget"},
			})
	}},
	{Name: "carioca", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCariocaCuiController(usecase.NewCariocaInteractor(
				domain.NewDefaultCarioca(), new(presenter.CariocaCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "carioca.helpTitle",
				CommandKeys: []string{
					"carioca.helpDrawStock",
					"carioca.helpDrawDiscard",
					"carioca.helpMeldContract",
					"carioca.helpMeldExtra",
					"carioca.helpLayoff",
					"carioca.helpDiscard",
					"carioca.helpNextRound",
				},
				SettingKeys: []string{"carioca.helpSetPlayers", "carioca.helpSetDifficulty", "carioca.helpSetPenalty"},
			})
	}},
	{Name: "samba", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewSambaCuiController(usecase.NewSambaInteractor(
				domain.NewDefaultSamba(), new(presenter.SambaCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "samba.helpTitle",
				CommandKeys: []string{
					"samba.helpDrawStock",
					"samba.helpDrawDiscard",
					"samba.helpMeld",
					"samba.helpSkipMeld",
					"samba.helpDiscard",
					"samba.helpGoOut",
					"samba.helpNextRound",
				},
				SettingKeys: []string{"samba.helpSetDifficulty", "samba.helpSetLimit"},
			})
	}},
	{Name: "anaconda", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewAnacondaCuiController(usecase.NewAnacondaInteractor(
				domain.NewDefaultAnaconda(), new(presenter.AnacondaCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "anaconda.helpTitle",
				CommandKeys:       []string{"anaconda.helpPass", "anaconda.helpKeep", "anaconda.helpBet", "anaconda.helpNext"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"anaconda.helpSetPlayers", "anaconda.helpSetAnte", "anaconda.helpSetChips", "anaconda.helpSetRounds"},
			})
	}},
	{Name: "machiavelli", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewMachiavelliCuiController(usecase.NewMachiavelliInteractor(
				domain.NewDefaultMachiavelli(), new(presenter.MachiavelliCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "machiavelli.helpTitle",
				CommandKeys: []string{
					"machiavelli.helpDraw",
					"machiavelli.helpNewMeld",
					"machiavelli.helpLayoff",
					"machiavelli.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"machiavelli.helpSetPlayers", "machiavelli.helpSetDifficulty", "machiavelli.helpSetRounds"},
			})
	}},
	{Name: "pan", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewPanCuiController(usecase.NewPanInteractor(
				domain.NewDefaultPan(), new(presenter.PanCuiPresenter))),
			CuiHelpSpec{
				TitleKey: "pan.helpTitle",
				CommandKeys: []string{
					"pan.helpDrawStock",
					"pan.helpDrawDiscard",
					"pan.helpMeld",
					"pan.helpLayoff",
					"pan.helpDiscard",
					"pan.helpNextRound",
				},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"pan.helpSetPlayers", "pan.helpSetDifficulty", "pan.helpSetRounds"},
			})
	}},
	{Name: "wizard", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewWizardCuiController(usecase.NewWizardInteractor(
				domain.NewDefaultWizard(), new(presenter.WizardCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "wizard.helpTitle",
				CommandKeys:       []string{"wizard.helpBid", "wizard.helpPlay", "wizard.helpNext", "wizard.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"wizard.helpSetDifficulty"},
			})
	}},
	{Name: "oichokabu", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewOichoKabuCuiController(usecase.NewOichoKabuInteractor(
				domain.NewDefaultOichoKabu(), new(presenter.OichoKabuCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "oichokabu.helpTitle",
				CommandKeys:       []string{"oichokabu.helpBet", "oichokabu.helpDraw", "oichokabu.helpStand"},
				ExtraCommandLines: []string{"  log                  action log"},
			})
	}},
	{Name: "rook", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewRookCuiController(usecase.NewRookInteractor(
				domain.NewDefaultRook(), new(presenter.RookCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "rook.helpTitle",
				CommandKeys:       []string{"rook.helpBid", "rook.helpPass", "rook.helpExchange", "rook.helpPlay", "rook.helpNext", "rook.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"rook.helpSetDifficulty", "rook.helpSetTarget"},
			})
	}},
	{Name: "koikoi", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewKoiKoiCuiController(usecase.NewKoiKoiInteractor(
				domain.NewDefaultKoiKoi(), new(presenter.KoiKoiCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "koikoi.helpTitle",
				CommandKeys:       []string{"koikoi.helpPlay", "koikoi.helpKoiKoi", "koikoi.helpStop", "koikoi.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"koikoi.helpSetDifficulty"},
			})
	}},
	{Name: "gostop", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewGoStopCuiController(usecase.NewGoStopInteractor(
				domain.NewDefaultGoStop(), new(presenter.GoStopCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "gostop.helpTitle",
				CommandKeys:       []string{"gostop.helpPlay", "gostop.helpGo", "gostop.helpStop", "gostop.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"gostop.helpSetDifficulty"},
			})
	}},
	{Name: "hachihachi", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewHachiHachiCuiController(usecase.NewHachiHachiInteractor(
				domain.NewDefaultHachiHachi(), new(presenter.HachiHachiCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "hachihachi.helpTitle",
				CommandKeys:       []string{"hachihachi.helpPlay", "hachihachi.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"hachihachi.helpSetDifficulty"},
			})
	}},
	{Name: "frenchtarot", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewFrenchTarotCuiController(usecase.NewFrenchTarotInteractor(
				domain.NewDefaultFrenchTarot(), new(presenter.FrenchTarotCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "frenchtarot.helpTitle",
				CommandKeys:       []string{"frenchtarot.helpBid", "frenchtarot.helpPass", "frenchtarot.helpDiscard", "frenchtarot.helpPlay", "frenchtarot.helpNext", "frenchtarot.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"frenchtarot.helpSetDifficulty"},
			})
	}},
	{Name: "koenigrufen", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewKoenigrufenCuiController(usecase.NewKoenigrufenInteractor(
				domain.NewDefaultKoenigrufen(), new(presenter.KoenigrufenCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "koenigrufen.helpTitle",
				CommandKeys:       []string{"koenigrufen.helpBid", "koenigrufen.helpPass", "koenigrufen.helpCallKing", "koenigrufen.helpDiscard", "koenigrufen.helpPlay", "koenigrufen.helpNext", "koenigrufen.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"koenigrufen.helpSetDifficulty"},
			})
	}},
	{Name: "scarto", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewScartoCuiController(usecase.NewScartoInteractor(
				domain.NewDefaultScarto(), new(presenter.ScartoCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "scarto.helpTitle",
				CommandKeys:       []string{"scarto.helpScarto", "scarto.helpPlay", "scarto.helpNext", "scarto.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"scarto.helpSetDifficulty"},
			})
	}},
	{Name: "cego", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewCegoCuiController(usecase.NewCegoInteractor(
				domain.NewDefaultCego(), new(presenter.CegoCuiPresenter))),
			CuiHelpSpec{
				TitleKey:          "cego.helpTitle",
				CommandKeys:       []string{"cego.helpBid", "cego.helpPass", "cego.helpCego", "cego.helpHandspiel", "cego.helpDiscard", "cego.helpPlay", "cego.helpNext", "cego.helpNextRound"},
				ExtraCommandLines: []string{"  l                    action log"},
				SettingKeys:       []string{"cego.helpSetDifficulty"},
			})
	}},
	{Name: "zheng", NewCui: func() cuiGame {
		return cuiEntry(
			controller.NewZhengCuiController(usecase.NewZhengInteractor(
				domain.NewDefaultZheng(), new(presenter.ZhengCuiPresenter))),
			CuiHelpSpec{
				TitleKey:    "zheng.helpTitle",
				CommandKeys: []string{"zheng.helpPlay"},
				SettingKeys: []string{"zheng.helpSetDifficulty"},
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
	"7stud":   "sevencardstud",
	"7cs":     "sevencardstud",
	"clock":   "clocksolitaire",
	"crazy8":  "crazyeights",
	"indian":  "indianpoker",
	"video":   "videopoker",
	"deuces":  "deuceswild",
	"joker":   "jokerpoker",
	"short":   "shortdeck",
	"6plus":   "shortdeck",
	"gin":     "ginrummy",
	"3card":   "threecard",
	"csp":     "caribbeanstud",
	"stud":    "caribbeanstud",
	"oasis":   "oasispoker",
	"oasp":    "oasispoker",
	"thb":     "texasholdembonus",
	"thbp":    "texasholdembonus",
	"ch":      "casinoholdem",
	"choldem": "casinoholdem",
	"uth":     "ultimatetexasholdem",
	"uthe":    "ultimatetexasholdem",
	"40t":     "fortythieves",
	"pgp":     "paigow",
	"lir":     "letitride",
	"ride":    "letitride",
	"ms":      "mississippistud",
	"mstud":   "mississippistud",
	"sp21":    "spanish21",
	"s21":     "spanish21",
	"rummy":   "rummy500",
	"500":     "rummy500",
	"r500":    "rummy500",
	"cpk":     "chinesepoker",
	"cpkr":    "chinesepoker",
	"scg":     "sixcardgolf",
	"6golf":   "sixcardgolf",
	"ddz":     "doudizhu",
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
