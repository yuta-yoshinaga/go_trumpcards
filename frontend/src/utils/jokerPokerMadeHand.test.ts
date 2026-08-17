import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { evaluateJokerPokerMadeHand, evaluateVideoPokerMadeHand } from './jokerPokerMadeHand';
import { videoPokerPayoutRows } from './videoPokerPayout';

/** Build a card from a suit + value shorthand. */
const c = (design: Card['design'], value: number): Card => ({ design, value });
const JOKER = c('JOKER', 0);

describe('evaluateJokerPokerMadeHand', () => {
  it('returns null for a non-5-card hand', () => {
    expect(evaluateJokerPokerMadeHand([])).toBeNull();
    expect(evaluateJokerPokerMadeHand([c('SPADE', 5), c('HEART', 5)])).toBeNull();
  });

  it('detects a full house (no joker)', () => {
    const hand = [c('HEART', 10), c('DIAMOND', 10), c('SPADE', 10), c('CLOVER', 5), c('HEART', 5)];
    expect(evaluateJokerPokerMadeHand(hand)?.rowKey).toBe('fullHouse');
  });

  it('pays a pair of kings (Kings or Better)', () => {
    const hand = [c('HEART', 13), c('DIAMOND', 13), c('SPADE', 9), c('CLOVER', 5), c('HEART', 2)];
    expect(evaluateJokerPokerMadeHand(hand)?.rowKey).toBe('kingsOrBetter');
  });

  it('pays a pair of aces (Kings or Better)', () => {
    const hand = [c('HEART', 1), c('DIAMOND', 1), c('SPADE', 9), c('CLOVER', 5), c('HEART', 2)];
    expect(evaluateJokerPokerMadeHand(hand)?.rowKey).toBe('kingsOrBetter');
  });

  it('does NOT pay a low pair (below kings)', () => {
    const hand = [c('HEART', 10), c('DIAMOND', 10), c('SPADE', 9), c('CLOVER', 5), c('HEART', 2)];
    expect(evaluateJokerPokerMadeHand(hand)?.rowKey).toBeNull();
  });

  it('does NOT pay a high card hand', () => {
    const hand = [c('HEART', 2), c('DIAMOND', 5), c('SPADE', 9), c('CLOVER', 11), c('HEART', 13)];
    expect(evaluateJokerPokerMadeHand(hand)?.rowKey).toBeNull();
  });

  it('promotes a low pair + joker to three of a kind', () => {
    // 5-5 + joker → joker becomes a third 5 → Three of a Kind (pays).
    const hand = [c('HEART', 5), c('DIAMOND', 5), JOKER, c('CLOVER', 9), c('HEART', 2)];
    expect(evaluateJokerPokerMadeHand(hand)?.rowKey).toBe('threeOfAKind');
  });

  it('makes five of a kind from four of a kind + joker', () => {
    const hand = [c('HEART', 8), c('DIAMOND', 8), c('SPADE', 8), c('CLOVER', 8), JOKER];
    expect(evaluateJokerPokerMadeHand(hand)?.rowKey).toBe('fiveOfAKind');
  });

  it('makes a wild royal flush when the joker completes T-J-Q-K + joker suited', () => {
    const hand = [c('HEART', 10), c('HEART', 11), c('HEART', 12), c('HEART', 13), JOKER];
    expect(evaluateJokerPokerMadeHand(hand)?.rowKey).toBe('wildRoyalFlush');
  });

  it('recognises a natural royal flush (no wild used)', () => {
    const hand = [c('HEART', 1), c('HEART', 10), c('HEART', 11), c('HEART', 12), c('HEART', 13)];
    expect(evaluateJokerPokerMadeHand(hand)?.rowKey).toBe('naturalRoyalFlush');
  });

  it('lone joker with junk cards yields no paying hand', () => {
    // Best rank is a joker-formed pair, but Kings-or-Better requires a *natural*
    // pair of aces/kings (jokers excluded), so this pays nothing — matching Go.
    const hand = [JOKER, c('SPADE', 2), c('HEART', 5), c('DIAMOND', 9), c('CLOVER', 12)];
    expect(evaluateJokerPokerMadeHand(hand)?.rowKey).toBeNull();
  });
});

