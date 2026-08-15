//go:build !js || !wasm || extra3

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CegoCuiController チェゴ (Cego) のCUIコントローラークラス
type CegoCuiController struct {
	di usecase.CegoInteractorIF
}

// NewCegoCuiController コンストラクタ
func NewCegoCuiController(di usecase.CegoInteractorIF) *CegoCuiController {
	return &CegoCuiController{di: di}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                          → ゲーム終了 ("bye.")
//	r / reset                         → ゲームリセット (設定保持)
//	bid play                          → 入札 (プレイ宣言)
//	pass                              → パス
//	cego                              → コントラクト Cego (場札交換)
//	handspiel                         → コントラクト Handspiel (手札のまま)
//	discard <i>                       → Cego 交換で残す 1 枚を選ぶ
//	<n> / play <n>                    → カードをプレイ (プレイフェーズ)
//	n / next                          → 次のトリックへ
//	nr / nextround                    → 次のディールへ (スコアリング)
//	sd / setdifficulty <0-2>          → CPU難易度設定
//	h / hint                          → ヒント表示
//	log / l                           → 棋譜表示
func (c *CegoCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.di.GetConfig()
			return c.di.ResetWithConfig(cfg)
		},
		[]string{
			"bid", "pass", "cego", "handspiel", "discard", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "bid":
				return c.execBid(args)
			case "pass":
				return c.di.Pass(), true
			case "cego":
				return c.di.ChooseContract(domain.CegoContractCego), true
			case "handspiel", "solo":
				return c.di.ChooseContract(domain.CegoContractHandspiel), true
			case "discard":
				return cuiutil.WithParsedIntKeys(args, "cardIndexToKeepRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, func(v int) string {
					return c.di.Discard([]int{v})
				})
			case "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.di.Play)
			case "n", "next":
				return c.di.NextTrick(), true
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuDifficulty = domain.CegoCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}

// execBid bid サブコマンドを解釈する。
func (c *CegoCuiController) execBid(args []string) (string, bool) {
	if len(args) == 0 {
		return invalidArg("bidRequiredPlay"), true
	}
	bid := cegoParseBid(args[0])
	if bid == domain.CegoBidPass {
		return "Invalid bid: " + args[0] + ". Please enter play.", true
	}
	return c.di.Bid(bid), true
}
