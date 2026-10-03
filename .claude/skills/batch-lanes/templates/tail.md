
## 共通の規則
- **まず前提を実コードで確かめること。** issue の主張 (件数・行番号・関数名) は古い SHA で書かれている。既に実装済み・前提が誤りなら**ファイルを一切変更せず**、最終報告の先頭行に `VERDICT: NOT_PLANNED` と根拠 (file:line) を書く。前提が一部だけ誤りなら、正しい部分を実装し、報告に差異を書く。実装したら先頭行は `VERDICT: IMPLEMENTED`。
- TDD: 試験を先に書き、**実装を一時的に戻すと試験が落ちる**ことを確かめてから戻す (戻しは手で行い git checkout を使わない)。
- 文言の assert はキー名ではなく**解決後の実際の文言**で書く (i18n.T は未定義キーでキー名を返すので、キー名の assert は常に通る)。Go の試験で i18n.SetLang を変えたら t.Cleanup で元に戻す。
- ロケールは ja/en を**キー単位で対に**更新する。ja の値に英語を残さない。使われなくなったキーは両方から消す。
- 範囲外の改名・整形・無関係なファイル変更をしない。dead code になったものは消す。
- Go を変えたら: `goimports -w <変更した .go>`、`go build ./...`、`go vet -tags test ./...`、`flock /tmp/ib-gotest.lock go test -tags test ./...` を**丸ごと 1 回** (flock は同時に走る別作業とのメモリ競合を避けるため。必ず付ける) (-run で絞らない。横断ガード試験が多数ある)。golangci-lint は実行しないこと (並列不可)。
- frontend を変えたら: `cd frontend && bunx vitest run <変更・影響した試験ファイル>` (-t で絞らない)、`bun run check`、`bun run typecheck` (素の tsc は使わない)。

## 契約
**コミットは絶対にしないこと。** 一時ファイルは必ず削除すること。
**実装を足すなら同じ変更に単体試験も足すこと** (CLAUDE.md の必須要件)。
既存の未コミット変更はそのまま保持すること。git reset / checkout / stash / restore を実行しないこと。
最後に `git status --porcelain` と `git log --oneline -1` を報告すること。
