package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// actionLogToJSON 棋譜をJSON文字列に変換する
func actionLogToJSON(entries []*domain.ActionLogEntry) string {
	out := &controller.ActionLogWebOutput{
		Entries: make([]*controller.ActionLogWebEntry, len(entries)),
	}
	for i, e := range entries {
		out.Entries[i] = &controller.ActionLogWebEntry{
			TurnNumber: e.TurnNumber,
			PlayerIdx:  e.PlayerIdx,
			ActionType: e.ActionType,
			Detail:     e.Detail,
			Cards:      cardsToOutput(e.Cards),
		}
	}
	return marshalOrError(out)
}

// actionLogToText 棋譜をテキスト形式に変換する
func actionLogToText(entries []*domain.ActionLogEntry) string {
	if len(entries) == 0 {
		return "棋譜はありません。\n"
	}
	var sb strings.Builder
	sb.WriteString("========== 棋譜 ==========\n")
	for _, e := range entries {
		player := "SYSTEM"
		if e.PlayerIdx >= 0 {
			player = fmt.Sprintf("Player %d", e.PlayerIdx)
		}
		fmt.Fprintf(&sb, "T%d [%s] %s: %s", e.TurnNumber, player, e.ActionType, e.Detail)
		if len(e.Cards) > 0 {
			cardStrs := make([]string, len(e.Cards))
			for i, c := range e.Cards {
				cardStrs[i] = cuiCardStr(c)
			}
			fmt.Fprintf(&sb, " [%s]", strings.Join(cardStrs, ", "))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
