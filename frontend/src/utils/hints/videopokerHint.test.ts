import { describe, expect, it } from 'vitest';
import type { Card, VideoPokerResponse } from '../../types/card';
import { VideoPokerPhase } from '../../types/phases';
import { getVideoPokerHint } from './videopokerHint';

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

describe('getVideoPokerHint', () => {
  it('returns null in BET phase', () => {
    expect(getVideoPokerHint(makeState({ phase: VideoPokerPhase.BET }))).toBeNull();
  });

  it('returns null in RESULT phase', () => {
    expect(getVideoPokerHint(makeState({ phase: VideoPokerPhase.RESULT }))).toBeNull();
  });

  it('suggests holding a pair', () => {
    const hand = [
      makeCard('SPADE', 10),
      makeCard('HEART', 10),
      makeCard('DIAMOND', 3),
      makeCard('CLOVER', 8),
      makeCard('SPADE', 5),
    ];
    const result = getVideoPokerHint(makeState({ hand }));
    expect(result?.reason).toBe('hint.holdPair');
  });

  it('suggests holding high cards', () => {
    const hand = [
      makeCard('SPADE', 11),
      makeCard('HEART', 3),
      makeCard('DIAMOND', 4),
      makeCard('CLOVER', 6),
      makeCard('SPADE', 13),
    ];
    const result = getVideoPokerHint(makeState({ hand }));
    expect(result?.reason).toBe('hint.holdHighCards');
  });

  it('does not treat 2s as wild', () => {
    const hand = [
      makeCard('SPADE', 2),
      makeCard('HEART', 4),
      makeCard('DIAMOND', 6),
      makeCard('CLOVER', 8),
      makeCard('SPADE', 10),
    ];
    const result = getVideoPokerHint(makeState({ hand }));
    expect(result?.reason).toBe('hint.drawAll');
  });

  // The negative control for #6301: the same hand that Deuces Wild must not
  // call a paying pair is one Jacks or Better should, so the fix cannot have
  // been a blanket removal.
  it('still recommends a paying pair in Jacks or Better', () => {
    const hint = getVideoPokerHint(
      makeState({
        hand: [
          { design: 'SPADE', value: 13 },
          { design: 'HEART', value: 13 },
          { design: 'SPADE', value: 12 },
          { design: 'SPADE', value: 11 },
          { design: 'SPADE', value: 10 },
        ],
      }),
    );
    expect(hint?.reason).toBe('hint.holdPair');
  });
});
