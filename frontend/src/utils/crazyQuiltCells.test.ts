import { describe, expect, it } from 'vitest';
import { isCrazyQuiltVertical } from './crazyQuiltCells';

describe('isCrazyQuiltVertical', () => {
  it('alternates orientation across rows and columns', () => {
    expect(isCrazyQuiltVertical(0)).toBe(true);
    expect(isCrazyQuiltVertical(1)).toBe(false);
    expect(isCrazyQuiltVertical(8)).toBe(false);
    expect(isCrazyQuiltVertical(9)).toBe(true);
    expect(isCrazyQuiltVertical(63)).toBe(true);
  });

  it('treats out-of-range cells as horizontal', () => {
    expect(isCrazyQuiltVertical(-1)).toBe(false);
    expect(isCrazyQuiltVertical(64)).toBe(false);
  });
});
