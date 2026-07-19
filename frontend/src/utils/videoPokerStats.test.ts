import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  emptyVideoPokerStats,
  readVideoPokerStats,
  recordVideoPokerResult,
  videoPokerNet,
  videoPokerStatsKey,
  videoPokerWinRate,
  writeVideoPokerStats,
} from './videoPokerStats';

beforeEach(() => {
  localStorage.clear();
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('emptyVideoPokerStats', () => {
  it('returns a fresh zeroed object with its own handCounts map', () => {
    const a = emptyVideoPokerStats();
    const b = emptyVideoPokerStats();
    expect(a).toEqual({ hands: 0, wins: 0, totalBet: 0, totalPayout: 0, handCounts: {} });
    expect(a.handCounts).not.toBe(b.handCounts);
  });
});

describe('recordVideoPokerResult', () => {
  it('counts a win, its bet, its payout, and its hand key', () => {
    const next = recordVideoPokerResult(emptyVideoPokerStats(), { bet: 5, payout: 25, rowKey: 'fourOfAKind' });
    expect(next).toEqual({
      hands: 1,
      wins: 1,
      totalBet: 5,
      totalPayout: 25,
      handCounts: { fourOfAKind: 1 },
    });
  });

  it('counts a loss without incrementing wins or handCounts', () => {
    const next = recordVideoPokerResult(emptyVideoPokerStats(), { bet: 3, payout: 0, rowKey: null });
    expect(next.hands).toBe(1);
    expect(next.wins).toBe(0);
    expect(next.totalBet).toBe(3);
    expect(next.totalPayout).toBe(0);
    expect(next.handCounts).toEqual({});
  });

  it('does not mutate the input stats (pure reducer)', () => {
    const prev = emptyVideoPokerStats();
    const next = recordVideoPokerResult(prev, { bet: 1, payout: 2, rowKey: 'twoPair' });
    expect(prev).toEqual(emptyVideoPokerStats());
    expect(next).not.toBe(prev);
    expect(next.handCounts).not.toBe(prev.handCounts);
  });

  it('accumulates repeated wins of the same hand', () => {
    let stats = emptyVideoPokerStats();
    stats = recordVideoPokerResult(stats, { bet: 1, payout: 5, rowKey: 'jacksOrBetter' });
    stats = recordVideoPokerResult(stats, { bet: 2, payout: 10, rowKey: 'jacksOrBetter' });
    expect(stats.hands).toBe(2);
    expect(stats.wins).toBe(2);
    expect(stats.handCounts).toEqual({ jacksOrBetter: 2 });
  });

  it('treats a payout with a null rowKey as a win but records no hand key', () => {
    const next = recordVideoPokerResult(emptyVideoPokerStats(), { bet: 1, payout: 4, rowKey: null });
    expect(next.wins).toBe(1);
    expect(next.handCounts).toEqual({});
  });
});

describe('videoPokerNet / videoPokerWinRate', () => {
  it('computes net as payouts minus wagers (can be negative)', () => {
    const stats = { hands: 3, wins: 1, totalBet: 9, totalPayout: 5, handCounts: {} };
    expect(videoPokerNet(stats)).toBe(-4);
  });

  it('computes win rate as a 0..1 fraction, 0 when no hands', () => {
    expect(videoPokerWinRate(emptyVideoPokerStats())).toBe(0);
    expect(videoPokerWinRate({ hands: 4, wins: 1, totalBet: 4, totalPayout: 0, handCounts: {} })).toBe(0.25);
  });
});

describe('localStorage persistence', () => {
  it('round-trips stats through write then read, keyed per variant', () => {
    const stats = recordVideoPokerResult(emptyVideoPokerStats(), { bet: 5, payout: 4000, rowKey: 'royalFlush' });
    writeVideoPokerStats('videopoker', stats);
    expect(localStorage.getItem(videoPokerStatsKey('videopoker'))).not.toBeNull();
    expect(readVideoPokerStats('videopoker')).toEqual(stats);
    // A different variant has its own independent bucket.
    expect(readVideoPokerStats('deuceswild')).toEqual(emptyVideoPokerStats());
  });

  it('returns empty stats when nothing is stored', () => {
    expect(readVideoPokerStats('videopoker')).toEqual(emptyVideoPokerStats());
  });

  it('returns empty stats on malformed JSON', () => {
    localStorage.setItem(videoPokerStatsKey('videopoker'), '{not json');
    expect(readVideoPokerStats('videopoker')).toEqual(emptyVideoPokerStats());
  });

  it('normalizes partial/invalid stored objects, dropping non-positive counts', () => {
    localStorage.setItem(
      videoPokerStatsKey('videopoker'),
      JSON.stringify({ hands: 5, wins: 'x', totalPayout: 10, handCounts: { flush: 2, bogus: 0, bad: 'no' } }),
    );
    expect(readVideoPokerStats('videopoker')).toEqual({
      hands: 5,
      wins: 0,
      totalBet: 0,
      totalPayout: 10,
      handCounts: { flush: 2 },
    });
  });

  it('returns empty stats when the stored value is not an object', () => {
    localStorage.setItem(videoPokerStatsKey('videopoker'), '42');
    expect(readVideoPokerStats('videopoker')).toEqual(emptyVideoPokerStats());
  });

  it('swallows read errors (localStorage unavailable)', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('blocked');
    });
    expect(readVideoPokerStats('videopoker')).toEqual(emptyVideoPokerStats());
  });

  it('swallows write errors (quota exceeded)', () => {
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('quota');
    });
    expect(() => writeVideoPokerStats('videopoker', emptyVideoPokerStats())).not.toThrow();
  });
});
