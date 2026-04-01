import { describe, expect, it } from 'vitest';
import type { Card, VideoPokerResponse } from '../../types/card';
import { VideoPokerPhase } from '../../types/phases';
import { getDeucesWildHint } from './deuceswildHint';

function makeCard(design: Card['design'], value: number): Card {
  return { design, value };
}

function makeState(overrides: Partial<VideoPokerResponse> = {}): VideoPokerResponse {
  return {
    hand: [
      makeCard('SPADE', 10),
      makeCard('HEART', 5),
      makeCard('DIAMOND', 3),
      makeCard('CLOVER', 8),
      makeCard('SPADE', 12),
    ],
    phase: VideoPokerPhase.DRAW,
    chips: 100,
    betAmount: 1,
    result: 0,
    payout: 0,
    handRank: 0,
    handName: '',
    heldIndices: [false, false, false, false, false],
    variantName: '',
    message: '',
    ...overrides,
  };
}

describe('getDeucesWildHint', () => {
  it('returns null in BET phase', () => {
    expect(getDeucesWildHint(makeState({ phase: VideoPokerPhase.BET }))).toBeNull();
  });

  it('treats 2s as wild cards', () => {
    const hand = [
      makeCard('SPADE', 2),
      makeCard('HEART', 5),
      makeCard('DIAMOND', 3),
      makeCard('CLOVER', 8),
      makeCard('SPADE', 10),
    ];
    const result = getDeucesWildHint(makeState({ hand }));
    expect(result?.reason).toBe('hint.holdWild');
    expect(result?.confidence).toBe('strong');
    expect(result?.targetAction).toContain('0');
  });

  it('holds wild 2s together with non-wild pairs', () => {
    const hand = [
      makeCard('SPADE', 2),
      makeCard('HEART', 7),
      makeCard('DIAMOND', 7),
      makeCard('CLOVER', 3),
      makeCard('SPADE', 10),
    ];
    const result = getDeucesWildHint(makeState({ hand }));
    expect(result?.targetAction).toBe('hold:0,1,2');
  });

  it('suggests standard hold for hands without 2s', () => {
    const hand = [
      makeCard('SPADE', 10),
      makeCard('HEART', 10),
      makeCard('DIAMOND', 3),
      makeCard('CLOVER', 8),
      makeCard('SPADE', 5),
    ];
    const result = getDeucesWildHint(makeState({ hand }));
    expect(result?.reason).toBe('hint.holdPair');
  });

  it('suggests draw all when nothing to hold and no wilds', () => {
    const hand = [
      makeCard('SPADE', 3),
      makeCard('HEART', 5),
      makeCard('DIAMOND', 7),
      makeCard('CLOVER', 9),
      makeCard('SPADE', 10),
    ];
    const result = getDeucesWildHint(makeState({ hand }));
    expect(result?.reason).toBe('hint.drawAll');
  });
});
