package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// goMapKeyRe は presenter が map リテラルで組む応答のキー。
//
// **タグだけ見ると誤検知する。** MonteBank のヒントは
// `map[string]any{"pickIdx": ..., "reason": ...}` で組まれていて、`json:` タグは
// どこにも無い。それでも本物のフィールドなので、ここも数える。
var goMapKeyRe = regexp.MustCompile(`"([A-Za-z][A-Za-z0-9_]*)"\s*:`)

// openapiStructuralKeys は payload ではなく OpenAPI 自身の語彙。
var openapiStructuralKeys = map[string]bool{
	"schema": true, "content": true, "properties": true, "items": true,
	"type": true, "description": true, "example": true, "enum": true,
	"required": true, "nullable": true, "additionalProperties": true,
	"allOf": true, "oneOf": true, "anyOf": true, "format": true,
	// リクエストの共通項目。どのゲームのコントローラにも構造体としては無い。
	"command": true, "sessionId": true,
}

// TestOpenAPIDeclaresNothingTheCodeCannotProduce は**逆向き**の検査。
//
// #7048 で「返っているのに書かれていない」を 0 にした。その逆 ──
// **書かれているのに返さない**項目が 23 件あった (#7050)。クライアントを
// 生成すると存在しないフィールドの取り出しコードができるし、仕様書を読んだ人は
// 無いものを探す。
//
// たとえば Bridge は `vulnerable` と書いてあったが、実際に返るのは
// `vulnerability`。**名前だけ違う**ので「その名前がどこかにあるか」を見る
// #6984 のガードでも、「返ってきた項目」を見る #7048 のガードでも出ない。
//
// 実際に返るかどうかは配りに依る (`omitempty`) ので、ここでは**静的に**見る:
// spec が宣言した名前が、`json:` タグにも presenter の map リテラルにも
// 無ければ、コードにその項目を作る手段が無い。
func TestOpenAPIDeclaresNothingTheCodeCannotProduce(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	raw, err := os.ReadFile(filepath.Join(root, "api", "openapi.yaml")) //nolint:gosec // test-only, fixed path
	if err != nil {
		t.Fatalf("openapi.yaml が読めない: %v", err)
	}
	var spec liveSpec
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		t.Fatalf("openapi.yaml が YAML として壊れている: %v", err)
	}

	produced := map[string]bool{}
	goFiles := 0
	walkErr := filepath.Walk(filepath.Join(root, "internal"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		src, readErr := os.ReadFile(path) //nolint:gosec // test-only, repo-relative
		if readErr != nil {
			return readErr
		}
		goFiles++
		for _, m := range jsonTagRe.FindAllStringSubmatch(string(src), -1) {
			produced[m[1]] = true
		}
		for _, m := range goMapKeyRe.FindAllStringSubmatch(string(src), -1) {
			produced[m[1]] = true
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("internal/ が歩けない: %v", walkErr)
	}
	// **空振りしない床。** 集められていないと「全部実在する」に見える。
	if goFiles < 500 || len(produced) < 1000 {
		t.Fatalf("Go ファイル %d 本 / 名前 %d 件しか集まっていない。走査が壊れている", goFiles, len(produced))
	}

	declared := map[string]bool{}
	for path, ops := range spec.Paths {
		if gameExecPathRe.FindStringSubmatch(path) == nil {
			continue
		}
		collectNames(&spec, ops.Post.Responses.OK.Content.JSON.Schema, declared, 0)
		collectNames(&spec, ops.Post.RequestBody.Content.JSON.Schema, declared, 0)
	}
	if len(declared) < 500 {
		t.Fatalf("spec から拾えた項目名が %d 件しかない", len(declared))
	}

	var ghosts []string
	for name := range declared {
		if !produced[name] && !openapiStructuralKeys[name] {
			ghosts = append(ghosts, name)
		}
	}
	sort.Strings(ghosts)
	if len(ghosts) > 0 {
		t.Errorf("コードが作れない項目を spec が宣言している (%d 件): %v\n"+
			"名前違いなら正しい名前に直し、実在しないなら消すこと。", len(ghosts), ghosts)
	}
}

// collectNames は spec の項目名を集める。requestBody 側も同じ形で歩く。
func collectNames(spec *liveSpec, sch *liveSchema, out map[string]bool, depth int) {
	sch = spec.deref(sch)
	if sch == nil || depth > liveMaxDepth {
		return
	}
	for name, sub := range sch.Properties {
		out[name] = true
		collectNames(spec, sub, out, depth+1)
	}
	for _, group := range [][]*liveSchema{sch.AllOf, sch.OneOf, sch.AnyOf} {
		for _, sub := range group {
			collectNames(spec, sub, out, depth+1)
		}
	}
	collectNames(spec, sch.Items, out, depth+1)
	collectNames(spec, sch.AdditionalProperties.Schema, out, depth+1)
}
