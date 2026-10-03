import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { canastaFamilyCardValue, canastaFamilySelectionPoints } from './canastaFamilyScore';

const c = (design: Card['design'], value: number): Card => ({ design, value });
describe('canasta-family scoring', () => {
  it('scores representative cards', () => {
    expect(canastaFamilyCardValue(c('JOKER', 0))).toBe(50);
    expect(canastaFamilyCardValue(c('SPADE', 2))).toBe(20);
    expect(canastaFamilyCardValue(c('HEART', 1))).toBe(20);
    expect(canastaFamilyCardValue(c('SPADE', 3))).toBe(5);
    expect(canastaFamilyCardValue(c('CLOVER', 3))).toBe(5);
    expect(canastaFamilyCardValue(c('HEART', 4))).toBe(5);
    expect(canastaFamilyCardValue(c('DIAMOND', 3))).toBe(5);
    expect(canastaFamilyCardValue(c('HEART', 8))).toBe(10);
    expect(canastaFamilyCardValue(c('DIAMOND', 13))).toBe(10);
    expect(canastaFamilyCardValue(c('DIAMOND', 5))).toBe(5);
  });
  it('sums selected cards', () => {
    expect(canastaFamilySelectionPoints([c('JOKER', 0), c('SPADE', 1), c('HEART', 8)])).toBe(80);
    expect(canastaFamilySelectionPoints([c('JOKER', 0), c('SPADE', 2), c('HEART', 4)])).toBe(75);
    expect(canastaFamilySelectionPoints([])).toBe(0);
  });
});
