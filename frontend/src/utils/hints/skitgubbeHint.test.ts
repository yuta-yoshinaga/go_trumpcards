import { describe, expect, it } from 'vitest';
import type { SkitgubbeResponse } from '../../types/card';
import { getSkitgubbeHint } from './skitgubbeHint';

const base = {
  players: [],
  phase: 0,
  currentPlayerIdx: 0,
  stockCount: 0,
  trumpSuit: -1,
  duel: [],
  duelLeader: 0,
  pile: [],
  validIndices: [],
  canPickUp: false,
  gameEndFlag: false,
  loserIdx: -1,
  message: '',
};

const state = (reason?: string, extra?: Record<string, number | boolean>): SkitgubbeResponse =>
  ({ ...base, hint: reason ? { reason, pickUp: false, ...extra } : undefined }) as SkitgubbeResponse;

describe('getSkitgubbeHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getSkitgubbeHint(state())).toBeNull();
  });

  it('returns null for the non-suggestions', () => {
    for (const r of ['game_end', 'not_your_turn', 'none']) {
      expect(getSkitgubbeHint(state(`skitgubbe.hint.${r}`))).toBeNull();
    }
  });

  it('maps a duel suggestion to the play action', () => {
    expect(getSkitgubbeHint(state('skitgubbe.hint.duel', { cardIndex: 2 }))).toEqual({
      targetAction: 'play',
      reason: 'hint.duel',
      confidence: 'moderate',
    });
  });

  it('maps a beat suggestion to the play action', () => {
    expect(getSkitgubbeHint(state('skitgubbe.hint.beat', { cardIndex: 0 }))?.targetAction).toBe('play');
  });

  it('maps a pick-up suggestion to its own action', () => {
    // The pick-up is a footer button, not a card, so it must not point at a
    // hand index that would then be highlighted.
    expect(getSkitgubbeHint(state('skitgubbe.hint.pickup', { pickUp: true }))).toEqual({
      targetAction: 'pickup',
      reason: 'hint.pickup',
      confidence: 'moderate',
    });
  });
});
