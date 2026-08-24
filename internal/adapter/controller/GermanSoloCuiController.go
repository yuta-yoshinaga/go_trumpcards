//go:build !js || !wasm || classic

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GermanSoloCuiController ジャーマン・ソロ (GermanSolo) のCUIコントローラークラス
type GermanSoloCuiController struct {
	di usecase.GermanSoloInteractorIF
}

// NewGermanSoloCuiController コンストラクタ
func NewGermanSoloCuiController(di usecase.GermanSoloInteractorIF) *GermanSoloCuiController {
	return &GermanSoloCuiController{di: di}
}

// germanSoloParseSuit スート文字列を CardDesign 値に変換する (-1=不正)。
func germanSoloParseSuit(s string) int {
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
//	bid frage <s|c|h|d>            → Frage 宣言 (切り札スート指定、味方を呼ぶ)
//	bid solo <s|c|h|d>             → Solo 宣言 (切り札スート指定、単独 5 トリック)
//	bid tout <s|c|h|d>             → Tout 宣言 (切り札スート指定、単独 8 トリック)
//	ace <s|c|h|d>                 → 味方を呼ぶエースを指名 (エース呼びフェーズ)
//	play <n>                       → カードをプレイ (プレイフェーズ)
//	n / next                       → 次のトリックへ
//	nr / nextround                 → 次のディールへ (スコアリング)
//	sd / setdifficulty <0-2>       → CPU難易度設定
//	h / hint                       → ヒント表示
//	log / l                        → 棋譜表示
func (c *GermanSoloCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.di.GetConfig()
			return c.di.ResetWithConfig(cfg)
		},
		[]string{
			"bid", "ace", "a", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "bid":
				return c.execBid(args)
			case "ace", "a":
				return c.execAce(args)
			case "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.di.Play)
			case "n", "next":
				return c.di.NextTrick(), true
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuDifficulty = domain.GermanSoloCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}

// execAce ace サブコマンドを解釈する。
//
// **エース呼びフェーズを抜ける唯一の入力。** 落札の直後はこのフェーズで、
// エースが呼ばれるまで play は「フェーズが違う」で弾かれる。
// 引数はビッドと同じスート語 (spade/s, club/c, heart/h, diamond/d)。
func (c *GermanSoloCuiController) execAce(args []string) (string, bool) {
	if len(args) == 0 {
		return invalidArg("aceSuitRequiredWords"), true
	}
	suit := germanSoloParseSuit(args[0])
	if suit < 0 {
		return invalidArg("invalidAceSuitSCHD", "val", args[0]), true
	}
	return c.di.CallAce(suit), true
}

// execBid bid サブコマンドを解釈する。
func (c *GermanSoloCuiController) execBid(args []string) (string, bool) {
	if len(args) == 0 {
		return invalidArg("bidActionRequiredFrage"), true
	}
	switch args[0] {
	case "pass", "p":
		return c.di.Bid(domain.GermanSoloBidNone, -1), true
	case "frage", "f":
		return c.bidWithTrump(domain.GermanSoloBidFrage, args), true
	case "solo", "s":
		return c.bidWithTrump(domain.GermanSoloBidSolo, args), true
	case "tout", "t":
		return c.bidWithTrump(domain.GermanSoloBidTout, args), true
	default:
		return invalidArg("invalidBidActionFrage", "val", args[0]), true
	}
}

// bidWithTrump frage/solo/tout の切り札スート引数を解釈してビッドする。
func (c *GermanSoloCuiController) bidWithTrump(bid domain.GermanSoloBid, args []string) string {
	if len(args) < 2 {
		return invalidArg("trumpSuitRequiredWords")
	}
	suit := germanSoloParseSuit(args[1])
	if suit < 0 {
		return invalidArg("invalidTrumpSuitSCHD", "val", args[1])
	}
	return c.di.Bid(bid, suit)
}
