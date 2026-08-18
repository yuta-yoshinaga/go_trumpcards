import { describe, expect, it } from 'vitest';
import { cuckooSwapTarget } from './cuckooSwapTarget';

const seats = (eliminated: number[], count = 4) =>
  Array.from({ length: count }, (_, id) => ({ id, isEliminated: eliminated.includes(id) }));

describe('cuckooSwapTarget', () => {
  it('is the next seat when nobody is out', () => {
    expect(cuckooSwapTarget(seats([]), 0)).toBe(1);
    expect(cuckooSwapTarget(seats([]), 3)).toBe(0);
  });

  // **脱落者は飛ばす。**「隣」が席順の隣とは限らないのが、この表示が要る理由。
  it('skips eliminated seats', () => {
    expect(cuckooSwapTarget(seats([1]), 0)).toBe(2);
    expect(cuckooSwapTarget(seats([1, 2]), 0)).toBe(3);
  });

  it('wraps around the table', () => {
    expect(cuckooSwapTarget(seats([0]), 3)).toBe(1);
  });

  // 他に生き残りがいなければスワップは保持と同義 (domain の attemptSwap)。
  it('is null when no one else is left', () => {
    expect(cuckooSwapTarget(seats([1, 2, 3]), 0)).toBeNull();
  });

  it('is null for an out-of-range seat', () => {
    expect(cuckooSwapTarget(seats([]), -1)).toBeNull();
    expect(cuckooSwapTarget(seats([]), 9)).toBeNull();
  });
});
