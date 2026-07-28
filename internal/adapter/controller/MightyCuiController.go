//go:build !js || !wasm || extra2

package controller

import (
	"math"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MightyCuiController マイティCUIコントローラークラス
type MightyCuiController struct {
	mi usecase.MightyInteractorIF
}

// NewMightyCuiController コンストラクタ
func NewMightyCuiController(mi usecase.MightyInteractorIF) *MightyCuiController {
	return &MightyCuiController{mi: mi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                              → ゲーム終了 ("bye.")
//	r / reset                             → ゲームリセット (設定保持)
//	b / bid <n> [nt]                      → ビッドを宣言 (0=パス、nt=ノートランプ)
//	t / trump <suit> <partnerSuit> <partnerVal> → 切り札とパートナーを宣言 (suit=-1 でノートランプ)
//	e / exchange <i> <j> <k>              → 場札交換 (捨てるカード3枚)
//	p / play <i>                          → カードをプレイ
//	jl / jokerlead <i> <suit>             → ジョーカーをリード (要求スート指定)
//	n / next                              → 次のトリックへ
//	nr / nextround                        → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2>              → CPU難易度設定
//	sl / setlimit <n>                     → ポイント上限設定
//	sm / setminbid <n>                    → 最低ビッド設定
//	sn / setnotrumpextra <n>              → ノートランプ加算値設定
//	h / hint                              → ヒント表示
//	log / l                               → 棋譜表示
func (c *MightyCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.mi.GetConfig()
			return c.mi.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid", "t", "trump", "e", "exchange", "p", "play",
			"jl", "jokerlead",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit", "sm", "setminbid", "sn", "setnotrumpextra",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				if len(args) == 0 {
					return "Bid value is required (0=pass, 13-20). Add 'nt' for no-trump.", true
				}
				noTrump := false
				bidArgs := args
				for i, a := range args {
					la := strings.ToLower(a)
					if la == "nt" || la == "notrump" {
						noTrump = true
						bidArgs = append(append([]string{}, args[:i]...), args[i+1:]...)
						break
					}
				}
				val, errMsg, ok := cuiutil.ParseIntArg(bidArgs, "Bid value is required (0=pass, 13-20).", "Invalid bid value: %s.", 0, domain.MightyMaxPoints)
				if !ok {
					return errMsg, true
				}
				return c.mi.Bid(val, noTrump), true
			case "t", "trump":
				if len(args) < 3 {
					return "Usage: trump <suit> <partnerSuit> <partnerVal>\n  suit: -1=No-Trump 1=Spade 2=Clover 3=Heart 4=Diamond\n  partnerSuit: 0=Joker 1=Spade 2=Clover 3=Heart 4=Diamond\n  partnerVal: 1=A 2-10 11=J 12=Q 13=K (Joker: 1)\n", true
				}
				suit, errMsg, ok := cuiutil.ParseIntArg(args[:1], "", "Invalid suit: %s.", -1, 4)
				if !ok {
					return errMsg, true
				}
				partnerSuit, errMsg, ok := cuiutil.ParseIntArg(args[1:2], "", "Invalid partner suit: %s.", 0, 4)
				if !ok {
					return errMsg, true
				}
				partnerVal, errMsg, ok := cuiutil.ParseIntArg(args[2:3], "", "Invalid partner value: %s.", 1, 13)
				if !ok {
					return errMsg, true
				}
				return c.mi.DeclareTrumpAndFriend(suit, partnerSuit, partnerVal), true
			case "e", "exchange":
				if len(args) < 3 {
					return "Usage: exchange <i> <j> <k>  (three card indices to discard from kitty pickup)\n", true
				}
				idxs := make([]int, 3)
				for i := 0; i < 3; i++ {
					v, errMsg, ok := cuiutil.ParseIntArg(args[i:i+1], "", "Invalid card index: %s.", 0, math.MaxInt)
					if !ok {
						return errMsg, true
					}
					idxs[i] = v
				}
				return c.mi.ExchangeKitty(idxs), true
			case "p", "play":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax, c.mi.Play)
			case "jl", "jokerlead":
				if len(args) < 2 {
					return "Usage: jokerlead <cardIndex> <demandSuit>\n  demandSuit: 1=Spade 2=Clover 3=Heart 4=Diamond\n", true
				}
				cardIdx, errMsg, ok := cuiutil.ParseIntArg(args[:1], "", "Invalid card index: %s.", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				demandSuit, errMsg, ok := cuiutil.ParseIntArg(args[1:2], "", "Invalid demand suit: %s.", 1, 4)
				if !ok {
					return errMsg, true
				}
				return c.mi.PlayJokerLead(cardIdx, demandSuit), true
			case "n", "next":
				return c.mi.NextTrick(), true
			case "nr", "nextround":
				return c.mi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.mi.GetConfig()
					cfg.CpuDifficulty = domain.MightyCpuDifficulty(v)
					return c.mi.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedInt(args, "Point limit is required.", "Invalid point limit: %s. Please enter 1 or more.", 1, math.MaxInt, func(v int) string {
					cfg := c.mi.GetConfig()
					cfg.PointLimit = v
					return c.mi.ResetWithConfig(cfg)
				})
			case "sm", "setminbid":
				return cuiutil.WithParsedInt(args, "Min bid is required.", "Invalid min bid: %s.", 1, domain.MightyMaxPoints, func(v int) string {
					cfg := c.mi.GetConfig()
					cfg.MinBid = v
					return c.mi.ResetWithConfig(cfg)
				})
			case "sn", "setnotrumpextra":
				return cuiutil.WithParsedInt(args, "No-trump extra is required.", "Invalid no-trump extra: %s.", 0, domain.MightyMaxPoints-1, func(v int) string {
					cfg := c.mi.GetConfig()
					cfg.NoTrumpExtra = v
					return c.mi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.mi.Hint, c.mi.ActionLog)
			}
		},
	)
}
