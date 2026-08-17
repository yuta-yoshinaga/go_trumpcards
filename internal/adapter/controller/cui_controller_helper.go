package controller

import (
	"sort"
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
		return i18n.MarkError(i18n.Tf("unknownCommandWithSuggestion", "cmd", command, "suggestion", suggestion))
	}
	return i18n.MarkError(i18n.Tf("unknownCommand", "cmd", command))
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

// solitaireCuiFns is the interactor surface of the tableau-solitaire CUI
// command set, taken as function values because each game has its own
// interactor type with no shared interface.
type solitaireCuiFns struct {
	reset        func() string
	move         func(args []string) string
	giveUp       func() string
	autoComplete func() string
	undo         func() string
	hint         func() string
	actionLog    func() string
	// extraCommands はゲーム固有の追加コマンド (名前 → ハンドラ)。エイリアスは
	// 別のキーとして登録する。
	//
	// **共有の一覧に直接足すと 6 ゲーム全部に生える。**受け取る側の能力
	// (例: LegalTargets) はゲームごとに違うので、名前だけ生えて動かない
	// コマンドができてしまう (#5581)。
	extraCommands map[string]func(args []string) string
}

// execSolitaireCui runs the move/giveup/autocomplete/undo command set shared by
// the tableau solitaires, delegating quit/reset and unknown-command suggestions
// to execCuiCommand.
//
// Consolidates 6 byte-identical Exec bodies: BakersDozen, BeleagueredCastle,
// Bisley, FlowerGarden, KingAlbert, StreetsAndAlleys. They differed only in
// receiver name and in which interactor the closures called — see issue #5368.
//
// The command list is passed to execCuiCommand unchanged, so a typo still
// reaches the shared suggestion path ("もしかして 'move' ですか？") instead of
// being swallowed here.
// solitaireCuiCommandNames は候補一覧を組む。追加コマンドも入れる ── 入れないと
// 打ち間違いが「もしかして」で拾われず、存在しないコマンド扱いになる。
func solitaireCuiCommandNames(fns solitaireCuiFns) []string {
	names := []string{"m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"}
	extra := make([]string, 0, len(fns.extraCommands))
	for name := range fns.extraCommands {
		extra = append(extra, name)
	}
	// map の反復順は毎回変わる。候補の並びが実行ごとに変わらないよう揃える。
	sort.Strings(extra)
	return append(names, extra...)
}

func execSolitaireCui(command string, fns solitaireCuiFns) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return fns.reset() },
		solitaireCuiCommandNames(fns),
		func(cmd string, args []string) (string, bool) {
			if h, ok := fns.extraCommands[cmd]; ok {
				return h(args), true
			}
			switch cmd {
			case "m", "move":
				return fns.move(args), true
			case "g", "giveup":
				return fns.giveUp(), true
			case "ac", "autocomplete":
				return fns.autoComplete(), true
			case "u", "undo":
				return fns.undo(), true
			default:
				return handleCuiHintAndLog(cmd, fns.hint, fns.actionLog)
			}
		},
	)
}
