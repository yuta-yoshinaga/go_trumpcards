package games_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// openapiPropertyRe は openapi.yaml のキー行を拾う。
//
// **プロパティ以外のキーにも当たる。** パス名やリクエスト側の項目も含むので、
// 「載っている」の判定は緩い。緩い側に倒しているのは意図的で、このガードの
// 役目は**未記載を増やさないこと**であって、記載の質を測ることではない。
// 厳しくすると既存 389 件の許可リストが膨らみ、誰も削らなくなる。
var openapiPropertyRe = regexp.MustCompile(`(?m)^\s*([A-Za-z][A-Za-z0-9_]*):\s*$`)

// jsonTagRe は Go の出力構造体の `json:"name"` からフィールド名を取る。
var jsonTagRe = regexp.MustCompile("`json:\"([A-Za-z][A-Za-z0-9_]*)[\",]")

// undocumentedAllowFile は「まだ openapi.yaml に書かれていない」既知の欠落。
const undocumentedAllowFile = "testdata_openapi_undocumented.json"

// TestOpenAPIDocumentsEveryResponseField は Web API のレスポンス項目が
// openapi.yaml に載っていることを守る。
//
// **CI が突き合わせていたのは root の `tags:` ブロックだけだった。** 各ゲームの
// レスポンス本体は誰も見ておらず、実際に返している項目のうち 389 件 (98 ゲーム分)
// が仕様書に一度も書かれないまま増え続けていた (#6984)。出荷済みでフロントが
// 実際に読んでいる項目まで含まれる。
//
// **既存分は許可リストに入れ、増やせないようにするだけ。** 389 件をこの場で
// 書き切るのは現実的でないので、
//
//   - 許可リストに無い未記載を見つけたら落ちる  → 新規の追加は今日から止まる
//   - 許可リストに載っているのに実際は記載済み/削除済みなら落ちる → 件数は減る方向にしか動かない
//
// の 2 方向で締める。2 つ目があるので、項目を書いたら許可リストから消す必要が
// あり、放置すると赤くなる。
func TestOpenAPIDocumentsEveryResponseField(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	spec, err := os.ReadFile(filepath.Join(root, "api", "openapi.yaml")) //nolint:gosec // test-only, fixed path
	if err != nil {
		t.Fatalf("openapi.yaml が読めない: %v", err)
	}
	documented := map[string]bool{}
	for _, m := range openapiPropertyRe.FindAllStringSubmatch(string(spec), -1) {
		documented[m[1]] = true
	}
	// **抽出が全部を見落としていないことを先に確かめる。** 0 件しか拾えていない
	// と「全部記載済み」に見えてしまい、このテストは永久に緑になる。
	if len(documented) < 500 {
		t.Fatalf("openapi.yaml から拾えたキーが %d 件しかない。抽出が壊れている", len(documented))
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
		seen := map[string]bool{}
		var missing []string
		for _, m := range jsonTagRe.FindAllStringSubmatch(string(src), -1) {
			if documented[m[1]] || seen[m[1]] {
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

	// go test の cwd は必ずパッケージのディレクトリなので、testdata/ の 1 経路だけでよい。
	allowRaw, err := os.ReadFile(filepath.Join("testdata", undocumentedAllowFile))
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

	// (1) 許可リストに無い未記載 = 新しく増えたぶん。
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
		t.Errorf("openapi.yaml に無いレスポンス項目が増えている (%d 件): %v\n"+
			"レスポンスに項目を足したら api/openapi.yaml にも同じコミットで書くこと。"+
			"意図的に保留するなら %s に足す。", len(added), added, undocumentedAllowFile)
	}

	// (2) 許可リストに載っているのに、いま未記載でないもの = 書けたか消えたぶん。
	//     残しておくと件数が減らないので落とす。
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
		t.Errorf("許可リストの項目が %d 件、もう未記載ではない: %v\n"+
			"書けたぶんは %s から消すこと (件数は減る方向にしか動かせない)。",
			len(stale), stale, undocumentedAllowFile)
	}
}
