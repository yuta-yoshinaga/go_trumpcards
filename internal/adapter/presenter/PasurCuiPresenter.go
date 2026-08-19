//go:build !js || !wasm || extra

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// pasurPlayerStr returns the display string for a single player.
func pasurPlayerStr(s interfaces.PasurGame, idx int, current bool) string {
	player := s.GetPlayer(idx)
	var b strings.Builder
	marker := " "
	if current {
		marker = ">"
	}
	role := ""
	if idx == s.GetLastCaptureIdx() {
		role = i18n.T("pasur.roleLastCapture")
	}
	b.WriteString(marker + i18n.Tf("pasur.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"captured", strconv.Itoa(player.GetCapturedCount()),
		"soors", strconv.Itoa(player.GetSoors()),
		"score", strconv.Itoa(s.GetScore(idx)),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// PasurCuiPresenter renders the Pasur CUI view.
type PasurCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *PasurCuiPresenter) Output(s interfaces.PasurGame, lastErr error) string {
	return buildCuiOutput(i18n.T("pasur.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("pasur.header",
			"pack", strconv.Itoa(s.GetPacksDealt()),
			"deck", strconv.Itoa(s.GetDeckRemaining())) + "\n")
		// **11 の合計と絵札の扱いがこのゲームの規則そのもの。** 毎回書く。
		sb.WriteString(i18n.T("pasur.rule") + "\n")

		// **場の札に番号を振る。** `play <i> <t...>` の t はこの番号。
		table := s.GetTableCards()
		if len(table) == 0 {
			sb.WriteString(i18n.T("pasur.tableEmpty") + "\n")
		} else {
			parts := make([]string, 0, len(table))
			for i, c := range table {
				parts = append(parts, "["+strconv.Itoa(i)+"]"+cuiCardStr(c))
			}
			sb.WriteString(i18n.Tf("pasur.tableLine", "cards", strings.Join(parts, " ")) + "\n")
		}

		for i := 0; i < s.GetPlayerCnt(); i++ {
			sb.WriteString(pasurPlayerStr(s, i, i == s.GetCurrentPlayerIdx() && !s.GetGameEndFlag()))
		}

		sb.WriteString("----------\n")

		cuiErrorBlock(sb, lastErr)

		if s.GetGameEndFlag() {
			winners := s.GetWinners()
			var banner string
			switch {
			case len(winners) > 1:
				banner = i18n.Tf("pasur.gameEndTie", "n", strconv.Itoa(len(winners)))
			case len(winners) == 1 && winners[0] == 0:
				banner = i18n.T("pasur.gameEndYou")
			case len(winners) == 1:
				banner = i18n.Tf("pasur.gameEndCpu",
					"name", cuiPlayerName(s.GetPlayer(winners[0]), winners[0]))
			default:
				banner = i18n.Tf("pasur.gameEndTie", "n", "0")
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		currentIdx := s.GetCurrentPlayerIdx()
		sb.WriteString(i18n.Tf("pasur.promptCurrentPlayer",
			"name", cuiPlayerName(s.GetPlayer(currentIdx), currentIdx)) + "\n")
		// **スールは「取った結果、場が空になる」こと** (#5762)。倍化を狙うなら
		// 場の枚数と取る枚数を突き合わせる必要があるので、その基準を出す。
		if n := len(s.GetTableCards()); n > 0 {
			sb.WriteString(i18n.Tf("pasur.soorNote",
				"n", strconv.Itoa(n),
				"mult", strconv.Itoa(domain.PasurSoorMultiplier)) + "\n")
		}
		sb.WriteString(i18n.T("pasur.promptPlay") + "\n")
	})
}

// HintOutput emits the current hint.
func (p *PasurCuiPresenter) HintOutput(s interfaces.PasurGame) string {
	hint := s.GetHint()
	if hint == nil || hint.CardIndex == nil {
		return i18n.T("pasur.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, pasurHintReasonKeys)
	card := s.GetPlayer(s.GetCurrentPlayerIdx()).GetCard(*hint.CardIndex)
	if len(hint.TableIndices) == 0 {
		// **取れないときは置くことを勧める。** 取る札を書かない。
		return color.Yellow(i18n.Tf("pasur.hintTrail",
			"idx", strconv.Itoa(*hint.CardIndex),
			"card", cuiCardStr(card),
			"reason", reason)) + "\n"
	}
	nums := make([]string, 0, len(hint.TableIndices))
	for _, i := range hint.TableIndices {
		nums = append(nums, strconv.Itoa(i))
	}
	line := i18n.Tf("pasur.hintCapture",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"table", strings.Join(nums, " "),
		"reason", reason)
	// 勧めている取り方が場を空にするなら、そこも言う。
	if len(hint.TableIndices) == len(s.GetTableCards()) {
		line += i18n.Tf("pasur.hintSoorMark", "mult", strconv.Itoa(domain.PasurSoorMultiplier))
	}
	return color.Yellow(line) + "\n"
}

// pasurHintReasonKeys maps hint-reason identifiers to their i18n keys.
var pasurHintReasonKeys = map[string]string{
	"pasurSoor":    "pasur.hintReasonSoor",
	"pasurCapture": "pasur.hintReasonCapture",
	"pasurTrail":   "pasur.hintReasonTrail",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *PasurCuiPresenter) ActionLogOutput(s interfaces.PasurGame) string {
	return actionLogOutputText(s)
}
