//go:build test

package ui

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// helpRawKeyRe matches a help line that is nothing but an i18n key, which is
// what i18n.T returns when the key has no translation.
var helpRawKeyRe = regexp.MustCompile(`^[a-z0-9]+\.[a-zA-Z][a-zA-Z0-9]*$`)

// TestCuiHelpRendersNoRawKeys fails when a game's help shows an untranslated
// key instead of a command row.
//
// Ten games shipped this way: bezique, courtpiece, ecarte, ganjifa, pitch,
// sevencardstudhilo, tarneeb, teenpatti, threecardbrag and vira rendered 58
// lines that read `vira.helpBid`, `pitch.helpPlay` and so on -- their whole
// command table was unreadable, in both languages, and `tarneeb` had no locale
// file at all. Nothing caught it: the parity checker compares ja against en, and
// these were missing from both, so the two agreed perfectly.
func TestCuiHelpRendersNoRawKeys(t *testing.T) {
	registry := GameRegistry()
	if len(registry) < 300 {
		t.Fatalf("only %d games in the registry -- the walk broke", len(registry))
	}

	original := i18n.Lang()
	t.Cleanup(func() { i18n.SetLang(original) })

	checked := 0
	var bad []string
	for _, lang := range []string{"ja", "en"} {
		i18n.SetLang(lang)
		for _, entry := range registry {
			for _, line := range entry.NewCui().HelpLines() {
				checked++
				if s := strings.TrimSpace(line); helpRawKeyRe.MatchString(s) {
					bad = append(bad, lang+"/"+entry.Name+": "+s)
				}
			}
		}
	}
	if checked < 300 {
		t.Fatalf("only %d help lines were read -- the walk broke", checked)
	}
	sort.Strings(bad)
	if len(bad) > 0 {
		t.Errorf("help lines rendering an untranslated i18n key (%d of %d lines):\n  %s",
			len(bad), checked, strings.Join(bad, "\n  "))
	}
}

// TestCuiHelpRawKeyPatternMatchesAKey pins the pattern, because a regex that
// stopped matching would make the guard above report zero for every possible
// state of the locale files.
func TestCuiHelpRawKeyPatternMatchesAKey(t *testing.T) {
	for _, s := range []string{"vira.helpBid", "pitch.helpPlay", "tarneeb.helpTitle"} {
		if !helpRawKeyRe.MatchString(s) {
			t.Errorf("%q should look like an untranslated key", s)
		}
	}
	for _, s := range []string{"b <n>", "bid (0=pass, 7-13)", "Tarneeb", "スエカ (Sueca)", ""} {
		if helpRawKeyRe.MatchString(s) {
			t.Errorf("%q should NOT look like an untranslated key", s)
		}
	}
}

// helpVerbIsUnknown reports whether ctrl answered `verb` with its
// unknown-command message rather than dispatching it.
func helpVerbIsUnknown(out, verb string) bool {
	body, _ := i18n.StripErrorPrefix(out)
	full := i18n.Tf("unknownCommand", "cmd", verb)
	// The suggestion variant appends to the same prefix, so match on the head
	// rather than the whole string.
	head := strings.SplitN(full, ":", 2)[0] + ": " + verb
	return strings.Contains(body, full) || strings.HasPrefix(body, head)
}

// TestCuiHelpTableVerbsDispatch is the direction TestCuiHelpCommandKeysCoverLocale
// does not check: that one starts from the parser and asks whether the table
// lists it, this one starts from the table and asks whether the parser answers.
// A row documenting a command the controller never had would pass the former.
func TestCuiHelpTableVerbsDispatch(t *testing.T) {
	registry := GameRegistry()
	if len(registry) < 300 {
		t.Fatalf("only %d games in the registry -- the walk broke", len(registry))
	}

	// Positive control first: if the detector cannot recognise a command that is
	// certainly unknown, "0 undispatchable" below means nothing.
	detected := 0
	for _, entry := range registry {
		c := entry.NewCui().Controller()
		c.Exec("r")
		if helpVerbIsUnknown(c.Exec("zzbogusverb"), "zzbogusverb") {
			detected++
		}
	}
	if detected != len(registry) {
		t.Fatalf("the unknown-command detector recognised a bogus verb in only %d of %d games; "+
			"fix helpVerbIsUnknown rather than trusting the result below", detected, len(registry))
	}

	checked := 0
	var bad []string
	for _, entry := range registry {
		g := entry.NewCui()
		commands, _ := helpSections(g.HelpLines())
		ctrl := g.Controller()
		ctrl.Exec("r")
		for _, line := range commands {
			verb := helpLineVerb(line)
			if verb == "" {
				continue
			}
			checked++
			if helpVerbIsUnknown(ctrl.Exec(verb), verb) {
				bad = append(bad, entry.Name+": documents `"+verb+"`, which its controller answers as unknown")
			}
		}
	}
	if checked < len(registry) {
		t.Fatalf("only %d command verbs were read for %d games -- the walk broke", checked, len(registry))
	}
	sort.Strings(bad)
	if len(bad) > 0 {
		t.Errorf("documented commands the parser does not have (%d of %d):\n  %s",
			len(bad), checked, strings.Join(bad, "\n  "))
	}
}
