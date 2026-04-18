import { describe, expect, it } from 'vitest';
import { isTableauAllFaceUp } from './solitaireUtils';

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
