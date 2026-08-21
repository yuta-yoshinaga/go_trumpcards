//go:build !js || !wasm || extra4

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PerseveranceCuiController パーシビアランスCUIコントローラークラス
type PerseveranceCuiController struct {
	bi usecase.PerseveranceInteractorIF
}

// NewPerseveranceCuiController コンストラクタ
func NewPerseveranceCuiController(bi usecase.PerseveranceInteractorIF) *PerseveranceCuiController {
	return &PerseveranceCuiController{bi: bi}
}

// Exec コマンド実行
func (c *PerseveranceCuiController) Exec(command string) string {
	return execSolitaireCui(command, solitaireCuiFns{
		reset:        c.bi.Reset,
		move:         c.handleMove,
		giveUp:       c.bi.GiveUp,
		autoComplete: c.bi.AutoComplete,
		undo:         c.bi.Undo,
		hint:         c.bi.Hint,
		actionLog:    c.bi.ActionLog,
		// **共有の一覧には足さない。**LegalTargets を持たない 5 ゲームにも
		// 名前だけ生えてしまうので、このゲームの追加コマンドとして登録する。
		extraCommands: map[string]func([]string) string{
			"t":       c.handleTargets,
			"targets": c.handleTargets,
			"rd":      func([]string) string { return c.bi.Redeal() },
			"redeal":  func([]string) string { return c.bi.Redeal() },
		},
	})
}

// handleTargets は `t <col>` / `targets <col>` を処理する。
//
// 12 列 + 4 組札を押して試すのは現実的でない (#5581)。列番号だけを取り、
// 置ける先はサーバの判定 (LegalTargets) がそのまま答える。
func (c *PerseveranceCuiController) handleTargets(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "t {0}")
	}
	col, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}
	return c.bi.Targets(col)
}

// handleMove 移動コマンドを処理
// Perseverance has no waste/stock; supported syntax:
//
//	m <fromCol> <toCol>        - move top card between tableau columns
//	m <fromCol> <idx> <toCol>  - move the run starting at <idx> between columns
//	m <fromCol> f              - move top card to foundation
//
// **3 引数形が無いと、同スート降順の並びを一括で動かせない。**ドメインは
// cardIndex を受け取れるのに、CUI から名指しする文法が無ければ看板ルールが
// CUI では死ぬ。索引の妥当性はドメイン (範囲 + isRun) に検査させる。
func (c *PerseveranceCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m {0}")
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("perseverance.promptToZone"), fmt.Sprintf("m %s {0}", args[0]))
	}
	if args[1] == "f" {
		return c.bi.MoveTableauToFoundation(fromCol)
	}
	// 3 引数形: 真ん中が列内の開始位置。
	if len(args) >= 3 {
		cardIndex, err := strconv.Atoi(args[1])
		if err != nil {
			return invalidArg("invalidColumn", "val", args[1])
		}
		toCol, err := strconv.Atoi(args[2])
		if err != nil {
			return invalidArg("invalidColumn", "val", args[2])
		}
		return c.bi.MoveTableauToTableau(fromCol, cardIndex, toCol)
	}
	toCol, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[1])
	}
	return c.bi.MoveTableauToTableau(fromCol, -1, toCol)
}
