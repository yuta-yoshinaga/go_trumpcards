import { describe, expect, it } from 'vitest';
import {
  VIDEO_POKER_MAX_BET,
  videoPokerHandNameToRowKey,
  videoPokerPayoutCell,
  videoPokerPayoutRows,
  videoPokerRowKey,
} from './videoPokerPayout';

describe('videoPokerPayoutRows', () => {
  it('returns the Jacks or Better table for videopoker', () => {
    const rows = videoPokerPayoutRows('videopoker');
    expect(rows[0].key).toBe('royalFlush');
    expect(rows.at(-1)?.key).toBe('jacksOrBetter');
    expect(rows).toHaveLength(9);
  });
  it('returns the Deuces Wild table with four deuces and wild royal', () => {
    const rows = videoPokerPayoutRows('deuceswild');
    const keys = rows.map((r) => r.key);
    expect(keys).toContain('fourDeuces');
    expect(keys).toContain('wildRoyalFlush');
    expect(keys).toContain('naturalRoyalFlush');
    expect(rows).toHaveLength(10);
  });
  it('returns the Joker Poker table with five of a kind and kings or better', () => {
    const rows = videoPokerPayoutRows('jokerpoker');
    const keys = rows.map((r) => r.key);
    expect(keys).toContain('fiveOfAKind');
    expect(keys).toContain('kingsOrBetter');
    expect(rows).toHaveLength(11);
  });
});

describe('videoPokerPayoutCell', () => {
  const royal = videoPokerPayoutRows('videopoker')[0];
  const straightFlush = videoPokerPayoutRows('videopoker')[1];

  it('scales linearly with the bet for normal hands', () => {
    expect(videoPokerPayoutCell(straightFlush, 1)).toBe(50);
    expect(videoPokerPayoutCell(straightFlush, 3)).toBe(150);
    expect(videoPokerPayoutCell(straightFlush, 5)).toBe(250);
  });
  it('pays the royal jackpot only at max bet', () => {
    expect(videoPokerPayoutCell(royal, 1)).toBe(250);
    expect(videoPokerPayoutCell(royal, 4)).toBe(1000);
    expect(videoPokerPayoutCell(royal, VIDEO_POKER_MAX_BET)).toBe(4000);
  });
});

describe('videoPokerHandNameToRowKey', () => {
  it('maps server hand names to row keys', () => {
    expect(videoPokerHandNameToRowKey('Royal Flush')).toBe('royalFlush');
    expect(videoPokerHandNameToRowKey('Natural Royal Flush')).toBe('naturalRoyalFlush');
    expect(videoPokerHandNameToRowKey('Wild Royal Flush')).toBe('wildRoyalFlush');
    expect(videoPokerHandNameToRowKey('Four Deuces')).toBe('fourDeuces');
    expect(videoPokerHandNameToRowKey('Five of a Kind')).toBe('fiveOfAKind');
    expect(videoPokerHandNameToRowKey('Four of a Kind')).toBe('fourOfAKind');
    expect(videoPokerHandNameToRowKey('Full House')).toBe('fullHouse');
    expect(videoPokerHandNameToRowKey('Straight Flush')).toBe('straightFlush');
    expect(videoPokerHandNameToRowKey('Straight')).toBe('straight');
    expect(videoPokerHandNameToRowKey('Three of a Kind')).toBe('threeOfAKind');
    expect(videoPokerHandNameToRowKey('Two Pair')).toBe('twoPair');
    expect(videoPokerHandNameToRowKey('Jacks or Better')).toBe('jacksOrBetter');
    expect(videoPokerHandNameToRowKey('Kings or Better')).toBe('kingsOrBetter');
    expect(videoPokerHandNameToRowKey('Flush')).toBe('flush');
  });
  it('returns null for non-paying / unknown hands', () => {
    expect(videoPokerHandNameToRowKey('High Card')).toBeNull();
    expect(videoPokerHandNameToRowKey('')).toBeNull();
  });
});

describe('videoPokerRowKey', () => {
  it('prefers the stable handKey over the English handName', () => {
    // handKey wins even when handName would map elsewhere.
    expect(videoPokerRowKey('wildRoyalFlush', 'Four of a Kind')).toBe('wildRoyalFlush');
    expect(videoPokerRowKey('fiveOfAKind', '')).toBe('fiveOfAKind');
  });
  it('falls back to the handName reverse-lookup when handKey is absent', () => {
    expect(videoPokerRowKey(undefined, 'Wild Royal Flush')).toBe('wildRoyalFlush');
    expect(videoPokerRowKey('', 'Four Deuces')).toBe('fourDeuces');
  });
  it('returns null when neither a key nor a known name is present', () => {
    expect(videoPokerRowKey(undefined, '')).toBeNull();
    expect(videoPokerRowKey('', 'High Card')).toBeNull();
  });
});
