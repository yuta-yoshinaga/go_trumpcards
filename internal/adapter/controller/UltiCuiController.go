//go:build !js || !wasm || extra3

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// UltiCuiController ウルティ (Ulti) のCUIコントローラークラス
type UltiCuiController struct {
	di usecase.UltiInteractorIF
}

// NewUltiCuiController コンストラクタ
func NewUltiCuiController(di usecase.UltiInteractorIF) *UltiCuiController {
	return &UltiCuiController{di: di}
}

// ultiParseSuit スート文字列を CardDesign 値に変換する (-1=不正)。
func ultiParseSuit(s string) int {
	switch s {
	case "spade", "spades", "s":
		return domain.CardDesignSpade
	case "club", "clubs", "c":
		return domain.CardDesignClover
	case "heart", "hearts", "h":
		return domain.CardDesignHeart
	case "diamond", "diamonds", "d":
		return domain.CardDesignDiamond
	default:
		return -1
	}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                       → ゲーム終了 ("bye.")
//	r / reset                      → ゲームリセット (設定保持)
//	bid party <s|c|h|d>            → Party 宣言 (切り札スート指定)
//	bid betli                      → Betli 宣言
//	bid durchmarsch                → Durchmarsch 宣言
//	bid ulti <s|c|h|d>             → Ulti 宣言 (切り札の 7 で最終トリック勝ち)
//	discard <i> <j>                → タロン受け取り後に 2 枚を捨てる
//	<n> / play <n>                 → カードをプレイ (プレイフェーズ)
//	n / next                       → 次のトリックへ
//	nr / nextround                 → 次のディールへ (スコアリング)
//	sd / setdifficulty <0-2>       → CPU難易度設定
//	h / hint                       → ヒント表示
//	log / l                        → 棋譜表示
func (c *UltiCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.di.GetConfig()
			return c.di.ResetWithConfig(cfg)
		},
		[]string{
			"bid", "discard", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "bid":
				return c.execBid(args)
			case "discard":
				return c.execDiscard(args)
			case "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.di.Play)
			case "n", "next":
				return c.di.NextTrick(), true
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuDifficulty = domain.UltiCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}

// execBid bid サブコマンドを解釈する。
func (c *UltiCuiController) execBid(args []string) (string, bool) {
	if len(args) == 0 {
		return invalidArg("bidActionRequiredParty"), true
	}
	switch args[0] {
	case "party", "p":
		return c.bidWithTrump(domain.UltiContractParty, args), true
	case "betli", "b":
		return c.di.Bid(domain.UltiContractBetli, -1), true
	case "durchmarsch", "d":
		return c.di.Bid(domain.UltiContractDurchmarsch, -1), true
	case "ulti", "u":
		return c.bidWithTrump(domain.UltiContractUlti, args), true
	default:
		return invalidArg("invalidBidActionParty", "val", args[0]), true
	}
}

// bidWithTrump 切り札スート引数を解釈して切り札コントラクト (Party / Ulti) を宣言する。
func (c *UltiCuiController) bidWithTrump(contract domain.UltiContract, args []string) string {
	if len(args) < 2 {
		return invalidArg("trumpSuitRequiredWords")
	}
	suit := ultiParseSuit(args[1])
	if suit < 0 {
		return invalidArg("invalidTrumpSuitSCHD", "val", args[1])
	}
	return c.di.Bid(contract, suit)
}

// execDiscard discard サブコマンドを解釈する (2 枚のインデックス)。
func (c *UltiCuiController) execDiscard(args []string) (string, bool) {
	if len(args) < domain.UltiDiscardSize {
		return invalidArg("twoIndicesRequiredDiscard"), true
	}
	indices, skipped := cuiutil.ParseIntSlice(args)
	if len(skipped) > 0 {
		return invalidArg("invalidCardIndex", "val", skipped[0]), true
	}
	return c.di.Discard(indices), true
}
