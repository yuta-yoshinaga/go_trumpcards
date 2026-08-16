//go:build test

package controller_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// hintDispatchers are the helper calls that make "h"/"hint" reach the interactor.
// A controller may also answer it inline, or pass a hint func into a shared
// dispatcher struct, so the struct-field form counts too.
var hintDispatchers = []string{
	"dispatchHintAndLog",
	"dispatchResetHintAndLog",
	"hint:",
	`case "h", "hint"`,
	`case "hint"`,
}

var interactorRe = regexp.MustCompile(`usecase\.(\w+)InteractorIF`)

// hintMethodRe matches a Hint() declaration at a token boundary, so ToggleHint()
// and similar suffixes do not count as the interactor exposing Hint().
var hintMethodRe = regexp.MustCompile(`(^|[^\w])Hint\(\) string`)

// TestWebControllersDispatchHintWhenTheInteractorHasIt is a structural guard.
//
// The Web CLI sends "hint" to the REST endpoint, so the *Web* controller has to
// route it. That is a different file from the terminal *CuiController, and the
// distinction is not academic: #5791 shipped a CLI hint command for Durak and
// Crazy Eights whose Web controllers fell through to dispatchLog, which matches
// neither "h" nor "hint" -- the request came back 400 "Unsupported command.",
// strictly worse than the local rejection it replaced. Review caught it; nothing
// in the test suite did.
//
// The rule is keyed on capability rather than a list of game names: a Web
// controller must dispatch hint exactly when its own interactor exposes Hint().
// Controllers whose interactor has no Hint() are skipped without needing an
// allowlist entry, and start being checked the moment one is added.
func TestWebControllersDispatchHintWhenTheInteractorHasIt(t *testing.T) {
	files, err := filepath.Glob("*WebController.go")
	assert.NoError(t, err)
	assert.NotEmpty(t, files, "no Web controllers found -- the glob is wrong, not the code")

	capable, skipped := 0, 0
	var missing []string

	for _, path := range files {
		b, err := os.ReadFile(path) //nolint:gosec // test-only, fixed glob
		assert.NoError(t, err)
		src := string(b)

		m := interactorRe.FindStringSubmatch(src)
		if m == nil {
			skipped++
			continue
		}
		if !interactorExposesHint(t, m[1]) {
			skipped++
			continue
		}
		capable++

		found := false
		for _, d := range hintDispatchers {
			if strings.Contains(src, d) {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, path)
		}
	}

	// Negative control on the detection itself: if either counter collapses the
	// guard is inspecting nothing and would pass however broken the code was.
	assert.Greater(t, capable, 150,
		"far fewer hint-capable controllers than expected -- the detection is broken, not the code")
	assert.Greater(t, skipped, 0,
		"expected some controllers whose interactor has no Hint()")
	assert.Empty(t, missing,
		"these Web controllers have an interactor with Hint() but never dispatch it, "+
			"so the Web CLI's hint command would 400")
}

// interactorExposesHint reports whether <name>InteractorIF declares Hint().
func interactorExposesHint(t *testing.T, name string) bool {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "usecase", name+"Interactor.go")) //nolint:gosec // derived from a matched Go identifier
	if err != nil {
		return false
	}
	body := string(b)
	i := strings.Index(body, name+"InteractorIF interface")
	if i < 0 {
		return false
	}
	end := strings.Index(body[i:], "\n}")
	if end < 0 {
		return false
	}
	// Must be the method Hint, not a method whose name merely ends in it:
	// BlackJackInteractorIF has ToggleHint() string, and a substring match
	// reported it as hint-capable, which sent a "fix" at a method that does
	// not exist and broke the build.
	return hintMethodRe.MatchString(body[i : i+end])
}
