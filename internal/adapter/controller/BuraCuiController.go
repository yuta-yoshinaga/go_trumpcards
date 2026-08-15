//go:build !js || !wasm || extra3

package controller

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BuraCuiController ブラCUIコントローラークラス
type BuraCuiController struct {
	bi usecase.BuraInteractorIF
}

// NewBuraCuiController コンストラクタ
func NewBuraCuiController(bi usecase.BuraInteractorIF) *BuraCuiController {
	return &BuraCuiController{bi: bi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit           → ゲーム終了 ("bye.")
//	r / reset          → ゲームリセット (設定保持)
//	p / play <i...>    → カードをプレイ (リードは同スート最大3枚まとめて可)
//	c / claim          → 31点到達を宣言する
//	d / declare        → 手札の即勝ち役を宣言する
//	h / hint           → ヒント表示
//	log / l            → 棋譜表示
func (c *BuraCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.bi.GetConfig()
			return c.bi.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "c", "claim", "d", "declare", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				indices, msg := buraParseIndices(args)
				if msg != "" {
					return msg + "\n", true
				}
				return c.bi.Play(indices), true
			case "c", "claim":
				return c.bi.Claim(), true
			case "d", "declare":
				return c.bi.Declare(), true
			default:
				return handleCuiHintAndLog(cmd, c.bi.Hint, c.bi.ActionLog)
			}
		},
	)
}

// buraParseIndices parses the card indices for `play`, returning the indices
// or a user-facing message (never both).
//
// Bura leads up to three cards at once, so this takes a list rather than the
// single index cuiutil.WithParsedInt handles. Duplicates are rejected here
// rather than left to the domain: repeating an index would otherwise read as
// a longer play than the hand can support.
//
// The failure is a message rather than an error because it is printed to the
// player verbatim, and Go error strings are not capitalised.
func buraParseIndices(args []string) ([]int, string) {
	if len(args) == 0 {
		return nil, i18n.T("cardIndexRequired")
	}
	seen := map[int]bool{}
	indices := make([]int, 0, len(args))
	for _, a := range args {
		n, err := strconv.Atoi(strings.TrimSpace(a))
		if err != nil {
			return nil, i18n.Tf("invalidCardIndex", "val", a)
		}
		if seen[n] {
			return nil, fmt.Sprintf("Duplicate card index: %d.", n)
		}
		seen[n] = true
		indices = append(indices, n)
	}
	return indices, ""
}
