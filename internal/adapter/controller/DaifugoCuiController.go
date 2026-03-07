package controller

import (
	"strconv"

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
	return execCuiCommand(
		command,
		func(_ []string) string { return c.dgi.Reset() },
		unknownCommandMessage,
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				indices := []int{}
				for _, f := range args {
					if idx, err := strconv.Atoi(f); err == nil {
						indices = append(indices, idx)
					}
				}
				return c.dgi.Play(indices), true
			case "sort":
				mode := domain.DaifugoSortByStrength
				if len(args) > 0 {
					if m, err := strconv.Atoi(args[0]); err == nil && m >= int(domain.DaifugoSortByStrength) && m <= int(domain.DaifugoSortByNumber) {
						mode = domain.DaifugoSortMode(m)
					}
				}
				return c.dgi.Sort(mode), true
			}
			return "", false
		},
	)
}
