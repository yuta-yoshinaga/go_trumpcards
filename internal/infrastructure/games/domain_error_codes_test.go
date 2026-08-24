package games_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// domainErrorCodeRe captures the message-code literal handed to
// NewDomainErrorCode. The key has to be a literal at the call site for this
// guard (and check-message-codes.mjs on the frontend side) to see it at all.
var domainErrorCodeRe = regexp.MustCompile(`NewDomainErrorCode\([^,]+,\s*"([^"]+)"`)

// TestDomainErrorCodesResolveInGoLocales asserts every message code raised by
// the domain has a Japanese *and* an English translation in internal/i18n.
//
// **The CUI prints the code itself when the key is missing.** cuiErrorBlock
// translates via i18n.Tf, and i18n.T returns its argument for an unknown key,
// so a forgotten entry ships as a player-visible `<game>.errSomething` line
// with no error, no warning and every unit test still green — the frontend
// guard (check-message-codes.mjs) only ever looks at
// frontend/src/i18n/locales/{ja,en}/common.json, never at the Go locales.
func TestDomainErrorCodesResolveInGoLocales(t *testing.T) {
	codes := collectDomainErrorCodes(t)
	if len(codes) < 50 {
		t.Fatalf("only %d domain error codes found — the regex has stopped matching", len(codes))
	}

	locales := map[string]map[string]map[string]string{}
	for _, lang := range []string{"ja", "en"} {
		locales[lang] = map[string]map[string]string{}
	}

	var missing []string
	for _, code := range codes {
		ns, key, ok := strings.Cut(code, ".")
		if !ok {
			t.Errorf("%q is not namespaced as <game>.<key>", code)
			continue
		}
		for _, lang := range []string{"ja", "en"} {
			table, err := loadGoLocale(locales[lang], lang, ns)
			if err != nil {
				missing = append(missing, fmt.Sprintf("%s [%s]: %v", code, lang, err))
				continue
			}
			if _, found := table[key]; !found {
				missing = append(missing, fmt.Sprintf("%s [%s]", code, lang))
			}
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("%d domain error code(s) have no Go translation and print raw to the CUI:\n  %s",
			len(missing), strings.Join(missing, "\n  "))
	}
}

// collectDomainErrorCodes returns every message code literal in internal/domain.
func collectDomainErrorCodes(t *testing.T) []string {
	t.Helper()
	dir := filepath.Join(repoRoot, "internal", "domain")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	seen := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		for _, m := range domainErrorCodeRe.FindAllSubmatch(data, -1) {
			seen[string(m[1])] = true
		}
	}
	codes := make([]string, 0, len(seen))
	for c := range seen {
		codes = append(codes, c)
	}
	sort.Strings(codes)
	return codes
}

// loadGoLocale reads internal/i18n/locales/<lang>/<ns>.json, caching by name.
func loadGoLocale(cache map[string]map[string]string, lang, ns string) (map[string]string, error) {
	if table, ok := cache[ns]; ok {
		return table, nil
	}
	path := filepath.Join(repoRoot, "internal", "i18n", "locales", lang, ns+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no locale file %s", filepath.Join("internal/i18n/locales", lang, ns+".json"))
	}
	var table map[string]string
	if err := json.Unmarshal(data, &table); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	cache[ns] = table
	return table, nil
}
