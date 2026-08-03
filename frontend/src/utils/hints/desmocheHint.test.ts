import { describe, expect, it } from 'vitest';
import type { DesmocheResponse } from '../../types/card';
import { getDesmocheHint } from './desmocheHint';

const base = {
  players: [],
  phase: 0,
  currentPlayerIdx: 0,
  stockCount: 15,
  melds: [],
  roundNo: 0,
  pot: 40,
  goOutSize: 10,
  roundWinner: -1,
  roundExhausted: false,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
};

const state = (reason?: string, extra?: Record<string, unknown>): DesmocheResponse =>
  ({ ...base, hint: reason ? { reason, drawStock: false, ...extra } : undefined }) as DesmocheResponse;

describe('getDesmocheHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getDesmocheHint(state())).toBeNull();
  });

  it('returns null for the non-suggestions', () => {
    for (const r of ['game_end', 'not_your_turn', 'none']) {
      expect(getDesmocheHint(state(`desmoche.hint.${r}`))).toBeNull();
    }
  });

  it('sends each suggestion to its own control', () => {
    // 引く・出す・捨てるは別のボタン。同じ action に潰すと、押せない場所が光る。
    expect(getDesmocheHint(state('desmoche.hint.draw', { drawStock: true }))?.targetAction).toBe('draw');
    expect(getDesmocheHint(state('desmoche.hint.meld', { cardIndices: [0, 1, 2] }))?.targetAction).toBe('meld');
    expect(getDesmocheHint(state('desmoche.hint.discard', { cardIndex: 3 }))?.targetAction).toBe('discard');
  });

  it('carries the reason key', () => {
    expect(getDesmocheHint(state('desmoche.hint.meld', { cardIndices: [0, 1, 2] }))).toEqual({
      targetAction: 'meld',
      reason: 'hint.meld',
      confidence: 'moderate',
    });
  });
});
