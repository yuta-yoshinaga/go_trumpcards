import { describe, expect, it } from 'vitest';
import type { NainJauneResponse } from '../../types/card';
import { getNainJauneHint } from './nainjauneHint';

const base = {
  players: [],
  phase: 0,
  currentPlayerIdx: 0,
  boxes: [],
  talonCount: 4,
  awards: [],
  playedPile: [],
  runRank: 0,
  dealNo: 0,
  targetDeals: 5,
  dealWinner: -1,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
};

const state = (reason?: string, extra?: Record<string, unknown>): NainJauneResponse =>
  ({ ...base, hint: reason ? { reason, ...extra } : undefined }) as NainJauneResponse;

describe('getNainJauneHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getNainJauneHint(state())).toBeNull();
  });

  it('returns null for the non-suggestions', () => {
    for (const r of ['game_end', 'deal_end', 'not_your_turn', 'none']) {
      expect(getNainJauneHint(state(`nainjaune.hint.${r}`))).toBeNull();
    }
  });

  // 止まっているか・続きか・区画を取れるかで「なぜその札か」がまるで違うので、
  // reason は分ける。
  it('keeps lead, follow and box as separate reasons', () => {
    expect(getNainJauneHint(state('nainjaune.hint.lead', { cardIndex: 1 }))).toEqual({
      targetAction: 'play',
      reason: 'hint.lead',
      confidence: 'moderate',
    });
    expect(getNainJauneHint(state('nainjaune.hint.follow', { cardIndex: 2 }))?.reason).toBe('hint.follow');
    expect(getNainJauneHint(state('nainjaune.hint.box', { cardIndex: 3 }))?.reason).toBe('hint.box');
  });

  it('sends them all to the play control', () => {
    for (const r of ['lead', 'follow', 'box']) {
      expect(getNainJauneHint(state(`nainjaune.hint.${r}`, { cardIndex: 1 }))?.targetAction).toBe('play');
    }
  });
});
