import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { evaluateGutsGuide } from './gutsGuideUtils';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('evaluateGutsGuide', () => {
  it('returns null for an empty hand', () => {
    expect(evaluateGutsGuide([])).toBeNull();
  });

  it('rates a pair as a strong (high) hand', () => {
    const guide = evaluateGutsGuide([card('SPADE', 8), card('HEART', 8)]);
    expect(guide).toEqual({ handKey: 'pair', tier: 'high' });
  });

  it('treats a low Ace as the top rank (Ace-high) for high-card hands', () => {
    const guide = evaluateGutsGuide([card('SPADE', 1), card('HEART', 4)]);
    expect(guide).toEqual({ handKey: 'highcard', tier: 'medium' });
  });

  it('rates a King-high hand as medium', () => {
    const guide = evaluateGutsGuide([card('SPADE', 13), card('HEART', 11)]);
    expect(guide).toEqual({ handKey: 'highcard', tier: 'medium' });
  });

  it('rates a low high-card hand as weak (low)', () => {
    const guide = evaluateGutsGuide([card('SPADE', 2), card('HEART', 7)]);
    expect(guide).toEqual({ handKey: 'highcard', tier: 'low' });
  });

  it('rates a pair of aces as a strong hand', () => {
    const guide = evaluateGutsGuide([card('SPADE', 1), card('HEART', 1)]);
    expect(guide).toEqual({ handKey: 'pair', tier: 'high' });
  });
});
