import { describe, expect, it } from 'vitest';
import { splitBalanced } from './splitBalanced';

describe('splitBalanced', () => {
  it('splits contiguously and puts remainder items in earlier rows', () => {
    expect(splitBalanced([0, 1, 2, 3, 4], 2)).toEqual([
      { items: [0, 1, 2], start: 0 },
      { items: [3, 4], start: 3 },
    ]);
  });

  it('handles empty items and limits rows to the item count', () => {
    expect(splitBalanced([], 3)).toEqual([{ items: [], start: 0 }]);
    expect(splitBalanced(['a'], 4)).toEqual([{ items: ['a'], start: 0 }]);
  });
});
