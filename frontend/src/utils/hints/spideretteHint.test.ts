import { describe, expect, it } from 'vitest';
import { getSpideretteHint } from './spideretteHint';

describe('getSpideretteHint', () => {
  it('returns null for any state (intentional stub)', () => {
    expect(getSpideretteHint(null)).toBeNull();
    expect(getSpideretteHint(undefined)).toBeNull();
    expect(
      getSpideretteHint({
        tableau: [],
        stockCount: 0,
        completedSuits: 0,
        score: 500,
        phase: 0,
        moveCount: 0,
        canUndo: false,
        isStalemate: false,
        message: '',
      }),
    ).toBeNull();
  });
});
