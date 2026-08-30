import { describe, expect, it } from 'vitest';
import type { TarneebPlayerData } from '../types/card';
import { groupTarneebPlayersByTeam } from './tarneebTeams';

function player(id: number, team: number, trickCount: number, roundScore = 0): TarneebPlayerData {
  return {
    id,
    isHuman: id === 0,
    team,
    cardCount: 0,
    cards: [],
    bid: -1,
    roundScore,
    cumulativeScore: 0,
    trickCount,
  };
}

describe('groupTarneebPlayersByTeam', () => {
  it('groups players by team and sums each team’s round tricks', () => {
    const players = [player(0, 0, 3, 4), player(1, 1, 2, 2), player(2, 0, 1, 4), player(3, 1, 0, 2)];
    const result = groupTarneebPlayersByTeam(players, 2);

    expect(result).toHaveLength(2);
    expect(result[0].team).toBe(0);
    expect(result[0].members.map((p) => p.id)).toEqual([0, 2]);
    expect(result[0].roundTricks).toBe(4);
    expect(result[0].roundScore).toBe(4);
    expect(result[1].team).toBe(1);
    expect(result[1].members.map((p) => p.id)).toEqual([1, 3]);
    expect(result[1].roundTricks).toBe(2);
    expect(result[1].roundScore).toBe(2);
  });

  it('returns a single member’s roundScore without summing (does not double)', () => {
    // Both members in team 0 have roundScore = 8 (replicated delta).
    // Summing would yield 16, which is wrong. Correct value is 8.
    const players = [player(0, 0, 5, 8), player(1, 1, 4, 5), player(2, 0, 3, 8), player(3, 1, 1, 5)];
    const result = groupTarneebPlayersByTeam(players, 2);

    expect(result[0].roundScore).toBe(8);
    expect(result[0].roundScore).not.toBe(16);
    expect(result[0].roundTricks).toBe(8); // sum of 5 + 3 tricks

    expect(result[1].roundScore).toBe(5);
    expect(result[1].roundScore).not.toBe(10);
    expect(result[1].roundTricks).toBe(5); // sum of 4 + 1 tricks
  });

  it('returns negative roundScore correctly when a bid fails', () => {
    // Team 0 failed bid: both members have roundScore = -8.
    // Team 1 defenders: both members have roundScore = 5 (positive).
    const players = [player(0, 0, 4, -8), player(1, 1, 3, 5), player(2, 0, 2, -8), player(3, 1, 4, 5)];
    const result = groupTarneebPlayersByTeam(players, 2);

    expect(result[0].roundScore).toBe(-8);
    expect(result[0].roundTricks).toBe(6);
    expect(result[1].roundScore).toBe(5);
    expect(result[1].roundTricks).toBe(7);
  });

  it('returns an empty team with zero tricks and zero roundScore when no player belongs to it', () => {
    const result = groupTarneebPlayersByTeam([player(0, 0, 5, 5)], 2);
    expect(result[1].members).toEqual([]);
    expect(result[1].roundTricks).toBe(0);
    expect(result[1].roundScore).toBe(0);
  });
});
