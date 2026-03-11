package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DoubtCui ダウトCUIクラス
type DoubtCui struct {
	dc      *controller.DoubtCuiController
	game    interfaces.DoubtGame // read-only: phase/state inspection
	inputCh chan string
}

// NewDoubtCui コンストラクタ
func NewDoubtCui() *DoubtCui {
	players := []*domain.DoubtPlayer{
		domain.NewDoubtPlayer(true),
		domain.NewDoubtPlayer(false),
		domain.NewDoubtPlayer(false),
		domain.NewDoubtPlayer(false),
	}
	game := domain.NewDoubt(domain.NewTrumpCards(0), players)
	dc := controller.NewDoubtCuiController(
		usecase.NewDoubtInteractor(game, new(presenter.DoubtCuiPresenter)),
	)
	return &DoubtCui{
		dc:      dc,
		game:    game,
		inputCh: make(chan string, 10),
	}
}

// Controller returns the game controller.
func (cui *DoubtCui) Controller() CuiExecer { return cui.dc }

// HelpLines returns the game's help lines.
func (cui *DoubtCui) HelpLines() []string {
	return []string{
		i18n.T("doubt.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("doubt.helpPlay"),
		i18n.T("doubt.helpDoubt"),
		i18n.T("doubt.helpSkip"),
		"",
		i18n.T("settings"),
		i18n.T("doubt.helpSetWindow"),
		i18n.T("doubt.helpSetMemory"),
		i18n.T("doubt.helpSetPenalty"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
}

// inputReader 標準入力を読み込み inputCh に送るゴルーチン
func (cui *DoubtCui) inputReader() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		cui.inputCh <- scanner.Text()
	}
}

// drainInput inputCh に残っている入力を捨てる (ダウト待機ウィンドウ後のクリア)
func (cui *DoubtCui) drainInput() {
	for {
		select {
		case <-cui.inputCh:
		default:
			return
		}
	}
}

// Exec ゲームメインループ
func (cui *DoubtCui) Exec() {
	go cui.inputReader()
	fmt.Println(cui.dc.Exec("r"))
	fmt.Println(i18n.T("typeHelp"))
	for !cui.game.GetGameEndFlag() {
		switch {
		case cui.game.GetPhase() == domain.DoubtPhasePlay && cui.game.IsHumanTurn():
			cui.drainInput()
			input := <-cui.inputCh
			trimmed := strings.TrimSpace(input)
			if trimmed == "help" || trimmed == "?" {
				for _, line := range cui.HelpLines() {
					fmt.Println(line)
				}
				continue
			}
			res := cui.dc.Exec(input)
			if res == i18n.QuitSentinel {
				fmt.Println(i18n.T("bye"))
				return
			}
			fmt.Println(res)
		case cui.game.GetPhase() == domain.DoubtPhaseDoubt:
			cui.handleDoubtWindow()
		}
	}
}

// handleDoubtWindow ダウト待機ウィンドウを処理する
// 人間がカードを出した場合: CPUダウターのみ即時処理
// CPUがカードを出した場合: 10秒間人間のダウト入力を待ち、CPUダウターと合算して処理
func (cui *DoubtCui) handleDoubtWindow() {
	lastAction := cui.game.GetLastAction()
	if lastAction == nil {
		// ダウトフェーズだが lastAction がない場合はスキップ
		res := cui.dc.Exec("s")
		fmt.Println(res)
		return
	}

	cpuDoubters := cui.game.GetCpuDoubters()

	// 人間プレイヤーのインデックスを動的に取得
	humanIdx := -1
	for i := 0; i < cui.game.GetPlayerCnt(); i++ {
		if cui.game.GetPlayer(i).GetIsHuman() {
			humanIdx = i
			break
		}
	}
	humanCanDoubt := humanIdx >= 0 && lastAction.PlayerIdx != humanIdx // 人間はカード出しプレイヤーをダウトできない

	var allDoubters []int

	if humanCanDoubt {
		timeout := time.Duration(cui.game.GetConfig().DoubtWindowSec) * time.Second
		fmt.Println(i18n.Tf("doubtPrompt", "sec", fmt.Sprintf("%d", cui.game.GetConfig().DoubtWindowSec)))
		select {
		case input := <-cui.inputCh:
			trimmed := strings.TrimSpace(input)
			fields := strings.Fields(trimmed)
			if len(fields) > 0 && (fields[0] == "d" || fields[0] == "doubt") {
				allDoubters = append(allDoubters, humanIdx) // 人間がダウト
			}
		case <-time.After(timeout):
			fmt.Println(i18n.T("timeout"))
		}
	}

	allDoubters = append(allDoubters, cpuDoubters...)

	if len(allDoubters) > 0 {
		args := "d"
		for _, idx := range allDoubters {
			args += fmt.Sprintf(" %d", idx)
		}
		res := cui.dc.Exec(args)
		fmt.Println(res)
	} else {
		res := cui.dc.Exec("s")
		fmt.Println(res)
	}
}
