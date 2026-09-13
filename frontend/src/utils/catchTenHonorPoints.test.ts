import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { catchTenHonorPoints } from './catchTenHonorPoints';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('catchTenHonorPoints', () => {
  it.each([
    [11, 11],
    [10, 10],
    [1, 4],
    [13, 3],
    [12, 2],
    [9, 0],
  ])('scores trump rank %i as %i points', (value, points) => {
    expect(catchTenHonorPoints(card('SPADE', value), 1)).toBe(points);
  });

  it('scores every non-trump rank as zero', () => {
    expect(catchTenHonorPoints(card('HEART', 11), 1)).toBe(0);
    expect(catchTenHonorPoints(card('HEART', 10), 1)).toBe(0);
    expect(catchTenHonorPoints(card('HEART', 1), 1)).toBe(0);
    expect(catchTenHonorPoints(card('HEART', 13), 1)).toBe(0);
    expect(catchTenHonorPoints(card('HEART', 12), 1)).toBe(0);
  });
});
