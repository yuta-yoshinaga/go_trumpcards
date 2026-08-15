package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NertzCuiController Nertz / Pounce CUI コントローラー
type NertzCuiController struct {
	ni usecase.NertzInteractorIF
}

// NewNertzCuiController コンストラクタ
func NewNertzCuiController(ni usecase.NertzInteractorIF) *NertzCuiController {
	return &NertzCuiController{ni: ni}
}

// Exec コマンド実行
//   - d   <p>                   : ストック → ウェイスト
//   - mnf <p> <f>               : ナッツ → ファウンデーション
//   - mnt <p> <c>               : ナッツ → タブロー
//   - mwf <p> <f>               : ウェイスト → ファウンデーション
//   - mwt <p> <c>               : ウェイスト → タブロー
//   - mtf <p> <c> <f>           : タブロー → ファウンデーション
//   - mtt <p> <c> <i> <c2>      : タブロー → タブロー
//   - tick                      : CPU を 1tick 進める
//   - nr                        : 次ラウンド開始
//   - u, undo                   : 直前の操作を取り消す
//   - h, hint                   : ヒント
//   - log, l                    : 棋譜
func (c *NertzCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ni.Reset() },
		[]string{"d", "draw", "mnf", "mnt", "mwf", "mwt", "mtf", "mtt", "tick", "nr", "u", "undo", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.handleDraw(args), true
			case "mnf":
				return c.handleMoveNF(args), true
			case "mnt":
				return c.handleMoveNT(args), true
			case "mwf":
				return c.handleMoveWF(args), true
			case "mwt":
				return c.handleMoveWT(args), true
			case "mtf":
				return c.handleMoveTF(args), true
			case "mtt":
				return c.handleMoveTT(args), true
			case "tick":
				return c.ni.Tick(), true
			case "nr":
				return c.ni.NextRound(), true
			case "u", "undo":
				return c.ni.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.ni.Hint, c.ni.ActionLog)
			}
		},
	)
}

func (c *NertzCuiController) handleDraw(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("nertz.promptPlayerIdx"), "d {0}")
	}
	p, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[0])
	}
	return c.ni.Draw(p)
}

func (c *NertzCuiController) handleMoveNF(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("nertz.promptPlayerIdx"), "mnf {0} {1}")
	}
	p, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("nertz.promptFoundationIdx"), "mnf "+args[0]+" {0}")
	}
	f, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[1])
	}
	return c.ni.MoveNertzToFoundation(p, f)
}

func (c *NertzCuiController) handleMoveNT(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("nertz.promptPlayerIdx"), "mnt {0} {1}")
	}
	p, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("nertz.promptTableauCol"), "mnt "+args[0]+" {0}")
	}
	col, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[1])
	}
	return c.ni.MoveNertzToTableau(p, col)
}

func (c *NertzCuiController) handleMoveWF(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("nertz.promptPlayerIdx"), "mwf {0} {1}")
	}
	p, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("nertz.promptFoundationIdx"), "mwf "+args[0]+" {0}")
	}
	f, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[1])
	}
	return c.ni.MoveWasteToFoundation(p, f)
}

func (c *NertzCuiController) handleMoveWT(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("nertz.promptPlayerIdx"), "mwt {0} {1}")
	}
	p, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("nertz.promptTableauCol"), "mwt "+args[0]+" {0}")
	}
	col, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[1])
	}
	return c.ni.MoveWasteToTableau(p, col)
}

func (c *NertzCuiController) handleMoveTF(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("nertz.promptPlayerIdx"), "mtf {0} {1} {2}")
	}
	p, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("nertz.promptTableauCol"), "mtf "+args[0]+" {0} {1}")
	}
	col, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[1])
	}
	if len(args) < 3 {
		return cuiutil.PromptRequest(i18n.T("nertz.promptFoundationIdx"), "mtf "+args[0]+" "+args[1]+" {0}")
	}
	f, err := strconv.Atoi(args[2])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[2])
	}
	return c.ni.MoveTableauToFoundation(p, col, f)
}

func (c *NertzCuiController) handleMoveTT(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("nertz.promptPlayerIdx"), "mtt {0} {1} {2} {3}")
	}
	p, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("nertz.promptTableauCol"), "mtt "+args[0]+" {0} {1} {2}")
	}
	fromCol, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[1])
	}
	if len(args) < 3 {
		return cuiutil.PromptRequest(i18n.T("nertz.promptTableauFromIdx"), "mtt "+args[0]+" "+args[1]+" {0} {1}")
	}
	fromIdx, err := strconv.Atoi(args[2])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[2])
	}
	if len(args) < 4 {
		return cuiutil.PromptRequest(i18n.T("nertz.promptTableauToCol"), "mtt "+args[0]+" "+args[1]+" "+args[2]+" {0}")
	}
	toCol, err := strconv.Atoi(args[3])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[3])
	}
	return c.ni.MoveTableauToTableau(p, fromCol, fromIdx, toCol)
}
