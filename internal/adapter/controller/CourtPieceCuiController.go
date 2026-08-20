//go:build !js || !wasm || casino

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CourtPieceCuiController Court Piece CUI コントローラークラス
type CourtPieceCuiController struct {
	ti usecase.CourtPieceInteractorIF
}

// NewCourtPieceCuiController コンストラクタ
func NewCourtPieceCuiController(ti usecase.CourtPieceInteractorIF) *CourtPieceCuiController {
	return &CourtPieceCuiController{ti: ti}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	t / trump <suit>         → トランプを宣言 (1=♠ 2=♣ 3=♥ 4=♦)
//	p / play <i>             → カードをプレイ
//	n / next                 → 次のトリックへ
//	nr / nextround           → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	sl / setlimit <n>        → ポイント上限設定
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *CourtPieceCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ti.GetConfig()
			return c.ti.ResetWithConfig(cfg)
		},
		[]string{
			"t", "trump", "p", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "t", "trump":
				return cuiutil.WithParsedIntKeys(args, "trumpSuitRequiredNames", "invalidSuit", domain.CardDesignSpade, domain.CardDesignDiamond, c.ti.DeclareTrump)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ti.Play)
			case "n", "next":
				return c.ti.NextTrick(), true
			case "nr", "nextround":
				return c.ti.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.CpuDifficulty = domain.CourtPieceCpuDifficulty(v)
					return c.ti.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedIntKeys(args, "pointLimitRequired", "invalidPointLimit150", 1, domain.CourtPieceMaxPointLimit, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.PointLimit = v
					return c.ti.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ti.Hint, c.ti.ActionLog)
			}
		},
	)
}
