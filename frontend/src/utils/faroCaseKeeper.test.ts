import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { FARO_RANK_COUNT, faroCardKey, mergeSeenCards, remainingByRank } from './faroCaseKeeper';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('faroCardKey', () => {
  it('builds a unique key from suit design and value', () => {
    expect(faroCardKey(card('SPADE', 3))).toBe('SPADE-3');
    expect(faroCardKey(card('HEART', 13))).toBe('HEART-13');
  });
});

describe('mergeSeenCards', () => {
  it('adds revealed cards to the running set', () => {
    const seen = mergeSeenCards(new Set(), [card('SPADE', 3), card('HEART', 7)]);
    expect(seen.has('SPADE-3')).toBe(true);
    expect(seen.has('HEART-7')).toBe(true);
    expect(seen.size).toBe(2);
  });

  it('ignores null and undefined entries', () => {
    const seen = mergeSeenCards(new Set(), [null, card('SPADE', 3), undefined]);
    expect(seen.size).toBe(1);
    expect(seen.has('SPADE-3')).toBe(true);
  });

  it('deduplicates a card revealed again by an idempotent refresh', () => {
    const first = mergeSeenCards(new Set(), [card('SPADE', 3)]);
    const second = mergeSeenCards(first, [card('SPADE', 3), card('CLOVER', 3)]);
    expect(second.size).toBe(2);
  });

  it('does not mutate the previous set', () => {
    const prev = new Set(['SPADE-3']);
    mergeSeenCards(prev, [card('HEART', 7)]);
    expect(prev.size).toBe(1);
  });
});

describe('remainingByRank', () => {
  it('starts every rank at four when nothing has been seen', () => {
    const remaining = remainingByRank(new Set());
    for (let rank = 1; rank <= 13; rank++) {
      expect(remaining[rank]).toBe(FARO_RANK_COUNT);
    }
  });

  it('subtracts one per distinct seen card of a rank', () => {
    const seen = new Set(['SPADE-3', 'HEART-3', 'SPADE-7']);
    const remaining = remainingByRank(seen);
    expect(remaining[3]).toBe(2);
    expect(remaining[7]).toBe(3);
    expect(remaining[5]).toBe(4);
  });

  it('clamps at zero when all four of a rank are seen', () => {
    const seen = new Set(['SPADE-9', 'HEART-9', 'CLOVER-9', 'DIAMOND-9']);
    const remaining = remainingByRank(seen);
    expect(remaining[9]).toBe(0);
  });

  it('ignores keys whose rank is out of range', () => {
    const seen = new Set(['JOKER-0', 'SPADE-14', 'HEART-5']);
    const remaining = remainingByRank(seen);
    expect(remaining[5]).toBe(3);
    expect(remaining[0]).toBeUndefined();
    expect(remaining[14]).toBeUndefined();
  });
});
