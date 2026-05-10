package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// realtimeQuitCommand is the controller command produced when the user
// presses the quit key. The realtime loop short-circuits on this value
// rather than dispatching it to the controller — it represents loop exit,
// not a game action.
const realtimeQuitCommand = "q"

// realtimeTickCommand is the controller command produced on each tick
// goroutine pulse. Slapjack and Egyptian Ratscrew interactor expose this
// to advance CPU pending actions; the runner is generic over the command
// name so other realtime games can substitute their own.
const realtimeTickCommand = "tick"

// SlapjackRealtimeTickInterval is how often the tick goroutine pulses.
// 200 ms is fast enough to feel snappy and slow enough to leave room for
// human reaction time without flooding the terminal with redraws.
const SlapjackRealtimeTickInterval = 200 * time.Millisecond

// SlapjackRealtimeKeyMap maps single-keystroke input to the controller
// command the realtime CUI runner should execute. Used by Slapjack and
// Egyptian Ratscrew (their controllers share the same command surface).
//
//   - space → "j" (slap) — the headline real-time action
//   - s/S   → "s" (step) — flip your top card to the pile
//   - r/R   → "r" (reset) — start a fresh game
//   - q/Q   → "q" (quit) — exit the realtime loop
//   - l/L   → "log" (action log)
var SlapjackRealtimeKeyMap = map[rune]string{
	' ':  "j",
	's':  "s",
	'S':  "s",
	'r':  "r",
	'R':  "r",
	'q':  realtimeQuitCommand,
	'Q':  realtimeQuitCommand,
	'l':  "log",
	'L':  "log",
	0x03: realtimeQuitCommand, // Ctrl+C
	0x04: realtimeQuitCommand, // Ctrl+D
}

// realtimeCuiCore drives the realtime CUI loop without touching the
// terminal or starting goroutines. Tests feed deterministic key/tick
// channels; the production runner wires raw stdin and a time.Ticker on
// top. Loop exits when the keys channel closes (stdin EOF), the quit
// channel closes (signal received), or a key maps to realtimeQuitCommand.
//
// The initial reset ("r") is dispatched once on entry so the player sees
// a fresh game without having to type anything.
func realtimeCuiCore(execer CuiExecer, keys <-chan rune, ticks <-chan struct{}, quit <-chan struct{}, w io.Writer, mapping map[rune]string) {
	writeRealtimeOutput(w, execer.Exec("r"))
	for {
		select {
		case k, ok := <-keys:
			if !ok {
				// stdin closed: nobody can quit interactively; exit so
				// the production loop can restore the terminal.
				return
			}
			cmd, mapped := mapping[k]
			if !mapped {
				continue
			}
			if cmd == realtimeQuitCommand {
				return
			}
			writeRealtimeOutput(w, execer.Exec(cmd))
		case _, ok := <-ticks:
			if !ok {
				// Ticker stopped: continue serving keys until they close
				// too. Disable the ticks branch so the closed channel
				// doesn't busy-loop the select.
				ticks = nil
				continue
			}
			writeRealtimeOutput(w, execer.Exec(realtimeTickCommand))
		case <-quit:
			return
		}
	}
}

// writeRealtimeOutput prints out unless empty. The realtime loop dispatches
// "tick" repeatedly and the interactor frequently returns "" when no CPU
// action was pending; suppressing those keeps the screen readable.
func writeRealtimeOutput(w io.Writer, out string) {
	if out == "" {
		return
	}
	_, _ = fmt.Fprintln(w, out)
}

// RunRealtimeCuiLoop runs the realtime CUI loop for Slapjack and Egyptian
// Ratscrew. When stdin is not a TTY (piped input, CI, WSL2 with redirected
// stdio), it falls back to the standard line-mode loop so non-interactive
// scripts keep working.
//
// Signal handling is local rather than via the package-level
// setupSignalHandler — the latter calls os.Exit on SIGINT/SIGTERM, which
// bypasses the deferred term.Restore and leaves the user's terminal in
// raw mode. Here we close a quit channel instead, letting the loop return
// normally and the deferred restore run.
func RunRealtimeCuiLoop(gameName string, controller CuiExecer, helpLines []string) {
	stdinFd := int(os.Stdin.Fd())
	if !term.IsTerminal(stdinFd) {
		RunCuiLoop(gameName, controller, helpLines)
		return
	}
	state, err := term.MakeRaw(stdinFd)
	if err != nil {
		// Raw mode failed (rare — e.g., unusual shells); degrade to line mode.
		RunCuiLoop(gameName, controller, helpLines)
		return
	}
	defer func() { _ = term.Restore(stdinFd, state) }()

	fmt.Println(i18n.T("realtime.banner"))
	for _, line := range helpLines {
		fmt.Println(line)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	quit := make(chan struct{})
	go func() {
		<-sigCh
		close(quit)
	}()

	keys := make(chan rune)
	go readRealtimeKeys(os.Stdin, keys)

	ticker := time.NewTicker(SlapjackRealtimeTickInterval)
	defer ticker.Stop()
	ticks := make(chan struct{}, 1)
	done := make(chan struct{})
	go func() {
		defer close(ticks)
		for {
			select {
			case <-ticker.C:
				// Non-blocking send so a busy core (e.g., printing a long
				// log) cannot back up a tick queue and trigger a flurry of
				// stale ticks once it catches up. Reflex gameplay only
				// cares about the next tick, not missed ones.
				select {
				case ticks <- struct{}{}:
				default:
				}
			case <-done:
				return
			}
		}
	}()

	realtimeCuiCore(controller, keys, ticks, quit, os.Stdout, SlapjackRealtimeKeyMap)
	close(done)
	fmt.Println(i18n.T("bye"))
}

// readRealtimeKeys reads single bytes from r and forwards them as runes.
// Closes keys on EOF or unrecoverable error so the core loop exits.
func readRealtimeKeys(r io.Reader, keys chan<- rune) {
	defer close(keys)
	br := bufio.NewReader(r)
	for {
		b, err := br.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			return
		}
		keys <- rune(b)
	}
}
