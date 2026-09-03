package games_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// unreachableAllowFile は「そのゲームのスキーマから辿り着けない」既知の項目。
const unreachableAllowFile = "testdata_openapi_unreachable.json"

// gameExecPathRe は `/<game>/exec` のパスからゲーム名を取る。
var gameExecPathRe = regexp.MustCompile(`^/([a-z0-9]+)/exec$`)

// routeOverride はコントローラ名を小文字にしても経路にならないゲーム。
//
// **登録名は構造体名の小文字とは限らない。** この 4 本は経路を持っているのに
// `strings.ToLower` では引けず、黙って検査の外に出ていた (レビュー指摘)。
var routeOverride = map[string]string{
	"SetteEMezzo":           "settemezzo",
	"DoubleAttackBlackjack": "doubleattack",
	"FreeBetBlackjack":      "freebet",
	"PigsTail":              "pigtail",
}

// routeOf はコントローラ名から `/<route>/exec` の route を返す。
func routeOf(game string) string {
	if r, ok := routeOverride[game]; ok {
		return r
	}
	return strings.ToLower(game)
}

// TestOpenAPISchemaReachesEveryField は各ゲームの項目が**そのゲームの**
// スキーマから辿れることを守る。
//
// **名前の一致では置き場所の誤りが見えない。**
// `TestOpenAPIDocumentsEveryResponseField` は「その名前が spec のどこかに
// あるか」しか見ていない (そう書いてある)。だから `scartoCount` を 1 ゲームに
// 書けば同名を返す他のゲームも記載済みになるし、**API が返さないキーを
// 書いても通る** ── 実際 PR #7045 で `result:` (実際は `lastResult`) など
// 4 件を書いてしまい、1 件はレビューで、3 件は自分の突き合わせで見つかった。
//
// こちらは応答 (と requestBody) のスキーマを `$ref` / `allOf` / `items` /
// `additionalProperties` まで展開して、そこから辿れる名前の集合を作る。
// 起票時は 148 ゲーム / 1,741 項目が辿れなかったので、#6984 と同じく許可リストで
// 増やせなくする形にした。**その後ぜんぶ埋めたので、いま許可リストは空**
// (`testdata_openapi_unreachable.json` = `{}`)。仕組みは残してある:
//
//   - 許可リストに無い到達不能を見つけたら落ちる  → 新規の追加は止まる
//   - 許可リストにあるのに実際は到達できるなら落ちる → 件数は減る方向にしか動かない
//
// **空の許可リストがいちばん強い状態**なので、項目を足したくなったら
// まず spec に書けないか考えること。
//
// (#7046)
func TestOpenAPISchemaReachesEveryField(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	raw, err := os.ReadFile(filepath.Join(root, "api", "openapi.yaml")) //nolint:gosec // test-only, fixed path
	if err != nil {
		t.Fatalf("openapi.yaml が読めない: %v", err)
	}
	var spec openAPIDoc
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		t.Fatalf("openapi.yaml が YAML として壊れている: %v", err)
	}
	// **展開が空振りしていないことを先に確かめる。** 0 件しか集められないと
	// 「全部到達不能」でも「全部到達可能」でも見た目が変わらなくなる。
	if len(spec.Paths) < 300 {
		t.Fatalf("`/<game>/exec` のパスが %d 件しか読めていない。展開が壊れている", len(spec.Paths))
	}

	reachable := map[string]map[string]bool{}
	for path, ops := range spec.Paths {
		m := gameExecPathRe.FindStringSubmatch(path)
		if m == nil {
			continue
		}
		got := map[string]bool{}
		collectKeys(&spec, ops.Post.Responses.OK.Content.JSON.Schema, got, 0)
		collectKeys(&spec, ops.Post.RequestBody.Content.JSON.Schema, got, 0)
		reachable[m[1]] = got
	}
	if len(reachable) < 300 {
		t.Fatalf("ゲームの応答スキーマが %d 件しか集まっていない", len(reachable))
	}

	controllers, err := filepath.Glob(filepath.Join(root, "internal", "adapter", "controller", "*WebController.go"))
	if err != nil || len(controllers) < 200 {
		t.Fatalf("WebController が %d 件しか見つからない (err=%v)", len(controllers), err)
	}

	actual := map[string][]string{}
	for _, path := range controllers {
		src, readErr := os.ReadFile(path) //nolint:gosec // test-only, repo-relative
		if readErr != nil {
			t.Fatalf("%s が読めない: %v", path, readErr)
		}
		game := strings.TrimSuffix(filepath.Base(path), "WebController.go")
		got, ok := reachable[routeOf(game)]
		// **経路が引けないなら落とす。** 黙って飛ばすと、そのゲームは
		// この ratchet の外に出たまま二度と検査されない。実測で 352 本すべてに
		// 経路があるので、引けないのは対応表の漏れであって「CUI 専用」ではない
		// (レビュー指摘: 4 本が黙って飛ばされていた)。
		if !ok {
			t.Fatalf("%s の経路 (%q) が openapi.yaml に無い。routeOverride に足すこと", game, routeOf(game))
		}
		seen := map[string]bool{}
		var missing []string
		for _, m := range jsonTagRe.FindAllStringSubmatch(string(src), -1) {
			if got[m[1]] || seen[m[1]] {
				continue
			}
			seen[m[1]] = true
			missing = append(missing, m[1])
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			actual[game] = missing
		}
	}

	allowRaw, err := os.ReadFile(filepath.Join("testdata", unreachableAllowFile))
	if err != nil {
		t.Fatalf("許可リストが読めない: %v", err)
	}
	var allow map[string][]string
	if err := json.Unmarshal(allowRaw, &allow); err != nil {
		t.Fatalf("許可リストが JSON として壊れている: %v", err)
	}
	allowSet := map[string]map[string]bool{}
	for game, fields := range allow {
		allowSet[game] = map[string]bool{}
		for _, f := range fields {
			allowSet[game][f] = true
		}
	}

	var added []string
	for game, fields := range actual {
		for _, f := range fields {
			if !allowSet[game][f] {
				added = append(added, game+"."+f)
			}
		}
	}
	sort.Strings(added)
	if len(added) > 0 {
		t.Errorf("そのゲームのスキーマから辿れない項目が増えている (%d 件): %v\n"+
			"項目を足したら、そのゲームの応答スキーマ (または requestBody) にも書くこと。"+
			"意図的に保留するなら %s に足す。", len(added), added, unreachableAllowFile)
	}

	actualSet := map[string]map[string]bool{}
	for game, fields := range actual {
		actualSet[game] = map[string]bool{}
		for _, f := range fields {
			actualSet[game][f] = true
		}
	}
	var stale []string
	for game, fields := range allow {
		for _, f := range fields {
			if !actualSet[game][f] {
				stale = append(stale, game+"."+f)
			}
		}
	}
	sort.Strings(stale)
	if len(stale) > 0 {
		t.Errorf("許可リストの項目が %d 件、もう到達不能ではない: %v\n"+
			"書けたぶんは %s から消すこと (件数は減る方向にしか動かせない)。",
			len(stale), stale, unreachableAllowFile)
	}
}

