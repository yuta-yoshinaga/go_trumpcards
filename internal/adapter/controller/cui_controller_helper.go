package controller

import (
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// suggestMaxDistance はコマンド提案のための最大編集距離。
const suggestMaxDistance = 2

// unknownCommandMessage は不明なコマンドに対する統一エラーメッセージを返す。
// validCommands が指定されている場合、Levenshtein距離に基づく提案を含める。
func unknownCommandMessage(command string, validCommands []string) string {
	if suggestion := cuiutil.SuggestCommand(command, validCommands, suggestMaxDistance); suggestion != "" {
		return i18n.Tf("unknownCommandWithSuggestion", "cmd", command, "suggestion", suggestion)
	}
	return i18n.Tf("unknownCommand", "cmd", command)
}

// commonCommands は全CUIコントローラーで共通のコマンド。
var commonCommands = []string{"q", "quit", "exit", "r", "reset", "help", "?"}

// handleCuiLog は "log"/"l" コマンドを処理する。処理した場合 true を返す。
func handleCuiLog(cmd string, actionLogFn func() string) (string, bool) {
	switch cmd {
	case "log", "l":
		return actionLogFn(), true
	}
	return "", false
}

// handleCuiHintAndLog は "h"/"hint" と "log"/"l" コマンドを処理する。処理した場合 true を返す。
func handleCuiHintAndLog(cmd string, hintFn, actionLogFn func() string) (string, bool) {
	switch cmd {
	case "h", "hint":
		return hintFn(), true
	case "log", "l":
		return actionLogFn(), true
	}
	return "", false
}

// execCuiCommand は全CUIコントローラーで共通のコマンド解析を行うヘルパー関数。
// command を strings.Fields で分割し、空入力・q/quit・r/reset を共通処理する。
// ゲーム固有コマンドは gameHandler で処理し、未知コマンドは validCommands に基づく提案付きメッセージで応答する。
func execCuiCommand(
	command string,
	resetFn func(args []string) string,
	validCommands []string,
	gameHandler func(cmd string, args []string) (string, bool),
) string {
	allCommands := append(commonCommands, validCommands...)
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return i18n.T("emptyInputHint")
	}
	switch fields[0] {
	case "q", "quit", "exit":
		return i18n.QuitSentinel
	case "r", "reset":
		return resetFn(fields[1:])
	default:
		if result, handled := gameHandler(fields[0], fields[1:]); handled {
			return result
		}
		return unknownCommandMessage(fields[0], allCommands)
	}
}
