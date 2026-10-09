import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { estimateGutsWinChance, evaluateGutsGuide } from './gutsGuideUtils';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('evaluateGutsGuide', () => {
  it('returns zero when a win chance is requested without a complete hand', () => {
    expect(estimateGutsWinChance([card('SPADE', 8)], [], 0)).toBe(0);
  });

  it('returns null for an empty hand', () => {
    expect(evaluateGutsGuide([])).toBeNull();
  });

  it('rates a pair as a strong (high) hand', () => {
    const guide = evaluateGutsGuide([card('SPADE', 8), card('HEART', 8)]);
    expect(guide).toMatchObject({ handKey: 'pair', tier: 'high' });
    expect(guide?.winChance).toBe(100);
  });

  it('treats a low Ace as the top rank (Ace-high) for high-card hands', () => {
    const guide = evaluateGutsGuide([card('SPADE', 1), card('HEART', 4)]);
    expect(guide).toMatchObject({ handKey: 'highcard', tier: 'medium' });
  });

  it('rates a King-high hand as medium', () => {
    const guide = evaluateGutsGuide([card('SPADE', 13), card('HEART', 11)]);
    expect(guide).toMatchObject({ handKey: 'highcard', tier: 'medium' });
  });

  it('rates a low high-card hand as weak (low)', () => {
    const guide = evaluateGutsGuide([card('SPADE', 2), card('HEART', 7)]);
    expect(guide).toMatchObject({ handKey: 'highcard', tier: 'low' });
  });

  it('rates a pair of aces as a strong hand', () => {
    const guide = evaluateGutsGuide([card('SPADE', 1), card('HEART', 1)]);
    expect(guide).toMatchObject({ handKey: 'pair', tier: 'high' });
  });

  it('estimates lower win chances with more active CPUs', () => {
    const hand = [card('SPADE', 2), card('HEART', 7)];
    const twoPlayers = estimateGutsWinChance(hand, [{ seat: 1, out: false }], 0);
    const fourPlayers = estimateGutsWinChance(
      hand,
      [
        { seat: 1, out: false },
        { seat: 2, out: false },
        { seat: 3, out: false },
      ],
      0,
    );
    expect(fourPlayers).toBeLessThan(twoPlayers);
  });

  it('applies lower-seat priority when equal hands are compared', () => {
    const pair = [card('SPADE', 8), card('HEART', 8)];
    const lowerSeatOpponent = estimateGutsWinChance(pair, [{ seat: 0, out: false }], 1);
    const higherSeatOpponent = estimateGutsWinChance(pair, [{ seat: 2, out: false }], 1);
    expect(lowerSeatOpponent).toBeLessThan(higherSeatOpponent);
  });
});
