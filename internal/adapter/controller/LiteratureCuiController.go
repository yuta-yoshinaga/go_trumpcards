//go:build !js || !wasm || solo

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// LiteratureCuiController リテラチャー (Literature) CUIコントローラークラス
type LiteratureCuiController struct {
	li usecase.LiteratureInteractorIF
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
			cfg := c.li.GetConfig()
			return c.li.ResetWithConfig(cfg)
		},
		[]string{"a", "ask", "c", "claim", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "a", "ask":
				return literatureParseAsk(args, c.li)
			case "c", "claim":
				return literatureParseClaim(args, c.li)
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

// literatureParseClaim は `c <組> <席×6>` を解釈する。
//
// **6 枚すべての所在を申告しなければ宣言にならない。**組の札順は
// LiteratureHalfSuitCards の並びに対応する。
func literatureParseClaim(args []string, li usecase.LiteratureInteractorIF) (string, bool) {
	if len(args) < 1+domain.LiteratureHalfSuitSize {
		return invalidArg("claimNeedsHalfSuitAndHolders"), true
	}
	half, err := strconv.Atoi(args[0])
	if err != nil || half < 0 || half >= domain.LiteratureHalfSuitCnt {
		return invalidArg("invalidHalfSuit0Max", "val", args[0], "max", strconv.Itoa(domain.LiteratureHalfSuitCnt-1)), true
	}
	holders := make([]int, 0, domain.LiteratureHalfSuitSize)
	for i := range domain.LiteratureHalfSuitSize {
		seat, err := strconv.Atoi(args[1+i])
		if err != nil || seat < 0 || seat >= domain.LiteraturePlayerCnt {
			return invalidArg("invalidSeat0Max", "val", args[1+i], "max", strconv.Itoa(domain.LiteraturePlayerCnt-1)), true
		}
		holders = append(holders, seat)
	}
	return li.Claim(half, holders), true
}
