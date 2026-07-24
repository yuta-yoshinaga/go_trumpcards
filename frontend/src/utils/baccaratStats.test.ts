import { describe, expect, it } from 'vitest';
import { computeBaccaratShoeStats, ROAD_BANKER, ROAD_PLAYER } from './baccaratStats';

const TIE = 2;

describe('computeBaccaratShoeStats', () => {
  it('returns zeroed stats for empty history', () => {
    const s = computeBaccaratShoeStats([]);
    expect(s).toEqual({
      playerCount: 0,
      bankerCount: 0,
      tieCount: 0,
      total: 0,
      playerPct: 0,
      bankerPct: 0,
      tiePct: 0,
      streakSide: null,
      streakCount: 0,
    });
  });

  it('counts Player/Banker/Tie occurrences', () => {
    const s = computeBaccaratShoeStats([ROAD_PLAYER, ROAD_BANKER, ROAD_BANKER, TIE]);
    expect(s.playerCount).toBe(1);
    expect(s.bankerCount).toBe(2);
    expect(s.tieCount).toBe(1);
    expect(s.total).toBe(4);
  });

  it('computes integer appearance rates against the total', () => {
    const s = computeBaccaratShoeStats([ROAD_PLAYER, ROAD_BANKER, ROAD_BANKER, TIE]);
    expect(s.playerPct).toBe(25);
    expect(s.bankerPct).toBe(50);
    expect(s.tiePct).toBe(25);
  });

  it('rounds rates to the nearest integer', () => {
    // 1 player, 2 banker of 3 -> 33% / 67% / 0%
    const s = computeBaccaratShoeStats([ROAD_PLAYER, ROAD_BANKER, ROAD_BANKER]);
    expect(s.playerPct).toBe(33);
    expect(s.bankerPct).toBe(67);
    expect(s.tiePct).toBe(0);
  });

  it('reports the current banker streak', () => {
    const s = computeBaccaratShoeStats([ROAD_PLAYER, ROAD_BANKER, ROAD_BANKER, ROAD_BANKER]);
    expect(s.streakSide).toBe(ROAD_BANKER);
    expect(s.streakCount).toBe(3);
  });

  it('reports the current player streak', () => {
    const s = computeBaccaratShoeStats([ROAD_BANKER, ROAD_PLAYER, ROAD_PLAYER]);
    expect(s.streakSide).toBe(ROAD_PLAYER);
    expect(s.streakCount).toBe(2);
  });

  it('does not break the streak on a trailing tie', () => {
    const s = computeBaccaratShoeStats([ROAD_PLAYER, ROAD_BANKER, ROAD_BANKER, TIE]);
    expect(s.streakSide).toBe(ROAD_BANKER);
    expect(s.streakCount).toBe(2);
  });

  it('ignores an interior tie within a streak', () => {
    const s = computeBaccaratShoeStats([ROAD_BANKER, TIE, ROAD_BANKER]);
    expect(s.streakSide).toBe(ROAD_BANKER);
    expect(s.streakCount).toBe(2);
  });

  it('resets the streak when the trailing side changes', () => {
    const s = computeBaccaratShoeStats([ROAD_BANKER, ROAD_BANKER, ROAD_PLAYER]);
    expect(s.streakSide).toBe(ROAD_PLAYER);
    expect(s.streakCount).toBe(1);
  });

  it('yields a null streak side for a tie-only history', () => {
    const s = computeBaccaratShoeStats([TIE, TIE]);
    expect(s.streakSide).toBeNull();
    expect(s.streakCount).toBe(0);
    expect(s.tieCount).toBe(2);
  });
});
