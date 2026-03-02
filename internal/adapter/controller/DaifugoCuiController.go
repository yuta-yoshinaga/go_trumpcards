package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DaifugoCuiController 大富豪CUIコントローラークラス
type DaifugoCuiController struct {
	dgi usecase.DaifugoInteractorIF
}

// NewDaifugoCuiController コンストラクタ
func NewDaifugoCuiController(dgi usecase.DaifugoInteractorIF) *DaifugoCuiController {
	return &DaifugoCuiController{dgi: dgi}
}

// Exec コマンド実行
// play コマンドは "p 0 2" または "play 0 2" の形式でカードインデックスを指定。
// インデックスなしの場合はパス扱い。例: "p" → パス / "p 0 2" → 0番と2番のカードを出す
func (c *DaifugoCuiController) Exec(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return "コマンドが不明です: " + command
	}
	switch fields[0] {
	case "q", "quit":
		return "bye."
	case "r", "reset":
		return c.dgi.Reset()
	case "p", "play":
		indices := []int{}
		for _, f := range fields[1:] {
			if idx, err := strconv.Atoi(f); err == nil {
				indices = append(indices, idx)
			}
		}
		return c.dgi.Play(indices)
	case "sort":
		mode := domain.DaifugoSortByStrength
		if len(fields) > 1 {
			if m, err := strconv.Atoi(fields[1]); err == nil && m >= int(domain.DaifugoSortByStrength) && m <= int(domain.DaifugoSortByNumber) {
				mode = domain.DaifugoSortMode(m)
			}
		}
		return c.dgi.Sort(mode)
	default:
		return "コマンドが不明です: " + command
	}
}
