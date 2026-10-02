import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { rankLabel, redDogRank, reddogWinningCardOdds, reddogWinningRanks } from './reddogWinningRanks';

const c = (value: number): Card => ({ design: 'SPADE', value });

describe('reddogWinningRanks', () => {
  it('returns the ranks strictly between the two initial cards', () => {
    expect(reddogWinningRanks([c(4), c(9)])).toEqual([5, 6, 7, 8]);
  });

  it('treats ace as the high rank (14)', () => {
    expect(reddogWinningRanks([c(11), c(1)])).toEqual([12, 13]);
  });

  it('returns [] on pair', () => {
    expect(reddogWinningRanks([c(7), c(7)])).toEqual([]);
  });

  it('returns [] on consecutive ranks', () => {
    expect(reddogWinningRanks([c(7), c(8)])).toEqual([]);
  });

  it('returns [] when initial does not have exactly 2 cards', () => {
    expect(reddogWinningRanks([])).toEqual([]);
    expect(reddogWinningRanks([c(5)])).toEqual([]);
  });

  it('is unaffected by input order', () => {
    expect(reddogWinningRanks([c(10), c(3)])).toEqual([4, 5, 6, 7, 8, 9]);
  });
});

describe('reddogWinningCardOdds', () => {
  it('counts remaining cards for the winning ranks from the 50-card undealt deck', () => {
    expect(reddogWinningCardOdds([c(5), c(10)])).toEqual({ winningCards: 16, remainingCards: 50, probability: 0.32 });
  });

  it('counts the pair matching rank as the winning condition', () => {
    expect(reddogWinningCardOdds([c(7), c(7)])).toEqual({ winningCards: 2, remainingCards: 50, probability: 0.04 });
  });

  it('returns null when there is no winning rank', () => {
    expect(reddogWinningCardOdds([c(7), c(8)])).toBeNull();
    expect(reddogWinningCardOdds([])).toBeNull();
  });
});

describe('redDogRank', () => {
  it('promotes ace to 14', () => {
    expect(redDogRank(c(1))).toBe(14);
  });
  it('returns the value as-is for 2..K', () => {
    expect(redDogRank(c(2))).toBe(2);
    expect(redDogRank(c(13))).toBe(13);
  });
});

describe('rankLabel', () => {
  it('formats face cards', () => {
    expect(rankLabel(14)).toBe('A');
    expect(rankLabel(13)).toBe('K');
    expect(rankLabel(12)).toBe('Q');
    expect(rankLabel(11)).toBe('J');
  });
  it('formats number cards', () => {
    expect(rankLabel(7)).toBe('7');
  });
});
