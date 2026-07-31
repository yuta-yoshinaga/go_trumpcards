import { describe, expect, it } from 'vitest';
import type { TrexResponse } from '../../types/card';
import { getTrexHint } from './trexHint';

const base = {
  players: [],
  phase: 0,
  currentPlayerIdx: 0,
  kingIdx: 0,
  contract: 5,
  availableContracts: [],
  isTrix: false,
  dealNo: 0,
  totalDeals: 20,
  trick: [],
  trickNo: 0,
  runs: [],
  finishOrder: [],
  validIndices: [],
  canPass: false,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
};

const state = (reason?: string, extra?: Record<string, number | boolean>): TrexResponse =>
  ({ ...base, hint: reason ? { reason, pass: false, ...extra } : undefined }) as TrexResponse;

describe('getTrexHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getTrexHint(state())).toBeNull();
  });

  it('returns null for the non-suggestions', () => {
    for (const r of ['game_end', 'not_your_turn', 'none']) {
      expect(getTrexHint(state(`trex.hint.${r}`))).toBeNull();
    }
  });

  it('points a contract choice at the contract buttons, not at a card', () => {
    // play に落とすと、選択フェーズで押せないカードがハイライトされる。
    expect(getTrexHint(state('trex.hint.choose', { contract: 2 }))).toEqual({
      targetAction: 'choose',
      reason: 'hint.choose',
      confidence: 'moderate',
    });
  });

  it('points a pass at the pass button', () => {
    expect(getTrexHint(state('trex.hint.pass', { pass: true }))?.targetAction).toBe('pass');
  });

  it('points a play at the cards', () => {
    expect(getTrexHint(state('trex.hint.play', { cardIndex: 3 }))?.targetAction).toBe('play');
  });
});
