import { describe, expect, it } from 'vitest';
import type { ToepenResponse } from '../../types/card';
import { getToepenHint } from './toepenHint';

const base = {
  players: [],
  phase: 0,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 0,
  currentTrick: [],
  leadSuit: -1,
  trickNumber: 0,
  handNumber: 1,
  stake: 1,
  knockerIdx: -1,
  pendingRespondent: -1,
  lastTrickWinner: -1,
  maxLives: 10,
  validPlayIndices: [],
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
};

const state = (reason?: string, extra?: Record<string, unknown>): ToepenResponse =>
  ({ ...base, hint: reason ? { reason, ...extra } : undefined }) as ToepenResponse;

describe('getToepenHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getToepenHint(state())).toBeNull();
  });

  it('returns null for the non-suggestions', () => {
    for (const r of ['game_end', 'hand_end', 'not_your_turn', 'none']) {
      expect(getToepenHint(state(`toepen.hint.${r}`))).toBeNull();
    }
  });

  it('maps a card suggestion to the play action', () => {
    expect(getToepenHint(state('toepen.hint.play', { cardIndex: 1 }))).toEqual({
      targetAction: 'play',
      reason: 'hint.play',
      confidence: 'moderate',
    });
  });

  it('reports folding as strong and staying as moderate', () => {
    // Folding costs the stake BEFORE the raise, so it is the cheap exit from a
    // hand you cannot win -- worth a firmer nudge than staying in.
    expect(getToepenHint(state('toepen.hint.fold', { fold: true }))).toEqual({
      targetAction: 'fold',
      reason: 'hint.fold',
      confidence: 'strong',
    });
    expect(getToepenHint(state('toepen.hint.stay', { fold: false }))?.confidence).toBe('moderate');
  });
});
