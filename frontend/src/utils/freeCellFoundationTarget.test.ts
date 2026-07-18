import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { freeCellFoundationTarget } from './freeCellFoundationTarget';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const emptyFoundation = (): Card[][] => [[], [], [], []];

describe('freeCellFoundationTarget', () => {
  it('sends an Ace to its empty foundation pile', () => {
    expect(freeCellFoundationTarget(card('SPADE', 1), emptyFoundation())).toEqual({ zone: 'foundation', col: 0 });
    expect(freeCellFoundationTarget(card('CLOVER', 1), emptyFoundation())).toEqual({ zone: 'foundation', col: 1 });
    expect(freeCellFoundationTarget(card('HEART', 1), emptyFoundation())).toEqual({ zone: 'foundation', col: 2 });
    expect(freeCellFoundationTarget(card('DIAMOND', 1), emptyFoundation())).toEqual({ zone: 'foundation', col: 3 });
  });

  it('rejects a non-Ace on an empty pile', () => {
    expect(freeCellFoundationTarget(card('SPADE', 2), emptyFoundation())).toBeNull();
    expect(freeCellFoundationTarget(card('HEART', 13), emptyFoundation())).toBeNull();
  });

  it('sends the next rank up onto a matching-suit pile', () => {
    const foundation: Card[][] = [[card('SPADE', 1)], [], [card('HEART', 1), card('HEART', 2)], []];
    expect(freeCellFoundationTarget(card('SPADE', 2), foundation)).toEqual({ zone: 'foundation', col: 0 });
    expect(freeCellFoundationTarget(card('HEART', 3), foundation)).toEqual({ zone: 'foundation', col: 2 });
  });

  it('rejects a card whose rank is not exactly one above the pile top', () => {
    const foundation: Card[][] = [[card('SPADE', 1)], [], [], []];
    expect(freeCellFoundationTarget(card('SPADE', 3), foundation)).toBeNull();
    expect(freeCellFoundationTarget(card('SPADE', 1), foundation)).toBeNull();
  });

  it('rejects a card when its own suit pile is not ready even if another suit would accept it', () => {
    // Spade pile empty (needs an Ace), so the 2♠ has no legal target regardless
    // of the heart pile's contents — the target pile is chosen by suit.
    const foundation: Card[][] = [[], [], [card('HEART', 1)], []];
    expect(freeCellFoundationTarget(card('SPADE', 2), foundation)).toBeNull();
  });

  it('returns null for a card with an unknown design', () => {
    expect(freeCellFoundationTarget({ design: 'JOKER', value: 1 }, emptyFoundation())).toBeNull();
  });
});
