package games_test

import (
	"path/filepath"
	"regexp"
	"testing"
)

// TestPasurSoorConditionAgreesAcrossTheUIs guards the "this take is a soor" test.
//
// The domain rule lives in `capturesWholeTable`: a soor takes every table index,
// leaving the table empty. The Web decides which button
// to badge from `option.length === state.table.length` and the CUI compares the
// hint's indices to the table (#5762). Both are restatements of the same domain
// condition, so a change there has to reach both — this fails when either side
// stops comparing against the whole table.
func TestPasurSoorConditionAgreesAcrossTheUIs(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	domainSrc := readFileForTest(t, filepath.Join(root, "internal", "domain", "Pasur.go"))
	domainGuard := regexp.MustCompile(`func \(p \*Pasur\) capturesWholeTable\(tableIndices \[\]int\) bool \{([\s\S]*?)\n\}`).FindStringSubmatch(domainSrc)
	if domainGuard == nil {
		t.Fatal("the domain no longer defines capturesWholeTable")
	}
	if !regexp.MustCompile(`len\(seen\) == len\(p\.tableCards\)`).MatchString(domainGuard[1]) {
		t.Fatal("capturesWholeTable no longer requires every table index to be captured")
	}

	web := readFileForTest(t, filepath.Join(root, "frontend", "src", "pages", "PasurPage.tsx"))
	if !regexp.MustCompile(`option\.length === state\.table\.length`).MatchString(web) {
		t.Error("PasurPage.tsx no longer badges the option that empties the table")
	}

	cui := readFileForTest(t, filepath.Join(root, "internal", "adapter", "presenter", "PasurCuiPresenter.go"))
	if !regexp.MustCompile(`len\(hint\.TableIndices\) == len\(s\.GetTableCards\(\)\)`).MatchString(cui) {
		t.Error("the CUI no longer marks a hint that empties the table")
	}
}
