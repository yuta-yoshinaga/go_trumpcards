import { describe, expect, it } from 'vitest';
import type { HeartsPlayerData } from '../types/card';
import { shootTheMoonAlertIdx } from './heartsShootMoonAlert';

const player = (id: number, roundScore: number): HeartsPlayerData => ({
  id,
  isHuman: id === 0,
  cardCount: 0,
  cards: [],
  roundScore,
  cumulativeScore: 0,
  trickCount: 0,
});

describe('shootTheMoonAlertIdx', () => {
  it('returns null when no one has points yet', () => {
    expect(shootTheMoonAlertIdx([player(0, 0), player(1, 0), player(2, 0), player(3, 0)])).toBeNull();
  });

  it('returns null when the total is below the threshold', () => {
    expect(shootTheMoonAlertIdx([player(0, 0), player(1, 0), player(2, 0), player(3, 5)])).toBeNull();
  });

  it('returns the leader index when one player holds all 13+ points', () => {
    expect(shootTheMoonAlertIdx([player(0, 0), player(1, 0), player(2, 13), player(3, 0)])).toBe(2);
  });

  it('returns null when the points are split between two players', () => {
    expect(shootTheMoonAlertIdx([player(0, 0), player(1, 7), player(2, 7), player(3, 0)])).toBeNull();
  });

  it('returns the leader index in extreme scenarios (Q♠ + many hearts)', () => {
    expect(shootTheMoonAlertIdx([player(0, 0), player(1, 0), player(2, 20), player(3, 0)])).toBe(2);
  });
});
