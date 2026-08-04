import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import golden from './__fixtures__/guandanCombo.golden.json';
import { guandanEvaluate } from './guandanCombo';

/**
 * Differential fixture: 220 selections classified by the Go evaluator, covering
 * all eleven combo kinds. `guandanEvaluate` is a hand port of
 * `GuandanEvaluate`, and a preview that disagrees with the server is worse than
 * no preview at all — so the port is pinned against the original's own output
 * rather than against a second reading of the rules.
 *
 * Regenerate after changing `internal/domain/Guandan.go`:
 *
 * ```sh
 * GD_FIXTURE=/tmp/gd.json go test -tags test ./internal/domain -run TestGenGuandanFixture
 * jq -c '[group_by(.kind)[]|.[0:24]]|flatten|map({c:[.cards[]|(.design[0:1]+(.value|tostring))],l:.level,k:.kind,r:.rank,s:.size})' \
 *   /tmp/gd.json > frontend/src/utils/__fixtures__/guandanCombo.golden.json
 * ```
 */
const DESIGNS: Readonly<Record<string, CardDesign>> = {
  S: 'SPADE',
  H: 'HEART',
  D: 'DIAMOND',
  C: 'CLOVER',
  J: 'JOKER',
};

function decode(token: string): Card {
  const design = DESIGNS[token[0] ?? ''];
  if (!design) throw new Error(`bad fixture token: ${token}`);
  return { design, value: Number(token.slice(1)) };
}

describe('guandanEvaluate against the Go evaluator', () => {
  it('covers every combo kind, so a regression cannot hide in an untaken branch', () => {
    expect(new Set(golden.map((g) => g.k)).size).toBe(11);
    expect(golden.length).toBeGreaterThanOrEqual(200);
  });

  it('agrees with the Go evaluator on kind, rank and size for every case', () => {
    const mismatches = golden
      .map((g) => ({ g, got: guandanEvaluate(g.c.map(decode), g.l) ?? { kind: 0, rank: 0, size: 0 } }))
      .filter(({ g, got }) => got.kind !== g.k || got.rank !== g.r || got.size !== g.s)
      .map(({ g, got }) => ({ cards: g.c, level: g.l, want: [g.k, g.r, g.s], got: [got.kind, got.rank, got.size] }));
    expect(mismatches).toEqual([]);
  });
});
