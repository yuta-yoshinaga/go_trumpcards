import { describe, expect, it } from 'vitest';
import type { LobaResponse } from '../../types/card';
import { getLobaHint } from './lobaHint';

const base = {
  players: [],
  phase: 0,
  currentPlayerIdx: 0,
  stockCount: 70,
  melds: [],
  roundNo: 0,
  knockOut: 101,
  roundWinner: -1,
  roundClean: false,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
};

const state = (reason?: string, extra?: Record<string, unknown>): LobaResponse =>
  ({ ...base, hint: reason ? { reason, drawStock: false, ...extra } : undefined }) as LobaResponse;

describe('getLobaHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getLobaHint(state())).toBeNull();
  });

  it('returns null for the non-suggestions', () => {
    for (const r of ['game_end', 'not_your_turn', 'none']) {
      expect(getLobaHint(state(`loba.hint.${r}`))).toBeNull();
    }
  });

  it('sends each suggestion to its own control', () => {
    // 引く・出す・捨てるは別のボタン。同じ action に潰すと、押せない場所が光る。
    expect(getLobaHint(state('loba.hint.draw', { drawStock: true }))?.targetAction).toBe('draw');
    expect(getLobaHint(state('loba.hint.meld', { cardIndices: [0, 1, 2] }))?.targetAction).toBe('meld');
    expect(getLobaHint(state('loba.hint.discard', { cardIndex: 3 }))?.targetAction).toBe('discard');
  });

  it('carries the reason key', () => {
    expect(getLobaHint(state('loba.hint.meld', { cardIndices: [0, 1, 2] }))).toEqual({
      targetAction: 'meld',
      reason: 'hint.meld',
      confidence: 'moderate',
    });
  });
});
