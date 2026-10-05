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

  it.each([
    ['handIsWorthTheBuy', 'pay'],
    ['buyIsCheapEnough', 'pay'],
    ['buyCostsTooMuch', 'fold'],
    ['strongEnoughToBet', 'bet'],
    ['seeAnotherCard', 'check'],
    ['strongEnoughToRaise', 'raise'],
    ['worthACall', 'call'],
    ['cheapToStay', 'call'],
    ['wildsRaiseTheBar', 'fold'],
  ])('サーバ助言 %s を対応するキーと操作へ写す', (reason, action) => {
    const hint = getBaseballpokerHint(state({ serverHint: { action, reason } }));
    expect(hint).toEqual({
      targetAction: action,
      reason: `frontendHint.baseball${reason[0].toUpperCase()}${reason.slice(1)}`,
      confidence: reason === 'seeAnotherCard' ? 'strong' : 'moderate',
    });
  });

  it('serverHint が無ければ助言しない', () => {
    expect(getBaseballpokerHint(state())).toBeNull();
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
