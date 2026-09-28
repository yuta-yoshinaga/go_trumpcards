//go:build test

package domain_test

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const mapRangeUnreviewedLimit = 0
const mapRangeMinimumCount = 100

type mapRangeEntry struct {
	file, function, expression, verdict, reason string
}

func TestMapRangeGuard(t *testing.T) {
	entries := enumerateMapRanges(t)
	if len(entries) < mapRangeMinimumCount {
		t.Fatalf("enumerated only %d map ranges; scan may be broken", len(entries))
	}

	allowlistPath := filepath.Join("testdata", "map_range_allowlist.tsv")
	allowlist, err := readMapRangeAllowlist(allowlistPath)
	if err != nil {
		if os.Getenv("GOLDEN_UPDATE") != "1" || !os.IsNotExist(err) {
			t.Fatal(err)
		}
		allowlist = make(map[string]mapRangeEntry)
	}
	if os.Getenv("GOLDEN_UPDATE") == "1" {
		for key, entry := range entries {
			if old, ok := allowlist[key]; ok {
				entry.verdict, entry.reason = old.verdict, old.reason
			} else if entry.file == "map_order.go" && entry.function == "sortedIntKeys" {
				entry.verdict = "independent"
				entry.reason = "キーを集めてから sort.Ints で並べる"
			} else {
				entry.verdict = "unreviewed"
			}
			entries[key] = entry
		}
		if err := writeMapRangeAllowlist(allowlistPath, entries); err != nil {
			t.Fatal(err)
		}
		t.Logf("enumerated %d map ranges; unreviewed %d", len(entries), countUnreviewed(entries))
		return
	}

	var problems []string
	for key, entry := range entries {
		allowed, ok := allowlist[key]
		if !ok {
			problems = append(problems, fmt.Sprintf("new map iteration: %s\t%s\t%s; iterate in ascending order with sortedIntKeys or add it to the allowlist with a reason it is order-independent", entry.file, entry.function, entry.expression))
			continue
		}
		if allowed.verdict != "independent" && allowed.verdict != "unreviewed" {
			problems = append(problems, fmt.Sprintf("invalid verdict %q for %s", allowed.verdict, key))
		}
		if allowed.verdict == "independent" && strings.TrimSpace(allowed.reason) == "" {
			problems = append(problems, fmt.Sprintf("independent entry has no reason: %s", key))
		}
	}
	for key := range allowlist {
		if _, ok := entries[key]; !ok {
			problems = append(problems, "stale allowlist entry; remove it: "+key)
		}
	}
	unreviewed := countUnreviewed(allowlist)
	if unreviewed > mapRangeUnreviewedLimit {
		problems = append(problems, fmt.Sprintf("unreviewed map ranges %d exceeds limit %d", unreviewed, mapRangeUnreviewedLimit))
	}
	t.Logf("enumerated %d map ranges; unreviewed %d", len(entries), unreviewed)
	if len(problems) > 0 {
		sort.Strings(problems)
		t.Fatalf("map range guard found problems:\n  %s", strings.Join(problems, "\n  "))
	}
}

func enumerateMapRanges(t *testing.T) map[string]mapRangeEntry {
	t.Helper()
	const root = "."
	fset := token.NewFileSet()
	dirEntries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read domain directory: %v", err)
	}
	packages := make(map[string][]*ast.File)
	for _, entry := range dirEntries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(root, name), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		packages[file.Name.Name] = append(packages[file.Name.Name], file)
	}
	entries := make(map[string]mapRangeEntry)
	for _, files := range packages {
		info := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}
		conf := types.Config{Importer: importer.ForCompiler(fset, "source", nil), Error: func(error) {}}
		_, _ = conf.Check("domain", fset, files, info)
		for _, file := range files {
			function := ""
			ast.Inspect(file, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.FuncDecl:
					previous := function
					function = n.Name.Name
					if n.Recv != nil && len(n.Recv.List) > 0 {
						function = types.ExprString(n.Recv.List[0].Type) + "." + function
					}
					if n.Body != nil {
						ast.Inspect(n.Body, func(child ast.Node) bool {
							if child == nil {
								return true
							}
							if rs, ok := child.(*ast.RangeStmt); ok {
								if tv, ok := info.Types[rs.X]; ok && tv.Type != nil {
									if _, isMap := tv.Type.Underlying().(*types.Map); isMap {
										fileName := filepath.Base(fset.Position(rs.Pos()).Filename)
										expr := types.ExprString(rs.X)
										entryExpr := expr
										key := fileName + "\t" + function + "\t" + entryExpr
										for duplicate := 2; ; duplicate++ {
											if _, exists := entries[key]; !exists {
												break
											}
											entryExpr = fmt.Sprintf("%s #%d", expr, duplicate)
											key = fileName + "\t" + function + "\t" + entryExpr
										}
										entries[key] = mapRangeEntry{file: fileName, function: function, expression: entryExpr}
									}
								}
							}
							return true
						})
					}
					function = previous
					return false
				}
				return true
			})
		}
	}
	return entries
}

func readMapRangeAllowlist(path string) (map[string]mapRangeEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	entries := make(map[string]mapRangeEntry)
	scanner := bufio.NewScanner(file)
	for line := 1; scanner.Scan(); line++ {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) != 5 {
			return nil, fmt.Errorf("%s:%d: want 5 tab-separated columns", path, line)
		}
		entry := mapRangeEntry{fields[0], fields[1], fields[2], fields[3], fields[4]}
		key := strings.Join(fields[:3], "\t")
		if _, exists := entries[key]; exists {
			return nil, fmt.Errorf("%s:%d: duplicate key %s", path, line, key)
		}
		entries[key] = entry
	}
	return entries, scanner.Err()
}

func writeMapRangeAllowlist(path string, entries map[string]mapRangeEntry) error {
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	for _, key := range keys {
		entry := entries[key]
		if _, err := fmt.Fprintf(file, "%s\t%s\t%s\t%s\t%s\n", entry.file, entry.function, entry.expression, entry.verdict, entry.reason); err != nil {
			return err
		}
	}
	if err := file.Close(); err != nil {
		return err
	}
	return nil
}

func countUnreviewed(entries map[string]mapRangeEntry) int {
	count := 0
	for _, entry := range entries {
		if entry.verdict == "unreviewed" {
			count++
		}
	}
	return count
}
