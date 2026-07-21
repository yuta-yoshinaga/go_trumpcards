import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { evaluateJokerPokerMadeHand } from './jokerPokerMadeHand';

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
