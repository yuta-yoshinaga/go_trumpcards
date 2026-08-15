//go:build !js || !wasm || extra

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TrenteEtQuaranteCuiController はトラント・エ・カラント (Trente et Quarante) の
// CUI コントローラークラス。
type TrenteEtQuaranteCuiController struct {
	bi usecase.TrenteEtQuaranteInteractorIF
}

// NewTrenteEtQuaranteCuiController コンストラクタ。
func NewTrenteEtQuaranteCuiController(bi usecase.TrenteEtQuaranteInteractorIF) *TrenteEtQuaranteCuiController {
	return &TrenteEtQuaranteCuiController{bi: bi}
}

// Exec コマンド実行。
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	b / bet <t> <stake>      → ベット種別 t (0=Noir,1=Rouge,2=Couleur,3=Inverse) と
//	                           ステークを賭けてラウンドを解決する
//	nr / nextround / n       → 次のラウンドへ
//	sb / setdefaultbet <0-3> → デフォルトベット種別設定
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *TrenteEtQuaranteCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.bi.GetConfig()
			return c.bi.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bet", "n", "next", "nr", "nextround",
			"sb", "setdefaultbet", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				return c.handleBet(args), true
			case "n", "next", "nr", "nextround":
				return c.bi.NextRound(), true
			case "sb", "setdefaultbet":
				return cuiutil.WithParsedIntKeys(args, "defaultBetRequired0Noir1Rouge2Couleur3Inverse", "invalidBet03", 0, 3, func(v int) string {
					cfg := c.bi.GetConfig()
					cfg.DefaultBet = domain.TrenteEtQuaranteBet(v)
					return c.bi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.bi.Hint, c.bi.ActionLog)
			}
		},
	)
}

// handleBet は "b <betType> <stake>" を解析して Bet を呼ぶ。
func (c *TrenteEtQuaranteCuiController) handleBet(args []string) string {
	if len(args) < 2 {
		return invalidArg("betTypeAndStakeRequired")
	}
	bet, err := strconv.Atoi(args[0])
	if err != nil || bet < int(domain.TrenteEtQuaranteBetNoir) || bet > int(domain.TrenteEtQuaranteBetInverse) {
		return "Invalid bet type: " + args[0] + " (0=Noir, 1=Rouge, 2=Couleur, 3=Inverse)."
	}
	stake, err := strconv.Atoi(args[1])
	if err != nil {
		return "Invalid stake: " + args[1] + "."
	}
	return c.bi.Bet(domain.TrenteEtQuaranteBet(bet), stake)
}
