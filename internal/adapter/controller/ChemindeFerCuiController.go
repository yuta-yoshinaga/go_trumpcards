//go:build !js || !wasm || extra4

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ChemindeFerCuiController シュマン・ド・フェールCUIコントローラークラス
type ChemindeFerCuiController struct {
	ci usecase.ChemindeFerInteractorIF
}

// NewChemindeFerCuiController コンストラクタ
func NewChemindeFerCuiController(ci usecase.ChemindeFerInteractorIF) *ChemindeFerCuiController {
	return &ChemindeFerCuiController{ci: ci}
}

// Exec ゲーム実行
// コマンド例: "r", "stake 200", "bet 50", "draw", "stand", "pass", "next", "hint", "log", "q"
//   - draw / stand は**手番の側**に効きます (子の判断中なら子、親の判断中なら親)
//   - bet 0 は「降りる」
func (cc *ChemindeFerCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"stake", "bet", "draw", "stand", "pass", "next", "giveup", "hint", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "stake":
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args, "stakeRequired", "invalidStakeANumber", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return cc.ci.SetStake(amount), true
			case "bet":
				// **0 は「降りる」。** 下限を 0 にしておかないと降りられない。
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args, "betAmountRequired", "invalidBetANumber", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return cc.ci.PlaceBet(chemindeFerHumanSeat, amount), true
			case "draw":
				return cc.drawOrStand(args, true), true
			case "stand":
				return cc.drawOrStand(args, false), true
			case "pass":
				return cc.ci.PassBank(), true
			case "next":
				return cc.ci.NextRound(), true
			case "giveup":
				return cc.ci.GiveUp(), true
			case "hint":
				return cc.ci.Hint(), true
			default:
				return handleCuiLog(cmd, cc.ci.ActionLog)
			}
		},
	)
}

// drawOrStand は "draw" / "stand" を親と子のどちらの操作として送るか決める。
//
// **CUI では側を明示させない。** 手番でない側の操作はドメインが弾くので、両方を
// 順に試して通ったほうを返せばよい。ここで側を推測して 1 つだけ送ると、実装が
// フェーズを二重に持つことになってドメインとずれる。
// 明示したい場合は "draw banker" / "stand punter" のように側を書ける。
func (cc *ChemindeFerCuiController) drawOrStand(args []string, draw bool) string {
	side := ""
	if len(args) > 0 {
		side = args[0]
	}
	punter, banker := cc.ci.PunterDraw, cc.ci.BankerDraw
	if !draw {
		punter, banker = cc.ci.PunterStand, cc.ci.BankerStand
	}
	switch side {
	case "p", "punter":
		return punter()
	case "b", "banker":
		return banker()
	default:
		return cc.ci.DrawOrStand(draw)
	}
}
