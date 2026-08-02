import { describe, expect, it } from 'vitest';
import type { Card, ChinesePokerResponse } from '../../types/card';
import { ChinesePokerPhase } from '../../types/phases';
import { getChinesePokerHint } from './chinesepokerHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

/**
 * 素直に並べても反則にならない 13 枚。後列 A-A-A-K-Q はスリーカード、中列 J-10-9-8-6 と
 * 前列 5-4-2 はハイカード。ランクは実測して選んである —— 最初に書いた 8-7-6-5-4 は
 * 中列がストレートになって後列を抜き、次に書いた前列 4-3-2 は **3 枚役ではストレート**
 * (`cpEvalThreeCardHand` は 4)で中列を抜いた。どちらも「合法な例」のつもりが反則例だった。
 *
 * A を 3 枚入れてあるのは、A を 1 のまま並べるとこの手が反則になるため
 * (`chinesePokerIsFoul` で確認済み)。エースの読み替えを外すと下のテストが落ちる。
 */
const LEGAL_HAND: Card[] = [
  card('SPADE', 1),
  card('HEART', 1),
  card('DIAMOND', 1),
  card('CLOVER', 13),
  card('SPADE', 12),
  card('HEART', 11),
  card('DIAMOND', 10),
  card('CLOVER', 9),
  card('SPADE', 8),
  card('HEART', 6),
  card('DIAMOND', 5),
  card('CLOVER', 4),
  card('SPADE', 2),
];

/** 一番低い 3 枚が 2 のスリーカードになる 13 枚。素直に並べると前列が中列を抜く。 */
const FOULING_HAND: Card[] = [
  card('SPADE', 1),
  card('HEART', 1),
  card('DIAMOND', 1),
  card('CLOVER', 13),
  card('SPADE', 12),
  card('HEART', 11),
  card('DIAMOND', 10),
  card('CLOVER', 9),
  card('SPADE', 8),
  card('HEART', 6),
  card('SPADE', 2),
  card('HEART', 2),
  card('DIAMOND', 2),
];

function base(overrides: Partial<ChinesePokerResponse> = {}) {
  return {
    playerCards: LEGAL_HAND,
    dealerCards: [],
    playerFront: [],
    playerMiddle: [],
    playerBack: [],
    dealerFront: [],
    dealerMiddle: [],
    dealerBack: [],
    phase: ChinesePokerPhase.SET_HANDS,
    chips: 1000,
    bet: 10,
    result: 0,
    frontResult: 0,
    middleResult: 0,
    backResult: 0,
    payout: 0,
    playerFrontRank: 0,
    playerMiddleRank: 0,
    playerBackRank: 0,
    dealerFrontRank: 0,
    dealerMiddleRank: 0,
    dealerBackRank: 0,
    playerRoyalty: 0,
    dealerRoyalty: 0,
    scoop: false,
    ...overrides,
  } as ChinesePokerResponse;
}

describe('getChinesePokerHint', () => {
  it('suggests betting while chips remain', () => {
    const hint = getChinesePokerHint(base({ phase: ChinesePokerPhase.BET }));
    expect(hint?.targetAction).toBe('bet');
    expect(hint?.reason).toBe('frontendHint.chinesepokerBet');
  });

  it('says nothing in the bet phase once the chips are gone', () => {
    expect(getChinesePokerHint(base({ phase: ChinesePokerPhase.BET, chips: 0 }))).toBeNull();
  });

  it('returns null after the showdown', () => {
    expect(getChinesePokerHint(base({ phase: ChinesePokerPhase.END }))).toBeNull();
  });

  it('recommends the rank-ordered split when it is legal', () => {
    expect(getChinesePokerHint(base())?.reason).toBe('frontendHint.chinesepokerSplit');
  });

  it('warns instead of recommending a split that would foul', () => {
    // 前列が 2 のスリーカード、中列は 8 ハイ。高い順に並べただけでは反則になる。
    expect(getChinesePokerHint(base({ playerCards: FOULING_HAND }))?.reason).toBe('frontendHint.chinesepokerFoulRisk');
  });

  it('treats the ace as the highest card, not as a one', () => {
    // A を 1 のまま扱うと前列に落ちる。そうなると前列 A-3-2 が中列 8 ハイを抜き、
    // 通るはずの手が反則警告になる。
    expect(getChinesePokerHint(base())?.reason).not.toBe('frontendHint.chinesepokerFoulRisk');
  });

  it('returns null before a full thirteen cards have been dealt', () => {
    expect(getChinesePokerHint(base({ playerCards: LEGAL_HAND.slice(0, 5) }))).toBeNull();
  });
});
