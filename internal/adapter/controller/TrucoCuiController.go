package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TrucoCuiController トゥルコCUIコントローラークラス
type TrucoCuiController struct {
	ti usecase.TrucoInteractorIF
}

// NewTrucoCuiController コンストラクタ
func NewTrucoCuiController(ti usecase.TrucoInteractorIF) *TrucoCuiController {
	return &TrucoCuiController{ti: ti}
}

// setMatchTarget はマッチ目標点を設定してリセットする。
// 値を文言に埋め込まず {{min}}/{{max}} で渡すのは、定数を変えたときに
// メッセージだけ古いまま残るのを避けるため。
func (c *TrucoCuiController) setMatchTarget(args []string) string {
	if len(args) < 1 {
		return invalidArg("targetScoreRequired")
	}
	v, err := strconv.Atoi(args[0])
	if err != nil || v < domain.TrucoMinMatchTarget || v > domain.TrucoMaxMatchTarget {
		return invalidArg("invalidTargetScoreRange",
			"val", args[0],
			"min", strconv.Itoa(domain.TrucoMinMatchTarget),
			"max", strconv.Itoa(domain.TrucoMaxMatchTarget))
	}
	cfg := c.ti.GetConfig()
	cfg.MatchTarget = v
	return c.ti.ResetWithConfig(cfg)
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit         → ゲーム終了 ("bye.")
//	r / reset        → マッチリセット (設定保持)
//	p / play <i>     → カードをプレイ
//	t / truco        → Truco を宣言 (または再引き上げ)
//	a / accept       → 宣言を受諾 (Quiero)
//	d / decline      → 宣言を拒否 (No Quiero)
//	n / next         → 次のバサ / マノへ
//	sm / setmatchtarget <n> → マッチ目標点を変更してリセット
//	h / hint         → ヒント表示
//	log / l          → 棋譜表示
func (c *TrucoCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ti.GetConfig()
			return c.ti.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "t", "truco", "a", "accept", "d", "decline",
			"n", "next", "sm", "setmatchtarget", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ti.Play)
			case "t", "truco":
				return c.ti.Truco(), true
			case "a", "accept":
				return c.ti.Respond(true), true
			case "d", "decline":
				return c.ti.Respond(false), true
			case "sm", "setmatchtarget":
				// **範囲はドメインの定数から取る。**Web (TrucoWebController) も同じ
				// 定数でクランプしているので、両方の入り口で同じ値が通る (#5618)。
				return c.setMatchTarget(args), true
			case "n", "next":
				return c.ti.Next(), true
			default:
				return handleCuiHintAndLog(cmd, c.ti.Hint, c.ti.ActionLog)
			}
		},
	)
}
