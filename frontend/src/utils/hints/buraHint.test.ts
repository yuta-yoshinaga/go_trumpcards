import { describe, expect, it } from 'vitest';
import type { BuraResponse } from '../../types/card';
import { getBuraHint } from './buraHint';

const base = {
  players: [],
  phase: 0,
  trickNumber: 0,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  currentLead: [],
  trumpSuit: 2,
  stockRemaining: 0,
  winThreshold: 31,
  gameEndFlag: false,
  winnerIdx: -1,
  isDraw: false,
  message: '',
};

const state = (reason?: string, cardIndices?: number[]): BuraResponse =>
  ({ ...base, hint: reason ? { reason, cardIndices } : undefined }) as BuraResponse;

describe('getBuraHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getBuraHint(state())).toBeNull();
  });

  it('returns null for the non-suggestions', () => {
    // These two say "there is nothing to suggest", so surfacing them as a
    // tooltip would put an empty recommendation on screen.
    expect(getBuraHint(state('bura.hint.game_end'))).toBeNull();
    expect(getBuraHint(state('bura.hint.not_your_turn'))).toBeNull();
  });

  it('maps a card suggestion to the play action with moderate confidence', () => {
    expect(getBuraHint(state('bura.hint.lead', [0, 1]))).toEqual({
      targetAction: 'play',
      reason: 'hint.lead',
      confidence: 'moderate',
    });
    expect(getBuraHint(state('bura.hint.respond', [2]))?.targetAction).toBe('play');
  });

  it('reports claim and declare as strong -- both end the round on the spot', () => {
    expect(getBuraHint(state('bura.hint.claim'))).toEqual({
      targetAction: 'claim',
      reason: 'hint.claim',
      confidence: 'strong',
    });
    expect(getBuraHint(state('bura.hint.declare'))?.confidence).toBe('strong');
  });
});
