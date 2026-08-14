import { describe, expect, it } from 'vitest';
import type { Card, CincinnatiResponse } from '../../types/card';
import { CincinnatiPhase } from '../../types/phases';
import { getCincinnatiHint } from './cincinnatiHint';

const card = (value: number): Card => ({ design: 'SPADE', value });

const seat = (over: Partial<CincinnatiResponse['seats'][number]> = {}) =>
  ({
    name: 'YOU',
    isHuman: true,
    chips: 1000,
    bet: 0,
    cards: [card(1), card(2), card(3), card(4), card(5)],
    folded: false,
    allIn: false,
    isTurn: true,
    handRank: 0,
    bestHand: [],
    wonAmount: 0,
    ...over,
  }) as CincinnatiResponse['seats'][number];

const state = (over: Partial<CincinnatiResponse> = {}) =>
  ({
    phase: CincinnatiPhase.BETTING,
    seats: [seat()],
    community: [],
    revealedCount: 0,
    communityTotal: 5,
    pot: 40,
    currentBet: 0,
    toCall: 0,
    raiseCount: 0,
    canRaise: true,
    turnSeat: 0,
    humanSeat: 0,
    isHumanTurn: true,
    handNumber: 1,
    remainingCards: 30,
    winnerSeat: 0,
    gameEndFlag: false,
    message: '',
    config: { seats: 4, initialChips: 1000, ante: 10 },
    ...over,
  }) as CincinnatiResponse;

describe('getCincinnatiHint', () => {
  it('終局・他フェーズ・他人の手番では助言しない', () => {
    expect(getCincinnatiHint(state({ gameEndFlag: true }))).toBeNull();
    expect(getCincinnatiHint(state({ phase: CincinnatiPhase.SHOWDOWN }))).toBeNull();
    expect(getCincinnatiHint(state({ isHumanTurn: false }))).toBeNull();
    expect(getCincinnatiHint(state({ seats: [] }))).toBeNull();
  });

  it('賭けが無ければただで次を見る', () => {
    const hint = getCincinnatiHint(state());
    expect(hint?.targetAction).toBe('check');
    expect(hint?.reason).toBe('frontendHint.cincinnatiSeeAnotherCard');
  });

  it('役があって賭けが無ければ賭ける', () => {
    const hint = getCincinnatiHint(state({ seats: [seat({ handRank: 3 })] }));
    expect(hint?.targetAction).toBe('bet');
  });

  it('レイズ上限に達していれば賭けを薦めない', () => {
    const hint = getCincinnatiHint(state({ seats: [seat({ handRank: 3 })], canRaise: false }));
    expect(hint?.targetAction).toBe('check');
  });

  it('強い役で賭けを受けたら押す', () => {
    const hint = getCincinnatiHint(state({ toCall: 20, seats: [seat({ handRank: 4 })] }));
    expect(hint?.targetAction).toBe('raise');
  });

  it('役があればコールする', () => {
    const hint = getCincinnatiHint(state({ toCall: 20, seats: [seat({ handRank: 2 })] }));
    expect(hint?.targetAction).toBe('call');
    expect(hint?.reason).toBe('frontendHint.cincinnatiWorthACall');
  });

  // **五回ぶんのベットが残っている。** 役の無い手は安易に持ち越さない。
  it('役が無く高ければ降りる', () => {
    const hint = getCincinnatiHint(state({ toCall: 200 }));
    expect(hint?.targetAction).toBe('fold');
    expect(hint?.reason).toBe('frontendHint.cincinnatiNotWorthIt');
  });

  it('役が無くても安ければ見る', () => {
    const hint = getCincinnatiHint(state({ toCall: 10 }));
    expect(hint?.targetAction).toBe('call');
    expect(hint?.reason).toBe('frontendHint.cincinnatiCheapToStay');
  });
});
