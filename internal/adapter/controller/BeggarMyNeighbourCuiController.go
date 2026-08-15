package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BeggarMyNeighbourCuiController Beggar-My-Neighbour CUIコントローラークラス
type BeggarMyNeighbourCuiController struct {
	wi usecase.BeggarMyNeighbourInteractorIF
}

// NewBeggarMyNeighbourCuiController コンストラクタ
func NewBeggarMyNeighbourCuiController(wi usecase.BeggarMyNeighbourInteractorIF) *BeggarMyNeighbourCuiController {
	return &BeggarMyNeighbourCuiController{wi: wi}
}

// Exec コマンド実行
//
//	q / quit        → ゲーム終了
//	r / reset       → ゲームリセット
//	s / step        → 次のカードを出す / ペナルティを払う / 場を回収する
//	a / autoplay    → 決着まで自動で進める
//	sm / setmax <n> → MaxRounds を設定してリセット
//	log / l         → 棋譜表示
func (c *BeggarMyNeighbourCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.wi.GetConfig()
			return c.wi.ResetWithConfig(cfg)
		},
		[]string{"s", "step", "a", "autoplay", "sm", "setmax", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "s", "step":
				return c.wi.Step(), true
			case "a", "autoplay":
				return c.wi.AutoPlay(), true
			case "sm", "setmax":
				val, errMsg, ok := cuiutil.ParseOptionalIntKeys(args, 0, domain.BeggarMyNeighbourDefaultMaxRounds, "invalidMaxRounds")
				if !ok {
					return errMsg, true
				}
				cfg := c.wi.GetConfig()
				cfg.MaxRounds = val
				return c.wi.ResetWithConfig(cfg), true
			case "log", "l":
				return c.wi.ActionLog(), true
			default:
				return "", false
			}
		},
	)
}
