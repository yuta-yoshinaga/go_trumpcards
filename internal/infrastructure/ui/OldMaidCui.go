package ui

import (
	"bufio"
	"fmt"
	"os"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OldMaidCui ババ抜きCUIクラス
type OldMaidCui struct {
	omc *controller.OldMaidCuiController
}

// NewOldMaidCui コンストラクタ
func NewOldMaidCui() *OldMaidCui {
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(true),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
	}
	oldMaid := domain.NewOldMaid(domain.NewTrumpCards(1), players)
	return &OldMaidCui{
		omc: controller.NewOldMaidCuiController(
			usecase.NewOldMaidInteractor(oldMaid, presenter.NewOldMaidCuiPresenter()),
		),
	}
}

// Exec ゲーム実行
func (cui *OldMaidCui) Exec() {
	fmt.Println(cui.omc.Exec("r"))
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("コマンドを入力してください。")
		fmt.Println("q・・・quit")
		fmt.Println("r・・・reset")
		fmt.Println("d・・・draw (カードを引く)")
		if !scanner.Scan() {
			break
		}
		res := cui.omc.Exec(scanner.Text())
		fmt.Println(res)
		if res == "bye." {
			break
		}
	}
}
