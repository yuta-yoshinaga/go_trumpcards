import { describe, expect, it } from 'vitest';
import {
  VIDEO_POKER_MAX_BET,
  videoPokerHandNameToRowKey,
  videoPokerPayoutCell,
  videoPokerPayoutRows,
} from './videoPokerPayout';

describe('videoPokerPayoutRows', () => {
  it('returns the Jacks or Better table for videopoker', () => {
    const rows = videoPokerPayoutRows('videopoker');
    expect(rows[0].key).toBe('royalFlush');
    expect(rows.at(-1)?.key).toBe('jacksOrBetter');
    expect(rows).toHaveLength(9);
  });
  it('returns the Deuces Wild table with four deuces and wild royal', () => {
    const keys = videoPokerPayoutRows('deuceswild').map((r) => r.key);
    expect(keys).toContain('fourDeuces');
    expect(keys).toContain('wildRoyalFlush');
    expect(keys).toContain('naturalRoyalFlush');
  });
  it('returns the Joker Poker table with five of a kind and kings or better', () => {
    const keys = videoPokerPayoutRows('jokerpoker').map((r) => r.key);
    expect(keys).toContain('fiveOfAKind');
    expect(keys).toContain('kingsOrBetter');
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
    expect(videoPokerHandNameToRowKey('Natural Royal Flush')).toBe('naturalRoyalFlush');
    expect(videoPokerHandNameToRowKey('Jacks or Better')).toBe('jacksOrBetter');
    expect(videoPokerHandNameToRowKey('Kings or Better')).toBe('kingsOrBetter');
    expect(videoPokerHandNameToRowKey('Flush')).toBe('flush');
  });
  it('returns null for non-paying / unknown hands', () => {
    expect(videoPokerHandNameToRowKey('High Card')).toBeNull();
    expect(videoPokerHandNameToRowKey('')).toBeNull();
  });
});
