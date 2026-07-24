import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { basraFindCaptures, isBasraJack, resolveBasraAction } from './basraCaptures';

/** Builds a card with the given rank value (suit is irrelevant to capture logic). */
const c = (value: number, design: Card['design'] = 'SPADE'): Card => ({ design, value });

describe('basraFindCaptures', () => {
  it('captures a same-rank table card with a number card', () => {
    // A 5 captures the table 5 by matching rank.
    expect(basraFindCaptures(c(5), [c(5), c(9)])).toEqual([0]);
  });

  it('captures a two-card subset summing to the played value', () => {
    // A 5 captures 2 + 3 (indices 0 and 1); the 9 is untouched.
    expect(basraFindCaptures(c(5), [c(2), c(3), c(9)])).toEqual([0, 1]);
  });

  it('greedily captures both a same-rank single and a summing subset', () => {
    // A 5 takes the table 5 (index 0) AND the 2 + 3 pair (indices 1, 2).
    expect(basraFindCaptures(c(5), [c(5), c(2), c(3)])).toEqual([0, 1, 2]);
  });

  it('captures nothing when no rank matches and no subset sums (trail)', () => {
    expect(basraFindCaptures(c(3), [c(5), c(9)])).toEqual([]);
  });

  it('sweeps every non-Jack table card when a Jack is played', () => {
    // The Jack (value 11) sweeps indices 0 and 2 but leaves the table Jack (index 1).
    expect(basraFindCaptures(c(11), [c(5), c(11), c(9)])).toEqual([0, 2]);
  });

  it('captures faces (Q/K) by matching rank only, never by sum', () => {
    // A Queen (12) captures the table Queen only; 5 + 7 never sums into a face.
    expect(basraFindCaptures(c(12), [c(12), c(5), c(7)])).toEqual([0]);
  });

  it('captures only other Aces with an Ace (value 1)', () => {
    expect(basraFindCaptures(c(1), [c(1), c(1), c(9)])).toEqual([0, 1]);
  });
});

describe('isBasraJack', () => {
  it('is true only for value-11 cards', () => {
    expect(isBasraJack(c(11))).toBe(true);
    expect(isBasraJack(c(12))).toBe(false);
    expect(isBasraJack(null)).toBe(false);
  });
});

describe('resolveBasraAction', () => {
  it('is idle when no card is selected', () => {
    expect(resolveBasraAction(null, [], [])).toEqual({ kind: 'idle', count: 0 });
  });

  it('sweeps for a Jack, counting every capturable card regardless of selection', () => {
    expect(resolveBasraAction(c(11), [0, 1], [])).toEqual({ kind: 'sweep', count: 2 });
  });

  it('captures the selected table cards for a non-Jack selection', () => {
    expect(resolveBasraAction(c(5), [0, 1], [0])).toEqual({ kind: 'capture', count: 1 });
  });

  it('is capturable (awaiting selection) when captures exist but none are picked', () => {
    expect(resolveBasraAction(c(5), [0], [])).toEqual({ kind: 'capturable', count: 1 });
  });

  it('trails when a non-Jack card can capture nothing', () => {
    expect(resolveBasraAction(c(3), [], [])).toEqual({ kind: 'trail', count: 0 });
  });
});
