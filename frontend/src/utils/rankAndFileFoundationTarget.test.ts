import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { rankAndFileFoundationTarget } from './rankAndFileFoundationTarget';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const emptyFoundation = (): Card[][] => [[], [], [], [], [], [], [], []];

describe('rankAndFileFoundationTarget', () => {
  it('sends an Ace to the first empty foundation pile', () => {
    // Piles are not suit-locked, so any Ace lands on pile 0 when all are empty.
    expect(rankAndFileFoundationTarget(card('SPADE', 1), emptyFoundation())).toEqual({ zone: 'foundation', col: 0 });
    expect(rankAndFileFoundationTarget(card('DIAMOND', 1), emptyFoundation())).toEqual({ zone: 'foundation', col: 0 });
  });

  it('rejects a non-Ace on an all-empty foundation', () => {
    expect(rankAndFileFoundationTarget(card('SPADE', 2), emptyFoundation())).toBeNull();
    expect(rankAndFileFoundationTarget(card('HEART', 13), emptyFoundation())).toBeNull();
  });

  it('sends the next rank up onto the matching-suit pile', () => {
    const foundation: Card[][] = [[card('SPADE', 1)], [card('HEART', 1), card('HEART', 2)], [], [], [], [], [], []];
    expect(rankAndFileFoundationTarget(card('SPADE', 2), foundation)).toEqual({ zone: 'foundation', col: 0 });
    expect(rankAndFileFoundationTarget(card('HEART', 3), foundation)).toEqual({ zone: 'foundation', col: 1 });
  });

  it('rejects a card whose rank is not exactly one above its matching pile top', () => {
    const foundation: Card[][] = [[card('SPADE', 1)], [], [], [], [], [], [], []];
    // 3♠ needs a 2♠ on top; pile 0 holds A♠ and no empty pile accepts a non-Ace.
    expect(rankAndFileFoundationTarget(card('SPADE', 3), foundation)).toBeNull();
  });

  it('routes a duplicate Ace to the next empty pile (two-deck game)', () => {
    // A♠ is already on pile 0, so the second A♠ lands on the first empty pile.
    const foundation: Card[][] = [[card('SPADE', 1)], [], [], [], [], [], [], []];
    expect(rankAndFileFoundationTarget(card('SPADE', 1), foundation)).toEqual({ zone: 'foundation', col: 1 });
  });

  it('requires a matching suit even when a same-rank slot exists on another suit', () => {
    const foundation: Card[][] = [[card('HEART', 1)], [], [], [], [], [], [], []];
    // 2♠ cannot stack on the 1♥ pile, and no empty pile accepts a non-Ace.
    expect(rankAndFileFoundationTarget(card('SPADE', 2), foundation)).toBeNull();
  });

  it('finds a later pile when earlier piles do not accept the card', () => {
    // Ace already placed on pile 0; the 2♥ must skip to its matching suit pile.
    const foundation: Card[][] = [[card('SPADE', 1)], [], [], [], [], [], [card('HEART', 1)], []];
    expect(rankAndFileFoundationTarget(card('HEART', 2), foundation)).toEqual({ zone: 'foundation', col: 6 });
  });
});
