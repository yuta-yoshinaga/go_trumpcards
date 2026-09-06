import { describe, expect, it } from 'vitest';
import type { OmiPlayerData, OmiResponse } from '../types/card';
import { omiSittingOutIdx } from './omiSittingOut';

function p(id: number, team: number): OmiPlayerData {
  return { id, isHuman: id === 0, cardCount: 5, cards: [], team, trickCount: 0 };
}

function s(overrides: Partial<OmiResponse> = {}): Pick<OmiResponse, 'players' | 'goingAlone' | 'goingAlonePlayerIdx'> {
  return {
    players: [p(0, 0), p(1, 1), p(2, 0), p(3, 1)],
    goingAlone: false,
    goingAlonePlayerIdx: -1,
    ...overrides,
  };
}

describe('omiSittingOutIdx', () => {
  it('returns null when no one is going alone', () => {
    expect(omiSittingOutIdx(s())).toBeNull();
  });

  it('returns the teammate of the going-alone player', () => {
    expect(omiSittingOutIdx(s({ goingAlone: true, goingAlonePlayerIdx: 0 }))).toBe(2);
    expect(omiSittingOutIdx(s({ goingAlone: true, goingAlonePlayerIdx: 1 }))).toBe(3);
  });

  it('returns null when goingAlonePlayerIdx is not in players', () => {
    expect(omiSittingOutIdx(s({ goingAlone: true, goingAlonePlayerIdx: 99 }))).toBeNull();
  });

  it('returns null when goingAlone but team has no partner (edge case)', () => {
    expect(
      omiSittingOutIdx({
        players: [p(0, 0), p(1, 1), p(2, 1), p(3, 1)],
        goingAlone: true,
        goingAlonePlayerIdx: 0,
      }),
    ).toBeNull();
  });
});
