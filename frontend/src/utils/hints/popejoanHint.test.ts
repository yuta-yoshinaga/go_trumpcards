import { describe, expect, it } from 'vitest';
import type { PopeJoanResponse } from '../../types/card';
import { getPopeJoanHint } from './popejoanHint';

const base = {
  players: [],
  phase: 0,
  validPlays: [],
  currentPlayerIdx: 0,
  compartments: [],
  trumpSuit: 1,
  awards: [],
  playedPile: [],
  runSuit: -1,
  runRank: 0,
  dealNo: 0,
  targetDeals: 5,
  dealWinner: -1,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
};

const state = (reason?: string, extra?: Record<string, unknown>): PopeJoanResponse =>
  ({ ...base, hint: reason ? { reason, ...extra } : undefined }) as PopeJoanResponse;

describe('getPopeJoanHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getPopeJoanHint(state())).toBeNull();
  });

  it('returns null for the non-suggestions', () => {
    for (const r of ['game_end', 'deal_end', 'not_your_turn', 'none']) {
      expect(getPopeJoanHint(state(`popejoan.hint.${r}`))).toBeNull();
    }
  });

  // 止まっているか途中かで「なぜその札か」がまるで違うので、reason は分ける。
  it('keeps lead and follow as separate reasons', () => {
    expect(getPopeJoanHint(state('popejoan.hint.lead', { cardIndex: 1 }))).toEqual({
      targetAction: 'play',
      reason: 'hint.lead',
      confidence: 'moderate',
    });
    expect(getPopeJoanHint(state('popejoan.hint.follow', { cardIndex: 2 }))?.reason).toBe('hint.follow');
  });

  it('sends both to the play control', () => {
    expect(getPopeJoanHint(state('popejoan.hint.lead', { cardIndex: 1 }))?.targetAction).toBe('play');
    expect(getPopeJoanHint(state('popejoan.hint.follow', { cardIndex: 2 }))?.targetAction).toBe('play');
  });
});
