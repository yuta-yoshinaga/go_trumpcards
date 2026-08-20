package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// gameEndLogger is a minimal interface satisfied by all game types that support action logs.
type gameEndLogger interface {
	GetGameEndFlag() bool
	GetActionLog() []*domain.ActionLogEntry
}

// actionLogOutputText returns the action log as plain text, or an empty log if the game is not finished.
func actionLogOutputText(game gameEndLogger) string {
	if !game.GetGameEndFlag() {
		return actionLogToText(nil)
	}
	return actionLogToText(game.GetActionLog())
}

// actionLogOutputTextWithNames is actionLogOutputText with seat names resolved
// by the caller, so the transcript calls the seats what the rest of the screen
// calls them (#5977).
func actionLogOutputTextWithNames(game gameEndLogger, nameOf func(idx int) string) string {
	if !game.GetGameEndFlag() {
		return actionLogToTextWithNames(nil, nameOf)
	}
	return actionLogToTextWithNames(game.GetActionLog(), nameOf)
}

// seatedGame is a game that can name the seat behind a log entry's PlayerIdx.
type seatedGame[P cuiPlayer] interface {
	gameEndLogger
	GetPlayer(int) P
}

// actionLogOutputTextForSeats renders the transcript naming seats the way the
// rest of the screen names them.
//
// **クロージャはここに 1 つだけ置く。**呼び出し側 86 ファイルにインラインで
// 書くと、席名を引く行が「棋譜が空の局面しか見ていないテスト」では通らず、
// 全ファイルが部分的に未到達として並ぶ。
//
// P は明示する (`actionLogOutputTextForSeats[*domain.XPlayer](g)`)。
// メソッドの戻り値からの型推論は Go には無い。
func actionLogOutputTextForSeats[P cuiPlayer, G seatedGame[P]](game G) string {
	return actionLogOutputTextWithNames(game, func(idx int) string { return cuiPlayerName(game.GetPlayer(idx), idx) })
}

// actionLogOutputJSON returns the action log as JSON, or an empty log if the game is not finished.
func actionLogOutputJSON(game gameEndLogger) string {
	if !game.GetGameEndFlag() {
		return actionLogToJSON(nil)
	}
	return actionLogToJSON(game.GetActionLog())
}

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

// actionLogToText 棋譜をテキスト形式に変換する。
//
// 座席名は既定で "Player 0" 相当（Web GUI と同じ表記）。ゲーム側が席の一覧を
// 持っているなら actionLogToTextWithNames で cuiPlayerName に寄せられる。
func actionLogToText(entries []*domain.ActionLogEntry) string {
	return actionLogToTextWithNames(entries, nil)
}

// actionLogToTextWithNames は座席名の解決を呼び出し側に委ねる版。
//
// **棋譜だけ呼称が食い違っていた。**同じ画面の他の行は cuiPlayerName 経由で
// 「あなた」「CPU 1」と出すのに、棋譜は英語固定の "Player 0" だった (#5977)。
// nameOf が nil か空文字を返した席は既定表記に落とす。
func actionLogToTextWithNames(entries []*domain.ActionLogEntry, nameOf func(idx int) string) string {
	if len(entries) == 0 {
		return i18n.T("cuiActionLogEmpty") + "\n"
	}
	var sb strings.Builder
	sb.WriteString(i18n.T("cuiActionLogHeader") + "\n")
	for _, e := range entries {
		player := i18n.T("cuiActionLogSystem")
		if e.PlayerIdx >= 0 {
			player = ""
			if nameOf != nil {
				player = nameOf(e.PlayerIdx)
			}
			if player == "" {
				player = i18n.Tf("cuiActionLogPlayer", "idx", strconv.Itoa(e.PlayerIdx))
			}
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
