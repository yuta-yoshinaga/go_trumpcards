import { describe, expect, it } from 'vitest';
import type { ChineseTenResponse } from '../../types/card';
import { getChineseTenHint } from './chinesetenHint';

const base = {
  players: [],
  layout: [],
  phase: 0,
  currentPlayerIdx: 0,
  stockCount: 0,
  selectableIndices: [],
  tieScore: 105,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
};

const state = (reason?: string, extra?: Record<string, number>): ChineseTenResponse =>
  ({ ...base, hint: reason ? { reason, ...extra } : undefined }) as ChineseTenResponse;

describe('getChineseTenHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getChineseTenHint(state())).toBeNull();
  });

  it('returns null for the non-suggestions', () => {
    for (const r of ['game_end', 'not_your_turn', 'none']) {
      expect(getChineseTenHint(state(`chineseten.hint.${r}`))).toBeNull();
    }
  });

  it('maps a play suggestion', () => {
    expect(getChineseTenHint(state('chineseten.hint.play', { cardIndex: 2 }))).toEqual({
      targetAction: 'play',
      reason: 'hint.play',
      confidence: 'moderate',
    });
  });

  it('maps a select suggestion to the select action', () => {
    expect(getChineseTenHint(state('chineseten.hint.select', { layoutIndex: 1 }))?.targetAction).toBe('select');
  });
});
