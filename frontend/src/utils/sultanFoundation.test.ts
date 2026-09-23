import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import {
  canPlaceCardOnSultanFoundation,
  SULTAN_CARD_VALUE_MAX,
  SULTAN_FOUNDATION_FULL,
  sultanFoundationInfo,
} from './sultanFoundation';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('sultanFoundationInfo', () => {
  it('reads the suit from the King base of a fresh pile', () => {
    const info = sultanFoundationInfo([card('SPADE', 13)]);
    expect(info.suit).toBe('♠');
    expect(info.design).toBe('SPADE');
    expect(info.count).toBe(1);
    expect(info.complete).toBe(false);
  });

  it('keeps the King base suit as the pile grows', () => {
    const info = sultanFoundationInfo([card('HEART', 13), card('HEART', 1), card('HEART', 2)]);
    expect(info.suit).toBe('♥');
    expect(info.design).toBe('HEART');
    expect(info.count).toBe(3);
    expect(info.complete).toBe(false);
  });

  it('marks a pile of 13 cards as complete', () => {
    const pile = Array.from({ length: SULTAN_FOUNDATION_FULL }, () => card('DIAMOND', 13));
    const info = sultanFoundationInfo(pile);
    expect(info.complete).toBe(true);
    expect(info.suit).toBe('♦');
  });

  it('reports no suit for an empty pile', () => {
    const info = sultanFoundationInfo([]);
    expect(info.suit).toBe('');
    expect(info.design).toBeNull();
    expect(info.count).toBe(0);
    expect(info.complete).toBe(false);
  });
});

describe('canPlaceCardOnSultanFoundation', () => {
  it('allows the next card of the same suit', () => {
    expect(canPlaceCardOnSultanFoundation(card('SPADE', 2), [[card('SPADE', 13), card('SPADE', 1)]])).toBe(true);
  });

  it('allows an Ace on a King', () => {
    expect(canPlaceCardOnSultanFoundation(card('HEART', 1), [[card('HEART', SULTAN_CARD_VALUE_MAX)]])).toBe(true);
  });

  it('rejects a two on a King', () => {
    expect(canPlaceCardOnSultanFoundation(card('HEART', 2), [[card('HEART', SULTAN_CARD_VALUE_MAX)]])).toBe(false);
  });

  it('rejects every card on a completed Queen pile', () => {
    expect(canPlaceCardOnSultanFoundation(card('HEART', 1), [[card('HEART', 13), card('HEART', 12)]])).toBe(false);
  });

  it('rejects a card of a different suit', () => {
    expect(canPlaceCardOnSultanFoundation(card('DIAMOND', 2), [[card('SPADE', 13), card('SPADE', 1)]])).toBe(false);
  });

  it('allows a card when either of two same-suit foundations accepts it', () => {
    expect(
      canPlaceCardOnSultanFoundation(card('SPADE', 2), [[card('SPADE', 13)], [card('SPADE', 13), card('SPADE', 1)]]),
    ).toBe(true);
  });
});
