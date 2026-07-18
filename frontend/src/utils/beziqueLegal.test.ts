import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { beziqueEndgameLegalIndices, beziqueSuitDesign, isBeziqueEndgameLegalPlay } from './beziqueLegal';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('beziqueSuitDesign', () => {
  it('maps the 1-4 ordinals to their design strings', () => {
    expect(beziqueSuitDesign(1)).toBe('SPADE');
    expect(beziqueSuitDesign(2)).toBe('CLOVER');
    expect(beziqueSuitDesign(3)).toBe('HEART');
    expect(beziqueSuitDesign(4)).toBe('DIAMOND');
  });

  it('returns null for an out-of-range ordinal (no trump fixed)', () => {
    expect(beziqueSuitDesign(0)).toBeNull();
    expect(beziqueSuitDesign(5)).toBeNull();
  });
});

describe('isBeziqueEndgameLegalPlay', () => {
  it('requires following the led suit when the follower holds it', () => {
    // Hand holds ♠; lead is ♠. Only the ♠ card follows.
    const hand = [c('SPADE', 12), c('DIAMOND', 11), c('HEART', 1)];
    const lead = c('SPADE', 13); // ♠K (higher than ♠Q → no in-suit winner)
    expect(isBeziqueEndgameLegalPlay(hand[0], hand, lead, 'HEART')).toBe(true); // ♠Q follows
    expect(isBeziqueEndgameLegalPlay(hand[1], hand, lead, 'HEART')).toBe(false); // ♦J off-suit
    expect(isBeziqueEndgameLegalPlay(hand[2], hand, lead, 'HEART')).toBe(false); // ♥A off-suit
  });

  it('forces a beating card when the follower can win in the led suit', () => {
    // Hand holds ♠A and ♠7; lead ♠K. Only ♠A beats ♠K, so ♠7 is illegal.
    const hand = [c('SPADE', 1), c('SPADE', 7), c('DIAMOND', 11)];
    const lead = c('SPADE', 13);
    expect(isBeziqueEndgameLegalPlay(hand[0], hand, lead, 'HEART')).toBe(true); // ♠A beats ♠K
    expect(isBeziqueEndgameLegalPlay(hand[1], hand, lead, 'HEART')).toBe(false); // ♠7 cannot beat
  });

  it('allows any in-suit card when none can beat the lead', () => {
    // Hand holds ♠Q and ♠7; lead ♠A (unbeatable). Both ♠ cards are legal.
    const hand = [c('SPADE', 12), c('SPADE', 7)];
    const lead = c('SPADE', 1);
    expect(isBeziqueEndgameLegalPlay(hand[0], hand, lead, 'HEART')).toBe(true);
    expect(isBeziqueEndgameLegalPlay(hand[1], hand, lead, 'HEART')).toBe(true);
  });

  it('requires trumping when void in the led suit but holding trump', () => {
    // Void in ♠ (lead), holds ♥ (trump). Must play a trump.
    const hand = [c('HEART', 7), c('DIAMOND', 11)];
    const lead = c('SPADE', 13);
    expect(isBeziqueEndgameLegalPlay(hand[0], hand, lead, 'HEART')).toBe(true); // ♥ trump
    expect(isBeziqueEndgameLegalPlay(hand[1], hand, lead, 'HEART')).toBe(false); // ♦ neither
  });

  it('allows any card when void in both the led suit and trump', () => {
    const hand = [c('DIAMOND', 11), c('CLOVER', 12)];
    const lead = c('SPADE', 13);
    expect(isBeziqueEndgameLegalPlay(hand[0], hand, lead, 'HEART')).toBe(true);
    expect(isBeziqueEndgameLegalPlay(hand[1], hand, lead, 'HEART')).toBe(true);
  });

  it('a higher trump must overtrump a trump lead', () => {
    // Lead is a trump ♥10; hand holds ♥A (beats) and ♥7 (loses). Must overtrump.
    const hand = [c('HEART', 1), c('HEART', 7)];
    const lead = c('HEART', 10);
    expect(isBeziqueEndgameLegalPlay(hand[0], hand, lead, 'HEART')).toBe(true); // ♥A > ♥10
    expect(isBeziqueEndgameLegalPlay(hand[1], hand, lead, 'HEART')).toBe(false); // ♥7 < ♥10
  });
});

describe('beziqueEndgameLegalIndices', () => {
  it('returns only the legal indices for the follower', () => {
    const hand = [c('SPADE', 12), c('DIAMOND', 11), c('HEART', 1)];
    const lead = c('SPADE', 13);
    expect(beziqueEndgameLegalIndices(hand, lead, 'HEART')).toEqual([0]);
  });

  it('returns every index when void in both led suit and trump', () => {
    const hand = [c('DIAMOND', 11), c('CLOVER', 12)];
    const lead = c('SPADE', 13);
    expect(beziqueEndgameLegalIndices(hand, lead, 'HEART')).toEqual([0, 1]);
  });
});
