package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// pokerActionInteractor is the common interface for poker-family game interactors.
// All five poker games (Holdem, Omaha, ShortDeck, Pineapple, SevenCardStud) satisfy this.
type pokerActionInteractor interface {
	Action(action int, amount int, humanPlayMs int) string
	Rebuy() string
	SkipRebuy() string
	Addon() string
	SkipAddon() string
	Muck() string
	ShowHand() string
	ActionLog() string
}

// pokerAction describes a command-to-action mapping.
type pokerAction struct {
	actionType int
	usesAmount bool
}

// pokerActionMap maps command strings to action types for all poker-family games.
// All games share the same int values for fold/check/call/bet/raise/allin.
var pokerActionMap = map[string]pokerAction{
	"f":     {domain.HoldemActionFold, false},
	"fold":  {domain.HoldemActionFold, false},
	"ck":    {domain.HoldemActionCheck, false},
	"check": {domain.HoldemActionCheck, false},
	"c":     {domain.HoldemActionCall, false},
	"call":  {domain.HoldemActionCall, false},
	"a":     {domain.HoldemActionAllIn, false},
	"allin": {domain.HoldemActionAllIn, false},
	"b":     {domain.HoldemActionBet, true},
	"bet":   {domain.HoldemActionBet, true},
	"ra":    {domain.HoldemActionRaise, true},
	"raise": {domain.HoldemActionRaise, true},
}

// pokerTournamentMap maps tournament-related commands to methods.
var pokerTournamentMap = map[string]func(pokerActionInteractor) string{
	"rb":        func(i pokerActionInteractor) string { return i.Rebuy() },
	"rebuy":     func(i pokerActionInteractor) string { return i.Rebuy() },
	"sr":        func(i pokerActionInteractor) string { return i.SkipRebuy() },
	"skiprebuy": func(i pokerActionInteractor) string { return i.SkipRebuy() },
	"ad":        func(i pokerActionInteractor) string { return i.Addon() },
	"addon":     func(i pokerActionInteractor) string { return i.Addon() },
	"sa":        func(i pokerActionInteractor) string { return i.SkipAddon() },
	"skipaddon": func(i pokerActionInteractor) string { return i.SkipAddon() },
	"m":         func(i pokerActionInteractor) string { return i.Muck() },
	"muck":      func(i pokerActionInteractor) string { return i.Muck() },
	"sh":        func(i pokerActionInteractor) string { return i.ShowHand() },
	"show":      func(i pokerActionInteractor) string { return i.ShowHand() },
}

// dispatchPokerAction handles commands common to all poker-family games:
// action commands (fold/check/call/bet/raise/allin) and tournament commands
// (rebuy/addon/muck/show). Returns true if the command was handled.
func dispatchPokerAction(bc *baseController, w http.ResponseWriter, interactor pokerActionInteractor, command string, amount int, humanPlayMs int) bool {
	if pa, ok := pokerActionMap[command]; ok {
		amt := 0
		if pa.usesAmount {
			amt = amount
		}
		bc.writePresenterResponse(w, interactor.Action(pa.actionType, amt, humanPlayMs))
		return true
	}
	if fn, ok := pokerTournamentMap[command]; ok {
		bc.writePresenterResponse(w, fn(interactor))
		return true
	}
	return false
}
