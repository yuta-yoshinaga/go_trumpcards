# ADR-0011: i18n with react-i18next and browser language detection

## Status

Accepted

## Date

2026-03-05

## Context

日本語のみのUIでは国際的なユーザーが利用しにくい。Web GUI と CLI の両方で多言語対応が必要だった。

## Decision

- **Web GUI**: `react-i18next` + `i18next-browser-languagedetector` で日本語/英語をサポート
- **CLI**: `--lang` フラグで言語切替
- **翻訳ファイル**: `frontend/src/i18n/locales/{ja,en}/` にゲームごとに配置
- **サーバーレスポンス**: `messageCode` と `messageParams` を `message` と併せて返し、フロントエンド側で翻訳

## Consequences

- ブラウザの言語設定に基づく自動言語検出
- ゲームごとに翻訳ファイルが分離され、保守しやすい
- 新しいゲーム追加時に翻訳ファイル（ja/en）の作成が必要
- サーバーが `messageCode` を返すことで、翻訳ロジックがフロントエンドに集約
