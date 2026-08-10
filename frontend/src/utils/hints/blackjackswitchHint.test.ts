import { describe, expect, it } from 'vitest';
import type { BlackJackSwitchHand, BlackJackSwitchResponse, Card } from '../../types/card';
import { BlackJackSwitchPhase } from '../../types/phases';
import { getBlackjackswitchHint } from './blackjackswitchHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function hand(score: number, second: Card, overrides: Partial<BlackJackSwitchHand> = {}): BlackJackSwitchHand {
  return {
    cards: [card('SPADE', 10), second],
    score,
    bet: 10,
    stood: false,
    doubled: false,
    busted: false,
    isBJ: false,
    result: 0,
    payout: 0,
    ...overrides,
  } as BlackJackSwitchHand;
}

function base(overrides: Partial<BlackJackSwitchResponse> = {}): BlackJackSwitchResponse {
  return {
    hands: [hand(16, card('HEART', 6)), hand(15, card('CLOVER', 5))],
    dealerCards: [card('DIAMOND', 9), null],
    dealerScore: 9,
    phase: BlackJackSwitchPhase.ACTION,
    currentHandIdx: 0,
    chips: 1000,
    switched: false,
    dealerPushed22: false,
    overallResult: 0,
    totalPayout: 0,
    message: '',
    ...overrides,
  } as BlackJackSwitchResponse;
}

describe('getBlackjackswitchHint', () => {
  it('suggests betting while chips remain', () => {
    expect(getBlackjackswitchHint(base({ phase: BlackJackSwitchPhase.BET }))?.targetAction).toBe('bet');
  });

  it('says nothing in the bet phase once the chips are gone', () => {
    expect(getBlackjackswitchHint(base({ phase: BlackJackSwitchPhase.BET, chips: 0 }))).toBeNull();
  });

  it('returns null after the hand is settled', () => {
    expect(getBlackjackswitchHint(base({ phase: BlackJackSwitchPhase.END }))).toBeNull();
  });

  it('takes the switch when it turns one usable hand into two', () => {
    // 10+2=12 と 9+7=16。どちらも使えない。2 と 7 を交換すると 10+7=17 が立つ。
    // **対称な手では改善しない**（20 と 12 を入れ替えても合計は同じ）ので、
    // 片側だけが 17 に届く組を探して置いてある。
    const s = base({
      phase: BlackJackSwitchPhase.SWITCH,
      hands: [
        { ...hand(12, card('HEART', 2)), cards: [card('SPADE', 10), card('HEART', 2)] },
        { ...hand(16, card('CLOVER', 7)), cards: [card('SPADE', 9), card('CLOVER', 7)] },
      ],
    });
    expect(getBlackjackswitchHint(s)?.targetAction).toBe('switch');
  });

  it('keeps the deal when swapping makes it no better', () => {
    const s = base({
      phase: BlackJackSwitchPhase.SWITCH,
      hands: [hand(18, card('HEART', 8)), hand(18, card('CLOVER', 8))],
    });
    expect(getBlackjackswitchHint(s)?.targetAction).toBe('keep');
  });

  it('stands on seventeen or more', () => {
    const s = base({ hands: [hand(18, card('HEART', 8)), hand(15, card('CLOVER', 5))] });
    expect(getBlackjackswitchHint(s)?.targetAction).toBe('stand');
  });

  it('hits a low hand against a strong upcard', () => {
    const s = base({ dealerCards: [card('DIAMOND', 10), null] });
    expect(getBlackjackswitchHint(s)?.targetAction).toBe('hit');
  });

  it('stands on a marginal hand against a weak upcard, citing the push-22 rule', () => {
    // **ディーラー 22 はプッシュ。**崩れを当てにできないぶん、通常の BJ より
    // 手前で降りる判断になる。
    const s = base({ dealerCards: [card('DIAMOND', 5), null] });
    const hint = getBlackjackswitchHint(s);
    expect(hint?.targetAction).toBe('stand');
    expect(hint?.reason).toBe('frontendHint.blackjackswitchStandPush22');
  });

  it('hits a hand below twelve even against a weak upcard', () => {
    const s = base({
      hands: [hand(9, card('HEART', 4)), hand(15, card('CLOVER', 5))],
      dealerCards: [card('DIAMOND', 5), null],
    });
    expect(getBlackjackswitchHint(s)?.targetAction).toBe('hit');
  });

  it('says nothing for a hand that has already stood', () => {
    const s = base({ hands: [hand(15, card('HEART', 5), { stood: true }), hand(15, card('CLOVER', 5))] });
    expect(getBlackjackswitchHint(s)).toBeNull();
  });

  it('reads the ace as eleven when judging the upcard', () => {
    // A を 1 と数えると「弱い見せ札」に入ってしまい、立つ側に倒れる。
    const s = base({ dealerCards: [card('DIAMOND', 1), null] });
    expect(getBlackjackswitchHint(s)?.targetAction).toBe('hit');
  });
});

