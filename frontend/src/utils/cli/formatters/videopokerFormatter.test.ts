import { describe, expect, it } from 'vitest';
import type { VideoPokerResponse } from '../../../types/card';
import { formatVideopokerState } from './videopokerFormatter';

function makeState(overrides?: Partial<VideoPokerResponse>): VideoPokerResponse {
  return {
    hand: [],
    phase: 1,
    chips: 1000,
    betAmount: 0,
    result: 0,
    payout: 0,
    handRank: 0,
    handName: '',
    heldIndices: [false, false, false, false, false],
    variantName: 'jacksorbetter',
    hands: 0,
    winRate: 0,
    net: 0,
    message: '',
    ...overrides,
  };
}

describe('formatVideopokerState', () => {
  it('shows zero-valued session stats', () => {
    expect(formatVideopokerState(makeState())).toContain('hands: 0  win: 0%  net: +0');
  });

  it('shows a negative net with its sign', () => {
    expect(formatVideopokerState(makeState({ hands: 12, winRate: 41, net: -35 }))).toContain(
      'hands: 12  win: 41%  net: -35',
    );
  });
});
