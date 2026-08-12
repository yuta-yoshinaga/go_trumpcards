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

// BindCuiFor assembles one gameRegistry entry, giving the CLI the same
// registration shape the other four registration points already use:
// games_server.go has BindWebControllerFor and the per-category worker packages
// have RegisterKVGame, while this one was still hand-writing the nested
// constructors internal/CLAUDE.md tells you not to write.
//
// Two type parameters, not one: interactor constructors return a concrete type
// (*BlackJackInteractor) while CUI controller constructors take the interface
// (BlackJackInteractorIF), and Go function types are not covariant in their
// results. Constraining C by CuiExecer rather than returning it lets a
// controller constructor be passed by name. newInteractor must still be an
// interface-annotated closure for the same reason — which is why
// BindWebControllerFor's call sites look the way they do.
//
// newInteractor is called inside the NewCui closure, never here: gameRegistry is
// a package-level var, so building interactors eagerly would construct all 264
// games' domain state at process start instead of when one is played.
// See issue #5187.
func BindCuiFor[I any, C CuiExecer](
	name string,
	newInteractor func() I,
	newCtrl func(I) C,
	spec CuiHelpSpec,
) GameRegistryEntry {
	return GameRegistryEntry{
		Name: name,
		NewCui: func() cuiGame {
			return cuiEntry(newCtrl(newInteractor()), spec)
		},
	}
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
	BindCuiFor("blackjack",
		func() usecase.BlackJackInteractorIF {
			return usecase.NewBlackJackInteractor(domain.NewDefaultBlackJack(), new(presenter.BlackJackCuiPresenter))
		},
		controller.NewBlackJackCuiController,
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
		}),
	BindCuiFor("poker",
		func() usecase.PokerInteractorIF {
			return usecase.NewPokerInteractor(domain.NewDefaultPoker(), new(presenter.PokerCuiPresenter))
		},
		controller.NewPokerCuiController,
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
		}),
	BindCuiFor("oldmaid",
		func() usecase.OldMaidInteractorIF {
			return usecase.NewOldMaidInteractor(domain.NewDefaultOldMaid(), new(presenter.OldMaidCuiPresenter))
		},
		controller.NewOldMaidCuiController,
		CuiHelpSpec{
			TitleKey:    "oldmaid.helpTitle",
			CommandKeys: []string{"oldmaid.helpDraw", "oldmaid.helpShuffle", "oldmaid.helpReorder"},
			SettingKeys: []string{"oldmaid.helpSetMode", "oldmaid.helpSetPlacement", "oldmaid.helpSetMemoryAI"},
		}),
	BindCuiFor("daifugo",
		func() usecase.DaifugoInteractorIF {
			return usecase.NewDaifugoInteractor(domain.NewDefaultDaifugo(), new(presenter.DaifugoCuiPresenter))
		},
		controller.NewDaifugoCuiController,
		CuiHelpSpec{
			TitleKey:    "daifugo.helpTitle",
			CommandKeys: []string{"daifugo.helpPlay", "daifugo.helpSort"},
			SettingKeys: []string{"daifugo.helpSetDifficulty", "daifugo.helpSetJoker", "daifugo.helpSetRule"},
		}),
	BindCuiFor("bigtwo",
		func() usecase.BigTwoInteractorIF {
			return usecase.NewBigTwoInteractor(domain.NewDefaultBigTwo(), new(presenter.BigTwoCuiPresenter))
		},
		controller.NewBigTwoCuiController,
		CuiHelpSpec{
			TitleKey:    "bigtwo.helpTitle",
			CommandKeys: []string{"bigtwo.helpPlay"},
			SettingKeys: []string{"bigtwo.helpSetDifficulty"},
		}),
	BindCuiFor("sevens",
		func() usecase.SevensInteractorIF {
			return usecase.NewSevensInteractor(domain.NewDefaultSevens(), new(presenter.SevensCuiPresenter))
		},
		controller.NewSevensCuiController,
		CuiHelpSpec{
			TitleKey:      "sevens.helpTitle",
			CommandKeys:   []string{"sevens.helpPlay"},
			ResetOverride: "  r [tunnel] [joker=N] [strategy] [passes=N]  reset with options",
		}),
	{Name: "doubt", NewCui: func() cuiGame { return NewDoubtCui() }},
	BindCuiFor("holdem",
		func() usecase.HoldemInteractorIF {
			return usecase.NewHoldemInteractor(domain.NewDefaultHoldem(), new(presenter.HoldemCuiPresenter))
		},
		controller.NewHoldemCuiController,
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
		}),
	BindCuiFor("omaha",
		func() usecase.OmahaInteractorIF {
			return usecase.NewOmahaInteractor(domain.NewDefaultOmaha(), new(presenter.OmahaCuiPresenter))
		},
		controller.NewOmahaCuiController,
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
		}),
	BindCuiFor("omahahilo",
		func() usecase.OmahaInteractorIF {
			return usecase.NewOmahaInteractor(domain.NewDefaultOmahaHiLo(), new(presenter.OmahaCuiPresenter))
		},
		controller.NewOmahaCuiController,
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
		}),
	BindCuiFor("bigo",
		func() usecase.OmahaInteractorIF {
			return usecase.NewOmahaInteractor(domain.NewDefaultBigO(), new(presenter.OmahaCuiPresenter))
		},
		controller.NewOmahaCuiController,
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
		}),
	BindCuiFor("bigohilo",
		func() usecase.OmahaInteractorIF {
			return usecase.NewOmahaInteractor(domain.NewDefaultBigOHiLo(), new(presenter.OmahaCuiPresenter))
		},
		controller.NewOmahaCuiController,
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
		}),
	BindCuiFor("shortdeck",
		func() usecase.ShortDeckInteractorIF {
			return usecase.NewShortDeckInteractor(domain.NewDefaultShortDeck(), new(presenter.ShortDeckCuiPresenter))
		},
		controller.NewShortDeckCuiController,
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
		}),
	BindCuiFor("pineapple",
		func() usecase.PineappleInteractorIF {
			return usecase.NewPineappleInteractor(domain.NewDefaultPineapple(), new(presenter.PineappleCuiPresenter))
		},
		controller.NewPineappleCuiController,
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
		}),
	BindCuiFor("crazypineapple",
		func() usecase.PineappleInteractorIF {
			return usecase.NewPineappleInteractor(domain.NewDefaultCrazyPineapple(), new(presenter.PineappleCuiPresenter))
		},
		controller.NewPineappleCuiController,
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
		}),
	BindCuiFor("irishpoker",
		func() usecase.PineappleInteractorIF {
			return usecase.NewPineappleInteractor(domain.NewDefaultIrishPoker(), new(presenter.PineappleCuiPresenter))
		},
		controller.NewPineappleCuiController,
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
		}),
	BindCuiFor("hearts",
		func() usecase.HeartsInteractorIF {
			return usecase.NewHeartsInteractor(domain.NewDefaultHearts(), new(presenter.HeartsCuiPresenter))
		},
		controller.NewHeartsCuiController,
		CuiHelpSpec{
			TitleKey:          "hearts.helpTitle",
			CommandKeys:       []string{"hearts.helpPass", "hearts.helpPlay", "hearts.helpNext", "hearts.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"hearts.helpSetDifficulty", "hearts.helpSetLimit"},
		}),
	BindCuiFor("memory",
		func() usecase.MemoryInteractorIF {
			return usecase.NewMemoryInteractor(domain.NewDefaultMemory(), new(presenter.MemoryCuiPresenter))
		},
		controller.NewMemoryCuiController,
		CuiHelpSpec{
			TitleKey:          "memory.helpTitle",
			CommandKeys:       []string{"memory.helpFlip", "memory.helpNext"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"memory.helpSetDifficulty"},
		}),
	BindCuiFor("klondike",
		func() usecase.KlondikeInteractorIF {
			return usecase.NewKlondikeInteractor(domain.NewDefaultKlondike(), new(presenter.KlondikeCuiPresenter))
		},
		controller.NewKlondikeCuiController,
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
		}),
	BindCuiFor("freecell",
		func() usecase.FreeCellInteractorIF {
			return usecase.NewFreeCellInteractor(domain.NewDefaultFreeCell(), new(presenter.FreeCellCuiPresenter))
		},
		controller.NewFreeCellCuiController,
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
		}),
	BindCuiFor("seahaventowers",
		func() usecase.SeahavenTowersInteractorIF {
			return usecase.NewSeahavenTowersInteractor(domain.NewDefaultSeahavenTowers(), new(presenter.SeahavenTowersCuiPresenter))
		},
		controller.NewSeahavenTowersCuiController,
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
		}),
	BindCuiFor("cruel",
		func() usecase.CruelInteractorIF {
			return usecase.NewCruelInteractor(domain.NewDefaultCruel(), new(presenter.CruelCuiPresenter))
		},
		controller.NewCruelCuiController,
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
		}),
	BindCuiFor("baccarat",
		func() usecase.BaccaratInteractorIF {
			return usecase.NewBaccaratInteractor(domain.NewDefaultBaccarat(), new(presenter.BaccaratCuiPresenter))
		},
		controller.NewBaccaratCuiController,
		CuiHelpSpec{
			TitleKey:    "baccarat.helpTitle",
			CommandKeys: []string{"baccarat.helpBet"},
			ExtraCommandLines: []string{
				"  log                  action log",
				"  ch                   clear history",
			},
		}),
	BindCuiFor("spades",
		func() usecase.SpadesInteractorIF {
			return usecase.NewSpadesInteractor(domain.NewDefaultSpades(), new(presenter.SpadesCuiPresenter))
		},
		controller.NewSpadesCuiController,
		CuiHelpSpec{
			TitleKey:          "spades.helpTitle",
			CommandKeys:       []string{"spades.helpBid", "spades.helpPlay", "spades.helpNext", "spades.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"spades.helpSetDifficulty", "spades.helpSetLimit"},
		}),
	BindCuiFor("crazyeights",
		func() usecase.CrazyEightsInteractorIF {
			return usecase.NewCrazyEightsInteractor(domain.NewDefaultCrazyEights(), new(presenter.CrazyEightsCuiPresenter))
		},
		controller.NewCrazyEightsCuiController,
		CuiHelpSpec{
			TitleKey:          "crazyeights.helpTitle",
			CommandKeys:       []string{"crazyeights.helpPlay", "crazyeights.helpDraw", "crazyeights.helpSuit", "crazyeights.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"crazyeights.helpSetDifficulty", "crazyeights.helpSetLimit"},
		}),
	BindCuiFor("ginrummy",
		func() usecase.GinRummyInteractorIF {
			return usecase.NewGinRummyInteractor(domain.NewDefaultGinRummy(), new(presenter.GinRummyCuiPresenter))
		},
		controller.NewGinRummyCuiController,
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
		}),
	BindCuiFor("indianrummy",
		func() usecase.IndianRummyInteractorIF {
			return usecase.NewIndianRummyInteractor(domain.NewDefaultIndianRummy(), new(presenter.IndianRummyCuiPresenter))
		},
		controller.NewIndianRummyCuiController,
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
		}),
	BindCuiFor("canasta",
		func() usecase.CanastaInteractorIF {
			return usecase.NewCanastaInteractor(domain.NewDefaultCanasta(), new(presenter.CanastaCuiPresenter))
		},
		controller.NewCanastaCuiController,
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
		}}),
	BindCuiFor("spider",
		func() usecase.SpiderInteractorIF {
			return usecase.NewSpiderInteractor(domain.NewDefaultSpider(), new(presenter.SpiderCuiPresenter))
		},
		controller.NewSpiderCuiController,
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
		}),
	BindCuiFor("napoleon",
		func() usecase.NapoleonInteractorIF {
			return usecase.NewNapoleonInteractor(domain.NewDefaultNapoleon(), new(presenter.NapoleonCuiPresenter))
		},
		controller.NewNapoleonCuiController,
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
		}),
	BindCuiFor("indianpoker",
		func() usecase.IndianPokerInteractorIF {
			return usecase.NewIndianPokerInteractor(domain.NewDefaultIndianPoker(), new(presenter.IndianPokerCuiPresenter))
		},
		controller.NewIndianPokerCuiController,
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
		}),
	BindCuiFor("videopoker",
		func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(domain.NewDefaultVideoPoker(), new(presenter.VideoPokerCuiPresenter))
		},
		controller.NewVideoPokerCuiController,
		CuiHelpSpec{
			TitleKey:          "videopoker.helpTitle",
			CommandKeys:       []string{"videopoker.helpBet", "videopoker.helpHold", "videopoker.helpHint"},
			ExtraCommandLines: []string{"  log                  action log"},
		}),
	BindCuiFor("deuceswild",
		func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(domain.NewDeucesWildVideoPoker(), new(presenter.VideoPokerCuiPresenter))
		},
		controller.NewVideoPokerCuiController,
		CuiHelpSpec{
			TitleKey:          "deuceswild.helpTitle",
			CommandKeys:       []string{"videopoker.helpBet", "videopoker.helpHold", "videopoker.helpHint"},
			ExtraCommandLines: []string{"  log                  action log"},
		}),
	BindCuiFor("jokerpoker",
		func() usecase.VideoPokerInteractorIF {
			return usecase.NewVideoPokerInteractor(domain.NewJokerPokerVideoPoker(), new(presenter.VideoPokerCuiPresenter))
		},
		controller.NewVideoPokerCuiController,
		CuiHelpSpec{
			TitleKey:          "jokerpoker.helpTitle",
			CommandKeys:       []string{"videopoker.helpBet", "videopoker.helpHold", "videopoker.helpHint"},
			ExtraCommandLines: []string{"  log                  action log"},
		}),
	BindCuiFor("euchre",
		func() usecase.EuchreInteractorIF {
			return usecase.NewEuchreInteractor(domain.NewDefaultEuchre(), new(presenter.EuchreCuiPresenter))
		},
		controller.NewEuchreCuiController,
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
		}),
	BindCuiFor("pyramid",
		func() usecase.PyramidInteractorIF {
			return usecase.NewPyramidInteractor(domain.NewDefaultPyramid(), new(presenter.PyramidCuiPresenter))
		},
		controller.NewPyramidCuiController,
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
		}),
	BindCuiFor("tripeaks",
		func() usecase.TriPeaksInteractorIF {
			return usecase.NewTriPeaksInteractor(domain.NewDefaultTriPeaks(), new(presenter.TriPeaksCuiPresenter))
		},
		controller.NewTriPeaksCuiController,
		CuiHelpSpec{
			TitleKey: "tripeaks.helpTitle",
			CommandKeys: []string{
				"tripeaks.helpDraw",
				"tripeaks.helpRemove",
				"tripeaks.helpGiveUp",
				"tripeaks.helpHint",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("cribbage",
		func() usecase.CribbageInteractorIF {
			return usecase.NewCribbageInteractor(domain.NewDefaultCribbage(), new(presenter.CribbageCuiPresenter))
		},
		controller.NewCribbageCuiController,
		CuiHelpSpec{
			TitleKey:          "cribbage.helpTitle",
			CommandKeys:       []string{"cribbage.helpDiscard", "cribbage.helpCut", "cribbage.helpPeg", "cribbage.helpGo", "cribbage.helpHint", "cribbage.helpShowNext", "cribbage.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"cribbage.helpSetDifficulty", "cribbage.helpSetLimit"},
		}),
	BindCuiFor("threecard",
		func() usecase.ThreeCardInteractorIF {
			return usecase.NewThreeCardInteractor(domain.NewDefaultThreeCard(), new(presenter.ThreeCardCuiPresenter))
		},
		controller.NewThreeCardCuiController,
		CuiHelpSpec{
			TitleKey:          "threecard.helpTitle",
			CommandKeys:       []string{"threecard.helpBet", "threecard.helpPlay", "threecard.helpFold"},
			ExtraCommandLines: []string{"  log                  action log"},
		}),
	BindCuiFor("ohhell",
		func() usecase.OhHellInteractorIF {
			return usecase.NewOhHellInteractor(domain.NewDefaultOhHell(), new(presenter.OhHellCuiPresenter))
		},
		controller.NewOhHellCuiController,
		CuiHelpSpec{
			TitleKey:          "ohhell.helpTitle",
			CommandKeys:       []string{"ohhell.helpBid", "ohhell.helpPlay", "ohhell.helpNext", "ohhell.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"ohhell.helpSetDifficulty", "ohhell.helpSetMaxHand"},
		}),
	BindCuiFor("ninetynine",
		func() usecase.NinetyNineInteractorIF {
			return usecase.NewNinetyNineInteractor(domain.NewDefaultNinetyNine(), new(presenter.NinetyNineCuiPresenter))
		},
		controller.NewNinetyNineCuiController,
		CuiHelpSpec{
			TitleKey:          "ninetynine.helpTitle",
			CommandKeys:       []string{"ninetynine.helpBid", "ninetynine.helpPlay", "ninetynine.helpNext", "ninetynine.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"ninetynine.helpSetDifficulty", "ninetynine.helpSetTarget"},
		}),
	BindCuiFor("bridge",
		func() usecase.BridgeInteractorIF {
			return usecase.NewBridgeInteractor(domain.NewDefaultBridge(), new(presenter.BridgeCuiPresenter))
		},
		controller.NewBridgeCuiController,
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
		}}),
	BindCuiFor("speed",
		func() usecase.SpeedInteractorIF {
			return usecase.NewSpeedInteractor(domain.NewDefaultSpeed(), new(presenter.SpeedCuiPresenter))
		},
		controller.NewSpeedCuiController,
		CuiHelpSpec{
			TitleKey:    "speed.helpTitle",
			CommandKeys: []string{"speed.helpPlay", "speed.helpFlip", "speed.helpHint"},
			SettingKeys: []string{"speed.helpSetDifficulty"},
		}),
	BindCuiFor("gofish",
		func() usecase.GoFishInteractorIF {
			return usecase.NewGoFishInteractor(domain.NewDefaultGoFish(), new(presenter.GoFishCuiPresenter))
		},
		controller.NewGoFishCuiController,
		CuiHelpSpec{
			TitleKey:    "gofish.helpTitle",
			CommandKeys: []string{"gofish.helpAsk"},
			SettingKeys: []string{"gofish.helpSetDifficulty"},
		}),
	BindCuiFor("pinochle",
		func() usecase.PinochleInteractorIF {
			return usecase.NewPinochleInteractor(domain.NewDefaultPinochle(), new(presenter.PinochleCuiPresenter))
		},
		controller.NewPinochleCuiController,
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
		}),
	BindCuiFor("golf",
		func() usecase.GolfInteractorIF {
			return usecase.NewGolfInteractor(domain.NewDefaultGolf(), new(presenter.GolfCuiPresenter))
		},
		controller.NewGolfCuiController,
		CuiHelpSpec{
			TitleKey:          "golf.helpTitle",
			CommandKeys:       []string{"golf.helpDraw", "golf.helpRemove", "golf.helpGiveUp", "golf.helpHint"},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("pigtail",
		func() usecase.PigsTailInteractorIF {
			return usecase.NewPigsTailInteractor(domain.NewDefaultPigsTail(), new(presenter.PigsTailCuiPresenter))
		},
		controller.NewPigsTailCuiController,
		CuiHelpSpec{
			TitleKey:    "pigtail.helpTitle",
			CommandKeys: []string{"pigtail.helpAction"},
		}),
	BindCuiFor("sevencardstud",
		func() usecase.SevenCardStudInteractorIF {
			return usecase.NewSevenCardStudInteractor(domain.NewDefaultSevenCardStud(), new(presenter.SevenCardStudCuiPresenter))
		},
		controller.NewSevenCardStudCuiController,
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
		}),
	BindCuiFor("clocksolitaire",
		func() usecase.ClockSolitaireInteractorIF {
			return usecase.NewClockSolitaireInteractor(domain.NewDefaultClockSolitaire(), new(presenter.ClockSolitaireCuiPresenter))
		},
		controller.NewClockSolitaireCuiController,
		CuiHelpSpec{
			TitleKey:          "clocksolitaire.helpTitle",
			CommandKeys:       []string{"clocksolitaire.helpStep", "clocksolitaire.helpAutoPlay", "clocksolitaire.helpUndo"},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("durak",
		func() usecase.DurakInteractorIF {
			return usecase.NewDurakInteractor(domain.NewDefaultDurak(), new(presenter.DurakCuiPresenter))
		},
		controller.NewDurakCuiController,
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
		}}),
	BindCuiFor("fortythieves",
		func() usecase.FortyThievesInteractorIF {
			return usecase.NewFortyThievesInteractor(domain.NewDefaultFortyThieves(), new(presenter.FortyThievesCuiPresenter))
		},
		controller.NewFortyThievesCuiController,
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
		}),
	BindCuiFor("paigow",
		func() usecase.PaiGowInteractorIF {
			return usecase.NewPaiGowInteractor(domain.NewDefaultPaiGow(), new(presenter.PaiGowCuiPresenter))
		},
		controller.NewPaiGowCuiController,
		CuiHelpSpec{
			TitleKey:          "paigow.helpTitle",
			CommandKeys:       []string{"paigow.helpBet", "paigow.helpSet", "paigow.helpAuto", "paigow.helpHint"},
			ExtraCommandLines: []string{"  log                  action log"},
		}),
	BindCuiFor("twotenjack",
		func() usecase.TwoTenJackInteractorIF {
			return usecase.NewTwoTenJackInteractor(domain.NewDefaultTwoTenJack(), new(presenter.TwoTenJackCuiPresenter))
		},
		controller.NewTwoTenJackCuiController,
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
		}),
	BindCuiFor("caribbeanstud",
		func() usecase.CaribbeanStudInteractorIF {
			return usecase.NewCaribbeanStudInteractor(domain.NewDefaultCaribbeanStud(), new(presenter.CaribbeanStudCuiPresenter))
		},
		controller.NewCaribbeanStudCuiController,
		CuiHelpSpec{
			TitleKey:          "caribbeanstud.helpTitle",
			CommandKeys:       []string{"caribbeanstud.helpBet", "caribbeanstud.helpPlay", "caribbeanstud.helpFold"},
			ExtraCommandLines: []string{"  log                  action log"},
		}),
	BindCuiFor("texasholdembonus",
		func() usecase.TexasHoldemBonusInteractorIF {
			return usecase.NewTexasHoldemBonusInteractor(domain.NewDefaultTexasHoldemBonus(), new(presenter.TexasHoldemBonusCuiPresenter))
		},
		controller.NewTexasHoldemBonusCuiController,
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
		}),
	BindCuiFor("war",
		func() usecase.WarInteractorIF {
			return usecase.NewWarInteractor(domain.NewDefaultWar(), new(presenter.WarCuiPresenter))
		},
		controller.NewWarCuiController,
		CuiHelpSpec{
			TitleKey:    "war.helpTitle",
			CommandKeys: []string{"war.helpStep", "war.helpAutoPlay"},
			SettingKeys: []string{"war.helpSetMax"},
		}),
	BindCuiFor("canfield",
		func() usecase.CanfieldInteractorIF {
			return usecase.NewCanfieldInteractor(domain.NewDefaultCanfield(), new(presenter.CanfieldCuiPresenter))
		},
		controller.NewCanfieldCuiController,
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
		}),
	BindCuiFor("fiftyone",
		func() usecase.FiftyOneInteractorIF {
			return usecase.NewFiftyOneInteractor(domain.NewDefaultFiftyOne(), new(presenter.FiftyOneCuiPresenter))
		},
		controller.NewFiftyOneCuiController,
		CuiHelpSpec{
			TitleKey:    "fiftyone.helpTitle",
			CommandKeys: []string{"fiftyone.helpPlay", "fiftyone.helpAll", "fiftyone.helpStop"},
			SettingKeys: []string{"fiftyone.helpSetDifficulty"},
		}),
	BindCuiFor("yukon",
		func() usecase.YukonInteractorIF {
			return usecase.NewYukonInteractor(domain.NewDefaultYukon(), new(presenter.YukonCuiPresenter))
		},
		controller.NewYukonCuiController,
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
		}),
	BindCuiFor("russiansolitaire",
		func() usecase.RussianSolitaireInteractorIF {
			return usecase.NewRussianSolitaireInteractor(domain.NewDefaultRussianSolitaire(), new(presenter.RussianSolitaireCuiPresenter))
		},
		controller.NewRussianSolitaireCuiController,
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
		}),
	BindCuiFor("whist",
		func() usecase.WhistInteractorIF {
			return usecase.NewWhistInteractor(domain.NewDefaultWhist(), new(presenter.WhistCuiPresenter))
		},
		controller.NewWhistCuiController,
		CuiHelpSpec{
			TitleKey:          "whist.helpTitle",
			CommandKeys:       []string{"whist.helpPlay", "whist.helpNext", "whist.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"whist.helpSetDifficulty", "whist.helpSetLimit"},
		}),
	BindCuiFor("catchten",
		func() usecase.CatchTenInteractorIF {
			return usecase.NewCatchTenInteractor(domain.NewDefaultCatchTen(), new(presenter.CatchTenCuiPresenter))
		},
		controller.NewCatchTenCuiController,
		CuiHelpSpec{
			TitleKey:          "catchten.helpTitle",
			CommandKeys:       []string{"catchten.helpPlay", "catchten.helpNext", "catchten.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"catchten.helpSetDifficulty", "catchten.helpSetLimit"},
		}),
	BindCuiFor("letitride",
		func() usecase.LetItRideInteractorIF {
			return usecase.NewLetItRideInteractor(domain.NewDefaultLetItRide(), new(presenter.LetItRideCuiPresenter))
		},
		controller.NewLetItRideCuiController,
		CuiHelpSpec{
			TitleKey:          "letitride.helpTitle",
			CommandKeys:       []string{"letitride.helpBet", "letitride.helpPull", "letitride.helpLetItRide"},
			ExtraCommandLines: []string{"  log                  action log"},
		}),
	BindCuiFor("pokersquares",
		func() usecase.PokerSquaresInteractorIF {
			return usecase.NewPokerSquaresInteractor(domain.NewDefaultPokerSquares(), new(presenter.PokerSquaresCuiPresenter))
		},
		controller.NewPokerSquaresCuiController,
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
		}}),
	BindCuiFor("pageone",
		func() usecase.PageOneInteractorIF {
			return usecase.NewPageOneInteractor(domain.NewDefaultPageOne(), new(presenter.PageOneCuiPresenter))
		},
		controller.NewPageOneCuiController,
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
		}),
	BindCuiFor("reddog",
		func() usecase.RedDogInteractorIF {
			return usecase.NewRedDogInteractor(domain.NewDefaultRedDog(), new(presenter.RedDogCuiPresenter))
		},
		controller.NewRedDogCuiController,
		CuiHelpSpec{
			TitleKey:          "reddog.helpTitle",
			CommandKeys:       []string{"reddog.helpBet", "reddog.helpRaise", "reddog.helpStay", "reddog.helpHint"},
			ExtraCommandLines: []string{"  log                  action log"},
		}),
	BindCuiFor("badugi",
		func() usecase.BadugiInteractorIF {
			return usecase.NewBadugiInteractor(domain.NewDefaultBadugi(), new(presenter.BadugiCuiPresenter))
		},
		controller.NewBadugiCuiController,
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
		}),
	BindCuiFor("deucetoseven",
		func() usecase.DeuceToSevenInteractorIF {
			return usecase.NewDeuceToSevenInteractor(domain.NewDefaultDeuceToSeven(), new(presenter.DeuceToSevenCuiPresenter))
		},
		controller.NewDeuceToSevenCuiController,
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
		}),
	BindCuiFor("razz",
		func() usecase.SevenCardStudInteractorIF {
			return usecase.NewSevenCardStudInteractor(domain.NewDefaultRazz(), new(presenter.SevenCardStudCuiPresenter))
		},
		controller.NewSevenCardStudCuiController,
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
		}),
	BindCuiFor("sevencardstudhilo",
		func() usecase.SevenCardStudInteractorIF {
			return usecase.NewSevenCardStudInteractor(domain.NewDefaultSevenCardStudHiLo(), new(presenter.SevenCardStudCuiPresenter))
		},
		controller.NewSevenCardStudCuiController,
		CuiHelpSpec{
			TitleKey: "sevencardstudhilo.helpTitle",
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
		}),
	BindCuiFor("scorpion",
		func() usecase.ScorpionInteractorIF {
			return usecase.NewScorpionInteractor(domain.NewDefaultScorpion(), new(presenter.ScorpionCuiPresenter))
		},
		controller.NewScorpionCuiController,
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
		}),
	BindCuiFor("wasp",
		func() usecase.WaspInteractorIF {
			return usecase.NewWaspInteractor(domain.NewDefaultWasp(), new(presenter.WaspCuiPresenter))
		},
		controller.NewWaspCuiController,
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
		}),
	BindCuiFor("accordion",
		func() usecase.AccordionInteractorIF {
			return usecase.NewAccordionInteractor(domain.NewDefaultAccordion(), new(presenter.AccordionCuiPresenter))
		},
		controller.NewAccordionCuiController,
		CuiHelpSpec{
			TitleKey: "accordion.helpTitle",
			CommandKeys: []string{
				"accordion.helpMove",
				"accordion.helpGiveup",
				"accordion.helpHint",
				"accordion.helpLog",
				"accordion.helpUndo",
			},
		}),
	BindCuiFor("trash",
		func() usecase.TrashInteractorIF {
			return usecase.NewTrashInteractor(domain.NewDefaultTrash(), new(presenter.TrashCuiPresenter))
		},
		controller.NewTrashCuiController,
		CuiHelpSpec{
			TitleKey: "trash.helpTitle",
			CommandKeys: []string{
				"trash.helpDraw",
				"trash.helpPlace",
				"trash.helpCpu",
				"trash.helpHint",
				"trash.helpLog",
			},
		}),
	BindCuiFor("sevenbridge",
		func() usecase.SevenBridgeInteractorIF {
			return usecase.NewSevenBridgeInteractor(domain.NewDefaultSevenBridge(), new(presenter.SevenBridgeCuiPresenter))
		},
		controller.NewSevenBridgeCuiController,
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
		}),
	BindCuiFor("president",
		func() usecase.PresidentInteractorIF {
			return usecase.NewPresidentInteractor(domain.NewDefaultPresident(), new(presenter.PresidentCuiPresenter))
		},
		controller.NewPresidentCuiController,
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
		}),
	BindCuiFor("cassino",
		func() usecase.CassinoInteractorIF {
			return usecase.NewCassinoInteractor(domain.NewDefaultCassino(), new(presenter.CassinoCuiPresenter))
		},
		controller.NewCassinoCuiController,
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
		}),
	BindCuiFor("spanish21",
		func() usecase.BlackJackInteractorIF {
			return usecase.NewBlackJackInteractor(domain.NewSpanish21BlackJack(), new(presenter.BlackJackCuiPresenter))
		},
		controller.NewBlackJackCuiController,
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
		}),
	BindCuiFor("calculation",
		func() usecase.CalculationInteractorIF {
			return usecase.NewCalculationInteractor(domain.NewDefaultCalculation(), new(presenter.CalculationCuiPresenter))
		},
		controller.NewCalculationCuiController,
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
		}),
	BindCuiFor("sirtommy",
		func() usecase.SirTommyInteractorIF {
			return usecase.NewSirTommyInteractor(domain.NewDefaultSirTommy(), new(presenter.SirTommyCuiPresenter))
		},
		controller.NewSirTommyCuiController,
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
		}),
	BindCuiFor("bisley",
		func() usecase.BisleyInteractorIF {
			return usecase.NewBisleyInteractor(domain.NewDefaultBisley(), new(presenter.BisleyCuiPresenter))
		},
		controller.NewBisleyCuiController,
		CuiHelpSpec{
			TitleKey: "bisley.helpTitle",
			CommandKeys: []string{
				"bisley.helpMoveTA",
				"bisley.helpMoveTK",
				"bisley.helpMoveTT",
				"bisley.helpGiveUp",
				"bisley.helpHint",
				"bisley.helpAutoComplete",
				"bisley.helpUndo",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("napoleonssquare",
		func() usecase.NapoleonsSquareInteractorIF {
			return usecase.NewNapoleonsSquareInteractor(domain.NewDefaultNapoleonsSquare(), new(presenter.NapoleonsSquareCuiPresenter))
		},
		controller.NewNapoleonsSquareCuiController,
		CuiHelpSpec{
			TitleKey: "napoleonssquare.helpTitle",
			CommandKeys: []string{
				"napoleonssquare.helpDraw",
				"napoleonssquare.helpMoveWF",
				"napoleonssquare.helpMoveWT",
				"napoleonssquare.helpMoveTF",
				"napoleonssquare.helpMoveTT",
				"napoleonssquare.helpGiveUp",
				"napoleonssquare.helpHint",
				"napoleonssquare.helpAutoComplete",
				"napoleonssquare.helpUndo",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("grandfathersclock",
		func() usecase.GrandfathersClockInteractorIF {
			return usecase.NewGrandfathersClockInteractor(domain.NewDefaultGrandfathersClock(), new(presenter.GrandfathersClockCuiPresenter))
		},
		controller.NewGrandfathersClockCuiController,
		CuiHelpSpec{
			TitleKey: "grandfathersclock.helpTitle",
			CommandKeys: []string{
				"grandfathersclock.helpMoveTF",
				"grandfathersclock.helpMoveTT",
				"grandfathersclock.helpGiveUp",
				"grandfathersclock.helpHint",
				"grandfathersclock.helpAutoComplete",
				"grandfathersclock.helpUndo",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("missmilligan",
		func() usecase.MissMilliganInteractorIF {
			return usecase.NewMissMilliganInteractor(domain.NewDefaultMissMilligan(), new(presenter.MissMilliganCuiPresenter))
		},
		controller.NewMissMilliganCuiController,
		CuiHelpSpec{
			TitleKey: "missmilligan.helpTitle",
			CommandKeys: []string{
				"missmilligan.helpDeal",
				"missmilligan.helpMoveTF",
				"missmilligan.helpMoveTT",
				"missmilligan.helpWaive",
				"missmilligan.helpMoveWT",
				"missmilligan.helpMoveWF",
				"missmilligan.helpGiveUp",
				"missmilligan.helpHint",
				"missmilligan.helpAutoComplete",
				"missmilligan.helpUndo",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("duchess",
		func() usecase.DuchessInteractorIF {
			return usecase.NewDuchessInteractor(domain.NewDefaultDuchess(), new(presenter.DuchessCuiPresenter))
		},
		controller.NewDuchessCuiController,
		CuiHelpSpec{
			TitleKey: "duchess.helpTitle",
			CommandKeys: []string{
				"duchess.helpBase",
				"duchess.helpDraw",
				"duchess.helpMoveRF",
				"duchess.helpMoveRT",
				"duchess.helpMoveWF",
				"duchess.helpMoveWT",
				"duchess.helpMoveTF",
				"duchess.helpMoveTT",
				"duchess.helpGiveUp",
				"duchess.helpHint",
				"duchess.helpAutoComplete",
				"duchess.helpUndo",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("windmill",
		func() usecase.WindmillInteractorIF {
			return usecase.NewWindmillInteractor(domain.NewDefaultWindmill(), new(presenter.WindmillCuiPresenter))
		},
		controller.NewWindmillCuiController,
		CuiHelpSpec{
			TitleKey: "windmill.helpTitle",
			CommandKeys: []string{
				"windmill.helpDraw",
				"windmill.helpMoveSC",
				"windmill.helpMoveSK",
				"windmill.helpMoveWC",
				"windmill.helpMoveWK",
				"windmill.helpMoveKC",
				"windmill.helpGiveUp",
				"windmill.helpHint",
				"windmill.helpAutoComplete",
				"windmill.helpUndo",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("americantoad",
		func() usecase.AmericanToadInteractorIF {
			return usecase.NewAmericanToadInteractor(domain.NewDefaultAmericanToad(), new(presenter.AmericanToadCuiPresenter))
		},
		controller.NewAmericanToadCuiController,
		CuiHelpSpec{
			TitleKey: "americantoad.helpTitle",
			CommandKeys: []string{
				"americantoad.helpDraw",
				"americantoad.helpMoveRF",
				"americantoad.helpMoveRT",
				"americantoad.helpMoveWF",
				"americantoad.helpMoveWT",
				"americantoad.helpMoveTF",
				"americantoad.helpMoveTT",
				"americantoad.helpGiveUp",
				"americantoad.helpHint",
				"americantoad.helpAutoComplete",
				"americantoad.helpUndo",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("spiteandmalice",
		func() usecase.SpiteAndMaliceInteractorIF {
			return usecase.NewSpiteAndMaliceInteractor(domain.NewDefaultSpiteAndMalice(), new(presenter.SpiteAndMaliceCuiPresenter))
		},
		controller.NewSpiteAndMaliceCuiController,
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
		}),
	BindCuiFor("skat",
		func() usecase.SkatInteractorIF {
			return usecase.NewSkatInteractor(domain.NewDefaultSkat(), new(presenter.SkatCuiPresenter))
		},
		controller.NewSkatCuiController,
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
		}),
	BindCuiFor("congress",
		func() usecase.CongressInteractorIF {
			return usecase.NewCongressInteractor(domain.NewDefaultCongress(), new(presenter.CongressCuiPresenter))
		},
		controller.NewCongressCuiController,
		CuiHelpSpec{
			TitleKey: "congress.helpTitle",
			CommandKeys: []string{
				"congress.helpDraw",
				"congress.helpMoveTF",
				"congress.helpMoveTT",
				"congress.helpMoveWF",
				"congress.helpMoveWT",
				"congress.helpMoveST",
				"congress.helpGiveUp",
				"congress.helpHint",
				"congress.helpAutoComplete",
				"congress.helpUndo",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("terrace",
		func() usecase.TerraceInteractorIF {
			return usecase.NewTerraceInteractor(domain.NewDefaultTerrace(), new(presenter.TerraceCuiPresenter))
		},
		controller.NewTerraceCuiController,
		CuiHelpSpec{
			TitleKey: "terrace.helpTitle",
			CommandKeys: []string{
				"terrace.helpDraw",
				"terrace.helpMoveRF",
				"terrace.helpMoveWF",
				"terrace.helpMoveWT",
				"terrace.helpMoveTF",
				"terrace.helpMoveTT",
				"terrace.helpGiveUp",
				"terrace.helpHint",
				"terrace.helpAutoComplete",
				"terrace.helpUndo",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("braid",
		func() usecase.BraidInteractorIF {
			return usecase.NewBraidInteractor(domain.NewDefaultBraid(), new(presenter.BraidCuiPresenter))
		},
		controller.NewBraidCuiController,
		CuiHelpSpec{
			TitleKey: "braid.helpTitle",
			CommandKeys: []string{
				"braid.helpDraw",
				"braid.helpDirection",
				"braid.helpMoveBrF",
				"braid.helpMoveFdF",
				"braid.helpMoveHpF",
				"braid.helpMoveWF",
				"braid.helpMoveWHp",
				"braid.helpGiveUp",
				"braid.helpHint",
				"braid.helpAutoComplete",
				"braid.helpUndo",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("pontoon",
		func() usecase.PontoonInteractorIF {
			return usecase.NewPontoonInteractor(domain.NewDefaultPontoon(), new(presenter.PontoonCuiPresenter))
		},
		controller.NewPontoonCuiController,
		CuiHelpSpec{
			TitleKey: "pontoon.helpTitle",
			CommandKeys: []string{
				"pontoon.helpBet",
				"pontoon.helpDeal",
				"pontoon.helpStick",
				"pontoon.helpTwist",
				"pontoon.helpBuy",
				"pontoon.helpSplit",
				"pontoon.helpBankerTwist",
				"pontoon.helpBankerStay",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("settemezzo",
		func() usecase.SetteEMezzoInteractorIF {
			return usecase.NewSetteEMezzoInteractor(domain.NewDefaultSetteEMezzo(), new(presenter.SetteEMezzoCuiPresenter))
		},
		controller.NewSetteEMezzoCuiController,
		CuiHelpSpec{
			TitleKey: "settemezzo.helpTitle",
			CommandKeys: []string{
				"settemezzo.helpBet",
				"settemezzo.helpDeal",
				"settemezzo.helpHit",
				"settemezzo.helpStand",
				"settemezzo.helpMatta",
				"settemezzo.helpBankerHit",
				"settemezzo.helpBankerStand",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("niuniu",
		func() usecase.NiuNiuInteractorIF {
			return usecase.NewNiuNiuInteractor(domain.NewDefaultNiuNiu(), new(presenter.NiuNiuCuiPresenter))
		},
		controller.NewNiuNiuCuiController,
		CuiHelpSpec{
			TitleKey:          "niuniu.helpTitle",
			CommandKeys:       []string{"niuniu.helpBet"},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("bura",
		func() usecase.BuraInteractorIF {
			return usecase.NewBuraInteractor(domain.NewDefaultBura(), new(presenter.BuraCuiPresenter))
		},
		controller.NewBuraCuiController,
		CuiHelpSpec{
			TitleKey: "bura.helpTitle",
			CommandKeys: []string{
				"bura.helpPlay",
				"bura.helpClaim",
				"bura.helpDeclare",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("mushi",
		func() usecase.MushiInteractorIF {
			return usecase.NewMushiInteractor(domain.NewDefaultMushi(), new(presenter.MushiCuiPresenter))
		},
		controller.NewMushiCuiController,
		CuiHelpSpec{
			TitleKey: "mushi.helpTitle",
			CommandKeys: []string{
				"mushi.helpPlay",
				"mushi.helpSelect",
				"mushi.helpNext",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("toepen",
		func() usecase.ToepenInteractorIF {
			return usecase.NewToepenInteractor(domain.NewDefaultToepen(), new(presenter.ToepenCuiPresenter))
		},
		controller.NewToepenCuiController,
		CuiHelpSpec{
			TitleKey: "toepen.helpTitle",
			CommandKeys: []string{
				"toepen.helpPlay",
				"toepen.helpToep",
				"toepen.helpStay",
				"toepen.helpFold",
				"toepen.helpRedeal",
				"toepen.helpNext",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("chineseten",
		func() usecase.ChineseTenInteractorIF {
			return usecase.NewChineseTenInteractor(domain.NewDefaultChineseTen(), new(presenter.ChineseTenCuiPresenter))
		},
		controller.NewChineseTenCuiController,
		CuiHelpSpec{
			TitleKey: "chineseten.helpTitle",
			CommandKeys: []string{
				"chineseten.helpPlay",
				"chineseten.helpSelect",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("trex",
		func() usecase.TrexInteractorIF {
			return usecase.NewTrexInteractor(domain.NewDefaultTrex(), new(presenter.TrexCuiPresenter))
		},
		controller.NewTrexCuiController,
		CuiHelpSpec{
			TitleKey: "trex.helpTitle",
			CommandKeys: []string{
				"trex.helpChoose",
				"trex.helpPlay",
				"trex.helpPass",
				"trex.helpNext",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("skitgubbe",
		func() usecase.SkitgubbeInteractorIF {
			return usecase.NewSkitgubbeInteractor(domain.NewDefaultSkitgubbe(), new(presenter.SkitgubbeCuiPresenter))
		},
		controller.NewSkitgubbeCuiController,
		CuiHelpSpec{
			TitleKey: "skitgubbe.helpTitle",
			CommandKeys: []string{
				"skitgubbe.helpPlay",
				"skitgubbe.helpPickUp",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("loba",
		func() usecase.LobaInteractorIF {
			return usecase.NewLobaInteractor(domain.NewDefaultLoba(), new(presenter.LobaCuiPresenter))
		},
		controller.NewLobaCuiController,
		CuiHelpSpec{
			TitleKey: "loba.helpTitle",
			CommandKeys: []string{
				"loba.helpDrawStock",
				"loba.helpDrawDiscard",
				"loba.helpMeld",
				"loba.helpLayOff",
				"loba.helpDiscard",
				"loba.helpNext",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("sjavs",
		func() usecase.SjavsInteractorIF {
			return usecase.NewSjavsInteractor(domain.NewDefaultSjavs(), new(presenter.SjavsCuiPresenter))
		},
		controller.NewSjavsCuiController,
		CuiHelpSpec{
			TitleKey: "sjavs.helpTitle",
			CommandKeys: []string{
				"sjavs.helpBid",
				"sjavs.helpPlay",
				"sjavs.helpNext",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("laughandliedown",
		func() usecase.LaughAndLieDownInteractorIF {
			return usecase.NewLaughAndLieDownInteractor(domain.NewDefaultLaughAndLieDown(), new(presenter.LaughAndLieDownCuiPresenter))
		},
		controller.NewLaughAndLieDownCuiController,
		CuiHelpSpec{
			TitleKey: "laughandliedown.helpTitle",
			CommandKeys: []string{
				"laughandliedown.helpPlay",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("shithead",
		func() usecase.ShitheadInteractorIF {
			return usecase.NewShitheadInteractor(domain.NewDefaultShithead(), new(presenter.ShitheadCuiPresenter))
		},
		controller.NewShitheadCuiController,
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
		}),
	BindCuiFor("nertz",
		func() usecase.NertzInteractorIF {
			return usecase.NewNertzInteractor(domain.NewDefaultNertz(), new(presenter.NertzCuiPresenter))
		},
		controller.NewNertzCuiController,
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
		}),
	BindCuiFor("slapjack",
		func() usecase.SlapjackInteractorIF {
			return usecase.NewSlapjackInteractor(domain.NewDefaultSlapjack(), new(presenter.SlapjackCuiPresenter))
		},
		controller.NewSlapjackCuiController,
		CuiHelpSpec{
			TitleKey: "slapjack.helpTitle",
			CommandKeys: []string{
				"slapjack.helpStep",
				"slapjack.helpSlap",
				"slapjack.helpTick",
				"slapjack.helpLog",
			},
			SettingKeys: []string{"slapjack.helpSetDifficulty"},
		}),
	BindCuiFor("egyptianratscrew",
		func() usecase.EgyptianRatscrewInteractorIF {
			return usecase.NewEgyptianRatscrewInteractor(domain.NewDefaultEgyptianRatscrew(), new(presenter.EgyptianRatscrewCuiPresenter))
		},
		controller.NewEgyptianRatscrewCuiController,
		CuiHelpSpec{
			TitleKey: "egyptianratscrew.helpTitle",
			CommandKeys: []string{
				"egyptianratscrew.helpStep",
				"egyptianratscrew.helpSlap",
				"egyptianratscrew.helpTick",
				"egyptianratscrew.helpLog",
			},
			SettingKeys: []string{"egyptianratscrew.helpSetDifficulty"},
		}),
	BindCuiFor("bakersdozen",
		func() usecase.BakersDozenInteractorIF {
			return usecase.NewBakersDozenInteractor(domain.NewDefaultBakersDozen(), new(presenter.BakersDozenCuiPresenter))
		},
		controller.NewBakersDozenCuiController,
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
		}),
	BindCuiFor("tonk",
		func() usecase.TonkInteractorIF {
			return usecase.NewTonkInteractor(domain.NewDefaultTonk(), new(presenter.TonkCuiPresenter))
		},
		controller.NewTonkCuiController,
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
		}),
	BindCuiFor("casinowar",
		func() usecase.CasinoWarInteractorIF {
			return usecase.NewCasinoWarInteractor(domain.NewDefaultCasinoWar(), new(presenter.CasinoWarCuiPresenter))
		},
		controller.NewCasinoWarCuiController,
		CuiHelpSpec{
			TitleKey:          "casinowar.helpTitle",
			CommandKeys:       []string{"casinowar.helpBet", "casinowar.helpSurrender", "casinowar.helpWar"},
			ExtraCommandLines: []string{"  log                  action log"},
		}),
	BindCuiFor("pitch",
		func() usecase.PitchInteractorIF {
			return usecase.NewPitchInteractor(domain.NewDefaultPitch(), new(presenter.PitchCuiPresenter))
		},
		controller.NewPitchCuiController,
		CuiHelpSpec{
			TitleKey:          "pitch.helpTitle",
			CommandKeys:       []string{"pitch.helpBid", "pitch.helpPlay", "pitch.helpNext", "pitch.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"pitch.helpSetDifficulty", "pitch.helpSetLimit"},
		}),
	BindCuiFor("dragontiger",
		func() usecase.DragonTigerInteractorIF {
			return usecase.NewDragonTigerInteractor(domain.NewDefaultDragonTiger(), new(presenter.DragonTigerCuiPresenter))
		},
		controller.NewDragonTigerCuiController,
		CuiHelpSpec{
			TitleKey:          "dragontiger.helpTitle",
			CommandKeys:       []string{"dragontiger.helpBet", "dragontiger.helpClear"},
			ExtraCommandLines: []string{"  log                  action log"},
		}),
	BindCuiFor("blackjackswitch",
		func() usecase.BlackJackSwitchInteractorIF {
			return usecase.NewBlackJackSwitchInteractor(domain.NewDefaultBlackJackSwitch(), new(presenter.BlackJackSwitchCuiPresenter))
		},
		controller.NewBlackJackSwitchCuiController,
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
		}),
	BindCuiFor("montecarlo",
		func() usecase.MonteCarloInteractorIF {
			return usecase.NewMonteCarloInteractor(domain.NewDefaultMonteCarlo(), new(presenter.MonteCarloCuiPresenter))
		},
		controller.NewMonteCarloCuiController,
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
		}),
	BindCuiFor("contractrummy",
		func() usecase.ContractRummyInteractorIF {
			return usecase.NewContractRummyInteractor(domain.NewDefaultContractRummy(), new(presenter.ContractRummyCuiPresenter))
		},
		controller.NewContractRummyCuiController,
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
		}),
	BindCuiFor("ultimatetexasholdem",
		func() usecase.UltimateTexasHoldemInteractorIF {
			return usecase.NewUltimateTexasHoldemInteractor(domain.NewDefaultUltimateTexasHoldem(), new(presenter.UltimateTexasHoldemCuiPresenter))
		},
		controller.NewUltimateTexasHoldemCuiController,
		CuiHelpSpec{
			TitleKey: "ultimatetexasholdem.helpTitle",
			CommandKeys: []string{
				"ultimatetexasholdem.helpBet",
				"ultimatetexasholdem.helpPlay",
				"ultimatetexasholdem.helpCheck",
				"ultimatetexasholdem.helpFold",
			},
			ExtraCommandLines: []string{"  log                  action log"},
		}),
	BindCuiFor("crescent",
		func() usecase.CrescentInteractorIF {
			return usecase.NewCrescentInteractor(domain.NewDefaultCrescent(), new(presenter.CrescentCuiPresenter))
		},
		controller.NewCrescentCuiController,
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
		}),
	BindCuiFor("mississippistud",
		func() usecase.MississippiStudInteractorIF {
			return usecase.NewMississippiStudInteractor(domain.NewDefaultMississippiStud(), new(presenter.MississippiStudCuiPresenter))
		},
		controller.NewMississippiStudCuiController,
		CuiHelpSpec{
			TitleKey: "mississippistud.helpTitle",
			CommandKeys: []string{
				"mississippistud.helpBet",
				"mississippistud.helpPlay",
				"mississippistud.helpFold",
			},
			ExtraCommandLines: []string{"  log                  action log"},
		}),
	BindCuiFor("belote",
		func() usecase.BeloteInteractorIF {
			return usecase.NewBeloteInteractor(domain.NewDefaultBelote(), new(presenter.BeloteCuiPresenter))
		},
		controller.NewBeloteCuiController,
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
		}),
	BindCuiFor("spiderette",
		func() usecase.SpideretteInteractorIF {
			return usecase.NewSpideretteInteractor(domain.NewDefaultSpiderette(), new(presenter.SpideretteCuiPresenter))
		},
		controller.NewSpideretteCuiController,
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
		}),
	BindCuiFor("mighty",
		func() usecase.MightyInteractorIF {
			return usecase.NewMightyInteractor(domain.NewDefaultMighty(), new(presenter.MightyCuiPresenter))
		},
		controller.NewMightyCuiController,
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
		}),
	BindCuiFor("oasispoker",
		func() usecase.OasisPokerInteractorIF {
			return usecase.NewOasisPokerInteractor(domain.NewDefaultOasisPoker(), new(presenter.OasisPokerCuiPresenter))
		},
		controller.NewOasisPokerCuiController,
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
		}),
	BindCuiFor("beleagueredcastle",
		func() usecase.BeleagueredCastleInteractorIF {
			return usecase.NewBeleagueredCastleInteractor(domain.NewDefaultBeleagueredCastle(), new(presenter.BeleagueredCastleCuiPresenter))
		},
		controller.NewBeleagueredCastleCuiController,
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
		}),
	BindCuiFor("streetsandalleys",
		func() usecase.StreetsAndAlleysInteractorIF {
			return usecase.NewStreetsAndAlleysInteractor(domain.NewDefaultStreetsAndAlleys(), new(presenter.StreetsAndAlleysCuiPresenter))
		},
		controller.NewStreetsAndAlleysCuiController,
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
		}),
	BindCuiFor("kingalbert",
		func() usecase.KingAlbertInteractorIF {
			return usecase.NewKingAlbertInteractor(domain.NewDefaultKingAlbert(), new(presenter.KingAlbertCuiPresenter))
		},
		controller.NewKingAlbertCuiController,
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
		}),
	BindCuiFor("flowergarden",
		func() usecase.FlowerGardenInteractorIF {
			return usecase.NewFlowerGardenInteractor(domain.NewDefaultFlowerGarden(), new(presenter.FlowerGardenCuiPresenter))
		},
		controller.NewFlowerGardenCuiController,
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
		}),
	BindCuiFor("fortyandeight",
		func() usecase.FortyAndEightInteractorIF {
			return usecase.NewFortyAndEightInteractor(domain.NewDefaultFortyAndEight(), new(presenter.FortyAndEightCuiPresenter))
		},
		controller.NewFortyAndEightCuiController,
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
		}),
	BindCuiFor("agnes",
		func() usecase.AgnesInteractorIF {
			return usecase.NewAgnesInteractor(domain.NewDefaultAgnes(), new(presenter.AgnesCuiPresenter))
		},
		controller.NewAgnesCuiController,
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
		}),
	BindCuiFor("sultan",
		func() usecase.SultanInteractorIF {
			return usecase.NewSultanInteractor(domain.NewDefaultSultan(), new(presenter.SultanCuiPresenter))
		},
		controller.NewSultanCuiController,
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
		}),
	BindCuiFor("piquet",
		func() usecase.PiquetInteractorIF {
			return usecase.NewPiquetInteractor(domain.NewDefaultPiquet(), new(presenter.PiquetCuiPresenter))
		},
		controller.NewPiquetCuiController,
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
		}),
	BindCuiFor("casinoholdem",
		func() usecase.CasinoHoldemInteractorIF {
			return usecase.NewCasinoHoldemInteractor(domain.NewDefaultCasinoHoldem(), new(presenter.CasinoHoldemCuiPresenter))
		},
		controller.NewCasinoHoldemCuiController,
		CuiHelpSpec{
			TitleKey: "casinoholdem.helpTitle",
			CommandKeys: []string{
				"casinoholdem.helpBet",
				"casinoholdem.helpCall",
				"casinoholdem.helpFold",
			},
			ExtraCommandLines: []string{"  log                  action log"},
		}),
	BindCuiFor("callbreak",
		func() usecase.CallBreakInteractorIF {
			return usecase.NewCallBreakInteractor(domain.NewDefaultCallBreak(), new(presenter.CallBreakCuiPresenter))
		},
		controller.NewCallBreakCuiController,
		CuiHelpSpec{
			TitleKey:          "callbreak.helpTitle",
			CommandKeys:       []string{"callbreak.helpBid", "callbreak.helpPlay", "callbreak.helpNext", "callbreak.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"callbreak.helpSetDifficulty", "callbreak.helpSetRounds"},
		}),
	BindCuiFor("tarneeb",
		func() usecase.TarneebInteractorIF {
			return usecase.NewTarneebInteractor(domain.NewDefaultTarneeb(), new(presenter.TarneebCuiPresenter))
		},
		controller.NewTarneebCuiController,
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
		}),
	BindCuiFor("highcardflush",
		func() usecase.HighCardFlushInteractorIF {
			return usecase.NewHighCardFlushInteractor(domain.NewDefaultHighCardFlush(), new(presenter.HighCardFlushCuiPresenter))
		},
		controller.NewHighCardFlushCuiController,
		CuiHelpSpec{
			TitleKey:          "highcardflush.helpTitle",
			CommandKeys:       []string{"highcardflush.helpBet", "highcardflush.helpRaise", "highcardflush.helpFold"},
			ExtraCommandLines: []string{"  log                  action log"},
		}),
	BindCuiFor("briscola",
		func() usecase.BriscolaInteractorIF {
			return usecase.NewBriscolaInteractor(domain.NewDefaultBriscola(), new(presenter.BriscolaCuiPresenter))
		},
		controller.NewBriscolaCuiController,
		CuiHelpSpec{
			TitleKey:          "briscola.helpTitle",
			CommandKeys:       []string{"briscola.helpPlay", "briscola.helpNext"},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("gaps",
		func() usecase.GapsInteractorIF {
			return usecase.NewGapsInteractor(domain.NewDefaultGaps(), new(presenter.GapsCuiPresenter))
		},
		controller.NewGapsCuiController,
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
		}),
	BindCuiFor("fourcardpoker",
		func() usecase.FourCardPokerInteractorIF {
			return usecase.NewFourCardPokerInteractor(domain.NewDefaultFourCardPoker(), new(presenter.FourCardPokerCuiPresenter))
		},
		controller.NewFourCardPokerCuiController,
		CuiHelpSpec{
			TitleKey:          "fourcardpoker.helpTitle",
			CommandKeys:       []string{"fourcardpoker.helpBet", "fourcardpoker.helpPlay", "fourcardpoker.helpFold"},
			ExtraCommandLines: []string{"  log                  action log"},
		}),
	BindCuiFor("rummy500",
		func() usecase.Rummy500InteractorIF {
			return usecase.NewRummy500Interactor(domain.NewDefaultRummy500(), new(presenter.Rummy500CuiPresenter))
		},
		controller.NewRummy500CuiController,
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
		}),
	BindCuiFor("eightoff",
		func() usecase.EightOffInteractorIF {
			return usecase.NewEightOffInteractor(domain.NewDefaultEightOff(), new(presenter.EightOffCuiPresenter))
		},
		controller.NewEightOffCuiController,
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
		}),
	BindCuiFor("russianpoker",
		func() usecase.RussianPokerInteractorIF {
			return usecase.NewRussianPokerInteractor(domain.NewDefaultRussianPoker(), new(presenter.RussianPokerCuiPresenter))
		},
		controller.NewRussianPokerCuiController,
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
		}),
	BindCuiFor("penguin",
		func() usecase.PenguinInteractorIF {
			return usecase.NewPenguinInteractor(domain.NewDefaultPenguin(), new(presenter.PenguinCuiPresenter))
		},
		controller.NewPenguinCuiController,
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
		}),
	BindCuiFor("chinesepoker",
		func() usecase.ChinesePokerInteractorIF {
			return usecase.NewChinesePokerInteractor(domain.NewDefaultChinesePoker(), new(presenter.ChinesePokerCuiPresenter))
		},
		controller.NewChinesePokerCuiController,
		CuiHelpSpec{
			TitleKey:          "chinesepoker.helpTitle",
			CommandKeys:       []string{"chinesepoker.helpBet", "chinesepoker.helpSet"},
			ExtraCommandLines: []string{"  l                                            action log"},
		}),
	BindCuiFor("sixcardgolf",
		func() usecase.SixCardGolfInteractorIF {
			return usecase.NewSixCardGolfInteractor(domain.NewDefaultSixCardGolf(), new(presenter.SixCardGolfCuiPresenter))
		},
		controller.NewSixCardGolfCuiController,
		CuiHelpSpec{
			TitleKey:    "sixcardgolf.helpTitle",
			CommandKeys: []string{"sixcardgolf.helpFlipInitial", "sixcardgolf.helpDrawStock", "sixcardgolf.helpDrawDiscard", "sixcardgolf.helpSwap", "sixcardgolf.helpDiscard", "sixcardgolf.helpFlip", "sixcardgolf.helpSkipFlip", "sixcardgolf.helpNextRound"},
			SettingKeys: []string{"sixcardgolf.helpSetDifficulty", "sixcardgolf.helpSetPlayers", "sixcardgolf.helpSetRounds"},
		}),
	BindCuiFor("doudizhu",
		func() usecase.DoudizhuInteractorIF {
			return usecase.NewDoudizhuInteractor(domain.NewDefaultDoudizhu(), new(presenter.DoudizhuCuiPresenter))
		},
		controller.NewDoudizhuCuiController,
		CuiHelpSpec{
			TitleKey:    "doudizhu.helpTitle",
			CommandKeys: []string{"doudizhu.helpPlay", "doudizhu.helpBid"},
			SettingKeys: []string{"doudizhu.helpSetDifficulty"},
		}),
	BindCuiFor("truco",
		func() usecase.TrucoInteractorIF {
			return usecase.NewTrucoInteractor(domain.NewDefaultTruco(), new(presenter.TrucoCuiPresenter))
		},
		controller.NewTrucoCuiController,
		CuiHelpSpec{
			TitleKey:    "truco.helpTitle",
			CommandKeys: []string{"truco.helpPlay", "truco.helpTruco", "truco.helpRespond", "truco.helpNext"},
		}),
	BindCuiFor("scopa",
		func() usecase.ScopaInteractorIF {
			return usecase.NewScopaInteractor(domain.NewDefaultScopa(), new(presenter.ScopaCuiPresenter))
		},
		controller.NewScopaCuiController,
		CuiHelpSpec{
			TitleKey:    "scopa.helpTitle",
			CommandKeys: []string{"scopa.helpPlay", "scopa.helpNext"},
			SettingKeys: []string{"scopa.helpSetDifficulty"},
		}),
	BindCuiFor("acesup",
		func() usecase.AcesUpInteractorIF {
			return usecase.NewAcesUpInteractor(domain.NewDefaultAcesUp(), new(presenter.AcesUpCuiPresenter))
		},
		controller.NewAcesUpCuiController,
		CuiHelpSpec{
			TitleKey:          "acesup.helpTitle",
			CommandKeys:       []string{"acesup.helpDraw", "acesup.helpRemove", "acesup.helpMove", "acesup.helpGiveUp", "acesup.helpHint"},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("barbu",
		func() usecase.BarbuInteractorIF {
			return usecase.NewBarbuInteractor(domain.NewDefaultBarbu(), new(presenter.BarbuCuiPresenter))
		},
		controller.NewBarbuCuiController,
		CuiHelpSpec{
			TitleKey:    "barbu.helpTitle",
			CommandKeys: []string{"barbu.helpContract", "barbu.helpPlay", "barbu.helpNext"},
			SettingKeys: []string{"barbu.helpSetDifficulty"},
		}),
	BindCuiFor("macau",
		func() usecase.MacauInteractorIF {
			return usecase.NewMacauInteractor(domain.NewDefaultMacau(), new(presenter.MacauCuiPresenter))
		},
		controller.NewMacauCuiController,
		CuiHelpSpec{
			TitleKey:          "macau.helpTitle",
			CommandKeys:       []string{"macau.helpPlay", "macau.helpDraw", "macau.helpSuit", "macau.helpDeclare", "macau.helpSkipDeclare", "macau.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"macau.helpSetDifficulty", "macau.helpSetLimit"},
		}),
	BindCuiFor("thirtyone",
		func() usecase.ThirtyOneInteractorIF {
			return usecase.NewThirtyOneInteractor(domain.NewDefaultThirtyOne(), new(presenter.ThirtyOneCuiPresenter))
		},
		controller.NewThirtyOneCuiController,
		CuiHelpSpec{
			TitleKey: "thirtyone.helpTitle",
			CommandKeys: []string{
				"thirtyone.helpDrawStock",
				"thirtyone.helpDrawDiscard",
				"thirtyone.helpDiscard",
				"thirtyone.helpKnock",
				"thirtyone.helpNextRound",
				"thirtyone.helpHint",
			},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"thirtyone.helpSetDifficulty", "thirtyone.helpSetLives"},
		}),
	BindCuiFor("tienlen",
		func() usecase.TienLenInteractorIF {
			return usecase.NewTienLenInteractor(domain.NewDefaultTienLen(), new(presenter.TienLenCuiPresenter))
		},
		controller.NewTienLenCuiController,
		CuiHelpSpec{
			TitleKey:    "tienlen.helpTitle",
			CommandKeys: []string{"tienlen.helpPlay"},
			SettingKeys: []string{"tienlen.helpSetDifficulty"},
		}),
	BindCuiFor("osmosis",
		func() usecase.OsmosisInteractorIF {
			return usecase.NewOsmosisInteractor(domain.NewDefaultOsmosis(), new(presenter.OsmosisCuiPresenter))
		},
		controller.NewOsmosisCuiController,
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
		}),
	BindCuiFor("fivehundred",
		func() usecase.FiveHundredInteractorIF {
			return usecase.NewFiveHundredInteractor(domain.NewDefaultFiveHundred(), new(presenter.FiveHundredCuiPresenter))
		},
		controller.NewFiveHundredCuiController,
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
		}),
	BindCuiFor("schnapsen",
		func() usecase.SchnapsenInteractorIF {
			return usecase.NewSchnapsenInteractor(domain.NewDefaultSchnapsen(), new(presenter.SchnapsenCuiPresenter))
		},
		controller.NewSchnapsenCuiController,
		CuiHelpSpec{
			TitleKey: "schnapsen.helpTitle",
			CommandKeys: []string{
				"schnapsen.helpPlay",
				"schnapsen.helpMarriage",
				"schnapsen.helpNext",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("burraco",
		func() usecase.BurracoInteractorIF {
			return usecase.NewBurracoInteractor(domain.NewDefaultBurraco(), new(presenter.BurracoCuiPresenter))
		},
		controller.NewBurracoCuiController,
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
		}}),
	BindCuiFor("yaniv",
		func() usecase.YanivInteractorIF {
			return usecase.NewYanivInteractor(domain.NewDefaultYaniv(), new(presenter.YanivCuiPresenter))
		},
		controller.NewYanivCuiController,
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
		}),
	BindCuiFor("gongzhu",
		func() usecase.GongZhuInteractorIF {
			return usecase.NewGongZhuInteractor(domain.NewDefaultGongZhu(), new(presenter.GongZhuCuiPresenter))
		},
		controller.NewGongZhuCuiController,
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
		}),
	BindCuiFor("bristol",
		func() usecase.BristolInteractorIF {
			return usecase.NewBristolInteractor(domain.NewDefaultBristol(), new(presenter.BristolCuiPresenter))
		},
		controller.NewBristolCuiController,
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
		}),
	BindCuiFor("bidwhist",
		func() usecase.BidWhistInteractorIF {
			return usecase.NewBidWhistInteractor(domain.NewDefaultBidWhist(), new(presenter.BidWhistCuiPresenter))
		},
		controller.NewBidWhistCuiController,
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
		}),
	BindCuiFor("tressette",
		func() usecase.TressetteInteractorIF {
			return usecase.NewTressetteInteractor(domain.NewDefaultTressette(), new(presenter.TressetteCuiPresenter))
		},
		controller.NewTressetteCuiController,
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
		}),
	BindCuiFor("easthaven",
		func() usecase.EasthavenInteractorIF {
			return usecase.NewEasthavenInteractor(domain.NewDefaultEasthaven(), new(presenter.EasthavenCuiPresenter))
		},
		controller.NewEasthavenCuiController,
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
		}),
	BindCuiFor("tichu",
		func() usecase.TichuInteractorIF {
			return usecase.NewTichuInteractor(domain.NewDefaultTichu(), new(presenter.TichuCuiPresenter))
		},
		controller.NewTichuCuiController,
		CuiHelpSpec{
			TitleKey:    "tichu.helpTitle",
			CommandKeys: []string{"tichu.helpPlay", "tichu.helpDeclare"},
			SettingKeys: []string{"tichu.helpSetDifficulty"},
		}),
	// Baker's Game reuses the FreeCell interactor/controller; only the
	// domain (same-suit stacking) and presenter (i18n namespace) differ.
	BindCuiFor("bakersgame",
		func() usecase.FreeCellInteractorIF {
			return usecase.NewFreeCellInteractor(domain.NewDefaultBakersGame(), new(presenter.BakersGameCuiPresenter))
		},
		controller.NewFreeCellCuiController,
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
		}),
	BindCuiFor("bourre",
		func() usecase.BourreInteractorIF {
			return usecase.NewBourreInteractor(domain.NewDefaultBourre(), new(presenter.BourreCuiPresenter))
		},
		controller.NewBourreCuiController,
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
		}),
	BindCuiFor("sheepshead",
		func() usecase.SheepsheadInteractorIF {
			return usecase.NewSheepsheadInteractor(domain.NewDefaultSheepshead(), new(presenter.SheepsheadCuiPresenter))
		},
		controller.NewSheepsheadCuiController,
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
		}),
	BindCuiFor("doppelkopf",
		func() usecase.DoppelkopfInteractorIF {
			return usecase.NewDoppelkopfInteractor(domain.NewDefaultDoppelkopf(), new(presenter.DoppelkopfCuiPresenter))
		},
		controller.NewDoppelkopfCuiController,
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
		}),
	BindCuiFor("mus",
		func() usecase.MusInteractorIF {
			return usecase.NewMusInteractor(domain.NewDefaultMus(), new(presenter.MusCuiPresenter))
		},
		controller.NewMusCuiController,
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
		}),
	BindCuiFor("tute",
		func() usecase.TuteInteractorIF {
			return usecase.NewTuteInteractor(domain.NewDefaultTute(), new(presenter.TuteCuiPresenter))
		},
		controller.NewTuteCuiController,
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
		}),
	BindCuiFor("sueca",
		func() usecase.SuecaInteractorIF {
			return usecase.NewSuecaInteractor(domain.NewDefaultSueca(), new(presenter.SuecaCuiPresenter))
		},
		controller.NewSuecaCuiController,
		CuiHelpSpec{
			TitleKey: "sueca.helpTitle",
			CommandKeys: []string{
				"sueca.helpPlay",
				"sueca.helpNext",
				"sueca.helpNextRound",
			},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"sueca.helpSetDifficulty"},
		}),
	BindCuiFor("fortyfives",
		func() usecase.FortyFivesInteractorIF {
			return usecase.NewFortyFivesInteractor(domain.NewDefaultFortyFives(), new(presenter.FortyFivesCuiPresenter))
		},
		controller.NewFortyFivesCuiController,
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
		}),
	BindCuiFor("twentynine",
		func() usecase.TwentyNineInteractorIF {
			return usecase.NewTwentyNineInteractor(domain.NewDefaultTwentyNine(), new(presenter.TwentyNineCuiPresenter))
		},
		controller.NewTwentyNineCuiController,
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
		}),
	BindCuiFor("klaverjas",
		func() usecase.KlaverjasInteractorIF {
			return usecase.NewKlaverjasInteractor(domain.NewDefaultKlaverjas(), new(presenter.KlaverjasCuiPresenter))
		},
		controller.NewKlaverjasCuiController,
		CuiHelpSpec{
			TitleKey: "klaverjas.helpTitle",
			CommandKeys: []string{
				"klaverjas.helpPlay",
				"klaverjas.helpNext",
				"klaverjas.helpNextRound",
			},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"klaverjas.helpSetDifficulty"},
		}),
	BindCuiFor("manille",
		func() usecase.ManilleInteractorIF {
			return usecase.NewManilleInteractor(domain.NewDefaultManille(), new(presenter.ManilleCuiPresenter))
		},
		controller.NewManilleCuiController,
		CuiHelpSpec{
			TitleKey: "manille.helpTitle",
			CommandKeys: []string{
				"manille.helpPlay",
				"manille.helpNext",
				"manille.helpNextRound",
			},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"manille.helpSetDifficulty"},
		}),
	BindCuiFor("marias",
		func() usecase.MariasInteractorIF {
			return usecase.NewMariasInteractor(domain.NewDefaultMarias(), new(presenter.MariasCuiPresenter))
		},
		controller.NewMariasCuiController,
		CuiHelpSpec{
			TitleKey: "marias.helpTitle",
			CommandKeys: []string{
				"marias.helpPlay",
				"marias.helpNext",
				"marias.helpNextRound",
			},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"marias.helpSetDifficulty"},
		}),
	BindCuiFor("sedma",
		func() usecase.SedmaInteractorIF {
			return usecase.NewSedmaInteractor(domain.NewDefaultSedma(), new(presenter.SedmaCuiPresenter))
		},
		controller.NewSedmaCuiController,
		CuiHelpSpec{
			TitleKey: "sedma.helpTitle",
			CommandKeys: []string{
				"sedma.helpPlay",
				"sedma.helpNext",
				"sedma.helpNextRound",
			},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"sedma.helpSetDifficulty"},
		}),
	BindCuiFor("solowhist",
		func() usecase.SoloWhistInteractorIF {
			return usecase.NewSoloWhistInteractor(domain.NewDefaultSoloWhist(), new(presenter.SoloWhistCuiPresenter))
		},
		controller.NewSoloWhistCuiController,
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
		}),
	BindCuiFor("knockoutwhist",
		func() usecase.KnockoutWhistInteractorIF {
			return usecase.NewKnockoutWhistInteractor(domain.NewDefaultKnockoutWhist(), new(presenter.KnockoutWhistCuiPresenter))
		},
		controller.NewKnockoutWhistCuiController,
		CuiHelpSpec{
			TitleKey: "knockoutwhist.helpTitle",
			CommandKeys: []string{
				"knockoutwhist.helpPlay",
				"knockoutwhist.helpNext",
				"knockoutwhist.helpNextRound",
			},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"knockoutwhist.helpSetDifficulty"},
		}),
	BindCuiFor("nap",
		func() usecase.NapInteractorIF {
			return usecase.NewNapInteractor(domain.NewDefaultNap(), new(presenter.NapCuiPresenter))
		},
		controller.NewNapCuiController,
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
		}),
	BindCuiFor("preference",
		func() usecase.PreferenceInteractorIF {
			return usecase.NewPreferenceInteractor(domain.NewDefaultPreference(), new(presenter.PreferenceCuiPresenter))
		},
		controller.NewPreferenceCuiController,
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
		}),
	BindCuiFor("ganjifa",
		func() usecase.GanjifaInteractorIF {
			return usecase.NewGanjifaInteractor(domain.NewDefaultGanjifa(), new(presenter.GanjifaCuiPresenter))
		},
		controller.NewGanjifaCuiController,
		CuiHelpSpec{
			TitleKey: "ganjifa.helpTitle",
			CommandKeys: []string{
				"ganjifa.helpPlay",
				"ganjifa.helpNext",
				"ganjifa.helpNextRound",
			},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"ganjifa.helpSetDifficulty"},
		}),
	BindCuiFor("vira",
		func() usecase.ViraInteractorIF {
			return usecase.NewViraInteractor(domain.NewDefaultVira(), new(presenter.ViraCuiPresenter))
		},
		controller.NewViraCuiController,
		CuiHelpSpec{
			TitleKey: "vira.helpTitle",
			CommandKeys: []string{
				"vira.helpBid",
				"vira.helpPass",
				"vira.helpPlay",
				"vira.helpNext",
				"vira.helpNextRound",
			},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"vira.helpSetDifficulty"},
		}),
	BindCuiFor("spoilfive",
		func() usecase.SpoilFiveInteractorIF {
			return usecase.NewSpoilFiveInteractor(domain.NewDefaultSpoilFive(), new(presenter.SpoilFiveCuiPresenter))
		},
		controller.NewSpoilFiveCuiController,
		CuiHelpSpec{
			TitleKey: "spoilfive.helpTitle",
			CommandKeys: []string{
				"spoilfive.helpPlay",
				"spoilfive.helpNext",
				"spoilfive.helpNextRound",
			},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"spoilfive.helpSetDifficulty"},
		}),
	BindCuiFor("courtpiece",
		func() usecase.CourtPieceInteractorIF {
			return usecase.NewCourtPieceInteractor(domain.NewDefaultCourtPiece(), new(presenter.CourtPieceCuiPresenter))
		},
		controller.NewCourtPieceCuiController,
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
		}),
	BindCuiFor("bezique",
		func() usecase.BeziqueInteractorIF {
			return usecase.NewBeziqueInteractor(domain.NewDefaultBezique(), new(presenter.BeziqueCuiPresenter))
		},
		controller.NewBeziqueCuiController,
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
		}),
	BindCuiFor("ecarte",
		func() usecase.EcarteInteractorIF {
			return usecase.NewEcarteInteractor(domain.NewDefaultEcarte(), new(presenter.EcarteCuiPresenter))
		},
		controller.NewEcarteCuiController,
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
		}),
	BindCuiFor("threecardbrag",
		func() usecase.ThreeCardBragInteractorIF {
			return usecase.NewThreeCardBragInteractor(domain.NewDefaultThreeCardBrag(), new(presenter.ThreeCardBragCuiPresenter))
		},
		controller.NewThreeCardBragCuiController,
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
		}),
	BindCuiFor("teenpatti",
		func() usecase.TeenPattiInteractorIF {
			return usecase.NewTeenPattiInteractor(domain.NewDefaultTeenPatti(), new(presenter.TeenPattiCuiPresenter))
		},
		controller.NewTeenPattiCuiController,
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
		}),
	BindCuiFor("scopone",
		func() usecase.ScoponeInteractorIF {
			return usecase.NewScoponeInteractor(domain.NewDefaultScopone(), new(presenter.ScoponeCuiPresenter))
		},
		controller.NewScoponeCuiController,
		CuiHelpSpec{
			TitleKey: "scopone.helpTitle",
			CommandKeys: []string{
				"scopone.helpPlay",
				"scopone.helpNext",
			},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"scopone.helpSetDifficulty"},
		}),
	BindCuiFor("escoba",
		func() usecase.EscobaInteractorIF {
			return usecase.NewEscobaInteractor(domain.NewDefaultEscoba(), new(presenter.EscobaCuiPresenter))
		},
		controller.NewEscobaCuiController,
		CuiHelpSpec{
			TitleKey: "escoba.helpTitle",
			CommandKeys: []string{
				"escoba.helpPlay",
				"escoba.helpNext",
			},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"escoba.helpSetDifficulty"},
		}),
	BindCuiFor("handandfoot",
		func() usecase.HandAndFootInteractorIF {
			return usecase.NewHandAndFootInteractor(domain.NewDefaultHandAndFoot(), new(presenter.HandAndFootCuiPresenter))
		},
		controller.NewHandAndFootCuiController,
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
		}),
	BindCuiFor("conquian",
		func() usecase.ConquianInteractorIF {
			return usecase.NewConquianInteractor(domain.NewDefaultConquian(), new(presenter.ConquianCuiPresenter))
		},
		controller.NewConquianCuiController,
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
		}),
	BindCuiFor("chinchon",
		func() usecase.ChinchonInteractorIF {
			return usecase.NewChinchonInteractor(domain.NewDefaultChinchon(), new(presenter.ChinchonCuiPresenter))
		},
		controller.NewChinchonCuiController,
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
		}),
	BindCuiFor("kalooki",
		func() usecase.KalookiInteractorIF {
			return usecase.NewKalookiInteractor(domain.NewDefaultKalooki(), new(presenter.KalookiCuiPresenter))
		},
		controller.NewKalookiCuiController,
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
		}),
	BindCuiFor("threethirteen",
		func() usecase.ThreeThirteenInteractorIF {
			return usecase.NewThreeThirteenInteractor(domain.NewDefaultThreeThirteen(), new(presenter.ThreeThirteenCuiPresenter))
		},
		controller.NewThreeThirteenCuiController,
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
		}),
	BindCuiFor("mao",
		func() usecase.MaoInteractorIF {
			return usecase.NewMaoInteractor(domain.NewDefaultMao(), new(presenter.MaoCuiPresenter))
		},
		controller.NewMaoCuiController,
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
		}),
	BindCuiFor("spoons",
		func() usecase.SpoonsInteractorIF {
			return usecase.NewSpoonsInteractor(domain.NewDefaultSpoons(), new(presenter.SpoonsCuiPresenter))
		},
		controller.NewSpoonsCuiController,
		CuiHelpSpec{
			TitleKey: "spoons.helpTitle",
			CommandKeys: []string{
				"spoons.helpPass",
				"spoons.helpGrab",
				"spoons.helpNext",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("kemps",
		func() usecase.KempsInteractorIF {
			return usecase.NewKempsInteractor(domain.NewDefaultKemps(), new(presenter.KempsCuiPresenter))
		},
		controller.NewKempsCuiController,
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
		}),
	BindCuiFor("cuckoo",
		func() usecase.CuckooInteractorIF {
			return usecase.NewCuckooInteractor(domain.NewDefaultCuckoo(), new(presenter.CuckooCuiPresenter))
		},
		controller.NewCuckooCuiController,
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
		}),
	BindCuiFor("pishti",
		func() usecase.PishtiInteractorIF {
			return usecase.NewPishtiInteractor(domain.NewDefaultPishti(), new(presenter.PishtiCuiPresenter))
		},
		controller.NewPishtiCuiController,
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
		}),
	BindCuiFor("cuarenta",
		func() usecase.CuarentaInteractorIF {
			return usecase.NewCuarentaInteractor(domain.NewDefaultCuarenta(), new(presenter.CuarentaCuiPresenter))
		},
		controller.NewCuarentaCuiController,
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
		}),
	BindCuiFor("fivecardstud",
		func() usecase.FiveCardStudInteractorIF {
			return usecase.NewFiveCardStudInteractor(domain.NewDefaultFiveCardStud(), new(presenter.FiveCardStudCuiPresenter))
		},
		controller.NewFiveCardStudCuiController,
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
		}),
	BindCuiFor("faro",
		func() usecase.FaroInteractorIF {
			return usecase.NewFaroInteractor(domain.NewDefaultFaro(), new(presenter.FaroCuiPresenter))
		},
		controller.NewFaroCuiController,
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
		}),
	BindCuiFor("openfacechinese",
		func() usecase.OpenFaceChineseInteractorIF {
			return usecase.NewOpenFaceChineseInteractor(domain.NewDefaultOpenFaceChinese(), new(presenter.OpenFaceChineseCuiPresenter))
		},
		controller.NewOpenFaceChineseCuiController,
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
		}),
	BindCuiFor("russianbank",
		func() usecase.RussianBankInteractorIF {
			return usecase.NewRussianBankInteractor(domain.NewDefaultRussianBank(), new(presenter.RussianBankCuiPresenter))
		},
		controller.NewRussianBankCuiController,
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
		}),
	BindCuiFor("labellelucie",
		func() usecase.LaBelleLucieInteractorIF {
			return usecase.NewLaBelleLucieInteractor(domain.NewDefaultLaBelleLucie(), new(presenter.LaBelleLucieCuiPresenter))
		},
		controller.NewLaBelleLucieCuiController,
		CuiHelpSpec{
			TitleKey: "labellelucie.helpTitle",
			CommandKeys: []string{
				"labellelucie.helpMove",
				"labellelucie.helpRedeal",
				"labellelucie.helpAutoComplete",
				"labellelucie.helpUndo",
				"labellelucie.helpGiveUp",
			},
		}),
	BindCuiFor("simplesimon",
		func() usecase.SimpleSimonInteractorIF {
			return usecase.NewSimpleSimonInteractor(domain.NewDefaultSimpleSimon(), new(presenter.SimpleSimonCuiPresenter))
		},
		controller.NewSimpleSimonCuiController,
		CuiHelpSpec{
			TitleKey: "simplesimon.helpTitle",
			CommandKeys: []string{
				"simplesimon.helpMove",
				"simplesimon.helpUndo",
				"simplesimon.helpGiveUp",
			},
		}),
	BindCuiFor("doubleklondike",
		func() usecase.DoubleKlondikeInteractorIF {
			return usecase.NewDoubleKlondikeInteractor(domain.NewDefaultDoubleKlondike(), new(presenter.DoubleKlondikeCuiPresenter))
		},
		controller.NewDoubleKlondikeCuiController,
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
		}),
	BindCuiFor("blackhole",
		func() usecase.BlackHoleInteractorIF {
			return usecase.NewBlackHoleInteractor(domain.NewDefaultBlackHole(), new(presenter.BlackHoleCuiPresenter))
		},
		controller.NewBlackHoleCuiController,
		CuiHelpSpec{
			TitleKey: "blackhole.helpTitle",
			CommandKeys: []string{
				"blackhole.helpMove",
				"blackhole.helpUndo",
				"blackhole.helpGiveUp",
			},
		}),
	BindCuiFor("beggarmyneighbour",
		func() usecase.BeggarMyNeighbourInteractorIF {
			return usecase.NewBeggarMyNeighbourInteractor(domain.NewDefaultBeggarMyNeighbour(), new(presenter.BeggarMyNeighbourCuiPresenter))
		},
		controller.NewBeggarMyNeighbourCuiController,
		CuiHelpSpec{
			TitleKey:    "beggarmyneighbour.helpTitle",
			CommandKeys: []string{"beggarmyneighbour.helpStep", "beggarmyneighbour.helpAutoPlay"},
			SettingKeys: []string{"beggarmyneighbour.helpSetMax"},
		}),
	BindCuiFor("allfours",
		func() usecase.AllFoursInteractorIF {
			return usecase.NewAllFoursInteractor(domain.NewDefaultAllFours(), new(presenter.AllFoursCuiPresenter))
		},
		controller.NewAllFoursCuiController,
		CuiHelpSpec{
			TitleKey: "allfours.helpTitle",
			CommandKeys: []string{
				"allfours.helpStand", "allfours.helpBeg", "allfours.helpGift",
				"allfours.helpRun", "allfours.helpPlay", "allfours.helpNext", "allfours.helpNextRound",
			},
			SettingKeys: []string{"allfours.helpSetDifficulty", "allfours.helpSetLimit"},
		}),
	BindCuiFor("prsi",
		func() usecase.PrsiInteractorIF {
			return usecase.NewPrsiInteractor(domain.NewDefaultPrsi(), new(presenter.PrsiCuiPresenter))
		},
		controller.NewPrsiCuiController,
		CuiHelpSpec{
			TitleKey:          "prsi.helpTitle",
			CommandKeys:       []string{"prsi.helpPlay", "prsi.helpDraw"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"prsi.helpSetDifficulty"},
		}),
	BindCuiFor("jass",
		func() usecase.JassInteractorIF {
			return usecase.NewJassInteractor(domain.NewDefaultJass(), new(presenter.JassCuiPresenter))
		},
		controller.NewJassCuiController,
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
		}),
	BindCuiFor("gaigel",
		func() usecase.GaigelInteractorIF {
			return usecase.NewGaigelInteractor(domain.NewDefaultGaigel(), new(presenter.GaigelCuiPresenter))
		},
		controller.NewGaigelCuiController,
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
		}),
	BindCuiFor("tysiac",
		func() usecase.TysiacInteractorIF {
			return usecase.NewTysiacInteractor(domain.NewDefaultTysiac(), new(presenter.TysiacCuiPresenter))
		},
		controller.NewTysiacCuiController,
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
		}),
	BindCuiFor("calabresella",
		func() usecase.CalabresellaInteractorIF {
			return usecase.NewCalabresellaInteractor(domain.NewDefaultCalabresella(), new(presenter.CalabresellaCuiPresenter))
		},
		controller.NewCalabresellaCuiController,
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
		}),
	BindCuiFor("ombre",
		func() usecase.OmbreInteractorIF {
			return usecase.NewOmbreInteractor(domain.NewDefaultOmbre(), new(presenter.OmbreCuiPresenter))
		},
		controller.NewOmbreCuiController,
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
		}),
	BindCuiFor("ulti",
		func() usecase.UltiInteractorIF {
			return usecase.NewUltiInteractor(domain.NewDefaultUlti(), new(presenter.UltiCuiPresenter))
		},
		controller.NewUltiCuiController,
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
		}),
	BindCuiFor("king",
		func() usecase.KingInteractorIF {
			return usecase.NewKingInteractor(domain.NewDefaultKing(), new(presenter.KingCuiPresenter))
		},
		controller.NewKingCuiController,
		CuiHelpSpec{
			TitleKey:          "king.helpTitle",
			CommandKeys:       []string{"king.helpContract", "king.helpPlay", "king.helpNext"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"king.helpSetDifficulty"},
		}),
	BindCuiFor("cinch",
		func() usecase.CinchInteractorIF {
			return usecase.NewCinchInteractor(domain.NewDefaultCinch(), new(presenter.CinchCuiPresenter))
		},
		controller.NewCinchCuiController,
		CuiHelpSpec{
			TitleKey:          "cinch.helpTitle",
			CommandKeys:       []string{"cinch.helpBid", "cinch.helpTrump", "cinch.helpPlay", "cinch.helpNext"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"cinch.helpSetDifficulty"},
		}),
	BindCuiFor("loo",
		func() usecase.LooInteractorIF {
			return usecase.NewLooInteractor(domain.NewDefaultLoo(), new(presenter.LooCuiPresenter))
		},
		controller.NewLooCuiController,
		CuiHelpSpec{
			TitleKey:          "loo.helpTitle",
			CommandKeys:       []string{"loo.helpDecide", "loo.helpPlay", "loo.helpNext"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"loo.helpSetDifficulty"},
		}),
	BindCuiFor("basra",
		func() usecase.BasraInteractorIF {
			return usecase.NewBasraInteractor(domain.NewDefaultBasra(), new(presenter.BasraCuiPresenter))
		},
		controller.NewBasraCuiController,
		CuiHelpSpec{
			TitleKey:          "basra.helpTitle",
			CommandKeys:       []string{"basra.helpPlay", "basra.helpNext"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"basra.helpSetDifficulty"},
		}),
	BindCuiFor("tablanet",
		func() usecase.TablanetInteractorIF {
			return usecase.NewTablanetInteractor(domain.NewDefaultTablanet(), new(presenter.TablanetCuiPresenter))
		},
		controller.NewTablanetCuiController,
		CuiHelpSpec{
			TitleKey:          "tablanet.helpTitle",
			CommandKeys:       []string{"tablanet.helpPlay", "tablanet.helpNext"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"tablanet.helpSetDifficulty"},
		}),
	BindCuiFor("trenteetquarante",
		func() usecase.TrenteEtQuaranteInteractorIF {
			return usecase.NewTrenteEtQuaranteInteractor(domain.NewDefaultTrenteEtQuarante(), new(presenter.TrenteEtQuaranteCuiPresenter))
		},
		controller.NewTrenteEtQuaranteCuiController,
		CuiHelpSpec{
			TitleKey:          "trenteetquarante.helpTitle",
			CommandKeys:       []string{"trenteetquarante.helpBet", "trenteetquarante.helpNext"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"trenteetquarante.helpSetDefaultBet"},
		}),
	BindCuiFor("guts",
		func() usecase.GutsInteractorIF {
			return usecase.NewGutsInteractor(domain.NewDefaultGuts(), new(presenter.GutsCuiPresenter))
		},
		controller.NewGutsCuiController,
		CuiHelpSpec{
			TitleKey:          "guts.helpTitle",
			CommandKeys:       []string{"guts.helpDeclare", "guts.helpNext"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"guts.helpSetPlayers", "guts.helpSetAnte", "guts.helpSetChips", "guts.helpSetRounds"},
		}),
	BindCuiFor("bouillotte",
		func() usecase.BouillotteInteractorIF {
			return usecase.NewBouillotteInteractor(domain.NewDefaultBouillotte(), new(presenter.BouillotteCuiPresenter))
		},
		controller.NewBouillotteCuiController,
		CuiHelpSpec{
			TitleKey:          "bouillotte.helpTitle",
			CommandKeys:       []string{"bouillotte.helpBet", "bouillotte.helpNext"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"bouillotte.helpSetPlayers", "bouillotte.helpSetAnte", "bouillotte.helpSetChips", "bouillotte.helpSetRounds"},
		}),
	BindCuiFor("primero",
		func() usecase.PrimeroInteractorIF {
			return usecase.NewPrimeroInteractor(domain.NewDefaultPrimero(), new(presenter.PrimeroCuiPresenter))
		},
		controller.NewPrimeroCuiController,
		CuiHelpSpec{
			TitleKey:          "primero.helpTitle",
			CommandKeys:       []string{"primero.helpBet", "primero.helpNext"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"primero.helpSetPlayers", "primero.helpSetAnte", "primero.helpSetChips", "primero.helpSetRounds"},
		}),
	BindCuiFor("michigan",
		func() usecase.MichiganInteractorIF {
			return usecase.NewMichiganInteractor(domain.NewDefaultMichigan(), new(presenter.MichiganCuiPresenter))
		},
		controller.NewMichiganCuiController,
		CuiHelpSpec{
			TitleKey:          "michigan.helpTitle",
			CommandKeys:       []string{"michigan.helpBet", "michigan.helpPlay", "michigan.helpNext"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"michigan.helpSetPlayers", "michigan.helpSetAnte", "michigan.helpSetChips", "michigan.helpSetRounds"},
		}),
	BindCuiFor("watten",
		func() usecase.WattenInteractorIF {
			return usecase.NewWattenInteractor(domain.NewDefaultWatten(), new(presenter.WattenCuiPresenter))
		},
		controller.NewWattenCuiController,
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
		}),
	BindCuiFor("carioca",
		func() usecase.CariocaInteractorIF {
			return usecase.NewCariocaInteractor(domain.NewDefaultCarioca(), new(presenter.CariocaCuiPresenter))
		},
		controller.NewCariocaCuiController,
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
		}),
	BindCuiFor("samba",
		func() usecase.SambaInteractorIF {
			return usecase.NewSambaInteractor(domain.NewDefaultSamba(), new(presenter.SambaCuiPresenter))
		},
		controller.NewSambaCuiController,
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
		}),
	BindCuiFor("anaconda",
		func() usecase.AnacondaInteractorIF {
			return usecase.NewAnacondaInteractor(domain.NewDefaultAnaconda(), new(presenter.AnacondaCuiPresenter))
		},
		controller.NewAnacondaCuiController,
		CuiHelpSpec{
			TitleKey:          "anaconda.helpTitle",
			CommandKeys:       []string{"anaconda.helpPass", "anaconda.helpKeep", "anaconda.helpBet", "anaconda.helpNext"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"anaconda.helpSetPlayers", "anaconda.helpSetAnte", "anaconda.helpSetChips", "anaconda.helpSetRounds"},
		}),
	BindCuiFor("machiavelli",
		func() usecase.MachiavelliInteractorIF {
			return usecase.NewMachiavelliInteractor(domain.NewDefaultMachiavelli(), new(presenter.MachiavelliCuiPresenter))
		},
		controller.NewMachiavelliCuiController,
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
		}),
	BindCuiFor("pan",
		func() usecase.PanInteractorIF {
			return usecase.NewPanInteractor(domain.NewDefaultPan(), new(presenter.PanCuiPresenter))
		},
		controller.NewPanCuiController,
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
		}),
	BindCuiFor("wizard",
		func() usecase.WizardInteractorIF {
			return usecase.NewWizardInteractor(domain.NewDefaultWizard(), new(presenter.WizardCuiPresenter))
		},
		controller.NewWizardCuiController,
		CuiHelpSpec{
			TitleKey:          "wizard.helpTitle",
			CommandKeys:       []string{"wizard.helpBid", "wizard.helpPlay", "wizard.helpNext", "wizard.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"wizard.helpSetDifficulty"},
		}),
	BindCuiFor("oichokabu",
		func() usecase.OichoKabuInteractorIF {
			return usecase.NewOichoKabuInteractor(domain.NewDefaultOichoKabu(), new(presenter.OichoKabuCuiPresenter))
		},
		controller.NewOichoKabuCuiController,
		CuiHelpSpec{
			TitleKey:          "oichokabu.helpTitle",
			CommandKeys:       []string{"oichokabu.helpBet", "oichokabu.helpDraw", "oichokabu.helpStand"},
			ExtraCommandLines: []string{"  log                  action log"},
		}),
	BindCuiFor("rook",
		func() usecase.RookInteractorIF {
			return usecase.NewRookInteractor(domain.NewDefaultRook(), new(presenter.RookCuiPresenter))
		},
		controller.NewRookCuiController,
		CuiHelpSpec{
			TitleKey:          "rook.helpTitle",
			CommandKeys:       []string{"rook.helpBid", "rook.helpPass", "rook.helpExchange", "rook.helpPlay", "rook.helpNext", "rook.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"rook.helpSetDifficulty", "rook.helpSetTarget"},
		}),
	BindCuiFor("koikoi",
		func() usecase.KoiKoiInteractorIF {
			return usecase.NewKoiKoiInteractor(domain.NewDefaultKoiKoi(), new(presenter.KoiKoiCuiPresenter))
		},
		controller.NewKoiKoiCuiController,
		CuiHelpSpec{
			TitleKey:          "koikoi.helpTitle",
			CommandKeys:       []string{"koikoi.helpPlay", "koikoi.helpKoiKoi", "koikoi.helpStop", "koikoi.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"koikoi.helpSetDifficulty"},
		}),
	BindCuiFor("gostop",
		func() usecase.GoStopInteractorIF {
			return usecase.NewGoStopInteractor(domain.NewDefaultGoStop(), new(presenter.GoStopCuiPresenter))
		},
		controller.NewGoStopCuiController,
		CuiHelpSpec{
			TitleKey:          "gostop.helpTitle",
			CommandKeys:       []string{"gostop.helpPlay", "gostop.helpGo", "gostop.helpStop", "gostop.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"gostop.helpSetDifficulty"},
		}),
	BindCuiFor("hachihachi",
		func() usecase.HachiHachiInteractorIF {
			return usecase.NewHachiHachiInteractor(domain.NewDefaultHachiHachi(), new(presenter.HachiHachiCuiPresenter))
		},
		controller.NewHachiHachiCuiController,
		CuiHelpSpec{
			TitleKey:          "hachihachi.helpTitle",
			CommandKeys:       []string{"hachihachi.helpPlay", "hachihachi.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"hachihachi.helpSetDifficulty"},
		}),
	BindCuiFor("frenchtarot",
		func() usecase.FrenchTarotInteractorIF {
			return usecase.NewFrenchTarotInteractor(domain.NewDefaultFrenchTarot(), new(presenter.FrenchTarotCuiPresenter))
		},
		controller.NewFrenchTarotCuiController,
		CuiHelpSpec{
			TitleKey:          "frenchtarot.helpTitle",
			CommandKeys:       []string{"frenchtarot.helpBid", "frenchtarot.helpPass", "frenchtarot.helpDiscard", "frenchtarot.helpPlay", "frenchtarot.helpNext", "frenchtarot.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"frenchtarot.helpSetDifficulty"},
		}),
	BindCuiFor("koenigrufen",
		func() usecase.KoenigrufenInteractorIF {
			return usecase.NewKoenigrufenInteractor(domain.NewDefaultKoenigrufen(), new(presenter.KoenigrufenCuiPresenter))
		},
		controller.NewKoenigrufenCuiController,
		CuiHelpSpec{
			TitleKey:          "koenigrufen.helpTitle",
			CommandKeys:       []string{"koenigrufen.helpBid", "koenigrufen.helpPass", "koenigrufen.helpCallKing", "koenigrufen.helpDiscard", "koenigrufen.helpPlay", "koenigrufen.helpNext", "koenigrufen.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"koenigrufen.helpSetDifficulty"},
		}),
	BindCuiFor("aluette",
		func() usecase.AluetteInteractorIF {
			return usecase.NewAluetteInteractor(domain.NewDefaultAluette(), new(presenter.AluetteCuiPresenter))
		},
		controller.NewAluetteCuiController,
		CuiHelpSpec{
			TitleKey: "aluette.helpTitle",
			CommandKeys: []string{
				"aluette.helpPlay",
				"aluette.helpNext",
				"aluette.helpNextRound",
			},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"aluette.helpSetDifficulty"},
		}),
	BindCuiFor("minchiate",
		func() usecase.MinchiateInteractorIF {
			return usecase.NewMinchiateInteractor(domain.NewDefaultMinchiate(), new(presenter.MinchiateCuiPresenter))
		},
		controller.NewMinchiateCuiController,
		CuiHelpSpec{
			TitleKey: "minchiate.helpTitle",
			CommandKeys: []string{
				"minchiate.helpScarto",
				"minchiate.helpPlay",
				"minchiate.helpNext",
				"minchiate.helpNextRound",
			},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"minchiate.helpSetDifficulty"},
		}),
	BindCuiFor("tarocchini",
		func() usecase.TarocchiniInteractorIF {
			return usecase.NewTarocchiniInteractor(domain.NewDefaultTarocchini(), new(presenter.TarocchiniCuiPresenter))
		},
		controller.NewTarocchiniCuiController,
		CuiHelpSpec{
			TitleKey: "tarocchini.helpTitle",
			CommandKeys: []string{
				"tarocchini.helpScarto",
				"tarocchini.helpPlay",
				"tarocchini.helpNext",
				"tarocchini.helpNextRound",
			},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"tarocchini.helpSetDifficulty"},
		}),
	BindCuiFor("scarto",
		func() usecase.ScartoInteractorIF {
			return usecase.NewScartoInteractor(domain.NewDefaultScarto(), new(presenter.ScartoCuiPresenter))
		},
		controller.NewScartoCuiController,
		CuiHelpSpec{
			TitleKey:          "scarto.helpTitle",
			CommandKeys:       []string{"scarto.helpScarto", "scarto.helpPlay", "scarto.helpNext", "scarto.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"scarto.helpSetDifficulty"},
		}),
	BindCuiFor("cego",
		func() usecase.CegoInteractorIF {
			return usecase.NewCegoInteractor(domain.NewDefaultCego(), new(presenter.CegoCuiPresenter))
		},
		controller.NewCegoCuiController,
		CuiHelpSpec{
			TitleKey:          "cego.helpTitle",
			CommandKeys:       []string{"cego.helpBid", "cego.helpPass", "cego.helpCego", "cego.helpHandspiel", "cego.helpDiscard", "cego.helpPlay", "cego.helpNext", "cego.helpNextRound"},
			ExtraCommandLines: []string{"  l                    action log"},
			SettingKeys:       []string{"cego.helpSetDifficulty"},
		}),
	BindCuiFor("zheng",
		func() usecase.ZhengInteractorIF {
			return usecase.NewZhengInteractor(domain.NewDefaultZheng(), new(presenter.ZhengCuiPresenter))
		},
		controller.NewZhengCuiController,
		CuiHelpSpec{
			TitleKey:    "zheng.helpTitle",
			CommandKeys: []string{"zheng.helpPlay"},
			SettingKeys: []string{"zheng.helpSetDifficulty"},
		}),
	BindCuiFor("desmoche",
		func() usecase.DesmocheInteractorIF {
			return usecase.NewDesmocheInteractor(domain.NewDefaultDesmoche(), new(presenter.DesmocheCuiPresenter))
		},
		controller.NewDesmocheCuiController,
		CuiHelpSpec{
			TitleKey: "desmoche.helpTitle",
			CommandKeys: []string{
				"desmoche.helpDrawStock",
				"desmoche.helpDrawDiscard",
				"desmoche.helpMeld",
				"desmoche.helpLayOff",
				"desmoche.helpDesmoche",
				"desmoche.helpDiscard",
				"desmoche.helpNext",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("zwicker",
		func() usecase.ZwickerInteractorIF {
			return usecase.NewZwickerInteractor(domain.NewDefaultZwicker(), new(presenter.ZwickerCuiPresenter))
		},
		controller.NewZwickerCuiController,
		CuiHelpSpec{
			TitleKey: "zwicker.helpTitle",
			CommandKeys: []string{
				"zwicker.helpTake",
				"zwicker.helpBuild",
				"zwicker.helpTrail",
				"zwicker.helpNext",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("poch",
		func() usecase.PochInteractorIF {
			return usecase.NewPochInteractor(domain.NewDefaultPoch(), new(presenter.PochCuiPresenter))
		},
		controller.NewPochCuiController,
		CuiHelpSpec{
			TitleKey: "poch.helpTitle",
			CommandKeys: []string{
				"poch.helpBet",
				"poch.helpFold",
				"poch.helpPlay",
				"poch.helpNext",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("popejoan",
		func() usecase.PopeJoanInteractorIF {
			return usecase.NewPopeJoanInteractor(domain.NewDefaultPopeJoan(), new(presenter.PopeJoanCuiPresenter))
		},
		controller.NewPopeJoanCuiController,
		CuiHelpSpec{
			TitleKey: "popejoan.helpTitle",
			CommandKeys: []string{
				"popejoan.helpPlay",
				"popejoan.helpNext",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("nainjaune",
		func() usecase.NainJauneInteractorIF {
			return usecase.NewNainJauneInteractor(domain.NewDefaultNainJaune(), new(presenter.NainJauneCuiPresenter))
		},
		controller.NewNainJauneCuiController,
		CuiHelpSpec{
			TitleKey: "nainjaune.helpTitle",
			CommandKeys: []string{
				"nainjaune.helpPlay",
				"nainjaune.helpNext",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("kille",
		func() usecase.KilleInteractorIF {
			return usecase.NewKilleInteractor(domain.NewDefaultKille(), new(presenter.KilleCuiPresenter))
		},
		controller.NewKilleCuiController,
		CuiHelpSpec{
			TitleKey: "kille.helpTitle",
			CommandKeys: []string{
				"kille.helpExchange",
				"kille.helpSatisfied",
				"kille.helpReenter",
				"kille.helpNextRound",
			},
			ExtraCommandLines: []string{"  l                        action log"},
			SettingKeys: []string{
				"kille.helpSetStake",
			},
		}),
	BindCuiFor("klaberjass",
		func() usecase.KlaberjassInteractorIF {
			return usecase.NewKlaberjassInteractor(domain.NewDefaultKlaberjass(), new(presenter.KlaberjassCuiPresenter))
		},
		controller.NewKlaberjassCuiController,
		CuiHelpSpec{
			TitleKey: "klaberjass.helpTitle",
			CommandKeys: []string{
				"klaberjass.helpAccept",
				"klaberjass.helpCall",
				"klaberjass.helpPass",
				"klaberjass.helpSchmeiss",
				"klaberjass.helpAnswer",
				"klaberjass.helpPlay",
				"klaberjass.helpNext",
			},
			ExtraCommandLines: []string{"  l                        action log"},
			SettingKeys: []string{
				"klaberjass.helpSetTarget",
			},
		}),
	BindCuiFor("kaiser",
		func() usecase.KaiserInteractorIF {
			return usecase.NewKaiserInteractor(domain.NewDefaultKaiser(), new(presenter.KaiserCuiPresenter))
		},
		controller.NewKaiserCuiController,
		CuiHelpSpec{
			TitleKey: "kaiser.helpTitle",
			CommandKeys: []string{
				"kaiser.helpBid",
				"kaiser.helpPass",
				"kaiser.helpTrump",
				"kaiser.helpDiscard",
				"kaiser.helpPlay",
				"kaiser.helpNext",
				"kaiser.helpHint",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("boston",
		func() usecase.BostonInteractorIF {
			return usecase.NewBostonInteractor(domain.NewDefaultBoston(), new(presenter.BostonCuiPresenter))
		},
		controller.NewBostonCuiController,
		CuiHelpSpec{
			TitleKey: "boston.helpTitle",
			CommandKeys: []string{
				"boston.helpBid",
				"boston.helpPass",
				"boston.helpCallPartner",
				"boston.helpPlay",
				"boston.helpNext",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("vint",
		func() usecase.VintInteractorIF {
			return usecase.NewVintInteractor(domain.NewDefaultVint(), new(presenter.VintCuiPresenter))
		},
		controller.NewVintCuiController,
		CuiHelpSpec{
			TitleKey: "vint.helpTitle",
			CommandKeys: []string{
				"vint.helpBid",
				"vint.helpPass",
				"vint.helpPlay",
				"vint.helpNext",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("bideuchre",
		func() usecase.BidEuchreInteractorIF {
			return usecase.NewBidEuchreInteractor(domain.NewDefaultBidEuchre(), new(presenter.BidEuchreCuiPresenter))
		},
		controller.NewBidEuchreCuiController,
		CuiHelpSpec{
			TitleKey: "bideuchre.helpTitle",
			CommandKeys: []string{
				"bideuchre.helpBid",
				"bideuchre.helpPass",
				"bideuchre.helpTrump",
				"bideuchre.helpPlay",
				"bideuchre.helpNext",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("sixbidsolo",
		func() usecase.SixBidSoloInteractorIF {
			return usecase.NewSixBidSoloInteractor(domain.NewDefaultSixBidSolo(), new(presenter.SixBidSoloCuiPresenter))
		},
		controller.NewSixBidSoloCuiController,
		CuiHelpSpec{
			TitleKey: "sixbidsolo.helpTitle",
			CommandKeys: []string{
				"sixbidsolo.helpBid",
				"sixbidsolo.helpPass",
				"sixbidsolo.helpDeclare",
				"sixbidsolo.helpPlay",
				"sixbidsolo.helpNext",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("karnoffel",
		func() usecase.KarnoffelInteractorIF {
			return usecase.NewKarnoffelInteractor(domain.NewDefaultKarnoffel(), new(presenter.KarnoffelCuiPresenter))
		},
		controller.NewKarnoffelCuiController,
		CuiHelpSpec{
			TitleKey: "karnoffel.helpTitle",
			CommandKeys: []string{
				"karnoffel.helpPlay",
				"karnoffel.helpNext",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("literature",
		func() usecase.LiteratureInteractorIF {
			return usecase.NewLiteratureInteractor(domain.NewDefaultLiterature(), new(presenter.LiteratureCuiPresenter))
		},
		controller.NewLiteratureCuiController,
		CuiHelpSpec{
			TitleKey: "literature.helpTitle",
			CommandKeys: []string{
				"literature.helpAsk",
				"literature.helpClaim",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("guandan",
		func() usecase.GuandanInteractorIF {
			return usecase.NewGuandanInteractor(domain.NewDefaultGuandan(), new(presenter.GuandanCuiPresenter))
		},
		controller.NewGuandanCuiController,
		CuiHelpSpec{
			TitleKey: "guandan.helpTitle",
			CommandKeys: []string{
				"guandan.helpPlay",
				"guandan.helpPass",
				"guandan.helpTribute",
				"guandan.helpNext",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("shengji",
		func() usecase.ShengJiInteractorIF {
			return usecase.NewShengJiInteractor(domain.NewDefaultShengJi(), new(presenter.ShengJiCuiPresenter))
		},
		controller.NewShengJiCuiController,
		CuiHelpSpec{
			TitleKey: "shengji.helpTitle",
			CommandKeys: []string{
				"shengji.helpDeclare",
				"shengji.helpBury",
				"shengji.helpPlay",
				"shengji.helpNext",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("auldlangsyne",
		func() usecase.AuldLangSyneInteractorIF {
			return usecase.NewAuldLangSyneInteractor(domain.NewDefaultAuldLangSyne(), new(presenter.AuldLangSyneCuiPresenter))
		},
		controller.NewAuldLangSyneCuiController,
		CuiHelpSpec{
			TitleKey: "auldlangsyne.helpTitle",
			CommandKeys: []string{
				"auldlangsyne.helpDeal",
				"auldlangsyne.helpWasteToFoundation",
				"auldlangsyne.helpGiveUp",
				"auldlangsyne.helpHint",
				"auldlangsyne.helpAutoComplete",
				"auldlangsyne.helpUndo",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("soko",
		func() usecase.FiveCardStudInteractorIF {
			return usecase.NewFiveCardStudInteractor(domain.NewDefaultSoko(), new(presenter.FiveCardStudCuiPresenter))
		},
		controller.NewFiveCardStudCuiController,
		CuiHelpSpec{
			TitleKey: "soko.helpTitle",
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
		}),
	BindCuiFor("fourseasons",
		func() usecase.FourSeasonsInteractorIF {
			return usecase.NewFourSeasonsInteractor(domain.NewDefaultFourSeasons(), new(presenter.FourSeasonsCuiPresenter))
		},
		controller.NewFourSeasonsCuiController,
		CuiHelpSpec{
			TitleKey: "fourseasons.helpTitle",
			CommandKeys: []string{
				"fourseasons.helpDraw",
				"fourseasons.helpMoveWaste",
				"fourseasons.helpMoveTableau",
				"fourseasons.helpGiveUp",
				"fourseasons.helpHint",
				"fourseasons.helpAutoComplete",
				"fourseasons.helpUndo",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("colorado",
		func() usecase.ColoradoInteractorIF {
			return usecase.NewColoradoInteractor(domain.NewDefaultColorado(), new(presenter.ColoradoCuiPresenter))
		},
		controller.NewColoradoCuiController,
		CuiHelpSpec{
			TitleKey: "colorado.helpTitle",
			CommandKeys: []string{
				"colorado.helpDraw",
				"colorado.helpMoveTF",
				"colorado.helpMoveWF",
				"colorado.helpMoveWT",
				"colorado.helpMoveST",
				"colorado.helpGiveUp",
				"colorado.helpHint",
				"colorado.helpAutoComplete",
				"colorado.helpUndo",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("cribbagesquares",
		func() usecase.CribbageSquaresInteractorIF {
			return usecase.NewCribbageSquaresInteractor(domain.NewDefaultCribbageSquares(), new(presenter.CribbageSquaresCuiPresenter))
		},
		controller.NewCribbageSquaresCuiController,
		CuiHelpSpec{
			TitleKey: "cribbagesquares.helpTitle",
			CommandKeys: []string{
				"cribbagesquares.helpPlace",
				"cribbagesquares.helpHint",
				"cribbagesquares.helpUndo",
				"cribbagesquares.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("diplomat",
		func() usecase.DiplomatInteractorIF {
			return usecase.NewDiplomatInteractor(domain.NewDefaultDiplomat(), new(presenter.DiplomatCuiPresenter))
		},
		controller.NewDiplomatCuiController,
		CuiHelpSpec{
			TitleKey: "diplomat.helpTitle",
			CommandKeys: []string{
				"diplomat.helpDraw",
				"diplomat.helpMoveTF",
				"diplomat.helpMoveTT",
				"diplomat.helpMoveWF",
				"diplomat.helpMoveWT",
				"diplomat.helpGiveUp",
				"diplomat.helpHint",
				"diplomat.helpAutoComplete",
				"diplomat.helpUndo",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("royalcotillion",
		func() usecase.RoyalCotillionInteractorIF {
			return usecase.NewRoyalCotillionInteractor(domain.NewDefaultRoyalCotillion(), new(presenter.RoyalCotillionCuiPresenter))
		},
		controller.NewRoyalCotillionCuiController,
		CuiHelpSpec{
			TitleKey: "royalcotillion.helpTitle",
			CommandKeys: []string{
				"royalcotillion.helpDraw",
				"royalcotillion.helpMoveTF",
				"royalcotillion.helpMoveRF",
				"royalcotillion.helpMoveWF",
				"royalcotillion.helpMoveWT",
				"royalcotillion.helpMoveST",
				"royalcotillion.helpGiveUp",
				"royalcotillion.helpHint",
				"royalcotillion.helpAutoComplete",
				"royalcotillion.helpUndo",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("crazyquilt",
		func() usecase.CrazyQuiltInteractorIF {
			return usecase.NewCrazyQuiltInteractor(domain.NewDefaultCrazyQuilt(), new(presenter.CrazyQuiltCuiPresenter))
		},
		controller.NewCrazyQuiltCuiController,
		CuiHelpSpec{
			TitleKey: "crazyquilt.helpTitle",
			CommandKeys: []string{
				"crazyquilt.helpDraw",
				"crazyquilt.helpMoveQF",
				"crazyquilt.helpMoveQW",
				"crazyquilt.helpMoveWF",
				"crazyquilt.helpGiveUp",
				"crazyquilt.helpHint",
				"crazyquilt.helpAutoComplete",
				"crazyquilt.helpUndo",
			},
			ExtraCommandLines: []string{"  l                        action log"},
		}),
	BindCuiFor("germanwhist",
		func() usecase.GermanWhistInteractorIF {
			return usecase.NewGermanWhistInteractor(domain.NewDefaultGermanWhist(), new(presenter.GermanWhistCuiPresenter))
		},
		controller.NewGermanWhistCuiController,
		CuiHelpSpec{
			TitleKey: "germanwhist.helpTitle",
			CommandKeys: []string{
				"germanwhist.helpPlay",
				"germanwhist.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("slobberhannes",
		func() usecase.SlobberhannesInteractorIF {
			return usecase.NewSlobberhannesInteractor(domain.NewDefaultSlobberhannes(), new(presenter.SlobberhannesCuiPresenter))
		},
		controller.NewSlobberhannesCuiController,
		CuiHelpSpec{
			TitleKey: "slobberhannes.helpTitle",
			CommandKeys: []string{
				"slobberhannes.helpPlay",
				"slobberhannes.helpNext",
				"slobberhannes.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("polignac",
		func() usecase.PolignacInteractorIF {
			return usecase.NewPolignacInteractor(domain.NewDefaultPolignac(), new(presenter.PolignacCuiPresenter))
		},
		controller.NewPolignacCuiController,
		CuiHelpSpec{
			TitleKey: "polignac.helpTitle",
			CommandKeys: []string{
				"polignac.helpCapot",
				"polignac.helpPass",
				"polignac.helpPlay",
				"polignac.helpNext",
				"polignac.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("reversis",
		func() usecase.ReversisInteractorIF {
			return usecase.NewReversisInteractor(domain.NewDefaultReversis(), new(presenter.ReversisCuiPresenter))
		},
		controller.NewReversisCuiController,
		CuiHelpSpec{
			TitleKey: "reversis.helpTitle",
			CommandKeys: []string{
				"reversis.helpPlay",
				"reversis.helpNext",
				"reversis.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("rams",
		func() usecase.RamsInteractorIF {
			return usecase.NewRamsInteractor(domain.NewDefaultRams(), new(presenter.RamsCuiPresenter))
		},
		controller.NewRamsCuiController,
		CuiHelpSpec{
			TitleKey: "rams.helpTitle",
			CommandKeys: []string{
				"rams.helpIn",
				"rams.helpOut",
				"rams.helpCard",
				"rams.helpNext",
				"rams.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("tarabish",
		func() usecase.TarabishInteractorIF {
			return usecase.NewTarabishInteractor(domain.NewDefaultTarabish(), new(presenter.TarabishCuiPresenter))
		},
		controller.NewTarabishCuiController,
		CuiHelpSpec{
			TitleKey: "tarabish.helpTitle",
			CommandKeys: []string{
				"tarabish.helpTake",
				"tarabish.helpPass",
				"tarabish.helpPlay",
				"tarabish.helpNext",
				"tarabish.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("baloot",
		func() usecase.BalootInteractorIF {
			return usecase.NewBalootInteractor(domain.NewDefaultBaloot(), new(presenter.BalootCuiPresenter))
		},
		controller.NewBalootCuiController,
		CuiHelpSpec{
			TitleKey: "baloot.helpTitle",
			CommandKeys: []string{
				"baloot.helpSun",
				"baloot.helpHokom",
				"baloot.helpPass",
				"baloot.helpPlay",
				"baloot.helpNext",
				"baloot.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("estimation",
		func() usecase.EstimationInteractorIF {
			return usecase.NewEstimationInteractor(domain.NewDefaultEstimation(), new(presenter.EstimationCuiPresenter))
		},
		controller.NewEstimationCuiController,
		CuiHelpSpec{
			TitleKey: "estimation.helpTitle",
			CommandKeys: []string{
				"estimation.helpTrump",
				"estimation.helpBid",
				"estimation.helpPlay",
				"estimation.helpNext",
				"estimation.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("israeliwhist",
		func() usecase.IsraeliWhistInteractorIF {
			return usecase.NewIsraeliWhistInteractor(domain.NewDefaultIsraeliWhist(), new(presenter.IsraeliWhistCuiPresenter))
		},
		controller.NewIsraeliWhistCuiController,
		CuiHelpSpec{
			TitleKey: "israeliwhist.helpTitle",
			CommandKeys: []string{
				"israeliwhist.helpAuction",
				"israeliwhist.helpPass",
				"israeliwhist.helpBid",
				"israeliwhist.helpPlay",
				"israeliwhist.helpNext",
				"israeliwhist.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("hokm",
		func() usecase.HokmInteractorIF {
			return usecase.NewHokmInteractor(domain.NewDefaultHokm(), new(presenter.HokmCuiPresenter))
		},
		controller.NewHokmCuiController,
		CuiHelpSpec{
			TitleKey: "hokm.helpTitle",
			CommandKeys: []string{
				"hokm.helpTrump",
				"hokm.helpPlay",
				"hokm.helpNext",
				"hokm.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("shelem",
		func() usecase.ShelemInteractorIF {
			return usecase.NewShelemInteractor(domain.NewDefaultShelem(), new(presenter.ShelemCuiPresenter))
		},
		controller.NewShelemCuiController,
		CuiHelpSpec{
			TitleKey: "shelem.helpTitle",
			CommandKeys: []string{
				"shelem.helpBid",
				"shelem.helpShelem",
				"shelem.helpPass",
				"shelem.helpDiscard",
				"shelem.helpPlay",
				"shelem.helpNext",
				"shelem.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("mendikot",
		func() usecase.MendikotInteractorIF {
			return usecase.NewMendikotInteractor(domain.NewDefaultMendikot(), new(presenter.MendikotCuiPresenter))
		},
		controller.NewMendikotCuiController,
		CuiHelpSpec{
			TitleKey: "mendikot.helpTitle",
			CommandKeys: []string{
				"mendikot.helpPlay",
				"mendikot.helpNext",
				"mendikot.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("bhabhi",
		func() usecase.BhabhiInteractorIF {
			return usecase.NewBhabhiInteractor(domain.NewDefaultBhabhi(), new(presenter.BhabhiCuiPresenter))
		},
		controller.NewBhabhiCuiController,
		CuiHelpSpec{
			TitleKey: "bhabhi.helpTitle",
			CommandKeys: []string{
				"bhabhi.helpPlay",
				"bhabhi.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("teendopaanch",
		func() usecase.TeenDoPaanchInteractorIF {
			return usecase.NewTeenDoPaanchInteractor(domain.NewDefaultTeenDoPaanch(), new(presenter.TeenDoPaanchCuiPresenter))
		},
		controller.NewTeenDoPaanchCuiController,
		CuiHelpSpec{
			TitleKey: "teendopaanch.helpTitle",
			CommandKeys: []string{
				"teendopaanch.helpTrump",
				"teendopaanch.helpPlay",
				"teendopaanch.helpNext",
				"teendopaanch.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("hasenpfeffer",
		func() usecase.HasenpfefferInteractorIF {
			return usecase.NewHasenpfefferInteractor(domain.NewDefaultHasenpfeffer(), new(presenter.HasenpfefferCuiPresenter))
		},
		controller.NewHasenpfefferCuiController,
		CuiHelpSpec{
			TitleKey: "hasenpfeffer.helpTitle",
			CommandKeys: []string{
				"hasenpfeffer.helpBid",
				"hasenpfeffer.helpPass",
				"hasenpfeffer.helpDiscard",
				"hasenpfeffer.helpPlay",
				"hasenpfeffer.helpNext",
				"hasenpfeffer.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("sergeantmajor",
		func() usecase.SergeantMajorInteractorIF {
			return usecase.NewSergeantMajorInteractor(domain.NewDefaultSergeantMajor(), new(presenter.SergeantMajorCuiPresenter))
		},
		controller.NewSergeantMajorCuiController,
		CuiHelpSpec{
			TitleKey: "sergeantmajor.helpTitle",
			CommandKeys: []string{
				"sergeantmajor.helpTrump",
				"sergeantmajor.helpDiscard",
				"sergeantmajor.helpPlay",
				"sergeantmajor.helpNext",
				"sergeantmajor.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("honeymoonbridge",
		func() usecase.HoneymoonBridgeInteractorIF {
			return usecase.NewHoneymoonBridgeInteractor(domain.NewDefaultHoneymoonBridge(), new(presenter.HoneymoonBridgeCuiPresenter))
		},
		controller.NewHoneymoonBridgeCuiController,
		CuiHelpSpec{
			TitleKey: "honeymoonbridge.helpTitle",
			CommandKeys: []string{
				"honeymoonbridge.helpBid",
				"honeymoonbridge.helpPass",
				"honeymoonbridge.helpPlay",
				"honeymoonbridge.helpNext",
				"honeymoonbridge.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("minibridge",
		func() usecase.MinibridgeInteractorIF {
			return usecase.NewMinibridgeInteractor(domain.NewDefaultMinibridge(), new(presenter.MinibridgeCuiPresenter))
		},
		controller.NewMinibridgeCuiController,
		CuiHelpSpec{
			TitleKey: "minibridge.helpTitle",
			CommandKeys: []string{
				"minibridge.helpContract",
				"minibridge.helpPlay",
				"minibridge.helpNext",
				"minibridge.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("pasur",
		func() usecase.PasurInteractorIF {
			return usecase.NewPasurInteractor(domain.NewDefaultPasur(), new(presenter.PasurCuiPresenter))
		},
		controller.NewPasurCuiController,
		CuiHelpSpec{
			TitleKey: "pasur.helpTitle",
			CommandKeys: []string{
				"pasur.helpPlay",
				"pasur.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("snap",
		func() usecase.SnapInteractorIF {
			return usecase.NewSnapInteractor(domain.NewDefaultSnap(), new(presenter.SnapCuiPresenter))
		},
		controller.NewSnapCuiController,
		CuiHelpSpec{
			TitleKey: "snap.helpTitle",
			CommandKeys: []string{
				"snap.helpStep",
				"snap.helpSnap",
				"snap.helpTick",
				"snap.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("rollingstone",
		func() usecase.RollingStoneInteractorIF {
			return usecase.NewRollingStoneInteractor(domain.NewDefaultRollingStone(), new(presenter.RollingStoneCuiPresenter))
		},
		controller.NewRollingStoneCuiController,
		CuiHelpSpec{
			TitleKey: "rollingstone.helpTitle",
			CommandKeys: []string{
				"rollingstone.helpPlay",
				"rollingstone.helpPickUp",
				"rollingstone.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("lingerlonger",
		func() usecase.LingerLongerInteractorIF {
			return usecase.NewLingerLongerInteractor(domain.NewDefaultLingerLonger(), new(presenter.LingerLongerCuiPresenter))
		},
		controller.NewLingerLongerCuiController,
		CuiHelpSpec{
			TitleKey: "lingerlonger.helpTitle",
			CommandKeys: []string{
				"lingerlonger.helpPlay",
				"lingerlonger.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("pig",
		func() usecase.PigInteractorIF {
			return usecase.NewPigInteractor(domain.NewDefaultPig(), new(presenter.PigCuiPresenter))
		},
		controller.NewPigCuiController,
		CuiHelpSpec{
			TitleKey: "pig.helpTitle",
			CommandKeys: []string{
				"pig.helpPass",
				"pig.helpSignal",
				"pig.helpNext",
				"pig.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("stealingbundles",
		func() usecase.StealingBundlesInteractorIF {
			return usecase.NewStealingBundlesInteractor(domain.NewDefaultStealingBundles(), new(presenter.StealingBundlesCuiPresenter))
		},
		controller.NewStealingBundlesCuiController,
		CuiHelpSpec{
			TitleKey: "stealingbundles.helpTitle",
			CommandKeys: []string{
				"stealingbundles.helpTake",
				"stealingbundles.helpSteal",
				"stealingbundles.helpTrail",
				"stealingbundles.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("cucumber",
		func() usecase.CucumberInteractorIF {
			return usecase.NewCucumberInteractor(domain.NewDefaultCucumber(), new(presenter.CucumberCuiPresenter))
		},
		controller.NewCucumberCuiController,
		CuiHelpSpec{
			TitleKey: "cucumber.helpTitle",
			CommandKeys: []string{
				"cucumber.helpPlay",
				"cucumber.helpNext",
				"cucumber.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("goofspiel",
		func() usecase.GoofspielInteractorIF {
			return usecase.NewGoofspielInteractor(domain.NewDefaultGoofspiel(), new(presenter.GoofspielCuiPresenter))
		},
		controller.NewGoofspielCuiController,
		CuiHelpSpec{
			TitleKey: "goofspiel.helpTitle",
			CommandKeys: []string{
				"goofspiel.helpBid",
				"goofspiel.helpNext",
				"goofspiel.helpGiveUp",
			},
			ExtraCommandLines: []string{"  l                    action log"},
		}),
	BindCuiFor("andarbahar",
		func() usecase.AndarBaharInteractorIF {
			return usecase.NewAndarBaharInteractor(domain.NewDefaultAndarBahar(), new(presenter.AndarBaharCuiPresenter))
		},
		controller.NewAndarBaharCuiController,
		CuiHelpSpec{
			TitleKey: "andarbahar.helpTitle",
			CommandKeys: []string{
				"andarbahar.helpBet",
				"andarbahar.helpClear",
				"andarbahar.helpHint",
			},
			ExtraCommandLines: []string{"  log                  action log"},
		}),
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
