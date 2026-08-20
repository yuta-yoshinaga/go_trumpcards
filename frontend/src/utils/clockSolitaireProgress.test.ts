import { describe, expect, it } from 'vitest';
import { CLOCK_PILE_COUNT, completedClockPiles } from './clockSolitaireProgress';

describe('completedClockPiles', () => {
  it('counts only the piles with all four cards face up', () => {
    expect(completedClockPiles([4, 3, 4, 0, 4])).toBe(3);
  });

  it('is 0 at the start', () => {
    expect(completedClockPiles(Array(CLOCK_PILE_COUNT).fill(0))).toBe(0);
  });

  // **クリア時は 13/13。**中央のKの山も数に入る。
  it('is 13 once every pile is done', () => {
    expect(completedClockPiles(Array(CLOCK_PILE_COUNT).fill(4))).toBe(CLOCK_PILE_COUNT);
  });

  // 3枚では完成でない -- 境界を off-by-one で数えると進捗が先走る。
  it('does not count a pile of three', () => {
    expect(completedClockPiles([3, 3, 3])).toBe(0);
  });
});
