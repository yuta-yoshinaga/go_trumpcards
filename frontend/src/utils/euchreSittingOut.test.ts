import { describe, expect, it } from 'vitest';
import type { EuchrePlayerData, EuchreResponse } from '../types/card';
import { euchreSittingOutIdx } from './euchreSittingOut';

function p(id: number, team: number): EuchrePlayerData {
  return { id, isHuman: id === 0, cardCount: 5, cards: [], team, trickCount: 0 };
}

function s(
  overrides: Partial<EuchreResponse> = {},
): Pick<EuchreResponse, 'players' | 'goingAlone' | 'goingAlonePlayerIdx'> {
  return {
    players: [p(0, 0), p(1, 1), p(2, 0), p(3, 1)],
    goingAlone: false,
    goingAlonePlayerIdx: -1,
    ...overrides,
  };
}

describe('euchreSittingOutIdx', () => {
  it('returns null when no one is going alone', () => {
    expect(euchreSittingOutIdx(s())).toBeNull();
  });

  it('returns the teammate of the going-alone player', () => {
    expect(euchreSittingOutIdx(s({ goingAlone: true, goingAlonePlayerIdx: 0 }))).toBe(2);
    expect(euchreSittingOutIdx(s({ goingAlone: true, goingAlonePlayerIdx: 1 }))).toBe(3);
  });

  it('returns null when goingAlonePlayerIdx is not in players', () => {
    expect(euchreSittingOutIdx(s({ goingAlone: true, goingAlonePlayerIdx: 99 }))).toBeNull();
  });

  it('returns null when goingAlone but team has no partner (edge case)', () => {
    expect(
      euchreSittingOutIdx({
        players: [p(0, 0), p(1, 1), p(2, 1), p(3, 1)],
        goingAlone: true,
        goingAlonePlayerIdx: 0,
      }),
    ).toBeNull();
  });
});
