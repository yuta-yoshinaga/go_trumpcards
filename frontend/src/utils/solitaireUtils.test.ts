import { describe, expect, it } from 'vitest';
import { isInMoveBlock, isTableauAllFaceUp } from './solitaireUtils';

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

describe('isInMoveBlock', () => {
  it('returns false when nothing is selected', () => {
    expect(isInMoveBlock(null, 0, 0)).toBe(false);
  });

  it('includes the selected card and every card below it in the same column', () => {
    const selected = { col: 2, cardIndex: 3 };
    expect(isInMoveBlock(selected, 2, 3)).toBe(true); // the selected card itself
    expect(isInMoveBlock(selected, 2, 4)).toBe(true); // one below
    expect(isInMoveBlock(selected, 2, 9)).toBe(true); // further below
  });

  it('excludes cards above the selection', () => {
    const selected = { col: 2, cardIndex: 3 };
    expect(isInMoveBlock(selected, 2, 2)).toBe(false);
    expect(isInMoveBlock(selected, 2, 0)).toBe(false);
  });

  it('excludes cards in a different column', () => {
    const selected = { col: 2, cardIndex: 3 };
    expect(isInMoveBlock(selected, 1, 3)).toBe(false);
    expect(isInMoveBlock(selected, 3, 4)).toBe(false);
  });
});
