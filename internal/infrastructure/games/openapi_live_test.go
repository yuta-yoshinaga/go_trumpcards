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
//
// **打った後にしか現れる形もある。** `humanAction` は 1 手打つまで null なので、
// 進行系のコマンドを何巡か回す。知らないコマンドは各ゲームが穏当に断るだけなので、
// 当たらないものが混ざっていて構わない ── 届く状態を増やす方が大事。
var liveProbeCommands = buildProbeCommands()

// buildProbeCommands は `reset` のあと、進行系のコマンドを 3 巡ぶん並べる。
func buildProbeCommands() []string {
	turns := []string{
		"n", "next", "nr", "hint", "log",
		"p", "play", "d", "draw", "s", "stand", "h", "hit", "c", "call", "k", "check",
	}
	out := []string{"reset"}
	for range 3 {
		out = append(out, turns...)
	}
	return out
}

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
//
// **落ちたらフレークではなく本物。** 配りは毎回違うので、届く盤面も毎回違う。
// この検査は「返ってきた項目」しか見ないので**見落とすことはあっても、
// 無い欠落を報告することはない**。実際 CI が 1 回だけ出した
// `teenpatti: messageParams` も、20 回回して 2 回だけ出た
// `currentLowHand` と `gofish.books` も、全部本物の食い違いだった。
// 再実行で緑になっても直すこと。
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
		bad := path.Post.Responses.Bad.Content.JSON.Schema
		ctrl := g.NewWebController()
		gaps := map[string]bool{}
		for _, cmd := range liveProbeCommands {
			body, code := probe(t, ctrl, g.Name, cmd)
			if body == nil {
				continue
			}
			checked++
			// **エラーの応答もスキーマを持っている。** 400 の本体はゲームの
			// 状態一式 (ただしゼロ値) + メッセージなので、`{message}` だけの
			// スキーマでは足りない。enum は見ない ── 盤面がゼロ値なので
			// `phase: 0` を違反として報告してしまう (#7055)。
			if code != http.StatusOK {
				spec.skipEnum = true
				spec.walk(bad, body, "", gaps, 0)
				continue
			}
			// **ヒントの応答は盤面が入っていないことがある。** 54 の presenter は
			// `HintOutput` で空の出力を作ってヒントだけ詰めるので、`phase` や
			// `chips` がゼロ値のまま返る (#7053)。項目の有無と型は正しいので
			// そちらは見るが、enum を比べると「盤面が無い」だけの応答を違反として
			// 報告してしまう。`h` はゲームによってヒントの別名なので一緒に外す。
			spec.skipEnum = cmd == "hint" || cmd == "h"
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

// checkEnum は宣言した enum が実際の値を許すか見る。
//
// **enum は「載っている」だけでは足りない検査。** `rebuyPhaseType` は
// `HoldemRebuyPhaseNone = 0` を返すのに `enum: [1, 2]` と書いてあり、
// Dramaha は共有スキーマに無いフェーズ 8 を返していた (#7052)。名前も型も
// 合っているので、ここを見ないと出ない。
//
// **200 の応答でだけ見る。** そのゲームが受け付けないコマンドは 400 を返し、
// 本体には 0 値の盤面が乗る。状態コードを見ずに突き合わせると、初期化されて
// いない `phase: 0` や `tableSize: 0` が違反として山ほど出て本物が埋もれる ──
// サーバは正しく、検査の方が間違っていた。`probe` が 200 だけを通すように
// なったので、enum も全コマンドの応答で見てよい。
func (s *liveSpec) checkEnum(sch *liveSchema, v any, path string, gaps map[string]bool) {
	if s.skipEnum {
		return
	}
	d := s.deref(sch)
	if d == nil || len(d.Enum) == 0 {
		return
	}
	switch v.(type) {
	case map[string]any, []any, nil:
		return // enum は目盛りのある値にしか意味が無い
	}
	for _, want := range d.Enum {
		if enumEqual(want, v) {
			return
		}
	}
	gaps[fmt.Sprintf("%s: enum=%v 実際=%v", strings.TrimPrefix(path, "."), d.Enum, v)] = true
}

// enumEqual は YAML の enum 値と JSON の値を比べる。
//
// **数値の型がそろわない。** YAML は 1 を int で、JSON は float64 で持つので、
// そのまま比べると常に食い違う。
func enumEqual(want, got any) bool {
	if wf, ok := toFloat(want); ok {
		if gf, ok2 := toFloat(got); ok2 {
			return wf == gf
		}
		return false
	}
	return want == got
}

// toFloat は数値なら float64 にして返す。
func toFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case float64:
		return t, true
	}
	return 0, false
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

// probe は 1 コマンドを叩いて応答と状態コードを返す。
//
// **状態コードを返すこと。** 呼び出し側はこれで 200 と 400 を振り分け、
// それぞれのスキーマと突き合わせる。見ないまま 400 の本体を 200 のスキーマと
// 比べると、0 値の盤面が enum 違反として山ほど出る ── サーバは正しく、
// 検査の方が間違っていた。
func probe(t *testing.T, ctrl games.WebController, name, cmd string) (map[string]any, int) {
	t.Helper()
	payload := fmt.Sprintf(`{"command":%q,"sessionId":"live-%s"}`, cmd, name)
	req := httptest.NewRequest(http.MethodPost, "/"+name+"/exec", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctrl.Exec(rec, req)
	var body map[string]any
	if json.Unmarshal(rec.Body.Bytes(), &body) != nil {
		return nil, rec.Code
	}
	return body, rec.Code
}

// --- openapi.yaml の必要な部分だけを読む型 ---

type liveSpec struct {
	// skipEnum は「いまの応答では enum を見ない」。盤面の入らない `hint` の応答用。
	skipEnum   bool
	Paths      map[string]livePath `yaml:"paths"`
	Components struct {
		Schemas map[string]*liveSchema `yaml:"schemas"`
	} `yaml:"components"`
}

type livePath struct {
	Post struct {
		// RequestBody は #7050 の逆向き検査 (宣言だけあって作れない項目) でも歩く。
		RequestBody liveBody `yaml:"requestBody"`
		Responses   struct {
			OK  liveBody `yaml:"200"`
			Bad liveBody `yaml:"400"`
		} `yaml:"responses"`
	} `yaml:"post"`
}

// liveBody は requestBody と 200 応答に共通の中身。
type liveBody struct {
	Content struct {
		JSON struct {
			Schema *liveSchema `yaml:"schema"`
		} `yaml:"application/json"`
	} `yaml:"content"`
}

type liveSchema struct {
	Type                 string                 `yaml:"type"`
	Enum                 []any                  `yaml:"enum"`
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
				s.checkEnum(props[k], sub, path+"."+k, gaps)
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
