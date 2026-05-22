import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { evaluateFiveCardHand, PokerHand, pokerSquaresRankToScore } from './pokerSquaresUtils';

const card = (design: CardDesign, value: number): Card => ({ design, value });

describe('evaluateFiveCardHand', () => {
  it('returns null when fewer than 5 cards are given', () => {
    expect(evaluateFiveCardHand([card('SPADE', 5)])).toBeNull();
  });

  it('detects high card', () => {
    expect(
      evaluateFiveCardHand([
        card('SPADE', 2),
        card('CLOVER', 5),
        card('HEART', 9),
        card('DIAMOND', 11),
        card('SPADE', 13),
      ]),
    ).toBe(PokerHand.HighCard);
  });

  it('detects one pair', () => {
    expect(
      evaluateFiveCardHand([
        card('SPADE', 2),
        card('CLOVER', 2),
        card('HEART', 9),
        card('DIAMOND', 11),
        card('SPADE', 13),
      ]),
    ).toBe(PokerHand.OnePair);
  });

  it('detects two pair', () => {
    expect(
      evaluateFiveCardHand([
        card('SPADE', 2),
        card('CLOVER', 2),
        card('HEART', 9),
        card('DIAMOND', 9),
        card('SPADE', 13),
      ]),
    ).toBe(PokerHand.TwoPair);
  });

  it('detects three of a kind', () => {
    expect(
      evaluateFiveCardHand([
        card('SPADE', 7),
        card('CLOVER', 7),
        card('HEART', 7),
        card('DIAMOND', 11),
        card('SPADE', 13),
      ]),
    ).toBe(PokerHand.ThreeOfAKind);
  });

  it('detects straight (sequential)', () => {
    expect(
      evaluateFiveCardHand([
        card('SPADE', 5),
        card('CLOVER', 6),
        card('HEART', 7),
        card('DIAMOND', 8),
        card('SPADE', 9),
      ]),
    ).toBe(PokerHand.Straight);
  });

  it('detects A-2-3-4-5 wheel straight', () => {
    expect(
      evaluateFiveCardHand([
        card('SPADE', 1),
        card('CLOVER', 2),
        card('HEART', 3),
        card('DIAMOND', 4),
        card('SPADE', 5),
      ]),
    ).toBe(PokerHand.Straight);
  });

  it('detects 10-J-Q-K-A high straight', () => {
    expect(
      evaluateFiveCardHand([
        card('SPADE', 10),
        card('CLOVER', 11),
        card('HEART', 12),
        card('DIAMOND', 13),
        card('SPADE', 1),
      ]),
    ).toBe(PokerHand.Straight);
  });

  it('detects flush (same suit, non-straight)', () => {
    expect(
      evaluateFiveCardHand([
        card('SPADE', 2),
        card('SPADE', 5),
        card('SPADE', 9),
        card('SPADE', 11),
        card('SPADE', 13),
      ]),
    ).toBe(PokerHand.Flush);
  });

  it('detects full house', () => {
    expect(
      evaluateFiveCardHand([
        card('SPADE', 7),
        card('CLOVER', 7),
        card('HEART', 7),
        card('DIAMOND', 11),
        card('SPADE', 11),
      ]),
    ).toBe(PokerHand.FullHouse);
  });

  it('detects four of a kind', () => {
    expect(
      evaluateFiveCardHand([
        card('SPADE', 9),
        card('CLOVER', 9),
        card('HEART', 9),
        card('DIAMOND', 9),
        card('SPADE', 13),
      ]),
    ).toBe(PokerHand.FourOfAKind);
  });

  it('detects straight flush', () => {
    expect(
      evaluateFiveCardHand([card('SPADE', 5), card('SPADE', 6), card('SPADE', 7), card('SPADE', 8), card('SPADE', 9)]),
    ).toBe(PokerHand.StraightFlush);
  });

  it('detects royal flush', () => {
    expect(
      evaluateFiveCardHand([
        card('SPADE', 10),
        card('SPADE', 11),
        card('SPADE', 12),
        card('SPADE', 13),
        card('SPADE', 1),
      ]),
    ).toBe(PokerHand.RoyalFlush);
  });
});

describe('pokerSquaresRankToScore', () => {
  it('matches the American scoring table', () => {
    expect(pokerSquaresRankToScore(PokerHand.HighCard)).toBe(0);
    expect(pokerSquaresRankToScore(PokerHand.OnePair)).toBe(2);
    expect(pokerSquaresRankToScore(PokerHand.TwoPair)).toBe(5);
    expect(pokerSquaresRankToScore(PokerHand.ThreeOfAKind)).toBe(10);
    expect(pokerSquaresRankToScore(PokerHand.Straight)).toBe(15);
    expect(pokerSquaresRankToScore(PokerHand.Flush)).toBe(20);
    expect(pokerSquaresRankToScore(PokerHand.FullHouse)).toBe(25);
    expect(pokerSquaresRankToScore(PokerHand.FourOfAKind)).toBe(50);
    expect(pokerSquaresRankToScore(PokerHand.StraightFlush)).toBe(75);
    expect(pokerSquaresRankToScore(PokerHand.RoyalFlush)).toBe(100);
  });
});
