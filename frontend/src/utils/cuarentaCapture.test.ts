import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { cuarentaCaptureIndices } from './cuarentaCapture';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('cuarentaCaptureIndices', () => {
  it('returns an empty set when no hand card is selected', () => {
    const table = [card('SPADE', 5), card('HEART', 5)];
    expect(cuarentaCaptureIndices(null, table).size).toBe(0);
  });

  it('captures every table card of the same rank regardless of suit', () => {
    const table = [card('SPADE', 7), card('HEART', 3), card('DIAMOND', 7)];
    const result = cuarentaCaptureIndices(card('CLOVER', 7), table);
    expect([...result].sort()).toEqual([0, 2]);
  });

  it('returns an empty set when no table card matches the rank', () => {
    const table = [card('SPADE', 5), card('HEART', 3)];
    expect(cuarentaCaptureIndices(card('CLOVER', 12), table).size).toBe(0);
  });

  it('captures a single matching table card', () => {
    const table = [card('SPADE', 1), card('HEART', 11), card('DIAMOND', 2)];
    const result = cuarentaCaptureIndices(card('CLOVER', 11), table);
    expect([...result]).toEqual([1]);
  });

  it('handles an empty table', () => {
    expect(cuarentaCaptureIndices(card('CLOVER', 7), []).size).toBe(0);
  });
});
