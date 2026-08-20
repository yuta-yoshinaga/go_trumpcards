package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"

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
	game := domain.NewDefaultDoubt()
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
	return BuildCuiHelp(CuiHelpSpec{
		TitleKey:    "doubt.helpTitle",
		CommandKeys: []string{"doubt.helpPlay", "doubt.helpDoubt", "doubt.helpSkip", "doubt.helpLog"},
		SettingKeys: []string{"doubt.helpSetWindow", "doubt.helpSetMemory", "doubt.helpSetPenalty", "doubt.helpSetHesitation"},
	})
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
	// The Doubt loop reads plain (cooked) stdin and holds no terminal raw
	// mode or history file, so there is nothing to clean up on signal — but
	// it still needs a handler to exit on SIGINT/SIGTERM (issue #2096).
	defer runSignalWatcher(nil)()
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
// CPUがカードを出した場合: 設定秒数だけ人間のダウト入力を待ち、CPUダウターと合算して処理
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
		// Drop any keystrokes the user typed during the preceding CPU turn so
		// the doubt window only sees input entered after the prompt appears.
		cui.drainInput()
		windowSec := cui.game.GetConfig().DoubtWindowSec
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		ticks := make(chan struct{})
		done := make(chan struct{})
		defer close(done)
		go func() {
			defer close(ticks)
			for {
				select {
				case <-ticker.C:
					select {
					case ticks <- struct{}{}:
					case <-done:
						return
					}
				case <-done:
					return
				}
			}
		}()
		tty := term.IsTerminal(int(os.Stdout.Fd()))
		if runDoubtCountdown(os.Stdout, windowSec, cui.inputCh, ticks, tty) {
			allDoubters = append(allDoubters, humanIdx) // 人間がダウト
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

// runDoubtCountdown renders the human doubt-window prompt and waits for
// either an input on inputCh or windowSec ticks. Returns true when the
// caller typed "d"/"doubt" (first whitespace token) before the window
// expired, false on any other input or on timeout.
//
// In TTY mode the helper emits BEL (\a) once on entry so a user looking
// away from the terminal hears the alert, then rewrites the seconds
// indicator in place via \r + ANSI erase-to-end on every tick. In
// non-TTY mode (CI logs, piped output) it falls back to a single
// prompt line so the transcript stays readable. See issue #1862.
//
// The tick channel is injected so tests can drive the countdown
// deterministically; production code wires it to a 1s time.Ticker.
func runDoubtCountdown(w io.Writer, windowSec int, inputCh <-chan string, ticks <-chan struct{}, tty bool) bool {
	promptFor := func(sec int) string {
		return i18n.Tf("doubt.doubtPrompt", "sec", strconv.Itoa(sec))
	}

	if tty {
		// BEL + first prompt without a trailing newline so the next tick can
		// overwrite the same row.
		_, _ = fmt.Fprint(w, "\a"+promptFor(windowSec))
	} else {
		_, _ = fmt.Fprintln(w, promptFor(windowSec))
	}

	remaining := windowSec
	for {
		select {
		case input := <-inputCh:
			// In TTY mode the terminal's Enter echo already advanced the cursor
			// to a fresh row, so the helper does not emit its own newline; in
			// non-TTY mode the prompt was printed with a trailing newline up
			// front. strings.Fields already trims whitespace on both sides.
			fields := strings.Fields(input)
			return len(fields) > 0 && (fields[0] == "d" || fields[0] == "doubt")
		case <-ticks:
			remaining--
			if remaining <= 0 {
				if tty {
					_, _ = fmt.Fprintln(w)
				}
				_, _ = fmt.Fprintln(w, i18n.T("doubt.timeout"))
				return false
			}
			if tty {
				// \r repositions to column 0; \x1b[K clears anything left over
				// from a longer previous render (e.g., 10→9 loses a digit).
				_, _ = fmt.Fprint(w, "\r"+promptFor(remaining)+"\x1b[K")
			}
		}
	}
}
