//go:build !js || !wasm || extra3

package controller

import (
	"fmt"
	"strconv"
	"strings"

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
				indices, err := buraParseIndices(args)
				if err != nil {
					return err.Error() + "\n", true
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

// buraParseIndices parses the card indices for `play`.
//
// Bura leads up to three cards at once, so this takes a list rather than the
// single index cuiutil.WithParsedInt handles. Duplicates are rejected here
// rather than left to the domain: repeating an index would otherwise read as
// a longer play than the hand can support.
func buraParseIndices(args []string) ([]int, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("Card index is required")
	}
	seen := map[int]bool{}
	indices := make([]int, 0, len(args))
	for _, a := range args {
		n, err := strconv.Atoi(strings.TrimSpace(a))
		if err != nil {
			return nil, fmt.Errorf("Invalid card index: %s", a)
		}
		if seen[n] {
			return nil, fmt.Errorf("Duplicate card index: %d", n)
		}
		seen[n] = true
		indices = append(indices, n)
	}
	return indices, nil
}
