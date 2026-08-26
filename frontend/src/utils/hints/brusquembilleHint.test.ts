import { describe, expect, it } from 'vitest';
import type { BrusquembilleResponse } from '../../types/card';
import { getBrusquembilleHint } from './brusquembilleHint';

function makeState(overrides?: Partial<BrusquembilleResponse>): BrusquembilleResponse {
  return {
    players: [],
    phase: 0,
    trickNumber: 1,
    currentPlayerIdx: 0,
    currentTrick: [],
    trumpSuit: 1,
    deckCount: 30,
    gameEndFlag: false,
    winnerIdx: -1,
    leadPlayerIdx: 0,
    message: '',
    ...overrides,
  } as BrusquembilleResponse;
}

describe('getBrusquembilleHint', () => {
  it('points at the card the backend recommends', () => {
    const hint = getBrusquembilleHint(makeState({ hint: { cardIndex: 2, reason: 'lead_low' } }));
    expect(hint?.targetAction).toBe('play');
    expect(hint?.reason).toBe('hint.lead_low');
  });

  it('returns null without a hint', () => {
    expect(getBrusquembilleHint(makeState())).toBeNull();
  });

  // **カード番号の無いヒントもある。**その場合は指す先が無い。
  it('returns null when the hint names no card', () => {
    expect(getBrusquembilleHint(makeState({ hint: { reason: 'lead_low' } }))).toBeNull();
    expect(getBrusquembilleHint(makeState({ hint: { cardIndex: -1, reason: 'lead_low' } }))).toBeNull();
  });
});
