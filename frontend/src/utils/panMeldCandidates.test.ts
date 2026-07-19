import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { type PanMeldCandidate, panLayoffIndices, panMeldCandidates } from './panMeldCandidates';

const c = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

/** Sorts candidates into a stable, comparison-friendly order. */
function normalize(cands: PanMeldCandidate[]): PanMeldCandidate[] {
  return [...cands]
    .map((cand) => ({ kind: cand.kind, indices: [...cand.indices].sort((a, b) => a - b) }))
    .sort((a, b) => a.indices.join(',').localeCompare(b.indices.join(',')));
}

describe('panMeldCandidates', () => {
  it('detects a set of three of the same rank (duplicate suits allowed)', () => {
    const cards = [c('SPADE', 5), c('HEART', 5), c('SPADE', 5), c('DIAMOND', 2)];
    expect(normalize(panMeldCandidates(cards))).toEqual([{ kind: 'set', indices: [0, 1, 2] }]);
  });

  it('detects a same-suit run of three regardless of input order', () => {
    const cards = [c('SPADE', 6), c('SPADE', 4), c('SPADE', 5), c('HEART', 13)];
    expect(normalize(panMeldCandidates(cards))).toEqual([{ kind: 'run', indices: [0, 1, 2] }]);
  });

  it('treats 7 and J as adjacent (8/9/10 removed from the deck)', () => {
    // 7-J-Q is a valid Pan run because the 8, 9 and 10 are not in the deck.
    const cards = [c('CLOVER', 7), c('CLOVER', 11), c('CLOVER', 12)];
    expect(normalize(panMeldCandidates(cards))).toEqual([{ kind: 'run', indices: [0, 1, 2] }]);
  });

  it('allows an Ace-low run (A-2-3)', () => {
    const cards = [c('DIAMOND', 1), c('DIAMOND', 2), c('DIAMOND', 3)];
    expect(normalize(panMeldCandidates(cards))).toEqual([{ kind: 'run', indices: [0, 1, 2] }]);
  });

  it('allows an Ace-high run (Q-K-A)', () => {
    const cards = [c('HEART', 12), c('HEART', 13), c('HEART', 1)];
    expect(normalize(panMeldCandidates(cards))).toEqual([{ kind: 'run', indices: [0, 1, 2] }]);
  });

  it('does NOT suggest a two-card pair or partial run', () => {
    expect(panMeldCandidates([c('SPADE', 5), c('HEART', 5)])).toEqual([]);
    expect(panMeldCandidates([c('SPADE', 4), c('SPADE', 5)])).toEqual([]);
  });

  it('does NOT connect a run across different suits', () => {
    expect(panMeldCandidates([c('SPADE', 4), c('HEART', 5), c('SPADE', 6)])).toEqual([]);
  });

  it('does NOT treat 6-7-J as a run (would wrap over the removed 8/9/10 only if adjacent)', () => {
    // 6-7 are adjacent, 7-J are adjacent, but 6-7-J is 3 consecutive indices (5,6,7) -> valid.
    // The invalid near-miss is 6-7 then a gap: 6-J-Q skips 7.
    expect(panMeldCandidates([c('SPADE', 6), c('SPADE', 11), c('SPADE', 12)])).toEqual([]);
  });

  it('does NOT wrap K-A-2 around the deck', () => {
    // K(9)-A(high 10) is adjacent, but 2(1) is not — no wrap-around run.
    expect(panMeldCandidates([c('SPADE', 13), c('SPADE', 1), c('SPADE', 2)])).toEqual([]);
  });

  it('does NOT form a run from same-suit cards with a duplicate rank', () => {
    // 4-4-5 same suit: duplicate rank breaks the run, and 3 distinct ranks are absent.
    expect(panMeldCandidates([c('SPADE', 4), c('SPADE', 4), c('SPADE', 5)])).toEqual([]);
  });

  it('finds both a set and a run in the same hand', () => {
    const cards = [
      c('SPADE', 5),
      c('HEART', 5),
      c('DIAMOND', 5), // set of 5s -> indices 0,1,2
      c('CLOVER', 3),
      c('CLOVER', 4),
      c('CLOVER', 5), // run 3-4-5 clovers -> indices 3,4,5
    ];
    expect(normalize(panMeldCandidates(cards))).toEqual([
      { kind: 'set', indices: [0, 1, 2] },
      { kind: 'run', indices: [3, 4, 5] },
    ]);
  });
});

describe('panLayoffIndices', () => {
  it('marks a card that extends an existing set', () => {
    const hand = [c('SPADE', 4), c('HEART', 9)];
    const melds = [[c('DIAMOND', 4), c('HEART', 4), c('CLOVER', 4)]]; // set of 4s
    expect(panLayoffIndices(hand, melds)).toEqual(new Set([0]));
  });

  it('marks a card that extends an existing run at either end', () => {
    const hand = [c('SPADE', 6), c('SPADE', 2), c('SPADE', 7)];
    const melds = [[c('SPADE', 3), c('SPADE', 4), c('SPADE', 5)]]; // run 3-4-5 spades
    // 6 extends the top, 2 extends the bottom; 7 does not (gap after 5-6 would need the 6 first).
    expect(panLayoffIndices(hand, melds)).toEqual(new Set([0, 1]));
  });

  it('does NOT mark a card of a different suit for a run', () => {
    const hand = [c('HEART', 6)];
    const melds = [[c('SPADE', 3), c('SPADE', 4), c('SPADE', 5)]];
    expect(panLayoffIndices(hand, melds)).toEqual(new Set());
  });

  it('does NOT mark a card of a different rank for a set', () => {
    const hand = [c('SPADE', 5)];
    const melds = [[c('DIAMOND', 4), c('HEART', 4), c('CLOVER', 4)]];
    expect(panLayoffIndices(hand, melds)).toEqual(new Set());
  });

  it('returns empty when there are no table melds', () => {
    expect(panLayoffIndices([c('SPADE', 4)], [])).toEqual(new Set());
  });
});
