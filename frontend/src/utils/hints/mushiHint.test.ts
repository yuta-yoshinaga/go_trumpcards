import { describe, expect, it } from 'vitest';
import type { MushiResponse } from '../../types/card';
import { getMushiHint } from './mushiHint';

const base = {
  players: [],
  field: [],
  phase: 0,
  currentPlayerIdx: 0,
  dealerIdx: 0,
  roundNumber: 1,
  targetRounds: 12,
  stockCount: 0,
  selectableIndices: [],
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
};

const state = (reason?: string, extra?: Record<string, number>): MushiResponse =>
  ({ ...base, hint: reason ? { reason, ...extra } : undefined }) as MushiResponse;

describe('getMushiHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getMushiHint(state())).toBeNull();
  });

  it('returns null for the non-suggestions', () => {
    // Each of these says "there is nothing to suggest"; surfacing them would
    // put an empty recommendation on screen.
    for (const r of ['game_end', 'round_end', 'not_your_turn', 'none']) {
      expect(getMushiHint(state(`mushi.hint.${r}`))).toBeNull();
    }
  });

  it('maps a play suggestion', () => {
    expect(getMushiHint(state('mushi.hint.play', { cardIndex: 2 }))).toEqual({
      targetAction: 'play',
      reason: 'hint.play',
      confidence: 'moderate',
    });
  });

  it('maps a select suggestion to the select action', () => {
    expect(getMushiHint(state('mushi.hint.select', { fieldIndex: 1 }))?.targetAction).toBe('select');
  });
});
