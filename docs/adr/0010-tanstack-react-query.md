# ADR-0010: TanStack React Query for API state management

## Status

Accepted

## Date

2026-03-04

## Context

カスタムの `useGameApi` フックで各ページがローディング・エラー状態を手動管理しており、ボイラープレートが多かった。

## Decision

TanStack React Query（`useMutation`）を採用し、API状態管理を統一する。`QueryClientProvider` でアプリ全体をラップ。

## Consequences

- ローディング・エラー・成功状態の管理が宣言的になり、ボイラープレート削減
- テストでは `renderWithProviders` で `QueryClientProvider` をラップする必要がある
- リトライ、キャッシュ等のReact Queryの機能を将来活用可能
