import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { classifyYanivDiscard, isValidYanivDiscard } from './yanivCombos';

const card = (design: Card['design'], value: number): Card => ({ design, value });
const joker = (value: number): Card => ({ design: 'JOKER', value });

describe('classifyYanivDiscard', () => {
  it('classifies a single card as a valid single', () => {
    const res = classifyYanivDiscard([card('HEART', 7)]);
    expect(res.kind).toBe('single');
    expect(res.reasonKey).toBeUndefined();
  });

  it('classifies a single joker as a valid single', () => {
    expect(classifyYanivDiscard([joker(1)]).kind).toBe('single');
  });

  it('classifies a same-rank pair as a valid set', () => {
    const res = classifyYanivDiscard([card('HEART', 8), card('SPADE', 8)]);
    expect(res.kind).toBe('set');
    expect(res.reasonKey).toBeUndefined();
  });

  it('classifies a same-rank triple as a valid set', () => {
    expect(classifyYanivDiscard([card('HEART', 8), card('SPADE', 8), card('CLOVER', 8)]).kind).toBe('set');
  });

  it('classifies a same-suit 3-run as a valid run regardless of order', () => {
    const res = classifyYanivDiscard([card('DIAMOND', 6), card('DIAMOND', 4), card('DIAMOND', 5)]);
    expect(res.kind).toBe('run');
    expect(res.reasonKey).toBeUndefined();
  });

  it('rejects an empty selection', () => {
    const res = classifyYanivDiscard([]);
    expect(res.kind).toBe('invalid');
    expect(res.reasonKey).toBe('discardWarn.empty');
  });

  it('rejects a mixed 2-card non-pair', () => {
    const res = classifyYanivDiscard([card('HEART', 8), card('SPADE', 9)]);
    expect(res.kind).toBe('invalid');
    expect(res.reasonKey).toBe('discardWarn.pair');
  });

  it('rejects a two-card selection containing a joker', () => {
    const res = classifyYanivDiscard([card('HEART', 8), joker(1)]);
    expect(res.kind).toBe('invalid');
    expect(res.reasonKey).toBe('discardWarn.joker');
  });

  it('rejects a pair of jokers (jokers cannot form a value set)', () => {
    const res = classifyYanivDiscard([joker(1), joker(2)]);
    expect(res.kind).toBe('invalid');
    expect(res.reasonKey).toBe('discardWarn.joker');
  });

  it('rejects a same-suit run that is not consecutive', () => {
    const res = classifyYanivDiscard([card('DIAMOND', 4), card('DIAMOND', 5), card('DIAMOND', 7)]);
    expect(res.kind).toBe('invalid');
    expect(res.reasonKey).toBe('discardWarn.general');
  });

  it('rejects a consecutive run of mixed suits', () => {
    const res = classifyYanivDiscard([card('DIAMOND', 4), card('HEART', 5), card('DIAMOND', 6)]);
    expect(res.kind).toBe('invalid');
    expect(res.reasonKey).toBe('discardWarn.general');
  });

  it('rejects a joker in the middle of an otherwise valid run', () => {
    const res = classifyYanivDiscard([card('DIAMOND', 4), joker(1), card('DIAMOND', 6)]);
    expect(res.kind).toBe('invalid');
    expect(res.reasonKey).toBe('discardWarn.joker');
  });
});

describe('isValidYanivDiscard', () => {
  it('is true for a valid single, set and run', () => {
    expect(isValidYanivDiscard([card('HEART', 7)])).toBe(true);
    expect(isValidYanivDiscard([card('HEART', 8), card('SPADE', 8)])).toBe(true);
    expect(isValidYanivDiscard([card('DIAMOND', 4), card('DIAMOND', 5), card('DIAMOND', 6)])).toBe(true);
  });

  it('is false for an invalid selection and an empty selection', () => {
    expect(isValidYanivDiscard([card('HEART', 8), card('SPADE', 9)])).toBe(false);
    expect(isValidYanivDiscard([])).toBe(false);
  });
});
