import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { sambaCardValue, sambaMinMeld, sambaSelectionPoints } from './sambaScore';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('sambaMinMeld', () => {
  it('returns 15 for a negative team score', () => {
    expect(sambaMinMeld(-100)).toBe(15);
  });
  it('returns 50 below 1500', () => {
    expect(sambaMinMeld(0)).toBe(50);
    expect(sambaMinMeld(1499)).toBe(50);
  });
  it('returns 90 below 3000', () => {
    expect(sambaMinMeld(1500)).toBe(90);
    expect(sambaMinMeld(2999)).toBe(90);
  });
  it('returns 120 at 3000 and above', () => {
    expect(sambaMinMeld(3000)).toBe(120);
  });
});

describe('sambaCardValue', () => {
  it('scores jokers at 50', () => {
    expect(sambaCardValue(card('JOKER', 0))).toBe(50);
  });
  it('scores 2s and aces at 20', () => {
    expect(sambaCardValue(card('SPADE', 2))).toBe(20);
    expect(sambaCardValue(card('HEART', 1))).toBe(20);
  });
  it('scores black 3s at 5', () => {
    expect(sambaCardValue(card('SPADE', 3))).toBe(5);
    expect(sambaCardValue(card('CLOVER', 3))).toBe(5);
  });
  it('scores 8+ at 10', () => {
    expect(sambaCardValue(card('HEART', 8))).toBe(10);
    expect(sambaCardValue(card('DIAMOND', 13))).toBe(10);
  });
  it('scores low ranks at 5', () => {
    expect(sambaCardValue(card('HEART', 4))).toBe(5);
    expect(sambaCardValue(card('DIAMOND', 3))).toBe(5);
  });
});

describe('sambaSelectionPoints', () => {
  it('sums the card values', () => {
    expect(sambaSelectionPoints([card('JOKER', 0), card('SPADE', 2), card('HEART', 4)])).toBe(75);
  });
  it('returns 0 for an empty selection', () => {
    expect(sambaSelectionPoints([])).toBe(0);
  });
});
