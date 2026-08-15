//go:build !js || !wasm || extra3

package controller

import (
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MaoCuiController マオCUIコントローラークラス
type MaoCuiController struct {
	ci usecase.MaoInteractorIF
}

// NewMaoCuiController コンストラクタ
func NewMaoCuiController(ci usecase.MaoInteractorIF) *MaoCuiController {
	return &MaoCuiController{ci: ci}
}

// playWithPrompts runs Play(idx) and inlines a follow-up prompt when the secret
// rule fired (word declaration), the human just played an 8 (suit choice) or
// just reached one card (Mao declaration), so the user does not have to type the
// follow-up command on a separate line.
func (c *MaoCuiController) playWithPrompts(cardIndex int) string {
	res := c.ci.Play(cardIndex)
	if c.ci.IsHumanAwaitingWord() {
		return cuiutil.PromptRequest(i18n.T("mao.promptWord"), "dw {0}")
	}
	if c.ci.IsHumanChooseSuitTurn() {
		return cuiutil.PromptRequest(i18n.T("mao.promptSuit"), "s {0}")
	}
	if c.ci.IsHumanDeclareTurn() {
		return cuiutil.PromptRequest(i18n.T("mao.promptDeclare"), "dc")
	}
	return res
}

// Exec コマンド実行
func (c *MaoCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "d", "draw", "s", "suit",
			"dc", "declare", "sk", "skipdeclare", "dw", "declareword",
			"nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.playWithPrompts)
			case "d", "draw":
				return c.ci.Draw(), true
			case "s", "suit":
				return cuiutil.WithParsedIntKeys(args, "suitRequiredSymbols", "invalidSuitRange", 1, 4, c.ci.ChooseSuit)
			case "dc", "declare":
				return c.ci.Declare(), true
			case "sk", "skipdeclare":
				return c.ci.SkipDeclare(), true
			case "dw", "declareword":
				word := strings.TrimSpace(strings.Join(args, " "))
				if word == "" {
					return "A word is required.", true
				}
				return c.ci.DeclareWord(word), true
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.MaoCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				// Cap matches the Web controller's bound (MaoWebConfig.ToConfig) so the
				// CUI and Web layers enforce the same maximum point limit.
				return cuiutil.WithParsedIntKeys(args, "pointLimitRequired", "invalidPointLimit11000", 1, 1000, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.PointLimit = v
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.ci.ActionLog)
			}
		},
	)
}