// #5506: Jacks or Better では madeHand が常に null だった。ワイルドが無いだけで
// 評価器自体は流用でき、配当最低ラインが J 以上のペアに変わるだけ。
describe('evaluateVideoPokerMadeHand for Jacks or Better', () => {
  it('names a paying hand', () => {
    expect(
      evaluateVideoPokerMadeHand('videopoker', [
        c('SPADE', 5),
        c('HEART', 5),
        c('CLOVER', 5),
        c('DIAMOND', 9),
        c('SPADE', 2),
      ]),
    ).toEqual({ rowKey: 'threeOfAKind' });
  });

  // **最低ラインは J 以上のペア。** Joker Poker の K 以上とは違う。
  it('pays a pair of jacks but not a pair of tens', () => {
    const pair = (v: number) => [c('SPADE', v), c('HEART', v), c('CLOVER', 3), c('DIAMOND', 7), c('SPADE', 9)];
    expect(evaluateVideoPokerMadeHand('videopoker', pair(11))).toEqual({ rowKey: 'jacksOrBetter' });
    expect(evaluateVideoPokerMadeHand('videopoker', pair(12))).toEqual({ rowKey: 'jacksOrBetter' });
    expect(evaluateVideoPokerMadeHand('videopoker', pair(13))).toEqual({ rowKey: 'jacksOrBetter' });
    expect(evaluateVideoPokerMadeHand('videopoker', pair(1))).toEqual({ rowKey: 'jacksOrBetter' });
    expect(evaluateVideoPokerMadeHand('videopoker', pair(10))).toEqual({ rowKey: null });
  });

  // ワイルドが無いので、ロイヤルは natural/wild に分かれず 'royalFlush' 1本。
  it('uses the single royalFlush row rather than the wild split', () => {
    expect(
      evaluateVideoPokerMadeHand('videopoker', [
        c('SPADE', 1),
        c('SPADE', 13),
        c('SPADE', 12),
        c('SPADE', 11),
        c('SPADE', 10),
      ]),
    ).toEqual({ rowKey: 'royalFlush' });
  });

  // **既存の Joker Poker は変わらない。** K 以上のペアが最低ラインのまま。
  it('leaves Joker Poker on its own minimum', () => {
    const pair = (v: number) => [c('SPADE', v), c('HEART', v), c('CLOVER', 3), c('DIAMOND', 7), c('SPADE', 9)];
    expect(evaluateVideoPokerMadeHand('jokerpoker', pair(13))).toEqual({ rowKey: 'kingsOrBetter' });
    expect(evaluateVideoPokerMadeHand('jokerpoker', pair(11))).toEqual({ rowKey: null });
  });

  // 出力する行キーが、実際の配当表に存在すること。存在しないキーを返すと
  // 画面に生キーが出る。
  it('only returns keys that exist in the variant paytable', () => {
    for (const variant of ['videopoker', 'jokerpoker'] as const) {
      const keys = new Set(videoPokerPayoutRows(variant).map((r) => r.key));
      const hands: Card[][] = [
        [c('SPADE', 1), c('SPADE', 13), c('SPADE', 12), c('SPADE', 11), c('SPADE', 10)],
        [c('SPADE', 5), c('HEART', 5), c('CLOVER', 5), c('DIAMOND', 5), c('SPADE', 2)],
        [c('SPADE', 11), c('HEART', 11), c('CLOVER', 3), c('DIAMOND', 7), c('SPADE', 9)],
        [c('SPADE', 13), c('HEART', 13), c('CLOVER', 3), c('DIAMOND', 7), c('SPADE', 9)],
      ];
      for (const hand of hands) {
        const made = evaluateVideoPokerMadeHand(variant, hand);
        if (made?.rowKey) expect(keys).toContain(made.rowKey);
      }
    }
  });
});
