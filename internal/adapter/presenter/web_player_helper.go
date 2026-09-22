package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// webPlayerName returns the localized seat label for a web response.
// It mirrors cuiPlayerName but omits the ANSI styling, because web
// responses are consumed by the browser rather than a terminal.
func webPlayerName(isHuman bool, idx int) string {
	if isHuman {
		return i18n.T("cuiPlayerYou")
	}
	return i18n.Tf("cuiPlayerCpu", "idx", strconv.Itoa(idx))
}
