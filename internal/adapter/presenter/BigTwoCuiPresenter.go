package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// bigTwoPlayerStr returns the display string for a single Big Two player.
func bigTwoPlayerStr(player *domain.BigTwoPlayer, i int) string {
	var b strings.Builder
	b.WriteString(cuiPlayerName(player, i))
	if player.GetIsFinished() {
		b.WriteString(i18n.Tf("bigtwo.playerFinished", "rank", strconv.Itoa(player.GetRank())) + "\n")
	} else {
		b.WriteString(i18n.Tf("bigtwo.playerCardCount", "count", strconv.Itoa(player.GetCardsSize())) + "\n")
		if player.GetIsHuman() {
			b.WriteString(cuiIndexedCardListStr(player) + "\n")
		}
	}
	return b.String()
}

// BigTwoCuiPresenter renders the Big Two CUI view.
type BigTwoCuiPresenter struct{}

// bigTwoPlayTypeName returns the localized name of a play type, or "" for
// BigTwoPlayInvalid (nothing on the table). Mirrors the web GUI's
// `bigTwoPlayTypeKey`, which keys the same eight names off the same enum.
func bigTwoPlayTypeName(t domain.BigTwoPlayType) string {
	keys := map[domain.BigTwoPlayType]string{
		domain.BigTwoPlaySingle:        "bigtwo.playTypeSingle",
		domain.BigTwoPlayPair:          "bigtwo.playTypePair",
		domain.BigTwoPlayTriple:        "bigtwo.playTypeTriple",
		domain.BigTwoPlayStraight:      "bigtwo.playTypeStraight",
		domain.BigTwoPlayFlush:         "bigtwo.playTypeFlush",
		domain.BigTwoPlayFullHouse:     "bigtwo.playTypeFullHouse",
		domain.BigTwoPlayFourOfAKind:   "bigtwo.playTypeFourOfAKind",
		domain.BigTwoPlayStraightFlush: "bigtwo.playTypeStraightFlush",
	}
	key, ok := keys[t]
	if !ok {
		return ""
	}
	return i18n.T(key)
}

// Output renders the current game state for the active locale.
func (p *BigTwoCuiPresenter) Output(bg interfaces.BigTwoGame, lastErr error) string {
	return buildCuiOutput(i18n.T("bigtwo.helpTitle"), func(b *strings.Builder) {
		for i := 0; i < bg.GetPlayerCnt(); i++ {
			b.WriteString(bigTwoPlayerStr(bg.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		if bg.GetTableCards() != nil {
			// **役名を添える。**Web は `bt-table-playtype` バッジで場の役を常時出して
			// いるのに、CUI は生のカード列だけで、何を出せば通るのかを自分で
			// 読み取るしかなかった (#4859)。
			line := i18n.T("bigtwo.tableCards") + ": " + cuiCardSliceStr(bg.GetTableCards())
			if name := bigTwoPlayTypeName(bg.GetTablePlayType()); name != "" {
				line += " " + i18n.Tf("bigtwo.tablePlayType", "type", name)
			}
			b.WriteString(line + "\n")
		} else {
			b.WriteString(i18n.T("bigtwo.tableEmpty") + "\n")
		}

		if bg.GetGameEndFlag() {
			b.WriteString(i18n.T("bigtwo.gameEnd") + "\n")
			for i := 0; i < bg.GetPlayerCnt(); i++ {
				player := bg.GetPlayer(i)
				if player.GetRank() >= 1 {
					name := cuiPlayerName(player, i)
					fmt.Fprintf(b, "  %s: %s\n", name, i18n.Tf("bigtwo.rankN", "rank", strconv.Itoa(player.GetRank())))
				}
			}
		} else {
			for _, action := range bg.GetCpuActions() {
				var actionStr string
				if action.PlayedCards == nil {
					actionStr = i18n.T("bigtwo.actionPass")
				} else {
					actionStr = cuiCardSliceStr(action.PlayedCards)
				}
				fmt.Fprintf(b, "%s: %s\n", cuiPlayerName(bg.GetPlayer(action.PlayerIdx), action.PlayerIdx), actionStr)
			}
			if bg.IsHumanTurn() {
				b.WriteString(i18n.T("bigtwo.yourTurn") + "\n")
			}
		}

		if lastErr != nil {
			b.WriteString(lastErr.Error() + "\n")
		}
	})
}

// ActionLogOutput 棋譜を出力
func (p *BigTwoCuiPresenter) ActionLogOutput(bg interfaces.BigTwoGame) string {
	return actionLogOutputText(bg)
}
