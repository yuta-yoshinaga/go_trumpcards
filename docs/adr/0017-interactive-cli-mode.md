# ADR-0017: Interactive CLI mode with game switching

## Status

Accepted

## Date

2026-03-11

## Context

ゲームを切り替えるたびにCLIを再起動する必要があり、ユーザー体験が悪かった。

## Decision

引数なしで起動するインタラクティブモードを追加:

- `switch <game>` コマンドでゲーム切替
- `games` コマンドで利用可能なゲーム一覧表示
- マップベースのディスパッチ（switchステートメントを置換）
- Levenshteinベースのタイポ提案（ゲーム名の入力ミス時）

## Consequences

- ゲーム間のシームレスな切替が可能
- マップベースのディスパッチにより新しいゲームの追加が1行で完了
- タイポ提案によりUXが向上
