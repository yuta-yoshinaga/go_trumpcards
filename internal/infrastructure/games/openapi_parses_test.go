package games_test

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestOpenAPISpecParses is the guard that was missing when a hand-edited
// insertion split a folded `description: >-` block and left the file invalid.
//
// Nothing else reads api/openapi.yaml as YAML: the drift guards match it with
// regexes, which keep working on a broken file, so a malformed spec shipped
// green (#5757 review). Parsing it is the cheapest check that the document is
// still a document.
func TestOpenAPISpecParses(t *testing.T) {
	path := filepath.Join("..", "..", "..", "api", "openapi.yaml")
	data, err := os.ReadFile(path) //nolint:gosec // test-only, fixed path
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("api/openapi.yaml is not valid YAML: %v", err)
	}

	// 空振り防止: 読めてはいるが中身が空、を成功と読ませない。
	for _, key := range []string{"openapi", "paths", "components"} {
		if _, ok := doc[key]; !ok {
			t.Errorf("the spec has no %q section", key)
		}
	}
	paths, ok := doc["paths"].(map[string]any)
	if !ok || len(paths) < 100 {
		t.Errorf("the spec declares %d paths, far fewer than expected", len(paths))
	}
}
