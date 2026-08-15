//go:build !js || !wasm || casino

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TarneebCuiController Tarneeb CUI コントローラークラス
type TarneebCuiController struct {
	ti usecase.TarneebInteractorIF
}

// NewTarneebCuiController コンストラクタ
func NewTarneebCuiController(ti usecase.TarneebInteractorIF) *TarneebCuiController {
	return &TarneebCuiController{ti: ti}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	b / bid <n>              → ビッドを宣言 (0=パス、7-13=ビッド)
//	t / trump <suit>         → トランプを宣言 (1=♠ 2=♣ 3=♥ 4=♦)
//	p / play <i>             → カードをプレイ
//	n / next                 → 次のトリックへ
//	nr / nextround           → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	sl / setlimit <n>        → ポイント上限設定
//	sm / setminbid <n>       → 最低ビッド設定 (1-13)
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *TarneebCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ti.GetConfig()
			return c.ti.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid", "t", "trump", "p", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit", "sm", "setminbid",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return cuiutil.WithParsedIntKeys(args, "bidValueRequiredPass713", "invalidBidValue", 0, domain.TarneebMaxBid, c.ti.Bid)
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
					cfg.CpuDifficulty = domain.TarneebCpuDifficulty(v)
					return c.ti.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedIntKeys(args, "pointLimitRequired", "invalidPointLimit1200", 1, domain.TarneebMaxPointLimit, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.PointLimit = v
					return c.ti.ResetWithConfig(cfg)
				})
			case "sm", "setminbid":
				return cuiutil.WithParsedIntKeys(args, "minimumBidRequired113", "invalidMinBid", 1, 13, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.MinBid = v
					return c.ti.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ti.Hint, c.ti.ActionLog)
			}
		},
	)
}
