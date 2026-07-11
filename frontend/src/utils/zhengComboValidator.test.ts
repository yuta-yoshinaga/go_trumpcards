import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { classifyZhengCombo, isValidZhengCombo, zhengRankStrength } from './zhengComboValidator';

const card = (design: Card['design'], value: number): Card => ({ design, value });
const smallJoker = card('JOKER', 1);
const bigJoker = card('JOKER', 2);

describe('zhengRankStrength', () => {
  it('ranks 3 lowest and the big joker highest, ignoring suits', () => {
    expect(zhengRankStrength(card('SPADE', 3))).toBe(0);
    expect(zhengRankStrength(card('HEART', 3))).toBe(0);
    expect(zhengRankStrength(card('CLOVER', 13))).toBe(10);
    expect(zhengRankStrength(card('DIAMOND', 1))).toBe(11);
    expect(zhengRankStrength(card('SPADE', 2))).toBe(12);
    expect(zhengRankStrength(smallJoker)).toBe(13);
    expect(zhengRankStrength(bigJoker)).toBe(14);
  });
});

describe('classifyZhengCombo', () => {
  it('classifies singles, pairs, and triples', () => {
    expect(classifyZhengCombo([card('SPADE', 5)])).toBe('single');
    expect(classifyZhengCombo([card('SPADE', 5), card('HEART', 5)])).toBe('pair');
    expect(classifyZhengCombo([card('SPADE', 5), card('HEART', 5), card('CLOVER', 5)])).toBe('triple');
  });

  it('rejects mixed-rank pairs and joker pairs', () => {
    expect(classifyZhengCombo([card('SPADE', 5), card('HEART', 6)])).toBe('invalid');
    expect(classifyZhengCombo([smallJoker, card('SPADE', 5)])).toBe('invalid');
  });

  it('classifies two jokers as a joker bomb, not a pair', () => {
    expect(classifyZhengCombo([smallJoker, bigJoker])).toBe('jokerBomb');
    expect(classifyZhengCombo([bigJoker, smallJoker])).toBe('jokerBomb');
  });

  it('classifies four of a kind as a bomb', () => {
    expect(classifyZhengCombo([card('SPADE', 9), card('HEART', 9), card('CLOVER', 9), card('DIAMOND', 9)])).toBe(
      'bomb',
    );
  });

  it('classifies straights of 3+ with mixed suits', () => {
    expect(classifyZhengCombo([card('SPADE', 3), card('HEART', 4), card('SPADE', 5)])).toBe('straight');
    expect(classifyZhengCombo([card('SPADE', 11), card('HEART', 12), card('SPADE', 13), card('CLOVER', 1)])).toBe(
      'straight',
    );
    expect(
      classifyZhengCombo([card('SPADE', 4), card('HEART', 5), card('SPADE', 6), card('CLOVER', 7), card('HEART', 8)]),
    ).toBe('straight');
  });

  it('rejects straights containing 2 or a joker', () => {
    expect(classifyZhengCombo([card('SPADE', 13), card('HEART', 1), card('SPADE', 2)])).toBe('invalid');
    expect(classifyZhengCombo([card('SPADE', 3), card('HEART', 4), smallJoker])).toBe('invalid');
  });

  it('classifies three consecutive pairs as a pair run', () => {
    const run = [
      card('SPADE', 4),
      card('HEART', 4),
      card('SPADE', 5),
      card('CLOVER', 5),
      card('HEART', 6),
      card('DIAMOND', 6),
    ];
    expect(classifyZhengCombo(run)).toBe('pairRun');
  });

  it('rejects non-consecutive or 2-containing pair runs', () => {
    const gap = [
      card('SPADE', 4),
      card('HEART', 4),
      card('SPADE', 6),
      card('CLOVER', 6),
      card('HEART', 7),
      card('DIAMOND', 7),
    ];
    expect(classifyZhengCombo(gap)).toBe('invalid');
    const withTwo = [
      card('SPADE', 1),
      card('HEART', 1),
      card('SPADE', 2),
      card('CLOVER', 2),
      card('HEART', 3),
      card('DIAMOND', 3),
    ];
    expect(classifyZhengCombo(withTwo)).toBe('invalid');
  });
});

describe('isValidZhengCombo', () => {
  it('rejects an empty selection', () => {
    expect(isValidZhengCombo([])).toBe(false);
  });

  it('accepts a legal combo and rejects garbage', () => {
    expect(isValidZhengCombo([card('SPADE', 5), card('HEART', 5)])).toBe(true);
    expect(isValidZhengCombo([card('SPADE', 5), card('HEART', 6), card('CLOVER', 8)])).toBe(false);
  });
});
