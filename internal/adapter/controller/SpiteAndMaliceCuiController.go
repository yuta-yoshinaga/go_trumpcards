package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SpiteAndMaliceCuiController Spite & Malice CUI コントローラー
type SpiteAndMaliceCuiController struct {
	si usecase.SpiteAndMaliceInteractorIF
}

// NewSpiteAndMaliceCuiController コンストラクタ
func NewSpiteAndMaliceCuiController(si usecase.SpiteAndMaliceInteractorIF) *SpiteAndMaliceCuiController {
	return &SpiteAndMaliceCuiController{si: si}
}

// Exec コマンド実行
// - ph <handIdx> <fIdx>  : 手札 → ファウンデーション
// - pg <fIdx>            : ゴール → ファウンデーション
// - ps <sideIdx> <fIdx>  : サイド → ファウンデーション
// - d <handIdx> <sideIdx>: ディスカード (ターン終了)
// - cpu                  : CPU を 1 ステップ進める
// - hint / h             : ヒント
// - log / l              : 棋譜
func (c *SpiteAndMaliceCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.si.Reset() },
		[]string{"ph", "pg", "ps", "d", "discard", "cpu", "ac", "autocomplete", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "ph":
				return c.handlePlayFromHand(args), true
			case "pg":
				return c.handlePlayFromGoal(args), true
			case "ps":
				return c.handlePlayFromSide(args), true
			case "d", "discard":
				return c.handleDiscard(args), true
			case "cpu":
				return c.si.CpuStep(), true
			case "ac", "autocomplete":
				return c.si.AutoComplete(), true
			default:
				return handleCuiHintAndLog(cmd, c.si.Hint, c.si.ActionLog)
			}
		},
	)
}

func (c *SpiteAndMaliceCuiController) handlePlayFromHand(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("spiteandmalice.promptHandIdx"), "ph {0} {1}")
	}
	handIdx, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("spiteandmalice.promptFoundationIdx"), "ph "+args[0]+" {0}")
	}
	fIdx, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[1])
	}
	return c.si.PlayFromHand(handIdx, fIdx)
}

func (c *SpiteAndMaliceCuiController) handlePlayFromGoal(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("spiteandmalice.promptFoundationIdx"), "pg {0}")
	}
	fIdx, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[0])
	}
	return c.si.PlayFromGoal(fIdx)
}

func (c *SpiteAndMaliceCuiController) handlePlayFromSide(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("spiteandmalice.promptSideIdx"), "ps {0} {1}")
	}
	sideIdx, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("spiteandmalice.promptFoundationIdx"), "ps "+args[0]+" {0}")
	}
	fIdx, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[1])
	}
	return c.si.PlayFromSide(sideIdx, fIdx)
}

func (c *SpiteAndMaliceCuiController) handleDiscard(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("spiteandmalice.promptHandIdx"), "d {0} {1}")
	}
	handIdx, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("spiteandmalice.promptSideIdx"), "d "+args[0]+" {0}")
	}
	sideIdx, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[1])
	}
	return c.si.Discard(handIdx, sideIdx)
}
