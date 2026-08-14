import { describe, expect, it } from 'vitest';
import type { Card, DoubleAttackResponse } from '../../types/card';
import { DoubleAttackPhase } from '../../types/phases';
import { getDoubleattackHint } from './doubleattackHint';

const card = (value: number): Card => ({ design: 'SPADE', value });

const base = {
  phase: DoubleAttackPhase.BET,
  gameEndFlag: false,
  dealerCards: [],
  hands: [],
  activeHand: 0,
} as unknown as DoubleAttackResponse;

const at = (over: Partial<DoubleAttackResponse>) => ({ ...base, ...over }) as DoubleAttackResponse;

describe('getDoubleattackHint', () => {
  it('賭けフェーズと決着後は助言しない', () => {
    expect(getDoubleattackHint(at({ phase: DoubleAttackPhase.BET }))).toBeNull();
    expect(getDoubleattackHint(at({ phase: DoubleAttackPhase.RESULT }))).toBeNull();
  });

  it('終局後は助言しない', () => {
    expect(
      getDoubleattackHint(at({ phase: DoubleAttackPhase.ATTACK, gameEndFlag: true, dealerCards: [card(5)] })),
    ).toBeNull();
  });

  // **弱いアップカード (2〜6) でだけ賭け増しを薦める。**
  it('弱いアップカードなら賭け増しを薦める', () => {
    for (const v of [2, 3, 4, 5, 6]) {
      const hint = getDoubleattackHint(at({ phase: DoubleAttackPhase.ATTACK, dealerCards: [card(v)] }));
      expect(hint?.targetAction, `up-card ${v}`).toBe('attack');
      expect(hint?.reason).toBe('frontendHint.doubleAttackWeakUp');
    }
  });

  // **10 が抜けているぶんディーラーはバストしにくい。** 強い札では見送り。
  it('強いアップカードなら見送りを薦める', () => {
    for (const v of [7, 8, 9, 11, 12, 13, 1]) {
      const hint = getDoubleattackHint(at({ phase: DoubleAttackPhase.ATTACK, dealerCards: [card(v)] }));
      expect(hint?.targetAction, `up-card ${v}`).toBe('decline');
      expect(hint?.reason).toBe('frontendHint.doubleAttackStrongUp');
    }
  });

  it('アップカードが無ければ助言しない', () => {
    expect(getDoubleattackHint(at({ phase: DoubleAttackPhase.ATTACK, dealerCards: [] }))).toBeNull();
  });

  it('11 以下は引く、17 以上は立つ', () => {
    const hand = (score: number) =>
      at({
        phase: DoubleAttackPhase.PLAY,
        activeHand: 0,
        hands: [{ score } as DoubleAttackResponse['hands'][number]],
      });
    expect(getDoubleattackHint(hand(9))?.targetAction).toBe('hit');
    expect(getDoubleattackHint(hand(11))?.targetAction).toBe('hit');
    expect(getDoubleattackHint(hand(17))?.targetAction).toBe('stand');
    expect(getDoubleattackHint(hand(20))?.targetAction).toBe('stand');
    expect(getDoubleattackHint(hand(15))?.reason).toBe('frontendHint.doubleAttackBorderline');
  });

  it('手札が無ければ助言しない', () => {
    expect(getDoubleattackHint(at({ phase: DoubleAttackPhase.PLAY, hands: [], activeHand: 0 }))).toBeNull();
  });
});
