# ADR-0031: 4 レイヤーゲームレジストリの取り扱い方針

## Status

Accepted

## Date

2026-05-02

## Context

`internal/infrastructure/games/registry.go` がメタデータ (`{Name, Category, Description}`) の SSoT を持っているにもかかわらず、ゲーム1つを追加・改名するたびに **4 ファイル** を同期更新する必要が残っている (Issue #1616)。

| # | ファイル | 役割 | 1 ゲームあたり行数 |
|---|---------|-----|------------------|
| ① | `internal/infrastructure/games/registry.go` | `{Name, Category, Description}` 宣言 (build tag なし) | 1 行 |
| ② | `internal/infrastructure/games/games_server.go` | `BindWebControllerFor` 呼び出し (`!js \|\| !wasm`) | 5 行 |
| ③ | `internal/infrastructure/games/{casino,classic,solo}/<cat>.go` | `RegisterKVGame` 呼び出し (`js && wasm`) | 8 行 |
| ④ | `internal/infrastructure/ui/GameManager.go` | `gameRegistry` の CUI ファクトリ | 15-30 行 |

合計 ~73 ゲーム × 平均 7 行 ≈ 510 行のメカニカルなコピペが、4 ファイルにまたがって存在している。

### この構造になっている理由

1. **Cloudflare Workers の 1 MB gzipped 制限** ([ADR-0027](0027-cloudflare-workers-wasm.md))。3 つの Worker (casino / classic / solo) に build tag で分割し、TinyGo の dead-code elimination により他カテゴリのゲームコードを除外する必要がある。これは構造的な要請であり、回避不可。
2. **`registry.go` は build tag なし** で全ビルドに含まれるため、`domain` / `usecase` / `controller` を import できない (WASM サイズが膨らむ)。よって型情報を持つファクトリは別ファイルから後付けする pattern になる。
3. CLI は build tag `!js \|\| !wasm` 配下で完結するため `gameRegistry` を持っても害はないが、CUI 専用の `CuiHelpSpec` を `games` パッケージに置くと無関係な依存が増えるため、`ui` パッケージ側に置かれている。

これらは設計上の必然である一方、**「1 ゲーム = 1 イベント」という認知上の単位が 4 箇所に分散している** のが本 Issue の指摘点である。

### ガードレールの現状

`internal/infrastructure/games/registry_test.go` に以下のテストがある:

- `TestAllReturnsExpectedTotal` — 総数の固定アサート (`expectedTotal = 73`)
- `TestByCategoryCounts` — カテゴリ別件数の固定アサート
- `TestAllEntriesAreValid` — 各エントリの Name 一意性、`NewWebController` 非 nil、Category 妥当性
- `TestRegistryMatchesCLI` — CLI の `gameRegistry` と順序ごと一致

つまり ① + ② + ④ の不整合は CI で検出される。**③ Worker 側の登録漏れだけが既存テストでカバーされていない** (テストは `!js \|\| !wasm` でしか走らないため)。

## Decision

検討した 3 案を以下に整理する。**本 ADR は方針合意のための提案として起草され、Issue #1616 での合意形成を経て案3 を採択済み (Status: Accepted)**。採択の経緯は末尾「採択結果 (2026-05-02)」節を参照。

### 案1: 1 ゲーム = 1 ディレクトリ (per-game packages)

```
internal/infrastructure/games/
  registry.go                   (no tag, types only)
  blackjack/
    blackjack.go                (no tag, registers metadata)
    blackjack_server.go         (!js || !wasm, server factory)
    blackjack_worker.go         (js && wasm, worker factory)
    blackjack_cui.go            (!js || !wasm, CUI factory)
  casino/casino.go              (js && wasm, blank-imports casino/* games)
```

各ゲームの登録は 1 ディレクトリで完結。新規追加時の編集箇所は **4 → 1**。

#### Pros

- 認知負荷の劇的な低下: 「BlackJack 関連は `blackjack/` を見ろ」で完結。
- `TestRegistryMatchesCLI` のような cross-file 整合性テストが構造的に不要 (1 ゲーム = 1 init() の宣言で 4 レイヤー全てが登録される)。
- 削除・改名の影響範囲が `git mv` 1 回で済む。

#### Cons

- **新規ファイル数が増える**: 1 ゲーム × 4 ファイル ≈ **292 個の新規ファイル**。各ファイルは小さいが、ディレクトリ走査・grep・PR の認知コストは増える。
- **build tag 配線が脆い**: `_server.go` / `_worker.go` / `_cui.go` のタグを 1 つでも誤ると、対応する init() が走らず該当ゲームがそのビルドから消える。テストで検出はできるが、書き間違いの局所性は失われる (registry.go の 1 行の方が見つけやすい)。
- **WASM サイズへのリスク**: 各 per-game パッケージが `domain` / `usecase` / `controller` / `presenter` を import する。TinyGo の dead-code elimination は per-symbol だが、blank import の到達可能性に依存する。リファクタ後にカテゴリ間で意図せぬ symbol leak が起きると、worker サイズが 1 MB を超え CI で気付くことになる。
- **移行コストが大きい**: 73 ゲーム × 4 箇所 = 292 ファイルへの分散は段階的にしか進められない。1 ゲームずつ移行する PR 群か、全ゲーム一括の巨大 PR の二択。

### 案2: code generation (`go generate` ベース)

```go
// registry.go (人間が編集する唯一の場所)
//go:generate go run ./internal/cmd/genregistry
var registry = []*Game{
    {Name: "blackjack", Category: CategoryCasino, Description: "BlackJack (...)"},
    ...
}
```

`go run ./internal/cmd/genregistry` が ② / ③ / ④ を自動生成。`registry.go` の 1 行追加 + `go generate` で完了。

#### Pros

- **新規追加時の手作業は 1 行のみ**。
- **build tag / Worker 構造を変えない** → WASM サイズへの影響ゼロ。
- 生成物はコミットされる (git 上で diff レビュー可能)。
- ジェネレータ自体は ~200 LoC で完結する見込み (テンプレート + struct → file、案2 採択時の見積もりと整合)。

#### Cons

- **ジェネレータがメンテ対象になる**: テンプレート変更は全ゲームに伝播する (≒ 一回の変更が広い blast radius を持つ)。
- **CI に `go generate && git diff --exit-code` のステップが必要** (生成漏れの検出)。
- 生成コードのナビゲーション体験は悪化 (grep で `func NewBlackJackWebController` を辿ると生成ファイルに着地する)。
- **`gameRegistry` (④) のテンプレート化が難しい**: BlackJack の `CuiHelpSpec` は他ゲームと違って "Insurance" や "Decline Insurance" など独自キーを持つ。テンプレートで全ゲーム分を表現するには、`registry.go` に i18n キーや CommandKeys 配列まで宣言する必要があり、SSoT が肥大化する。

### 案3: 現状維持 + ガードレール強化

`registry.go` を SSoT のまま、4 ファイルへの分散も維持。代わりに **「③ Worker 側登録漏れ」を CI で検出する仕組みを追加**。

実装上の制約: 既存の `casino/casino.go` 等は `//go:build js && wasm` で gated されており、これらの init() を実行して `Game.RegisterWorker` の populate 後を検査するには、テストも同じビルドタグ下で **WASM ランタイム上で実行** する必要がある (TinyGo + wasmer か `GOOS=wasip1` のいずれか — 現行 CI には未導入)。

そのため、**ランタイム実行ではなくソースコードの AST 解析で同じ漏れを検出するアプローチを推奨**:

```go
// internal/infrastructure/games/registry_worker_consistency_test.go
// (build tag なし — 通常の go test ./... で実行可能)

func TestWorkerRegistrationsCoverAllGames(t *testing.T) {
    // 1. games.All() からカテゴリごとの期待ゲーム名集合を取得。
    // 2. {casino,classic,solo}/{cat}.go を go/parser で AST に展開。
    // 3. 各ファイルの init() 内の RegisterKVGame("name", ...) 呼び出しから
    //    name 引数を抽出し、期待集合と一致することを assert。
}
```

実行時に `RegisterWorker != nil` を assert したい場合は、`registry.go:RegisterCategory` が既に同等のチェックを startup-time で行っており (ローカル `make build-worker-casino` でビルドしたワーカーを起動すれば漏れがあれば即座に panic。`cmd/workers/*` は `//go:build js && wasm` なので `go run` では建てられない)、CI の AST 検査と二重で守られる形になる。

加えて `docs/new-game-checklist.md` を「リトマステスト的なテスト」へ寄せる:

- ① + ② + ④ の不整合 → 既存 `TestRegistryMatchesCLI` で検出済み。
- ③ の不整合 → 上記の `TestWorkerRegistrationsCoverAllGames` を新規追加。
- 結果: **4 ファイル中いずれか 1 つでも漏れたら CI が落ちる** 状態。

#### Pros

- **着手コストが小さい** (1 PR、~90 LoC: テストファイル 1 つ + 軽い AST ヘルパー)。
- **既存の build tag 構造を壊さない** → WASM サイズ影響ゼロ。
- 73 ゲーム × 4 ファイルの認知負荷は残るが、**「漏れたら CI で気付ける」事実そのものが最大の改善**。
- 案1・案2 と互換: 後から本格的なリファクタに移行しても、ガードレール自体は流用可能。

#### Cons

- 新規ゲーム追加時の 4 ファイル編集自体は残る (案1・2 と比べると見劣りする)。
- 「このゲームの登録はどこ?」という質問への答えは依然として「4 つのファイルを順に追え」のまま。

### 推奨

**案3 (現状維持 + ガードレール強化) を第一候補とする**。理由:

1. **着手コストとリスクのバランス**: 案1・2 は移行に数 PR / 数週間を要するうえ、build tag の取り扱いを誤ると WASM ビルドが silent に壊れる (1 MB 超過時のみ CI で検出)。案3 は ~90 LoC の test-only 変更 (AST ベース) で「漏れ検出」というコア課題を解決できる。
2. **新規ゲーム追加の頻度**: 直近 12 ヶ月で 73 ゲームに到達した後、ペースは緩やかになりつつある。残存する ROI は「加える側の摩擦低減」より「漏れの早期検出」の方が大きい。
3. **方針変更の選択肢を保持**: ガードレールがあれば、後で案1 (per-game packages) に移行しても安全。一方、いきなり案1 に着手すると、移行中に発生する登録漏れを既存テストで検出できない。

ただし、案1 が望ましいケースもある: 「将来的にゲーム数を 200+ に拡張する予定がある」「ゲーム単位でデプロイ・バージョニングしたい」など、`games/` の物理的なディレクトリ分割そのものに価値がある場合。本プロジェクトは現状そのいずれにも該当しないため、案3 を推奨する。

**本 ADR は提案段階を経て、Issue #1616 のコメントでの合意形成を踏まえ案3 を採択した (Status: Accepted)**。採択結果は次節「採択結果 (2026-05-02)」のとおり。

### 採択結果 (2026-05-02)

Issue #1616 で **案3 (現状維持 + ガードレール強化)** を採択。実装は `internal/infrastructure/games/registry_worker_consistency_test.go` の `TestWorkerRegistrationsCoverAllGames` として導入し、3 つの worker サブパッケージ (`casino` / `classic` / `solo`) のソースを `go/parser` で AST に展開して `RegisterKVGame` 呼び出しの第 1 引数を抽出、`games.ByCategory(...)` と集合一致を検査する形になった。`//go:build js && wasm` 配下のランタイムテストを書く案も検討したが、CI に wasm ランタイム (TinyGo / `GOOS=wasip1`) を新規導入する追加コストが見合わないと判断。AST 解析であれば既存の `go test -tags test ./...` で実行できる。

## Consequences

(※ 本 ADR は提案段階のため、各案を採択した場合の影響を併記する)

### 案3 採択時 (推奨)

- **追加コスト**: `internal/infrastructure/games/registry_worker_consistency_test.go` 1 ファイル (build tag なし、AST 解析で 3 カテゴリを一括チェック) ≈ ~90 LoC。
- **CI コスト**: ゼロ。`go test -tags test ./...` で実行されるため、現行の CI に追加ステップ不要。
  - 代替案として `//go:build js && wasm` 下でランタイムテスト (`Game.RegisterWorker != nil` を assert) を書く方法もあるが、これは TinyGo + wasmer か `GOOS=wasip1 GOARCH=wasm` のいずれかを CI に追加する必要があり、Worker サブパッケージのビルドタグ変更も伴うため、**AST 解析ルートを推奨**。
- **`docs/new-game-checklist.md` を更新**: 「漏れたら CI で気付く」前提に書き換え、心理的ハードルを下げる。
- **後方互換**: 既存ファイル構造は無変更、外部 API への影響もなし。

### 案1 採択時

- **大規模リファクタ**: 73 ゲーム × 4 ファイル = ~292 ファイルの新規作成 + 既存 4 ファイルの空化または削除。1 ゲームずつのインクリメンタル PR (推定 73 PR、数週間) または全件一括 PR (レビュー困難) の二択。
- **CI 強化必須**: WASM ビルドサイズチェック (現行 [ADR-0027](0027-cloudflare-workers-wasm.md) 通り) に加えて、各カテゴリのゲーム数固定アサートを worker サブパッケージ側にも追加。
- **`new-game-checklist.md` 大幅簡素化**: 「`internal/infrastructure/games/<game>/` を作って 4 ファイル書く」の 1 ステップへ短縮。
- **deprecation 期間**: 旧構造と新構造の並存期間が必要なら、`registry.go` のメタデータ宣言は当面維持。

### 案2 採択時

- **新規ツール**: `internal/cmd/genregistry/main.go` (テンプレート + 出力ロジック)、推定 200 LoC。
- **CI 追加**: `go generate ./... && git diff --exit-code` ステップ。
- **`registry.go` の肥大化**: i18n 用 CommandKeys / SettingKeys / TitleKey をメタデータに追加する必要がある。
- **テンプレート保守**: 登録 API の signature 変更時、テンプレートも追従する。`BindWebControllerFor` / `RegisterKVGame` の入力 / 出力型は generics で安定しているため大きな変更は想定しにくいが、ジェネレータ独自の障害ポイントが 1 つ増える。

## References

- Issue [#1616](https://github.com/yuta-yoshinaga/go_trumpcards/issues/1616) — 元の課題提起
- [ADR-0027](0027-cloudflare-workers-wasm.md) — 1 MB gzipped 制限の根拠
- [ADR-0028](0028-kv-session-persistence.md) — KV セッション永続化 (案1・3 で worker 構造を維持する根拠)
- `docs/new-game-checklist.md` — 現状の手作業チェックリスト
- `internal/infrastructure/games/registry_test.go` — 既存ガードレール
