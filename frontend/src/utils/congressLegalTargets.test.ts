import { describe, expect, it } from 'vitest';
import type { Card } from '../types/common';
import { congressLegalTargets } from './congressLegalTargets';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('congressLegalTargets', () => {
  it('finds descending tableau and same-suit ascending foundation targets', () => {
    const result = congressLegalTargets(
      [[card('HEART', 6)], [], [card('SPADE', 1)]],
      [[card('SPADE', 4)], [], [], [], [], [], [], []],
      card('SPADE', 5),
      'tableau',
    );

    expect(result.tableau).toEqual(new Set([0]));
    expect(result.foundation).toEqual(new Set([0]));
  });

  it('does not highlight an empty tableau for a tableau source', () => {
    const result = congressLegalTargets([[], [card('HEART', 8)]], [], card('SPADE', 9), 'tableau');

    expect(result.tableau).toEqual(new Set());
  });

  it('highlights empty tableau piles for a stock source', () => {
    const result = congressLegalTargets([[], [card('HEART', 8)], []], [], null, 'stock');

    expect(result.tableau).toEqual(new Set([0, 2]));
    expect(result.foundation).toEqual(new Set());
  });

  it('does not allow a card onto an ace or a full foundation', () => {
    const fullFoundation = Array.from({ length: 13 }, (_, i) => card('SPADE', i + 1));
    const result = congressLegalTargets(
      Array.from({ length: 8 }, () => [card('SPADE', 1)]),
      [fullFoundation],
      card('HEART', 2),
      'waste',
    );

    expect(result.tableau).toEqual(new Set());
    expect(result.foundation).toEqual(new Set());
  });
});
