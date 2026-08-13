import { describe, expect, it } from 'vitest';
import type { FreeBetResponse } from '../../types/card';
import { FreeBetPhase } from '../../types/phases';
import { getFreebetHint } from './freebetHint';

const hand = (score: number) =>
  ({
    cards: [],
    score,
    bet: 50,
    freeBet: 0,
    isSoft: false,
    stood: false,
    doubled: false,
    busted: false,
    blackjack: false,
    result: 0,
  }) as FreeBetResponse['hands'][number];

const state = (over: Partial<FreeBetResponse> = {}) =>
  ({
    phase: FreeBetPhase.PLAY,
    hands: [hand(15)],
    activeHand: 0,
    dealerCards: [],
    dealerScore: 0,
    dealerPushed22: false,
    canFreeDouble: false,
    canFreeSplit: false,
    anteBet: 50,
    payout: 0,
    chips: 1000,
    roundNumber: 1,
    remainingCards: 312,
    gameEndFlag: false,
    message: '',
    ...over,
  }) as FreeBetResponse;

describe('getFreebetHint', () => {
  it('終局と賭けフェーズでは助言しない', () => {
    expect(getFreebetHint(state({ gameEndFlag: true }))).toBeNull();
    expect(getFreebetHint(state({ phase: FreeBetPhase.BET }))).toBeNull();
    expect(getFreebetHint(state({ phase: FreeBetPhase.RESULT }))).toBeNull();
  });

  // **無料の操作が最優先。** 上乗せぶんを失うことがないので、勝率が五分を
  // 割る手札でも取るのが正しい。ここを点数の定石より後ろに置くと助言が誤る。
  it('無料スプリットが使えるならそれを薦める', () => {
    const hint = getFreebetHint(state({ canFreeSplit: true, hands: [hand(20)] }));
    expect(hint?.targetAction).toBe('freesplit');
    expect(hint?.confidence).toBe('strong');
  });

  it('無料ダブルが使えるならそれを薦める', () => {
    const hint = getFreebetHint(state({ canFreeDouble: true, hands: [hand(11)] }));
    expect(hint?.targetAction).toBe('freedouble');
  });

  it('両方使えるならスプリットを先に薦める', () => {
    const hint = getFreebetHint(state({ canFreeDouble: true, canFreeSplit: true }));
    expect(hint?.targetAction).toBe('freesplit');
  });

  it('11以下は引く', () => {
    expect(getFreebetHint(state({ hands: [hand(11)] }))?.targetAction).toBe('hit');
    expect(getFreebetHint(state({ hands: [hand(5)] }))?.targetAction).toBe('hit');
  });

  it('17以上は立つ', () => {
    expect(getFreebetHint(state({ hands: [hand(17)] }))?.targetAction).toBe('stand');
    expect(getFreebetHint(state({ hands: [hand(20)] }))?.targetAction).toBe('stand');
  });

  // 12〜16 は、ディーラーの 22 が引き分けになるぶん立つほうが得。
  it('12〜16は立つ', () => {
    const hint = getFreebetHint(state({ hands: [hand(15)] }));
    expect(hint?.targetAction).toBe('stand');
    expect(hint?.reason).toBe('frontendHint.freeBetDealerMayBust');
  });

  it('操作中の手札が無ければ助言しない', () => {
    expect(getFreebetHint(state({ hands: [], activeHand: 0 }))).toBeNull();
  });
});
