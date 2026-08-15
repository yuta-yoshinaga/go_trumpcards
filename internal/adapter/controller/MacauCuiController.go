//go:build !js || !wasm || solo

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MacauCuiController マカオCUIコントローラークラス
type MacauCuiController struct {
	ci usecase.MacauInteractorIF
}

// NewMacauCuiController コンストラクタ
func NewMacauCuiController(ci usecase.MacauInteractorIF) *MacauCuiController {
	return &MacauCuiController{ci: ci}
}

// playWithPrompts runs Play(idx) and inlines a follow-up prompt when the human
// just played an 8 (suit choice) or just reached one card (Macau declaration),
// so the user does not have to type the follow-up command on a separate line.
func (c *MacauCuiController) playWithPrompts(cardIndex int) string {
	res := c.ci.Play(cardIndex)
	if c.ci.IsHumanChooseSuitTurn() {
		return cuiutil.PromptRequest(i18n.T("macau.promptSuit"), "s {0}")
	}
	if c.ci.IsHumanDeclareTurn() {
		return cuiutil.PromptRequest(i18n.T("macau.promptDeclare"), "dc")
	}
	return res
}

// Exec コマンド実行
func (c *MacauCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "d", "draw", "s", "suit",
			"dc", "declare", "sk", "skipdeclare",
			"nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.playWithPrompts)
			case "d", "draw":
				return c.ci.Draw(), true
			case "s", "suit":
				return cuiutil.WithParsedInt(args, "Suit is required (1=♠, 2=♣, 3=♥, 4=♦).", "Invalid suit: %s. Please enter 1-4.", 1, 4, c.ci.ChooseSuit)
			case "dc", "declare":
				return c.ci.Declare(), true
			case "sk", "skipdeclare":
				return c.ci.SkipDeclare(), true
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.MacauCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				// Cap matches the Web controller's bound (MacauWebConfig.ToConfig) so the
				// CUI and Web layers enforce the same maximum point limit.
				return cuiutil.WithParsedInt(args, "Point limit is required.", "Invalid point limit: %s. Please enter 1-1000.", 1, 1000, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.PointLimit = v
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ci.Hint, c.ci.ActionLog)
			}
		},
	)
}
