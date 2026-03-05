package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HoldemCuiController テキサスホールデムCUIコントローラークラス
type HoldemCuiController struct {
	hi usecase.HoldemInteractorIF
}

// NewHoldemCuiController コンストラクタ
func NewHoldemCuiController(hi usecase.HoldemInteractorIF) *HoldemCuiController {
	return &HoldemCuiController{hi: hi}
}

// Exec コマンド実行
func (c *HoldemCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.hi.Reset() },
		func(command string) string { return "コマンドが不明です: " + command },
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "f", "fold":
				return c.hi.Action(domain.HoldemActionFold, 0), true
			case "ck", "check":
				return c.hi.Action(domain.HoldemActionCheck, 0), true
			case "c", "call":
				return c.hi.Action(domain.HoldemActionCall, 0), true
			case "b", "bet":
				amount, err := parseAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.hi.Action(domain.HoldemActionBet, amount), true
			case "ra", "raise":
				amount, err := parseAmount(args)
				if err != nil {
					return err.Error(), true
				}
				return c.hi.Action(domain.HoldemActionRaise, amount), true
			case "a", "allin":
				return c.hi.Action(domain.HoldemActionAllIn, 0), true
			}
			return "", false
		},
	)
}

// parseAmount 引数スライスからベット額を抽出する
func parseAmount(args []string) (int, error) {
	if len(args) < 1 {
		return 0, fmt.Errorf("ベット/レイズには金額の指定が必要です。")
	}
	amount, err := strconv.Atoi(args[0])
	if err != nil || amount <= 0 {
		return 0, fmt.Errorf("無効な金額です: %s", args[0])
	}
	return amount, nil
}
