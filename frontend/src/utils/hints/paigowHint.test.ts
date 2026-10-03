import { describe, expect, it } from 'vitest';
import { makePaiGowState } from '../../test/stateFactories';
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
  return makePaiGowState({ playerCards: PLAIN_HAND, phase: PaiGowPhase.SET_HANDS, bet: 10, ...overrides });
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

  it('points at the auto-split button when the server supplies a split', () => {
    const hint = getPaiGowHint(base({ hint: { lowIdx0: 0, lowIdx1: 1, lowIsPair: false, reason: 'house_way' } }));
    expect(hint?.targetAction).toBe('autoSet');
    expect(hint?.reason).toBe('frontendHint.paigowAutoSplit');
  });

  it('points at auto-split for joker hands when the server supplies a split', () => {
    const hint = getPaiGowHint(
      base({ playerCards: JOKER_HAND, hint: { lowIdx0: 1, lowIdx1: 2, lowIsPair: false, reason: 'house_way' } }),
    );
    expect(hint?.reason).toBe('frontendHint.paigowAutoSplit');
  });

  it('explains the rule when the server has no split hint', () => {
    const hint = getPaiGowHint(base({ hint: null }));
    expect(hint?.reason).toBe('frontendHint.paigowSplitByHand');
  });
});
