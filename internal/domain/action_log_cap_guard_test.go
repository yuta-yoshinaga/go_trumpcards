//go:build test
// +build test

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

func TestActionLogCapGuard(t *testing.T) {
	violations := []string{}
	filesScanned := 0
	checks := []*regexp.Regexp{
		regexp.MustCompile(`\.actionLog = append\(`),
		regexp.MustCompile(`appendLogCodeAt\(len\(`),
		regexp.MustCompile(`TurnNumber:\s+len\(`),
		regexp.MustCompile(`^\s+actionLog\s+\[\]\*ActionLogEntry`),
	}
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") || name == "action_log_base.go" {
			continue
		}
		filesScanned++
		file, err := os.Open(name)
		if err != nil {
			t.Fatal(err)
		}
		scanner := bufio.NewScanner(file)
		lineNumber := 0
		for scanner.Scan() {
			lineNumber++
			line := scanner.Text()
			for _, check := range checks {
				if check.MatchString(line) {
					violations = append(violations, filepath.Base(name)+":"+strconv.Itoa(lineNumber))
					break
				}
			}
		}
		if err := scanner.Err(); err != nil {
			_ = file.Close()
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if filesScanned < 300 {
		t.Fatalf("scanned only %d non-test Go files; want at least 300", filesScanned)
	}
	if len(violations) > 0 {
		t.Fatalf("uncapped action log writes found: %s", strings.Join(violations, ", "))
	}
}
