# ADR-0006: OpenAPI specification as API contract

## Status

Accepted

## Date

2026-02-21

## Context

エンドポイント数が増加するにつれ、API仕様の正式なドキュメントが必要になった。開発者間でのAPI契約の共有と、ドキュメントの一元管理が求められた。

## Decision

`api/openapi.yaml` をREST APIの唯一の信頼できるソース（Single Source of Truth）とする。エンドポイントやスキーマの変更時は必ずOpenAPI仕様を更新する。

## Consequences

- API仕様が形式的に定義され、フロントエンド開発者との認識齟齬を防止
- ドキュメントメンテナンスルールにより、コード変更とドキュメント更新が同期
- OpenAPI仕様の保守コストが発生するが、11エンドポイントの規模では管理可能
