package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strings"
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

// realtimeHelpCommand is the pseudo-command produced by the help key. Like
// realtimeQuitCommand it is handled by the loop rather than dispatched to the
// controller — redisplaying the key legend is not a game action. The value is
// not a real controller command, so it must never reach Exec.
const realtimeHelpCommand = "\x00help"

// SlapjackRealtimeKeyMap maps single-keystroke input to the controller
// command the realtime CUI runner should execute. Used by Slapjack and
// Egyptian Ratscrew (their controllers share the same command surface).
//
// Every entry must also appear in realtimeCommandLabelKeys, which is what the
// on-screen legend is built from — TestRealtimeLegendCoversKeyMap checks both
// directions. The `log` key sat here undocumented for months because the
// legend was hand-written prose (issue #5179).
var SlapjackRealtimeKeyMap = map[rune]string{
	' ': "j",
	's': "s",
	'S': "s",
	'r': "r",
	'R': "r",
	'q': realtimeQuitCommand,
	'Q': realtimeQuitCommand,
	'l': "log",
	'L': "log",
	'h': realtimeHelpCommand,
	'H': realtimeHelpCommand,
	'?': realtimeHelpCommand,
	// `sd <n>` cannot be typed one byte at a time, so the three difficulties get
	// their own keys. 1/2/3 rather than 0/1/2 so the keys read as
	// easy/normal/hard; the argument stays the domain's 0-based value.
	'1':  "sd 0",
	'2':  "sd 1",
	'3':  "sd 2",
	0x03: realtimeQuitCommand, // Ctrl+C
	0x04: realtimeQuitCommand, // Ctrl+D
}

// realtimeCommandLabelKeys maps each command reachable from a realtime key map
// to the i18n key describing it. The legend is rendered from this plus the key
// map, so a key added to one without the other fails the round-trip test rather
// than silently shipping an undocumented command.
var realtimeCommandLabelKeys = map[string]string{
	"j":                 "realtime.labelSlap",
	"s":                 "realtime.labelStep",
	"r":                 "realtime.labelReset",
	"log":               "realtime.labelLog",
	"sd 0":              "realtime.labelEasy",
	"sd 1":              "realtime.labelNormal",
	"sd 2":              "realtime.labelHard",
	realtimeHelpCommand: "realtime.labelHelp",
	realtimeQuitCommand: "realtime.labelQuit",
}

// realtimeKeyLabel renders a key for display: space and the control codes have
// no printable form, so they get spelled out.
func realtimeKeyLabel(k rune) string {
	switch k {
	case ' ':
		return i18n.T("realtime.keySpace")
	case 0x03:
		return "Ctrl+C"
	case 0x04:
		return "Ctrl+D"
	}
	return string(k)
}

// realtimeLegendLines renders the key legend from mapping: one line per
// command, listing every key bound to it in a stable order. Generated rather
// than written so it cannot drift from the key map.
func realtimeLegendLines(mapping map[rune]string) []string {
	keysByCommand := make(map[string][]rune, len(mapping))
	for k, cmd := range mapping {
		keysByCommand[cmd] = append(keysByCommand[cmd], k)
	}
	// Commands in a fixed display order; anything unlisted is appended sorted so
	// a new command still shows up (the round-trip test keeps the list honest).
	order := []string{"j", "s", "r", "sd 0", "sd 1", "sd 2", "log", realtimeHelpCommand, realtimeQuitCommand}
	seen := make(map[string]bool, len(order))
	for _, c := range order {
		seen[c] = true
	}
	rest := make([]string, 0, len(keysByCommand))
	for c := range keysByCommand {
		if !seen[c] {
			rest = append(rest, c)
		}
	}
	sort.Strings(rest)

	lines := make([]string, 0, len(keysByCommand)+1)
	lines = append(lines, i18n.T("realtime.banner"))
	for _, cmd := range append(order, rest...) {
		keys := keysByCommand[cmd]
		if len(keys) == 0 {
			continue
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		labels := make([]string, 0, len(keys))
		for _, k := range keys {
			labels = append(labels, realtimeKeyLabel(k))
		}
		labelKey, ok := realtimeCommandLabelKeys[cmd]
		if !ok {
			continue // guarded by TestRealtimeLegendCoversKeyMap
		}
		lines = append(lines, fmt.Sprintf("  %-14s %s", strings.Join(labels, ", "), i18n.T(labelKey)))
	}
	return lines
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
			if cmd == realtimeHelpCommand {
				// Loop-level, not a game action: the 200 ms tick scrolls the
				// legend off screen within seconds and there was no way to get
				// it back (issue #5179).
				for _, line := range realtimeLegendLines(mapping) {
					writeRealtimeOutput(w, line)
				}
				continue
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
// Returns 0 on normal exit and the line-mode loop's code (0 normally, 1 on
// non-EOF stdin error — issue #1839) when it falls back. The realtime loop
// itself only exits on user quit / signal / stdin EOF, all of which are
// clean shutdowns.
//
// Signal handling is local rather than via the package-level
// setupSignalHandler — the latter calls os.Exit on SIGINT/SIGTERM, which
// bypasses the deferred term.Restore and leaves the user's terminal in
// raw mode. Here we close a quit channel instead, letting the loop return
// normally and the deferred restore run.
func RunRealtimeCuiLoop(gameName string, controller CuiExecer, helpLines []string) int {
	stdinFd := int(os.Stdin.Fd())
	if !term.IsTerminal(stdinFd) {
		return RunCuiLoop(gameName, controller, helpLines)
	}
	state, err := term.MakeRaw(stdinFd)
	if err != nil {
		// Raw mode failed (rare — e.g., unusual shells); degrade to line mode.
		return RunCuiLoop(gameName, controller, helpLines)
	}
	defer func() { _ = term.Restore(stdinFd, state) }()

	// The line-mode helpLines are NOT shown here. They document `j`, `tick` and
	// `sd <n>`, none of which can be typed one byte at a time -- pressing `j`
	// did nothing at all, because unmapped keys are silently dropped. The
	// legend below is generated from the key map instead, so it can only ever
	// list keys that actually work. See issue #5179.
	for _, line := range realtimeLegendLines(SlapjackRealtimeKeyMap) {
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
	// Buffered so the reader goroutine never blocks waiting for us to
	// consume the final error after we've already returned from the core.
	errCh := make(chan error, 1)
	go readRealtimeKeys(os.Stdin, keys, errCh)

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
	// Pick up any non-EOF stdin error the reader recorded before the keys
	// channel closed. Non-blocking: EOF / signal exits leave errCh empty.
	select {
	case err := <-errCh:
		if err != nil {
			fmt.Fprintln(os.Stderr, i18n.Tf("inputReadError", "error", err.Error()))
			return 1
		}
	default:
	}
	fmt.Println(i18n.T("bye"))
	return 0
}

// readRealtimeKeys reads single bytes from r and forwards them as runes.
// Closes keys on EOF or unrecoverable error so the core loop exits, and
// sends the error on errCh for non-EOF failures so the caller can surface
// them as exit 1 instead of swallowing them. EOF sends nothing — it is a
// clean shutdown.
func readRealtimeKeys(r io.Reader, keys chan<- rune, errCh chan<- error) {
	defer close(keys)
	br := bufio.NewReader(r)
	for {
		b, err := br.ReadByte()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				// Non-blocking: errCh is buffered with cap 1 and is only
				// ever sent to here, so this select cannot block; the
				// default branch guards against any future redesign that
				// might bypass that invariant.
				select {
				case errCh <- err:
				default:
				}
			}
			return
		}
		keys <- rune(b)
	}
}
