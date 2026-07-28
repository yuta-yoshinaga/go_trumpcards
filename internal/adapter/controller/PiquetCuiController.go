//go:build !js || !wasm || extra3

package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PiquetCuiController Piquet CUIコントローラークラス
type PiquetCuiController struct {
	pi usecase.PiquetInteractorIF
}

// NewPiquetCuiController コンストラクタ
func NewPiquetCuiController(pi usecase.PiquetInteractorIF) *PiquetCuiController {
	return &PiquetCuiController{pi: pi}
}

// Exec コマンド実行
func (c *PiquetCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.pi.Reset() },
		[]string{"e", "elder", "y", "younger", "d", "declare", "p", "play", "nd", "nextdeal", "h", "hint", "log", "l"},
		c.dispatch,
	)
}

// dispatch sub-command dispatcher
func (c *PiquetCuiController) dispatch(cmd string, args []string) (string, bool) {
	switch cmd {
	case "e", "elder":
		return c.handleExchange(args, c.pi.ExchangeElder), true
	case "y", "younger":
		return c.handleExchange(args, c.pi.ExchangeYounger), true
	case "d", "declare":
		return c.pi.ResolveDeclaration(), true
	case "p", "play":
		return c.handlePlay(args), true
	case "nd", "nextdeal":
		return c.pi.NextDeal(), true
	default:
		return handleCuiHintAndLog(cmd, c.pi.Hint, c.pi.ActionLog)
	}
}

// handleExchange は交換コマンドを処理する。
//
//	e 0,1,2     ... Elder が手札 [0,1,2] を捨ててタロンから3枚受け取る
//	y 0         ... Younger が手札 [0] を捨ててタロンから1枚受け取る
//	y           ... Younger が0枚交換 (パス)
func (c *PiquetCuiController) handleExchange(args []string, exec func([]int) string) string {
	if len(args) == 0 {
		return exec(nil)
	}
	indices, err := parseIndexList(args[0])
	if err != nil {
		return i18n.Tf("invalidIndex", "val", args[0])
	}
	return exec(indices)
}

// handlePlay プレイコマンド
func (c *PiquetCuiController) handlePlay(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("piquet.promptCardIndex"), "p {0}")
	}
	idx, err := strconv.Atoi(args[0])
	if err != nil {
		return i18n.Tf("invalidIndex", "val", args[0])
	}
	return c.pi.Play(idx)
}

// parseIndexList "0,1,2" -> []int{0,1,2}
func parseIndexList(s string) ([]int, error) {
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}