// --- openapi.yaml の必要な部分だけを読む型 ---

type openAPIDoc struct {
	Paths      map[string]openAPIPathItem `yaml:"paths"`
	Components struct {
		Schemas map[string]*openAPISchema `yaml:"schemas"`
	} `yaml:"components"`
}

type openAPIPathItem struct {
	Post struct {
		RequestBody openAPIBody `yaml:"requestBody"`
		Responses   struct {
			OK openAPIBody `yaml:"200"`
		} `yaml:"responses"`
	} `yaml:"post"`
}

type openAPIBody struct {
	Content struct {
		JSON struct {
			Schema *openAPISchema `yaml:"schema"`
		} `yaml:"application/json"`
	} `yaml:"content"`
}

type openAPISchema struct {
	Ref                  string                    `yaml:"$ref"`
	Properties           map[string]*openAPISchema `yaml:"properties"`
	Items                *openAPISchema            `yaml:"items"`
	AllOf                []*openAPISchema          `yaml:"allOf"`
	OneOf                []*openAPISchema          `yaml:"oneOf"`
	AnyOf                []*openAPISchema          `yaml:"anyOf"`
	AdditionalProperties additionalProps           `yaml:"additionalProperties"`
}

// additionalProps は `additionalProperties` がスキーマにも真偽値にもなる
// (`additionalProperties: true` が 50 ゲームにある) のを受ける。
type additionalProps struct{ schema *openAPISchema }

// UnmarshalYAML は真偽値のときは何も持たない値として読む。
func (a *additionalProps) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return nil // `additionalProperties: true` -- 展開する子は無い
	}
	var s openAPISchema
	if err := node.Decode(&s); err != nil {
		return err
	}
	a.schema = &s
	return nil
}

// openAPIRefPrefix は spec 内のスキーマ参照の接頭辞。
const openAPIRefPrefix = "#/components/schemas/"

// openAPIMaxDepth は展開の深さの上限。相互参照で止まらなくなるのを防ぐ。
const openAPIMaxDepth = 8

// collectKeys は s から辿れるプロパティ名を out に集める。
func collectKeys(doc *openAPIDoc, s *openAPISchema, out map[string]bool, depth int) {
	if s == nil || depth > openAPIMaxDepth {
		return
	}
	if s.Ref != "" {
		name := strings.TrimPrefix(s.Ref, openAPIRefPrefix)
		collectKeys(doc, doc.Components.Schemas[name], out, depth+1)
		return
	}
	for name, sub := range s.Properties {
		out[name] = true
		collectKeys(doc, sub, out, depth+1)
	}
	for _, group := range [][]*openAPISchema{s.AllOf, s.OneOf, s.AnyOf} {
		for _, sub := range group {
			collectKeys(doc, sub, out, depth+1)
		}
	}
	collectKeys(doc, s.Items, out, depth+1)
	collectKeys(doc, s.AdditionalProperties.schema, out, depth+1)
}
