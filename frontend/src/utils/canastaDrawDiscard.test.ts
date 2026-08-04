import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { canastaDrawDiscardProblem, canastaIsWild } from './canastaDrawDiscard';

const c = (design: CardDesign, value: number): Card => ({ design, value });
const top = c('SPADE', 9);

describe('canastaIsWild', () => {
  it('treats jokers and twos as wild', () => {
    expect(canastaIsWild(c('JOKER', 0))).toBe(true);
    expect(canastaIsWild(c('HEART', 2))).toBe(true);
  });

  it('treats every other card as natural', () => {
    expect(canastaIsWild(c('HEART', 3))).toBe(false);
    expect(canastaIsWild(c('SPADE', 9))).toBe(false);
  });
});

describe('canastaDrawDiscardProblem', () => {
  it('asks for two cards when none are chosen', () => {
    expect(canastaDrawDiscardProblem([], top)).toBe('selectTwo');
  });

  it('asks for one more when only one is chosen', () => {
    expect(canastaDrawDiscardProblem([c('HEART', 9)], top)).toBe('selectOneMore');
  });

  it('rejects more than two', () => {
    expect(canastaDrawDiscardProblem([c('HEART', 9), c('CLOVER', 9), c('DIAMOND', 9)], top)).toBe('tooMany');
  });

  it('accepts a natural pair of the discard top rank', () => {
    expect(canastaDrawDiscardProblem([c('HEART', 9), c('CLOVER', 9)], top)).toBeNull();
  });

  it('rejects a wild in the pair, which the old count-only check let through', () => {
    expect(canastaDrawDiscardProblem([c('HEART', 9), c('JOKER', 0)], top)).toBe('wildInPair');
    expect(canastaDrawDiscardProblem([c('HEART', 2), c('CLOVER', 9)], top)).toBe('wildInPair');
  });

  it('rejects a pair that does not match the discard top', () => {
    expect(canastaDrawDiscardProblem([c('HEART', 8), c('CLOVER', 8)], top)).toBe('rankMismatch');
    // One matching card is not enough — both must.
    expect(canastaDrawDiscardProblem([c('HEART', 9), c('CLOVER', 8)], top)).toBe('rankMismatch');
  });

  it('rejects taking from an empty pile', () => {
    expect(canastaDrawDiscardProblem([c('HEART', 9), c('CLOVER', 9)], null)).toBe('pileEmpty');
  });
});
