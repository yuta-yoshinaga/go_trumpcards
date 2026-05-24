import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { utHoldemPreflopStrength } from './utHoldemPreflop';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('utHoldemPreflopStrength', () => {
  it.each([
    ['pair of 2s', [c('SPADE', 2), c('HEART', 2)], 'strong'],
    ['any ace', [c('SPADE', 1), c('HEART', 5)], 'strong'],
    ['suited king', [c('SPADE', 13), c('SPADE', 4)], 'strong'],
    ['K-Q offsuit', [c('SPADE', 13), c('HEART', 12)], 'strong'],
    ['Q-J offsuit', [c('SPADE', 12), c('HEART', 11)], 'strong'],
    ['K-7 offsuit', [c('SPADE', 13), c('HEART', 7)], 'moderate'],
    ['suited 8-9', [c('SPADE', 8), c('SPADE', 9)], 'moderate'],
    ['Q-4 suited', [c('SPADE', 12), c('SPADE', 4)], 'moderate'],
    ['7-2 offsuit', [c('SPADE', 7), c('HEART', 2)], 'weak'],
    ['empty hand', [], 'weak'],
  ])('%s -> %s', (_, hand, expected) => {
    expect(utHoldemPreflopStrength(hand as Card[])).toBe(expected);
  });
});
