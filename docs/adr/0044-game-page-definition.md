# ADR-0044: ゲームページを GameDefinition から組み立てる (createGamePage)

## Status

Accepted（2026-10-19、2 ページ目の Hearts が定義の拡張だけで移行できたため）

## Date

2026-10-19

## Context

バックエンドは `games_server.go` とレジストリが 1 つで各ゲームはプラグイン。フロントは 387 個の `pages/*Page.tsx` がそれぞれ合成ルートとして横断機能を手で配線している。依存の向きは健全（components/hooks/utils/api/types → pages の逆参照 0 件、components → api 0 件）で、問題は向きではなく N ページ × M 横断機能の配線。

計測 (2026-10-19): `withTutorial(` 377 / `<GamePageShell` 372 / `useGameHint(` 363 / `<ErrorAlert` 355 / `<GameSkeleton` 347 / `useCliMode(` 342 / `<SettingsPanel` 342 / `<GameResetButton` 341 / `<ActionLogSection` 329 / `hintCheckboxItem(` 315 ページ (grep)。委譲スキャンで 12 ページを行単位で分類すると 9,700 行中 4,116 行 (約 42%、モデルによる帰属なので幅のある値) が横断配線。`git log --first-parent 2026-07-01` 以降で 5 ページ以上を触ったコミット 88 件、ページ編集の約 32% がそこに集中し、上位は「共通化を 1 機能ずつ全ページへ撒く」リファクタ (#4336 151 ページ / #4339 115 / #10915 69)。`frontend/scripts/check-*.mjs` 約 40 本の多くが「全ページが X を配線したか」を後から保証する guard。

検討した代替案:

- (a) Feature-Sliced Design や層の import 制限 (Biome noRestrictedImports + overrides) — 現状違反 0 件なので配線の重複は減らない。
- (b) 現状維持で共通化を 1 機能ずつ撒き続ける — 1 回 100 ページ規模の変更が今後も続く。
- (c) 宣言的な `GameDefinition` と 1 つの `createGamePage` — 採用。

Issue #11912。

## Decision

`GameDefinition { key, useGame, useView?, tutorial, skeleton, cli, busy?, phase, shortcuts?, shortcutsEnabled?, shortcutsTestId?, primaryActions?, message?, cardNav?, settings?, settingsPlacement?, landscapeBanner?, resetTutorialId?, actionsTutorialId?, HeaderExtra?, LiveRegion?, Board, FooterHand?, FooterStatus?, FooterActions?, shell?, reset?, contentClassName? }` を宣言し `createGamePage(def)` が横断配線を 1 か所で行う。ページは named export (`export const XPage = createGamePage(def)`) を保つので `App.tsx` の `import.meta.glob` ルーティングは無変更。

`cli` は `parseCommand`/`formatResponse`/`helpText` の名前のままページに置く (`check-cli-hint-reachable` がページの `parseCommand:` を grep するため)。ヒントの描画はゲーム固有なので `Board` に残す (`check-hint-live-region` / `check-hint-gate` はページファイルを読むので無変更で通る)。

試作結果 (AcesUp): 496 行 → 372 行、import 40 → 20、手配線していた横断シンボル 19 種 → 0 種。`AcesUpPage.test.tsx` (37 件) と `e2e/acesup.spec.ts` は無変更で通過。guard スクリプトの変更は 0 本。

2 ページ目 (Hearts、trick-taking・カード選択・パスフェーズ): 744 行 → 680 行、import 43 → 26、横断シンボルの手配線は `hintCheckboxItem` (ゲーム固有設定と並べるため) を除いて 0。足した口は `cardNav` (カード番号キー操作、62 ページが `useCardKeyboardNav` を使う)、`FooterHand` / `FooterStatus`、`actionsTutorialId`、`settingsPlacement` (設定パネルが盤面の上か下かは概算 169 / 172 ページで二分。ソース上の行位置の比較による)、`landscapeBanner` (91 ページだけが出すので既定は false)。`HeartsPage.test.tsx` と `e2e/hearts.spec.ts` は無変更で通過。

3 ページ目 (Spades、固定アクション付きフッター): 676 行 → 588 行、import 41 → 24。足した口は `primaryActions` (定義すると GameFooter の固定形になり、返り値をそのまま `actions` に渡すので null で操作行が消える。115 ページが固定形) と `message`。固定形では ErrorAlert・FooterStatus・操作行・ショートカットをスクロール領域の末尾 (ActionLogSection の直後) に置く。`SpadesPage.test.tsx` は無変更で通過。

## Consequences

次の横断機能は `createGamePage` 1 ファイルの変更で済む。一斉移行はしない: 新規ゲームは new-game スキルから定義で作る、既存ページは触るついでに移行。`frontend/scripts/check-game-page-definition.mjs` は手配線の新規ページを拒み、`frontend/scripts/game-page-definition-baseline.txt` の行数が未移行ページ数になる（移行した PR でその行を消す、行を足すのは禁止）。行を足すこと自体は機械的には止めていない: CI の浅い checkout には develop 側のベースラインが無いので、足された行はレビューで見る。ページ単位の guard を `createGamePage` 1 ファイルの検査へ置き換えられる時期は、この行数で判断する。数えるのはゲームとして経路に載り、自分で配線しているページだけ。別のページをそのまま描く派生ページ 8 本（DoubleExposure など）は元のページと一緒に移り、ゲームでないページ 4 本（Discover、DiscoverResult、Legal、NotFound）は対象外なので、ベースラインから外した（384 → 372 行）。

まだ表現できないもの: CPU ログやモバイル/デスクトップで分かれるサイドバー — Hearts では不要だった (サイドバーは `lgTwoColGrid` を Board 内で完結)。必要になるページを移すときに定義へ足す (#11915)。

スコープ外: フロントのルール二重実装 (#11913)、`types/` の `openapi.yaml` からの生成。

## 追記 (2026-10-24): 試験移行の実測

「触るついでに移行」のままにするか一斉移行へ進むかを数字で決めるため、trick-taking 系の 4 ページを試験的に移した (#12042 Doppelkopf、#12043 Tute、#12044 Schafkopf、Euchre)。

- **効果（直近 6 週の実測）:** 10 ページ以上を一度に変えたコミットが 24 本あった（ページの変更は延べ 639 件）。上位 12 本のうち、移行済みなら `createGamePage` 1 ファイルの変更で済んだのは 2〜3 本だった（マウント時リセット #10915、操作ボタンの固定 #11628）。残りは色トークン、レイアウト、記号表、Biome 警告など Board の中身の変更なので、移行してもページごとに残る。
- **縮み方:** 4 ページとも import は 38〜40 から 23 になった。行数は 466→428、496→419、535→491、679→629。
- **不具合率:** 4 ページ中 1 ページ。Doppelkopf の委譲がフェーズ表示に翻訳前のキー（`play`）を渡していた。既存 28 件の試験はこの文字列を見ておらず、全部緑のまま通った。`delegation-diff-auditor` が見つけたので、試験を足して直した。ほかの 3 ページは監査で振る舞いの差が見つからず、既存の単体試験と E2E（`--repeat-each`）も無変更で通った。
- **工数:** 委譲 1 ページあたり 4〜11 分。既定の codex モデルは 4 ページ中 3 ページを未着手で返し、別バックエンドへ再委譲して完了した。受け取りでは、ページごとに監査を 1 回と、ビルドし直したうえでの E2E を 1 回ずつ回した。

決定 (2026-10-25、ユーザ判断): **系統ごとに移行を続ける。** 次の系統を 4〜8 ページずつ、下の受け入れ条件で移す。一度に全ページは移さない。効果は 1 ファイル化できる横断変更に限られ、移行が既存試験をすり抜けて振る舞いを変えることもある (1/4)。そのため、受け入れ条件に監査と E2E を含める。条件は次の 3 つ。

1. 既存の単体試験と E2E を変えずに通す。
2. フェーズ名を `usePhaseNames` 経由にする。
3. `delegation-diff-auditor` の判定が OK である。

上の「まだ表現できないもの」は #11915 で解消済み。CPU ログも分割サイドバーも `Board` の中に置けるので、定義に口は足していない (2026-10-20、NOT_PLANNED)。

## 追記 (2026-10-25): 系統移行 2 回目で見つかったファクトリの不足

OldMaid、Ecarte、Jass、Belote を移した。監査を通ったのは Ecarte だけだった。残りの 3 ページは委譲の誤りではなく、ファクトリ側の不足で振る舞いが変わった。上の「CPU ログも分割サイドバーも `Board` に置ける」は正しかった（OldMaid の CPU ログは Board の中で同じ位置に出た）。ただし、`Board` だけでは表せない形が 4 つあった。そこで次の口を足した:

- `cli` の引数に `frontendHint` を渡す。OldMaid の CLI `hint` はヒントの ON/OFF に従う必要があったが、`cli` からはヒントの状態が見えなかった。
- `AboveContent`: 上部の設定パネルとスクロール領域の間に置き、スクロールしない要素のための口。OldMaid の metaAI 表示と再生速度の設定がこの位置にある。`Board` に入れるとスクロールし、余白も二重になった。
- `actionsClassName` と `resetButton: 'custom'`（`renderResetButton` を context で渡す）: フッターの操作列は `flex` で左寄せになり、リセットが最後に並ぶ。中央寄せでリセットが途中に並ぶページ（OldMaid）は、そのままでは元の見た目に戻せなかった。
- `shell` に `gameThemeBg` を渡せるようにした（OldMaid の揺れ演出用の背景）。

CLI モードの無いページは、移行の前に CLI を正式な機能として足す (2026-10-25、ユーザ判断)。対象は未移行 368 ページ中 33 ページ。`cli` は必須のままにした。委譲が既存の無いページで CLI を即席で書き、試験も無く、設定を無視する reset を入れた (Jass、Belote) ためである。

委譲の前に各候補をこの 4 つの観点で grep し、今のファクトリで表せるページだけを渡す。

CallBreak のチュートリアル対象は固定フッター自体なので、`primaryActions` と組み合わせて固定フッターの操作行に `data-tutorial` を付ける `footerTutorialId` を追加する。
