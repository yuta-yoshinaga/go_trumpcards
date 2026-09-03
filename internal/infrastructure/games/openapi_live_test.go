//go:build !js || !wasm

package games_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
)

// liveProbeCommands は各ゲームに投げるコマンド。
//
// **`reset` だけでは全部の形は出ない。** ヒントの塊も、ラウンドの集計も、
// 棋譜も、打ってからでないと返らない。実際 `log` の `entries` は 317 ゲームで
// 一度も書かれていなかったが、`reset` しか投げない検査では永久に見つからない。
var liveProbeCommands = []string{"reset", "hint", "n", "next", "nr", "log"}

// TestOpenAPIMatchesLiveResponses は**実際に返る JSON**とスキーマを突き合わせる。
//
// 構造体の `json:` タグを見る `TestOpenAPISchemaReachesEveryField` では
// 原理的に見えないものが 2 種類ある (#7048):
//
//   - **埋め込み構造体。** `WebOutputBase` の `messageCode` は各ゲームの
//     コントローラのファイルに現れないので、タグの走査に掛からない。実測で
//     47 ゲームが未記載だった
//   - **共有スキーマの取りこぼし。** `Card` は `glyph` / `label` / `color` /
//     `deck` を返すのに書かれていなかった (非 52 枚デッキの手続き描画に使う)
//
// 応答とスキーマを**並べて降りる**ので、名前ではなく位置で見る。
// `additionalProperties: true` のノードはそこで打ち切る ── 宣言どおり
// 何を入れてもよい場所なので、欠落ではない。
func TestOpenAPIMatchesLiveResponses(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	raw, err := os.ReadFile(filepath.Join(root, "api", "openapi.yaml")) //nolint:gosec // test-only, fixed path
	if err != nil {
		t.Fatalf("openapi.yaml が読めない: %v", err)
	}
	var spec liveSpec
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		t.Fatalf("openapi.yaml が YAML として壊れている: %v", err)
	}

	all := games.All()
	// **空振りしない床。** 0 件でも「全部一致」に見えてしまう。
	if len(all) < 300 {
		t.Fatalf("ゲームが %d 件しか登録されていない", len(all))
	}

	var problems []string
	checked := 0
	for _, g := range all {
		path, ok := spec.Paths["/"+g.Name+"/exec"]
		if !ok {
			t.Fatalf("%s の経路が openapi.yaml に無い", g.Name)
		}
		sch := path.Post.Responses.OK.Content.JSON.Schema
		ctrl := g.NewWebController()
		gaps := map[string]bool{}
		for _, cmd := range liveProbeCommands {
			body := probe(t, ctrl, g.Name, cmd)
			if body == nil {
				continue
			}
			checked++
			spec.walk(sch, body, "", gaps, 0)
		}
		ctrl.Stop()
		if len(gaps) > 0 {
			keys := make([]string, 0, len(gaps))
			for k := range gaps {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			problems = append(problems, fmt.Sprintf("%s: %s", g.Name, strings.Join(keys, ", ")))
		}
	}
	if checked < 300 {
		t.Fatalf("応答が %d 件しか取れていない。叩き方が壊れている", checked)
	}
	if len(problems) > 0 {
		sort.Strings(problems)
		t.Errorf("実際に返っているのにスキーマに無い項目が %d ゲームにある:\n  %s\n"+
			"レスポンスに項目を足したら api/openapi.yaml にも同じコミットで書くこと。",
			len(problems), strings.Join(problems, "\n  "))
	}
}

// joinKeys は型の集合を読める形にする。
func joinKeys(m map[string]bool) string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return strings.Join(out, "|")
}

// probe は 1 コマンドを叩いて応答を返す。JSON でなければ nil。
func probe(t *testing.T, ctrl games.WebController, name, cmd string) map[string]any {
	t.Helper()
	payload := fmt.Sprintf(`{"command":%q,"sessionId":"live-%s"}`, cmd, name)
	req := httptest.NewRequest(http.MethodPost, "/"+name+"/exec", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctrl.Exec(rec, req)
	var body map[string]any
	if json.Unmarshal(rec.Body.Bytes(), &body) != nil {
		return nil
	}
	return body
}

// --- openapi.yaml の必要な部分だけを読む型 ---

type liveSpec struct {
	Paths      map[string]livePath `yaml:"paths"`
	Components struct {
		Schemas map[string]*liveSchema `yaml:"schemas"`
	} `yaml:"components"`
}

type livePath struct {
	Post struct {
		Responses struct {
			OK struct {
				Content struct {
					JSON struct {
						Schema *liveSchema `yaml:"schema"`
					} `yaml:"application/json"`
				} `yaml:"content"`
			} `yaml:"200"`
		} `yaml:"responses"`
	} `yaml:"post"`
}

type liveSchema struct {
	Type                 string                 `yaml:"type"`
	Ref                  string                 `yaml:"$ref"`
	Properties           map[string]*liveSchema `yaml:"properties"`
	Items                *liveSchema            `yaml:"items"`
	AllOf                []*liveSchema          `yaml:"allOf"`
	OneOf                []*liveSchema          `yaml:"oneOf"`
	AnyOf                []*liveSchema          `yaml:"anyOf"`
	AdditionalProperties liveAdditional         `yaml:"additionalProperties"`
}

// liveAdditional は `additionalProperties` がスキーマにも真偽値にもなるのを受ける。
type liveAdditional struct {
	Open   bool
	Schema *liveSchema
}

// UnmarshalYAML は真偽値とスキーマの両方を読む。
func (a *liveAdditional) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return node.Decode(&a.Open)
	}
	var s liveSchema
	if err := node.Decode(&s); err != nil {
		return err
	}
	a.Schema = &s
	return nil
}

