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

`GameDefinition { key, useGame, useView?, tutorial, skeleton, cli, busy?, phase, shortcuts?, shortcutsEnabled?, shortcutsTestId?, cardNav?, settings?, settingsPlacement?, landscapeBanner?, resetTutorialId?, actionsTutorialId?, HeaderExtra?, LiveRegion?, Board, FooterHand?, FooterStatus?, FooterActions?, shell?, reset?, contentClassName? }` を宣言し `createGamePage(def)` が横断配線を 1 か所で行う。ページは named export (`export const XPage = createGamePage(def)`) を保つので `App.tsx` の `import.meta.glob` ルーティングは無変更。

`cli` は `parseCommand`/`formatResponse`/`helpText` の名前のままページに置く (`check-cli-hint-reachable` がページの `parseCommand:` を grep するため)。ヒントの描画はゲーム固有なので `Board` に残す (`check-hint-live-region` / `check-hint-gate` はページファイルを読むので無変更で通る)。

試作結果 (AcesUp): 496 行 → 372 行、import 40 → 20、手配線していた横断シンボル 19 種 → 0 種。`AcesUpPage.test.tsx` (37 件) と `e2e/acesup.spec.ts` は無変更で通過。guard スクリプトの変更は 0 本。

2 ページ目 (Hearts、trick-taking・カード選択・パスフェーズ): 744 行 → 680 行、import 43 → 26、横断シンボルの手配線は `hintCheckboxItem` (ゲーム固有設定と並べるため) を除いて 0。足した口は `cardNav` (カード番号キー操作、62 ページが `useCardKeyboardNav` を使う)、`FooterHand` / `FooterStatus`、`actionsTutorialId`、`settingsPlacement` (設定パネルが盤面の上か下かは 169 / 172 ページで二分)、`landscapeBanner` (91 ページだけが出すので既定は false)。`HeartsPage.test.tsx` と `e2e/hearts.spec.ts` は無変更で通過。

## Consequences

次の横断機能は `createGamePage` 1 ファイルの変更で済む。一斉移行はしない: 新規ゲームは new-game スキルから定義で作る、既存ページは触るついでに移行。ページ単位の guard は、移行済みページが増えたら `createGamePage` 1 ファイルの検査に置き換えられる（「387 ページ → 1 ファイル」になった guard 数を成果指標にする）。

まだ表現できないもの: CPU ログやモバイル/デスクトップで分かれるサイドバー — Hearts では不要だった (サイドバーは `lgTwoColGrid` を Board 内で完結)。必要になるページを移すときに定義へ足す (#11915)。

スコープ外: フロントのルール二重実装 (#11913)、`types/` の `openapi.yaml` からの生成。
