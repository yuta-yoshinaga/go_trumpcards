import { describe, expect, it } from 'vitest';
import type { TarneebPlayerData } from '../types/card';
import { groupTarneebPlayersByTeam } from './tarneebTeams';

function player(id: number, team: number, trickCount: number): TarneebPlayerData {
  return {
    id,
    isHuman: id === 0,
    team,
    cardCount: 0,
    cards: [],
    bid: -1,
    roundScore: 0,
    cumulativeScore: 0,
    trickCount,
  };
}

describe('groupTarneebPlayersByTeam', () => {
  it('groups players by team and sums each team’s round tricks', () => {
    const players = [player(0, 0, 3), player(1, 1, 2), player(2, 0, 1), player(3, 1, 0)];
    const result = groupTarneebPlayersByTeam(players, 2);

    expect(result).toHaveLength(2);
    expect(result[0].team).toBe(0);
    expect(result[0].members.map((p) => p.id)).toEqual([0, 2]);
    expect(result[0].roundTricks).toBe(4);
    expect(result[1].team).toBe(1);
    expect(result[1].members.map((p) => p.id)).toEqual([1, 3]);
    expect(result[1].roundTricks).toBe(2);
  });

  it('returns an empty team with zero tricks when no player belongs to it', () => {
    const result = groupTarneebPlayersByTeam([player(0, 0, 5)], 2);
    expect(result[1].members).toEqual([]);
    expect(result[1].roundTricks).toBe(0);
  });
});
