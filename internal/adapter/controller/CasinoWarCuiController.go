//go:build !js || !wasm || extra4

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CasinoWarCuiController カジノウォーCUIコントローラークラス
type CasinoWarCuiController struct {
	ci usecase.CasinoWarInteractorIF
	// lastBet は直近に受け付けたベット額。`rb` はこれを reset の後に打ち直す (#6379)。
	//
	// **Web は 1 クリックで同じ賭けを繰り返せる** (`cw-rebet-button`) のに、CUI は
	// 毎ラウンド `r` のあとフルの `b <額>` を打ち直す必要があった。
	lastBet *int
}

// NewCasinoWarCuiController コンストラクタ
func NewCasinoWarCuiController(ci usecase.CasinoWarInteractorIF) *CasinoWarCuiController {
	return &CasinoWarCuiController{ci: ci}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "rb", "rebet", "surrender", "war", "log", "q"
func (cc *CasinoWarCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"b", "bet", "rb", "rebet", "surrender", "war", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args, "betAmountRequired", "invalidBetAmount", domain.CasinoWarMinBet, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				// **受け付けた後に覚える。**額が不正なまま覚えると、
				// `rb` が通らないベットを繰り返す。
				cc.lastBet = &amount
				return cc.ci.Bet(amount), true
			case "rb", "rebet":
				return cc.handleRebet(), true
			case "surrender":
				return cc.ci.Surrender(), true
			case "war":
				return cc.ci.War(), true
			default:
				return handleCuiLog(cmd, cc.ci.ActionLog)
			}
		},
	)
}

// handleRebet は直前と同じ賭けを、リセットの後に打ち直す。
//
// Web の `handleRebet` と同じ順序 (reset → bet)。履歴が無ければ、黙って
// 何もせずにエラーを返す ── 「リセットだけされた」状態は、打ち直したつもりの
// プレイヤーには気づけない。
func (cc *CasinoWarCuiController) handleRebet() string {
	if cc.lastBet == nil {
		return invalidArg("casinowar.noPreviousBet")
	}
	cc.ci.Reset()
	return cc.ci.Bet(*cc.lastBet)
}
