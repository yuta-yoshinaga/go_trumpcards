import { describe, expect, it } from 'vitest';
import { isTableauAllFaceUp, spiderMovableRun } from './solitaireUtils';

/** Builds a face-up Spider tableau card from suit (`design`) + rank (`value`). */
const up = (design: string, value: number) => ({ card: { design, value }, faceUp: true });
/** Builds a face-down Spider tableau card (face hidden). */
const down = (design: string, value: number) => ({ card: { design, value }, faceUp: false });

describe('isTableauAllFaceUp', () => {
  it('returns true for an empty tableau', () => {
    expect(isTableauAllFaceUp([])).toBe(true);
  });

  it('returns true when every cell is face-up', () => {
    expect(isTableauAllFaceUp([[{ faceUp: true }], [{ faceUp: true }, { faceUp: true }]])).toBe(true);
  });

  it('returns false when any cell is face-down', () => {
    expect(isTableauAllFaceUp([[{ faceUp: true }], [{ faceUp: false }, { faceUp: true }]])).toBe(false);
  });

  it('returns true when columns are empty arrays', () => {
    expect(isTableauAllFaceUp([[], [], []])).toBe(true);
  });
});

describe('spiderMovableRun', () => {
  it('rings the whole tail when it is a same-suit descending run to the bottom', () => {
    // ♠10 ♠9 ♠8 all face-up.
    const col = [up('spade', 10), up('spade', 9), up('spade', 8)];
    expect(spiderMovableRun(col, 0)).toEqual([0, 1, 2]);
  });

  it('rings only the valid suffix when a hovered card starts a shorter run', () => {
    const col = [up('spade', 10), up('spade', 9), up('spade', 8)];
    expect(spiderMovableRun(col, 1)).toEqual([1, 2]);
  });

  it('rings only itself for the bottom card', () => {
    const col = [up('spade', 10), up('spade', 9), up('spade', 8)];
    expect(spiderMovableRun(col, 2)).toEqual([2]);
  });

  it('returns empty when the tail breaks below the hovered card (suit mismatch)', () => {
    // ♠10 then ♥9 ♥8: grabbing ♠10 drags a broken sequence, so it is not movable.
    const col = [up('spade', 10), up('heart', 9), up('heart', 8)];
    expect(spiderMovableRun(col, 0)).toEqual([]);
    // The valid heart suffix still rings on its own.
    expect(spiderMovableRun(col, 1)).toEqual([1, 2]);
  });

  it('returns empty when the ranks are not strictly descending by one', () => {
    const col = [up('spade', 10), up('spade', 8)];
    expect(spiderMovableRun(col, 0)).toEqual([]);
  });

  it('returns empty for a face-down hovered card', () => {
    const col = [down('spade', 10), up('spade', 9)];
    expect(spiderMovableRun(col, 0)).toEqual([]);
  });

  it('returns empty when a card below the hovered card is face-down', () => {
    const col = [up('spade', 10), down('spade', 9)];
    expect(spiderMovableRun(col, 0)).toEqual([]);
  });

  it('returns empty for out-of-range or null-card indices', () => {
    const col = [up('spade', 10)];
    expect(spiderMovableRun(col, -1)).toEqual([]);
    expect(spiderMovableRun(col, 5)).toEqual([]);
    expect(spiderMovableRun([{ card: null, faceUp: true }], 0)).toEqual([]);
  });
});