// **入れ替え後のスコアを引き算で作ってはいけない (#4708)。**エースは
// 21 を超えると 11 → 1 に落ちるので、`score - 自分の2枚目 + 相手の2枚目`
// は実際の点数と一致しない。判定は blackjackSwitchPreviewScores と同じ
// 「手札を組み直して数え直す」やり方に揃える。
describe('getBlackjackswitchHint — switch decision (sync: blackjackSwitchPreviewScores)', () => {
  const switchState = (a: BlackJackSwitchHand, b: BlackJackSwitchHand): BlackJackSwitchResponse =>
    base({ phase: BlackJackSwitchPhase.SWITCH, hands: [a, b] });

  const handOf = (first: Card, second: Card, score: number): BlackJackSwitchHand =>
    ({
      cards: [first, second],
      score,
      bet: 10,
      stood: false,
      doubled: false,
      busted: false,
      isBJ: false,
      result: 0,
      payout: 0,
    }) as BlackJackSwitchHand;

  // ♠A ♥A (=12) と ♣9 ♦8 (=17)。2枚目を入れ替えると ♠A ♦8 (=19) と
  // ♣9 ♥A (=20) で、**両手とも使える形になる**。
  // 引き算で作ると 12 - 11 + 8 = 9 になり、この入れ替えを見逃す。
  it('recommends switching a soft hand the subtraction shortcut would miss', () => {
    const hint = getBlackjackswitchHint(
      switchState(handOf(card('SPADE', 1), card('HEART', 1), 12), handOf(card('CLOVER', 9), card('DIAMOND', 8), 17)),
    );
    expect(hint?.targetAction).toBe('switch');
  });

  it('recommends keeping when the swap makes both hands worse', () => {
    // ♠10 ♥9 (=19) と ♣10 ♦8 (=18)。入れ替えると 18 と 19 で改善しない。
    const hint = getBlackjackswitchHint(
      switchState(handOf(card('SPADE', 10), card('HEART', 9), 19), handOf(card('CLOVER', 10), card('DIAMOND', 8), 18)),
    );
    expect(hint?.targetAction).toBe('keep');
  });

  // ♠10 ♥10 (=20, 使える) と ♣7 ♦7 (=14, 使えない) → 入れ替えると 17 と 17 で
  // **両手とも使える形になる**。使える手が 1 → 2 に増えるので入れ替える。
  it('recommends switching when it turns one usable hand into two', () => {
    const hint = getBlackjackswitchHint(
      switchState(handOf(card('SPADE', 10), card('HEART', 10), 20), handOf(card('CLOVER', 7), card('DIAMOND', 7), 14)),
    );
    expect(hint?.targetAction).toBe('switch');
  });

  // **片方を強くして片方を潰す入れ替えは勧めない。**♠10 ♥2 (=12) と
  // ♣9 ♦9 (=18) は入れ替えると 19 と 11 で、使える手の数は 1 のまま。
  it('does not recommend a swap that just moves the problem', () => {
    const hint = getBlackjackswitchHint(
      switchState(handOf(card('SPADE', 10), card('HEART', 2), 12), handOf(card('CLOVER', 9), card('DIAMOND', 9), 18)),
    );
    expect(hint?.targetAction).toBe('keep');
  });
});
