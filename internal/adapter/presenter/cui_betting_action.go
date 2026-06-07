//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// cui_betting_action.go holds the betting-action display helper. It references
// the casino-only domain.PokerAction* constants and is used only by casino
// (betting) CUI presenters, so it carries the casino build tag — keeping the
// shared cui_card_helper.go free of casino dependencies (#2126).

// bettingActionKeys maps betting action constants to i18n keys in
// cui_common.json. Used by cuiBettingActionName so the displayed action
// text follows the active locale (issue #1699 Phase 1).
var bettingActionKeys = map[int]string{
	domain.PokerActionFold:  "cuiBettingActionFold",
	domain.PokerActionCheck: "cuiBettingActionCheck",
	domain.PokerActionCall:  "cuiBettingActionCall",
	domain.PokerActionBet:   "cuiBettingActionBet",
	domain.PokerActionRaise: "cuiBettingActionRaise",
	domain.PokerActionAllIn: "cuiBettingActionAllIn",
}

// cuiBettingActionName returns the localized name for a betting action.
func cuiBettingActionName(action int) string {
	if key, ok := bettingActionKeys[action]; ok {
		return i18n.T(key)
	}
	return i18n.T("cuiBettingActionUnknown")
}
