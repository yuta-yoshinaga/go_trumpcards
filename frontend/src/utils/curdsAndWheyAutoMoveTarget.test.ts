import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { curdsAndWheyAutoMoveTarget } from './curdsAndWheyAutoMoveTarget';

const card = (design: CardDesign, value: number): Card => ({ design, value });

/** Builds 10 columns from a sparse map so tests only spell out the relevant piles. */
const board = (piles: Record<number, Card[]>): Card[][] => Array.from({ length: 10 }, (_, i) => piles[i] ?? []);

describe('curdsAndWheyAutoMoveTarget', () => {
  it('returns null when the source card index is out of range', () => {
    const columns = board({ 0: [card('SPADE', 5)] });
    expect(curdsAndWheyAutoMoveTarget(columns, 0, 3)).toBeNull();
  });

  it('prefers a same-suit link over a same-rank link and an empty column', () => {
    // Moving ♠7 (col 0). col 1 = ♥7 (same rank), col 2 = ♠8 (same suit), col 3 empty.
    const columns = board({
      0: [card('CLOVER', 9), card('SPADE', 7)],
      1: [card('HEART', 7)],
      2: [card('DIAMOND', 10), card('SPADE', 8)],
      3: [],
    });
    expect(curdsAndWheyAutoMoveTarget(columns, 0, 1)).toBe(2);
  });

  // Simple Simon accepted "one rank higher, any suit"; Curds and Whey does not.
  // A ♥8 is no longer a destination for a ♠7 -- only an empty column is left.
  it('does not offer a different-suit card one rank higher', () => {
    // ♥8 in col 3 was Simple Simon's "rank-only" destination and must now be
    // ignored, leaving only the empty-column fallback -- which is column 1,
    // the lowest empty one, since board() leaves every unnamed column empty.
    const columns = board({
      0: [card('CLOVER', 9), card('SPADE', 7)],
      3: [card('HEART', 8)],
    });
    expect(curdsAndWheyAutoMoveTarget(columns, 0, 1)).toBe(1);
  });

  it('falls back to a same-rank link when no same-suit link exists', () => {
    const columns = board({
      0: [card('CLOVER', 9), card('SPADE', 7)],
      3: [card('HEART', 7)],
      5: [],
    });
    expect(curdsAndWheyAutoMoveTarget(columns, 0, 1)).toBe(3);
  });

  it('picks the lowest column index among equal same-rank links', () => {
    const columns = board({
      0: [card('CLOVER', 9), card('SPADE', 7)],
      6: [card('HEART', 7)],
      2: [card('DIAMOND', 7)],
    });
    expect(curdsAndWheyAutoMoveTarget(columns, 0, 1)).toBe(2);
  });

  it('falls back to the lowest empty column only when the run is not the whole pile', () => {
    // ♠7 sits under ♣9, so cardIndex 1 leaves a card behind: empty col is allowed.
    // Fill cols 1-3 so col 4 is the only (and thus lowest) empty column.
    const columns = board({
      0: [card('CLOVER', 9), card('SPADE', 7)],
      1: [card('HEART', 2)],
      2: [card('HEART', 2)],
      3: [card('HEART', 2)],
      4: [],
    });
    expect(curdsAndWheyAutoMoveTarget(columns, 0, 1)).toBe(4);
  });

  it('does not send a whole column onto an empty column (no progress)', () => {
    // Moving the entire col 0 (cardIndex 0) to an empty col makes no progress.
    const columns = board({
      0: [card('SPADE', 7)],
      4: [],
    });
    expect(curdsAndWheyAutoMoveTarget(columns, 0, 0)).toBeNull();
  });

  it('never targets the source column even if its top would otherwise link', () => {
    // col 0 top ♠8 links to ♠7, but ♠7 is IN col 0. Fill every other column so no
    // empty fallback and no external value-8 top exists → no destination.
    const columns = board({
      0: [card('SPADE', 8), card('SPADE', 7)],
      1: [card('HEART', 2)],
      2: [card('HEART', 2)],
      3: [card('HEART', 2)],
      4: [card('HEART', 2)],
      5: [card('HEART', 2)],
      6: [card('HEART', 2)],
      7: [card('HEART', 2)],
      8: [card('HEART', 2)],
      9: [card('HEART', 2)],
    });
    expect(curdsAndWheyAutoMoveTarget(columns, 0, 1)).toBeNull();
  });

  it('returns null when no legal destination exists', () => {
    const columns = board({
      0: [card('SPADE', 7)],
      1: [card('HEART', 3)],
      2: [card('CLOVER', 5)],
    });
    expect(curdsAndWheyAutoMoveTarget(columns, 0, 0)).toBeNull();
  });
});
