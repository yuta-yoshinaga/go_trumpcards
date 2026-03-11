package controller

import (
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// unknownCommandMessage は不明なコマンドに対する統一エラーメッセージを返す。
func unknownCommandMessage(command string) string {
	return i18n.Tf("unknownCommand", "cmd", command)
}

// execCuiCommand は全CUIコントローラーで共通のコマンド解析を行うヘルパー関数。
// command を strings.Fields で分割し、空入力・q/quit・r/reset を共通処理する。
// ゲーム固有コマンドは gameHandler で処理し、未知コマンドは unknownMsg で応答する。
func execCuiCommand(
	command string,
	resetFn func(args []string) string,
	unknownMsg func(string) string,
	gameHandler func(cmd string, args []string) (string, bool),
) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return unknownMsg(command)
	}
	switch fields[0] {
	case "q", "quit":
		return "bye."
	case "r", "reset":
		return resetFn(fields[1:])
	default:
		if result, handled := gameHandler(fields[0], fields[1:]); handled {
			return result
		}
		return unknownMsg(command)
	}
}
