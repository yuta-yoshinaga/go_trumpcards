//go:build !js || !wasm || classic

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KarnoffelCuiController カルニッフェル (Karnöffel) CUIコントローラークラス
type KarnoffelCuiController struct {
	ki usecase.KarnoffelInteractorIF
}

// NewKarnoffelCuiController コンストラクタ
func NewKarnoffelCuiController(ki usecase.KarnoffelInteractorIF) *KarnoffelCuiController {
	return &KarnoffelCuiController{ki: ki}
}

// Exec コマンド実行
func (c *KarnoffelCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ki.GetConfig()
			return c.ki.ResetWithConfig(cfg)
		},
		[]string{"p", "play", "n", "next", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					0, domain.KarnoffelHandSize-1, func(v int) string {
						return c.ki.PlayCard(v)
					})
			case "n", "next":
				return c.ki.NextHand(), true
			default:
				return handleCuiLog(cmd, c.ki.ActionLog)
			}
		},
	)
}
