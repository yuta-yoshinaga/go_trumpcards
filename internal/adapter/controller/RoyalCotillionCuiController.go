//go:build !js || !wasm || classic

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RoyalCotillionCuiController ロイヤルコティヨン CUI コントローラークラス
type RoyalCotillionCuiController struct {
	ci usecase.RoyalCotillionInteractorIF
}

// NewRoyalCotillionCuiController コンストラクタ
func NewRoyalCotillionCuiController(ci usecase.RoyalCotillionInteractorIF) *RoyalCotillionCuiController {
	return &RoyalCotillionCuiController{ci: ci}
}

// Exec コマンド実行
func (c *RoyalCotillionCuiController) Exec(command string) string {
	return execSolitaireCui(command, solitaireCuiFns{
		reset:        c.ci.Reset,
		move:         c.handleMove,
		giveUp:       c.ci.GiveUp,
		autoComplete: c.ci.AutoComplete,
		undo:         c.ci.Undo,
		hint:         c.ci.Hint,
		actionLog:    c.ci.ActionLog,
		extraCommands: map[string]func([]string) string{
			"d": func([]string) string { return c.ci.Draw() }, "draw": func([]string) string { return c.ci.Draw() },
		},
	})
}

// handleMove 移動コマンドを処理。supported syntax:
//
//	m t <slot>          - a tableau card to a foundation (its only move)
//	m r <pile>          - a reserve top to a foundation (its only move)
//	m w f               - the waste top to a foundation
//	m w t <slot>        - the waste top into an empty tableau slot
//	m s t <slot>        - the stock top straight into an empty tableau slot
func (c *RoyalCotillionCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("royalcotillion.promptSourceZone"), "m {0}")
	}
	switch args[0] {
	case "t":
		return c.handleMoveFromTableau(args[1:])
	case "r":
		return c.handleMoveFromReserve(args[1:])
	case "w":
		return c.handleMoveFromWaste(args[1:])
	case "s":
		return c.handleMoveFromStock(args[1:])
	default:
		return invalidArg("royalcotillion.invalidFromZone", "val", args[0])
	}
}

// handleMoveFromTableau タブロー枠の札は基礎札へしか送れないので行き先を尋ねない。
// 「m t 3 f」と打たれても同じ手になるよう、末尾の f は受け流す。
func (c *RoyalCotillionCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("royalcotillion.promptFromPile"), "m t {0}")
	}
	slot, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args, "", "royalcotillion.invalidPile", cuiutil.NoMin, cuiutil.NoMax)
	if !parseOK {
		return parseMsg
	}
	if len(args) >= 2 && args[1] != "f" {
		return invalidArg("royalcotillion.invalidToZone", "val", args[1])
	}
	return c.ci.MoveTableauToFoundation(slot)
}

// handleMoveFromReserve リザーブも基礎札へしか送れない。
func (c *RoyalCotillionCuiController) handleMoveFromReserve(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("royalcotillion.promptFromReserve"), "m r {0}")
	}
	pile, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args, "", "royalcotillion.invalidReserve", cuiutil.NoMin, cuiutil.NoMax)
	if !parseOK {
		return parseMsg
	}
	if len(args) >= 2 && args[1] != "f" {
		return invalidArg("royalcotillion.invalidToZone", "val", args[1])
	}
	return c.ci.MoveReserveToFoundation(pile)
}

func (c *RoyalCotillionCuiController) handleMoveFromWaste(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("royalcotillion.promptToZone"), "m w {0}")
	}
	switch args[0] {
	case "f":
		return c.ci.MoveWasteToFoundation()
	case "t":
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("royalcotillion.promptToPile"), "m w t {0}")
		}
		pile, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[1:], "", "royalcotillion.invalidPile", cuiutil.NoMin, cuiutil.NoMax)
		if !parseOK {
			return parseMsg
		}
		return c.ci.MoveWasteToTableau(pile)
	default:
		return invalidArg("royalcotillion.invalidToZone", "val", args[0])
	}
}

// handleMoveFromStock 山札からは空き山を埋める手しかない。基礎札へ直接は送れない。
func (c *RoyalCotillionCuiController) handleMoveFromStock(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("royalcotillion.promptToZone"), "m s {0}")
	}
	if args[0] != "t" {
		return invalidArg("royalcotillion.invalidToZone", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("royalcotillion.promptToPile"), "m s t {0}")
	}
	pile, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[1:], "", "royalcotillion.invalidPile", cuiutil.NoMin, cuiutil.NoMax)
	if !parseOK {
		return parseMsg
	}
	return c.ci.MoveStockToTableau(pile)
}
