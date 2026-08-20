//go:build !js || !wasm || extra4

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DragonTigerCuiController ドラゴンタイガーCUIコントローラークラス
type DragonTigerCuiController struct {
	di usecase.DragonTigerInteractorIF
	// lastBet は直近に受け付けたベット。`rb` はこれを reset の後に打ち直す (#5585)。
	//
	// **Web は 1 クリックで同じ賭けを繰り返せる** (`dt-rebet-button`) のに、CUI は
	// 毎ラウンド `r` のあとフルの `b <額> <種別>` を打ち直す必要があった。
	lastBet *dragonTigerBet
}

// dragonTigerBet は再賭けのために覚えておく 1 回分のベット。
type dragonTigerBet struct {
	amount  int
	betType int
}

// NewDragonTigerCuiController コンストラクタ
func NewDragonTigerCuiController(di usecase.DragonTigerInteractorIF) *DragonTigerCuiController {
	return &DragonTigerCuiController{di: di}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100 d", "b 100 t", "b 100 e", "clear", "log", "q"
//   - dragon: d / dragon / 0
//   - tiger:  t / tiger  / 1
//   - tie:    e / tie    / 2
func (dc *DragonTigerCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return dc.di.Reset() },
		[]string{"b", "bet", "rb", "rebet", "clear", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				if len(args) < 2 {
					return invalidArg("betAmountAndTypeRequired"), true
				}
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args[:1], "betAmountRequired", "invalidBetAmount", domain.DragonTigerMinBet, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				betType, ok := dragonTigerParseBetType(args[1])
				if !ok {
					return invalidArg("invalidBetTypeDragonTiger"), true
				}
				// **受け付けた後に覚える。**額や種別が不正なまま覚えると、
				// `rb` が通らないベットを繰り返す。
				dc.lastBet = &dragonTigerBet{amount: amount, betType: betType}
				return dc.di.Bet(amount, betType), true
			case "rb", "rebet":
				return dc.handleRebet(), true
			case "clear":
				return dc.di.ClearHistory(), true
			default:
				return handleCuiLog(cmd, dc.di.ActionLog)
			}
		},
	)
}

// handleRebet は直前と同じ賭けを、リセットの後に打ち直す。
//
// Web の `handleRebet` と同じ順序 (reset → bet)。履歴が無ければ、黙って
// 何もせずにエラーを返す ── 「リセットだけされた」状態は、打ち直したつもりの
// プレイヤーには気づけない。
func (dc *DragonTigerCuiController) handleRebet() string {
	if dc.lastBet == nil {
		return invalidArg("dragontiger.noPreviousBet")
	}
	dc.di.Reset()
	return dc.di.Bet(dc.lastBet.amount, dc.lastBet.betType)
}

// dragonTigerParseBetType ベットタイプを文字列から解析する。
func dragonTigerParseBetType(arg string) (int, bool) {
	switch arg {
	case "d", "dragon", "0":
		return domain.DragonTigerBetDragon, true
	case "t", "tiger", "1":
		return domain.DragonTigerBetTiger, true
	case "e", "tie", "2":
		return domain.DragonTigerBetTie, true
	default:
		return 0, false
	}
}
