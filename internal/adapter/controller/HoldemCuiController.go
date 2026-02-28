package controller

import (
	"strconv"
	"strings"

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
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return "コマンドが不明です: " + command
	}
	switch fields[0] {
	case "q", "quit":
		return "bye."
	case "r", "reset":
		return c.hi.Reset()
	case "f", "fold":
		return c.hi.Action(domain.HoldemActionFold, 0)
	case "ck", "check":
		return c.hi.Action(domain.HoldemActionCheck, 0)
	case "c", "call":
		return c.hi.Action(domain.HoldemActionCall, 0)
	case "b", "bet":
		amount, err := parseAmount(fields)
		if err != "" {
			return err
		}
		return c.hi.Action(domain.HoldemActionBet, amount)
	case "ra", "raise":
		amount, err := parseAmount(fields)
		if err != "" {
			return err
		}
		return c.hi.Action(domain.HoldemActionRaise, amount)
	case "a", "allin":
		return c.hi.Action(domain.HoldemActionAllIn, 0)
	default:
		return "コマンドが不明です: " + command
	}
}

// parseAmount コマンドのスライスからベット額を抽出する
// 戻り値: (金額, エラーメッセージ)。エラーメッセージが空なら成功。
func parseAmount(fields []string) (int, string) {
	if len(fields) < 2 {
		return 0, "ベット/レイズには金額の指定が必要です。"
	}
	amount, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, "無効な金額です: " + fields[1]
	}
	return amount, ""
}
