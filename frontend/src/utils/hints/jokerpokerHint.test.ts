import { describe, expect, it } from 'vitest';
import type { Card, VideoPokerResponse } from '../../types/card';
import { VideoPokerPhase } from '../../types/phases';
import { getJokerPokerHint } from './jokerpokerHint';

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

describe('getJokerPokerHint', () => {
  it('returns null in RESULT phase', () => {
    expect(getJokerPokerHint(makeState({ phase: VideoPokerPhase.RESULT }))).toBeNull();
  });

  it('treats JOKER design as wild', () => {
    const hand = [
      makeCard('JOKER', 0),
      makeCard('HEART', 5),
      makeCard('DIAMOND', 3),
      makeCard('CLOVER', 8),
      makeCard('SPADE', 10),
    ];
    const result = getJokerPokerHint(makeState({ hand }));
    expect(result?.reason).toBe('hint.holdWild');
    expect(result?.confidence).toBe('strong');
    expect(result?.targetAction).toContain('0');
  });

  it('does not treat 2s as wild', () => {
    const hand = [
      makeCard('SPADE', 2),
      makeCard('HEART', 4),
      makeCard('DIAMOND', 6),
      makeCard('CLOVER', 8),
      makeCard('SPADE', 10),
    ];
    const result = getJokerPokerHint(makeState({ hand }));
    expect(result?.reason).toBe('hint.drawAll');
  });

  it('suggests standard hold for hands without jokers', () => {
    const hand = [
      makeCard('SPADE', 10),
      makeCard('HEART', 10),
      makeCard('DIAMOND', 3),
      makeCard('CLOVER', 8),
      makeCard('SPADE', 5),
    ];
    const result = getJokerPokerHint(makeState({ hand }));
    expect(result?.reason).toBe('hint.holdPair');
  });

  it('holds joker with non-wild pairs', () => {
    const hand = [
      makeCard('JOKER', 0),
      makeCard('HEART', 7),
      makeCard('DIAMOND', 7),
      makeCard('CLOVER', 3),
      makeCard('SPADE', 10),
    ];
    const result = getJokerPokerHint(makeState({ hand }));
    expect(result?.targetAction).toBe('hold:0,1,2');
  });
});
