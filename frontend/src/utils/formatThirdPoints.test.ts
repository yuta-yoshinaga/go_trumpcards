import { describe, expect, it } from 'vitest';
import { formatThirdPoints } from './formatThirdPoints';

describe('formatThirdPoints', () => {
  it.each([
    [4, '1+1/3'],
    [6, '2'],
    [2, '2/3'],
  ])('formats %i thirds as %s', (thirds, expected) => {
    expect(formatThirdPoints(thirds)).toBe(expected);
  });
});
