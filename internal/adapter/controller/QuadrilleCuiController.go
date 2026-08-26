//go:build !js || !wasm || extra4

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// QuadrilleCuiController カドリール (Quadrille) のCUIコントローラークラス
type QuadrilleCuiController struct {
	di usecase.QuadrilleInteractorIF
}

// NewQuadrilleCuiController コンストラクタ
func NewQuadrilleCuiController(di usecase.QuadrilleInteractorIF) *QuadrilleCuiController {
	return &QuadrilleCuiController{di: di}
}

// quadrilleParseSuit スート文字列を CardDesign 値に変換する (-1=不正)。
func quadrilleParseSuit(s string) int {
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
//	king <s|c|h|d>                 → 味方を呼ぶ王を指名 (王呼びフェーズ)
//	<n> / play <n>                 → カードをプレイ (プレイフェーズ)
//	n / next                       → 次のトリックへ
//	nr / nextround                 → 次のディールへ (スコアリング)
//	sd / setdifficulty <0-2>       → CPU難易度設定
//	h / hint                       → ヒント表示
//	log / l                        → 棋譜表示
func (c *QuadrilleCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.di.GetConfig()
			return c.di.ResetWithConfig(cfg)
		},
		[]string{
			"bid", "king", "k", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "bid":
				return c.execBid(args)
			case "king", "k":
				return c.execKing(args)
			case "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.di.Play)
			case "n", "next":
				return c.di.NextTrick(), true
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuDifficulty = domain.QuadrilleCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}

// execKing king サブコマンドを解釈する。
//
// **王呼びフェーズを抜ける唯一の入力。** 落札の直後はこのフェーズで、
// 王が呼ばれるまで play は「フェーズが違う」で弾かれる。
// 引数はビッドと同じスート語 (spade/s, club/c, heart/h, diamond/d)。
func (c *QuadrilleCuiController) execKing(args []string) (string, bool) {
	if len(args) == 0 {
		return invalidArg("kingSuitRequiredWords"), true
	}
	suit := quadrilleParseSuit(args[0])
	if suit < 0 {
		return invalidArg("invalidKingSuitSCHD", "val", args[0]), true
	}
	return c.di.CallKing(suit), true
}

// execBid bid サブコマンドを解釈する。
func (c *QuadrilleCuiController) execBid(args []string) (string, bool) {
	if len(args) == 0 {
		return invalidArg("bidActionRequiredEntrar"), true
	}
	switch args[0] {
	case "pass", "p":
		return c.di.Bid(domain.QuadrilleBidNone, -1), true
	case "entrar", "e":
		return c.bidWithTrump(domain.QuadrilleBidEntrar, args), true
	case "solo", "s":
		return c.bidWithTrump(domain.QuadrilleBidSolo, args), true
	default:
		return invalidArg("invalidBidActionEntrar", "val", args[0]), true
	}
}

// bidWithTrump entrar/solo の切り札スート引数を解釈してビッドする。
func (c *QuadrilleCuiController) bidWithTrump(bid domain.QuadrilleBid, args []string) string {
	if len(args) < 2 {
		return invalidArg("trumpSuitRequiredWords")
	}
	suit := quadrilleParseSuit(args[1])
	if suit < 0 {
		return invalidArg("invalidTrumpSuitSCHD", "val", args[1])
	}
	return c.di.Bid(bid, suit)
}
