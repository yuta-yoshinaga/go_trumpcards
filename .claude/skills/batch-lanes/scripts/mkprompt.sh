#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# usage: mkprompt.sh <issue#>  -> prints delegation prompt
n=$1
body=$(cd "$REPO" && gh issue view $n --json title,body -q '"# "+.title+"\n\n"+.body')
cat <<P
GitHub issue #$n を実装してください。issue 本文:

$body

## 手順
0. **必要な値が Web 応答に無いことは NOT_PLANNED の理由にならない。** その場合はドメインに getter を足し、Web 応答 (controller の json・presenter・api/openapi.yaml・frontend/src/types・stateFactories) に足して実装する (#9530 / #9609 で 2 回止まった)。
1. **まず前提を実コードで確かめること。** 提案が既に実装済み、または前提が誤っている (その要素は既に存在する、共有 hook/子コンポーネントで既に満たされている等) なら、**ファイルを一切変更せず**、最終報告の先頭行に \`VERDICT: NOT_PLANNED\` と書き、根拠をファイル:行で示すこと。
2. 前提が正しければ TDD で実装する。最終報告の先頭行は \`VERDICT: IMPLEMENTED\`。
   - ページ変更は \`frontend/src/pages/<X>Page.test.tsx\` に試験を足す (renderWithProviders を使う既存形に倣う)。
   - 文言は必ず i18n。\`frontend/src/i18n/locales/{ja,en}/<game>.json\` を**キー単位で対に**更新する。ja の値に英語を残さない。既存の近い文言・キー名を先に grep して揃える。
   - Go を変えるなら \`internal/i18n/locales/{ja,en}/\` も対に。試験は \`go test -tags test ./internal/<該当pkg>/...\`。
   - 文字列を列挙で繋ぐときは区切りをハードコードせず、ゲームの ja/en に listSeparator (ja「、」/ en ", ") を置いて t('listSeparator') を使う。
   - 番号 (列・席・扇など) を表示/読み上げに出すときは、そのゲームの棋譜・CUI・既存の aria-label と同じ体系 (0 始まりか 1 始まりか) に揃える。
   - 新しいライブ領域 (role=status / aria-live) は常設にし、中身だけを切り替える (条件付きでマウントしない)。常設の外枠には枠線・背景・余白を付けない (空のとき空の箱が見えるため)。見た目のスタイルは中身があるときだけ内側の要素に付ける。画面にすでに見えている文と同じ内容を読ませるだけのライブ領域は \`className="sr-only"\` にする (付け忘れると同じ文が画面に 2 重に出る)。
   - カード名を読み上げ・表示するときは共有の \`cardAlt(card)\` (\`frontend/src/utils/cardAlt.ts\`) を使う。スート名・ランク名の i18n キーをゲームのロケールに新設しない (#9008/#9039 で 2 回差し戻した)。
   - aria-disabled にしたカードの onClick ガードは、native \`disabled\` が既に塞いでいる条件 (loading・手番外) を再導出しない。\`if (playable) handle(idx)\` のように合法性だけを見る (\`state?.\` などの到達不能分岐は codecov の部分行になる)。
   - 応答 (state) のフィールドを新しく表示するときは、internal/domain/ でそのフィールドが何か (設定値か、累計か、現在値か) を必ず確かめる。名前だけで意味を推測しない (例: Whitehead の drawCount は「配った回数」ではなく 1/3 枚めくりの設定)。
   - 同じ札を札の同一性 (design/value) で差分・検索するときは、そのゲームのデッキ数を確かめる (2 デッキなら同じ札が 2 枚ある)。
   - frontend/src/components/ の共有コンポーネントを変えるときは、他のゲームの挙動を一切変えない形 (省略可能な prop を足し、渡したページだけで効く) にする。共有コンポーネントの中でゲーム固有の翻訳名前空間 (useTranslation('<game>')) を使わない。まずページ側だけで実装できないかを考える。
   - 「山札が空なら引けない/パス」のような表示をするときは、ドメインの引く処理が捨て札から山札を作り直す (recycle) かを確かめる。作り直すなら山札 0 だけを条件にしない。
   - ボタンの disabled 条件を足すときは、その操作がドメイン上で唯一の進行手段 (例: 引けないときの山札ボタンが実はパス) でないかを確かめる。塞ぐと進行不能になる操作は disabled にしない。
   - Go の試験で i18n.SetLang を変えたら、必ず t.Cleanup で元の言語に戻す (戻さないと同じパッケージの別の試験が落ちる)。Go を変えたら該当パッケージ全体の go test を -run で絞らずに 1 回回す。
   - Go のインターフェース (internal/domain/interfaces/*.go) にメソッドを足したら、同じディレクトリのモック (*_mock.go) にも同じメソッドを足す。go vet -tags test ./internal/... で確かめる。
   - 段階の途中の状態 (例: 決定済み・待機中) を表示するときは、ドメインがその操作の同じ呼び出しの中で段階を進めてしまわないかを確かめる。画面に届かない状態のための表示は作らない。
   - **Web 応答に項目を足したら** (internal/adapter/controller/*WebController.go の json タグ)、同じ変更で api/openapi.yaml (そのゲームの \`<Game>Response\` schema の中に足す。\`grep -n "^    <Game>Response:"\` で位置を確かめ、別ゲームの同名プロパティの隣に挿さない。折り畳みスカラー \`>-\` の途中に挿さない)・frontend/src/types・frontend/src/test/stateFactories.ts・golden-body 試験ヘルパにも足し、\`go test -tags test ./internal/infrastructure/games/\` (openapi と実応答の突き合わせ) を必ず通す。ページが新しい state.X を読むなら、それを送る json タグがあることを確かめる。
   - **CPU の手番を示す表示 (「CPU の番」「思考中」「待機中」) を作る前に**、その CPU がどう動くかを確かめる: usecase の runCpuTurns / runCpuTurnsLoop やドメインの advanceCpu / runCpuActions が人間の手番まで同じ応答内で CPU を進めるなら、CPU が手番の状態は画面に届かない (Rikken / Tute / Pineapple で 3 回)。その場合は VERDICT: NOT_PLANNED とし、ループの file:line を根拠に書く。画面側が 1 手ずつ CPU を進める (exec('cpu') など) ゲームだけが対象。
   - 画面に文字列を増やしたら (既存の用語の再利用を含む)、frontend/e2e/<game>*.spec.ts の getByText / getByRole の name がその文字列にも当たって複数一致 (strict mode violation) にならないかを grep で確かめ、当たるならロケータを testid 等で絞る。
   - 操作できない札・ボタンを示すときに native の \`disabled\` を新しく足さない (リポジトリ方針 #8524/#8787)。\`aria-disabled="true"\` + \`aria-describedby\` (理由の文) にしてフォーカス可能のまま残し、onClick の先頭で弾いて API を送らない (既存の native disabled を aria-disabled に置き換えるときは、**それまで disabled にしていた条件 (loading など) をすべて**ガードに含める。ガードを忘れると通信中の二重送信になる)。見た目も native disabled と同じく減光させる (\`disabled:opacity-*\` は aria-disabled に効かないので、先例 CrescentPage.tsx / BristolPage.tsx と同じ書き方で aria-disabled のときの減光・カーソルを付ける)。試験は「押しても exec が呼ばれない」と「aria-disabled が付く/付かない」の両側を書く。
   - **操作の可否を画面で判定する (無効化・候補表示) ときは、ドメインの合法判定関数 (isBidLegal / canPlay 等) が実際に比べている値と同じ値を使う。** 名前が近い応答フィールド (例: winningBid は競りが終わってから設定される確定値で、競り中は 0) で代用しない。その値が応答に無ければ、Web 応答に足す (下の「Web 応答に項目を足したら」に従う)。
   - \`<fieldset>\` を使うときは、リポジトリの他の fieldset と同じく \`border-0 p-0 m-0 min-w-0\` を付けてブラウザ既定の枠・余白を消す。
   - 複数の variant が共有するページ (例: PineapplePage / BlackJackPage / SevenCardStudPage。\`t\` は variant ごとの名前空間) に文言を足すときは、使う variant の名前空間すべてにキーを足すか、その要素ごと対象 variant のときだけ描画する。aria-label などの属性に未定義キーを渡すと生キーが読み上げられる。
   - 新しく出したパネル・ダイアログ (ログ等) が開いている間は、そのページのキーボードショートカット (useActionShortcuts の enabled) が実操作を送らないようにする。
   - 0 始まりの番号に「目」「番目」「第」などの序数を付けない (「0列目」になる)。ja は「列{{col}}」の形、en は「column {{col}}」。
   - 英語で件数を含む文は \`_one\` / \`_other\` のキーに分ける (先例 penguin.json tableauCardCount)。ja は 1 キーのまま。
   - 文の区切り (、 / ", ") を JSX のテンプレートリテラルにハードコードしない。区切りを含む文全体をロケールのテンプレートにする。
   - 訳語は同じゲームのロケールの既存の表記に合わせる (\`git grep <語> origin/develop -- frontend/src/i18n/locales/ja/<game>.json\` で確認。例: euchre は「バウアー」)。
   - 3 項目以上の新しいボタンや分岐を足すときも、既存のボタン・操作を隠したり意味を変えたりしない。
   - **画面の差分 (前回の state との比較) で「出た札」「勝者」を読み上げるときは**、usecase の Interactor が人間の操作の中で CPU を何手進めるか (runCpuTurns 等) を先に読む。1 回の応答で複数の札が増える・トリック確定と次のリードが同時に起きる前提で、増えた札すべてと勝者の両方を出す。
   - 札の画像 (CardImage の alt) を中に持つボタンに aria-label を足すと、alt はアクセシブル名に入らなくなる。札がある状態では aria-label にも札名 (cardAlt) を含める。
   - api/openapi.yaml にプロパティを足すときは、既存プロパティの \`type:\` と \`description:\` の間に挿入しない (挿入位置の前後 5 行を必ず確認。\`mapping key "description" already defined\` で YAML が壊れる事故が 2 回)。共有スキーマ (複数の paths が \$ref する) に足すと他ゲームでも宣言されるので、そのゲームの応答が参照するスキーマを paths から辿って確認する。
   - **issue の「対象ファイル」は目安であり制約ではない。** 規則由来の値 (得点・必要点・増減・履歴) を出すのにドメイン・presenter・WebController・api/openapi.yaml・TS 型・stateFactories の変更が要るなら、それらも変更してよい (途中で止めて報告するのではなく、そこまで実装して試験を足す)。値の計算は 1 か所 (ドメインで実際に点を動かす所) にだけ書き、表示用 getter で同じ計算を繰り返さない。
   - 払戻額・純損益・精算額をページで計算しない (アンテや持ち越しが入る)。ドメインの精算が実際に動かしたチップを Web 応答 (roundPayout / netChange 等) に載せて表示する (#8791 の方針)。
   - コール額・不足額などの金額を表示するときは、ドメインが実際に請求する式 (例: Anaconda の need := currentBet - streetBet) を grep し、同じフィールドを使う。名前の近い累計 (roundBet 等) で代用しない。
   - 要素を aria-disabled にしたら、frontend/e2e/<game>*.spec.ts でその要素をクリックしているロケータ (例 button[aria-pressed]:has(img) の first()) を grep し、:not([aria-disabled="true"]) で除く。Playwright の click は aria-disabled の要素が有効になるまで待ち続けてタイムアウトする (#9811)。
   - 成績・統計 (クリア回数・平均手数・勝敗など) を localStorage に残すときは、共有の frontend/src/hooks/createLocalStorageStats.ts とゲームごとの hooks/use<Game>Stats.ts (先例 useSomersetStats / useMrsMopStats) の形に合わせ、ページに localStorage の読み書きを直書きしない。フックに単体試験を付ける。
   - presenter が新しいゲッターを呼ぶようにしたら、internal/adapter/presenter/<Game>WebPresenter_test.go のモック設定 (m.On("…")) にも同じゲッターの期待値を足し、\`go test -tags test ./internal/adapter/presenter/\` を -run で絞らずに回す (#9427 で予期しない呼び出しにより落ちた)。
   - デザイントークンのみ (\`text-ds-*\`, \`bg-ds-*\`)。生のパレットや \`text-white/N\` は不可。
   - 追加した試験が**実装を壊すと落ちる**ことを確かめる (一時的に実装を戻して失敗を見て、元に戻す。戻しは手で行い git checkout を使わない)。
   - \`waitFor(() => expect(...).not...)\` のような、描画が空でも通る否定だけの assert を書かない。
   - サーバ側のヒント (domain の GetHint) を画面で使う issue: ページに hint ボタン + useHintRequest が既にあれば、その経路 (hint コマンド → HintOutput → res.hint) を直すだけにする。無ければ state 変化ごとに hint コマンドを別リクエストで叩く形にはせず、Web 出力 struct に serverHint (omitempty) を足して状態応答 (Output) に GetHint() を詰め (openapi・types・stateFactories も)、getXxxHint(state) で state.serverHint を HintResult に写す (BlackHole / TuSac が先例)。共有の useGameHint / FrontendHintTooltip は外さない。
   - Web 応答で札を伏せるとき: 偽の札 (design "BACK" など) を作らない (Card スキーマの enum 違反)。札は送らず空配列にし、枚数は定数か既存の count フィールドで表す。同じ札が別フィールド (例 dealerHighHand) から漏れていないか presenter の全フィールドを確認する。
   - 判定ロジックをドメインへ移して フロントの試験を消すときは、消した試験の場面を同じ数だけドメイン試験に移植する。
   - Go の presenter / domain を触ったら、そのファイルの //go:build にある worker タグで GOOS=js GOARCH=wasm go build -tags <worker> -o /dev/null ./cmd/workers/<worker> を通す。別タグのファイルの関数は使わず、共有したい関数はタグ無しの共有ヘルパへ移す。
   - 新しく足した関数・export は、定義以外から使われていることを git grep で確かめる (使われないものは足さない)。
3. 受け入れ条件 (issue の各項目) を 1 対 1 で満たすこと。範囲外の改名・整形・無関係なファイル変更をしない。

## 検証 (すべて通すこと)
- \`cd frontend && bunx vitest run <変更したテストファイル>\` (**-t で絞らずファイル丸ごと**。アクセシブル名を変えると既存試験のクエリが落ちる: #9631 で 12 本)
- \`cd frontend && bun run check && bun run typecheck\`
- Go を変えた場合: \`goimports -w <file>\`、\`go build ./...\`、該当パッケージの \`go test -tags test\`。golangci-lint は実行しないこと (並列不可)。

## 契約
**コミットは絶対にしないこと。** 一時ファイルは必ず削除すること。
**実装を足すなら同じ変更に単体試験も足すこと** (CLAUDE.md の必須要件)。
既存の未コミット変更はそのまま保持すること。git reset / checkout / stash / restore を実行しないこと。
最後に \`git status --porcelain\` と \`git log --oneline -1\` を報告すること。
P
