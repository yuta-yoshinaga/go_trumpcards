package domain

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestHelperPlacementGuard(t *testing.T) {
	root := filepath.Join("..")
	declaration := regexp.MustCompile(`^func (\([^)]*\) )?\w+(Public|ForTest)\(`)
	violations := make([]string, 0)
	goFiles := 0

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		goFiles++
		name := filepath.Base(path)
		if strings.HasSuffix(name, "_test.go") || strings.HasSuffix(name, "_testhelpers.go") {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(file)
		for line := 1; scanner.Scan(); line++ {
			content := scanner.Text()
			if declaration.MatchString(content) {
				violations = append(violations, filepath.Clean(path)+":"+strconv.Itoa(line)+":"+content)
			}
		}
		scanErr := scanner.Err()
		if closeErr := file.Close(); scanErr == nil {
			scanErr = closeErr
		}
		return scanErr
	})
	if err != nil {
		t.Fatal(err)
	}
	if goFiles < 100 {
		t.Errorf("walked %d Go files under internal/, want at least 100", goFiles)
	}
	for _, violation := range violations {
		t.Errorf("%s: test-only helpers must live in <Game>_testhelpers.go with //go:build test (#8020)", violation)
	}
}
