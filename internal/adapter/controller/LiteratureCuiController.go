//go:build !js || !wasm || extra4

package controller

import (
	"slices"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// literaturePendingClaim は確認待ちの宣言。
type literaturePendingClaim struct {
	half    int
	holders []int
}

// LiteratureCuiController リテラチャー (Literature) CUIコントローラークラス
type LiteratureCuiController struct {
	li usecase.LiteratureInteractorIF
	// pendingClaim は確認待ちの宣言 (nil = 待ちなし)。
	//
	// **宣言はこのゲームで最も重い一手。**1 席でも誤ると組は無効化されるか
	// 相手チームに渡る。Web は #4822 で確認ダイアログを挟んだのに、CUI は
	// 6 個の席番号を 1 行打った瞬間に確定していた (#5733)。
	pendingClaim *literaturePendingClaim
}

// NewLiteratureCuiController コンストラクタ
func NewLiteratureCuiController(li usecase.LiteratureInteractorIF) *LiteratureCuiController {
	return &LiteratureCuiController{li: li}
}

// Exec コマンド実行
func (c *LiteratureCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			// **r / reset は gameHandler を通らない** (execCuiCommand が先に拾う)。
			// ここで消さないと、リセット直後の y が配り直す前の宣言を確定させる
			// (レビュー指摘 #6070)。
			c.pendingClaim = nil
			cfg := c.li.GetConfig()
			return c.li.ResetWithConfig(cfg)
		},
		[]string{"a", "ask", "c", "claim", "y", "yes", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			// **確認待ちは宣言と確認以外のどのコマンドでも取り消す。**残したまま
			// だと、あとで打った y が意図しない宣言を確定させてしまう。
			pending := c.pendingClaim
			switch cmd {
			case "c", "claim", "y", "yes":
			default:
				c.pendingClaim = nil
			}

			switch cmd {
			case "a", "ask":
				return literatureParseAsk(args, c.li)
			case "c", "claim":
				return c.handleClaim(args, pending)
			case "y", "yes":
				if pending == nil {
					return invalidArg("literature.nothingToConfirm"), true
				}
				c.pendingClaim = nil
				return c.li.Claim(pending.half, pending.holders), true
			default:
				return handleCuiLog(cmd, c.li.ActionLog)
			}
		},
	)
}

// literatureParseAsk は `a <相手> <スート> <ランク>` を解釈する。
//
// **相手・スート・ランクの 3 つが揃って初めて要求になる。**
func literatureParseAsk(args []string, li usecase.LiteratureInteractorIF) (string, bool) {
	if len(args) < 3 {
		return invalidArg("askNeedsSeatSuitRank"), true
	}
	to, err := strconv.Atoi(args[0])
	if err != nil || to < 0 || to >= domain.LiteraturePlayerCnt {
		return invalidArg("invalidSeat0Max", "val", args[0], "max", strconv.Itoa(domain.LiteraturePlayerCnt-1)), true
	}
	suit, err := strconv.Atoi(args[1])
	if err != nil || suit < domain.CardDesignSpade || suit > domain.CardDesignDiamond {
		return invalidArg("invalidSuit14Letters", "val", args[1]), true
	}
	value, err := strconv.Atoi(args[2])
	if err != nil || value < 1 || value > 13 {
		return invalidArg("invalidRank113", "val", args[2]), true
	}
	return li.Ask(to, suit, value), true
}

// handleClaim は宣言を 2 段階にする。
//
// 1 度目は内容をそのまま読み上げるだけで確定させず、**同じ内容をもう一度打つか
// y で確認**して初めて Claim を呼ぶ。
func (c *LiteratureCuiController) handleClaim(args []string, pending *literaturePendingClaim) (string, bool) {
	half, holders, errMsg, ok := literatureParseClaim(args)
	if !ok {
		// 打ち間違いは待ちを引き継がない。次の y が古い内容を確定させてしまう。
		c.pendingClaim = nil
		return errMsg, true
	}
	if pending != nil && pending.half == half && slices.Equal(pending.holders, holders) {
		c.pendingClaim = nil
		return c.li.Claim(half, holders), true
	}
	c.pendingClaim = &literaturePendingClaim{half: half, holders: holders}
	return literatureClaimPreview(half, holders), true
}

// literatureClaimPreview は確定前に申告内容を読み上げる。
//
// **札と席を組にして出す。**6 個の数字が並んでいるだけでは、打ち間違いを
// 見つけられない。
func literatureClaimPreview(half int, holders []int) string {
	cards := domain.LiteratureHalfSuitCards(half)
	var b strings.Builder
	for i, seat := range holders {
		if i > 0 {
			b.WriteString(", ")
		}
		// 添字は literatureParseClaim が検証済みで、有効な half の
		// LiteratureHalfSuitCards は必ず 6 枚返る。
		b.WriteString(i18n.Tf("literature.claimPreviewPair",
			"rank", strconv.Itoa(cards[i].GetValue()), "seat", strconv.Itoa(seat)))
	}
	return i18n.Tf("literature.claimPreview",
		"half", strconv.Itoa(half), "pairs", b.String())
}

// literatureParseClaim は `c <組> <席×6>` を解釈する。
//
// **6 枚すべての所在を申告しなければ宣言にならない。**組の札順は
// LiteratureHalfSuitCards の並びに対応する。
func literatureParseClaim(args []string) (half int, holders []int, errMsg string, ok bool) {
	if len(args) < 1+domain.LiteratureHalfSuitSize {
		return 0, nil, invalidArg("claimNeedsHalfSuitAndHolders"), false
	}
	half, err := strconv.Atoi(args[0])
	if err != nil || half < 0 || half >= domain.LiteratureHalfSuitCnt {
		return 0, nil, invalidArg("invalidHalfSuit0Max", "val", args[0], "max", strconv.Itoa(domain.LiteratureHalfSuitCnt-1)), false
	}
	holders = make([]int, 0, domain.LiteratureHalfSuitSize)
	for i := range domain.LiteratureHalfSuitSize {
		seat, err := strconv.Atoi(args[1+i])
		if err != nil || seat < 0 || seat >= domain.LiteraturePlayerCnt {
			return 0, nil, invalidArg("invalidSeat0Max", "val", args[1+i], "max", strconv.Itoa(domain.LiteraturePlayerCnt-1)), false
		}
		holders = append(holders, seat)
	}
	return half, holders, "", true
}
