import { describe, expect, it } from 'vitest';
import {
  applyTriPeaksScore,
  chainRemovalPoints,
  initialTriPeaksScore,
  TRIPEAKS_PEAK_BONUS,
  TRIPEAKS_POINTS_PER_CHAIN,
  type TriPeaksSnapshot,
} from './tripeaksScore';

const snap = (moveCount: number, stockCount: number, peaksCleared = 0): TriPeaksSnapshot => ({
  moveCount,
  stockCount,
  peaksCleared,
});

describe('chainRemovalPoints', () => {
  it('scales linearly with chain length', () => {
    expect(chainRemovalPoints(1)).toBe(TRIPEAKS_POINTS_PER_CHAIN);
    expect(chainRemovalPoints(3)).toBe(3 * TRIPEAKS_POINTS_PER_CHAIN);
  });
});

describe('applyTriPeaksScore', () => {
  it('returns the initial state when moveCount is 0 (fresh game)', () => {
    const result = applyTriPeaksScore({ score: 999, chain: 4 }, snap(5, 10), snap(0, 24));
    expect(result).toEqual(initialTriPeaksScore);
  });

  it('makes no change before the first snapshot (prev null)', () => {
    const state = { score: 100, chain: 1 };
    expect(applyTriPeaksScore(state, null, snap(2, 10))).toBe(state);
  });

  it('adds chain × POINTS_PER_CHAIN for a consecutive removal', () => {
    const first = applyTriPeaksScore(initialTriPeaksScore, snap(3, 10), snap(4, 10));
    expect(first).toEqual({ score: 100, chain: 1 });
    const second = applyTriPeaksScore(first, snap(4, 10), snap(5, 10));
    expect(second).toEqual({ score: 100 + 200, chain: 2 });
  });

  it('adds a peak bonus for each newly cleared peak', () => {
    const result = applyTriPeaksScore({ score: 0, chain: 0 }, snap(4, 10, 0), snap(5, 10, 1));
    expect(result).toEqual({ score: TRIPEAKS_POINTS_PER_CHAIN + TRIPEAKS_PEAK_BONUS, chain: 1 });
  });

  it('resets the chain but keeps the score on a draw (stock change)', () => {
    const result = applyTriPeaksScore({ score: 300, chain: 2 }, snap(5, 10), snap(5, 9));
    expect(result).toEqual({ score: 300, chain: 0 });
  });

  it('resets the chain but keeps the score on an undo (moveCount decrease)', () => {
    const result = applyTriPeaksScore({ score: 300, chain: 2 }, snap(5, 10), snap(4, 10));
    expect(result).toEqual({ score: 300, chain: 0 });
  });

  it('leaves state unchanged when nothing relevant moved', () => {
    const state = { score: 300, chain: 2 };
    expect(applyTriPeaksScore(state, snap(5, 10), snap(5, 10))).toBe(state);
  });

  it('never awards a negative peak bonus when a peak count decreases (undo refill)', () => {
    // moveCount rising while peaksCleared drops (an escape/undo edge): no negative bonus.
    const result = applyTriPeaksScore({ score: 0, chain: 0 }, snap(4, 10, 2), snap(5, 10, 1));
    expect(result).toEqual({ score: TRIPEAKS_POINTS_PER_CHAIN, chain: 1 });
  });
});
