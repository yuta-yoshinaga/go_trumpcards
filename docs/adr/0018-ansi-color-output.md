# ADR-0018: ANSI color output for CLI with TTY auto-detection

## Status

Accepted

## Date

2026-03-14

## Context

CUIのテキスト出力がプレーンテキストのみで、視覚的に情報を把握しにくかった。

## Decision

- 全CUIプレゼンターにANSIカラー出力を追加
- TTY自動検出: `stdout` がターミナルでない場合はカラーを無効化

## Consequences

- ターミナルでの視認性が向上（カード、メッセージ、エラー等の色分け）
- パイプやリダイレクト時にはカラーコードが含まれず、プログラム的な処理が容易
- `golang.org/x/term` への依存が追加
