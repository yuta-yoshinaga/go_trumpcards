import { describe, expect, it } from 'vitest';
import type { LaughAndLieDownResponse } from '../../types/card';
import { getLaughAndLieDownHint } from './laughandliedownHint';

const base = {
  players: [],
  layout: [],
  phase: 0,
  currentPlayerIdx: 0,
  validIndices: [],
  threeTakeIndices: [],
  dealerIdx: 0,
  lastInIdx: -1,
  lastInBonus: 5,
  pot: 11,
  gameEndFlag: false,
  message: '',
};

const state = (reason?: string, extra?: Record<string, number>): LaughAndLieDownResponse =>
  ({ ...base, hint: reason ? { reason, takeCount: 1, ...extra } : undefined }) as LaughAndLieDownResponse;

describe('getLaughAndLieDownHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getLaughAndLieDownHint(state())).toBeNull();
  });

  it('returns null for the non-suggestions', () => {
    for (const r of ['game_end', 'not_your_turn']) {
      expect(getLaughAndLieDownHint(state(`laughandliedown.hint.${r}`))).toBeNull();
    }
  });

  it('maps a one-card take', () => {
    expect(getLaughAndLieDownHint(state('laughandliedown.hint.take_one', { cardIndex: 2 }))).toEqual({
      targetAction: 'play',
      reason: 'hint.take_one',
      confidence: 'moderate',
    });
  });

  it('maps a three-card take to its own reason', () => {
    // 1 枚取りと 3 枚取りは別の手なので、同じ文言にまとめてはいけない。
    expect(getLaughAndLieDownHint(state('laughandliedown.hint.take_three', { cardIndex: 0 }))?.reason).toBe(
      'hint.take_three',
    );
  });

  it('still reports that the hand is about to go to the table', () => {
    // 選べる手ではないが、手札 8 枚が丸ごと相手の獲物になる直前なので黙らない。
    expect(getLaughAndLieDownHint(state('laughandliedown.hint.must_lie_down'))?.reason).toBe('hint.must_lie_down');
  });
});
