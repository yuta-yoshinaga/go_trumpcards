import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { SULTAN_FOUNDATION_FULL, sultanFoundationInfo } from './sultanFoundation';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('sultanFoundationInfo', () => {
  it('reads the suit from the King base of a fresh pile', () => {
    const info = sultanFoundationInfo([card('SPADE', 13)]);
    expect(info.suit).toBe('♠');
    expect(info.design).toBe('SPADE');
    expect(info.count).toBe(1);
    expect(info.complete).toBe(false);
  });

  it('keeps the King base suit as the pile grows', () => {
    const info = sultanFoundationInfo([card('HEART', 13), card('HEART', 1), card('HEART', 2)]);
    expect(info.suit).toBe('♥');
    expect(info.design).toBe('HEART');
    expect(info.count).toBe(3);
    expect(info.complete).toBe(false);
  });

  it('marks a pile of 13 cards as complete', () => {
    const pile = Array.from({ length: SULTAN_FOUNDATION_FULL }, () => card('DIAMOND', 13));
    const info = sultanFoundationInfo(pile);
    expect(info.complete).toBe(true);
    expect(info.suit).toBe('♦');
  });

  it('reports no suit for an empty pile', () => {
    const info = sultanFoundationInfo([]);
    expect(info.suit).toBe('');
    expect(info.design).toBeNull();
    expect(info.count).toBe(0);
    expect(info.complete).toBe(false);
  });
});
