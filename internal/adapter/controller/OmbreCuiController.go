//go:build !js || !wasm || extra3

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OmbreCuiController オンブル (Ombre) のCUIコントローラークラス
type OmbreCuiController struct {
	di usecase.OmbreInteractorIF
}

// NewOmbreCuiController コンストラクタ
func NewOmbreCuiController(di usecase.OmbreInteractorIF) *OmbreCuiController {
	return &OmbreCuiController{di: di}
}

// ombreParseSuit スート文字列を CardDesign 値に変換する (-1=不正)。
func ombreParseSuit(s string) int {
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
//	bid pass                       → パス (降りる)
//	bid entrar <s|c|h|d>           → Entrar 宣言 (切り札スート指定)
//	bid solo <s|c|h|d>             → Solo 宣言 (切り札スート指定)
//	<n> / play <n>                 → カードをプレイ (プレイフェーズ)
//	n / next                       → 次のトリックへ
//	nr / nextround                 → 次のディールへ (スコアリング)
//	sd / setdifficulty <0-2>       → CPU難易度設定
//	h / hint                       → ヒント表示
//	log / l                        → 棋譜表示
func (c *OmbreCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.di.GetConfig()
			return c.di.ResetWithConfig(cfg)
		},
		[]string{
			"bid", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "bid":
				return c.execBid(args)
			case "play":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax, c.di.Play)
			case "n", "next":
				return c.di.NextTrick(), true
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuDifficulty = domain.OmbreCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}

// execBid bid サブコマンドを解釈する。
func (c *OmbreCuiController) execBid(args []string) (string, bool) {
	if len(args) == 0 {
		return "Bid action is required (pass, entrar <suit>, or solo <suit>).", true
	}
	switch args[0] {
	case "pass", "p":
		return c.di.Bid(domain.OmbreBidNone, -1), true
	case "entrar", "e":
		return c.bidWithTrump(domain.OmbreBidEntrar, args), true
	case "solo", "s":
		return c.bidWithTrump(domain.OmbreBidSolo, args), true
	default:
		return "Invalid bid action: " + args[0] + ". Please enter pass, entrar <suit>, or solo <suit>.", true
	}
}

// bidWithTrump entrar/solo の切り札スート引数を解釈してビッドする。
func (c *OmbreCuiController) bidWithTrump(bid domain.OmbreBid, args []string) string {
	if len(args) < 2 {
		return "Trump suit is required (s=spade, c=club, h=heart, d=diamond)."
	}
	suit := ombreParseSuit(args[1])
	if suit < 0 {
		return "Invalid trump suit: " + args[1] + ". Please enter s, c, h, or d."
	}
	return c.di.Bid(bid, suit)
}
