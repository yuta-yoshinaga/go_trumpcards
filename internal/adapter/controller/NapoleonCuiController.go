package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NapoleonCuiController ナポレオンCUIコントローラークラス
type NapoleonCuiController struct {
	ni usecase.NapoleonInteractorIF
}

// NewNapoleonCuiController コンストラクタ
func NewNapoleonCuiController(ni usecase.NapoleonInteractorIF) *NapoleonCuiController {
	return &NapoleonCuiController{ni: ni}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit         → ゲーム終了 ("bye.")
//	r / reset        → ゲームリセット (設定保持)
//	b / bid <n>      → ビッドを宣言 (0=パス)
//	t / trump <suit> <adjSuit> <adjVal> → 切り札と副官を宣言
//	e / exchange <i> → 場札交換 (捨てるカードインデックス)
//	p / play <i>     → カードをプレイ
//	n / next         → 次のトリックへ
//	nr / nextround   → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	sl / setlimit <n> → ポイント上限設定
//	sm / setminbid <n> → 最低ビッド設定
//	h / hint         → ヒント表示
//	log / l          → 棋譜表示
func (c *NapoleonCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ni.GetConfig()
			return c.ni.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid", "t", "trump", "e", "exchange", "p", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit", "sm", "setminbid",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				v, errMsg, ok := cuiutil.ParseIntArg(args, "Bid value is required (0=pass, 12-17).", "Invalid bid value: %s.", 0, domain.NapoleonMaxPictureCards)
				if !ok {
					return errMsg, true
				}
				return c.ni.Bid(v), true
			case "t", "trump":
				if len(args) < 3 {
					return "Usage: trump <suit> <adjSuit> <adjVal>\n  suit: 1=Spade 2=Club 3=Heart 4=Diamond\n  adjSuit: 0=Joker 1=Spade 2=Club 3=Heart 4=Diamond\n  adjVal: 1=A 2-10 11=J 12=Q 13=K (Joker: 1)\n", true
				}
				suit, errMsg, ok := cuiutil.ParseIntArg(args[:1], "", "Invalid suit: %s.", 1, 4)
				if !ok {
					return errMsg, true
				}
				adjSuit, errMsg, ok := cuiutil.ParseIntArg(args[1:2], "", "Invalid adjutant suit: %s.", 0, 4)
				if !ok {
					return errMsg, true
				}
				adjVal, errMsg, ok := cuiutil.ParseIntArg(args[2:3], "", "Invalid adjutant value: %s.", 1, 13)
				if !ok {
					return errMsg, true
				}
				return c.ni.DeclareTrump(suit, adjSuit, adjVal), true
			case "e", "exchange":
				idx, errMsg, ok := cuiutil.ParseIntArg(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax)
				if !ok {
					return errMsg, true
				}
				return c.ni.ExchangeKitty(idx), true
			case "p", "play":
				idx, errMsg, ok := cuiutil.ParseIntArg(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax)
				if !ok {
					return errMsg, true
				}
				return c.ni.Play(idx), true
			case "n", "next":
				return c.ni.NextTrick(), true
			case "nr", "nextround":
				return c.ni.NextRound(), true
			case "sd", "setdifficulty":
				v, errMsg, ok := cuiutil.ParseIntArg(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2)
				if !ok {
					return errMsg, true
				}
				cfg := c.ni.GetConfig()
				cfg.CpuDifficulty = domain.NapoleonCpuDifficulty(v)
				return c.ni.ResetWithConfig(cfg), true
			case "sl", "setlimit":
				v, errMsg, ok := cuiutil.ParseIntArg(args, "Point limit is required.", "Invalid point limit: %s. Please enter 1 or more.", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				cfg := c.ni.GetConfig()
				cfg.PointLimit = v
				return c.ni.ResetWithConfig(cfg), true
			case "sm", "setminbid":
				v, errMsg, ok := cuiutil.ParseIntArg(args, "Min bid is required.", "Invalid min bid: %s.", 1, domain.NapoleonMaxPictureCards)
				if !ok {
					return errMsg, true
				}
				cfg := c.ni.GetConfig()
				cfg.MinBid = v
				return c.ni.ResetWithConfig(cfg), true
			case "h", "hint":
				return c.ni.Hint(), true
			case "log", "l":
				return c.ni.ActionLog(), true
			}
			return "", false
		},
	)
}
