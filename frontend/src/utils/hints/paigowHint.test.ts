import { describe, expect, it } from 'vitest';
import type { Card, PaiGowResponse } from '../../types/card';
import { PaiGowPhase } from '../../types/phases';
import { getPaiGowHint } from './paigowHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

/** ジョーカーを含まない 7 枚。自動分割が必ず答えを返す。 */
const PLAIN_HAND: Card[] = [
  card('SPADE', 13),
  card('HEART', 11),
  card('DIAMOND', 9),
  card('CLOVER', 7),
  card('SPADE', 5),
  card('HEART', 4),
  card('DIAMOND', 2),
];

/** ジョーカー入りの 7 枚。`paiGowAutoSplit` は null を返し、ボタンも無効になる。 */
const JOKER_HAND: Card[] = [card('JOKER', 0), ...PLAIN_HAND.slice(1)];

function base(overrides: Partial<PaiGowResponse> = {}) {
  return {
    playerCards: PLAIN_HAND,
    dealerCards: [],
    playerHighHand: [],
    playerLowHand: [],
    dealerHighHand: [],
    dealerLowHand: [],
    phase: PaiGowPhase.SET_HANDS,
    chips: 1000,
    bet: 10,
    result: 0,
    highHandResult: 0,
    lowHandResult: 0,
    payout: 0,
    commission: 0,
    playerHighRank: 0,
    playerLowRank: 0,
    dealerHighRank: 0,
    dealerLowRank: 0,
    ...overrides,
  } as PaiGowResponse;
}

describe('getPaiGowHint', () => {
  it('suggests betting while chips remain', () => {
    const hint = getPaiGowHint(base({ phase: PaiGowPhase.BET }));
    expect(hint?.targetAction).toBe('bet');
  });

  it('says nothing in the bet phase once the chips are gone', () => {
    expect(getPaiGowHint(base({ phase: PaiGowPhase.BET, chips: 0 }))).toBeNull();
  });

  it('returns null after the showdown', () => {
    expect(getPaiGowHint(base({ phase: PaiGowPhase.END }))).toBeNull();
  });

  it('points at the auto-split button when the page can compute one', () => {
    const hint = getPaiGowHint(base());
    expect(hint?.targetAction).toBe('autoSet');
    expect(hint?.reason).toBe('frontendHint.paigowAutoSplit');
  });

  it('explains the rule instead when the joker disables the auto split', () => {
    // ジョーカー入りでは `paiGowAutoSplit` が null を返し、A キーのボタンも無効。
    // ここで自動分割を勧めると、押しても何も起きないボタンを指すことになる。
    const hint = getPaiGowHint(base({ playerCards: JOKER_HAND }));
    expect(hint?.reason).toBe('frontendHint.paigowSplitByHand');
  });

  it('explains the rule when fewer than seven cards have been dealt', () => {
    // 枚数が揃っていないときも自動分割は null を返す。
    expect(getPaiGowHint(base({ playerCards: PLAIN_HAND.slice(0, 3) }))?.reason).toBe('frontendHint.paigowSplitByHand');
  });
});
