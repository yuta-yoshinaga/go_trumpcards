//go:build test

package controller

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestUnknownCommandMessage(t *testing.T) {
	t.Run("no suggestion", func(t *testing.T) {
		assert.Equal(t, i18n.ErrorPrefix+"コマンドが不明です: foo", unknownCommandMessage("foo", nil))
	})
	t.Run("empty command", func(t *testing.T) {
		assert.Equal(t, i18n.ErrorPrefix+"コマンドが不明です: ", unknownCommandMessage("", nil))
	})
	t.Run("with suggestion", func(t *testing.T) {
		assert.Equal(t, i18n.ErrorPrefix+"コマンドが不明です: hti。もしかして 'hit' ですか？", unknownCommandMessage("hti", []string{"hit", "stand"}))
	})
	t.Run("no close match", func(t *testing.T) {
		assert.Equal(t, i18n.ErrorPrefix+"コマンドが不明です: zzzzzzz", unknownCommandMessage("zzzzzzz", []string{"hit", "stand"}))
	})
}

func TestExecCuiCommand(t *testing.T) {
	resetFn := func(args []string) string {
		if len(args) > 0 {
			return "reset:" + args[0]
		}
		return "reset"
	}
	validCmds := []string{"g", "game"}
	gameHandler := func(cmd string, args []string) (string, bool) {
		if cmd == "g" {
			return "game", true
		}
		return "", false
	}

	t.Run("empty input", func(t *testing.T) {
		result := execCuiCommand("", resetFn, validCmds, gameHandler)
		assert.Equal(t, "'help' でコマンド一覧を表示します。", result)
	})

	t.Run("whitespace only input", func(t *testing.T) {
		result := execCuiCommand("   ", resetFn, validCmds, gameHandler)
		assert.Equal(t, "'help' でコマンド一覧を表示します。", result)
	})

	t.Run("q command", func(t *testing.T) {
		assert.Equal(t, "bye.", execCuiCommand("q", resetFn, validCmds, gameHandler))
	})

	t.Run("quit command", func(t *testing.T) {
		assert.Equal(t, "bye.", execCuiCommand("quit", resetFn, validCmds, gameHandler))
	})

	t.Run("exit command", func(t *testing.T) {
		assert.Equal(t, "bye.", execCuiCommand("exit", resetFn, validCmds, gameHandler))
	})

	t.Run("r command without args", func(t *testing.T) {
		assert.Equal(t, "reset", execCuiCommand("r", resetFn, validCmds, gameHandler))
	})

	t.Run("reset command without args", func(t *testing.T) {
		assert.Equal(t, "reset", execCuiCommand("reset", resetFn, validCmds, gameHandler))
	})

	t.Run("r command with args", func(t *testing.T) {
		assert.Equal(t, "reset:tunnel", execCuiCommand("r tunnel", resetFn, validCmds, gameHandler))
	})

	t.Run("handled game command", func(t *testing.T) {
		assert.Equal(t, "game", execCuiCommand("g", resetFn, validCmds, gameHandler))
	})

	t.Run("unhandled game command with suggestion", func(t *testing.T) {
		result := execCuiCommand("gam", resetFn, validCmds, gameHandler)
		assert.Equal(t, i18n.ErrorPrefix+"コマンドが不明です: gam。もしかして 'game' ですか？", result)
	})

	t.Run("unhandled game command no suggestion", func(t *testing.T) {
		result := execCuiCommand("zzzzzzz", resetFn, validCmds, gameHandler)
		assert.Equal(t, i18n.ErrorPrefix+"コマンドが不明です: zzzzzzz", result)
	})
}

// execSolitaireCui consolidates 6 byte-identical Exec bodies across the
// tableau solitaires (BakersDozen, BeleagueredCastle, Bisley, FlowerGarden,
// KingAlbert, StreetsAndAlleys). They differed only in receiver and in which
// interactor the closures called — see issue #5368.
func TestExecSolitaireCui(t *testing.T) {
	fns := func(calls *[]string) solitaireCuiFns {
		rec := func(name string) func() string {
			return func() string { *calls = append(*calls, name); return name }
		}
		return solitaireCuiFns{
			reset:        rec("reset"),
			giveUp:       rec("giveup"),
			autoComplete: rec("autocomplete"),
			undo:         rec("undo"),
			hint:         rec("hint"),
			actionLog:    rec("log"),
			move: func(args []string) string {
				*calls = append(*calls, "move:"+strings.Join(args, ","))
				return "moved"
			},
		}
	}

	cases := []struct{ cmd, want string }{
		{"r", "reset"}, {"reset", "reset"},
		{"g", "giveup"}, {"giveup", "giveup"},
		{"ac", "autocomplete"}, {"autocomplete", "autocomplete"},
		{"u", "undo"}, {"undo", "undo"},
		{"h", "hint"}, {"hint", "hint"},
		{"l", "log"}, {"log", "log"},
	}
	for _, tc := range cases {
		t.Run(tc.cmd, func(t *testing.T) {
			var calls []string
			execSolitaireCui(tc.cmd, fns(&calls))
			assert.Equal(t, []string{tc.want}, calls)
		})
	}

	t.Run("move forwards its arguments", func(t *testing.T) {
		var calls []string
		got := execSolitaireCui("m w 1", fns(&calls))
		assert.Equal(t, "moved", got)
		assert.Equal(t, []string{"move:w,1"}, calls)
	})

	// An unknown command must reach the shared suggestion path rather than
	// being answered here: that is what produces "もしかして…" for a typo.
	t.Run("unknown command is not answered by any of the callbacks", func(t *testing.T) {
		var calls []string
		out := execSolitaireCui("frobnicate", fns(&calls))
		assert.Empty(t, calls, "no interactor method should have run")
		assert.NotEmpty(t, out, "the shared helper still answers with a suggestion")
	})
}

