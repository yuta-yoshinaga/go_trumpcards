import { describe, expect, it } from 'vitest';
import type { BaseballPokerResponse, Card } from '../../types/card';
import { BaseballPhase } from '../../types/phases';
import { getBaseballpokerHint, isBaseballWild } from './baseballpokerHint';

const card = (value: number): Card => ({ design: 'SPADE', value });

const seat = (over: Partial<BaseballPokerResponse['seats'][number]> = {}) =>
  ({
    name: 'YOU',
    isHuman: true,
    chips: 1000,
    bet: 0,
    cards: [card(1), card(2), card(5)],
    faceUp: [false, false, true],
    bonusCards: 0,
    folded: false,
    allIn: false,
    isTurn: true,
    isBuying: false,
    handRank: 0,
    usedWild: false,
    bestHand: [],
    wonAmount: 0,
    ...over,
  }) as BaseballPokerResponse['seats'][number];

const state = (over: Partial<BaseballPokerResponse> = {}) =>
  ({
    phase: BaseballPhase.BETTING,
    seats: [seat()],
    street: 1,
    streetTotal: 4,
    wildValues: [3, 9],
    bonusValue: 4,
    buyInValue: 3,
    pot: 40,
    currentBet: 0,
    toCall: 0,
    raiseCount: 0,
    canRaise: true,
    turnSeat: 0,
    humanSeat: 0,
    isHumanTurn: true,
    buyerSeat: -1,
    buyCost: 0,
    isBuying: false,
    handNumber: 1,
    remainingCards: 30,
    winnerSeat: 0,
    gameEndFlag: false,
    message: '',
    config: { seats: 4, initialChips: 1000, ante: 10 },
    ...over,
  }) as BaseballPokerResponse;

describe('getBaseballpokerHint', () => {
  it('終局では助言しない', () => {
    expect(getBaseballpokerHint(state({ gameEndFlag: true }))).toBeNull();
  });

  it('他人の手番では助言しない', () => {
    expect(getBaseballpokerHint(state({ isHumanTurn: false }))).toBeNull();
  });

  it('席が見つからなければ助言しない', () => {
    expect(getBaseballpokerHint(state({ humanSeat: 9 }))).toBeNull();
  });

  // **買い増しの返事が最優先。** その場で払うか降りるかしかない。
  it('役が強ければ買い増しを薦める', () => {
    const hint = getBaseballpokerHint(
      state({ isBuying: true, phase: BaseballPhase.BUY_IN, buyCost: 400, seats: [seat({ handRank: 3 })] }),
    );
    expect(hint?.targetAction).toBe('pay');
  });

  it('安ければ役が無くても買い増しを薦める', () => {
    const hint = getBaseballpokerHint(state({ isBuying: true, phase: BaseballPhase.BUY_IN, buyCost: 100 }));
    expect(hint?.targetAction).toBe('pay');
  });

  it('手持ちに対して重ければ降りるよう薦める', () => {
    const hint = getBaseballpokerHint(state({ isBuying: true, phase: BaseballPhase.BUY_IN, buyCost: 900 }));
    expect(hint?.targetAction).toBe('fold');
  });

  it('自分が迫られていない買い増し中は助言しない', () => {
    expect(getBaseballpokerHint(state({ phase: BaseballPhase.BUY_IN, isBuying: false }))).toBeNull();
  });

  it('賭けが無ければチェックを薦める', () => {
    expect(getBaseballpokerHint(state())?.targetAction).toBe('check');
  });

  it('スリーカード以上なら賭けを薦める', () => {
    expect(getBaseballpokerHint(state({ seats: [seat({ handRank: 4 })] }))?.targetAction).toBe('bet');
  });

  it('レイズ上限に達していたら賭けを薦めない', () => {
    expect(getBaseballpokerHint(state({ seats: [seat({ handRank: 4 })], canRaise: false }))?.targetAction).toBe(
      'check',
    );
  });

  it('フラッシュ以上ならレイズを薦める', () => {
    expect(getBaseballpokerHint(state({ toCall: 20, seats: [seat({ handRank: 6 })] }))?.targetAction).toBe('raise');
  });

  // **ワイルドが8枚あるので相場が上がる。** ツーペア未満はコールに値しない。
  it('二段以上ならコール、それ未満で高ければ降りるよう薦める', () => {
    expect(getBaseballpokerHint(state({ toCall: 20, seats: [seat({ handRank: 3 })] }))?.targetAction).toBe('call');
    expect(getBaseballpokerHint(state({ toCall: 200, seats: [seat({ handRank: 2 })] }))?.targetAction).toBe('fold');
  });

  it('安ければ役が無くてもコールを薦める', () => {
    expect(getBaseballpokerHint(state({ toCall: 5 }))?.targetAction).toBe('call');
  });
});

describe('isBaseballWild', () => {
  // **サーバが送った値で判定する。** 画面が 3 と 9 を持たない証拠。
  it('サーバが送った値だけをワイルドとする', () => {
    expect(isBaseballWild(state(), 3)).toBe(true);
    expect(isBaseballWild(state(), 9)).toBe(true);
    expect(isBaseballWild(state(), 4)).toBe(false);
    expect(isBaseballWild(state({ wildValues: [2] }), 3)).toBe(false);
    expect(isBaseballWild(state({ wildValues: [2] }), 2)).toBe(true);
  });
});
