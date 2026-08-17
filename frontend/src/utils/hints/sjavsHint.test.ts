import { describe, expect, it } from 'vitest';
import type { SjavsResponse } from '../../types/card';
import { getSjavsHint } from './sjavsHint';

const base = {
  players: [],
  phase: 0,
  currentPlayerIdx: 0,
  dealerIdx: 0,
  trumpSuit: -1,
  trumpCount: 0,
  bidderIdx: -1,
  bidLength: 0,
  minBid: 5,
  myLongest: 6,
  trick: [],
  trickNo: 0,
  validIndices: [],
  trumpIndices: [],
  teamPoints: [0, 0],
  remaining: [24, 24],
  crosses: [0, 0],
  carryOver: 0,
  gameEndFlag: false,
  winnerTeam: -1,
  doubleVictory: false,
  message: '',
};

const state = (reason?: string, extra?: Record<string, number>): SjavsResponse =>
  ({ ...base, hint: reason ? { reason, ...extra } : undefined }) as SjavsResponse;

describe('getSjavsHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getSjavsHint(state())).toBeNull();
  });

  it('returns null for the non-suggestions', () => {
    for (const r of ['game_end', 'not_your_turn', 'none']) {
      expect(getSjavsHint(state(`sjavs.hint.${r}`))).toBeNull();
    }
  });

  it('points a bid at the bid controls, not at a card', () => {
    // ビッドとパスはカードではなくボタンを指す。play に落とすと、押せない
    // カードがハイライトされる。
    expect(getSjavsHint(state('sjavs.hint.bid', { bidLength: 6 }))).toEqual({
      targetAction: 'bid',
      reason: 'hint.bid',
      confidence: 'moderate',
    });
    expect(getSjavsHint(state('sjavs.hint.pass', { bidLength: 0 }))?.targetAction).toBe('bid');
  });

  it('points a play at the cards', () => {
    expect(getSjavsHint(state('sjavs.hint.play', { cardIndex: 2 }))).toEqual({
      targetAction: 'play',
      reason: 'hint.play',
      confidence: 'moderate',
    });
  });
});
