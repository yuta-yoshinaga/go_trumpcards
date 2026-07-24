import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { computeSpoonsRankGroups } from './spoonsRankGroups';

const card = (value: number, design: Card['design'] = 'SPADE'): Card => ({ design, value });

describe('computeSpoonsRankGroups', () => {
  it('returns null colorIndex for a hand of all singletons', () => {
    const groups = computeSpoonsRankGroups([card(2), card(5), card(9), card(13)]);
    expect(groups.map((g) => g.colorIndex)).toEqual([null, null, null, null]);
    expect(groups.every((g) => g.count === 1)).toBe(true);
  });

  it('assigns a shared color to a pair and none to singletons', () => {
    // Hand: 7♠ 7♥ 3♣ K♦ — the two 7s form a group, others are singletons.
    const groups = computeSpoonsRankGroups([
      card(7, 'SPADE'),
      card(7, 'HEART'),
      card(3, 'CLOVER'),
      card(13, 'DIAMOND'),
    ]);
    expect(groups[0].colorIndex).toBe(0);
    expect(groups[1].colorIndex).toBe(0);
    expect(groups[0].colorIndex).toBe(groups[1].colorIndex);
    expect(groups[2].colorIndex).toBeNull();
    expect(groups[3].colorIndex).toBeNull();
    expect(groups.map((g) => g.count)).toEqual([2, 2, 1, 1]);
  });

  it('gives two pairs distinct color indices, deterministic by ascending rank', () => {
    // Hand order puts the higher pair first; lower rank must still get index 0.
    const groups = computeSpoonsRankGroups([card(10), card(4), card(10), card(4)]);
    // Rank 4 (lower) -> index 0, rank 10 (higher) -> index 1.
    expect(groups[0].colorIndex).toBe(1); // 10
    expect(groups[1].colorIndex).toBe(0); // 4
    expect(groups[2].colorIndex).toBe(1); // 10
    expect(groups[3].colorIndex).toBe(0); // 4
    expect(groups.map((g) => g.count)).toEqual([2, 2, 2, 2]);
  });

  it('flags a three-of-a-kind (reach) with count 3 sharing one color', () => {
    const groups = computeSpoonsRankGroups([card(8), card(8), card(8), card(2)]);
    expect(groups.slice(0, 3).every((g) => g.colorIndex === 0)).toBe(true);
    expect(groups.slice(0, 3).every((g) => g.count === 3)).toBe(true);
    expect(groups[3].colorIndex).toBeNull();
    expect(groups[3].count).toBe(1);
  });

  it('handles an empty hand', () => {
    expect(computeSpoonsRankGroups([])).toEqual([]);
  });
});
