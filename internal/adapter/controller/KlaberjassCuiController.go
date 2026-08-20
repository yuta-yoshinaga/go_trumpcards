//go:build !js || !wasm || extra3

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KlaberjassCuiController クラバーヤス (Klaberjass) CUIコントローラークラス
type KlaberjassCuiController struct {
	ki usecase.KlaberjassInteractorIF
}

// NewKlaberjassCuiController コンストラクタ
func NewKlaberjassCuiController(ki usecase.KlaberjassInteractorIF) *KlaberjassCuiController {
	return &KlaberjassCuiController{ki: ki}
}

// Exec コマンド実行
func (c *KlaberjassCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ki.GetConfig()
			return c.ki.ResetWithConfig(cfg)
		},
		[]string{
			"a", "accept", "c", "call", "ps", "pass",
			"sm", "schmeiss", "y", "yes", "no",
			"p", "play", "n", "next",
			"st", "settarget", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "a", "accept":
				return c.ki.AcceptTrump(), true
			case "c", "call":
				return cuiutil.WithParsedIntKeys(args, "suitRequiredLetters", "invalidSuitRange", domain.CardDesignSpade, domain.CardDesignDiamond, func(v int) string {
					return c.ki.CallTrump(v)
				})
			case "ps", "pass":
				return c.ki.Pass(), true
			case "sm", "schmeiss":
				return c.ki.Schmeiss(), true
			case "y", "yes":
				return c.ki.AnswerSchmeiss(true), true
			case "no":
				return c.ki.AnswerSchmeiss(false), true
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", 0, 8, func(v int) string {
					return c.ki.PlayCard(v)
				})
			case "n", "next":
				return c.ki.NextDeal(), true
			case "st", "settarget":
				return cuiutil.WithParsedIntKeys(args, "targetScoreRequired", "invalidTargetScorePlain", domain.KlaberjassTargetScoreMin, domain.KlaberjassTargetScoreMax, func(v int) string {
					cfg := c.ki.GetConfig()
					cfg.TargetScore = v
					return c.ki.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.ki.ActionLog)
			}
		},
	)
}
