import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { eightOffFoundationTarget } from './eightOffFoundationTarget';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('eightOffFoundationTarget', () => {
  it('sends an Ace to its empty suit pile', () => {
    expect(eightOffFoundationTarget(c('HEART', 1), [[], [], [], []])).toEqual({ zone: 'foundation', col: 2 });
  });

  it('sends the next rank onto a matching-suit pile', () => {
    const foundation = [[c('SPADE', 1)], [], [], []];
    expect(eightOffFoundationTarget(c('SPADE', 2), foundation)).toEqual({ zone: 'foundation', col: 0 });
  });

  it('returns null for a non-Ace on an empty pile', () => {
    expect(eightOffFoundationTarget(c('CLOVER', 5), [[], [], [], []])).toBeNull();
  });

  it('returns null when the rank is not exactly one higher', () => {
    const foundation = [[], [c('CLOVER', 1)], [], []];
    expect(eightOffFoundationTarget(c('CLOVER', 3), foundation)).toBeNull();
  });

  it('returns null when the suit does not match the target pile top', () => {
    // Diamond maps to pile index 3, which is empty and only accepts an Ace.
    const foundation = [[c('SPADE', 1)], [], [], []];
    expect(eightOffFoundationTarget(c('DIAMOND', 2), foundation)).toBeNull();
  });

  it('returns null for an unknown design', () => {
    expect(eightOffFoundationTarget(c('JOKER' as Card['design'], 1), [[], [], [], []])).toBeNull();
  });
});
