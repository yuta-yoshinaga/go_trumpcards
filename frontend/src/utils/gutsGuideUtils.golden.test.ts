import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import golden from './__fixtures__/gutsGuide.golden.json';
import { evaluateGutsGuide } from './gutsGuideUtils';

/**
 * The Guts declaration guide lives twice: `GutsEvaluateGuide` in
 * `internal/domain/Guts.go` (which the CUI shows) and this module (which the Web
 * page shows). These golden vectors are also asserted by
 * `TestGutsEvaluateGuide_GoldenVectors`, so changing the rules on one side alone
 * fails that side, and regenerating the vectors to fix it fails the other.
 */
const DESIGNS: Card['design'][] = ['JOKER', 'SPADE', 'CLOVER', 'HEART', 'DIAMOND'];

describe('evaluateGutsGuide golden vectors (shared with the Go domain)', () => {
  it('has vectors to check', () => {
    expect(golden.cases.length).toBeGreaterThan(0);
  });

  it.each(golden.cases)('$name', (c) => {
    const cards: Card[] = c.cards.map((cd) => ({ design: DESIGNS[cd.suit], value: cd.value }));

    const guide = evaluateGutsGuide(cards);

    expect(guide).not.toBeNull();
    expect(guide?.handKey).toBe(c.pair ? 'pair' : 'highcard');
    expect(guide?.tier).toBe(c.tier);
  });
});
