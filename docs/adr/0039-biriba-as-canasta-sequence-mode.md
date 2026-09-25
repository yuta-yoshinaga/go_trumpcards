# ADR-0039: Biriba を Canasta のシーケンスモードとして実装

## Status

Accepted

## Date

2026-09-10

## Context

Biriba は同じスートの連続ランクによるシーケンスが主役である。一方、Canasta
ドメインの `validateNewMeld` は同ランクのセットしか認めていない。
Canasta と Burraco・Biriba は Canasta ドメインを共有している（エイリアス）。
Samba と Bolivia は独立したドメイン型で、同一の純粋ロジックだけを
`samba_bolivia_shared.go` の共有関数にしている (#8056)。

検討した選択肢は次のとおりである。

1. Canasta を複製して独立ドメインにする。1670 行規模の複製になり、さらに
   当時は TinyGo の `json.Marshaler` / `json.Unmarshaler` 保持問題により、
   全 Worker が不要な直列化コードを抱える懸念があったため却下した（当時の懸念。
   現在は Worker ごとのビルドタグで domain が分かれているため、この点は当てはまらない。
   1,670 行規模の複製になることが主な理由）。
2. Burraco と同じくエイリアスと Pozzetto だけを追加する。セット制のままでは Biriba が
   Burraco のほぼ別名となり、issue #7093 の「単なる別名ではない」という要件を満たさないため却下した。
3. Canasta に `UseBiriba` でシーケンス制メルドを追加し、Biriba はエイリアスで公開する。

## Decision

3 を採用する。`UseBiriba` が有効なとき、同スートのシーケンス、106 枚デッキ、2 山の
Pozzetto、および純ビリバのボーナスを有効にする。Biriba は `type Biriba = Canasta` などの
エイリアスだけで公開し、新しいドメイン型を増やさない。

Biriba は2山の Pozzetto を必須とするため、`UseBiriba` は `UsePozzetto` を含意する共通判定に
した。これにより設定利用者が両方のフラグを立てる必要はなく、既存の `UsePozzetto` の動作は
変更しない。

## Consequences

共有ドメインが2つのメルド規則を持つことになる。既定（`UseBiriba=false`）の経路は不変であり、
Biriba テストの否定対照でセット制とシーケンス拒否を守る。新しい型は増えないため、Worker の
サイズ増加は最小限になる。
