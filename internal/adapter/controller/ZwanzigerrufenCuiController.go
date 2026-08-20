//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ZwanzigerrufenCuiController ツヴァンツィガールーフェンの CUI コントローラー。
type ZwanzigerrufenCuiController struct {
	zi usecase.ZwanzigerrufenInteractorIF
}

// NewZwanzigerrufenCuiController コンストラクタ。
func NewZwanzigerrufenCuiController(zi usecase.ZwanzigerrufenInteractorIF) *ZwanzigerrufenCuiController {
	return &ZwanzigerrufenCuiController{zi: zi}
}

// Exec コマンド実行。
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	bid rufer|solo           → 入札
//	pass                     → パス
//	discard <i0> ... <i5>    → 場札交換で 6 枚を伏せる
//	play <n>                 → カードをプレイ
//	n / next                 → 次のトリックへ
//	nr / nextround           → 次のディールへ
//	sd / setdifficulty <0-2> → CPU 難易度設定
//	st / setdeals <1-12>     → ディール数設定
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *ZwanzigerrufenCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.zi.ResetWithConfig(c.zi.GetConfig())
		},
		[]string{
			"bid", "pass", "discard", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "st", "setdeals", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "bid":
				return c.execBid(args)
			case "pass":
				return c.zi.Pass(), true
			case "discard":
				return c.execDiscard(args)
			case "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired",
					"invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.zi.Play)
			case "n", "next":
				return c.zi.NextTrick(), true
			case "nr", "nextround":
				return c.zi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args,
					"cpuDifficultyRequired",
					"invalidCpuDifficulty", 0, 2, func(v int) string {
						cfg := c.zi.GetConfig()
						cfg.CpuDifficulty = domain.ZwanzigerrufenCpuDifficulty(v)
						return c.zi.ResetWithConfig(cfg)
					})
			case "st", "setdeals":
				return cuiutil.WithParsedIntKeys(args, "numberOfDealsRequired112", "invalidNumberOfDeals112",
					domain.ZwanzigerrufenMinDeals, domain.ZwanzigerrufenMaxDeals, func(v int) string {
						cfg := c.zi.GetConfig()
						cfg.TargetDeals = v
						return c.zi.ResetWithConfig(cfg)
					})
			default:
				return handleCuiHintAndLog(cmd, c.zi.Hint, c.zi.ActionLog)
			}
		},
	)
}

// execBid bid サブコマンドを解釈する。
//
// **trischaken は受け付けない。** 全員パスの結果としてしか成立しない契約なので、
// 打てるコマンドにすると「誰も落札しなかった」という前提が崩れる。
func (c *ZwanzigerrufenCuiController) execBid(args []string) (string, bool) {
	if len(args) == 0 {
		return invalidArg("bidRequiredRuferOrSolo"), true
	}
	bid := zwanzigerrufenParseBid(args[0])
	if bid == domain.ZwanzigerrufenBidPass {
		return invalidArg("invalidBidRuferOrSolo", "val", args[0]), true
	}
	return c.zi.Bid(bid), true
}

// execDiscard discard サブコマンドを解釈する (6 枚のインデックス)。
func (c *ZwanzigerrufenCuiController) execDiscard(args []string) (string, bool) {
	if len(args) < domain.ZwanzigerrufenTalonSize {
		return invalidArg("sixIndicesRequiredDiscard"), true
	}
	indices, skipped := cuiutil.ParseIntSlice(args)
	if len(skipped) > 0 {
		return invalidArg("invalidCardIndex", "val", skipped[0]), true
	}
	return c.zi.Discard(indices), true
}