func TestSolitaireCuiCommandNamesMatchMigratedControllers(t *testing.T) {
	cases := []struct {
		name          string
		oldCommands   []string
		extraCommands []string
	}{
		{"Alaska", []string{"m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"}, nil},
		{"AmericanToad", []string{"d", "draw", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"}, []string{"d", "draw"}},
		{"BigBen", []string{"d", "deal", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"}, []string{"d", "deal"}},
		{"Braid", []string{"d", "draw", "dir", "direction", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"}, []string{"d", "draw", "dir", "direction"}},
		{"Colorado", []string{"d", "draw", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"}, []string{"d", "draw"}},
		{"Congress", []string{"d", "draw", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"}, []string{"d", "draw"}},
		{"CrazyQuilt", []string{"d", "draw", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"}, []string{"d", "draw"}},
		{"Cruel", []string{"m", "move", "s", "shift", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"}, []string{"s", "shift"}},
		{"Diplomat", []string{"d", "draw", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"}, []string{"d", "draw"}},
		{"Duchess", []string{"b", "base", "d", "draw", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"}, []string{"b", "base", "d", "draw"}},
		{"Easthaven", []string{"m", "move", "d", "deal", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"}, []string{"d", "deal"}},
		{"EightOff", []string{"m", "move", "f", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"}, []string{"f"}},
		{"FreeCell", []string{"m", "move", "f", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"}, []string{"f"}},
		{"GrandfathersClock", []string{"m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo", "redo"}, []string{"redo"}},
		{"Matrimony", []string{"d", "draw", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"}, []string{"d", "draw"}},
		{"MissMilligan", []string{"d", "deal", "m", "move", "wv", "waive", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"}, []string{"d", "deal", "wv", "waive"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			extra := make(map[string]func([]string) string, len(tc.extraCommands))
			for _, name := range tc.extraCommands {
				extra[name] = func([]string) string { return "" }
			}
			fns := solitaireCuiFns{extraCommands: extra}
			assert.ElementsMatch(t, tc.oldCommands, solitaireCuiCommandNames(fns))
		})
	}

	// Exercise each real Exec as well: this catches a command accidentally
	// omitted from a controller's extraCommands map.
	execs := map[string]func(string) string{
		"Alaska":            (&AlaskaCuiController{ri: new(mockusecase.MockAlaskaInteractor)}).Exec,
		"AmericanToad":      (&AmericanToadCuiController{ai: new(mockusecase.MockAmericanToadInteractor)}).Exec,
		"BigBen":            (&BigBenCuiController{gi: new(mockusecase.MockBigBenInteractor)}).Exec,
		"Braid":             (&BraidCuiController{bi: new(mockusecase.MockBraidInteractor)}).Exec,
		"Colorado":          (&ColoradoCuiController{ci: new(mockusecase.MockColoradoInteractor)}).Exec,
		"Congress":          (&CongressCuiController{ci: new(mockusecase.MockCongressInteractor)}).Exec,
		"CrazyQuilt":        (&CrazyQuiltCuiController{ci: new(mockusecase.MockCrazyQuiltInteractor)}).Exec,
		"Cruel":             (&CruelCuiController{ci: new(mockusecase.MockCruelInteractor)}).Exec,
		"Diplomat":          (&DiplomatCuiController{ci: new(mockusecase.MockDiplomatInteractor)}).Exec,
		"Duchess":           (&DuchessCuiController{di: new(mockusecase.MockDuchessInteractor)}).Exec,
		"Easthaven":         (&EasthavenCuiController{ei: new(mockusecase.MockEasthavenInteractor)}).Exec,
		"EightOff":          (&EightOffCuiController{ei: new(mockusecase.MockEightOffInteractor)}).Exec,
		"FreeCell":          (&FreeCellCuiController{fi: new(mockusecase.MockFreeCellInteractor)}).Exec,
		"GrandfathersClock": (&GrandfathersClockCuiController{gi: new(mockusecase.MockGrandfathersClockInteractor)}).Exec,
		"Matrimony":         (&MatrimonyCuiController{ci: new(mockusecase.MockMatrimonyInteractor)}).Exec,
		"MissMilligan":      (&MissMilliganCuiController{mi: new(mockusecase.MockMissMilliganInteractor)}).Exec,
	}
	for _, tc := range cases {
		t.Run(tc.name+" Exec candidates", func(t *testing.T) {
			for _, command := range tc.oldCommands {
				out := execs[tc.name](command + "x")
				assert.Contains(t, out, "'"+command+"'", "command %q should remain a typo suggestion", command)
			}
		})
	}
}
