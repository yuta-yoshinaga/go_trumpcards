//go:build !js || !wasm || extra3

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MichiganCuiController はミシガン (Michigan) の CUI コントローラークラス。
type MichiganCuiController struct {
	ti usecase.MichiganInteractorIF
}

// NewMichiganCuiController コンストラクタ。
func NewMichiganCuiController(ti usecase.MichiganInteractorIF) *MichiganCuiController {
	return &MichiganCuiController{ti: ti}
}

// Exec コマンド実行。
//
//	q / quit                  → ゲーム終了 ("bye.")
//	r / reset                 → ゲームリセット (設定保持)
//	b / bet <h> <c> <d> <s>   → ブードルへの賭け (♥ / ♣ / ♦ / ♠ 順、合計 = アンティ)
//	p / play <n>              → 手札インデックス n のカードを出す
//	n / next / nr / nextround → 次のラウンドへ
//	sp <n> / setplayers <n>   → プレイヤー数設定 (3-8)
//	sa <n> / setante <n>      → アンティ額設定
//	sc <n> / setchips <n>     → 初期チップ設定
//	st <n> / setrounds <n>    → 実施ラウンド数設定
//	h / hint                  → ヒント表示
//	l / log                   → 棋譜表示
func (c *MichiganCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ti.GetConfig()
			return c.ti.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bet", "p", "play",
			"n", "next", "nr", "nextround",
			"sp", "setplayers", "sa", "setante",
			"sc", "setchips", "st", "setrounds",
			"h", "hint", "l", "log",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				return c.handleBet(args), true
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequiredExample", "invalidCardIndex", 0, 51, func(v int) string {
					return c.ti.Play(v)
				})
			case "n", "next", "nr", "nextround":
				return c.ti.NextRound(), true
			case "sp", "setplayers":
				return cuiutil.WithParsedIntKeys(args, "playerCountRequiredEGSp4", "invalidPlayerCount38", domain.MichiganMinPlayerCount, domain.MichiganMaxPlayerCount, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.PlayerCount = v
					return c.ti.ResetWithConfig(cfg)
				})
			case "sa", "setante":
				return cuiutil.WithParsedIntKeys(args, "anteRequiredEGSa8", "invalidAntePlain", domain.MichiganMinAnte, domain.MichiganMaxAnte, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.Ante = v
					return c.ti.ResetWithConfig(cfg)
				})
			case "sc", "setchips":
				return cuiutil.WithParsedIntKeys(args, "startingChipsRequired", "invalidStartingChips", domain.MichiganMinStartingChips, domain.MichiganMaxStartingChips, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.StartingChips = v
					return c.ti.ResetWithConfig(cfg)
				})
			case "st", "setrounds":
				return cuiutil.WithParsedIntKeys(args, "targetRoundsRequired", "invalidTargetRounds", domain.MichiganMinTargetRounds, domain.MichiganMaxTargetRounds, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.TargetRounds = v
					return c.ti.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ti.Hint, c.ti.ActionLog)
			}
		},
	)
}

// handleBet は "bet <h> <c> <d> <s>" を解析して 4 つのブードルへの賭けを適用する。
func (c *MichiganCuiController) handleBet(args []string) string {
	if len(args) < domain.MichiganBoodleCount {
		return i18n.MarkError(i18n.T("michigan.betArgsRequired"))
	}
	bets := make([]int, domain.MichiganBoodleCount)
	for i := 0; i < domain.MichiganBoodleCount; i++ {
		v, err := strconv.Atoi(args[i])
		if err != nil || v < 0 {
			return i18n.MarkError(i18n.Tf("michigan.betInvalid", "value", args[i]))
		}
		bets[i] = v
	}
	return c.ti.Bet(bets)
}