// liveMaxDepth は降りる深さの上限。相互参照で止まらなくなるのを防ぐ。
const liveMaxDepth = 6

// jsonTypeOf は実際の値の JSON 型を返す。
func jsonTypeOf(v any) string {
	switch t := v.(type) {
	case bool:
		return "boolean"
	case float64:
		// **JSON に整数型は無い。** 整数値は `integer` として報告し、
		// `number` と宣言された側で受け入れる (下の typeAllows)。
		if t == float64(int64(t)) {
			return "integer"
		}
		return "number"
	case string:
		return "string"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	}
	return "null"
}

// typeAllows は宣言した型が実際の値を許すか。
//
// **`number` は整数値も許す。** JSON の数値型は 1 つしか無いので、float の
// フィールドがたまたま整数で返るのは食い違いではない。
func typeAllows(want, got string) bool {
	return want == got || (want == "number" && got == "integer")
}

// types は sch とその選択肢が許す型を集める。
func (s *liveSpec) types(sch *liveSchema) map[string]bool {
	out := map[string]bool{}
	if sch.Type != "" {
		out[sch.Type] = true
	}
	for _, group := range [][]*liveSchema{sch.AllOf, sch.OneOf, sch.AnyOf} {
		for _, sub := range group {
			if d := s.deref(sub); d != nil {
				for k := range s.types(d) {
					out[k] = true
				}
			}
		}
	}
	return out
}

// deref は `$ref` を解決する。
func (s *liveSpec) deref(n *liveSchema) *liveSchema {
	for i := 0; n != nil && n.Ref != "" && i < liveMaxDepth; i++ {
		n = s.Components.Schemas[strings.TrimPrefix(n.Ref, openAPIRefPrefix)]
	}
	return n
}

// props は自分の properties と、allOf/oneOf/anyOf が足すものを合わせて返す。
//
// **oneOf も数える。** どれか 1 つを満たせば妥当なので、選択肢のどれかに
// 載っていれば記載済み。allOf しか見ないと Gaps の `grid`
// (`oneOf: [Card, null]`) を未記載と誤って報告する。
func (s *liveSpec) props(sch *liveSchema) map[string]*liveSchema {
	out := map[string]*liveSchema{}
	for k, v := range sch.Properties {
		out[k] = v
	}
	for _, group := range [][]*liveSchema{sch.AllOf, sch.OneOf, sch.AnyOf} {
		for _, sub := range group {
			d := s.deref(sub)
			if d == nil {
				continue
			}
			for k, v := range d.Properties {
				if _, ok := out[k]; !ok {
					out[k] = v
				}
			}
		}
	}
	return out
}

// walk は応答とスキーマを並べて降り、スキーマに無いキーを gaps に積む。
func (s *liveSpec) walk(sch *liveSchema, body any, path string, gaps map[string]bool, depth int) {
	sch = s.deref(sch)
	if sch == nil || depth > liveMaxDepth {
		return
	}
	switch v := body.(type) {
	case map[string]any:
		if sch.AdditionalProperties.Open {
			return // 宣言どおり何でも入る
		}
		props := s.props(sch)
		for k, sub := range v {
			switch {
			case props[k] != nil:
				// **名前があっても形が合っているとは限らない。** 出荷済みの
				// 誤りは 3 件とも「載ってはいるが型が違う」形だった
				// (Bristol の legalTargets、Rummy500 の layoffTargets、
				// ソリティア 12 本の tableau)。名前だけ見る検査では出ない。
				if d := s.deref(props[k]); d != nil {
					if want := s.types(d); len(want) > 0 {
						got := jsonTypeOf(sub)
						allowed := got == "null"
						for w := range want {
							if typeAllows(w, got) {
								allowed = true
							}
						}
						if !allowed {
							gaps[fmt.Sprintf("%s: 宣言=%s 実際=%s",
								strings.TrimPrefix(path+"."+k, "."), joinKeys(want), got)] = true
						}
					}
				}
				s.walk(props[k], sub, path+"."+k, gaps, depth+1)
			case sch.AdditionalProperties.Schema != nil:
				s.walk(sch.AdditionalProperties.Schema, sub, path+".*", gaps, depth+1)
			default:
				gaps[strings.TrimPrefix(path+"."+k, ".")] = true
			}
		}
	case []any:
		if sch.Items == nil {
			return
		}
		for i, el := range v {
			if i >= 3 {
				break
			}
			s.walk(sch.Items, el, path+"[]", gaps, depth+1)
		}
	}
}
